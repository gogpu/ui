package widget

import (
	"testing"

	"github.com/gogpu/gg/scene"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
)

func TestWidgetBase_SetRepaintBoundary(t *testing.T) {
	w := NewWidgetBase()

	if w.IsRepaintBoundary() {
		t.Error("expected IsRepaintBoundary=false by default")
	}

	w.SetRepaintBoundary(true)
	if !w.IsRepaintBoundary() {
		t.Error("expected IsRepaintBoundary=true after SetRepaintBoundary(true)")
	}
	if w.BoundaryCacheKey() == 0 {
		t.Error("expected non-zero cache key after enabling boundary")
	}
	if !w.IsSceneDirty() {
		t.Error("expected sceneDirty=true after enabling boundary (first draw needs record)")
	}

	// Disable boundary.
	w.SetRepaintBoundary(false)
	if w.IsRepaintBoundary() {
		t.Error("expected IsRepaintBoundary=false after SetRepaintBoundary(false)")
	}
	if w.CachedScene() != nil {
		t.Error("expected cached scene to be nil after disabling boundary")
	}
}

func TestWidgetBase_SetRepaintBoundary_Idempotent(t *testing.T) {
	w := NewWidgetBase()
	w.SetRepaintBoundary(true)
	key1 := w.BoundaryCacheKey()

	// Setting again should not change the key.
	w.SetRepaintBoundary(true)
	key2 := w.BoundaryCacheKey()

	if key1 != key2 {
		t.Errorf("cache key changed on repeated SetRepaintBoundary: %d != %d", key1, key2)
	}
}

func TestWidgetBase_InvalidateScene(t *testing.T) {
	w := NewWidgetBase()

	// No-op when not a boundary.
	w.InvalidateScene()
	if w.IsSceneDirty() {
		t.Error("InvalidateScene should be no-op when not a boundary")
	}

	// Enable boundary, clear dirty, then invalidate.
	w.SetRepaintBoundary(true)
	w.ClearSceneDirty() // Simulate successful record.
	if w.IsSceneDirty() {
		t.Error("expected sceneDirty=false after ClearSceneDirty")
	}

	w.InvalidateScene()
	if !w.IsSceneDirty() {
		t.Error("expected sceneDirty=true after InvalidateScene")
	}
}

func TestWidgetBase_InvalidateScene_Callback(t *testing.T) {
	w := NewWidgetBase()
	w.SetRepaintBoundary(true)
	w.ClearSceneDirty()

	callbackCalled := false
	w.SetOnBoundaryDirty(func() {
		callbackCalled = true
	})

	w.InvalidateScene()
	if !callbackCalled {
		t.Error("onBoundaryDirty callback not called on InvalidateScene")
	}

	// Second call should be no-op (already dirty).
	callbackCalled = false
	w.InvalidateScene()
	if callbackCalled {
		t.Error("onBoundaryDirty callback should not be called when already dirty")
	}
}

func TestWidgetBase_SceneCacheSize(t *testing.T) {
	w := NewWidgetBase()
	w.SetRepaintBoundary(true)

	w.SetSceneCacheSize(200, 100)
	width, height := w.SceneCacheSize()
	if width != 200 || height != 100 {
		t.Errorf("SceneCacheSize = (%d, %d), want (200, 100)", width, height)
	}
}

func TestWidgetBase_CachedScene(t *testing.T) {
	w := NewWidgetBase()
	w.SetRepaintBoundary(true)

	if w.CachedScene() != nil {
		t.Error("expected nil cached scene initially")
	}

	sc := scene.NewScene()
	w.SetCachedScene(sc)
	if w.CachedScene() != sc {
		t.Error("expected same scene reference after SetCachedScene")
	}

	// Disable boundary releases scene.
	w.SetRepaintBoundary(false)
	if w.CachedScene() != nil {
		t.Error("expected nil cached scene after disabling boundary")
	}
}

func TestWidgetBase_SceneCacheVersion(t *testing.T) {
	w := NewWidgetBase()
	w.SetRepaintBoundary(true)

	v0 := w.SceneCacheVersion()
	w.ClearSceneDirty()
	v1 := w.SceneCacheVersion()

	if v1 <= v0 {
		t.Errorf("SceneCacheVersion should increment on ClearSceneDirty: %d <= %d", v1, v0)
	}
}

func TestPropagateDirtyUpward_BoundaryProperty(t *testing.T) {
	// Create a tree: parent (boundary) → child.
	parent := &boundaryTestWidget{}
	parent.SetVisible(true)
	parent.SetEnabled(true)
	parent.SetRepaintBoundary(true)
	parent.ClearSceneDirty() // Simulate previous successful draw.

	child := &boundaryTestWidget{}
	child.SetVisible(true)
	child.SetEnabled(true)
	child.SetParent(parent)

	// Marking child dirty should propagate to parent boundary.
	child.SetNeedsRedraw(true)

	if !parent.IsSceneDirty() {
		t.Error("parent boundary should be dirty after child SetNeedsRedraw")
	}
}

func TestPropagateDirtyUpward_LegacyBoundary(t *testing.T) {
	// Create a mock that implements RepaintBoundaryMarker (legacy pattern).
	boundary := &legacyBoundaryTestWidget{}
	boundary.SetVisible(true)
	boundary.SetEnabled(true)

	child := &boundaryTestWidget{}
	child.SetVisible(true)
	child.SetEnabled(true)
	child.SetParent(boundary)

	// Marking child dirty should call legacy MarkBoundaryDirty.
	child.SetNeedsRedraw(true)

	if !boundary.markedDirty {
		t.Error("legacy boundary should be marked dirty after child SetNeedsRedraw")
	}
}

