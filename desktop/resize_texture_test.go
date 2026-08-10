package desktop

import (
	"testing"

	"github.com/gogpu/gg"
	"github.com/gogpu/ui/app"
	"github.com/gogpu/ui/compositor"
)

func TestSurfaceResizePending_ForcesFullSurfaceFrameWithoutDirtyingBoundaries(t *testing.T) {
	win := app.New().Window()
	win.ClearAfterPaint()
	win.ClearDirtyBoundaries()
	win.ClearAnimationFrame()
	if win.NeedsRedraw() || win.HasDirtyBoundaries() || win.NeedsAnimationFrame() {
		t.Fatal("test window is not clean")
	}

	rl := &renderLoop{surfaceResizePending: true}
	if !rl.needsFrame(win) {
		t.Error("a clean window skipped the frame required by a canvas resize")
	}
	if !rl.requiresFullSurfaceRender() {
		t.Error("a canvas resize did not require full surface composition")
	}

	entry := &boundaryTexEntry{sceneVersion: 3}
	pic := compositor.NewPictureLayer()
	pic.SetSceneVersion(3)
	pic.ClearDirty()
	if !rl.isBoundaryClean(entry, pic, dummyScene()) {
		t.Error("surface-only resize invalidated an unchanged boundary texture")
	}
}

func TestEnsureBoundaryTexture_DoesNotDirtyCleanSibling(t *testing.T) {
	ctx := gg.NewContext(800, 600)
	t.Cleanup(func() {
		if err := ctx.Close(); err != nil {
			t.Errorf("close context: %v", err)
		}
	})
	rl := &renderLoop{
		boundaryTextures: map[uint64]*boundaryTexEntry{
			2: {width: 50, height: 50, sceneVersion: 3},
		},
	}

	// Allocating the resized root texture must invalidate only that entry,
	// not every boundary consulted later in the same tree walk.
	rootEntry := rl.ensureBoundaryTexture(1, 800, 600, ctx)
	if rl.fullRedrawNeeded {
		t.Fatal("allocating one boundary texture forced a full-tree redraw")
	}

	rootPic := compositor.NewPictureLayer()
	rootPic.SetSceneVersion(1)
	rootPic.ClearDirty()
	if rl.isBoundaryClean(rootEntry, rootPic, dummyScene()) {
		t.Error("fresh root texture was considered clean before its first render")
	}

	siblingPic := compositor.NewPictureLayer()
	siblingPic.SetSceneVersion(3)
	siblingPic.ClearDirty()
	if !rl.isBoundaryClean(rl.boundaryTextures[2], siblingPic, dummyScene()) {
		t.Error("unchanged sibling texture became dirty when another boundary was allocated")
	}
}

func TestEnsureBoundaryTexture_ReleasesOnlySizeChangedEntry(t *testing.T) {
	ctx := gg.NewContext(800, 600)
	t.Cleanup(func() {
		if err := ctx.Close(); err != nil {
			t.Errorf("close context: %v", err)
		}
	})
	releases := 0
	existing := &boundaryTexEntry{
		width:   50,
		height:  50,
		release: func() { releases++ },
	}
	rl := &renderLoop{
		boundaryTextures: map[uint64]*boundaryTexEntry{
			1: existing,
			2: {width: 25, height: 25, release: func() { releases++ }},
		},
	}

	if got := rl.ensureBoundaryTexture(1, 50, 50, ctx); got != existing {
		t.Error("same-size boundary replaced its retained texture entry")
	}
	if releases != 0 {
		t.Fatalf("same-size boundary released %d textures, want 0", releases)
	}

	if got := rl.ensureBoundaryTexture(1, 75, 50, ctx); got == existing {
		t.Error("size-changed boundary retained its stale texture entry")
	}
	if releases != 1 {
		t.Fatalf("size change released %d textures, want only the changed entry", releases)
	}
	if rl.boundaryTextures[2] == nil {
		t.Error("unrelated boundary texture entry was discarded")
	}
}
