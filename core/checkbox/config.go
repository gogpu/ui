package checkbox

import (
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// config holds the checkbox's configuration, set at construction time via options.
type config struct {
	label          string
	labelFn        func() string
	labelSignal    state.Signal[string]
	checked        bool
	checkedFn      func() bool
	checkedSignal  state.Signal[bool]
	onToggle       func(checked bool)
	disabled       bool
	disabledFn     func() bool
	disabledSignal state.Signal[bool]
	indeterminate  bool
	a11yHint       string
	// styling overrides (nil/zero means use defaults)
	background *widget.Color
	painter    Painter
}

// ResolvedLabel returns the current display label.
// Priority: Signal > Fn > Static.
func (c *config) ResolvedLabel() string {
	if c.labelSignal != nil {
		return c.labelSignal.Get()
	}
	if c.labelFn != nil {
		return c.labelFn()
	}
	return c.label
}

// ResolvedChecked returns the current checked state.
// Priority: Signal > Fn > Static.
func (c *config) ResolvedChecked() bool {
	if c.checkedSignal != nil {
		return c.checkedSignal.Get()
	}
	if c.checkedFn != nil {
		return c.checkedFn()
	}
	return c.checked
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
