package devtools

import (
	"github.com/gogpu/ui/core/splitview"
	"github.com/gogpu/ui/widget"
)

// SplitViewPainter renders split view dividers using DevTools design tokens.
// DevTools split views use a minimal 1px divider line (Gray3) with a 3px
// invisible drag zone (cursor change on hover), no shadow and no handle
// decoration. On hover the divider brightens to Gray5, matching JetBrains
// IDE panel divider styling.
//
// If Theme is nil, SplitViewPainter falls back to the default DevTools dark palette.
type SplitViewPainter struct {
	Theme *Theme // nil uses default DevTools dark fallback
}

// resolveColors returns the DividerColorScheme derived from the painter's Theme.
func (p SplitViewPainter) resolveColors() splitview.DividerColorScheme {
	if p.Theme == nil {
		return dtDefaultSplitViewColors
	}
	cs := p.Theme.Colors
	return splitview.DividerColorScheme{
		Divider:      cs.Border,
		DividerHover: cs.BorderStrong,
		DividerDrag:  cs.Primary.WithAlpha(0.5),
		Handle:       cs.OnSurfaceSecondary,
	}
}

// PaintDivider renders a split view divider according to DevTools specifications.
// DevTools dividers are minimal: a single line with no handle decoration.
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
	bgColor := dtResolvedSplitDividerColor(ps, colors)

	// Draw divider as a simple line (no handle decoration in DevTools style).
	canvas.DrawRect(ps.DividerRect, bgColor)
}

// dtResolvedSplitDividerColor returns the divider color based on interaction state.
func dtResolvedSplitDividerColor(ps splitview.PaintState, colors splitview.DividerColorScheme) widget.Color {
	if ps.Dragging {
		return colors.DividerDrag
	}
	if ps.Hovered {
		return colors.DividerHover
	}
	return colors.Divider
}

// dtDefaultSplitViewColors holds the default DevTools dark color scheme for split view dividers.
var dtDefaultSplitViewColors = splitview.DividerColorScheme{
	Divider:      widget.Hex(0x393B40),                // Gray3 (border)
	DividerHover: widget.Hex(0x4E5157),                // Gray5 (border strong)
	DividerDrag:  widget.Hex(0x3574F0).WithAlpha(0.5), // Blue6 @ 50%
	Handle:       widget.Hex(0x9DA0A8),                // Gray9 (secondary)
}

// Compile-time check that SplitViewPainter implements Painter.
var _ splitview.Painter = SplitViewPainter{}
