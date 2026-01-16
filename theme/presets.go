package theme

import "github.com/gogpu/ui/widget"

// DefaultLight returns a pre-configured light theme following Material 3 guidelines.
//
// The light theme uses a white background with dark text and blue primary colors.
// It's suitable for well-lit environments and is often the default choice for
// applications.
//
// Color scheme:
//   - Primary: Blue (#1976D2)
//   - Secondary: Teal (#009688)
//   - Background: White (#FFFFFF)
//   - Surface: Light gray (#FAFAFA)
//   - Error: Red (#D32F2F)
//
// Example:
//
//	theme := theme.DefaultLight()
//	primaryColor := theme.Colors.Primary
//	bodyStyle := theme.Typography.BodyMedium
func DefaultLight() *Theme {
	return &Theme{
		Name: "Light",
		Mode: ModeLight,
		Colors: ColorPalette{
			// Primary colors - Material Blue
			Primary:      widget.Hex(0x1976D2), // Blue 700
			PrimaryLight: widget.Hex(0x63A4FF), // Blue 400
			PrimaryDark:  widget.Hex(0x004BA0), // Blue 900

			// Secondary colors - Material Teal
			Secondary:      widget.Hex(0x009688), // Teal 500
			SecondaryLight: widget.Hex(0x52C7B8), // Teal 300
			SecondaryDark:  widget.Hex(0x00675B), // Teal 800

			// Surface colors
			Background:     widget.Hex(0xFFFFFF), // White
			Surface:        widget.Hex(0xFAFAFA), // Gray 50
			SurfaceVariant: widget.Hex(0xF5F5F5), // Gray 100

			// Semantic colors
			Error:   widget.Hex(0xD32F2F), // Red 700
			Warning: widget.Hex(0xED6C02), // Orange 700
			Success: widget.Hex(0x2E7D32), // Green 700
			Info:    widget.Hex(0x0288D1), // Light Blue 700

			// On-colors (text/icons on corresponding backgrounds)
			OnPrimary:    widget.Hex(0xFFFFFF), // White on blue
			OnSecondary:  widget.Hex(0xFFFFFF), // White on teal
			OnBackground: widget.Hex(0x212121), // Gray 900 on white
			OnSurface:    widget.Hex(0x212121), // Gray 900 on surface
			OnError:      widget.Hex(0xFFFFFF), // White on red

			// UI element colors
			Divider: widget.RGBA(0, 0, 0, 0.12), // 12% black
			Outline: widget.RGBA(0, 0, 0, 0.23), // 23% black
			Shadow:  widget.RGBA(0, 0, 0, 0.20), // 20% black
		},
		Typography: DefaultTypography(),
		Spacing:    DefaultSpacing(),
		Shadows:    DefaultShadowsLight(),
		Radii:      DefaultRadii(),
		Extensions: make(map[string]any),
	}
}

// DefaultDark returns a pre-configured dark theme following Material 3 guidelines.
//
// The dark theme uses a dark gray background with light text and lighter
// primary colors for better contrast. It's suitable for low-light environments
// and can reduce eye strain and save battery on OLED displays.
//
// Color scheme:
//   - Primary: Light Blue (#64B5F6)
//   - Secondary: Light Teal (#4DB6AC)
//   - Background: Dark gray (#121212)
//   - Surface: Slightly lighter gray (#1E1E1E)
//   - Error: Light Red (#EF5350)
//
// Example:
//
//	theme := theme.DefaultDark()
//	backgroundColor := theme.Colors.Background
func DefaultDark() *Theme {
	return &Theme{
		Name: "Dark",
		Mode: ModeDark,
		Colors: ColorPalette{
			// Primary colors - Lighter blue for dark backgrounds
			Primary:      widget.Hex(0x64B5F6), // Blue 300
			PrimaryLight: widget.Hex(0x9BE7FF), // Blue 100
			PrimaryDark:  widget.Hex(0x2286C3), // Blue 600

			// Secondary colors - Lighter teal for dark backgrounds
			Secondary:      widget.Hex(0x4DB6AC), // Teal 300
			SecondaryLight: widget.Hex(0x82E9DE), // Teal 100
			SecondaryDark:  widget.Hex(0x00867D), // Teal 600

			// Surface colors - Dark grays
			Background:     widget.Hex(0x121212), // Material dark background
			Surface:        widget.Hex(0x1E1E1E), // Elevated surface
			SurfaceVariant: widget.Hex(0x2C2C2C), // Higher elevation

			// Semantic colors - Lighter versions for dark backgrounds
			Error:   widget.Hex(0xEF5350), // Red 400
			Warning: widget.Hex(0xFFA726), // Orange 400
			Success: widget.Hex(0x66BB6A), // Green 400
			Info:    widget.Hex(0x29B6F6), // Light Blue 400

			// On-colors (text/icons on corresponding backgrounds)
			OnPrimary:    widget.Hex(0x000000), // Black on light blue
			OnSecondary:  widget.Hex(0x000000), // Black on light teal
			OnBackground: widget.Hex(0xE0E0E0), // Gray 300 on dark
			OnSurface:    widget.Hex(0xE0E0E0), // Gray 300 on surface
			OnError:      widget.Hex(0x000000), // Black on light red

			// UI element colors
			Divider: widget.RGBA(1, 1, 1, 0.12), // 12% white
			Outline: widget.RGBA(1, 1, 1, 0.23), // 23% white
			Shadow:  widget.RGBA(0, 0, 0, 0.50), // 50% black (stronger for dark)
		},
		Typography: DefaultTypography(),
		Spacing:    DefaultSpacing(),
		Shadows:    DefaultShadowsDark(),
		Radii:      DefaultRadii(),
		Extensions: make(map[string]any),
	}
}

