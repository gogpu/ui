package focus

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/widget"
)

// Manager tracks keyboard focus within a widget tree.
//
// It maintains a reference to the widget tree root and the currently
// focused widget. Tab and Shift+Tab key events are handled automatically
// to cycle focus through focusable widgets in depth-first order.
//
// The Manager also supports global keyboard shortcuts that take precedence
// over tab navigation.
type Manager struct {
	root      widget.Widget
	focused   widget.Focusable
	shortcuts []shortcutEntry
}

// New creates a new focus Manager for the given widget tree root.
//
// The root widget is the top-level widget whose subtree will be
// traversed for focus management. It may be nil, in which case the
// manager operates as a no-op until SetRoot is called.
func New(root widget.Widget) *Manager {
	return &Manager{
		root: root,
	}
}

// SetRoot replaces the widget tree root.
//
// If the currently focused widget is no longer in the new tree,
// focus is cleared.
func (m *Manager) SetRoot(root widget.Widget) {
	m.root = root

	// If there is a focused widget, verify it is still reachable.
	if m.focused != nil {
		focusable := collectFocusable(m.root)
		if indexOf(focusable, m.focused) < 0 {
			m.focused.SetFocused(false)
			m.focused = nil
		}
	}
}

// Focus sets keyboard focus to the given widget.
//
// If another widget currently has focus, it is blurred first.
// If w is nil, focus is cleared (equivalent to [Manager.Blur]).
// If w is not focusable (IsFocusable returns false), this is a no-op.
func (m *Manager) Focus(w widget.Focusable) {
	if w != nil && !w.IsFocusable() {
		return
	}

	if m.focused == w {
		return
	}

	// Blur current.
	if m.focused != nil {
		m.focused.SetFocused(false)
	}

	m.focused = w

	if m.focused != nil {
		m.focused.SetFocused(true)
	}
}

// Blur removes focus from the currently focused widget.
//
// After calling Blur, [Manager.Focused] returns nil.
func (m *Manager) Blur() {
	if m.focused != nil {
		m.focused.SetFocused(false)
		m.focused = nil
	}
}

// Focused returns the currently focused widget, or nil if no widget has focus.
func (m *Manager) Focused() widget.Focusable {
	return m.focused
}

// Next moves focus to the next focusable widget in tab order.
//
// Tab order is determined by depth-first traversal of the widget tree.
// If no widget currently has focus, focus moves to the first focusable widget.
// If the last focusable widget has focus, focus wraps to the first.
// If no focusable widgets exist, this is a no-op.
func (m *Manager) Next() {
	focusable := collectFocusable(m.root)
	if len(focusable) == 0 {
		return
	}

	if m.focused == nil {
		m.Focus(focusable[0])
		return
	}

	idx := indexOf(focusable, m.focused)
	if idx < 0 {
		m.Focus(focusable[0])
		return
	}

	next := (idx + 1) % len(focusable)
	m.Focus(focusable[next])
}

// Previous moves focus to the previous focusable widget in tab order.
//
// If no widget currently has focus, focus moves to the last focusable widget.
// If the first focusable widget has focus, focus wraps to the last.
// If no focusable widgets exist, this is a no-op.
func (m *Manager) Previous() {
	focusable := collectFocusable(m.root)
	if len(focusable) == 0 {
		return
	}

	if m.focused == nil {
		m.Focus(focusable[len(focusable)-1])
		return
	}

	idx := indexOf(focusable, m.focused)
	if idx < 0 {
		m.Focus(focusable[len(focusable)-1])
		return
	}

	prev := (idx - 1 + len(focusable)) % len(focusable)
	m.Focus(focusable[prev])
}

// HandleKeyEvent processes a key event for focus management.
//
// The event is checked in the following order:
//  1. Registered keyboard shortcuts
//  2. Tab key (moves focus to next widget)
//  3. Shift+Tab (moves focus to previous widget)
//
// Returns true if the event was consumed (shortcut matched or tab navigation
// occurred), false otherwise.
func (m *Manager) HandleKeyEvent(e *event.KeyEvent) bool {
	if e == nil {
		return false
	}

	// Check shortcuts on key press only.
	if e.KeyType == event.KeyPress && m.matchShortcut(e) {
		return true
	}

	// Tab navigation on press and repeat.
	if e.Key == event.KeyTab {
		return m.handleTab(e)
	}

	return false
}

// matchShortcut checks registered shortcuts and runs the first matching handler.
func (m *Manager) matchShortcut(e *event.KeyEvent) bool {
	for _, entry := range m.shortcuts {
		if entry.shortcut.Matches(e) {
			entry.handler()
			return true
		}
	}
	return false
}

// handleTab processes Tab key events for focus navigation.
func (m *Manager) handleTab(e *event.KeyEvent) bool {
	switch e.KeyType {
	case event.KeyPress, event.KeyRepeat:
		if e.Modifiers().IsShift() {
			m.Previous()
		} else {
			m.Next()
		}
		return true
	case event.KeyRelease:
		return true
	default:
		return false
	}
}

// RegisterShortcut registers a global keyboard shortcut.
//
// When the shortcut's key combination is pressed, the handler function
// is called. Shortcuts take precedence over tab navigation.
func (m *Manager) RegisterShortcut(s Shortcut, handler func()) {
	if handler == nil {
		return
	}
	m.shortcuts = append(m.shortcuts, shortcutEntry{
		shortcut: s,
		handler:  handler,
	})
}

// UnregisterShortcut removes all handlers for the given shortcut.
func (m *Manager) UnregisterShortcut(s Shortcut) {
	filtered := m.shortcuts[:0]
	for _, entry := range m.shortcuts {
		if entry.shortcut != s {
			filtered = append(filtered, entry)
		}
	}
	for i := len(filtered); i < len(m.shortcuts); i++ {
		m.shortcuts[i] = shortcutEntry{}
	}
	m.shortcuts = filtered
}
