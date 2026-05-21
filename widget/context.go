package widget

import (
	"image"
	"sync"
	"time"

	"github.com/gogpu/ui/geometry"
)

// Context provides access to UI state during layout, drawing, and event handling.
//
// Context is passed through the widget tree during all phases (Layout, Draw, Event).
// It provides:
//   - Focus management: Request/release focus, query focused widget
//   - Time information: Current time and delta for animations
//   - Invalidation: Mark areas as needing redraw
//   - Cursor management: Change the mouse cursor
//   - Theme access: Query the current visual theme
//
// Thread Safety:
//
// Context implementations must be safe for concurrent access. The default
// implementation [ContextImpl] uses a mutex to protect internal state.
//
// Example:
//
//	func (w *MyWidget) Event(ctx widget.Context, e event.Event) bool {
//	    if clicked {
//	        ctx.RequestFocus(w)
//	        ctx.Invalidate()
//	        return true
//	    }
//	    return false
//	}
type Context interface {
	// RequestFocus requests focus for the given widget.
	//
	// If another widget currently has focus, it will receive a focus lost event.
	// The widget parameter should implement the Widget interface.
	RequestFocus(w Widget)

	// ReleaseFocus releases focus from the given widget.
	//
	// If the widget doesn't have focus, this is a no-op.
	// After calling this, FocusedWidget() will return nil.
	ReleaseFocus(w Widget)

	// IsFocused returns true if the given widget currently has focus.
	IsFocused(w Widget) bool

	// FocusedWidget returns the currently focused widget, or nil if none.
	FocusedWidget() Widget

	// Now returns the current time.
	//
	// This is the time at the start of the current frame/event cycle.
	// Use this for animations and time-based effects.
	Now() time.Time

	// DeltaTime returns the time elapsed since the previous frame.
	//
	// This is useful for smooth animations that should be frame-rate independent.
	// Returns 0 for the first frame.
	DeltaTime() time.Duration

	// Invalidate marks the entire window as needing a redraw.
	//
	// Call this when widget state changes require visual updates.
	// Multiple calls per frame are coalesced into a single redraw.
	Invalidate()

	// InvalidateRect marks a specific rectangular area as needing a redraw.
	//
	// Use this for more efficient partial redraws when only a small
	// part of the UI has changed.
	InvalidateRect(r geometry.Rect)

	// Cursor returns the current cursor type.
	Cursor() CursorType

	// SetCursor changes the mouse cursor.
	//
	// The cursor is typically reset to CursorDefault at the start of each frame.
	SetCursor(cursor CursorType)

	// Scale returns the display scale factor (DPI scaling).
	//
	// Returns 1.0 for standard displays, 2.0 for Retina/HiDPI displays, etc.
	// Use this to scale coordinates and sizes for proper rendering.
	Scale() float32

	// ThemeProvider returns the current theme for this context.
	//
	// Returns nil if no theme is set (headless mode without a theme).
	// Widgets should check for nil before using the returned provider.
	ThemeProvider() ThemeProvider

	// OverlayManager returns the overlay manager for pushing/removing overlays.
	//
	// Returns nil if no overlay manager is set (headless mode without a window).
	// Widgets should check for nil before calling overlay methods.
	OverlayManager() OverlayManager

	// WindowSize returns the current window size in logical pixels.
	WindowSize() geometry.Size

	// Scheduler returns the signal scheduler for this context.
	//
	// Returns nil if no scheduler is set (headless mode without signal support).
	// Widgets should check for nil before using the returned scheduler.
	Scheduler() SchedulerRef
}

// SchedulerRef is a minimal interface for the signal scheduler.
// It is defined in the widget package to avoid circular imports
// between widget and state packages.
type SchedulerRef interface {
	MarkDirty(w Widget)
}

// OverlayManager provides methods for pushing and removing overlays from the
// window's overlay stack. This interface lives in the widget package to avoid
// circular imports: the overlay package imports widget, so widget cannot
// import overlay. Instead, widgets call OverlayManager methods on the Context
// without needing to know the concrete overlay.Stack type.
type OverlayManager interface {
	// PushOverlay pushes a widget as an overlay. The onDismiss callback is
	// called when the overlay should be closed (e.g. click outside, Escape key).
	PushOverlay(w Widget, onDismiss func())

	// PopOverlay removes the topmost overlay from the stack.
	PopOverlay()

	// RemoveOverlay removes a specific overlay widget from the stack.
	RemoveOverlay(w Widget)
}

