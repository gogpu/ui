package plugin

import (
	"testing"
)

// TestGlobalFunctions tests the package-level global functions.
// These tests modify global state, so they should be run carefully.
func TestGlobalFunctions(t *testing.T) {
	// Clear global state before and after
	_ = ClearGlobalManager()
	t.Cleanup(func() {
		_ = ClearGlobalManager()
	})

	// Test Register
	p := newMockPlugin("global-test", "1.0.0")
	err := Register(p, PluginInfo{
		Description: "Global test plugin",
	})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Test Has
	if !Has("global-test") {
		t.Error("Has should return true")
	}
	if Has("nonexistent") {
		t.Error("Has should return false for nonexistent")
	}

	// Test Get
	got, ok := Get("global-test")
	if !ok {
		t.Fatal("Get should return true")
	}
	if got != p {
		t.Error("Get returned wrong plugin")
	}

	// Test Count
	if Count() != 1 {
		t.Errorf("Count() = %d, want 1", Count())
	}

	// Test List
	list := List()
	if len(list) != 1 || list[0] != "global-test" {
		t.Errorf("List() = %v, want [global-test]", list)
	}

	// Test Info
	info, ok := Info("global-test")
	if !ok {
		t.Fatal("Info should return true")
	}
	if info.Name != "global-test" {
		t.Errorf("Info.Name = %q, want %q", info.Name, "global-test")
	}
	if info.Description != "Global test plugin" {
		t.Errorf("Info.Description = %q, want %q", info.Description, "Global test plugin")
	}

	// Test AllInfo
	allInfo := AllInfo()
	if len(allInfo) != 1 {
		t.Errorf("AllInfo() returned %d items, want 1", len(allInfo))
	}

	// Test IsInitialized before init
	if IsInitialized() {
		t.Error("Should not be initialized yet")
	}

	// Test LoadOrder before init
	if LoadOrder() != nil {
		t.Error("LoadOrder should be nil before init")
	}

	// Test Initialize
	err = Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Test IsInitialized after init
	if !IsInitialized() {
		t.Error("Should be initialized")
	}

	// Test LoadOrder after init
	order := LoadOrder()
	if len(order) != 1 || order[0] != "global-test" {
		t.Errorf("LoadOrder() = %v, want [global-test]", order)
	}

	// Test Shutdown
	err = Shutdown()
	if err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}

	// Test IsInitialized after shutdown
	if IsInitialized() {
		t.Error("Should not be initialized after shutdown")
	}
}

// TestGlobalMustRegister tests MustRegister.
func TestGlobalMustRegister(t *testing.T) {
	_ = ClearGlobalManager()
	t.Cleanup(func() {
		_ = ClearGlobalManager()
	})

	p := newMockPlugin("must-register-test", "1.0.0")

	// Should not panic
	MustRegister(p)

	if !Has("must-register-test") {
		t.Error("Plugin should be registered")
	}
}

// TestGlobalMustRegisterPanic tests MustRegister panic.
func TestGlobalMustRegisterPanic(t *testing.T) {
	_ = ClearGlobalManager()
	t.Cleanup(func() {
		_ = ClearGlobalManager()
	})

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic")
		}
	}()

	MustRegister(nil)
}

// TestGlobalInitializeWithContext tests InitializeWithContext.
func TestGlobalInitializeWithContext(t *testing.T) {
	_ = ClearGlobalManager()
	t.Cleanup(func() {
		_ = ClearGlobalManager()
	})

	var receivedCtx *PluginContext
	p := newMockPlugin("ctx-test", "1.0.0")
	p.initFunc = func(ctx *PluginContext) error {
		receivedCtx = ctx
		return nil
	}

	_ = Register(p)

	ctx := NewDefaultPluginContext()
	err := InitializeWithContext(ctx)
	if err != nil {
		t.Fatalf("InitializeWithContext failed: %v", err)
	}

	if receivedCtx != ctx {
		t.Error("Plugin should receive the provided context")
	}

	_ = Shutdown()
}

// TestGlobalManager tests GlobalManager.
func TestGlobalManager(t *testing.T) {
	m := GlobalManager()
	if m == nil {
		t.Fatal("GlobalManager returned nil")
	}
	if m != globalManager {
		t.Error("GlobalManager should return the global manager")
	}
}

// TestClearGlobalManager tests ClearGlobalManager.
func TestClearGlobalManager(t *testing.T) {
	_ = ClearGlobalManager()

	_ = Register(newMockPlugin("clear-test", "1.0.0"))
	_ = Initialize()

	if Count() != 1 {
		t.Errorf("Count() = %d, want 1", Count())
	}

	err := ClearGlobalManager()
	if err != nil {
		t.Fatalf("ClearGlobalManager failed: %v", err)
	}

	if Count() != 0 {
		t.Errorf("Count() = %d, want 0", Count())
	}
}
