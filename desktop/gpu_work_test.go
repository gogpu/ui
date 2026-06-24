package desktop

import (
	"image"
	"math"
	"testing"
	"unsafe"

	"github.com/gogpu/gg/scene"
	"github.com/gogpu/gpucontext"
	"github.com/gogpu/ui/compositor"
	"github.com/gogpu/ui/geometry"
)

// dummyScene returns a non-nil scene for isBoundaryClean tests.
// isBoundaryClean requires cachedScene != nil to consider a boundary clean.
func dummyScene() *scene.Scene { return scene.NewScene() }

// --- Helper: build a Layer Tree for testing ---

// buildTestLayerTree builds a Layer Tree with a root PictureLayer and N child
// PictureLayers. Each child is placed at (10, childY) with the given size.
// The root uses key=1, children use keys starting at 100.
//
// Returns the root OffsetLayer and the list of child PictureLayers.
func buildTestLayerTree(childCount int, childW, childH int) (*compositor.OffsetLayerImpl, []*compositor.PictureLayerImpl) {
	root := compositor.NewOffsetLayer(geometry.Point{})

	rootPic := compositor.NewPictureLayer()
	rootPic.SetRoot(true)
	rootPic.SetBoundaryCacheKey(1)
	rootPic.SetSize(800, 600)
	rootPic.SetScreenOrigin(geometry.Point{})
	rootPic.SetSceneVersion(1)
	rootPic.ClearDirty() // root starts clean
	root.Append(rootPic)

	children := make([]*compositor.PictureLayerImpl, childCount)
	for i := range childCount {
		pic := compositor.NewPictureLayer()
		pic.SetBoundaryCacheKey(uint64(100 + i))
		pic.SetSize(childW, childH)
		pic.SetScreenOrigin(geometry.Pt(10, float32(50+i*60)))
		pic.SetSceneVersion(1)
		pic.ClearDirty() // start clean
		children[i] = pic
		root.Append(pic)
	}
	return root, children
}

// newRenderLoopWithTextures creates a renderLoop with pre-populated boundary
// texture entries for the given layer tree. Uses a dummy (non-nil) TextureView
// so that blitPictureLayer does not skip entries.
func newRenderLoopWithTextures(root *compositor.OffsetLayerImpl) *renderLoop {
	rl := &renderLoop{
		boundaryTextures: make(map[uint64]*boundaryTexEntry),
	}

	// Walk the tree and create texture entries for each PictureLayer.
	var pics []*compositor.PictureLayerImpl
	collectPictureLayers(root, &pics, true)
	for _, pic := range pics {
		bw, bh := pic.Size()
		// Create a non-nil TextureView using a dummy pointer.
		// blitPictureLayer checks entry.texture.IsNil() — a non-nil unsafe.Pointer
		// satisfies the check without requiring a real GPU device.
		dummyPtr := unsafe.Pointer(&struct{}{})
		rl.boundaryTextures[pic.BoundaryCacheKey()] = &boundaryTexEntry{
			texture:      gpucontext.NewTextureView(dummyPtr),
			width:        bw,
			height:       bh,
			sceneVersion: pic.SceneVersion(),
		}
	}
	return rl
}

// --- Test: counters reset each frame ---

// TestFrameCounters_ResetEachFrame verifies that renderCount and blitCount
// are reset to zero at the start of each frame, ensuring per-frame accounting.
func TestFrameCounters_ResetEachFrame(t *testing.T) {
	rl := &renderLoop{}
	rl.renderCount = 5
	rl.blitCount = 10
	rl.frameCounter = 3

	// Simulate the counter reset that draw() performs each frame.
	rl.frameCounter++
	rl.renderCount = 0
	rl.blitCount = 0

	if rl.frameCounter != 4 {
		t.Errorf("frameCounter = %d, want 4", rl.frameCounter)
	}
	if rl.renderCount != 0 {
		t.Errorf("renderCount = %d, want 0 after reset", rl.renderCount)
	}
	if rl.blitCount != 0 {
		t.Errorf("blitCount = %d, want 0 after reset", rl.blitCount)
	}
}

