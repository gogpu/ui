package plugin

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
)

// TestNoopAssetLoader tests the no-op asset loader.
func TestNoopAssetLoader(t *testing.T) {
	loader := &noopAssetLoader{}

	// All methods should succeed silently
	if err := loader.LoadFont("test", []byte("data")); err != nil {
		t.Errorf("LoadFont returned error: %v", err)
	}
	if err := loader.LoadIcon("test", []byte("data")); err != nil {
		t.Errorf("LoadIcon returned error: %v", err)
	}
	if err := loader.LoadImage("test", []byte("data")); err != nil {
		t.Errorf("LoadImage returned error: %v", err)
	}
}

// TestNoopAssetLoaderInterface verifies interface implementation.
func TestNoopAssetLoaderInterface(t *testing.T) {
	var _ AssetLoader = (*noopAssetLoader)(nil)
}

// TestMemoryAssetLoaderInterface verifies interface implementation.
func TestMemoryAssetLoaderInterface(t *testing.T) {
	var _ AssetLoader = (*MemoryAssetLoader)(nil)
}

// TestNewMemoryAssetLoader tests creating a new memory asset loader.
func TestNewMemoryAssetLoader(t *testing.T) {
	loader := NewMemoryAssetLoader()

	if loader == nil {
		t.Fatal("NewMemoryAssetLoader returned nil")
	}
	if loader.fonts == nil {
		t.Error("fonts map is nil")
	}
	if loader.icons == nil {
		t.Error("icons map is nil")
	}
	if loader.images == nil {
		t.Error("images map is nil")
	}
}

// TestMemoryAssetLoaderFont tests font loading and retrieval.
func TestMemoryAssetLoaderFont(t *testing.T) {
	loader := NewMemoryAssetLoader()
	data := []byte("font-data")

	// Load font
	if err := loader.LoadFont("roboto", data); err != nil {
		t.Fatalf("LoadFont failed: %v", err)
	}

	// Retrieve font
	retrieved, ok := loader.GetFont("roboto")
	if !ok {
		t.Fatal("Font not found")
	}
	if !bytes.Equal(retrieved, data) {
		t.Errorf("Font data mismatch: got %v, want %v", retrieved, data)
	}

	// Count
	if loader.FontCount() != 1 {
		t.Errorf("FontCount() = %d, want 1", loader.FontCount())
	}
}

// TestMemoryAssetLoaderIcon tests icon loading and retrieval.
func TestMemoryAssetLoaderIcon(t *testing.T) {
	loader := NewMemoryAssetLoader()
	data := []byte("icon-data")

	// Load icon
	if err := loader.LoadIcon("add", data); err != nil {
		t.Fatalf("LoadIcon failed: %v", err)
	}

	// Retrieve icon
	retrieved, ok := loader.GetIcon("add")
	if !ok {
		t.Fatal("Icon not found")
	}
	if !bytes.Equal(retrieved, data) {
		t.Errorf("Icon data mismatch: got %v, want %v", retrieved, data)
	}

	// Count
	if loader.IconCount() != 1 {
		t.Errorf("IconCount() = %d, want 1", loader.IconCount())
	}
}

// TestMemoryAssetLoaderImage tests image loading and retrieval.
func TestMemoryAssetLoaderImage(t *testing.T) {
	loader := NewMemoryAssetLoader()
	data := []byte("image-data")

	// Load image
	if err := loader.LoadImage("logo", data); err != nil {
		t.Fatalf("LoadImage failed: %v", err)
	}

	// Retrieve image
	retrieved, ok := loader.GetImage("logo")
	if !ok {
		t.Fatal("Image not found")
	}
	if !bytes.Equal(retrieved, data) {
		t.Errorf("Image data mismatch: got %v, want %v", retrieved, data)
	}

	// Count
	if loader.ImageCount() != 1 {
		t.Errorf("ImageCount() = %d, want 1", loader.ImageCount())
	}
}

// TestMemoryAssetLoaderNotFound tests retrieval of non-existent assets.
func TestMemoryAssetLoaderNotFound(t *testing.T) {
	loader := NewMemoryAssetLoader()

	if _, ok := loader.GetFont("nonexistent"); ok {
		t.Error("GetFont should return false for non-existent font")
	}
	if _, ok := loader.GetIcon("nonexistent"); ok {
		t.Error("GetIcon should return false for non-existent icon")
	}
	if _, ok := loader.GetImage("nonexistent"); ok {
		t.Error("GetImage should return false for non-existent image")
	}
}

// TestMemoryAssetLoaderOverwrite tests overwriting existing assets.
func TestMemoryAssetLoaderOverwrite(t *testing.T) {
	loader := NewMemoryAssetLoader()

	data1 := []byte("original")
	data2 := []byte("updated")

	// Load original
	_ = loader.LoadFont("test", data1)

	// Overwrite
	_ = loader.LoadFont("test", data2)

	// Should have updated data
	retrieved, _ := loader.GetFont("test")
	if !bytes.Equal(retrieved, data2) {
		t.Errorf("Font not overwritten: got %v, want %v", retrieved, data2)
	}

	// Count should still be 1
	if loader.FontCount() != 1 {
		t.Errorf("FontCount() = %d, want 1", loader.FontCount())
	}
}

