package material3

import (
	"image"
	"testing"

	"github.com/gogpu/ui/core/popover"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

func TestPopoverPainter_CompileTimeCheck(t *testing.T) {
	var _ popover.Painter = PopoverPainter{}
}

func TestPopoverPainter_Popover_EmptyBounds(t *testing.T) {
	p := PopoverPainter{}
	canvas := &popMockCanvas{}

	p.PaintPopover(canvas, &popover.PopoverPaintState{})

	if canvas.drawCount > 0 {
		t.Error("should not draw anything with empty bounds")
	}
}

func TestPopoverPainter_Tooltip_EmptyBounds(t *testing.T) {
	p := PopoverPainter{}
	canvas := &popMockCanvas{}

	p.PaintTooltip(canvas, &popover.TooltipPaintState{})

	if canvas.drawCount > 0 {
		t.Error("should not draw anything with empty bounds")
	}
}

func TestPopoverPainter_Popover_NilTheme(t *testing.T) {
	p := PopoverPainter{Theme: nil}
	canvas := &popMockCanvas{}

	p.PaintPopover(canvas, &popover.PopoverPaintState{
		Bounds: geometry.NewRect(100, 100, 200, 150),
	})

	if canvas.drawCount == 0 {
		t.Error("should draw with nil theme (default colors)")
	}
	// Should have DrawRoundRect (shadow + bg) + StrokeRoundRect (border).
	if canvas.drawRoundRectCount < 2 {
		t.Errorf("popover should draw at least 2 DrawRoundRect (shadow + bg), got %d", canvas.drawRoundRectCount)
	}
	if canvas.strokeRoundRectCount == 0 {
		t.Error("popover should draw border (StrokeRoundRect)")
	}
}

func TestPopoverPainter_Popover_WithTheme(t *testing.T) {
	theme := New(widget.Hex(0x6750A4))
	p := PopoverPainter{Theme: theme}
	canvas := &popMockCanvas{}

	p.PaintPopover(canvas, &popover.PopoverPaintState{
		Bounds: geometry.NewRect(100, 100, 200, 150),
	})

	if canvas.drawCount == 0 {
		t.Error("should draw with theme")
	}
}

func TestPopoverPainter_Popover_WithColorScheme(t *testing.T) {
	p := PopoverPainter{}
	canvas := &popMockCanvas{}

	p.PaintPopover(canvas, &popover.PopoverPaintState{
		Bounds: geometry.NewRect(100, 100, 200, 150),
		ColorScheme: popover.PopoverColorScheme{
			Background: widget.ColorWhite,
			Border:     widget.ColorBlack,
			Shadow:     widget.RGBA(0, 0, 0, 0.2),
			ShadowBlur: 8,
		},
	})

	if canvas.drawCount == 0 {
		t.Error("should draw with custom color scheme")
	}
}

func TestPopoverPainter_Tooltip_NilTheme(t *testing.T) {
	p := PopoverPainter{Theme: nil}
	canvas := &popMockCanvas{}

	p.PaintTooltip(canvas, &popover.TooltipPaintState{
		Bounds: geometry.NewRect(100, 100, 120, 30),
		Text:   "Tooltip text",
	})

	if canvas.drawCount == 0 {
		t.Error("should draw with nil theme")
	}
	if canvas.drawRoundRectCount == 0 {
		t.Error("tooltip should draw background (DrawRoundRect)")
	}
	if canvas.drawTextCount == 0 {
		t.Error("tooltip should draw text (DrawText)")
	}
}

func TestPopoverPainter_Tooltip_WithTheme(t *testing.T) {
	theme := New(widget.Hex(0xFF0000))
	p := PopoverPainter{Theme: theme}
	canvas := &popMockCanvas{}

	p.PaintTooltip(canvas, &popover.TooltipPaintState{
		Bounds: geometry.NewRect(100, 100, 120, 30),
		Text:   "Tooltip text",
	})

	if canvas.drawCount == 0 {
		t.Error("should draw with theme")
	}
}

func TestPopoverPainter_Tooltip_WithColorScheme(t *testing.T) {
	p := PopoverPainter{}
	canvas := &popMockCanvas{}

	p.PaintTooltip(canvas, &popover.TooltipPaintState{
		Bounds: geometry.NewRect(100, 100, 120, 30),
		Text:   "Tooltip text",
		ColorScheme: popover.TooltipColorScheme{
			Background: widget.ColorBlack,
			TextColor:  widget.ColorWhite,
			Border:     widget.ColorGray,
		},
	})

	if canvas.drawCount == 0 {
		t.Error("should draw with custom tooltip color scheme")
	}
}

// --- popMockCanvas records basic draw calls ---

type popMockCanvas struct {
	drawCount            int
	drawRoundRectCount   int
	strokeRoundRectCount int
	drawTextCount        int
}

func (c *popMockCanvas) Clear(_ widget.Color)                                  {}
func (c *popMockCanvas) DrawRect(_ geometry.Rect, _ widget.Color)              { c.drawCount++ }
func (c *popMockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) { c.drawCount++ }
func (c *popMockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32) {
	c.drawCount++
	c.drawRoundRectCount++
}
func (c *popMockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {
	c.drawCount++
	c.strokeRoundRectCount++
}
func (c *popMockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color) { c.drawCount++ }
func (c *popMockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {
	c.drawCount++
}
func (c *popMockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32) { c.drawCount++ }
func (c *popMockCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
	c.drawCount++
	c.drawTextCount++
}

func (c *popMockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}
func (c *popMockCanvas) DrawImage(_ image.Image, _ geometry.Point)    { c.drawCount++ }
func (c *popMockCanvas) PushClip(_ geometry.Rect)                     {}
func (c *popMockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *popMockCanvas) PopClip()                                     {}
func (c *popMockCanvas) PushTransform(_ geometry.Point)               {}
func (c *popMockCanvas) PopTransform()                                {}
func (c *popMockCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }
