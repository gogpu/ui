// Package state provides reactive state management for the gogpu/ui widget tree.
//
// It wraps [github.com/coregx/signals] with UI-specific helpers for
// binding reactive values to widget invalidation and for batching
// multiple state changes into a single render pass.
//
// # Core Concepts
//
// A Signal holds a single value and notifies subscribers when that value changes.
// A Computed signal derives its value from other signals and recomputes lazily.
// An Effect runs a side-effect whenever its dependencies change.
// A Binding connects a signal to a widget so the widget is automatically
// invalidated (marked for re-render) when the signal changes.
// A Scheduler collects dirty widgets and flushes them in one batch.
//
// # Quick Start
//
//	// 1. Create a signal for some piece of UI state.
//	counter := state.NewSignal(0)
//
//	// 2. Derive a display string from it.
//	label := state.NewComputed(func() string {
//	    return fmt.Sprintf("Count: %d", counter.Get())
//	}, counter.AsReadonly())
//
//	// 3. Bind the signal to a widget so it re-renders on change.
//	binding := state.Bind(counter, ctx)
//	defer binding.Unbind()
//
//	// 4. Batch multiple updates into one render.
//	sched := state.NewScheduler(func(dirty []widget.Widget) {
//	    for _, w := range dirty {
//	        renderWidget(w)
//	    }
//	})
//	sched.Batch(func() {
//	    counter.Set(1)
//	    counter.Set(2)
//	    counter.Set(3)
//	})
//	sched.Flush() // processes all dirty widgets once
//
// # Thread Safety
//
// All types in this package are safe for concurrent use.
// Scheduler and Binding protect their internal state with mutexes.
// The underlying signals library is also fully thread-safe.
//
// # Memory Safety
//
// Every Binding must be cleaned up by calling Unbind when the widget
// is removed from the tree. Failing to do so leaks the subscription.
// Effects returned by NewEffect must be stopped via the returned
// EffectRef.Stop method.
package state
