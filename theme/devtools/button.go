package devtools

import (
	"github.com/gogpu/ui/core/button"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// ButtonPainter renders buttons using DevTools design tokens.
// DevTools buttons are compact (28px height) with 4px corner radius,
// matching JetBrains IDE toolbar and dialog button styling.
//
// If Theme is nil, ButtonPainter falls back to the default DevTools dark palette.
type ButtonPainter struct {
	Theme *Theme // nil uses default DevTools dark fallback
}

// resolveColors returns the ButtonColorScheme derived from the painter's Theme.
func (p ButtonPainter) resolveColors() button.ButtonColorScheme {
	if p.Theme == nil {
		return dtDefaultButtonColors
	}
	cs := p.Theme.Colors
	return button.ButtonColorScheme{
		FilledBg:       cs.Primary,
		FilledFg:       cs.OnPrimary,
		OutlinedBorder: cs.BorderStrong,
		TextBgHover:    cs.ControlHover,
		TonalBg:        cs.ControlFill,
		TonalFg:        cs.OnSurface,
		Primary:        cs.Primary,
		DisabledBg:     cs.ControlFill,
		DisabledFg:     cs.OnSurfaceDisabled,
		FocusRing:      cs.BorderFocus,
	}
}

// PaintButton renders a button according to DevTools design specifications.
func (p ButtonPainter) PaintButton(canvas widget.Canvas, state button.PaintState) {
	if state.Bounds.IsEmpty() {
		return
	}

	radius := dtButtonRadius
	if state.Radius != nil {
		radius = *state.Radius
	}

	colors := state.ColorScheme
	if colors == (button.ButtonColorScheme{}) {
		colors = p.resolveColors()
	}

	disabled := state.Disabled
	bg := dtButtonBackground(state, disabled, colors)
	fg := dtButtonForeground(state.Variant, disabled, colors)

	// Draw background based on variant.
	switch state.Variant {
	case button.Filled, button.Tonal:
		canvas.DrawRoundRect(state.Bounds, bg, radius)
	case button.Outlined:
		if state.Hovered || state.Pressed {
			canvas.DrawRoundRect(state.Bounds, colors.Primary.WithAlpha(0.08), radius)
		}
		canvas.StrokeRoundRect(state.Bounds, bg, radius, dtButtonBorderWidth)
	case button.TextOnly:
		if state.Hovered || state.Pressed {
			canvas.DrawRoundRect(state.Bounds, bg, radius)
		}
	}

	// Draw text centered.
	fontSize := dtButtonFontSize(state.Size)
	canvas.DrawText(state.Text, state.Bounds, fontSize, fg, false, dtTextAlignCenter)

	// Focus indicator: 1px border (DevTools style).
	if state.Focused && !disabled {
		dtDrawFocusRing(canvas, state.Bounds, radius, colors.FocusRing)
	}
}

// dtButtonBackground returns the DevTools background color for the current variant and state.
func dtButtonBackground(state button.PaintState, disabled bool, colors button.ButtonColorScheme) widget.Color {
	if disabled {
		return dtButtonDisabledBg(state.Variant, colors)
	}
	if state.Background != nil {
		return dtApplyState(*state.Background, state.Hovered, state.Pressed)
	}
	return dtButtonVariantBg(state.Variant, state.Hovered, state.Pressed, colors)
}

// dtButtonForeground returns the DevTools text color for the current variant.
func dtButtonForeground(v button.Variant, disabled bool, colors button.ButtonColorScheme) widget.Color {
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

// dtButtonVariantBg returns the background for a variant and interaction state.
func dtButtonVariantBg(v button.Variant, hovered, pressed bool, colors button.ButtonColorScheme) widget.Color {
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
	return dtApplyState(base, hovered, pressed)
}

// dtButtonDisabledBg returns the disabled background for a variant.
func dtButtonDisabledBg(v button.Variant, colors button.ButtonColorScheme) widget.Color {
	switch v {
	case button.Outlined, button.TextOnly:
		return widget.ColorTransparent
	default:
		return colors.DisabledBg
	}
}

// dtButtonFontSize returns the DevTools font size for a button size.
func dtButtonFontSize(s button.Size) float32 {
	switch s {
	case button.Small:
		return 11
	case button.Large:
		return 14
	default:
		return 13
	}
}

// dtDrawFocusRing draws a DevTools-style focus ring (1px accent border).
func dtDrawFocusRing(canvas widget.Canvas, bounds geometry.Rect, radius float32, color widget.Color) {
	canvas.StrokeRoundRect(bounds, color, radius, dtFocusRingStrokeWidth)
}

// dtApplyState adjusts a color based on DevTools interaction state.
// DevTools uses subtle lighten on hover, slight darken on press.
func dtApplyState(base widget.Color, hovered, pressed bool) widget.Color {
	if pressed {
		return base.Lerp(widget.ColorBlack, dtPressedDarkenFactor)
	}
	if hovered {
		return base.Lerp(widget.ColorWhite, dtHoverLightenFactor)
	}
	return base
}

// dtDefaultButtonColors holds the default DevTools dark color scheme for buttons.
var dtDefaultButtonColors = button.ButtonColorScheme{
	FilledBg:       DefaultAccentColor,
	FilledFg:       widget.ColorWhite,
	OutlinedBorder: widget.Hex(0x4E5157), // Gray5
	TextBgHover:    widget.Hex(0x43454A), // Gray4
	TonalBg:        widget.Hex(0x393B40), // Gray3
	TonalFg:        widget.Hex(0xDFE1E5), // Gray12
	Primary:        DefaultAccentColor,
	DisabledBg:     widget.Hex(0x393B40), // Gray3
	DisabledFg:     widget.Hex(0x6F737A), // Gray7
	FocusRing:      DefaultAccentColor,
}

// DevTools button drawing constants.
const (
	dtButtonRadius         float32 = 4
	dtButtonBorderWidth    float32 = 1
	dtTextAlignCenter              = widget.TextAlignCenter
	dtFocusRingStrokeWidth float32 = 1
	dtHoverLightenFactor   float32 = 0.08
	dtPressedDarkenFactor  float32 = 0.10
)

// Compile-time check that ButtonPainter implements Painter.
var _ button.Painter = ButtonPainter{}
