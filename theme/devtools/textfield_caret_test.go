package devtools_test

import (
	"testing"
	"time"

	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/theme/devtools"
	"github.com/gogpu/ui/widget"
)

// TestTextFieldCaret pins the DevTools caret to the IntelliJ text field
// caret: a thin bar in the text color (TextField.caretForeground), not the
// accent, blinking every 500 ms.
func TestTextFieldCaret(t *testing.T) {
	theme := devtools.NewDarkTheme()
	for _, tt := range []struct {
		name    string
		painter devtools.TextFieldPainter
		want    widget.Color
	}{
		{name: "theme", painter: devtools.TextFieldPainter{Theme: theme}, want: theme.Colors.OnSurface},
		{name: "default palette", painter: devtools.TextFieldPainter{}, want: widget.Hex(0xDFE1E5)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			canvas := &recordCanvas{}
			tt.painter.PaintTextField(canvas, &textfield.PaintState{
				Text:        "Caret",
				DisplayText: "Caret",
				Bounds:      testBounds(),
				ContentRect: geometry.NewRect(8, 6, 184, 18),
				TextRect:    geometry.NewRect(8, 6, 184, 18),
				Focused:     true,
				ShowCursor:  true,
				CursorRect:  geometry.NewRect(20, 6, 1, 18),
				FontSize:    13,
			})
			lines := canvas.methodCalls(methodDrawLine)
			if len(lines) != 1 {
				t.Fatalf("DrawLine calls = %d, want 1 caret", len(lines))
			}
			if lines[0].color != tt.want {
				t.Errorf("caret color = %v, want the text color %v", lines[0].color, tt.want)
			}
			if lines[0].strokeWidth != 1 {
				t.Errorf("caret width = %v, want 1", lines[0].strokeWidth)
			}
		})
	}

	blink := devtools.TextFieldPainter{}.TextFieldCaretBlink()
	if blink.Opacity(499*time.Millisecond) != 1 || blink.Opacity(500*time.Millisecond) != 0 {
		t.Error("caret should be visible for 500 ms, then hidden")
	}
}
