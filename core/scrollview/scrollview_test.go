package scrollview_test

import (
	"image"
	"testing"

	"github.com/gogpu/ui/core/scrollview"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// --- Test helpers ---

// stubWidget is a minimal widget for testing scroll view content.
type stubWidget struct {
	widget.WidgetBase
	preferredSize geometry.Size
}

func newStub(w, h float32) *stubWidget {
	s := &stubWidget{preferredSize: geometry.Sz(w, h)}
	s.SetVisible(true)
	s.SetEnabled(true)
	return s
}

func (s *stubWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(s.preferredSize)
}

func (s *stubWidget) Draw(_ widget.Context, _ widget.Canvas) {}

func (s *stubWidget) Event(_ widget.Context, _ event.Event) bool { return false }

func (s *stubWidget) Children() []widget.Widget { return nil }

// stubCanvas implements widget.Canvas for testing draw calls.
type stubCanvas struct {
	clipStack        []geometry.Rect
	transformStack   []geometry.Point
	clipsPopped      int
	transformsPopped int
}

func (c *stubCanvas) Clear(_ widget.Color)                                                  {}
func (c *stubCanvas) DrawRect(_ geometry.Rect, _ widget.Color)                              {}
func (c *stubCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32)                 {}
func (c *stubCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32)              {}
func (c *stubCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {}
func (c *stubCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)                {}
func (c *stubCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32)   {}
func (c *stubCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32)               {}
func (c *stubCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
}

func (c *stubCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}
func (c *stubCanvas) DrawImage(_ image.Image, _ geometry.Point) {}

func (c *stubCanvas) PushClip(r geometry.Rect) {
	c.clipStack = append(c.clipStack, r)
}
func (c *stubCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}

func (c *stubCanvas) PopClip() {
	c.clipsPopped++
}

func (c *stubCanvas) PushTransform(offset geometry.Point) {
	c.transformStack = append(c.transformStack, offset)
}

func (c *stubCanvas) PopTransform() {
	c.transformsPopped++
}
func (c *stubCanvas) TransformOffset() geometry.Point { return geometry.Point{} }

// --- Construction Tests ---

func TestNew_Defaults(t *testing.T) {
	content := newStub(200, 800)
	sv := scrollview.New(content)

	if !sv.IsVisible() {
		t.Error("should be visible by default")
	}
	if !sv.IsEnabled() {
		t.Error("should be enabled by default")
	}
	if !sv.IsFocusable() {
		t.Error("should be focusable by default")
	}
	children := sv.Children()
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}
}

func TestNew_NilContent(t *testing.T) {
	sv := scrollview.New(nil)

	if sv.Children() != nil {
		t.Error("nil content should yield nil children")
	}
}

func TestNew_WithOptions(t *testing.T) {
	scrolled := false
	sv := scrollview.New(newStub(200, 800),
		scrollview.DirectionOpt(scrollview.Both),
		scrollview.ScrollbarOpt(scrollview.ScrollbarAlways),
		scrollview.ScrollX(10),
		scrollview.ScrollY(20),
		scrollview.ScrollStep(60),
		scrollview.OnScroll(func(_, _ float32) { scrolled = true }),
	)

	if !sv.IsVisible() {
		t.Error("should be visible")
	}
	_ = scrolled
}

// --- Layout Tests ---

func TestLayout_ViewportSize(t *testing.T) {
	ctx := widget.NewContext()
	content := newStub(200, 800)
	sv := scrollview.New(content)

	constraints := geometry.Tight(geometry.Sz(300, 400))
	size := sv.Layout(ctx, constraints)

	if size.Width != 300 || size.Height != 400 {
		t.Errorf("viewport = %v, want (300, 400)", size)
	}
}

func TestLayout_ContentSizeMeasured(t *testing.T) {
	ctx := widget.NewContext()
	content := newStub(200, 800)
	sv := scrollview.New(content)

	sv.SetBounds(geometry.NewRect(0, 0, 300, 400))
	sv.Layout(ctx, geometry.Tight(geometry.Sz(300, 400)))

	cs := sv.ContentSize()
	if cs.Height != 800 {
		t.Errorf("content height = %v, want 800", cs.Height)
	}
}

func TestLayout_HorizontalDirection(t *testing.T) {
	ctx := widget.NewContext()
	content := newStub(1000, 200)
	sv := scrollview.New(content, scrollview.DirectionOpt(scrollview.Horizontal))

	sv.SetBounds(geometry.NewRect(0, 0, 300, 400))
	sv.Layout(ctx, geometry.Tight(geometry.Sz(300, 400)))

	cs := sv.ContentSize()
	if cs.Width != 1000 {
		t.Errorf("content width = %v, want 1000", cs.Width)
	}
}

func TestLayout_BothDirection(t *testing.T) {
	ctx := widget.NewContext()
	content := newStub(1000, 2000)
	sv := scrollview.New(content, scrollview.DirectionOpt(scrollview.Both))

	sv.SetBounds(geometry.NewRect(0, 0, 300, 400))
	sv.Layout(ctx, geometry.Tight(geometry.Sz(300, 400)))

	cs := sv.ContentSize()
	if cs.Width != 1000 || cs.Height != 2000 {
		t.Errorf("content = %v, want (1000, 2000)", cs)
	}
}

// --- Draw Tests ---

func TestDraw_PushesClipAndTransform(t *testing.T) {
	ctx := widget.NewContext()
	content := newStub(200, 800)
	sv := scrollview.New(content,
		scrollview.ScrollY(100),
		scrollview.ScrollbarOpt(scrollview.ScrollbarNever),
	)

	bounds := geometry.NewRect(10, 20, 300, 400)
	sv.SetBounds(bounds)
	sv.Layout(ctx, geometry.Tight(geometry.Sz(300, 400)))

	canvas := &stubCanvas{}
	sv.Draw(ctx, canvas)

	if len(canvas.clipStack) == 0 {
		t.Fatal("expected at least one PushClip call")
	}
	if canvas.clipStack[0] != bounds {
		t.Errorf("clip = %v, want %v", canvas.clipStack[0], bounds)
	}
	if canvas.clipsPopped == 0 {
		t.Error("expected PopClip to be called")
	}
	if len(canvas.transformStack) == 0 {
		t.Fatal("expected PushTransform call")
	}
	if canvas.transformsPopped == 0 {
		t.Error("expected PopTransform to be called")
	}
}

func TestDraw_EmptyBoundsSkips(t *testing.T) {
	ctx := widget.NewContext()
	sv := scrollview.New(newStub(200, 800))
	// Don't set bounds — empty.

	canvas := &stubCanvas{}
	sv.Draw(ctx, canvas)

	if len(canvas.clipStack) != 0 {
		t.Error("should not push clip with empty bounds")
	}
}

// --- Wheel Event Tests ---

func TestWheelEvent_VerticalScroll(t *testing.T) {
	ctx := widget.NewContext()
	content := newStub(200, 800)
	sv := scrollview.New(content)

	bounds := geometry.NewRect(0, 0, 200, 400)
	sv.SetBounds(bounds)
	sv.Layout(ctx, geometry.Tight(geometry.Sz(200, 400)))

	// Simulate scroll down.
	wheel := &event.WheelEvent{
		Position: geometry.Pt(100, 200),
		Delta:    geometry.Pt(0, 1),
	}

	consumed := sv.Event(ctx, wheel)
	if !consumed {
		t.Error("wheel event should be consumed")
	}

	_, scrollY := sv.ScrollOffset()
	if scrollY <= 0 {
		t.Errorf("scrollY = %v, expected > 0 after scrolling down", scrollY)
	}
}

func TestWheelEvent_HorizontalBlocked(t *testing.T) {
	ctx := widget.NewContext()
	content := newStub(200, 800)
	// Default is Vertical -- horizontal wheel should not scroll.
	sv := scrollview.New(content)

	bounds := geometry.NewRect(0, 0, 200, 400)
	sv.SetBounds(bounds)
	sv.Layout(ctx, geometry.Tight(geometry.Sz(200, 400)))

	wheel := &event.WheelEvent{
		Position: geometry.Pt(100, 200),
		Delta:    geometry.Pt(1, 0),
	}

	consumed := sv.Event(ctx, wheel)
	if consumed {
		t.Error("horizontal wheel should not be consumed in Vertical mode")
	}
}

func TestWheelEvent_OutsideBounds(t *testing.T) {
	ctx := widget.NewContext()
	sv := scrollview.New(newStub(200, 800))
	sv.SetBounds(geometry.NewRect(0, 0, 200, 400))
	sv.Layout(ctx, geometry.Tight(geometry.Sz(200, 400)))

	wheel := &event.WheelEvent{
		Position: geometry.Pt(500, 500), // outside
		Delta:    geometry.Pt(0, 1),
	}

	if sv.Event(ctx, wheel) {
		t.Error("wheel event outside bounds should not be consumed")
	}
}

// --- Keyboard Event Tests ---

func TestKeyEvent_ArrowDown(t *testing.T) {
	ctx := widget.NewContext()
	content := newStub(200, 800)
	sv := scrollview.New(content)

	bounds := geometry.NewRect(0, 0, 200, 400)
	sv.SetBounds(bounds)
	sv.Layout(ctx, geometry.Tight(geometry.Sz(200, 400)))

	// Focus the widget.
	ctx.RequestFocus(sv)

	key := &event.KeyEvent{
		KeyType: event.KeyPress,
		Key:     event.KeyDown,
	}

	consumed := sv.Event(ctx, key)
	if !consumed {
		t.Error("arrow down should be consumed when focused")
	}

	_, scrollY := sv.ScrollOffset()
	if scrollY <= 0 {
		t.Error("scrollY should increase after arrow down")
	}
}

func TestKeyEvent_NotFocused(t *testing.T) {
	ctx := widget.NewContext()
	sv := scrollview.New(newStub(200, 800))
	sv.SetBounds(geometry.NewRect(0, 0, 200, 400))
	sv.Layout(ctx, geometry.Tight(geometry.Sz(200, 400)))

	key := &event.KeyEvent{
		KeyType: event.KeyPress,
		Key:     event.KeyDown,
	}

	if sv.Event(ctx, key) {
		t.Error("key event should not be consumed when not focused")
	}
}

func TestKeyEvent_HomeEnd(t *testing.T) {
	ctx := widget.NewContext()
	content := newStub(200, 800)
	sv := scrollview.New(content, scrollview.ScrollY(200))

	bounds := geometry.NewRect(0, 0, 200, 400)
	sv.SetBounds(bounds)
	sv.Layout(ctx, geometry.Tight(geometry.Sz(200, 400)))
	ctx.RequestFocus(sv)

	// Home should scroll to top.
	home := &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyHome}
	sv.Event(ctx, home)
	_, y := sv.ScrollOffset()
	if y != 0 {
		t.Errorf("Home: scrollY = %v, want 0", y)
	}

	// End should scroll to max.
	end := &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyEnd}
	sv.Event(ctx, end)
	_, y = sv.ScrollOffset()
	maxY := float32(800 - 400) // content - viewport
	if y != maxY {
		t.Errorf("End: scrollY = %v, want %v", y, maxY)
	}
}

