package app

import (
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

type fixedSizeResizeChild struct {
	widget.WidgetBase
	layoutCalls int
}

func newFixedSizeResizeChild() *fixedSizeResizeChild {
	child := &fixedSizeResizeChild{}
	child.SetVisible(true)
	child.SetEnabled(true)
	child.SetRepaintBoundary(true)
	return child
}

func (c *fixedSizeResizeChild) Layout(_ widget.Context, _ geometry.Constraints) geometry.Size {
	c.layoutCalls++
	return geometry.Sz(50, 50)
}

func (c *fixedSizeResizeChild) Draw(_ widget.Context, _ widget.Canvas) {}

func (c *fixedSizeResizeChild) Event(_ widget.Context, _ event.Event) bool { return false }

type widthDependentResizeChild struct {
	widget.WidgetBase
	layoutCalls int
}

func newWidthDependentResizeChild() *widthDependentResizeChild {
	child := &widthDependentResizeChild{}
	child.SetVisible(true)
	child.SetEnabled(true)
	child.SetRepaintBoundary(true)
	return child
}

func (c *widthDependentResizeChild) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	c.layoutCalls++
	return constraints.Constrain(geometry.Sz(50, 50))
}

func (c *widthDependentResizeChild) Draw(_ widget.Context, _ widget.Canvas) {}

func (c *widthDependentResizeChild) Event(_ widget.Context, _ event.Event) bool { return false }

type selectiveResizeRoot struct {
	widget.WidgetBase
	fixedChild *fixedSizeResizeChild
	widthChild *widthDependentResizeChild
}

func newSelectiveResizeRoot() *selectiveResizeRoot {
	root := &selectiveResizeRoot{
		fixedChild: newFixedSizeResizeChild(),
		widthChild: newWidthDependentResizeChild(),
	}
	root.SetVisible(true)
	root.SetEnabled(true)
	root.AddChild(root.fixedChild)
	root.AddChild(root.widthChild)
	return root
}

func (r *selectiveResizeRoot) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	size := constraints.Constrain(geometry.Sz(800, 600))
	widget.LayoutChild(r.fixedChild, ctx, geometry.Tight(geometry.Sz(50, 50)))
	widget.LayoutChild(r.widthChild, ctx, geometry.Tight(geometry.Sz(size.Width/2, 50)))
	return size
}

func (r *selectiveResizeRoot) Draw(_ widget.Context, _ widget.Canvas) {}

func (r *selectiveResizeRoot) Event(_ widget.Context, _ event.Event) bool { return false }

func TestHandleResize_InvalidatesOnlyConstraintChangedBoundaries(t *testing.T) {
	provider := &mockWindowProvider{width: 800, height: 600, scale: 1}
	w := New(WithWindowProvider(provider)).Window()
	t.Cleanup(w.Close)
	root := newSelectiveResizeRoot()
	w.SetRoot(root)
	w.Frame()

	// Model a completed paint: both boundaries and all widget redraw flags
	// start clean before the resize event arrives.
	widget.ClearRedrawInTree(root)
	root.ClearSceneDirty()
	root.fixedChild.ClearSceneDirty()
	root.widthChild.ClearSceneDirty()
	root.fixedChild.layoutCalls = 0
	root.widthChild.layoutCalls = 0

	provider.width, provider.height = 1024, 768
	w.HandleResize(provider.width, provider.height)
	w.Frame()

	if root.fixedChild.layoutCalls != 0 {
		t.Errorf("fixed-size child Layout called %d times, want 0 for unchanged constraints", root.fixedChild.layoutCalls)
	}
	if root.fixedChild.IsSceneDirty() {
		t.Error("fixed-size child boundary became scene-dirty after unrelated window resize")
	}
	if root.widthChild.layoutCalls != 1 {
		t.Errorf("width-dependent child Layout called %d times, want 1 for changed constraints", root.widthChild.layoutCalls)
	}
	if !root.widthChild.IsSceneDirty() {
		t.Error("width-dependent child boundary remained clean after its constraints changed")
	}
	if !root.IsSceneDirty() {
		t.Error("root boundary remained clean after its window constraints changed")
	}
}
