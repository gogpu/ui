package app

import (
	"testing"

	"github.com/gogpu/gg/scene"
	"github.com/gogpu/ui/compositor"
	"github.com/gogpu/ui/core/progress"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	internalRender "github.com/gogpu/ui/internal/render"
	"github.com/gogpu/ui/widget"
)

// boxContainer is a test container that draws children via DrawChild.
type boxContainer struct {
	widget.WidgetBase
	kids []widget.Widget
}

func (w *boxContainer) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(800, 600))
}

func (w *boxContainer) Draw(ctx widget.Context, canvas widget.Canvas) {
	canvas.DrawRect(w.Bounds(), widget.RGBA8(240, 240, 240, 255))
	for _, child := range w.kids {
		widget.DrawChild(child, ctx, canvas)
	}
}

func (w *boxContainer) Event(_ widget.Context, _ event.Event) bool { return false }
func (w *boxContainer) Children() []widget.Widget                  { return w.kids }

// TestSpinnerE2E_VisibleInCompositor is the definitive end-to-end test.
// Uses REAL progress.Widget (not mock) with REAL SceneCanvas recording.
// Verifies spinner is found by PaintBoundaryLayers AND visible in composed scene.
func TestSpinnerE2E_VisibleInCompositor(t *testing.T) {
	prev := widget.GetSceneRecorderFactory()
	widget.RegisterSceneRecorder(func(s widget.SceneCache, w, h int) (widget.Canvas, func()) {
		rec := internalRender.NewSceneCanvas(s.(*scene.Scene), w, h)
		return rec, rec.Close
	})
	defer widget.RegisterSceneRecorder(prev)

	// Build: root(boundary) → container → spinner(boundary)
	root := &boxContainer{}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))
	root.SetScreenOrigin(geometry.Pt(0, 0))

	container := &boxContainer{}
	container.SetVisible(true)
	container.SetBounds(geometry.NewRect(0, 200, 800, 400))
	root.kids = []widget.Widget{container}

	spinner := progress.New(progress.Indeterminate(true), progress.Size(48))
	spinner.SetBounds(geometry.NewRect(100, 10, 148, 58))
	spinner.SetScreenOrigin(geometry.Pt(100, 10))
	spinner.SetParent(container)
	container.kids = []widget.Widget{spinner}

	ctx := widget.NewContext()
	ctx.SetOnInvalidateRect(func(_ geometry.Rect) {})

	// Step 1: PaintBoundaryLayers
	PaintBoundaryLayersWithContext(root, nil, ctx)

	// Verify root scene recorded.
	if root.CachedScene() == nil || root.CachedScene().IsEmpty() {
		t.Fatal("root CachedScene is nil/empty after PaintBoundaryLayers")
	}
	t.Logf("root scene empty=%v", root.CachedScene().IsEmpty())

	// Verify spinner scene recorded.
	if spinner.CachedScene() == nil {
		t.Fatal("spinner CachedScene is nil — PaintBoundaryLayers didn't reach spinner")
	}
	if spinner.CachedScene().IsEmpty() {
		t.Fatal("spinner CachedScene is empty — spinner.Draw didn't produce scene content")
	}
	t.Logf("spinner scene empty=%v", spinner.CachedScene().IsEmpty())

	// Step 2: BuildLayerTree
	layerTree := BuildLayerTree(root)
	t.Logf("layer tree root children: %d", len(layerTree.Children()))

	// Walk layer tree and count PictureLayers with non-empty scenes.
	nonEmptyPictures := 0
	var walkLayers func(compositor.Layer, string)
	walkLayers = func(l compositor.Layer, indent string) {
		if po, ok := l.(compositor.PictureOwner); ok {
			pic := po.Picture()
			empty := pic == nil || pic.IsEmpty()
			t.Logf("%sPictureLayer: empty=%v", indent, empty)
			if !empty {
				nonEmptyPictures++
			}
		}
		if cl, ok := l.(compositor.ContainerLayer); ok {
			t.Logf("%sContainerLayer: offset=%v children=%d", indent, l.Offset(), len(cl.Children()))
			for _, child := range cl.Children() {
				walkLayers(child, indent+"  ")
			}
		}
	}
	walkLayers(layerTree, "")

	if nonEmptyPictures < 2 {
		t.Errorf("expected >= 2 non-empty PictureLayers (root + spinner), got %d", nonEmptyPictures)
	}

	// Step 3: Compositor.Compose
	comp := compositor.New()
	composed := comp.Compose(layerTree)

	if composed.IsEmpty() {
		t.Fatal("composed scene is EMPTY — spinner invisible in final output")
	}
	bounds := composed.Bounds()
	t.Logf("composed scene: empty=%v, version=%d, bounds=(%f,%f)-(%f,%f)",
		composed.IsEmpty(), composed.Version(),
		bounds.MinX, bounds.MinY, bounds.MaxX, bounds.MaxY)

	// Bounds must extend to spinner area (100+48=148, 210+48=258).
	if bounds.MaxX < 140 {
		t.Errorf("composed bounds.MaxX=%f, want >= 148 (spinner at X=100, width=48)", bounds.MaxX)
	}
	if bounds.MaxY < 250 {
		t.Errorf("composed bounds.MaxY=%f, want >= 258 (spinner at Y=210, height=48)", bounds.MaxY)
	}
}
