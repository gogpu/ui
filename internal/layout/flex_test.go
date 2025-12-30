package layout

import (
	"testing"

	"github.com/gogpu/ui/geometry"
)

func TestDirection_String(t *testing.T) {
	tests := []struct {
		dir  Direction
		want string
	}{
		{Row, "Row"},
		{RowReverse, "RowReverse"},
		{Column, "Column"},
		{ColumnReverse, "ColumnReverse"},
		{Direction(99), "Unknown"},
	}

	for _, tt := range tests {
		got := tt.dir.String()
		if got != tt.want {
			t.Errorf("Direction(%d).String() = %q, want %q", tt.dir, got, tt.want)
		}
	}
}

func TestDirection_IsHorizontal(t *testing.T) {
	tests := []struct {
		dir  Direction
		want bool
	}{
		{Row, true},
		{RowReverse, true},
		{Column, false},
		{ColumnReverse, false},
	}

	for _, tt := range tests {
		got := tt.dir.IsHorizontal()
		if got != tt.want {
			t.Errorf("%s.IsHorizontal() = %v, want %v", tt.dir, got, tt.want)
		}
	}
}

func TestDirection_IsReversed(t *testing.T) {
	tests := []struct {
		dir  Direction
		want bool
	}{
		{Row, false},
		{RowReverse, true},
		{Column, false},
		{ColumnReverse, true},
	}

	for _, tt := range tests {
		got := tt.dir.IsReversed()
		if got != tt.want {
			t.Errorf("%s.IsReversed() = %v, want %v", tt.dir, got, tt.want)
		}
	}
}

func TestJustifyContent_String(t *testing.T) {
	tests := []struct {
		j    JustifyContent
		want string
	}{
		{JustifyStart, "Start"},
		{JustifyEnd, "End"},
		{JustifyCenter, "Center"},
		{JustifySpaceBetween, "SpaceBetween"},
		{JustifySpaceAround, "SpaceAround"},
		{JustifySpaceEvenly, "SpaceEvenly"},
		{JustifyContent(99), "Unknown"},
	}

	for _, tt := range tests {
		got := tt.j.String()
		if got != tt.want {
			t.Errorf("JustifyContent(%d).String() = %q, want %q", tt.j, got, tt.want)
		}
	}
}

func TestAlignItems_String(t *testing.T) {
	tests := []struct {
		a    AlignItems
		want string
	}{
		{AlignStart, "Start"},
		{AlignEnd, "End"},
		{AlignCenter, "Center"},
		{AlignStretch, "Stretch"},
		{AlignBaseline, "Baseline"},
		{AlignItems(99), "Unknown"},
	}

	for _, tt := range tests {
		got := tt.a.String()
		if got != tt.want {
			t.Errorf("AlignItems(%d).String() = %q, want %q", tt.a, got, tt.want)
		}
	}
}

func TestWrapMode_String(t *testing.T) {
	tests := []struct {
		w    WrapMode
		want string
	}{
		{NoWrap, "NoWrap"},
		{Wrap, "Wrap"},
		{WrapReverse, "WrapReverse"},
		{WrapMode(99), "Unknown"},
	}

	for _, tt := range tests {
		got := tt.w.String()
		if got != tt.want {
			t.Errorf("WrapMode(%d).String() = %q, want %q", tt.w, got, tt.want)
		}
	}
}

func TestNewFlexContainer(t *testing.T) {
	flex := NewFlexContainer(Row, JustifyCenter, AlignStretch)

	if flex.Direction != Row {
		t.Errorf("Direction = %v, want Row", flex.Direction)
	}
	if flex.JustifyContent != JustifyCenter {
		t.Errorf("JustifyContent = %v, want JustifyCenter", flex.JustifyContent)
	}
	if flex.AlignItems != AlignStretch {
		t.Errorf("AlignItems = %v, want AlignStretch", flex.AlignItems)
	}
	if len(flex.Items) != 0 {
		t.Errorf("Items length = %d, want 0", len(flex.Items))
	}
}

