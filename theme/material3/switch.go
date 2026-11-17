package material3

import (
	"time"

	"github.com/gogpu/ui/animation"
	"github.com/gogpu/ui/core/toggleswitch"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// SwitchPainter renders switches using Material 3 design tokens.
//
// The thumb sits in a 52×32 pill track. It is small and outlined while off,
// grows when on, stretches into a pill on the way across, and grows again
// while pressed. The thumb lands with a slight overshoot, and hover and
// press show the M3 state layer around it.
//
// If Theme is nil, SwitchPainter falls back to the default M3 purple palette.
type SwitchPainter struct {
	Theme *Theme // nil uses default M3 purple fallback
}

// m3SwitchColors holds the color roles a switch uses.
type m3SwitchColors struct {
	primary                 widget.Color
	onPrimary               widget.Color
	primaryContainer        widget.Color
	surface                 widget.Color
	onSurface               widget.Color
	onSurfaceVariant        widget.Color
	surfaceContainerHighest widget.Color
	outline                 widget.Color
}

// resolveColors returns the switch colors of the painter's Theme, or the
// default M3 purple palette if Theme is nil.
func (p SwitchPainter) resolveColors() m3SwitchColors {
	if p.Theme == nil {
		return m3DefaultSwitchColors
	}
	cs := &p.Theme.Colors
	return m3SwitchColors{
		primary:                 cs.Primary,
		onPrimary:               cs.OnPrimary,
		primaryContainer:        cs.PrimaryContainer,
		surface:                 cs.Surface,
		onSurface:               cs.OnSurface,
		onSurfaceVariant:        cs.OnSurfaceVariant,
		surfaceContainerHighest: cs.SurfaceContainerHighest,
		outline:                 cs.Outline,
	}
}

// PaintSwitch renders a switch according to Material 3 specifications.
func (p SwitchPainter) PaintSwitch(canvas widget.Canvas, state toggleswitch.PaintState) {
	if state.ControlBounds.IsEmpty() {
		return
	}

	colors := p.resolveColors()
	on := m3SwitchColorProgress(&state)
	track := m3SwitchTrackRect(state.ControlBounds)
	m3PaintSwitchTrack(canvas, track, &colors, on, state.Disabled)

	center := geometry.Pt(
		track.Min.X+m3SwitchTrackHeight/2+state.Position*m3SwitchThumbTravel,
		track.Min.Y+m3SwitchTrackHeight/2,
	)

	// The state layer sits between the track and the thumb.
	if !state.Disabled {
		alpha := max(m3SwitchHoverAlpha*state.HoverAmount, m3SwitchPressAlpha*state.PressAmount)
		if alpha > 0 {
			layer := colors.onSurface.Lerp(colors.primary, on)
			canvas.DrawCircle(center, m3SwitchStateLayerRadius, layer.WithAlpha(alpha))
		}
	}

	size := m3SwitchThumbSize(&state)
	thumb := geometry.NewRect(center.X-size.Width/2, center.Y-size.Height/2, size.Width, size.Height)
	canvas.DrawRoundRect(thumb, m3SwitchThumbColor(&colors, &state, on), min(size.Width, size.Height)/2)

	if state.Label != "" && !state.LabelBounds.IsEmpty() {
		fg := colors.onSurface
		if state.Disabled {
			fg = fg.WithAlpha(m3SwitchDisabledContentAlpha)
		}
		canvas.DrawText(state.Label, state.LabelBounds, m3SwitchFontSize, fg, false, widget.TextAlignLeft)
	}

	if state.Focused && !state.Disabled {
		canvas.StrokeRoundRect(
			track.Expand(m3SwitchFocusRingOffset),
			colors.primary.WithAlpha(m3SwitchFocusRingAlpha),
			m3SwitchTrackRadius+m3SwitchFocusRingOffset,
			m3SwitchFocusRingStrokeWidth,
		)
	}
}

// m3PaintSwitchTrack fills the track and draws its outline, which fades out
// as the switch turns on.
func m3PaintSwitchTrack(canvas widget.Canvas, track geometry.Rect, c *m3SwitchColors, on float32, disabled bool) {
	offFill, onFill, outline := c.surfaceContainerHighest, c.primary, c.outline
	if disabled {
		offFill = c.surfaceContainerHighest.WithAlpha(m3SwitchDisabledTrackAlpha)
		onFill = c.onSurface.WithAlpha(m3SwitchDisabledTrackAlpha)
		outline = c.onSurface.WithAlpha(m3SwitchDisabledTrackAlpha)
	}
	canvas.DrawRoundRect(track, offFill.Lerp(onFill, on), m3SwitchTrackRadius)

	if on < 1 {
		inset := m3SwitchOutlineWidth / 2
		canvas.StrokeRoundRect(track.Expand(-inset), outline.WithAlpha(outline.A*(1-on)),
			m3SwitchTrackRadius-inset, m3SwitchOutlineWidth)
	}
}

// m3SwitchThumbColor returns the thumb color. Hover, press, and focus switch
// it to the emphasized variant of the current side.
func m3SwitchThumbColor(c *m3SwitchColors, state *toggleswitch.PaintState, on float32) widget.Color {
	if state.Disabled {
		return c.onSurface.WithAlpha(m3SwitchDisabledContentAlpha).Lerp(c.surface, on)
	}
	active := max(state.HoverAmount, state.PressAmount)
	if state.Focused {
		active = 1
	}
	off := c.outline.Lerp(c.onSurfaceVariant, active)
	onColor := c.onPrimary.Lerp(c.primaryContainer, active)
	return off.Lerp(onColor, on)
}

// m3SwitchColorProgress returns how far the colors have moved toward the on
// state. Like Flutter, colors change quickly at the start of each trip.
func m3SwitchColorProgress(state *toggleswitch.PaintState) float32 {
	if state.TurningOn {
		return m3SwitchColorEaseOut(state.Progress)
	}
	return m3SwitchColorEaseIn(state.Progress)
}

// m3SwitchThumbSize returns the thumb size for the travel progress, grown
// toward the pressed size by the press reaction.
func m3SwitchThumbSize(state *toggleswitch.PaintState) geometry.Size {
	size := m3SwitchTravelThumbSize(state.Progress, state.TurningOn)
	return m3LerpSize(size, m3SwitchPressedThumb, state.PressAmount)
}

// m3SwitchTravelThumbSize returns the thumb size at linear travel progress t.
// The thumb stretches into the transitional pill early in the trip and
// settles to its resting size before it arrives. Turning off plays the same
// sequence mirrored in time, as Flutter's Material 3 switch does.
func m3SwitchTravelThumbSize(t float32, turningOn bool) geometry.Size {
	if turningOn {
		switch {
		case t < m3SwitchStretchEnd:
			return m3LerpSize(m3SwitchInactiveThumb, m3SwitchTransitionalThumb,
				m3SwitchStretchEasing(t/m3SwitchStretchEnd))
		case t < m3SwitchSettleEnd:
			return m3LerpSize(m3SwitchTransitionalThumb, m3SwitchActiveThumb,
				m3SwitchSettleEasing((t-m3SwitchStretchEnd)/(m3SwitchSettleEnd-m3SwitchStretchEnd)))
		default:
			return m3SwitchActiveThumb
		}
	}

	switch {
	case t < 1-m3SwitchSettleEnd:
		return m3SwitchInactiveThumb
	case t < 1-m3SwitchStretchEnd:
		u := (t - (1 - m3SwitchSettleEnd)) / (m3SwitchSettleEnd - m3SwitchStretchEnd)
		return m3LerpSize(m3SwitchInactiveThumb, m3SwitchTransitionalThumb, 1-m3SwitchSettleEasing(1-u))
	default:
		u := (t - (1 - m3SwitchStretchEnd)) / m3SwitchStretchEnd
		return m3LerpSize(m3SwitchTransitionalThumb, m3SwitchActiveThumb, 1-m3SwitchStretchEasing(1-u))
	}
}

// m3LerpSize interpolates between two sizes.
func m3LerpSize(a, b geometry.Size, t float32) geometry.Size {
	return geometry.Sz(a.Width+(b.Width-a.Width)*t, a.Height+(b.Height-a.Height)*t)
}

// m3SwitchTrackRect returns the track centered in the control bounds.
func m3SwitchTrackRect(control geometry.Rect) geometry.Rect {
	return geometry.NewRect(
		control.Min.X+(control.Width()-m3SwitchTrackWidth)/2,
		control.Min.Y+(control.Height()-m3SwitchTrackHeight)/2,
		m3SwitchTrackWidth, m3SwitchTrackHeight,
	)
}

// SwitchSize returns the M3 switch size: the track plus room for the state
// layer around the thumb.
func (SwitchPainter) SwitchSize() geometry.Size {
	return geometry.Sz(m3SwitchTrackWidth+2*m3SwitchMargin, m3SwitchTrackHeight+2*m3SwitchMargin)
}

// SwitchThumbTravel returns the distance between the off and on thumb centers.
func (SwitchPainter) SwitchThumbTravel() float32 { return m3SwitchThumbTravel }

// SwitchLabelGap returns the M3 gap between switch and label.
func (SwitchPainter) SwitchLabelGap() float32 { return m3SwitchLabelGap }

// SwitchFontSize returns the M3 label font size.
func (SwitchPainter) SwitchFontSize() float32 { return m3SwitchFontSize }

// SwitchMotion returns the M3 switch timing: a 300 ms toggle that lands with
// easeOutBack, and 100 ms state layer reactions.
func (SwitchPainter) SwitchMotion() toggleswitch.Motion {
	return toggleswitch.Motion{
		Toggle:       m3SwitchToggleDuration,
		ToggleEasing: m3SwitchEaseOutBack,
		Hover:        m3SwitchReactionDuration,
		Press:        m3SwitchReactionDuration,
	}
}

// m3DefaultSwitchColors holds the baseline M3 purple roles used when no
// Theme is provided.
var m3DefaultSwitchColors = m3SwitchColors{
	primary:                 widget.Hex(0x6750A4),
	onPrimary:               widget.ColorWhite,
	primaryContainer:        widget.Hex(0xEADDFF),
	surface:                 widget.Hex(0xFEF7FF),
	onSurface:               widget.Hex(0x1D1B20),
	onSurfaceVariant:        widget.Hex(0x49454F),
	surfaceContainerHighest: widget.Hex(0xE6E0E9),
	outline:                 widget.Hex(0x79747E),
}

// M3 switch curves, as used by Flutter's Material 3 switch.
var (
	m3SwitchEaseOutBack   = animation.CubicBezier(0.175, 0.885, 0.32, 1.275)
	m3SwitchColorEaseOut  = animation.CubicBezier(0, 0, 0.58, 1)
	m3SwitchColorEaseIn   = animation.CubicBezier(0.42, 0, 1, 1)
	m3SwitchStretchEasing = animation.CubicBezier(0.31, 0, 0.56, 1)
	m3SwitchSettleEasing  = animation.CubicBezier(0.2, 0, 0, 1)
)

// M3 thumb sizes.
var (
	m3SwitchInactiveThumb     = geometry.Sz(16, 16)
	m3SwitchActiveThumb       = geometry.Sz(24, 24)
	m3SwitchTransitionalThumb = geometry.Sz(34, 22)
	m3SwitchPressedThumb      = geometry.Sz(28, 28)
)

// M3 switch drawing constants.
const (
	m3SwitchTrackWidth           float32 = 52
	m3SwitchTrackHeight          float32 = 32
	m3SwitchTrackRadius          float32 = 16
	m3SwitchOutlineWidth         float32 = 2
	m3SwitchThumbTravel          float32 = m3SwitchTrackWidth - m3SwitchTrackHeight
	m3SwitchStateLayerRadius     float32 = 20
	m3SwitchMargin               float32 = m3SwitchStateLayerRadius - m3SwitchTrackHeight/2
	m3SwitchLabelGap             float32 = 8
	m3SwitchFontSize             float32 = 14
	m3SwitchHoverAlpha           float32 = 0.08
	m3SwitchPressAlpha           float32 = 0.1
	m3SwitchDisabledTrackAlpha   float32 = 0.12
	m3SwitchDisabledContentAlpha float32 = 0.38
	m3SwitchFocusRingOffset      float32 = 2
	m3SwitchFocusRingStrokeWidth float32 = 2
	m3SwitchFocusRingAlpha       float32 = 0.7

	// Ends of the thumb stretch and settle phases, as fractions of the trip.
	m3SwitchStretchEnd float32 = 0.11
	m3SwitchSettleEnd  float32 = 0.83

	m3SwitchToggleDuration   = 300 * time.Millisecond
	m3SwitchReactionDuration = 100 * time.Millisecond
)

// Compile-time checks.
var (
	_ toggleswitch.Painter        = SwitchPainter{}
	_ toggleswitch.LayoutMetrics  = SwitchPainter{}
	_ toggleswitch.MotionProvider = SwitchPainter{}
)
