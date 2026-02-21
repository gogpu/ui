package overlay_test

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/overlay"
)

func TestPositionBelow(t *testing.T) {
	anchor := geometry.NewRect(100, 100, 200, 40) // x=100, y=100, w=200, h=40
	overlaySize := geometry.Sz(200, 150)
	windowSize := geometry.Sz(800, 600)

	pos := overlay.Position(overlay.PlacementBelow, anchor, overlaySize, windowSize, 4)

	if pos.X != 100 {
		t.Errorf("X = %f, want 100", pos.X)
	}
	// anchor.Max.Y = 140, gap = 4 -> y = 144
	if pos.Y != 144 {
		t.Errorf("Y = %f, want 144", pos.Y)
	}
}

func TestPositionAbove(t *testing.T) {
	anchor := geometry.NewRect(100, 300, 200, 40)
	overlaySize := geometry.Sz(200, 100)
	windowSize := geometry.Sz(800, 600)

	pos := overlay.Position(overlay.PlacementAbove, anchor, overlaySize, windowSize, 4)

	if pos.X != 100 {
		t.Errorf("X = %f, want 100", pos.X)
	}
	// anchor.Min.Y = 300, overlaySize.Height = 100, gap = 4 -> y = 196
	if pos.Y != 196 {
		t.Errorf("Y = %f, want 196", pos.Y)
	}
}

func TestPositionRight(t *testing.T) {
	anchor := geometry.NewRect(100, 100, 50, 40)
	overlaySize := geometry.Sz(120, 80)
	windowSize := geometry.Sz(800, 600)

	pos := overlay.Position(overlay.PlacementRight, anchor, overlaySize, windowSize, 4)

	// anchor.Max.X = 150, gap = 4 -> x = 154
	if pos.X != 154 {
		t.Errorf("X = %f, want 154", pos.X)
	}
	if pos.Y != 100 {
		t.Errorf("Y = %f, want 100", pos.Y)
	}
}

func TestPositionLeft(t *testing.T) {
	anchor := geometry.NewRect(300, 100, 50, 40)
	overlaySize := geometry.Sz(120, 80)
	windowSize := geometry.Sz(800, 600)

	pos := overlay.Position(overlay.PlacementLeft, anchor, overlaySize, windowSize, 4)

	// anchor.Min.X = 300, overlaySize.Width = 120, gap = 4 -> x = 176
	if pos.X != 176 {
		t.Errorf("X = %f, want 176", pos.X)
	}
	if pos.Y != 100 {
		t.Errorf("Y = %f, want 100", pos.Y)
	}
}

func TestPositionFlipBelow(t *testing.T) {
	// Anchor near the bottom, not enough room below -> flip above
	anchor := geometry.NewRect(100, 500, 200, 40) // bottom at 540
	overlaySize := geometry.Sz(200, 100)
	windowSize := geometry.Sz(800, 600)

	pos := overlay.Position(overlay.PlacementBelow, anchor, overlaySize, windowSize, 4)

	// Should flip above: anchor.Min.Y = 500, overlaySize.Height = 100, gap = 4 -> y = 396
	if pos.Y != 396 {
		t.Errorf("Y = %f, want 396 (flipped above)", pos.Y)
	}
}

func TestPositionFlipAbove(t *testing.T) {
	// Anchor near the top, not enough room above -> flip below
	anchor := geometry.NewRect(100, 20, 200, 40) // top at 20
	overlaySize := geometry.Sz(200, 100)
	windowSize := geometry.Sz(800, 600)

	pos := overlay.Position(overlay.PlacementAbove, anchor, overlaySize, windowSize, 4)

	// Cannot fit above (20 - 100 - 4 = -84), should flip below.
	// anchor.Max.Y = 60, gap = 4 -> y = 64
	if pos.Y != 64 {
		t.Errorf("Y = %f, want 64 (flipped below)", pos.Y)
	}
}

func TestPositionClampRight(t *testing.T) {
	// Anchor near the right edge, overlay extends past window.
	anchor := geometry.NewRect(700, 100, 80, 40)
	overlaySize := geometry.Sz(200, 100)
	windowSize := geometry.Sz(800, 600)

	pos := overlay.Position(overlay.PlacementBelow, anchor, overlaySize, windowSize, 4)

	// x = 700, but overlay would extend to 900 (> 800)
	// Clamped: x = 800 - 200 = 600
	if pos.X != 600 {
		t.Errorf("X = %f, want 600 (clamped)", pos.X)
	}
}

func TestPositionClampLeft(t *testing.T) {
	// Anchor at left edge, placement Left would go negative.
	anchor := geometry.NewRect(10, 100, 50, 40)
	overlaySize := geometry.Sz(120, 80)
	windowSize := geometry.Sz(800, 600)

	pos := overlay.Position(overlay.PlacementLeft, anchor, overlaySize, windowSize, 4)

	// anchor.Min.X = 10, overlaySize.Width = 120, gap = 4 -> x = -114
	// Flip: anchor.Max.X = 60, gap = 4 -> x = 64
	// Clamped: max(0, 64) = 64
	if pos.X < 0 {
		t.Errorf("X = %f, should be >= 0", pos.X)
	}
}

func TestPositionClampBottom(t *testing.T) {
	// Both flip and clamp needed.
	anchor := geometry.NewRect(100, 550, 200, 40)
	overlaySize := geometry.Sz(200, 200)
	windowSize := geometry.Sz(800, 600)

	pos := overlay.Position(overlay.PlacementBelow, anchor, overlaySize, windowSize, 4)

	// Below: y = 594, extends to 794 (out of 600). Flip above: y = 550 - 200 - 4 = 346.
	// 346 >= 0, so flip works.
	// After clamp: y = max(0, min(400, 346)) = 346... but 346 + 200 = 546 <= 600, fine.
	if pos.Y+overlaySize.Height > windowSize.Height {
		t.Errorf("Y + height = %f, exceeds window height %f", pos.Y+overlaySize.Height, windowSize.Height)
	}
	if pos.Y < 0 {
		t.Errorf("Y = %f, should be >= 0", pos.Y)
	}
}
