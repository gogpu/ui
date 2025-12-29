package event

import (
	"strings"
	"testing"
	"time"
)

func TestFocusEventType_String(t *testing.T) {
	tests := []struct {
		name string
		typ  FocusEventType
		want string
	}{
		{"Gained", FocusGained, "Gained"},
		{"Lost", FocusLost, "Lost"},
		{"Unknown", FocusEventType(99), "Unknown"},
		{"Zero", FocusEventType(0), "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.typ.String(); got != tt.want {
				t.Errorf("FocusEventType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewFocusEvent(t *testing.T) {
	tests := []struct {
		name      string
		focusType FocusEventType
	}{
		{"Gained", FocusGained},
		{"Lost", FocusLost},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			e := NewFocusEvent(tt.focusType)
			after := time.Now()

			if e.Type() != TypeFocus {
				t.Errorf("Type() = %v, want %v", e.Type(), TypeFocus)
			}
			if e.FocusType != tt.focusType {
				t.Errorf("FocusType = %v, want %v", e.FocusType, tt.focusType)
			}
			if e.Modifiers() != ModNone {
				t.Errorf("Modifiers() = %v, want %v", e.Modifiers(), ModNone)
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

func TestNewFocusEventWithTime(t *testing.T) {
	fixedTime := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name      string
		focusType FocusEventType
		time      time.Time
	}{
		{"Gained", FocusGained, fixedTime},
		{"Lost", FocusLost, fixedTime.Add(time.Hour)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewFocusEventWithTime(tt.focusType, tt.time)

			if e.Type() != TypeFocus {
				t.Errorf("Type() = %v, want %v", e.Type(), TypeFocus)
			}
			if e.FocusType != tt.focusType {
				t.Errorf("FocusType = %v, want %v", e.FocusType, tt.focusType)
			}
			if !e.Time().Equal(tt.time) {
				t.Errorf("Time() = %v, want %v", e.Time(), tt.time)
			}
		})
	}
}

func TestFocusEvent_IsGained(t *testing.T) {
	tests := []struct {
		name      string
		focusType FocusEventType
		want      bool
	}{
		{"Gained", FocusGained, true},
		{"Lost", FocusLost, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewFocusEvent(tt.focusType)
			if got := e.IsGained(); got != tt.want {
				t.Errorf("IsGained() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFocusEvent_IsLost(t *testing.T) {
	tests := []struct {
		name      string
		focusType FocusEventType
		want      bool
	}{
		{"Lost", FocusLost, true},
		{"Gained", FocusGained, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewFocusEvent(tt.focusType)
			if got := e.IsLost(); got != tt.want {
				t.Errorf("IsLost() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFocusEvent_String(t *testing.T) {
	tests := []struct {
		name      string
		focusType FocusEventType
		contains  []string
	}{
		{"Gained", FocusGained, []string{"FocusEvent", "Gained"}},
		{"Lost", FocusLost, []string{"FocusEvent", "Lost"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewFocusEvent(tt.focusType)
			s := e.String()
			for _, c := range tt.contains {
				if !strings.Contains(s, c) {
					t.Errorf("String() = %v, should contain %v", s, c)
				}
			}
		})
	}
}

func TestFocusEvent_SetHandled(t *testing.T) {
	e := NewFocusEvent(FocusGained)

	if e.Handled() {
		t.Error("Handled() should be false initially")
	}

	e.SetHandled()

	if !e.Handled() {
		t.Error("Handled() should be true after SetHandled()")
	}
}

func TestFocusEvent_ImplementsEvent(t *testing.T) {
	var e Event = NewFocusEvent(FocusGained)

	if e.Type() != TypeFocus {
		t.Errorf("Type() = %v, want %v", e.Type(), TypeFocus)
	}
	if e.Handled() {
		t.Error("Handled() should be false initially")
	}
	if e.Modifiers() != ModNone {
		t.Errorf("Modifiers() = %v, want %v", e.Modifiers(), ModNone)
	}
	e.SetHandled()
	if !e.Handled() {
		t.Error("Handled() should be true after SetHandled()")
	}
}

func TestFocusEvent_Time(t *testing.T) {
	before := time.Now()
	e := NewFocusEvent(FocusGained)
	after := time.Now()

	if e.Time().Before(before) {
		t.Errorf("Time() = %v is before creation time %v", e.Time(), before)
	}
	if e.Time().After(after) {
		t.Errorf("Time() = %v is after creation time %v", e.Time(), after)
	}
}

// BenchmarkNewFocusEvent benchmarks creating a new FocusEvent.
func BenchmarkNewFocusEvent(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = NewFocusEvent(FocusGained)
	}
}

// BenchmarkNewFocusEventWithTime benchmarks creating a FocusEvent with time.
func BenchmarkNewFocusEventWithTime(b *testing.B) {
	t := time.Now()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewFocusEventWithTime(FocusLost, t)
	}
}

// BenchmarkFocusEvent_IsGained benchmarks checking focus gained.
func BenchmarkFocusEvent_IsGained(b *testing.B) {
	e := NewFocusEvent(FocusGained)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.IsGained()
	}
}

// BenchmarkFocusEvent_IsLost benchmarks checking focus lost.
func BenchmarkFocusEvent_IsLost(b *testing.B) {
	e := NewFocusEvent(FocusLost)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.IsLost()
	}
}

// BenchmarkFocusEvent_String benchmarks string conversion.
func BenchmarkFocusEvent_String(b *testing.B) {
	e := NewFocusEvent(FocusGained)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = e.String()
	}
}

// BenchmarkFocusEventType_String benchmarks getting focus type string.
func BenchmarkFocusEventType_String(b *testing.B) {
	typ := FocusGained
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = typ.String()
	}
}
