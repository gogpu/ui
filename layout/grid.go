package layout

import (
	"github.com/gogpu/ui/geometry"
)

// gridLayoutName is the constant name for grid layout.
const gridLayoutName = "grid"

// GridTrackSizing specifies how a grid track (row or column) is sized.
type GridTrackSizing int

const (
	// GridTrackAuto sizes the track to fit its content.
	GridTrackAuto GridTrackSizing = iota
	// GridTrackFixed uses a fixed pixel size.
	GridTrackFixed
	// GridTrackFraction uses a fraction of available space (like CSS fr unit).
	GridTrackFraction
)

// GridTrack defines a grid row or column.
type GridTrack struct {
	// Sizing specifies how the track is sized.
	Sizing GridTrackSizing

	// Value is the size value. Meaning depends on Sizing:
	//   - GridTrackAuto: ignored
	//   - GridTrackFixed: pixel size
	//   - GridTrackFraction: fraction value (like 1fr, 2fr)
	Value float32
}

// AutoTrack creates an auto-sized track.
func AutoTrack() GridTrack {
	return GridTrack{Sizing: GridTrackAuto}
}

// FixedTrack creates a fixed-size track.
func FixedTrack(size float32) GridTrack {
	return GridTrack{Sizing: GridTrackFixed, Value: size}
}

// FractionTrack creates a fractional track.
func FractionTrack(fr float32) GridTrack {
	return GridTrack{Sizing: GridTrackFraction, Value: fr}
}

// GridLayout implements CSS Grid-style layout.
//
// GridLayout arranges children in a grid of rows and columns,
// with support for fixed, auto, and fractional track sizes.
type GridLayout struct {
	// Columns defines the column tracks.
	Columns []GridTrack

	// Rows defines the row tracks.
	Rows []GridTrack

	// ColumnGap is the space between columns.
	ColumnGap float32

	// RowGap is the space between rows.
	RowGap float32
}

// Name returns "grid".
func (g *GridLayout) Name() string {
	return gridLayoutName
}

// Compute performs grid layout on the tree starting at root.
func (g *GridLayout) Compute(tree LayoutTree, root NodeID, available geometry.Size) Result {
	childCount := tree.ChildCount(root)
	if childCount == 0 {
		return Result{}
	}

	columns := g.getColumns()
	cells, maxRow := g.buildCells(tree, root, childCount, len(columns))
	rows := g.buildRows(maxRow)

	columnSizes := g.calculateTrackSizes(columns, available.Width, g.ColumnGap)
	rowSizes := g.initRowSizes(rows)

	g.measureCells(tree, cells, rows, columnSizes, rowSizes)
	g.resolveFractionalRows(rows, rowSizes, available.Height)

	g.positionCells(tree, cells, columnSizes, rowSizes)

	return Result{
		Size: geometry.Size{
			Width:  g.totalSize(columnSizes, g.ColumnGap),
			Height: g.totalSize(rowSizes, g.RowGap),
		},
	}
}

// getColumns returns the column tracks, defaulting to a single fractional column.
func (g *GridLayout) getColumns() []GridTrack {
	if len(g.Columns) == 0 {
		return []GridTrack{FractionTrack(1)}
	}
	return g.Columns
}

// buildCells creates grid cells for each child and determines max row count.
func (g *GridLayout) buildCells(tree LayoutTree, root NodeID, childCount, numColumns int) ([]gridCell, int) {
	cells := make([]gridCell, childCount)
	maxRow := 0
	for i := 0; i < childCount; i++ {
		row := i / numColumns
		col := i % numColumns
		cells[i] = gridCell{
			id:      tree.ChildAt(root, i),
			row:     row,
			col:     col,
			rowSpan: 1,
			colSpan: 1,
		}
		if row+1 > maxRow {
			maxRow = row + 1
		}
	}
	return cells, maxRow
}

// buildRows ensures we have enough row definitions for all cells.
func (g *GridLayout) buildRows(maxRow int) []GridTrack {
	rows := make([]GridTrack, 0, maxRow)
	rows = append(rows, g.Rows...)
	for len(rows) < maxRow {
		rows = append(rows, AutoTrack())
	}
	return rows
}

// initRowSizes initializes row sizes based on track definitions.
func (g *GridLayout) initRowSizes(rows []GridTrack) []float32 {
	sizes := make([]float32, len(rows))
	for i, row := range rows {
		if row.Sizing == GridTrackFixed {
			sizes[i] = row.Value
		}
	}
	return sizes
}

