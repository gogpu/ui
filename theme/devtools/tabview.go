package devtools

import (
	"github.com/gogpu/ui/core/tabview"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// TabViewPainter renders tab bars using DevTools design tokens.
// DevTools tabs use a flat underline indicator (2px Blue6) instead of pills,
// matching JetBrains IDE editor tab styling. Unselected tabs show Gray9 text,
// selected tabs show Gray12 text with a bold bottom accent line.
//
// If Theme is nil, TabViewPainter falls back to the default DevTools dark palette.
type TabViewPainter struct {
	Theme *Theme // nil uses default DevTools dark fallback
}

// resolveColors returns the TabColorScheme derived from the painter's Theme.
func (p TabViewPainter) resolveColors() tabview.TabColorScheme {
	if p.Theme == nil {
		return dtDefaultTabColors
	}
	cs := p.Theme.Colors
	return tabview.TabColorScheme{
		Background:      cs.Surface,
		SelectedText:    cs.OnSurface,
		UnselectedText:  cs.OnSurfaceSecondary,
		Indicator:       cs.Primary,
		HoverBackground: cs.SurfaceElevated,
		CloseButton:     cs.OnSurfaceSecondary,
		FocusRing:       cs.BorderFocus,
	}
}

// PaintTabBar renders a tab bar according to DevTools design specifications.
func (p TabViewPainter) PaintTabBar(canvas widget.Canvas, ps tabview.PaintState) {
	if ps.Bounds.IsEmpty() {
		return
	}

	colors := ps.ColorScheme
	if colors == (tabview.TabColorScheme{}) {
		colors = p.resolveColors()
	}

	// Tab bar background.
	canvas.DrawRect(ps.Bounds, colors.Background)

	// Draw each tab.
	for i := range ps.Tabs {
		ts := &ps.Tabs[i]
		dtPaintTab(canvas, ts, colors)
	}

	// Selected indicator: 2px underline (DevTools style, no pill).
	if ps.SelectedIdx >= 0 && ps.SelectedIdx < len(ps.Tabs) {
		selected := &ps.Tabs[ps.SelectedIdx]
		dtPaintTabIndicator(canvas, selected.Bounds, ps.Position, colors.Indicator)
	}

	// Focus ring.
	if ps.Focused {
		canvas.StrokeRect(ps.Bounds, colors.FocusRing, dtTabFocusRingStroke)
	}
}

// dtPaintTab draws a single tab with DevTools styling.
func dtPaintTab(canvas widget.Canvas, ts *tabview.TabState, colors tabview.TabColorScheme) {
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
		fg = colors.UnselectedText.WithAlpha(dtTabDisabledAlpha)
	}

	// Text bounds (shrink if closeable).
	textBounds := ts.Bounds
	if ts.Closeable {
		textBounds = geometry.NewRect(
			ts.Bounds.Min.X,
			ts.Bounds.Min.Y,
			ts.Bounds.Width()-dtTabCloseButtonSize-dtTabCloseButtonPad,
			ts.Bounds.Height(),
		)
	}

	canvas.DrawText(ts.Label, textBounds, dtTabFontSize, fg, ts.Selected, dtTabTextAlign)

	// Close button.
	if ts.Closeable && !ts.Disabled && !ts.CloseButtonBounds.IsEmpty() {
		dtPaintCloseButton(canvas, ts.CloseButtonBounds, colors.CloseButton)
	}
}

// dtPaintTabIndicator draws the DevTools tab indicator (flat 2px underline).
func dtPaintTabIndicator(canvas widget.Canvas, tabBounds geometry.Rect, pos tabview.TabPosition, color widget.Color) {
	tabW := tabBounds.Width()
	var indicatorBounds geometry.Rect
	if pos == tabview.Bottom {
		indicatorBounds = geometry.NewRect(
			tabBounds.Min.X,
			tabBounds.Min.Y,
			tabW,
			dtTabIndicatorHeight,
		)
	} else {
		indicatorBounds = geometry.NewRect(
			tabBounds.Min.X,
			tabBounds.Max.Y-dtTabIndicatorHeight,
			tabW,
			dtTabIndicatorHeight,
		)
	}
	// DevTools uses a flat rectangle indicator, no rounding.
	canvas.DrawRect(indicatorBounds, color)
}

// dtPaintCloseButton draws the close button (X icon).
func dtPaintCloseButton(canvas widget.Canvas, bounds geometry.Rect, color widget.Color) {
	centerX := (bounds.Min.X + bounds.Max.X) / 2
	centerY := (bounds.Min.Y + bounds.Max.Y) / 2
	halfSize := dtTabCloseIconHalf

	canvas.DrawLine(
		geometry.Pt(centerX-halfSize, centerY-halfSize),
		geometry.Pt(centerX+halfSize, centerY+halfSize),
		color, dtTabCloseStroke,
	)
	canvas.DrawLine(
		geometry.Pt(centerX+halfSize, centerY-halfSize),
		geometry.Pt(centerX-halfSize, centerY+halfSize),
		color, dtTabCloseStroke,
	)
}

// dtDefaultTabColors holds the default DevTools dark tab color scheme.
var dtDefaultTabColors = tabview.TabColorScheme{
	Background:      widget.Hex(0x2B2D30), // Gray2 (Surface)
	SelectedText:    widget.Hex(0xDFE1E5), // Gray12
	UnselectedText:  widget.Hex(0x9DA0A8), // Gray9
	Indicator:       DefaultAccentColor,
	HoverBackground: widget.Hex(0x393B40), // Gray3 (SurfaceElevated)
	CloseButton:     widget.Hex(0x9DA0A8), // Gray9
	FocusRing:       DefaultAccentColor,
}

// DevTools tab bar drawing constants.
const (
	dtTabFontSize        float32 = 13
	dtTabTextAlign               = widget.TextAlignCenter
	dtTabIndicatorHeight float32 = 2
	dtTabDisabledAlpha   float32 = 0.38
	dtTabFocusRingStroke float32 = 1
	dtTabCloseButtonSize float32 = 16
	dtTabCloseButtonPad  float32 = 8
	dtTabCloseIconHalf   float32 = 4
	dtTabCloseStroke     float32 = 1.5
)

// Compile-time check that TabViewPainter implements Painter.
var _ tabview.Painter = TabViewPainter{}
