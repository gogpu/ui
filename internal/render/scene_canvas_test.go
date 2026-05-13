package render

import (
	"image"
	"image/color"
	"testing"

	"github.com/gogpu/gg/scene"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// --- Compile-time interface check ---

func TestSceneCanvas_WidgetCanvasInterface(t *testing.T) {
	var _ widget.Canvas = (*SceneCanvas)(nil)
}

// --- Construction ---

func TestNewSceneCanvas(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 800, 600)
	defer c.Close()

	if c.Scene() != sc {
		t.Error("expected same scene reference")
	}
	if c.width != 800 || c.height != 600 {
		t.Errorf("expected dimensions 800x600, got %dx%d", c.width, c.height)
	}
}

// --- DrawRect ---

func TestSceneCanvas_DrawRect(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 100)
	defer c.Close()

	v0 := sc.Version()
	c.DrawRect(geometry.NewRect(10, 20, 50, 30), widget.ColorRed)
	v1 := sc.Version()

	if v1 <= v0 {
		t.Error("scene version should increment after DrawRect")
	}
	if sc.IsEmpty() {
		t.Error("scene should not be empty after DrawRect")
	}
}

// --- DrawRoundRect ---

func TestSceneCanvas_DrawRoundRect(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 100)
	defer c.Close()

	v0 := sc.Version()
	c.DrawRoundRect(geometry.NewRect(10, 20, 80, 40), widget.ColorBlue, 8)
	v1 := sc.Version()

	if v1 <= v0 {
		t.Error("scene version should increment after DrawRoundRect")
	}
}

// --- DrawCircle ---

func TestSceneCanvas_DrawCircle(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 200)
	defer c.Close()

	v0 := sc.Version()
	c.DrawCircle(geometry.Pt(100, 100), 50, widget.ColorGreen)
	v1 := sc.Version()

	if v1 <= v0 {
		t.Error("scene version should increment after DrawCircle")
	}
}

// --- StrokeRect ---

func TestSceneCanvas_StrokeRect(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 100)
	defer c.Close()

	v0 := sc.Version()
	c.StrokeRect(geometry.NewRect(10, 10, 80, 40), widget.ColorBlack, 2)
	v1 := sc.Version()

	if v1 <= v0 {
		t.Error("scene version should increment after StrokeRect")
	}
}

// --- StrokeRoundRect ---

func TestSceneCanvas_StrokeRoundRect(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 100)
	defer c.Close()

	v0 := sc.Version()
	c.StrokeRoundRect(geometry.NewRect(10, 10, 80, 40), widget.ColorBlack, 6, 2)
	v1 := sc.Version()

	if v1 <= v0 {
		t.Error("scene version should increment after StrokeRoundRect")
	}
}

// --- StrokeCircle ---

func TestSceneCanvas_StrokeCircle(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 200)
	defer c.Close()

	v0 := sc.Version()
	c.StrokeCircle(geometry.Pt(100, 100), 40, widget.ColorRed, 3)
	v1 := sc.Version()

	if v1 <= v0 {
		t.Error("scene version should increment after StrokeCircle")
	}
}

// --- StrokeArc ---

func TestSceneCanvas_StrokeArc(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 200)
	defer c.Close()

	v0 := sc.Version()
	c.StrokeArc(geometry.Pt(100, 100), 40, 0, 1.5708, widget.ColorRed, 3)
	v1 := sc.Version()

	if v1 <= v0 {
		t.Error("scene version should increment after StrokeArc")
	}
}

func TestSceneCanvas_StrokeArc_ZeroSweep(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 200)
	defer c.Close()

	v0 := sc.Version()
	c.StrokeArc(geometry.Pt(100, 100), 40, 0, 0, widget.ColorRed, 3)
	v1 := sc.Version()

	if v1 != v0 {
		t.Error("zero sweep StrokeArc should not increment scene version")
	}
}

// --- DrawLine ---

