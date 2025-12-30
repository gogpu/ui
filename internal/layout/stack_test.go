package layout

import (
	"testing"

	"github.com/gogpu/ui/geometry"
)

func TestStackAlignment_String(t *testing.T) {
	tests := []struct {
		a    StackAlignment
		want string
	}{
		{StackAlignStart, "Start"},
		{StackAlignCenter, "Center"},
		{StackAlignEnd, "End"},
		{StackAlignStretch, "Stretch"},
		{StackAlignment(99), "Unknown"},
	}

	for _, tt := range tests {
		got := tt.a.String()
		if got != tt.want {
			t.Errorf("StackAlignment(%d).String() = %q, want %q", tt.a, got, tt.want)
		}
	}
}

func TestZStackAlignment_String(t *testing.T) {
	tests := []struct {
		a    ZStackAlignment
		want string
	}{
		{ZAlignTopLeft, "TopLeft"},
		{ZAlignTop, "Top"},
		{ZAlignTopRight, "TopRight"},
		{ZAlignLeft, "Left"},
		{ZAlignCenter, "Center"},
		{ZAlignRight, "Right"},
		{ZAlignBottomLeft, "BottomLeft"},
		{ZAlignBottom, "Bottom"},
		{ZAlignBottomRight, "BottomRight"},
		{ZStackAlignment(99), "Unknown"},
	}

	for _, tt := range tests {
		got := tt.a.String()
		if got != tt.want {
			t.Errorf("ZStackAlignment(%d).String() = %q, want %q", tt.a, got, tt.want)
		}
	}
}

// VStack tests

func TestNewVStack(t *testing.T) {
	vstack := NewVStack(10, StackAlignCenter)

	if vstack.Spacing != 10 {
		t.Errorf("Spacing = %v, want 10", vstack.Spacing)
	}
	if vstack.Alignment != StackAlignCenter {
		t.Errorf("Alignment = %v, want StackAlignCenter", vstack.Alignment)
	}
	if len(vstack.Children) != 0 {
		t.Errorf("Children length = %d, want 0", len(vstack.Children))
	}
}

func TestVStack_ID(t *testing.T) {
	vstack := NewVStack(0, StackAlignStart)

	if vstack.ID() != 0 {
		t.Errorf("ID() = %d, want 0", vstack.ID())
	}

	vstack.SetID(42)

	if vstack.ID() != 42 {
		t.Errorf("ID() = %d, want 42", vstack.ID())
	}
}

func TestVStack_AddChild(t *testing.T) {
	vstack := NewVStack(0, StackAlignStart)
	child := &mockLayoutable{preferredSize: geometry.Sz(100, 50)}
	vstack.AddChild(child)

	if len(vstack.Children) != 1 {
		t.Fatalf("Children length = %d, want 1", len(vstack.Children))
	}
	if vstack.Children[0].Element != child {
		t.Error("Child element mismatch")
	}
}

func TestVStack_Clear(t *testing.T) {
	vstack := NewVStack(0, StackAlignStart)
	vstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(100, 50)})
	vstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(100, 50)})

	vstack.Clear()

	if len(vstack.Children) != 0 {
		t.Errorf("Children length = %d, want 0", len(vstack.Children))
	}
}

func TestVStack_Layout_Empty(t *testing.T) {
	vstack := NewVStack(0, StackAlignStart)

	size := vstack.Layout(geometry.Loose(geometry.Sz(200, 200)))

	if size.Width != 0 || size.Height != 0 {
		t.Errorf("Size = %v, want (0, 0)", size)
	}
}

func TestVStack_Layout_Basic(t *testing.T) {
	vstack := NewVStack(0, StackAlignStart)
	vstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(100, 50)})
	vstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(80, 40)})

	size := vstack.Layout(geometry.Loose(geometry.Sz(200, 200)))

	// Width = max(100, 80) = 100, Height = 50 + 40 = 90
	if size.Width != 100 {
		t.Errorf("Width = %v, want 100", size.Width)
	}
	if size.Height != 90 {
		t.Errorf("Height = %v, want 90", size.Height)
	}

	// Check positions
	pos0 := vstack.ChildPosition(0)
	pos1 := vstack.ChildPosition(1)

	if pos0.Y != 0 {
		t.Errorf("Child 0 Y = %v, want 0", pos0.Y)
	}
	if pos1.Y != 50 {
		t.Errorf("Child 1 Y = %v, want 50", pos1.Y)
	}
}

