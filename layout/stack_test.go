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
		if got := tt.a.String(); got != tt.want {
			t.Errorf("StackAlignment(%d).String() = %q, want %q", tt.a, got, tt.want)
		}
	}
}

func TestStackLayout_Name(t *testing.T) {
	tests := []struct {
		direction StackDirection
		want      string
	}{
		{StackVertical, "vstack"},
		{StackHorizontal, "hstack"},
		{StackZ, "zstack"},
		{StackDirection(99), "stack"},
	}

	for _, tt := range tests {
		s := &StackLayout{Direction: tt.direction}
		if got := s.Name(); got != tt.want {
			t.Errorf("StackLayout{Direction: %d}.Name() = %q, want %q", tt.direction, got, tt.want)
		}
	}
}

func TestVStack_Basic(t *testing.T) {
	stack := &StackLayout{Direction: StackVertical, Alignment: StackAlignStart}
	tree := newTestTree()

	tree.AddChild(1, 10)
	tree.AddChild(1, 11)
	tree.AddChild(1, 12)

	tree.SetPreferredSize(10, geometry.Size{Width: 100, Height: 50})
	tree.SetPreferredSize(11, geometry.Size{Width: 150, Height: 60})
	tree.SetPreferredSize(12, geometry.Size{Width: 80, Height: 40})

	result := stack.Compute(tree, 1, geometry.Size{Width: 200, Height: 400})

	// Width should be max child width (150)
	if result.Size.Width != 150 {
		t.Errorf("Result.Size.Width = %v, want 150", result.Size.Width)
	}

	// Height should be sum (50 + 60 + 40 = 150)
	if result.Size.Height != 150 {
		t.Errorf("Result.Size.Height = %v, want 150", result.Size.Height)
	}

	// Check positions
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

func TestVStack_WithSpacing(t *testing.T) {
	stack := &StackLayout{Direction: StackVertical, Spacing: 10}
	tree := newTestTree()

	tree.AddChild(1, 10)
	tree.AddChild(1, 11)
	tree.AddChild(1, 12)

	tree.SetPreferredSize(10, geometry.Size{Width: 100, Height: 50})
	tree.SetPreferredSize(11, geometry.Size{Width: 100, Height: 60})
	tree.SetPreferredSize(12, geometry.Size{Width: 100, Height: 40})

	result := stack.Compute(tree, 1, geometry.Size{Width: 200, Height: 400})

	// Height = 50 + 10 + 60 + 10 + 40 = 170
	if result.Size.Height != 170 {
		t.Errorf("Result.Size.Height = %v, want 170", result.Size.Height)
	}

	layout11 := tree.GetLayout(11)
	if layout11.Position.Y != 60 { // 50 + 10
		t.Errorf("child 11 Y = %v, want 60", layout11.Position.Y)
	}

	layout12 := tree.GetLayout(12)
	if layout12.Position.Y != 130 { // 50 + 10 + 60 + 10
		t.Errorf("child 12 Y = %v, want 130", layout12.Position.Y)
	}
}

func TestVStack_AlignCenter(t *testing.T) {
	stack := &StackLayout{Direction: StackVertical, Alignment: StackAlignCenter}
	tree := newTestTree()

	tree.AddChild(1, 10)
	tree.SetPreferredSize(10, geometry.Size{Width: 100, Height: 50})

	_ = stack.Compute(tree, 1, geometry.Size{Width: 200, Height: 200})

	layout := tree.GetLayout(10)
	// Centered: (100 - 100) / 2 = 0 (width = max width = 100)
	if layout.Position.X != 0 {
		t.Errorf("child X = %v, want 0 (centered at max width)", layout.Position.X)
	}
}

func TestVStack_AlignEnd(t *testing.T) {
	stack := &StackLayout{Direction: StackVertical, Alignment: StackAlignEnd}
	tree := newTestTree()

	tree.AddChild(1, 10)
	tree.AddChild(1, 11)

	tree.SetPreferredSize(10, geometry.Size{Width: 50, Height: 50})
	tree.SetPreferredSize(11, geometry.Size{Width: 100, Height: 50})

	_ = stack.Compute(tree, 1, geometry.Size{Width: 200, Height: 200})

	layout10 := tree.GetLayout(10)
	// End aligned: max width (100) - child width (50) = 50
	if layout10.Position.X != 50 {
		t.Errorf("child 10 X = %v, want 50 (end aligned)", layout10.Position.X)
	}
}

func TestHStack_Basic(t *testing.T) {
	stack := &StackLayout{Direction: StackHorizontal, Alignment: StackAlignStart}
	tree := newTestTree()

	tree.AddChild(1, 10)
	tree.AddChild(1, 11)
	tree.AddChild(1, 12)

	tree.SetPreferredSize(10, geometry.Size{Width: 100, Height: 50})
	tree.SetPreferredSize(11, geometry.Size{Width: 150, Height: 60})
	tree.SetPreferredSize(12, geometry.Size{Width: 80, Height: 40})

	result := stack.Compute(tree, 1, geometry.Size{Width: 500, Height: 200})

	// Width should be sum (100 + 150 + 80 = 330)
	if result.Size.Width != 330 {
		t.Errorf("Result.Size.Width = %v, want 330", result.Size.Width)
	}

	// Height should be max child height (60)
	if result.Size.Height != 60 {
		t.Errorf("Result.Size.Height = %v, want 60", result.Size.Height)
	}

	// Check positions
	layout10 := tree.GetLayout(10)
	if layout10.Position.X != 0 {
		t.Errorf("child 10 X = %v, want 0", layout10.Position.X)
	}

	layout11 := tree.GetLayout(11)
	if layout11.Position.X != 100 {
		t.Errorf("child 11 X = %v, want 100", layout11.Position.X)
	}

	layout12 := tree.GetLayout(12)
	if layout12.Position.X != 250 {
		t.Errorf("child 12 X = %v, want 250", layout12.Position.X)
	}
}

func TestHStack_WithSpacing(t *testing.T) {
	stack := &StackLayout{Direction: StackHorizontal, Spacing: 20}
	tree := newTestTree()

	tree.AddChild(1, 10)
	tree.AddChild(1, 11)

	tree.SetPreferredSize(10, geometry.Size{Width: 100, Height: 50})
	tree.SetPreferredSize(11, geometry.Size{Width: 100, Height: 50})

	result := stack.Compute(tree, 1, geometry.Size{Width: 500, Height: 100})

	// Width = 100 + 20 + 100 = 220
	if result.Size.Width != 220 {
		t.Errorf("Result.Size.Width = %v, want 220", result.Size.Width)
	}

	layout11 := tree.GetLayout(11)
	if layout11.Position.X != 120 { // 100 + 20
		t.Errorf("child 11 X = %v, want 120", layout11.Position.X)
	}
}

func TestZStack_Basic(t *testing.T) {
	stack := &StackLayout{Direction: StackZ, Alignment: StackAlignCenter}
	tree := newTestTree()

	tree.AddChild(1, 10)
	tree.AddChild(1, 11)

	tree.SetPreferredSize(10, geometry.Size{Width: 200, Height: 150})
	tree.SetPreferredSize(11, geometry.Size{Width: 100, Height: 80})

	result := stack.Compute(tree, 1, geometry.Size{Width: 300, Height: 300})

	// Size should be max of children
	if result.Size.Width != 200 {
		t.Errorf("Result.Size.Width = %v, want 200", result.Size.Width)
	}
	if result.Size.Height != 150 {
		t.Errorf("Result.Size.Height = %v, want 150", result.Size.Height)
	}

	// First child should be centered
	layout10 := tree.GetLayout(10)
	if layout10.Position.X != 0 || layout10.Position.Y != 0 {
		t.Errorf("child 10 Position = %v, want {0, 0}", layout10.Position)
	}

	// Second child (smaller) should be centered
	layout11 := tree.GetLayout(11)
	// Centered: (200 - 100) / 2 = 50, (150 - 80) / 2 = 35
	if layout11.Position.X != 50 {
		t.Errorf("child 11 X = %v, want 50", layout11.Position.X)
	}
	if layout11.Position.Y != 35 {
		t.Errorf("child 11 Y = %v, want 35", layout11.Position.Y)
	}
}

func TestZStackLayout_Alignments(t *testing.T) {
	// With two children of different sizes, we can test alignment
	// Stack size becomes max of children: 200x150
	// Smaller child (100x80) gets positioned according to alignment
	tests := []struct {
		alignment ZStackAlignment
		wantX     float32
		wantY     float32
	}{
		{ZAlignTopLeft, 0, 0},
		{ZAlignTop, 50, 0},       // (200-100)/2 = 50
		{ZAlignTopRight, 100, 0}, // 200-100 = 100
		{ZAlignLeft, 0, 35},      // (150-80)/2 = 35
		{ZAlignCenter, 50, 35},
		{ZAlignRight, 100, 35},
		{ZAlignBottomLeft, 0, 70}, // 150-80 = 70
		{ZAlignBottom, 50, 70},
		{ZAlignBottomRight, 100, 70},
	}

	for _, tt := range tests {
		t.Run(tt.alignment.String(), func(t *testing.T) {
			stack := &ZStackLayout{Alignment: tt.alignment}
			tree := newTestTree()

			// First child is larger (defines stack size)
			tree.AddChild(1, 10)
			tree.AddChild(1, 11)
			tree.SetPreferredSize(10, geometry.Size{Width: 200, Height: 150})
			tree.SetPreferredSize(11, geometry.Size{Width: 100, Height: 80})

			_ = stack.Compute(tree, 1, geometry.Size{Width: 300, Height: 300})

			// Check alignment of smaller child (11)
			layout := tree.GetLayout(11)
			if layout.Position.X != tt.wantX {
				t.Errorf("X = %v, want %v", layout.Position.X, tt.wantX)
			}
			if layout.Position.Y != tt.wantY {
				t.Errorf("Y = %v, want %v", layout.Position.Y, tt.wantY)
			}
		})
	}
}

// String representation for ZStackAlignment (for test names).
func (a ZStackAlignment) String() string {
	names := []string{
		"TopLeft", "Top", "TopRight",
		"Left", "Center", "Right",
		"BottomLeft", "Bottom", "BottomRight",
	}
	if int(a) < len(names) {
		return names[a]
	}
	return "Unknown"
}

func TestStack_Empty(t *testing.T) {
	tests := []struct {
		name      string
		direction StackDirection
	}{
		{"vstack empty", StackVertical},
		{"hstack empty", StackHorizontal},
		{"zstack empty", StackZ},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stack := &StackLayout{Direction: tt.direction}
			tree := newTestTree()

			result := stack.Compute(tree, 1, geometry.Size{Width: 200, Height: 200})

			if !result.Size.IsZero() {
				t.Errorf("Result.Size = %v, want zero", result.Size)
			}
		})
	}
}

func TestStack_Registered(t *testing.T) {
	// All stacks should be registered via init()
	tests := []string{"vstack", "hstack", "zstack"}

	for _, name := range tests {
		if !Has(name) {
			t.Errorf("%s layout should be registered", name)
		}
	}
}

func TestStack_Overflow(t *testing.T) {
	stack := &StackLayout{Direction: StackVertical}
	tree := newTestTree()

	tree.AddChild(1, 10)
	tree.SetPreferredSize(10, geometry.Size{Width: 100, Height: 300})

	result := stack.Compute(tree, 1, geometry.Size{Width: 200, Height: 100})

	if !result.Overflow {
		t.Error("Result.Overflow should be true when content exceeds available space")
	}
}
