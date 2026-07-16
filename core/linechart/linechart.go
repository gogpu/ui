package linechart

import (
	"fmt"
	"sync"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// DataPoint represents a single data value in a series.
type DataPoint struct {
	Value float64
}

// Series represents a named collection of data points with a display color.
type Series struct {
	Label  string
	Color  widget.Color
	Points []DataPoint
}

// Option configures a line chart during construction.
type Option func(*config)

// config holds the chart's configuration, set at construction time via options.
type config struct {
	maxPoints  int
	yMin       float64
	yMax       float64
	showGrid   bool
	showLabels bool
	gridColor  widget.Color
	background widget.Color
	painter    Painter

	// Signal bindings (4-level priority: ReadonlySignal > Signal > Fn > Static).
	series               []Series
	seriesFn             func() []Series
	seriesSignal         state.Signal[[]Series]
	readonlySeriesSignal state.ReadonlySignal[[]Series]
}

// resolvedSeries returns the current series data.
// Priority: ReadonlySignal > Signal > Fn > Static.
func (c *config) resolvedSeries() []Series {
	if c.readonlySeriesSignal != nil {
		return c.readonlySeriesSignal.Get()
	}
	if c.seriesSignal != nil {
		return c.seriesSignal.Get()
	}
	if c.seriesFn != nil {
		return c.seriesFn()
	}
	return c.series
}

// Default configuration values.
const (
	defaultMaxPoints         = 60
	defaultYMin      float64 = 0
	defaultYMax      float64 = 100
	defaultWidth     float32 = 300
	defaultHeight    float32 = 150
	defaultPadding   float32 = 4
	labelAreaWidth   float32 = 40
	gridDivisions    int     = 4
	lineWidth        float32 = 1.5
	gridLineWidth    float32 = 0.5
	labelFontSize    float32 = 10
	labelAlign               = widget.TextAlignRight
	labelPadding     float32 = 4
	percentMax       float64 = 100
	percentScale     float64 = 100
	percentFmt               = "%.0f%%"
	valueFmt                 = "%.1f"
	zeroThreshold    float64 = 0.0001
)

// Default colors for DefaultPainter.
var (
	defaultBackground = widget.RGBA(0.05, 0.05, 0.08, 1.0)
	defaultGridColor  = widget.RGBA(0.2, 0.2, 0.25, 1.0)
	defaultLabelColor = widget.RGBA(0.6, 0.6, 0.65, 1.0)
)

// MaxPoints sets the rolling window size (number of data points displayed).
// Default is 60.
func MaxPoints(n int) Option {
	return func(c *config) {
		if n > 0 {
			c.maxPoints = n
		}
	}
}

// YRange sets the Y axis range [yMin, yMax].
// Default is [0, 100].
func YRange(yMin, yMax float64) Option {
	return func(c *config) {
		c.yMin = yMin
		c.yMax = yMax
	}
}

// ShowGrid enables or disables grid lines. Default is false.
func ShowGrid(v bool) Option {
	return func(c *config) {
		c.showGrid = v
	}
}

// ShowLabels enables or disables Y axis labels. Default is false.
func ShowLabels(v bool) Option {
	return func(c *config) {
		c.showLabels = v
	}
}

// GridColor sets the color used for grid lines.
func GridColor(color widget.Color) Option {
	return func(c *config) {
		c.gridColor = color
	}
}

// BackgroundColor sets the chart background fill color.
func BackgroundColor(color widget.Color) Option {
	return func(c *config) {
		c.background = color
	}
}

// PainterOpt sets the painter used to render the chart.
// Each design system provides its own painter. If not set,
// [DefaultPainter] is used.
func PainterOpt(p Painter) Option {
	return func(c *config) {
		c.painter = p
	}
}

// SeriesData sets the initial static series data.
func SeriesData(s []Series) Option {
	return func(c *config) {
		c.series = s
	}
}

// SeriesFn sets a dynamic function that returns the current series data.
// When set, this takes precedence over the static data but not over
// a signal set via [SeriesSignal].
func SeriesFn(fn func() []Series) Option {
	return func(c *config) {
		c.seriesFn = fn
	}
}

// SeriesSignal binds the chart's series data to a reactive signal.
// This is a TWO-WAY binding: the widget reads series from the signal,
// and when PushValue modifies series, the signal is updated.
func SeriesSignal(sig state.Signal[[]Series]) Option {
	return func(c *config) {
		c.seriesSignal = sig
	}
}

// SeriesReadonlySignal binds the chart's series data to a read-only signal.
// This is useful for computed signals. When set, this takes highest precedence.
func SeriesReadonlySignal(sig state.ReadonlySignal[[]Series]) Option {
	return func(c *config) {
		c.readonlySeriesSignal = sig
	}
}

// Widget implements a real-time line chart for visualizing time-series data.
//
// A line chart is created with [New] using functional options:
//
//	chart := linechart.New(
//	    linechart.MaxPoints(60),
//	    linechart.YRange(0, 100),
//	    linechart.ShowGrid(true),
//	)
//
// Data is pushed with [Widget.AddSeries] and [Widget.PushValue], which are
// safe to call from any goroutine.
type Widget struct {
	widget.WidgetBase
	cfg     config
	painter Painter

	// mu protects series data for concurrent PushValue calls.
	mu     sync.Mutex
	series []Series

	// scheduler routes invalidation from background goroutines through the
	// main-thread scheduler instead of calling SetNeedsRedraw directly (#182).
	// Set during Mount, cleared during Unmount.
	scheduler widget.SchedulerRef

	// Styling overrides set via fluent methods.
	padding float32
}

// New creates a new line chart Widget with the given options.
//
// The returned widget is visible and enabled by default.
// The default Y range is [0, 100] with a rolling window of 60 points.
func New(opts ...Option) *Widget {
	w := &Widget{
		padding: defaultPadding,
		painter: DefaultPainter{},
	}
	w.SetVisible(true)
	w.SetEnabled(true)

	// Default config.
	w.cfg.maxPoints = defaultMaxPoints
	w.cfg.yMin = defaultYMin
	w.cfg.yMax = defaultYMax
	w.cfg.background = defaultBackground
	w.cfg.gridColor = defaultGridColor

	for _, opt := range opts {
		opt(&w.cfg)
	}

	// Apply painter from config if set.
	if w.cfg.painter != nil {
		w.painter = w.cfg.painter
	}

	// Copy initial static series into mutable state.
	if len(w.cfg.series) > 0 {
		w.series = make([]Series, len(w.cfg.series))
		copy(w.series, w.cfg.series)
	}

	return w
}

// AddSeries adds a named data series with the given color.
// If a series with the same label already exists, this is a no-op.
// Safe to call from any goroutine.
func (w *Widget) AddSeries(label string, color widget.Color) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, s := range w.series {
		if s.Label == label {
			return
		}
	}

	w.series = append(w.series, Series{
		Label:  label,
		Color:  color,
		Points: make([]DataPoint, 0, w.cfg.maxPoints),
	})
	w.syncToSignal()
	w.requestRedraw()
}

