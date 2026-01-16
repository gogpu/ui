package plugin

import (
	"errors"
	"testing"
)

// mockPlugin is a test implementation of the Plugin interface.
type mockPlugin struct {
	name           string
	version        string
	dependencies   []Dependency
	initFunc       func(*PluginContext) error
	shutdownFunc   func() error
	initCalled     bool
	shutdownCalled bool
}

func newMockPlugin(name, version string) *mockPlugin {
	return &mockPlugin{
		name:    name,
		version: version,
	}
}

func (p *mockPlugin) Name() string    { return p.name }
func (p *mockPlugin) Version() string { return p.version }

func (p *mockPlugin) Dependencies() []Dependency {
	return p.dependencies
}

func (p *mockPlugin) Init(ctx *PluginContext) error {
	p.initCalled = true
	if p.initFunc != nil {
		return p.initFunc(ctx)
	}
	return nil
}

func (p *mockPlugin) Shutdown() error {
	p.shutdownCalled = true
	if p.shutdownFunc != nil {
		return p.shutdownFunc()
	}
	return nil
}

// TestPluginInterface verifies that mockPlugin implements Plugin.
func TestPluginInterface(t *testing.T) {
	var _ Plugin = (*mockPlugin)(nil)
}

// TestDependency tests the Dependency struct.
func TestDependency(t *testing.T) {
	dep := Dependency{
		Name:    "base-plugin",
		Version: ">=1.0.0",
	}

	if dep.Name != "base-plugin" {
		t.Errorf("Name = %q, want %q", dep.Name, "base-plugin")
	}
	if dep.Version != ">=1.0.0" {
		t.Errorf("Version = %q, want %q", dep.Version, ">=1.0.0")
	}
}

// TestPluginInfo tests the PluginInfo struct.
func TestPluginInfo(t *testing.T) {
	info := PluginInfo{
		Name:        "test-plugin",
		Description: "A test plugin",
		Version:     "1.0.0",
		Author:      "Test Author",
		License:     "MIT",
		Homepage:    "https://example.com",
		Dependencies: []Dependency{
			{Name: "dep1", Version: ">=1.0.0"},
		},
	}

	if info.Name != "test-plugin" {
		t.Errorf("Name = %q, want %q", info.Name, "test-plugin")
	}
	if info.Description != "A test plugin" {
		t.Errorf("Description = %q, want %q", info.Description, "A test plugin")
	}
	if info.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", info.Version, "1.0.0")
	}
	if info.Author != "Test Author" {
		t.Errorf("Author = %q, want %q", info.Author, "Test Author")
	}
	if info.License != "MIT" {
		t.Errorf("License = %q, want %q", info.License, "MIT")
	}
	if info.Homepage != "https://example.com" {
		t.Errorf("Homepage = %q, want %q", info.Homepage, "https://example.com")
	}
	if len(info.Dependencies) != 1 {
		t.Errorf("Dependencies length = %d, want 1", len(info.Dependencies))
	}
}

// TestPluginErrors tests that error variables are properly defined.
func TestPluginErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"ErrPluginNotFound", ErrPluginNotFound, "plugin not found"},
		{"ErrPluginExists", ErrPluginExists, "plugin already registered"},
		{"ErrNilPlugin", ErrNilPlugin, "plugin cannot be nil"},
		{"ErrEmptyName", ErrEmptyName, "plugin name cannot be empty"},
		{"ErrAlreadyInitialized", ErrAlreadyInitialized, "plugins already initialized"},
		{"ErrNotInitialized", ErrNotInitialized, "plugins not initialized"},
		{"ErrCircularDependency", ErrCircularDependency, "circular dependency detected"},
		{"ErrDependencyNotFound", ErrDependencyNotFound, "required dependency not found"},
		{"ErrVersionMismatch", ErrVersionMismatch, "version constraint not satisfied"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Fatal("error is nil")
			}
			if tt.err.Error() != tt.want {
				t.Errorf("Error() = %q, want %q", tt.err.Error(), tt.want)
			}
		})
	}
}

// TestErrorWrapping tests that errors can be wrapped properly.
func TestErrorWrapping(t *testing.T) {
	wrapped := errors.New("wrapped: plugin not found")
	if errors.Is(wrapped, ErrPluginNotFound) {
		t.Error("errors.Is should not match without wrapping")
	}
}
