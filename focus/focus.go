package focus

import (
	ifocus "github.com/gogpu/ui/internal/focus"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// DefaultFocusRingOffset is the distance between the widget bounds and
// the focus ring, in logical pixels.
const DefaultFocusRingOffset = ifocus.DefaultFocusRingOffset

// DefaultFocusRingStrokeWidth is the line width of the focus ring stroke,
// in logical pixels.
const DefaultFocusRingStrokeWidth = ifocus.DefaultFocusRingStrokeWidth

// Manager tracks keyboard focus within a widget tree.
//
// It maintains a reference to the widget tree root and the currently
// focused widget. Tab and Shift+Tab key events are handled automatically
// to cycle focus through focusable widgets in depth-first order.
//
// Create a new Manager with [New].
type Manager struct {
	impl *ifocus.Manager
}

// New creates a new focus [Manager] for the given widget tree root.
//
// The root widget is the top-level widget whose subtree will be
// traversed for focus management. It may be nil, in which case the
// manager operates as a no-op until SetRoot is called.
func New(root widget.Widget) *Manager {
	return &Manager{impl: ifocus.New(root)}
}

// SetRoot replaces the widget tree root.
//
// If the currently focused widget is no longer in the new tree,
// focus is cleared.
func (m *Manager) SetRoot(root widget.Widget) {
	m.impl.SetRoot(root)
}

// Focus sets keyboard focus to the given widget.
//
// If another widget currently has focus, it is blurred first.
// If w is nil, focus is cleared (equivalent to [Manager.Blur]).
// If w is not focusable (IsFocusable returns false), this is a no-op.
func (m *Manager) Focus(w widget.Focusable) {
	m.impl.Focus(w)
}

// Blur removes focus from the currently focused widget.
//
// After calling Blur, [Manager.Focused] returns nil.
func (m *Manager) Blur() {
	m.impl.Blur()
}

// Focused returns the currently focused widget, or nil if no widget has focus.
func (m *Manager) Focused() widget.Focusable {
	return m.impl.Focused()
}

// Next moves focus to the next focusable widget in tab order.
//
// Tab order is determined by depth-first traversal of the widget tree.
// If no widget currently has focus, focus moves to the first focusable widget.
// If the last focusable widget has focus, focus wraps to the first.
// If no focusable widgets exist, this is a no-op.
func (m *Manager) Next() {
	m.impl.Next()
}

// Previous moves focus to the previous focusable widget in tab order.
//
// If no widget currently has focus, focus moves to the last focusable widget.
// If the first focusable widget has focus, focus wraps to the last.
// If no focusable widgets exist, this is a no-op.
func (m *Manager) Previous() {
	m.impl.Previous()
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
	return m.impl.HandleKeyEvent(e)
}

// RegisterShortcut registers a global keyboard shortcut.
//
// When the shortcut's key combination is pressed, the handler function
// is called. Shortcuts take precedence over tab navigation.
func (m *Manager) RegisterShortcut(s Shortcut, handler func()) {
	m.impl.RegisterShortcut(ifocus.Shortcut{
		Key:   s.Key,
		Ctrl:  s.Ctrl,
		Shift: s.Shift,
		Alt:   s.Alt,
	}, handler)
}

// UnregisterShortcut removes all handlers for the given shortcut.
func (m *Manager) UnregisterShortcut(s Shortcut) {
	m.impl.UnregisterShortcut(ifocus.Shortcut{
		Key:   s.Key,
		Ctrl:  s.Ctrl,
		Shift: s.Shift,
		Alt:   s.Alt,
	})
}

// DrawFocusRing draws a focus indicator around the given bounds.
//
// The ring is drawn as a rounded rectangle outline, offset slightly
// outside the bounds. The radius controls corner rounding; use 0 for
// square corners.
//
// Example:
//
//	if w.IsFocused() {
//	    focus.DrawFocusRing(canvas, w.Bounds(), widget.ColorBlue, 4.0)
//	}
func DrawFocusRing(canvas widget.Canvas, bounds geometry.Rect, color widget.Color, radius float32) {
	ifocus.DrawFocusRing(canvas, bounds, color, radius)
}
