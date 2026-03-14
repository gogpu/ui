package popover

import "github.com/gogpu/ui/geometry"

// Placement defines where the popover or tooltip appears relative to its
// trigger widget. There are 12 placements: 4 sides with 3 alignment
// variants each (start, center, end).
type Placement uint8

// Placement constants.
const (
	// Bottom places the overlay below the trigger, centered horizontally.
	Bottom Placement = iota

	// BottomStart places the overlay below, aligned to the trigger's start edge.
	BottomStart

	// BottomEnd places the overlay below, aligned to the trigger's end edge.
	BottomEnd

	// Top places the overlay above the trigger, centered horizontally.
	Top

	// TopStart places the overlay above, aligned to the trigger's start edge.
	TopStart

	// TopEnd places the overlay above, aligned to the trigger's end edge.
	TopEnd

	// Left places the overlay to the left of the trigger, centered vertically.
	Left

	// LeftStart places the overlay to the left, aligned to the trigger's top edge.
	LeftStart

	// LeftEnd places the overlay to the left, aligned to the trigger's bottom edge.
	LeftEnd

	// Right places the overlay to the right of the trigger, centered vertically.
	Right

	// RightStart places the overlay to the right, aligned to the trigger's top edge.
	RightStart

	// RightEnd places the overlay to the right, aligned to the trigger's bottom edge.
	RightEnd
)

// placementNames maps placements to their human-readable names.
var placementNames = [...]string{
	Bottom:      "Bottom",
	BottomStart: "BottomStart",
	BottomEnd:   "BottomEnd",
	Top:         "Top",
	TopStart:    "TopStart",
	TopEnd:      "TopEnd",
	Left:        "Left",
	LeftStart:   "LeftStart",
	LeftEnd:     "LeftEnd",
	Right:       "Right",
	RightStart:  "RightStart",
	RightEnd:    "RightEnd",
}

// String returns a human-readable name for the placement.
func (p Placement) String() string {
	if int(p) < len(placementNames) {
		return placementNames[p]
	}
	return "Unknown"
}

// opposite returns the flipped placement (top <-> bottom, left <-> right).
func (p Placement) opposite() Placement {
	switch p {
	case Bottom:
		return Top
	case BottomStart:
		return TopStart
	case BottomEnd:
		return TopEnd
	case Top:
		return Bottom
	case TopStart:
		return BottomStart
	case TopEnd:
		return BottomEnd
	case Left:
		return Right
	case LeftStart:
		return RightStart
	case LeftEnd:
		return RightEnd
	case Right:
		return Left
	case RightStart:
		return LeftStart
	case RightEnd:
		return LeftEnd
	default:
		return p
	}
}

// defaultGap is the spacing between the trigger and the overlay.
const defaultGap float32 = 4

// CalculatePosition computes the overlay position given the trigger bounds,
// overlay size, and window size. It applies auto-flip logic when the overlay
// would overflow the viewport, then clamps to the window bounds.
func CalculatePosition(
	placement Placement,
	triggerBounds geometry.Rect,
	overlaySize geometry.Size,
	windowSize geometry.Size,
	gap float32,
) geometry.Point {
	x, y := computePosition(placement, triggerBounds, overlaySize, gap)

	// Auto-flip if overflows viewport.
	if overflows(x, y, overlaySize, windowSize) {
		fx, fy := computePosition(placement.opposite(), triggerBounds, overlaySize, gap)
		if !overflows(fx, fy, overlaySize, windowSize) {
			x, y = fx, fy
		}
	}

	// Clamp to viewport bounds.
	x, y = clampToViewport(x, y, overlaySize, windowSize)

	return geometry.Pt(x, y)
}

// computePosition calculates the raw position without any clamping or flipping.
func computePosition(
	placement Placement,
	anchor geometry.Rect,
	overlaySize geometry.Size,
	gap float32,
) (float32, float32) {
	var x, y float32

	switch placement {
	case Bottom:
		x = anchor.Center().X - overlaySize.Width/2
		y = anchor.Max.Y + gap
	case BottomStart:
		x = anchor.Min.X
		y = anchor.Max.Y + gap
	case BottomEnd:
		x = anchor.Max.X - overlaySize.Width
		y = anchor.Max.Y + gap

	case Top:
		x = anchor.Center().X - overlaySize.Width/2
		y = anchor.Min.Y - overlaySize.Height - gap
	case TopStart:
		x = anchor.Min.X
		y = anchor.Min.Y - overlaySize.Height - gap
	case TopEnd:
		x = anchor.Max.X - overlaySize.Width
		y = anchor.Min.Y - overlaySize.Height - gap

	case Left:
		x = anchor.Min.X - overlaySize.Width - gap
		y = anchor.Center().Y - overlaySize.Height/2
	case LeftStart:
		x = anchor.Min.X - overlaySize.Width - gap
		y = anchor.Min.Y
	case LeftEnd:
		x = anchor.Min.X - overlaySize.Width - gap
		y = anchor.Max.Y - overlaySize.Height

	case Right:
		x = anchor.Max.X + gap
		y = anchor.Center().Y - overlaySize.Height/2
	case RightStart:
		x = anchor.Max.X + gap
		y = anchor.Min.Y
	case RightEnd:
		x = anchor.Max.X + gap
		y = anchor.Max.Y - overlaySize.Height
	}

	return x, y
}

// overflows returns true if the overlay at (x, y) would exceed the viewport.
func overflows(x, y float32, overlaySize geometry.Size, windowSize geometry.Size) bool {
	return x < 0 || y < 0 ||
		x+overlaySize.Width > windowSize.Width ||
		y+overlaySize.Height > windowSize.Height
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