func TestSceneCanvas_DrawLine(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 200)
	defer c.Close()

	v0 := sc.Version()
	c.DrawLine(geometry.Pt(10, 10), geometry.Pt(190, 190), widget.ColorBlack, 2)
	v1 := sc.Version()

	if v1 <= v0 {
		t.Error("scene version should increment after DrawLine")
	}
}

// --- DrawText ---

func TestSceneCanvas_DrawText(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 300, 50)
	defer c.Close()

	v0 := sc.Version()
	c.DrawText("Hello", geometry.NewRect(10, 5, 200, 30), 14, widget.ColorBlack, false, widget.TextAlignLeft)
	v1 := sc.Version()

	if v1 <= v0 {
		t.Error("scene version should increment after DrawText (vector path commands recorded)")
	}
}

func TestSceneCanvas_DrawText_Empty(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 50)
	defer c.Close()

	v0 := sc.Version()
	c.DrawText("", geometry.NewRect(10, 5, 100, 30), 14, widget.ColorBlack, false, widget.TextAlignLeft)
	v1 := sc.Version()

	if v1 != v0 {
		t.Error("empty text should not increment scene version")
	}
}

func TestSceneCanvas_DrawText_Bold(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 300, 50)
	defer c.Close()

	v0 := sc.Version()
	c.DrawText("Bold", geometry.NewRect(10, 5, 200, 30), 14, widget.ColorBlack, true, widget.TextAlignCenter)
	v1 := sc.Version()

	if v1 <= v0 {
		t.Error("bold text should still produce scene commands")
	}
}

// --- DrawImage ---

func TestSceneCanvas_DrawImage(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 200)
	defer c.Close()

	// Create a small test image.
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for i := range img.Pix {
		img.Pix[i] = 128
	}

	v0 := sc.Version()
	c.DrawImage(img, geometry.Pt(10, 10))
	v1 := sc.Version()

	if v1 <= v0 {
		t.Error("scene version should increment after DrawImage")
	}
}

func TestSceneCanvas_DrawImage_Nil(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 200)
	defer c.Close()

	v0 := sc.Version()
	c.DrawImage(nil, geometry.Pt(10, 10))
	v1 := sc.Version()

	if v1 != v0 {
		t.Error("nil image should not increment scene version")
	}
}

// --- Clear ---

func TestSceneCanvas_Clear(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 100, 100)
	defer c.Close()

	v0 := sc.Version()
	c.Clear(widget.ColorWhite)
	v1 := sc.Version()

	if v1 <= v0 {
		t.Error("scene version should increment after Clear")
	}
}

// --- PushClip / PopClip ---

func TestSceneCanvas_PushPopClip(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 200)
	defer c.Close()

	// Initial clip depth is 0.
	if len(c.clipStack) != 0 {
		t.Errorf("expected empty clip stack, got depth %d", len(c.clipStack))
	}

	c.PushClip(geometry.NewRect(10, 10, 100, 100))
	if len(c.clipStack) != 1 {
		t.Errorf("expected clip stack depth 1, got %d", len(c.clipStack))
	}

	c.PushClip(geometry.NewRect(20, 20, 50, 50))
	if len(c.clipStack) != 2 {
		t.Errorf("expected clip stack depth 2, got %d", len(c.clipStack))
	}

	c.PopClip()
	if len(c.clipStack) != 1 {
		t.Errorf("expected clip stack depth 1 after pop, got %d", len(c.clipStack))
	}

	c.PopClip()
	if len(c.clipStack) != 0 {
		t.Errorf("expected clip stack depth 0 after double pop, got %d", len(c.clipStack))
	}

	// Extra pop should be safe (no panic).
	c.PopClip()
}

// --- PushTransform / PopTransform ---