// DirtyTrackerRef is a minimal interface for querying spatial dirty regions
// during the draw pass. It is defined in the widget package (rather than
// importing internal/dirty) so that widgets like RepaintBoundary can check
// whether their bounds intersect any dirty region without depending on
// internal packages.
//
// The internal dirty.Tracker satisfies this interface through structural
// typing — no explicit implementation is needed.
type DirtyTrackerRef interface {
	// Intersects returns true if the given bounds intersect any dirty region.
	// Used by RepaintBoundary for O(regions) early exit: when bounds don't
	// overlap any dirty region, the subtree is guaranteed clean and the
	// expensive O(tree_depth) NeedsRedrawInTree walk can be skipped.
	Intersects(bounds geometry.Rect) bool
}

// DirtyTrackerProvider is an optional interface implemented by Context
// implementations that provide access to the frame's dirty region tracker.
//
// During a draw pass, widgets like RepaintBoundary type-assert the Context
// to DirtyTrackerProvider to query spatial dirty regions. This uses the
// established "interface extension via type assertion" pattern (same as
// DrawStatsProvider, Focusable) to avoid adding methods to the Context
// interface (which would be a breaking change).
//
// Example usage in a widget's Draw method:
//
//	if provider, ok := ctx.(widget.DirtyTrackerProvider); ok {
//	    if tracker := provider.DirtyTracker(); tracker != nil {
//	        if !tracker.Intersects(rb.Bounds()) {
//	            // Bounds don't overlap any dirty region — guaranteed clean.
//	            return
//	        }
//	    }
//	}
type DirtyTrackerProvider interface {
	DirtyTracker() DirtyTrackerRef
}

// DirtyBoundaryRegistrar is an optional interface implemented by Context
// implementations that support O(1) flat dirty boundary tracking.
//
// During upward dirty propagation, when a RepaintBoundary's onBoundaryDirty
// callback fires, it type-asserts the Context to DirtyBoundaryRegistrar
// and registers the boundary in the Window's flat dirty set. This replaces
// O(n) NeedsRedrawInTreeNonBoundary tree walks with O(1) map lookup.
//
// This is the Flutter _nodesNeedingPaint pattern: a flat list of dirty
// RenderObjects, populated during markNeedsPaint, consumed during flushPaint.
//
// Example usage in onBoundaryDirty callback:
//
//	if reg, ok := ctx.(widget.DirtyBoundaryRegistrar); ok {
//	    reg.RegisterDirtyBoundary(key)
//	}
type DirtyBoundaryRegistrar interface {
	RegisterDirtyBoundary(key uint64)
}

// ImageCacheRef is a minimal interface for a centralized RepaintBoundary
// pixel cache with LRU eviction. It is defined in the widget package so that
// primitives/repaint_boundary.go can use the cache without importing
// internal/render (which would be an import cycle).
//
// The internal render.ImageCache satisfies this interface through structural
// typing — no explicit implementation is needed.
type ImageCacheRef interface {
	// Get retrieves a cached image by key. Returns the image and true if
	// found, nil and false otherwise. On hit, the entry is promoted in LRU.
	Get(key uint64) (*image.RGBA, bool)

	// Put stores an image in the cache with the given key and version.
	// Evicts LRU entries if the memory budget is exceeded.
	Put(key uint64, img *image.RGBA, version uint64)

	// Invalidate removes a specific entry from the cache by key.
	Invalidate(key uint64)
}

// ImageCacheProvider is an optional interface implemented by Context
// implementations that provide a centralized RepaintBoundary pixel cache.
//
// During a draw pass, RepaintBoundary type-asserts the Context to
// ImageCacheProvider to access the shared LRU cache. This uses the
// established "interface extension via type assertion" pattern (same as
// DrawStatsProvider, DirtyTrackerProvider) to avoid adding methods to
// the Context interface (which would be a breaking change).
//
// When no ImageCacheProvider is available (e.g., in headless tests),
// RepaintBoundary falls back to its local per-widget cache field.
//
// Example usage in a widget's Draw method:
//
//	if provider, ok := ctx.(widget.ImageCacheProvider); ok {
//	    if cache := provider.ImageCache(); cache != nil {
//	        img, ok := cache.Get(rb.cacheKey)
//	        // ...
//	    }
//	}
type ImageCacheProvider interface {
	ImageCache() ImageCacheRef
}

