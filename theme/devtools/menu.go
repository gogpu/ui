package devtools

import (
	"slices"

	"github.com/gogpu/ui/core/menu"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/widget"
)

// MenuPainter renders menus and menu bars using DevTools design tokens.
// DevTools menus use a dark popup (SurfaceElevated #393B40) with 8px radius,
// 1px border, 28px item height, right-aligned keyboard shortcut text in Gray9,
// and a ControlHover (#43454A) highlight, matching JetBrains IDE menu styling.
//
// If Theme is nil, MenuPainter falls back to the default DevTools dark palette.
type MenuPainter struct {
	Theme *Theme // nil uses default DevTools dark fallback
}

// resolveColors returns DevTools-derived colors for menu painting.
func (p MenuPainter) resolveColors() menu.MenuColorScheme {
	if p.Theme == nil {
		return dtDefaultMenuColors
	}
	cs := p.Theme.Colors
	return menu.MenuColorScheme{
		BarBackground:    cs.Surface,
		BarText:          cs.OnSurface,
		BarHover:         cs.ControlHover,
		BarActiveText:    cs.Primary,
		MenuBackground:   cs.SurfaceElevated,
		MenuBorder:       cs.BorderStrong,
		ItemText:         cs.OnSurface,
		ItemHover:        cs.ControlHover,
		ItemDisabledText: cs.OnSurfaceDisabled,
		ShortcutText:     cs.OnSurfaceSecondary,
		SeparatorColor:   cs.BorderStrong,
		SubMenuArrow:     cs.OnSurfaceSecondary,
	}
}

// PaintMenuBar renders a menu bar with DevTools surface and border styling.
func (p MenuPainter) PaintMenuBar(canvas widget.Canvas, st *menu.MenuBarPaintState) {
	if st.Bounds.IsEmpty() {
		return
	}

	colors := p.effectiveBarColors(st.ColorScheme)

	// Bar background.
	canvas.DrawRect(st.Bounds, colors.BarBackground)

	// Bottom border.
	borderRect := geometry.Rect{
		Min: geometry.Pt(st.Bounds.Min.X, st.Bounds.Max.Y-dtMenuBorderWidth),
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
			canvas.DrawRect(r, colors.ItemHover)
		case st.HoveredIndex:
			canvas.DrawRect(r, colors.BarHover)
		}

		// Label text.
		textColor := colors.BarText
		if i == st.OpenIndex {
			textColor = colors.BarActiveText
		}
		textBounds := geometry.Rect{
			Min: geometry.Pt(r.Min.X+dtMenuBarPaddingH, r.Min.Y),
			Max: geometry.Pt(r.Max.X-dtMenuBarPaddingH, r.Max.Y),
		}
		canvas.DrawText(m.Label, textBounds, dtMenuFontSize, textColor, false, widget.TextAlignCenter)
	}
}

