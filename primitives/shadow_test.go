package primitives

import "testing"

// TestShadowLayersLevelZero verifies that level 0 produces no shadow layers.
func TestShadowLayersLevelZero(t *testing.T) {
	layers := shadowLayers(0)
	if layers != nil {
		t.Errorf("level 0 should return nil, got %d layers", len(layers))
	}
}

// TestShadowLayersNegativeLevel verifies that negative levels produce no layers.
func TestShadowLayersNegativeLevel(t *testing.T) {
	layers := shadowLayers(-5)
	if layers != nil {
		t.Errorf("negative level should return nil, got %d layers", len(layers))
	}
}

// TestShadowLayersCountPerLevel verifies the expected number of layers at
// each elevation level.
func TestShadowLayersCountPerLevel(t *testing.T) {
	tests := []struct {
		name  string
		level int
		want  int
	}{
		{"level 1 has 2 layers", 1, 2},
		{"level 2 has 3 layers", 2, 3},
		{"level 3 has 3 layers", 3, 3},
		{"level 4 has 4 layers", 4, 4},
		{"level 5 has 4 layers", 5, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layers := shadowLayers(tt.level)
			if len(layers) != tt.want {
				t.Errorf("shadowLayers(%d) returned %d layers, want %d", tt.level, len(layers), tt.want)
			}
		})
	}
}

// TestShadowLayersOverflowClamped verifies that levels beyond maxShadowLevel
// are clamped to level 5.
func TestShadowLayersOverflowClamped(t *testing.T) {
	layers := shadowLayers(99)
	expected := shadowLayers(maxShadowLevel)
	if len(layers) != len(expected) {
		t.Errorf("overflow level should clamp to %d, got %d layers (want %d)",
			maxShadowLevel, len(layers), len(expected))
	}
}

// TestShadowLayersProgressiveElevation verifies that higher levels have
// larger total alpha (more visible shadow) and larger total spread.
func TestShadowLayersProgressiveElevation(t *testing.T) {
	var prevAlpha, prevSpread float32
	for level := 1; level <= maxShadowLevel; level++ {
		layers := shadowLayers(level)

		var totalAlpha, totalSpread float32
		for _, l := range layers {
			totalAlpha += l.alpha
			totalSpread += l.spread
		}

		if level > 1 {
			if totalAlpha < prevAlpha {
				t.Errorf("level %d total alpha (%.3f) < level %d (%.3f): shadow should grow",
					level, totalAlpha, level-1, prevAlpha)
			}
			if totalSpread < prevSpread {
				t.Errorf("level %d total spread (%.1f) < level %d (%.1f): shadow should grow",
					level, totalSpread, level-1, prevSpread)
			}
		}
		prevAlpha = totalAlpha
		prevSpread = totalSpread
	}
}

// TestShadowLayersAlphaRange verifies that all alpha values are within
// the valid range (0, 1].
func TestShadowLayersAlphaRange(t *testing.T) {
	for level := 1; level <= maxShadowLevel; level++ {
		for i, l := range shadowLayers(level) {
			if l.alpha <= 0 || l.alpha > 1 {
				t.Errorf("level %d layer %d: alpha %.3f out of range (0, 1]", level, i, l.alpha)
			}
		}
	}
}

// TestShadowLayersOutermostFirst verifies that layers are ordered with the
// outermost (largest spread) first and innermost (smallest spread) last.
func TestShadowLayersOutermostFirst(t *testing.T) {
	for level := 1; level <= maxShadowLevel; level++ {
		layers := shadowLayers(level)
		for i := 1; i < len(layers); i++ {
			if layers[i].spread > layers[i-1].spread {
				t.Errorf("level %d: layer %d spread (%.1f) > layer %d spread (%.1f): should decrease",
					level, i, layers[i].spread, i-1, layers[i-1].spread)
			}
		}
	}
}
