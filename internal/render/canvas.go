package render

import (
	"image"
	stdcolor "image/color"
	"math"
	"sync"

	"github.com/gogpu/gg"
	"github.com/gogpu/gg/scene"
	"github.com/gogpu/gg/svg"
	"github.com/gogpu/gg/text"
	"github.com/gogpu/gpucontext"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/internal/render/fonts"
	"github.com/gogpu/ui/widget"
)

// Canvas implements [widget.Canvas] using gogpu/gg as the 2D drawing backend.
//
// Canvas wraps a gg.Context and provides all drawing operations required by
// the widget system. It manages clip and transform stacks internally.
//
// Canvas is NOT thread-safe. All drawing operations must occur on the main/UI
// thread during the Draw phase.
type Canvas struct {
	dc     *gg.Context
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
func NewCanvas(dc *gg.Context, width, height int) *Canvas {
	// NOTE: SetLCDLayout removed — it clears the shared GlyphMask atlas,
	// breaking GPU text in offscreen contexts (RepaintBoundary).
	// LCD layout should be set once at app init if needed.

	return &Canvas{
		dc:             dc,
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
	return c.dc
}

// Clear fills the entire canvas with the given color.
func (c *Canvas) Clear(color widget.Color) {
	c.dc.ClearWithColor(ToGGColor(color))
}

// DrawRect fills a rectangle with the given color.
func (c *Canvas) DrawRect(r geometry.Rect, color widget.Color) {
	// Apply current transform
	r = c.applyTransform(r)

	// Skip if outside clip bounds
	if !c.isVisible(r) {
		return
	}

	c.dc.SetRGBA(float64(color.R), float64(color.G), float64(color.B), float64(color.A))
	c.dc.DrawRectangle(
		float64(r.Min.X),
		float64(r.Min.Y),
		float64(r.Width()),
		float64(r.Height()),
	)
	c.dc.Fill()
	c.dc.ClearPath()
}

// FillRectDirect fills a rectangle directly on the CPU pixmap, bypassing the
// GPU shape accelerator. This prevents SDF shapes from being queued on the
// compositor canvas, enabling the non-MSAA blit-only fast path (ADR-016).
func (c *Canvas) FillRectDirect(r geometry.Rect, color widget.Color) {
	r = c.applyTransform(r)
	if !c.isVisible(r) {
		return
	}
	c.dc.FillRectCPU(
		float64(r.Min.X), float64(r.Min.Y),
		float64(r.Width()), float64(r.Height()),
		gg.RGBA{R: float64(color.R), G: float64(color.G), B: float64(color.B), A: float64(color.A)},
	)
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

	c.dc.SetRGBA(float64(color.R), float64(color.G), float64(color.B), float64(color.A))
	c.dc.SetLineWidth(float64(strokeWidth))
	c.dc.DrawRectangle(
		float64(r.Min.X),
		float64(r.Min.Y),
		float64(r.Width()),
		float64(r.Height()),
	)
	c.dc.Stroke()
	c.dc.ClearPath()
}

// DrawRoundRect fills a rounded rectangle with the given color.
func (c *Canvas) DrawRoundRect(r geometry.Rect, color widget.Color, radius float32) {
	// Apply current transform
	r = c.applyTransform(r)

	// Skip if outside clip bounds
	if !c.isVisible(r) {
		return
	}

	c.dc.SetRGBA(float64(color.R), float64(color.G), float64(color.B), float64(color.A))
	c.dc.DrawRoundedRectangle(
		float64(r.Min.X),
		float64(r.Min.Y),
		float64(r.Width()),
		float64(r.Height()),
		float64(radius),
	)
	c.dc.Fill()
	c.dc.ClearPath()
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

	c.dc.SetRGBA(float64(color.R), float64(color.G), float64(color.B), float64(color.A))
	c.dc.SetLineWidth(float64(strokeWidth))
	c.dc.DrawRoundedRectangle(
		float64(r.Min.X),
		float64(r.Min.Y),
		float64(r.Width()),
		float64(r.Height()),
		float64(radius),
	)
	c.dc.Stroke()
	c.dc.ClearPath()
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

	c.dc.SetRGBA(float64(color.R), float64(color.G), float64(color.B), float64(color.A))
	c.dc.DrawCircle(float64(center.X), float64(center.Y), float64(radius))
	c.dc.Fill()
	c.dc.ClearPath()
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

	c.dc.SetRGBA(float64(color.R), float64(color.G), float64(color.B), float64(color.A))
	c.dc.SetLineWidth(float64(strokeWidth))
	c.dc.DrawCircle(float64(center.X), float64(center.Y), float64(radius))
	c.dc.Stroke()
	c.dc.ClearPath()
}

// StrokeArc draws a circular arc outline from startAngle with the given sweep.
// Angles are in radians. startAngle=0 is 3 o'clock, positive is counterclockwise.
// Uses gg.DrawArc which approximates arcs with cubic Bézier curves.
func (c *Canvas) StrokeArc(center geometry.Point, radius float32,
	startAngle, sweepAngle float64, color widget.Color, strokeWidth float32) {
	if sweepAngle == 0 {
		return
	}

	center = c.applyTransformPoint(center)

	// Visibility check using the arc bounding box (conservative: full circle).
	bounds := geometry.NewRect(
		center.X-radius-strokeWidth/2,
		center.Y-radius-strokeWidth/2,
		(radius+strokeWidth/2)*2,
		(radius+strokeWidth/2)*2,
	)
	if !c.isVisible(bounds) {
		return
	}

	c.dc.SetRGBA(float64(color.R), float64(color.G), float64(color.B), float64(color.A))
	c.dc.SetLineWidth(float64(strokeWidth))
	// gg.DrawArc takes absolute start and end angles.
	c.dc.DrawArc(float64(center.X), float64(center.Y), float64(radius),
		startAngle, startAngle+sweepAngle)
	c.dc.Stroke()
	c.dc.ClearPath()
}

// StrokeArcStyled draws a circular arc with the specified line cap style.
func (c *Canvas) StrokeArcStyled(center geometry.Point, radius float32,
	startAngle, sweepAngle float64, color widget.Color, strokeWidth float32, lineCap widget.LineCap) {
	if sweepAngle == 0 {
		return
	}

	center = c.applyTransformPoint(center)

	bounds := geometry.NewRect(
		center.X-radius-strokeWidth/2,
		center.Y-radius-strokeWidth/2,
		(radius+strokeWidth/2)*2,
		(radius+strokeWidth/2)*2,
	)
	if !c.isVisible(bounds) {
		return
	}

	c.dc.SetRGBA(float64(color.R), float64(color.G), float64(color.B), float64(color.A))
	c.dc.SetLineWidth(float64(strokeWidth))
	oldCap := c.dc.GetStroke().Cap
	c.dc.SetLineCap(toGGLineCap(lineCap))
	c.dc.DrawArc(float64(center.X), float64(center.Y), float64(radius),
		startAngle, startAngle+sweepAngle)
	c.dc.Stroke()
	c.dc.ClearPath()
	c.dc.SetLineCap(oldCap)
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

	c.dc.SetRGBA(float64(color.R), float64(color.G), float64(color.B), float64(color.A))
	c.dc.SetLineWidth(float64(strokeWidth))
	c.dc.DrawLine(float64(from.X), float64(from.Y), float64(to.X), float64(to.Y))
	c.dc.Stroke()
	c.dc.ClearPath()
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

	// Set clip on gg context. Currently gg uses this for CPU-side
	// ClipCoverage masking. Once gg implements GPU scissor rect (Phase 1),
	// this will also drive hardware scissor for GPU rendering.
	c.dc.Push()
	clip := c.currentClip
	c.dc.ClipRect(float64(clip.Min.X), float64(clip.Min.Y),
		float64(clip.Width()), float64(clip.Height()))
}

// PushClipRoundRect pushes a rounded rectangle clipping region.
//
// Uses gg.ClipRoundRect which activates GPU SDF-based clipping for
// rounded rectangles. All subsequent draw operations will be clipped
// to the rounded rect shape.
func (c *Canvas) PushClipRoundRect(r geometry.Rect, radius float32) {
	r = c.applyTransform(r)
	c.clipStack = append(c.clipStack, c.currentClip)
	c.currentClip = c.currentClip.Intersection(r)

	c.dc.Push()
	c.dc.ClipRoundRect(
		float64(r.Min.X), float64(r.Min.Y),
		float64(r.Width()), float64(r.Height()),
		float64(radius),
	)
}

// PopClip removes the most recently pushed clipping region.
//
// Must be called for each PushClip or PushClipRoundRect call.
func (c *Canvas) PopClip() {
	if len(c.clipStack) == 0 {
		return
	}

	// Restore previous clip bounds
	lastIdx := len(c.clipStack) - 1
	c.currentClip = c.clipStack[lastIdx]
	c.clipStack = c.clipStack[:lastIdx]

	// Restore gg context state
	c.dc.Pop()
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

// ScreenOriginBase returns the screen-space base offset for this canvas.
// For the main window canvas this is always (0,0).
func (c *Canvas) ScreenOriginBase() geometry.Point { return geometry.Point{} }

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
		defaultRegular, _ = text.NewFontSource(fonts.InterRegular)
		defaultBold, _ = text.NewFontSource(fonts.InterBold)
	})
}

