package theme

import "testing"

func TestDefaultSpacing(t *testing.T) {
	s := DefaultSpacing()

	expected := map[string]float32{
		"XXS":  2,
		"XS":   4,
		"S":    8,
		"M":    16,
		"L":    24,
		"XL":   32,
		"XXL":  48,
		"XXXL": 64,
	}

	checks := []struct {
		name string
		got  float32
		want float32
	}{
		{"XXS", s.XXS, expected["XXS"]},
		{"XS", s.XS, expected["XS"]},
		{"S", s.S, expected["S"]},
		{"M", s.M, expected["M"]},
		{"L", s.L, expected["L"]},
		{"XL", s.XL, expected["XL"]},
		{"XXL", s.XXL, expected["XXL"]},
		{"XXXL", s.XXXL, expected["XXXL"]},
	}

	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
		}
	}
}

func TestSpacingScale_Scale(t *testing.T) {
	s := DefaultSpacing()
	result := s.Scale(2)

	if result.M != 32 {
		t.Errorf("M = %v, want 32 (16 * 2)", result.M)
	}
	if result.S != 16 {
		t.Errorf("S = %v, want 16 (8 * 2)", result.S)
	}

	// Original unchanged
	if s.M != 16 {
		t.Errorf("Original M changed: %v", s.M)
	}
}

func TestSpacingScale_Inset(t *testing.T) {
	s := DefaultSpacing()
	top, right, bottom, left := s.Inset(16)

	if top != 16 || right != 16 || bottom != 16 || left != 16 {
		t.Errorf("Inset = (%v, %v, %v, %v), want (16, 16, 16, 16)", top, right, bottom, left)
	}
}

func TestSpacingScale_InsetHorizontal(t *testing.T) {
	s := DefaultSpacing()
	top, right, bottom, left := s.InsetHorizontal(16)

	if top != 0 || bottom != 0 {
		t.Errorf("Vertical should be 0, got top=%v bottom=%v", top, bottom)
	}
	if right != 16 || left != 16 {
		t.Errorf("Horizontal should be 16, got right=%v left=%v", right, left)
	}
}

func TestSpacingScale_InsetVertical(t *testing.T) {
	s := DefaultSpacing()
	top, right, bottom, left := s.InsetVertical(16)

	if top != 16 || bottom != 16 {
		t.Errorf("Vertical should be 16, got top=%v bottom=%v", top, bottom)
	}
	if right != 0 || left != 0 {
		t.Errorf("Horizontal should be 0, got right=%v left=%v", right, left)
	}
}

func TestSpacingScale_Compact(t *testing.T) {
	s := DefaultSpacing()
	result := s.Compact()

	expected := s.M * 0.75
	if result.M != expected {
		t.Errorf("Compact M = %v, want %v", result.M, expected)
	}
}

func TestSpacingScale_Relaxed(t *testing.T) {
	s := DefaultSpacing()
	result := s.Relaxed()

	expected := s.M * 1.5
	if result.M != expected {
		t.Errorf("Relaxed M = %v, want %v", result.M, expected)
	}
}

func TestDenseSpacing(t *testing.T) {
	s := DenseSpacing()

	if s.M != 8 {
		t.Errorf("Dense M = %v, want 8", s.M)
	}
	if s.S != 4 {
		t.Errorf("Dense S = %v, want 4", s.S)
	}
}

func TestComfortableSpacing(t *testing.T) {
	s := ComfortableSpacing()

	if s.M != 24 {
		t.Errorf("Comfortable M = %v, want 24", s.M)
	}
	if s.S != 12 {
		t.Errorf("Comfortable S = %v, want 12", s.S)
	}
}

func TestSpacingScale_AllValuesScaled(t *testing.T) {
	s := DefaultSpacing()
	factor := float32(2.0)
	result := s.Scale(factor)

	checks := []struct {
		name     string
		original float32
		scaled   float32
	}{
		{"XXS", s.XXS, result.XXS},
		{"XS", s.XS, result.XS},
		{"S", s.S, result.S},
		{"M", s.M, result.M},
		{"L", s.L, result.L},
		{"XL", s.XL, result.XL},
		{"XXL", s.XXL, result.XXL},
		{"XXXL", s.XXXL, result.XXXL},
	}

	for _, c := range checks {
		expected := c.original * factor
		if c.scaled != expected {
			t.Errorf("%s = %v, want %v", c.name, c.scaled, expected)
		}
	}
}
