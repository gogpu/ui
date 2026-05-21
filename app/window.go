package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/gogpu/gpucontext"
	ui "github.com/gogpu/ui"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/internal/dirty"
	ifocus "github.com/gogpu/ui/internal/focus"
	internalRender "github.com/gogpu/ui/internal/render"
	"github.com/gogpu/ui/overlay"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/theme"
	"github.com/gogpu/ui/widget"
)

// dirtyBoundaryEntry tracks a RepaintBoundary that has been marked dirty
// by upward propagation. The key is the boundary's CacheKey for deduplication.
// The struct is intentionally minimal — only the key matters for O(1) lookup.
// Future: add depth field for deepest-first paint ordering.
type dirtyBoundaryEntry struct {
	// present is always true. The struct exists so the map value is not empty,
	// allowing future extension (depth, boundary reference) without changing
	// the AddDirtyBoundary signature.
	present bool
}

const (
	// defaultWidth is the default window width in headless mode.
	defaultWidth = 800
	// defaultHeight is the default window height in headless mode.
	defaultHeight = 600
	// defaultScale is the default DPI scale factor.
	defaultScale = 1.0
)

// Window manages the widget tree and coordinates layout, draw, and
// event dispatch for a single window.
//
// Window is created by [App] and should not be instantiated directly.
// It is NOT safe for concurrent access.
// animationStopper can stop a continuous render session.
type animationStopper interface {
	Stop()
}

type Window struct {
	root      widget.Widget
	ctx       *widget.ContextImpl
	wp        gpucontext.WindowProvider
	pp        gpucontext.PlatformProvider
	scheduler *state.Scheduler
	theme     *theme.Theme
	overlays  *overlay.Stack
	focusMgr  *ifocus.Manager

	// renderMode controls background ownership and incremental rendering.
	// See RenderMode documentation for details.
	renderMode RenderMode

	// animToken is non-nil while continuous rendering is active for animations.
	animToken animationStopper

	// animIdleFrames counts consecutive frames with no Invalidate call.
	// Used to stop the animation pumper after animations complete.
	animIdleFrames int

	// needsAnimationFrame is set by ScheduleAnimationFrame (during Draw)
	// and persists across ClearAfterPaint. Checked by desktop.draw frame
	// skip to ensure animated boundary frames are not skipped.
	// Flutter equivalent: _hasScheduledFrame.
	needsAnimationFrame bool

	// needsLayout indicates that layout should be recalculated.
	needsLayout bool

	// needsRedraw indicates that the draw pass should be performed.
	// This is set when any widget in the tree needs re-rendering,
	// and cleared after a successful draw. When false, DrawTo can
	// skip rendering entirely because the previous frame's output
	// is still valid in the GPU framebuffer.
	needsRedraw bool

	// needsFullRepaint forces a complete redraw of the entire widget tree
	// on the next DrawTo call. Set on first frame, resize, theme change,
	// and SetRoot. When true, DrawTo clears the entire pixmap and draws
	// all widgets unconditionally instead of using incremental dirty regions.
	needsFullRepaint bool

	// dirtyTracker collects and merges dirty regions for incremental redraw.
	// Populated by dirtyCollector before each DrawTo call.
	dirtyTracker *dirty.Tracker

	// dirtyCollector walks the widget tree to find dirty widgets and records
	// their bounds in dirtyTracker.
	dirtyCollector *dirty.Collector

	// lastDrawStats holds per-widget statistics from the most recent
	// draw traversal. Updated by DrawTo() and the headless draw() path.
	lastDrawStats widget.DrawStats

	// lastDirtyUnion is the union of all dirty regions from the most recent
	// dirty-region-only repaint. Zero when the last frame was a full repaint
	// or frame skip. Used by desktop.Run to pass dirty region to ggcanvas
	// for partial texture upload.
	lastDirtyUnion geometry.Rect

	// lastWasFullRepaint indicates that the most recent DrawTo performed a
	// full repaint (first frame, resize, theme change). When true, the
	// entire pixmap was modified and needs full GPU upload.
	lastWasFullRepaint bool

	// hoveredWidget tracks the widget currently under the mouse pointer.
	// Used for sending MouseEnter/MouseLeave events to individual widgets
	// as the mouse moves across the widget tree.
	hoveredWidget widget.Widget

	// capturedWidget receives all mouse events directly during drag
	// operations, bypassing tree hit-testing (ADR-031 PointerCapturer).
	// Set by CapturePointer, cleared by ReleasePointer or auto-released
	// on MouseRelease when all buttons are up.
	// Enterprise references: Win32 SetCapture, Qt grabMouse.
	capturedWidget widget.Widget

	// mouseButtonsHeld tracks pressed mouse buttons to prevent cursor
	// reset during drag operations (Frame.ResetCursor skipped while dragging).
	mouseButtonsHeld event.ButtonState

	// windowSize tracks the last known window size in physical pixels.
	windowSize geometry.Size

	// frameCallback, if set, is called after each frame with statistics.
	frameCallback FrameCallback

	// imageCache is the centralized LRU cache for RepaintBoundary pixel
	// buffers. All RepaintBoundary instances in this window share this
	// cache, which enforces a memory budget (default 64MB) and evicts
	// least-recently-used entries. Phase 5, ADR-004.
	imageCache *internalRender.ImageCache

	// dirtyBoundaries collects RepaintBoundary instances marked dirty by
	// upward propagation (ADR-007, Task 1e). Populated by the
	// onBoundaryDirty callback wired during mount. Used by future Phase 2
	// PaintDirtyBoundaries to repaint only changed boundaries.
	dirtyBoundaries map[uint64]dirtyBoundaryEntry
}

