package collapsible_test

import (
	"image"
	"testing"
	"time"

	"github.com/gogpu/ui/core/collapsible"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// --- Construction Tests ---

func TestNew_Defaults(t *testing.T) {
	w := collapsible.New()

	if !w.IsVisible() {
		t.Error("default collapsible should be visible")
	}
	if !w.IsEnabled() {
		t.Error("default collapsible should be enabled")
	}
	if !w.IsFocusable() {
		t.Error("default collapsible should be focusable")
	}
	if w.IsExpanded() {
		t.Error("default collapsible should be collapsed")
	}
	if w.Progress() != 0.0 {
		t.Errorf("default progress = %v, want 0.0", w.Progress())
	}
}

func TestNew_Expanded(t *testing.T) {
	w := collapsible.New(collapsible.Expanded(true))

	if !w.IsExpanded() {
		t.Error("should be expanded")
	}
	if w.Progress() != 1.0 {
		t.Errorf("expanded progress = %v, want 1.0", w.Progress())
	}
}

func TestNew_WithAllOptions(t *testing.T) {
	toggled := false
	content := &mockWidget{}
	w := collapsible.New(
		collapsible.Title("CPU Usage"),
		collapsible.Content(content),
		collapsible.Expanded(true),
		collapsible.HeaderHeight(48),
		collapsible.HeaderColor(widget.ColorBlue),
		collapsible.ArrowColor(widget.ColorWhite),
		collapsible.Animated(false),
		collapsible.Duration(300*time.Millisecond),
		collapsible.OnToggle(func(expanded bool) { toggled = expanded }),
	)

	if !w.IsExpanded() {
		t.Error("should be expanded")
	}
	_ = toggled
}

func TestNew_WithContent(t *testing.T) {
	content := &mockWidget{}
	w := collapsible.New(collapsible.Content(content))

	children := w.Children()
	if len(children) != 1 {
		t.Fatalf("Children() = %d, want 1", len(children))
	}
	if children[0] != content {
		t.Error("child should be the content widget")
	}
}

func TestNew_NoContent(t *testing.T) {
	w := collapsible.New()

	if children := w.Children(); children != nil {
		t.Errorf("Children() should be nil without content, got %v", children)
	}
}

func TestNew_WithPainter(t *testing.T) {
	p := &testPainter{}
	w := collapsible.New(
		collapsible.Title("Test"),
		collapsible.PainterOpt(p),
	)
	w.SetBounds(geometry.NewRect(0, 0, 200, 36))
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	w.Draw(ctx, canvas)

	if !p.called {
		t.Error("Draw should delegate to the configured painter")
	}
	if p.state.Title != "Test" {
		t.Errorf("HeaderState.Title = %q, want %q", p.state.Title, "Test")
	}
}

// --- Toggle Tests ---

func TestToggle_ClickHeader(t *testing.T) {
	toggled := false
	var newState bool
	w := collapsible.New(
		collapsible.Title("Section"),
		collapsible.Animated(false),
		collapsible.OnToggle(func(expanded bool) {
			toggled = true
			newState = expanded
		}),
	)
	w.SetBounds(geometry.NewRect(0, 0, 200, 36))
	ctx := widget.NewContext()

	// Full click cycle on header.
	simulateClick(w, ctx, geometry.Pt(100, 18))

	if !toggled {
		t.Error("onToggle should have been called after click")
	}
	if !newState {
		t.Error("should expand from collapsed state")
	}
	if !w.IsExpanded() {
		t.Error("should be expanded after toggle")
	}
	if w.Progress() != 1.0 {
		t.Errorf("progress = %v, want 1.0 (animated=false)", w.Progress())
	}
}

func TestToggle_DoubleClick(t *testing.T) {
	toggleCount := 0
	w := collapsible.New(
		collapsible.Animated(false),
		collapsible.OnToggle(func(bool) { toggleCount++ }),
	)
	w.SetBounds(geometry.NewRect(0, 0, 200, 36))
	ctx := widget.NewContext()

	simulateClick(w, ctx, geometry.Pt(100, 18))
	if toggleCount != 1 {
		t.Fatalf("toggleCount = %d after first click, want 1", toggleCount)
	}
	if !w.IsExpanded() {
		t.Error("should be expanded after first toggle")
	}

	simulateClick(w, ctx, geometry.Pt(100, 18))
	if toggleCount != 2 {
		t.Fatalf("toggleCount = %d after second click, want 2", toggleCount)
	}
	if w.IsExpanded() {
		t.Error("should be collapsed after second toggle")
	}
}

func TestToggle_Keyboard_Space(t *testing.T) {
	toggled := false
	w := collapsible.New(
		collapsible.Animated(false),
		collapsible.OnToggle(func(bool) { toggled = true }),
	)
	w.SetFocused(true)
	ctx := widget.NewContext()

	press := event.NewKeyEvent(event.KeyPress, event.KeySpace, 0, event.ModNone)
	w.Event(ctx, press)

	release := event.NewKeyEvent(event.KeyRelease, event.KeySpace, 0, event.ModNone)
	w.Event(ctx, release)

	if !toggled {
		t.Error("Space key should toggle collapsible")
	}
	if !w.IsExpanded() {
		t.Error("should be expanded after Space toggle")
	}
}

func TestToggle_Keyboard_Enter(t *testing.T) {
	toggled := false
	w := collapsible.New(
		collapsible.Animated(false),
		collapsible.OnToggle(func(bool) { toggled = true }),
	)
	w.SetFocused(true)
	ctx := widget.NewContext()

	press := event.NewKeyEvent(event.KeyPress, event.KeyEnter, 0, event.ModNone)
	w.Event(ctx, press)

	release := event.NewKeyEvent(event.KeyRelease, event.KeyEnter, 0, event.ModNone)
	w.Event(ctx, release)

	if !toggled {
		t.Error("Enter key should toggle collapsible")
	}
}

func TestToggle_Keyboard_OtherKey_Ignored(t *testing.T) {
	w := collapsible.New(collapsible.Animated(false))
	w.SetFocused(true)
	ctx := widget.NewContext()

	press := event.NewKeyEvent(event.KeyPress, event.KeyA, 'a', event.ModNone)
	consumed := w.Event(ctx, press)

	if consumed {
		t.Error("key A should not be consumed")
	}
}

func TestToggle_Keyboard_NotFocused_Ignored(t *testing.T) {
	toggled := false
	w := collapsible.New(
		collapsible.Animated(false),
		collapsible.OnToggle(func(bool) { toggled = true }),
	)
	w.SetFocused(false)
	ctx := widget.NewContext()

	press := event.NewKeyEvent(event.KeyPress, event.KeySpace, 0, event.ModNone)
	consumed := w.Event(ctx, press)

	if consumed {
		t.Error("should not consume key when not focused")
	}
	if toggled {
		t.Error("should not toggle when not focused")
	}
}

func TestSetExpanded_Programmatic(t *testing.T) {
	w := collapsible.New(collapsible.Animated(false))

	w.SetExpanded(true)
	if !w.IsExpanded() {
		t.Error("should be expanded after SetExpanded(true)")
	}
	if w.Progress() != 1.0 {
		t.Errorf("progress = %v, want 1.0", w.Progress())
	}

	w.SetExpanded(false)
	if w.IsExpanded() {
		t.Error("should be collapsed after SetExpanded(false)")
	}
	if w.Progress() != 0.0 {
		t.Errorf("progress = %v, want 0.0", w.Progress())
	}
}

func TestSetExpanded_SameState_Noop(t *testing.T) {
	toggleCount := 0
	w := collapsible.New(
		collapsible.Animated(false),
		collapsible.OnToggle(func(bool) { toggleCount++ }),
	)

	w.SetExpanded(false) // Already collapsed.
	if toggleCount != 0 {
		t.Error("setting same state should be a noop")
	}
}

func TestToggle_Method(t *testing.T) {
	w := collapsible.New(collapsible.Animated(false))

	w.Toggle()
	if !w.IsExpanded() {
		t.Error("should be expanded after Toggle()")
	}

	w.Toggle()
	if w.IsExpanded() {
		t.Error("should be collapsed after second Toggle()")
	}
}

// --- Click Outside Header ---

func TestClick_OutsideHeader_NotConsumed(t *testing.T) {
	content := &mockWidget{}
	w := collapsible.New(
		collapsible.Content(content),
		collapsible.Expanded(true),
		collapsible.Animated(false),
	)
	w.SetBounds(geometry.NewRect(0, 0, 200, 136)) // 36 header + 100 content
	ctx := widget.NewContext()

	// Click in content area (below header).
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(100, 80), geometry.Pt(100, 80), event.ModNone)
	consumed := w.Event(ctx, press)

	if consumed {
		t.Error("click in content area should not be consumed by header handler")
	}
}

