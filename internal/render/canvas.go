package render

import (
	"math"
	"sync"

	"github.com/gogpu/gg"
	"github.com/gogpu/gg/text"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
)

// Canvas implements [widget.Canvas] using gogpu/gg as the 2D drawing backend.
//
// Canvas wraps a gg.Context and provides all drawing operations required by
// the widget system. It manages clip and transform stacks internally.
//
// Canvas is NOT thread-safe. All drawing operations must occur on the main/UI
// thread during the Draw phase.
type Canvas struct {
	ctx    *gg.Context
	width  int
	height int

	// Clip stack: stores previous clip bounds for each PushClip
	clipStack []geometry.Rect
	// Current clip bounds (intersection of all pushed clips)
	currentClip geometry.Rect

	// Transform stack: stores cumulative offsets for each PushTransform
	transformStack []geometry.Point
	// Current cumulative transform offset
	currentOffset geometry.Point
}

// NewCanvas creates a new Canvas wrapping the given gg.Context.
//
// The width and height specify the canvas dimensions in logical pixels.
// The gg.Context should already be created with matching dimensions.
func NewCanvas(ctx *gg.Context, width, height int) *Canvas {
	return &Canvas{
		ctx:            ctx,
		width:          width,
		height:         height,
		clipStack:      make([]geometry.Rect, 0, 8),
		currentClip:    geometry.NewRect(0, 0, float32(width), float32(height)),
		transformStack: make([]geometry.Point, 0, 8),
		currentOffset:  geometry.Point{},
	}
}

// Width returns the canvas width in logical pixels.
func (c *Canvas) Width() int {
	return c.width
}

// Height returns the canvas height in logical pixels.
func (c *Canvas) Height() int {
	return c.height
}

// Context returns the underlying gg.Context.
//
// This is provided for advanced use cases where direct access to gg
// functionality is needed. Use with caution as it bypasses Canvas state.
func (c *Canvas) Context() *gg.Context {
	return c.ctx
}

// Clear fills the entire canvas with the given color.
func (c *Canvas) Clear(color widget.Color) {
	c.ctx.ClearWithColor(ToGGColor(color))
}

// DrawRect fills a rectangle with the given color.
func (c *Canvas) DrawRect(r geometry.Rect, color widget.Color) {
	// Apply current transform
	r = c.applyTransform(r)

	// Skip if outside clip bounds
	if !c.isVisible(r) {
		return
	}

	c.ctx.SetRGBA(float64(color.R), float64(color.G), float64(color.B), float64(color.A))
	c.ctx.DrawRectangle(
		float64(r.Min.X),
		float64(r.Min.Y),
		float64(r.Width()),
		float64(r.Height()),
	)
	c.ctx.Fill()
	c.ctx.ClearPath()
}

// StrokeRect draws the outline of a rectangle.
func (c *Canvas) StrokeRect(r geometry.Rect, color widget.Color, strokeWidth float32) {
	// Apply current transform
	r = c.applyTransform(r)

	// Skip if outside clip bounds (with stroke width consideration)
	expanded := r.Expand(strokeWidth / 2)
	if !c.isVisible(expanded) {
		return
	}

	c.ctx.SetRGBA(float64(color.R), float64(color.G), float64(color.B), float64(color.A))
	c.ctx.SetLineWidth(float64(strokeWidth))
	c.ctx.DrawRectangle(
		float64(r.Min.X),
		float64(r.Min.Y),
		float64(r.Width()),
		float64(r.Height()),
	)
	c.ctx.Stroke()
	c.ctx.ClearPath()
}

// DrawRoundRect fills a rounded rectangle with the given color.
func (c *Canvas) DrawRoundRect(r geometry.Rect, color widget.Color, radius float32) {
	// Apply current transform
	r = c.applyTransform(r)

	// Skip if outside clip bounds
	if !c.isVisible(r) {
		return
	}

	c.ctx.SetRGBA(float64(color.R), float64(color.G), float64(color.B), float64(color.A))
	c.ctx.DrawRoundedRectangle(
		float64(r.Min.X),
		float64(r.Min.Y),
		float64(r.Width()),
		float64(r.Height()),
		float64(radius),
	)
	c.ctx.Fill()
	c.ctx.ClearPath()
}

