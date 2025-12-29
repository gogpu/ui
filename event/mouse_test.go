package event

import (
	"strings"
	"testing"
	"time"

	"github.com/gogpu/ui/geometry"
)

func TestMouseEventType_String(t *testing.T) {
	tests := []struct {
		name string
		typ  MouseEventType
		want string
	}{
		{"Press", MousePress, "Press"},
		{"Release", MouseRelease, "Release"},
		{"Move", MouseMove, "Move"},
		{"Enter", MouseEnter, "Enter"},
		{"Leave", MouseLeave, "Leave"},
		{"Drag", MouseDrag, "Drag"},
		{"DoubleClick", MouseDoubleClick, "DoubleClick"},
		{"Unknown", MouseEventType(99), "Unknown"},
		{"Zero", MouseEventType(0), "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.typ.String(); got != tt.want {
				t.Errorf("MouseEventType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestButton_String(t *testing.T) {
	tests := []struct {
		name string
		b    Button
		want string
	}{
		{"None", ButtonNone, "None"},
		{"Left", ButtonLeft, "Left"},
		{"Right", ButtonRight, "Right"},
		{"Middle", ButtonMiddle, "Middle"},
		{"X1", ButtonX1, "X1"},
		{"X2", ButtonX2, "X2"},
		{"Unknown", Button(99), "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b.String(); got != tt.want {
				t.Errorf("Button.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestButtonState_Has(t *testing.T) {
	tests := []struct {
		name  string
		state ButtonState
		check ButtonState
		want  bool
	}{
		{"Empty has none", 0, 0, true},
		{"Left has Left", ButtonStateLeft, ButtonStateLeft, true},
		{"Left+Right has Left", ButtonStateLeft | ButtonStateRight, ButtonStateLeft, true},
		{"Left+Right has Right", ButtonStateLeft | ButtonStateRight, ButtonStateRight, true},
		{"Left does not have Right", ButtonStateLeft, ButtonStateRight, false},
		{"Empty does not have Left", 0, ButtonStateLeft, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.Has(tt.check); got != tt.want {
				t.Errorf("Has(%v) = %v, want %v", tt.check, got, tt.want)
			}
		})
	}
}

func TestButtonState_IsLeftPressed(t *testing.T) {
	tests := []struct {
		name  string
		state ButtonState
		want  bool
	}{
		{"Empty", 0, false},
		{"Left", ButtonStateLeft, true},
		{"Right", ButtonStateRight, false},
		{"Left+Right", ButtonStateLeft | ButtonStateRight, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.IsLeftPressed(); got != tt.want {
				t.Errorf("IsLeftPressed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestButtonState_IsRightPressed(t *testing.T) {
	tests := []struct {
		name  string
		state ButtonState
		want  bool
	}{
		{"Empty", 0, false},
		{"Right", ButtonStateRight, true},
		{"Left", ButtonStateLeft, false},
		{"Left+Right", ButtonStateLeft | ButtonStateRight, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.IsRightPressed(); got != tt.want {
				t.Errorf("IsRightPressed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestButtonState_IsMiddlePressed(t *testing.T) {
	tests := []struct {
		name  string
		state ButtonState
		want  bool
	}{
		{"Empty", 0, false},
		{"Middle", ButtonStateMiddle, true},
		{"Left", ButtonStateLeft, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.IsMiddlePressed(); got != tt.want {
				t.Errorf("IsMiddlePressed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestButtonState_IsX1Pressed(t *testing.T) {
	tests := []struct {
		name  string
		state ButtonState
		want  bool
	}{
		{"Empty", 0, false},
		{"X1", ButtonStateX1, true},
		{"Left", ButtonStateLeft, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.IsX1Pressed(); got != tt.want {
				t.Errorf("IsX1Pressed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestButtonState_IsX2Pressed(t *testing.T) {
	tests := []struct {
		name  string
		state ButtonState
		want  bool
	}{
		{"Empty", 0, false},
		{"X2", ButtonStateX2, true},
		{"Left", ButtonStateLeft, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.IsX2Pressed(); got != tt.want {
				t.Errorf("IsX2Pressed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestButtonState_AnyPressed(t *testing.T) {
	tests := []struct {
		name  string
		state ButtonState
		want  bool
	}{
		{"Empty", 0, false},
		{"Left", ButtonStateLeft, true},
		{"Right", ButtonStateRight, true},
		{"Middle", ButtonStateMiddle, true},
		{"Multiple", ButtonStateLeft | ButtonStateRight, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.AnyPressed(); got != tt.want {
				t.Errorf("AnyPressed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewMouseEvent(t *testing.T) {
	pos := geometry.Pt(100, 200)
	globalPos := geometry.Pt(500, 600)

	tests := []struct {
		name      string
		mouseType MouseEventType
		button    Button
		buttons   ButtonState
		mods      Modifiers
	}{
		{"Press Left", MousePress, ButtonLeft, ButtonStateLeft, ModNone},
		{"Release Right", MouseRelease, ButtonRight, 0, ModShift},
		{"Move", MouseMove, ButtonNone, 0, ModNone},
		{"Drag with Ctrl", MouseDrag, ButtonLeft, ButtonStateLeft, ModCtrl},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			e := NewMouseEvent(tt.mouseType, tt.button, tt.buttons, pos, globalPos, tt.mods)
			after := time.Now()

			if e.Type() != TypeMouse {
				t.Errorf("Type() = %v, want %v", e.Type(), TypeMouse)
			}
			if e.MouseType != tt.mouseType {
				t.Errorf("MouseType = %v, want %v", e.MouseType, tt.mouseType)
			}
			if e.Button != tt.button {
				t.Errorf("Button = %v, want %v", e.Button, tt.button)
			}
			if e.Buttons != tt.buttons {
				t.Errorf("Buttons = %v, want %v", e.Buttons, tt.buttons)
			}
			if e.Position != pos {
				t.Errorf("Position = %v, want %v", e.Position, pos)
			}
			if e.GlobalPosition != globalPos {
				t.Errorf("GlobalPosition = %v, want %v", e.GlobalPosition, globalPos)
			}
			if e.Modifiers() != tt.mods {
				t.Errorf("Modifiers() = %v, want %v", e.Modifiers(), tt.mods)
			}
			if e.ClickCount != 1 {
				t.Errorf("ClickCount = %v, want 1", e.ClickCount)
			}
			if e.Time().Before(before) || e.Time().After(after) {
				t.Errorf("Time() = %v, want between %v and %v", e.Time(), before, after)
			}
		})
	}
}

func TestNewMouseEventWithTime(t *testing.T) {
	pos := geometry.Pt(100, 200)
	globalPos := geometry.Pt(500, 600)
	fixedTime := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

	e := NewMouseEventWithTime(MousePress, ButtonLeft, ButtonStateLeft, pos, globalPos, ModCtrl, fixedTime)

	if !e.Time().Equal(fixedTime) {
		t.Errorf("Time() = %v, want %v", e.Time(), fixedTime)
	}
	if e.Type() != TypeMouse {
		t.Errorf("Type() = %v, want %v", e.Type(), TypeMouse)
	}
}

func TestMouseEvent_IsPress(t *testing.T) {
	tests := []struct {
		name      string
		mouseType MouseEventType
		want      bool
	}{
		{"Press", MousePress, true},
		{"Release", MouseRelease, false},
		{"Move", MouseMove, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewMouseEvent(tt.mouseType, ButtonLeft, 0, geometry.Point{}, geometry.Point{}, ModNone)
			if got := e.IsPress(); got != tt.want {
				t.Errorf("IsPress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMouseEvent_IsRelease(t *testing.T) {
	tests := []struct {
		name      string
		mouseType MouseEventType
		want      bool
	}{
		{"Release", MouseRelease, true},
		{"Press", MousePress, false},
		{"Move", MouseMove, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewMouseEvent(tt.mouseType, ButtonLeft, 0, geometry.Point{}, geometry.Point{}, ModNone)
			if got := e.IsRelease(); got != tt.want {
				t.Errorf("IsRelease() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMouseEvent_IsMove(t *testing.T) {
	tests := []struct {
		name      string
		mouseType MouseEventType
		want      bool
	}{
		{"Move", MouseMove, true},
		{"Press", MousePress, false},
		{"Drag", MouseDrag, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewMouseEvent(tt.mouseType, ButtonNone, 0, geometry.Point{}, geometry.Point{}, ModNone)
			if got := e.IsMove(); got != tt.want {
				t.Errorf("IsMove() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMouseEvent_IsEnter(t *testing.T) {
	e := NewMouseEvent(MouseEnter, ButtonNone, 0, geometry.Point{}, geometry.Point{}, ModNone)
	if !e.IsEnter() {
		t.Error("IsEnter() = false, want true")
	}
	e2 := NewMouseEvent(MouseLeave, ButtonNone, 0, geometry.Point{}, geometry.Point{}, ModNone)
	if e2.IsEnter() {
		t.Error("IsEnter() = true for Leave event, want false")
	}
}

func TestMouseEvent_IsLeave(t *testing.T) {
	e := NewMouseEvent(MouseLeave, ButtonNone, 0, geometry.Point{}, geometry.Point{}, ModNone)
	if !e.IsLeave() {
		t.Error("IsLeave() = false, want true")
	}
	e2 := NewMouseEvent(MouseEnter, ButtonNone, 0, geometry.Point{}, geometry.Point{}, ModNone)
	if e2.IsLeave() {
		t.Error("IsLeave() = true for Enter event, want false")
	}
}

func TestMouseEvent_IsDrag(t *testing.T) {
	e := NewMouseEvent(MouseDrag, ButtonLeft, ButtonStateLeft, geometry.Point{}, geometry.Point{}, ModNone)
	if !e.IsDrag() {
		t.Error("IsDrag() = false, want true")
	}
	e2 := NewMouseEvent(MouseMove, ButtonNone, 0, geometry.Point{}, geometry.Point{}, ModNone)
	if e2.IsDrag() {
		t.Error("IsDrag() = true for Move event, want false")
	}
}

func TestMouseEvent_IsDoubleClick(t *testing.T) {
	e := NewMouseEvent(MouseDoubleClick, ButtonLeft, 0, geometry.Point{}, geometry.Point{}, ModNone)
	if !e.IsDoubleClick() {
		t.Error("IsDoubleClick() = false, want true")
	}
	e2 := NewMouseEvent(MousePress, ButtonLeft, ButtonStateLeft, geometry.Point{}, geometry.Point{}, ModNone)
	if e2.IsDoubleClick() {
		t.Error("IsDoubleClick() = true for Press event, want false")
	}
}

func TestMouseEvent_IsLeftButton(t *testing.T) {
	tests := []struct {
		name   string
		button Button
		want   bool
	}{
		{"Left", ButtonLeft, true},
		{"Right", ButtonRight, false},
		{"Middle", ButtonMiddle, false},
		{"None", ButtonNone, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewMouseEvent(MousePress, tt.button, 0, geometry.Point{}, geometry.Point{}, ModNone)
			if got := e.IsLeftButton(); got != tt.want {
				t.Errorf("IsLeftButton() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMouseEvent_IsRightButton(t *testing.T) {
	tests := []struct {
		name   string
		button Button
		want   bool
	}{
		{"Right", ButtonRight, true},
		{"Left", ButtonLeft, false},
		{"Middle", ButtonMiddle, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewMouseEvent(MousePress, tt.button, 0, geometry.Point{}, geometry.Point{}, ModNone)
			if got := e.IsRightButton(); got != tt.want {
				t.Errorf("IsRightButton() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMouseEvent_IsMiddleButton(t *testing.T) {
	tests := []struct {
		name   string
		button Button
		want   bool
	}{
		{"Middle", ButtonMiddle, true},
		{"Left", ButtonLeft, false},
		{"Right", ButtonRight, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewMouseEvent(MousePress, tt.button, 0, geometry.Point{}, geometry.Point{}, ModNone)
			if got := e.IsMiddleButton(); got != tt.want {
				t.Errorf("IsMiddleButton() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMouseEvent_String(t *testing.T) {
	e := NewMouseEvent(MousePress, ButtonLeft, ButtonStateLeft, geometry.Pt(100, 200), geometry.Pt(500, 600), ModCtrl|ModShift)

	s := e.String()
	if !strings.Contains(s, "MouseEvent") {
		t.Errorf("String() should contain 'MouseEvent', got %v", s)
	}
	if !strings.Contains(s, "Press") {
		t.Errorf("String() should contain 'Press', got %v", s)
	}
	if !strings.Contains(s, "Left") {
		t.Errorf("String() should contain 'Left', got %v", s)
	}
}

func TestMouseEvent_ImplementsEvent(t *testing.T) {
	var e Event = NewMouseEvent(MousePress, ButtonLeft, 0, geometry.Point{}, geometry.Point{}, ModNone)

	if e.Type() != TypeMouse {
		t.Errorf("Type() = %v, want %v", e.Type(), TypeMouse)
	}
	if e.Handled() {
		t.Error("Handled() should be false initially")
	}
	e.SetHandled()
	if !e.Handled() {
		t.Error("Handled() should be true after SetHandled()")
	}
}

// BenchmarkNewMouseEvent benchmarks creating a new MouseEvent.
func BenchmarkNewMouseEvent(b *testing.B) {
	pos := geometry.Pt(100, 200)
	globalPos := geometry.Pt(500, 600)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = NewMouseEvent(MousePress, ButtonLeft, ButtonStateLeft, pos, globalPos, ModCtrl)
	}
}

// BenchmarkMouseEvent_IsPress benchmarks checking press type.
func BenchmarkMouseEvent_IsPress(b *testing.B) {
	e := NewMouseEvent(MousePress, ButtonLeft, 0, geometry.Point{}, geometry.Point{}, ModNone)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.IsPress()
	}
}

// BenchmarkMouseEvent_String benchmarks string conversion.
func BenchmarkMouseEvent_String(b *testing.B) {
	e := NewMouseEvent(MousePress, ButtonLeft, ButtonStateLeft, geometry.Pt(100, 200), geometry.Pt(500, 600), ModCtrl|ModShift)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = e.String()
	}
}

// BenchmarkButtonState_Has benchmarks checking button state.
func BenchmarkButtonState_Has(b *testing.B) {
	state := ButtonStateLeft | ButtonStateRight | ButtonStateMiddle
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = state.Has(ButtonStateLeft)
	}
}
