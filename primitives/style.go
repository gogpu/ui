package primitives

import "github.com/gogpu/ui/widget"

// unknownStr is the string representation for unknown/unrecognized values.
const unknownStr = "Unknown"

// TextAlign is an alias for widget.TextAlign for backward compatibility.
type TextAlign = widget.TextAlign

// TextAlign constants mapped to widget.TextAlign values.
const (
	// TextAlignStart aligns text to the start edge (left in LTR, right in RTL).
	TextAlignStart = widget.TextAlignLeft

	// TextAlignCenter centers text horizontally.
	TextAlignCenter = widget.TextAlignCenter

	// TextAlignEnd aligns text to the end edge (right in LTR, left in RTL).
	TextAlignEnd = widget.TextAlignRight
)

// TextOverflow specifies how text behaves when it exceeds its container.
type TextOverflow uint8

// TextOverflow constants.
const (
	// TextOverflowClip clips overflowing text at the container boundary.
	TextOverflowClip TextOverflow = iota

	// TextOverflowEllipsis truncates overflowing text and appends "...".
	TextOverflowEllipsis
)

// textOverflowNames maps each TextOverflow to its human-readable name.
var textOverflowNames = [...]string{
	TextOverflowClip:     "Clip",
	TextOverflowEllipsis: "Ellipsis",
}

// String returns a human-readable name for the text overflow mode.
func (o TextOverflow) String() string {
	if int(o) < len(textOverflowNames) {
		return textOverflowNames[o]
	}
	return unknownStr
}

// ImageFit specifies how an image is sized relative to its container.
type ImageFit uint8

// ImageFit constants.
const (
	// ImageFitContain scales the image to fit entirely within the container
	// while preserving aspect ratio. The image may not fill the entire
	// container.
	ImageFitContain ImageFit = iota

	// ImageFitCover scales the image to completely cover the container while
	// preserving aspect ratio. Parts of the image may be clipped.
	ImageFitCover

	// ImageFitFill stretches the image to exactly fill the container,
	// ignoring aspect ratio.
	ImageFitFill

	// ImageFitNone displays the image at its natural size, centered within
	// the container. Overflow is clipped.
	ImageFitNone
)

// imageFitNames maps each ImageFit to its human-readable name.
var imageFitNames = [...]string{
	ImageFitContain: "Contain",
	ImageFitCover:   "Cover",
	ImageFitFill:    "Fill",
	ImageFitNone:    "None",
}

// String returns a human-readable name for the image fit mode.
func (f ImageFit) String() string {
	if int(f) < len(imageFitNames) {
		return imageFitNames[f]
	}
	return unknownStr
}

// Border describes the border of a widget.
type Border struct {
	// Width is the border thickness in logical pixels.
	Width float32
	// Color is the border color.
	Color widget.Color
}

// IsZero returns true if the border has no visible thickness.
func (b Border) IsZero() bool {
	return b.Width <= 0 || b.Color.IsTransparent()
}

// Shadow describes a box shadow with an integer elevation level (0-5).
//
// Level 0 means no shadow. Higher levels produce progressively larger and
// more diffuse shadows. The rendering of each level is determined by the
// current theme; the primitives package uses a simple constant lookup.
type Shadow struct {
	// Level is the elevation level from 0 (none) to maxShadowLevel.
	Level int
}

// maxShadowLevel is the maximum supported shadow elevation.
const maxShadowLevel = 5

// IsZero returns true if the shadow has no elevation.
func (s Shadow) IsZero() bool {
	return s.Level <= 0
}

// shadowLayer describes a single concentric layer in a multi-layer shadow.
//
// Multiple layers with decreasing alpha and increasing spread approximate
// a Gaussian blur without requiring a GPU compute pass.
type shadowLayer struct {
	// offsetX is the horizontal offset in logical pixels.
	offsetX float32
	// offsetY is the vertical offset in logical pixels.
	offsetY float32
	// spread expands the rect by this amount on all sides.
	spread float32
	// alpha is the opacity of this layer (0..1).
	alpha float32
	// radiusExtra is added to the box corner radius for this layer.
	radiusExtra float32
}

// shadowLayers returns the multi-layer shadow definition for a given
// elevation level. Each level uses progressively more layers with larger
// offsets and spreads to approximate Gaussian blur.
//
// Level 0 returns nil (no shadow). Levels 1-5 return 2-4 layers ordered
// from outermost (softest, most spread) to innermost (sharpest, least spread).
func shadowLayers(level int) []shadowLayer {
	if level <= 0 {
		return nil
	}
	if level > maxShadowLevel {
		level = maxShadowLevel
	}
	return shadowPresets[level-1]
}

// shadowPresets contains pre-computed shadow layers for levels 1-5.
// Each preset is ordered outermost-first so that inner (darker) layers
// paint on top of outer (lighter) ones.
//
// The values approximate Material Design elevation shadows using
// concentric rounded rectangles with varying alpha and spread.
var shadowPresets = [maxShadowLevel][]shadowLayer{
	// Level 1: subtle elevation (cards at rest). 2 layers.
	{
		{offsetX: 0, offsetY: 0.5, spread: 3, alpha: 0.04, radiusExtra: 3},
		{offsetX: 0, offsetY: 1, spread: 1, alpha: 0.08, radiusExtra: 1},
	},
	// Level 2: medium elevation (hovering buttons, raised cards). 3 layers.
	{
		{offsetX: 0, offsetY: 1, spread: 6, alpha: 0.04, radiusExtra: 4},
		{offsetX: 0, offsetY: 2, spread: 3, alpha: 0.06, radiusExtra: 2},
		{offsetX: 0, offsetY: 2, spread: 1, alpha: 0.10, radiusExtra: 0},
	},
	// Level 3: pronounced elevation (navigation drawers, menus). 3 layers.
	{
		{offsetX: 0, offsetY: 2, spread: 10, alpha: 0.04, radiusExtra: 5},
		{offsetX: 0, offsetY: 4, spread: 5, alpha: 0.06, radiusExtra: 3},
		{offsetX: 0, offsetY: 4, spread: 2, alpha: 0.12, radiusExtra: 1},
	},
	// Level 4: strong elevation (dialogs). 4 layers.
	{
		{offsetX: 0, offsetY: 3, spread: 14, alpha: 0.03, radiusExtra: 6},
		{offsetX: 0, offsetY: 6, spread: 8, alpha: 0.05, radiusExtra: 4},
		{offsetX: 0, offsetY: 6, spread: 4, alpha: 0.08, radiusExtra: 2},
		{offsetX: 0, offsetY: 8, spread: 2, alpha: 0.12, radiusExtra: 0},
	},
	// Level 5: highest elevation (tooltips, overlays). 4 layers.
	{
		{offsetX: 0, offsetY: 4, spread: 18, alpha: 0.03, radiusExtra: 8},
		{offsetX: 0, offsetY: 8, spread: 10, alpha: 0.05, radiusExtra: 5},
		{offsetX: 0, offsetY: 10, spread: 5, alpha: 0.08, radiusExtra: 2},
		{offsetX: 0, offsetY: 12, spread: 2, alpha: 0.14, radiusExtra: 0},
	},
}

// Direction specifies the layout direction for child widgets within a container.
type Direction uint8

// Direction constants.
const (
	// DirectionVertical lays out children from top to bottom (default).
	DirectionVertical Direction = iota

	// DirectionHorizontal lays out children from left to right.
	DirectionHorizontal
)

// directionNames maps each Direction to its human-readable name.
var directionNames = [...]string{
	DirectionVertical:   "Vertical",
	DirectionHorizontal: "Horizontal",
}

// String returns a human-readable name for the layout direction.
func (d Direction) String() string {
	if int(d) < len(directionNames) {
		return directionNames[d]
	}
	return unknownStr
}

// ImageSource provides the natural dimensions of an image.
//
// Concrete implementations are provided by the rendering backend. The
// primitives package only queries the image bounds for layout; actual
// pixel data is drawn by the backend-specific Canvas implementation.
type ImageSource interface {
	// Bounds returns the natural pixel dimensions of the image.
	Bounds() [2]float32
}
