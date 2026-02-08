package focus

import "github.com/gogpu/ui/event"

// Shortcut defines a keyboard shortcut as a combination of a key and modifier flags.
//
// Shortcuts are matched against key press events. A shortcut matches when
// the key matches and all specified modifiers are held.
type Shortcut struct {
	// Key is the key code that triggers this shortcut.
	Key event.Key

	// Ctrl indicates whether the Control modifier must be held.
	Ctrl bool

	// Shift indicates whether the Shift modifier must be held.
	Shift bool

	// Alt indicates whether the Alt modifier must be held.
	Alt bool
}

// Matches reports whether the shortcut matches the given key event.
func (s Shortcut) Matches(e *event.KeyEvent) bool {
	if e.Key != s.Key {
		return false
	}

	mods := e.Modifiers()

	if s.Ctrl != mods.IsCtrl() {
		return false
	}
	if s.Shift != mods.IsShift() {
		return false
	}
	if s.Alt != mods.IsAlt() {
		return false
	}

	return true
}

// shortcutEntry pairs a shortcut with its handler.
type shortcutEntry struct {
	shortcut Shortcut
	handler  func()
}
