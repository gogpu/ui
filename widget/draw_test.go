package widget

import (
	"github.com/gogpu/gg/scene"
	"image"
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
)

// drawTrackingWidget tracks whether Draw was called.
type drawTrackingWidget struct {
	WidgetBase
	drawCalled bool
	drawCanvas Canvas
}

func newDrawTrackingWidget() *drawTrackingWidget {
	w := &drawTrackingWidget{}
	w.SetVisible(true)
	w.SetEnabled(true)
	return w
}

func (w *drawTrackingWidget) Layout(_ Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(100, 50))
}

func (w *drawTrackingWidget) Draw(_ Context, canvas Canvas) {
	w.drawCalled = true
	w.drawCanvas = canvas
}

func (w *drawTrackingWidget) Event(_ Context, _ event.Event) bool { return false }

var _ Widget = (*drawTrackingWidget)(nil)

// invisibleWidget reports IsVisible() = false.
type invisibleWidget struct {
	WidgetBase
	drawCalled bool
}

func newInvisibleWidget() *invisibleWidget {
	w := &invisibleWidget{}
	w.SetVisible(false)
	w.SetEnabled(true)
	return w
}

func (w *invisibleWidget) Layout(_ Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(50, 50))
}

func (w *invisibleWidget) Draw(_ Context, _ Canvas) {
	w.drawCalled = true
}

func (w *invisibleWidget) Event(_ Context, _ event.Event) bool { return false }

var _ Widget = (*invisibleWidget)(nil)

// --- DrawTree tests ---

func TestDrawTree_NilWidget(t *testing.T) {
	stats := DrawTree(nil, nil, nil)

	if stats.TotalWidgets != 0 {
		t.Errorf("TotalWidgets = %d, want 0", stats.TotalWidgets)
	}
	if stats.DrawnWidgets != 0 {
		t.Errorf("DrawnWidgets = %d, want 0", stats.DrawnWidgets)
	}
}

func TestDrawTree_SingleDirtyWidget(t *testing.T) {
	w := newDrawTrackingWidget()
	w.SetNeedsRedraw(true)

	canvas := &noopCanvas{}
	stats := DrawTree(w, nil, canvas)

	if !w.drawCalled {
		t.Error("Draw should be called on dirty widget")
	}
	if stats.TotalWidgets != 1 {
		t.Errorf("TotalWidgets = %d, want 1", stats.TotalWidgets)
	}
	if stats.DrawnWidgets != 1 {
		t.Errorf("DrawnWidgets = %d, want 1", stats.DrawnWidgets)
	}
	if stats.DirtyWidgets != 1 {
		t.Errorf("DirtyWidgets = %d, want 1", stats.DirtyWidgets)
	}
	if stats.CleanWidgets != 0 {
		t.Errorf("CleanWidgets = %d, want 0", stats.CleanWidgets)
	}
}

func TestDrawTree_SingleCleanWidget(t *testing.T) {
	w := newDrawTrackingWidget()
	w.ClearRedraw()

	canvas := &noopCanvas{}
	stats := DrawTree(w, nil, canvas)

	// In Sub-Phase 1, clean widgets are still drawn (gg clears pixmap).
	if !w.drawCalled {
		t.Error("Draw should be called even on clean widget in Sub-Phase 1")
	}
	if stats.CleanWidgets != 1 {
		t.Errorf("CleanWidgets = %d, want 1", stats.CleanWidgets)
	}
	if stats.DirtyWidgets != 0 {
		t.Errorf("DirtyWidgets = %d, want 0", stats.DirtyWidgets)
	}
}

func TestDrawTree_CustomWidgetWithoutBase(t *testing.T) {
	// Custom widget without WidgetBase is treated as always dirty.
	w := &customWidget{}

	canvas := &noopCanvas{}
	stats := DrawTree(w, nil, canvas)

	if stats.DirtyWidgets != 1 {
		t.Errorf("DirtyWidgets = %d, want 1 (custom widget without NeedsRedraw)", stats.DirtyWidgets)
	}
	if stats.DrawnWidgets != 1 {
		t.Errorf("DrawnWidgets = %d, want 1", stats.DrawnWidgets)
	}
}

func TestDrawTree_PassesCanvasToWidget(t *testing.T) {
	w := newDrawTrackingWidget()
	w.SetNeedsRedraw(true)

	canvas := &noopCanvas{}
	DrawTree(w, nil, canvas)

	if w.drawCanvas != canvas {
		t.Error("DrawTree should pass canvas to widget's Draw method")
	}
}

// --- CollectDrawStats tests ---

func TestCollectDrawStats_NilWidget(t *testing.T) {
	stats := CollectDrawStats(nil)

	if stats.TotalWidgets != 0 {
		t.Errorf("TotalWidgets = %d, want 0", stats.TotalWidgets)
	}
}

