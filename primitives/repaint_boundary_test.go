package primitives_test

import (
	"image"
	"testing"

	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
)

// drawCountingWidget tracks how many times Draw is called.
type drawCountingWidget struct {
	widget.WidgetBase
	drawCount int
}

func newDrawCountingWidget() *drawCountingWidget {
	w := &drawCountingWidget{}
	w.SetVisible(true)
	w.SetEnabled(true)
	w.SetNeedsRedraw(true)
	return w
}

func (w *drawCountingWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(100, 50))
}

func (w *drawCountingWidget) Draw(_ widget.Context, _ widget.Canvas) {
	w.drawCount++
}

func (w *drawCountingWidget) Event(_ widget.Context, _ event.Event) bool { return false }

var _ widget.Widget = (*drawCountingWidget)(nil)

// imageRecordingCanvas records DrawImage and ReplayScene calls for validation.
// ADR-007: RepaintBoundary now uses ReplayScene instead of DrawImage, so both
// are tracked. replaySceneCalls is the primary assertion target.
type imageRecordingCanvas struct {
	mockCanvas
	drawImageCalls   []drawImageCall
	replaySceneCalls []widget.SceneCache
}

type drawImageCall struct {
	img image.Image
	at  geometry.Point
}

func (c *imageRecordingCanvas) DrawImage(img image.Image, at geometry.Point) {
	c.drawImageCalls = append(c.drawImageCalls, drawImageCall{img: img, at: at})
}

func (c *imageRecordingCanvas) ReplayScene(s widget.SceneCache) {
	c.replaySceneCalls = append(c.replaySceneCalls, s)
}

// --- Construction Tests ---

func TestNewRepaintBoundary_NilChild(t *testing.T) {
	rb := primitives.NewRepaintBoundary(nil)
	if rb == nil {
		t.Fatal("NewRepaintBoundary should never return nil")
	}
	if rb.Child() != nil {
		t.Error("expected nil child")
	}
	if rb.Children() != nil {
		t.Error("expected nil children slice for nil child")
	}
}

func TestNewRepaintBoundary_WithChild(t *testing.T) {
	child := primitives.Text("hello")
	rb := primitives.NewRepaintBoundary(child)

	if rb.Child() != child {
		t.Error("expected child to be returned")
	}
	children := rb.Children()
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}
	if children[0] != child {
		t.Error("expected child in Children() slice")
	}
}

func TestNewRepaintBoundary_DefaultState(t *testing.T) {
	rb := primitives.NewRepaintBoundary(primitives.Text("x"))

	if !rb.IsVisible() {
		t.Error("should be visible by default")
	}
	if !rb.IsEnabled() {
		t.Error("should be enabled by default")
	}
	if rb.CacheValid() {
		t.Error("cache should not be valid initially")
	}
	if rb.CacheHits() != 0 {
		t.Error("cache hits should be 0 initially")
	}
	if rb.DebugLabel() != "" {
		t.Error("debug label should be empty by default")
	}
}

func TestNewRepaintBoundary_WithDebugLabel(t *testing.T) {
	rb := primitives.NewRepaintBoundary(nil, primitives.WithDebugLabel("chart"))
	if rb.DebugLabel() != "chart" {
		t.Errorf("expected debug label 'chart', got %q", rb.DebugLabel())
	}
}

// --- Layout Tests ---

func TestRepaintBoundary_Layout_NilChild(t *testing.T) {
	rb := primitives.NewRepaintBoundary(nil)
	constraints := geometry.Tight(geometry.Sz(200, 100))
	size := rb.Layout(nil, constraints)

	if size.Width != 200 || size.Height != 100 {
		t.Errorf("expected tight size 200x100, got %v", size)
	}
}

func TestRepaintBoundary_Layout_DelegatesToChild(t *testing.T) {
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)

	constraints := geometry.BoxConstraints(0, 200, 0, 100)
	size := rb.Layout(nil, constraints)

	// drawCountingWidget returns Constrain(100, 50)
	if size.Width != 100 || size.Height != 50 {
		t.Errorf("expected 100x50, got %v", size)
	}
}

