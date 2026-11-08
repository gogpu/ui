package dropdown

import "github.com/gogpu/ui/state"

// Option configures a dropdown during construction.
type Option func(*config)

// config holds the dropdown's configuration, set at construction time via options.
type config struct {
	items               []ItemDef
	itemsFn             func() []string
	itemsSignal         state.Signal[[]string]
	readonlyItemsSignal state.ReadonlySignal[[]string]
	selectedIndex       int
	placeholder         string
	onChange            func(index int, value string)
	disabled            bool
	disabledFn          func() bool
	maxVisibleItems     int
	painter             Painter
	signal              state.Signal[int]
	a11yHint            string
}

// defaultMaxVisible is the default number of visible items before scrolling.
const defaultMaxVisible = 8

// itemDefsFromStrings converts string items to ItemDefs. Each string becomes
// both the value and the label, as documented on [Items].
func itemDefsFromStrings(items []string) []ItemDef {
	if items == nil {
		return nil
	}
	defs := make([]ItemDef, len(items))
	for i, s := range items {
		defs[i] = ItemDef{Value: s}
	}
	return defs
}

// ResolvedDisabled returns the current disabled state, preferring the
// dynamic function over the static bool.
func (c *config) ResolvedDisabled() bool {
	if c.disabledFn != nil {
		return c.disabledFn()
	}
	return c.disabled
}

// ResolvedItems returns the current dropdown items.
// Priority: ReadonlySignal > Signal > Fn > Static.
//
// The signal and function sources are converted to ItemDefs on every call,
// so the returned slice is fresh and safe for the caller to keep (the menu
// overlay captures it when it opens).
func (c *config) ResolvedItems() []ItemDef {
	if c.readonlyItemsSignal != nil {
		return itemDefsFromStrings(c.readonlyItemsSignal.Get())
	}
	if c.itemsSignal != nil {
		return itemDefsFromStrings(c.itemsSignal.Get())
	}
	if c.itemsFn != nil {
		return itemDefsFromStrings(c.itemsFn())
	}
	return c.items
}

// Items sets the dropdown items from strings. Each string becomes both
// the value and label.
func Items(items ...string) Option {
	return func(c *config) {
		c.items = itemDefsFromStrings(items)
	}
}

// ItemsFn sets a dynamic function that supplies the dropdown items as strings.
// The function is evaluated whenever the widget reads its items.
// When set, it takes precedence over the static [Items] and [ItemDefs] but not
// over [ItemsSignal].
func ItemsFn(fn func() []string) Option {
	return func(c *config) {
		c.itemsFn = fn
	}
}

// ItemsSignal binds the dropdown's items to a reactive signal. When the
// signal is set, the trigger paints the new items and the selected index is
// clamped to a valid item (the last one, or -1 for an empty list, so the
// trigger falls back to the placeholder).
//
// A menu that is already open keeps the items it was built with; it picks up
// the new list on the next [Widget.Open], which always rebuilds the menu.
//
// The clamp is widget state only: a signal passed to [SelectedSignal] is not
// rewritten by it, and keeps the index the application last set. It is
// updated again when the user picks an item or [Widget.SetSelectedIndex]
// is called.
func ItemsSignal(sig state.Signal[[]string]) Option {
	return func(c *config) {
		c.itemsSignal = sig
	}
}

// ItemsReadonlySignal binds the dropdown's items to a read-only signal, for
// example one created with [state.NewComputed]. It takes precedence over
// [ItemsSignal], [ItemsFn] and the static items. The clamping and the
// open-menu behavior are the same as for [ItemsSignal].
func ItemsReadonlySignal(sig state.ReadonlySignal[[]string]) Option {
	return func(c *config) {
		c.readonlyItemsSignal = sig
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

// SelectedSignal binds the dropdown's selected index to a reactive signal for
// two-way data binding. When the selection changes, the signal is updated.
// When the signal is set externally, the dropdown reflects the change.
func SelectedSignal(sig state.Signal[int]) Option {
	return func(c *config) {
		c.signal = sig
	}
}

// Deprecated: Use SelectedSignal instead.
func Signal(s state.Signal[int]) Option {
	return SelectedSignal(s)
}

// A11yHint sets the accessibility hint text for the dropdown.
func A11yHint(hint string) Option {
	return func(c *config) {
		c.a11yHint = hint
	}
}