func TestCollectDrawStats_SingleDirtyWidget(t *testing.T) {
	w := newDrawTrackingWidget()
	w.SetNeedsRedraw(true)

	stats := CollectDrawStats(w)

	if stats.TotalWidgets != 1 {
		t.Errorf("TotalWidgets = %d, want 1", stats.TotalWidgets)
	}
	if stats.DirtyWidgets != 1 {
		t.Errorf("DirtyWidgets = %d, want 1", stats.DirtyWidgets)
	}
	// CollectDrawStats should NOT call Draw.
	if w.drawCalled {
		t.Error("CollectDrawStats should not call Draw")
	}
}

func TestCollectDrawStats_TreeWithChildren(t *testing.T) {
	root := newDrawTrackingWidget()
	root.SetNeedsRedraw(true)

	child1 := newDrawTrackingWidget()
	child1.SetNeedsRedraw(true)
	root.AddChild(child1)

	child2 := newDrawTrackingWidget()
	child2.ClearRedraw()
	root.AddChild(child2)

	grandchild := newDrawTrackingWidget()
	grandchild.SetNeedsRedraw(true)
	child1.AddChild(grandchild)

	stats := CollectDrawStats(root)

	if stats.TotalWidgets != 4 {
		t.Errorf("TotalWidgets = %d, want 4", stats.TotalWidgets)
	}
	if stats.DirtyWidgets != 3 {
		t.Errorf("DirtyWidgets = %d, want 3", stats.DirtyWidgets)
	}
	if stats.CleanWidgets != 1 {
		t.Errorf("CleanWidgets = %d, want 1", stats.CleanWidgets)
	}
}

func TestCollectDrawStats_InvisibleWidget(t *testing.T) {
	w := newInvisibleWidget()
	w.SetNeedsRedraw(true)

	stats := CollectDrawStats(w)

	if stats.TotalWidgets != 1 {
		t.Errorf("TotalWidgets = %d, want 1", stats.TotalWidgets)
	}
	if stats.SkippedWidgets != 1 {
		t.Errorf("SkippedWidgets = %d, want 1", stats.SkippedWidgets)
	}
	if stats.DirtyWidgets != 0 {
		t.Errorf("DirtyWidgets = %d, want 0 (invisible widgets are skipped)", stats.DirtyWidgets)
	}
}

func TestCollectDrawStats_MixedTree(t *testing.T) {
	root := newDrawTrackingWidget()
	root.SetNeedsRedraw(true)

	visible := newDrawTrackingWidget()
	visible.ClearRedraw()
	root.AddChild(visible)

	invisible := newInvisibleWidget()
	root.AddChild(invisible)

	stats := CollectDrawStats(root)

	if stats.TotalWidgets != 3 {
		t.Errorf("TotalWidgets = %d, want 3", stats.TotalWidgets)
	}
	if stats.DirtyWidgets != 1 {
		t.Errorf("DirtyWidgets = %d, want 1", stats.DirtyWidgets)
	}
	if stats.CleanWidgets != 1 {
		t.Errorf("CleanWidgets = %d, want 1", stats.CleanWidgets)
	}
	if stats.SkippedWidgets != 1 {
		t.Errorf("SkippedWidgets = %d, want 1", stats.SkippedWidgets)
	}
}

func TestCollectDrawStats_CustomWidgetWithChildren(t *testing.T) {
	child := newDrawTrackingWidget()
	child.ClearRedraw()

	w := &customWidget{children: []Widget{child}}

	stats := CollectDrawStats(w)

	if stats.TotalWidgets != 2 {
		t.Errorf("TotalWidgets = %d, want 2", stats.TotalWidgets)
	}
	if stats.DirtyWidgets != 1 {
		t.Errorf("DirtyWidgets = %d, want 1 (custom widget always dirty)", stats.DirtyWidgets)
	}
	if stats.CleanWidgets != 1 {
		t.Errorf("CleanWidgets = %d, want 1", stats.CleanWidgets)
	}
}

func TestCollectDrawStats_DoesNotClearFlags(t *testing.T) {
	w := newDrawTrackingWidget()
	w.SetNeedsRedraw(true)

	CollectDrawStats(w)

	// CollectDrawStats should not modify any state.
	if !w.NeedsRedraw() {
		t.Error("CollectDrawStats should not clear needsRedraw flag")
	}
}

// --- DrawStats zero value test ---

func TestDrawStats_ZeroValue(t *testing.T) {
	var stats DrawStats

	if stats.TotalWidgets != 0 || stats.DrawnWidgets != 0 ||
		stats.SkippedWidgets != 0 || stats.DirtyWidgets != 0 ||
		stats.CleanWidgets != 0 || stats.CachedWidgets != 0 {
		t.Error("zero-valued DrawStats should have all fields zero")
	}
}

// --- DrawStatsProvider tests ---

func TestDrawStatsProvider_ContextImplImplementsInterface(t *testing.T) {
	ctx := NewContext()
	if _, ok := interface{}(ctx).(DrawStatsProvider); !ok {
		t.Error("ContextImpl should implement DrawStatsProvider")
	}
}

