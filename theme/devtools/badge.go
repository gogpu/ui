package devtools

import (
	"github.com/gogpu/ui/core/badge"
	"github.com/gogpu/ui/widget"
)

// BadgePainter renders badges using DevTools (JetBrains Int UI) design tokens.
// DevTools badges use the error color (Red7) for the background with white text,
// matching JetBrains IDE notification badge styling.
//
// If Theme is nil, BadgePainter falls back to the default DevTools dark palette.
type BadgePainter struct {
	Theme *Theme // nil uses default DevTools dark fallback
}

// resolveColors returns the BadgeColorScheme derived from the painter's Theme.
// If Theme is nil, it returns the default DevTools badge color scheme.
func (p BadgePainter) resolveColors() badge.BadgeColorScheme {
	if p.Theme == nil {
		return dtDefaultBadgeColors
	}
	cs := p.Theme.Colors
	return badge.BadgeColorScheme{
		Background:         cs.Error,
		Label:              cs.OnPrimary,
		DisabledBackground: cs.OnSurfaceDisabled,
		DisabledLabel:      cs.OnSurfaceDisabled.WithAlpha(0.38),
	}
}

// PaintBadge renders a badge according to DevTools specifications.
// It delegates to the core [badge.DefaultPainter] with DevTools-derived colors.
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

// dtDefaultBadgeColors holds the default DevTools badge color scheme.
// Uses Red7 (error) for background with white text on dark theme.
var dtDefaultBadgeColors = badge.BadgeColorScheme{
	Background:         widget.Hex(0xDB5C5C),                // Red7 (error)
	Label:              widget.ColorWhite,                   // on-primary
	DisabledBackground: widget.Hex(0x6F737A),                // Gray7 (disabled)
	DisabledLabel:      widget.RGBA(0.44, 0.45, 0.48, 0.38), // Gray7 @ 38%
}

// Compile-time check that BadgePainter implements Painter.
var _ badge.Painter = BadgePainter{}
