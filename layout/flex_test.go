package layout

import (
	"testing"

	"github.com/gogpu/ui/geometry"
)

func TestFlexLayout_Name(t *testing.T) {
	flex := &FlexLayout{}
	if name := flex.Name(); name != "flex" {
		t.Errorf("Name() = %q, want %q", name, "flex")
	}
}

func TestFlexLayout_EmptyChildren(t *testing.T) {
	flex := &FlexLayout{}
	tree := newTestTree()

	result := flex.Compute(tree, 1, geometry.Size{Width: 300, Height: 200})

	if !result.Size.IsZero() {
		t.Errorf("Result.Size = %v, want zero", result.Size)
	}
}

func TestFlexLayout_Row(t *testing.T) {
	flex := &FlexLayout{}
	tree := newTestTree()

	// Setup: 3 children in a row
	tree.AddChild(1, 10)
	tree.AddChild(1, 11)
	tree.AddChild(1, 12)

	tree.SetPreferredSize(10, geometry.Size{Width: 100, Height: 50})
	tree.SetPreferredSize(11, geometry.Size{Width: 100, Height: 60})
	tree.SetPreferredSize(12, geometry.Size{Width: 100, Height: 40})

	tree.SetStyle(1, &Style{
		FlexDirection:  FlexRow,
		JustifyContent: JustifyStart,
		AlignItems:     AlignItemsStart,
	})

	result := flex.Compute(tree, 1, geometry.Size{Width: 500, Height: 200})

	// Total width should be 300 (3 x 100)
	if result.Size.Width != 300 {
		t.Errorf("Result.Size.Width = %v, want 300", result.Size.Width)
	}

	// Height should be max child height (60)
	if result.Size.Height != 60 {
		t.Errorf("Result.Size.Height = %v, want 60", result.Size.Height)
	}

	// Check child positions
	layout10 := tree.GetLayout(10)
	if layout10.Position.X != 0 {
		t.Errorf("child 10 X = %v, want 0", layout10.Position.X)
	}

	layout11 := tree.GetLayout(11)
	if layout11.Position.X != 100 {
		t.Errorf("child 11 X = %v, want 100", layout11.Position.X)
	}

	layout12 := tree.GetLayout(12)
	if layout12.Position.X != 200 {
		t.Errorf("child 12 X = %v, want 200", layout12.Position.X)
	}
}

func TestFlexLayout_Column(t *testing.T) {
	flex := &FlexLayout{}
	tree := newTestTree()

	// Setup: 3 children in a column
	tree.AddChild(1, 10)
	tree.AddChild(1, 11)
	tree.AddChild(1, 12)

	tree.SetPreferredSize(10, geometry.Size{Width: 100, Height: 50})
	tree.SetPreferredSize(11, geometry.Size{Width: 150, Height: 60})
	tree.SetPreferredSize(12, geometry.Size{Width: 80, Height: 40})

	tree.SetStyle(1, &Style{
		FlexDirection:  FlexColumn,
		JustifyContent: JustifyStart,
		AlignItems:     AlignItemsStart,
	})

	result := flex.Compute(tree, 1, geometry.Size{Width: 200, Height: 500})

	// Width should be max child width (150)
	if result.Size.Width != 150 {
		t.Errorf("Result.Size.Width = %v, want 150", result.Size.Width)
	}

	// Total height should be 150 (50 + 60 + 40)
	if result.Size.Height != 150 {
		t.Errorf("Result.Size.Height = %v, want 150", result.Size.Height)
	}

	// Check child positions
	layout10 := tree.GetLayout(10)
	if layout10.Position.Y != 0 {
		t.Errorf("child 10 Y = %v, want 0", layout10.Position.Y)
	}

	layout11 := tree.GetLayout(11)
	if layout11.Position.Y != 50 {
		t.Errorf("child 11 Y = %v, want 50", layout11.Position.Y)
	}

	layout12 := tree.GetLayout(12)
	if layout12.Position.Y != 110 {
		t.Errorf("child 12 Y = %v, want 110", layout12.Position.Y)
	}
}

func TestFlexLayout_FlexGrow(t *testing.T) {
	flex := &FlexLayout{}
	tree := newTestTree()

	// Setup: 2 children, one grows
	tree.AddChild(1, 10)
	tree.AddChild(1, 11)

	tree.SetPreferredSize(10, geometry.Size{Width: 100, Height: 50})
	tree.SetPreferredSize(11, geometry.Size{Width: 100, Height: 50})

	tree.SetStyle(1, &Style{FlexDirection: FlexRow})
	tree.SetStyle(10, &Style{FlexGrow: 0})
	tree.SetStyle(11, &Style{FlexGrow: 1})

	result := flex.Compute(tree, 1, geometry.Size{Width: 400, Height: 100})

	// Total should be sum of final sizes (100 + 300 = 400)
	// Child 11 grows to fill remaining space (400 - 100 = 300)
	if result.Size.Width != 400 {
		t.Errorf("Result.Size.Width = %v, want 400", result.Size.Width)
	}

	// Child 10 stays at 100, child 11 grows to 300
	layout10 := tree.GetLayout(10)
	if layout10.Size.Width != 100 {
		t.Errorf("child 10 Width = %v, want 100", layout10.Size.Width)
	}

	layout11 := tree.GetLayout(11)
	if layout11.Size.Width != 300 {
		t.Errorf("child 11 Width = %v, want 300", layout11.Size.Width)
	}
}

func TestFlexLayout_JustifyCenter(t *testing.T) {
	flex := &FlexLayout{}
	tree := newTestTree()

	tree.AddChild(1, 10)
	tree.SetPreferredSize(10, geometry.Size{Width: 100, Height: 50})

	tree.SetStyle(1, &Style{
		FlexDirection:  FlexRow,
		JustifyContent: JustifyCenter,
	})

	_ = flex.Compute(tree, 1, geometry.Size{Width: 300, Height: 100})

	layout := tree.GetLayout(10)
	// Centered: (300 - 100) / 2 = 100
	if layout.Position.X != 100 {
		t.Errorf("child X = %v, want 100 (centered)", layout.Position.X)
	}
}

func TestFlexLayout_JustifySpaceBetween(t *testing.T) {
	flex := &FlexLayout{}
	tree := newTestTree()

	tree.AddChild(1, 10)
	tree.AddChild(1, 11)
	tree.AddChild(1, 12)

	tree.SetPreferredSize(10, geometry.Size{Width: 50, Height: 50})
	tree.SetPreferredSize(11, geometry.Size{Width: 50, Height: 50})
	tree.SetPreferredSize(12, geometry.Size{Width: 50, Height: 50})

	tree.SetStyle(1, &Style{
		FlexDirection:  FlexRow,
		JustifyContent: JustifySpaceBetween,
	})

	_ = flex.Compute(tree, 1, geometry.Size{Width: 300, Height: 100})

	// Free space = 300 - 150 = 150
	// Space between = 150 / 2 = 75

	layout10 := tree.GetLayout(10)
	if layout10.Position.X != 0 {
		t.Errorf("child 10 X = %v, want 0", layout10.Position.X)
	}

	layout11 := tree.GetLayout(11)
	if layout11.Position.X != 125 { // 50 + 75
		t.Errorf("child 11 X = %v, want 125", layout11.Position.X)
	}

	layout12 := tree.GetLayout(12)
	if layout12.Position.X != 250 { // 50 + 75 + 50 + 75
		t.Errorf("child 12 X = %v, want 250", layout12.Position.X)
	}
}

func TestFlexLayout_AlignItemsStretch(t *testing.T) {
	flex := &FlexLayout{}
	tree := newTestTree()

	tree.AddChild(1, 10)
	tree.AddChild(1, 11)

	tree.SetPreferredSize(10, geometry.Size{Width: 100, Height: 30})
	tree.SetPreferredSize(11, geometry.Size{Width: 100, Height: 60})

	tree.SetStyle(1, &Style{
		FlexDirection: FlexRow,
		AlignItems:    AlignItemsStretch,
	})

	_ = flex.Compute(tree, 1, geometry.Size{Width: 300, Height: 100})

	// Both should stretch to max cross size (60)
	layout10 := tree.GetLayout(10)
	if layout10.Size.Height != 60 {
		t.Errorf("child 10 Height = %v, want 60 (stretched)", layout10.Size.Height)
	}
}

func TestFlexLayout_WithGap(t *testing.T) {
	flex := &FlexLayout{}
	tree := newTestTree()

	tree.AddChild(1, 10)
	tree.AddChild(1, 11)
	tree.AddChild(1, 12)

	tree.SetPreferredSize(10, geometry.Size{Width: 50, Height: 50})
	tree.SetPreferredSize(11, geometry.Size{Width: 50, Height: 50})
	tree.SetPreferredSize(12, geometry.Size{Width: 50, Height: 50})

	tree.SetStyle(1, &Style{
		FlexDirection:  FlexRow,
		JustifyContent: JustifyStart,
		Gap:            10,
	})

	result := flex.Compute(tree, 1, geometry.Size{Width: 300, Height: 100})

	// Total width = 50 + 10 + 50 + 10 + 50 = 170
	if result.Size.Width != 170 {
		t.Errorf("Result.Size.Width = %v, want 170", result.Size.Width)
	}

	layout11 := tree.GetLayout(11)
	if layout11.Position.X != 60 { // 50 + 10
		t.Errorf("child 11 X = %v, want 60", layout11.Position.X)
	}

	layout12 := tree.GetLayout(12)
	if layout12.Position.X != 120 { // 50 + 10 + 50 + 10
		t.Errorf("child 12 X = %v, want 120", layout12.Position.X)
	}
}

func TestFlexLayout_RowReverse(t *testing.T) {
	flex := &FlexLayout{}
	tree := newTestTree()

	tree.AddChild(1, 10)
	tree.AddChild(1, 11)

	tree.SetPreferredSize(10, geometry.Size{Width: 100, Height: 50})
	tree.SetPreferredSize(11, geometry.Size{Width: 100, Height: 50})

	tree.SetStyle(1, &Style{
		FlexDirection:  FlexRowReverse,
		JustifyContent: JustifyStart,
	})

	_ = flex.Compute(tree, 1, geometry.Size{Width: 300, Height: 100})

	// In reverse, first child should be at the right
	layout10 := tree.GetLayout(10)
	if layout10.Position.X != 200 { // 300 - 100
		t.Errorf("child 10 X = %v, want 200 (reversed)", layout10.Position.X)
	}

	layout11 := tree.GetLayout(11)
	if layout11.Position.X != 100 { // 200 - 100
		t.Errorf("child 11 X = %v, want 100 (reversed)", layout11.Position.X)
	}
}

func TestFlexLayout_Registered(t *testing.T) {
	// Flex should be registered via init()
	if !Has("flex") {
		t.Error("flex layout should be registered")
	}

	algo, ok := Get("flex")
	if !ok {
		t.Fatal("Get('flex') should return true")
	}

	if algo.Name() != "flex" {
		t.Errorf("algorithm Name() = %q, want %q", algo.Name(), "flex")
	}
}
