package toggleswitch

import (
	"time"

	"github.com/gogpu/ui/animation"
)

// Motion describes the timing of a switch's animations. Painters supply it
// through [MotionProvider]. A zero duration makes the matching change
// instant.
type Motion struct {
	// Toggle is the time the thumb takes to cross the whole track. Shorter
	// trips, such as settling after a drag or reversing mid-flight, take
	// proportionally less.
	Toggle time.Duration

	// ToggleEasing shapes each thumb trip. It is applied from wherever the
	// thumb currently is, so reversing mid-flight never makes it jump.
	// Nil means linear.
	ToggleEasing animation.Easing

	// Hover is the fade time of the hover highlight.
	Hover time.Duration

	// Press is the fade time of the press reaction.
	Press time.Duration
}

// MotionProvider lets a painter time the switch animations to match its
// design system. Painters that do not implement it get the timing of
// [DefaultPainter].
type MotionProvider interface {
	// SwitchMotion returns the animation timing for the switch.
	SwitchMotion() Motion
}

// tween moves one value toward a target over time. Each retarget starts a
// new segment from the value currently shown.
type tween struct {
	from, to float32
	start    time.Time
	duration time.Duration
	easing   animation.Easing
}

// fraction returns the linear completion of the current segment, in [0, 1].
func (t *tween) fraction(now time.Time) float32 {
	if t.duration <= 0 || t.from == t.to {
		return 1
	}
	elapsed := now.Sub(t.start)
	if elapsed <= 0 {
		return 0
	}
	if elapsed >= t.duration {
		return 1
	}
	return float32(elapsed) / float32(t.duration)
}

// linear returns the value at now without easing.
func (t *tween) linear(now time.Time) float32 {
	return t.from + (t.to-t.from)*t.fraction(now)
}

// value returns the eased value at now.
func (t *tween) value(now time.Time) float32 {
	f := t.fraction(now)
	if t.easing != nil && f > 0 && f < 1 {
		f = t.easing(f)
	}
	return t.from + (t.to-t.from)*f
}

// running reports whether the segment is still in progress at now.
func (t *tween) running(now time.Time) bool {
	return t.fraction(now) < 1
}

// rising reports whether the current segment moves toward larger values.
func (t *tween) rising() bool {
	return t.to > t.from
}

// jump shows v immediately, ending any segment in progress.
func (t *tween) jump(v float32) {
	t.from, t.to, t.duration = v, v, 0
}

// animateTo starts a segment from the value shown at now toward target.
// full is the duration of a complete 0 to 1 trip; the segment lasts in
// proportion to the distance it covers.
func (t *tween) animateTo(target float32, now time.Time, full time.Duration, easing animation.Easing) {
	if target == t.to {
		return
	}
	current := t.value(now)
	distance := target - current
	if distance < 0 {
		distance = -distance
	}
	distance = min(distance, 1)
	t.from, t.to = current, target
	t.start = now
	t.duration = time.Duration(float64(full) * float64(distance))
	t.easing = easing
}
