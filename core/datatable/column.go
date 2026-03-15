package datatable

import (
	"github.com/gogpu/ui/widget"
)

// Column describes a single column in the data table.
//
// Each column has a unique Key used to identify it in callbacks (CellValue, OnSort).
// Width and MinWidth control sizing; if Width is 0 the column receives equal share
// of remaining space after fixed-width columns are allocated.
type Column struct {
	// Key is the unique identifier for this column, used in CellValue and OnSort callbacks.
	Key string

	// Title is the display text shown in the header row.
	Title string

	// Width is the fixed column width in logical pixels. 0 means auto (flex).
	Width float32

	// MinWidth is the minimum column width for auto-sized columns.
	// Ignored when Width > 0. Default: 50.
	MinWidth float32

	// Sortable enables click-to-sort on this column's header.
	Sortable bool

	// Align controls horizontal text alignment for cells in this column.
	// Default: TextAlignLeft.
	Align widget.TextAlign
}

// SortDirection represents the sort state of a column.
type SortDirection uint8

// SortDirection constants.
const (
	// SortNone means the column is not sorted.
	SortNone SortDirection = iota

	// SortAscending sorts the column in ascending order.
	SortAscending

	// SortDescending sorts the column in descending order.
	SortDescending
)

// String returns a human-readable name for the sort direction.
func (d SortDirection) String() string {
	switch d {
	case SortNone:
		return sortNoneStr
	case SortAscending:
		return sortAscStr
	case SortDescending:
		return sortDescStr
	default:
		return sortUnknownStr
	}
}

// Indicator returns the unicode arrow character for the sort direction.
// Returns an empty string for SortNone.
func (d SortDirection) Indicator() string {
	switch d {
	case SortAscending:
		return sortIndicatorAsc
	case SortDescending:
		return sortIndicatorDesc
	default:
		return ""
	}
}

// nextDirection cycles through sort directions: None -> Asc -> Desc -> None.
func (d SortDirection) nextDirection() SortDirection {
	switch d {
	case SortNone:
		return SortAscending
	case SortAscending:
		return SortDescending
	default:
		return SortNone
	}
}

// Sort direction string constants.
const (
	sortNoneStr    = "None"
	sortAscStr     = "Ascending"
	sortDescStr    = "Descending"
	sortUnknownStr = "Unknown"

	sortIndicatorAsc  = "\u25B2" // ▲
	sortIndicatorDesc = "\u25BC" // ▼
)
