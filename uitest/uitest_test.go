package uitest_test

import (
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"
)

// --- MockCanvas Tests ---

func TestMockCanvas_ImplementsInterface(t *testing.T) {
	var _ widget.Canvas = (*uitest.MockCanvas)(nil)
}

func TestMockCanvas_RecordsDrawRect(t *testing.T) {
	c := &uitest.MockCanvas{}
	r := geometry.NewRect(10, 20, 100, 50)
	color := widget.ColorRed

	c.DrawRect(r, color)

	if len(c.Rects) != 1 {
		t.Fatalf("Rects count = %d, want 1", len(c.Rects))
	}
	if c.Rects[0].Bounds != r {
		t.Errorf("Rects[0].Bounds = %v, want %v", c.Rects[0].Bounds, r)
	}
	if c.Rects[0].Color != color {
		t.Errorf("Rects[0].Color = %v, want %v", c.Rects[0].Color, color)
	}
}

func TestMockCanvas_RecordsStrokeRect(t *testing.T) {
	c := &uitest.MockCanvas{}
	r := geometry.NewRect(0, 0, 50, 50)

	c.StrokeRect(r, widget.ColorBlack, 2.0)

	if len(c.StrokeRects) != 1 {
		t.Fatalf("StrokeRects count = %d, want 1", len(c.StrokeRects))
	}
	if c.StrokeRects[0].StrokeWidth != 2.0 {
		t.Errorf("StrokeWidth = %f, want 2.0", c.StrokeRects[0].StrokeWidth)
	}
}

func TestMockCanvas_RecordsDrawRoundRect(t *testing.T) {
	c := &uitest.MockCanvas{}
	r := geometry.NewRect(0, 0, 100, 40)

	c.DrawRoundRect(r, widget.ColorBlue, 8.0)

	if len(c.RoundRects) != 1 {
		t.Fatalf("RoundRects count = %d, want 1", len(c.RoundRects))
	}
	if c.RoundRects[0].Radius != 8.0 {
		t.Errorf("Radius = %f, want 8.0", c.RoundRects[0].Radius)
	}
}

func TestMockCanvas_RecordsStrokeRoundRect(t *testing.T) {
	c := &uitest.MockCanvas{}
	r := geometry.NewRect(0, 0, 100, 40)

	c.StrokeRoundRect(r, widget.ColorGreen, 4.0, 1.5)

	if len(c.StrokeRoundRects) != 1 {
		t.Fatalf("StrokeRoundRects count = %d, want 1", len(c.StrokeRoundRects))
	}
	if c.StrokeRoundRects[0].Radius != 4.0 {
		t.Errorf("Radius = %f, want 4.0", c.StrokeRoundRects[0].Radius)
	}
	if c.StrokeRoundRects[0].StrokeWidth != 1.5 {
		t.Errorf("StrokeWidth = %f, want 1.5", c.StrokeRoundRects[0].StrokeWidth)
	}
}

func TestMockCanvas_RecordsDrawCircle(t *testing.T) {
	c := &uitest.MockCanvas{}
	center := geometry.Pt(50, 50)

	c.DrawCircle(center, 25.0, widget.ColorYellow)

	if len(c.Circles) != 1 {
		t.Fatalf("Circles count = %d, want 1", len(c.Circles))
	}
	if c.Circles[0].Center != center {
		t.Errorf("Center = %v, want %v", c.Circles[0].Center, center)
	}
	if c.Circles[0].Radius != 25.0 {
		t.Errorf("Radius = %f, want 25.0", c.Circles[0].Radius)
	}
}

func TestMockCanvas_RecordsStrokeCircle(t *testing.T) {
	c := &uitest.MockCanvas{}
	center := geometry.Pt(50, 50)

	c.StrokeCircle(center, 25.0, widget.ColorCyan, 3.0)

	if len(c.StrokeCircles) != 1 {
		t.Fatalf("StrokeCircles count = %d, want 1", len(c.StrokeCircles))
	}
	if c.StrokeCircles[0].StrokeWidth != 3.0 {
		t.Errorf("StrokeWidth = %f, want 3.0", c.StrokeCircles[0].StrokeWidth)
	}
}

