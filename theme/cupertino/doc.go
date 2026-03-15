// Package cupertino provides an Apple Human Interface Guidelines (HIG) theme.
//
// The Cupertino theme follows Apple's design language with system blue accent
// colors, rounded pill-shaped buttons, iOS-style toggle switches for checkboxes,
// thin auto-hiding scrollbars, sheet-style dialogs, and segmented control tab bars.
//
// # Creating a Theme
//
// Create a theme with the default system blue accent:
//
//	theme := cupertino.NewTheme()
//	accent := theme.Colors.Accent
//	radius := theme.Radius
//
// Or customize the accent color:
//
//	theme := cupertino.NewTheme(cupertino.WithAccentColor(widget.Hex(0x34C759)))
//
// # Light and Dark Schemes
//
// Both light and dark themes are available:
//
//	lightTheme := cupertino.NewTheme()
//	darkTheme := cupertino.NewDarkTheme()
//
// # Component Painters
//
// The package provides painter implementations that render UI components
// according to Apple HIG specifications. Painters implement the interfaces
// defined in core packages and can be injected into widgets:
//
//   - [ButtonPainter] implements [button.Painter] for rounded pill buttons
//   - [CheckboxPainter] implements [checkbox.Painter] for iOS toggle switches
//   - [RadioPainter] implements [radio.Painter] for radio items
//   - [TextFieldPainter] implements [textfield.Painter] for rounded text fields
//   - [DropdownPainter] implements [dropdown.Painter] for dropdown menus
//   - [SliderPainter] implements [slider.Painter] for thin track sliders
//   - [DialogPainter] implements [dialog.Painter] for sheet-style dialogs
//   - [ScrollbarPainter] implements [scrollview.Painter] for thin scrollbars
//   - [TabViewPainter] implements [tabview.Painter] for segmented controls
//
// Example:
//
//	btn := button.New(
//	    button.Text("Submit"),
//	    button.PainterOpt(cupertino.ButtonPainter{Theme: cupertinoTheme}),
//	)
package cupertino
