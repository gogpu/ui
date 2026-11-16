package material3_test

import (
	"testing"
	"time"

	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func caretPaintState(hasError bool) *textfield.PaintState {
	return &textfield.PaintState{
		Text:        "Caret",
		DisplayText: "Caret",
		Bounds:      geometry.NewRect(0, 0, 240, 56),
		ContentRect: geometry.NewRect(16, 12, 208, 32),
		TextRect:    geometry.NewRect(16, 12, 208, 32),
		Focused:     true,
		HasError:    hasError,
		ShowCursor:  true,
		CursorRect:  geometry.NewRect(40, 16, 2, 24),
		FontSize:    16,
	}
}

// TestTextFieldCaret pins the M3 caret: a square-ended 2 dp bar in primary,
// error while the field shows an error, blinking sharply every 500 ms.
func TestTextFieldCaret(t *testing.T) {
	theme := material3.New(widget.Hex(0x6750A4))
	painter := material3.TextFieldPainter{Theme: theme}

	for _, tt := range []struct {
		name     string
		hasError bool
		want     widget.Color
	}{
		{name: "normal", want: theme.Colors.Primary},
		{name: "error", hasError: true, want: theme.Colors.Error},
	} {
		t.Run(tt.name, func(t *testing.T) {
			canvas := &recordCanvas{}
			painter.PaintTextField(canvas, caretPaintState(tt.hasError))
			lines := canvas.methodCalls(methodDrawLine)
			if len(lines) != 1 {
				t.Fatalf("DrawLine calls = %d, want 1 caret", len(lines))
			}
			if lines[0].color != tt.want {
				t.Errorf("caret color = %v, want %v", lines[0].color, tt.want)
			}
			if lines[0].strokeWidth != 2 {
				t.Errorf("caret width = %v, want 2", lines[0].strokeWidth)
			}
		})
	}

	if got := painter.TextFieldCursorWidth(); got != 2 {
		t.Errorf("TextFieldCursorWidth = %v, want 2", got)
	}
	blink := painter.TextFieldCaretBlink()
	if blink.Opacity(499*time.Millisecond) != 1 || blink.Opacity(500*time.Millisecond) != 0 {
		t.Error("caret should be visible for 500 ms, then hidden")
	}
	if got := blink.NextChange(0); got != 500*time.Millisecond {
		t.Errorf("first blink step = %v, want 500ms", got)
	}
}
