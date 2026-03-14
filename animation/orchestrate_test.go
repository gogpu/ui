package animation

import (
	"math"
	"testing"
	"time"
)

func TestStaggerBasic(t *testing.T) {
	sig1 := newMockSignal(0)
	sig2 := newMockSignal(0)
	sig3 := newMockSignal(0)
	ctrl := NewController()

	Stagger(50*time.Millisecond,
		To(sig1, 1.0).From(0.0).Duration(100*time.Millisecond).Ease(Linear),
		To(sig2, 1.0).From(0.0).Duration(100*time.Millisecond).Ease(Linear),
		To(sig3, 1.0).From(0.0).Duration(100*time.Millisecond).Ease(Linear),
	).Start(ctrl)

	// At t=50ms: sig1 at 50%, sig2 just started (0%), sig3 not started.
	ctrl.Tick(50 * time.Millisecond)
	if math.Abs(float64(sig1.Get()-0.5)) > 0.05 {
		t.Errorf("stagger sig1 at 50ms: got %v, want ~0.5", sig1.Get())
	}
	if sig2.Get() != 0 {
		t.Errorf("stagger sig2 at 50ms: got %v, want 0", sig2.Get())
	}
	if sig3.Get() != 0 {
		t.Errorf("stagger sig3 at 50ms: got %v, want 0", sig3.Get())
	}

	// At t=100ms: sig1 complete, sig2 at 50%, sig3 just started.
	ctrl.Tick(50 * time.Millisecond)
	if sig1.Get() != 1.0 {
		t.Errorf("stagger sig1 at 100ms: got %v, want 1.0", sig1.Get())
	}
	if math.Abs(float64(sig2.Get()-0.5)) > 0.05 {
		t.Errorf("stagger sig2 at 100ms: got %v, want ~0.5", sig2.Get())
	}
	if sig3.Get() != 0 {
		t.Errorf("stagger sig3 at 100ms: got %v, want 0", sig3.Get())
	}

	// At t=200ms: all complete.
	ctrl.Tick(100 * time.Millisecond)
	if sig1.Get() != 1.0 {
		t.Errorf("stagger sig1 at 200ms: got %v, want 1.0", sig1.Get())
	}
	if sig2.Get() != 1.0 {
		t.Errorf("stagger sig2 at 200ms: got %v, want 1.0", sig2.Get())
	}
	if sig3.Get() != 1.0 {
		t.Errorf("stagger sig3 at 200ms: got %v, want 1.0", sig3.Get())
	}
}

func TestStaggerEmpty(t *testing.T) {
	ctrl := NewController()

	done := false
	Stagger(50 * time.Millisecond).OnDone(func() { done = true }).Start(ctrl)

	ctrl.Tick(time.Millisecond)
	if !done {
		t.Error("empty stagger should complete immediately")
	}
}

func TestStaggerSingleItem(t *testing.T) {
	sig := newMockSignal(0)
	ctrl := NewController()

	Stagger(50*time.Millisecond,
		To(sig, 1.0).From(0.0).Duration(100*time.Millisecond).Ease(Linear),
	).Start(ctrl)

	ctrl.Tick(100 * time.Millisecond)
	if sig.Get() != 1.0 {
		t.Errorf("stagger single: got %v, want 1.0", sig.Get())
	}
}

func TestChainBasic(t *testing.T) {
	sig1 := newMockSignal(0)
	sig2 := newMockSignal(0)
	ctrl := NewController()

	Chain(
		To(sig1, 1.0).From(0.0).Duration(50*time.Millisecond).Ease(Linear),
		To(sig2, 1.0).From(0.0).Duration(50*time.Millisecond).Ease(Linear),
	).Start(ctrl)

	ctrl.Tick(50 * time.Millisecond)
	if sig1.Get() != 1.0 {
		t.Errorf("chain sig1: got %v, want 1.0", sig1.Get())
	}
	if sig2.Get() != 0 {
		t.Errorf("chain sig2 before start: got %v, want 0", sig2.Get())
	}

	ctrl.Tick(50 * time.Millisecond)
	if sig2.Get() != 1.0 {
		t.Errorf("chain sig2: got %v, want 1.0", sig2.Get())
	}
}

