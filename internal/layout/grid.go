package layout

import (
	"github.com/gogpu/ui/geometry"
)

// TrackSizing specifies how a grid track (row or column) is sized.
type TrackSizing int

const (
	// TrackAuto sizes the track to fit its content.
	TrackAuto TrackSizing = iota
	// TrackFixed uses a fixed pixel size.
	TrackFixed
	// TrackFraction uses a fraction of available space (like CSS fr unit).
	TrackFraction
)

// Track defines a grid row or column.
type Track struct {
	// Sizing specifies how the track is sized.
	Sizing TrackSizing

	// Value is the size value. Meaning depends on Sizing:
	//   - TrackAuto: ignored
	//   - TrackFixed: pixel size
	//   - TrackFraction: fraction value (like 1fr, 2fr)
	Value float32

	// computed values after layout
	offset float32
	size   float32
}

// AutoTrack creates an auto-sized track.
func AutoTrack() Track {
	return Track{Sizing: TrackAuto}
}

// FixedTrack creates a fixed-size track.
func FixedTrack(size float32) Track {
	return Track{Sizing: TrackFixed, Value: size}
}

// FractionTrack creates a fractional track.
func FractionTrack(fr float32) Track {
	return Track{Sizing: TrackFraction, Value: fr}
}

// GridCell represents a child in the grid.
type GridCell struct {
	// Element is the layoutable child.
	Element Layoutable

	// Row is the row index (0-based).
	Row int

	// Column is the column index (0-based).
	Column int

	// RowSpan is the number of rows to span (default 1).
	RowSpan int

	// ColSpan is the number of columns to span (default 1).
	ColSpan int

	// computed values after layout
	position geometry.Point
	size     geometry.Size
}

// GridContainer implements grid layout.
//
// GridContainer arranges children in a grid of rows and columns.
// Each track (row/column) can be auto-sized, fixed, or fractional.
type GridContainer struct {
	id uint64

	// Columns defines the column tracks.
	Columns []Track

	// Rows defines the row tracks.
	Rows []Track

	// ColumnGap is the space between columns.
	ColumnGap float32

	// RowGap is the space between rows.
	RowGap float32

	// Cells are the grid children.
	Cells []GridCell

	// computed size after layout
	size geometry.Size
}

// NewGridContainer creates a new grid container.
func NewGridContainer(columns, rows []Track) *GridContainer {
	return &GridContainer{
		Columns: columns,
		Rows:    rows,
		Cells:   make([]GridCell, 0, 8),
	}
}

// NewSimpleGrid creates a grid with equal fractional columns and auto rows.
func NewSimpleGrid(numColumns int) *GridContainer {
	columns := make([]Track, numColumns)
	for i := range columns {
		columns[i] = FractionTrack(1)
	}
	return &GridContainer{
		Columns: columns,
		Rows:    nil, // Auto-create rows as needed
		Cells:   make([]GridCell, 0, 8),
	}
}

// SetID sets the unique identifier for caching.
func (g *GridContainer) SetID(id uint64) {
	g.id = id
}

// ID returns the unique identifier.
func (g *GridContainer) ID() uint64 {
	return g.id
}

// SetGap sets both column and row gaps.
func (g *GridContainer) SetGap(gap float32) {
	g.ColumnGap = gap
	g.RowGap = gap
}

// AddCell adds a child at the specified grid position.
func (g *GridContainer) AddCell(element Layoutable, row, column int) {
	g.Cells = append(g.Cells, GridCell{
		Element: element,
		Row:     row,
		Column:  column,
		RowSpan: 1,
		ColSpan: 1,
	})
}

// AddCellWithSpan adds a child that spans multiple cells.
func (g *GridContainer) AddCellWithSpan(element Layoutable, row, column, rowSpan, colSpan int) {
	if rowSpan < 1 {
		rowSpan = 1
	}
	if colSpan < 1 {
		colSpan = 1
	}
	g.Cells = append(g.Cells, GridCell{
		Element: element,
		Row:     row,
		Column:  column,
		RowSpan: rowSpan,
		ColSpan: colSpan,
	})
}

