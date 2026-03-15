// Package fluent provides a Microsoft Fluent Design System theme.
//
// Fluent Design uses an accent color to derive a consistent color scheme
// with subtle depth cues, clean typography, and restrained rounded corners.
// The default accent color is Windows Blue (#0078D4), but it can be
// customized to match any brand.
//
// # Creating a Theme
//
// Create a theme from an accent color:
//
//	t := fluent.NewTheme(fluent.WithAccentColor(widget.Hex(0x0078D4)))
//	primary := t.Colors.Primary
//
// Or use defaults:
//
//	lightTheme := fluent.NewTheme()        // light mode
//	darkTheme := fluent.NewDarkTheme()     // dark mode
//
// # Design Characteristics
//
//   - Accent color: Windows Blue (#0078D4) by default, customizable
//   - Rounded corners: 4px default (smaller than Material 3)
//   - Shadows: Subtle elevation with lighter shadow values
//   - Typography: Segoe UI-like metrics using Inter font
//   - Focus: Inner focus ring style
//
// # Component Painters
//
// The package provides painter implementations for all core widgets:
//
//   - [ButtonPainter] implements [button.Painter]
//   - [CheckboxPainter] implements [checkbox.Painter]
//   - [RadioPainter] implements [radio.Painter]
//   - [TextFieldPainter] implements [textfield.Painter]
//   - [DropdownPainter] implements [dropdown.Painter]
//   - [SliderPainter] implements [slider.Painter]
//   - [DialogPainter] implements [dialog.Painter]
//   - [ScrollbarPainter] implements [scrollview.Painter]
//   - [TabViewPainter] implements [tabview.Painter]
//
// Example:
//
//	btn := button.New(
//	    button.Text("Submit"),
//	    button.PainterOpt(fluent.ButtonPainter{Theme: fluentTheme}),
//	)
package fluent