func TestClick_RightButton_Ignored(t *testing.T) {
	w := collapsible.New(collapsible.Animated(false))
	w.SetBounds(geometry.NewRect(0, 0, 200, 36))
	ctx := widget.NewContext()

	press := event.NewMouseEvent(event.MousePress, event.ButtonRight, event.ButtonStateRight,
		geometry.Pt(100, 18), geometry.Pt(100, 18), event.ModNone)
	consumed := w.Event(ctx, press)

	if consumed {
		t.Error("right button press should not be consumed")
	}
}

// --- Focus Tests ---

func TestFocusable_Interface(t *testing.T) {
	var f widget.Focusable = collapsible.New()
	_ = f
}

func TestFocusable_VisibleAndEnabled(t *testing.T) {
	w := collapsible.New()

	if !w.IsFocusable() {
		t.Error("visible+enabled should be focusable")
	}

	w.SetVisible(false)
	if w.IsFocusable() {
		t.Error("invisible should not be focusable")
	}

	w.SetVisible(true)
	w.SetEnabled(false)
	if w.IsFocusable() {
		t.Error("disabled should not be focusable")
	}
}

// --- Layout Tests ---

func TestLayout_Collapsed(t *testing.T) {
	w := collapsible.New(
		collapsible.HeaderHeight(40),
		collapsible.Content(&mockWidget{preferredSize: geometry.Sz(200, 100)}),
		collapsible.Animated(false),
	)
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(300, 500))

	size := w.Layout(ctx, constraints)

	if size.Height != 40 {
		t.Errorf("collapsed height = %v, want 40 (header only)", size.Height)
	}
}

