// Package radio provides a mutually-exclusive radio group widget.
//
// A radio group contains a set of radio items, exactly one of which
// may be selected at a time. Construction uses functional options for
// immutable configuration:
//
//	rg := radio.NewGroup(
//	    radio.Items(
//	        radio.ItemDef{Value: "s", Label: "Small"},
//	        radio.ItemDef{Value: "m", Label: "Medium"},
//	        radio.ItemDef{Value: "l", Label: "Large"},
//	    ),
//	    radio.Selected("m"),
//	    radio.OnChange(handleChange),
//	)
//
// # Visual Style
//
// The visual rendering is provided by a [Painter] implementation.
// Each design system (Material 3, Fluent, Cupertino) supplies its own
// painter to render radio items in the appropriate visual style.
//
// If no painter is set, [DefaultPainter] is used, which draws a minimal
// gray radio button suitable for testing and prototyping.
//
// # Direction
//
// Items can be laid out vertically (default) or horizontally:
//   - [Vertical] -- items stacked top-to-bottom
//   - [Horizontal] -- items placed left-to-right
//
// # Keyboard Navigation
//
// Arrow keys navigate between items within a group:
//   - Up/Down for [Vertical] layout
//   - Left/Right for [Horizontal] layout
//
// Space or Enter on a focused item selects it.
//
// # Interaction
//
// Clicking a radio item selects it and deselects the previously selected
// item. Disabled groups ignore all interaction and are drawn with a
// dimmed appearance.
//
// # Focus
//
// Individual radio items implement [widget.Focusable] and participate in
// tab navigation. The group manages focus transfer between items via
// arrow keys.
package radio
