package state

import (
	"github.com/coregx/signals"
)

// EffectRef represents a running side effect that can be stopped.
//
// Effects run immediately upon creation and re-run whenever any of their
// dependencies change. Call Stop to clean up the effect and unsubscribe
// from all dependencies.
type EffectRef = signals.EffectRef

// NewEffect creates a side effect that runs immediately and re-runs whenever
// any dependency changes.
//
// Dependencies must be passed explicitly. The effect function should not
// return anything; for effects that need cleanup use [NewEffectWithCleanup].
//
// Example:
//
//	count := state.NewSignal(0)
//
//	eff := state.NewEffect(func() {
//	    fmt.Println("count is", count.Get())
//	}, count.AsReadonly())
//	defer eff.Stop()
//
//	count.Set(5) // prints "count is 5"
func NewEffect(fn func(), deps ...any) EffectRef {
	return signals.Effect(fn, deps...)
}

// NewEffectWithCleanup creates a side effect whose function returns a cleanup
// callback.
//
// The cleanup callback is called:
//   - Before the next execution of the effect
//   - When Stop is called
//
// This is useful for canceling timers, closing connections, or removing
// event listeners established by the effect.
//
// Example:
//
//	interval := state.NewSignal(time.Second)
//
//	eff := state.NewEffectWithCleanup(func() func() {
//	    ticker := time.NewTicker(interval.Get())
//	    go func() {
//	        for range ticker.C {
//	            fmt.Println("tick")
//	        }
//	    }()
//	    return func() { ticker.Stop() }
//	}, interval.AsReadonly())
//	defer eff.Stop()
func NewEffectWithCleanup(fn func() func(), deps ...any) EffectRef {
	return signals.EffectWithCleanup(fn, deps...)
}