func TestRepaintBoundary_Layout_InvalidatesCacheOnSizeChange(t *testing.T) {
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)

	// First layout
	rb.Layout(nil, geometry.BoxConstraints(0, 200, 0, 100))

	// Force a draw to populate cache
	canvas := &imageRecordingCanvas{}
	rb.Draw(nil, canvas)

	if !rb.CacheValid() {
		t.Error("cache should be valid after first draw")
	}

	// Change constraints to produce different size
	rb.Layout(nil, geometry.Tight(geometry.Sz(50, 25)))

	if rb.CacheValid() {
		t.Error("cache should be invalidated after size change")
	}
}

// --- Draw Tests ---

func TestRepaintBoundary_Draw_Invisible(t *testing.T) {
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)
	rb.SetVisible(false)

	canvas := &imageRecordingCanvas{}
	rb.Draw(nil, canvas)

	if child.drawCount > 0 {
		t.Error("invisible boundary should not draw child")
	}
	if len(canvas.replaySceneCalls) > 0 {
		t.Error("invisible boundary should not call DrawImage")
	}
}

func TestRepaintBoundary_Draw_NilChild(t *testing.T) {
	rb := primitives.NewRepaintBoundary(nil)
	canvas := &imageRecordingCanvas{}
	rb.Draw(nil, canvas)

	if len(canvas.replaySceneCalls) > 0 {
		t.Error("nil child should not call DrawImage")
	}
}

func TestRepaintBoundary_Draw_FirstDrawRendersChild(t *testing.T) {
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)

	rb.Layout(nil, geometry.BoxConstraints(0, 200, 0, 100))

	canvas := &imageRecordingCanvas{}
	rb.Draw(nil, canvas)

	if child.drawCount != 1 {
		t.Errorf("expected child Draw called once, got %d", child.drawCount)
	}
	if len(canvas.replaySceneCalls) != 1 {
		t.Fatalf("expected 1 ReplayScene call, got %d", len(canvas.replaySceneCalls))
	}
	if rb.CacheHits() != 0 {
		t.Error("first draw should not be a cache hit")
	}
	if !rb.CacheValid() {
		t.Error("cache should be valid after draw")
	}
}

func TestRepaintBoundary_Draw_SecondDrawUsesCacheWhenClean(t *testing.T) {
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)

	rb.Layout(nil, geometry.BoxConstraints(0, 200, 0, 100))

	canvas := &imageRecordingCanvas{}
	// First draw: renders child, populates cache.
	rb.Draw(nil, canvas)

	// Child is now clean (ClearRedrawInTree called by RepaintBoundary.Draw).
	// Second draw should use cache.
	canvas2 := &imageRecordingCanvas{}
	rb.Draw(nil, canvas2)

	if child.drawCount != 1 {
		t.Errorf("expected child Draw called once total, got %d", child.drawCount)
	}
	if len(canvas2.replaySceneCalls) != 1 {
		t.Fatalf("expected 1 ReplayScene call on second draw, got %d", len(canvas2.replaySceneCalls))
	}
	if rb.CacheHits() != 1 {
		t.Errorf("expected 1 cache hit, got %d", rb.CacheHits())
	}
}

func TestRepaintBoundary_Draw_DirtyChildInvalidatesCache(t *testing.T) {
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)

	rb.Layout(nil, geometry.BoxConstraints(0, 200, 0, 100))

	canvas := &imageRecordingCanvas{}
	rb.Draw(nil, canvas) // First draw

	// Mark boundary dirty (simulates upward propagation from child).
	rb.MarkBoundaryDirty()

	canvas2 := &imageRecordingCanvas{}
	rb.Draw(nil, canvas2) // Second draw: dirty → re-record.

	if child.drawCount != 2 {
		t.Errorf("expected child Draw called twice, got %d", child.drawCount)
	}
	if rb.CacheHits() != 0 {
		t.Errorf("expected 0 cache hits (dirty child), got %d", rb.CacheHits())
	}
}

func TestRepaintBoundary_Draw_ManualInvalidation(t *testing.T) {
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)

	rb.Layout(nil, geometry.BoxConstraints(0, 200, 0, 100))

	canvas := &imageRecordingCanvas{}
	rb.Draw(nil, canvas) // First draw

	// Manually invalidate cache.
	rb.InvalidateCache()

	if rb.CacheValid() {
		t.Error("cache should be invalid after InvalidateCache")
	}

	canvas2 := &imageRecordingCanvas{}
	rb.Draw(nil, canvas2)

	if child.drawCount != 2 {
		t.Errorf("expected child Draw called twice after manual invalidation, got %d", child.drawCount)
	}
}

func TestRepaintBoundary_Draw_ZeroSizeSkipsRendering(t *testing.T) {
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)

	// Set zero-size bounds.
	rb.SetBounds(geometry.NewRect(0, 0, 0, 0))

	canvas := &imageRecordingCanvas{}
	rb.Draw(nil, canvas)

	if child.drawCount > 0 {
		t.Error("zero-size boundary should not draw child")
	}
	if len(canvas.replaySceneCalls) > 0 {
		t.Error("zero-size boundary should not call DrawImage")
	}
}

func TestRepaintBoundary_Draw_ReplaySceneCalledOnDraw(t *testing.T) {
	// ReplayScene should be called regardless of widget position.
	// Position is embedded in the scene commands, not passed as a parameter.
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)

	rb.Layout(nil, geometry.BoxConstraints(0, 200, 0, 100))
	rb.SetBounds(geometry.NewRect(50, 30, 100, 50))

	canvas := &imageRecordingCanvas{}
	rb.Draw(nil, canvas)

	if len(canvas.replaySceneCalls) != 1 {
		t.Fatalf("expected 1 ReplayScene call, got %d", len(canvas.replaySceneCalls))
	}
	if canvas.replaySceneCalls[0] == nil {
		t.Error("ReplayScene received nil scene")
	}
}

// --- Nested RepaintBoundary Tests ---

func TestRepaintBoundary_Nested(t *testing.T) {
	innerChild := newDrawCountingWidget()
	inner := primitives.NewRepaintBoundary(innerChild)

	outer := primitives.NewRepaintBoundary(inner)
	outer.Layout(nil, geometry.BoxConstraints(0, 200, 0, 100))

	canvas := &imageRecordingCanvas{}
	outer.Draw(nil, canvas) // First draw: both render.

	if innerChild.drawCount != 1 {
		t.Errorf("expected inner child Draw once, got %d", innerChild.drawCount)
	}

	// Second draw: outer serves from cache (inner is also clean).
	canvas2 := &imageRecordingCanvas{}
	outer.Draw(nil, canvas2)

	if innerChild.drawCount != 1 {
		t.Errorf("expected inner child Draw still 1 (outer cached), got %d", innerChild.drawCount)
	}
	if outer.CacheHits() != 1 {
		t.Errorf("expected outer cache hit, got %d", outer.CacheHits())
	}
}

// --- Event Tests ---

func TestRepaintBoundary_Event_DelegatesToChild(t *testing.T) {
	consumed := false
	child := &eventTestWidget{
		onEvent: func() { consumed = true },
	}
	child.SetVisible(true)
	child.SetEnabled(true)
	child.SetBounds(geometry.NewRect(0, 0, 100, 50))

	rb := primitives.NewRepaintBoundary(child)
	rb.SetBounds(geometry.NewRect(10, 10, 100, 50))

	// Send key event (non-mouse, no coordinate translation needed).
	ke := event.NewKeyEvent(event.KeyPress, event.KeyA, 'a', 0)
	result := rb.Event(nil, ke)

	if !consumed {
		t.Error("event should be dispatched to child")
	}
	if !result {
		t.Error("event should be consumed")
	}
}

func TestRepaintBoundary_Event_TranslatesMouseCoordinates(t *testing.T) {
	var receivedPos geometry.Point
	child := &mouseTrackingWidget{
		onMouse: func(pos geometry.Point) { receivedPos = pos },
	}
	child.SetVisible(true)
	child.SetEnabled(true)
	child.SetBounds(geometry.NewRect(0, 0, 100, 50))

	rb := primitives.NewRepaintBoundary(child)
	rb.SetBounds(geometry.NewRect(20, 30, 100, 50))

	pos := geometry.Pt(50, 40)
	me := event.NewMouseEvent(event.MousePress, event.ButtonLeft, 0, pos, pos, 0)
	rb.Event(nil, me)

	// Mouse position should be translated: (50-20, 40-30) = (30, 10)
	if receivedPos.X != 30 || receivedPos.Y != 10 {
		t.Errorf("expected translated position (30,10), got (%v,%v)", receivedPos.X, receivedPos.Y)
	}
}

