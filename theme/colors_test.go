package theme

import (
	"testing"

	"github.com/gogpu/ui/widget"
)

func TestColorPalette_WithAlpha(t *testing.T) {
	palette := &ColorPalette{
		Primary:   widget.RGB(1, 0, 0),
		Secondary: widget.RGB(0, 1, 0),
		Error:     widget.RGB(1, 0, 0),
	}

	result := palette.WithAlpha(0.5)

	if result.Primary.A != 0.5 {
		t.Errorf("Primary alpha = %v, want 0.5", result.Primary.A)
	}
	if result.Secondary.A != 0.5 {
		t.Errorf("Secondary alpha = %v, want 0.5", result.Secondary.A)
	}
	if result.Error.A != 0.5 {
		t.Errorf("Error alpha = %v, want 0.5", result.Error.A)
	}
}

func TestColorPalette_Lerp(t *testing.T) {
	p1 := &ColorPalette{
		Primary: widget.RGB(0, 0, 0),
	}
	p2 := &ColorPalette{
		Primary: widget.RGB(1, 1, 1),
	}

	tests := []struct {
		name string
		t    float32
		want float32
	}{
		{"t=0 returns first", 0, 0},
		{"t=1 returns second", 1, 1},
		{"t=0.5 returns middle", 0.5, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p1.Lerp(p2, tt.t)
			if result.Primary.R != tt.want {
				t.Errorf("Lerp Primary.R = %v, want %v", result.Primary.R, tt.want)
			}
		})
	}
}

func TestContrastColor(t *testing.T) {
	onLight := widget.RGB(0, 0, 0) // Black
	onDark := widget.RGB(1, 1, 1)  // White

	tests := []struct {
		name       string
		background widget.Color
		wantDark   bool
	}{
		{"white background gets dark text", widget.ColorWhite, true},
		{"black background gets light text", widget.ColorBlack, false},
		{"light gray gets dark text", widget.RGB(0.8, 0.8, 0.8), true},
		{"dark gray gets light text", widget.RGB(0.2, 0.2, 0.2), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContrastColor(tt.background, onLight, onDark)
			if tt.wantDark {
				if result != onLight {
					t.Errorf("expected dark text (onLight), got onDark")
				}
			} else {
				if result != onDark {
					t.Errorf("expected light text (onDark), got onLight")
				}
			}
		})
	}
}

func TestLighten(t *testing.T) {
	red := widget.RGB(1, 0, 0)

	tests := []struct {
		name   string
		amount float32
		wantR  float32
		wantG  float32
	}{
		{"no lightening", 0, 1, 0},
		{"full lightening is white", 1, 1, 1},
		{"half lightening", 0.5, 1, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Lighten(red, tt.amount)
			if result.R != tt.wantR {
				t.Errorf("R = %v, want %v", result.R, tt.wantR)
			}
			if result.G != tt.wantG {
				t.Errorf("G = %v, want %v", result.G, tt.wantG)
			}
		})
	}
}

func TestDarken(t *testing.T) {
	white := widget.RGB(1, 1, 1)

	tests := []struct {
		name   string
		amount float32
		wantR  float32
	}{
		{"no darkening", 0, 1},
		{"full darkening is black", 1, 0},
		{"half darkening", 0.5, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Darken(white, tt.amount)
			if result.R != tt.wantR {
				t.Errorf("R = %v, want %v", result.R, tt.wantR)
			}
		})
	}
}

func TestWithOpacity(t *testing.T) {
	red := widget.RGBA(1, 0, 0, 1)

	tests := []struct {
		name    string
		opacity float32
		wantA   float32
	}{
		{"full opacity", 1, 1},
		{"half opacity", 0.5, 0.5},
		{"no opacity", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WithOpacity(red, tt.opacity)
			if result.A != tt.wantA {
				t.Errorf("A = %v, want %v", result.A, tt.wantA)
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
		{"negative clamps to 0", -1, 0},
		{"zero stays zero", 0, 0},
		{"one stays one", 1, 1},
		{"greater than one clamps to 1", 2, 1},
		{"middle value unchanged", 0.5, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clamp01(tt.value); got != tt.want {
				t.Errorf("clamp01(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestLighten_ClampsAmount(t *testing.T) {
	red := widget.RGB(1, 0, 0)

	// Amount > 1 should be clamped to 1
	result := Lighten(red, 2.0)
	if result.G != 1 {
		t.Errorf("G = %v, want 1 (amount should be clamped)", result.G)
	}

	// Amount < 0 should be clamped to 0
	result = Lighten(red, -1.0)
	if result.G != 0 {
		t.Errorf("G = %v, want 0 (amount should be clamped)", result.G)
	}
}

func TestDarken_ClampsAmount(t *testing.T) {
	white := widget.RGB(1, 1, 1)

	// Amount > 1 should be clamped to 1
	result := Darken(white, 2.0)
	if result.R != 0 {
		t.Errorf("R = %v, want 0 (amount should be clamped)", result.R)
	}

	// Amount < 0 should be clamped to 0
	result = Darken(white, -1.0)
	if result.R != 1 {
		t.Errorf("R = %v, want 1 (amount should be clamped)", result.R)
	}
}

func TestWithOpacity_ClampsOpacity(t *testing.T) {
	red := widget.RGBA(1, 0, 0, 1)

	// Opacity > 1 should be clamped to 1
	result := WithOpacity(red, 2.0)
	if result.A != 1 {
		t.Errorf("A = %v, want 1 (opacity should be clamped)", result.A)
	}

	// Opacity < 0 should be clamped to 0
	result = WithOpacity(red, -1.0)
	if result.A != 0 {
		t.Errorf("A = %v, want 0 (opacity should be clamped)", result.A)
	}
}

func TestWithOpacity_MultipliesExistingAlpha(t *testing.T) {
	// Color with 0.5 alpha
	semiRed := widget.RGBA(1, 0, 0, 0.5)

	// Apply 0.5 opacity = 0.5 * 0.5 = 0.25
	result := WithOpacity(semiRed, 0.5)
	if result.A != 0.25 {
		t.Errorf("A = %v, want 0.25 (0.5 * 0.5)", result.A)
	}
}
