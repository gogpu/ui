package material3_test

import (
	"testing"

	"github.com/gogpu/ui/core/radio"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

// radioTestBounds returns a standard non-empty rectangle for radio tests.
func radioTestBounds() geometry.Rect {
	return geometry.NewRect(10, 10, 120, 30)
}

func TestRadioPainter_Implements_Interface(t *testing.T) {
	// Compile-time check: assignment would fail if RadioPainter
	// did not implement radio.Painter.
	var _ radio.Painter = material3.RadioPainter{}
}

func TestRadioPainter_EmptyBounds(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.RadioPainter{}

	painter.PaintRadio(canvas, radio.PaintState{
		Label:  "Test",
		Bounds: geometry.Rect{}, // empty bounds
	})

	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no draw calls, got %d", len(canvas.calls))
	}
}

func TestRadioPainter_Unselected(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.RadioPainter{}

	painter.PaintRadio(canvas, radio.PaintState{
		Label:    "Option A",
		Selected: false,
		Bounds:   radioTestBounds(),
	})

	// Unselected: StrokeCircle (border) + DrawText (label).
	strokes := canvas.methodCalls(methodStrokeCircle)
	if len(strokes) != 1 {
		t.Errorf("Unselected should draw 1 StrokeCircle (border), got %d", len(strokes))
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("Unselected should draw 1 DrawText (label), got %d", len(texts))
	}
	if texts[0].text != "Option A" {
		t.Errorf("text should be 'Option A', got %q", texts[0].text)
	}
}

func TestRadioPainter_Selected(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.RadioPainter{}

	painter.PaintRadio(canvas, radio.PaintState{
		Label:    "Option A",
		Selected: true,
		Bounds:   radioTestBounds(),
	})

	// Selected: DrawCircle (outer filled) + DrawCircle (inner dot) + DrawText (label).
	circles := canvas.methodCalls(methodDrawCircle)
	if len(circles) != 2 {
		t.Errorf("Selected should draw 2 DrawCircle (outer + inner), got %d", len(circles))
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("Selected should draw 1 DrawText (label), got %d", len(texts))
	}
}

func TestRadioPainter_Disabled(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.RadioPainter{}

	painter.PaintRadio(canvas, radio.PaintState{
		Label:    "Disabled",
		Selected: true,
		Disabled: true,
		Bounds:   radioTestBounds(),
	})

	// Disabled + Selected: should draw circles + text, no focus ring.
	circles := canvas.methodCalls(methodDrawCircle)
	if len(circles) != 2 {
		t.Errorf("Disabled selected should draw 2 DrawCircle, got %d", len(circles))
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("Disabled should draw 1 DrawText, got %d", len(texts))
	}
}

func TestRadioPainter_Focused(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.RadioPainter{}

	painter.PaintRadio(canvas, radio.PaintState{
		Label:    "Focused",
		Selected: false,
		Focused:  true,
		Bounds:   radioTestBounds(),
	})

	// Focused: should draw border + text + focus ring (StrokeCircle).
	strokes := canvas.methodCalls(methodStrokeCircle)
	if len(strokes) < 2 {
		t.Errorf("Focused unselected should draw at least 2 StrokeCircle (border + focus ring), got %d", len(strokes))
	}
}

func TestRadioPainter_FocusedDisabled(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.RadioPainter{}

	painter.PaintRadio(canvas, radio.PaintState{
		Label:    "Focused+Disabled",
		Selected: false,
		Focused:  true,
		Disabled: true,
		Bounds:   radioTestBounds(),
	})

	// Focused+Disabled: should NOT draw focus ring.
	strokes := canvas.methodCalls(methodStrokeCircle)
	if len(strokes) != 1 {
		t.Errorf("Focused+Disabled should draw 1 StrokeCircle (border only), got %d", len(strokes))
	}
}

func TestRadioPainter_NoLabel(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.RadioPainter{}

	painter.PaintRadio(canvas, radio.PaintState{
		Label:    "",
		Selected: false,
		Bounds:   radioTestBounds(),
	})

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 0 {
		t.Errorf("No label should produce 0 DrawText, got %d", len(texts))
	}
}

