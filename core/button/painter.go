package button

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// Painter draws the visual representation of a button.
// Each design system (Material 3, Fluent, Cupertino) provides its own
// Painter implementation to render the button in its visual style.
//
// If no Painter is set, the button uses [DefaultPainter].
type Painter interface {
	PaintButton(canvas widget.Canvas, state PaintState)
}

// PaintState provides the current button state to the painter.
type PaintState struct {
	Text     string
	Variant  Variant
	Size     Size
	Hovered  bool
	Pressed  bool
	Focused  bool
	Disabled bool
	Bounds   geometry.Rect

	// Styling overrides (zero value means use design system defaults).
	Background  *widget.Color
	Radius      *float32
	ColorScheme ButtonColorScheme
}

// ButtonColorScheme provides theme-derived colors for button painting.
// Zero value means the painter should use its built-in defaults.
type ButtonColorScheme struct {
	FilledBg       widget.Color
	FilledFg       widget.Color
	OutlinedBorder widget.Color
	TextBgHover    widget.Color
	TonalBg        widget.Color
	TonalFg        widget.Color
	Primary        widget.Color
	DisabledBg     widget.Color
	DisabledFg     widget.Color
	FocusRing      widget.Color
}

// DefaultPainter provides a minimal fallback painter with no design system styling.
// It draws a simple gray button -- useful for testing and as a base reference.
type DefaultPainter struct{}

// PaintButton renders a minimal button with gray background and black text.
// If state.ColorScheme is non-zero, its colors are used instead of built-in defaults.
func (p DefaultPainter) PaintButton(canvas widget.Canvas, state PaintState) {
	if state.Bounds.IsEmpty() {
		return
	}

	radius := defaultRadius
	if state.Radius != nil {
		radius = *state.Radius
	}

	hasScheme := state.ColorScheme != (ButtonColorScheme{})

	// Background.
	bg := defaultBg
	if hasScheme {
		bg = state.ColorScheme.FilledBg
	}
	if state.Background != nil {
		bg = *state.Background
	}
	if state.Disabled {
		if hasScheme {
			bg = state.ColorScheme.DisabledBg
		} else {
			bg = defaultDisabledBg
		}
	}
	bg = applyStateModifier(bg, state.Hovered, state.Pressed)
	canvas.DrawRoundRect(state.Bounds, bg, radius)

	// Text.
	fg := defaultFg
	if hasScheme {
		fg = state.ColorScheme.FilledFg
	}
	if state.Disabled {
		if hasScheme {
			fg = state.ColorScheme.DisabledFg
		} else {
			fg = defaultDisabledFg
		}
	}
	fontSize := sizeFontSize(state.Size)
	canvas.DrawText(state.Text, state.Bounds, fontSize, fg, false, textAlignCenter)

	// Focus ring.
	if state.Focused && !state.Disabled {
		drawFocusIndicator(canvas, state.Bounds, radius)
	}
}

// Default colors for DefaultPainter.
var (
	defaultBg         = widget.RGBA(0.85, 0.85, 0.85, 1.0)
	defaultFg         = widget.RGBA(0.1, 0.1, 0.1, 1.0)
	defaultDisabledBg = widget.RGBA(0.92, 0.92, 0.92, 1.0)
	defaultDisabledFg = widget.RGBA(0.62, 0.62, 0.62, 1.0)
)
