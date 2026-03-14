package dirty

import (
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// testWidget is a mock widget that embeds WidgetBase for testing the collector.
type testWidget struct {
	widget.WidgetBase
}

func (w *testWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(100, 50))
}

func (w *testWidget) Draw(_ widget.Context, _ widget.Canvas) {}

func (w *testWidget) Event(_ widget.Context, _ event.Event) bool { return false }

func newTestWidget(x, y, w, h float32) *testWidget {
	tw := &testWidget{}
	tw.SetVisible(true)
	tw.SetEnabled(true)
	tw.SetBounds(geometry.NewRect(x, y, w, h))
	return tw
}

// customTestWidget does not embed WidgetBase — no NeedsRedraw method.
type customTestWidget struct {
	bounds   geometry.Rect
	children []widget.Widget
}

func (w *customTestWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(w.bounds.Size())
}

func (w *customTestWidget) Draw(_ widget.Context, _ widget.Canvas) {}

func (w *customTestWidget) Event(_ widget.Context, _ event.Event) bool { return false }

func (w *customTestWidget) Children() []widget.Widget { return w.children }

func (w *customTestWidget) Bounds() geometry.Rect { return w.bounds }

// invisibleWidget is visible=false.
type invisibleWidget struct {
	widget.WidgetBase
}

func (w *invisibleWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(50, 50))
}

func (w *invisibleWidget) Draw(_ widget.Context, _ widget.Canvas) {}

func (w *invisibleWidget) Event(_ widget.Context, _ event.Event) bool { return false }

func TestCollector_NilRoot(t *testing.T) {
	tr := NewTracker()
	c := NewCollector(tr)
	c.Collect(nil) // should not panic
	if !tr.IsEmpty() {
		t.Error("collecting nil root should produce no dirty regions")
	}
}

func TestCollector_SingleDirtyWidget(t *testing.T) {
	tr := NewTracker()
	c := NewCollector(tr)

	w := newTestWidget(10, 20, 100, 50)
	w.SetNeedsRedraw(true)

	c.Collect(w)

	if tr.IsEmpty() {
		t.Error("dirty widget should produce a dirty region")
	}
	if tr.RegionCount() != 1 {
		t.Errorf("region count = %d, want 1", tr.RegionCount())
	}
	expected := geometry.NewRect(10, 20, 100, 50)
	if tr.DirtyRegions()[0].Bounds != expected {
		t.Errorf("region = %v, want %v", tr.DirtyRegions()[0].Bounds, expected)
	}
}

func TestCollector_SingleCleanWidget(t *testing.T) {
	tr := NewTracker()
	c := NewCollector(tr)

	w := newTestWidget(10, 20, 100, 50)
	w.ClearRedraw()

	c.Collect(w)

	if !tr.IsEmpty() {
		t.Error("clean widget should produce no dirty regions")
	}
}

func TestCollector_DirtyChild(t *testing.T) {
	tr := NewTracker()
	c := NewCollector(tr)

	parent := newTestWidget(0, 0, 300, 300)
	parent.ClearRedraw()

	child := newTestWidget(50, 50, 100, 50)
	child.SetNeedsRedraw(true)
	parent.AddChild(child)

	c.Collect(parent)

	if tr.RegionCount() != 1 {
		t.Errorf("region count = %d, want 1", tr.RegionCount())
	}
	expected := geometry.NewRect(50, 50, 100, 50)
	if tr.DirtyRegions()[0].Bounds != expected {
		t.Errorf("region = %v, want %v", tr.DirtyRegions()[0].Bounds, expected)
	}
}

func TestCollector_MultipleDirtyWidgets(t *testing.T) {
	tr := NewTracker()
	c := NewCollector(tr)

	parent := newTestWidget(0, 0, 500, 500)
	parent.ClearRedraw()

	child1 := newTestWidget(10, 10, 50, 50)
	child1.SetNeedsRedraw(true)

	child2 := newTestWidget(200, 200, 50, 50)
	child2.ClearRedraw()

	child3 := newTestWidget(400, 400, 50, 50)
	child3.SetNeedsRedraw(true)

	parent.AddChild(child1)
	parent.AddChild(child2)
	parent.AddChild(child3)

	c.Collect(parent)

	if tr.RegionCount() != 2 {
		t.Errorf("region count = %d, want 2 (two dirty children)", tr.RegionCount())
	}
}

