package theme

import "testing"

// testBundle is a minimal Bundle implementation for testing the interface.
type testBundle struct {
	name     string
	theme    *Theme
	painters map[string]any
}

func (b *testBundle) Name() string      { return b.name }
func (b *testBundle) BaseTheme() *Theme { return b.theme }
func (b *testBundle) Painter(w string) any {
	return b.painters[w]
}
func (b *testBundle) Painters() map[string]any {
	out := make(map[string]any, len(b.painters))
	for k, v := range b.painters {
		out[k] = v
	}
	return out
}

func TestBundle_Interface(t *testing.T) {
	th := DefaultLight()
	b := &testBundle{
		name:  "Test Bundle",
		theme: th,
		painters: map[string]any{
			"button":   "mock-button-painter",
			"checkbox": "mock-checkbox-painter",
		},
	}

	// Compile-time interface satisfaction.
	var _ Bundle = b

	if b.Name() != "Test Bundle" {
		t.Errorf("Name() = %q, want %q", b.Name(), "Test Bundle")
	}

	if b.BaseTheme() != th {
		t.Error("BaseTheme() should return the provided theme")
	}

	if b.Painter("button") != "mock-button-painter" {
		t.Errorf("Painter(button) = %v, want mock-button-painter", b.Painter("button"))
	}

	if b.Painter("nonexistent") != nil {
		t.Errorf("Painter(nonexistent) = %v, want nil", b.Painter("nonexistent"))
	}

	painters := b.Painters()
	if len(painters) != 2 {
		t.Errorf("Painters() returned %d entries, want 2", len(painters))
	}

	// Verify snapshot semantics: mutating returned map doesn't affect bundle.
	painters["extra"] = "should-not-leak"
	if b.Painter("extra") != nil {
		t.Error("mutating Painters() result should not affect bundle")
	}
}

func TestBundle_EmptyPainters(t *testing.T) {
	b := &testBundle{
		name:     "Empty",
		theme:    DefaultDark(),
		painters: map[string]any{},
	}

	if len(b.Painters()) != 0 {
		t.Error("empty bundle should return empty map")
	}

	if b.Painter("button") != nil {
		t.Error("empty bundle should return nil for any widget")
	}
}
