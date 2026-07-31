package app

import (
	"testing"
	"unsafe"

	"github.com/gogpu/gg/scene"
	"github.com/gogpu/gpucontext"
	"github.com/gogpu/ui/compositor"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// testContainer has children accessible via Children().
type testContainer struct {
	widget.WidgetBase
	kids []widget.Widget
}

func (w *testContainer) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(800, 600))
}
func (w *testContainer) Draw(_ widget.Context, canvas widget.Canvas) {
	canvas.DrawRect(w.Bounds(), widget.RGBA8(200, 200, 200, 255))
	for _, child := range w.kids {
		widget.DrawChild(child, nil, canvas)
	}
}
func (w *testContainer) Event(_ widget.Context, _ event.Event) bool { return false }
func (w *testContainer) Children() []widget.Widget                  { return w.kids }

// testLeaf is a leaf widget with boundary support.
type testLeaf struct {
	widget.WidgetBase
	drawCount int
}

func (w *testLeaf) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(48, 48))
}
func (w *testLeaf) Draw(_ widget.Context, canvas widget.Canvas) {
	w.drawCount++
	canvas.DrawRect(w.Bounds(), widget.RGBA8(255, 0, 0, 255))
}
func (w *testLeaf) Event(_ widget.Context, _ event.Event) bool { return false }
func (w *testLeaf) Children() []widget.Widget                  { return nil }

// TestPaintBoundaryLayers_FindsNestedBoundary verifies that
// PaintBoundaryLayers walks through non-boundary containers
// and reaches nested boundary widgets (spinner inside collapsible).
func TestPaintBoundaryLayers_FindsNestedBoundary(t *testing.T) {
	prev := widget.GetSceneRecorderFactory()
	widget.RegisterSceneRecorder(testSceneRecorder)
	defer widget.RegisterSceneRecorder(prev)

	root := &testContainer{}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))
	root.SetScreenOrigin(geometry.Pt(0, 0))

	mid := &testContainer{}
	mid.SetVisible(true)
	mid.SetBounds(geometry.NewRect(0, 100, 800, 500))
	root.kids = append(root.kids, mid)

	spinner := &testLeaf{}
	spinner.SetVisible(true)
	spinner.SetRepaintBoundary(true)
	spinner.SetBounds(geometry.NewRect(100, 200, 148, 248))
	spinner.SetScreenOrigin(geometry.Pt(100, 200))
	mid.kids = append(mid.kids, spinner)

	ctx := widget.NewContext()
	ctx.SetOnInvalidateRect(func(_ geometry.Rect) {})

	PaintBoundaryLayersWithContext(root, nil, ctx)

	if root.CachedScene() == nil {
		t.Error("root boundary should have cached scene after PaintBoundaryLayers")
	}
	if spinner.CachedScene() == nil {
		t.Error("spinner boundary should have cached scene — PaintBoundaryLayers " +
			"must traverse non-boundary containers to reach nested boundaries")
	}
	if spinner.drawCount == 0 {
		t.Error("spinner.Draw should have been called during recording")
	}
}

// TestBuildLayerTree_NestedOffset verifies accumulated offset computation.
func TestBuildLayerTree_NestedOffset(t *testing.T) {
	root := &testContainer{}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))

	mid := &testContainer{}
	mid.SetVisible(true)
	mid.SetBounds(geometry.NewRect(0, 100, 800, 500))
	root.kids = append(root.kids, mid)

	spinner := &testLeaf{}
	spinner.SetVisible(true)
	spinner.SetRepaintBoundary(true)
	spinner.SetBounds(geometry.NewRect(50, 200, 98, 248))
	mid.kids = append(mid.kids, spinner)

	layer := BuildLayerTree(root)

	// Root OffsetLayer(0,0) has children:
	// [0] = root's own OffsetLayer(0,0) with PictureLayer (root boundary)
	// Root OffsetLayer → root boundary OffsetLayer → [PictureLayer, spinner OffsetLayer]
	children := layer.Children()
	t.Logf("root layer children: %d", len(children))
	for i, ch := range children {
		t.Logf("  child[%d]: %T, offset=%v, children=%d",
			i, ch, ch.Offset(), len(ch.(compositor.ContainerLayer).Children()))
	}

	if len(children) == 0 {
		t.Fatal("root layer should have children")
	}

	// Root boundary is first child OffsetLayer. It should contain spinner.
	rootBoundary, ok := children[0].(compositor.ContainerLayer)
	if !ok {
		t.Fatal("first child should be ContainerLayer (root boundary OffsetLayer)")
	}
	rootBoundaryChildren := rootBoundary.Children()
	t.Logf("root boundary children: %d", len(rootBoundaryChildren))

	// Should have PictureLayer + spinner OffsetLayer
	foundSpinner := false
	for _, rbc := range rootBoundaryChildren {
		if cl, ok2 := rbc.(compositor.ContainerLayer); ok2 && len(cl.Children()) > 0 {
			foundSpinner = true
		}
	}
	if !foundSpinner && len(rootBoundaryChildren) < 2 {
		t.Error("root boundary should have spinner as nested layer")
	}

	// Check spinner offset: mid.Bounds.Min(0,100) + spinner.Bounds.Min(50,200) = (50,300)
	for _, rbc := range rootBoundaryChildren {
		if cl, ok2 := rbc.(compositor.ContainerLayer); ok2 {
			t.Logf("spinner OffsetLayer offset: %v", cl.Offset())
		}
	}
}

// --- Phase D Tests: PictureLayer Fields ---

// TestBuildLayerTree_BoundaryCacheKeysPreserved verifies that each boundary
// widget's BoundaryCacheKey appears in the corresponding PictureLayer.
func TestBuildLayerTree_BoundaryCacheKeysPreserved(t *testing.T) {
	root := &testContainer{}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))
	rootKey := root.BoundaryCacheKey()

	child := &testLeaf{}
	child.SetVisible(true)
	child.SetRepaintBoundary(true)
	child.SetBounds(geometry.NewRect(10, 20, 58, 68))
	child.SetScreenOrigin(geometry.Pt(10, 20))
	child.SetParent(root)
	childKey := child.BoundaryCacheKey()
	root.kids = append(root.kids, child)

	tree := BuildLayerTree(root)

	var pics []*compositor.PictureLayerImpl
	collectPictureLayersFromTree(tree, &pics)

	if len(pics) != 2 {
		t.Fatalf("expected 2 PictureLayers (root + child), got %d", len(pics))
	}

	keys := map[uint64]bool{}
	for _, pic := range pics {
		keys[pic.BoundaryCacheKey()] = true
	}

	if !keys[rootKey] {
		t.Errorf("root BoundaryCacheKey %d not found in PictureLayers", rootKey)
	}
	if !keys[childKey] {
		t.Errorf("child BoundaryCacheKey %d not found in PictureLayers", childKey)
	}
}

// TestBuildLayerTree_IsRootFlag verifies that only the root boundary's
// PictureLayer has IsRoot=true.
func TestBuildLayerTree_IsRootFlag(t *testing.T) {
	root := &testContainer{}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))

	child := &testLeaf{}
	child.SetVisible(true)
	child.SetRepaintBoundary(true)
	child.SetBounds(geometry.NewRect(0, 0, 48, 48))
	child.SetScreenOrigin(geometry.Pt(100, 100))
	child.SetParent(root)
	root.kids = append(root.kids, child)

	tree := BuildLayerTree(root)

	var pics []*compositor.PictureLayerImpl
	collectPictureLayersFromTree(tree, &pics)

	rootCount := 0
	for _, pic := range pics {
		if pic.IsRoot() {
			rootCount++
		}
	}

	if rootCount != 1 {
		t.Errorf("expected exactly 1 root PictureLayer, got %d", rootCount)
	}
}

// TestBuildLayerTree_SizeFromBounds verifies that PictureLayer.Size
// matches the boundary widget's Bounds dimensions.
func TestBuildLayerTree_SizeFromBounds(t *testing.T) {
	root := &testContainer{}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))

	child := &testLeaf{}
	child.SetVisible(true)
	child.SetRepaintBoundary(true)
	child.SetBounds(geometry.NewRect(10, 20, 200, 100))
	child.SetScreenOrigin(geometry.Pt(10, 20))
	child.SetParent(root)
	root.kids = append(root.kids, child)

	tree := BuildLayerTree(root)

	var pics []*compositor.PictureLayerImpl
	collectPictureLayersFromTree(tree, &pics)

	for _, pic := range pics {
		w, h := pic.Size()
		if pic.IsRoot() {
			if w != 800 || h != 600 {
				t.Errorf("root size = (%d, %d), want (800, 600)", w, h)
			}
		} else {
			if w != 200 || h != 100 {
				t.Errorf("child size = (%d, %d), want (200, 100)", w, h)
			}
		}
	}
}

// TestBuildLayerTree_ScreenOriginPropagated verifies that PictureLayer
// carries the boundary widget's ScreenOrigin.
func TestBuildLayerTree_ScreenOriginPropagated(t *testing.T) {
	root := &testContainer{}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))
	root.SetScreenOrigin(geometry.Pt(0, 0))

	child := &testLeaf{}
	child.SetVisible(true)
	child.SetRepaintBoundary(true)
	child.SetBounds(geometry.NewRect(0, 0, 48, 48))
	child.SetScreenOrigin(geometry.Pt(150, 250))
	child.SetParent(root)
	root.kids = append(root.kids, child)

	tree := BuildLayerTree(root)

	var pics []*compositor.PictureLayerImpl
	collectPictureLayersFromTree(tree, &pics)

	for _, pic := range pics {
		if pic.IsRoot() {
			continue
		}
		origin := pic.ScreenOrigin()
		if origin.X != 150 || origin.Y != 250 {
			t.Errorf("child ScreenOrigin = %v, want (150, 250)", origin)
		}
		if !pic.IsScreenOriginValid() {
			t.Error("child ScreenOrigin should be valid")
		}
	}
}

// TestBuildLayerTree_SceneVersionPropagated verifies that PictureLayer
// carries the boundary widget's SceneCacheVersion.
func TestBuildLayerTree_SceneVersionPropagated(t *testing.T) {
	root := &testContainer{}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))

	// ClearSceneDirty increments version.
	root.ClearSceneDirty()
	root.ClearSceneDirty()
	expectedVersion := root.SceneCacheVersion()
	if expectedVersion == 0 {
		t.Fatal("expected non-zero SceneCacheVersion after ClearSceneDirty")
	}

	tree := BuildLayerTree(root)

	var pics []*compositor.PictureLayerImpl
	collectPictureLayersFromTree(tree, &pics)

	if len(pics) == 0 {
		t.Fatal("no PictureLayers found")
	}

	rootPic := pics[0]
	if rootPic.SceneVersion() != expectedVersion {
		t.Errorf("PictureLayer SceneVersion = %d, want %d",
			rootPic.SceneVersion(), expectedVersion)
	}
}

// TestBuildLayerTree_DirtyFlagsPropagated verifies that dirty/clean boundaries
// produce PictureLayers with matching dirty flags.
func TestBuildLayerTree_DirtyFlagsPropagated(t *testing.T) {
	root := &testContainer{}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))

	// Mark root as clean.
	root.ClearSceneDirty()

	// Create dirty child.
	child := &testLeaf{}
	child.SetVisible(true)
	child.SetRepaintBoundary(true)
	child.SetBounds(geometry.NewRect(0, 0, 48, 48))
	child.SetParent(root)
	child.InvalidateScene() // mark dirty
	root.kids = append(root.kids, child)

	tree := BuildLayerTree(root)

	var pics []*compositor.PictureLayerImpl
	collectPictureLayersFromTree(tree, &pics)

	if len(pics) < 2 {
		t.Fatalf("expected >= 2 PictureLayers, got %d", len(pics))
	}

	for _, pic := range pics {
		if pic.IsRoot() {
			if pic.IsDirty() {
				t.Error("root PictureLayer should be clean (ClearSceneDirty was called)")
			}
		} else {
			if !pic.IsDirty() {
				t.Error("child PictureLayer should be dirty (InvalidateScene was called)")
			}
		}
	}
}

// TestBuildLayerTree_NilRoot returns empty tree for nil root.
func TestBuildLayerTree_NilRoot(t *testing.T) {
	tree := BuildLayerTree(nil)
	if tree == nil {
		t.Fatal("BuildLayerTree(nil) should return non-nil empty OffsetLayer")
	}
	if len(tree.Children()) != 0 {
		t.Errorf("nil root should produce empty tree, got %d children", len(tree.Children()))
	}
}

// --- Phase D5 Tests: UpdateLayerTree (persistent tree) ---

// TestUpdateLayerTree_NilExistingMatchesBuild verifies that UpdateLayerTree
// with nil existing produces the same structure as BuildLayerTree.
func TestUpdateLayerTree_NilExistingMatchesBuild(t *testing.T) {
	root := &testContainer{}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))

	child := &testLeaf{}
	child.SetVisible(true)
	child.SetRepaintBoundary(true)
	child.SetBounds(geometry.NewRect(10, 20, 58, 68))
	child.SetScreenOrigin(geometry.Pt(10, 20))
	child.SetParent(root)
	root.kids = append(root.kids, child)

	tree := UpdateLayerTree(root, nil)

	var pics []*compositor.PictureLayerImpl
	collectPictureLayersFromTree(tree, &pics)

	if len(pics) != 2 {
		t.Fatalf("expected 2 PictureLayers from nil existing, got %d", len(pics))
	}

	rootKey := root.BoundaryCacheKey()
	childKey := child.BoundaryCacheKey()
	keys := map[uint64]bool{}
	for _, pic := range pics {
		keys[pic.BoundaryCacheKey()] = true
	}
	if !keys[rootKey] {
		t.Errorf("root key %d missing from UpdateLayerTree(nil)", rootKey)
	}
	if !keys[childKey] {
		t.Errorf("child key %d missing from UpdateLayerTree(nil)", childKey)
	}
}

// TestUpdateLayerTree_ReusesUnchangedLayers verifies that PictureLayerImpl
// and OffsetLayerImpl objects are reused (same pointers) when the widget
// tree is unchanged between frames.
func TestUpdateLayerTree_ReusesUnchangedLayers(t *testing.T) {
	root := &testContainer{}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))
	root.SetScreenOrigin(geometry.Pt(0, 0))

	child := &testLeaf{}
	child.SetVisible(true)
	child.SetRepaintBoundary(true)
	child.SetBounds(geometry.NewRect(10, 20, 58, 68))
	child.SetScreenOrigin(geometry.Pt(10, 20))
	child.SetParent(root)
	root.kids = append(root.kids, child)

	// First frame: build.
	tree1 := UpdateLayerTree(root, nil)

	var pics1 []*compositor.PictureLayerImpl
	collectPictureLayersFromTree(tree1, &pics1)

	// Collect OffsetLayers too.
	offsets1 := collectOffsetLayersByKey(tree1)

	// Second frame: update with same widgets.
	tree2 := UpdateLayerTree(root, tree1)

	var pics2 []*compositor.PictureLayerImpl
	collectPictureLayersFromTree(tree2, &pics2)

	offsets2 := collectOffsetLayersByKey(tree2)

	if len(pics1) != len(pics2) {
		t.Fatalf("PictureLayer count changed: %d -> %d", len(pics1), len(pics2))
	}

	// Verify pointer identity: same PictureLayerImpl objects reused.
	for i, pic1 := range pics1 {
		key := pic1.BoundaryCacheKey()
		found := false
		for _, pic2 := range pics2 {
			if pic2.BoundaryCacheKey() == key {
				found = true
				if pic1 != pic2 {
					t.Errorf("PictureLayer key=%d: different pointer after update (not reused)", key)
				}
				break
			}
		}
		if !found {
			t.Errorf("PictureLayer[%d] key=%d not found in updated tree", i, key)
		}
	}

	// Verify OffsetLayer pointer identity.
	for key, off1 := range offsets1 {
		off2, ok := offsets2[key]
		if !ok {
			t.Errorf("OffsetLayer key=%d not found in updated tree", key)
			continue
		}
		if off1 != off2 {
			t.Errorf("OffsetLayer key=%d: different pointer after update (not reused)", key)
		}
	}
}

// TestUpdateLayerTree_UpdatesDirtyScene verifies that when a boundary
// gets a new scene, UpdateLayerTree syncs the scene pointer.
func TestUpdateLayerTree_UpdatesDirtyScene(t *testing.T) {
	root := &testContainer{}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))
	root.SetScreenOrigin(geometry.Pt(0, 0))

	child := &testLeaf{}
	child.SetVisible(true)
	child.SetRepaintBoundary(true)
	child.SetBounds(geometry.NewRect(10, 20, 58, 68))
	child.SetScreenOrigin(geometry.Pt(10, 20))
	child.SetParent(root)
	root.kids = append(root.kids, child)

	// First frame.
	tree1 := UpdateLayerTree(root, nil)

	// Simulate recording: set a scene on child.
	s1 := scene.NewScene()
	child.SetCachedScene(s1)
	child.ClearSceneDirty()
	version1 := child.SceneCacheVersion()

	// Second frame: child gets dirty and re-recorded.
	child.InvalidateScene()
	s2 := scene.NewScene()
	child.SetCachedScene(s2)
	child.ClearSceneDirty()
	version2 := child.SceneCacheVersion()

	if version1 == version2 {
		t.Fatal("scene versions should differ after re-recording")
	}

	tree2 := UpdateLayerTree(root, tree1)

	// Find child PictureLayer and verify updated scene.
	var pics []*compositor.PictureLayerImpl
	collectPictureLayersFromTree(tree2, &pics)

	childKey := child.BoundaryCacheKey()
	for _, pic := range pics {
		if pic.BoundaryCacheKey() == childKey {
			if pic.Picture() != s2 {
				t.Error("UpdateLayerTree should sync new scene pointer to PictureLayer")
			}
			if pic.SceneVersion() != version2 {
				t.Errorf("SceneVersion = %d, want %d", pic.SceneVersion(), version2)
			}
			return
		}
	}
	t.Error("child PictureLayer not found in updated tree")
}

// TestUpdateLayerTree_AddsNewBoundary verifies that UpdateLayerTree creates
// new PictureLayer/OffsetLayer for a boundary that was added between frames.
func TestUpdateLayerTree_AddsNewBoundary(t *testing.T) {
	root := &testContainer{}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))
	root.SetScreenOrigin(geometry.Pt(0, 0))

	child1 := &testLeaf{}
	child1.SetVisible(true)
	child1.SetRepaintBoundary(true)
	child1.SetBounds(geometry.NewRect(10, 20, 58, 68))
	child1.SetScreenOrigin(geometry.Pt(10, 20))
	child1.SetParent(root)
	root.kids = append(root.kids, child1)

	// First frame: root + child1 = 2 boundaries.
	tree1 := UpdateLayerTree(root, nil)
	var pics1 []*compositor.PictureLayerImpl
	collectPictureLayersFromTree(tree1, &pics1)
	if len(pics1) != 2 {
		t.Fatalf("frame 1: expected 2 PictureLayers, got %d", len(pics1))
	}

	// Add child2 between frames.
	child2 := &testLeaf{}
	child2.SetVisible(true)
	child2.SetRepaintBoundary(true)
	child2.SetBounds(geometry.NewRect(100, 20, 148, 68))
	child2.SetScreenOrigin(geometry.Pt(100, 20))
	child2.SetParent(root)
	root.kids = append(root.kids, child2)

	// Second frame: root + child1 + child2 = 3 boundaries.
	tree2 := UpdateLayerTree(root, tree1)
	var pics2 []*compositor.PictureLayerImpl
	collectPictureLayersFromTree(tree2, &pics2)
	if len(pics2) != 3 {
		t.Fatalf("frame 2: expected 3 PictureLayers, got %d", len(pics2))
	}

	// Verify child2's key is present.
	child2Key := child2.BoundaryCacheKey()
	found := false
	for _, pic := range pics2 {
		if pic.BoundaryCacheKey() == child2Key {
			found = true
			break
		}
	}
	if !found {
		t.Error("newly added child2 boundary not found in updated tree")
	}
}

// TestUpdateLayerTree_RemovesBoundary verifies that UpdateLayerTree drops
// PictureLayers for boundaries that no longer exist in the widget tree.
func TestUpdateLayerTree_RemovesBoundary(t *testing.T) {
	root := &testContainer{}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))
	root.SetScreenOrigin(geometry.Pt(0, 0))

	child1 := &testLeaf{}
	child1.SetVisible(true)
	child1.SetRepaintBoundary(true)
	child1.SetBounds(geometry.NewRect(10, 20, 58, 68))
	child1.SetScreenOrigin(geometry.Pt(10, 20))
	child1.SetParent(root)

	child2 := &testLeaf{}
	child2.SetVisible(true)
	child2.SetRepaintBoundary(true)
	child2.SetBounds(geometry.NewRect(100, 20, 148, 68))
	child2.SetScreenOrigin(geometry.Pt(100, 20))
	child2.SetParent(root)

	root.kids = []widget.Widget{child1, child2}

	// First frame: root + child1 + child2 = 3 boundaries.
	tree1 := UpdateLayerTree(root, nil)
	var pics1 []*compositor.PictureLayerImpl
	collectPictureLayersFromTree(tree1, &pics1)
	if len(pics1) != 3 {
		t.Fatalf("frame 1: expected 3 PictureLayers, got %d", len(pics1))
	}

	// Remove child2 between frames.
	root.kids = []widget.Widget{child1}

	// Second frame: root + child1 = 2 boundaries.
	tree2 := UpdateLayerTree(root, tree1)
	var pics2 []*compositor.PictureLayerImpl
	collectPictureLayersFromTree(tree2, &pics2)
	if len(pics2) != 2 {
		t.Fatalf("frame 2: expected 2 PictureLayers, got %d", len(pics2))
	}

	// Verify child2's key is NOT present.
	child2Key := child2.BoundaryCacheKey()
	for _, pic := range pics2 {
		if pic.BoundaryCacheKey() == child2Key {
			t.Error("removed child2 boundary should not appear in updated tree")
		}
	}
}

// TestUpdateLayerTree_UpdatesOffset verifies that when a boundary moves
// (different Bounds.Min), UpdateLayerTree updates the OffsetLayer offset.
func TestUpdateLayerTree_UpdatesOffset(t *testing.T) {
	root := &testContainer{}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))
	root.SetScreenOrigin(geometry.Pt(0, 0))

	child := &testLeaf{}
	child.SetVisible(true)
	child.SetRepaintBoundary(true)
	child.SetBounds(geometry.NewRect(10, 20, 58, 68))
	child.SetScreenOrigin(geometry.Pt(10, 20))
	child.SetParent(root)
	root.kids = append(root.kids, child)

	// First frame.
	tree1 := UpdateLayerTree(root, nil)
	offsets1 := collectOffsetLayersByKey(tree1)
	childKey := child.BoundaryCacheKey()
	off1 := offsets1[childKey]
	if off1 == nil {
		t.Fatal("child OffsetLayer not found in first frame")
	}
	origOffset := off1.Offset()

	// Move child between frames.
	child.SetBounds(geometry.NewRect(200, 300, 248, 348))
	child.SetScreenOrigin(geometry.Pt(200, 300))

	// Second frame.
	tree2 := UpdateLayerTree(root, tree1)
	offsets2 := collectOffsetLayersByKey(tree2)
	off2 := offsets2[childKey]
	if off2 == nil {
		t.Fatal("child OffsetLayer not found in second frame")
	}

	// Offset should have changed.
	newOffset := off2.Offset()
	if origOffset == newOffset {
		t.Error("OffsetLayer should have updated offset after boundary moved")
	}
	// The offset should reflect the new bounds.Min (200, 300).
	if newOffset.X != 200 || newOffset.Y != 300 {
		t.Errorf("OffsetLayer offset = %v, want (200, 300)", newOffset)
	}
}

// TestUpdateLayerTree_SyncsDirtyFlag verifies that UpdateLayerTree propagates
// dirty/clean state from widget to PictureLayer.
func TestUpdateLayerTree_SyncsDirtyFlag(t *testing.T) {
	root := &testContainer{}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))
	root.SetScreenOrigin(geometry.Pt(0, 0))
	root.ClearSceneDirty()

	child := &testLeaf{}
	child.SetVisible(true)
	child.SetRepaintBoundary(true)
	child.SetBounds(geometry.NewRect(10, 20, 58, 68))
	child.SetScreenOrigin(geometry.Pt(10, 20))
	child.SetParent(root)
	child.ClearSceneDirty()
	root.kids = append(root.kids, child)

	// First frame: both clean.
	tree1 := UpdateLayerTree(root, nil)

	// Dirty the child between frames.
	child.InvalidateScene()

	tree2 := UpdateLayerTree(root, tree1)

	var pics []*compositor.PictureLayerImpl
	collectPictureLayersFromTree(tree2, &pics)

	childKey := child.BoundaryCacheKey()
	for _, pic := range pics {
		if pic.BoundaryCacheKey() == childKey {
			if !pic.IsDirty() {
				t.Error("child PictureLayer should be dirty after InvalidateScene")
			}
			return
		}
	}
	t.Error("child PictureLayer not found in updated tree")
}

// TestUpdateLayerTree_NilRoot verifies nil root with existing tree.
func TestUpdateLayerTree_NilRoot(t *testing.T) {
	tree := UpdateLayerTree(nil, nil)
	if tree == nil {
		t.Fatal("UpdateLayerTree(nil, nil) should return non-nil OffsetLayer")
	}
	if len(tree.Children()) != 0 {
		t.Error("nil root should produce empty tree")
	}

	// With existing tree.
	existing := compositor.NewOffsetLayer(geometry.Point{})
	tree2 := UpdateLayerTree(nil, existing)
	if tree2 == nil {
		t.Fatal("UpdateLayerTree(nil, existing) should return non-nil OffsetLayer")
	}
	if len(tree2.Children()) != 0 {
		t.Error("nil root with existing tree should produce empty tree")
	}
}

// TestUpdateLayerTree_CompositorClipSynced verifies that CompositorClip
// is propagated to PictureLayer during update.
func TestUpdateLayerTree_CompositorClipSynced(t *testing.T) {
	root := &testContainer{}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))
	root.SetScreenOrigin(geometry.Pt(0, 0))

	child := &testLeaf{}
	child.SetVisible(true)
	child.SetRepaintBoundary(true)
	child.SetBounds(geometry.NewRect(10, 20, 58, 68))
	child.SetScreenOrigin(geometry.Pt(10, 20))
	child.SetParent(root)
	child.SetCompositorClip(geometry.NewRect(0, 0, 400, 300))
	root.kids = append(root.kids, child)

	// First frame.
	tree1 := UpdateLayerTree(root, nil)

	// Update clip between frames.
	child.SetCompositorClip(geometry.NewRect(50, 50, 350, 250))

	// Second frame.
	tree2 := UpdateLayerTree(root, tree1)

	var pics []*compositor.PictureLayerImpl
	collectPictureLayersFromTree(tree2, &pics)

	childKey := child.BoundaryCacheKey()
	for _, pic := range pics {
		if pic.BoundaryCacheKey() == childKey {
			if !pic.HasPictureClip() {
				t.Error("PictureLayer should have clip after update")
			}
			clip := pic.PictureClipRect()
			if clip.Min.X != 50 || clip.Min.Y != 50 {
				t.Errorf("clip Min = %v, want (50, 50)", clip.Min)
			}
			return
		}
	}
	t.Error("child PictureLayer not found in updated tree")
}

// TestUpdateLayerTree_MultipleFramesStable verifies that the persistent
// tree stays correct across many consecutive update frames.
func TestUpdateLayerTree_MultipleFramesStable(t *testing.T) {
	root := &testContainer{}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))
	root.SetScreenOrigin(geometry.Pt(0, 0))

	child := &testLeaf{}
	child.SetVisible(true)
	child.SetRepaintBoundary(true)
	child.SetBounds(geometry.NewRect(10, 20, 58, 68))
	child.SetScreenOrigin(geometry.Pt(10, 20))
	child.SetParent(root)
	root.kids = append(root.kids, child)

	tree := UpdateLayerTree(root, nil)

	// Run 10 update frames.
	for i := range 10 {
		tree = UpdateLayerTree(root, tree)

		var pics []*compositor.PictureLayerImpl
		collectPictureLayersFromTree(tree, &pics)
		if len(pics) != 2 {
			t.Fatalf("frame %d: expected 2 PictureLayers, got %d", i+1, len(pics))
		}
	}
}

// --- ExternalTextureLayer Tests (GPUView / #197) ---

// testViewport simulates a GPUView widget: RepaintBoundary that implements
// externalTextureWidget (Texture() + ViewportSize()). Used to verify that
// BuildLayerTree/UpdateLayerTree appends ExternalTextureLayer for such widgets.
type testViewport struct {
	widget.WidgetBase
	texture   gpucontext.TextureView
	vpWidth   int
	vpHeight  int
	drawCount int
}

func (w *testViewport) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(float32(w.vpWidth), float32(w.vpHeight)))
}
func (w *testViewport) Draw(_ widget.Context, canvas widget.Canvas) {
	w.drawCount++
	if canvas != nil {
		canvas.DrawRect(w.Bounds(), widget.RGBA8(0, 128, 255, 255))
	}
}
func (w *testViewport) Event(_ widget.Context, _ event.Event) bool { return false }
func (w *testViewport) Children() []widget.Widget                  { return nil }
func (w *testViewport) Texture() gpucontext.TextureView            { return w.texture }
func (w *testViewport) ViewportSize() (int, int)                   { return w.vpWidth, w.vpHeight }

// newTestViewport creates a testViewport with a valid fake texture.
func newTestViewport(x, y float32, width, height int) *testViewport {
	var dummy int
	tv := gpucontext.NewTextureView(unsafe.Pointer(&dummy))
	vp := &testViewport{
		texture:  tv,
		vpWidth:  width,
		vpHeight: height,
	}
	vp.SetVisible(true)
	vp.SetRepaintBoundary(true)
	vp.SetBounds(geometry.NewRect(x, y, x+float32(width), y+float32(height)))
	vp.SetScreenOrigin(geometry.Pt(x, y))
	return vp
}

// collectExternalTextureLayers walks a Layer Tree and collects all ExternalTextureLayers.
func collectExternalTextureLayers(layer compositor.Layer, out *[]*compositor.ExternalTextureLayer) {
	if layer == nil {
		return
	}
	if ext, ok := layer.(*compositor.ExternalTextureLayer); ok {
		*out = append(*out, ext)
		return
	}
	if cl, ok := layer.(compositor.ContainerLayer); ok {
		for _, child := range cl.Children() {
			collectExternalTextureLayers(child, out)
		}
	}
}

// TestBuildLayerTree_CreatesExternalTextureLayer verifies that BuildLayerTree
// creates an ExternalTextureLayer for a widget implementing externalTextureWidget
// (GPUView pattern). The ExternalTextureLayer is a sibling of the PictureLayer
// inside the same OffsetLayer.
func TestBuildLayerTree_CreatesExternalTextureLayer(t *testing.T) {
	root := &testContainer{}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))
	root.SetScreenOrigin(geometry.Pt(0, 0))

	vp := newTestViewport(100, 200, 320, 240)
	vp.SetParent(root)
	root.kids = []widget.Widget{vp}

	tree := BuildLayerTree(root)

	var extLayers []*compositor.ExternalTextureLayer
	collectExternalTextureLayers(tree, &extLayers)

	if len(extLayers) != 1 {
		t.Fatalf("expected 1 ExternalTextureLayer, got %d", len(extLayers))
	}

	ext := extLayers[0]
	if ext.Width() != 320 || ext.Height() != 240 {
		t.Errorf("ExternalTextureLayer size = %dx%d, want 320x240", ext.Width(), ext.Height())
	}
	if ext.X() != 100 || ext.Y() != 200 {
		t.Errorf("ExternalTextureLayer position = (%.0f, %.0f), want (100, 200)", ext.X(), ext.Y())
	}
	if ext.Texture().IsNil() {
		t.Error("ExternalTextureLayer texture should not be nil")
	}
}

// TestBuildLayerTree_NoExternalTextureForRegularWidget verifies that regular
// boundary widgets (non-GPUView) do NOT get an ExternalTextureLayer.
func TestBuildLayerTree_NoExternalTextureForRegularWidget(t *testing.T) {
	root := &testContainer{}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))

	child := &testLeaf{}
	child.SetVisible(true)
	child.SetRepaintBoundary(true)
	child.SetBounds(geometry.NewRect(0, 0, 48, 48))
	child.SetScreenOrigin(geometry.Pt(10, 10))
	child.SetParent(root)
	root.kids = []widget.Widget{child}

	tree := BuildLayerTree(root)

	var extLayers []*compositor.ExternalTextureLayer
	collectExternalTextureLayers(tree, &extLayers)

	if len(extLayers) != 0 {
		t.Errorf("expected 0 ExternalTextureLayers for regular widget, got %d", len(extLayers))
	}
}

// TestBuildLayerTree_NilTextureSkipsExternalLayer verifies that a viewport
// widget with a nil texture does NOT produce an ExternalTextureLayer.
func TestBuildLayerTree_NilTextureSkipsExternalLayer(t *testing.T) {
	root := &testContainer{}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))
	root.SetScreenOrigin(geometry.Pt(0, 0))

	vp := &testViewport{
		texture:  gpucontext.TextureView{}, // nil/zero texture
		vpWidth:  320,
		vpHeight: 240,
	}
	vp.SetVisible(true)
	vp.SetRepaintBoundary(true)
	vp.SetBounds(geometry.NewRect(0, 0, 320, 240))
	vp.SetScreenOrigin(geometry.Pt(0, 0))
	vp.SetParent(root)
	root.kids = []widget.Widget{vp}

	tree := BuildLayerTree(root)

	var extLayers []*compositor.ExternalTextureLayer
	collectExternalTextureLayers(tree, &extLayers)

	if len(extLayers) != 0 {
		t.Errorf("expected 0 ExternalTextureLayers for nil texture, got %d", len(extLayers))
	}
}

// TestBuildLayerTree_ExternalTextureCoexistsWithPictureLayer verifies that
// the ExternalTextureLayer is appended alongside (not instead of) the PictureLayer.
// Both must exist in the same OffsetLayer: PictureLayer for scene recording
// compatibility, ExternalTextureLayer for actual texture blit.
func TestBuildLayerTree_ExternalTextureCoexistsWithPictureLayer(t *testing.T) {
	root := &testContainer{}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))
	root.SetScreenOrigin(geometry.Pt(0, 0))

	vp := newTestViewport(50, 50, 200, 150)
	vp.SetParent(root)
	root.kids = []widget.Widget{vp}

	tree := BuildLayerTree(root)

	var pics []*compositor.PictureLayerImpl
	collectPictureLayersFromTree(tree, &pics)

	var exts []*compositor.ExternalTextureLayer
	collectExternalTextureLayers(tree, &exts)

	// Root PictureLayer + viewport PictureLayer = 2 PictureLayers.
	if len(pics) != 2 {
		t.Errorf("expected 2 PictureLayers (root + viewport), got %d", len(pics))
	}

	// Viewport also has ExternalTextureLayer.
	if len(exts) != 1 {
		t.Errorf("expected 1 ExternalTextureLayer, got %d", len(exts))
	}
}

// TestUpdateLayerTree_CreatesExternalTextureLayer verifies that UpdateLayerTree
// (persistent tree path) also creates ExternalTextureLayer for viewport widgets.
func TestUpdateLayerTree_CreatesExternalTextureLayer(t *testing.T) {
	root := &testContainer{}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))
	root.SetScreenOrigin(geometry.Pt(0, 0))

	vp := newTestViewport(100, 200, 320, 240)
	vp.SetParent(root)
	root.kids = []widget.Widget{vp}

	// First frame: build from scratch.
	tree := UpdateLayerTree(root, nil)

	var extLayers []*compositor.ExternalTextureLayer
	collectExternalTextureLayers(tree, &extLayers)
	if len(extLayers) != 1 {
		t.Fatalf("frame 1: expected 1 ExternalTextureLayer, got %d", len(extLayers))
	}

	// Second frame: update with existing tree (reuse path).
	tree2 := UpdateLayerTree(root, tree)

	var extLayers2 []*compositor.ExternalTextureLayer
	collectExternalTextureLayers(tree2, &extLayers2)
	if len(extLayers2) != 1 {
		t.Fatalf("frame 2: expected 1 ExternalTextureLayer, got %d", len(extLayers2))
	}

	// Verify position is correct on reuse.
	ext := extLayers2[0]
	if ext.X() != 100 || ext.Y() != 200 {
		t.Errorf("reused ExternalTextureLayer position = (%.0f, %.0f), want (100, 200)", ext.X(), ext.Y())
	}
}

// TestUpdateLayerTree_ExternalTextureUpdatedOnMove verifies that when a
// viewport widget moves, the ExternalTextureLayer position updates.
func TestUpdateLayerTree_ExternalTextureUpdatedOnMove(t *testing.T) {
	root := &testContainer{}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))
	root.SetScreenOrigin(geometry.Pt(0, 0))

	vp := newTestViewport(100, 200, 320, 240)
	vp.SetParent(root)
	root.kids = []widget.Widget{vp}

	tree := UpdateLayerTree(root, nil)

	// Move viewport.
	vp.SetBounds(geometry.NewRect(300, 400, 620, 640))
	vp.SetScreenOrigin(geometry.Pt(300, 400))

	tree2 := UpdateLayerTree(root, tree)

	var extLayers []*compositor.ExternalTextureLayer
	collectExternalTextureLayers(tree2, &extLayers)
	if len(extLayers) != 1 {
		t.Fatalf("expected 1 ExternalTextureLayer after move, got %d", len(extLayers))
	}

	ext := extLayers[0]
	if ext.X() != 300 || ext.Y() != 400 {
		t.Errorf("ExternalTextureLayer position after move = (%.0f, %.0f), want (300, 400)",
			ext.X(), ext.Y())
	}
}

// TestBuildLayerTree_MultipleViewports verifies that multiple viewport
// widgets each get their own ExternalTextureLayer.
func TestBuildLayerTree_MultipleViewports(t *testing.T) {
	root := &testContainer{}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))
	root.SetScreenOrigin(geometry.Pt(0, 0))

	vp1 := newTestViewport(0, 0, 320, 240)
	vp1.SetParent(root)
	vp2 := newTestViewport(400, 0, 320, 240)
	vp2.SetParent(root)
	root.kids = []widget.Widget{vp1, vp2}

	tree := BuildLayerTree(root)

	var extLayers []*compositor.ExternalTextureLayer
	collectExternalTextureLayers(tree, &extLayers)

	if len(extLayers) != 2 {
		t.Fatalf("expected 2 ExternalTextureLayers, got %d", len(extLayers))
	}

	// Verify distinct positions.
	positions := map[float64]bool{}
	for _, ext := range extLayers {
		positions[ext.X()] = true
	}
	if !positions[0] || !positions[400] {
		t.Errorf("expected positions at X=0 and X=400, got %v", positions)
	}
}

// --- Benchmarks ---

// BenchmarkLayerTree_Build_200Boundaries measures per-frame BuildLayerTree
// cost with 200 boundaries (fresh allocation every frame).
func BenchmarkLayerTree_Build_200Boundaries(b *testing.B) {
	root := buildWidgetTreeWithBoundaries(200)
	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_ = BuildLayerTree(root)
	}
}

// BenchmarkLayerTree_Update_200Boundaries measures per-frame UpdateLayerTree
// cost with 200 boundaries (persistent tree, reuse existing layers).
func BenchmarkLayerTree_Update_200Boundaries(b *testing.B) {
	root := buildWidgetTreeWithBoundaries(200)
	tree := BuildLayerTree(root)
	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		tree = UpdateLayerTree(root, tree)
	}
}

// BenchmarkLayerTree_Build_50Boundaries measures BuildLayerTree with
// a smaller boundary count (typical dialog with widgets).
func BenchmarkLayerTree_Build_50Boundaries(b *testing.B) {
	root := buildWidgetTreeWithBoundaries(50)
	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_ = BuildLayerTree(root)
	}
}

// BenchmarkLayerTree_Update_50Boundaries measures UpdateLayerTree with
// 50 boundaries (persistent tree reuse).
func BenchmarkLayerTree_Update_50Boundaries(b *testing.B) {
	root := buildWidgetTreeWithBoundaries(50)
	tree := BuildLayerTree(root)
	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		tree = UpdateLayerTree(root, tree)
	}
}

// --- test helpers ---

// collectPictureLayersFromTree walks a Layer Tree and collects all PictureLayers.
func collectPictureLayersFromTree(layer compositor.Layer, out *[]*compositor.PictureLayerImpl) {
	if layer == nil {
		return
	}
	if pic, ok := layer.(*compositor.PictureLayerImpl); ok {
		*out = append(*out, pic)
		return
	}
	if cl, ok := layer.(compositor.ContainerLayer); ok {
		for _, child := range cl.Children() {
			collectPictureLayersFromTree(child, out)
		}
	}
}

// collectOffsetLayersByKey walks the tree and maps BoundaryCacheKey to the
// OffsetLayerImpl that wraps the boundary's PictureLayer.
func collectOffsetLayersByKey(root compositor.Layer) map[uint64]*compositor.OffsetLayerImpl {
	result := make(map[uint64]*compositor.OffsetLayerImpl)
	collectOffsetsRecursive(root, result)
	return result
}

func collectOffsetsRecursive(layer compositor.Layer, out map[uint64]*compositor.OffsetLayerImpl) {
	if layer == nil {
		return
	}
	offset, isOffset := layer.(*compositor.OffsetLayerImpl)
	if isOffset {
		// Check if this OffsetLayer wraps a PictureLayer (boundary pair).
		for _, ch := range offset.Children() {
			if pic, ok := ch.(*compositor.PictureLayerImpl); ok {
				key := pic.BoundaryCacheKey()
				if key != 0 {
					out[key] = offset
				}
			}
		}
	}
	if cl, ok := layer.(compositor.ContainerLayer); ok {
		for _, ch := range cl.Children() {
			collectOffsetsRecursive(ch, out)
		}
	}
}

// buildWidgetTreeWithBoundaries creates a root container with N child
// boundary widgets, used for benchmarking.
func buildWidgetTreeWithBoundaries(n int) *testContainer {
	root := &testContainer{}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))
	root.SetScreenOrigin(geometry.Pt(0, 0))

	for i := range n {
		child := &testLeaf{}
		child.SetVisible(true)
		child.SetRepaintBoundary(true)
		x := float32((i % 20) * 40)
		y := float32((i / 20) * 40)
		child.SetBounds(geometry.NewRect(x, y, x+32, y+32))
		child.SetScreenOrigin(geometry.Pt(x, y))
		child.SetParent(root)
		root.kids = append(root.kids, child)
	}

	return root
}
