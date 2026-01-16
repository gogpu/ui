package plugin

import (
	"errors"
	"sync"
	"testing"
)

// TestPluginManagerNew tests creating a new plugin manager.
func TestPluginManagerNew(t *testing.T) {
	m := NewPluginManager()

	if m == nil {
		t.Fatal("NewPluginManager returned nil")
	}
	if m.plugins == nil {
		t.Error("plugins map is nil")
	}
	if m.info == nil {
		t.Error("info map is nil")
	}
	if m.initialized {
		t.Error("should not be initialized")
	}
}

// TestPluginManagerRegister tests plugin registration.
func TestPluginManagerRegister(t *testing.T) {
	m := NewPluginManager()
	p := newMockPlugin("test", "1.0.0")

	err := m.Register(p)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if !m.Has("test") {
		t.Error("Plugin should be registered")
	}
}

// TestPluginManagerRegisterWithInfo tests registration with PluginInfo.
func TestPluginManagerRegisterWithInfo(t *testing.T) {
	m := NewPluginManager()
	p := newMockPlugin("test", "1.0.0")
	info := PluginInfo{
		Description: "Test plugin",
		Author:      "Test Author",
		License:     "MIT",
		Homepage:    "https://example.com",
	}

	err := m.Register(p, info)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	gotInfo, ok := m.Info("test")
	if !ok {
		t.Fatal("Info not found")
	}
	if gotInfo.Name != "test" {
		t.Errorf("Name = %q, want %q", gotInfo.Name, "test")
	}
	if gotInfo.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", gotInfo.Version, "1.0.0")
	}
	if gotInfo.Description != "Test plugin" {
		t.Errorf("Description = %q, want %q", gotInfo.Description, "Test plugin")
	}
}

// TestPluginManagerRegisterNil tests registering nil plugin.
func TestPluginManagerRegisterNil(t *testing.T) {
	m := NewPluginManager()

	err := m.Register(nil)
	if !errors.Is(err, ErrNilPlugin) {
		t.Errorf("Expected ErrNilPlugin, got %v", err)
	}
}

// TestPluginManagerRegisterEmptyName tests registering plugin with empty name.
func TestPluginManagerRegisterEmptyName(t *testing.T) {
	m := NewPluginManager()
	p := newMockPlugin("", "1.0.0")

	err := m.Register(p)
	if !errors.Is(err, ErrEmptyName) {
		t.Errorf("Expected ErrEmptyName, got %v", err)
	}
}

// TestPluginManagerRegisterDuplicate tests registering duplicate plugin.
func TestPluginManagerRegisterDuplicate(t *testing.T) {
	m := NewPluginManager()
	p1 := newMockPlugin("test", "1.0.0")
	p2 := newMockPlugin("test", "2.0.0")

	_ = m.Register(p1)
	err := m.Register(p2)

	if !errors.Is(err, ErrPluginExists) {
		t.Errorf("Expected ErrPluginExists, got %v", err)
	}
}

// TestPluginManagerMustRegister tests MustRegister.
func TestPluginManagerMustRegister(t *testing.T) {
	m := NewPluginManager()
	p := newMockPlugin("test", "1.0.0")

	// Should not panic
	m.MustRegister(p)

	if !m.Has("test") {
		t.Error("Plugin should be registered")
	}
}

// TestPluginManagerMustRegisterPanic tests MustRegister panic.
func TestPluginManagerMustRegisterPanic(t *testing.T) {
	m := NewPluginManager()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic")
		}
	}()

	m.MustRegister(nil)
}

// TestPluginManagerInitialize tests plugin initialization.
func TestPluginManagerInitialize(t *testing.T) {
	m := NewPluginManager()
	p := newMockPlugin("test", "1.0.0")
	_ = m.Register(p)

	err := m.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if !m.IsInitialized() {
		t.Error("Should be initialized")
	}
	if !p.initCalled {
		t.Error("Init should have been called")
	}
}

// TestPluginManagerInitializeDouble tests double initialization.
func TestPluginManagerInitializeDouble(t *testing.T) {
	m := NewPluginManager()
	p := newMockPlugin("test", "1.0.0")
	_ = m.Register(p)
	_ = m.Initialize()

	err := m.Initialize()
	if !errors.Is(err, ErrAlreadyInitialized) {
		t.Errorf("Expected ErrAlreadyInitialized, got %v", err)
	}
}

// TestPluginManagerInitializeError tests initialization error.
func TestPluginManagerInitializeError(t *testing.T) {
	m := NewPluginManager()
	p := newMockPlugin("test", "1.0.0")
	p.initFunc = func(_ *PluginContext) error {
		return errors.New("init failed")
	}
	_ = m.Register(p)

	err := m.Initialize()
	if err == nil {
		t.Error("Expected error")
	}
}

// TestPluginManagerShutdown tests plugin shutdown.
func TestPluginManagerShutdown(t *testing.T) {
	m := NewPluginManager()
	p := newMockPlugin("test", "1.0.0")
	_ = m.Register(p)
	_ = m.Initialize()

	err := m.Shutdown()
	if err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}

	if m.IsInitialized() {
		t.Error("Should not be initialized")
	}
	if !p.shutdownCalled {
		t.Error("Shutdown should have been called")
	}
}

