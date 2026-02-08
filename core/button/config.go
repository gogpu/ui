package button

import "github.com/gogpu/ui/widget"

// config holds the button's configuration, set at construction time via options.
type config struct {
	text       string
	textFn     func() string
	onClick    func()
	disabled   bool
	disabledFn func() bool
	variant    Variant
	size       Size
	a11yHint   string
	// styling overrides (nil/zero means use defaults)
	background *widget.Color
	rounded    *float32
	painter    Painter
}

// ResolvedText returns the current display text, preferring the dynamic
// text function over the static string.
func (c *config) ResolvedText() string {
	if c.textFn != nil {
		return c.textFn()
	}
	return c.text
}

// ResolvedDisabled returns the current disabled state, preferring the
// dynamic function over the static bool.
func (c *config) ResolvedDisabled() bool {
	if c.disabledFn != nil {
		return c.disabledFn()
	}
	return c.disabled
}
