package radio

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

// radioCircleGeometry returns the center point and radius for the outer radio circle.
func radioCircleGeometry(bounds geometry.Rect) (geometry.Point, float32) {
	h := bounds.Height()
	cx := bounds.Min.X + outerRadius
	cy := bounds.Min.Y + h/2
	return geometry.Pt(cx, cy), outerRadius
}

// radioLabelBounds returns the label area to the right of the radio circle.
func radioLabelBounds(bounds geometry.Rect) geometry.Rect {
	labelX := bounds.Min.X + outerRadius*2 + labelGap
	labelW := bounds.Width() - outerRadius*2 - labelGap
	if labelW < 0 {
		labelW = 0
	}
	return geometry.NewRect(labelX, bounds.Min.Y, labelW, bounds.Height())
}

// paintSelectedRadio draws the radio item in selected state.
func paintSelectedRadio(canvas widget.Canvas, center geometry.Point, radius float32, state PaintState, hasScheme bool) {
	// Outer filled circle.
	bg := defaultSelectedBg
	if hasScheme {
		bg = state.ColorScheme.SelectedBg
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
	canvas.DrawCircle(center, radius, bg)

	// Inner dot.
	fg := defaultSelectedFg
	if hasScheme {
		fg = state.ColorScheme.SelectedFg
	}
	if state.Disabled {
		if hasScheme {
			fg = state.ColorScheme.DisabledFg
		} else {
			fg = defaultDisabledFg
		}
	}
	canvas.DrawCircle(center, innerRadius, fg)
}

// paintUnselectedRadio draws the radio item in unselected state.
func paintUnselectedRadio(canvas widget.Canvas, center geometry.Point, radius float32, state PaintState, hasScheme bool) {
	borderColor := defaultUnselectedBorder
	if hasScheme {
		borderColor = state.ColorScheme.UnselectedBorder
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
	canvas.StrokeCircle(center, radius, borderColor, borderStrokeWidth)
}

// drawFocusIndicator draws a focus ring around the radio circle.
func drawFocusIndicator(canvas widget.Canvas, center geometry.Point, radius float32, state PaintState, hasScheme bool) {
	ringColor := defaultFocusRingColor
	if hasScheme {
		ringColor = state.ColorScheme.FocusRing
	}
	canvas.StrokeCircle(center, radius+focusRingOffset, ringColor, focusRingStrokeWidth)
}

// Painting constants.
const (
	outerRadius          float32 = 9
	innerRadius          float32 = 4.5
	labelGap             float32 = 8
	borderStrokeWidth    float32 = 2
	defaultFontSize      float32 = 14
	textAlignLeft                = widget.TextAlignLeft
	focusRingOffset      float32 = 2
	focusRingStrokeWidth float32 = 2
	hoverLightenFactor   float32 = 0.1
	pressedDarkenFactor  float32 = 0.15
)

// Default colors for DefaultPainter.
var (
	defaultSelectedBg       = widget.Hex(0x6750A4)
	defaultSelectedFg       = widget.ColorWhite
	defaultUnselectedBorder = widget.RGBA(0.45, 0.45, 0.45, 1.0)
	defaultLabelColor       = widget.RGBA(0.1, 0.1, 0.1, 1.0)
	defaultDisabledBg       = widget.RGBA(0.12, 0.12, 0.13, 0.12)
	defaultDisabledFg       = widget.RGBA(0.12, 0.12, 0.13, 0.38)
	defaultFocusRingColor   = widget.Hex(0x6750A4).WithAlpha(0.7)
)
