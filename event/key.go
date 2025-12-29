package event

import (
	"fmt"
	"time"
)

// KeyEventType represents the specific type of keyboard event.
type KeyEventType uint8

// Keyboard event type constants.
const (
	// KeyPress indicates a key was pressed.
	KeyPress KeyEventType = iota + 1

	// KeyRelease indicates a key was released.
	KeyRelease

	// KeyRepeat indicates a key is being held and auto-repeating.
	KeyRepeat
)

// String returns a human-readable name for the key event type.
func (t KeyEventType) String() string {
	switch t {
	case KeyPress:
		return "Press"
	case KeyRelease:
		return "Release"
	case KeyRepeat:
		return "Repeat"
	default:
		return unknownStr
	}
}

// Key represents a keyboard key code.
//
// Key codes represent physical or virtual keys, not characters.
// Use Rune for the character that was typed.
type Key uint16

// Letter key constants (A-Z).
const (
	KeyA Key = iota + 1
	KeyB
	KeyC
	KeyD
	KeyE
	KeyF
	KeyG
	KeyH
	KeyI
	KeyJ
	KeyK
	KeyL
	KeyM
	KeyN
	KeyO
	KeyP
	KeyQ
	KeyR
	KeyS
	KeyT
	KeyU
	KeyV
	KeyW
	KeyX
	KeyY
	KeyZ
)

// Number key constants (0-9).
const (
	Key0 Key = iota + 100
	Key1
	Key2
	Key3
	Key4
	Key5
	Key6
	Key7
	Key8
	Key9
)

// Function key constants (F1-F24).
const (
	KeyF1 Key = iota + 200
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
	KeyF13
	KeyF14
	KeyF15
	KeyF16
	KeyF17
	KeyF18
	KeyF19
	KeyF20
	KeyF21
	KeyF22
	KeyF23
	KeyF24
)

// Navigation key constants.
const (
	KeyUp Key = iota + 300
	KeyDown
	KeyLeft
	KeyRight
	KeyHome
	KeyEnd
	KeyPageUp
	KeyPageDown
)

// Editing key constants.
const (
	KeyEnter Key = iota + 400
	KeyTab
	KeyBackspace
	KeyDelete
	KeyInsert
	KeyEscape
	KeySpace
)

// Modifier key constants.
const (
	KeyLeftShift Key = iota + 500
	KeyRightShift
	KeyLeftCtrl
	KeyRightCtrl
	KeyLeftAlt
	KeyRightAlt
	KeyLeftSuper
	KeyRightSuper
	KeyCapsLock
	KeyNumLock
	KeyScrollLock
)

// Numpad key constants.
const (
	KeyNumpad0 Key = iota + 600
	KeyNumpad1
	KeyNumpad2
	KeyNumpad3
	KeyNumpad4
	KeyNumpad5
	KeyNumpad6
	KeyNumpad7
	KeyNumpad8
	KeyNumpad9
	KeyNumpadDecimal
	KeyNumpadEnter
	KeyNumpadAdd
	KeyNumpadSubtract
	KeyNumpadMultiply
	KeyNumpadDivide
)

// Symbol and punctuation key constants.
const (
	KeyMinus Key = iota + 700
	KeyEqual
	KeyLeftBracket
	KeyRightBracket
	KeyBackslash
	KeySemicolon
	KeyApostrophe
	KeyGrave
	KeyComma
	KeyPeriod
	KeySlash
)

// Media and system key constants.
const (
	KeyPrintScreen Key = iota + 800
	KeyPause
	KeyMenu
	KeyMute
	KeyVolumeUp
	KeyVolumeDown
	KeyMediaPlay
	KeyMediaStop
	KeyMediaNext
	KeyMediaPrev
)

// Special key constant.
const (
	// KeyUnknown represents an unknown or unmapped key.
	KeyUnknown Key = 0
)

