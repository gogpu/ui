package registry

import (
	"errors"
	"sync"
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// mockWidget is a simple Widget implementation for testing.
type mockWidget struct {
	widget.WidgetBase
	label string
	value int
}

func (m *mockWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(100, 50))
}

func (m *mockWidget) Draw(_ widget.Context, _ widget.Canvas) {
	// No-op for testing
}

func (m *mockWidget) Event(_ widget.Context, _ event.Event) bool {
	return false
}

// mockWidgetFactory creates a mockWidget from config.
func mockWidgetFactory(config map[string]any) (Widget, error) {
	w := &mockWidget{}
	if config != nil {
		if label, ok := config["label"].(string); ok {
			w.label = label
		}
		if value, ok := config["value"].(int); ok {
			w.value = value
		}
	}
	return w, nil
}

// errorWidgetFactory always returns an error.
func errorWidgetFactory(_ map[string]any) (Widget, error) {
	return nil, errors.New("factory error")
}

// validatingWidgetFactory requires a "label" config parameter.
func validatingWidgetFactory(config map[string]any) (Widget, error) {
	if config == nil {
		return nil, errors.New("config is required")
	}
	label, ok := config["label"].(string)
	if !ok || label == "" {
		return nil, errors.New("label is required")
	}
	return &mockWidget{label: label}, nil
}

// newTestRegistry creates a fresh registry for testing.
func newTestRegistry() *WidgetRegistry {
	return NewWidgetRegistry()
}

// TestCategory tests the Category type.
func TestCategory(t *testing.T) {
	tests := []struct {
		name     string
		category Category
		str      string
		isValid  bool
	}{
		{"input category", CategoryInput, "input", true},
		{"display category", CategoryDisplay, "display", true},
		{"container category", CategoryContainer, "container", true},
		{"custom category", CategoryCustom, "custom", true},
		{"unknown category", Category("unknown"), "unknown", false},
		{"empty category", Category(""), "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.category.String(); got != tt.str {
				t.Errorf("String() = %q, want %q", got, tt.str)
			}
			if got := tt.category.IsValid(); got != tt.isValid {
				t.Errorf("IsValid() = %v, want %v", got, tt.isValid)
			}
		})
	}
}

