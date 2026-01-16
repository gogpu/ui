package layout

import (
	"testing"

	"github.com/gogpu/ui/geometry"
)

func TestGridLayout_Name(t *testing.T) {
	grid := &GridLayout{}
	if name := grid.Name(); name != "grid" {
		t.Errorf("Name() = %q, want %q", name, "grid")
	}
}

func TestGridLayout_Empty(t *testing.T) {
	grid := &GridLayout{}
	tree := newTestTree()

	result := grid.Compute(tree, 1, geometry.Size{Width: 300, Height: 200})

	if !result.Size.IsZero() {
		t.Errorf("Result.Size = %v, want zero", result.Size)
	}
}

func TestGridLayout_SingleColumn(t *testing.T) {
	grid := &GridLayout{
		Columns: []GridTrack{FractionTrack(1)},
	}
	tree := newTestTree()

	tree.AddChild(1, 10)
	tree.AddChild(1, 11)
	tree.AddChild(1, 12)

	tree.SetPreferredSize(10, geometry.Size{Width: 100, Height: 50})
	tree.SetPreferredSize(11, geometry.Size{Width: 100, Height: 60})
	tree.SetPreferredSize(12, geometry.Size{Width: 100, Height: 40})

	result := grid.Compute(tree, 1, geometry.Size{Width: 200, Height: 300})

	// Single column should use full width
	if result.Size.Width != 200 {
		t.Errorf("Result.Size.Width = %v, want 200", result.Size.Width)
	}

	// Check that children are stacked vertically
	layout10 := tree.GetLayout(10)
	if layout10.Position.Y != 0 {
		t.Errorf("child 10 Y = %v, want 0", layout10.Position.Y)
	}

	layout11 := tree.GetLayout(11)
	if layout11.Position.Y != 50 {
		t.Errorf("child 11 Y = %v, want 50", layout11.Position.Y)
	}

	layout12 := tree.GetLayout(12)
	if layout12.Position.Y != 110 { // 50 + 60
		t.Errorf("child 12 Y = %v, want 110", layout12.Position.Y)
	}
}

func TestGridLayout_TwoColumns(t *testing.T) {
	grid := &GridLayout{
		Columns: []GridTrack{FractionTrack(1), FractionTrack(1)},
	}
	tree := newTestTree()

	tree.AddChild(1, 10)
	tree.AddChild(1, 11)
	tree.AddChild(1, 12)
	tree.AddChild(1, 13)

	tree.SetPreferredSize(10, geometry.Size{Width: 50, Height: 50})
	tree.SetPreferredSize(11, geometry.Size{Width: 50, Height: 50})
	tree.SetPreferredSize(12, geometry.Size{Width: 50, Height: 50})
	tree.SetPreferredSize(13, geometry.Size{Width: 50, Height: 50})

	result := grid.Compute(tree, 1, geometry.Size{Width: 200, Height: 300})

	// Should have 2 columns of 100px each
	if result.Size.Width != 200 {
		t.Errorf("Result.Size.Width = %v, want 200", result.Size.Width)
	}

	// Check positions: 2x2 grid
	layout10 := tree.GetLayout(10)
	if layout10.Position.X != 0 || layout10.Position.Y != 0 {
		t.Errorf("child 10 Position = %v, want {0, 0}", layout10.Position)
	}

	layout11 := tree.GetLayout(11)
	if layout11.Position.X != 100 || layout11.Position.Y != 0 {
		t.Errorf("child 11 Position = %v, want {100, 0}", layout11.Position)
	}

	layout12 := tree.GetLayout(12)
	if layout12.Position.X != 0 || layout12.Position.Y != 50 {
		t.Errorf("child 12 Position = %v, want {0, 50}", layout12.Position)
	}

	layout13 := tree.GetLayout(13)
	if layout13.Position.X != 100 || layout13.Position.Y != 50 {
		t.Errorf("child 13 Position = %v, want {100, 50}", layout13.Position)
	}
}

func TestGridLayout_WithGap(t *testing.T) {
	grid := &GridLayout{
		Columns:   []GridTrack{FractionTrack(1), FractionTrack(1)},
		ColumnGap: 20,
		RowGap:    10,
	}
	tree := newTestTree()

	tree.AddChild(1, 10)
	tree.AddChild(1, 11)
	tree.AddChild(1, 12)
	tree.AddChild(1, 13)

	tree.SetPreferredSize(10, geometry.Size{Width: 50, Height: 50})
	tree.SetPreferredSize(11, geometry.Size{Width: 50, Height: 50})
	tree.SetPreferredSize(12, geometry.Size{Width: 50, Height: 50})
	tree.SetPreferredSize(13, geometry.Size{Width: 50, Height: 50})

	result := grid.Compute(tree, 1, geometry.Size{Width: 200, Height: 300})

	// Column width = (200 - 20) / 2 = 90
	// Row 2 Y = 50 + 10 = 60
	layout12 := tree.GetLayout(12)
	if layout12.Position.Y != 60 {
		t.Errorf("child 12 Y = %v, want 60", layout12.Position.Y)
	}

	layout11 := tree.GetLayout(11)
	if layout11.Position.X != 110 { // 90 + 20
		t.Errorf("child 11 X = %v, want 110", layout11.Position.X)
	}

	// Total height = 50 + 10 + 50 = 110
	if result.Size.Height != 110 {
		t.Errorf("Result.Size.Height = %v, want 110", result.Size.Height)
	}
}