func TestRepaintBoundary_Event_InvisibleIgnoresEvents(t *testing.T) {
	consumed := false
	child := &eventTestWidget{
		onEvent: func() { consumed = true },
	}
	child.SetVisible(true)
	child.SetEnabled(true)

	rb := primitives.NewRepaintBoundary(child)
	rb.SetVisible(false)

	ke := event.NewKeyEvent(event.KeyPress, event.KeyA, 'a', 0)
	rb.Event(nil, ke)

	if consumed {
		t.Error("invisible boundary should not dispatch events")
	}
}

func TestRepaintBoundary_Event_NilChild(t *testing.T) {
	rb := primitives.NewRepaintBoundary(nil)
	ke := event.NewKeyEvent(event.KeyPress, event.KeyA, 'a', 0)
	result := rb.Event(nil, ke)

	if result {
		t.Error("nil child should not consume events")
	}
}

// --- Unmount Tests ---

func TestRepaintBoundary_Unmount_FreesCache(t *testing.T) {
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)

	rb.Layout(nil, geometry.BoxConstraints(0, 200, 0, 100))

	canvas := &imageRecordingCanvas{}
	rb.Draw(nil, canvas)

	if !rb.CacheValid() {
		t.Error("cache should be valid after draw")
	}

	rb.Unmount()

	if rb.CacheValid() {
		t.Error("cache should be invalid after Unmount")
	}
	if rb.CacheHits() != 0 {
		t.Error("cache hits should be reset after Unmount")
	}
}

// --- Accessibility Tests ---

func TestRepaintBoundary_Accessibility(t *testing.T) {
	rb := primitives.NewRepaintBoundary(nil, primitives.WithDebugLabel("chart"))

	acc, ok := interface{}(rb).(a11y.Accessible)
	if !ok {
		t.Fatal("RepaintBoundary should implement a11y.Accessible")
	}

	if acc.AccessibilityRole() != a11y.RoleGenericContainer {
		t.Errorf("expected RoleGenericContainer, got %v", acc.AccessibilityRole())
	}
	if acc.AccessibilityLabel() != "chart" {
		t.Errorf("expected label 'chart', got %q", acc.AccessibilityLabel())
	}
	if acc.AccessibilityHint() != "" {
		t.Error("expected empty hint")
	}
	if acc.AccessibilityValue() != "" {
		t.Error("expected empty value")
	}

	state := acc.AccessibilityState()
	if state.Disabled {
		t.Error("should not be disabled by default")
	}
	if state.Hidden {
		t.Error("should not be hidden by default")
	}
	if acc.AccessibilityActions() != nil {
		t.Error("should have no actions")
	}
}

// --- DrawStats Integration Tests ---

func TestRepaintBoundary_Draw_CacheHitIncrementsDrawStats(t *testing.T) {
	// When RepaintBoundary serves from cache, it should increment
	// DrawStats.CachedWidgets via the DrawStatsProvider interface.
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)

	rb.Layout(nil, geometry.BoxConstraints(0, 200, 0, 100))

	ctx := widget.NewContext()
	var stats widget.DrawStats
	ctx.SetDrawStats(&stats)

	canvas := &imageRecordingCanvas{}

	// First draw: cache miss → child rendered.
	rb.Draw(ctx, canvas)
	if stats.CachedWidgets != 0 {
		t.Errorf("CachedWidgets = %d after first draw, want 0 (cache miss)", stats.CachedWidgets)
	}

	// Second draw: child is clean → cache hit.
	canvas2 := &imageRecordingCanvas{}
	rb.Draw(ctx, canvas2)
	if stats.CachedWidgets != 1 {
		t.Errorf("CachedWidgets = %d after second draw, want 1 (cache hit)", stats.CachedWidgets)
	}

	// Third draw: still clean → another cache hit.
	canvas3 := &imageRecordingCanvas{}
	rb.Draw(ctx, canvas3)
	if stats.CachedWidgets != 2 {
		t.Errorf("CachedWidgets = %d after third draw, want 2", stats.CachedWidgets)
	}

	ctx.SetDrawStats(nil)
}

