package layout

import (
	"testing"

	"github.com/gogpu/ui/geometry"
)

func TestAutoTrack(t *testing.T) {
	track := AutoTrack()

	if track.Sizing != TrackAuto {
		t.Errorf("Sizing = %v, want TrackAuto", track.Sizing)
	}
}

func TestFixedTrack(t *testing.T) {
	track := FixedTrack(100)

	if track.Sizing != TrackFixed {
		t.Errorf("Sizing = %v, want TrackFixed", track.Sizing)
	}
	if track.Value != 100 {
		t.Errorf("Value = %v, want 100", track.Value)
	}
}

func TestFractionTrack(t *testing.T) {
	track := FractionTrack(2)

	if track.Sizing != TrackFraction {
		t.Errorf("Sizing = %v, want TrackFraction", track.Sizing)
	}
	if track.Value != 2 {
		t.Errorf("Value = %v, want 2", track.Value)
	}
}

func TestNewGridContainer(t *testing.T) {
	columns := []Track{FractionTrack(1), FractionTrack(2)}
	rows := []Track{AutoTrack(), FixedTrack(50)}
	grid := NewGridContainer(columns, rows)

	if len(grid.Columns) != 2 {
		t.Errorf("Columns length = %d, want 2", len(grid.Columns))
	}
	if len(grid.Rows) != 2 {
		t.Errorf("Rows length = %d, want 2", len(grid.Rows))
	}
	if len(grid.Cells) != 0 {
		t.Errorf("Cells length = %d, want 0", len(grid.Cells))
	}
}

func TestNewSimpleGrid(t *testing.T) {
	grid := NewSimpleGrid(3)

	if len(grid.Columns) != 3 {
		t.Errorf("Columns length = %d, want 3", len(grid.Columns))
	}
	for i, col := range grid.Columns {
		if col.Sizing != TrackFraction {
			t.Errorf("Column %d Sizing = %v, want TrackFraction", i, col.Sizing)
		}
		if col.Value != 1 {
			t.Errorf("Column %d Value = %v, want 1", i, col.Value)
		}
	}
}

func TestGridContainer_ID(t *testing.T) {
	grid := NewSimpleGrid(2)

	if grid.ID() != 0 {
		t.Errorf("ID() = %d, want 0", grid.ID())
	}

	grid.SetID(42)

	if grid.ID() != 42 {
		t.Errorf("ID() = %d, want 42", grid.ID())
	}
}

func TestGridContainer_SetGap(t *testing.T) {
	grid := NewSimpleGrid(2)
	grid.SetGap(10)

	if grid.ColumnGap != 10 {
		t.Errorf("ColumnGap = %v, want 10", grid.ColumnGap)
	}
	if grid.RowGap != 10 {
		t.Errorf("RowGap = %v, want 10", grid.RowGap)
	}
}

func TestGridContainer_AddCell(t *testing.T) {
	grid := NewSimpleGrid(2)
	child := &mockLayoutable{preferredSize: geometry.Sz(50, 30)}
	grid.AddCell(child, 1, 0)

	if len(grid.Cells) != 1 {
		t.Fatalf("Cells length = %d, want 1", len(grid.Cells))
	}

	cell := grid.Cells[0]
	if cell.Element != child {
		t.Error("Cell element mismatch")
	}
	if cell.Row != 1 {
		t.Errorf("Row = %d, want 1", cell.Row)
	}
	if cell.Column != 0 {
		t.Errorf("Column = %d, want 0", cell.Column)
	}
	if cell.RowSpan != 1 {
		t.Errorf("RowSpan = %d, want 1", cell.RowSpan)
	}
	if cell.ColSpan != 1 {
		t.Errorf("ColSpan = %d, want 1", cell.ColSpan)
	}
}

func TestGridContainer_AddCellWithSpan(t *testing.T) {
	grid := NewSimpleGrid(3)
	child := &mockLayoutable{preferredSize: geometry.Sz(100, 60)}
	grid.AddCellWithSpan(child, 0, 0, 2, 2)

	if len(grid.Cells) != 1 {
		t.Fatalf("Cells length = %d, want 1", len(grid.Cells))
	}

	cell := grid.Cells[0]
	if cell.RowSpan != 2 {
		t.Errorf("RowSpan = %d, want 2", cell.RowSpan)
	}
	if cell.ColSpan != 2 {
		t.Errorf("ColSpan = %d, want 2", cell.ColSpan)
	}
}