func TestVStack_Layout_WithSpacing(t *testing.T) {
	vstack := NewVStack(10, StackAlignStart)
	vstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(100, 50)})
	vstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(100, 50)})
	vstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(100, 50)})

	size := vstack.Layout(geometry.Loose(geometry.Sz(200, 300)))

	// Height = 50*3 + 10*2 = 170
	if size.Height != 170 {
		t.Errorf("Height = %v, want 170", size.Height)
	}

	pos0 := vstack.ChildPosition(0)
	pos1 := vstack.ChildPosition(1)
	pos2 := vstack.ChildPosition(2)

	if pos0.Y != 0 {
		t.Errorf("Child 0 Y = %v, want 0", pos0.Y)
	}
	if pos1.Y != 60 { // 50 + 10
		t.Errorf("Child 1 Y = %v, want 60", pos1.Y)
	}
	if pos2.Y != 120 { // 50 + 10 + 50 + 10
		t.Errorf("Child 2 Y = %v, want 120", pos2.Y)
	}
}

func TestVStack_Layout_AlignCenter(t *testing.T) {
	vstack := NewVStack(0, StackAlignCenter)
	vstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(60, 50)})
	vstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(100, 50)})

	_ = vstack.Layout(geometry.Loose(geometry.Sz(200, 200)))

	pos0 := vstack.ChildPosition(0)
	pos1 := vstack.ChildPosition(1)

	// Max width = 100, child 0 width = 60, centered at (100-60)/2 = 20
	if pos0.X != 20 {
		t.Errorf("Child 0 X = %v, want 20", pos0.X)
	}
	// Child 1 width = 100, centered at 0
	if pos1.X != 0 {
		t.Errorf("Child 1 X = %v, want 0", pos1.X)
	}
}

func TestVStack_Layout_AlignEnd(t *testing.T) {
	vstack := NewVStack(0, StackAlignEnd)
	vstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(60, 50)})
	vstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(100, 50)})

	_ = vstack.Layout(geometry.Loose(geometry.Sz(200, 200)))

	pos0 := vstack.ChildPosition(0)

	// Max width = 100, child 0 width = 60, right-aligned at 100-60 = 40
	if pos0.X != 40 {
		t.Errorf("Child 0 X = %v, want 40", pos0.X)
	}
}

func TestVStack_ChildBounds(t *testing.T) {
	vstack := NewVStack(0, StackAlignStart)
	vstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(100, 50)})

	_ = vstack.Layout(geometry.Loose(geometry.Sz(200, 200)))

	bounds := vstack.ChildBounds(0)

	if bounds.Min.X != 0 || bounds.Min.Y != 0 {
		t.Errorf("bounds.Min = %v, want (0, 0)", bounds.Min)
	}
	if bounds.Width() != 100 {
		t.Errorf("bounds.Width() = %v, want 100", bounds.Width())
	}
	if bounds.Height() != 50 {
		t.Errorf("bounds.Height() = %v, want 50", bounds.Height())
	}
}

func TestVStack_ChildPosition_OutOfBounds(t *testing.T) {
	vstack := NewVStack(0, StackAlignStart)

	pos := vstack.ChildPosition(-1)
	if pos.X != 0 || pos.Y != 0 {
		t.Errorf("ChildPosition(-1) = %v, want (0, 0)", pos)
	}

	pos = vstack.ChildPosition(100)
	if pos.X != 0 || pos.Y != 0 {
		t.Errorf("ChildPosition(100) = %v, want (0, 0)", pos)
	}
}

func TestVStack_NilElement(t *testing.T) {
	vstack := NewVStack(0, StackAlignStart)
	vstack.AddChild(nil)
	vstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(100, 50)})

	size := vstack.Layout(geometry.Loose(geometry.Sz(200, 200)))

	if size.Width != 100 {
		t.Errorf("Width = %v, want 100", size.Width)
	}
}

