package material3

import (
	"github.com/gogpu/ui/core/gridview"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// GridViewPainter renders grid view elements using Material 3 design tokens.
// It maps cell states (selected, hovered, focused) to the M3 color scheme
// with primary for selection, secondary-container for hover, and surface for cells.
//
// If Theme is nil, GridViewPainter falls back to the default M3 purple palette.
type GridViewPainter struct {
	Theme *Theme // nil uses default M3 purple fallback
}

// resolveColors returns the GridColorScheme derived from the painter's Theme.
// If Theme is nil, it returns the default M3 purple color scheme.
func (p GridViewPainter) resolveColors() gridview.GridColorScheme {
	if p.Theme == nil {
		return m3DefaultGridColors
	}
	cs := p.Theme.Colors
	return gridview.GridColorScheme{
		SelectionColor: cs.Primary.WithAlpha(0.12),
		HoverColor:     cs.SecondaryContainer.WithAlpha(0.3),
		FocusColor:     cs.Primary.WithAlpha(0.7),
		EmptyTextColor: cs.OnSurfaceVariant,
		CellBackground: cs.Surface,
	}
}

// PaintCellBackground draws the background for a grid cell.
// This is called before the cell widget's own Draw method.
func (p GridViewPainter) PaintCellBackground(canvas widget.Canvas, cps gridview.CellPaintState) {
	if cps.Bounds.IsEmpty() {
		return
	}

	// Determine the color scheme to use.
	colors := cps.ColorScheme
	if colors == (gridview.GridColorScheme{}) {
		colors = p.resolveColors()
	}

	if cps.Hovered && !cps.Disabled {
		canvas.DrawRoundRect(cps.Bounds, colors.HoverColor, m3GridCellRadius)
	}
}

// PaintSelection draws the selection highlight for a selected cell.
// This is called before the cell widget's own Draw method.
func (p GridViewPainter) PaintSelection(canvas widget.Canvas, cps gridview.CellPaintState) {
	if cps.Bounds.IsEmpty() || !cps.Selected {
		return
	}

	// Determine the color scheme to use.
	colors := cps.ColorScheme
	if colors == (gridview.GridColorScheme{}) {
		colors = p.resolveColors()
	}

	canvas.DrawRoundRect(cps.Bounds, colors.SelectionColor, m3GridCellRadius)

	// Draw focus border.
	if cps.Focused && !cps.Disabled {
		canvas.StrokeRoundRect(cps.Bounds, colors.FocusColor, m3GridCellRadius, m3GridFocusBorderWidth)
	}
}

// PaintEmptyState draws a centered empty state message.
func (p GridViewPainter) PaintEmptyState(canvas widget.Canvas, bounds geometry.Rect) {
	if bounds.IsEmpty() {
		return
	}

	textColor := m3GridDefaultEmptyText
	if p.Theme != nil {
		textColor = p.Theme.Colors.OnSurfaceVariant
	}

	canvas.DrawText(m3GridEmptyStateText, bounds, m3GridEmptyFontSize, textColor, false, m3GridEmptyTextAlign)
}

// m3DefaultGridColors holds the default M3 purple color scheme for grid views.
// Used as a fallback when no Theme is provided.
var m3DefaultGridColors = gridview.GridColorScheme{
	SelectionColor: widget.Hex(0x6750A4).WithAlpha(0.12), // M3 primary with low alpha
	HoverColor:     widget.Hex(0xE8DEF8).WithAlpha(0.3),  // M3 secondary container with alpha
	FocusColor:     widget.Hex(0x6750A4).WithAlpha(0.7),  // M3 primary focus ring
	EmptyTextColor: widget.Hex(0x49454F),                 // M3 on-surface-variant
	CellBackground: widget.Hex(0xFFFBFE),                 // M3 surface
}

// Default M3 empty text color.
var m3GridDefaultEmptyText = widget.Hex(0x49454F)

// M3 grid view drawing constants.
const (
	m3GridCellRadius       float32 = 8
	m3GridFocusBorderWidth float32 = 2
	m3GridEmptyFontSize    float32 = 14
	m3GridEmptyTextAlign           = widget.TextAlignCenter
	m3GridEmptyStateText           = "No items"
)

// Compile-time check that GridViewPainter implements Painter.
var _ gridview.Painter = GridViewPainter{}
