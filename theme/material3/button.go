package material3

import (
	"github.com/gogpu/ui/core/button"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// ButtonPainter renders buttons using Material 3 design tokens.
// It maps button variants (Filled, Outlined, TextOnly, Tonal) to
// the M3 color scheme and applies appropriate interaction feedback.
//
// If Theme is nil, ButtonPainter falls back to the default M3 purple palette.
type ButtonPainter struct {
	Theme *Theme // nil uses default M3 purple fallback
}

// resolveColors returns the ButtonColorScheme derived from the painter's Theme.
// If Theme is nil, it returns the default M3 purple color scheme.
func (p ButtonPainter) resolveColors() button.ButtonColorScheme {
	if p.Theme == nil {
		return m3DefaultColors
	}
	cs := p.Theme.Colors
	return button.ButtonColorScheme{
		FilledBg:       cs.Primary,
		FilledFg:       cs.OnPrimary,
		OutlinedBorder: cs.Outline,
		TextBgHover:    cs.Primary.WithAlpha(0.08),
		TonalBg:        cs.SecondaryContainer,
		TonalFg:        cs.OnSecondaryContainer,
		Primary:        cs.Primary,
		DisabledBg:     cs.OnSurface.WithAlpha(0.12),
		DisabledFg:     cs.OnSurface.WithAlpha(0.38),
		FocusRing:      cs.Primary.WithAlpha(0.7),
	}
}

// PaintButton renders a button according to Material 3 specifications.
func (p ButtonPainter) PaintButton(canvas widget.Canvas, state button.PaintState) {
	if state.Bounds.IsEmpty() {
		return
	}

	radius := m3DefaultRadius
	if state.Radius != nil {
		radius = *state.Radius
	}

	// Determine the color scheme to use.
	colors := state.ColorScheme
	if colors == (button.ButtonColorScheme{}) {
		colors = p.resolveColors()
	}

	disabled := state.Disabled
	bg := m3ResolvedBackground(state, disabled, colors)
	fg := m3ResolvedForeground(state.Variant, disabled, colors)

	// Draw background based on variant.
	switch state.Variant {
	case button.Filled, button.Tonal:
		canvas.DrawRoundRect(state.Bounds, bg, radius)
	case button.Outlined:
		if state.Hovered || state.Pressed {
			canvas.DrawRoundRect(state.Bounds, colors.Primary.WithAlpha(0.08), radius)
		}
		canvas.StrokeRoundRect(state.Bounds, bg, radius, m3OutlineStrokeWidth)
	case button.TextOnly:
		if state.Hovered || state.Pressed {
			canvas.DrawRoundRect(state.Bounds, bg, radius)
		}
	}

	// Draw text centered in bounds.
	fontSize := m3FontSize(state.Size)
	bold := state.Variant == button.Filled
	canvas.DrawText(state.Text, state.Bounds, fontSize, fg, bold, m3TextAlignCenter)

	// Draw focus ring when focused.
	if state.Focused && !disabled {
		m3DrawFocusIndicator(canvas, state.Bounds, radius, colors)
	}
}

// m3ResolvedBackground returns the M3 background color for the current variant and state.
func m3ResolvedBackground(state button.PaintState, disabled bool, colors button.ButtonColorScheme) widget.Color {
	if disabled {
		return m3DisabledBackground(state.Variant, colors)
	}
	if state.Background != nil {
		return m3ApplyState(*state.Background, state.Hovered, state.Pressed)
	}
	return m3VariantBackground(state.Variant, state.Hovered, state.Pressed, colors)
}

// m3ResolvedForeground returns the M3 text color for the current variant.
func m3ResolvedForeground(v button.Variant, disabled bool, colors button.ButtonColorScheme) widget.Color {
	if disabled {
		return colors.DisabledFg
	}
	switch v {
	case button.Filled:
		return colors.FilledFg
	case button.Outlined, button.TextOnly:
		return colors.Primary
	case button.Tonal:
		return colors.TonalFg
	default:
		return colors.FilledFg
	}
}

// m3VariantBackground returns the M3 background color for a variant and state.
func m3VariantBackground(v button.Variant, hovered, pressed bool, colors button.ButtonColorScheme) widget.Color {
	var base widget.Color
	switch v {
	case button.Filled:
		base = colors.FilledBg
	case button.Outlined:
		base = colors.OutlinedBorder
	case button.TextOnly:
		base = colors.TextBgHover
	case button.Tonal:
		base = colors.TonalBg
	default:
		base = colors.FilledBg
	}
	return m3ApplyState(base, hovered, pressed)
}

// m3DisabledBackground returns the M3 disabled background color for a variant.
func m3DisabledBackground(v button.Variant, colors button.ButtonColorScheme) widget.Color {
	switch v {
	case button.TextOnly:
		return widget.ColorTransparent
	case button.Outlined:
		return colors.DisabledFg
	default:
		return colors.DisabledBg
	}
}

// m3ApplyState adjusts a color based on interaction state.
func m3ApplyState(base widget.Color, hovered, pressed bool) widget.Color {
	if pressed {
		return base.Lerp(widget.ColorBlack, m3PressedDarkenFactor)
	}
	if hovered {
		return base.Lerp(widget.ColorWhite, m3HoverLightenFactor)
	}
	return base
}

// m3DrawFocusIndicator draws a focus ring around the button bounds.
func m3DrawFocusIndicator(canvas widget.Canvas, bounds geometry.Rect, radius float32, colors button.ButtonColorScheme) {
	ringBounds := bounds.Expand(m3FocusRingOffset)
	ringRadius := radius + m3FocusRingOffset
	canvas.StrokeRoundRect(ringBounds, colors.FocusRing, ringRadius, m3FocusRingStrokeWidth)
}

// m3FontSize returns the M3 font size for a button size.
func m3FontSize(s button.Size) float32 {
	switch s {
	case button.Small:
		return 12
	case button.Large:
		return 16
	default:
		return 14
	}
}

// m3DefaultColors holds the default M3 purple color scheme for buttons.
// Used as a fallback when no Theme is provided.
var m3DefaultColors = button.ButtonColorScheme{
	FilledBg:       widget.Hex(0x6750A4),                // M3 primary
	FilledFg:       widget.ColorWhite,                   // M3 on-primary
	OutlinedBorder: widget.Hex(0x79747E),                // M3 outline
	TextBgHover:    widget.RGBA(0.4, 0.31, 0.64, 0.08),  // M3 primary hover overlay
	TonalBg:        widget.Hex(0xE8DEF8),                // M3 secondary container
	TonalFg:        widget.Hex(0x1D192B),                // M3 on secondary container
	Primary:        widget.Hex(0x6750A4),                // M3 primary
	DisabledBg:     widget.RGBA(0.12, 0.12, 0.13, 0.12), // M3 disabled bg
	DisabledFg:     widget.RGBA(0.12, 0.12, 0.13, 0.38), // M3 disabled fg
	FocusRing:      widget.Hex(0x6750A4).WithAlpha(0.7), // M3 focus ring
}

// M3 drawing constants.
const (
	m3DefaultRadius        float32 = 8
	m3OutlineStrokeWidth   float32 = 1.5
	m3FocusRingOffset      float32 = 2
	m3FocusRingStrokeWidth float32 = 2
	m3TextAlignCenter              = widget.TextAlignCenter
	m3HoverLightenFactor   float32 = 0.1
	m3PressedDarkenFactor  float32 = 0.15
)

// ButtonHeight returns the M3 height for the given button size.
func (ButtonPainter) ButtonHeight(size button.Size) float32 {
	switch size {
	case button.Small:
		return 32
	case button.Large:
		return 48
	default:
		return 40
	}
}

// ButtonPadding returns the M3 horizontal and vertical padding.
func (ButtonPainter) ButtonPadding(size button.Size) (float32, float32) {
	switch size {
	case button.Small:
		return 12, 6
	case button.Large:
		return 24, 12
	default:
		return 16, 8
	}
}

// ButtonFontSize returns the M3 font size for the given button size.
func (ButtonPainter) ButtonFontSize(size button.Size) float32 {
	return m3FontSize(size)
}

// ButtonRadius returns the M3 default corner radius.
func (ButtonPainter) ButtonRadius() float32 { return m3DefaultRadius }

// Compile-time checks.
var (
	_ button.Painter       = ButtonPainter{}
	_ button.LayoutMetrics = ButtonPainter{}
)