// TestWidgetInfoValidate tests WidgetInfo validation.
func TestWidgetInfoValidate(t *testing.T) {
	tests := []struct {
		name    string
		info    WidgetInfo
		wantErr bool
	}{
		{
			name:    "valid info with all fields",
			info:    WidgetInfo{Name: "test", Description: "Test widget", Category: CategoryCustom, Version: "1.0.0"},
			wantErr: false,
		},
		{
			name:    "valid info with name only",
			info:    WidgetInfo{Name: "test"},
			wantErr: false,
		},
		{
			name:    "invalid info with empty name",
			info:    WidgetInfo{Description: "Test widget"},
			wantErr: true,
		},
		{
			name:    "empty info",
			info:    WidgetInfo{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.info.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestNewWidgetRegistry tests registry creation.
func TestNewWidgetRegistry(t *testing.T) {
	r := NewWidgetRegistry()
	if r == nil {
		t.Fatal("NewWidgetRegistry() returned nil")
	}
	if r.widgets == nil {
		t.Error("widgets map is nil")
	}
	if r.info == nil {
		t.Error("info map is nil")
	}
	if r.Count() != 0 {
		t.Errorf("Count() = %d, want 0", r.Count())
	}
}

// TestRegister tests widget registration.
func TestRegister(t *testing.T) {
	tests := []struct {
		name       string
		widgetName string
		factory    WidgetFactory
		info       []WidgetInfo
		wantErr    error
	}{
		{
			name:       "valid registration without info",
			widgetName: "test-widget",
			factory:    mockWidgetFactory,
			info:       nil,
			wantErr:    nil,
		},
		{
			name:       "valid registration with info",
			widgetName: "test-widget-2",
			factory:    mockWidgetFactory,
			info: []WidgetInfo{{
				Name:        "test-widget-2",
				Description: "A test widget",
				Category:    CategoryCustom,
				Version:     "1.0.0",
			}},
			wantErr: nil,
		},
		{
			name:       "registration with empty name",
			widgetName: "",
			factory:    mockWidgetFactory,
			info:       nil,
			wantErr:    ErrEmptyName,
		},
		{
			name:       "registration with nil factory",
			widgetName: "nil-factory",
			factory:    nil,
			info:       nil,
			wantErr:    ErrNilFactory,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestRegistry()
			err := r.Register(tt.widgetName, tt.factory, tt.info...)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Register() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

// TestRegisterDuplicate tests that duplicate registration returns an error.
func TestRegisterDuplicate(t *testing.T) {
	r := newTestRegistry()

	// First registration should succeed
	err := r.Register("widget", mockWidgetFactory)
	if err != nil {
		t.Fatalf("first Register() error = %v", err)
	}

	// Second registration should fail
	err = r.Register("widget", mockWidgetFactory)
	if !errors.Is(err, ErrWidgetExists) {
		t.Errorf("Register() error = %v, want %v", err, ErrWidgetExists)
	}
}

// TestRegisterInfoNameInheritance tests that info.Name inherits registration name.
func TestRegisterInfoNameInheritance(t *testing.T) {
	r := newTestRegistry()

	// Register with empty Name in info
	err := r.Register("my-widget", mockWidgetFactory, WidgetInfo{
		Description: "Test widget",
		Category:    CategoryInput,
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	// Verify info.Name is set to registration name
	info, ok := r.Info("my-widget")
	if !ok {
		t.Fatal("Info() returned false")
	}
	if info.Name != "my-widget" {
		t.Errorf("info.Name = %q, want %q", info.Name, "my-widget")
	}
}

// TestMustRegister tests MustRegister panics on error.
func TestMustRegister(t *testing.T) {
	r := newTestRegistry()

	// Valid registration should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustRegister() panicked unexpectedly: %v", r)
			}
		}()
		r.MustRegister("valid-widget", mockWidgetFactory)
	}()

	// Invalid registration should panic
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustRegister() did not panic for nil factory")
			}
		}()
		r.MustRegister("invalid-widget", nil)
	}()
}

// TestUnregister tests widget unregistration.
func TestUnregister(t *testing.T) {
	r := newTestRegistry()

	// Register a widget
	_ = r.Register("widget", mockWidgetFactory, WidgetInfo{Name: "widget"})

	// Unregister should succeed
	err := r.Unregister("widget")
	if err != nil {
		t.Errorf("Unregister() error = %v", err)
	}

	// Widget should no longer exist
	if r.Has("widget") {
		t.Error("Has() = true after Unregister()")
	}

	// Info should also be removed
	if _, ok := r.Info("widget"); ok {
		t.Error("Info() returned true after Unregister()")
	}

	// Unregister again should fail
	err = r.Unregister("widget")
	if !errors.Is(err, ErrWidgetNotFound) {
		t.Errorf("Unregister() error = %v, want %v", err, ErrWidgetNotFound)
	}
}

// TestCreate tests widget creation.
func TestCreate(t *testing.T) {
	r := newTestRegistry()
	_ = r.Register("mock", mockWidgetFactory)
	_ = r.Register("error", errorWidgetFactory)
	_ = r.Register("validating", validatingWidgetFactory)

	tests := []struct {
		name       string
		widgetName string
		config     map[string]any
		wantErr    bool
		checkErr   error
	}{
		{
			name:       "create with nil config",
			widgetName: "mock",
			config:     nil,
			wantErr:    false,
		},
		{
			name:       "create with config",
			widgetName: "mock",
			config:     map[string]any{"label": "test", "value": 42},
			wantErr:    false,
		},
		{
			name:       "create unregistered widget",
			widgetName: "unregistered",
			config:     nil,
			wantErr:    true,
			checkErr:   ErrWidgetNotFound,
		},
		{
			name:       "create with factory error",
			widgetName: "error",
			config:     nil,
			wantErr:    true,
		},
		{
			name:       "create with validation error",
			widgetName: "validating",
			config:     nil,
			wantErr:    true,
		},
		{
			name:       "create with valid validation",
			widgetName: "validating",
			config:     map[string]any{"label": "test"},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			created, err := r.Create(tt.widgetName, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.checkErr != nil && !errors.Is(err, tt.checkErr) {
				t.Errorf("Create() error = %v, want %v", err, tt.checkErr)
			}
			if !tt.wantErr && created == nil {
				t.Error("Create() returned nil widget without error")
			}
		})
	}
}

// TestCreateWidgetValues tests that created widgets have correct values.
func TestCreateWidgetValues(t *testing.T) {
	r := newTestRegistry()
	_ = r.Register("mock", mockWidgetFactory)

	created, err := r.Create("mock", map[string]any{
		"label": "test-label",
		"value": 123,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	mw, ok := created.(*mockWidget)
	if !ok {
		t.Fatal("created widget is not *mockWidget")
	}

	if mw.label != "test-label" {
		t.Errorf("label = %q, want %q", mw.label, "test-label")
	}
	if mw.value != 123 {
		t.Errorf("value = %d, want %d", mw.value, 123)
	}
}

// TestHas tests the Has method.
func TestHas(t *testing.T) {
	r := newTestRegistry()
	_ = r.Register("exists", mockWidgetFactory)

	if !r.Has("exists") {
		t.Error("Has() = false for registered widget")
	}
	if r.Has("not-exists") {
		t.Error("Has() = true for unregistered widget")
	}
}

// TestInfo tests the Info method.
func TestInfo(t *testing.T) {
	r := newTestRegistry()
	_ = r.Register("widget", mockWidgetFactory, WidgetInfo{
		Name:        "widget",
		Description: "Test widget",
		Category:    CategoryCustom,
		Version:     "1.0.0",
	})

	info, ok := r.Info("widget")
	if !ok {
		t.Fatal("Info() returned false for registered widget")
	}
	if info.Name != "widget" {
		t.Errorf("Name = %q, want %q", info.Name, "widget")
	}
	if info.Description != "Test widget" {
		t.Errorf("Description = %q, want %q", info.Description, "Test widget")
	}
	if info.Category != CategoryCustom {
		t.Errorf("Category = %v, want %v", info.Category, CategoryCustom)
	}
	if info.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", info.Version, "1.0.0")
	}

	// Check unregistered widget
	_, ok = r.Info("not-exists")
	if ok {
		t.Error("Info() returned true for unregistered widget")
	}
}

// TestList tests the List method.
func TestList(t *testing.T) {
	r := newTestRegistry()

	// Empty registry
	if list := r.List(); len(list) != 0 {
		t.Errorf("List() = %v, want empty slice", list)
	}

	// Register some widgets
	_ = r.Register("charlie", mockWidgetFactory)
	_ = r.Register("alpha", mockWidgetFactory)
	_ = r.Register("bravo", mockWidgetFactory)

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

// TestListByCategory tests the ListByCategory method.
func TestListByCategory(t *testing.T) {
	r := newTestRegistry()

	_ = r.Register("button", mockWidgetFactory, WidgetInfo{Name: "button", Category: CategoryInput})
	_ = r.Register("textfield", mockWidgetFactory, WidgetInfo{Name: "textfield", Category: CategoryInput})
	_ = r.Register("label", mockWidgetFactory, WidgetInfo{Name: "label", Category: CategoryDisplay})
	_ = r.Register("panel", mockWidgetFactory, WidgetInfo{Name: "panel", Category: CategoryContainer})

	tests := []struct {
		category Category
		want     []string
	}{
		{CategoryInput, []string{"button", "textfield"}},
		{CategoryDisplay, []string{"label"}},
		{CategoryContainer, []string{"panel"}},
		{CategoryCustom, []string{}},
	}

	for _, tt := range tests {
		t.Run(string(tt.category), func(t *testing.T) {
			got := r.ListByCategory(tt.category)
			if len(got) != len(tt.want) {
				t.Fatalf("ListByCategory(%s) length = %d, want %d", tt.category, len(got), len(tt.want))
			}
			for i, name := range tt.want {
				if got[i] != name {
					t.Errorf("ListByCategory(%s)[%d] = %q, want %q", tt.category, i, got[i], name)
				}
			}
		})
	}
}

// TestCount tests the Count method.
func TestCount(t *testing.T) {
	r := newTestRegistry()

	if c := r.Count(); c != 0 {
		t.Errorf("Count() = %d, want 0", c)
	}

	_ = r.Register("widget1", mockWidgetFactory)
	if c := r.Count(); c != 1 {
		t.Errorf("Count() = %d, want 1", c)
	}

	_ = r.Register("widget2", mockWidgetFactory)
	if c := r.Count(); c != 2 {
		t.Errorf("Count() = %d, want 2", c)
	}

	_ = r.Unregister("widget1")
	if c := r.Count(); c != 1 {
		t.Errorf("Count() = %d, want 1", c)
	}
}

// TestClear tests the Clear method.
func TestClear(t *testing.T) {
	r := newTestRegistry()

	_ = r.Register("widget1", mockWidgetFactory)
	_ = r.Register("widget2", mockWidgetFactory)
	_ = r.Register("widget3", mockWidgetFactory)

	if r.Count() != 3 {
		t.Fatalf("Count() = %d, want 3", r.Count())
	}

	r.Clear()

	if r.Count() != 0 {
		t.Errorf("Count() after Clear() = %d, want 0", r.Count())
	}

	if len(r.List()) != 0 {
		t.Error("List() after Clear() is not empty")
	}

	// Should be able to register again
	err := r.Register("widget1", mockWidgetFactory)
	if err != nil {
		t.Errorf("Register() after Clear() error = %v", err)
	}
}

// TestAllInfo tests the AllInfo method.
func TestAllInfo(t *testing.T) {
	r := newTestRegistry()

	_ = r.Register("beta", mockWidgetFactory, WidgetInfo{Name: "beta", Description: "Beta widget"})
	_ = r.Register("alpha", mockWidgetFactory, WidgetInfo{Name: "alpha", Description: "Alpha widget"})

	infos := r.AllInfo()
	if len(infos) != 2 {
		t.Fatalf("AllInfo() length = %d, want 2", len(infos))
	}

	// Should be sorted by name
	if infos[0].Name != "alpha" {
		t.Errorf("AllInfo()[0].Name = %q, want %q", infos[0].Name, "alpha")
	}
	if infos[1].Name != "beta" {
		t.Errorf("AllInfo()[1].Name = %q, want %q", infos[1].Name, "beta")
	}
}

// TestConcurrentAccess tests thread safety.
func TestConcurrentAccess(t *testing.T) {
	r := newTestRegistry()

	var wg sync.WaitGroup
	const goroutines = 100
	const operations = 100

	// Mix of operations
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				name := "widget"
				switch j % 5 {
				case 0:
					_ = r.Register(name, mockWidgetFactory)
				case 1:
					_ = r.Unregister(name)
				case 2:
					_, _ = r.Create(name, nil)
				case 3:
					_ = r.List()
				case 4:
					_ = r.Has(name)
				}
			}
		}()
	}

	wg.Wait()
	// No race conditions or panics = test passes
}

// TestConcurrentRegisterDifferentNames tests concurrent registration of different widgets.
func TestConcurrentRegisterDifferentNames(t *testing.T) {
	r := newTestRegistry()

	var wg sync.WaitGroup
	const count = 100

	wg.Add(count)
	for i := 0; i < count; i++ {
		go func(id int) {
			defer wg.Done()
			name := "widget-" + string(rune('a'+id%26)) + string(rune('0'+id/26))
			_ = r.Register(name, mockWidgetFactory)
		}(i)
	}

	wg.Wait()

	// All unique widgets should be registered
	if c := r.Count(); c != count {
		t.Errorf("Count() = %d, want %d", c, count)
	}
}

// --- Global registry tests ---

// TestGlobalRegistry tests the global registry instance.
func TestGlobalRegistry(t *testing.T) {
	// Clear global registry for testing
	ClearGlobalRegistry()
	defer ClearGlobalRegistry()

	if GlobalRegistry() == nil {
		t.Fatal("GlobalRegistry() returned nil")
	}
}

// TestGlobalFunctions tests package-level functions.
func TestGlobalFunctions(t *testing.T) {
	ClearGlobalRegistry()
	defer ClearGlobalRegistry()

	// RegisterWidget
	err := RegisterWidget("global-widget", mockWidgetFactory, WidgetInfo{
		Name:        "global-widget",
		Description: "Global test widget",
		Category:    CategoryInput,
		Version:     "1.0.0",
	})
	if err != nil {
		t.Fatalf("RegisterWidget() error = %v", err)
	}

	// HasWidget
	if !HasWidget("global-widget") {
		t.Error("HasWidget() = false for registered widget")
	}

	// GetWidgetInfo
	info, ok := GetWidgetInfo("global-widget")
	if !ok {
		t.Fatal("GetWidgetInfo() returned false")
	}
	if info.Description != "Global test widget" {
		t.Errorf("info.Description = %q, want %q", info.Description, "Global test widget")
	}

	// CreateWidget
	created, err := CreateWidget("global-widget", nil)
	if err != nil {
		t.Fatalf("CreateWidget() error = %v", err)
	}
	if created == nil {
		t.Error("CreateWidget() returned nil widget")
	}

	// ListWidgets
	list := ListWidgets()
	if len(list) != 1 || list[0] != "global-widget" {
		t.Errorf("ListWidgets() = %v, want [global-widget]", list)
	}

	// WidgetCount
	if c := WidgetCount(); c != 1 {
		t.Errorf("WidgetCount() = %d, want 1", c)
	}

	// AllWidgetInfo
	allInfo := AllWidgetInfo()
	if len(allInfo) != 1 {
		t.Fatalf("AllWidgetInfo() length = %d, want 1", len(allInfo))
	}

	// UnregisterWidget
	err = UnregisterWidget("global-widget")
	if err != nil {
		t.Errorf("UnregisterWidget() error = %v", err)
	}

	if HasWidget("global-widget") {
		t.Error("HasWidget() = true after UnregisterWidget()")
	}
}

// TestMustRegisterWidget tests MustRegisterWidget function.
func TestMustRegisterWidget(t *testing.T) {
	ClearGlobalRegistry()
	defer ClearGlobalRegistry()

	// Valid registration should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustRegisterWidget() panicked unexpectedly: %v", r)
			}
		}()
		MustRegisterWidget("must-widget", mockWidgetFactory)
	}()

	// Duplicate registration should panic
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustRegisterWidget() did not panic for duplicate")
			}
		}()
		MustRegisterWidget("must-widget", mockWidgetFactory)
	}()
}