func TestLayout_Expanded(t *testing.T) {
	w := collapsible.New(
		collapsible.HeaderHeight(40),
		collapsible.Content(&mockWidget{preferredSize: geometry.Sz(200, 100)}),
		collapsible.Expanded(true),
		collapsible.Animated(false),
	)
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(300, 500))

	size := w.Layout(ctx, constraints)

	if size.Height != 140 {
		t.Errorf("expanded height = %v, want 140 (40 header + 100 content)", size.Height)
	}
}

func TestLayout_NoContent(t *testing.T) {
	w := collapsible.New(
		collapsible.HeaderHeight(36),
		collapsible.Expanded(true),
		collapsible.Animated(false),
	)
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(300, 500))

	size := w.Layout(ctx, constraints)

	if size.Height != 36 {
		t.Errorf("height without content = %v, want 36 (header only)", size.Height)
	}
}

func TestLayout_UsesMaxWidth(t *testing.T) {
	w := collapsible.New(collapsible.HeaderHeight(36))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 500))

	size := w.Layout(ctx, constraints)

	if size.Width != 400 {
		t.Errorf("width = %v, want 400 (max width)", size.Width)
	}
}

// --- Draw Tests ---

func TestDraw_EmptyBounds(t *testing.T) {
	w := collapsible.New(collapsible.Title("Test"))
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	// No bounds set, should not draw.
	w.Draw(ctx, canvas)

	if canvas.drawRectCount > 0 || canvas.drawTextCount > 0 {
		t.Error("should not draw anything with empty bounds")
	}
}

func TestDraw_Collapsed_NoContent(t *testing.T) {
	content := &mockWidget{}
	w := collapsible.New(
		collapsible.Title("Section"),
		collapsible.Content(content),
		collapsible.Animated(false),
	)
	w.SetBounds(geometry.NewRect(0, 0, 200, 36))
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	w.Draw(ctx, canvas)

	if content.drawCalled {
		t.Error("collapsed content should not be drawn")
	}
}

