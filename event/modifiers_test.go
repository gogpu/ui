package event

import (
	"testing"
)

func TestModifiers_Has(t *testing.T) {
	tests := []struct {
		name string
		m    Modifiers
		mod  Modifiers
		want bool
	}{
		{"None has None", ModNone, ModNone, true},
		{"Shift has Shift", ModShift, ModShift, true},
		{"Ctrl has Ctrl", ModCtrl, ModCtrl, true},
		{"Shift does not have Ctrl", ModShift, ModCtrl, false},
		{"Ctrl+Shift has Ctrl", ModCtrl | ModShift, ModCtrl, true},
		{"Ctrl+Shift has Shift", ModCtrl | ModShift, ModShift, true},
		{"Ctrl+Shift has Ctrl+Shift", ModCtrl | ModShift, ModCtrl | ModShift, true},
		{"Ctrl does not have Ctrl+Shift", ModCtrl, ModCtrl | ModShift, false},
		{"All modifiers has Alt", ModCtrl | ModShift | ModAlt | ModSuper, ModAlt, true},
		{"All modifiers has Ctrl+Alt", ModCtrl | ModShift | ModAlt | ModSuper, ModCtrl | ModAlt, true},
		{"None does not have Shift", ModNone, ModShift, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.Has(tt.mod); got != tt.want {
				t.Errorf("Has(%v) = %v, want %v", tt.mod, got, tt.want)
			}
		})
	}
}

func TestModifiers_HasAny(t *testing.T) {
	tests := []struct {
		name string
		m    Modifiers
		mod  Modifiers
		want bool
	}{
		{"None has any None", ModNone, ModNone, false},
		{"Shift has any Shift", ModShift, ModShift, true},
		{"Ctrl has any Ctrl+Shift", ModCtrl, ModCtrl | ModShift, true},
		{"Shift has any Ctrl+Shift", ModShift, ModCtrl | ModShift, true},
		{"None has any Ctrl+Shift", ModNone, ModCtrl | ModShift, false},
		{"Alt has any Ctrl+Shift", ModAlt, ModCtrl | ModShift, false},
		{"Ctrl+Alt has any Shift+Super", ModCtrl | ModAlt, ModShift | ModSuper, false},
		{"Ctrl+Alt has any Alt+Super", ModCtrl | ModAlt, ModAlt | ModSuper, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.HasAny(tt.mod); got != tt.want {
				t.Errorf("HasAny(%v) = %v, want %v", tt.mod, got, tt.want)
			}
		})
	}
}