// TestMemoryAssetLoaderClear tests clearing all assets.
func TestMemoryAssetLoaderClear(t *testing.T) {
	loader := NewMemoryAssetLoader()

	// Load some assets
	_ = loader.LoadFont("font1", []byte("data"))
	_ = loader.LoadFont("font2", []byte("data"))
	_ = loader.LoadIcon("icon1", []byte("data"))
	_ = loader.LoadImage("image1", []byte("data"))

	if loader.FontCount() != 2 {
		t.Errorf("FontCount() = %d, want 2", loader.FontCount())
	}

	// Clear
	loader.Clear()

	// All counts should be 0
	if loader.FontCount() != 0 {
		t.Errorf("FontCount() after clear = %d, want 0", loader.FontCount())
	}
	if loader.IconCount() != 0 {
		t.Errorf("IconCount() after clear = %d, want 0", loader.IconCount())
	}
	if loader.ImageCount() != 0 {
		t.Errorf("ImageCount() after clear = %d, want 0", loader.ImageCount())
	}
}

// TestMemoryAssetLoaderDataIsolation tests that stored data is copied.
func TestMemoryAssetLoaderDataIsolation(t *testing.T) {
	loader := NewMemoryAssetLoader()

	data := []byte("original")
	_ = loader.LoadFont("test", data)

	// Modify the original slice
	data[0] = 'X'

	// Retrieved data should be unchanged
	retrieved, _ := loader.GetFont("test")
	if retrieved[0] == 'X' {
		t.Error("Stored data was modified by changing original slice")
	}
}

// TestMemoryAssetLoaderConcurrency tests thread safety.
func TestMemoryAssetLoaderConcurrency(t *testing.T) {
	loader := NewMemoryAssetLoader()
	const n = 100

	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < n; i++ {
		wg.Add(3)
		go func(i int) {
			defer wg.Done()
			_ = loader.LoadFont("font", []byte{byte(i)})
		}(i)
		go func(i int) {
			defer wg.Done()
			_ = loader.LoadIcon("icon", []byte{byte(i)})
		}(i)
		go func(i int) {
			defer wg.Done()
			_ = loader.LoadImage("image", []byte{byte(i)})
		}(i)
	}

	// Concurrent reads
	for i := 0; i < n; i++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			_, _ = loader.GetFont("font")
		}()
		go func() {
			defer wg.Done()
			_, _ = loader.GetIcon("icon")
		}()
		go func() {
			defer wg.Done()
			_, _ = loader.GetImage("image")
		}()
	}

	wg.Wait()

	// Should not panic or have race conditions
	// Just verify we have some data
	if loader.FontCount() != 1 {
		t.Errorf("FontCount() = %d, want 1", loader.FontCount())
	}
}

// TestMemoryAssetLoaderEmptyData tests loading empty data.
func TestMemoryAssetLoaderEmptyData(t *testing.T) {
	loader := NewMemoryAssetLoader()

	// Loading empty data should succeed
	if err := loader.LoadFont("empty", []byte{}); err != nil {
		t.Errorf("LoadFont with empty data failed: %v", err)
	}

	data, ok := loader.GetFont("empty")
	if !ok {
		t.Error("Empty font not found")
	}
	if len(data) != 0 {
		t.Errorf("Expected empty data, got %v", data)
	}
}

// TestMemoryAssetLoaderNilData tests loading nil data.
func TestMemoryAssetLoaderNilData(t *testing.T) {
	loader := NewMemoryAssetLoader()

	// Loading nil data should succeed (treated as empty)
	if err := loader.LoadFont("nil", nil); err != nil {
		t.Errorf("LoadFont with nil data failed: %v", err)
	}

	data, ok := loader.GetFont("nil")
	if !ok {
		t.Error("Nil font not found")
	}
	if len(data) != 0 {
		t.Errorf("Expected empty data, got %v", data)
	}
}

// TestMemoryAssetLoaderFontRegisterer tests that LoadFont calls the registerer.
func TestMemoryAssetLoaderFontRegisterer(t *testing.T) {
	loader := NewMemoryAssetLoader()

	var registeredName string
	var registeredData []byte
	loader.SetFontRegisterer(func(name string, data []byte) error {
		registeredName = name
		registeredData = data
		return nil
	})

	fontData := []byte("fake-font-data")
	if err := loader.LoadFont("TestFont", fontData); err != nil {
		t.Fatalf("LoadFont failed: %v", err)
	}

	if registeredName != "TestFont" {
		t.Errorf("registerer name = %q, want TestFont", registeredName)
	}
	if !bytes.Equal(registeredData, fontData) {
		t.Error("registerer data should match loaded font data")
	}
}

// TestMemoryAssetLoaderFontRegisterer_Error tests that registerer errors propagate.
func TestMemoryAssetLoaderFontRegisterer_Error(t *testing.T) {
	loader := NewMemoryAssetLoader()
	loader.SetFontRegisterer(func(_ string, _ []byte) error {
		return errForTesting
	})

	err := loader.LoadFont("Bad", []byte("data"))
	if err == nil {
		t.Fatal("LoadFont should propagate registerer error")
	}

	// Font should still be stored in memory (store happens before register).
	_, ok := loader.GetFont("Bad")
	if !ok {
		t.Error("Font data should still be stored despite registerer error")
	}
}

// TestMemoryAssetLoaderFontRegisterer_NilRegisterer tests no panic without registerer.
func TestMemoryAssetLoaderFontRegisterer_NilRegisterer(t *testing.T) {
	loader := NewMemoryAssetLoader()
	// No SetFontRegisterer call -- should not panic.

	if err := loader.LoadFont("Test", []byte("data")); err != nil {
		t.Fatalf("LoadFont without registerer should succeed: %v", err)
	}
}

// errForTesting is a sentinel error for tests.
var errForTesting = fmt.Errorf("test error")
