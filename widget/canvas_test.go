package widget

import (
	"testing"
)

func TestRGBA(t *testing.T) {
	c := RGBA(0.5, 0.25, 0.75, 1.0)
	if c.R != 0.5 {
		t.Errorf("R = %v, want 0.5", c.R)
	}
	if c.G != 0.25 {
		t.Errorf("G = %v, want 0.25", c.G)
	}
	if c.B != 0.75 {
		t.Errorf("B = %v, want 0.75", c.B)
	}
	if c.A != 1.0 {
		t.Errorf("A = %v, want 1.0", c.A)
	}
}

func TestRGB(t *testing.T) {
	c := RGB(0.5, 0.25, 0.75)
	if c.R != 0.5 {
		t.Errorf("R = %v, want 0.5", c.R)
	}
	if c.G != 0.25 {
		t.Errorf("G = %v, want 0.25", c.G)
	}
	if c.B != 0.75 {
		t.Errorf("B = %v, want 0.75", c.B)
	}
	if c.A != 1.0 {
		t.Errorf("A = %v, want 1.0 (opaque)", c.A)
	}
}

func TestRGBA8(t *testing.T) {
	tests := []struct {
		name    string
		r, g, b uint8
		a       uint8
		wantR   float32
		wantG   float32
		wantB   float32
		wantA   float32
	}{
		{"black", 0, 0, 0, 255, 0, 0, 0, 1},
		{"white", 255, 255, 255, 255, 1, 1, 1, 1},
		{"red", 255, 0, 0, 255, 1, 0, 0, 1},
		{"green", 0, 255, 0, 255, 0, 1, 0, 1},
		{"blue", 0, 0, 255, 255, 0, 0, 1, 1},
		{"semi-transparent", 128, 128, 128, 128, 128.0 / 255, 128.0 / 255, 128.0 / 255, 128.0 / 255},
		{"transparent", 255, 255, 255, 0, 1, 1, 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := RGBA8(tt.r, tt.g, tt.b, tt.a)
			if !floatEqual(c.R, tt.wantR) {
				t.Errorf("R = %v, want %v", c.R, tt.wantR)
			}
			if !floatEqual(c.G, tt.wantG) {
				t.Errorf("G = %v, want %v", c.G, tt.wantG)
			}
			if !floatEqual(c.B, tt.wantB) {
				t.Errorf("B = %v, want %v", c.B, tt.wantB)
			}
			if !floatEqual(c.A, tt.wantA) {
				t.Errorf("A = %v, want %v", c.A, tt.wantA)
			}
		})
	}
}

func TestRGB8(t *testing.T) {
	c := RGB8(128, 64, 192)
	if !floatEqual(c.R, 128.0/255) {
		t.Errorf("R = %v, want %v", c.R, 128.0/255)
	}
	if !floatEqual(c.G, 64.0/255) {
		t.Errorf("G = %v, want %v", c.G, 64.0/255)
	}
	if !floatEqual(c.B, 192.0/255) {
		t.Errorf("B = %v, want %v", c.B, 192.0/255)
	}
	if c.A != 1.0 {
		t.Errorf("A = %v, want 1.0 (opaque)", c.A)
	}
}

func TestHex(t *testing.T) {
	tests := []struct {
		name  string
		hex   uint32
		wantR float32
		wantG float32
		wantB float32
	}{
		{"black", 0x000000, 0, 0, 0},
		{"white", 0xFFFFFF, 1, 1, 1},
		{"red", 0xFF0000, 1, 0, 0},
		{"green", 0x00FF00, 0, 1, 0},
		{"blue", 0x0000FF, 0, 0, 1},
		{"sky blue", 0x87CEEB, 0x87 / 255.0, 0xCE / 255.0, 0xEB / 255.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Hex(tt.hex)
			if !floatEqual(c.R, tt.wantR) {
				t.Errorf("R = %v, want %v", c.R, tt.wantR)
			}
			if !floatEqual(c.G, tt.wantG) {
				t.Errorf("G = %v, want %v", c.G, tt.wantG)
			}
			if !floatEqual(c.B, tt.wantB) {
				t.Errorf("B = %v, want %v", c.B, tt.wantB)
			}
			if c.A != 1.0 {
				t.Errorf("A = %v, want 1.0 (opaque)", c.A)
			}
		})
	}
}

func TestHexA(t *testing.T) {
	tests := []struct {
		name  string
		hex   uint32
		wantR float32
		wantG float32
		wantB float32
		wantA float32
	}{
		{"transparent black", 0x00000000, 0, 0, 0, 0},
		{"opaque white", 0xFFFFFFFF, 1, 1, 1, 1},
		{"semi-red", 0xFF000080, 1, 0, 0, 128.0 / 255},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := HexA(tt.hex)
			if !floatEqual(c.R, tt.wantR) {
				t.Errorf("R = %v, want %v", c.R, tt.wantR)
			}
			if !floatEqual(c.G, tt.wantG) {
				t.Errorf("G = %v, want %v", c.G, tt.wantG)
			}
			if !floatEqual(c.B, tt.wantB) {
				t.Errorf("B = %v, want %v", c.B, tt.wantB)
			}
			if !floatEqual(c.A, tt.wantA) {
				t.Errorf("A = %v, want %v", c.A, tt.wantA)
			}
		})
	}
}

func TestColor_WithAlpha(t *testing.T) {
	red := ColorRed
	semiRed := red.WithAlpha(0.5)

	if semiRed.R != red.R {
		t.Error("WithAlpha should preserve R")
	}
	if semiRed.G != red.G {
		t.Error("WithAlpha should preserve G")
	}
	if semiRed.B != red.B {
		t.Error("WithAlpha should preserve B")
	}
	if semiRed.A != 0.5 {
		t.Errorf("A = %v, want 0.5", semiRed.A)
	}
}

func TestColor_Lerp(t *testing.T) {
	tests := []struct {
		name  string
		from  Color
		to    Color
		t     float32
		wantR float32
		wantG float32
		wantB float32
		wantA float32
	}{
		{"t=0 (start)", ColorRed, ColorBlue, 0, 1, 0, 0, 1},
		{"t=1 (end)", ColorRed, ColorBlue, 1, 0, 0, 1, 1},
		{"t=0.5 (middle)", ColorBlack, ColorWhite, 0.5, 0.5, 0.5, 0.5, 1},
		{"alpha interpolation", RGBA(1, 0, 0, 0), RGBA(1, 0, 0, 1), 0.5, 1, 0, 0, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.from.Lerp(tt.to, tt.t)
			if !floatEqual(c.R, tt.wantR) {
				t.Errorf("R = %v, want %v", c.R, tt.wantR)
			}
			if !floatEqual(c.G, tt.wantG) {
				t.Errorf("G = %v, want %v", c.G, tt.wantG)
			}
			if !floatEqual(c.B, tt.wantB) {
				t.Errorf("B = %v, want %v", c.B, tt.wantB)
			}
			if !floatEqual(c.A, tt.wantA) {
				t.Errorf("A = %v, want %v", c.A, tt.wantA)
			}
		})
	}
}

func TestColor_IsOpaque(t *testing.T) {
	tests := []struct {
		name   string
		color  Color
		opaque bool
	}{
		{"fully opaque", ColorRed, true},
		{"fully transparent", ColorTransparent, false},
		{"semi-transparent", RGBA(1, 0, 0, 0.5), false},
		{"exactly 1.0", RGBA(1, 1, 1, 1.0), true},
		{"slightly over 1.0", RGBA(1, 1, 1, 1.1), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.color.IsOpaque()
			if got != tt.opaque {
				t.Errorf("IsOpaque() = %v, want %v", got, tt.opaque)
			}
		})
	}
}

func TestColor_IsTransparent(t *testing.T) {
	tests := []struct {
		name        string
		color       Color
		transparent bool
	}{
		{"fully opaque", ColorRed, false},
		{"fully transparent", ColorTransparent, true},
		{"semi-transparent", RGBA(1, 0, 0, 0.5), false},
		{"exactly 0.0", RGBA(1, 1, 1, 0.0), true},
		{"slightly negative", RGBA(1, 1, 1, -0.1), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.color.IsTransparent()
			if got != tt.transparent {
				t.Errorf("IsTransparent() = %v, want %v", got, tt.transparent)
			}
		})
	}
}

func TestColor_RGBA8(t *testing.T) {
	tests := []struct {
		name                string
		color               Color
		wantR, wantG, wantB uint8
		wantA               uint8
	}{
		{"black", ColorBlack, 0, 0, 0, 255},
		{"white", ColorWhite, 255, 255, 255, 255},
		{"red", ColorRed, 255, 0, 0, 255},
		{"transparent", ColorTransparent, 0, 0, 0, 0},
		{"gray", ColorGray, 127, 127, 127, 255},
		// Clamping tests
		{"over 1.0", RGBA(1.5, 0, 0, 1), 255, 0, 0, 255},
		{"under 0.0", RGBA(-0.5, 0, 0, 1), 0, 0, 0, 255},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, g, b, a := tt.color.RGBA8()
			if r != tt.wantR {
				t.Errorf("R = %v, want %v", r, tt.wantR)
			}
			if g != tt.wantG {
				t.Errorf("G = %v, want %v", g, tt.wantG)
			}
			if b != tt.wantB {
				t.Errorf("B = %v, want %v", b, tt.wantB)
			}
			if a != tt.wantA {
				t.Errorf("A = %v, want %v", a, tt.wantA)
			}
		})
	}
}

func TestColorConstants(t *testing.T) {
	// Verify color constant values
	tests := []struct {
		name  string
		color Color
		wantR float32
		wantG float32
		wantB float32
		wantA float32
	}{
		{"transparent", ColorTransparent, 0, 0, 0, 0},
		{"black", ColorBlack, 0, 0, 0, 1},
		{"white", ColorWhite, 1, 1, 1, 1},
		{"red", ColorRed, 1, 0, 0, 1},
		{"green", ColorGreen, 0, 1, 0, 1},
		{"blue", ColorBlue, 0, 0, 1, 1},
		{"yellow", ColorYellow, 1, 1, 0, 1},
		{"cyan", ColorCyan, 0, 1, 1, 1},
		{"magenta", ColorMagenta, 1, 0, 1, 1},
		{"gray", ColorGray, 0.5, 0.5, 0.5, 1},
		{"light gray", ColorLightGray, 0.75, 0.75, 0.75, 1},
		{"dark gray", ColorDarkGray, 0.25, 0.25, 0.25, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.color.R != tt.wantR {
				t.Errorf("R = %v, want %v", tt.color.R, tt.wantR)
			}
			if tt.color.G != tt.wantG {
				t.Errorf("G = %v, want %v", tt.color.G, tt.wantG)
			}
			if tt.color.B != tt.wantB {
				t.Errorf("B = %v, want %v", tt.color.B, tt.wantB)
			}
			if tt.color.A != tt.wantA {
				t.Errorf("A = %v, want %v", tt.color.A, tt.wantA)
			}
		})
	}
}

func TestClamp01(t *testing.T) {
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
			got := clamp01(tt.value)
			if got != tt.want {
				t.Errorf("clamp01(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

// floatEqual compares two float32 values with tolerance for floating point errors.
func floatEqual(a, b float32) bool {
	const epsilon = 0.001
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < epsilon
}

func BenchmarkColor_Lerp(b *testing.B) {
	c1 := ColorRed
	c2 := ColorBlue

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c1.Lerp(c2, 0.5)
	}
}

func BenchmarkColor_RGBA8(b *testing.B) {
	c := ColorGray

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, _ = c.RGBA8()
	}
}

func BenchmarkHex(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Hex(0x87CEEB)
	}
}

func BenchmarkRGBA8_Constructor(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RGBA8(128, 64, 192, 255)
	}
}
