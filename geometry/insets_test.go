package geometry

import (
	"math"
	"testing"
)

func TestUniformInsets(t *testing.T) {
	i := UniformInsets(16)
	if i.Top != 16 || i.Right != 16 || i.Bottom != 16 || i.Left != 16 {
		t.Errorf("UniformInsets(16) = %v, want all 16", i)
	}
}

func TestSymmetricInsets(t *testing.T) {
	i := SymmetricInsets(16, 8)
	if i.Top != 8 || i.Bottom != 8 {
		t.Errorf("SymmetricInsets() vertical = %v,%v, want 8,8", i.Top, i.Bottom)
	}
	if i.Left != 16 || i.Right != 16 {
		t.Errorf("SymmetricInsets() horizontal = %v,%v, want 16,16", i.Left, i.Right)
	}
}

func TestInsetsLTRB(t *testing.T) {
	i := InsetsLTRB(10, 20, 30, 40)
	if i.Left != 10 {
		t.Errorf("InsetsLTRB() left = %v, want 10", i.Left)
	}
	if i.Top != 20 {
		t.Errorf("InsetsLTRB() top = %v, want 20", i.Top)
	}
	if i.Right != 30 {
		t.Errorf("InsetsLTRB() right = %v, want 30", i.Right)
	}
	if i.Bottom != 40 {
		t.Errorf("InsetsLTRB() bottom = %v, want 40", i.Bottom)
	}
}

func TestInsetsTRBL(t *testing.T) {
	i := InsetsTRBL(20, 30, 40, 10)
	if i.Top != 20 {
		t.Errorf("InsetsTRBL() top = %v, want 20", i.Top)
	}
	if i.Right != 30 {
		t.Errorf("InsetsTRBL() right = %v, want 30", i.Right)
	}
	if i.Bottom != 40 {
		t.Errorf("InsetsTRBL() bottom = %v, want 40", i.Bottom)
	}
	if i.Left != 10 {
		t.Errorf("InsetsTRBL() left = %v, want 10", i.Left)
	}
}

func TestInsetsOnly(t *testing.T) {
	i := InsetsOnly(10, 0, 0, 0)
	if i.Top != 10 || i.Right != 0 || i.Bottom != 0 || i.Left != 0 {
		t.Errorf("InsetsOnly(10,0,0,0) = %v, want top=10 only", i)
	}
}

