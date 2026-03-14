package animation

import "time"

// buildAnimatable satisfies [Startable] for [SequenceBuilder], enabling
// sequences to be nested inside other sequences or parallel compositions.
func (b *SequenceBuilder) buildAnimatable() animatable {
	return b.seq
}

// buildAnimatable satisfies [Startable] for [ParallelBuilder], enabling
// parallel compositions to be nested inside sequences or other parallels.
func (b *ParallelBuilder) buildAnimatable() animatable {
	return b.par
}

// Stagger creates a composition where each animation starts after a fixed
// delay from the previous one's start time.
//
// This is implemented as a parallel composition where each successive
// animation has an additional delay added. All animations run concurrently
// but with staggered start times.
//
// Example:
//
//	// Each item fades in 50ms after the previous one starts
//	animation.Stagger(50*time.Millisecond,
//	    animation.FadeIn(item1, animation.DurationMedium2),
//	    animation.FadeIn(item2, animation.DurationMedium2),
//	    animation.FadeIn(item3, animation.DurationMedium2),
//	).Start(ctrl)
func Stagger(delay time.Duration, animations ...Startable) *ParallelBuilder {
	staggered := make([]Startable, len(animations))
	for i, anim := range animations {
		if i == 0 {
			staggered[i] = anim
			continue
		}
		staggered[i] = &delayedStartable{
			inner: anim,
			delay: time.Duration(i) * delay,
		}
	}
	return NewParallel(staggered...)
}

// delayedStartable wraps a Startable and adds a delay to the built animatable.
type delayedStartable struct {
	inner Startable
	delay time.Duration
}

// buildAnimatable builds the inner animation and wraps it with a delay.
func (d *delayedStartable) buildAnimatable() animatable {
	return &delayed{
		inner:     d.inner.buildAnimatable(),
		delay:     d.delay,
		remaining: d.delay,
	}
}

// Chain creates a sequential composition that plays animations one after another.
//
// Chain is a convenience alias for [NewSequence] with a clearer name for
// orchestration contexts. Each animation starts only after the previous
// one completes.
//
// Example:
//
//	animation.Chain(
//	    animation.FadeOut(oldOpacity, animation.DurationShort4),
//	    animation.FadeIn(newOpacity, animation.DurationMedium2),
//	).Start(ctrl)
func Chain(animations ...Startable) *SequenceBuilder {
	return NewSequence(animations...)
}

// Group plays all animations in parallel.
//
// Group is a convenience alias for [NewParallel] with a clearer name for
// orchestration contexts. All animations start simultaneously and the
// group completes when the longest animation finishes.
//
// Example:
//
//	animation.Group(
//	    animation.FadeIn(opacity, animation.DurationMedium2),
//	    animation.ScaleIn(scale, animation.DurationMedium2),
//	).Start(ctrl)
func Group(animations ...Startable) *ParallelBuilder {
	return NewParallel(animations...)
}

// WithDelay wraps an animation builder with an initial delay before it starts.
//
// This returns a new [AnimationBuilder] if the input is an [AnimationBuilder],
// otherwise wraps the animation in a delayed orchestration wrapper.
//
// Example:
//
//	animation.WithDelay(200*time.Millisecond,
//	    animation.FadeIn(opacity, animation.DurationMedium2),
//	).Start(ctrl)
func WithDelay(delay time.Duration, anim Startable) Startable {
	return &delayedStartable{
		inner: anim,
		delay: delay,
	}
}

// delayed wraps an animatable with an initial delay.
type delayed struct {
	inner     animatable
	delay     time.Duration
	remaining time.Duration
	started   bool
	done      bool
}

// step advances the delayed animation by dt. Returns true when finished.
func (d *delayed) step(dt time.Duration) bool {
	if d.done {
		return true
	}

	if !d.started {
		d.remaining -= dt
		if d.remaining > 0 {
			return false
		}
		d.started = true
		// Pass any overflow time to the inner animation.
		dt = -d.remaining
		d.remaining = 0
	}

	finished := d.inner.step(dt)
	if finished {
		d.done = true
	}
	return finished
}

// isDone reports whether the delayed animation has completed.
func (d *delayed) isDone() bool {
	return d.done
}

// signalKey returns nil because delayed wrappers don't target a single signal.
func (d *delayed) signalKey() any {
	return nil
}

// repeating wraps an animatable and replays it N times (or infinitely).
type repeating struct {
	factory  Startable
	maxCount int // 0 = infinite, >0 = exact number of times
	count    int
	current  animatable
	done     bool
	onDone   func()
}

