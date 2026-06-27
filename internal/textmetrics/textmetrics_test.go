package textmetrics

import (
	"image"
	"testing"

	"github.com/gogpu/gg/scene"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// mockCanvas provides MeasureText using a fixed character width ratio.
type mockCanvas struct {
	charWidthRatio float32
}

func (c *mockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * c.charWidthRatio
}

func (c *mockCanvas) Clear(_ widget.Color)                                                  {}
func (c *mockCanvas) DrawRect(_ geometry.Rect, _ widget.Color)                              {}
func (c *mockCanvas) FillRectDirect(_ geometry.Rect, _ widget.Color)                        {}
func (c *mockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32)                 {}
func (c *mockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32)              {}
func (c *mockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {}
func (c *mockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)                {}
func (c *mockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32)   {}

func (c *mockCanvas) StrokeArc(_ geometry.Point, _ float32, _, _ float64, _ widget.Color, _ float32) {
}

func (c *mockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32) {}
func (c *mockCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
}
func (c *mockCanvas) DrawImage(_ image.Image, _ geometry.Point)    {}
func (c *mockCanvas) PushClip(_ geometry.Rect)                     {}
func (c *mockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *mockCanvas) PopClip()                                     {}
func (c *mockCanvas) PushTransform(_ geometry.Point)               {}
func (c *mockCanvas) PopTransform()                                {}
func (c *mockCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }
func (c *mockCanvas) ScreenOriginBase() geometry.Point             { return geometry.Point{} }
func (c *mockCanvas) ClipBounds() geometry.Rect                    { return geometry.NewRect(0, 0, 10000, 10000) }
func (c *mockCanvas) ReplayScene(_ *scene.Scene)                   {}

func newMetrics() *Metrics {
	return &Metrics{
		Canvas:   &mockCanvas{charWidthRatio: 0.5},
		FontSize: 14,
	}
}

func testContentRect() geometry.Rect {
	return geometry.NewRect(10, 5, 200, 30)
}

// --- CursorX Tests ---

func TestCursorX_AtStart(t *testing.T) {
	m := newMetrics()
	cr := testContentRect()

	x := m.CursorX(cr, "Hello", 0)
	if x != cr.Min.X {
		t.Errorf("CursorX(0) = %v, want %v", x, cr.Min.X)
	}
}

func TestCursorX_AtEnd(t *testing.T) {
	m := newMetrics()
	cr := testContentRect()

	x := m.CursorX(cr, "Hi", 2)
	// "Hi" = 2 runes * 14 * 0.5 = 14px
	want := cr.Min.X + 14
	if x != want {
		t.Errorf("CursorX(2) = %v, want %v", x, want)
	}
}

func TestCursorX_BeyondTextLength(t *testing.T) {
	m := newMetrics()
	cr := testContentRect()

	// Position beyond text length should clamp to text length.
	x := m.CursorX(cr, "Hi", 10)
	want := cr.Min.X + 14 // same as end of "Hi"
	if x != want {
		t.Errorf("CursorX(10) = %v, want %v (clamped to text end)", x, want)
	}
}

func TestCursorX_NegativePosition(t *testing.T) {
	m := newMetrics()
	cr := testContentRect()

	x := m.CursorX(cr, "Hello", -1)
	if x != cr.Min.X {
		t.Errorf("CursorX(-1) = %v, want %v (clamped to start)", x, cr.Min.X)
	}
}

func TestCursorX_EmptyText(t *testing.T) {
	m := newMetrics()
	cr := testContentRect()

	x := m.CursorX(cr, "", 0)
	if x != cr.Min.X {
		t.Errorf("CursorX empty = %v, want %v", x, cr.Min.X)
	}
}

// --- RuneIndexFromX Tests ---

func TestRuneIndexFromX_AtStart(t *testing.T) {
	m := newMetrics()
	cr := testContentRect()

	idx := m.RuneIndexFromX(cr, "Hello", cr.Min.X)
	if idx != 0 {
		t.Errorf("RuneIndexFromX(start) = %d, want 0", idx)
	}
}

func TestRuneIndexFromX_BeforeStart(t *testing.T) {
	m := newMetrics()
	cr := testContentRect()

	idx := m.RuneIndexFromX(cr, "Hello", cr.Min.X-10)
	if idx != 0 {
		t.Errorf("RuneIndexFromX(before start) = %d, want 0", idx)
	}
}

func TestRuneIndexFromX_AfterEnd(t *testing.T) {
	m := newMetrics()
	cr := testContentRect()

	// Well past the end of text.
	idx := m.RuneIndexFromX(cr, "Hi", cr.Min.X+1000)
	if idx != 2 {
		t.Errorf("RuneIndexFromX(past end) = %d, want 2", idx)
	}
}

func TestRuneIndexFromX_MiddleOfChar(t *testing.T) {
	m := newMetrics()
	cr := testContentRect()

	// Each char = 14*0.5 = 7px. Halfway through first char = 3.5px.
	// At 3px -> closer to start (0) than end of char (7) -> index 0.
	idx := m.RuneIndexFromX(cr, "Hello", cr.Min.X+3)
	if idx != 0 {
		t.Errorf("RuneIndexFromX(3px in) = %d, want 0", idx)
	}

	// At 5px -> closer to end of char (7) than start (0) -> index 1.
	idx = m.RuneIndexFromX(cr, "Hello", cr.Min.X+5)
	if idx != 1 {
		t.Errorf("RuneIndexFromX(5px in) = %d, want 1", idx)
	}
}

func TestRuneIndexFromX_Unicode(t *testing.T) {
	m := newMetrics()
	cr := testContentRect()

	// Unicode bullets (6 chars), each 7px wide.
	text := string([]rune{'\u2022', '\u2022', '\u2022', '\u2022', '\u2022', '\u2022'})
	idx := m.RuneIndexFromX(cr, text, cr.Min.X+21) // 21px = 3 chars
	if idx != 3 {
		t.Errorf("RuneIndexFromX(unicode 21px) = %d, want 3", idx)
	}
}

// --- CursorRect Tests ---

func TestCursorRect_Basic(t *testing.T) {
	m := newMetrics()
	cr := testContentRect()

	rect := m.CursorRect(cr, "Hello", 2, 2.0)

	// X should be at position 2 = 2*7 = 14px from content start.
	wantX := cr.Min.X + 14
	if rect.Min.X != wantX {
		t.Errorf("CursorRect.Min.X = %v, want %v", rect.Min.X, wantX)
	}

	// Width should be cursor width.
	if rect.Width() != 2.0 {
		t.Errorf("CursorRect.Width = %v, want 2.0", rect.Width())
	}

	// Height should be based on lineHeight (fontSize * 1.4) minus
	// 2*caretHeightOffset (Flutter _kCaretHeightOffset = 2.0).
	lineHeight := m.FontSize * 1.4
	wantHeight := lineHeight - 2*caretHeightOffset
	gotHeight := rect.Height()
	if gotHeight < wantHeight-0.1 || gotHeight > wantHeight+0.1 {
		t.Errorf("CursorRect.Height = %v, want ~%v", gotHeight, wantHeight)
	}
}

func TestCursorRect_VerticallyCentered(t *testing.T) {
	m := newMetrics()
	cr := testContentRect()

	rect := m.CursorRect(cr, "Hi", 0, 1.0)

	centerY := (cr.Min.Y + cr.Max.Y) / 2
	rectCenterY := (rect.Min.Y + rect.Max.Y) / 2
	if rectCenterY < centerY-0.5 || rectCenterY > centerY+0.5 {
		t.Errorf("CursorRect center Y = %v, want ~%v (vertically centered)", rectCenterY, centerY)
	}
}

// --- SelectionRect Tests ---

func TestSelectionRect_Basic(t *testing.T) {
	m := newMetrics()
	cr := testContentRect()

	rect := m.SelectionRect(cr, "Hello", 1, 3)

	// Start X: 1 char = 7px, End X: 3 chars = 21px.
	wantX1 := cr.Min.X + 7
	wantX2 := cr.Min.X + 21

	if rect.Min.X != wantX1 {
		t.Errorf("SelectionRect.Min.X = %v, want %v", rect.Min.X, wantX1)
	}
	if rect.Max.X != wantX2 {
		t.Errorf("SelectionRect.Max.X = %v, want %v", rect.Max.X, wantX2)
	}

	// Y should span full content rect height.
	if rect.Min.Y != cr.Min.Y {
		t.Errorf("SelectionRect.Min.Y = %v, want %v", rect.Min.Y, cr.Min.Y)
	}
	if rect.Max.Y != cr.Max.Y {
		t.Errorf("SelectionRect.Max.Y = %v, want %v", rect.Max.Y, cr.Max.Y)
	}
}

func TestSelectionRect_ReversedRange(t *testing.T) {
	m := newMetrics()
	cr := testContentRect()

	// Start > end should be handled (reversed).
	rect := m.SelectionRect(cr, "Hello", 3, 1)

	wantX1 := cr.Min.X + 7 // min(1,3) = 1 -> 7px
	wantX2 := cr.Min.X + 21

	if rect.Min.X != wantX1 {
		t.Errorf("SelectionRect(reversed).Min.X = %v, want %v", rect.Min.X, wantX1)
	}
	if rect.Max.X != wantX2 {
		t.Errorf("SelectionRect(reversed).Max.X = %v, want %v", rect.Max.X, wantX2)
	}
}

func TestSelectionRect_FullText(t *testing.T) {
	m := newMetrics()
	cr := testContentRect()

	rect := m.SelectionRect(cr, "Hi", 0, 2)

	if rect.Min.X != cr.Min.X {
		t.Errorf("SelectionRect(all).Min.X = %v, want %v", rect.Min.X, cr.Min.X)
	}
	wantX2 := cr.Min.X + 14 // 2 chars * 7px
	if rect.Max.X != wantX2 {
		t.Errorf("SelectionRect(all).Max.X = %v, want %v", rect.Max.X, wantX2)
	}
}

func TestSelectionRect_ClampedToContentRect(t *testing.T) {
	m := newMetrics()
	// Small content rect to test clamping.
	cr := geometry.NewRect(10, 5, 20, 30)

	// 10 chars * 7px = 70px > 20px width.
	rect := m.SelectionRect(cr, "HelloWorld", 0, 10)

	if rect.Max.X > cr.Max.X {
		t.Errorf("SelectionRect.Max.X = %v, should be clamped to %v", rect.Max.X, cr.Max.X)
	}
}
