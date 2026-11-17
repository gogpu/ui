package toggleswitch

import "github.com/gogpu/ui/state"

// Option configures a switch during construction.
type Option func(*config)

// Label sets the static label drawn next to the switch.
func Label(s string) Option {
	return func(c *config) {
		c.label = s
	}
}

// LabelFn sets a label function evaluated on each layout and draw.
// It takes precedence over [Label] but not over [LabelSignal].
func LabelFn(fn func() string) Option {
	return func(c *config) {
		c.labelFn = fn
	}
}

// LabelSignal binds the label to a reactive signal.
// It takes precedence over [LabelFn] and [Label] but not over
// [LabelReadonlySignal].
func LabelSignal(sig state.Signal[string]) Option {
	return func(c *config) {
		c.labelSignal = sig
	}
}

// LabelReadonlySignal binds the label to a read-only signal, such as one
// created with [state.NewComputed]. It takes precedence over every other
// label source.
func LabelReadonlySignal(sig state.ReadonlySignal[string]) Option {
	return func(c *config) {
		c.readonlyLabelSig = sig
	}
}

// Checked sets the initial value: true for on, false for off.
func Checked(on bool) Option {
	return func(c *config) {
		c.checked = on
	}
}

// CheckedFn sets a function that reports the value on each draw. The caller
// owns the state: toggling only calls [OnToggle]. It takes precedence over
// [Checked] but not over [CheckedSignal].
func CheckedFn(fn func() bool) Option {
	return func(c *config) {
		c.checkedFn = fn
	}
}

// CheckedSignal binds the value to a reactive signal. This is a TWO-WAY
// binding: the switch reads its value from the signal, and a user toggle
// writes the new value back. It takes precedence over [CheckedFn] and
// [Checked].
func CheckedSignal(sig state.Signal[bool]) Option {
	return func(c *config) {
		c.checkedSignal = sig
	}
}

// OnToggle sets the callback invoked after the user toggles the switch.
// The callback receives the new value.
func OnToggle(fn func(checked bool)) Option {
	return func(c *config) {
		c.onToggle = fn
	}
}

// Disabled sets the disabled state. A disabled switch ignores input and is
// drawn dimmed.
func Disabled(d bool) Option {
	return func(c *config) {
		c.disabled = d
	}
}

// DisabledFn sets a function that reports the disabled state. It takes
// precedence over [Disabled] but not over [DisabledSignal].
func DisabledFn(fn func() bool) Option {
	return func(c *config) {
		c.disabledFn = fn
	}
}

// DisabledSignal binds the disabled state to a reactive signal. It takes
// precedence over [DisabledFn] and [Disabled] but not over
// [DisabledReadonlySignal].
func DisabledSignal(sig state.Signal[bool]) Option {
	return func(c *config) {
		c.disabledSignal = sig
	}
}

// DisabledReadonlySignal binds the disabled state to a read-only signal,
// such as one created with [state.NewComputed]. It takes precedence over
// every other disabled source.
func DisabledReadonlySignal(sig state.ReadonlySignal[bool]) Option {
	return func(c *config) {
		c.readonlyDisabledSig = sig
	}
}

// A11yHint sets the accessibility hint announced for the switch.
func A11yHint(hint string) Option {
	return func(c *config) {
		c.a11yHint = hint
	}
}

// PainterOpt sets the painter that renders the switch. Each design system
// provides one. Without it, [DefaultPainter] is used.
func PainterOpt(p Painter) Option {
	return func(c *config) {
		c.painter = p
	}
}
