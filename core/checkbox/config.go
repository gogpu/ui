package checkbox

import "github.com/gogpu/ui/widget"

// config holds the checkbox's configuration, set at construction time via options.
type config struct {
	label         string
	labelFn       func() string
	checked       bool
	checkedFn     func() bool
	onToggle      func(checked bool)
	disabled      bool
	disabledFn    func() bool
	indeterminate bool
	a11yHint      string
	// styling overrides (nil/zero means use defaults)
	background *widget.Color
	painter    Painter
}

// ResolvedLabel returns the current display label, preferring the dynamic
// label function over the static string.
func (c *config) ResolvedLabel() string {
	if c.labelFn != nil {
		return c.labelFn()
	}
	return c.label
}

// ResolvedChecked returns the current checked state, preferring the dynamic
// function over the static bool.
func (c *config) ResolvedChecked() bool {
	if c.checkedFn != nil {
		return c.checkedFn()
	}
	return c.checked
}

// ResolvedDisabled returns the current disabled state, preferring the
// dynamic function over the static bool.
func (c *config) ResolvedDisabled() bool {
	if c.disabledFn != nil {
		return c.disabledFn()
	}
	return c.disabled
}