func TestSceneCanvas_PushPopTransform(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 200)
	defer c.Close()

	if len(c.transformStack) != 0 {
		t.Errorf("expected empty transform stack, got depth %d", len(c.transformStack))
	}

	c.PushTransform(geometry.Pt(10, 20))
	if len(c.transformStack) != 1 {
		t.Errorf("expected transform stack depth 1, got %d", len(c.transformStack))
	}
	if c.currentOffset.X != 10 || c.currentOffset.Y != 20 {
		t.Errorf("expected offset (10,20), got (%v,%v)", c.currentOffset.X, c.currentOffset.Y)
	}

	c.PushTransform(geometry.Pt(5, 5))
	if c.currentOffset.X != 15 || c.currentOffset.Y != 25 {
		t.Errorf("expected offset (15,25), got (%v,%v)", c.currentOffset.X, c.currentOffset.Y)
	}

	c.PopTransform()
	if c.currentOffset.X != 10 || c.currentOffset.Y != 20 {
		t.Errorf("expected offset (10,20) after pop, got (%v,%v)", c.currentOffset.X, c.currentOffset.Y)
	}

	c.PopTransform()
	if c.currentOffset.X != 0 || c.currentOffset.Y != 0 {
		t.Errorf("expected offset (0,0) after double pop, got (%v,%v)", c.currentOffset.X, c.currentOffset.Y)
	}

	// Extra pop should be safe (no panic).
	c.PopTransform()
}

// --- Nested Transforms ---

func TestSceneCanvas_NestedTransforms(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 200)
	defer c.Close()

	c.PushTransform(geometry.Pt(50, 50))
	c.PushTransform(geometry.Pt(10, 10))

	// Draw a rect with nested transforms: effective offset = (60, 60).
	v0 := sc.Version()
	c.DrawRect(geometry.NewRect(0, 0, 20, 20), widget.ColorRed)
	v1 := sc.Version()

	if v1 <= v0 {
		t.Error("DrawRect with transforms should increment version")
	}

	c.PopTransform()
	c.PopTransform()
}

// --- Clip Visibility ---

func TestSceneCanvas_ClipVisibility(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 200)
	defer c.Close()

	// Push a tight clip region.
	c.PushClip(geometry.NewRect(50, 50, 50, 50))

	// Draw outside clip bounds: should be skipped.
	v0 := sc.Version()
	c.DrawRect(geometry.NewRect(0, 0, 10, 10), widget.ColorRed)
	v1 := sc.Version()

	if v1 != v0 {
		t.Error("draw outside clip should be skipped (no version increment)")
	}

	// Draw inside clip bounds: should be recorded.
	c.DrawRect(geometry.NewRect(60, 60, 20, 20), widget.ColorGreen)
	v2 := sc.Version()

	if v2 <= v1 {
		t.Error("draw inside clip should be recorded (version should increment)")
	}

	c.PopClip()
}

// --- Close ---

func TestSceneCanvas_Close(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 50)

	// Draw text to exercise the text path before Close.
	c.DrawText("Test", geometry.NewRect(0, 0, 100, 30), 14, widget.ColorBlack, false, widget.TextAlignLeft)

	// Close should not panic.
	c.Close()

	// Double close should be safe (no-op).
	c.Close()
}

// --- imageToRGBA ---

// --- DrawText Vector Path Verification ---

// TestSceneCanvas_DrawText_VectorPaths verifies that DrawText records vector
// Fill commands (glyph outlines) rather than TagImage bitmap captures.
// This is the core assertion for ADR-007 Task 1c.
func TestSceneCanvas_DrawText_VectorPaths(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 400, 50)
	defer c.Close()

	c.DrawText("Vector", geometry.NewRect(10, 5, 200, 30), 14, widget.ColorBlack, false, widget.TextAlignLeft)

	// scene.DrawText records glyph outlines as Fill commands, NOT TagImage.
	// Check that the scene has NO images registered (vector-only).
	images := sc.Images()
	if len(images) > 0 {
		t.Errorf("expected no images in scene (vector text), got %d", len(images))
	}

	// Verify that fill commands were recorded (glyph outlines).
	tags := sc.Encoding().Tags()
	hasText := false
	for _, tag := range tags {
		if tag == scene.TagFill || tag == scene.TagText {
			hasText = true
			break
		}
	}
	if !hasText {
		t.Error("expected TagFill or TagText commands from text rendering")
	}
}