// String returns a human-readable name for the key.
//
//nolint:gocyclo,cyclop,funlen,maintidx // Key mapping requires a large switch statement by design
func (k Key) String() string {
	switch k {
	case KeyUnknown:
		return unknownStr
	case KeyA:
		return "A"
	case KeyB:
		return "B"
	case KeyC:
		return "C"
	case KeyD:
		return "D"
	case KeyE:
		return "E"
	case KeyF:
		return "F"
	case KeyG:
		return "G"
	case KeyH:
		return "H"
	case KeyI:
		return "I"
	case KeyJ:
		return "J"
	case KeyK:
		return "K"
	case KeyL:
		return "L"
	case KeyM:
		return "M"
	case KeyN:
		return "N"
	case KeyO:
		return "O"
	case KeyP:
		return "P"
	case KeyQ:
		return "Q"
	case KeyR:
		return "R"
	case KeyS:
		return "S"
	case KeyT:
		return "T"
	case KeyU:
		return "U"
	case KeyV:
		return "V"
	case KeyW:
		return "W"
	case KeyX:
		return "X"
	case KeyY:
		return "Y"
	case KeyZ:
		return "Z"
	case Key0:
		return "0"
	case Key1:
		return "1"
	case Key2:
		return "2"
	case Key3:
		return "3"
	case Key4:
		return "4"
	case Key5:
		return "5"
	case Key6:
		return "6"
	case Key7:
		return "7"
	case Key8:
		return "8"
	case Key9:
		return "9"
	case KeyF1:
		return "F1"
	case KeyF2:
		return "F2"
	case KeyF3:
		return "F3"
	case KeyF4:
		return "F4"
	case KeyF5:
		return "F5"
	case KeyF6:
		return "F6"
	case KeyF7:
		return "F7"
	case KeyF8:
		return "F8"
	case KeyF9:
		return "F9"
	case KeyF10:
		return "F10"
	case KeyF11:
		return "F11"
	case KeyF12:
		return "F12"
	case KeyF13:
		return "F13"
	case KeyF14:
		return "F14"
	case KeyF15:
		return "F15"
	case KeyF16:
		return "F16"
	case KeyF17:
		return "F17"
	case KeyF18:
		return "F18"
	case KeyF19:
		return "F19"
	case KeyF20:
		return "F20"
	case KeyF21:
		return "F21"
	case KeyF22:
		return "F22"
	case KeyF23:
		return "F23"
	case KeyF24:
		return "F24"
	case KeyUp:
		return "Up"
	case KeyDown:
		return "Down"
	case KeyLeft:
		return "Left"
	case KeyRight:
		return "Right"
	case KeyHome:
		return "Home"
	case KeyEnd:
		return "End"
	case KeyPageUp:
		return "PageUp"
	case KeyPageDown:
		return "PageDown"
	case KeyEnter:
		return "Enter"
	case KeyTab:
		return "Tab"
	case KeyBackspace:
		return "Backspace"
	case KeyDelete:
		return "Delete"
	case KeyInsert:
		return "Insert"
	case KeyEscape:
		return "Escape"
	case KeySpace:
		return "Space"
	case KeyLeftShift:
		return "LeftShift"
	case KeyRightShift:
		return "RightShift"
	case KeyLeftCtrl:
		return "LeftCtrl"
	case KeyRightCtrl:
		return "RightCtrl"
	case KeyLeftAlt:
		return "LeftAlt"
	case KeyRightAlt:
		return "RightAlt"
	case KeyLeftSuper:
		return "LeftSuper"
	case KeyRightSuper:
		return "RightSuper"
	case KeyCapsLock:
		return "CapsLock"
	case KeyNumLock:
		return "NumLock"
	case KeyScrollLock:
		return "ScrollLock"
	case KeyNumpad0:
		return "Numpad0"
	case KeyNumpad1:
		return "Numpad1"
	case KeyNumpad2:
		return "Numpad2"
	case KeyNumpad3:
		return "Numpad3"
	case KeyNumpad4:
		return "Numpad4"
	case KeyNumpad5:
		return "Numpad5"
	case KeyNumpad6:
		return "Numpad6"
	case KeyNumpad7:
		return "Numpad7"
	case KeyNumpad8:
		return "Numpad8"
	case KeyNumpad9:
		return "Numpad9"
	case KeyNumpadDecimal:
		return "NumpadDecimal"
	case KeyNumpadEnter:
		return "NumpadEnter"
	case KeyNumpadAdd:
		return "NumpadAdd"
	case KeyNumpadSubtract:
		return "NumpadSubtract"
	case KeyNumpadMultiply:
		return "NumpadMultiply"
	case KeyNumpadDivide:
		return "NumpadDivide"
	case KeyMinus:
		return "Minus"
	case KeyEqual:
		return "Equal"
	case KeyLeftBracket:
		return "LeftBracket"
	case KeyRightBracket:
		return "RightBracket"
	case KeyBackslash:
		return "Backslash"
	case KeySemicolon:
		return "Semicolon"
	case KeyApostrophe:
		return "Apostrophe"
	case KeyGrave:
		return "Grave"
	case KeyComma:
		return "Comma"
	case KeyPeriod:
		return "Period"
	case KeySlash:
		return "Slash"
	case KeyPrintScreen:
		return "PrintScreen"
	case KeyPause:
		return "Pause"
	case KeyMenu:
		return "Menu"
	case KeyMute:
		return "Mute"
	case KeyVolumeUp:
		return "VolumeUp"
	case KeyVolumeDown:
		return "VolumeDown"
	case KeyMediaPlay:
		return "MediaPlay"
	case KeyMediaStop:
		return "MediaStop"
	case KeyMediaNext:
		return "MediaNext"
	case KeyMediaPrev:
		return "MediaPrev"
	default:
		return fmt.Sprintf("Key(%d)", k)
	}
}

