package render

import (
	"image"
	stdcolor "image/color"
	"image/draw"
	"math"

	"github.com/gogpu/gg"
	"github.com/gogpu/gg/scene"
	"github.com/gogpu/gg/text"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/theme/font"
	"github.com/gogpu/ui/widget"
)

// SceneCanvas implements [widget.Canvas] by recording drawing commands into
// a [scene.Scene]. This is used inside RepaintBoundary for tile-parallel
// rendering of widget subtrees.
//
// Shape drawing (rect, round rect, circle, line) is recorded directly into
// the scene's encoding as scene shapes. Text rendering uses
// scene.DrawText which converts glyph outlines to vector Fill paths —
// scalable and resolution-independent. No bitmap capture needed.
//
// SceneCanvas is NOT thread-safe. All drawing operations must occur on the
// main/UI thread during the Draw phase.
type SceneCanvas struct {
	sc     *scene.Scene
	width  int
	height int

	// Clip stack: stores previous clip bounds for each PushClip.
	clipStack []geometry.Rect
	// Current clip bounds (intersection of all pushed clips).
	currentClip geometry.Rect

	// Transform stack: stores cumulative offsets for each PushTransform.
	transformStack []geometry.Point
	// Current cumulative transform offset.
	currentOffset geometry.Point

	// screenOriginBase is the screen-space position of the RepaintBoundary
	// that owns this SceneCanvas. Set before recording child drawing so
	// StampScreenOrigin produces correct screen-space ScreenOrigin values.
	screenOriginBase geometry.Point

	// deviceScale is the display scale factor (DPI scaling). SVG icons are
	// rasterized at ceil(logicalSize * deviceScale) physical pixels, then
	// drawn with an inverse-scale affine transform so they appear at the
	// correct logical size but with crisp, HiDPI-quality rendering.
	// Follows the Qt6/Chromium/IntelliJ pattern (ADR-026).
	// A value <= 0 is treated as 1.0.
	deviceScale float32
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
// Currently a no-op since vector text rendering does not hold persistent
// resources, but retained for lifecycle symmetry with NewSceneCanvas.
func (c *SceneCanvas) Close() {
	// No resources to release — vector text uses no persistent state.
}

// Scene returns the underlying scene.Scene.
func (c *SceneCanvas) Scene() *scene.Scene {
	return c.sc
}

// IsBoundaryRecording returns true. SceneCanvas records into a boundary's
// scene.Scene. DrawChild uses this to skip child boundaries — they have
// their own PictureLayers in the compositor (Flutter paintChild pattern).
func (c *SceneCanvas) IsBoundaryRecording() bool {
	return true
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

// FillRectDirect fills a rectangle via the scene graph. SceneCanvas does not
// use the GPU SDF accelerator, so this is equivalent to DrawRect.
func (c *SceneCanvas) FillRectDirect(r geometry.Rect, color widget.Color) {
	c.DrawRect(r, color)
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

// StrokeArc draws a circular arc outline from startAngle with the given sweep.
// Uses scene.Path.Arc to record the arc as cubic Bézier curves into the scene.
func (c *SceneCanvas) StrokeArc(center geometry.Point, radius float32,
	startAngle, sweepAngle float64, color widget.Color, strokeWidth float32) {
	if sweepAngle == 0 {
		return
	}

	center = c.applyTransformPoint(center)

	// Visibility check (conservative: full circle bounding box).
	bounds := geometry.NewRect(
		center.X-radius-strokeWidth/2,
		center.Y-radius-strokeWidth/2,
		(radius+strokeWidth/2)*2,
		(radius+strokeWidth/2)*2,
	)
	if !c.isVisible(bounds) {
		return
	}

	// Build a scene path for the arc.
	endAngle := startAngle + sweepAngle
	sweepClockwise := sweepAngle > 0
	path := scene.NewPath().Arc(
		center.X, center.Y,
		radius, radius,
		float32(startAngle), float32(endAngle),
		sweepClockwise,
	)

	brush := scene.SolidBrush(ToGGColor(color))
	shape := scene.NewPathShape(path)
	style := &scene.StrokeStyle{
		Width:      strokeWidth,
		MiterLimit: 10.0,
		Cap:        scene.LineCapButt,
		Join:       scene.LineJoinMiter,
	}
	c.sc.Stroke(style, scene.IdentityAffine(), brush, shape)
}

// StrokeArcStyled draws a circular arc with the specified line cap style.
func (c *SceneCanvas) StrokeArcStyled(center geometry.Point, radius float32,
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

	endAngle := startAngle + sweepAngle
	sweepClockwise := sweepAngle > 0
	path := scene.NewPath().Arc(
		center.X, center.Y,
		radius, radius,
		float32(startAngle), float32(endAngle),
		sweepClockwise,
	)

	brush := scene.SolidBrush(ToGGColor(color))
	shape := scene.NewPathShape(path)
	style := &scene.StrokeStyle{
		Width:      strokeWidth,
		MiterLimit: 10.0,
		Cap:        toSceneLineCap(lineCap),
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

// DrawText draws text within the given bounding rectangle using vector paths.
//
// Text is rendered via scene.DrawText which converts glyph outlines to vector
// Fill paths in the scene — scalable, resolution-independent, and enterprise-
// quality. No bitmap capture or temporary gg.Context needed.
//
// Alignment is handled by measuring the text advance and computing the x offset
// before calling scene.DrawText with the final position.
func (c *SceneCanvas) DrawText(s string, bounds geometry.Rect, fontSize float32, color widget.Color, bold bool, align widget.TextAlign) {
	if s == "" {
		return
	}

	bounds = c.applyTransform(bounds)
	if !c.isVisible(bounds) {
		return
	}

	if bounds.Width() <= 0 || bounds.Height() <= 0 {
		return
	}

	// Resolve font face.
	ensureDefaultFonts()
	source := defaultRegular
	if bold {
		source = defaultBold
	}
	if source == nil {
		return
	}

	face := source.Face(float64(fontSize))

	// Calculate baseline Y by centering text vertically within bounds.
	// Same logic as Canvas.DrawText (canvas.go) for visual consistency.
	metrics := face.Metrics()
	textHeight := metrics.Ascent + metrics.Descent
	baselineY := math.Round(float64(bounds.Min.Y) + (float64(bounds.Height())-textHeight)/2 + metrics.Ascent)

	// Calculate x position based on alignment using text.Measure for
	// standalone measurement (no gg.Context needed).
	tw, _ := text.Measure(s, face)
	available := float64(bounds.Width())
	x := float64(bounds.Min.X)
	if tw < available {
		x += (available - tw) * align.Float64()
	}
	x = math.Round(x)

	// Record text as vector glyph outlines into the scene.
	brush := scene.SolidBrush(ToGGColor(color))
	_ = c.sc.DrawText(s, face, float32(x), float32(baselineY), brush)
}

// MeasureText returns the width in pixels of the given text string
// when rendered at the specified font size and weight.
//
// Uses text.Measure from the gg text package for accurate, standalone
// measurement — no gg.Context required.
func (c *SceneCanvas) MeasureText(s string, fontSize float32, bold bool) float32 {
	if s == "" {
		return 0
	}

	ensureDefaultFonts()
	source := defaultRegular
	if bold {
		source = defaultBold
	}
	if source != nil {
		face := source.Face(float64(fontSize))
		w, _ := text.Measure(s, face)
		return float32(w)
	}

	// Fallback: approximate with average character width.
	return float32(len([]rune(s))) * fontSize * 0.5
}

// DrawStyledText draws text within the given bounding rectangle using
// a font resolved from the global [FontRegistry]. This enables rendering
// with custom fonts (CJK, icon fonts, etc.) loaded via the plugin system.
//
// If the FontFamily is empty or not found, the default Inter font is used.
func (c *SceneCanvas) DrawStyledText(s string, bounds geometry.Rect, style widget.TextStyle) {
	if s == "" {
		return
	}

	bounds = c.applyTransform(bounds)
	if !c.isVisible(bounds) {
		return
	}

	if bounds.Width() <= 0 || bounds.Height() <= 0 {
		return
	}

	source := resolveStyledFontSource(style)
	if source == nil {
		return
	}

	face := source.Face(float64(style.FontSize))

	// Calculate baseline Y by centering text vertically within bounds.
	metrics := face.Metrics()
	textHeight := metrics.Ascent + metrics.Descent
	baselineY := math.Round(float64(bounds.Min.Y) + (float64(bounds.Height())-textHeight)/2 + metrics.Ascent)

	// Calculate x position based on alignment.
	tw, _ := text.Measure(s, face)
	available := float64(bounds.Width())
	x := float64(bounds.Min.X)
	if tw < available {
		x += (available - tw) * style.Align.Float64()
	}
	x = math.Round(x)

	// Record text as vector glyph outlines into the scene.
	brush := scene.SolidBrush(ToGGColor(style.Color))
	_ = c.sc.DrawText(s, face, float32(x), float32(baselineY), brush)
}

// MeasureStyledText returns the width in pixels of the given text string
// when rendered with the specified [widget.TextStyle]. The font is resolved
// from the global [FontRegistry].
func (c *SceneCanvas) MeasureStyledText(s string, style widget.TextStyle) float32 {
	if s == "" {
		return 0
	}

	source := resolveStyledFontSource(style)
	if source != nil {
		face := source.Face(float64(style.FontSize))
		w, _ := text.Measure(s, face)
		return float32(w)
	}

	// Fallback: approximate with average character width.
	return float32(len([]rune(s))) * style.FontSize * 0.5
}

// resolveStyledFontSource resolves a [text.FontSource] from a [widget.TextStyle]
// using the global [FontRegistry]. Shared by Canvas and SceneCanvas.
func resolveStyledFontSource(style widget.TextStyle) *text.FontSource {
	reg := GlobalFontRegistry()

	family := style.FontFamily
	if family == "" {
		family = defaultFontFamily
	}

	weight := font.Regular
	if style.Bold {
		weight = font.Bold
	}

	fontStyle := font.Normal
	if style.Italic {
		fontStyle = font.Italic
	}

	return reg.Resolve(family, weight, fontStyle)
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

// ScreenOriginBase returns the screen-space base offset for this SceneCanvas.
// For RepaintBoundary recording, this is the boundary widget's screen position
// so that StampScreenOrigin computes correct screen-space coordinates even
// after PushTransform(-bounds.Min) shifts to local coordinates.
func (c *SceneCanvas) ScreenOriginBase() geometry.Point { return c.screenOriginBase }

// SetScreenOriginBase sets the screen-space base offset for this SceneCanvas.
func (c *SceneCanvas) SetScreenOriginBase(p geometry.Point) { c.screenOriginBase = p }

// DeviceScale returns the display scale factor used for SVG rasterization.
// Returns 1.0 if no scale has been set.
func (c *SceneCanvas) DeviceScale() float32 {
	if c.deviceScale <= 0 {
		return 1
	}
	return c.deviceScale
}

// SetDeviceScale sets the display scale factor for HiDPI-aware SVG icon
// rasterization (ADR-026). Icons are rasterized at physical pixel size
// (ceil(logical * scale)) and drawn with an inverse-scale transform.
func (c *SceneCanvas) SetDeviceScale(scale float32) { c.deviceScale = scale }

// ClipBounds returns the current clip rectangle.
func (c *SceneCanvas) ClipBounds() geometry.Rect {
	return c.currentClip
}

// ReplayScene merges a child scene.Scene into this canvas's parent scene
// with translation offset. This is the scene-concatenation path (ADR-007)
// used when a RepaintBoundary replays its cached display list inside
// another SceneCanvas (nested boundaries).
//
// The child scene was recorded in local coordinates (0,0 = boundary origin).
// AppendWithTranslation offsets all path coordinates by the current cumulative
// transform offset, following the Vello pattern (encoding.rs:162-169).
func (c *SceneCanvas) ReplayScene(s *scene.Scene) {
	if s == nil || s.IsEmpty() {
		return
	}
	c.sc.AppendWithTranslation(s, c.currentOffset.X, c.currentOffset.Y)
}

// --- Internal helpers ---

// svgDrawTransform builds the affine transform for drawing a rasterized SVG
// icon that was rendered at physical pixel size. When scale == 1, this is a
// pure translation. When scale > 1, the image is drawn at 1/scale size so
// the oversized raster maps back to the correct logical pixel area.
//
// The transform is: translate(tx, ty) * scale(1/s, 1/s).
// Matrix form:
//
//	| 1/s   0   tx |
//	|  0   1/s  ty |
func svgDrawTransform(tx, ty, scale float32) scene.Affine {
	if scale <= 1 {
		return scene.TranslateAffine(tx, ty)
	}
	inv := 1.0 / scale
	return scene.NewAffine(inv, 0, tx, 0, inv, ty)
}

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
//
// When deviceScale > 1, the icon is rasterized at physical pixel size
// (ceil(logical * scale)) and drawn with an inverse-scale affine transform
// so it appears at the correct logical size with crisp HiDPI rendering.
// This follows the Qt6/Chromium/IntelliJ pattern (ADR-026).
//
// Results are cached in the global icon cache: the rasterized scene.Image is
// keyed by (svgData pointer, width, height, color). Cache hits skip parsing
// and rasterization entirely — only a map lookup + scene.DrawImage.
func (c *SceneCanvas) FillSVGPath(svgData string, viewBox float32, bounds geometry.Rect, color widget.Color) {
	if svgData == "" || viewBox <= 0 {
		return
	}

	bounds = c.applyTransform(bounds)
	if !c.isVisible(bounds) {
		return
	}

	dpiScale := c.DeviceScale()

	// Physical pixel dimensions for rasterization.
	physW := int(math.Ceil(float64(bounds.Width()) * float64(dpiScale)))
	physH := int(math.Ceil(float64(bounds.Height()) * float64(dpiScale)))
	if physW <= 0 || physH <= 0 {
		return
	}

	// Icon cache lookup (Level 2: rasterized image).
	// Key uses physical dimensions — different scales produce different entries.
	key := iconImageKey{
		svgPtr: svgStringPtr(svgData),
		width:  physW,
		height: physH,
		color:  packColor(color),
	}
	if cached := globalIconCache.getImage(key); cached != nil {
		c.sc.DrawImage(cached, svgDrawTransform(bounds.Min.X, bounds.Min.Y, dpiScale))
		return
	}

	// Cache miss: parse + rasterize at physical resolution.
	path, err := gg.ParseSVGPath(svgData)
	if err != nil {
		return
	}

	dc := gg.NewContext(physW, physH)
	dc.SetRasterizerMode(gg.RasterizerAnalytic) // CPU-only: bypass GPU queueing
	// Scale SVG viewBox to physical pixel dimensions.
	svgScale := float64(physW) / float64(viewBox)
	svgScaleY := float64(physH) / float64(viewBox)
	if svgScaleY < svgScale {
		svgScale = svgScaleY
	}
	dc.Scale(svgScale, svgScale)
	dc.SetRGBA(float64(color.R), float64(color.G), float64(color.B), float64(color.A))
	dc.SetFillRule(gg.FillRuleEvenOdd)
	dc.FillPath(path)

	img := dc.Image()
	rgba := imageToRGBA(img)
	scImg := scene.NewImage(physW, physH)
	scImg.Data = rgba.Pix
	c.sc.DrawImage(scImg, svgDrawTransform(bounds.Min.X, bounds.Min.Y, dpiScale))
	_ = dc.Close()

	// Store in cache for next frame.
	globalIconCache.putImage(key, scImg)
}

// toSceneLineCap converts widget.LineCap to scene.LineCap.
func toSceneLineCap(lc widget.LineCap) scene.LineCap {
	switch lc {
	case widget.LineCapRound:
		return scene.LineCapRound
	case widget.LineCapSquare:
		return scene.LineCapSquare
	default:
		return scene.LineCapButt
	}
}

// SetTextMode is a no-op on SceneCanvas. Scene text uses TagText which
// handles mode selection at replay time via GPUSceneRenderer.
func (c *SceneCanvas) SetTextMode(_ widget.TextMode) {}

// TextMode always returns TextModeAuto on SceneCanvas.
func (c *SceneCanvas) TextMode() widget.TextMode { return widget.TextModeAuto }

// RenderSVG rasterizes full SVG XML to bitmap and encodes as scene image.
//
// When deviceScale > 1, the SVG is rasterized at physical pixel size
// (ceil(logical * scale)) and drawn with an inverse-scale affine transform
// so it appears at the correct logical size with crisp HiDPI rendering.
// This follows the Qt6/Chromium/IntelliJ pattern (ADR-026).
//
// Uses the global icon cache for both parsing (Level 1: svg.Document by
// data pointer) and rasterization (Level 2: scene.Image by pointer+size+color).
// On cache hit, the entire method reduces to a map lookup + scene.DrawImage.
func (c *SceneCanvas) RenderSVG(svgXML []byte, bounds geometry.Rect, color widget.Color) {
	if len(svgXML) == 0 {
		return
	}
	bounds = c.applyTransform(bounds)

	dpiScale := c.DeviceScale()

	// Physical pixel dimensions for rasterization.
	physW := int(math.Ceil(float64(bounds.Width()) * float64(dpiScale)))
	physH := int(math.Ceil(float64(bounds.Height()) * float64(dpiScale)))
	if physW <= 0 || physH <= 0 {
		return
	}

	// Icon cache lookup (Level 2: rasterized image).
	// Key uses physical dimensions — different scales produce different entries.
	key := iconImageKey{
		svgPtr: svgSlicePtr(svgXML),
		width:  physW,
		height: physH,
		color:  packColor(color),
	}
	if cached := globalIconCache.getImage(key); cached != nil {
		c.sc.DrawImage(cached, svgDrawTransform(bounds.Min.X, bounds.Min.Y, dpiScale))
		return
	}

	// Cache miss: parse SVG (Level 1 cache) + rasterize at physical resolution.
	doc := globalIconCache.getDoc(svgXML)
	if doc == nil {
		return
	}

	dc := gg.NewContext(physW, physH)
	dc.SetRasterizerMode(gg.RasterizerAnalytic) // CPU-only: bypass GPU queueing
	r8, g8, b8, a8 := color.RGBA8()
	doc.RenderToWithColor(dc, 0, 0, float64(physW), float64(physH),
		stdcolor.NRGBA{R: r8, G: g8, B: b8, A: a8})

	rgba := imageToRGBA(dc.Image())
	scImg := scene.NewImage(physW, physH)
	scImg.Data = rgba.Pix
	c.sc.DrawImage(scImg, svgDrawTransform(bounds.Min.X, bounds.Min.Y, dpiScale))
	_ = dc.Close()

	// Store in cache for next frame.
	globalIconCache.putImage(key, scImg)
}

// Verify SceneCanvas implements widget.Canvas.
var _ widget.Canvas = (*SceneCanvas)(nil)

// Verify SceneCanvas implements widget.ArcStroker.
var _ widget.ArcStroker = (*SceneCanvas)(nil)

// Verify SceneCanvas implements widget.SVGFiller.
var _ widget.SVGFiller = (*SceneCanvas)(nil)

// Verify SceneCanvas implements widget.SVGRenderer.
var _ widget.SVGRenderer = (*SceneCanvas)(nil)

// Verify SceneCanvas implements widget.StyledTextDrawer.
var _ widget.StyledTextDrawer = (*SceneCanvas)(nil)
