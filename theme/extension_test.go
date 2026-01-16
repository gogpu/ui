package theme

import (
	"sync"
	"testing"

	"github.com/gogpu/ui/widget"
)

// MockExtension is a test implementation of ThemeExtension.
type MockExtension struct {
	name       string
	Color      widget.Color
	Label      string
	Size       float32
	MergeCalls int
	LerpCalls  int
}

func (m *MockExtension) Name() string {
	return m.name
}

func (m *MockExtension) Merge(other ThemeExtension) ThemeExtension {
	m.MergeCalls++
	if o, ok := other.(*MockExtension); ok {
		return &MockExtension{
			name:  m.name,
			Color: o.Color,
			Label: o.Label,
			Size:  o.Size,
		}
	}
	return m
}

func (m *MockExtension) Lerp(other ThemeExtension, t float32) ThemeExtension {
	m.LerpCalls++
	if o, ok := other.(*MockExtension); ok {
		return &MockExtension{
			name:  m.name,
			Color: m.Color.Lerp(o.Color, t),
			Label: LerpString(m.Label, o.Label, t),
			Size:  LerpFloat32(m.Size, o.Size, t),
		}
	}
	return m
}

func (m *MockExtension) CopyWith(overrides map[string]any) ThemeExtension {
	result := *m
	if v, ok := overrides["color"].(widget.Color); ok {
		result.Color = v
	}
	if v, ok := overrides["label"].(string); ok {
		result.Label = v
	}
	if v, ok := overrides["size"].(float32); ok {
		result.Size = v
	}
	return &result
}

// NewMockExtension creates a new MockExtension with defaults.
func NewMockExtension(name string) *MockExtension {
	return &MockExtension{
		name:  name,
		Color: widget.Hex(0xFF0000),
		Label: "default",
		Size:  10.0,
	}
}

// DifferentExtension is a different type for testing type mismatch.
type DifferentExtension struct {
	name  string
	Value int
}

func (d *DifferentExtension) Name() string                                        { return d.name }
func (d *DifferentExtension) Merge(other ThemeExtension) ThemeExtension           { return d }
func (d *DifferentExtension) Lerp(other ThemeExtension, t float32) ThemeExtension { return d }
func (d *DifferentExtension) CopyWith(overrides map[string]any) ThemeExtension    { return d }

// --- typedExtensions tests ---

func TestNewTypedExtensions(t *testing.T) {
	te := newTypedExtensions()

	if te == nil {
		t.Fatal("newTypedExtensions() returned nil")
	}
	if te.extensions == nil {
		t.Error("extensions map should be initialized")
	}
}

func TestTypedExtensions_Register(t *testing.T) {
	te := newTypedExtensions()
	ext := NewMockExtension("test")

	te.register(ext)

	got := te.get("test")
	if got == nil {
		t.Fatal("extension not found after register")
	}
	if got.Name() != "test" {
		t.Errorf("got Name() = %v, want test", got.Name())
	}
}

func TestTypedExtensions_RegisterOverwrite(t *testing.T) {
	te := newTypedExtensions()
	ext1 := &MockExtension{name: "test", Label: "first"}
	ext2 := &MockExtension{name: "test", Label: "second"}

	te.register(ext1)
	te.register(ext2)

	got := te.get("test")
	if mock, ok := got.(*MockExtension); ok {
		if mock.Label != "second" {
			t.Errorf("Label = %v, want second", mock.Label)
		}
	} else {
		t.Fatal("failed to cast to MockExtension")
	}
}

func TestTypedExtensions_Get_NotFound(t *testing.T) {
	te := newTypedExtensions()

	got := te.get("nonexistent")
	if got != nil {
		t.Errorf("got = %v, want nil for nonexistent key", got)
	}
}

