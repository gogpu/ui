package chip

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// Painter draws the visual representation of a chip.
// Each design system (Material 3, Fluent, Cupertino) provides its own
// Painter implementation. If no Painter is set, the chip uses [DefaultPainter].
type Painter interface {
	PaintChip(canvas widget.Canvas, state PaintState)
}

// LayoutMetrics allows theme painters to provide spatial metrics used by the
// widget's Layout method to compute chip dimensions.
//
// Painters that implement this interface provide custom metrics.
// Painters that do not implement it get default values from [DefaultPainter].
type LayoutMetrics interface {
	// ChipFontSize returns the font size for the chip label.
	ChipFontSize() float32

	// ChipMinWidth returns the minimum chip width in logical pixels.
	ChipMinWidth() float32

	// ChipPadding returns the horizontal padding inside the chip.
	ChipPadding() float32

	// ChipRadius returns the corner radius for the chip.
	ChipRadius() float32
}

// PaintState provides the current chip state to the painter.
type PaintState struct {
	Label       string          // chip label text
	Bounds      geometry.Rect   // total widget bounds (excludes outer padding)
	Radius      float32         // corner radius
	FontSize    float32         // label font size in logical pixels (0 = painter default)
	Selectable  bool            // chip toggles a selected state
	Selected    bool            // chip is currently selected
	Hovered     bool            // pointer is over the chip
	Pressed     bool            // chip is being pressed
	Focused     bool            // chip has keyboard focus
	Disabled    bool            // chip is disabled
	ColorScheme ChipColorScheme // theme-derived colors (zero = use defaults)
}

// DefaultPainter provides a minimal fallback painter with Material-3-flavored
// styling: an outlined pill when unselected and a filled pill when selected,
// with a translucent state layer for hover and press.
type DefaultPainter struct{}

// PaintChip renders the chip. If state.ColorScheme is non-zero, its colors are
// used instead of the built-in defaults.
func (p DefaultPainter) PaintChip(canvas widget.Canvas, ps PaintState) {
	if ps.Bounds.IsEmpty() {
		return
	}

	hasScheme := ps.ColorScheme != (ChipColorScheme{})
	bounds := ps.Bounds
	radius := ps.Radius

	labelColor := resolveLabelColor(ps, hasScheme)

	switch {
	case ps.Disabled:
		canvas.DrawRoundRect(bounds, resolveDisabledBackground(ps, hasScheme), radius)
	case ps.Selected:
		canvas.DrawRoundRect(bounds, resolveSelectedBackground(ps, hasScheme), radius)
	default:
		// Outlined: optional fill plus a border.
		if bg := resolveBackground(ps, hasScheme); bg.A > 0 {
			canvas.DrawRoundRect(bounds, bg, radius)
		}
		canvas.StrokeRoundRect(bounds, resolveBorder(ps, hasScheme), radius, borderWidth)
	}

	// State layer for hover/press (skipped when disabled).
	if !ps.Disabled {
		if alpha := stateLayerAlpha(ps); alpha > 0 {
			canvas.DrawRoundRect(bounds, labelColor.WithAlpha(alpha), radius)
		}
	}

	if ps.Label != "" {
		fontSize := ps.FontSize
		if fontSize <= 0 {
			fontSize = defaultFontSize
		}
		canvas.DrawText(ps.Label, bounds, fontSize, labelColor, false, widget.TextAlignCenter)
	}

	if ps.Focused && !ps.Disabled {
		ringBounds := bounds.Expand(focusRingOffset)
		canvas.StrokeRoundRect(ringBounds, focusRingColor, radius+focusRingOffset, focusRingStrokeWidth)
	}
}

// stateLayerAlpha returns the translucent overlay alpha for the current
// interaction state, following Material 3 state-layer opacities.
func stateLayerAlpha(ps PaintState) float32 {
	switch {
	case ps.Pressed:
		return pressStateAlpha
	case ps.Hovered:
		return hoverStateAlpha
	default:
		return 0
	}
}

// Color resolution helpers.

func resolveBackground(ps PaintState, hasScheme bool) widget.Color {
	if hasScheme {
		return ps.ColorScheme.Background
	}
	return defaultBackground
}

func resolveBorder(ps PaintState, hasScheme bool) widget.Color {
	if hasScheme {
		return ps.ColorScheme.Border
	}
	return defaultBorder
}

func resolveLabelColor(ps PaintState, hasScheme bool) widget.Color {
	if ps.Disabled {
		if hasScheme {
			return ps.ColorScheme.DisabledLabel
		}
		return defaultDisabledLabel
	}
	if ps.Selected {
		if hasScheme {
			return ps.ColorScheme.SelectedLabel
		}
		return defaultSelectedLabel
	}
	if hasScheme {
		return ps.ColorScheme.Label
	}
	return defaultLabel
}

func resolveSelectedBackground(ps PaintState, hasScheme bool) widget.Color {
	if hasScheme {
		return ps.ColorScheme.SelectedBackground
	}
	return defaultSelectedBackground
}

func resolveDisabledBackground(ps PaintState, hasScheme bool) widget.Color {
	if hasScheme {
		return ps.ColorScheme.DisabledBackground
	}
	return defaultDisabledBackground
}

// ChipFontSize returns the default font size.
func (DefaultPainter) ChipFontSize() float32 { return defaultFontSize }

// ChipMinWidth returns the default minimum chip width.
func (DefaultPainter) ChipMinWidth() float32 { return minChipWidth }

// ChipPadding returns the default horizontal padding.
func (DefaultPainter) ChipPadding() float32 { return labelPaddingX }

// ChipRadius returns the default corner radius.
func (DefaultPainter) ChipRadius() float32 { return defaultChipRadius }

// resolveChipLayoutMetrics returns the LayoutMetrics from the painter if it
// implements that interface, otherwise returns DefaultPainter metrics.
func resolveChipLayoutMetrics(p Painter) LayoutMetrics {
	if lm, ok := p.(LayoutMetrics); ok {
		return lm
	}
	return DefaultPainter{}
}

// Compile-time checks.
var _ LayoutMetrics = DefaultPainter{}

// Painting constants.
const (
	borderWidth          float32 = 1
	focusRingOffset      float32 = 2
	focusRingStrokeWidth float32 = 2
	hoverStateAlpha      float32 = 0.08
	pressStateAlpha      float32 = 0.12
)

// focusRingColor is the default color for focus indicators.
var focusRingColor = widget.Hex(0x6750A4).WithAlpha(0.7)

// Default colors for DefaultPainter (Material 3 flavored).
var (
	defaultBackground         = widget.RGBA(0, 0, 0, 0) // transparent (outlined)
	defaultBorder             = widget.Hex(0x79747E)    // M3 outline
	defaultLabel              = widget.Hex(0x1D1B20)    // M3 on-surface
	defaultSelectedBackground = widget.Hex(0xE8DEF8)    // M3 secondary container
	defaultSelectedLabel      = widget.Hex(0x1D192B)    // M3 on-secondary container
	defaultDisabledBackground = widget.RGBA(0.88, 0.88, 0.88, 1.0)
	defaultDisabledLabel      = widget.RGBA(0.60, 0.60, 0.60, 1.0)
)
