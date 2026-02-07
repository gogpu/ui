package primitives

import "github.com/gogpu/ui/widget"

// unknownStr is the string representation for unknown/unrecognized values.
const unknownStr = "Unknown"

// TextAlign specifies horizontal text alignment within its bounding box.
type TextAlign uint8

// TextAlign constants.
const (
	// TextAlignStart aligns text to the start edge (left in LTR, right in RTL).
	TextAlignStart TextAlign = iota

	// TextAlignCenter centers text horizontally.
	TextAlignCenter

	// TextAlignEnd aligns text to the end edge (right in LTR, left in RTL).
	TextAlignEnd
)

// textAlignNames maps each TextAlign to its human-readable name.
var textAlignNames = [...]string{
	TextAlignStart:  "Start",
	TextAlignCenter: "Center",
	TextAlignEnd:    "End",
}

// String returns a human-readable name for the text alignment.
func (a TextAlign) String() string {
	if int(a) < len(textAlignNames) {
		return textAlignNames[a]
	}
	return unknownStr
}

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

// shadowAlpha returns the alpha value for a given shadow level.
// Each level increases opacity slightly.
func shadowAlpha(level int) float32 {
	if level <= 0 {
		return 0
	}
	if level > maxShadowLevel {
		level = maxShadowLevel
	}
	// Linear ramp: level 1 = 0.08, level 5 = 0.40
	return float32(level) * 0.08
}

// shadowOffset returns the vertical offset for a given shadow level.
func shadowOffset(level int) float32 {
	if level <= 0 {
		return 0
	}
	if level > maxShadowLevel {
		level = maxShadowLevel
	}
	return float32(level) * 2
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
