package cupertino_test

import (
	"testing"

	"github.com/gogpu/ui/core/chip"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/theme/cupertino"
	"github.com/gogpu/ui/widget"
)

// --- Chip painter tests ---

func TestChipPainterImplementsInterface(t *testing.T) {
	var _ chip.Painter = cupertino.ChipPainter{}
}

func TestChipPaintEmptyBounds(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.ChipPainter{}
	painter.PaintChip(canvas, chip.PaintState{
		Label:  "Tag",
		Bounds: geometry.Rect{},
	})
	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no draw calls, got %d", len(canvas.calls))
	}
}

func TestChipPaintUnselected(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.ChipPainter{}
	painter.PaintChip(canvas, chip.PaintState{
		Label:  "Tag",
		Bounds: testBounds(),
		Radius: 10,
	})

	// Unselected outlined: StrokeRoundRect (border) + DrawText.
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

func TestChipPaintSelected(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.ChipPainter{}
	painter.PaintChip(canvas, chip.PaintState{
		Label:    "Filter",
		Bounds:   testBounds(),
		Radius:   10,
		Selected: true,
	})

	// Selected: DrawRoundRect (selected fill) + DrawText.
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Errorf("selected chip should draw 1 DrawRoundRect, got %d", len(roundRects))
	}
}

func TestChipPaintDisabled(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.ChipPainter{}
	painter.PaintChip(canvas, chip.PaintState{
		Label:    "Disabled",
		Bounds:   testBounds(),
		Radius:   10,
		Disabled: true,
	})

	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Errorf("disabled chip should draw 1 DrawRoundRect, got %d", len(roundRects))
	}
}

func TestChipPaintFocused(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.ChipPainter{}
	painter.PaintChip(canvas, chip.PaintState{
		Label:   "Focus",
		Bounds:  testBounds(),
		Radius:  10,
		Focused: true,
	})

	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) < 2 { //nolint:mnd // border + focus ring
		t.Errorf("focused chip should draw at least 2 StrokeRoundRect (border + focus ring), got %d", len(strokes))
	}
}

func TestChipPaintHovered(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.ChipPainter{}
	painter.PaintChip(canvas, chip.PaintState{
		Label:   "Hover",
		Bounds:  testBounds(),
		Radius:  10,
		Hovered: true,
	})

	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Errorf("hovered chip should draw 1 DrawRoundRect (state layer), got %d", len(roundRects))
	}
}

func TestChipPaintWithTheme(t *testing.T) {
	theme := cupertino.NewTheme()
	painter := cupertino.ChipPainter{Theme: theme}
	canvas := &recordCanvas{}

	painter.PaintChip(canvas, chip.PaintState{
		Label:  "Themed",
		Bounds: testBounds(),
		Radius: 10,
	})

	if len(canvas.calls) == 0 {
		t.Fatal("themed painter should produce draw calls")
	}
}

func TestChipPaintWithDarkTheme(t *testing.T) {
	theme := cupertino.NewDarkTheme()
	painter := cupertino.ChipPainter{Theme: theme}
	canvas := &recordCanvas{}

	painter.PaintChip(canvas, chip.PaintState{
		Label:    "Dark",
		Bounds:   testBounds(),
		Radius:   10,
		Selected: true,
	})

	if len(canvas.calls) == 0 {
		t.Fatal("dark theme painter should produce draw calls")
	}
}

func TestChipPaintColorSchemeOverride(t *testing.T) {
	painter := cupertino.ChipPainter{}
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
		Radius:      10,
		ColorScheme: customColors,
	})

	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Errorf("custom colors should draw 1 DrawRoundRect (bg), got %d", len(roundRects))
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("expected 1 DrawText, got %d", len(texts))
	}
	if texts[0].color != widget.Hex(0x00FF00) {
		t.Errorf("label color should be custom green, got %v", texts[0].color)
	}
}

func TestChipPaintCustomAccent(t *testing.T) {
	greenTheme := cupertino.NewTheme(cupertino.WithAccentColor(widget.Hex(0x34C759)))
	blueTheme := cupertino.NewTheme()

	canvasGreen := &recordCanvas{}
	canvasBlue := &recordCanvas{}

	state := chip.PaintState{
		Label:    "Color",
		Bounds:   testBounds(),
		Radius:   10,
		Selected: true,
	}

	cupertino.ChipPainter{Theme: greenTheme}.PaintChip(canvasGreen, state)
	cupertino.ChipPainter{Theme: blueTheme}.PaintChip(canvasBlue, state)

	greenRects := canvasGreen.methodCalls(methodDrawRoundRect)
	blueRects := canvasBlue.methodCalls(methodDrawRoundRect)

	if len(greenRects) < 1 || len(blueRects) < 1 {
		t.Fatalf("both should draw DrawRoundRect, got green=%d blue=%d",
			len(greenRects), len(blueRects))
	}

	if greenRects[0].color == blueRects[0].color {
		t.Error("green and blue accents should produce different selected backgrounds")
	}
}
