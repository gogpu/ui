package material3

import (
	"github.com/gogpu/gg/scene"
	"image"
	"testing"

	"github.com/gogpu/ui/core/linechart"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

func TestLineChartPainter_CompileTimeCheck(t *testing.T) {
	var _ linechart.Painter = LineChartPainter{}
}

func TestLineChartPainter_EmptyBounds(t *testing.T) {
	p := LineChartPainter{}
	canvas := &chartMockCanvas{}

	p.PaintChart(canvas, linechart.PaintState{})

	if canvas.drawCount > 0 {
		t.Error("should not draw anything with empty bounds")
	}
}

func TestLineChartPainter_NilTheme_UsesDefaults(t *testing.T) {
	p := LineChartPainter{Theme: nil}
	canvas := &chartMockCanvas{}

	state := linechart.PaintState{
		Series:    []linechart.Series{{Color: widget.ColorRed, Points: []linechart.DataPoint{{Value: 1}, {Value: 2}}}},
		MaxPoints: 10,
		YMin:      0,
		YMax:      10,
		ShowGrid:  true,
	}

	state.Bounds = geometry.NewRect(0, 0, 200, 100)
	p.PaintChart(canvas, state)

	if canvas.drawCount == 0 {
		t.Error("should draw with nil theme (default colors)")
	}
}

func TestLineChartPainter_WithTheme(t *testing.T) {
	theme := New(widget.Hex(0x6750A4))
	p := LineChartPainter{Theme: theme}
	canvas := &chartMockCanvas{}

	state := linechart.PaintState{
		Series:    []linechart.Series{{Color: widget.ColorBlue, Points: []linechart.DataPoint{{Value: 1}, {Value: 5}}}},
		MaxPoints: 10,
		YMin:      0,
		YMax:      10,
	}

	state.Bounds = geometry.NewRect(0, 0, 200, 100)
	p.PaintChart(canvas, state)

	if canvas.drawCount == 0 {
		t.Error("should draw with theme")
	}
}

func TestLineChartPainter_WithGrid(t *testing.T) {
	p := LineChartPainter{}
	canvas := &chartMockCanvas{}

	state := linechart.PaintState{
		ShowGrid:  true,
		MaxPoints: 10,
		YMin:      0,
		YMax:      100,
	}

	state.Bounds = geometry.NewRect(0, 0, 200, 100)
	p.PaintChart(canvas, state)

	// Grid draws m3ChartGridDivisions+1 lines.
	if canvas.lineCount < m3ChartGridDivisions+1 {
		t.Errorf("grid should draw at least %d lines, got %d", m3ChartGridDivisions+1, canvas.lineCount)
	}
}

func TestLineChartPainter_WithLabels(t *testing.T) {
	p := LineChartPainter{}
	canvas := &chartMockCanvas{}

	state := linechart.PaintState{
		ShowLabels: true,
		MaxPoints:  10,
		YMin:       0,
		YMax:       100,
	}

	state.Bounds = geometry.NewRect(0, 0, 200, 100)
	p.PaintChart(canvas, state)

	if canvas.textCount < m3ChartGridDivisions+1 {
		t.Errorf("labels should draw at least %d texts, got %d", m3ChartGridDivisions+1, canvas.textCount)
	}
}

func TestLineChartPainter_ResolveColors_NilTheme(t *testing.T) {
	p := LineChartPainter{Theme: nil}
	colors := p.resolveColors()

	if colors != m3DefaultLineChartColors {
		t.Error("nil theme should return default M3 chart colors")
	}
}

func TestLineChartPainter_ResolveColors_WithTheme(t *testing.T) {
	theme := New(widget.Hex(0xFF0000))
	p := LineChartPainter{Theme: theme}
	colors := p.resolveColors()

	if colors == m3DefaultLineChartColors {
		t.Error("themed colors should differ from default purple")
	}
	if colors.Background != theme.Colors.Surface {
		t.Errorf("Background = %v, want %v", colors.Background, theme.Colors.Surface)
	}
}

func TestLineChartPainter_NoSeries(t *testing.T) {
	p := LineChartPainter{}
	canvas := &chartMockCanvas{}

	state := linechart.PaintState{
		MaxPoints: 10,
		YMin:      0,
		YMax:      10,
	}

	// Should not panic.
	state.Bounds = geometry.NewRect(0, 0, 200, 100)
	p.PaintChart(canvas, state)
}

func TestLineChartPainter_SinglePoint(t *testing.T) {
	p := LineChartPainter{}
	canvas := &chartMockCanvas{}

	state := linechart.PaintState{
		Series:    []linechart.Series{{Color: widget.ColorRed, Points: []linechart.DataPoint{{Value: 5}}}},
		MaxPoints: 10,
		YMin:      0,
		YMax:      10,
	}

	// Should not draw lines for single point.
	state.Bounds = geometry.NewRect(0, 0, 200, 100)
	p.PaintChart(canvas, state)
}

// --- chartMockCanvas ---

type chartMockCanvas struct {
	drawCount int
	lineCount int
	textCount int
}

func (c *chartMockCanvas) Clear(_ widget.Color)                                     {}
func (c *chartMockCanvas) DrawRect(_ geometry.Rect, _ widget.Color)                 { c.drawCount++ }
func (c *chartMockCanvas) FillRectDirect(_ geometry.Rect, _ widget.Color)           {}
func (c *chartMockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32)    { c.drawCount++ }
func (c *chartMockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32) { c.drawCount++ }
func (c *chartMockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {
	c.drawCount++
}
func (c *chartMockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color) { c.drawCount++ }
func (c *chartMockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {
	c.drawCount++
}
func (c *chartMockCanvas) StrokeArc(_ geometry.Point, _ float32, _, _ float64, _ widget.Color, _ float32) {
}
func (c *chartMockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32) {
	c.drawCount++
	c.lineCount++
}
func (c *chartMockCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
	c.drawCount++
	c.textCount++
}

func (c *chartMockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}
func (c *chartMockCanvas) DrawImage(_ image.Image, _ geometry.Point)    { c.drawCount++ }
func (c *chartMockCanvas) PushClip(_ geometry.Rect)                     {}
func (c *chartMockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *chartMockCanvas) PopClip()                                     {}
func (c *chartMockCanvas) PushTransform(_ geometry.Point)               {}
func (c *chartMockCanvas) PopTransform()                                {}
func (c *chartMockCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }
func (c *chartMockCanvas) ScreenOriginBase() geometry.Point             { return geometry.Point{} }
func (c *chartMockCanvas) ClipBounds() geometry.Rect                    { return geometry.NewRect(0, 0, 10000, 10000) }
func (c *chartMockCanvas) ReplayScene(_ *scene.Scene)                   {}
