package textfield

import (
	"image"
	"testing"

	"github.com/gogpu/gpucontext"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/internal/textmetrics"
	"github.com/gogpu/ui/widget"
)

type imeMetricsCanvas struct{}

func (imeMetricsCanvas) Clear(widget.Color)                                                         {}
func (imeMetricsCanvas) DrawRect(geometry.Rect, widget.Color)                                       {}
func (imeMetricsCanvas) FillRectDirect(geometry.Rect, widget.Color)                                 {}
func (imeMetricsCanvas) StrokeRect(geometry.Rect, widget.Color, float32)                            {}
func (imeMetricsCanvas) DrawRoundRect(geometry.Rect, widget.Color, float32)                         {}
func (imeMetricsCanvas) StrokeRoundRect(geometry.Rect, widget.Color, float32, float32)              {}
func (imeMetricsCanvas) DrawCircle(geometry.Point, float32, widget.Color)                           {}
func (imeMetricsCanvas) StrokeCircle(geometry.Point, float32, widget.Color, float32)                {}
func (imeMetricsCanvas) StrokeArc(geometry.Point, float32, float64, float64, widget.Color, float32) {}
func (imeMetricsCanvas) DrawLine(geometry.Point, geometry.Point, widget.Color, float32)             {}
func (imeMetricsCanvas) DrawText(string, geometry.Rect, float32, widget.Color, bool, widget.TextAlign) {
}
func (imeMetricsCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize
}
func (imeMetricsCanvas) DrawImage(image.Image, geometry.Point)    {}
func (imeMetricsCanvas) PushClip(geometry.Rect)                   {}
func (imeMetricsCanvas) PushClipRoundRect(geometry.Rect, float32) {}
func (imeMetricsCanvas) PopClip()                                 {}
func (imeMetricsCanvas) PushTransform(geometry.Point)             {}
func (imeMetricsCanvas) PopTransform()                            {}
func (imeMetricsCanvas) TransformOffset() geometry.Point          { return geometry.Point{} }
func (imeMetricsCanvas) ScreenOriginBase() geometry.Point         { return geometry.Point{} }
func (imeMetricsCanvas) ClipBounds() geometry.Rect                { return geometry.NewRect(0, 0, 1000, 1000) }
func (imeMetricsCanvas) ReplayScene(widget.SceneCache)            {}

func TestIMEGeometryValidationBranches(t *testing.T) {
	if got, ok := byteOffsetToRune("é你", -1); ok || got != 0 {
		t.Fatalf("negative byte offset = (%d,%v)", got, ok)
	}
	if got, ok := byteOffsetToRune(string([]byte{0xff}), 0); ok || got != 0 {
		t.Fatalf("invalid UTF-8 offset = (%d,%v)", got, ok)
	}
	if got, ok := byteOffsetToRune("é", 1); ok || got != 0 {
		t.Fatalf("mid-rune byte offset = (%d,%v)", got, ok)
	}
	if got, ok := byteOffsetToRune("é", len("é")); !ok || got != 1 {
		t.Fatalf("terminal byte offset = (%d,%v)", got, ok)
	}
	if got, ok := compositionRangeToRunes("é", 0, 0); !ok || got != nil {
		t.Fatalf("empty composition range = (%v,%v)", got, ok)
	}
	if got, ok := compositionRangeToRunes("é", 1, 0); ok || got != nil {
		t.Fatalf("reversed composition range = (%v,%v)", got, ok)
	}
	if got, ok := compositionRangeToRunes("abc", 2, 1); ok || got != nil {
		t.Fatalf("ordered-byte reversed range = (%v,%v)", got, ok)
	}

	w := New()
	tm := &textmetrics.Metrics{Canvas: imeMetricsCanvas{}, FontSize: 12}
	content := geometry.NewRect(0, 0, 200, 30)
	if text, _, _, _ := w.compositionPaintGeometry(tm, content, ""); text != "" {
		t.Fatal("inactive composition produced geometry")
	}
	w.composing = true
	if text, _, _, _ := w.compositionPaintGeometry(tm, content, ""); text != "" {
		t.Fatal("empty composition produced geometry")
	}
	w.composition = gpucontext.IMEComposition{CompositionText: "é", CursorBegin: 0, CursorEnd: 0, SelectionStart: 1, SelectionEnd: 0}
	if text, _, _, _ := w.compositionPaintGeometry(tm, content, ""); text != "" {
		t.Fatal("invalid range produced geometry")
	}
	w.composition = gpucontext.IMEComposition{CompositionText: "é你", CursorBegin: 2, CursorEnd: 2, SelectionStart: 0, SelectionEnd: len("é你")}
	if text, _, _, cursor := w.compositionPaintGeometry(tm, content, ""); text == "" || cursor.IsEmpty() {
		t.Fatal("valid cursor range did not produce caret")
	}
	w.composition = gpucontext.IMEComposition{CompositionText: "é你", CursorBegin: -1, CursorEnd: -1, SelectionStart: 0, SelectionEnd: len("é你")}
	if text, _, _, cursor := w.compositionPaintGeometry(tm, content, ""); text == "" || !cursor.IsEmpty() {
		t.Fatal("hidden cursor invented caret")
	}
	w.cfg.inputType = TypePassword
	if text, _, _, _ := w.compositionPaintGeometry(tm, content, ""); text != "" {
		t.Fatal("password composition geometry leaked preedit")
	}
}

func TestIMEDeleteSurroundingRejectsInvalidDirectRequests(t *testing.T) {
	if (&Widget{}).deleteSurrounding(0, 0) {
		t.Fatal("empty text delete should be rejected")
	}
	w := New(InitialValue("abc"))
	if w.deleteSurrounding(4, 0) || w.deleteSurrounding(0, 4) {
		t.Fatal("out-of-bounds delete should be rejected")
	}
	w.SetText("é")
	w.sel.SetCursor(1)
	if w.deleteSurrounding(1, 0) {
		t.Fatal("mid-rune before delete should be rejected")
	}
	w.sel.SetCursor(0)
	if w.deleteSurrounding(0, 1) {
		t.Fatal("mid-rune after delete should be rejected")
	}
	w.SetText("abc")
	w.sel.SetCursor(1)
	if w.deleteSurrounding(-1, -1) {
		t.Fatal("reversed surrounding range should be rejected")
	}
	if w.deleteSurrounding(0, 0) {
		t.Fatal("empty surrounding range should be rejected")
	}
}

func TestIMEHandleEventGuardBranches(t *testing.T) {
	ctx := widget.NewContext()
	w := New()
	if handleIMEEvent(w, ctx, event.NewIMECompositionStart()) {
		t.Fatal("unfocused handleIMEEvent should be rejected")
	}
	w.SetFocused(true)
	w.SetEnabled(false)
	if handleIMEEvent(w, ctx, event.NewIMECompositionStart()) {
		t.Fatal("disabled handleIMEEvent should be rejected")
	}
}