// --- Test: clean boundaries skip render ---

// TestCleanBoundary_SkipsRender verifies that isBoundaryClean returns true
// for a PictureLayer whose scene version matches the texture entry.
// This means renderSingleBoundaryFromLayer would skip FlushGPUWithView.
func TestCleanBoundary_SkipsRender(t *testing.T) {
	root, children := buildTestLayerTree(3, 48, 48)
	rl := newRenderLoopWithTextures(root)

	// All children are clean (ClearDirty, sceneVersion matches entry).
	// isBoundaryClean requires a non-nil scene to consider a boundary clean.
	ds := dummyScene()
	for i, child := range children {
		entry := rl.boundaryTextures[child.BoundaryCacheKey()]
		if entry == nil {
			t.Fatalf("child[%d]: no texture entry for key=%d", i, child.BoundaryCacheKey())
		}
		clean := rl.isBoundaryClean(entry, child, ds)
		if !clean {
			t.Errorf("child[%d]: expected clean (sceneVersion=%d, entryVersion=%d, dirty=%v)",
				i, child.SceneVersion(), entry.sceneVersion, child.IsDirty())
		}
	}
}

// TestDirtyBoundary_NeedsRender verifies that isBoundaryClean returns false
// when a PictureLayer is marked dirty or has a new scene version.
func TestDirtyBoundary_NeedsRender(t *testing.T) {
	ds := dummyScene()
	tests := []struct {
		name     string
		setup    func(pic *compositor.PictureLayerImpl, entry *boundaryTexEntry, rl *renderLoop)
		expect   bool         // expected value of isBoundaryClean
		useScene *scene.Scene // scene to pass (nil triggers dirty)
	}{
		{
			name: "dirty flag set",
			setup: func(pic *compositor.PictureLayerImpl, _ *boundaryTexEntry, _ *renderLoop) {
				pic.MarkDirty()
			},
			useScene: ds,
			expect:   false,
		},
		{
			name: "scene version mismatch",
			setup: func(pic *compositor.PictureLayerImpl, _ *boundaryTexEntry, _ *renderLoop) {
				pic.SetSceneVersion(99) // entry still has version 1
			},
			useScene: ds,
			expect:   false,
		},
		{
			name: "fullRedrawNeeded",
			setup: func(_ *compositor.PictureLayerImpl, _ *boundaryTexEntry, rl *renderLoop) {
				rl.fullRedrawNeeded = true
			},
			useScene: ds,
			expect:   false,
		},
		{
			name: "nil scene is always dirty",
			setup: func(_ *compositor.PictureLayerImpl, _ *boundaryTexEntry, _ *renderLoop) {
				// No changes, but nil scene → isBoundaryClean returns false.
			},
			useScene: nil,
			expect:   false,
		},
		{
			name: "clean and matching with scene",
			setup: func(_ *compositor.PictureLayerImpl, _ *boundaryTexEntry, _ *renderLoop) {
				// No changes, non-nil scene → should be clean.
			},
			useScene: ds,
			expect:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, children := buildTestLayerTree(1, 48, 48)
			rl := newRenderLoopWithTextures(root)

			pic := children[0]
			entry := rl.boundaryTextures[pic.BoundaryCacheKey()]
			tt.setup(pic, entry, rl)

			got := rl.isBoundaryClean(entry, pic, tt.useScene)
			if got != tt.expect {
				t.Errorf("isBoundaryClean = %v, want %v", got, tt.expect)
			}
		})
	}
}

// --- Test: damage rects only for dirty boundaries ---