// TestListWidgetsByCategory tests the ListWidgetsByCategory global function.
func TestListWidgetsByCategory(t *testing.T) {
	ClearGlobalRegistry()
	defer ClearGlobalRegistry()

	_ = RegisterWidget("button", mockWidgetFactory, WidgetInfo{Name: "button", Category: CategoryInput})
	_ = RegisterWidget("checkbox", mockWidgetFactory, WidgetInfo{Name: "checkbox", Category: CategoryInput})
	_ = RegisterWidget("label", mockWidgetFactory, WidgetInfo{Name: "label", Category: CategoryDisplay})

	inputWidgets := ListWidgetsByCategory(CategoryInput)
	if len(inputWidgets) != 2 {
		t.Errorf("ListWidgetsByCategory(input) length = %d, want 2", len(inputWidgets))
	}

	displayWidgets := ListWidgetsByCategory(CategoryDisplay)
	if len(displayWidgets) != 1 {
		t.Errorf("ListWidgetsByCategory(display) length = %d, want 1", len(displayWidgets))
	}

	containerWidgets := ListWidgetsByCategory(CategoryContainer)
	if len(containerWidgets) != 0 {
		t.Errorf("ListWidgetsByCategory(container) length = %d, want 0", len(containerWidgets))
	}
}

// --- Benchmarks ---

