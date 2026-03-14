package material3

import (
	"github.com/gogpu/ui/core/listview"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// ListViewPainter renders list view elements using Material 3 design tokens.
// It maps list states (divider, empty, hover, selection) to the M3 color scheme
// and applies appropriate visual feedback.
//
// If Theme is nil, ListViewPainter falls back to the default M3 purple palette.
type ListViewPainter struct {
	Theme *Theme // nil uses default M3 purple fallback
}

// resolveColors returns the ListColorScheme derived from the painter's Theme.
func (p ListViewPainter) resolveColors() listview.ListColorScheme {
	if p.Theme == nil {
		return m3DefaultListColors
	}
	cs := p.Theme.Colors
	return listview.ListColorScheme{
		DividerColor:      cs.OutlineVariant,
		SelectionColor:    cs.SecondaryContainer,
		HoverColor:        cs.OnSurface.WithAlpha(0.08),
		FocusColor:        cs.Primary.WithAlpha(0.7),
		EmptyTextColor:    cs.OnSurfaceVariant,
		ItemBackground:    widget.ColorTransparent,
		ItemBackgroundAlt: widget.ColorTransparent,
	}
}

// PaintDivider draws a M3-styled divider between list items.
func (p ListViewPainter) PaintDivider(canvas widget.Canvas, ds listview.DividerState) {
	if ds.Bounds.IsEmpty() {
		return
	}
	colors := ds.ColorScheme
	if colors == (listview.ListColorScheme{}) {
		colors = p.resolveColors()
	}
	ds.ColorScheme = colors
	listview.DefaultPainter{}.PaintDivider(canvas, ds)
}

// PaintEmptyState draws a M3-styled empty state message.
func (p ListViewPainter) PaintEmptyState(canvas widget.Canvas, bounds geometry.Rect) {
	if bounds.IsEmpty() {
		return
	}
	colors := p.resolveColors()
	canvas.DrawText(
		m3EmptyStateText,
		bounds,
		m3EmptyStateFontSize,
		colors.EmptyTextColor,
		false,
		m3EmptyStateAlign,
	)
}

// PaintItemBackground draws the M3 item background with hover state.
func (p ListViewPainter) PaintItemBackground(canvas widget.Canvas, ips listview.ItemPaintState) {
	if ips.Bounds.IsEmpty() {
		return
	}
	colors := ips.ColorScheme
	if colors == (listview.ListColorScheme{}) {
		colors = p.resolveColors()
	}
	ips.ColorScheme = colors
	listview.DefaultPainter{}.PaintItemBackground(canvas, ips)
}

// PaintSelection draws the M3 selection highlight.
func (p ListViewPainter) PaintSelection(canvas widget.Canvas, ips listview.ItemPaintState) {
	if ips.Bounds.IsEmpty() || !ips.Selected {
		return
	}
	colors := ips.ColorScheme
	if colors == (listview.ListColorScheme{}) {
		colors = p.resolveColors()
	}
	ips.ColorScheme = colors
	listview.DefaultPainter{}.PaintSelection(canvas, ips)
}

// m3DefaultListColors holds the default M3 purple color scheme for lists.
var m3DefaultListColors = listview.ListColorScheme{
	DividerColor:      widget.Hex(0xCAC4D0),                 // M3 outline variant
	SelectionColor:    widget.Hex(0xE8DEF8),                 // M3 secondary container
	HoverColor:        widget.Hex(0x1D1B20).WithAlpha(0.08), // M3 on-surface 8%
	FocusColor:        widget.Hex(0x6750A4).WithAlpha(0.7),  // M3 primary 70%
	EmptyTextColor:    widget.Hex(0x49454F),                 // M3 on-surface-variant
	ItemBackground:    widget.ColorTransparent,
	ItemBackgroundAlt: widget.ColorTransparent,
}

// M3 list painting constants.
const (
	m3EmptyStateFontSize float32 = 14
	m3EmptyStateAlign            = widget.TextAlignCenter
	m3EmptyStateText             = "No items"
)

// Compile-time check that ListViewPainter implements Painter.
var _ listview.Painter = ListViewPainter{}
