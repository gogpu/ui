package fluent

import (
	"github.com/gogpu/ui/core/slider"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// SliderPainter renders sliders using Fluent Design tokens.
//
// If Theme is nil, SliderPainter falls back to the default Fluent Blue palette.
type SliderPainter struct {
	Theme *Theme // nil uses default Fluent Blue fallback
}

// resolveColors returns the SliderColorScheme derived from the painter's Theme.
func (p SliderPainter) resolveColors() slider.SliderColorScheme {
	if p.Theme == nil {
		return flDefaultSliderColors
	}
	cs := p.Theme.Colors
	return slider.SliderColorScheme{
		ActiveTrack:   cs.Accent,
		InactiveTrack: cs.FillDefault,
		Thumb:         cs.Accent,
		ThumbBorder:   cs.Surface,
		FocusRing:     cs.StrokeFocus,
		DisabledTrack: cs.FillDisable,
		DisabledThumb: cs.OnSurfaceSecond.WithAlpha(flDisabledAlpha),
		MarkColor:     cs.OnSurfaceSecond,
	}
}

// PaintSlider renders a slider according to Fluent Design specifications.
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
		flPaintVerticalSlider(canvas, ps, disabled, colors)
	} else {
		flPaintHorizontalSlider(canvas, ps, disabled, colors)
	}
}

// flPaintHorizontalSlider renders a horizontal Fluent slider.
func flPaintHorizontalSlider(canvas widget.Canvas, ps slider.PaintState, disabled bool, colors slider.SliderColorScheme) {
	bounds := ps.Bounds
	trackY := bounds.Min.Y + bounds.Height()/2
	trackLeft := bounds.Min.X + flSliderThumbRadius
	trackRight := bounds.Max.X - flSliderThumbRadius
	trackWidth := trackRight - trackLeft

	if trackWidth <= 0 {
		return
	}

	thumbX := trackLeft + ps.Progress*trackWidth

	// Inactive track.
	inactiveColor := flResolvedInactiveTrack(disabled, colors)
	inactiveRect := geometry.NewRect(trackLeft, trackY-flSliderTrackHeight/2, trackWidth, flSliderTrackHeight)
	canvas.DrawRoundRect(inactiveRect, inactiveColor, flSliderTrackHeight/2)

	// Active track.
	activeWidth := thumbX - trackLeft
	if activeWidth > 0 {
		activeColor := flResolvedActiveTrack(disabled, colors)
		activeRect := geometry.NewRect(trackLeft, trackY-flSliderTrackHeight/2, activeWidth, flSliderTrackHeight)
		canvas.DrawRoundRect(activeRect, activeColor, flSliderTrackHeight/2)
	}

	// Marks.
	flPaintSliderMarks(canvas, ps, colors, trackLeft, trackWidth, trackY)

	// Thumb.
	thumbCenter := geometry.Pt(thumbX, trackY)
	flPaintSliderThumb(canvas, ps, disabled, colors, thumbCenter)
}

// flPaintVerticalSlider renders a vertical Fluent slider.
func flPaintVerticalSlider(canvas widget.Canvas, ps slider.PaintState, disabled bool, colors slider.SliderColorScheme) {
	bounds := ps.Bounds
	trackX := bounds.Min.X + bounds.Width()/2
	trackTop := bounds.Min.Y + flSliderThumbRadius
	trackBottom := bounds.Max.Y - flSliderThumbRadius
	trackLen := trackBottom - trackTop

	if trackLen <= 0 {
		return
	}

	thumbY := trackBottom - ps.Progress*trackLen

	// Inactive track.
	inactiveColor := flResolvedInactiveTrack(disabled, colors)
	inactiveRect := geometry.NewRect(trackX-flSliderTrackHeight/2, trackTop, flSliderTrackHeight, trackLen)
	canvas.DrawRoundRect(inactiveRect, inactiveColor, flSliderTrackHeight/2)

	// Active track.
	activeLen := trackBottom - thumbY
	if activeLen > 0 {
		activeColor := flResolvedActiveTrack(disabled, colors)
		activeRect := geometry.NewRect(trackX-flSliderTrackHeight/2, thumbY, flSliderTrackHeight, activeLen)
		canvas.DrawRoundRect(activeRect, activeColor, flSliderTrackHeight/2)
	}

	// Thumb.
	thumbCenter := geometry.Pt(trackX, thumbY)
	flPaintSliderThumb(canvas, ps, disabled, colors, thumbCenter)
}

// flPaintSliderThumb draws the Fluent slider thumb.
// Fluent thumb has a white border (surface color) around the accent fill.
func flPaintSliderThumb(canvas widget.Canvas, ps slider.PaintState, disabled bool, colors slider.SliderColorScheme, center geometry.Point) {
	thumbColor := flResolvedThumbColor(ps, disabled, colors)
	// Outer white circle (border) then inner accent circle.
	canvas.DrawCircle(center, flSliderThumbRadius, colors.ThumbBorder)
	canvas.DrawCircle(center, flSliderThumbRadius-flSliderThumbBorderWidth, thumbColor)

	// Focus ring.
	if ps.Focused && !disabled {
		canvas.StrokeCircle(center, flSliderThumbRadius+flSliderFocusRingOffset, colors.FocusRing, flSliderFocusRingStrokeWidth)
	}
}

// flPaintSliderMarks draws tick marks on the Fluent slider track.
func flPaintSliderMarks(canvas widget.Canvas, ps slider.PaintState, colors slider.SliderColorScheme, trackLeft, trackWidth, trackY float32) {
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
		canvas.DrawCircle(geometry.Pt(markX, trackY), flSliderMarkRadius, colors.MarkColor)
	}
}

// Color resolution helpers.

func flResolvedActiveTrack(disabled bool, colors slider.SliderColorScheme) widget.Color {
	if disabled {
		return colors.DisabledTrack
	}
	return colors.ActiveTrack
}

func flResolvedInactiveTrack(disabled bool, colors slider.SliderColorScheme) widget.Color {
	if disabled {
		return colors.DisabledTrack
	}
	return colors.InactiveTrack
}

func flResolvedThumbColor(ps slider.PaintState, disabled bool, colors slider.SliderColorScheme) widget.Color {
	if disabled {
		return colors.DisabledThumb
	}
	base := colors.Thumb
	return flApplyState(base, ps.Hovered, ps.Dragging)
}

// flDefaultSliderColors holds the default Fluent slider color scheme.
var flDefaultSliderColors = slider.SliderColorScheme{
	ActiveTrack:   DefaultAccentColor,
	InactiveTrack: widget.RGBA(0, 0, 0, 0.04),
	Thumb:         DefaultAccentColor,
	ThumbBorder:   widget.ColorWhite,
	FocusRing:     DefaultAccentColor,
	DisabledTrack: widget.RGBA(0, 0, 0, 0.04),
	DisabledThumb: widget.RGBA(0.38, 0.38, 0.38, 0.38),
	MarkColor:     widget.Hex(0x616161),
}

// Fluent slider drawing constants.
const (
	flSliderTrackHeight          float32 = 4
	flSliderThumbRadius          float32 = 10
	flSliderThumbBorderWidth     float32 = 2
	flSliderFocusRingOffset      float32 = 2
	flSliderFocusRingStrokeWidth float32 = 2
	flSliderMarkRadius           float32 = 2
)

// Compile-time check that SliderPainter implements Painter.
var _ slider.Painter = SliderPainter{}
