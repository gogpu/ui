package uitest

import (
	"time"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// defaultDelta is the default frame delta time (approximately 60 FPS).
const defaultDelta = 16 * time.Millisecond

// MockContext provides a configurable test implementation of [widget.Context].
//
// All fields are exported for direct configuration. Use [NewMockContext] for
// sensible defaults, then override fields as needed.
//
// Example:
//
//	ctx := uitest.NewMockContext()
//	ctx.ScaleVal = 2.0  // HiDPI
//	myWidget.Draw(ctx, canvas)
//	if ctx.Invalidated {
//	    t.Log("widget requested redraw")
//	}
type MockContext struct {
	// FocusedVal is the widget that currently has focus.
	FocusedVal widget.Widget

	// TimeVal is returned by Now(). Defaults to a fixed time.
	TimeVal time.Time

	// DeltaVal is returned by DeltaTime(). Defaults to 16ms.
	DeltaVal time.Duration

	// ScaleVal is returned by Scale(). Defaults to 1.0.
	ScaleVal float32

	// CursorVal is returned by Cursor() and set by SetCursor().
	CursorVal widget.CursorType

	// ThemeVal is returned by ThemeProvider(). Defaults to nil.
	ThemeVal widget.ThemeProvider

	// OverlayVal is returned by OverlayManager(). Defaults to nil.
	OverlayVal widget.OverlayManager

	// WindowSizeVal is returned by WindowSize(). Defaults to 800x600.
	WindowSizeVal geometry.Size

	// SchedulerVal is returned by Scheduler(). Defaults to nil.
	SchedulerVal widget.SchedulerRef

	// Invalidated is set to true when Invalidate() is called.
	Invalidated bool

	// InvalidatedRects collects all rectangles passed to InvalidateRect().
	InvalidatedRects []geometry.Rect

	// InvalidateCount tracks how many times Invalidate() was called.
	InvalidateCount int
}

// NewMockContext creates a MockContext with sensible defaults for testing.
//
// Defaults:
//   - Scale: 1.0
//   - DeltaTime: 16ms
//   - WindowSize: 800x600
//   - Time: 2025-01-01 00:00:00 UTC (deterministic for tests)
func NewMockContext() *MockContext {
	return &MockContext{
		TimeVal:       time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		DeltaVal:      defaultDelta,
		ScaleVal:      1.0,
		WindowSizeVal: geometry.Sz(800, 600),
	}
}

// RequestFocus sets FocusedVal to the given widget.
func (c *MockContext) RequestFocus(w widget.Widget) {
	// Notify previous widget of focus loss.
	if c.FocusedVal != nil && c.FocusedVal != w {
		if setter, ok := c.FocusedVal.(interface{ SetFocused(bool) }); ok {
			setter.SetFocused(false)
		}
	}
	c.FocusedVal = w
	if w != nil {
		if setter, ok := w.(interface{ SetFocused(bool) }); ok {
			setter.SetFocused(true)
		}
	}
}

// ReleaseFocus clears focus if the given widget currently has it.
func (c *MockContext) ReleaseFocus(w widget.Widget) {
	if c.FocusedVal != w {
		return
	}
	if setter, ok := c.FocusedVal.(interface{ SetFocused(bool) }); ok {
		setter.SetFocused(false)
	}
	c.FocusedVal = nil
}

// IsFocused returns true if w is the currently focused widget.
func (c *MockContext) IsFocused(w widget.Widget) bool {
	return c.FocusedVal == w
}

// FocusedWidget returns the currently focused widget.
func (c *MockContext) FocusedWidget() widget.Widget {
	return c.FocusedVal
}

// Now returns TimeVal.
func (c *MockContext) Now() time.Time {
	return c.TimeVal
}

// DeltaTime returns DeltaVal.
func (c *MockContext) DeltaTime() time.Duration {
	return c.DeltaVal
}

// Invalidate sets Invalidated to true and increments InvalidateCount.
func (c *MockContext) Invalidate() {
	c.Invalidated = true
	c.InvalidateCount++
}

// InvalidateRect appends the rectangle to InvalidatedRects.
func (c *MockContext) InvalidateRect(r geometry.Rect) {
	c.InvalidatedRects = append(c.InvalidatedRects, r)
}

// Cursor returns CursorVal.
func (c *MockContext) Cursor() widget.CursorType {
	return c.CursorVal
}

// SetCursor sets CursorVal.
func (c *MockContext) SetCursor(cursor widget.CursorType) {
	c.CursorVal = cursor
}

// Scale returns ScaleVal.
func (c *MockContext) Scale() float32 {
	return c.ScaleVal
}

// ThemeProvider returns ThemeVal.
func (c *MockContext) ThemeProvider() widget.ThemeProvider {
	return c.ThemeVal
}

// OverlayManager returns OverlayVal.
func (c *MockContext) OverlayManager() widget.OverlayManager {
	return c.OverlayVal
}

// WindowSize returns WindowSizeVal.
func (c *MockContext) WindowSize() geometry.Size {
	return c.WindowSizeVal
}

// Scheduler returns SchedulerVal.
func (c *MockContext) Scheduler() widget.SchedulerRef {
	return c.SchedulerVal
}

// Reset clears all recorded state (invalidation, cursor) while preserving configuration.
func (c *MockContext) Reset() {
	c.Invalidated = false
	c.InvalidateCount = 0
	c.InvalidatedRects = c.InvalidatedRects[:0]
	c.CursorVal = widget.CursorDefault
}

// Compile-time interface check.
var _ widget.Context = (*MockContext)(nil)
