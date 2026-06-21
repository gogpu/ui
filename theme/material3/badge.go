package material3

import (
	"github.com/gogpu/ui/core/badge"
	"github.com/gogpu/ui/widget"
)

// BadgePainter renders badges using Material 3 design tokens.
// M3 badges use the error container color (red) with onError text (white),
// following the Material 3 specification for notification badges.
//
// If Theme is nil, BadgePainter falls back to the default M3 purple-derived palette.
type BadgePainter struct {
	Theme *Theme // nil uses default M3 fallback
}

// resolveColors returns the BadgeColorScheme derived from the painter's Theme.
// If Theme is nil, it returns the default M3 badge color scheme.
func (p BadgePainter) resolveColors() badge.BadgeColorScheme {
	if p.Theme == nil {
		return m3DefaultBadgeColors
	}
	cs := p.Theme.Colors
	return badge.BadgeColorScheme{
		Background:         cs.Error,
		Label:              cs.OnError,
		DisabledBackground: cs.OnSurface.WithAlpha(0.12),
		DisabledLabel:      cs.OnSurface.WithAlpha(0.38),
	}
}

// PaintBadge renders a badge according to Material 3 specifications.
// It delegates to the core [badge.DefaultPainter] with M3-derived colors.
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

// m3DefaultBadgeColors holds the default M3 badge color scheme.
// M3 badges use error color for background with onError text.
var m3DefaultBadgeColors = badge.BadgeColorScheme{
	Background:         widget.Hex(0xB3261E),                // M3 error
	Label:              widget.ColorWhite,                   // M3 on-error
	DisabledBackground: widget.RGBA(0.12, 0.12, 0.13, 0.12), // M3 on-surface @ 12%
	DisabledLabel:      widget.RGBA(0.12, 0.12, 0.13, 0.38), // M3 on-surface @ 38%
}

// Compile-time check that BadgePainter implements Painter.
var _ badge.Painter = BadgePainter{}
