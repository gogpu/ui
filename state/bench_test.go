package state

import (
	"testing"
)

func BenchmarkSignalGet(b *testing.B) {
	sig := NewSignal(42)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = sig.Get()
	}
}

func BenchmarkSignalSet(b *testing.B) {
	sig := NewSignal(0)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		sig.Set(42)
	}
}

func BenchmarkSignalSetWithSubscriber(b *testing.B) {
	sig := NewSignal(0)
	_ = SubscribeForever[int](sig, func(_ int) {})
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		sig.Set(42)
	}
}

func BenchmarkComputedGet(b *testing.B) {
	sig := NewSignal(10)
	comp := NewComputed(func() int {
		return sig.Get() * 2
	}, sig)
	// Warm up: first Get triggers computation.
	_ = comp.Get()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = comp.Get()
	}
}

func BenchmarkComputedRecompute(b *testing.B) {
	sig := NewSignal(0)
	comp := NewComputed(func() int {
		return sig.Get() * 2
	}, sig)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		sig.Set(i)
		_ = comp.Get()
	}
}

func BenchmarkEffectTrigger(b *testing.B) {
	sig := NewSignal(0)
	eff := NewEffect(func() {
		_ = sig.Get()
	}, sig)
	defer eff.Stop()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		sig.Set(i)
	}
}

func BenchmarkSignalChain10(b *testing.B) {
	// Build a chain: sig -> c1 -> c2 -> ... -> c10
	sig := NewSignal(0)
	var prev ReadonlySignal[int] = sig
	computeds := make([]ReadonlySignal[int], 10)
	for i := range computeds {
		p := prev // capture
		computeds[i] = NewComputed(func() int {
			return p.Get() + 1
		}, p)
		prev = computeds[i]
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		sig.Set(i)
		// Read the end of the chain to trigger all recomputations.
		_ = computeds[len(computeds)-1].Get()
	}
}
