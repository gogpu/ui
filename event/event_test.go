package event

import (
	"testing"
	"time"
)

func TestType_String(t *testing.T) {
	tests := []struct {
		name string
		typ  Type
		want string
	}{
		{"Mouse", TypeMouse, "Mouse"},
		{"Key", TypeKey, "Key"},
		{"Focus", TypeFocus, "Focus"},
		{"Wheel", TypeWheel, "Wheel"},
		{"Touch", TypeTouch, "Touch"},
		{"Text", TypeText, "Text"},
		{"Drop", TypeDrop, "Drop"},
		{"Resize", TypeResize, "Resize"},
		{"Unknown", Type(99), "Unknown"},
		{"Zero", Type(0), "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.typ.String(); got != tt.want {
				t.Errorf("Type.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewBase(t *testing.T) {
	tests := []struct {
		name      string
		eventType Type
		mods      Modifiers
	}{
		{"Mouse with no mods", TypeMouse, ModNone},
		{"Key with Ctrl", TypeKey, ModCtrl},
		{"Focus with Shift", TypeFocus, ModShift},
		{"Wheel with multiple mods", TypeWheel, ModCtrl | ModShift | ModAlt},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			b := NewBase(tt.eventType, tt.mods)
			after := time.Now()

			if b.Type() != tt.eventType {
				t.Errorf("Type() = %v, want %v", b.Type(), tt.eventType)
			}
			if b.Modifiers() != tt.mods {
				t.Errorf("Modifiers() = %v, want %v", b.Modifiers(), tt.mods)
			}
			if b.Time().Before(before) || b.Time().After(after) {
				t.Errorf("Time() = %v, want between %v and %v", b.Time(), before, after)
			}
			if b.Handled() {
				t.Error("Handled() = true, want false for new event")
			}
		})
	}
}

func TestNewBaseWithTime(t *testing.T) {
	fixedTime := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name      string
		eventType Type
		mods      Modifiers
		time      time.Time
	}{
		{"Mouse event", TypeMouse, ModNone, fixedTime},
		{"Key event", TypeKey, ModCtrl, fixedTime.Add(time.Hour)},
		{"Focus event", TypeFocus, ModShift, fixedTime.Add(-time.Hour)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBaseWithTime(tt.eventType, tt.time, tt.mods)

			if b.Type() != tt.eventType {
				t.Errorf("Type() = %v, want %v", b.Type(), tt.eventType)
			}
			if b.Modifiers() != tt.mods {
				t.Errorf("Modifiers() = %v, want %v", b.Modifiers(), tt.mods)
			}
			if !b.Time().Equal(tt.time) {
				t.Errorf("Time() = %v, want %v", b.Time(), tt.time)
			}
			if b.Handled() {
				t.Error("Handled() = true, want false for new event")
			}
		})
	}
}

func TestBase_SetHandled(t *testing.T) {
	b := NewBase(TypeMouse, ModNone)

	if b.Handled() {
		t.Error("Handled() should be false initially")
	}

	b.SetHandled()

	if !b.Handled() {
		t.Error("Handled() should be true after SetHandled()")
	}

	// Setting handled again should be idempotent
	b.SetHandled()
	if !b.Handled() {
		t.Error("Handled() should remain true")
	}
}

func TestBase_Type(t *testing.T) {
	tests := []struct {
		name      string
		eventType Type
	}{
		{"Mouse", TypeMouse},
		{"Key", TypeKey},
		{"Focus", TypeFocus},
		{"Wheel", TypeWheel},
		{"Touch", TypeTouch},
		{"Text", TypeText},
		{"Drop", TypeDrop},
		{"Resize", TypeResize},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBase(tt.eventType, ModNone)
			if got := b.Type(); got != tt.eventType {
				t.Errorf("Type() = %v, want %v", got, tt.eventType)
			}
		})
	}
}

func TestBase_Modifiers(t *testing.T) {
	tests := []struct {
		name string
		mods Modifiers
	}{
		{"None", ModNone},
		{"Shift", ModShift},
		{"Ctrl", ModCtrl},
		{"Alt", ModAlt},
		{"Super", ModSuper},
		{"Ctrl+Shift", ModCtrl | ModShift},
		{"All", ModCtrl | ModShift | ModAlt | ModSuper},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBase(TypeMouse, tt.mods)
			if got := b.Modifiers(); got != tt.mods {
				t.Errorf("Modifiers() = %v, want %v", got, tt.mods)
			}
		})
	}
}

// BenchmarkNewBase benchmarks creating a new Base event.
func BenchmarkNewBase(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = NewBase(TypeMouse, ModCtrl)
	}
}

// BenchmarkNewBaseWithTime benchmarks creating a Base event with specific time.
func BenchmarkNewBaseWithTime(b *testing.B) {
	t := time.Now()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewBaseWithTime(TypeKey, t, ModShift)
	}
}

// BenchmarkBase_Handled benchmarks checking handled state.
func BenchmarkBase_Handled(b *testing.B) {
	base := NewBase(TypeMouse, ModNone)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = base.Handled()
	}
}

// BenchmarkBase_SetHandled benchmarks setting handled state.
func BenchmarkBase_SetHandled(b *testing.B) {
	base := NewBase(TypeMouse, ModNone)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		base.handled = false // Reset for benchmark
		base.SetHandled()
	}
}

// BenchmarkType_String benchmarks getting event type string.
func BenchmarkType_String(b *testing.B) {
	typ := TypeMouse
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = typ.String()
	}
}
