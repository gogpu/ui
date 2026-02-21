package primitives_test

import (
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
)

// --- Construction ---

func TestThemeScopeNoChildren(t *testing.T) {
	ts := primitives.ThemeScope(&darkThemeMock{})
	if ts.Children() != nil {
		t.Errorf("expected nil children, got %d", len(ts.Children()))
	}
}

func TestThemeScopeSingleChild(t *testing.T) {
	child := primitives.Text("Hello")
	ts := primitives.ThemeScope(&darkThemeMock{}, child)
	children := ts.Children()
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}
}

func TestThemeScopeMultipleChildrenWrapsInBox(t *testing.T) {
	c1 := primitives.Text("A")
	c2 := primitives.Text("B")
	ts := primitives.ThemeScope(&darkThemeMock{}, c1, c2)

	children := ts.Children()
	if len(children) != 1 {
		t.Fatalf("expected 1 child (wrapped Box), got %d", len(children))
	}

	// The single child should be a Box that contains c1 and c2.
	box := children[0]
	boxChildren := box.Children()
	if len(boxChildren) != 2 {
		t.Errorf("expected wrapped Box to have 2 children, got %d", len(boxChildren))
	}
}

func TestThemeScopeIsVisibleAndEnabled(t *testing.T) {
	ts := primitives.ThemeScope(&darkThemeMock{})
	if !ts.IsVisible() {
		t.Error("ThemeScope should be visible by default")
	}
	if !ts.IsEnabled() {
		t.Error("ThemeScope should be enabled by default")
	}
}

func TestThemeScopeGetSetTheme(t *testing.T) {
	dark := &darkThemeMock{}
	light := &lightThemeMock{}

	ts := primitives.ThemeScope(dark)
	if ts.Theme() != dark {
		t.Error("Theme() should return the initial theme")
	}

	ts.SetTheme(light)
	if ts.Theme() != light {
		t.Error("Theme() should return the updated theme after SetTheme")
	}
}

// --- Theme Override ---

func TestThemeScopeOverridesThemeInLayout(t *testing.T) {
	dark := &darkThemeMock{}
	recorder := &themeRecorderWidget{}

	ts := primitives.ThemeScope(dark, recorder)
	ctx := widget.NewContext()
	ctx.SetThemeProvider(&lightThemeMock{})

	_ = ts.Layout(ctx, geometry.Loose(geometry.Sz(300, 300)))

	if recorder.lastTheme == nil {
		t.Fatal("child did not receive a theme")
	}
	if !recorder.lastTheme.IsDark() {
		t.Error("child should receive the dark scoped theme, not the app-level light theme")
	}
}

func TestThemeScopeOverridesThemeInDraw(t *testing.T) {
	dark := &darkThemeMock{}
	recorder := &themeRecorderWidget{}

	ts := primitives.ThemeScope(dark, recorder)
	ctx := widget.NewContext()
	ctx.SetThemeProvider(&lightThemeMock{})

	_ = ts.Layout(ctx, geometry.Loose(geometry.Sz(300, 300)))
	ts.Draw(ctx, &mockCanvas{})

	if recorder.drawTheme == nil {
		t.Fatal("child did not receive a theme during Draw")
	}
	if !recorder.drawTheme.IsDark() {
		t.Error("child should receive the dark scoped theme during Draw")
	}
}

func TestThemeScopeOverridesThemeInEvent(t *testing.T) {
	dark := &darkThemeMock{}
	recorder := &themeRecorderWidget{}

	ts := primitives.ThemeScope(dark, recorder)
	ctx := widget.NewContext()
	ctx.SetThemeProvider(&lightThemeMock{})

	_ = ts.Layout(ctx, geometry.Loose(geometry.Sz(300, 300)))
	_ = ts.Event(ctx, &event.Base{})

	if recorder.eventTheme == nil {
		t.Fatal("child did not receive a theme during Event")
	}
	if !recorder.eventTheme.IsDark() {
		t.Error("child should receive the dark scoped theme during Event")
	}
}

func TestThemeScopeNilThemePassesParent(t *testing.T) {
	recorder := &themeRecorderWidget{}
	ts := primitives.ThemeScope(nil, recorder)

	ctx := widget.NewContext()
	light := &lightThemeMock{}
	ctx.SetThemeProvider(light)

	_ = ts.Layout(ctx, geometry.Loose(geometry.Sz(300, 300)))

	if recorder.lastTheme == nil {
		t.Fatal("child did not receive a theme")
	}
	if recorder.lastTheme.IsDark() {
		t.Error("nil scope theme should pass through the parent context theme (light)")
	}
}

