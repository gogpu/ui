package cupertino

import (
	"github.com/gogpu/ui/core/button"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// ButtonPainter renders buttons using Apple HIG design tokens.
// Cupertino buttons use rounded pill shapes with system blue fills.
//
// Button variants map to iOS styles:
//   - Filled: rounded rect with accent color fill (primary CTA)
//   - Outlined: rounded rect with accent color border
//   - TextOnly: text-only button with accent color text
//   - Tonal: rounded rect with light accent tint fill
//
// If Theme is nil, ButtonPainter falls back to the default system blue palette.
type ButtonPainter struct {
	Theme *Theme // nil uses default system blue fallback
}

// resolveColors returns the ButtonColorScheme derived from the painter's Theme.
func (p ButtonPainter) resolveColors() button.ButtonColorScheme {
	if p.Theme == nil {
		return cupDefaultButtonColors
	}
	cs := p.Theme.Colors
	return button.ButtonColorScheme{
		FilledBg:       cs.Accent,
		FilledFg:       cs.OnAccent,
		OutlinedBorder: cs.Accent,
		TextBgHover:    cs.Accent.WithAlpha(cupBtnHoverAlpha),
		TonalBg:        cs.Accent.WithAlpha(cupBtnTonalAlpha),
		TonalFg:        cs.Accent,
		Primary:        cs.Accent,
		DisabledBg:     cs.QuaternaryLabel,
		DisabledFg:     cs.TertiaryLabel,
		FocusRing:      cs.Accent.WithAlpha(cupBtnFocusAlpha),
	}
}

// PaintButton renders a button according to Apple HIG specifications.
func (p ButtonPainter) PaintButton(canvas widget.Canvas, state button.PaintState) {
	if state.Bounds.IsEmpty() {
		return
	}

	radius := cupBtnDefaultRadius
	if state.Radius != nil {
		radius = *state.Radius
	}

	colors := state.ColorScheme
	if colors == (button.ButtonColorScheme{}) {
		colors = p.resolveColors()
	}

	disabled := state.Disabled
	bg := cupResolvedBtnBackground(state, disabled, colors)
	fg := cupResolvedBtnForeground(state.Variant, disabled, colors)

	switch state.Variant {
	case button.Filled, button.Tonal:
		canvas.DrawRoundRect(state.Bounds, bg, radius)
	case button.Outlined:
		if state.Hovered || state.Pressed {
			canvas.DrawRoundRect(state.Bounds, colors.Primary.WithAlpha(0.08), radius)
		}
		canvas.StrokeRoundRect(state.Bounds, bg, radius, cupBtnOutlineStroke)
	case button.TextOnly:
		if state.Hovered || state.Pressed {
			canvas.DrawRoundRect(state.Bounds, bg, radius)
		}
	}

	fontSize := cupBtnFontSize(state.Size)
	bold := state.Variant == button.Filled
	canvas.DrawText(state.Text, state.Bounds, fontSize, fg, bold, cupBtnTextAlign)

	if state.Focused && !disabled {
		cupDrawBtnFocusRing(canvas, state.Bounds, radius, colors)
	}
}

// cupResolvedBtnBackground returns the background color for the current variant and state.
func cupResolvedBtnBackground(state button.PaintState, disabled bool, colors button.ButtonColorScheme) widget.Color {
	if disabled {
		return cupDisabledBtnBackground(state.Variant, colors)
	}
	if state.Background != nil {
		return cupApplyBtnState(*state.Background, state.Hovered, state.Pressed)
	}
	return cupVariantBtnBackground(state.Variant, state.Hovered, state.Pressed, colors)
}

// cupResolvedBtnForeground returns the text color for the current variant.
func cupResolvedBtnForeground(v button.Variant, disabled bool, colors button.ButtonColorScheme) widget.Color {
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

// cupVariantBtnBackground returns the background color for a variant and state.
func cupVariantBtnBackground(v button.Variant, hovered, pressed bool, colors button.ButtonColorScheme) widget.Color {
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
	return cupApplyBtnState(base, hovered, pressed)
}

// cupDisabledBtnBackground returns the disabled background for a variant.
func cupDisabledBtnBackground(v button.Variant, colors button.ButtonColorScheme) widget.Color {
	switch v {
	case button.Outlined, button.TextOnly:
		return widget.ColorTransparent
	default:
		return colors.DisabledBg
	}
}

// cupApplyBtnState adjusts a color based on interaction state.
func cupApplyBtnState(base widget.Color, hovered, pressed bool) widget.Color {
	if pressed {
		return base.Lerp(widget.ColorBlack, cupBtnPressedDarken)
	}
	if hovered {
		return base.Lerp(widget.ColorWhite, cupBtnHoverLighten)
	}
	return base
}

// cupDrawBtnFocusRing draws the blue focus ring around the button.
func cupDrawBtnFocusRing(canvas widget.Canvas, bounds geometry.Rect, radius float32, colors button.ButtonColorScheme) {
	ringBounds := bounds.Expand(cupBtnFocusRingOffset)
	ringRadius := radius + cupBtnFocusRingOffset
	canvas.StrokeRoundRect(ringBounds, colors.FocusRing, ringRadius, cupBtnFocusRingStroke)
}

// cupBtnFontSize returns the font size for a button size.
func cupBtnFontSize(s button.Size) float32 {
	switch s {
	case button.Small:
		return 13
	case button.Large:
		return 17
	default:
		return 15
	}
}

// cupDefaultButtonColors holds the default system blue color scheme for buttons.
var cupDefaultButtonColors = button.ButtonColorScheme{
	FilledBg:       systemBlue,
	FilledFg:       widget.ColorWhite,
	OutlinedBorder: systemBlue,
	TextBgHover:    systemBlue.WithAlpha(cupBtnHoverAlpha),
	TonalBg:        systemBlue.WithAlpha(cupBtnTonalAlpha),
	TonalFg:        systemBlue,
	Primary:        systemBlue,
	DisabledBg:     widget.RGBA(0.235, 0.235, 0.263, 0.18),
	DisabledFg:     widget.RGBA(0.235, 0.235, 0.263, 0.3),
	FocusRing:      systemBlue.WithAlpha(cupBtnFocusAlpha),
}

// Cupertino button drawing constants.
const (
	cupBtnDefaultRadius   float32 = 10
	cupBtnOutlineStroke   float32 = 1.5
	cupBtnFocusRingOffset float32 = 3
	cupBtnFocusRingStroke float32 = 2.5
	cupBtnTextAlign               = widget.TextAlignCenter
	cupBtnHoverLighten    float32 = 0.08
	cupBtnPressedDarken   float32 = 0.12
	cupBtnHoverAlpha      float32 = 0.08
	cupBtnTonalAlpha      float32 = 0.15
	cupBtnFocusAlpha      float32 = 0.6
)

// Compile-time check that ButtonPainter implements Painter.
var _ button.Painter = ButtonPainter{}
