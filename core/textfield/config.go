package textfield

import "github.com/gogpu/ui/state"

// config holds the text field's configuration, set at construction time via options.
type config struct {
	placeholder string
	value       string
	signal      state.Signal[string]
	onChange    func(string)
	onSubmit    func(string)
	inputType   InputType
	maxLength   int
	validation  []ValidationFunc
	disabled    bool
	disabledFn  func() bool
	a11yLabel   string
	painter     Painter
}

// ResolvedDisabled returns the current disabled state, preferring the
// dynamic function over the static bool.
func (c *config) ResolvedDisabled() bool {
	if c.disabledFn != nil {
		return c.disabledFn()
	}
	return c.disabled
}
