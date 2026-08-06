package app

import (
	"fmt"
	"testing"

	"github.com/gogpu/gg/scene"
	"github.com/gogpu/ui/core/listview"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	internalRender "github.com/gogpu/ui/internal/render"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
)

// itemWidget is a minimal widget used as a ListView item in tests.
// It draws a colored rectangle so its scene is non-empty.
type itemWidget struct {
	widget.WidgetBase
	index     int
	drawCount int
}

func newItemWidget(index int) *itemWidget {
	w := &itemWidget{index: index}
	w.SetVisible(true)
	w.SetEnabled(true)
	return w
}

func (w *itemWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	// Fixed height items for predictable test behavior.
	return c.Constrain(geometry.Sz(c.MaxWidth, 48))
}

func (w *itemWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	w.drawCount++
	// Draw a colored rectangle so the recorded scene is non-empty.
	canvas.DrawRect(w.Bounds(), widget.RGBA8(100, 150, 200, 255))
}

func (w *itemWidget) Event(_ widget.Context, _ event.Event) bool { return false }
func (w *itemWidget) Children() []widget.Widget                  { return nil }

// TestDrawChildSkip_ListViewItemBoundaries is the primary diagnostic test for
// the DrawChild skip pattern with ListView items.
//
// Scenario: root (boundary) contains a ListView with 5 items. Each item is
// automatically wrapped as a RepaintBoundary by the widget cache.
//
// Expected behavior (Flutter paintChild pattern):
//  1. PaintBoundaryLayers records root boundary -> calls root.Draw()
//  2. root.Draw -> ListView.Draw -> ScrollView.Draw -> VirtualContent.Draw
//  3. VirtualContent.Draw populates cache (items created, bounds set)
//  4. DrawChild SKIPS boundary items during recording (BoundaryRecorder)
//  5. After root recording, PaintBoundaryLayers RECURSES into children
//  6. Recursion reaches virtualContent.Children() -> finds item widgets
//  7. Each item has IsRepaintBoundary=true, sceneDirty=true -> recordBoundary
//  8. Item scenes are recorded with their content
//
// This test verifies every step of this chain.
func TestDrawChildSkip_ListViewItemBoundaries(t *testing.T) {
	// Register SceneRecorder factory (required for boundary recording).
	prev := widget.GetSceneRecorderFactory()
	widget.RegisterSceneRecorder(func(s widget.SceneCache, w, h int) (widget.Canvas, func()) {
		rec := internalRender.NewSceneCanvas(s.(*scene.Scene), w, h)
		return rec, rec.Close
	})
	defer widget.RegisterSceneRecorder(prev)

	const itemCount = 5

	// Build ListView with simple item widgets.
	lv := listview.New(
		listview.ItemCount(itemCount),
		listview.FixedItemHeight(48),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget {
			return newItemWidget(ctx.Index)
		}),
	)

	// Root container (boundary) containing the ListView.
	root := &listViewTestContainer{kids: []widget.Widget{lv}}
	root.SetVisible(true)
	root.SetEnabled(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 400, 600))
	root.SetScreenOrigin(geometry.Pt(0, 0))

	// Layout the tree so widgets have proper dimensions.
	ctx := widget.NewContext()
	ctx.SetOnInvalidateRect(func(_ geometry.Rect) {})
	rootConstraints := geometry.Tight(geometry.Sz(400, 600))
	root.Layout(ctx, rootConstraints)

	// Set ListView bounds after layout.
	lv.SetBounds(geometry.NewRect(0, 0, 400, 600))
	lv.Layout(ctx, rootConstraints)

	// Mount the tree to wire parent chain.
	widget.MountTree(root, ctx)

	// STEP 1: Verify root is dirty before painting.
	if !root.IsSceneDirty() {
		t.Fatal("root should be sceneDirty=true before first PaintBoundaryLayers")
	}

	// STEP 2: Paint boundary layers -- this is the function under test.
	PaintBoundaryLayersWithContext(root, nil, ctx)

	// STEP 3: Verify root was recorded (has non-nil, non-empty scene).
	rootScene := root.CachedScene()
	if rootScene == nil {
		t.Fatal("root.CachedScene() is nil after PaintBoundaryLayers")
	}
	if rootScene.IsEmpty() {
		t.Fatal("root.CachedScene() is empty -- root Draw was not recorded properly")
	}

	// STEP 4: Collect item widgets via tree traversal.
	// After root recording, virtualContent.Draw populated the cache.
	// virtualContent.Children() should return the item widgets.
	items := collectBoundaryDescendants(root)
	t.Logf("found %d boundary descendants (excluding root)", len(items))

	if len(items) == 0 {
		// Detailed diagnostics: walk the tree manually.
		t.Log("=== DIAGNOSTIC: Walking tree to find items ===")
		walkTreeDiag(t, root, 0)
		t.Fatal("no boundary descendants found -- items not visible to tree walk")
	}

	// STEP 5: Verify each item.
	for i, item := range items {
		// 5a: Item has IsRepaintBoundary.
		bc, ok := item.(interface{ IsRepaintBoundary() bool })
		if !ok || !bc.IsRepaintBoundary() {
			t.Errorf("item[%d]: IsRepaintBoundary should be true", i)
			continue
		}

		// 5b: Item has non-zero bounds.
		bg, ok := item.(interface{ Bounds() geometry.Rect })
		if !ok {
			t.Errorf("item[%d]: does not implement Bounds()", i)
			continue
		}
		bounds := bg.Bounds()
		if bounds.Width() <= 0 || bounds.Height() <= 0 {
			t.Errorf("item[%d]: bounds are zero/negative: %v (width=%.1f, height=%.1f)",
				i, bounds, bounds.Width(), bounds.Height())
			continue
		}
		t.Logf("item[%d]: bounds=%v (%.0fx%.0f)", i, bounds, bounds.Width(), bounds.Height())

		// 5c: Item has cached scene (recorded by PaintBoundaryLayers recursion).
		sc, ok := item.(interface{ CachedScene() widget.SceneCache })
		if !ok {
			t.Errorf("item[%d]: does not implement CachedScene()", i)
			continue
		}
		cachedScene := sc.CachedScene()
		if cachedScene == nil {
			t.Errorf("item[%d]: CachedScene is nil -- PaintBoundaryLayers did not "+
				"reach this boundary during recursion", i)
			continue
		}

		// 5d: Item scene is non-empty (has actual draw commands).
		if cachedScene.IsEmpty() {
			t.Errorf("item[%d]: CachedScene is empty -- recordBoundary was called "+
				"but item.Draw() did not produce any draw commands", i)
			continue
		}

		t.Logf("item[%d]: OK (scene recorded, non-empty)", i)
	}

	// STEP 6: Verify we found the expected number of items.
	if len(items) < itemCount {
		t.Errorf("expected at least %d item boundaries, found %d", itemCount, len(items))
	}
}

// TestDrawChildSkip_RootRecordingSkipsItems verifies that during root boundary
// recording, DrawChild correctly skips child boundaries (BoundaryRecorder check).
// Items should NOT appear in the root's scene -- they have their own scenes.
func TestDrawChildSkip_RootRecordingSkipsItems(t *testing.T) {
	prev := widget.GetSceneRecorderFactory()
	widget.RegisterSceneRecorder(func(s widget.SceneCache, w, h int) (widget.Canvas, func()) {
		rec := internalRender.NewSceneCanvas(s.(*scene.Scene), w, h)
		return rec, rec.Close
	})
	defer widget.RegisterSceneRecorder(prev)

	// Track which boundaries get recorded.
	var recordedBoundaries []string

	const itemCount = 3
	lv := listview.New(
		listview.ItemCount(itemCount),
		listview.FixedItemHeight(48),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget {
			return newItemWidget(ctx.Index)
		}),
	)

	root := &recordingContainer{
		name: "root",
		kids: []widget.Widget{lv},
		onDraw: func(name string) {
			recordedBoundaries = append(recordedBoundaries, name)
		},
	}
	root.SetVisible(true)
	root.SetEnabled(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 400, 300))
	root.SetScreenOrigin(geometry.Pt(0, 0))

	ctx := widget.NewContext()
	ctx.SetOnInvalidateRect(func(_ geometry.Rect) {})

	constraints := geometry.Tight(geometry.Sz(400, 300))
	root.Layout(ctx, constraints)
	lv.SetBounds(geometry.NewRect(0, 0, 400, 300))
	lv.Layout(ctx, constraints)

	widget.MountTree(root, ctx)

	PaintBoundaryLayersWithContext(root, nil, ctx)

	// Root should be in the recorded list (its Draw was called).
	if len(recordedBoundaries) == 0 || recordedBoundaries[0] != "root" {
		t.Errorf("root Draw was not called, recorded=%v", recordedBoundaries)
	}
	t.Logf("recorded boundaries: %v", recordedBoundaries)

	// After PaintBoundaryLayers, item boundaries should also have scenes.
	items := collectBoundaryDescendants(root)
	for i, item := range items {
		if sc, ok := item.(interface{ CachedScene() widget.SceneCache }); ok {
			cs := sc.CachedScene()
			if cs == nil {
				t.Errorf("item[%d]: CachedScene nil after PaintBoundaryLayers", i)
			} else if cs.IsEmpty() {
				t.Errorf("item[%d]: CachedScene empty after PaintBoundaryLayers", i)
			}
		}
	}
}

// TestDrawChildSkip_ItemsExistAfterRootRecording verifies that item widgets
// exist in the tree (via Children()) AFTER root recording completes, even
// though they were created dynamically during VirtualContent.Draw().
func TestDrawChildSkip_ItemsExistAfterRootRecording(t *testing.T) {
	prev := widget.GetSceneRecorderFactory()
	widget.RegisterSceneRecorder(func(s widget.SceneCache, w, h int) (widget.Canvas, func()) {
		rec := internalRender.NewSceneCanvas(s.(*scene.Scene), w, h)
		return rec, rec.Close
	})
	defer widget.RegisterSceneRecorder(prev)

	const itemCount = 5
	lv := listview.New(
		listview.ItemCount(itemCount),
		listview.FixedItemHeight(48),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget {
			return newItemWidget(ctx.Index)
		}),
	)

	root := &listViewTestContainer{kids: []widget.Widget{lv}}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 400, 600))

	ctx := widget.NewContext()
	ctx.SetOnInvalidateRect(func(_ geometry.Rect) {})
	constraints := geometry.Tight(geometry.Sz(400, 600))
	root.Layout(ctx, constraints)
	lv.SetBounds(geometry.NewRect(0, 0, 400, 600))
	lv.Layout(ctx, constraints)

	// BEFORE root recording: items should NOT exist yet.
	itemsBefore := collectBoundaryDescendants(root)
	t.Logf("items BEFORE root recording: %d", len(itemsBefore))

	// Record root boundary only (simulates what recordBoundary does).
	rootScene := scene.NewScene()
	recorder, cleanup := widget.GetSceneRecorderFactory()(rootScene, 400, 600)
	recorder.PushTransform(geometry.Pt(0, 0))
	root.Draw(ctx, recorder)
	recorder.PopTransform()
	cleanup()

	// AFTER root recording: items should exist (cache populated by VirtualContent.Draw).
	itemsAfter := collectBoundaryDescendants(root)
	t.Logf("items AFTER root recording: %d", len(itemsAfter))

	if len(itemsAfter) == 0 {
		t.Log("=== DIAGNOSTIC: Walking tree after root.Draw ===")
		walkTreeDiag(t, root, 0)
		t.Fatal("no items found after root recording -- VirtualContent.Draw did not " +
			"populate cache, or virtualContent.Children() does not expose cached items")
	}

	// Verify items have valid bounds (set during VirtualContent.Draw).
	for i, item := range itemsAfter {
		if bg, ok := item.(interface{ Bounds() geometry.Rect }); ok {
			bounds := bg.Bounds()
			if bounds.Width() <= 0 || bounds.Height() <= 0 {
				t.Errorf("item[%d]: bounds invalid after root recording: %v", i, bounds)
			} else {
				t.Logf("item[%d]: bounds=%v OK", i, bounds)
			}
		}
	}
}

// TestDrawChildSkip_BoxTextItems_ProductionScenario tests the exact scenario
// from the hello example: ListView items are primitives.Box(primitives.Text(...)).
// This verifies that PaintBoundaryLayers records item scenes that contain
// both the Box background and the Text content.
func TestDrawChildSkip_BoxTextItems_ProductionScenario(t *testing.T) {
	prev := widget.GetSceneRecorderFactory()
	widget.RegisterSceneRecorder(func(s widget.SceneCache, w, h int) (widget.Canvas, func()) {
		rec := internalRender.NewSceneCanvas(s.(*scene.Scene), w, h)
		return rec, rec.Close
	})
	defer widget.RegisterSceneRecorder(prev)

	const itemCount = 5

	// Build ListView with Box(Text) items -- same pattern as hello example.
	lv := listview.New(
		listview.ItemCount(itemCount),
		listview.FixedItemHeight(36),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget {
			return primitives.Box(
				primitives.Text(fmt.Sprintf("Item %d", ctx.Index)).
					FontSize(14).
					Color(widget.RGBA8(33, 33, 33, 255)),
			).PaddingXY(12, 8)
		}),
	)

	root := &listViewTestContainer{kids: []widget.Widget{lv}}
	root.SetVisible(true)
	root.SetEnabled(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 400, 600))
	root.SetScreenOrigin(geometry.Pt(0, 0))

	ctx := widget.NewContext()
	ctx.SetOnInvalidateRect(func(_ geometry.Rect) {})
	constraints := geometry.Tight(geometry.Sz(400, 600))
	root.Layout(ctx, constraints)
	lv.SetBounds(geometry.NewRect(0, 0, 400, 600))
	lv.Layout(ctx, constraints)
	widget.MountTree(root, ctx)

	// Paint all boundaries.
	PaintBoundaryLayersWithContext(root, nil, ctx)

	// Verify root recorded.
	if root.CachedScene() == nil || root.CachedScene().IsEmpty() {
		t.Fatal("root scene should be non-nil and non-empty")
	}

	// Verify item boundaries.
	items := collectBoundaryDescendants(root)
	t.Logf("found %d boundary descendants", len(items))

	if len(items) < itemCount {
		t.Log("=== DIAGNOSTIC: Tree after PaintBoundaryLayers ===")
		walkTreeDiag(t, root, 0)
		t.Fatalf("expected at least %d items, found %d", itemCount, len(items))
	}

	for i, item := range items {
		sc, ok := item.(interface{ CachedScene() widget.SceneCache })
		if !ok {
			t.Errorf("item[%d]: does not implement CachedScene()", i)
			continue
		}
		cs := sc.CachedScene()
		if cs == nil {
			t.Errorf("item[%d]: CachedScene nil", i)
			continue
		}
		if cs.IsEmpty() {
			t.Errorf("item[%d]: CachedScene empty -- Box+Text content not recorded", i)
			continue
		}
		t.Logf("item[%d]: OK (Box+Text scene recorded)", i)
	}
}

// --- Test Helpers ---

// listViewTestContainer is a simple container that draws children via DrawChild.
type listViewTestContainer struct {
	widget.WidgetBase
	kids []widget.Widget
}

func (w *listViewTestContainer) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	// Layout children.
	for _, child := range w.kids {
		child.Layout(ctx, c)
	}
	return c.Constrain(geometry.Sz(400, 600))
}

func (w *listViewTestContainer) Draw(ctx widget.Context, canvas widget.Canvas) {
	canvas.DrawRect(w.Bounds(), widget.RGBA8(240, 240, 240, 255))
	for _, child := range w.kids {
		widget.StampScreenOrigin(child, canvas)
		widget.DrawChild(child, ctx, canvas)
	}
}

func (w *listViewTestContainer) Event(_ widget.Context, _ event.Event) bool { return false }
func (w *listViewTestContainer) Children() []widget.Widget                  { return w.kids }

// recordingContainer records when its Draw is called.
type recordingContainer struct {
	widget.WidgetBase
	name   string
	kids   []widget.Widget
	onDraw func(name string)
}

func (w *recordingContainer) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	for _, child := range w.kids {
		child.Layout(ctx, c)
	}
	return c.Constrain(geometry.Sz(400, 300))
}

func (w *recordingContainer) Draw(ctx widget.Context, canvas widget.Canvas) {
	if w.onDraw != nil {
		w.onDraw(w.name)
	}
	canvas.DrawRect(w.Bounds(), widget.RGBA8(240, 240, 240, 255))
	for _, child := range w.kids {
		widget.StampScreenOrigin(child, canvas)
		widget.DrawChild(child, ctx, canvas)
	}
}

func (w *recordingContainer) Event(_ widget.Context, _ event.Event) bool { return false }
func (w *recordingContainer) Children() []widget.Widget                  { return w.kids }

// collectBoundaryDescendants walks the widget tree and returns all widgets
// (excluding root) that have IsRepaintBoundary=true.
func collectBoundaryDescendants(root widget.Widget) []widget.Widget {
	var result []widget.Widget
	collectBoundaryDescendantsRecursive(root, &result, true)
	return result
}

func collectBoundaryDescendantsRecursive(w widget.Widget, result *[]widget.Widget, isRoot bool) {
	if w == nil {
		return
	}

	if !isRoot {
		if bc, ok := w.(interface{ IsRepaintBoundary() bool }); ok && bc.IsRepaintBoundary() {
			*result = append(*result, w)
		}
	}

	for _, child := range w.Children() {
		collectBoundaryDescendantsRecursive(child, result, false)
	}
}

// walkTreeDiag prints a diagnostic tree walk showing widget types, bounds,
// and boundary status.
func walkTreeDiag(t *testing.T, w widget.Widget, depth int) {
	t.Helper()
	if w == nil {
		return
	}

	indent := ""
	for range depth {
		indent += "  "
	}

	isBoundary := false
	if bc, ok := w.(interface{ IsRepaintBoundary() bool }); ok {
		isBoundary = bc.IsRepaintBoundary()
	}

	bounds := geometry.Rect{}
	if bg, ok := w.(interface{ Bounds() geometry.Rect }); ok {
		bounds = bg.Bounds()
	}

	sceneDirty := false
	hasScene := false
	if sd, ok := w.(interface{ IsSceneDirty() bool }); ok {
		sceneDirty = sd.IsSceneDirty()
	}
	if sc, ok := w.(interface{ CachedScene() widget.SceneCache }); ok {
		hasScene = sc.CachedScene() != nil
	}

	t.Logf("%s%T boundary=%v bounds=%v sceneDirty=%v hasScene=%v children=%d",
		indent, w, isBoundary, bounds, sceneDirty, hasScene, len(w.Children()))

	for _, child := range w.Children() {
		walkTreeDiag(t, child, depth+1)
	}
}
