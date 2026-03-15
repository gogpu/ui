package material3

import (
	"github.com/gogpu/ui/core/datatable"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// DataTablePainter renders data tables using Material 3 design tokens.
// It maps M3 color roles to table elements: surface-container for header,
// on-surface for text, surface-variant for zebra striping, and primary for selection.
//
// If Theme is nil, DataTablePainter falls back to the default M3 purple palette.
type DataTablePainter struct {
	Theme *Theme // nil uses default M3 purple fallback
}

// resolveColors returns M3-derived colors for table painting.
func (p DataTablePainter) resolveColors() datatable.TableColorScheme {
	if p.Theme == nil {
		return m3DefaultTableColors
	}
	cs := p.Theme.Colors
	return datatable.TableColorScheme{
		HeaderBackground: cs.SurfaceContainerHighest,
		HeaderText:       cs.OnSurface,
		RowBackground:    cs.Surface,
		RowAlternate:     cs.SurfaceVariant.WithAlpha(0.3),
		SelectionColor:   cs.Primary.WithAlpha(0.12),
		HoverColor:       cs.OnSurface.WithAlpha(0.04),
		FocusColor:       cs.Primary.WithAlpha(0.7),
		CellText:         cs.OnSurface,
		Divider:          cs.OutlineVariant,
		EmptyText:        cs.OnSurfaceVariant,
	}
}

// PaintHeader draws the table header background with M3 surface-container color.
func (p DataTablePainter) PaintHeader(canvas widget.Canvas, bounds geometry.Rect, hps datatable.HeaderPaintState) {
	if bounds.IsEmpty() {
		return
	}
	colors := p.effectiveHeaderColors(hps.ColorScheme)
	canvas.DrawRect(bounds, colors.HeaderBackground)

	// Bottom divider line.
	dividerRect := geometry.NewRect(bounds.Min.X, bounds.Max.Y-m3TableDividerHeight, bounds.Width(), m3TableDividerHeight)
	canvas.DrawRect(dividerRect, colors.Divider)
}

// PaintHeaderCell draws a column header with title and sort indicator.
func (p DataTablePainter) PaintHeaderCell(canvas widget.Canvas, bounds geometry.Rect, hcs datatable.HeaderCellPaintState) {
	if bounds.IsEmpty() {
		return
	}

	colors := p.effectiveHeaderCellColors(hcs.ColorScheme)

	// Highlight on hover for sortable columns.
	if hcs.Hovered && hcs.Sortable && !hcs.Disabled {
		canvas.DrawRect(bounds, colors.HoverColor)
	}

	// Build display text with sort indicator.
	displayText := hcs.Title
	if indicator := hcs.SortDir.Indicator(); indicator != "" {
		displayText = hcs.Title + " " + indicator
	}

	// Inset for text padding.
	textBounds := geometry.NewRect(
		bounds.Min.X+m3TableCellPaddingH,
		bounds.Min.Y,
		bounds.Width()-m3TableCellPaddingH*2,
		bounds.Height(),
	)

	fg := colors.HeaderText
	if hcs.Disabled {
		fg = fg.WithAlpha(m3TableDisabledAlpha)
	}
	canvas.DrawText(displayText, textBounds, m3TableHeaderFontSize, fg, true, hcs.Align)
}

// PaintRow draws the row background with zebra striping and M3 selection/hover highlights.
func (p DataTablePainter) PaintRow(canvas widget.Canvas, rps datatable.RowPaintState) {
	if rps.Bounds.IsEmpty() {
		return
	}
	colors := p.effectiveRowColors(rps.ColorScheme)

	// Zebra striping for alternate rows.
	if rps.RowIndex%2 == 1 {
		canvas.DrawRect(rps.Bounds, colors.RowAlternate)
	}

	// Selection highlight.
	if rps.Selected {
		canvas.DrawRect(rps.Bounds, colors.SelectionColor)
	}

	// Hover highlight (only when not selected to avoid double-tinting).
	if rps.Hovered && !rps.Selected && !rps.Disabled {
		canvas.DrawRect(rps.Bounds, colors.HoverColor)
	}

	// Focus ring.
	if rps.Focused && !rps.Disabled {
		canvas.StrokeRect(rps.Bounds, colors.FocusColor, m3TableFocusBorderWidth)
	}
}

// PaintCell draws a single data cell with M3 on-surface text color.
func (p DataTablePainter) PaintCell(canvas widget.Canvas, cps datatable.CellPaintState) {
	if cps.Bounds.IsEmpty() {
		return
	}
	colors := p.effectiveCellColors(cps.ColorScheme)
	fg := colors.CellText
	if cps.Disabled {
		fg = fg.WithAlpha(m3TableDisabledAlpha)
	}

	textBounds := geometry.NewRect(
		cps.Bounds.Min.X+m3TableCellPaddingH,
		cps.Bounds.Min.Y,
		cps.Bounds.Width()-m3TableCellPaddingH*2,
		cps.Bounds.Height(),
	)
	canvas.DrawText(cps.Value, textBounds, m3TableCellFontSize, fg, false, cps.Align)
}

// PaintEmptyState draws a centered "No data" message with M3 on-surface-variant color.
func (p DataTablePainter) PaintEmptyState(canvas widget.Canvas, bounds geometry.Rect) {
	if bounds.IsEmpty() {
		return
	}
	colors := p.resolveColors()
	canvas.DrawText(m3TableEmptyText, bounds, m3TableEmptyFontSize, colors.EmptyText, false, widget.TextAlignCenter)
}

// effectiveHeaderColors returns colors, preferring the paint state's ColorScheme.
func (p DataTablePainter) effectiveHeaderColors(cs datatable.TableColorScheme) datatable.TableColorScheme {
	if cs != (datatable.TableColorScheme{}) {
		return cs
	}
	return p.resolveColors()
}

// effectiveHeaderCellColors returns colors for header cell painting.
func (p DataTablePainter) effectiveHeaderCellColors(cs datatable.TableColorScheme) datatable.TableColorScheme {
	if cs != (datatable.TableColorScheme{}) {
		return cs
	}
	return p.resolveColors()
}

// effectiveRowColors returns colors for row painting.
func (p DataTablePainter) effectiveRowColors(cs datatable.TableColorScheme) datatable.TableColorScheme {
	if cs != (datatable.TableColorScheme{}) {
		return cs
	}
	return p.resolveColors()
}

// effectiveCellColors returns colors for cell painting.
func (p DataTablePainter) effectiveCellColors(cs datatable.TableColorScheme) datatable.TableColorScheme {
	if cs != (datatable.TableColorScheme{}) {
		return cs
	}
	return p.resolveColors()
}

// m3DefaultTableColors holds default M3 purple fallback colors for data tables.
var m3DefaultTableColors = datatable.TableColorScheme{
	HeaderBackground: widget.Hex(0xE6E0E9),               // M3 surface-container-highest
	HeaderText:       widget.Hex(0x1C1B1F),               // M3 on-surface
	RowBackground:    widget.Hex(0xFFFBFE),               // M3 surface
	RowAlternate:     widget.RGBA(0.91, 0.87, 0.93, 0.3), // M3 surface-variant @ 30%
	SelectionColor:   widget.Hex(0x6750A4).WithAlpha(0.12),
	HoverColor:       widget.RGBA(0.12, 0.12, 0.13, 0.04),
	FocusColor:       widget.Hex(0x6750A4).WithAlpha(0.7),
	CellText:         widget.Hex(0x1C1B1F), // M3 on-surface
	Divider:          widget.Hex(0xCAC4D0), // M3 outline-variant
	EmptyText:        widget.Hex(0x49454F), // M3 on-surface-variant
}

// M3 data table drawing constants.
const (
	m3TableCellPaddingH     float32 = 12
	m3TableHeaderFontSize   float32 = 13
	m3TableCellFontSize     float32 = 14
	m3TableFocusBorderWidth float32 = 2
	m3TableDividerHeight    float32 = 1
	m3TableEmptyFontSize    float32 = 14
	m3TableDisabledAlpha    float32 = 0.38
	m3TableEmptyText                = "No data"
)

// Compile-time check that DataTablePainter implements Painter.
var _ datatable.Painter = DataTablePainter{}
