package cupertino_test

import (
	"testing"
	"time"

	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/theme/cupertino"
)

// TestTextFieldCaret pins the Cupertino caret: a 2 pt bar with rounded ends
// in the accent color, fading on the UIKit keyframes.
func TestTextFieldCaret(t *testing.T) {
	theme := cupertino.NewTheme()
	painter := cupertino.TextFieldPainter{Theme: theme}
	caret := geometry.NewRect(30, 10, 2, 20)

	canvas := &recordCanvas{}
	painter.PaintTextField(canvas, &textfield.PaintState{
		Text:          "Caret",
		DisplayText:   "Caret",
		Bounds:        geometry.NewRect(0, 0, 240, 40),
		ContentRect:   geometry.NewRect(12, 10, 216, 20),
		TextRect:      geometry.NewRect(12, 10, 216, 20),
		Focused:       true,
		ShowCursor:    true,
		CursorOpacity: 0.5,
		CursorRect:    caret,
		FontSize:      15,
	})
	if lines := canvas.methodCalls(methodDrawLine); len(lines) != 0 {
		t.Errorf("caret drawn as %d square-ended lines, want a rounded bar", len(lines))
	}
	var bars []drawCall
	for _, c := range canvas.methodCalls(methodDrawRoundRect) {
		if c.bounds == geometry.NewRect(29, 10, 2, 20) {
			bars = append(bars, c)
		}
	}
	if len(bars) != 1 {
		t.Fatalf("rounded caret bars centered on x=30 = %d, want 1", len(bars))
	}
	if bars[0].radius != 1 {
		t.Errorf("caret radius = %v, want 1 (2 pt capped at half the width)", bars[0].radius)
	}
	if want := theme.Colors.Accent.WithAlpha(theme.Colors.Accent.A * 0.5); bars[0].color != want {
		t.Errorf("caret color = %v, want the accent at the blink opacity %v", bars[0].color, want)
	}

	blink := painter.TextFieldCaretBlink()
	ms := time.Millisecond
	for _, step := range []struct {
		at      time.Duration
		opacity float32
	}{
		{0, 1}, {537 * ms, 1}, {540 * ms, 0.75}, {575 * ms, 0.5}, {620 * ms, 0.25},
		{650 * ms, 0}, {849 * ms, 0}, {890 * ms, 0.25}, {930 * ms, 0.5}, {970 * ms, 0.75},
		{time.Second, 1},
	} {
		if got := blink.Opacity(step.at); got != step.opacity {
			t.Errorf("Opacity(%v) = %v, want %v", step.at, got, step.opacity)
		}
	}
	if got := blink.NextChange(0); got != 537500*time.Microsecond {
		t.Errorf("first blink step = %v, want 537.5ms (the 500ms keyframe keeps full opacity)", got)
	}
}