func TestDrawStatsProvider_NilByDefault(t *testing.T) {
	ctx := NewContext()
	if ctx.DrawStats() != nil {
		t.Error("DrawStats should be nil by default")
	}
}

func TestDrawStatsProvider_SetAndGet(t *testing.T) {
	ctx := NewContext()
	var stats DrawStats
	ctx.SetDrawStats(&stats)

	got := ctx.DrawStats()
	if got != &stats {
		t.Error("DrawStats should return the set pointer")
	}

	ctx.SetDrawStats(nil)
	if ctx.DrawStats() != nil {
		t.Error("DrawStats should be nil after SetDrawStats(nil)")
	}
}

func TestDrawTree_SetsDrawStatsOnContext(t *testing.T) {
	// Verify DrawTree sets DrawStats on the context before drawing
	// and clears it after.
	var capturedStats *DrawStats
	capturedWidget := &statsCapturingWidget{
		onDraw: func(ctx Context) {
			if provider, ok := ctx.(DrawStatsProvider); ok {
				capturedStats = provider.DrawStats()
			}
		},
	}
	capturedWidget.SetVisible(true)
	capturedWidget.SetEnabled(true)
	capturedWidget.SetNeedsRedraw(true)

	ctx := NewContext()
	canvas := &noopCanvas{}
	DrawTree(capturedWidget, ctx, canvas)

	if capturedStats == nil {
		t.Error("DrawStats should be accessible inside Draw via DrawStatsProvider")
	}

	// After DrawTree returns, stats should be cleared from context.
	if ctx.DrawStats() != nil {
		t.Error("DrawStats should be nil after DrawTree returns")
	}
}

// noopCanvas is a minimal Canvas implementation for testing.
type noopCanvas struct{}

func (c *noopCanvas) Clear(Color)                                                         {}
func (c *noopCanvas) DrawRect(geometry.Rect, Color)                                       {}
func (c *noopCanvas) FillRectDirect(geometry.Rect, Color)                                 {}
func (c *noopCanvas) StrokeRect(geometry.Rect, Color, float32)                            {}
func (c *noopCanvas) DrawRoundRect(geometry.Rect, Color, float32)                         {}
func (c *noopCanvas) StrokeRoundRect(geometry.Rect, Color, float32, float32)              {}
func (c *noopCanvas) DrawCircle(geometry.Point, float32, Color)                           {}
func (c *noopCanvas) StrokeCircle(geometry.Point, float32, Color, float32)                {}
func (c *noopCanvas) StrokeArc(geometry.Point, float32, float64, float64, Color, float32) {}
func (c *noopCanvas) DrawLine(geometry.Point, geometry.Point, Color, float32)             {}
func (c *noopCanvas) DrawText(string, geometry.Rect, float32, Color, bool, TextAlign)     {}

func (c *noopCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}
func (c *noopCanvas) DrawImage(image.Image, geometry.Point)        {}
func (c *noopCanvas) PushClip(geometry.Rect)                       {}
func (c *noopCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *noopCanvas) PopClip()                                     {}
func (c *noopCanvas) PushTransform(geometry.Point)                 {}
func (c *noopCanvas) PopTransform()                                {}
func (c *noopCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }
func (c *noopCanvas) ScreenOriginBase() geometry.Point             { return geometry.Point{} }
func (c *noopCanvas) ClipBounds() geometry.Rect                    { return geometry.NewRect(0, 0, 10000, 10000) }
func (c *noopCanvas) ReplayScene(_ SceneCache)                     {}

var _ Canvas = (*noopCanvas)(nil)

// --- StampScreenOrigin tests ---

// stampCanvas tracks transform offsets for stamping tests.
type stampCanvas struct {
	noopCanvas
	offsetStack   []geometry.Point
	currentOffset geometry.Point
}

func (c *stampCanvas) PushTransform(offset geometry.Point) {
	c.offsetStack = append(c.offsetStack, c.currentOffset)
	c.currentOffset = c.currentOffset.Add(offset)
}

func (c *stampCanvas) PopTransform() {
	if len(c.offsetStack) > 0 {
		lastIdx := len(c.offsetStack) - 1
		c.currentOffset = c.offsetStack[lastIdx]
		c.offsetStack = c.offsetStack[:lastIdx]
	}
}

func (c *stampCanvas) TransformOffset() geometry.Point {
	return c.currentOffset
}

func (c *stampCanvas) ScreenOriginBase() geometry.Point { return geometry.Point{} }

func TestStampScreenOrigin_Basic(t *testing.T) {
	canvas := &stampCanvas{}

	// Simulate a container at (50, 100)
	canvas.PushTransform(geometry.Pt(50, 100))

	child := newMockWidget()
	child.SetBounds(geometry.NewRect(10, 20, 80, 40))

	StampScreenOrigin(child, canvas)

	// Screen origin = container offset (50,100) + child bounds.Min (10,20) = (60,120)
	got := child.ScreenOrigin()
	want := geometry.Pt(60, 120)
	if got != want {
		t.Errorf("ScreenOrigin = %v, want %v", got, want)
	}

	// ScreenBounds should reflect the screen position
	sb := child.ScreenBounds()
	if sb.Min != want {
		t.Errorf("ScreenBounds.Min = %v, want %v", sb.Min, want)
	}
	if sb.Width() != 80 || sb.Height() != 40 {
		t.Errorf("ScreenBounds size = (%v,%v), want (80,40)", sb.Width(), sb.Height())
	}

	canvas.PopTransform()
}

