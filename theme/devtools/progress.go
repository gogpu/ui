package devtools

import (
	"math"

	"github.com/gogpu/ui/core/progress"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// ProgressPainter renders progress indicators using DevTools design tokens.
// DevTools progress uses a 2px stroke: Gray3 track with Blue6 fill for both
// linear and circular indicators. The indeterminate mode uses an animated
// Blue6 arc, matching JetBrains IDE progress bar styling.
//
// If Theme is nil, ProgressPainter falls back to the default DevTools dark palette.
type ProgressPainter struct {
	Theme *Theme // nil uses default DevTools dark fallback
}

// PaintProgress renders a circular progress indicator according to DevTools specifications.
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

// paintDeterminate draws a track circle and a progress arc using DevTools colors.
func (p ProgressPainter) paintDeterminate(canvas widget.Canvas, ps progress.PaintState, center geometry.Point, radius float32) {
	trackColor, indicatorColor, labelColor := p.resolveProgressColors(ps)

	// Draw track circle (full 360 degrees).
	canvas.StrokeCircle(center, radius, trackColor, ps.StrokeWidth)

	// Draw progress arc (0 to value*360 degrees, starting from top).
	if ps.Value > 0 {
		startAngle := -math.Pi / 2
		sweepAngle := ps.Value * 2 * math.Pi
		dtDrawProgressArc(canvas, center, radius, startAngle, sweepAngle, indicatorColor, ps.StrokeWidth)
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
		canvas.DrawText(ps.Label, labelBounds, dtProgressFontSize, labelColor, false, dtProgressTextAlign)
	}
}

// paintIndeterminate draws a rotating partial arc using DevTools colors.
func (p ProgressPainter) paintIndeterminate(canvas widget.Canvas, ps progress.PaintState, center geometry.Point, radius float32) {
	trackColor, indicatorColor, _ := p.resolveProgressColors(ps)

	// Draw track circle.
	canvas.StrokeCircle(center, radius, trackColor, ps.StrokeWidth)

	// Draw rotating arc.
	startAngle := -math.Pi/2 + ps.Rotation
	dtDrawProgressArc(canvas, center, radius, startAngle, dtProgressIndeterminateSpan, indicatorColor, ps.StrokeWidth)
}

// resolveProgressColors returns track, indicator, and label colors for the current state.
func (p ProgressPainter) resolveProgressColors(ps progress.PaintState) (track, indicator, label widget.Color) {
	if ps.Disabled {
		return dtProgressDisabledTrack, dtProgressDisabledIndicator, dtProgressDisabledIndicator
	}

	// Use the color scheme from PaintState if provided.
	hasScheme := ps.ColorScheme != (progress.ProgressColorScheme{})
	if hasScheme {
		return ps.ColorScheme.Track, ps.ColorScheme.Indicator, ps.ColorScheme.Label
	}

	// Resolve from theme.
	if p.Theme != nil {
		cs := p.Theme.Colors
		return cs.Border, cs.Primary, cs.OnSurface
	}

	return dtProgressDefaultTrack, dtProgressDefaultIndicator, dtProgressDefaultLabel
}

// dtDrawProgressArc approximates a circular arc using line segments.
func dtDrawProgressArc(canvas widget.Canvas, center geometry.Point, radius float32, startAngle, sweepAngle float64, color widget.Color, strokeWidth float32) {
	if sweepAngle == 0 {
		return
	}

	segments := int(math.Ceil(float64(dtProgressArcSegments) * math.Abs(sweepAngle) / (2 * math.Pi)))
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

// Default DevTools colors for progress indicators.
var (
	dtProgressDefaultIndicator  = widget.Hex(0x3574F0)                // Blue6 (primary)
	dtProgressDefaultTrack      = widget.Hex(0x393B40)                // Gray3 (border)
	dtProgressDefaultLabel      = widget.Hex(0xDFE1E5)                // Gray12 (on-surface)
	dtProgressDisabledIndicator = widget.RGBA(0.44, 0.45, 0.48, 0.38) // Gray7 @ 38%
	dtProgressDisabledTrack     = widget.RGBA(0.22, 0.23, 0.25, 0.38) // Gray3 @ 38%
)

// DevTools progress drawing constants.
const (
	dtProgressFontSize          float32 = 11
	dtProgressTextAlign                 = widget.TextAlignCenter
	dtProgressArcSegments               = 64
	dtProgressIndeterminateSpan float64 = math.Pi * 0.75 // 270 degree arc for spinner
)

// Compile-time check that ProgressPainter implements Painter.
var _ progress.Painter = ProgressPainter{}
