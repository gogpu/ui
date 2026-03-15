package uitest

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
)

// Click creates a left mouse button press event at the given position.
//
// Both Position and GlobalPosition are set to (x, y) with no modifiers.
func Click(x, y float32) *event.MouseEvent {
	pos := geometry.Pt(x, y)
	return event.NewMouseEvent(
		event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		pos, pos, event.ModNone,
	)
}

// Release creates a left mouse button release event at the given position.
//
// Both Position and GlobalPosition are set to (x, y) with no modifiers.
func Release(x, y float32) *event.MouseEvent {
	pos := geometry.Pt(x, y)
	return event.NewMouseEvent(
		event.MouseRelease, event.ButtonLeft, 0,
		pos, pos, event.ModNone,
	)
}

// DoubleClick creates a double-click event at the given position.
//
// Both Position and GlobalPosition are set to (x, y) with ClickCount=2.
func DoubleClick(x, y float32) *event.MouseEvent {
	pos := geometry.Pt(x, y)
	e := event.NewMouseEvent(
		event.MouseDoubleClick, event.ButtonLeft, event.ButtonStateLeft,
		pos, pos, event.ModNone,
	)
	e.ClickCount = 2
	return e
}

// RightClick creates a right mouse button press event at the given position.
func RightClick(x, y float32) *event.MouseEvent {
	pos := geometry.Pt(x, y)
	return event.NewMouseEvent(
		event.MousePress, event.ButtonRight, event.ButtonStateRight,
		pos, pos, event.ModNone,
	)
}

// MouseMove creates a mouse move event at the given position.
func MouseMove(x, y float32) *event.MouseEvent {
	pos := geometry.Pt(x, y)
	return event.NewMouseEvent(
		event.MouseMove, event.ButtonNone, 0,
		pos, pos, event.ModNone,
	)
}

// MouseEnter creates a mouse enter event at the given position.
func MouseEnter(x, y float32) *event.MouseEvent {
	pos := geometry.Pt(x, y)
	return event.NewMouseEvent(
		event.MouseEnter, event.ButtonNone, 0,
		pos, pos, event.ModNone,
	)
}

// MouseLeave creates a mouse leave event at the given position.
func MouseLeave(x, y float32) *event.MouseEvent {
	pos := geometry.Pt(x, y)
	return event.NewMouseEvent(
		event.MouseLeave, event.ButtonNone, 0,
		pos, pos, event.ModNone,
	)
}

// MouseDrag creates a mouse drag event at the given position with the left button held.
func MouseDrag(x, y float32) *event.MouseEvent {
	pos := geometry.Pt(x, y)
	return event.NewMouseEvent(
		event.MouseDrag, event.ButtonLeft, event.ButtonStateLeft,
		pos, pos, event.ModNone,
	)
}

// KeyPress creates a key press event for the given key code with no modifiers.
func KeyPress(code event.Key, mods event.Modifiers) *event.KeyEvent {
	return event.NewKeyEvent(event.KeyPress, code, 0, mods)
}

// KeyRelease creates a key release event for the given key code.
func KeyRelease(code event.Key, mods event.Modifiers) *event.KeyEvent {
	return event.NewKeyEvent(event.KeyRelease, code, 0, mods)
}

// KeyType creates a key press event that also produces a character (rune).
//
// Use this for simulating text input where both the key code and the
// typed character matter.
func KeyType(code event.Key, r rune, mods event.Modifiers) *event.KeyEvent {
	return event.NewKeyEvent(event.KeyPress, code, r, mods)
}

// WheelScroll creates a wheel scroll event at position (x, y) with the given vertical delta.
//
// Positive deltaY scrolls up (content moves down), negative scrolls down.
func WheelScroll(x, y, deltaY float32) *event.WheelEvent {
	pos := geometry.Pt(x, y)
	delta := geometry.Pt(0, deltaY)
	return event.NewWheelEvent(delta, pos, pos, event.ModNone)
}

// WheelScrollH creates a wheel scroll event with both horizontal and vertical deltas.
func WheelScrollH(x, y, deltaX, deltaY float32) *event.WheelEvent {
	pos := geometry.Pt(x, y)
	delta := geometry.Pt(deltaX, deltaY)
	return event.NewWheelEvent(delta, pos, pos, event.ModNone)
}

// FocusGained creates a focus gained event.
func FocusGained() *event.FocusEvent {
	return event.NewFocusEvent(event.FocusGained)
}

// FocusLost creates a focus lost event.
func FocusLost() *event.FocusEvent {
	return event.NewFocusEvent(event.FocusLost)
}
