package material3_test

import (
	"testing"

	"github.com/gogpu/ui/core/checkbox"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

// checkboxTestBounds returns a standard non-empty rectangle for checkbox tests.
func checkboxTestBounds() geometry.Rect {
	return geometry.NewRect(10, 10, 120, 40)
}

func TestCheckboxPainter_Implements_Interface(t *testing.T) {
	// Compile-time check: assignment would fail if CheckboxPainter
	// did not implement checkbox.Painter.
	var _ checkbox.Painter = material3.CheckboxPainter{}
}

func TestCheckboxPainter_EmptyBounds(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.CheckboxPainter{}

	painter.PaintCheckbox(canvas, checkbox.PaintState{
		Label:  "Test",
		Bounds: geometry.Rect{}, // empty bounds
	})

	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no draw calls, got %d", len(canvas.calls))
	}
}

func TestCheckboxPainter_Unchecked(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.CheckboxPainter{}

	painter.PaintCheckbox(canvas, checkbox.PaintState{
		Label:   "Accept",
		Checked: false,
		Bounds:  checkboxTestBounds(),
	})

	// Unchecked: StrokeRoundRect (border) + DrawText (label).
	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) != 1 {
		t.Errorf("Unchecked should draw 1 StrokeRoundRect (border), got %d", len(strokes))
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("Unchecked should draw 1 DrawText (label), got %d", len(texts))
	}
	if texts[0].text != "Accept" {
		t.Errorf("text should be 'Accept', got %q", texts[0].text)
	}
}

func TestCheckboxPainter_Checked(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.CheckboxPainter{}

	painter.PaintCheckbox(canvas, checkbox.PaintState{
		Label:   "Accept",
		Checked: true,
		Bounds:  checkboxTestBounds(),
	})

	// Checked: DrawRoundRect (filled box) + 2x DrawLine (checkmark) + DrawText (label).
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Errorf("Checked should draw 1 DrawRoundRect (filled box), got %d", len(roundRects))
	}

	lines := canvas.methodCalls(methodDrawLine)
	if len(lines) != 2 {
		t.Errorf("Checked should draw 2 DrawLine (checkmark), got %d", len(lines))
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("Checked should draw 1 DrawText (label), got %d", len(texts))
	}
}

func TestCheckboxPainter_Indeterminate(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.CheckboxPainter{}

	painter.PaintCheckbox(canvas, checkbox.PaintState{
		Label:         "Select all",
		Indeterminate: true,
		Bounds:        checkboxTestBounds(),
	})

	// Indeterminate: DrawRoundRect (filled box) + 1x DrawLine (dash) + DrawText (label).
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Errorf("Indeterminate should draw 1 DrawRoundRect, got %d", len(roundRects))
	}

	lines := canvas.methodCalls(methodDrawLine)
	if len(lines) != 1 {
		t.Errorf("Indeterminate should draw 1 DrawLine (dash), got %d", len(lines))
	}
}

func TestCheckboxPainter_Disabled(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.CheckboxPainter{}

	painter.PaintCheckbox(canvas, checkbox.PaintState{
		Label:    "Disabled",
		Checked:  true,
		Disabled: true,
		Bounds:   checkboxTestBounds(),
	})

	// Disabled + Checked: should draw filled box + checkmark + text, no focus ring.
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Errorf("Disabled checked should draw 1 DrawRoundRect, got %d", len(roundRects))
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("Disabled should draw 1 DrawText, got %d", len(texts))
	}
}

func TestCheckboxPainter_Focused(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.CheckboxPainter{}

	painter.PaintCheckbox(canvas, checkbox.PaintState{
		Label:   "Focused",
		Checked: false,
		Focused: true,
		Bounds:  checkboxTestBounds(),
	})

	// Focused: should draw border + text + focus ring (StrokeRoundRect).
	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) < 2 {
		t.Errorf("Focused unchecked should draw at least 2 StrokeRoundRect (border + focus ring), got %d", len(strokes))
	}
}

func TestCheckboxPainter_FocusedDisabled(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.CheckboxPainter{}

	painter.PaintCheckbox(canvas, checkbox.PaintState{
		Label:    "Focused+Disabled",
		Checked:  false,
		Focused:  true,
		Disabled: true,
		Bounds:   checkboxTestBounds(),
	})

	// Focused+Disabled: should NOT draw focus ring.
	// Only 1 StrokeRoundRect for the border, no extra for focus.
	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) != 1 {
		t.Errorf("Focused+Disabled should draw 1 StrokeRoundRect (border only), got %d", len(strokes))
	}
}

func TestCheckboxPainter_NoLabel(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.CheckboxPainter{}

	painter.PaintCheckbox(canvas, checkbox.PaintState{
		Label:   "",
		Checked: false,
		Bounds:  checkboxTestBounds(),
	})

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 0 {
		t.Errorf("No label should produce 0 DrawText, got %d", len(texts))
	}
}

func TestCheckboxPainter_CustomBackground(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.CheckboxPainter{}

	customBg := widget.Hex(0xFF0000) // red
	painter.PaintCheckbox(canvas, checkbox.PaintState{
		Label:      "Custom",
		Checked:    true,
		Bounds:     checkboxTestBounds(),
		Background: &customBg,
	})

	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Fatalf("Custom bg should draw 1 DrawRoundRect, got %d", len(roundRects))
	}

	got := roundRects[0].color
	if got != customBg {
		t.Errorf("custom background: got %v, want %v", got, customBg)
	}
}

