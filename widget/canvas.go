package widget

import (
	"image"

	"github.com/gogpu/gg/scene"
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

// unknownStr is the string representation for unknown/unrecognized values.
const unknownStr = "Unknown"

// String returns a human-readable name for the text alignment.
func (a TextAlign) String() string {
	if int(a) < len(textAlignNames) {
		return textAlignNames[a]
	}
	return unknownStr
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

	// FillRectDirect fills a rectangle using CPU-only rendering, bypassing
	// the GPU shape accelerator. Used for dirty-region background clearing
	// where GPU acceleration is counterproductive (it blocks the compositor
	// non-MSAA blit-only fast path, causing 10x GPU overhead).
	FillRectDirect(r geometry.Rect, color Color)

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

	// StrokeArc draws a circular arc outline from startAngle with the given sweep.
	// Angles are in radians. startAngle=0 is 3 o'clock, positive is counterclockwise.
	// The arc is rendered as a single stroke operation using cubic Bézier curves.
	StrokeArc(center geometry.Point, radius float32, startAngle, sweepAngle float64,
		color Color, strokeWidth float32)

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
	// currently on the transform stack.
	TransformOffset() geometry.Point

	// ScreenOriginBase returns the screen-space base offset for this canvas.
	//
	// For the main window canvas this is (0,0). For SceneCanvas inside
	// RepaintBoundary recording, this is the boundary widget's screen position.
	// Without this, PushTransform(-bounds.Min) for local coords makes
	// TransformOffset() return local values, and StampScreenOrigin computes
	// wrong ScreenOrigin → dirty.Collector reports regions at (0,0).
	//
	// Flutter equivalent: PaintingContext.offset in RenderObject.paint().
	ScreenOriginBase() geometry.Point

	// ClipBounds returns the current clip rectangle.
	//
	// This is the intersection of all PushClip/PushClipRoundRect regions
	// currently on the clip stack. Used for viewport culling: containers
	// skip Draw on children whose bounds don't intersect the clip region,
	// preventing offscreen widgets from ticking animations.
	ClipBounds() geometry.Rect

	// ReplayScene renders a previously recorded scene.Scene display list
	// into this canvas. Used by RepaintBoundary to replay cached content
	// without re-executing the child widget's Draw method.
	//
	// For Canvas (gg.Context wrapper): renders the scene via GPU scene
	// renderer, which routes commands through gg.Context's GPU accelerator.
	// For SceneCanvas: merges the child scene into the parent scene via
	// Scene.Append (O(commands), zero re-encoding).
	//
	// If s is nil or empty, this is a no-op.
	ReplayScene(s *scene.Scene)
}

// DamageController can suppress damage tracking during rendering.
// Implemented by render.Canvas for retained-mode optimization.
// RepaintBoundary uses this to suppress damage during cached scene replay.
type DamageController interface {
	SetDamageTracking(enabled bool)
}

// BoundaryRecorder is implemented by canvases that record into a boundary's
// scene.Scene. When DrawChild encounters a child that IS a boundary, it skips
// drawing — the child boundary has its own PictureLayer in the compositor.
//
// Flutter equivalent: PaintingContext knows it's recording into a boundary's
// PictureRecorder. paintChild checks isRepaintBoundary → appendLayer instead
// of painting into the current recorder.
type BoundaryRecorder interface {
	IsBoundaryRecording() bool
}

// LineCap specifies how the endpoints of stroked arcs and lines are drawn.
type LineCap uint8

const (
	LineCapButt   LineCap = iota // flat end, stops exactly at endpoint
	LineCapRound                 // semicircle extending past endpoint
	LineCapSquare                // half-square extending past endpoint
)

// ArcStroker is an optional interface for canvases that support styled arc strokes.
// Use type assertion to check: if s, ok := canvas.(ArcStroker); ok { ... }
type ArcStroker interface {
	// StrokeArcStyled draws a circular arc with the specified line cap style.
	StrokeArcStyled(center geometry.Point, radius float32, startAngle, sweepAngle float64,
		color Color, strokeWidth float32, lineCap LineCap)
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

// DeviceScaler is an optional interface for canvases that support DPI-aware
// SVG icon rasterization (ADR-026). When a canvas implements DeviceScaler,
// the boundary recording infrastructure sets the display scale factor so
// that SVG icons are rasterized at physical pixel size (ceil(logical * scale))
// and drawn with an inverse-scale transform for crisp HiDPI rendering.
//
// Use type assertion to set the scale after creating a recording canvas:
//
//	if ds, ok := recorder.(widget.DeviceScaler); ok {
//	    ds.SetDeviceScale(ctx.Scale())
//	}
//
// This follows the Qt6/Chromium/IntelliJ pattern where icon assets are
// rendered at the device's native resolution rather than logical pixel size.
type DeviceScaler interface {
	SetDeviceScale(scale float32)
}

// TextMode controls text rendering strategy selection.
//
// Maps 1:1 to gg.TextMode. The default TextModeAuto selects the best strategy
// automatically based on GPU availability, transform, and font size.
type TextMode int

const (
	TextModeAuto      TextMode = iota // auto-select based on context
	TextModeMSDF                      // GPU MSDF atlas (games, animations)
	TextModeVector                    // vector outlines (quality-critical)
	TextModeBitmap                    // CPU bitmap (export, static)
	TextModeGlyphMask                 // GPU glyph mask (UI labels, <48px)
)

// textModeNames maps each TextMode to its human-readable name.
var textModeNames = [...]string{
	TextModeAuto:      "Auto",
	TextModeMSDF:      "MSDF",
	TextModeVector:    "Vector",
	TextModeBitmap:    "Bitmap",
	TextModeGlyphMask: "GlyphMask",
}

// String returns the text mode name.
func (m TextMode) String() string {
	if int(m) >= 0 && int(m) < len(textModeNames) {
		return textModeNames[m]
	}
	return unknownStr
}

// TextModeController is an optional interface for canvases that support
// explicit text rendering mode control.
//
// Use type assertion to check availability:
//
//	if tc, ok := canvas.(widget.TextModeController); ok {
//	    tc.SetTextMode(widget.TextModeMSDF)
//	    defer tc.SetTextMode(widget.TextModeAuto)
//	}
//
// This is primarily useful during zoom/scale operations where the default
// auto-selection may cause atlas pressure (issue #94).
//
// On SceneCanvas (RepaintBoundary recording), SetTextMode is a no-op because
// scene text uses TagText which handles mode selection at replay time.
type TextModeController interface {
	SetTextMode(mode TextMode)
	TextMode() TextMode
}

// TextStyle specifies font properties for styled text rendering.
//
// TextStyle is used with [StyledTextDrawer] to render text with custom fonts
// loaded through the plugin system or font registry. When FontFamily is empty,
// the default embedded Inter font is used.
//
// Example:
//
//	style := widget.TextStyle{
//	    FontFamily: "NotoSansCJK",
//	    FontSize:   16,
//	    Bold:       true,
//	    Color:      widget.ColorBlack,
//	    Align:      widget.TextAlignLeft,
//	}
type TextStyle struct {
	// FontFamily is the font family name (e.g., "Inter", "NotoSansCJK").
	// Empty string falls back to the default embedded font.
	FontFamily string

	// FontSize is the font size in logical pixels.
	FontSize float32

	// Bold selects bold weight (700). When false, regular weight (400) is used.
	Bold bool

	// Italic selects italic style. When false, normal style is used.
	Italic bool

	// Color is the text color.
	Color Color

	// Align controls horizontal text alignment within bounds.
	Align TextAlign
}

// StyledTextDrawer is an optional interface for canvases that support
// rendering text with custom fonts from the font registry.
//
// Use type assertion to check availability:
//
//	if sd, ok := canvas.(widget.StyledTextDrawer); ok {
//	    sd.DrawStyledText("Hello", bounds, widget.TextStyle{
//	        FontFamily: "NotoSansCJK",
//	        FontSize:   16,
//	        Color:      widget.ColorBlack,
//	    })
//	    width := sd.MeasureStyledText("Hello", widget.TextStyle{
//	        FontFamily: "NotoSansCJK",
//	        FontSize:   16,
//	    })
//	}
//
// Widgets that need custom font support should check for this interface
// and fall back to [Canvas.DrawText] when it is not available. This follows
// the same optional interface pattern as [ArcStroker], [SVGFiller], and
// [TextModeController].
type StyledTextDrawer interface {
	// DrawStyledText draws text within the given bounding rectangle using
	// the font specified in the TextStyle. The font is resolved from the
	// global font registry using CSS weight matching.
	DrawStyledText(text string, bounds geometry.Rect, style TextStyle)

	// MeasureStyledText returns the width in pixels of the given text
	// when rendered with the specified TextStyle.
	MeasureStyledText(text string, style TextStyle) float32
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
