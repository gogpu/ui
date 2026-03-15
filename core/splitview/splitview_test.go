package splitview_test

import (
	"image"
	"testing"
	"time"

	"github.com/gogpu/ui/core/splitview"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// --- Test Helpers ---

// mockWidget is a minimal widget for testing.
type mockWidget struct {
	widget.WidgetBase
	layoutCalled bool
	drawCalled   bool
	eventResult  bool
	lastEvent    event.Event
}

func newMockWidget() *mockWidget {
	w := &mockWidget{}
	w.SetVisible(true)
	w.SetEnabled(true)
	return w
}

func (m *mockWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	m.layoutCalled = true
	return c.Biggest()
}

func (m *mockWidget) Draw(_ widget.Context, _ widget.Canvas) {
	m.drawCalled = true
}

func (m *mockWidget) Event(_ widget.Context, e event.Event) bool {
	m.lastEvent = e
	return m.eventResult
}

// canvasRecorder implements widget.Canvas for testing draw calls.
type canvasRecorder struct {
	drawRectCount   int
	drawCircleCount int
}

func (c *canvasRecorder) Clear(_ widget.Color)                                                  {}
func (c *canvasRecorder) DrawRect(_ geometry.Rect, _ widget.Color)                              { c.drawRectCount++ }
func (c *canvasRecorder) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32)                 {}
func (c *canvasRecorder) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32)              {}
func (c *canvasRecorder) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {}
func (c *canvasRecorder) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)                { c.drawCircleCount++ }
func (c *canvasRecorder) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32)   {}
func (c *canvasRecorder) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32)               {}
func (c *canvasRecorder) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
}

func (c *canvasRecorder) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}
func (c *canvasRecorder) DrawImage(_ image.Image, _ geometry.Point)    {}
func (c *canvasRecorder) PushClip(_ geometry.Rect)                     {}
func (c *canvasRecorder) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *canvasRecorder) PopClip()                                     {}
func (c *canvasRecorder) PushTransform(_ geometry.Point)               {}
func (c *canvasRecorder) PopTransform()                                {}
func (c *canvasRecorder) TransformOffset() geometry.Point              { return geometry.Point{} }

// Compile-time check.
var _ widget.Canvas = (*canvasRecorder)(nil)

func setupSplit(opts ...splitview.Option) (*splitview.Widget, *widget.ContextImpl) {
	sv := splitview.New(opts...)
	ctx := widget.NewContext()
	bounds := geometry.NewRect(0, 0, 600, 400)
	sv.SetBounds(bounds)
	_ = sv.Layout(ctx, geometry.Constraints{
		MinWidth: 600, MaxWidth: 600,
		MinHeight: 400, MaxHeight: 400,
	})
	return sv, ctx
}

// --- Tests ---

func TestNew_Defaults(t *testing.T) {
	sv := splitview.New()

	if !sv.IsVisible() {
		t.Error("expected visible by default")
	}
	if !sv.IsEnabled() {
		t.Error("expected enabled by default")
	}
	if sv.Ratio() != 0.5 {
		t.Errorf("expected default ratio 0.5, got %f", sv.Ratio())
	}
	if sv.IsCollapsed() {
		t.Error("expected not collapsed by default")
	}
	if sv.FirstPanel() != nil {
		t.Error("expected nil first panel by default")
	}
	if sv.SecondPanel() != nil {
		t.Error("expected nil second panel by default")
	}
	if children := sv.Children(); len(children) != 0 {
		t.Errorf("expected 0 children, got %d", len(children))
	}
}

func TestNew_WithOptions(t *testing.T) {
	first := newMockWidget()
	second := newMockWidget()

	sv := splitview.New(
		splitview.First(first),
		splitview.Second(second),
		splitview.OrientationOpt(splitview.Vertical),
		splitview.InitialRatio(0.3),
		splitview.MinFirst(100),
		splitview.MinSecond(50),
		splitview.DividerWidth(8),
		splitview.CollapsibleOpt(true),
	)

	if sv.FirstPanel() != first {
		t.Error("expected first panel to be set")
	}
	if sv.SecondPanel() != second {
		t.Error("expected second panel to be set")
	}
	if sv.Ratio() != 0.3 {
		t.Errorf("expected ratio 0.3, got %f", sv.Ratio())
	}
}

func TestNew_WithPainter(t *testing.T) {
	painter := splitview.DefaultPainter{}
	sv := splitview.New(splitview.PainterOpt(painter))
	// Just ensure it doesn't panic.
	if sv == nil {
		t.Error("expected non-nil widget")
	}
}

