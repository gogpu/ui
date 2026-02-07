package state

import (
	"sync"

	"github.com/gogpu/ui/widget"
)

// Binding connects a reactive signal to a widget's invalidation lifecycle.
//
// When the bound signal's value changes the widget's context is invalidated,
// which marks the widget for re-render. A Binding must be cleaned up via
// [Binding.Unbind] when the widget is removed from the tree; otherwise the
// subscription leaks.
//
// Create a Binding with [Bind].
type Binding struct {
	mu      sync.Mutex
	cleanup Unsubscribe
	active  bool
}

// Unbind stops the binding so that future signal changes no longer
// invalidate the widget. Safe to call multiple times; subsequent calls
// are no-ops.
func (b *Binding) Unbind() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.active {
		return
	}
	b.active = false
	if b.cleanup != nil {
		b.cleanup()
		b.cleanup = nil
	}
}

// IsActive reports whether the binding is still active (not yet unbound).
func (b *Binding) IsActive() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.active
}

// Bind creates a [Binding] that invalidates ctx whenever sig changes.
//
// The type parameter T must match the signal's value type. The binding
// subscribes to the signal using SubscribeForever; the caller must call
// [Binding.Unbind] to release the subscription.
//
// Example:
//
//	counter := state.NewSignal(0)
//	binding := state.Bind(counter, ctx)
//	defer binding.Unbind()
//
//	counter.Set(1) // ctx.Invalidate() is called automatically
func Bind[T any](sig ReadonlySignal[T], ctx widget.Context) *Binding {
	b := &Binding{active: true}
	unsub := sig.SubscribeForever(func(_ T) {
		ctx.Invalidate()
	})
	b.cleanup = unsub
	return b
}

// BindToScheduler creates a [Binding] that marks w as dirty in sched
// whenever sig changes.
//
// Use this instead of [Bind] when you want fine-grained control over
// render batching. The scheduler collects dirty widgets and processes
// them in a single flush.
//
// Example:
//
//	counter := state.NewSignal(0)
//	sched := state.NewScheduler(flushFn)
//	binding := state.BindToScheduler(counter, myWidget, sched)
//	defer binding.Unbind()
//
//	counter.Set(1) // sched.MarkDirty(myWidget) is called
func BindToScheduler[T any](sig ReadonlySignal[T], w widget.Widget, sched *Scheduler) *Binding {
	b := &Binding{active: true}
	unsub := sig.SubscribeForever(func(_ T) {
		sched.MarkDirty(w)
	})
	b.cleanup = unsub
	return b
}