func TestCollector_InvisibleWidgetSkipped(t *testing.T) {
	tr := NewTracker()
	c := NewCollector(tr)

	parent := newTestWidget(0, 0, 300, 300)
	parent.ClearRedraw()

	inv := &invisibleWidget{}
	inv.SetVisible(false)
	inv.SetBounds(geometry.NewRect(50, 50, 100, 100))
	inv.SetNeedsRedraw(true)
	parent.AddChild(inv)

	c.Collect(parent)

	if !tr.IsEmpty() {
		t.Error("invisible dirty widget should be skipped")
	}
}

func TestCollector_InvisibleSubtreeSkipped(t *testing.T) {
	tr := NewTracker()
	c := NewCollector(tr)

	root := newTestWidget(0, 0, 500, 500)
	root.ClearRedraw()

	invisParent := &invisibleWidget{}
	invisParent.SetVisible(false)
	invisParent.SetBounds(geometry.NewRect(0, 0, 200, 200))
	invisParent.ClearRedraw()

	dirtyChild := newTestWidget(10, 10, 50, 50)
	dirtyChild.SetNeedsRedraw(true)
	invisParent.AddChild(dirtyChild)

	root.AddChild(invisParent)

	c.Collect(root)

	if !tr.IsEmpty() {
		t.Error("children of invisible widget should be skipped")
	}
}

func TestCollector_CustomWidgetTreatedAsDirty(t *testing.T) {
	tr := NewTracker()
	c := NewCollector(tr)

	cw := &customTestWidget{
		bounds: geometry.NewRect(10, 10, 80, 80),
	}
	c.Collect(cw)

	if tr.IsEmpty() {
		t.Error("custom widget without NeedsRedraw should be treated as dirty")
	}
	expected := geometry.NewRect(10, 10, 80, 80)
	if tr.DirtyRegions()[0].Bounds != expected {
		t.Errorf("region = %v, want %v", tr.DirtyRegions()[0].Bounds, expected)
	}
}

func TestCollector_CustomWidgetWithChildren(t *testing.T) {
	tr := NewTracker()
	c := NewCollector(tr)

	child := newTestWidget(50, 50, 30, 30)
	child.SetNeedsRedraw(true)

	cw := &customTestWidget{
		bounds:   geometry.NewRect(0, 0, 200, 200),
		children: []widget.Widget{child},
	}
	c.Collect(cw)

	// Custom widget itself + dirty child = 2 regions.
	if tr.RegionCount() != 2 {
		t.Errorf("region count = %d, want 2", tr.RegionCount())
	}
}

func TestCollector_DeeplyNested(t *testing.T) {
	tr := NewTracker()
	c := NewCollector(tr)

	root := newTestWidget(0, 0, 500, 500)
	root.ClearRedraw()

	mid := newTestWidget(100, 100, 300, 300)
	mid.ClearRedraw()
	root.AddChild(mid)

	leaf := newTestWidget(150, 150, 50, 50)
	leaf.SetNeedsRedraw(true)
	mid.AddChild(leaf)

	c.Collect(root)

	if tr.RegionCount() != 1 {
		t.Errorf("region count = %d, want 1", tr.RegionCount())
	}
	expected := geometry.NewRect(150, 150, 50, 50)
	if tr.DirtyRegions()[0].Bounds != expected {
		t.Errorf("region = %v, want %v", tr.DirtyRegions()[0].Bounds, expected)
	}
}

func TestCollector_AllClean(t *testing.T) {
	tr := NewTracker()
	c := NewCollector(tr)

	root := newTestWidget(0, 0, 500, 500)
	root.ClearRedraw()

	child1 := newTestWidget(10, 10, 50, 50)
	child1.ClearRedraw()
	child2 := newTestWidget(100, 100, 50, 50)
	child2.ClearRedraw()
	root.AddChild(child1)
	root.AddChild(child2)

	c.Collect(root)

	if !tr.IsEmpty() {
		t.Error("all-clean tree should produce no dirty regions")
	}
}