// newWindow creates a Window with the given providers.
func newWindow(
	wp gpucontext.WindowProvider,
	pp gpucontext.PlatformProvider,
	scheduler *state.Scheduler,
	t *theme.Theme,
	renderMode RenderMode,
) *Window {
	ctx := widget.NewContext()

	tracker := dirty.NewTracker()
	imgCache := internalRender.DefaultImageCache()
	w := &Window{
		ctx:              ctx,
		wp:               wp,
		pp:               pp,
		scheduler:        scheduler,
		theme:            t,
		renderMode:       renderMode,
		focusMgr:         ifocus.New(nil),
		needsLayout:      true,
		needsFullRepaint: true,
		dirtyTracker:     tracker,
		dirtyCollector:   dirty.NewCollector(tracker),
		imageCache:       imgCache,
		dirtyBoundaries:  make(map[uint64]dirtyBoundaryEntry),
	}

	// Set centralized image cache on context so RepaintBoundary instances
	// use the shared LRU cache with memory budget enforcement (Phase 5).
	ctx.SetImageCache(imgCache)

	// Initialize overlay stack.
	w.overlays = overlay.NewStack(func() {
		w.needsLayout = true
		w.needsRedraw = true
		if w.wp != nil {
			w.wp.RequestRedraw()
		}
	})

	// Set scheduler on context so widgets can bind signals on mount.
	ctx.SetScheduler(scheduler)

	// Set the theme provider on the context so widgets can access theme.
	if t != nil {
		ctx.SetThemeProvider(t)
	}

	// Set initial scale from WindowProvider.
	w.updateScale()

	// Set initial window size.
	w.updateWindowSize()

	// Set overlay manager on context so widgets can push/remove overlays.
	ctx.SetOverlayManager(&windowOverlayManager{window: w})

	// Wire invalidation callback to request redraw.
	// ADR-028: Invalidate triggers layout + redraw, but does NOT mark
	// ALL widgets dirty. Widgets that need redraw call SetNeedsRedraw
	// themselves. MarkRedrawInTree was the source of full-window dirty
	// on every ctx.Invalidate() call (ScrollView full-window green).
	ctx.SetOnInvalidate(func() {
		w.needsLayout = true
		w.needsRedraw = true
		if w.wp != nil {
			w.wp.RequestRedraw()
		}
	})

	// Wire partial invalidation callback.
	// InvalidateRect = visual change in a specific region (redraw only, no layout).
	// Used by widgets with local animations (spinner, progress) that don't
	// affect layout of other widgets.
	ctx.SetOnInvalidateRect(func(_ geometry.Rect) {
		w.needsRedraw = true
		if w.wp != nil {
			w.wp.RequestRedraw()
		}
	})

	// Wire animation frame scheduling (Flutter scheduleFrame pattern).
	// Animated widgets call ScheduleAnimationFrame() instead of InvalidateRect()
	// to avoid triggering immediate RequestRedraw. This keeps the animPumper
	// alive without forcing a render on every call. The animPumper ticks at
	// its configured rate (30fps default) and triggers renders.
	ctx.SetOnScheduleAnimation(func() {
		w.animIdleFrames = 0
		w.needsAnimationFrame = true
		if w.animToken == nil && w.wp != nil {
			w.animToken = newAnimPumper(w.wp)
		}
	})

	// Wire pointer capture callbacks (ADR-031). Widgets call CapturePointer
	// during drag start to receive mouse events directly, bypassing tree
	// hit-testing. Auto-released on MouseRelease when all buttons are up.
	ctx.SetOnCapturePointer(func(wgt widget.Widget) {
		w.capturedWidget = wgt
	})
	ctx.SetOnReleasePointer(func(wgt widget.Widget) {
		if w.capturedWidget == wgt {
			w.capturedWidget = nil
		}
	})

	// Wire dirty boundary registration so upward propagation populates the
	// flat dirty boundary set. This replaces O(n) NeedsRedrawInTreeNonBoundary
	// tree walks with O(1) HasDirtyBoundaries map lookup for frame skip.
	// Flutter equivalent: markNeedsPaint adds to _nodesNeedingPaint list.
	ctx.SetOnRegisterDirtyBoundary(func(key uint64) {
		w.AddDirtyBoundary(key)
		// Wake the render loop WITHOUT setting needsRedraw (which would
		// force root re-recording). HasDirtyBoundaries is sufficient for
		// frame skip. RequestRedraw only wakes the loop.
		if w.wp != nil {
			w.wp.RequestRedraw()
		}
	})

	// Wire scheduler to wake render loop when signals change.
	// Signal dirty = visual content changed (redraw only).
	// Layout is NOT needed — widget size/position unchanged.
	// Widgets that need relayout after signal change call ctx.Invalidate().
	if wp != nil {
		scheduler.SetOnDirty(func() {
			w.needsRedraw = true
			if w.wp != nil {
				w.wp.RequestRedraw()
			}
		})
	}

	return w
}

// SetRoot sets the root widget for this window.
//
// Setting a new root triggers a full layout on the next Frame call.
// The old root tree is unmounted and the new root tree is mounted,
// which triggers signal binding setup/teardown via [widget.Lifecycle].
func (w *Window) SetRoot(root widget.Widget) {
	// Unmount old tree (triggers RepaintBoundary.Unmount which invalidates
	// individual cache entries from the shared ImageCache).
	if w.root != nil {
		widget.UnmountTree(w.root)
	}

	// Clear the entire image cache since the old widget tree is gone and
	// the new tree will have different boundaries with different keys.
	if w.imageCache != nil {
		w.imageCache.InvalidateAll()
	}

	w.root = root

	// ADR-007 Phase 5: Root IS boundary (Flutter RenderView.isRepaintBoundary).
	// DrawChild skips child boundaries during recording (BoundaryRecorder).
	// Compositor Layer Tree assembles all boundary scenes by reference.
	type boundaryEnabler interface {
		SetRepaintBoundary(bool)
	}
	if be, ok := root.(boundaryEnabler); ok {
		be.SetRepaintBoundary(true)
	}

	w.needsLayout = true
	w.needsRedraw = true
	w.needsFullRepaint = true

	// Clear references to old tree widgets to prevent pinning the
	// entire unmounted tree in memory via parent/children links.
	w.hoveredWidget = nil
	w.capturedWidget = nil

	// Clear references to old tree widgets to prevent pinning the
	// entire unmounted tree in memory via parent/children links.
	w.hoveredWidget = nil
	w.capturedWidget = nil

	// Update focus manager with new root so Tab navigation
	// traverses the correct widget tree.
	w.focusMgr.SetRoot(root)

	// Mount new tree and mark all widgets as needing redraw.
	if w.root != nil {
		widget.MountTree(w.root, w.ctx)
		widget.MarkRedrawInTree(w.root)
	}
}

// Root returns the current root widget, or nil if none is set.
func (w *Window) Root() widget.Widget {
	return w.root
}

// Context returns the widget context used for layout, draw, and events.
func (w *Window) Context() *widget.ContextImpl {
	return w.ctx
}

// Theme returns the window's current theme.
func (w *Window) Theme() *theme.Theme {
	return w.theme
}

// setTheme updates the window's theme and marks layout as dirty.
func (w *Window) setTheme(t *theme.Theme) {
	w.theme = t
	w.ctx.SetThemeProvider(t)
	w.needsLayout = true
	w.needsRedraw = true
	w.needsFullRepaint = true
	if w.root != nil {
		widget.MarkRedrawInTree(w.root)
	}
}