func TestTypedExtensions_Clone(t *testing.T) {
	te := newTypedExtensions()
	ext := NewMockExtension("test")
	te.register(ext)

	cloned := te.clone()

	// Verify clone has the extension
	got := cloned.get("test")
	if got == nil {
		t.Fatal("cloned extensions should have test")
	}

	// Verify modifying clone doesn't affect original
	cloned.register(&MockExtension{name: "test", Label: "modified"})
	original := te.get("test")
	if mock, ok := original.(*MockExtension); ok {
		if mock.Label == "modified" {
			t.Error("modifying clone affected original")
		}
	}
}

func TestTypedExtensions_All(t *testing.T) {
	te := newTypedExtensions()
	te.register(NewMockExtension("ext1"))
	te.register(NewMockExtension("ext2"))
	te.register(NewMockExtension("ext3"))

	all := te.all()

	if len(all) != 3 {
		t.Errorf("len(all) = %d, want 3", len(all))
	}
	if _, ok := all["ext1"]; !ok {
		t.Error("ext1 not found")
	}
	if _, ok := all["ext2"]; !ok {
		t.Error("ext2 not found")
	}
	if _, ok := all["ext3"]; !ok {
		t.Error("ext3 not found")
	}

	// Verify returned map is a copy
	delete(all, "ext1")
	if te.get("ext1") == nil {
		t.Error("modifying all() result affected original")
	}
}

func TestTypedExtensions_Concurrent(t *testing.T) {
	te := newTypedExtensions()
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			te.register(&MockExtension{name: "concurrent", Size: float32(n)})
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = te.get("concurrent")
		}()
	}

	// Concurrent clones
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = te.clone()
		}()
	}

	// Concurrent all
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = te.all()
		}()
	}

	wg.Wait()
}

// --- ExtensionAs tests ---

func TestExtensionAs(t *testing.T) {
	th := New("Test", ModeLight)
	ext := NewMockExtension("corporate")
	th.RegisterExtension(ext)

	got, ok := ExtensionAs[*MockExtension](th, "corporate")
	if !ok {
		t.Fatal("ExtensionAs returned false")
	}
	if got == nil {
		t.Fatal("ExtensionAs returned nil")
	}
	if got.Name() != "corporate" {
		t.Errorf("got.Name() = %v, want corporate", got.Name())
	}
}

func TestExtensionAs_NotFound(t *testing.T) {
	th := New("Test", ModeLight)

	got, ok := ExtensionAs[*MockExtension](th, "nonexistent")
	if ok {
		t.Error("ExtensionAs should return false for nonexistent")
	}
	if got != nil {
		t.Error("got should be nil for nonexistent")
	}
}

func TestExtensionAs_WrongType(t *testing.T) {
	th := New("Test", ModeLight)
	th.RegisterExtension(&DifferentExtension{name: "test", Value: 42})

	got, ok := ExtensionAs[*MockExtension](th, "test")
	if ok {
		t.Error("ExtensionAs should return false for wrong type")
	}
	if got != nil {
		t.Error("got should be nil for wrong type")
	}
}

func TestExtensionAs_NilTheme(t *testing.T) {
	var th *Theme

	got, ok := ExtensionAs[*MockExtension](th, "test")
	if ok {
		t.Error("ExtensionAs should return false for nil theme")
	}
	if got != nil {
		t.Error("got should be nil for nil theme")
	}
}

// --- Helper function tests ---

func TestLerpString(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		t    float32
		want string
	}{
		{"t=0", "hello", "world", 0.0, "hello"},
		{"t=0.25", "hello", "world", 0.25, "hello"},
		{"t=0.49", "hello", "world", 0.49, "hello"},
		{"t=0.5", "hello", "world", 0.5, "world"},
		{"t=0.75", "hello", "world", 0.75, "world"},
		{"t=1", "hello", "world", 1.0, "world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LerpString(tt.a, tt.b, tt.t)
			if got != tt.want {
				t.Errorf("LerpString(%q, %q, %v) = %q, want %q", tt.a, tt.b, tt.t, got, tt.want)
			}
		})
	}
}

