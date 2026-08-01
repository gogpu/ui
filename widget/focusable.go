package widget

import "github.com/gogpu/ui/event"

// Focusable is implemented by widgets that can receive keyboard focus.
//
// Widgets that support keyboard interaction (text inputs, buttons, etc.)
// should implement this interface in addition to the Widget interface.
// The focus manager uses this interface to determine which widgets
// participate in tab navigation.
//
// WidgetBase already implements SetFocused and IsFocused, so concrete
// widgets only need to implement IsFocusable to opt into focus management.
//
// Example:
//
//	type TextInput struct {
//	    widget.WidgetBase
//	}
//
//	func (t *TextInput) IsFocusable() bool {
//	    return t.IsEnabled() && t.IsVisible()
//	}
type Focusable interface {
	// IsFocusable reports whether this widget can currently receive focus.
	//
	// A widget may return false if it is disabled, invisible, or otherwise
	// unable to accept keyboard input at this time.
	IsFocusable() bool

	// SetFocused sets the widget's focus state.
	//
	// The focus manager calls this when focus is granted or revoked.
	// Widgets should update their visual appearance accordingly.
	SetFocused(focused bool)

	// IsFocused reports whether this widget currently has keyboard focus.
	IsFocused() bool
}

// KeyGrabber is implemented by focused widgets that need a key the framework
// otherwise reserves for itself. Today that is Tab and Shift+Tab, which the
// focus manager consumes for traversal before the widget tree sees them.
//
// A terminal emulator is the canonical case: Tab is shell completion, and a
// terminal that eats it is broken. Code editors (indent) and game or 3D views
// (raw input) have the same need. Other toolkits expose the same escape
// hatch — Qt via focusNextPrevChild, the web via preventDefault on Tab.
//
// Only the FOCUSED widget is asked, and only for keys the manager was about to
// consume, so implementing this cannot break traversal elsewhere.
//
// Example:
//
//	func (t *Terminal) GrabsKey(k event.Key) bool {
//	    return k == event.KeyTab // forward to the shell, don't traverse
//	}
type KeyGrabber interface {
	// GrabsKey reports whether this widget wants key k delivered to it
	// instead of being handled by the framework.
	GrabsKey(k event.Key) bool
}
