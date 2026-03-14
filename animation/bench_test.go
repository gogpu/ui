package animation

import (
	"testing"
	"time"
)

// benchSignal is a minimal signalFloat32 for benchmarks.
type benchSignal struct {
	value float32
}

func (s *benchSignal) Get() float32  { return s.value }
func (s *benchSignal) Set(v float32) { s.value = v }

func BenchmarkTweenTick(b *testing.B) {
	sig := &benchSignal{value: 0}
	anim := To(sig, 1.0).
		From(0.0).
		Duration(300 * time.Millisecond).
		Ease(M3Standard).
		Build()

	dt := 16 * time.Millisecond // ~60 FPS
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		// Reset animation state for each iteration.
		anim.elapsed = 0
		anim.started = false
		anim.done = false
		anim.iter = 0
		anim.step(dt)
	}
}

func BenchmarkSpringTick(b *testing.B) {
	sig := &benchSignal{value: 0}
	spring := SpringTo(sig, 100.0).
		Stiffness(StiffnessMedium).
		DampingRatio(DampingNoBouncy).
		Build()

	dt := 16 * time.Millisecond
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		// Reset spring state for each iteration.
		spring.position = 0
		spring.velocity = 0
		spring.done = false
		spring.step(dt)
	}
}

func BenchmarkSequenceTick10(b *testing.B) {
	signals := make([]*benchSignal, 10)
	builders := make([]Startable, 10)
	for i := range signals {
		signals[i] = &benchSignal{value: 0}
		builders[i] = To(signals[i], 1.0).
			From(0.0).
			Duration(100 * time.Millisecond).
			Ease(Linear)
	}

	dt := 16 * time.Millisecond
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		// Build a fresh sequence each iteration so it starts from scratch.
		seq := NewSequence(builders...)
		seq.seq.step(dt)
	}
}

func BenchmarkParallelTick10(b *testing.B) {
	signals := make([]*benchSignal, 10)
	builders := make([]Startable, 10)
	for i := range signals {
		signals[i] = &benchSignal{value: 0}
		builders[i] = To(signals[i], 1.0).
			From(0.0).
			Duration(100 * time.Millisecond).
			Ease(Linear)
	}

	dt := 16 * time.Millisecond
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		par := NewParallel(builders...)
		par.par.step(dt)
	}
}

func BenchmarkEasingEaseInOutCubic(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		// Vary t across [0,1] to avoid constant folding.
		t := float32(i%100) / 100.0
		EaseInOutCubic(t)
	}
}

func BenchmarkEasingM3Standard(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		t := float32(i%100) / 100.0
		M3Standard(t)
	}
}

func BenchmarkControllerTick(b *testing.B) {
	ctrl := NewController()
	// Add 10 concurrent animations.
	for i := 0; i < 10; i++ {
		sig := &benchSignal{value: 0}
		To(sig, 1.0).
			From(0.0).
			Duration(10 * time.Second). // Long duration so they don't finish.
			Ease(Linear).
			Start(ctrl)
	}

	dt := 16 * time.Millisecond
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		ctrl.Tick(dt)
	}
}

func BenchmarkTweenAt(b *testing.B) {
	tw := NewFloat32Tween(0.0, 100.0)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		t := float32(i%100) / 100.0
		tw.At(t)
	}
}
