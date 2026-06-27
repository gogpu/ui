package cupertino

import (
	"github.com/gogpu/ui/core/checkbox"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// CheckboxPainter renders checkboxes as iOS-style toggle switches.
// Instead of the traditional square checkbox, Cupertino uses a pill-shaped
// toggle switch that slides between on/off states.
//
// The toggle track is a rounded pill. When checked (on), the track fills
// with the accent color and the thumb knob slides to the right.
// When unchecked (off), the track shows a gray outline.
//
// If Theme is nil, CheckboxPainter falls back to the default system blue palette.
type CheckboxPainter struct {
	Theme *Theme // nil uses default system blue fallback
}

// resolveColors returns the CheckboxColorScheme derived from the painter's Theme.
func (p CheckboxPainter) resolveColors() checkbox.CheckboxColorScheme {
	if p.Theme == nil {
		return cupDefaultCheckboxColors
	}
	cs := p.Theme.Colors
	return checkbox.CheckboxColorScheme{
		CheckedBg:       cs.Accent,
		CheckedFg:       cs.OnAccent,
		UncheckedBorder: cs.OpaqueSeparator,
		LabelColor:      cs.Label,
		DisabledBg:      cs.QuaternaryLabel,
		DisabledFg:      cs.TertiaryLabel,
		FocusRing:       cs.Accent.WithAlpha(cupCbFocusAlpha),
	}
}

// PaintCheckbox renders a checkbox as an iOS toggle switch.
func (p CheckboxPainter) PaintCheckbox(canvas widget.Canvas, state checkbox.PaintState) {
	if state.Bounds.IsEmpty() {
		return
	}

	colors := state.ColorScheme
	if colors == (checkbox.CheckboxColorScheme{}) {
		colors = p.resolveColors()
	}

	disabled := state.Disabled
	trackRect := cupToggleTrackRect(state.Bounds)

	if state.Checked || state.Indeterminate {
		cupPaintToggleOn(canvas, trackRect, state, disabled, colors)
	} else {
		cupPaintToggleOff(canvas, trackRect, state, disabled, colors)
	}

	// Draw label if present.
	if state.Label != "" {
		fg := colors.LabelColor
		if disabled {
			fg = colors.DisabledFg
		}
		labelBounds := cupToggleLabelBounds(state.Bounds)
		canvas.DrawText(state.Label, labelBounds, cupCbFontSize, fg, false, cupCbTextAlignLeft)
	}

	// Focus ring.
	if state.Focused && !disabled {
		cupDrawToggleFocusRing(canvas, trackRect, colors)
	}
}

// cupPaintToggleOn draws the toggle in the ON state (checked/indeterminate).
func cupPaintToggleOn(canvas widget.Canvas, trackRect geometry.Rect, state checkbox.PaintState, disabled bool, colors checkbox.CheckboxColorScheme) {
	// Track fill.
	bg := colors.CheckedBg
	if state.Background != nil {
		bg = *state.Background
	}
	if disabled {
		bg = colors.DisabledBg
	} else {
		bg = cupApplyCbState(bg, state.Hovered, state.Pressed)
	}
	trackRadius := trackRect.Height() / 2
	canvas.DrawRoundRect(trackRect, bg, trackRadius)

	// Thumb knob (right side for ON).
	thumbCenter := cupToggleThumbCenter(trackRect, true)
	thumbColor := widget.ColorWhite
	if disabled {
		thumbColor = colors.DisabledFg
	}
	canvas.DrawCircle(thumbCenter, cupCbThumbRadius, thumbColor)

	// Dash indicator for indeterminate.
	if state.Indeterminate {
		dashY := thumbCenter.Y
		dashLeft := trackRect.Min.X + trackRect.Width()*cupCbDashStartX
		dashRight := trackRect.Min.X + trackRect.Width()*cupCbDashEndX
		canvas.DrawLine(
			geometry.Pt(dashLeft, dashY),
			geometry.Pt(dashRight, dashY),
			colors.CheckedFg, cupCbDashStroke,
		)
	}
}

// cupPaintToggleOff draws the toggle in the OFF state (unchecked).
func cupPaintToggleOff(canvas widget.Canvas, trackRect geometry.Rect, state checkbox.PaintState, disabled bool, colors checkbox.CheckboxColorScheme) {
	trackRadius := trackRect.Height() / 2

	// Track outline only.
	borderColor := colors.UncheckedBorder
	if disabled {
		borderColor = colors.DisabledFg
	} else {
		borderColor = cupApplyCbState(borderColor, state.Hovered, state.Pressed)
	}
	canvas.StrokeRoundRect(trackRect, borderColor, trackRadius, cupCbTrackBorderWidth)

	// Thumb knob (left side for OFF).
	thumbCenter := cupToggleThumbCenter(trackRect, false)
	thumbColor := colors.UncheckedBorder
	if disabled {
		thumbColor = colors.DisabledFg
	}
	canvas.DrawCircle(thumbCenter, cupCbThumbRadius, thumbColor)
}

// cupApplyCbState adjusts a color based on interaction state.
func cupApplyCbState(base widget.Color, hovered, pressed bool) widget.Color {
	if pressed {
		return base.Lerp(widget.ColorBlack, cupCbPressedDarken)
	}
	if hovered {
		return base.Lerp(widget.ColorWhite, cupCbHoverLighten)
	}
	return base
}

// cupToggleTrackRect returns the pill-shaped track rectangle.
// Inset by half the border width so centered strokes stay within bounds (#117).
func cupToggleTrackRect(bounds geometry.Rect) geometry.Rect {
	h := bounds.Height()
	trackY := bounds.Min.Y + (h-cupCbTrackHeight)/2
	return geometry.NewRect(bounds.Min.X+cupCbTrackBorderWidth/2, trackY, cupCbTrackWidth, cupCbTrackHeight)
}

// cupToggleThumbCenter returns the center point for the thumb knob.
func cupToggleThumbCenter(trackRect geometry.Rect, on bool) geometry.Point {
	cy := trackRect.Min.Y + trackRect.Height()/2
	if on {
		return geometry.Pt(trackRect.Max.X-cupCbThumbInset-cupCbThumbRadius, cy)
	}
	return geometry.Pt(trackRect.Min.X+cupCbThumbInset+cupCbThumbRadius, cy)
}

// cupToggleLabelBounds returns the label area to the right of the toggle.
func cupToggleLabelBounds(bounds geometry.Rect) geometry.Rect {
	offset := cupCbTrackBorderWidth / 2
	labelX := bounds.Min.X + offset + cupCbTrackWidth + cupCbLabelGap
	labelW := bounds.Width() - offset - cupCbTrackWidth - cupCbLabelGap
	if labelW < 0 {
		labelW = 0
	}
	return geometry.NewRect(labelX, bounds.Min.Y, labelW, bounds.Height())
}

// cupDrawToggleFocusRing draws the blue focus ring around the toggle track.
func cupDrawToggleFocusRing(canvas widget.Canvas, trackRect geometry.Rect, colors checkbox.CheckboxColorScheme) {
	trackRadius := trackRect.Height() / 2
	ringBounds := trackRect.Expand(cupCbFocusRingOffset)
	ringRadius := trackRadius + cupCbFocusRingOffset
	canvas.StrokeRoundRect(ringBounds, colors.FocusRing, ringRadius, cupCbFocusRingStroke)
}

// cupDefaultCheckboxColors holds the default system blue color scheme for toggle switches.
var cupDefaultCheckboxColors = checkbox.CheckboxColorScheme{
	CheckedBg:       systemBlue,
	CheckedFg:       widget.ColorWhite,
	UncheckedBorder: widget.Hex(0xC6C6C8),
	LabelColor:      widget.RGBA(0.0, 0.0, 0.0, 1.0),
	DisabledBg:      widget.RGBA(0.235, 0.235, 0.263, 0.18),
	DisabledFg:      widget.RGBA(0.235, 0.235, 0.263, 0.3),
	FocusRing:       systemBlue.WithAlpha(0.6),
}

// Cupertino toggle switch drawing constants.
const (
	cupCbTrackWidth       float32 = 51
	cupCbTrackHeight      float32 = 31
	cupCbThumbRadius      float32 = 13.5
	cupCbThumbInset       float32 = 2
	cupCbTrackBorderWidth float32 = 2
	cupCbLabelGap         float32 = 8
	cupCbFontSize         float32 = 15
	cupCbTextAlignLeft            = widget.TextAlignLeft
	cupCbFocusRingOffset  float32 = 3
	cupCbFocusRingStroke  float32 = 2.5
	cupCbFocusAlpha       float32 = 0.6
	cupCbHoverLighten     float32 = 0.08
	cupCbPressedDarken    float32 = 0.12
	cupCbDashStartX       float32 = 0.3
	cupCbDashEndX         float32 = 0.7
	cupCbDashStroke       float32 = 2
)

// CheckboxBoxSize returns the Cupertino toggle track width.
// The Cupertino checkbox renders as a toggle switch, so the "box size"
// represents the track width for layout purposes.
func (CheckboxPainter) CheckboxBoxSize() float32 { return cupCbTrackWidth }

// CheckboxLabelGap returns the Cupertino gap between toggle and label.
func (CheckboxPainter) CheckboxLabelGap() float32 { return cupCbLabelGap }

// CheckboxFontSize returns the Cupertino label font size.
func (CheckboxPainter) CheckboxFontSize() float32 { return cupCbFontSize }

// Compile-time checks.
var (
	_ checkbox.Painter       = CheckboxPainter{}
	_ checkbox.LayoutMetrics = CheckboxPainter{}
)