// PointerCapturer is an optional interface implemented by Context
// implementations that support mouse capture during drag operations.
//
// When a widget captures the pointer, all subsequent mouse events (Move,
// Release, Press) are delivered directly to it, bypassing tree hit-testing.
// This is essential for drag operations (slider thumb, splitview divider,
// scrollbar thumb, text selection) where the mouse may move outside the
// widget's bounds during the gesture.
//
// The capture auto-releases on MouseRelease when all buttons are up,
// preventing stale captures. Widgets should also call ReleasePointer
// explicitly when the drag operation ends.
//
// Enterprise references:
//   - Win32: SetCapture(HWND) / ReleaseCapture()
//   - Qt: QWidget::grabMouse() / releaseMouse()
//   - Flutter: GestureArena implicit capture
//   - GTK4: gtk_gesture_set_state(CLAIMED)
//
// This uses the established "interface extension via type assertion" pattern
// (same as DirtyBoundaryRegistrar, AnimationScheduler, ImageCacheProvider)
// to avoid adding methods to the Context interface (which would be a breaking
// change for all implementors).
//
// Example usage in a widget's Event handler:
//
//	case event.MousePress:
//	    if cap, ok := ctx.(widget.PointerCapturer); ok {
//	        cap.CapturePointer(w)
//	    }
//	case event.MouseRelease:
//	    if cap, ok := ctx.(widget.PointerCapturer); ok {
//	        cap.ReleasePointer(w)
//	    }
type PointerCapturer interface {
	// CapturePointer requests exclusive mouse event delivery for the widget.
	// While captured, MouseMove, MouseRelease, and MousePress events are
	// delivered directly to this widget without tree hit-testing.
	// Only one widget can have capture at a time.
	CapturePointer(w Widget)

	// ReleasePointer releases the pointer capture. This is a no-op if the
	// given widget does not currently hold the capture.
	ReleasePointer(w Widget)
}

// CursorType represents the type of mouse cursor to display.
type CursorType uint8

// Cursor type constants.
const (
	// CursorDefault is the standard arrow cursor.
	CursorDefault CursorType = iota

	// CursorPointer is the pointing hand cursor, typically for links.
	CursorPointer

	// CursorText is the I-beam cursor for text selection.
	CursorText

	// CursorCrosshair is the crosshair cursor for precise selection.
	CursorCrosshair

	// CursorMove is the four-arrow move cursor.
	CursorMove

	// CursorResizeNS is the north-south (vertical) resize cursor.
	CursorResizeNS

	// CursorResizeEW is the east-west (horizontal) resize cursor.
	CursorResizeEW

	// CursorResizeNESW is the diagonal (northeast-southwest) resize cursor.
	CursorResizeNESW

	// CursorResizeNWSE is the diagonal (northwest-southeast) resize cursor.
	CursorResizeNWSE

	// CursorNotAllowed is the circle with a line through it (forbidden) cursor.
	CursorNotAllowed

	// CursorWait is the wait/busy cursor (hourglass or spinner).
	CursorWait

	// CursorNone hides the cursor.
	CursorNone
)

// String returns a human-readable name for the cursor type.
func (c CursorType) String() string {
	switch c {
	case CursorDefault:
		return "Default"
	case CursorPointer:
		return "Pointer"
	case CursorText:
		return "Text"
	case CursorCrosshair:
		return "Crosshair"
	case CursorMove:
		return "Move"
	case CursorResizeNS:
		return "ResizeNS"
	case CursorResizeEW:
		return "ResizeEW"
	case CursorResizeNESW:
		return "ResizeNESW"
	case CursorResizeNWSE:
		return "ResizeNWSE"
	case CursorNotAllowed:
		return "NotAllowed"
	case CursorWait:
		return "Wait"
	case CursorNone:
		return "None"
	default:
		return unknownStr
	}
}