// HStack tests

func TestNewHStack(t *testing.T) {
	hstack := NewHStack(10, StackAlignCenter)

	if hstack.Spacing != 10 {
		t.Errorf("Spacing = %v, want 10", hstack.Spacing)
	}
	if hstack.Alignment != StackAlignCenter {
		t.Errorf("Alignment = %v, want StackAlignCenter", hstack.Alignment)
	}
}

func TestHStack_ID(t *testing.T) {
	hstack := NewHStack(0, StackAlignStart)

	if hstack.ID() != 0 {
		t.Errorf("ID() = %d, want 0", hstack.ID())
	}

	hstack.SetID(42)

	if hstack.ID() != 42 {
		t.Errorf("ID() = %d, want 42", hstack.ID())
	}
}

func TestHStack_Layout_Basic(t *testing.T) {
	hstack := NewHStack(0, StackAlignStart)
	hstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(50, 100)})
	hstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(40, 80)})

	size := hstack.Layout(geometry.Loose(geometry.Sz(200, 200)))

	// Width = 50 + 40 = 90, Height = max(100, 80) = 100
	if size.Width != 90 {
		t.Errorf("Width = %v, want 90", size.Width)
	}
	if size.Height != 100 {
		t.Errorf("Height = %v, want 100", size.Height)
	}

	pos0 := hstack.ChildPosition(0)
	pos1 := hstack.ChildPosition(1)

	if pos0.X != 0 {
		t.Errorf("Child 0 X = %v, want 0", pos0.X)
	}
	if pos1.X != 50 {
		t.Errorf("Child 1 X = %v, want 50", pos1.X)
	}
}

func TestHStack_Layout_WithSpacing(t *testing.T) {
	hstack := NewHStack(10, StackAlignStart)
	hstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(50, 100)})
	hstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(50, 100)})

	size := hstack.Layout(geometry.Loose(geometry.Sz(200, 200)))

	// Width = 50*2 + 10 = 110
	if size.Width != 110 {
		t.Errorf("Width = %v, want 110", size.Width)
	}

	pos1 := hstack.ChildPosition(1)
	if pos1.X != 60 { // 50 + 10
		t.Errorf("Child 1 X = %v, want 60", pos1.X)
	}
}

func TestHStack_Layout_AlignCenter(t *testing.T) {
	hstack := NewHStack(0, StackAlignCenter)
	hstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(50, 60)})
	hstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(50, 100)})

	_ = hstack.Layout(geometry.Loose(geometry.Sz(200, 200)))

	pos0 := hstack.ChildPosition(0)

	// Max height = 100, child 0 height = 60, centered at (100-60)/2 = 20
	if pos0.Y != 20 {
		t.Errorf("Child 0 Y = %v, want 20", pos0.Y)
	}
}

func TestHStack_Layout_Empty(t *testing.T) {
	hstack := NewHStack(0, StackAlignStart)

	size := hstack.Layout(geometry.Loose(geometry.Sz(200, 200)))

	if !size.IsZero() {
		t.Errorf("Size = %v, want zero", size)
	}
}

func TestHStack_Clear(t *testing.T) {
	hstack := NewHStack(0, StackAlignStart)
	hstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(50, 100)})

	hstack.Clear()

	if len(hstack.Children) != 0 {
		t.Errorf("Children length = %d, want 0", len(hstack.Children))
	}
}

func TestHStack_ChildBounds(t *testing.T) {
	hstack := NewHStack(0, StackAlignStart)
	hstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(50, 100)})

	_ = hstack.Layout(geometry.Loose(geometry.Sz(200, 200)))

	bounds := hstack.ChildBounds(0)

	if bounds.Width() != 50 || bounds.Height() != 100 {
		t.Errorf("bounds = %v, want size (50, 100)", bounds)
	}
}

func TestHStack_ChildPosition_OutOfBounds(t *testing.T) {
	hstack := NewHStack(0, StackAlignStart)

	pos := hstack.ChildPosition(-1)
	if !pos.IsZero() {
		t.Errorf("ChildPosition(-1) = %v, want zero", pos)
	}
}