func TestDrawBoundaryWidget_FallbackOnZeroBounds(t *testing.T) {
	// Widget with boundary enabled but zero bounds should fall back to normal draw.
	RegisterSceneFactory(func() SceneCache { return scene.NewScene() })
	defer RegisterSceneFactory(nil)
	RegisterSceneRecorder(func(s SceneCache, w, h int) (Canvas, func()) {
		return &noopCanvas{}, func() {}
	})
	defer RegisterSceneRecorder(nil)

	root := &drawCountWidget{}
	root.SetVisible(true)
	root.SetEnabled(true)
	root.SetRepaintBoundary(true)
	// Don't set bounds — they're zero.

	ctx := NewContext()
	canvas := &noopCanvas{}

	drawBoundaryWidget(root, ctx, canvas, nil)

	if root.drawCount != 1 {
		t.Errorf("drawCount = %d, want 1 (fallback on zero bounds)", root.drawCount)
	}
}

func TestDrawBoundaryWidget_CacheHit(t *testing.T) {
	RegisterSceneFactory(func() SceneCache { return scene.NewScene() })
	defer RegisterSceneFactory(nil)
	RegisterSceneRecorder(func(s SceneCache, w, h int) (Canvas, func()) {
		return &noopCanvas{}, func() {}
	})
	defer RegisterSceneRecorder(nil)

	root := &drawCountWidget{}
	root.SetVisible(true)
	root.SetEnabled(true)
	root.SetBounds(geometry.NewRect(0, 0, 100, 50))
	root.SetRepaintBoundary(true)

	ctx := NewContext()
	canvas := &noopCanvas{}

	// First draw: cache miss (sceneDirty=true).
	var stats DrawStats
	drawBoundaryWidget(root, ctx, canvas, &stats)
	if root.drawCount != 1 {
		t.Errorf("first draw: drawCount = %d, want 1", root.drawCount)
	}
	if stats.CachedWidgets != 0 {
		t.Errorf("first draw: CachedWidgets = %d, want 0", stats.CachedWidgets)
	}

	// Second draw: cache hit (sceneDirty=false, scene exists).
	stats = DrawStats{}
	drawBoundaryWidget(root, ctx, canvas, &stats)
	if root.drawCount != 1 {
		t.Errorf("second draw: drawCount = %d, want 1 (cache hit)", root.drawCount)
	}
	if stats.CachedWidgets != 1 {
		t.Errorf("second draw: CachedWidgets = %d, want 1", stats.CachedWidgets)
	}
}

func TestDrawBoundaryWidget_CacheInvalidation(t *testing.T) {
	RegisterSceneFactory(func() SceneCache { return scene.NewScene() })
	defer RegisterSceneFactory(nil)
	RegisterSceneRecorder(func(s SceneCache, w, h int) (Canvas, func()) {
		return &noopCanvas{}, func() {}
	})
	defer RegisterSceneRecorder(nil)

	root := &drawCountWidget{}
	root.SetVisible(true)
	root.SetEnabled(true)
	root.SetBounds(geometry.NewRect(0, 0, 100, 50))
	root.SetRepaintBoundary(true)

	ctx := NewContext()
	canvas := &noopCanvas{}

	// First draw: cache miss.
	drawBoundaryWidget(root, ctx, canvas, nil)
	if root.drawCount != 1 {
		t.Fatalf("first draw: drawCount = %d, want 1", root.drawCount)
	}

	// Invalidate scene.
	root.InvalidateScene()

	// Next draw: cache miss again (invalidated).
	drawBoundaryWidget(root, ctx, canvas, nil)
	if root.drawCount != 2 {
		t.Errorf("after invalidation: drawCount = %d, want 2", root.drawCount)
	}
}

// --- test helpers ---

// boundaryTestWidget is a minimal widget for boundary testing.
type boundaryTestWidget struct {
	WidgetBase
}

func (w *boundaryTestWidget) Layout(_ Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(100, 50))
}

func (w *boundaryTestWidget) Draw(_ Context, _ Canvas) {}

func (w *boundaryTestWidget) Event(_ Context, _ event.Event) bool {
	return false
}

// drawCountWidget counts Draw calls for cache hit/miss verification.
type drawCountWidget struct {
	WidgetBase
	drawCount int
}

func (w *drawCountWidget) Layout(_ Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(100, 50))
}

func (w *drawCountWidget) Draw(_ Context, _ Canvas) {
	w.drawCount++
}

func (w *drawCountWidget) Event(_ Context, _ event.Event) bool {
	return false
}

// legacyBoundaryTestWidget implements RepaintBoundaryMarker (legacy pattern).
type legacyBoundaryTestWidget struct {
	WidgetBase
	markedDirty bool
}

func (w *legacyBoundaryTestWidget) Layout(_ Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(100, 50))
}

func (w *legacyBoundaryTestWidget) Draw(_ Context, _ Canvas) {}

func (w *legacyBoundaryTestWidget) Event(_ Context, _ event.Event) bool {
	return false
}

func (w *legacyBoundaryTestWidget) MarkBoundaryDirty() {
	w.markedDirty = true
}