// AddChildAutoFlow adds a child, automatically placing it in the next available cell.
// Cells are filled left-to-right, top-to-bottom.
func (g *GridContainer) AddChildAutoFlow(element Layoutable) {
	numCols := len(g.Columns)
	if numCols == 0 {
		numCols = 1
	}

	// Find next available position
	cellIndex := len(g.Cells)
	row := cellIndex / numCols
	column := cellIndex % numCols

	g.AddCell(element, row, column)
}

// Clear removes all cells.
func (g *GridContainer) Clear() {
	g.Cells = g.Cells[:0]
}

// Children returns child layoutables.
func (g *GridContainer) Children() []Layoutable {
	children := make([]Layoutable, len(g.Cells))
	for i, cell := range g.Cells {
		children[i] = cell.Element
	}
	return children
}

// Layout performs grid layout.
func (g *GridContainer) Layout(constraints geometry.Constraints) geometry.Size {
	if len(g.Cells) == 0 {
		g.size = constraints.Smallest()
		return g.size
	}

	if len(g.Columns) == 0 {
		g.Columns = []Track{FractionTrack(1)}
	}

	// Determine number of rows needed
	maxRow := 0
	for _, cell := range g.Cells {
		endRow := cell.Row + cell.RowSpan
		if endRow > maxRow {
			maxRow = endRow
		}
	}

	// Ensure we have enough row definitions
	for len(g.Rows) < maxRow {
		g.Rows = append(g.Rows, AutoTrack())
	}

	// Phase 1: Calculate column sizes
	columnSizes := g.calculateTrackSizes(g.Columns, constraints.MaxWidth, g.ColumnGap, true)

	// Phase 2: Calculate row sizes
	rowSizes := g.calculateTrackSizes(g.Rows, constraints.MaxHeight, g.RowGap, false)

	// Phase 3: Layout cells and refine auto row sizes
	for i := range g.Cells {
		cell := &g.Cells[i]
		if cell.Element == nil {
			continue
		}

		// Calculate cell bounds
		cellWidth := g.spanSize(columnSizes, cell.Column, cell.ColSpan, g.ColumnGap)
		cellHeight := g.spanSize(rowSizes, cell.Row, cell.RowSpan, g.RowGap)

		// Layout cell with constraints
		cellConstraints := geometry.Tight(geometry.Size{Width: cellWidth, Height: cellHeight})
		cell.size = cell.Element.Layout(cellConstraints)

		// For auto rows, update row size if content is larger
		if cell.Row < len(g.Rows) && g.Rows[cell.Row].Sizing == TrackAuto {
			if cell.size.Height > rowSizes[cell.Row] {
				rowSizes[cell.Row] = cell.size.Height
			}
		}
	}

	// Phase 4: Calculate offsets and position cells
	columnOffsets := g.calculateOffsets(columnSizes, g.ColumnGap)
	rowOffsets := g.calculateOffsets(rowSizes, g.RowGap)

	for i := range g.Cells {
		cell := &g.Cells[i]
		if cell.Element == nil {
			continue
		}

		x := float32(0)
		if cell.Column < len(columnOffsets) {
			x = columnOffsets[cell.Column]
		}

		y := float32(0)
		if cell.Row < len(rowOffsets) {
			y = rowOffsets[cell.Row]
		}

		cell.position = geometry.Point{X: x, Y: y}
	}

	// Calculate total size
	totalWidth := g.totalSize(columnSizes, g.ColumnGap)
	totalHeight := g.totalSize(rowSizes, g.RowGap)

	g.size = constraints.Constrain(geometry.Size{Width: totalWidth, Height: totalHeight})
	return g.size
}

// calculateTrackSizes calculates sizes for tracks based on available space.
func (g *GridContainer) calculateTrackSizes(tracks []Track, available float32, gap float32, isColumn bool) []float32 {
	sizes := make([]float32, len(tracks))

	// Handle unbounded available space
	if available >= geometry.Infinity {
		available = 1000 // fallback for unbounded
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
		case TrackFixed:
			sizes[i] = track.Value
			fixedTotal += track.Value
		case TrackAuto:
			// For auto, we need to measure content
			// Start with a minimum size, will be adjusted during cell layout
			minSize := g.measureAutoTrack(i, isColumn)
			sizes[i] = minSize
			fixedTotal += minSize
		case TrackFraction:
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
			if track.Sizing == TrackFraction {
				sizes[i] = remainingSpace * (track.Value / fractionTotal)
			}
		}
	}

	return sizes
}

