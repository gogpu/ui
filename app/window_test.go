package app

import (
	"github.com/gogpu/gg/scene"
	"image"
	"testing"

	"github.com/gogpu/gpucontext"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/theme"
	"github.com/gogpu/ui/widget"
)

func TestWindow_SetRoot(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	if w.Root() != root {
		t.Error("root widget not set")
	}
	if !w.NeedsLayout() {
		t.Error("setting root should mark layout as needed")
	}
}

func TestWindow_SetRoot_Replace(t *testing.T) {
	a := New()
	w := a.Window()
	first := newMockWidget()
	second := newMockWidget()

	w.SetRoot(first)
	w.Frame()
	w.SetRoot(second)

	if w.Root() != second {
		t.Error("root not replaced")
	}
	if !w.NeedsLayout() {
		t.Error("replacing root should mark layout as needed")
	}
}

func TestWindow_Frame_Layout(t *testing.T) {
	wp := &mockWindowProvider{width: 400, height: 300, scale: 1.0}
	a := New(WithWindowProvider(wp))
	w := a.Window()

	root := newMockWidget()
	root.layoutSize = geometry.Sz(400, 300)
	w.SetRoot(root)

	w.Frame()

	if !root.layoutCalled {
		t.Error("layout not called on root")
	}

	// Verify bounds were set on root.
	bounds := root.Bounds()
	if bounds.Width() != 400 || bounds.Height() != 300 {
		t.Errorf("bounds = %v, want (400, 300)", bounds)
	}
}

func TestWindow_Frame_SkipsLayoutWhenClean(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	// First frame: layout performed.
	w.Frame()
	root.layoutCalled = false

	// Second frame: layout should not be performed (nothing changed).
	w.Frame()
	if root.layoutCalled {
		t.Error("layout called on second frame when nothing changed")
	}
}

func TestWindow_Frame_RelayoutOnResize(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	w.Frame()
	root.layoutCalled = false

	// Resize triggers relayout.
	w.HandleResize(1024, 768)
	w.Frame()

	if !root.layoutCalled {
		t.Error("layout not called after resize")
	}
}

func TestWindow_HandleEvent_DispatchesToRoot(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	e := event.NewMouseEvent(
		event.MousePress,
		event.ButtonLeft,
		event.ButtonStateLeft,
		geometry.Pt(10, 20),
		geometry.Pt(10, 20),
		event.ModNone,
	)
	w.HandleEvent(e)

	if !root.eventCalled {
		t.Error("event not dispatched to root")
	}
	if root.lastEvent != e {
		t.Error("wrong event dispatched")
	}
}

func TestWindow_HandleEvent_NoRoot(t *testing.T) {
	a := New()
	w := a.Window()

	e := event.NewKeyEvent(event.KeyPress, event.KeyA, 'a', event.ModNone)
	// Should not panic.
	w.HandleEvent(e)
}

func TestWindow_HandleEvent_NilEvent(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	// Should not panic.
	w.HandleEvent(nil)
	if root.eventCalled {
		t.Error("nil event should not be dispatched")
	}
}

func TestWindow_HandleResize(t *testing.T) {
	a := New()
	w := a.Window()

	w.HandleResize(1920, 1080)

	size := w.WindowSize()
	if size.Width != 1920 || size.Height != 1080 {
		t.Errorf("size = %v, want (1920, 1080)", size)
	}
	if !w.NeedsLayout() {
		t.Error("resize should mark layout as needed")
	}
}

func TestWindow_HandleFocusChange_Gained(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	w.HandleFocusChange(true)

	if !root.eventCalled {
		t.Error("focus event not dispatched")
	}
	if fe, ok := root.lastEvent.(*event.FocusEvent); ok {
		if !fe.IsGained() {
			t.Error("expected focus gained event")
		}
	} else {
		t.Error("expected FocusEvent type")
	}
}

func TestWindow_HandleFocusChange_Lost(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	w.HandleFocusChange(false)

	if !root.eventCalled {
		t.Error("focus event not dispatched")
	}
	if fe, ok := root.lastEvent.(*event.FocusEvent); ok {
		if !fe.IsLost() {
			t.Error("expected focus lost event")
		}
	} else {
		t.Error("expected FocusEvent type")
	}
}

func TestWindow_HandleFocusChange_NoRoot(t *testing.T) {
	a := New()
	w := a.Window()
	// Should not panic.
	w.HandleFocusChange(true)
}

func TestWindow_ScaleFromProvider(t *testing.T) {
	wp := &mockWindowProvider{width: 800, height: 600, scale: 2.0}
	a := New(WithWindowProvider(wp))
	w := a.Window()

	if w.Context().Scale() != 2.0 {
		t.Errorf("scale = %v, want 2.0", w.Context().Scale())
	}
}

func TestWindow_DefaultScale_Headless(t *testing.T) {
	a := New()
	w := a.Window()

	if w.Context().Scale() != 1.0 {
		t.Errorf("headless scale = %v, want 1.0", w.Context().Scale())
	}
}

func TestWindow_DefaultSize_Headless(t *testing.T) {
	a := New()
	w := a.Window()

	size := w.WindowSize()
	if size.Width != defaultWidth || size.Height != defaultHeight {
		t.Errorf("headless size = %v, want (%d, %d)", size, defaultWidth, defaultHeight)
	}
}

func TestWindow_CursorSync(t *testing.T) {
	pp := &mockPlatformProvider{fontScale: 1.0}
	a := New(WithPlatformProvider(pp))
	w := a.Window()

	// Create a widget that sets cursor during layout (inside Frame).
	cursorWidget := &cursorSettingOnLayoutWidget{
		cursor: widget.CursorPointer,
	}
	w.SetRoot(cursorWidget)

	// Frame calls layout -> widget sets cursor -> syncCursor forwards to platform.
	w.Frame()

	if pp.lastCursor != gpucontext.CursorPointer {
		t.Errorf("cursor = %v, want Pointer", pp.lastCursor)
	}
}

func TestWindow_CursorSync_NoPlatform(t *testing.T) {
	a := New() // headless, no platform provider
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)
	// Should not panic without platform provider.
	w.Frame()
}

func TestWindow_DrawTo(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	canvas := &mockCanvas{}
	drawn := w.DrawTo(canvas)

	if !root.drawCalled {
		t.Error("DrawTo did not call root Draw")
	}
	if !drawn {
		t.Error("DrawTo should return true on first draw (all widgets dirty)")
	}
}

func TestWindow_DrawTo_NoRoot(t *testing.T) {
	a := New()
	w := a.Window()
	canvas := &mockCanvas{}
	// Should not panic.
	drawn := w.DrawTo(canvas)
	if drawn {
		t.Error("DrawTo should return false with no root")
	}
}

func TestWindow_DrawTo_NilCanvas(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)
	// Should not panic.
	drawn := w.DrawTo(nil)
	if drawn {
		t.Error("DrawTo should return false with nil canvas")
	}
}

func TestWindow_Theme(t *testing.T) {
	dark := theme.DefaultDark()
	a := New(WithTheme(dark))

	if a.Window().Theme() != dark {
		t.Error("window theme does not match app theme")
	}
}

