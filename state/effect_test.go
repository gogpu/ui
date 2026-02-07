package state_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/gogpu/ui/state"
)

func TestNewEffect(t *testing.T) {
	var ran atomic.Int32
	src := state.NewSignal(0)

	eff := state.NewEffect(func() {
		ran.Add(1)
	}, src.AsReadonly())
	defer eff.Stop()

	// Effect runs immediately on creation.
	if got := ran.Load(); got != 1 {
		t.Errorf("effect ran %d times initially, want 1", got)
	}

	src.Set(5)
	if got := ran.Load(); got != 2 {
		t.Errorf("after Set, effect ran %d times, want 2", got)
	}
}

func TestNewEffectStop(t *testing.T) {
	var ran atomic.Int32
	src := state.NewSignal(0)

	eff := state.NewEffect(func() {
		ran.Add(1)
	}, src.AsReadonly())

	eff.Stop()
	src.Set(10)

	// Give time for potential async notification.
	time.Sleep(20 * time.Millisecond)

	// Only the initial run should have happened.
	if got := ran.Load(); got != 1 {
		t.Errorf("after Stop, effect ran %d times, want 1", got)
	}
}

func TestNewEffectWithCleanup(t *testing.T) {
	var cleanupCalled atomic.Int32
	src := state.NewSignal(0)

	eff := state.NewEffectWithCleanup(func() func() {
		return func() {
			cleanupCalled.Add(1)
		}
	}, src.AsReadonly())

	// Trigger re-run: old cleanup should be called.
	src.Set(1)
	if got := cleanupCalled.Load(); got != 1 {
		t.Errorf("cleanup called %d times after re-run, want 1", got)
	}

	// Stop triggers final cleanup.
	eff.Stop()
	if got := cleanupCalled.Load(); got != 2 {
		t.Errorf("cleanup called %d times after Stop, want 2", got)
	}
}

func TestNewEffectMultipleDeps(t *testing.T) {
	var ran atomic.Int32
	a := state.NewSignal(0)
	b := state.NewSignal("x")

	eff := state.NewEffect(func() {
		_ = a.Get()
		_ = b.Get()
		ran.Add(1)
	}, a.AsReadonly(), b.AsReadonly())
	defer eff.Stop()

	// Initial run.
	if got := ran.Load(); got != 1 {
		t.Fatalf("initial runs = %d, want 1", got)
	}

	a.Set(1)
	b.Set("y")

	if got := ran.Load(); got < 3 {
		t.Errorf("after two dep changes, runs = %d, want >= 3", got)
	}
}

func TestNewEffectStopIdempotent(t *testing.T) {
	src := state.NewSignal(0)
	eff := state.NewEffect(func() {
		_ = src.Get()
	}, src.AsReadonly())

	eff.Stop()
	eff.Stop() // second call should be safe
	eff.Stop() // third call should be safe
}