func TestGroupBasic(t *testing.T) {
	sig1 := newMockSignal(0)
	sig2 := newMockSignal(0)
	ctrl := NewController()

	Group(
		To(sig1, 1.0).From(0.0).Duration(50*time.Millisecond).Ease(Linear),
		To(sig2, 2.0).From(0.0).Duration(100*time.Millisecond).Ease(Linear),
	).Start(ctrl)

	// Both run simultaneously.
	ctrl.Tick(50 * time.Millisecond)
	if sig1.Get() != 1.0 {
		t.Errorf("group sig1 at 50ms: got %v, want 1.0", sig1.Get())
	}
	if math.Abs(float64(sig2.Get()-1.0)) > 0.05 {
		t.Errorf("group sig2 at 50ms: got %v, want ~1.0", sig2.Get())
	}

	ctrl.Tick(50 * time.Millisecond)
	if sig2.Get() != 2.0 {
		t.Errorf("group sig2 at 100ms: got %v, want 2.0", sig2.Get())
	}
}

func TestWithDelayBasic(t *testing.T) {
	sig := newMockSignal(0)
	ctrl := NewController()

	builder := To(sig, 1.0).From(0.0).Duration(100 * time.Millisecond).Ease(Linear)
	delayed := WithDelay(50*time.Millisecond, builder)

	// Use in a parallel composition to start it.
	NewParallel(delayed).Start(ctrl)

	// During delay (30ms < 50ms delay).
	ctrl.Tick(30 * time.Millisecond)
	if sig.Get() != 0 {
		t.Errorf("WithDelay during delay: got %v, want 0", sig.Get())
	}

	// At 50ms total: delay exactly consumed, animation starts with 0ms overflow.
	ctrl.Tick(20 * time.Millisecond)
	// Animation just started, 0ms of progress.

	// At 100ms total: 50ms into the 100ms animation = 50%.
	ctrl.Tick(50 * time.Millisecond)
	if math.Abs(float64(sig.Get()-0.5)) > 0.05 {
		t.Errorf("WithDelay midpoint: got %v, want ~0.5", sig.Get())
	}

	// Complete the animation (50ms more).
	ctrl.Tick(50 * time.Millisecond)
	if sig.Get() != 1.0 {
		t.Errorf("WithDelay complete: got %v, want 1.0", sig.Get())
	}
}

func TestWithDelayOverflow(t *testing.T) {
	sig := newMockSignal(0)
	ctrl := NewController()

	// 50ms delay + 100ms animation.
	// Single tick of 100ms should consume 50ms delay + 50ms of animation.
	delayed := WithDelay(50*time.Millisecond,
		To(sig, 1.0).From(0.0).Duration(100*time.Millisecond).Ease(Linear),
	)
	NewParallel(delayed).Start(ctrl)

	ctrl.Tick(100 * time.Millisecond)
	if math.Abs(float64(sig.Get()-0.5)) > 0.05 {
		t.Errorf("WithDelay overflow: got %v, want ~0.5", sig.Get())
	}

	ctrl.Tick(50 * time.Millisecond)
	if sig.Get() != 1.0 {
		t.Errorf("WithDelay overflow complete: got %v, want 1.0", sig.Get())
	}
}

func TestWithDelayZero(t *testing.T) {
	sig := newMockSignal(0)
	ctrl := NewController()

	delayed := WithDelay(0, To(sig, 1.0).From(0.0).Duration(50*time.Millisecond).Ease(Linear))
	NewParallel(delayed).Start(ctrl)

	ctrl.Tick(50 * time.Millisecond)
	if sig.Get() != 1.0 {
		t.Errorf("WithDelay(0) complete: got %v, want 1.0", sig.Get())
	}
}

func TestRepeatNBasic(t *testing.T) {
	sig := newMockSignal(0)
	ctrl := NewController()

	done := false
	RepeatN(3,
		To(sig, 1.0).From(0.0).Duration(50*time.Millisecond).Ease(Linear),
	).OnDone(func() { done = true }).Start(ctrl)

	// First iteration.
	ctrl.Tick(50 * time.Millisecond)
	if done {
		t.Error("should not be done after 1 iteration")
	}

	// Second iteration.
	ctrl.Tick(50 * time.Millisecond)
	if done {
		t.Error("should not be done after 2 iterations")
	}

	// Third iteration.
	ctrl.Tick(50 * time.Millisecond)
	if !done {
		t.Error("should be done after 3 iterations")
	}
}

func TestRepeatNOne(t *testing.T) {
	sig := newMockSignal(0)
	ctrl := NewController()

	done := false
	RepeatN(1,
		To(sig, 1.0).From(0.0).Duration(50*time.Millisecond).Ease(Linear),
	).OnDone(func() { done = true }).Start(ctrl)

	ctrl.Tick(50 * time.Millisecond)
	if !done {
		t.Error("RepeatN(1) should complete after one iteration")
	}
}

func TestRepeatForever(t *testing.T) {
	sig := newMockSignal(0)
	ctrl := NewController()

	RepeatForever(
		To(sig, 1.0).From(0.0).Duration(50 * time.Millisecond).Ease(Linear),
	).Start(ctrl)

	// Should still be active after many iterations.
	for range 20 {
		ctrl.Tick(50 * time.Millisecond)
	}
	if !ctrl.HasActive() {
		t.Error("RepeatForever should still be active")
	}
}