func TestRepaintBoundary_Draw_DirtySubtreeDoesNotIncrementCachedWidgets(t *testing.T) {
	// When the boundary is dirty, RepaintBoundary re-records the scene
	// and should NOT increment CachedWidgets.
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)

	rb.Layout(nil, geometry.BoxConstraints(0, 200, 0, 100))

	ctx := widget.NewContext()
	var stats widget.DrawStats
	ctx.SetDrawStats(&stats)

	canvas := &imageRecordingCanvas{}
	rb.Draw(ctx, canvas) // First draw (cache miss).

	// Mark boundary dirty (simulates upward propagation from child).
	rb.MarkBoundaryDirty()

	canvas2 := &imageRecordingCanvas{}
	rb.Draw(ctx, canvas2) // Second draw: dirty → re-record, not cache hit.

	if stats.CachedWidgets != 0 {
		t.Errorf("CachedWidgets = %d, want 0 (dirty boundary should not count as cache hit)",
			stats.CachedWidgets)
	}

	ctx.SetDrawStats(nil)
}

func TestRepaintBoundary_Draw_NilDrawStatsDoesNotPanic(t *testing.T) {
	// When context does not provide DrawStats (nil), cache hit should
	// still work correctly (just no stats recorded).
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)

	rb.Layout(nil, geometry.BoxConstraints(0, 200, 0, 100))

	canvas := &imageRecordingCanvas{}
	rb.Draw(nil, canvas) // First draw: nil ctx.
	rb.Draw(nil, canvas) // Second draw: cache hit, nil ctx → no panic.

	if rb.CacheHits() != 1 {
		t.Errorf("CacheHits = %d, want 1 (cache hit should work without stats)", rb.CacheHits())
	}
}

func TestRepaintBoundary_Draw_NonProviderContextDoesNotPanic(t *testing.T) {
	// When context does not implement DrawStatsProvider, cache hit
	// should still work (just no stats recorded). Uses a plain Context.
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)

	rb.Layout(nil, geometry.BoxConstraints(0, 200, 0, 100))

	// Use ContextImpl but without SetDrawStats → DrawStats() returns nil.
	ctx := widget.NewContext()

	canvas := &imageRecordingCanvas{}
	rb.Draw(ctx, canvas) // First draw.
	rb.Draw(ctx, canvas) // Second draw: cache hit.

	if rb.CacheHits() != 1 {
		t.Errorf("CacheHits = %d, want 1", rb.CacheHits())
	}
}

func TestRepaintBoundary_DrawViaDrawTree_CachedWidgetsPopulated(t *testing.T) {
	// When drawn through widget.DrawTree, the stats should include
	// CachedWidgets from RepaintBoundary cache hits.
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)

	rb.Layout(nil, geometry.BoxConstraints(0, 200, 0, 100))

	ctx := widget.NewContext()
	canvas := &imageRecordingCanvas{}

	// First draw: cache miss.
	stats1 := widget.DrawTree(rb, ctx, canvas)
	if stats1.CachedWidgets != 0 {
		t.Errorf("first DrawTree: CachedWidgets = %d, want 0", stats1.CachedWidgets)
	}

	// Second draw: cache hit → CachedWidgets should be 1.
	canvas2 := &imageRecordingCanvas{}
	stats2 := widget.DrawTree(rb, ctx, canvas2)
	if stats2.CachedWidgets != 1 {
		t.Errorf("second DrawTree: CachedWidgets = %d, want 1", stats2.CachedWidgets)
	}
}

// --- ADR-007 Boundary Dirty Flag Tests ---

