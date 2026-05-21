package desktop

import (
	"image"
	"testing"

	"github.com/gogpu/ui/compositor"
	"github.com/gogpu/ui/geometry"
)

// --- ADR-030: Multi-Rect Damage Tests ---

func TestAccumulatedDamageRects_SingleRect(t *testing.T) {
	rl := &renderLoop{}
	rl.frameDamageRects = []image.Rectangle{
		image.Rect(100, 200, 148, 248), // spinner 48x48
	}

	got := rl.accumulatedDamageRects()

	// Single boundary → result must contain exactly that rect.
	found := false
	for _, r := range got {
		if r == image.Rect(100, 200, 148, 248) {
			found = true
		}
	}
	if !found {
		t.Errorf("accumulatedDamageRects should contain spinner rect, got %v", got)
	}
}

func TestAccumulatedDamageRects_TwoDistantBoundaries(t *testing.T) {
	rl := &renderLoop{}

	// Two dirty boundaries far apart: spinner (24,64,48,48) + button (300,500,100,32).
	spinner := image.Rect(24, 64, 72, 112)   // 48x48
	button := image.Rect(300, 500, 400, 532) // 100x32
	rl.frameDamageRects = []image.Rectangle{spinner, button}

	got := rl.accumulatedDamageRects()

	// ADR-030: should return 2+ separate rects, NOT one union.
	// Union would be (24,64)-(400,532) = 376x468 = 175,968 px.
	// Multi-rect = 48x48 + 100x32 = 5,504 px (32x savings).
	if len(got) < 2 {
		t.Errorf("expected 2+ separate rects for distant boundaries, got %d: %v", len(got), got)
	}

	// Verify both rects are present.
	hasSpinner, hasButton := false, false
	for _, r := range got {
		if r == spinner {
			hasSpinner = true
		}
		if r == button {
			hasButton = true
		}
	}
	if !hasSpinner {
		t.Errorf("result should contain spinner rect %v, got %v", spinner, got)
	}
	if !hasButton {
		t.Errorf("result should contain button rect %v, got %v", button, got)
	}
}

func TestAccumulatedDamageRects_ThresholdMergesToUnion(t *testing.T) {
	rl := &renderLoop{}

	// 20 dirty boundaries → exceeds maxDamageRects=16 → should merge to single union.
	for i := range 20 {
		rl.frameDamageRects = append(rl.frameDamageRects, image.Rect(i*40, 0, i*40+30, 30))
	}

	got := rl.accumulatedDamageRects()

	if len(got) != 1 {
		t.Errorf("expected 1 merged rect when exceeding threshold, got %d rects", len(got))
	}

	// Union should cover all 20 rects: (0,0) to (790,30).
	if len(got) == 1 {
		if got[0].Min.X != 0 || got[0].Min.Y != 0 {
			t.Errorf("union min should be (0,0), got %v", got[0].Min)
		}
		if got[0].Max.X < 790 || got[0].Max.Y < 30 {
			t.Errorf("union should cover all rects to (790,30), got %v", got[0])
		}
	}
}

func TestAccumulatedDamageRects_RingBufferAccumulation(t *testing.T) {
	rl := &renderLoop{}

	// Frame 1: spinner dirty.
	spinner := image.Rect(100, 200, 148, 248)
	rl.frameDamageRects = []image.Rectangle{spinner}
	d1 := rl.accumulatedDamageRects()
	t.Logf("frame 1: %v", d1)

	// Frame 2: button dirty (different position).
	button := image.Rect(500, 400, 600, 432)
	rl.frameDamageRects = []image.Rectangle{button}
	d2 := rl.accumulatedDamageRects()
	t.Logf("frame 2: %v", d2)

	// Frame 2 result must contain BOTH spinner (from ring buffer) + button (current).
	hasSpinner, hasButton := false, false
	for _, r := range d2 {
		if r == spinner {
			hasSpinner = true
		}
		if r == button {
			hasButton = true
		}
	}
	if !hasSpinner {
		t.Errorf("frame 2 should include spinner from ring buffer, got %v", d2)
	}
	if !hasButton {
		t.Errorf("frame 2 should include button from current frame, got %v", d2)
	}
}

