package textmetrics

import (
	"testing"

	"github.com/gogpu/gg/scene"
	"github.com/gogpu/ui/internal/render"
	"github.com/gogpu/ui/widget"
)

// TestCaretOffsetsFallBackToMeasureText covers the canvas that does not
// implement [widget.TextCaretPositioner]: the offsets must still be the
// measured prefix widths, so no existing canvas changes behavior.
func TestCaretOffsetsFallBackToMeasureText(t *testing.T) {
	const fontSize float32 = 10

	canvas := &mockCanvas{charWidthRatio: 0.6}
	if _, ok := any(canvas).(widget.TextCaretPositioner); ok {
		t.Fatal("mockCanvas implements TextCaretPositioner: it no longer covers the fallback")
	}

	m := &Metrics{Canvas: canvas, FontSize: fontSize}
	const s = "AV Wave"

	offsets := m.CaretOffsets(s)
	if want := len([]rune(s)) + 1; len(offsets) != want {
		t.Fatalf("len(CaretOffsets(%q)) = %d, want %d", s, len(offsets), want)
	}
	for i, got := range offsets {
		want := float32(i) * fontSize * 0.6
		if absf(got-want) > caretTolerance {
			t.Errorf("CaretOffsets(%q)[%d] = %.4f, want %.4f", s, i, got, want)
		}
	}
}

// TestCaretOffsetsHandleEmptyText pins the one-entry contract, which callers
// index without a length check.
func TestCaretOffsetsHandleEmptyText(t *testing.T) {
	for name, canvas := range map[string]widget.Canvas{
		"shaping": render.NewSceneCanvas(scene.NewScene(), 100, 20),
		"measure": &mockCanvas{charWidthRatio: 0.5},
	} {
		t.Run(name, func(t *testing.T) {
			m := &Metrics{Canvas: canvas, FontSize: 14}
			offsets := m.CaretOffsets("")
			if len(offsets) != 1 || offsets[0] != 0 {
				t.Fatalf("CaretOffsets(\"\") = %v, want [0]", offsets)
			}
			if got := m.TextWidth(""); got != 0 {
				t.Errorf("TextWidth(\"\") = %v, want 0", got)
			}
		})
	}
}