func TestRepaintBoundary_Draw_MultipleBoundaries_OneDirty(t *testing.T) {
	// End-to-end test: 12 boundaries, only 1 marked dirty.
	// Only the dirty boundary should re-record; the rest serve cached scenes.
	const numBoundaries = 12
	children := make([]*drawCountingWidget, numBoundaries)
	boundaries := make([]*primitives.RepaintBoundary, numBoundaries)

	for i := range numBoundaries {
		children[i] = newDrawCountingWidget()
		boundaries[i] = primitives.NewRepaintBoundary(children[i])
		boundaries[i].Layout(nil, geometry.BoxConstraints(0, 200, 0, 50))
	}

	// First draw: all boundaries render their children.
	for i := range numBoundaries {
		canvas := &imageRecordingCanvas{}
		boundaries[i].Draw(nil, canvas)
	}

	// Verify all caches are valid and all children drawn exactly once.
	for i := range numBoundaries {
		if !boundaries[i].CacheValid() {
			t.Fatalf("boundary[%d] cache not valid after first draw", i)
		}
		if children[i].drawCount != 1 {
			t.Fatalf("children[%d] drawn %d times, want 1", i, children[i].drawCount)
		}
	}

	// Mark only boundary[3] dirty via MarkBoundaryDirty (upward propagation).
	boundaries[3].MarkBoundaryDirty()

	ctx := widget.NewContext()
	var stats widget.DrawStats
	ctx.SetDrawStats(&stats)

	// Second draw: all boundaries.
	for i := range numBoundaries {
		canvas := &imageRecordingCanvas{}
		boundaries[i].Draw(ctx, canvas)
	}

	// Verify: only child[3] was re-rendered.
	for i := range numBoundaries {
		if i == 3 {
			if children[i].drawCount != 2 {
				t.Errorf("children[%d] drawn %d times, want 2 (dirty boundary)", i, children[i].drawCount)
			}
		} else {
			if children[i].drawCount != 1 {
				t.Errorf("children[%d] drawn %d times, want 1 (clean boundary)", i, children[i].drawCount)
			}
		}
	}

	// CachedWidgets should be numBoundaries - 1 (all except the dirty one).
	if stats.CachedWidgets != numBoundaries-1 {
		t.Errorf("CachedWidgets = %d, want %d", stats.CachedWidgets, numBoundaries-1)
	}

	ctx.SetDrawStats(nil)
}

// --- Helper test widgets ---

type eventTestWidget struct {
	widget.WidgetBase
	onEvent func()
}

func (w *eventTestWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(100, 50))
}

func (w *eventTestWidget) Draw(_ widget.Context, _ widget.Canvas) {}

func (w *eventTestWidget) Event(_ widget.Context, _ event.Event) bool {
	if w.onEvent != nil {
		w.onEvent()
	}
	return true
}

var _ widget.Widget = (*eventTestWidget)(nil)

type mouseTrackingWidget struct {
	widget.WidgetBase
	onMouse func(geometry.Point)
}

func (w *mouseTrackingWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(100, 50))
}

func (w *mouseTrackingWidget) Draw(_ widget.Context, _ widget.Canvas) {}

func (w *mouseTrackingWidget) Event(_ widget.Context, e event.Event) bool {
	if me, ok := e.(*event.MouseEvent); ok {
		if w.onMouse != nil {
			w.onMouse(me.Position)
		}
		return true
	}
	return false
}

var _ widget.Widget = (*mouseTrackingWidget)(nil)

// --- GPU Texture Compositing Tests ---

func TestRepaintBoundary_Draw_SceneRecordAndReplay(t *testing.T) {
	// Verify that the scene-based cache records on miss and replays on hit.
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)
	rb.Layout(nil, geometry.BoxConstraints(0, 200, 0, 100))

	canvas := &imageRecordingCanvas{}
	rb.Draw(nil, canvas)

	if child.drawCount != 1 {
		t.Errorf("child drawn %d times, want 1", child.drawCount)
	}
	if len(canvas.replaySceneCalls) != 1 {
		t.Errorf("ReplayScene called %d times, want 1", len(canvas.replaySceneCalls))
	}
	if !rb.CacheValid() {
		t.Error("cache should be valid after scene recording")
	}
}

func TestRepaintBoundary_Draw_SceneCacheHitWorks(t *testing.T) {
	// Cache hit should replay the same scene without re-drawing child.
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)
	rb.Layout(nil, geometry.BoxConstraints(0, 200, 0, 100))

	canvas := &imageRecordingCanvas{}
	rb.Draw(nil, canvas) // Cache miss.

	canvas2 := &imageRecordingCanvas{}
	rb.Draw(nil, canvas2) // Cache hit.

	if child.drawCount != 1 {
		t.Errorf("child drawn %d times, want 1 (cache hit)", child.drawCount)
	}
	if len(canvas2.replaySceneCalls) != 1 {
		t.Errorf("ReplayScene called %d times, want 1 (cache hit)", len(canvas2.replaySceneCalls))
	}
	if rb.CacheHits() != 1 {
		t.Errorf("CacheHits = %d, want 1", rb.CacheHits())
	}
}

