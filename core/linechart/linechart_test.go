package linechart

import (
	"image"
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// --- Construction Tests ---

func TestNew_Defaults(t *testing.T) {
	w := New()

	if !w.IsVisible() {
		t.Error("default chart should be visible")
	}
	if !w.IsEnabled() {
		t.Error("default chart should be enabled")
	}
	if w.Children() != nil {
		t.Error("chart should have no children")
	}
	if w.cfg.maxPoints != defaultMaxPoints {
		t.Errorf("maxPoints = %d, want %d", w.cfg.maxPoints, defaultMaxPoints)
	}
	if w.cfg.yMin != defaultYMin {
		t.Errorf("yMin = %f, want %f", w.cfg.yMin, defaultYMin)
	}
	if w.cfg.yMax != defaultYMax {
		t.Errorf("yMax = %f, want %f", w.cfg.yMax, defaultYMax)
	}
	if w.cfg.showGrid {
		t.Error("default chart should not show grid")
	}
	if w.cfg.showLabels {
		t.Error("default chart should not show labels")
	}
}

func TestNew_WithOptions(t *testing.T) {
	bg := widget.RGB(0.1, 0.1, 0.1)
	gc := widget.RGB(0.3, 0.3, 0.3)
	w := New(
		MaxPoints(120),
		YRange(-10, 200),
		ShowGrid(true),
		ShowLabels(true),
		GridColor(gc),
		BackgroundColor(bg),
	)

	if w.cfg.maxPoints != 120 {
		t.Errorf("maxPoints = %d, want 120", w.cfg.maxPoints)
	}
	if w.cfg.yMin != -10 {
		t.Errorf("yMin = %f, want -10", w.cfg.yMin)
	}
	if w.cfg.yMax != 200 {
		t.Errorf("yMax = %f, want 200", w.cfg.yMax)
	}
	if !w.cfg.showGrid {
		t.Error("showGrid should be true")
	}
	if !w.cfg.showLabels {
		t.Error("showLabels should be true")
	}
	if w.cfg.gridColor != gc {
		t.Errorf("gridColor = %v, want %v", w.cfg.gridColor, gc)
	}
	if w.cfg.background != bg {
		t.Errorf("background = %v, want %v", w.cfg.background, bg)
	}
}

func TestNew_MaxPointsZeroIgnored(t *testing.T) {
	w := New(MaxPoints(0))
	if w.cfg.maxPoints != defaultMaxPoints {
		t.Errorf("maxPoints = %d, want %d (zero should be ignored)", w.cfg.maxPoints, defaultMaxPoints)
	}
}

func TestNew_MaxPointsNegativeIgnored(t *testing.T) {
	w := New(MaxPoints(-5))
	if w.cfg.maxPoints != defaultMaxPoints {
		t.Errorf("maxPoints = %d, want %d (negative should be ignored)", w.cfg.maxPoints, defaultMaxPoints)
	}
}

func TestNew_WithCustomPainter(t *testing.T) {
	p := &mockPainter{}
	w := New(PainterOpt(p))
	if w.painter != p {
		t.Error("painter should be the custom painter")
	}
}

func TestNew_WithInitialSeries(t *testing.T) {
	series := []Series{
		{Label: "CPU", Color: widget.ColorRed, Points: []DataPoint{{Value: 10}, {Value: 20}}},
	}
	w := New(SeriesData(series))

	if len(w.series) != 1 {
		t.Fatalf("series count = %d, want 1", len(w.series))
	}
	if w.series[0].Label != "CPU" {
		t.Errorf("series label = %q, want %q", w.series[0].Label, "CPU")
	}
	if len(w.series[0].Points) != 2 {
		t.Errorf("points count = %d, want 2", len(w.series[0].Points))
	}
}

// --- Layout Tests ---

func TestLayout_RespectsConstraints(t *testing.T) {
	ctx := widget.NewContext()
	constraints := geometry.Tight(geometry.Sz(400, 200))

	w := New()
	size := w.Layout(ctx, constraints)

	if size.Width != 400 {
		t.Errorf("width = %v, want 400 (tight)", size.Width)
	}
	if size.Height != 200 {
		t.Errorf("height = %v, want 200 (tight)", size.Height)
	}
}

func TestLayout_PreferredSize(t *testing.T) {
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 600))

	w := New()
	size := w.Layout(ctx, constraints)

	expectedW := defaultWidth + defaultPadding*2
	expectedH := defaultHeight + defaultPadding*2

	if size.Width != expectedW {
		t.Errorf("width = %v, want %v", size.Width, expectedW)
	}
	if size.Height != expectedH {
		t.Errorf("height = %v, want %v", size.Height, expectedH)
	}
}