func TestNew_WithColorScheme(t *testing.T) {
	cs := splitview.DividerColorScheme{
		Divider:      widget.RGBA(1, 0, 0, 1),
		DividerHover: widget.RGBA(0, 1, 0, 1),
		DividerDrag:  widget.RGBA(0, 0, 1, 1),
		Handle:       widget.RGBA(1, 1, 0, 1),
	}
	sv := splitview.New(splitview.ColorSchemeOpt(cs))
	if sv == nil {
		t.Error("expected non-nil widget")
	}
}

func TestChildren(t *testing.T) {
	tests := []struct {
		name    string
		first   widget.Widget
		second  widget.Widget
		wantLen int
	}{
		{"no panels", nil, nil, 0},
		{"first only", newMockWidget(), nil, 1},
		{"second only", nil, newMockWidget(), 1},
		{"both panels", newMockWidget(), newMockWidget(), 2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var opts []splitview.Option
			if tc.first != nil {
				opts = append(opts, splitview.First(tc.first))
			}
			if tc.second != nil {
				opts = append(opts, splitview.Second(tc.second))
			}
			sv := splitview.New(opts...)
			if got := len(sv.Children()); got != tc.wantLen {
				t.Errorf("Children() len = %d, want %d", got, tc.wantLen)
			}
		})
	}
}

func TestLayout_Horizontal(t *testing.T) {
	first := newMockWidget()
	second := newMockWidget()
	sv, _ := setupSplit(
		splitview.First(first),
		splitview.Second(second),
		splitview.InitialRatio(0.5),
	)

	if !first.layoutCalled {
		t.Error("expected first panel Layout to be called")
	}
	if !second.layoutCalled {
		t.Error("expected second panel Layout to be called")
	}

	// Default divider width = 6. Total space = 600 - 6 = 594.
	// ratio 0.5 -> first gets 297, second gets 297.
	firstBounds := first.Bounds()
	if firstBounds.Width() < 290 || firstBounds.Width() > 300 {
		t.Errorf("first width = %f, expected ~297", firstBounds.Width())
	}

	secondBounds := second.Bounds()
	if secondBounds.Width() < 290 || secondBounds.Width() > 300 {
		t.Errorf("second width = %f, expected ~297", secondBounds.Width())
	}

	_ = sv // keep reference
}

func TestLayout_Vertical(t *testing.T) {
	first := newMockWidget()
	second := newMockWidget()

	sv, ctx := setupSplit(
		splitview.First(first),
		splitview.Second(second),
		splitview.OrientationOpt(splitview.Vertical),
		splitview.InitialRatio(0.25),
	)

	_ = sv
	_ = ctx

	firstBounds := first.Bounds()
	// Total height = 400 - 6 (divider) = 394. ratio 0.25 -> ~98.5
	if firstBounds.Height() < 95 || firstBounds.Height() > 103 {
		t.Errorf("first height = %f, expected ~98.5", firstBounds.Height())
	}
}

func TestLayout_NilPanels(t *testing.T) {
	sv, ctx := setupSplit() // no panels
	size := sv.Layout(ctx, geometry.Constraints{
		MinWidth: 600, MaxWidth: 600,
		MinHeight: 400, MaxHeight: 400,
	})
	if size.Width != 600 || size.Height != 400 {
		t.Errorf("size = %v, expected (600, 400)", size)
	}
}

func TestLayout_ZeroConstraints(t *testing.T) {
	sv := splitview.New()
	ctx := widget.NewContext()
	size := sv.Layout(ctx, geometry.Constraints{
		MinWidth: 0, MaxWidth: 0,
		MinHeight: 0, MaxHeight: 0,
	})
	// Should use defaults.
	if size.Width <= 0 || size.Height <= 0 {
		t.Errorf("expected positive fallback size, got %v", size)
	}
}

func TestDraw_BothPanels(t *testing.T) {
	first := newMockWidget()
	second := newMockWidget()
	sv, ctx := setupSplit(
		splitview.First(first),
		splitview.Second(second),
	)

	canvas := &canvasRecorder{}
	sv.Draw(ctx, canvas)

	if !first.drawCalled {
		t.Error("expected first panel Draw to be called")
	}
	if !second.drawCalled {
		t.Error("expected second panel Draw to be called")
	}
	// Divider draws: 1 DrawRect (bg) + 3 DrawCircle (handle dots).
	if canvas.drawRectCount < 1 {
		t.Error("expected at least 1 DrawRect for divider")
	}
	if canvas.drawCircleCount < 3 {
		t.Errorf("expected 3 handle dots, got %d DrawCircle calls", canvas.drawCircleCount)
	}
}

