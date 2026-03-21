package icon

import "github.com/gogpu/ui/widget"

// Color key constants used by built-in multi-color icons.
const (
	// KeyPrimary is the default foreground color key (outlines, text).
	KeyPrimary = "primary"

	// KeyAccent is the primary accent color key (highlights, indicators).
	KeyAccent = "accent"

	// KeySecondary is a secondary/muted foreground color key.
	KeySecondary = "secondary"

	// KeySuccess is the success/positive color key (green tones).
	KeySuccess = "success"

	// KeyError is the error/negative color key (red tones).
	KeyError = "error"

	// KeyWarning is the warning/caution color key (yellow/orange tones).
	KeyWarning = "warning"

	// KeyGo is the Go language brand color key (cyan).
	KeyGo = "go"

	// KeyJSON is the JSON file type color key (gold).
	KeyJSON = "json"

	// KeyYAML is the YAML file type color key (green).
	KeyYAML = "yaml"

	// KeyRust is the Rust language brand color key (orange).
	KeyRust = "rust"

	// KeyPython is the Python language brand color key (blue).
	KeyPython = "python"

	// KeyMarkdown is the Markdown file type color key (gray).
	KeyMarkdown = "markdown"
)

// DefaultDarkPalette returns a palette suitable for dark themes.
//
// The color values are inspired by JetBrains IDE dark themes and provide
// good contrast on dark backgrounds.
func DefaultDarkPalette() Palette {
	return Palette{
		KeyPrimary:   widget.Hex(0xDFE1E5), // Gray12 — default foreground
		KeyAccent:    widget.Hex(0x3574F0), // Blue6 — primary accent
		KeySecondary: widget.Hex(0x9DA0A8), // Gray9
		KeySuccess:   widget.Hex(0x57965C), // Green6
		KeyError:     widget.Hex(0xDB5C5C), // Red7
		KeyWarning:   widget.Hex(0xC69026), // Yellow6
		KeyGo:        widget.Hex(0x00ADD8), // Go cyan
		KeyJSON:      widget.Hex(0xEDA200), // Gold
		KeyYAML:      widget.Hex(0x59A869), // Green
		KeyRust:      widget.Hex(0xDEA584), // Rust orange
		KeyPython:    widget.Hex(0x3776AB), // Python blue
		KeyMarkdown:  widget.Hex(0x9DA0A8), // Gray
	}
}

// DefaultLightPalette returns a palette suitable for light themes.
//
// The color values provide good contrast on light backgrounds with slightly
// darker tones than the dark palette counterparts.
func DefaultLightPalette() Palette {
	return Palette{
		KeyPrimary:   widget.Hex(0x3D3F43), // Dark gray — default foreground
		KeyAccent:    widget.Hex(0x2B5FC3), // Blue — primary accent
		KeySecondary: widget.Hex(0x6E7076), // Medium gray
		KeySuccess:   widget.Hex(0x3B7A40), // Dark green
		KeyError:     widget.Hex(0xC4314B), // Dark red
		KeyWarning:   widget.Hex(0xA67A1A), // Dark gold
		KeyGo:        widget.Hex(0x007D9C), // Go dark cyan
		KeyJSON:      widget.Hex(0xB88200), // Dark gold
		KeyYAML:      widget.Hex(0x3B7A40), // Dark green
		KeyRust:      widget.Hex(0xB07858), // Rust dark orange
		KeyPython:    widget.Hex(0x2B5C8A), // Python dark blue
		KeyMarkdown:  widget.Hex(0x6E7076), // Medium gray
	}
}
