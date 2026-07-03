package widget_test

import (
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"
)

// clwWidget is a minimal WidgetBase-embedding widget that counts Layout calls
// and returns a configurable size, for exercising the layout cache.
type clwWidget struct {
	widget.WidgetBase
	size        geometry.Size
	layoutCalls int
}

func (w *clwWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	if !widget.IsLayoutVerifying() {
		w.layoutCalls++
	}
	return c.Constrain(w.size)
}
func (w *clwWidget) Draw(_ widget.Context, _ widget.Canvas)     {}
func (w *clwWidget) Event(_ widget.Context, _ event.Event) bool { return false }
func (w *clwWidget) Children() []widget.Widget                  { return nil }

// noBaseWidget implements Widget without embedding WidgetBase, so it cannot
// participate in layout caching.
type noBaseWidget struct{ layoutCalls int }

func (w *noBaseWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	w.layoutCalls++
	return c.Constrain(geometry.Sz(10, 10))
}
func (w *noBaseWidget) Draw(_ widget.Context, _ widget.Canvas)     {}
func (w *noBaseWidget) Event(_ widget.Context, _ event.Event) bool { return false }
func (w *noBaseWidget) Children() []widget.Widget                  { return nil }

func looseConstraints() geometry.Constraints {
	return geometry.Constraints{MinWidth: 0, MaxWidth: 500, MinHeight: 0, MaxHeight: 500}
}

func newWB() *clwWidget {
	w := &clwWidget{size: geometry.Sz(100, 50)}
	w.SetVisible(true)
	w.SetEnabled(true)
	return w
}

// --- Cache hit / miss ---

func TestLayoutChild_CachesOnSameConstraints(t *testing.T) {
	w := newWB()
	ctx := uitest.NewMockContext()
	c := looseConstraints()

	s1 := widget.LayoutChild(w, ctx, c)
	s2 := widget.LayoutChild(w, ctx, c)

	if w.layoutCalls != 1 {
		t.Errorf("layoutCalls = %d, want 1 (second call should hit cache)", w.layoutCalls)
	}
	if s1 != (geometry.Sz(100, 50)) || s2 != s1 {
		t.Errorf("sizes = %v, %v, want both (100,50)", s1, s2)
	}
	if !w.IsLayoutCacheValid() {
		t.Error("cache should be valid after a layout")
	}
}

func TestLayoutChild_RecomputesOnDifferentConstraints(t *testing.T) {
	w := newWB()
	ctx := uitest.NewMockContext()

	widget.LayoutChild(w, ctx, looseConstraints())
	widget.LayoutChild(w, ctx, geometry.Constraints{MaxWidth: 300, MaxHeight: 300})

	if w.layoutCalls != 2 {
		t.Errorf("layoutCalls = %d, want 2 (different constraints miss)", w.layoutCalls)
	}
}

func TestLayoutChild_NilReturnsZero(t *testing.T) {
	if got := widget.LayoutChild(nil, uitest.NewMockContext(), looseConstraints()); got != (geometry.Size{}) {
		t.Errorf("LayoutChild(nil) = %v, want zero", got)
	}
}

func TestLayoutChild_NonCacheableAlwaysLayouts(t *testing.T) {
	w := &noBaseWidget{}
	ctx := uitest.NewMockContext()
	c := looseConstraints()

	widget.LayoutChild(w, ctx, c)
	widget.LayoutChild(w, ctx, c)

	if w.layoutCalls != 2 {
		t.Errorf("layoutCalls = %d, want 2 (no caching without WidgetBase)", w.layoutCalls)
	}
}

// --- Invalidation ---

func TestMarkNeedsLayout_InvalidatesCache(t *testing.T) {
	w := newWB()
	ctx := uitest.NewMockContext()
	c := looseConstraints()

	widget.LayoutChild(w, ctx, c)
	w.MarkNeedsLayout()

	if w.IsLayoutCacheValid() {
		t.Error("cache should be invalid after MarkNeedsLayout")
	}
	widget.LayoutChild(w, ctx, c)
	if w.layoutCalls != 2 {
		t.Errorf("layoutCalls = %d, want 2 (recompute after invalidation)", w.layoutCalls)
	}
}

func TestInvalidateLayoutCache_DoesNotPropagateUpward(t *testing.T) {
	parent, child := newWB(), newWB()
	child.SetParent(parent)
	ctx := uitest.NewMockContext()
	c := looseConstraints()

	widget.LayoutChild(parent, ctx, c)
	widget.LayoutChild(child, ctx, c)

	child.InvalidateLayoutCache() // self-only

	if child.IsLayoutCacheValid() {
		t.Error("child cache should be invalid")
	}
	if !parent.IsLayoutCacheValid() {
		t.Error("InvalidateLayoutCache must not touch ancestors")
	}
}

func TestMarkNeedsLayout_PropagatesCacheInvalidationUpward(t *testing.T) {
	parent, child := newWB(), newWB()
	child.SetParent(parent)
	ctx := uitest.NewMockContext()
	c := looseConstraints()

	widget.LayoutChild(parent, ctx, c)
	widget.LayoutChild(child, ctx, c)

	child.MarkNeedsLayout()

	if parent.IsLayoutCacheValid() {
		t.Error("ancestor cache must be invalidated (its size depends on the child)")
	}
}

func TestMarkNeedsLayout_StopsAtNonWidgetBaseAncestor(t *testing.T) {
	// A parent that implements Widget but not WidgetBase: it neither caches
	// nor exposes Parent(), so upward propagation must skip it and stop.
	child := newWB()
	child.SetParent(&noBaseWidget{})
	// Must not panic and must still invalidate the child itself.
	widget.LayoutChild(child, uitest.NewMockContext(), looseConstraints())
	child.MarkNeedsLayout()
	if child.IsLayoutCacheValid() {
		t.Error("child cache should be invalid after MarkNeedsLayout")
	}
}

// --- GAP-1: paint dirty timing ---

func TestMarkNeedsLayout_MarksPaintSelfOnly(t *testing.T) {
	parent, child := newWB(), newWB()
	child.SetParent(parent)
	ctx := uitest.NewMockContext()
	c := looseConstraints()
	// A valid cache is the "clean" marker the guard checks, so lay both out first.
	widget.LayoutChild(parent, ctx, c)
	widget.LayoutChild(child, ctx, c)
	parent.SetNeedsRedraw(false)
	child.SetNeedsRedraw(false)

	child.MarkNeedsLayout()

	if !child.NeedsRedraw() {
		t.Error("MarkNeedsLayout should mark the widget itself paint-dirty")
	}
	if parent.NeedsRedraw() {
		t.Error("GAP-1: the upward walk must clear the parent's cache but NOT mark it paint-dirty")
	}
	if parent.IsLayoutCacheValid() {
		t.Error("the parent's layout cache should be invalidated (its size depends on the child)")
	}
}

func TestLayoutChild_CacheMissMarksPaint(t *testing.T) {
	w := newWB()
	w.SetNeedsRedraw(false)
	widget.LayoutChild(w, uitest.NewMockContext(), looseConstraints())
	if !w.NeedsRedraw() {
		t.Error("cache miss should mark the child paint-dirty (changed subtree repaints)")
	}
}

func TestLayoutChild_CacheHitDoesNotMarkPaint(t *testing.T) {
	w := newWB()
	ctx := uitest.NewMockContext()
	c := looseConstraints()
	widget.LayoutChild(w, ctx, c) // miss → marks paint
	w.SetNeedsRedraw(false)
	widget.LayoutChild(w, ctx, c) // hit
	if w.NeedsRedraw() {
		t.Error("cache hit must not mark paint-dirty (unchanged subtree)")
	}
}

// --- Root callback wiring ---

func TestOnLayoutDirty_FiresWhenInvalidationReachesRoot(t *testing.T) {
	root, child := newWB(), newWB()
	child.SetParent(root)
	ctx := uitest.NewMockContext()
	c := looseConstraints()
	fired := 0
	root.SetOnLayoutDirty(func() { fired++ })

	widget.LayoutChild(root, ctx, c)
	widget.LayoutChild(child, ctx, c)

	child.MarkNeedsLayout() // propagates to root → fires once
	if fired != 1 {
		t.Errorf("onLayoutDirty fired %d times, want 1 (reached root via propagation)", fired)
	}

	widget.LayoutChild(root, ctx, c) // re-validate root (it was invalidated above)
	root.MarkNeedsLayout()           // root invalidating itself → fires once
	if fired != 2 {
		t.Errorf("onLayoutDirty fired %d times, want 2 (root invalidating itself)", fired)
	}
}

