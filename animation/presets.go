package animation

import "time"

// Preset easing aliases for convenience.
//
// These are shorthand aliases for the M3 easing curves defined in m3.go,
// using names that match the Material Design 3 motion specification directly.
var (
	// EasingStandard is the standard M3 easing for utility animations
	// that begin and end on screen. Alias for [M3Standard].
	EasingStandard = M3Standard

	// EasingStandardDecelerate is the M3 easing for utility animations
	// entering the screen. Alias for [M3StandardDecelerate].
	EasingStandardDecelerate = M3StandardDecelerate

	// EasingStandardAccelerate is the M3 easing for utility animations
	// exiting the screen. Alias for [M3StandardAccelerate].
	EasingStandardAccelerate = M3StandardAccelerate

	// EasingEmphasized is the primary M3 easing for common animations
	// on screen. Alias for [M3Emphasized].
	EasingEmphasized = M3Emphasized

	// EasingEmphasizedDecelerate is the M3 easing for styled animations
	// entering the screen. Alias for [M3EmphasizedDecelerate].
	EasingEmphasizedDecelerate = M3EmphasizedDecelerate

	// EasingEmphasizedAccelerate is the M3 easing for styled animations
	// exiting the screen. Alias for [M3EmphasizedAccelerate].
	EasingEmphasizedAccelerate = M3EmphasizedAccelerate
)

// FadeIn creates a tween animation builder that animates opacity from 0 to 1.
//
// Uses M3 standard decelerate easing (entering). The returned builder targets
// the provided signal and can be further configured before starting.
//
// Example:
//
//	opacity := state.NewSignal[float32](0)
//	animation.FadeIn(opacity, animation.DurationMedium2).Start(ctrl)
func FadeIn(signal signalFloat32, duration time.Duration) *AnimationBuilder {
	return To(signal, 1.0).
		From(0.0).
		Duration(duration).
		Ease(M3StandardDecelerate)
}

// FadeOut creates a tween animation builder that animates opacity from 1 to 0.
//
// Uses M3 standard accelerate easing (exiting). The returned builder targets
// the provided signal and can be further configured before starting.
//
// Example:
//
//	opacity := state.NewSignal[float32](1)
//	animation.FadeOut(opacity, animation.DurationMedium2).Start(ctrl)
func FadeOut(signal signalFloat32, duration time.Duration) *AnimationBuilder {
	return To(signal, 0.0).
		From(1.0).
		Duration(duration).
		Ease(M3StandardAccelerate)
}

// SlideInFromBottom creates a tween animation builder that animates a vertical
// offset from +distance to 0 (entering from below).
//
// Uses M3 emphasized decelerate easing (entering).
//
// Example:
//
//	translateY := state.NewSignal[float32](100)
//	animation.SlideInFromBottom(translateY, 100, animation.DurationMedium2).Start(ctrl)
func SlideInFromBottom(signal signalFloat32, distance float32, duration time.Duration) *AnimationBuilder {
	return To(signal, 0.0).
		From(distance).
		Duration(duration).
		Ease(M3EmphasizedDecelerate)
}

// SlideInFromTop creates a tween animation builder that animates a vertical
// offset from -distance to 0 (entering from above).
//
// Uses M3 emphasized decelerate easing (entering).
func SlideInFromTop(signal signalFloat32, distance float32, duration time.Duration) *AnimationBuilder {
	return To(signal, 0.0).
		From(-distance).
		Duration(duration).
		Ease(M3EmphasizedDecelerate)
}

// SlideInFromLeft creates a tween animation builder that animates a horizontal
// offset from -distance to 0 (entering from the left).
//
// Uses M3 emphasized decelerate easing (entering).
func SlideInFromLeft(signal signalFloat32, distance float32, duration time.Duration) *AnimationBuilder {
	return To(signal, 0.0).
		From(-distance).
		Duration(duration).
		Ease(M3EmphasizedDecelerate)
}

// SlideInFromRight creates a tween animation builder that animates a horizontal
// offset from +distance to 0 (entering from the right).
//
// Uses M3 emphasized decelerate easing (entering).
func SlideInFromRight(signal signalFloat32, distance float32, duration time.Duration) *AnimationBuilder {
	return To(signal, 0.0).
		From(distance).
		Duration(duration).
		Ease(M3EmphasizedDecelerate)
}

