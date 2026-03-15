package material3

import (
	"github.com/gogpu/ui/core/collapsible"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/widget"
)

// CollapsiblePainter renders collapsible section headers using Material 3 design tokens.
// It maps header states (expanded, hovered, pressed, focused) to the M3 color scheme
// with surface colors for the header and on-surface colors for text.
//
// If Theme is nil, CollapsiblePainter falls back to the default M3 purple palette.
type CollapsiblePainter struct {
	Theme *Theme // nil uses default M3 purple fallback
}

// PaintHeader renders a collapsible header according to Material 3 specifications.
func (p CollapsiblePainter) PaintHeader(canvas widget.Canvas, s collapsible.HeaderState) {
	if s.Bounds.IsEmpty() {
		return
	}

	// Resolve colors from theme or defaults.
	bgColor, textColor, arrowColor, focusColor := p.resolveHeaderColors(s)

	// Apply interaction state to background.
	bgColor = m3ApplyCollapsibleState(bgColor, s.Hovered, s.Pressed)

	// Draw header background with rounded corners.
	canvas.DrawRoundRect(s.Bounds, bgColor, m3CollapsibleRadius)

	// Draw arrow indicator.
	m3DrawCollapsibleArrow(canvas, s.Bounds, arrowColor, s.ArrowProgress)

	// Draw title text.
	if s.Title != "" {
		titleBounds := m3CollapsibleTitleBounds(s.Bounds)
		canvas.DrawText(s.Title, titleBounds, m3CollapsibleFontSize, textColor, true, m3CollapsibleTextAlign)
	}

	// Focus ring.
	if s.Focused {
		ringBounds := s.Bounds.Expand(m3CollapsibleFocusRingOffset)
		ringRadius := m3CollapsibleRadius + m3CollapsibleFocusRingOffset
		canvas.StrokeRoundRect(ringBounds, focusColor, ringRadius, m3CollapsibleFocusRingStrokeWidth)
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
			bg = cs.SurfaceContainerLow
		}
		if arrow == (widget.Color{}) {
			arrow = cs.OnSurfaceVariant
		}
		return bg, cs.OnSurface, arrow, cs.Primary.WithAlpha(0.7)
	}

	// Default M3 purple fallback.
	if bg == (widget.Color{}) {
		bg = m3CollapsibleDefaultBg
	}
	if arrow == (widget.Color{}) {
		arrow = m3CollapsibleDefaultArrow
	}
	return bg, m3CollapsibleDefaultText, arrow, m3CollapsibleDefaultFocus
}

// m3ApplyCollapsibleState adjusts a color based on interaction state.
func m3ApplyCollapsibleState(base widget.Color, hovered, pressed bool) widget.Color {
	if pressed {
		return base.Lerp(widget.ColorBlack, m3CollapsiblePressedDarkenFactor)
	}
	if hovered {
		return base.Lerp(widget.ColorWhite, m3CollapsibleHoverLightenFactor)
	}
	return base
}

// m3DrawCollapsibleArrow draws a chevron icon indicator.
// progress 0.0 = right-pointing (>), 1.0 = down-pointing (v).
// Uses icon.ChevronRight / icon.ChevronDown for crisp vector rendering.
func m3DrawCollapsibleArrow(canvas widget.Canvas, headerBounds geometry.Rect, color widget.Color, progress float32) {
	h := headerBounds.Height()
	iconSize := h * m3CollapsibleArrowSizeRatio * 2
	x := headerBounds.Min.X + m3CollapsibleArrowPadding
	y := headerBounds.Min.Y + (h-iconSize)/2
	bounds := geometry.NewRect(x, y, iconSize, iconSize)

	data := icon.ChevronRight
	if progress > 0.5 {
		data = icon.ChevronDown
	}
	icon.Draw(canvas, data, bounds, color)
}

// m3CollapsibleTitleBounds returns the bounds for the title text within the header.
func m3CollapsibleTitleBounds(headerBounds geometry.Rect) geometry.Rect {
	return geometry.NewRect(
		headerBounds.Min.X+m3CollapsibleTitleLeftOffset,
		headerBounds.Min.Y,
		headerBounds.Width()-m3CollapsibleTitleLeftOffset-m3CollapsibleTitleRightPadding,
		headerBounds.Height(),
	)
}

// Default M3 colors for collapsible headers.
var (
	m3CollapsibleDefaultBg    = widget.Hex(0xF7F2FA) // M3 surface container low
	m3CollapsibleDefaultText  = widget.Hex(0x1C1B1F) // M3 on-surface
	m3CollapsibleDefaultArrow = widget.Hex(0x49454F) // M3 on-surface-variant
	m3CollapsibleDefaultFocus = widget.Hex(0x6750A4).WithAlpha(0.7)
)

// M3 collapsible drawing constants.
const (
	m3CollapsibleRadius               float32 = 8
	m3CollapsibleFontSize             float32 = 14
	m3CollapsibleTextAlign                    = widget.TextAlignLeft
	m3CollapsibleArrowPadding         float32 = 8
	m3CollapsibleArrowSizeRatio       float32 = 0.35
	m3CollapsibleArrowStrokeWidth     float32 = 2
	m3CollapsibleTitleLeftOffset      float32 = 36
	m3CollapsibleTitleRightPadding    float32 = 8
	m3CollapsibleFocusRingOffset      float32 = 2
	m3CollapsibleFocusRingStrokeWidth float32 = 2
	m3CollapsibleHoverLightenFactor   float32 = 0.08
	m3CollapsiblePressedDarkenFactor  float32 = 0.12
)

// Compile-time check that CollapsiblePainter implements Painter.
var _ collapsible.Painter = CollapsiblePainter{}
