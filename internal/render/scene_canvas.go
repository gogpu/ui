package render

import (
	"image"
	"image/draw"
	"math"

	"github.com/gogpu/gg"
	"github.com/gogpu/gg/scene"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// SceneCanvas implements [widget.Canvas] by recording drawing commands into
// a [scene.Scene]. This is used inside RepaintBoundary for tile-parallel
// rendering of widget subtrees.
//
// Shape drawing (rect, round rect, circle, line) is recorded directly into
// the scene's encoding as scene shapes. Text rendering is delegated to a
// temporary gg.Context to preserve MSDF text quality, then captured as an
// image and drawn into the scene.
//
// SceneCanvas is NOT thread-safe. All drawing operations must occur on the
// main/UI thread during the Draw phase.
type SceneCanvas struct {
	sc     *scene.Scene
	width  int
	height int

	// Text rendering: delegate to gg.Context (preserves MSDF quality).
	textDC *gg.Context

	// Clip stack: stores previous clip bounds for each PushClip.
	clipStack []geometry.Rect
	// Current clip bounds (intersection of all pushed clips).
	currentClip geometry.Rect

	// Transform stack: stores cumulative offsets for each PushTransform.
	transformStack []geometry.Point
	// Current cumulative transform offset.
	currentOffset geometry.Point
}

// NewSceneCanvas creates a new SceneCanvas that records drawing commands
// into the given scene. The width and height specify the canvas dimensions
// in logical pixels.
func NewSceneCanvas(sc *scene.Scene, width, height int) *SceneCanvas {
	return &SceneCanvas{
		sc:             sc,
		width:          width,
		height:         height,
		clipStack:      make([]geometry.Rect, 0, 8),
		currentClip:    geometry.NewRect(0, 0, float32(width), float32(height)),
		transformStack: make([]geometry.Point, 0, 8),
		currentOffset:  geometry.Point{},
	}
}

// Close releases resources held by the SceneCanvas.
// Must be called after drawing is complete to free the temporary text context.
func (c *SceneCanvas) Close() {
	if c.textDC != nil {
		_ = c.textDC.Close()
		c.textDC = nil
	}
}

// Scene returns the underlying scene.Scene.
func (c *SceneCanvas) Scene() *scene.Scene {
	return c.sc
}

// --- widget.Canvas interface ---

// Clear fills the entire canvas with the given color.
func (c *SceneCanvas) Clear(color widget.Color) {
	brush := scene.SolidBrush(ToGGColor(color))
	shape := scene.NewRectShape(0, 0, float32(c.width), float32(c.height))
	c.sc.Fill(scene.FillNonZero, scene.IdentityAffine(), brush, shape)
}

// DrawRect fills a rectangle with the given color.
func (c *SceneCanvas) DrawRect(r geometry.Rect, color widget.Color) {
	r = c.applyTransform(r)
	if !c.isVisible(r) {
		return
	}

	brush := scene.SolidBrush(ToGGColor(color))
	shape := scene.NewRectShape(r.Min.X, r.Min.Y, r.Width(), r.Height())
	c.sc.Fill(scene.FillNonZero, scene.IdentityAffine(), brush, shape)
}

// StrokeRect draws the outline of a rectangle.
func (c *SceneCanvas) StrokeRect(r geometry.Rect, color widget.Color, strokeWidth float32) {
	r = c.applyTransform(r)
	expanded := r.Expand(strokeWidth / 2)
	if !c.isVisible(expanded) {
		return
	}

	brush := scene.SolidBrush(ToGGColor(color))
	shape := scene.NewRectShape(r.Min.X, r.Min.Y, r.Width(), r.Height())
	style := &scene.StrokeStyle{
		Width:      strokeWidth,
		MiterLimit: 10.0,
		Cap:        scene.LineCapButt,
		Join:       scene.LineJoinMiter,
	}
	c.sc.Stroke(style, scene.IdentityAffine(), brush, shape)
}

// DrawRoundRect fills a rounded rectangle with the given color.
func (c *SceneCanvas) DrawRoundRect(r geometry.Rect, color widget.Color, radius float32) {
	r = c.applyTransform(r)
	if !c.isVisible(r) {
		return
	}

	brush := scene.SolidBrush(ToGGColor(color))
	shape := scene.NewRoundedRectShape(r.Min.X, r.Min.Y, r.Width(), r.Height(), radius)
	c.sc.Fill(scene.FillNonZero, scene.IdentityAffine(), brush, shape)
}

// StrokeRoundRect draws the outline of a rounded rectangle.
func (c *SceneCanvas) StrokeRoundRect(r geometry.Rect, color widget.Color, radius float32, strokeWidth float32) {
	r = c.applyTransform(r)
	expanded := r.Expand(strokeWidth / 2)
	if !c.isVisible(expanded) {
		return
	}

	brush := scene.SolidBrush(ToGGColor(color))
	shape := scene.NewRoundedRectShape(r.Min.X, r.Min.Y, r.Width(), r.Height(), radius)
	style := &scene.StrokeStyle{
		Width:      strokeWidth,
		MiterLimit: 10.0,
		Cap:        scene.LineCapButt,
		Join:       scene.LineJoinMiter,
	}
	c.sc.Stroke(style, scene.IdentityAffine(), brush, shape)
}

// DrawCircle fills a circle with the given color.
func (c *SceneCanvas) DrawCircle(center geometry.Point, radius float32, color widget.Color) {
	center = c.applyTransformPoint(center)
	bounds := geometry.NewRect(
		center.X-radius,
		center.Y-radius,
		radius*2,
		radius*2,
	)
	if !c.isVisible(bounds) {
		return
	}

	brush := scene.SolidBrush(ToGGColor(color))
	shape := scene.NewCircleShape(center.X, center.Y, radius)
	c.sc.Fill(scene.FillNonZero, scene.IdentityAffine(), brush, shape)
}

// StrokeCircle draws the outline of a circle.
func (c *SceneCanvas) StrokeCircle(center geometry.Point, radius float32, color widget.Color, strokeWidth float32) {
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

	brush := scene.SolidBrush(ToGGColor(color))
	shape := scene.NewCircleShape(center.X, center.Y, radius)
	style := &scene.StrokeStyle{
		Width:      strokeWidth,
		MiterLimit: 10.0,
		Cap:        scene.LineCapButt,
		Join:       scene.LineJoinMiter,
	}
	c.sc.Stroke(style, scene.IdentityAffine(), brush, shape)
}

// DrawLine draws a line between two points.
func (c *SceneCanvas) DrawLine(from, to geometry.Point, color widget.Color, strokeWidth float32) {
	from = c.applyTransformPoint(from)
	to = c.applyTransformPoint(to)

	bounds := geometry.FromMinMax(from, to).Expand(strokeWidth / 2)
	if !c.isVisible(bounds) {
		return
	}

	brush := scene.SolidBrush(ToGGColor(color))
	shape := scene.NewLineShape(from.X, from.Y, to.X, to.Y)
	style := &scene.StrokeStyle{
		Width:      strokeWidth,
		MiterLimit: 10.0,
		Cap:        scene.LineCapButt,
		Join:       scene.LineJoinMiter,
	}
	c.sc.Stroke(style, scene.IdentityAffine(), brush, shape)
}

// DrawText draws text within the given bounding rectangle.
// Text is rendered via gg.Context (MSDF pipeline) and captured as an image
// to preserve high-quality text rendering.
func (c *SceneCanvas) DrawText(s string, bounds geometry.Rect, fontSize float32, color widget.Color, bold bool, align widget.TextAlign) {
	if s == "" {
		return
	}

	bounds = c.applyTransform(bounds)
	if !c.isVisible(bounds) {
		return
	}

	w := int(math.Ceil(float64(bounds.Width())))
	h := int(math.Ceil(float64(bounds.Height())))
	if w <= 0 || h <= 0 {
		return
	}

	// Create or resize the text gg.Context.
	if c.textDC == nil || c.textDC.Width() != w || c.textDC.Height() != h {
		if c.textDC != nil {
			_ = c.textDC.Close()
		}
		c.textDC = gg.NewContext(w, h)
	}
	c.textDC.ClearWithColor(gg.Transparent)

	// Render text using gg's MSDF pipeline (same logic as Canvas.DrawText).
	ensureDefaultFonts()
	source := defaultRegular
	if bold {
		source = defaultBold
	}
	if source == nil {
		return
	}

	face := source.Face(float64(fontSize))
	c.textDC.SetFont(face)
	c.textDC.SetRGBA(float64(color.R), float64(color.G), float64(color.B), float64(color.A))

	metrics := face.Metrics()
	textHeight := metrics.Ascent + metrics.Descent
	baselineY := math.Round((float64(h)-textHeight)/2 + metrics.Ascent)

	tw, _ := c.textDC.MeasureString(s)
	x := 0.0
	if tw < float64(w) {
		x = (float64(w) - tw) * align.Float64()
	}
	x = math.Round(x)

	c.textDC.DrawString(s, x, baselineY)

	// Capture pixels and add to scene as image.
	img := c.textDC.Image()
	rgba := imageToRGBA(img)
	scImg := scene.NewImage(w, h)
	scImg.Data = rgba.Pix
	c.sc.DrawImage(scImg, scene.TranslateAffine(bounds.Min.X, bounds.Min.Y))
}

// MeasureText returns the width in pixels of the given text string
// when rendered at the specified font size and weight.
// SceneCanvas approximates using average character width since the text
// rendering context may not be initialized.
func (c *SceneCanvas) MeasureText(s string, fontSize float32, bold bool) float32 {
	if s == "" {
		return 0
	}

	// Try to use the gg text context for accurate measurement.
	ensureDefaultFonts()
	source := defaultRegular
	if bold {
		source = defaultBold
	}
	if source != nil {
		// Lazily create a small text context for measurement.
		if c.textDC == nil {
			c.textDC = gg.NewContext(1, 1)
		}
		face := source.Face(float64(fontSize))
		c.textDC.SetFont(face)
		w, _ := c.textDC.MeasureString(s)
		return float32(w)
	}

	// Fallback: approximate with average character width.
	return float32(len([]rune(s))) * fontSize * 0.5
}

// DrawImage draws an image at the specified position.
func (c *SceneCanvas) DrawImage(img image.Image, at geometry.Point) {
	if img == nil {
		return
	}

	at = c.applyTransformPoint(at)

	imgBounds := img.Bounds()
	imgW := float32(imgBounds.Dx())
	imgH := float32(imgBounds.Dy())
	drawRect := geometry.NewRect(at.X, at.Y, imgW, imgH)
	if !c.isVisible(drawRect) {
		return
	}

	w := imgBounds.Dx()
	h := imgBounds.Dy()
	rgba := imageToRGBA(img)
	scImg := scene.NewImage(w, h)
	scImg.Data = rgba.Pix
	c.sc.DrawImage(scImg, scene.TranslateAffine(at.X, at.Y))
}

// PushClip pushes a clipping rectangle onto the clip stack.
func (c *SceneCanvas) PushClip(r geometry.Rect) {
	r = c.applyTransform(r)
	c.clipStack = append(c.clipStack, c.currentClip)
	c.currentClip = c.currentClip.Intersection(r)

	clipShape := scene.NewRectShape(r.Min.X, r.Min.Y, r.Width(), r.Height())
	c.sc.PushClip(clipShape)
}

// PushClipRoundRect pushes a rounded rectangle clipping region.
// SceneCanvas falls back to rectangular clip (scene.Scene does not
// yet support rounded clip shapes).
func (c *SceneCanvas) PushClipRoundRect(r geometry.Rect, radius float32) {
	// TODO: use rounded rect clip shape when scene.Scene supports it.
	// For now, fall back to rectangular clip.
	_ = radius
	c.PushClip(r)
}

// PopClip removes the most recently pushed clipping region.
func (c *SceneCanvas) PopClip() {
	if len(c.clipStack) == 0 {
		return
	}

	lastIdx := len(c.clipStack) - 1
	c.currentClip = c.clipStack[lastIdx]
	c.clipStack = c.clipStack[:lastIdx]

	c.sc.PopClip()
}

// PushTransform pushes a translation transform onto the transform stack.
func (c *SceneCanvas) PushTransform(offset geometry.Point) {
	c.transformStack = append(c.transformStack, c.currentOffset)
	c.currentOffset = c.currentOffset.Add(offset)
}

// PopTransform removes the most recently pushed transform.
func (c *SceneCanvas) PopTransform() {
	if len(c.transformStack) == 0 {
		return
	}

	lastIdx := len(c.transformStack) - 1
	c.currentOffset = c.transformStack[lastIdx]
	c.transformStack = c.transformStack[:lastIdx]
}

// TransformOffset returns the current cumulative transform offset.
func (c *SceneCanvas) TransformOffset() geometry.Point {
	return c.currentOffset
}

// --- Internal helpers ---

// applyTransform applies the current transform offset to a rectangle
// and snaps to pixel grid.
func (c *SceneCanvas) applyTransform(r geometry.Rect) geometry.Rect {
	r = r.Translate(c.currentOffset)
	r.Min.X = snapF(r.Min.X)
	r.Min.Y = snapF(r.Min.Y)
	r.Max.X = snapF(r.Max.X)
	r.Max.Y = snapF(r.Max.Y)
	return r
}

// applyTransformPoint applies the current transform offset to a point
// and snaps to pixel grid.
func (c *SceneCanvas) applyTransformPoint(p geometry.Point) geometry.Point {
	p = p.Add(c.currentOffset)
	p.X = snapF(p.X)
	p.Y = snapF(p.Y)
	return p
}

// isVisible returns true if the rectangle intersects with the current clip bounds.
func (c *SceneCanvas) isVisible(r geometry.Rect) bool {
	return c.currentClip.Intersects(r)
}

// imageToRGBA converts an image.Image to *image.RGBA.
// If the image is already *image.RGBA, it is returned directly.
func imageToRGBA(img image.Image) *image.RGBA {
	if rgba, ok := img.(*image.RGBA); ok {
		return rgba
	}
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
	return rgba
}

// FillSVGPath fills an SVG path within the given bounds using a temporary gg.Context.
func (c *SceneCanvas) FillSVGPath(svgData string, viewBox float32, bounds geometry.Rect, color widget.Color) {
	if svgData == "" || viewBox <= 0 {
		return
	}

	bounds = c.applyTransform(bounds)
	if !c.isVisible(bounds) {
		return
	}

	w := int(math.Ceil(float64(bounds.Width())))
	h := int(math.Ceil(float64(bounds.Height())))
	if w <= 0 || h <= 0 {
		return
	}

	path, err := gg.ParseSVGPath(svgData)
	if err != nil {
		return
	}

	dc := gg.NewContext(w, h)
	scale := float64(bounds.Width()) / float64(viewBox)
	scaleY := float64(bounds.Height()) / float64(viewBox)
	if scaleY < scale {
		scale = scaleY
	}
	dc.Scale(scale, scale)
	dc.SetRGBA(float64(color.R), float64(color.G), float64(color.B), float64(color.A))
	dc.SetFillRule(gg.FillRuleEvenOdd)
	dc.FillPath(path)

	img := dc.Image()
	rgba := imageToRGBA(img)
	scImg := scene.NewImage(w, h)
	scImg.Data = rgba.Pix
	c.sc.DrawImage(scImg, scene.TranslateAffine(bounds.Min.X, bounds.Min.Y))
	_ = dc.Close()
}

// Verify SceneCanvas implements widget.Canvas.
var _ widget.Canvas = (*SceneCanvas)(nil)
