package widget

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