func TestLerpFloat32(t *testing.T) {
	tests := []struct {
		name string
		a    float32
		b    float32
		t    float32
		want float32
	}{
		{"t=0", 0.0, 100.0, 0.0, 0.0},
		{"t=0.5", 0.0, 100.0, 0.5, 50.0},
		{"t=1", 0.0, 100.0, 1.0, 100.0},
		{"negative to positive", -50.0, 50.0, 0.5, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LerpFloat32(tt.a, tt.b, tt.t)
			if got != tt.want {
				t.Errorf("LerpFloat32(%v, %v, %v) = %v, want %v", tt.a, tt.b, tt.t, got, tt.want)
			}
		})
	}
}

func TestLerpInt(t *testing.T) {
	tests := []struct {
		name string
		a    int
		b    int
		t    float32
		want int
	}{
		{"t=0", 0, 100, 0.0, 0},
		{"t=0.5", 0, 100, 0.5, 50},
		{"t=1", 0, 100, 1.0, 100},
		{"t=0.25 rounds", 0, 10, 0.25, 3}, // 2.5 + 0.5 = 3
		{"negative", -10, 10, 0.5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LerpInt(tt.a, tt.b, tt.t)
			if got != tt.want {
				t.Errorf("LerpInt(%v, %v, %v) = %v, want %v", tt.a, tt.b, tt.t, got, tt.want)
			}
		})
	}
}

// --- Theme extension method tests ---

func TestTheme_RegisterExtension(t *testing.T) {
	th := New("Test", ModeLight)
	ext := NewMockExtension("corporate")

	th.RegisterExtension(ext)

	got := th.TypedExtension("corporate")
	if got == nil {
		t.Fatal("extension not found after RegisterExtension")
	}
	if got.Name() != "corporate" {
		t.Errorf("Name() = %v, want corporate", got.Name())
	}
}

func TestTheme_RegisterExtension_NilTypedExts(t *testing.T) {
	th := &Theme{} // No typedExts initialized
	ext := NewMockExtension("test")

	th.RegisterExtension(ext)

	got := th.TypedExtension("test")
	if got == nil {
		t.Fatal("extension not found")
	}
}

func TestTheme_TypedExtension_NilTypedExts(t *testing.T) {
	th := &Theme{} // No typedExts initialized

	got := th.TypedExtension("test")
	if got != nil {
		t.Errorf("got = %v, want nil for uninitialized typedExts", got)
	}
}

func TestTheme_TypedExtensions(t *testing.T) {
	th := New("Test", ModeLight)
	th.RegisterExtension(NewMockExtension("ext1"))
	th.RegisterExtension(NewMockExtension("ext2"))

	all := th.TypedExtensions()

	if len(all) != 2 {
		t.Errorf("len(all) = %d, want 2", len(all))
	}
}

func TestTheme_TypedExtensions_NilTypedExts(t *testing.T) {
	th := &Theme{}

	all := th.TypedExtensions()
	if all != nil {
		t.Errorf("got = %v, want nil for uninitialized typedExts", all)
	}
}

func TestTheme_MergeExtensions(t *testing.T) {
	base := New("Base", ModeLight)
	base.RegisterExtension(&MockExtension{name: "shared", Label: "base", Size: 10})
	base.RegisterExtension(&MockExtension{name: "baseonly", Label: "base only"})

	child := New("Child", ModeLight)
	child.RegisterExtension(&MockExtension{name: "shared", Label: "child", Size: 20})
	child.RegisterExtension(&MockExtension{name: "childonly", Label: "child only"})

	base.MergeExtensions(child)

	// Check shared extension was merged
	shared := base.TypedExtension("shared")
	if shared == nil {
		t.Fatal("shared extension should exist")
	}
	if mock, ok := shared.(*MockExtension); ok {
		// Merge returns child's values
		if mock.Label != "child" {
			t.Errorf("merged Label = %v, want child", mock.Label)
		}
	}

	// Check child-only extension was copied
	childOnly := base.TypedExtension("childonly")
	if childOnly == nil {
		t.Fatal("childonly extension should exist after merge")
	}

	// Check base-only extension still exists
	baseOnly := base.TypedExtension("baseonly")
	if baseOnly == nil {
		t.Fatal("baseonly extension should still exist")
	}
}

func TestTheme_MergeExtensions_NilOther(t *testing.T) {
	th := New("Test", ModeLight)
	th.RegisterExtension(NewMockExtension("test"))

	// Should not panic
	th.MergeExtensions(nil)

	// Extension should still exist
	if th.TypedExtension("test") == nil {
		t.Error("extension should not be affected by nil merge")
	}
}

func TestTheme_MergeExtensions_NilOtherTypedExts(t *testing.T) {
	th := New("Test", ModeLight)
	th.RegisterExtension(NewMockExtension("test"))
	other := &Theme{}

	// Should not panic
	th.MergeExtensions(other)
}

func TestTheme_MergeExtensions_NilThisTypedExts(t *testing.T) {
	th := &Theme{}
	other := New("Other", ModeLight)
	other.RegisterExtension(NewMockExtension("test"))

	th.MergeExtensions(other)

	if th.TypedExtension("test") == nil {
		t.Error("extension should be copied from other")
	}
}

func TestTheme_LerpExtensions(t *testing.T) {
	th1 := New("Theme1", ModeLight)
	th1.RegisterExtension(&MockExtension{
		name:  "test",
		Color: widget.RGB(1, 0, 0), // Red
		Size:  0,
	})

	th2 := New("Theme2", ModeLight)
	th2.RegisterExtension(&MockExtension{
		name:  "test",
		Color: widget.RGB(0, 0, 1), // Blue
		Size:  100,
	})

	th1.LerpExtensions(th2, 0.5)

	ext := th1.TypedExtension("test")
	if ext == nil {
		t.Fatal("extension should exist after lerp")
	}

	mock, ok := ext.(*MockExtension)
	if !ok {
		t.Fatal("failed to cast to MockExtension")
	}

	// Color should be interpolated (purple-ish)
	if mock.Color.R != 0.5 {
		t.Errorf("Color.R = %v, want 0.5", mock.Color.R)
	}
	if mock.Size != 50 {
		t.Errorf("Size = %v, want 50", mock.Size)
	}
}

func TestTheme_LerpExtensions_NilOther(t *testing.T) {
	th := New("Test", ModeLight)
	th.RegisterExtension(NewMockExtension("test"))

	// Should not panic
	th.LerpExtensions(nil, 0.5)
}

func TestTheme_LerpExtensions_NilOtherTypedExts(t *testing.T) {
	th := New("Test", ModeLight)
	th.RegisterExtension(NewMockExtension("test"))
	other := &Theme{}

	// Should not panic
	th.LerpExtensions(other, 0.5)
}

func TestTheme_LerpExtensions_NilThisTypedExts(t *testing.T) {
	th := &Theme{}
	other := New("Other", ModeLight)
	other.RegisterExtension(NewMockExtension("test"))

	// Should not panic
	th.LerpExtensions(other, 0.5)
}

func TestTheme_LerpExtensions_NoMatchingExtension(t *testing.T) {
	th1 := New("Theme1", ModeLight)
	th1.RegisterExtension(&MockExtension{name: "ext1"})

	th2 := New("Theme2", ModeLight)
	th2.RegisterExtension(&MockExtension{name: "ext2"})

	// Should not panic, ext1 has no matching ext2
	th1.LerpExtensions(th2, 0.5)

	// ext1 should still exist
	if th1.TypedExtension("ext1") == nil {
		t.Error("ext1 should still exist")
	}
}

func TestTheme_Clone_WithTypedExtensions(t *testing.T) {
	th := New("Test", ModeLight)
	th.RegisterExtension(&MockExtension{name: "test", Label: "original"})

	clone := th.Clone()

	// Verify extension was cloned
	ext := clone.TypedExtension("test")
	if ext == nil {
		t.Fatal("extension should be cloned")
	}

	// Verify modifying clone doesn't affect original
	clone.RegisterExtension(&MockExtension{name: "test", Label: "modified"})

	origExt := th.TypedExtension("test")
	if mock, ok := origExt.(*MockExtension); ok {
		if mock.Label == "modified" {
			t.Error("modifying clone affected original")
		}
	}
}

