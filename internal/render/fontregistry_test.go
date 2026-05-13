package render

import (
	"sync"
	"testing"

	"github.com/gogpu/ui/internal/render/fonts"
	"github.com/gogpu/ui/theme/font"
)

func TestNewFontRegistry_InterPreRegistered(t *testing.T) {
	r := NewFontRegistry()

	if !r.HasFamily("Inter") {
		t.Fatal("NewFontRegistry should pre-register Inter family")
	}

	names := r.FamilyNames()
	found := false
	for _, n := range names {
		if n == "Inter" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("FamilyNames() = %v, want to contain Inter", names)
	}
}

func TestFontRegistry_Resolve_InterRegular(t *testing.T) {
	r := NewFontRegistry()

	src := r.Resolve("Inter", font.Regular, font.Normal)
	if src == nil {
		t.Fatal("Resolve(Inter, Regular, Normal) returned nil")
	}
}

func TestFontRegistry_Resolve_InterBold(t *testing.T) {
	r := NewFontRegistry()

	src := r.Resolve("Inter", font.Bold, font.Normal)
	if src == nil {
		t.Fatal("Resolve(Inter, Bold, Normal) returned nil")
	}
}

func TestFontRegistry_Resolve_FallbackToInter(t *testing.T) {
	r := NewFontRegistry()

	// Unknown family should fall back to Inter.
	src := r.Resolve("NonExistent", font.Regular, font.Normal)
	if src == nil {
		t.Fatal("Resolve for unknown family should fall back to Inter, got nil")
	}
}

func TestFontRegistry_Register_ValidFont(t *testing.T) {
	r := NewFontRegistry()

	// Register Inter data under a custom family name.
	err := r.Register("TestFamily", font.Regular, font.Normal, fonts.InterRegular)
	if err != nil {
		t.Fatalf("Register valid font data: %v", err)
	}

	if !r.HasFamily("TestFamily") {
		t.Fatal("HasFamily(TestFamily) should be true after Register")
	}

	src := r.Resolve("TestFamily", font.Regular, font.Normal)
	if src == nil {
		t.Fatal("Resolve(TestFamily) should return a FontSource after Register")
	}
}

func TestFontRegistry_Register_EmptyData(t *testing.T) {
	r := NewFontRegistry()

	err := r.Register("Empty", font.Regular, font.Normal, nil)
	if err == nil {
		t.Fatal("Register with nil data should return error")
	}

	err = r.Register("Empty", font.Regular, font.Normal, []byte{})
	if err == nil {
		t.Fatal("Register with empty data should return error")
	}
}

func TestFontRegistry_Register_InvalidData(t *testing.T) {
	r := NewFontRegistry()

	err := r.Register("Bad", font.Regular, font.Normal, []byte("not a font"))
	if err == nil {
		t.Fatal("Register with invalid font data should return error")
	}
}

func TestFontRegistry_Register_MultipleWeights(t *testing.T) {
	r := NewFontRegistry()

	if err := r.Register("Multi", font.Regular, font.Normal, fonts.InterRegular); err != nil {
		t.Fatalf("Register regular: %v", err)
	}
	if err := r.Register("Multi", font.Bold, font.Normal, fonts.InterBold); err != nil {
		t.Fatalf("Register bold: %v", err)
	}

	regular := r.Resolve("Multi", font.Regular, font.Normal)
	bold := r.Resolve("Multi", font.Bold, font.Normal)

	if regular == nil || bold == nil {
		t.Fatal("Both regular and bold should resolve for Multi family")
	}

	// They should be different FontSource instances (different font data).
	if regular == bold {
		t.Error("Regular and Bold should resolve to different FontSources")
	}
}

func TestFontRegistry_Resolve_CSSWeightFallback(t *testing.T) {
	r := NewFontRegistry()

	// Register only Bold for a family.
	if err := r.Register("BoldOnly", font.Bold, font.Normal, fonts.InterBold); err != nil {
		t.Fatalf("Register: %v", err)
	}

	// Request Regular — CSS weight matching should fall back to Bold.
	src := r.Resolve("BoldOnly", font.Regular, font.Normal)
	if src == nil {
		t.Fatal("CSS weight fallback should resolve to Bold when Regular unavailable")
	}
}

func TestFontRegistry_Resolve_Caching(t *testing.T) {
	r := NewFontRegistry()

	// First resolve creates and caches.
	src1 := r.Resolve("Inter", font.Regular, font.Normal)
	// Second resolve should return the cached instance.
	src2 := r.Resolve("Inter", font.Regular, font.Normal)

	if src1 != src2 {
		t.Error("Resolve should return cached FontSource on subsequent calls")
	}
}

func TestFontRegistry_Concurrent(t *testing.T) {
	r := NewFontRegistry()

	var wg sync.WaitGroup
	const goroutines = 10

	// Concurrent reads.
	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = r.Resolve("Inter", font.Regular, font.Normal)
			_ = r.HasFamily("Inter")
			_ = r.FamilyNames()
		}()
	}

	// Concurrent write + reads.
	for i := range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Register with unique names to avoid racing on same entry.
			_ = r.Register("ConcFamily", font.Weight(400+i*100), font.Normal, fonts.InterRegular)
		}()
	}

	wg.Wait()
}

func TestGlobalFontRegistry_Singleton(t *testing.T) {
	r1 := GlobalFontRegistry()
	r2 := GlobalFontRegistry()

	if r1 != r2 {
		t.Error("GlobalFontRegistry should return the same singleton")
	}

	if r1 == nil {
		t.Fatal("GlobalFontRegistry should not return nil")
	}

	if !r1.HasFamily("Inter") {
		t.Error("Global registry should have Inter pre-registered")
	}
}

func TestFontRegistry_FamilyNames_Sorted(t *testing.T) {
	r := NewFontRegistry()

	// Register additional families.
	_ = r.Register("Zebra", font.Regular, font.Normal, fonts.InterRegular)
	_ = r.Register("Alpha", font.Regular, font.Normal, fonts.InterRegular)

	names := r.FamilyNames()
	for i := 1; i < len(names); i++ {
		if names[i-1] > names[i] {
			t.Errorf("FamilyNames not sorted: %v", names)
			break
		}
	}
}
