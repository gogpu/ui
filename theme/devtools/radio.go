package devtools

import (
	"github.com/gogpu/ui/core/radio"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// RadioPainter renders radio items using DevTools design tokens.
// DevTools radio buttons are 16px diameter circles with a 6px inner dot
// when selected, matching JetBrains IDE settings panel radio styling.
//
// If Theme is nil, RadioPainter falls back to the default DevTools dark palette.
type RadioPainter struct {
	Theme *Theme // nil uses default DevTools dark fallback
}

// resolveColors returns the RadioColorScheme derived from the painter's Theme.
func (p RadioPainter) resolveColors() radio.RadioColorScheme {
	if p.Theme == nil {
		return dtDefaultRadioColors
	}
	cs := p.Theme.Colors
	return radio.RadioColorScheme{
		SelectedBg:       cs.Primary,
		SelectedFg:       cs.OnPrimary,
		UnselectedBorder: cs.BorderStrong,
		LabelColor:       cs.OnSurface,
		DisabledBg:       cs.ControlFill,
		DisabledFg:       cs.OnSurfaceDisabled,
		FocusRing:        cs.BorderFocus,
	}
}

// PaintRadio renders a radio item according to DevTools design specifications.
func (p RadioPainter) PaintRadio(canvas widget.Canvas, state radio.PaintState) {
	if state.Bounds.IsEmpty() {
		return
	}

	colors := state.ColorScheme
	if colors == (radio.RadioColorScheme{}) {
		colors = p.resolveColors()
	}

	disabled := state.Disabled
	center, radius := dtRadioCircleGeometry(state.Bounds)

	if state.Selected {
		dtPaintSelectedRadio(canvas, center, radius, state, disabled, colors)
	} else {
		dtPaintUnselectedRadio(canvas, center, radius, state, disabled, colors)
	}

	// Label.
	if state.Label != "" {
		fg := colors.LabelColor
		if disabled {
			fg = colors.DisabledFg
		}
		labelBounds := dtRadioLabelBounds(state.Bounds)
		canvas.DrawText(state.Label, labelBounds, dtRadioFontSize, fg, false, dtRadioTextAlignLeft)
	}

	// Focus ring.
	if state.Focused && !disabled {
		canvas.StrokeCircle(center, radius+dtRadioFocusRingOffset, colors.FocusRing, dtRadioFocusRingStrokeWidth)
	}
}

// dtPaintSelectedRadio draws a selected radio item.
// DevTools style: accent border circle + accent inner dot.
func dtPaintSelectedRadio(canvas widget.Canvas, center geometry.Point, radius float32, state radio.PaintState, disabled bool, colors radio.RadioColorScheme) {
	// Outer ring with accent color.
	borderColor := colors.SelectedBg
	if disabled {
		borderColor = colors.DisabledBg
	} else {
		borderColor = dtApplyState(borderColor, state.Hovered, state.Pressed)
	}
	canvas.StrokeCircle(center, radius, borderColor, dtRadioBorderWidth)

	// Inner dot.
	fg := colors.SelectedBg
	if disabled {
		fg = colors.DisabledFg
	}
	canvas.DrawCircle(center, dtRadioInnerRadius, fg)
}

// dtPaintUnselectedRadio draws an unselected radio item.
func dtPaintUnselectedRadio(canvas widget.Canvas, center geometry.Point, radius float32, state radio.PaintState, disabled bool, colors radio.RadioColorScheme) {
	borderColor := colors.UnselectedBorder
	if disabled {
		borderColor = colors.DisabledFg
	} else {
		borderColor = dtApplyState(borderColor, state.Hovered, state.Pressed)
	}
	canvas.StrokeCircle(center, radius, borderColor, dtRadioBorderWidth)
}

// dtRadioCircleGeometry returns the center and radius for the outer radio circle.
func dtRadioCircleGeometry(bounds geometry.Rect) (geometry.Point, float32) {
	h := bounds.Height()
	cx := bounds.Min.X + dtRadioOuterRadius
	cy := bounds.Min.Y + h/2
	return geometry.Pt(cx, cy), dtRadioOuterRadius
}

// dtRadioLabelBounds returns the label area.
func dtRadioLabelBounds(bounds geometry.Rect) geometry.Rect {
	labelX := bounds.Min.X + dtRadioOuterRadius*2 + dtRadioLabelGap
	labelW := bounds.Width() - dtRadioOuterRadius*2 - dtRadioLabelGap
	if labelW < 0 {
		labelW = 0
	}
	return geometry.NewRect(labelX, bounds.Min.Y, labelW, bounds.Height())
}

// dtDefaultRadioColors holds the default DevTools dark radio color scheme.
var dtDefaultRadioColors = radio.RadioColorScheme{
	SelectedBg:       DefaultAccentColor,
	SelectedFg:       widget.ColorWhite,
	UnselectedBorder: widget.Hex(0x4E5157), // Gray5
	LabelColor:       widget.Hex(0xDFE1E5), // Gray12
	DisabledBg:       widget.Hex(0x393B40), // Gray3
	DisabledFg:       widget.Hex(0x6F737A), // Gray7
	FocusRing:        DefaultAccentColor,
}

// DevTools radio drawing constants.
const (
	dtRadioOuterRadius          float32 = 8
	dtRadioInnerRadius          float32 = 3
	dtRadioLabelGap             float32 = 8
	dtRadioBorderWidth          float32 = 1.5
	dtRadioFontSize             float32 = 13
	dtRadioTextAlignLeft                = widget.TextAlignLeft
	dtRadioFocusRingOffset      float32 = 2
	dtRadioFocusRingStrokeWidth float32 = 1
)

// Compile-time check that RadioPainter implements Painter.
var _ radio.Painter = RadioPainter{}
