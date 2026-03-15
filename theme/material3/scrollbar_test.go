package material3

import (
	"image"
	"testing"

	"github.com/gogpu/ui/core/scrollview"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

func TestScrollbarPainter_NilTheme(t *testing.T) {
	p := ScrollbarPainter{Theme: nil}
	canvas := &scrollbarMockCanvas{}

	ps := scrollview.PaintState{
		Bounds:         geometry.NewRect(0, 0, 200, 300),
		VScrollVisible: true,
		VThumbRect:     geometry.NewRect(188, 10, 8, 50),
		VTrackRect:     geometry.NewRect(186, 0, 12, 300),
	}

	p.PaintScrollbar(canvas, ps)

	if len(canvas.drawRoundRects) < 2 {
		t.Errorf("should draw track + thumb, got %d DrawRoundRect calls", len(canvas.drawRoundRects))
	}
}

func TestScrollbarPainter_WithTheme(t *testing.T) {
	theme := New(widget.Hex(0x6750A4)) // Default M3 purple
	p := ScrollbarPainter{Theme: theme}
	canvas := &scrollbarMockCanvas{}

	ps := scrollview.PaintState{
		Bounds:         geometry.NewRect(0, 0, 200, 300),
		VScrollVisible: true,
		VThumbRect:     geometry.NewRect(188, 10, 8, 50),
		VTrackRect:     geometry.NewRect(186, 0, 12, 300),
	}

	p.PaintScrollbar(canvas, ps)

	if len(canvas.drawRoundRects) < 2 {
		t.Errorf("should draw track + thumb with theme, got %d DrawRoundRect calls", len(canvas.drawRoundRects))
	}
}

func TestScrollbarPainter_EmptyBounds(t *testing.T) {
	p := ScrollbarPainter{}
	canvas := &scrollbarMockCanvas{}

	ps := scrollview.PaintState{
		Bounds: geometry.Rect{},
	}

	p.PaintScrollbar(canvas, ps)

	if len(canvas.drawRoundRects) != 0 {
		t.Error("should not draw anything for empty bounds")
	}
}

func TestScrollbarPainter_WithColorScheme(t *testing.T) {
	p := ScrollbarPainter{}
	canvas := &scrollbarMockCanvas{}

	cs := scrollview.ScrollbarColorScheme{
		Track:      widget.RGBA(0.1, 0.1, 0.1, 1),
		Thumb:      widget.RGBA(0.5, 0.5, 0.5, 1),
		ThumbHover: widget.RGBA(0.6, 0.6, 0.6, 1),
		ThumbDrag:  widget.RGBA(0.7, 0.7, 0.7, 1),
	}

	ps := scrollview.PaintState{
		Bounds:         geometry.NewRect(0, 0, 200, 300),
		VScrollVisible: true,
		VThumbRect:     geometry.NewRect(188, 10, 8, 50),
		VTrackRect:     geometry.NewRect(186, 0, 12, 300),
		ColorScheme:    cs,
	}

	p.PaintScrollbar(canvas, ps)

	if len(canvas.drawRoundRects) < 2 {
		t.Error("should draw with provided color scheme")
	}
}

func TestScrollbarPainter_ResolveColors_NilTheme(t *testing.T) {
	p := ScrollbarPainter{Theme: nil}
	colors := p.resolveColors()

	if colors.Track.A <= 0 {
		t.Error("default track color should have alpha > 0")
	}
	if colors.Thumb.A <= 0 {
		t.Error("default thumb color should have alpha > 0")
	}
}

func TestScrollbarPainter_ResolveColors_WithTheme(t *testing.T) {
	theme := New(widget.Hex(0xFF0000)) // Red seed
	p := ScrollbarPainter{Theme: theme}
	colors := p.resolveColors()

	if colors.Track.A <= 0 {
		t.Error("theme track color should have alpha > 0")
	}
	if colors.Thumb.A <= 0 {
		t.Error("theme thumb color should have alpha > 0")
	}
}

func TestScrollbarPainter_BothScrollbars(t *testing.T) {
	p := ScrollbarPainter{}
	canvas := &scrollbarMockCanvas{}

	ps := scrollview.PaintState{
		Bounds:         geometry.NewRect(0, 0, 200, 300),
		VScrollVisible: true,
		VThumbRect:     geometry.NewRect(188, 10, 8, 50),
		VTrackRect:     geometry.NewRect(186, 0, 12, 288),
		HScrollVisible: true,
		HThumbRect:     geometry.NewRect(10, 288, 50, 8),
		HTrackRect:     geometry.NewRect(0, 286, 186, 12),
	}

	p.PaintScrollbar(canvas, ps)

	if len(canvas.drawRoundRects) < 4 {
		t.Errorf("should draw 4 rects (2 tracks + 2 thumbs), got %d", len(canvas.drawRoundRects))
	}
}

// Compile-time check.
func TestScrollbarPainter_ImplementsPainter(t *testing.T) {
	var _ scrollview.Painter = ScrollbarPainter{}
}

// --- scrollbarMockCanvas ---

type scrollbarMockCanvas struct {
	drawRoundRects []scrollbarDrawRoundRectCall
}

type scrollbarDrawRoundRectCall struct {
	r      geometry.Rect
	color  widget.Color
	radius float32
}

func (c *scrollbarMockCanvas) Clear(_ widget.Color)                                  {}
func (c *scrollbarMockCanvas) DrawRect(_ geometry.Rect, _ widget.Color)              {}
func (c *scrollbarMockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) {}

func (c *scrollbarMockCanvas) DrawRoundRect(r geometry.Rect, color widget.Color, radius float32) {
	c.drawRoundRects = append(c.drawRoundRects, scrollbarDrawRoundRectCall{r: r, color: color, radius: radius})
}

func (c *scrollbarMockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {
}
func (c *scrollbarMockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)              {}
func (c *scrollbarMockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {}
func (c *scrollbarMockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32)             {}
func (c *scrollbarMockCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
}

func (c *scrollbarMockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}
func (c *scrollbarMockCanvas) DrawImage(_ image.Image, _ geometry.Point)    {}
func (c *scrollbarMockCanvas) PushClip(_ geometry.Rect)                     {}
func (c *scrollbarMockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *scrollbarMockCanvas) PopClip()                                     {}
func (c *scrollbarMockCanvas) PushTransform(_ geometry.Point)               {}
func (c *scrollbarMockCanvas) PopTransform()                                {}
func (c *scrollbarMockCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }
