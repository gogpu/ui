package linechart

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// GridLine holds a pre-computed Y grid position and its formatted label.
type GridLine struct {
	Y     float32 // Y pixel coordinate within the plot area
	Label string  // formatted value string
}

// PaintState holds the read-only snapshot passed to the painter.
type PaintState struct {
	Bounds     geometry.Rect // total widget bounds
	Series     []Series
	MaxPoints  int
	YMin       float64
	YMax       float64
	ShowGrid   bool
	ShowLabels bool
	GridColor  widget.Color
	Background widget.Color

	// Pre-computed plot geometry (ADR-034 Phase 4).
	// The widget computes these from Bounds and data. Painters should use
	// these instead of recomputing data-to-pixel mapping.
	PlotArea    geometry.Rect      // computed plot area (after label margins)
	GridLines   []GridLine         // pre-computed Y grid positions + labels
	SeriesLines [][]geometry.Point // pre-computed pixel coordinates per series
}

// Painter renders the chart visuals.
// Each design system (Material 3, Fluent, Cupertino) provides its own
// Painter implementation to render the chart in its visual style.
//
// If no Painter is set, [DefaultPainter] is used.
type Painter interface {
	PaintChart(canvas widget.Canvas, state PaintState)
}

// DefaultPainter provides a minimal fallback painter with no design system styling.
// It draws background, grid lines, Y-axis labels, and line segments for each series.
type DefaultPainter struct{}

// PaintChart renders the chart with background, optional grid, optional labels,
// and line segments for each data series.
func (p DefaultPainter) PaintChart(canvas widget.Canvas, cs PaintState) {
	bounds := cs.Bounds
	if bounds.IsEmpty() {
		return
	}

	// Background fill.
	canvas.DrawRect(bounds, cs.Background)

	// Prefer pre-computed plot data (ADR-034 Phase 4).
	plotArea := cs.PlotArea
	if plotArea.IsEmpty() {
		// Legacy fallback.
		plotArea = computePlotArea(bounds, cs.ShowLabels)
	}

	if plotArea.Width() <= 0 || plotArea.Height() <= 0 {
		return
	}

	// Clip to bounds.
	canvas.PushClip(bounds)
	defer canvas.PopClip()

	// Grid lines.
	if cs.ShowGrid {
		if len(cs.GridLines) > 0 {
			drawPrecomputedGrid(canvas, plotArea, cs.GridLines, cs.GridColor)
		} else {
			drawGrid(canvas, plotArea, cs)
		}
	}

	// Y-axis labels.
	if cs.ShowLabels {
		if len(cs.GridLines) > 0 {
			drawPrecomputedLabels(canvas, bounds, cs.GridLines, defaultLabelColor)
		} else {
			drawLabels(canvas, bounds, plotArea, cs)
		}
	}

	// Data lines.
	if len(cs.SeriesLines) == len(cs.Series) {
		for i, series := range cs.Series {
			drawPrecomputedSeries(canvas, cs.SeriesLines[i], series.Color)
		}
	} else {
		for _, series := range cs.Series {
			drawSeriesLine(canvas, plotArea, series, cs)
		}
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

// drawPrecomputedGrid draws horizontal grid lines at pre-computed Y positions.
func drawPrecomputedGrid(canvas widget.Canvas, plotArea geometry.Rect, gridLines []GridLine, gridColor widget.Color) {
	for _, gl := range gridLines {
		from := geometry.Pt(plotArea.Min.X, gl.Y)
		to := geometry.Pt(plotArea.Max.X, gl.Y)
		canvas.DrawLine(from, to, gridColor, gridLineWidth)
	}
}

// drawPrecomputedLabels draws Y-axis labels at pre-computed positions.
func drawPrecomputedLabels(canvas widget.Canvas, bounds geometry.Rect, gridLines []GridLine, color widget.Color) {
	for _, gl := range gridLines {
		labelBounds := geometry.NewRect(
			bounds.Min.X,
			gl.Y-labelFontSize/2,
			labelAreaWidth-labelPadding,
			labelFontSize,
		)
		canvas.DrawText(gl.Label, labelBounds, labelFontSize, color, false, labelAlign)
	}
}

// drawPrecomputedSeries draws connected line segments from pre-computed pixel coordinates.
func drawPrecomputedSeries(canvas widget.Canvas, points []geometry.Point, color widget.Color) {
	for i := 1; i < len(points); i++ {
		canvas.DrawLine(points[i-1], points[i], color, lineWidth)
	}
}
