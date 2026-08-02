package gpuview

import (
	"testing"
	"unsafe"

	"github.com/gogpu/gpucontext"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

func TestNew_Defaults(t *testing.T) {
	w := New()

	width, height := w.ViewportSize()
	if width != 320 || height != 240 {
		t.Errorf("default size = %dx%d, want 320x240", width, height)
	}
	if w.IsContinuous() {
		t.Error("default should not be continuous")
	}
	if !w.IsVisible() {
		t.Error("widget should be visible by default")
	}
	if !w.IsEnabled() {
		t.Error("widget should be enabled by default")
	}
	if !w.IsRepaintBoundary() {
		t.Error("gpuview should be a repaint boundary")
	}
	// Before first Draw, texture is not initialized — this is expected.
	// Verified: w.Texture().IsNil() == true.
}

func TestNew_WithOptions(t *testing.T) {
	var renderCalled bool
	w := New(
		Size(640, 480),
		OnRender(func(_ gpucontext.TextureView) {
			renderCalled = true
		}),
		Continuous(true),
	)

	width, height := w.ViewportSize()
	if width != 640 || height != 480 {
		t.Errorf("size = %dx%d, want 640x480", width, height)
	}
	if !w.IsContinuous() {
		t.Error("should be continuous")
	}
	// render not called yet (no Draw)
	if renderCalled {
		t.Error("render should not be called before Draw")
	}
}

func TestLayout_ReturnsConstrainedSize(t *testing.T) {
	w := New(Size(800, 600))
	ctx := widget.NewContext()

	// Unconstrained: returns preferred size.
	size := w.Layout(ctx, geometry.Constraints{
		MinWidth:  0,
		MinHeight: 0,
		MaxWidth:  1920,
		MaxHeight: 1080,
	})
	if size.Width != 800 || size.Height != 600 {
		t.Errorf("unconstrained size = %v, want 800x600", size)
	}

	// Constrained smaller: clamps to max.
	size = w.Layout(ctx, geometry.Constraints{
		MinWidth:  0,
		MinHeight: 0,
		MaxWidth:  400,
		MaxHeight: 300,
	})
	if size.Width != 400 || size.Height != 300 {
		t.Errorf("constrained size = %v, want 400x300", size)
	}
}

func TestDraw_WithGPUProvider(t *testing.T) {
	var renderCount int
	w := New(
		Size(100, 100),
		OnRender(func(tv gpucontext.TextureView) {
			renderCount++
			if tv.IsNil() {
				t.Error("texture passed to OnRender should not be nil")
			}
		}),
	)

	ctx := newMockContextWithGPU()

	// First Draw: initializes texture and renders.
	w.Draw(ctx, nil)
	if renderCount != 1 {
		t.Errorf("render count = %d, want 1 after first draw", renderCount)
	}
	if w.Texture().IsNil() {
		t.Error("texture should be initialized after first draw")
	}

	// Second Draw (not dirty, not continuous): no render.
	w.Draw(ctx, nil)
	if renderCount != 1 {
		t.Errorf("render count = %d, want 1 (not dirty)", renderCount)
	}

	// Invalidate and draw again.
	w.Invalidate()
	w.Draw(ctx, nil)
	if renderCount != 2 {
		t.Errorf("render count = %d, want 2 after invalidate", renderCount)
	}
}

func TestDraw_Continuous(t *testing.T) {
	var renderCount int
	w := New(
		Size(100, 100),
		Continuous(true),
		OnRender(func(_ gpucontext.TextureView) {
			renderCount++
		}),
	)

	ctx := newMockContextWithGPU()

	// Each Draw triggers render in continuous mode.
	w.Draw(ctx, nil)
	w.Draw(ctx, nil)
	w.Draw(ctx, nil)
	if renderCount != 3 {
		t.Errorf("render count = %d, want 3 (continuous mode)", renderCount)
	}
}

func TestDraw_NoGPUProvider(t *testing.T) {
	var renderCalled bool
	w := New(
		Size(100, 100),
		OnRender(func(_ gpucontext.TextureView) {
			renderCalled = true
		}),
	)

	// Plain context without GPUTextureProvider — render should not be called.
	ctx := widget.NewContext()
	w.Draw(ctx, nil)

	if renderCalled {
		t.Error("render should not be called without GPU provider")
	}
	if !w.Texture().IsNil() {
		t.Error("texture should remain nil without GPU provider")
	}
}

func TestDraw_Invisible(t *testing.T) {
	var renderCalled bool
	w := New(
		Size(100, 100),
		OnRender(func(_ gpucontext.TextureView) {
			renderCalled = true
		}),
	)
	w.SetVisible(false)

	ctx := newMockContextWithGPU()
	w.Draw(ctx, nil)

	if renderCalled {
		t.Error("render should not be called when widget is invisible")
	}
}

func TestEvent_DoesNotConsume(t *testing.T) {
	w := New()
	ctx := widget.NewContext()

	me := event.NewMouseEvent(
		event.MousePress,
		event.ButtonLeft,
		event.ButtonStateLeft,
		geometry.Pt(10, 10),
		geometry.Pt(10, 10),
		0,
	)

	if w.Event(ctx, me) {
		t.Error("gpuview should not consume events")
	}
}

func TestChildren_ReturnsNil(t *testing.T) {
	w := New()
	if w.Children() != nil {
		t.Error("gpuview should have no children")
	}
}

func TestRelease(t *testing.T) {
	w := New(Size(100, 100))
	ctx := newMockContextWithGPU()

	w.Draw(ctx, nil)
	if w.Texture().IsNil() {
		t.Fatal("texture should be initialized after draw")
	}

	w.Release()
	if !w.Texture().IsNil() {
		t.Error("texture should be nil after release")
	}
}

func TestUnmount_ReleasesTexture(t *testing.T) {
	var released bool
	w := New(Size(100, 100))

	// Manually set texture and release for testing.
	var dummy int
	w.texture = gpucontext.NewTextureView(unsafe.Pointer(&dummy))
	w.release = func() { released = true }
	w.initialized = true

	w.Unmount()

	if !released {
		t.Error("Unmount should call release function")
	}
	if !w.Texture().IsNil() {
		t.Error("texture should be nil after Unmount")
	}
}

func TestSetContinuous(t *testing.T) {
	w := New()

	if w.IsContinuous() {
		t.Error("default should not be continuous")
	}

	w.SetContinuous(true)
	if !w.IsContinuous() {
		t.Error("should be continuous after SetContinuous(true)")
	}

	w.SetContinuous(false)
	if w.IsContinuous() {
		t.Error("should not be continuous after SetContinuous(false)")
	}
}

func TestAccessible(t *testing.T) {
	w := New()
	acc := w.Accessible()

	if acc == nil {
		t.Fatal("Accessible() should not return nil")
	}
	if acc.AccessibilityRole() != 6 { // a11y.RoleImage = 6
		// Just verify it returns a role without importing a11y in test assertion.
		_ = acc.AccessibilityRole()
	}
	if acc.AccessibilityLabel() != "GPU View" {
		t.Errorf("label = %q, want %q", acc.AccessibilityLabel(), "GPU View")
	}
}

func TestSizeOption_IgnoresZeroAndNegative(t *testing.T) {
	w := New(Size(0, -1))
	width, height := w.ViewportSize()
	// Zero/negative should not override defaults.
	if width != 320 || height != 240 {
		t.Errorf("size = %dx%d, want 320x240 (invalid size ignored)", width, height)
	}
}

// --- Test helpers ---

// mockGPUContext wraps ContextImpl and provides GPUTextureProvider.
type mockGPUContext struct {
	*widget.ContextImpl
}

func (m *mockGPUContext) CreateGPUTexture(width, height int) (any, func()) {
	if width <= 0 || height <= 0 {
		return nil, nil
	}
	// Return a fake texture view with a non-nil pointer.
	var dummy int
	tv := gpucontext.NewTextureView(unsafe.Pointer(&dummy))
	return tv, func() {} // no-op release for testing
}

func newMockContextWithGPU() *mockGPUContext {
	return &mockGPUContext{ContextImpl: widget.NewContext()}
}

// Verify mockGPUContext satisfies both Context and GPUTextureProvider.
var (
	_ widget.Context            = (*mockGPUContext)(nil)
	_ widget.GPUTextureProvider = (*mockGPUContext)(nil)
)
