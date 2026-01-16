package theme

import "testing"

func TestDefaultRadii(t *testing.T) {
	r := DefaultRadii()

	expected := map[string]float32{
		"None": 0,
		"XS":   2,
		"S":    4,
		"M":    8,
		"L":    12,
		"XL":   16,
		"XXL":  24,
		"Full": 9999,
	}

	checks := []struct {
		name string
		got  float32
		want float32
	}{
		{"None", r.None, expected["None"]},
		{"XS", r.XS, expected["XS"]},
		{"S", r.S, expected["S"]},
		{"M", r.M, expected["M"]},
		{"L", r.L, expected["L"]},
		{"XL", r.XL, expected["XL"]},
		{"XXL", r.XXL, expected["XXL"]},
		{"Full", r.Full, expected["Full"]},
	}

	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
		}
	}
}

func TestRadiusScale_Scale(t *testing.T) {
	r := DefaultRadii()
	result := r.Scale(2)

	// None stays 0, Full stays 9999
	if result.None != 0 {
		t.Errorf("None should stay 0, got %v", result.None)
	}
	if result.Full != 9999 {
		t.Errorf("Full should stay 9999, got %v", result.Full)
	}

	// Others should be scaled
	if result.M != 16 {
		t.Errorf("M = %v, want 16 (8 * 2)", result.M)
	}
	if result.S != 8 {
		t.Errorf("S = %v, want 8 (4 * 2)", result.S)
	}

	// Original unchanged
	if r.M != 8 {
		t.Errorf("Original M changed: %v", r.M)
	}
}

func TestRadiusScale_Clamp(t *testing.T) {
	r := DefaultRadii()

	tests := []struct {
		name  string
		value float32
		min   float32
		max   float32
		want  float32
	}{
		{"value in range", 10, 0, 20, 10},
		{"value below min", -5, 0, 20, 0},
		{"value above max", 30, 0, 20, 20},
		{"value equals min", 0, 0, 20, 0},
		{"value equals max", 20, 0, 20, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := r.Clamp(tt.value, tt.min, tt.max); got != tt.want {
				t.Errorf("Clamp(%v, %v, %v) = %v, want %v",
					tt.value, tt.min, tt.max, got, tt.want)
			}
		})
	}
}

func TestSharpRadii(t *testing.T) {
	r := SharpRadii()

	// Sharp radii should be smaller than default
	d := DefaultRadii()
	if r.M >= d.M {
		t.Errorf("Sharp M (%v) should be smaller than default M (%v)", r.M, d.M)
	}

	// But None and Full should remain the same
	if r.None != 0 {
		t.Errorf("Sharp None = %v, want 0", r.None)
	}
	if r.Full != 9999 {
		t.Errorf("Sharp Full = %v, want 9999", r.Full)
	}
}

func TestSoftRadii(t *testing.T) {
	r := SoftRadii()

	// Soft radii should be larger than default
	d := DefaultRadii()
	if r.M <= d.M {
		t.Errorf("Soft M (%v) should be larger than default M (%v)", r.M, d.M)
	}

	// But None and Full should remain the same
	if r.None != 0 {
		t.Errorf("Soft None = %v, want 0", r.None)
	}
	if r.Full != 9999 {
		t.Errorf("Soft Full = %v, want 9999", r.Full)
	}
}

func TestUniform(t *testing.T) {
	r := Uniform(8)

	if r.TopLeft != 8 {
		t.Errorf("TopLeft = %v, want 8", r.TopLeft)
	}
	if r.TopRight != 8 {
		t.Errorf("TopRight = %v, want 8", r.TopRight)
	}
	if r.BottomRight != 8 {
		t.Errorf("BottomRight = %v, want 8", r.BottomRight)
	}
	if r.BottomLeft != 8 {
		t.Errorf("BottomLeft = %v, want 8", r.BottomLeft)
	}
}

func TestTop(t *testing.T) {
	r := Top(8)

	if r.TopLeft != 8 || r.TopRight != 8 {
		t.Error("Top corners should be 8")
	}
	if r.BottomLeft != 0 || r.BottomRight != 0 {
		t.Error("Bottom corners should be 0")
	}
}

func TestBottom(t *testing.T) {
	r := Bottom(8)

	if r.TopLeft != 0 || r.TopRight != 0 {
		t.Error("Top corners should be 0")
	}
	if r.BottomLeft != 8 || r.BottomRight != 8 {
		t.Error("Bottom corners should be 8")
	}
}

func TestLeft(t *testing.T) {
	r := Left(8)

	if r.TopLeft != 8 || r.BottomLeft != 8 {
		t.Error("Left corners should be 8")
	}
	if r.TopRight != 0 || r.BottomRight != 0 {
		t.Error("Right corners should be 0")
	}
}

func TestRight(t *testing.T) {
	r := Right(8)

	if r.TopLeft != 0 || r.BottomLeft != 0 {
		t.Error("Left corners should be 0")
	}
	if r.TopRight != 8 || r.BottomRight != 8 {
		t.Error("Right corners should be 8")
	}
}

func TestCornerRadius_IsUniform(t *testing.T) {
	tests := []struct {
		name   string
		radius CornerRadius
		want   bool
	}{
		{"uniform", Uniform(8), true},
		{"top only", Top(8), false},
		{"bottom only", Bottom(8), false},
		{"left only", Left(8), false},
		{"right only", Right(8), false},
		{"all zero is uniform", CornerRadius{0, 0, 0, 0}, true},
		{"mixed values", CornerRadius{1, 2, 3, 4}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.radius.IsUniform(); got != tt.want {
				t.Errorf("IsUniform() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCornerRadius_Max(t *testing.T) {
	tests := []struct {
		name   string
		radius CornerRadius
		want   float32
	}{
		{"uniform", Uniform(8), 8},
		{"top left is max", CornerRadius{10, 5, 3, 2}, 10},
		{"top right is max", CornerRadius{2, 10, 3, 5}, 10},
		{"bottom right is max", CornerRadius{2, 3, 10, 5}, 10},
		{"bottom left is max", CornerRadius{2, 3, 5, 10}, 10},
		{"all zero", CornerRadius{0, 0, 0, 0}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.radius.Max(); got != tt.want {
				t.Errorf("Max() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCornerRadius_Scale(t *testing.T) {
	r := Uniform(8)
	result := r.Scale(2)

	if result.TopLeft != 16 {
		t.Errorf("TopLeft = %v, want 16", result.TopLeft)
	}
	if result.TopRight != 16 {
		t.Errorf("TopRight = %v, want 16", result.TopRight)
	}
	if result.BottomRight != 16 {
		t.Errorf("BottomRight = %v, want 16", result.BottomRight)
	}
	if result.BottomLeft != 16 {
		t.Errorf("BottomLeft = %v, want 16", result.BottomLeft)
	}

	// Original unchanged
	if r.TopLeft != 8 {
		t.Errorf("Original TopLeft changed: %v", r.TopLeft)
	}
}

func TestCornerRadius_Scale_NonUniform(t *testing.T) {
	r := CornerRadius{1, 2, 3, 4}
	result := r.Scale(2)

	if result.TopLeft != 2 {
		t.Errorf("TopLeft = %v, want 2", result.TopLeft)
	}
	if result.TopRight != 4 {
		t.Errorf("TopRight = %v, want 4", result.TopRight)
	}
	if result.BottomRight != 6 {
		t.Errorf("BottomRight = %v, want 6", result.BottomRight)
	}
	if result.BottomLeft != 8 {
		t.Errorf("BottomLeft = %v, want 8", result.BottomLeft)
	}
}
