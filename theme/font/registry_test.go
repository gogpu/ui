package font

import (
	"sync"
	"testing"
)

// testData creates distinguishable font data bytes for test assertions.
func testData(id byte) []byte {
	return []byte{id}
}

func TestNewRegistry(t *testing.T) {
	reg := NewRegistry()
	if reg == nil {
		t.Fatal("NewRegistry returned nil")
	}
	if len(reg.FamilyNames()) != 0 {
		t.Errorf("expected empty registry, got %d families", len(reg.FamilyNames()))
	}
}

func TestRegisterFamily(t *testing.T) {
	reg := NewRegistry()
	reg.RegisterFamily(Family{
		Name: "Inter",
		Faces: []Face{
			{Weight: Regular, Style: Normal, Data: testData(1)},
			{Weight: Bold, Style: Normal, Data: testData(2)},
		},
	})

	if !reg.HasFamily("Inter") {
		t.Fatal("expected Inter to be registered")
	}
	if reg.FaceCount("Inter") != 2 {
		t.Errorf("expected 2 faces, got %d", reg.FaceCount("Inter"))
	}
}

func TestRegisterFamily_Append(t *testing.T) {
	reg := NewRegistry()
	reg.RegisterFamily(Family{
		Name:  "Inter",
		Faces: []Face{{Weight: Regular, Style: Normal, Data: testData(1)}},
	})
	reg.RegisterFamily(Family{
		Name:  "Inter",
		Faces: []Face{{Weight: Bold, Style: Normal, Data: testData(2)}},
	})

	if reg.FaceCount("Inter") != 2 {
		t.Errorf("expected 2 faces after append, got %d", reg.FaceCount("Inter"))
	}
}

func TestRegisterFamily_Replace(t *testing.T) {
	reg := NewRegistry()
	reg.RegisterFamily(Family{
		Name:  "Inter",
		Faces: []Face{{Weight: Regular, Style: Normal, Data: testData(1)}},
	})
	reg.RegisterFamily(Family{
		Name:  "Inter",
		Faces: []Face{{Weight: Regular, Style: Normal, Data: testData(99)}},
	})

	if reg.FaceCount("Inter") != 1 {
		t.Errorf("expected 1 face after replace, got %d", reg.FaceCount("Inter"))
	}
	data, ok := reg.Resolve("Inter", Regular, Normal)
	if !ok {
		t.Fatal("expected resolve to succeed")
	}
	if data[0] != 99 {
		t.Errorf("expected replaced data (99), got %d", data[0])
	}
}

func TestResolve_ExactMatch(t *testing.T) {
	reg := NewRegistry()
	reg.RegisterFamily(Family{
		Name: "Test",
		Faces: []Face{
			{Weight: Regular, Style: Normal, Data: testData(1)},
			{Weight: Bold, Style: Normal, Data: testData(2)},
			{Weight: Regular, Style: Italic, Data: testData(3)},
			{Weight: Bold, Style: Italic, Data: testData(4)},
		},
	})

	tests := []struct {
		name     string
		weight   Weight
		style    Style
		wantByte byte
	}{
		{"regular normal", Regular, Normal, 1},
		{"bold normal", Bold, Normal, 2},
		{"regular italic", Regular, Italic, 3},
		{"bold italic", Bold, Italic, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, ok := reg.Resolve("Test", tt.weight, tt.style)
			if !ok {
				t.Fatal("expected resolve to succeed")
			}
			if data[0] != tt.wantByte {
				t.Errorf("got %d, want %d", data[0], tt.wantByte)
			}
		})
	}
}

func TestResolve_UnknownFamily(t *testing.T) {
	reg := NewRegistry()
	data, ok := reg.Resolve("NonExistent", Regular, Normal)
	if ok {
		t.Error("expected resolve to fail for unknown family")
	}
	if data != nil {
		t.Error("expected nil data for unknown family")
	}
}

func TestResolve_ItalicFallbackToNormal(t *testing.T) {
	reg := NewRegistry()
	reg.RegisterFamily(Family{
		Name: "NoItalic",
		Faces: []Face{
			{Weight: Regular, Style: Normal, Data: testData(1)},
			{Weight: Bold, Style: Normal, Data: testData(2)},
		},
	})

	// Request italic, but only normal exists. Should fall back to normal style.
	data, ok := reg.Resolve("NoItalic", Regular, Italic)
	if !ok {
		t.Fatal("expected resolve to succeed")
	}
	if data[0] != 1 {
		t.Errorf("expected regular normal data (1), got %d", data[0])
	}

	data, ok = reg.Resolve("NoItalic", Bold, Italic)
	if !ok {
		t.Fatal("expected resolve to succeed")
	}
	if data[0] != 2 {
		t.Errorf("expected bold normal data (2), got %d", data[0])
	}
}