func TestFlexContainer_ID(t *testing.T) {
	flex := NewFlexContainer(Row, JustifyStart, AlignStart)

	if flex.ID() != 0 {
		t.Errorf("ID() = %d, want 0", flex.ID())
	}

	flex.SetID(42)

	if flex.ID() != 42 {
		t.Errorf("ID() = %d, want 42", flex.ID())
	}
}

func TestFlexContainer_AddChild(t *testing.T) {
	flex := NewFlexContainer(Row, JustifyStart, AlignStart)

	child := &mockLayoutable{preferredSize: geometry.Sz(100, 50)}
	flex.AddChild(child)

	if len(flex.Items) != 1 {
		t.Fatalf("Items length = %d, want 1", len(flex.Items))
	}

	item := flex.Items[0]
	if item.Element != child {
		t.Error("item.Element mismatch")
	}
	if item.Grow != 0 {
		t.Errorf("item.Grow = %v, want 0", item.Grow)
	}
	if item.Shrink != 1 {
		t.Errorf("item.Shrink = %v, want 1", item.Shrink)
	}
}

func TestFlexContainer_AddChildWithFlex(t *testing.T) {
	flex := NewFlexContainer(Row, JustifyStart, AlignStart)

	child := &mockLayoutable{preferredSize: geometry.Sz(100, 50)}
	flex.AddChildWithFlex(child, 2.0, 0.5, 50)

	if len(flex.Items) != 1 {
		t.Fatalf("Items length = %d, want 1", len(flex.Items))
	}

	item := flex.Items[0]
	if item.Grow != 2.0 {
		t.Errorf("item.Grow = %v, want 2.0", item.Grow)
	}
	if item.Shrink != 0.5 {
		t.Errorf("item.Shrink = %v, want 0.5", item.Shrink)
	}
	if item.Basis != 50 {
		t.Errorf("item.Basis = %v, want 50", item.Basis)
	}
}

func TestFlexContainer_Clear(t *testing.T) {
	flex := NewFlexContainer(Row, JustifyStart, AlignStart)
	flex.AddChild(&mockLayoutable{preferredSize: geometry.Sz(100, 50)})
	flex.AddChild(&mockLayoutable{preferredSize: geometry.Sz(100, 50)})

	flex.Clear()

	if len(flex.Items) != 0 {
		t.Errorf("Items length = %d, want 0", len(flex.Items))
	}
}

func TestFlexContainer_Children(t *testing.T) {
	flex := NewFlexContainer(Row, JustifyStart, AlignStart)
	child1 := &mockLayoutable{preferredSize: geometry.Sz(100, 50)}
	child2 := &mockLayoutable{preferredSize: geometry.Sz(100, 50)}
	flex.AddChild(child1)
	flex.AddChild(child2)

	children := flex.Children()

	if len(children) != 2 {
		t.Fatalf("Children length = %d, want 2", len(children))
	}
	if children[0] != child1 {
		t.Error("children[0] mismatch")
	}
	if children[1] != child2 {
		t.Error("children[1] mismatch")
	}
}

func TestFlexContainer_Layout_Empty(t *testing.T) {
	flex := NewFlexContainer(Row, JustifyStart, AlignStart)

	size := flex.Layout(geometry.Loose(geometry.Sz(200, 100)))

	if size.Width != 0 {
		t.Errorf("Width = %v, want 0", size.Width)
	}
	if size.Height != 0 {
		t.Errorf("Height = %v, want 0", size.Height)
	}
}

func TestFlexContainer_Layout_Row(t *testing.T) {
	flex := NewFlexContainer(Row, JustifyStart, AlignStart)
	flex.AddChild(&mockLayoutable{preferredSize: geometry.Sz(50, 30)})
	flex.AddChild(&mockLayoutable{preferredSize: geometry.Sz(60, 40)})

	size := flex.Layout(geometry.Loose(geometry.Sz(200, 100)))

	// Total width = 50 + 60 = 110, height = max(30, 40) = 40
	if size.Width != 110 {
		t.Errorf("Width = %v, want 110", size.Width)
	}
	if size.Height != 40 {
		t.Errorf("Height = %v, want 40", size.Height)
	}

	// Check positions
	pos0 := flex.ItemPosition(0)
	if pos0.X != 0 {
		t.Errorf("Item 0 X = %v, want 0", pos0.X)
	}

	pos1 := flex.ItemPosition(1)
	if pos1.X != 50 {
		t.Errorf("Item 1 X = %v, want 50", pos1.X)
	}
}