func TestReverseAnimation(t *testing.T) {
	sig := newMockSignal(0)
	ctrl := NewController()

	// FadeIn goes 0 -> 1. Reverse should go 1 -> 0.
	reversed := Reverse(
		To(sig, 1.0).From(0.0).Duration(100 * time.Millisecond).Ease(Linear),
	)
	NewParallel(reversed).Start(ctrl)

	ctrl.Tick(50 * time.Millisecond)
	if math.Abs(float64(sig.Get()-0.5)) > 0.05 {
		t.Errorf("Reverse midpoint: got %v, want ~0.5", sig.Get())
	}

	ctrl.Tick(50 * time.Millisecond)
	if sig.Get() != 0.0 {
		t.Errorf("Reverse complete: got %v, want 0.0", sig.Get())
	}
}

func TestReverseNonAnimation(t *testing.T) {
	// For non-Animation types (like sequences), Reverse wraps in a reversed struct.
	sig := newMockSignal(0)
	ctrl := NewController()

	reversed := Reverse(
		NewSequence(
			To(sig, 1.0).From(0.0).Duration(50 * time.Millisecond).Ease(Linear),
		),
	)
	NewParallel(reversed).Start(ctrl)

	// The sequence should still complete (reversed wrapper delegates step).
	ctrl.Tick(50 * time.Millisecond)
	if sig.Get() != 1.0 {
		t.Errorf("Reverse sequence: got %v, want 1.0", sig.Get())
	}
}

func TestSequenceBuilderStartable(t *testing.T) {
	// SequenceBuilder can now be used as Startable in other compositions.
	sig1 := newMockSignal(0)
	sig2 := newMockSignal(0)
	ctrl := NewController()

	inner := NewSequence(
		To(sig1, 1.0).From(0.0).Duration(50 * time.Millisecond).Ease(Linear),
	)

	NewSequence(
		inner,
		To(sig2, 1.0).From(0.0).Duration(50*time.Millisecond).Ease(Linear),
	).Start(ctrl)

	ctrl.Tick(50 * time.Millisecond)
	if sig1.Get() != 1.0 {
		t.Errorf("nested sequence sig1: got %v, want 1.0", sig1.Get())
	}

	ctrl.Tick(50 * time.Millisecond)
	if sig2.Get() != 1.0 {
		t.Errorf("nested sequence sig2: got %v, want 1.0", sig2.Get())
	}
}

func TestParallelBuilderStartable(t *testing.T) {
	// ParallelBuilder can now be used as Startable in sequences.
	sig1 := newMockSignal(0)
	sig2 := newMockSignal(0)
	sig3 := newMockSignal(0)
	ctrl := NewController()

	par := NewParallel(
		To(sig1, 1.0).From(0.0).Duration(50*time.Millisecond).Ease(Linear),
		To(sig2, 1.0).From(0.0).Duration(50*time.Millisecond).Ease(Linear),
	)

	NewSequence(
		par,
		To(sig3, 1.0).From(0.0).Duration(50*time.Millisecond).Ease(Linear),
	).Start(ctrl)

	// Parallel part completes.
	ctrl.Tick(50 * time.Millisecond)
	if sig1.Get() != 1.0 || sig2.Get() != 1.0 {
		t.Errorf("parallel in sequence: sig1=%v, sig2=%v, want both 1.0", sig1.Get(), sig2.Get())
	}
	if sig3.Get() != 0 {
		t.Errorf("sig3 should not have started: got %v", sig3.Get())
	}

	// Sequential part.
	ctrl.Tick(50 * time.Millisecond)
	if sig3.Get() != 1.0 {
		t.Errorf("sequential after parallel: sig3=%v, want 1.0", sig3.Get())
	}
}

func TestStaggerWithPresets(t *testing.T) {
	sig1 := newMockSignal(0)
	sig2 := newMockSignal(0)
	ctrl := NewController()

	Stagger(30*time.Millisecond,
		FadeIn(sig1, 50*time.Millisecond).Ease(Linear),
		FadeIn(sig2, 50*time.Millisecond).Ease(Linear),
	).Start(ctrl)

	// At 50ms: sig1 complete, sig2 at (50-30)/50 = 40%.
	ctrl.Tick(50 * time.Millisecond)
	if sig1.Get() != 1.0 {
		t.Errorf("stagger preset sig1: got %v, want 1.0", sig1.Get())
	}

	// sig2 started at 30ms, so at 50ms it has 20ms of progress = 20/50 = 0.4.
	if math.Abs(float64(sig2.Get()-0.4)) > 0.05 {
		t.Errorf("stagger preset sig2 at 50ms: got %v, want ~0.4", sig2.Get())
	}

	// Complete.
	ctrl.Tick(30 * time.Millisecond)
	if sig2.Get() != 1.0 {
		t.Errorf("stagger preset sig2 complete: got %v, want 1.0", sig2.Get())
	}
}

