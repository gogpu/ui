package datatable

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// Painter draws table-specific visual elements.
// Each design system (Material 3, Fluent, Cupertino) provides its own
// Painter implementation to render the table in its visual style.
//
// If no Painter is set, the data table uses [DefaultPainter].
type Painter interface {
	// PaintHeader draws the table header background.
	PaintHeader(canvas widget.Canvas, bounds geometry.Rect, state HeaderPaintState)

	// PaintHeaderCell draws a single header cell (column title + sort indicator).
	PaintHeaderCell(canvas widget.Canvas, bounds geometry.Rect, state HeaderCellPaintState)

	// PaintRow draws the background for a data row.
	PaintRow(canvas widget.Canvas, state RowPaintState)

	// PaintCell draws a single data cell.
	PaintCell(canvas widget.Canvas, state CellPaintState)

	// PaintEmptyState draws the empty state when RowCount is 0.
	PaintEmptyState(canvas widget.Canvas, bounds geometry.Rect)
}

// HeaderPaintState provides context for header background painting.
type HeaderPaintState struct {
	// Disabled is true if the table is disabled.
	Disabled bool

	// ColorScheme provides theme-derived colors.
	ColorScheme TableColorScheme
}

// HeaderCellPaintState provides context for a single header cell.
type HeaderCellPaintState struct {
	// Title is the column display text.
	Title string

	// Align is the column text alignment.
	Align widget.TextAlign

	// Sortable is true if this column supports sorting.
	Sortable bool

	// SortDir is the current sort direction for this column.
	SortDir SortDirection

	// Hovered is true if the mouse is over this header cell.
	Hovered bool

	// Disabled is true if the table is disabled.
	Disabled bool

	// ColorScheme provides theme-derived colors.
	ColorScheme TableColorScheme
}

// RowPaintState provides context for row background painting.
type RowPaintState struct {
	// Bounds is the row's bounding rectangle.
	Bounds geometry.Rect

	// RowIndex is the zero-based row index in the data source.
	RowIndex int

	// Selected is true if this row is selected.
	Selected bool

	// Focused is true if this row has focus within the table.
	Focused bool

	// Hovered is true if the mouse is over this row.
	Hovered bool

	// Disabled is true if the table is disabled.
	Disabled bool

	// ColorScheme provides theme-derived colors.
	ColorScheme TableColorScheme
}

// CellPaintState provides context for a single data cell.
type CellPaintState struct {
	// Bounds is the cell's bounding rectangle.
	Bounds geometry.Rect

	// Value is the cell's text content.
	Value string

	// Align is the column text alignment.
	Align widget.TextAlign

	// RowIndex is the zero-based row index.
	RowIndex int

	// ColIndex is the zero-based column index.
	ColIndex int

	// Selected is true if the parent row is selected.
	Selected bool

	// Disabled is true if the table is disabled.
	Disabled bool

	// ColorScheme provides theme-derived colors.
	ColorScheme TableColorScheme
}

// TableColorScheme provides theme-derived colors for table painting.
// Zero value means the painter should use its built-in defaults.
type TableColorScheme struct {
	HeaderBackground widget.Color // header row background
	HeaderText       widget.Color // header text color
	RowBackground    widget.Color // normal row background
	RowAlternate     widget.Color // alternate row background (zebra)
	SelectionColor   widget.Color // selected row background
	HoverColor       widget.Color // hovered row background
	FocusColor       widget.Color // focused row border
	CellText         widget.Color // cell text color
	Divider          widget.Color // column/row divider color
	EmptyText        widget.Color // empty state text color
}

// DefaultPainter provides a minimal fallback painter with no design system styling.
type DefaultPainter struct{}

// PaintHeader draws the header background.
func (p DefaultPainter) PaintHeader(canvas widget.Canvas, bounds geometry.Rect, hps HeaderPaintState) {
	if bounds.IsEmpty() {
		return
	}
	bg := defaultHeaderBackground
	if hps.ColorScheme != (TableColorScheme{}) {
		bg = hps.ColorScheme.HeaderBackground
	}
	canvas.DrawRect(bounds, bg)
}

