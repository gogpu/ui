package app

import (
	"time"

	"github.com/gogpu/gpucontext"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
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
type Window struct {
	root      widget.Widget
	ctx       *widget.ContextImpl
	wp        gpucontext.WindowProvider
	pp        gpucontext.PlatformProvider
	scheduler *state.Scheduler
	theme     *theme.Theme
	overlays  *overlay.Stack

	// needsLayout indicates that layout should be recalculated.
	needsLayout bool

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
		needsLayout: true,
	}

	// Initialize overlay stack.
	w.overlays = overlay.NewStack(func() {
		w.needsLayout = true
		if w.wp != nil {
			w.wp.RequestRedraw()
		}
	})

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
		if w.wp != nil {
			w.wp.RequestRedraw()
		}
	})

	return w
}

// SetRoot sets the root widget for this window.
//
// Setting a new root triggers a full layout on the next Frame call.
func (w *Window) SetRoot(root widget.Widget) {
	w.root = root
	w.needsLayout = true
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
}

// HandleEvent dispatches a single event to the widget tree.
//
// Events are first offered to the overlay stack (top overlay has priority).
// If no overlay consumes the event (and no modal overlay blocks it),
// the event is propagated to the root widget.
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

	// Dispatch event to root widget.
	_ = w.root.Event(w.ctx, e)
}

// HandleResize processes a window resize.
//
// This updates the window size and marks layout as needing recalculation.
func (w *Window) HandleResize(width, height int) {
	w.windowSize = geometry.Sz(float32(width), float32(height))
	w.needsLayout = true
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

	// Update time.
	w.ctx.SetNow(frameStart)

	// Reset cursor for this frame.
	w.ctx.ResetCursor()

	// Flush any pending state changes.
	w.scheduler.Flush()

	// Update scale factor (may change between frames on multi-monitor setups).
	w.updateScale()

	// Update window size from provider.
	w.updateWindowSize()

	// Perform layout if needed.
	var layoutDur time.Duration
	if w.needsLayout {
		layoutStart := time.Now()
		w.layout()
		layoutDur = time.Since(layoutStart)
		w.needsLayout = false
	}

	// Draw the widget tree.
	drawStart := time.Now()
	w.draw()
	drawDur := time.Since(drawStart)

	// Sync cursor to platform.
	w.syncCursor()

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
		})
	}
}

// NeedsLayout returns true if layout needs recalculation.
func (w *Window) NeedsLayout() bool {
	return w.needsLayout
}

// WindowSize returns the current window size in logical pixels.
func (w *Window) WindowSize() geometry.Size {
	return w.windowSize
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
}

// draw performs the draw pass on the widget tree.
func (w *Window) draw() {
	if w.root == nil {
		return
	}

	// Draw uses a nil canvas in headless mode.
	// In real usage, the host application provides the canvas through
	// a different mechanism (e.g., gg.Context wrapped in render.Canvas).
	// For now, we call Draw with a nil canvas which widgets should handle
	// gracefully, or the host provides one via DrawTo.
	//
	// TODO: Integrate with internal/render.Canvas when the full pipeline
	// is connected through gpucontext.
}

// DrawTo performs the draw pass using the provided canvas.
//
// This method is called by the host application when it has a canvas
// ready for drawing. It draws the root widget tree and then all overlays
// on top.
func (w *Window) DrawTo(canvas widget.Canvas) {
	if w.root == nil || canvas == nil {
		return
	}

	w.root.Draw(w.ctx, canvas)

	// Draw overlays on top (bottom to top).
	w.overlays.Draw(w.ctx, canvas)
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