// ContextImpl is the standard implementation of the Context interface.
//
// It provides thread-safe focus management, time tracking, and invalidation.
// Create a new ContextImpl with [NewContext].
//
// Example:
//
//	ctx := widget.NewContext()
//	ctx.SetNow(time.Now())
//	// Pass to widget tree during layout/draw/event
type ContextImpl struct {
	mu sync.RWMutex

	// Focus state
	focusedWidget Widget

	// Time tracking
	now       time.Time
	lastFrame time.Time
	deltaTime time.Duration

	// Invalidation
	invalidated    bool
	invalidateRect geometry.Rect

	// Cursor
	cursor CursorType

	// Display scale
	scale float32

	// Theme provider
	themeProvider ThemeProvider

	// Callback for invalidation (called when Invalidate is called)
	onInvalidate func()

	// Callback for invalidate rect (called when InvalidateRect is called)
	onInvalidateRect func(geometry.Rect)

	// Callback for animation frame scheduling (deferred, not immediate)
	onScheduleAnimation func()

	// Overlay manager
	overlayManager OverlayManager

	// Window size
	windowSize geometry.Size

	// Signal scheduler
	scheduler SchedulerRef

	// drawStats collects per-frame rendering statistics during the draw pass.
	// Set by DrawTree/drawWidgetsInRegion before drawing, read by
	// RepaintBoundary to increment CachedWidgets on cache hit.
	drawStats *DrawStats

	// dirtyTracker provides spatial dirty region queries during the draw pass.
	// Set by Window.DrawTo/drawIncremental before drawing, cleared after.
	// RepaintBoundary uses this for O(regions) fast path to avoid expensive
	// O(tree_depth) NeedsRedrawInTree checks when bounds don't intersect
	// any dirty region.
	dirtyTracker DirtyTrackerRef

	// imageCache provides a centralized LRU cache for RepaintBoundary pixel
	// buffers. When set, RepaintBoundary uses this shared cache instead of
	// per-widget local cache fields. This enables memory budget enforcement
	// and LRU eviction across all boundaries in a window.
	// Set by Window during initialization, cleared on close.
	imageCache ImageCacheRef

	// onRegisterDirtyBoundary is called when a RepaintBoundary transitions
	// from clean to dirty via upward propagation. The Window wires this
	// callback to AddDirtyBoundary during initialization, populating the
	// flat dirty boundary set for O(1) frame skip decisions.
	// This is the Flutter _nodesNeedingPaint.add() equivalent.
	onRegisterDirtyBoundary func(key uint64)

	// onCapturePointer is called when a widget requests pointer capture
	// during drag operations (ADR-031). The Window wires this callback
	// to set capturedWidget. When set, mouse events are delivered directly
	// to the captured widget, bypassing tree hit-testing.
	onCapturePointer func(w Widget)

	// onReleasePointer is called when a widget releases pointer capture.
	// The Window wires this callback to clear capturedWidget.
	onReleasePointer func(w Widget)
}

// NewContext creates a new ContextImpl with default settings.
//
// The context is initialized with:
//   - No focused widget
//   - Current time set to time.Now()
//   - Scale factor of 1.0
//   - Default cursor
func NewContext() *ContextImpl {
	now := time.Now()
	return &ContextImpl{
		now:       now,
		lastFrame: now,
		scale:     1.0,
		cursor:    CursorDefault,
	}
}

// RequestFocus requests focus for the given widget.
func (c *ContextImpl) RequestFocus(w Widget) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.focusedWidget == w {
		return // Already focused
	}

	// Clear focus from previous widget
	if c.focusedWidget != nil {
		if setter, ok := c.focusedWidget.(interface{ SetFocused(bool) }); ok {
			setter.SetFocused(false)
		}
	}

	// Set focus to new widget
	c.focusedWidget = w
	if w != nil {
		if setter, ok := w.(interface{ SetFocused(bool) }); ok {
			setter.SetFocused(true)
		}
	}
}

// ReleaseFocus releases focus from the given widget.
func (c *ContextImpl) ReleaseFocus(w Widget) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.focusedWidget != w {
		return // Widget doesn't have focus
	}

	if setter, ok := c.focusedWidget.(interface{ SetFocused(bool) }); ok {
		setter.SetFocused(false)
	}
	c.focusedWidget = nil
}

