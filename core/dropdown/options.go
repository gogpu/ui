package dropdown

import "github.com/gogpu/ui/state"

// Option configures a dropdown during construction.
type Option func(*config)

// config holds the dropdown's configuration, set at construction time via options.
type config struct {
	items           []ItemDef
	selectedIndex   int
	placeholder     string
	onChange        func(index int, value string)
	disabled        bool
	disabledFn      func() bool
	maxVisibleItems int
	painter         Painter
	signal          state.Signal[int]
	a11yHint        string
}

// defaultMaxVisible is the default number of visible items before scrolling.
const defaultMaxVisible = 8

// ResolvedDisabled returns the current disabled state, preferring the
// dynamic function over the static bool.
func (c *config) ResolvedDisabled() bool {
	if c.disabledFn != nil {
		return c.disabledFn()
	}
	return c.disabled
}

// Items sets the dropdown items from strings. Each string becomes both
// the value and label.
func Items(items ...string) Option {
	return func(c *config) {
		defs := make([]ItemDef, len(items))
		for i, s := range items {
			defs[i] = ItemDef{Value: s}
		}
		c.items = defs
	}
}

// ItemDefs sets the dropdown items from ItemDef definitions.
func ItemDefs(items []ItemDef) Option {
	return func(c *config) {
		c.items = items
	}
}

// Selected sets the initially selected item index. Use -1 for no selection.
func Selected(index int) Option {
	return func(c *config) {
		c.selectedIndex = index
	}
}

// Placeholder sets the text shown when no item is selected.
func Placeholder(text string) Option {
	return func(c *config) {
		c.placeholder = text
	}
}

// OnChange sets the callback invoked when the selection changes.
// The callback receives the new index and the selected item's value.
func OnChange(fn func(index int, value string)) Option {
	return func(c *config) {
		c.onChange = fn
	}
}

// Disabled sets the dropdown's disabled state.
func Disabled(d bool) Option {
	return func(c *config) {
		c.disabled = d
	}
}

// DisabledFn sets a dynamic function that determines whether the dropdown
// is disabled. When set, this takes precedence over the static value.
func DisabledFn(fn func() bool) Option {
	return func(c *config) {
		c.disabledFn = fn
	}
}

// MaxVisibleItems sets the maximum number of items visible in the menu
// before scrolling is required. Defaults to 8.
func MaxVisibleItems(n int) Option {
	return func(c *config) {
		if n > 0 {
			c.maxVisibleItems = n
		}
	}
}

// PainterOpt sets the painter used to render the dropdown.
func PainterOpt(p Painter) Option {
	return func(c *config) {
		c.painter = p
	}
}

// Signal binds the dropdown's selected index to a reactive signal for
// two-way data binding. When the selection changes, the signal is updated.
// When the signal is set externally, the dropdown reflects the change.
func Signal(s state.Signal[int]) Option {
	return func(c *config) {
		c.signal = s
	}
}

// A11yHint sets the accessibility hint text for the dropdown.
func A11yHint(hint string) Option {
	return func(c *config) {
		c.a11yHint = hint
	}
}