// ScaleIn creates a tween animation builder that animates scale from 0.8 to 1.0
// (growing into view).
//
// Uses M3 emphasized decelerate easing (entering).
//
// Example:
//
//	scale := state.NewSignal[float32](0.8)
//	animation.ScaleIn(scale, animation.DurationMedium2).Start(ctrl)
func ScaleIn(signal signalFloat32, duration time.Duration) *AnimationBuilder {
	return To(signal, 1.0).
		From(0.8).
		Duration(duration).
		Ease(M3EmphasizedDecelerate)
}

// ScaleOut creates a tween animation builder that animates scale from 1.0 to 0.8
// (shrinking out of view).
//
// Uses M3 emphasized accelerate easing (exiting).
func ScaleOut(signal signalFloat32, duration time.Duration) *AnimationBuilder {
	return To(signal, 0.8).
		From(1.0).
		Duration(duration).
		Ease(M3EmphasizedAccelerate)
}

// DialogEnter creates a composite animation for dialog enter transitions.
//
// The animation plays fade in and scale in simultaneously, matching the
// Material Design 3 dialog container transform pattern. Uses M3 emphasized
// decelerate easing with DurationMedium2 (300ms).
//
// Parameters:
//   - opacity: signal driven from 0 to 1
//   - scale: signal driven from 0.8 to 1.0
func DialogEnter(opacity, scale signalFloat32) *ParallelBuilder {
	return NewParallel(
		FadeIn(opacity, DurationMedium2),
		ScaleIn(scale, DurationMedium2),
	)
}

// DialogExit creates a composite animation for dialog exit transitions.
//
// The animation plays fade out and scale out simultaneously, matching the
// Material Design 3 dialog container transform pattern. Uses M3 emphasized
// accelerate easing with DurationShort4 (200ms).
//
// Parameters:
//   - opacity: signal driven from 1 to 0
//   - scale: signal driven from 1.0 to 0.8
func DialogExit(opacity, scale signalFloat32) *ParallelBuilder {
	return NewParallel(
		FadeOut(opacity, DurationShort4),
		ScaleOut(scale, DurationShort4),
	)
}

// MenuEnter creates a composite animation for menu enter transitions.
//
// The animation plays fade in and slide from top simultaneously, matching
// the Material Design 3 menu reveal pattern. Uses DurationShort4 (200ms).
//
// Parameters:
//   - opacity: signal driven from 0 to 1
//   - translateY: signal driven from -distance to 0
//   - distance: the slide distance in pixels
func MenuEnter(opacity, translateY signalFloat32, distance float32) *ParallelBuilder {
	return NewParallel(
		FadeIn(opacity, DurationShort4),
		SlideInFromTop(translateY, distance, DurationShort4),
	)
}

// MenuExit creates a composite animation for menu exit transitions.
//
// The animation plays fade out and slide to top simultaneously, matching
// the Material Design 3 menu dismiss pattern. Uses DurationShort3 (150ms).
//
// Parameters:
//   - opacity: signal driven from 1 to 0
//   - translateY: signal driven from 0 to -distance
//   - distance: the slide distance in pixels
func MenuExit(opacity, translateY signalFloat32, distance float32) *ParallelBuilder {
	return NewParallel(
		FadeOut(opacity, DurationShort3),
		To(translateY, -distance).
			From(0).
			Duration(DurationShort3).
			Ease(M3StandardAccelerate),
	)
}

// SnackbarEnter creates an animation for snackbar enter transitions.
//
// The animation slides the snackbar in from the bottom, matching the
// Material Design 3 snackbar pattern. Uses M3 emphasized decelerate
// easing with DurationMedium1 (250ms).
//
// Parameters:
//   - translateY: signal driven from +distance to 0
//   - distance: the slide distance in pixels
func SnackbarEnter(translateY signalFloat32, distance float32) *AnimationBuilder {
	return SlideInFromBottom(translateY, distance, DurationMedium1)
}

// SnackbarExit creates an animation for snackbar exit transitions.
//
// The animation fades the snackbar out, matching the Material Design 3
// snackbar pattern. Uses M3 standard accelerate easing with DurationShort4 (200ms).
//
// Parameters:
//   - opacity: signal driven from 1 to 0
func SnackbarExit(opacity signalFloat32) *AnimationBuilder {
	return FadeOut(opacity, DurationShort4)
}
