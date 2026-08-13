package app

import (
	"testing"
	"time"

	"github.com/gogpu/gpucontext"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/gesture"
)

func TestConvertPointerEvent_AllTypes(t *testing.T) {
	tests := []struct {
		name     string
		inType   gpucontext.PointerEventType
		wantType gesture.PointerEventType
		wantOK   bool
	}{
		{"Down", gpucontext.PointerDown, gesture.PointerDown, true},
		{"Up", gpucontext.PointerUp, gesture.PointerUp, true},
		{"Move", gpucontext.PointerMove, gesture.PointerMove, true},
		{"Cancel", gpucontext.PointerCancel, gesture.PointerCancel, true},
		{"Enter", gpucontext.PointerEnter, 0, false},
		{"Leave", gpucontext.PointerLeave, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pev := gpucontext.PointerEvent{
				Type:      tt.inType,
				PointerID: 42,
				X:         100.0,
				Y:         200.0,
			}
			gev, ok := convertPointerEvent(pev)
			if ok != tt.wantOK {
				t.Fatalf("convertPointerEvent ok = %v, want %v", ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if gev.EventType != tt.wantType {
				t.Errorf("EventType = %v, want %v", gev.EventType, tt.wantType)
			}
			if gev.PointerID != 42 {
				t.Errorf("PointerID = %d, want 42", gev.PointerID)
			}
		})
	}
}

func TestConvertPointerEvent_Position(t *testing.T) {
	pev := gpucontext.PointerEvent{
		Type: gpucontext.PointerDown,
		X:    150.5,
		Y:    250.75,
	}
	gev, ok := convertPointerEvent(pev)
	if !ok {
		t.Fatal("convertPointerEvent returned false")
	}
	wantPos := geometry.Pt(150.5, 250.75)
	if gev.Position != wantPos {
		t.Errorf("Position = %v, want %v", gev.Position, wantPos)
	}
	if gev.GlobalPosition != wantPos {
		t.Errorf("GlobalPosition = %v, want %v", gev.GlobalPosition, wantPos)
	}
}

func TestConvertPointerEvent_RichFields(t *testing.T) {
	ts := 500 * time.Millisecond
	pev := gpucontext.PointerEvent{
		Type:        gpucontext.PointerDown,
		PointerID:   3,
		X:           10.0,
		Y:           20.0,
		Pressure:    0.75,
		TiltX:       15.0,
		TiltY:       -10.0,
		Twist:       45.0,
		Width:       5.0,
		Height:      8.0,
		PointerType: gpucontext.PointerTypePen,
		Button:      gpucontext.ButtonLeft,
		Buttons:     gpucontext.ButtonsLeft,
		DeltaX:      1.5,
		DeltaY:      -2.5,
		Timestamp:   ts,
		Modifiers:   gpucontext.ModShift,
	}

	gev, ok := convertPointerEvent(pev)
	if !ok {
		t.Fatal("convertPointerEvent returned false")
	}

	if gev.PointerID != 3 {
		t.Errorf("PointerID = %d, want 3", gev.PointerID)
	}
	if gev.PointerType != gesture.PointerTypePen {
		t.Errorf("PointerType = %v, want Pen", gev.PointerType)
	}
	if gev.Pressure != 0.75 {
		t.Errorf("Pressure = %v, want 0.75", gev.Pressure)
	}
	if gev.TiltX != 15.0 {
		t.Errorf("TiltX = %v, want 15.0", gev.TiltX)
	}
	if gev.TiltY != -10.0 {
		t.Errorf("TiltY = %v, want -10.0", gev.TiltY)
	}
	if gev.Twist != 45.0 {
		t.Errorf("Twist = %v, want 45.0", gev.Twist)
	}
	if gev.ContactWidth != 5.0 {
		t.Errorf("ContactWidth = %v, want 5.0", gev.ContactWidth)
	}
	if gev.ContactHeight != 8.0 {
		t.Errorf("ContactHeight = %v, want 8.0", gev.ContactHeight)
	}
	if gev.Button != event.ButtonLeft {
		t.Errorf("Button = %v, want Left", gev.Button)
	}
	if !gev.Buttons.IsLeftPressed() {
		t.Error("Buttons should have Left pressed")
	}
	if gev.Delta.X != 1.5 || gev.Delta.Y != -2.5 {
		t.Errorf("Delta = %v, want (1.5, -2.5)", gev.Delta)
	}
	if gev.Timestamp != ts {
		t.Errorf("Timestamp = %v, want %v", gev.Timestamp, ts)
	}
	if !gev.Modifiers().IsShift() {
		t.Error("expected Shift modifier")
	}
}

func TestConvertPointerType(t *testing.T) {
	tests := []struct {
		name string
		in   gpucontext.PointerType
		want gesture.PointerType
	}{
		{"Mouse", gpucontext.PointerTypeMouse, gesture.PointerTypeMouse},
		{"Touch", gpucontext.PointerTypeTouch, gesture.PointerTypeTouch},
		{"Pen", gpucontext.PointerTypePen, gesture.PointerTypePen},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertPointerType(tt.in)
			if got != tt.want {
				t.Errorf("convertPointerType(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestConvertPointerButton(t *testing.T) {
	tests := []struct {
		name string
		in   gpucontext.Button
		want event.Button
	}{
		{"Left", gpucontext.ButtonLeft, event.ButtonLeft},
		{"Right", gpucontext.ButtonRight, event.ButtonRight},
		{"Middle", gpucontext.ButtonMiddle, event.ButtonMiddle},
		{"X1", gpucontext.ButtonX1, event.ButtonX1},
		{"X2", gpucontext.ButtonX2, event.ButtonX2},
		{"None", gpucontext.ButtonNone, event.ButtonNone},
		{"Eraser", gpucontext.ButtonEraser, event.ButtonNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertPointerButton(tt.in)
			if got != tt.want {
				t.Errorf("convertPointerButton(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestConvertPointerButtons(t *testing.T) {
	tests := []struct {
		name string
		in   gpucontext.Buttons
		want event.ButtonState
	}{
		{"None", gpucontext.ButtonsNone, 0},
		{"Left", gpucontext.ButtonsLeft, event.ButtonStateLeft},
		{"Right", gpucontext.ButtonsRight, event.ButtonStateRight},
		{"Middle", gpucontext.ButtonsMiddle, event.ButtonStateMiddle},
		{"X1", gpucontext.ButtonsX1, event.ButtonStateX1},
		{"X2", gpucontext.ButtonsX2, event.ButtonStateX2},
		{"LeftRight", gpucontext.ButtonsLeft | gpucontext.ButtonsRight,
			event.ButtonStateLeft | event.ButtonStateRight},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertPointerButtons(tt.in)
			if got != tt.want {
				t.Errorf("convertPointerButtons(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestSynthesizePointerEvent_Press(t *testing.T) {
	pos := geometry.Pt(100, 200)
	gev := synthesizePointerEvent(
		event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		pos, event.ModShift,
	)

	if gev.EventType != gesture.PointerDown {
		t.Errorf("EventType = %v, want PointerDown", gev.EventType)
	}
	if gev.PointerID != 1 {
		t.Errorf("PointerID = %d, want 1", gev.PointerID)
	}
	if gev.PointerType != gesture.PointerTypeMouse {
		t.Errorf("PointerType = %v, want Mouse", gev.PointerType)
	}
	if gev.Position != pos {
		t.Errorf("Position = %v, want %v", gev.Position, pos)
	}
	if gev.Pressure != 0.5 {
		t.Errorf("Pressure = %v, want 0.5", gev.Pressure)
	}
	if gev.Button != event.ButtonLeft {
		t.Errorf("Button = %v, want Left", gev.Button)
	}
	if !gev.Buttons.IsLeftPressed() {
		t.Error("Buttons should have Left pressed")
	}
	if gev.Timestamp == 0 {
		t.Error("Timestamp should be non-zero for synthesized event")
	}
	if !gev.Modifiers().IsShift() {
		t.Error("expected Shift modifier")
	}
}

func TestSynthesizePointerEvent_Release(t *testing.T) {
	pos := geometry.Pt(50, 60)
	gev := synthesizePointerEvent(
		event.MouseRelease, event.ButtonLeft, 0,
		pos, event.ModNone,
	)

	if gev.EventType != gesture.PointerUp {
		t.Errorf("EventType = %v, want PointerUp", gev.EventType)
	}
	if gev.Pressure != 0.0 {
		t.Errorf("Pressure = %v, want 0.0 (no buttons pressed)", gev.Pressure)
	}
}

func TestSynthesizePointerEvent_Move(t *testing.T) {
	pos := geometry.Pt(75, 80)
	gev := synthesizePointerEvent(
		event.MouseMove, event.ButtonNone, 0,
		pos, event.ModNone,
	)

	if gev.EventType != gesture.PointerMove {
		t.Errorf("EventType = %v, want PointerMove", gev.EventType)
	}
}

func TestSynthesizePointerEvent_Drag(t *testing.T) {
	pos := geometry.Pt(120, 130)
	gev := synthesizePointerEvent(
		event.MouseDrag, event.ButtonLeft, event.ButtonStateLeft,
		pos, event.ModNone,
	)

	if gev.EventType != gesture.PointerMove {
		t.Errorf("EventType = %v, want PointerMove (drag maps to move)", gev.EventType)
	}
	if gev.Pressure != 0.5 {
		t.Errorf("Pressure = %v, want 0.5 (button pressed during drag)", gev.Pressure)
	}
}

func TestSynthesizePointerEvent_NonGestureTypes(t *testing.T) {
	// Enter, Leave, DoubleClick should produce zero-value events.
	for _, mt := range []event.MouseEventType{
		event.MouseEnter, event.MouseLeave, event.MouseDoubleClick,
	} {
		gev := synthesizePointerEvent(mt, event.ButtonNone, 0, geometry.Point{}, event.ModNone)
		if gev.EventType != 0 {
			t.Errorf("synthesizePointerEvent(%v) EventType = %v, want 0", mt, gev.EventType)
		}
	}
}

func TestSynthesizePointerEvent_ContactDefaults(t *testing.T) {
	gev := synthesizePointerEvent(
		event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(10, 10), event.ModNone,
	)

	if gev.ContactWidth != 1.0 {
		t.Errorf("ContactWidth = %v, want 1.0", gev.ContactWidth)
	}
	if gev.ContactHeight != 1.0 {
		t.Errorf("ContactHeight = %v, want 1.0", gev.ContactHeight)
	}
}