func TestDraw_Expanded_DrawsContent(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 100)}
	w := collapsible.New(
		collapsible.Title("Section"),
		collapsible.Content(content),
		collapsible.Expanded(true),
		collapsible.Animated(false),
	)
	w.SetBounds(geometry.NewRect(0, 0, 200, 136))
	ctx := widget.NewContext()

	// Need to layout first to set content size.
	w.Layout(ctx, geometry.Loose(geometry.Sz(200, 500)))
	w.SetBounds(geometry.NewRect(0, 0, 200, 136))

	canvas := &mockCanvas{}
	w.Draw(ctx, canvas)

	if !content.drawCalled {
		t.Error("expanded content should be drawn")
	}
	if canvas.pushClipCount != 1 {
		t.Errorf("PushClip called %d times, want 1", canvas.pushClipCount)
	}
	if canvas.popClipCount != 1 {
		t.Errorf("PopClip called %d times, want 1", canvas.popClipCount)
	}
}

func TestDraw_DelegatesToPainter(t *testing.T) {
	p := &testPainter{}
	w := collapsible.New(
		collapsible.Title("Header"),
		collapsible.Expanded(true),
		collapsible.PainterOpt(p),
	)
	w.SetBounds(geometry.NewRect(10, 10, 200, 136))
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	w.Draw(ctx, canvas)

	if !p.called {
		t.Error("should delegate to painter")
	}
	if p.state.Title != "Header" {
		t.Errorf("title = %q, want %q", p.state.Title, "Header")
	}
	if !p.state.Expanded {
		t.Error("state should show expanded")
	}
	if p.state.ArrowProgress != 1.0 {
		t.Errorf("arrow progress = %v, want 1.0", p.state.ArrowProgress)
	}
}

// --- Animation Tests ---

func TestAnimation_Expand(t *testing.T) {
	w := collapsible.New(
		collapsible.Animated(true),
		collapsible.Duration(100*time.Millisecond),
	)

	if w.Progress() != 0.0 {
		t.Fatalf("initial progress = %v, want 0.0", w.Progress())
	}

	w.Toggle()

	if !w.IsAnimating() {
		t.Error("should be animating after toggle")
	}

	// Simulate animation ticks via Layout.
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 500))

	// Advance 50ms (halfway).
	ctx.SetNow(ctx.Now().Add(50 * time.Millisecond))
	w.Layout(ctx, constraints)

	if w.Progress() <= 0.0 || w.Progress() >= 1.0 {
		t.Errorf("progress at 50ms = %v, should be between 0 and 1", w.Progress())
	}

	// Advance to 100ms (complete).
	ctx.SetNow(ctx.Now().Add(50 * time.Millisecond))
	w.Layout(ctx, constraints)

	if w.Progress() != 1.0 {
		t.Errorf("progress at 100ms = %v, want 1.0", w.Progress())
	}
}

func TestAnimation_Collapse(t *testing.T) {
	w := collapsible.New(
		collapsible.Expanded(true),
		collapsible.Animated(true),
		collapsible.Duration(100*time.Millisecond),
	)

	if w.Progress() != 1.0 {
		t.Fatalf("initial progress = %v, want 1.0", w.Progress())
	}

	w.Toggle()

	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 500))

	// Advance to completion.
	ctx.SetNow(ctx.Now().Add(100 * time.Millisecond))
	w.Layout(ctx, constraints)

	if w.Progress() != 0.0 {
		t.Errorf("progress after collapse = %v, want 0.0", w.Progress())
	}
}

func TestAnimation_Disabled(t *testing.T) {
	w := collapsible.New(collapsible.Animated(false))

	w.Toggle()

	if w.IsAnimating() {
		t.Error("should not animate when animated=false")
	}
	if w.Progress() != 1.0 {
		t.Errorf("instant expand progress = %v, want 1.0", w.Progress())
	}

	w.Toggle()

	if w.Progress() != 0.0 {
		t.Errorf("instant collapse progress = %v, want 0.0", w.Progress())
	}
}

// --- Signal Tests ---