// IsFocused returns true if the given widget currently has focus.
func (c *ContextImpl) IsFocused(w Widget) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.focusedWidget == w
}

// FocusedWidget returns the currently focused widget, or nil if none.
func (c *ContextImpl) FocusedWidget() Widget {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.focusedWidget
}

// Now returns the current time.
func (c *ContextImpl) Now() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.now
}

// DeltaTime returns the time elapsed since the previous frame.
func (c *ContextImpl) DeltaTime() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.deltaTime
}

// Invalidate marks the entire window as needing a redraw.
func (c *ContextImpl) Invalidate() {
	c.mu.Lock()
	c.invalidated = true
	callback := c.onInvalidate
	c.mu.Unlock()

	if callback != nil {
		callback()
	}
}

// InvalidateRect marks a specific rectangular area as needing a redraw.
func (c *ContextImpl) InvalidateRect(r geometry.Rect) {
	c.mu.Lock()
	if c.invalidated {
		// Already doing a full invalidation, no need for partial
		c.mu.Unlock()
		return
	}
	if c.invalidateRect.IsEmpty() {
		c.invalidateRect = r
	} else {
		c.invalidateRect = c.invalidateRect.Union(r)
	}
	callback := c.onInvalidateRect
	c.mu.Unlock()

	if callback != nil {
		callback(r)
	}
}

// Cursor returns the current cursor type.
func (c *ContextImpl) Cursor() CursorType {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cursor
}

// SetCursor changes the mouse cursor.
func (c *ContextImpl) SetCursor(cursor CursorType) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cursor = cursor
}

// Scale returns the display scale factor.
func (c *ContextImpl) Scale() float32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.scale
}

// SetScale sets the display scale factor.
func (c *ContextImpl) SetScale(scale float32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.scale = scale
}

// ThemeProvider returns the current theme for this context.
//
// Returns nil if no theme is set (headless mode without a theme).
func (c *ContextImpl) ThemeProvider() ThemeProvider {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.themeProvider
}

// SetThemeProvider sets the theme provider for this context.
//
// Pass nil to clear the theme provider (e.g., for headless testing).
func (c *ContextImpl) SetThemeProvider(tp ThemeProvider) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.themeProvider = tp
}

// maxDeltaTime caps the delta to prevent animation jumps when the app
// resumes from background, after a debug pause, or on the first frame.
// This is standard practice in game engines and UI frameworks (Flutter
// clamps to 100ms, Qt handles this internally).
const maxDeltaTime = 100 * time.Millisecond

// SetNow updates the current time and calculates delta time.
//
// Call this at the start of each frame before processing events and layout.
// DeltaTime is clamped to [0, 100ms] to prevent animation jumps after
// long pauses (e.g., background/resume, debugger breakpoints).
func (c *ContextImpl) SetNow(now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = now
}

// BeginFrame updates the frame timing. DeltaTime is calculated from the
// previous BeginFrame call, not from the last SetNow. This ensures
// DeltaTime always reflects the inter-frame interval regardless of how
// many HandleEvent calls happen between frames. See issue #53.
func (c *ContextImpl) BeginFrame(now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	dt := now.Sub(c.lastFrame)
	if dt < 0 {
		dt = 0
	}
	if dt > maxDeltaTime {
		dt = maxDeltaTime
	}
	c.deltaTime = dt
	c.lastFrame = now
	c.now = now
}

// IsInvalidated returns true if the window needs a redraw.
func (c *ContextImpl) IsInvalidated() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.invalidated
}

// InvalidatedRect returns the area that needs redrawing.
//
// Returns an empty rect if no partial invalidation was requested,
// or if a full invalidation was requested (check IsInvalidated).
func (c *ContextImpl) InvalidatedRect() geometry.Rect {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.invalidateRect
}

// ClearInvalidation clears the invalidation state.
//
// Call this after processing a redraw.
func (c *ContextImpl) ClearInvalidation() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.invalidated = false
	c.invalidateRect = geometry.Rect{}
}

// ResetCursor resets the cursor to default.
//
// Call this at the start of each frame before processing events.
func (c *ContextImpl) ResetCursor() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cursor = CursorDefault
}

// SetOnInvalidate sets a callback function called when Invalidate is called.
func (c *ContextImpl) SetOnInvalidate(callback func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onInvalidate = callback
}