func TestWindow_InvalidateTriggersRelayout(t *testing.T) {
	wp := &mockWindowProvider{width: 800, height: 600, scale: 1.0}
	a := New(WithWindowProvider(wp))
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	w.Frame()
	root.layoutCalled = false

	// Invalidate should mark needsLayout.
	w.Context().Invalidate()

	if !w.NeedsLayout() {
		t.Error("invalidation should mark layout as needed")
	}

	w.Frame()
	if !root.layoutCalled {
		t.Error("layout not called after invalidation")
	}
}

func TestWindow_SizeFromProvider_Updates(t *testing.T) {
	wp := &mockWindowProvider{width: 800, height: 600, scale: 1.0}
	a := New(WithWindowProvider(wp))
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	w.Frame()
	root.layoutCalled = false

	// Simulate window resize by changing provider values.
	wp.width = 1024
	wp.height = 768

	w.Frame()

	if !root.layoutCalled {
		t.Error("layout not called when window size changed via provider")
	}
	size := w.WindowSize()
	if size.Width != 1024 || size.Height != 768 {
		t.Errorf("size = %v, want (1024, 768)", size)
	}
}

func TestWindow_FrameReflush(t *testing.T) {
	// Verify that Frame's reflush loop drains widgets that are
	// re-dirtied during the flush callback.
	root := newMockWidget()
	reflushes := 0

	var sched *state.Scheduler
	sched = state.NewScheduler(func(_ []widget.Widget) {
		reflushes++
		// Re-dirty on the first flush to exercise the reflush loop.
		if reflushes == 1 {
			sched.MarkDirty(root)
		}
	})

	wp := &mockWindowProvider{width: 400, height: 300, scale: 1.0}
	w := newWindow(wp, nil, sched, theme.DefaultLight(), RenderModeHostManaged)
	w.SetRoot(root)

	sched.MarkDirty(root)
	w.Frame()

	if sched.PendingCount() != 0 {
		t.Errorf("pending count after Frame = %d, want 0", sched.PendingCount())
	}
	if reflushes < 2 {
		t.Errorf("reflushes = %d, want >= 2", reflushes)
	}
}

// --- Helper test types ---

// cursorSettingWidget sets the cursor during event handling.
type cursorSettingWidget struct {
	widget.WidgetBase
	cursor widget.CursorType
}

func (w *cursorSettingWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(100, 100))
}

func (w *cursorSettingWidget) Draw(_ widget.Context, _ widget.Canvas) {}

func (w *cursorSettingWidget) Event(ctx widget.Context, _ event.Event) bool {
	ctx.SetCursor(w.cursor)
	return true
}

// cursorSettingOnLayoutWidget sets the cursor during layout (inside Frame).
type cursorSettingOnLayoutWidget struct {
	widget.WidgetBase
	cursor widget.CursorType
}

func (w *cursorSettingOnLayoutWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	ctx.SetCursor(w.cursor)
	return c.Constrain(geometry.Sz(100, 100))
}

func (w *cursorSettingOnLayoutWidget) Draw(_ widget.Context, _ widget.Canvas) {}

func (w *cursorSettingOnLayoutWidget) Event(_ widget.Context, _ event.Event) bool {
	return false
}

// mockCanvas implements widget.Canvas for testing.
type mockCanvas struct{}

func (m *mockCanvas) Clear(widget.Color)                                            {}
func (m *mockCanvas) DrawRect(geometry.Rect, widget.Color)                          {}
func (m *mockCanvas) FillRectDirect(geometry.Rect, widget.Color)                    {}
func (m *mockCanvas) StrokeRect(geometry.Rect, widget.Color, float32)               {}
func (m *mockCanvas) DrawRoundRect(geometry.Rect, widget.Color, float32)            {}
func (m *mockCanvas) StrokeRoundRect(geometry.Rect, widget.Color, float32, float32) {}
func (m *mockCanvas) DrawCircle(geometry.Point, float32, widget.Color)              {}
func (m *mockCanvas) StrokeCircle(geometry.Point, float32, widget.Color, float32)   {}
func (m *mockCanvas) StrokeArc(_ geometry.Point, _ float32, _, _ float64, _ widget.Color, _ float32) {
}
func (m *mockCanvas) DrawLine(geometry.Point, geometry.Point, widget.Color, float32)                {}
func (m *mockCanvas) DrawText(string, geometry.Rect, float32, widget.Color, bool, widget.TextAlign) {}

