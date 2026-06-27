package devtools

import (
	"github.com/gogpu/ui/core/linechart"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// LineChartPainter renders line charts using DevTools design tokens.
// DevTools charts use a dark grid (Gray3 lines on Background), muted axes
// (Gray7 text, Gray3 axis lines), and 1.5px line width with Blue6 as the
// default series color, matching JetBrains IDE performance monitoring charts.
//
// If Theme is nil, LineChartPainter falls back to the default DevTools dark palette.
type LineChartPainter struct {
	Theme *Theme // nil uses default DevTools dark fallback
}

// resolveColors returns DevTools-derived colors for chart painting.
func (p LineChartPainter) resolveColors() dtLineChartColors {
	if p.Theme == nil {
		return dtDefaultLineChartColors
	}
	cs := p.Theme.Colors
	return dtLineChartColors{
		Background: cs.Background,
		GridColor:  cs.Border,
		LabelColor: cs.OnSurfaceDisabled,
		LineColor:  cs.Primary,
	}
}

// PaintChart renders a line chart according to DevTools specifications.
func (p LineChartPainter) PaintChart(canvas widget.Canvas, state linechart.PaintState) {
	bounds := state.Bounds
	if bounds.IsEmpty() {
		return
	}

	colors := p.resolveColors()

	// Background fill.
	bg := state.Background
	if bg == (widget.Color{}) {
		bg = colors.Background
	}
	canvas.DrawRect(bounds, bg)

	// Use pre-computed plot area (ADR-034 Phase 4).
	plotArea := state.PlotArea
	if plotArea.IsEmpty() {
		plotArea = dtChartPlotArea(bounds, state.ShowLabels)
	}
	if plotArea.Width() <= 0 || plotArea.Height() <= 0 {
		return
	}

	// Clip to bounds.
	canvas.PushClip(bounds)
	defer canvas.PopClip()

	// Grid lines.
	if state.ShowGrid {
		gridColor := state.GridColor
		if gridColor == (widget.Color{}) {
			gridColor = colors.GridColor
		}
		if len(state.GridLines) > 0 {
			dtChartDrawPrecomputedGrid(canvas, plotArea, state.GridLines, gridColor)
		} else {
			dtChartDrawGrid(canvas, plotArea, gridColor)
		}
	}

	// Y-axis labels.
	if state.ShowLabels {
		if len(state.GridLines) > 0 {
			dtChartDrawPrecomputedLabels(canvas, bounds, state.GridLines, colors.LabelColor)
		} else {
			dtChartDrawLabels(canvas, bounds, plotArea, state, colors.LabelColor)
		}
	}

	// Data lines.
	if len(state.SeriesLines) == len(state.Series) {
		for i, series := range state.Series {
			lineColor := series.Color
			if lineColor == (widget.Color{}) {
				lineColor = colors.LineColor
			}
			dtChartDrawPrecomputedSeries(canvas, state.SeriesLines[i], lineColor)
		}
	} else {
		for _, series := range state.Series {
			lineColor := series.Color
			if lineColor == (widget.Color{}) {
				lineColor = colors.LineColor
			}
			dtChartDrawSeries(canvas, plotArea, series, state, lineColor)
		}
	}
}

// dtChartDrawPrecomputedGrid draws grid lines at pre-computed Y positions.
func dtChartDrawPrecomputedGrid(canvas widget.Canvas, plotArea geometry.Rect, gridLines []linechart.GridLine, color widget.Color) {
	for _, gl := range gridLines {
		from := geometry.Pt(plotArea.Min.X, gl.Y)
		to := geometry.Pt(plotArea.Max.X, gl.Y)
		canvas.DrawLine(from, to, color, dtChartGridLineWidth)
	}
}

// dtChartDrawPrecomputedLabels draws labels at pre-computed grid line positions.
func dtChartDrawPrecomputedLabels(canvas widget.Canvas, bounds geometry.Rect, gridLines []linechart.GridLine, color widget.Color) {
	for _, gl := range gridLines {
		labelBounds := geometry.NewRect(
			bounds.Min.X,
			gl.Y-dtChartLabelFontSize/2,
			dtChartLabelAreaWidth-dtChartLabelPadding,
			dtChartLabelFontSize,
		)
		canvas.DrawText(gl.Label, labelBounds, dtChartLabelFontSize, color, false, widget.TextAlignRight)
	}
}

// dtChartDrawPrecomputedSeries draws connected line segments from pre-computed coordinates.
func dtChartDrawPrecomputedSeries(canvas widget.Canvas, points []geometry.Point, color widget.Color) {
	for i := 1; i < len(points); i++ {
		canvas.DrawLine(points[i-1], points[i], color, dtChartLineWidth)
	}
}

// dtLineChartColors holds the resolved DevTools color roles for chart painting.
type dtLineChartColors struct {
	Background widget.Color
	GridColor  widget.Color
	LabelColor widget.Color
	LineColor  widget.Color
}

// dtDefaultLineChartColors holds default DevTools dark fallback colors.
var dtDefaultLineChartColors = dtLineChartColors{
	Background: widget.Hex(0x1E1F22), // Gray1 (background)
	GridColor:  widget.Hex(0x393B40), // Gray3 (border)
	LabelColor: widget.Hex(0x6F737A), // Gray7 (muted)
	LineColor:  widget.Hex(0x3574F0), // Blue6 (primary)
}

// Chart painting constants.
const (
	dtChartGridDivisions          = 5
	dtChartGridLineWidth  float32 = 1
	dtChartLineWidth      float32 = 1.5
	dtChartLabelAreaWidth float32 = 48
	dtChartLabelPadding   float32 = 4
	dtChartLabelFontSize  float32 = 11
	dtChartZeroThreshold          = 1e-9
)

// dtChartPlotArea returns the rectangle where data lines are drawn.
func dtChartPlotArea(bounds geometry.Rect, showLabels bool) geometry.Rect {
	if showLabels {
		return geometry.NewRect(
			bounds.Min.X+dtChartLabelAreaWidth,
			bounds.Min.Y,
			bounds.Width()-dtChartLabelAreaWidth,
			bounds.Height(),
		)
	}
	return bounds
}

// dtChartDrawGrid draws horizontal grid lines with DevTools styling.
func dtChartDrawGrid(canvas widget.Canvas, plotArea geometry.Rect, color widget.Color) {
	for i := 0; i <= dtChartGridDivisions; i++ {
		t := float32(i) / float32(dtChartGridDivisions)
		y := plotArea.Max.Y - t*plotArea.Height()

		from := geometry.Pt(plotArea.Min.X, y)
		to := geometry.Pt(plotArea.Max.X, y)
		canvas.DrawLine(from, to, color, dtChartGridLineWidth)
	}
}

// dtChartDrawLabels draws Y-axis labels with DevTools typography.
func dtChartDrawLabels(canvas widget.Canvas, bounds, plotArea geometry.Rect, state linechart.PaintState, color widget.Color) {
	yRange := state.YMax - state.YMin
	for i := 0; i <= dtChartGridDivisions; i++ {
		t := float64(i) / float64(dtChartGridDivisions)
		value := state.YMin + t*yRange
		y := plotArea.Max.Y - float32(t)*plotArea.Height()

		labelBounds := geometry.NewRect(
			bounds.Min.X,
			y-dtChartLabelFontSize/2,
			dtChartLabelAreaWidth-dtChartLabelPadding,
			dtChartLabelFontSize,
		)
		text := dtChartFormatLabel(value)
		canvas.DrawText(text, labelBounds, dtChartLabelFontSize, color, false, widget.TextAlignRight)
	}
}

// dtChartDrawSeries draws connected line segments for a single data series.
func dtChartDrawSeries(canvas widget.Canvas, plotArea geometry.Rect, s linechart.Series, state linechart.PaintState, color widget.Color) {
	pointCount := len(s.Points)
	if pointCount < 2 {
		return
	}

	yRange := state.YMax - state.YMin
	if yRange <= dtChartZeroThreshold && yRange >= -dtChartZeroThreshold {
		return
	}

	slots := state.MaxPoints - 1
	if slots < 1 {
		slots = 1
	}
	xStep := plotArea.Width() / float32(slots)
	startX := plotArea.Max.X - float32(pointCount-1)*xStep

	for i := 1; i < pointCount; i++ {
		x1 := startX + float32(i-1)*xStep
		x2 := startX + float32(i)*xStep

		y1 := dtChartYForValue(s.Points[i-1].Value, plotArea, state.YMin, yRange)
		y2 := dtChartYForValue(s.Points[i].Value, plotArea, state.YMin, yRange)

		canvas.DrawLine(
			geometry.Pt(x1, y1),
			geometry.Pt(x2, y2),
			color,
			dtChartLineWidth,
		)
	}
}

// dtChartYForValue converts a data value to a Y pixel coordinate.
func dtChartYForValue(value float64, plotArea geometry.Rect, yMin, yRange float64) float32 {
	t := (value - yMin) / yRange
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return plotArea.Max.Y - float32(t)*plotArea.Height()
}

// dtChartFormatLabel formats a numeric value as a label string.
func dtChartFormatLabel(value float64) string {
	if value == float64(int64(value)) {
		return dtFormatInt(int64(value))
	}
	return dtFormatFloat(value)
}

// dtFormatInt converts an int64 to a string without importing strconv.
func dtFormatInt(v int64) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	buf := [20]byte{}
	i := len(buf) - 1
	for v > 0 {
		buf[i] = byte('0' + v%10)
		i--
		v /= 10
	}
	if neg {
		buf[i] = '-'
		i--
	}
	return string(buf[i+1:])
}

// dtFormatFloat formats a float64 with one decimal place.
func dtFormatFloat(v float64) string {
	neg := v < 0
	if neg {
		v = -v
	}
	intPart := int64(v)
	fracPart := int64((v - float64(intPart)) * 10)
	if fracPart < 0 {
		fracPart = -fracPart
	}

	digit := byte('0' + fracPart%10)
	s := dtFormatInt(intPart) + "." + string(digit)
	if neg {
		s = "-" + s
	}
	return s
}

// Compile-time check that LineChartPainter implements Painter.
var _ linechart.Painter = LineChartPainter{}