// BenchmarkRegister benchmarks widget registration.
func BenchmarkRegister(b *testing.B) {
	r := newTestRegistry()
	info := WidgetInfo{Name: "widget", Category: CategoryCustom}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Clear()
		_ = r.Register("widget", mockWidgetFactory, info)
	}
}

// BenchmarkCreate benchmarks widget creation.
func BenchmarkCreate(b *testing.B) {
	r := newTestRegistry()
	_ = r.Register("widget", mockWidgetFactory)
	config := map[string]any{"label": "test", "value": 42}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.Create("widget", config)
	}
}

// BenchmarkHas benchmarks widget lookup.
func BenchmarkHas(b *testing.B) {
	r := newTestRegistry()
	_ = r.Register("widget", mockWidgetFactory)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Has("widget")
	}
}

// BenchmarkList benchmarks listing widgets.
func BenchmarkList(b *testing.B) {
	r := newTestRegistry()
	for i := 0; i < 100; i++ {
		_ = r.Register("widget-"+string(rune('a'+i%26))+string(rune('0'+i/26)), mockWidgetFactory)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.List()
	}
}

// BenchmarkConcurrentCreate benchmarks concurrent widget creation.
func BenchmarkConcurrentCreate(b *testing.B) {
	r := newTestRegistry()
	_ = r.Register("widget", mockWidgetFactory)
	config := map[string]any{"label": "test"}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = r.Create("widget", config)
		}
	})
}