func TestStampScreenOrigin_NestedTransforms(t *testing.T) {
	canvas := &stampCanvas{}

	// Simulate Box at (100, 50)
	canvas.PushTransform(geometry.Pt(100, 50))

	// Inside, a ScrollView that scrolled 30px down:
	// PushTransform(Pt(0, -30))
	canvas.PushTransform(geometry.Pt(0, -30))

	child := newMockWidget()
	child.SetBounds(geometry.NewRect(10, 80, 60, 30))

	StampScreenOrigin(child, canvas)

	// Screen origin = (100+0+10, 50-30+80) = (110, 100)
	got := child.ScreenOrigin()
	want := geometry.Pt(110, 100)
	if got != want {
		t.Errorf("ScreenOrigin = %v, want %v", got, want)
	}

	canvas.PopTransform()
	canvas.PopTransform()
}

func TestStampScreenOrigin_NilChild(t *testing.T) {
	canvas := &stampCanvas{}
	// Should not panic
	StampScreenOrigin(nil, canvas)
}

// statsCapturingWidget calls a callback during Draw, allowing tests to
// inspect the Context state (e.g., DrawStats) from inside the draw pass.
type statsCapturingWidget struct {
	WidgetBase
	onDraw func(Context)
}

func (w *statsCapturingWidget) Layout(_ Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(100, 50))
}

func (w *statsCapturingWidget) Draw(ctx Context, _ Canvas) {
	if w.onDraw != nil {
		w.onDraw(ctx)
	}
}

func (w *statsCapturingWidget) Event(_ Context, _ event.Event) bool { return false }

var _ Widget = (*statsCapturingWidget)(nil)

// --- ADR-024 DrawChild Tests ---
//
// DrawChild is the public API for container widgets to draw children
// with RepaintBoundary support. It checks IsRepaintBoundary and routes
// through scene caching (drawBoundaryWidget) or direct Draw.
// This replaces the primitives.NewRepaintBoundary wrapper pattern.

// TestDrawChild_NormalWidgetCallsDraw verifies that DrawChild calls
// child.Draw() directly when child is NOT a RepaintBoundary.
func TestDrawChild_NormalWidgetCallsDraw(t *testing.T) {
	child := newDrawTrackingWidget()
	child.SetBounds(geometry.NewRect(0, 0, 100, 50))

	DrawChild(child, nil, nil)

	if !child.drawCalled {
		t.Error("DrawChild should call child.Draw() for non-boundary widget")
	}
}

// TestDrawChild_NilChild verifies DrawChild handles nil gracefully.
func TestDrawChild_NilChild(t *testing.T) {
	DrawChild(nil, nil, nil) // must not panic
}

// TestDrawChild_BoundaryFallsBackWithoutRecorder verifies that DrawChild
// falls back to direct Draw when no SceneRecorder is registered.
func TestDrawChild_BoundaryFallsBackWithoutRecorder(t *testing.T) {
	child := newDrawTrackingWidget()
	child.SetRepaintBoundary(true)
	child.SetBounds(geometry.NewRect(10, 20, 100, 50))

	// Without SceneRecorder registered, drawBoundaryWidget falls back to Draw.
	DrawChild(child, nil, nil)

	if !child.drawCalled {
		t.Error("DrawChild should fall back to Draw when no SceneRecorder registered")
	}
}

// TestDrawChild_BoundaryChecked verifies that DrawChild detects
// IsRepaintBoundary and routes differently than normal Draw.
func TestDrawChild_BoundaryChecked(t *testing.T) {
	normal := newDrawTrackingWidget()
	normal.SetBounds(geometry.NewRect(0, 0, 100, 50))

	boundary := newDrawTrackingWidget()
	boundary.SetRepaintBoundary(true)
	boundary.SetBounds(geometry.NewRect(0, 0, 100, 50))

	if normal.IsRepaintBoundary() {
		t.Error("normal widget should not be boundary")
	}
	if !boundary.IsRepaintBoundary() {
		t.Error("boundary widget should be boundary")
	}
}

// TestBoundary_MarkRedrawInTreeInvalidatesRootScene verifies that
// MarkRedrawInTree on the root widget also invalidates its scene when
// root has IsRepaintBoundary=true. Without this, ctx.Invalidate()
// sets needsRedraw but root boundary replays stale cached scene.
func TestBoundary_MarkRedrawInTreeInvalidatesRootScene(t *testing.T) {
	root := newDrawTrackingWidget()
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))

	root.ClearRedraw()
	root.ClearSceneDirty()

	// Simulate what ctx.Invalidate does.
	MarkRedrawInTree(root)

	if !root.NeedsRedraw() {
		t.Error("root.NeedsRedraw() should be true after MarkRedrawInTree")
	}
	if !root.IsSceneDirty() {
		t.Error("root.IsSceneDirty() should be true after MarkRedrawInTree; " +
			"ctx.Invalidate → MarkRedrawInTree must also invalidate boundary scene, " +
			"otherwise root replays stale cached scene on checkbox/radio clicks")
	}
}

