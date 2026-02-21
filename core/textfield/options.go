package textfield

import "github.com/gogpu/ui/state"

// Option configures a text field during construction.
type Option func(*config)

// Placeholder sets the placeholder text shown when the field is empty and unfocused.
func Placeholder(s string) Option {
	return func(c *config) {
		c.placeholder = s
	}
}

// InitialValue sets the initial text value for the field.
func InitialValue(s string) Option {
	return func(c *config) {
		c.value = s
	}
}

// Value binds the text field to a reactive signal for two-way data binding.
// When the user types, the signal is updated. When the signal is set externally,
// the text field reflects the change.
func Value(sig state.Signal[string]) Option {
	return func(c *config) {
		c.signal = sig
	}
}

// OnChange sets the callback invoked when the text value changes.
// The callback receives the new text value.
func OnChange(fn func(string)) Option {
	return func(c *config) {
		c.onChange = fn
	}
}

// OnSubmit sets the callback invoked when the user presses Enter
// in a single-line text field. The callback receives the current text value.
func OnSubmit(fn func(string)) Option {
	return func(c *config) {
		c.onSubmit = fn
	}
}

// InputTypeOpt sets the input type (text, password, email, number, search).
func InputTypeOpt(t InputType) Option {
	return func(c *config) {
		c.inputType = t
	}
}

// MaxLength sets the maximum number of characters allowed.
// A value of 0 means no limit.
func MaxLength(n int) Option {
	return func(c *config) {
		c.maxLength = n
	}
}

// Validation sets one or more validation functions.
// Each function receives the current value and returns an error message
// (empty string means valid). Validation runs on every text change.
func Validation(fns ...ValidationFunc) Option {
	return func(c *config) {
		c.validation = append(c.validation, fns...)
	}
}

// Disabled sets the text field's disabled state. A disabled text field
// does not respond to user input and is drawn with a dimmed appearance.
func Disabled(d bool) Option {
	return func(c *config) {
		c.disabled = d
	}
}

// DisabledFn sets a dynamic function that is evaluated to determine whether
// the text field is disabled. When set, this takes precedence over the static value.
func DisabledFn(fn func() bool) Option {
	return func(c *config) {
		c.disabledFn = fn
	}
}

// A11yLabel sets the accessibility label for the text field.
// This is announced by screen readers to describe the field's purpose.
func A11yLabel(label string) Option {
	return func(c *config) {
		c.a11yLabel = label
	}
}

// PainterOpt sets the painter used to render the text field.
// Each design system provides its own painter. If not set,
// [DefaultPainter] is used.
func PainterOpt(p Painter) Option {
	return func(c *config) {
		c.painter = p
	}
}
