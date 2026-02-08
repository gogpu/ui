package focus

import (
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// mockWidget is a test widget that implements Widget and optionally Focusable.
type mockWidget struct {
	widget.WidgetBase
	focusable bool
	children  []widget.Widget
}

func newMockWidget(id string, focusable bool) *mockWidget {
	w := &mockWidget{focusable: focusable}
	w.SetID(id)
	w.SetVisible(true)
	w.SetEnabled(true)
	return w
}

func (w *mockWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Smallest()
}

func (w *mockWidget) Draw(_ widget.Context, _ widget.Canvas) {}

func (w *mockWidget) Event(_ widget.Context, _ event.Event) bool {
	return false
}

func (w *mockWidget) Children() []widget.Widget {
	return w.children
}

func (w *mockWidget) IsFocusable() bool {
	return w.focusable && w.IsVisible() && w.IsEnabled()
}

func (w *mockWidget) addChild(child *mockWidget) {
	w.children = append(w.children, child)
}

// --- Traversal Tests ---

func TestCollectFocusable_Order(t *testing.T) {
	root := newMockWidget("root", false)
	a := newMockWidget("a", true)
	b := newMockWidget("b", false)
	c := newMockWidget("c", true)
	d := newMockWidget("d", true)

	b.addChild(c)
	root.addChild(a)
	root.addChild(b)
	root.addChild(d)

	result := collectFocusable(root)

	if len(result) != 3 {
		t.Fatalf("expected 3 focusable, got %d", len(result))
	}
	if result[0] != a {
		t.Error("first focusable should be a")
	}
	if result[1] != c {
		t.Error("second focusable should be c")
	}
	if result[2] != d {
		t.Error("third focusable should be d")
	}
}

func TestCollectFocusable_NilRoot(t *testing.T) {
	result := collectFocusable(nil)
	if result != nil {
		t.Error("nil root should return nil")
	}
}

func TestCollectFocusable_NoFocusable(t *testing.T) {
	root := newMockWidget("root", false)
	child := newMockWidget("child", false)
	root.addChild(child)

	result := collectFocusable(root)
	if len(result) != 0 {
		t.Errorf("expected 0 focusable, got %d", len(result))
	}
}

func TestIndexOf(t *testing.T) {
	a := newMockWidget("a", true)
	b := newMockWidget("b", true)
	c := newMockWidget("c", true)
	list := []widget.Focusable{a, b, c}

	if idx := indexOf(list, a); idx != 0 {
		t.Errorf("indexOf(a) = %d, want 0", idx)
	}
	if idx := indexOf(list, b); idx != 1 {
		t.Errorf("indexOf(b) = %d, want 1", idx)
	}
	if idx := indexOf(list, c); idx != 2 {
		t.Errorf("indexOf(c) = %d, want 2", idx)
	}

	notInList := newMockWidget("x", true)
	if idx := indexOf(list, notInList); idx != -1 {
		t.Errorf("indexOf(not in list) = %d, want -1", idx)
	}

	if idx := indexOf(nil, a); idx != -1 {
		t.Errorf("indexOf(nil list) = %d, want -1", idx)
	}
}

func TestDrawFocusRing_Constants(t *testing.T) {
	if DefaultFocusRingOffset != 2.0 {
		t.Errorf("DefaultFocusRingOffset = %v, want 2.0", DefaultFocusRingOffset)
	}
	if DefaultFocusRingStrokeWidth != 2.0 {
		t.Errorf("DefaultFocusRingStrokeWidth = %v, want 2.0", DefaultFocusRingStrokeWidth)
	}
}