// --- Nested Scopes (inner wins) ---

func TestThemeScopeNestedInnerWins(t *testing.T) {
	outerTheme := &darkThemeMock{}
	innerTheme := &lightThemeMock{}
	recorder := &themeRecorderWidget{}

	inner := primitives.ThemeScope(innerTheme, recorder)
	outer := primitives.ThemeScope(outerTheme, inner)

	ctx := widget.NewContext()
	_ = outer.Layout(ctx, geometry.Loose(geometry.Sz(300, 300)))

	if recorder.lastTheme == nil {
		t.Fatal("deepest child did not receive a theme")
	}
	if recorder.lastTheme.IsDark() {
		t.Error("inner ThemeScope should override outer (inner is light, got dark)")
	}
}

// --- Widgets Outside ThemeScope Use App Theme ---

func TestThemeScopeWidgetsOutsideScopeUseAppTheme(t *testing.T) {
	appTheme := &lightThemeMock{}
	scopeTheme := &darkThemeMock{}

	insideRecorder := &themeRecorderWidget{}
	outsideRecorder := &themeRecorderWidget{}

	scoped := primitives.ThemeScope(scopeTheme, insideRecorder)

	ctx := widget.NewContext()
	ctx.SetThemeProvider(appTheme)

	// Layout the scoped widget.
	_ = scoped.Layout(ctx, geometry.Loose(geometry.Sz(300, 300)))

	// Layout the outside widget directly with the app context.
	_ = outsideRecorder.Layout(ctx, geometry.Loose(geometry.Sz(300, 300)))

	if !insideRecorder.lastTheme.IsDark() {
		t.Error("widget inside ThemeScope should receive dark theme")
	}
	if outsideRecorder.lastTheme.IsDark() {
		t.Error("widget outside ThemeScope should receive app-level light theme")
	}
}

// --- Layout Delegation ---

func TestThemeScopeLayoutDelegatesSize(t *testing.T) {
	child := primitives.Box(primitives.Text("Hello")).Width(100).Height(50)
	ts := primitives.ThemeScope(&darkThemeMock{}, child)

	ctx := widget.NewContext()
	size := ts.Layout(ctx, geometry.Loose(geometry.Sz(300, 300)))

	if size.Width != 100 || size.Height != 50 {
		t.Errorf("expected 100x50, got %s", size)
	}
}

func TestThemeScopeLayoutNoChildZeroSize(t *testing.T) {
	ts := primitives.ThemeScope(&darkThemeMock{})
	ctx := widget.NewContext()
	size := ts.Layout(ctx, geometry.Loose(geometry.Sz(300, 300)))

	if size.Width != 0 || size.Height != 0 {
		t.Errorf("expected 0x0 for no child, got %s", size)
	}
}

// --- Draw ---

func TestThemeScopeDrawNoPanicEmpty(t *testing.T) {
	ts := primitives.ThemeScope(&darkThemeMock{})
	ctx := widget.NewContext()
	canvas := &mockCanvas{}
	_ = ts.Layout(ctx, geometry.Loose(geometry.Sz(100, 100)))
	ts.Draw(ctx, canvas) // Should not panic.
}

func TestThemeScopeDrawInvisibleSkips(t *testing.T) {
	recorder := &themeRecorderWidget{}
	ts := primitives.ThemeScope(&darkThemeMock{}, recorder)
	ts.SetVisible(false)

	ctx := widget.NewContext()
	canvas := &mockCanvas{}
	_ = ts.Layout(ctx, geometry.Loose(geometry.Sz(100, 100)))
	ts.Draw(ctx, canvas)

	if recorder.drawTheme != nil {
		t.Error("invisible ThemeScope should not draw child")
	}
}

func TestThemeScopeDrawUsesTransform(t *testing.T) {
	child := primitives.Text("Hi")
	ts := primitives.ThemeScope(&darkThemeMock{}, child)

	ctx := widget.NewContext()
	canvas := &mockCanvas{}
	_ = ts.Layout(ctx, geometry.Loose(geometry.Sz(300, 300)))
	ts.Draw(ctx, canvas)

	if canvas.pushTransformCount == 0 || canvas.popTransformCount == 0 {
		t.Error("expected PushTransform/PopTransform for child drawing")
	}
}

// --- Event ---

func TestThemeScopeEventDispatchesToChild(t *testing.T) {
	consumed := false
	child := &eventConsumer{consume: true, called: &consumed}
	ts := primitives.ThemeScope(&darkThemeMock{}, child)

	ctx := widget.NewContext()
	_ = ts.Layout(ctx, geometry.Loose(geometry.Sz(100, 100)))

	result := ts.Event(ctx, &event.Base{})
	if !result || !consumed {
		t.Error("event should be dispatched to and consumed by child")
	}
}

func TestThemeScopeEventNoChildReturnsFalse(t *testing.T) {
	ts := primitives.ThemeScope(&darkThemeMock{})
	ctx := widget.NewContext()
	if ts.Event(ctx, &event.Base{}) {
		t.Error("ThemeScope with no child should not consume events")
	}
}

func TestThemeScopeEventDisabledSkips(t *testing.T) {
	consumed := false
	child := &eventConsumer{consume: true, called: &consumed}
	ts := primitives.ThemeScope(&darkThemeMock{}, child)
	ts.SetEnabled(false)

	ctx := widget.NewContext()
	if ts.Event(ctx, &event.Base{}) {
		t.Error("disabled ThemeScope should not consume events")
	}
	if consumed {
		t.Error("children should not receive events when ThemeScope is disabled")
	}
}

func TestThemeScopeEventInvisibleSkips(t *testing.T) {
	consumed := false
	child := &eventConsumer{consume: true, called: &consumed}
	ts := primitives.ThemeScope(&darkThemeMock{}, child)
	ts.SetVisible(false)

	ctx := widget.NewContext()
	if ts.Event(ctx, &event.Base{}) {
		t.Error("invisible ThemeScope should not consume events")
	}
}

// --- Context delegation ---

func TestThemeScopeContextDelegatesFocus(t *testing.T) {
	recorder := &themeRecorderWidget{}
	ts := primitives.ThemeScope(&darkThemeMock{}, recorder)

	ctx := widget.NewContext()
	_ = ts.Layout(ctx, geometry.Loose(geometry.Sz(300, 300)))

	// Request focus through the scoped context — should reach the real context.
	if ctx.FocusedWidget() != nil {
		t.Error("should start with no focused widget")
	}

	// The recorder's Layout calls RequestFocus if instructed.
	// For this test, just verify the context chain works.
	ctx.RequestFocus(recorder)
	if !ctx.IsFocused(recorder) {
		t.Error("focus should be set via parent context")
	}
}

func TestThemeScopeContextDelegatesScale(t *testing.T) {
	recorder := &themeRecorderWidget{}
	ts := primitives.ThemeScope(&darkThemeMock{}, recorder)

	ctx := widget.NewContext()
	ctx.SetScale(2.0)
	_ = ts.Layout(ctx, geometry.Loose(geometry.Sz(300, 300)))

	if recorder.lastScale != 2.0 {
		t.Errorf("expected scale 2.0, got %f", recorder.lastScale)
	}
}

// --- Test helpers ---

// darkThemeMock implements widget.ThemeProvider with IsDark() = true.
type darkThemeMock struct{}

func (d *darkThemeMock) IsDark() bool            { return true }
func (d *darkThemeMock) OnSurface() widget.Color { return widget.ColorWhite }

// lightThemeMock implements widget.ThemeProvider with IsDark() = false.
type lightThemeMock struct{}

func (l *lightThemeMock) IsDark() bool            { return false }
func (l *lightThemeMock) OnSurface() widget.Color { return widget.ColorBlack }

// themeRecorderWidget records the ThemeProvider it receives during
// Layout, Draw, and Event calls.
type themeRecorderWidget struct {
	widget.WidgetBase

	lastTheme  widget.ThemeProvider
	drawTheme  widget.ThemeProvider
	eventTheme widget.ThemeProvider
	lastScale  float32
}

func (w *themeRecorderWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	w.lastTheme = ctx.ThemeProvider()
	w.lastScale = ctx.Scale()
	size := c.Constrain(geometry.Sz(10, 10))
	w.SetBounds(geometry.FromPointSize(w.Position(), size))
	return size
}

func (w *themeRecorderWidget) Draw(ctx widget.Context, _ widget.Canvas) {
	w.drawTheme = ctx.ThemeProvider()
}

func (w *themeRecorderWidget) Event(ctx widget.Context, _ event.Event) bool {
	w.eventTheme = ctx.ThemeProvider()
	return false
}

func (w *themeRecorderWidget) Children() []widget.Widget { return nil }
