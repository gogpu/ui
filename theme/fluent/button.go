package fluent

import (
	"github.com/gogpu/ui/core/button"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// ButtonPainter renders buttons using Fluent Design tokens.
// It maps button variants (Filled/Accent, Outlined/Standard, TextOnly/Subtle,
// Tonal) to the Fluent color scheme.
//
// If Theme is nil, ButtonPainter falls back to the default Fluent Blue palette.
type ButtonPainter struct {
	Theme *Theme // nil uses default Fluent Blue fallback
}

// resolveColors returns the ButtonColorScheme derived from the painter's Theme.
func (p ButtonPainter) resolveColors() button.ButtonColorScheme {
	if p.Theme == nil {
		return flDefaultButtonColors
	}
	cs := p.Theme.Colors
	return button.ButtonColorScheme{
		FilledBg:       cs.Accent,
		FilledFg:       cs.OnAccent,
		OutlinedBorder: cs.StrokeDefault,
		TextBgHover:    cs.FillSecond,
		TonalBg:        cs.AccentLight,
		TonalFg:        cs.AccentDark,
		Primary:        cs.Accent,
		DisabledBg:     cs.FillDisable,
		DisabledFg:     cs.OnSurfaceSecond.WithAlpha(flDisabledAlpha),
		FocusRing:      cs.StrokeFocus,
	}
}

// PaintButton renders a button according to Fluent Design specifications.
func (p ButtonPainter) PaintButton(canvas widget.Canvas, state button.PaintState) {
	if state.Bounds.IsEmpty() {
		return
	}

	radius := flButtonRadius
	if state.Radius != nil {
		radius = *state.Radius
	}

	colors := state.ColorScheme
	if colors == (button.ButtonColorScheme{}) {
		colors = p.resolveColors()
	}

	disabled := state.Disabled
	bg := flButtonBackground(state, disabled, colors)
	fg := flButtonForeground(state.Variant, disabled, colors)

	// Draw background based on variant.
	switch state.Variant {
	case button.Filled, button.Tonal:
		canvas.DrawRoundRect(state.Bounds, bg, radius)
	case button.Outlined:
		if state.Hovered || state.Pressed {
			canvas.DrawRoundRect(state.Bounds, colors.Primary.WithAlpha(0.08), radius)
		}
		canvas.StrokeRoundRect(state.Bounds, bg, radius, flButtonBorderWidth)
	case button.TextOnly:
		if state.Hovered || state.Pressed {
			canvas.DrawRoundRect(state.Bounds, bg, radius)
		}
	}

	// Draw text centered.
	fontSize := flButtonFontSize(state.Size)
	canvas.DrawText(state.Text, state.Bounds, fontSize, fg, false, flTextAlignCenter)

	// Focus indicator: inner ring (Fluent style).
	if state.Focused && !disabled {
		flDrawFocusRing(canvas, state.Bounds, radius, colors.FocusRing)
	}
}

// flButtonBackground returns the Fluent background color for the current variant and state.
func flButtonBackground(state button.PaintState, disabled bool, colors button.ButtonColorScheme) widget.Color {
	if disabled {
		return flButtonDisabledBg(state.Variant, colors)
	}
	if state.Background != nil {
		return flApplyState(*state.Background, state.Hovered, state.Pressed)
	}
	return flButtonVariantBg(state.Variant, state.Hovered, state.Pressed, colors)
}

// flButtonForeground returns the Fluent text color for the current variant.
func flButtonForeground(v button.Variant, disabled bool, colors button.ButtonColorScheme) widget.Color {
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

// flButtonVariantBg returns the background for a variant and interaction state.
func flButtonVariantBg(v button.Variant, hovered, pressed bool, colors button.ButtonColorScheme) widget.Color {
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
	return flApplyState(base, hovered, pressed)
}

// flButtonDisabledBg returns the disabled background for a variant.
func flButtonDisabledBg(v button.Variant, colors button.ButtonColorScheme) widget.Color {
	switch v {
	case button.Outlined, button.TextOnly:
		return widget.ColorTransparent
	default:
		return colors.DisabledBg
	}
}

// flButtonFontSize returns the Fluent font size for a button size.
func flButtonFontSize(s button.Size) float32 {
	switch s {
	case button.Small:
		return 12
	case button.Large:
		return 16
	default:
		return 14
	}
}

// flDrawFocusRing draws a Fluent-style inner focus ring.
func flDrawFocusRing(canvas widget.Canvas, bounds geometry.Rect, radius float32, color widget.Color) {
	innerBounds := bounds.Expand(-flFocusRingInset)
	innerRadius := radius - flFocusRingInset
	if innerRadius < 0 {
		innerRadius = 0
	}
	canvas.StrokeRoundRect(innerBounds, color, innerRadius, flFocusRingStrokeWidth)
}

// flApplyState adjusts a color based on Fluent interaction state.
func flApplyState(base widget.Color, hovered, pressed bool) widget.Color {
	if pressed {
		return base.Lerp(widget.ColorBlack, flPressedDarkenFactor)
	}
	if hovered {
		return base.Lerp(widget.ColorWhite, flHoverLightenFactor)
	}
	return base
}

// flDefaultButtonColors holds the default Fluent Blue color scheme for buttons.
var flDefaultButtonColors = button.ButtonColorScheme{
	FilledBg:       DefaultAccentColor,
	FilledFg:       widget.ColorWhite,
	OutlinedBorder: widget.RGBA(0, 0, 0, 0.14),
	TextBgHover:    widget.RGBA(0, 0, 0, 0.06),
	TonalBg:        lighten(DefaultAccentColor, 0.85),
	TonalFg:        darken(DefaultAccentColor, 0.25),
	Primary:        DefaultAccentColor,
	DisabledBg:     widget.RGBA(0, 0, 0, 0.04),
	DisabledFg:     widget.RGBA(0.38, 0.38, 0.38, 0.38),
	FocusRing:      DefaultAccentColor,
}

// Fluent button drawing constants.
const (
	flButtonRadius         float32 = 4
	flButtonBorderWidth    float32 = 1
	flTextAlignCenter              = widget.TextAlignCenter
	flFocusRingInset       float32 = 2
	flFocusRingStrokeWidth float32 = 2
	flHoverLightenFactor   float32 = 0.08
	flPressedDarkenFactor  float32 = 0.12
	flDisabledAlpha        float32 = 0.38
)

// ButtonHeight returns the Fluent height for the given button size.
func (ButtonPainter) ButtonHeight(size button.Size) float32 {
	switch size {
	case button.Small:
		return 28
	case button.Large:
		return 44
	default:
		return 36
	}
}

// ButtonPadding returns the Fluent horizontal and vertical padding.
func (ButtonPainter) ButtonPadding(size button.Size) (float32, float32) {
	switch size {
	case button.Small:
		return 10, 5
	case button.Large:
		return 20, 10
	default:
		return 16, 8
	}
}

// ButtonFontSize returns the Fluent font size for the given button size.
func (ButtonPainter) ButtonFontSize(size button.Size) float32 {
	return flButtonFontSize(size)
}

// ButtonRadius returns the Fluent default corner radius.
func (ButtonPainter) ButtonRadius() float32 { return flButtonRadius }

// Compile-time checks.
var (
	_ button.Painter       = ButtonPainter{}
	_ button.LayoutMetrics = ButtonPainter{}
)