func TestRepaintBoundary_Unmount_ClearsSceneCache(t *testing.T) {
	// Unmount should clear all scene cache state.
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)
	rb.Layout(nil, geometry.BoxConstraints(0, 200, 0, 100))

	canvas := &imageRecordingCanvas{}
	rb.Draw(nil, canvas) // Populate cache.

	if !rb.CacheValid() {
		t.Fatal("cache should be valid before Unmount")
	}

	rb.Unmount()

	if rb.CacheValid() {
		t.Error("cache should be invalid after Unmount")
	}
	if rb.CacheHits() != 0 {
		t.Error("cache hits should be 0 after Unmount")
	}
}

func TestRepaintBoundary_Layout_SizeChange_InvalidatesSceneCache(t *testing.T) {
	// When size changes, the scene cache should be invalidated.
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)

	// First layout + draw.
	rb.Layout(nil, geometry.BoxConstraints(0, 200, 0, 100))
	canvas := &imageRecordingCanvas{}
	rb.Draw(nil, canvas)
	if !rb.CacheValid() {
		t.Fatal("cache should be valid after first draw")
	}

	// Layout with different size.
	rb.Layout(nil, geometry.Tight(geometry.Sz(50, 25)))
	if rb.CacheValid() {
		t.Error("cache should be invalidated after size change")
	}

	// Draw after size change should re-record.
	canvas2 := &imageRecordingCanvas{}
	rb.Draw(nil, canvas2)
	if child.drawCount != 2 {
		t.Errorf("child drawn %d times, want 2 (re-record after size change)", child.drawCount)
	}
}

// --- MarkBoundaryDirty Tests (ADR-007 Task 1d) ---

func TestRepaintBoundary_MarkBoundaryDirty(t *testing.T) {
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)

	if rb.IsBoundaryDirty() {
		t.Error("new RepaintBoundary should not be dirty")
	}

	rb.MarkBoundaryDirty()

	if !rb.IsBoundaryDirty() {
		t.Error("should be dirty after MarkBoundaryDirty")
	}
	if rb.CacheValid() {
		t.Error("cache should be invalidated after MarkBoundaryDirty")
	}
}

func TestRepaintBoundary_MarkBoundaryDirty_O1Guard(t *testing.T) {
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)

	callCount := 0
	rb.SetOnBoundaryDirty(func(_ *primitives.RepaintBoundary) {
		callCount++
	})

	rb.MarkBoundaryDirty()
	rb.MarkBoundaryDirty() // Second call — should be O(1) no-op.

	if callCount != 1 {
		t.Errorf("onBoundaryDirty should be called once (O(1) guard), got %d", callCount)
	}
}

func TestRepaintBoundary_ClearBoundaryDirty(t *testing.T) {
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)

	rb.MarkBoundaryDirty()
	rb.ClearBoundaryDirty()

	if rb.IsBoundaryDirty() {
		t.Error("should not be dirty after ClearBoundaryDirty")
	}
}

func TestRepaintBoundary_OnBoundaryDirtyCallback(t *testing.T) {
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)

	var notifiedRB *primitives.RepaintBoundary
	rb.SetOnBoundaryDirty(func(r *primitives.RepaintBoundary) {
		notifiedRB = r
	})

	rb.MarkBoundaryDirty()

	if notifiedRB != rb {
		t.Error("callback should receive the RepaintBoundary that was marked dirty")
	}
}

func TestRepaintBoundary_MarkClearRemark(t *testing.T) {
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)

	callCount := 0
	rb.SetOnBoundaryDirty(func(_ *primitives.RepaintBoundary) {
		callCount++
	})

	rb.MarkBoundaryDirty()
	rb.ClearBoundaryDirty()
	rb.MarkBoundaryDirty() // Should fire callback again after clear.

	if callCount != 2 {
		t.Errorf("callback should fire on each clean→dirty transition, got %d", callCount)
	}
}

// TestRepaintBoundary_ImplementsRepaintBoundaryMarker verifies the compile-time
// interface check for widget.RepaintBoundaryMarker.
func TestRepaintBoundary_ImplementsRepaintBoundaryMarker(t *testing.T) {
	var rb interface{} = &primitives.RepaintBoundary{}
	if _, ok := rb.(widget.RepaintBoundaryMarker); !ok {
		t.Error("RepaintBoundary should implement widget.RepaintBoundaryMarker")
	}
}
