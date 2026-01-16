// Package theme provides a comprehensive theming system for gogpu/ui.
//
// The theme package enables consistent visual styling across an application by
// providing centralized definitions for colors, typography, spacing, shadows,
// and border radii. It supports both light and dark modes, with automatic
// system preference detection (platform-dependent).
//
// # Quick Start
//
// Use the default themes for immediate styling:
//
//	// Get the default light theme
//	light := theme.DefaultLight()
//
//	// Get the default dark theme
//	dark := theme.DefaultDark()
//
//	// Access theme properties
//	primaryColor := light.Colors.Primary
//	bodyFont := light.Typography.BodyMedium
//	padding := light.Spacing.M
//
// # Theme Structure
//
// A Theme consists of several components:
//
//   - Colors: Semantic color palette (primary, secondary, background, etc.)
//   - Typography: Font families and text styles (display, headline, body, label)
//   - Spacing: Consistent spacing scale (XS to XXL)
//   - Shadows: Elevation-based shadow definitions
//   - Radii: Border radius scale for rounded corners
//   - Mode: Light, Dark, or System preference
//
// # Color Palette
//
// The color palette uses semantic naming following Material 3 guidelines:
//
//	theme.Colors.Primary       // Brand primary color
//	theme.Colors.Secondary     // Secondary accent color
//	theme.Colors.Background    // Window/page background
//	theme.Colors.Surface       // Card/panel surfaces
//	theme.Colors.Error         // Error states
//	theme.Colors.OnPrimary     // Text on primary color
//	theme.Colors.OnSurface     // Text on surface color
//
// # Typography
//
// Typography follows the Material 3 type scale:
//
//	// Display - Large, impactful text
//	theme.Typography.DisplayLarge
//	theme.Typography.DisplayMedium
//	theme.Typography.DisplaySmall
//
//	// Headline - Section headers
//	theme.Typography.HeadlineLarge
//	theme.Typography.HeadlineMedium
//	theme.Typography.HeadlineSmall
//
//	// Title - Smaller headings
//	theme.Typography.TitleLarge
//	theme.Typography.TitleMedium
//	theme.Typography.TitleSmall
//
//	// Body - Primary reading text
//	theme.Typography.BodyLarge
//	theme.Typography.BodyMedium
//	theme.Typography.BodySmall
//
//	// Label - UI labels and buttons
//	theme.Typography.LabelLarge
//	theme.Typography.LabelMedium
//	theme.Typography.LabelSmall
//
// # Thread Safety
//
// Themes are designed to be immutable by convention. Once created, a theme
// should not be modified. This allows themes to be safely shared across
// goroutines without synchronization. If you need different themes for
// different parts of your application, create separate Theme instances.
//
// # Extensions
//
// The Theme struct includes an extensions map for custom theme data:
//
//	theme := theme.DefaultLight()
//	theme.Extensions["my-component"] = MyComponentTheme{...}
//
// This allows third-party components to add their own theme properties
// while maintaining compatibility with the core theme system.
//
// # Creating Custom Themes
//
// Build custom themes by modifying a base theme:
//
//	func MyBrandTheme() *theme.Theme {
//	    t := theme.DefaultLight()
//	    t.Colors.Primary = widget.Hex(0x6200EE)      // Purple
//	    t.Colors.Secondary = widget.Hex(0x03DAC6)    // Teal
//	    return t
//	}
//
// # System Preference Detection
//
// The ThemeMode.System value indicates that the theme should follow the
// operating system's color scheme preference. Actual detection is handled
// by the platform integration layer (not this package).
//
//	if theme.Mode == theme.ModeSystem {
//	    // Platform layer should detect OS preference
//	    // and apply appropriate light/dark theme
//	}
package theme
