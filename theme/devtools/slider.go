package devtools

import (
	"github.com/gogpu/ui/core/slider"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// SliderPainter renders sliders using DevTools design tokens.
// DevTools sliders have a 2px track and 12px thumb circle with Blue6 accent,
// matching JetBrains IDE preference panel slider styling.
//
// If Theme is nil, SliderPainter falls back to the default DevTools dark palette.
type SliderPainter struct {
	Theme *Theme // nil uses default DevTools dark fallback
}

// resolveColors returns the SliderColorScheme derived from the painter's Theme.
func (p SliderPainter) resolveColors() slider.SliderColorScheme {
	if p.Theme == nil {
		return dtDefaultSliderColors
	}
	cs := p.Theme.Colors
	return slider.SliderColorScheme{
		ActiveTrack:   cs.Primary,
		InactiveTrack: cs.ControlFill,
		Thumb:         cs.Primary,
		ThumbBorder:   cs.Surface,
		FocusRing:     cs.BorderFocus,
		DisabledTrack: cs.ControlFill,
		DisabledThumb: cs.OnSurfaceDisabled,
		MarkColor:     cs.OnSurfaceSecondary,
	}
}

// PaintSlider renders a slider according to DevTools design specifications.
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
		dtPaintVerticalSlider(canvas, ps, disabled, colors)
	} else {
		dtPaintHorizontalSlider(canvas, ps, disabled, colors)
	}
}

// dtPaintHorizontalSlider renders a horizontal DevTools slider.
func dtPaintHorizontalSlider(canvas widget.Canvas, ps slider.PaintState, disabled bool, colors slider.SliderColorScheme) {
	bounds := ps.Bounds
	trackY := bounds.Min.Y + bounds.Height()/2
	trackLeft := bounds.Min.X + dtSliderThumbRadius
	trackRight := bounds.Max.X - dtSliderThumbRadius
	trackWidth := trackRight - trackLeft

	if trackWidth <= 0 {
		return
	}

	thumbX := trackLeft + ps.Progress*trackWidth

	// Inactive track.
	inactiveColor := dtResolvedInactiveTrack(disabled, colors)
	inactiveRect := geometry.NewRect(trackLeft, trackY-dtSliderTrackHeight/2, trackWidth, dtSliderTrackHeight)
	canvas.DrawRoundRect(inactiveRect, inactiveColor, dtSliderTrackHeight/2)

	// Active track.
	activeWidth := thumbX - trackLeft
	if activeWidth > 0 {
		activeColor := dtResolvedActiveTrack(disabled, colors)
		activeRect := geometry.NewRect(trackLeft, trackY-dtSliderTrackHeight/2, activeWidth, dtSliderTrackHeight)
		canvas.DrawRoundRect(activeRect, activeColor, dtSliderTrackHeight/2)
	}

	// Marks.
	dtPaintSliderMarks(canvas, ps, colors, trackLeft, trackWidth, trackY)

	// Thumb.
	thumbCenter := geometry.Pt(thumbX, trackY)
	dtPaintSliderThumb(canvas, ps, disabled, colors, thumbCenter)
}

// dtPaintVerticalSlider renders a vertical DevTools slider.
func dtPaintVerticalSlider(canvas widget.Canvas, ps slider.PaintState, disabled bool, colors slider.SliderColorScheme) {
	bounds := ps.Bounds
	trackX := bounds.Min.X + bounds.Width()/2
	trackTop := bounds.Min.Y + dtSliderThumbRadius
	trackBottom := bounds.Max.Y - dtSliderThumbRadius
	trackLen := trackBottom - trackTop

	if trackLen <= 0 {
		return
	}

	thumbY := trackBottom - ps.Progress*trackLen

	// Inactive track.
	inactiveColor := dtResolvedInactiveTrack(disabled, colors)
	inactiveRect := geometry.NewRect(trackX-dtSliderTrackHeight/2, trackTop, dtSliderTrackHeight, trackLen)
	canvas.DrawRoundRect(inactiveRect, inactiveColor, dtSliderTrackHeight/2)

	// Active track.
	activeLen := trackBottom - thumbY
	if activeLen > 0 {
		activeColor := dtResolvedActiveTrack(disabled, colors)
		activeRect := geometry.NewRect(trackX-dtSliderTrackHeight/2, thumbY, dtSliderTrackHeight, activeLen)
		canvas.DrawRoundRect(activeRect, activeColor, dtSliderTrackHeight/2)
	}

	// Thumb.
	thumbCenter := geometry.Pt(trackX, thumbY)
	dtPaintSliderThumb(canvas, ps, disabled, colors, thumbCenter)
}

// dtPaintSliderThumb draws the DevTools slider thumb.
// DevTools thumb: accent-colored filled circle.
func dtPaintSliderThumb(canvas widget.Canvas, ps slider.PaintState, disabled bool, colors slider.SliderColorScheme, center geometry.Point) {
	thumbColor := dtResolvedThumbColor(ps, disabled, colors)
	canvas.DrawCircle(center, dtSliderThumbRadius, thumbColor)

	// Focus ring.
	if ps.Focused && !disabled {
		canvas.StrokeCircle(center, dtSliderThumbRadius+dtSliderFocusRingOffset, colors.FocusRing, dtSliderFocusRingStrokeWidth)
	}
}

// dtPaintSliderMarks draws tick marks on the DevTools slider track.
func dtPaintSliderMarks(canvas widget.Canvas, ps slider.PaintState, colors slider.SliderColorScheme, trackLeft, trackWidth, trackY float32) {
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
		canvas.DrawCircle(geometry.Pt(markX, trackY), dtSliderMarkRadius, colors.MarkColor)
	}
}

// Color resolution helpers.

func dtResolvedActiveTrack(disabled bool, colors slider.SliderColorScheme) widget.Color {
	if disabled {
		return colors.DisabledTrack
	}
	return colors.ActiveTrack
}

func dtResolvedInactiveTrack(disabled bool, colors slider.SliderColorScheme) widget.Color {
	if disabled {
		return colors.DisabledTrack
	}
	return colors.InactiveTrack
}

func dtResolvedThumbColor(ps slider.PaintState, disabled bool, colors slider.SliderColorScheme) widget.Color {
	if disabled {
		return colors.DisabledThumb
	}
	base := colors.Thumb
	return dtApplyState(base, ps.Hovered, ps.Dragging)
}

// dtDefaultSliderColors holds the default DevTools dark slider color scheme.
var dtDefaultSliderColors = slider.SliderColorScheme{
	ActiveTrack:   DefaultAccentColor,
	InactiveTrack: widget.Hex(0x393B40), // Gray3
	Thumb:         DefaultAccentColor,
	ThumbBorder:   widget.Hex(0x2B2D30), // Gray2 (Surface)
	FocusRing:     DefaultAccentColor,
	DisabledTrack: widget.Hex(0x393B40), // Gray3
	DisabledThumb: widget.Hex(0x6F737A), // Gray7
	MarkColor:     widget.Hex(0x9DA0A8), // Gray9
}

// DevTools slider drawing constants.
const (
	dtSliderTrackHeight          float32 = 2
	dtSliderThumbRadius          float32 = 6
	dtSliderFocusRingOffset      float32 = 2
	dtSliderFocusRingStrokeWidth float32 = 1
	dtSliderMarkRadius           float32 = 2
)

// Compile-time check that SliderPainter implements Painter.
var _ slider.Painter = SliderPainter{}
