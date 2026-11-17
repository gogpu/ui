package devtools

import (
	"time"

	"github.com/gogpu/ui/animation"
	"github.com/gogpu/ui/core/toggleswitch"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// SwitchPainter renders switches using DevTools design tokens, matching the
// JetBrains Int UI toggle.
//
// The 26×16 track is translucent gray with a small ring at the left while
// off, and fills with the accent color with a white dot at the right when
// on. The ring slides across and fills into the dot.
//
// If Theme is nil, SwitchPainter falls back to the default DevTools dark palette.
type SwitchPainter struct {
	Theme *Theme // nil uses default DevTools dark fallback
}

// dtSwitchColors holds the colors a switch uses.
type dtSwitchColors struct {
	offFill        widget.Color
	offFillHover   widget.Color
	offFillPressed widget.Color
	offKnob        widget.Color
	onFill         widget.Color
	onFillHover    widget.Color
	onFillPressed  widget.Color
	onKnob         widget.Color
	disabledFill   widget.Color
	disabledBorder widget.Color
	disabledKnob   widget.Color
	label          widget.Color
	labelDisabled  widget.Color
	focus          widget.Color
}

// resolveColors returns the switch colors for the painter's Theme.
func (p SwitchPainter) resolveColors() dtSwitchColors {
	if p.Theme == nil {
		return dtSwitchColorsFor(&dtDefaultSwitchScheme, true)
	}
	return dtSwitchColorsFor(&p.Theme.Colors, p.Theme.IsDark())
}

// dtSwitchColorsFor returns the switch colors for a color scheme.
func dtSwitchColorsFor(cs *ColorScheme, dark bool) dtSwitchColors {
	c := dtSwitchColors{
		onFill:        cs.Primary,
		onFillHover:   cs.PrimaryHover,
		onFillPressed: cs.PrimaryPress,
		onKnob:        cs.OnPrimary,
		label:         cs.OnSurface,
		labelDisabled: cs.OnSurfaceDisabled,
		focus:         cs.BorderFocus,
	}
	if dark {
		c.offFill = widget.HexA(0xFFFFFF3B)
		c.offFillHover = widget.HexA(0xFFFFFF4A)
		c.offFillPressed = widget.HexA(0xFFFFFF59)
		c.offKnob = widget.Hex(0xD1D3D9)
		c.disabledFill = widget.ColorTransparent
		c.disabledBorder = widget.Hex(0x33353B)
		c.disabledKnob = widget.Hex(0x5F6269)
		return c
	}
	c.offFill = widget.HexA(0x00000030)
	c.offFillHover = widget.HexA(0x0000003D)
	c.offFillPressed = widget.HexA(0x0000004A)
	c.offKnob = widget.ColorWhite
	c.disabledFill = widget.Hex(0xF7F8F9)
	c.disabledBorder = widget.Hex(0xDDDFE4)
	c.disabledKnob = widget.Hex(0xC3C5CB)
	return c
}

// PaintSwitch renders a switch according to DevTools design specifications.
func (p SwitchPainter) PaintSwitch(canvas widget.Canvas, state toggleswitch.PaintState) {
	if state.ControlBounds.IsEmpty() {
		return
	}

	colors := p.resolveColors()
	on := min(max(state.Position, 0), 1)
	track := dtSwitchTrackRect(state.ControlBounds)

	knobColor := colors.disabledKnob
	if state.Disabled {
		canvas.DrawRoundRect(track, colors.disabledFill, dtSwitchTrackRadius)
		inset := dtSwitchBorderWidth / 2
		canvas.StrokeRoundRect(track.Expand(-inset), colors.disabledBorder,
			dtSwitchTrackRadius-inset, dtSwitchBorderWidth)
	} else {
		off := colors.offFill.Lerp(colors.offFillHover, state.HoverAmount).
			Lerp(colors.offFillPressed, state.PressAmount)
		onFill := colors.onFill.Lerp(colors.onFillHover, state.HoverAmount).
			Lerp(colors.onFillPressed, state.PressAmount)
		canvas.DrawRoundRect(track, off.Lerp(onFill, on), dtSwitchTrackRadius)
		knobColor = colors.offKnob.Lerp(colors.onKnob, on)
	}

	dtPaintSwitchKnob(canvas, track, state.Position, on, knobColor)

	if state.Label != "" && !state.LabelBounds.IsEmpty() {
		fg := colors.label
		if state.Disabled {
			fg = colors.labelDisabled
		}
		canvas.DrawText(state.Label, state.LabelBounds, dtSwitchFontSize, fg, false, widget.TextAlignLeft)
	}

	if state.Focused && !state.Disabled {
		canvas.StrokeRoundRect(track.Expand(dtSwitchFocusRingOutset), colors.focus,
			dtSwitchTrackRadius+dtSwitchFocusRingOutset, dtSwitchFocusRingStroke)
	}
}

// dtPaintSwitchKnob draws the knob at position pos. It is a ring while off
// and fills into a larger dot as fill goes to 1.
func dtPaintSwitchKnob(canvas widget.Canvas, track geometry.Rect, pos, fill float32, color widget.Color) {
	center := geometry.Pt(
		track.Min.X+dtSwitchKnobOffX+pos*dtSwitchThumbTravel,
		track.Min.Y+track.Height()/2,
	)
	if fill >= 1 {
		canvas.DrawCircle(center, dtSwitchDotRadius, color)
		return
	}
	// Grow the ring's outer edge to the dot radius and close its hole.
	inner := (dtSwitchRingRadius - dtSwitchRingStroke/2) * (1 - fill)
	outer := dtSwitchRingRadius + dtSwitchRingStroke/2
	outer += (dtSwitchDotRadius - outer) * fill
	canvas.StrokeCircle(center, (outer+inner)/2, color, outer-inner)
}

// dtSwitchTrackRect returns the track centered in the control bounds.
func dtSwitchTrackRect(control geometry.Rect) geometry.Rect {
	return geometry.NewRect(
		control.Min.X+(control.Width()-dtSwitchTrackWidth)/2,
		control.Min.Y+(control.Height()-dtSwitchTrackHeight)/2,
		dtSwitchTrackWidth, dtSwitchTrackHeight,
	)
}

// SwitchSize returns the DevTools switch size: the track plus room for the
// focus ring.
func (SwitchPainter) SwitchSize() geometry.Size {
	return geometry.Sz(dtSwitchTrackWidth+2*dtSwitchMargin, dtSwitchTrackHeight+2*dtSwitchMargin)
}

// SwitchThumbTravel returns the distance the knob moves between off and on.
func (SwitchPainter) SwitchThumbTravel() float32 { return dtSwitchThumbTravel }

// SwitchLabelGap returns the DevTools gap between switch and label.
func (SwitchPainter) SwitchLabelGap() float32 { return dtSwitchLabelGap }

// SwitchFontSize returns the DevTools label font size.
func (SwitchPainter) SwitchFontSize() float32 { return dtSwitchFontSize }

// SwitchMotion returns the DevTools switch timing: a quick 100 ms slide and
// short hover and press fades.
func (SwitchPainter) SwitchMotion() toggleswitch.Motion {
	return toggleswitch.Motion{
		Toggle:       dtSwitchToggleDuration,
		ToggleEasing: animation.EaseOutQuad,
		Hover:        dtSwitchReactionDuration,
		Press:        dtSwitchReactionDuration,
	}
}

// dtDefaultSwitchScheme is the palette used when no Theme is provided.
var dtDefaultSwitchScheme = DarkScheme()

// DevTools switch drawing constants, from the Int UI toggle icons.
const (
	dtSwitchTrackWidth       float32 = 26
	dtSwitchTrackHeight      float32 = 16
	dtSwitchTrackRadius      float32 = dtSwitchTrackHeight / 2
	dtSwitchBorderWidth      float32 = 1
	dtSwitchMargin           float32 = 3
	dtSwitchKnobOffX         float32 = 8
	dtSwitchThumbTravel      float32 = 10
	dtSwitchRingRadius       float32 = 3
	dtSwitchRingStroke       float32 = 2
	dtSwitchDotRadius        float32 = 5
	dtSwitchFocusRingOutset  float32 = 2
	dtSwitchFocusRingStroke  float32 = 2
	dtSwitchLabelGap         float32 = 6
	dtSwitchFontSize         float32 = 13
	dtSwitchToggleDuration           = 100 * time.Millisecond
	dtSwitchReactionDuration         = 80 * time.Millisecond
)

// Compile-time checks.
var (
	_ toggleswitch.Painter        = SwitchPainter{}
	_ toggleswitch.LayoutMetrics  = SwitchPainter{}
	_ toggleswitch.MotionProvider = SwitchPainter{}
)
