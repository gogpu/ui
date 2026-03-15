package fluent

import (
	"github.com/gogpu/ui/core/tabview"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// TabViewPainter renders tab bars using Fluent Design tokens.
// Fluent tabs have a subtle selected indicator and clean typography.
//
// If Theme is nil, TabViewPainter falls back to the default Fluent Blue palette.
type TabViewPainter struct {
	Theme *Theme // nil uses default Fluent Blue fallback
}

// resolveColors returns the TabColorScheme derived from the painter's Theme.
func (p TabViewPainter) resolveColors() tabview.TabColorScheme {
	if p.Theme == nil {
		return flDefaultTabColors
	}
	cs := p.Theme.Colors
	return tabview.TabColorScheme{
		Background:      cs.Surface,
		SelectedText:    cs.OnSurface,
		UnselectedText:  cs.OnSurfaceSecond,
		Indicator:       cs.Accent,
		HoverBackground: cs.FillSecond,
		CloseButton:     cs.OnSurfaceSecond,
		FocusRing:       cs.StrokeFocus,
	}
}

// PaintTabBar renders a tab bar according to Fluent Design specifications.
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
		flPaintTab(canvas, ts, colors)
	}

	// Selected indicator.
	if ps.SelectedIdx >= 0 && ps.SelectedIdx < len(ps.Tabs) {
		selected := &ps.Tabs[ps.SelectedIdx]
		flPaintTabIndicator(canvas, selected.Bounds, ps.Position, colors.Indicator)
	}

	// Focus ring.
	if ps.Focused {
		innerBounds := ps.Bounds.Expand(-flTabFocusRingInset)
		canvas.StrokeRoundRect(innerBounds, colors.FocusRing, flTabFocusRadius, flTabFocusRingStroke)
	}
}

// flPaintTab draws a single tab with Fluent styling.
func flPaintTab(canvas widget.Canvas, ts *tabview.TabState, colors tabview.TabColorScheme) {
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
		fg = colors.UnselectedText.WithAlpha(flTabDisabledAlpha)
	}

	// Text bounds (shrink if closeable).
	textBounds := ts.Bounds
	if ts.Closeable {
		textBounds = geometry.NewRect(
			ts.Bounds.Min.X,
			ts.Bounds.Min.Y,
			ts.Bounds.Width()-flTabCloseButtonSize-flTabCloseButtonPad,
			ts.Bounds.Height(),
		)
	}

	// Fluent uses semibold for selected, regular for unselected.
	canvas.DrawText(ts.Label, textBounds, flTabFontSize, fg, ts.Selected, flTabTextAlign)

	// Close button.
	if ts.Closeable && !ts.Disabled && !ts.CloseButtonBounds.IsEmpty() {
		flPaintCloseButton(canvas, ts.CloseButtonBounds, colors.CloseButton)
	}
}

// flPaintTabIndicator draws the Fluent tab indicator line.
func flPaintTabIndicator(canvas widget.Canvas, tabBounds geometry.Rect, pos tabview.TabPosition, color widget.Color) {
	tabW := tabBounds.Width()
	var indicatorBounds geometry.Rect
	if pos == tabview.Bottom {
		indicatorBounds = geometry.NewRect(
			tabBounds.Min.X,
			tabBounds.Min.Y,
			tabW,
			flTabIndicatorHeight,
		)
	} else {
		indicatorBounds = geometry.NewRect(
			tabBounds.Min.X,
			tabBounds.Max.Y-flTabIndicatorHeight,
			tabW,
			flTabIndicatorHeight,
		)
	}
	canvas.DrawRoundRect(indicatorBounds, color, flTabIndicatorRadius)
}

// flPaintCloseButton draws the close button (X icon).
func flPaintCloseButton(canvas widget.Canvas, bounds geometry.Rect, color widget.Color) {
	centerX := (bounds.Min.X + bounds.Max.X) / 2
	centerY := (bounds.Min.Y + bounds.Max.Y) / 2
	halfSize := flTabCloseIconHalf

	canvas.DrawLine(
		geometry.Pt(centerX-halfSize, centerY-halfSize),
		geometry.Pt(centerX+halfSize, centerY+halfSize),
		color, flTabCloseStroke,
	)
	canvas.DrawLine(
		geometry.Pt(centerX+halfSize, centerY-halfSize),
		geometry.Pt(centerX-halfSize, centerY+halfSize),
		color, flTabCloseStroke,
	)
}

// flDefaultTabColors holds the default Fluent tab color scheme.
var flDefaultTabColors = tabview.TabColorScheme{
	Background:      widget.ColorWhite,
	SelectedText:    widget.Hex(0x1A1A1A),
	UnselectedText:  widget.Hex(0x616161),
	Indicator:       DefaultAccentColor,
	HoverBackground: widget.RGBA(0, 0, 0, 0.06),
	CloseButton:     widget.Hex(0x616161),
	FocusRing:       DefaultAccentColor,
}

// Fluent tab bar drawing constants.
const (
	flTabFontSize        float32 = 14
	flTabTextAlign               = widget.TextAlignCenter
	flTabIndicatorHeight float32 = 2
	flTabIndicatorRadius float32 = 1
	flTabDisabledAlpha   float32 = 0.38
	flTabFocusRingInset  float32 = 1
	flTabFocusRingStroke float32 = 2
	flTabFocusRadius     float32 = 4
	flTabCloseButtonSize float32 = 16
	flTabCloseButtonPad  float32 = 8
	flTabCloseIconHalf   float32 = 4
	flTabCloseStroke     float32 = 1.5
)

// Compile-time check that TabViewPainter implements Painter.
var _ tabview.Painter = TabViewPainter{}
