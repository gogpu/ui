package material3

import (
	"image"
	"testing"

	"github.com/gogpu/ui/core/tabview"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

func TestTabViewPainter_Interface(t *testing.T) {
	var _ tabview.Painter = TabViewPainter{}
}

func TestTabViewPainter_NilTheme(t *testing.T) {
	p := TabViewPainter{Theme: nil}
	canvas := &tabMockCanvas{}

	ps := tabview.PaintState{
		Bounds: geometry.NewRect(0, 0, 300, 48),
		Tabs: []tabview.TabState{
			{Label: "Tab1", Bounds: geometry.NewRect(0, 0, 150, 48), Selected: true},
			{Label: "Tab2", Bounds: geometry.NewRect(150, 0, 300, 48)},
		},
		SelectedIdx: 0,
		Position:    tabview.Top,
	}

	// Should not panic.
	p.PaintTabBar(canvas, ps)
}

func TestTabViewPainter_WithTheme(t *testing.T) {
	m3 := New(widget.Hex(0x6750A4))
	p := TabViewPainter{Theme: m3}
	canvas := &tabMockCanvas{}

	ps := tabview.PaintState{
		Bounds: geometry.NewRect(0, 0, 300, 48),
		Tabs: []tabview.TabState{
			{Label: "Tab1", Bounds: geometry.NewRect(0, 0, 150, 48), Selected: true},
			{Label: "Tab2", Bounds: geometry.NewRect(150, 0, 300, 48), Hovered: true},
		},
		SelectedIdx: 0,
		Position:    tabview.Top,
	}

	p.PaintTabBar(canvas, ps)
}

func TestTabViewPainter_DarkTheme(t *testing.T) {
	m3 := NewDark(widget.Hex(0x6750A4))
	p := TabViewPainter{Theme: m3}
	canvas := &tabMockCanvas{}

	ps := tabview.PaintState{
		Bounds: geometry.NewRect(0, 0, 300, 48),
		Tabs: []tabview.TabState{
			{Label: "Tab1", Bounds: geometry.NewRect(0, 0, 150, 48)},
			{Label: "Tab2", Bounds: geometry.NewRect(150, 0, 300, 48), Selected: true},
		},
		SelectedIdx: 1,
		Position:    tabview.Top,
	}

	p.PaintTabBar(canvas, ps)
}

func TestTabViewPainter_Focused(t *testing.T) {
	p := TabViewPainter{Theme: nil}
	canvas := &tabMockCanvas{}

	ps := tabview.PaintState{
		Bounds: geometry.NewRect(0, 0, 300, 48),
		Tabs: []tabview.TabState{
			{Label: "Tab1", Bounds: geometry.NewRect(0, 0, 300, 48), Selected: true},
		},
		SelectedIdx: 0,
		Position:    tabview.Top,
		Focused:     true,
	}

	p.PaintTabBar(canvas, ps)

	if canvas.strokeRoundRectCount == 0 {
		t.Error("should draw focus ring when focused")
	}
}

func TestTabViewPainter_DisabledTab(t *testing.T) {
	p := TabViewPainter{Theme: nil}
	canvas := &tabMockCanvas{}

	ps := tabview.PaintState{
		Bounds: geometry.NewRect(0, 0, 300, 48),
		Tabs: []tabview.TabState{
			{Label: "Tab1", Bounds: geometry.NewRect(0, 0, 150, 48), Disabled: true},
			{Label: "Tab2", Bounds: geometry.NewRect(150, 0, 300, 48), Selected: true},
		},
		SelectedIdx: 1,
		Position:    tabview.Top,
	}

	p.PaintTabBar(canvas, ps)
}

func TestTabViewPainter_CloseableTab(t *testing.T) {
	p := TabViewPainter{Theme: nil}
	canvas := &tabMockCanvas{}

	closeBounds := geometry.NewRect(120, 16, 136, 32)
	ps := tabview.PaintState{
		Bounds: geometry.NewRect(0, 0, 300, 48),
		Tabs: []tabview.TabState{
			{
				Label:             "Tab1",
				Bounds:            geometry.NewRect(0, 0, 150, 48),
				Selected:          true,
				Closeable:         true,
				CloseButtonBounds: closeBounds,
			},
		},
		SelectedIdx: 0,
		Position:    tabview.Top,
	}

	p.PaintTabBar(canvas, ps)

	if canvas.drawLineCount < 2 {
		t.Errorf("drawLineCount = %d, want >= 2 (X icon)", canvas.drawLineCount)
	}
}

func TestTabViewPainter_BottomPosition(t *testing.T) {
	p := TabViewPainter{Theme: nil}
	canvas := &tabMockCanvas{}

	ps := tabview.PaintState{
		Bounds: geometry.NewRect(0, 252, 300, 300),
		Tabs: []tabview.TabState{
			{Label: "Tab1", Bounds: geometry.NewRect(0, 252, 300, 300), Selected: true},
		},
		SelectedIdx: 0,
		Position:    tabview.Bottom,
	}

	p.PaintTabBar(canvas, ps)
}

func TestTabViewPainter_EmptyBounds(t *testing.T) {
	p := TabViewPainter{Theme: nil}
	canvas := &tabMockCanvas{}

	// Should not panic with empty bounds.
	p.PaintTabBar(canvas, tabview.PaintState{})

	if canvas.drawRectCount != 0 {
		t.Error("should not draw anything with empty bounds")
	}
}

func TestTabViewPainter_WithColorScheme(t *testing.T) {
	p := TabViewPainter{Theme: nil}
	canvas := &tabMockCanvas{}

	cs := tabview.TabColorScheme{
		Background:      widget.Hex(0xFF0000),
		SelectedText:    widget.Hex(0x00FF00),
		UnselectedText:  widget.Hex(0x0000FF),
		Indicator:       widget.Hex(0xFF00FF),
		HoverBackground: widget.Hex(0x00FFFF),
		CloseButton:     widget.Hex(0xFFFF00),
		FocusRing:       widget.Hex(0x808080),
	}

	ps := tabview.PaintState{
		Bounds: geometry.NewRect(0, 0, 200, 48),
		Tabs: []tabview.TabState{
			{Label: "Tab1", Bounds: geometry.NewRect(0, 0, 200, 48), Selected: true},
		},
		SelectedIdx: 0,
		Position:    tabview.Top,
		ColorScheme: cs,
	}

	p.PaintTabBar(canvas, ps)
}

func TestTabViewPainter_EmptyTabSlice(t *testing.T) {
	p := TabViewPainter{Theme: nil}
	canvas := &tabMockCanvas{}

	ps := tabview.PaintState{
		Bounds: geometry.NewRect(0, 0, 300, 48),
		Tabs:   nil,
	}

	// Should not panic.
	p.PaintTabBar(canvas, ps)
}

func TestTabViewPainter_SelectedOutOfRange(t *testing.T) {
	p := TabViewPainter{Theme: nil}
	canvas := &tabMockCanvas{}

	ps := tabview.PaintState{
		Bounds: geometry.NewRect(0, 0, 300, 48),
		Tabs: []tabview.TabState{
			{Label: "Tab1", Bounds: geometry.NewRect(0, 0, 300, 48)},
		},
		SelectedIdx: 5, // Out of range.
	}

	// Should not panic.
	p.PaintTabBar(canvas, ps)
}

// tabMockCanvas counts draw operations.
type tabMockCanvas struct {
	drawRectCount        int
	strokeRoundRectCount int
	drawLineCount        int
}

func (c *tabMockCanvas) Clear(_ widget.Color) {}

func (c *tabMockCanvas) DrawRect(_ geometry.Rect, _ widget.Color) {
	c.drawRectCount++
}

func (c *tabMockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) {}

func (c *tabMockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32) {}

func (c *tabMockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {
	c.strokeRoundRectCount++
}

func (c *tabMockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color) {}

func (c *tabMockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {}

func (c *tabMockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32) {
	c.drawLineCount++
}

func (c *tabMockCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
}

func (c *tabMockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}

func (c *tabMockCanvas) DrawImage(_ image.Image, _ geometry.Point)    {}
func (c *tabMockCanvas) PushClip(_ geometry.Rect)                     {}
func (c *tabMockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *tabMockCanvas) PopClip()                                     {}
func (c *tabMockCanvas) PushTransform(_ geometry.Point)               {}
func (c *tabMockCanvas) PopTransform()                                {}
func (c *tabMockCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }
