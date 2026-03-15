package transition

import (
	"image"
	"testing"
	"time"

	"github.com/gogpu/ui/animation"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// --- Test helpers ---

// mockWidget implements widget.Widget for testing.
type mockWidget struct {
	widget.WidgetBase
	drawn    int
	lastCtx  widget.Context
	eventRet bool
}

func newMockWidget() *mockWidget {
	w := &mockWidget{}
	w.SetVisible(true)
	w.SetEnabled(true)
	return w
}

func (m *mockWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(100, 50))
}

func (m *mockWidget) Draw(ctx widget.Context, _ widget.Canvas) {
	m.drawn++
	m.lastCtx = ctx
}

func (m *mockWidget) Event(_ widget.Context, _ event.Event) bool {
	return m.eventRet
}

func (m *mockWidget) Children() []widget.Widget { return nil }

// mockCanvas implements widget.Canvas for testing.
type mockCanvas struct {
	transforms []geometry.Point
	clips      int
}

func (c *mockCanvas) Clear(_ widget.Color)                                     {}
func (c *mockCanvas) DrawRect(_ geometry.Rect, _ widget.Color)                 {}
func (c *mockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32)    {}
func (c *mockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32) {}
func (c *mockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {
}
func (c *mockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color) {}
func (c *mockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {
}
func (c *mockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32) {}
func (c *mockCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
}

func (c *mockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}
func (c *mockCanvas) DrawImage(_ image.Image, _ geometry.Point)    {}
func (c *mockCanvas) PushClip(_ geometry.Rect)                     { c.clips++ }
func (c *mockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) { c.clips++ }
func (c *mockCanvas) PopClip()                                     { c.clips-- }
func (c *mockCanvas) PushTransform(offset geometry.Point) {
	c.transforms = append(c.transforms, offset)
}
func (c *mockCanvas) PopTransform() {
	if len(c.transforms) > 0 {
		c.transforms = c.transforms[:len(c.transforms)-1]
	}
}
func (c *mockCanvas) TransformOffset() geometry.Point { return geometry.Point{} }

// opacityCanvas extends mockCanvas with OpacityPusher support.
type opacityCanvas struct {
	mockCanvas
	opacities []float64
}

func (c *opacityCanvas) PushOpacity(opacity float64) {
	c.opacities = append(c.opacities, opacity)
}

func (c *opacityCanvas) PopOpacity() {
	if len(c.opacities) > 0 {
		c.opacities = c.opacities[:len(c.opacities)-1]
	}
}

// mockContext implements widget.Context for testing.
type mockContext struct {
	now         time.Time
	delta       time.Duration
	invalidated bool
}

func newMockContext(now time.Time) *mockContext {
	return &mockContext{now: now}
}

func (c *mockContext) RequestFocus(_ widget.Widget)          {}
func (c *mockContext) ReleaseFocus(_ widget.Widget)          {}
func (c *mockContext) IsFocused(_ widget.Widget) bool        { return false }
func (c *mockContext) FocusedWidget() widget.Widget          { return nil }
func (c *mockContext) Now() time.Time                        { return c.now }
func (c *mockContext) DeltaTime() time.Duration              { return c.delta }
func (c *mockContext) Invalidate()                           { c.invalidated = true }
func (c *mockContext) InvalidateRect(_ geometry.Rect)        {}
func (c *mockContext) Cursor() widget.CursorType             { return widget.CursorDefault }
func (c *mockContext) SetCursor(_ widget.CursorType)         {}
func (c *mockContext) Scale() float32                        { return 1.0 }
func (c *mockContext) ThemeProvider() widget.ThemeProvider   { return nil }
func (c *mockContext) OverlayManager() widget.OverlayManager { return nil }
func (c *mockContext) WindowSize() geometry.Size             { return geometry.Sz(800, 600) }
func (c *mockContext) Scheduler() widget.SchedulerRef        { return nil }

// --- Effect tests ---

func TestNoneEffect(t *testing.T) {
	e := None()
	if !e.IsNone() {
		t.Error("None() effect should report IsNone")
	}
}

func TestFadeInEffect(t *testing.T) {
	e := FadeIn()
	if e.IsNone() {
		t.Error("FadeIn should not be None")
	}
	if e.OpacityStart != 0 || e.OpacityEnd != 1 {
		t.Errorf("FadeIn opacity: got %v->%v, want 0->1", e.OpacityStart, e.OpacityEnd)
	}
}

func TestFadeOutEffect(t *testing.T) {
	e := FadeOut()
	if e.IsNone() {
		t.Error("FadeOut should not be None")
	}
	if e.OpacityStart != 1 || e.OpacityEnd != 0 {
		t.Errorf("FadeOut opacity: got %v->%v, want 1->0", e.OpacityStart, e.OpacityEnd)
	}
}

func TestSlideInEffect(t *testing.T) {
	tests := []struct {
		dir          Direction
		wantX, wantY float64
	}{
		{FromTop, 0, -1},
		{FromBottom, 0, 1},
		{FromLeft, -1, 0},
		{FromRight, 1, 0},
		{ToTop, 0, -1},   // aliased
		{ToBottom, 0, 1}, // aliased
		{ToLeft, -1, 0},  // aliased
		{ToRight, 1, 0},  // aliased
	}
	for _, tt := range tests {
		e := SlideIn(tt.dir)
		if e.IsNone() {
			t.Errorf("SlideIn(%d) should not be None", tt.dir)
		}
		if e.TranslateXFraction != tt.wantX || e.TranslateYFraction != tt.wantY {
			t.Errorf("SlideIn(%d): got (%v,%v), want (%v,%v)",
				tt.dir, e.TranslateXFraction, e.TranslateYFraction, tt.wantX, tt.wantY)
		}
	}
}

func TestSlideOutEffect(t *testing.T) {
	tests := []struct {
		dir          Direction
		wantX, wantY float64
	}{
		{ToTop, 0, -1},
		{ToBottom, 0, 1},
		{ToLeft, -1, 0},
		{ToRight, 1, 0},
	}
	for _, tt := range tests {
		e := SlideOut(tt.dir)
		if e.IsNone() {
			t.Errorf("SlideOut(%d) should not be None", tt.dir)
		}
		if e.TranslateXFraction != tt.wantX || e.TranslateYFraction != tt.wantY {
			t.Errorf("SlideOut(%d): got (%v,%v), want (%v,%v)",
				tt.dir, e.TranslateXFraction, e.TranslateYFraction, tt.wantX, tt.wantY)
		}
	}
}

func TestScaleInEffect(t *testing.T) {
	e := ScaleIn()
	if e.IsNone() {
		t.Error("ScaleIn should not be None")
	}
	if e.ScaleStart != 0.8 || e.ScaleEnd != 1.0 {
		t.Errorf("ScaleIn scale: got %v->%v, want 0.8->1.0", e.ScaleStart, e.ScaleEnd)
	}
	if e.OpacityStart != 0 || e.OpacityEnd != 1 {
		t.Errorf("ScaleIn opacity: got %v->%v, want 0->1", e.OpacityStart, e.OpacityEnd)
	}
}

func TestScaleOutEffect(t *testing.T) {
	e := ScaleOut()
	if e.IsNone() {
		t.Error("ScaleOut should not be None")
	}
	if e.ScaleStart != 1.0 || e.ScaleEnd != 0.8 {
		t.Errorf("ScaleOut scale: got %v->%v, want 1.0->0.8", e.ScaleStart, e.ScaleEnd)
	}
}

func TestLerp(t *testing.T) {
	tests := []struct {
		a, b, t float64
		want    float64
	}{
		{0, 1, 0, 0},
		{0, 1, 1, 1},
		{0, 1, 0.5, 0.5},
		{10, 20, 0.25, 12.5},
		{-1, 1, 0.5, 0},
	}
	for _, tt := range tests {
		got := lerp(tt.a, tt.b, tt.t)
		if got != tt.want {
			t.Errorf("lerp(%v, %v, %v) = %v, want %v", tt.a, tt.b, tt.t, got, tt.want)
		}
	}
}

// --- Wrap and option tests ---

func TestWrapDefaults(t *testing.T) {
	child := newMockWidget()
	tr := Wrap(child)

	if tr.Child() != child {
		t.Error("Child() should return wrapped widget")
	}
	if !tr.IsShown() {
		t.Error("should start visible")
	}
	if tr.IsAnimating() {
		t.Error("should not be animating initially")
	}
	if tr.enterEffect.IsNone() != true {
		t.Error("default enter effect should be None")
	}
	if tr.exitEffect.IsNone() != true {
		t.Error("default exit effect should be None")
	}
	if tr.duration != defaultDuration {
		t.Errorf("default duration: got %v, want %v", tr.duration, defaultDuration)
	}
}

func TestWrapWithOptions(t *testing.T) {
	child := newMockWidget()
	dur := 500 * time.Millisecond
	tr := Wrap(child,
		EnterEffect(FadeIn()),
		ExitEffect(FadeOut()),
		Duration(dur),
		Easing(animation.Linear),
	)

	if tr.enterEffect.OpacityStart != 0 {
		t.Error("enter effect should be FadeIn")
	}
	if tr.exitEffect.OpacityStart != 1 {
		t.Error("exit effect should be FadeOut")
	}
	if tr.duration != dur {
		t.Errorf("duration: got %v, want %v", tr.duration, dur)
	}
}

func TestWrapNilChild(t *testing.T) {
	tr := Wrap(nil)
	if tr.Children() != nil {
		t.Error("Children should return nil for nil child")
	}
	// Layout/Draw/Event should not panic with nil child.
	ctx := newMockContext(time.Now())
	canvas := &mockCanvas{}
	_ = tr.Layout(ctx, geometry.Constraints{MaxWidth: 100, MaxHeight: 100})
	tr.Draw(ctx, canvas)
	if tr.Event(ctx, &event.MouseEvent{}) {
		t.Error("Event should return false for nil child")
	}
}

// --- Show/Hide tests ---

func TestShowHideNoEffect(t *testing.T) {
	child := newMockWidget()
	tr := Wrap(child) // default None effects

	tr.Hide()
	if tr.IsShown() {
		t.Error("should be hidden after Hide with no exit effect")
	}
	if tr.IsAnimating() {
		t.Error("should not animate with None effect")
	}

	tr.Show()
	if !tr.IsShown() {
		t.Error("should be shown after Show with no enter effect")
	}
	if tr.IsAnimating() {
		t.Error("should not animate with None effect")
	}
}

func TestShowStartsAnimation(t *testing.T) {
	child := newMockWidget()
	tr := Wrap(child, EnterEffect(FadeIn()))
	tr.shown = false // start hidden

	tr.Show()
	if !tr.IsShown() {
		t.Error("should be shown after Show")
	}
	if !tr.IsAnimating() {
		t.Error("should be animating after Show with FadeIn")
	}
	if !tr.entering {
		t.Error("should be in entering state")
	}
}

func TestHideStartsAnimation(t *testing.T) {
	child := newMockWidget()
	tr := Wrap(child, ExitEffect(FadeOut()))

	tr.Hide()
	// Should still be "shown" during exit animation.
	if !tr.IsAnimating() {
		t.Error("should be animating after Hide with FadeOut")
	}
	if tr.entering {
		t.Error("should be in exiting state")
	}
}

func TestShowAlreadyVisible(t *testing.T) {
	child := newMockWidget()
	tr := Wrap(child, EnterEffect(FadeIn()))
	// Already shown, not animating.
	tr.Show()
	if tr.IsAnimating() {
		t.Error("Show on already-visible widget should not start animation")
	}
}

func TestHideAlreadyHidden(t *testing.T) {
	child := newMockWidget()
	tr := Wrap(child, ExitEffect(FadeOut()))
	tr.shown = false
	tr.animating = false

	tr.Hide()
	if tr.IsAnimating() {
		t.Error("Hide on already-hidden widget should not start animation")
	}
}

func TestShowWithZeroDuration(t *testing.T) {
	child := newMockWidget()
	tr := Wrap(child, EnterEffect(FadeIn()), Duration(0))
	tr.shown = false

	tr.Show()
	if tr.IsAnimating() {
		t.Error("should not animate with zero duration")
	}
	if !tr.IsShown() {
		t.Error("should be shown immediately with zero duration")
	}
}

// --- Layout tests ---

func TestLayoutDelegatesToChild(t *testing.T) {
	child := newMockWidget()
	tr := Wrap(child)
	ctx := newMockContext(time.Now())

	tr.SetBounds(geometry.NewRect(10, 20, 200, 200))
	size := tr.Layout(ctx, geometry.Constraints{MaxWidth: 200, MaxHeight: 200})

	if size.Width != 100 || size.Height != 50 {
		t.Errorf("Layout size: got %v, want (100,50)", size)
	}
	if tr.childSize.Width != 100 || tr.childSize.Height != 50 {
		t.Errorf("cached childSize: got %v, want (100,50)", tr.childSize)
	}
}

// --- Draw tests ---

func TestDrawHiddenDoesNothing(t *testing.T) {
	child := newMockWidget()
	tr := Wrap(child)
	tr.shown = false
	tr.animating = false

	ctx := newMockContext(time.Now())
	canvas := &mockCanvas{}
	tr.Draw(ctx, canvas)

	if child.drawn != 0 {
		t.Error("hidden transition should not draw child")
	}
}

func TestDrawVisibleDrawsChild(t *testing.T) {
	child := newMockWidget()
	tr := Wrap(child)

	ctx := newMockContext(time.Now())
	canvas := &mockCanvas{}
	tr.Draw(ctx, canvas)

	if child.drawn != 1 {
		t.Errorf("visible transition should draw child once, got %d", child.drawn)
	}
}

func TestDrawFadeEnterWithOpacityCanvas(t *testing.T) {
	child := newMockWidget()
	dur := 200 * time.Millisecond
	tr := Wrap(child, EnterEffect(FadeIn()), Duration(dur), Easing(animation.Linear))
	tr.shown = false

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx := newMockContext(now)
	canvas := &opacityCanvas{}

	// Start show.
	tr.Show()

	// First draw: initializes start time, progress = 0.
	tr.Draw(ctx, canvas)
	if child.drawn != 1 {
		t.Errorf("draw count: got %d, want 1", child.drawn)
	}

	// Advance halfway.
	ctx.now = now.Add(100 * time.Millisecond)
	tr.Draw(ctx, canvas)
	if child.drawn != 2 {
		t.Errorf("draw count: got %d, want 2", child.drawn)
	}
	if !tr.IsAnimating() {
		t.Error("should still be animating at 50%")
	}

	// Advance past duration.
	ctx.now = now.Add(300 * time.Millisecond)
	tr.Draw(ctx, canvas)
	if tr.IsAnimating() {
		t.Error("should stop animating after duration")
	}
	if !tr.IsShown() {
		t.Error("should be shown after enter completes")
	}
}

func TestDrawFadeExitHidesAfterComplete(t *testing.T) {
	child := newMockWidget()
	dur := 100 * time.Millisecond
	tr := Wrap(child, ExitEffect(FadeOut()), Duration(dur), Easing(animation.Linear))

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx := newMockContext(now)
	canvas := &opacityCanvas{}

	tr.Hide()
	// Still "shown" during exit animation; we check after draw completes below.

	// First draw.
	tr.Draw(ctx, canvas)

	// Advance past duration.
	ctx.now = now.Add(200 * time.Millisecond)
	tr.Draw(ctx, canvas)
	if tr.IsAnimating() {
		t.Error("should stop animating after exit duration")
	}
	if tr.IsShown() {
		t.Error("should be hidden after exit completes")
	}

	// Further draws should not draw child.
	child.drawn = 0
	tr.Draw(ctx, canvas)
	if child.drawn != 0 {
		t.Error("hidden transition should not draw child")
	}
}

func TestDrawSlideEnter(t *testing.T) {
	child := newMockWidget()
	dur := 200 * time.Millisecond
	tr := Wrap(child, EnterEffect(SlideIn(FromLeft)), Duration(dur), Easing(animation.Linear))
	tr.shown = false

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx := newMockContext(now)
	canvas := &mockCanvas{}

	// Layout first to set child size.
	tr.SetBounds(geometry.NewRect(0, 0, 200, 200))
	tr.Layout(ctx, geometry.Constraints{MaxWidth: 200, MaxHeight: 200})

	tr.Show()

	// Draw at t=0: full offset.
	tr.Draw(ctx, canvas)
	if child.drawn != 1 {
		t.Errorf("draw count: got %d, want 1", child.drawn)
	}
	// Transform should have been pushed and popped (net zero in canvas state).
	if len(canvas.transforms) != 0 {
		t.Errorf("transforms should be popped after draw, got %d", len(canvas.transforms))
	}
}

func TestDrawScaleEnter(t *testing.T) {
	child := newMockWidget()
	dur := 200 * time.Millisecond
	tr := Wrap(child, EnterEffect(ScaleIn()), Duration(dur), Easing(animation.Linear))
	tr.shown = false

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx := newMockContext(now)
	canvas := &opacityCanvas{}

	tr.SetBounds(geometry.NewRect(0, 0, 200, 200))
	tr.Layout(ctx, geometry.Constraints{MaxWidth: 200, MaxHeight: 200})

	tr.Show()
	tr.Draw(ctx, canvas)

	if child.drawn != 1 {
		t.Errorf("draw count: got %d, want 1", child.drawn)
	}
}

func TestDrawWithCanvasWithoutOpacity(t *testing.T) {
	child := newMockWidget()
	dur := 100 * time.Millisecond
	tr := Wrap(child, EnterEffect(FadeIn()), Duration(dur), Easing(animation.Linear))
	tr.shown = false

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx := newMockContext(now)
	canvas := &mockCanvas{} // no OpacityPusher

	tr.Show()
	tr.Draw(ctx, canvas)

	// Should still draw child even without opacity support.
	if child.drawn != 1 {
		t.Errorf("should draw child even without opacity canvas, got %d draws", child.drawn)
	}
}

// --- Children tests ---

func TestChildrenReturnsChild(t *testing.T) {
	child := newMockWidget()
	tr := Wrap(child)
	children := tr.Children()
	if len(children) != 1 || children[0] != child {
		t.Error("Children should return single-element slice with child")
	}
}

func TestChildrenNilChild(t *testing.T) {
	tr := Wrap(nil)
	if tr.Children() != nil {
		t.Error("Children should return nil for nil child")
	}
}

// --- Event tests ---

func TestEventForwardsWhenShown(t *testing.T) {
	child := newMockWidget()
	child.eventRet = true
	tr := Wrap(child)

	ctx := newMockContext(time.Now())
	me := &event.MouseEvent{}
	if !tr.Event(ctx, me) {
		t.Error("Event should return child's result when shown")
	}
}

func TestEventBlockedWhenHidden(t *testing.T) {
	child := newMockWidget()
	child.eventRet = true
	tr := Wrap(child)
	tr.shown = false

	ctx := newMockContext(time.Now())
	me := &event.MouseEvent{}
	if tr.Event(ctx, me) {
		t.Error("Event should return false when hidden")
	}
}

func TestEventNilChild(t *testing.T) {
	tr := Wrap(nil)
	ctx := newMockContext(time.Now())
	if tr.Event(ctx, &event.MouseEvent{}) {
		t.Error("Event should return false for nil child")
	}
}

// --- Animation completion tests ---

func TestAnimationInvalidatesOnProgress(t *testing.T) {
	child := newMockWidget()
	dur := 200 * time.Millisecond
	tr := Wrap(child, EnterEffect(FadeIn()), Duration(dur))
	tr.shown = false

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx := newMockContext(now)
	canvas := &mockCanvas{}

	tr.Show()
	tr.Draw(ctx, canvas)

	// Mid-animation should have invalidated.
	ctx.now = now.Add(50 * time.Millisecond)
	ctx.invalidated = false
	tr.Draw(ctx, canvas)
	if !ctx.invalidated {
		t.Error("should invalidate during animation")
	}
}

func TestAnimationSetsNeedsRedraw(t *testing.T) {
	child := newMockWidget()
	dur := 200 * time.Millisecond
	tr := Wrap(child, EnterEffect(FadeIn()), Duration(dur))
	tr.shown = false

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx := newMockContext(now)
	canvas := &mockCanvas{}

	tr.Show()
	tr.SetNeedsRedraw(false)
	tr.Draw(ctx, canvas)

	// After first draw (still animating), should set needs redraw.
	// Progress is 0 on first draw since start == now.
	if !tr.NeedsRedraw() {
		// First draw at t=0 sets progress to 0, which is < 1, so it sets NeedsRedraw.
		// Actually at t=0, elapsed=0, progress=0, which is < 1.0, so NeedsRedraw should be true.
		t.Error("should set NeedsRedraw during animation")
	}
}

// --- scaleBoundsFromCenter test ---

func TestScaleBoundsFromCenter(t *testing.T) {
	bounds := geometry.NewRect(10, 20, 100, 80) // Min(10,20) Max(110,100)
	scaled := scaleBoundsFromCenter(bounds, 0.5)

	// Center: (60, 60), half-width: 25, half-height: 20
	wantMinX := float32(60 - 25)
	wantMinY := float32(60 - 20)
	wantMaxX := float32(60 + 25)
	wantMaxY := float32(60 + 20)

	if scaled.Min.X != wantMinX || scaled.Min.Y != wantMinY ||
		scaled.Max.X != wantMaxX || scaled.Max.Y != wantMaxY {
		t.Errorf("scaleBoundsFromCenter: got %v, want (%v,%v)-(%v,%v)",
			scaled, wantMinX, wantMinY, wantMaxX, wantMaxY)
	}
}

// --- computeTranslation test ---

func TestComputeTranslationEnter(t *testing.T) {
	child := newMockWidget()
	tr := Wrap(child, EnterEffect(SlideIn(FromLeft)))
	tr.entering = true

	bounds := geometry.NewRect(0, 0, 100, 50)
	dx, dy := tr.computeTranslation(SlideIn(FromLeft), bounds, 0.5)

	// FromLeft: TranslateXFraction = -1.0
	// Enter: lerp(-1.0*100, 0, 0.5) = lerp(-100, 0, 0.5) = -50
	if dx != -50 {
		t.Errorf("dx: got %v, want -50", dx)
	}
	if dy != 0 {
		t.Errorf("dy: got %v, want 0", dy)
	}
}

func TestComputeTranslationExit(t *testing.T) {
	child := newMockWidget()
	tr := Wrap(child, ExitEffect(SlideOut(ToRight)))
	tr.entering = false

	bounds := geometry.NewRect(0, 0, 100, 50)
	dx, dy := tr.computeTranslation(SlideOut(ToRight), bounds, 0.5)

	// ToRight: TranslateXFraction = 1.0
	// Exit: lerp(0, 1.0*100, 0.5) = 50
	if dx != 50 {
		t.Errorf("dx: got %v, want 50", dx)
	}
	if dy != 0 {
		t.Errorf("dy: got %v, want 0", dy)
	}
}

// --- Widget interface verification ---

func TestTransitionImplementsWidget(t *testing.T) {
	var _ widget.Widget = (*Transition)(nil)
}

// --- setChildBounds / childBoundsOf helpers ---

func TestSetChildBoundsTypeAssertion(t *testing.T) {
	child := newMockWidget()
	bounds := geometry.NewRect(10, 20, 100, 50)
	setChildBounds(child, bounds)

	got := child.Bounds()
	if got != bounds {
		t.Errorf("child bounds: got %v, want %v", got, bounds)
	}
}

func TestChildBoundsOfWithWidget(t *testing.T) {
	child := newMockWidget()
	bounds := geometry.NewRect(10, 20, 100, 50)
	child.SetBounds(bounds)

	got := childBoundsOf(child, geometry.Sz(100, 50), geometry.Pt(0, 0))
	if got != bounds {
		t.Errorf("childBoundsOf: got %v, want %v", got, bounds)
	}
}

// --- Direction constants test ---

func TestDirectionConstants(t *testing.T) {
	// Verify all direction constants are distinct.
	dirs := []Direction{FromTop, FromBottom, FromLeft, FromRight, ToTop, ToBottom, ToLeft, ToRight}
	seen := make(map[Direction]bool)
	for _, d := range dirs {
		if seen[d] {
			t.Errorf("duplicate direction constant: %d", d)
		}
		seen[d] = true
	}
}

// --- Show/Hide interrupt test ---

func TestShowInterruptsExit(t *testing.T) {
	child := newMockWidget()
	tr := Wrap(child,
		EnterEffect(FadeIn()),
		ExitEffect(FadeOut()),
		Duration(200*time.Millisecond),
	)

	tr.Hide()
	if !tr.IsAnimating() {
		t.Error("should be animating exit")
	}

	tr.Show()
	if !tr.IsAnimating() {
		t.Error("should be animating enter after interrupt")
	}
	if !tr.entering {
		t.Error("should be entering after Show interrupts Hide")
	}
}

func TestHideInterruptsEnter(t *testing.T) {
	child := newMockWidget()
	tr := Wrap(child,
		EnterEffect(FadeIn()),
		ExitEffect(FadeOut()),
		Duration(200*time.Millisecond),
	)
	tr.shown = false

	tr.Show()
	if !tr.entering {
		t.Error("should be entering")
	}

	tr.Hide()
	if !tr.IsAnimating() {
		t.Error("should be animating exit after interrupt")
	}
	if tr.entering {
		t.Error("should be exiting after Hide interrupts Show")
	}
}
