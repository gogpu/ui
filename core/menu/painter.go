package menu

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// Painter draws the visual representation of menus and menu bars.
// Each design system (Material 3, Fluent, Cupertino) provides its own
// Painter implementation. If no Painter is set, [DefaultPainter] is used.
type Painter interface {
	// PaintMenuBar draws the horizontal menu bar background and top-level labels.
	PaintMenuBar(canvas widget.Canvas, state *MenuBarPaintState)

	// PaintMenu draws a popup menu panel with its items.
	PaintMenu(canvas widget.Canvas, state *MenuPaintState)
}

// MenuBarPaintState provides the current menu bar state to the painter.
type MenuBarPaintState struct {
	// Bounds is the menu bar's bounds.
	Bounds geometry.Rect

	// Menus is the list of top-level menu definitions.
	Menus []TopMenu

	// MenuRects contains the bounding rectangles for each top-level menu label.
	MenuRects []geometry.Rect

	// OpenIndex is the index of the currently open top-level menu (-1 for none).
	OpenIndex int

	// HoveredIndex is the index of the hovered top-level menu label (-1 for none).
	HoveredIndex int

	// Focused is true if the menu bar has keyboard focus.
	Focused bool

	// ColorScheme provides theme-derived colors. Zero value means use defaults.
	ColorScheme MenuColorScheme
}

// MenuPaintState provides the current popup menu state to the painter.
type MenuPaintState struct {
	// Bounds is the menu panel's bounds in window coordinates.
	Bounds geometry.Rect

	// Items is the list of menu items.
	Items []MenuItem

	// HighlightedIndex is the index of the currently highlighted item (-1 for none).
	HighlightedIndex int

	// ItemHeight is the height of each regular item row.
	ItemHeight float32

	// SeparatorHeight is the height of separator rows.
	SeparatorHeight float32

	// SubMenuOpenIndex is the index of the item whose submenu is open (-1 for none).
	SubMenuOpenIndex int

	// ColorScheme provides theme-derived colors. Zero value means use defaults.
	ColorScheme MenuColorScheme
}

// MenuColorScheme provides theme-derived colors for menu painting.
// Zero value means the painter should use its built-in defaults.
type MenuColorScheme struct {
	BarBackground    widget.Color
	BarText          widget.Color
	BarHover         widget.Color
	BarActiveText    widget.Color
	MenuBackground   widget.Color
	MenuBorder       widget.Color
	ItemText         widget.Color
	ItemHover        widget.Color
	ItemDisabledText widget.Color
	ShortcutText     widget.Color
	SeparatorColor   widget.Color
	SubMenuArrow     widget.Color
}

// DefaultPainter provides a minimal fallback painter for menus.
// It draws simple rectangles and text without design system styling.
type DefaultPainter struct{}

// PaintMenuBar renders a minimal menu bar.
func (p DefaultPainter) PaintMenuBar(canvas widget.Canvas, st *MenuBarPaintState) {
	if st.Bounds.IsEmpty() {
		return
	}

	// Bar background.
	canvas.DrawRect(st.Bounds, dfltBarBg)

	// Bottom border line.
	borderRect := geometry.Rect{
		Min: geometry.Pt(st.Bounds.Min.X, st.Bounds.Max.Y-1),
		Max: st.Bounds.Max,
	}
	canvas.DrawRect(borderRect, dfltBarBorder)

	// Draw each top-level menu label.
	for i, m := range st.Menus {
		if i >= len(st.MenuRects) {
			break
		}
		r := st.MenuRects[i]

		// Highlight background.
		switch i {
		case st.OpenIndex:
			canvas.DrawRect(r, dfltBarActiveItem)
		case st.HoveredIndex:
			canvas.DrawRect(r, dfltBarHoverItem)
		}

		// Label text.
		textColor := dfltBarTextColor
		if i == st.OpenIndex {
			textColor = dfltBarActiveText
		}
		textBounds := geometry.Rect{
			Min: geometry.Pt(r.Min.X+dfltBarPaddingH, r.Min.Y),
			Max: geometry.Pt(r.Max.X-dfltBarPaddingH, r.Max.Y),
		}
		canvas.DrawText(m.Label, textBounds, dfltFontSize, textColor, false, widget.TextAlignCenter)
	}
}