func TestFlexContainer_Layout_Column(t *testing.T) {
	flex := NewFlexContainer(Column, JustifyStart, AlignStart)
	flex.AddChild(&mockLayoutable{preferredSize: geometry.Sz(50, 30)})
	flex.AddChild(&mockLayoutable{preferredSize: geometry.Sz(60, 40)})

	size := flex.Layout(geometry.Loose(geometry.Sz(200, 200)))

	// Width = max(50, 60) = 60, Total height = 30 + 40 = 70
	if size.Width != 60 {
		t.Errorf("Width = %v, want 60", size.Width)
	}
	if size.Height != 70 {
		t.Errorf("Height = %v, want 70", size.Height)
	}

	// Check positions
	pos0 := flex.ItemPosition(0)
	if pos0.Y != 0 {
		t.Errorf("Item 0 Y = %v, want 0", pos0.Y)
	}

	pos1 := flex.ItemPosition(1)
	if pos1.Y != 30 {
		t.Errorf("Item 1 Y = %v, want 30", pos1.Y)
	}
}

func TestFlexContainer_Layout_WithGap(t *testing.T) {
	flex := NewFlexContainer(Row, JustifyStart, AlignStart)
	flex.Gap = 10
	flex.AddChild(&mockLayoutable{preferredSize: geometry.Sz(50, 30)})
	flex.AddChild(&mockLayoutable{preferredSize: geometry.Sz(50, 30)})
	flex.AddChild(&mockLayoutable{preferredSize: geometry.Sz(50, 30)})

	size := flex.Layout(geometry.Loose(geometry.Sz(300, 100)))

	// Total width = 50*3 + 10*2 = 170
	if size.Width != 170 {
		t.Errorf("Width = %v, want 170", size.Width)
	}

	// Check positions with gaps
	pos0 := flex.ItemPosition(0)
	pos1 := flex.ItemPosition(1)
	pos2 := flex.ItemPosition(2)

	if pos0.X != 0 {
		t.Errorf("Item 0 X = %v, want 0", pos0.X)
	}
	if pos1.X != 60 { // 50 + 10
		t.Errorf("Item 1 X = %v, want 60", pos1.X)
	}
	if pos2.X != 120 { // 50 + 10 + 50 + 10
		t.Errorf("Item 2 X = %v, want 120", pos2.X)
	}
}

func TestFlexContainer_Layout_FlexGrow(t *testing.T) {
	flex := NewFlexContainer(Row, JustifyStart, AlignStart)
	flex.AddChildWithFlex(&mockLayoutable{preferredSize: geometry.Sz(50, 30)}, 1, 1, 0)
	flex.AddChildWithFlex(&mockLayoutable{preferredSize: geometry.Sz(50, 30)}, 2, 1, 0)

	// Available width = 300, children want 100, 200 remaining to distribute
	// Child 0 gets 200/3 = 66.67, Child 1 gets 400/3 = 133.33
	size := flex.Layout(geometry.Tight(geometry.Sz(300, 100)))

	if size.Width != 300 {
		t.Errorf("Width = %v, want 300", size.Width)
	}

	// Check that items received flex space
	size0 := flex.ItemSize(0)
	size1 := flex.ItemSize(1)

	// Item 1 should be roughly 2x the size of item 0 (excluding original preferred sizes)
	// Original: 50 each, extra: 200 to distribute (1:2 ratio)
	// Item 0: 50 + 200/3 ≈ 116.67
	// Item 1: 50 + 400/3 ≈ 183.33
	if size0.Width < 100 || size0.Width > 120 {
		t.Errorf("Item 0 width = %v, expected ~116", size0.Width)
	}
	if size1.Width < 180 || size1.Width > 190 {
		t.Errorf("Item 1 width = %v, expected ~183", size1.Width)
	}
}

