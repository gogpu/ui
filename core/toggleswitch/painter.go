package toggleswitch

import (
	"time"

	"github.com/gogpu/ui/animation"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// Painter draws the visual representation of a switch. Each design system
// provides its own implementation.
//
// If no Painter is set, the switch uses [DefaultPainter].
type Painter interface {
	PaintSwitch(canvas widget.Canvas, state PaintState)
}

// PaintState provides the current switch state to the painter.
//
// The boolean fields describe where the switch is heading; the float fields
// describe what is on screen right now, mid-animation included. Painters
// should draw from the float fields so that every transition animates.
type PaintState struct {
	Label string

	// Checked is the switch value. It flips as soon as the user toggles,
	// before the thumb has finished moving.
	Checked bool

	// Hovered reports that the pointer is over the switch.
	Hovered bool

	// Pressed reports that the pointer button or the Space key is held on
	// the switch, including while the thumb is dragged.
	Pressed bool

	// Dragging reports that the thumb follows the pointer. Position is then
	// the raw drag position, without easing.
	Dragging bool

	Focused  bool
	Disabled bool

	// Position places the thumb, from 0 (off) to 1 (on), with the painter's
	// [Motion.ToggleEasing] applied. Easings that overshoot can carry it
	// slightly outside [0, 1].
	Position float32

	// Progress is the thumb travel without easing, in [0, 1]. Painters use
	// it to drive secondary changes, such as a thumb that changes size on
	// the way.
	Progress float32

	// TurningOn reports the direction of the latest thumb trip: true while
	// moving toward on. At rest it equals Checked.
	TurningOn bool

	// HoverAmount fades the hover highlight from 0 (none) to 1 (full).
	// It stays 0 while the switch is disabled.
	HoverAmount float32

	// PressAmount fades the press reaction from 0 (none) to 1 (full).
	// It stays 0 while the switch is disabled.
	PressAmount float32

	// Bounds is the whole widget area.
	Bounds geometry.Rect

	// ControlBounds is the area of the switch graphic, sized by
	// [LayoutMetrics.SwitchSize] and centered vertically in Bounds.
	ControlBounds geometry.Rect

	// LabelBounds is the area for the label, right of the switch graphic.
	// It is empty when there is no label.
	LabelBounds geometry.Rect
}

// LayoutMetrics lets a painter size the switch. Painters that do not
// implement it get the metrics of [DefaultPainter].
type LayoutMetrics interface {
	// SwitchSize returns the size of the switch graphic, including room
	// for its focus ring and state highlights.
	SwitchSize() geometry.Size

	// SwitchThumbTravel returns the distance the thumb moves between off
	// and on. Dragging the pointer this far moves the thumb across.
	SwitchThumbTravel() float32

	// SwitchLabelGap returns the gap between the switch graphic and the label.
	SwitchLabelGap() float32

	// SwitchFontSize returns the label font size.
	SwitchFontSize() float32
}

// DefaultPainter draws a neutral switch without design system styling.
// It is useful for tests and as a reference for custom painters.
//
// DefaultPainter also implements [LayoutMetrics] and [MotionProvider],
// supplying the defaults used when a painter does not implement them.
type DefaultPainter struct{}

// PaintSwitch renders a gray-and-purple switch.
func (DefaultPainter) PaintSwitch(canvas widget.Canvas, state PaintState) {
	if state.ControlBounds.IsEmpty() {
		return
	}

	track := centeredRect(state.ControlBounds, defaultTrackWidth, defaultTrackHeight)
	radius := track.Height() / 2

	trackColor := defaultOffTrack.Lerp(defaultOnTrack, clampUnit(state.Position))
	trackColor = trackColor.Lerp(widget.ColorWhite, defaultHoverLighten*state.HoverAmount)
	thumbColor := widget.ColorWhite
	if state.Disabled {
		trackColor = trackColor.WithAlpha(trackColor.A * defaultDisabledAlpha)
		thumbColor = thumbColor.WithAlpha(defaultDisabledThumbAlpha)
	}
	canvas.DrawRoundRect(track, trackColor, radius)

	center := geometry.Pt(
		track.Min.X+radius+state.Position*(track.Width()-track.Height()),
		track.Min.Y+radius,
	)
	canvas.DrawCircle(center, defaultThumbRadius+defaultPressGrow*state.PressAmount, thumbColor)

	if state.Label != "" && !state.LabelBounds.IsEmpty() {
		labelColor := defaultLabelColor
		if state.Disabled {
			labelColor = defaultDisabledLabel
		}
		canvas.DrawText(state.Label, state.LabelBounds, defaultFontSize, labelColor, false, widget.TextAlignLeft)
	}

	if state.Focused && !state.Disabled {
		canvas.StrokeRoundRect(track.Expand(defaultFocusOffset), defaultFocusColor,
			radius+defaultFocusOffset, defaultFocusStroke)
	}
}

// SwitchSize returns the default switch graphic size.
func (DefaultPainter) SwitchSize() geometry.Size {
	return geometry.Sz(defaultTrackWidth+2*defaultMargin, defaultTrackHeight+2*defaultMargin)
}

// SwitchThumbTravel returns the default thumb travel.
func (DefaultPainter) SwitchThumbTravel() float32 { return defaultTrackWidth - defaultTrackHeight }

// SwitchLabelGap returns the default gap between switch and label.
func (DefaultPainter) SwitchLabelGap() float32 { return defaultLabelGap }

// SwitchFontSize returns the default label font size.
func (DefaultPainter) SwitchFontSize() float32 { return defaultFontSize }

// SwitchMotion returns the default animation timing.
func (DefaultPainter) SwitchMotion() Motion {
	return Motion{
		Toggle:       defaultToggleDuration,
		ToggleEasing: animation.EaseInOutCubic,
		Hover:        defaultFadeDuration,
		Press:        defaultFadeDuration,
	}
}

// resolveLayoutMetrics returns the painter's LayoutMetrics, falling back to
// DefaultPainter.
func resolveLayoutMetrics(p Painter) LayoutMetrics {
	if lm, ok := p.(LayoutMetrics); ok {
		return lm
	}
	return DefaultPainter{}
}

// resolveMotion returns the painter's Motion, falling back to DefaultPainter.
func resolveMotion(p Painter) Motion {
	if mp, ok := p.(MotionProvider); ok {
		return mp.SwitchMotion()
	}
	return DefaultPainter{}.SwitchMotion()
}

// centeredRect returns a w×h rectangle centered in r.
func centeredRect(r geometry.Rect, w, h float32) geometry.Rect {
	return geometry.NewRect(r.Min.X+(r.Width()-w)/2, r.Min.Y+(r.Height()-h)/2, w, h)
}

// clampUnit clamps v to [0, 1].
func clampUnit(v float32) float32 {
	return min(max(v, 0), 1)
}

// DefaultPainter geometry and timing.
const (
	defaultTrackWidth         float32 = 36
	defaultTrackHeight        float32 = 20
	defaultMargin             float32 = 4
	defaultThumbRadius        float32 = 7
	defaultPressGrow          float32 = 1
	defaultLabelGap           float32 = 8
	defaultFontSize           float32 = 14
	defaultFocusOffset        float32 = 2
	defaultFocusStroke        float32 = 2
	defaultHoverLighten       float32 = 0.12
	defaultDisabledAlpha      float32 = 0.38
	defaultDisabledThumbAlpha float32 = 0.7
	defaultToggleDuration             = 150 * time.Millisecond
	defaultFadeDuration               = 100 * time.Millisecond
)

// DefaultPainter colors.
var (
	defaultOnTrack       = widget.Hex(0x6750A4)
	defaultOffTrack      = widget.RGBA(0.62, 0.62, 0.65, 1)
	defaultLabelColor    = widget.RGBA(0.1, 0.1, 0.1, 1)
	defaultDisabledLabel = widget.RGBA(0.12, 0.12, 0.13, 0.38)
	defaultFocusColor    = widget.Hex(0x6750A4).WithAlpha(0.7)
)

// Compile-time checks.
var (
	_ Painter        = DefaultPainter{}
	_ LayoutMetrics  = DefaultPainter{}
	_ MotionProvider = DefaultPainter{}
)