func TestDelayedIsDone(t *testing.T) {
	sig := newMockSignal(0)
	builder := To(sig, 1.0).From(0.0).Duration(50 * time.Millisecond).Ease(Linear)
	ds := &delayedStartable{inner: builder, delay: 20 * time.Millisecond}
	d := ds.buildAnimatable()

	if d.isDone() {
		t.Error("delayed should not be done initially")
	}
	if d.signalKey() != nil {
		t.Error("delayed signalKey should be nil")
	}

	// Step through delay.
	d.step(20 * time.Millisecond)
	// Step through animation.
	d.step(50 * time.Millisecond)
	if !d.isDone() {
		t.Error("delayed should be done after completing")
	}
}

func TestRepeatingIsDone(t *testing.T) {
	sig := newMockSignal(0)
	rep := &repeating{
		factory:  To(sig, 1.0).From(0.0).Duration(50 * time.Millisecond).Ease(Linear),
		maxCount: 1,
	}

	if rep.isDone() {
		t.Error("repeating should not be done initially")
	}
	if rep.signalKey() != nil {
		t.Error("repeating signalKey should be nil")
	}
}

func TestReversedIsDone(t *testing.T) {
	sig := newMockSignal(0)
	inner := To(sig, 1.0).From(0.0).Duration(50 * time.Millisecond).Ease(Linear).Build()
	rev := &reversed{inner: inner}

	if rev.isDone() {
		t.Error("reversed should not be done initially")
	}
	if rev.signalKey() != nil {
		t.Error("reversed signalKey should be nil")
	}

	rev.step(50 * time.Millisecond)
	if !rev.isDone() {
		t.Error("reversed should be done after completing")
	}
}

func TestDelayedStepAlreadyDone(t *testing.T) {
	sig := newMockSignal(0)
	ds := &delayedStartable{
		inner: To(sig, 1.0).From(0.0).Duration(50 * time.Millisecond).Ease(Linear),
		delay: 0,
	}
	d := ds.buildAnimatable()

	d.step(50 * time.Millisecond)
	if !d.isDone() {
		t.Error("should be done")
	}

	// Stepping again should return true immediately.
	if !d.step(10 * time.Millisecond) {
		t.Error("step on done delayed should return true")
	}
}

func TestRepeatingStepAlreadyDone(t *testing.T) {
	sig := newMockSignal(0)
	rep := &repeating{
		factory:  To(sig, 1.0).From(0.0).Duration(50 * time.Millisecond).Ease(Linear),
		maxCount: 1,
	}

	rep.step(50 * time.Millisecond)
	if !rep.isDone() {
		t.Error("should be done")
	}

	// Stepping again should return true immediately.
	if !rep.step(10 * time.Millisecond) {
		t.Error("step on done repeating should return true")
	}
}

func TestChainOnDone(t *testing.T) {
	sig := newMockSignal(0)
	ctrl := NewController()

	done := false
	Chain(
		To(sig, 1.0).From(0.0).Duration(50 * time.Millisecond).Ease(Linear),
	).OnDone(func() { done = true }).Start(ctrl)

	ctrl.Tick(50 * time.Millisecond)
	if !done {
		t.Error("Chain OnDone was not called")
	}
}

func TestGroupOnDone(t *testing.T) {
	sig := newMockSignal(0)
	ctrl := NewController()

	done := false
	Group(
		To(sig, 1.0).From(0.0).Duration(50 * time.Millisecond).Ease(Linear),
	).OnDone(func() { done = true }).Start(ctrl)

	ctrl.Tick(50 * time.Millisecond)
	if !done {
		t.Error("Group OnDone was not called")
	}
}

func TestRepeatingBuilderStartable(t *testing.T) {
	sig := newMockSignal(0)
	ctrl := NewController()

	// RepeatingBuilder should work as Startable in compositions.
	repeater := RepeatN(2,
		To(sig, 1.0).From(0.0).Duration(50*time.Millisecond).Ease(Linear),
	)

	NewSequence(repeater).Start(ctrl)

	// Two iterations.
	ctrl.Tick(50 * time.Millisecond)
	ctrl.Tick(50 * time.Millisecond)

	if ctrl.HasActive() {
		t.Error("repeater in sequence should be done after 2 iterations")
	}
}