// HandleEvent dispatches a single event to the widget tree.
//
// Events are first offered to the overlay stack (top overlay has priority).
// If no overlay consumes the event (and no modal overlay blocks it),
// key events are offered to the focus manager for Tab/Shift+Tab navigation
// and registered shortcuts. Finally, unconsumed events are propagated to
// the root widget.
func (w *Window) HandleEvent(e event.Event) {
	if w.root == nil || e == nil {
		return
	}

	// Update context time for event processing.
	w.ctx.SetNow(time.Now())

	// Overlays get priority.
	if w.overlays.HandleEvent(w.ctx, e) {
		return
	}

	// Let focus manager intercept Tab/Shift+Tab and shortcuts.
	if ke, ok := e.(*event.KeyEvent); ok {
		w.syncContextFocusToManager()

		if w.focusMgr.HandleKeyEvent(ke) {
			w.syncManagerFocusToContext()
			return
		}
	}

	// ADR-031: Pointer capture — deliver mouse events directly to the
	// captured widget, bypassing tree hit-testing. This enables drag
	// operations (slider, splitview, scrollbar, text selection) to work
	// when the mouse moves outside the widget's bounds.
	// Auto-releases on MouseRelease when all buttons are up.
	if me, ok := e.(*event.MouseEvent); ok && w.capturedWidget != nil {
		switch me.MouseType {
		case event.MouseMove, event.MouseRelease, event.MousePress:
			consumed := w.capturedWidget.Event(w.ctx, me)
			// Auto-release when all mouse buttons are up (drag ended).
			if me.MouseType == event.MouseRelease && me.Buttons == 0 {
				w.capturedWidget = nil
			}
			// Track mouse button state even during capture.
			w.mouseButtonsHeld = me.Buttons
			if consumed {
				return
			}
		}
	}

	// Track widget-level hover for MouseEnter/MouseLeave dispatch.
	// Skip hover updates during drag (button pressed) to prevent cursor
	// from resetting when dragging over other widgets (e.g., SplitView divider).
	if me, ok := e.(*event.MouseEvent); ok {
		switch me.MouseType {
		case event.MouseMove:
			if me.Buttons == 0 {
				w.updateHover(me.Position, me.Buttons, me.Modifiers())
			}
		case event.MouseLeave:
			// Mouse left the window entirely — clear hover state.
			// The window-level leave event still propagates to the root below.
			w.clearHover(me.Buttons, me.Modifiers())
		}
	}

	// Track mouse button state for drag cursor protection.
	if me, ok := e.(*event.MouseEvent); ok {
		w.mouseButtonsHeld = me.Buttons
	}

	// Dispatch event to root widget.
	_ = w.root.Event(w.ctx, e)

	// Sync cursor immediately after event dispatch so hover cursor
	// changes are visible without waiting for the next Frame() tick.
	if w.pp != nil {
		w.syncCursor()
	}

	// After widget tree processes a mouse press, a widget may have called
	// ctx.RequestFocus. Sync that to the focus manager so subsequent
	// Tab navigation starts from the correct position.
	if me, ok := e.(*event.MouseEvent); ok && me.MouseType == event.MousePress {
		w.syncContextFocusToManager()
	}
}

// HandleResize processes a window resize.
//
// This updates the window size and marks layout as needing recalculation.
func (w *Window) HandleResize(width, height int) {
	w.windowSize = geometry.Sz(float32(width), float32(height))
	w.needsLayout = true
	w.needsRedraw = true
	w.needsFullRepaint = true
	if w.root != nil {
		widget.MarkRedrawInTree(w.root)
	}
}

// HandleFocusChange processes a window focus change.
func (w *Window) HandleFocusChange(focused bool) {
	if w.root == nil {
		return
	}

	var focusType event.FocusEventType
	if focused {
		focusType = event.FocusGained
	} else {
		focusType = event.FocusLost
	}

	e := event.NewFocusEvent(focusType)
	_ = w.root.Event(w.ctx, e)

	// Request redraw so widgets can update visual state (e.g. titlebar
	// active/inactive appearance, focus rings).
	w.needsRedraw = true
	if w.wp != nil {
		w.wp.RequestRedraw()
	}
}

// Frame performs one complete frame cycle.
//
// The frame cycle consists of:
//  1. Update time tracking
//  2. Flush pending scheduler updates
//  3. Perform layout if needed
//  4. Draw the widget tree
//  5. Sync cursor state to platform
//  6. Clear invalidation flags
//
// Frame is a no-op if there is no root widget.
func (w *Window) Frame() {
	if w.root == nil {
		return
	}

	frameStart := time.Now()
	didLayout := w.needsLayout

	// Begin frame timing. DeltaTime = time since last BeginFrame.
	w.ctx.BeginFrame(frameStart)

	// Cursor is managed in Event handlers (updateHover resets on target
	// change, widgets set Pointer in MouseEnter/Default in MouseLeave).
	// No ResetCursor here — Frame runs after Event and would overwrite
	// the cursor set by the widget, causing flash on next syncCursor.

	// Flush pending signal changes (may trigger new dirty marks).
	// The scheduler's flushFn sets persistent needsRedraw flags on widgets.
	const maxReflushes = 2
	for i := 0; i <= maxReflushes; i++ {
		w.scheduler.Flush()
		if w.scheduler.PendingCount() == 0 {
			break
		}
	}

	// Update scale factor (may change between frames on multi-monitor setups).
	w.updateScale()

	// Update window size from provider.
	w.updateWindowSize()

	// Perform layout if needed.
	// Layout changes always require a redraw since widget positions may shift.
	var layoutDur time.Duration
	if w.needsLayout {
		ui.Logger().Info("[LAYOUT-TRIGGER]")
		layoutStart := time.Now()
		w.layout()
		layoutDur = time.Since(layoutStart)
		// Clear needsLayout only if no widget re-invalidated during layout.
		// Animations call ctx.Invalidate() from tickAnimation() during layout,
		// which sets needsLayout back to true — we must not clobber that.
		if !w.ctx.IsInvalidated() {
			w.needsLayout = false
		}
		// Layout completed — widgets with changed positions need redraw.
		// ADR-028: do NOT MarkRedrawInTree(root) — that marks ALL widgets
		// dirty → full screen repaint. Only widgets that actually changed
		// should be dirty (they called SetNeedsRedraw during layout).
		w.needsRedraw = true
	}

	// ADR-028 Phase C: O(1) dirty check. The needsRedraw flag is set by
	// onInvalidate, onInvalidateRect, and scheduler.SetOnDirty callbacks.
	// Boundary widgets populate dirtyBoundaries via RegisterDirtyBoundary.
	// No O(n) NeedsRedrawInTreeNonBoundary tree walk needed.

	// Draw the widget tree.
	// In hosted mode (wp != nil), DrawTo() is called later by the host
	// application with a real canvas — so we must NOT clear redraw flags here.
	// In headless mode (wp == nil), there is no DrawTo() call, so we collect
	// stats and clear flags ourselves.
	drawStart := time.Now()
	drawSkipped := !w.needsRedraw
	if w.needsRedraw && w.wp == nil {
		w.draw()
		w.needsRedraw = false
	}
	drawDur := time.Since(drawStart)

	// Sync cursor to platform.
	w.syncCursor()

	// Manage continuous rendering for animations.
	// If a widget called Invalidate during this frame (e.g., animation tick),
	// enter continuous (vsync) rendering mode via StartAnimation.
	// When no more animations are active, stop continuous mode.
	// Manage animation frame pumping.
	// Start pumper when any animation is active (Invalidate from tickAnimation).
	// Keep pumper running for a few extra frames to handle animation completion
	// and prevent start/stop thrashing from periodic data updates.
	if w.ctx.IsInvalidated() || !w.ctx.InvalidatedRect().IsEmpty() || w.needsAnimationFrame {
		w.animIdleFrames = 0
		if w.animToken == nil && w.wp != nil {
			w.animToken = newAnimPumper(w.wp)
		}
	} else if w.animToken != nil {
		w.animIdleFrames++
		// Stop pumper after 3 consecutive idle frames (no Invalidate).
		// This handles the case where data sim triggers periodic Invalidate
		// but no animation is running.
		if w.animIdleFrames > 3 {
			w.animToken.Stop()
			w.animToken = nil
			w.animIdleFrames = 0
		}
	}

	// Clear invalidation state.
	w.ctx.ClearInvalidation()

	// Report frame statistics if callback is set.
	if w.frameCallback != nil {
		w.frameCallback(FrameStats{
			FrameStart:      frameStart,
			LayoutDuration:  layoutDur,
			DrawDuration:    drawDur,
			TotalDuration:   time.Since(frameStart),
			LayoutPerformed: didLayout,
			DrawSkipped:     drawSkipped,
			DrawStats:       w.lastDrawStats,
		})
	}
}