func TestMockCanvas_RecordsDrawLine(t *testing.T) {
	c := &uitest.MockCanvas{}
	from := geometry.Pt(0, 0)
	to := geometry.Pt(100, 100)

	c.DrawLine(from, to, widget.ColorWhite, 1.0)

	if len(c.Lines) != 1 {
		t.Fatalf("Lines count = %d, want 1", len(c.Lines))
	}
	if c.Lines[0].From != from {
		t.Errorf("From = %v, want %v", c.Lines[0].From, from)
	}
	if c.Lines[0].To != to {
		t.Errorf("To = %v, want %v", c.Lines[0].To, to)
	}
}

func TestMockCanvas_RecordsDrawText(t *testing.T) {
	c := &uitest.MockCanvas{}
	bounds := geometry.NewRect(10, 10, 200, 30)

	c.DrawText("Hello", bounds, 14.0, widget.ColorBlack, true, widget.TextAlignCenter)

	if len(c.Texts) != 1 {
		t.Fatalf("Texts count = %d, want 1", len(c.Texts))
	}
	dt := c.Texts[0]
	if dt.Text != "Hello" {
		t.Errorf("Text = %q, want %q", dt.Text, "Hello")
	}
	if dt.FontSize != 14.0 {
		t.Errorf("FontSize = %f, want 14.0", dt.FontSize)
	}
	if !dt.Bold {
		t.Error("Bold = false, want true")
	}
	if dt.Align != widget.TextAlignCenter {
		t.Errorf("Align = %v, want %v", dt.Align, widget.TextAlignCenter)
	}
}

func TestMockCanvas_RecordsDrawImage(t *testing.T) {
	c := &uitest.MockCanvas{}
	at := geometry.Pt(5, 10)

	c.DrawImage(nil, at)

	if len(c.Images) != 1 {
		t.Fatalf("Images count = %d, want 1", len(c.Images))
	}
	if c.Images[0].At != at {
		t.Errorf("At = %v, want %v", c.Images[0].At, at)
	}
}

func TestMockCanvas_RecordsClear(t *testing.T) {
	c := &uitest.MockCanvas{}

	c.Clear(widget.ColorWhite)

	if len(c.Clears) != 1 {
		t.Fatalf("Clears count = %d, want 1", len(c.Clears))
	}
	if c.Clears[0] != widget.ColorWhite {
		t.Errorf("Clear color = %v, want %v", c.Clears[0], widget.ColorWhite)
	}
}

func TestMockCanvas_RecordsClipAndTransform(t *testing.T) {
	c := &uitest.MockCanvas{}
	clip := geometry.NewRect(0, 0, 200, 200)
	offset := geometry.Pt(10, 20)

	c.PushClip(clip)
	c.PushClipRoundRect(clip, 8.0)
	c.PushTransform(offset)
	c.PopClip()
	c.PopClip()
	c.PopTransform()

	if len(c.Clips) != 1 {
		t.Errorf("Clips count = %d, want 1", len(c.Clips))
	}
	if len(c.ClipRoundRects) != 1 {
		t.Errorf("ClipRoundRects count = %d, want 1", len(c.ClipRoundRects))
	}
	if c.ClipRoundRects[0].Radius != 8.0 {
		t.Errorf("ClipRoundRect Radius = %f, want 8.0", c.ClipRoundRects[0].Radius)
	}
	if len(c.Transforms) != 1 {
		t.Errorf("Transforms count = %d, want 1", len(c.Transforms))
	}
	if c.PopClipCount != 2 {
		t.Errorf("PopClipCount = %d, want 2", c.PopClipCount)
	}
	if c.PopTransformCount != 1 {
		t.Errorf("PopTransformCount = %d, want 1", c.PopTransformCount)
	}
}

func TestMockCanvas_Reset(t *testing.T) {
	c := &uitest.MockCanvas{}
	c.DrawRect(geometry.NewRect(0, 0, 10, 10), widget.ColorRed)
	c.DrawText("test", geometry.NewRect(0, 0, 50, 20), 12, widget.ColorBlack, false, widget.TextAlignLeft)
	c.PushClip(geometry.NewRect(0, 0, 100, 100))
	c.PopClip()
	c.PushTransform(geometry.Pt(5, 5))
	c.PopTransform()
	c.Clear(widget.ColorWhite)

	c.Reset()

	if len(c.Rects) != 0 {
		t.Errorf("Rects should be empty after reset, got %d", len(c.Rects))
	}
	if len(c.Texts) != 0 {
		t.Errorf("Texts should be empty after reset, got %d", len(c.Texts))
	}
	if len(c.Clips) != 0 {
		t.Errorf("Clips should be empty after reset, got %d", len(c.Clips))
	}
	if len(c.Transforms) != 0 {
		t.Errorf("Transforms should be empty after reset, got %d", len(c.Transforms))
	}
	if len(c.Clears) != 0 {
		t.Errorf("Clears should be empty after reset, got %d", len(c.Clears))
	}
	if c.PopClipCount != 0 {
		t.Errorf("PopClipCount should be 0 after reset, got %d", c.PopClipCount)
	}
	if c.PopTransformCount != 0 {
		t.Errorf("PopTransformCount should be 0 after reset, got %d", c.PopTransformCount)
	}
}

