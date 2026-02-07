package state_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

func TestSchedulerMarkDirtyAndFlush(t *testing.T) {
	var flushed []widget.Widget
	sched := state.NewScheduler(func(dirty []widget.Widget) {
		flushed = append(flushed, dirty...)
	})

	w1 := &mockWidget{name: "w1"}
	w2 := &mockWidget{name: "w2"}

	sched.MarkDirty(w1)
	sched.MarkDirty(w2)
	sched.Flush()

	if len(flushed) != 2 {
		t.Fatalf("flushed count = %d, want 2", len(flushed))
	}
}

func TestSchedulerDeduplication(t *testing.T) {
	var flushed []widget.Widget
	sched := state.NewScheduler(func(dirty []widget.Widget) {
		flushed = append(flushed, dirty...)
	})

	w := &mockWidget{name: "w"}
	sched.MarkDirty(w)
	sched.MarkDirty(w)
	sched.MarkDirty(w)
	sched.MarkDirty(w)
	sched.MarkDirty(w)
	sched.Flush()

	if len(flushed) != 1 {
		t.Fatalf("flushed count = %d, want 1 (dedup)", len(flushed))
	}
}

func TestSchedulerEmptyFlush(t *testing.T) {
	called := false
	sched := state.NewScheduler(func(_ []widget.Widget) {
		called = true
	})

	sched.Flush()
	if called {
		t.Error("empty flush should not call flushFn")
	}
}

func TestSchedulerFlushClearsPending(t *testing.T) {
	sched := state.NewScheduler(func(_ []widget.Widget) {})
	w := &mockWidget{name: "w"}

	sched.MarkDirty(w)
	if got := sched.PendingCount(); got != 1 {
		t.Fatalf("before flush, pending = %d, want 1", got)
	}

	sched.Flush()
	if got := sched.PendingCount(); got != 0 {
		t.Errorf("after flush, pending = %d, want 0", got)
	}
}

func TestSchedulerBatch(t *testing.T) {
	var flushCount int
	sched := state.NewScheduler(func(_ []widget.Widget) {
		flushCount++
	})

	w1 := &mockWidget{name: "w1"}
	w2 := &mockWidget{name: "w2"}
	w3 := &mockWidget{name: "w3"}

	sched.Batch(func() {
		sched.MarkDirty(w1)
		sched.MarkDirty(w2)
		sched.MarkDirty(w3)
	})

	if got := sched.PendingCount(); got != 3 {
		t.Fatalf("after batch, pending = %d, want 3", got)
	}

	sched.Flush()
	if flushCount != 1 {
		t.Errorf("flush count = %d, want 1", flushCount)
	}
}

func TestSchedulerBatchWithDedup(t *testing.T) {
	var flushed []widget.Widget
	sched := state.NewScheduler(func(dirty []widget.Widget) {
		flushed = dirty
	})

	w := &mockWidget{name: "w"}

	sched.Batch(func() {
		sched.MarkDirty(w)
		sched.MarkDirty(w)
		sched.MarkDirty(w)
	})

	sched.Flush()
	if len(flushed) != 1 {
		t.Errorf("flushed count = %d, want 1 (dedup inside batch)", len(flushed))
	}
}

func TestSchedulerMarkDirtyNil(t *testing.T) {
	sched := state.NewScheduler(func(_ []widget.Widget) {})
	sched.MarkDirty(nil) // should not panic
	if got := sched.PendingCount(); got != 0 {
		t.Errorf("pending = %d, want 0 after nil MarkDirty", got)
	}
}

func TestSchedulerConcurrentMarkDirty(t *testing.T) {
	var flushCount atomic.Int32
	sched := state.NewScheduler(func(_ []widget.Widget) {
		flushCount.Add(1)
	})

	const goroutines = 50
	const iterations = 100

	widgets := make([]*mockWidget, goroutines)
	for i := range widgets {
		widgets[i] = &mockWidget{name: "w"}
	}

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := range goroutines {
		go func(idx int) {
			defer wg.Done()
			for range iterations {
				sched.MarkDirty(widgets[idx])
			}
		}(i)
	}
	wg.Wait()

	// All widgets should be pending (deduplicated).
	if got := sched.PendingCount(); got > goroutines {
		t.Errorf("pending = %d, should be <= %d", got, goroutines)
	}
	if got := sched.PendingCount(); got == 0 {
		t.Error("pending should be > 0 after concurrent MarkDirty")
	}

	sched.Flush()
	if got := sched.PendingCount(); got != 0 {
		t.Errorf("after flush, pending = %d, want 0", got)
	}
}

