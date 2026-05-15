# Render Pipeline

> How gogpu/ui renders a frame. Read this before debugging the render loop.

## Overview

gogpu/ui uses a retained-mode render pipeline inspired by Flutter, Chrome, and Qt6. The core idea: each `RepaintBoundary` widget owns an offscreen GPU texture. When a widget changes, only its boundary's texture is re-rendered. Everything else is reused from the previous frame.

The pipeline lives in `desktop/desktop.go` (`renderLoop.draw`).

## Frame Lifecycle

```
User interaction (click, hover, signal change)
  → SetNeedsRedraw(true) on widget
  → propagateDirtyUpward to nearest RepaintBoundary
  → InvalidateScene() on that boundary
  → RegisterDirtyBoundary() → flat dirty set (O(1))
  → RequestRedraw() → next frame scheduled
```

### Frame Skip (O(1))

Before doing any work, the render loop checks:

```go
if !w.HasDirtyBoundaries() && !w.NeedsRedraw() && !w.NeedsAnimationFrame() {
    return  // nothing changed — 0% GPU
}
```

This is O(1) — a flat set length check, not a tree walk. Static UI = zero GPU.

## The 10-Step Pipeline

Each frame executes these steps in order:

### Step 1: Frame Setup

```
Frame()                         // flush signals, layout, animations
BeginAcceleratorFrame()          // reset GPU frame state
BeginGPUFrame()                  // prepare gg render context
ResetFrameDamage()               // clear damage tracking
```

`Frame()` runs the signal scheduler (up to 2 re-flushes for cascading changes), layout pass if needed, and animation tick.

### Step 2: Root Invalidation

If `NeedsRedraw()` or `fullRedrawNeeded`, the root boundary's scene is invalidated. This forces the root to re-record its content.

The `SuppressDirtyCallback` mechanism prevents this from restarting the animation pumper — we're already inside the render loop.

### Step 3: Collect Dirty Regions

```
CollectDirtyRegions()            // capture dirty widget rects
prePaintDirtyRegions = DirtyRegions()
```

Done BEFORE PaintBoundaryLayers because that step clears `NeedsRedraw` flags. These rects feed the debug overlay and OS partial present.

### Step 4: Paint Boundary Layers (Flutter `flushPaint`)

```
PaintBoundaryLayersWithContext(root, nil, ctx)
```

Walks the flat dirty boundary set. Each dirty boundary re-records its `scene.Scene` display list via `SceneCanvas`. Clean boundaries are skipped entirely.

**DrawChild skip pattern:** During recording, child boundaries are SKIPPED — they have their own GPU textures. The parent scene contains only non-boundary children (text, backgrounds, dividers).

### Step 5: Paint Overlay Boundaries

```
PaintOverlayBoundaries(overlayWidgets, ctx)
```

Dropdown menus, dialogs, popovers — same boundary pipeline as main widgets. Each overlay content widget is a `RepaintBoundary`.

### Step 6: Update Layer Tree (Persistent)

```
layerTree = UpdateLayerTree(root, layerTree)
AppendOverlaysToLayerTree(layerTree, overlayWidgets, layerTree)
```

Builds/updates a persistent Layer Tree (`compositor/`). Layer types:
- **OffsetLayer** — positions boundaries in window coordinates
- **PictureLayer** — owns scene + BoundaryCacheKey + ScreenOrigin + ClipRect
- **ClipRectLayer** — viewport clipping (ScrollView)
- **OpacityLayer** — alpha blending

Persistent = layer objects reused across frames (97.9% fewer allocs for 200 boundaries). Overlays appended AFTER main tree for correct Z-order.

### Step 7: Render Boundary Textures

```
cc.SetDamageTracking(false)      // offscreen renders must not pollute surface damage
renderBoundaryTexturesFromTree(layerTree, cc)
cc.SetDamageTracking(true)
```

Walks the Layer Tree. For each `PictureLayer`:
- **Dirty boundary** → create/reuse offscreen MSAA texture → `GPUSceneRenderer.RenderScene(scene)` → GPU texture filled
- **Clean boundary** → skip (reuse previous texture, 0 GPU work)
- **Invisible boundary** (scrolled out) → skip entirely

### Step 8: Composite Textures

```
compositeTexturesFromTree(layerTree, cc, width, height)
```

Walks the Layer Tree again. Blits all boundary textures onto the surface via non-MSAA path (`DrawGPUTextureBase` for root, `DrawGPUTexture` for children). Each texture blitted at its `ScreenOrigin` with `ClipRect` scissor.

### Step 9: Overlays + Debug

```
DrawOverlayScrim()               // modal backdrop only (non-modal = no scrim)
debugOverlay.draw()              // cyan flash on dirty widgets (GOGPU_DEBUG_DIRTY=1)
```

### Step 10: Present

Two paths:

**Damage-aware (default):** When only child boundaries changed (root unchanged), uses `RenderDirectWithDamageRects` — `LoadOpLoad` + per-draw scissor. Previous swapchain content preserved. Zero pixel waste.

```
Spinner (48×48) + button (100×32) = two scissor rects
NOT one big union rect
```

