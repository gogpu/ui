package gridview

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// Painter draws grid-specific visual elements.
// Each design system (Material 3, Fluent, Cupertino) provides its own
// Painter implementation to render the grid in its visual style.
//
// If no Painter is set, the grid view uses [DefaultPainter].
type Painter interface {
	// PaintCellBackground draws the background for a grid cell.
	// This is called before the cell widget's own Draw method.
	PaintCellBackground(canvas widget.Canvas, state CellPaintState)

	// PaintSelection draws the selection highlight for a selected cell.
	// This is called before the cell widget's own Draw method.
	PaintSelection(canvas widget.Canvas, state CellPaintState)

	// PaintEmptyState draws the empty state when ItemCount is 0.
	PaintEmptyState(canvas widget.Canvas, bounds geometry.Rect)
}

// CellPaintState provides context for cell background and selection painting.
type CellPaintState struct {
	// Bounds is the cell's bounding rectangle.
	Bounds geometry.Rect

	// Index is the cell's zero-based index in the data source.
	Index int

	// Row is the cell's row index (zero-based).
	Row int

	// Col is the cell's column index (zero-based).
	Col int

	// Selected is true if the cell is selected.
	Selected bool

	// Focused is true if the cell has keyboard focus within the grid.
	Focused bool

	// Hovered is true if the mouse cursor is over this cell.
	Hovered bool

	// Disabled is true if the grid is disabled.
	Disabled bool

	// ColorScheme provides theme-derived colors.
	ColorScheme GridColorScheme
}

// GridColorScheme provides theme-derived colors for grid painting.
// Zero value means the painter should use its built-in defaults.
type GridColorScheme struct {
	SelectionColor widget.Color // selected cell background
	HoverColor     widget.Color // hovered cell background
	FocusColor     widget.Color // focused cell border/background
	EmptyTextColor widget.Color // empty state text color
	CellBackground widget.Color // cell background
}

// DefaultPainter provides a minimal fallback painter with no design system styling.
// It draws simple grid visuals -- useful for testing and as a base reference.
type DefaultPainter struct{}

// PaintCellBackground draws the background for a grid cell.
func (p DefaultPainter) PaintCellBackground(canvas widget.Canvas, cps CellPaintState) {
	if cps.Bounds.IsEmpty() {
		return
	}
	hasScheme := cps.ColorScheme != (GridColorScheme{})
	if cps.Hovered && !cps.Disabled {
		color := defaultHoverColor
		if hasScheme {
			color = cps.ColorScheme.HoverColor
		}
		canvas.DrawRect(cps.Bounds, color)
	}
}

// PaintSelection draws the selection highlight for a selected cell.
func (p DefaultPainter) PaintSelection(canvas widget.Canvas, cps CellPaintState) {
	if cps.Bounds.IsEmpty() || !cps.Selected {
		return
	}
	hasScheme := cps.ColorScheme != (GridColorScheme{})
	color := defaultSelectionColor
	if hasScheme {
		color = cps.ColorScheme.SelectionColor
	}
	canvas.DrawRect(cps.Bounds, color)

	// Draw focus border.
	if cps.Focused && !cps.Disabled {
		focusColor := defaultFocusBorderColor
		if hasScheme {
			focusColor = cps.ColorScheme.FocusColor
		}
		canvas.StrokeRect(cps.Bounds, focusColor, focusBorderWidth)
	}
}

// PaintEmptyState draws a centered "No items" message.
func (p DefaultPainter) PaintEmptyState(canvas widget.Canvas, bounds geometry.Rect) {
	if bounds.IsEmpty() {
		return
	}
	canvas.DrawText(emptyStateText, bounds, emptyStateFontSize, defaultEmptyTextColor, false, emptyStateAlign)
}

// Painting constants.
const (
	focusBorderWidth float32 = 2

	emptyStateFontSize float32 = 14
	emptyStateAlign            = widget.TextAlignCenter
	emptyStateText             = "No items"
)

// Default colors for DefaultPainter.
var (
	defaultSelectionColor   = widget.RGBA(0.23, 0.51, 0.96, 0.12)
	defaultHoverColor       = widget.RGBA(0.0, 0.0, 0.0, 0.04)
	defaultFocusBorderColor = widget.Hex(0x6750A4).WithAlpha(0.7)
	defaultEmptyTextColor   = widget.RGBA(0.5, 0.5, 0.5, 1.0)
)