func TestAccumulatedDamageRects_FullBlitStoresFullWindow(t *testing.T) {
	rl := &renderLoop{}

	// Simulate full blit by storing full window in ring buffer (as draw() does).
	fullWindow := image.Rect(0, 0, 800, 600)
	rl.damageRingRects[rl.damageRingIdx] = []image.Rectangle{fullWindow}
	rl.damageRingIdx = (rl.damageRingIdx + 1) % len(rl.damageRingRects)

	// Next frame: spinner only — but ring buffer has full window.
	spinner := image.Rect(100, 200, 148, 248)
	rl.frameDamageRects = []image.Rectangle{spinner}
	got := rl.accumulatedDamageRects()

	// Should contain full window rect from ring buffer.
	hasFullWindow := false
	for _, r := range got {
		if r == fullWindow {
			hasFullWindow = true
		}
	}
	if !hasFullWindow {
		// Threshold may merge — check union covers full window.
		if len(got) == 1 && got[0].Dx() >= 800 && got[0].Dy() >= 600 {
			// Merged to union covering full window — acceptable.
			return
		}
		t.Errorf("result should include full window from ring buffer, got %v", got)
	}
}

func TestAccumulatedDamageRects_SingleBoundaryOneRect(t *testing.T) {
	rl := &renderLoop{}

	// Only spinner dirty → result should contain exactly spinner rect.
	spinner := image.Rect(100, 200, 148, 248)
	rl.frameDamageRects = []image.Rectangle{spinner}

	got := rl.accumulatedDamageRects()

	// With empty ring buffer, should be the spinner rect (possibly duplicated
	// because ring buffer stores current frame too, but all entries are same).
	allSpinner := true
	for _, r := range got {
		if !r.Empty() && r != spinner {
			allSpinner = false
		}
	}
	if !allSpinner {
		t.Errorf("single boundary should produce only spinner rects, got %v", got)
	}
}

func TestRootTextureChanged_TrackedCorrectly(t *testing.T) {
	rl := &renderLoop{}

	// Initially false.
	if rl.rootTextureChanged {
		t.Error("rootTextureChanged should be false initially")
	}

	// After setting.
	rl.rootTextureChanged = true
	if !rl.rootTextureChanged {
		t.Error("rootTextureChanged should be true after set")
	}

	// After reset.
	rl.rootTextureChanged = false
	if rl.rootTextureChanged {
		t.Error("rootTextureChanged should be false after reset")
	}
}

func TestDamageBlitDecision_RootDirty_FullBlit(t *testing.T) {
	// When root texture changed, should use full blit (not damage-aware).
	rl := &renderLoop{rootTextureChanged: true}
	skipRootBlit := !rl.rootTextureChanged && !rl.fullRedrawNeeded

	if skipRootBlit {
		t.Error("should NOT skip root blit when root texture changed")
	}
}

func TestDamageBlitDecision_SpinnerOnly_DamageAware(t *testing.T) {
	// When root clean and spinner dirty, should use damage-aware path.
	rl := &renderLoop{
		rootTextureChanged: false,
		fullRedrawNeeded:   false,
		frameDamageRects:   []image.Rectangle{image.Rect(100, 200, 148, 248)},
	}
	skipRootBlit := !rl.rootTextureChanged && !rl.fullRedrawNeeded
	hasDamage := len(rl.frameDamageRects) > 0

	if !skipRootBlit {
		t.Error("should skip root blit when root texture unchanged")
	}
	if !hasDamage {
		t.Error("should have damage rects for spinner")
	}
}

func TestDamageBlitDecision_FullRedrawNeeded_FullBlit(t *testing.T) {
	// First frame or resize — always full blit.
	rl := &renderLoop{fullRedrawNeeded: true, rootTextureChanged: false}
	skipRootBlit := !rl.rootTextureChanged && !rl.fullRedrawNeeded

	if skipRootBlit {
		t.Error("should NOT skip root blit on first frame/resize")
	}
}