func TestDraw_EmptyBounds(t *testing.T) {
	sv := splitview.New()
	ctx := widget.NewContext()
	canvas := &canvasRecorder{}
	// Don't set bounds -- default is empty.
	sv.Draw(ctx, canvas)
	// Should not panic, should not draw anything.
	if canvas.drawRectCount != 0 {
		t.Errorf("expected no draws on empty bounds, got %d", canvas.drawRectCount)
	}
}

func TestDivider_Drag_Horizontal(t *testing.T) {
	sv, ctx := setupSplit(
		splitview.First(newMockWidget()),
		splitview.Second(newMockWidget()),
		splitview.InitialRatio(0.5),
	)

	// Divider center at x = 297 (approx). Press on divider.
	press := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(300, 200),
		Buttons:   event.ButtonStateLeft,
	}
	if !sv.Event(ctx, press) {
		t.Error("expected press on divider to be consumed")
	}

	// Drag to the right by 60 pixels.
	move := &event.MouseEvent{
		MouseType: event.MouseMove,
		Position:  geometry.Pt(360, 200),
		Buttons:   event.ButtonStateLeft,
	}
	if !sv.Event(ctx, move) {
		t.Error("expected drag move to be consumed")
	}

	// Ratio should have increased.
	if sv.Ratio() <= 0.5 {
		t.Errorf("expected ratio > 0.5 after drag right, got %f", sv.Ratio())
	}

	// Release.
	release := &event.MouseEvent{
		MouseType: event.MouseRelease,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(360, 200),
	}
	if !sv.Event(ctx, release) {
		t.Error("expected release to be consumed after drag")
	}
}

func TestDivider_Drag_Vertical(t *testing.T) {
	sv, ctx := setupSplit(
		splitview.First(newMockWidget()),
		splitview.Second(newMockWidget()),
		splitview.OrientationOpt(splitview.Vertical),
		splitview.InitialRatio(0.5),
	)

	// Divider center at y = 197 (approx). Press on divider.
	press := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(300, 199),
		Buttons:   event.ButtonStateLeft,
	}
	if !sv.Event(ctx, press) {
		t.Error("expected press on divider to be consumed")
	}

	// Drag down by 40 pixels.
	move := &event.MouseEvent{
		MouseType: event.MouseMove,
		Position:  geometry.Pt(300, 239),
		Buttons:   event.ButtonStateLeft,
	}
	if !sv.Event(ctx, move) {
		t.Error("expected drag move to be consumed")
	}

	if sv.Ratio() <= 0.5 {
		t.Errorf("expected ratio > 0.5 after drag down, got %f", sv.Ratio())
	}

	// Release.
	release := &event.MouseEvent{
		MouseType: event.MouseRelease,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(300, 239),
	}
	sv.Event(ctx, release)
}

func TestDivider_Drag_MinConstraints(t *testing.T) {
	sv, ctx := setupSplit(
		splitview.First(newMockWidget()),
		splitview.Second(newMockWidget()),
		splitview.InitialRatio(0.5),
		splitview.MinFirst(200),
		splitview.MinSecond(200),
	)

	// Press on divider.
	press := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(300, 200),
		Buttons:   event.ButtonStateLeft,
	}
	sv.Event(ctx, press)

	// Drag far left -- should be clamped by MinFirst.
	move := &event.MouseEvent{
		MouseType: event.MouseMove,
		Position:  geometry.Pt(50, 200),
		Buttons:   event.ButtonStateLeft,
	}
	sv.Event(ctx, move)

	// With total space = 594, min first = 200 => min ratio ~0.337.
	if sv.Ratio() < 0.33 {
		t.Errorf("expected ratio >= ~0.337 due to MinFirst, got %f", sv.Ratio())
	}

	// Drag far right -- should be clamped by MinSecond.
	move2 := &event.MouseEvent{
		MouseType: event.MouseMove,
		Position:  geometry.Pt(550, 200),
		Buttons:   event.ButtonStateLeft,
	}
	sv.Event(ctx, move2)

	// max ratio = 1 - 200/594 ~= 0.663.
	if sv.Ratio() > 0.67 {
		t.Errorf("expected ratio <= ~0.663 due to MinSecond, got %f", sv.Ratio())
	}

	release := &event.MouseEvent{
		MouseType: event.MouseRelease,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(550, 200),
	}
	sv.Event(ctx, release)
}

