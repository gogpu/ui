package radio

// GroupOption configures a radio group during construction.
type GroupOption func(*groupConfig)

// OnChange sets the callback invoked when the selected item changes.
// The callback receives the value of the newly selected item.
func OnChange(fn func(value string)) GroupOption {
	return func(c *groupConfig) {
		c.onChange = fn
	}
}

// Selected sets the initially selected item by value.
// If no item matches the value, no item is selected.
func Selected(value string) GroupOption {
	return func(c *groupConfig) {
		c.selected = value
	}
}

// DirectionOpt sets the layout direction for the group's items.
func DirectionOpt(d Direction) GroupOption {
	return func(c *groupConfig) {
		c.direction = d
	}
}

// GroupDisabled sets the group's disabled state. A disabled group does not
// respond to user input and all items are drawn with a dimmed appearance.
func GroupDisabled(d bool) GroupOption {
	return func(c *groupConfig) {
		c.disabled = d
	}
}

// GroupDisabledFn sets a dynamic function that is evaluated to determine
// whether the group is disabled. When set, this takes precedence over the
// static value.
func GroupDisabledFn(fn func() bool) GroupOption {
	return func(c *groupConfig) {
		c.disabledFn = fn
	}
}

// GroupA11yLabel sets the accessibility label for the radio group.
func GroupA11yLabel(s string) GroupOption {
	return func(c *groupConfig) {
		c.a11yLabel = s
	}
}

// GroupPainter sets the painter used to render each radio item.
// Each design system provides its own painter. If not set,
// [DefaultPainter] is used.
func GroupPainter(p Painter) GroupOption {
	return func(c *groupConfig) {
		c.painter = p
	}
}

// Items sets the item definitions for the group.
// Each ItemDef describes a single radio item's value and label.
func Items(defs ...ItemDef) GroupOption {
	return func(c *groupConfig) {
		c.items = defs
	}
}

// ItemDef describes a radio item's value and display label.
type ItemDef struct {
	// Value is the programmatic identifier returned by [Group.Selected].
	Value string

	// Label is the human-readable text displayed next to the radio circle.
	Label string
}
