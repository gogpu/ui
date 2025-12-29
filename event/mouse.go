package event

import (
	"fmt"
	"time"

	"github.com/gogpu/ui/geometry"
)

// MouseEventType represents the specific type of mouse event.
type MouseEventType uint8

// Mouse event type constants.
const (
	// MousePress indicates a mouse button was pressed.
	MousePress MouseEventType = iota + 1

	// MouseRelease indicates a mouse button was released.
	MouseRelease

	// MouseMove indicates the mouse pointer moved.
	MouseMove

	// MouseEnter indicates the mouse pointer entered a widget's bounds.
	MouseEnter

	// MouseLeave indicates the mouse pointer left a widget's bounds.
	MouseLeave

	// MouseDrag indicates the mouse is moving while a button is held.
	MouseDrag

	// MouseDoubleClick indicates a double-click occurred.
	MouseDoubleClick
)

// String returns a human-readable name for the mouse event type.
func (t MouseEventType) String() string {
	switch t {
	case MousePress:
		return "Press"
	case MouseRelease:
		return "Release"
	case MouseMove:
		return "Move"
	case MouseEnter:
		return "Enter"
	case MouseLeave:
		return "Leave"
	case MouseDrag:
		return "Drag"
	case MouseDoubleClick:
		return "DoubleClick"
	default:
		return unknownStr
	}
}

// Button represents a mouse button.
type Button uint8

// Mouse button constants.
const (
	// ButtonNone indicates no button (used for move events).
	ButtonNone Button = iota

	// ButtonLeft is the primary mouse button.
	ButtonLeft

	// ButtonRight is the secondary mouse button.
	ButtonRight

	// ButtonMiddle is the middle mouse button (scroll wheel click).
	ButtonMiddle

	// ButtonX1 is the first extended button (back button).
	ButtonX1

	// ButtonX2 is the second extended button (forward button).
	ButtonX2
)

// String returns a human-readable name for the button.
func (b Button) String() string {
	switch b {
	case ButtonNone:
		return "None"
	case ButtonLeft:
		return "Left"
	case ButtonRight:
		return "Right"
	case ButtonMiddle:
		return "Middle"
	case ButtonX1:
		return "X1"
	case ButtonX2:
		return "X2"
	default:
		return unknownStr
	}
}

// ButtonState represents which mouse buttons are currently pressed.
type ButtonState uint8

// Button state bit flags.
const (
	// ButtonStateLeft indicates the left button is pressed.
	ButtonStateLeft ButtonState = 1 << iota

	// ButtonStateRight indicates the right button is pressed.
	ButtonStateRight

	// ButtonStateMiddle indicates the middle button is pressed.
	ButtonStateMiddle

	// ButtonStateX1 indicates the X1 button is pressed.
	ButtonStateX1

	// ButtonStateX2 indicates the X2 button is pressed.
	ButtonStateX2
)

// Has returns true if the specified button is pressed.
func (s ButtonState) Has(state ButtonState) bool {
	return s&state == state
}

// IsLeftPressed returns true if the left button is pressed.
func (s ButtonState) IsLeftPressed() bool {
	return s.Has(ButtonStateLeft)
}

// IsRightPressed returns true if the right button is pressed.
func (s ButtonState) IsRightPressed() bool {
	return s.Has(ButtonStateRight)
}

// IsMiddlePressed returns true if the middle button is pressed.
func (s ButtonState) IsMiddlePressed() bool {
	return s.Has(ButtonStateMiddle)
}

// IsX1Pressed returns true if the X1 button is pressed.
func (s ButtonState) IsX1Pressed() bool {
	return s.Has(ButtonStateX1)
}

// IsX2Pressed returns true if the X2 button is pressed.
func (s ButtonState) IsX2Pressed() bool {
	return s.Has(ButtonStateX2)
}

// AnyPressed returns true if any button is pressed.
func (s ButtonState) AnyPressed() bool {
	return s != 0
}

// MouseEvent represents a mouse input event.
//
// Position is relative to the widget receiving the event.
// GlobalPosition is the position in screen coordinates.
type MouseEvent struct {
	Base

	// MouseType is the specific type of mouse event.
	MouseType MouseEventType

	// Button is the button involved in press/release events.
	Button Button

	// Buttons is the state of all mouse buttons.
	Buttons ButtonState

	// Position is the mouse position relative to the widget.
	Position geometry.Point

	// GlobalPosition is the mouse position in screen coordinates.
	GlobalPosition geometry.Point

	// ClickCount is the number of consecutive clicks (1 for single, 2 for double).
	ClickCount int
}

// NewMouseEvent creates a new mouse event with the current time.
func NewMouseEvent(
	mouseType MouseEventType,
	button Button,
	buttons ButtonState,
	position geometry.Point,
	globalPosition geometry.Point,
	mods Modifiers,
) *MouseEvent {
	return &MouseEvent{
		Base:           NewBase(TypeMouse, mods),
		MouseType:      mouseType,
		Button:         button,
		Buttons:        buttons,
		Position:       position,
		GlobalPosition: globalPosition,
		ClickCount:     1,
	}
}

// NewMouseEventWithTime creates a new mouse event with a specific timestamp.
func NewMouseEventWithTime(
	mouseType MouseEventType,
	button Button,
	buttons ButtonState,
	position geometry.Point,
	globalPosition geometry.Point,
	mods Modifiers,
	t time.Time,
) *MouseEvent {
	return &MouseEvent{
		Base:           NewBaseWithTime(TypeMouse, t, mods),
		MouseType:      mouseType,
		Button:         button,
		Buttons:        buttons,
		Position:       position,
		GlobalPosition: globalPosition,
		ClickCount:     1,
	}
}

// IsPress returns true if this is a button press event.
func (e *MouseEvent) IsPress() bool {
	return e.MouseType == MousePress
}

// IsRelease returns true if this is a button release event.
func (e *MouseEvent) IsRelease() bool {
	return e.MouseType == MouseRelease
}

// IsMove returns true if this is a mouse move event.
func (e *MouseEvent) IsMove() bool {
	return e.MouseType == MouseMove
}

// IsEnter returns true if this is a mouse enter event.
func (e *MouseEvent) IsEnter() bool {
	return e.MouseType == MouseEnter
}

// IsLeave returns true if this is a mouse leave event.
func (e *MouseEvent) IsLeave() bool {
	return e.MouseType == MouseLeave
}

// IsDrag returns true if this is a drag event.
func (e *MouseEvent) IsDrag() bool {
	return e.MouseType == MouseDrag
}

// IsDoubleClick returns true if this is a double-click event.
func (e *MouseEvent) IsDoubleClick() bool {
	return e.MouseType == MouseDoubleClick
}

// IsLeftButton returns true if the left button is involved.
func (e *MouseEvent) IsLeftButton() bool {
	return e.Button == ButtonLeft
}

// IsRightButton returns true if the right button is involved.
func (e *MouseEvent) IsRightButton() bool {
	return e.Button == ButtonRight
}

// IsMiddleButton returns true if the middle button is involved.
func (e *MouseEvent) IsMiddleButton() bool {
	return e.Button == ButtonMiddle
}

// String returns a human-readable representation of the event.
func (e *MouseEvent) String() string {
	return fmt.Sprintf("MouseEvent{Type: %s, Button: %s, Position: %s, Mods: %s}",
		e.MouseType, e.Button, e.Position, e.Modifiers())
}