// NeedsLayout returns true if layout needs recalculation.
func (w *Window) NeedsLayout() bool {
	return w.needsLayout
}

// NeedsRedraw reports whether the window-level redraw flag is set.
//
// This is an O(1) check — the flag is set by onInvalidate, onInvalidateRect,
// and scheduler.SetOnDirty callbacks. No tree walk is performed.
//
// ADR-028 Phase C: removed O(n) NeedsRedrawInTreeNonBoundary fallback.
// All dirty propagation paths now set this flag or populate dirtyBoundaries.
func (w *Window) NeedsRedraw() bool {
	return w.needsRedraw
}

// LastDrawStats returns the per-widget statistics from the most recent
// draw traversal (either via [Window.DrawTo] or the headless draw path
// inside [Window.Frame]).
//
// When the last frame was skipped (no dirty widgets), the returned stats
// are zero-valued.
func (w *Window) LastDrawStats() widget.DrawStats {
	return w.lastDrawStats
}

// DirtyRegionCount returns the number of dirty regions from the most recent
// DrawTo call. This reflects the state after Collector.Collect + Tracker.Optimize
// in the last frame.
//
// Returns 0 when the last frame was a full repaint (no incremental regions)
// or when the frame was skipped (nothing dirty).
//
// This is an observability hook for monitoring incremental rendering efficiency.
func (w *Window) DirtyRegionCount() int {
	return w.dirtyTracker.RegionCount()
}

// DirtyRegions returns the list of dirty widget regions from the most
// recent DrawTo call. Each region corresponds to a widget (or group of
// nearby widgets) that needed redraw.
//
// Used by desktop.Run for debug overlay (GOGPU_DEBUG_DIRTY=1) and for
// passing damage rects to the OS compositor (SetDamageRects).
func (w *Window) DirtyRegions() []geometry.Rect {
	regions := w.dirtyTracker.DirtyRegions()
	rects := make([]geometry.Rect, len(regions))
	for i, r := range regions {
		rects[i] = r.Bounds
	}
	return rects
}

// LastDirtyUnion returns the union of all dirty regions from the most
// recent dirty-region-only repaint. Returns a zero Rect when the last
// frame was a full repaint or a frame skip.
//
// Used by desktop.Run to pass the dirty region to ggcanvas for partial
// texture upload — only the modified region is uploaded to the GPU
// instead of the entire pixmap.
func (w *Window) LastDirtyUnion() geometry.Rect {
	return w.lastDirtyUnion
}

// WasFullRepaint returns true if the most recent DrawTo performed a full
// repaint (first frame, resize, theme change, SetRoot). When true, the
// entire pixmap was modified and needs full GPU upload.
func (w *Window) WasFullRepaint() bool {
	return w.lastWasFullRepaint
}

// WindowSize returns the current window size in logical pixels.
func (w *Window) WindowSize() geometry.Size {
	return w.windowSize
}

// syncManagerFocusToContext updates the widget context's focused widget
// to match the focus manager's state. Called after the focus manager
// moves focus via Tab/Shift+Tab navigation.
func (w *Window) syncManagerFocusToContext() {
	focused := w.focusMgr.Focused()
	if focused == nil {
		// Focus manager cleared focus.
		current := w.ctx.FocusedWidget()
		if current != nil {
			w.ctx.ReleaseFocus(current)
		}
		return
	}

	// The Focusable interface doesn't embed Widget, but in practice all
	// focusable widgets implement Widget (they embed WidgetBase). Use
	// type assertion to get the Widget interface for the context.
	if fw, ok := focused.(widget.Widget); ok {
		w.ctx.RequestFocus(fw)
		w.needsRedraw = true
		if w.wp != nil {
			w.wp.RequestRedraw()
		}
	}
}

// syncContextFocusToManager updates the focus manager's state to match
// the widget context. Called before Tab processing so navigation starts
// from the widget that received focus via mouse click or programmatic
// ctx.RequestFocus.
func (w *Window) syncContextFocusToManager() {
	ctxFocused := w.ctx.FocusedWidget()
	mgrFocused := w.focusMgr.Focused()

	// Check if they already agree.
	if ctxFocused == nil && mgrFocused == nil {
		return
	}

	// If context has a focused widget, sync it to the manager.
	if ctxFocused != nil {
		if f, ok := ctxFocused.(widget.Focusable); ok {
			if mgrFocused != f {
				w.focusMgr.Focus(f)
			}
			return
		}
	}

	// Context has no focus (or non-focusable widget); clear manager focus.
	if mgrFocused != nil {
		w.focusMgr.Blur()
	}
}

// layout performs the layout pass on the widget tree and overlays.
func (w *Window) layout() {
	if w.root == nil {
		return
	}

	// Update window size on context so widgets can access it.
	w.ctx.SetWindowSize(w.windowSize)

	// Create tight constraints matching the window size.
	constraints := geometry.Tight(w.windowSize)

	// Layout the root widget.
	size := w.root.Layout(w.ctx, constraints)

	// Set root bounds to fill the window from origin.
	if setter, ok := w.root.(interface{ SetBounds(geometry.Rect) }); ok {
		setter.SetBounds(geometry.NewRect(0, 0, size.Width, size.Height))
	}

	// Layout overlays with window-sized constraints.
	w.overlays.Layout(w.ctx, w.windowSize)

	// Refresh focus manager's root after layout so the focusable
	// widget list reflects any tree changes (widgets added/removed,
	// visibility/enabled state changes).
	w.focusMgr.SetRoot(w.root)
}

// draw performs the draw pass on the widget tree in headless mode.
func (w *Window) draw() {
	if w.root == nil {
		return
	}

	// Headless mode: no canvas available, so we collect statistics and
	// clear redraw flags without actually calling Draw on widgets.
	// Real rendering happens via DrawTo() when the host provides a canvas.
	w.lastDrawStats = widget.CollectDrawStats(w.root)
	widget.ClearRedrawInTree(w.root)
}

// RenderMode returns the window's current rendering mode.
func (w *Window) RenderMode() RenderMode {
	return w.renderMode
}

// SetRenderMode changes the window's rendering mode at runtime.
//
// Switching to [RenderModeFrameworkManaged] enables frame skip and
// dirty-region rendering. Switching to [RenderModeHostManaged] restores
// full-tree draw every frame.
//
// A full repaint is forced after the switch to ensure the persistent
// pixmap is populated correctly in the new mode.
func (w *Window) SetRenderMode(mode RenderMode) {
	w.renderMode = mode
	w.needsFullRepaint = true
	if w.root != nil {
		widget.MarkRedrawInTree(w.root)
	}
}

