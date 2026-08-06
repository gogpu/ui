package primitives_test

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
)

// --- ADR-007 Phase 2: Scene Display List Cache Tests ---

// replayRecordingCanvas is a mock canvas that records ReplayScene calls.
type replayRecordingCanvas struct {
	mockCanvas
	replayCount  int
	replayScenes []widget.SceneCache
}

func (c *replayRecordingCanvas) ReplayScene(s widget.SceneCache) {
	c.replayCount++
	c.replayScenes = append(c.replayScenes, s)
}

func TestRepaintBoundary_SceneCache_CacheMissRecordsScene(t *testing.T) {
	// First draw is always a cache miss — child.Draw is executed and the
	// result is recorded into a scene.Scene, then replayed via ReplayScene.
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)
	rb.Layout(nil, geometry.BoxConstraints(0, 200, 0, 100))

	canvas := &replayRecordingCanvas{}
	rb.Draw(nil, canvas)

	if child.drawCount != 1 {
		t.Errorf("child drawn %d times, want 1", child.drawCount)
	}
	if canvas.replayCount != 1 {
		t.Errorf("ReplayScene called %d times, want 1", canvas.replayCount)
	}
	if canvas.replayScenes[0] == nil {
		t.Error("ReplayScene received nil scene")
	}
}

func TestRepaintBoundary_SceneCache_CacheHitReplaysWithoutRedraw(t *testing.T) {
	// On second draw with clean subtree, the cached scene should be
	// replayed without re-executing child.Draw.
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)
	rb.Layout(nil, geometry.BoxConstraints(0, 200, 0, 100))

	canvas1 := &replayRecordingCanvas{}
	rb.Draw(nil, canvas1) // Cache miss.

	canvas2 := &replayRecordingCanvas{}
	rb.Draw(nil, canvas2) // Cache hit.

	if child.drawCount != 1 {
		t.Errorf("child drawn %d times, want 1 (second draw should use cache)", child.drawCount)
	}
	if rb.CacheHits() != 1 {
		t.Errorf("CacheHits = %d, want 1", rb.CacheHits())
	}
	if canvas2.replayCount != 1 {
		t.Errorf("second draw: ReplayScene called %d times, want 1", canvas2.replayCount)
	}
}

func TestRepaintBoundary_SceneCache_DirtySubtreeReRecords(t *testing.T) {
	// When a descendant is dirty, the boundary should re-record the scene.
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)
	rb.Layout(nil, geometry.BoxConstraints(0, 200, 0, 100))

	canvas := &replayRecordingCanvas{}
	rb.Draw(nil, canvas) // First draw: cache miss.

	// Mark boundary dirty (simulating upward dirty propagation).
	rb.MarkBoundaryDirty()

	canvas2 := &replayRecordingCanvas{}
	rb.Draw(nil, canvas2) // Second draw: dirty → re-record.

	if child.drawCount != 2 {
		t.Errorf("child drawn %d times, want 2 (dirty child re-recorded)", child.drawCount)
	}
	if canvas2.replayCount != 1 {
		t.Errorf("ReplayScene called %d times, want 1", canvas2.replayCount)
	}
}

func TestRepaintBoundary_SceneCache_InvalidateCausesMiss(t *testing.T) {
	// InvalidateCache should force a re-record on the next draw.
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)
	rb.Layout(nil, geometry.BoxConstraints(0, 200, 0, 100))

	canvas := &replayRecordingCanvas{}
	rb.Draw(nil, canvas) // First draw.

	rb.InvalidateCache()

	canvas2 := &replayRecordingCanvas{}
	rb.Draw(nil, canvas2) // Should re-record.

	if child.drawCount != 2 {
		t.Errorf("child drawn %d times, want 2 after InvalidateCache", child.drawCount)
	}
}

func TestRepaintBoundary_SceneCache_SizeChangeInvalidates(t *testing.T) {
	// When the boundary size changes, the cached scene should be re-recorded.
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)
	// drawCountingWidget returns Constrain(100, 50), so use Tight to force size.
	rb.Layout(nil, geometry.Tight(geometry.Sz(100, 50)))

	canvas := &replayRecordingCanvas{}
	rb.Draw(nil, canvas) // First draw at 100x50.

	// Force different size via Tight constraints.
	rb.Layout(nil, geometry.Tight(geometry.Sz(200, 100)))

	canvas2 := &replayRecordingCanvas{}
	rb.Draw(nil, canvas2) // Should re-record at new size.

	if child.drawCount != 2 {
		t.Errorf("child drawn %d times, want 2 after size change", child.drawCount)
	}
}

func TestRepaintBoundary_SceneCache_UnmountClearsCache(t *testing.T) {
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)
	rb.Layout(nil, geometry.BoxConstraints(0, 200, 0, 100))

	canvas := &replayRecordingCanvas{}
	rb.Draw(nil, canvas) // Populate cache.

	if !rb.CacheValid() {
		t.Fatal("pre-condition: cache should be valid")
	}

	rb.Unmount()

	if rb.CacheValid() {
		t.Error("cache should be invalid after Unmount")
	}
}

