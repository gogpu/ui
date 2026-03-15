// Package dropdown provides a dropdown/select widget that displays a list of
// items in a floating menu when activated.
//
// The user selects an item from the list using mouse clicks or keyboard
// navigation (Up/Down arrows to highlight, Enter to confirm, Escape to close).
// The dropdown consists of two parts: a trigger element that shows the
// current selection, and a menu overlay that appears below the trigger
// when activated.
//
// # Construction
//
// Construction uses functional options for immutable configuration:
//
//	dd := dropdown.New(
//	    dropdown.Items("Red", "Green", "Blue"),
//	    dropdown.Selected(0),
//	    dropdown.Placeholder("Choose a color..."),
//	    dropdown.OnChange(func(index int, value string) {
//	        fmt.Println("Selected:", value)
//	    }),
//	)
//
// For items with separate labels and values, use [ItemDefs]:
//
//	dd := dropdown.New(
//	    dropdown.ItemDefs([]dropdown.ItemDef{
//	        {Value: "sm", Label: "Small"},
//	        {Value: "md", Label: "Medium"},
//	        {Value: "lg", Label: "Large"},
//	    }),
//	)
//
// # Visual Style
//
// The visual rendering is provided by a [Painter] implementation.
// Each design system (Material 3, Fluent, Cupertino) supplies its own
// painter to render dropdowns in the appropriate visual style.
//
// If no painter is set, [DefaultPainter] is used, which draws a minimal
// dropdown suitable for testing and prototyping.
//
// # Overlay Menu
//
// The menu is displayed as an [overlay.Overlay] obtained from
// [widget.Context]. This keeps the menu above other widgets and allows
// it to extend beyond the dropdown's own bounds. The menu supports:
//   - Scrolling when items exceed [MaxVisibleItems]
//   - Keyboard navigation within the menu
//   - Click-outside dismissal
//
// # Signal Binding
//
// The selected index can be bound to a reactive signal for two-way
// synchronization:
//
//	idx := state.NewSignal(0)
//	dd := dropdown.New(
//	    dropdown.Items("A", "B", "C"),
//	    dropdown.SelectedSignal(idx),
//	)
//
// # Focus
//
// Dropdowns implement [widget.Focusable] and participate in tab navigation.
// When focused, pressing Space or Enter opens the menu.
package dropdown
