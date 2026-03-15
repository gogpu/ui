package fluent

import (
	"github.com/gogpu/ui/core/dropdown"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// DropdownPainter renders dropdowns using Fluent Design tokens.
//
// If Theme is nil, DropdownPainter falls back to the default Fluent Blue palette.
type DropdownPainter struct {
	Theme *Theme // nil uses default Fluent Blue fallback
}

// resolveColors returns the DropdownColorScheme derived from the painter's Theme.
func (p DropdownPainter) resolveColors() dropdown.DropdownColorScheme {
	if p.Theme == nil {
		return flDefaultDropdownColors
	}
	cs := p.Theme.Colors
	return dropdown.DropdownColorScheme{
		Background:      cs.SurfaceTertiary,
		Border:          cs.StrokeDefault,
		FocusBorder:     cs.Accent,
		TextColor:       cs.OnSurface,
		PlaceholderText: cs.OnSurfaceSecond,
		DisabledBg:      cs.FillDisable,
		DisabledFg:      cs.OnSurfaceSecond.WithAlpha(flDisabledAlpha),
		MenuBg:          cs.Surface,
		MenuBorder:      cs.StrokeDefault,
		ItemHover:       cs.FillSecond,
		ItemSelected:    cs.AccentLight,
		ItemDisabled:    cs.OnSurfaceSecond.WithAlpha(flDisabledAlpha),
		ChevronColor:    cs.OnSurfaceSecond,
		FocusRing:       cs.StrokeFocus,
	}
}

// PaintTrigger renders a dropdown trigger according to Fluent Design specifications.
func (p DropdownPainter) PaintTrigger(canvas widget.Canvas, st *dropdown.TriggerPaintState) {
	if st.Bounds.IsEmpty() {
		return
	}

	colors := st.ColorScheme
	if colors == (dropdown.DropdownColorScheme{}) {
		colors = p.resolveColors()
	}

	flDDPaintTriggerBg(canvas, st, colors)
	flDDPaintTriggerBorder(canvas, st, colors)
	flDDPaintTriggerText(canvas, st, colors)
	flDDPaintTriggerChevron(canvas, st, colors)

	if st.Focused && !st.Disabled {
		flDrawFocusRing(canvas, st.Bounds, flDDTriggerRadius, colors.FocusRing)
	}
}

// PaintMenu renders a dropdown menu according to Fluent Design specifications.
func (p DropdownPainter) PaintMenu(canvas widget.Canvas, st *dropdown.MenuPaintState) {
	if st.Bounds.IsEmpty() {
		return
	}

	colors := st.ColorScheme
	if colors == (dropdown.DropdownColorScheme{}) {
		colors = p.resolveColors()
	}

	// Menu background with subtle border.
	canvas.DrawRoundRect(st.Bounds, colors.MenuBg, flDDMenuRadius)
	canvas.StrokeRoundRect(st.Bounds, colors.MenuBorder, flDDMenuRadius, flDDBorderWidth)

	canvas.PushClip(st.Bounds)
	defer canvas.PopClip()

	endIndex := st.ScrollOffset + st.VisibleCount
	if endIndex > len(st.Items) {
		endIndex = len(st.Items)
	}

	for i := st.ScrollOffset; i < endIndex; i++ {
		item := st.Items[i]
		row := i - st.ScrollOffset
		itemRect := geometry.Rect{
			Min: geometry.Pt(st.Bounds.Min.X, st.Bounds.Min.Y+float32(row)*st.ItemHeight),
			Max: geometry.Pt(st.Bounds.Max.X, st.Bounds.Min.Y+float32(row+1)*st.ItemHeight),
		}

		flDDPaintMenuItem(canvas, itemRect, item, i, st.HighlightedIndex, st.SelectedIndex, colors)
	}
}

// flDDPaintTriggerBg draws the trigger background.
func flDDPaintTriggerBg(canvas widget.Canvas, st *dropdown.TriggerPaintState, colors dropdown.DropdownColorScheme) {
	bg := colors.Background
	if st.Disabled {
		bg = colors.DisabledBg
	}
	canvas.DrawRoundRect(st.Bounds, bg, flDDTriggerRadius)
}

// flDDPaintTriggerBorder draws the trigger outline.
func flDDPaintTriggerBorder(canvas widget.Canvas, st *dropdown.TriggerPaintState, colors dropdown.DropdownColorScheme) {
	borderColor := colors.Border
	strokeWidth := flDDBorderWidth

	switch {
	case st.Disabled:
		borderColor = colors.DisabledFg
	case st.Focused:
		borderColor = colors.FocusBorder
		strokeWidth = flDDFocusBorderWidth
	case st.Hovered:
		borderColor = colors.TextColor
	}

	canvas.StrokeRoundRect(st.Bounds, borderColor, flDDTriggerRadius, strokeWidth)
}

// flDDPaintTriggerText draws the selected text or placeholder.
func flDDPaintTriggerText(canvas widget.Canvas, st *dropdown.TriggerPaintState, colors dropdown.DropdownColorScheme) {
	textColor := colors.TextColor
	if st.IsPlaceholder {
		textColor = colors.PlaceholderText
	}
	if st.Disabled {
		textColor = colors.DisabledFg
	}

	textBounds := geometry.Rect{
		Min: geometry.Pt(st.Bounds.Min.X+flDDContentPaddingH, st.Bounds.Min.Y),
		Max: geometry.Pt(st.Bounds.Max.X-flDDChevronWidth-flDDContentPaddingH, st.Bounds.Max.Y),
	}
	canvas.DrawText(st.SelectedText, textBounds, flDDFontSize, textColor, false, flDDTextAlignLeft)
}

// flDDPaintTriggerChevron draws the chevron indicator.
func flDDPaintTriggerChevron(canvas widget.Canvas, st *dropdown.TriggerPaintState, colors dropdown.DropdownColorScheme) {
	chevronColor := colors.ChevronColor
	if st.Disabled {
		chevronColor = colors.DisabledFg
	}

	chevronX := st.Bounds.Max.X - flDDChevronWidth - flDDContentPaddingH/2
	chevronY := st.Bounds.Center().Y
	flDDDrawChevron(canvas, geometry.Pt(chevronX, chevronY), st.Open, chevronColor)
}

// flDDDrawChevron draws a Fluent up/down chevron indicator.
func flDDDrawChevron(canvas widget.Canvas, center geometry.Point, open bool, color widget.Color) {
	size := flDDChevronSize
	if open {
		canvas.DrawLine(
			geometry.Pt(center.X-size, center.Y+size/2),
			geometry.Pt(center.X, center.Y-size/2),
			color, flDDChevronStroke,
		)
		canvas.DrawLine(
			geometry.Pt(center.X, center.Y-size/2),
			geometry.Pt(center.X+size, center.Y+size/2),
			color, flDDChevronStroke,
		)
	} else {
		canvas.DrawLine(
			geometry.Pt(center.X-size, center.Y-size/2),
			geometry.Pt(center.X, center.Y+size/2),
			color, flDDChevronStroke,
		)
		canvas.DrawLine(
			geometry.Pt(center.X, center.Y+size/2),
			geometry.Pt(center.X+size, center.Y-size/2),
			color, flDDChevronStroke,
		)
	}
}

// flDDPaintMenuItem draws a single menu item.
func flDDPaintMenuItem(
	canvas widget.Canvas,
	itemRect geometry.Rect,
	item dropdown.ItemDef,
	index int,
	highlightedIndex int,
	selectedIndex int,
	colors dropdown.DropdownColorScheme,
) {
	switch index {
	case highlightedIndex:
		canvas.DrawRect(itemRect, colors.ItemHover)
	case selectedIndex:
		canvas.DrawRect(itemRect, colors.ItemSelected)
	}

	textColor := colors.TextColor
	if item.Disabled {
		textColor = colors.ItemDisabled
	}

	textRect := geometry.Rect{
		Min: geometry.Pt(itemRect.Min.X+flDDContentPaddingH, itemRect.Min.Y),
		Max: geometry.Pt(itemRect.Max.X-flDDContentPaddingH, itemRect.Max.Y),
	}
	canvas.DrawText(item.DisplayText(), textRect, flDDFontSize, textColor, false, flDDTextAlignLeft)
}

// flDefaultDropdownColors holds the default Fluent dropdown color scheme.
var flDefaultDropdownColors = dropdown.DropdownColorScheme{
	Background:      widget.Hex(0xFAFAFA),
	Border:          widget.RGBA(0, 0, 0, 0.14),
	FocusBorder:     DefaultAccentColor,
	TextColor:       widget.Hex(0x1A1A1A),
	PlaceholderText: widget.Hex(0x616161),
	DisabledBg:      widget.RGBA(0, 0, 0, 0.04),
	DisabledFg:      widget.RGBA(0.38, 0.38, 0.38, 0.38),
	MenuBg:          widget.ColorWhite,
	MenuBorder:      widget.RGBA(0, 0, 0, 0.14),
	ItemHover:       widget.RGBA(0, 0, 0, 0.06),
	ItemSelected:    lighten(DefaultAccentColor, 0.85),
	ItemDisabled:    widget.RGBA(0.38, 0.38, 0.38, 0.38),
	ChevronColor:    widget.Hex(0x616161),
	FocusRing:       DefaultAccentColor,
}

// Fluent dropdown drawing constants.
const (
	flDDTriggerRadius    float32 = 4
	flDDMenuRadius       float32 = 4
	flDDBorderWidth      float32 = 1
	flDDFocusBorderWidth float32 = 2
	flDDContentPaddingH  float32 = 12
	flDDFontSize         float32 = 14
	flDDTextAlignLeft            = widget.TextAlignLeft
	flDDChevronWidth     float32 = 24
	flDDChevronSize      float32 = 5
	flDDChevronStroke    float32 = 1.5
)

// Compile-time check that DropdownPainter implements Painter.
var _ dropdown.Painter = DropdownPainter{}
