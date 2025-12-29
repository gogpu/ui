package event

import (
	"strings"
	"testing"
	"time"
)

func TestKeyEventType_String(t *testing.T) {
	tests := []struct {
		name string
		typ  KeyEventType
		want string
	}{
		{"Press", KeyPress, "Press"},
		{"Release", KeyRelease, "Release"},
		{"Repeat", KeyRepeat, "Repeat"},
		{"Unknown", KeyEventType(99), "Unknown"},
		{"Zero", KeyEventType(0), "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.typ.String(); got != tt.want {
				t.Errorf("KeyEventType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKey_String(t *testing.T) {
	tests := []struct {
		name string
		k    Key
		want string
	}{
		{"Unknown", KeyUnknown, "Unknown"},
		{"A", KeyA, "A"},
		{"Z", KeyZ, "Z"},
		{"0", Key0, "0"},
		{"9", Key9, "9"},
		{"F1", KeyF1, "F1"},
		{"F12", KeyF12, "F12"},
		{"F24", KeyF24, "F24"},
		{"Up", KeyUp, "Up"},
		{"Down", KeyDown, "Down"},
		{"Left", KeyLeft, "Left"},
		{"Right", KeyRight, "Right"},
		{"Enter", KeyEnter, "Enter"},
		{"Tab", KeyTab, "Tab"},
		{"Backspace", KeyBackspace, "Backspace"},
		{"Delete", KeyDelete, "Delete"},
		{"Escape", KeyEscape, "Escape"},
		{"Space", KeySpace, "Space"},
		{"LeftShift", KeyLeftShift, "LeftShift"},
		{"RightShift", KeyRightShift, "RightShift"},
		{"LeftCtrl", KeyLeftCtrl, "LeftCtrl"},
		{"RightCtrl", KeyRightCtrl, "RightCtrl"},
		{"LeftAlt", KeyLeftAlt, "LeftAlt"},
		{"RightAlt", KeyRightAlt, "RightAlt"},
		{"LeftSuper", KeyLeftSuper, "LeftSuper"},
		{"RightSuper", KeyRightSuper, "RightSuper"},
		{"CapsLock", KeyCapsLock, "CapsLock"},
		{"NumLock", KeyNumLock, "NumLock"},
		{"ScrollLock", KeyScrollLock, "ScrollLock"},
		{"Numpad0", KeyNumpad0, "Numpad0"},
		{"Numpad9", KeyNumpad9, "Numpad9"},
		{"NumpadDecimal", KeyNumpadDecimal, "NumpadDecimal"},
		{"NumpadEnter", KeyNumpadEnter, "NumpadEnter"},
		{"NumpadAdd", KeyNumpadAdd, "NumpadAdd"},
		{"NumpadSubtract", KeyNumpadSubtract, "NumpadSubtract"},
		{"NumpadMultiply", KeyNumpadMultiply, "NumpadMultiply"},
		{"NumpadDivide", KeyNumpadDivide, "NumpadDivide"},
		{"Minus", KeyMinus, "Minus"},
		{"Equal", KeyEqual, "Equal"},
		{"LeftBracket", KeyLeftBracket, "LeftBracket"},
		{"RightBracket", KeyRightBracket, "RightBracket"},
		{"Backslash", KeyBackslash, "Backslash"},
		{"Semicolon", KeySemicolon, "Semicolon"},
		{"Apostrophe", KeyApostrophe, "Apostrophe"},
		{"Grave", KeyGrave, "Grave"},
		{"Comma", KeyComma, "Comma"},
		{"Period", KeyPeriod, "Period"},
		{"Slash", KeySlash, "Slash"},
		{"PrintScreen", KeyPrintScreen, "PrintScreen"},
		{"Pause", KeyPause, "Pause"},
		{"Menu", KeyMenu, "Menu"},
		{"Mute", KeyMute, "Mute"},
		{"VolumeUp", KeyVolumeUp, "VolumeUp"},
		{"VolumeDown", KeyVolumeDown, "VolumeDown"},
		{"MediaPlay", KeyMediaPlay, "MediaPlay"},
		{"MediaStop", KeyMediaStop, "MediaStop"},
		{"MediaNext", KeyMediaNext, "MediaNext"},
		{"MediaPrev", KeyMediaPrev, "MediaPrev"},
		{"Unknown value", Key(9999), "Key(9999)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.k.String(); got != tt.want {
				t.Errorf("Key.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKey_String_AllLetters(t *testing.T) {
	letters := []struct {
		key  Key
		want string
	}{
		{KeyA, "A"}, {KeyB, "B"}, {KeyC, "C"}, {KeyD, "D"},
		{KeyE, "E"}, {KeyF, "F"}, {KeyG, "G"}, {KeyH, "H"},
		{KeyI, "I"}, {KeyJ, "J"}, {KeyK, "K"}, {KeyL, "L"},
		{KeyM, "M"}, {KeyN, "N"}, {KeyO, "O"}, {KeyP, "P"},
		{KeyQ, "Q"}, {KeyR, "R"}, {KeyS, "S"}, {KeyT, "T"},
		{KeyU, "U"}, {KeyV, "V"}, {KeyW, "W"}, {KeyX, "X"},
		{KeyY, "Y"}, {KeyZ, "Z"},
	}
	for _, tt := range letters {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.key.String(); got != tt.want {
				t.Errorf("Key.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKey_String_AllDigits(t *testing.T) {
	digits := []struct {
		key  Key
		want string
	}{
		{Key0, "0"}, {Key1, "1"}, {Key2, "2"}, {Key3, "3"},
		{Key4, "4"}, {Key5, "5"}, {Key6, "6"}, {Key7, "7"},
		{Key8, "8"}, {Key9, "9"},
	}
	for _, tt := range digits {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.key.String(); got != tt.want {
				t.Errorf("Key.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKey_String_AllFunctionKeys(t *testing.T) {
	fkeys := []struct {
		key  Key
		want string
	}{
		{KeyF1, "F1"}, {KeyF2, "F2"}, {KeyF3, "F3"}, {KeyF4, "F4"},
		{KeyF5, "F5"}, {KeyF6, "F6"}, {KeyF7, "F7"}, {KeyF8, "F8"},
		{KeyF9, "F9"}, {KeyF10, "F10"}, {KeyF11, "F11"}, {KeyF12, "F12"},
		{KeyF13, "F13"}, {KeyF14, "F14"}, {KeyF15, "F15"}, {KeyF16, "F16"},
		{KeyF17, "F17"}, {KeyF18, "F18"}, {KeyF19, "F19"}, {KeyF20, "F20"},
		{KeyF21, "F21"}, {KeyF22, "F22"}, {KeyF23, "F23"}, {KeyF24, "F24"},
	}
	for _, tt := range fkeys {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.key.String(); got != tt.want {
				t.Errorf("Key.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKey_String_Navigation(t *testing.T) {
	navKeys := []struct {
		key  Key
		want string
	}{
		{KeyHome, "Home"}, {KeyEnd, "End"},
		{KeyPageUp, "PageUp"}, {KeyPageDown, "PageDown"},
		{KeyInsert, "Insert"},
	}
	for _, tt := range navKeys {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.key.String(); got != tt.want {
				t.Errorf("Key.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKey_String_AllNumpad(t *testing.T) {
	numpadKeys := []struct {
		key  Key
		want string
	}{
		{KeyNumpad0, "Numpad0"}, {KeyNumpad1, "Numpad1"}, {KeyNumpad2, "Numpad2"},
		{KeyNumpad3, "Numpad3"}, {KeyNumpad4, "Numpad4"}, {KeyNumpad5, "Numpad5"},
		{KeyNumpad6, "Numpad6"}, {KeyNumpad7, "Numpad7"}, {KeyNumpad8, "Numpad8"},
		{KeyNumpad9, "Numpad9"},
	}
	for _, tt := range numpadKeys {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.key.String(); got != tt.want {
				t.Errorf("Key.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKey_IsLetter(t *testing.T) {
	tests := []struct {
		name string
		k    Key
		want bool
	}{
		{"A", KeyA, true},
		{"Z", KeyZ, true},
		{"M (middle)", KeyM, true},
		{"0", Key0, false},
		{"F1", KeyF1, false},
		{"Enter", KeyEnter, false},
		{"Unknown", KeyUnknown, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.k.IsLetter(); got != tt.want {
				t.Errorf("IsLetter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKey_IsDigit(t *testing.T) {
	tests := []struct {
		name string
		k    Key
		want bool
	}{
		{"0", Key0, true},
		{"9", Key9, true},
		{"5 (middle)", Key5, true},
		{"A", KeyA, false},
		{"Numpad0", KeyNumpad0, false},
		{"F1", KeyF1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.k.IsDigit(); got != tt.want {
				t.Errorf("IsDigit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKey_IsFunction(t *testing.T) {
	tests := []struct {
		name string
		k    Key
		want bool
	}{
		{"F1", KeyF1, true},
		{"F12", KeyF12, true},
		{"F24", KeyF24, true},
		{"A", KeyA, false},
		{"Enter", KeyEnter, false},
		{"Unknown", KeyUnknown, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.k.IsFunction(); got != tt.want {
				t.Errorf("IsFunction() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKey_IsNavigation(t *testing.T) {
	tests := []struct {
		name string
		k    Key
		want bool
	}{
		{"Up", KeyUp, true},
		{"Down", KeyDown, true},
		{"Left", KeyLeft, true},
		{"Right", KeyRight, true},
		{"Home", KeyHome, true},
		{"End", KeyEnd, true},
		{"PageUp", KeyPageUp, true},
		{"PageDown", KeyPageDown, true},
		{"A", KeyA, false},
		{"Enter", KeyEnter, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.k.IsNavigation(); got != tt.want {
				t.Errorf("IsNavigation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKey_IsModifier(t *testing.T) {
	tests := []struct {
		name string
		k    Key
		want bool
	}{
		{"LeftShift", KeyLeftShift, true},
		{"RightShift", KeyRightShift, true},
		{"LeftCtrl", KeyLeftCtrl, true},
		{"RightCtrl", KeyRightCtrl, true},
		{"LeftAlt", KeyLeftAlt, true},
		{"RightAlt", KeyRightAlt, true},
		{"LeftSuper", KeyLeftSuper, true},
		{"RightSuper", KeyRightSuper, true},
		{"CapsLock", KeyCapsLock, true},
		{"NumLock", KeyNumLock, true},
		{"ScrollLock", KeyScrollLock, true},
		{"A", KeyA, false},
		{"Enter", KeyEnter, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.k.IsModifier(); got != tt.want {
				t.Errorf("IsModifier() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKey_IsNumpad(t *testing.T) {
	tests := []struct {
		name string
		k    Key
		want bool
	}{
		{"Numpad0", KeyNumpad0, true},
		{"Numpad9", KeyNumpad9, true},
		{"NumpadDecimal", KeyNumpadDecimal, true},
		{"NumpadEnter", KeyNumpadEnter, true},
		{"NumpadAdd", KeyNumpadAdd, true},
		{"NumpadSubtract", KeyNumpadSubtract, true},
		{"NumpadMultiply", KeyNumpadMultiply, true},
		{"NumpadDivide", KeyNumpadDivide, true},
		{"0", Key0, false},
		{"Enter", KeyEnter, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.k.IsNumpad(); got != tt.want {
				t.Errorf("IsNumpad() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewKeyEvent(t *testing.T) {
	tests := []struct {
		name    string
		keyType KeyEventType
		key     Key
		r       rune
		mods    Modifiers
	}{
		{"Press A", KeyPress, KeyA, 'a', ModNone},
		{"Press Shift+A", KeyPress, KeyA, 'A', ModShift},
		{"Release Enter", KeyRelease, KeyEnter, 0, ModNone},
		{"Repeat F1 with Ctrl", KeyRepeat, KeyF1, 0, ModCtrl},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			e := NewKeyEvent(tt.keyType, tt.key, tt.r, tt.mods)
			after := time.Now()

			if e.Type() != TypeKey {
				t.Errorf("Type() = %v, want %v", e.Type(), TypeKey)
			}
			if e.KeyType != tt.keyType {
				t.Errorf("KeyType = %v, want %v", e.KeyType, tt.keyType)
			}
			if e.Key != tt.key {
				t.Errorf("Key = %v, want %v", e.Key, tt.key)
			}
			if e.Rune != tt.r {
				t.Errorf("Rune = %v, want %v", e.Rune, tt.r)
			}
			if e.Modifiers() != tt.mods {
				t.Errorf("Modifiers() = %v, want %v", e.Modifiers(), tt.mods)
			}
			if e.Time().Before(before) || e.Time().After(after) {
				t.Errorf("Time() = %v, want between %v and %v", e.Time(), before, after)
			}
		})
	}
}

func TestNewKeyEventWithTime(t *testing.T) {
	fixedTime := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

	e := NewKeyEventWithTime(KeyPress, KeyA, 'a', ModShift, fixedTime)

	if !e.Time().Equal(fixedTime) {
		t.Errorf("Time() = %v, want %v", e.Time(), fixedTime)
	}
	if e.Type() != TypeKey {
		t.Errorf("Type() = %v, want %v", e.Type(), TypeKey)
	}
}

func TestKeyEvent_IsPress(t *testing.T) {
	tests := []struct {
		name    string
		keyType KeyEventType
		want    bool
	}{
		{"Press", KeyPress, true},
		{"Release", KeyRelease, false},
		{"Repeat", KeyRepeat, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewKeyEvent(tt.keyType, KeyA, 'a', ModNone)
			if got := e.IsPress(); got != tt.want {
				t.Errorf("IsPress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeyEvent_IsRelease(t *testing.T) {
	tests := []struct {
		name    string
		keyType KeyEventType
		want    bool
	}{
		{"Release", KeyRelease, true},
		{"Press", KeyPress, false},
		{"Repeat", KeyRepeat, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewKeyEvent(tt.keyType, KeyA, 'a', ModNone)
			if got := e.IsRelease(); got != tt.want {
				t.Errorf("IsRelease() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeyEvent_IsRepeat(t *testing.T) {
	tests := []struct {
		name    string
		keyType KeyEventType
		want    bool
	}{
		{"Repeat", KeyRepeat, true},
		{"Press", KeyPress, false},
		{"Release", KeyRelease, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewKeyEvent(tt.keyType, KeyA, 'a', ModNone)
			if got := e.IsRepeat(); got != tt.want {
				t.Errorf("IsRepeat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeyEvent_HasRune(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want bool
	}{
		{"Has rune 'a'", 'a', true},
		{"Has rune 'A'", 'A', true},
		{"Has rune space", ' ', true},
		{"No rune (zero)", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewKeyEvent(KeyPress, KeyA, tt.r, ModNone)
			if got := e.HasRune(); got != tt.want {
				t.Errorf("HasRune() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeyEvent_IsShift(t *testing.T) {
	tests := []struct {
		name string
		mods Modifiers
		want bool
	}{
		{"No modifiers", ModNone, false},
		{"Shift only", ModShift, true},
		{"Ctrl only", ModCtrl, false},
		{"Shift+Ctrl", ModShift | ModCtrl, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewKeyEvent(KeyPress, KeyA, 'a', tt.mods)
			if got := e.IsShift(); got != tt.want {
				t.Errorf("IsShift() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeyEvent_IsCtrl(t *testing.T) {
	tests := []struct {
		name string
		mods Modifiers
		want bool
	}{
		{"No modifiers", ModNone, false},
		{"Ctrl only", ModCtrl, true},
		{"Shift only", ModShift, false},
		{"Ctrl+Shift", ModCtrl | ModShift, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewKeyEvent(KeyPress, KeyA, 'a', tt.mods)
			if got := e.IsCtrl(); got != tt.want {
				t.Errorf("IsCtrl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeyEvent_IsAlt(t *testing.T) {
	tests := []struct {
		name string
		mods Modifiers
		want bool
	}{
		{"No modifiers", ModNone, false},
		{"Alt only", ModAlt, true},
		{"Shift only", ModShift, false},
		{"Alt+Ctrl", ModAlt | ModCtrl, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewKeyEvent(KeyPress, KeyA, 'a', tt.mods)
			if got := e.IsAlt(); got != tt.want {
				t.Errorf("IsAlt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeyEvent_IsSuper(t *testing.T) {
	tests := []struct {
		name string
		mods Modifiers
		want bool
	}{
		{"No modifiers", ModNone, false},
		{"Super only", ModSuper, true},
		{"Shift only", ModShift, false},
		{"Super+Ctrl", ModSuper | ModCtrl, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewKeyEvent(KeyPress, KeyA, 'a', tt.mods)
			if got := e.IsSuper(); got != tt.want {
				t.Errorf("IsSuper() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeyEvent_String(t *testing.T) {
	tests := []struct {
		name     string
		keyType  KeyEventType
		key      Key
		r        rune
		mods     Modifiers
		contains []string
	}{
		{
			name:     "Press A with rune",
			keyType:  KeyPress,
			key:      KeyA,
			r:        'a',
			mods:     ModNone,
			contains: []string{"KeyEvent", "Press", "A", "'a'"},
		},
		{
			name:     "Press F1 without rune",
			keyType:  KeyPress,
			key:      KeyF1,
			r:        0,
			mods:     ModCtrl,
			contains: []string{"KeyEvent", "Press", "F1", "Ctrl"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewKeyEvent(tt.keyType, tt.key, tt.r, tt.mods)
			s := e.String()
			for _, c := range tt.contains {
				if !strings.Contains(s, c) {
					t.Errorf("String() = %v, should contain %v", s, c)
				}
			}
		})
	}
}

func TestKeyEvent_ImplementsEvent(t *testing.T) {
	var e Event = NewKeyEvent(KeyPress, KeyA, 'a', ModNone)

	if e.Type() != TypeKey {
		t.Errorf("Type() = %v, want %v", e.Type(), TypeKey)
	}
	if e.Handled() {
		t.Error("Handled() should be false initially")
	}
	e.SetHandled()
	if !e.Handled() {
		t.Error("Handled() should be true after SetHandled()")
	}
}

// BenchmarkNewKeyEvent benchmarks creating a new KeyEvent.
func BenchmarkNewKeyEvent(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = NewKeyEvent(KeyPress, KeyA, 'a', ModCtrl|ModShift)
	}
}

// BenchmarkKeyEvent_IsPress benchmarks checking press type.
func BenchmarkKeyEvent_IsPress(b *testing.B) {
	e := NewKeyEvent(KeyPress, KeyA, 'a', ModNone)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.IsPress()
	}
}

// BenchmarkKeyEvent_HasRune benchmarks checking rune presence.
func BenchmarkKeyEvent_HasRune(b *testing.B) {
	e := NewKeyEvent(KeyPress, KeyA, 'a', ModNone)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.HasRune()
	}
}

// BenchmarkKey_IsLetter benchmarks checking if key is letter.
func BenchmarkKey_IsLetter(b *testing.B) {
	k := KeyA
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = k.IsLetter()
	}
}

// BenchmarkKey_String benchmarks getting key string.
func BenchmarkKey_String(b *testing.B) {
	k := KeyA
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = k.String()
	}
}

// BenchmarkKeyEvent_String benchmarks string conversion.
func BenchmarkKeyEvent_String(b *testing.B) {
	e := NewKeyEvent(KeyPress, KeyA, 'a', ModCtrl|ModShift)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = e.String()
	}
}