func TestTheme_Clone_NilTypedExts(t *testing.T) {
	th := &Theme{
		Name: "Test",
		Mode: ModeLight,
	}

	clone := th.Clone()

	// Should not panic and should have empty typedExts
	if clone.typedExts == nil {
		t.Error("clone should have initialized typedExts")
	}
}

// --- MockExtension method tests ---

func TestMockExtension_Name(t *testing.T) {
	ext := NewMockExtension("myext")

	if ext.Name() != "myext" {
		t.Errorf("Name() = %v, want myext", ext.Name())
	}
}

func TestMockExtension_Merge(t *testing.T) {
	ext1 := &MockExtension{name: "test", Label: "first", Size: 10}
	ext2 := &MockExtension{name: "test", Label: "second", Size: 20}

	merged := ext1.Merge(ext2)

	mock, ok := merged.(*MockExtension)
	if !ok {
		t.Fatal("Merge should return MockExtension")
	}
	if mock.Label != "second" {
		t.Errorf("Label = %v, want second", mock.Label)
	}
	if mock.Size != 20 {
		t.Errorf("Size = %v, want 20", mock.Size)
	}
}

func TestMockExtension_Merge_WrongType(t *testing.T) {
	ext1 := &MockExtension{name: "test", Label: "original"}
	ext2 := &DifferentExtension{name: "test", Value: 42}

	merged := ext1.Merge(ext2)

	// Should return original when types don't match
	if merged != ext1 {
		t.Error("Merge with wrong type should return original")
	}
}

func TestMockExtension_Lerp(t *testing.T) {
	ext1 := &MockExtension{
		name:  "test",
		Color: widget.RGB(1, 0, 0),
		Size:  0,
	}
	ext2 := &MockExtension{
		name:  "test",
		Color: widget.RGB(0, 0, 1),
		Size:  100,
	}

	lerped := ext1.Lerp(ext2, 0.5)

	mock, ok := lerped.(*MockExtension)
	if !ok {
		t.Fatal("Lerp should return MockExtension")
	}
	if mock.Color.R != 0.5 {
		t.Errorf("Color.R = %v, want 0.5", mock.Color.R)
	}
	if mock.Size != 50 {
		t.Errorf("Size = %v, want 50", mock.Size)
	}
}

func TestMockExtension_Lerp_WrongType(t *testing.T) {
	ext1 := &MockExtension{name: "test", Label: "original"}
	ext2 := &DifferentExtension{name: "test", Value: 42}

	lerped := ext1.Lerp(ext2, 0.5)

	// Should return original when types don't match
	if lerped != ext1 {
		t.Error("Lerp with wrong type should return original")
	}
}

func TestMockExtension_CopyWith(t *testing.T) {
	ext := &MockExtension{
		name:  "test",
		Color: widget.RGB(1, 0, 0),
		Label: "original",
		Size:  10,
	}

	copied := ext.CopyWith(map[string]any{
		"label": "modified",
		"size":  float32(20),
	})

	mock, ok := copied.(*MockExtension)
	if !ok {
		t.Fatal("CopyWith should return MockExtension")
	}
	if mock.Label != "modified" {
		t.Errorf("Label = %v, want modified", mock.Label)
	}
	if mock.Size != 20 {
		t.Errorf("Size = %v, want 20", mock.Size)
	}
	// Color should be unchanged
	if mock.Color.R != 1 {
		t.Errorf("Color.R = %v, want 1", mock.Color.R)
	}
	// Original should be unchanged
	if ext.Label != "original" {
		t.Error("CopyWith modified original")
	}
}

func TestMockExtension_CopyWith_ColorOverride(t *testing.T) {
	ext := &MockExtension{
		name:  "test",
		Color: widget.RGB(1, 0, 0),
	}

	copied := ext.CopyWith(map[string]any{
		"color": widget.RGB(0, 1, 0),
	})

	mock := copied.(*MockExtension)
	if mock.Color.G != 1 {
		t.Errorf("Color.G = %v, want 1", mock.Color.G)
	}
}
