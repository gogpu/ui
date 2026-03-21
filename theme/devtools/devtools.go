package devtools

import (
	"github.com/gogpu/ui/theme"
	"github.com/gogpu/ui/widget"
)

// DefaultAccentColor is the JetBrains default blue accent (#3574F0).
var DefaultAccentColor = widget.Hex(0x3574F0)

// Theme provides JetBrains-inspired DevTools design tokens.
//
// A Theme contains the complete set of design tokens needed to style
// a DevTools application: colors (based on JetBrains Int UI gray scale
// with a customizable accent), and references to the shared theme.Theme
// for typography, spacing, shadows, and radii.
//
// DevTools is dark-first: [NewDarkTheme] produces the primary theme,
// while [NewTheme] produces a light variant. Both themes feature a dark
// header toolbar, matching JetBrains IDE behavior.
//
// Create a theme:
//
//	dark := devtools.NewDarkTheme()                                         // default dark
//	dark := devtools.NewDarkTheme(devtools.WithAccentColor(widget.Hex(0x57965C))) // green accent
//	light := devtools.NewTheme()                                           // default light
type Theme struct {
	// Colors holds the DevTools color scheme.
	Colors ColorScheme

	// dark indicates whether this theme uses a dark color scheme.
	dark bool
}

// Option configures a DevTools theme.
type Option func(*themeConfig)

// themeConfig holds configuration for theme creation.
type themeConfig struct {
	accent widget.Color
}

// WithAccentColor sets the accent color for the DevTools theme.
// If not provided, the default JetBrains Blue (#3574F0) is used.
//
// The accent color is applied to Primary, PrimaryHover, PrimaryPress,
// BorderFocus, and Info roles. Hover and press variants are derived
// automatically.
func WithAccentColor(accent widget.Color) Option {
	return func(c *themeConfig) {
		c.accent = accent
	}
}

// NewTheme creates a DevTools light theme.
//
// By default it uses JetBrains Blue (#3574F0) as the accent color.
// Use [WithAccentColor] to customize:
//
//	t := devtools.NewTheme(devtools.WithAccentColor(widget.Hex(0x57965C)))
func NewTheme(opts ...Option) *Theme {
	cfg := themeConfig{accent: DefaultAccentColor}
	for _, o := range opts {
		o(&cfg)
	}
	cs := LightScheme()
	applyAccent(&cs, cfg.accent, false)
	return &Theme{Colors: cs}
}

