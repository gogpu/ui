package fluent

import (
	"github.com/gogpu/ui/core/radio"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// RadioPainter renders radio items using Fluent Design tokens.
//
// If Theme is nil, RadioPainter falls back to the default Fluent Blue palette.
type RadioPainter struct {
	Theme *Theme // nil uses default Fluent Blue fallback
}

// resolveColors returns the RadioColorScheme derived from the painter's Theme.
func (p RadioPainter) resolveColors() radio.RadioColorScheme {
	if p.Theme == nil {
		return flDefaultRadioColors
	}
	cs := p.Theme.Colors
	return radio.RadioColorScheme{
		SelectedBg:       cs.Accent,
		SelectedFg:       cs.OnAccent,
		UnselectedBorder: cs.StrokeDefault,
		LabelColor:       cs.OnSurface,
		DisabledBg:       cs.FillDisable,
		DisabledFg:       cs.OnSurfaceSecond.WithAlpha(flDisabledAlpha),
		FocusRing:        cs.StrokeFocus,
	}
}

// PaintRadio renders a radio item according to Fluent Design specifications.
func (p RadioPainter) PaintRadio(canvas widget.Canvas, state radio.PaintState) {
	if state.Bounds.IsEmpty() {
		return
	}

	colors := state.ColorScheme
	if colors == (radio.RadioColorScheme{}) {
		colors = p.resolveColors()
	}

	disabled := state.Disabled
	center, radius := flRadioCircleGeometry(state.Bounds)

	if state.Selected {
		flPaintSelectedRadio(canvas, center, radius, state, disabled, colors)
	} else {
		flPaintUnselectedRadio(canvas, center, radius, state, disabled, colors)
	}

	// Label.
	if state.Label != "" {
		fg := colors.LabelColor
		if disabled {
			fg = colors.DisabledFg
		}
		labelBounds := flRadioLabelBounds(state.Bounds)
		canvas.DrawText(state.Label, labelBounds, flRadioFontSize, fg, false, flRadioTextAlignLeft)
	}

	// Focus ring.
	if state.Focused && !disabled {
		canvas.StrokeCircle(center, radius+flRadioFocusRingOffset, colors.FocusRing, flRadioFocusRingStrokeWidth)
	}
}

// flPaintSelectedRadio draws a selected radio item.
func flPaintSelectedRadio(canvas widget.Canvas, center geometry.Point, radius float32, state radio.PaintState, disabled bool, colors radio.RadioColorScheme) {
	bg := colors.SelectedBg
	if disabled {
		bg = colors.DisabledBg
	} else {
		bg = flApplyState(bg, state.Hovered, state.Pressed)
	}
	canvas.DrawCircle(center, radius, bg)

	fg := colors.SelectedFg
	if disabled {
		fg = colors.DisabledFg
	}
	canvas.DrawCircle(center, flRadioInnerRadius, fg)
}

// flPaintUnselectedRadio draws an unselected radio item.
func flPaintUnselectedRadio(canvas widget.Canvas, center geometry.Point, radius float32, state radio.PaintState, disabled bool, colors radio.RadioColorScheme) {
	borderColor := colors.UnselectedBorder
	if disabled {
		borderColor = colors.DisabledFg
	} else {
		borderColor = flApplyState(borderColor, state.Hovered, state.Pressed)
	}
	canvas.StrokeCircle(center, radius, borderColor, flRadioBorderWidth)
}

// flRadioCircleGeometry returns the center and radius for the outer radio circle.
func flRadioCircleGeometry(bounds geometry.Rect) (geometry.Point, float32) {
	h := bounds.Height()
	cx := bounds.Min.X + flRadioOuterRadius
	cy := bounds.Min.Y + h/2
	return geometry.Pt(cx, cy), flRadioOuterRadius
}

// flRadioLabelBounds returns the label area.
func flRadioLabelBounds(bounds geometry.Rect) geometry.Rect {
	labelX := bounds.Min.X + flRadioOuterRadius*2 + flRadioLabelGap
	labelW := bounds.Width() - flRadioOuterRadius*2 - flRadioLabelGap
	if labelW < 0 {
		labelW = 0
	}
	return geometry.NewRect(labelX, bounds.Min.Y, labelW, bounds.Height())
}

// flDefaultRadioColors holds the default Fluent Blue radio color scheme.
var flDefaultRadioColors = radio.RadioColorScheme{
	SelectedBg:       DefaultAccentColor,
	SelectedFg:       widget.ColorWhite,
	UnselectedBorder: widget.RGBA(0, 0, 0, 0.14),
	LabelColor:       widget.Hex(0x1A1A1A),
	DisabledBg:       widget.RGBA(0, 0, 0, 0.04),
	DisabledFg:       widget.RGBA(0.38, 0.38, 0.38, 0.38),
	FocusRing:        DefaultAccentColor,
}

// Fluent radio drawing constants.
const (
	flRadioOuterRadius          float32 = 9
	flRadioInnerRadius          float32 = 4
	flRadioLabelGap             float32 = 8
	flRadioBorderWidth          float32 = 1.5
	flRadioFontSize             float32 = 14
	flRadioTextAlignLeft                = widget.TextAlignLeft
	flRadioFocusRingOffset      float32 = 2
	flRadioFocusRingStrokeWidth float32 = 2
)

// Compile-time check that RadioPainter implements Painter.
var _ radio.Painter = RadioPainter{}
