package material3

import (
	"github.com/gogpu/ui/core/dropdown"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// DropdownPainter renders dropdowns using Material 3 design tokens.
// It implements the outlined dropdown variant with theme-derived colors.
//
// If Theme is nil, DropdownPainter falls back to the default M3 purple palette.
type DropdownPainter struct {
	Theme *Theme // nil uses default M3 purple fallback
}

// resolveColors returns the DropdownColorScheme derived from the painter's Theme.
// If Theme is nil, it returns the default M3 dropdown color scheme.
func (p DropdownPainter) resolveColors() dropdown.DropdownColorScheme {
	if p.Theme == nil {
		return m3DefaultDropdownColors
	}
	cs := p.Theme.Colors
	return dropdown.DropdownColorScheme{
		Background:      cs.Surface,
		Border:          cs.Outline,
		FocusBorder:     cs.Primary,
		TextColor:       cs.OnSurface,
		PlaceholderText: cs.OnSurfaceVariant,
		DisabledBg:      cs.OnSurface.WithAlpha(0.04),
		DisabledFg:      cs.OnSurface.WithAlpha(0.38),
		MenuBg:          cs.Surface,
		MenuBorder:      cs.OutlineVariant,
		ItemHover:       cs.Primary.WithAlpha(0.08),
		ItemSelected:    cs.Primary.WithAlpha(0.12),
		ItemDisabled:    cs.OnSurface.WithAlpha(0.38),
		ChevronColor:    cs.OnSurfaceVariant,
		FocusRing:       cs.Primary.WithAlpha(0.7),
	}
}

// PaintTrigger renders a dropdown trigger according to Material 3 specifications.
func (p DropdownPainter) PaintTrigger(canvas widget.Canvas, st *dropdown.TriggerPaintState) {
	if st.Bounds.IsEmpty() {
		return
	}

	colors := st.ColorScheme
	if colors == (dropdown.DropdownColorScheme{}) {
		colors = p.resolveColors()
	}

	m3DDPaintTriggerBg(canvas, st, colors)
	m3DDPaintTriggerBorder(canvas, st, colors)
	m3DDPaintTriggerText(canvas, st, colors)
	m3DDPaintTriggerChevron(canvas, st, colors)

	if st.Focused && !st.Disabled {
		m3DDDrawFocusIndicator(canvas, st.Bounds, colors)
	}
}

// PaintMenu renders a dropdown menu according to Material 3 specifications.
func (p DropdownPainter) PaintMenu(canvas widget.Canvas, st *dropdown.MenuPaintState) {
	if st.Bounds.IsEmpty() {
		return
	}

	colors := st.ColorScheme
	if colors == (dropdown.DropdownColorScheme{}) {
		colors = p.resolveColors()
	}

	// Menu background with elevation shadow (M3 surface tint).
	canvas.DrawRoundRect(st.Bounds, colors.MenuBg, m3DDMenuRadius)
	canvas.StrokeRoundRect(st.Bounds, colors.MenuBorder, m3DDMenuRadius, m3DDBorderWidth)

	// Clip to menu bounds for items.
	canvas.PushClip(st.Bounds)
	defer canvas.PopClip()

	// Draw visible items.
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

		m3DDPaintMenuItem(canvas, itemRect, item, i, st.HighlightedIndex, st.SelectedIndex, colors)
	}
}

// m3DDPaintTriggerBg draws the trigger background.
func m3DDPaintTriggerBg(canvas widget.Canvas, st *dropdown.TriggerPaintState, colors dropdown.DropdownColorScheme) {
	bg := colors.Background
	if st.Disabled {
		bg = colors.DisabledBg
	}
	canvas.DrawRoundRect(st.Bounds, bg, m3DDTriggerRadius)
}

// m3DDPaintTriggerBorder draws the trigger outline.
func m3DDPaintTriggerBorder(canvas widget.Canvas, st *dropdown.TriggerPaintState, colors dropdown.DropdownColorScheme) {
	borderColor := colors.Border
	strokeWidth := m3DDBorderWidth

	switch {
	case st.Disabled:
		borderColor = colors.DisabledFg
	case st.Focused:
		borderColor = colors.FocusBorder
		strokeWidth = m3DDFocusBorderWidth
	case st.Hovered:
		borderColor = colors.TextColor
	}

	canvas.StrokeRoundRect(st.Bounds, borderColor, m3DDTriggerRadius, strokeWidth)
}

// m3DDPaintTriggerText draws the selected text or placeholder.
func m3DDPaintTriggerText(canvas widget.Canvas, st *dropdown.TriggerPaintState, colors dropdown.DropdownColorScheme) {
	textColor := colors.TextColor
	if st.IsPlaceholder {
		textColor = colors.PlaceholderText
	}
	if st.Disabled {
		textColor = colors.DisabledFg
	}

	textBounds := geometry.Rect{
		Min: geometry.Pt(st.Bounds.Min.X+m3DDContentPaddingH, st.Bounds.Min.Y),
		Max: geometry.Pt(st.Bounds.Max.X-m3DDChevronWidth-m3DDContentPaddingH, st.Bounds.Max.Y),
	}
	canvas.DrawText(st.SelectedText, textBounds, m3DDFontSize, textColor, false, m3DDTextAlignLeft)
}

