package material3

import (
	"github.com/gogpu/ui/core/menu"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// MenuPainter renders menus and menu bars using Material 3 design tokens.
// It maps M3 color roles to menu elements: surface for containers, elevation
// shadow for depth, on-surface for text, and primary for highlight states.
//
// If Theme is nil, MenuPainter falls back to the default M3 purple palette.
type MenuPainter struct {
	Theme *Theme // nil uses default M3 purple fallback
}

// resolveColors returns M3-derived colors for menu painting.
func (p MenuPainter) resolveColors() menu.MenuColorScheme {
	if p.Theme == nil {
		return m3DefaultMenuColors
	}
	cs := p.Theme.Colors
	return menu.MenuColorScheme{
		BarBackground:    cs.SurfaceContainer,
		BarText:          cs.OnSurface,
		BarHover:         cs.OnSurface.WithAlpha(0.04),
		BarActiveText:    cs.Primary,
		MenuBackground:   cs.SurfaceContainerLow,
		MenuBorder:       cs.OutlineVariant,
		ItemText:         cs.OnSurface,
		ItemHover:        cs.Primary.WithAlpha(0.08),
		ItemDisabledText: cs.OnSurface.WithAlpha(0.38),
		ShortcutText:     cs.OnSurfaceVariant,
		SeparatorColor:   cs.OutlineVariant,
		SubMenuArrow:     cs.OnSurfaceVariant,
	}
}

// PaintMenuBar renders a menu bar with M3 surface and elevation styling.
func (p MenuPainter) PaintMenuBar(canvas widget.Canvas, st *menu.MenuBarPaintState) {
	if st.Bounds.IsEmpty() {
		return
	}

	colors := p.effectiveBarColors(st.ColorScheme)

	// Bar background.
	canvas.DrawRect(st.Bounds, colors.BarBackground)

	// Bottom border.
	borderRect := geometry.Rect{
		Min: geometry.Pt(st.Bounds.Min.X, st.Bounds.Max.Y-m3MenuBorderWidth),
		Max: st.Bounds.Max,
	}
	canvas.DrawRect(borderRect, colors.MenuBorder)

	// Draw each top-level menu label.
	for i, m := range st.Menus {
		if i >= len(st.MenuRects) {
			break
		}
		r := st.MenuRects[i]

		// Highlight background.
		switch i {
		case st.OpenIndex:
			canvas.DrawRoundRect(r, colors.ItemHover, m3MenuBarItemRadius)
		case st.HoveredIndex:
			canvas.DrawRoundRect(r, colors.BarHover, m3MenuBarItemRadius)
		}

		// Label text.
		textColor := colors.BarText
		if i == st.OpenIndex {
			textColor = colors.BarActiveText
		}
		textBounds := geometry.Rect{
			Min: geometry.Pt(r.Min.X+m3MenuBarPaddingH, r.Min.Y),
			Max: geometry.Pt(r.Max.X-m3MenuBarPaddingH, r.Max.Y),
		}
		canvas.DrawText(m.Label, textBounds, m3MenuFontSize, textColor, false, widget.TextAlignCenter)
	}
}

// PaintMenu renders a popup menu panel with M3 surface and elevation styling.
func (p MenuPainter) PaintMenu(canvas widget.Canvas, st *menu.MenuPaintState) {
	if st.Bounds.IsEmpty() {
		return
	}

	colors := p.effectiveMenuColors(st.ColorScheme)

	// Elevation shadow.
	shadowRect := geometry.Rect{
		Min: geometry.Pt(st.Bounds.Min.X+m3MenuShadowOffset, st.Bounds.Min.Y+m3MenuShadowOffset),
		Max: geometry.Pt(st.Bounds.Max.X+m3MenuShadowOffset, st.Bounds.Max.Y+m3MenuShadowOffset),
	}
	canvas.DrawRoundRect(shadowRect, m3MenuShadowColor, m3MenuRadius)

	// Menu surface.
	canvas.DrawRoundRect(st.Bounds, colors.MenuBackground, m3MenuRadius)
	canvas.StrokeRoundRect(st.Bounds, colors.MenuBorder, m3MenuRadius, m3MenuBorderWidth)

	// Clip to menu bounds.
	canvas.PushClip(st.Bounds)
	defer canvas.PopClip()

	// Draw items.
	y := st.Bounds.Min.Y + m3MenuPaddingV
	for i, item := range st.Items {
		if item.IsSeparator() {
			m3PaintMenuSeparator(canvas, st.Bounds.Min.X, y, st.Bounds.Width(), st.SeparatorHeight, colors.SeparatorColor)
			y += st.SeparatorHeight
			continue
		}

		itemRect := geometry.Rect{
			Min: geometry.Pt(st.Bounds.Min.X, y),
			Max: geometry.Pt(st.Bounds.Max.X, y+st.ItemHeight),
		}

		// Highlight.
		if i == st.HighlightedIndex && !item.Disabled {
			canvas.DrawRoundRect(
				geometry.NewRect(itemRect.Min.X+m3MenuItemInset, itemRect.Min.Y, itemRect.Width()-m3MenuItemInset*2, st.ItemHeight),
				colors.ItemHover, m3MenuItemRadius,
			)
		}

		// Label.
		textColor := colors.ItemText
		if item.Disabled {
			textColor = colors.ItemDisabledText
		}
		labelRect := geometry.Rect{
			Min: geometry.Pt(itemRect.Min.X+m3MenuItemPaddingH, itemRect.Min.Y),
			Max: geometry.Pt(itemRect.Max.X-m3MenuShortcutWidth-m3MenuItemPaddingH, itemRect.Max.Y),
		}
		canvas.DrawText(item.Label, labelRect, m3MenuFontSize, textColor, false, widget.TextAlignLeft)

		// Shortcut text or submenu arrow.
		rightRect := geometry.Rect{
			Min: geometry.Pt(itemRect.Max.X-m3MenuShortcutWidth, itemRect.Min.Y),
			Max: geometry.Pt(itemRect.Max.X-m3MenuItemPaddingH, itemRect.Max.Y),
		}
		if item.HasChildren() {
			arrowColor := colors.SubMenuArrow
			if item.Disabled {
				arrowColor = colors.ItemDisabledText
			}
			canvas.DrawText(m3MenuRightArrow, rightRect, m3MenuFontSize, arrowColor, false, widget.TextAlignRight)
		} else if item.Shortcut != "" {
			shortcutColor := colors.ShortcutText
			if item.Disabled {
				shortcutColor = colors.ItemDisabledText
			}
			canvas.DrawText(item.Shortcut, rightRect, m3MenuShortcutFontSize, shortcutColor, false, widget.TextAlignRight)
		}

		y += st.ItemHeight
	}
}

// effectiveBarColors returns colors for bar painting, preferring state's ColorScheme.
func (p MenuPainter) effectiveBarColors(cs menu.MenuColorScheme) menu.MenuColorScheme {
	if cs != (menu.MenuColorScheme{}) {
		return cs
	}
	return p.resolveColors()
}

// effectiveMenuColors returns colors for menu painting, preferring state's ColorScheme.
func (p MenuPainter) effectiveMenuColors(cs menu.MenuColorScheme) menu.MenuColorScheme {
	if cs != (menu.MenuColorScheme{}) {
		return cs
	}
	return p.resolveColors()
}

// m3PaintMenuSeparator draws a horizontal separator line.
func m3PaintMenuSeparator(canvas widget.Canvas, x, y, width, height float32, color widget.Color) {
	lineY := y + height/2
	canvas.DrawLine(
		geometry.Pt(x+m3MenuSepPaddingH, lineY),
		geometry.Pt(x+width-m3MenuSepPaddingH, lineY),
		color, m3MenuSepWidth,
	)
}

// m3DefaultMenuColors holds default M3 purple fallback colors for menus.
var m3DefaultMenuColors = menu.MenuColorScheme{
	BarBackground:    widget.Hex(0xECE6F0), // M3 surface-container
	BarText:          widget.Hex(0x1C1B1F), // M3 on-surface
	BarHover:         widget.RGBA(0.12, 0.12, 0.13, 0.04),
	BarActiveText:    widget.Hex(0x6750A4), // M3 primary
	MenuBackground:   widget.Hex(0xF7F2FA), // M3 surface-container-low
	MenuBorder:       widget.Hex(0xCAC4D0), // M3 outline-variant
	ItemText:         widget.Hex(0x1C1B1F), // M3 on-surface
	ItemHover:        widget.Hex(0x6750A4).WithAlpha(0.08),
	ItemDisabledText: widget.RGBA(0.12, 0.12, 0.13, 0.38),
	ShortcutText:     widget.Hex(0x49454F), // M3 on-surface-variant
	SeparatorColor:   widget.Hex(0xCAC4D0), // M3 outline-variant
	SubMenuArrow:     widget.Hex(0x49454F), // M3 on-surface-variant
}

// M3 menu drawing constants.
const (
	m3MenuRadius           float32 = 4
	m3MenuBorderWidth      float32 = 1
	m3MenuFontSize         float32 = 14
	m3MenuShortcutFontSize float32 = 12
	m3MenuBarPaddingH      float32 = 12
	m3MenuBarItemRadius    float32 = 4
	m3MenuItemPaddingH     float32 = 16
	m3MenuItemInset        float32 = 4
	m3MenuItemRadius       float32 = 4
	m3MenuPaddingV         float32 = 4
	m3MenuShortcutWidth    float32 = 80
	m3MenuSepPaddingH      float32 = 12
	m3MenuSepWidth         float32 = 1
	m3MenuShadowOffset     float32 = 2
	m3MenuRightArrow               = ">"
)

// m3MenuShadowColor is the M3 elevation shadow color for popup menus.
var m3MenuShadowColor = widget.RGBA(0, 0, 0, 0.12)

// Compile-time check that MenuPainter implements Painter.
var _ menu.Painter = MenuPainter{}
