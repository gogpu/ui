package devtools

import (
	"github.com/gogpu/ui/core/collapsible"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/widget"
)

// CollapsiblePainter renders collapsible section headers using DevTools design tokens.
// DevTools collapsibles use minimal styling: Surface background, a subtle chevron
// (Gray7 -> Gray12 on hover), 1px bottom border separator, and compact 8px padding,
// matching JetBrains IDE tool window section headers.
//
// If Theme is nil, CollapsiblePainter falls back to the default DevTools dark palette.
type CollapsiblePainter struct {
	Theme *Theme // nil uses default DevTools dark fallback
}

// PaintHeader renders a collapsible header according to DevTools specifications.
func (p CollapsiblePainter) PaintHeader(canvas widget.Canvas, s collapsible.HeaderState) {
	if s.Bounds.IsEmpty() {
		return
	}

	// Resolve colors from theme or defaults.
	bgColor, textColor, arrowColor, focusColor := p.resolveHeaderColors(s)

	// Apply interaction state to background.
	bgColor = dtApplyCollapsibleState(bgColor, s.Hovered, s.Pressed)

	// Draw header background (flat, no rounded corners).
	canvas.DrawRect(s.Bounds, bgColor)

	// 1px bottom border separator.
	borderRect := geometry.NewRect(
		s.Bounds.Min.X,
		s.Bounds.Max.Y-dtCollapsibleBorderWidth,
		s.Bounds.Width(),
		dtCollapsibleBorderWidth,
	)
	canvas.DrawRect(borderRect, dtCollapsibleBorderColor(p.Theme))

	// Draw arrow indicator.
	dtDrawCollapsibleArrow(canvas, s.Bounds, arrowColor, s.ArrowProgress)

	// Draw title text.
	if s.Title != "" {
		titleBounds := dtCollapsibleTitleBounds(s.Bounds)
		canvas.DrawText(s.Title, titleBounds, dtCollapsibleFontSize, textColor, true, dtCollapsibleTextAlign)
	}

	// Focus ring (1px DevTools style).
	if s.Focused {
		canvas.StrokeRect(s.Bounds, focusColor, dtCollapsibleFocusRingWidth)
	}
}

// resolveHeaderColors returns background, text, arrow, and focus colors.
func (p CollapsiblePainter) resolveHeaderColors(s collapsible.HeaderState) (bg, text, arrow, focus widget.Color) {
	// Use explicit overrides first.
	bg = s.HeaderColor
	arrow = s.ArrowColor

	if p.Theme != nil {
		cs := p.Theme.Colors
		if bg == (widget.Color{}) {
			bg = cs.Surface
		}
		if arrow == (widget.Color{}) {
			arrow = cs.OnSurfaceDisabled
		}
		return bg, cs.OnSurface, arrow, cs.BorderFocus
	}

	// Default DevTools dark fallback.
	if bg == (widget.Color{}) {
		bg = dtCollapsibleDefaultBg
	}
	if arrow == (widget.Color{}) {
		arrow = dtCollapsibleDefaultArrow
	}
	return bg, dtCollapsibleDefaultText, arrow, dtCollapsibleDefaultFocus
}

// dtApplyCollapsibleState adjusts a color based on interaction state.
func dtApplyCollapsibleState(base widget.Color, hovered, pressed bool) widget.Color {
	if pressed {
		return base.Lerp(widget.ColorBlack, dtCollapsiblePressedFactor)
	}
	if hovered {
		return base.Lerp(widget.ColorWhite, dtCollapsibleHoverFactor)
	}
	return base
}

// dtDrawCollapsibleArrow draws a chevron icon indicator.
// progress 0.0 = right-pointing (>), 1.0 = down-pointing (v).
func dtDrawCollapsibleArrow(canvas widget.Canvas, headerBounds geometry.Rect, color widget.Color, progress float32) {
	h := headerBounds.Height()
	iconSize := h * dtCollapsibleArrowSizeRatio * 2
	x := headerBounds.Min.X + dtCollapsibleArrowPadding
	y := headerBounds.Min.Y + (h-iconSize)/2
	bounds := geometry.NewRect(x, y, iconSize, iconSize)

	data := icon.ChevronRight
	if progress > 0.5 {
		data = icon.ChevronDown
	}
	icon.Draw(canvas, data, bounds, color)
}

// dtCollapsibleTitleBounds returns the bounds for the title text within the header.
func dtCollapsibleTitleBounds(headerBounds geometry.Rect) geometry.Rect {
	return geometry.NewRect(
		headerBounds.Min.X+dtCollapsibleTitleLeftOffset,
		headerBounds.Min.Y,
		headerBounds.Width()-dtCollapsibleTitleLeftOffset-dtCollapsibleTitleRightPadding,
		headerBounds.Height(),
	)
}

// dtCollapsibleBorderColor returns the border color for the bottom separator.
func dtCollapsibleBorderColor(theme *Theme) widget.Color {
	if theme != nil {
		return theme.Colors.Border
	}
	return widget.Hex(0x393B40) // Gray3
}

// Default DevTools colors for collapsible headers.
var (
	dtCollapsibleDefaultBg    = widget.Hex(0x2B2D30) // Gray2 (surface)
	dtCollapsibleDefaultText  = widget.Hex(0xDFE1E5) // Gray12 (on-surface)
	dtCollapsibleDefaultArrow = widget.Hex(0x6F737A) // Gray7 (muted)
	dtCollapsibleDefaultFocus = widget.Hex(0x3574F0) // Blue6 (accent)
)

// DevTools collapsible drawing constants.
const (
	dtCollapsibleFontSize          float32 = 13
	dtCollapsibleTextAlign                 = widget.TextAlignLeft
	dtCollapsibleArrowPadding      float32 = 8
	dtCollapsibleArrowSizeRatio    float32 = 0.35
	dtCollapsibleTitleLeftOffset   float32 = 28
	dtCollapsibleTitleRightPadding float32 = 8
	dtCollapsibleFocusRingWidth    float32 = 1
	dtCollapsibleBorderWidth       float32 = 1
	dtCollapsibleHoverFactor       float32 = 0.06
	dtCollapsiblePressedFactor     float32 = 0.08
)

// Compile-time check that CollapsiblePainter implements Painter.
var _ collapsible.Painter = CollapsiblePainter{}
