package material3

import (
	"testing"

	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/geometry"
)

func TestPaintTextFieldComposition(t *testing.T) {
	canvas := &pbMockCanvas{}
	(TextFieldPainter{}).PaintTextField(canvas, &textfield.PaintState{
		DisplayText: "committed", TextRect: geometry.NewRect(0, 0, 100, 40), Bounds: geometry.NewRect(0, 0, 100, 40),
		Focused: true, ShowComposition: true,
		CompositionText: "preedit", CompositionTextRect: geometry.NewRect(10, 10, 30, 20),
		CompositionSelectionRect: geometry.NewRect(10, 10, 10, 20),
		CompositionCursorRect:    geometry.NewRect(20, 10, 1, 20),
	})
	if canvas.drawCount == 0 {
		t.Fatal("composition painter produced no draw calls")
	}
}