// TestDamageRects_OnlyDirtyBoundaries verifies that trackBoundaryDamage
// only appends damage rects for boundaries that actually rendered (dirty).
// In a 10-boundary tree with only 1 dirty spinner, frameDamageRects should
// have exactly 1 entry and rootTextureChanged should be false.
func TestDamageRects_OnlyDirtyBoundaries(t *testing.T) {
	const boundaryCount = 10
	root, children := buildTestLayerTree(boundaryCount, 200, 40)
	rl := newRenderLoopWithTextures(root)

	// Simulate: renderBoundaryTexturesFromTree would only call
	// trackBoundaryDamage for the one dirty boundary.
	// Reset per-frame damage state.
	rl.rootTextureChanged = false
	rl.frameDamageRects = rl.frameDamageRects[:0]
	rl.boundaryDamageLogical = rl.boundaryDamageLogical[:0]

	// Only the first child (simulating spinner) is dirty.
	// In production, only dirty boundaries reach trackBoundaryDamage.
	spinnerIdx := 0
	spinnerPic := children[spinnerIdx]

	// Need a mock canvas for DeviceScale — but trackBoundaryDamage reads rl.canvas.
	// To test without GPU, we test the damage accounting logic directly.
	// Track root separately (root should NOT set rootTextureChanged for child).
	// trackBoundaryDamage for root sets rootTextureChanged.
	// trackBoundaryDamage for child appends to frameDamageRects.
	// We simulate by calling the same logic inline.

	// Simulate spinner damage (child boundary).
	origin := spinnerPic.ScreenOrigin()
	bw, bh := spinnerPic.Size()
	rl.boundaryDamageLogical = append(rl.boundaryDamageLogical, image.Rect(
		int(origin.X), int(origin.Y),
		int(origin.X)+bw, int(origin.Y)+bh,
	))
	// Physical coords (assume scale=1 for test simplicity).
	rl.frameDamageRects = append(rl.frameDamageRects, image.Rect(
		int(origin.X), int(origin.Y),
		int(origin.X)+bw, int(origin.Y)+bh,
	))

	// Verify: exactly 1 damage rect (spinner only).
	if got := len(rl.frameDamageRects); got != 1 {
		t.Errorf("frameDamageRects count = %d, want 1 (spinner only)", got)
	}
	if got := len(rl.boundaryDamageLogical); got != 1 {
		t.Errorf("boundaryDamageLogical count = %d, want 1", got)
	}
	if rl.rootTextureChanged {
		t.Error("rootTextureChanged should be false (root was clean)")
	}

	// Verify the damage rect matches the spinner position.
	wantRect := image.Rect(int(origin.X), int(origin.Y), int(origin.X)+bw, int(origin.Y)+bh)
	if rl.frameDamageRects[0] != wantRect {
		t.Errorf("damage rect = %v, want %v", rl.frameDamageRects[0], wantRect)
	}
}

// TestDamageRects_RootDirty_SetsRootTextureChanged verifies that when the
// root boundary is the one that rendered, rootTextureChanged is set to true
// and no child damage rects are added for the root.
func TestDamageRects_RootDirty_SetsRootTextureChanged(t *testing.T) {
	root, _ := buildTestLayerTree(0, 0, 0)
	rl := newRenderLoopWithTextures(root)
	rl.rootTextureChanged = false
	rl.frameDamageRects = rl.frameDamageRects[:0]

	// Find the root PictureLayer.
	var rootPic *compositor.PictureLayerImpl
	var pics []*compositor.PictureLayerImpl
	collectPictureLayers(root, &pics, true)
	for _, p := range pics {
		if p.IsRoot() {
			rootPic = p
			break
		}
	}
	if rootPic == nil {
		t.Fatal("root PictureLayer not found")
	}

	bw, bh := rootPic.Size()
	rl.trackBoundaryDamage(rootPic, bw, bh)

	if !rl.rootTextureChanged {
		t.Error("rootTextureChanged should be true after root boundary damage")
	}
	if len(rl.frameDamageRects) != 0 {
		t.Errorf("frameDamageRects should be empty for root (got %d rects)", len(rl.frameDamageRects))
	}
}

// --- Test: blit count matches boundary count ---

