package fluent

import (
	"github.com/gogpu/ui/theme"
	"github.com/gogpu/ui/widget"
)

// DefaultAccentColor is the standard Windows Blue accent color.
var DefaultAccentColor = widget.Hex(0x0078D4)

// Theme provides Fluent Design System design tokens.
//
// A Theme contains the complete set of design tokens needed to style
// a Fluent Design application: colors (derived from an accent color),
// and references to the shared theme.Theme for typography, spacing,
// shadows, and radii.
//
// Create a theme with [NewTheme] or [NewDarkTheme]:
//
//	t := fluent.NewTheme()                                          // default light
//	t := fluent.NewTheme(fluent.WithAccentColor(widget.Hex(0x744DA9))) // custom accent
//	dark := fluent.NewDarkTheme()                                   // default dark
type Theme struct {
	// Colors holds the Fluent Design color scheme derived from the accent color.
	Colors ColorScheme

	// dark indicates whether this theme uses a dark color scheme.
	dark bool
}

// Option configures a Fluent theme.
type Option func(*themeConfig)

// themeConfig holds configuration for theme creation.
type themeConfig struct {
	accent widget.Color
}

// WithAccentColor sets the accent color for the Fluent theme.
// If not provided, the default Windows Blue (#0078D4) is used.
func WithAccentColor(accent widget.Color) Option {
	return func(c *themeConfig) {
		c.accent = accent
	}
}

// NewTheme creates a Fluent Design light theme.
//
// By default it uses Windows Blue (#0078D4) as the accent color.
// Use [WithAccentColor] to customize:
//
//	t := fluent.NewTheme(fluent.WithAccentColor(widget.Hex(0x744DA9)))
func NewTheme(opts ...Option) *Theme {
	cfg := themeConfig{accent: DefaultAccentColor}
	for _, o := range opts {
		o(&cfg)
	}
	return &Theme{
		Colors: LightScheme(cfg.accent),
	}
}

// NewDarkTheme creates a Fluent Design dark theme.
//
// By default it uses Windows Blue (#0078D4) as the accent color.
// Use [WithAccentColor] to customize.
func NewDarkTheme(opts ...Option) *Theme {
	cfg := themeConfig{accent: DefaultAccentColor}
	for _, o := range opts {
		o(&cfg)
	}
	return &Theme{
		Colors: DarkScheme(cfg.accent),
		dark:   true,
	}
}

// IsDark returns true if this theme uses a dark color scheme.
func (t *Theme) IsDark() bool {
	return t.dark
}

// OnSurface returns the default text/icon color for surface backgrounds.
//
// This satisfies the widget.ThemeProvider interface.
func (t *Theme) OnSurface() widget.Color {
	return t.Colors.OnSurface
}

// AsTheme converts the Fluent theme to a theme.Theme for use with the
// generic theme system. This maps Fluent color roles to the shared
// ColorPalette structure.
func (t *Theme) AsTheme() *theme.Theme {
	cs := t.Colors
	mode := theme.ModeLight
	name := "Fluent Light"
	shadows := theme.DefaultShadowsLight()
	if t.dark {
		mode = theme.ModeDark
		name = "Fluent Dark"
		shadows = theme.DefaultShadowsDark()
	}

	return &theme.Theme{
		Name: name,
		Mode: mode,
		Colors: theme.ColorPalette{
			Primary:        cs.Accent,
			PrimaryLight:   cs.AccentLight,
			PrimaryDark:    cs.AccentDark,
			Secondary:      cs.Accent,
			SecondaryLight: cs.AccentLight,
			SecondaryDark:  cs.AccentDark,
			Background:     cs.Surface,
			Surface:        cs.SurfaceSecondary,
			SurfaceVariant: cs.SurfaceTertiary,
			Error:          cs.Error,
			Warning:        cs.Warning,
			Success:        cs.Success,
			Info:           cs.Accent,
			OnPrimary:      cs.OnAccent,
			OnSecondary:    cs.OnAccent,
			OnBackground:   cs.OnSurface,
			OnSurface:      cs.OnSurface,
			OnError:        widget.ColorWhite,
			Divider:        cs.StrokeDefault,
			Outline:        cs.StrokeDefault,
			Shadow:         cs.Shadow,
		},
		Typography: theme.DefaultTypography(),
		Spacing:    theme.DefaultSpacing(),
		Shadows:    shadows,
		Radii:      fluentRadii(),
		Extensions: make(map[string]any),
	}
}

// fluentRadii returns the Fluent Design radius scale.
// Fluent uses smaller radii than Material 3 (4px default).
func fluentRadii() theme.RadiusScale {
	return theme.RadiusScale{
		None: 0,
		XS:   2,
		S:    4,
		M:    4,
		L:    8,
		XL:   12,
		XXL:  16,
		Full: 9999,
	}
}

// Compile-time check that Theme implements ThemeProvider.
var _ widget.ThemeProvider = (*Theme)(nil)
