package radio

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// Painter draws the visual representation of a radio item.
// Each design system (Material 3, Fluent, Cupertino) provides its own
// Painter implementation to render the radio item in its visual style.
//
// If no Painter is set, the radio group uses [DefaultPainter].
type Painter interface {
	PaintRadio(canvas widget.Canvas, state PaintState)
}

// PaintState provides the current radio item state to the painter.
type PaintState struct {
	Label    string
	Selected bool
	Hovered  bool
	Pressed  bool
	Focused  bool
	Disabled bool
	Bounds   geometry.Rect

	// ColorScheme provides theme-derived colors (zero value means use defaults).
	ColorScheme RadioColorScheme
}

// RadioColorScheme provides theme-derived colors for radio painting.
// Zero value means the painter should use its built-in defaults.
type RadioColorScheme struct {
	SelectedBg       widget.Color // Filled circle when selected
	SelectedFg       widget.Color // Inner dot color
	UnselectedBorder widget.Color // Circle border when unselected
	LabelColor       widget.Color
	DisabledBg       widget.Color
	DisabledFg       widget.Color
	FocusRing        widget.Color
}

// DefaultPainter provides a minimal fallback painter with no design system styling.
// It draws a simple radio button -- useful for testing and as a base reference.
type DefaultPainter struct{}

// PaintRadio renders a minimal radio item with gray colors.
// If state.ColorScheme is non-zero, its colors are used instead of built-in defaults.
func (p DefaultPainter) PaintRadio(canvas widget.Canvas, state PaintState) {
	if state.Bounds.IsEmpty() {
		return
	}

	hasScheme := state.ColorScheme != (RadioColorScheme{})

	circleCenter, circleRadius := radioCircleGeometry(state.Bounds)

	if state.Selected {
		paintSelectedRadio(canvas, circleCenter, circleRadius, state, hasScheme)
	} else {
		paintUnselectedRadio(canvas, circleCenter, circleRadius, state, hasScheme)
	}

	// Draw label if present.
	if state.Label != "" {
		labelColor := resolveLabelColor(state, hasScheme)
		labelBounds := radioLabelBounds(state.Bounds)
		canvas.DrawText(state.Label, labelBounds, defaultFontSize, labelColor, false, textAlignLeft)
	}

	// Focus ring.
	if state.Focused && !state.Disabled {
		drawFocusIndicator(canvas, circleCenter, circleRadius, state, hasScheme)
	}
}

// resolveLabelColor determines the label text color based on state and color scheme.
func resolveLabelColor(state PaintState, hasScheme bool) widget.Color {
	if state.Disabled && hasScheme {
		return state.ColorScheme.DisabledFg
	}
	if state.Disabled {
		return defaultDisabledFg
	}
	if hasScheme {
		return state.ColorScheme.LabelColor
	}
	return defaultLabelColor
}