func TestExpandedSignal_TwoWay(t *testing.T) {
	sig := state.NewSignal(false)
	w := collapsible.New(
		collapsible.ExpandedSignal(sig),
		collapsible.Animated(false),
	)

	if w.IsExpanded() {
		t.Error("should start collapsed (signal=false)")
	}

	// Signal -> widget.
	sig.Set(true)
	if !w.IsExpanded() {
		t.Error("should reflect signal change to expanded")
	}

	// Widget -> signal (via Toggle).
	w.Toggle()
	if sig.Get() {
		t.Error("signal should be false after toggle from expanded")
	}
}

func TestExpandedSignal_ClickWritesBack(t *testing.T) {
	sig := state.NewSignal(false)
	w := collapsible.New(
		collapsible.ExpandedSignal(sig),
		collapsible.Animated(false),
	)
	w.SetBounds(geometry.NewRect(0, 0, 200, 36))
	ctx := widget.NewContext()

	simulateClick(w, ctx, geometry.Pt(100, 18))

	if !sig.Get() {
		t.Error("signal should be true after click toggle")
	}
}

func TestExpandedReadonlySignal(t *testing.T) {
	computed := state.NewComputed(func() bool { return true })
	w := collapsible.New(collapsible.ExpandedReadonlySignal(computed))

	if !w.IsExpanded() {
		t.Error("should reflect readonly signal as expanded")
	}
}

func TestExpandedSignal_Mount_CreatesBinding(t *testing.T) {
	sig := state.NewSignal(false)
	w := collapsible.New(
		collapsible.ExpandedSignal(sig),
		collapsible.Animated(false),
	)

	sched := state.NewScheduler(func(_ []widget.Widget) {})
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	w.Mount(ctx)

	dirtyCount := 0
	sched.SetOnDirty(func() { dirtyCount++ })
	sig.Set(true)

	if dirtyCount == 0 {
		t.Error("signal change should mark widget dirty after mount")
	}
}

func TestExpandedSignal_Unmount_CleansBindings(t *testing.T) {
	sig := state.NewSignal(false)
	w := collapsible.New(
		collapsible.ExpandedSignal(sig),
		collapsible.Animated(false),
	)

	sched := state.NewScheduler(func(_ []widget.Widget) {})
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	w.Mount(ctx)
	w.CleanupBindings()
	w.Unmount()

	sig.Set(true)

	if sched.PendingCount() != 0 {
		t.Error("signal change after unmount should not mark widget dirty")
	}
}

func TestExpandedReadonlySignal_Mount_CreatesBinding(t *testing.T) {
	base := state.NewSignal(false)
	computed := state.NewComputed(func() bool {
		return !base.Get()
	}, base)

	w := collapsible.New(collapsible.ExpandedReadonlySignal(computed))

	sched := state.NewScheduler(func(_ []widget.Widget) {})
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	w.Mount(ctx)

	dirtyCount := 0
	sched.SetOnDirty(func() { dirtyCount++ })
	base.Set(true)

	if dirtyCount == 0 {
		t.Error("computed signal dependency change should mark widget dirty after mount")
	}
}

// --- TitleFn Tests ---

func TestTitleFn(t *testing.T) {
	counter := 0
	w := collapsible.New(collapsible.TitleFn(func() string {
		counter++
		return "Dynamic Title"
	}))

	p := &testPainter{}
	w2 := collapsible.New(
		collapsible.TitleFn(func() string { return "Dynamic" }),
		collapsible.PainterOpt(p),
	)
	w2.SetBounds(geometry.NewRect(0, 0, 200, 36))
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	w2.Draw(ctx, canvas)

	if p.state.Title != "Dynamic" {
		t.Errorf("title = %q, want %q", p.state.Title, "Dynamic")
	}
	_ = w
	_ = counter
}

// --- Widget Interface Compliance ---

func TestWidgetInterface(t *testing.T) {
	var w widget.Widget = collapsible.New()
	_ = w
}

func TestLifecycleInterface(t *testing.T) {
	var _ widget.Lifecycle = collapsible.New()
}

// --- Unmount Cancels Animation ---