func TestKeyEvent_PageUpDown(t *testing.T) {
	ctx := widget.NewContext()
	content := newStub(200, 2000)
	sv := scrollview.New(content, scrollview.ScrollY(500))

	bounds := geometry.NewRect(0, 0, 200, 400)
	sv.SetBounds(bounds)
	sv.Layout(ctx, geometry.Tight(geometry.Sz(200, 400)))
	ctx.RequestFocus(sv)

	// Page Down: should increase by viewport height (400).
	pd := &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyPageDown}
	sv.Event(ctx, pd)
	_, y := sv.ScrollOffset()
	if y != 900 {
		t.Errorf("PageDown: scrollY = %v, want 900", y)
	}

	// Page Up: should decrease by viewport height (400).
	pu := &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyPageUp}
	sv.Event(ctx, pu)
	_, y = sv.ScrollOffset()
	if y != 500 {
		t.Errorf("PageUp: scrollY = %v, want 500", y)
	}
}

// --- Scroll Clamping Tests ---

func TestScrollClamping_MinZero(t *testing.T) {
	ctx := widget.NewContext()
	content := newStub(200, 800)
	sv := scrollview.New(content, scrollview.ScrollY(0))

	bounds := geometry.NewRect(0, 0, 200, 400)
	sv.SetBounds(bounds)
	sv.Layout(ctx, geometry.Tight(geometry.Sz(200, 400)))
	ctx.RequestFocus(sv)

	// Scroll up from 0 should stay at 0.
	up := &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyUp}
	sv.Event(ctx, up)
	_, y := sv.ScrollOffset()
	if y != 0 {
		t.Errorf("scrollY = %v, want 0 (clamped)", y)
	}
}