// StrokeRoundRect draws the outline of a rounded rectangle.
func (c *Canvas) StrokeRoundRect(r geometry.Rect, color widget.Color, radius float32, strokeWidth float32) {
	// Apply current transform
	r = c.applyTransform(r)

	// Skip if outside clip bounds (with stroke width consideration)
	expanded := r.Expand(strokeWidth / 2)
	if !c.isVisible(expanded) {
		return
	}

	c.ctx.SetRGBA(float64(color.R), float64(color.G), float64(color.B), float64(color.A))
	c.ctx.SetLineWidth(float64(strokeWidth))
	c.ctx.DrawRoundedRectangle(
		float64(r.Min.X),
		float64(r.Min.Y),
		float64(r.Width()),
		float64(r.Height()),
		float64(radius),
	)
	c.ctx.Stroke()
	c.ctx.ClearPath()
}

// DrawCircle fills a circle with the given color.
func (c *Canvas) DrawCircle(center geometry.Point, radius float32, color widget.Color) {
	// Apply current transform to center
	center = c.applyTransformPoint(center)

	// Create bounding rect for visibility check
	bounds := geometry.NewRect(
		center.X-radius,
		center.Y-radius,
		radius*2,
		radius*2,
	)
	if !c.isVisible(bounds) {
		return
	}

	c.ctx.SetRGBA(float64(color.R), float64(color.G), float64(color.B), float64(color.A))
	c.ctx.DrawCircle(float64(center.X), float64(center.Y), float64(radius))
	c.ctx.Fill()
	c.ctx.ClearPath()
}

// StrokeCircle draws the outline of a circle.
func (c *Canvas) StrokeCircle(center geometry.Point, radius float32, color widget.Color, strokeWidth float32) {
	// Apply current transform to center
	center = c.applyTransformPoint(center)

	// Create bounding rect for visibility check (with stroke consideration)
	bounds := geometry.NewRect(
		center.X-radius-strokeWidth/2,
		center.Y-radius-strokeWidth/2,
		(radius+strokeWidth/2)*2,
		(radius+strokeWidth/2)*2,
	)
	if !c.isVisible(bounds) {
		return
	}

	c.ctx.SetRGBA(float64(color.R), float64(color.G), float64(color.B), float64(color.A))
	c.ctx.SetLineWidth(float64(strokeWidth))
	c.ctx.DrawCircle(float64(center.X), float64(center.Y), float64(radius))
	c.ctx.Stroke()
	c.ctx.ClearPath()
}

// DrawLine draws a line between two points.
func (c *Canvas) DrawLine(from, to geometry.Point, color widget.Color, strokeWidth float32) {
	// Apply current transform
	from = c.applyTransformPoint(from)
	to = c.applyTransformPoint(to)

	// Create bounding rect for visibility check
	bounds := geometry.FromMinMax(from, to).Expand(strokeWidth / 2)
	if !c.isVisible(bounds) {
		return
	}

	c.ctx.SetRGBA(float64(color.R), float64(color.G), float64(color.B), float64(color.A))
	c.ctx.SetLineWidth(float64(strokeWidth))
	c.ctx.DrawLine(float64(from.X), float64(from.Y), float64(to.X), float64(to.Y))
	c.ctx.Stroke()
	c.ctx.ClearPath()
}

// PushClip pushes a clipping rectangle onto the clip stack.
//
// All subsequent drawing operations will be clipped to this rectangle
// intersected with any parent clip rectangles.
func (c *Canvas) PushClip(r geometry.Rect) {
	// Apply current transform to clip rect
	r = c.applyTransform(r)

	// Save current clip bounds
	c.clipStack = append(c.clipStack, c.currentClip)

	// Compute new clip as intersection with current
	c.currentClip = c.currentClip.Intersection(r)

	// Apply clip to gg context using Push/Pop state
	c.ctx.Push()

	// Draw clip rectangle path and apply as clip
	// Note: gg doesn't have a direct Clip() method, so we work around it
	// by checking visibility in each draw operation
}

// PopClip removes the most recently pushed clipping rectangle.
//
// Must be called for each PushClip call.
func (c *Canvas) PopClip() {
	if len(c.clipStack) == 0 {
		return
	}

	// Restore previous clip bounds
	lastIdx := len(c.clipStack) - 1
	c.currentClip = c.clipStack[lastIdx]
	c.clipStack = c.clipStack[:lastIdx]

	// Restore gg context state
	c.ctx.Pop()
}

// PushTransform pushes a translation transform onto the transform stack.
//
// All subsequent drawing operations will be offset by the given point.
func (c *Canvas) PushTransform(offset geometry.Point) {
	// Save current offset
	c.transformStack = append(c.transformStack, c.currentOffset)

	// Add new offset to current cumulative offset
	c.currentOffset = c.currentOffset.Add(offset)
}