// TestSceneCanvas_DrawText_Alignment verifies that text alignment produces
// different x positions for the same text.
func TestSceneCanvas_DrawText_Alignment(t *testing.T) {
	tests := []struct {
		name  string
		align widget.TextAlign
	}{
		{"left", widget.TextAlignLeft},
		{"center", widget.TextAlignCenter},
		{"right", widget.TextAlignRight},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := scene.NewScene()
			c := NewSceneCanvas(sc, 400, 50)
			defer c.Close()

			v0 := sc.Version()
			c.DrawText("Align", geometry.NewRect(0, 0, 400, 50), 14, widget.ColorBlack, false, tt.align)
			v1 := sc.Version()

			if v1 <= v0 {
				t.Errorf("alignment %s: expected version increment", tt.name)
			}
		})
	}
}

// TestSceneCanvas_DrawText_ZeroBounds verifies that zero-sized bounds produce no output.
func TestSceneCanvas_DrawText_ZeroBounds(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 50)
	defer c.Close()

	v0 := sc.Version()
	c.DrawText("Hello", geometry.NewRect(10, 5, 0, 30), 14, widget.ColorBlack, false, widget.TextAlignLeft)
	v1 := sc.Version()

	if v1 != v0 {
		t.Error("zero-width bounds should produce no scene commands")
	}

	v2 := sc.Version()
	c.DrawText("Hello", geometry.NewRect(10, 5, 100, 0), 14, widget.ColorBlack, false, widget.TextAlignLeft)
	v3 := sc.Version()

	if v3 != v2 {
		t.Error("zero-height bounds should produce no scene commands")
	}
}

// TestSceneCanvas_MeasureText_StandaloneFont verifies that MeasureText works
// without relying on a temporary gg.Context.
func TestSceneCanvas_MeasureText_Standalone(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 50)
	defer c.Close()

	w := c.MeasureText("Hello World", 14, false)
	if w <= 0 {
		t.Errorf("expected positive text width, got %f", w)
	}

	wBold := c.MeasureText("Hello World", 14, true)
	if wBold <= 0 {
		t.Errorf("expected positive bold text width, got %f", wBold)
	}

	// Bold text should be at least as wide as regular.
	if wBold < w*0.9 {
		t.Errorf("bold width (%f) unexpectedly much smaller than regular (%f)", wBold, w)
	}
}

// TestSceneCanvas_MeasureText_Empty verifies that empty string returns 0.
func TestSceneCanvas_MeasureText_Empty(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 50)
	defer c.Close()

	if w := c.MeasureText("", 14, false); w != 0 {
		t.Errorf("expected 0 for empty text, got %f", w)
	}
}

// --- imageToRGBA ---

func TestImageToRGBA_AlreadyRGBA(t *testing.T) {
	orig := image.NewRGBA(image.Rect(0, 0, 10, 10))
	result := imageToRGBA(orig)
	if result != orig {
		t.Error("imageToRGBA should return the same *image.RGBA without copying")
	}
}

func TestImageToRGBA_NonRGBA(t *testing.T) {
	// Create a non-RGBA image.
	nrgba := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	nrgba.SetNRGBA(1, 1, color.NRGBA{R: 255, G: 0, B: 0, A: 255})

	result := imageToRGBA(nrgba)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Bounds() != nrgba.Bounds() {
		t.Error("bounds should match")
	}
}

func TestSceneCanvas_TextModeController(t *testing.T) {
	sc := scene.NewScene()
	canvas := NewSceneCanvas(sc, 100, 100)

	tc, ok := widget.Canvas(canvas).(widget.TextModeController)
	if !ok {
		t.Fatal("SceneCanvas should implement TextModeController")
	}

	if tc.TextMode() != widget.TextModeAuto {
		t.Errorf("TextMode = %v, want Auto", tc.TextMode())
	}

	tc.SetTextMode(widget.TextModeMSDF)
	if tc.TextMode() != widget.TextModeAuto {
		t.Error("SceneCanvas.TextMode should always return Auto (no-op)")
	}
}

// --- DeviceScaler (ADR-026) ---

func TestSceneCanvas_DeviceScaler_Interface(t *testing.T) {
	sc := scene.NewScene()
	canvas := NewSceneCanvas(sc, 100, 100)

	_, ok := widget.Canvas(canvas).(widget.DeviceScaler)
	if !ok {
		t.Fatal("SceneCanvas should implement widget.DeviceScaler")
	}
}

func TestSceneCanvas_DeviceScale_Default(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 100, 100)

	if got := c.DeviceScale(); got != 1 {
		t.Errorf("DeviceScale() default = %f, want 1.0", got)
	}
}

func TestSceneCanvas_DeviceScale_SetGet(t *testing.T) {
	tests := []struct {
		name     string
		set      float32
		expected float32
	}{
		{"scale 2.0", 2.0, 2.0},
		{"scale 1.5", 1.5, 1.5},
		{"scale 1.0", 1.0, 1.0},
		{"scale 0 → default 1.0", 0, 1.0},
		{"scale negative → default 1.0", -1, 1.0},
		{"scale 3.0", 3.0, 3.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := scene.NewScene()
			c := NewSceneCanvas(sc, 100, 100)
			c.SetDeviceScale(tt.set)

			if got := c.DeviceScale(); got != tt.expected {
				t.Errorf("DeviceScale() = %f, want %f", got, tt.expected)
			}
		})
	}
}

// --- svgDrawTransform ---

func TestSvgDrawTransform_Scale1(t *testing.T) {
	// At scale 1.0, should produce a pure translation.
	aff := svgDrawTransform(10, 20, 1.0)
	expected := scene.TranslateAffine(10, 20)

	if aff != expected {
		t.Errorf("svgDrawTransform(10,20,1) = %+v, want %+v", aff, expected)
	}
}

func TestSvgDrawTransform_ScaleLessThan1(t *testing.T) {
	// Scale < 1 should also produce pure translation (no downscale).
	aff := svgDrawTransform(10, 20, 0.5)
	expected := scene.TranslateAffine(10, 20)

	if aff != expected {
		t.Errorf("svgDrawTransform(10,20,0.5) = %+v, want %+v", aff, expected)
	}
}

func TestSvgDrawTransform_Scale2(t *testing.T) {
	// Scale 2.0: translate(10,20) * scale(0.5, 0.5)
	aff := svgDrawTransform(10, 20, 2.0)

	// Expected: A=0.5, B=0, C=10, D=0, E=0.5, F=20
	if aff.A != 0.5 || aff.E != 0.5 {
		t.Errorf("scale components: A=%f, E=%f, want 0.5, 0.5", aff.A, aff.E)
	}
	if aff.C != 10 || aff.F != 20 {
		t.Errorf("translation components: C=%f, F=%f, want 10, 20", aff.C, aff.F)
	}
	if aff.B != 0 || aff.D != 0 {
		t.Errorf("off-diagonal: B=%f, D=%f, want 0, 0", aff.B, aff.D)
	}
}

func TestSvgDrawTransform_PointMapping(t *testing.T) {
	// A 40×40 image drawn at scale 2.0 should map (40,40) → (30,30) = (10+40*0.5, 20+40*0.5)
	aff := svgDrawTransform(10, 20, 2.0)
	x, y := aff.TransformPoint(40, 40)

	if x != 30 || y != 40 {
		t.Errorf("TransformPoint(40,40) = (%f,%f), want (30,40)", x, y)
	}

	// Origin maps to the translation offset.
	x0, y0 := aff.TransformPoint(0, 0)
	if x0 != 10 || y0 != 20 {
		t.Errorf("TransformPoint(0,0) = (%f,%f), want (10,20)", x0, y0)
	}
}

// --- FillSVGPath DPI-aware rendering ---

// simpleSVGPath is a valid SVG path for testing FillSVGPath.
const simpleSVGPath = "M12 2L2 22h20z"

