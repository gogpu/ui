package fluent

import "github.com/gogpu/ui/widget"

// ColorScheme holds all Fluent Design color roles.
//
// Fluent Design organizes colors around an accent color with neutral
// surface tones. Light mode uses white/gray backgrounds with dark text,
// while dark mode uses dark gray backgrounds (#1F1F1F, #2D2D2D) with
// light text.
type ColorScheme struct {
	// Accent group: brand color and variants.
	Accent      widget.Color
	AccentHover widget.Color
	AccentPress widget.Color
	OnAccent    widget.Color
	AccentLight widget.Color // lighter accent for subtle backgrounds
	AccentDark  widget.Color // darker accent for emphasis

	// Surface group: neutral backgrounds.
	Surface          widget.Color
	SurfaceSecondary widget.Color // slightly different surface (cards)
	SurfaceTertiary  widget.Color // input field backgrounds
	OnSurface        widget.Color // primary text
	OnSurfaceSecond  widget.Color // secondary text (dimmer)

	// Stroke group: borders and outlines.
	StrokeDefault widget.Color // control borders
	StrokeFocus   widget.Color // focused control borders
	StrokeDisable widget.Color // disabled borders

	// Fill group: control backgrounds.
	FillDefault  widget.Color // standard control fill
	FillSecond   widget.Color // secondary fill (hover)
	FillTertiary widget.Color // tertiary fill (pressed)
	FillDisable  widget.Color // disabled fill

	// Semantic colors.
	Error   widget.Color
	Warning widget.Color
	Success widget.Color

	// Overlay and shadow.
	Backdrop widget.Color
	Shadow   widget.Color
}

// LightScheme generates a Fluent Design light color scheme from an accent color.
func LightScheme(accent widget.Color) ColorScheme {
	return ColorScheme{
		Accent:      accent,
		AccentHover: darken(accent, 0.08),
		AccentPress: darken(accent, 0.16),
		OnAccent:    widget.ColorWhite,
		AccentLight: lighten(accent, 0.85),
		AccentDark:  darken(accent, 0.25),

		Surface:          widget.Hex(0xFFFFFF),
		SurfaceSecondary: widget.Hex(0xF3F3F3),
		SurfaceTertiary:  widget.Hex(0xFAFAFA),
		OnSurface:        widget.Hex(0x1A1A1A),
		OnSurfaceSecond:  widget.Hex(0x616161),

		StrokeDefault: widget.RGBA(0, 0, 0, 0.14),
		StrokeFocus:   accent,
		StrokeDisable: widget.RGBA(0, 0, 0, 0.06),

		FillDefault:  widget.RGBA(0, 0, 0, 0.04),
		FillSecond:   widget.RGBA(0, 0, 0, 0.06),
		FillTertiary: widget.RGBA(0, 0, 0, 0.03),
		FillDisable:  widget.RGBA(0, 0, 0, 0.04),

		Error:   widget.Hex(0xC42B1C),
		Warning: widget.Hex(0x9D5D00),
		Success: widget.Hex(0x0F7B0F),

		Backdrop: widget.RGBA(0, 0, 0, 0.30),
		Shadow:   widget.RGBA(0, 0, 0, 0.14),
	}
}

// DarkScheme generates a Fluent Design dark color scheme from an accent color.
func DarkScheme(accent widget.Color) ColorScheme {
	// In dark mode, accent is lighter for contrast on dark backgrounds.
	lightAccent := lighten(accent, 0.25)
	return ColorScheme{
		Accent:      lightAccent,
		AccentHover: lighten(lightAccent, 0.08),
		AccentPress: darken(lightAccent, 0.08),
		OnAccent:    widget.Hex(0x1A1A1A),
		AccentLight: accent.WithAlpha(0.15),
		AccentDark:  accent,

		Surface:          widget.Hex(0x1F1F1F),
		SurfaceSecondary: widget.Hex(0x2D2D2D),
		SurfaceTertiary:  widget.Hex(0x282828),
		OnSurface:        widget.Hex(0xE4E4E4),
		OnSurfaceSecond:  widget.Hex(0x9E9E9E),

		StrokeDefault: widget.RGBA(1, 1, 1, 0.10),
		StrokeFocus:   lightAccent,
		StrokeDisable: widget.RGBA(1, 1, 1, 0.06),

		FillDefault:  widget.RGBA(1, 1, 1, 0.06),
		FillSecond:   widget.RGBA(1, 1, 1, 0.08),
		FillTertiary: widget.RGBA(1, 1, 1, 0.04),
		FillDisable:  widget.RGBA(1, 1, 1, 0.04),

		Error:   widget.Hex(0xFF99A4),
		Warning: widget.Hex(0xFCE100),
		Success: widget.Hex(0x6CCB5F),

		Backdrop: widget.RGBA(0, 0, 0, 0.50),
		Shadow:   widget.RGBA(0, 0, 0, 0.36),
	}
}

// lighten blends a color toward white by the given amount (0..1).
func lighten(c widget.Color, amount float32) widget.Color {
	return c.Lerp(widget.ColorWhite, clamp01(amount))
}

// darken blends a color toward black by the given amount (0..1).
func darken(c widget.Color, amount float32) widget.Color {
	return c.Lerp(widget.ColorBlack, clamp01(amount))
}

// clamp01 clamps a float32 value to [0, 1].
func clamp01(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