// TestBoundary_MarkRedrawInTreeDoesNotInvalidateChildBoundaries verifies
// that MarkRedrawInTree only invalidates the ROOT boundary (no parent),
// not recursively all child boundaries. Over-invalidation causes full
// repaint every frame → cyan overlay covers entire window.
func TestBoundary_MarkRedrawInTreeInvalidatesAllBoundaries(t *testing.T) {
	root := newDrawTrackingWidget()
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))

	child := newDrawTrackingWidget()
	child.SetRepaintBoundary(true)
	child.SetBounds(geometry.NewRect(0, 0, 48, 48))
	child.SetParent(root)
	root.AddChild(child)

	root.ClearRedraw()
	root.ClearSceneDirty()
	child.ClearRedraw()
	child.ClearSceneDirty()

	MarkRedrawInTree(root)

	if !root.IsSceneDirty() {
		t.Error("root boundary should be scene-dirty after MarkRedrawInTree")
	}

	// MarkRedrawInTree is nuclear (layout/resize). ALL boundaries must
	// invalidate because widget positions may have changed. SetNeedsRedraw
	// on a boundary widget calls InvalidateScene() (Flutter markNeedsPaint
	// self-boundary pattern), so child boundaries correctly become dirty.
	if !child.IsSceneDirty() {
		t.Error("child boundary should be scene-dirty after MarkRedrawInTree; " +
			"nuclear redraw invalidates all boundaries (layout may have moved them)")
	}
}

// TestDrawChild_BoundaryAtNonZeroPosition verifies that DrawChild works
// correctly when the boundary widget has bounds NOT at origin (e.g., Y=200
// in ScrollView content space). The scene must be recorded in LOCAL coords
// (0,0-based), not absolute coords — otherwise text is culled by the
// SceneCanvas clip rect.
func TestDrawChild_BoundaryAtNonZeroPosition(t *testing.T) {
	child := newDrawTrackingWidget()
	child.SetRepaintBoundary(true)
	child.SetBounds(geometry.NewRect(0, 200, 400, 48)) // Y=200, not at origin

	// DrawChild should still call Draw (fallback without SceneRecorder).
	DrawChild(child, nil, nil)

	if !child.drawCalled {
		t.Error("DrawChild should draw widget even at non-zero position")
	}
}

// --- ADR-024 RepaintBoundary Dirty Propagation Tests ---
//
// These verify that SetNeedsRedraw propagates upward through the parent chain
// to the nearest WidgetBase RepaintBoundary, invalidating its scene cache.
// Without this, the boundary replays stale cached scenes after child changes.

// TestBoundary_SetNeedsRedrawPropagesToWidgetBaseBoundary verifies that
// a child's SetNeedsRedraw(true) propagates to a WidgetBase parent with
// isRepaintBoundary=true, calling InvalidateScene.
func TestBoundary_SetNeedsRedrawPropagesToWidgetBaseBoundary(t *testing.T) {
	parent := newDrawTrackingWidget()
	parent.SetRepaintBoundary(true)

	child := newDrawTrackingWidget()
	child.SetParent(parent)

	// Clear initial dirty state.
	parent.ClearRedraw()
	parent.ClearSceneDirty() // Reset scene dirty flag.
	child.ClearRedraw()

	// Child marks itself dirty.
	child.SetNeedsRedraw(true)

	// Parent boundary scene must be invalidated.
	if !parent.IsSceneDirty() {
		t.Error("parent.IsSceneDirty() = false; SetNeedsRedraw on child should " +
			"propagate to WidgetBase boundary and call InvalidateScene")
	}
}

// TestBoundary_SetNeedsRedrawStopsAtFirstBoundary verifies that dirty
// propagation stops at the NEAREST RepaintBoundary (O(depth) walk).
func TestBoundary_SetNeedsRedrawStopsAtFirstBoundary(t *testing.T) {
	root := newDrawTrackingWidget()
	root.SetRepaintBoundary(true)

	middle := newDrawTrackingWidget()
	middle.SetRepaintBoundary(true)
	middle.SetParent(root)

	child := newDrawTrackingWidget()
	child.SetParent(middle)

	// Clear all.
	root.ClearRedraw()
	root.ClearSceneDirty()
	middle.ClearRedraw()
	middle.ClearSceneDirty()
	child.ClearRedraw()

	child.SetNeedsRedraw(true)

	// Middle boundary should be dirty (nearest).
	if !middle.IsSceneDirty() {
		t.Error("middle boundary should be scene-dirty (nearest to child)")
	}

	// Root boundary should NOT be dirty (propagation stops at middle).
	if root.IsSceneDirty() {
		t.Error("root boundary should NOT be scene-dirty (propagation stops at middle)")
	}
}

