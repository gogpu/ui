package material3

import (
	"image"
	"testing"

	"github.com/gogpu/ui/core/toolbar"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/widget"
)

func TestToolbarPainter_CompileTimeCheck(t *testing.T) {
	var _ toolbar.Painter = ToolbarPainter{}
}

func TestToolbarPainter_PaintToolbar_EmptyBounds(t *testing.T) {
	p := ToolbarPainter{}
	canvas := &toolbarMockCanvas{}

	p.PaintToolbar(canvas, toolbar.PaintToolbarState{})

	if canvas.drawCount > 0 {
		t.Error("should not draw with empty bounds")
	}
}

func TestToolbarPainter_PaintToolbar(t *testing.T) {
	p := ToolbarPainter{}
	canvas := &toolbarMockCanvas{}

	p.PaintToolbar(canvas, toolbar.PaintToolbarState{
		Bounds: geometry.NewRect(0, 0, 400, 40),
	})

	if canvas.rectCount == 0 {
		t.Error("should draw toolbar background")
	}
}

func TestToolbarPainter_PaintButtonItem_EmptyBounds(t *testing.T) {
	p := ToolbarPainter{}
	canvas := &toolbarMockCanvas{}

	p.PaintButtonItem(canvas, toolbar.PaintButtonState{})

	if canvas.drawCount > 0 {
		t.Error("should not draw with empty bounds")
	}
}

func TestToolbarPainter_PaintButtonItem_Hovered(t *testing.T) {
	p := ToolbarPainter{}
	canvas := &toolbarMockCanvas{}

	p.PaintButtonItem(canvas, toolbar.PaintButtonState{
		Bounds:  geometry.NewRect(0, 0, 40, 40),
		Hovered: true,
		Icon:    icon.Settings,
	})

	if canvas.roundRectCount == 0 {
		t.Error("hovered button should draw rounded rect background")
	}
}

func TestToolbarPainter_PaintButtonItem_Pressed(t *testing.T) {
	p := ToolbarPainter{}
	canvas := &toolbarMockCanvas{}

	p.PaintButtonItem(canvas, toolbar.PaintButtonState{
		Bounds:  geometry.NewRect(0, 0, 40, 40),
		Pressed: true,
		Icon:    icon.Settings,
	})

	if canvas.roundRectCount == 0 {
		t.Error("pressed button should draw rounded rect background")
	}
}

func TestToolbarPainter_PaintButtonItem_Disabled(t *testing.T) {
	p := ToolbarPainter{}
	canvas := &toolbarMockCanvas{}

	p.PaintButtonItem(canvas, toolbar.PaintButtonState{
		Bounds:   geometry.NewRect(0, 0, 40, 40),
		Disabled: true,
		Hovered:  true,
		Icon:     icon.Settings,
	})

	// Disabled items should not get hover background.
	if canvas.roundRectCount > 0 {
		t.Error("disabled hovered button should not draw hover background")
	}
}

func TestToolbarPainter_PaintButtonItem_WithLabel(t *testing.T) {
	p := ToolbarPainter{}
	canvas := &toolbarMockCanvas{}

	p.PaintButtonItem(canvas, toolbar.PaintButtonState{
		Bounds:    geometry.NewRect(0, 0, 120, 40),
		Label:     "Save",
		ShowLabel: true,
		Icon:      icon.Settings,
	})

	if canvas.textCount == 0 {
		t.Error("should draw label text when ShowLabel is true")
	}
}

func TestToolbarPainter_PaintButtonItem_Focused(t *testing.T) {
	p := ToolbarPainter{}
	canvas := &toolbarMockCanvas{}

	p.PaintButtonItem(canvas, toolbar.PaintButtonState{
		Bounds:  geometry.NewRect(0, 0, 40, 40),
		Focused: true,
		Icon:    icon.Settings,
	})

	if canvas.strokeRoundRectCount == 0 {
		t.Error("focused button should draw focus ring")
	}
}

func TestToolbarPainter_PaintSeparator_EmptyBounds(t *testing.T) {
	p := ToolbarPainter{}
	canvas := &toolbarMockCanvas{}

	p.PaintSeparator(canvas, geometry.Rect{})

	if canvas.drawCount > 0 {
		t.Error("should not draw with empty bounds")
	}
}

func TestToolbarPainter_PaintSeparator(t *testing.T) {
	p := ToolbarPainter{}
	canvas := &toolbarMockCanvas{}

	p.PaintSeparator(canvas, geometry.NewRect(0, 0, 8, 40))

	if canvas.lineCount == 0 {
		t.Error("should draw separator line")
	}
}

func TestToolbarPainter_ResolveColors_NilTheme(t *testing.T) {
	p := ToolbarPainter{Theme: nil}
	colors := p.resolveColors()

	if colors != m3DefaultToolbarColors {
		t.Error("nil theme should return default M3 toolbar colors")
	}
}

func TestToolbarPainter_ResolveColors_WithTheme(t *testing.T) {
	theme := New(widget.Hex(0xFF0000))
	p := ToolbarPainter{Theme: theme}
	colors := p.resolveColors()

	if colors.Background != theme.Colors.SurfaceContainer {
		t.Errorf("Background = %v, want %v", colors.Background, theme.Colors.SurfaceContainer)
	}
	if colors.IconColor != theme.Colors.OnSurface {
		t.Errorf("IconColor = %v, want %v", colors.IconColor, theme.Colors.OnSurface)
	}
}

// --- toolbarMockCanvas ---

type toolbarMockCanvas struct {
	drawCount            int
	rectCount            int
	roundRectCount       int
	strokeRoundRectCount int
	lineCount            int
	textCount            int
}

func (c *toolbarMockCanvas) Clear(_ widget.Color) {}
func (c *toolbarMockCanvas) DrawRect(_ geometry.Rect, _ widget.Color) {
	c.drawCount++
	c.rectCount++
}
func (c *toolbarMockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) { c.drawCount++ }
func (c *toolbarMockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32) {
	c.drawCount++
	c.roundRectCount++
}
func (c *toolbarMockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {
	c.drawCount++
	c.strokeRoundRectCount++
}
func (c *toolbarMockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color) { c.drawCount++ }
func (c *toolbarMockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {
	c.drawCount++
}
func (c *toolbarMockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32) {
	c.drawCount++
	c.lineCount++
}
func (c *toolbarMockCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
	c.drawCount++
	c.textCount++
}

func (c *toolbarMockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}
func (c *toolbarMockCanvas) DrawImage(_ image.Image, _ geometry.Point)    { c.drawCount++ }
func (c *toolbarMockCanvas) PushClip(_ geometry.Rect)                     {}
func (c *toolbarMockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *toolbarMockCanvas) PopClip()                                     {}
func (c *toolbarMockCanvas) PushTransform(_ geometry.Point)               {}
func (c *toolbarMockCanvas) PopTransform()                                {}
func (c *toolbarMockCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }
