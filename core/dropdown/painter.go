package dropdown

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// Painter draws the visual representation of a dropdown.
// Each design system (Material 3, Fluent, Cupertino) provides its own
// Painter implementation. If no Painter is set, [DefaultPainter] is used.
type Painter interface {
	// PaintTrigger draws the closed dropdown trigger (shows current selection).
	PaintTrigger(canvas widget.Canvas, state *TriggerPaintState)

	// PaintMenu draws the open dropdown menu.
	PaintMenu(canvas widget.Canvas, state *MenuPaintState)
}

// TriggerPaintState provides the current dropdown trigger state to the painter.
type TriggerPaintState struct {
	// Bounds is the trigger widget's bounds.
	Bounds geometry.Rect

	// SelectedText is the display text of the currently selected item,
	// or the placeholder if nothing is selected.
	SelectedText string

	// IsPlaceholder is true if SelectedText is the placeholder (nothing selected).
	IsPlaceholder bool

	// Open is true if the dropdown menu is currently visible.
	Open bool

	// Focused is true if the dropdown has keyboard focus.
	Focused bool

	// Hovered is true if the mouse is over the trigger.
	Hovered bool

	// Disabled is true if the dropdown is disabled.
	Disabled bool

	// ColorScheme provides theme-derived colors. Zero value means use defaults.
	ColorScheme DropdownColorScheme
}

// MenuPaintState provides the current menu state to the painter.
type MenuPaintState struct {
	// Bounds is the menu's bounds in window coordinates.
	Bounds geometry.Rect

	// Items is the list of menu items.
	Items []ItemDef

	// HighlightedIndex is the index of the currently highlighted item (-1 for none).
	HighlightedIndex int

	// SelectedIndex is the index of the currently selected item (-1 for none).
	SelectedIndex int

	// ScrollOffset is the index of the first visible item.
	ScrollOffset int

	// VisibleCount is the number of items visible without scrolling.
	VisibleCount int

	// ItemHeight is the height of each item in the menu.
	ItemHeight float32

	// ColorScheme provides theme-derived colors. Zero value means use defaults.
	ColorScheme DropdownColorScheme
}

// DropdownColorScheme provides theme-derived colors for dropdown painting.
// Zero value means the painter should use its built-in defaults.
type DropdownColorScheme struct {
	Background      widget.Color
	Border          widget.Color
	FocusBorder     widget.Color
	TextColor       widget.Color
	PlaceholderText widget.Color
	DisabledBg      widget.Color
	DisabledFg      widget.Color
	MenuBg          widget.Color
	MenuBorder      widget.Color
	ItemHover       widget.Color
	ItemSelected    widget.Color
	ItemDisabled    widget.Color
	ChevronColor    widget.Color
	FocusRing       widget.Color
}

// DefaultPainter provides a minimal fallback painter for dropdowns.
// It draws simple rectangles and text without design system styling.
type DefaultPainter struct{}

// PaintTrigger renders a minimal dropdown trigger.
func (p DefaultPainter) PaintTrigger(canvas widget.Canvas, st *TriggerPaintState) {
	if st.Bounds.IsEmpty() {
		return
	}

	// Background.
	bg := dfltTriggerBg
	if st.Disabled {
		bg = dfltDisabledBg
	}
	canvas.DrawRoundRect(st.Bounds, bg, dfltRadius)

	// Border.
	borderColor := dfltBorder
	borderWidth := dfltBorderWidth
	if st.Focused {
		borderColor = dfltFocusBorder
		borderWidth = dfltFocusBorderWidth
	}
	if st.Disabled {
		borderColor = dfltDisabledFg
	}
	canvas.StrokeRoundRect(st.Bounds, borderColor, dfltRadius, borderWidth)

	// Text.
	textColor := dfltTextColor
	if st.IsPlaceholder {
		textColor = dfltPlaceholder
	}
	if st.Disabled {
		textColor = dfltDisabledFg
	}
	textBounds := geometry.Rect{
		Min: geometry.Pt(st.Bounds.Min.X+dfltPaddingH, st.Bounds.Min.Y),
		Max: geometry.Pt(st.Bounds.Max.X-dfltChevronWidth-dfltPaddingH, st.Bounds.Max.Y),
	}
	canvas.DrawText(st.SelectedText, textBounds, dfltFontSize, textColor, false, dfltTextAlignLeft)

	// Chevron indicator.
	chevronColor := dfltChevronColor
	if st.Disabled {
		chevronColor = dfltDisabledFg
	}
	chevronX := st.Bounds.Max.X - dfltChevronWidth - dfltPaddingH/2
	chevronY := st.Bounds.Center().Y
	drawChevron(canvas, geometry.Pt(chevronX, chevronY), st.Open, chevronColor)
}

