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

	// Use pre-computed geometry when available (ADR-034 Phase 4).
	center := ps.Center
	radius := ps.Radius
	if radius <= 0 {
		// Legacy fallback: compute from bounds (for direct PaintState construction).
		center, radius = progress.ComputeCenterRadius(ps)
		if radius <= 0 {
			return
		}
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
		strokeArcWithCap(canvas, center, radius, startAngle, sweepAngle, indicatorColor, ps.StrokeWidth, widget.LineCapRound)
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

// paintIndeterminate draws a variable-length rotating arc per M3 spec.
// The arc grows from ~0 to ~270 degrees then shrinks back on a 1.333s cycle
// while continuously rotating (Flutter progress_indicator.dart reference).
func (p ProgressPainter) paintIndeterminate(canvas widget.Canvas, ps progress.PaintState, center geometry.Point, radius float32) {
	trackColor, indicatorColor, _ := p.resolveProgressColors(ps)

	// Draw track circle.
	canvas.StrokeCircle(center, radius, trackColor, ps.StrokeWidth)

	// Use pre-computed arc angles (ADR-034 Phase 4).
	arcStart := ps.ArcStartAngle
	arcSweep := ps.ArcSweepAngle
	if arcSweep == 0 {
		// Legacy fallback: compute from AnimationPhase.
		arcStart, arcSweep = progress.ComputeArcAngles(ps.AnimationPhase, ps.Rotation)
	}

	strokeArcWithCap(canvas, center, radius, arcStart, arcSweep, indicatorColor, ps.StrokeWidth, widget.LineCapRound)
}

// strokeArcWithCap draws an arc using StrokeArcStyled if available, falling back to StrokeArc.
func strokeArcWithCap(canvas widget.Canvas, center geometry.Point, radius float32,
	startAngle, sweepAngle float64, color widget.Color, strokeWidth float32, lineCap widget.LineCap) {
	if s, ok := canvas.(widget.ArcStroker); ok {
		s.StrokeArcStyled(center, radius, startAngle, sweepAngle, color, strokeWidth, lineCap)
		return
	}
	canvas.StrokeArc(center, radius, startAngle, sweepAngle, color, strokeWidth)
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
	m3ProgressFontSize  float32 = 12
	m3ProgressTextAlign         = widget.TextAlignCenter
)

// Compile-time check that ProgressPainter implements Painter.
var _ progress.Painter = ProgressPainter{}
