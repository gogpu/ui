package material3

import (
	"github.com/gogpu/ui/core/linechart"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// LineChartPainter renders line charts using Material 3 design tokens.
// It maps M3 color roles to chart elements: primary for data lines,
// surface for background, and on-surface-variant for grid/labels.
//
// If Theme is nil, LineChartPainter falls back to the default M3 purple palette.
type LineChartPainter struct {
	Theme *Theme // nil uses default M3 purple fallback
}

// resolveColors returns M3-derived colors for chart painting.
func (p LineChartPainter) resolveColors() lineChartColors {
	if p.Theme == nil {
		return m3DefaultLineChartColors
	}
	cs := p.Theme.Colors
	return lineChartColors{
		Background: cs.Surface,
		GridColor:  cs.OutlineVariant,
		LabelColor: cs.OnSurfaceVariant,
		LineColor:  cs.Primary,
	}
}

// PaintChart renders a line chart according to Material 3 specifications.
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
		plotArea = m3ChartPlotArea(bounds, state.ShowLabels)
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
			m3ChartDrawPrecomputedGrid(canvas, plotArea, state.GridLines, gridColor)
		} else {
			m3ChartDrawGrid(canvas, plotArea, gridColor)
		}
	}

	// Y-axis labels.
	if state.ShowLabels {
		if len(state.GridLines) > 0 {
			m3ChartDrawPrecomputedLabels(canvas, bounds, state.GridLines, colors.LabelColor)
		} else {
			m3ChartDrawLabels(canvas, bounds, plotArea, state, colors.LabelColor)
		}
	}

	// Data lines.
	if len(state.SeriesLines) == len(state.Series) {
		for i, series := range state.Series {
			lineColor := series.Color
			if lineColor == (widget.Color{}) {
				lineColor = colors.LineColor
			}
			m3ChartDrawPrecomputedSeries(canvas, state.SeriesLines[i], lineColor)
		}
	} else {
		for _, series := range state.Series {
			lineColor := series.Color
			if lineColor == (widget.Color{}) {
				lineColor = colors.LineColor
			}
			m3ChartDrawSeries(canvas, plotArea, series, state, lineColor)
		}
	}
}

// m3ChartDrawPrecomputedGrid draws grid lines at pre-computed Y positions.
func m3ChartDrawPrecomputedGrid(canvas widget.Canvas, plotArea geometry.Rect, gridLines []linechart.GridLine, color widget.Color) {
	for _, gl := range gridLines {
		from := geometry.Pt(plotArea.Min.X, gl.Y)
		to := geometry.Pt(plotArea.Max.X, gl.Y)
		canvas.DrawLine(from, to, color, m3ChartGridLineWidth)
	}
}

// m3ChartDrawPrecomputedLabels draws labels at pre-computed grid line positions.
func m3ChartDrawPrecomputedLabels(canvas widget.Canvas, bounds geometry.Rect, gridLines []linechart.GridLine, color widget.Color) {
	for _, gl := range gridLines {
		labelBounds := geometry.NewRect(
			bounds.Min.X,
			gl.Y-m3ChartLabelFontSize/2,
			m3ChartLabelAreaWidth-m3ChartLabelPadding,
			m3ChartLabelFontSize,
		)
		canvas.DrawText(gl.Label, labelBounds, m3ChartLabelFontSize, color, false, widget.TextAlignRight)
	}
}

// m3ChartDrawPrecomputedSeries draws connected line segments from pre-computed coordinates.
func m3ChartDrawPrecomputedSeries(canvas widget.Canvas, points []geometry.Point, color widget.Color) {
	for i := 1; i < len(points); i++ {
		canvas.DrawLine(points[i-1], points[i], color, m3ChartLineWidth)
	}
}

// lineChartColors holds the resolved M3 color roles for chart painting.
type lineChartColors struct {
	Background widget.Color
	GridColor  widget.Color
	LabelColor widget.Color
	LineColor  widget.Color
}

// m3DefaultLineChartColors holds default M3 purple fallback colors.
var m3DefaultLineChartColors = lineChartColors{
	Background: widget.Hex(0xFFFBFE), // M3 surface
	GridColor:  widget.Hex(0xCAC4D0), // M3 outline-variant
	LabelColor: widget.Hex(0x49454F), // M3 on-surface-variant
	LineColor:  widget.Hex(0x6750A4), // M3 primary
}

