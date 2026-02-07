package state

import (
	"github.com/coregx/signals"
)

// NewComputed creates a read-only signal whose value is derived from the
// given computation function.
//
// The compute function must be pure: it should read signals and return
// a result without side effects. Dependencies must be passed explicitly
// so that the computed signal is marked dirty when any dependency changes.
//
// The computed value uses lazy evaluation and memoisation: it is only
// recomputed when accessed after a dependency change.
//
// Example:
//
//	firstName := state.NewSignal("John")
//	lastName := state.NewSignal("Doe")
//
//	fullName := state.NewComputed(func() string {
//	    return firstName.Get() + " " + lastName.Get()
//	}, firstName.AsReadonly(), lastName.AsReadonly())
//
//	fmt.Println(fullName.Get()) // "John Doe"
//	firstName.Set("Jane")
//	fmt.Println(fullName.Get()) // "Jane Doe"
func NewComputed[T any](compute func() T, deps ...any) ReadonlySignal[T] {
	return signals.Computed(compute, deps...)
}

// NewComputedWithOptions creates a computed signal with custom options.
//
// Use this when you need custom panic handling for the compute function or
// its subscribers.
func NewComputedWithOptions[T any](compute func() T, opts Options[T], deps ...any) ReadonlySignal[T] {
	return signals.ComputedWithOptions(compute, opts, deps...)
}