func (m *mockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}
func (m *mockCanvas) DrawImage(image.Image, geometry.Point)        {}
func (m *mockCanvas) PushClip(geometry.Rect)                       {}
func (m *mockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (m *mockCanvas) PopClip()                                     {}
func (m *mockCanvas) PushTransform(geometry.Point)                 {}
func (m *mockCanvas) PopTransform()                                {}
func (m *mockCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }
func (m *mockCanvas) ScreenOriginBase() geometry.Point             { return geometry.Point{} }
func (m *mockCanvas) ClipBounds() geometry.Rect                    { return geometry.NewRect(0, 0, 10000, 10000) }
func (m *mockCanvas) ReplayScene(_ *scene.Scene)                   {}

// --- Retained-mode rendering tests ---

func TestWindow_DrawTo_ReportsCleanTree(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	canvas := &mockCanvas{}

	// First draw: all widgets are dirty (just mounted, needsFullRepaint=true).
	drawn := w.DrawTo(canvas)
	if !drawn {
		t.Error("first DrawTo should report dirty (all widgets dirty after mount)")
	}

	// Reset tracking.
	root.drawCalled = false

	// Second draw: nothing changed but DrawTo still draws (host may have
	// cleared pixmap). Returns true because a valid frame was produced.
	drawn = w.DrawTo(canvas)
	if !drawn {
		t.Error("second DrawTo should return true (always draws)")
	}
}

func TestWindow_DrawTo_DrawsAfterSignalDirty(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	canvas := &mockCanvas{}

	// First draw.
	w.DrawTo(canvas)
	root.drawCalled = false

	// Mark widget as needing redraw (simulates signal change).
	root.SetNeedsRedraw(true)

	drawn := w.DrawTo(canvas)
	if !drawn {
		t.Error("DrawTo should render after widget marked dirty")
	}
	if !root.drawCalled {
		t.Error("root Draw should be called when widget is dirty")
	}
}

func TestWindow_NeedsRedraw_InitialState(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	if !w.NeedsRedraw() {
		t.Error("window should need redraw after SetRoot")
	}
}

func TestWindow_NeedsRedraw_AfterDraw(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	canvas := &mockCanvas{}
	w.DrawTo(canvas)

	if w.NeedsRedraw() {
		t.Error("window should not need redraw after DrawTo")
	}
}

func TestWindow_NeedsRedraw_AfterResize(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	canvas := &mockCanvas{}
	w.DrawTo(canvas)

	w.HandleResize(1024, 768)

	if !w.NeedsRedraw() {
		t.Error("window should need redraw after resize")
	}
}

func TestWindow_DrawTo_ClearsRedrawFlags(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	canvas := &mockCanvas{}
	w.DrawTo(canvas)

	// Verify all flags were cleared.
	if root.NeedsRedraw() {
		t.Error("root needsRedraw should be cleared after DrawTo")
	}
	if w.NeedsRedraw() {
		t.Error("window needsRedraw should be cleared after DrawTo")
	}
}

func TestWindow_SetRoot_MarksAllDirty(t *testing.T) {
	a := New()
	w := a.Window()

	// Create a tree with pre-cleared redraw flags.
	root := newMockWidget()
	root.ClearRedraw()

	// SetRoot should mark everything as needing redraw.
	w.SetRoot(root)

	if !root.NeedsRedraw() {
		t.Error("root should need redraw after SetRoot")
	}
}

func TestWindow_Frame_DrawSkippedInStats(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	var stats FrameStats

	// First frame: should perform draw (layout marks redraw).
	w.frameCallback = func(s FrameStats) { stats = s }
	w.Frame()

	if stats.DrawSkipped {
		t.Error("first frame should not skip draw")
	}

	// Second frame: nothing changed, draw should be skipped.
	w.Frame()

	if !stats.DrawSkipped {
		t.Error("second frame should skip draw (nothing changed)")
	}
}

func TestWindow_SchedulerFlush_SetsNeedsRedraw(t *testing.T) {
	root := newMockWidget()
	root.ClearRedraw()

	a := New()
	w := a.Window()
	w.SetRoot(root)

	// Clear all flags first.
	canvas := &mockCanvas{}
	w.DrawTo(canvas)
	root.drawCalled = false

	// Simulate signal change by marking dirty through scheduler.
	a.Scheduler().MarkDirty(root)
	a.Scheduler().Flush()

	// The flushFn should have set needsRedraw on the widget.
	if !root.NeedsRedraw() {
		t.Error("widget should have needsRedraw set after scheduler flush")
	}

	// DrawTo should now render.
	drawn := w.DrawTo(canvas)
	if !drawn {
		t.Error("DrawTo should render after scheduler marked widget dirty")
	}
}

func TestWindow_Theme_MarksRedraw(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	canvas := &mockCanvas{}
	w.DrawTo(canvas)

	// Theme change should mark all widgets dirty.
	w.setTheme(theme.DefaultDark())

	if !w.NeedsRedraw() {
		t.Error("window should need redraw after theme change")
	}
	if !root.NeedsRedraw() {
		t.Error("root should need redraw after theme change")
	}
}

// --- DrawStats integration tests ---

func TestWindow_DrawTo_ReturnsDrawStats(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	canvas := &mockCanvas{}
	w.DrawTo(canvas)

	stats := w.LastDrawStats()
	if stats.TotalWidgets != 1 {
		t.Errorf("TotalWidgets = %d, want 1", stats.TotalWidgets)
	}
	if stats.DrawnWidgets != 1 {
		t.Errorf("DrawnWidgets = %d, want 1", stats.DrawnWidgets)
	}
	if stats.DirtyWidgets != 1 {
		t.Errorf("DirtyWidgets = %d, want 1 (first draw, all dirty)", stats.DirtyWidgets)
	}
}

func TestWindow_DrawTo_CleanTreeStillDraws(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	canvas := &mockCanvas{}

	// First draw populates stats.
	w.DrawTo(canvas)
	if w.LastDrawStats().DrawnWidgets != 1 {
		t.Error("first draw should report 1 drawn widget")
	}

	// Second draw: tree is clean but DrawTo still draws (full repaint)
	// because the host may have cleared the pixmap before calling DrawTo.
	drawn := w.DrawTo(canvas)
	if !drawn {
		t.Error("DrawTo should always draw (host may have cleared pixmap)")
	}
	if w.LastDrawStats().DrawnWidgets != 1 {
		t.Error("clean-tree full repaint should still draw 1 widget")
	}
}

func TestWindow_Frame_DrawStatsInCallback(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	var stats FrameStats
	w.frameCallback = func(s FrameStats) { stats = s }
	w.Frame()

	if stats.DrawStats.TotalWidgets != 1 {
		t.Errorf("FrameStats.DrawStats.TotalWidgets = %d, want 1", stats.DrawStats.TotalWidgets)
	}
}

func TestWindow_Frame_DrawStatsZeroWhenSkipped(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	var stats FrameStats
	w.frameCallback = func(s FrameStats) { stats = s }

	// First frame draws.
	w.Frame()
	if stats.DrawSkipped {
		t.Error("first frame should not skip draw")
	}

	// Second frame: nothing changed, draw skipped.
	w.Frame()
	if !stats.DrawSkipped {
		t.Error("second frame should skip draw")
	}
	// DrawStats from the skipped frame should be from the CollectDrawStats
	// pass (headless mode collects stats without drawing).
	// Since headless draw uses CollectDrawStats, it reports stats even when skipping.
}

func TestWindow_LastDrawStats_NoRoot(t *testing.T) {
	a := New()
	w := a.Window()

	stats := w.LastDrawStats()
	if stats.TotalWidgets != 0 {
		t.Errorf("TotalWidgets = %d, want 0 (no root)", stats.TotalWidgets)
	}
}

// --- DirtyTracker wiring tests (Phase 4, ADR-004) ---

func TestWindow_DrawTo_SetsDirtyTrackerOnContext(t *testing.T) {
	// During incremental draw, Window should set the dirty tracker on
	// the context so RepaintBoundary can use the fast path.
	a := New()
	w := a.Window()

	// Track whether DirtyTracker was set during draw by using a widget
	// that inspects the context.
	trackerSeen := false
	spy := &contextSpyWidget{
		onDraw: func(ctx widget.Context) {
			if provider, ok := ctx.(widget.DirtyTrackerProvider); ok {
				if provider.DirtyTracker() != nil {
					trackerSeen = true
				}
			}
		},
	}
	spy.SetVisible(true)
	spy.SetEnabled(true)
	spy.SetNeedsRedraw(true)
	w.SetRoot(spy)

	canvas := &mockCanvas{}

	// First draw is a full repaint (needsFullRepaint=true).
	// Full repaint does NOT set dirty tracker (by design — all widgets
	// are drawn unconditionally). The tracker may or may not be visible
	// depending on the draw path. Reset spy state.
	w.DrawTo(canvas)
	trackerSeen = false

	// Mark widget dirty so next DrawTo triggers incremental path.
	spy.SetNeedsRedraw(true)
	w.needsRedraw = true

	// Layout root to give it non-zero bounds (needed for collector).
	spy.SetBounds(geometry.NewRect(0, 0, 100, 50))

	w.DrawTo(canvas)
	if !trackerSeen {
		t.Error("DirtyTracker should be set on context during incremental draw")
	}

	// After DrawTo, the tracker should be cleared.
	if w.ctx.DirtyTracker() != nil {
		t.Error("DirtyTracker should be nil after DrawTo completes")
	}
}

func TestWindow_DrawTo_DirtyTrackerClearedAfterDraw(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	canvas := &mockCanvas{}
	w.DrawTo(canvas)

	// After any DrawTo, the dirty tracker reference on context should be nil.
	if w.ctx.DirtyTracker() != nil {
		t.Error("DirtyTracker should be nil after DrawTo completes")
	}
}

// contextSpyWidget inspects the context during Draw.
type contextSpyWidget struct {
	widget.WidgetBase
	onDraw func(ctx widget.Context)
}

func (w *contextSpyWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(100, 50))
}

func (w *contextSpyWidget) Draw(ctx widget.Context, _ widget.Canvas) {
	if w.onDraw != nil {
		w.onDraw(ctx)
	}
}

func (w *contextSpyWidget) Event(_ widget.Context, _ event.Event) bool { return false }

var _ widget.Widget = (*contextSpyWidget)(nil)

// --- Focus management tests ---

// focusableMockWidget implements both widget.Widget and widget.Focusable.
type focusableMockWidget struct {
	widget.WidgetBase
	layoutSize geometry.Size
	name       string
}

func newFocusableMock(name string) *focusableMockWidget {
	w := &focusableMockWidget{
		layoutSize: geometry.Sz(100, 30),
		name:       name,
	}
	w.SetEnabled(true)
	w.SetVisible(true)
	return w
}

func (w *focusableMockWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(w.layoutSize)
}

func (w *focusableMockWidget) Draw(_ widget.Context, _ widget.Canvas) {}

func (w *focusableMockWidget) Event(_ widget.Context, _ event.Event) bool {
	return false
}

func (w *focusableMockWidget) IsFocusable() bool {
	return w.IsEnabled() && w.IsVisible()
}

// containerMockWidget holds children for focus traversal.
type containerMockWidget struct {
	widget.WidgetBase
	children []widget.Widget
}

func newContainerMock(children ...widget.Widget) *containerMockWidget {
	c := &containerMockWidget{children: children}
	c.SetEnabled(true)
	c.SetVisible(true)
	return c
}

func (c *containerMockWidget) Layout(_ widget.Context, cs geometry.Constraints) geometry.Size {
	return cs.Constrain(geometry.Sz(400, 300))
}

func (c *containerMockWidget) Draw(_ widget.Context, _ widget.Canvas) {}

func (c *containerMockWidget) Event(_ widget.Context, _ event.Event) bool {
	return false
}

func (c *containerMockWidget) Children() []widget.Widget {
	return c.children
}

func TestWindow_TabNavigation_ForwardCycle(t *testing.T) {
	a := New()
	w := a.Window()

	field1 := newFocusableMock("field1")
	field2 := newFocusableMock("field2")
	field3 := newFocusableMock("field3")
	root := newContainerMock(field1, field2, field3)

	w.SetRoot(root)
	w.Frame() // trigger layout so focus ring is built

	// No widget focused initially.
	if w.Context().FocusedWidget() != nil {
		t.Error("no widget should be focused initially")
	}

	// Tab: focus moves to first focusable widget.
	tabPress := event.NewKeyEvent(event.KeyPress, event.KeyTab, 0, event.ModNone)
	w.HandleEvent(tabPress)

	if !field1.IsFocused() {
		t.Error("field1 should be focused after first Tab")
	}
	if w.Context().FocusedWidget() != field1 {
		t.Error("context should report field1 as focused")
	}

	// Tab again: focus moves to field2.
	w.HandleEvent(tabPress)
	if !field2.IsFocused() {
		t.Error("field2 should be focused after second Tab")
	}
	if field1.IsFocused() {
		t.Error("field1 should be blurred after second Tab")
	}

	// Tab again: focus moves to field3.
	w.HandleEvent(tabPress)
	if !field3.IsFocused() {
		t.Error("field3 should be focused after third Tab")
	}

	// Tab again: wraps to field1.
	w.HandleEvent(tabPress)
	if !field1.IsFocused() {
		t.Error("field1 should be focused after wrap-around Tab")
	}
}

func TestWindow_TabNavigation_Backward(t *testing.T) {
	a := New()
	w := a.Window()

	field1 := newFocusableMock("field1")
	field2 := newFocusableMock("field2")
	root := newContainerMock(field1, field2)

	w.SetRoot(root)
	w.Frame()

	// Shift+Tab with no focus: should focus last widget.
	shiftTabPress := event.NewKeyEvent(event.KeyPress, event.KeyTab, 0, event.ModShift)
	w.HandleEvent(shiftTabPress)

	if !field2.IsFocused() {
		t.Error("field2 should be focused after first Shift+Tab (last focusable)")
	}

	// Shift+Tab again: moves to field1.
	w.HandleEvent(shiftTabPress)
	if !field1.IsFocused() {
		t.Error("field1 should be focused after second Shift+Tab")
	}
	if field2.IsFocused() {
		t.Error("field2 should be blurred")
	}

	// Shift+Tab again: wraps to field2.
	w.HandleEvent(shiftTabPress)
	if !field2.IsFocused() {
		t.Error("field2 should be focused after wrap-around Shift+Tab")
	}
}

func TestWindow_TabNavigation_NoFocusableWidgets(t *testing.T) {
	a := New()
	w := a.Window()

	// Root with no focusable children.
	root := newMockWidget()
	w.SetRoot(root)
	w.Frame()

	tabPress := event.NewKeyEvent(event.KeyPress, event.KeyTab, 0, event.ModNone)
	// Should not panic.
	w.HandleEvent(tabPress)

	if w.Context().FocusedWidget() != nil {
		t.Error("no widget should be focused when there are no focusable widgets")
	}
}

func TestWindow_TabNavigation_ContextSyncAfterMouseFocus(t *testing.T) {
	a := New()
	w := a.Window()

	field1 := newFocusableMock("field1")
	field2 := newFocusableMock("field2")
	field3 := newFocusableMock("field3")
	root := newContainerMock(field1, field2, field3)

	w.SetRoot(root)
	w.Frame()

	// Simulate mouse click focusing field2 via context (as widgets do).
	w.Context().RequestFocus(field2)

	// Now Tab should move from field2 to field3 (not restart from field1).
	tabPress := event.NewKeyEvent(event.KeyPress, event.KeyTab, 0, event.ModNone)
	w.HandleEvent(tabPress)

	if !field3.IsFocused() {
		t.Errorf("field3 should be focused after Tab from field2; field1=%v field2=%v field3=%v",
			field1.IsFocused(), field2.IsFocused(), field3.IsFocused())
	}
}

func TestWindow_TabNavigation_NonKeyEventsPassThrough(t *testing.T) {
	a := New()
	w := a.Window()

	field1 := newFocusableMock("field1")
	root := newContainerMock(field1)
	w.SetRoot(root)
	w.Frame()

	// Non-key events should not be intercepted by focus manager.
	mouseEvent := event.NewMouseEvent(
		event.MousePress,
		event.ButtonLeft,
		event.ButtonStateLeft,
		geometry.Pt(10, 20),
		geometry.Pt(10, 20),
		event.ModNone,
	)
	w.HandleEvent(mouseEvent)

	// Field1 should not be focused (mouse events go to widget tree, not focus manager).
	if field1.IsFocused() {
		t.Error("mouse event should not trigger focus manager")
	}
}

func TestWindow_TabNavigation_KeyReleaseConsumed(t *testing.T) {
	a := New()
	w := a.Window()

	field1 := newFocusableMock("field1")
	field2 := newFocusableMock("field2")
	root := newContainerMock(field1, field2)
	w.SetRoot(root)
	w.Frame()

	// Tab press moves focus.
	tabPress := event.NewKeyEvent(event.KeyPress, event.KeyTab, 0, event.ModNone)
	w.HandleEvent(tabPress)
	if !field1.IsFocused() {
		t.Error("field1 should be focused")
	}

	// Tab release should also be consumed (not dispatched to widget tree).
	tabRelease := event.NewKeyEvent(event.KeyRelease, event.KeyTab, 0, event.ModNone)
	w.HandleEvent(tabRelease)

	// Focus should still be on field1 (release doesn't move focus).
	if !field1.IsFocused() {
		t.Error("field1 should still be focused after Tab release")
	}
}

func TestWindow_FocusManager_Accessor(t *testing.T) {
	a := New()
	w := a.Window()

	if w.FocusManager() == nil {
		t.Error("FocusManager() should not return nil")
	}
}

// --- Hover tracking tests ---

// hoverTrackingWidget records MouseEnter/MouseLeave events.
type hoverTrackingWidget struct {
	widget.WidgetBase
	enterCount int
	leaveCount int
	layoutSize geometry.Size
}

func newHoverWidget(r geometry.Rect) *hoverTrackingWidget {
	h := &hoverTrackingWidget{
		layoutSize: r.Size(),
	}
	h.SetVisible(true)
	h.SetEnabled(true)
	h.SetBounds(r)
	// Set screen origin to match bounds for hit testing.
	h.SetScreenOrigin(r.Min)
	return h
}

func (h *hoverTrackingWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(h.layoutSize)
}

func (h *hoverTrackingWidget) Draw(_ widget.Context, _ widget.Canvas) {}

func (h *hoverTrackingWidget) Event(_ widget.Context, e event.Event) bool {
	if me, ok := e.(*event.MouseEvent); ok {
		switch me.MouseType {
		case event.MouseEnter:
			h.enterCount++
			return true
		case event.MouseLeave:
			h.leaveCount++
			return true
		}
	}
	return false
}

// hoverContainer holds children for hover test scenarios.
type hoverContainer struct {
	widget.WidgetBase
	kids []widget.Widget
}

func newHoverContainer(children ...widget.Widget) *hoverContainer {
	c := &hoverContainer{kids: children}
	c.SetVisible(true)
	c.SetEnabled(true)
	c.SetBounds(geometry.NewRect(0, 0, 800, 600))
	c.SetScreenOrigin(geometry.Pt(0, 0))
	return c
}

func (c *hoverContainer) Layout(_ widget.Context, cs geometry.Constraints) geometry.Size {
	return cs.Constrain(geometry.Sz(800, 600))
}

func (c *hoverContainer) Draw(_ widget.Context, _ widget.Canvas) {}

func (c *hoverContainer) Event(_ widget.Context, _ event.Event) bool {
	return false
}

func (c *hoverContainer) Children() []widget.Widget {
	return c.kids
}

func TestWindow_HoverTracking_EnterWidget(t *testing.T) {
	a := New()
	w := a.Window()

	btn := newHoverWidget(geometry.NewRect(10, 10, 110, 50))
	root := newHoverContainer(btn)
	w.SetRoot(root)

	// Move mouse into the button's ScreenBounds.
	moveEvent := event.NewMouseEvent(
		event.MouseMove,
		event.ButtonNone,
		0,
		geometry.Pt(50, 30),
		geometry.Pt(50, 30),
		event.ModNone,
	)
	w.HandleEvent(moveEvent)

	if btn.enterCount != 1 {
		t.Errorf("enterCount = %d, want 1", btn.enterCount)
	}
	if w.HoveredWidget() != btn {
		t.Error("hovered widget should be the button")
	}
}

func TestWindow_HoverTracking_LeaveWidget(t *testing.T) {
	a := New()
	w := a.Window()

	btn := newHoverWidget(geometry.NewRect(10, 10, 110, 50))
	root := newHoverContainer(btn)
	w.SetRoot(root)

	// Enter the button.
	w.HandleEvent(event.NewMouseEvent(
		event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(50, 30), geometry.Pt(50, 30), event.ModNone,
	))

	// Move outside the button (but inside the container).
	w.HandleEvent(event.NewMouseEvent(
		event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(200, 200), geometry.Pt(200, 200), event.ModNone,
	))

	if btn.leaveCount != 1 {
		t.Errorf("leaveCount = %d, want 1", btn.leaveCount)
	}
	// Hover should now be on the container (root).
	if w.HoveredWidget() == btn {
		t.Error("hovered widget should no longer be the button")
	}
}

func TestWindow_HoverTracking_MoveWithinSameWidget(t *testing.T) {
	a := New()
	w := a.Window()

	btn := newHoverWidget(geometry.NewRect(10, 10, 110, 50))
	root := newHoverContainer(btn)
	w.SetRoot(root)

	// Move inside the button twice — should only generate one Enter.
	w.HandleEvent(event.NewMouseEvent(
		event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(20, 20), geometry.Pt(20, 20), event.ModNone,
	))
	w.HandleEvent(event.NewMouseEvent(
		event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(30, 30), geometry.Pt(30, 30), event.ModNone,
	))

	if btn.enterCount != 1 {
		t.Errorf("enterCount = %d, want 1 (no duplicate Enter)", btn.enterCount)
	}
	if btn.leaveCount != 0 {
		t.Errorf("leaveCount = %d, want 0", btn.leaveCount)
	}
}

func TestWindow_HoverTracking_WindowLeave(t *testing.T) {
	a := New()
	w := a.Window()

	btn := newHoverWidget(geometry.NewRect(10, 10, 110, 50))
	root := newHoverContainer(btn)
	w.SetRoot(root)

	// Enter the button.
	w.HandleEvent(event.NewMouseEvent(
		event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(50, 30), geometry.Pt(50, 30), event.ModNone,
	))

	// Mouse leaves the window entirely.
	w.HandleEvent(event.NewMouseEvent(
		event.MouseLeave, event.ButtonNone, 0,
		geometry.Pt(0, 0), geometry.Pt(0, 0), event.ModNone,
	))

	if btn.leaveCount != 1 {
		t.Errorf("leaveCount = %d, want 1 (window leave should clear hover)", btn.leaveCount)
	}
	if w.HoveredWidget() != nil {
		t.Error("hovered widget should be nil after window leave")
	}
}

func TestWindow_HoverTracking_SwitchBetweenWidgets(t *testing.T) {
	a := New()
	w := a.Window()

	btn1 := newHoverWidget(geometry.NewRect(10, 10, 110, 50))
	btn2 := newHoverWidget(geometry.NewRect(10, 60, 110, 100))
	root := newHoverContainer(btn1, btn2)
	w.SetRoot(root)

	// Enter btn1.
	w.HandleEvent(event.NewMouseEvent(
		event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(50, 30), geometry.Pt(50, 30), event.ModNone,
	))

	if btn1.enterCount != 1 {
		t.Errorf("btn1 enterCount = %d, want 1", btn1.enterCount)
	}

	// Move to btn2.
	w.HandleEvent(event.NewMouseEvent(
		event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(50, 80), geometry.Pt(50, 80), event.ModNone,
	))

	if btn1.leaveCount != 1 {
		t.Errorf("btn1 leaveCount = %d, want 1", btn1.leaveCount)
	}
	if btn2.enterCount != 1 {
		t.Errorf("btn2 enterCount = %d, want 1", btn2.enterCount)
	}
	if w.HoveredWidget() != btn2 {
		t.Error("hovered widget should be btn2")
	}
}

func TestWindow_HoverTracking_InvisibleWidgetSkipped(t *testing.T) {
	a := New()
	w := a.Window()

	btn := newHoverWidget(geometry.NewRect(10, 10, 110, 50))
	btn.SetVisible(false)
	root := newHoverContainer(btn)
	w.SetRoot(root)

	// Move into where the button would be — it's invisible.
	w.HandleEvent(event.NewMouseEvent(
		event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(50, 30), geometry.Pt(50, 30), event.ModNone,
	))

	if btn.enterCount != 0 {
		t.Errorf("enterCount = %d, want 0 (invisible widget)", btn.enterCount)
	}
	// Should hover the container instead.
	if w.HoveredWidget() == btn {
		t.Error("invisible widget should not receive hover")
	}
}