// PaintHeaderCell draws a column header with title and sort indicator.
func (p DefaultPainter) PaintHeaderCell(canvas widget.Canvas, bounds geometry.Rect, hcs HeaderCellPaintState) {
	if bounds.IsEmpty() {
		return
	}

	textColor := defaultHeaderTextColor
	if hcs.ColorScheme != (TableColorScheme{}) {
		textColor = hcs.ColorScheme.HeaderText
	}

	// Highlight on hover for sortable columns.
	if hcs.Hovered && hcs.Sortable && !hcs.Disabled {
		hoverColor := defaultHeaderHoverColor
		if hcs.ColorScheme != (TableColorScheme{}) {
			hoverColor = hcs.ColorScheme.HoverColor
		}
		canvas.DrawRect(bounds, hoverColor)
	}

	// Build display text with sort indicator.
	displayText := hcs.Title
	if indicator := hcs.SortDir.Indicator(); indicator != "" {
		displayText = hcs.Title + " " + indicator
	}

	// Inset for text padding.
	textBounds := geometry.NewRect(
		bounds.Min.X+cellPaddingH,
		bounds.Min.Y,
		bounds.Width()-cellPaddingH*2,
		bounds.Height(),
	)
	canvas.DrawText(displayText, textBounds, headerFontSize, textColor, true, hcs.Align)
}

// PaintRow draws the row background with zebra striping and selection/hover highlights.
func (p DefaultPainter) PaintRow(canvas widget.Canvas, rps RowPaintState) {
	if rps.Bounds.IsEmpty() {
		return
	}
	hasScheme := rps.ColorScheme != (TableColorScheme{})

	// Zebra striping.
	if rps.RowIndex%2 == 1 {
		bg := defaultRowAlternate
		if hasScheme {
			bg = rps.ColorScheme.RowAlternate
		}
		canvas.DrawRect(rps.Bounds, bg)
	}

	// Selection.
	if rps.Selected {
		color := defaultSelectionColor
		if hasScheme {
			color = rps.ColorScheme.SelectionColor
		}
		canvas.DrawRect(rps.Bounds, color)
	}

	// Hover (only if not selected, to avoid double-tinting).
	if rps.Hovered && !rps.Selected && !rps.Disabled {
		color := defaultHoverColor
		if hasScheme {
			color = rps.ColorScheme.HoverColor
		}
		canvas.DrawRect(rps.Bounds, color)
	}

	// Focus ring.
	if rps.Focused && !rps.Disabled {
		focusColor := defaultFocusBorderColor
		if hasScheme {
			focusColor = rps.ColorScheme.FocusColor
		}
		canvas.StrokeRect(rps.Bounds, focusColor, focusBorderWidth)
	}
}

// PaintCell draws a single data cell with text.
func (p DefaultPainter) PaintCell(canvas widget.Canvas, cps CellPaintState) {
	if cps.Bounds.IsEmpty() {
		return
	}
	textColor := defaultCellTextColor
	if cps.ColorScheme != (TableColorScheme{}) {
		textColor = cps.ColorScheme.CellText
	}

	textBounds := geometry.NewRect(
		cps.Bounds.Min.X+cellPaddingH,
		cps.Bounds.Min.Y,
		cps.Bounds.Width()-cellPaddingH*2,
		cps.Bounds.Height(),
	)
	canvas.DrawText(cps.Value, textBounds, cellFontSize, textColor, false, cps.Align)
}

// PaintEmptyState draws a centered "No data" message.
func (p DefaultPainter) PaintEmptyState(canvas widget.Canvas, bounds geometry.Rect) {
	if bounds.IsEmpty() {
		return
	}
	canvas.DrawText(emptyStateText, bounds, emptyStateFontSize, defaultEmptyTextColor, false, widget.TextAlignCenter)
}

// Painting constants.
const (
	cellPaddingH       float32 = 8
	headerFontSize     float32 = 13
	cellFontSize       float32 = 13
	focusBorderWidth   float32 = 2
	emptyStateFontSize float32 = 14
	emptyStateText             = "No data"
)

// Default colors.
var (
	defaultHeaderBackground = widget.RGBA(0.95, 0.95, 0.97, 1.0)
	defaultHeaderTextColor  = widget.RGBA(0.2, 0.2, 0.2, 1.0)
	defaultHeaderHoverColor = widget.RGBA(0.0, 0.0, 0.0, 0.06)
	defaultCellTextColor    = widget.RGBA(0.13, 0.13, 0.13, 1.0)
	defaultRowAlternate     = widget.RGBA(0.0, 0.0, 0.0, 0.02)
	defaultSelectionColor   = widget.RGBA(0.23, 0.51, 0.96, 0.12)
	defaultHoverColor       = widget.RGBA(0.0, 0.0, 0.0, 0.04)
	defaultFocusBorderColor = widget.Hex(0x6750A4).WithAlpha(0.7)
	defaultEmptyTextColor   = widget.RGBA(0.5, 0.5, 0.5, 1.0)
)