// measureAutoTrack measures the minimum size needed for an auto track.
func (g *GridContainer) measureAutoTrack(trackIndex int, isColumn bool) float32 {
	var maxSize float32

	for _, cell := range g.Cells {
		if cell.Element == nil {
			continue
		}

		var matchesTrack bool
		if isColumn {
			matchesTrack = cell.Column <= trackIndex && trackIndex < cell.Column+cell.ColSpan
		} else {
			matchesTrack = cell.Row <= trackIndex && trackIndex < cell.Row+cell.RowSpan
		}

		if !matchesTrack {
			continue
		}

		// Measure element with loose constraints
		childSize := cell.Element.Layout(geometry.Expand())

		var size float32
		if isColumn {
			// For spanning cells, distribute size across tracks
			size = childSize.Width / float32(cell.ColSpan)
		} else {
			size = childSize.Height / float32(cell.RowSpan)
		}

		if size > maxSize {
			maxSize = size
		}
	}

	return maxSize
}

// spanSize calculates the total size of a span of tracks.
func (g *GridContainer) spanSize(trackSizes []float32, start, span int, gap float32) float32 {
	var total float32
	for i := start; i < start+span && i < len(trackSizes); i++ {
		total += trackSizes[i]
		if i > start {
			total += gap
		}
	}
	return total
}

// calculateOffsets calculates the offset for each track.
func (g *GridContainer) calculateOffsets(sizes []float32, gap float32) []float32 {
	offsets := make([]float32, len(sizes))
	var offset float32
	for i := range sizes {
		offsets[i] = offset
		offset += sizes[i] + gap
	}
	return offsets
}

// totalSize calculates the total size of all tracks including gaps.
func (g *GridContainer) totalSize(sizes []float32, gap float32) float32 {
	var total float32
	for i, size := range sizes {
		total += size
		if i < len(sizes)-1 {
			total += gap
		}
	}
	return total
}

// Size returns the computed size after layout.
func (g *GridContainer) Size() geometry.Size {
	return g.size
}

// CellPosition returns the position of a cell after layout.
func (g *GridContainer) CellPosition(index int) geometry.Point {
	if index < 0 || index >= len(g.Cells) {
		return geometry.Point{}
	}
	return g.Cells[index].position
}

// CellSize returns the computed size of a cell after layout.
func (g *GridContainer) CellSize(index int) geometry.Size {
	if index < 0 || index >= len(g.Cells) {
		return geometry.Size{}
	}
	return g.Cells[index].size
}

// CellBounds returns the bounds of a cell after layout.
func (g *GridContainer) CellBounds(index int) geometry.Rect {
	pos := g.CellPosition(index)
	size := g.CellSize(index)
	return geometry.FromPointSize(pos, size)
}

// ColumnCount returns the number of columns.
func (g *GridContainer) ColumnCount() int {
	return len(g.Columns)
}

// RowCount returns the number of rows.
func (g *GridContainer) RowCount() int {
	return len(g.Rows)
}

// ColumnOffset returns the X offset for a column.
func (g *GridContainer) ColumnOffset(column int) float32 {
	if column < 0 || column >= len(g.Columns) {
		return 0
	}
	return g.Columns[column].offset
}

// RowOffset returns the Y offset for a row.
func (g *GridContainer) RowOffset(row int) float32 {
	if row < 0 || row >= len(g.Rows) {
		return 0
	}
	return g.Rows[row].offset
}

// ColumnSize returns the width of a column after layout.
func (g *GridContainer) ColumnSize(column int) float32 {
	if column < 0 || column >= len(g.Columns) {
		return 0
	}
	return g.Columns[column].size
}

// RowSize returns the height of a row after layout.
func (g *GridContainer) RowSize(row int) float32 {
	if row < 0 || row >= len(g.Rows) {
		return 0
	}
	return g.Rows[row].size
}
