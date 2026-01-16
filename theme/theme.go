package theme

// Theme defines the complete visual style for an application.
//
// Theme combines all visual properties needed to style UI components:
// colors, typography, spacing, shadows, and border radii. Themes are
// designed to be immutable by convention - once created, they should
// not be modified. This allows safe sharing across goroutines.
//
// Use [DefaultLight] or [DefaultDark] for pre-configured themes, or
// build custom themes by modifying a base theme:
//
//	func MyBrandTheme() *Theme {
//	    t := DefaultLight()
//	    t.Colors.Primary = widget.Hex(0x6200EE)
//	    return t
//	}
//
// Themes can be extended with custom data using the Extensions map,
// which allows third-party components to add their own theme properties.
type Theme struct {
	// Name is a human-readable identifier for the theme.
	//
	// Examples: "Light", "Dark", "Ocean Blue", "High Contrast"
	Name string

	// Mode indicates whether this is a light, dark, or system-following theme.
	Mode ThemeMode

	// Colors defines the semantic color palette.
	Colors ColorPalette

	// Typography defines the type scale and text styles.
	Typography Typography

	// Spacing defines the spacing scale for margins and padding.
	Spacing SpacingScale

	// Shadows defines the elevation-based shadow styles.
	Shadows ShadowStyles

	// Radii defines the border radius scale.
	Radii RadiusScale

	// Extensions allows custom theme data for third-party components.
	//
	// Components can store their own theme properties here using a unique
	// key (typically the component package path). This enables theming
	// of custom components while maintaining compatibility with the core
	// theme system.
	//
	// Example:
	//
	//	type MyComponentTheme struct {
	//	    AccentColor widget.Color
	//	    BorderWidth float32
	//	}
	//
	//	// Store custom theme data
	//	theme.Extensions["myorg/mycomponent"] = MyComponentTheme{
	//	    AccentColor: widget.Hex(0xFF5722),
	//	    BorderWidth: 2,
	//	}
	//
	//	// Retrieve custom theme data
	//	if ext, ok := theme.Extensions["myorg/mycomponent"].(MyComponentTheme); ok {
	//	    // Use ext.AccentColor, ext.BorderWidth
	//	}
	Extensions map[string]any
}

// New creates a new Theme with the given name and mode.
//
// The theme is initialized with default values for all properties.
// Use this as a starting point for custom themes.
//
// Example:
//
//	theme := theme.New("My Brand", theme.ModeLight)
//	theme.Colors.Primary = widget.Hex(0x6200EE)
func New(name string, mode ThemeMode) *Theme {
	return &Theme{
		Name:       name,
		Mode:       mode,
		Colors:     ColorPalette{}, // Will be filled by caller
		Typography: DefaultTypography(),
		Spacing:    DefaultSpacing(),
		Shadows:    DefaultShadowsLight(),
		Radii:      DefaultRadii(),
		Extensions: make(map[string]any),
	}
}

// Clone creates a deep copy of the theme.
//
// This is useful when you want to create a variant of an existing theme
// without modifying the original.
//
// Note: Extensions are shallow-copied. If you have mutable extension
// values, you should deep-copy them manually.
func (t *Theme) Clone() *Theme {
	if t == nil {
		return nil
	}

	clone := &Theme{
		Name:       t.Name,
		Mode:       t.Mode,
		Colors:     t.Colors,
		Typography: t.Typography,
		Spacing:    t.Spacing,
		Shadows:    t.Shadows,
		Radii:      t.Radii,
		Extensions: make(map[string]any, len(t.Extensions)),
	}

	// Shallow copy extensions
	for k, v := range t.Extensions {
		clone.Extensions[k] = v
	}

	return clone
}

// WithName returns a copy of the theme with a different name.
func (t *Theme) WithName(name string) *Theme {
	clone := t.Clone()
	clone.Name = name
	return clone
}

// WithMode returns a copy of the theme with a different mode.
//
// Note: This only changes the Mode field. For a full light/dark switch,
// you should also update Colors and Shadows appropriately.
func (t *Theme) WithMode(mode ThemeMode) *Theme {
	clone := t.Clone()
	clone.Mode = mode
	return clone
}

// WithColors returns a copy of the theme with a different color palette.
func (t *Theme) WithColors(colors *ColorPalette) *Theme {
	clone := t.Clone()
	clone.Colors = *colors
	return clone
}

// WithTypography returns a copy of the theme with different typography.
func (t *Theme) WithTypography(typography *Typography) *Theme {
	clone := t.Clone()
	clone.Typography = *typography
	return clone
}

// WithSpacing returns a copy of the theme with a different spacing scale.
func (t *Theme) WithSpacing(spacing SpacingScale) *Theme {
	clone := t.Clone()
	clone.Spacing = spacing
	return clone
}

// WithShadows returns a copy of the theme with different shadow styles.
func (t *Theme) WithShadows(shadows *ShadowStyles) *Theme {
	clone := t.Clone()
	clone.Shadows = *shadows
	return clone
}

// WithRadii returns a copy of the theme with a different radius scale.
func (t *Theme) WithRadii(radii RadiusScale) *Theme {
	clone := t.Clone()
	clone.Radii = radii
	return clone
}

// SetExtension sets a custom extension value on the theme.
//
// This modifies the theme in place. For immutable usage, use
// Clone() first.
func (t *Theme) SetExtension(key string, value any) {
	if t.Extensions == nil {
		t.Extensions = make(map[string]any)
	}
	t.Extensions[key] = value
}

// GetExtension retrieves a custom extension value from the theme.
//
// Returns the value and true if found, or nil and false if not found.
func (t *Theme) GetExtension(key string) (any, bool) {
	if t.Extensions == nil {
		return nil, false
	}
	v, ok := t.Extensions[key]
	return v, ok
}

// IsLight returns true if this theme uses a light color scheme.
func (t *Theme) IsLight() bool {
	return t.Mode == ModeLight
}

// IsDark returns true if this theme uses a dark color scheme.
func (t *Theme) IsDark() bool {
	return t.Mode == ModeDark
}

// ScaleTypography returns a copy of the theme with scaled typography.
//
// This is useful for accessibility settings that increase text size.
// Factor of 1.0 keeps original sizes, 1.25 increases by 25%, etc.
func (t *Theme) ScaleTypography(factor float32) *Theme {
	clone := t.Clone()
	clone.Typography = t.Typography.Scale(factor)
	return clone
}

// ScaleSpacing returns a copy of the theme with scaled spacing.
//
// This is useful for density adjustments.
// Factor of 1.0 keeps original spacing, 0.75 creates compact spacing, etc.
func (t *Theme) ScaleSpacing(factor float32) *Theme {
	clone := t.Clone()
	clone.Spacing = t.Spacing.Scale(factor)
	return clone
}

// Compact returns a copy of the theme with reduced spacing (75%).
//
// This creates a denser layout suitable for data-heavy interfaces.
func (t *Theme) Compact() *Theme {
	return t.ScaleSpacing(0.75)
}

// Comfortable returns a copy of the theme with increased spacing (125%).
//
// This creates a more spacious layout suitable for touch interfaces.
func (t *Theme) Comfortable() *Theme {
	return t.ScaleSpacing(1.25)
}
