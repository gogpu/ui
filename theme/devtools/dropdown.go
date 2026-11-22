package devtools

import (
	"github.com/gogpu/ui/core/dropdown"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/widget"
)

// DropdownPainter renders dropdowns using DevTools design tokens.
// DevTools dropdowns use the same InputBackground as text fields,
// with SurfaceElevated popup menus and compact spacing.
//
// If Theme is nil, DropdownPainter falls back to the default DevTools dark palette.
type DropdownPainter struct {
	Theme *Theme // nil uses default DevTools dark fallback
}

// resolveColors returns the DropdownColorScheme derived from the painter's Theme.
func (p DropdownPainter) resolveColors() dropdown.DropdownColorScheme {
	if p.Theme == nil {
		return dtDefaultDropdownColors
	}
	cs := p.Theme.Colors
	return dropdown.DropdownColorScheme{
		Background:      cs.InputBackground,
		Border:          cs.BorderStrong,
		FocusBorder:     cs.BorderFocus,
		TextColor:       cs.OnSurface,
		PlaceholderText: cs.OnSurfaceSecondary,
		DisabledBg:      cs.ControlFill,
		DisabledFg:      cs.OnSurfaceDisabled,
		MenuBg:          cs.SurfaceElevated,
		MenuBorder:      cs.BorderStrong,
		ItemHover:       cs.ControlHover,
		ItemSelected:    cs.Selection,
		ItemDisabled:    cs.OnSurfaceDisabled,
		ChevronColor:    cs.OnSurfaceSecondary,
		FocusRing:       cs.BorderFocus,
	}
}

// PaintTrigger renders a dropdown trigger according to DevTools design specifications.
func (p DropdownPainter) PaintTrigger(canvas widget.Canvas, st *dropdown.TriggerPaintState) {
	if st.Bounds.IsEmpty() {
		return
	}

	colors := st.ColorScheme
	if colors == (dropdown.DropdownColorScheme{}) {
		colors = p.resolveColors()
	}

	dtDDPaintTriggerBg(canvas, st, colors)
	dtDDPaintTriggerBorder(canvas, st, colors)
	dtDDPaintTriggerText(canvas, st, colors)
	dtDDPaintTriggerChevron(canvas, st, colors)

	if st.Focused && !st.Disabled {
		dtDrawFocusRing(canvas, st.Bounds, dtDDTriggerRadius, colors.FocusRing)
	}
}

// PaintMenu renders a dropdown menu according to DevTools design specifications.
func (p DropdownPainter) PaintMenu(canvas widget.Canvas, st *dropdown.MenuPaintState) {
	if st.Bounds.IsEmpty() {
		return
	}

	colors := st.ColorScheme
	if colors == (dropdown.DropdownColorScheme{}) {
		colors = p.resolveColors()
	}

	// Subtle elevation shadow.
	shadow := geometry.Rect{
		Min: geometry.Pt(st.Bounds.Min.X, st.Bounds.Min.Y+2),
		Max: geometry.Pt(st.Bounds.Max.X, st.Bounds.Max.Y+3),
	}
	shadowColor := widget.RGBA(0, 0, 0, 0.28)
	if p.Theme != nil {
		shadowColor = p.Theme.Colors.Shadow.WithAlpha(min(p.Theme.Colors.Shadow.A, 0.28))
	}
	canvas.DrawRoundRect(shadow, shadowColor, dtDDMenuRadius)

	// Menu background with border.
	canvas.DrawRoundRect(st.Bounds, colors.MenuBg, dtDDMenuRadius)
	canvas.StrokeRoundRect(st.Bounds, colors.MenuBorder, dtDDMenuRadius, dtDDBorderWidth)

	canvas.PushClip(st.Bounds)
	defer canvas.PopClip()

	endIndex := min(st.ScrollOffset+st.VisibleCount, len(st.Items))

	for i := st.ScrollOffset; i < endIndex; i++ {
		item := st.Items[i]
		row := i - st.ScrollOffset
		top := st.Bounds.Min.Y + st.MenuPaddingV + float32(row)*st.ItemHeight
		itemRect := geometry.Rect{
			Min: geometry.Pt(st.Bounds.Min.X, top),
			Max: geometry.Pt(st.Bounds.Max.X, top+st.ItemHeight),
		}

		dtDDPaintMenuItem(canvas, itemRect, item, i, st.HighlightedIndex, st.SelectedIndex, colors)
	}
}

// dtDDPaintTriggerBg draws the trigger background.
func dtDDPaintTriggerBg(canvas widget.Canvas, st *dropdown.TriggerPaintState, colors dropdown.DropdownColorScheme) {
	bg := colors.Background
	if st.Disabled {
		bg = colors.DisabledBg
	}
	canvas.DrawRoundRect(st.Bounds, bg, dtDDTriggerRadius)
}

// dtDDPaintTriggerBorder draws the trigger outline.
func dtDDPaintTriggerBorder(canvas widget.Canvas, st *dropdown.TriggerPaintState, colors dropdown.DropdownColorScheme) {
	borderColor := colors.Border
	strokeWidth := dtDDBorderWidth

	switch {
	case st.Disabled:
		borderColor = colors.DisabledFg
	case st.Focused:
		borderColor = colors.FocusBorder
		strokeWidth = dtDDBorderWidth
	case st.Hovered:
		borderColor = colors.TextColor
	}

	canvas.StrokeRoundRect(st.Bounds, borderColor, dtDDTriggerRadius, strokeWidth)
}

