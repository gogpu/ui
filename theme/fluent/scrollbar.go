package fluent

import (
	"github.com/gogpu/ui/core/scrollview"
	"github.com/gogpu/ui/widget"
)

// ScrollbarPainter renders scrollbars using Fluent Design tokens.
// Fluent scrollbars are thin and subtle, appearing on hover.
//
// If Theme is nil, ScrollbarPainter falls back to the default Fluent scrollbar palette.
type ScrollbarPainter struct {
	Theme *Theme // nil uses default Fluent fallback
}

// resolveColors returns the ScrollbarColorScheme derived from the painter's Theme.
func (p ScrollbarPainter) resolveColors() scrollview.ScrollbarColorScheme {
	if p.Theme == nil {
		return flDefaultScrollbarColors
	}
	cs := p.Theme.Colors
	return scrollview.ScrollbarColorScheme{
		Track:      cs.FillTertiary,
		Thumb:      cs.OnSurfaceSecond.WithAlpha(0.4),
		ThumbHover: cs.OnSurfaceSecond.WithAlpha(0.6),
		ThumbDrag:  cs.OnSurfaceSecond.WithAlpha(0.8),
	}
}

// PaintScrollbar renders scrollbars according to Fluent Design specifications.
func (p ScrollbarPainter) PaintScrollbar(canvas widget.Canvas, ps scrollview.PaintState) {
	if ps.Bounds.IsEmpty() {
		return
	}

	colors := ps.ColorScheme
	if colors == (scrollview.ScrollbarColorScheme{}) {
		colors = p.resolveColors()
	}

	// Override paint state with resolved colors and delegate.
	ps.ColorScheme = colors
	scrollview.DefaultPainter{}.PaintScrollbar(canvas, ps)
}

// flDefaultScrollbarColors holds the default Fluent scrollbar color scheme.
var flDefaultScrollbarColors = scrollview.ScrollbarColorScheme{
	Track:      widget.RGBA(0, 0, 0, 0.03),
	Thumb:      widget.RGBA(0.38, 0.38, 0.38, 0.4),
	ThumbHover: widget.RGBA(0.38, 0.38, 0.38, 0.6),
	ThumbDrag:  widget.RGBA(0.38, 0.38, 0.38, 0.8),
}

// Compile-time check that ScrollbarPainter implements Painter.
var _ scrollview.Painter = ScrollbarPainter{}
