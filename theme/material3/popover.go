package material3

import (
	"github.com/gogpu/ui/core/popover"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// PopoverPainter renders popovers and tooltips using Material 3 design tokens.
// It applies M3 surface container elevation, outline-variant borders,
// and inverse-surface tooltip styling.
//
// If Theme is nil, PopoverPainter falls back to the default M3 purple palette.
type PopoverPainter struct {
	Theme *Theme // nil uses default M3 purple fallback
}

// PaintPopover renders a popover background with shadow according to M3 specifications.
func (p PopoverPainter) PaintPopover(canvas widget.Canvas, st *popover.PopoverPaintState) {
	if st.Bounds.IsEmpty() {
		return
	}

	bg, border, shadow := p.resolvePopoverColors(st.ColorScheme)
	shadowBlur := m3PopoverShadowBlur
	if st.ColorScheme != (popover.PopoverColorScheme{}) && st.ColorScheme.ShadowBlur > 0 {
		shadowBlur = st.ColorScheme.ShadowBlur
	}

	// Shadow (offset slightly down).
	shadowBounds := geometry.Rect{
		Min: geometry.Pt(st.Bounds.Min.X+m3PopoverShadowOffX, st.Bounds.Min.Y+shadowBlur/2),
		Max: geometry.Pt(st.Bounds.Max.X+m3PopoverShadowOffX, st.Bounds.Max.Y+shadowBlur/2),
	}
	canvas.DrawRoundRect(shadowBounds, shadow, m3PopoverRadius)

	// Background.
	canvas.DrawRoundRect(st.Bounds, bg, m3PopoverRadius)

	// Border.
	canvas.StrokeRoundRect(st.Bounds, border, m3PopoverRadius, m3PopoverBorderWidth)
}

// PaintTooltip renders a tooltip background and text according to M3 specifications.
func (p PopoverPainter) PaintTooltip(canvas widget.Canvas, st *popover.TooltipPaintState) {
	if st.Bounds.IsEmpty() {
		return
	}

	bg, textColor, border := p.resolveTooltipColors(st.ColorScheme)

	// Background.
	canvas.DrawRoundRect(st.Bounds, bg, m3TooltipRadius)

	// Border (subtle).
	if border != (widget.Color{}) {
		canvas.StrokeRoundRect(st.Bounds, border, m3TooltipRadius, m3TooltipBorderWidth)
	}

	// Text.
	textBounds := geometry.Rect{
		Min: geometry.Pt(st.Bounds.Min.X+m3TooltipPadH, st.Bounds.Min.Y),
		Max: geometry.Pt(st.Bounds.Max.X-m3TooltipPadH, st.Bounds.Max.Y),
	}
	canvas.DrawText(st.Text, textBounds, m3TooltipFontSize, textColor, false, m3TooltipTextAlign)
}

// resolvePopoverColors returns background, border, and shadow colors for the popover.
func (p PopoverPainter) resolvePopoverColors(scheme popover.PopoverColorScheme) (bg, border, shadow widget.Color) {
	if scheme != (popover.PopoverColorScheme{}) {
		return scheme.Background, scheme.Border, scheme.Shadow
	}
	if p.Theme != nil {
		cs := p.Theme.Colors
		return cs.SurfaceContainerHigh, cs.OutlineVariant, widget.RGBA(0, 0, 0, 0.15)
	}
	return m3PopoverDefaultBg, m3PopoverDefaultBorder, m3PopoverDefaultShadow
}

// resolveTooltipColors returns background, text, and border colors for the tooltip.
func (p PopoverPainter) resolveTooltipColors(scheme popover.TooltipColorScheme) (bg, text, border widget.Color) {
	if scheme != (popover.TooltipColorScheme{}) {
		return scheme.Background, scheme.TextColor, scheme.Border
	}
	if p.Theme != nil {
		cs := p.Theme.Colors
		return cs.InverseSurface, cs.InverseOnSurface, widget.Color{}
	}
	return m3TooltipDefaultBg, m3TooltipDefaultText, widget.Color{}
}

// Default M3 colors for popovers.
var (
	m3PopoverDefaultBg     = widget.Hex(0xECE6F0) // M3 surface container high
	m3PopoverDefaultBorder = widget.Hex(0xCAC4D0) // M3 outline variant
	m3PopoverDefaultShadow = widget.RGBA(0, 0, 0, 0.15)
)

// Default M3 colors for tooltips.
var (
	m3TooltipDefaultBg   = widget.Hex(0x313033) // M3 inverse surface
	m3TooltipDefaultText = widget.Hex(0xF4EFF4) // M3 inverse on-surface
)

// M3 popover drawing constants.
const (
	m3PopoverRadius      float32 = 12
	m3PopoverBorderWidth float32 = 1
	m3PopoverShadowOffX  float32 = 0
	m3PopoverShadowBlur  float32 = 4
)

// M3 tooltip drawing constants.
const (
	m3TooltipRadius      float32 = 4
	m3TooltipBorderWidth float32 = 0.5
	m3TooltipFontSize    float32 = 12
	m3TooltipPadH        float32 = 8
	m3TooltipTextAlign           = widget.TextAlignLeft
)

// Compile-time check that PopoverPainter implements Painter.
var _ popover.Painter = PopoverPainter{}
