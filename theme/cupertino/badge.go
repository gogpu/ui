package cupertino

import (
	"github.com/gogpu/ui/core/badge"
	"github.com/gogpu/ui/widget"
)

// BadgePainter renders badges using Apple HIG (Cupertino) design tokens.
// Cupertino badges use System Red for the background with white text,
// matching the iOS/macOS notification badge style.
//
// If Theme is nil, BadgePainter falls back to the default Cupertino light palette.
type BadgePainter struct {
	Theme *Theme // nil uses default Cupertino fallback
}

// resolveColors returns the BadgeColorScheme derived from the painter's Theme.
// If Theme is nil, it returns the default Cupertino badge color scheme.
func (p BadgePainter) resolveColors() badge.BadgeColorScheme {
	if p.Theme == nil {
		return cupDefaultBadgeColors
	}
	cs := p.Theme.Colors
	return badge.BadgeColorScheme{
		Background:         cs.SystemRed,
		Label:              widget.ColorWhite,
		DisabledBackground: cs.QuaternaryLabel,
		DisabledLabel:      cs.TertiaryLabel,
	}
}

// PaintBadge renders a badge according to Cupertino specifications.
// It delegates to the core [badge.DefaultPainter] with Cupertino-derived colors.
func (p BadgePainter) PaintBadge(canvas widget.Canvas, ps badge.PaintState) {
	if ps.Bounds.IsEmpty() {
		return
	}

	// Use the color scheme from PaintState if provided, otherwise resolve from theme.
	colors := ps.ColorScheme
	if colors == (badge.BadgeColorScheme{}) {
		colors = p.resolveColors()
	}

	ps.ColorScheme = colors
	badge.DefaultPainter{}.PaintBadge(canvas, ps)
}

// cupDefaultBadgeColors holds the default Cupertino badge color scheme.
// Uses System Red (#FF3B30) with white text, matching iOS notification badges.
var cupDefaultBadgeColors = badge.BadgeColorScheme{
	Background:         widget.Hex(0xFF3B30),                   // System Red (light)
	Label:              widget.ColorWhite,                      // white text
	DisabledBackground: widget.RGBA(0.235, 0.235, 0.263, 0.18), // quaternary label (light)
	DisabledLabel:      widget.RGBA(0.235, 0.235, 0.263, 0.3),  // tertiary label (light)
}

// Compile-time check that BadgePainter implements Painter.
var _ badge.Painter = BadgePainter{}