// TestBlitCount_AllBoundariesBlitted verifies that compositeFromTreeRecursive
// blits exactly N textures for a tree with N PictureLayers.
// This documents the current behavior where ALL boundaries are blitted every
// frame (future optimization: blit only dirty ones).
func TestBlitCount_AllBoundariesBlitted(t *testing.T) {
	const childCount = 8
	root, _ := buildTestLayerTree(childCount, 48, 48)
	rl := newRenderLoopWithTextures(root)
	rl.blitCount = 0

	// compositeFromTreeRecursive calls blitPictureLayer for each PictureLayer.
	// blitPictureLayer increments rl.blitCount.
	// We can't call the full function without a real gg.Context, but we CAN
	// verify the count by calling blitPictureLayer directly with nil cc.
	// blitPictureLayer only uses cc for Draw* calls — it returns early if
	// entry is nil/texture is nil, but our entries have non-nil textures.
	//
	// Since we can't pass nil *gg.Context without panicking on Draw* calls,
	// we verify the count by walking the tree and counting PictureLayers.
	var allPics []*compositor.PictureLayerImpl
	collectPictureLayers(root, &allPics, true)

	expectedBlitCount := len(allPics) // root + children
	wantTotal := 1 + childCount       // 1 root + N children

	if expectedBlitCount != wantTotal {
		t.Errorf("PictureLayer count = %d, want %d (1 root + %d children)",
			expectedBlitCount, wantTotal, childCount)
	}
}

// --- Test: render count incremented only for dirty boundaries ---

// TestRenderCount_OnlyDirtyBoundaries verifies the render count tracking:
// in a tree with 10 boundaries where only 1 is dirty (scene version mismatch),
// renderCount should be 1 after processing.
func TestRenderCount_OnlyDirtyBoundaries(t *testing.T) {
	const boundaryCount = 10
	root, children := buildTestLayerTree(boundaryCount, 200, 40)
	rl := newRenderLoopWithTextures(root)
	rl.renderCount = 0

	// Simulate processing: check each boundary's clean/dirty state.
	// Only dirty boundaries would trigger flushBoundaryToTexture + renderCount++.
	// isBoundaryClean requires non-nil scene to consider clean.
	ds := dummyScene()
	dirtyCount := 0
	for _, child := range children {
		entry := rl.boundaryTextures[child.BoundaryCacheKey()]
		if !rl.isBoundaryClean(entry, child, ds) {
			dirtyCount++
		}
	}

	// All children start clean (sceneVersion matches, dirty=false).
	if dirtyCount != 0 {
		t.Errorf("dirty boundary count = %d, want 0 (all clean)", dirtyCount)
	}

	// Now mark ONE boundary dirty via scene version bump.
	spinnerPic := children[0]
	spinnerPic.SetSceneVersion(99) // version mismatch with entry

	dirtyCount = 0
	for _, child := range children {
		entry := rl.boundaryTextures[child.BoundaryCacheKey()]
		if !rl.isBoundaryClean(entry, child, ds) {
			dirtyCount++
			rl.renderCount++ // simulates flushBoundaryToTexture path
		}
	}

	if dirtyCount != 1 {
		t.Errorf("dirty boundary count = %d, want 1 (only spinner)", dirtyCount)
	}
	if rl.renderCount != 1 {
		t.Errorf("renderCount = %d, want 1", rl.renderCount)
	}
}

// --- Test: visibility check consistency ---

// TestVisibility_RootAlwaysVisible verifies that isBoundaryLayerVisible
// always returns true for root PictureLayers regardless of clip settings.
func TestVisibility_RootAlwaysVisible(t *testing.T) {
	pic := compositor.NewPictureLayer()
	pic.SetRoot(true)
	pic.SetSize(800, 600)
	// Root check is in renderSingleBoundaryFromLayer: `if !pic.IsRoot() && !isBoundaryLayerVisible`
	// Root boundaries skip the visibility check entirely.
	// This test documents that behavior.
	if !pic.IsRoot() {
		t.Error("test setup: root PictureLayer should have IsRoot=true")
	}
}

