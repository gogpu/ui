package event

import "strings"

// Modifiers represents a bitmask of modifier keys held during an event.
//
// Multiple modifiers can be combined using bitwise OR:
//
//	mods := ModCtrl | ModShift
type Modifiers uint8

// Modifier key constants.
const (
	// ModNone indicates no modifier keys are pressed.
	ModNone Modifiers = 0

	// ModShift indicates the Shift key is pressed.
	ModShift Modifiers = 1 << iota

	// ModCtrl indicates the Control key is pressed.
	ModCtrl

	// ModAlt indicates the Alt key (Option on macOS) is pressed.
	ModAlt

	// ModSuper indicates the Super key (Windows key, Cmd on macOS) is pressed.
	ModSuper

	// ModCapsLock indicates Caps Lock is active.
	ModCapsLock

	// ModNumLock indicates Num Lock is active.
	ModNumLock
)

// Has returns true if all the specified modifier bits are set.
//
// Example:
//
//	if mods.Has(ModCtrl | ModShift) {
//	    // Both Ctrl and Shift are pressed
//	}
func (m Modifiers) Has(mod Modifiers) bool {
	return m&mod == mod
}

// HasAny returns true if any of the specified modifier bits are set.
//
// Example:
//
//	if mods.HasAny(ModCtrl | ModSuper) {
//	    // Either Ctrl or Super is pressed
//	}
func (m Modifiers) HasAny(mod Modifiers) bool {
	return m&mod != 0
}

// IsShift returns true if the Shift modifier is set.
func (m Modifiers) IsShift() bool {
	return m.Has(ModShift)
}

// IsCtrl returns true if the Control modifier is set.
func (m Modifiers) IsCtrl() bool {
	return m.Has(ModCtrl)
}

// IsAlt returns true if the Alt modifier is set.
func (m Modifiers) IsAlt() bool {
	return m.Has(ModAlt)
}

// IsSuper returns true if the Super modifier is set.
func (m Modifiers) IsSuper() bool {
	return m.Has(ModSuper)
}

// IsCapsLock returns true if Caps Lock is active.
func (m Modifiers) IsCapsLock() bool {
	return m.Has(ModCapsLock)
}

// IsNumLock returns true if Num Lock is active.
func (m Modifiers) IsNumLock() bool {
	return m.Has(ModNumLock)
}

// With returns a new Modifiers value with the specified modifier added.
func (m Modifiers) With(mod Modifiers) Modifiers {
	return m | mod
}

// Without returns a new Modifiers value with the specified modifier removed.
func (m Modifiers) Without(mod Modifiers) Modifiers {
	return m &^ mod
}

// String returns a human-readable representation of the modifiers.
//
// Example:
//
//	(ModCtrl | ModShift).String() // "Ctrl+Shift"
func (m Modifiers) String() string {
	if m == ModNone {
		return "None"
	}

	var parts []string
	if m.Has(ModCtrl) {
		parts = append(parts, "Ctrl")
	}
	if m.Has(ModAlt) {
		parts = append(parts, "Alt")
	}
	if m.Has(ModShift) {
		parts = append(parts, "Shift")
	}
	if m.Has(ModSuper) {
		parts = append(parts, "Super")
	}
	if m.Has(ModCapsLock) {
		parts = append(parts, "CapsLock")
	}
	if m.Has(ModNumLock) {
		parts = append(parts, "NumLock")
	}

	return strings.Join(parts, "+")
}
