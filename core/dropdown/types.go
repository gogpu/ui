// Package dropdown provides a dropdown/select widget that displays a list of
// items in a floating menu when activated. The user can select an item from the
// list using mouse clicks or keyboard navigation (Up/Down/Enter/Escape).
//
// The dropdown consists of two parts:
//   - A trigger element (button-like) that shows the current selection
//   - A menu overlay that appears below the trigger when activated
//
// Dropdown follows the functional options pattern for construction:
//
//	dd := dropdown.New(
//	    dropdown.Items("Red", "Green", "Blue"),
//	    dropdown.Selected(0),
//	    dropdown.OnChange(func(index int, value string) {
//	        fmt.Println("Selected:", value)
//	    }),
//	)
//
// The visual appearance is controlled by a pluggable [Painter] interface.
// The default painter provides a minimal fallback; use a design system painter
// (e.g., material3.DropdownPainter) for production styling.
package dropdown

// ItemDef defines a single item in the dropdown list.
type ItemDef struct {
	// Value is the internal value associated with this item.
	Value string

	// Label is the display text. If empty, Value is used as the display text.
	Label string

	// Disabled prevents this item from being selected.
	Disabled bool
}

// DisplayText returns the text to display for this item.
// Returns Label if set, otherwise Value.
func (d ItemDef) DisplayText() string {
	if d.Label != "" {
		return d.Label
	}
	return d.Value
}
