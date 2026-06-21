package material3_test

import (
	"testing"

	"github.com/gogpu/ui/core/chip"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func TestChipPainterImplementsInterface(t *testing.T) {
	var _ chip.Painter = material3.ChipPainter{}
}

func TestChipPaintEmptyBounds(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.ChipPainter{}

	painter.PaintChip(canvas, chip.PaintState{
		Label:  "Tag",
		Bounds: geometry.Rect{}, // empty bounds
	})

	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no draw calls, got %d", len(canvas.calls))
	}
}

func TestChipPaintNilThemeUnselected(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.ChipPainter{} // nil Theme

	painter.PaintChip(canvas, chip.PaintState{
		Label:  "Tag",
		Bounds: testBounds(),
		Radius: 8,
	})

	// Unselected: outlined style = StrokeRoundRect (border) + DrawText.
	// No DrawRoundRect because background is transparent.
	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) != 1 {
		t.Errorf("unselected chip should draw 1 StrokeRoundRect (border), got %d", len(strokes))
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("unselected chip should draw 1 DrawText, got %d", len(texts))
	}
	if texts[0].text != "Tag" {
		t.Errorf("label = %q, want 'Tag'", texts[0].text)
	}
}

func TestChipPaintNilThemeSelected(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.ChipPainter{}

	painter.PaintChip(canvas, chip.PaintState{
		Label:    "Filter",
		Bounds:   testBounds(),
		Radius:   8,
		Selected: true,
	})

	// Selected: DrawRoundRect (selected background) + DrawText.
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Errorf("selected chip should draw 1 DrawRoundRect, got %d", len(roundRects))
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("selected chip should draw 1 DrawText, got %d", len(texts))
	}
}

func TestChipPaintDisabled(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.ChipPainter{}

	painter.PaintChip(canvas, chip.PaintState{
		Label:    "Disabled",
		Bounds:   testBounds(),
		Radius:   8,
		Disabled: true,
	})

	// Disabled: DrawRoundRect (disabled bg) + DrawText, no state layer, no focus ring.
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Errorf("disabled chip should draw 1 DrawRoundRect (disabled bg), got %d", len(roundRects))
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("disabled chip should draw 1 DrawText, got %d", len(texts))
	}
}

func TestChipPaintFocused(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.ChipPainter{}

	painter.PaintChip(canvas, chip.PaintState{
		Label:   "Focus",
		Bounds:  testBounds(),
		Radius:  8,
		Focused: true,
	})

	// Focused: border + text + focus ring (second StrokeRoundRect).
	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) < 2 { //nolint:mnd // border + focus ring
		t.Errorf("focused chip should draw at least 2 StrokeRoundRect (border + focus ring), got %d", len(strokes))
	}
}

func TestChipPaintFocusedDisabled(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.ChipPainter{}

	painter.PaintChip(canvas, chip.PaintState{
		Label:    "Focus+Disabled",
		Bounds:   testBounds(),
		Radius:   8,
		Focused:  true,
		Disabled: true,
	})

	// Focused+Disabled: no focus ring drawn.
	// Should only have DrawRoundRect (disabled bg) + DrawText.
	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) != 0 {
		t.Errorf("focused+disabled chip should not draw focus ring, got %d StrokeRoundRect", len(strokes))
	}
}

func TestChipPaintHovered(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.ChipPainter{}

	painter.PaintChip(canvas, chip.PaintState{
		Label:   "Hover",
		Bounds:  testBounds(),
		Radius:  8,
		Hovered: true,
	})

	// Hovered: border + state layer (DrawRoundRect) + text.
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Errorf("hovered chip should draw 1 DrawRoundRect (state layer), got %d", len(roundRects))
	}
}

func TestChipPaintPressed(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.ChipPainter{}

	painter.PaintChip(canvas, chip.PaintState{
		Label:   "Press",
		Bounds:  testBounds(),
		Radius:  8,
		Pressed: true,
	})

	// Pressed: border + state layer (DrawRoundRect) + text.
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Errorf("pressed chip should draw 1 DrawRoundRect (state layer), got %d", len(roundRects))
	}
}

