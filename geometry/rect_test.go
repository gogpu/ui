package geometry

import (
	"math"
	"testing"
)

func TestNewRect(t *testing.T) {
	r := NewRect(10, 20, 100, 50)
	if r.Min.X != 10 || r.Min.Y != 20 || r.Max.X != 110 || r.Max.Y != 70 {
		t.Errorf("NewRect(10, 20, 100, 50) = %v, want Min(10,20) Max(110,70)", r)
	}
}

func TestFromPointSize(t *testing.T) {
	r := FromPointSize(Pt(10, 20), Sz(100, 50))
	want := NewRect(10, 20, 100, 50)
	if r != want {
		t.Errorf("FromPointSize() = %v, want %v", r, want)
	}
}

func TestFromCenter(t *testing.T) {
	tests := []struct {
		name   string
		center Point
		size   Size
		want   Rect
	}{
		{"centered_at_origin", Pt(0, 0), Sz(100, 50), Rect{Min: Pt(-50, -25), Max: Pt(50, 25)}},
		{"centered_at_50_50", Pt(50, 50), Sz(100, 50), Rect{Min: Pt(0, 25), Max: Pt(100, 75)}},
		{"zero_size", Pt(50, 50), Sz(0, 0), Rect{Min: Pt(50, 50), Max: Pt(50, 50)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FromCenter(tt.center, tt.size)
			if got != tt.want {
				t.Errorf("FromCenter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromMinMax(t *testing.T) {
	tests := []struct {
		name string
		p1   Point
		p2   Point
		want Rect
	}{
		{"normal_order", Pt(0, 0), Pt(100, 50), Rect{Min: Pt(0, 0), Max: Pt(100, 50)}},
		{"reversed_order", Pt(100, 50), Pt(0, 0), Rect{Min: Pt(0, 0), Max: Pt(100, 50)}},
		{"mixed", Pt(100, 0), Pt(0, 50), Rect{Min: Pt(0, 0), Max: Pt(100, 50)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FromMinMax(tt.p1, tt.p2)
			if got != tt.want {
				t.Errorf("FromMinMax() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_Size(t *testing.T) {
	tests := []struct {
		name string
		r    Rect
		want Size
	}{
		{"positive", NewRect(0, 0, 100, 50), Sz(100, 50)},
		{"zero", Rect{}, Sz(0, 0)},
		{"offset", NewRect(10, 20, 100, 50), Sz(100, 50)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.r.Size()
			if got != tt.want {
				t.Errorf("Rect.Size() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_Width(t *testing.T) {
	r := NewRect(10, 20, 100, 50)
	if got := r.Width(); got != 100 {
		t.Errorf("Rect.Width() = %v, want 100", got)
	}
}

func TestRect_Height(t *testing.T) {
	r := NewRect(10, 20, 100, 50)
	if got := r.Height(); got != 50 {
		t.Errorf("Rect.Height() = %v, want 50", got)
	}
}

func TestRect_Center(t *testing.T) {
	tests := []struct {
		name string
		r    Rect
		want Point
	}{
		{"positive", NewRect(0, 0, 100, 50), Pt(50, 25)},
		{"offset", NewRect(100, 100, 200, 100), Pt(200, 150)},
		{"zero", Rect{}, Pt(0, 0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.r.Center()
			if got != tt.want {
				t.Errorf("Rect.Center() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_Corners(t *testing.T) {
	r := NewRect(10, 20, 100, 50)

	if got := r.TopLeft(); got != Pt(10, 20) {
		t.Errorf("TopLeft() = %v, want Point(10, 20)", got)
	}
	if got := r.TopRight(); got != Pt(110, 20) {
		t.Errorf("TopRight() = %v, want Point(110, 20)", got)
	}
	if got := r.BottomLeft(); got != Pt(10, 70) {
		t.Errorf("BottomLeft() = %v, want Point(10, 70)", got)
	}
	if got := r.BottomRight(); got != Pt(110, 70) {
		t.Errorf("BottomRight() = %v, want Point(110, 70)", got)
	}
}

func TestRect_Contains(t *testing.T) {
	r := NewRect(0, 0, 100, 50)
	tests := []struct {
		name string
		p    Point
		want bool
	}{
		{"inside", Pt(50, 25), true},
		{"on_min_edge", Pt(0, 0), true},
		{"on_max_edge", Pt(100, 50), true},
		{"outside_right", Pt(150, 25), false},
		{"outside_left", Pt(-50, 25), false},
		{"outside_top", Pt(50, -25), false},
		{"outside_bottom", Pt(50, 75), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.Contains(tt.p)
			if got != tt.want {
				t.Errorf("Rect.Contains(%v) = %v, want %v", tt.p, got, tt.want)
			}
		})
	}
}

func TestRect_ContainsRect(t *testing.T) {
	outer := NewRect(0, 0, 100, 100)
	tests := []struct {
		name  string
		other Rect
		want  bool
	}{
		{"inside", NewRect(10, 10, 50, 50), true},
		{"exact", NewRect(0, 0, 100, 100), true},
		{"partially_outside", NewRect(50, 50, 100, 100), false},
		{"completely_outside", NewRect(200, 200, 50, 50), false},
		{"touching_edge", NewRect(0, 0, 50, 50), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := outer.ContainsRect(tt.other)
			if got != tt.want {
				t.Errorf("Rect.ContainsRect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_Intersects(t *testing.T) {
	r := NewRect(0, 0, 100, 100)
	tests := []struct {
		name  string
		other Rect
		want  bool
	}{
		{"overlapping", NewRect(50, 50, 100, 100), true},
		{"inside", NewRect(25, 25, 50, 50), true},
		{"outside_right", NewRect(150, 0, 50, 50), false},
		{"outside_left", NewRect(-100, 0, 50, 50), false},
		{"touching_edge", NewRect(100, 0, 50, 50), false}, // Touching but not overlapping
		{"same", NewRect(0, 0, 100, 100), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.Intersects(tt.other)
			if got != tt.want {
				t.Errorf("Rect.Intersects() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_Intersection(t *testing.T) {
	tests := []struct {
		name  string
		r     Rect
		other Rect
		want  Rect
	}{
		{
			"overlapping",
			NewRect(0, 0, 100, 100),
			NewRect(50, 50, 100, 100),
			Rect{Min: Pt(50, 50), Max: Pt(100, 100)},
		},
		{
			"no_overlap",
			NewRect(0, 0, 50, 50),
			NewRect(100, 100, 50, 50),
			Rect{}, // Empty
		},
		{
			"inside",
			NewRect(0, 0, 100, 100),
			NewRect(25, 25, 50, 50),
			NewRect(25, 25, 50, 50),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.r.Intersection(tt.other)
			if got != tt.want {
				t.Errorf("Rect.Intersection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_Union(t *testing.T) {
	tests := []struct {
		name  string
		r     Rect
		other Rect
		want  Rect
	}{
		{
			"overlapping",
			NewRect(0, 0, 100, 100),
			NewRect(50, 50, 100, 100),
			Rect{Min: Pt(0, 0), Max: Pt(150, 150)},
		},
		{
			"separate",
			NewRect(0, 0, 50, 50),
			NewRect(100, 100, 50, 50),
			Rect{Min: Pt(0, 0), Max: Pt(150, 150)},
		},
		{
			"with_empty",
			NewRect(0, 0, 100, 100),
			Rect{},
			NewRect(0, 0, 100, 100),
		},
		{
			"empty_with_valid",
			Rect{},
			NewRect(0, 0, 100, 100),
			NewRect(0, 0, 100, 100),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.r.Union(tt.other)
			if got != tt.want {
				t.Errorf("Rect.Union() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_Inset(t *testing.T) {
	tests := []struct {
		name   string
		r      Rect
		insets Insets
		want   Rect
	}{
		{
			"uniform",
			NewRect(0, 0, 100, 100),
			UniformInsets(10),
			Rect{Min: Pt(10, 10), Max: Pt(90, 90)},
		},
		{
			"asymmetric",
			NewRect(0, 0, 100, 100),
			InsetsLTRB(10, 20, 30, 40),
			Rect{Min: Pt(10, 20), Max: Pt(70, 60)},
		},
		{
			"negative_expands",
			NewRect(10, 10, 80, 80),
			UniformInsets(-10),
			Rect{Min: Pt(0, 0), Max: Pt(100, 100)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.r.Inset(tt.insets)
			if got != tt.want {
				t.Errorf("Rect.Inset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_Expand(t *testing.T) {
	tests := []struct {
		name  string
		r     Rect
		delta float32
		want  Rect
	}{
		{
			"positive",
			NewRect(10, 10, 80, 80),
			10,
			Rect{Min: Pt(0, 0), Max: Pt(100, 100)},
		},
		{
			"negative_shrinks",
			NewRect(0, 0, 100, 100),
			-10,
			Rect{Min: Pt(10, 10), Max: Pt(90, 90)},
		},
		{
			"zero",
			NewRect(0, 0, 100, 100),
			0,
			NewRect(0, 0, 100, 100),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.r.Expand(tt.delta)
			if got != tt.want {
				t.Errorf("Rect.Expand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_Translate(t *testing.T) {
	tests := []struct {
		name   string
		r      Rect
		offset Point
		want   Rect
	}{
		{
			"positive",
			NewRect(0, 0, 100, 50),
			Pt(10, 20),
			NewRect(10, 20, 100, 50),
		},
		{
			"negative",
			NewRect(50, 50, 100, 50),
			Pt(-50, -50),
			NewRect(0, 0, 100, 50),
		},
		{
			"zero",
			NewRect(0, 0, 100, 50),
			Pt(0, 0),
			NewRect(0, 0, 100, 50),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.r.Translate(tt.offset)
			if got != tt.want {
				t.Errorf("Rect.Translate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_TranslateXY(t *testing.T) {
	r := NewRect(0, 0, 100, 50)
	got := r.TranslateXY(10, 20)
	want := NewRect(10, 20, 100, 50)
	if got != want {
		t.Errorf("Rect.TranslateXY() = %v, want %v", got, want)
	}
}

func TestRect_WithSize(t *testing.T) {
	r := NewRect(10, 20, 100, 50)
	got := r.WithSize(Sz(200, 100))
	want := NewRect(10, 20, 200, 100)
	if got != want {
		t.Errorf("Rect.WithSize() = %v, want %v", got, want)
	}
}

func TestRect_WithCenter(t *testing.T) {
	r := NewRect(0, 0, 100, 50)
	got := r.WithCenter(Pt(100, 100))
	want := FromCenter(Pt(100, 100), Sz(100, 50))
	if got != want {
		t.Errorf("Rect.WithCenter() = %v, want %v", got, want)
	}
}

func TestRect_IsZero(t *testing.T) {
	tests := []struct {
		name string
		r    Rect
		want bool
	}{
		{"zero", Rect{}, true},
		{"non_zero", NewRect(0, 0, 100, 50), false},
		{"min_non_zero", Rect{Min: Pt(1, 0)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.r.IsZero()
			if got != tt.want {
				t.Errorf("Rect.IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_IsEmpty(t *testing.T) {
	tests := []struct {
		name string
		r    Rect
		want bool
	}{
		{"valid", NewRect(0, 0, 100, 50), false},
		{"zero_size", Rect{}, true},
		{"zero_width", NewRect(0, 0, 0, 50), true},
		{"zero_height", NewRect(0, 0, 100, 0), true},
		{"negative_width", Rect{Min: Pt(100, 0), Max: Pt(0, 50)}, true},
		{"negative_height", Rect{Min: Pt(0, 100), Max: Pt(100, 0)}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.r.IsEmpty()
			if got != tt.want {
				t.Errorf("Rect.IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_Area(t *testing.T) {
	tests := []struct {
		name string
		r    Rect
		want float32
	}{
		{"positive", NewRect(0, 0, 100, 50), 5000},
		{"empty", Rect{}, 0},
		{"invalid", Rect{Min: Pt(100, 0), Max: Pt(0, 50)}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.r.Area()
			if got != tt.want {
				t.Errorf("Rect.Area() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_Normalize(t *testing.T) {
	tests := []struct {
		name string
		r    Rect
		want Rect
	}{
		{
			"already_normal",
			Rect{Min: Pt(0, 0), Max: Pt(100, 50)},
			Rect{Min: Pt(0, 0), Max: Pt(100, 50)},
		},
		{
			"reversed",
			Rect{Min: Pt(100, 50), Max: Pt(0, 0)},
			Rect{Min: Pt(0, 0), Max: Pt(100, 50)},
		},
		{
			"mixed",
			Rect{Min: Pt(100, 0), Max: Pt(0, 50)},
			Rect{Min: Pt(0, 0), Max: Pt(100, 50)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.r.Normalize()
			if got != tt.want {
				t.Errorf("Rect.Normalize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_IsNaN(t *testing.T) {
	nan := float32(math.NaN())
	tests := []struct {
		name string
		r    Rect
		want bool
	}{
		{"normal", NewRect(0, 0, 100, 50), false},
		{"nan_min_x", Rect{Min: Point{X: nan, Y: 0}, Max: Pt(100, 50)}, true},
		{"nan_max_y", Rect{Min: Pt(0, 0), Max: Point{X: 100, Y: nan}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.r.IsNaN()
			if got != tt.want {
				t.Errorf("Rect.IsNaN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_Sanitize(t *testing.T) {
	nan := float32(math.NaN())
	tests := []struct {
		name string
		r    Rect
		want Rect
	}{
		{"normal", NewRect(0, 0, 100, 50), NewRect(0, 0, 100, 50)},
		{
			"nan_values",
			Rect{Min: Point{X: nan, Y: 0}, Max: Point{X: 100, Y: nan}},
			Rect{Min: Pt(0, 0), Max: Pt(100, 0)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.r.Sanitize()
			if got != tt.want {
				t.Errorf("Rect.Sanitize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_String(t *testing.T) {
	tests := []struct {
		name string
		r    Rect
		want string
	}{
		{"integer", NewRect(10, 20, 100, 50), "Rect(10, 20, 100x50)"},
		{"decimal", NewRect(10.5, 20.5, 100.5, 50.5), "Rect(10.5, 20.5, 100.5x50.5)"},
		{"zero", Rect{}, "Rect(0, 0, 0x0)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.r.String()
			if got != tt.want {
				t.Errorf("Rect.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test that operations don't mutate original
func TestRect_Immutability(t *testing.T) {
	original := NewRect(10, 20, 100, 50)
	copyRect := original

	_ = original.Inset(UniformInsets(5))
	_ = original.Expand(10)
	_ = original.Translate(Pt(10, 10))
	_ = original.WithSize(Sz(200, 100))
	_ = original.WithCenter(Pt(100, 100))

	if original != copyRect {
		t.Errorf("Rect operations mutated original: got %v, want %v", original, copyRect)
	}
}

// Benchmarks
func BenchmarkRect_Contains(b *testing.B) {
	r := NewRect(0, 0, 100, 100)
	p := Pt(50, 50)
	for i := 0; i < b.N; i++ {
		_ = r.Contains(p)
	}
}

func BenchmarkRect_Intersects(b *testing.B) {
	r1 := NewRect(0, 0, 100, 100)
	r2 := NewRect(50, 50, 100, 100)
	for i := 0; i < b.N; i++ {
		_ = r1.Intersects(r2)
	}
}

func BenchmarkRect_Intersection(b *testing.B) {
	r1 := NewRect(0, 0, 100, 100)
	r2 := NewRect(50, 50, 100, 100)
	for i := 0; i < b.N; i++ {
		_ = r1.Intersection(r2)
	}
}

func BenchmarkRect_Union(b *testing.B) {
	r1 := NewRect(0, 0, 100, 100)
	r2 := NewRect(50, 50, 100, 100)
	for i := 0; i < b.N; i++ {
		_ = r1.Union(r2)
	}
}

func BenchmarkRect_Inset(b *testing.B) {
	r := NewRect(0, 0, 100, 100)
	insets := UniformInsets(10)
	for i := 0; i < b.N; i++ {
		_ = r.Inset(insets)
	}
}
