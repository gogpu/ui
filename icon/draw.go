package icon

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// defaultStrokeWidth is the base stroke width in viewbox units.
// Scaled proportionally when the icon is rendered at different sizes.
const defaultStrokeWidth float32 = 1.5

// Draw renders an icon's path operations onto the canvas within the given
// bounds, using the specified color.
//
// The icon is uniformly scaled to fit the bounds while preserving aspect ratio
// (the viewbox is square). The stroke width scales proportionally.
//
// Draw can be called directly for custom rendering, or is used internally by
// [IconWidget].
func Draw(canvas widget.Canvas, data IconData, bounds geometry.Rect, color widget.Color) {
	if len(data.Ops) == 0 || data.ViewBox <= 0 {
		return
	}
	if bounds.IsEmpty() {
		return
	}

	bw := bounds.Width()
	bh := bounds.Height()
	scale := bw / data.ViewBox
	if s := bh / data.ViewBox; s < scale {
		scale = s
	}

	// Center the icon within bounds.
	offsetX := bounds.Min.X + (bw-data.ViewBox*scale)/2
	offsetY := bounds.Min.Y + (bh-data.ViewBox*scale)/2

	strokeW := defaultStrokeWidth * scale

	var startX, startY float32 // sub-path start (for Close)
	var curX, curY float32

	for _, op := range data.Ops {
		switch op.Cmd {
		case CmdMoveTo:
			curX = op.Params[0]*scale + offsetX
			curY = op.Params[1]*scale + offsetY
			startX = curX
			startY = curY

		case CmdLineTo:
			newX := op.Params[0]*scale + offsetX
			newY := op.Params[1]*scale + offsetY
			canvas.DrawLine(
				geometry.Pt(curX, curY),
				geometry.Pt(newX, newY),
				color, strokeW,
			)
			curX = newX
			curY = newY

		case CmdCubicTo:
			// Canvas does not have a native cubic Bezier stroke method.
			// Approximate with line segments using De Casteljau subdivision.
			drawCubic(canvas, color, strokeW, scale, offsetX, offsetY,
				curX, curY, op.Params, &curX, &curY)

		case CmdClose:
			if curX != startX || curY != startY {
				canvas.DrawLine(
					geometry.Pt(curX, curY),
					geometry.Pt(startX, startY),
					color, strokeW,
				)
			}
			curX = startX
			curY = startY
		}
	}
}

// cubicSegments is the number of line segments used to approximate a cubic
// Bezier curve. 8 segments provides smooth appearance at typical icon sizes.
const cubicSegments = 8

// drawCubic approximates a cubic Bezier curve with line segments.
//
// The params array contains [cx1, cy1, cx2, cy2, x, y] in viewbox coordinates.
// The current point is updated via outX and outY pointers.
func drawCubic(
	canvas widget.Canvas,
	color widget.Color,
	strokeW, scale, offsetX, offsetY float32,
	curX, curY float32,
	params [maxParams]float32,
	outX, outY *float32,
) {
	// Transform control/end points to canvas coordinates.
	cx1 := params[0]*scale + offsetX
	cy1 := params[1]*scale + offsetY
	cx2 := params[2]*scale + offsetX
	cy2 := params[3]*scale + offsetY
	endX := params[4]*scale + offsetX
	endY := params[5]*scale + offsetY

	prevX, prevY := curX, curY
	for i := 1; i <= cubicSegments; i++ {
		t := float32(i) / float32(cubicSegments)
		t1 := 1 - t

		// De Casteljau cubic evaluation.
		x := t1*t1*t1*prevX + 3*t1*t1*t*cx1 + 3*t1*t*t*cx2 + t*t*t*endX
		y := t1*t1*t1*prevY + 3*t1*t1*t*cy1 + 3*t1*t*t*cy2 + t*t*t*endY

		// The first segment starts from curX/curY, not prevX/prevY of the
		// interpolation. We need the actual current canvas point.
		if i == 1 {
			canvas.DrawLine(geometry.Pt(curX, curY), geometry.Pt(x, y), color, strokeW)
		} else {
			canvas.DrawLine(geometry.Pt(prevX, prevY), geometry.Pt(x, y), color, strokeW)
		}
		prevX = x
		prevY = y
	}

	*outX = endX
	*outY = endY
}