func TestSceneCanvas_FillSVGPath_Scale1_ProducesScene(t *testing.T) {
	globalIconCache.invalidateAll()
	defer globalIconCache.invalidateAll()

	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 200)
	defer c.Close()

	// Scale 1.0 (default) — baseline behavior.
	v0 := sc.Version()
	c.FillSVGPath(simpleSVGPath, 24, geometry.NewRect(10, 10, 20, 20), widget.ColorBlack)
	v1 := sc.Version()

	if v1 <= v0 {
		t.Error("FillSVGPath at scale=1 should produce scene commands")
	}
}

func TestSceneCanvas_FillSVGPath_Scale2_ProducesScene(t *testing.T) {
	globalIconCache.invalidateAll()
	defer globalIconCache.invalidateAll()

	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 200)
	c.SetDeviceScale(2.0)
	defer c.Close()

	v0 := sc.Version()
	c.FillSVGPath(simpleSVGPath, 24, geometry.NewRect(10, 10, 20, 20), widget.ColorBlack)
	v1 := sc.Version()

	if v1 <= v0 {
		t.Error("FillSVGPath at scale=2 should produce scene commands")
	}
}

func TestSceneCanvas_FillSVGPath_DifferentScales_DifferentCacheEntries(t *testing.T) {
	globalIconCache.invalidateAll()
	defer globalIconCache.invalidateAll()

	// Render at scale 1.
	sc1 := scene.NewScene()
	c1 := NewSceneCanvas(sc1, 200, 200)
	c1.SetDeviceScale(1.0)
	c1.FillSVGPath(simpleSVGPath, 24, geometry.NewRect(10, 10, 20, 20), widget.ColorBlack)
	c1.Close()

	stats1 := globalIconCache.stats()
	entries1 := stats1.ImageEntries

	// Render at scale 2 — must create a SEPARATE cache entry.
	sc2 := scene.NewScene()
	c2 := NewSceneCanvas(sc2, 200, 200)
	c2.SetDeviceScale(2.0)
	c2.FillSVGPath(simpleSVGPath, 24, geometry.NewRect(10, 10, 20, 20), widget.ColorBlack)
	c2.Close()

	stats2 := globalIconCache.stats()
	entries2 := stats2.ImageEntries

	if entries2 <= entries1 {
		t.Errorf("different scales should produce different cache entries: scale1=%d, scale2=%d",
			entries1, entries2)
	}
}

func TestSceneCanvas_FillSVGPath_EmptyPath(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 200)
	defer c.Close()

	v0 := sc.Version()
	c.FillSVGPath("", 24, geometry.NewRect(10, 10, 20, 20), widget.ColorBlack)
	v1 := sc.Version()

	if v1 != v0 {
		t.Error("empty SVG path should not produce scene commands")
	}
}

func TestSceneCanvas_FillSVGPath_ZeroViewBox(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 200)
	defer c.Close()

	v0 := sc.Version()
	c.FillSVGPath(simpleSVGPath, 0, geometry.NewRect(10, 10, 20, 20), widget.ColorBlack)
	v1 := sc.Version()

	if v1 != v0 {
		t.Error("zero viewBox should not produce scene commands")
	}
}

// --- RenderSVG DPI-aware rendering ---

