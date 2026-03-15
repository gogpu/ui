package material3

import (
	"image"
	"testing"

	"github.com/gogpu/ui/core/progressbar"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

func TestProgressBarPainter_CompileTimeCheck(t *testing.T) {
	var _ progressbar.Painter = ProgressBarPainter{}
}

func TestProgressBarPainter_EmptyBounds(t *testing.T) {
	p := ProgressBarPainter{}
	canvas := &pbMockCanvas{}

	p.PaintProgressBar(canvas, progressbar.PaintState{})

	if canvas.drawCount > 0 {
		t.Error("should not draw anything with empty bounds")
	}
}

func TestProgressBarPainter_NilTheme_UsesDefaults(t *testing.T) {
	p := ProgressBarPainter{Theme: nil}
	canvas := &pbMockCanvas{}

	ps := progressbar.PaintState{
		Value:     0.5,
		Bounds:    geometry.NewRect(0, 0, 200, 20),
		BarHeight: 8,
		Radius:    4,
	}

	p.PaintProgressBar(canvas, ps)

	if canvas.drawCount == 0 {
		t.Error("should draw with nil theme (default colors)")
	}
}

func TestProgressBarPainter_WithTheme(t *testing.T) {
	theme := New(widget.Hex(0x6750A4))
	p := ProgressBarPainter{Theme: theme}
	canvas := &pbMockCanvas{}

	ps := progressbar.PaintState{
		Value:     0.75,
		Bounds:    geometry.NewRect(0, 0, 200, 20),
		BarHeight: 8,
		Radius:    4,
	}

	p.PaintProgressBar(canvas, ps)

	if canvas.drawCount == 0 {
		t.Error("should draw with theme")
	}
}

func TestProgressBarPainter_Disabled(t *testing.T) {
	p := ProgressBarPainter{}
	canvas := &pbMockCanvas{}

	ps := progressbar.PaintState{
		Value:     0.5,
		Bounds:    geometry.NewRect(0, 0, 200, 20),
		BarHeight: 8,
		Radius:    4,
		Disabled:  true,
	}

	p.PaintProgressBar(canvas, ps)

	if canvas.drawCount == 0 {
		t.Error("disabled progress bar should still draw")
	}
}

func TestProgressBarPainter_ZeroValue(t *testing.T) {
	p := ProgressBarPainter{}
	canvas := &pbMockCanvas{}

	ps := progressbar.PaintState{
		Value:     0,
		Bounds:    geometry.NewRect(0, 0, 200, 20),
		BarHeight: 8,
		Radius:    4,
	}

	// Should not panic.
	p.PaintProgressBar(canvas, ps)

	// Only track should be drawn, no fill.
	if canvas.drawRoundRectCount != 1 {
		t.Errorf("zero value should draw 1 DrawRoundRect (track only), got %d", canvas.drawRoundRectCount)
	}
}

func TestProgressBarPainter_FullValue(t *testing.T) {
	p := ProgressBarPainter{}
	canvas := &pbMockCanvas{}

	ps := progressbar.PaintState{
		Value:     1.0,
		Bounds:    geometry.NewRect(0, 0, 200, 20),
		BarHeight: 8,
		Radius:    4,
	}

	// Should not panic.
	p.PaintProgressBar(canvas, ps)

	// Track + fill.
	if canvas.drawRoundRectCount < 2 {
		t.Errorf("full value should draw at least 2 DrawRoundRect (track + fill), got %d", canvas.drawRoundRectCount)
	}
}

func TestProgressBarPainter_WithLabel(t *testing.T) {
	p := ProgressBarPainter{}
	canvas := &pbMockCanvas{}

	ps := progressbar.PaintState{
		Value:     0.5,
		Bounds:    geometry.NewRect(0, 0, 200, 20),
		BarHeight: 8,
		Radius:    4,
		ShowLabel: true,
		Label:     "50%",
	}

	p.PaintProgressBar(canvas, ps)

	if canvas.drawTextCount != 1 {
		t.Errorf("should draw 1 DrawText for label, got %d", canvas.drawTextCount)
	}
}

func TestProgressBarPainter_WithColorScheme(t *testing.T) {
	p := ProgressBarPainter{}
	canvas := &pbMockCanvas{}

	scheme := progressbar.ProgressBarColorScheme{
		Bar:           widget.ColorRed,
		Track:         widget.ColorGray,
		Label:         widget.ColorWhite,
		DisabledBar:   widget.ColorDarkGray,
		DisabledTrack: widget.ColorLightGray,
	}

	ps := progressbar.PaintState{
		Value:                  0.5,
		Bounds:                 geometry.NewRect(0, 0, 200, 20),
		BarHeight:              8,
		Radius:                 4,
		ProgressBarColorScheme: scheme,
	}

	p.PaintProgressBar(canvas, ps)

	if canvas.drawCount == 0 {
		t.Error("should draw with custom color scheme")
	}
}

func TestProgressBarPainter_ResolveColors_NilTheme(t *testing.T) {
	p := ProgressBarPainter{Theme: nil}
	colors := p.resolveColors()

	if colors == (progressbar.ProgressBarColorScheme{}) {
		t.Error("should return non-zero default colors")
	}
}

func TestProgressBarPainter_ResolveColors_WithTheme(t *testing.T) {
	theme := New(widget.Hex(0xFF0000))
	p := ProgressBarPainter{Theme: theme}
	colors := p.resolveColors()

	if colors == (progressbar.ProgressBarColorScheme{}) {
		t.Error("should return non-zero colors from theme")
	}
	if colors.Bar == m3DefaultProgressBarColors.Bar {
		t.Error("themed colors should differ from default purple")
	}
}

// --- pbMockCanvas records basic draw calls ---

type pbMockCanvas struct {
	drawCount          int
	drawRoundRectCount int
	drawTextCount      int
}

func (c *pbMockCanvas) Clear(_ widget.Color)                                  {}
func (c *pbMockCanvas) DrawRect(_ geometry.Rect, _ widget.Color)              { c.drawCount++ }
func (c *pbMockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) { c.drawCount++ }
func (c *pbMockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32) {
	c.drawCount++
	c.drawRoundRectCount++
}
func (c *pbMockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {
	c.drawCount++
}
func (c *pbMockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color) { c.drawCount++ }
func (c *pbMockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {
	c.drawCount++
}
func (c *pbMockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32) { c.drawCount++ }
func (c *pbMockCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
	c.drawCount++
	c.drawTextCount++
}

func (c *pbMockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}
func (c *pbMockCanvas) DrawImage(_ image.Image, _ geometry.Point)    { c.drawCount++ }
func (c *pbMockCanvas) PushClip(_ geometry.Rect)                     {}
func (c *pbMockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *pbMockCanvas) PopClip()                                     {}
func (c *pbMockCanvas) PushTransform(_ geometry.Point)               {}
func (c *pbMockCanvas) PopTransform()                                {}
func (c *pbMockCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }
