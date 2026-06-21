package devtools_test

import (
	"testing"

	"github.com/gogpu/ui/core/chip"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/theme/devtools"
	"github.com/gogpu/ui/widget"
)

// --- Chip painter tests ---

func TestChipPainterImplementsInterface(t *testing.T) {
	var _ chip.Painter = devtools.ChipPainter{}
}

func TestChipPaintEmptyBounds(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.ChipPainter{}
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
	painter := devtools.ChipPainter{}
	painter.PaintChip(canvas, chip.PaintState{
		Label:  "Tag",
		Bounds: testBounds(),
		Radius: 8,
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
	painter := devtools.ChipPainter{}
	painter.PaintChip(canvas, chip.PaintState{
		Label:    "Filter",
		Bounds:   testBounds(),
		Radius:   8,
		Selected: true,
	})

	// Selected: DrawRoundRect (selected fill) + DrawText.
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
	painter := devtools.ChipPainter{}
	painter.PaintChip(canvas, chip.PaintState{
		Label:    "Disabled",
		Bounds:   testBounds(),
		Radius:   8,
		Disabled: true,
	})

	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Errorf("disabled chip should draw 1 DrawRoundRect, got %d", len(roundRects))
	}

	// No focus ring when disabled.
	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) != 0 {
		t.Errorf("disabled chip should not draw focus ring, got %d StrokeRoundRect", len(strokes))
	}
}

func TestChipPaintFocused(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.ChipPainter{}
	painter.PaintChip(canvas, chip.PaintState{
		Label:   "Focus",
		Bounds:  testBounds(),
		Radius:  8,
		Focused: true,
	})

	// Focused: border + focus ring = at least 2 StrokeRoundRect.
	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) < 2 { //nolint:mnd // border + focus ring
		t.Errorf("focused chip should draw at least 2 StrokeRoundRect (border + focus ring), got %d", len(strokes))
	}
}

func TestChipPaintHovered(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.ChipPainter{}
	painter.PaintChip(canvas, chip.PaintState{
		Label:   "Hover",
		Bounds:  testBounds(),
		Radius:  8,
		Hovered: true,
	})

	// Hovered: state layer overlay = DrawRoundRect.
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Errorf("hovered chip should draw 1 DrawRoundRect (state layer), got %d", len(roundRects))
	}
}

func TestChipPaintWithDarkTheme(t *testing.T) {
	theme := devtools.NewDarkTheme()
	painter := devtools.ChipPainter{Theme: theme}
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

func TestChipPaintWithLightTheme(t *testing.T) {
	theme := devtools.NewTheme()
	painter := devtools.ChipPainter{Theme: theme}
	canvas := &recordCanvas{}

	painter.PaintChip(canvas, chip.PaintState{
		Label:    "Light",
		Bounds:   testBounds(),
		Radius:   8,
		Selected: true,
	})

	if len(canvas.calls) == 0 {
		t.Fatal("light theme painter should produce draw calls")
	}
}

func TestChipPaintColorSchemeOverride(t *testing.T) {
	painter := devtools.ChipPainter{}
	canvas := &recordCanvas{}

	customColors := chip.ChipColorScheme{
		Background:         widget.Hex(0xFF0000),
		Border:             widget.Hex(0x00FF00),
		Label:              widget.Hex(0x0000FF),
		SelectedBackground: widget.Hex(0xFFFF00),
		SelectedLabel:      widget.Hex(0xFF00FF),
		DisabledBackground: widget.Hex(0x888888),
		DisabledLabel:      widget.Hex(0xAAAAAA),
	}

	painter.PaintChip(canvas, chip.PaintState{
		Label:       "Custom",
		Bounds:      testBounds(),
		Radius:      8,
		ColorScheme: customColors,
	})

	// With a non-transparent background, should draw it.
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Errorf("custom color scheme should draw 1 DrawRoundRect (bg), got %d", len(roundRects))
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("expected 1 DrawText, got %d", len(texts))
	}
	if texts[0].color != widget.Hex(0x0000FF) {
		t.Errorf("label color should be custom blue, got %v", texts[0].color)
	}
}
