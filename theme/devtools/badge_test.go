package devtools

import (
	"github.com/gogpu/gg/scene"
	"image"
	"testing"

	"github.com/gogpu/ui/core/badge"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// --- BadgePainter Tests ---

func TestBadgePainter_CompileTimeCheck(t *testing.T) {
	var _ badge.Painter = BadgePainter{}
}

func TestBadgePainter_EmptyBounds(t *testing.T) {
	p := BadgePainter{}
	canvas := &bgMockCanvas{}

	p.PaintBadge(canvas, badge.PaintState{})

	if canvas.drawCount > 0 {
		t.Error("should not draw anything with empty bounds")
	}
}

func TestBadgePainter_NilTheme_UsesDefaults(t *testing.T) {
	p := BadgePainter{Theme: nil}
	canvas := &bgMockCanvas{}

	ps := badge.PaintState{
		Bounds: geometry.NewRect(0, 0, 30, 16),
		Label:  "5",
	}

	p.PaintBadge(canvas, ps)

	if canvas.drawCount == 0 {
		t.Error("should draw with nil theme (default colors)")
	}
}

func TestBadgePainter_WithTheme(t *testing.T) {
	p := BadgePainter{Theme: NewDarkTheme()}
	canvas := &bgMockCanvas{}

	ps := badge.PaintState{
		Bounds: geometry.NewRect(0, 0, 30, 16),
		Label:  "3",
	}

	p.PaintBadge(canvas, ps)

	if canvas.drawCount == 0 {
		t.Error("should draw with theme")
	}
}

func TestBadgePainter_Disabled(t *testing.T) {
	p := BadgePainter{}
	canvas := &bgMockCanvas{}

	ps := badge.PaintState{
		Bounds:   geometry.NewRect(0, 0, 30, 16),
		Label:    "5",
		Disabled: true,
	}

	p.PaintBadge(canvas, ps)

	if canvas.drawCount == 0 {
		t.Error("disabled badge should still draw")
	}
}

func TestBadgePainter_DotMode(t *testing.T) {
	p := BadgePainter{}
	canvas := &bgMockCanvas{}

	ps := badge.PaintState{
		Bounds: geometry.NewRect(0, 0, 8, 8),
		Dot:    true,
	}

	p.PaintBadge(canvas, ps)

	if canvas.drawCircleCount == 0 {
		t.Error("dot badge should draw a circle")
	}
	if canvas.drawTextCount != 0 {
		t.Errorf("dot mode should not draw text, got %d", canvas.drawTextCount)
	}
}

func TestBadgePainter_CountMode(t *testing.T) {
	p := BadgePainter{}
	canvas := &bgMockCanvas{}

	ps := badge.PaintState{
		Bounds: geometry.NewRect(0, 0, 30, 16),
		Label:  "42",
	}

	p.PaintBadge(canvas, ps)

	if canvas.drawRoundRectCount == 0 {
		t.Error("count mode should draw a round rect (pill)")
	}
	if canvas.drawTextCount == 0 {
		t.Error("count mode should draw text label")
	}
}

func TestBadgePainter_WithColorScheme(t *testing.T) {
	p := BadgePainter{}
	canvas := &bgMockCanvas{}

	scheme := badge.BadgeColorScheme{
		Background:         widget.ColorRed,
		Label:              widget.ColorWhite,
		DisabledBackground: widget.ColorDarkGray,
		DisabledLabel:      widget.ColorLightGray,
	}

	ps := badge.PaintState{
		Bounds:      geometry.NewRect(0, 0, 30, 16),
		Label:       "7",
		ColorScheme: scheme,
	}

	p.PaintBadge(canvas, ps)

	if canvas.drawCount == 0 {
		t.Error("should draw with custom color scheme")
	}
}

func TestBadgePainter_resolveColors_NilTheme(t *testing.T) {
	p := BadgePainter{Theme: nil}
	colors := p.resolveColors()

	if colors == (badge.BadgeColorScheme{}) {
		t.Error("should return non-zero default colors")
	}
	if colors.Background != dtDefaultBadgeColors.Background {
		t.Error("nil theme should return default badge background")
	}
}

func TestBadgePainter_resolveColors_WithTheme(t *testing.T) {
	dt := NewDarkTheme()
	p := BadgePainter{Theme: dt}
	colors := p.resolveColors()

	if colors == (badge.BadgeColorScheme{}) {
		t.Error("should return non-zero colors from theme")
	}
	if colors.Background != dt.Colors.Error {
		t.Errorf("badge background should be theme Error color")
	}
}

func TestBadgePainter_LightTheme(t *testing.T) {
	p := BadgePainter{Theme: NewTheme()}
	canvas := &bgMockCanvas{}

	ps := badge.PaintState{
		Bounds: geometry.NewRect(0, 0, 30, 16),
		Label:  "9",
	}

	p.PaintBadge(canvas, ps)

	if canvas.drawCount == 0 {
		t.Error("light theme should draw badge")
	}
}

// --- Mock Canvas for badge tests ---

type bgMockCanvas struct {
	drawCount          int
	drawCircleCount    int
	drawRoundRectCount int
	drawTextCount      int
}

func (c *bgMockCanvas) Clear(_ widget.Color)                                  {}
func (c *bgMockCanvas) DrawRect(_ geometry.Rect, _ widget.Color)              { c.drawCount++ }
func (c *bgMockCanvas) FillRectDirect(_ geometry.Rect, _ widget.Color)        {}
func (c *bgMockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) { c.drawCount++ }
func (c *bgMockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32) {
	c.drawCount++
	c.drawRoundRectCount++
}
func (c *bgMockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {
	c.drawCount++
}
func (c *bgMockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color) {
	c.drawCount++
	c.drawCircleCount++
}
func (c *bgMockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {
	c.drawCount++
}
func (c *bgMockCanvas) StrokeArc(_ geometry.Point, _ float32, _, _ float64, _ widget.Color, _ float32) {
}
func (c *bgMockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32) { c.drawCount++ }
func (c *bgMockCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
	c.drawCount++
	c.drawTextCount++
}
func (c *bgMockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}
func (c *bgMockCanvas) DrawImage(_ image.Image, _ geometry.Point)    { c.drawCount++ }
func (c *bgMockCanvas) PushClip(_ geometry.Rect)                     {}
func (c *bgMockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *bgMockCanvas) PopClip()                                     {}
func (c *bgMockCanvas) PushTransform(_ geometry.Point)               {}
func (c *bgMockCanvas) PopTransform()                                {}
func (c *bgMockCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }
func (c *bgMockCanvas) ScreenOriginBase() geometry.Point             { return geometry.Point{} }
func (c *bgMockCanvas) ClipBounds() geometry.Rect                    { return geometry.NewRect(0, 0, 10000, 10000) }
func (c *bgMockCanvas) ReplayScene(_ *scene.Scene)                   {}
