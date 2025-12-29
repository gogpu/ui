package render

import (
	"github.com/gogpu/gg"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// Renderer orchestrates the render cycle for the UI framework.
//
// Renderer manages:
//   - Frame begin/end lifecycle
//   - Canvas creation and management
//   - Surface abstraction for different backends
//
// Renderer is NOT thread-safe. All operations must occur on the main/UI thread.
type Renderer struct {
	width  int
	height int

	// gg context for software rendering
	ctx *gg.Context

	// Canvas wrapper
	canvas *Canvas

	// Frame state
	inFrame bool
}

// NewRenderer creates a new Renderer with the given dimensions.
//
// The dimensions specify the render target size in logical pixels.
func NewRenderer(width, height int) *Renderer {
	ctx := gg.NewContext(width, height)
	return &Renderer{
		width:  width,
		height: height,
		ctx:    ctx,
		canvas: NewCanvas(ctx, width, height),
	}
}

// Width returns the renderer width in logical pixels.
func (r *Renderer) Width() int {
	return r.width
}

// Height returns the renderer height in logical pixels.
func (r *Renderer) Height() int {
	return r.height
}

// Resize changes the renderer dimensions.
//
// This recreates the underlying context and canvas. Any in-progress
// frame is implicitly ended.
func (r *Renderer) Resize(width, height int) {
	if r.width == width && r.height == height {
		return
	}

	// End any in-progress frame
	r.inFrame = false

	// Close old context
	if r.ctx != nil {
		_ = r.ctx.Close() // Ignore error, just cleanup
	}

	// Create new context and canvas
	r.width = width
	r.height = height
	r.ctx = gg.NewContext(width, height)
	r.canvas = NewCanvas(r.ctx, width, height)
}

// BeginFrame starts a new render frame.
//
// This must be called before any drawing operations. It clears the canvas
// with the given background color and resets the canvas state.
//
// Returns the Canvas to use for drawing.
func (r *Renderer) BeginFrame(background widget.Color) *Canvas {
	if r.inFrame {
		// Already in a frame, just reset and continue
		r.canvas.Reset()
	}

	r.inFrame = true
	r.canvas.Reset()
	r.canvas.Clear(background)

	return r.canvas
}

// EndFrame finishes the current render frame.
//
// This should be called after all drawing operations are complete.
// It returns the gg.Context which can be used to extract the rendered image.
func (r *Renderer) EndFrame() *gg.Context {
	r.inFrame = false
	return r.ctx
}

// Canvas returns the current Canvas.
//
// This is a convenience method for cases where the Canvas is needed
// outside of the BeginFrame/EndFrame cycle.
func (r *Renderer) Canvas() *Canvas {
	return r.canvas
}

// Context returns the underlying gg.Context.
//
// This is provided for advanced use cases where direct access is needed.
func (r *Renderer) Context() *gg.Context {
	return r.ctx
}

// InFrame returns true if currently between BeginFrame and EndFrame.
func (r *Renderer) InFrame() bool {
	return r.inFrame
}

// Close releases resources associated with the Renderer.
//
// After Close, the Renderer should not be used.
func (r *Renderer) Close() error {
	r.inFrame = false
	if r.ctx != nil {
		return r.ctx.Close()
	}
	return nil
}

// RenderConfig contains configuration options for rendering.
type RenderConfig struct {
	// BackgroundColor is the color used to clear the canvas at frame start.
	BackgroundColor widget.Color

	// Scale is the display scale factor (1.0 for standard, 2.0 for HiDPI).
	Scale float32
}

// DefaultRenderConfig returns the default render configuration.
func DefaultRenderConfig() RenderConfig {
	return RenderConfig{
		BackgroundColor: widget.ColorWhite,
		Scale:           1.0,
	}
}

// RenderTarget represents an abstract render target.
//
// This interface allows the renderer to work with different output targets
// (software buffer, GPU texture, etc.).
type RenderTarget interface {
	// Width returns the target width in pixels.
	Width() int

	// Height returns the target height in pixels.
	Height() int

	// Bounds returns the target bounds as a geometry.Rect.
	Bounds() geometry.Rect
}

// SoftwareTarget is a RenderTarget backed by a software buffer (gg.Context).
type SoftwareTarget struct {
	ctx *gg.Context
}

// NewSoftwareTarget creates a new software render target.
func NewSoftwareTarget(width, height int) *SoftwareTarget {
	return &SoftwareTarget{
		ctx: gg.NewContext(width, height),
	}
}

// Width returns the target width.
func (t *SoftwareTarget) Width() int {
	return t.ctx.Width()
}

// Height returns the target height.
func (t *SoftwareTarget) Height() int {
	return t.ctx.Height()
}

// Bounds returns the target bounds.
func (t *SoftwareTarget) Bounds() geometry.Rect {
	return geometry.NewRect(0, 0, float32(t.ctx.Width()), float32(t.ctx.Height()))
}

// Context returns the underlying gg.Context.
func (t *SoftwareTarget) Context() *gg.Context {
	return t.ctx
}

// Close releases resources.
func (t *SoftwareTarget) Close() error {
	return t.ctx.Close()
}

// Verify SoftwareTarget implements RenderTarget.
var _ RenderTarget = (*SoftwareTarget)(nil)