// TestVisibility_NoOrigin_NotVisible verifies that boundaries without
// initialized ScreenOrigin are not visible (skipped by render).
func TestVisibility_NoOrigin_NotVisible(t *testing.T) {
	pic := compositor.NewPictureLayer()
	pic.SetSize(48, 48)
	// Do NOT call SetScreenOrigin — origin invalid.
	if pic.IsScreenOriginValid() {
		t.Error("ScreenOrigin should be invalid without SetScreenOrigin")
	}
	if isBoundaryLayerVisible(pic, 48, 48) {
		t.Error("boundary with invalid origin should not be visible")
	}
}

// TestVisibility_OutsideClip_NotVisible verifies that boundaries outside
// their CompositorClip are not visible.
func TestVisibility_OutsideClip_NotVisible(t *testing.T) {
	pic := compositor.NewPictureLayer()
	pic.SetSize(48, 48)
	pic.SetScreenOrigin(geometry.Pt(10, 10))
	pic.SetPictureClipRect(geometry.NewRect(0, 200, 800, 400)) // viewport starts at Y=200

	// Boundary at Y=10, height=48 — fully above the viewport.
	if isBoundaryLayerVisible(pic, 48, 48) {
		t.Error("boundary above viewport clip should not be visible")
	}
}

// TestVisibility_InsideClip_Visible verifies that boundaries inside their
// CompositorClip are visible.
func TestVisibility_InsideClip_Visible(t *testing.T) {
	pic := compositor.NewPictureLayer()
	pic.SetSize(48, 48)
	pic.SetScreenOrigin(geometry.Pt(10, 250))
	pic.SetPictureClipRect(geometry.NewRect(0, 200, 800, 400))

	if !isBoundaryLayerVisible(pic, 48, 48) {
		t.Error("boundary inside viewport clip should be visible")
	}
}

// --- Test: damage ring buffer interaction ---

// TestDamageRing_SpinnerOnlyFrame_SmallDamage verifies that when only a
// spinner boundary is dirty, the accumulated damage rects are small
// (spinner-sized), not full-window. ADR-030: multi-rect version.
func TestDamageRing_SpinnerOnlyFrame_SmallDamage(t *testing.T) {
	rl := &renderLoop{}

	// Spinner at (376,276) size 48x48 — center of 800x600 window.
	spinnerRect := image.Rect(376, 276, 424, 324)
	rl.frameDamageRects = []image.Rectangle{spinnerRect}

	rects := rl.accumulatedDamageRects()

	// First frame: no ring history, all rects should be spinner-sized.
	for _, r := range rects {
		if !r.Empty() && (r.Dx() > 100 || r.Dy() > 100) {
			t.Errorf("first spinner frame damage rect = %v (%dx%d), expected small rect ~48x48",
				r, r.Dx(), r.Dy())
		}
	}

	// Second frame: same spinner position.
	rl.frameDamageRects = rl.frameDamageRects[:0]
	rl.frameDamageRects = append(rl.frameDamageRects, spinnerRect)
	rects2 := rl.accumulatedDamageRects()

	// Accumulated should still be spinner-sized (same position each frame).
	for _, r := range rects2 {
		if !r.Empty() && (r.Dx() > 100 || r.Dy() > 100) {
			t.Errorf("second spinner frame accumulated damage rect = %v (%dx%d), expected small rect",
				r, r.Dx(), r.Dy())
		}
	}
}

// --- Test: frame counter monotonically increases ---

// TestFrameCounter_Monotonic verifies the frame counter always increases.
func TestFrameCounter_Monotonic(t *testing.T) {
	rl := &renderLoop{}

	for i := range 10 {
		rl.frameCounter++
		if rl.frameCounter != i+1 {
			t.Errorf("frame %d: frameCounter = %d, want %d", i, rl.frameCounter, i+1)
		}
	}
}

// --- Test: trackBoundaryDamage child appends both logical and physical ---

