package collapsible

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// Painter draws the visual representation of a collapsible section header.
// Each design system (Material 3, Fluent, Cupertino) provides its own
// Painter implementation.
//
// If no Painter is set, the section uses [DefaultPainter].
type Painter interface {
	PaintHeader(canvas widget.Canvas, state HeaderState)
}

// HeaderState provides the current state to the painter for header rendering.
type HeaderState struct {
	Title         string
	Expanded      bool
	Hovered       bool
	Pressed       bool
	Focused       bool
	Bounds        geometry.Rect
	ArrowProgress float32 // 0.0 = collapsed (right arrow), 1.0 = expanded (down arrow)

	// Styling overrides (zero value means use painter defaults).
	HeaderColor widget.Color
	ArrowColor  widget.Color
}

// DefaultPainter provides a minimal fallback header painter.
// It draws a background, title text, and an arrow indicator.
type DefaultPainter struct{}

// PaintHeader renders a minimal collapsible header.
func (p DefaultPainter) PaintHeader(canvas widget.Canvas, s HeaderState) {
	if s.Bounds.IsEmpty() {
		return
	}

	bg := resolveHeaderBg(s)
	bg = applyStateModifier(bg, s.Hovered, s.Pressed)
	canvas.DrawRect(s.Bounds, bg)

	// Draw arrow indicator.
	arrowColor := resolveArrowColor(s)
	drawArrow(canvas, s.Bounds, arrowColor, s.ArrowProgress)

	// Draw title text.
	if s.Title != "" {
		titleBounds := titleTextBounds(s.Bounds)
		canvas.DrawText(s.Title, titleBounds, defaultFontSize, defaultTitleColor, true, textAlignLeft)
	}

	// Focus ring.
	if s.Focused {
		canvas.StrokeRect(s.Bounds, focusRingColor, focusRingStrokeWidth)
	}
}

// resolveHeaderBg returns the header background color.
func resolveHeaderBg(s HeaderState) widget.Color {
	if s.HeaderColor != (widget.Color{}) {
		return s.HeaderColor
	}
	return defaultHeaderBg
}

// resolveArrowColor returns the arrow indicator color.
func resolveArrowColor(s HeaderState) widget.Color {
	if s.ArrowColor != (widget.Color{}) {
		return s.ArrowColor
	}
	return defaultArrowColor
}

// applyStateModifier adjusts a color based on interaction state.
func applyStateModifier(base widget.Color, hovered, pressed bool) widget.Color {
	if pressed {
		return base.Lerp(widget.ColorBlack, pressedDarkenFactor)
	}
	if hovered {
		return base.Lerp(widget.ColorWhite, hoverLightenFactor)
	}
	return base
}

// drawArrow draws a triangular arrow indicator.
// progress 0.0 = right-pointing, 1.0 = down-pointing.
func drawArrow(canvas widget.Canvas, headerBounds geometry.Rect, color widget.Color, progress float32) {
	h := headerBounds.Height()
	arrowSize := h * arrowSizeRatio
	cx := headerBounds.Min.X + arrowPadding + arrowSize/2
	cy := headerBounds.Min.Y + h/2

	// Interpolate between right arrow (>) and down arrow (v).
	// Right arrow: tip right, base left.
	// Down arrow: tip down, base top.
	half := arrowSize / 2

	// Right arrow points: left-top, right-center, left-bottom.
	// Down arrow points: left-top, center-bottom, right-top.
	// We interpolate between these.
	p1x := cx - half + progress*half
	p1y := cy - half + progress*half
	p2x := cx + half - progress*half*2
	p2y := cy + progress*half
	p3x := cx - half + progress*half*2
	p3y := cy + half - progress*half

	canvas.DrawLine(geometry.Pt(p1x, p1y), geometry.Pt(p2x, p2y), color, arrowStrokeWidth)
	canvas.DrawLine(geometry.Pt(p2x, p2y), geometry.Pt(p3x, p3y), color, arrowStrokeWidth)
}

// titleTextBounds returns the bounds for the title text within the header.
func titleTextBounds(headerBounds geometry.Rect) geometry.Rect {
	return geometry.NewRect(
		headerBounds.Min.X+titleLeftOffset,
		headerBounds.Min.Y,
		headerBounds.Width()-titleLeftOffset-titleRightPadding,
		headerBounds.Height(),
	)
}

// Painting constants.
const (
	defaultFontSize      float32 = 14
	textAlignLeft                = widget.TextAlignLeft
	arrowPadding         float32 = 8
	arrowSizeRatio       float32 = 0.35
	arrowStrokeWidth     float32 = 2
	titleLeftOffset      float32 = 32
	titleRightPadding    float32 = 8
	focusRingStrokeWidth float32 = 2
	hoverLightenFactor   float32 = 0.1
	pressedDarkenFactor  float32 = 0.15
)

// Default colors for DefaultPainter.
var (
	defaultHeaderBg   = widget.RGBA(0.93, 0.93, 0.93, 1.0)
	defaultArrowColor = widget.RGBA(0.3, 0.3, 0.3, 1.0)
	defaultTitleColor = widget.RGBA(0.1, 0.1, 0.1, 1.0)
	focusRingColor    = widget.Hex(0x6750A4).WithAlpha(0.7)
)
