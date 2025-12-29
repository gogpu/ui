package event

import "time"

// unknownStr is the string representation for unknown/unrecognized values.
const unknownStr = "Unknown"

// Type represents the category of an event.
type Type uint8

// Event type constants.
const (
	// TypeMouse represents mouse button and movement events.
	TypeMouse Type = iota + 1

	// TypeKey represents keyboard events.
	TypeKey

	// TypeFocus represents focus change events.
	TypeFocus

	// TypeWheel represents mouse wheel scroll events.
	TypeWheel

	// TypeTouch represents touch screen events.
	TypeTouch

	// TypeText represents text input events.
	TypeText

	// TypeDrop represents drag-and-drop events.
	TypeDrop

	// TypeResize represents window/widget resize events.
	TypeResize
)

// String returns a human-readable name for the event type.
func (t Type) String() string {
	switch t {
	case TypeMouse:
		return "Mouse"
	case TypeKey:
		return "Key"
	case TypeFocus:
		return "Focus"
	case TypeWheel:
		return "Wheel"
	case TypeTouch:
		return "Touch"
	case TypeText:
		return "Text"
	case TypeDrop:
		return "Drop"
	case TypeResize:
		return "Resize"
	default:
		return unknownStr
	}
}

// Event is the interface implemented by all event types.
//
// Events carry information about user input and can be marked as handled
// to prevent further propagation through the widget tree.
type Event interface {
	// Type returns the category of this event.
	Type() Type

	// Time returns when the event occurred.
	Time() time.Time

	// Handled returns true if the event has been handled.
	Handled() bool

	// SetHandled marks the event as handled, preventing further propagation.
	SetHandled()

	// Modifiers returns the modifier keys held when the event occurred.
	Modifiers() Modifiers
}

// Base provides common fields and methods for all event types.
//
// Embed this struct in concrete event types to inherit the base implementation.
// All concrete event types must use pointer receivers to allow SetHandled to work.
type Base struct {
	eventType Type
	time      time.Time
	handled   bool
	modifiers Modifiers
}

// NewBase creates a new Base event with the given type and current time.
func NewBase(eventType Type, mods Modifiers) Base {
	return Base{
		eventType: eventType,
		time:      time.Now(),
		modifiers: mods,
	}
}

// NewBaseWithTime creates a new Base event with the given type and timestamp.
func NewBaseWithTime(eventType Type, t time.Time, mods Modifiers) Base {
	return Base{
		eventType: eventType,
		time:      t,
		modifiers: mods,
	}
}

// Type returns the category of this event.
func (b *Base) Type() Type {
	return b.eventType
}

// Time returns when the event occurred.
func (b *Base) Time() time.Time {
	return b.time
}

// Handled returns true if the event has been handled.
func (b *Base) Handled() bool {
	return b.handled
}

// SetHandled marks the event as handled, preventing further propagation.
func (b *Base) SetHandled() {
	b.handled = true
}

// Modifiers returns the modifier keys held when the event occurred.
func (b *Base) Modifiers() Modifiers {
	return b.modifiers
}
