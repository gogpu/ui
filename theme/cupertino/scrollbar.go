package cupertino

import (
	"github.com/gogpu/ui/core/scrollview"
	"github.com/gogpu/ui/widget"
)

// ScrollbarPainter renders scrollbars using Apple HIG design tokens.
// Cupertino scrollbars are very thin (3px), use rounded ends, and appear
// only during scrolling with a fade-out effect. They overlay content
// rather than taking up layout space.
//
// If Theme is nil, ScrollbarPainter falls back to the default Cupertino palette.
type ScrollbarPainter struct {
	Theme *Theme // nil uses default fallback
}

// resolveColors returns the ScrollbarColorScheme derived from the painter's Theme.
func (p ScrollbarPainter) resolveColors() scrollview.ScrollbarColorScheme {
	if p.Theme == nil {
		return cupDefaultScrollbarColors
	}
	cs := p.Theme.Colors
	return scrollview.ScrollbarColorScheme{
		Track:      widget.ColorTransparent,
		Thumb:      cs.Label.WithAlpha(cupSBThumbAlpha),
		ThumbHover: cs.Label.WithAlpha(cupSBThumbHoverAlpha),
		ThumbDrag:  cs.Label.WithAlpha(cupSBThumbDragAlpha),
	}
}

// PaintScrollbar renders thin Cupertino-style scrollbars.
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

// cupDefaultScrollbarColors holds the default Cupertino scrollbar colors.
// Track is transparent (overlay scrollbars), thumb is semi-transparent.
var cupDefaultScrollbarColors = scrollview.ScrollbarColorScheme{
	Track:      widget.ColorTransparent,
	Thumb:      widget.RGBA(0, 0, 0, 0.3),
	ThumbHover: widget.RGBA(0, 0, 0, 0.5),
	ThumbDrag:  widget.RGBA(0, 0, 0, 0.6),
}

// Cupertino scrollbar constants.
const (
	cupSBThumbAlpha      float32 = 0.3
	cupSBThumbHoverAlpha float32 = 0.5
	cupSBThumbDragAlpha  float32 = 0.6
)

// Compile-time check that ScrollbarPainter implements Painter.
var _ scrollview.Painter = ScrollbarPainter{}