**Full blit (fallback):** When root changed, overlays present, or first frame. Uses `canvas.Render` — `LoadOpClear` + full surface blit.

Ring buffer stores damage rects across N swapchain buffers. Threshold: >16 rects merges to union (GDK/Sway pattern).

## Data Flow

```
Widget state change
  → SetNeedsRedraw(true)
    → propagateDirtyUpward → nearest RepaintBoundary
      → InvalidateScene() → scene version incremented
        → RegisterDirtyBoundary() → flat set in Window
          → RequestRedraw()
            → desktop.draw():
              1. Frame() (signals, layout, animation)
              2. Root invalidation (if needed)
              3. CollectDirtyRegions
              4. PaintBoundaryLayers → re-record dirty scenes
              5. PaintOverlayBoundaries
              6. UpdateLayerTree (persistent)
              7. renderBoundaryTextures → GPU textures (MSAA)
              8. compositeTextures → blit to surface (non-MSAA)
              9. Scrim + debug overlay
             10. Present (damage-aware or full blit)
```

## Key Types

| Type | Location | Purpose |
|------|----------|---------|
| `renderLoop` | `desktop/desktop.go` | Frame state, texture cache, damage ring buffer |
| `boundaryTexEntry` | `desktop/desktop.go` | Per-boundary GPU texture + metadata |
| `OffsetLayerImpl` | `compositor/layer.go` | Layer Tree node with position |
| `PictureLayerImpl` | `compositor/layer.go` | Leaf node with scene + cache key |
| `ClipRectLayerImpl` | `compositor/layer.go` | Viewport clip (ScrollView) |
| `FontRegistry` | `internal/render/fontregistry.go` | Global font resolution |

## Debug Tools

```bash
# Cyan flash on dirty widget regions (ui level)
GOGPU_DEBUG_DIRTY=1 go run ./examples/gallery/

# Green flash on damage regions + diagnostic logging (GPU level)
GOGPU_DEBUG_DAMAGE=1 go run ./examples/gallery/

# Disable damage-aware blit (force full render every frame)
GOGPU_DAMAGE_BLIT=0 go run ./examples/gallery/
```

`GOGPU_DEBUG_DAMAGE=1` prints per-frame diagnostic log:
```
[FRAME] #42 needsRedraw=false dirtyBoundaries=1 animFrame=true fullRedraw=false
[RENDER-CHECK] frame=42 key=5 root=false size=48x48 dirty=true originValid=true
[RENDER] frame=42 key=5 root=false size=48x48 sceneVersion=42
[DAMAGE-TRACK] frame=42 source=child-boundary key=5 rect=(24,64)-(72,112)
[BLIT-PATH] frame=42 damageEnabled=true skipRoot=true hasOverlays=false damageRects=1
```

## Performance Characteristics

| Scenario | GPU Work |
|----------|----------|
| Static UI (no interaction) | 0% — frame skip |
| Hover over button | Root re-record + blit (1 boundary) |
| Spinner animating | 48×48 scissor blit at 30fps |
| Spinner scrolled offscreen | 0% — boundary culled |
| Dropdown open | Overlay boundary + scrim |
| Window resize | Full redraw (all boundaries) |

## Enterprise References

| Pattern | Our Implementation | Reference |
|---------|-------------------|-----------|
| Layer Tree | `compositor/` | Flutter `Layer`, Chrome `cc::Layer` |
| Persistent Tree | `UpdateLayerTree` | Flutter `addRetained`, Android `RenderNode` |
| flushPaint | `PaintBoundaryLayers` | Flutter `PipelineOwner.flushPaint` |
| DrawChild skip | `BoundaryRecorder` | Flutter `PaintingContext.paintChild` |
| Damage tracking | `CollectDirtyRegions` | Chrome `DamageTracker` |
| Frame skip | `HasDirtyBoundaries` O(1) | Flutter `_nodesNeedingPaint` |
| Multi-rect damage | `accumulatedDamageRects` | GDK, Sway, VK_KHR_incremental_present |

## Files

| File | What |
|------|------|
| `desktop/desktop.go` | Render loop, texture cache, damage blit |
| `app/layer_tree.go` | `UpdateLayerTree`, `PaintBoundaryLayers`, `AppendOverlays` |
| `app/window.go` | `HasDirtyBoundaries`, `RegisterDirtyBoundary`, `CollectDirtyRegions` |
| `compositor/layer.go` | Layer types (Offset, Picture, ClipRect, Opacity) |
| `widget/base.go` | `SetNeedsRedraw`, `propagateDirtyUpward` |
| `widget/boundary.go` | `InvalidateScene`, `SceneCacheVersion` |
| `widget/stamp.go` | `StampScreenOrigin`, `stampCompositorClip` |
| `internal/render/canvas.go` | Canvas (gg.Context wrapper), `DrawStyledText` |
| `internal/render/scene_canvas.go` | SceneCanvas (scene.Scene recorder) |
| `internal/dirty/collector.go` | Dirty region collection + merge |

---

*v0.1.26 — May 2026*
