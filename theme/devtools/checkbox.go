package devtools

import (
	"github.com/gogpu/ui/core/checkbox"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// CheckboxPainter renders checkboxes using DevTools design tokens.
// DevTools checkboxes are 16px with 2px corner radius, matching JetBrains
// IDE settings panel checkbox styling.
//
// If Theme is nil, CheckboxPainter falls back to the default DevTools dark palette.
type CheckboxPainter struct {
	Theme *Theme // nil uses default DevTools dark fallback
}

// resolveColors returns the CheckboxColorScheme derived from the painter's Theme.
func (p CheckboxPainter) resolveColors() checkbox.CheckboxColorScheme {
	if p.Theme == nil {
		return dtDefaultCheckboxColors
	}
	cs := p.Theme.Colors
	return checkbox.CheckboxColorScheme{
		CheckedBg:       cs.Primary,
		CheckedFg:       cs.OnPrimary,
		UncheckedBorder: cs.BorderStrong,
		LabelColor:      cs.OnSurface,
		DisabledBg:      cs.ControlFill,
		DisabledFg:      cs.OnSurfaceDisabled,
		FocusRing:       cs.BorderFocus,
	}
}

// PaintCheckbox renders a checkbox according to DevTools design specifications.
func (p CheckboxPainter) PaintCheckbox(canvas widget.Canvas, state checkbox.PaintState) {
	if state.Bounds.IsEmpty() {
		return
	}

	colors := state.ColorScheme
	if colors == (checkbox.CheckboxColorScheme{}) {
		colors = p.resolveColors()
	}

	boxRect := dtCheckboxBoxRect(state.Bounds)
	disabled := state.Disabled

	if state.Checked || state.Indeterminate {
		dtPaintCheckedBox(canvas, boxRect, state, disabled, colors)
	} else {
		dtPaintUncheckedBox(canvas, boxRect, state, disabled, colors)
	}

	// Label.
	if state.Label != "" {
		fg := colors.LabelColor
		if disabled {
			fg = colors.DisabledFg
		}
		labelBounds := dtCheckboxLabelBounds(state.Bounds)
		canvas.DrawText(state.Label, labelBounds, dtCBFontSize, fg, false, dtCBTextAlignLeft)
	}

	// Focus indicator.
	if state.Focused && !disabled {
		dtDrawFocusRing(canvas, boxRect, dtCBCornerRadius, colors.FocusRing)
	}
}

// dtPaintCheckedBox draws the checkbox in checked/indeterminate state.
func dtPaintCheckedBox(canvas widget.Canvas, boxRect geometry.Rect, state checkbox.PaintState, disabled bool, colors checkbox.CheckboxColorScheme) {
	bg := colors.CheckedBg
	if state.Background != nil {
		bg = *state.Background
	}
	if disabled {
		bg = colors.DisabledBg
	} else {
		bg = dtApplyState(bg, state.Hovered, state.Pressed)
	}
	canvas.DrawRoundRect(boxRect, bg, dtCBCornerRadius)

	fg := colors.CheckedFg
	if disabled {
		fg = colors.DisabledFg
	}

	if state.Indeterminate {
		dtDrawDash(canvas, boxRect, fg)
	} else {
		dtDrawCheckmark(canvas, boxRect, fg)
	}
}

// dtPaintUncheckedBox draws the checkbox in unchecked state.
func dtPaintUncheckedBox(canvas widget.Canvas, boxRect geometry.Rect, state checkbox.PaintState, disabled bool, colors checkbox.CheckboxColorScheme) {
	borderColor := colors.UncheckedBorder
	if disabled {
		borderColor = colors.DisabledFg
	} else {
		borderColor = dtApplyState(borderColor, state.Hovered, state.Pressed)
	}
	canvas.StrokeRoundRect(boxRect, borderColor, dtCBCornerRadius, dtCBBorderWidth)
}

// dtDrawCheckmark draws a checkmark inside the box.
func dtDrawCheckmark(canvas widget.Canvas, box geometry.Rect, color widget.Color) {
	cx := box.Min.X
	cy := box.Min.Y
	s := box.Width()

	p1 := geometry.Pt(cx+s*dtCheckP1X, cy+s*dtCheckP1Y)
	p2 := geometry.Pt(cx+s*dtCheckP2X, cy+s*dtCheckP2Y)
	p3 := geometry.Pt(cx+s*dtCheckP3X, cy+s*dtCheckP3Y)

	canvas.DrawLine(p1, p2, color, dtCheckStrokeWidth)
	canvas.DrawLine(p2, p3, color, dtCheckStrokeWidth)
}

// dtDrawDash draws a horizontal dash for indeterminate state.
func dtDrawDash(canvas widget.Canvas, box geometry.Rect, color widget.Color) {
	cx := box.Min.X
	cy := box.Min.Y
	s := box.Width()
	midY := cy + s*dtDashMidY

	from := geometry.Pt(cx+s*dtDashStartX, midY)
	to := geometry.Pt(cx+s*dtDashEndX, midY)

	canvas.DrawLine(from, to, color, dtDashStrokeWidth)
}

// dtCheckboxBoxRect returns the checkbox box area.
// Inset by half the border width so centered strokes stay within bounds (#117).
func dtCheckboxBoxRect(bounds geometry.Rect) geometry.Rect {
	h := bounds.Height()
	boxY := bounds.Min.Y + (h-dtCBBoxSize)/2
	return geometry.NewRect(bounds.Min.X+dtCBBorderWidth/2, boxY, dtCBBoxSize, dtCBBoxSize)
}

// dtCheckboxLabelBounds returns the label area.
func dtCheckboxLabelBounds(bounds geometry.Rect) geometry.Rect {
	offset := dtCBBorderWidth / 2
	return geometry.NewRect(
		bounds.Min.X+offset+dtCBBoxSize+dtCBLabelGap,
		bounds.Min.Y,
		bounds.Width()-offset-dtCBBoxSize-dtCBLabelGap,
		bounds.Height(),
	)
}

// dtDefaultCheckboxColors holds the default DevTools dark checkbox color scheme.
var dtDefaultCheckboxColors = checkbox.CheckboxColorScheme{
	CheckedBg:       DefaultAccentColor,
	CheckedFg:       widget.ColorWhite,
	UncheckedBorder: widget.Hex(0x4E5157), // Gray5
	LabelColor:      widget.Hex(0xDFE1E5), // Gray12
	DisabledBg:      widget.Hex(0x393B40), // Gray3
	DisabledFg:      widget.Hex(0x6F737A), // Gray7
	FocusRing:       DefaultAccentColor,
}

// DevTools checkbox drawing constants.
const (
	dtCBBoxSize       float32 = 16
	dtCBCornerRadius  float32 = 2
	dtCBLabelGap      float32 = 8
	dtCBBorderWidth   float32 = 1.5
	dtCBFontSize      float32 = 13
	dtCBTextAlignLeft         = widget.TextAlignLeft

	// Checkmark geometry (relative to box size 0..1).
	dtCheckP1X         float32 = 0.2
	dtCheckP1Y         float32 = 0.5
	dtCheckP2X         float32 = 0.4
	dtCheckP2Y         float32 = 0.7
	dtCheckP3X         float32 = 0.8
	dtCheckP3Y         float32 = 0.3
	dtCheckStrokeWidth float32 = 1.5

	// Dash geometry.
	dtDashStartX      float32 = 0.25
	dtDashEndX        float32 = 0.75
	dtDashMidY        float32 = 0.5
	dtDashStrokeWidth float32 = 1.5
)

// CheckboxBoxSize returns the DevTools checkbox box size.
func (CheckboxPainter) CheckboxBoxSize() float32 { return dtCBBoxSize }

// CheckboxLabelGap returns the DevTools gap between box and label.
func (CheckboxPainter) CheckboxLabelGap() float32 { return dtCBLabelGap }

// CheckboxFontSize returns the DevTools label font size.
func (CheckboxPainter) CheckboxFontSize() float32 { return dtCBFontSize }

// Compile-time checks.
var (
	_ checkbox.Painter       = CheckboxPainter{}
	_ checkbox.LayoutMetrics = CheckboxPainter{}
)
