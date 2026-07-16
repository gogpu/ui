// Package linechart safety_test.go tests goroutine safety for concurrent
// data operations on the line chart widget. Regression tests for fix #182
// (linechart data race in PushValue from background goroutines).
//
// The key fix routes invalidation through the scheduler (SchedulerRef)
// when mounted, falling back to direct SetNeedsRedraw when unmounted.
// All data operations are protected by the widget's mutex.
package linechart

import (
	"sync"
	"testing"

	"github.com/gogpu/ui/widget"
)

func TestPushValue_ConcurrentSafe(t *testing.T) {
	// This test verifies goroutine safety — run with -race flag.
	// Multiple goroutines push data simultaneously to the same series.
	w := New(MaxPoints(200))
	w.AddSeries("cpu", widget.ColorRed)

	var wg sync.WaitGroup
	const goroutines = 10
	const pushesPerGoroutine = 100

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < pushesPerGoroutine; j++ {
				w.PushValue("cpu", float64(id*100+j))
			}
		}(i)
	}

	wg.Wait()

	// After all pushes, the series should have data (up to maxPoints).
	w.mu.Lock()
	count := len(w.series[0].Points)
	w.mu.Unlock()

	if count == 0 {
		t.Error("expected data points after concurrent pushes, got 0")
	}
	if count > 200 {
		t.Errorf("point count = %d, exceeds maxPoints 200", count)
	}
}

func TestPushValue_ConcurrentMultipleSeries(t *testing.T) {
	// Concurrent pushes to different series should not race.
	w := New(MaxPoints(50))
	w.AddSeries("cpu", widget.ColorRed)
	w.AddSeries("mem", widget.ColorBlue)
	w.AddSeries("gpu", widget.ColorGreen)

	var wg sync.WaitGroup
	series := []string{"cpu", "mem", "gpu"}

	for _, label := range series {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				w.PushValue(s, float64(i))
			}
		}(label)
	}

	wg.Wait()

	// All series should have data.
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, s := range w.series {
		if len(s.Points) == 0 {
			t.Errorf("series %q has 0 points after concurrent pushes", s.Label)
		}
	}
}

func TestRequestRedraw_UsesSchedulerWhenMounted(t *testing.T) {
	// After Mount, requestRedraw should route through the scheduler
	// instead of calling SetNeedsRedraw directly.
	w := New()
	w.AddSeries("s1", widget.ColorRed)

	sched := &mockScheduler{}
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	w.Mount(ctx)

	// PushValue triggers requestRedraw which should go through scheduler.
	w.PushValue("s1", 42.0)

	if sched.dirtyCount == 0 {
		t.Error("scheduler.MarkDirty should have been called after PushValue (mounted)")
	}
}

func TestRequestRedraw_FallsBackWhenUnmounted(t *testing.T) {
	// Before mount or after unmount, requestRedraw should use SetNeedsRedraw.
	w := New()
	w.AddSeries("s1", widget.ColorRed)

	// Push before mount — should not panic, uses SetNeedsRedraw.
	w.PushValue("s1", 1.0)

	sched := &mockScheduler{}
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	w.Mount(ctx)
	sched.dirtyCount = 0

	w.Unmount()

	// After unmount, scheduler reference is cleared.
	w.PushValue("s1", 2.0)

	if sched.dirtyCount != 0 {
		t.Error("scheduler should NOT be called after Unmount — expected SetNeedsRedraw fallback")
	}
}

func TestAddSeries_ConcurrentSafe(t *testing.T) {
	// Concurrent AddSeries calls should not race.
	w := New()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			label := string(rune('a' + id))
			w.AddSeries(label, widget.ColorRed)
		}(i)
	}

	wg.Wait()

	if w.SeriesCount() != 10 {
		t.Errorf("SeriesCount() = %d, want 10", w.SeriesCount())
	}
}

func TestClearSeries_ConcurrentSafe(t *testing.T) {
	// Concurrent clear and push should not race.
	w := New(MaxPoints(50))
	w.AddSeries("s1", widget.ColorRed)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			w.PushValue("s1", float64(i))
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			w.ClearSeries("s1")
		}
	}()

	wg.Wait()
	// No race or panic = success.
}