func TestFlexContainer_Layout_JustifyCenter(t *testing.T) {
	flex := NewFlexContainer(Row, JustifyCenter, AlignStart)
	flex.AddChild(&mockLayoutable{preferredSize: geometry.Sz(50, 30)})
	flex.AddChild(&mockLayoutable{preferredSize: geometry.Sz(50, 30)})

	// Total content width = 100, available = 200, free space = 100
	// Center: items start at 50
	size := flex.Layout(geometry.Tight(geometry.Sz(200, 100)))

	if size.Width != 200 {
		t.Errorf("Width = %v, want 200", size.Width)
	}

	pos0 := flex.ItemPosition(0)
	if pos0.X != 50 {
		t.Errorf("Item 0 X = %v, want 50", pos0.X)
	}
}

func TestFlexContainer_Layout_JustifyEnd(t *testing.T) {
	flex := NewFlexContainer(Row, JustifyEnd, AlignStart)
	flex.AddChild(&mockLayoutable{preferredSize: geometry.Sz(50, 30)})
	flex.AddChild(&mockLayoutable{preferredSize: geometry.Sz(50, 30)})

	// Total content width = 100, available = 200, free space = 100
	// End: items start at 100
	_ = flex.Layout(geometry.Tight(geometry.Sz(200, 100)))

	pos0 := flex.ItemPosition(0)
	if pos0.X != 100 {
		t.Errorf("Item 0 X = %v, want 100", pos0.X)
	}
}

func TestFlexContainer_Layout_JustifySpaceBetween(t *testing.T) {
	flex := NewFlexContainer(Row, JustifySpaceBetween, AlignStart)
	flex.AddChild(&mockLayoutable{preferredSize: geometry.Sz(50, 30)})
	flex.AddChild(&mockLayoutable{preferredSize: geometry.Sz(50, 30)})

	// Total content width = 100, available = 200, free space = 100
	// SpaceBetween: first at 0, second at 150
	_ = flex.Layout(geometry.Tight(geometry.Sz(200, 100)))

	pos0 := flex.ItemPosition(0)
	pos1 := flex.ItemPosition(1)

	if pos0.X != 0 {
		t.Errorf("Item 0 X = %v, want 0", pos0.X)
	}
	if pos1.X != 150 {
		t.Errorf("Item 1 X = %v, want 150", pos1.X)
	}
}

func TestFlexContainer_Layout_AlignCenter(t *testing.T) {
	flex := NewFlexContainer(Row, JustifyStart, AlignCenter)
	flex.AddChild(&mockLayoutable{preferredSize: geometry.Sz(50, 30)})
	flex.AddChild(&mockLayoutable{preferredSize: geometry.Sz(50, 50)})

	// Max cross size = max(30, 50) = 50
	// Item 0 height = 30, centered at (50-30)/2 = 10
	// Item 1 height = 50, centered at (50-50)/2 = 0
	_ = flex.Layout(geometry.Tight(geometry.Sz(200, 100)))

	pos0 := flex.ItemPosition(0)
	if pos0.Y != 10 {
		t.Errorf("Item 0 Y = %v, want 10", pos0.Y)
	}

	pos1 := flex.ItemPosition(1)
	if pos1.Y != 0 {
		t.Errorf("Item 1 Y = %v, want 0", pos1.Y)
	}
}

func TestFlexContainer_Layout_RowReverse(t *testing.T) {
	flex := NewFlexContainer(RowReverse, JustifyStart, AlignStart)
	flex.AddChild(&mockLayoutable{preferredSize: geometry.Sz(50, 30)})
	flex.AddChild(&mockLayoutable{preferredSize: geometry.Sz(60, 30)})

	// RowReverse: items positioned from right to left
	_ = flex.Layout(geometry.Tight(geometry.Sz(200, 100)))

	pos0 := flex.ItemPosition(0)
	pos1 := flex.ItemPosition(1)

	// First item at right edge - 50
	if pos0.X != 150 {
		t.Errorf("Item 0 X = %v, want 150", pos0.X)
	}
	// Second item to the left of first
	if pos1.X != 90 {
		t.Errorf("Item 1 X = %v, want 90", pos1.X)
	}
}

