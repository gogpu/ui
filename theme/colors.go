package theme

import "github.com/gogpu/ui/widget"

// ColorPalette defines the semantic colors for a theme.
//
// ColorPalette follows Material 3 design guidelines with semantic naming
// that describes the purpose of each color rather than its appearance.
// This allows themes to be switched between light and dark modes while
// maintaining consistent meaning.
//
// # Color Categories
//
// Primary colors:
//   - Primary: The brand's main color, used for key UI elements
//   - PrimaryLight: A lighter variant for hover states or containers
//   - PrimaryDark: A darker variant for pressed states or emphasis
//
// Secondary colors:
//   - Secondary: An accent color for less prominent elements
//   - SecondaryLight: A lighter secondary variant
//   - SecondaryDark: A darker secondary variant
//
// Surface colors:
//   - Background: The main window/page background
//   - Surface: Card, panel, and dialog surfaces
//   - SurfaceVariant: Alternative surface for visual distinction
//
// Semantic colors:
//   - Error: Error states and destructive actions
//   - Warning: Warning states and cautionary actions
//   - Success: Success states and confirmations
//   - Info: Informational states and tips
//
// On-colors (text/icon colors for use ON the corresponding background):
//   - OnPrimary: Text/icons on Primary color
//   - OnSecondary: Text/icons on Secondary color
//   - OnBackground: Text/icons on Background color
//   - OnSurface: Text/icons on Surface color
//   - OnError: Text/icons on Error color
//
// UI element colors:
//   - Divider: Lines separating content
//   - Outline: Borders and outlines
//   - Shadow: Shadow color (typically semi-transparent black)
type ColorPalette struct {
	// Primary brand colors
	Primary      widget.Color
	PrimaryLight widget.Color
	PrimaryDark  widget.Color

	// Secondary accent colors
	Secondary      widget.Color
	SecondaryLight widget.Color
	SecondaryDark  widget.Color

	// Surface colors
	Background     widget.Color
	Surface        widget.Color
	SurfaceVariant widget.Color

	// Semantic colors
	Error   widget.Color
	Warning widget.Color
	Success widget.Color
	Info    widget.Color

	// On-colors (text/icons on corresponding backgrounds)
	OnPrimary    widget.Color
	OnSecondary  widget.Color
	OnBackground widget.Color
	OnSurface    widget.Color
	OnError      widget.Color

	// UI element colors
	Divider widget.Color
	Outline widget.Color
	Shadow  widget.Color
}

// WithAlpha returns a copy of the palette with all colors adjusted to the given alpha.
//
// This is useful for creating disabled or overlay states.
func (p *ColorPalette) WithAlpha(alpha float32) ColorPalette {
	return ColorPalette{
		Primary:        p.Primary.WithAlpha(alpha),
		PrimaryLight:   p.PrimaryLight.WithAlpha(alpha),
		PrimaryDark:    p.PrimaryDark.WithAlpha(alpha),
		Secondary:      p.Secondary.WithAlpha(alpha),
		SecondaryLight: p.SecondaryLight.WithAlpha(alpha),
		SecondaryDark:  p.SecondaryDark.WithAlpha(alpha),
		Background:     p.Background.WithAlpha(alpha),
		Surface:        p.Surface.WithAlpha(alpha),
		SurfaceVariant: p.SurfaceVariant.WithAlpha(alpha),
		Error:          p.Error.WithAlpha(alpha),
		Warning:        p.Warning.WithAlpha(alpha),
		Success:        p.Success.WithAlpha(alpha),
		Info:           p.Info.WithAlpha(alpha),
		OnPrimary:      p.OnPrimary.WithAlpha(alpha),
		OnSecondary:    p.OnSecondary.WithAlpha(alpha),
		OnBackground:   p.OnBackground.WithAlpha(alpha),
		OnSurface:      p.OnSurface.WithAlpha(alpha),
		OnError:        p.OnError.WithAlpha(alpha),
		Divider:        p.Divider.WithAlpha(alpha),
		Outline:        p.Outline.WithAlpha(alpha),
		Shadow:         p.Shadow.WithAlpha(alpha),
	}
}

// Lerp returns a color palette linearly interpolated between p and other.
//
// t=0 returns p, t=1 returns other. This is useful for animating
// between themes.
func (p *ColorPalette) Lerp(other *ColorPalette, t float32) ColorPalette {
	return ColorPalette{
		Primary:        p.Primary.Lerp(other.Primary, t),
		PrimaryLight:   p.PrimaryLight.Lerp(other.PrimaryLight, t),
		PrimaryDark:    p.PrimaryDark.Lerp(other.PrimaryDark, t),
		Secondary:      p.Secondary.Lerp(other.Secondary, t),
		SecondaryLight: p.SecondaryLight.Lerp(other.SecondaryLight, t),
		SecondaryDark:  p.SecondaryDark.Lerp(other.SecondaryDark, t),
		Background:     p.Background.Lerp(other.Background, t),
		Surface:        p.Surface.Lerp(other.Surface, t),
		SurfaceVariant: p.SurfaceVariant.Lerp(other.SurfaceVariant, t),
		Error:          p.Error.Lerp(other.Error, t),
		Warning:        p.Warning.Lerp(other.Warning, t),
		Success:        p.Success.Lerp(other.Success, t),
		Info:           p.Info.Lerp(other.Info, t),
		OnPrimary:      p.OnPrimary.Lerp(other.OnPrimary, t),
		OnSecondary:    p.OnSecondary.Lerp(other.OnSecondary, t),
		OnBackground:   p.OnBackground.Lerp(other.OnBackground, t),
		OnSurface:      p.OnSurface.Lerp(other.OnSurface, t),
		OnError:        p.OnError.Lerp(other.OnError, t),
		Divider:        p.Divider.Lerp(other.Divider, t),
		Outline:        p.Outline.Lerp(other.Outline, t),
		Shadow:         p.Shadow.Lerp(other.Shadow, t),
	}
}

// ContrastColor returns an appropriate text color for the given background.
//
// This uses a simple luminance calculation to determine whether white or
// black text would provide better contrast on the given background color.
//
// The onLight color is returned for light backgrounds (luminance >= 0.5),
// and onDark is returned for dark backgrounds.
func ContrastColor(background, onLight, onDark widget.Color) widget.Color {
	// Calculate relative luminance using sRGB coefficients
	// This is a simplified version; full WCAG compliance would need gamma correction
	luminance := 0.299*background.R + 0.587*background.G + 0.114*background.B

	if luminance >= 0.5 {
		return onLight
	}
	return onDark
}

// Lighten returns a lighter version of the color by the given amount.
//
// The amount should be in the range [0, 1], where 0 returns the original
// color and 1 returns white.
func Lighten(c widget.Color, amount float32) widget.Color {
	return c.Lerp(widget.ColorWhite, clamp01(amount))
}

// Darken returns a darker version of the color by the given amount.
//
// The amount should be in the range [0, 1], where 0 returns the original
// color and 1 returns black.
func Darken(c widget.Color, amount float32) widget.Color {
	return c.Lerp(widget.ColorBlack, clamp01(amount))
}

// WithOpacity returns a copy of the color with the given opacity applied.
//
// This multiplies the existing alpha by the opacity factor.
// opacity should be in the range [0, 1].
func WithOpacity(c widget.Color, opacity float32) widget.Color {
	return widget.Color{
		R: c.R,
		G: c.G,
		B: c.B,
		A: c.A * clamp01(opacity),
	}
}

// clamp01 clamps a float32 value to the range [0, 1].
func clamp01(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