func TestCheckboxPainter_WithTheme(t *testing.T) {
	// CheckboxPainter with the default M3 purple seed should produce
	// output consistent with a nil-Theme painter.
	defaultSeed := widget.Hex(0x6750A4)
	painterWithTheme := material3.CheckboxPainter{Theme: material3.New(defaultSeed)}
	painterNilTheme := material3.CheckboxPainter{}

	canvasA := &recordCanvas{}
	canvasB := &recordCanvas{}

	state := checkbox.PaintState{
		Label:   "Test",
		Checked: true,
		Bounds:  checkboxTestBounds(),
	}

	painterWithTheme.PaintCheckbox(canvasA, state)
	painterNilTheme.PaintCheckbox(canvasB, state)

	if len(canvasA.calls) == 0 {
		t.Fatal("themed painter should produce draw calls")
	}
	if len(canvasB.calls) == 0 {
		t.Fatal("nil-theme painter should produce draw calls")
	}

	if len(canvasA.calls) != len(canvasB.calls) {
		t.Errorf("call count mismatch: themed=%d, nil-theme=%d",
			len(canvasA.calls), len(canvasB.calls))
	}
}

func TestCheckboxPainter_NilTheme_Fallback(t *testing.T) {
	painter := material3.CheckboxPainter{}
	canvas := &recordCanvas{}

	painter.PaintCheckbox(canvas, checkbox.PaintState{
		Label:   "Default",
		Checked: true,
		Bounds:  checkboxTestBounds(),
	})

	// Should produce DrawRoundRect (filled box) + DrawLine (checkmark) + DrawText (label).
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Fatalf("nil-theme Checked should draw 1 DrawRoundRect, got %d", len(roundRects))
	}

	lines := canvas.methodCalls(methodDrawLine)
	if len(lines) != 2 {
		t.Fatalf("nil-theme Checked should draw 2 DrawLine (checkmark), got %d", len(lines))
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("nil-theme Checked should draw 1 DrawText, got %d", len(texts))
	}
}

func TestCheckboxPainter_CustomSeed_DifferentColors(t *testing.T) {
	redTheme := material3.New(widget.Hex(0xFF0000))
	purpleTheme := material3.New(widget.Hex(0x6750A4))

	painterRed := material3.CheckboxPainter{Theme: redTheme}
	painterPurple := material3.CheckboxPainter{Theme: purpleTheme}

	canvasRed := &recordCanvas{}
	canvasPurple := &recordCanvas{}

	state := checkbox.PaintState{
		Label:   "Color",
		Checked: true,
		Bounds:  checkboxTestBounds(),
	}

	painterRed.PaintCheckbox(canvasRed, state)
	painterPurple.PaintCheckbox(canvasPurple, state)

	redRects := canvasRed.methodCalls(methodDrawRoundRect)
	purpleRects := canvasPurple.methodCalls(methodDrawRoundRect)

	if len(redRects) != 1 || len(purpleRects) != 1 {
		t.Fatalf("both should draw 1 DrawRoundRect, got red=%d purple=%d",
			len(redRects), len(purpleRects))
	}

	if redRects[0].color == purpleRects[0].color {
		t.Error("red and purple themes should produce different checked backgrounds")
	}
}

func TestCheckboxPainter_HoverState(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.CheckboxPainter{}

	painter.PaintCheckbox(canvas, checkbox.PaintState{
		Label:   "Hover",
		Checked: true,
		Hovered: true,
		Bounds:  checkboxTestBounds(),
	})

	// Hovered color should differ from normal.
	canvasNormal := &recordCanvas{}
	painter.PaintCheckbox(canvasNormal, checkbox.PaintState{
		Label:   "Normal",
		Checked: true,
		Bounds:  checkboxTestBounds(),
	})

	hoveredRects := canvas.methodCalls(methodDrawRoundRect)
	normalRects := canvasNormal.methodCalls(methodDrawRoundRect)

	if len(hoveredRects) != 1 || len(normalRects) != 1 {
		t.Fatalf("both should draw 1 DrawRoundRect, got hover=%d normal=%d",
			len(hoveredRects), len(normalRects))
	}

	if hoveredRects[0].color == normalRects[0].color {
		t.Error("hovered color should differ from normal color")
	}
}

func TestCheckboxPainter_PressedState(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.CheckboxPainter{}

	painter.PaintCheckbox(canvas, checkbox.PaintState{
		Label:   "Pressed",
		Checked: true,
		Pressed: true,
		Bounds:  checkboxTestBounds(),
	})

	canvasNormal := &recordCanvas{}
	painter.PaintCheckbox(canvasNormal, checkbox.PaintState{
		Label:   "Normal",
		Checked: true,
		Bounds:  checkboxTestBounds(),
	})

	pressedRects := canvas.methodCalls(methodDrawRoundRect)
	normalRects := canvasNormal.methodCalls(methodDrawRoundRect)

	if len(pressedRects) != 1 || len(normalRects) != 1 {
		t.Fatalf("both should draw 1 DrawRoundRect, got pressed=%d normal=%d",
			len(pressedRects), len(normalRects))
	}

	if pressedRects[0].color == normalRects[0].color {
		t.Error("pressed color should differ from normal color")
	}
}