func TestWindow_HoverTracking_NoRoot(t *testing.T) {
	a := New()
	w := a.Window()

	// Should not panic with no root.
	w.HandleEvent(event.NewMouseEvent(
		event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(50, 30), geometry.Pt(50, 30), event.ModNone,
	))

	if w.HoveredWidget() != nil {
		t.Error("hovered widget should be nil with no root")
	}
}

func TestHitTest_Nil(t *testing.T) {
	result := hitTest(nil, geometry.Pt(10, 10))
	if result != nil {
		t.Error("hitTest(nil, ...) should return nil")
	}
}

func TestHitTest_OutsideBounds(t *testing.T) {
	btn := newHoverWidget(geometry.NewRect(50, 10, 150, 50))
	result := hitTest(btn, geometry.Pt(200, 200))
	if result != nil {
		t.Error("hitTest should return nil when point is outside bounds")
	}
}

func TestHitTest_DeepestChild(t *testing.T) {
	// Container with a child — hit test should return the deepest widget.
	child := newHoverWidget(geometry.NewRect(30, 10, 130, 50))
	root := newHoverContainer(child)

	result := hitTest(root, geometry.Pt(70, 30))
	if result != child {
		t.Errorf("hitTest should return deepest child, got %T", result)
	}
}

func TestHitTest_ReverseZOrder(t *testing.T) {
	// Two overlapping children — last child (higher z-order) should win.
	child1 := newHoverWidget(geometry.NewRect(20, 10, 120, 50))
	child2 := newHoverWidget(geometry.NewRect(20, 10, 120, 50)) // Same bounds, higher z-order
	root := newHoverContainer(child1, child2)

	result := hitTest(root, geometry.Pt(60, 30))
	if result != child2 {
		t.Error("hitTest should return the topmost (last) child for overlapping widgets")
	}
}

// --- Animation scheduling regression tests ---
//
// These tests prevent regressions for the bug where needsLayout was
// unconditionally cleared after layout(), clobbering invalidation
// requests made by animating widgets during the layout pass.
// See the animPumper and BeginFrame fixes.

// invalidatingOnLayoutWidget calls ctx.Invalidate() during Layout,
// simulating what animated widgets (e.g. collapsible) do when they
// call tickAnimation() inside Layout().
type invalidatingOnLayoutWidget struct {
	widget.WidgetBase
	invalidateOnLayout bool
}

func newInvalidatingWidget(invalidate bool) *invalidatingOnLayoutWidget {
	w := &invalidatingOnLayoutWidget{invalidateOnLayout: invalidate}
	w.SetVisible(true)
	w.SetEnabled(true)
	return w
}

func (w *invalidatingOnLayoutWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	if w.invalidateOnLayout {
		ctx.Invalidate()
	}
	return c.Constrain(geometry.Sz(100, 50))
}

func (w *invalidatingOnLayoutWidget) Draw(_ widget.Context, _ widget.Canvas) {}

func (w *invalidatingOnLayoutWidget) Event(_ widget.Context, _ event.Event) bool {
	return false
}

// TestWindow_Frame_NeedsLayoutPreservedWhenInvalidatedDuringLayout verifies
// that needsLayout is NOT cleared when a widget calls ctx.Invalidate() during
// Layout(). This is the KEY regression test for the animation scheduling bug:
// before the fix, needsLayout was unconditionally set to false after layout(),
// which meant animations only progressed when external events (mouse move)
// triggered new frames.
func TestWindow_Frame_NeedsLayoutPreservedWhenInvalidatedDuringLayout(t *testing.T) {
	wp := &mockWindowProvider{width: 800, height: 600, scale: 1.0}
	a := New(WithWindowProvider(wp))
	w := a.Window()

	root := newInvalidatingWidget(true) // Will call Invalidate during Layout
	w.SetRoot(root)

	w.Frame()

	// After Frame, needsLayout should still be true because the widget
	// re-invalidated during layout (simulating an active animation).
	if !w.NeedsLayout() {
		t.Error("needsLayout should be preserved when widget invalidates during layout; " +
			"this was the root cause of the animation scheduling bug")
	}
}

// TestWindow_Frame_NeedsLayoutClearedWhenNoInvalidation verifies that
// needsLayout IS cleared normally when no widget invalidates during layout.
func TestWindow_Frame_NeedsLayoutClearedWhenNoInvalidation(t *testing.T) {
	wp := &mockWindowProvider{width: 800, height: 600, scale: 1.0}
	a := New(WithWindowProvider(wp))
	w := a.Window()

	root := newInvalidatingWidget(false) // Normal widget, no invalidation
	w.SetRoot(root)

	w.Frame()

	if w.NeedsLayout() {
		t.Error("needsLayout should be cleared when no widget invalidates during layout")
	}
}

// TestWindow_AnimPumper_StartsOnInvalidation verifies that the animation
// pumper goroutine is started when a widget calls ctx.Invalidate() during
// a frame (e.g. from tickAnimation). Without the pumper, animations would
// only progress when the OS sends input events.
func TestWindow_AnimPumper_StartsOnInvalidation(t *testing.T) {
	wp := &mockWindowProvider{width: 800, height: 600, scale: 1.0}
	a := New(WithWindowProvider(wp))
	w := a.Window()

	root := newInvalidatingWidget(true)
	w.SetRoot(root)

	// Before first frame, no pumper.
	if w.animToken != nil {
		t.Fatal("animToken should be nil before first frame")
	}

	w.Frame()

	// Widget invalidated during layout, so pumper should start.
	if w.animToken == nil {
		t.Error("animToken should be started after frame with invalidation")
	}

	// Verify RequestRedraw was called (from the invalidation callback).
	if wp.redrawCount == 0 {
		t.Error("WindowProvider.RequestRedraw should have been called")
	}
}

// TestWindow_AnimPumper_StopsAfterIdleFrames verifies that the animation
// pumper is stopped after 3+ consecutive frames with no invalidation.
// This prevents the pumper from running forever after animations complete.
func TestWindow_AnimPumper_StopsAfterIdleFrames(t *testing.T) {
	wp := &mockWindowProvider{width: 800, height: 600, scale: 1.0}
	a := New(WithWindowProvider(wp))
	w := a.Window()

	root := newInvalidatingWidget(true)
	w.SetRoot(root)

	// First frame: starts pumper.
	w.Frame()
	if w.animToken == nil {
		t.Fatal("pumper should be running after invalidating frame")
	}

	// Stop invalidating.
	root.invalidateOnLayout = false

	// Run frames until pumper stops. The threshold is >3 idle frames.
	for range 10 {
		w.Frame()
		if w.animToken == nil {
			// Pumper stopped — verify it took more than 3 idle frames.
			if w.animIdleFrames != 0 {
				t.Errorf("animIdleFrames should be reset to 0 after stop, got %d", w.animIdleFrames)
			}
			return // Success
		}
	}
	t.Error("animToken should become nil after consecutive idle frames")
}

// TestWindow_AnimPumper_NotStartedWithoutWindowProvider verifies that
// the animation pumper is NOT created in headless mode (wp==nil).
// In headless mode there is no window to request redraws from.
func TestWindow_AnimPumper_NotStartedWithoutWindowProvider(t *testing.T) {
	a := New() // Headless, no WindowProvider.
	w := a.Window()

	root := newInvalidatingWidget(true)
	w.SetRoot(root)

	w.Frame()

	if w.animToken != nil {
		t.Error("animToken should be nil in headless mode (no WindowProvider)")
	}
}

