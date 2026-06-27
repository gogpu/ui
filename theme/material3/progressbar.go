package material3

import (
	"github.com/gogpu/ui/core/progressbar"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// ProgressBarPainter renders progress bars using Material 3 design tokens.
// It maps progress bar states (normal, disabled) to the M3 color scheme
// with primary color for the filled bar and surface variant for the track.
//
// If Theme is nil, ProgressBarPainter falls back to the default M3 purple palette.
type ProgressBarPainter struct {
	Theme *Theme // nil uses default M3 purple fallback
}

// resolveColors returns the ProgressBarColorScheme derived from the painter's Theme.
// If Theme is nil, it returns the default M3 purple color scheme.
func (p ProgressBarPainter) resolveColors() progressbar.ProgressBarColorScheme {
	if p.Theme == nil {
		return m3DefaultProgressBarColors
	}
	cs := p.Theme.Colors
	return progressbar.ProgressBarColorScheme{
		Bar:           cs.Primary,
		Track:         cs.SurfaceVariant,
		Label:         cs.OnPrimary,
		DisabledBar:   cs.OnSurface.WithAlpha(0.38),
		DisabledTrack: cs.OnSurface.WithAlpha(0.12),
	}
}

// PaintProgressBar renders a progress bar according to Material 3 specifications.
func (p ProgressBarPainter) PaintProgressBar(canvas widget.Canvas, ps progressbar.PaintState) {
	if ps.Bounds.IsEmpty() {
		return
	}

	// Determine the color scheme to use.
	colors := ps.ColorScheme
	if colors == (progressbar.ProgressBarColorScheme{}) {
		colors = p.resolveColors()
	}

	radius := ps.Radius
	if radius <= 0 {
		radius = m3ProgressBarRadius
	}

	bounds := ps.Bounds

	// Calculate bar rect centered vertically in bounds.
	barY := bounds.Min.Y + (bounds.Height()-ps.BarHeight)/2
	barRect := geometry.NewRect(bounds.Min.X, barY, bounds.Width(), ps.BarHeight)

	// Draw track.
	trackColor := colors.Track
	if ps.Disabled {
		trackColor = colors.DisabledTrack
	}
	canvas.DrawRoundRect(barRect, trackColor, radius)

	// Draw filled portion.
	fillWidth := barRect.Width() * float32(ps.Value)
	if fillWidth > 0 {
		barColor := colors.Bar
		if ps.Disabled {
			barColor = colors.DisabledBar
		}
		fillRect := geometry.NewRect(barRect.Min.X, barRect.Min.Y, fillWidth, ps.BarHeight)

		// Clip to the track's rounded rect so the fill never exceeds the track shape.
		canvas.PushClipRoundRect(barRect, radius)
		canvas.DrawRoundRect(fillRect, barColor, radius)
		canvas.PopClip()
	}

	// Draw label if enabled.
	if ps.ShowLabel && ps.Label != "" {
		canvas.DrawText(ps.Label, barRect, m3ProgressBarFontSize, colors.Label, false, m3ProgressBarTextAlign)
	}
}

// m3DefaultProgressBarColors holds the default M3 purple color scheme for progress bars.
// Used as a fallback when no Theme is provided.
var m3DefaultProgressBarColors = progressbar.ProgressBarColorScheme{
	Bar:           widget.Hex(0x6750A4),                // M3 primary
	Track:         widget.Hex(0xE7E0EC),                // M3 surface variant
	Label:         widget.ColorWhite,                   // M3 on-primary
	DisabledBar:   widget.RGBA(0.12, 0.12, 0.13, 0.38), // M3 disabled fg
	DisabledTrack: widget.RGBA(0.12, 0.12, 0.13, 0.12), // M3 disabled bg
}

// M3 progress bar drawing constants.
const (
	m3ProgressBarRadius    float32 = 4
	m3ProgressBarFontSize  float32 = 11
	m3ProgressBarTextAlign         = widget.TextAlignCenter
)

// Compile-time check that ProgressBarPainter implements Painter.
var _ progressbar.Painter = ProgressBarPainter{}