func TestCollector_ParentDirtyChildrenClean(t *testing.T) {
	tr := NewTracker()
	c := NewCollector(tr)

	parent := newTestWidget(0, 0, 300, 300)
	parent.SetNeedsRedraw(true)

	child := newTestWidget(50, 50, 50, 50)
	child.ClearRedraw()
	parent.AddChild(child)

	c.Collect(parent)

	// Only parent is dirty, child is clean.
	if tr.RegionCount() != 1 {
		t.Errorf("region count = %d, want 1", tr.RegionCount())
	}
	expected := geometry.NewRect(0, 0, 300, 300)
	if tr.DirtyRegions()[0].Bounds != expected {
		t.Errorf("region = %v, want %v", tr.DirtyRegions()[0].Bounds, expected)
	}
}

func TestCollector_BothParentAndChildDirty(t *testing.T) {
	tr := NewTracker()
	c := NewCollector(tr)

	parent := newTestWidget(0, 0, 300, 300)
	parent.SetNeedsRedraw(true)

	child := newTestWidget(50, 50, 50, 50)
	child.SetNeedsRedraw(true)
	parent.AddChild(child)

	c.Collect(parent)

	// Both parent and child are dirty — 2 regions (optimization merges later).
	if tr.RegionCount() != 2 {
		t.Errorf("region count = %d, want 2", tr.RegionCount())
	}
}

// noBoundsWidget has no Bounds() method and no NeedsRedraw — tests the markWidgetDirty fallback.
type noBoundsWidget struct {
	children []widget.Widget
}

func (w *noBoundsWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(50, 50))
}

func (w *noBoundsWidget) Draw(_ widget.Context, _ widget.Canvas) {}

func (w *noBoundsWidget) Event(_ widget.Context, _ event.Event) bool { return false }

func (w *noBoundsWidget) Children() []widget.Widget { return w.children }

func TestCollector_NoBoundsWidget(t *testing.T) {
	tr := NewTracker()
	c := NewCollector(tr)

	// Widget without Bounds method — isWidgetDirty returns true (no NeedsRedraw),
	// but markWidgetDirty can't get bounds, so no region added.
	w := &noBoundsWidget{}
	c.Collect(w)

	if !tr.IsEmpty() {
		t.Error("widget without Bounds should not add a region")
	}
}

func TestNewCollector(t *testing.T) {
	tr := NewTracker()
	c := NewCollector(tr)
	if c == nil {
		t.Fatal("NewCollector returned nil")
	}
}

// Integration test: collect + optimize + intersect.
func TestCollectOptimizeIntersect(t *testing.T) {
	tr := NewTracker()
	c := NewCollector(tr)

	root := newTestWidget(0, 0, 800, 600)
	root.ClearRedraw()

	// Two nearby dirty widgets should merge.
	w1 := newTestWidget(10, 10, 40, 40)
	w1.SetNeedsRedraw(true)
	w2 := newTestWidget(55, 10, 40, 40)
	w2.SetNeedsRedraw(true)
	// One far-away dirty widget.
	w3 := newTestWidget(500, 500, 40, 40)
	w3.SetNeedsRedraw(true)

	root.AddChild(w1)
	root.AddChild(w2)
	root.AddChild(w3)

	c.Collect(root)
	tr.Optimize()

	// w1 and w2 should merge (within default 16px gap), w3 stays separate.
	if tr.RegionCount() != 2 {
		t.Errorf("after optimize: region count = %d, want 2", tr.RegionCount())
	}

	// Widget in merged region should intersect.
	if !tr.Intersects(geometry.NewRect(30, 20, 10, 10)) {
		t.Error("should intersect merged region")
	}
	// Widget far away should not intersect.
	if tr.Intersects(geometry.NewRect(300, 300, 10, 10)) {
		t.Error("should not intersect between regions")
	}
}
