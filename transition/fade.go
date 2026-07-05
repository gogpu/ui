package transition

import (
	"time"

	"github.com/gogpu/ui/animation"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// Fade animates its child's opacity from transparent to opaque (fade-in)
// or opaque to transparent (fade-out).
//
// Fade uses a background-color overlay approach: the child is drawn
// normally, then a rectangle with the background color at (1-opacity)
// alpha is drawn on top. This simulates opacity without requiring
// offscreen rendering or per-pixel alpha compositing.
//
// When the canvas implements [OpacityPusher], true per-pixel opacity
// is used instead of the overlay approach.
//
// Create a Fade with [NewFade] and functional options:
//
//	fade := transition.NewFade(myWidget,
//	    transition.FadeDuration(200 * time.Millisecond),
//	    transition.FadeEasing(animation.EaseInOutCubic),
//	)
//
// By default, the animation starts automatically on [Fade.Mount] (autoStart).
// Call [Fade.FadeIn] or [Fade.FadeOut] to trigger animations manually.
type Fade struct {
	widget.WidgetBase

	child     widget.Widget
	duration  time.Duration
	easing    animation.Easing
	autoStart bool
	bgColor   widget.Color // background color for overlay approach

	// Animation state.
	opacity   float32   // 0.0 = invisible, 1.0 = fully opaque
	animating bool      // animation is in progress
	fadingIn  bool      // true = fading in (0->1), false = fading out (1->0)
	startTime time.Time // set on first Draw tick from ctx.Now()
	progress  float32   // raw animation progress 0-1
}

// Default fade configuration values.
const (
	defaultFadeDuration = 200 * time.Millisecond
)

// FadeOption configures a [Fade] widget during construction.
type FadeOption func(*fadeConfig)

// fadeConfig holds Fade configuration applied via functional options.
type fadeConfig struct {
	duration  time.Duration
	easing    animation.Easing
	autoStart *bool // nil = use default (true)
	bgColor   *widget.Color
}

// FadeDuration sets the animation duration. Defaults to 200ms.
func FadeDuration(d time.Duration) FadeOption {
	return func(c *fadeConfig) {
		c.duration = d
	}
}

// FadeEasing sets the easing function. Defaults to [animation.EaseInOutCubic].
func FadeEasing(e animation.Easing) FadeOption {
	return func(c *fadeConfig) {
		c.easing = e
	}
}

// FadeAutoStart controls whether the fade-in animation starts automatically
// on [Fade.Mount]. Defaults to true.
func FadeAutoStart(start bool) FadeOption {
	return func(c *fadeConfig) {
		c.autoStart = &start
	}
}

// FadeBackground sets the background color used for the overlay approach.
// When the canvas does not implement [OpacityPusher], the fade effect is
// achieved by drawing a semi-transparent rectangle of this color over the
// child. Defaults to white.
func FadeBackground(color widget.Color) FadeOption {
	return func(c *fadeConfig) {
		c.bgColor = &color
	}
}

// NewFade creates a new Fade transition wrapping the given child widget.
//
// The child fades from transparent to opaque on fade-in, or opaque to
// transparent on fade-out. By default the fade-in animation starts
// automatically when the widget is mounted.
//
//	fade := transition.NewFade(myWidget,
//	    transition.FadeDuration(200 * time.Millisecond),
//	    transition.FadeEasing(animation.EaseInOutCubic),
//	)
func NewFade(child widget.Widget, opts ...FadeOption) *Fade {
	cfg := fadeConfig{
		duration: defaultFadeDuration,
		easing:   animation.EaseInOutCubic,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	autoStart := true
	if cfg.autoStart != nil {
		autoStart = *cfg.autoStart
	}

	bgColor := widget.ColorWhite
	if cfg.bgColor != nil {
		bgColor = *cfg.bgColor
	}

	f := &Fade{
		child:     child,
		duration:  cfg.duration,
		easing:    cfg.easing,
		autoStart: autoStart,
		bgColor:   bgColor,
		opacity:   0, // start invisible for fade-in
	}
	f.SetVisible(true)
	f.SetEnabled(true)

	return f
}

// FadeIn triggers a fade-in animation (opacity 0 to 1).
//
// If an animation is already in progress, it is restarted from the
// beginning.
func (f *Fade) FadeIn() {
	f.fadingIn = true
	f.animating = true
	f.progress = 0
	f.startTime = time.Time{} // set on first Draw from ctx.Now()
}

// FadeOut triggers a fade-out animation (opacity 1 to 0).
//
// If an animation is already in progress, it is restarted from the
// beginning.
func (f *Fade) FadeOut() {
	f.fadingIn = false
	f.animating = true
	f.progress = 0
	f.opacity = 1.0 // ensure we start from fully visible
	f.startTime = time.Time{}
}

// IsAnimating reports whether a fade animation is currently in progress.
func (f *Fade) IsAnimating() bool {
	return f.animating
}

// Opacity returns the current opacity value (0.0 to 1.0).
func (f *Fade) Opacity() float32 {
	return f.opacity
}

// SetOpacity sets the opacity directly without animation.
// The value is clamped to [0, 1].
func (f *Fade) SetOpacity(opacity float32) {
	f.opacity = clampFloat32(opacity)
	f.animating = false
	f.SetNeedsRedraw(true)
}

// Child returns the wrapped child widget.
func (f *Fade) Child() widget.Widget {
	return f.child
}

// Layout delegates to the child and returns its preferred size.
func (f *Fade) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	if f.child == nil {
		return constraints.Constrain(geometry.Size{})
	}
	size := widget.LayoutChild(f.child, ctx, constraints)
	origin := f.Bounds().Min
	setChildBounds(f.child, geometry.FromPointSize(origin, size))
	return size
}

// Draw renders the child with a fade effect based on current opacity.
func (f *Fade) Draw(ctx widget.Context, canvas widget.Canvas) {
	if f.child == nil {
		return
	}

	// Update animation progress.
	if f.animating {
		f.updateAnimation(ctx)
	}

	// At zero opacity, skip drawing entirely.
	if f.opacity <= 0 {
		return
	}

	// At full opacity, draw child directly (no overhead).
	if f.opacity >= 1.0 {
		f.child.Draw(ctx, canvas)
		return
	}

	// Try OpacityPusher for true per-pixel opacity.
	if op, ok := canvas.(OpacityPusher); ok {
		op.PushOpacity(float64(f.opacity))
		f.child.Draw(ctx, canvas)
		op.PopOpacity()
		return
	}

	// Fallback: draw child, then overlay with background color at (1-opacity) alpha.
	f.child.Draw(ctx, canvas)
	overlayAlpha := 1.0 - f.opacity
	overlay := widget.RGBA(f.bgColor.R, f.bgColor.G, f.bgColor.B, overlayAlpha)
	bounds := childBoundsOf(f.child, f.Bounds().Size(), f.Bounds().Min)
	canvas.DrawRect(bounds, overlay)
}

// updateAnimation advances the animation based on elapsed time.
func (f *Fade) updateAnimation(ctx widget.Context) {
	now := ctx.Now()

	// Initialize start time on first tick.
	if f.startTime.IsZero() {
		f.startTime = now
	}

	elapsed := now.Sub(f.startTime)
	if f.duration > 0 {
		f.progress = float32(float64(elapsed) / float64(f.duration))
	} else {
		f.progress = 1.0
	}

	if f.progress >= 1.0 {
		f.progress = 1.0
		f.animating = false
	} else {
		f.SetNeedsRedraw(true)
	}

	// Compute opacity from eased progress.
	eased := f.progress
	if f.easing != nil {
		eased = f.easing(f.progress)
	}

	if f.fadingIn {
		f.opacity = eased // 0 -> 1
	} else {
		f.opacity = 1.0 - eased // 1 -> 0
	}
}

// Event dispatches events to the child widget.
func (f *Fade) Event(ctx widget.Context, e event.Event) bool {
	if f.child == nil {
		return false
	}
	return f.child.Event(ctx, e)
}

// Children returns the wrapped child as a single-element slice.
func (f *Fade) Children() []widget.Widget {
	if f.child == nil {
		return nil
	}
	return []widget.Widget{f.child}
}

// Mount starts the auto-start fade-in animation if enabled.
// Implements [widget.Lifecycle].
func (f *Fade) Mount(_ widget.Context) {
	if f.autoStart {
		f.FadeIn()
	}
}

// Unmount is called when the widget is removed from the tree.
// Implements [widget.Lifecycle].
func (f *Fade) Unmount() {
	f.animating = false
}

// clampFloat32 clamps v to the range [0, 1].
func clampFloat32(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// Compile-time interface checks.
var (
	_ widget.Widget    = (*Fade)(nil)
	_ widget.Lifecycle = (*Fade)(nil)
)