func TestScrollClamping_MaxBound(t *testing.T) {
	ctx := widget.NewContext()
	content := newStub(200, 800)
	maxScroll := float32(400) // 800 - 400
	sv := scrollview.New(content, scrollview.ScrollY(maxScroll))

	bounds := geometry.NewRect(0, 0, 200, 400)
	sv.SetBounds(bounds)
	sv.Layout(ctx, geometry.Tight(geometry.Sz(200, 400)))
	ctx.RequestFocus(sv)

	// Scroll down from max should stay at max.
	down := &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyDown}
	sv.Event(ctx, down)
	_, y := sv.ScrollOffset()
	if y != maxScroll {
		t.Errorf("scrollY = %v, want %v (clamped)", y, maxScroll)
	}
}

// --- Signal Binding Tests ---

func TestSignalBinding_ScrollY(t *testing.T) {
	scrollY := state.NewSignal(float32(0))

	ctx := widget.NewContext()
	content := newStub(200, 800)
	sv := scrollview.New(content, scrollview.ScrollYSignal(scrollY))

	bounds := geometry.NewRect(0, 0, 200, 400)
	sv.SetBounds(bounds)
	sv.Layout(ctx, geometry.Tight(geometry.Sz(200, 400)))

	// Scroll via wheel.
	wheel := &event.WheelEvent{
		Position: geometry.Pt(100, 200),
		Delta:    geometry.Pt(0, 1),
	}
	sv.Event(ctx, wheel)

	if scrollY.Get() <= 0 {
		t.Errorf("signal scrollY = %v, expected > 0 after scroll", scrollY.Get())
	}
}