// measureCells measures each cell and updates auto row sizes.
func (g *GridLayout) measureCells(tree LayoutTree, cells []gridCell, rows []GridTrack, colSizes, rowSizes []float32) {
	for i := range cells {
		cell := &cells[i]
		cellWidth := g.spanSize(colSizes, cell.col, cell.colSpan, g.ColumnGap)
		constraints := geometry.BoxConstraints(cellWidth, cellWidth, 0, geometry.Infinity)
		cell.size = tree.Measure(cell.id, constraints)

		if cell.row < len(rows) && rows[cell.row].Sizing == GridTrackAuto {
			if cell.size.Height > rowSizes[cell.row] {
				rowSizes[cell.row] = cell.size.Height
			}
		}
	}
}

// resolveFractionalRows calculates sizes for fractional rows.
func (g *GridLayout) resolveFractionalRows(rows []GridTrack, rowSizes []float32, availableHeight float32) {
	var fixedTotal, fractionTotal float32
	for i, row := range rows {
		switch row.Sizing {
		case GridTrackFixed, GridTrackAuto:
			fixedTotal += rowSizes[i]
		case GridTrackFraction:
			fractionTotal += row.Value
		}
	}

	numGaps := len(rows) - 1
	if numGaps < 0 {
		numGaps = 0
	}
	remainingHeight := availableHeight - g.RowGap*float32(numGaps) - fixedTotal
	if remainingHeight < 0 {
		remainingHeight = 0
	}

	if fractionTotal > 0 {
		for i, row := range rows {
			if row.Sizing == GridTrackFraction {
				rowSizes[i] = remainingHeight * (row.Value / fractionTotal)
			}
		}
	}
}

// positionCells sets the layout for each cell based on calculated sizes.
func (g *GridLayout) positionCells(tree LayoutTree, cells []gridCell, colSizes, rowSizes []float32) {
	colOffsets := g.calculateOffsets(colSizes, g.ColumnGap)
	rowOffsets := g.calculateOffsets(rowSizes, g.RowGap)

	for i := range cells {
		cell := &cells[i]
		var x, y float32
		if cell.col < len(colOffsets) {
			x = colOffsets[cell.col]
		}
		if cell.row < len(rowOffsets) {
			y = rowOffsets[cell.row]
		}
		tree.SetLayout(cell.id, NodeLayout{
			Position: geometry.Point{X: x, Y: y},
			Size:     cell.size,
		})
	}
}

// gridCell stores information about a child in the grid.
type gridCell struct {
	id      NodeID
	row     int
	col     int
	rowSpan int
	colSpan int
	size    geometry.Size
}

func (g *GridLayout) calculateTrackSizes(tracks []GridTrack, available float32, gap float32) []float32 {
	sizes := make([]float32, len(tracks))

	// Handle unbounded available space
	if available >= geometry.Infinity {
		available = 1000 // fallback
	}

	// Calculate total gaps
	numGaps := len(tracks) - 1
	if numGaps < 0 {
		numGaps = 0
	}
	totalGaps := gap * float32(numGaps)

	// First pass: calculate fixed and auto sizes
	var fixedTotal float32
	var fractionTotal float32

	for i, track := range tracks {
		switch track.Sizing {
		case GridTrackFixed:
			sizes[i] = track.Value
			fixedTotal += track.Value
		case GridTrackAuto:
			// Start with minimum size; will be updated during cell measurement
			sizes[i] = 0
		case GridTrackFraction:
			fractionTotal += track.Value
		}
	}

	// Second pass: distribute remaining space to fractional tracks
	remainingSpace := available - totalGaps - fixedTotal
	if remainingSpace < 0 {
		remainingSpace = 0
	}

	if fractionTotal > 0 {
		for i, track := range tracks {
			if track.Sizing == GridTrackFraction {
				sizes[i] = remainingSpace * (track.Value / fractionTotal)
			}
		}
	}

	return sizes
}

func (g *GridLayout) spanSize(trackSizes []float32, start, span int, gap float32) float32 {
	var total float32
	for i := start; i < start+span && i < len(trackSizes); i++ {
		total += trackSizes[i]
		if i > start {
			total += gap
		}
	}
	return total
}

func (g *GridLayout) calculateOffsets(sizes []float32, gap float32) []float32 {
	offsets := make([]float32, len(sizes))
	var offset float32
	for i := range sizes {
		offsets[i] = offset
		offset += sizes[i] + gap
	}
	return offsets
}

func (g *GridLayout) totalSize(sizes []float32, gap float32) float32 {
	var total float32
	for i, size := range sizes {
		total += size
		if i < len(sizes)-1 {
			total += gap
		}
	}
	return total
}

// SimpleGrid creates a grid layout with equal fractional columns.
func SimpleGrid(numColumns int, gap float32) *GridLayout {
	columns := make([]GridTrack, numColumns)
	for i := range columns {
		columns[i] = FractionTrack(1)
	}
	return &GridLayout{
		Columns:   columns,
		ColumnGap: gap,
		RowGap:    gap,
	}
}

func init() {
	Register(&GridLayout{
		Columns:   []GridTrack{FractionTrack(1)},
		ColumnGap: 0,
		RowGap:    0,
	})
}
