package textfield

import "github.com/gogpu/ui/a11y"

// AccessibleRole returns the accessibility role for the text field.
func (w *Widget) AccessibleRole() a11y.Role {
	return a11y.RoleTextField
}

// AccessibleLabel returns the accessibility label for the text field.
// If an explicit a11y label was set via [A11yLabel], it is returned.
// Otherwise, the placeholder text is used as a fallback label.
func (w *Widget) AccessibleLabel() string {
	if w.cfg.a11yLabel != "" {
		return w.cfg.a11yLabel
	}
	return w.cfg.placeholder
}

// AccessibleValue returns the current text value for assistive technology.
// For password fields, the value is masked.
func (w *Widget) AccessibleValue() string {
	text := w.resolvedText()
	if w.cfg.inputType == TypePassword {
		return maskText(len([]rune(text)))
	}
	return text
}