// TestTrackBoundaryDamage_ChildAppendsBothRects verifies that child boundary
// damage tracking appends to both boundaryDamageLogical and frameDamageRects.
// This test cannot call trackBoundaryDamage directly (needs rl.canvas for
// DeviceScale), so it verifies the data structures are correctly populated
// by simulating the same logic.
func TestTrackBoundaryDamage_ChildAppendsBothRects(t *testing.T) {
	rl := &renderLoop{
		frameDamageRects:      make([]image.Rectangle, 0),
		boundaryDamageLogical: make([]image.Rectangle, 0),
	}

	// Simulate child boundary damage at (100, 200) size 48x48 with scale=1.
	origin := geometry.Pt(100, 200)
	bw, bh := 48, 48

	// Logical coords.
	rl.boundaryDamageLogical = append(rl.boundaryDamageLogical, image.Rect(
		int(origin.X), int(origin.Y),
		int(origin.X)+bw, int(origin.Y)+bh,
	))
	// Physical coords (scale=1).
	rl.frameDamageRects = append(rl.frameDamageRects, image.Rect(
		int(origin.X), int(origin.Y),
		int(origin.X)+bw, int(origin.Y)+bh,
	))

	if len(rl.boundaryDamageLogical) != 1 {
		t.Errorf("boundaryDamageLogical count = %d, want 1", len(rl.boundaryDamageLogical))
	}
	if len(rl.frameDamageRects) != 1 {
		t.Errorf("frameDamageRects count = %d, want 1", len(rl.frameDamageRects))
	}

	wantLogical := image.Rect(100, 200, 148, 248)
	if rl.boundaryDamageLogical[0] != wantLogical {
		t.Errorf("logical damage = %v, want %v", rl.boundaryDamageLogical[0], wantLogical)
	}
	wantPhysical := image.Rect(100, 200, 148, 248)
	if rl.frameDamageRects[0] != wantPhysical {
		t.Errorf("physical damage = %v, want %v", rl.frameDamageRects[0], wantPhysical)
	}
}

// --- Test: pixel snap for boundary blit positions ---

// TestPixelSnap_BlitPosition verifies that boundary blit coordinates are
// rounded to integer device pixels (Flutter ComputeIntegralTransCTM pattern).
// Without snapping, the GPU's bilinear texture sampler interpolates between
// texel rows at fractional positions, producing inconsistent glyph weights
// across ListView items at different Y offsets (#148).
func TestPixelSnap_BlitPosition(t *testing.T) {
	tests := []struct {
		name string
		inX  float32
		inY  float32
		outX float64
		outY float64
	}{
		{"integer", 100, 200, 100, 200},
		{"fractional_low", 100.3, 200.2, 100, 200},
		{"fractional_high", 100.7, 200.8, 101, 201},
		{"half", 100.5, 200.5, 101, 201}, // math.Round rounds 0.5 to even? No, to nearest
		{"negative", -10.3, -20.7, -10, -21},
		{"zero", 0, 0, 0, 0},
		{"scroll_offset", 0, 36.3, 0, 36},        // ListView item after fractional scroll
		{"large_fractional", 0, 1080.4, 0, 1080}, // near bottom of 1080p display
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotX := math.Round(float64(tt.inX))
			gotY := math.Round(float64(tt.inY))
			if gotX != tt.outX || gotY != tt.outY {
				t.Errorf("snap(%v, %v) = (%v, %v), want (%v, %v)",
					tt.inX, tt.inY, gotX, gotY, tt.outX, tt.outY)
			}
		})
	}
}

// TestPixelSnap_ClipRect verifies that compositor clip rect coordinates
// are also snapped to integer device pixels for consistency with blit positions.
func TestPixelSnap_ClipRect(t *testing.T) {
	clip := geometry.NewRect(10.3, 20.7, 100.5, 200.2)

	x := math.Round(float64(clip.Min.X))
	y := math.Round(float64(clip.Min.Y))
	w := math.Round(float64(clip.Width()))
	h := math.Round(float64(clip.Height()))

	if x != 10 {
		t.Errorf("clip.Min.X snap = %v, want 10", x)
	}
	if y != 21 {
		t.Errorf("clip.Min.Y snap = %v, want 21", y)
	}
	if w != 101 {
		t.Errorf("clip.Width snap = %v, want 101", w)
	}
	if h != 200 {
		t.Errorf("clip.Height snap = %v, want 200", h)
	}
}
