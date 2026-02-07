package state_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/gogpu/ui/state"
)

func TestNewComputed(t *testing.T) {
	src := state.NewSignal(5)
	comp := state.NewComputed(func() int {
		return src.Get() * 2
	}, src.AsReadonly())

	if got := comp.Get(); got != 10 {
		t.Errorf("computed.Get() = %d, want 10", got)
	}
}

func TestComputedAutoUpdate(t *testing.T) {
	src := state.NewSignal(3)
	comp := state.NewComputed(func() int {
		return src.Get() + 1
	}, src.AsReadonly())

	if got := comp.Get(); got != 4 {
		t.Errorf("initial computed = %d, want 4", got)
	}

	src.Set(10)
	if got := comp.Get(); got != 11 {
		t.Errorf("after src.Set(10), computed = %d, want 11", got)
	}
}

func TestComputedMultipleDeps(t *testing.T) {
	first := state.NewSignal("John")
	last := state.NewSignal("Doe")

	full := state.NewComputed(func() string {
		return first.Get() + " " + last.Get()
	}, first.AsReadonly(), last.AsReadonly())

	if got := full.Get(); got != "John Doe" {
		t.Errorf("computed = %q, want %q", got, "John Doe")
	}

	first.Set("Jane")
	if got := full.Get(); got != "Jane Doe" {
		t.Errorf("after first.Set, computed = %q, want %q", got, "Jane Doe")
	}

	last.Set("Smith")
	if got := full.Get(); got != "Jane Smith" {
		t.Errorf("after last.Set, computed = %q, want %q", got, "Jane Smith")
	}
}

func TestComputedChained(t *testing.T) {
	base := state.NewSignal(2)
	doubled := state.NewComputed(func() int {
		return base.Get() * 2
	}, base.AsReadonly())
	quadrupled := state.NewComputed(func() int {
		return doubled.Get() * 2
	}, doubled)

	if got := quadrupled.Get(); got != 8 {
		t.Errorf("quadrupled = %d, want 8", got)
	}

	base.Set(5)
	if got := quadrupled.Get(); got != 20 {
		t.Errorf("after base.Set(5), quadrupled = %d, want 20", got)
	}
}

func TestComputedSubscribe(t *testing.T) {
	src := state.NewSignal(0)
	comp := state.NewComputed(func() int {
		return src.Get() + 100
	}, src.AsReadonly())

	var received []int
	unsub := comp.SubscribeForever(func(v int) {
		received = append(received, v)
	})
	defer unsub()

	src.Set(1)
	src.Set(2)

	if len(received) < 2 {
		t.Fatalf("expected at least 2 notifications, got %d", len(received))
	}
	if received[0] != 101 {
		t.Errorf("received[0] = %d, want 101", received[0])
	}
	if received[1] != 102 {
		t.Errorf("received[1] = %d, want 102", received[1])
	}
}

func TestComputedLazyEvaluation(t *testing.T) {
	var evalCount int
	src := state.NewSignal(0)
	comp := state.NewComputed(func() int {
		evalCount++
		return src.Get()
	}, src.AsReadonly())

	// First Get triggers evaluation.
	_ = comp.Get()
	first := evalCount

	// Second Get should use cached value.
	_ = comp.Get()
	if evalCount != first {
		t.Errorf("second Get re-evaluated, count = %d", evalCount)
	}

	// After dependency change, next Get re-evaluates.
	src.Set(1)
	_ = comp.Get()
	if evalCount <= first {
		t.Errorf("Get after dep change did not re-evaluate, count = %d", evalCount)
	}
}

func TestComputedConcurrent(t *testing.T) {
	src := state.NewSignal(0)
	comp := state.NewComputed(func() string {
		return fmt.Sprintf("val=%d", src.Get())
	}, src.AsReadonly())

	var wg sync.WaitGroup
	const goroutines = 20
	const iterations = 50

	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			for j := range iterations {
				src.Set(j)
				_ = comp.Get()
			}
		}()
	}
	wg.Wait()
}

func TestComputedWithOptions(t *testing.T) {
	src := state.NewSignal(3)
	var panicCaught bool
	comp := state.NewComputedWithOptions(func() int {
		return src.Get() * 10
	}, state.Options[int]{
		OnPanic: func(_ any, _ []byte) {
			panicCaught = true
		},
	}, src.AsReadonly())

	if got := comp.Get(); got != 30 {
		t.Errorf("computed = %d, want 30", got)
	}
	// panicCaught should remain false in normal operation.
	if panicCaught {
		t.Error("unexpected panic caught")
	}
}

func BenchmarkComputedGet(b *testing.B) {
	src := state.NewSignal(42)
	comp := state.NewComputed(func() int {
		return src.Get() * 2
	}, src.AsReadonly())
	// Warm cache.
	_ = comp.Get()
	b.ResetTimer()
	for range b.N {
		_ = comp.Get()
	}
}
