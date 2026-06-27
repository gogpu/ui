package slider

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// Painter draws the visual representation of a slider.
// Each design system (Material 3, Fluent, Cupertino) provides its own
// Painter implementation to render the slider in its visual style.
//
// If no Painter is set, the slider uses [DefaultPainter].
type Painter interface {
	PaintSlider(canvas widget.Canvas, state PaintState)
}

// PaintState provides the current slider state to the painter.
type PaintState struct {
	Value       float32           // current value (within [Min, Max])
	Min         float32           // minimum value
	Max         float32           // maximum value
	Progress    float32           // normalized 0..1 progress
	Hovered     bool              // mouse is over the slider
	Dragging    bool              // thumb is being dragged
	Focused     bool              // slider has keyboard focus
	Disabled    bool              // slider is disabled
	Bounds      geometry.Rect     // total widget bounds
	Orientation Orientation       // horizontal or vertical
	Marks       []Mark            // optional tick marks
	ColorScheme SliderColorScheme // theme-derived colors
}

// SliderColorScheme provides theme-derived colors for slider painting.
// Zero value means the painter should use its built-in defaults.
type SliderColorScheme struct {
	ActiveTrack   widget.Color // active (filled) portion of the track
	InactiveTrack widget.Color // inactive (empty) portion of the track
	Thumb         widget.Color // thumb circle fill
	ThumbBorder   widget.Color // thumb circle border
	FocusRing     widget.Color // focus indicator color
	DisabledTrack widget.Color // track color when disabled
	DisabledThumb widget.Color // thumb color when disabled
	MarkColor     widget.Color // tick mark color
}

// LayoutMetrics allows theme painters to provide spatial metrics used by the
// widget's Layout method to compute thumb and track dimensions.
//
// Painters that implement this interface provide custom metrics.
// Painters that do not implement it get default values from [DefaultPainter].
type LayoutMetrics interface {
	// SliderThumbRadius returns the thumb circle radius in logical pixels.
	SliderThumbRadius() float32

	// SliderTrackHeight returns the track bar height in logical pixels.
	SliderTrackHeight() float32
}

// DefaultPainter provides a minimal fallback painter with no design system styling.
// It draws a simple slider -- useful for testing and as a base reference.
//
// DefaultPainter also implements [LayoutMetrics], providing the default spatial
// values used when a painter does not implement that interface.
type DefaultPainter struct{}

// PaintSlider renders a minimal slider with gray track, blue active segment,
// and a white thumb circle with border.
// If state.ColorScheme is non-zero, its colors are used instead of built-in defaults.
func (p DefaultPainter) PaintSlider(canvas widget.Canvas, ps PaintState) {
	if ps.Bounds.IsEmpty() {
		return
	}

	hasScheme := ps.ColorScheme != (SliderColorScheme{})

	if ps.Orientation == Vertical {
		paintVerticalSlider(canvas, ps, hasScheme)
	} else {
		paintHorizontalSlider(canvas, ps, hasScheme)
	}
}

// paintHorizontalSlider renders a horizontal slider.
func paintHorizontalSlider(canvas widget.Canvas, ps PaintState, hasScheme bool) {
	bounds := ps.Bounds
	trackY := bounds.Min.Y + bounds.Height()/2
	trackLeft := bounds.Min.X + thumbRadius
	trackRight := bounds.Max.X - thumbRadius
	trackWidth := trackRight - trackLeft

	if trackWidth <= 0 {
		return
	}

	thumbX := trackLeft + ps.Progress*trackWidth
	trackRect := geometry.NewRect(trackLeft, trackY-trackHeight/2, trackWidth, trackHeight)

	// Draw inactive track.
	inactiveColor := resolveInactiveTrack(ps, hasScheme)
	canvas.DrawRoundRect(trackRect, inactiveColor, trackHeight/2)

	// Draw active track.
	activeWidth := thumbX - trackLeft
	if activeWidth > 0 {
		activeColor := resolveActiveTrack(ps, hasScheme)
		activeRect := geometry.NewRect(trackLeft, trackY-trackHeight/2, activeWidth, trackHeight)
		canvas.DrawRoundRect(activeRect, activeColor, trackHeight/2)
	}

	// Draw marks.
	paintMarks(canvas, ps, hasScheme, trackLeft, trackWidth, trackY)

	// Draw thumb.
	thumbCenter := geometry.Pt(thumbX, trackY)
	paintThumb(canvas, ps, hasScheme, thumbCenter)
}

// paintVerticalSlider renders a vertical slider.
func paintVerticalSlider(canvas widget.Canvas, ps PaintState, hasScheme bool) {
	bounds := ps.Bounds
	trackX := bounds.Min.X + bounds.Width()/2
	trackTop := bounds.Min.Y + thumbRadius
	trackBottom := bounds.Max.Y - thumbRadius
	trackLen := trackBottom - trackTop

	if trackLen <= 0 {
		return
	}

	// In vertical orientation, progress 0 is at bottom, 1 is at top.
	thumbY := trackBottom - ps.Progress*trackLen
	trackRect := geometry.NewRect(trackX-trackHeight/2, trackTop, trackHeight, trackLen)

	// Draw inactive track.
	inactiveColor := resolveInactiveTrack(ps, hasScheme)
	canvas.DrawRoundRect(trackRect, inactiveColor, trackHeight/2)

	// Draw active track (from bottom to thumb).
	activeLen := trackBottom - thumbY
	if activeLen > 0 {
		activeColor := resolveActiveTrack(ps, hasScheme)
		activeRect := geometry.NewRect(trackX-trackHeight/2, thumbY, trackHeight, activeLen)
		canvas.DrawRoundRect(activeRect, activeColor, trackHeight/2)
	}

	// Draw thumb.
	thumbCenter := geometry.Pt(trackX, thumbY)
	paintThumb(canvas, ps, hasScheme, thumbCenter)
}