// --- Draw Tests ---

func TestDraw_EmptyBounds(t *testing.T) {
	w := New()
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	w.Draw(ctx, canvas)

	if canvas.drawCount > 0 {
		t.Error("should not draw anything with empty bounds")
	}
}

func TestDraw_NormalState(t *testing.T) {
	w := New(ShowGrid(true), ShowLabels(true))
	w.SetBounds(geometry.NewRect(0, 0, 400, 200))
	w.AddSeries("CPU", widget.ColorRed)
	w.PushValue("CPU", 25)
	w.PushValue("CPU", 50)
	w.PushValue("CPU", 75)

	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	w.Draw(ctx, canvas)

	if canvas.drawCount == 0 {
		t.Error("should draw something with valid bounds and data")
	}
}

func TestDraw_NoData(t *testing.T) {
	w := New()
	w.SetBounds(geometry.NewRect(0, 0, 400, 200))

	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	w.Draw(ctx, canvas)

	// Should still draw background at minimum.
	if canvas.rectCount == 0 {
		t.Error("should draw background even without data")
	}
}

func TestDraw_SinglePoint(t *testing.T) {
	w := New()
	w.SetBounds(geometry.NewRect(0, 0, 400, 200))
	w.AddSeries("CPU", widget.ColorRed)
	w.PushValue("CPU", 50)

	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	w.Draw(ctx, canvas)

	// Single point = no line segments (need at least 2).
	if canvas.lineCount > 0 {
		t.Error("should not draw lines with only 1 data point")
	}
}

func TestDraw_MultipleSeries(t *testing.T) {
	w := New()
	w.SetBounds(geometry.NewRect(0, 0, 400, 200))
	w.AddSeries("CPU", widget.ColorRed)
	w.AddSeries("Memory", widget.ColorBlue)

	for i := 0; i < 5; i++ {
		w.PushValue("CPU", float64(i*10))
		w.PushValue("Memory", float64(i*5))
	}

	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	w.Draw(ctx, canvas)

	// Each series with 5 points = 4 line segments each = 8 total.
	if canvas.lineCount != 8 {
		t.Errorf("lineCount = %d, want 8 (2 series x 4 segments)", canvas.lineCount)
	}
}

func TestDraw_WithGridAndLabels(t *testing.T) {
	w := New(ShowGrid(true), ShowLabels(true))
	w.SetBounds(geometry.NewRect(0, 0, 400, 200))

	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	w.Draw(ctx, canvas)

	// Grid: 5 horizontal lines (0%, 25%, 50%, 75%, 100%).
	if canvas.lineCount != gridDivisions+1 {
		t.Errorf("grid lineCount = %d, want %d", canvas.lineCount, gridDivisions+1)
	}

	// Labels: 5 text draws.
	if canvas.textCount != gridDivisions+1 {
		t.Errorf("label textCount = %d, want %d", canvas.textCount, gridDivisions+1)
	}
}

// --- Data Management Tests ---

func TestAddSeries(t *testing.T) {
	w := New()
	w.AddSeries("CPU", widget.ColorRed)

	if w.SeriesCount() != 1 {
		t.Fatalf("series count = %d, want 1", w.SeriesCount())
	}
}

func TestAddSeries_Duplicate(t *testing.T) {
	w := New()
	w.AddSeries("CPU", widget.ColorRed)
	w.AddSeries("CPU", widget.ColorBlue) // should be no-op

	if w.SeriesCount() != 1 {
		t.Errorf("series count = %d, want 1 (duplicate ignored)", w.SeriesCount())
	}
}

func TestPushValue_RollingWindow(t *testing.T) {
	w := New(MaxPoints(5))
	w.AddSeries("CPU", widget.ColorRed)

	for i := 0; i < 10; i++ {
		w.PushValue("CPU", float64(i))
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.series[0].Points) != 5 {
		t.Fatalf("points = %d, want 5 (rolling window)", len(w.series[0].Points))
	}

	// Should contain the last 5 values: 5, 6, 7, 8, 9.
	for i, expected := range []float64{5, 6, 7, 8, 9} {
		if w.series[0].Points[i].Value != expected {
			t.Errorf("point[%d] = %v, want %v", i, w.series[0].Points[i].Value, expected)
		}
	}
}

func TestPushValue_UnknownSeries(t *testing.T) {
	w := New()
	// Should not panic.
	w.PushValue("nonexistent", 42)
}