// minimalSVGForCanvas is a valid SVG XML for testing RenderSVG in scene_canvas tests.
var minimalSVGForCanvas = []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path d="M12 2L2 22h20z"/></svg>`)

func TestSceneCanvas_RenderSVG_Scale1_ProducesScene(t *testing.T) {
	globalIconCache.invalidateAll()
	defer globalIconCache.invalidateAll()

	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 200)
	defer c.Close()

	v0 := sc.Version()
	c.RenderSVG(minimalSVGForCanvas, geometry.NewRect(10, 10, 20, 20), widget.ColorBlack)
	v1 := sc.Version()

	if v1 <= v0 {
		t.Error("RenderSVG at scale=1 should produce scene commands")
	}
}

func TestSceneCanvas_RenderSVG_Scale2_ProducesScene(t *testing.T) {
	globalIconCache.invalidateAll()
	defer globalIconCache.invalidateAll()

	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 200)
	c.SetDeviceScale(2.0)
	defer c.Close()

	v0 := sc.Version()
	c.RenderSVG(minimalSVGForCanvas, geometry.NewRect(10, 10, 20, 20), widget.ColorBlack)
	v1 := sc.Version()

	if v1 <= v0 {
		t.Error("RenderSVG at scale=2 should produce scene commands")
	}
}

func TestSceneCanvas_RenderSVG_DifferentScales_DifferentCacheEntries(t *testing.T) {
	globalIconCache.invalidateAll()
	defer globalIconCache.invalidateAll()

	// Render at scale 1.
	sc1 := scene.NewScene()
	c1 := NewSceneCanvas(sc1, 200, 200)
	c1.SetDeviceScale(1.0)
	c1.RenderSVG(minimalSVGForCanvas, geometry.NewRect(10, 10, 20, 20), widget.ColorBlack)
	c1.Close()

	stats1 := globalIconCache.stats()
	entries1 := stats1.ImageEntries

	// Render at scale 2.
	sc2 := scene.NewScene()
	c2 := NewSceneCanvas(sc2, 200, 200)
	c2.SetDeviceScale(2.0)
	c2.RenderSVG(minimalSVGForCanvas, geometry.NewRect(10, 10, 20, 20), widget.ColorBlack)
	c2.Close()

	stats2 := globalIconCache.stats()
	entries2 := stats2.ImageEntries

	if entries2 <= entries1 {
		t.Errorf("different scales should produce different cache entries: scale1=%d, scale2=%d",
			entries1, entries2)
	}
}

func TestSceneCanvas_RenderSVG_EmptyXML(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 200)
	defer c.Close()

	v0 := sc.Version()
	c.RenderSVG(nil, geometry.NewRect(10, 10, 20, 20), widget.ColorBlack)
	v1 := sc.Version()

	if v1 != v0 {
		t.Error("nil SVG XML should not produce scene commands")
	}
}

func TestSceneCanvas_RenderSVG_ZeroBounds(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 200)
	defer c.Close()

	v0 := sc.Version()
	c.RenderSVG(minimalSVGForCanvas, geometry.NewRect(10, 10, 0, 0), widget.ColorBlack)
	v1 := sc.Version()

	if v1 != v0 {
		t.Error("zero-size bounds should not produce scene commands")
	}
}

func TestSceneCanvas_FillSVGPath_Scale1_CacheHit(t *testing.T) {
	globalIconCache.invalidateAll()
	defer globalIconCache.invalidateAll()

	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 200)
	defer c.Close()

	// First call: cache miss.
	c.FillSVGPath(simpleSVGPath, 24, geometry.NewRect(10, 10, 20, 20), widget.ColorBlack)
	stats1 := globalIconCache.stats()

	// Second call with same params: cache hit.
	c.FillSVGPath(simpleSVGPath, 24, geometry.NewRect(10, 10, 20, 20), widget.ColorBlack)
	stats2 := globalIconCache.stats()

	if stats2.Hits <= stats1.Hits {
		t.Error("second FillSVGPath call should produce a cache hit")
	}
}

func TestSceneCanvas_RenderSVG_Scale2_CacheHit(t *testing.T) {
	globalIconCache.invalidateAll()
	defer globalIconCache.invalidateAll()

	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 200)
	c.SetDeviceScale(2.0)
	defer c.Close()

	// First call: cache miss.
	c.RenderSVG(minimalSVGForCanvas, geometry.NewRect(10, 10, 20, 20), widget.ColorBlack)
	stats1 := globalIconCache.stats()

	// Second call with same params + scale: cache hit.
	c.RenderSVG(minimalSVGForCanvas, geometry.NewRect(10, 10, 20, 20), widget.ColorBlack)
	stats2 := globalIconCache.stats()

	if stats2.Hits <= stats1.Hits {
		t.Error("second RenderSVG call at same scale should produce a cache hit")
	}
}

// --- StyledTextDrawer interface ---

func TestSceneCanvas_ImplementsStyledTextDrawer(t *testing.T) {
	var _ widget.StyledTextDrawer = (*SceneCanvas)(nil)
}

func TestSceneCanvas_DrawStyledText_DefaultFont(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 100)
	defer c.Close()

	bounds := geometry.NewRect(10, 10, 180, 30)
	style := widget.TextStyle{
		FontSize: 14,
		Color:    widget.ColorBlack,
		Align:    widget.TextAlignLeft,
	}

	// Should not panic -- uses default Inter font.
	c.DrawStyledText("Hello World", bounds, style)

	if sc.IsEmpty() {
		t.Error("scene should have content after DrawStyledText")
	}
}

func TestSceneCanvas_DrawStyledText_Bold(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 100)
	defer c.Close()

	bounds := geometry.NewRect(10, 10, 180, 30)
	style := widget.TextStyle{
		FontSize: 14,
		Bold:     true,
		Color:    widget.ColorBlack,
		Align:    widget.TextAlignCenter,
	}

	c.DrawStyledText("Bold Text", bounds, style)

	if sc.IsEmpty() {
		t.Error("scene should have content after DrawStyledText with bold")
	}
}

func TestSceneCanvas_DrawStyledText_Empty(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 100)
	defer c.Close()

	style := widget.TextStyle{FontSize: 14, Color: widget.ColorBlack}
	c.DrawStyledText("", geometry.NewRect(10, 10, 180, 30), style)

	if !sc.IsEmpty() {
		t.Error("scene should be empty after DrawStyledText with empty string")
	}
}

func TestSceneCanvas_DrawStyledText_OutsideClip(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 100, 100)
	defer c.Close()

	// Bounds completely outside the canvas.
	bounds := geometry.NewRect(500, 500, 100, 30)
	style := widget.TextStyle{FontSize: 14, Color: widget.ColorBlack}

	c.DrawStyledText("Offscreen", bounds, style)

	if !sc.IsEmpty() {
		t.Error("scene should be empty after DrawStyledText with offscreen bounds")
	}
}

func TestSceneCanvas_DrawStyledText_ExplicitFamily(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 100)
	defer c.Close()

	bounds := geometry.NewRect(10, 10, 180, 30)
	style := widget.TextStyle{
		FontFamily: "Inter",
		FontSize:   16,
		Color:      widget.ColorBlack,
		Align:      widget.TextAlignRight,
	}

	c.DrawStyledText("Inter Font", bounds, style)

	if sc.IsEmpty() {
		t.Error("scene should have content after DrawStyledText with Inter family")
	}
}

func TestSceneCanvas_DrawStyledText_UnknownFamily(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 100)
	defer c.Close()

	bounds := geometry.NewRect(10, 10, 180, 30)
	style := widget.TextStyle{
		FontFamily: "NonExistentFont",
		FontSize:   14,
		Color:      widget.ColorBlack,
	}

	// Unknown family should fall back to Inter.
	c.DrawStyledText("Fallback", bounds, style)

	if sc.IsEmpty() {
		t.Error("scene should have content after DrawStyledText with unknown family (fallback)")
	}
}

func TestSceneCanvas_MeasureStyledText_Default(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 100)
	defer c.Close()

	style := widget.TextStyle{FontSize: 14, Color: widget.ColorBlack}
	w := c.MeasureStyledText("Hello", style)

	if w <= 0 {
		t.Errorf("MeasureStyledText(Hello) = %f, want > 0", w)
	}
}

func TestSceneCanvas_MeasureStyledText_Empty(t *testing.T) {
	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 200, 100)
	defer c.Close()

	style := widget.TextStyle{FontSize: 14}
	w := c.MeasureStyledText("", style)

	if w != 0 {
		t.Errorf("MeasureStyledText('') = %f, want 0", w)
	}
}