// TestPluginManagerShutdownNotInitialized tests shutdown without init.
func TestPluginManagerShutdownNotInitialized(t *testing.T) {
	m := NewPluginManager()

	err := m.Shutdown()
	if !errors.Is(err, ErrNotInitialized) {
		t.Errorf("Expected ErrNotInitialized, got %v", err)
	}
}

// TestPluginManagerShutdownError tests shutdown error.
func TestPluginManagerShutdownError(t *testing.T) {
	m := NewPluginManager()
	p := newMockPlugin("test", "1.0.0")
	p.shutdownFunc = func() error {
		return errors.New("shutdown failed")
	}
	_ = m.Register(p)
	_ = m.Initialize()

	err := m.Shutdown()
	if err == nil {
		t.Error("Expected error")
	}
}

// TestPluginManagerList tests listing plugins.
func TestPluginManagerList(t *testing.T) {
	m := NewPluginManager()
	_ = m.Register(newMockPlugin("charlie", "1.0.0"))
	_ = m.Register(newMockPlugin("alpha", "1.0.0"))
	_ = m.Register(newMockPlugin("bravo", "1.0.0"))

	list := m.List()
	if len(list) != 3 {
		t.Fatalf("Expected 3 plugins, got %d", len(list))
	}

	// Should be sorted
	expected := []string{"alpha", "bravo", "charlie"}
	for i, name := range expected {
		if list[i] != name {
			t.Errorf("list[%d] = %q, want %q", i, list[i], name)
		}
	}
}

// TestPluginManagerCount tests counting plugins.
func TestPluginManagerCount(t *testing.T) {
	m := NewPluginManager()

	if m.Count() != 0 {
		t.Errorf("Count() = %d, want 0", m.Count())
	}

	_ = m.Register(newMockPlugin("test1", "1.0.0"))
	_ = m.Register(newMockPlugin("test2", "1.0.0"))

	if m.Count() != 2 {
		t.Errorf("Count() = %d, want 2", m.Count())
	}
}

// TestPluginManagerGet tests getting a plugin.
func TestPluginManagerGet(t *testing.T) {
	m := NewPluginManager()
	p := newMockPlugin("test", "1.0.0")
	_ = m.Register(p)

	got, ok := m.Get("test")
	if !ok {
		t.Fatal("Plugin not found")
	}
	if got != p {
		t.Error("Got wrong plugin")
	}

	_, ok = m.Get("nonexistent")
	if ok {
		t.Error("Should not find nonexistent plugin")
	}
}

// TestPluginManagerLoadOrder tests initialization order.
func TestPluginManagerLoadOrder(t *testing.T) {
	m := NewPluginManager()

	// A depends on B, B depends on C
	a := newMockPlugin("A", "1.0.0")
	a.dependencies = []Dependency{{Name: "B"}}
	b := newMockPlugin("B", "1.0.0")
	b.dependencies = []Dependency{{Name: "C"}}
	c := newMockPlugin("C", "1.0.0")

	_ = m.Register(a)
	_ = m.Register(b)
	_ = m.Register(c)
	_ = m.Initialize()

	order := m.LoadOrder()
	if len(order) != 3 {
		t.Fatalf("Expected 3 items, got %d", len(order))
	}

	// Find positions
	positions := make(map[string]int)
	for i, name := range order {
		positions[name] = i
	}

	if positions["C"] > positions["B"] {
		t.Error("C should come before B")
	}
	if positions["B"] > positions["A"] {
		t.Error("B should come before A")
	}
}

// TestPluginManagerShutdownOrder tests shutdown order (reverse of init).
func TestPluginManagerShutdownOrder(t *testing.T) {
	m := NewPluginManager()

	var shutdownOrder []string
	var mu sync.Mutex

	a := newMockPlugin("A", "1.0.0")
	a.dependencies = []Dependency{{Name: "B"}}
	a.shutdownFunc = func() error {
		mu.Lock()
		shutdownOrder = append(shutdownOrder, "A")
		mu.Unlock()
		return nil
	}

	b := newMockPlugin("B", "1.0.0")
	b.shutdownFunc = func() error {
		mu.Lock()
		shutdownOrder = append(shutdownOrder, "B")
		mu.Unlock()
		return nil
	}

	_ = m.Register(a)
	_ = m.Register(b)
	_ = m.Initialize()
	_ = m.Shutdown()

	// A should shut down before B (reverse of init order)
	if len(shutdownOrder) != 2 {
		t.Fatalf("Expected 2 shutdowns, got %d", len(shutdownOrder))
	}
	if shutdownOrder[0] != "A" {
		t.Errorf("Expected A first, got %s", shutdownOrder[0])
	}
	if shutdownOrder[1] != "B" {
		t.Errorf("Expected B second, got %s", shutdownOrder[1])
	}
}

