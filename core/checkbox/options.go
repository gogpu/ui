package checkbox

import (
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// Option configures a checkbox during construction.
type Option func(*config)

// LabelOpt sets the checkbox's static display label.
func LabelOpt(s string) Option {
	return func(c *config) {
		c.label = s
	}
}

// LabelFn sets a dynamic label function that is evaluated on each draw.
// When set, this takes precedence over the static label but not over
// a signal set via [LabelSignal].
func LabelFn(fn func() string) Option {
	return func(c *config) {
		c.labelFn = fn
	}
}

// LabelSignal binds the checkbox's display label to a reactive signal.
// When set, the signal value takes precedence over both [LabelFn] and [LabelOpt].
func LabelSignal(sig state.Signal[string]) Option {
	return func(c *config) {
		c.labelSignal = sig
	}
}

// Checked sets the checkbox's initial checked state.
func Checked(b bool) Option {
	return func(c *config) {
		c.checked = b
	}
}

// CheckedFn sets a dynamic function that is evaluated to determine whether
// the checkbox is checked. When set, this takes precedence over the static value
// but not over a signal set via [CheckedSignal].
func CheckedFn(fn func() bool) Option {
	return func(c *config) {
		c.checkedFn = fn
	}
}

// CheckedSignal binds the checkbox's checked state to a reactive signal.
// This is a TWO-WAY binding: the widget reads the checked state from the signal,
// and when the user toggles the checkbox, the new state is written back to the signal.
// When set, the signal value takes precedence over both [CheckedFn] and [Checked].
func CheckedSignal(sig state.Signal[bool]) Option {
	return func(c *config) {
		c.checkedSignal = sig
	}
}

// OnToggle sets the callback invoked when the checkbox is toggled.
// The callback receives the new checked state.
func OnToggle(fn func(checked bool)) Option {
	return func(c *config) {
		c.onToggle = fn
	}
}

// Disabled sets the checkbox's disabled state. A disabled checkbox does not
// respond to user input and is drawn with a dimmed appearance.
func Disabled(d bool) Option {
	return func(c *config) {
		c.disabled = d
	}
}

// DisabledFn sets a dynamic function that is evaluated to determine whether
// the checkbox is disabled. When set, this takes precedence over the static value
// but not over a signal set via [DisabledSignal].
func DisabledFn(fn func() bool) Option {
	return func(c *config) {
		c.disabledFn = fn
	}
}

// DisabledSignal binds the checkbox's disabled state to a reactive signal.
// When set, the signal value takes precedence over both [DisabledFn] and [Disabled].
func DisabledSignal(sig state.Signal[bool]) Option {
	return func(c *config) {
		c.disabledSignal = sig
	}
}

// Indeterminate sets the checkbox to the indeterminate (mixed) state.
// An indeterminate checkbox displays a horizontal dash instead of a checkmark.
func Indeterminate(b bool) Option {
	return func(c *config) {
		c.indeterminate = b
	}
}

// A11yHint sets the accessibility hint text for the checkbox.
func A11yHint(hint string) Option {
	return func(c *config) {
		c.a11yHint = hint
	}
}

// BackgroundOpt sets a custom background color override.
func BackgroundOpt(color widget.Color) Option {
	return func(c *config) {
		c.background = &color
	}
}

// PainterOpt sets the painter used to render the checkbox.
// Each design system provides its own painter. If not set,
// [DefaultPainter] is used.
func PainterOpt(p Painter) Option {
	return func(c *config) {
		c.painter = p
	}
}
