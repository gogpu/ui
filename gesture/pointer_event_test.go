package gesture

import (
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
)

func TestPointerEventType_String(t *testing.T) {
	tests := []struct {
		typ  PointerEventType
		want string
	}{
		{PointerDown, "Down"},
		{PointerUp, "Up"},
		{PointerMove, "Move"},
		{PointerCancel, "Cancel"},
		{PointerEventType(99), "Unknown"},
	}

	for _, tt := range tests {
		got := tt.typ.String()
		if got != tt.want {
			t.Errorf("PointerEventType(%d).String() = %q, want %q", tt.typ, got, tt.want)
		}
	}
}

func TestPointerType_String(t *testing.T) {
	tests := []struct {
		typ  PointerType
		want string
	}{
		{PointerTypeMouse, "Mouse"},
		{PointerTypeTouch, "Touch"},
		{PointerTypePen, "Pen"},
		{PointerType(99), "Unknown"},
	}

	for _, tt := range tests {
		got := tt.typ.String()
		if got != tt.want {
			t.Errorf("PointerType(%d).String() = %q, want %q", tt.typ, got, tt.want)
		}
	}
}

func TestDeviceKind_String(t *testing.T) {
	tests := []struct {
		kind DeviceKind
		want string
	}{
		{DeviceKindPrecise, "Precise"},
		{DeviceKindTouch, "Touch"},
		{DeviceKind(99), "Unknown"},
	}

	for _, tt := range tests {
		got := tt.kind.String()
		if got != tt.want {
			t.Errorf("DeviceKind(%d).String() = %q, want %q", tt.kind, got, tt.want)
		}
	}
}

func TestPointerType_DeviceKind(t *testing.T) {
	tests := []struct {
		typ  PointerType
		want DeviceKind
	}{
		{PointerTypeMouse, DeviceKindPrecise},
		{PointerTypeTouch, DeviceKindTouch},
		{PointerTypePen, DeviceKindTouch},
	}

	for _, tt := range tests {
		got := tt.typ.DeviceKind()
		if got != tt.want {
			t.Errorf("PointerType(%d).DeviceKind() = %v, want %v", tt.typ, got, tt.want)
		}
	}
}

func TestPointerEvent_String(t *testing.T) {
	ev := &PointerEvent{
		Base:        event.NewBase(event.TypeMouse, event.ModNone),
		EventType:   PointerDown,
		PointerID:   1,
		PointerType: PointerTypeMouse,
		Position:    geometry.Pt(10, 20),
	}

	s := ev.String()
	if s == "" {
		t.Error("String() should return non-empty string")
	}
}

func TestSlopForDevice(t *testing.T) {
	if got := SlopForDevice(DeviceKindPrecise); got != PrecisePointerSlop {
		t.Errorf("SlopForDevice(Precise) = %f, want %f", got, PrecisePointerSlop)
	}
	if got := SlopForDevice(DeviceKindTouch); got != TouchSlop {
		t.Errorf("SlopForDevice(Touch) = %f, want %f", got, TouchSlop)
	}
}