// NewDarkTheme creates a DevTools dark theme.
//
// This is the primary theme for DevTools, optimized for long sessions.
// By default it uses JetBrains Blue (#3574F0) as the accent color.
// Use [WithAccentColor] to customize.
func NewDarkTheme(opts ...Option) *Theme {
	cfg := themeConfig{accent: DefaultAccentColor}
	for _, o := range opts {
		o(&cfg)
	}
	cs := DarkScheme()
	applyAccent(&cs, cfg.accent, true)
	return &Theme{Colors: cs, dark: true}
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

// AsTheme converts the DevTools theme to a theme.Theme for use with
// the generic theme system. This maps DevTools color roles to the
// shared ColorPalette structure.
func (t *Theme) AsTheme() *theme.Theme {
	cs := t.Colors
	mode := theme.ModeLight
	name := "DevTools Light"
	shadows := theme.DefaultShadowsLight()
	if t.dark {
		mode = theme.ModeDark
		name = "DevTools Dark"
		shadows = theme.DefaultShadowsDark()
	}

	return &theme.Theme{
		Name: name,
		Mode: mode,
		Colors: theme.ColorPalette{
			Primary:        cs.Primary,
			PrimaryLight:   cs.PrimaryHover,
			PrimaryDark:    cs.PrimaryPress,
			Secondary:      cs.Primary,
			SecondaryLight: cs.PrimaryHover,
			SecondaryDark:  cs.PrimaryPress,
			Background:     cs.Background,
			Surface:        cs.Surface,
			SurfaceVariant: cs.SurfaceElevated,
			Error:          cs.Error,
			Warning:        cs.Warning,
			Success:        cs.Success,
			Info:           cs.Info,
			OnPrimary:      cs.OnPrimary,
			OnSecondary:    cs.OnPrimary,
			OnBackground:   cs.OnSurface,
			OnSurface:      cs.OnSurface,
			OnError:        widget.ColorWhite,
			Divider:        cs.Border,
			Outline:        cs.BorderStrong,
			Shadow:         cs.Shadow,
		},
		Typography: devtoolsTypography(),
		Spacing:    devtoolsSpacing(),
		Shadows:    shadows,
		Radii:      devtoolsRadii(),
		Extensions: make(map[string]any),
	}
}

// applyAccent overrides accent-derived color roles with the given accent.
func applyAccent(cs *ColorScheme, accent widget.Color, dark bool) {
	cs.Primary = accent
	cs.BorderFocus = accent
	cs.Info = accent

	if dark {
		cs.PrimaryHover = lighten(accent, 0.10)
		cs.PrimaryPress = lighten(accent, 0.20)
	} else {
		cs.PrimaryHover = darken(accent, 0.10)
		cs.PrimaryPress = darken(accent, 0.20)
	}
}

// devtoolsTypography returns a compact type scale with 13px base size.
//
// JetBrains IDEs use 13px as the default UI font size, which is 1px smaller
// than Material 3 and Fluent Design. This creates a high-density information
// display suitable for developer tools.
func devtoolsTypography() theme.Typography {
	font := "System"
	ts := func(size, lineHeight float32, weight theme.FontWeight) theme.TextStyle {
		return theme.TextStyle{
			Font:       font,
			Size:       size,
			Weight:     weight,
			LineHeight: lineHeight,
		}
	}

	return theme.Typography{
		FontFamily: font,
		// Display: scaled down proportionally from M3
		DisplayLarge:  ts(48, 56, theme.FontWeightNormal),
		DisplayMedium: ts(38, 46, theme.FontWeightNormal),
		DisplaySmall:  ts(30, 38, theme.FontWeightNormal),
		// Headline
		HeadlineLarge:  ts(26, 34, theme.FontWeightNormal),
		HeadlineMedium: ts(22, 30, theme.FontWeightNormal),
		HeadlineSmall:  ts(18, 26, theme.FontWeightBold),
		// Title
		TitleLarge:  ts(18, 24, theme.FontWeightNormal),
		TitleMedium: ts(14, 20, theme.FontWeightSemiBold),
		TitleSmall:  ts(13, 18, theme.FontWeightSemiBold),
		// Body (13px base)
		BodyLarge:  ts(14, 20, theme.FontWeightNormal),
		BodyMedium: ts(13, 18, theme.FontWeightNormal),
		BodySmall:  ts(12, 16, theme.FontWeightNormal),
		// Label
		LabelLarge:  ts(13, 18, theme.FontWeightMedium),
		LabelMedium: ts(12, 16, theme.FontWeightMedium),
		LabelSmall:  ts(11, 14, theme.FontWeightMedium),
	}
}

// devtoolsSpacing returns a compact spacing scale for high-density layouts.
//
// DevTools spacing is consistently smaller than Material 3 to maximize
// information density, matching JetBrains IDE panel and tool window spacing.
func devtoolsSpacing() theme.SpacingScale {
	return theme.SpacingScale{
		XXS:  2,
		XS:   4,
		S:    6,
		M:    8,
		L:    12,
		XL:   16,
		XXL:  24,
		XXXL: 32,
	}
}

// devtoolsRadii returns the DevTools border radius scale.
//
// DevTools uses conservative radii: 4px default for most components (slightly
// less rounded than Material 3's 8px default). This matches JetBrains Int UI
// component styling.
func devtoolsRadii() theme.RadiusScale {
	return theme.RadiusScale{
		None: 0,
		XS:   2,
		S:    4,
		M:    8,
		L:    8,
		XL:   12,
		XXL:  12,
		Full: 9999,
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

// Compile-time check that Theme implements ThemeProvider.
var _ widget.ThemeProvider = (*Theme)(nil)
