package stripe

import (
	"math"
	"testing"

	"github.com/gogpu/ui/geometry"
)

// isInteger reports whether f is an integer value (no fractional part).
func isInteger(f float32) bool {
	return f == float32(math.Round(float64(f)))
}

func TestButtonIconBounds_PixelSnap(t *testing.T) {
	tests := []struct {
		name      string
		btnBounds geometry.Rect
		showLabel bool
	}{
		{
			name:      "even width no label",
			btnBounds: geometry.NewRect(0, 0, 40, 40),
			showLabel: false,
		},
		{
			name:      "odd width no label",
			btnBounds: geometry.NewRect(0, 0, 41, 41),
			showLabel: false,
		},
		{
			name:      "even width with label",
			btnBounds: geometry.NewRect(0, 0, 64, 56),
			showLabel: true,
		},
		{
			name:      "odd width with label",
			btnBounds: geometry.NewRect(0, 0, 63, 55),
			showLabel: true,
		},
		{
			name:      "fractional offset no label",
			btnBounds: geometry.NewRect(0.5, 0.5, 40, 40),
			showLabel: false,
		},
		{
			name:      "fractional offset with label",
			btnBounds: geometry.NewRect(0.5, 0.5, 64, 56),
			showLabel: true,
		},
		{
			name:      "large odd width no label",
			btnBounds: geometry.NewRect(10, 20, 77, 77),
			showLabel: false,
		},
		{
			name:      "large odd width with label",
			btnBounds: geometry.NewRect(10, 20, 77, 77),
			showLabel: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := buttonIconBounds(tt.btnBounds, tt.showLabel)

			if !isInteger(r.Min.X) {
				t.Errorf("Min.X = %v, want integer (pixel-snapped)", r.Min.X)
			}
			if !isInteger(r.Min.Y) {
				t.Errorf("Min.Y = %v, want integer (pixel-snapped)", r.Min.Y)
			}
			if r.Width() != defaultIconSize {
				t.Errorf("Width() = %v, want %v", r.Width(), defaultIconSize)
			}
			if r.Height() != defaultIconSize {
				t.Errorf("Height() = %v, want %v", r.Height(), defaultIconSize)
			}
		})
	}
}

func TestDefaultIconSize_MatchesJetBrainsReference(t *testing.T) {
	// JetBrains expui icons have two variants: 16x16 (standard) and
	// @20x20 (stripe/toolbar). Stripe icon size should match the 20x20
	// reference. Verified: intellij-community/platform/icons/src/expui/
	// toolwindows/ contains both project.svg (16x16) and project@20x20.svg.
	// With vector rendering (gg#464), non-integer viewBox scaling is handled
	// correctly via stroke hinting (gg#463).
	const referenceSize float32 = 20
	if defaultIconSize != referenceSize {
		t.Errorf("defaultIconSize = %v, want %v (JetBrains @20x20 stripe variant)",
			defaultIconSize, referenceSize)
	}
}

func TestButtonIconBounds_CenteredInButton(t *testing.T) {
	btn := geometry.NewRect(0, 0, 40, 40)
	r := buttonIconBounds(btn, false)

	// Icon should be approximately centered (within 1px due to rounding).
	iconCenterX := r.Min.X + r.Width()/2
	btnCenterX := btn.Min.X + btn.Width()/2
	if diff := math.Abs(float64(iconCenterX - btnCenterX)); diff > 1.0 {
		t.Errorf("icon center X=%v, button center X=%v, diff=%v > 1px",
			iconCenterX, btnCenterX, diff)
	}

	iconCenterY := r.Min.Y + r.Height()/2
	btnCenterY := btn.Min.Y + btn.Height()/2
	if diff := math.Abs(float64(iconCenterY - btnCenterY)); diff > 1.0 {
		t.Errorf("icon center Y=%v, button center Y=%v, diff=%v > 1px",
			iconCenterY, btnCenterY, diff)
	}
}

func TestButtonIconBounds_WithLabel_TopAligned(t *testing.T) {
	btn := geometry.NewRect(0, 0, 64, 56)
	r := buttonIconBounds(btn, true)

	// With labels, icon should be at the top with defaultIconPaddingLabel offset.
	expectedY := float32(math.Round(float64(btn.Min.Y + defaultIconPaddingLabel)))
	if r.Min.Y != expectedY {
		t.Errorf("Min.Y = %v, want %v (top-aligned with padding)", r.Min.Y, expectedY)
	}
}

func TestButtonTextBounds_BelowIcon(t *testing.T) {
	btn := geometry.NewRect(0, 0, 64, 56)
	iconRect := geometry.NewRect(24, 4, 16, 16)

	r := buttonTextBounds(btn, iconRect)

	expectedY := iconRect.Max.Y + defaultIconTextGap
	if r.Min.Y != expectedY {
		t.Errorf("text Min.Y = %v, want %v (below icon with gap)", r.Min.Y, expectedY)
	}

	expectedWidth := btn.Width() - defaultTextPaddingH*2
	if r.Width() != expectedWidth {
		t.Errorf("text Width = %v, want %v", r.Width(), expectedWidth)
	}
}
