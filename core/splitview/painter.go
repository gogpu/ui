package splitview

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// Painter draws the visual representation of the split view divider.
// Each design system (Material 3, Fluent, Cupertino) provides its own
// Painter implementation to render the divider in its visual style.
//
// If no Painter is set, the split view uses [DefaultPainter].
type Painter interface {
	PaintDivider(canvas widget.Canvas, state PaintState)
}

// PaintState provides the current divider state to the painter.
type PaintState struct {
	// DividerRect is the bounding rectangle of the divider area.
	DividerRect geometry.Rect

	// Orientation is the split orientation (Horizontal or Vertical).
	Orientation Orientation

	// Hovered indicates that the mouse is over the divider.
	Hovered bool

	// Dragging indicates that the divider is currently being dragged.
	Dragging bool

	// Collapsed indicates that the first panel is collapsed.
	Collapsed bool

	// ColorScheme provides theme-derived colors for divider painting.
	// Zero value means the painter should use its built-in defaults.
	ColorScheme DividerColorScheme
}

// DividerColorScheme provides theme-derived colors for divider painting.
// Zero value means the painter should use its built-in defaults.
type DividerColorScheme struct {
	Divider      widget.Color // divider background
	DividerHover widget.Color // divider background when hovered
	DividerDrag  widget.Color // divider background when dragging
	Handle       widget.Color // center handle indicator
}

// DefaultPainter provides a minimal fallback painter with no design system styling.
// It draws a simple divider line with a center handle indicator.
type DefaultPainter struct{}

// PaintDivider renders a minimal divider with a center handle.
func (p DefaultPainter) PaintDivider(canvas widget.Canvas, ps PaintState) {
	if ps.DividerRect.IsEmpty() {
		return
	}

	// Draw divider background.
	bgColor := resolveDividerColor(ps)
	canvas.DrawRect(ps.DividerRect, bgColor)

	// Draw center handle indicator (a small line/dot in the middle).
	handleColor := defaultHandleColor
	hasScheme := ps.ColorScheme != (DividerColorScheme{})
	if hasScheme {
		handleColor = ps.ColorScheme.Handle
	}

	paintHandle(canvas, ps.DividerRect, ps.Orientation, handleColor)
}

// resolveDividerColor returns the divider color based on interaction state.
func resolveDividerColor(ps PaintState) widget.Color {
	hasScheme := ps.ColorScheme != (DividerColorScheme{})

	if ps.Dragging {
		if hasScheme {
			return ps.ColorScheme.DividerDrag
		}
		return defaultDividerDragColor
	}
	if ps.Hovered {
		if hasScheme {
			return ps.ColorScheme.DividerHover
		}
		return defaultDividerHoverColor
	}
	if hasScheme {
		return ps.ColorScheme.Divider
	}
	return defaultDividerColor
}

// paintHandle draws a small handle indicator in the center of the divider.
func paintHandle(canvas widget.Canvas, divider geometry.Rect, orient Orientation, color widget.Color) {
	cx := (divider.Min.X + divider.Max.X) / 2
	cy := (divider.Min.Y + divider.Max.Y) / 2

	if orient == Horizontal {
		// Vertical line of dots for horizontal split.
		for i := -1; i <= 1; i++ {
			y := cy + float32(i)*handleSpacing
			canvas.DrawCircle(geometry.Pt(cx, y), handleRadius, color)
		}
	} else {
		// Horizontal line of dots for vertical split.
		for i := -1; i <= 1; i++ {
			x := cx + float32(i)*handleSpacing
			canvas.DrawCircle(geometry.Pt(x, cy), handleRadius, color)
		}
	}
}

// Handle dimensions.
const (
	handleRadius  float32 = 2 // radius of each handle dot
	handleSpacing float32 = 6 // spacing between handle dots
)

// Default colors for DefaultPainter.
var (
	defaultDividerColor      = widget.RGBA(0.85, 0.85, 0.85, 1)
	defaultDividerHoverColor = widget.RGBA(0.70, 0.70, 0.70, 1)
	defaultDividerDragColor  = widget.RGBA(0.55, 0.55, 0.55, 1)
	defaultHandleColor       = widget.RGBA(0.50, 0.50, 0.50, 0.8)
)
