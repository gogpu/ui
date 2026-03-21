package devtools

import (
	"github.com/gogpu/ui/core/popover"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// PopoverPainter renders popovers and tooltips using DevTools design tokens.
// DevTools popovers use SurfaceElevated (#393B40) background with 4px border
// radius, 1px Gray5 border, 8px padding, and a subtle shadow, matching
// JetBrains IDE popup and tooltip styling.
//
// If Theme is nil, PopoverPainter falls back to the default DevTools dark palette.
type PopoverPainter struct {
	Theme *Theme // nil uses default DevTools dark fallback
}

// PaintPopover renders a popover background with shadow according to DevTools specifications.
func (p PopoverPainter) PaintPopover(canvas widget.Canvas, st *popover.PopoverPaintState) {
	if st.Bounds.IsEmpty() {
		return
	}

	bg, border, shadow := p.resolvePopoverColors(st.ColorScheme)

	// Shadow (offset slightly down).
	shadowBounds := geometry.Rect{
		Min: geometry.Pt(st.Bounds.Min.X+dtPopoverShadowOffX, st.Bounds.Min.Y+dtPopoverShadowOffY),
		Max: geometry.Pt(st.Bounds.Max.X+dtPopoverShadowOffX, st.Bounds.Max.Y+dtPopoverShadowOffY),
	}
	canvas.DrawRoundRect(shadowBounds, shadow, dtPopoverRadius)

	// Background.
	canvas.DrawRoundRect(st.Bounds, bg, dtPopoverRadius)

	// Border.
	canvas.StrokeRoundRect(st.Bounds, border, dtPopoverRadius, dtPopoverBorderWidth)
}

// PaintTooltip renders a tooltip background and text according to DevTools specifications.
func (p PopoverPainter) PaintTooltip(canvas widget.Canvas, st *popover.TooltipPaintState) {
	if st.Bounds.IsEmpty() {
		return
	}

	bg, textColor, border := p.resolveTooltipColors(st.ColorScheme)

	// Background.
	canvas.DrawRoundRect(st.Bounds, bg, dtTooltipRadius)

	// Border.
	canvas.StrokeRoundRect(st.Bounds, border, dtTooltipRadius, dtTooltipBorderWidth)

	// Text.
	textBounds := geometry.Rect{
		Min: geometry.Pt(st.Bounds.Min.X+dtTooltipPadH, st.Bounds.Min.Y),
		Max: geometry.Pt(st.Bounds.Max.X-dtTooltipPadH, st.Bounds.Max.Y),
	}
	canvas.DrawText(st.Text, textBounds, dtTooltipFontSize, textColor, false, dtTooltipTextAlign)
}

// resolvePopoverColors returns background, border, and shadow colors for the popover.
func (p PopoverPainter) resolvePopoverColors(scheme popover.PopoverColorScheme) (bg, border, shadow widget.Color) {
	if scheme != (popover.PopoverColorScheme{}) {
		return scheme.Background, scheme.Border, scheme.Shadow
	}
	if p.Theme != nil {
		cs := p.Theme.Colors
		return cs.SurfaceElevated, cs.BorderStrong, cs.Shadow
	}
	return dtPopoverDefaultBg, dtPopoverDefaultBorder, dtPopoverDefaultShadow
}

// resolveTooltipColors returns background, text, and border colors for the tooltip.
func (p PopoverPainter) resolveTooltipColors(scheme popover.TooltipColorScheme) (bg, text, border widget.Color) {
	if scheme != (popover.TooltipColorScheme{}) {
		return scheme.Background, scheme.TextColor, scheme.Border
	}
	if p.Theme != nil {
		cs := p.Theme.Colors
		return cs.SurfaceElevated, cs.OnSurface, cs.BorderStrong
	}
	return dtTooltipDefaultBg, dtTooltipDefaultText, dtTooltipDefaultBorder
}

// Default DevTools colors for popovers.
var (
	dtPopoverDefaultBg     = widget.Hex(0x393B40) // Gray3 (elevated surface)
	dtPopoverDefaultBorder = widget.Hex(0x4E5157) // Gray5 (border strong)
	dtPopoverDefaultShadow = widget.RGBA(0, 0, 0, 0.50)
)

// Default DevTools colors for tooltips.
var (
	dtTooltipDefaultBg     = widget.Hex(0x393B40) // Gray3 (elevated surface)
	dtTooltipDefaultText   = widget.Hex(0xDFE1E5) // Gray12 (on-surface)
	dtTooltipDefaultBorder = widget.Hex(0x4E5157) // Gray5 (border strong)
)

// DevTools popover drawing constants.
const (
	dtPopoverRadius      float32 = 4
	dtPopoverBorderWidth float32 = 1
	dtPopoverShadowOffX  float32 = 0
	dtPopoverShadowOffY  float32 = 2
)

// DevTools tooltip drawing constants.
const (
	dtTooltipRadius      float32 = 4
	dtTooltipBorderWidth float32 = 1
	dtTooltipFontSize    float32 = 12
	dtTooltipPadH        float32 = 8
	dtTooltipTextAlign           = widget.TextAlignLeft
)

// Compile-time check that PopoverPainter implements Painter.
var _ popover.Painter = PopoverPainter{}
