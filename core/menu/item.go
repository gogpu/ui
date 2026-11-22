package menu

// MenuItem defines a single entry in a menu. It can be a regular action item,
// a separator, or a submenu parent containing children.
type MenuItem struct {
	// Label is the display text for this item.
	Label string

	// Shortcut is the keyboard shortcut display text (e.g., "Ctrl+S").
	// This is cosmetic only; actual shortcut handling is separate.
	Shortcut string

	// OnAction is the callback invoked when the item is activated.
	// Nil for separators and submenu parents.
	OnAction func()

	// Disabled prevents this item from being activated.
	Disabled bool

	// Checkable marks this item as having a checkable state gutter.
	Checkable bool

	// Checked renders a checkmark beside this item when Checkable is true.
	Checked bool

	// Children holds submenu items. If non-empty, this item opens a submenu
	// on hover instead of invoking OnAction.
	Children []MenuItem

	// isSeparator marks this item as a horizontal separator line.
	isSeparator bool
}

// IsSeparator reports whether this item is a separator line.
func (m MenuItem) IsSeparator() bool {
	return m.isSeparator
}

// HasChildren reports whether this item has a submenu.
func (m MenuItem) HasChildren() bool {
	return len(m.Children) > 0
}

// Item creates a regular menu item with a label, shortcut display text,
// and action callback.
func Item(label, shortcut string, onAction func()) MenuItem {
	return MenuItem{
		Label:    label,
		Shortcut: shortcut,
		OnAction: onAction,
	}
}

// ItemCheck creates a checkable menu item with an optional checkmark indicator.
func ItemCheck(label, shortcut string, checked bool, onAction func()) MenuItem {
	return MenuItem{
		Label:     label,
		Shortcut:  shortcut,
		OnAction:  onAction,
		Checkable: true,
		Checked:   checked,
	}
}

// ItemDisabled creates a disabled menu item that cannot be activated.
func ItemDisabled(label, shortcut string) MenuItem {
	return MenuItem{
		Label:    label,
		Shortcut: shortcut,
		Disabled: true,
	}
}

// Sep creates a separator menu item displayed as a horizontal line.
func Sep() MenuItem {
	return MenuItem{isSeparator: true}
}

// SubMenu creates a menu item that opens a nested submenu on hover.
func SubMenu(label string, children ...MenuItem) MenuItem {
	return MenuItem{
		Label:    label,
		Children: children,
	}
}

// TopMenu defines a top-level menu in a MenuBar. It has a label shown
// in the bar and a list of items displayed in the dropdown.
type TopMenu struct {
	// Label is the text displayed in the menu bar.
	Label string

	// Items are the menu entries shown when this top-level menu is opened.
	Items []MenuItem
}

// BarMenu creates a TopMenu definition for use with NewBar.
func BarMenu(label string, items ...MenuItem) TopMenu {
	return TopMenu{
		Label: label,
		Items: items,
	}
}