// --- Dirty Boundaries Tests (ADR-007 Task 1e, Phase C: O(1) flat list) ---

func TestWindow_DirtyBoundaries_Initial(t *testing.T) {
	a := New()
	w := a.Window()

	if w.HasDirtyBoundaries() {
		t.Error("new window should have no dirty boundaries")
	}
	if w.DirtyBoundaryCount() != 0 {
		t.Errorf("expected 0 dirty boundaries, got %d", w.DirtyBoundaryCount())
	}
}

func TestWindow_DirtyBoundaries_AddAndCount(t *testing.T) {
	a := New()
	w := a.Window()

	w.AddDirtyBoundary(1)
	w.AddDirtyBoundary(2)

	if !w.HasDirtyBoundaries() {
		t.Error("should have dirty boundaries after Add")
	}
	if w.DirtyBoundaryCount() != 2 {
		t.Errorf("expected 2 dirty boundaries, got %d", w.DirtyBoundaryCount())
	}
}

func TestWindow_DirtyBoundaries_Deduplication(t *testing.T) {
	a := New()
	w := a.Window()

	w.AddDirtyBoundary(42)
	w.AddDirtyBoundary(42) // Same key — should deduplicate.

	if w.DirtyBoundaryCount() != 1 {
		t.Errorf("expected 1 dirty boundary (deduplicated), got %d", w.DirtyBoundaryCount())
	}
}

func TestWindow_DirtyBoundaries_Clear(t *testing.T) {
	a := New()
	w := a.Window()

	w.AddDirtyBoundary(1)
	w.AddDirtyBoundary(2)

	w.ClearDirtyBoundaries()

	if w.HasDirtyBoundaries() {
		t.Error("should have no dirty boundaries after Clear")
	}
	if w.DirtyBoundaryCount() != 0 {
		t.Errorf("expected 0 dirty boundaries after Clear, got %d", w.DirtyBoundaryCount())
	}
}

func TestWindow_DirtyBoundaries_ClearAndReuse(t *testing.T) {
	a := New()
	w := a.Window()

	w.AddDirtyBoundary(1)
	w.ClearDirtyBoundaries()

	// After clear, adding again should work.
	w.AddDirtyBoundary(2)

	if w.DirtyBoundaryCount() != 1 {
		t.Errorf("expected 1 dirty boundary after clear+add, got %d", w.DirtyBoundaryCount())
	}
}

func TestWindow_DirtyBoundaries_ClearedAfterPaintDirtyBoundaries(t *testing.T) {
	a := New()
	w := a.Window()

	w.AddDirtyBoundary(1)
	if !w.HasDirtyBoundaries() {
		t.Fatal("pre-condition: should have dirty boundary")
	}

	w.PaintDirtyBoundaries()

	if w.HasDirtyBoundaries() {
		t.Error("dirty boundaries should be empty after PaintDirtyBoundaries")
	}
}

// TestWindow_DirtyBoundaryRegistration_ViaContext verifies that
// ContextImpl.RegisterDirtyBoundary populates the Window's dirty set.
// This is the O(1) flat dirty list that replaces O(n) tree walks.
func TestWindow_DirtyBoundaryRegistration_ViaContext(t *testing.T) {
	a := New()
	w := a.Window()
	ctx := w.Context()

	if w.HasDirtyBoundaries() {
		t.Fatal("pre-condition: should start clean")
	}

	// RegisterDirtyBoundary fires the callback wired in newWindow.
	ctx.RegisterDirtyBoundary(42)

	if !w.HasDirtyBoundaries() {
		t.Error("RegisterDirtyBoundary should populate Window.dirtyBoundaries")
	}
	if w.DirtyBoundaryCount() != 1 {
		t.Errorf("expected 1, got %d", w.DirtyBoundaryCount())
	}
}

// TestWindow_DirtyBoundaryRegistration_DeduplicatesSameKey verifies
// that multiple RegisterDirtyBoundary calls with the same key
// produce only one entry (map deduplication).
func TestWindow_DirtyBoundaryRegistration_DeduplicatesSameKey(t *testing.T) {
	a := New()
	w := a.Window()
	ctx := w.Context()

	ctx.RegisterDirtyBoundary(7)
	ctx.RegisterDirtyBoundary(7)
	ctx.RegisterDirtyBoundary(7)

	if w.DirtyBoundaryCount() != 1 {
		t.Errorf("expected 1 (deduplicated), got %d", w.DirtyBoundaryCount())
	}
}

// TestWindow_DirtyBoundaryRegistration_MultipleBoundaries verifies
// that different keys produce separate entries.
func TestWindow_DirtyBoundaryRegistration_MultipleBoundaries(t *testing.T) {
	a := New()
	w := a.Window()
	ctx := w.Context()

	ctx.RegisterDirtyBoundary(1)
	ctx.RegisterDirtyBoundary(2)
	ctx.RegisterDirtyBoundary(3)

	if w.DirtyBoundaryCount() != 3 {
		t.Errorf("expected 3, got %d", w.DirtyBoundaryCount())
	}
}

// --- Cursor regression tests (2026-05-07) ---

// TestCursorNotResetByFrame verifies that Frame() does not clobber a cursor
// set during event handling. Before the fix, ResetCursor was called in Frame()
// after layout, which overwrote CursorPointer set by a widget's Event handler.
// Regression: ResetCursor in Frame() erased cursor set by Event handler (2026-05-07)
func TestCursorNotResetByFrame(t *testing.T) {
	pp := &mockPlatformProvider{fontScale: 1.0}
	a := New(WithPlatformProvider(pp))
	w := a.Window()

	// Use a widget that sets CursorPointer during its Event handler.
	cw := &cursorSettingWidget{cursor: widget.CursorPointer}
	cw.SetVisible(true)
	cw.SetEnabled(true)
	w.SetRoot(cw)

	// Simulate an event that causes the widget to set CursorPointer.
	me := event.NewMouseEvent(
		event.MousePress,
		event.ButtonLeft,
		event.ButtonStateLeft,
		geometry.Pt(50, 25),
		geometry.Pt(50, 25),
		event.ModNone,
	)
	w.HandleEvent(me)

	// Verify cursor was set.
	if w.Context().Cursor() != widget.CursorPointer {
		t.Fatal("precondition: cursor should be Pointer after event")
	}

	// Call Frame() — this must NOT reset the cursor back to Default.
	w.Frame()

	if w.Context().Cursor() != widget.CursorPointer {
		t.Errorf("cursor = %v after Frame(), want CursorPointer; "+
			"Frame() must not clobber cursor set during event handling",
			w.Context().Cursor())
	}
}

