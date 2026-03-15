package material3

import (
	"image"
	"testing"

	"github.com/gogpu/ui/core/docking"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

func TestDockingPainter_CompileTimeCheck(t *testing.T) {
	var _ docking.Painter = DockingPainter{}
}

func TestDockingPainter_PaintZoneTabs_EmptyBounds(t *testing.T) {
	p := DockingPainter{}
	canvas := &dockingMockCanvas{}

	p.PaintZoneTabs(canvas, docking.ZoneTabsPaintState{})

	if canvas.drawCount > 0 {
		t.Error("should not draw with empty bounds")
	}
}

func TestDockingPainter_PaintZoneTabs(t *testing.T) {
	p := DockingPainter{}
	canvas := &dockingMockCanvas{}

	p.PaintZoneTabs(canvas, docking.ZoneTabsPaintState{
		Zone:         docking.Left,
		TabBarBounds: geometry.NewRect(0, 0, 200, 32),
		Tabs: []docking.ZoneTabState{
			{Title: "Panel 1", Bounds: geometry.NewRect(0, 0, 100, 32), Active: true},
			{Title: "Panel 2", Bounds: geometry.NewRect(100, 0, 100, 32)},
		},
		ActiveIdx: 0,
	})

	if canvas.rectCount == 0 {
		t.Error("should draw tab bar background")
	}
	if canvas.textCount < 2 {
		t.Errorf("should draw 2 tab labels, got %d", canvas.textCount)
	}
}

func TestDockingPainter_PaintZoneTabs_ActiveIndicator(t *testing.T) {
	p := DockingPainter{}
	canvas := &dockingMockCanvas{}

	p.PaintZoneTabs(canvas, docking.ZoneTabsPaintState{
		TabBarBounds: geometry.NewRect(0, 0, 200, 32),
		Tabs: []docking.ZoneTabState{
			{Title: "Panel", Bounds: geometry.NewRect(0, 0, 100, 32), Active: true},
		},
		ActiveIdx: 0,
	})

	// Should draw bg + indicator (at least 2 rects).
	if canvas.rectCount < 2 {
		t.Errorf("should draw bg + active indicator, got %d rects", canvas.rectCount)
	}
}

func TestDockingPainter_PaintZoneTabs_Hovered(t *testing.T) {
	p := DockingPainter{}
	canvas := &dockingMockCanvas{}

	p.PaintZoneTabs(canvas, docking.ZoneTabsPaintState{
		TabBarBounds: geometry.NewRect(0, 0, 200, 32),
		Tabs: []docking.ZoneTabState{
			{Title: "Panel", Bounds: geometry.NewRect(0, 0, 100, 32), Hovered: true},
		},
		ActiveIdx: -1,
	})

	// Hovered tab draws an additional rect.
	if canvas.rectCount < 2 {
		t.Errorf("hovered tab should draw hover background, got %d rects", canvas.rectCount)
	}
}

func TestDockingPainter_PaintZoneTabs_Closeable(t *testing.T) {
	p := DockingPainter{}
	canvas := &dockingMockCanvas{}

	p.PaintZoneTabs(canvas, docking.ZoneTabsPaintState{
		TabBarBounds: geometry.NewRect(0, 0, 200, 32),
		Tabs: []docking.ZoneTabState{
			{
				Title:             "Panel",
				Bounds:            geometry.NewRect(0, 0, 100, 32),
				Active:            true,
				Closeable:         true,
				CloseButtonBounds: geometry.NewRect(80, 8, 14, 14),
			},
		},
		ActiveIdx: 0,
	})

	// Close button draws 2 lines (X mark).
	if canvas.lineCount < 2 {
		t.Errorf("closeable tab should draw X close button (2 lines), got %d", canvas.lineCount)
	}
}

func TestDockingPainter_PaintZoneBorder_EmptyBounds(t *testing.T) {
	p := DockingPainter{}
	canvas := &dockingMockCanvas{}

	p.PaintZoneBorder(canvas, geometry.Rect{}, docking.Left)

	if canvas.drawCount > 0 {
		t.Error("should not draw with empty bounds")
	}
}

func TestDockingPainter_PaintZoneBorder(t *testing.T) {
	p := DockingPainter{}
	canvas := &dockingMockCanvas{}

	p.PaintZoneBorder(canvas, geometry.NewRect(200, 0, 1, 400), docking.Left)

	if canvas.rectCount == 0 {
		t.Error("should draw zone border")
	}
}

func TestDockingPainter_ResolveColors_NilTheme(t *testing.T) {
	p := DockingPainter{Theme: nil}
	colors := p.resolveColors()

	if colors != m3DefaultDockingColors {
		t.Error("nil theme should return default M3 docking colors")
	}
}

func TestDockingPainter_ResolveColors_WithTheme(t *testing.T) {
	theme := New(widget.Hex(0xFF0000))
	p := DockingPainter{Theme: theme}
	colors := p.resolveColors()

	if colors.TabBarBackground != theme.Colors.SurfaceContainerHigh {
		t.Errorf("TabBarBackground = %v, want %v", colors.TabBarBackground, theme.Colors.SurfaceContainerHigh)
	}
	if colors.ActiveTabBackground != theme.Colors.Primary {
		t.Errorf("ActiveTabBackground = %v, want %v", colors.ActiveTabBackground, theme.Colors.Primary)
	}
}

func TestDockingPainter_WithTheme(t *testing.T) {
	theme := New(widget.Hex(0x6750A4))
	p := DockingPainter{Theme: theme}
	canvas := &dockingMockCanvas{}

	p.PaintZoneTabs(canvas, docking.ZoneTabsPaintState{
		TabBarBounds: geometry.NewRect(0, 0, 200, 32),
		Tabs: []docking.ZoneTabState{
			{Title: "Panel", Bounds: geometry.NewRect(0, 0, 100, 32)},
		},
		ActiveIdx: -1,
	})

	if canvas.drawCount == 0 {
		t.Error("should draw with theme")
	}
}

// --- dockingMockCanvas ---

type dockingMockCanvas struct {
	drawCount int
	rectCount int
	lineCount int
	textCount int
}

func (c *dockingMockCanvas) Clear(_ widget.Color) {}
func (c *dockingMockCanvas) DrawRect(_ geometry.Rect, _ widget.Color) {
	c.drawCount++
	c.rectCount++
}
func (c *dockingMockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32)    { c.drawCount++ }
func (c *dockingMockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32) { c.drawCount++ }
func (c *dockingMockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {
	c.drawCount++
}
func (c *dockingMockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color) { c.drawCount++ }
func (c *dockingMockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {
	c.drawCount++
}
func (c *dockingMockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32) {
	c.drawCount++
	c.lineCount++
}
func (c *dockingMockCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
	c.drawCount++
	c.textCount++
}

func (c *dockingMockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}
func (c *dockingMockCanvas) DrawImage(_ image.Image, _ geometry.Point)    { c.drawCount++ }
func (c *dockingMockCanvas) PushClip(_ geometry.Rect)                     {}
func (c *dockingMockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *dockingMockCanvas) PopClip()                                     {}
func (c *dockingMockCanvas) PushTransform(_ geometry.Point)               {}
func (c *dockingMockCanvas) PopTransform()                                {}
func (c *dockingMockCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }
