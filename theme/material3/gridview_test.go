package material3

import (
	"image"
	"testing"

	"github.com/gogpu/ui/core/gridview"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

func TestGridViewPainter_CompileTimeCheck(t *testing.T) {
	var _ gridview.Painter = GridViewPainter{}
}

func TestGridViewPainter_CellBackground_EmptyBounds(t *testing.T) {
	p := GridViewPainter{}
	canvas := &gvMockCanvas{}

	p.PaintCellBackground(canvas, gridview.CellPaintState{})

	if canvas.drawCount > 0 {
		t.Error("should not draw anything with empty bounds")
	}
}

func TestGridViewPainter_CellBackground_Normal(t *testing.T) {
	p := GridViewPainter{}
	canvas := &gvMockCanvas{}

	cps := gridview.CellPaintState{
		Bounds: geometry.NewRect(0, 0, 100, 100),
		Index:  0,
	}

	p.PaintCellBackground(canvas, cps)

	// Normal (not hovered) should not draw.
	if canvas.drawCount > 0 {
		t.Error("normal cell should not draw background")
	}
}

func TestGridViewPainter_CellBackground_Hovered(t *testing.T) {
	p := GridViewPainter{}
	canvas := &gvMockCanvas{}

	cps := gridview.CellPaintState{
		Bounds:  geometry.NewRect(0, 0, 100, 100),
		Index:   0,
		Hovered: true,
	}

	p.PaintCellBackground(canvas, cps)

	if canvas.drawRoundRectCount == 0 {
		t.Error("hovered cell should draw background (DrawRoundRect)")
	}
}

func TestGridViewPainter_CellBackground_HoveredDisabled(t *testing.T) {
	p := GridViewPainter{}
	canvas := &gvMockCanvas{}

	cps := gridview.CellPaintState{
		Bounds:   geometry.NewRect(0, 0, 100, 100),
		Index:    0,
		Hovered:  true,
		Disabled: true,
	}

	p.PaintCellBackground(canvas, cps)

	if canvas.drawCount > 0 {
		t.Error("hovered+disabled should not draw hover background")
	}
}

func TestGridViewPainter_Selection_EmptyBounds(t *testing.T) {
	p := GridViewPainter{}
	canvas := &gvMockCanvas{}

	p.PaintSelection(canvas, gridview.CellPaintState{Selected: true})

	if canvas.drawCount > 0 {
		t.Error("should not draw anything with empty bounds")
	}
}

func TestGridViewPainter_Selection_NotSelected(t *testing.T) {
	p := GridViewPainter{}
	canvas := &gvMockCanvas{}

	cps := gridview.CellPaintState{
		Bounds:   geometry.NewRect(0, 0, 100, 100),
		Index:    0,
		Selected: false,
	}

	p.PaintSelection(canvas, cps)

	if canvas.drawCount > 0 {
		t.Error("unselected cell should not draw selection")
	}
}

func TestGridViewPainter_Selection_Selected(t *testing.T) {
	p := GridViewPainter{}
	canvas := &gvMockCanvas{}

	cps := gridview.CellPaintState{
		Bounds:   geometry.NewRect(0, 0, 100, 100),
		Index:    0,
		Selected: true,
	}

	p.PaintSelection(canvas, cps)

	if canvas.drawRoundRectCount == 0 {
		t.Error("selected cell should draw selection (DrawRoundRect)")
	}
}

func TestGridViewPainter_Selection_SelectedFocused(t *testing.T) {
	p := GridViewPainter{}
	canvas := &gvMockCanvas{}

	cps := gridview.CellPaintState{
		Bounds:   geometry.NewRect(0, 0, 100, 100),
		Index:    0,
		Selected: true,
		Focused:  true,
	}

	p.PaintSelection(canvas, cps)

	if canvas.strokeRoundRectCount == 0 {
		t.Error("selected+focused should draw focus border (StrokeRoundRect)")
	}
}

func TestGridViewPainter_Selection_SelectedFocusedDisabled(t *testing.T) {
	p := GridViewPainter{}
	canvas := &gvMockCanvas{}

	cps := gridview.CellPaintState{
		Bounds:   geometry.NewRect(0, 0, 100, 100),
		Index:    0,
		Selected: true,
		Focused:  true,
		Disabled: true,
	}

	p.PaintSelection(canvas, cps)

	if canvas.strokeRoundRectCount > 0 {
		t.Error("selected+focused+disabled should not draw focus border")
	}
}

func TestGridViewPainter_EmptyState_EmptyBounds(t *testing.T) {
	p := GridViewPainter{}
	canvas := &gvMockCanvas{}

	p.PaintEmptyState(canvas, geometry.Rect{})

	if canvas.drawCount > 0 {
		t.Error("should not draw anything with empty bounds")
	}
}

func TestGridViewPainter_EmptyState_NilTheme(t *testing.T) {
	p := GridViewPainter{Theme: nil}
	canvas := &gvMockCanvas{}

	p.PaintEmptyState(canvas, geometry.NewRect(0, 0, 300, 200))

	if canvas.drawTextCount == 0 {
		t.Error("should draw empty state text")
	}
}

func TestGridViewPainter_EmptyState_WithTheme(t *testing.T) {
	theme := New(widget.Hex(0x6750A4))
	p := GridViewPainter{Theme: theme}
	canvas := &gvMockCanvas{}

	p.PaintEmptyState(canvas, geometry.NewRect(0, 0, 300, 200))

	if canvas.drawTextCount == 0 {
		t.Error("should draw empty state text with theme")
	}
}

func TestGridViewPainter_WithColorScheme(t *testing.T) {
	p := GridViewPainter{}
	canvas := &gvMockCanvas{}

	scheme := gridview.GridColorScheme{
		SelectionColor: widget.ColorRed,
		HoverColor:     widget.ColorBlue,
		FocusColor:     widget.ColorGreen,
		EmptyTextColor: widget.ColorGray,
		CellBackground: widget.ColorWhite,
	}

	cps := gridview.CellPaintState{
		Bounds:      geometry.NewRect(0, 0, 100, 100),
		Index:       0,
		Selected:    true,
		Hovered:     true,
		ColorScheme: scheme,
	}

	p.PaintCellBackground(canvas, cps)
	p.PaintSelection(canvas, cps)

	if canvas.drawCount == 0 {
		t.Error("should draw with custom color scheme")
	}
}

func TestGridViewPainter_ResolveColors_NilTheme(t *testing.T) {
	p := GridViewPainter{Theme: nil}
	colors := p.resolveColors()

	if colors == (gridview.GridColorScheme{}) {
		t.Error("should return non-zero default colors")
	}
}

func TestGridViewPainter_ResolveColors_WithTheme(t *testing.T) {
	theme := New(widget.Hex(0xFF0000))
	p := GridViewPainter{Theme: theme}
	colors := p.resolveColors()

	if colors == (gridview.GridColorScheme{}) {
		t.Error("should return non-zero colors from theme")
	}
	if colors.SelectionColor == m3DefaultGridColors.SelectionColor {
		t.Error("themed colors should differ from default purple")
	}
}

// --- gvMockCanvas records basic draw calls ---

type gvMockCanvas struct {
	drawCount            int
	drawRoundRectCount   int
	strokeRoundRectCount int
	drawTextCount        int
}

func (c *gvMockCanvas) Clear(_ widget.Color)                                  {}
func (c *gvMockCanvas) DrawRect(_ geometry.Rect, _ widget.Color)              { c.drawCount++ }
func (c *gvMockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) { c.drawCount++ }
func (c *gvMockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32) {
	c.drawCount++
	c.drawRoundRectCount++
}
func (c *gvMockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {
	c.drawCount++
	c.strokeRoundRectCount++
}
func (c *gvMockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color) { c.drawCount++ }
func (c *gvMockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {
	c.drawCount++
}
func (c *gvMockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32) { c.drawCount++ }
func (c *gvMockCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
	c.drawCount++
	c.drawTextCount++
}

func (c *gvMockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}
func (c *gvMockCanvas) DrawImage(_ image.Image, _ geometry.Point)    { c.drawCount++ }
func (c *gvMockCanvas) PushClip(_ geometry.Rect)                     {}
func (c *gvMockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *gvMockCanvas) PopClip()                                     {}
func (c *gvMockCanvas) PushTransform(_ geometry.Point)               {}
func (c *gvMockCanvas) PopTransform()                                {}
func (c *gvMockCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }
