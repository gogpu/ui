package material3

import (
	"github.com/gogpu/ui/core/splitview"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// SplitViewPainter renders split view dividers using Material 3 design tokens.
// It maps divider states (normal, hovered, dragging) to the M3 color scheme
// with outline-variant for the divider and primary color for the handle.
//
// If Theme is nil, SplitViewPainter falls back to the default M3 purple palette.
type SplitViewPainter struct {
	Theme *Theme // nil uses default M3 purple fallback
}

// resolveColors returns the DividerColorScheme derived from the painter's Theme.
// If Theme is nil, it returns the default M3 purple color scheme.
func (p SplitViewPainter) resolveColors() splitview.DividerColorScheme {
	if p.Theme == nil {
		return m3DefaultSplitViewColors
	}
	cs := p.Theme.Colors
	return splitview.DividerColorScheme{
		Divider:      cs.OutlineVariant,
		DividerHover: cs.Outline,
		DividerDrag:  cs.Primary.WithAlpha(0.3),
		Handle:       cs.OnSurfaceVariant,
	}
}

// PaintDivider renders a split view divider according to Material 3 specifications.
func (p SplitViewPainter) PaintDivider(canvas widget.Canvas, ps splitview.PaintState) {
	if ps.DividerRect.IsEmpty() {
		return
	}

	// Determine the color scheme to use.
	colors := ps.ColorScheme
	if colors == (splitview.DividerColorScheme{}) {
		colors = p.resolveColors()
	}

	// Resolve divider background color based on interaction state.
	bgColor := m3ResolvedSplitDividerColor(ps, colors)

	// Draw divider background.
	canvas.DrawRect(ps.DividerRect, bgColor)

	// Draw center handle indicator.
	m3PaintSplitHandle(canvas, ps.DividerRect, ps.Orientation, colors.Handle)
}

// m3ResolvedSplitDividerColor returns the divider color based on interaction state.
func m3ResolvedSplitDividerColor(ps splitview.PaintState, colors splitview.DividerColorScheme) widget.Color {
	if ps.Dragging {
		return colors.DividerDrag
	}
	if ps.Hovered {
		return colors.DividerHover
	}
	return colors.Divider
}

// m3PaintSplitHandle draws a handle indicator (three dots) in the center of the divider.
func m3PaintSplitHandle(canvas widget.Canvas, divider geometry.Rect, orient splitview.Orientation, color widget.Color) {
	cx := (divider.Min.X + divider.Max.X) / 2
	cy := (divider.Min.Y + divider.Max.Y) / 2

	if orient == splitview.Horizontal {
		// Vertical line of dots for horizontal split.
		for i := -1; i <= 1; i++ {
			y := cy + float32(i)*m3SplitHandleSpacing
			canvas.DrawCircle(geometry.Pt(cx, y), m3SplitHandleRadius, color)
		}
	} else {
		// Horizontal line of dots for vertical split.
		for i := -1; i <= 1; i++ {
			x := cx + float32(i)*m3SplitHandleSpacing
			canvas.DrawCircle(geometry.Pt(x, cy), m3SplitHandleRadius, color)
		}
	}
}

// m3DefaultSplitViewColors holds the default M3 purple color scheme for split view dividers.
// Used as a fallback when no Theme is provided.
var m3DefaultSplitViewColors = splitview.DividerColorScheme{
	Divider:      widget.Hex(0xCAC4D0),                // M3 outline variant
	DividerHover: widget.Hex(0x79747E),                // M3 outline
	DividerDrag:  widget.Hex(0x6750A4).WithAlpha(0.3), // M3 primary with low alpha
	Handle:       widget.Hex(0x49454F),                // M3 on-surface-variant
}

// M3 split view drawing constants.
const (
	m3SplitHandleRadius  float32 = 2.5
	m3SplitHandleSpacing float32 = 7
)

// Compile-time check that SplitViewPainter implements Painter.
var _ splitview.Painter = SplitViewPainter{}