// PushValue appends a data point to the named series. If the series has
// reached MaxPoints, the oldest point is removed (rolling window).
// Safe to call from any goroutine.
func (w *Widget) PushValue(label string, value float64) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for i := range w.series {
		if w.series[i].Label != label {
			continue
		}

		pts := w.series[i].Points
		if len(pts) >= w.cfg.maxPoints {
			// Shift left: drop oldest point.
			copy(pts, pts[1:])
			pts[len(pts)-1] = DataPoint{Value: value}
		} else {
			pts = append(pts, DataPoint{Value: value})
		}
		w.series[i].Points = pts
		w.syncToSignal()
		w.requestRedraw()
		return
	}
}

// ClearSeries removes all data points from the named series.
// Safe to call from any goroutine.
func (w *Widget) ClearSeries(label string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for i := range w.series {
		if w.series[i].Label == label {
			w.series[i].Points = w.series[i].Points[:0]
			w.syncToSignal()
			w.requestRedraw()
			return
		}
	}
}

// SeriesCount returns the number of series. Safe to call from any goroutine.
func (w *Widget) SeriesCount() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.series)
}

// requestRedraw routes invalidation through the scheduler when available,
// falling back to direct SetNeedsRedraw for unmounted usage.
// Must be called with w.mu held (scheduler field is read-only after Mount).
func (w *Widget) requestRedraw() {
	if w.scheduler != nil {
		w.scheduler.MarkDirty(w)
	} else {
		w.SetNeedsRedraw(true)
	}
}

// syncToSignal writes the current series data back to the signal if bound.
// Must be called with w.mu held.
func (w *Widget) syncToSignal() {
	if w.cfg.seriesSignal != nil {
		// Copy to avoid sharing mutable slices.
		cp := make([]Series, len(w.series))
		copy(cp, w.series)
		w.cfg.seriesSignal.Set(cp)
	}
}

// Padding sets the padding around the chart area.
// Returns the widget for method chaining.
func (w *Widget) Padding(v float32) *Widget {
	w.padding = v
	return w
}

// Layout calculates the chart's preferred size within the given constraints.
func (w *Widget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	width := constraints.MaxWidth
	if width <= 0 || width == geometry.Infinity {
		width = defaultWidth + w.padding*2
	}
	height := defaultHeight + w.padding*2
	return constraints.Constrain(geometry.Sz(width, height))
}

// Draw renders the chart to the canvas.
func (w *Widget) Draw(_ widget.Context, canvas widget.Canvas) {
	bounds := w.Bounds()
	if bounds.IsEmpty() {
		return
	}

	// Build PaintState from current data.
	chartState := w.buildPaintState(bounds)
	w.painter.PaintChart(canvas, chartState)
}

