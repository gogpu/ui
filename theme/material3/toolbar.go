package material3

import (
	"github.com/gogpu/ui/core/toolbar"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/widget"
)

// ToolbarPainter renders toolbars using Material 3 design tokens.
// It maps M3 color roles to toolbar elements: surface for background,
// on-surface for icons, and primary-container for hover highlights.
//
// If Theme is nil, ToolbarPainter falls back to the default M3 purple palette.
type ToolbarPainter struct {
	Theme *Theme // nil uses default M3 purple fallback
}

// resolveColors returns M3-derived colors for toolbar painting.
func (p ToolbarPainter) resolveColors() toolbarColors {
	if p.Theme == nil {
		return m3DefaultToolbarColors
	}
	cs := p.Theme.Colors
	return toolbarColors{
		Background:  cs.SurfaceContainer,
		IconColor:   cs.OnSurface,
		HoverBg:     cs.PrimaryContainer.WithAlpha(0.3),
		PressedBg:   cs.PrimaryContainer.WithAlpha(0.5),
		DisabledFg:  cs.OnSurface.WithAlpha(0.38),
		SeparatorFg: cs.OutlineVariant,
		FocusRing:   cs.Primary.WithAlpha(0.7),
	}
}

// PaintToolbar renders the toolbar background using M3 surface color.
func (p ToolbarPainter) PaintToolbar(canvas widget.Canvas, state toolbar.PaintToolbarState) {
	if state.Bounds.IsEmpty() {
		return
	}
	colors := p.resolveColors()
	canvas.DrawRect(state.Bounds, colors.Background)
}

// PaintButtonItem renders a button item with icon and optional label using M3 colors.
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
			canvas.DrawRoundRect(state.Bounds, bg, m3ToolbarItemRadius)
		}
	}

	// Icon color.
	fg := colors.IconColor
	if state.Disabled {
		fg = colors.DisabledFg
	}

	// Draw icon centered in the button area.
	iconBounds := m3ToolbarIconBounds(state.Bounds, state.ShowLabel)
	icon.Draw(canvas, state.Icon, iconBounds, fg)

	// Draw label text if ShowLabel is true.
	if state.ShowLabel && state.Label != "" {
		textBounds := m3ToolbarTextBounds(state.Bounds, iconBounds)
		canvas.DrawText(state.Label, textBounds, m3ToolbarFontSize, fg, false, widget.TextAlignLeft)
	}

	// Focus ring.
	if state.Focused && !state.Disabled {
		ringBounds := state.Bounds.Expand(m3ToolbarFocusRingOffset)
		canvas.StrokeRoundRect(ringBounds, colors.FocusRing, m3ToolbarItemRadius+m3ToolbarFocusRingOffset, m3ToolbarFocusRingWidth)
	}
}

// PaintSeparator renders a vertical separator line using M3 outline-variant.
func (p ToolbarPainter) PaintSeparator(canvas widget.Canvas, bounds geometry.Rect) {
	if bounds.IsEmpty() {
		return
	}
	colors := p.resolveColors()
	centerX := bounds.Min.X + bounds.Width()/2
	canvas.DrawLine(
		geometry.Pt(centerX, bounds.Min.Y+m3ToolbarSepInset),
		geometry.Pt(centerX, bounds.Max.Y-m3ToolbarSepInset),
		colors.SeparatorFg, m3ToolbarSepWidth,
	)
}

// toolbarColors holds resolved M3 color roles for toolbar painting.
type toolbarColors struct {
	Background  widget.Color
	IconColor   widget.Color
	HoverBg     widget.Color
	PressedBg   widget.Color
	DisabledFg  widget.Color
	SeparatorFg widget.Color
	FocusRing   widget.Color
}

// m3DefaultToolbarColors holds default M3 purple fallback colors.
var m3DefaultToolbarColors = toolbarColors{
	Background:  widget.Hex(0xECE6F0),                // M3 surface-container
	IconColor:   widget.Hex(0x1C1B1F),                // M3 on-surface
	HoverBg:     widget.Hex(0xEADDFF).WithAlpha(0.3), // M3 primary-container @ 30%
	PressedBg:   widget.Hex(0xEADDFF).WithAlpha(0.5), // M3 primary-container @ 50%
	DisabledFg:  widget.RGBA(0.12, 0.12, 0.13, 0.38),
	SeparatorFg: widget.Hex(0xCAC4D0), // M3 outline-variant
	FocusRing:   widget.Hex(0x6750A4).WithAlpha(0.7),
}

// m3ToolbarIconBounds calculates the icon bounds within a button item.
func m3ToolbarIconBounds(itemBounds geometry.Rect, showLabel bool) geometry.Rect {
	h := itemBounds.Height() - m3ToolbarIconPadding*2
	if h < 0 {
		h = 0
	}
	iconSize := h
	if iconSize > m3ToolbarMaxIconSize {
		iconSize = m3ToolbarMaxIconSize
	}
	if showLabel {
		x := itemBounds.Min.X + m3ToolbarIconPadding
		y := itemBounds.Min.Y + (itemBounds.Height()-iconSize)/2
		return geometry.NewRect(x, y, iconSize, iconSize)
	}
	centerX := itemBounds.Min.X + itemBounds.Width()/2
	centerY := itemBounds.Min.Y + itemBounds.Height()/2
	return geometry.NewRect(centerX-iconSize/2, centerY-iconSize/2, iconSize, iconSize)
}

// m3ToolbarTextBounds calculates label text bounds next to the icon.
func m3ToolbarTextBounds(itemBounds, iconRect geometry.Rect) geometry.Rect {
	x := iconRect.Max.X + m3ToolbarTextIconGap
	return geometry.NewRect(
		x,
		itemBounds.Min.Y,
		itemBounds.Max.X-m3ToolbarIconPadding-x,
		itemBounds.Height(),
	)
}

// M3 toolbar drawing constants.
const (
	m3ToolbarItemRadius      float32 = 8
	m3ToolbarIconPadding     float32 = 6
	m3ToolbarMaxIconSize     float32 = 20
	m3ToolbarTextIconGap     float32 = 4
	m3ToolbarFontSize        float32 = 12
	m3ToolbarSepInset        float32 = 6
	m3ToolbarSepWidth        float32 = 1
	m3ToolbarFocusRingOffset float32 = 2
	m3ToolbarFocusRingWidth  float32 = 2
)

// Compile-time check that ToolbarPainter implements Painter.
var _ toolbar.Painter = ToolbarPainter{}