func TestInsets_Horizontal(t *testing.T) {
	tests := []struct {
		name string
		i    Insets
		want float32
	}{
		{"uniform", UniformInsets(16), 32},
		{"asymmetric", InsetsLTRB(10, 0, 20, 0), 30},
		{"zero", Insets{}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.i.Horizontal()
			if got != tt.want {
				t.Errorf("Horizontal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInsets_Vertical(t *testing.T) {
	tests := []struct {
		name string
		i    Insets
		want float32
	}{
		{"uniform", UniformInsets(16), 32},
		{"asymmetric", InsetsLTRB(0, 10, 0, 20), 30},
		{"zero", Insets{}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.i.Vertical()
			if got != tt.want {
				t.Errorf("Vertical() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInsets_Size(t *testing.T) {
	tests := []struct {
		name string
		i    Insets
		want Size
	}{
		{"uniform", UniformInsets(16), Sz(32, 32)},
		{"asymmetric", SymmetricInsets(16, 8), Sz(32, 16)},
		{"zero", Insets{}, Sz(0, 0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.i.Size()
			if got != tt.want {
				t.Errorf("Size() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInsets_TopLeft(t *testing.T) {
	i := InsetsLTRB(10, 20, 30, 40)
	got := i.TopLeft()
	want := Pt(10, 20)
	if got != want {
		t.Errorf("TopLeft() = %v, want %v", got, want)
	}
}

func TestInsets_BottomRight(t *testing.T) {
	i := InsetsLTRB(10, 20, 30, 40)
	got := i.BottomRight()
	want := Pt(30, 40)
	if got != want {
		t.Errorf("BottomRight() = %v, want %v", got, want)
	}
}

func TestInsets_Add(t *testing.T) {
	tests := []struct {
		name  string
		i     Insets
		other Insets
		want  Insets
	}{
		{"uniform", UniformInsets(10), UniformInsets(5), UniformInsets(15)},
		{"asymmetric", InsetsLTRB(10, 20, 30, 40), InsetsLTRB(1, 2, 3, 4), InsetsLTRB(11, 22, 33, 44)},
		{"zero", UniformInsets(10), Insets{}, UniformInsets(10)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.i.Add(tt.other)
			if got != tt.want {
				t.Errorf("Add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInsets_Sub(t *testing.T) {
	tests := []struct {
		name  string
		i     Insets
		other Insets
		want  Insets
	}{
		{"uniform", UniformInsets(10), UniformInsets(5), UniformInsets(5)},
		{"asymmetric", InsetsLTRB(10, 20, 30, 40), InsetsLTRB(1, 2, 3, 4), InsetsLTRB(9, 18, 27, 36)},
		{"zero", UniformInsets(10), Insets{}, UniformInsets(10)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.i.Sub(tt.other)
			if got != tt.want {
				t.Errorf("Sub() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInsets_Scale(t *testing.T) {
	tests := []struct {
		name   string
		i      Insets
		scalar float32
		want   Insets
	}{
		{"double", UniformInsets(10), 2, UniformInsets(20)},
		{"half", UniformInsets(10), 0.5, UniformInsets(5)},
		{"zero", UniformInsets(10), 0, Insets{}},
		{"negative", UniformInsets(10), -1, UniformInsets(-10)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.i.Scale(tt.scalar)
			if got != tt.want {
				t.Errorf("Scale() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInsets_Negate(t *testing.T) {
	tests := []struct {
		name string
		i    Insets
		want Insets
	}{
		{"positive", UniformInsets(10), UniformInsets(-10)},
		{"negative", UniformInsets(-10), UniformInsets(10)},
		{"zero", Insets{}, Insets{}},
		{"mixed", InsetsLTRB(10, -20, 30, -40), InsetsLTRB(-10, 20, -30, 40)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.i.Negate()
			if got != tt.want {
				t.Errorf("Negate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInsets_Min(t *testing.T) {
	tests := []struct {
		name  string
		i     Insets
		other Insets
		want  Insets
	}{
		{"same", UniformInsets(10), UniformInsets(10), UniformInsets(10)},
		{"first_smaller", UniformInsets(5), UniformInsets(10), UniformInsets(5)},
		{"second_smaller", UniformInsets(10), UniformInsets(5), UniformInsets(5)},
		{"mixed", InsetsLTRB(5, 20, 15, 8), InsetsLTRB(10, 10, 10, 10), InsetsLTRB(5, 10, 10, 8)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.i.Min(tt.other)
			if got != tt.want {
				t.Errorf("Min() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInsets_Max(t *testing.T) {
	tests := []struct {
		name  string
		i     Insets
		other Insets
		want  Insets
	}{
		{"same", UniformInsets(10), UniformInsets(10), UniformInsets(10)},
		{"first_larger", UniformInsets(15), UniformInsets(10), UniformInsets(15)},
		{"second_larger", UniformInsets(10), UniformInsets(15), UniformInsets(15)},
		{"mixed", InsetsLTRB(5, 20, 15, 8), InsetsLTRB(10, 10, 10, 10), InsetsLTRB(10, 20, 15, 10)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.i.Max(tt.other)
			if got != tt.want {
				t.Errorf("Max() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInsets_Clamp(t *testing.T) {
	tests := []struct {
		name string
		i    Insets
		minI Insets
		maxI Insets
		want Insets
	}{
		{"within", UniformInsets(10), UniformInsets(5), UniformInsets(15), UniformInsets(10)},
		{"below_min", UniformInsets(3), UniformInsets(5), UniformInsets(15), UniformInsets(5)},
		{"above_max", UniformInsets(20), UniformInsets(5), UniformInsets(15), UniformInsets(15)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.i.Clamp(tt.minI, tt.maxI)
			if got != tt.want {
				t.Errorf("Clamp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInsets_IsZero(t *testing.T) {
	tests := []struct {
		name string
		i    Insets
		want bool
	}{
		{"zero", Insets{}, true},
		{"non_zero_top", InsetsOnly(1, 0, 0, 0), false},
		{"non_zero_right", InsetsOnly(0, 1, 0, 0), false},
		{"non_zero_bottom", InsetsOnly(0, 0, 1, 0), false},
		{"non_zero_left", InsetsOnly(0, 0, 0, 1), false},
		{"uniform", UniformInsets(10), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.i.IsZero()
			if got != tt.want {
				t.Errorf("IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInsets_IsUniform(t *testing.T) {
	tests := []struct {
		name string
		i    Insets
		want bool
	}{
		{"uniform", UniformInsets(10), true},
		{"zero", Insets{}, true},
		{"asymmetric", SymmetricInsets(10, 20), false},
		{"different", InsetsLTRB(1, 2, 3, 4), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.i.IsUniform()
			if got != tt.want {
				t.Errorf("IsUniform() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInsets_IsSymmetric(t *testing.T) {
	tests := []struct {
		name string
		i    Insets
		want bool
	}{
		{"symmetric", SymmetricInsets(10, 20), true},
		{"uniform", UniformInsets(10), true},
		{"asymmetric", InsetsLTRB(10, 20, 30, 40), false},
		{"symmetric_left_right", InsetsLTRB(10, 20, 10, 40), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.i.IsSymmetric()
			if got != tt.want {
				t.Errorf("IsSymmetric() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInsets_IsNonNegative(t *testing.T) {
	tests := []struct {
		name string
		i    Insets
		want bool
	}{
		{"positive", UniformInsets(10), true},
		{"zero", Insets{}, true},
		{"negative_top", InsetsOnly(-1, 0, 0, 0), false},
		{"negative_right", InsetsOnly(0, -1, 0, 0), false},
		{"negative_bottom", InsetsOnly(0, 0, -1, 0), false},
		{"negative_left", InsetsOnly(0, 0, 0, -1), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.i.IsNonNegative()
			if got != tt.want {
				t.Errorf("IsNonNegative() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInsets_IsNaN(t *testing.T) {
	nan := float32(math.NaN())
	tests := []struct {
		name string
		i    Insets
		want bool
	}{
		{"normal", UniformInsets(10), false},
		{"nan_top", Insets{Top: nan, Right: 0, Bottom: 0, Left: 0}, true},
		{"nan_right", Insets{Top: 0, Right: nan, Bottom: 0, Left: 0}, true},
		{"nan_bottom", Insets{Top: 0, Right: 0, Bottom: nan, Left: 0}, true},
		{"nan_left", Insets{Top: 0, Right: 0, Bottom: 0, Left: nan}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.i.IsNaN()
			if got != tt.want {
				t.Errorf("IsNaN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInsets_Sanitize(t *testing.T) {
	nan := float32(math.NaN())
	tests := []struct {
		name string
		i    Insets
		want Insets
	}{
		{"normal", UniformInsets(10), UniformInsets(10)},
		{"nan_values", Insets{Top: nan, Right: nan, Bottom: nan, Left: nan}, Insets{}},
		{"partial_nan", Insets{Top: nan, Right: 10, Bottom: nan, Left: 20}, Insets{Top: 0, Right: 10, Bottom: 0, Left: 20}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.i.Sanitize()
			if got != tt.want {
				t.Errorf("Sanitize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInsets_Abs(t *testing.T) {
	tests := []struct {
		name string
		i    Insets
		want Insets
	}{
		{"positive", UniformInsets(10), UniformInsets(10)},
		{"negative", UniformInsets(-10), UniformInsets(10)},
		{"mixed", InsetsLTRB(-10, 20, -30, 40), InsetsLTRB(10, 20, 30, 40)},
		{"zero", Insets{}, Insets{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.i.Abs()
			if got != tt.want {
				t.Errorf("Abs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInsets_String(t *testing.T) {
	tests := []struct {
		name string
		i    Insets
		want string
	}{
		{"uniform", UniformInsets(16), "Insets(16, 16, 16, 16)"},
		{"asymmetric", InsetsLTRB(10, 20, 30, 40), "Insets(20, 30, 40, 10)"},
		{"zero", Insets{}, "Insets(0, 0, 0, 0)"},
		{"decimal", Insets{Top: 10.5, Right: 20.25, Bottom: 30.5, Left: 40.25}, "Insets(10.5, 20.25, 30.5, 40.25)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.i.String()
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test immutability
func TestInsets_Immutability(t *testing.T) {
	original := UniformInsets(10)
	copyInsets := original

	_ = original.Add(UniformInsets(5))
	_ = original.Sub(UniformInsets(5))
	_ = original.Scale(2)
	_ = original.Negate()
	_ = original.Abs()

	if original != copyInsets {
		t.Errorf("Insets operations mutated original: got %v, want %v", original, copyInsets)
	}
}

// Benchmarks
func BenchmarkInsets_Add(b *testing.B) {
	i1 := UniformInsets(10)
	i2 := UniformInsets(5)
	for i := 0; i < b.N; i++ {
		_ = i1.Add(i2)
	}
}

func BenchmarkInsets_Horizontal(b *testing.B) {
	i := SymmetricInsets(16, 8)
	for j := 0; j < b.N; j++ {
		_ = i.Horizontal()
	}
}

func BenchmarkInsets_Size(b *testing.B) {
	i := UniformInsets(16)
	for j := 0; j < b.N; j++ {
		_ = i.Size()
	}
}

func BenchmarkInsets_IsUniform(b *testing.B) {
	i := UniformInsets(10)
	for j := 0; j < b.N; j++ {
		_ = i.IsUniform()
	}
}