func TestHStack_NilElement(t *testing.T) {
	hstack := NewHStack(0, StackAlignStart)
	hstack.AddChild(nil)
	hstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(50, 100)})

	size := hstack.Layout(geometry.Loose(geometry.Sz(200, 200)))

	if size.Height != 100 {
		t.Errorf("Height = %v, want 100", size.Height)
	}
}

// ZStack tests

func TestNewZStack(t *testing.T) {
	zstack := NewZStack(ZAlignCenter)

	if zstack.Alignment != ZAlignCenter {
		t.Errorf("Alignment = %v, want ZAlignCenter", zstack.Alignment)
	}
	if len(zstack.Children) != 0 {
		t.Errorf("Children length = %d, want 0", len(zstack.Children))
	}
}

func TestZStack_ID(t *testing.T) {
	zstack := NewZStack(ZAlignCenter)

	if zstack.ID() != 0 {
		t.Errorf("ID() = %d, want 0", zstack.ID())
	}

	zstack.SetID(42)

	if zstack.ID() != 42 {
		t.Errorf("ID() = %d, want 42", zstack.ID())
	}
}

func TestZStack_Layout_Empty(t *testing.T) {
	zstack := NewZStack(ZAlignCenter)

	size := zstack.Layout(geometry.Loose(geometry.Sz(200, 200)))

	if !size.IsZero() {
		t.Errorf("Size = %v, want zero", size)
	}
}

func TestZStack_Layout_Basic(t *testing.T) {
	zstack := NewZStack(ZAlignCenter)
	zstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(100, 80)})
	zstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(60, 40)})

	size := zstack.Layout(geometry.Loose(geometry.Sz(200, 200)))

	// Size = max dimensions
	if size.Width != 100 {
		t.Errorf("Width = %v, want 100", size.Width)
	}
	if size.Height != 80 {
		t.Errorf("Height = %v, want 80", size.Height)
	}
}

func TestZStack_Layout_AlignCenter(t *testing.T) {
	zstack := NewZStack(ZAlignCenter)
	zstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(100, 80)})
	zstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(60, 40)})

	_ = zstack.Layout(geometry.Loose(geometry.Sz(200, 200)))

	// First child (100x80) centered in 100x80: at (0, 0)
	pos0 := zstack.ChildPosition(0)
	if pos0.X != 0 || pos0.Y != 0 {
		t.Errorf("Child 0 position = %v, want (0, 0)", pos0)
	}

	// Second child (60x40) centered in 100x80: at (20, 20)
	pos1 := zstack.ChildPosition(1)
	if pos1.X != 20 {
		t.Errorf("Child 1 X = %v, want 20", pos1.X)
	}
	if pos1.Y != 20 {
		t.Errorf("Child 1 Y = %v, want 20", pos1.Y)
	}
}

func TestZStack_Layout_AlignTopLeft(t *testing.T) {
	zstack := NewZStack(ZAlignTopLeft)
	zstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(100, 80)})
	zstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(60, 40)})

	_ = zstack.Layout(geometry.Loose(geometry.Sz(200, 200)))

	pos1 := zstack.ChildPosition(1)
	if pos1.X != 0 || pos1.Y != 0 {
		t.Errorf("Child 1 position = %v, want (0, 0)", pos1)
	}
}

func TestZStack_Layout_AlignBottomRight(t *testing.T) {
	zstack := NewZStack(ZAlignBottomRight)
	zstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(100, 80)})
	zstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(60, 40)})

	_ = zstack.Layout(geometry.Loose(geometry.Sz(200, 200)))

	// Second child (60x40) at bottom-right of 100x80: (40, 40)
	pos1 := zstack.ChildPosition(1)
	if pos1.X != 40 {
		t.Errorf("Child 1 X = %v, want 40", pos1.X)
	}
	if pos1.Y != 40 {
		t.Errorf("Child 1 Y = %v, want 40", pos1.Y)
	}
}