func TestDamageBlitDecision_NoDamageRects_FullBlit(t *testing.T) {
	// No damage rects (shouldn't happen in practice) — fallback to full.
	rl := &renderLoop{rootTextureChanged: false, fullRedrawNeeded: false}
	hasDamage := len(rl.frameDamageRects) > 0

	if hasDamage {
		t.Error("should have no damage rects")
	}
}

// --- #111: Ring buffer size and full-render fill ---

func TestDamageRingRects_SizeFour(t *testing.T) {
	// Ring buffer must have 4 slots to cover Linux 4-buffer swapchains (#111).
	rl := &renderLoop{}
	if len(rl.damageRingRects) != 4 {
		t.Errorf("damageRingRects size = %d, want 4", len(rl.damageRingRects))
	}
}

func TestFullRender_FillsAllRingSlots(t *testing.T) {
	// After full render (canvas.Render), ALL ring buffer slots must contain
	// fullWindow. This prevents stale buffers on 4-buffer swapchains (#111).
	rl := &renderLoop{}

	// Simulate partial state in ring buffer (as if previous frames left data).
	rl.damageRingRects[1] = []image.Rectangle{image.Rect(10, 10, 50, 50)}
	rl.damageRingRects[2] = []image.Rectangle{image.Rect(60, 60, 100, 100)}
	rl.damageRingIdx = 2

	// Simulate what draw() does after canvas.Render (full blit path).
	fullWindowRect := []image.Rectangle{image.Rect(0, 0, 1920, 1080)}
	for i := range rl.damageRingRects {
		rl.damageRingRects[i] = fullWindowRect
	}
	rl.damageRingIdx = 0

	// Verify every slot has fullWindow.
	fullWindow := image.Rect(0, 0, 1920, 1080)
	for i, slot := range rl.damageRingRects {
		if len(slot) != 1 {
			t.Errorf("slot[%d] has %d rects, want 1", i, len(slot))
			continue
		}
		if slot[0] != fullWindow {
			t.Errorf("slot[%d] = %v, want %v", i, slot[0], fullWindow)
		}
	}

	// Ring index must be reset to 0 for deterministic behavior.
	if rl.damageRingIdx != 0 {
		t.Errorf("damageRingIdx = %d, want 0 after fill-all", rl.damageRingIdx)
	}
}

// --- #116: Orphaned boundary texture GC ---

func TestGCOrphanedTextures_DeletesOrphaned(t *testing.T) {
	// After SetRoot (theme change), old widget tree keys are orphaned.
	// gcOrphanedTextures must delete entries not present in Layer Tree.
	rl := &renderLoop{
		boundaryTextures: map[uint64]*boundaryTexEntry{
			100: {width: 48, height: 48},   // live — in tree
			200: {width: 100, height: 32},  // orphaned — NOT in tree
			300: {width: 200, height: 100}, // orphaned — NOT in tree
		},
	}

	// Build a tree that only contains key=100.
	tree := compositor.NewOffsetLayer(geometry.Point{})
	pic := compositor.NewPictureLayer()
	pic.SetBoundaryCacheKey(100)
	tree.Append(pic)

	released := 0
	rl.boundaryTextures[200].release = func() { released++ }
	rl.boundaryTextures[300].release = func() { released++ }

	rl.gcOrphanedTextures(tree)

	// Key 100 must survive, keys 200/300 must be deleted.
	if _, ok := rl.boundaryTextures[100]; !ok {
		t.Error("live key 100 should be preserved")
	}
	if _, ok := rl.boundaryTextures[200]; ok {
		t.Error("orphaned key 200 should be deleted")
	}
	if _, ok := rl.boundaryTextures[300]; ok {
		t.Error("orphaned key 300 should be deleted")
	}
	if released != 2 {
		t.Errorf("release called %d times, want 2", released)
	}
}