func TestMarkNeedsLayout_SelfGuardIsIdempotent(t *testing.T) {
	root := newWB()
	ctx := uitest.NewMockContext()
	c := looseConstraints()
	fired := 0
	root.SetOnLayoutDirty(func() { fired++ })

	widget.LayoutChild(root, ctx, c)
	root.MarkNeedsLayout() // valid → fires, invalidates
	root.MarkNeedsLayout() // already invalid → no-op
	if fired != 1 {
		t.Errorf("onLayoutDirty fired %d times, want 1 (second call is a no-op)", fired)
	}
}

func TestMarkNeedsLayout_SiblingBatchFiresRootOnce(t *testing.T) {
	// root → box → {c1, c2}. Invalidating both children in a batch must fire the
	// root callback exactly once; the second child short-circuits at box.
	root, box, c1, c2 := newWB(), newWB(), newWB(), newWB()
	box.SetParent(root)
	c1.SetParent(box)
	c2.SetParent(box)
	ctx := uitest.NewMockContext()
	cc := looseConstraints()
	fired := 0
	root.SetOnLayoutDirty(func() { fired++ })

	for _, w := range []*clwWidget{root, box, c1, c2} {
		widget.LayoutChild(w, ctx, cc)
	}

	c1.MarkNeedsLayout() // full walk c1→box→root, fires once
	c2.MarkNeedsLayout() // box already invalid → stops, no extra fire

	if fired != 1 {
		t.Errorf("onLayoutDirty fired %d times, want 1 (sibling batch coalesces)", fired)
	}
	for _, w := range []*clwWidget{box, c1, c2} {
		if w.IsLayoutCacheValid() {
			t.Error("all invalidated nodes should have cleared caches")
		}
	}
}

// --- Debug verifier (ADR-032 Phase 1c) ---

func TestDebugVerifier_PanicsOnStaleCache(t *testing.T) {
	defer widget.SetLayoutDebug(widget.SetLayoutDebug(true))
	w := newWB()
	ctx := uitest.NewMockContext()
	c := looseConstraints()

	widget.LayoutChild(w, ctx, c) // cache (100,50)
	w.size = geometry.Sz(200, 50) // layout-affecting change WITHOUT MarkNeedsLayout

	defer func() {
		if recover() == nil {
			t.Error("expected panic: stale cache hit should be caught by the verifier")
		}
	}()
	widget.LayoutChild(w, ctx, c) // hit → verifier recomputes → mismatch → panic
}

func TestDebugVerifier_NoPanicWhenStable(t *testing.T) {
	defer widget.SetLayoutDebug(widget.SetLayoutDebug(true))
	w := newWB()
	ctx := uitest.NewMockContext()
	c := looseConstraints()

	widget.LayoutChild(w, ctx, c)
	widget.LayoutChild(w, ctx, c) // hit, verified, matches — must not panic
}

func TestDebugVerifier_SentinelSuppressesCount(t *testing.T) {
	defer widget.SetLayoutDebug(widget.SetLayoutDebug(true))
	w := newWB()
	ctx := uitest.NewMockContext()
	c := looseConstraints()

	widget.LayoutChild(w, ctx, c) // miss: layoutCalls=1
	widget.LayoutChild(w, ctx, c) // hit: verifier re-runs but sentinel suppresses count

	if w.layoutCalls != 1 {
		t.Errorf("layoutCalls = %d, want 1 (verifier pass should be invisible)", w.layoutCalls)
	}
}

func TestSetLayoutDebug_RestoresPrevious(t *testing.T) {
	prev := widget.SetLayoutDebug(true)
	if widget.SetLayoutDebug(prev) != true {
		t.Error("SetLayoutDebug should return the value it just set")
	}
	// Leave the flag as it started.
	widget.SetLayoutDebug(prev)
}
