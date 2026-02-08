// Package button provides a clickable button widget.
//
// Construction uses functional options for immutable configuration,
// while fluent methods handle mutable styling:
//
//	btn := button.New(
//	    button.Text("Submit"),
//	    button.OnClick(handleSubmit),
//	    button.VariantOpt(button.Filled),
//	).Padding(16).Rounded(8)
//
// # Visual Style
//
// The visual rendering is provided by a [Painter] implementation.
// Each design system (Material 3, Fluent, Cupertino) supplies its own
// painter to render buttons in the appropriate visual style.
//
// If no painter is set, [DefaultPainter] is used, which draws a minimal
// gray button suitable for testing and prototyping.
//
// # Variants
//
// Four semantic variants are available:
//   - [Filled] (default) -- solid background with contrasting text
//   - [Outlined] -- transparent background with a border
//   - [TextOnly] -- no background or border, only text
//   - [Tonal] -- tinted background (lower emphasis than Filled)
//
// The interpretation of each variant depends on the active [Painter].
//
// # Sizes
//
// Three sizes control the button height:
//   - [Small] -- 32px height
//   - [Medium] (default) -- 40px height
//   - [Large] -- 48px height
//
// # Interaction
//
// Buttons respond to mouse events (hover, press, click) and keyboard
// activation (Enter or Space when focused). Disabled buttons ignore
// all interaction and are drawn with a dimmed appearance.
//
// # Focus
//
// Buttons implement [widget.Focusable] and participate in tab navigation.
// A focus ring is drawn when the button has keyboard focus.
package button
