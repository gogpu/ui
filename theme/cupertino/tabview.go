package cupertino

import (
	"github.com/gogpu/ui/core/tabview"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// TabViewPainter renders tab bars as iOS-style segmented controls.
// Instead of a traditional tab bar with an underline indicator, Cupertino
// uses a pill-shaped segmented control where the selected segment has a
// raised white background.
//
// If Theme is nil, TabViewPainter falls back to the default system blue palette.
type TabViewPainter struct {
	Theme *Theme // nil uses default system blue fallback
}

// resolveColors returns the TabColorScheme derived from the painter's Theme.
func (p TabViewPainter) resolveColors() tabview.TabColorScheme {
	if p.Theme == nil {
		return cupDefaultTabColors
	}
	cs := p.Theme.Colors
	return tabview.TabColorScheme{
		Background:      cs.SecondarySystemBackground,
		SelectedText:    cs.Label,
		UnselectedText:  cs.SecondaryLabel,
		Indicator:       cs.TertiarySystemBackground,
		HoverBackground: cs.SystemFill,
		CloseButton:     cs.SecondaryLabel,
		FocusRing:       cs.Accent.WithAlpha(cupTabFocusAlpha),
	}
}

// PaintTabBar renders a segmented control according to Apple HIG specifications.
func (p TabViewPainter) PaintTabBar(canvas widget.Canvas, ps tabview.PaintState) {
	if ps.Bounds.IsEmpty() {
		return
	}

	colors := ps.ColorScheme
	if colors == (tabview.TabColorScheme{}) {
		colors = p.resolveColors()
	}

	// Segmented control background (pill shape).
	canvas.DrawRoundRect(ps.Bounds, colors.Background, cupTabSegmentRadius)

	// Draw selected segment indicator first (behind text).
	if ps.SelectedIdx >= 0 && ps.SelectedIdx < len(ps.Tabs) {
		selected := &ps.Tabs[ps.SelectedIdx]
		if !selected.Bounds.IsEmpty() {
			segInset := cupTabSegmentInset
			segBounds := geometry.NewRect(
				selected.Bounds.Min.X+segInset,
				selected.Bounds.Min.Y+segInset,
				selected.Bounds.Width()-segInset*2,
				selected.Bounds.Height()-segInset*2,
			)
			canvas.DrawRoundRect(segBounds, colors.Indicator, cupTabSelectedRadius)
		}
	}

	// Draw each tab.
	for i := range ps.Tabs {
		ts := &ps.Tabs[i]
		cupPaintTab(canvas, ts, colors)
	}

	// Focus ring around entire segmented control.
	if ps.Focused {
		ringBounds := ps.Bounds.Expand(cupTabFocusRingOffset)
		canvas.StrokeRoundRect(ringBounds, colors.FocusRing, cupTabSegmentRadius+cupTabFocusRingOffset, cupTabFocusRingStroke)
	}
}

// cupPaintTab draws a single tab segment with Cupertino styling.
func cupPaintTab(canvas widget.Canvas, ts *tabview.TabState, colors tabview.TabColorScheme) {
	if ts.Bounds.IsEmpty() {
		return
	}

	// Hover overlay.
	if ts.Hovered && !ts.Selected && !ts.Disabled {
		canvas.DrawRoundRect(ts.Bounds, colors.HoverBackground, cupTabSelectedRadius)
	}

	// Text color.
	fg := colors.UnselectedText
	if ts.Selected {
		fg = colors.SelectedText
	}
	if ts.Disabled {
		fg = colors.UnselectedText.WithAlpha(cupTabDisabledAlpha)
	}

	// Text bounds (shrink if closeable).
	textBounds := ts.Bounds
	if ts.Closeable {
		textBounds = geometry.NewRect(
			ts.Bounds.Min.X,
			ts.Bounds.Min.Y,
			ts.Bounds.Width()-cupTabCloseButtonSize-cupTabCloseButtonPad,
			ts.Bounds.Height(),
		)
	}

	canvas.DrawText(ts.Label, textBounds, cupTabFontSize, fg, ts.Selected, cupTabTextAlign)

	// Close button.
	if ts.Closeable && !ts.Disabled && !ts.CloseButtonBounds.IsEmpty() {
		cupPaintCloseButton(canvas, ts.CloseButtonBounds, colors.CloseButton)
	}
}

// cupPaintCloseButton draws a small X icon.
func cupPaintCloseButton(canvas widget.Canvas, bounds geometry.Rect, color widget.Color) {
	centerX := (bounds.Min.X + bounds.Max.X) / 2
	centerY := (bounds.Min.Y + bounds.Max.Y) / 2
	halfSize := cupTabCloseIconHalf

	canvas.DrawLine(
		geometry.Pt(centerX-halfSize, centerY-halfSize),
		geometry.Pt(centerX+halfSize, centerY+halfSize),
		color, cupTabCloseStroke,
	)
	canvas.DrawLine(
		geometry.Pt(centerX+halfSize, centerY-halfSize),
		geometry.Pt(centerX-halfSize, centerY+halfSize),
		color, cupTabCloseStroke,
	)
}

// cupDefaultTabColors holds the default Cupertino segmented control colors.
var cupDefaultTabColors = tabview.TabColorScheme{
	Background:      widget.Hex(0xF2F2F7),
	SelectedText:    widget.RGBA(0.0, 0.0, 0.0, 1.0),
	UnselectedText:  widget.RGBA(0.235, 0.235, 0.263, 0.6),
	Indicator:       widget.ColorWhite,
	HoverBackground: widget.RGBA(0.47, 0.47, 0.50, 0.12),
	CloseButton:     widget.RGBA(0.235, 0.235, 0.263, 0.6),
	FocusRing:       systemBlue.WithAlpha(0.6),
}

// Cupertino segmented control drawing constants.
const (
	cupTabSegmentRadius   float32 = 8
	cupTabSelectedRadius  float32 = 7
	cupTabSegmentInset    float32 = 2
	cupTabFontSize        float32 = 13
	cupTabTextAlign               = widget.TextAlignCenter
	cupTabDisabledAlpha   float32 = 0.3
	cupTabFocusAlpha      float32 = 0.6
	cupTabFocusRingOffset float32 = 3
	cupTabFocusRingStroke float32 = 2.5
	cupTabCloseButtonSize float32 = 16
	cupTabCloseButtonPad  float32 = 8
	cupTabCloseIconHalf   float32 = 4
	cupTabCloseStroke     float32 = 1.5
)

// Compile-time check that TabViewPainter implements Painter.
var _ tabview.Painter = TabViewPainter{}