func TestDivider_NoConsume_OutsideDivider(t *testing.T) {
	sv, ctx := setupSplit(
		splitview.First(newMockWidget()),
		splitview.Second(newMockWidget()),
	)

	// Click outside divider area.
	press := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(100, 200),
		Buttons:   event.ButtonStateLeft,
	}
	if sv.Event(ctx, press) {
		t.Error("expected press outside divider to not be consumed by divider handler")
	}
}

func TestDivider_RightButton_Ignored(t *testing.T) {
	sv, ctx := setupSplit(
		splitview.First(newMockWidget()),
		splitview.Second(newMockWidget()),
	)

	press := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonRight,
		Position:  geometry.Pt(300, 200),
	}
	if sv.Event(ctx, press) {
		t.Error("expected right button press to not be consumed")
	}
}

func TestCollapse_DoubleClick(t *testing.T) {
	sv, ctx := setupSplit(
		splitview.First(newMockWidget()),
		splitview.Second(newMockWidget()),
		splitview.InitialRatio(0.5),
		splitview.CollapsibleOpt(true),
	)

	divPos := geometry.Pt(300, 200)

	// First click.
	click1 := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  divPos,
		Buttons:   event.ButtonStateLeft,
	}
	sv.Event(ctx, click1)
	release1 := &event.MouseEvent{
		MouseType: event.MouseRelease,
		Button:    event.ButtonLeft,
		Position:  divPos,
	}
	sv.Event(ctx, release1)

	// Advance time slightly.
	ctx.SetNow(ctx.Now().Add(100 * time.Millisecond))

	// Re-layout so divider rect is current.
	sv.Layout(ctx, geometry.Constraints{
		MinWidth: 600, MaxWidth: 600,
		MinHeight: 400, MaxHeight: 400,
	})

	// Second click (double-click). Must hit the divider again.
	click2 := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  divPos,
		Buttons:   event.ButtonStateLeft,
	}
	consumed := sv.Event(ctx, click2)
	if !consumed {
		t.Error("expected double-click to be consumed")
	}

	if !sv.IsCollapsed() {
		t.Error("expected first panel to be collapsed after double-click")
	}

	// Re-layout so divider rect reflects collapsed state (divider at x=0).
	sv.Layout(ctx, geometry.Constraints{
		MinWidth: 600, MaxWidth: 600,
		MinHeight: 400, MaxHeight: 400,
	})

	// Advance time again.
	ctx.SetNow(ctx.Now().Add(100 * time.Millisecond))

	// Double-click again to restore.
	// Divider is now at x=0 since ratio=0.
	newDivPos := geometry.Pt(3, 200) // center of divider at x=0, width=6

	click3 := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  newDivPos,
		Buttons:   event.ButtonStateLeft,
	}
	sv.Event(ctx, click3)
	release3 := &event.MouseEvent{
		MouseType: event.MouseRelease,
		Button:    event.ButtonLeft,
		Position:  newDivPos,
	}
	sv.Event(ctx, release3)

	ctx.SetNow(ctx.Now().Add(100 * time.Millisecond))
	click4 := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  newDivPos,
		Buttons:   event.ButtonStateLeft,
	}
	sv.Event(ctx, click4)

	if sv.IsCollapsed() {
		t.Error("expected first panel to be restored after second double-click")
	}

	if sv.Ratio() < 0.45 || sv.Ratio() > 0.55 {
		t.Errorf("expected ratio ~0.5 after restore, got %f", sv.Ratio())
	}
}

func TestCollapse_Disabled(t *testing.T) {
	sv, ctx := setupSplit(
		splitview.First(newMockWidget()),
		splitview.Second(newMockWidget()),
		splitview.InitialRatio(0.5),
		// CollapsibleOpt not set (default false).
	)

	divPos := geometry.Pt(300, 200)

	// Double-click.
	click1 := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  divPos,
		Buttons:   event.ButtonStateLeft,
	}
	sv.Event(ctx, click1)
	sv.Event(ctx, &event.MouseEvent{
		MouseType: event.MouseRelease,
		Button:    event.ButtonLeft,
		Position:  divPos,
	})

	ctx.SetNow(ctx.Now().Add(100 * time.Millisecond))

	click2 := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  divPos,
		Buttons:   event.ButtonStateLeft,
	}
	sv.Event(ctx, click2)

	if sv.IsCollapsed() {
		t.Error("expected collapse to be disabled when CollapsibleOpt is not set")
	}
}