func TestModifiers_IsShift(t *testing.T) {
	tests := []struct {
		name string
		m    Modifiers
		want bool
	}{
		{"None", ModNone, false},
		{"Shift", ModShift, true},
		{"Ctrl", ModCtrl, false},
		{"Ctrl+Shift", ModCtrl | ModShift, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.IsShift(); got != tt.want {
				t.Errorf("IsShift() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModifiers_IsCtrl(t *testing.T) {
	tests := []struct {
		name string
		m    Modifiers
		want bool
	}{
		{"None", ModNone, false},
		{"Ctrl", ModCtrl, true},
		{"Shift", ModShift, false},
		{"Ctrl+Shift", ModCtrl | ModShift, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.IsCtrl(); got != tt.want {
				t.Errorf("IsCtrl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModifiers_IsAlt(t *testing.T) {
	tests := []struct {
		name string
		m    Modifiers
		want bool
	}{
		{"None", ModNone, false},
		{"Alt", ModAlt, true},
		{"Shift", ModShift, false},
		{"Alt+Shift", ModAlt | ModShift, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.IsAlt(); got != tt.want {
				t.Errorf("IsAlt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModifiers_IsSuper(t *testing.T) {
	tests := []struct {
		name string
		m    Modifiers
		want bool
	}{
		{"None", ModNone, false},
		{"Super", ModSuper, true},
		{"Shift", ModShift, false},
		{"Super+Ctrl", ModSuper | ModCtrl, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.IsSuper(); got != tt.want {
				t.Errorf("IsSuper() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModifiers_IsCapsLock(t *testing.T) {
	tests := []struct {
		name string
		m    Modifiers
		want bool
	}{
		{"None", ModNone, false},
		{"CapsLock", ModCapsLock, true},
		{"Shift", ModShift, false},
		{"CapsLock+Shift", ModCapsLock | ModShift, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.IsCapsLock(); got != tt.want {
				t.Errorf("IsCapsLock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModifiers_IsNumLock(t *testing.T) {
	tests := []struct {
		name string
		m    Modifiers
		want bool
	}{
		{"None", ModNone, false},
		{"NumLock", ModNumLock, true},
		{"Shift", ModShift, false},
		{"NumLock+Ctrl", ModNumLock | ModCtrl, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.IsNumLock(); got != tt.want {
				t.Errorf("IsNumLock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModifiers_With(t *testing.T) {
	tests := []struct {
		name string
		m    Modifiers
		mod  Modifiers
		want Modifiers
	}{
		{"None with Shift", ModNone, ModShift, ModShift},
		{"Shift with Ctrl", ModShift, ModCtrl, ModShift | ModCtrl},
		{"Ctrl with Ctrl (idempotent)", ModCtrl, ModCtrl, ModCtrl},
		{"Multiple with Alt", ModCtrl | ModShift, ModAlt, ModCtrl | ModShift | ModAlt},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.With(tt.mod); got != tt.want {
				t.Errorf("With(%v) = %v, want %v", tt.mod, got, tt.want)
			}
		})
	}
}

func TestModifiers_Without(t *testing.T) {
	tests := []struct {
		name string
		m    Modifiers
		mod  Modifiers
		want Modifiers
	}{
		{"Shift without Shift", ModShift, ModShift, ModNone},
		{"Ctrl+Shift without Shift", ModCtrl | ModShift, ModShift, ModCtrl},
		{"Ctrl without Shift (no change)", ModCtrl, ModShift, ModCtrl},
		{"None without Ctrl (no change)", ModNone, ModCtrl, ModNone},
		{"All without Ctrl+Alt", ModCtrl | ModShift | ModAlt | ModSuper, ModCtrl | ModAlt, ModShift | ModSuper},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.Without(tt.mod); got != tt.want {
				t.Errorf("Without(%v) = %v, want %v", tt.mod, got, tt.want)
			}
		})
	}
}

func TestModifiers_String(t *testing.T) {
	tests := []struct {
		name string
		m    Modifiers
		want string
	}{
		{"None", ModNone, "None"},
		{"Shift", ModShift, "Shift"},
		{"Ctrl", ModCtrl, "Ctrl"},
		{"Alt", ModAlt, "Alt"},
		{"Super", ModSuper, "Super"},
		{"CapsLock", ModCapsLock, "CapsLock"},
		{"NumLock", ModNumLock, "NumLock"},
		{"Ctrl+Shift", ModCtrl | ModShift, "Ctrl+Shift"},
		{"Ctrl+Alt+Shift", ModCtrl | ModAlt | ModShift, "Ctrl+Alt+Shift"},
		{"All modifiers", ModCtrl | ModAlt | ModShift | ModSuper | ModCapsLock | ModNumLock, "Ctrl+Alt+Shift+Super+CapsLock+NumLock"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModifiers_Constants(t *testing.T) {
	// Verify that modifier constants are distinct bits
	if ModShift&ModCtrl != 0 {
		t.Error("ModShift and ModCtrl should not overlap")
	}
	if ModCtrl&ModAlt != 0 {
		t.Error("ModCtrl and ModAlt should not overlap")
	}
	if ModAlt&ModSuper != 0 {
		t.Error("ModAlt and ModSuper should not overlap")
	}
	if ModSuper&ModCapsLock != 0 {
		t.Error("ModSuper and ModCapsLock should not overlap")
	}
	if ModCapsLock&ModNumLock != 0 {
		t.Error("ModCapsLock and ModNumLock should not overlap")
	}

	// Verify ModNone is zero
	if ModNone != 0 {
		t.Errorf("ModNone should be 0, got %v", ModNone)
	}
}

// BenchmarkModifiers_Has benchmarks checking modifier presence.
func BenchmarkModifiers_Has(b *testing.B) {
	mods := ModCtrl | ModShift | ModAlt
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = mods.Has(ModCtrl)
	}
}

// BenchmarkModifiers_HasAny benchmarks checking any modifier presence.
func BenchmarkModifiers_HasAny(b *testing.B) {
	mods := ModCtrl | ModShift
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = mods.HasAny(ModAlt | ModSuper)
	}
}

// BenchmarkModifiers_String benchmarks getting modifier string.
func BenchmarkModifiers_String(b *testing.B) {
	mods := ModCtrl | ModShift | ModAlt
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = mods.String()
	}
}

// BenchmarkModifiers_With benchmarks adding modifiers.
func BenchmarkModifiers_With(b *testing.B) {
	mods := ModCtrl
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = mods.With(ModShift)
	}
}

// BenchmarkModifiers_Without benchmarks removing modifiers.
func BenchmarkModifiers_Without(b *testing.B) {
	mods := ModCtrl | ModShift | ModAlt
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = mods.Without(ModShift)
	}
}
