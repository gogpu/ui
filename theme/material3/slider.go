package material3

import (
	"github.com/gogpu/ui/core/slider"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// SliderPainter renders sliders using Material 3 design tokens.
// It maps slider states (normal, hover, dragging, disabled) to
// the M3 color scheme and applies appropriate interaction feedback.
//
// If Theme is nil, SliderPainter falls back to the default M3 purple palette.
type SliderPainter struct {
	Theme *Theme // nil uses default M3 purple fallback
}

// resolveColors returns the SliderColorScheme derived from the painter's Theme.
// If Theme is nil, it returns the default M3 purple color scheme.
func (p SliderPainter) resolveColors() slider.SliderColorScheme {
	if p.Theme == nil {
		return m3DefaultSliderColors
	}
	cs := p.Theme.Colors
	return slider.SliderColorScheme{
		ActiveTrack:   cs.Primary,
		InactiveTrack: cs.SurfaceVariant,
		Thumb:         cs.Primary,
		ThumbBorder:   cs.Primary,
		FocusRing:     cs.Primary.WithAlpha(0.7),
		DisabledTrack: cs.OnSurface.WithAlpha(0.12),
		DisabledThumb: cs.OnSurface.WithAlpha(0.38),
		MarkColor:     cs.OnSurfaceVariant,
	}
}

// PaintSlider renders a slider according to Material 3 specifications.
func (p SliderPainter) PaintSlider(canvas widget.Canvas, ps slider.PaintState) {
	if ps.Bounds.IsEmpty() {
		return
	}

	// Determine the color scheme to use.
	colors := ps.ColorScheme
	if colors == (slider.SliderColorScheme{}) {
		colors = p.resolveColors()
	}

	disabled := ps.Disabled

	if ps.Orientation == slider.Vertical {
		m3PaintVerticalSlider(canvas, ps, disabled, colors)
	} else {
		m3PaintHorizontalSlider(canvas, ps, disabled, colors)
	}
}

// m3PaintHorizontalSlider renders a horizontal M3 slider.
func m3PaintHorizontalSlider(canvas widget.Canvas, ps slider.PaintState, disabled bool, colors slider.SliderColorScheme) {
	bounds := ps.Bounds
	trackY := bounds.Min.Y + bounds.Height()/2
	trackLeft := bounds.Min.X + m3SliderThumbRadius
	trackRight := bounds.Max.X - m3SliderThumbRadius
	trackWidth := trackRight - trackLeft

	if trackWidth <= 0 {
		return
	}

	thumbX := trackLeft + ps.Progress*trackWidth

	// Draw inactive track.
	inactiveColor := m3ResolvedInactiveTrack(disabled, colors)
	inactiveRect := geometry.NewRect(trackLeft, trackY-m3SliderTrackHeight/2, trackWidth, m3SliderTrackHeight)
	canvas.DrawRoundRect(inactiveRect, inactiveColor, m3SliderTrackHeight/2)

	// Draw active track.
	activeWidth := thumbX - trackLeft
	if activeWidth > 0 {
		activeColor := m3ResolvedActiveTrack(disabled, colors)
		activeRect := geometry.NewRect(trackLeft, trackY-m3SliderTrackHeight/2, activeWidth, m3SliderTrackHeight)
		canvas.DrawRoundRect(activeRect, activeColor, m3SliderTrackHeight/2)
	}

	// Draw marks.
	m3PaintMarks(canvas, ps, colors, trackLeft, trackWidth, trackY)

	// Draw thumb.
	thumbCenter := geometry.Pt(thumbX, trackY)
	m3PaintThumb(canvas, ps, disabled, colors, thumbCenter)
}

// m3PaintVerticalSlider renders a vertical M3 slider.
func m3PaintVerticalSlider(canvas widget.Canvas, ps slider.PaintState, disabled bool, colors slider.SliderColorScheme) {
	bounds := ps.Bounds
	trackX := bounds.Min.X + bounds.Width()/2
	trackTop := bounds.Min.Y + m3SliderThumbRadius
	trackBottom := bounds.Max.Y - m3SliderThumbRadius
	trackLen := trackBottom - trackTop

	if trackLen <= 0 {
		return
	}

	thumbY := trackBottom - ps.Progress*trackLen

	// Draw inactive track.
	inactiveColor := m3ResolvedInactiveTrack(disabled, colors)
	inactiveRect := geometry.NewRect(trackX-m3SliderTrackHeight/2, trackTop, m3SliderTrackHeight, trackLen)
	canvas.DrawRoundRect(inactiveRect, inactiveColor, m3SliderTrackHeight/2)

	// Draw active track.
	activeLen := trackBottom - thumbY
	if activeLen > 0 {
		activeColor := m3ResolvedActiveTrack(disabled, colors)
		activeRect := geometry.NewRect(trackX-m3SliderTrackHeight/2, thumbY, m3SliderTrackHeight, activeLen)
		canvas.DrawRoundRect(activeRect, activeColor, m3SliderTrackHeight/2)
	}

	// Draw thumb.
	thumbCenter := geometry.Pt(trackX, thumbY)
	m3PaintThumb(canvas, ps, disabled, colors, thumbCenter)
}

// m3PaintThumb draws the M3 slider thumb.
func m3PaintThumb(canvas widget.Canvas, ps slider.PaintState, disabled bool, colors slider.SliderColorScheme, center geometry.Point) {
	thumbColor := m3ResolvedThumbColor(ps, disabled, colors)
	canvas.DrawCircle(center, m3SliderThumbRadius, thumbColor)

	// Focus ring.
	if ps.Focused && !disabled {
		canvas.StrokeCircle(center, m3SliderThumbRadius+m3SliderFocusRingOffset, colors.FocusRing, m3SliderFocusRingStrokeWidth)
	}
}

// m3PaintMarks draws tick marks on the M3 slider track.
func m3PaintMarks(canvas widget.Canvas, ps slider.PaintState, colors slider.SliderColorScheme, trackLeft, trackWidth, trackY float32) {
	if len(ps.Marks) == 0 {
		return
	}

	rangeVal := ps.Max - ps.Min
	if rangeVal <= 0 {
		return
	}

	markColor := colors.MarkColor

	for _, m := range ps.Marks {
		progress := (m.Value - ps.Min) / rangeVal
		if progress < 0 || progress > 1 {
			continue
		}
		markX := trackLeft + progress*trackWidth
		// Draw a small dot on the track.
		canvas.DrawCircle(geometry.Pt(markX, trackY), m3SliderMarkRadius, markColor)
	}
}

// Color resolution helpers.

func m3ResolvedActiveTrack(disabled bool, colors slider.SliderColorScheme) widget.Color {
	if disabled {
		return colors.DisabledTrack
	}
	return colors.ActiveTrack
}

func m3ResolvedInactiveTrack(disabled bool, colors slider.SliderColorScheme) widget.Color {
	if disabled {
		return colors.DisabledTrack
	}
	return colors.InactiveTrack
}

func m3ResolvedThumbColor(ps slider.PaintState, disabled bool, colors slider.SliderColorScheme) widget.Color {
	if disabled {
		return colors.DisabledThumb
	}
	base := colors.Thumb
	return m3ApplySliderState(base, ps.Hovered, ps.Dragging)
}

// m3ApplySliderState adjusts a color based on interaction state.
func m3ApplySliderState(base widget.Color, hovered, dragging bool) widget.Color {
	if dragging {
		return base.Lerp(widget.ColorBlack, m3SliderPressedDarkenFactor)
	}
	if hovered {
		return base.Lerp(widget.ColorWhite, m3SliderHoverLightenFactor)
	}
	return base
}

// m3DefaultSliderColors holds the default M3 purple color scheme for sliders.
// Used as a fallback when no Theme is provided.
var m3DefaultSliderColors = slider.SliderColorScheme{
	ActiveTrack:   widget.Hex(0x6750A4),                // M3 primary
	InactiveTrack: widget.Hex(0xE7E0EC),                // M3 surface variant
	Thumb:         widget.Hex(0x6750A4),                // M3 primary
	ThumbBorder:   widget.Hex(0x6750A4),                // M3 primary
	FocusRing:     widget.Hex(0x6750A4).WithAlpha(0.7), // M3 focus ring
	DisabledTrack: widget.RGBA(0.12, 0.12, 0.13, 0.12), // M3 disabled
	DisabledThumb: widget.RGBA(0.12, 0.12, 0.13, 0.38), // M3 disabled fg
	MarkColor:     widget.Hex(0x49454F),                // M3 on surface variant
}

// M3 slider drawing constants.
const (
	m3SliderTrackHeight          float32 = 4
	m3SliderThumbRadius          float32 = 10
	m3SliderFocusRingOffset      float32 = 2
	m3SliderFocusRingStrokeWidth float32 = 2
	m3SliderMarkRadius           float32 = 2
	m3SliderHoverLightenFactor   float32 = 0.1
	m3SliderPressedDarkenFactor  float32 = 0.15
)

// SliderThumbRadius returns the M3 thumb radius.
func (SliderPainter) SliderThumbRadius() float32 { return m3SliderThumbRadius }

// SliderTrackHeight returns the M3 track height.
func (SliderPainter) SliderTrackHeight() float32 { return m3SliderTrackHeight }

// Compile-time checks.
var (
	_ slider.Painter       = SliderPainter{}
	_ slider.LayoutMetrics = SliderPainter{}
)