// TestPluginManagerDependencyNotFound tests missing dependency.
func TestPluginManagerDependencyNotFound(t *testing.T) {
	m := NewPluginManager()

	a := newMockPlugin("A", "1.0.0")
	a.dependencies = []Dependency{{Name: "B"}}

	_ = m.Register(a)

	err := m.Initialize()
	if !errors.Is(err, ErrDependencyNotFound) {
		t.Errorf("Expected ErrDependencyNotFound, got %v", err)
	}
}

// TestPluginManagerVersionMismatch tests version constraint failure.
func TestPluginManagerVersionMismatch(t *testing.T) {
	m := NewPluginManager()

	a := newMockPlugin("A", "1.0.0")
	a.dependencies = []Dependency{{Name: "B", Version: ">=2.0.0"}}

	b := newMockPlugin("B", "1.0.0")

	_ = m.Register(a)
	_ = m.Register(b)

	err := m.Initialize()
	if !errors.Is(err, ErrVersionMismatch) {
		t.Errorf("Expected ErrVersionMismatch, got %v", err)
	}
}

// TestPluginManagerCircularDependency tests circular dependency detection.
func TestPluginManagerCircularDependency(t *testing.T) {
	m := NewPluginManager()

	a := newMockPlugin("A", "1.0.0")
	a.dependencies = []Dependency{{Name: "B"}}

	b := newMockPlugin("B", "1.0.0")
	b.dependencies = []Dependency{{Name: "A"}}

	_ = m.Register(a)
	_ = m.Register(b)

	err := m.Initialize()
	if !errors.Is(err, ErrCircularDependency) {
		t.Errorf("Expected ErrCircularDependency, got %v", err)
	}
}

// TestPluginManagerClear tests clearing the manager.
func TestPluginManagerClear(t *testing.T) {
	m := NewPluginManager()
	_ = m.Register(newMockPlugin("test", "1.0.0"))
	_ = m.Initialize()

	err := m.Clear()
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	if m.Count() != 0 {
		t.Errorf("Count() = %d, want 0", m.Count())
	}
	if m.IsInitialized() {
		t.Error("Should not be initialized")
	}
}

// TestPluginManagerAllInfo tests getting all plugin info.
func TestPluginManagerAllInfo(t *testing.T) {
	m := NewPluginManager()
	_ = m.Register(newMockPlugin("charlie", "3.0.0"))
	_ = m.Register(newMockPlugin("alpha", "1.0.0"))
	_ = m.Register(newMockPlugin("bravo", "2.0.0"))

	infos := m.AllInfo()
	if len(infos) != 3 {
		t.Fatalf("Expected 3 infos, got %d", len(infos))
	}

	// Should be sorted by name
	expected := []string{"alpha", "bravo", "charlie"}
	for i, name := range expected {
		if infos[i].Name != name {
			t.Errorf("infos[%d].Name = %q, want %q", i, infos[i].Name, name)
		}
	}
}

// TestPluginManagerConcurrency tests thread safety.
func TestPluginManagerConcurrency(t *testing.T) {
	m := NewPluginManager()
	const n = 100

	var wg sync.WaitGroup

	// Concurrent registrations
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p := newMockPlugin("plugin-concurrent", "1.0.0")
			_ = m.Register(p) // May fail due to duplicate, that's OK
		}()
	}

	// Concurrent reads
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = m.List()
			_ = m.Count()
			_, _ = m.Get("plugin-concurrent")
			_, _ = m.Info("plugin-concurrent")
		}()
	}

	wg.Wait()

	// Should have exactly 1 plugin registered
	if m.Count() != 1 {
		t.Errorf("Count() = %d, want 1", m.Count())
	}
}

// TestPluginManagerLoadOrderNotInitialized tests LoadOrder when not initialized.
func TestPluginManagerLoadOrderNotInitialized(t *testing.T) {
	m := NewPluginManager()

	order := m.LoadOrder()
	if order != nil {
		t.Errorf("Expected nil, got %v", order)
	}
}

// TestPluginManagerRegisterAfterInit tests registration after initialization.
func TestPluginManagerRegisterAfterInit(t *testing.T) {
	m := NewPluginManager()
	_ = m.Register(newMockPlugin("test1", "1.0.0"))
	_ = m.Initialize()

	err := m.Register(newMockPlugin("test2", "1.0.0"))
	if !errors.Is(err, ErrAlreadyInitialized) {
		t.Errorf("Expected ErrAlreadyInitialized, got %v", err)
	}
}

// TestPluginManagerInitializeWithContext tests InitializeWithContext.
func TestPluginManagerInitializeWithContext(t *testing.T) {
	m := NewPluginManager()

	var receivedCtx *PluginContext
	p := newMockPlugin("test", "1.0.0")
	p.initFunc = func(ctx *PluginContext) error {
		receivedCtx = ctx
		return nil
	}

	_ = m.Register(p)

	ctx := NewDefaultPluginContext()
	err := m.InitializeWithContext(ctx)
	if err != nil {
		t.Fatalf("InitializeWithContext failed: %v", err)
	}

	if receivedCtx != ctx {
		t.Error("Plugin should receive the provided context")
	}
}
