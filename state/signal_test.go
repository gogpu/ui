package state_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gogpu/ui/state"
)

func TestNewSignal(t *testing.T) {
	tests := []struct {
		name    string
		initial int
	}{
		{"zero", 0},
		{"positive", 42},
		{"negative", -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig := state.NewSignal(tt.initial)
			if got := sig.Get(); got != tt.initial {
				t.Errorf("NewSignal(%d).Get() = %d", tt.initial, got)
			}
		})
	}
}

func TestSignalSetAndGet(t *testing.T) {
	sig := state.NewSignal(0)
	sig.Set(10)
	if got := sig.Get(); got != 10 {
		t.Errorf("after Set(10), Get() = %d, want 10", got)
	}
	sig.Set(20)
	if got := sig.Get(); got != 20 {
		t.Errorf("after Set(20), Get() = %d, want 20", got)
	}
}

func TestSignalUpdate(t *testing.T) {
	sig := state.NewSignal(5)
	sig.Update(func(v int) int { return v * 2 })
	if got := sig.Get(); got != 10 {
		t.Errorf("after Update(*2), Get() = %d, want 10", got)
	}
}

func TestSignalSubscribeForever(t *testing.T) {
	sig := state.NewSignal(0)
	var received []int
	unsub := state.SubscribeForever(sig.AsReadonly(), func(v int) {
		received = append(received, v)
	})

	sig.Set(1)
	sig.Set(2)
	sig.Set(3)
	unsub()
	sig.Set(4) // should not be received

	if len(received) != 3 {
		t.Fatalf("expected 3 notifications, got %d", len(received))
	}
	want := []int{1, 2, 3}
	for i, v := range received {
		if v != want[i] {
			t.Errorf("received[%d] = %d, want %d", i, v, want[i])
		}
	}
}

func TestSignalSubscribeWithContext(t *testing.T) {
	sig := state.NewSignal(0)
	ctx, cancel := context.WithCancel(context.Background())

	var count atomic.Int32
	unsub := state.Subscribe(sig.AsReadonly(), ctx, func(_ int) {
		count.Add(1)
	})
	defer unsub()

	sig.Set(1)
	if got := count.Load(); got != 1 {
		t.Fatalf("after Set(1), count = %d, want 1", got)
	}

	cancel()
	// Give goroutine time to process context cancellation.
	time.Sleep(50 * time.Millisecond)

	sig.Set(2)
	// After context cancel the subscriber should be removed.
	// Give a small window for the notification to arrive (it should not).
	time.Sleep(20 * time.Millisecond)
	if got := count.Load(); got != 1 {
		t.Errorf("after cancel, count = %d, want 1", got)
	}
}

func TestSignalWithEqualFunc(t *testing.T) {
	sig := state.NewSignalWithOptions(0, state.Options[int]{
		Equal: func(a, b int) bool { return a == b },
	})

	var count int
	unsub := sig.SubscribeForever(func(_ int) { count++ })
	defer unsub()

	sig.Set(0) // same value, should not notify
	if count != 0 {
		t.Errorf("equal value notified, count = %d, want 0", count)
	}

	sig.Set(1) // different value, should notify
	if count != 1 {
		t.Errorf("different value not notified, count = %d, want 1", count)
	}
}

func TestSignalStringType(t *testing.T) {
	sig := state.NewSignal("hello")
	if got := sig.Get(); got != "hello" {
		t.Errorf("Get() = %q, want %q", got, "hello")
	}
	sig.Set("world")
	if got := sig.Get(); got != "world" {
		t.Errorf("after Set, Get() = %q, want %q", got, "world")
	}
}

func TestSignalConcurrentAccess(t *testing.T) {
	sig := state.NewSignal(0)
	var wg sync.WaitGroup

	const goroutines = 50
	const iterations = 100

	wg.Add(goroutines)
	for i := range goroutines {
		go func(id int) {
			defer wg.Done()
			for j := range iterations {
				sig.Set(id*iterations + j)
				_ = sig.Get()
			}
		}(i)
	}
	wg.Wait()

	// No panic or race is the success criterion.
	// Just verify we can still read.
	_ = sig.Get()
}

func TestSignalAsReadonly(t *testing.T) {
	sig := state.NewSignal(42)
	ro := sig.AsReadonly()
	if got := ro.Get(); got != 42 {
		t.Errorf("AsReadonly().Get() = %d, want 42", got)
	}

	sig.Set(100)
	if got := ro.Get(); got != 100 {
		t.Errorf("after Set(100), readonly.Get() = %d, want 100", got)
	}
}

func BenchmarkSignalGet(b *testing.B) {
	sig := state.NewSignal(42)
	b.ResetTimer()
	for range b.N {
		_ = sig.Get()
	}
}

func BenchmarkSignalSet(b *testing.B) {
	sig := state.NewSignal(0)
	b.ResetTimer()
	for i := range b.N {
		sig.Set(i)
	}
}

func BenchmarkSignalSetWithSubscriber(b *testing.B) {
	sig := state.NewSignal(0)
	unsub := sig.SubscribeForever(func(_ int) {})
	defer unsub()
	b.ResetTimer()
	for i := range b.N {
		sig.Set(i)
	}
}