func TestGridContainer_AddCellWithSpan_MinimumValues(t *testing.T) {
	grid := NewSimpleGrid(2)
	child := &mockLayoutable{preferredSize: geometry.Sz(50, 30)}
	grid.AddCellWithSpan(child, 0, 0, 0, -1) // Invalid spans

	cell := grid.Cells[0]
	// Should be clamped to minimum of 1
	if cell.RowSpan != 1 {
		t.Errorf("RowSpan = %d, want 1", cell.RowSpan)
	}
	if cell.ColSpan != 1 {
		t.Errorf("ColSpan = %d, want 1", cell.ColSpan)
	}
}

func TestGridContainer_AddChildAutoFlow(t *testing.T) {
	grid := NewSimpleGrid(3)
	grid.AddChildAutoFlow(&mockLayoutable{preferredSize: geometry.Sz(50, 30)})
	grid.AddChildAutoFlow(&mockLayoutable{preferredSize: geometry.Sz(50, 30)})
	grid.AddChildAutoFlow(&mockLayoutable{preferredSize: geometry.Sz(50, 30)})
	grid.AddChildAutoFlow(&mockLayoutable{preferredSize: geometry.Sz(50, 30)})

	if len(grid.Cells) != 4 {
		t.Fatalf("Cells length = %d, want 4", len(grid.Cells))
	}

	// Check positions: row-major order
	expected := []struct{ row, col int }{
		{0, 0}, {0, 1}, {0, 2}, {1, 0},
	}
	for i, exp := range expected {
		if grid.Cells[i].Row != exp.row || grid.Cells[i].Column != exp.col {
			t.Errorf("Cell %d position = (%d, %d), want (%d, %d)",
				i, grid.Cells[i].Row, grid.Cells[i].Column, exp.row, exp.col)
		}
	}
}

func TestGridContainer_Clear(t *testing.T) {
	grid := NewSimpleGrid(2)
	grid.AddCell(&mockLayoutable{preferredSize: geometry.Sz(50, 30)}, 0, 0)

	grid.Clear()

	if len(grid.Cells) != 0 {
		t.Errorf("Cells length = %d, want 0", len(grid.Cells))
	}
}

func TestGridContainer_Children(t *testing.T) {
	grid := NewSimpleGrid(2)
	child1 := &mockLayoutable{preferredSize: geometry.Sz(50, 30)}
	child2 := &mockLayoutable{preferredSize: geometry.Sz(50, 30)}
	grid.AddCell(child1, 0, 0)
	grid.AddCell(child2, 0, 1)

	children := grid.Children()

	if len(children) != 2 {
		t.Fatalf("Children length = %d, want 2", len(children))
	}
	if children[0] != child1 || children[1] != child2 {
		t.Error("Children mismatch")
	}
}

func TestGridContainer_Layout_Empty(t *testing.T) {
	grid := NewSimpleGrid(2)

	size := grid.Layout(geometry.Loose(geometry.Sz(200, 200)))

	if size.Width != 0 || size.Height != 0 {
		t.Errorf("Size = %v, want zero", size)
	}
}

func TestGridContainer_Layout_Simple(t *testing.T) {
	grid := NewSimpleGrid(2)
	grid.AddCell(&mockLayoutable{preferredSize: geometry.Sz(50, 30)}, 0, 0)
	grid.AddCell(&mockLayoutable{preferredSize: geometry.Sz(50, 30)}, 0, 1)

	// With 200 width and 2 equal columns, each column is 100
	size := grid.Layout(geometry.Tight(geometry.Sz(200, 100)))

	if size.Width != 200 {
		t.Errorf("Width = %v, want 200", size.Width)
	}
}

func TestGridContainer_Layout_FixedColumns(t *testing.T) {
	columns := []Track{FixedTrack(100), FixedTrack(50)}
	grid := NewGridContainer(columns, nil)
	grid.AddCell(&mockLayoutable{preferredSize: geometry.Sz(100, 30)}, 0, 0)
	grid.AddCell(&mockLayoutable{preferredSize: geometry.Sz(50, 30)}, 0, 1)

	size := grid.Layout(geometry.Loose(geometry.Sz(300, 200)))

	// Fixed columns total = 100 + 50 = 150
	if size.Width != 150 {
		t.Errorf("Width = %v, want 150", size.Width)
	}
}