func TestSignalBinding_ScrollX(t *testing.T) {
	scrollX := state.NewSignal(float32(0))

	ctx := widget.NewContext()
	content := newStub(1000, 200)
	sv := scrollview.New(content,
		scrollview.DirectionOpt(scrollview.Horizontal),
		scrollview.ScrollXSignal(scrollX),
	)

	bounds := geometry.NewRect(0, 0, 300, 200)
	sv.SetBounds(bounds)
	sv.Layout(ctx, geometry.Tight(geometry.Sz(300, 200)))

	wheel := &event.WheelEvent{
		Position: geometry.Pt(100, 100),
		Delta:    geometry.Pt(1, 0),
	}
	sv.Event(ctx, wheel)

	if scrollX.Get() <= 0 {
		t.Errorf("signal scrollX = %v, expected > 0 after scroll", scrollX.Get())
	}
}

// --- OnScroll Callback Test ---

func TestOnScroll_CallbackFired(t *testing.T) {
	var callbackX, callbackY float32
	called := false

	ctx := widget.NewContext()
	content := newStub(200, 800)
	sv := scrollview.New(content,
		scrollview.OnScroll(func(x, y float32) {
			called = true
			callbackX = x
			callbackY = y
		}),
	)

	bounds := geometry.NewRect(0, 0, 200, 400)
	sv.SetBounds(bounds)
	sv.Layout(ctx, geometry.Tight(geometry.Sz(200, 400)))

	wheel := &event.WheelEvent{
		Position: geometry.Pt(100, 200),
		Delta:    geometry.Pt(0, 1),
	}
	sv.Event(ctx, wheel)

	if !called {
		t.Error("OnScroll callback should have been called")
	}
	if callbackY <= 0 {
		t.Errorf("callbackY = %v, expected > 0", callbackY)
	}
	_ = callbackX
}

// --- Direction Tests ---