// PopTransform removes the most recently pushed transform.
//
// Must be called for each PushTransform call.
func (c *Canvas) PopTransform() {
	if len(c.transformStack) == 0 {
		return
	}

	// Restore previous offset
	lastIdx := len(c.transformStack) - 1
	c.currentOffset = c.transformStack[lastIdx]
	c.transformStack = c.transformStack[:lastIdx]
}

// ClipBounds returns the current clip bounds.
//
// This is useful for widgets that want to optimize drawing by skipping
// elements outside the visible area.
func (c *Canvas) ClipBounds() geometry.Rect {
	return c.currentClip
}

// TransformOffset returns the current cumulative transform offset.
//
// This is useful for converting between local and global coordinates.
func (c *Canvas) TransformOffset() geometry.Point {
	return c.currentOffset
}

// ClipDepth returns the current depth of the clip stack.
func (c *Canvas) ClipDepth() int {
	return len(c.clipStack)
}

// TransformDepth returns the current depth of the transform stack.
func (c *Canvas) TransformDepth() int {
	return len(c.transformStack)
}

// Reset clears the clip and transform stacks, restoring initial state.
func (c *Canvas) Reset() {
	c.clipStack = c.clipStack[:0]
	c.currentClip = geometry.NewRect(0, 0, float32(c.width), float32(c.height))
	c.transformStack = c.transformStack[:0]
	c.currentOffset = geometry.Point{}
}

// default font sources, loaded lazily.
var (
	defaultFontsOnce sync.Once
	defaultRegular   *text.FontSource
	defaultBold      *text.FontSource
)

func ensureDefaultFonts() {
	defaultFontsOnce.Do(func() {
		defaultRegular, _ = text.NewFontSource(goregular.TTF)
		defaultBold, _ = text.NewFontSource(gobold.TTF)
	})
}

// DrawText draws text within the given bounding rectangle.
func (c *Canvas) DrawText(s string, bounds geometry.Rect, fontSize float32, color widget.Color, bold bool, align float32) {
	if s == "" {
		return
	}

	bounds = c.applyTransform(bounds)
	if !c.isVisible(bounds) {
		return
	}

	ensureDefaultFonts()

	source := defaultRegular
	if bold {
		source = defaultBold
	}
	if source == nil {
		return
	}

	face := source.Face(float64(fontSize))
	c.ctx.SetFont(face)
	c.ctx.SetRGBA(float64(color.R), float64(color.G), float64(color.B), float64(color.A))

	// Calculate baseline Y by centering text vertically within bounds.
	// Round to pixel grid for crisp text rendering.
	metrics := face.Metrics()
	textHeight := metrics.Ascent + metrics.Descent
	baselineY := math.Round(float64(bounds.Min.Y) + (float64(bounds.Height())-textHeight)/2 + metrics.Ascent)

	// Calculate x position based on alignment.
	// Round to pixel grid.
	w, _ := c.ctx.MeasureString(s)
	available := float64(bounds.Width())
	x := float64(bounds.Min.X)
	if w < available {
		x += (available - w) * float64(align)
	}
	x = math.Round(x)

	c.ctx.DrawString(s, x, baselineY)
}

// snapF rounds a float32 to the nearest integer for pixel-perfect rendering.
// Sub-pixel coordinates cause blurry lines, asymmetric AA on circles, and
// fuzzy text. Snapping to the pixel grid eliminates these artifacts.
func snapF(v float32) float32 {
	return float32(math.Round(float64(v)))
}

// applyTransform applies the current transform offset to a rectangle
// and snaps the result to the pixel grid for crisp rendering.
func (c *Canvas) applyTransform(r geometry.Rect) geometry.Rect {
	r = r.Translate(c.currentOffset)
	r.Min.X = snapF(r.Min.X)
	r.Min.Y = snapF(r.Min.Y)
	r.Max.X = snapF(r.Max.X)
	r.Max.Y = snapF(r.Max.Y)
	return r
}

// applyTransformPoint applies the current transform offset to a point
// and snaps the result to the pixel grid for crisp rendering.
func (c *Canvas) applyTransformPoint(p geometry.Point) geometry.Point {
	p = p.Add(c.currentOffset)
	p.X = snapF(p.X)
	p.Y = snapF(p.Y)
	return p
}

// isVisible returns true if the rectangle intersects with the current clip bounds.
func (c *Canvas) isVisible(r geometry.Rect) bool {
	return c.currentClip.Intersects(r)
}

// Verify Canvas implements widget.Canvas.
var _ widget.Canvas = (*Canvas)(nil)
