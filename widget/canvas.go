package widget

import (
	"image"

	"github.com/gogpu/ui/geometry"
)

// TextAlign specifies horizontal text alignment within bounds.
type TextAlign uint8

const (
	// TextAlignLeft aligns text to the left edge (default).
	TextAlignLeft TextAlign = iota

	// TextAlignCenter centers text horizontally.
	TextAlignCenter

	// TextAlignRight aligns text to the right edge.
	TextAlignRight
)

// textAlignNames maps each TextAlign to its human-readable name.
var textAlignNames = [...]string{
	TextAlignLeft:   "Left",
	TextAlignCenter: "Center",
	TextAlignRight:  "Right",
}

// String returns a human-readable name for the text alignment.
func (a TextAlign) String() string {
	if int(a) < len(textAlignNames) {
		return textAlignNames[a]
	}
	return "Unknown"
}

// Float64 returns the alignment as a float64 value for rendering backends.
// Left=0, Center=0.5, Right=1.
func (a TextAlign) Float64() float64 {
	switch a {
	case TextAlignCenter:
		return 0.5
	case TextAlignRight:
		return 1.0
	default:
		return 0.0
	}
}

// Canvas provides drawing operations for widgets.
//
// Canvas is passed to widgets during the Draw phase. It provides methods
// for drawing shapes, text, and images. The full implementation will be
// in the render package; this is a placeholder interface.
//
// Coordinate System:
//
// Canvas uses a coordinate system where (0,0) is the top-left corner of
// the window, X increases to the right, and Y increases downward.
// All coordinates are in logical pixels, which may be scaled for HiDPI displays.
//
// Example:
//
//	func (w *MyWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
//	    // Fill background
//	    canvas.DrawRect(w.Bounds(), widget.ColorWhite)
//	    // Draw border
//	    canvas.StrokeRect(w.Bounds(), widget.ColorBlack, 1.0)
//	}
//
// Thread Safety:
//
// Canvas is NOT thread-safe. All drawing operations must occur on the
// main/UI thread during the Draw phase.
type Canvas interface {
	// Clear fills the entire canvas with the given color.
	Clear(color Color)

	// DrawRect fills a rectangle with the given color.
	DrawRect(r geometry.Rect, color Color)

	// StrokeRect draws the outline of a rectangle.
	//
	// The strokeWidth specifies the line thickness in logical pixels.
	StrokeRect(r geometry.Rect, color Color, strokeWidth float32)

	// DrawRoundRect fills a rounded rectangle with the given color.
	//
	// The radius specifies the corner radius.
	DrawRoundRect(r geometry.Rect, color Color, radius float32)

	// StrokeRoundRect draws the outline of a rounded rectangle.
	StrokeRoundRect(r geometry.Rect, color Color, radius float32, strokeWidth float32)

	// DrawCircle fills a circle with the given color.
	//
	// The center and radius specify the circle geometry.
	DrawCircle(center geometry.Point, radius float32, color Color)

	// StrokeCircle draws the outline of a circle.
	StrokeCircle(center geometry.Point, radius float32, color Color, strokeWidth float32)

	// DrawLine draws a line between two points.
	DrawLine(from, to geometry.Point, color Color, strokeWidth float32)

	// DrawText draws text within the given bounding rectangle.
	//
	// The fontSize is in logical pixels. The bold flag selects bold weight.
	// The align parameter controls horizontal alignment.
	DrawText(text string, bounds geometry.Rect, fontSize float32, color Color, bold bool, align TextAlign)

	// MeasureText returns the width in pixels of the given text string
	// when rendered at the specified font size and weight.
	// This is essential for accurate cursor positioning in text fields.
	MeasureText(text string, fontSize float32, bold bool) float32

	// DrawImage draws an image at the specified position.
	//
	// The image is drawn with its top-left corner at the given point.
	// The image is composited using source-over blending. This method
	// is used by RepaintBoundary to blit cached subtree renders.
	DrawImage(img image.Image, at geometry.Point)

	// PushClip pushes a clipping rectangle onto the clip stack.
	//
	// All subsequent drawing operations will be clipped to this rectangle
	// intersected with any parent clip rectangles.
	PushClip(r geometry.Rect)

	// PushClipRoundRect pushes a rounded rectangle clipping region.
	//
	// All subsequent drawing operations will be clipped to this rounded
	// rectangle. Uses GPU SDF-based clipping when available (gg.ClipRoundRect).
	PushClipRoundRect(r geometry.Rect, radius float32)

	// PopClip removes the most recently pushed clipping region.
	//
	// Must be called for each PushClip or PushClipRoundRect call.
	PopClip()

	// PushTransform pushes a translation transform onto the transform stack.
	//
	// All subsequent drawing operations will be offset by the given point.
	PushTransform(offset geometry.Point)

	// PopTransform removes the most recently pushed transform.
	//
	// Must be called for each PushTransform call.
	PopTransform()

	// TransformOffset returns the current cumulative transform offset.
	//
	// This is the total translation applied by all PushTransform calls
	// currently on the transform stack. It represents the mapping from
	// local widget coordinates to window (screen) coordinates.
	//
	// Used by [StampScreenOrigin] to compute a widget's screen-space
	// position during the Draw pass.
	TransformOffset() geometry.Point
}

// SVGFiller is an optional interface for canvases that support SVG path fill.
// Use type assertion to check: if f, ok := canvas.(SVGFiller); ok { ... }
type SVGFiller interface {
	// FillSVGPath fills an SVG path within the given bounds.
	FillSVGPath(svgData string, viewBox float32, bounds geometry.Rect, color Color)
}

// SVGRenderer is an optional interface for canvases that support full SVG rendering.
// Full SVG XML is rasterized to bitmap and drawn at the specified bounds.
type SVGRenderer interface {
	// RenderSVG renders full SVG XML within the given bounds with color override.
	RenderSVG(svgXML []byte, bounds geometry.Rect, color Color)
}

// Color represents an RGBA color with float32 components.
//
// Each component is in the range [0, 1], where 0 is minimum intensity
// and 1 is maximum intensity. For alpha, 0 is fully transparent and
// 1 is fully opaque.
//
// Color values use premultiplied alpha for efficient blending.
type Color struct {
	R, G, B, A float32
}

// RGBA creates a Color from red, green, blue, and alpha components.
//
// All components should be in the range [0, 1].
//
// Example:
//
//	red := widget.RGBA(1, 0, 0, 1)      // Solid red
//	semiRed := widget.RGBA(1, 0, 0, 0.5) // 50% transparent red
func RGBA(r, g, b, a float32) Color {
	return Color{R: r, G: g, B: b, A: a}
}

// RGB creates an opaque Color from red, green, and blue components.
//
// All components should be in the range [0, 1].
//
// Example:
//
//	white := widget.RGB(1, 1, 1)
//	black := widget.RGB(0, 0, 0)
func RGB(r, g, b float32) Color {
	return Color{R: r, G: g, B: b, A: 1}
}

// RGBA8 creates a Color from 8-bit RGBA values (0-255).
//
// Example:
//
//	red := widget.RGBA8(255, 0, 0, 255)
func RGBA8(r, g, b, a uint8) Color {
	return Color{
		R: float32(r) / colorMax8,
		G: float32(g) / colorMax8,
		B: float32(b) / colorMax8,
		A: float32(a) / colorMax8,
	}
}

// RGB8 creates an opaque Color from 8-bit RGB values (0-255).
//
// Example:
//
//	white := widget.RGB8(255, 255, 255)
func RGB8(r, g, b uint8) Color {
	return Color{
		R: float32(r) / colorMax8,
		G: float32(g) / colorMax8,
		B: float32(b) / colorMax8,
		A: 1,
	}
}

// Hex creates a Color from a 24-bit hex value (0xRRGGBB).
//
// Example:
//
//	skyBlue := widget.Hex(0x87CEEB)
//	coral := widget.Hex(0xFF7F50)
func Hex(hex uint32) Color {
	return Color{
		R: float32((hex>>16)&0xFF) / colorMax8,
		G: float32((hex>>8)&0xFF) / colorMax8,
		B: float32(hex&0xFF) / colorMax8,
		A: 1,
	}
}

// HexA creates a Color from a 32-bit hex value with alpha (0xRRGGBBAA).
//
// Example:
//
//	semiBlue := widget.HexA(0x0000FF80) // 50% transparent blue
func HexA(hex uint32) Color {
	return Color{
		R: float32((hex>>24)&0xFF) / colorMax8,
		G: float32((hex>>16)&0xFF) / colorMax8,
		B: float32((hex>>8)&0xFF) / colorMax8,
		A: float32(hex&0xFF) / colorMax8,
	}
}

// colorMax8 is the maximum value for 8-bit color components.
const colorMax8 = 255.0

// WithAlpha returns a copy of the color with the given alpha value.
//
// Example:
//
//	semiRed := widget.ColorRed.WithAlpha(0.5)
func (c Color) WithAlpha(a float32) Color {
	return Color{R: c.R, G: c.G, B: c.B, A: a}
}

// Lerp returns a color linearly interpolated between c and other.
//
// t=0 returns c, t=1 returns other.
//
// Example:
//
//	// Fade from red to blue
//	mid := widget.ColorRed.Lerp(widget.ColorBlue, 0.5)
func (c Color) Lerp(other Color, t float32) Color {
	return Color{
		R: c.R + (other.R-c.R)*t,
		G: c.G + (other.G-c.G)*t,
		B: c.B + (other.B-c.B)*t,
		A: c.A + (other.A-c.A)*t,
	}
}

// IsOpaque returns true if the color is fully opaque (alpha == 1).
func (c Color) IsOpaque() bool {
	return c.A >= 1.0
}

// IsTransparent returns true if the color is fully transparent (alpha == 0).
func (c Color) IsTransparent() bool {
	return c.A <= 0.0
}

// RGBA8 returns the color as 8-bit RGBA components (0-255).
func (c Color) RGBA8() (r, g, b, a uint8) {
	return uint8(clamp01(c.R) * colorMax8),
		uint8(clamp01(c.G) * colorMax8),
		uint8(clamp01(c.B) * colorMax8),
		uint8(clamp01(c.A) * colorMax8)
}

// clamp01 clamps a float32 value to the range [0, 1].
func clamp01(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// Common color constants.
var (
	// ColorTransparent is fully transparent (alpha = 0).
	ColorTransparent = Color{R: 0, G: 0, B: 0, A: 0}

	// ColorBlack is solid black.
	ColorBlack = Color{R: 0, G: 0, B: 0, A: 1}

	// ColorWhite is solid white.
	ColorWhite = Color{R: 1, G: 1, B: 1, A: 1}

	// ColorRed is solid red.
	ColorRed = Color{R: 1, G: 0, B: 0, A: 1}

	// ColorGreen is solid green.
	ColorGreen = Color{R: 0, G: 1, B: 0, A: 1}

	// ColorBlue is solid blue.
	ColorBlue = Color{R: 0, G: 0, B: 1, A: 1}

	// ColorYellow is solid yellow.
	ColorYellow = Color{R: 1, G: 1, B: 0, A: 1}

	// ColorCyan is solid cyan.
	ColorCyan = Color{R: 0, G: 1, B: 1, A: 1}

	// ColorMagenta is solid magenta.
	ColorMagenta = Color{R: 1, G: 0, B: 1, A: 1}

	// ColorGray is medium gray (50% brightness).
	ColorGray = Color{R: 0.5, G: 0.5, B: 0.5, A: 1}

	// ColorLightGray is light gray (75% brightness).
	ColorLightGray = Color{R: 0.75, G: 0.75, B: 0.75, A: 1}

	// ColorDarkGray is dark gray (25% brightness).
	ColorDarkGray = Color{R: 0.25, G: 0.25, B: 0.25, A: 1}
)