func TestResolve_WeightCSS400(t *testing.T) {
	// CSS spec: weight 400 tries 500, then descending from 400, then ascending from 500.
	tests := []struct {
		name      string
		available []Weight
		wantIdx   int // index into available after sorting
		wantDesc  string
	}{
		{
			name:      "400 tries 500",
			available: []Weight{300, 500, 700},
			wantIdx:   1, // 500
			wantDesc:  "should prefer 500 for 400 request",
		},
		{
			name:      "400 descends if no 500",
			available: []Weight{200, 300, 700},
			wantIdx:   1, // 300
			wantDesc:  "should pick nearest lighter (300)",
		},
		{
			name:      "400 ascends if nothing lighter",
			available: []Weight{600, 800},
			wantIdx:   0, // 600
			wantDesc:  "should pick nearest heavier (600)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := NewRegistry()
			faces := make([]Face, len(tt.available))
			for i, w := range tt.available {
				faces[i] = Face{Weight: w, Style: Normal, Data: testData(byte(w / 100))}
			}
			reg.RegisterFamily(Family{Name: "T", Faces: faces})

			data, ok := reg.Resolve("T", Regular, Normal)
			if !ok {
				t.Fatal("expected resolve to succeed")
			}
			wantByte := byte(tt.available[tt.wantIdx] / 100)
			if data[0] != wantByte {
				t.Errorf("%s: got weight %d00, want weight %d00", tt.wantDesc, data[0], wantByte)
			}
		})
	}
}

func TestResolve_WeightCSS500(t *testing.T) {
	// CSS spec: weight 500 tries 400, then descending from 400, then ascending from 500.
	tests := []struct {
		name      string
		available []Weight
		wantByte  byte
	}{
		{
			name:      "500 tries 400",
			available: []Weight{300, 400, 700},
			wantByte:  4, // 400
		},
		{
			name:      "500 descends if no 400",
			available: []Weight{200, 300, 700},
			wantByte:  3, // 300
		},
		{
			name:      "500 ascends if nothing lighter",
			available: []Weight{600, 800},
			wantByte:  6, // 600
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := NewRegistry()
			faces := make([]Face, len(tt.available))
			for i, w := range tt.available {
				faces[i] = Face{Weight: w, Style: Normal, Data: testData(byte(w / 100))}
			}
			reg.RegisterFamily(Family{Name: "T", Faces: faces})

			data, ok := reg.Resolve("T", Medium, Normal)
			if !ok {
				t.Fatal("expected resolve to succeed")
			}
			if data[0] != tt.wantByte {
				t.Errorf("got weight %d00, want weight %d00", data[0], tt.wantByte)
			}
		})
	}
}

func TestResolve_WeightLightward(t *testing.T) {
	// Weights < 400: try lighter first, then heavier.
	tests := []struct {
		name      string
		target    Weight
		available []Weight
		wantByte  byte
	}{
		{
			name:      "300 finds 200 (lighter)",
			target:    Light,
			available: []Weight{200, 600},
			wantByte:  2,
		},
		{
			name:      "300 finds 600 (heavier, no lighter)",
			target:    Light,
			available: []Weight{600, 900},
			wantByte:  6,
		},
		{
			name:      "200 finds 100 (lighter)",
			target:    ExtraLight,
			available: []Weight{100, 700},
			wantByte:  1,
		},
		{
			name:      "100 finds 300 (heavier, nothing lighter)",
			target:    Thin,
			available: []Weight{300, 900},
			wantByte:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := NewRegistry()
			faces := make([]Face, len(tt.available))
			for i, w := range tt.available {
				faces[i] = Face{Weight: w, Style: Normal, Data: testData(byte(w / 100))}
			}
			reg.RegisterFamily(Family{Name: "T", Faces: faces})

			data, ok := reg.Resolve("T", tt.target, Normal)
			if !ok {
				t.Fatal("expected resolve to succeed")
			}
			if data[0] != tt.wantByte {
				t.Errorf("got weight %d00, want weight %d00", data[0], tt.wantByte)
			}
		})
	}
}