// m3DDPaintTriggerChevron draws the chevron indicator.
func m3DDPaintTriggerChevron(canvas widget.Canvas, st *dropdown.TriggerPaintState, colors dropdown.DropdownColorScheme) {
	chevronColor := colors.ChevronColor
	if st.Disabled {
		chevronColor = colors.DisabledFg
	}

	chevronX := st.Bounds.Max.X - m3DDChevronWidth - m3DDContentPaddingH/2
	chevronY := st.Bounds.Center().Y
	m3DDDrawChevron(canvas, geometry.Pt(chevronX, chevronY), st.Open, chevronColor)
}

// m3DDDrawChevron draws an M3 up/down chevron indicator.
func m3DDDrawChevron(canvas widget.Canvas, center geometry.Point, open bool, color widget.Color) {
	size := m3DDChevronSize
	if open {
		// Up chevron.
		canvas.DrawLine(
			geometry.Pt(center.X-size, center.Y+size/2),
			geometry.Pt(center.X, center.Y-size/2),
			color, m3DDChevronStroke,
		)
		canvas.DrawLine(
			geometry.Pt(center.X, center.Y-size/2),
			geometry.Pt(center.X+size, center.Y+size/2),
			color, m3DDChevronStroke,
		)
	} else {
		// Down chevron.
		canvas.DrawLine(
			geometry.Pt(center.X-size, center.Y-size/2),
			geometry.Pt(center.X, center.Y+size/2),
			color, m3DDChevronStroke,
		)
		canvas.DrawLine(
			geometry.Pt(center.X, center.Y+size/2),
			geometry.Pt(center.X+size, center.Y-size/2),
			color, m3DDChevronStroke,
		)
	}
}

// m3DDPaintMenuItem draws a single menu item.
func m3DDPaintMenuItem(
	canvas widget.Canvas,
	itemRect geometry.Rect,
	item dropdown.ItemDef,
	index int,
	highlightedIndex int,
	selectedIndex int,
	colors dropdown.DropdownColorScheme,
) {
	// Highlight or selection background.
	switch index {
	case highlightedIndex:
		canvas.DrawRect(itemRect, colors.ItemHover)
	case selectedIndex:
		canvas.DrawRect(itemRect, colors.ItemSelected)
	}

	// Item text.
	textColor := colors.TextColor
	if item.Disabled {
		textColor = colors.ItemDisabled
	}

	textRect := geometry.Rect{
		Min: geometry.Pt(itemRect.Min.X+m3DDContentPaddingH, itemRect.Min.Y),
		Max: geometry.Pt(itemRect.Max.X-m3DDContentPaddingH, itemRect.Max.Y),
	}
	canvas.DrawText(item.DisplayText(), textRect, m3DDFontSize, textColor, false, m3DDTextAlignLeft)
}

// m3DDDrawFocusIndicator draws a focus ring around the trigger.
func m3DDDrawFocusIndicator(canvas widget.Canvas, bounds geometry.Rect, colors dropdown.DropdownColorScheme) {
	ringBounds := bounds.Expand(m3DDFocusRingOffset)
	ringRadius := m3DDTriggerRadius + m3DDFocusRingOffset
	canvas.StrokeRoundRect(ringBounds, colors.FocusRing, ringRadius, m3DDFocusRingStroke)
}

// m3DefaultDropdownColors holds the default M3 dropdown color scheme.
var m3DefaultDropdownColors = dropdown.DropdownColorScheme{
	Background:      widget.ColorWhite,                   // M3 surface
	Border:          widget.Hex(0x79747E),                // M3 outline
	FocusBorder:     widget.Hex(0x6750A4),                // M3 primary
	TextColor:       widget.Hex(0x1C1B1F),                // M3 on-surface
	PlaceholderText: widget.Hex(0x49454F),                // M3 on-surface-variant
	DisabledBg:      widget.RGBA(0.12, 0.12, 0.13, 0.04), // M3 disabled surface
	DisabledFg:      widget.RGBA(0.12, 0.12, 0.13, 0.38), // M3 disabled fg
	MenuBg:          widget.ColorWhite,                   // M3 surface
	MenuBorder:      widget.Hex(0xCAC4D0),                // M3 outline-variant
	ItemHover:       widget.RGBA(0.4, 0.31, 0.64, 0.08),  // M3 primary 8%
	ItemSelected:    widget.RGBA(0.4, 0.31, 0.64, 0.12),  // M3 primary 12%
	ItemDisabled:    widget.RGBA(0.12, 0.12, 0.13, 0.38), // M3 disabled fg
	ChevronColor:    widget.Hex(0x49454F),                // M3 on-surface-variant
	FocusRing:       widget.Hex(0x6750A4).WithAlpha(0.7), // M3 primary 70%
}

// M3 dropdown drawing constants.
const (
	m3DDTriggerRadius    float32 = 4
	m3DDMenuRadius       float32 = 4
	m3DDBorderWidth      float32 = 1
	m3DDFocusBorderWidth float32 = 2
	m3DDContentPaddingH  float32 = 16
	m3DDFontSize         float32 = 16
	m3DDTextAlignLeft    float32 = 0
	m3DDChevronWidth     float32 = 24
	m3DDChevronSize      float32 = 5
	m3DDChevronStroke    float32 = 1.5
	m3DDFocusRingOffset  float32 = 2
	m3DDFocusRingStroke  float32 = 2
)

// Compile-time check that DropdownPainter implements Painter.
var _ dropdown.Painter = DropdownPainter{}