func TestRadioPainter_WithTheme(t *testing.T) {
	defaultSeed := widget.Hex(0x6750A4)
	painterWithTheme := material3.RadioPainter{Theme: material3.New(defaultSeed)}
	painterNilTheme := material3.RadioPainter{}

	canvasA := &recordCanvas{}
	canvasB := &recordCanvas{}

	state := radio.PaintState{
		Label:    "Test",
		Selected: true,
		Bounds:   radioTestBounds(),
	}

	painterWithTheme.PaintRadio(canvasA, state)
	painterNilTheme.PaintRadio(canvasB, state)

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

func TestRadioPainter_NilTheme_Fallback(t *testing.T) {
	painter := material3.RadioPainter{}
	canvas := &recordCanvas{}

	painter.PaintRadio(canvas, radio.PaintState{
		Label:    "Default",
		Selected: true,
		Bounds:   radioTestBounds(),
	})

	// Should produce DrawCircle (outer) + DrawCircle (inner dot) + DrawText.
	circles := canvas.methodCalls(methodDrawCircle)
	if len(circles) != 2 {
		t.Fatalf("nil-theme Selected should draw 2 DrawCircle, got %d", len(circles))
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("nil-theme Selected should draw 1 DrawText, got %d", len(texts))
	}
}

func TestRadioPainter_CustomSeed_DifferentColors(t *testing.T) {
	redTheme := material3.New(widget.Hex(0xFF0000))
	purpleTheme := material3.New(widget.Hex(0x6750A4))

	painterRed := material3.RadioPainter{Theme: redTheme}
	painterPurple := material3.RadioPainter{Theme: purpleTheme}

	canvasRed := &recordCanvas{}
	canvasPurple := &recordCanvas{}

	state := radio.PaintState{
		Label:    "Color",
		Selected: true,
		Bounds:   radioTestBounds(),
	}

	painterRed.PaintRadio(canvasRed, state)
	painterPurple.PaintRadio(canvasPurple, state)

	redCircles := canvasRed.methodCalls(methodDrawCircle)
	purpleCircles := canvasPurple.methodCalls(methodDrawCircle)

	if len(redCircles) != 2 || len(purpleCircles) != 2 {
		t.Fatalf("both should draw 2 DrawCircle, got red=%d purple=%d",
			len(redCircles), len(purpleCircles))
	}

	// The outer circle (first) background colors should differ between themes.
	if redCircles[0].color == purpleCircles[0].color {
		t.Error("red and purple themes should produce different selected backgrounds")
	}
}

func TestRadioPainter_HoverState(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.RadioPainter{}

	painter.PaintRadio(canvas, radio.PaintState{
		Label:    "Hover",
		Selected: true,
		Hovered:  true,
		Bounds:   radioTestBounds(),
	})

	canvasNormal := &recordCanvas{}
	painter.PaintRadio(canvasNormal, radio.PaintState{
		Label:    "Normal",
		Selected: true,
		Bounds:   radioTestBounds(),
	})

	hoveredCircles := canvas.methodCalls(methodDrawCircle)
	normalCircles := canvasNormal.methodCalls(methodDrawCircle)

	if len(hoveredCircles) != 2 || len(normalCircles) != 2 {
		t.Fatalf("both should draw 2 DrawCircle, got hover=%d normal=%d",
			len(hoveredCircles), len(normalCircles))
	}

	if hoveredCircles[0].color == normalCircles[0].color {
		t.Error("hovered color should differ from normal color")
	}
}

func TestRadioPainter_PressedState(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.RadioPainter{}

	painter.PaintRadio(canvas, radio.PaintState{
		Label:    "Pressed",
		Selected: true,
		Pressed:  true,
		Bounds:   radioTestBounds(),
	})

	canvasNormal := &recordCanvas{}
	painter.PaintRadio(canvasNormal, radio.PaintState{
		Label:    "Normal",
		Selected: true,
		Bounds:   radioTestBounds(),
	})

	pressedCircles := canvas.methodCalls(methodDrawCircle)
	normalCircles := canvasNormal.methodCalls(methodDrawCircle)

	if len(pressedCircles) != 2 || len(normalCircles) != 2 {
		t.Fatalf("both should draw 2 DrawCircle, got pressed=%d normal=%d",
			len(pressedCircles), len(normalCircles))
	}

	if pressedCircles[0].color == normalCircles[0].color {
		t.Error("pressed color should differ from normal color")
	}
}