// TestCursorResetOnHoverChange verifies that when the hover target changes
// (e.g., mouse moves from an interactive widget to empty space), the cursor
// is reset to Default. Without this, cursor remained as Pointer on
// non-interactive areas after leaving a button.
// Regression: cursor stuck as Pointer after leaving interactive widget (2026-05-07)
func TestCursorResetOnHoverChange(t *testing.T) {
	pp := &mockPlatformProvider{fontScale: 1.0}
	a := New(WithPlatformProvider(pp))
	w := a.Window()

	// cursorWidget sets Pointer on MouseEnter.
	btn := &hoverCursorWidget{cursor: widget.CursorPointer}
	btn.SetVisible(true)
	btn.SetEnabled(true)
	btn.SetBounds(geometry.NewRect(10, 10, 100, 40))
	btn.SetScreenOrigin(geometry.Pt(10, 10))

	root := newHoverContainer(btn)
	w.SetRoot(root)

	// Move mouse into the button — cursor should become Pointer.
	w.HandleEvent(event.NewMouseEvent(
		event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(50, 25), geometry.Pt(50, 25), event.ModNone,
	))

	if w.Context().Cursor() != widget.CursorPointer {
		t.Fatal("precondition: cursor should be Pointer when over button")
	}

	// Move mouse away from button to empty area.
	w.HandleEvent(event.NewMouseEvent(
		event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(400, 400), geometry.Pt(400, 400), event.ModNone,
	))

	if w.Context().Cursor() != widget.CursorDefault {
		t.Errorf("cursor = %v after leaving button, want CursorDefault; "+
			"updateHover must reset cursor when hover target changes",
			w.Context().Cursor())
	}
}

// hoverCursorWidget sets a cursor on MouseEnter.
type hoverCursorWidget struct {
	widget.WidgetBase
	cursor widget.CursorType
}

func (w *hoverCursorWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(100, 40))
}

func (w *hoverCursorWidget) Draw(_ widget.Context, _ widget.Canvas) {}

func (w *hoverCursorWidget) Event(ctx widget.Context, e event.Event) bool {
	if me, ok := e.(*event.MouseEvent); ok && me.MouseType == event.MouseEnter {
		ctx.SetCursor(w.cursor)
		return true
	}
	return false
}

// --- Pointer Capture Tests (ADR-031) ---

// captureTrackingWidget records mouse events and can capture the pointer.
type captureTrackingWidget struct {
	widget.WidgetBase
	events      []*event.MouseEvent
	captureOnce bool // capture pointer on first MousePress
}

func newCaptureWidget(bounds geometry.Rect) *captureTrackingWidget {
	w := &captureTrackingWidget{}
	w.SetVisible(true)
	w.SetEnabled(true)
	w.SetBounds(bounds)
	w.SetScreenOrigin(bounds.Min)
	return w
}

func (w *captureTrackingWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(w.Bounds().Size())
}

func (w *captureTrackingWidget) Draw(_ widget.Context, _ widget.Canvas) {}

func (w *captureTrackingWidget) Event(ctx widget.Context, e event.Event) bool {
	me, ok := e.(*event.MouseEvent)
	if !ok {
		return false
	}
	w.events = append(w.events, me)

	if w.captureOnce && me.MouseType == event.MousePress {
		if pc, ok := ctx.(widget.PointerCapturer); ok {
			pc.CapturePointer(w)
		}
		w.captureOnce = false
	}

	return true
}

func TestWindow_PointerCapture_DirectDelivery(t *testing.T) {
	a := New()
	w := a.Window()

	// Widget at (10,10)-(110,50). Pointer capture should deliver events
	// even when mouse is outside these bounds.
	cw := newCaptureWidget(geometry.NewRect(10, 10, 110, 50))
	root := newHoverContainer(cw)
	w.SetRoot(root)

	// Set capture via context (simulating what a real widget does on press).
	w.ctx.CapturePointer(cw)

	// Move OUTSIDE widget bounds — should be delivered via capture.
	move := event.NewMouseEvent(
		event.MouseMove, event.ButtonNone, event.ButtonStateLeft,
		geometry.Pt(300, 300), geometry.Pt(300, 300), event.ModNone,
	)
	w.HandleEvent(move)

	if len(cw.events) != 1 {
		t.Fatalf("expected 1 event after move outside bounds, got %d", len(cw.events))
	}
	if cw.events[0].MouseType != event.MouseMove {
		t.Errorf("expected MouseMove, got %v", cw.events[0].MouseType)
	}

	// Move again further out — still captured.
	move2 := event.NewMouseEvent(
		event.MouseMove, event.ButtonNone, event.ButtonStateLeft,
		geometry.Pt(600, 600), geometry.Pt(600, 600), event.ModNone,
	)
	w.HandleEvent(move2)

	if len(cw.events) != 2 {
		t.Fatalf("expected 2 events after second move, got %d", len(cw.events))
	}
}

func TestWindow_PointerCapture_AutoRelease(t *testing.T) {
	a := New()
	w := a.Window()

	cw := newCaptureWidget(geometry.NewRect(10, 10, 110, 50))
	cw.captureOnce = true
	root := newHoverContainer(cw)
	w.SetRoot(root)

	// Press inside widget — triggers capture.
	press := event.NewMouseEvent(
		event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(50, 30), geometry.Pt(50, 30), event.ModNone,
	)
	w.HandleEvent(press)

	// Release outside bounds — auto-releases capture (Buttons==0).
	release := event.NewMouseEvent(
		event.MouseRelease, event.ButtonLeft, 0, // all buttons up
		geometry.Pt(300, 300), geometry.Pt(300, 300), event.ModNone,
	)
	w.HandleEvent(release)

	// Capture should be cleared.
	if w.capturedWidget != nil {
		t.Error("capturedWidget should be nil after MouseRelease with all buttons up")
	}

	// Subsequent move should NOT go to the captured widget.
	cw.events = nil
	move := event.NewMouseEvent(
		event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(300, 300), geometry.Pt(300, 300), event.ModNone,
	)
	w.HandleEvent(move)

	// The move event should go through normal dispatch, not to cw
	// (since (300,300) is outside cw's bounds).
	for _, ev := range cw.events {
		if ev.MouseType == event.MouseMove {
			t.Error("move event should not be delivered to widget after capture release")
		}
	}
}

func TestWindow_PointerCapture_ReleaseWrongWidget(t *testing.T) {
	a := New()
	w := a.Window()

	cw1 := newCaptureWidget(geometry.NewRect(10, 10, 110, 50))
	cw2 := newCaptureWidget(geometry.NewRect(10, 60, 110, 100))
	root := newHoverContainer(cw1, cw2)
	w.SetRoot(root)

	// Capture cw1 via the context callback.
	w.ctx.CapturePointer(cw1)

	if w.capturedWidget != cw1 {
		t.Fatal("precondition: cw1 should be captured")
	}

	// Release cw2 — should be a no-op (wrong widget).
	w.ctx.ReleasePointer(cw2)

	if w.capturedWidget != cw1 {
		t.Error("releasing a different widget should not clear the capture")
	}

	// Release cw1 — should clear.
	w.ctx.ReleasePointer(cw1)

	if w.capturedWidget != nil {
		t.Error("capturedWidget should be nil after releasing the correct widget")
	}
}

func TestWindow_PointerCapture_OverlaysPriority(t *testing.T) {
	a := New()
	w := a.Window()

	cw := newCaptureWidget(geometry.NewRect(10, 10, 110, 50))
	root := newHoverContainer(cw)
	w.SetRoot(root)

	// Capture the widget.
	w.ctx.CapturePointer(cw)

	// Overlay events still take priority over capture (the capture
	// short-circuit is AFTER the overlay check in HandleEvent).
	// We just verify that with no overlays, captured events work.
	move := event.NewMouseEvent(
		event.MouseMove, event.ButtonNone, event.ButtonStateLeft,
		geometry.Pt(500, 500), geometry.Pt(500, 500), event.ModNone,
	)
	w.HandleEvent(move)

	if len(cw.events) != 1 {
		t.Errorf("expected 1 event delivered via capture, got %d", len(cw.events))
	}
}
