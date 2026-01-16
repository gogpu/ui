package theme

import (
	"testing"

	"github.com/gogpu/ui/widget"
)

func TestShadow_WithAlpha(t *testing.T) {
	s := Shadow{
		OffsetX: 0,
		OffsetY: 2,
		Blur:    4,
		Spread:  0,
		Color:   widget.RGBA(0, 0, 0, 1),
	}

	result := s.WithAlpha(0.5)

	if result.Color.A != 0.5 {
		t.Errorf("Color.A = %v, want 0.5", result.Color.A)
	}
	// Other properties unchanged
	if result.Blur != 4 {
		t.Errorf("Blur changed: %v", result.Blur)
	}
}

func TestShadow_WithBlur(t *testing.T) {
	s := Shadow{Blur: 4}
	result := s.WithBlur(8)

	if result.Blur != 8 {
		t.Errorf("Blur = %v, want 8", result.Blur)
	}
	// Original unchanged
	if s.Blur != 4 {
		t.Errorf("Original Blur changed: %v", s.Blur)
	}
}

func TestShadow_WithOffset(t *testing.T) {
	s := Shadow{OffsetX: 0, OffsetY: 2}
	result := s.WithOffset(4, 8)

	if result.OffsetX != 4 {
		t.Errorf("OffsetX = %v, want 4", result.OffsetX)
	}
	if result.OffsetY != 8 {
		t.Errorf("OffsetY = %v, want 8", result.OffsetY)
	}
}

func TestShadow_Scale(t *testing.T) {
	s := Shadow{
		OffsetX: 2,
		OffsetY: 4,
		Blur:    8,
		Spread:  1,
		Color:   widget.RGBA(0, 0, 0, 0.5),
	}

	result := s.Scale(2)

	if result.OffsetX != 4 {
		t.Errorf("OffsetX = %v, want 4", result.OffsetX)
	}
	if result.OffsetY != 8 {
		t.Errorf("OffsetY = %v, want 8", result.OffsetY)
	}
	if result.Blur != 16 {
		t.Errorf("Blur = %v, want 16", result.Blur)
	}
	if result.Spread != 2 {
		t.Errorf("Spread = %v, want 2", result.Spread)
	}
	// Color unchanged
	if result.Color.A != 0.5 {
		t.Errorf("Color.A changed: %v", result.Color.A)
	}
}

func TestShadow_IsZero(t *testing.T) {
	tests := []struct {
		name   string
		shadow Shadow
		want   bool
	}{
		{
			"zero shadow is zero",
			Shadow{Blur: 0, Spread: 0, Color: widget.ColorTransparent},
			true,
		},
		{
			"shadow with blur is not zero",
			Shadow{Blur: 4, Spread: 0, Color: widget.ColorTransparent},
			false,
		},
		{
			"shadow with spread is not zero",
			Shadow{Blur: 0, Spread: 2, Color: widget.ColorTransparent},
			false,
		},
		{
			"shadow with color is not zero",
			Shadow{Blur: 0, Spread: 0, Color: widget.RGBA(0, 0, 0, 0.5)},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.shadow.IsZero(); got != tt.want {
				t.Errorf("IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestElevationShadow_WithAlpha(t *testing.T) {
	e := ElevationShadow{
		Key:     Shadow{Color: widget.RGBA(0, 0, 0, 1)},
		Ambient: Shadow{Color: widget.RGBA(0, 0, 0, 1)},
	}

	result := e.WithAlpha(0.5)

	if result.Key.Color.A != 0.5 {
		t.Errorf("Key.Color.A = %v, want 0.5", result.Key.Color.A)
	}
	if result.Ambient.Color.A != 0.5 {
		t.Errorf("Ambient.Color.A = %v, want 0.5", result.Ambient.Color.A)
	}
}

func TestElevationShadow_Scale(t *testing.T) {
	e := ElevationShadow{
		Key:     Shadow{Blur: 4},
		Ambient: Shadow{Blur: 2},
	}

	result := e.Scale(2)

	if result.Key.Blur != 8 {
		t.Errorf("Key.Blur = %v, want 8", result.Key.Blur)
	}
	if result.Ambient.Blur != 4 {
		t.Errorf("Ambient.Blur = %v, want 4", result.Ambient.Blur)
	}
}

func TestShadowStyles_ForElevation(t *testing.T) {
	s := DefaultShadowsLight()

	tests := []struct {
		name  string
		level int
	}{
		{"level 0", 0},
		{"level 1", 1},
		{"level 2", 2},
		{"level 3", 3},
		{"level 4", 4},
		{"level 5", 5},
		{"level negative clamps to 0", -1},
		{"level 6 clamps to 5", 6},
		{"level 100 clamps to 5", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.ForElevation(tt.level)
			// Just verify it doesn't panic and returns something
			_ = result.Key
			_ = result.Ambient
		})
	}
}

func TestDefaultShadowsLight(t *testing.T) {
	s := DefaultShadowsLight()

	// Level 0 should have no shadow
	if !s.Level0.Key.IsZero() {
		t.Error("Level0 Key should be zero")
	}
	if !s.Level0.Ambient.IsZero() {
		t.Error("Level0 Ambient should be zero")
	}

	// Higher levels should have larger blur
	if s.Level1.Key.Blur >= s.Level2.Key.Blur {
		t.Error("Level2 should have larger blur than Level1")
	}
	if s.Level4.Key.Blur >= s.Level5.Key.Blur {
		t.Error("Level5 should have larger blur than Level4")
	}
}

func TestDefaultShadowsDark(t *testing.T) {
	s := DefaultShadowsDark()

	// Level 0 should have no shadow
	if !s.Level0.Key.IsZero() {
		t.Error("Level0 Key should be zero")
	}

	// Dark shadows should be stronger (higher alpha) than light
	lightShadows := DefaultShadowsLight()
	if s.Level1.Key.Color.A <= lightShadows.Level1.Key.Color.A {
		t.Error("Dark shadow should have higher alpha than light shadow")
	}
}

func TestShadowStyles_ElevationProgression(t *testing.T) {
	s := DefaultShadowsLight()

	levels := []ElevationShadow{
		s.Level0, s.Level1, s.Level2, s.Level3, s.Level4, s.Level5,
	}

	// Each level should have equal or greater blur than the previous
	for i := 1; i < len(levels); i++ {
		if levels[i].Key.Blur < levels[i-1].Key.Blur {
			t.Errorf("Level%d blur (%v) should be >= Level%d blur (%v)",
				i, levels[i].Key.Blur, i-1, levels[i-1].Key.Blur)
		}
	}
}