// DrawTo performs the draw pass using the provided canvas.
//
// Behavior depends on the [RenderMode]:
//
// In RenderModeHostManaged (default):
//   - The host draws background before calling DrawTo.
//   - DrawTo does NOT call canvas.Clear (host already painted background).
//   - DrawTo always draws the full widget tree and returns true.
//   - dirty.Tracker is populated for RepaintBoundary Intersects fast path.
//
// In RenderModeFrameworkManaged:
//   - Level 1 (Frame Skip): If nothing changed since the last draw,
//     returns false immediately — zero CPU, zero GPU upload.
//   - Level 2 (Dirty Region Redraw): Only dirty regions are redrawn on
//     the persistent pixmap. Clean regions retain previous pixels
//     (Qt QBackingStore pattern).
//   - Full Repaint: On first frame, resize, theme change, or SetRoot,
//     the entire pixmap is cleared and redrawn.
//
// Returns true if rendering was performed, false if the draw was skipped.
func (w *Window) DrawTo(canvas widget.Canvas) bool {
	if w.root == nil || canvas == nil {
		return false
	}

	// Collect dirty regions (always — for RepaintBoundary Intersects fast path).
	w.dirtyTracker.Reset()
	w.dirtyCollector.Collect(w.root)
	w.dirtyTracker.Optimize()

	var drawn bool
	switch w.renderMode {
	case RenderModeFrameworkManaged:
		drawn = w.drawFrameworkManaged(canvas)
	default:
		drawn = w.drawHostManaged(canvas)
	}

	if drawn {
		ui.Logger().LogAttrs(context.Background(), slog.LevelDebug,
			"DrawTo: rendered",
			slog.Int("dirty", w.lastDrawStats.DirtyWidgets),
			slog.Int("clean", w.lastDrawStats.CleanWidgets),
			slog.Int("cached", w.lastDrawStats.CachedWidgets),
			slog.Int("total", w.lastDrawStats.TotalWidgets),
			slog.Int("dirtyRegions", w.dirtyTracker.RegionCount()),
			slog.String("mode", w.renderMode.String()),
		)
	}

	return drawn
}

// drawHostManaged draws the full widget tree without clearing the canvas.
// The host draws background before calling DrawTo.
//
// Dirty-region optimization is handled by RepaintBoundary: clean boundaries
// serve cached GPU textured quads (cheap blit), dirty boundaries re-render
// their subtree. The dirty.Tracker feeds RepaintBoundary.Intersects for
// O(regions) fast-path cache validation instead of O(tree_depth) walks.
//
// Canvas-level dirty clip is NOT used here because the host draws full
// background every frame — clipping would prevent clean widgets from
// being redrawn over the fresh background.
func (w *Window) drawHostManaged(canvas widget.Canvas) bool {
	w.ctx.SetDirtyTracker(w.dirtyTracker)
	defer w.ctx.SetDirtyTracker(nil)

	w.lastDrawStats = widget.DrawTree(w.root, w.ctx, canvas)

	w.overlays.Draw(w.ctx, canvas)
	widget.ClearRedrawInTree(w.root)
	w.needsRedraw = false
	w.needsFullRepaint = false
	return true
}

// drawFrameworkManaged implements the three-level incremental rendering
// pipeline (ADR-004). The framework owns the pixmap lifecycle and draws
// the theme background when needed.
//
// Level 1 (frame skip): if nothing changed and the persistent pixmap is
// valid, returns false immediately -- zero CPU, zero GPU upload.
//
// Level 2 (dirty-region-only repaint): only the regions that changed are
// cleared and redrawn on the persistent pixmap. Clean regions retain the
// previous frame's pixels (Qt QBackingStore / Win32 partial-paint pattern).
//
// Full repaint: on first frame, resize, theme change, SetRoot, layout
// change, or when too many dirty regions accumulate (>16), the entire
// pixmap is cleared and redrawn.
func (w *Window) drawFrameworkManaged(canvas widget.Canvas) bool {
	hasTreeDirty := !w.dirtyTracker.IsEmpty() || w.needsRedraw || widget.NeedsRedrawInTree(w.root)

	// Level 1: Frame skip — nothing changed and pixmap is persistent.
	if !hasTreeDirty && !w.needsFullRepaint {
		return false
	}

	// Clear redraw flags BEFORE draw so flags set during Draw (e.g.,
	// spinner's SetNeedsRedraw(true)) survive until the next frame's
	// collector run. If cleared AFTER draw, the collector on the next
	// frame sees NeedsRedraw=false and misses the widget.
	widget.ClearRedrawInTree(w.root)

	w.ctx.SetDirtyTracker(w.dirtyTracker)
	defer w.ctx.SetDirtyTracker(nil)

	// Reset dirty union from previous frame.
	w.lastDirtyUnion = geometry.Rect{}
	w.lastWasFullRepaint = false

	// Full repaint on first frame, resize, theme change, SetRoot,
	// layout change, or when too many dirty regions accumulate.
	if w.needsFullRepaint || w.dirtyTracker.NeedsFullRepaint() {
		w.drawFullRepaint(canvas)
	} else {
		// Level 2: Dirty-region-only repaint.
		w.drawDirtyRegions(canvas)
	}

	w.overlays.Draw(w.ctx, canvas)
	w.needsRedraw = false
	w.needsFullRepaint = false
	return true
}

// drawFullRepaint clears the entire canvas and redraws the full widget tree.
// Used in FrameworkManaged mode on first frame, resize, theme change, and SetRoot.
func (w *Window) drawFullRepaint(canvas widget.Canvas) {
	bg := w.ThemeBackground()
	canvas.Clear(bg)

	w.lastDrawStats = widget.DrawTree(w.root, w.ctx, canvas)
	w.lastWasFullRepaint = true
}

// drawDirtyRegions clears and redraws only the dirty regions.
// Clean regions retain previous frame's pixels (persistent pixmap).
// Follows the Qt QBackingStore partial-paint pattern.
//
// The algorithm:
//  1. Compute the union of all dirty regions for a single clip rect.
//  2. Clear each dirty region with the theme background.
//  3. Clip to the dirty union and draw the full tree. Widgets outside
//     the clip early-exit from visibility checks, so only widgets
//     overlapping dirty regions actually render.
func (w *Window) drawDirtyRegions(canvas widget.Canvas) {
	bg := w.ThemeBackground()
	regions := w.dirtyTracker.DirtyRegions()
	if len(regions) == 0 {
		return
	}

	// Compute union of all dirty regions for a single clip.
	union := regions[0].Bounds
	for i := 1; i < len(regions); i++ {
		union = union.Union(regions[i].Bounds)
	}

	// Clear dirty regions with theme background using CPU-only fill.
	// DrawRect would trigger the GPU SDF accelerator, queuing SDF shapes
	// on the compositor canvas and blocking the non-MSAA blit-only fast
	// path (ADR-016). FillRectDirect writes directly to the CPU pixmap.
	for _, region := range regions {
		canvas.FillRectDirect(region.Bounds, bg)
	}

	// Clip to dirty union and draw tree.
	// Widgets outside the clip early-exit from isVisible checks,
	// so only widgets overlapping dirty regions actually render.
	canvas.PushClip(union)
	w.lastDrawStats = widget.DrawTree(w.root, w.ctx, canvas)
	canvas.PopClip()

	w.lastDirtyUnion = union
}

