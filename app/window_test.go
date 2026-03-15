package app

import (
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
	w := newWindow(wp, nil, sched, theme.DefaultLight())
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

func (m *mockCanvas) Clear(widget.Color)                                                            {}
func (m *mockCanvas) DrawRect(geometry.Rect, widget.Color)                                          {}
func (m *mockCanvas) StrokeRect(geometry.Rect, widget.Color, float32)                               {}
func (m *mockCanvas) DrawRoundRect(geometry.Rect, widget.Color, float32)                            {}
func (m *mockCanvas) StrokeRoundRect(geometry.Rect, widget.Color, float32, float32)                 {}
func (m *mockCanvas) DrawCircle(geometry.Point, float32, widget.Color)                              {}
func (m *mockCanvas) StrokeCircle(geometry.Point, float32, widget.Color, float32)                   {}
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

// --- Retained-mode rendering tests ---

func TestWindow_DrawTo_ReportsCleanTree(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	canvas := &mockCanvas{}

	// First draw: all widgets are dirty (just mounted).
	drawn := w.DrawTo(canvas)
	if !drawn {
		t.Error("first DrawTo should report dirty (all widgets dirty after mount)")
	}

	// Reset tracking.
	root.drawCalled = false

	// Second draw: nothing changed. In Sub-Phase 1, DrawTo still draws
	// (existing widgets don't self-dirty on event state changes yet),
	// but returns false to indicate the tree was clean.
	drawn = w.DrawTo(canvas)
	if drawn {
		t.Error("second DrawTo should report clean (no widgets dirty)")
	}
	if !root.drawCalled {
		t.Error("root Draw should still be called (Sub-Phase 1 always draws)")
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

func TestWindow_DrawTo_SkippedFrameHasZeroStats(t *testing.T) {
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

	// Second draw is skipped (nothing dirty).
	drawn := w.DrawTo(canvas)
	if drawn {
		t.Error("second draw should be skipped")
	}
	// Stats from last actual draw are retained.
	if w.LastDrawStats().DrawnWidgets != 1 {
		t.Error("stats should be retained from last actual draw")
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