func TestMockCanvas_TotalDrawCalls(t *testing.T) {
	c := &uitest.MockCanvas{}
	c.Clear(widget.ColorWhite)
	c.DrawRect(geometry.NewRect(0, 0, 10, 10), widget.ColorRed)
	c.DrawText("hi", geometry.NewRect(0, 0, 50, 20), 12, widget.ColorBlack, false, widget.TextAlignLeft)
	c.PushClip(geometry.NewRect(0, 0, 100, 100)) // Not counted as draw call.

	if c.TotalDrawCalls() != 3 {
		t.Errorf("TotalDrawCalls = %d, want 3", c.TotalDrawCalls())
	}
}

// --- MockContext Tests ---

func TestMockContext_ImplementsInterface(t *testing.T) {
	var _ widget.Context = (*uitest.MockContext)(nil)
}

func TestNewMockContext_Defaults(t *testing.T) {
	ctx := uitest.NewMockContext()

	if ctx.Scale() != 1.0 {
		t.Errorf("Scale = %f, want 1.0", ctx.Scale())
	}
	if ctx.DeltaTime() == 0 {
		t.Error("DeltaTime should not be zero")
	}
	if ctx.WindowSize().Width != 800 || ctx.WindowSize().Height != 600 {
		t.Errorf("WindowSize = %v, want 800x600", ctx.WindowSize())
	}
	if ctx.ThemeProvider() != nil {
		t.Error("ThemeProvider should be nil by default")
	}
	if ctx.OverlayManager() != nil {
		t.Error("OverlayManager should be nil by default")
	}
	if ctx.Scheduler() != nil {
		t.Error("Scheduler should be nil by default")
	}
	if ctx.FocusedWidget() != nil {
		t.Error("FocusedWidget should be nil by default")
	}
	if ctx.Cursor() != widget.CursorDefault {
		t.Errorf("Cursor = %v, want Default", ctx.Cursor())
	}
}

func TestMockContext_Focus(t *testing.T) {
	ctx := uitest.NewMockContext()
	w := &testWidget{}

	ctx.RequestFocus(w)
	if !ctx.IsFocused(w) {
		t.Error("widget should be focused after RequestFocus")
	}
	if ctx.FocusedWidget() != w {
		t.Error("FocusedWidget should return the focused widget")
	}

	ctx.ReleaseFocus(w)
	if ctx.IsFocused(w) {
		t.Error("widget should not be focused after ReleaseFocus")
	}
	if ctx.FocusedWidget() != nil {
		t.Error("FocusedWidget should be nil after release")
	}
}

func TestMockContext_FocusSwitch(t *testing.T) {
	ctx := uitest.NewMockContext()
	w1 := &testWidget{}
	w2 := &testWidget{}

	ctx.RequestFocus(w1)
	ctx.RequestFocus(w2)

	if ctx.IsFocused(w1) {
		t.Error("w1 should not be focused after w2 gets focus")
	}
	if !ctx.IsFocused(w2) {
		t.Error("w2 should be focused")
	}
	if w1.IsFocused() {
		t.Error("w1.IsFocused() should return false")
	}
}

func TestMockContext_ReleaseFocus_WrongWidget(t *testing.T) {
	ctx := uitest.NewMockContext()
	w1 := &testWidget{}
	w2 := &testWidget{}

	ctx.RequestFocus(w1)
	ctx.ReleaseFocus(w2) // w2 doesn't have focus, should be no-op

	if !ctx.IsFocused(w1) {
		t.Error("w1 should still be focused")
	}
}

func TestMockContext_Invalidation(t *testing.T) {
	ctx := uitest.NewMockContext()

	if ctx.Invalidated {
		t.Error("should not be invalidated initially")
	}

	ctx.Invalidate()
	if !ctx.Invalidated {
		t.Error("should be invalidated after Invalidate()")
	}
	if ctx.InvalidateCount != 1 {
		t.Errorf("InvalidateCount = %d, want 1", ctx.InvalidateCount)
	}

	ctx.Invalidate()
	if ctx.InvalidateCount != 2 {
		t.Errorf("InvalidateCount = %d, want 2", ctx.InvalidateCount)
	}
}