func TestRepaintBoundary_SceneCache_NilChildNoOp(t *testing.T) {
	// RepaintBoundary with nil child should not panic.
	rb := primitives.NewRepaintBoundary(nil)
	rb.Layout(nil, geometry.BoxConstraints(0, 200, 0, 100))

	canvas := &replayRecordingCanvas{}
	rb.Draw(nil, canvas)

	if canvas.replayCount != 0 {
		t.Errorf("ReplayScene called %d times for nil child, want 0", canvas.replayCount)
	}
}

func TestRepaintBoundary_SceneCache_ZeroSizeNoOp(t *testing.T) {
	// Zero-size boundary should not attempt rendering.
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)
	rb.Layout(nil, geometry.BoxConstraints(0, 0, 0, 0))

	canvas := &replayRecordingCanvas{}
	rb.Draw(nil, canvas)

	if child.drawCount != 0 {
		t.Errorf("child drawn %d times for zero size, want 0", child.drawCount)
	}
	if canvas.replayCount != 0 {
		t.Errorf("ReplayScene called %d times for zero size, want 0", canvas.replayCount)
	}
}

func TestRepaintBoundary_SceneCache_SameSceneReused(t *testing.T) {
	// The same scene.Scene object should be replayed on cache hit
	// (not re-created each frame).
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)
	rb.Layout(nil, geometry.BoxConstraints(0, 200, 0, 100))

	canvas1 := &replayRecordingCanvas{}
	rb.Draw(nil, canvas1) // Cache miss.

	canvas2 := &replayRecordingCanvas{}
	rb.Draw(nil, canvas2) // Cache hit.

	if canvas1.replayScenes[0] != canvas2.replayScenes[0] {
		t.Error("cache hit should replay the same scene.Scene object")
	}
}

func TestRepaintBoundary_CacheKey_UniquePerBoundary(t *testing.T) {
	// Every RepaintBoundary should get a unique, monotonically increasing key.
	keys := make(map[uint64]bool)
	for range 100 {
		rb := primitives.NewRepaintBoundary(nil)
		if keys[rb.CacheKey()] {
			t.Fatalf("duplicate cache key: %d", rb.CacheKey())
		}
		keys[rb.CacheKey()] = true
	}
}

func TestRepaintBoundary_SceneCache_MultipleBoundariesIndependent(t *testing.T) {
	// Multiple RepaintBoundary instances should maintain independent caches.
	child1 := newDrawCountingWidget()
	child2 := newDrawCountingWidget()

	rb1 := primitives.NewRepaintBoundary(child1)
	rb2 := primitives.NewRepaintBoundary(child2)

	for _, rb := range []*primitives.RepaintBoundary{rb1, rb2} {
		rb.Layout(nil, geometry.BoxConstraints(0, 200, 0, 100))
	}

	// Draw both.
	canvas1 := &replayRecordingCanvas{}
	rb1.Draw(nil, canvas1)
	canvas2 := &replayRecordingCanvas{}
	rb2.Draw(nil, canvas2)

	// Dirty only rb1.
	rb1.MarkBoundaryDirty()

	canvas3 := &replayRecordingCanvas{}
	rb1.Draw(nil, canvas3) // Re-record.
	canvas4 := &replayRecordingCanvas{}
	rb2.Draw(nil, canvas4) // Cache hit.

	if child1.drawCount != 2 {
		t.Errorf("child1 drawn %d times, want 2", child1.drawCount)
	}
	if child2.drawCount != 1 {
		t.Errorf("child2 drawn %d times, want 1 (should be cached)", child2.drawCount)
	}
}

func TestRepaintBoundary_SceneCache_BoundaryDirtyCallback(t *testing.T) {
	// SetOnBoundaryDirty callback should fire when boundary becomes dirty.
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)

	var callbackCount int
	rb.SetOnBoundaryDirty(func(_ *primitives.RepaintBoundary) {
		callbackCount++
	})

	rb.MarkBoundaryDirty()
	rb.MarkBoundaryDirty() // Second call — O(1) guard, should not fire again.

	if callbackCount != 1 {
		t.Errorf("callback fired %d times, want 1 (O(1) guard)", callbackCount)
	}
}

func TestRepaintBoundary_SceneCache_DrawStatsRecorded(t *testing.T) {
	// Cache hits should be recorded in DrawStats when context provides them.
	child := newDrawCountingWidget()
	rb := primitives.NewRepaintBoundary(child)
	rb.Layout(nil, geometry.BoxConstraints(0, 200, 0, 100))

	ctx := widget.NewContext()
	stats := &widget.DrawStats{}
	ctx.SetDrawStats(stats)

	canvas := &replayRecordingCanvas{}
	rb.Draw(ctx, canvas) // Cache miss — no stat.
	rb.Draw(ctx, canvas) // Cache hit — stat incremented.

	if stats.CachedWidgets != 1 {
		t.Errorf("DrawStats.CachedWidgets = %d, want 1", stats.CachedWidgets)
	}
}
