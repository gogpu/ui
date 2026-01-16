package layout

import (
	"testing"

	"github.com/gogpu/ui/geometry"
)

func TestLayoutTreeAdapter_Style(t *testing.T) {
	adapter := NewLayoutTreeAdapter()

	// No style set - should return default
	style := adapter.Style(1)
	if style == nil {
		t.Fatal("Style should never return nil")
	}

	// Set a style
	customStyle := &Style{FlexGrow: 2}
	adapter.SetStyle(1, customStyle)

	style = adapter.Style(1)
	if style.FlexGrow != 2 {
		t.Errorf("Style.FlexGrow = %v, want 2", style.FlexGrow)
	}

	// Other nodes still get default
	style = adapter.Style(2)
	if style.FlexGrow != 0 {
		t.Errorf("unset node Style.FlexGrow = %v, want 0", style.FlexGrow)
	}
}

func TestLayoutTreeAdapter_Layout(t *testing.T) {
	adapter := NewLayoutTreeAdapter()

	// No layout set - should return zero
	layout := adapter.GetLayout(1)
	if !layout.IsZero() {
		t.Errorf("GetLayout for unset node = %v, want zero", layout)
	}

	// Set a layout
	nodeLayout := NodeLayout{
		Position: geometry.Point{X: 10, Y: 20},
		Size:     geometry.Size{Width: 100, Height: 50},
	}
	adapter.SetLayout(1, nodeLayout)

	layout = adapter.GetLayout(1)
	if layout.Position.X != 10 || layout.Position.Y != 20 {
		t.Errorf("GetLayout Position = %v, want {10, 20}", layout.Position)
	}
	if layout.Size.Width != 100 || layout.Size.Height != 50 {
		t.Errorf("GetLayout Size = %v, want {100, 50}", layout.Size)
	}
}

func TestLayoutTreeAdapter_Clear(t *testing.T) {
	adapter := NewLayoutTreeAdapter()

	adapter.SetStyle(1, &Style{FlexGrow: 1})
	adapter.SetLayout(1, NodeLayout{Size: geometry.Size{Width: 100}})

	adapter.Clear()

	// Should be empty after clear
	style := adapter.Style(1)
	if style.FlexGrow != 0 {
		t.Error("Style should be default after Clear")
	}

	layout := adapter.GetLayout(1)
	if !layout.IsZero() {
		t.Error("Layout should be zero after Clear")
	}
}

func TestLayoutTreeAdapter_NilMaps(t *testing.T) {
	// Test that methods work even with nil maps
	adapter := &LayoutTreeAdapter{}

	// SetLayout should initialize maps
	adapter.SetLayout(1, NodeLayout{Size: geometry.Size{Width: 100}})
	if adapter.Layouts == nil {
		t.Error("SetLayout should initialize Layouts map")
	}

	// SetStyle should initialize maps
	adapter.SetStyle(2, &Style{FlexGrow: 1})
	if adapter.Styles == nil {
		t.Error("SetStyle should initialize Styles map")
	}

	// GetLayout with nil map should return zero
	adapter2 := &LayoutTreeAdapter{}
	layout := adapter2.GetLayout(1)
	if !layout.IsZero() {
		t.Error("GetLayout with nil map should return zero")
	}
}

// testTree implements LayoutTree for testing.
type testTree struct {
	*LayoutTreeAdapter
	children map[NodeID][]NodeID
	sizes    map[NodeID]geometry.Size
}

func newTestTree() *testTree {
	return &testTree{
		LayoutTreeAdapter: NewLayoutTreeAdapter(),
		children:          make(map[NodeID][]NodeID),
		sizes:             make(map[NodeID]geometry.Size),
	}
}

func (t *testTree) ChildCount(parent NodeID) int {
	return len(t.children[parent])
}

func (t *testTree) ChildAt(parent NodeID, index int) NodeID {
	children := t.children[parent]
	if index < 0 || index >= len(children) {
		return InvalidNodeID
	}
	return children[index]
}

func (t *testTree) Measure(node NodeID, constraints geometry.Constraints) geometry.Size {
	size := t.sizes[node]
	return constraints.Constrain(size)
}

func (t *testTree) AddChild(parent, child NodeID) {
	t.children[parent] = append(t.children[parent], child)
}

func (t *testTree) SetPreferredSize(node NodeID, size geometry.Size) {
	t.sizes[node] = size
}

func TestTestTree_Implementation(t *testing.T) {
	tree := newTestTree()

	// Setup tree structure
	tree.AddChild(1, 10)
	tree.AddChild(1, 11)
	tree.AddChild(1, 12)

	tree.SetPreferredSize(10, geometry.Size{Width: 100, Height: 50})
	tree.SetPreferredSize(11, geometry.Size{Width: 150, Height: 60})
	tree.SetPreferredSize(12, geometry.Size{Width: 80, Height: 40})

	// Test ChildCount
	if count := tree.ChildCount(1); count != 3 {
		t.Errorf("ChildCount(1) = %d, want 3", count)
	}

	// Test ChildAt
	if child := tree.ChildAt(1, 0); child != 10 {
		t.Errorf("ChildAt(1, 0) = %d, want 10", child)
	}
	if child := tree.ChildAt(1, 1); child != 11 {
		t.Errorf("ChildAt(1, 1) = %d, want 11", child)
	}
	if child := tree.ChildAt(1, 99); child != InvalidNodeID {
		t.Errorf("ChildAt(1, 99) = %d, want InvalidNodeID", child)
	}

	// Test Measure
	size := tree.Measure(10, geometry.Loose(geometry.Size{Width: 200, Height: 200}))
	if size.Width != 100 || size.Height != 50 {
		t.Errorf("Measure(10) = %v, want {100, 50}", size)
	}

	// Test Measure with tight constraints
	size = tree.Measure(10, geometry.Tight(geometry.Size{Width: 80, Height: 40}))
	if size.Width != 80 || size.Height != 40 {
		t.Errorf("Measure(10) with tight = %v, want {80, 40}", size)
	}
}