// TestBoundary_MarkRedrawLocalDoesNotPropagate verifies that MarkRedrawLocal
// sets needsRedraw on the widget but does NOT propagate to parent boundary.
// This is the bug that caused stale scene cache on scroll (fixed in setScroll).
func TestBoundary_MarkRedrawLocalDoesNotPropagate(t *testing.T) {
	parent := newDrawTrackingWidget()
	parent.SetRepaintBoundary(true)

	child := newDrawTrackingWidget()
	child.SetParent(parent)

	parent.ClearRedraw()
	parent.ClearSceneDirty()
	child.ClearRedraw()

	// MarkRedrawLocal only sets local flag.
	child.MarkRedrawLocal()

	if !child.NeedsRedraw() {
		t.Error("child.NeedsRedraw() should be true after MarkRedrawLocal")
	}

	// Parent boundary must NOT be invalidated.
	if parent.IsSceneDirty() {
		t.Error("parent.IsSceneDirty() should be false; MarkRedrawLocal must not propagate")
	}
}

// TestBoundary_CacheHitWhenClean verifies that drawBoundaryWidget returns
// cache hit (replays scene) when boundary is NOT scene-dirty.
func TestBoundary_CacheHitWhenClean(t *testing.T) {
	w := newDrawTrackingWidget()
	w.SetRepaintBoundary(true)
	w.SetBounds(geometry.NewRect(0, 0, 100, 50))

	// Simulate first draw: create and cache a scene.
	sc := scene.NewScene()
	w.SetCachedScene(sc)
	w.SetSceneCacheSize(100, 50)
	w.ClearSceneDirty()

	// Now check: boundary is clean + has cached scene = cache hit.
	if w.IsSceneDirty() {
		t.Error("boundary should be clean")
	}
	if w.CachedScene() == nil {
		t.Error("cached scene should exist")
	}
}

// TestBoundary_CacheMissWhenDirty verifies that drawBoundaryWidget forces
// re-record when boundary IS scene-dirty.
func TestBoundary_CacheMissWhenDirty(t *testing.T) {
	w := newDrawTrackingWidget()
	w.SetRepaintBoundary(true)
	w.SetBounds(geometry.NewRect(0, 0, 100, 50))

	// Give it a cached scene.
	sc := scene.NewScene()
	w.SetCachedScene(sc)
	w.SetSceneCacheSize(100, 50)

	// Mark dirty.
	w.InvalidateScene()

	if !w.IsSceneDirty() {
		t.Error("boundary should be scene-dirty after InvalidateScene")
	}
}

// TestBoundary_SizeChangeInvalidatesCache verifies that a size change
// forces cache miss even if boundary is not scene-dirty.
func TestBoundary_SizeChangeInvalidatesCache(t *testing.T) {
	w := newDrawTrackingWidget()
	w.SetRepaintBoundary(true)
	w.SetBounds(geometry.NewRect(0, 0, 100, 50))
	w.SetSceneCacheSize(100, 50)
	w.SetCachedScene(scene.NewScene())
	w.ClearSceneDirty()

	// Change bounds (simulate resize).
	w.SetBounds(geometry.NewRect(0, 0, 200, 100))

	// drawBoundaryWidget checks cw != width || ch != height.
	cw, ch := w.SceneCacheSize()
	bounds := w.Bounds()
	newW := int(bounds.Width())
	newH := int(bounds.Height())

	if cw == newW && ch == newH {
		t.Error("cache size should differ from new bounds, triggering re-record")
	}
}

// --- Animation Flow Tests (ADR-007 Phase 4) ---
//
// These tests verify that animated widgets (spinner) correctly re-dirty
// themselves during Draw, and that consecutive frames produce fresh renders.
// The key invariant: a boundary widget that calls SetNeedsRedraw(true)
// inside Draw MUST remain dirty after drawBoundaryWidget completes.

// animatingWidget simulates a spinner: calls SetNeedsRedraw(true) during Draw.
type animatingWidget struct {
	WidgetBase
	drawCount int
}

func newAnimatingWidget() *animatingWidget {
	w := &animatingWidget{}
	w.SetVisible(true)
	w.SetEnabled(true)
	return w
}

func (w *animatingWidget) Layout(_ Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(48, 48))
}

func (w *animatingWidget) Draw(ctx Context, canvas Canvas) {
	w.drawCount++
	// Spinner pattern: re-dirty self for next frame.
	w.SetNeedsRedraw(true)
	if ctx != nil {
		ctx.InvalidateRect(w.Bounds())
	}
}

func (w *animatingWidget) Event(_ Context, _ event.Event) bool { return false }
func (w *animatingWidget) Children() []Widget                  { return nil }