// DrawText draws text within the given bounding rectangle.
func (c *Canvas) DrawText(s string, bounds geometry.Rect, fontSize float32, color widget.Color, bold bool, align widget.TextAlign) {
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
	c.dc.SetFont(face)
	c.dc.SetRGBA(float64(color.R), float64(color.G), float64(color.B), float64(color.A))

	// Calculate baseline Y by centering text vertically within bounds.
	// Round to pixel grid for crisp text rendering.
	metrics := face.Metrics()
	textHeight := metrics.Ascent + metrics.Descent
	baselineY := math.Round(float64(bounds.Min.Y) + (float64(bounds.Height())-textHeight)/2 + metrics.Ascent)

	// Calculate x position based on alignment.
	// Round to pixel grid.
	w, _ := c.dc.MeasureString(s)
	available := float64(bounds.Width())
	x := float64(bounds.Min.X)
	if w < available {
		x += (available - w) * align.Float64()
	}
	x = math.Round(x)

	c.dc.DrawString(s, x, baselineY)
}

// MeasureText returns the width in pixels of the given text string
// when rendered at the specified font size and weight.
func (c *Canvas) MeasureText(s string, fontSize float32, bold bool) float32 {
	if s == "" {
		return 0
	}

	ensureDefaultFonts()

	source := defaultRegular
	if bold {
		source = defaultBold
	}
	if source == nil {
		return float32(len([]rune(s))) * fontSize * 0.5
	}

	face := source.Face(float64(fontSize))
	c.dc.SetFont(face)
	w, _ := c.dc.MeasureString(s)
	return float32(w)
}

