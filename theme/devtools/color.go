package devtools

import "github.com/gogpu/ui/widget"

// ColorScheme holds all DevTools design system color roles.
//
// The scheme is organized around a neutral gray scale with a blue accent,
// following the JetBrains Int UI color system. Dark mode uses deep grays
// (#1E1F22 through #393B40) for surfaces with light text (#DFE1E5), while
// light mode inverts to white/light-gray surfaces with dark text.
//
// Both modes share the same blue accent (#3574F0) and a dark header toolbar,
// matching JetBrains IDEs where the main toolbar is always dark.
type ColorScheme struct {
	// Surface group: neutral backgrounds from darkest to lightest.
	Background         widget.Color // deepest background (editor area)
	Surface            widget.Color // panel/surface background
	SurfaceElevated    widget.Color // elevated surface, separator
	OnSurface          widget.Color // primary text/foreground
	OnSurfaceSecondary widget.Color // secondary text (dimmer)
	OnSurfaceDisabled  widget.Color // disabled/muted text

	// Accent/Primary: brand accent color and interactive states.
	Primary      widget.Color // primary accent (buttons, links, focus ring)
	PrimaryHover widget.Color // hover state
	PrimaryPress widget.Color // pressed state
	OnPrimary    widget.Color // text/icon on primary-colored surfaces

	// Selection: text/item selection background.
	Selection widget.Color

	// Borders: varying strengths for visual hierarchy.
	Border       widget.Color // default border (subtle)
	BorderStrong widget.Color // stronger border (emphasis)
	BorderFocus  widget.Color // focused element border (accent)

	// Input/Control: fills for interactive controls.
	InputBackground widget.Color // text input background
	ControlFill     widget.Color // default control fill
	ControlHover    widget.Color // control hover fill

	// Header/Toolbar: always dark in both light and dark modes.
	HeaderBackground widget.Color // dark toolbar background
	HeaderForeground widget.Color // toolbar text/icons

	// Semantic: status-indicating colors.
	Error           widget.Color // error border/icon
	ErrorBackground widget.Color // error surface tint
	Warning         widget.Color // warning icon
	Success         widget.Color // success icon
	Info            widget.Color // info (same as primary)

	// Overlay: backdrops and shadows.
	Backdrop widget.Color // modal backdrop
	Shadow   widget.Color // drop shadow

	// Scrollbar: thumb and track colors.
	ScrollbarThumb      widget.Color // scrollbar thumb at rest
	ScrollbarThumbHover widget.Color // scrollbar thumb hovered
	ScrollbarTrack      widget.Color // scrollbar track (typically transparent)
}

