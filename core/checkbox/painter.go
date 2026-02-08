package checkbox

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// Painter draws the visual representation of a checkbox.
// Each design system (Material 3, Fluent, Cupertino) provides its own
// Painter implementation to render the checkbox in its visual style.
//
// If no Painter is set, the checkbox uses [DefaultPainter].
type Painter interface {
	PaintCheckbox(canvas widget.Canvas, state PaintState)
}

// PaintState provides the current checkbox state to the painter.
type PaintState struct {
	Label         string
	Checked       bool
	Indeterminate bool
	Hovered       bool
	Pressed       bool
	Focused       bool
	Disabled      bool
	Bounds        geometry.Rect

	// Styling overrides (zero value means use design system defaults).
	Background  *widget.Color
	ColorScheme CheckboxColorScheme
}

// CheckboxColorScheme provides theme-derived colors for checkbox painting.
// Zero value means the painter should use its built-in defaults.
type CheckboxColorScheme struct {
	CheckedBg       widget.Color // Filled box background when checked
	CheckedFg       widget.Color // Checkmark color
	UncheckedBorder widget.Color // Border color when unchecked
	LabelColor      widget.Color // Label text color
	DisabledBg      widget.Color
	DisabledFg      widget.Color
	FocusRing       widget.Color
}

// DefaultPainter provides a minimal fallback painter with no design system styling.
// It draws a simple checkbox -- useful for testing and as a base reference.
type DefaultPainter struct{}

// PaintCheckbox renders a minimal checkbox with gray colors.
// If state.ColorScheme is non-zero, its colors are used instead of built-in defaults.
func (p DefaultPainter) PaintCheckbox(canvas widget.Canvas, state PaintState) {
	if state.Bounds.IsEmpty() {
		return
	}

	hasScheme := state.ColorScheme != (CheckboxColorScheme{})

	boxRect := checkboxBoxRect(state.Bounds)

	if state.Checked || state.Indeterminate {
		paintCheckedBox(canvas, boxRect, state, hasScheme)
	} else {
		paintUncheckedBox(canvas, boxRect, state, hasScheme)
	}

	// Draw label if present.
	if state.Label != "" {
		labelColor := resolveLabelColor(state, hasScheme)
		labelBounds := checkboxLabelBounds(state.Bounds)
		canvas.DrawText(state.Label, labelBounds, defaultFontSize, labelColor, false, textAlignLeft)
	}

	// Focus ring.
	if state.Focused && !state.Disabled {
		drawFocusIndicator(canvas, boxRect)
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
