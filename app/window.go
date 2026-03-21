package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/gogpu/gpucontext"
	ui "github.com/gogpu/ui"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	ifocus "github.com/gogpu/ui/internal/focus"
	"github.com/gogpu/ui/overlay"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/theme"
	"github.com/gogpu/ui/widget"
)

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

	// animToken is non-nil while continuous rendering is active for animations.
	animToken animationStopper

	// animIdleFrames counts consecutive frames with no Invalidate call.
	// Used to stop the animation pumper after animations complete.
	animIdleFrames int

	// needsLayout indicates that layout should be recalculated.
	needsLayout bool

	// needsRedraw indicates that the draw pass should be performed.
	// This is set when any widget in the tree needs re-rendering,
	// and cleared after a successful draw. When false, DrawTo can
	// skip rendering entirely because the previous frame's output
	// is still valid in the GPU framebuffer.
	needsRedraw bool

	// lastDrawStats holds per-widget statistics from the most recent
	// draw traversal. Updated by DrawTo() and the headless draw() path.
	lastDrawStats widget.DrawStats

	// hoveredWidget tracks the widget currently under the mouse pointer.
	// Used for sending MouseEnter/MouseLeave events to individual widgets
	// as the mouse moves across the widget tree.
	hoveredWidget widget.Widget

	// mouseButtonsHeld tracks pressed mouse buttons to prevent cursor
	// reset during drag operations (Frame.ResetCursor skipped while dragging).
	mouseButtonsHeld event.ButtonState

	// windowSize tracks the last known window size in physical pixels.
	windowSize geometry.Size

	// frameCallback, if set, is called after each frame with statistics.
	frameCallback FrameCallback
}

// newWindow creates a Window with the given providers.
func newWindow(
	wp gpucontext.WindowProvider,
	pp gpucontext.PlatformProvider,
	scheduler *state.Scheduler,
	t *theme.Theme,
) *Window {
	ctx := widget.NewContext()

	w := &Window{
		ctx:         ctx,
		wp:          wp,
		pp:          pp,
		scheduler:   scheduler,
		theme:       t,
		focusMgr:    ifocus.New(nil),
		needsLayout: true,
	}

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
	ctx.SetOnInvalidate(func() {
		w.needsLayout = true
		w.needsRedraw = true
		if w.root != nil {
			widget.MarkRedrawInTree(w.root)
		}
		if w.wp != nil {
			w.wp.RequestRedraw()
		}
	})

	// Wire scheduler to wake render loop when signals change.
	if wp != nil {
		scheduler.SetOnDirty(func() {
			w.needsLayout = true
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
	// Unmount old tree.
	if w.root != nil {
		widget.UnmountTree(w.root)
	}

	w.root = root
	w.needsLayout = true
	w.needsRedraw = true

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
	// In event-driven mode (ContinuousRender=false), Frame() only
	// runs when a redraw is needed, but cursor changes from hover
	// don't trigger redraws.
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

	// Reset cursor for this frame — but not during drag operations.
	// During drag, the dragging widget (SplitView, Slider) sets cursor
	// on every MouseMove; resetting here would flash default between frames.
	if w.mouseButtonsHeld == 0 {
		w.ctx.ResetCursor()
	}

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
		layoutStart := time.Now()
		w.layout()
		layoutDur = time.Since(layoutStart)
		// Clear needsLayout only if no widget re-invalidated during layout.
		// Animations call ctx.Invalidate() from tickAnimation() during layout,
		// which sets needsLayout back to true — we must not clobber that.
		if !w.ctx.IsInvalidated() {
			w.needsLayout = false
		}
		// Layout changes require full redraw since positions may have shifted.
		w.needsRedraw = true
		widget.MarkRedrawInTree(w.root)
	}

	// Determine if any widget in the tree needs redraw.
	// This check is O(n) in the worst case but short-circuits on first dirty widget.
	if !w.needsRedraw {
		w.needsRedraw = widget.NeedsRedrawInTree(w.root)
	}

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
	if w.ctx.IsInvalidated() {
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

// NeedsRedraw reports whether any widget in the tree needs re-rendering.
//
// When this returns false, the host application can skip calling [Window.DrawTo]
// and reuse the previous frame's output from the GPU framebuffer. This is the
// primary optimization of retained-mode rendering: idle UIs consume zero CPU
// for rendering.
func (w *Window) NeedsRedraw() bool {
	if w.needsRedraw {
		return true
	}
	return widget.NeedsRedrawInTree(w.root)
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

// DrawTo performs the draw pass using the provided canvas.
//
// This method is called by the host application when it has a canvas
// ready for drawing. It draws the root widget tree and then all overlays
// on top.
//
// With retained-mode rendering, DrawTo checks whether any widget in the
// tree actually needs re-rendering. If not, it returns false to indicate
// that the host application can skip uploading the canvas to the GPU
// (the previous frame's output is still valid).
//
// Returns true if rendering was performed, false if the draw was skipped
// because no widgets changed since the last draw.
func (w *Window) DrawTo(canvas widget.Canvas) bool {
	if w.root == nil || canvas == nil {
		return false
	}

	// Collect dirty/clean statistics before drawing.
	// In Sub-Phase 1 we always draw (existing widgets don't self-dirty on
	// event-driven state changes yet). The stats provide observability for
	// monitoring and for validating the dirty-tracking system.
	//
	// Sub-Phase 2 will add per-widget self-dirtying + pixel caching, at which
	// point clean subtrees can be composited from cache instead of re-drawn.
	treeClean := !w.needsRedraw && !widget.NeedsRedrawInTree(w.root)

	// Draw the widget tree, collecting per-widget statistics.
	w.lastDrawStats = widget.DrawTree(w.root, w.ctx, canvas)

	// Draw overlays on top (bottom to top).
	w.overlays.Draw(w.ctx, canvas)

	// Clear all redraw flags in the tree after successful draw.
	widget.ClearRedrawInTree(w.root)
	w.needsRedraw = false

	if treeClean {
		ui.Logger().LogAttrs(context.Background(), slog.LevelDebug, "DrawTo: tree was clean (could skip in Phase 2)")
	} else {
		ui.Logger().LogAttrs(context.Background(), slog.LevelDebug, "DrawTo: rendered",
			slog.Int("dirty", w.lastDrawStats.DirtyWidgets),
			slog.Int("clean", w.lastDrawStats.CleanWidgets),
			slog.Int("cached", w.lastDrawStats.CachedWidgets),
			slog.Int("total", w.lastDrawStats.TotalWidgets),
		)
	}

	return !treeClean
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
func (m *windowOverlayManager) PushOverlay(w widget.Widget, onDismiss func()) {
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
// This uses ScreenBounds (computed during the Draw pass) for correct
// coordinate mapping, which accounts for scroll offsets, box positions,
// and all parent transforms.
func (w *Window) updateHover(pos geometry.Point, buttons event.ButtonState, mods event.Modifiers) {
	target := hitTest(w.root, pos)
	if target == w.hoveredWidget {
		return
	}

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

// animPumper pumps frames at ~60fps for smooth animation.
// Stopped when animation completes.
type animPumper struct {
	stop chan struct{}
}

func newAnimPumper(wp gpucontext.WindowProvider) *animPumper {
	p := &animPumper{stop: make(chan struct{})}
	go func() {
		ticker := time.NewTicker(16 * time.Millisecond) // ~60fps
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
