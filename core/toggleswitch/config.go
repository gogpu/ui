package toggleswitch

import "github.com/gogpu/ui/state"

// config holds the switch configuration, set at construction time via options.
type config struct {
	label               string
	labelFn             func() string
	labelSignal         state.Signal[string]
	readonlyLabelSig    state.ReadonlySignal[string]
	checked             bool
	checkedFn           func() bool
	checkedSignal       state.Signal[bool]
	onToggle            func(checked bool)
	disabled            bool
	disabledFn          func() bool
	disabledSignal      state.Signal[bool]
	readonlyDisabledSig state.ReadonlySignal[bool]
	a11yHint            string
	painter             Painter
}

// ResolvedLabel returns the current display label.
// Priority: ReadonlySignal > Signal > Fn > Static.
func (c *config) ResolvedLabel() string {
	if c.readonlyLabelSig != nil {
		return c.readonlyLabelSig.Get()
	}
	if c.labelSignal != nil {
		return c.labelSignal.Get()
	}
	if c.labelFn != nil {
		return c.labelFn()
	}
	return c.label
}

// ResolvedChecked returns the current on/off value.
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
// Priority: ReadonlySignal > Signal > Fn > Static.
func (c *config) ResolvedDisabled() bool {
	if c.readonlyDisabledSig != nil {
		return c.readonlyDisabledSig.Get()
	}
	if c.disabledSignal != nil {
		return c.disabledSignal.Get()
	}
	if c.disabledFn != nil {
		return c.disabledFn()
	}
	return c.disabled
}
