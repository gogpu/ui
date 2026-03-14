package progressbar

import (
	"github.com/gogpu/ui/state"
)

// Option configures a progress bar during construction.
type Option func(*config)

// Value sets the progress bar's initial static value (0.0 to 1.0).
// Values outside [0, 1] are clamped during rendering.
func Value(v float64) Option {
	return func(c *config) {
		c.value = v
	}
}

// ValueFn sets a dynamic value function that is evaluated on each draw.
// When set, this takes precedence over the static value but not over
// a signal set via [ValueSignal] or [ValueReadonlySignal].
func ValueFn(fn func() float64) Option {
	return func(c *config) {
		c.valueFn = fn
	}
}

// ValueSignal binds the progress bar's value to a reactive signal.
// This is a one-way read binding: the widget reads the value from the signal.
// When set, the signal value takes precedence over both [ValueFn] and [Value]
// but not over [ValueReadonlySignal].
func ValueSignal(sig state.Signal[float64]) Option {
	return func(c *config) {
		c.valueSignal = sig
	}
}

// ValueReadonlySignal binds the progress bar's value to a read-only signal.
// This is useful for computed signals created via [state.NewComputed].
// When set, this takes highest precedence over all other value sources.
func ValueReadonlySignal(sig state.ReadonlySignal[float64]) Option {
	return func(c *config) {
		c.readonlyValueSignal = sig
	}
}

// Height sets the bar height in logical pixels. Default is 8.
func Height(h float32) Option {
	return func(c *config) {
		c.height = h
	}
}

// Radius sets the corner radius for rounded bar ends. Default is 4.
// Set to 0 for square corners.
func Radius(r float32) Option {
	return func(c *config) {
		c.radius = r
		c.radiusSet = true
	}
}

// ShowLabel enables or disables the percentage label overlay.
// When enabled, the label is drawn centered over the bar.
func ShowLabel(show bool) Option {
	return func(c *config) {
		c.showLabel = show
	}
}

// FormatLabelFn sets a custom label formatting function.
// The function receives the current value (0.0 to 1.0) and returns
// the label string. If nil, the default "65%" format is used.
func FormatLabelFn(fn func(float64) string) Option {
	return func(c *config) {
		c.formatLabel = fn
	}
}

// ColorSchemeOpt sets the color scheme for painting.
// This overrides the painter's built-in defaults.
func ColorSchemeOpt(cs ProgressBarColorScheme) Option {
	return func(c *config) {
		c.colorScheme = cs
	}
}

// Disabled sets the progress bar's disabled state.
func Disabled(d bool) Option {
	return func(c *config) {
		c.disabled = d
	}
}

// DisabledFn sets a dynamic function for the disabled state.
// When set, this takes precedence over the static value but not
// over a signal set via [DisabledSignal].
func DisabledFn(fn func() bool) Option {
	return func(c *config) {
		c.disabledFn = fn
	}
}

// DisabledSignal binds the disabled state to a reactive signal.
// When set, the signal value takes precedence over both [DisabledFn]
// and [Disabled] but not over [DisabledReadonlySignal].
func DisabledSignal(sig state.Signal[bool]) Option {
	return func(c *config) {
		c.disabledSignal = sig
	}
}

// DisabledReadonlySignal binds the disabled state to a read-only signal.
// When set, this takes highest precedence over all other disabled sources.
func DisabledReadonlySignal(sig state.ReadonlySignal[bool]) Option {
	return func(c *config) {
		c.readonlyDisabledSignal = sig
	}
}

// PainterOpt sets the painter used to render the progress bar.
// Each design system provides its own painter. If not set,
// [DefaultPainter] is used.
func PainterOpt(p Painter) Option {
	return func(c *config) {
		c.painter = p
	}
}
