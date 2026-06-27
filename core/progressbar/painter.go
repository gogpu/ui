package progressbar

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// Painter draws the visual representation of a progress bar.
// Each design system (Material 3, Fluent, Cupertino) provides its own
// Painter implementation to render the bar in its visual style.
//
// If no Painter is set, the progress bar uses [DefaultPainter].
type Painter interface {
	PaintProgressBar(canvas widget.Canvas, state PaintState)
}

// PaintState provides the current progress bar state to the painter.
type PaintState struct {
	Value       float64                // current value clamped to [0, 1]
	Bounds      geometry.Rect          // total widget bounds
	BarHeight   float32                // height of the bar in logical pixels
	Radius      float32                // corner radius for rounded ends
	ShowLabel   bool                   // whether to show percentage label
	Label       string                 // pre-formatted label text (empty if ShowLabel is false)
	Disabled    bool                   // widget is disabled
	ColorScheme ProgressBarColorScheme // theme-derived colors (zero = use defaults)
}

// DefaultPainter provides a minimal fallback painter with no design system styling.
// It draws a simple progress bar -- useful for testing and as a base reference.
type DefaultPainter struct{}

// PaintProgressBar renders a minimal progress bar with a gray track,
// blue fill, and optional centered label.
// If state.ColorScheme is non-zero, its colors are used instead of built-in defaults.
func (p DefaultPainter) PaintProgressBar(canvas widget.Canvas, ps PaintState) {
	if ps.Bounds.IsEmpty() {
		return
	}

	hasScheme := ps.ColorScheme != (ProgressBarColorScheme{})
	bounds := ps.Bounds

	// Calculate bar rect centered vertically in bounds.
	barY := bounds.Min.Y + (bounds.Height()-ps.BarHeight)/2
	barRect := geometry.NewRect(bounds.Min.X, barY, bounds.Width(), ps.BarHeight)

	// Draw track.
	trackColor := resolveTrackColor(ps, hasScheme)
	if ps.Radius > 0 {
		canvas.DrawRoundRect(barRect, trackColor, ps.Radius)
	} else {
		canvas.DrawRect(barRect, trackColor)
	}

	// Draw filled portion.
	fillWidth := barRect.Width() * float32(ps.Value)
	if fillWidth > 0 {
		barColor := resolveBarColor(ps, hasScheme)
		fillRect := geometry.NewRect(barRect.Min.X, barRect.Min.Y, fillWidth, ps.BarHeight)

		if ps.Radius > 0 {
			// Clip to the track's rounded rect so the fill never exceeds
			// the track shape, then draw the fill with rounded corners.
			canvas.PushClipRoundRect(barRect, ps.Radius)
			canvas.DrawRoundRect(fillRect, barColor, ps.Radius)
			canvas.PopClip()
		} else {
			canvas.DrawRect(fillRect, barColor)
		}
	}

	// Draw label if enabled.
	if ps.ShowLabel && ps.Label != "" {
		labelColor := resolveLabelColor(ps, hasScheme)
		canvas.DrawText(ps.Label, barRect, defaultFontSize, labelColor, false, labelAlignCenter)
	}
}

// Color resolution helpers.

func resolveTrackColor(ps PaintState, hasScheme bool) widget.Color {
	if ps.Disabled {
		if hasScheme {
			return ps.ColorScheme.DisabledTrack
		}
		return defaultDisabledTrack
	}
	if hasScheme {
		return ps.ColorScheme.Track
	}
	return defaultTrackColor
}

func resolveBarColor(ps PaintState, hasScheme bool) widget.Color {
	if ps.Disabled {
		if hasScheme {
			return ps.ColorScheme.DisabledBar
		}
		return defaultDisabledBar
	}
	if hasScheme {
		return ps.ColorScheme.Bar
	}
	return defaultBarColor
}

func resolveLabelColor(ps PaintState, hasScheme bool) widget.Color {
	if hasScheme {
		return ps.ColorScheme.Label
	}
	return defaultLabelColor
}

// Label alignment constant.
var labelAlignCenter = widget.TextAlignCenter

// Default colors for DefaultPainter.
var (
	defaultBarColor      = widget.Hex(0x6750A4) // Material primary
	defaultTrackColor    = widget.RGBA(0.90, 0.90, 0.90, 1.0)
	defaultLabelColor    = widget.ColorWhite
	defaultDisabledBar   = widget.RGBA(0.70, 0.70, 0.70, 1.0)
	defaultDisabledTrack = widget.RGBA(0.93, 0.93, 0.93, 1.0)
)