func TestUnmount_CancelsAnimation(t *testing.T) {
	w := collapsible.New(
		collapsible.Animated(true),
		collapsible.Duration(500*time.Millisecond),
	)

	w.Toggle()
	if !w.IsAnimating() {
		t.Fatal("should be animating after toggle")
	}

	w.Unmount()

	if w.IsAnimating() {
		t.Error("animation should be canceled after unmount")
	}
}

// --- DefaultPainter Tests ---

func TestDefaultPainter_EmptyBounds(t *testing.T) {
	p := collapsible.DefaultPainter{}
	canvas := &mockCanvas{}

	p.PaintHeader(canvas, collapsible.HeaderState{})

	if canvas.drawRectCount > 0 {
		t.Error("should not paint with empty bounds")
	}
}

func TestDefaultPainter_DrawsBackground(t *testing.T) {
	p := collapsible.DefaultPainter{}
	canvas := &recordingCanvas{}

	p.PaintHeader(canvas, collapsible.HeaderState{
		Title:  "Test",
		Bounds: geometry.NewRect(0, 0, 200, 36),
	})

	if len(canvas.drawRects) == 0 {
		t.Error("should draw background rect")
	}
}

func TestDefaultPainter_DrawsTitle(t *testing.T) {
	p := collapsible.DefaultPainter{}
	canvas := &recordingCanvas{}

	p.PaintHeader(canvas, collapsible.HeaderState{
		Title:  "My Section",
		Bounds: geometry.NewRect(0, 0, 200, 36),
	})

	if len(canvas.drawTexts) == 0 {
		t.Fatal("should draw title text")
	}
	if canvas.drawTexts[0].text != "My Section" {
		t.Errorf("text = %q, want %q", canvas.drawTexts[0].text, "My Section")
	}
}

func TestDefaultPainter_DrawsArrow(t *testing.T) {
	p := collapsible.DefaultPainter{}
	canvas := &recordingCanvas{}

	p.PaintHeader(canvas, collapsible.HeaderState{
		Bounds:        geometry.NewRect(0, 0, 200, 36),
		ArrowProgress: 0.5,
	})

	if len(canvas.drawLines) < 2 {
		t.Errorf("should draw arrow lines, got %d", len(canvas.drawLines))
	}
}

func TestDefaultPainter_FocusRing(t *testing.T) {
	p := collapsible.DefaultPainter{}
	canvas := &recordingCanvas{}

	p.PaintHeader(canvas, collapsible.HeaderState{
		Bounds:  geometry.NewRect(0, 0, 200, 36),
		Focused: true,
	})

	if len(canvas.strokeRects) == 0 {
		t.Error("focused header should draw focus ring")
	}
}

func TestDefaultPainter_NoTitle(t *testing.T) {
	p := collapsible.DefaultPainter{}
	canvas := &recordingCanvas{}

	p.PaintHeader(canvas, collapsible.HeaderState{
		Bounds: geometry.NewRect(0, 0, 200, 36),
	})

	if len(canvas.drawTexts) > 0 {
		t.Error("should not draw text when title is empty")
	}
}

func TestDefaultPainter_CustomColors(t *testing.T) {
	p := collapsible.DefaultPainter{}
	canvas := &recordingCanvas{}

	p.PaintHeader(canvas, collapsible.HeaderState{
		Title:       "Test",
		Bounds:      geometry.NewRect(0, 0, 200, 36),
		HeaderColor: widget.ColorBlue,
		ArrowColor:  widget.ColorRed,
	})

	if len(canvas.drawRects) == 0 {
		t.Fatal("should draw background")
	}
	// Background should reflect custom header color (may be modified by state).
	bg := canvas.drawRects[0].color
	if bg.B < 0.5 {
		t.Error("header background should reflect custom blue color")
	}
}

// --- Helper functions ---

func simulateClick(w *collapsible.Widget, ctx widget.Context, pos geometry.Point) {
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		pos, pos, event.ModNone)
	w.Event(ctx, press)

	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		pos, pos, event.ModNone)
	w.Event(ctx, release)
}

// --- testPainter records the call to PaintHeader ---

type testPainter struct {
	called bool
	state  collapsible.HeaderState
}

func (p *testPainter) PaintHeader(_ widget.Canvas, hs collapsible.HeaderState) {
	p.called = true
	p.state = hs
}