func TestSignal_TwoWayBinding(t *testing.T) {
	sig := state.NewSignal[float32](0.5)

	sv, ctx := setupSplit(
		splitview.First(newMockWidget()),
		splitview.Second(newMockWidget()),
		splitview.RatioSignal(sig),
	)

	// Ratio should come from signal.
	if sv.Ratio() != 0.5 {
		t.Errorf("expected ratio 0.5 from signal, got %f", sv.Ratio())
	}

	// Change signal externally.
	sig.Set(0.7)
	if sv.Ratio() != 0.7 {
		t.Errorf("expected ratio 0.7 after signal update, got %f", sv.Ratio())
	}

	// Drag divider -- should write back to signal.
	// Re-layout with new ratio.
	sv.Layout(ctx, geometry.Constraints{
		MinWidth: 600, MaxWidth: 600,
		MinHeight: 400, MaxHeight: 400,
	})

	// Press on divider at new position.
	// ratio 0.7 -> first = 594*0.7 = 415.8, divider at ~416.
	press := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(418, 200),
		Buttons:   event.ButtonStateLeft,
	}
	sv.Event(ctx, press)

	// Drag left.
	move := &event.MouseEvent{
		MouseType: event.MouseMove,
		Position:  geometry.Pt(358, 200),
		Buttons:   event.ButtonStateLeft,
	}
	sv.Event(ctx, move)

	// Signal should have been updated.
	if sig.Get() >= 0.7 {
		t.Errorf("expected signal < 0.7 after drag left, got %f", sig.Get())
	}

	release := &event.MouseEvent{
		MouseType: event.MouseRelease,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(358, 200),
	}
	sv.Event(ctx, release)
}

func TestSignal_ReadonlyBinding(t *testing.T) {
	readonlySig := state.NewComputed(func() float32 { return 0.3 })

	sv := splitview.New(
		splitview.First(newMockWidget()),
		splitview.Second(newMockWidget()),
		splitview.RatioReadonlySignal(readonlySig),
		splitview.InitialRatio(0.5), // should be overridden by signal
	)

	if sv.Ratio() != 0.3 {
		t.Errorf("expected ratio 0.3 from readonly signal, got %f", sv.Ratio())
	}
}

func TestOnRatioChange_Callback(t *testing.T) {
	var callbackRatio float32
	var callbackCalled bool

	sv, ctx := setupSplit(
		splitview.First(newMockWidget()),
		splitview.Second(newMockWidget()),
		splitview.InitialRatio(0.5),
		splitview.OnRatioChange(func(r float32) {
			callbackRatio = r
			callbackCalled = true
		}),
	)

	// Drag divider.
	press := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(300, 200),
		Buttons:   event.ButtonStateLeft,
	}
	sv.Event(ctx, press)

	move := &event.MouseEvent{
		MouseType: event.MouseMove,
		Position:  geometry.Pt(350, 200),
		Buttons:   event.ButtonStateLeft,
	}
	sv.Event(ctx, move)

	if !callbackCalled {
		t.Error("expected OnRatioChange callback to be called")
	}
	if callbackRatio <= 0.5 {
		t.Errorf("expected callback ratio > 0.5, got %f", callbackRatio)
	}

	release := &event.MouseEvent{
		MouseType: event.MouseRelease,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(350, 200),
	}
	sv.Event(ctx, release)
}

func TestInitialRatio_Clamping(t *testing.T) {
	tests := []struct {
		name      string
		input     float32
		wantRatio float32
	}{
		{"below zero", -0.5, 0},
		{"zero", 0, 0},
		{"normal", 0.3, 0.3},
		{"one", 1.0, 1.0},
		{"above one", 1.5, 1.0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sv := splitview.New(splitview.InitialRatio(tc.input))
			if sv.Ratio() != tc.wantRatio {
				t.Errorf("Ratio() = %f, want %f", sv.Ratio(), tc.wantRatio)
			}
		})
	}
}

func TestOrientation_String(t *testing.T) {
	tests := []struct {
		o    splitview.Orientation
		want string
	}{
		{splitview.Horizontal, "Horizontal"},
		{splitview.Vertical, "Vertical"},
		{splitview.Orientation(99), "Unknown"},
	}

	for _, tc := range tests {
		if got := tc.o.String(); got != tc.want {
			t.Errorf("Orientation(%d).String() = %q, want %q", tc.o, got, tc.want)
		}
	}
}

