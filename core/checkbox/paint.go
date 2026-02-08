package checkbox

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// applyStateModifier adjusts a color based on interaction state.
func applyStateModifier(base widget.Color, hovered, pressed bool) widget.Color {
	if pressed {
		return base.Lerp(widget.ColorBlack, pressedDarkenFactor)
	}
	if hovered {
		return base.Lerp(widget.ColorWhite, hoverLightenFactor)
	}
	return base
}

// drawFocusIndicator draws a focus ring around the checkbox box.
func drawFocusIndicator(canvas widget.Canvas, boxRect geometry.Rect) {
	ringBounds := boxRect.Expand(focusRingOffset)
	ringRadius := boxCornerRadius + focusRingOffset
	canvas.StrokeRoundRect(ringBounds, focusRingColor, ringRadius, focusRingStrokeWidth)
}

// checkboxBoxRect returns the square box area on the left side of the bounds.
func checkboxBoxRect(bounds geometry.Rect) geometry.Rect {
	h := bounds.Height()
	boxY := bounds.Min.Y + (h-boxSize)/2
	return geometry.NewRect(bounds.Min.X, boxY, boxSize, boxSize)
}

// checkboxLabelBounds returns the label area to the right of the checkbox box.
func checkboxLabelBounds(bounds geometry.Rect) geometry.Rect {
	return geometry.NewRect(
		bounds.Min.X+boxSize+labelGap,
		bounds.Min.Y,
		bounds.Width()-boxSize-labelGap,
		bounds.Height(),
	)
}

// paintCheckedBox draws the checkbox in checked or indeterminate state.
func paintCheckedBox(canvas widget.Canvas, boxRect geometry.Rect, state PaintState, hasScheme bool) {
	bg := defaultCheckedBg
	if hasScheme {
		bg = state.ColorScheme.CheckedBg
	}
	if state.Background != nil {
		bg = *state.Background
	}
	if state.Disabled {
		if hasScheme {
			bg = state.ColorScheme.DisabledBg
		} else {
			bg = defaultDisabledBg
		}
	} else {
		bg = applyStateModifier(bg, state.Hovered, state.Pressed)
	}
	canvas.DrawRoundRect(boxRect, bg, boxCornerRadius)

	// Draw checkmark or dash.
	fg := defaultCheckedFg
	if hasScheme {
		fg = state.ColorScheme.CheckedFg
	}
	if state.Disabled {
		if hasScheme {
			fg = state.ColorScheme.DisabledFg
		} else {
			fg = defaultDisabledFg
		}
	}

	if state.Indeterminate {
		drawDash(canvas, boxRect, fg)
	} else {
		drawCheckmark(canvas, boxRect, fg)
	}
}

// paintUncheckedBox draws the checkbox in unchecked state.
func paintUncheckedBox(canvas widget.Canvas, boxRect geometry.Rect, state PaintState, hasScheme bool) {
	borderColor := defaultUncheckedBorder
	if hasScheme {
		borderColor = state.ColorScheme.UncheckedBorder
	}
	if state.Disabled {
		if hasScheme {
			borderColor = state.ColorScheme.DisabledFg
		} else {
			borderColor = defaultDisabledFg
		}
	} else {
		borderColor = applyStateModifier(borderColor, state.Hovered, state.Pressed)
	}
	canvas.StrokeRoundRect(boxRect, borderColor, boxCornerRadius, borderStrokeWidth)
}

// drawCheckmark draws a checkmark (two lines forming a V shape) inside the box.
func drawCheckmark(canvas widget.Canvas, box geometry.Rect, color widget.Color) {
	cx := box.Min.X
	cy := box.Min.Y
	s := box.Width()

	// Checkmark points (relative to box, normalized to box size).
	// Start from lower-left area, dip to bottom-center, up to top-right.
	p1 := geometry.Pt(cx+s*checkP1X, cy+s*checkP1Y)
	p2 := geometry.Pt(cx+s*checkP2X, cy+s*checkP2Y)
	p3 := geometry.Pt(cx+s*checkP3X, cy+s*checkP3Y)

	canvas.DrawLine(p1, p2, color, checkStrokeWidth)
	canvas.DrawLine(p2, p3, color, checkStrokeWidth)
}

// drawDash draws a horizontal dash inside the box for indeterminate state.
func drawDash(canvas widget.Canvas, box geometry.Rect, color widget.Color) {
	cx := box.Min.X
	cy := box.Min.Y
	s := box.Width()
	midY := cy + s*dashMidY

	from := geometry.Pt(cx+s*dashStartX, midY)
	to := geometry.Pt(cx+s*dashEndX, midY)

	canvas.DrawLine(from, to, color, dashStrokeWidth)
}

// Painting constants.
const (
	boxSize           float32 = 18
	boxCornerRadius   float32 = 3
	labelGap          float32 = 8
	borderStrokeWidth float32 = 2
	defaultFontSize   float32 = 14
	textAlignLeft     float32 = 0

	focusRingOffset      float32 = 2
	focusRingStrokeWidth float32 = 2
	hoverLightenFactor   float32 = 0.1
	pressedDarkenFactor  float32 = 0.15

	// Checkmark geometry (relative to box size 0..1).
	checkP1X         float32 = 0.2
	checkP1Y         float32 = 0.5
	checkP2X         float32 = 0.4
	checkP2Y         float32 = 0.7
	checkP3X         float32 = 0.8
	checkP3Y         float32 = 0.3
	checkStrokeWidth float32 = 2

	// Dash geometry (relative to box size 0..1).
	dashStartX      float32 = 0.25
	dashEndX        float32 = 0.75
	dashMidY        float32 = 0.5
	dashStrokeWidth float32 = 2
)

// Default colors for DefaultPainter.
var (
	defaultCheckedBg       = widget.Hex(0x6750A4)
	defaultCheckedFg       = widget.ColorWhite
	defaultUncheckedBorder = widget.RGBA(0.45, 0.45, 0.45, 1.0)
	defaultLabelColor      = widget.RGBA(0.1, 0.1, 0.1, 1.0)
	defaultDisabledBg      = widget.RGBA(0.12, 0.12, 0.13, 0.12)
	defaultDisabledFg      = widget.RGBA(0.12, 0.12, 0.13, 0.38)
	focusRingColor         = widget.Hex(0x6750A4).WithAlpha(0.7)
)