// --- mockWidget simulates a content widget ---

type mockWidget struct {
	widget.WidgetBase
	preferredSize geometry.Size
	drawCalled    bool
}

func (m *mockWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	if m.preferredSize == (geometry.Size{}) {
		return c.Constrain(geometry.Sz(100, 100))
	}
	return c.Constrain(m.preferredSize)
}

func (m *mockWidget) Draw(_ widget.Context, _ widget.Canvas) {
	m.drawCalled = true
}

func (m *mockWidget) Event(_ widget.Context, _ event.Event) bool {
	return false
}

// --- mockCanvas for counting operations ---

type mockCanvas struct {
	drawRectCount   int
	drawTextCount   int
	pushClipCount   int
	popClipCount    int
	drawLineCount   int
	strokeRectCount int
}

func (c *mockCanvas) Clear(_ widget.Color)                                                  {}
func (c *mockCanvas) DrawRect(_ geometry.Rect, _ widget.Color)                              { c.drawRectCount++ }
func (c *mockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32)                 { c.strokeRectCount++ }
func (c *mockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32)              {}
func (c *mockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {}
func (c *mockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)                {}
func (c *mockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32)   {}
func (c *mockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32)               { c.drawLineCount++ }

func (c *mockCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
	c.drawTextCount++
}

func (c *mockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}

func (c *mockCanvas) DrawImage(_ image.Image, _ geometry.Point)    {}
func (c *mockCanvas) PushClip(_ geometry.Rect)                     { c.pushClipCount++ }
func (c *mockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *mockCanvas) PopClip()                                     { c.popClipCount++ }
func (c *mockCanvas) PushTransform(_ geometry.Point)               {}
func (c *mockCanvas) PopTransform()                                {}
func (c *mockCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }

// --- recordingCanvas records draw calls for detailed verification ---

type recordingCanvas struct {
	drawRects   []drawRectCall
	strokeRects []strokeRectCall
	drawTexts   []drawTextCall
	drawLines   []drawLineCall
}

type drawRectCall struct {
	r     geometry.Rect
	color widget.Color
}

type strokeRectCall struct {
	r           geometry.Rect
	color       widget.Color
	strokeWidth float32
}

type drawTextCall struct {
	text     string
	bounds   geometry.Rect
	fontSize float32
	color    widget.Color
	bold     bool
	align    widget.TextAlign
}

type drawLineCall struct {
	from, to    geometry.Point
	color       widget.Color
	strokeWidth float32
}

func (c *recordingCanvas) Clear(_ widget.Color) {}

func (c *recordingCanvas) DrawRect(r geometry.Rect, color widget.Color) {
	c.drawRects = append(c.drawRects, drawRectCall{r: r, color: color})
}

func (c *recordingCanvas) StrokeRect(r geometry.Rect, color widget.Color, strokeWidth float32) {
	c.strokeRects = append(c.strokeRects, strokeRectCall{r: r, color: color, strokeWidth: strokeWidth})
}

func (c *recordingCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32)              {}
func (c *recordingCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {}
func (c *recordingCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)                {}
func (c *recordingCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32)   {}

func (c *recordingCanvas) DrawLine(from, to geometry.Point, color widget.Color, strokeWidth float32) {
	c.drawLines = append(c.drawLines, drawLineCall{from: from, to: to, color: color, strokeWidth: strokeWidth})
}

func (c *recordingCanvas) DrawText(text string, bounds geometry.Rect, fontSize float32, color widget.Color, bold bool, align widget.TextAlign) {
	c.drawTexts = append(c.drawTexts, drawTextCall{text: text, bounds: bounds, fontSize: fontSize, color: color, bold: bold, align: align})
}

func (c *recordingCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}

func (c *recordingCanvas) DrawImage(_ image.Image, _ geometry.Point)    {}
func (c *recordingCanvas) PushClip(_ geometry.Rect)                     {}
func (c *recordingCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *recordingCanvas) PopClip()                                     {}
func (c *recordingCanvas) PushTransform(_ geometry.Point)               {}
func (c *recordingCanvas) PopTransform()                                {}
func (c *recordingCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }
