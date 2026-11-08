package primitives_test

import (
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
)

// Runtime child mutation (P1-WIDGET-001, issue #243).

// recordLeaf is a leaf widget that records layout/draw/event participation and
// exposes hooks so a test can mutate the tree from inside a callback.
type recordLeaf struct {
	widget.WidgetBase
	name       string
	size       geometry.Size
	layoutCall int
	drawCall   int
	onLayout   func()
	onDraw     func()
}

func newRecordLeaf(name string, size geometry.Size) *recordLeaf {
	l := &recordLeaf{name: name, size: size}
	l.SetVisible(true)
	l.SetEnabled(true)
	return l
}

func (l *recordLeaf) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	l.layoutCall++
	if l.onLayout != nil {
		l.onLayout()
	}
	return c.Constrain(l.size)
}

func (l *recordLeaf) Draw(_ widget.Context, _ widget.Canvas) {
	l.drawCall++
	if l.onDraw != nil {
		l.onDraw()
	}
}

func (l *recordLeaf) Event(_ widget.Context, _ event.Event) bool { return false }

// mountLeaf counts Mount/Unmount calls and records the context it was given.
type mountLeaf struct {
	widget.WidgetBase
	mountCall   int
	unmountCall int
	mountedCtx  widget.Context
}

func (m *mountLeaf) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Smallest()
}

func (m *mountLeaf) Draw(_ widget.Context, _ widget.Canvas) {}

func (m *mountLeaf) Event(_ widget.Context, _ event.Event) bool { return false }

func (m *mountLeaf) Mount(ctx widget.Context) {
	m.mountCall++
	m.mountedCtx = ctx
}

func (m *mountLeaf) Unmount() { m.unmountCall++ }

func TestBox_AddChildAppearsInLayout(t *testing.T) {
	box := primitives.Box()
	child := newRecordLeaf("A", geometry.Sz(10, 10))

	box.AddChild(child)
	box.Layout(widget.NewContext(), geometry.Loose(geometry.Sz(100, 100)))

	if child.layoutCall != 1 {
		t.Errorf("child added at runtime was laid out %d times, want 1", child.layoutCall)
	}
	if child.Bounds().Width() != 10 {
		t.Errorf("child bounds not set by layout: %v", child.Bounds())
	}
}

func TestBox_InsertChildPreservesOrder(t *testing.T) {
	a := newRecordLeaf("A", geometry.Sz(10, 10))
	b := newRecordLeaf("B", geometry.Sz(10, 10))
	c := newRecordLeaf("C", geometry.Sz(10, 10))
	box := primitives.Box(a, b, c)

	d := newRecordLeaf("D", geometry.Sz(10, 10))
	box.InsertChild(1, d)

	var order []string
	for _, l := range []*recordLeaf{a, b, c, d} {
		l.onLayout = func() { order = append(order, l.name) }
	}
	box.Layout(widget.NewContext(), geometry.Loose(geometry.Sz(1000, 1000)))

	want := []string{"A", "D", "B", "C"}
	if len(order) != len(want) {
		t.Fatalf("layout order = %v, want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("layout order = %v, want %v", order, want)
		}
	}
}

func TestBox_ChildCountAndChildAt(t *testing.T) {
	a := newRecordLeaf("A", geometry.Sz(10, 10))
	b := newRecordLeaf("B", geometry.Sz(10, 10))
	c := newRecordLeaf("C", geometry.Sz(10, 10))
	box := primitives.Box(a, b, c)

	if box.ChildCount() != 3 {
		t.Errorf("ChildCount = %d, want 3", box.ChildCount())
	}
	if box.ChildAt(0) != a || box.ChildAt(1) != b || box.ChildAt(2) != c {
		t.Error("ChildAt does not match construction order")
	}
}

func TestBox_AddChildSetsParent(t *testing.T) {
	box := primitives.Box()
	child := newRecordLeaf("A", geometry.Sz(10, 10))

	box.AddChild(child)

	if child.Parent() != box {
		t.Error("AddChild must set the child's parent to the box")
	}
}

func TestBox_RemoveChildClearsParent(t *testing.T) {
	box := primitives.Box()
	child := newRecordLeaf("A", geometry.Sz(10, 10))
	box.AddChild(child)

	ctx := widget.NewContext()
	widget.MountTree(box, ctx)
	if !child.IsMounted() {
		t.Fatal("precondition: child should be mounted")
	}

	if !box.RemoveChild(child) {
		t.Fatal("RemoveChild should return true")
	}
	if child.Parent() != nil {
		t.Error("removed child should have no parent")
	}
	if child.IsMounted() {
		t.Error("removed child should be unmounted")
	}
}

func TestBox_AddChildMountsWhenParentMounted(t *testing.T) {
	box := primitives.Box()
	ctx := widget.NewContext()
	widget.MountTree(box, ctx)

	child := &mountLeaf{}
	box.AddChild(child)

	if !child.IsMounted() {
		t.Error("child added to a mounted box should be mounted")
	}
	if child.mountCall != 1 {
		t.Errorf("Mount calls = %d, want 1", child.mountCall)
	}
	if child.mountedCtx != ctx {
		t.Error("child should be mounted with the container's context")
	}

	box.RemoveChild(child)
	if child.unmountCall != 1 {
		t.Errorf("Unmount calls = %d, want 1", child.unmountCall)
	}
}

func TestBox_RemoveChildMarksLayoutAndRepaint(t *testing.T) {
	box := primitives.Box()
	ctx := widget.NewContext()
	layoutDirty := false
	box.SetOnLayoutDirty(func() { layoutDirty = true })

	child := newRecordLeaf("A", geometry.Sz(10, 10))
	box.AddChild(child)

	// Prime the layout cache so MarkNeedsLayout is not a no-op (it returns
	// early when the cache is already invalid).
	widget.LayoutChild(box, ctx, geometry.Loose(geometry.Sz(1000, 1000)))
	box.ClearRedraw()
	layoutDirty = false

	box.RemoveChild(child)

	if !layoutDirty {
		t.Error("removing a child must invalidate the parent's layout")
	}
	if !box.NeedsRedraw() {
		t.Error("removing a child must mark the parent for repaint")
	}
}

func TestBox_MutationDuringDraw(t *testing.T) {
	a := newRecordLeaf("A", geometry.Sz(10, 10))
	b := newRecordLeaf("B", geometry.Sz(10, 10))
	c := newRecordLeaf("C", geometry.Sz(10, 10))
	box := primitives.Box(a, b, c)

	// B removes itself from inside its own Draw. The draw pass must not skip C,
	// and the removal must actually take effect in the child list.
	b.onDraw = func() { box.RemoveChild(b) }

	ctx := widget.NewContext()
	box.Layout(ctx, geometry.Loose(geometry.Sz(1000, 1000)))
	box.Draw(ctx, &mockCanvas{})

	if a.drawCall != 1 || b.drawCall != 1 || c.drawCall != 1 {
		t.Errorf("sibling skipped during draw: A=%d B=%d C=%d", a.drawCall, b.drawCall, c.drawCall)
	}
	remaining := box.Children()
	if len(remaining) != 2 || remaining[0] != a || remaining[1] != c {
		t.Errorf("after mutation-in-draw Children() = %v, want [A C]", leafNames(remaining))
	}
}

func leafNames(ws []widget.Widget) []string {
	names := make([]string, len(ws))
	for i, w := range ws {
		if l, ok := w.(*recordLeaf); ok {
			names[i] = l.name
		}
	}
	return names
}

func TestBox_RemoveFocusedChild_ClearsFocus(t *testing.T) {
	box := primitives.Box()
	ctx := widget.NewContext()
	widget.MountTree(box, ctx)

	child := newRecordLeaf("A", geometry.Sz(10, 10))
	box.AddChild(child)
	ctx.RequestFocus(child)
	if ctx.FocusedWidget() != child {
		t.Fatal("precondition: child should be focused")
	}

	box.RemoveChild(child)

	if ctx.FocusedWidget() != nil {
		t.Error("focus must be released synchronously when the focused child is removed")
	}
}

func TestBox_ClearChildrenUnmountsAll(t *testing.T) {
	box := primitives.Box()
	ctx := widget.NewContext()
	widget.MountTree(box, ctx)

	c1 := &mountLeaf{}
	c2 := &mountLeaf{}
	box.AddChild(c1)
	box.AddChild(c2)

	box.ClearChildren()

	if box.ChildCount() != 0 {
		t.Errorf("ChildCount = %d after ClearChildren, want 0", box.ChildCount())
	}
	if c1.unmountCall != 1 || c2.unmountCall != 1 {
		t.Errorf("ClearChildren must unmount every child: c1=%d c2=%d", c1.unmountCall, c2.unmountCall)
	}
	if c1.IsMounted() || c2.IsMounted() {
		t.Error("cleared children must be unmounted")
	}
}

func TestBox_AddChildToUnmountedTree_MountsOnSetRoot(t *testing.T) {
	box := primitives.Box()
	c1 := &mountLeaf{}
	c2 := &mountLeaf{}

	// Added while the tree is not yet mounted.
	box.AddChild(c1)
	box.AddChild(c2)
	if c1.IsMounted() || c2.IsMounted() {
		t.Fatal("children should not be mounted before the tree is")
	}

	widget.MountTree(box, widget.NewContext())

	if c1.mountCall != 1 || c2.mountCall != 1 {
		t.Errorf("each child must be mounted exactly once: c1=%d c2=%d", c1.mountCall, c2.mountCall)
	}
	if !c1.IsMounted() || !c2.IsMounted() {
		t.Error("children should be mounted after MountTree")
	}
}

func TestBox_MutationInsideRepaintBoundary_DoesNotDirtyAncestors(t *testing.T) {
	outer := primitives.Box()
	outer.SetRepaintBoundary(true)
	inner := primitives.Box()
	inner.SetRepaintBoundary(true)

	leaf := newRecordLeaf("leaf", geometry.Sz(10, 10))
	inner.AddChild(leaf)
	outer.AddChild(inner)

	ctx := widget.NewContext()
	// Prime the inner layout cache via the parent path.
	widget.LayoutChild(inner, ctx, geometry.Loose(geometry.Sz(1000, 1000)))

	// Establish a clean baseline for both boundaries and the inner redraw flag.
	outer.ClearSceneDirty()
	inner.ClearRedraw()

	inner.RemoveChild(leaf)

	if !inner.NeedsRedraw() {
		t.Error("the mutated container must be marked for repaint")
	}
	if outer.IsSceneDirty() {
		t.Error("ADR-024 isolation: mutation inside a boundary must not dirty the ancestor boundary")
	}
}