func TestMockContext_InvalidateRect(t *testing.T) {
	ctx := uitest.NewMockContext()
	r := geometry.NewRect(10, 20, 100, 50)

	ctx.InvalidateRect(r)

	if len(ctx.InvalidatedRects) != 1 {
		t.Fatalf("InvalidatedRects count = %d, want 1", len(ctx.InvalidatedRects))
	}
	if ctx.InvalidatedRects[0] != r {
		t.Errorf("InvalidatedRects[0] = %v, want %v", ctx.InvalidatedRects[0], r)
	}
}

func TestMockContext_SetCursor(t *testing.T) {
	ctx := uitest.NewMockContext()

	ctx.SetCursor(widget.CursorPointer)
	if ctx.Cursor() != widget.CursorPointer {
		t.Errorf("Cursor = %v, want Pointer", ctx.Cursor())
	}
}

func TestMockContext_Reset(t *testing.T) {
	ctx := uitest.NewMockContext()
	ctx.Invalidate()
	ctx.InvalidateRect(geometry.NewRect(0, 0, 10, 10))
	ctx.SetCursor(widget.CursorText)

	ctx.Reset()

	if ctx.Invalidated {
		t.Error("Invalidated should be false after reset")
	}
	if ctx.InvalidateCount != 0 {
		t.Errorf("InvalidateCount = %d, want 0", ctx.InvalidateCount)
	}
	if len(ctx.InvalidatedRects) != 0 {
		t.Errorf("InvalidatedRects should be empty, got %d", len(ctx.InvalidatedRects))
	}
	if ctx.Cursor() != widget.CursorDefault {
		t.Errorf("Cursor = %v, want Default", ctx.Cursor())
	}
}

// --- Event Factory Tests ---

func TestClick(t *testing.T) {
	e := uitest.Click(50, 25)

	if e.MouseType != event.MousePress {
		t.Errorf("MouseType = %v, want Press", e.MouseType)
	}
	if e.Button != event.ButtonLeft {
		t.Errorf("Button = %v, want Left", e.Button)
	}
	if e.Position.X != 50 || e.Position.Y != 25 {
		t.Errorf("Position = %v, want (50, 25)", e.Position)
	}
}

func TestRelease(t *testing.T) {
	e := uitest.Release(50, 25)

	if e.MouseType != event.MouseRelease {
		t.Errorf("MouseType = %v, want Release", e.MouseType)
	}
	if e.Button != event.ButtonLeft {
		t.Errorf("Button = %v, want Left", e.Button)
	}
}

func TestDoubleClick(t *testing.T) {
	e := uitest.DoubleClick(30, 40)

	if e.MouseType != event.MouseDoubleClick {
		t.Errorf("MouseType = %v, want DoubleClick", e.MouseType)
	}
	if e.ClickCount != 2 {
		t.Errorf("ClickCount = %d, want 2", e.ClickCount)
	}
}

func TestRightClick(t *testing.T) {
	e := uitest.RightClick(10, 20)

	if e.MouseType != event.MousePress {
		t.Errorf("MouseType = %v, want Press", e.MouseType)
	}
	if e.Button != event.ButtonRight {
		t.Errorf("Button = %v, want Right", e.Button)
	}
}

func TestMouseMove(t *testing.T) {
	e := uitest.MouseMove(100, 200)

	if e.MouseType != event.MouseMove {
		t.Errorf("MouseType = %v, want Move", e.MouseType)
	}
	if e.Button != event.ButtonNone {
		t.Errorf("Button = %v, want None", e.Button)
	}
}

func TestMouseEnterLeave(t *testing.T) {
	enter := uitest.MouseEnter(10, 20)
	leave := uitest.MouseLeave(10, 20)

	if enter.MouseType != event.MouseEnter {
		t.Errorf("Enter MouseType = %v, want Enter", enter.MouseType)
	}
	if leave.MouseType != event.MouseLeave {
		t.Errorf("Leave MouseType = %v, want Leave", leave.MouseType)
	}
}