// DrawStyledText draws text within the given bounding rectangle using
// a font resolved from the global [FontRegistry]. This enables rendering
// with custom fonts (CJK, icon fonts, etc.) loaded via the plugin system.
//
// If the FontFamily is empty or not found, the default Inter font is used.
func (c *Canvas) DrawStyledText(s string, bounds geometry.Rect, style widget.TextStyle) {
	if s == "" {
		return
	}

	bounds = c.applyTransform(bounds)
	if !c.isVisible(bounds) {
		return
	}

	source := c.resolveStyledFont(style)
	if source == nil {
		return
	}

	face := source.Face(float64(style.FontSize))
	c.dc.SetFont(face)
	c.dc.SetRGBA(float64(style.Color.R), float64(style.Color.G), float64(style.Color.B), float64(style.Color.A))

	// Calculate baseline Y by centering text vertically within bounds.
	metrics := face.Metrics()
	textHeight := metrics.Ascent + metrics.Descent
	baselineY := math.Round(float64(bounds.Min.Y) + (float64(bounds.Height())-textHeight)/2 + metrics.Ascent)

	// Calculate x position based on alignment.
	w, _ := c.dc.MeasureString(s)
	available := float64(bounds.Width())
	x := float64(bounds.Min.X)
	if w < available {
		x += (available - w) * style.Align.Float64()
	}
	x = math.Round(x)

	c.dc.DrawString(s, x, baselineY)
}

