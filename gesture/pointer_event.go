package gesture

import (
	"fmt"
	"time"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
)

// PointerEventType indicates the type of pointer event.
type PointerEventType uint8

const (
	// PointerDown indicates a pointer became active (button pressed, finger touched).
	PointerDown PointerEventType = iota

	// PointerUp indicates a pointer was deactivated (button released, finger lifted).
	PointerUp

	// PointerMove indicates a pointer position changed.
	PointerMove

	// PointerCancel indicates the system canceled the pointer sequence.
	PointerCancel
)

// PointerEventType string constants for goconst compliance.
const (
	pointerDownStr   = "Down"
	pointerUpStr     = "Up"
	pointerMoveStr   = "Move"
	pointerCancelStr = "Cancel"
)

// String returns a human-readable name for the pointer event type.
func (t PointerEventType) String() string {
	switch t {
	case PointerDown:
		return pointerDownStr
	case PointerUp:
		return pointerUpStr
	case PointerMove:
		return pointerMoveStr
	case PointerCancel:
		return pointerCancelStr
	default:
		return pointerUnknownStr
	}
}

// PointerType distinguishes the category of pointing device.
type PointerType uint8

const (
	// PointerTypeMouse is a mouse or trackpad pointer.
	PointerTypeMouse PointerType = iota

	// PointerTypeTouch is a finger on a touch screen.
	PointerTypeTouch

	// PointerTypePen is a stylus or pen input device.
	PointerTypePen
)

// PointerType string constants for goconst compliance.
const (
	pointerMouseStr   = "Mouse"
	pointerTouchStr   = "Touch"
	pointerPenStr     = "Pen"
	pointerUnknownStr = "Unknown"
)

// String returns a human-readable name for the pointer type.
func (t PointerType) String() string {
	switch t {
	case PointerTypeMouse:
		return pointerMouseStr
	case PointerTypeTouch:
		return pointerTouchStr
	case PointerTypePen:
		return pointerPenStr
	default:
		return pointerUnknownStr
	}
}

// DeviceKind classifies a pointing device for threshold selection.
type DeviceKind uint8

const (
	// DeviceKindPrecise classifies mouse and trackpad pointers.
	DeviceKindPrecise DeviceKind = iota

	// DeviceKindTouch classifies touch and pen pointers.
	DeviceKindTouch
)

// DeviceKind string constants for goconst compliance.
const (
	devicePreciseStr = "Precise"
	deviceTouchStr   = "Touch"
)

// String returns a human-readable name for the device kind.
func (k DeviceKind) String() string {
	switch k {
	case DeviceKindPrecise:
		return devicePreciseStr
	case DeviceKindTouch:
		return deviceTouchStr
	default:
		return pointerUnknownStr
	}
}

// DeviceKind returns a classification used for threshold selection.
// Touch and Pen use touch thresholds; Mouse uses precise thresholds.
func (t PointerType) DeviceKind() DeviceKind {
	switch t {
	case PointerTypeTouch, PointerTypePen:
		return DeviceKindTouch
	default:
		return DeviceKindPrecise
	}
}

// PointerEvent carries unified pointer data for gesture recognition.
// It extends event.Base with W3C Pointer Events Level 3 fields from
// gpucontext.PointerEvent, adding widget-relative positioning.
//
// PointerEvent is the sole input type for all Recognizer implementations.
// The event_bridge constructs these from gpucontext.PointerEvent, enriching
// them with widget-relative coordinates during tree dispatch.
type PointerEvent struct {
	event.Base

	// EventType is the pointer event type (Down, Up, Move, Cancel).
	EventType PointerEventType

	// PointerID uniquely identifies this pointer across its lifetime.
	// Mouse: always 1. Touch: per-finger. Pen: per-stylus.
	PointerID int

	// PointerType distinguishes the input device.
	PointerType PointerType

	// Position is the pointer location relative to the receiving widget.
	Position geometry.Point

	// GlobalPosition is the pointer location in window coordinates.
	GlobalPosition geometry.Point

	// Pressure is the normalized pressure (0.0-1.0).
	// Mouse: 0.5 when pressed, 0.0 when not. Touch/Pen: actual pressure.
	Pressure float32

	// TiltX is the pen tilt angle around the X axis in degrees (-90 to 90).
	TiltX float32

	// TiltY is the pen tilt angle around the Y axis in degrees (-90 to 90).
	TiltY float32

	// Twist is the pen rotation in degrees (0 to 359).
	Twist float32

	// ContactWidth is the touch contact width in logical pixels.
	// 1.0 for devices without contact geometry (mouse).
	ContactWidth float32

	// ContactHeight is the touch contact height in logical pixels.
	// 1.0 for devices without contact geometry (mouse).
	ContactHeight float32

	// Button is the button that triggered this event (Down/Up only).
	Button event.Button

	// Buttons is the bitmask of all currently pressed buttons.
	Buttons event.ButtonState

	// Delta is the relative movement since the last event.
	// Non-zero only during pointer-locked mode.
	Delta geometry.Point

	// Timestamp is the platform event time for velocity calculation.
	// Zero if the platform does not provide timestamps.
	Timestamp time.Duration
}

// String returns a human-readable representation of the pointer event.
func (e *PointerEvent) String() string {
	return fmt.Sprintf("PointerEvent{Type: %s, ID: %d, Pointer: %s, Pos: %s}",
		e.EventType, e.PointerID, e.PointerType, e.Position)
}