func TestDirection_String(t *testing.T) {
	tests := []struct {
		d    scrollview.ScrollDirection
		want string
	}{
		{scrollview.Vertical, "Vertical"},
		{scrollview.Horizontal, "Horizontal"},
		{scrollview.Both, "Both"},
		{scrollview.ScrollDirection(99), "Unknown"},
	}
	for _, tt := range tests {
		if got := tt.d.String(); got != tt.want {
			t.Errorf("%d.String() = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestScrollbarVisibility_String(t *testing.T) {
	tests := []struct {
		v    scrollview.ScrollbarVisibility
		want string
	}{
		{scrollview.ScrollbarAuto, "Auto"},
		{scrollview.ScrollbarAlways, "Always"},
		{scrollview.ScrollbarNever, "Never"},
		{scrollview.ScrollbarVisibility(99), "Unknown"},
	}
	for _, tt := range tests {
		if got := tt.v.String(); got != tt.want {
			t.Errorf("%d.String() = %q, want %q", tt.v, got, tt.want)
		}
	}
}

// --- Painter Tests ---

func TestDefaultPainter_EmptyBounds(t *testing.T) {
	p := scrollview.DefaultPainter{}
	canvas := &stubCanvas{}
	ps := scrollview.PaintState{
		Bounds: geometry.Rect{}, // empty
	}
	// Should not panic.
	p.PaintScrollbar(canvas, ps)
}

func TestDefaultPainter_DrawsScrollbar(t *testing.T) {
	p := scrollview.DefaultPainter{}
	canvas := &stubCanvas{}
	ps := scrollview.PaintState{
		Bounds:         geometry.NewRect(0, 0, 300, 400),
		VScrollVisible: true,
		VThumbRect:     geometry.NewRect(288, 10, 8, 80),
		VTrackRect:     geometry.NewRect(288, 0, 12, 400),
	}
	// Should draw track and thumb (no panic).
	p.PaintScrollbar(canvas, ps)
}

// --- Content Tests ---

func TestContent_ReturnsChild(t *testing.T) {
	content := newStub(200, 800)
	sv := scrollview.New(content)

	if sv.Content() != content {
		t.Error("Content() should return the wrapped widget")
	}
}

// --- ScrollTo Tests ---

func TestScrollOffset_ReturnsCurrentPosition(t *testing.T) {
	ctx := widget.NewContext()
	content := newStub(200, 800)
	sv := scrollview.New(content, scrollview.ScrollY(100))

	bounds := geometry.NewRect(0, 0, 200, 400)
	sv.SetBounds(bounds)
	sv.Layout(ctx, geometry.Tight(geometry.Sz(200, 400)))

	_, y := sv.ScrollOffset()
	if y != 100 {
		t.Errorf("scrollY = %v, want 100", y)
	}
}

// --- Lifecycle Tests ---

func TestMount_WithSignals(t *testing.T) {
	ctx := widget.NewContext()
	scrollY := state.NewSignal(float32(0))
	sv := scrollview.New(newStub(200, 800), scrollview.ScrollYSignal(scrollY))

	// Mount should not panic even without scheduler.
	sv.Mount(ctx)
}

func TestUnmount(t *testing.T) {
	sv := scrollview.New(newStub(200, 800))
	// Should not panic.
	sv.Unmount()
}

// --- Scrollbar Visibility Tests ---

func TestScrollbar_AutoHidesWhenContentFits(t *testing.T) {
	ctx := widget.NewContext()
	content := newStub(200, 200) // fits in 300x400
	sv := scrollview.New(content)

	bounds := geometry.NewRect(0, 0, 300, 400)
	sv.SetBounds(bounds)
	sv.Layout(ctx, geometry.Tight(geometry.Sz(300, 400)))

	canvas := &stubCanvas{}
	sv.Draw(ctx, canvas)
	// Scrollbar auto — content fits, so no scrollbar rect calls
	// (just verify no panic and clip/transform used).
}

func TestScrollbar_AlwaysShows(t *testing.T) {
	ctx := widget.NewContext()
	content := newStub(200, 200)
	sv := scrollview.New(content,
		scrollview.ScrollbarOpt(scrollview.ScrollbarAlways),
	)

	bounds := geometry.NewRect(0, 0, 300, 400)
	sv.SetBounds(bounds)
	sv.Layout(ctx, geometry.Tight(geometry.Sz(300, 400)))

	canvas := &stubCanvas{}
	sv.Draw(ctx, canvas)
	// Just verify no panic.
}

func TestScrollbar_NeverHides(t *testing.T) {
	ctx := widget.NewContext()
	content := newStub(200, 800)
	sv := scrollview.New(content,
		scrollview.ScrollbarOpt(scrollview.ScrollbarNever),
	)

	bounds := geometry.NewRect(0, 0, 300, 400)
	sv.SetBounds(bounds)
	sv.Layout(ctx, geometry.Tight(geometry.Sz(300, 400)))

	canvas := &stubCanvas{}
	sv.Draw(ctx, canvas)
	// Just verify no panic.
}

// --- Interface Compliance ---

func TestWidgetInterface(t *testing.T) {
	var _ widget.Widget = scrollview.New(nil)
}

func TestFocusableInterface(t *testing.T) {
	var _ widget.Focusable = scrollview.New(nil)
}

func TestLifecycleInterface(t *testing.T) {
	var _ widget.Lifecycle = scrollview.New(nil)
}
