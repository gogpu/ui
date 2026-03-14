package material3

import (
	"github.com/gogpu/ui/core/tabview"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// TabViewPainter renders tab bars using Material 3 design tokens.
// It maps tab states to the M3 color scheme and applies appropriate
// interaction feedback including an animated indicator.
//
// If Theme is nil, TabViewPainter falls back to the default M3 purple palette.
type TabViewPainter struct {
	Theme *Theme // nil uses default M3 purple fallback
}

// resolveColors returns the TabColorScheme derived from the painter's Theme.
func (p TabViewPainter) resolveColors() tabview.TabColorScheme {
	if p.Theme == nil {
		return m3DefaultTabColors
	}
	cs := p.Theme.Colors
	return tabview.TabColorScheme{
		Background:      cs.Surface,
		SelectedText:    cs.Primary,
		UnselectedText:  cs.OnSurfaceVariant,
		Indicator:       cs.Primary,
		HoverBackground: cs.OnSurface.WithAlpha(m3TabHoverAlpha),
		CloseButton:     cs.OnSurfaceVariant,
		FocusRing:       cs.Primary.WithAlpha(m3TabFocusAlpha),
	}
}

// PaintTabBar renders a tab bar according to Material 3 specifications.
func (p TabViewPainter) PaintTabBar(canvas widget.Canvas, ps tabview.PaintState) {
	if ps.Bounds.IsEmpty() {
		return
	}

	// Determine the color scheme to use.
	colors := ps.ColorScheme
	if colors == (tabview.TabColorScheme{}) {
		colors = p.resolveColors()
	}

	// Tab bar background.
	canvas.DrawRect(ps.Bounds, colors.Background)

	// Draw each tab.
	for i := range ps.Tabs {
		ts := &ps.Tabs[i]
		m3PaintTab(canvas, ts, colors)
	}

	// Draw indicator under/over selected tab.
	if ps.SelectedIdx >= 0 && ps.SelectedIdx < len(ps.Tabs) {
		selected := &ps.Tabs[ps.SelectedIdx]
		m3PaintIndicator(canvas, selected.Bounds, ps.Position, colors.Indicator)
	}

	// Focus ring.
	if ps.Focused {
		ringBounds := ps.Bounds.Expand(m3TabFocusRingOffset)
		canvas.StrokeRoundRect(ringBounds, colors.FocusRing, m3TabFocusRadius, m3TabFocusRingStroke)
	}
}

// m3PaintTab draws a single tab with M3 styling.
func m3PaintTab(canvas widget.Canvas, ts *tabview.TabState, colors tabview.TabColorScheme) {
	if ts.Bounds.IsEmpty() {
		return
	}

	// Hover overlay.
	if ts.Hovered && !ts.Disabled {
		canvas.DrawRect(ts.Bounds, colors.HoverBackground)
	}

	// Text color.
	fg := colors.UnselectedText
	if ts.Selected {
		fg = colors.SelectedText
	}
	if ts.Disabled {
		fg = colors.UnselectedText.WithAlpha(m3TabDisabledAlpha)
	}

	// Text bounds (shrink if closeable).
	textBounds := ts.Bounds
	if ts.Closeable {
		textBounds = geometry.NewRect(
			ts.Bounds.Min.X,
			ts.Bounds.Min.Y,
			ts.Bounds.Width()-m3TabCloseButtonSize-m3TabCloseButtonPad,
			ts.Bounds.Height(),
		)
	}

	canvas.DrawText(ts.Label, textBounds, m3TabFontSize, fg, ts.Selected, m3TabTextAlign)

	// Close button.
	if ts.Closeable && !ts.Disabled && !ts.CloseButtonBounds.IsEmpty() {
		m3PaintCloseButton(canvas, ts.CloseButtonBounds, colors.CloseButton)
	}
}

// m3PaintIndicator draws the M3 tab indicator line.
func m3PaintIndicator(canvas widget.Canvas, tabBounds geometry.Rect, pos tabview.TabPosition, color widget.Color) {
	tabW := tabBounds.Width()
	var indicatorBounds geometry.Rect
	if pos == tabview.Bottom {
		indicatorBounds = geometry.NewRect(
			tabBounds.Min.X,
			tabBounds.Min.Y,
			tabW,
			m3TabIndicatorHeight,
		)
	} else {
		indicatorBounds = geometry.NewRect(
			tabBounds.Min.X,
			tabBounds.Max.Y-m3TabIndicatorHeight,
			tabW,
			m3TabIndicatorHeight,
		)
	}
	canvas.DrawRoundRect(indicatorBounds, color, m3TabIndicatorRadius)
}

// m3PaintCloseButton draws the M3 close button (X icon).
func m3PaintCloseButton(canvas widget.Canvas, bounds geometry.Rect, color widget.Color) {
	centerX := (bounds.Min.X + bounds.Max.X) / 2
	centerY := (bounds.Min.Y + bounds.Max.Y) / 2
	halfSize := m3TabCloseIconHalf

	canvas.DrawLine(
		geometry.Pt(centerX-halfSize, centerY-halfSize),
		geometry.Pt(centerX+halfSize, centerY+halfSize),
		color, m3TabCloseStroke,
	)
	canvas.DrawLine(
		geometry.Pt(centerX+halfSize, centerY-halfSize),
		geometry.Pt(centerX-halfSize, centerY+halfSize),
		color, m3TabCloseStroke,
	)
}

// m3DefaultTabColors holds the default M3 purple color scheme for tabs.
var m3DefaultTabColors = tabview.TabColorScheme{
	Background:      widget.Hex(0xFEF7FF),                // M3 surface
	SelectedText:    widget.Hex(0x6750A4),                // M3 primary
	UnselectedText:  widget.Hex(0x49454F),                // M3 on-surface-variant
	Indicator:       widget.Hex(0x6750A4),                // M3 primary
	HoverBackground: widget.RGBA(0.12, 0.12, 0.13, 0.08), // M3 hover overlay
	CloseButton:     widget.Hex(0x49454F),                // M3 on-surface-variant
	FocusRing:       widget.Hex(0x6750A4).WithAlpha(0.7), // M3 focus ring
}

// M3 tab bar drawing constants.
const (
	m3TabFontSize        float32 = 14
	m3TabTextAlign               = widget.TextAlignCenter
	m3TabIndicatorHeight float32 = 3
	m3TabIndicatorRadius float32 = 1.5
	m3TabHoverAlpha      float32 = 0.08
	m3TabDisabledAlpha   float32 = 0.38
	m3TabFocusAlpha      float32 = 0.7
	m3TabFocusRingOffset float32 = 2
	m3TabFocusRingStroke float32 = 2
	m3TabFocusRadius     float32 = 4
	m3TabCloseButtonSize float32 = 16
	m3TabCloseButtonPad  float32 = 8
	m3TabCloseIconHalf   float32 = 4
	m3TabCloseStroke     float32 = 1.5
)

// Compile-time check that TabViewPainter implements Painter.
var _ tabview.Painter = TabViewPainter{}