func TestMount_Unmount(t *testing.T) {
	sig := state.NewSignal[float32](0.5)
	sched := state.NewScheduler(func(_ []widget.Widget) {})

	sv := splitview.New(
		splitview.RatioSignal(sig),
	)

	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	// Mount should create bindings.
	sv.Mount(ctx)

	// Unmount should be a no-op (bindings cleaned up by WidgetBase).
	sv.Unmount()
}

func TestMount_ReadonlySignal(t *testing.T) {
	readonlySig := state.NewComputed(func() float32 { return 0.5 })
	sched := state.NewScheduler(func(_ []widget.Widget) {})

	sv := splitview.New(
		splitview.RatioReadonlySignal(readonlySig),
	)

	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	// Mount should create binding for readonly signal.
	sv.Mount(ctx)
	sv.Unmount()
}

func TestMount_NilScheduler(t *testing.T) {
	sv := splitview.New(
		splitview.RatioSignal(state.NewSignal[float32](0.5)),
	)

	ctx := widget.NewContext()
	// No scheduler set. Mount should not panic.
	sv.Mount(ctx)
}

func TestEvent_DispatchToChildren(t *testing.T) {
	first := newMockWidget()
	second := newMockWidget()
	first.eventResult = true // first consumes

	sv, ctx := setupSplit(
		splitview.First(first),
		splitview.Second(second),
	)

	// Click on first panel area.
	click := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(50, 200),
		Buttons:   event.ButtonStateLeft,
	}
	consumed := sv.Event(ctx, click)
	if !consumed {
		t.Error("expected event to be consumed by first panel")
	}
	if first.lastEvent != click {
		t.Error("expected first panel to receive the event")
	}
}

func TestEvent_SecondPanelReceives(t *testing.T) {
	first := newMockWidget()
	second := newMockWidget()
	second.eventResult = true

	sv, ctx := setupSplit(
		splitview.First(first),
		splitview.Second(second),
	)

	// Click on second panel area.
	click := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(500, 200),
		Buttons:   event.ButtonStateLeft,
	}
	consumed := sv.Event(ctx, click)
	if !consumed {
		t.Error("expected event to be consumed by second panel")
	}
}

func TestDrag_LostButton(t *testing.T) {
	sv, ctx := setupSplit(
		splitview.First(newMockWidget()),
		splitview.Second(newMockWidget()),
		splitview.InitialRatio(0.5),
	)

	// Start drag.
	press := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(300, 200),
		Buttons:   event.ButtonStateLeft,
	}
	sv.Event(ctx, press)

	// Move with button released (simulates lost MouseRelease).
	move := &event.MouseEvent{
		MouseType: event.MouseMove,
		Position:  geometry.Pt(350, 200),
		Buttons:   0, // no buttons pressed
	}
	consumed := sv.Event(ctx, move)

	// Should clear drag state and not consume.
	if consumed {
		t.Error("expected lost-button move to not be consumed")
	}
}

func TestDivider_CursorChange(t *testing.T) {
	sv, ctx := setupSplit(
		splitview.First(newMockWidget()),
		splitview.Second(newMockWidget()),
		splitview.InitialRatio(0.5),
	)

	// Hover over divider.
	enter := &event.MouseEvent{
		MouseType: event.MouseMove,
		Position:  geometry.Pt(300, 200),
	}
	sv.Event(ctx, enter)

	if ctx.Cursor() != widget.CursorResizeEW {
		t.Errorf("expected ResizeEW cursor when hovering horizontal divider, got %s", ctx.Cursor())
	}

	// Move away from divider.
	leave := &event.MouseEvent{
		MouseType: event.MouseMove,
		Position:  geometry.Pt(100, 200),
	}
	sv.Event(ctx, leave)

	if ctx.Cursor() != widget.CursorDefault {
		t.Errorf("expected Default cursor after leaving divider, got %s", ctx.Cursor())
	}
}

func TestDivider_CursorChange_Vertical(t *testing.T) {
	sv, ctx := setupSplit(
		splitview.First(newMockWidget()),
		splitview.Second(newMockWidget()),
		splitview.OrientationOpt(splitview.Vertical),
		splitview.InitialRatio(0.5),
	)

	// Hover over divider (vertical: divider at y ~= 197).
	hover := &event.MouseEvent{
		MouseType: event.MouseMove,
		Position:  geometry.Pt(300, 199),
	}
	sv.Event(ctx, hover)

	if ctx.Cursor() != widget.CursorResizeNS {
		t.Errorf("expected ResizeNS cursor when hovering vertical divider, got %s", ctx.Cursor())
	}
}