func TestClearSeries(t *testing.T) {
	w := New()
	w.AddSeries("CPU", widget.ColorRed)
	w.PushValue("CPU", 10)
	w.PushValue("CPU", 20)
	w.ClearSeries("CPU")

	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.series[0].Points) != 0 {
		t.Errorf("points = %d, want 0 after clear", len(w.series[0].Points))
	}
}

func TestClearSeries_UnknownSeries(t *testing.T) {
	w := New()
	// Should not panic.
	w.ClearSeries("nonexistent")
}

func TestPushValue_SetsNeedsRedraw(t *testing.T) {
	w := New()
	w.AddSeries("CPU", widget.ColorRed)
	w.SetNeedsRedraw(false)

	w.PushValue("CPU", 50)

	if !w.NeedsRedraw() {
		t.Error("PushValue should set needsRedraw")
	}
}

func TestAddSeries_SetsNeedsRedraw(t *testing.T) {
	w := New()
	w.SetNeedsRedraw(false)

	w.AddSeries("CPU", widget.ColorRed)

	if !w.NeedsRedraw() {
		t.Error("AddSeries should set needsRedraw")
	}
}

// --- Signal Binding Tests ---

func TestSeriesSignal_ReadFromSignal(t *testing.T) {
	series := []Series{
		{Label: "CPU", Color: widget.ColorRed, Points: []DataPoint{{Value: 30}, {Value: 60}}},
	}
	sig := state.NewSignal(series)
	w := New(SeriesSignal(sig))
	w.SetBounds(geometry.NewRect(0, 0, 400, 200))

	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	w.Draw(ctx, canvas)

	// Should draw 1 line segment (2 points).
	if canvas.lineCount != 1 {
		t.Errorf("lineCount = %d, want 1 from signal data", canvas.lineCount)
	}
}

func TestSeriesReadonlySignal_ReadFromSignal(t *testing.T) {
	series := []Series{
		{Label: "CPU", Color: widget.ColorRed, Points: []DataPoint{{Value: 10}, {Value: 50}, {Value: 90}}},
	}
	sig := state.NewSignal(series)
	w := New(SeriesReadonlySignal(sig))
	w.SetBounds(geometry.NewRect(0, 0, 400, 200))

	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	w.Draw(ctx, canvas)

	// Should draw 2 line segments (3 points).
	if canvas.lineCount != 2 {
		t.Errorf("lineCount = %d, want 2 from readonly signal data", canvas.lineCount)
	}
}

func TestSeriesFn_ReadFromFunction(t *testing.T) {
	series := []Series{
		{Label: "CPU", Color: widget.ColorRed, Points: []DataPoint{{Value: 25}, {Value: 75}}},
	}
	w := New(SeriesFn(func() []Series { return series }))
	w.SetBounds(geometry.NewRect(0, 0, 400, 200))

	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	w.Draw(ctx, canvas)

	if canvas.lineCount != 1 {
		t.Errorf("lineCount = %d, want 1 from series function", canvas.lineCount)
	}
}

func TestSeriesSignal_TwoWayBinding(t *testing.T) {
	sig := state.NewSignal[[]Series](nil)
	w := New(SeriesSignal(sig), MaxPoints(10))

	w.AddSeries("CPU", widget.ColorRed)
	w.PushValue("CPU", 42)

	// Signal should reflect the updated data.
	result := sig.Get()
	if len(result) != 1 {
		t.Fatalf("signal series count = %d, want 1", len(result))
	}
	if len(result[0].Points) != 1 {
		t.Fatalf("signal points count = %d, want 1", len(result[0].Points))
	}
	if result[0].Points[0].Value != 42 {
		t.Errorf("signal value = %v, want 42", result[0].Points[0].Value)
	}
}

// --- Config Resolution Tests ---

func TestResolvedSeries_Priority(t *testing.T) {
	staticSeries := []Series{{Label: "Static"}}
	fnSeries := []Series{{Label: "Fn"}}
	sigSeries := []Series{{Label: "Signal"}}
	roSeries := []Series{{Label: "ReadonlySignal"}}

	t.Run("static", func(t *testing.T) {
		c := config{series: staticSeries}
		got := c.resolvedSeries()
		if len(got) != 1 || got[0].Label != "Static" {
			t.Errorf("expected Static, got %v", got)
		}
	})

	t.Run("fn over static", func(t *testing.T) {
		c := config{series: staticSeries, seriesFn: func() []Series { return fnSeries }}
		got := c.resolvedSeries()
		if len(got) != 1 || got[0].Label != "Fn" {
			t.Errorf("expected Fn, got %v", got)
		}
	})

	t.Run("signal over fn", func(t *testing.T) {
		sig := state.NewSignal(sigSeries)
		c := config{series: staticSeries, seriesFn: func() []Series { return fnSeries }, seriesSignal: sig}
		got := c.resolvedSeries()
		if len(got) != 1 || got[0].Label != "Signal" {
			t.Errorf("expected Signal, got %v", got)
		}
	})

	t.Run("readonly signal over signal", func(t *testing.T) {
		sig := state.NewSignal(sigSeries)
		roSig := state.NewSignal(roSeries)
		c := config{seriesSignal: sig, readonlySeriesSignal: roSig}
		got := c.resolvedSeries()
		if len(got) != 1 || got[0].Label != "ReadonlySignal" {
			t.Errorf("expected ReadonlySignal, got %v", got)
		}
	})
}