// Chart painting constants.
const (
	m3ChartGridDivisions          = 5
	m3ChartGridLineWidth  float32 = 1
	m3ChartLineWidth      float32 = 2
	m3ChartLabelAreaWidth float32 = 48
	m3ChartLabelPadding   float32 = 4
	m3ChartLabelFontSize  float32 = 11
	m3ChartZeroThreshold          = 1e-9
)

// m3ChartPlotArea returns the rectangle where data lines are drawn.
func m3ChartPlotArea(bounds geometry.Rect, showLabels bool) geometry.Rect {
	if showLabels {
		return geometry.NewRect(
			bounds.Min.X+m3ChartLabelAreaWidth,
			bounds.Min.Y,
			bounds.Width()-m3ChartLabelAreaWidth,
			bounds.Height(),
		)
	}
	return bounds
}

// m3ChartDrawGrid draws horizontal grid lines with M3 styling.
func m3ChartDrawGrid(canvas widget.Canvas, plotArea geometry.Rect, color widget.Color) {
	for i := 0; i <= m3ChartGridDivisions; i++ {
		t := float32(i) / float32(m3ChartGridDivisions)
		y := plotArea.Max.Y - t*plotArea.Height()

		from := geometry.Pt(plotArea.Min.X, y)
		to := geometry.Pt(plotArea.Max.X, y)
		canvas.DrawLine(from, to, color, m3ChartGridLineWidth)
	}
}

// m3ChartDrawLabels draws Y-axis labels with M3 typography.
func m3ChartDrawLabels(canvas widget.Canvas, bounds, plotArea geometry.Rect, state linechart.PaintState, color widget.Color) {
	yRange := state.YMax - state.YMin
	for i := 0; i <= m3ChartGridDivisions; i++ {
		t := float64(i) / float64(m3ChartGridDivisions)
		value := state.YMin + t*yRange
		y := plotArea.Max.Y - float32(t)*plotArea.Height()

		labelBounds := geometry.NewRect(
			bounds.Min.X,
			y-m3ChartLabelFontSize/2,
			m3ChartLabelAreaWidth-m3ChartLabelPadding,
			m3ChartLabelFontSize,
		)
		text := m3ChartFormatLabel(value)
		canvas.DrawText(text, labelBounds, m3ChartLabelFontSize, color, false, widget.TextAlignRight)
	}
}

// m3ChartDrawSeries draws connected line segments for a single data series.
func m3ChartDrawSeries(canvas widget.Canvas, plotArea geometry.Rect, s linechart.Series, state linechart.PaintState, color widget.Color) {
	pointCount := len(s.Points)
	if pointCount < 2 {
		return
	}

	yRange := state.YMax - state.YMin
	if yRange <= m3ChartZeroThreshold && yRange >= -m3ChartZeroThreshold {
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

		y1 := m3ChartYForValue(s.Points[i-1].Value, plotArea, state.YMin, yRange)
		y2 := m3ChartYForValue(s.Points[i].Value, plotArea, state.YMin, yRange)

		canvas.DrawLine(
			geometry.Pt(x1, y1),
			geometry.Pt(x2, y2),
			color,
			m3ChartLineWidth,
		)
	}
}

// m3ChartYForValue converts a data value to a Y pixel coordinate.
func m3ChartYForValue(value float64, plotArea geometry.Rect, yMin, yRange float64) float32 {
	t := (value - yMin) / yRange
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return plotArea.Max.Y - float32(t)*plotArea.Height()
}

// m3ChartFormatLabel formats a numeric value as a label string.
func m3ChartFormatLabel(value float64) string {
	if value == float64(int64(value)) {
		return formatInt(int64(value))
	}
	return formatFloat(value)
}

// formatInt converts an int64 to a string without importing strconv.
func formatInt(v int64) string {
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

// formatFloat formats a float64 with one decimal place.
func formatFloat(v float64) string {
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
	s := formatInt(intPart) + "." + string(digit)
	if neg {
		s = "-" + s
	}
	return s
}

// Compile-time check that LineChartPainter implements Painter.
var _ linechart.Painter = LineChartPainter{}
