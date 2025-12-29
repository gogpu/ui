package widget

import (
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
		return "Unknown"
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

	// Callback for invalidation (called when Invalidate is called)
	onInvalidate func()

	// Callback for invalidate rect (called when InvalidateRect is called)
	onInvalidateRect func(geometry.Rect)
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

// SetNow updates the current time and calculates delta time.
//
// Call this at the start of each frame before processing events and layout.
func (c *ContextImpl) SetNow(now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.deltaTime = now.Sub(c.now)
	c.lastFrame = c.now
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

// Verify ContextImpl implements Context.
var _ Context = (*ContextImpl)(nil)
