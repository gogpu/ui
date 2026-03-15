package cupertino

import (
	"github.com/gogpu/ui/core/radio"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// RadioPainter renders radio items using Apple HIG design tokens.
// Cupertino radio buttons use filled circles with accent-colored selection.
//
// If Theme is nil, RadioPainter falls back to the default system blue palette.
type RadioPainter struct {
	Theme *Theme // nil uses default system blue fallback
}

// resolveColors returns the RadioColorScheme derived from the painter's Theme.
func (p RadioPainter) resolveColors() radio.RadioColorScheme {
	if p.Theme == nil {
		return cupDefaultRadioColors
	}
	cs := p.Theme.Colors
	return radio.RadioColorScheme{
		SelectedBg:       cs.Accent,
		SelectedFg:       cs.OnAccent,
		UnselectedBorder: cs.OpaqueSeparator,
		LabelColor:       cs.Label,
		DisabledBg:       cs.QuaternaryLabel,
		DisabledFg:       cs.TertiaryLabel,
		FocusRing:        cs.Accent.WithAlpha(cupRadioFocusAlpha),
	}
}

// PaintRadio renders a radio item according to Apple HIG specifications.
func (p RadioPainter) PaintRadio(canvas widget.Canvas, state radio.PaintState) {
	if state.Bounds.IsEmpty() {
		return
	}

	colors := state.ColorScheme
	if colors == (radio.RadioColorScheme{}) {
		colors = p.resolveColors()
	}

	disabled := state.Disabled
	center, radius := cupRadioCircleGeometry(state.Bounds)

	if state.Selected {
		cupPaintSelectedRadio(canvas, center, radius, state, disabled, colors)
	} else {
		cupPaintUnselectedRadio(canvas, center, radius, state, disabled, colors)
	}

	// Label.
	if state.Label != "" {
		fg := colors.LabelColor
		if disabled {
			fg = colors.DisabledFg
		}
		labelBounds := cupRadioLabelBounds(state.Bounds)
		canvas.DrawText(state.Label, labelBounds, cupRadioFontSize, fg, false, cupRadioTextAlignLeft)
	}

	// Focus ring.
	if state.Focused && !disabled {
		canvas.StrokeCircle(center, radius+cupRadioFocusRingOffset, colors.FocusRing, cupRadioFocusRingStroke)
	}
}

// cupPaintSelectedRadio draws the radio in selected state.
func cupPaintSelectedRadio(canvas widget.Canvas, center geometry.Point, radius float32, state radio.PaintState, disabled bool, colors radio.RadioColorScheme) {
	bg := colors.SelectedBg
	if disabled {
		bg = colors.DisabledBg
	} else {
		bg = cupApplyRadioState(bg, state.Hovered, state.Pressed)
	}
	canvas.DrawCircle(center, radius, bg)

	fg := colors.SelectedFg
	if disabled {
		fg = colors.DisabledFg
	}
	canvas.DrawCircle(center, cupRadioInnerRadius, fg)
}

// cupPaintUnselectedRadio draws the radio in unselected state.
func cupPaintUnselectedRadio(canvas widget.Canvas, center geometry.Point, radius float32, state radio.PaintState, disabled bool, colors radio.RadioColorScheme) {
	borderColor := colors.UnselectedBorder
	if disabled {
		borderColor = colors.DisabledFg
	} else {
		borderColor = cupApplyRadioState(borderColor, state.Hovered, state.Pressed)
	}
	canvas.StrokeCircle(center, radius, borderColor, cupRadioBorderWidth)
}

// cupApplyRadioState adjusts a color based on interaction state.
func cupApplyRadioState(base widget.Color, hovered, pressed bool) widget.Color {
	if pressed {
		return base.Lerp(widget.ColorBlack, cupRadioPressedDarken)
	}
	if hovered {
		return base.Lerp(widget.ColorWhite, cupRadioHoverLighten)
	}
	return base
}

// cupRadioCircleGeometry returns the center and radius for the outer radio circle.
func cupRadioCircleGeometry(bounds geometry.Rect) (geometry.Point, float32) {
	h := bounds.Height()
	cx := bounds.Min.X + cupRadioOuterRadius
	cy := bounds.Min.Y + h/2
	return geometry.Pt(cx, cy), cupRadioOuterRadius
}

// cupRadioLabelBounds returns the label area to the right of the radio circle.
func cupRadioLabelBounds(bounds geometry.Rect) geometry.Rect {
	labelX := bounds.Min.X + cupRadioOuterRadius*2 + cupRadioLabelGap
	labelW := bounds.Width() - cupRadioOuterRadius*2 - cupRadioLabelGap
	if labelW < 0 {
		labelW = 0
	}
	return geometry.NewRect(labelX, bounds.Min.Y, labelW, bounds.Height())
}

// cupDefaultRadioColors holds the default system blue color scheme for radio items.
var cupDefaultRadioColors = radio.RadioColorScheme{
	SelectedBg:       systemBlue,
	SelectedFg:       widget.ColorWhite,
	UnselectedBorder: widget.Hex(0xC6C6C8),
	LabelColor:       widget.RGBA(0.0, 0.0, 0.0, 1.0),
	DisabledBg:       widget.RGBA(0.235, 0.235, 0.263, 0.18),
	DisabledFg:       widget.RGBA(0.235, 0.235, 0.263, 0.3),
	FocusRing:        systemBlue.WithAlpha(0.6),
}

// Cupertino radio drawing constants.
const (
	cupRadioOuterRadius     float32 = 10
	cupRadioInnerRadius     float32 = 4
	cupRadioLabelGap        float32 = 8
	cupRadioBorderWidth     float32 = 2
	cupRadioFontSize        float32 = 15
	cupRadioTextAlignLeft           = widget.TextAlignLeft
	cupRadioFocusRingOffset float32 = 3
	cupRadioFocusRingStroke float32 = 2.5
	cupRadioFocusAlpha      float32 = 0.6
	cupRadioHoverLighten    float32 = 0.08
	cupRadioPressedDarken   float32 = 0.12
)

// Compile-time check that RadioPainter implements Painter.
var _ radio.Painter = RadioPainter{}