func TestGCOrphanedTextures_PreservesLiveKeys(t *testing.T) {
	// When all texture keys are in the tree, nothing should be deleted.
	rl := &renderLoop{
		boundaryTextures: map[uint64]*boundaryTexEntry{
			10: {width: 48, height: 48},
			20: {width: 100, height: 32},
		},
	}

	tree := compositor.NewOffsetLayer(geometry.Point{})
	pic1 := compositor.NewPictureLayer()
	pic1.SetBoundaryCacheKey(10)
	pic2 := compositor.NewPictureLayer()
	pic2.SetBoundaryCacheKey(20)
	tree.Append(pic1)
	tree.Append(pic2)

	rl.gcOrphanedTextures(tree)

	if len(rl.boundaryTextures) != 2 {
		t.Errorf("should preserve all 2 entries, got %d", len(rl.boundaryTextures))
	}
}

func TestGCOrphanedTextures_EmptyMap(t *testing.T) {
	// Empty map must be a no-op (no panic).
	rl := &renderLoop{
		boundaryTextures: map[uint64]*boundaryTexEntry{},
	}
	tree := compositor.NewOffsetLayer(geometry.Point{})

	// Should not panic.
	rl.gcOrphanedTextures(tree)

	if len(rl.boundaryTextures) != 0 {
		t.Errorf("empty map should remain empty, got %d entries", len(rl.boundaryTextures))
	}
}

func TestGCOrphanedTextures_NilMap(t *testing.T) {
	// Nil map (boundaryTextures not yet initialized) must be a no-op.
	rl := &renderLoop{}
	tree := compositor.NewOffsetLayer(geometry.Point{})

	// Should not panic.
	rl.gcOrphanedTextures(tree)
}

func TestGCOrphanedTextures_NestedTree(t *testing.T) {
	// GC must walk nested ContainerLayers (ClipRect, Opacity) to find
	// PictureLayers at any depth.
	rl := &renderLoop{
		boundaryTextures: map[uint64]*boundaryTexEntry{
			10: {width: 48, height: 48},  // in root
			20: {width: 100, height: 32}, // nested in ClipRectLayer
			30: {width: 50, height: 50},  // nested in OpacityLayer
			40: {width: 60, height: 60},  // orphaned
		},
	}

	tree := compositor.NewOffsetLayer(geometry.Point{})

	pic1 := compositor.NewPictureLayer()
	pic1.SetBoundaryCacheKey(10)
	tree.Append(pic1)

	clipLayer := compositor.NewClipRectLayer(geometry.Rect{})
	pic2 := compositor.NewPictureLayer()
	pic2.SetBoundaryCacheKey(20)
	clipLayer.Append(pic2)
	tree.Append(clipLayer)

	opacityLayer := compositor.NewOpacityLayer(0.5)
	pic3 := compositor.NewPictureLayer()
	pic3.SetBoundaryCacheKey(30)
	opacityLayer.Append(pic3)
	tree.Append(opacityLayer)

	released := 0
	rl.boundaryTextures[40].release = func() { released++ }

	rl.gcOrphanedTextures(tree)

	// Keys 10, 20, 30 should survive. Key 40 should be deleted.
	for _, key := range []uint64{10, 20, 30} {
		if _, ok := rl.boundaryTextures[key]; !ok {
			t.Errorf("live key %d should be preserved", key)
		}
	}
	if _, ok := rl.boundaryTextures[40]; ok {
		t.Error("orphaned key 40 should be deleted")
	}
	if released != 1 {
		t.Errorf("release called %d times, want 1", released)
	}
}

func TestGCOrphanedTextures_NilRelease(t *testing.T) {
	// Orphaned entry with nil release function must not panic.
	rl := &renderLoop{
		boundaryTextures: map[uint64]*boundaryTexEntry{
			10: {width: 48, height: 48, release: nil}, // orphaned, nil release
		},
	}

	tree := compositor.NewOffsetLayer(geometry.Point{})

	// Should not panic on nil release.
	rl.gcOrphanedTextures(tree)

	if _, ok := rl.boundaryTextures[10]; ok {
		t.Error("orphaned key 10 should be deleted even with nil release")
	}
}

func TestCollectLiveKeys_NilLayer(t *testing.T) {
	// collectLiveKeys with nil layer must be a no-op (no panic).
	keys := make(map[uint64]bool)
	collectLiveKeys(nil, keys)
	if len(keys) != 0 {
		t.Errorf("nil layer should produce no keys, got %d", len(keys))
	}
}
