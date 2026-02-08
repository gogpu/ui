package material3

import "github.com/gogpu/ui/widget"

// ColorScheme holds all Material 3 color roles.
//
// Material 3 defines a comprehensive set of color roles organized in
// primary/secondary/tertiary/error groups, each with container variants
// and corresponding "on" colors for content placed on top. Surface
// colors provide a neutral backdrop with multiple elevation levels.
type ColorScheme struct {
	// Primary group: key brand color and its variants.
	Primary            widget.Color
	OnPrimary          widget.Color
	PrimaryContainer   widget.Color
	OnPrimaryContainer widget.Color

	// Secondary group: accent color for less prominent elements.
	Secondary            widget.Color
	OnSecondary          widget.Color
	SecondaryContainer   widget.Color
	OnSecondaryContainer widget.Color

	// Tertiary group: complementary color for contrast and balance.
	Tertiary            widget.Color
	OnTertiary          widget.Color
	TertiaryContainer   widget.Color
	OnTertiaryContainer widget.Color

	// Error group: error states and destructive actions.
	Error            widget.Color
	OnError          widget.Color
	ErrorContainer   widget.Color
	OnErrorContainer widget.Color

	// Surface group: neutral backgrounds at various elevation levels.
	Surface                 widget.Color
	OnSurface               widget.Color
	SurfaceVariant          widget.Color
	OnSurfaceVariant        widget.Color
	SurfaceContainerLowest  widget.Color
	SurfaceContainerLow     widget.Color
	SurfaceContainer        widget.Color
	SurfaceContainerHigh    widget.Color
	SurfaceContainerHighest widget.Color

	// Background colors.
	Background   widget.Color
	OnBackground widget.Color

	// Outline colors for borders and dividers.
	Outline        widget.Color
	OutlineVariant widget.Color

	// Inverse colors for elements that appear on inverse surfaces.
	InverseSurface   widget.Color
	InverseOnSurface widget.Color
	InversePrimary   widget.Color
}

// Light generates a light color scheme from a seed color.
//
// The seed color is used to derive a full set of harmonious colors
// following Material 3 tonal mapping rules for light themes.
func Light(seed widget.Color) ColorScheme {
	p := newCorePalette(seed)
	return lightFromPalette(p)
}

// Dark generates a dark color scheme from a seed color.
//
// The seed color is used to derive a full set of harmonious colors
// following Material 3 tonal mapping rules for dark themes.
func Dark(seed widget.Color) ColorScheme {
	p := newCorePalette(seed)
	return darkFromPalette(p)
}

// lightFromPalette maps a core palette to a light color scheme.
//
// Tone mappings follow the Material 3 specification:
// https://m3.material.io/styles/color/the-color-system/color-roles
func lightFromPalette(p corePalette) ColorScheme {
	return ColorScheme{
		Primary:            p.Primary.tone(40),
		OnPrimary:          p.Primary.tone(100),
		PrimaryContainer:   p.Primary.tone(90),
		OnPrimaryContainer: p.Primary.tone(10),

		Secondary:            p.Secondary.tone(40),
		OnSecondary:          p.Secondary.tone(100),
		SecondaryContainer:   p.Secondary.tone(90),
		OnSecondaryContainer: p.Secondary.tone(10),

		Tertiary:            p.Tertiary.tone(40),
		OnTertiary:          p.Tertiary.tone(100),
		TertiaryContainer:   p.Tertiary.tone(90),
		OnTertiaryContainer: p.Tertiary.tone(10),

		Error:            p.Error.tone(40),
		OnError:          p.Error.tone(100),
		ErrorContainer:   p.Error.tone(90),
		OnErrorContainer: p.Error.tone(10),

		Surface:                 p.Neutral.tone(99),
		OnSurface:               p.Neutral.tone(10),
		SurfaceVariant:          p.Neutral.tone(90),
		OnSurfaceVariant:        p.Neutral.tone(30),
		SurfaceContainerLowest:  p.Neutral.tone(100),
		SurfaceContainerLow:     p.Neutral.tone(96),
		SurfaceContainer:        p.Neutral.tone(94),
		SurfaceContainerHigh:    p.Neutral.tone(92),
		SurfaceContainerHighest: p.Neutral.tone(90),

		Background:   p.Neutral.tone(99),
		OnBackground: p.Neutral.tone(10),

		Outline:        p.Neutral.tone(50),
		OutlineVariant: p.Neutral.tone(80),

		InverseSurface:   p.Neutral.tone(20),
		InverseOnSurface: p.Neutral.tone(95),
		InversePrimary:   p.Primary.tone(80),
	}
}

// darkFromPalette maps a core palette to a dark color scheme.
//
// Dark schemes use higher tones for primary colors and lower tones
// for surfaces, providing adequate contrast on dark backgrounds.
func darkFromPalette(p corePalette) ColorScheme {
	return ColorScheme{
		Primary:            p.Primary.tone(80),
		OnPrimary:          p.Primary.tone(20),
		PrimaryContainer:   p.Primary.tone(30),
		OnPrimaryContainer: p.Primary.tone(90),

		Secondary:            p.Secondary.tone(80),
		OnSecondary:          p.Secondary.tone(20),
		SecondaryContainer:   p.Secondary.tone(30),
		OnSecondaryContainer: p.Secondary.tone(90),

		Tertiary:            p.Tertiary.tone(80),
		OnTertiary:          p.Tertiary.tone(20),
		TertiaryContainer:   p.Tertiary.tone(30),
		OnTertiaryContainer: p.Tertiary.tone(90),

		Error:            p.Error.tone(80),
		OnError:          p.Error.tone(20),
		ErrorContainer:   p.Error.tone(30),
		OnErrorContainer: p.Error.tone(90),

		Surface:                 p.Neutral.tone(6),
		OnSurface:               p.Neutral.tone(90),
		SurfaceVariant:          p.Neutral.tone(30),
		OnSurfaceVariant:        p.Neutral.tone(80),
		SurfaceContainerLowest:  p.Neutral.tone(4),
		SurfaceContainerLow:     p.Neutral.tone(10),
		SurfaceContainer:        p.Neutral.tone(12),
		SurfaceContainerHigh:    p.Neutral.tone(17),
		SurfaceContainerHighest: p.Neutral.tone(22),

		Background:   p.Neutral.tone(6),
		OnBackground: p.Neutral.tone(90),

		Outline:        p.Neutral.tone(60),
		OutlineVariant: p.Neutral.tone(30),

		InverseSurface:   p.Neutral.tone(90),
		InverseOnSurface: p.Neutral.tone(20),
		InversePrimary:   p.Primary.tone(40),
	}
}