// MeasureStyledText returns the width in pixels of the given text string
// when rendered with the specified [widget.TextStyle]. The font is resolved
// from the global [FontRegistry].
func (c *Canvas) MeasureStyledText(s string, style widget.TextStyle) float32 {
	if s == "" {
		return 0
	}

	source := c.resolveStyledFont(style)
	if source == nil {
		return float32(len([]rune(s))) * style.FontSize * 0.5
	}

	face := source.Face(float64(style.FontSize))
	c.dc.SetFont(face)
	w, _ := c.dc.MeasureString(s)
	return float32(w)
}

// resolveStyledFont resolves a [text.FontSource] from a [widget.TextStyle]
// using the global [FontRegistry]. Falls back to the default Inter font
// when the family is empty or not found.
func (c *Canvas) resolveStyledFont(style widget.TextStyle) *text.FontSource {
	return resolveStyledFontSource(style)
}

// DrawImage draws an image at the specified position.
//
// The image is composited using source-over blending via gg.DrawImage.
// The position is adjusted by the current transform offset and snapped
// to the pixel grid.
func (c *Canvas) DrawImage(img image.Image, at geometry.Point) {
	if img == nil {
		return
	}

	// Apply current transform and snap to pixel grid.
	at = c.applyTransformPoint(at)

	// Create bounding rect for visibility check.
	bounds := img.Bounds()
	imgW := float32(bounds.Dx())
	imgH := float32(bounds.Dy())
	drawRect := geometry.NewRect(at.X, at.Y, imgW, imgH)
	if !c.isVisible(drawRect) {
		return
	}

	// Convert image.Image to gg.ImageBuf and draw via gg.Context.
	buf := gg.ImageBufFromImage(img)
	c.dc.DrawImage(buf, float64(at.X), float64(at.Y))
}

