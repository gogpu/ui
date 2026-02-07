package a11y

// CheckedState represents the checked state of a checkbox or similar control.
type CheckedState uint8

// Checked state constants.
const (
	// CheckedFalse indicates the control is not checked.
	CheckedFalse CheckedState = iota

	// CheckedTrue indicates the control is checked.
	CheckedTrue

	// CheckedMixed indicates the control is in a mixed/indeterminate state.
	// This is used when a checkbox represents a group where some items
	// are checked and some are not.
	CheckedMixed
)

// String returns a human-readable name for the checked state.
func (c CheckedState) String() string {
	switch c {
	case CheckedFalse:
		return "Unchecked"
	case CheckedTrue:
		return "Checked"
	case CheckedMixed:
		return "Mixed"
	default:
		return unknownStr
	}
}

// State holds the dynamic accessibility state of a UI element.
//
// State is a value type that captures a snapshot of an element's current
// accessibility-relevant properties. It is returned by [Accessible.AccessibilityState]
// and stored in [Node] instances for the accessibility tree.
//
// All fields have meaningful zero values: the zero State represents an enabled,
// unchecked, visible, non-expandable element with no numeric value.
//
// # Expandable Elements
//
// The Expanded field uses a *bool to distinguish between three cases:
//   - nil: the element is not expandable (e.g., a button)
//   - *false: the element is expandable but currently collapsed (e.g., a closed tree node)
//   - *true: the element is expandable and currently expanded
//
// # Numeric Values
//
// ValueMin, ValueMax, and ValueNow use *float64 to indicate whether numeric
// value semantics apply. When all are nil, the element has no numeric value
// (e.g., a button). When set, they describe a range (e.g., a slider from 0 to 100).
type State struct {
	// Disabled indicates the element cannot be interacted with.
	// Disabled elements are still visible and present in the accessibility tree.
	Disabled bool

	// Selected indicates the element is currently selected within a group
	// (e.g., a selected list item or tab).
	Selected bool

	// Checked indicates the checked state for checkboxes, radio buttons,
	// and similar controls.
	Checked CheckedState

	// Expanded indicates whether an expandable element is expanded or collapsed.
	// nil means the element is not expandable.
	Expanded *bool

	// ReadOnly indicates the element's value cannot be modified by the user,
	// but the content can still be read and focused.
	ReadOnly bool

	// Required indicates the element must have a value before a form can
	// be submitted.
	Required bool

	// Busy indicates the element is being modified and assistive technology
	// should wait before exposing changes to the user.
	Busy bool

	// Hidden indicates the element should be excluded from the accessibility tree.
	// Hidden elements are not visible to assistive technology.
	Hidden bool

	// Focused indicates the element currently has keyboard focus.
	Focused bool

	// Modal indicates the element is a modal dialog that restricts interaction
	// to its descendants.
	Modal bool

	// Multiselectable indicates the element allows selecting multiple items.
	Multiselectable bool

	// ValueMin is the minimum allowed numeric value, or nil if not applicable.
	ValueMin *float64

	// ValueMax is the maximum allowed numeric value, or nil if not applicable.
	ValueMax *float64

	// ValueNow is the current numeric value, or nil if not applicable.
	ValueNow *float64

	// ValueText is a human-readable representation of the current value.
	// For a slider, this might be "50%" or "Medium". When empty, assistive
	// technology will typically use the numeric value.
	ValueText string

	// Level indicates the heading level (1-6) or tree item depth.
	// Zero means no level is applicable.
	Level int
}

// BoolPtr returns a pointer to the given boolean value.
//
// This is a convenience function for setting the [State.Expanded] field:
//
//	state := a11y.State{
//	    Expanded: a11y.BoolPtr(true),  // expanded
//	}
func BoolPtr(v bool) *bool {
	return &v
}

// Float64Ptr returns a pointer to the given float64 value.
//
// This is a convenience function for setting numeric value fields:
//
//	state := a11y.State{
//	    ValueMin: a11y.Float64Ptr(0),
//	    ValueMax: a11y.Float64Ptr(100),
//	    ValueNow: a11y.Float64Ptr(50),
//	}
func Float64Ptr(v float64) *float64 {
	return &v
}

// HasNumericValue returns true if the state has numeric value semantics.
//
// This is true when at least one of ValueMin, ValueMax, or ValueNow is set.
func (s State) HasNumericValue() bool {
	return s.ValueMin != nil || s.ValueMax != nil || s.ValueNow != nil
}

// IsExpandable returns true if the state represents an expandable element.
//
// This is true when Expanded is not nil.
func (s State) IsExpandable() bool {
	return s.Expanded != nil
}

// IsExpanded returns true if the state represents an expanded element.
//
// Returns false if the element is not expandable or is collapsed.
func (s State) IsExpanded() bool {
	return s.Expanded != nil && *s.Expanded
}