// ThemeBackground returns the window background color from the current theme.
// Falls back to white if no theme is configured.
func (w *Window) ThemeBackground() widget.Color {
	if w.theme != nil {
		return w.theme.Colors.Background
	}
	return widget.ColorWhite
}

// syncCursor forwards the cursor state to the platform provider.
func (w *Window) syncCursor() {
	if w.pp == nil {
		return
	}

	cursor := w.ctx.Cursor()
	w.pp.SetCursor(widgetCursorToPlatform(cursor))
}

// updateScale reads the scale factor from the WindowProvider and
// updates the context.
func (w *Window) updateScale() {
	scale := float32(defaultScale)
	if w.wp != nil {
		scale = float32(w.wp.ScaleFactor())
	}
	w.ctx.SetScale(scale)
}

// updateWindowSize reads the window size from the WindowProvider.
func (w *Window) updateWindowSize() {
	if w.wp != nil {
		width, height := w.wp.Size()
		newSize := geometry.Sz(float32(width), float32(height))
		if newSize != w.windowSize {
			w.windowSize = newSize
			w.needsLayout = true
		}
	} else if w.windowSize.Width == 0 && w.windowSize.Height == 0 {
		// Default size for headless mode.
		w.windowSize = geometry.Sz(defaultWidth, defaultHeight)
	}
}

// Overlays returns the window's overlay stack.
func (w *Window) Overlays() *overlay.Stack {
	return w.overlays
}

// FocusManager returns the window's focus manager.
//
// The focus manager handles Tab/Shift+Tab navigation between focusable
// widgets and supports registering global keyboard shortcuts.
func (w *Window) FocusManager() *ifocus.Manager {
	return w.focusMgr
}

// windowOverlayManager adapts the Window's overlay.Stack to the
// widget.OverlayManager interface. This avoids circular imports since
// the widget package cannot import the overlay package.
type windowOverlayManager struct {
	window *Window
}

// PushOverlay wraps the widget in an overlay.Container and pushes it.
// The content widget is promoted to a RepaintBoundary via SetRepaintBoundary(true)
// (ADR-024 WidgetBase property). No wrapper widget created — the content widget
// itself becomes the boundary. Clean overlays = texture blit (zero re-render).
// Dirty overlays (hover) = re-render only content texture.
func (m *windowOverlayManager) PushOverlay(w widget.Widget, onDismiss func()) {
	// ADR-024 + ADR-029: promote content to RepaintBoundary for damage isolation.
	type boundarySetter interface{ SetRepaintBoundary(bool) }
	if bs, ok := w.(boundarySetter); ok {
		bs.SetRepaintBoundary(true)
	}
	container := overlay.NewContainer(w, m.window.windowSize,
		overlay.WithOnDismiss(func() {
			if onDismiss != nil {
				onDismiss()
			}
		}),
	)
	m.window.overlays.Push(container)
}

// PopOverlay removes the topmost overlay.
func (m *windowOverlayManager) PopOverlay() {
	m.window.overlays.Pop()
}

// RemoveOverlay finds and removes the overlay containing the given widget.
func (m *windowOverlayManager) RemoveOverlay(w widget.Widget) {
	for _, o := range m.window.overlays.List() {
		if c, ok := o.(*overlay.Container); ok {
			if c.Content() == w {
				m.window.overlays.Remove(o)
				return
			}
		}
	}
}

// Compile-time check.
var _ widget.OverlayManager = (*windowOverlayManager)(nil)

// updateHover performs hit-testing to find the widget under the mouse and
// sends MouseEnter/MouseLeave events to individual widgets as the mouse
// moves across the widget tree.
//
// When overlays are open (dropdowns, dialogs), hover is directed to the
// overlay stack first. The topmost overlay's content widget tree receives
// hover hit-testing. If the mouse is outside the overlay content, hover is
// blocked from reaching background widgets (Flutter ModalBarrier pattern).
// This prevents background ListView items from highlighting while a
// dropdown menu is open on top of them.
//
// This uses ScreenBounds (computed during the Draw pass) for correct
// coordinate mapping, which accounts for scroll offsets, box positions,
// and all parent transforms.
func (w *Window) updateHover(pos geometry.Point, buttons event.ButtonState, mods event.Modifiers) {
	target := w.overlayAwareHitTest(pos)
	if target == w.hoveredWidget {
		return
	}

	// Hover target changed — reset cursor to Default.
	// Old widget's MouseLeave will set Default, new widget's MouseEnter
	// will set Pointer if it's interactive.
	w.ctx.ResetCursor()

	// Send MouseLeave to the old hovered widget.
	if w.hoveredWidget != nil {
		leave := event.NewMouseEvent(
			event.MouseLeave,
			event.ButtonNone,
			buttons,
			pos, pos,
			mods,
		)
		_ = w.hoveredWidget.Event(w.ctx, leave)
	}

	// Send MouseEnter to the new hovered widget.
	if target != nil {
		enter := event.NewMouseEvent(
			event.MouseEnter,
			event.ButtonNone,
			buttons,
			pos, pos,
			mods,
		)
		_ = target.Event(w.ctx, enter)
	}

	w.hoveredWidget = target
}

// clearHover sends MouseLeave to the currently hovered widget and clears
// the hover state. Called when the mouse leaves the window entirely.
func (w *Window) clearHover(buttons event.ButtonState, mods event.Modifiers) {
	if w.hoveredWidget == nil {
		return
	}

	leave := event.NewMouseEvent(
		event.MouseLeave,
		event.ButtonNone,
		buttons,
		geometry.Point{}, geometry.Point{},
		mods,
	)
	_ = w.hoveredWidget.Event(w.ctx, leave)
	w.hoveredWidget = nil
}

// HoveredWidget returns the widget currently under the mouse pointer,
// or nil if no widget is hovered.
func (w *Window) HoveredWidget() widget.Widget {
	return w.hoveredWidget
}

// overlayAwareHitTest performs hit-testing that respects the overlay stack.
//
// When overlays are open, the topmost overlay's widget tree is tested first.
// If a widget inside the overlay content matches, it is returned. If no
// overlay widget matches (mouse outside overlay content), nil is returned
// to block hover from reaching background widgets. This is the Flutter
// ModalBarrier pattern: overlays absorb hover to prevent background
// interaction while a dropdown or dialog is open.
//
// When no overlays are open, falls through to normal root tree hit-testing.
func (w *Window) overlayAwareHitTest(pos geometry.Point) widget.Widget {
	if w.overlays != nil && w.overlays.Len() > 0 {
		// Walk overlays top-to-bottom (highest z-order first).
		overlayList := w.overlays.List()
		for i := len(overlayList) - 1; i >= 0; i-- {
			o := overlayList[i]
			// Hit-test the overlay widget tree (Container + content).
			if hit := hitTest(o, pos); hit != nil {
				// Ignore hits on the Container itself (full-window backdrop).
				// Only return hits on actual content widgets inside the overlay.
				if hit == o {
					continue
				}
				return hit
			}
		}
		// Overlays are open but mouse is not over any overlay content.
		// Block hover from reaching background widgets.
		return nil
	}
	return hitTest(w.root, pos)
}

