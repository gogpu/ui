package focus

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// DefaultFocusRingOffset is the distance between the widget bounds and
// the focus ring, in logical pixels.
const DefaultFocusRingOffset float32 = 2.0

// DefaultFocusRingStrokeWidth is the line width of the focus ring stroke,
// in logical pixels.
const DefaultFocusRingStrokeWidth float32 = 2.0

// DrawFocusRing draws a focus indicator around the given bounds.
//
// The ring is drawn as a rounded rectangle outline, offset slightly
// outside the bounds. The radius controls corner rounding; use 0 for
// square corners.
func DrawFocusRing(canvas widget.Canvas, bounds geometry.Rect, color widget.Color, radius float32) {
	ringBounds := bounds.Expand(DefaultFocusRingOffset)
	ringRadius := radius + DefaultFocusRingOffset

	canvas.StrokeRoundRect(ringBounds, color, ringRadius, DefaultFocusRingStrokeWidth)
}
