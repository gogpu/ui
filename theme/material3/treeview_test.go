package material3

import (
	"image"
	"testing"

	"github.com/gogpu/ui/core/treeview"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

func TestTreeViewPainter_CompileTimeCheck(t *testing.T) {
	var _ treeview.Painter = TreeViewPainter{}
}

func TestTreeViewPainter_PaintRowBackground_EmptyBounds(t *testing.T) {
	p := TreeViewPainter{}
	canvas := &treeMockCanvas{}

	p.PaintRowBackground(canvas, treeview.RowPaintState{})

	if canvas.drawCount > 0 {
		t.Error("should not draw with empty bounds")
	}
}

func TestTreeViewPainter_PaintRowBackground_Hovered(t *testing.T) {
	p := TreeViewPainter{}
	canvas := &treeMockCanvas{}

	p.PaintRowBackground(canvas, treeview.RowPaintState{
		Bounds:  geometry.NewRect(0, 0, 200, 30),
		Hovered: true,
	})

	if canvas.roundRectCount == 0 {
		t.Error("hovered row should draw rounded rect background")
	}
}

func TestTreeViewPainter_PaintSelection_Selected(t *testing.T) {
	p := TreeViewPainter{}
	canvas := &treeMockCanvas{}

	p.PaintSelection(canvas, treeview.RowPaintState{
		Bounds:   geometry.NewRect(0, 0, 200, 30),
		Selected: true,
	})

	if canvas.roundRectCount == 0 {
		t.Error("selected row should draw selection highlight")
	}
}

func TestTreeViewPainter_PaintSelection_Focused(t *testing.T) {
	p := TreeViewPainter{}
	canvas := &treeMockCanvas{}

	p.PaintSelection(canvas, treeview.RowPaintState{
		Bounds:   geometry.NewRect(0, 0, 200, 30),
		Selected: true,
		Focused:  true,
	})

	if canvas.strokeRoundRectCount == 0 {
		t.Error("focused selected row should draw focus ring")
	}
}

func TestTreeViewPainter_PaintSelection_FocusedDisabled(t *testing.T) {
	p := TreeViewPainter{}
	canvas := &treeMockCanvas{}

	p.PaintSelection(canvas, treeview.RowPaintState{
		Bounds:   geometry.NewRect(0, 0, 200, 30),
		Selected: true,
		Focused:  true,
		Disabled: true,
	})

	if canvas.strokeRoundRectCount > 0 {
		t.Error("disabled focused row should not draw focus ring")
	}
}

func TestTreeViewPainter_PaintExpandIcon_Expanded(t *testing.T) {
	p := TreeViewPainter{}
	canvas := &treeMockCanvas{}

	p.PaintExpandIcon(canvas, treeview.ExpandIconState{
		Bounds:   geometry.NewRect(0, 0, 16, 16),
		Expanded: true,
	})

	if canvas.lineCount < 2 {
		t.Errorf("expanded icon should draw 2 lines (chevron), got %d", canvas.lineCount)
	}
}

func TestTreeViewPainter_PaintExpandIcon_Collapsed(t *testing.T) {
	p := TreeViewPainter{}
	canvas := &treeMockCanvas{}

	p.PaintExpandIcon(canvas, treeview.ExpandIconState{
		Bounds: geometry.NewRect(0, 0, 16, 16),
	})

	if canvas.lineCount < 2 {
		t.Errorf("collapsed icon should draw 2 lines (chevron), got %d", canvas.lineCount)
	}
}

func TestTreeViewPainter_PaintConnectorLines(t *testing.T) {
	p := TreeViewPainter{}
	canvas := &treeMockCanvas{}

	p.PaintConnectorLines(canvas, treeview.ConnectorState{
		RowBounds:     geometry.NewRect(0, 0, 200, 30),
		Depth:         2,
		IndentWidth:   24,
		IsLastChild:   false,
		ParentHasMore: []bool{true},
	})

	if canvas.lineCount < 2 {
		t.Errorf("connector lines should draw at least 2 lines, got %d", canvas.lineCount)
	}
}

func TestTreeViewPainter_PaintConnectorLines_DepthZero(t *testing.T) {
	p := TreeViewPainter{}
	canvas := &treeMockCanvas{}

	p.PaintConnectorLines(canvas, treeview.ConnectorState{
		RowBounds: geometry.NewRect(0, 0, 200, 30),
		Depth:     0,
	})

	if canvas.lineCount > 0 {
		t.Error("depth 0 should not draw connector lines")
	}
}

func TestTreeViewPainter_PaintLabel(t *testing.T) {
	p := TreeViewPainter{}
	canvas := &treeMockCanvas{}

	p.PaintLabel(canvas, treeview.LabelState{
		Bounds: geometry.NewRect(0, 0, 200, 30),
		Text:   "Node",
	})

	if canvas.textCount == 0 {
		t.Error("should draw label text")
	}
}

func TestTreeViewPainter_PaintLabel_Empty(t *testing.T) {
	p := TreeViewPainter{}
	canvas := &treeMockCanvas{}

	p.PaintLabel(canvas, treeview.LabelState{
		Bounds: geometry.NewRect(0, 0, 200, 30),
		Text:   "",
	})

	if canvas.textCount > 0 {
		t.Error("should not draw empty label")
	}
}

func TestTreeViewPainter_PaintEmptyState(t *testing.T) {
	p := TreeViewPainter{}
	canvas := &treeMockCanvas{}

	p.PaintEmptyState(canvas, geometry.NewRect(0, 0, 200, 100))

	if canvas.textCount == 0 {
		t.Error("should draw empty state text")
	}
}

func TestTreeViewPainter_ResolveColors_NilTheme(t *testing.T) {
	p := TreeViewPainter{Theme: nil}
	colors := p.resolveColors()

	if colors != m3DefaultTreeColors {
		t.Error("nil theme should return default M3 tree colors")
	}
}

func TestTreeViewPainter_ResolveColors_WithTheme(t *testing.T) {
	theme := New(widget.Hex(0xFF0000))
	p := TreeViewPainter{Theme: theme}
	colors := p.resolveColors()

	if colors.LabelColor != theme.Colors.OnSurface {
		t.Errorf("LabelColor = %v, want %v", colors.LabelColor, theme.Colors.OnSurface)
	}
}

func TestTreeViewPainter_WithTheme(t *testing.T) {
	theme := New(widget.Hex(0x6750A4))
	p := TreeViewPainter{Theme: theme}
	canvas := &treeMockCanvas{}

	p.PaintRowBackground(canvas, treeview.RowPaintState{
		Bounds:  geometry.NewRect(0, 0, 200, 30),
		Hovered: true,
	})

	if canvas.drawCount == 0 {
		t.Error("should draw with theme")
	}
}

// --- treeMockCanvas ---

type treeMockCanvas struct {
	drawCount            int
	roundRectCount       int
	strokeRoundRectCount int
	lineCount            int
	textCount            int
}

func (c *treeMockCanvas) Clear(_ widget.Color)                                  {}
func (c *treeMockCanvas) DrawRect(_ geometry.Rect, _ widget.Color)              { c.drawCount++ }
func (c *treeMockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) { c.drawCount++ }
func (c *treeMockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32) {
	c.drawCount++
	c.roundRectCount++
}
func (c *treeMockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {
	c.drawCount++
	c.strokeRoundRectCount++
}
func (c *treeMockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color) { c.drawCount++ }
func (c *treeMockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {
	c.drawCount++
}
func (c *treeMockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32) {
	c.drawCount++
	c.lineCount++
}
func (c *treeMockCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
	c.drawCount++
	c.textCount++
}

func (c *treeMockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}
func (c *treeMockCanvas) DrawImage(_ image.Image, _ geometry.Point)    { c.drawCount++ }
func (c *treeMockCanvas) PushClip(_ geometry.Rect)                     {}
func (c *treeMockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *treeMockCanvas) PopClip()                                     {}
func (c *treeMockCanvas) PushTransform(_ geometry.Point)               {}
func (c *treeMockCanvas) PopTransform()                                {}
func (c *treeMockCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }
