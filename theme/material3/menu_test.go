package material3

import (
	"image"
	"testing"

	"github.com/gogpu/ui/core/menu"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

func TestMenuPainter_CompileTimeCheck(t *testing.T) {
	var _ menu.Painter = MenuPainter{}
}

func TestMenuPainter_PaintMenuBar_EmptyBounds(t *testing.T) {
	p := MenuPainter{}
	canvas := &menuMockCanvas{}

	p.PaintMenuBar(canvas, &menu.MenuBarPaintState{})

	if canvas.drawCount > 0 {
		t.Error("should not draw with empty bounds")
	}
}

func TestMenuPainter_PaintMenuBar(t *testing.T) {
	p := MenuPainter{}
	canvas := &menuMockCanvas{}

	p.PaintMenuBar(canvas, &menu.MenuBarPaintState{
		Bounds: geometry.NewRect(0, 0, 400, 32),
		Menus: []menu.TopMenu{
			{Label: "File"},
			{Label: "Edit"},
		},
		MenuRects: []geometry.Rect{
			geometry.NewRect(0, 0, 50, 32),
			geometry.NewRect(50, 0, 50, 32),
		},
		OpenIndex:    -1,
		HoveredIndex: -1,
	})

	if canvas.rectCount == 0 {
		t.Error("should draw menu bar background")
	}
	if canvas.textCount < 2 {
		t.Errorf("should draw 2 menu labels, got %d", canvas.textCount)
	}
}

func TestMenuPainter_PaintMenuBar_Hovered(t *testing.T) {
	p := MenuPainter{}
	canvas := &menuMockCanvas{}

	p.PaintMenuBar(canvas, &menu.MenuBarPaintState{
		Bounds:       geometry.NewRect(0, 0, 400, 32),
		Menus:        []menu.TopMenu{{Label: "File"}},
		MenuRects:    []geometry.Rect{geometry.NewRect(0, 0, 50, 32)},
		OpenIndex:    -1,
		HoveredIndex: 0,
	})

	if canvas.roundRectCount == 0 {
		t.Error("hovered menu label should draw rounded hover background")
	}
}

func TestMenuPainter_PaintMenuBar_Open(t *testing.T) {
	p := MenuPainter{}
	canvas := &menuMockCanvas{}

	p.PaintMenuBar(canvas, &menu.MenuBarPaintState{
		Bounds:       geometry.NewRect(0, 0, 400, 32),
		Menus:        []menu.TopMenu{{Label: "File"}},
		MenuRects:    []geometry.Rect{geometry.NewRect(0, 0, 50, 32)},
		OpenIndex:    0,
		HoveredIndex: -1,
	})

	if canvas.roundRectCount == 0 {
		t.Error("open menu label should draw active background")
	}
}

func TestMenuPainter_PaintMenu_EmptyBounds(t *testing.T) {
	p := MenuPainter{}
	canvas := &menuMockCanvas{}

	p.PaintMenu(canvas, &menu.MenuPaintState{})

	if canvas.drawCount > 0 {
		t.Error("should not draw with empty bounds")
	}
}

func TestMenuPainter_PaintMenu(t *testing.T) {
	p := MenuPainter{}
	canvas := &menuMockCanvas{}

	p.PaintMenu(canvas, &menu.MenuPaintState{
		Bounds: geometry.NewRect(0, 32, 200, 100),
		Items: []menu.MenuItem{
			{Label: "New"},
			{Label: "Open", Shortcut: "Ctrl+O"},
		},
		HighlightedIndex: -1,
		SubMenuOpenIndex: -1,
		ItemHeight:       32,
		SeparatorHeight:  8,
	})

	// Should draw shadow + surface + border + items.
	if canvas.roundRectCount < 2 {
		t.Errorf("should draw shadow + surface (at least 2 round rects), got %d", canvas.roundRectCount)
	}
	if canvas.textCount < 2 {
		t.Errorf("should draw 2 item labels, got %d", canvas.textCount)
	}
}

func TestMenuPainter_PaintMenu_Highlighted(t *testing.T) {
	p := MenuPainter{}
	canvas := &menuMockCanvas{}

	p.PaintMenu(canvas, &menu.MenuPaintState{
		Bounds: geometry.NewRect(0, 32, 200, 100),
		Items: []menu.MenuItem{
			{Label: "New"},
		},
		HighlightedIndex: 0,
		SubMenuOpenIndex: -1,
		ItemHeight:       32,
		SeparatorHeight:  8,
	})

	// Highlighted item draws extra rounded rect.
	if canvas.roundRectCount < 3 {
		t.Errorf("highlighted item should draw highlight rect, got %d round rects", canvas.roundRectCount)
	}
}

func TestMenuPainter_PaintMenu_Separator(t *testing.T) {
	p := MenuPainter{}
	canvas := &menuMockCanvas{}

	p.PaintMenu(canvas, &menu.MenuPaintState{
		Bounds: geometry.NewRect(0, 32, 200, 100),
		Items: []menu.MenuItem{
			{Label: "New"},
			menu.Sep(),
			{Label: "Quit"},
		},
		HighlightedIndex: -1,
		SubMenuOpenIndex: -1,
		ItemHeight:       32,
		SeparatorHeight:  8,
	})

	if canvas.lineCount == 0 {
		t.Error("separator should draw a line")
	}
}

func TestMenuPainter_ResolveColors_NilTheme(t *testing.T) {
	p := MenuPainter{Theme: nil}
	colors := p.resolveColors()

	if colors != m3DefaultMenuColors {
		t.Error("nil theme should return default M3 menu colors")
	}
}

func TestMenuPainter_ResolveColors_WithTheme(t *testing.T) {
	theme := New(widget.Hex(0xFF0000))
	p := MenuPainter{Theme: theme}
	colors := p.resolveColors()

	if colors.BarBackground != theme.Colors.SurfaceContainer {
		t.Errorf("BarBackground = %v, want %v", colors.BarBackground, theme.Colors.SurfaceContainer)
	}
	if colors.BarActiveText != theme.Colors.Primary {
		t.Errorf("BarActiveText = %v, want %v", colors.BarActiveText, theme.Colors.Primary)
	}
}

// --- menuMockCanvas ---

type menuMockCanvas struct {
	drawCount            int
	rectCount            int
	roundRectCount       int
	strokeRoundRectCount int
	lineCount            int
	textCount            int
}

func (c *menuMockCanvas) Clear(_ widget.Color) {}
func (c *menuMockCanvas) DrawRect(_ geometry.Rect, _ widget.Color) {
	c.drawCount++
	c.rectCount++
}
func (c *menuMockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) { c.drawCount++ }
func (c *menuMockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32) {
	c.drawCount++
	c.roundRectCount++
}
func (c *menuMockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {
	c.drawCount++
	c.strokeRoundRectCount++
}
func (c *menuMockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color) { c.drawCount++ }
func (c *menuMockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {
	c.drawCount++
}
func (c *menuMockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32) {
	c.drawCount++
	c.lineCount++
}
func (c *menuMockCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
	c.drawCount++
	c.textCount++
}

func (c *menuMockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}
func (c *menuMockCanvas) DrawImage(_ image.Image, _ geometry.Point)    { c.drawCount++ }
func (c *menuMockCanvas) PushClip(_ geometry.Rect)                     {}
func (c *menuMockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *menuMockCanvas) PopClip()                                     {}
func (c *menuMockCanvas) PushTransform(_ geometry.Point)               {}
func (c *menuMockCanvas) PopTransform()                                {}
func (c *menuMockCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }
