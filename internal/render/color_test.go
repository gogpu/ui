package render

import (
	"testing"

	"github.com/gogpu/gg"
	"github.com/gogpu/ui/widget"
)

func TestToGGColor(t *testing.T) {
	tests := []struct {
		name  string
		input widget.Color
		want  gg.RGBA
	}{
		{
			name:  "black",
			input: widget.Color{R: 0, G: 0, B: 0, A: 1},
			want:  gg.RGBA{R: 0, G: 0, B: 0, A: 1},
		},
		{
			name:  "white",
			input: widget.Color{R: 1, G: 1, B: 1, A: 1},
			want:  gg.RGBA{R: 1, G: 1, B: 1, A: 1},
		},
		{
			name:  "red",
			input: widget.Color{R: 1, G: 0, B: 0, A: 1},
			want:  gg.RGBA{R: 1, G: 0, B: 0, A: 1},
		},
		{
			name:  "semi-transparent green",
			input: widget.Color{R: 0, G: 1, B: 0, A: 0.5},
			want:  gg.RGBA{R: 0, G: 1, B: 0, A: 0.5},
		},
		{
			name:  "transparent",
			input: widget.Color{R: 0, G: 0, B: 0, A: 0},
			want:  gg.RGBA{R: 0, G: 0, B: 0, A: 0},
		},
		{
			name:  "arbitrary color",
			input: widget.Color{R: 0.25, G: 0.5, B: 0.75, A: 0.9},
			want:  gg.RGBA{R: 0.25, G: 0.5, B: 0.75, A: 0.9},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToGGColor(tt.input)
			if !float64Equal(got.R, tt.want.R) {
				t.Errorf("R = %v, want %v", got.R, tt.want.R)
			}
			if !float64Equal(got.G, tt.want.G) {
				t.Errorf("G = %v, want %v", got.G, tt.want.G)
			}
			if !float64Equal(got.B, tt.want.B) {
				t.Errorf("B = %v, want %v", got.B, tt.want.B)
			}
			if !float64Equal(got.A, tt.want.A) {
				t.Errorf("A = %v, want %v", got.A, tt.want.A)
			}
		})
	}
}

func TestFromGGColor(t *testing.T) {
	tests := []struct {
		name  string
		input gg.RGBA
		want  widget.Color
	}{
		{
			name:  "black",
			input: gg.RGBA{R: 0, G: 0, B: 0, A: 1},
			want:  widget.Color{R: 0, G: 0, B: 0, A: 1},
		},
		{
			name:  "white",
			input: gg.RGBA{R: 1, G: 1, B: 1, A: 1},
			want:  widget.Color{R: 1, G: 1, B: 1, A: 1},
		},
		{
			name:  "arbitrary color",
			input: gg.RGBA{R: 0.25, G: 0.5, B: 0.75, A: 0.9},
			want:  widget.Color{R: 0.25, G: 0.5, B: 0.75, A: 0.9},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FromGGColor(tt.input)
			if !float32Equal(got.R, tt.want.R) {
				t.Errorf("R = %v, want %v", got.R, tt.want.R)
			}
			if !float32Equal(got.G, tt.want.G) {
				t.Errorf("G = %v, want %v", got.G, tt.want.G)
			}
			if !float32Equal(got.B, tt.want.B) {
				t.Errorf("B = %v, want %v", got.B, tt.want.B)
			}
			if !float32Equal(got.A, tt.want.A) {
				t.Errorf("A = %v, want %v", got.A, tt.want.A)
			}
		})
	}
}

func TestToGGColorPremultiplied(t *testing.T) {
	tests := []struct {
		name  string
		input widget.Color
		want  gg.RGBA
	}{
		{
			name:  "opaque red",
			input: widget.Color{R: 1, G: 0, B: 0, A: 1},
			want:  gg.RGBA{R: 1, G: 0, B: 0, A: 1},
		},
		{
			name:  "semi-transparent red",
			input: widget.Color{R: 1, G: 0, B: 0, A: 0.5},
			want:  gg.RGBA{R: 0.5, G: 0, B: 0, A: 0.5},
		},
		{
			name:  "transparent",
			input: widget.Color{R: 1, G: 1, B: 1, A: 0},
			want:  gg.RGBA{R: 0, G: 0, B: 0, A: 0},
		},
		{
			name:  "quarter alpha",
			input: widget.Color{R: 0.8, G: 0.4, B: 0.2, A: 0.25},
			want:  gg.RGBA{R: 0.2, G: 0.1, B: 0.05, A: 0.25},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToGGColorPremultiplied(tt.input)
			if !float64Equal(got.R, tt.want.R) {
				t.Errorf("R = %v, want %v", got.R, tt.want.R)
			}
			if !float64Equal(got.G, tt.want.G) {
				t.Errorf("G = %v, want %v", got.G, tt.want.G)
			}
			if !float64Equal(got.B, tt.want.B) {
				t.Errorf("B = %v, want %v", got.B, tt.want.B)
			}
			if !float64Equal(got.A, tt.want.A) {
				t.Errorf("A = %v, want %v", got.A, tt.want.A)
			}
		})
	}
}

func TestClamp01Float64(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  float64
	}{
		{"zero", 0, 0},
		{"one", 1, 1},
		{"middle", 0.5, 0.5},
		{"negative", -0.5, 0},
		{"over one", 1.5, 1},
		{"large negative", -100, 0},
		{"large positive", 100, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Clamp01Float64(tt.value)
			if got != tt.want {
				t.Errorf("Clamp01Float64(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestClamp01Float32(t *testing.T) {
	tests := []struct {
		name  string
		value float32
		want  float32
	}{
		{"zero", 0, 0},
		{"one", 1, 1},
		{"middle", 0.5, 0.5},
		{"negative", -0.5, 0},
		{"over one", 1.5, 1},
		{"large negative", -100, 0},
		{"large positive", 100, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Clamp01Float32(tt.value)
			if got != tt.want {
				t.Errorf("Clamp01Float32(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestLerpColor(t *testing.T) {
	tests := []struct {
		name string
		a    gg.RGBA
		b    gg.RGBA
		t    float64
		want gg.RGBA
	}{
		{
			name: "t=0 (start)",
			a:    gg.RGBA{R: 1, G: 0, B: 0, A: 1},
			b:    gg.RGBA{R: 0, G: 0, B: 1, A: 1},
			t:    0,
			want: gg.RGBA{R: 1, G: 0, B: 0, A: 1},
		},
		{
			name: "t=1 (end)",
			a:    gg.RGBA{R: 1, G: 0, B: 0, A: 1},
			b:    gg.RGBA{R: 0, G: 0, B: 1, A: 1},
			t:    1,
			want: gg.RGBA{R: 0, G: 0, B: 1, A: 1},
		},
		{
			name: "t=0.5 (middle)",
			a:    gg.RGBA{R: 0, G: 0, B: 0, A: 0},
			b:    gg.RGBA{R: 1, G: 1, B: 1, A: 1},
			t:    0.5,
			want: gg.RGBA{R: 0.5, G: 0.5, B: 0.5, A: 0.5},
		},
		{
			name: "t=0.25",
			a:    gg.RGBA{R: 0, G: 0, B: 0, A: 1},
			b:    gg.RGBA{R: 1, G: 1, B: 1, A: 1},
			t:    0.25,
			want: gg.RGBA{R: 0.25, G: 0.25, B: 0.25, A: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LerpColor(tt.a, tt.b, tt.t)
			if !float64Equal(got.R, tt.want.R) {
				t.Errorf("R = %v, want %v", got.R, tt.want.R)
			}
			if !float64Equal(got.G, tt.want.G) {
				t.Errorf("G = %v, want %v", got.G, tt.want.G)
			}
			if !float64Equal(got.B, tt.want.B) {
				t.Errorf("B = %v, want %v", got.B, tt.want.B)
			}
			if !float64Equal(got.A, tt.want.A) {
				t.Errorf("A = %v, want %v", got.A, tt.want.A)
			}
		})
	}
}

func TestColorRoundTrip(t *testing.T) {
	// Test that converting to gg and back preserves values
	original := widget.Color{R: 0.25, G: 0.5, B: 0.75, A: 0.9}
	ggColor := ToGGColor(original)
	roundTrip := FromGGColor(ggColor)

	if !float32Equal(roundTrip.R, original.R) {
		t.Errorf("R roundtrip: got %v, want %v", roundTrip.R, original.R)
	}
	if !float32Equal(roundTrip.G, original.G) {
		t.Errorf("G roundtrip: got %v, want %v", roundTrip.G, original.G)
	}
	if !float32Equal(roundTrip.B, original.B) {
		t.Errorf("B roundtrip: got %v, want %v", roundTrip.B, original.B)
	}
	if !float32Equal(roundTrip.A, original.A) {
		t.Errorf("A roundtrip: got %v, want %v", roundTrip.A, original.A)
	}
}

// Helper functions for float comparison
func float64Equal(a, b float64) bool {
	const epsilon = 0.0001
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < epsilon
}

func float32Equal(a, b float32) bool {
	const epsilon = 0.0001
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < epsilon
}

// Benchmarks

func BenchmarkToGGColor(b *testing.B) {
	c := widget.Color{R: 0.5, G: 0.5, B: 0.5, A: 1.0}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ToGGColor(c)
	}
}

func BenchmarkFromGGColor(b *testing.B) {
	c := gg.RGBA{R: 0.5, G: 0.5, B: 0.5, A: 1.0}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FromGGColor(c)
	}
}

func BenchmarkToGGColorPremultiplied(b *testing.B) {
	c := widget.Color{R: 0.5, G: 0.5, B: 0.5, A: 0.5}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ToGGColorPremultiplied(c)
	}
}

func BenchmarkLerpColor(b *testing.B) {
	a := gg.RGBA{R: 0, G: 0, B: 0, A: 1}
	c := gg.RGBA{R: 1, G: 1, B: 1, A: 1}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = LerpColor(a, c, 0.5)
	}
}
