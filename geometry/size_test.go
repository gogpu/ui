package geometry

import (
	"math"
	"testing"
)

func TestSz(t *testing.T) {
	s := Sz(100, 50)
	if s.Width != 100 || s.Height != 50 {
		t.Errorf("Sz(100, 50) = %v, want Size{100, 50}", s)
	}
}

func TestSize_Add(t *testing.T) {
	tests := []struct {
		name string
		s    Size
		q    Size
		want Size
	}{
		{"positive", Sz(100, 50), Sz(10, 20), Sz(110, 70)},
		{"negative", Sz(100, 50), Sz(-10, -20), Sz(90, 30)},
		{"zero", Sz(100, 50), Sz(0, 0), Sz(100, 50)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Add(tt.q)
			if got != tt.want {
				t.Errorf("Size.Add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSize_Sub(t *testing.T) {
	tests := []struct {
		name string
		s    Size
		q    Size
		want Size
	}{
		{"positive", Sz(100, 50), Sz(10, 20), Sz(90, 30)},
		{"negative", Sz(100, 50), Sz(-10, -20), Sz(110, 70)},
		{"zero", Sz(100, 50), Sz(0, 0), Sz(100, 50)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Sub(tt.q)
			if got != tt.want {
				t.Errorf("Size.Sub() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSize_Scale(t *testing.T) {
	tests := []struct {
		name   string
		s      Size
		scalar float32
		want   Size
	}{
		{"double", Sz(100, 50), 2, Sz(200, 100)},
		{"half", Sz(100, 50), 0.5, Sz(50, 25)},
		{"zero", Sz(100, 50), 0, Sz(0, 0)},
		{"negative", Sz(100, 50), -1, Sz(-100, -50)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Scale(tt.scalar)
			if got != tt.want {
				t.Errorf("Size.Scale() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSize_Area(t *testing.T) {
	tests := []struct {
		name string
		s    Size
		want float32
	}{
		{"positive", Sz(100, 50), 5000},
		{"zero_width", Sz(0, 50), 0},
		{"zero_height", Sz(100, 0), 0},
		{"negative", Sz(-10, 50), -500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Area()
			if got != tt.want {
				t.Errorf("Size.Area() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSize_IsZero(t *testing.T) {
	tests := []struct {
		name string
		s    Size
		want bool
	}{
		{"zero", Sz(0, 0), true},
		{"non_zero_width", Sz(1, 0), false},
		{"non_zero_height", Sz(0, 1), false},
		{"non_zero_both", Sz(1, 1), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.IsZero()
			if got != tt.want {
				t.Errorf("Size.IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSize_IsEmpty(t *testing.T) {
	tests := []struct {
		name string
		s    Size
		want bool
	}{
		{"positive", Sz(100, 50), false},
		{"zero_width", Sz(0, 50), true},
		{"zero_height", Sz(100, 0), true},
		{"negative_width", Sz(-1, 50), true},
		{"negative_height", Sz(100, -1), true},
		{"both_zero", Sz(0, 0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.IsEmpty()
			if got != tt.want {
				t.Errorf("Size.IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSize_Contains(t *testing.T) {
	tests := []struct {
		name string
		s    Size
		q    Size
		want bool
	}{
		{"contains", Sz(100, 50), Sz(80, 40), true},
		{"exact", Sz(100, 50), Sz(100, 50), true},
		{"too_wide", Sz(100, 50), Sz(110, 40), false},
		{"too_tall", Sz(100, 50), Sz(80, 60), false},
		{"zero_contains_zero", Sz(0, 0), Sz(0, 0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Contains(tt.q)
			if got != tt.want {
				t.Errorf("Size.Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSize_Expand(t *testing.T) {
	tests := []struct {
		name  string
		s     Size
		delta float32
		want  Size
	}{
		{"positive", Sz(100, 50), 10, Sz(120, 70)},
		{"negative", Sz(100, 50), -10, Sz(80, 30)},
		{"zero", Sz(100, 50), 0, Sz(100, 50)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Expand(tt.delta)
			if got != tt.want {
				t.Errorf("Size.Expand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSize_Contract(t *testing.T) {
	tests := []struct {
		name  string
		s     Size
		delta float32
		want  Size
	}{
		{"positive", Sz(100, 50), 10, Sz(80, 30)},
		{"negative", Sz(100, 50), -10, Sz(120, 70)},
		{"zero", Sz(100, 50), 0, Sz(100, 50)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Contract(tt.delta)
			if got != tt.want {
				t.Errorf("Size.Contract() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSize_Min(t *testing.T) {
	tests := []struct {
		name string
		s    Size
		q    Size
		want Size
	}{
		{"mixed", Sz(100, 30), Sz(80, 50), Sz(80, 30)},
		{"same", Sz(100, 50), Sz(100, 50), Sz(100, 50)},
		{"first_smaller", Sz(50, 25), Sz(100, 50), Sz(50, 25)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Min(tt.q)
			if got != tt.want {
				t.Errorf("Size.Min() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSize_Max(t *testing.T) {
	tests := []struct {
		name string
		s    Size
		q    Size
		want Size
	}{
		{"mixed", Sz(100, 30), Sz(80, 50), Sz(100, 50)},
		{"same", Sz(100, 50), Sz(100, 50), Sz(100, 50)},
		{"second_larger", Sz(50, 25), Sz(100, 50), Sz(100, 50)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Max(tt.q)
			if got != tt.want {
				t.Errorf("Size.Max() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSize_Clamp(t *testing.T) {
	tests := []struct {
		name string
		s    Size
		minS Size
		maxS Size
		want Size
	}{
		{"inside", Sz(75, 75), Sz(50, 50), Sz(100, 100), Sz(75, 75)},
		{"above_max", Sz(150, 150), Sz(50, 50), Sz(100, 100), Sz(100, 100)},
		{"below_min", Sz(25, 25), Sz(50, 50), Sz(100, 100), Sz(50, 50)},
		{"mixed", Sz(150, 25), Sz(50, 50), Sz(100, 100), Sz(100, 50)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Clamp(tt.minS, tt.maxS)
			if got != tt.want {
				t.Errorf("Size.Clamp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSize_ToPoint(t *testing.T) {
	s := Sz(100, 50)
	want := Pt(100, 50)
	got := s.ToPoint()
	if got != want {
		t.Errorf("Size.ToPoint() = %v, want %v", got, want)
	}
}

func TestSize_AspectRatio(t *testing.T) {
	tests := []struct {
		name string
		s    Size
		want float32
	}{
		{"2:1", Sz(200, 100), 2},
		{"1:2", Sz(100, 200), 0.5},
		{"square", Sz(100, 100), 1},
		{"zero_height", Sz(100, 0), 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.AspectRatio()
			if got != tt.want {
				t.Errorf("Size.AspectRatio() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSize_FitIn(t *testing.T) {
	tests := []struct {
		name string
		s    Size
		maxS Size
		want Size
	}{
		{"wider", Sz(200, 100), Sz(100, 100), Sz(100, 50)},
		{"taller", Sz(100, 200), Sz(100, 100), Sz(50, 100)},
		{"fits", Sz(50, 50), Sz(100, 100), Sz(50, 50)},
		{"empty_source", Sz(0, 100), Sz(100, 100), Sz(0, 0)},
		{"empty_target", Sz(100, 100), Sz(0, 0), Sz(0, 0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.FitIn(tt.maxS)
			if !sizeEquals(got, tt.want, 0.0001) {
				t.Errorf("Size.FitIn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSize_FillIn(t *testing.T) {
	tests := []struct {
		name string
		s    Size
		maxS Size
		want Size
	}{
		{"wider", Sz(200, 100), Sz(100, 100), Sz(200, 100)},
		{"taller", Sz(100, 200), Sz(100, 100), Sz(100, 200)},
		{"square", Sz(50, 50), Sz(100, 100), Sz(100, 100)},
		{"empty_source", Sz(0, 100), Sz(100, 100), Sz(0, 0)},
		{"empty_target", Sz(100, 100), Sz(0, 0), Sz(0, 0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.FillIn(tt.maxS)
			if !sizeEquals(got, tt.want, 0.0001) {
				t.Errorf("Size.FillIn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSize_IsNaN(t *testing.T) {
	nan := float32(math.NaN())
	tests := []struct {
		name string
		s    Size
		want bool
	}{
		{"normal", Sz(100, 50), false},
		{"nan_width", Size{Width: nan, Height: 50}, true},
		{"nan_height", Size{Width: 100, Height: nan}, true},
		{"nan_both", Size{Width: nan, Height: nan}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.IsNaN()
			if got != tt.want {
				t.Errorf("Size.IsNaN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSize_Sanitize(t *testing.T) {
	nan := float32(math.NaN())
	tests := []struct {
		name string
		s    Size
		want Size
	}{
		{"normal", Sz(100, 50), Sz(100, 50)},
		{"nan_width", Size{Width: nan, Height: 50}, Sz(0, 50)},
		{"nan_height", Size{Width: 100, Height: nan}, Sz(100, 0)},
		{"nan_both", Size{Width: nan, Height: nan}, Sz(0, 0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Sanitize()
			if got != tt.want {
				t.Errorf("Size.Sanitize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSize_String(t *testing.T) {
	tests := []struct {
		name string
		s    Size
		want string
	}{
		{"integer", Sz(100, 50), "Size(100, 50)"},
		{"decimal", Sz(100.5, 50.25), "Size(100.5, 50.25)"},
		{"zero", Sz(0, 0), "Size(0, 0)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.String()
			if got != tt.want {
				t.Errorf("Size.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test that operations don't mutate original
func TestSize_Immutability(t *testing.T) {
	original := Sz(100, 50)
	saved := original

	_ = original.Add(Sz(10, 10))
	_ = original.Sub(Sz(10, 10))
	_ = original.Scale(2)
	_ = original.Expand(10)
	_ = original.Contract(10)

	if original != saved {
		t.Errorf("Size operations mutated original: got %v, want %v", original, saved)
	}
}

// Benchmarks
func BenchmarkSize_Add(b *testing.B) {
	s1 := Sz(100, 50)
	s2 := Sz(10, 20)
	for i := 0; i < b.N; i++ {
		_ = s1.Add(s2)
	}
}

func BenchmarkSize_Area(b *testing.B) {
	s := Sz(100, 50)
	for i := 0; i < b.N; i++ {
		_ = s.Area()
	}
}

func BenchmarkSize_FitIn(b *testing.B) {
	s := Sz(200, 100)
	maxS := Sz(100, 100)
	for i := 0; i < b.N; i++ {
		_ = s.FitIn(maxS)
	}
}

// Helper function
func sizeEquals(a, b Size, epsilon float32) bool {
	return floatEquals(a.Width, b.Width, epsilon) && floatEquals(a.Height, b.Height, epsilon)
}