// --- Lifecycle Tests ---

func TestMount_CreatesBindings(t *testing.T) {
	sig := state.NewSignal[[]Series](nil)
	w := New(SeriesSignal(sig))

	sched := &mockScheduler{}
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	w.Mount(ctx)

	// Change signal value — scheduler should be notified.
	sig.Set([]Series{{Label: "test"}})

	if sched.dirtyCount == 0 {
		t.Error("scheduler should have been notified after signal change")
	}
}

func TestMount_ReadonlySignal(t *testing.T) {
	sig := state.NewSignal[[]Series](nil)
	w := New(SeriesReadonlySignal(sig))

	sched := &mockScheduler{}
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	w.Mount(ctx)

	sig.Set([]Series{{Label: "test"}})

	if sched.dirtyCount == 0 {
		t.Error("scheduler should have been notified for readonly signal")
	}
}

func TestMount_NoScheduler(t *testing.T) {
	sig := state.NewSignal[[]Series](nil)
	w := New(SeriesSignal(sig))
	ctx := widget.NewContext()

	// Should not panic with nil scheduler.
	w.Mount(ctx)
}

func TestUnmount_DoesNotPanic(t *testing.T) {
	w := New()
	w.Unmount()
}

// --- Event Tests ---

func TestEvent_DoesNotConsume(t *testing.T) {
	w := New()
	w.SetBounds(geometry.NewRect(0, 0, 400, 200))
	ctx := widget.NewContext()

	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(100, 100), geometry.Pt(100, 100), event.ModNone)

	if w.Event(ctx, press) {
		t.Error("linechart should not consume mouse events")
	}
}

// --- Fluent Styling Tests ---

func TestPadding_Chaining(t *testing.T) {
	w := New()
	result := w.Padding(16)

	if result != w {
		t.Error("Padding should return same widget for chaining")
	}
}

// --- Painter Tests ---

func TestDefaultPainter_EmptyBounds(t *testing.T) {
	p := DefaultPainter{}
	canvas := &recordingCanvas{}
	p.PaintChart(canvas, geometry.Rect{}, PaintState{})

	if canvas.drawCount > 0 {
		t.Error("should not draw with empty bounds")
	}
}

func TestDefaultPainter_ZeroYRange(t *testing.T) {
	p := DefaultPainter{}
	canvas := &recordingCanvas{}
	cs := PaintState{
		YMin:      50,
		YMax:      50,
		MaxPoints: 60,
		Series: []Series{
			{Label: "CPU", Color: widget.ColorRed, Points: []DataPoint{{Value: 50}, {Value: 50}}},
		},
		Background: defaultBackground,
	}
	bounds := geometry.NewRect(0, 0, 400, 200)
	p.PaintChart(canvas, bounds, cs)

	// Should draw background but no lines (zero range).
	if canvas.lineCount > 0 {
		t.Error("should not draw lines with zero Y range")
	}
}

func TestDefaultPainter_ClampValues(t *testing.T) {
	p := DefaultPainter{}
	canvas := &recordingCanvas{}
	cs := PaintState{
		YMin:       0,
		YMax:       100,
		MaxPoints:  60,
		Background: defaultBackground,
		Series: []Series{
			{Label: "CPU", Color: widget.ColorRed, Points: []DataPoint{
				{Value: -50}, // below min
				{Value: 200}, // above max
			}},
		},
	}
	bounds := geometry.NewRect(0, 0, 400, 200)
	p.PaintChart(canvas, bounds, cs)

	// Should draw 1 line segment, clamped to bounds.
	if canvas.lineCount != 1 {
		t.Errorf("lineCount = %d, want 1", canvas.lineCount)
	}
}

// --- FormatLabel Tests ---

func TestFormatLabel_Percentage(t *testing.T) {
	tests := []struct {
		value float64
		want  string
	}{
		{0, "0%"},
		{50, "50%"},
		{100, "100%"},
		{25.7, "26%"},
	}
	for _, tc := range tests {
		got := formatLabel(tc.value, 0, 100)
		if got != tc.want {
			t.Errorf("formatLabel(%v, 0, 100) = %q, want %q", tc.value, got, tc.want)
		}
	}
}

