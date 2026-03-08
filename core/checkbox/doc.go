// Package checkbox provides a toggleable checkbox widget.
//
// Construction uses functional options for immutable configuration,
// while fluent methods handle mutable styling:
//
//	cb := checkbox.New(
//	    checkbox.Label("Accept terms"),
//	    checkbox.OnToggle(handleToggle),
//	    checkbox.Checked(true),
//	).Padding(8)
//
// # Visual Style
//
// The visual rendering is provided by a [Painter] implementation.
// Each design system (Material 3, Fluent, Cupertino) supplies its own
// painter to render checkboxes in the appropriate visual style.
//
// If no painter is set, [DefaultPainter] is used, which draws a minimal
// gray checkbox suitable for testing and prototyping.
//
// # States
//
// A checkbox has three visual check states:
//   - Unchecked (default) -- empty box with a border
//   - Checked -- filled box with a checkmark
//   - Indeterminate -- filled box with a horizontal dash
//
// The indeterminate state is used for "select all" checkboxes when only
// some items are selected.
//
// # Signal Binding
//
// Checkbox properties can be bound to reactive signals from the [state] package.
// Signals take highest priority in value resolution (Signal > Fn > Static).
//
//   - [LabelSignal] — one-way binding for the display label
//   - [CheckedSignal] — TWO-WAY binding: reads from signal AND writes back on toggle
//   - [DisabledSignal] — one-way binding for the disabled state
//
// Example:
//
//	checked := state.NewSignal(false)
//	cb := checkbox.New(
//	    checkbox.Label("Accept"),
//	    checkbox.CheckedSignal(checked),
//	)
//	// User toggles checkbox → checked.Get() returns new value
//	// checked.Set(true) → checkbox reflects new state on next draw
//
// # Interaction
//
// Checkboxes respond to mouse click (left button) and keyboard activation
// (Space when focused). Each activation toggles the checked state and
// invokes the [OnToggle] callback. Disabled checkboxes ignore all
// interaction and are drawn with a dimmed appearance.
//
// # Focus
//
// Checkboxes implement [widget.Focusable] and participate in tab navigation.
// A focus ring is drawn when the checkbox has keyboard focus.
package checkbox
