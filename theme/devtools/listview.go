package devtools

import (
	"github.com/gogpu/ui/core/listview"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// ListViewPainter renders list view elements using DevTools design tokens.
// DevTools lists use wide full-row selection (Blue2 background), 24px row
// height, ControlHover highlight, and optional alternating rows
// (Background/Surface), matching JetBrains IDE file list styling.
//
// If Theme is nil, ListViewPainter falls back to the default DevTools dark palette.
type ListViewPainter struct {
	Theme *Theme // nil uses default DevTools dark fallback
}

// resolveColors returns the ListColorScheme derived from the painter's Theme.
func (p ListViewPainter) resolveColors() listview.ListColorScheme {
	if p.Theme == nil {
		return dtDefaultListColors
	}
	cs := p.Theme.Colors
	return listview.ListColorScheme{
		DividerColor:      cs.Border,
		SelectionColor:    cs.Selection,
		HoverColor:        cs.ControlHover,
		FocusColor:        cs.BorderFocus,
		EmptyTextColor:    cs.OnSurfaceSecondary,
		ItemBackground:    cs.Background,
		ItemBackgroundAlt: cs.Surface,
	}
}

// PaintDivider draws a DevTools-styled divider between list items.
func (p ListViewPainter) PaintDivider(canvas widget.Canvas, ds listview.DividerState) {
	if ds.Bounds.IsEmpty() {
		return
	}
	colors := ds.ColorScheme
	if colors == (listview.ListColorScheme{}) {
		colors = p.resolveColors()
	}
	canvas.DrawRect(ds.Bounds, colors.DividerColor)
}

// PaintEmptyState draws a centered "No items" message.
func (p ListViewPainter) PaintEmptyState(canvas widget.Canvas, bounds geometry.Rect) {
	if bounds.IsEmpty() {
		return
	}
	colors := p.resolveColors()
	canvas.DrawText(
		dtListEmptyStateText,
		bounds,
		dtListEmptyStateFontSize,
		colors.EmptyTextColor,
		false,
		dtListEmptyStateAlign,
	)
}

// PaintItemBackground draws the DevTools item background with hover state.
func (p ListViewPainter) PaintItemBackground(canvas widget.Canvas, ips listview.ItemPaintState) {
	if ips.Bounds.IsEmpty() {
		return
	}
	colors := ips.ColorScheme
	if colors == (listview.ListColorScheme{}) {
		colors = p.resolveColors()
	}

	// Hover highlight.
	if ips.Hovered && !ips.Disabled {
		canvas.DrawRect(ips.Bounds, colors.HoverColor)
	}
}

// PaintSelection draws the DevTools selection highlight (full-row, no rounded corners).
func (p ListViewPainter) PaintSelection(canvas widget.Canvas, ips listview.ItemPaintState) {
	if ips.Bounds.IsEmpty() || !ips.Selected {
		return
	}
	colors := ips.ColorScheme
	if colors == (listview.ListColorScheme{}) {
		colors = p.resolveColors()
	}
	canvas.DrawRect(ips.Bounds, colors.SelectionColor)

	// Focus border (1px DevTools style).
	if ips.Focused && !ips.Disabled {
		canvas.StrokeRect(ips.Bounds, colors.FocusColor, dtListFocusBorderWidth)
	}
}

// dtDefaultListColors holds the default DevTools dark color scheme for lists.
var dtDefaultListColors = listview.ListColorScheme{
	DividerColor:      widget.Hex(0x393B40), // Gray3 (border)
	SelectionColor:    widget.Hex(0x2E436E), // Blue2 (selection)
	HoverColor:        widget.Hex(0x43454A), // Gray4 (control hover)
	FocusColor:        widget.Hex(0x3574F0), // Blue6 (accent)
	EmptyTextColor:    widget.Hex(0x9DA0A8), // Gray9 (secondary text)
	ItemBackground:    widget.Hex(0x1E1F22), // Gray1 (background)
	ItemBackgroundAlt: widget.Hex(0x2B2D30), // Gray2 (surface)
}

// DevTools list painting constants.
const (
	dtListEmptyStateFontSize float32 = 13
	dtListEmptyStateAlign            = widget.TextAlignCenter
	dtListEmptyStateText             = "No items"
	dtListFocusBorderWidth   float32 = 1
)

// Compile-time check that ListViewPainter implements Painter.
var _ listview.Painter = ListViewPainter{}