// PaintMenu renders a popup menu panel with DevTools elevated surface styling.
func (p MenuPainter) PaintMenu(canvas widget.Canvas, st *menu.MenuPaintState) {
	if st.Bounds.IsEmpty() {
		return
	}

	colors := p.effectiveMenuColors(st.ColorScheme)

	// Elevation shadow.
	shadowRect := geometry.Rect{
		Min: geometry.Pt(st.Bounds.Min.X+dtMenuShadowOffset, st.Bounds.Min.Y+dtMenuShadowOffset),
		Max: geometry.Pt(st.Bounds.Max.X+dtMenuShadowOffset, st.Bounds.Max.Y+dtMenuShadowOffset),
	}
	shadowColor := dtMenuShadowColor
	if p.Theme != nil {
		shadowColor = p.Theme.Colors.Shadow.WithAlpha(min(p.Theme.Colors.Shadow.A, 0.28))
	}
	canvas.DrawRoundRect(shadowRect, shadowColor, dtMenuRadius)

	// Menu surface.
	canvas.DrawRoundRect(st.Bounds, colors.MenuBackground, dtMenuRadius)
	canvas.StrokeRoundRect(st.Bounds, colors.MenuBorder, dtMenuRadius, dtMenuBorderWidth)

	// Clip to menu bounds.
	canvas.PushClip(st.Bounds)
	defer canvas.PopClip()

	// Draw items.
	padV := st.PaddingV
	if padV <= 0 {
		padV = dtMenuPaddingV
	}
	hasCheck := slices.ContainsFunc(st.Items, func(it menu.MenuItem) bool {
		return it.Checkable || it.Checked
	})
	y := st.Bounds.Min.Y + padV
	for i, item := range st.Items {
		if item.IsSeparator() {
			dtPaintMenuSeparator(canvas, st.Bounds.Min.X, y, st.Bounds.Width(), st.SeparatorHeight, colors.SeparatorColor)
			y += st.SeparatorHeight
			continue
		}

		itemRect := geometry.Rect{
			Min: geometry.Pt(st.Bounds.Min.X, y),
			Max: geometry.Pt(st.Bounds.Max.X, y+st.ItemHeight),
		}

		// Highlight.
		if i == st.HighlightedIndex && !item.Disabled {
			fillRect := geometry.Rect{
				Min: geometry.Pt(itemRect.Min.X+dtMenuItemInsetH, itemRect.Min.Y),
				Max: geometry.Pt(itemRect.Max.X-dtMenuItemInsetH, itemRect.Max.Y),
			}
			canvas.DrawRoundRect(fillRect, colors.ItemHover, dtMenuItemRadius)
		}

		// Label and optional check indicator.
		textColor := colors.ItemText
		if item.Disabled {
			textColor = colors.ItemDisabledText
		}
		labelMinX := itemRect.Min.X + dtMenuItemPaddingH
		if hasCheck {
			if item.Checked {
				checkRect := geometry.NewRect(itemRect.Min.X+dtMenuItemPaddingH, itemRect.Center().Y-7, 14, 14)
				icon.Draw(canvas, icon.ExpCheck, checkRect, textColor)
			}
			labelMinX += dtMenuCheckGutter
		}
		labelMaxX := itemRect.Max.X - dtMenuItemPaddingH
		if item.HasChildren() || item.Shortcut != "" {
			labelMaxX -= dtMenuShortcutWidth
		}
		labelRect := geometry.Rect{
			Min: geometry.Pt(labelMinX, itemRect.Min.Y),
			Max: geometry.Pt(max(labelMinX, labelMaxX), itemRect.Max.Y),
		}
		if !labelRect.IsEmpty() {
			canvas.PushClip(labelRect)
			canvas.DrawText(item.Label, labelRect, dtMenuFontSize, textColor, false, widget.TextAlignLeft)
			canvas.PopClip()
		}

		// Shortcut text or submenu arrow.
		rightRect := geometry.Rect{
			Min: geometry.Pt(itemRect.Max.X-dtMenuShortcutWidth, itemRect.Min.Y),
			Max: geometry.Pt(itemRect.Max.X-dtMenuItemPaddingH, itemRect.Max.Y),
		}
		if item.HasChildren() {
			arrowColor := colors.SubMenuArrow
			if item.Disabled {
				arrowColor = colors.ItemDisabledText
			}
			arrowRect := geometry.NewRect(itemRect.Max.X-dtMenuItemPaddingH-12, itemRect.Center().Y-6, 12, 12)
			icon.Draw(canvas, icon.ExpChevronRight, arrowRect, arrowColor)
		} else if item.Shortcut != "" {
			shortcutColor := colors.ShortcutText
			if item.Disabled {
				shortcutColor = colors.ItemDisabledText
			}
			canvas.DrawText(item.Shortcut, rightRect, dtMenuShortcutFontSize, shortcutColor, false, widget.TextAlignRight)
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

// dtPaintMenuSeparator draws a horizontal separator line.
func dtPaintMenuSeparator(canvas widget.Canvas, x, y, width, height float32, color widget.Color) {
	lineY := y + height/2
	canvas.DrawLine(
		geometry.Pt(x+dtMenuSepPaddingH, lineY),
		geometry.Pt(x+width-dtMenuSepPaddingH, lineY),
		color, dtMenuSepWidth,
	)
}

// dtDefaultMenuColors holds default DevTools dark fallback colors for menus.
var dtDefaultMenuColors = menu.MenuColorScheme{
	BarBackground:    widget.Hex(0x2B2D30), // Gray2 (surface)
	BarText:          widget.Hex(0xDFE1E5), // Gray12 (on-surface)
	BarHover:         widget.Hex(0x43454A), // Gray4 (control hover)
	BarActiveText:    widget.Hex(0x3574F0), // Blue6 (primary)
	MenuBackground:   widget.Hex(0x393B40), // Gray3 (elevated surface)
	MenuBorder:       widget.Hex(0x4E5157), // Gray5 (border strong)
	ItemText:         widget.Hex(0xDFE1E5), // Gray12 (on-surface)
	ItemHover:        widget.Hex(0x43454A), // Gray4 (control hover)
	ItemDisabledText: widget.Hex(0x6F737A), // Gray7 (disabled)
	ShortcutText:     widget.Hex(0x9DA0A8), // Gray9 (secondary text)
	SeparatorColor:   widget.Hex(0x4E5157), // Gray5 (border strong)
	SubMenuArrow:     widget.Hex(0x9DA0A8), // Gray9 (secondary)
}

// MenuItemHeight returns the compact DevTools popup menu item height (28px).
func (MenuPainter) MenuItemHeight() float32 { return dtMenuItemHeight }

// MenuSeparatorHeight returns the DevTools popup menu separator height (9px).
func (MenuPainter) MenuSeparatorHeight() float32 { return dtMenuSepHeight }

// MenuPaddingV returns the top and bottom popup menu padding (4px).
func (MenuPainter) MenuPaddingV() float32 { return dtMenuPaddingV }

// DevTools menu drawing constants.
const (
	dtMenuItemHeight       float32 = 28
	dtMenuSepHeight        float32 = 9
	dtMenuRadius           float32 = 6
	dtMenuItemRadius       float32 = 5
	dtMenuItemInsetH       float32 = 4
	dtMenuCheckGutter      float32 = 20
	dtMenuBorderWidth      float32 = 1
	dtMenuFontSize         float32 = 13
	dtMenuShortcutFontSize float32 = 11
	dtMenuBarPaddingH      float32 = 8
	dtMenuItemPaddingH     float32 = 12
	dtMenuPaddingV         float32 = 4
	dtMenuShortcutWidth    float32 = 80
	dtMenuSepPaddingH      float32 = 8
	dtMenuSepWidth         float32 = 1
	dtMenuShadowOffset     float32 = 2
)

// dtMenuShadowColor is the DevTools elevation shadow color for popup menus.
var dtMenuShadowColor = widget.RGBA(0, 0, 0, 0.28)

// Compile-time checks that MenuPainter implements Painter and LayoutMetrics.
var (
	_ menu.Painter       = MenuPainter{}
	_ menu.LayoutMetrics = MenuPainter{}
)
