package a11y

// Action represents an accessibility action that can be performed on a UI element.
//
// Actions are the operations that assistive technology can invoke on behalf
// of the user. For example, a screen reader might trigger [ActionClick] when
// the user activates a button, or [ActionSetValue] when the user changes
// a slider value.
//
// Widgets report their supported actions through [Accessible.AccessibilityActions].
// Platform adapters translate these into platform-specific accessibility actions.
type Action uint8

// Action constants define standard accessibility actions.
//
// The set of actions follows the AccessKit and WAI-ARIA action model.
const (
	// ActionClick performs the default action for the element.
	// For buttons this triggers activation; for links this navigates.
	ActionClick Action = iota + 1

	// ActionFocus moves keyboard focus to the element.
	ActionFocus

	// ActionBlur removes keyboard focus from the element.
	ActionBlur

	// ActionSetValue sets the element's value.
	// This is used for text fields, sliders, and other value-bearing controls.
	ActionSetValue

	// ActionIncrement increases the element's value by one step.
	// This is used for sliders, spin buttons, and similar controls.
	ActionIncrement

	// ActionDecrement decreases the element's value by one step.
	// This is used for sliders, spin buttons, and similar controls.
	ActionDecrement

	// ActionExpand expands a collapsible element.
	// This is used for tree items, disclosure triangles, and similar controls.
	ActionExpand

	// ActionCollapse collapses an expanded element.
	// This is used for tree items, disclosure triangles, and similar controls.
	ActionCollapse

	// ActionSelect selects the element within its container.
	// This is used for list items, tabs, and similar selectable elements.
	ActionSelect

	// ActionScrollIntoView scrolls the element into the visible area
	// of its scroll container.
	ActionScrollIntoView

	// ActionScrollUp scrolls the content upward.
	ActionScrollUp

	// ActionScrollDown scrolls the content downward.
	ActionScrollDown

	// ActionScrollLeft scrolls the content to the left.
	ActionScrollLeft

	// ActionScrollRight scrolls the content to the right.
	ActionScrollRight

	// ActionShowContextMenu opens the element's context menu.
	ActionShowContextMenu

	// ActionDismiss dismisses a transient element such as a tooltip,
	// dialog, or popup menu.
	ActionDismiss
)

// String returns a human-readable name for the action.
func (a Action) String() string {
	switch a {
	case ActionClick:
		return "Click"
	case ActionFocus:
		return "Focus"
	case ActionBlur:
		return "Blur"
	case ActionSetValue:
		return "SetValue"
	case ActionIncrement:
		return "Increment"
	case ActionDecrement:
		return "Decrement"
	case ActionExpand:
		return "Expand"
	case ActionCollapse:
		return "Collapse"
	case ActionSelect:
		return "Select"
	case ActionScrollIntoView:
		return "ScrollIntoView"
	case ActionScrollUp:
		return "ScrollUp"
	case ActionScrollDown:
		return "ScrollDown"
	case ActionScrollLeft:
		return "ScrollLeft"
	case ActionScrollRight:
		return "ScrollRight"
	case ActionShowContextMenu:
		return "ShowContextMenu"
	case ActionDismiss:
		return "Dismiss"
	default:
		return unknownStr
	}
}
