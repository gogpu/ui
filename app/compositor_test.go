package app

import (
	"testing"

	"github.com/gogpu/gg/scene"
	"github.com/gogpu/ui/compositor"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	internalRender "github.com/gogpu/ui/internal/render"
	"github.com/gogpu/ui/widget"
)

// animWidget simulates a spinner: calls SetNeedsRedraw during Draw.
type animWidget struct {
	widget.WidgetBase
	drawCount int
}

func (w *animWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(48, 48))
}

func (w *animWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	w.drawCount++
	// Draw a rect so the scene is non-empty.
	canvas.DrawRect(w.Bounds(), widget.RGBA8(255, 0, 0, 255))
	w.SetNeedsRedraw(true)
	if ctx != nil {
		ctx.InvalidateRect(w.Bounds())
	}
}

func (w *animWidget) Event(_ widget.Context, _ event.Event) bool { return false }
func (w *animWidget) Children() []widget.Widget                  { return nil }

// staticWidget is a non-animated widget.
type staticWidget struct {
	widget.WidgetBase
	drawCount int
}

func (w *staticWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(800, 40))
}

func (w *staticWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	w.drawCount++
	canvas.DrawRect(w.Bounds(), widget.RGBA8(128, 128, 128, 255))
}

func (w *staticWidget) Event(_ widget.Context, _ event.Event) bool { return false }
func (w *staticWidget) Children() []widget.Widget                  { return nil }

// TestBuildLayerTree_RootBoundary verifies layer tree construction
// from a widget tree with root boundary.
func TestBuildLayerTree_RootBoundary(t *testing.T) {
	root := &staticWidget{}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))

	layer := BuildLayerTree(root)
	if layer == nil {
		t.Fatal("BuildLayerTree should return non-nil layer")
	}
}

// TestBuildLayerTree_NestedBoundaries verifies that nested boundary
// widgets produce nested layers in the tree.
func TestBuildLayerTree_NestedBoundaries(t *testing.T) {
	root := &containerTestWidget{children: make([]widget.Widget, 0)}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))

	child := &staticWidget{}
	child.SetVisible(true)
	child.SetRepaintBoundary(true)
	child.SetBounds(geometry.NewRect(100, 200, 148, 248))
	child.SetParent(root)
	root.children = append(root.children, child)

	layer := BuildLayerTree(root)
	if layer == nil {
		t.Fatal("BuildLayerTree returned nil")
	}

	// Root should have at least one child layer (the child boundary).
	children := layer.Children()
	if len(children) == 0 {
		t.Fatal("root layer should have child layers for nested boundaries")
	}
}

// TestCompositorIntegration_SpinnerAnimation is the END-TO-END test
// that validates the full pipeline: spinner re-records → compositor
// produces fresh composed scene → animation not frozen.
//
// This is the exact scenario that was broken before Layer Tree.
func TestCompositorIntegration_SpinnerAnimation(t *testing.T) {
	prev := widget.GetSceneRecorderFactory()
	widget.RegisterSceneRecorder(testSceneRecorder)
	defer widget.RegisterSceneRecorder(prev)
	root := &containerTestWidget{children: make([]widget.Widget, 0)}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))
	root.SetScreenOrigin(geometry.Pt(0, 0))

	spinner := &animWidget{}
	spinner.SetVisible(true)
	spinner.SetRepaintBoundary(true)
	spinner.SetBounds(geometry.NewRect(100, 200, 148, 248))
	spinner.SetScreenOrigin(geometry.Pt(100, 200))
	spinner.SetParent(root)
	root.children = append(root.children, spinner)

	comp := compositor.New()

	ctx := widget.NewContext()
	ctx.SetOnInvalidateRect(func(_ geometry.Rect) {})

	// Frame 1: paint boundaries then build layer tree.
	PaintBoundaryLayersWithContext(root, nil, ctx)

	// Verify recording happened.
	if spinner.drawCount == 0 {
		t.Fatal("frame 1: spinner.Draw should have been called by PaintBoundaryLayers")
	}
	if root.drawCount == 0 {
		t.Fatal("frame 1: root.Draw should have been called by PaintBoundaryLayers")
	}
	if root.CachedScene() == nil {
		t.Fatal("frame 1: root.CachedScene() is nil after PaintBoundaryLayers")
	}
	if root.CachedScene().IsEmpty() {
		t.Fatal("frame 1: root.CachedScene() is empty after PaintBoundaryLayers")
	}

	layerTree := BuildLayerTree(root)
	scene1 := comp.Compose(layerTree)

	if scene1.IsEmpty() {
		t.Fatal("frame 1: composed scene should not be empty")
	}
	v1 := scene1.Version()

	// Frame 2: spinner re-dirtied itself. Only spinner needs re-paint.
	spinnerDrew := spinner.drawCount
	layerTree = BuildLayerTree(root) // rebuild to pick up fresh scenes
	PaintBoundaryLayersWithContext(root, layerTree, ctx)
	scene2 := comp.Compose(layerTree)

	if spinner.drawCount <= spinnerDrew {
		t.Error("frame 2: spinner.Draw should have been called again (animation)")
	}
	v2 := scene2.Version()

	if v2 <= v1 {
		t.Errorf("frame 2: composed version %d <= frame 1 version %d; "+
			"animation frozen — composed scene is stale", v2, v1)
	}

	// With depth > 0 (all child boundaries render inline in root scene),
	// root IS re-recorded when spinner is dirty. This is expected —
	// inline rendering requires parent scene to include updated child content.
	if root.drawCount < 1 {
		t.Errorf("root.drawCount = %d; root should have drawn at least once", root.drawCount)
	}
}

// containerTestWidget is a widget with explicit children list.
type containerTestWidget struct {
	widget.WidgetBase
	drawCount int
	children  []widget.Widget
}

func (w *containerTestWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(800, 600))
}

func (w *containerTestWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	w.drawCount++
	canvas.DrawRect(w.Bounds(), widget.RGBA8(200, 200, 200, 255))
}

func (w *containerTestWidget) Event(_ widget.Context, _ event.Event) bool { return false }
func (w *containerTestWidget) Children() []widget.Widget                  { return w.children }

// testSceneRecorder creates a SceneCanvas for recording into SceneCache.
func testSceneRecorder(s widget.SceneCache, w, h int) (widget.Canvas, func()) {
	rec := internalRender.NewSceneCanvas(s.(*scene.Scene), w, h)
	return rec, rec.Close
}
