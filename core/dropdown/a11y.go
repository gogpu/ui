package dropdown

// A11yRole returns the ARIA role for the dropdown widget.
// Returns "combobox" per WAI-ARIA 1.2 spec for dropdown/select controls.
func (w *Widget) A11yRole() string {
	return "combobox"
}

// A11yLabel returns the accessible label for the dropdown.
// It uses the configured accessibility hint if set, or falls back to
// "dropdown" as a generic label.
func (w *Widget) A11yLabel() string {
	if w.cfg.a11yHint != "" {
		return w.cfg.a11yHint
	}
	return "dropdown"
}

// A11yValue returns the currently displayed value for assistive technology.
// If no item is selected, it returns the placeholder text.
func (w *Widget) A11yValue() string {
	if w.selectedIndex >= 0 && w.selectedIndex < len(w.cfg.items) {
		return w.cfg.items[w.selectedIndex].DisplayText()
	}
	if w.cfg.placeholder != "" {
		return w.cfg.placeholder
	}
	return ""
}

// A11yExpanded returns true if the dropdown menu is currently visible.
// Maps to the aria-expanded attribute.
func (w *Widget) A11yExpanded() bool {
	return w.open
}