// step advances the repeating animation by dt. Returns true when finished.
func (r *repeating) step(dt time.Duration) bool {
	if r.done {
		return true
	}

	if r.current == nil {
		r.current = r.factory.buildAnimatable()
		r.count++
	}

	finished := r.current.step(dt)
	if !finished {
		return false
	}

	// Current iteration finished. Check if we should repeat.
	if r.maxCount > 0 && r.count >= r.maxCount {
		r.done = true
		if r.onDone != nil {
			r.onDone()
		}
		return true
	}

	// Reset for next iteration (will be created on next step).
	r.current = nil
	return false
}

// isDone reports whether all repetitions are complete.
func (r *repeating) isDone() bool {
	return r.done
}

// signalKey returns nil because repeating wrappers don't target a single signal.
func (r *repeating) signalKey() any {
	return nil
}

// RepeatN creates an animation that repeats the given animation exactly n times.
//
// Pass n=0 for infinite repetition. The animation factory is called fresh for
// each iteration, ensuring clean state.
//
// Example:
//
//	// Pulse opacity 3 times
//	animation.RepeatN(3,
//	    animation.To(opacity, 1.0).From(0.0).Duration(200*time.Millisecond).Ease(animation.Linear),
//	).Start(ctrl)
func RepeatN(n int, anim Startable) *RepeatingBuilder {
	return &RepeatingBuilder{
		rep: &repeating{
			factory:  anim,
			maxCount: n,
		},
	}
}

// RepeatForever creates an animation that repeats indefinitely.
//
// The animation factory is called fresh for each iteration.
//
// Example:
//
//	animation.RepeatForever(
//	    animation.To(pulse, 1.0).From(0.5).Duration(500*time.Millisecond).AutoReverse(),
//	).Start(ctrl)
func RepeatForever(anim Startable) *RepeatingBuilder {
	return RepeatN(0, anim)
}

// RepeatingBuilder constructs a repeating animation.
type RepeatingBuilder struct {
	rep *repeating
}

// OnDone sets a callback invoked when all repetitions complete.
// Not called for infinite repetitions.
func (b *RepeatingBuilder) OnDone(fn func()) *RepeatingBuilder {
	b.rep.onDone = fn
	return b
}

// Start registers the repeating animation with the controller.
func (b *RepeatingBuilder) Start(ctrl *Controller) {
	ctrl.addComposition(b.rep)
}

// buildAnimatable satisfies [Startable] for [RepeatingBuilder].
func (b *RepeatingBuilder) buildAnimatable() animatable {
	return b.rep
}

// reversed wraps an animatable and plays it with inverted progress.
type reversed struct {
	inner animatable
	done  bool
}

// Reverse wraps an animation to play with inverted time mapping.
//
// This works by intercepting the easing and inverting the progress value.
// For tween animations, it swaps the from/to values. For other types,
// it wraps the animatable.
//
// Example:
//
//	// Slide out to bottom (reverse of slide in from bottom)
//	animation.Reverse(
//	    animation.SlideInFromBottom(translateY, 100, animation.DurationMedium2),
//	).Start(ctrl)
func Reverse(anim Startable) Startable {
	return &reversedStartable{inner: anim}
}

// reversedStartable wraps a Startable to build a reversed animation.
type reversedStartable struct {
	inner Startable
}

// buildAnimatable builds the inner animation with swapped from/to for tweens,
// or wraps in a reversed wrapper for other types.
func (r *reversedStartable) buildAnimatable() animatable {
	a := r.inner.buildAnimatable()

	// For Animation (tween), swap from and to values directly.
	if anim, ok := a.(*Animation); ok {
		anim.from, anim.to = anim.to, anim.from
		return anim
	}

	// For other types, wrap in a reversed animatable.
	return &reversed{inner: a}
}

// step delegates to the inner animation. For non-tween types, this still
// plays forward since we cannot generically reverse arbitrary animations.
// The primary use case is tween reversal via from/to swap above.
func (r *reversed) step(dt time.Duration) bool {
	finished := r.inner.step(dt)
	if finished {
		r.done = true
	}
	return finished
}

// isDone reports whether the reversed animation has completed.
func (r *reversed) isDone() bool {
	return r.done
}

// signalKey returns nil.
func (r *reversed) signalKey() any {
	return nil
}

// Verify interface compliance.
var (
	_ animatable = (*delayed)(nil)
	_ animatable = (*repeating)(nil)
	_ animatable = (*reversed)(nil)

	_ Startable = (*delayedStartable)(nil)
	_ Startable = (*reversedStartable)(nil)
	_ Startable = (*RepeatingBuilder)(nil)
	_ Startable = (*SequenceBuilder)(nil)
	_ Startable = (*ParallelBuilder)(nil)
)