// dtDDPaintTriggerText draws the selected text or placeholder.
func dtDDPaintTriggerText(canvas widget.Canvas, st *dropdown.TriggerPaintState, colors dropdown.DropdownColorScheme) {
	textColor := colors.TextColor
	if st.IsPlaceholder {
		textColor = colors.PlaceholderText
	}
	if st.Disabled {
		textColor = colors.DisabledFg
	}

	textBounds := geometry.Rect{
		Min: geometry.Pt(st.Bounds.Min.X+dtDDContentPaddingH, st.Bounds.Min.Y),
		Max: geometry.Pt(st.Bounds.Max.X-dtDDChevronWidth-dtDDContentPaddingH, st.Bounds.Max.Y),
	}
	if !textBounds.IsEmpty() {
		canvas.PushClip(textBounds)
		canvas.DrawText(st.SelectedText, textBounds, dtDDFontSize, textColor, false, dtDDTextAlignLeft)
		canvas.PopClip()
	}
}

// dtDDPaintTriggerChevron draws the chevron indicator.
func dtDDPaintTriggerChevron(canvas widget.Canvas, st *dropdown.TriggerPaintState, colors dropdown.DropdownColorScheme) {
	chevronColor := colors.ChevronColor
	if st.Disabled {
		chevronColor = colors.DisabledFg
	}

	glyph := icon.ExpChevronDown
	if st.Open {
		glyph = icon.ExpChevronUp
	}
	centerY := st.Bounds.Center().Y
	chevronRect := geometry.NewRect(st.Bounds.Max.X-19, centerY-7, 14, 14)
	icon.Draw(canvas, glyph, chevronRect, chevronColor)
}

// dtDDPaintMenuItem draws a single menu item.
func dtDDPaintMenuItem(
	canvas widget.Canvas,
	itemRect geometry.Rect,
	item dropdown.ItemDef,
	index int,
	highlightedIndex int,
	selectedIndex int,
	colors dropdown.DropdownColorScheme,
) {
	fillRect := geometry.Rect{
		Min: geometry.Pt(itemRect.Min.X+dtDDItemInsetH, itemRect.Min.Y),
		Max: geometry.Pt(itemRect.Max.X-dtDDItemInsetH, itemRect.Max.Y),
	}
	switch index {
	case highlightedIndex:
		canvas.DrawRoundRect(fillRect, colors.ItemHover, dtDDItemRadius)
	case selectedIndex:
		canvas.DrawRoundRect(fillRect, colors.ItemSelected, dtDDItemRadius)
	}

	textColor := colors.TextColor
	if item.Disabled {
		textColor = colors.ItemDisabled
	}

	textRect := geometry.Rect{
		Min: geometry.Pt(itemRect.Min.X+dtDDContentPaddingH, itemRect.Min.Y),
		Max: geometry.Pt(itemRect.Max.X-dtDDContentPaddingH, itemRect.Max.Y),
	}
	if !textRect.IsEmpty() {
		canvas.PushClip(textRect)
		canvas.DrawText(item.DisplayText(), textRect, dtDDFontSize, textColor, false, dtDDTextAlignLeft)
		canvas.PopClip()
	}
}

// dtDefaultDropdownColors holds the default DevTools dark dropdown color scheme.
var dtDefaultDropdownColors = dropdown.DropdownColorScheme{
	Background:      widget.Hex(0x1E1F22), // Gray1 (InputBackground)
	Border:          widget.Hex(0x4E5157), // Gray5
	FocusBorder:     DefaultAccentColor,
	TextColor:       widget.Hex(0xDFE1E5), // Gray12
	PlaceholderText: widget.Hex(0x9DA0A8), // Gray9
	DisabledBg:      widget.Hex(0x393B40), // Gray3
	DisabledFg:      widget.Hex(0x6F737A), // Gray7
	MenuBg:          widget.Hex(0x393B40), // Gray3 (SurfaceElevated)
	MenuBorder:      widget.Hex(0x4E5157), // Gray5
	ItemHover:       widget.Hex(0x43454A), // Gray4 (ControlHover)
	ItemSelected:    widget.Hex(0x2E436E), // Blue2 (Selection)
	ItemDisabled:    widget.Hex(0x6F737A), // Gray7
	ChevronColor:    widget.Hex(0x9DA0A8), // Gray9
	FocusRing:       DefaultAccentColor,
}

// TriggerHeight returns the compact DevTools dropdown trigger height.
func (DropdownPainter) TriggerHeight() float32 { return dtDDTriggerHeight }

// MenuItemHeight returns the compact DevTools dropdown menu item height.
func (DropdownPainter) MenuItemHeight() float32 { return dtDDMenuItemHeight }

// MenuPaddingV returns the top and bottom vertical padding for DevTools dropdown menus.
func (DropdownPainter) MenuPaddingV() float32 { return dtDDMenuPaddingV }

// DevTools dropdown drawing constants.
const (
	dtDDTriggerHeight   float32 = 28
	dtDDMenuItemHeight  float32 = 28
	dtDDMenuPaddingV    float32 = 4
	dtDDTriggerRadius   float32 = 4
	dtDDMenuRadius      float32 = 6
	dtDDItemRadius      float32 = 5
	dtDDItemInsetH      float32 = 4
	dtDDBorderWidth     float32 = 1
	dtDDContentPaddingH float32 = 8
	dtDDFontSize        float32 = 13
	dtDDTextAlignLeft           = widget.TextAlignLeft
	dtDDChevronWidth    float32 = 20
)

// Compile-time checks that DropdownPainter implements Painter and LayoutMetrics.
var (
	_ dropdown.Painter       = DropdownPainter{}
	_ dropdown.LayoutMetrics = DropdownPainter{}
)
