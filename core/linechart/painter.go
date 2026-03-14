package linechart

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// PaintState holds the read-only snapshot passed to the painter.
type PaintState struct {
	Series     []Series
	MaxPoints  int
	YMin       float64
	YMax       float64
	ShowGrid   bool
	ShowLabels bool
	GridColor  widget.Color
	Background widget.Color
}

// Painter renders the chart visuals.
// Each design system (Material 3, Fluent, Cupertino) provides its own
// Painter implementation to render the chart in its visual style.
//
// If no Painter is set, [DefaultPainter] is used.
type Painter interface {
	PaintChart(canvas widget.Canvas, bounds geometry.Rect, state PaintState)
}

// DefaultPainter provides a minimal fallback painter with no design system styling.
// It draws background, grid lines, Y-axis labels, and line segments for each series.
type DefaultPainter struct{}

// PaintChart renders the chart with background, optional grid, optional labels,
// and line segments for each data series.
func (p DefaultPainter) PaintChart(canvas widget.Canvas, bounds geometry.Rect, cs PaintState) {
	if bounds.IsEmpty() {
		return
	}

	// Background fill.
	canvas.DrawRect(bounds, cs.Background)

	// Compute the plot area (inset for labels if enabled).
	plotArea := computePlotArea(bounds, cs.ShowLabels)

	if plotArea.Width() <= 0 || plotArea.Height() <= 0 {
		return
	}

	// Clip to bounds.
	canvas.PushClip(bounds)
	defer canvas.PopClip()

	// Grid lines.
	if cs.ShowGrid {
		drawGrid(canvas, plotArea, cs)
	}

	// Y-axis labels.
	if cs.ShowLabels {
		drawLabels(canvas, bounds, plotArea, cs)
	}

	// Data lines.
	for _, series := range cs.Series {
		drawSeriesLine(canvas, plotArea, series, cs)
	}
}

// computePlotArea returns the rectangle where data lines are drawn.
// When labels are enabled, the left side is inset to make room for them.
func computePlotArea(bounds geometry.Rect, showLabels bool) geometry.Rect {
	if showLabels {
		return geometry.NewRect(
			bounds.Min.X+labelAreaWidth,
			bounds.Min.Y,
			bounds.Width()-labelAreaWidth,
			bounds.Height(),
		)
	}
	return bounds
}

// drawGrid draws horizontal grid lines across the plot area.
func drawGrid(canvas widget.Canvas, plotArea geometry.Rect, cs PaintState) {
	for i := 0; i <= gridDivisions; i++ {
		t := float32(i) / float32(gridDivisions)
		y := plotArea.Max.Y - t*plotArea.Height()

		from := geometry.Pt(plotArea.Min.X, y)
		to := geometry.Pt(plotArea.Max.X, y)
		canvas.DrawLine(from, to, cs.GridColor, gridLineWidth)
	}
}

// drawLabels draws Y-axis labels along the left edge.
func drawLabels(canvas widget.Canvas, bounds geometry.Rect, plotArea geometry.Rect, cs PaintState) {
	yRange := cs.YMax - cs.YMin
	for i := 0; i <= gridDivisions; i++ {
		t := float64(i) / float64(gridDivisions)
		value := cs.YMin + t*yRange
		y := plotArea.Max.Y - float32(t)*plotArea.Height()

		labelBounds := geometry.NewRect(
			bounds.Min.X,
			y-labelFontSize/2,
			labelAreaWidth-labelPadding,
			labelFontSize,
		)
		text := formatLabel(value, cs.YMin, cs.YMax)
		canvas.DrawText(text, labelBounds, labelFontSize, defaultLabelColor, false, labelAlign)
	}
}

// drawSeriesLine draws connected line segments for a single data series.
func drawSeriesLine(canvas widget.Canvas, plotArea geometry.Rect, s Series, cs PaintState) {
	pointCount := len(s.Points)
	if pointCount < 2 {
		return
	}

	yRange := cs.YMax - cs.YMin
	if yRange <= zeroThreshold && yRange >= -zeroThreshold {
		return
	}

	// Determine x spacing based on maxPoints (not actual point count),
	// so that the chart scrolls from right to left as new data arrives.
	slots := cs.MaxPoints - 1
	if slots < 1 {
		slots = 1
	}
	xStep := plotArea.Width() / float32(slots)

	// Points are drawn right-aligned: most recent point at the right edge.
	startX := plotArea.Max.X - float32(pointCount-1)*xStep

	for i := 1; i < pointCount; i++ {
		x1 := startX + float32(i-1)*xStep
		x2 := startX + float32(i)*xStep

		y1 := yForValue(s.Points[i-1].Value, plotArea, cs.YMin, yRange)
		y2 := yForValue(s.Points[i].Value, plotArea, cs.YMin, yRange)

		canvas.DrawLine(
			geometry.Pt(x1, y1),
			geometry.Pt(x2, y2),
			s.Color,
			lineWidth,
		)
	}
}

// yForValue converts a data value to a Y pixel coordinate within the plot area.
// Y=0 (min) is at the bottom, Y=max is at the top.
func yForValue(value float64, plotArea geometry.Rect, yMin, yRange float64) float32 {
	t := (value - yMin) / yRange
	// Clamp to [0, 1].
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return plotArea.Max.Y - float32(t)*plotArea.Height()
}
