package devtools

import (
	"github.com/gogpu/ui/core/scrollview"
	"github.com/gogpu/ui/widget"
)

// ScrollbarPainter renders scrollbars using DevTools design tokens.
// DevTools scrollbars are thin (6px width, 4px idle expanding to 6px on hover)
// with transparent track — matching JetBrains IDE editor scrollbar styling.
//
// If Theme is nil, ScrollbarPainter falls back to the default DevTools dark palette.
type ScrollbarPainter struct {
	Theme *Theme // nil uses default DevTools dark fallback
}

// resolveColors returns the ScrollbarColorScheme derived from the painter's Theme.
func (p ScrollbarPainter) resolveColors() scrollview.ScrollbarColorScheme {
	if p.Theme == nil {
		return dtDefaultScrollbarColors
	}
	cs := p.Theme.Colors
	return scrollview.ScrollbarColorScheme{
		Track:      cs.ScrollbarTrack,
		Thumb:      cs.ScrollbarThumb,
		ThumbHover: cs.ScrollbarThumbHover,
		ThumbDrag:  cs.ScrollbarThumbHover,
	}
}

// PaintScrollbar renders scrollbars according to DevTools design specifications.
func (p ScrollbarPainter) PaintScrollbar(canvas widget.Canvas, ps scrollview.PaintState) {
	if ps.Bounds.IsEmpty() {
		return
	}

	colors := ps.ColorScheme
	if colors == (scrollview.ScrollbarColorScheme{}) {
		colors = p.resolveColors()
	}

	// Override paint state with resolved colors and delegate to the default painter.
	ps.ColorScheme = colors
	scrollview.DefaultPainter{}.PaintScrollbar(canvas, ps)
}

// dtDefaultScrollbarColors holds the default DevTools dark scrollbar color scheme.
var dtDefaultScrollbarColors = scrollview.ScrollbarColorScheme{
	Track:      widget.ColorTransparent,
	Thumb:      widget.Hex(0x4E5157), // Gray5 (ScrollbarThumb)
	ThumbHover: widget.Hex(0x6F737A), // Gray7 (ScrollbarThumbHover)
	ThumbDrag:  widget.Hex(0x6F737A), // Gray7
}

// Compile-time check that ScrollbarPainter implements Painter.
var _ scrollview.Painter = ScrollbarPainter{}