// PaintMenu renders a minimal popup menu panel.
func (p DefaultPainter) PaintMenu(canvas widget.Canvas, st *MenuPaintState) {
	if st.Bounds.IsEmpty() {
		return
	}

	// Menu background with shadow.
	shadowRect := geometry.Rect{
		Min: geometry.Pt(st.Bounds.Min.X+dfltShadowOffset, st.Bounds.Min.Y+dfltShadowOffset),
		Max: geometry.Pt(st.Bounds.Max.X+dfltShadowOffset, st.Bounds.Max.Y+dfltShadowOffset),
	}
	canvas.DrawRoundRect(shadowRect, dfltShadowColor, dfltMenuRadius)
	canvas.DrawRoundRect(st.Bounds, dfltMenuBg, dfltMenuRadius)
	canvas.StrokeRoundRect(st.Bounds, dfltMenuBorder, dfltMenuRadius, dfltBorderWidth)

	// Clip to menu bounds.
	canvas.PushClip(st.Bounds)
	defer canvas.PopClip()

	// Draw items.
	y := st.Bounds.Min.Y + dfltMenuPaddingV
	for i, item := range st.Items {
		if item.IsSeparator() {
			paintSeparator(canvas, st.Bounds.Min.X, y, st.Bounds.Width(), st.SeparatorHeight)
			y += st.SeparatorHeight
			continue
		}

		itemRect := geometry.Rect{
			Min: geometry.Pt(st.Bounds.Min.X, y),
			Max: geometry.Pt(st.Bounds.Max.X, y+st.ItemHeight),
		}

		// Highlight.
		if i == st.HighlightedIndex && !item.Disabled {
			canvas.DrawRect(itemRect, dfltItemHover)
		}

		// Label.
		textColor := dfltItemText
		if item.Disabled {
			textColor = dfltDisabledText
		}
		labelRect := geometry.Rect{
			Min: geometry.Pt(itemRect.Min.X+dfltMenuItemPaddingH, itemRect.Min.Y),
			Max: geometry.Pt(itemRect.Max.X-dfltShortcutAreaWidth-dfltMenuItemPaddingH, itemRect.Max.Y),
		}
		canvas.DrawText(item.Label, labelRect, dfltFontSize, textColor, false, widget.TextAlignLeft)

		// Shortcut text or submenu arrow.
		rightRect := geometry.Rect{
			Min: geometry.Pt(itemRect.Max.X-dfltShortcutAreaWidth, itemRect.Min.Y),
			Max: geometry.Pt(itemRect.Max.X-dfltMenuItemPaddingH, itemRect.Max.Y),
		}
		if item.HasChildren() {
			arrowColor := dfltSubMenuArrow
			if item.Disabled {
				arrowColor = dfltDisabledText
			}
			canvas.DrawText(rightArrowStr, rightRect, dfltFontSize, arrowColor, false, widget.TextAlignRight)
		} else if item.Shortcut != "" {
			shortcutColor := dfltShortcutText
			if item.Disabled {
				shortcutColor = dfltDisabledText
			}
			canvas.DrawText(item.Shortcut, rightRect, dfltShortcutFontSize, shortcutColor, false, widget.TextAlignRight)
		}

		y += st.ItemHeight
	}
}

// paintSeparator draws a horizontal line separator.
func paintSeparator(canvas widget.Canvas, x, y, width, height float32) {
	lineY := y + height/2
	canvas.DrawLine(
		geometry.Pt(x+dfltSepPaddingH, lineY),
		geometry.Pt(x+width-dfltSepPaddingH, lineY),
		dfltSepColor, dfltSepWidth,
	)
}

// rightArrowStr is a right-pointing arrow for submenu indicators.
const rightArrowStr = ">"

// Default painter constants.
var (
	dfltBarBg         = widget.Hex(0xF5F5F5)
	dfltBarBorder     = widget.Hex(0xE0E0E0)
	dfltBarTextColor  = widget.Hex(0x1C1B1F)
	dfltBarActiveText = widget.Hex(0x6750A4)
	dfltBarHoverItem  = widget.RGBA(0, 0, 0, 0.04)
	dfltBarActiveItem = widget.RGBA(0.4, 0.31, 0.64, 0.08)
	dfltMenuBg        = widget.ColorWhite
	dfltMenuBorder    = widget.Hex(0xCAC4D0)
	dfltItemText      = widget.Hex(0x1C1B1F)
	dfltItemHover     = widget.RGBA(0.4, 0.31, 0.64, 0.08)
	dfltDisabledText  = widget.RGBA(0.12, 0.12, 0.13, 0.38)
	dfltShortcutText  = widget.Hex(0x79747E)
	dfltSubMenuArrow  = widget.Hex(0x49454F)
	dfltSepColor      = widget.Hex(0xE0E0E0)
	dfltShadowColor   = widget.RGBA(0, 0, 0, 0.12)
)

const (
	dfltBarHeight         float32 = 32
	dfltBarPaddingH       float32 = 12
	dfltMenuRadius        float32 = 4
	dfltBorderWidth       float32 = 1
	dfltFontSize          float32 = 14
	dfltShortcutFontSize  float32 = 12
	dfltMenuItemPaddingH  float32 = 16
	dfltMenuPaddingV      float32 = 4
	dfltShortcutAreaWidth float32 = 80
	dfltSepPaddingH       float32 = 12
	dfltSepWidth          float32 = 1
	dfltShadowOffset      float32 = 2
)

// Compile-time check.
var _ Painter = DefaultPainter{}
