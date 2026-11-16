package fluent_test

import (
	"testing"
	"time"

	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/theme/fluent"
	"github.com/gogpu/ui/widget"
)

// TestTextFieldCaret pins the Fluent caret to the Windows text caret: a 1 px
// bar in the text color, not the accent, blinking every 530 ms.
func TestTextFieldCaret(t *testing.T) {
	theme := fluent.NewTheme()
	for _, tt := range []struct {
		name    string
		painter fluent.TextFieldPainter
		want    widget.Color
	}{
		{name: "theme", painter: fluent.TextFieldPainter{Theme: theme}, want: theme.Colors.OnSurface},
		{name: "default palette", painter: fluent.TextFieldPainter{}, want: widget.Hex(0x1A1A1A)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			canvas := &recordCanvas{}
			tt.painter.PaintTextField(canvas, &textfield.PaintState{
				Text:        "Caret",
				DisplayText: "Caret",
				Bounds:      geometry.NewRect(0, 0, 240, 32),
				ContentRect: geometry.NewRect(12, 8, 216, 16),
				TextRect:    geometry.NewRect(12, 8, 216, 16),
				Focused:     true,
				ShowCursor:  true,
				CursorRect:  geometry.NewRect(30, 8, 1, 16),
				FontSize:    14,
			})
			lines := canvas.methodCalls(methodDrawLine)
			if len(lines) != 1 {
				t.Fatalf("DrawLine calls = %d, want 1 caret", len(lines))
			}
			if lines[0].color != tt.want {
				t.Errorf("caret color = %v, want the text color %v", lines[0].color, tt.want)
			}
		})
	}

	painter := fluent.TextFieldPainter{}
	if got := painter.TextFieldCursorWidth(); got != 1 {
		t.Errorf("TextFieldCursorWidth = %v, want 1", got)
	}
	blink := painter.TextFieldCaretBlink()
	if blink.Opacity(529*time.Millisecond) != 1 || blink.Opacity(530*time.Millisecond) != 0 {
		t.Error("caret should be visible for 530 ms, then hidden")
	}
	if got := blink.NextChange(0); got != 530*time.Millisecond {
		t.Errorf("first blink step = %v, want 530ms", got)
	}
}
