package render

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

func TestNewRenderer(t *testing.T) {
	r := NewRenderer(800, 600)
	defer func() { _ = r.Close() }()

	if r.Width() != 800 {
		t.Errorf("Width() = %v, want 800", r.Width())
	}
	if r.Height() != 600 {
		t.Errorf("Height() = %v, want 600", r.Height())
	}
	if r.Context() == nil {
		t.Error("Context() should not be nil")
	}
	if r.Canvas() == nil {
		t.Error("Canvas() should not be nil")
	}
	if r.InFrame() {
		t.Error("InFrame() should be false initially")
	}
}

func TestRenderer_Resize(t *testing.T) {
	r := NewRenderer(800, 600)
	defer func() { _ = r.Close() }()

	// Resize to different dimensions
	r.Resize(1024, 768)

	if r.Width() != 1024 {
		t.Errorf("After resize, Width() = %v, want 1024", r.Width())
	}
	if r.Height() != 768 {
		t.Errorf("After resize, Height() = %v, want 768", r.Height())
	}
	if r.Canvas().Width() != 1024 {
		t.Errorf("After resize, Canvas().Width() = %v, want 1024", r.Canvas().Width())
	}
	if r.Canvas().Height() != 768 {
		t.Errorf("After resize, Canvas().Height() = %v, want 768", r.Canvas().Height())
	}
}

func TestRenderer_ResizeSameSize(t *testing.T) {
	r := NewRenderer(800, 600)
	defer func() { _ = r.Close() }()

	oldDC := r.Context()

	// Resize to same dimensions should be no-op
	r.Resize(800, 600)

	if r.Context() != oldDC {
		t.Error("Resize to same size should not recreate context")
	}
}

func TestRenderer_BeginEndFrame(t *testing.T) {
	r := NewRenderer(100, 100)
	defer func() { _ = r.Close() }()

	if r.InFrame() {
		t.Error("InFrame() should be false before BeginFrame")
	}

	// Begin frame
	canvas := r.BeginFrame(widget.ColorWhite)

	if !r.InFrame() {
		t.Error("InFrame() should be true after BeginFrame")
	}
	if canvas == nil {
		t.Error("BeginFrame should return non-nil canvas")
	}
	if canvas != r.Canvas() {
		t.Error("BeginFrame should return the renderer's canvas")
	}

	// End frame
	dc := r.EndFrame()

	if r.InFrame() {
		t.Error("InFrame() should be false after EndFrame")
	}
	if dc == nil {
		t.Error("EndFrame should return non-nil dc")
	}
	if dc != r.Context() {
		t.Error("EndFrame should return the renderer's dc")
	}
}

func TestRenderer_BeginFrameWhileInFrame(t *testing.T) {
	r := NewRenderer(100, 100)
	defer func() { _ = r.Close() }()

	// Begin first frame
	canvas1 := r.BeginFrame(widget.ColorWhite)

	// Add some state to canvas
	canvas1.PushClip(geometry.NewRect(10, 10, 80, 80))
	canvas1.PushTransform(geometry.Pt(5, 5))

	// Begin second frame while still in first (should reset)
	canvas2 := r.BeginFrame(widget.ColorBlack)

	if canvas2 != canvas1 {
		t.Error("BeginFrame should return same canvas")
	}
	if canvas2.ClipDepth() != 0 {
		t.Error("Canvas should be reset when BeginFrame called while in frame")
	}
	if canvas2.TransformDepth() != 0 {
		t.Error("Canvas should be reset when BeginFrame called while in frame")
	}
}

func TestRenderer_DrawingWorkflow(t *testing.T) {
	r := NewRenderer(100, 100)
	defer func() { _ = r.Close() }()

	// Complete drawing workflow
	canvas := r.BeginFrame(widget.ColorWhite)

	// Draw some content
	canvas.DrawRect(geometry.NewRect(10, 10, 30, 30), widget.ColorRed)
	canvas.DrawCircle(geometry.Pt(50, 50), 20, widget.ColorBlue)
	canvas.DrawLine(geometry.Pt(0, 0), geometry.Pt(100, 100), widget.ColorBlack, 1.0)

	dc := r.EndFrame()

	// Verify dc has valid image
	img := dc.Image()
	if img == nil {
		t.Error("Context.Image() should not be nil")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("Image bounds = %v, want 100x100", bounds)
	}
}

func TestRenderer_Close(t *testing.T) {
	r := NewRenderer(100, 100)

	err := r.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	// Close should be idempotent
	err = r.Close()
	if err != nil {
		t.Errorf("Second Close() returned error: %v", err)
	}

	if r.InFrame() {
		t.Error("InFrame() should be false after Close")
	}
}

func TestRenderer_CloseEndsFrame(t *testing.T) {
	r := NewRenderer(100, 100)

	// Start a frame
	_ = r.BeginFrame(widget.ColorWhite)
	if !r.InFrame() {
		t.Error("Should be in frame after BeginFrame")
	}

	// Close should end the frame
	_ = r.Close()
	if r.InFrame() {
		t.Error("InFrame() should be false after Close")
	}
}

func TestRenderer_ResizeEndsFrame(t *testing.T) {
	r := NewRenderer(100, 100)
	defer func() { _ = r.Close() }()

	// Start a frame
	_ = r.BeginFrame(widget.ColorWhite)
	if !r.InFrame() {
		t.Error("Should be in frame after BeginFrame")
	}

	// Resize should end the frame
	r.Resize(200, 200)
	if r.InFrame() {
		t.Error("InFrame() should be false after Resize")
	}
}

func TestDefaultRenderConfig(t *testing.T) {
	config := DefaultRenderConfig()

	if config.BackgroundColor != widget.ColorWhite {
		t.Errorf("BackgroundColor = %v, want white", config.BackgroundColor)
	}
	if config.Scale != 1.0 {
		t.Errorf("Scale = %v, want 1.0", config.Scale)
	}
}

func TestNewSoftwareTarget(t *testing.T) {
	target := NewSoftwareTarget(800, 600)
	defer func() { _ = target.Close() }()

	if target.Width() != 800 {
		t.Errorf("Width() = %v, want 800", target.Width())
	}
	if target.Height() != 600 {
		t.Errorf("Height() = %v, want 600", target.Height())
	}
	if target.Context() == nil {
		t.Error("Context() should not be nil")
	}
}

func TestSoftwareTarget_Bounds(t *testing.T) {
	target := NewSoftwareTarget(100, 200)
	defer func() { _ = target.Close() }()

	bounds := target.Bounds()

	if bounds.Min.X != 0 || bounds.Min.Y != 0 {
		t.Errorf("Bounds().Min = (%v, %v), want (0, 0)", bounds.Min.X, bounds.Min.Y)
	}
	if bounds.Width() != 100 || bounds.Height() != 200 {
		t.Errorf("Bounds() size = %vx%v, want 100x200", bounds.Width(), bounds.Height())
	}
}

func TestSoftwareTarget_ImplementsRenderTarget(t *testing.T) {
	target := NewSoftwareTarget(100, 100)
	defer func() { _ = target.Close() }()

	// Verify the interface is implemented
	var _ RenderTarget = target
}

func TestSoftwareTarget_Close(t *testing.T) {
	target := NewSoftwareTarget(100, 100)

	err := target.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}
}

// Benchmarks

func BenchmarkRenderer_BeginEndFrame(b *testing.B) {
	r := NewRenderer(800, 600)
	defer func() { _ = r.Close() }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.BeginFrame(widget.ColorWhite)
		_ = r.EndFrame()
	}
}

func BenchmarkRenderer_Resize(b *testing.B) {
	r := NewRenderer(800, 600)
	defer func() { _ = r.Close() }()

	sizes := [][2]int{{800, 600}, {1024, 768}, {1920, 1080}, {800, 600}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		size := sizes[i%len(sizes)]
		r.Resize(size[0], size[1])
	}
}

func BenchmarkNewRenderer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r := NewRenderer(800, 600)
		_ = r.Close()
	}
}

func BenchmarkNewSoftwareTarget(b *testing.B) {
	for i := 0; i < b.N; i++ {
		t := NewSoftwareTarget(800, 600)
		_ = t.Close()
	}
}
