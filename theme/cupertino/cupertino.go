package cupertino

import "github.com/gogpu/ui/widget"

// Theme provides Apple Human Interface Guidelines (HIG) design tokens.
//
// A Theme contains the complete set of design tokens needed to style
// a Cupertino-style application: colors, corner radius, and font size.
// It can be generated with customizable accent color using [NewTheme].
//
// Create a theme with default system blue:
//
//	theme := cupertino.NewTheme()
//	accent := theme.Colors.Accent
//
// Create a dark theme:
//
//	darkTheme := cupertino.NewDarkTheme()
//
// Customize accent color:
//
//	theme := cupertino.NewTheme(cupertino.WithAccentColor(widget.Hex(0x34C759)))
type Theme struct {
	// Colors holds the Cupertino color scheme.
	Colors ColorScheme

	// Radius is the default corner radius for interactive elements (8-12px).
	Radius float32

	// dark indicates whether this theme uses a dark color scheme.
	dark bool
}

// Option configures a Cupertino theme.
type Option func(*themeConfig)

// themeConfig holds configuration for theme creation.
type themeConfig struct {
	accent widget.Color
}

// WithAccentColor sets the accent (tint) color for the theme.
// Default is Apple System Blue (#007AFF).
func WithAccentColor(accent widget.Color) Option {
	return func(c *themeConfig) {
		c.accent = accent
	}
}

// NewTheme creates a Cupertino light theme.
//
// By default it uses Apple System Blue (#007AFF) as the accent color.
// Use [WithAccentColor] to customize.
func NewTheme(opts ...Option) *Theme {
	cfg := &themeConfig{
		accent: systemBlue,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	return &Theme{
		Colors: lightColors(cfg.accent),
		Radius: defaultRadius,
	}
}

// NewDarkTheme creates a Cupertino dark theme.
//
// By default it uses Apple System Blue (#007AFF) as the accent color.
// Use [WithAccentColor] to customize.
func NewDarkTheme(opts ...Option) *Theme {
	cfg := &themeConfig{
		accent: systemBlue,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	return &Theme{
		Colors: darkColors(cfg.accent),
		Radius: defaultRadius,
		dark:   true,
	}
}

// IsDark returns true if this theme uses a dark color scheme.
func (t *Theme) IsDark() bool {
	return t.dark
}

// OnSurface returns the default text/icon color for surface backgrounds.
//
// This satisfies the widget.ThemeProvider interface and returns the
// Label color from the theme's color scheme.
func (t *Theme) OnSurface() widget.Color {
	return t.Colors.Label
}

// Default Cupertino constants.
const (
	// defaultRadius is the standard corner radius for Apple HIG elements.
	defaultRadius float32 = 10
)
