package geometry

import (
	"math"
	"testing"
)

func TestPt(t *testing.T) {
	p := Pt(10, 20)
	if p.X != 10 || p.Y != 20 {
		t.Errorf("Pt(10, 20) = %v, want Point{10, 20}", p)
	}
}

func TestPoint_Add(t *testing.T) {
	tests := []struct {
		name string
		p    Point
		q    Point
		want Point
	}{
		{"positive", Pt(10, 20), Pt(5, 10), Pt(15, 30)},
		{"negative", Pt(10, 20), Pt(-5, -10), Pt(5, 10)},
		{"zero", Pt(10, 20), Pt(0, 0), Pt(10, 20)},
		{"both_zero", Pt(0, 0), Pt(0, 0), Pt(0, 0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.Add(tt.q)
			if got != tt.want {
				t.Errorf("Point.Add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPoint_Sub(t *testing.T) {
	tests := []struct {
		name string
		p    Point
		q    Point
		want Point
	}{
		{"positive", Pt(10, 20), Pt(5, 10), Pt(5, 10)},
		{"negative", Pt(10, 20), Pt(-5, -10), Pt(15, 30)},
		{"zero", Pt(10, 20), Pt(0, 0), Pt(10, 20)},
		{"self", Pt(10, 20), Pt(10, 20), Pt(0, 0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.Sub(tt.q)
			if got != tt.want {
				t.Errorf("Point.Sub() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPoint_Scale(t *testing.T) {
	tests := []struct {
		name   string
		p      Point
		scalar float32
		want   Point
	}{
		{"positive", Pt(10, 20), 2, Pt(20, 40)},
		{"negative", Pt(10, 20), -1, Pt(-10, -20)},
		{"zero", Pt(10, 20), 0, Pt(0, 0)},
		{"fraction", Pt(10, 20), 0.5, Pt(5, 10)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.Scale(tt.scalar)
			if got != tt.want {
				t.Errorf("Point.Scale() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPoint_Mul(t *testing.T) {
	tests := []struct {
		name string
		p    Point
		q    Point
		want Point
	}{
		{"positive", Pt(10, 20), Pt(2, 3), Pt(20, 60)},
		{"zero_x", Pt(10, 20), Pt(0, 3), Pt(0, 60)},
		{"negative", Pt(10, 20), Pt(-1, -1), Pt(-10, -20)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.Mul(tt.q)
			if got != tt.want {
				t.Errorf("Point.Mul() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPoint_Div(t *testing.T) {
	tests := []struct {
		name string
		p    Point
		q    Point
		want Point
	}{
		{"positive", Pt(10, 20), Pt(2, 4), Pt(5, 5)},
		{"div_by_zero_x", Pt(10, 20), Pt(0, 4), Pt(0, 5)},
		{"div_by_zero_y", Pt(10, 20), Pt(2, 0), Pt(5, 0)},
		{"div_by_zero_both", Pt(10, 20), Pt(0, 0), Pt(0, 0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.Div(tt.q)
			if got != tt.want {
				t.Errorf("Point.Div() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPoint_Distance(t *testing.T) {
	tests := []struct {
		name string
		p    Point
		q    Point
		want float32
	}{
		{"3-4-5", Pt(0, 0), Pt(3, 4), 5},
		{"same_point", Pt(5, 5), Pt(5, 5), 0},
		{"horizontal", Pt(0, 0), Pt(10, 0), 10},
		{"vertical", Pt(0, 0), Pt(0, 10), 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.Distance(tt.q)
			if !floatEquals(got, tt.want, 0.0001) {
				t.Errorf("Point.Distance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPoint_DistanceSquared(t *testing.T) {
	tests := []struct {
		name string
		p    Point
		q    Point
		want float32
	}{
		{"3-4-5", Pt(0, 0), Pt(3, 4), 25},
		{"same_point", Pt(5, 5), Pt(5, 5), 0},
		{"horizontal", Pt(0, 0), Pt(10, 0), 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.DistanceSquared(tt.q)
			if got != tt.want {
				t.Errorf("Point.DistanceSquared() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPoint_Lerp(t *testing.T) {
	tests := []struct {
		name string
		p    Point
		q    Point
		t    float32
		want Point
	}{
		{"start", Pt(0, 0), Pt(10, 20), 0, Pt(0, 0)},
		{"end", Pt(0, 0), Pt(10, 20), 1, Pt(10, 20)},
		{"middle", Pt(0, 0), Pt(10, 20), 0.5, Pt(5, 10)},
		{"quarter", Pt(0, 0), Pt(10, 20), 0.25, Pt(2.5, 5)},
		{"extrapolate", Pt(0, 0), Pt(10, 20), 2, Pt(20, 40)},
		{"negative_t", Pt(0, 0), Pt(10, 20), -1, Pt(-10, -20)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.Lerp(tt.q, tt.t)
			if got != tt.want {
				t.Errorf("Point.Lerp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPoint_Negate(t *testing.T) {
	tests := []struct {
		name string
		p    Point
		want Point
	}{
		{"positive", Pt(10, 20), Pt(-10, -20)},
		{"negative", Pt(-10, -20), Pt(10, 20)},
		{"zero", Pt(0, 0), Pt(0, 0)},
		{"mixed", Pt(10, -20), Pt(-10, 20)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.Negate()
			if got != tt.want {
				t.Errorf("Point.Negate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPoint_Normalize(t *testing.T) {
	tests := []struct {
		name string
		p    Point
		want Point
	}{
		{"3-4-5", Pt(3, 4), Pt(0.6, 0.8)},
		{"zero", Pt(0, 0), Pt(0, 0)},
		{"unit_x", Pt(1, 0), Pt(1, 0)},
		{"unit_y", Pt(0, 1), Pt(0, 1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.Normalize()
			if !pointEquals(got, tt.want, 0.0001) {
				t.Errorf("Point.Normalize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPoint_Length(t *testing.T) {
	tests := []struct {
		name string
		p    Point
		want float32
	}{
		{"3-4-5", Pt(3, 4), 5},
		{"zero", Pt(0, 0), 0},
		{"unit_x", Pt(1, 0), 1},
		{"unit_y", Pt(0, 1), 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.Length()
			if !floatEquals(got, tt.want, 0.0001) {
				t.Errorf("Point.Length() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPoint_LengthSquared(t *testing.T) {
	tests := []struct {
		name string
		p    Point
		want float32
	}{
		{"3-4-5", Pt(3, 4), 25},
		{"zero", Pt(0, 0), 0},
		{"unit_x", Pt(1, 0), 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.LengthSquared()
			if got != tt.want {
				t.Errorf("Point.LengthSquared() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPoint_Dot(t *testing.T) {
	tests := []struct {
		name string
		p    Point
		q    Point
		want float32
	}{
		{"positive", Pt(1, 2), Pt(3, 4), 11},
		{"perpendicular", Pt(1, 0), Pt(0, 1), 0},
		{"same", Pt(3, 4), Pt(3, 4), 25},
		{"negative", Pt(1, 2), Pt(-1, -2), -5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.Dot(tt.q)
			if got != tt.want {
				t.Errorf("Point.Dot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPoint_Min(t *testing.T) {
	tests := []struct {
		name string
		p    Point
		q    Point
		want Point
	}{
		{"mixed", Pt(10, 5), Pt(3, 20), Pt(3, 5)},
		{"same", Pt(10, 20), Pt(10, 20), Pt(10, 20)},
		{"negative", Pt(-10, -20), Pt(-5, -30), Pt(-10, -30)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.Min(tt.q)
			if got != tt.want {
				t.Errorf("Point.Min() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPoint_Max(t *testing.T) {
	tests := []struct {
		name string
		p    Point
		q    Point
		want Point
	}{
		{"mixed", Pt(10, 5), Pt(3, 20), Pt(10, 20)},
		{"same", Pt(10, 20), Pt(10, 20), Pt(10, 20)},
		{"negative", Pt(-10, -20), Pt(-5, -30), Pt(-5, -20)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.Max(tt.q)
			if got != tt.want {
				t.Errorf("Point.Max() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPoint_Clamp(t *testing.T) {
	tests := []struct {
		name string
		p    Point
		minP Point
		maxP Point
		want Point
	}{
		{"inside", Pt(5, 5), Pt(0, 0), Pt(10, 10), Pt(5, 5)},
		{"above_max", Pt(15, 15), Pt(0, 0), Pt(10, 10), Pt(10, 10)},
		{"below_min", Pt(-5, -5), Pt(0, 0), Pt(10, 10), Pt(0, 0)},
		{"mixed", Pt(15, -5), Pt(0, 0), Pt(10, 10), Pt(10, 0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.Clamp(tt.minP, tt.maxP)
			if got != tt.want {
				t.Errorf("Point.Clamp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPoint_IsZero(t *testing.T) {
	tests := []struct {
		name string
		p    Point
		want bool
	}{
		{"zero", Pt(0, 0), true},
		{"non_zero_x", Pt(1, 0), false},
		{"non_zero_y", Pt(0, 1), false},
		{"non_zero_both", Pt(1, 1), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.IsZero()
			if got != tt.want {
				t.Errorf("Point.IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPoint_IsNaN(t *testing.T) {
	nan := float32(math.NaN())
	tests := []struct {
		name string
		p    Point
		want bool
	}{
		{"normal", Pt(1, 2), false},
		{"nan_x", Point{X: nan, Y: 2}, true},
		{"nan_y", Point{X: 1, Y: nan}, true},
		{"nan_both", Point{X: nan, Y: nan}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.IsNaN()
			if got != tt.want {
				t.Errorf("Point.IsNaN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPoint_Sanitize(t *testing.T) {
	nan := float32(math.NaN())
	tests := []struct {
		name string
		p    Point
		want Point
	}{
		{"normal", Pt(1, 2), Pt(1, 2)},
		{"nan_x", Point{X: nan, Y: 2}, Pt(0, 2)},
		{"nan_y", Point{X: 1, Y: nan}, Pt(1, 0)},
		{"nan_both", Point{X: nan, Y: nan}, Pt(0, 0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.Sanitize()
			if got != tt.want {
				t.Errorf("Point.Sanitize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPoint_String(t *testing.T) {
	tests := []struct {
		name string
		p    Point
		want string
	}{
		{"integer", Pt(10, 20), "Point(10, 20)"},
		{"decimal", Pt(10.5, 20.25), "Point(10.5, 20.25)"},
		{"zero", Pt(0, 0), "Point(0, 0)"},
		{"negative", Pt(-10, -20), "Point(-10, -20)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.String()
			if got != tt.want {
				t.Errorf("Point.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test that operations don't mutate original
func TestPoint_Immutability(t *testing.T) {
	original := Pt(10, 20)
	saved := original

	_ = original.Add(Pt(5, 5))
	_ = original.Sub(Pt(5, 5))
	_ = original.Scale(2)
	_ = original.Negate()
	_ = original.Normalize()

	if original != saved {
		t.Errorf("Point operations mutated original: got %v, want %v", original, saved)
	}
}

// Benchmarks
func BenchmarkPoint_Add(b *testing.B) {
	p1 := Pt(10, 20)
	p2 := Pt(5, 10)
	for i := 0; i < b.N; i++ {
		_ = p1.Add(p2)
	}
}

func BenchmarkPoint_Distance(b *testing.B) {
	p1 := Pt(0, 0)
	p2 := Pt(3, 4)
	for i := 0; i < b.N; i++ {
		_ = p1.Distance(p2)
	}
}

func BenchmarkPoint_Normalize(b *testing.B) {
	p := Pt(3, 4)
	for i := 0; i < b.N; i++ {
		_ = p.Normalize()
	}
}

// Helper functions for tests
func floatEquals(a, b, epsilon float32) bool {
	return (a-b) < epsilon && (b-a) < epsilon
}

func pointEquals(a, b Point, epsilon float32) bool {
	return floatEquals(a.X, b.X, epsilon) && floatEquals(a.Y, b.Y, epsilon)
}