// buildPaintState creates a read-only snapshot for the painter.
// Pre-computes plot geometry so painters only draw (ADR-034 Phase 4).
func (w *Widget) buildPaintState(bounds geometry.Rect) PaintState {
	cs := PaintState{
		Bounds:     bounds,
		MaxPoints:  w.cfg.maxPoints,
		YMin:       w.cfg.yMin,
		YMax:       w.cfg.yMax,
		ShowGrid:   w.cfg.showGrid,
		ShowLabels: w.cfg.showLabels,
		GridColor:  w.cfg.gridColor,
		Background: w.cfg.background,
	}

	// Prefer signal-based series over local mutable state.
	if w.cfg.readonlySeriesSignal != nil || w.cfg.seriesSignal != nil || w.cfg.seriesFn != nil {
		cs.Series = w.cfg.resolvedSeries()
	} else {
		w.mu.Lock()
		cs.Series = make([]Series, len(w.series))
		copy(cs.Series, w.series)
		w.mu.Unlock()
	}

	// Pre-compute plot geometry (ADR-034 Phase 4).
	cs.PlotArea = computePlotArea(bounds, cs.ShowLabels)
	cs.GridLines = w.computeGridLines(cs.PlotArea, cs.YMin, cs.YMax)
	cs.SeriesLines = w.computeSeriesLines(cs.PlotArea, cs.Series, cs.YMin, cs.YMax)

	return cs
}

// computeGridLines pre-computes Y grid line positions and labels.
func (w *Widget) computeGridLines(plotArea geometry.Rect, yMin, yMax float64) []GridLine {
	if plotArea.IsEmpty() {
		return nil
	}
	lines := make([]GridLine, 0, gridDivisions+1)
	yRange := yMax - yMin
	for i := 0; i <= gridDivisions; i++ {
		t := float64(i) / float64(gridDivisions)
		value := yMin + t*yRange
		y := plotArea.Max.Y - float32(t)*plotArea.Height()
		lines = append(lines, GridLine{
			Y:     y,
			Label: formatLabel(value, yMin, yMax),
		})
	}
	return lines
}

// computeSeriesLines pre-computes pixel coordinates for each data series.
func (w *Widget) computeSeriesLines(plotArea geometry.Rect, series []Series, yMin, yMax float64) [][]geometry.Point {
	if plotArea.IsEmpty() || len(series) == 0 {
		return nil
	}

	yRange := yMax - yMin
	if yRange <= zeroThreshold && yRange >= -zeroThreshold {
		return nil
	}

	slots := w.cfg.maxPoints - 1
	if slots < 1 {
		slots = 1
	}
	xStep := plotArea.Width() / float32(slots)

	result := make([][]geometry.Point, len(series))
	for si, s := range series {
		pointCount := len(s.Points)
		if pointCount < 2 {
			result[si] = nil
			continue
		}
		startX := plotArea.Max.X - float32(pointCount-1)*xStep
		pts := make([]geometry.Point, pointCount)
		for i := 0; i < pointCount; i++ {
			x := startX + float32(i)*xStep
			y := yForValue(s.Points[i].Value, plotArea, yMin, yRange)
			pts[i] = geometry.Pt(x, y)
		}
		result[si] = pts
	}
	return result
}

// Event handles an input event and returns true if consumed.
// LineChart is a display-only widget and does not consume events.
func (w *Widget) Event(_ widget.Context, _ event.Event) bool {
	return false
}

// Children returns nil because a line chart is a leaf widget.
func (w *Widget) Children() []widget.Widget {
	return nil
}

// Mount creates signal bindings for push-based invalidation.
// Implements [widget.Lifecycle].
func (w *Widget) Mount(ctx widget.Context) {
	sched := ctx.Scheduler()

	// Store scheduler for goroutine-safe invalidation (#182).
	w.mu.Lock()
	w.scheduler = sched
	w.mu.Unlock()

	if sched == nil {
		return
	}
	if w.cfg.readonlySeriesSignal != nil {
		b := state.BindToScheduler(w.cfg.readonlySeriesSignal, w, sched)
		w.AddBinding(b)
	} else if w.cfg.seriesSignal != nil {
		b := state.BindToScheduler(w.cfg.seriesSignal, w, sched)
		w.AddBinding(b)
	}
}

// Unmount is called when the chart is removed from the widget tree.
// Implements [widget.Lifecycle].
func (w *Widget) Unmount() {
	// Clear scheduler reference — after unmount, direct SetNeedsRedraw is used.
	w.mu.Lock()
	w.scheduler = nil
	w.mu.Unlock()
	// Bindings are cleaned up automatically by WidgetBase.CleanupBindings().
}

// formatLabel formats a Y axis value as a label string.
// Values in [0, 100] range are displayed as percentages; others as decimals.
func formatLabel(value, yMin, yMax float64) string {
	if yMin >= 0 && yMax <= percentMax {
		return fmt.Sprintf(percentFmt, value)
	}
	return fmt.Sprintf(valueFmt, value)
}

// Verify Widget implements required interfaces at compile time.
var (
	_ widget.Widget    = (*Widget)(nil)
	_ widget.Lifecycle = (*Widget)(nil)
)
