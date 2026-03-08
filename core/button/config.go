package button

import (
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// config holds the button's configuration, set at construction time via options.
type config struct {
	text           string
	textFn         func() string
	textSignal     state.Signal[string]
	onClick        func()
	disabled       bool
	disabledFn     func() bool
	disabledSignal state.Signal[bool]
	variant        Variant
	size           Size
	a11yHint       string
	// styling overrides (nil/zero means use defaults)
	background *widget.Color
	rounded    *float32
	painter    Painter
}

// ResolvedText returns the current display text.
// Priority: Signal > Fn > Static.
func (c *config) ResolvedText() string {
	if c.textSignal != nil {
		return c.textSignal.Get()
	}
	if c.textFn != nil {
		return c.textFn()
	}
	return c.text
}

// ResolvedDisabled returns the current disabled state.
// Priority: Signal > Fn > Static.
func (c *config) ResolvedDisabled() bool {
	if c.disabledSignal != nil {
		return c.disabledSignal.Get()
	}
	if c.disabledFn != nil {
		return c.disabledFn()
	}
	return c.disabled
}