// hitTest walks the widget tree depth-first and returns the deepest
// visible widget whose ScreenBounds contains the given position.
//
// Children are checked in reverse order (top-most first in z-order)
// so that overlapping widgets receive events correctly.
// Returns nil if no widget contains the point.
func hitTest(root widget.Widget, pos geometry.Point) widget.Widget {
	if root == nil {
		return nil
	}
	return hitTestRecursive(root, pos)
}

// hitTestRecursive performs the recursive depth-first search.
func hitTestRecursive(w widget.Widget, pos geometry.Point) widget.Widget {
	// Check visibility — invisible widgets don't receive hover events.
	if base, ok := w.(interface{ IsVisible() bool }); ok && !base.IsVisible() {
		return nil
	}

	// Check if this widget's ScreenBounds contains the position.
	if sb, ok := w.(interface{ ScreenBounds() geometry.Rect }); ok {
		bounds := sb.ScreenBounds()
		if !bounds.Contains(pos) {
			return nil
		}
	}

	// Check children in reverse order (topmost first).
	children := w.Children()
	for i := len(children) - 1; i >= 0; i-- {
		child := children[i]
		if hit := hitTestRecursive(child, pos); hit != nil {
			return hit
		}
	}

	// No child hit — this widget itself is the target.
	return w
}

// widgetCursorToPlatform converts a widget.CursorType to gpucontext.CursorShape.
func widgetCursorToPlatform(c widget.CursorType) gpucontext.CursorShape {
	switch c {
	case widget.CursorDefault:
		return gpucontext.CursorDefault
	case widget.CursorPointer:
		return gpucontext.CursorPointer
	case widget.CursorText:
		return gpucontext.CursorText
	case widget.CursorCrosshair:
		return gpucontext.CursorCrosshair
	case widget.CursorMove:
		return gpucontext.CursorMove
	case widget.CursorResizeNS:
		return gpucontext.CursorResizeNS
	case widget.CursorResizeEW:
		return gpucontext.CursorResizeEW
	case widget.CursorResizeNESW:
		return gpucontext.CursorResizeNESW
	case widget.CursorResizeNWSE:
		return gpucontext.CursorResizeNWSE
	case widget.CursorNotAllowed:
		return gpucontext.CursorNotAllowed
	case widget.CursorWait:
		return gpucontext.CursorWait
	case widget.CursorNone:
		return gpucontext.CursorNone
	default:
		return gpucontext.CursorDefault
	}
}

// --- Dirty Boundary Management (ADR-007, Task 1e) ---

// AddDirtyBoundary registers a RepaintBoundary as dirty. Called by the
// onBoundaryDirty callback during upward dirty propagation.
//
// The key parameter is the boundary's unique cache key for deduplication.
// If the boundary is already in the set, this is a no-op (O(1) guard).
//
// This populates the flat dirty boundary set used by HasDirtyBoundaries
// for O(1) frame skip decisions, replacing O(n) NeedsRedrawInTreeNonBoundary.
func (w *Window) AddDirtyBoundary(key uint64) {
	if w.dirtyBoundaries == nil {
		w.dirtyBoundaries = make(map[uint64]dirtyBoundaryEntry)
	}
	w.dirtyBoundaries[key] = dirtyBoundaryEntry{present: true}
}

// HasDirtyBoundaries reports whether any RepaintBoundary has been marked
// dirty since the last paint pass.
func (w *Window) HasDirtyBoundaries() bool {
	return len(w.dirtyBoundaries) > 0
}

// DirtyBoundaryCount returns the number of dirty RepaintBoundary instances.
func (w *Window) DirtyBoundaryCount() int {
	return len(w.dirtyBoundaries)
}

// ClearDirtyBoundaries resets the dirty boundary set after painting.
// Each boundary's ClearBoundaryDirty is NOT called here — that is the
// responsibility of the PaintDirtyBoundaries method.
func (w *Window) ClearDirtyBoundaries() {
	// Clear map efficiently: delete all entries but keep the allocated map.
	for k := range w.dirtyBoundaries {
		delete(w.dirtyBoundaries, k)
	}
}

// PaintDirtyBoundaries clears the dirty boundary set after a frame.
//
// RepaintBoundary.Draw() handles cache invalidation internally: when
// boundaryDirty is true, it re-records child.Draw() into its scene.Scene.
// The full DrawTree pass triggers re-recording automatically — dirty
// boundaries re-record, clean ones replay cached scenes via ReplayScene.
//
// This is the Flutter flushPaint pattern: only dirty RepaintBoundary nodes
// re-record, clean ones replay cached scenes.
func (w *Window) PaintDirtyBoundaries() {
	w.ClearDirtyBoundaries()
}

// CollectDirtyRegions runs the dirty collector on the widget tree to populate
// the dirty tracker. Called by the compositor path in desktop.draw before
// painting, so debug overlays and damage rects have correct data.
//
// In the DrawTo path, this is called internally at the start of DrawTo.
// The compositor path must call it explicitly since it bypasses DrawTo.
func (w *Window) CollectDirtyRegions() {
	if w.root == nil {
		return
	}
	w.dirtyTracker.Reset()
	w.dirtyCollector.Collect(w.root)
	// ADR-029: also collect dirty regions from overlay content widgets.
	// Overlays are NOT in the root tree — without this, dirty overlay
	// widgets (hover on dropdown menu items) are invisible to both
	// cyan debug overlay (GOGPU_DEBUG_DIRTY) and green debug overlay
	// (GOGPU_DEBUG_DAMAGE via TrackDamageRect from prePaintDirtyRegions).
	// Use OverlayContentWidgets (not overlays.List) to reach the actual
	// content widgets directly, bypassing Container which may not expose
	// children through the standard Children() interface.
	for _, cw := range w.OverlayContentWidgets() {
		w.dirtyCollector.Collect(cw)
	}
	w.dirtyTracker.Optimize()
}

// ClearAfterPaint clears dirty flags and frame state after a paint pass.
// Called by the compositor path in desktop.draw after PaintBoundaryLayers
// and overlay drawing are complete.
//
// Flutter equivalent: flags are cleared at the end of flushPaint and
// after compositeFrame. We consolidate cleanup into one call.
func (w *Window) ClearAfterPaint() {
	// Do NOT call ClearRedrawInTree here. The paint pass (recordBoundary)
	// clears dirty flags BEFORE each boundary's Draw, so widgets that
	// re-dirty during Draw (spinner animation) keep their needsRedraw=true.
	// ClearRedrawInTree here would erase that re-dirty → spinner not found
	// by CollectDirtyRegions next frame → cyan overlay empty.
	w.needsRedraw = false
	w.needsFullRepaint = false
}

// NeedsAnimationFrame reports whether an animated boundary requested
// a frame via ScheduleAnimationFrame. This flag persists across
// ClearAfterPaint (unlike needsRedraw) to prevent frame skip from
// dropping animation frames. Flutter equivalent: _hasScheduledFrame.
func (w *Window) NeedsAnimationFrame() bool {
	return w.needsAnimationFrame
}