// paintThumb draws the slider thumb circle.
func paintThumb(canvas widget.Canvas, ps PaintState, hasScheme bool, center geometry.Point) {
	thumbColor := resolveThumbColor(ps, hasScheme)
	borderColor := resolveThumbBorder(ps, hasScheme)

	canvas.DrawCircle(center, thumbRadius, thumbColor)
	canvas.StrokeCircle(center, thumbRadius, borderColor, thumbBorderWidth)

	// Focus ring.
	if ps.Focused && !ps.Disabled {
		ringColor := resolveFocusRing(ps, hasScheme)
		canvas.StrokeCircle(center, thumbRadius+focusRingOffset, ringColor, focusRingStrokeWidth)
	}
}

// paintMarks draws tick marks on the track.
func paintMarks(canvas widget.Canvas, ps PaintState, hasScheme bool, trackLeft, trackWidth, trackY float32) {
	if len(ps.Marks) == 0 {
		return
	}

	rangeVal := ps.Max - ps.Min
	if rangeVal <= 0 {
		return
	}

	markColor := defaultMarkColor
	if hasScheme {
		markColor = ps.ColorScheme.MarkColor
	}

	for _, m := range ps.Marks {
		progress := (m.Value - ps.Min) / rangeVal
		if progress < 0 || progress > 1 {
			continue
		}
		markX := trackLeft + progress*trackWidth
		top := geometry.Pt(markX, trackY+trackHeight/2+markGap)
		bottom := geometry.Pt(markX, trackY+trackHeight/2+markGap+markLength)
		canvas.DrawLine(top, bottom, markColor, markStrokeWidth)
	}
}

// Color resolution helpers.

func resolveActiveTrack(ps PaintState, hasScheme bool) widget.Color {
	if ps.Disabled {
		if hasScheme {
			return ps.ColorScheme.DisabledTrack
		}
		return defaultDisabledTrack
	}
	if hasScheme {
		return ps.ColorScheme.ActiveTrack
	}
	return defaultActiveTrack
}

func resolveInactiveTrack(ps PaintState, hasScheme bool) widget.Color {
	if ps.Disabled {
		if hasScheme {
			return ps.ColorScheme.DisabledTrack
		}
		return defaultDisabledTrack
	}
	if hasScheme {
		return ps.ColorScheme.InactiveTrack
	}
	return defaultInactiveTrack
}

func resolveThumbColor(ps PaintState, hasScheme bool) widget.Color {
	if ps.Disabled {
		if hasScheme {
			return ps.ColorScheme.DisabledThumb
		}
		return defaultDisabledThumb
	}
	color := defaultThumbColor
	if hasScheme {
		color = ps.ColorScheme.Thumb
	}
	return applyStateModifier(color, ps.Hovered, ps.Dragging)
}

func resolveThumbBorder(ps PaintState, hasScheme bool) widget.Color {
	if ps.Disabled {
		if hasScheme {
			return ps.ColorScheme.DisabledTrack
		}
		return defaultDisabledTrack
	}
	if hasScheme {
		return ps.ColorScheme.ThumbBorder
	}
	return defaultThumbBorder
}

func resolveFocusRing(ps PaintState, hasScheme bool) widget.Color {
	if hasScheme {
		return ps.ColorScheme.FocusRing
	}
	return defaultFocusRingColor
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

// SliderThumbRadius returns the default thumb radius.
func (DefaultPainter) SliderThumbRadius() float32 { return thumbRadius }

// SliderTrackHeight returns the default track height.
func (DefaultPainter) SliderTrackHeight() float32 { return trackHeight }

// resolveSliderLayoutMetrics returns the LayoutMetrics from the painter if it
// implements that interface, otherwise returns DefaultPainter metrics.
func resolveSliderLayoutMetrics(p Painter) LayoutMetrics {
	if lm, ok := p.(LayoutMetrics); ok {
		return lm
	}
	return DefaultPainter{}
}

// Compile-time checks.
var _ LayoutMetrics = DefaultPainter{}

// Painting constants.
const (
	trackHeight      float32 = 4
	thumbRadius      float32 = 10
	thumbBorderWidth float32 = 2

	focusRingOffset      float32 = 2
	focusRingStrokeWidth float32 = 2
	hoverLightenFactor   float32 = 0.1
	pressedDarkenFactor  float32 = 0.15

	markGap         float32 = 2
	markLength      float32 = 6
	markStrokeWidth float32 = 1
)

// Default colors for DefaultPainter.
var (
	defaultActiveTrack    = widget.Hex(0x6750A4)
	defaultInactiveTrack  = widget.RGBA(0.78, 0.78, 0.78, 1.0)
	defaultThumbColor     = widget.ColorWhite
	defaultThumbBorder    = widget.Hex(0x6750A4)
	defaultDisabledTrack  = widget.RGBA(0.12, 0.12, 0.13, 0.12)
	defaultDisabledThumb  = widget.RGBA(0.92, 0.92, 0.92, 1.0)
	defaultFocusRingColor = widget.Hex(0x6750A4).WithAlpha(0.7)
	defaultMarkColor      = widget.RGBA(0.5, 0.5, 0.5, 1.0)
)
