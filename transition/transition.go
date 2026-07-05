package transition

import (
	"time"

	"github.com/gogpu/ui/animation"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// OpacityPusher is an optional Canvas capability for opacity effects.
//
// If the Canvas passed to Draw implements this interface, fade effects
// will use PushOpacity/PopOpacity. Otherwise, fade effects are silently
// skipped (graceful degradation).
type OpacityPusher interface {
	// PushOpacity pushes an opacity multiplier onto the draw state stack.
	// Opacity is in the range [0, 1]. Nested calls multiply.
	PushOpacity(opacity float64)

	// PopOpacity removes the most recently pushed opacity.
	PopOpacity()
}

// Default configuration values.
const (
	defaultDuration = 250 * time.Millisecond
)

// Option configures a [Transition] widget.
type Option func(*transitionConfig)

// transitionConfig holds the configuration applied via functional options.
type transitionConfig struct {
	enterEffect Effect
	exitEffect  Effect
	duration    time.Duration
	easing      animation.Easing
}

// EnterEffect sets the effect played when the widget appears.
func EnterEffect(e Effect) Option {
	return func(c *transitionConfig) {
		c.enterEffect = e
	}
}

// ExitEffect sets the effect played when the widget disappears.
func ExitEffect(e Effect) Option {
	return func(c *transitionConfig) {
		c.exitEffect = e
	}
}

// Duration sets the animation duration for both enter and exit effects.
func Duration(d time.Duration) Option {
	return func(c *transitionConfig) {
		c.duration = d
	}
}

// Easing sets the easing function for both enter and exit effects.
// Defaults to [animation.EaseOutCubic] if not set.
func Easing(e animation.Easing) Option {
	return func(c *transitionConfig) {
		c.easing = e
	}
}

// Transition wraps a child widget with enter/exit animation effects.
//
// When Show is called, the enter effect is played and the child becomes
// visible. When Hide is called, the exit effect is played and the child
// is hidden after the animation completes.
//
// Transition implements [widget.Widget].
type Transition struct {
	widget.WidgetBase

	child       widget.Widget
	enterEffect Effect
	exitEffect  Effect
	duration    time.Duration
	easing      animation.Easing

	// Animation state.
	shown     bool    // logical visibility (true after enter, false after exit completes)
	animating bool    // animation is in progress
	entering  bool    // true = enter animation, false = exit animation
	progress  float64 // 0.0 -> 1.0 normalized progress
	startTime time.Time

	// Cached child size from last layout (used for slide/scale calculations).
	childSize geometry.Size
}

// Wrap creates a new Transition that wraps the given child widget.
//
// By default the widget starts visible with no animation effects.
// Use [EnterEffect], [ExitEffect], [Duration], and [Easing] options
// to configure the transition behavior.
//
//	wrapped := transition.Wrap(myWidget,
//	    transition.EnterEffect(transition.FadeIn()),
//	    transition.ExitEffect(transition.FadeOut()),
//	    transition.Duration(300 * time.Millisecond),
//	)
func Wrap(child widget.Widget, opts ...Option) *Transition {
	cfg := transitionConfig{
		enterEffect: None(),
		exitEffect:  None(),
		duration:    defaultDuration,
		easing:      animation.EaseOutCubic,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	t := &Transition{
		child:       child,
		enterEffect: cfg.enterEffect,
		exitEffect:  cfg.exitEffect,
		duration:    cfg.duration,
		easing:      cfg.easing,
		shown:       true,
	}
	t.SetVisible(true)
	t.SetEnabled(true)
	return t
}

// Child returns the wrapped child widget.
func (t *Transition) Child() widget.Widget {
	return t.child
}

// IsShown reports whether the widget is logically visible (enter complete
// or in progress). Returns false if hidden or exit animation is complete.
func (t *Transition) IsShown() bool {
	return t.shown
}

// IsAnimating reports whether a transition animation is currently in progress.
func (t *Transition) IsAnimating() bool {
	return t.animating
}

// Show makes the widget visible, playing the enter effect if configured.
//
// If an exit animation is currently playing, it is replaced by the
// enter animation starting from the beginning.
func (t *Transition) Show() {
	if t.shown && !t.animating {
		return // already fully visible
	}
	t.shown = true
	if t.enterEffect.IsNone() || t.duration <= 0 {
		t.animating = false
		t.progress = 1.0
		return
	}
	t.entering = true
	t.animating = true
	t.progress = 0
	t.startTime = time.Time{} // set on first Draw from ctx.Now()
}

// Hide hides the widget, playing the exit effect if configured.
//
// If an enter animation is currently playing, it is replaced by the
// exit animation starting from the beginning.
func (t *Transition) Hide() {
	if !t.shown && !t.animating {
		return // already fully hidden
	}
	if t.exitEffect.IsNone() || t.duration <= 0 {
		t.shown = false
		t.animating = false
		t.progress = 0
		return
	}
	t.entering = false
	t.animating = true
	t.progress = 0
	t.startTime = time.Time{} // set on first Draw from ctx.Now()
}

// Layout calculates the transition widget size by delegating to the child.
func (t *Transition) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	if t.child == nil {
		return constraints.Constrain(geometry.Size{})
	}
	size := widget.LayoutChild(t.child, ctx, constraints)
	t.childSize = size
	origin := t.Bounds().Min
	setChildBounds(t.child, geometry.FromPointSize(origin, size))
	return size
}

// Draw renders the child widget with transition effects applied.
func (t *Transition) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !t.shown && !t.animating {
		return // fully hidden, nothing to draw
	}
	if t.child == nil {
		return
	}

	// Animation progress is updated by TickAnimation (called before layout).
	// Draw only reads the current progress — no mutation here.

	// Determine the active effect and eased progress.
	eff, easedProgress := t.currentEffect()

	// Apply effects, draw child, then pop effects in reverse.
	t.drawWithEffects(ctx, canvas, eff, easedProgress)
}