// DrawGPUTexture composites a pre-existing GPU texture view as a textured quad
// at the given position. The texture is drawn with dimensions width x height.
//
// This is the GPU layer compositing path used by RepaintBoundary — zero CPU
// readback, zero pixel upload. The view must originate from the same GPU device
// (typically obtained via gg.Context.CreateOffscreenTexture).
//
// The position is adjusted by the current transform offset and snapped to the
// pixel grid, consistent with DrawImage behavior.
func (c *Canvas) DrawGPUTexture(view gpucontext.TextureView, x, y float64, width, height int) {
	// Apply canvas transform offset.
	offset := c.currentOffset
	x += float64(offset.X)
	y += float64(offset.Y)

	// Snap to pixel grid.
	x = math.Round(x)
	y = math.Round(y)

	c.dc.DrawGPUTexture(view, x, y, width, height)
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

// FillSVGPath fills an SVG path within the given bounds.
func (c *Canvas) FillSVGPath(svgData string, viewBox float32, bounds geometry.Rect, color widget.Color) {
	if svgData == "" || viewBox <= 0 {
		return
	}

	bounds = c.applyTransform(bounds)
	if !c.isVisible(bounds) {
		return
	}

	path, err := gg.ParseSVGPath(svgData)
	if err != nil {
		return
	}

	// Scale path to fit bounds.
	scale := float64(bounds.Width()) / float64(viewBox)
	scaleY := float64(bounds.Height()) / float64(viewBox)
	if scaleY < scale {
		scale = scaleY
	}

	c.dc.Push()
	c.dc.Translate(float64(bounds.Min.X), float64(bounds.Min.Y))
	c.dc.Scale(scale, scale)
	c.dc.SetRGBA(float64(color.R), float64(color.G), float64(color.B), float64(color.A))
	c.dc.SetFillRule(gg.FillRuleEvenOdd)
	c.dc.FillPath(path)
	c.dc.Pop()
}

// ReplayScene renders a previously recorded scene.Scene display list into
// this canvas. The scene commands are decoded and routed through gg.Context's
// GPU accelerator, which auto-selects GPU or CPU rendering per shape.
//
// This is the retained-mode replay path (ADR-007): RepaintBoundary caches
// child drawing as a scene.Scene and replays it on cache hit instead of
// re-executing child.Draw().
func (c *Canvas) ReplayScene(s *scene.Scene) {
	if s == nil || s.IsEmpty() {
		return
	}
	// Apply current canvas offset so the scene renders at the correct
	// position within the parent. The scene was recorded in local
	// coordinates (0,0 based by SceneCanvas). We translate the gg.Context
	// to the current offset before replay, then restore after.
	//
	// We don't use GPUSceneRenderer here because it calls dc.Identity()
	// on TagTransform, which resets the parent transform. Instead, use
	// scene.Renderer (tile-parallel with GPU auto-select) which renders
	// the scene to a pixmap, then blit the result.
	//
	// For now, use the direct GPU renderer with Push/Pop + Translate
	// to position the scene output correctly.
	c.dc.Push()
	ox := float64(c.currentOffset.X)
	oy := float64(c.currentOffset.Y)
	c.dc.Translate(ox, oy)
	renderer := scene.NewGPUSceneRenderer(c.dc)
	_ = renderer.RenderScene(s)
	c.dc.Pop()
}

// SetDamageTracking enables or disables damage tracking on the underlying gg.Context.
// Implements widget.DamageController for retained-mode optimization.
func (c *Canvas) SetDamageTracking(enabled bool) {
	c.dc.SetDamageTracking(enabled)
}

// RenderSVG renders full SVG XML within the given bounds with color override.
// Uses gg/svg.Document.RenderToWithColor to draw directly into the gg.Context.
func (c *Canvas) RenderSVG(svgXML []byte, bounds geometry.Rect, color widget.Color) {
	if len(svgXML) == 0 {
		return
	}

	bounds = c.applyTransform(bounds)
	if !c.isVisible(bounds) {
		return
	}

	doc, err := svg.Parse(svgXML)
	if err != nil {
		return
	}

	r8, g8, b8, a8 := color.RGBA8()
	svgColor := stdcolor.NRGBA{R: r8, G: g8, B: b8, A: a8}
	doc.RenderToWithColor(c.dc, float64(bounds.Min.X), float64(bounds.Min.Y),
		float64(bounds.Width()), float64(bounds.Height()), svgColor)
}

// SetTextMode sets the text rendering strategy on the underlying gg.Context.
// Implements widget.TextModeController.
func (c *Canvas) SetTextMode(mode widget.TextMode) {
	c.dc.SetTextMode(gg.TextMode(mode))
}

// TextMode returns the current text rendering strategy.
// Implements widget.TextModeController.
func (c *Canvas) TextMode() widget.TextMode {
	return widget.TextMode(c.dc.TextMode())
}

// toGGLineCap converts widget.LineCap to gg.LineCap.
func toGGLineCap(lc widget.LineCap) gg.LineCap {
	switch lc {
	case widget.LineCapRound:
		return gg.LineCapRound
	case widget.LineCapSquare:
		return gg.LineCapSquare
	default:
		return gg.LineCapButt
	}
}

// Verify Canvas implements widget.Canvas.
var _ widget.Canvas = (*Canvas)(nil)

// Verify Canvas implements widget.ArcStroker.
var _ widget.ArcStroker = (*Canvas)(nil)

// Verify Canvas implements widget.StyledTextDrawer.
var _ widget.StyledTextDrawer = (*Canvas)(nil)
