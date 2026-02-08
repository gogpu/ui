// Package material3 provides a Google Material Design 3 (Material You) theme.
//
// Material 3 uses a seed color to generate a complete, harmonious color
// scheme using HCT (Hue, Chroma, Tone) color science. From one color,
// the system derives primary, secondary, tertiary, and neutral palettes
// with proper contrast ratios for accessibility.
//
// # Creating a Theme
//
// Create a theme from a seed color:
//
//	theme := material3.New(widget.Hex(0x6750A4))
//	primary := theme.Colors.Primary
//	fontSize := theme.Typography.BodyMedium.FontSize
//	radius := theme.Shape.Medium
//
// # Light and Dark Schemes
//
// Light and dark color schemes are available:
//
//	lightColors := material3.Light(seedColor)
//	darkColors := material3.Dark(seedColor)
//
// Or create a complete dark theme:
//
//	darkTheme := material3.NewDark(seedColor)
//
// # Color Roles
//
// The color scheme includes the following role groups:
//
//   - Primary: key brand color and its container variant
//   - Secondary: accent color for less prominent elements
//   - Tertiary: complementary color for contrast and balance
//   - Error: error states and destructive actions
//   - Surface: neutral backgrounds at various elevation levels
//   - Outline: borders and dividers
//
// Each color role has a corresponding "on" color for text/icons placed on top.
//
// # Typography
//
// The type scale provides 15 text styles organized into 5 categories:
//
//   - Display: large, impactful text
//   - Headline: section headers
//   - Title: component titles
//   - Body: primary reading text
//   - Label: UI labels and buttons
//
// # Shape
//
// The shape scale provides corner radius values from None (0dp) to Full (pill).
//
// # Component Painters
//
// The package provides painter implementations that render UI components
// according to Material 3 design specifications. Painters implement the
// interfaces defined in core packages and can be injected into widgets:
//
//   - [ButtonPainter] implements [button.Painter] for Material 3 button rendering
//
// Example:
//
//	btn := button.New(
//	    button.Text("Submit"),
//	    button.PainterOpt(material3.ButtonPainter{}),
//	)
package material3