// TickAnimation advances the transition animation based on elapsed time.
// Called by the framework BEFORE layout (ADR-032 GAP-3, Flutter pattern).
// Draw reads t.progress without mutation.
func (t *Transition) TickAnimation(ctx widget.Context) {
	if !t.animating {
		return
	}

	now := ctx.Now()

	if t.startTime.IsZero() {
		t.startTime = now
	}

	elapsed := now.Sub(t.startTime)
	if t.duration > 0 {
		t.progress = float64(elapsed) / float64(t.duration)
	} else {
		t.progress = 1.0
	}

	if t.progress >= 1.0 {
		t.progress = 1.0
		t.animating = false
		if !t.entering {
			t.shown = false
		}
		t.MarkNeedsLayout()
		t.SetNeedsRedraw(true)
	} else {
		t.MarkNeedsLayout()
		t.SetNeedsRedraw(true)
	}
}

// currentEffect returns the active effect and the eased progress value.
func (t *Transition) currentEffect() (Effect, float64) {
	eff := t.enterEffect
	if !t.entering {
		eff = t.exitEffect
	}

	eased := t.progress
	if t.easing != nil {
		eased = float64(t.easing(float32(t.progress)))
	}

	return eff, eased
}

// drawWithEffects applies the effect transforms, draws the child, and
// pops the transforms in reverse order.
func (t *Transition) drawWithEffects(
	ctx widget.Context,
	canvas widget.Canvas,
	eff Effect,
	progress float64,
) {
	pushedTransform := false
	pushedOpacity := false
	childBounds := childBoundsOf(t.child, t.childSize, t.Bounds().Min)

	// Slide: translate by fraction of widget size.
	if eff.TranslateXFraction != 0 || eff.TranslateYFraction != 0 {
		dx, dy := t.computeTranslation(eff, childBounds, progress)
		if dx != 0 || dy != 0 {
			canvas.PushTransform(geometry.Pt(dx, dy))
			pushedTransform = true
		}
	}

	// Scale: adjust bounds around center.
	if eff.ScaleStart >= 0 {
		scale := lerp(eff.ScaleStart, eff.ScaleEnd, progress)
		if scale != 1.0 {
			scaled := scaleBoundsFromCenter(childBounds, float32(scale))
			setChildBounds(t.child, scaled)
			// Restore original bounds after draw.
			defer setChildBounds(t.child, childBounds)
		}
	}

	// Opacity: use OpacityPusher if canvas supports it.
	if eff.OpacityStart >= 0 {
		opacity := lerp(eff.OpacityStart, eff.OpacityEnd, progress)
		if op, ok := canvas.(OpacityPusher); ok {
			op.PushOpacity(opacity)
			pushedOpacity = true
		}
	}

	// Draw the child widget.
	t.child.Draw(ctx, canvas)

	// Pop in reverse order.
	if pushedOpacity {
		if op, ok := canvas.(OpacityPusher); ok {
			op.PopOpacity()
		}
	}
	if pushedTransform {
		canvas.PopTransform()
	}
}

// Children returns the wrapped child as a single-element slice.
func (t *Transition) Children() []widget.Widget {
	if t.child == nil {
		return nil
	}
	return []widget.Widget{t.child}
}

// Event dispatches events to the child widget if visible.
func (t *Transition) Event(ctx widget.Context, e event.Event) bool {
	if !t.shown || t.child == nil {
		return false
	}
	return t.child.Event(ctx, e)
}

// computeTranslation calculates the pixel offset for slide effects.
func (t *Transition) computeTranslation(
	eff Effect,
	bounds geometry.Rect,
	progress float64,
) (float32, float32) {
	bw := float64(bounds.Width())
	bh := float64(bounds.Height())

	var dx, dy float64
	if t.entering {
		// Enter: offset -> 0 (start from fraction, end at zero).
		dx = lerp(eff.TranslateXFraction*bw, 0, progress)
		dy = lerp(eff.TranslateYFraction*bh, 0, progress)
	} else {
		// Exit: 0 -> offset (start at zero, end at fraction).
		dx = lerp(0, eff.TranslateXFraction*bw, progress)
		dy = lerp(0, eff.TranslateYFraction*bh, progress)
	}
	return float32(dx), float32(dy)
}

// scaleBoundsFromCenter returns bounds scaled around the center point.
func scaleBoundsFromCenter(bounds geometry.Rect, scale float32) geometry.Rect {
	cx := (bounds.Min.X + bounds.Max.X) / 2
	cy := (bounds.Min.Y + bounds.Max.Y) / 2
	hw := bounds.Width() / 2 * scale
	hh := bounds.Height() / 2 * scale
	return geometry.Rect{
		Min: geometry.Pt(cx-hw, cy-hh),
		Max: geometry.Pt(cx+hw, cy+hh),
	}
}

// setChildBounds sets bounds on a child widget via type assertion.
// This follows the same pattern used by primitives.BoxWidget.
func setChildBounds(child widget.Widget, bounds geometry.Rect) {
	if s, ok := child.(interface{ SetBounds(geometry.Rect) }); ok {
		s.SetBounds(bounds)
	}
}

// childBoundsOf returns the child's bounds via type assertion, or
// constructs bounds from the given size and origin as fallback.
func childBoundsOf(child widget.Widget, size geometry.Size, origin geometry.Point) geometry.Rect {
	if b, ok := child.(interface{ Bounds() geometry.Rect }); ok {
		return b.Bounds()
	}
	return geometry.FromPointSize(origin, size)
}

// Verify Transition implements Widget.
var _ widget.Widget = (*Transition)(nil)
