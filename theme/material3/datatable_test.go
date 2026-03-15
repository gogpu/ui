package material3

import (
	"image"
	"testing"

	"github.com/gogpu/ui/core/datatable"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

func TestDataTablePainter_CompileTimeCheck(t *testing.T) {
	var _ datatable.Painter = DataTablePainter{}
}

func TestDataTablePainter_PaintHeader_EmptyBounds(t *testing.T) {
	p := DataTablePainter{}
	canvas := &tableMockCanvas{}

	p.PaintHeader(canvas, geometry.Rect{}, datatable.HeaderPaintState{})

	if canvas.drawCount > 0 {
		t.Error("should not draw with empty bounds")
	}
}

func TestDataTablePainter_PaintHeader(t *testing.T) {
	p := DataTablePainter{}
	canvas := &tableMockCanvas{}

	p.PaintHeader(canvas, geometry.NewRect(0, 0, 400, 40), datatable.HeaderPaintState{})

	if canvas.rectCount < 2 {
		t.Errorf("should draw header bg + divider (at least 2 rects), got %d", canvas.rectCount)
	}
}

func TestDataTablePainter_PaintHeaderCell(t *testing.T) {
	p := DataTablePainter{}
	canvas := &tableMockCanvas{}

	p.PaintHeaderCell(canvas, geometry.NewRect(0, 0, 100, 40), datatable.HeaderCellPaintState{
		Title: "Name",
		Align: widget.TextAlignLeft,
	})

	if canvas.textCount == 0 {
		t.Error("should draw header cell text")
	}
}

func TestDataTablePainter_PaintHeaderCell_Hovered(t *testing.T) {
	p := DataTablePainter{}
	canvas := &tableMockCanvas{}

	p.PaintHeaderCell(canvas, geometry.NewRect(0, 0, 100, 40), datatable.HeaderCellPaintState{
		Title:    "Name",
		Sortable: true,
		Hovered:  true,
		Align:    widget.TextAlignLeft,
	})

	if canvas.rectCount == 0 {
		t.Error("hovered sortable header should draw hover highlight")
	}
}

func TestDataTablePainter_PaintRow_Zebra(t *testing.T) {
	p := DataTablePainter{}
	canvas := &tableMockCanvas{}

	p.PaintRow(canvas, datatable.RowPaintState{
		Bounds:   geometry.NewRect(0, 0, 400, 30),
		RowIndex: 1, // odd = alternate
	})

	if canvas.rectCount == 0 {
		t.Error("odd row should draw zebra background")
	}
}

func TestDataTablePainter_PaintRow_Selected(t *testing.T) {
	p := DataTablePainter{}
	canvas := &tableMockCanvas{}

	p.PaintRow(canvas, datatable.RowPaintState{
		Bounds:   geometry.NewRect(0, 0, 400, 30),
		RowIndex: 0,
		Selected: true,
	})

	if canvas.rectCount == 0 {
		t.Error("selected row should draw selection highlight")
	}
}

func TestDataTablePainter_PaintRow_Focused(t *testing.T) {
	p := DataTablePainter{}
	canvas := &tableMockCanvas{}

	p.PaintRow(canvas, datatable.RowPaintState{
		Bounds:   geometry.NewRect(0, 0, 400, 30),
		RowIndex: 0,
		Focused:  true,
	})

	if canvas.strokeRectCount == 0 {
		t.Error("focused row should draw focus ring")
	}
}

func TestDataTablePainter_PaintCell(t *testing.T) {
	p := DataTablePainter{}
	canvas := &tableMockCanvas{}

	p.PaintCell(canvas, datatable.CellPaintState{
		Bounds: geometry.NewRect(0, 0, 100, 30),
		Value:  "Hello",
		Align:  widget.TextAlignLeft,
	})

	if canvas.textCount == 0 {
		t.Error("should draw cell text")
	}
}

func TestDataTablePainter_PaintEmptyState(t *testing.T) {
	p := DataTablePainter{}
	canvas := &tableMockCanvas{}

	p.PaintEmptyState(canvas, geometry.NewRect(0, 0, 400, 200))

	if canvas.textCount == 0 {
		t.Error("should draw empty state text")
	}
}

func TestDataTablePainter_ResolveColors_NilTheme(t *testing.T) {
	p := DataTablePainter{Theme: nil}
	colors := p.resolveColors()

	if colors != m3DefaultTableColors {
		t.Error("nil theme should return default M3 table colors")
	}
}

func TestDataTablePainter_ResolveColors_WithTheme(t *testing.T) {
	theme := New(widget.Hex(0xFF0000))
	p := DataTablePainter{Theme: theme}
	colors := p.resolveColors()

	if colors.HeaderBackground != theme.Colors.SurfaceContainerHighest {
		t.Errorf("HeaderBackground = %v, want %v", colors.HeaderBackground, theme.Colors.SurfaceContainerHighest)
	}
	if colors.CellText != theme.Colors.OnSurface {
		t.Errorf("CellText = %v, want %v", colors.CellText, theme.Colors.OnSurface)
	}
}

func TestDataTablePainter_WithTheme(t *testing.T) {
	theme := New(widget.Hex(0x6750A4))
	p := DataTablePainter{Theme: theme}
	canvas := &tableMockCanvas{}

	p.PaintHeader(canvas, geometry.NewRect(0, 0, 400, 40), datatable.HeaderPaintState{})

	if canvas.drawCount == 0 {
		t.Error("should draw with theme")
	}
}

// --- tableMockCanvas ---

type tableMockCanvas struct {
	drawCount       int
	rectCount       int
	strokeRectCount int
	textCount       int
}

func (c *tableMockCanvas) Clear(_ widget.Color) {}
func (c *tableMockCanvas) DrawRect(_ geometry.Rect, _ widget.Color) {
	c.drawCount++
	c.rectCount++
}
func (c *tableMockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) {
	c.drawCount++
	c.strokeRectCount++
}
func (c *tableMockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32) { c.drawCount++ }
func (c *tableMockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {
	c.drawCount++
}
func (c *tableMockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color) { c.drawCount++ }
func (c *tableMockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {
	c.drawCount++
}
func (c *tableMockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32) { c.drawCount++ }
func (c *tableMockCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
	c.drawCount++
	c.textCount++
}

func (c *tableMockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}
func (c *tableMockCanvas) DrawImage(_ image.Image, _ geometry.Point)    { c.drawCount++ }
func (c *tableMockCanvas) PushClip(_ geometry.Rect)                     {}
func (c *tableMockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *tableMockCanvas) PopClip()                                     {}
func (c *tableMockCanvas) PushTransform(_ geometry.Point)               {}
func (c *tableMockCanvas) PopTransform()                                {}
func (c *tableMockCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }
