package uitest

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// LayoutWidget performs layout on a widget with the given available dimensions
// and returns the resulting size.
//
// The constraints are set to a loose box from (0,0) to (width, height).
// This simulates the common case where a parent offers available space
// and the widget sizes itself within those bounds.
func LayoutWidget(w widget.Widget, width, height float32) geometry.Size {
	ctx := NewMockContext()
	constraints := geometry.Constraints{
		MinWidth:  0,
		MaxWidth:  width,
		MinHeight: 0,
		MaxHeight: height,
	}
	return w.Layout(ctx, constraints)
}

// LayoutWidgetTight performs layout on a widget with tight constraints
// (the widget must be exactly the given size) and returns the resulting size.
func LayoutWidgetTight(w widget.Widget, width, height float32) geometry.Size {
	ctx := NewMockContext()
	constraints := geometry.Tight(geometry.Sz(width, height))
	return w.Layout(ctx, constraints)
}

// DrawWidget performs a draw on the widget using a fresh MockCanvas and
// MockContext, then returns the canvas for inspection.
//
// The widget should have its bounds set before calling this function
// (typically via SetBounds or after a Layout call).
func DrawWidget(w widget.Widget) *MockCanvas {
	canvas := &MockCanvas{}
	ctx := NewMockContext()
	w.Draw(ctx, canvas)
	return canvas
}

// DrawWidgetWithContext performs a draw on the widget using a fresh MockCanvas
// and the provided context, then returns the canvas for inspection.
func DrawWidgetWithContext(w widget.Widget, ctx widget.Context) *MockCanvas {
	canvas := &MockCanvas{}
	w.Draw(ctx, canvas)
	return canvas
}

// SimulateClick sends a press+release sequence at (x, y) to the widget and
// returns true if the press event was consumed.
//
// This simulates a complete click gesture: MousePress followed by MouseRelease
// at the same position.
func SimulateClick(w widget.Widget, x, y float32) bool {
	ctx := NewMockContext()
	press := Click(x, y)
	consumed := w.Event(ctx, press)
	release := Release(x, y)
	w.Event(ctx, release)
	return consumed
}

// SimulateClickWithContext sends a press+release sequence using the provided context.
func SimulateClickWithContext(w widget.Widget, ctx widget.Context, x, y float32) bool {
	press := Click(x, y)
	consumed := w.Event(ctx, press)
	release := Release(x, y)
	w.Event(ctx, release)
	return consumed
}

// SimulateKeyPress sends a key press event to the widget and returns
// true if the event was consumed.
func SimulateKeyPress(w widget.Widget, code event.Key) bool {
	ctx := NewMockContext()
	e := KeyPress(code, event.ModNone)
	return w.Event(ctx, e)
}

// SimulateKeyPressWithMods sends a key press event with modifiers to the widget
// and returns true if the event was consumed.
func SimulateKeyPressWithMods(w widget.Widget, code event.Key, mods event.Modifiers) bool {
	ctx := NewMockContext()
	e := KeyPress(code, mods)
	return w.Event(ctx, e)
}
