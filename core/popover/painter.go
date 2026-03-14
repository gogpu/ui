package popover

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// Painter draws the visual representation of popovers and tooltips.
// Each design system (Material 3, Fluent, Cupertino) provides its own
// Painter implementation. If no Painter is set, [DefaultPainter] is used.
type Painter interface {
	// PaintPopover draws the popover background (rounded rect + shadow).
	PaintPopover(canvas widget.Canvas, state *PopoverPaintState)

	// PaintTooltip draws the tooltip background and text.
	PaintTooltip(canvas widget.Canvas, state *TooltipPaintState)
}

// PopoverPaintState provides the current popover state to the painter.
type PopoverPaintState struct {
	// Bounds is the popover content area bounds in window coordinates.
	Bounds geometry.Rect

	// Placement is the resolved placement direction.
	Placement Placement

	// ColorScheme provides theme-derived colors. Zero value means use defaults.
	ColorScheme PopoverColorScheme
}

// TooltipPaintState provides the current tooltip state to the painter.
type TooltipPaintState struct {
	// Bounds is the tooltip's bounds in window coordinates.
	Bounds geometry.Rect

	// Text is the tooltip text to display.
	Text string

	// Placement is the resolved placement direction.
	Placement Placement

	// ColorScheme provides theme-derived colors. Zero value means use defaults.
	ColorScheme TooltipColorScheme
}

// PopoverColorScheme provides theme-derived colors for popover painting.
// Zero value means the painter should use its built-in defaults.
type PopoverColorScheme struct {
	Background widget.Color
	Border     widget.Color
	Shadow     widget.Color
	ShadowBlur float32
}

// TooltipColorScheme provides theme-derived colors for tooltip painting.
// Zero value means the painter should use its built-in defaults.
type TooltipColorScheme struct {
	Background widget.Color
	TextColor  widget.Color
	Border     widget.Color
}

// DefaultPainter provides a minimal fallback painter for popovers and
// tooltips. It draws rounded rectangles with a subtle shadow.
type DefaultPainter struct{}

// PaintPopover renders a popover background with shadow.
func (p DefaultPainter) PaintPopover(canvas widget.Canvas, st *PopoverPaintState) {
	if st.Bounds.IsEmpty() {
		return
	}

	// Shadow (offset slightly down-right).
	shadowBounds := geometry.Rect{
		Min: geometry.Pt(st.Bounds.Min.X+popShadowOffX, st.Bounds.Min.Y+popShadowOffY),
		Max: geometry.Pt(st.Bounds.Max.X+popShadowOffX, st.Bounds.Max.Y+popShadowOffY),
	}
	canvas.DrawRoundRect(shadowBounds, popShadowColor, popRadius)

	// Background.
	canvas.DrawRoundRect(st.Bounds, popBgColor, popRadius)

	// Border.
	canvas.StrokeRoundRect(st.Bounds, popBorderColor, popRadius, popBorderWidth)
}

// PaintTooltip renders a tooltip background and text.
func (p DefaultPainter) PaintTooltip(canvas widget.Canvas, st *TooltipPaintState) {
	if st.Bounds.IsEmpty() {
		return
	}

	// Background.
	canvas.DrawRoundRect(st.Bounds, tipBgColor, tipRadius)

	// Text.
	textBounds := geometry.Rect{
		Min: geometry.Pt(st.Bounds.Min.X+tooltipPadH, st.Bounds.Min.Y),
		Max: geometry.Pt(st.Bounds.Max.X-tooltipPadH, st.Bounds.Max.Y),
	}
	canvas.DrawText(st.Text, textBounds, tooltipFontSz, tipTextColor, false, widget.TextAlignLeft)
}

// Default painter constants for popovers.
var (
	popBgColor     = widget.ColorWhite
	popBorderColor = widget.Hex(0xCAC4D0)
	popShadowColor = widget.RGBA(0, 0, 0, 0.15)
)

const (
	popRadius      float32 = 8
	popBorderWidth float32 = 1
	popShadowOffX  float32 = 1
	popShadowOffY  float32 = 2
)

// Default painter constants for tooltips.
var (
	tipBgColor   = widget.Hex(0x313033)
	tipTextColor = widget.ColorWhite
)

const (
	tipRadius float32 = 4
)

// Tooltip sizing constants.
const (
	tooltipPadH           float32 = 8
	tooltipPadV           float32 = 6
	tooltipFontSz         float32 = 12
	tooltipMaxWidth       float32 = 300
	tooltipCharWidthRatio float32 = 0.5
)

// Compile-time check.
var _ Painter = DefaultPainter{}
