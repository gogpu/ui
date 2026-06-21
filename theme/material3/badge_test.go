package material3

import (
	"testing"

	"github.com/gogpu/ui/core/badge"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

func TestBadgePainter_CompileTimeCheck(t *testing.T) {
	var _ badge.Painter = BadgePainter{}
}

func TestBadgePainter_EmptyBounds(t *testing.T) {
	p := BadgePainter{}
	canvas := &pbMockCanvas{}

	p.PaintBadge(canvas, badge.PaintState{})

	if canvas.drawCount > 0 {
		t.Error("should not draw anything with empty bounds")
	}
}

func TestBadgePainter_NilTheme_UsesDefaults(t *testing.T) {
	p := BadgePainter{Theme: nil}
	canvas := &pbMockCanvas{}

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
	theme := New(widget.Hex(0x6750A4))
	p := BadgePainter{Theme: theme}
	canvas := &pbMockCanvas{}

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
	canvas := &pbMockCanvas{}

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
	canvas := &pbMockCanvas{}

	ps := badge.PaintState{
		Bounds: geometry.NewRect(0, 0, 8, 8),
		Dot:    true,
	}

	p.PaintBadge(canvas, ps)

	if canvas.drawCount == 0 {
		t.Error("dot badge should draw a circle")
	}
	// Dot mode draws a circle, not text.
	if canvas.drawTextCount != 0 {
		t.Errorf("dot mode should not draw text, got %d", canvas.drawTextCount)
	}
}

func TestBadgePainter_CountMode(t *testing.T) {
	p := BadgePainter{}
	canvas := &pbMockCanvas{}

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
	canvas := &pbMockCanvas{}

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

func TestBadgePainter_ResolveColors_NilTheme(t *testing.T) {
	p := BadgePainter{Theme: nil}
	colors := p.resolveColors()

	if colors == (badge.BadgeColorScheme{}) {
		t.Error("should return non-zero default colors")
	}
}

func TestBadgePainter_ResolveColors_WithTheme(t *testing.T) {
	theme := New(widget.Hex(0xFF0000))
	p := BadgePainter{Theme: theme}
	colors := p.resolveColors()

	if colors == (badge.BadgeColorScheme{}) {
		t.Error("should return non-zero colors from theme")
	}
	// Error color from a red seed should differ from default M3 purple-derived error.
	if colors.Background == m3DefaultBadgeColors.Background {
		t.Error("themed colors should differ from default palette")
	}
}