// DarkScheme returns the DevTools dark color scheme.
//
// This is the primary color scheme, optimized for long coding sessions.
// Colors are extracted from JetBrains Int UI dark theme (expUI/dark.theme.json).
func DarkScheme() ColorScheme {
	return ColorScheme{
		// Surface group (Gray1-Gray3)
		Background:         widget.Hex(0x1E1F22), // Gray1
		Surface:            widget.Hex(0x2B2D30), // Gray2
		SurfaceElevated:    widget.Hex(0x393B40), // Gray3
		OnSurface:          widget.Hex(0xDFE1E5), // Gray12
		OnSurfaceSecondary: widget.Hex(0x9DA0A8), // Gray9
		OnSurfaceDisabled:  widget.Hex(0x6F737A), // Gray7

		// Accent (Blue6-Blue8)
		Primary:      widget.Hex(0x3574F0), // Blue6
		PrimaryHover: widget.Hex(0x467FF2), // Blue7
		PrimaryPress: widget.Hex(0x548AF7), // Blue8
		OnPrimary:    widget.ColorWhite,

		// Selection (Blue2)
		Selection: widget.Hex(0x2E436E),

		// Borders (Gray3, Gray5, Blue6)
		Border:       widget.Hex(0x393B40), // Gray3
		BorderStrong: widget.Hex(0x4E5157), // Gray5
		BorderFocus:  widget.Hex(0x3574F0), // Blue6

		// Input/Control (Gray1, Gray3, Gray4)
		InputBackground: widget.Hex(0x1E1F22), // Gray1
		ControlFill:     widget.Hex(0x393B40), // Gray3
		ControlHover:    widget.Hex(0x43454A), // Gray4

		// Header (always dark)
		HeaderBackground: widget.Hex(0x27282E),
		HeaderForeground: widget.Hex(0xDFE1E5), // Gray12

		// Semantic
		Error:           widget.Hex(0xDB5C5C), // Red7
		ErrorBackground: widget.Hex(0x402929),
		Warning:         widget.Hex(0xC69026), // Yellow6
		Success:         widget.Hex(0x57965C), // Green6
		Info:            widget.Hex(0x3574F0), // Blue6

		// Overlay
		Backdrop: widget.RGBA(0, 0, 0, 0.50),
		Shadow:   widget.RGBA(0, 0, 0, 0.50),

		// Scrollbar (Gray5, Gray7)
		ScrollbarThumb:      widget.Hex(0x4E5157), // Gray5
		ScrollbarThumbHover: widget.Hex(0x6F737A), // Gray7
		ScrollbarTrack:      widget.ColorTransparent,
	}
}

// LightScheme returns the DevTools light color scheme.
//
// Light mode uses white/light-gray surfaces with dark text. Notably, the
// header toolbar remains dark (#27282E), matching JetBrains IDE behavior
// where the main toolbar is always dark even in light themes.
func LightScheme() ColorScheme {
	return ColorScheme{
		// Surface group (light grays)
		Background:         widget.Hex(0xFFFFFF), // Gray1 light
		Surface:            widget.Hex(0xF7F8FA), // Gray2 light
		SurfaceElevated:    widget.Hex(0xEBECF0), // Gray3 light
		OnSurface:          widget.Hex(0x000000), // Gray12 light
		OnSurfaceSecondary: widget.Hex(0x6C707E), // Gray9 light
		OnSurfaceDisabled:  widget.Hex(0x9DA0A8), // muted

		// Accent (same blue)
		Primary:      widget.Hex(0x3574F0), // Blue6
		PrimaryHover: widget.Hex(0x2A62D4), // slightly darker for light bg
		PrimaryPress: widget.Hex(0x2152B8), // darker pressed
		OnPrimary:    widget.ColorWhite,

		// Selection
		Selection: widget.Hex(0xC4D5F6), // light blue selection

		// Borders (light grays)
		Border:       widget.Hex(0xDFE1E5), // Gray4 light
		BorderStrong: widget.Hex(0xC9CCD6), // Gray5 light
		BorderFocus:  widget.Hex(0x3574F0), // Blue6

		// Input/Control
		InputBackground: widget.Hex(0xFFFFFF),
		ControlFill:     widget.Hex(0xEBECF0), // Gray3 light
		ControlHover:    widget.Hex(0xDFE1E5), // Gray4 light

		// Header (dark even in light mode!)
		HeaderBackground: widget.Hex(0x27282E),
		HeaderForeground: widget.Hex(0xDFE1E5),

		// Semantic
		Error:           widget.Hex(0xDB5C5C),
		ErrorBackground: widget.Hex(0xFCE8E8),
		Warning:         widget.Hex(0xC69026),
		Success:         widget.Hex(0x57965C),
		Info:            widget.Hex(0x3574F0),

		// Overlay
		Backdrop: widget.RGBA(0, 0, 0, 0.30),
		Shadow:   widget.RGBA(0, 0, 0, 0.14),

		// Scrollbar
		ScrollbarThumb:      widget.Hex(0xC9CCD6),
		ScrollbarThumbHover: widget.Hex(0x9DA0A8),
		ScrollbarTrack:      widget.ColorTransparent,
	}
}