func TestFlexContainer_Layout_AlignSelfOverride(t *testing.T) {
	flex := NewFlexContainer(Row, JustifyStart, AlignStart)
	// Add two children so cross axis has meaningful size
	flex.AddChild(&mockLayoutable{preferredSize: geometry.Sz(50, 80)})
	flex.AddFlexItem(FlexItem{
		Element:   &mockLayoutable{preferredSize: geometry.Sz(50, 30)},
		Grow:      0,
		Shrink:    1,
		AlignSelf: AlignSelfEnd,
	})

	// Container alignment is start, but second item overrides to end
	_ = flex.Layout(geometry.Tight(geometry.Sz(200, 100)))

	pos1 := flex.ItemPosition(1)
	// Max cross size = 80 (from first child), item 1 height = 30, aligned to end = 80-30 = 50
	if pos1.Y != 50 {
		t.Errorf("Item 1 Y = %v, want 50", pos1.Y)
	}
}

func TestFlexContainer_ItemBounds(t *testing.T) {
	flex := NewFlexContainer(Row, JustifyStart, AlignStart)
	flex.AddChild(&mockLayoutable{preferredSize: geometry.Sz(50, 30)})

	_ = flex.Layout(geometry.Loose(geometry.Sz(200, 100)))

	bounds := flex.ItemBounds(0)

	if bounds.Min.X != 0 || bounds.Min.Y != 0 {
		t.Errorf("bounds.Min = %v, want (0, 0)", bounds.Min)
	}
	if bounds.Width() != 50 {
		t.Errorf("bounds.Width() = %v, want 50", bounds.Width())
	}
	if bounds.Height() != 30 {
		t.Errorf("bounds.Height() = %v, want 30", bounds.Height())
	}
}

func TestFlexContainer_ItemPosition_OutOfBounds(t *testing.T) {
	flex := NewFlexContainer(Row, JustifyStart, AlignStart)

	pos := flex.ItemPosition(-1)
	if pos.X != 0 || pos.Y != 0 {
		t.Errorf("ItemPosition(-1) = %v, want (0, 0)", pos)
	}

	pos = flex.ItemPosition(100)
	if pos.X != 0 || pos.Y != 0 {
		t.Errorf("ItemPosition(100) = %v, want (0, 0)", pos)
	}
}

func TestFlexContainer_ItemSize_OutOfBounds(t *testing.T) {
	flex := NewFlexContainer(Row, JustifyStart, AlignStart)

	size := flex.ItemSize(-1)
	if !size.IsZero() {
		t.Errorf("ItemSize(-1) = %v, want zero", size)
	}

	size = flex.ItemSize(100)
	if !size.IsZero() {
		t.Errorf("ItemSize(100) = %v, want zero", size)
	}
}

func TestFlexContainer_NilElement(t *testing.T) {
	flex := NewFlexContainer(Row, JustifyStart, AlignStart)
	flex.AddChild(nil)
	flex.AddChild(&mockLayoutable{preferredSize: geometry.Sz(50, 30)})

	// Should handle nil elements gracefully
	size := flex.Layout(geometry.Loose(geometry.Sz(200, 100)))

	if size.Width != 50 {
		t.Errorf("Width = %v, want 50", size.Width)
	}
}

func TestFlexContainer_UnboundedConstraints(t *testing.T) {
	flex := NewFlexContainer(Row, JustifyStart, AlignStart)
	flex.AddChild(&mockLayoutable{preferredSize: geometry.Sz(50, 30)})

	// Layout with unbounded constraints
	size := flex.Layout(geometry.Expand())

	if size.Width != 50 {
		t.Errorf("Width = %v, want 50", size.Width)
	}
	if size.Height != 30 {
		t.Errorf("Height = %v, want 30", size.Height)
	}
}