// IsLetter returns true if the key is a letter key (A-Z).
func (k Key) IsLetter() bool {
	return k >= KeyA && k <= KeyZ
}

// IsDigit returns true if the key is a digit key (0-9).
func (k Key) IsDigit() bool {
	return k >= Key0 && k <= Key9
}

// IsFunction returns true if the key is a function key (F1-F24).
func (k Key) IsFunction() bool {
	return k >= KeyF1 && k <= KeyF24
}

// IsNavigation returns true if the key is a navigation key.
func (k Key) IsNavigation() bool {
	return k >= KeyUp && k <= KeyPageDown
}

// IsModifier returns true if the key is a modifier key.
func (k Key) IsModifier() bool {
	return k >= KeyLeftShift && k <= KeyScrollLock
}

// IsNumpad returns true if the key is a numpad key.
func (k Key) IsNumpad() bool {
	return k >= KeyNumpad0 && k <= KeyNumpadDivide
}

// KeyEvent represents a keyboard input event.
//
// Key represents the physical key pressed.
// Rune contains the character that was typed (if applicable).
type KeyEvent struct {
	Base

	// KeyType is the specific type of key event.
	KeyType KeyEventType

	// Key is the key code that was pressed/released.
	Key Key

	// Rune is the character that was typed, or 0 if not applicable.
	// For example, pressing 'A' with Shift produces Rune='A'.
	// Pressing 'F1' produces Rune=0.
	Rune rune

	// ScanCode is the platform-specific scan code.
	ScanCode uint32
}

// NewKeyEvent creates a new key event with the current time.
func NewKeyEvent(keyType KeyEventType, key Key, r rune, mods Modifiers) *KeyEvent {
	return &KeyEvent{
		Base:    NewBase(TypeKey, mods),
		KeyType: keyType,
		Key:     key,
		Rune:    r,
	}
}

// NewKeyEventWithTime creates a new key event with a specific timestamp.
func NewKeyEventWithTime(keyType KeyEventType, key Key, r rune, mods Modifiers, t time.Time) *KeyEvent {
	return &KeyEvent{
		Base:    NewBaseWithTime(TypeKey, t, mods),
		KeyType: keyType,
		Key:     key,
		Rune:    r,
	}
}

// IsPress returns true if this is a key press event.
func (e *KeyEvent) IsPress() bool {
	return e.KeyType == KeyPress
}

// IsRelease returns true if this is a key release event.
func (e *KeyEvent) IsRelease() bool {
	return e.KeyType == KeyRelease
}

// IsRepeat returns true if this is a key repeat event.
func (e *KeyEvent) IsRepeat() bool {
	return e.KeyType == KeyRepeat
}

// HasRune returns true if a character was typed.
func (e *KeyEvent) HasRune() bool {
	return e.Rune != 0
}

// IsShift returns true if Shift modifier is held.
func (e *KeyEvent) IsShift() bool {
	return e.Modifiers().IsShift()
}

// IsCtrl returns true if Ctrl modifier is held.
func (e *KeyEvent) IsCtrl() bool {
	return e.Modifiers().IsCtrl()
}

// IsAlt returns true if Alt modifier is held.
func (e *KeyEvent) IsAlt() bool {
	return e.Modifiers().IsAlt()
}

// IsSuper returns true if Super modifier is held.
func (e *KeyEvent) IsSuper() bool {
	return e.Modifiers().IsSuper()
}

// String returns a human-readable representation of the event.
func (e *KeyEvent) String() string {
	if e.HasRune() {
		return fmt.Sprintf("KeyEvent{Type: %s, Key: %s, Rune: %q, Mods: %s}",
			e.KeyType, e.Key, e.Rune, e.Modifiers())
	}
	return fmt.Sprintf("KeyEvent{Type: %s, Key: %s, Mods: %s}",
		e.KeyType, e.Key, e.Modifiers())
}
