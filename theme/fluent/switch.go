package fluent

import (
	"time"

	"github.com/gogpu/ui/animation"
	"github.com/gogpu/ui/core/toggleswitch"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// SwitchPainter renders switches as WinUI 3 ToggleSwitch controls.
//
// The 40×20 track has a 1 px border while off and fills with the accent
// color when on. The knob grows on hover, stretches toward the middle while
// pressed, and slides with the Fluent fast-out curve.
//
// If Theme is nil, SwitchPainter falls back to the default Fluent Blue
// light palette.
type SwitchPainter struct {
	Theme *Theme // nil uses default Fluent Blue fallback
}

// flSwitchColors holds the WinUI ToggleSwitch color tokens for one theme.
type flSwitchColors struct {
	offFill         widget.Color // ControlAltFillColorSecondary
	offFillHover    widget.Color // ControlAltFillColorTertiary
	offFillPressed  widget.Color // ControlAltFillColorQuarternary
	offFillDisabled widget.Color // ControlAltFillColorDisabled
	offStroke       widget.Color // ControlStrongStrokeColorDefault
	offStrokeDis    widget.Color // ControlStrongStrokeColorDisabled
	knobOff         widget.Color // TextFillColorSecondary
	knobOffDisabled widget.Color // TextFillColorDisabled
	onFillDisabled  widget.Color // AccentFillColorDisabled
	knobOnDisabled  widget.Color // TextOnAccentFillColorDisabled
	label           widget.Color // TextFillColorPrimary
	labelDisabled   widget.Color // TextFillColorDisabled
	accent          widget.Color // AccentFillColorDefault
	knobOn          widget.Color // TextOnAccentFillColorPrimary
	focus           widget.Color
}

// resolveColors returns the switch colors for the painter's Theme.
func (p SwitchPainter) resolveColors() flSwitchColors {
	if p.Theme == nil {
		return flLightSwitchColors(&flDefaultSwitchScheme)
	}
	if p.Theme.IsDark() {
		return flDarkSwitchColors(&p.Theme.Colors)
	}
	return flLightSwitchColors(&p.Theme.Colors)
}

// flLightSwitchColors returns the WinUI light theme tokens.
func flLightSwitchColors(cs *ColorScheme) flSwitchColors {
	return flSwitchColors{
		offFill:         widget.HexA(0x00000006),
		offFillHover:    widget.HexA(0x0000000F),
		offFillPressed:  widget.HexA(0x00000018),
		offFillDisabled: widget.ColorTransparent,
		offStroke:       widget.HexA(0x00000072),
		offStrokeDis:    widget.HexA(0x00000037),
		knobOff:         widget.HexA(0x0000009E),
		knobOffDisabled: widget.HexA(0x0000005C),
		onFillDisabled:  widget.HexA(0x00000037),
		knobOnDisabled:  widget.ColorWhite,
		label:           cs.OnSurface,
		labelDisabled:   widget.HexA(0x0000005C),
		accent:          cs.Accent,
		knobOn:          cs.OnAccent,
		focus:           cs.StrokeFocus,
	}
}

// flDarkSwitchColors returns the WinUI dark theme tokens.
func flDarkSwitchColors(cs *ColorScheme) flSwitchColors {
	return flSwitchColors{
		offFill:         widget.HexA(0x00000019),
		offFillHover:    widget.HexA(0xFFFFFF0B),
		offFillPressed:  widget.HexA(0xFFFFFF12),
		offFillDisabled: widget.ColorTransparent,
		offStroke:       widget.HexA(0xFFFFFF8B),
		offStrokeDis:    widget.HexA(0xFFFFFF28),
		knobOff:         widget.HexA(0xFFFFFFC5),
		knobOffDisabled: widget.HexA(0xFFFFFF5D),
		onFillDisabled:  widget.HexA(0xFFFFFF28),
		knobOnDisabled:  widget.HexA(0xFFFFFF87),
		label:           cs.OnSurface,
		labelDisabled:   widget.HexA(0xFFFFFF5D),
		accent:          cs.Accent,
		knobOn:          cs.OnAccent,
		focus:           cs.StrokeFocus,
	}
}

// PaintSwitch renders a switch according to Fluent Design specifications.
func (p SwitchPainter) PaintSwitch(canvas widget.Canvas, state toggleswitch.PaintState) {
	if state.ControlBounds.IsEmpty() {
		return
	}

	colors := p.resolveColors()
	on := min(max(state.Position, 0), 1)
	track := flSwitchTrackRect(state.ControlBounds)

	offFill, offStroke, onFill := colors.offFillDisabled, colors.offStrokeDis, colors.onFillDisabled
	knobOff, knobOn := colors.knobOffDisabled, colors.knobOnDisabled
	var hover, press float32
	if !state.Disabled {
		hover, press = state.HoverAmount, state.PressAmount
		offFill = colors.offFill.Lerp(colors.offFillHover, hover).Lerp(colors.offFillPressed, press)
		offStroke = colors.offStroke
		alpha := flLerp(flLerp(1, flSwitchAccentHoverAlpha, hover), flSwitchAccentPressAlpha, press)
		onFill = colors.accent.WithAlpha(colors.accent.A * alpha)
		knobOff, knobOn = colors.knobOff, colors.knobOn
	}

	canvas.DrawRoundRect(track, offFill.Lerp(onFill, on), flSwitchTrackRadius)
	inset := flSwitchBorderWidth / 2
	canvas.StrokeRoundRect(track.Expand(-inset), offStroke.Lerp(onFill, on),
		flSwitchTrackRadius-inset, flSwitchBorderWidth)

	knob := flSwitchKnobRect(track, state.Position, hover, press)
	canvas.DrawRoundRect(knob, knobOff.Lerp(knobOn, on), knob.Height()/2)

	if state.Label != "" && !state.LabelBounds.IsEmpty() {
		fg := colors.label
		if state.Disabled {
			fg = colors.labelDisabled
		}
		canvas.DrawText(state.Label, state.LabelBounds, flSwitchFontSize, fg, false, widget.TextAlignLeft)
	}

	if state.Focused && !state.Disabled {
		// flDrawFocusRing adds flFocusRingOffset outside the rect it is
		// given. Pull the track back by that offset so the ring lands
		// flSwitchFocusRingOutset outside the track.
		back := flSwitchFocusRingOutset - flFocusRingOffset
		flDrawFocusRing(canvas, track.Expand(back), flSwitchTrackRadius+back, colors.focus)
	}
}

// flSwitchKnobRect returns the knob rectangle. Hovering grows the knob by
// 1 px on each side; pressing also stretches it 3 px toward the middle.
func flSwitchKnobRect(track geometry.Rect, pos, hover, press float32) geometry.Rect {
	grow := max(hover, press)
	h := flSwitchKnobSize + 2*grow
	w := h + flSwitchKnobStretch*press
	offX := flSwitchKnobInset - grow
	onX := flSwitchTrackWidth - flSwitchKnobInset - flSwitchKnobSize - grow - flSwitchKnobStretch*press
	return geometry.NewRect(
		track.Min.X+flLerp(offX, onX, pos),
		track.Min.Y+(flSwitchTrackHeight-h)/2,
		w, h,
	)
}

// flSwitchTrackRect returns the track centered in the control bounds.
func flSwitchTrackRect(control geometry.Rect) geometry.Rect {
	return geometry.NewRect(
		control.Min.X+(control.Width()-flSwitchTrackWidth)/2,
		control.Min.Y+(control.Height()-flSwitchTrackHeight)/2,
		flSwitchTrackWidth, flSwitchTrackHeight,
	)
}

// flLerp interpolates between a and b.
func flLerp(a, b, t float32) float32 {
	return a + (b-a)*t
}

// SwitchSize returns the Fluent switch size: the track plus room for the
// focus ring.
func (SwitchPainter) SwitchSize() geometry.Size {
	return geometry.Sz(flSwitchTrackWidth+2*flSwitchMargin, flSwitchTrackHeight+2*flSwitchMargin)
}

// SwitchThumbTravel returns the distance the knob moves between off and on.
func (SwitchPainter) SwitchThumbTravel() float32 {
	return flSwitchTrackWidth - 2*flSwitchKnobInset - flSwitchKnobSize
}

// SwitchLabelGap returns the Fluent gap between switch and label.
func (SwitchPainter) SwitchLabelGap() float32 { return flSwitchLabelGap }

// SwitchFontSize returns the Fluent label font size.
func (SwitchPainter) SwitchFontSize() float32 { return flSwitchFontSize }

// SwitchMotion returns the WinUI ToggleSwitch timing: a 167 ms knob slide
// and 83 ms hover and press reactions, on the Fluent fast-out curve.
func (SwitchPainter) SwitchMotion() toggleswitch.Motion {
	return toggleswitch.Motion{
		Toggle:       flSwitchToggleDuration,
		ToggleEasing: flSwitchEasing,
		Hover:        flSwitchReactionDuration,
		Press:        flSwitchReactionDuration,
	}
}

// flSwitchEasing is the Fluent "fast out, slow in" curve that WinUI uses
// for the ToggleSwitch knob.
var flSwitchEasing = animation.CubicBezier(0, 0, 0, 1)

// flDefaultSwitchScheme is the palette used when no Theme is provided.
var flDefaultSwitchScheme = LightScheme(DefaultAccentColor)

// Fluent switch drawing constants, from the WinUI 3 ToggleSwitch template.
const (
	flSwitchTrackWidth       float32 = 40
	flSwitchTrackHeight      float32 = 20
	flSwitchTrackRadius      float32 = flSwitchTrackHeight / 2
	flSwitchBorderWidth      float32 = 1
	flSwitchKnobSize         float32 = 12
	flSwitchKnobInset        float32 = 4
	flSwitchKnobStretch      float32 = 3
	flSwitchMargin           float32 = 4
	flSwitchFocusRingOutset  float32 = 3
	flSwitchLabelGap         float32 = 8
	flSwitchFontSize         float32 = 14
	flSwitchAccentHoverAlpha float32 = 0.9
	flSwitchAccentPressAlpha float32 = 0.8

	flSwitchToggleDuration   = 167 * time.Millisecond
	flSwitchReactionDuration = 83 * time.Millisecond
)

// Compile-time checks.
var (
	_ toggleswitch.Painter        = SwitchPainter{}
	_ toggleswitch.LayoutMetrics  = SwitchPainter{}
	_ toggleswitch.MotionProvider = SwitchPainter{}
)
