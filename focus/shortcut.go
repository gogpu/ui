package focus

import ifocus "github.com/gogpu/ui/internal/focus"

// Shortcut defines a keyboard shortcut as a combination of a key and modifier flags.
//
// Shortcuts are matched against key press events. A shortcut matches when
// the key matches and all specified modifiers are held.
//
// Example:
//
//	save := focus.Shortcut{Key: event.KeyS, Ctrl: true}
//	redo := focus.Shortcut{Key: event.KeyZ, Ctrl: true, Shift: true}
type Shortcut = ifocus.Shortcut