func TestGridLayout_FixedColumns(t *testing.T) {
	grid := &GridLayout{
		Columns: []GridTrack{FixedTrack(100), FixedTrack(50)},
	}
	tree := newTestTree()

	tree.AddChild(1, 10)
	tree.AddChild(1, 11)

	tree.SetPreferredSize(10, geometry.Size{Width: 50, Height: 50})
	tree.SetPreferredSize(11, geometry.Size{Width: 50, Height: 50})

	result := grid.Compute(tree, 1, geometry.Size{Width: 300, Height: 100})

	// Total width should be 100 + 50 = 150
	if result.Size.Width != 150 {
		t.Errorf("Result.Size.Width = %v, want 150", result.Size.Width)
	}

	layout11 := tree.GetLayout(11)
	if layout11.Position.X != 100 {
		t.Errorf("child 11 X = %v, want 100", layout11.Position.X)
	}
}

func TestGridLayout_MixedTracks(t *testing.T) {
	grid := &GridLayout{
		Columns: []GridTrack{FixedTrack(50), FractionTrack(1), FractionTrack(2)},
	}
	tree := newTestTree()

	tree.AddChild(1, 10)
	tree.AddChild(1, 11)
	tree.AddChild(1, 12)

	tree.SetPreferredSize(10, geometry.Size{Width: 30, Height: 50})
	tree.SetPreferredSize(11, geometry.Size{Width: 30, Height: 50})
	tree.SetPreferredSize(12, geometry.Size{Width: 30, Height: 50})

	result := grid.Compute(tree, 1, geometry.Size{Width: 200, Height: 100})

	// Fixed: 50, remaining: 150 split 1:2 = 50, 100
	if result.Size.Width != 200 {
		t.Errorf("Result.Size.Width = %v, want 200", result.Size.Width)
	}

	layout10 := tree.GetLayout(10)
	if layout10.Position.X != 0 {
		t.Errorf("child 10 X = %v, want 0", layout10.Position.X)
	}

	layout11 := tree.GetLayout(11)
	if layout11.Position.X != 50 {
		t.Errorf("child 11 X = %v, want 50", layout11.Position.X)
	}

	layout12 := tree.GetLayout(12)
	if layout12.Position.X != 100 { // 50 + 50
		t.Errorf("child 12 X = %v, want 100", layout12.Position.X)
	}
}

func TestGridTrack_Constructors(t *testing.T) {
	auto := AutoTrack()
	if auto.Sizing != GridTrackAuto {
		t.Errorf("AutoTrack().Sizing = %v, want GridTrackAuto", auto.Sizing)
	}

	fixed := FixedTrack(100)
	if fixed.Sizing != GridTrackFixed || fixed.Value != 100 {
		t.Errorf("FixedTrack(100) = %+v, want {GridTrackFixed, 100}", fixed)
	}

	frac := FractionTrack(2)
	if frac.Sizing != GridTrackFraction || frac.Value != 2 {
		t.Errorf("FractionTrack(2) = %+v, want {GridTrackFraction, 2}", frac)
	}
}

func TestSimpleGrid(t *testing.T) {
	grid := SimpleGrid(3, 10)

	if len(grid.Columns) != 3 {
		t.Errorf("SimpleGrid(3, 10) column count = %d, want 3", len(grid.Columns))
	}

	for i, col := range grid.Columns {
		if col.Sizing != GridTrackFraction || col.Value != 1 {
			t.Errorf("column %d = %+v, want {GridTrackFraction, 1}", i, col)
		}
	}

	if grid.ColumnGap != 10 || grid.RowGap != 10 {
		t.Errorf("gaps = %v, %v, want 10, 10", grid.ColumnGap, grid.RowGap)
	}
}

func TestGridLayout_DefaultColumn(t *testing.T) {
	// Grid with no columns should default to single fractional column
	grid := &GridLayout{}
	tree := newTestTree()

	tree.AddChild(1, 10)
	tree.SetPreferredSize(10, geometry.Size{Width: 100, Height: 50})

	result := grid.Compute(tree, 1, geometry.Size{Width: 200, Height: 100})

	// Should use full width
	if result.Size.Width != 200 {
		t.Errorf("Result.Size.Width = %v, want 200", result.Size.Width)
	}
}

func TestGridLayout_Registered(t *testing.T) {
	if !Has("grid") {
		t.Error("grid layout should be registered")
	}

	algo, ok := Get("grid")
	if !ok {
		t.Fatal("Get('grid') should return true")
	}

	if algo.Name() != "grid" {
		t.Errorf("algorithm Name() = %q, want %q", algo.Name(), "grid")
	}
}