// DefaultHighContrast returns a high-contrast theme for accessibility.
//
// This theme uses maximum contrast ratios (black on white) for better
// readability by users with visual impairments. It follows WCAG AAA
// guidelines with contrast ratios of at least 7:1.
//
// Color scheme:
//   - Primary: Strong Blue (#0033CC)
//   - Background: White (#FFFFFF)
//   - Text: Black (#000000)
//   - Error: Strong Red (#CC0000)
//
// Example:
//
//	theme := theme.DefaultHighContrast()
func DefaultHighContrast() *Theme {
	return &Theme{
		Name: "High Contrast",
		Mode: ModeLight,
		Colors: ColorPalette{
			// High contrast primary colors
			Primary:      widget.Hex(0x0033CC), // Strong blue
			PrimaryLight: widget.Hex(0x3366FF), // Brighter blue
			PrimaryDark:  widget.Hex(0x002299), // Darker blue

			// High contrast secondary colors
			Secondary:      widget.Hex(0x006644), // Strong green
			SecondaryLight: widget.Hex(0x008855), // Brighter green
			SecondaryDark:  widget.Hex(0x004433), // Darker green

			// Maximum contrast surfaces
			Background:     widget.Hex(0xFFFFFF), // Pure white
			Surface:        widget.Hex(0xFFFFFF), // Pure white
			SurfaceVariant: widget.Hex(0xF0F0F0), // Very light gray

			// Strong semantic colors
			Error:   widget.Hex(0xCC0000), // Strong red
			Warning: widget.Hex(0xCC6600), // Strong orange
			Success: widget.Hex(0x006600), // Strong green
			Info:    widget.Hex(0x0066CC), // Strong blue

			// Maximum contrast text
			OnPrimary:    widget.Hex(0xFFFFFF), // White on blue
			OnSecondary:  widget.Hex(0xFFFFFF), // White on green
			OnBackground: widget.Hex(0x000000), // Black on white
			OnSurface:    widget.Hex(0x000000), // Black on white
			OnError:      widget.Hex(0xFFFFFF), // White on red

			// Clear UI element colors
			Divider: widget.Hex(0x000000),       // Black dividers
			Outline: widget.Hex(0x000000),       // Black outlines
			Shadow:  widget.RGBA(0, 0, 0, 0.40), // Strong shadow
		},
		Typography: DefaultTypography(),
		Spacing:    DefaultSpacing().Relaxed(), // More spacing for readability
		Shadows:    DefaultShadowsLight(),
		Radii:      SharpRadii(), // Sharp corners for clarity
		Extensions: make(map[string]any),
	}
}

// Purple returns a theme with purple primary colors.
//
// This is an example of a branded theme using Material Design purple palette.
//
// Color scheme:
//   - Primary: Purple (#7B1FA2)
//   - Secondary: Pink (#F06292)
//   - Background: White
func Purple() *Theme {
	t := DefaultLight()
	t.Name = "Purple"
	t.Colors.Primary = widget.Hex(0x7B1FA2)      // Purple 700
	t.Colors.PrimaryLight = widget.Hex(0xAE52D4) // Purple 400
	t.Colors.PrimaryDark = widget.Hex(0x4A0072)  // Purple 900
	t.Colors.Secondary = widget.Hex(0xF06292)    // Pink 300
	t.Colors.SecondaryLight = widget.Hex(0xFF94C2)
	t.Colors.SecondaryDark = widget.Hex(0xBA2D65)
	return t
}

// Green returns a theme with green primary colors.
//
// This is an example of a branded theme using Material Design green palette.
//
// Color scheme:
//   - Primary: Green (#388E3C)
//   - Secondary: Lime (#689F38)
//   - Background: White
func Green() *Theme {
	t := DefaultLight()
	t.Name = "Green"
	t.Colors.Primary = widget.Hex(0x388E3C)      // Green 700
	t.Colors.PrimaryLight = widget.Hex(0x6ABF69) // Green 400
	t.Colors.PrimaryDark = widget.Hex(0x00600F)  // Green 900
	t.Colors.Secondary = widget.Hex(0x689F38)    // Light Green 700
	t.Colors.SecondaryLight = widget.Hex(0x99D066)
	t.Colors.SecondaryDark = widget.Hex(0x387002)
	return t
}

// Orange returns a theme with orange primary colors.
//
// This is an example of a branded theme using Material Design orange palette.
//
// Color scheme:
//   - Primary: Orange (#F57C00)
//   - Secondary: Amber (#FFA000)
//   - Background: White
func Orange() *Theme {
	t := DefaultLight()
	t.Name = "Orange"
	t.Colors.Primary = widget.Hex(0xF57C00)      // Orange 700
	t.Colors.PrimaryLight = widget.Hex(0xFFAD42) // Orange 400
	t.Colors.PrimaryDark = widget.Hex(0xBB4D00)  // Orange 900
	t.Colors.Secondary = widget.Hex(0xFFA000)    // Amber 700
	t.Colors.SecondaryLight = widget.Hex(0xFFD149)
	t.Colors.SecondaryDark = widget.Hex(0xC67100)
	// Orange text needs dark text on light orange
	t.Colors.OnPrimary = widget.Hex(0x000000)
	t.Colors.OnSecondary = widget.Hex(0x000000)
	return t
}

// ForMode returns the appropriate default theme for the given mode.
//
// If mode is ModeSystem, this returns the light theme. Use your platform's
// system preference detection to determine the actual mode and call this
// with ModeLight or ModeDark.
func ForMode(mode ThemeMode) *Theme {
	switch mode {
	case ModeDark:
		return DefaultDark()
	default:
		return DefaultLight()
	}
}