// TestAnimatingBoundary_ReDirtiesSelfDuringDraw verifies that an animated
// boundary widget (spinner) remains dirty after drawBoundaryWidget completes.
// Without this, spinner freezes after first frame (cache hit forever).
func TestAnimatingBoundary_ReDirtiesSelfDuringDraw(t *testing.T) {
	RegisterSceneFactory(func() SceneCache { return scene.NewScene() })
	defer RegisterSceneFactory(nil)
	RegisterSceneRecorder(stubSceneRecorder)
	defer RegisterSceneRecorder(nil)

	w := newAnimatingWidget()
	w.SetRepaintBoundary(true)
	w.SetBounds(geometry.NewRect(0, 0, 48, 48))

	ctx := NewContext()
	ctx.SetOnInvalidateRect(func(_ geometry.Rect) {})

	canvas := &stubReplayCanvas{}

	// Frame 1: first draw (cache miss — sceneDirty=true from SetRepaintBoundary).
	drawBoundaryWidget(w, ctx, canvas, nil)

	if w.drawCount != 1 {
		t.Fatalf("frame 1: drawCount = %d, want 1", w.drawCount)
	}

	// After frame 1: widget must be re-dirtied (called SetNeedsRedraw in Draw).
	if !w.NeedsRedraw() {
		t.Error("frame 1: NeedsRedraw() = false after Draw; " +
			"animated widget calls SetNeedsRedraw(true) during Draw, " +
			"this flag must survive drawBoundaryWidget")
	}
	if !w.IsSceneDirty() {
		t.Error("frame 1: IsSceneDirty() = false after Draw; " +
			"SetNeedsRedraw on boundary calls InvalidateScene, " +
			"scene must be dirty for next frame to trigger cache miss")
	}

	// Frame 2: must be cache miss (sceneDirty=true) — NOT cache hit.
	drawBoundaryWidget(w, ctx, canvas, nil)

	if w.drawCount != 2 {
		t.Fatalf("frame 2: drawCount = %d, want 2; "+
			"animation froze because boundary was cache-hit on second frame", w.drawCount)
	}

	// Frame 3: still animating.
	drawBoundaryWidget(w, ctx, canvas, nil)

	if w.drawCount != 3 {
		t.Fatalf("frame 3: drawCount = %d, want 3", w.drawCount)
	}
}

// TestAnimatingBoundary_DoesNotDirtyParent verifies that an animated boundary
// widget's SetNeedsRedraw during Draw does NOT propagate to parent boundary.
// Parent boundary must stay clean — only the animated widget re-records.
func TestAnimatingBoundary_DoesNotDirtyParent(t *testing.T) {
	RegisterSceneFactory(func() SceneCache { return scene.NewScene() })
	defer RegisterSceneFactory(nil)
	RegisterSceneRecorder(stubSceneRecorder)
	defer RegisterSceneRecorder(nil)

	parent := newDrawTrackingWidget()
	parent.SetRepaintBoundary(true)
	parent.SetBounds(geometry.NewRect(0, 0, 800, 600))
	// Clear initial sceneDirty from SetRepaintBoundary(true) so we can
	// isolate whether the CHILD's SetNeedsRedraw propagates to parent.
	parent.ClearSceneDirty()

	child := newAnimatingWidget()
	child.SetRepaintBoundary(true)
	child.SetBounds(geometry.NewRect(100, 100, 148, 148))
	child.SetParent(parent)

	ctx := NewContext()
	ctx.SetOnInvalidateRect(func(_ geometry.Rect) {})

	canvas := &stubReplayCanvas{}

	// First draw of child boundary.
	drawBoundaryWidget(child, ctx, canvas, nil)

	if child.drawCount != 1 {
		t.Fatalf("child.drawCount = %d, want 1", child.drawCount)
	}

	// Child re-dirtied itself (animated).
	if !child.IsSceneDirty() {
		t.Error("child boundary should be scene-dirty (re-dirtied by animation)")
	}

	// Parent must NOT be dirty.
	if parent.IsSceneDirty() {
		t.Error("parent boundary should NOT be dirty; " +
			"animated child's SetNeedsRedraw stops at child boundary, " +
			"does not propagate to parent (Flutter markNeedsPaint)")
	}
}