func TestZStack_Layout_AllAlignments(t *testing.T) {
	// Test all alignment positions
	tests := []struct {
		alignment ZStackAlignment
		wantX     float32
		wantY     float32
	}{
		{ZAlignTopLeft, 0, 0},
		{ZAlignTop, 20, 0},
		{ZAlignTopRight, 40, 0},
		{ZAlignLeft, 0, 20},
		{ZAlignCenter, 20, 20},
		{ZAlignRight, 40, 20},
		{ZAlignBottomLeft, 0, 40},
		{ZAlignBottom, 20, 40},
		{ZAlignBottomRight, 40, 40},
	}

	for _, tt := range tests {
		t.Run(tt.alignment.String(), func(t *testing.T) {
			zstack := NewZStack(tt.alignment)
			zstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(100, 80)})
			zstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(60, 40)})

			_ = zstack.Layout(geometry.Loose(geometry.Sz(200, 200)))

			pos1 := zstack.ChildPosition(1)
			if pos1.X != tt.wantX {
				t.Errorf("X = %v, want %v", pos1.X, tt.wantX)
			}
			if pos1.Y != tt.wantY {
				t.Errorf("Y = %v, want %v", pos1.Y, tt.wantY)
			}
		})
	}
}

func TestZStack_Clear(t *testing.T) {
	zstack := NewZStack(ZAlignCenter)
	zstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(100, 80)})

	zstack.Clear()

	if len(zstack.Children) != 0 {
		t.Errorf("Children length = %d, want 0", len(zstack.Children))
	}
}

func TestZStack_ChildBounds(t *testing.T) {
	zstack := NewZStack(ZAlignCenter)
	zstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(100, 80)})

	_ = zstack.Layout(geometry.Loose(geometry.Sz(200, 200)))

	bounds := zstack.ChildBounds(0)

	if bounds.Width() != 100 || bounds.Height() != 80 {
		t.Errorf("bounds size = (%v, %v), want (100, 80)", bounds.Width(), bounds.Height())
	}
}

func TestZStack_ChildPosition_OutOfBounds(t *testing.T) {
	zstack := NewZStack(ZAlignCenter)

	pos := zstack.ChildPosition(-1)
	if !pos.IsZero() {
		t.Errorf("ChildPosition(-1) = %v, want zero", pos)
	}

	pos = zstack.ChildPosition(100)
	if !pos.IsZero() {
		t.Errorf("ChildPosition(100) = %v, want zero", pos)
	}
}

func TestZStack_NilElement(t *testing.T) {
	zstack := NewZStack(ZAlignCenter)
	zstack.AddChild(nil)
	zstack.AddChild(&mockLayoutable{preferredSize: geometry.Sz(100, 80)})

	size := zstack.Layout(geometry.Loose(geometry.Sz(200, 200)))

	if size.Width != 100 {
		t.Errorf("Width = %v, want 100", size.Width)
	}
}

// Test ChildLayoutables for interface compliance

func TestVStack_ChildLayoutables(t *testing.T) {
	vstack := NewVStack(0, StackAlignStart)
	child := &mockLayoutable{preferredSize: geometry.Sz(100, 50)}
	vstack.AddChild(child)

	children := vstack.ChildLayoutables()

	if len(children) != 1 {
		t.Fatalf("Children length = %d, want 1", len(children))
	}
	if children[0] != child {
		t.Error("Child mismatch")
	}
}

func TestHStack_ChildLayoutables(t *testing.T) {
	hstack := NewHStack(0, StackAlignStart)
	child := &mockLayoutable{preferredSize: geometry.Sz(50, 100)}
	hstack.AddChild(child)

	children := hstack.ChildLayoutables()

	if len(children) != 1 {
		t.Fatalf("Children length = %d, want 1", len(children))
	}
	if children[0] != child {
		t.Error("Child mismatch")
	}
}

func TestZStack_ChildLayoutables(t *testing.T) {
	zstack := NewZStack(ZAlignCenter)
	child := &mockLayoutable{preferredSize: geometry.Sz(100, 80)}
	zstack.AddChild(child)

	children := zstack.ChildLayoutables()

	if len(children) != 1 {
		t.Fatalf("Children length = %d, want 1", len(children))
	}
	if children[0] != child {
		t.Error("Child mismatch")
	}
}
