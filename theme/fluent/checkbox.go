package fluent

import (
	"github.com/gogpu/ui/core/checkbox"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// CheckboxPainter renders checkboxes using Fluent Design tokens.
//
// If Theme is nil, CheckboxPainter falls back to the default Fluent Blue palette.
type CheckboxPainter struct {
	Theme *Theme // nil uses default Fluent Blue fallback
}

// resolveColors returns the CheckboxColorScheme derived from the painter's Theme.
func (p CheckboxPainter) resolveColors() checkbox.CheckboxColorScheme {
	if p.Theme == nil {
		return flDefaultCheckboxColors
	}
	cs := p.Theme.Colors
	return checkbox.CheckboxColorScheme{
		CheckedBg:       cs.Accent,
		CheckedFg:       cs.OnAccent,
		UncheckedBorder: cs.StrokeDefault,
		LabelColor:      cs.OnSurface,
		DisabledBg:      cs.FillDisable,
		DisabledFg:      cs.OnSurfaceSecond.WithAlpha(flDisabledAlpha),
		FocusRing:       cs.StrokeFocus,
	}
}

// PaintCheckbox renders a checkbox according to Fluent Design specifications.
func (p CheckboxPainter) PaintCheckbox(canvas widget.Canvas, state checkbox.PaintState) {
	if state.Bounds.IsEmpty() {
		return
	}

	colors := state.ColorScheme
	if colors == (checkbox.CheckboxColorScheme{}) {
		colors = p.resolveColors()
	}

	boxRect := flCheckboxBoxRect(state.Bounds)
	disabled := state.Disabled

	if state.Checked || state.Indeterminate {
		flPaintCheckedBox(canvas, boxRect, state, disabled, colors)
	} else {
		flPaintUncheckedBox(canvas, boxRect, state, disabled, colors)
	}

	// Label.
	if state.Label != "" {
		fg := colors.LabelColor
		if disabled {
			fg = colors.DisabledFg
		}
		labelBounds := flCheckboxLabelBounds(state.Bounds)
		canvas.DrawText(state.Label, labelBounds, flCBFontSize, fg, false, flCBTextAlignLeft)
	}

	// Focus indicator.
	if state.Focused && !disabled {
		flDrawFocusRing(canvas, boxRect, flCBCornerRadius, colors.FocusRing)
	}
}

// flPaintCheckedBox draws the checkbox in checked/indeterminate state.
func flPaintCheckedBox(canvas widget.Canvas, boxRect geometry.Rect, state checkbox.PaintState, disabled bool, colors checkbox.CheckboxColorScheme) {
	bg := colors.CheckedBg
	if state.Background != nil {
		bg = *state.Background
	}
	if disabled {
		bg = colors.DisabledBg
	} else {
		bg = flApplyState(bg, state.Hovered, state.Pressed)
	}
	canvas.DrawRoundRect(boxRect, bg, flCBCornerRadius)

	fg := colors.CheckedFg
	if disabled {
		fg = colors.DisabledFg
	}

	if state.Indeterminate {
		flDrawDash(canvas, boxRect, fg)
	} else {
		flDrawCheckmark(canvas, boxRect, fg)
	}
}

// flPaintUncheckedBox draws the checkbox in unchecked state.
func flPaintUncheckedBox(canvas widget.Canvas, boxRect geometry.Rect, state checkbox.PaintState, disabled bool, colors checkbox.CheckboxColorScheme) {
	borderColor := colors.UncheckedBorder
	if disabled {
		borderColor = colors.DisabledFg
	} else {
		borderColor = flApplyState(borderColor, state.Hovered, state.Pressed)
	}
	canvas.StrokeRoundRect(boxRect, borderColor, flCBCornerRadius, flCBBorderWidth)
}

// flDrawCheckmark draws a checkmark inside the box.
func flDrawCheckmark(canvas widget.Canvas, box geometry.Rect, color widget.Color) {
	cx := box.Min.X
	cy := box.Min.Y
	s := box.Width()

	p1 := geometry.Pt(cx+s*flCheckP1X, cy+s*flCheckP1Y)
	p2 := geometry.Pt(cx+s*flCheckP2X, cy+s*flCheckP2Y)
	p3 := geometry.Pt(cx+s*flCheckP3X, cy+s*flCheckP3Y)

	canvas.DrawLine(p1, p2, color, flCheckStrokeWidth)
	canvas.DrawLine(p2, p3, color, flCheckStrokeWidth)
}

// flDrawDash draws a horizontal dash for indeterminate state.
func flDrawDash(canvas widget.Canvas, box geometry.Rect, color widget.Color) {
	cx := box.Min.X
	cy := box.Min.Y
	s := box.Width()
	midY := cy + s*flDashMidY

	from := geometry.Pt(cx+s*flDashStartX, midY)
	to := geometry.Pt(cx+s*flDashEndX, midY)

	canvas.DrawLine(from, to, color, flDashStrokeWidth)
}

// flCheckboxBoxRect returns the checkbox box area.
// Inset by half the border width so centered strokes stay within bounds (#117).
func flCheckboxBoxRect(bounds geometry.Rect) geometry.Rect {
	h := bounds.Height()
	boxY := bounds.Min.Y + (h-flCBBoxSize)/2
	return geometry.NewRect(bounds.Min.X+flCBBorderWidth/2, boxY, flCBBoxSize, flCBBoxSize)
}

// flCheckboxLabelBounds returns the label area.
func flCheckboxLabelBounds(bounds geometry.Rect) geometry.Rect {
	offset := flCBBorderWidth / 2
	return geometry.NewRect(
		bounds.Min.X+offset+flCBBoxSize+flCBLabelGap,
		bounds.Min.Y,
		bounds.Width()-offset-flCBBoxSize-flCBLabelGap,
		bounds.Height(),
	)
}

// flDefaultCheckboxColors holds the default Fluent Blue checkbox color scheme.
var flDefaultCheckboxColors = checkbox.CheckboxColorScheme{
	CheckedBg:       DefaultAccentColor,
	CheckedFg:       widget.ColorWhite,
	UncheckedBorder: widget.RGBA(0, 0, 0, 0.14),
	LabelColor:      widget.Hex(0x1A1A1A),
	DisabledBg:      widget.RGBA(0, 0, 0, 0.04),
	DisabledFg:      widget.RGBA(0.38, 0.38, 0.38, 0.38),
	FocusRing:       DefaultAccentColor,
}

// Fluent checkbox drawing constants.
const (
	flCBBoxSize       float32 = 18
	flCBCornerRadius  float32 = 3
	flCBLabelGap      float32 = 8
	flCBBorderWidth   float32 = 1.5
	flCBFontSize      float32 = 14
	flCBTextAlignLeft         = widget.TextAlignLeft

	// Checkmark geometry (relative to box size 0..1).
	flCheckP1X         float32 = 0.2
	flCheckP1Y         float32 = 0.5
	flCheckP2X         float32 = 0.4
	flCheckP2Y         float32 = 0.7
	flCheckP3X         float32 = 0.8
	flCheckP3Y         float32 = 0.3
	flCheckStrokeWidth float32 = 2

	// Dash geometry.
	flDashStartX      float32 = 0.25
	flDashEndX        float32 = 0.75
	flDashMidY        float32 = 0.5
	flDashStrokeWidth float32 = 2
)

// CheckboxBoxSize returns the Fluent checkbox box size.
func (CheckboxPainter) CheckboxBoxSize() float32 { return flCBBoxSize }

// CheckboxLabelGap returns the Fluent gap between box and label.
func (CheckboxPainter) CheckboxLabelGap() float32 { return flCBLabelGap }

// CheckboxFontSize returns the Fluent label font size.
func (CheckboxPainter) CheckboxFontSize() float32 { return flCBFontSize }

// Compile-time checks.
var (
	_ checkbox.Painter       = CheckboxPainter{}
	_ checkbox.LayoutMetrics = CheckboxPainter{}
)
