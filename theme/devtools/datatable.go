package devtools

import (
	"github.com/gogpu/ui/core/datatable"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// DataTablePainter renders data tables using DevTools design tokens.
// DevTools tables use striped rows (alternating Background/Surface), compact
// 24px row height, and a fixed header with SurfaceElevated background,
// matching JetBrains IDE table styling.
//
// If Theme is nil, DataTablePainter falls back to the default DevTools dark palette.
type DataTablePainter struct {
	Theme *Theme // nil uses default DevTools dark fallback
}

// resolveColors returns DevTools-derived colors for table painting.
func (p DataTablePainter) resolveColors() datatable.TableColorScheme {
	if p.Theme == nil {
		return dtDefaultTableColors
	}
	cs := p.Theme.Colors
	return datatable.TableColorScheme{
		HeaderBackground: cs.SurfaceElevated,
		HeaderText:       cs.OnSurface,
		RowBackground:    cs.Background,
		RowAlternate:     cs.Surface,
		SelectionColor:   cs.Selection,
		HoverColor:       cs.ControlHover,
		FocusColor:       cs.BorderFocus,
		CellText:         cs.OnSurface,
		Divider:          cs.Border,
		EmptyText:        cs.OnSurfaceSecondary,
	}
}

// PaintHeader draws the table header background with DevTools elevated surface color.
func (p DataTablePainter) PaintHeader(canvas widget.Canvas, bounds geometry.Rect, hps datatable.HeaderPaintState) {
	if bounds.IsEmpty() {
		return
	}
	colors := p.effectiveHeaderColors(hps.ColorScheme)
	canvas.DrawRect(bounds, colors.HeaderBackground)

	// Bottom divider line.
	dividerRect := geometry.NewRect(bounds.Min.X, bounds.Max.Y-dtTableDividerHeight, bounds.Width(), dtTableDividerHeight)
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
		bounds.Min.X+dtTableCellPaddingH,
		bounds.Min.Y,
		bounds.Width()-dtTableCellPaddingH*2,
		bounds.Height(),
	)

	fg := colors.HeaderText
	if hcs.Disabled {
		fg = fg.WithAlpha(dtTableDisabledAlpha)
	}
	canvas.DrawText(displayText, textBounds, dtTableHeaderFontSize, fg, true, hcs.Align)
}

// PaintRow draws the row background with zebra striping and DevTools selection/hover highlights.
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

	// Focus ring (1px DevTools style).
	if rps.Focused && !rps.Disabled {
		canvas.StrokeRect(rps.Bounds, colors.FocusColor, dtTableFocusBorderWidth)
	}
}

// PaintCell draws a single data cell with DevTools on-surface text color.
func (p DataTablePainter) PaintCell(canvas widget.Canvas, cps datatable.CellPaintState) {
	if cps.Bounds.IsEmpty() {
		return
	}
	colors := p.effectiveCellColors(cps.ColorScheme)
	fg := colors.CellText
	if cps.Disabled {
		fg = fg.WithAlpha(dtTableDisabledAlpha)
	}

	textBounds := geometry.NewRect(
		cps.Bounds.Min.X+dtTableCellPaddingH,
		cps.Bounds.Min.Y,
		cps.Bounds.Width()-dtTableCellPaddingH*2,
		cps.Bounds.Height(),
	)
	canvas.DrawText(cps.Value, textBounds, dtTableCellFontSize, fg, false, cps.Align)
}

// PaintEmptyState draws a centered "No data" message.
func (p DataTablePainter) PaintEmptyState(canvas widget.Canvas, bounds geometry.Rect) {
	if bounds.IsEmpty() {
		return
	}
	colors := p.resolveColors()
	canvas.DrawText(dtTableEmptyText, bounds, dtTableEmptyFontSize, colors.EmptyText, false, widget.TextAlignCenter)
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

// dtDefaultTableColors holds the default DevTools dark color scheme for data tables.
var dtDefaultTableColors = datatable.TableColorScheme{
	HeaderBackground: widget.Hex(0x393B40), // Gray3 (elevated surface)
	HeaderText:       widget.Hex(0xDFE1E5), // Gray12 (primary text)
	RowBackground:    widget.Hex(0x1E1F22), // Gray1 (background)
	RowAlternate:     widget.Hex(0x2B2D30), // Gray2 (surface)
	SelectionColor:   widget.Hex(0x2E436E), // Blue2 (selection)
	HoverColor:       widget.Hex(0x43454A), // Gray4 (control hover)
	FocusColor:       widget.Hex(0x3574F0), // Blue6 (accent)
	CellText:         widget.Hex(0xDFE1E5), // Gray12 (primary text)
	Divider:          widget.Hex(0x393B40), // Gray3 (border)
	EmptyText:        widget.Hex(0x9DA0A8), // Gray9 (secondary text)
}

// DevTools data table drawing constants.
const (
	dtTableCellPaddingH     float32 = 8
	dtTableHeaderFontSize   float32 = 12
	dtTableCellFontSize     float32 = 13
	dtTableFocusBorderWidth float32 = 1
	dtTableDividerHeight    float32 = 1
	dtTableEmptyFontSize    float32 = 13
	dtTableDisabledAlpha    float32 = 0.38
	dtTableEmptyText                = "No data"
)

// Compile-time check that DataTablePainter implements Painter.
var _ datatable.Painter = DataTablePainter{}
