package cupertino

import (
	"github.com/gogpu/ui/core/dropdown"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// DropdownPainter renders dropdowns using Apple HIG design tokens.
// Cupertino dropdowns use rounded popover menus with subtle shadows
// and a chevron indicator following iOS picker conventions.
//
// If Theme is nil, DropdownPainter falls back to the default system blue palette.
type DropdownPainter struct {
	Theme *Theme // nil uses default system blue fallback
}

// resolveColors returns the DropdownColorScheme derived from the painter's Theme.
func (p DropdownPainter) resolveColors() dropdown.DropdownColorScheme {
	if p.Theme == nil {
		return cupDefaultDropdownColors
	}
	cs := p.Theme.Colors
	return dropdown.DropdownColorScheme{
		Background:      cs.TertiarySystemBackground,
		Border:          cs.Separator,
		FocusBorder:     cs.Accent,
		TextColor:       cs.Label,
		PlaceholderText: cs.TertiaryLabel,
		DisabledBg:      cs.SecondarySystemBackground,
		DisabledFg:      cs.QuaternaryLabel,
		MenuBg:          cs.TertiarySystemBackground,
		MenuBorder:      cs.OpaqueSeparator,
		ItemHover:       cs.Accent.WithAlpha(cupDDItemHoverAlpha),
		ItemSelected:    cs.Accent.WithAlpha(cupDDItemSelectedAlpha),
		ItemDisabled:    cs.QuaternaryLabel,
		ChevronColor:    cs.SecondaryLabel,
		FocusRing:       cs.Accent.WithAlpha(cupDDFocusAlpha),
	}
}

// PaintTrigger renders a dropdown trigger according to Apple HIG specifications.
func (p DropdownPainter) PaintTrigger(canvas widget.Canvas, st *dropdown.TriggerPaintState) {
	if st.Bounds.IsEmpty() {
		return
	}

	colors := st.ColorScheme
	if colors == (dropdown.DropdownColorScheme{}) {
		colors = p.resolveColors()
	}

	cupDDPaintTriggerBg(canvas, st, colors)
	cupDDPaintTriggerBorder(canvas, st, colors)
	cupDDPaintTriggerText(canvas, st, colors)
	cupDDPaintTriggerChevron(canvas, st, colors)

	if st.Focused && !st.Disabled {
		cupDDDrawFocusRing(canvas, st.Bounds, colors)
	}
}

// PaintMenu renders a dropdown menu according to Apple HIG specifications.
func (p DropdownPainter) PaintMenu(canvas widget.Canvas, st *dropdown.MenuPaintState) {
	if st.Bounds.IsEmpty() {
		return
	}

	colors := st.ColorScheme
	if colors == (dropdown.DropdownColorScheme{}) {
		colors = p.resolveColors()
	}

	// Menu background with rounded corners (iOS popover).
	canvas.DrawRoundRect(st.Bounds, colors.MenuBg, cupDDMenuRadius)
	canvas.StrokeRoundRect(st.Bounds, colors.MenuBorder, cupDDMenuRadius, cupDDBorderWidth)

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

		cupDDPaintMenuItem(canvas, itemRect, item, i, st.HighlightedIndex, st.SelectedIndex, colors)
	}
}

// cupDDPaintTriggerBg draws the trigger background.
func cupDDPaintTriggerBg(canvas widget.Canvas, st *dropdown.TriggerPaintState, colors dropdown.DropdownColorScheme) {
	bg := colors.Background
	if st.Disabled {
		bg = colors.DisabledBg
	}
	canvas.DrawRoundRect(st.Bounds, bg, cupDDTriggerRadius)
}

// cupDDPaintTriggerBorder draws the trigger outline.
func cupDDPaintTriggerBorder(canvas widget.Canvas, st *dropdown.TriggerPaintState, colors dropdown.DropdownColorScheme) {
	borderColor := colors.Border
	strokeWidth := cupDDBorderWidth

	switch {
	case st.Disabled:
		borderColor = colors.DisabledFg
	case st.Focused:
		borderColor = colors.FocusBorder
		strokeWidth = cupDDFocusBorderWidth
	case st.Hovered:
		borderColor = colors.TextColor
	}

	canvas.StrokeRoundRect(st.Bounds, borderColor, cupDDTriggerRadius, strokeWidth)
}

// cupDDPaintTriggerText draws the selected text or placeholder.
func cupDDPaintTriggerText(canvas widget.Canvas, st *dropdown.TriggerPaintState, colors dropdown.DropdownColorScheme) {
	textColor := colors.TextColor
	if st.IsPlaceholder {
		textColor = colors.PlaceholderText
	}
	if st.Disabled {
		textColor = colors.DisabledFg
	}

	textBounds := geometry.Rect{
		Min: geometry.Pt(st.Bounds.Min.X+cupDDContentPaddingH, st.Bounds.Min.Y),
		Max: geometry.Pt(st.Bounds.Max.X-cupDDChevronWidth-cupDDContentPaddingH, st.Bounds.Max.Y),
	}
	canvas.DrawText(st.SelectedText, textBounds, cupDDFontSize, textColor, false, cupDDTextAlignLeft)
}

