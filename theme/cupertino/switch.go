package cupertino

import (
	"time"

	"github.com/gogpu/ui/animation"
	"github.com/gogpu/ui/core/toggleswitch"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// SwitchPainter renders switches as iOS switches (UISwitch).
//
// The 51×31 track fills with system green when on. The white thumb slides
// with the CSS ease curve and stretches toward the middle while pressed.
// Hover tints the track slightly, a disabled switch is drawn at half
// opacity, and keyboard focus draws the macOS focus ring.
//
// If Theme is nil, SwitchPainter falls back to the default light palette.
type SwitchPainter struct {
	Theme *Theme // nil uses the default light palette
}

// cupSwitchColors holds the colors a switch uses.
type cupSwitchColors struct {
	on            widget.Color
	off           widget.Color
	offHover      widget.Color
	hoverTint     widget.Color
	background    widget.Color
	label         widget.Color
	disabledLabel widget.Color
	accent        widget.Color
}

// resolveColors returns the switch colors of the painter's Theme, or the
// default light palette if Theme is nil.
func (p SwitchPainter) resolveColors() cupSwitchColors {
	cs := &cupDefaultSwitchScheme
	if p.Theme != nil {
		cs = &p.Theme.Colors
	}
	return cupSwitchColors{
		on:            cs.SystemGreen,
		off:           cs.SecondarySystemFill,
		offHover:      cs.SystemFill,
		hoverTint:     cs.Label.WithAlpha(1),
		background:    cs.SystemBackground,
		label:         cs.Label,
		disabledLabel: cs.TertiaryLabel,
		accent:        cs.Accent,
	}
}

// PaintSwitch renders a switch according to Apple HIG specifications.
func (p SwitchPainter) PaintSwitch(canvas widget.Canvas, state toggleswitch.PaintState) {
	if state.ControlBounds.IsEmpty() {
		return
	}

	colors := p.resolveColors()
	opacity := float32(1)
	if state.Disabled {
		opacity = cupSwitchDisabledOpacity
	}
	pos := min(max(state.Position, 0), 1)
	track := cupSwitchTrackRect(state.ControlBounds)

	off := colors.off.Lerp(colors.offHover, state.HoverAmount)
	on := colors.on.Lerp(colors.hoverTint, cupSwitchHoverTint*state.HoverAmount)
	trackColor := off.Lerp(on, pos)
	canvas.DrawRoundRect(track, trackColor.WithAlpha(trackColor.A*opacity), cupSwitchTrackRadius)

	extension := cupSwitchThumbExtension * cupSwitchEase(state.PressAmount)
	thumb := cupSwitchThumbRect(track, pos, extension)
	cupPaintSwitchThumb(canvas, thumb, &colors, opacity)

	if state.Label != "" && !state.LabelBounds.IsEmpty() {
		fg := colors.label
		if state.Disabled {
			fg = colors.disabledLabel
		}
		canvas.DrawText(state.Label, state.LabelBounds, cupSwitchFontSize, fg, false, widget.TextAlignLeft)
	}

	if state.Focused && !state.Disabled {
		canvas.StrokeRoundRect(track.Expand(cupSwitchFocusRingOutset), cupFocusColor(colors.accent),
			cupSwitchTrackRadius+cupSwitchFocusRingOutset, cupSwitchFocusRingStroke)
	}
}

// cupPaintSwitchThumb draws the thumb with its soft shadow and hairline
// border. At reduced opacity the thumb is blended with the system
// background, matching a switch faded as a whole.
func cupPaintSwitchThumb(canvas widget.Canvas, thumb geometry.Rect, c *cupSwitchColors, opacity float32) {
	radius := thumb.Height() / 2
	for _, s := range cupSwitchThumbShadows {
		canvas.DrawRoundRect(thumb.TranslateXY(0, s.dy).Expand(s.spread),
			widget.ColorBlack.WithAlpha(s.alpha*opacity), radius+s.spread)
	}
	canvas.DrawRoundRect(thumb.Expand(cupSwitchThumbBorder),
		widget.ColorBlack.WithAlpha(cupSwitchThumbBorderAlpha*opacity), radius+cupSwitchThumbBorder)

	fill := widget.ColorWhite
	if opacity < 1 {
		fill = fill.Lerp(c.background.WithAlpha(1), 1-opacity)
	}
	canvas.DrawRoundRect(thumb, fill, radius)
}

// cupSwitchThumbRect returns the thumb at position pos, stretched by
// extension toward the middle of the track, as in Flutter's CupertinoSwitch.
func cupSwitchThumbRect(track geometry.Rect, pos, extension float32) geometry.Rect {
	left := cupLerp(
		track.Min.X+cupSwitchInnerStart-cupSwitchThumbRadius,
		track.Min.X+cupSwitchInnerEnd-cupSwitchThumbRadius-extension,
		pos,
	)
	right := cupLerp(
		track.Min.X+cupSwitchInnerStart+cupSwitchThumbRadius+extension,
		track.Min.X+cupSwitchInnerEnd+cupSwitchThumbRadius,
		pos,
	)
	cy := track.Min.Y + track.Height()/2
	return geometry.NewRect(left, cy-cupSwitchThumbRadius, right-left, 2*cupSwitchThumbRadius)
}

// cupSwitchTrackRect returns the track centered in the control bounds.
func cupSwitchTrackRect(control geometry.Rect) geometry.Rect {
	return geometry.NewRect(
		control.Min.X+(control.Width()-cupSwitchTrackWidth)/2,
		control.Min.Y+(control.Height()-cupSwitchTrackHeight)/2,
		cupSwitchTrackWidth, cupSwitchTrackHeight,
	)
}

// cupLerp interpolates between a and b.
func cupLerp(a, b, t float32) float32 {
	return a + (b-a)*t
}

// SwitchSize returns the Cupertino switch size: the track plus room for the
// thumb shadow and the focus ring.
func (SwitchPainter) SwitchSize() geometry.Size {
	return geometry.Sz(cupSwitchTrackWidth+2*cupSwitchMargin, cupSwitchTrackHeight+2*cupSwitchMargin)
}

// SwitchThumbTravel returns the distance the thumb moves between off and on.
func (SwitchPainter) SwitchThumbTravel() float32 { return cupSwitchInnerEnd - cupSwitchInnerStart }

// SwitchLabelGap returns the Cupertino gap between switch and label.
func (SwitchPainter) SwitchLabelGap() float32 { return cupSwitchLabelGap }

// SwitchFontSize returns the Cupertino label font size.
func (SwitchPainter) SwitchFontSize() float32 { return cupSwitchFontSize }

// SwitchMotion returns the iOS switch timing: a 200 ms toggle and a 300 ms
// thumb stretch, both on the CSS ease curve.
func (SwitchPainter) SwitchMotion() toggleswitch.Motion {
	return toggleswitch.Motion{
		Toggle:       cupSwitchToggleDuration,
		ToggleEasing: cupSwitchEase,
		Hover:        cupSwitchHoverDuration,
		Press:        cupSwitchPressDuration,
	}
}

// cupSwitchShadow is one layer of the thumb shadow.
type cupSwitchShadow struct {
	dy, spread, alpha float32
}

// cupSwitchThumbShadows approximates the blurred UISwitch thumb shadow with
// stacked layers, widest first.
var cupSwitchThumbShadows = [...]cupSwitchShadow{
	{dy: 2, spread: 1.5, alpha: 0.03},
	{dy: 1.5, spread: 0.75, alpha: 0.04},
	{dy: 1, spread: 0, alpha: 0.05},
}

// cupSwitchEase is the CSS ease curve that UISwitch and Flutter's
// CupertinoSwitch animate with.
var cupSwitchEase = animation.CubicBezier(0.25, 0.1, 0.25, 1)

// cupDefaultSwitchScheme is the palette used when no Theme is provided.
var cupDefaultSwitchScheme = lightColors(systemBlue)

// Cupertino switch drawing constants, from UISwitch and Flutter's
// CupertinoSwitch.
const (
	cupSwitchTrackWidth       float32 = 51
	cupSwitchTrackHeight      float32 = 31
	cupSwitchTrackRadius      float32 = cupSwitchTrackHeight / 2
	cupSwitchThumbRadius      float32 = 14
	cupSwitchInnerStart       float32 = cupSwitchTrackHeight / 2
	cupSwitchInnerEnd         float32 = cupSwitchTrackWidth - cupSwitchTrackHeight/2
	cupSwitchThumbExtension   float32 = 7
	cupSwitchThumbBorder      float32 = 0.5
	cupSwitchThumbBorderAlpha float32 = 0.04
	cupSwitchMargin           float32 = 4
	cupSwitchLabelGap         float32 = 8
	cupSwitchFontSize         float32 = 15
	cupSwitchHoverTint        float32 = 0.08
	cupSwitchDisabledOpacity  float32 = 0.5
	cupSwitchFocusRingOutset  float32 = 1.75
	cupSwitchFocusRingStroke  float32 = 3.5

	cupSwitchToggleDuration = 200 * time.Millisecond
	cupSwitchPressDuration  = 300 * time.Millisecond
	cupSwitchHoverDuration  = 150 * time.Millisecond
)

// Compile-time checks.
var (
	_ toggleswitch.Painter        = SwitchPainter{}
	_ toggleswitch.LayoutMetrics  = SwitchPainter{}
	_ toggleswitch.MotionProvider = SwitchPainter{}
)
