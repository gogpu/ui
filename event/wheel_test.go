package event

import (
	"strings"
	"testing"
	"time"

	"github.com/gogpu/ui/geometry"
)

func TestNewWheelEvent(t *testing.T) {
	delta := geometry.Pt(1, -3)
	pos := geometry.Pt(100, 200)
	globalPos := geometry.Pt(500, 600)

	tests := []struct {
		name string
		mods Modifiers
	}{
		{"No modifiers", ModNone},
		{"With Ctrl", ModCtrl},
		{"With Shift", ModShift},
		{"With Ctrl+Shift", ModCtrl | ModShift},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			e := NewWheelEvent(delta, pos, globalPos, tt.mods)
			after := time.Now()

			if e.Type() != TypeWheel {
				t.Errorf("Type() = %v, want %v", e.Type(), TypeWheel)
			}
			if e.Delta != delta {
				t.Errorf("Delta = %v, want %v", e.Delta, delta)
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
			if e.Time().Before(before) || e.Time().After(after) {
				t.Errorf("Time() = %v, want between %v and %v", e.Time(), before, after)
			}
			if e.Handled() {
				t.Error("Handled() should be false initially")
			}
		})
	}
}

func TestNewWheelEventWithTime(t *testing.T) {
	delta := geometry.Pt(0, 3)
	pos := geometry.Pt(100, 200)
	globalPos := geometry.Pt(500, 600)
	fixedTime := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

	e := NewWheelEventWithTime(delta, pos, globalPos, ModCtrl, fixedTime)

	if !e.Time().Equal(fixedTime) {
		t.Errorf("Time() = %v, want %v", e.Time(), fixedTime)
	}
	if e.Type() != TypeWheel {
		t.Errorf("Type() = %v, want %v", e.Type(), TypeWheel)
	}
	if e.Delta != delta {
		t.Errorf("Delta = %v, want %v", e.Delta, delta)
	}
}