func TestSchedulerMultipleFlushes(t *testing.T) {
	var flushCount int
	sched := state.NewScheduler(func(_ []widget.Widget) {
		flushCount++
	})

	w := &mockWidget{name: "w"}

	sched.MarkDirty(w)
	sched.Flush()

	sched.MarkDirty(w)
	sched.Flush()

	sched.MarkDirty(w)
	sched.Flush()

	if flushCount != 3 {
		t.Errorf("flush count = %d, want 3", flushCount)
	}
}

func TestSchedulerIsFlushing(t *testing.T) {
	var wasFlushing bool
	sched := state.NewScheduler(func(_ []widget.Widget) {
		// Cannot check IsFlushing from inside the callback because
		// the lock is held differently, but we verify the flag is set.
	})

	// Before flush.
	if sched.IsFlushing() {
		t.Error("should not be flushing before Flush()")
	}

	w := &mockWidget{name: "w"}
	sched.MarkDirty(w)

	// Use a scheduler that checks flushing state inside the callback.
	var innerFlushing bool
	var sched2 *state.Scheduler
	sched2 = state.NewScheduler(func(_ []widget.Widget) {
		innerFlushing = sched2.IsFlushing()
	})
	sched2.MarkDirty(w)
	sched2.Flush()
	wasFlushing = innerFlushing

	if !wasFlushing {
		t.Error("IsFlushing should be true during Flush callback")
	}
}

func TestNewSchedulerNilPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("NewScheduler(nil) should panic")
		}
	}()
	state.NewScheduler(nil)
}

func TestSchedulerPendingCount(t *testing.T) {
	sched := state.NewScheduler(func(_ []widget.Widget) {})

	if got := sched.PendingCount(); got != 0 {
		t.Errorf("initial pending = %d, want 0", got)
	}

	w1 := &mockWidget{name: "w1"}
	w2 := &mockWidget{name: "w2"}

	sched.MarkDirty(w1)
	if got := sched.PendingCount(); got != 1 {
		t.Errorf("pending = %d, want 1", got)
	}

	sched.MarkDirty(w2)
	if got := sched.PendingCount(); got != 2 {
		t.Errorf("pending = %d, want 2", got)
	}

	sched.MarkDirty(w1) // duplicate
	if got := sched.PendingCount(); got != 2 {
		t.Errorf("pending after dup = %d, want 2", got)
	}
}

func TestSchedulerWidgetsDuringFlush(t *testing.T) {
	// Widgets added during flush should not be in the current flush.
	w1 := &mockWidget{name: "w1"}
	w2 := &mockWidget{name: "w2"}

	var firstFlush []widget.Widget
	var sched *state.Scheduler

	sched = state.NewScheduler(func(dirty []widget.Widget) {
		if firstFlush == nil {
			firstFlush = dirty
			// Add another widget during flush.
			sched.MarkDirty(w2)
		}
	})

	sched.MarkDirty(w1)
	sched.Flush()

	if len(firstFlush) != 1 {
		t.Fatalf("first flush got %d widgets, want 1", len(firstFlush))
	}

	// w2 should now be pending.
	if got := sched.PendingCount(); got != 1 {
		t.Errorf("after first flush, pending = %d, want 1 (w2)", got)
	}
}

func BenchmarkSchedulerMarkDirty(b *testing.B) {
	sched := state.NewScheduler(func(_ []widget.Widget) {})
	w := &mockWidget{name: "bench"}
	b.ResetTimer()
	for range b.N {
		sched.MarkDirty(w)
	}
}

func BenchmarkSchedulerFlush(b *testing.B) {
	w := &mockWidget{name: "bench"}
	sched := state.NewScheduler(func(_ []widget.Widget) {})
	b.ResetTimer()
	for range b.N {
		sched.MarkDirty(w)
		sched.Flush()
	}
}

func BenchmarkSchedulerConcurrentMarkDirty(b *testing.B) {
	sched := state.NewScheduler(func(_ []widget.Widget) {})
	w := &mockWidget{name: "bench"}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sched.MarkDirty(w)
		}
	})
}
