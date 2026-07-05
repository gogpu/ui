package transition

import (
	"time"

	"github.com/gogpu/ui/animation"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// Slide animates its child's position using [widget.Canvas.PushTransform].
//
// The child slides from off-screen (one full dimension away) to its
// natural position (slide-in), or from its natural position to off-screen
// (slide-out). This is useful for notification slide-ins, page transitions,
// and drawer animations.
//
// Create a Slide with [NewSlide] and functional options:
//
//	slide := transition.NewSlide(myWidget,
//	    transition.SlideFrom(transition.FromTop),
//	    transition.SlideDuration(300 * time.Millisecond),
//	    transition.SlideEasing(animation.EaseOutCubic),
//	)
//
// By default, the animation starts automatically on [Slide.Mount] (autoStart).
// Call [Slide.SlideIn] or [Slide.SlideOut] to trigger animations manually.
type Slide struct {
	widget.WidgetBase

	child     widget.Widget
	direction Direction
	duration  time.Duration
	easing    animation.Easing
	autoStart bool
	reverse   bool // true = slide OUT on auto-start instead of IN

	// Animation state.
	progress  float32   // 0.0 = hidden (offscreen), 1.0 = fully visible
	animating bool      // animation is in progress
	slideIn   bool      // true = sliding in, false = sliding out
	startTime time.Time // set on first Draw tick from ctx.Now()
}

// Default slide configuration values.
const (
	defaultSlideDuration = 300 * time.Millisecond
)

// SlideOption configures a [Slide] widget during construction.
type SlideOption func(*slideConfig)

// slideConfig holds Slide configuration applied via functional options.
type slideConfig struct {
	direction Direction
	duration  time.Duration
	easing    animation.Easing
	autoStart *bool // nil = use default (true)
	reverse   bool
}

// SlideFrom sets the direction from which the child slides in.
// Defaults to [FromTop].
func SlideFrom(dir Direction) SlideOption {
	return func(c *slideConfig) {
		c.direction = dir
	}
}

// SlideDuration sets the animation duration. Defaults to 300ms.
func SlideDuration(d time.Duration) SlideOption {
	return func(c *slideConfig) {
		c.duration = d
	}
}

// SlideEasing sets the easing function. Defaults to [animation.EaseOutCubic].
func SlideEasing(e animation.Easing) SlideOption {
	return func(c *slideConfig) {
		c.easing = e
	}
}

// SlideAutoStart controls whether the animation starts automatically on
// [Slide.Mount]. Defaults to true.
func SlideAutoStart(start bool) SlideOption {
	return func(c *slideConfig) {
		c.autoStart = &start
	}
}

// SlideReverse makes the auto-start animation slide OUT instead of IN.
// When true and autoStart is enabled, the widget starts visible and
// slides out on Mount. Defaults to false.
func SlideReverse(rev bool) SlideOption {
	return func(c *slideConfig) {
		c.reverse = rev
	}
}

// NewSlide creates a new Slide transition wrapping the given child widget.
//
// The child's position is animated from off-screen to its natural position
// (or vice versa) based on the configured direction. By default the
// animation starts automatically when the widget is mounted.
//
//	slide := transition.NewSlide(myWidget,
//	    transition.SlideFrom(transition.FromTop),
//	    transition.SlideDuration(300 * time.Millisecond),
//	)
func NewSlide(child widget.Widget, opts ...SlideOption) *Slide {
	cfg := slideConfig{
		direction: FromTop,
		duration:  defaultSlideDuration,
		easing:    animation.EaseOutCubic,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	autoStart := true
	if cfg.autoStart != nil {
		autoStart = *cfg.autoStart
	}

	s := &Slide{
		child:     child,
		direction: cfg.direction,
		duration:  cfg.duration,
		easing:    cfg.easing,
		autoStart: autoStart,
		reverse:   cfg.reverse,
	}
	s.SetVisible(true)
	s.SetEnabled(true)

	// If autoStart with reverse, start fully visible (will slide out on mount).
	// Otherwise start hidden (will slide in on mount).
	if autoStart && cfg.reverse {
		s.progress = 1.0
	}

	return s
}

// SlideIn triggers a slide-in animation (offscreen to natural position).
//
// If an animation is already in progress, it is restarted from the
// beginning. The child becomes visible immediately.
func (s *Slide) SlideIn() {
	s.slideIn = true
	s.animating = true
	s.progress = 0
	s.startTime = time.Time{} // set on first Draw from ctx.Now()
}

// SlideOut triggers a slide-out animation (natural position to offscreen).
//
// If an animation is already in progress, it is restarted from the
// beginning.
func (s *Slide) SlideOut() {
	s.slideIn = false
	s.animating = true
	s.progress = 0
	s.startTime = time.Time{} // set on first Draw from ctx.Now()
}

// IsAnimating reports whether a slide animation is currently in progress.
func (s *Slide) IsAnimating() bool {
	return s.animating
}

// Progress returns the current animation progress (0.0 to 1.0).
//
// For slide-in: 0 = fully offscreen, 1 = fully visible.
// For slide-out: 0 = fully visible, 1 = fully offscreen.
func (s *Slide) Progress() float32 {
	return s.progress
}

// SetChild replaces the wrapped child widget.
func (s *Slide) SetChild(child widget.Widget) {
	s.child = child
}

// Child returns the wrapped child widget.
func (s *Slide) Child() widget.Widget {
	return s.child
}

// Layout delegates to the child and returns its preferred size.
func (s *Slide) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	if s.child == nil {
		return constraints.Constrain(geometry.Size{})
	}
	size := widget.LayoutChild(s.child, ctx, constraints)
	origin := s.Bounds().Min
	setChildBounds(s.child, geometry.FromPointSize(origin, size))
	return size
}

// Draw renders the child with a translation offset based on animation progress.
func (s *Slide) Draw(ctx widget.Context, canvas widget.Canvas) {
	if s.child == nil {
		return
	}

	// Update animation progress.
	if s.animating {
		s.updateAnimation(ctx)
	}

	// Compute translation offset.
	offset := s.computeOffset()

	// Apply transform, draw child, pop transform.
	if offset.X != 0 || offset.Y != 0 {
		canvas.PushTransform(offset)
		s.child.Draw(ctx, canvas)
		canvas.PopTransform()
	} else {
		s.child.Draw(ctx, canvas)
	}
}

// updateAnimation advances the animation based on elapsed time.
func (s *Slide) updateAnimation(ctx widget.Context) {
	now := ctx.Now()

	// Initialize start time on first tick.
	if s.startTime.IsZero() {
		s.startTime = now
	}

	elapsed := now.Sub(s.startTime)
	if s.duration > 0 {
		s.progress = float32(float64(elapsed) / float64(s.duration))
	} else {
		s.progress = 1.0
	}

	if s.progress >= 1.0 {
		s.progress = 1.0
		s.animating = false
	} else {
		s.SetNeedsRedraw(true)
	}
}

// computeOffset returns the translation offset for the current animation state.
func (s *Slide) computeOffset() geometry.Point {
	bounds := childBoundsOf(s.child, s.Bounds().Size(), s.Bounds().Min)
	bw := bounds.Width()
	bh := bounds.Height()

	// Apply easing to get visual progress.
	eased := s.progress
	if s.easing != nil {
		eased = s.easing(s.progress)
	}

	// Determine the full-offset direction vector.
	var fullX, fullY float32
	switch s.direction {
	case FromTop, ToTop:
		fullY = -bh
	case FromBottom, ToBottom:
		fullY = bh
	case FromLeft, ToLeft:
		fullX = -bw
	case FromRight, ToRight:
		fullX = bw
	}

	if s.slideIn {
		// Slide in: start at full offset, end at zero.
		// remaining = 1 - eased
		remaining := 1 - eased
		return geometry.Pt(fullX*remaining, fullY*remaining)
	}
	// Slide out: start at zero, end at full offset.
	return geometry.Pt(fullX*eased, fullY*eased)
}

// Event dispatches events to the child widget.
func (s *Slide) Event(ctx widget.Context, e event.Event) bool {
	if s.child == nil {
		return false
	}
	return s.child.Event(ctx, e)
}

// Children returns the wrapped child as a single-element slice.
func (s *Slide) Children() []widget.Widget {
	if s.child == nil {
		return nil
	}
	return []widget.Widget{s.child}
}

// Mount starts the auto-start animation if enabled.
// Implements [widget.Lifecycle].
func (s *Slide) Mount(_ widget.Context) {
	if s.autoStart {
		if s.reverse {
			s.SlideOut()
		} else {
			s.SlideIn()
		}
	}
}

// Unmount is called when the widget is removed from the tree.
// Implements [widget.Lifecycle].
func (s *Slide) Unmount() {
	s.animating = false
}

// Compile-time interface checks.
var (
	_ widget.Widget    = (*Slide)(nil)
	_ widget.Lifecycle = (*Slide)(nil)
)
