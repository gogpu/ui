package cupertino

import (
	"github.com/gogpu/ui/core/slider"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// SliderPainter renders sliders using Apple HIG design tokens.
// Cupertino sliders use a thin track with accent-colored active fill
// and a white thumb with a subtle shadow, following iOS slider conventions.
//
// If Theme is nil, SliderPainter falls back to the default system blue palette.
type SliderPainter struct {
	Theme *Theme // nil uses default system blue fallback
}

// resolveColors returns the SliderColorScheme derived from the painter's Theme.
func (p SliderPainter) resolveColors() slider.SliderColorScheme {
	if p.Theme == nil {
		return cupDefaultSliderColors
	}
	cs := p.Theme.Colors
	return slider.SliderColorScheme{
		ActiveTrack:   cs.Accent,
		InactiveTrack: cs.SystemFill,
		Thumb:         widget.ColorWhite,
		ThumbBorder:   cs.Separator,
		FocusRing:     cs.Accent.WithAlpha(cupSliderFocusAlpha),
		DisabledTrack: cs.QuaternaryLabel,
		DisabledThumb: cs.TertiaryLabel,
		MarkColor:     cs.SecondaryLabel,
	}
}

// PaintSlider renders a slider according to Apple HIG specifications.
func (p SliderPainter) PaintSlider(canvas widget.Canvas, ps slider.PaintState) {
	if ps.Bounds.IsEmpty() {
		return
	}

	colors := ps.ColorScheme
	if colors == (slider.SliderColorScheme{}) {
		colors = p.resolveColors()
	}

	disabled := ps.Disabled

	if ps.Orientation == slider.Vertical {
		cupPaintVerticalSlider(canvas, ps, disabled, colors)
	} else {
		cupPaintHorizontalSlider(canvas, ps, disabled, colors)
	}
}

// cupPaintHorizontalSlider renders a horizontal Cupertino slider.
func cupPaintHorizontalSlider(canvas widget.Canvas, ps slider.PaintState, disabled bool, colors slider.SliderColorScheme) {
	bounds := ps.Bounds
	trackY := bounds.Min.Y + bounds.Height()/2
	trackLeft := bounds.Min.X + cupSliderThumbRadius
	trackRight := bounds.Max.X - cupSliderThumbRadius
	trackWidth := trackRight - trackLeft

	if trackWidth <= 0 {
		return
	}

	thumbX := trackLeft + ps.Progress*trackWidth

	// Inactive track.
	inactiveColor := cupResolvedInactiveTrack(disabled, colors)
	inactiveRect := geometry.NewRect(trackLeft, trackY-cupSliderTrackHeight/2, trackWidth, cupSliderTrackHeight)
	canvas.DrawRoundRect(inactiveRect, inactiveColor, cupSliderTrackHeight/2)

	// Active track.
	activeWidth := thumbX - trackLeft
	if activeWidth > 0 {
		activeColor := cupResolvedActiveTrack(disabled, colors)
		activeRect := geometry.NewRect(trackLeft, trackY-cupSliderTrackHeight/2, activeWidth, cupSliderTrackHeight)
		canvas.DrawRoundRect(activeRect, activeColor, cupSliderTrackHeight/2)
	}

	// Marks.
	cupPaintSliderMarks(canvas, ps, colors, trackLeft, trackWidth, trackY)

	// Thumb.
	thumbCenter := geometry.Pt(thumbX, trackY)
	cupPaintSliderThumb(canvas, ps, disabled, colors, thumbCenter)
}

// cupPaintVerticalSlider renders a vertical Cupertino slider.
func cupPaintVerticalSlider(canvas widget.Canvas, ps slider.PaintState, disabled bool, colors slider.SliderColorScheme) {
	bounds := ps.Bounds
	trackX := bounds.Min.X + bounds.Width()/2
	trackTop := bounds.Min.Y + cupSliderThumbRadius
	trackBottom := bounds.Max.Y - cupSliderThumbRadius
	trackLen := trackBottom - trackTop

	if trackLen <= 0 {
		return
	}

	thumbY := trackBottom - ps.Progress*trackLen

	// Inactive track.
	inactiveColor := cupResolvedInactiveTrack(disabled, colors)
	inactiveRect := geometry.NewRect(trackX-cupSliderTrackHeight/2, trackTop, cupSliderTrackHeight, trackLen)
	canvas.DrawRoundRect(inactiveRect, inactiveColor, cupSliderTrackHeight/2)

	// Active track.
	activeLen := trackBottom - thumbY
	if activeLen > 0 {
		activeColor := cupResolvedActiveTrack(disabled, colors)
		activeRect := geometry.NewRect(trackX-cupSliderTrackHeight/2, thumbY, cupSliderTrackHeight, activeLen)
		canvas.DrawRoundRect(activeRect, activeColor, cupSliderTrackHeight/2)
	}

	// Thumb.
	thumbCenter := geometry.Pt(trackX, thumbY)
	cupPaintSliderThumb(canvas, ps, disabled, colors, thumbCenter)
}

// cupPaintSliderThumb draws the iOS-style white thumb with border.
func cupPaintSliderThumb(canvas widget.Canvas, ps slider.PaintState, disabled bool, colors slider.SliderColorScheme, center geometry.Point) {
	thumbColor := colors.Thumb
	if disabled {
		thumbColor = colors.DisabledThumb
	} else {
		thumbColor = cupApplySliderState(thumbColor, ps.Hovered, ps.Dragging)
	}

	// Shadow border.
	borderColor := colors.ThumbBorder
	if disabled {
		borderColor = colors.DisabledTrack
	}

	canvas.DrawCircle(center, cupSliderThumbRadius, thumbColor)
	canvas.StrokeCircle(center, cupSliderThumbRadius, borderColor, cupSliderThumbBorderWidth)

	// Focus ring.
	if ps.Focused && !disabled {
		canvas.StrokeCircle(center, cupSliderThumbRadius+cupSliderFocusRingOffset, colors.FocusRing, cupSliderFocusRingStroke)
	}
}

// cupPaintSliderMarks draws tick marks.
func cupPaintSliderMarks(canvas widget.Canvas, ps slider.PaintState, colors slider.SliderColorScheme, trackLeft, trackWidth, trackY float32) {
	if len(ps.Marks) == 0 {
		return
	}

	rangeVal := ps.Max - ps.Min
	if rangeVal <= 0 {
		return
	}

	for _, m := range ps.Marks {
		progress := (m.Value - ps.Min) / rangeVal
		if progress < 0 || progress > 1 {
			continue
		}
		markX := trackLeft + progress*trackWidth
		canvas.DrawCircle(geometry.Pt(markX, trackY), cupSliderMarkRadius, colors.MarkColor)
	}
}

// cupResolvedActiveTrack returns the active track color.
func cupResolvedActiveTrack(disabled bool, colors slider.SliderColorScheme) widget.Color {
	if disabled {
		return colors.DisabledTrack
	}
	return colors.ActiveTrack
}

// cupResolvedInactiveTrack returns the inactive track color.
func cupResolvedInactiveTrack(disabled bool, colors slider.SliderColorScheme) widget.Color {
	if disabled {
		return colors.DisabledTrack
	}
	return colors.InactiveTrack
}

// cupApplySliderState adjusts a color based on interaction state.
func cupApplySliderState(base widget.Color, hovered, dragging bool) widget.Color {
	if dragging {
		return base.Lerp(widget.ColorBlack, cupSliderPressedDarken)
	}
	if hovered {
		return base.Lerp(widget.ColorWhite, cupSliderHoverLighten)
	}
	return base
}

// cupDefaultSliderColors holds the default system blue color scheme for sliders.
var cupDefaultSliderColors = slider.SliderColorScheme{
	ActiveTrack:   systemBlue,
	InactiveTrack: widget.RGBA(0.47, 0.47, 0.50, 0.2),
	Thumb:         widget.ColorWhite,
	ThumbBorder:   widget.RGBA(0.0, 0.0, 0.0, 0.1),
	FocusRing:     systemBlue.WithAlpha(0.6),
	DisabledTrack: widget.RGBA(0.235, 0.235, 0.263, 0.18),
	DisabledThumb: widget.RGBA(0.235, 0.235, 0.263, 0.3),
	MarkColor:     widget.RGBA(0.235, 0.235, 0.263, 0.6),
}

// Cupertino slider drawing constants.
const (
	cupSliderTrackHeight      float32 = 3
	cupSliderThumbRadius      float32 = 14
	cupSliderThumbBorderWidth float32 = 0.5
	cupSliderFocusRingOffset  float32 = 3
	cupSliderFocusRingStroke  float32 = 2.5
	cupSliderMarkRadius       float32 = 2
	cupSliderHoverLighten     float32 = 0.08
	cupSliderPressedDarken    float32 = 0.12
	cupSliderFocusAlpha       float32 = 0.6
)

// SliderThumbRadius returns the Cupertino thumb radius.
func (SliderPainter) SliderThumbRadius() float32 { return cupSliderThumbRadius }

// SliderTrackHeight returns the Cupertino track height.
func (SliderPainter) SliderTrackHeight() float32 { return cupSliderTrackHeight }

// Compile-time checks.
var (
	_ slider.Painter       = SliderPainter{}
	_ slider.LayoutMetrics = SliderPainter{}
)
