package a11y

// Accessible is the interface that widgets implement to expose accessibility semantics.
//
// Each method returns a snapshot of the widget's current accessibility state.
// Platform adapters query these methods to build and update the native accessibility
// tree consumed by screen readers and other assistive technology.
//
// All methods must be safe to call from the UI thread. Implementations should
// not block or perform expensive operations.
//
// # Minimal Implementation
//
// At minimum, a widget should provide a meaningful role and label:
//
//	func (b *Button) AccessibilityRole() a11y.Role  { return a11y.RoleButton }
//	func (b *Button) AccessibilityLabel() string     { return b.text }
//	func (b *Button) AccessibilityHint() string      { return "" }
//	func (b *Button) AccessibilityValue() string     { return "" }
//	func (b *Button) AccessibilityState() a11y.State { return a11y.State{} }
//	func (b *Button) AccessibilityActions() []a11y.Action {
//	    return []a11y.Action{a11y.ActionClick}
//	}
//
// # Rich Implementation
//
// A slider widget provides full numeric value information:
//
//	func (s *Slider) AccessibilityRole() a11y.Role  { return a11y.RoleSlider }
//	func (s *Slider) AccessibilityLabel() string     { return s.label }
//	func (s *Slider) AccessibilityHint() string      { return "Adjusts the value" }
//	func (s *Slider) AccessibilityValue() string     { return fmt.Sprintf("%.0f%%", s.value*100) }
//	func (s *Slider) AccessibilityState() a11y.State {
//	    return a11y.State{
//	        ValueMin: a11y.Float64Ptr(float64(s.min)),
//	        ValueMax: a11y.Float64Ptr(float64(s.max)),
//	        ValueNow: a11y.Float64Ptr(float64(s.value)),
//	    }
//	}
//	func (s *Slider) AccessibilityActions() []a11y.Action {
//	    return []a11y.Action{a11y.ActionIncrement, a11y.ActionDecrement, a11y.ActionSetValue}
//	}
type Accessible interface {
	// AccessibilityRole returns the semantic role of this element.
	//
	// The role determines how assistive technology presents the element
	// and what interactions are available. This should be stable for the
	// lifetime of the widget.
	AccessibilityRole() Role

	// AccessibilityLabel returns a human-readable label for this element.
	//
	// The label is the primary text that screen readers announce.
	// It should be concise and descriptive. For example, "Save" for a save button
	// or "Volume" for a volume slider.
	//
	// An empty string indicates the element has no explicit label.
	// Assistive technology may fall back to other sources (children, value, etc.).
	AccessibilityLabel() string

	// AccessibilityHint returns a description of the result of performing
	// the default action on this element.
	//
	// The hint provides additional context about what will happen when the user
	// activates the element. For example, "Opens the settings dialog" for a
	// settings button.
	//
	// An empty string indicates no hint is available.
	AccessibilityHint() string

	// AccessibilityValue returns the current value of this element as a string.
	//
	// This is used for elements that have a user-facing value, such as text fields
	// (the entered text), sliders (e.g., "50%"), and progress bars (e.g., "75% complete").
	//
	// An empty string indicates the element has no value to report.
	AccessibilityValue() string

	// AccessibilityState returns the current dynamic accessibility state.
	//
	// The state includes properties like disabled, selected, checked, expanded,
	// and numeric value range. These change as the user interacts with the element.
	AccessibilityState() State

	// AccessibilityActions returns the list of actions that assistive technology
	// can perform on this element.
	//
	// The returned slice should not be modified by the caller.
	// An empty or nil slice indicates the element supports no special actions.
	AccessibilityActions() []Action
}
