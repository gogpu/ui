package material3

import (
	"github.com/gogpu/ui/core/checkbox"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// CheckboxPainter renders checkboxes using Material 3 design tokens.
// It maps checkbox states (checked, unchecked, indeterminate) to
// the M3 color scheme and applies appropriate interaction feedback.
//
// If Theme is nil, CheckboxPainter falls back to the default M3 purple palette.
type CheckboxPainter struct {
	Theme *Theme // nil uses default M3 purple fallback
}

// resolveColors returns the CheckboxColorScheme derived from the painter's Theme.
// If Theme is nil, it returns the default M3 purple color scheme.
func (p CheckboxPainter) resolveColors() checkbox.CheckboxColorScheme {
	if p.Theme == nil {
		return m3DefaultCheckboxColors
	}
	cs := p.Theme.Colors
	return checkbox.CheckboxColorScheme{
		CheckedBg:       cs.Primary,
		CheckedFg:       cs.OnPrimary,
		UncheckedBorder: cs.Outline,
		LabelColor:      cs.OnSurface,
		DisabledBg:      cs.OnSurface.WithAlpha(0.12),
		DisabledFg:      cs.OnSurface.WithAlpha(0.38),
		FocusRing:       cs.Primary.WithAlpha(0.7),
	}
}

// PaintCheckbox renders a checkbox according to Material 3 specifications.
func (p CheckboxPainter) PaintCheckbox(canvas widget.Canvas, state checkbox.PaintState) {
	if state.Bounds.IsEmpty() {
		return
	}

	// Determine the color scheme to use.
	colors := state.ColorScheme
	if colors == (checkbox.CheckboxColorScheme{}) {
		colors = p.resolveColors()
	}

	boxRect := m3CheckboxBoxRect(state.Bounds)
	disabled := state.Disabled

	if state.Checked || state.Indeterminate {
		m3PaintCheckedBox(canvas, boxRect, state, disabled, colors)
	} else {
		m3PaintUncheckedBox(canvas, boxRect, state, disabled, colors)
	}

	// Draw label if present.
	if state.Label != "" {
		fg := colors.LabelColor
		if disabled {
			fg = colors.DisabledFg
		}
		labelBounds := m3CheckboxLabelBounds(state.Bounds)
		canvas.DrawText(state.Label, labelBounds, m3CheckboxFontSize, fg, false, m3CheckboxTextAlignLeft)
	}

	// Draw focus ring when focused.
	if state.Focused && !disabled {
		m3DrawCheckboxFocusIndicator(canvas, boxRect, colors)
	}
}

// m3PaintCheckedBox draws the checkbox in checked or indeterminate state with M3 colors.
func m3PaintCheckedBox(canvas widget.Canvas, boxRect geometry.Rect, state checkbox.PaintState, disabled bool, colors checkbox.CheckboxColorScheme) {
	bg := colors.CheckedBg
	if state.Background != nil {
		bg = *state.Background
	}
	if disabled {
		bg = colors.DisabledBg
	} else {
		bg = m3ApplyCheckboxState(bg, state.Hovered, state.Pressed)
	}
	canvas.DrawRoundRect(boxRect, bg, m3CheckboxCornerRadius)

	fg := colors.CheckedFg
	if disabled {
		fg = colors.DisabledFg
	}

	if state.Indeterminate {
		m3DrawDash(canvas, boxRect, fg)
	} else {
		m3DrawCheckmark(canvas, boxRect, fg)
	}
}

// m3PaintUncheckedBox draws the checkbox in unchecked state with M3 colors.
func m3PaintUncheckedBox(canvas widget.Canvas, boxRect geometry.Rect, state checkbox.PaintState, disabled bool, colors checkbox.CheckboxColorScheme) {
	borderColor := colors.UncheckedBorder
	if disabled {
		borderColor = colors.DisabledFg
	} else {
		borderColor = m3ApplyCheckboxState(borderColor, state.Hovered, state.Pressed)
	}
	canvas.StrokeRoundRect(boxRect, borderColor, m3CheckboxCornerRadius, m3CheckboxBorderWidth)
}

// m3ApplyCheckboxState adjusts a color based on interaction state.
func m3ApplyCheckboxState(base widget.Color, hovered, pressed bool) widget.Color {
	if pressed {
		return base.Lerp(widget.ColorBlack, m3CheckboxPressedDarkenFactor)
	}
	if hovered {
		return base.Lerp(widget.ColorWhite, m3CheckboxHoverLightenFactor)
	}
	return base
}

// m3DrawCheckmark draws a checkmark inside the box.
func m3DrawCheckmark(canvas widget.Canvas, box geometry.Rect, color widget.Color) {
	cx := box.Min.X
	cy := box.Min.Y
	s := box.Width()

	p1 := geometry.Pt(cx+s*m3CheckP1X, cy+s*m3CheckP1Y)
	p2 := geometry.Pt(cx+s*m3CheckP2X, cy+s*m3CheckP2Y)
	p3 := geometry.Pt(cx+s*m3CheckP3X, cy+s*m3CheckP3Y)

	canvas.DrawLine(p1, p2, color, m3CheckStrokeWidth)
	canvas.DrawLine(p2, p3, color, m3CheckStrokeWidth)
}

// m3DrawDash draws a horizontal dash inside the box for indeterminate state.
func m3DrawDash(canvas widget.Canvas, box geometry.Rect, color widget.Color) {
	cx := box.Min.X
	cy := box.Min.Y
	s := box.Width()
	midY := cy + s*m3DashMidY

	from := geometry.Pt(cx+s*m3DashStartX, midY)
	to := geometry.Pt(cx+s*m3DashEndX, midY)

	canvas.DrawLine(from, to, color, m3DashStrokeWidth)
}

// m3DrawCheckboxFocusIndicator draws a focus ring around the checkbox box.
func m3DrawCheckboxFocusIndicator(canvas widget.Canvas, boxRect geometry.Rect, colors checkbox.CheckboxColorScheme) {
	ringBounds := boxRect.Expand(m3CheckboxFocusRingOffset)
	ringRadius := m3CheckboxCornerRadius + m3CheckboxFocusRingOffset
	canvas.StrokeRoundRect(ringBounds, colors.FocusRing, ringRadius, m3CheckboxFocusRingStrokeWidth)
}

// m3CheckboxBoxRect returns the square box area on the left side of the bounds.
// Inset by half the border width so centered strokes stay within bounds (#117).
func m3CheckboxBoxRect(bounds geometry.Rect) geometry.Rect {
	h := bounds.Height()
	boxY := bounds.Min.Y + (h-m3CheckboxBoxSize)/2
	return geometry.NewRect(bounds.Min.X+m3CheckboxBorderWidth/2, boxY, m3CheckboxBoxSize, m3CheckboxBoxSize)
}

// m3CheckboxLabelBounds returns the label area to the right of the checkbox box.
func m3CheckboxLabelBounds(bounds geometry.Rect) geometry.Rect {
	offset := m3CheckboxBorderWidth / 2
	return geometry.NewRect(
		bounds.Min.X+offset+m3CheckboxBoxSize+m3CheckboxLabelGap,
		bounds.Min.Y,
		bounds.Width()-offset-m3CheckboxBoxSize-m3CheckboxLabelGap,
		bounds.Height(),
	)
}

// m3DefaultCheckboxColors holds the default M3 purple color scheme for checkboxes.
// Used as a fallback when no Theme is provided.
var m3DefaultCheckboxColors = checkbox.CheckboxColorScheme{
	CheckedBg:       widget.Hex(0x6750A4),                // M3 primary
	CheckedFg:       widget.ColorWhite,                   // M3 on-primary
	UncheckedBorder: widget.Hex(0x79747E),                // M3 outline
	LabelColor:      widget.Hex(0x1C1B1F),                // M3 on-surface
	DisabledBg:      widget.RGBA(0.12, 0.12, 0.13, 0.12), // M3 disabled bg
	DisabledFg:      widget.RGBA(0.12, 0.12, 0.13, 0.38), // M3 disabled fg
	FocusRing:       widget.Hex(0x6750A4).WithAlpha(0.7), // M3 focus ring
}

// M3 checkbox drawing constants.
const (
	m3CheckboxBoxSize              float32 = 18
	m3CheckboxCornerRadius         float32 = 3
	m3CheckboxLabelGap             float32 = 8
	m3CheckboxBorderWidth          float32 = 2
	m3CheckboxFontSize             float32 = 14
	m3CheckboxTextAlignLeft                = widget.TextAlignLeft
	m3CheckboxFocusRingOffset      float32 = 2
	m3CheckboxFocusRingStrokeWidth float32 = 2
	m3CheckboxHoverLightenFactor   float32 = 0.1
	m3CheckboxPressedDarkenFactor  float32 = 0.15

	// Checkmark geometry (relative to box size 0..1).
	m3CheckP1X         float32 = 0.2
	m3CheckP1Y         float32 = 0.5
	m3CheckP2X         float32 = 0.4
	m3CheckP2Y         float32 = 0.7
	m3CheckP3X         float32 = 0.8
	m3CheckP3Y         float32 = 0.3
	m3CheckStrokeWidth float32 = 2

	// Dash geometry (relative to box size 0..1).
	m3DashStartX      float32 = 0.25
	m3DashEndX        float32 = 0.75
	m3DashMidY        float32 = 0.5
	m3DashStrokeWidth float32 = 2
)

// CheckboxBoxSize returns the M3 checkbox box size.
func (CheckboxPainter) CheckboxBoxSize() float32 { return m3CheckboxBoxSize }

// CheckboxLabelGap returns the M3 gap between box and label.
func (CheckboxPainter) CheckboxLabelGap() float32 { return m3CheckboxLabelGap }

// CheckboxFontSize returns the M3 label font size.
func (CheckboxPainter) CheckboxFontSize() float32 { return m3CheckboxFontSize }

// Compile-time checks.
var (
	_ checkbox.Painter       = CheckboxPainter{}
	_ checkbox.LayoutMetrics = CheckboxPainter{}
)
