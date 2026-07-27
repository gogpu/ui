package compositor

import (
	"testing"
	"unsafe"

	"github.com/gogpu/gpucontext"
	"github.com/gogpu/ui/geometry"
)

func TestNewExternalTextureLayer(t *testing.T) {
	// Create a fake texture view (non-nil pointer for testing).
	var dummy int
	tv := gpucontext.NewTextureView(unsafe.Pointer(&dummy))

	l := NewExternalTextureLayer(tv, 320, 240, 10.5, 20.5)

	if l.Texture().IsNil() {
		t.Fatal("texture should not be nil")
	}
	if l.Width() != 320 {
		t.Errorf("width = %d, want 320", l.Width())
	}
	if l.Height() != 240 {
		t.Errorf("height = %d, want 240", l.Height())
	}
	if l.X() != 10.5 {
		t.Errorf("X() = %f, want 10.5", l.X())
	}
	if l.Y() != 20.5 {
		t.Errorf("Y() = %f, want 20.5", l.Y())
	}
	if !l.NeedsCompositing() {
		t.Error("new layer should need compositing")
	}
	if l.Parent() != nil {
		t.Error("new layer should have nil parent")
	}
}

func TestExternalTextureLayer_SetTexture(t *testing.T) {
	l := NewExternalTextureLayer(gpucontext.TextureView{}, 100, 100, 0, 0)

	if !l.Texture().IsNil() {
		t.Fatal("initial texture should be nil (zero value)")
	}

	var dummy int
	tv := gpucontext.NewTextureView(unsafe.Pointer(&dummy))
	l.SetTexture(tv)

	if l.Texture().IsNil() {
		t.Fatal("texture should not be nil after SetTexture")
	}
	if !l.NeedsCompositing() {
		t.Error("layer should need compositing after SetTexture")
	}
}

func TestExternalTextureLayer_SetSize(t *testing.T) {
	l := NewExternalTextureLayer(gpucontext.TextureView{}, 100, 100, 0, 0)

	l.ClearNeedsCompositing()
	l.SetSize(200, 150)

	if l.Width() != 200 {
		t.Errorf("width = %d, want 200", l.Width())
	}
	if l.Height() != 150 {
		t.Errorf("height = %d, want 150", l.Height())
	}
	if !l.NeedsCompositing() {
		t.Error("layer should need compositing after SetSize")
	}
}

func TestExternalTextureLayer_LayerInterface(t *testing.T) {
	l := NewExternalTextureLayer(gpucontext.TextureView{}, 100, 100, 5, 10)

	// Verify Layer interface compliance.
	var _ Layer = l

	if l.Offset() != (geometry.Point{X: 5, Y: 10}) {
		t.Errorf("offset = %v, want (5, 10)", l.Offset())
	}

	l.SetOffset(geometry.Pt(15, 25))
	if l.Offset() != (geometry.Point{X: 15, Y: 25}) {
		t.Errorf("offset after SetOffset = %v, want (15, 25)", l.Offset())
	}
}

func TestExternalTextureLayer_ParentAttachment(t *testing.T) {
	parent := NewOffsetLayer(geometry.Point{})
	l := NewExternalTextureLayer(gpucontext.TextureView{}, 100, 100, 0, 0)

	parent.Append(l)

	if l.Parent() != parent {
		t.Error("parent should be set after Append")
	}
	if len(parent.Children()) != 1 {
		t.Fatalf("parent children = %d, want 1", len(parent.Children()))
	}

	parent.Remove(l)

	if l.Parent() != nil {
		t.Error("parent should be nil after Remove")
	}
	if len(parent.Children()) != 0 {
		t.Fatalf("parent children = %d, want 0", len(parent.Children()))
	}
}