func TestResolve_WeightBoldward(t *testing.T) {
	// Weights > 500: try heavier first, then lighter.
	tests := []struct {
		name      string
		target    Weight
		available []Weight
		wantByte  byte
	}{
		{
			name:      "600 finds 700 (heavier)",
			target:    SemiBold,
			available: []Weight{400, 700},
			wantByte:  7,
		},
		{
			name:      "600 finds 400 (lighter, no heavier)",
			target:    SemiBold,
			available: []Weight{300, 400},
			wantByte:  4,
		},
		{
			name:      "700 finds 800 (heavier)",
			target:    Bold,
			available: []Weight{400, 800},
			wantByte:  8,
		},
		{
			name:      "900 finds 700 (lighter, nothing heavier)",
			target:    Black,
			available: []Weight{400, 700},
			wantByte:  7,
		},
		{
			name:      "800 finds 900 (heavier)",
			target:    ExtraBold,
			available: []Weight{400, 900},
			wantByte:  9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := NewRegistry()
			faces := make([]Face, len(tt.available))
			for i, w := range tt.available {
				faces[i] = Face{Weight: w, Style: Normal, Data: testData(byte(w / 100))}
			}
			reg.RegisterFamily(Family{Name: "T", Faces: faces})

			data, ok := reg.Resolve("T", tt.target, Normal)
			if !ok {
				t.Fatal("expected resolve to succeed")
			}
			if data[0] != tt.wantByte {
				t.Errorf("got weight %d00, want weight %d00", data[0], tt.wantByte)
			}
		})
	}
}

func TestResolve_SingleFaceFallback(t *testing.T) {
	// A family with only one face should always resolve to it.
	reg := NewRegistry()
	reg.RegisterFamily(Family{
		Name:  "OnlyRegular",
		Faces: []Face{{Weight: Regular, Style: Normal, Data: testData(42)}},
	})

	weights := []Weight{Thin, ExtraLight, Light, Regular, Medium, SemiBold, Bold, ExtraBold, Black}
	for _, w := range weights {
		for _, s := range []Style{Normal, Italic} {
			data, ok := reg.Resolve("OnlyRegular", w, s)
			if !ok {
				t.Errorf("expected resolve to succeed for weight=%d style=%s", w, s)
				continue
			}
			if data[0] != 42 {
				t.Errorf("weight=%d style=%s: expected data 42, got %d", w, s, data[0])
			}
		}
	}
}

func TestFamilyNames_Sorted(t *testing.T) {
	reg := NewRegistry()
	reg.RegisterFamily(Family{Name: "Zebra", Faces: []Face{{Weight: Regular, Style: Normal, Data: testData(1)}}})
	reg.RegisterFamily(Family{Name: "Alpha", Faces: []Face{{Weight: Regular, Style: Normal, Data: testData(2)}}})
	reg.RegisterFamily(Family{Name: "Middle", Faces: []Face{{Weight: Regular, Style: Normal, Data: testData(3)}}})

	names := reg.FamilyNames()
	if len(names) != 3 {
		t.Fatalf("expected 3 families, got %d", len(names))
	}
	if names[0] != "Alpha" || names[1] != "Middle" || names[2] != "Zebra" {
		t.Errorf("expected sorted names [Alpha Middle Zebra], got %v", names)
	}
}

func TestHasFamily(t *testing.T) {
	reg := NewRegistry()
	reg.RegisterFamily(Family{Name: "Inter", Faces: []Face{{Weight: Regular, Style: Normal, Data: testData(1)}}})

	if !reg.HasFamily("Inter") {
		t.Error("expected HasFamily(Inter) to be true")
	}
	if reg.HasFamily("Roboto") {
		t.Error("expected HasFamily(Roboto) to be false")
	}
}

func TestFaceCount(t *testing.T) {
	reg := NewRegistry()
	if reg.FaceCount("Missing") != 0 {
		t.Error("expected 0 for unregistered family")
	}

	reg.RegisterFamily(Family{
		Name: "Multi",
		Faces: []Face{
			{Weight: Regular, Style: Normal, Data: testData(1)},
			{Weight: Bold, Style: Normal, Data: testData(2)},
			{Weight: Regular, Style: Italic, Data: testData(3)},
		},
	})
	if reg.FaceCount("Multi") != 3 {
		t.Errorf("expected 3 faces, got %d", reg.FaceCount("Multi"))
	}
}

