package material3

import (
	"math"

	"github.com/gogpu/ui/core/progress"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// ProgressPainter renders circular progress indicators using Material 3 design tokens.
// It maps progress states (determinate, indeterminate, disabled) to the M3 color scheme
// with primary color for the arc and surface variant for the track.
//
// If Theme is nil, ProgressPainter falls back to the default M3 purple palette.
type ProgressPainter struct {
	Theme *Theme // nil uses default M3 purple fallback
}

// PaintProgress renders a circular progress indicator according to Material 3 specifications.
func (p ProgressPainter) PaintProgress(canvas widget.Canvas, ps progress.PaintState) {
	if ps.Bounds.IsEmpty() {
		return
	}

	bounds := ps.Bounds
	centerX := bounds.Min.X + bounds.Width()/2
	centerY := bounds.Min.Y + bounds.Height()/2
	center := geometry.Pt(centerX, centerY)

	// Use the smaller of width/height for the radius, minus stroke width.
	availDiameter := ps.Diameter
	if bounds.Width() < availDiameter {
		availDiameter = bounds.Width()
	}
	if bounds.Height() < availDiameter {
		availDiameter = bounds.Height()
	}
	radius := (availDiameter - ps.StrokeWidth) / 2
	if radius <= 0 {
		return
	}

	if ps.Indeterminate {
		p.paintIndeterminate(canvas, ps, center, radius)
	} else {
		p.paintDeterminate(canvas, ps, center, radius)
	}
}

// paintDeterminate draws a track circle and a progress arc using M3 colors.
func (p ProgressPainter) paintDeterminate(canvas widget.Canvas, ps progress.PaintState, center geometry.Point, radius float32) {
	trackColor, indicatorColor, labelColor := p.resolveProgressColors(ps)

	// Draw track circle (full 360 degrees).
	canvas.StrokeCircle(center, radius, trackColor, ps.StrokeWidth)

	// Draw progress arc (0 to value*360 degrees, starting from top).
	if ps.Value > 0 {
		startAngle := -math.Pi / 2
		sweepAngle := ps.Value * 2 * math.Pi
		m3DrawProgressArc(canvas, center, radius, startAngle, sweepAngle, indicatorColor, ps.StrokeWidth)
	}

	// Draw label centered if enabled.
	if ps.ShowLabel && ps.Label != "" {
		labelSize := ps.Diameter
		labelBounds := geometry.NewRect(
			center.X-labelSize/2,
			center.Y-labelSize/2,
			labelSize,
			labelSize,
		)
		canvas.DrawText(ps.Label, labelBounds, m3ProgressFontSize, labelColor, false, m3ProgressTextAlign)
	}
}

// paintIndeterminate draws a rotating partial arc using M3 colors.
func (p ProgressPainter) paintIndeterminate(canvas widget.Canvas, ps progress.PaintState, center geometry.Point, radius float32) {
	trackColor, indicatorColor, _ := p.resolveProgressColors(ps)

	// Draw track circle.
	canvas.StrokeCircle(center, radius, trackColor, ps.StrokeWidth)

	// Draw rotating arc.
	startAngle := -math.Pi/2 + ps.Rotation
	m3DrawProgressArc(canvas, center, radius, startAngle, m3ProgressIndeterminateSpan, indicatorColor, ps.StrokeWidth)
}

// resolveProgressColors returns track, indicator, and label colors for the current state.
func (p ProgressPainter) resolveProgressColors(ps progress.PaintState) (track, indicator, label widget.Color) {
	if ps.Disabled {
		return m3ProgressDisabledTrack, m3ProgressDisabledIndicator, m3ProgressDisabledIndicator
	}

	// Use the color scheme from PaintState if provided.
	hasScheme := ps.ColorScheme != (progress.ProgressColorScheme{})
	if hasScheme {
		return ps.ColorScheme.Track, ps.ColorScheme.Indicator, ps.ColorScheme.Label
	}

	// Resolve from theme.
	if p.Theme != nil {
		cs := p.Theme.Colors
		return cs.SurfaceVariant, cs.Primary, cs.OnSurface
	}

	return m3ProgressDefaultTrack, m3ProgressDefaultIndicator, m3ProgressDefaultLabel
}

// m3DrawProgressArc approximates a circular arc using line segments.
func m3DrawProgressArc(canvas widget.Canvas, center geometry.Point, radius float32, startAngle, sweepAngle float64, color widget.Color, strokeWidth float32) {
	if sweepAngle == 0 {
		return
	}

	segments := int(math.Ceil(float64(m3ProgressArcSegments) * math.Abs(sweepAngle) / (2 * math.Pi)))
	if segments < 2 {
		segments = 2
	}

	step := sweepAngle / float64(segments)
	prevX := center.X + radius*float32(math.Cos(startAngle))
	prevY := center.Y + radius*float32(math.Sin(startAngle))

	for i := 1; i <= segments; i++ {
		angle := startAngle + float64(i)*step
		curX := center.X + radius*float32(math.Cos(angle))
		curY := center.Y + radius*float32(math.Sin(angle))

		canvas.DrawLine(
			geometry.Pt(prevX, prevY),
			geometry.Pt(curX, curY),
			color,
			strokeWidth,
		)

		prevX = curX
		prevY = curY
	}
}

// Default M3 colors for circular progress.
var (
	m3ProgressDefaultIndicator  = widget.Hex(0x6750A4)                // M3 primary
	m3ProgressDefaultTrack      = widget.Hex(0xE7E0EC)                // M3 surface variant
	m3ProgressDefaultLabel      = widget.Hex(0x1C1B1F)                // M3 on-surface
	m3ProgressDisabledIndicator = widget.RGBA(0.12, 0.12, 0.13, 0.38) // M3 disabled fg
	m3ProgressDisabledTrack     = widget.RGBA(0.12, 0.12, 0.13, 0.12) // M3 disabled bg
)

// M3 circular progress drawing constants.
const (
	m3ProgressFontSize          float32 = 12
	m3ProgressTextAlign                 = widget.TextAlignCenter
	m3ProgressArcSegments               = 64
	m3ProgressIndeterminateSpan float64 = math.Pi * 0.75 // 270 degree arc for spinner
)

// Compile-time check that ProgressPainter implements Painter.
var _ progress.Painter = ProgressPainter{}