// ClearAnimationFrame resets the animation frame flag. Called by
// desktop.draw AFTER the frame skip check passes, ensuring the
// flag is consumed exactly once per frame.
func (w *Window) ClearAnimationFrame() {
	w.needsAnimationFrame = false
}

// HasOverlays reports whether any overlays (dropdowns, dialogs) are active.
func (w *Window) HasOverlays() bool {
	return w.overlays != nil && w.overlays.Len() > 0
}

// OverlayCount returns the number of active overlays.
func (w *Window) OverlayCount() int {
	if w.overlays == nil {
		return 0
	}
	return w.overlays.Len()
}

// HasDirtyOverlays reports whether any overlay widget has NeedsRedraw=true.
// Used to selectively enable damage tracking during DrawOverlays — unchanged
// overlays suppress tracking (avoid permanent green debug overlay), while
// changed overlays (hover) enable tracking for correct green flash.
func (w *Window) HasDirtyOverlays() bool {
	if w.overlays == nil || w.overlays.Len() == 0 {
		return false
	}
	for _, o := range w.overlays.List() {
		if widget.NeedsRedrawInTree(o) {
			return true
		}
	}
	return false
}

// ClearOverlayRedraw clears NeedsRedraw on all overlay widgets after
// drawing with damage tracking enabled.
func (w *Window) ClearOverlayRedraw() {
	if w.overlays == nil {
		return
	}
	for _, o := range w.overlays.List() {
		widget.ClearRedrawInTree(o)
	}
}

// DirtyOverlayContentRects returns the screen bounds of overlay CONTENT widgets
// (not the full-window Container backdrop) that have NeedsRedraw=true.
//
// This enables granular damage tracking for overlays. Without this, the Container
// backdrop (full-window scrim for modal overlays) registers the entire window as
// damage, causing GOGPU_DEBUG_DAMAGE green overlay on the full screen.
//
// ADR-029: retained-mode overlays. Flutter pattern: ModalBarrier is event-only
// (no draw contribution to damage), overlay content is in its own RepaintBoundary.
// Our equivalent: suppress damage for Container backdrop, track only content rects.
//
// For overlay types that implement ContentProvider (Container), the content widget's
// bounds are returned. For other overlay types, the overlay's own bounds are used.
func (w *Window) DirtyOverlayContentRects() []geometry.Rect {
	if w.overlays == nil || w.overlays.Len() == 0 {
		return nil
	}

	var rects []geometry.Rect
	for _, o := range w.overlays.List() {
		if !widget.NeedsRedrawInTree(o) {
			continue
		}

		// Try to get the content widget's bounds (Container pattern).
		// Content bounds are tighter than Container bounds (full window).
		type contentProvider interface {
			Content() widget.Widget
		}
		if cp, ok := o.(contentProvider); ok {
			content := cp.Content()
			if content != nil {
				type bounder interface{ Bounds() geometry.Rect }
				if b, ok2 := content.(bounder); ok2 {
					rects = append(rects, b.Bounds())
					continue
				}
			}
		}

		// Fallback: use overlay's own bounds (non-Container overlays).
		type bounder interface{ Bounds() geometry.Rect }
		if b, ok := o.(bounder); ok {
			rects = append(rects, b.Bounds())
		}
	}
	return rects
}

// DrawOverlays draws overlay widgets (dropdowns, dialogs) on the given canvas.
// In Flutter, overlays are part of the same widget tree. In our architecture,
// they are managed separately by overlay.Stack and drawn after the main scene.
func (w *Window) DrawOverlays(canvas widget.Canvas) {
	w.overlays.Draw(w.ctx, canvas)
}

// OverlayContentWidgets returns the content widgets from all active overlays.
// For Container overlays, this returns the inner content widget (which is
// marked as RepaintBoundary by PushOverlay). For non-Container overlays,
// the overlay itself is returned.
//
// Used by the compositor pipeline to include overlay boundaries in the
// Layer Tree alongside the main widget tree boundaries.
func (w *Window) OverlayContentWidgets() []widget.Widget {
	if w.overlays == nil || w.overlays.Len() == 0 {
		return nil
	}
	result := make([]widget.Widget, 0, w.overlays.Len())
	for _, o := range w.overlays.List() {
		type contentProvider interface {
			Content() widget.Widget
		}
		if cp, ok := o.(contentProvider); ok {
			content := cp.Content()
			if content != nil {
				result = append(result, content)
				continue
			}
		}
		// Fallback: non-Container overlay is its own widget.
		result = append(result, o)
	}
	return result
}

// DrawOverlayScrim draws only the modal backdrop scrim for overlay Containers.
// Non-modal overlays have no scrim. This is the minimal immediate-mode part
// that remains after overlay content moves to the boundary pipeline.
//
// Flutter equivalent: ModalBarrier.build() draws a full-screen gesture detector
// with optional color. The barrier is event-only in damage terms — no paint
// contribution. Our scrim draws a semi-transparent rect for visual feedback.
func (w *Window) DrawOverlayScrim(canvas widget.Canvas) {
	if w.overlays == nil || w.overlays.Len() == 0 || canvas == nil {
		return
	}
	for _, o := range w.overlays.List() {
		type modalChecker interface {
			Modal() bool
			Bounds() geometry.Rect
		}
		mc, ok := o.(modalChecker)
		if !ok || !mc.Modal() {
			continue
		}
		scrim := widget.RGBA(0, 0, 0, 0.32)
		canvas.DrawRect(mc.Bounds(), scrim)
	}
}

// HasDirtyBoundariesOrNeedsRedraw reports whether any rendering work is
// needed: either dirty boundaries from upward propagation or full-frame
// redraw flags (needsRedraw, needsFullRepaint).
func (w *Window) HasDirtyBoundariesOrNeedsRedraw() bool {
	return w.HasDirtyBoundaries() || w.needsRedraw || w.needsFullRepaint
}

// animPumper pumps frames at a configurable rate for smooth animation.
// Default 30fps (33ms) — sufficient for spinners and progress indicators.
// GPU cost scales linearly with frame rate (~0.17%/frame on Intel Iris Xe).
// 60fps for high-fidelity animations (transitions, physics).
// Stopped when animation completes (3 consecutive idle frames).
type animPumper struct {
	stop chan struct{}
}

// defaultAnimPumpInterval controls the animation frame pump rate.
// 33ms ≈ 30fps — visually smooth for indeterminate spinners and progress
// indicators. Saves ~50% GPU vs 60fps with no perceptible quality loss.
// Enterprise reference: Qt uses QTimer intervals, Ebiten uses SetTPS.
const defaultAnimPumpInterval = 33 * time.Millisecond // 30fps

func newAnimPumper(wp gpucontext.WindowProvider) *animPumper {
	return newAnimPumperWithInterval(wp, defaultAnimPumpInterval)
}

func newAnimPumperWithInterval(wp gpucontext.WindowProvider, interval time.Duration) *animPumper {
	p := &animPumper{stop: make(chan struct{})}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-p.stop:
				return
			case <-ticker.C:
				wp.RequestRedraw()
			}
		}
	}()
	return p
}

func (p *animPumper) Stop() {
	select {
	case p.stop <- struct{}{}:
	default:
	}
}
