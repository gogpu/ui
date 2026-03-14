package transition

// Direction specifies the origin or target edge for slide effects.
type Direction uint8

// Slide directions for enter effects (From*) and exit effects (To*).
const (
	FromTop    Direction = iota // Slide in from top edge.
	FromBottom                  // Slide in from bottom edge.
	FromLeft                    // Slide in from left edge.
	FromRight                   // Slide in from right edge.
	ToTop                       // Slide out toward top edge.
	ToBottom                    // Slide out toward bottom edge.
	ToLeft                      // Slide out toward left edge.
	ToRight                     // Slide out toward right edge.
)

// Effect describes the visual transformation applied during a transition.
//
// Each field pair (Start/End) defines the range of animation for that
// property. A negative Start value means the property is not animated.
// Effect is a value type; create instances via the provided constructors.
type Effect struct {
	// OpacityStart is the starting opacity. Negative means opacity is not animated.
	OpacityStart float64
	// OpacityEnd is the ending opacity.
	OpacityEnd float64

	// TranslateXFraction is the starting X offset as a fraction of widget width.
	// For example, -1.0 means start one full width to the left.
	TranslateXFraction float64
	// TranslateYFraction is the starting Y offset as a fraction of widget height.
	TranslateYFraction float64

	// ScaleStart is the starting scale factor. Negative means scale is not animated.
	ScaleStart float64
	// ScaleEnd is the ending scale factor.
	ScaleEnd float64
}

// noOpacity is the sentinel indicating opacity is not animated.
const noOpacity = -1.0

// noScale is the sentinel indicating scale is not animated.
const noScale = -1.0

// None returns an effect that performs no animation (instant transition).
func None() Effect {
	return Effect{
		OpacityStart: noOpacity,
		ScaleStart:   noScale,
	}
}

// IsNone reports whether the effect performs no animation.
func (e Effect) IsNone() bool {
	hasOpacity := e.OpacityStart >= 0
	hasTranslate := e.TranslateXFraction != 0 || e.TranslateYFraction != 0
	hasScale := e.ScaleStart >= 0
	return !hasOpacity && !hasTranslate && !hasScale
}

// FadeIn returns an effect that fades opacity from 0 to 1.
func FadeIn() Effect {
	return Effect{
		OpacityStart: 0,
		OpacityEnd:   1,
		ScaleStart:   noScale,
	}
}

// FadeOut returns an effect that fades opacity from 1 to 0.
func FadeOut() Effect {
	return Effect{
		OpacityStart: 1,
		OpacityEnd:   0,
		ScaleStart:   noScale,
	}
}

// SlideIn returns a slide-in effect from the given direction.
//
// The widget slides from off-screen (one full dimension away) to its
// final position. Only From* directions are meaningful for enter effects;
// To* directions are treated as their From* equivalents.
func SlideIn(dir Direction) Effect {
	e := Effect{
		OpacityStart: noOpacity,
		ScaleStart:   noScale,
	}
	switch dir {
	case FromTop, ToTop:
		e.TranslateYFraction = -1.0
	case FromBottom, ToBottom:
		e.TranslateYFraction = 1.0
	case FromLeft, ToLeft:
		e.TranslateXFraction = -1.0
	case FromRight, ToRight:
		e.TranslateXFraction = 1.0
	}
	return e
}

// SlideOut returns a slide-out effect toward the given direction.
//
// The widget slides from its current position to off-screen. Only To*
// directions are meaningful for exit effects; From* directions are
// treated as their To* equivalents.
func SlideOut(dir Direction) Effect {
	e := Effect{
		OpacityStart: noOpacity,
		ScaleStart:   noScale,
	}
	switch dir {
	case ToTop, FromTop:
		e.TranslateYFraction = -1.0
	case ToBottom, FromBottom:
		e.TranslateYFraction = 1.0
	case ToLeft, FromLeft:
		e.TranslateXFraction = -1.0
	case ToRight, FromRight:
		e.TranslateXFraction = 1.0
	}
	return e
}

// ScaleIn returns an effect that scales from 0.8 to 1.0 with a fade in.
func ScaleIn() Effect {
	return Effect{
		OpacityStart: 0,
		OpacityEnd:   1,
		ScaleStart:   0.8,
		ScaleEnd:     1.0,
	}
}

// ScaleOut returns an effect that scales from 1.0 to 0.8 with a fade out.
func ScaleOut() Effect {
	return Effect{
		OpacityStart: 1,
		OpacityEnd:   0,
		ScaleStart:   1.0,
		ScaleEnd:     0.8,
	}
}

// lerp linearly interpolates between a and b by t.
func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}
