package fluent

import (
	"github.com/gogpu/ui/core/badge"
	"github.com/gogpu/ui/widget"
)

// BadgePainter renders badges using Fluent Design tokens.
// Fluent badges use the accent color for the background with onAccent text,
// following the Microsoft Fluent Design notification badge pattern.
//
// If Theme is nil, BadgePainter falls back to the default Fluent (Windows Blue) palette.
type BadgePainter struct {
	Theme *Theme // nil uses default Fluent fallback
}

// resolveColors returns the BadgeColorScheme derived from the painter's Theme.
// If Theme is nil, it returns the default Fluent badge color scheme.
func (p BadgePainter) resolveColors() badge.BadgeColorScheme {
	if p.Theme == nil {
		return flDefaultBadgeColors
	}
	cs := p.Theme.Colors
	return badge.BadgeColorScheme{
		Background:         cs.Accent,
		Label:              cs.OnAccent,
		DisabledBackground: cs.FillDisable,
		DisabledLabel:      cs.OnSurfaceSecond,
	}
}

// PaintBadge renders a badge according to Fluent Design specifications.
// It delegates to the core [badge.DefaultPainter] with Fluent-derived colors.
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

// flDefaultBadgeColors holds the default Fluent badge color scheme.
// Uses Windows Blue accent with white text.
var flDefaultBadgeColors = badge.BadgeColorScheme{
	Background:         widget.Hex(0x0078D4),       // Windows Blue accent
	Label:              widget.ColorWhite,          // on-accent
	DisabledBackground: widget.RGBA(0, 0, 0, 0.04), // fill-disable (light)
	DisabledLabel:      widget.Hex(0x616161),       // on-surface-second (light)
}

// Compile-time check that BadgePainter implements Painter.
var _ badge.Painter = BadgePainter{}