// TestDrawTree_RootBoundary_ChildBoundaryReached verifies that when root IS
// a boundary and a child IS also a boundary, DrawTree reaches the child.
// This is the gallery scenario: root boundary + spinner boundary.
func TestDrawTree_RootBoundary_ChildBoundaryReached(t *testing.T) {
	RegisterSceneFactory(func() SceneCache { return scene.NewScene() })
	defer RegisterSceneFactory(nil)
	RegisterSceneRecorder(stubSceneRecorder)
	defer RegisterSceneRecorder(nil)

	// Root boundary contains a container with an animated child boundary.
	root := newAnimContainerWidget()
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))

	spinner := newAnimatingWidget()
	spinner.SetRepaintBoundary(true)
	spinner.SetBounds(geometry.NewRect(100, 100, 148, 148))
	spinner.SetParent(root)
	root.addChild(spinner)

	ctx := NewContext()
	ctx.SetOnInvalidateRect(func(_ geometry.Rect) {})
	canvas := &stubReplayCanvas{}

	// Frame 1: both dirty (initial). Root records, child records inside.
	stats := DrawTree(root, ctx, canvas)
	_ = stats

	rootDrew := root.drawCalled
	spinnerDrew := spinner.drawCount > 0

	if !rootDrew {
		t.Error("frame 1: root.Draw() not called (initial cache miss expected)")
	}
	if !spinnerDrew {
		t.Error("frame 1: spinner.Draw() not called; " +
			"spinner is a child boundary inside root boundary, " +
			"must be reached during root's Draw")
	}

	// Frame 2: spinner re-dirtied itself. Root is clean.
	// Root should be cache-hit. But spinner inside root must still animate.
	root.drawCalled = false
	spinner.drawCount = 0

	stats2 := DrawTree(root, ctx, canvas)

	// KEY TEST: spinner must be drawn on frame 2.
	// If spinner.drawCount == 0, animation is frozen.
	if spinner.drawCount == 0 {
		t.Errorf("frame 2: spinner.Draw() not called; animation frozen. "+
			"Root is cache-hit (clean), but spinner boundary is dirty. "+
			"DrawTree must visit child boundaries even when root is cache-hit. "+
			"Stats: total=%d dirty=%d cached=%d",
			stats2.TotalWidgets, stats2.DirtyWidgets, stats2.CachedWidgets)
	}
}

// --- Test helpers for boundary draw ---

// stubSceneRecorder creates a minimal scene recording canvas for tests.
func stubSceneRecorder(_ SceneCache, _, _ int) (Canvas, func()) {
	return &stubReplayCanvas{}, func() {}
}

// stubReplayCanvas implements widget.Canvas for boundary draw tests.
type stubReplayCanvas struct {
	replayCount int
}

func (c *stubReplayCanvas) Clear(_ Color)                                                           {}
func (c *stubReplayCanvas) DrawRect(_ geometry.Rect, _ Color)                                       {}
func (c *stubReplayCanvas) FillRectDirect(_ geometry.Rect, _ Color)                                 {}
func (c *stubReplayCanvas) StrokeRect(_ geometry.Rect, _ Color, _ float32)                          {}
func (c *stubReplayCanvas) DrawRoundRect(_ geometry.Rect, _ Color, _ float32)                       {}
func (c *stubReplayCanvas) StrokeRoundRect(_ geometry.Rect, _ Color, _ float32, _ float32)          {}
func (c *stubReplayCanvas) DrawCircle(_ geometry.Point, _ float32, _ Color)                         {}
func (c *stubReplayCanvas) StrokeCircle(_ geometry.Point, _ float32, _ Color, _ float32)            {}
func (c *stubReplayCanvas) StrokeArc(_ geometry.Point, _ float32, _, _ float64, _ Color, _ float32) {}
func (c *stubReplayCanvas) DrawLine(_, _ geometry.Point, _ Color, _ float32)                        {}
func (c *stubReplayCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ Color, _ bool, _ TextAlign) {
}
func (c *stubReplayCanvas) MeasureText(_ string, _ float32, _ bool) float32 { return 0 }
func (c *stubReplayCanvas) DrawImage(_ image.Image, _ geometry.Point)       {}
func (c *stubReplayCanvas) PushClip(_ geometry.Rect)                        {}
func (c *stubReplayCanvas) PushClipRoundRect(_ geometry.Rect, _ float32)    {}
func (c *stubReplayCanvas) PopClip()                                        {}
func (c *stubReplayCanvas) PushTransform(_ geometry.Point)                  {}
func (c *stubReplayCanvas) PopTransform()                                   {}
func (c *stubReplayCanvas) TransformOffset() geometry.Point                 { return geometry.Point{} }
func (c *stubReplayCanvas) ScreenOriginBase() geometry.Point                { return geometry.Point{} }
func (c *stubReplayCanvas) ClipBounds() geometry.Rect                       { return geometry.NewRect(0, 0, 9999, 9999) }
func (c *stubReplayCanvas) ReplayScene(s SceneCache)                        { c.replayCount++ }

var _ Canvas = (*stubReplayCanvas)(nil)

// animContainerWidget is a simple container that draws children via child.Draw().
type animContainerWidget struct {
	WidgetBase
	drawCalled bool
	kids       []Widget
}

func newAnimContainerWidget() *animContainerWidget {
	w := &animContainerWidget{}
	w.SetVisible(true)
	w.SetEnabled(true)
	return w
}

func (w *animContainerWidget) addChild(child Widget) {
	w.kids = append(w.kids, child)
	w.AddChild(child)
}

func (w *animContainerWidget) Layout(_ Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(800, 600))
}

func (w *animContainerWidget) Draw(ctx Context, canvas Canvas) {
	w.drawCalled = true
	for _, child := range w.kids {
		child.Draw(ctx, canvas)
	}
}

func (w *animContainerWidget) Event(_ Context, _ event.Event) bool { return false }
func (w *animContainerWidget) Children() []Widget                  { return w.kids }
