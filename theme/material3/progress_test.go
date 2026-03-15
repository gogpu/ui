package material3

import (
	"image"
	"testing"

	"github.com/gogpu/ui/core/progress"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

func TestProgressPainter_CompileTimeCheck(t *testing.T) {
	var _ progress.Painter = ProgressPainter{}
}

func TestProgressPainter_EmptyBounds(t *testing.T) {
	p := ProgressPainter{}
	canvas := &cpMockCanvas{}

	p.PaintProgress(canvas, progress.PaintState{})

	if canvas.drawCount > 0 {
		t.Error("should not draw anything with empty bounds")
	}
}

func TestProgressPainter_Determinate_NilTheme(t *testing.T) {
	p := ProgressPainter{Theme: nil}
	canvas := &cpMockCanvas{}

	ps := progress.PaintState{
		Value:       0.5,
		Bounds:      geometry.NewRect(0, 0, 48, 48),
		Diameter:    40,
		StrokeWidth: 4,
	}

	p.PaintProgress(canvas, ps)

	if canvas.drawCount == 0 {
		t.Error("should draw with nil theme (default colors)")
	}
	// Should have StrokeCircle (track) + DrawLine calls (arc).
	if canvas.strokeCircleCount == 0 {
		t.Error("determinate should draw track circle (StrokeCircle)")
	}
}

func TestProgressPainter_Determinate_WithTheme(t *testing.T) {
	theme := New(widget.Hex(0x6750A4))
	p := ProgressPainter{Theme: theme}
	canvas := &cpMockCanvas{}

	ps := progress.PaintState{
		Value:       0.75,
		Bounds:      geometry.NewRect(0, 0, 48, 48),
		Diameter:    40,
		StrokeWidth: 4,
	}

	p.PaintProgress(canvas, ps)

	if canvas.drawCount == 0 {
		t.Error("should draw with theme")
	}
}

func TestProgressPainter_Indeterminate(t *testing.T) {
	p := ProgressPainter{}
	canvas := &cpMockCanvas{}

	ps := progress.PaintState{
		Bounds:        geometry.NewRect(0, 0, 48, 48),
		Diameter:      40,
		StrokeWidth:   4,
		Indeterminate: true,
		Rotation:      1.5,
	}

	p.PaintProgress(canvas, ps)

	if canvas.drawCount == 0 {
		t.Error("indeterminate should draw")
	}
	if canvas.strokeCircleCount == 0 {
		t.Error("indeterminate should draw track circle")
	}
}

func TestProgressPainter_Disabled(t *testing.T) {
	p := ProgressPainter{}
	canvas := &cpMockCanvas{}

	ps := progress.PaintState{
		Value:       0.5,
		Bounds:      geometry.NewRect(0, 0, 48, 48),
		Diameter:    40,
		StrokeWidth: 4,
		Disabled:    true,
	}

	p.PaintProgress(canvas, ps)

	if canvas.drawCount == 0 {
		t.Error("disabled should still draw")
	}
}

func TestProgressPainter_WithLabel(t *testing.T) {
	p := ProgressPainter{}
	canvas := &cpMockCanvas{}

	ps := progress.PaintState{
		Value:       0.5,
		Bounds:      geometry.NewRect(0, 0, 48, 48),
		Diameter:    40,
		StrokeWidth: 4,
		ShowLabel:   true,
		Label:       "50%",
	}

	p.PaintProgress(canvas, ps)

	if canvas.drawTextCount != 1 {
		t.Errorf("should draw 1 DrawText for label, got %d", canvas.drawTextCount)
	}
}

func TestProgressPainter_ZeroValue(t *testing.T) {
	p := ProgressPainter{}
	canvas := &cpMockCanvas{}

	ps := progress.PaintState{
		Value:       0,
		Bounds:      geometry.NewRect(0, 0, 48, 48),
		Diameter:    40,
		StrokeWidth: 4,
	}

	// Should not panic.
	p.PaintProgress(canvas, ps)

	// Only track, no arc.
	if canvas.drawLineCount > 0 {
		t.Error("zero value should not draw arc lines")
	}
}

func TestProgressPainter_ZeroRadius(t *testing.T) {
	p := ProgressPainter{}
	canvas := &cpMockCanvas{}

	ps := progress.PaintState{
		Value:       0.5,
		Bounds:      geometry.NewRect(0, 0, 4, 4),
		Diameter:    40,
		StrokeWidth: 40, // larger than diameter -> zero radius
	}

	// Should not panic and should not draw.
	p.PaintProgress(canvas, ps)
}

// --- cpMockCanvas records basic draw calls ---

type cpMockCanvas struct {
	drawCount         int
	strokeCircleCount int
	drawLineCount     int
	drawTextCount     int
}

func (c *cpMockCanvas) Clear(_ widget.Color)                                  {}
func (c *cpMockCanvas) DrawRect(_ geometry.Rect, _ widget.Color)              { c.drawCount++ }
func (c *cpMockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) { c.drawCount++ }
func (c *cpMockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32) {
	c.drawCount++
}
func (c *cpMockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {
	c.drawCount++
}
func (c *cpMockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color) { c.drawCount++ }
func (c *cpMockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {
	c.drawCount++
	c.strokeCircleCount++
}
func (c *cpMockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32) {
	c.drawCount++
	c.drawLineCount++
}
func (c *cpMockCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
	c.drawCount++
	c.drawTextCount++
}

func (c *cpMockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}
func (c *cpMockCanvas) DrawImage(_ image.Image, _ geometry.Point)    { c.drawCount++ }
func (c *cpMockCanvas) PushClip(_ geometry.Rect)                     {}
func (c *cpMockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *cpMockCanvas) PopClip()                                     {}
func (c *cpMockCanvas) PushTransform(_ geometry.Point)               {}
func (c *cpMockCanvas) PopTransform()                                {}
func (c *cpMockCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }
