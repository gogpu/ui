package uitest

import (
	"image"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// DrawRectCall records a single DrawRect invocation.
type DrawRectCall struct {
	Bounds geometry.Rect
	Color  widget.Color
}

// StrokeRectCall records a single StrokeRect invocation.
type StrokeRectCall struct {
	Bounds      geometry.Rect
	Color       widget.Color
	StrokeWidth float32
}

// DrawRoundRectCall records a single DrawRoundRect invocation.
type DrawRoundRectCall struct {
	Bounds geometry.Rect
	Color  widget.Color
	Radius float32
}

// StrokeRoundRectCall records a single StrokeRoundRect invocation.
type StrokeRoundRectCall struct {
	Bounds      geometry.Rect
	Color       widget.Color
	Radius      float32
	StrokeWidth float32
}

// DrawCircleCall records a single DrawCircle invocation.
type DrawCircleCall struct {
	Center geometry.Point
	Radius float32
	Color  widget.Color
}

// StrokeCircleCall records a single StrokeCircle invocation.
type StrokeCircleCall struct {
	Center      geometry.Point
	Radius      float32
	Color       widget.Color
	StrokeWidth float32
}

// DrawLineCall records a single DrawLine invocation.
type DrawLineCall struct {
	From        geometry.Point
	To          geometry.Point
	Color       widget.Color
	StrokeWidth float32
}

// DrawTextCall records a single DrawText invocation.
type DrawTextCall struct {
	Text     string
	Bounds   geometry.Rect
	FontSize float32
	Color    widget.Color
	Bold     bool
	Align    widget.TextAlign
}

// DrawImageCall records a single DrawImage invocation.
type DrawImageCall struct {
	Image image.Image
	At    geometry.Point
}

// ClipRoundRectCall records a single PushClipRoundRect invocation.
type ClipRoundRectCall struct {
	Bounds geometry.Rect
	Radius float32
}

// MockCanvas records all draw calls for verification in tests.
//
// It implements [widget.Canvas] and stores each invocation in typed slices.
// Use [MockCanvas.Reset] to clear all recorded calls between test phases.
//
// Example:
//
//	canvas := &uitest.MockCanvas{}
//	myWidget.Draw(ctx, canvas)
//	if len(canvas.Rects) != 1 {
//	    t.Errorf("expected 1 rect, got %d", len(canvas.Rects))
//	}
type MockCanvas struct {
	// Clears records arguments passed to Clear.
	Clears []widget.Color

	// Rects records arguments passed to DrawRect.
	Rects []DrawRectCall

	// StrokeRects records arguments passed to StrokeRect.
	StrokeRects []StrokeRectCall

	// RoundRects records arguments passed to DrawRoundRect.
	RoundRects []DrawRoundRectCall

	// StrokeRoundRects records arguments passed to StrokeRoundRect.
	StrokeRoundRects []StrokeRoundRectCall

	// Circles records arguments passed to DrawCircle.
	Circles []DrawCircleCall

	// StrokeCircles records arguments passed to StrokeCircle.
	StrokeCircles []StrokeCircleCall

	// Lines records arguments passed to DrawLine.
	Lines []DrawLineCall

	// Texts records arguments passed to DrawText.
	Texts []DrawTextCall

	// Images records arguments passed to DrawImage.
	Images []DrawImageCall

	// Clips records arguments passed to PushClip.
	Clips []geometry.Rect

	// ClipRoundRects records arguments passed to PushClipRoundRect.
	ClipRoundRects []ClipRoundRectCall

	// Transforms records arguments passed to PushTransform.
	Transforms []geometry.Point

	// PopClipCount tracks how many times PopClip was called.
	PopClipCount int

	// PopTransformCount tracks how many times PopTransform was called.
	PopTransformCount int

	// transformStack tracks cumulative offsets for TransformOffset().
	transformStack []geometry.Point
	// currentOffset is the current cumulative transform offset.
	currentOffset geometry.Point
}

// Clear records the color argument.
func (c *MockCanvas) Clear(color widget.Color) {
	c.Clears = append(c.Clears, color)
}

// DrawRect records the rect and color arguments.
func (c *MockCanvas) DrawRect(r geometry.Rect, color widget.Color) {
	c.Rects = append(c.Rects, DrawRectCall{Bounds: r, Color: color})
}

// StrokeRect records the rect, color, and stroke width arguments.
func (c *MockCanvas) StrokeRect(r geometry.Rect, color widget.Color, strokeWidth float32) {
	c.StrokeRects = append(c.StrokeRects, StrokeRectCall{Bounds: r, Color: color, StrokeWidth: strokeWidth})
}

// DrawRoundRect records the rect, color, and radius arguments.
func (c *MockCanvas) DrawRoundRect(r geometry.Rect, color widget.Color, radius float32) {
	c.RoundRects = append(c.RoundRects, DrawRoundRectCall{Bounds: r, Color: color, Radius: radius})
}

// StrokeRoundRect records the rect, color, radius, and stroke width arguments.
func (c *MockCanvas) StrokeRoundRect(r geometry.Rect, color widget.Color, radius float32, strokeWidth float32) {
	c.StrokeRoundRects = append(c.StrokeRoundRects, StrokeRoundRectCall{
		Bounds: r, Color: color, Radius: radius, StrokeWidth: strokeWidth,
	})
}

// DrawCircle records the center, radius, and color arguments.
func (c *MockCanvas) DrawCircle(center geometry.Point, radius float32, color widget.Color) {
	c.Circles = append(c.Circles, DrawCircleCall{Center: center, Radius: radius, Color: color})
}

// StrokeCircle records the center, radius, color, and stroke width arguments.
func (c *MockCanvas) StrokeCircle(center geometry.Point, radius float32, color widget.Color, strokeWidth float32) {
	c.StrokeCircles = append(c.StrokeCircles, StrokeCircleCall{
		Center: center, Radius: radius, Color: color, StrokeWidth: strokeWidth,
	})
}

// DrawLine records the from, to, color, and stroke width arguments.
func (c *MockCanvas) DrawLine(from, to geometry.Point, color widget.Color, strokeWidth float32) {
	c.Lines = append(c.Lines, DrawLineCall{From: from, To: to, Color: color, StrokeWidth: strokeWidth})
}

// DrawText records all text drawing arguments.
func (c *MockCanvas) DrawText(text string, bounds geometry.Rect, fontSize float32, color widget.Color, bold bool, align widget.TextAlign) {
	c.Texts = append(c.Texts, DrawTextCall{
		Text: text, Bounds: bounds, FontSize: fontSize, Color: color, Bold: bold, Align: align,
	})
}

// MeasureText returns an approximate width for the given text.
// Uses average character width (0.5 * fontSize) for test predictability.
func (c *MockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}

// DrawImage records the image and position arguments.
func (c *MockCanvas) DrawImage(img image.Image, at geometry.Point) {
	c.Images = append(c.Images, DrawImageCall{Image: img, At: at})
}

// PushClip records the clip rectangle.
func (c *MockCanvas) PushClip(r geometry.Rect) {
	c.Clips = append(c.Clips, r)
}

// PushClipRoundRect records the clip rounded rectangle and radius.
func (c *MockCanvas) PushClipRoundRect(r geometry.Rect, radius float32) {
	c.ClipRoundRects = append(c.ClipRoundRects, ClipRoundRectCall{Bounds: r, Radius: radius})
}

// PopClip increments the pop clip counter.
func (c *MockCanvas) PopClip() {
	c.PopClipCount++
}

// PushTransform records the translation offset and updates the cumulative offset.
func (c *MockCanvas) PushTransform(offset geometry.Point) {
	c.Transforms = append(c.Transforms, offset)
	c.transformStack = append(c.transformStack, c.currentOffset)
	c.currentOffset = c.currentOffset.Add(offset)
}

// PopTransform restores the previous transform offset.
func (c *MockCanvas) PopTransform() {
	c.PopTransformCount++
	if len(c.transformStack) > 0 {
		lastIdx := len(c.transformStack) - 1
		c.currentOffset = c.transformStack[lastIdx]
		c.transformStack = c.transformStack[:lastIdx]
	}
}

// TransformOffset returns the current cumulative transform offset.
func (c *MockCanvas) TransformOffset() geometry.Point {
	return c.currentOffset
}

// Reset clears all recorded calls, returning the canvas to its initial state.
func (c *MockCanvas) Reset() {
	c.Clears = c.Clears[:0]
	c.Rects = c.Rects[:0]
	c.StrokeRects = c.StrokeRects[:0]
	c.RoundRects = c.RoundRects[:0]
	c.StrokeRoundRects = c.StrokeRoundRects[:0]
	c.Circles = c.Circles[:0]
	c.StrokeCircles = c.StrokeCircles[:0]
	c.Lines = c.Lines[:0]
	c.Texts = c.Texts[:0]
	c.Images = c.Images[:0]
	c.Clips = c.Clips[:0]
	c.ClipRoundRects = c.ClipRoundRects[:0]
	c.Transforms = c.Transforms[:0]
	c.PopClipCount = 0
	c.PopTransformCount = 0
	c.transformStack = c.transformStack[:0]
	c.currentOffset = geometry.Point{}
}

// TotalDrawCalls returns the total number of recorded draw operations
// (excluding clip and transform operations).
func (c *MockCanvas) TotalDrawCalls() int {
	return len(c.Clears) + len(c.Rects) + len(c.StrokeRects) +
		len(c.RoundRects) + len(c.StrokeRoundRects) +
		len(c.Circles) + len(c.StrokeCircles) +
		len(c.Lines) + len(c.Texts) + len(c.Images)
}

// Compile-time interface check.
var _ widget.Canvas = (*MockCanvas)(nil)
