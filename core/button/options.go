package button

import "github.com/gogpu/ui/widget"

// Option configures a button during construction.
type Option func(*config)

// TextOpt sets the button's static display text.
func TextOpt(s string) Option {
	return func(c *config) {
		c.text = s
	}
}

// TextFn sets a dynamic text function that is evaluated on each draw.
// When set, this takes precedence over the static text.
func TextFn(fn func() string) Option {
	return func(c *config) {
		c.textFn = fn
	}
}

// OnClick sets the callback invoked when the button is activated
// (mouse click or keyboard Enter/Space).
func OnClick(fn func()) Option {
	return func(c *config) {
		c.onClick = fn
	}
}

// Disabled sets the button's disabled state. A disabled button does not
// respond to user input and is drawn with a dimmed appearance.
func Disabled(d bool) Option {
	return func(c *config) {
		c.disabled = d
	}
}

// DisabledFn sets a dynamic function that is evaluated to determine whether
// the button is disabled. When set, this takes precedence over the static value.
func DisabledFn(fn func() bool) Option {
	return func(c *config) {
		c.disabledFn = fn
	}
}

// VariantOpt sets the button's visual variant.
func VariantOpt(v Variant) Option {
	return func(c *config) {
		c.variant = v
	}
}

// SizeOpt sets the button's size.
func SizeOpt(s Size) Option {
	return func(c *config) {
		c.size = s
	}
}

// A11yHint sets the accessibility hint text for the button.
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

// RoundedOpt sets a custom corner radius override.
func RoundedOpt(radius float32) Option {
	return func(c *config) {
		c.rounded = &radius
	}
}

// PainterOpt sets the painter used to render the button.
// Each design system provides its own painter. If not set,
// [DefaultPainter] is used.
func PainterOpt(p Painter) Option {
	return func(c *config) {
		c.painter = p
	}
}
