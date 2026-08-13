package app

import (
	"time"

	"github.com/gogpu/gpucontext"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/gesture"
)

// convertPointerEvent converts a gpucontext.PointerEvent to a gesture.PointerEvent.
//
// This maps W3C Pointer Events Level 3 fields from the platform layer into the
// ui-level gesture.PointerEvent that all gesture recognizers consume.
//
// Enter and Leave events are not gesture events (they have no pointer lifecycle);
// they are handled by the existing attachPointerBridge. This function returns
// a zero event and false for those types.
func convertPointerEvent(pev gpucontext.PointerEvent) (gesture.PointerEvent, bool) {
	var evType gesture.PointerEventType
	switch pev.Type {
	case gpucontext.PointerDown:
		evType = gesture.PointerDown
	case gpucontext.PointerUp:
		evType = gesture.PointerUp
	case gpucontext.PointerMove:
		evType = gesture.PointerMove
	case gpucontext.PointerCancel:
		evType = gesture.PointerCancel
	default:
		// Enter/Leave are not gesture events.
		return gesture.PointerEvent{}, false
	}

	pos := geometry.Pt(float32(pev.X), float32(pev.Y))

	gev := gesture.PointerEvent{
		Base:           event.NewBase(event.TypeMouse, translateModifiers(pev.Modifiers)),
		EventType:      evType,
		PointerID:      pev.PointerID,
		PointerType:    convertPointerType(pev.PointerType),
		Position:       pos,
		GlobalPosition: pos, // widget-relative adjusted during dispatch
		Pressure:       pev.Pressure,
		TiltX:          pev.TiltX,
		TiltY:          pev.TiltY,
		Twist:          pev.Twist,
		ContactWidth:   pev.Width,
		ContactHeight:  pev.Height,
		Button:         convertPointerButton(pev.Button),
		Buttons:        convertPointerButtons(pev.Buttons),
		Delta:          geometry.Pt(float32(pev.DeltaX), float32(pev.DeltaY)),
		Timestamp:      pev.Timestamp,
	}

	return gev, true
}

// synthesizePointerEvent creates a gesture.PointerEvent from legacy mouse callback
// data. Used when the platform does not provide PointerEventSource, so gesture
// recognition still works with mouse-only platforms.
//
// The synthesized event uses PointerID=1 (mouse is always a single pointer),
// PointerTypeMouse, default pressure (0.5 when pressed, 0.0 otherwise), and
// time.Now() as timestamp (since legacy callbacks do not provide timestamps).
func synthesizePointerEvent(
	mouseType event.MouseEventType,
	btn event.Button,
	buttons event.ButtonState,
	pos geometry.Point,
	mods event.Modifiers,
) gesture.PointerEvent {
	var evType gesture.PointerEventType
	switch mouseType {
	case event.MousePress:
		evType = gesture.PointerDown
	case event.MouseRelease:
		evType = gesture.PointerUp
	case event.MouseMove, event.MouseDrag:
		evType = gesture.PointerMove
	default:
		// Enter, Leave, DoubleClick are not gesture events.
		return gesture.PointerEvent{}
	}

	// Synthesize pressure: 0.5 when any button pressed (W3C mouse default).
	var pressure float32
	if buttons != 0 || mouseType == event.MousePress {
		pressure = 0.5
	}

	return gesture.PointerEvent{
		Base:           event.NewBase(event.TypeMouse, mods),
		EventType:      evType,
		PointerID:      1, // Mouse always uses pointer ID 1.
		PointerType:    gesture.PointerTypeMouse,
		Position:       pos,
		GlobalPosition: pos,
		Pressure:       pressure,
		ContactWidth:   1.0, // Default for mouse.
		ContactHeight:  1.0,
		Button:         btn,
		Buttons:        buttons,
		Timestamp:      time.Duration(time.Now().UnixNano()),
	}
}

// convertPointerType maps gpucontext.PointerType to gesture.PointerType.
func convertPointerType(pt gpucontext.PointerType) gesture.PointerType {
	switch pt {
	case gpucontext.PointerTypeTouch:
		return gesture.PointerTypeTouch
	case gpucontext.PointerTypePen:
		return gesture.PointerTypePen
	default:
		return gesture.PointerTypeMouse
	}
}

// convertPointerButton maps gpucontext.Button to event.Button.
func convertPointerButton(btn gpucontext.Button) event.Button {
	switch btn {
	case gpucontext.ButtonLeft:
		return event.ButtonLeft
	case gpucontext.ButtonRight:
		return event.ButtonRight
	case gpucontext.ButtonMiddle:
		return event.ButtonMiddle
	case gpucontext.ButtonX1:
		return event.ButtonX1
	case gpucontext.ButtonX2:
		return event.ButtonX2
	default:
		return event.ButtonNone
	}
}

// convertPointerButtons maps gpucontext.Buttons bitmask to event.ButtonState.
func convertPointerButtons(btns gpucontext.Buttons) event.ButtonState {
	var state event.ButtonState
	if btns.HasLeft() {
		state |= event.ButtonStateLeft
	}
	if btns.HasRight() {
		state |= event.ButtonStateRight
	}
	if btns.HasMiddle() {
		state |= event.ButtonStateMiddle
	}
	if btns.HasX1() {
		state |= event.ButtonStateX1
	}
	if btns.HasX2() {
		state |= event.ButtonStateX2
	}
	return state
}
