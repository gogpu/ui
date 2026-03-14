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
		t.Error("scene version should increment after DrawText (image command recorded)")
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

	// Force text DC creation.
	c.DrawText("Test", geometry.NewRect(0, 0, 100, 30), 14, widget.ColorBlack, false, widget.TextAlignLeft)

	if c.textDC == nil {
		t.Error("textDC should be created after DrawText")
	}

	c.Close()

	if c.textDC != nil {
		t.Error("textDC should be nil after Close")
	}

	// Double close should be safe.
	c.Close()
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
