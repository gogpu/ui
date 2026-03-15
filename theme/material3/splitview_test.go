package material3

import (
	"image"
	"testing"

	"github.com/gogpu/ui/core/splitview"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

func TestSplitViewPainter_CompileTimeCheck(t *testing.T) {
	var _ splitview.Painter = SplitViewPainter{}
}

func TestSplitViewPainter_EmptyBounds(t *testing.T) {
	p := SplitViewPainter{}
	canvas := &svMockCanvas{}

	p.PaintDivider(canvas, splitview.PaintState{})

	if canvas.drawCount > 0 {
		t.Error("should not draw anything with empty bounds")
	}
}

func TestSplitViewPainter_NilTheme_Horizontal(t *testing.T) {
	p := SplitViewPainter{Theme: nil}
	canvas := &svMockCanvas{}

	ps := splitview.PaintState{
		DividerRect: geometry.NewRect(195, 0, 10, 400),
		Orientation: splitview.Horizontal,
	}

	p.PaintDivider(canvas, ps)

	if canvas.drawCount == 0 {
		t.Error("should draw with nil theme (default colors)")
	}
	if canvas.drawRectCount == 0 {
		t.Error("should draw divider background (DrawRect)")
	}
	if canvas.drawCircleCount == 0 {
		t.Error("should draw handle dots (DrawCircle)")
	}
}

func TestSplitViewPainter_NilTheme_Vertical(t *testing.T) {
	p := SplitViewPainter{}
	canvas := &svMockCanvas{}

	ps := splitview.PaintState{
		DividerRect: geometry.NewRect(0, 195, 400, 10),
		Orientation: splitview.Vertical,
	}

	p.PaintDivider(canvas, ps)

	if canvas.drawCount == 0 {
		t.Error("vertical divider should draw")
	}
	// Vertical orientation should draw horizontal dots.
	if canvas.drawCircleCount < 3 {
		t.Errorf("should draw 3 handle dots, got %d", canvas.drawCircleCount)
	}
}

func TestSplitViewPainter_WithTheme(t *testing.T) {
	theme := New(widget.Hex(0x6750A4))
	p := SplitViewPainter{Theme: theme}
	canvas := &svMockCanvas{}

	ps := splitview.PaintState{
		DividerRect: geometry.NewRect(195, 0, 10, 400),
		Orientation: splitview.Horizontal,
	}

	p.PaintDivider(canvas, ps)

	if canvas.drawCount == 0 {
		t.Error("should draw with theme")
	}
}

func TestSplitViewPainter_Hovered(t *testing.T) {
	p := SplitViewPainter{}
	canvas := &svMockCanvas{}

	ps := splitview.PaintState{
		DividerRect: geometry.NewRect(195, 0, 10, 400),
		Orientation: splitview.Horizontal,
		Hovered:     true,
	}

	p.PaintDivider(canvas, ps)

	if canvas.drawCount == 0 {
		t.Error("hovered divider should draw")
	}
}

func TestSplitViewPainter_Dragging(t *testing.T) {
	p := SplitViewPainter{}
	canvas := &svMockCanvas{}

	ps := splitview.PaintState{
		DividerRect: geometry.NewRect(195, 0, 10, 400),
		Orientation: splitview.Horizontal,
		Dragging:    true,
	}

	p.PaintDivider(canvas, ps)

	if canvas.drawCount == 0 {
		t.Error("dragging divider should draw")
	}
}

func TestSplitViewPainter_WithColorScheme(t *testing.T) {
	p := SplitViewPainter{}
	canvas := &svMockCanvas{}

	scheme := splitview.DividerColorScheme{
		Divider:      widget.ColorGray,
		DividerHover: widget.ColorLightGray,
		DividerDrag:  widget.ColorBlue,
		Handle:       widget.ColorBlack,
	}

	ps := splitview.PaintState{
		DividerRect: geometry.NewRect(195, 0, 10, 400),
		Orientation: splitview.Horizontal,
		ColorScheme: scheme,
	}

	p.PaintDivider(canvas, ps)

	if canvas.drawCount == 0 {
		t.Error("should draw with custom color scheme")
	}
}

func TestSplitViewPainter_ResolveColors_NilTheme(t *testing.T) {
	p := SplitViewPainter{Theme: nil}
	colors := p.resolveColors()

	if colors == (splitview.DividerColorScheme{}) {
		t.Error("should return non-zero default colors")
	}
}

func TestSplitViewPainter_ResolveColors_WithTheme(t *testing.T) {
	theme := New(widget.Hex(0xFF0000))
	p := SplitViewPainter{Theme: theme}
	colors := p.resolveColors()

	if colors == (splitview.DividerColorScheme{}) {
		t.Error("should return non-zero colors from theme")
	}
	if colors.Divider == m3DefaultSplitViewColors.Divider {
		t.Error("themed colors should differ from default purple")
	}
}

func TestM3ResolvedSplitDividerColor(t *testing.T) {
	colors := m3DefaultSplitViewColors

	normal := m3ResolvedSplitDividerColor(splitview.PaintState{
		DividerRect: geometry.NewRect(0, 0, 10, 100),
	}, colors)
	if normal != colors.Divider {
		t.Error("normal should return Divider color")
	}

	hovered := m3ResolvedSplitDividerColor(splitview.PaintState{
		DividerRect: geometry.NewRect(0, 0, 10, 100),
		Hovered:     true,
	}, colors)
	if hovered != colors.DividerHover {
		t.Error("hovered should return DividerHover color")
	}

	dragging := m3ResolvedSplitDividerColor(splitview.PaintState{
		DividerRect: geometry.NewRect(0, 0, 10, 100),
		Dragging:    true,
	}, colors)
	if dragging != colors.DividerDrag {
		t.Error("dragging should return DividerDrag color")
	}

	// Dragging takes precedence.
	both := m3ResolvedSplitDividerColor(splitview.PaintState{
		DividerRect: geometry.NewRect(0, 0, 10, 100),
		Hovered:     true,
		Dragging:    true,
	}, colors)
	if both != colors.DividerDrag {
		t.Error("dragging should take precedence over hover")
	}
}

// --- svMockCanvas records basic draw calls ---

type svMockCanvas struct {
	drawCount       int
	drawRectCount   int
	drawCircleCount int
}

func (c *svMockCanvas) Clear(_ widget.Color)                     {}
func (c *svMockCanvas) DrawRect(_ geometry.Rect, _ widget.Color) { c.drawCount++; c.drawRectCount++ }
func (c *svMockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) {
	c.drawCount++
}
func (c *svMockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32) {
	c.drawCount++
}
func (c *svMockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {
	c.drawCount++
}
func (c *svMockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color) {
	c.drawCount++
	c.drawCircleCount++
}
func (c *svMockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {
	c.drawCount++
}
func (c *svMockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32) { c.drawCount++ }
func (c *svMockCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
	c.drawCount++
}

func (c *svMockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}
func (c *svMockCanvas) DrawImage(_ image.Image, _ geometry.Point)    { c.drawCount++ }
func (c *svMockCanvas) PushClip(_ geometry.Rect)                     {}
func (c *svMockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *svMockCanvas) PopClip()                                     {}
func (c *svMockCanvas) PushTransform(_ geometry.Point)               {}
func (c *svMockCanvas) PopTransform()                                {}
func (c *svMockCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }
