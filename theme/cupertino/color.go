package cupertino

import "github.com/gogpu/ui/widget"

// ColorScheme holds the Apple HIG color roles for the Cupertino theme.
//
// Apple uses a set of system-defined adaptive colors that automatically
// adjust between light and dark modes. This struct provides both semantic
// colors and the full set of system colors.
type ColorScheme struct {
	// Accent is the primary tint color (default: System Blue #007AFF).
	Accent widget.Color

	// OnAccent is the text/icon color on accent backgrounds.
	OnAccent widget.Color

	// Label is the primary text color.
	Label widget.Color

	// SecondaryLabel is the color for secondary text.
	SecondaryLabel widget.Color

	// TertiaryLabel is the color for tertiary text (placeholders).
	TertiaryLabel widget.Color

	// QuaternaryLabel is the color for quaternary text (disabled).
	QuaternaryLabel widget.Color

	// SystemBackground is the primary background color.
	SystemBackground widget.Color

	// SecondarySystemBackground is the grouped/secondary background.
	SecondarySystemBackground widget.Color

	// TertiarySystemBackground is the elevated/tertiary background.
	TertiarySystemBackground widget.Color

	// Separator is the color for thin separators.
	Separator widget.Color

	// OpaqueSeparator is the color for opaque separators.
	OpaqueSeparator widget.Color

	// SystemFill is the color for thin overlays on interactive elements.
	SystemFill widget.Color

	// SecondarySystemFill is the color for medium-thick overlays.
	SecondarySystemFill widget.Color

	// TertiarySystemFill is the color for thick overlays.
	TertiarySystemFill widget.Color

	// SystemRed is the destructive action color.
	SystemRed widget.Color

	// SystemGreen is the success/confirmation color.
	SystemGreen widget.Color

	// SystemOrange is the warning color.
	SystemOrange widget.Color
}

// lightColors returns the Apple HIG light mode color scheme with the given accent.
func lightColors(accent widget.Color) ColorScheme {
	return ColorScheme{
		Accent:                    accent,
		OnAccent:                  widget.ColorWhite,
		Label:                     widget.RGBA(0.0, 0.0, 0.0, 1.0),
		SecondaryLabel:            widget.RGBA(0.235, 0.235, 0.263, 0.6),
		TertiaryLabel:             widget.RGBA(0.235, 0.235, 0.263, 0.3),
		QuaternaryLabel:           widget.RGBA(0.235, 0.235, 0.263, 0.18),
		SystemBackground:          widget.ColorWhite,
		SecondarySystemBackground: widget.Hex(0xF2F2F7),
		TertiarySystemBackground:  widget.ColorWhite,
		Separator:                 widget.RGBA(0.235, 0.235, 0.263, 0.29),
		OpaqueSeparator:           widget.Hex(0xC6C6C8),
		SystemFill:                widget.RGBA(0.47, 0.47, 0.50, 0.2),
		SecondarySystemFill:       widget.RGBA(0.47, 0.47, 0.50, 0.16),
		TertiarySystemFill:        widget.RGBA(0.46, 0.46, 0.50, 0.12),
		SystemRed:                 widget.Hex(0xFF3B30),
		SystemGreen:               widget.Hex(0x34C759),
		SystemOrange:              widget.Hex(0xFF9500),
	}
}

// darkColors returns the Apple HIG dark mode color scheme with the given accent.
func darkColors(accent widget.Color) ColorScheme {
	return ColorScheme{
		Accent:                    accent,
		OnAccent:                  widget.ColorWhite,
		Label:                     widget.ColorWhite,
		SecondaryLabel:            widget.RGBA(0.922, 0.922, 0.961, 0.6),
		TertiaryLabel:             widget.RGBA(0.922, 0.922, 0.961, 0.3),
		QuaternaryLabel:           widget.RGBA(0.922, 0.922, 0.961, 0.18),
		SystemBackground:          widget.Hex(0x000000),
		SecondarySystemBackground: widget.Hex(0x1C1C1E),
		TertiarySystemBackground:  widget.Hex(0x2C2C2E),
		Separator:                 widget.RGBA(0.329, 0.329, 0.345, 0.6),
		OpaqueSeparator:           widget.Hex(0x38383A),
		SystemFill:                widget.RGBA(0.47, 0.47, 0.50, 0.36),
		SecondarySystemFill:       widget.RGBA(0.47, 0.47, 0.50, 0.32),
		TertiarySystemFill:        widget.RGBA(0.46, 0.46, 0.50, 0.24),
		SystemRed:                 widget.Hex(0xFF453A),
		SystemGreen:               widget.Hex(0x30D158),
		SystemOrange:              widget.Hex(0xFF9F0A),
	}
}

// Apple system colors (light mode).
var (
	systemBlue = widget.Hex(0x007AFF)
)