func TestGridContainer_Layout_MixedColumns(t *testing.T) {
	// 50px fixed + 1fr + 2fr in 350px container
	// Fixed: 50, Remaining: 300, 1fr=100, 2fr=200
	columns := []Track{FixedTrack(50), FractionTrack(1), FractionTrack(2)}
	grid := NewGridContainer(columns, nil)
	grid.AddCell(&mockLayoutable{preferredSize: geometry.Sz(50, 30)}, 0, 0)
	grid.AddCell(&mockLayoutable{preferredSize: geometry.Sz(100, 30)}, 0, 1)
	grid.AddCell(&mockLayoutable{preferredSize: geometry.Sz(200, 30)}, 0, 2)

	size := grid.Layout(geometry.Tight(geometry.Sz(350, 100)))

	if size.Width != 350 {
		t.Errorf("Width = %v, want 350", size.Width)
	}
}

func TestGridContainer_Layout_WithGap(t *testing.T) {
	grid := NewSimpleGrid(3)
	grid.SetGap(10)
	grid.AddCell(&mockLayoutable{preferredSize: geometry.Sz(50, 30)}, 0, 0)
	grid.AddCell(&mockLayoutable{preferredSize: geometry.Sz(50, 30)}, 0, 1)
	grid.AddCell(&mockLayoutable{preferredSize: geometry.Sz(50, 30)}, 0, 2)

	// With 200 width, 3 columns, and 10px gaps (2 gaps)
	// Available for columns: 200 - 20 = 180, each column = 60
	// But positions are offset by gaps
	_ = grid.Layout(geometry.Tight(geometry.Sz(200, 100)))

	pos0 := grid.CellPosition(0)
	pos1 := grid.CellPosition(1)
	pos2 := grid.CellPosition(2)

	if pos0.X != 0 {
		t.Errorf("Cell 0 X = %v, want 0", pos0.X)
	}
	// With gaps, positions are at column offsets
	if pos1.X == 0 {
		t.Error("Cell 1 X should not be 0 with gaps")
	}
	if pos2.X <= pos1.X {
		t.Errorf("Cell 2 X (%v) should be > Cell 1 X (%v)", pos2.X, pos1.X)
	}
}

func TestGridContainer_Layout_MultipleRows(t *testing.T) {
	grid := NewSimpleGrid(2)
	grid.AddCell(&mockLayoutable{preferredSize: geometry.Sz(50, 40)}, 0, 0)
	grid.AddCell(&mockLayoutable{preferredSize: geometry.Sz(50, 30)}, 0, 1)
	grid.AddCell(&mockLayoutable{preferredSize: geometry.Sz(50, 50)}, 1, 0)
	grid.AddCell(&mockLayoutable{preferredSize: geometry.Sz(50, 60)}, 1, 1)

	size := grid.Layout(geometry.Tight(geometry.Sz(200, 200)))

	// Height should accommodate both rows
	if size.Height == 0 {
		t.Error("Height should not be 0")
	}

	// Check row positions
	pos00 := grid.CellPosition(0) // row 0
	pos10 := grid.CellPosition(2) // row 1

	if pos00.Y != 0 {
		t.Errorf("Row 0 Y = %v, want 0", pos00.Y)
	}
	if pos10.Y <= pos00.Y {
		t.Errorf("Row 1 Y (%v) should be > Row 0 Y (%v)", pos10.Y, pos00.Y)
	}
}

func TestGridContainer_CellBounds(t *testing.T) {
	grid := NewSimpleGrid(2)
	grid.AddCell(&mockLayoutable{preferredSize: geometry.Sz(50, 30)}, 0, 0)

	_ = grid.Layout(geometry.Tight(geometry.Sz(200, 100)))

	bounds := grid.CellBounds(0)

	if bounds.Min.X != 0 || bounds.Min.Y != 0 {
		t.Errorf("bounds.Min = %v, want (0, 0)", bounds.Min)
	}
	// Size should match what was laid out
	if bounds.Width() == 0 || bounds.Height() == 0 {
		t.Error("bounds size should not be zero")
	}
}

func TestGridContainer_CellPosition_OutOfBounds(t *testing.T) {
	grid := NewSimpleGrid(2)

	pos := grid.CellPosition(-1)
	if !pos.IsZero() {
		t.Errorf("CellPosition(-1) = %v, want zero", pos)
	}

	pos = grid.CellPosition(100)
	if !pos.IsZero() {
		t.Errorf("CellPosition(100) = %v, want zero", pos)
	}
}

func TestGridContainer_CellSize_OutOfBounds(t *testing.T) {
	grid := NewSimpleGrid(2)

	size := grid.CellSize(-1)
	if !size.IsZero() {
		t.Errorf("CellSize(-1) = %v, want zero", size)
	}

	size = grid.CellSize(100)
	if !size.IsZero() {
		t.Errorf("CellSize(100) = %v, want zero", size)
	}
}

func TestGridContainer_ColumnRowCount(t *testing.T) {
	columns := []Track{FractionTrack(1), FractionTrack(1), FractionTrack(1)}
	rows := []Track{AutoTrack(), AutoTrack()}
	grid := NewGridContainer(columns, rows)

	if grid.ColumnCount() != 3 {
		t.Errorf("ColumnCount() = %d, want 3", grid.ColumnCount())
	}
	if grid.RowCount() != 2 {
		t.Errorf("RowCount() = %d, want 2", grid.RowCount())
	}
}

func TestGridContainer_ColumnRowOffset_OutOfBounds(t *testing.T) {
	grid := NewSimpleGrid(2)

	offset := grid.ColumnOffset(-1)
	if offset != 0 {
		t.Errorf("ColumnOffset(-1) = %v, want 0", offset)
	}

	offset = grid.ColumnOffset(100)
	if offset != 0 {
		t.Errorf("ColumnOffset(100) = %v, want 0", offset)
	}

	offset = grid.RowOffset(-1)
	if offset != 0 {
		t.Errorf("RowOffset(-1) = %v, want 0", offset)
	}

	offset = grid.RowOffset(100)
	if offset != 0 {
		t.Errorf("RowOffset(100) = %v, want 0", offset)
	}
}

func TestGridContainer_ColumnRowSize_OutOfBounds(t *testing.T) {
	grid := NewSimpleGrid(2)

	size := grid.ColumnSize(-1)
	if size != 0 {
		t.Errorf("ColumnSize(-1) = %v, want 0", size)
	}

	size = grid.ColumnSize(100)
	if size != 0 {
		t.Errorf("ColumnSize(100) = %v, want 0", size)
	}

	size = grid.RowSize(-1)
	if size != 0 {
		t.Errorf("RowSize(-1) = %v, want 0", size)
	}

	size = grid.RowSize(100)
	if size != 0 {
		t.Errorf("RowSize(100) = %v, want 0", size)
	}
}

func TestGridContainer_NilElement(t *testing.T) {
	grid := NewSimpleGrid(2)
	grid.AddCell(nil, 0, 0)
	grid.AddCell(&mockLayoutable{preferredSize: geometry.Sz(50, 30)}, 0, 1)

	// Should handle nil elements gracefully
	size := grid.Layout(geometry.Tight(geometry.Sz(200, 100)))

	if size.Width != 200 {
		t.Errorf("Width = %v, want 200", size.Width)
	}
}

func TestGridContainer_NoColumns(t *testing.T) {
	// Grid with no columns should default to 1 column
	grid := NewGridContainer(nil, nil)
	grid.AddChildAutoFlow(&mockLayoutable{preferredSize: geometry.Sz(50, 30)})

	size := grid.Layout(geometry.Tight(geometry.Sz(200, 100)))

	if len(grid.Columns) != 1 {
		t.Errorf("Columns length = %d, want 1", len(grid.Columns))
	}
	if size.Width != 200 {
		t.Errorf("Width = %v, want 200", size.Width)
	}
}

func TestGridContainer_AutoRowsCreation(t *testing.T) {
	// Grid should auto-create rows as needed
	grid := NewSimpleGrid(2)
	grid.AddCell(&mockLayoutable{preferredSize: geometry.Sz(50, 30)}, 5, 0) // Row 5

	_ = grid.Layout(geometry.Tight(geometry.Sz(200, 300)))

	if len(grid.Rows) < 6 {
		t.Errorf("Rows length = %d, want >= 6", len(grid.Rows))
	}
}

func TestGridContainer_CellSpanning(t *testing.T) {
	grid := NewSimpleGrid(3)
	grid.SetGap(0)
	// Cell spanning 2 columns
	grid.AddCellWithSpan(&mockLayoutable{preferredSize: geometry.Sz(100, 30)}, 0, 0, 1, 2)
	grid.AddCell(&mockLayoutable{preferredSize: geometry.Sz(50, 30)}, 0, 2)

	_ = grid.Layout(geometry.Tight(geometry.Sz(300, 100)))

	// First cell should be wider due to spanning
	size0 := grid.CellSize(0)
	size1 := grid.CellSize(1)

	// Cell 0 spans 2 columns, should be about twice as wide
	// With 300px and 3 equal columns, each column is 100px
	// Spanning 2 columns = ~200px
	if size0.Width < 180 {
		t.Errorf("Spanning cell width = %v, expected ~200", size0.Width)
	}
	if size1.Width > 110 {
		t.Errorf("Single cell width = %v, expected ~100", size1.Width)
	}
}