// PaintMenu renders a minimal dropdown menu.
func (p DefaultPainter) PaintMenu(canvas widget.Canvas, st *MenuPaintState) {
	if st.Bounds.IsEmpty() {
		return
	}

	// Menu background.
	canvas.DrawRoundRect(st.Bounds, dfltMenuBg, dfltMenuRadius)
	canvas.StrokeRoundRect(st.Bounds, dfltMenuBorder, dfltMenuRadius, dfltBorderWidth)

	// Clip to menu bounds.
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

		// Highlight.
		switch i {
		case st.HighlightedIndex:
			canvas.DrawRect(itemRect, dfltItemHover)
		case st.SelectedIndex:
			canvas.DrawRect(itemRect, dfltItemSelected)
		}

		// Text.
		textColor := dfltTextColor
		if item.Disabled {
			textColor = dfltDisabledFg
		}
		textRect := geometry.Rect{
			Min: geometry.Pt(itemRect.Min.X+dfltPaddingH, itemRect.Min.Y),
			Max: geometry.Pt(itemRect.Max.X-dfltPaddingH, itemRect.Max.Y),
		}
		canvas.DrawText(item.DisplayText(), textRect, dfltFontSize, textColor, false, dfltTextAlignLeft)
	}
}

// drawChevron draws a simple up/down chevron indicator.
func drawChevron(canvas widget.Canvas, center geometry.Point, open bool, color widget.Color) {
	size := dfltChevronSize
	if open {
		// Up chevron: ^
		canvas.DrawLine(
			geometry.Pt(center.X-size, center.Y+size/2),
			geometry.Pt(center.X, center.Y-size/2),
			color, 1.5,
		)
		canvas.DrawLine(
			geometry.Pt(center.X, center.Y-size/2),
			geometry.Pt(center.X+size, center.Y+size/2),
			color, 1.5,
		)
	} else {
		// Down chevron: v
		canvas.DrawLine(
			geometry.Pt(center.X-size, center.Y-size/2),
			geometry.Pt(center.X, center.Y+size/2),
			color, 1.5,
		)
		canvas.DrawLine(
			geometry.Pt(center.X, center.Y+size/2),
			geometry.Pt(center.X+size, center.Y-size/2),
			color, 1.5,
		)
	}
}

// Default painter constants.
var (
	dfltTriggerBg    = widget.ColorWhite
	dfltBorder       = widget.Hex(0x79747E)
	dfltFocusBorder  = widget.Hex(0x6750A4)
	dfltTextColor    = widget.Hex(0x1C1B1F)
	dfltPlaceholder  = widget.Hex(0x49454F)
	dfltDisabledBg   = widget.RGBA(0.12, 0.12, 0.13, 0.04)
	dfltDisabledFg   = widget.RGBA(0.12, 0.12, 0.13, 0.38)
	dfltChevronColor = widget.Hex(0x49454F)
	dfltMenuBg       = widget.ColorWhite
	dfltMenuBorder   = widget.Hex(0xCAC4D0)
	dfltItemHover    = widget.RGBA(0.4, 0.31, 0.64, 0.08)
	dfltItemSelected = widget.RGBA(0.4, 0.31, 0.64, 0.12)
)

const (
	dfltRadius           float32 = 4
	dfltMenuRadius       float32 = 4
	dfltBorderWidth      float32 = 1
	dfltFocusBorderWidth float32 = 2
	dfltPaddingH         float32 = 16
	dfltFontSize         float32 = 16
	dfltTextAlignLeft            = widget.TextAlignLeft
	dfltChevronWidth     float32 = 24
	dfltChevronSize      float32 = 5
)

// Compile-time check.
var _ Painter = DefaultPainter{}
