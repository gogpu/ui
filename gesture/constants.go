package gesture

import "time"

// Timing thresholds (from Flutter constants.dart, source-verified).
const (
	// PressTimeout is the duration before showing visual feedback (ripple).
	// The recognizer has not yet won the arena at this point.
	PressTimeout = 100 * time.Millisecond

	// LongPressTimeout is the duration a pointer must be held without
	// moving beyond slop to trigger a long-press gesture.
	LongPressTimeout = 500 * time.Millisecond

	// DoubleTapTimeout is the maximum time between taps for a multi-click
	// sequence. If more than this duration elapses between pointer-up and
	// the next pointer-down, the click count resets to 1.
	DoubleTapTimeout = 300 * time.Millisecond

	// DoubleTapMinTime is the minimum time between taps (anti-bounce).
	// Prevents hardware debounce glitches from being counted as double-taps.
	DoubleTapMinTime = 40 * time.Millisecond
)

// Spatial thresholds.
const (
	// TouchSlop is the minimum distance a touch pointer must move to be
	// considered a drag rather than a tap. Accounts for finger imprecision.
	// 18 logical pixels (Flutter kTouchSlop, Android ViewConfiguration).
	TouchSlop float32 = 18.0

	// PrecisePointerSlop is the minimum distance a mouse or trackpad pointer
	// must move to be considered a drag. Much smaller than TouchSlop because
	// precise pointers have sub-pixel accuracy.
	// 1 logical pixel (Flutter kPrecisePointerHitSlop).
	PrecisePointerSlop float32 = 1.0

	// DoubleTapSlop is the maximum distance between consecutive tap
	// positions for them to count as a multi-tap sequence (touch only).
	// Mouse has no distance constraint (cursor stays precise).
	// 100 logical pixels (Flutter kDoubleTapSlop).
	DoubleTapSlop float32 = 100.0
)

// Velocity thresholds.
const (
	// MinFlingVelocity is the minimum velocity (px/s) for a fling gesture.
	MinFlingVelocity float32 = 50.0

	// MaxFlingVelocity caps fling velocity to prevent extreme scrolling.
	MaxFlingVelocity float32 = 8000.0
)

// MaxClickCount is the maximum click count tracked.
// Chromium caps at 3 (single, double, triple). Going higher has no
// standard UI semantic.
const MaxClickCount = 3

// noPointer is the sentinel value for "no active pointer".
// PointerID 0 is valid (mouse is typically 1, but the W3C spec allows 0
// for system-generated events), so -1 is used to mean "no pointer tracked".
const noPointer = -1

// SlopForDevice returns the drag detection threshold for the given device kind.
func SlopForDevice(kind DeviceKind) float32 {
	if kind == DeviceKindTouch {
		return TouchSlop
	}
	return PrecisePointerSlop
}