// SetOnInvalidateRect sets a callback function called when InvalidateRect is called.
func (c *ContextImpl) SetOnInvalidateRect(callback func(geometry.Rect)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onInvalidateRect = callback
}

// AnimationScheduler is an optional interface for deferred animation frame
// requests. Animated widgets (spinners, progress bars) use this instead of
// ctx.InvalidateRect() to avoid triggering immediate RequestRedraw.
//
// The framework's animation pumper controls the actual frame rate —
// animated widgets just request "paint me on the next animation tick",
// not "paint me RIGHT NOW".
//
// Flutter equivalent: SchedulerBinding.scheduleFrame() — defers to next
// vsync, does NOT trigger immediate render. Multiple calls coalesce.
// Qt equivalent: QTimer-driven update() — deferred to event loop.
// Android equivalent: Choreographer.postFrameCallback() — next vsync.
//
// Usage in animated widgets:
//
//	if sched, ok := ctx.(widget.AnimationScheduler); ok {
//	    sched.ScheduleAnimationFrame()
//	} else {
//	    ctx.InvalidateRect(w.Bounds()) // fallback: immediate
//	}
type AnimationScheduler interface {
	ScheduleAnimationFrame()
}

// ScheduleAnimationFrame requests that the render loop stay active for
// animation. Unlike InvalidateRect, this does NOT trigger an immediate
// RequestRedraw — it ensures the animation pumper keeps ticking at its
// configured rate (default 30fps). The next pump tick will render any
// dirty boundaries.
func (c *ContextImpl) ScheduleAnimationFrame() {
	c.mu.RLock()
	cb := c.onScheduleAnimation
	c.mu.RUnlock()
	if cb != nil {
		cb()
		return
	}
	// Fallback: no animation scheduler wired → use immediate InvalidateRect.
	// This happens in headless tests and legacy contexts without Window.
	c.InvalidateRect(geometry.Rect{})
}

// SetOnScheduleAnimation sets the callback for ScheduleAnimationFrame.
func (c *ContextImpl) SetOnScheduleAnimation(callback func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onScheduleAnimation = callback
}

// OverlayManager returns the overlay manager, or nil if none is set.
func (c *ContextImpl) OverlayManager() OverlayManager {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.overlayManager
}

// SetOverlayManager sets the overlay manager for this context.
func (c *ContextImpl) SetOverlayManager(om OverlayManager) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.overlayManager = om
}

// WindowSize returns the current window size in logical pixels.
func (c *ContextImpl) WindowSize() geometry.Size {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.windowSize
}

// SetWindowSize sets the current window size.
func (c *ContextImpl) SetWindowSize(size geometry.Size) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.windowSize = size
}

// Scheduler returns the signal scheduler for this context.
//
// Returns nil if no scheduler is set.
func (c *ContextImpl) Scheduler() SchedulerRef {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.scheduler
}

// SetScheduler sets the signal scheduler for this context.
func (c *ContextImpl) SetScheduler(s SchedulerRef) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.scheduler = s
}

// DrawStats returns the current draw statistics collector for this frame.
//
// Returns nil when no draw pass is in progress. During a draw pass (DrawTree
// or drawWidgetsInRegion), the caller sets a *DrawStats via [SetDrawStats]
// so that widgets like RepaintBoundary can record cache hits.
//
// This method is on the concrete [ContextImpl] type (not the [Context]
// interface) because adding methods to the interface would be a breaking
// change for all implementors.
func (c *ContextImpl) DrawStats() *DrawStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.drawStats
}

// SetDrawStats sets the draw statistics collector for the current frame.
//
// Pass a non-nil *DrawStats before starting a draw pass so that widgets
// can record metrics. Pass nil after the draw pass to release the reference.
func (c *ContextImpl) SetDrawStats(stats *DrawStats) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.drawStats = stats
}

// DirtyTracker returns the dirty region tracker for the current draw pass.
//
// Returns nil when no draw pass is in progress or when the tracker was not
// set (e.g., during full repaints where all widgets are drawn regardless).
// During an incremental draw pass, the caller sets a DirtyTrackerRef via
// [SetDirtyTracker] so that widgets like RepaintBoundary can query spatial
// dirty regions for O(regions) early exit.
//
// This method is on the concrete [ContextImpl] type (not the [Context]
// interface) because adding methods to the interface would be a breaking
// change for all implementors.
func (c *ContextImpl) DirtyTracker() DirtyTrackerRef {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.dirtyTracker
}

