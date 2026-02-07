package state_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// mockContext is a minimal Context implementation for testing bindings.
type mockContext struct {
	widget.Context
	invalidated atomic.Int32
}

func (m *mockContext) Invalidate() {
	m.invalidated.Add(1)
}

// mockWidget is a minimal Widget implementation for testing scheduler bindings.
// It implements all methods required by the widget.Widget interface.
type mockWidget struct {
	widget.WidgetBase
	name string
}

func (m *mockWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(0, 0))
}

func (m *mockWidget) Draw(_ widget.Context, _ widget.Canvas) {}

func (m *mockWidget) Event(_ widget.Context, _ event.Event) bool { return false }

func TestBindInvalidatesOnChange(t *testing.T) {
	sig := state.NewSignal(0)
	ctx := &mockContext{}

	binding := state.Bind(sig.AsReadonly(), ctx)
	defer binding.Unbind()

	sig.Set(1)
	sig.Set(2)
	sig.Set(3)

	if got := ctx.invalidated.Load(); got != 3 {
		t.Errorf("invalidated count = %d, want 3", got)
	}
}

func TestBindUnbindStopsInvalidation(t *testing.T) {
	sig := state.NewSignal(0)
	ctx := &mockContext{}

	binding := state.Bind(sig.AsReadonly(), ctx)
	sig.Set(1)
	binding.Unbind()
	sig.Set(2)
	sig.Set(3)

	if got := ctx.invalidated.Load(); got != 1 {
		t.Errorf("invalidated count = %d, want 1 (after unbind)", got)
	}
}

func TestBindUnbindIdempotent(t *testing.T) {
	sig := state.NewSignal(0)
	ctx := &mockContext{}

	binding := state.Bind(sig.AsReadonly(), ctx)
	binding.Unbind()
	binding.Unbind() // second call should be safe
	binding.Unbind() // third call should be safe

	// No panic is the success criterion.
}

func TestBindIsActive(t *testing.T) {
	sig := state.NewSignal(0)
	ctx := &mockContext{}

	binding := state.Bind(sig.AsReadonly(), ctx)
	if !binding.IsActive() {
		t.Error("new binding should be active")
	}

	binding.Unbind()
	if binding.IsActive() {
		t.Error("unbound binding should not be active")
	}
}

func TestBindToScheduler(t *testing.T) {
	sig := state.NewSignal(0)
	w := &mockWidget{name: "test"}
	var flushed []widget.Widget
	sched := state.NewScheduler(func(dirty []widget.Widget) {
		flushed = append(flushed, dirty...)
	})

	binding := state.BindToScheduler(sig.AsReadonly(), w, sched)
	defer binding.Unbind()

	sig.Set(1)
	sig.Set(2)

	// Two changes, but widget should be deduplicated.
	if got := sched.PendingCount(); got != 1 {
		t.Errorf("pending count = %d, want 1 (dedup)", got)
	}

	sched.Flush()
	if len(flushed) != 1 {
		t.Fatalf("flushed count = %d, want 1", len(flushed))
	}
}

func TestBindToSchedulerUnbind(t *testing.T) {
	sig := state.NewSignal(0)
	w := &mockWidget{name: "test"}
	sched := state.NewScheduler(func(_ []widget.Widget) {})

	binding := state.BindToScheduler(sig.AsReadonly(), w, sched)
	sig.Set(1)
	binding.Unbind()
	// Clear pending from previous set.
	sched.Flush()

	sig.Set(2)
	if got := sched.PendingCount(); got != 0 {
		t.Errorf("after unbind, pending count = %d, want 0", got)
	}
}

func TestBindConcurrent(t *testing.T) {
	sig := state.NewSignal(0)
	ctx := &mockContext{}

	binding := state.Bind(sig.AsReadonly(), ctx)
	defer binding.Unbind()

	var wg sync.WaitGroup
	const goroutines = 20
	const iterations = 50

	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			for j := range iterations {
				sig.Set(j)
			}
		}()
	}
	wg.Wait()

	if got := ctx.invalidated.Load(); got == 0 {
		t.Error("expected at least one invalidation from concurrent sets")
	}
}

func TestBindStringSignal(t *testing.T) {
	sig := state.NewSignal("initial")
	ctx := &mockContext{}

	binding := state.Bind(sig.AsReadonly(), ctx)
	defer binding.Unbind()

	sig.Set("updated")
	if got := ctx.invalidated.Load(); got != 1 {
		t.Errorf("invalidated count = %d, want 1", got)
	}
}