func TestMouseDrag(t *testing.T) {
	e := uitest.MouseDrag(75, 80)

	if e.MouseType != event.MouseDrag {
		t.Errorf("MouseType = %v, want Drag", e.MouseType)
	}
	if e.Button != event.ButtonLeft {
		t.Errorf("Button = %v, want Left", e.Button)
	}
}

func TestKeyPress(t *testing.T) {
	e := uitest.KeyPress(event.KeyEnter, event.ModNone)

	if e.KeyType != event.KeyPress {
		t.Errorf("KeyType = %v, want Press", e.KeyType)
	}
	if e.Key != event.KeyEnter {
		t.Errorf("Key = %v, want Enter", e.Key)
	}
}

func TestKeyRelease(t *testing.T) {
	e := uitest.KeyRelease(event.KeyEscape, event.ModNone)

	if e.KeyType != event.KeyRelease {
		t.Errorf("KeyType = %v, want Release", e.KeyType)
	}
}

func TestKeyType(t *testing.T) {
	e := uitest.KeyType(event.KeyA, 'a', event.ModNone)

	if e.Key != event.KeyA {
		t.Errorf("Key = %v, want A", e.Key)
	}
	if e.Rune != 'a' {
		t.Errorf("Rune = %c, want 'a'", e.Rune)
	}
}

func TestKeyPressWithModifiers(t *testing.T) {
	e := uitest.KeyPress(event.KeyC, event.ModCtrl)

	if !e.Modifiers().IsCtrl() {
		t.Error("expected Ctrl modifier")
	}
}

func TestWheelScroll(t *testing.T) {
	e := uitest.WheelScroll(50, 50, 3.0)

	if e.DeltaY() != 3.0 {
		t.Errorf("DeltaY = %f, want 3.0", e.DeltaY())
	}
	if e.DeltaX() != 0 {
		t.Errorf("DeltaX = %f, want 0", e.DeltaX())
	}
	if e.Position.X != 50 || e.Position.Y != 50 {
		t.Errorf("Position = %v, want (50, 50)", e.Position)
	}
}

func TestWheelScrollH(t *testing.T) {
	e := uitest.WheelScrollH(10, 20, 1.5, -2.0)

	if e.DeltaX() != 1.5 {
		t.Errorf("DeltaX = %f, want 1.5", e.DeltaX())
	}
	if e.DeltaY() != -2.0 {
		t.Errorf("DeltaY = %f, want -2.0", e.DeltaY())
	}
}

func TestFocusEvents(t *testing.T) {
	gained := uitest.FocusGained()
	lost := uitest.FocusLost()

	if !gained.IsGained() {
		t.Error("FocusGained should return IsGained=true")
	}
	if !lost.IsLost() {
		t.Error("FocusLost should return IsLost=true")
	}
}

// --- Widget Helper Tests ---

func TestLayoutWidget(t *testing.T) {
	w := &testWidget{preferredSize: geometry.Sz(80, 30)}

	size := uitest.LayoutWidget(w, 200, 100)

	if size.Width != 80 || size.Height != 30 {
		t.Errorf("size = %v, want (80, 30)", size)
	}
}

func TestLayoutWidgetTight(t *testing.T) {
	w := &testWidget{preferredSize: geometry.Sz(80, 30)}

	size := uitest.LayoutWidgetTight(w, 200, 100)

	// Tight constraints force the size to 200x100.
	if size.Width != 200 || size.Height != 100 {
		t.Errorf("size = %v, want (200, 100)", size)
	}
}

func TestDrawWidget(t *testing.T) {
	w := &testWidget{drawText: "drawn"}
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))

	canvas := uitest.DrawWidget(w)

	if len(canvas.Texts) != 1 {
		t.Fatalf("Texts count = %d, want 1", len(canvas.Texts))
	}
	if canvas.Texts[0].Text != "drawn" {
		t.Errorf("Text = %q, want %q", canvas.Texts[0].Text, "drawn")
	}
}

func TestDrawWidgetWithContext(t *testing.T) {
	w := &testWidget{drawText: "ctx-draw"}
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := uitest.NewMockContext()
	ctx.ScaleVal = 2.0

	canvas := uitest.DrawWidgetWithContext(w, ctx)

	if len(canvas.Texts) != 1 {
		t.Fatalf("Texts count = %d, want 1", len(canvas.Texts))
	}
}

