package material3

import (
	"github.com/gogpu/ui/core/radio"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// RadioPainter renders radio items using Material 3 design tokens.
// It maps radio states (selected, unselected) to the M3 color scheme
// and applies appropriate interaction feedback.
//
// If Theme is nil, RadioPainter falls back to the default M3 purple palette.
type RadioPainter struct {
	Theme *Theme // nil uses default M3 purple fallback
}

// resolveColors returns the RadioColorScheme derived from the painter's Theme.
// If Theme is nil, it returns the default M3 purple color scheme.
func (p RadioPainter) resolveColors() radio.RadioColorScheme {
	if p.Theme == nil {
		return m3DefaultRadioColors
	}
	cs := p.Theme.Colors
	return radio.RadioColorScheme{
		SelectedBg:       cs.Primary,
		SelectedFg:       cs.OnPrimary,
		UnselectedBorder: cs.Outline,
		LabelColor:       cs.OnSurface,
		DisabledBg:       cs.OnSurface.WithAlpha(0.12),
		DisabledFg:       cs.OnSurface.WithAlpha(0.38),
		FocusRing:        cs.Primary.WithAlpha(0.7),
	}
}

// PaintRadio renders a radio item according to Material 3 specifications.
func (p RadioPainter) PaintRadio(canvas widget.Canvas, state radio.PaintState) {
	if state.Bounds.IsEmpty() {
		return
	}

	// Determine the color scheme to use.
	colors := state.ColorScheme
	if colors == (radio.RadioColorScheme{}) {
		colors = p.resolveColors()
	}

	disabled := state.Disabled
	center, radius := m3RadioCircleGeometry(state.Bounds)

	if state.Selected {
		m3PaintSelectedRadio(canvas, center, radius, state, disabled, colors)
	} else {
		m3PaintUnselectedRadio(canvas, center, radius, state, disabled, colors)
	}

	// Draw label if present.
	if state.Label != "" {
		fg := colors.LabelColor
		if disabled {
			fg = colors.DisabledFg
		}
		labelBounds := m3RadioLabelBounds(state.Bounds)
		canvas.DrawText(state.Label, labelBounds, m3RadioFontSize, fg, false, m3RadioTextAlignLeft)
	}

	// Draw focus ring when focused.
	if state.Focused && !disabled {
		m3DrawRadioFocusIndicator(canvas, center, radius, colors)
	}
}

// m3PaintSelectedRadio draws the radio item in selected state with M3 colors.
func m3PaintSelectedRadio(canvas widget.Canvas, center geometry.Point, radius float32, state radio.PaintState, disabled bool, colors radio.RadioColorScheme) {
	bg := colors.SelectedBg
	if disabled {
		bg = colors.DisabledBg
	} else {
		bg = m3ApplyRadioState(bg, state.Hovered, state.Pressed)
	}
	canvas.DrawCircle(center, radius, bg)

	fg := colors.SelectedFg
	if disabled {
		fg = colors.DisabledFg
	}
	canvas.DrawCircle(center, m3RadioInnerRadius, fg)
}

// m3PaintUnselectedRadio draws the radio item in unselected state with M3 colors.
func m3PaintUnselectedRadio(canvas widget.Canvas, center geometry.Point, radius float32, state radio.PaintState, disabled bool, colors radio.RadioColorScheme) {
	borderColor := colors.UnselectedBorder
	if disabled {
		borderColor = colors.DisabledFg
	} else {
		borderColor = m3ApplyRadioState(borderColor, state.Hovered, state.Pressed)
	}
	canvas.StrokeCircle(center, radius, borderColor, m3RadioBorderWidth)
}

// m3ApplyRadioState adjusts a color based on interaction state.
func m3ApplyRadioState(base widget.Color, hovered, pressed bool) widget.Color {
	if pressed {
		return base.Lerp(widget.ColorBlack, m3RadioPressedDarkenFactor)
	}
	if hovered {
		return base.Lerp(widget.ColorWhite, m3RadioHoverLightenFactor)
	}
	return base
}

// m3DrawRadioFocusIndicator draws a focus ring around the radio circle.
func m3DrawRadioFocusIndicator(canvas widget.Canvas, center geometry.Point, radius float32, colors radio.RadioColorScheme) {
	canvas.StrokeCircle(center, radius+m3RadioFocusRingOffset, colors.FocusRing, m3RadioFocusRingStrokeWidth)
}

// m3RadioCircleGeometry returns the center point and radius for the outer radio circle.
func m3RadioCircleGeometry(bounds geometry.Rect) (geometry.Point, float32) {
	h := bounds.Height()
	cx := bounds.Min.X + m3RadioOuterRadius
	cy := bounds.Min.Y + h/2
	return geometry.Pt(cx, cy), m3RadioOuterRadius
}

// m3RadioLabelBounds returns the label area to the right of the radio circle.
func m3RadioLabelBounds(bounds geometry.Rect) geometry.Rect {
	labelX := bounds.Min.X + m3RadioOuterRadius*2 + m3RadioLabelGap
	labelW := bounds.Width() - m3RadioOuterRadius*2 - m3RadioLabelGap
	if labelW < 0 {
		labelW = 0
	}
	return geometry.NewRect(labelX, bounds.Min.Y, labelW, bounds.Height())
}

// m3DefaultRadioColors holds the default M3 purple color scheme for radio items.
// Used as a fallback when no Theme is provided.
var m3DefaultRadioColors = radio.RadioColorScheme{
	SelectedBg:       widget.Hex(0x6750A4),                // M3 primary
	SelectedFg:       widget.ColorWhite,                   // M3 on-primary
	UnselectedBorder: widget.Hex(0x79747E),                // M3 outline
	LabelColor:       widget.Hex(0x1C1B1F),                // M3 on-surface
	DisabledBg:       widget.RGBA(0.12, 0.12, 0.13, 0.12), // M3 disabled bg
	DisabledFg:       widget.RGBA(0.12, 0.12, 0.13, 0.38), // M3 disabled fg
	FocusRing:        widget.Hex(0x6750A4).WithAlpha(0.7), // M3 focus ring
}

// M3 radio drawing constants.
const (
	m3RadioOuterRadius          float32 = 9
	m3RadioInnerRadius          float32 = 4.5
	m3RadioLabelGap             float32 = 8
	m3RadioBorderWidth          float32 = 2
	m3RadioFontSize             float32 = 14
	m3RadioTextAlignLeft                = widget.TextAlignLeft
	m3RadioFocusRingOffset      float32 = 2
	m3RadioFocusRingStrokeWidth float32 = 2
	m3RadioHoverLightenFactor   float32 = 0.1
	m3RadioPressedDarkenFactor  float32 = 0.15
	m3RadioItemPadding          float32 = 4
)

// RadioCircleRadius returns the M3 outer circle radius.
func (RadioPainter) RadioCircleRadius() float32 { return m3RadioOuterRadius }

// RadioLabelGap returns the M3 gap between circle and label.
func (RadioPainter) RadioLabelGap() float32 { return m3RadioLabelGap }

// RadioFontSize returns the M3 label font size.
func (RadioPainter) RadioFontSize() float32 { return m3RadioFontSize }

// RadioItemPadding returns the M3 radio item padding.
func (RadioPainter) RadioItemPadding() float32 { return m3RadioItemPadding }

// Compile-time checks.
var (
	_ radio.Painter       = RadioPainter{}
	_ radio.LayoutMetrics = RadioPainter{}
)
