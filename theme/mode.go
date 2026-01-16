package theme

// String constants for mode names.
const (
	unknownStr = "Unknown"
	modeLight  = "Light"
	modeDark   = "Dark"
	modeSystem = "System"
)

// ThemeMode represents the color scheme mode for a theme.
//
// ThemeMode determines whether a theme uses light colors, dark colors,
// or follows the operating system's preference.
type ThemeMode uint8

const (
	// ModeLight indicates a light color scheme with dark text on light backgrounds.
	//
	// Light mode is typically used in well-lit environments and is often
	// the default for daytime use.
	ModeLight ThemeMode = iota

	// ModeDark indicates a dark color scheme with light text on dark backgrounds.
	//
	// Dark mode reduces eye strain in low-light environments and can
	// save battery on OLED displays.
	ModeDark

	// ModeSystem indicates the theme should follow the operating system's
	// color scheme preference.
	//
	// When this mode is active, the platform integration layer should detect
	// the OS preference and apply the appropriate light or dark theme.
	// This package does not perform the actual detection; it only provides
	// the mode value.
	ModeSystem
)

// String returns a human-readable name for the theme mode.
func (m ThemeMode) String() string {
	switch m {
	case ModeLight:
		return modeLight
	case ModeDark:
		return modeDark
	case ModeSystem:
		return modeSystem
	default:
		return unknownStr
	}
}

// IsLight returns true if this mode represents a light color scheme.
//
// Note: This returns false for ModeSystem, as the actual scheme depends
// on the operating system preference which must be determined elsewhere.
func (m ThemeMode) IsLight() bool {
	return m == ModeLight
}

// IsDark returns true if this mode represents a dark color scheme.
//
// Note: This returns false for ModeSystem, as the actual scheme depends
// on the operating system preference which must be determined elsewhere.
func (m ThemeMode) IsDark() bool {
	return m == ModeDark
}

// IsSystem returns true if this mode follows the system preference.
func (m ThemeMode) IsSystem() bool {
	return m == ModeSystem
}

// ResolvedMode returns the effective mode when system preference is known.
//
// If the mode is ModeSystem, it returns lightIfSystem when preferLight is true,
// otherwise ModeDark. For explicit modes (ModeLight, ModeDark), it returns
// the mode unchanged.
//
// Example:
//
//	mode := theme.ModeSystem
//	effectiveMode := mode.ResolvedMode(osPrefersDarkMode)
//	// effectiveMode will be ModeDark if osPrefersDarkMode is true
func (m ThemeMode) ResolvedMode(preferLight bool) ThemeMode {
	if m != ModeSystem {
		return m
	}
	if preferLight {
		return ModeLight
	}
	return ModeDark
}
