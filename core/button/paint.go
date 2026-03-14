package button

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

// drawFocusIndicator draws a focus ring around the given bounds.
func drawFocusIndicator(canvas widget.Canvas, bounds geometry.Rect, radius float32) {
	ringBounds := bounds.Expand(focusRingOffset)
	ringRadius := radius + focusRingOffset
	canvas.StrokeRoundRect(ringBounds, focusRingColor, ringRadius, focusRingStrokeWidth)
}

// Painting constants.
const (
	defaultRadius        float32 = 8
	outlineStrokeWidth   float32 = 1.5
	focusRingOffset      float32 = 2
	focusRingStrokeWidth float32 = 2
	textAlignCenter              = widget.TextAlignCenter
	hoverLightenFactor   float32 = 0.1
	pressedDarkenFactor  float32 = 0.15
)

// focusRingColor is the default color for focus indicators.
var focusRingColor = widget.Hex(0x6750A4).WithAlpha(0.7)