func TestDefaultPainter_EmptyBounds(t *testing.T) {
	p := splitview.DefaultPainter{}
	canvas := &canvasRecorder{}

	ps := splitview.PaintState{
		DividerRect: geometry.Rect{}, // empty
	}
	p.PaintDivider(canvas, ps)

	if canvas.drawRectCount != 0 {
		t.Error("expected no draws on empty divider rect")
	}
}

func TestDefaultPainter_WithColorScheme(t *testing.T) {
	p := splitview.DefaultPainter{}
	canvas := &canvasRecorder{}

	cs := splitview.DividerColorScheme{
		Divider:      widget.RGBA(1, 0, 0, 1),
		DividerHover: widget.RGBA(0, 1, 0, 1),
		DividerDrag:  widget.RGBA(0, 0, 1, 1),
		Handle:       widget.RGBA(1, 1, 0, 1),
	}

	ps := splitview.PaintState{
		DividerRect: geometry.NewRect(100, 0, 6, 400),
		Orientation: splitview.Horizontal,
		ColorScheme: cs,
	}
	p.PaintDivider(canvas, ps)

	if canvas.drawRectCount != 1 {
		t.Errorf("expected 1 DrawRect, got %d", canvas.drawRectCount)
	}
	if canvas.drawCircleCount != 3 {
		t.Errorf("expected 3 DrawCircle for handle, got %d", canvas.drawCircleCount)
	}
}

func TestDefaultPainter_HoverState(t *testing.T) {
	p := splitview.DefaultPainter{}
	canvas := &canvasRecorder{}

	ps := splitview.PaintState{
		DividerRect: geometry.NewRect(100, 0, 6, 400),
		Orientation: splitview.Horizontal,
		Hovered:     true,
	}
	p.PaintDivider(canvas, ps)

	if canvas.drawRectCount != 1 {
		t.Errorf("expected 1 DrawRect for hover state, got %d", canvas.drawRectCount)
	}
}

func TestDefaultPainter_DragState(t *testing.T) {
	p := splitview.DefaultPainter{}
	canvas := &canvasRecorder{}

	ps := splitview.PaintState{
		DividerRect: geometry.NewRect(100, 0, 6, 400),
		Orientation: splitview.Vertical,
		Dragging:    true,
	}
	p.PaintDivider(canvas, ps)

	if canvas.drawRectCount != 1 {
		t.Errorf("expected 1 DrawRect for drag state, got %d", canvas.drawRectCount)
	}
}

func TestDefaultPainter_ColorScheme_States(t *testing.T) {
	cs := splitview.DividerColorScheme{
		Divider:      widget.RGBA(0.1, 0.1, 0.1, 1),
		DividerHover: widget.RGBA(0.2, 0.2, 0.2, 1),
		DividerDrag:  widget.RGBA(0.3, 0.3, 0.3, 1),
		Handle:       widget.RGBA(0.5, 0.5, 0.5, 1),
	}

	tests := []struct {
		name string
		ps   splitview.PaintState
	}{
		{"normal", splitview.PaintState{
			DividerRect: geometry.NewRect(0, 0, 6, 100),
			ColorScheme: cs,
		}},
		{"hovered", splitview.PaintState{
			DividerRect: geometry.NewRect(0, 0, 6, 100),
			ColorScheme: cs,
			Hovered:     true,
		}},
		{"dragging", splitview.PaintState{
			DividerRect: geometry.NewRect(0, 0, 6, 100),
			ColorScheme: cs,
			Dragging:    true,
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := splitview.DefaultPainter{}
			canvas := &canvasRecorder{}
			p.PaintDivider(canvas, tc.ps)
			if canvas.drawRectCount < 1 {
				t.Error("expected divider to be painted")
			}
		})
	}
}

