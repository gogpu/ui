package layout

import (
	"sync"
	"testing"

	"github.com/gogpu/ui/geometry"
)

// mockAlgorithm is a simple algorithm for testing.
type mockAlgorithm struct {
	name string
}

func (m *mockAlgorithm) Name() string { return m.name }

func (m *mockAlgorithm) Compute(_ LayoutTree, _ NodeID, available geometry.Size) Result {
	return Result{Size: available}
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()

	algo := &mockAlgorithm{name: "test"}
	r.Register(algo)

	if !r.Has("test") {
		t.Error("algorithm should be registered")
	}

	got, ok := r.Get("test")
	if !ok {
		t.Fatal("Get should return true for registered algorithm")
	}
	if got.Name() != "test" {
		t.Errorf("Name() = %q, want %q", got.Name(), "test")
	}
}

func TestRegistry_RegisterWithName(t *testing.T) {
	r := NewRegistry()

	algo := &mockAlgorithm{name: "original"}
	r.RegisterWithName("custom", algo)

	if r.Has("original") {
		t.Error("should not be registered under original name")
	}

	if !r.Has("custom") {
		t.Error("should be registered under custom name")
	}
}

func TestRegistry_Get_NotFound(t *testing.T) {
	r := NewRegistry()

	_, ok := r.Get("nonexistent")
	if ok {
		t.Error("Get should return false for unregistered algorithm")
	}
}

func TestRegistry_MustGet_Panic(t *testing.T) {
	r := NewRegistry()

	defer func() {
		if recover() == nil {
			t.Error("MustGet should panic for unregistered algorithm")
		}
	}()

	_ = r.MustGet("nonexistent")
}

func TestRegistry_MustGet_Success(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockAlgorithm{name: "test"})

	// Should not panic
	algo := r.MustGet("test")
	if algo.Name() != "test" {
		t.Errorf("MustGet returned wrong algorithm")
	}
}

func TestRegistry_Unregister(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockAlgorithm{name: "test"})

	if !r.Unregister("test") {
		t.Error("Unregister should return true for registered algorithm")
	}

	if r.Has("test") {
		t.Error("algorithm should no longer be registered")
	}

	if r.Unregister("test") {
		t.Error("Unregister should return false for unregistered algorithm")
	}
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockAlgorithm{name: "charlie"})
	r.Register(&mockAlgorithm{name: "alpha"})
	r.Register(&mockAlgorithm{name: "bravo"})

	list := r.List()

	if len(list) != 3 {
		t.Fatalf("List() length = %d, want 3", len(list))
	}

	// Should be sorted
	expected := []string{"alpha", "bravo", "charlie"}
	for i, name := range expected {
		if list[i] != name {
			t.Errorf("List()[%d] = %q, want %q", i, list[i], name)
		}
	}
}

func TestRegistry_Count(t *testing.T) {
	r := NewRegistry()

	if r.Count() != 0 {
		t.Errorf("Count() = %d, want 0", r.Count())
	}

	r.Register(&mockAlgorithm{name: "one"})
	r.Register(&mockAlgorithm{name: "two"})

	if r.Count() != 2 {
		t.Errorf("Count() = %d, want 2", r.Count())
	}
}

func TestRegistry_Clear(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockAlgorithm{name: "one"})
	r.Register(&mockAlgorithm{name: "two"})

	r.Clear()

	if r.Count() != 0 {
		t.Errorf("after Clear: Count() = %d, want 0", r.Count())
	}
}

func TestRegistry_Clone(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockAlgorithm{name: "one"})
	r.Register(&mockAlgorithm{name: "two"})

	clone := r.Clone()

	// Clone should have same algorithms
	if clone.Count() != 2 {
		t.Errorf("clone Count() = %d, want 2", clone.Count())
	}

	// Modifications to clone should not affect original
	clone.Register(&mockAlgorithm{name: "three"})
	if r.Count() != 2 {
		t.Error("modifying clone affected original")
	}
}

func TestRegistry_Replace(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockAlgorithm{name: "test"})

	// Register with same name should replace
	newAlgo := &mockAlgorithm{name: "test"}
	r.Register(newAlgo)

	if r.Count() != 1 {
		t.Errorf("Count() = %d, want 1 after replace", r.Count())
	}
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	r := NewRegistry()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := string(rune('a' + (i % 26)))
			r.Register(&mockAlgorithm{name: name})
			_, _ = r.Get(name)
			_ = r.List()
			_ = r.Has(name)
		}(i)
	}
	wg.Wait()

	// Should not panic and registry should be consistent
	if r.Count() < 1 {
		t.Error("registry should have at least one algorithm")
	}
}

func TestGlobalRegistry(t *testing.T) {
	// Save original state
	originalCount := Count()

	// Test global functions
	Register(&mockAlgorithm{name: "global-test"})
	defer Unregister("global-test")

	if !Has("global-test") {
		t.Error("global Has should return true")
	}

	algo, ok := Get("global-test")
	if !ok || algo.Name() != "global-test" {
		t.Error("global Get should return registered algorithm")
	}

	// Count should have increased
	if Count() != originalCount+1 {
		t.Errorf("Count() = %d, want %d", Count(), originalCount+1)
	}

	// List should include new algorithm
	list := List()
	found := false
	for _, name := range list {
		if name == "global-test" {
			found = true
			break
		}
	}
	if !found {
		t.Error("List should include global-test")
	}

	// GlobalRegistry should return the same instance
	if GlobalRegistry() == nil {
		t.Error("GlobalRegistry should not return nil")
	}
}

func TestLayoutFunc(t *testing.T) {
	called := false
	fn := LayoutFunc{
		NameValue: "test-func",
		ComputeFunc: func(_ LayoutTree, _ NodeID, available geometry.Size) Result {
			called = true
			return Result{Size: available}
		},
	}

	if fn.Name() != "test-func" {
		t.Errorf("Name() = %q, want %q", fn.Name(), "test-func")
	}

	result := fn.Compute(nil, 0, geometry.Size{Width: 100, Height: 100})
	if !called {
		t.Error("ComputeFunc should have been called")
	}
	if result.Size.Width != 100 || result.Size.Height != 100 {
		t.Errorf("Result.Size = %v, want {100, 100}", result.Size)
	}
}

func TestLayoutFunc_NilCompute(t *testing.T) {
	fn := LayoutFunc{NameValue: "nil-func"}

	result := fn.Compute(nil, 0, geometry.Size{Width: 100, Height: 100})
	if !result.IsZero() {
		t.Error("Compute with nil function should return zero Result")
	}
}