// cupDDPaintTriggerChevron draws the chevron indicator.
func cupDDPaintTriggerChevron(canvas widget.Canvas, st *dropdown.TriggerPaintState, colors dropdown.DropdownColorScheme) {
	chevronColor := colors.ChevronColor
	if st.Disabled {
		chevronColor = colors.DisabledFg
	}

	chevronX := st.Bounds.Max.X - cupDDChevronWidth - cupDDContentPaddingH/2
	chevronY := st.Bounds.Center().Y
	cupDDDrawChevron(canvas, geometry.Pt(chevronX, chevronY), st.Open, chevronColor)
}

// cupDDDrawChevron draws a Cupertino up/down chevron indicator.
func cupDDDrawChevron(canvas widget.Canvas, center geometry.Point, open bool, color widget.Color) {
	size := cupDDChevronSize
	if open {
		canvas.DrawLine(
			geometry.Pt(center.X-size, center.Y+size/2),
			geometry.Pt(center.X, center.Y-size/2),
			color, cupDDChevronStroke,
		)
		canvas.DrawLine(
			geometry.Pt(center.X, center.Y-size/2),
			geometry.Pt(center.X+size, center.Y+size/2),
			color, cupDDChevronStroke,
		)
	} else {
		canvas.DrawLine(
			geometry.Pt(center.X-size, center.Y-size/2),
			geometry.Pt(center.X, center.Y+size/2),
			color, cupDDChevronStroke,
		)
		canvas.DrawLine(
			geometry.Pt(center.X, center.Y+size/2),
			geometry.Pt(center.X+size, center.Y-size/2),
			color, cupDDChevronStroke,
		)
	}
}

// cupDDPaintMenuItem draws a single menu item.
func cupDDPaintMenuItem(
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
		Min: geometry.Pt(itemRect.Min.X+cupDDContentPaddingH, itemRect.Min.Y),
		Max: geometry.Pt(itemRect.Max.X-cupDDContentPaddingH, itemRect.Max.Y),
	}
	canvas.DrawText(item.DisplayText(), textRect, cupDDFontSize, textColor, false, cupDDTextAlignLeft)
}

// cupDDDrawFocusRing draws the blue focus ring around the trigger.
func cupDDDrawFocusRing(canvas widget.Canvas, bounds geometry.Rect, colors dropdown.DropdownColorScheme) {
	ringBounds := bounds.Expand(cupDDFocusRingOffset)
	ringRadius := cupDDTriggerRadius + cupDDFocusRingOffset
	canvas.StrokeRoundRect(ringBounds, colors.FocusRing, ringRadius, cupDDFocusRingStroke)
}

// cupDefaultDropdownColors holds the default Cupertino dropdown color scheme.
var cupDefaultDropdownColors = dropdown.DropdownColorScheme{
	Background:      widget.ColorWhite,
	Border:          widget.RGBA(0.235, 0.235, 0.263, 0.29),
	FocusBorder:     systemBlue,
	TextColor:       widget.RGBA(0.0, 0.0, 0.0, 1.0),
	PlaceholderText: widget.RGBA(0.235, 0.235, 0.263, 0.3),
	DisabledBg:      widget.Hex(0xF2F2F7),
	DisabledFg:      widget.RGBA(0.235, 0.235, 0.263, 0.18),
	MenuBg:          widget.ColorWhite,
	MenuBorder:      widget.Hex(0xC6C6C8),
	ItemHover:       systemBlue.WithAlpha(0.1),
	ItemSelected:    systemBlue.WithAlpha(0.15),
	ItemDisabled:    widget.RGBA(0.235, 0.235, 0.263, 0.18),
	ChevronColor:    widget.RGBA(0.235, 0.235, 0.263, 0.6),
	FocusRing:       systemBlue.WithAlpha(0.6),
}

// Cupertino dropdown drawing constants.
const (
	cupDDTriggerRadius     float32 = 8
	cupDDMenuRadius        float32 = 10
	cupDDBorderWidth       float32 = 0.5
	cupDDFocusBorderWidth  float32 = 2
	cupDDContentPaddingH   float32 = 12
	cupDDFontSize          float32 = 15
	cupDDTextAlignLeft             = widget.TextAlignLeft
	cupDDChevronWidth      float32 = 24
	cupDDChevronSize       float32 = 5
	cupDDChevronStroke     float32 = 1.5
	cupDDFocusRingOffset   float32 = 3
	cupDDFocusRingStroke   float32 = 2.5
	cupDDFocusAlpha        float32 = 0.6
	cupDDItemHoverAlpha    float32 = 0.1
	cupDDItemSelectedAlpha float32 = 0.15
)

// Compile-time check that DropdownPainter implements Painter.
var _ dropdown.Painter = DropdownPainter{}