func TestResolve_EmptyFaces(t *testing.T) {
	reg := NewRegistry()
	reg.RegisterFamily(Family{Name: "Empty", Faces: nil})

	data, ok := reg.Resolve("Empty", Regular, Normal)
	if ok {
		t.Error("expected resolve to fail for family with no faces")
	}
	if data != nil {
		t.Error("expected nil data for family with no faces")
	}
}

func TestConcurrentAccess(t *testing.T) {
	reg := NewRegistry()
	reg.RegisterFamily(Family{
		Name: "Inter",
		Faces: []Face{
			{Weight: Regular, Style: Normal, Data: testData(1)},
			{Weight: Bold, Style: Normal, Data: testData(2)},
		},
	})

	var wg sync.WaitGroup
	const goroutines = 50
	const iterations = 100

	// Concurrent reads.
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_, _ = reg.Resolve("Inter", Regular, Normal)
				_ = reg.FamilyNames()
				_ = reg.HasFamily("Inter")
				_ = reg.FaceCount("Inter")
			}
		}()
	}

	// Concurrent writes.
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				reg.RegisterFamily(Family{
					Name:  "Dynamic",
					Faces: []Face{{Weight: Weight(idx*100 + 100), Style: Normal, Data: testData(byte(idx))}},
				})
			}
		}(i)
	}

	wg.Wait()
}

func TestResolve_MultipleStylesWeightResolution(t *testing.T) {
	// Verify weight resolution is scoped to the requested style.
	reg := NewRegistry()
	reg.RegisterFamily(Family{
		Name: "Mixed",
		Faces: []Face{
			{Weight: Regular, Style: Normal, Data: testData(10)},
			{Weight: Bold, Style: Normal, Data: testData(20)},
			{Weight: Light, Style: Italic, Data: testData(30)},
			{Weight: ExtraBold, Style: Italic, Data: testData(40)},
		},
	})

	// Request Medium Italic: has italic faces, so should resolve among italic weights.
	// Medium=500 > 500 in boldward path: nearest heavier italic is ExtraBold (800).
	data, ok := reg.Resolve("Mixed", Medium, Italic)
	if !ok {
		t.Fatal("expected resolve to succeed")
	}
	// Exact Medium Italic doesn't exist. Among italic faces: Light(300), ExtraBold(800).
	// CSS for 500: try 400 (none), descend from 400 -> Light(300), that's the best lighter.
	if data[0] != 30 {
		t.Errorf("expected italic light (30), got %d", data[0])
	}
}

func TestWeight_String(t *testing.T) {
	tests := []struct {
		weight Weight
		want   string
	}{
		{Thin, "Thin"},
		{ExtraLight, "ExtraLight"},
		{Light, "Light"},
		{Regular, "Regular"},
		{Medium, "Medium"},
		{SemiBold, "SemiBold"},
		{Bold, "Bold"},
		{ExtraBold, "ExtraBold"},
		{Black, "Black"},
		{Weight(150), "Weight(150)"},
		{Weight(0), "Weight(0)"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.weight.String(); got != tt.want {
				t.Errorf("Weight(%d).String() = %q, want %q", int(tt.weight), got, tt.want)
			}
		})
	}
}

func TestWeight_IsBoldIsLight(t *testing.T) {
	tests := []struct {
		weight Weight
		bold   bool
		light  bool
	}{
		{Thin, false, true},
		{Light, false, true},
		{Regular, false, false},
		{Medium, false, false},
		{SemiBold, false, false},
		{Bold, true, false},
		{ExtraBold, true, false},
		{Black, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.weight.String(), func(t *testing.T) {
			if got := tt.weight.IsBold(); got != tt.bold {
				t.Errorf("IsBold() = %v, want %v", got, tt.bold)
			}
			if got := tt.weight.IsLight(); got != tt.light {
				t.Errorf("IsLight() = %v, want %v", got, tt.light)
			}
		})
	}
}

func TestStyle_String(t *testing.T) {
	tests := []struct {
		style Style
		want  string
	}{
		{Normal, "Normal"},
		{Italic, "Italic"},
		{Style(99), "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.style.String(); got != tt.want {
				t.Errorf("Style(%d).String() = %q, want %q", tt.style, got, tt.want)
			}
		})
	}
}
