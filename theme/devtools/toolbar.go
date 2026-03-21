package devtools

import (
	"github.com/gogpu/ui/core/toolbar"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/widget"
)

// ToolbarPainter renders toolbars using DevTools design tokens.
// DevTools toolbars are flat and transparent with no background, featuring
// a subtle 1px bottom border (Gray3) and compact 28px square icon buttons
// with 4px radius, matching JetBrains IDE toolbar styling.
//
// If Theme is nil, ToolbarPainter falls back to the default DevTools dark palette.
type ToolbarPainter struct {
	Theme *Theme // nil uses default DevTools dark fallback
}

// resolveColors returns DevTools-derived colors for toolbar painting.
func (p ToolbarPainter) resolveColors() dtToolbarColors {
	if p.Theme == nil {
		return dtDefaultToolbarColors
	}
	cs := p.Theme.Colors
	return dtToolbarColors{
		Background:  widget.ColorTransparent,
		BorderColor: cs.Border,
		IconColor:   cs.OnSurface,
		HoverBg:     cs.ControlFill,
		PressedBg:   cs.ControlHover,
		DisabledFg:  cs.OnSurfaceDisabled,
		SeparatorFg: cs.Border,
		FocusRing:   cs.BorderFocus,
	}
}

// PaintToolbar renders the toolbar background with a subtle bottom border.
func (p ToolbarPainter) PaintToolbar(canvas widget.Canvas, state toolbar.PaintToolbarState) {
	if state.Bounds.IsEmpty() {
		return
	}
	colors := p.resolveColors()

	// Transparent background (flat DevTools style).
	if !colors.Background.IsTransparent() {
		canvas.DrawRect(state.Bounds, colors.Background)
	}

	// 1px bottom border.
	borderRect := geometry.NewRect(
		state.Bounds.Min.X,
		state.Bounds.Max.Y-dtToolbarBorderWidth,
		state.Bounds.Width(),
		dtToolbarBorderWidth,
	)
	canvas.DrawRect(borderRect, colors.BorderColor)
}

// PaintButtonItem renders a button item with icon and optional text label.
func (p ToolbarPainter) PaintButtonItem(canvas widget.Canvas, state toolbar.PaintButtonState) {
	if state.Bounds.IsEmpty() {
		return
	}
	colors := p.resolveColors()

	// Background on hover/press (disabled items get no feedback).
	if !state.Disabled {
		bg := widget.ColorTransparent
		if state.Pressed {
			bg = colors.PressedBg
		} else if state.Hovered {
			bg = colors.HoverBg
		}
		if !bg.IsTransparent() {
			canvas.DrawRoundRect(state.Bounds, bg, dtToolbarItemRadius)
		}
	}

	// Icon color.
	fg := colors.IconColor
	if state.Disabled {
		fg = colors.DisabledFg
	}

	// Draw icon centered in the button area.
	iconBounds := dtToolbarIconBounds(state.Bounds, state.ShowLabel)
	icon.Draw(canvas, state.Icon, iconBounds, fg)

	// Draw label text if ShowLabel is true.
	if state.ShowLabel && state.Label != "" {
		textBounds := dtToolbarTextBounds(state.Bounds, iconBounds)
		canvas.DrawText(state.Label, textBounds, dtToolbarFontSize, fg, false, widget.TextAlignLeft)
	}

	// Focus ring (1px DevTools style).
	if state.Focused && !state.Disabled {
		canvas.StrokeRoundRect(state.Bounds, colors.FocusRing, dtToolbarItemRadius, dtToolbarFocusRingWidth)
	}
}

// PaintSeparator renders a vertical separator line.
func (p ToolbarPainter) PaintSeparator(canvas widget.Canvas, bounds geometry.Rect) {
	if bounds.IsEmpty() {
		return
	}
	colors := p.resolveColors()
	centerX := bounds.Min.X + bounds.Width()/2
	canvas.DrawLine(
		geometry.Pt(centerX, bounds.Min.Y+dtToolbarSepInset),
		geometry.Pt(centerX, bounds.Max.Y-dtToolbarSepInset),
		colors.SeparatorFg, dtToolbarSepWidth,
	)
}

// dtToolbarColors holds resolved DevTools color roles for toolbar painting.
type dtToolbarColors struct {
	Background  widget.Color
	BorderColor widget.Color
	IconColor   widget.Color
	HoverBg     widget.Color
	PressedBg   widget.Color
	DisabledFg  widget.Color
	SeparatorFg widget.Color
	FocusRing   widget.Color
}

// dtDefaultToolbarColors holds default DevTools dark fallback colors.
var dtDefaultToolbarColors = dtToolbarColors{
	Background:  widget.ColorTransparent,
	BorderColor: widget.Hex(0x393B40), // Gray3
	IconColor:   widget.Hex(0xDFE1E5), // Gray12
	HoverBg:     widget.Hex(0x393B40), // Gray3 (control fill)
	PressedBg:   widget.Hex(0x43454A), // Gray4 (control hover)
	DisabledFg:  widget.Hex(0x6F737A), // Gray7
	SeparatorFg: widget.Hex(0x393B40), // Gray3
	FocusRing:   widget.Hex(0x3574F0), // Blue6
}

// dtToolbarIconBounds calculates the icon bounds within a button item.
func dtToolbarIconBounds(itemBounds geometry.Rect, showLabel bool) geometry.Rect {
	h := itemBounds.Height() - dtToolbarIconPadding*2
	if h < 0 {
		h = 0
	}
	iconSize := h
	if iconSize > dtToolbarMaxIconSize {
		iconSize = dtToolbarMaxIconSize
	}
	if showLabel {
		x := itemBounds.Min.X + dtToolbarIconPadding
		y := itemBounds.Min.Y + (itemBounds.Height()-iconSize)/2
		return geometry.NewRect(x, y, iconSize, iconSize)
	}
	centerX := itemBounds.Min.X + itemBounds.Width()/2
	centerY := itemBounds.Min.Y + itemBounds.Height()/2
	return geometry.NewRect(centerX-iconSize/2, centerY-iconSize/2, iconSize, iconSize)
}

// dtToolbarTextBounds calculates label text bounds next to the icon.
func dtToolbarTextBounds(itemBounds, iconRect geometry.Rect) geometry.Rect {
	x := iconRect.Max.X + dtToolbarTextIconGap
	return geometry.NewRect(
		x,
		itemBounds.Min.Y,
		itemBounds.Max.X-dtToolbarIconPadding-x,
		itemBounds.Height(),
	)
}

// DevTools toolbar drawing constants.
const (
	dtToolbarItemRadius     float32 = 4
	dtToolbarIconPadding    float32 = 5
	dtToolbarMaxIconSize    float32 = 20 // JB: experimentalToolbarButtonIconSize = 20
	dtToolbarTextIconGap    float32 = 4
	dtToolbarFontSize       float32 = 12
	dtToolbarSepInset       float32 = 4
	dtToolbarSepWidth       float32 = 1
	dtToolbarBorderWidth    float32 = 1
	dtToolbarFocusRingWidth float32 = 1
)

// Compile-time check that ToolbarPainter implements Painter.
var _ toolbar.Painter = ToolbarPainter{}