// SetDirtyTracker sets the dirty region tracker for the current draw pass.
//
// Pass a non-nil DirtyTrackerRef before starting an incremental draw pass
// so that widgets can query spatial dirty regions. Pass nil after the draw
// pass to release the reference.
func (c *ContextImpl) SetDirtyTracker(tracker DirtyTrackerRef) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.dirtyTracker = tracker
}

// ImageCache returns the centralized RepaintBoundary pixel cache.
//
// Returns nil when no cache is configured (e.g., headless testing without
// a Window). When a Window owns an ImageCache, it calls [SetImageCache]
// during initialization so that all RepaintBoundary instances share a
// single LRU cache with memory budget enforcement.
//
// This method is on the concrete [ContextImpl] type (not the [Context]
// interface) because adding methods to the interface would be a breaking
// change for all implementors.
func (c *ContextImpl) ImageCache() ImageCacheRef {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.imageCache
}

// SetImageCache sets the centralized RepaintBoundary pixel cache.
//
// Pass a non-nil ImageCacheRef during Window initialization to enable
// shared caching. Pass nil to disable (RepaintBoundary falls back to
// per-widget local cache).
func (c *ContextImpl) SetImageCache(cache ImageCacheRef) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.imageCache = cache
}

// RegisterDirtyBoundary registers a RepaintBoundary as dirty in the
// Window's flat dirty boundary set. Called from the onBoundaryDirty
// callback wired by PaintBoundaryLayers.
//
// This populates the O(1) dirty boundary map that replaces O(n)
// NeedsRedrawInTreeNonBoundary tree walks for frame skip decisions.
// The key is the boundary's unique BoundaryCacheKey for deduplication.
//
// If no callback is wired (headless tests), this is a no-op.
func (c *ContextImpl) RegisterDirtyBoundary(key uint64) {
	c.mu.RLock()
	cb := c.onRegisterDirtyBoundary
	c.mu.RUnlock()
	if cb != nil {
		cb(key)
	}
}

// SetOnRegisterDirtyBoundary sets the callback for RegisterDirtyBoundary.
//
// The Window wires this during initialization to AddDirtyBoundary, so that
// upward dirty propagation populates the flat dirty set. This enables O(1)
// HasDirtyBoundaries checks instead of O(n) tree walks.
func (c *ContextImpl) SetOnRegisterDirtyBoundary(callback func(key uint64)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onRegisterDirtyBoundary = callback
}

// CapturePointer requests exclusive mouse event delivery for the widget.
// Events (Move, Release, Press) are delivered directly to the captured
// widget, bypassing tree hit-testing. Only one widget can be captured.
// Auto-releases on MouseRelease when all buttons are up (handled by Window).
//
// If no callback is wired (headless tests), this is a no-op.
func (c *ContextImpl) CapturePointer(w Widget) {
	c.mu.RLock()
	cb := c.onCapturePointer
	c.mu.RUnlock()
	if cb != nil {
		cb(w)
	}
}

// ReleasePointer releases pointer capture for the given widget.
// This is a no-op if the widget does not hold the capture or if no
// callback is wired (headless tests).
func (c *ContextImpl) ReleasePointer(w Widget) {
	c.mu.RLock()
	cb := c.onReleasePointer
	c.mu.RUnlock()
	if cb != nil {
		cb(w)
	}
}

// SetOnCapturePointer sets the callback for CapturePointer.
// The Window wires this during initialization so that widgets can request
// exclusive mouse delivery during drag operations (ADR-031).
func (c *ContextImpl) SetOnCapturePointer(callback func(w Widget)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onCapturePointer = callback
}

// SetOnReleasePointer sets the callback for ReleasePointer.
// The Window wires this during initialization so that widgets can release
// pointer capture when drag operations end.
func (c *ContextImpl) SetOnReleasePointer(callback func(w Widget)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onReleasePointer = callback
}

// Verify ContextImpl implements Context.
var _ Context = (*ContextImpl)(nil)