func TestChipPaintWithTheme(t *testing.T) {
	m3 := material3.New(widget.Hex(0x6750A4))
	painter := material3.ChipPainter{Theme: m3}
	canvas := &recordCanvas{}

	painter.PaintChip(canvas, chip.PaintState{
		Label:  "Themed",
		Bounds: testBounds(),
		Radius: 8,
	})

	if len(canvas.calls) == 0 {
		t.Fatal("themed painter should produce draw calls")
	}
}

func TestChipPaintWithRedTheme(t *testing.T) {
	redTheme := material3.New(widget.Hex(0xFF0000))
	purpleTheme := material3.New(widget.Hex(0x6750A4))

	canvasRed := &recordCanvas{}
	canvasPurple := &recordCanvas{}

	state := chip.PaintState{
		Label:    "Color",
		Bounds:   testBounds(),
		Radius:   8,
		Selected: true,
	}

	material3.ChipPainter{Theme: redTheme}.PaintChip(canvasRed, state)
	material3.ChipPainter{Theme: purpleTheme}.PaintChip(canvasPurple, state)

	redRects := canvasRed.methodCalls(methodDrawRoundRect)
	purpleRects := canvasPurple.methodCalls(methodDrawRoundRect)

	if len(redRects) < 1 || len(purpleRects) < 1 {
		t.Fatalf("both should draw DrawRoundRect, got red=%d purple=%d",
			len(redRects), len(purpleRects))
	}

	// Selected backgrounds should differ between red and purple themes.
	if redRects[0].color == purpleRects[0].color {
		t.Error("red and purple themes should produce different selected backgrounds")
	}
}

func TestChipPaintColorSchemeOverride(t *testing.T) {
	painter := material3.ChipPainter{}
	canvas := &recordCanvas{}

	customColors := chip.ChipColorScheme{
		Background:         widget.Hex(0xFFEEDD),
		Border:             widget.Hex(0xFF0000),
		Label:              widget.Hex(0x00FF00),
		SelectedBackground: widget.Hex(0x0000FF),
		SelectedLabel:      widget.ColorWhite,
		DisabledBackground: widget.Hex(0x888888),
		DisabledLabel:      widget.Hex(0xAAAAAA),
	}

	painter.PaintChip(canvas, chip.PaintState{
		Label:       "Custom",
		Bounds:      testBounds(),
		Radius:      8,
		ColorScheme: customColors,
	})

	// With a non-transparent Background, DefaultPainter draws it.
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Errorf("custom color scheme should draw 1 DrawRoundRect (bg), got %d", len(roundRects))
	}

	// Label should use the custom Label color.
	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("expected 1 DrawText, got %d", len(texts))
	}
	if texts[0].color != widget.Hex(0x00FF00) {
		t.Errorf("label color should be custom green, got %v", texts[0].color)
	}
}

func TestChipPaintSelectedWithColorSchemeOverride(t *testing.T) {
	painter := material3.ChipPainter{}
	canvas := &recordCanvas{}

	customColors := chip.ChipColorScheme{
		Background:         widget.ColorTransparent,
		Border:             widget.Hex(0xAAAAAA),
		Label:              widget.Hex(0x333333),
		SelectedBackground: widget.Hex(0x0000FF),
		SelectedLabel:      widget.ColorWhite,
		DisabledBackground: widget.Hex(0x888888),
		DisabledLabel:      widget.Hex(0xCCCCCC),
	}

	painter.PaintChip(canvas, chip.PaintState{
		Label:       "Selected",
		Bounds:      testBounds(),
		Radius:      8,
		Selected:    true,
		ColorScheme: customColors,
	})

	// Selected: should use SelectedBackground.
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Fatalf("selected with custom colors should draw 1 DrawRoundRect, got %d", len(roundRects))
	}
	if roundRects[0].color != widget.Hex(0x0000FF) {
		t.Errorf("selected background should be custom blue, got %v", roundRects[0].color)
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("expected 1 DrawText, got %d", len(texts))
	}
	if texts[0].color != widget.ColorWhite {
		t.Errorf("selected label should be white, got %v", texts[0].color)
	}
}