func TestFormatLabel_Decimal(t *testing.T) {
	tests := []struct {
		value float64
		yMin  float64
		yMax  float64
		want  string
	}{
		{0, -10, 200, "0.0"},
		{150.5, 0, 500, "150.5"},
		{-5, -100, 100, "-5.0"},
	}
	for _, tc := range tests {
		got := formatLabel(tc.value, tc.yMin, tc.yMax)
		if got != tc.want {
			t.Errorf("formatLabel(%v, %v, %v) = %q, want %q", tc.value, tc.yMin, tc.yMax, got, tc.want)
		}
	}
}

// --- YForValue Tests ---

func TestYForValue(t *testing.T) {
	plotArea := geometry.NewRect(0, 0, 400, 200)

	tests := []struct {
		name  string
		value float64
		yMin  float64
		yMax  float64
		wantY float32
	}{
		{"min value at bottom", 0, 0, 100, 200},
		{"max value at top", 100, 0, 100, 0},
		{"mid value at center", 50, 0, 100, 100},
		{"below min clamped", -50, 0, 100, 200},
		{"above max clamped", 200, 0, 100, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			yRange := tc.yMax - tc.yMin
			got := yForValue(tc.value, plotArea, tc.yMin, yRange)
			if got != tc.wantY {
				t.Errorf("yForValue(%v) = %v, want %v", tc.value, got, tc.wantY)
			}
		})
	}
}

// --- ComputePlotArea Tests ---

func TestComputePlotArea_WithoutLabels(t *testing.T) {
	bounds := geometry.NewRect(10, 20, 400, 200)
	plotArea := computePlotArea(bounds, false)

	if plotArea != bounds {
		t.Errorf("without labels, plot area should equal bounds: got %v, want %v", plotArea, bounds)
	}
}

func TestComputePlotArea_WithLabels(t *testing.T) {
	bounds := geometry.NewRect(10, 20, 400, 200)
	plotArea := computePlotArea(bounds, true)

	expectedX := bounds.Min.X + labelAreaWidth
	if plotArea.Min.X != expectedX {
		t.Errorf("plotArea.Min.X = %v, want %v", plotArea.Min.X, expectedX)
	}
	if plotArea.Width() != bounds.Width()-labelAreaWidth {
		t.Errorf("plotArea width = %v, want %v", plotArea.Width(), bounds.Width()-labelAreaWidth)
	}
}

// --- Interface Compliance Tests ---

func TestWidget_ImplementsWidget(t *testing.T) {
	var _ widget.Widget = (*Widget)(nil)
}

func TestWidget_ImplementsLifecycle(t *testing.T) {
	var _ widget.Lifecycle = (*Widget)(nil)
}

// --- Mock Types ---

type recordingCanvas struct {
	drawCount int
	rectCount int
	lineCount int
	textCount int
	clipCount int
}

func (c *recordingCanvas) Clear(_ widget.Color)                                     {}
func (c *recordingCanvas) DrawRect(_ geometry.Rect, _ widget.Color)                 { c.drawCount++; c.rectCount++ }
func (c *recordingCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32)    { c.drawCount++ }
func (c *recordingCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32) { c.drawCount++ }
func (c *recordingCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {
	c.drawCount++
}
func (c *recordingCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color) { c.drawCount++ }
func (c *recordingCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {
	c.drawCount++
}
func (c *recordingCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32) {
	c.drawCount++
	c.lineCount++
}
func (c *recordingCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
	c.drawCount++
	c.textCount++
}

func (c *recordingCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}
func (c *recordingCanvas) DrawImage(_ image.Image, _ geometry.Point) { c.drawCount++ }
func (c *recordingCanvas) PushClip(_ geometry.Rect)                  { c.clipCount++ }
func (c *recordingCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {
	c.clipCount++
}
func (c *recordingCanvas) PopClip()                        { c.clipCount++ }
func (c *recordingCanvas) PushTransform(_ geometry.Point)  {}
func (c *recordingCanvas) PopTransform()                   {}
func (c *recordingCanvas) TransformOffset() geometry.Point { return geometry.Point{} }

type mockPainter struct {
	called bool
}

func (p *mockPainter) PaintChart(_ widget.Canvas, _ geometry.Rect, _ PaintState) {
	p.called = true
}

type mockScheduler struct {
	dirtyCount int
}

func (s *mockScheduler) MarkDirty(_ widget.Widget) {
	s.dirtyCount++
}
