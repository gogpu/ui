package fluent

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
	canvas := &flBadgeMockCanvas{}

	p.PaintBadge(canvas, badge.PaintState{})

	if canvas.drawCount > 0 {
		t.Error("should not draw anything with empty bounds")
	}
}

func TestBadgePainter_NilTheme_UsesDefaults(t *testing.T) {
	p := BadgePainter{Theme: nil}
	canvas := &flBadgeMockCanvas{}

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
	p := BadgePainter{Theme: NewTheme()}
	canvas := &flBadgeMockCanvas{}

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
	canvas := &flBadgeMockCanvas{}

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
	canvas := &flBadgeMockCanvas{}

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
	canvas := &flBadgeMockCanvas{}

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
	canvas := &flBadgeMockCanvas{}

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
	if colors.Background != flDefaultBadgeColors.Background {
		t.Error("nil theme should return default badge background")
	}
}

func TestBadgePainter_resolveColors_WithTheme(t *testing.T) {
	theme := NewTheme()
	p := BadgePainter{Theme: theme}
	colors := p.resolveColors()

	if colors == (badge.BadgeColorScheme{}) {
		t.Error("should return non-zero colors from theme")
	}
	if colors.Background != theme.Colors.Accent {
		t.Error("badge background should be theme Accent color")
	}
}

func TestBadgePainter_DarkTheme(t *testing.T) {
	p := BadgePainter{Theme: NewDarkTheme()}
	canvas := &flBadgeMockCanvas{}

	ps := badge.PaintState{
		Bounds: geometry.NewRect(0, 0, 30, 16),
		Label:  "9",
	}

	p.PaintBadge(canvas, ps)

	if canvas.drawCount == 0 {
		t.Error("dark theme should draw badge")
	}
}

// --- Mock Canvas for badge tests ---

type flBadgeMockCanvas struct {
	drawCount          int
	drawCircleCount    int
	drawRoundRectCount int
	drawTextCount      int
}

func (c *flBadgeMockCanvas) Clear(_ widget.Color)                                  {}
func (c *flBadgeMockCanvas) DrawRect(_ geometry.Rect, _ widget.Color)              { c.drawCount++ }
func (c *flBadgeMockCanvas) FillRectDirect(_ geometry.Rect, _ widget.Color)        {}
func (c *flBadgeMockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) { c.drawCount++ }
func (c *flBadgeMockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32) {
	c.drawCount++
	c.drawRoundRectCount++
}
func (c *flBadgeMockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {
	c.drawCount++
}
func (c *flBadgeMockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color) {
	c.drawCount++
	c.drawCircleCount++
}
func (c *flBadgeMockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {
	c.drawCount++
}
func (c *flBadgeMockCanvas) StrokeArc(_ geometry.Point, _ float32, _, _ float64, _ widget.Color, _ float32) {
}
func (c *flBadgeMockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32) { c.drawCount++ }
func (c *flBadgeMockCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
	c.drawCount++
	c.drawTextCount++
}
func (c *flBadgeMockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}
func (c *flBadgeMockCanvas) DrawImage(_ image.Image, _ geometry.Point)    { c.drawCount++ }
func (c *flBadgeMockCanvas) PushClip(_ geometry.Rect)                     {}
func (c *flBadgeMockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *flBadgeMockCanvas) PopClip()                                     {}
func (c *flBadgeMockCanvas) PushTransform(_ geometry.Point)               {}
func (c *flBadgeMockCanvas) PopTransform()                                {}
func (c *flBadgeMockCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }
func (c *flBadgeMockCanvas) ScreenOriginBase() geometry.Point             { return geometry.Point{} }
func (c *flBadgeMockCanvas) ClipBounds() geometry.Rect                    { return geometry.NewRect(0, 0, 10000, 10000) }
func (c *flBadgeMockCanvas) ReplayScene(_ *scene.Scene)                   {}
