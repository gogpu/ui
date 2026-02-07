package state

import (
	"context"

	"github.com/coregx/signals"
)

// Signal is a writable reactive value that notifies subscribers when changed.
//
// It is a type alias for [signals.Signal] re-exported for convenience so
// that consumers of the state package do not need to import coregx/signals
// directly.
type Signal[T any] = signals.Signal[T]

// ReadonlySignal is a read-only view of a signal that supports Get and Subscribe
// but not Set or Update.
type ReadonlySignal[T any] = signals.ReadonlySignal[T]

// Unsubscribe is a function returned by Subscribe that removes the subscription.
// It must be called to prevent memory leaks.
type Unsubscribe = signals.Unsubscribe

// EqualFunc is a custom equality function used to determine whether a signal's
// value has changed. When provided, Set only notifies subscribers if the new
// value is not equal to the old value according to this function.
type EqualFunc[T any] = signals.EqualFunc[T]

// Options configures the behavior of a signal.
//
// Equal — optional custom equality function; if nil every Set notifies.
// OnPanic — optional panic handler for subscriber callbacks.
type Options[T any] = signals.Options[T]

// NewSignal creates a new writable signal with the given initial value.
//
// The signal uses default behavior: no equality checks and default panic
// handling (log and continue).
//
// Example:
//
//	count := state.NewSignal(0)
//	count.Set(5)
//	fmt.Println(count.Get()) // 5
func NewSignal[T any](initial T) Signal[T] {
	return signals.New(initial)
}

// NewSignalWithOptions creates a new writable signal with custom options.
//
// Use this when you need custom equality checks or custom panic handling.
//
// Example:
//
//	count := state.NewSignalWithOptions(0, state.Options[int]{
//	    Equal: func(a, b int) bool { return a == b },
//	})
func NewSignalWithOptions[T any](initial T, opts Options[T]) Signal[T] {
	return signals.NewWithOptions(initial, opts)
}

// Subscribe registers a callback on a readable signal that is automatically
// canceled when the context is done. Returns an Unsubscribe function for
// manual cleanup.
//
// This is a convenience wrapper so callers do not need to import coregx/signals.
func Subscribe[T any](sig ReadonlySignal[T], ctx context.Context, fn func(T)) Unsubscribe {
	return sig.Subscribe(ctx, fn)
}

// SubscribeForever registers a callback on a readable signal that is never
// automatically canceled. The caller must call the returned Unsubscribe to
// prevent memory leaks.
func SubscribeForever[T any](sig ReadonlySignal[T], fn func(T)) Unsubscribe {
	return sig.SubscribeForever(fn)
}
