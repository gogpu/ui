package material3

import (
	"image"
	"testing"

	"github.com/gogpu/ui/core/collapsible"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

func TestCollapsiblePainter_CompileTimeCheck(t *testing.T) {
	var _ collapsible.Painter = CollapsiblePainter{}
}

func TestCollapsiblePainter_EmptyBounds(t *testing.T) {
	p := CollapsiblePainter{}
	canvas := &colMockCanvas{}

	p.PaintHeader(canvas, collapsible.HeaderState{})

	if canvas.drawCount > 0 {
		t.Error("should not draw anything with empty bounds")
	}
}

func TestCollapsiblePainter_NilTheme_Collapsed(t *testing.T) {
	p := CollapsiblePainter{Theme: nil}
	canvas := &colMockCanvas{}

	s := collapsible.HeaderState{
		Title:         "Section",
		Expanded:      false,
		Bounds:        geometry.NewRect(0, 0, 300, 40),
		ArrowProgress: 0,
	}

	p.PaintHeader(canvas, s)

	if canvas.drawCount == 0 {
		t.Error("should draw with nil theme (default colors)")
	}
	// Should have DrawRoundRect (bg) + DrawLine (arrow) + DrawText (title).
	if canvas.drawRoundRectCount == 0 {
		t.Error("should draw header background (DrawRoundRect)")
	}
	if canvas.drawLineCount == 0 {
		t.Error("should draw arrow indicator (DrawLine)")
	}
	if canvas.drawTextCount == 0 {
		t.Error("should draw title text (DrawText)")
	}
}

func TestCollapsiblePainter_NilTheme_Expanded(t *testing.T) {
	p := CollapsiblePainter{}
	canvas := &colMockCanvas{}

	s := collapsible.HeaderState{
		Title:         "Section",
		Expanded:      true,
		Bounds:        geometry.NewRect(0, 0, 300, 40),
		ArrowProgress: 1.0,
	}

	p.PaintHeader(canvas, s)

	if canvas.drawCount == 0 {
		t.Error("expanded header should draw")
	}
}

func TestCollapsiblePainter_WithTheme(t *testing.T) {
	theme := New(widget.Hex(0x6750A4))
	p := CollapsiblePainter{Theme: theme}
	canvas := &colMockCanvas{}

	s := collapsible.HeaderState{
		Title:         "Section",
		Bounds:        geometry.NewRect(0, 0, 300, 40),
		ArrowProgress: 0.5,
	}

	p.PaintHeader(canvas, s)

	if canvas.drawCount == 0 {
		t.Error("should draw with theme")
	}
}

func TestCollapsiblePainter_Hovered(t *testing.T) {
	p := CollapsiblePainter{}
	canvas := &colMockCanvas{}

	s := collapsible.HeaderState{
		Title:         "Section",
		Bounds:        geometry.NewRect(0, 0, 300, 40),
		Hovered:       true,
		ArrowProgress: 0,
	}

	p.PaintHeader(canvas, s)

	if canvas.drawCount == 0 {
		t.Error("hovered header should draw")
	}
}

func TestCollapsiblePainter_Pressed(t *testing.T) {
	p := CollapsiblePainter{}
	canvas := &colMockCanvas{}

	s := collapsible.HeaderState{
		Title:         "Section",
		Bounds:        geometry.NewRect(0, 0, 300, 40),
		Pressed:       true,
		ArrowProgress: 0,
	}

	p.PaintHeader(canvas, s)

	if canvas.drawCount == 0 {
		t.Error("pressed header should draw")
	}
}

func TestCollapsiblePainter_Focused(t *testing.T) {
	p := CollapsiblePainter{}
	canvas := &colMockCanvas{}

	s := collapsible.HeaderState{
		Title:         "Section",
		Bounds:        geometry.NewRect(0, 0, 300, 40),
		Focused:       true,
		ArrowProgress: 0,
	}

	p.PaintHeader(canvas, s)

	if canvas.strokeRoundRectCount == 0 {
		t.Error("focused header should draw focus ring (StrokeRoundRect)")
	}
}

func TestCollapsiblePainter_CustomColors(t *testing.T) {
	p := CollapsiblePainter{}
	canvas := &colMockCanvas{}

	s := collapsible.HeaderState{
		Title:         "Section",
		Bounds:        geometry.NewRect(0, 0, 300, 40),
		ArrowProgress: 0,
		HeaderColor:   widget.ColorRed,
		ArrowColor:    widget.ColorBlue,
	}

	p.PaintHeader(canvas, s)

	if canvas.drawCount == 0 {
		t.Error("should draw with custom colors")
	}
}

func TestCollapsiblePainter_EmptyTitle(t *testing.T) {
	p := CollapsiblePainter{}
	canvas := &colMockCanvas{}

	s := collapsible.HeaderState{
		Title:         "",
		Bounds:        geometry.NewRect(0, 0, 300, 40),
		ArrowProgress: 0,
	}

	p.PaintHeader(canvas, s)

	if canvas.drawTextCount > 0 {
		t.Error("should not draw text for empty title")
	}
}

func TestM3ApplyCollapsibleState(t *testing.T) {
	base := widget.ColorRed

	normal := m3ApplyCollapsibleState(base, false, false)
	if normal != base {
		t.Error("normal state should return base color unchanged")
	}

	hovered := m3ApplyCollapsibleState(base, true, false)
	if hovered == base {
		t.Error("hover state should modify color")
	}

	pressed := m3ApplyCollapsibleState(base, false, true)
	if pressed == base {
		t.Error("pressed state should modify color")
	}

	// Pressed takes precedence.
	both := m3ApplyCollapsibleState(base, true, true)
	if both != pressed {
		t.Error("pressed should take precedence over hover")
	}
}

// --- colMockCanvas records basic draw calls ---

type colMockCanvas struct {
	drawCount            int
	drawRoundRectCount   int
	strokeRoundRectCount int
	drawLineCount        int
	drawTextCount        int
}

func (c *colMockCanvas) Clear(_ widget.Color)                                  {}
func (c *colMockCanvas) DrawRect(_ geometry.Rect, _ widget.Color)              { c.drawCount++ }
func (c *colMockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) { c.drawCount++ }
func (c *colMockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32) {
	c.drawCount++
	c.drawRoundRectCount++
}
func (c *colMockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {
	c.drawCount++
	c.strokeRoundRectCount++
}
func (c *colMockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color) { c.drawCount++ }
func (c *colMockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {
	c.drawCount++
}
func (c *colMockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32) {
	c.drawCount++
	c.drawLineCount++
}
func (c *colMockCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
	c.drawCount++
	c.drawTextCount++
}

func (c *colMockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}
func (c *colMockCanvas) DrawImage(_ image.Image, _ geometry.Point)    { c.drawCount++ }
func (c *colMockCanvas) PushClip(_ geometry.Rect)                     {}
func (c *colMockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *colMockCanvas) PopClip()                                     {}
func (c *colMockCanvas) PushTransform(_ geometry.Point)               {}
func (c *colMockCanvas) PopTransform()                                {}
func (c *colMockCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }
