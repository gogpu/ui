package overlay

import "github.com/gogpu/ui/geometry"

// Placement defines where an overlay appears relative to its anchor widget.
type Placement uint8

// Placement constants.
const (
	// PlacementBelow positions the overlay below the anchor (default for dropdowns).
	PlacementBelow Placement = iota

	// PlacementAbove positions the overlay above the anchor.
	PlacementAbove

	// PlacementLeft positions the overlay to the left of the anchor.
	PlacementLeft

	// PlacementRight positions the overlay to the right of the anchor.
	PlacementRight
)

// Position calculates the overlay position relative to an anchor rectangle.
//
// Parameters:
//   - placement: preferred placement direction
//   - anchorGlobal: the anchor widget's bounds in window coordinates
//   - overlaySize: the measured size of the overlay content
//   - windowSize: the total window size for viewport clamping
//   - gap: spacing between the anchor and overlay edges
//
// The function applies flip logic (tries the opposite side if the overlay
// would go out of bounds) and then clamps to the viewport.
func Position(
	placement Placement,
	anchorGlobal geometry.Rect,
	overlaySize geometry.Size,
	windowSize geometry.Size,
	gap float32,
) geometry.Point {
	var x, y float32

	switch placement {
	case PlacementBelow:
		x = anchorGlobal.Min.X
		y = anchorGlobal.Max.Y + gap
	case PlacementAbove:
		x = anchorGlobal.Min.X
		y = anchorGlobal.Min.Y - overlaySize.Height - gap
	case PlacementRight:
		x = anchorGlobal.Max.X + gap
		y = anchorGlobal.Min.Y
	case PlacementLeft:
		x = anchorGlobal.Min.X - overlaySize.Width - gap
		y = anchorGlobal.Min.Y
	}

	// Flip if the overlay goes out of viewport bounds.
	x, y = flip(placement, x, y, anchorGlobal, overlaySize, windowSize, gap)

	// Clamp to viewport.
	x, y = clampToViewport(x, y, overlaySize, windowSize)

	return geometry.Pt(x, y)
}

// flip tries the opposite placement direction if the overlay extends beyond
// the viewport boundary.
func flip(
	placement Placement,
	x, y float32,
	anchor geometry.Rect,
	overlaySize geometry.Size,
	windowSize geometry.Size,
	gap float32,
) (float32, float32) {
	switch placement {
	case PlacementBelow:
		if y+overlaySize.Height > windowSize.Height {
			flipped := anchor.Min.Y - overlaySize.Height - gap
			if flipped >= 0 {
				y = flipped
			}
		}
	case PlacementAbove:
		if y < 0 {
			flipped := anchor.Max.Y + gap
			if flipped+overlaySize.Height <= windowSize.Height {
				y = flipped
			}
		}
	case PlacementRight:
		if x+overlaySize.Width > windowSize.Width {
			flipped := anchor.Min.X - overlaySize.Width - gap
			if flipped >= 0 {
				x = flipped
			}
		}
	case PlacementLeft:
		if x < 0 {
			flipped := anchor.Max.X + gap
			if flipped+overlaySize.Width <= windowSize.Width {
				x = flipped
			}
		}
	}
	return x, y
}

// clampToViewport ensures the overlay stays within the window boundaries.
func clampToViewport(x, y float32, overlaySize geometry.Size, windowSize geometry.Size) (float32, float32) {
	if x+overlaySize.Width > windowSize.Width {
		x = windowSize.Width - overlaySize.Width
	}
	if x < 0 {
		x = 0
	}
	if y+overlaySize.Height > windowSize.Height {
		y = windowSize.Height - overlaySize.Height
	}
	if y < 0 {
		y = 0
	}
	return x, y
}