func TestSimulateClick(t *testing.T) {
	w := &testWidget{consumeEvents: true}
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))

	consumed := uitest.SimulateClick(w, 50, 20)

	if !consumed {
		t.Error("SimulateClick should return true when widget consumes events")
	}
	if w.eventCount < 2 {
		t.Errorf("eventCount = %d, want >= 2 (press + release)", w.eventCount)
	}
}

func TestSimulateClick_NotConsumed(t *testing.T) {
	w := &testWidget{consumeEvents: false}
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))

	consumed := uitest.SimulateClick(w, 50, 20)

	if consumed {
		t.Error("SimulateClick should return false when widget does not consume events")
	}
}

func TestSimulateClickWithContext(t *testing.T) {
	w := &testWidget{consumeEvents: true}
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := uitest.NewMockContext()

	consumed := uitest.SimulateClickWithContext(w, ctx, 50, 20)

	if !consumed {
		t.Error("should return true")
	}
}

func TestSimulateKeyPress(t *testing.T) {
	w := &testWidget{consumeEvents: true}

	consumed := uitest.SimulateKeyPress(w, event.KeyEnter)

	if !consumed {
		t.Error("SimulateKeyPress should return true when widget consumes events")
	}
}

func TestSimulateKeyPressWithMods(t *testing.T) {
	w := &testWidget{consumeEvents: true}

	consumed := uitest.SimulateKeyPressWithMods(w, event.KeyC, event.ModCtrl)

	if !consumed {
		t.Error("should return true")
	}
}

// --- Assert Tests ---

func TestAssertDrawnText_Pass(t *testing.T) {
	c := &uitest.MockCanvas{}
	c.DrawText("Hello", geometry.NewRect(0, 0, 100, 20), 14, widget.ColorBlack, false, widget.TextAlignLeft)

	// Should not fail.
	uitest.AssertDrawnText(t, c, "Hello")
}

func TestAssertNoDrawnText_Pass(t *testing.T) {
	c := &uitest.MockCanvas{}
	c.DrawText("Hello", geometry.NewRect(0, 0, 100, 20), 14, widget.ColorBlack, false, widget.TextAlignLeft)

	// Should not fail: "World" was not drawn.
	uitest.AssertNoDrawnText(t, c, "World")
}

func TestAssertRectDrawn_Pass(t *testing.T) {
	c := &uitest.MockCanvas{}
	r := geometry.NewRect(10, 20, 100, 50)
	c.DrawRect(r, widget.ColorRed)

	uitest.AssertRectDrawn(t, c, r)
}

func TestAssertInvalidated_Pass(t *testing.T) {
	ctx := uitest.NewMockContext()
	ctx.Invalidate()

	uitest.AssertInvalidated(t, ctx)
}

func TestAssertNotInvalidated_Pass(t *testing.T) {
	ctx := uitest.NewMockContext()

	uitest.AssertNotInvalidated(t, ctx)
}

func TestAssertCursor_Pass(t *testing.T) {
	ctx := uitest.NewMockContext()
	ctx.SetCursor(widget.CursorPointer)

	uitest.AssertCursor(t, ctx, widget.CursorPointer)
}

func TestAssertFocused_Pass(t *testing.T) {
	ctx := uitest.NewMockContext()
	w := &testWidget{}
	ctx.RequestFocus(w)

	uitest.AssertFocused(t, ctx, w)
}

func TestAssertNotFocused_Pass(t *testing.T) {
	ctx := uitest.NewMockContext()
	w := &testWidget{}

	uitest.AssertNotFocused(t, ctx, w)
}

func TestAssertColorEqual_Pass(t *testing.T) {
	uitest.AssertColorEqual(t, widget.ColorRed, widget.RGBA(1, 0, 0, 1))
}

func TestMockContext_Now(t *testing.T) {
	ctx := uitest.NewMockContext()
	expected := ctx.TimeVal

	if got := ctx.Now(); got != expected {
		t.Errorf("Now() = %v, want %v", got, expected)
	}
}

// --- Test helpers ---

// testWidget is a minimal Widget implementation for testing the test helpers.
type testWidget struct {
	widget.WidgetBase
	preferredSize geometry.Size
	drawText      string
	consumeEvents bool
	eventCount    int
}

func (w *testWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(w.preferredSize)
}

func (w *testWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	if w.drawText != "" {
		canvas.DrawText(w.drawText, w.Bounds(), 14, widget.ColorBlack, false, widget.TextAlignLeft)
	}
}

func (w *testWidget) Event(_ widget.Context, _ event.Event) bool {
	w.eventCount++
	return w.consumeEvents
}
