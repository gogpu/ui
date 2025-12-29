package geometry

import (
	"math"
	"testing"
)

func TestTight(t *testing.T) {
	c := Tight(Sz(100, 50))
	if c.MinWidth != 100 || c.MaxWidth != 100 || c.MinHeight != 50 || c.MaxHeight != 50 {
		t.Errorf("Tight() = %v, want min=max=100,50", c)
	}
}

func TestTightWidth(t *testing.T) {
	c := TightWidth(100)
	if c.MinWidth != 100 || c.MaxWidth != 100 {
		t.Errorf("TightWidth() width = %v,%v, want 100,100", c.MinWidth, c.MaxWidth)
	}
	if c.MinHeight != 0 || c.MaxHeight != Infinity {
		t.Errorf("TightWidth() height = %v,%v, want 0,Infinity", c.MinHeight, c.MaxHeight)
	}
}

func TestTightHeight(t *testing.T) {
	c := TightHeight(50)
	if c.MinWidth != 0 || c.MaxWidth != Infinity {
		t.Errorf("TightHeight() width = %v,%v, want 0,Infinity", c.MinWidth, c.MaxWidth)
	}
	if c.MinHeight != 50 || c.MaxHeight != 50 {
		t.Errorf("TightHeight() height = %v,%v, want 50,50", c.MinHeight, c.MaxHeight)
	}
}

func TestLoose(t *testing.T) {
	c := Loose(Sz(100, 50))
	if c.MinWidth != 0 || c.MaxWidth != 100 || c.MinHeight != 0 || c.MaxHeight != 50 {
		t.Errorf("Loose() = %v, want min=0, max=100,50", c)
	}
}

func TestExpand(t *testing.T) {
	c := Expand()
	if c.MinWidth != 0 || c.MinHeight != 0 {
		t.Errorf("Expand() min = %v,%v, want 0,0", c.MinWidth, c.MinHeight)
	}
	if c.MaxWidth < Infinity || c.MaxHeight < Infinity {
		t.Errorf("Expand() max = %v,%v, want Infinity,Infinity", c.MaxWidth, c.MaxHeight)
	}
}

func TestExpandWidth(t *testing.T) {
	c := ExpandWidth(100)
	if c.MaxWidth < Infinity {
		t.Errorf("ExpandWidth() maxWidth = %v, want Infinity", c.MaxWidth)
	}
	if c.MaxHeight != 100 {
		t.Errorf("ExpandWidth() maxHeight = %v, want 100", c.MaxHeight)
	}
}

func TestExpandHeight(t *testing.T) {
	c := ExpandHeight(100)
	if c.MaxWidth != 100 {
		t.Errorf("ExpandHeight() maxWidth = %v, want 100", c.MaxWidth)
	}
	if c.MaxHeight < Infinity {
		t.Errorf("ExpandHeight() maxHeight = %v, want Infinity", c.MaxHeight)
	}
}

func TestBoxConstraints(t *testing.T) {
	c := BoxConstraints(50, 200, 30, 100)
	if c.MinWidth != 50 || c.MaxWidth != 200 || c.MinHeight != 30 || c.MaxHeight != 100 {
		t.Errorf("BoxConstraints() = %v, want 50,200,30,100", c)
	}
}

func TestConstraints_Constrain(t *testing.T) {
	tests := []struct {
		name string
		c    Constraints
		size Size
		want Size
	}{
		{"within_bounds", BoxConstraints(50, 200, 30, 100), Sz(100, 50), Sz(100, 50)},
		{"below_min", BoxConstraints(50, 200, 30, 100), Sz(10, 10), Sz(50, 30)},
		{"above_max", BoxConstraints(50, 200, 30, 100), Sz(300, 200), Sz(200, 100)},
		{"mixed", BoxConstraints(50, 200, 30, 100), Sz(10, 200), Sz(50, 100)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.c.Constrain(tt.size)
			if got != tt.want {
				t.Errorf("Constrain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConstraints_ConstrainWidth(t *testing.T) {
	c := BoxConstraints(50, 200, 30, 100)
	tests := []struct {
		name  string
		width float32
		want  float32
	}{
		{"within", 100, 100},
		{"below", 10, 50},
		{"above", 300, 200},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.ConstrainWidth(tt.width)
			if got != tt.want {
				t.Errorf("ConstrainWidth(%v) = %v, want %v", tt.width, got, tt.want)
			}
		})
	}
}

func TestConstraints_ConstrainHeight(t *testing.T) {
	c := BoxConstraints(50, 200, 30, 100)
	tests := []struct {
		name   string
		height float32
		want   float32
	}{
		{"within", 50, 50},
		{"below", 10, 30},
		{"above", 200, 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.ConstrainHeight(tt.height)
			if got != tt.want {
				t.Errorf("ConstrainHeight(%v) = %v, want %v", tt.height, got, tt.want)
			}
		})
	}
}

func TestConstraints_Loosen(t *testing.T) {
	c := BoxConstraints(100, 200, 50, 100)
	got := c.Loosen()
	want := BoxConstraints(0, 200, 0, 100)
	if got != want {
		t.Errorf("Loosen() = %v, want %v", got, want)
	}
}

func TestConstraints_LoosenWidth(t *testing.T) {
	c := BoxConstraints(100, 200, 50, 100)
	got := c.LoosenWidth()
	want := BoxConstraints(0, 200, 50, 100)
	if got != want {
		t.Errorf("LoosenWidth() = %v, want %v", got, want)
	}
}

func TestConstraints_LoosenHeight(t *testing.T) {
	c := BoxConstraints(100, 200, 50, 100)
	got := c.LoosenHeight()
	want := BoxConstraints(100, 200, 0, 100)
	if got != want {
		t.Errorf("LoosenHeight() = %v, want %v", got, want)
	}
}

func TestConstraints_Tighten(t *testing.T) {
	tests := []struct {
		name string
		c    Constraints
		size Size
		want Constraints
	}{
		{
			"within_bounds",
			BoxConstraints(50, 200, 30, 100),
			Sz(150, 80),
			Tight(Sz(150, 80)),
		},
		{
			"clamped_to_max",
			BoxConstraints(50, 200, 30, 100),
			Sz(300, 200),
			Tight(Sz(200, 100)),
		},
		{
			"clamped_to_min",
			BoxConstraints(50, 200, 30, 100),
			Sz(10, 10),
			Tight(Sz(50, 30)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.c.Tighten(tt.size)
			if got != tt.want {
				t.Errorf("Tighten() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConstraints_TightenWidth(t *testing.T) {
	c := BoxConstraints(50, 200, 30, 100)
	got := c.TightenWidth(150)
	if got.MinWidth != 150 || got.MaxWidth != 150 {
		t.Errorf("TightenWidth() width = %v,%v, want 150,150", got.MinWidth, got.MaxWidth)
	}
	if got.MinHeight != 30 || got.MaxHeight != 100 {
		t.Errorf("TightenWidth() height = %v,%v, want 30,100", got.MinHeight, got.MaxHeight)
	}
}

func TestConstraints_TightenHeight(t *testing.T) {
	c := BoxConstraints(50, 200, 30, 100)
	got := c.TightenHeight(80)
	if got.MinWidth != 50 || got.MaxWidth != 200 {
		t.Errorf("TightenHeight() width = %v,%v, want 50,200", got.MinWidth, got.MaxWidth)
	}
	if got.MinHeight != 80 || got.MaxHeight != 80 {
		t.Errorf("TightenHeight() height = %v,%v, want 80,80", got.MinHeight, got.MaxHeight)
	}
}

func TestConstraints_Enforce(t *testing.T) {
	tests := []struct {
		name  string
		c     Constraints
		other Constraints
		want  Constraints
	}{
		{
			"more_restrictive",
			BoxConstraints(50, 200, 30, 100),
			BoxConstraints(100, 150, 50, 80),
			BoxConstraints(100, 150, 50, 80),
		},
		{
			"less_restrictive",
			BoxConstraints(100, 150, 50, 80),
			BoxConstraints(50, 200, 30, 100),
			BoxConstraints(100, 150, 50, 80),
		},
		{
			"mixed",
			BoxConstraints(50, 200, 30, 100),
			BoxConstraints(100, 300, 20, 80),
			BoxConstraints(100, 200, 30, 80),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.c.Enforce(tt.other)
			if got != tt.want {
				t.Errorf("Enforce() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConstraints_Deflate(t *testing.T) {
	tests := []struct {
		name   string
		c      Constraints
		insets Insets
		want   Constraints
	}{
		{
			"uniform",
			Tight(Sz(100, 100)),
			UniformInsets(10),
			Tight(Sz(80, 80)),
		},
		{
			"asymmetric",
			BoxConstraints(50, 200, 30, 100),
			InsetsLTRB(10, 5, 10, 5),
			BoxConstraints(30, 180, 20, 90),
		},
		{
			"larger_than_size",
			Tight(Sz(100, 100)),
			UniformInsets(60),
			Tight(Sz(0, 0)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.c.Deflate(tt.insets)
			if got != tt.want {
				t.Errorf("Deflate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConstraints_IsTight(t *testing.T) {
	tests := []struct {
		name string
		c    Constraints
		want bool
	}{
		{"tight", Tight(Sz(100, 50)), true},
		{"loose", Loose(Sz(100, 50)), false},
		{"tight_width_only", TightWidth(100), false},
		{"tight_height_only", TightHeight(50), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.c.IsTight()
			if got != tt.want {
				t.Errorf("IsTight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConstraints_IsTightWidth(t *testing.T) {
	tests := []struct {
		name string
		c    Constraints
		want bool
	}{
		{"tight", Tight(Sz(100, 50)), true},
		{"tight_width", TightWidth(100), true},
		{"loose", Loose(Sz(100, 50)), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.c.IsTightWidth()
			if got != tt.want {
				t.Errorf("IsTightWidth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConstraints_IsTightHeight(t *testing.T) {
	tests := []struct {
		name string
		c    Constraints
		want bool
	}{
		{"tight", Tight(Sz(100, 50)), true},
		{"tight_height", TightHeight(50), true},
		{"loose", Loose(Sz(100, 50)), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.c.IsTightHeight()
			if got != tt.want {
				t.Errorf("IsTightHeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConstraints_IsUnbounded(t *testing.T) {
	tests := []struct {
		name string
		c    Constraints
		want bool
	}{
		{"expand", Expand(), true},
		{"tight", Tight(Sz(100, 50)), false},
		{"bounded_width", ExpandHeight(100), false},
		{"bounded_height", ExpandWidth(100), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.c.IsUnbounded()
			if got != tt.want {
				t.Errorf("IsUnbounded() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConstraints_HasBoundedWidth(t *testing.T) {
	tests := []struct {
		name string
		c    Constraints
		want bool
	}{
		{"bounded", Loose(Sz(100, 50)), true},
		{"unbounded", Expand(), false},
		{"expand_width", ExpandWidth(100), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.c.HasBoundedWidth()
			if got != tt.want {
				t.Errorf("HasBoundedWidth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConstraints_HasBoundedHeight(t *testing.T) {
	tests := []struct {
		name string
		c    Constraints
		want bool
	}{
		{"bounded", Loose(Sz(100, 50)), true},
		{"unbounded", Expand(), false},
		{"expand_height", ExpandHeight(100), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.c.HasBoundedHeight()
			if got != tt.want {
				t.Errorf("HasBoundedHeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConstraints_IsSatisfiedBy(t *testing.T) {
	c := BoxConstraints(50, 200, 30, 100)
	tests := []struct {
		name string
		size Size
		want bool
	}{
		{"within", Sz(100, 50), true},
		{"at_min", Sz(50, 30), true},
		{"at_max", Sz(200, 100), true},
		{"below_min_width", Sz(40, 50), false},
		{"below_min_height", Sz(100, 20), false},
		{"above_max_width", Sz(210, 50), false},
		{"above_max_height", Sz(100, 110), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.IsSatisfiedBy(tt.size)
			if got != tt.want {
				t.Errorf("IsSatisfiedBy(%v) = %v, want %v", tt.size, got, tt.want)
			}
		})
	}
}

func TestConstraints_Normalize(t *testing.T) {
	tests := []struct {
		name string
		c    Constraints
		want Constraints
	}{
		{
			"already_normal",
			BoxConstraints(50, 200, 30, 100),
			BoxConstraints(50, 200, 30, 100),
		},
		{
			"min_exceeds_max",
			BoxConstraints(200, 100, 50, 30),
			BoxConstraints(100, 100, 30, 30),
		},
		{
			"negative_values",
			BoxConstraints(-50, 200, -30, 100),
			BoxConstraints(0, 200, 0, 100),
		},
		{
			"negative_max",
			BoxConstraints(0, -100, 0, -50),
			BoxConstraints(0, 0, 0, 0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.c.Normalize()
			if got != tt.want {
				t.Errorf("Normalize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConstraints_IsNormalized(t *testing.T) {
	tests := []struct {
		name string
		c    Constraints
		want bool
	}{
		{"normal", BoxConstraints(50, 200, 30, 100), true},
		{"min_exceeds_max", BoxConstraints(200, 100, 30, 100), false},
		{"negative_min", BoxConstraints(-50, 200, 30, 100), false},
		{"tight", Tight(Sz(100, 50)), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.c.IsNormalized()
			if got != tt.want {
				t.Errorf("IsNormalized() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConstraints_Smallest(t *testing.T) {
	c := BoxConstraints(50, 200, 30, 100)
	got := c.Smallest()
	want := Sz(50, 30)
	if got != want {
		t.Errorf("Smallest() = %v, want %v", got, want)
	}
}

func TestConstraints_Biggest(t *testing.T) {
	tests := []struct {
		name string
		c    Constraints
		want Size
	}{
		{"bounded", BoxConstraints(50, 200, 30, 100), Sz(200, 100)},
		{"unbounded", Expand(), Sz(0, 0)}, // Falls back to min
		{"unbounded_width", ExpandWidth(100), Sz(0, 100)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.c.Biggest()
			if got != tt.want {
				t.Errorf("Biggest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConstraints_BiggestFinite(t *testing.T) {
	tests := []struct {
		name     string
		c        Constraints
		fallback Size
		want     Size
	}{
		{"bounded", BoxConstraints(50, 200, 30, 100), Sz(500, 500), Sz(200, 100)},
		{"unbounded", Expand(), Sz(500, 500), Sz(500, 500)},
		{"mixed", ExpandWidth(100), Sz(500, 500), Sz(500, 100)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.c.BiggestFinite(tt.fallback.Width, tt.fallback.Height)
			if got != tt.want {
				t.Errorf("BiggestFinite() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConstraints_IsZero(t *testing.T) {
	tests := []struct {
		name string
		c    Constraints
		want bool
	}{
		{"zero", Constraints{}, true},
		{"non_zero", BoxConstraints(0, 100, 0, 100), false},
		{"all_zero", BoxConstraints(0, 0, 0, 0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.c.IsZero()
			if got != tt.want {
				t.Errorf("IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConstraints_IsNaN(t *testing.T) {
	nan := float32(math.NaN())
	tests := []struct {
		name string
		c    Constraints
		want bool
	}{
		{"normal", BoxConstraints(50, 200, 30, 100), false},
		{"nan_min_width", Constraints{MinWidth: nan, MaxWidth: 200, MinHeight: 30, MaxHeight: 100}, true},
		{"nan_max_height", Constraints{MinWidth: 50, MaxWidth: 200, MinHeight: 30, MaxHeight: nan}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.c.IsNaN()
			if got != tt.want {
				t.Errorf("IsNaN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConstraints_Sanitize(t *testing.T) {
	nan := float32(math.NaN())
	tests := []struct {
		name string
		c    Constraints
		want Constraints
	}{
		{
			"normal",
			BoxConstraints(50, 200, 30, 100),
			BoxConstraints(50, 200, 30, 100),
		},
		{
			"nan_values",
			Constraints{MinWidth: nan, MaxWidth: nan, MinHeight: nan, MaxHeight: nan},
			Constraints{MinWidth: 0, MaxWidth: Infinity, MinHeight: 0, MaxHeight: Infinity},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.c.Sanitize()
			if got != tt.want {
				t.Errorf("Sanitize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConstraints_String(t *testing.T) {
	tests := []struct {
		name string
		c    Constraints
		want string
	}{
		{"bounded", BoxConstraints(50, 200, 30, 100), "Constraints(50<=w<=200, 30<=h<=100)"},
		{"unbounded", Expand(), "Constraints(0<=w<=inf, 0<=h<=inf)"},
		{"tight", Tight(Sz(100, 50)), "Constraints(100<=w<=100, 50<=h<=50)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.c.String()
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test immutability
func TestConstraints_Immutability(t *testing.T) {
	original := BoxConstraints(50, 200, 30, 100)
	copyConstraints := original

	_ = original.Constrain(Sz(300, 300))
	_ = original.Loosen()
	_ = original.Tighten(Sz(100, 50))
	_ = original.Normalize()

	if original != copyConstraints {
		t.Errorf("Constraints operations mutated original: got %v, want %v", original, copyConstraints)
	}
}

// Benchmarks
func BenchmarkConstraints_Constrain(b *testing.B) {
	c := BoxConstraints(50, 200, 30, 100)
	s := Sz(150, 80)
	for i := 0; i < b.N; i++ {
		_ = c.Constrain(s)
	}
}

func BenchmarkConstraints_IsSatisfiedBy(b *testing.B) {
	c := BoxConstraints(50, 200, 30, 100)
	s := Sz(150, 80)
	for i := 0; i < b.N; i++ {
		_ = c.IsSatisfiedBy(s)
	}
}

func BenchmarkConstraints_Normalize(b *testing.B) {
	c := BoxConstraints(200, 100, 50, 30) // Invalid constraints
	for i := 0; i < b.N; i++ {
		_ = c.Normalize()
	}
}
