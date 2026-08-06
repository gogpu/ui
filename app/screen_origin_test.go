package app

import (
	"testing"

	"github.com/gogpu/gg/scene"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	internalRender "github.com/gogpu/ui/internal/render"
	"github.com/gogpu/ui/widget"
)

// TestRecordBoundary_ScreenOriginBase verifies that recordBoundary sets
// ScreenOriginBase on the recorder canvas so that nested StampScreenOrigin
// calls produce correct screen-space origins.
//
// Without ScreenOriginBase: children get ScreenOrigin relative to (0,0)
// instead of relative to the boundary's screen position → items render
// at window top-left corner.
func TestRecordBoundary_ScreenOriginBase(t *testing.T) {
	prev := widget.GetSceneRecorderFactory()
	widget.RegisterSceneRecorder(testSceneRecorder)
	defer widget.RegisterSceneRecorder(prev)

	// Root boundary at screen position (0,0).
	root := &screenOriginContainer{}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))
	root.SetScreenOrigin(geometry.Pt(0, 0))

	// ScrollView boundary inside root at position (56, 481).
	// Bounds are relative to parent (root), so Min = (56, 481).
	scrollView := &screenOriginContainer{}
	scrollView.SetVisible(true)
	scrollView.SetRepaintBoundary(true)
	scrollView.SetBounds(geometry.NewRect(56, 481, 672, 300))
	scrollView.SetParent(root)
	root.kids = []widget.Widget{scrollView}

	// Item inside scrollView at local position (0, 100).
	item := &screenOriginLeaf{}
	item.SetVisible(true)
	item.SetRepaintBoundary(true)
	item.SetBounds(geometry.NewRect(0, 100, 672, 32))
	item.SetParent(scrollView)
	scrollView.kids = []widget.Widget{item}

	ctx := widget.NewContext()
	ctx.SetOnInvalidateRect(func(_ geometry.Rect) {})

	// Record ScrollView boundary. This should set ScreenOriginBase on
	// the recorder so StampScreenOrigin inside Draw computes correct values.
	PaintBoundaryLayersWithContext(root, nil, ctx)

	// Item's ScreenOrigin should be ScrollView's screen pos + item's local pos.
	// ScrollView screen = (56, 481), item local = (0, 100) → item screen = (56, 581).
	gotOrigin := item.ScreenOrigin()
	wantOrigin := geometry.Pt(56, 581)

	if gotOrigin != wantOrigin {
		t.Errorf("item ScreenOrigin = %v, want %v\n"+
			"  scrollView.ScreenOrigin = %v\n"+
			"  item.Bounds.Min = %v\n"+
			"  If (0, 100): ScreenOriginBase not set on recorder canvas",
			gotOrigin, wantOrigin,
			scrollView.ScreenOrigin(),
			item.Bounds().Min,
		)
	}
}

// TestRecordBoundary_RootScreenOriginBase verifies that root boundary
// (ScreenOrigin=0,0) produces correct child origins.
func TestRecordBoundary_RootScreenOriginBase(t *testing.T) {
	prev := widget.GetSceneRecorderFactory()
	widget.RegisterSceneRecorder(testSceneRecorder)
	defer widget.RegisterSceneRecorder(prev)

	root := &screenOriginContainer{}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))
	root.SetScreenOrigin(geometry.Pt(0, 0))

	child := &screenOriginLeaf{}
	child.SetVisible(true)
	child.SetRepaintBoundary(true)
	child.SetBounds(geometry.NewRect(100, 200, 200, 48))
	child.SetParent(root)
	root.kids = []widget.Widget{child}

	ctx := widget.NewContext()
	ctx.SetOnInvalidateRect(func(_ geometry.Rect) {})

	PaintBoundaryLayersWithContext(root, nil, ctx)

	gotOrigin := child.ScreenOrigin()
	wantOrigin := geometry.Pt(100, 200)

	if gotOrigin != wantOrigin {
		t.Errorf("child ScreenOrigin = %v, want %v", gotOrigin, wantOrigin)
	}
}

// --- test helpers ---

type screenOriginContainer struct {
	widget.WidgetBase
	kids []widget.Widget
}

func (w *screenOriginContainer) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(800, 600))
}

func (w *screenOriginContainer) Draw(ctx widget.Context, canvas widget.Canvas) {
	for _, child := range w.kids {
		widget.StampScreenOrigin(child, canvas)
		widget.DrawChild(child, ctx, canvas)
	}
}

func (w *screenOriginContainer) Event(_ widget.Context, _ event.Event) bool { return false }
func (w *screenOriginContainer) Children() []widget.Widget                  { return w.kids }

type screenOriginLeaf struct {
	widget.WidgetBase
}

func (w *screenOriginLeaf) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(672, 32))
}

func (w *screenOriginLeaf) Draw(_ widget.Context, canvas widget.Canvas) {
	canvas.DrawRect(w.Bounds(), widget.RGBA8(100, 100, 100, 255))
}

func (w *screenOriginLeaf) Event(_ widget.Context, _ event.Event) bool { return false }
func (w *screenOriginLeaf) Children() []widget.Widget                  { return nil }

// testSceneRecorder is defined in compositor_test.go but redeclared here
// for this test file. Uses the same pattern.
func testSceneRecorderForOriginTests(s widget.SceneCache, w, h int) (widget.Canvas, func()) { //nolint:unused // retained for future screen origin test variants
	rec := internalRender.NewSceneCanvas(s.(*scene.Scene), w, h)
	return rec, rec.Close
}
