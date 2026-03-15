package uitest

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// colorEpsilon is the tolerance for float32 color comparison.
const colorEpsilon = 0.001

// AssertDrawnText checks that the canvas contains at least one DrawText call
// with the expected text string. It fails the test if no matching text is found.
func AssertDrawnText(t *testing.T, canvas *MockCanvas, expected string) {
	t.Helper()
	for _, dt := range canvas.Texts {
		if dt.Text == expected {
			return
		}
	}
	t.Errorf("expected text %q to be drawn, but it was not found in %d text calls", expected, len(canvas.Texts))
}

// AssertNoDrawnText checks that the canvas does not contain a DrawText call
// with the given text string. It fails the test if the text is found.
func AssertNoDrawnText(t *testing.T, canvas *MockCanvas, text string) {
	t.Helper()
	for _, dt := range canvas.Texts {
		if dt.Text == text {
			t.Errorf("expected text %q to NOT be drawn, but it was found", text)
			return
		}
	}
}

// AssertRectDrawn checks that a rectangle was drawn at the expected bounds.
// It compares using approximate float32 equality.
func AssertRectDrawn(t *testing.T, canvas *MockCanvas, expected geometry.Rect) {
	t.Helper()
	for _, r := range canvas.Rects {
		if rectsApproxEqual(r.Bounds, expected) {
			return
		}
	}
	t.Errorf("expected rect %v to be drawn, but it was not found in %d rect calls", expected, len(canvas.Rects))
}

// AssertInvalidated checks that the context was invalidated at least once.
func AssertInvalidated(t *testing.T, ctx *MockContext) {
	t.Helper()
	if !ctx.Invalidated {
		t.Error("expected context to be invalidated, but it was not")
	}
}

// AssertNotInvalidated checks that the context was not invalidated.
func AssertNotInvalidated(t *testing.T, ctx *MockContext) {
	t.Helper()
	if ctx.Invalidated {
		t.Error("expected context to NOT be invalidated, but it was")
	}
}

// AssertCursor checks that the context cursor matches the expected type.
func AssertCursor(t *testing.T, ctx *MockContext, expected widget.CursorType) {
	t.Helper()
	if ctx.CursorVal != expected {
		t.Errorf("cursor = %v, want %v", ctx.CursorVal, expected)
	}
}

// AssertFocused checks that the given widget currently has focus in the context.
func AssertFocused(t *testing.T, ctx *MockContext, w widget.Widget) {
	t.Helper()
	if ctx.FocusedVal != w {
		t.Errorf("expected widget to be focused, but it is not")
	}
}

// AssertNotFocused checks that the given widget does NOT have focus.
func AssertNotFocused(t *testing.T, ctx *MockContext, w widget.Widget) {
	t.Helper()
	if ctx.FocusedVal == w {
		t.Errorf("expected widget to NOT be focused, but it is")
	}
}

// AssertColorEqual checks that two colors are approximately equal
// within floating-point tolerance.
func AssertColorEqual(t *testing.T, got, want widget.Color) {
	t.Helper()
	if !colorsApproxEqual(got, want) {
		t.Errorf("color = {R:%.3f, G:%.3f, B:%.3f, A:%.3f}, want {R:%.3f, G:%.3f, B:%.3f, A:%.3f}",
			got.R, got.G, got.B, got.A, want.R, want.G, want.B, want.A)
	}
}

// colorsApproxEqual returns true if two colors are within epsilon on all components.
func colorsApproxEqual(a, b widget.Color) bool {
	return approxEqual(a.R, b.R) && approxEqual(a.G, b.G) &&
		approxEqual(a.B, b.B) && approxEqual(a.A, b.A)
}

// rectsApproxEqual returns true if two rects are within epsilon on all coordinates.
func rectsApproxEqual(a, b geometry.Rect) bool {
	return approxEqual(a.Min.X, b.Min.X) && approxEqual(a.Min.Y, b.Min.Y) &&
		approxEqual(a.Max.X, b.Max.X) && approxEqual(a.Max.Y, b.Max.Y)
}

// approxEqual returns true if two float32 values are within epsilon.
func approxEqual(a, b float32) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < colorEpsilon
}