func TestWheelEvent_DeltaX(t *testing.T) {
	tests := []struct {
		name  string
		delta geometry.Point
		want  float32
	}{
		{"Positive X", geometry.Pt(5, 0), 5},
		{"Negative X", geometry.Pt(-3, 0), -3},
		{"Zero X", geometry.Pt(0, 10), 0},
		{"Float X", geometry.Pt(2.5, 0), 2.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewWheelEvent(tt.delta, geometry.Point{}, geometry.Point{}, ModNone)
			if got := e.DeltaX(); got != tt.want {
				t.Errorf("DeltaX() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWheelEvent_DeltaY(t *testing.T) {
	tests := []struct {
		name  string
		delta geometry.Point
		want  float32
	}{
		{"Positive Y", geometry.Pt(0, 5), 5},
		{"Negative Y", geometry.Pt(0, -3), -3},
		{"Zero Y", geometry.Pt(10, 0), 0},
		{"Float Y", geometry.Pt(0, 2.5), 2.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewWheelEvent(tt.delta, geometry.Point{}, geometry.Point{}, ModNone)
			if got := e.DeltaY(); got != tt.want {
				t.Errorf("DeltaY() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWheelEvent_IsScrollUp(t *testing.T) {
	tests := []struct {
		name  string
		delta geometry.Point
		want  bool
	}{
		{"Scroll up", geometry.Pt(0, 3), true},
		{"Scroll down", geometry.Pt(0, -3), false},
		{"No vertical scroll", geometry.Pt(5, 0), false},
		{"Zero delta", geometry.Pt(0, 0), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewWheelEvent(tt.delta, geometry.Point{}, geometry.Point{}, ModNone)
			if got := e.IsScrollUp(); got != tt.want {
				t.Errorf("IsScrollUp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWheelEvent_IsScrollDown(t *testing.T) {
	tests := []struct {
		name  string
		delta geometry.Point
		want  bool
	}{
		{"Scroll down", geometry.Pt(0, -3), true},
		{"Scroll up", geometry.Pt(0, 3), false},
		{"No vertical scroll", geometry.Pt(5, 0), false},
		{"Zero delta", geometry.Pt(0, 0), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewWheelEvent(tt.delta, geometry.Point{}, geometry.Point{}, ModNone)
			if got := e.IsScrollDown(); got != tt.want {
				t.Errorf("IsScrollDown() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWheelEvent_IsScrollLeft(t *testing.T) {
	tests := []struct {
		name  string
		delta geometry.Point
		want  bool
	}{
		{"Scroll left", geometry.Pt(-3, 0), true},
		{"Scroll right", geometry.Pt(3, 0), false},
		{"No horizontal scroll", geometry.Pt(0, 5), false},
		{"Zero delta", geometry.Pt(0, 0), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewWheelEvent(tt.delta, geometry.Point{}, geometry.Point{}, ModNone)
			if got := e.IsScrollLeft(); got != tt.want {
				t.Errorf("IsScrollLeft() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWheelEvent_IsScrollRight(t *testing.T) {
	tests := []struct {
		name  string
		delta geometry.Point
		want  bool
	}{
		{"Scroll right", geometry.Pt(3, 0), true},
		{"Scroll left", geometry.Pt(-3, 0), false},
		{"No horizontal scroll", geometry.Pt(0, 5), false},
		{"Zero delta", geometry.Pt(0, 0), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewWheelEvent(tt.delta, geometry.Point{}, geometry.Point{}, ModNone)
			if got := e.IsScrollRight(); got != tt.want {
				t.Errorf("IsScrollRight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWheelEvent_IsHorizontal(t *testing.T) {
	tests := []struct {
		name  string
		delta geometry.Point
		want  bool
	}{
		{"Horizontal positive", geometry.Pt(3, 0), true},
		{"Horizontal negative", geometry.Pt(-3, 0), true},
		{"Vertical only", geometry.Pt(0, 5), false},
		{"Both", geometry.Pt(1, 1), true},
		{"Zero", geometry.Pt(0, 0), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewWheelEvent(tt.delta, geometry.Point{}, geometry.Point{}, ModNone)
			if got := e.IsHorizontal(); got != tt.want {
				t.Errorf("IsHorizontal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWheelEvent_IsVertical(t *testing.T) {
	tests := []struct {
		name  string
		delta geometry.Point
		want  bool
	}{
		{"Vertical positive", geometry.Pt(0, 3), true},
		{"Vertical negative", geometry.Pt(0, -3), true},
		{"Horizontal only", geometry.Pt(5, 0), false},
		{"Both", geometry.Pt(1, 1), true},
		{"Zero", geometry.Pt(0, 0), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewWheelEvent(tt.delta, geometry.Point{}, geometry.Point{}, ModNone)
			if got := e.IsVertical(); got != tt.want {
				t.Errorf("IsVertical() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWheelEvent_String(t *testing.T) {
	e := NewWheelEvent(geometry.Pt(1, -3), geometry.Pt(100, 200), geometry.Pt(500, 600), ModCtrl|ModShift)

	s := e.String()
	contains := []string{"WheelEvent", "Delta", "Position"}
	for _, c := range contains {
		if !strings.Contains(s, c) {
			t.Errorf("String() = %v, should contain %v", s, c)
		}
	}
}

func TestWheelEvent_SetHandled(t *testing.T) {
	e := NewWheelEvent(geometry.Pt(0, 1), geometry.Point{}, geometry.Point{}, ModNone)

	if e.Handled() {
		t.Error("Handled() should be false initially")
	}

	e.SetHandled()

	if !e.Handled() {
		t.Error("Handled() should be true after SetHandled()")
	}
}

func TestWheelEvent_ImplementsEvent(t *testing.T) {
	var e Event = NewWheelEvent(geometry.Pt(0, 1), geometry.Point{}, geometry.Point{}, ModNone)

	if e.Type() != TypeWheel {
		t.Errorf("Type() = %v, want %v", e.Type(), TypeWheel)
	}
	if e.Handled() {
		t.Error("Handled() should be false initially")
	}
	e.SetHandled()
	if !e.Handled() {
		t.Error("Handled() should be true after SetHandled()")
	}
}

func TestWheelEvent_Position(t *testing.T) {
	pos := geometry.Pt(150, 250)
	globalPos := geometry.Pt(550, 650)

	e := NewWheelEvent(geometry.Pt(0, 1), pos, globalPos, ModNone)

	if e.Position != pos {
		t.Errorf("Position = %v, want %v", e.Position, pos)
	}
	if e.GlobalPosition != globalPos {
		t.Errorf("GlobalPosition = %v, want %v", e.GlobalPosition, globalPos)
	}
}

func TestWheelEvent_Modifiers(t *testing.T) {
	tests := []struct {
		name string
		mods Modifiers
	}{
		{"None", ModNone},
		{"Ctrl", ModCtrl},
		{"Shift", ModShift},
		{"Alt", ModAlt},
		{"Super", ModSuper},
		{"All", ModCtrl | ModShift | ModAlt | ModSuper},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewWheelEvent(geometry.Pt(0, 1), geometry.Point{}, geometry.Point{}, tt.mods)
			if got := e.Modifiers(); got != tt.mods {
				t.Errorf("Modifiers() = %v, want %v", got, tt.mods)
			}
		})
	}
}

// BenchmarkNewWheelEvent benchmarks creating a new WheelEvent.
func BenchmarkNewWheelEvent(b *testing.B) {
	delta := geometry.Pt(0, 3)
	pos := geometry.Pt(100, 200)
	globalPos := geometry.Pt(500, 600)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = NewWheelEvent(delta, pos, globalPos, ModCtrl)
	}
}

// BenchmarkNewWheelEventWithTime benchmarks creating a WheelEvent with time.
func BenchmarkNewWheelEventWithTime(b *testing.B) {
	delta := geometry.Pt(0, 3)
	pos := geometry.Pt(100, 200)
	globalPos := geometry.Pt(500, 600)
	t := time.Now()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewWheelEventWithTime(delta, pos, globalPos, ModCtrl, t)
	}
}

// BenchmarkWheelEvent_IsScrollUp benchmarks checking scroll up.
func BenchmarkWheelEvent_IsScrollUp(b *testing.B) {
	e := NewWheelEvent(geometry.Pt(0, 3), geometry.Point{}, geometry.Point{}, ModNone)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.IsScrollUp()
	}
}

// BenchmarkWheelEvent_DeltaY benchmarks getting delta Y.
func BenchmarkWheelEvent_DeltaY(b *testing.B) {
	e := NewWheelEvent(geometry.Pt(1, 3), geometry.Point{}, geometry.Point{}, ModNone)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.DeltaY()
	}
}

// BenchmarkWheelEvent_String benchmarks string conversion.
func BenchmarkWheelEvent_String(b *testing.B) {
	e := NewWheelEvent(geometry.Pt(1, -3), geometry.Pt(100, 200), geometry.Pt(500, 600), ModCtrl|ModShift)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = e.String()
	}
}

// BenchmarkWheelEvent_IsHorizontal benchmarks checking horizontal scroll.
func BenchmarkWheelEvent_IsHorizontal(b *testing.B) {
	e := NewWheelEvent(geometry.Pt(3, 0), geometry.Point{}, geometry.Point{}, ModNone)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.IsHorizontal()
	}
}