func TestDividerWidth_Custom(t *testing.T) {
	sv, ctx := setupSplit(
		splitview.First(newMockWidget()),
		splitview.Second(newMockWidget()),
		splitview.DividerWidth(12),
		splitview.InitialRatio(0.5),
	)

	first := sv.FirstPanel().(*mockWidget)
	secondW := sv.SecondPanel().(*mockWidget)

	// Total space = 600 - 12 = 588. Each panel = 294.
	fb := first.Bounds()
	sb := secondW.Bounds()

	if fb.Width() < 290 || fb.Width() > 298 {
		t.Errorf("first width = %f, expected ~294 with divider=12", fb.Width())
	}
	if sb.Width() < 290 || sb.Width() > 298 {
		t.Errorf("second width = %f, expected ~294 with divider=12", sb.Width())
	}

	_ = ctx
}

func TestEvent_NonMouseEvent_Ignored(t *testing.T) {
	sv, ctx := setupSplit(
		splitview.First(newMockWidget()),
		splitview.Second(newMockWidget()),
	)

	// Key event should not be consumed by divider handler.
	key := &event.KeyEvent{
		KeyType: event.KeyPress,
		Key:     event.KeyA,
	}
	consumed := sv.Event(ctx, key)
	if consumed {
		t.Error("expected key event to not be consumed")
	}
}

func TestMouseEnter_Leave_Divider(t *testing.T) {
	sv, ctx := setupSplit(
		splitview.First(newMockWidget()),
		splitview.Second(newMockWidget()),
		splitview.InitialRatio(0.5),
	)

	// Enter on divider.
	enter := &event.MouseEvent{
		MouseType: event.MouseEnter,
		Position:  geometry.Pt(300, 200),
	}
	sv.Event(ctx, enter)

	// Leave.
	leave := &event.MouseEvent{
		MouseType: event.MouseLeave,
		Position:  geometry.Pt(100, 200),
	}
	sv.Event(ctx, leave)
	// Should not panic and should update state.
}

func TestRelease_WithoutDrag(t *testing.T) {
	sv, ctx := setupSplit(
		splitview.First(newMockWidget()),
		splitview.Second(newMockWidget()),
	)

	// Release without prior press.
	release := &event.MouseEvent{
		MouseType: event.MouseRelease,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(300, 200),
	}
	consumed := sv.Event(ctx, release)
	if consumed {
		t.Error("expected release without drag to not be consumed")
	}
}

func TestRelease_RightButton_Ignored(t *testing.T) {
	sv, ctx := setupSplit(
		splitview.First(newMockWidget()),
		splitview.Second(newMockWidget()),
	)

	// Start drag.
	press := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(300, 200),
		Buttons:   event.ButtonStateLeft,
	}
	sv.Event(ctx, press)

	// Release with right button.
	release := &event.MouseEvent{
		MouseType: event.MouseRelease,
		Button:    event.ButtonRight,
		Position:  geometry.Pt(300, 200),
	}
	consumed := sv.Event(ctx, release)
	if consumed {
		t.Error("expected right button release to not be consumed")
	}
}

func TestCompileTime_Interfaces(t *testing.T) {
	var _ widget.Widget = (*splitview.Widget)(nil)
	var _ widget.Lifecycle = (*splitview.Widget)(nil)
}

func TestDraw_NilPanels(t *testing.T) {
	sv, ctx := setupSplit() // no panels
	canvas := &canvasRecorder{}
	sv.Draw(ctx, canvas)
	// Should draw divider only.
	if canvas.drawRectCount < 1 {
		t.Error("expected divider to be drawn even without panels")
	}
}

func TestEvent_NilPanels(t *testing.T) {
	sv, ctx := setupSplit() // no panels
	click := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(300, 200),
		Buttons:   event.ButtonStateLeft,
	}
	// Should not panic.
	sv.Event(ctx, click)
}

func TestDrag_ZeroTotalSpace(t *testing.T) {
	sv := splitview.New(
		splitview.First(newMockWidget()),
		splitview.Second(newMockWidget()),
		splitview.InitialRatio(0.5),
	)
	ctx := widget.NewContext()
	// Layout with very small size.
	sv.SetBounds(geometry.NewRect(0, 0, 5, 5))
	_ = sv.Layout(ctx, geometry.Constraints{
		MinWidth: 5, MaxWidth: 5,
		MinHeight: 5, MaxHeight: 5,
	})

	// Try to drag -- should not panic even with zero total space.
	press := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(2, 2),
		Buttons:   event.ButtonStateLeft,
	}
	sv.Event(ctx, press)

	move := &event.MouseEvent{
		MouseType: event.MouseMove,
		Position:  geometry.Pt(4, 2),
		Buttons:   event.ButtonStateLeft,
	}
	sv.Event(ctx, move)
}
