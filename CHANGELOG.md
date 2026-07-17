# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.47] — 2026-07-16

### Changed

- **deps:** gg v0.50.6 → v0.50.7, gogpu v0.44.8 → v0.44.9, wgpu v0.30.21 → v0.30.22
  - **wgpu v0.30.22:** Metal MSAA storage mode fix — Intel integrated GPUs (Iris Plus) no longer crash on texture creation ([wgpu#271](https://github.com/gogpu/wgpu/issues/271)).
  - **gg v0.50.7:** Variable font gvar fix under transforms ([gg#405](https://github.com/gogpu/gg/issues/405)).

## [0.1.46] — 2026-07-16

### Fixed

- **ListView scene leak on scroll** ([#173](https://github.com/gogpu/ui/issues/173)) — evicted decorators now properly unmount (`UnmountTree`) and release cached `scene.Scene`. Previously, scrolling accumulated orphaned scene objects on the heap.
- **ListView row content never mounted** ([#174](https://github.com/gogpu/ui/issues/174)) — widgets created by `Content[C].Render()` now receive `MountTree` after creation, so signal bindings via `BindToScheduler` activate correctly. Previously, reactive updates in list item content were silently dead.
- **GridView cell content never mounted** ([#181](https://github.com/gogpu/ui/issues/181)) — same lifecycle fix as ListView applied to GridView cells. `UnmountTree` on eviction, `MountTree` after `Content[C].Render()`.
- **Overlay lifecycle bypass** ([#171](https://github.com/gogpu/ui/issues/171)) — `PushOverlay` now calls `MountTree` on the overlay container, `PopOverlay`/`RemoveOverlay` call `UnmountTree`. Previously, signal bindings in overlay content (dropdowns, dialogs) never registered, and cleanup never ran on dismiss — leaking goroutines and heap allocations.
- **Window.Close teardown** ([#175](https://github.com/gogpu/ui/issues/175)) — added `Window.Close()` method that stops the animation pumper goroutine, unmounts all overlay trees, and unmounts the root widget tree. Wired into `desktop.Run` OnClose callback. Previously, the animation pumper goroutine outlived the window.
- **linechart goroutine-safety** ([#182](https://github.com/gogpu/ui/issues/182)) — `PushValue`, `AddSeries`, `ClearSeries` now route invalidation through the scheduler (`MarkDirty`) instead of direct `SetNeedsRedraw`, eliminating data races when called from background goroutines. Falls back to `SetNeedsRedraw` when unmounted.

### Tests

- **28 regression tests** across 5 new test files covering all lifecycle fixes:
  - `core/listview/lifecycle_test.go` — mount/unmount on scroll, eviction, clear, re-scroll
  - `core/gridview/lifecycle_test.go` — cell mount/unmount, update, clear, invalidate
  - `app/overlay_lifecycle_test.go` — push/pop lifecycle, signal bindings, dismiss callback
  - `app/window_close_test.go` — close idempotency, unmount root + overlays, stop pumper
  - `core/linechart/safety_test.go` — concurrent PushValue (10 goroutines × 100 values), scheduler routing

## [0.1.45] — 2026-07-16

### Changed

- **deps:** gg v0.50.4 → v0.50.6, gogpu v0.44.1 → v0.44.8, wgpu v0.30.10 → v0.30.21, signals v0.1.0 → v0.1.1
  - **gg v0.50.6:** Unified backend-agnostic draw queue (ADR-051) + three-tier clip architecture (ADR-052). All rendering commands (shapes, text, GPU textures) now flow through a single dispatch pipeline regardless of backend. Software backend rendering is now architecturally correct — offscreen boundary textures, Layer Tree compositing, and damage-aware blit work on CPU adapters.
  - **gogpu v0.44.8:** Multi-backend auto-selection improvements, Wayland fractional scale, pixelPresented frame lifecycle.
  - **wgpu v0.30.21:** Software backend correctness fixes (BGRA swizzle, MSAA rejection, WriteTexture, blit optimization), format-aware copy commands.
  - **signals v0.1.1:** Minor improvements.

### Known Issues

- **Software backend performance:** The unified pipeline ensures correctness on CPU adapters, but performance is significantly slower than GPU backends. The SPIR-V interpreter processes shaders instruction-by-instruction on the CPU. Community contributions for optimization are welcome — see [Optimization Roadmap](#software-backend-optimization) below.

### Software Backend Optimization Roadmap

The software backend (`GOGPU_GRAPHICS_API=software`) now renders correctly through the same Layer Tree compositor pipeline as GPU backends. Performance optimization is the next priority:

1. **gg direct CPU rasterization** ([ADR-053](https://github.com/gogpu/gg)): gg already has fast native Go CPU rasterizers inspired by tiny-skia and Vello — AnalyticFiller (scanline AA), SparseStrips (4×4 tiles), TileCompute (16×16 Vello 9-stage). These are orders of magnitude faster than the SPIR-V interpreter. Smart dispatch routing: shapes/text rendered directly by gg's CPU rasterizers, only GPU-specific operations (texture compositing) go through the software HAL. This is the highest-impact optimization with minimal code changes.
2. **naga Go+SIMD backend** ([NAGA-FEAT-004](https://github.com/gogpu/naga)): Generate native Go code from WGSL shaders with `goexperiment.simd` (AVX-512/NEON) vectorization. Replaces SPIR-V interpreter entirely for shader execution. Expected 100x+ speedup. Reference: GoMLX PackGEMM achieved 14x speedup with pure Go SIMD on MatMul.
3. **SPIR-V interpreter SIMD** ([FEAT-SW-008](https://github.com/gogpu/wgpu)): `goexperiment.simd` Float32x4 for vec4 ops in the existing interpreter — estimated 2-4x speedup with ~500 LOC. Interim solution before naga Go backend.
4. **Multi-threaded CPU dispatch** ([ADR-053](https://github.com/gogpu/gg)): Parallel CPU dispatch for independent boundary textures. Each boundary = isolated pixmap → trivially parallel.

Contributions and profiling reports are welcome — see [issue #158](https://github.com/gogpu/ui/issues/158) for the software backend tracker.

## [0.1.44] — 2026-07-09

### Changed

- **deps:** gg v0.50.3 → v0.50.4, x/image v0.43.0 → v0.44.0, x/text v0.39.0 → v0.40.0
  - **gg v0.50.4:** Composite hardening (nil-safe `WithTransform`, `BeginClip`/`EndClip` balanced check, GlyphRun empty guard).

### Documentation

- **ARCHITECTURE.md:** comprehensive update — 26 widgets (added badge, chip, stripe, titlebar), 4 design systems with 70 painters (added DevTools 24), removed Engine references (CACHE-030), per-widget layout caching (ADR-032), updated all dependency versions.

## [0.1.43] — 2026-07-08

### Changed

- **deps:** gg v0.50.2 → v0.50.3, gogpu v0.43.4 → v0.44.1, wgpu v0.30.9 → v0.30.10
  - **gogpu v0.44.1:** Multi-backend auto-selection (old GPUs get DX12/GLES instead of software), macOS live resize fix.
  - **gg v0.50.3:** TT hinter improvements.

## [0.1.42] — 2026-07-05

### Added

- **Layout cache activation** (ADR-032 Phase 2b, [#160](https://github.com/gogpu/ui/pull/160), @TimLai666) — convert 32 parent→child Layout calls to `widget.LayoutChild` across 23 files, activating per-widget layout caching. Layout cost: O(total nodes) → O(affected subtree). Removes `MarkLayoutCleanRecursive` shim. Adds `InvalidateLayoutTree` for downward cache propagation on signal fires. `IsLayoutVerifying` sentinel for debug verifier (Flutter `debugCheckingIntrinsics` pattern).
- **Animation tick before layout** (ADR-032 GAP-3) — `AnimationTicker` interface + `tickAnimationsInTree` walk before layout pass. Collapsible and Transition animations now tick BEFORE layout (Flutter `handleBeginFrame` → `handleDrawFrame` pattern). Layout is a pure function of constraints + widget state, required for RelayoutBoundary (Phase 5).

### Removed

- **`internal/layout/Engine`** (ADR-032 CACHE-030) — centralized layout engine with 0 production usage. Per-widget layout caching is now fully handled by `widget.LayoutChild` on `WidgetBase`. -769 lines of dead code.

### Changed

- **deps:** gg v0.50.1 → v0.50.2, gogpu v0.43.1 → v0.43.4, wgpu v0.30.8 → v0.30.9, goffi v0.5.5 → v0.5.6

## [0.1.41] — 2026-07-01

### Fixed

- **Collapsible animation glitch on Wayland** ([#152](https://github.com/gogpu/ui/issues/152), [#157](https://github.com/gogpu/ui/pull/157)) — force final redraw after animation completes + full-window `wl_surface.damage_buffer` on root repaint. Confirmed by @porjo (Intel Iris Plus ICL GT2, Fedora 44, GNOME/Mutter).

### Changed

- **Event-driven rendering by default** (ADR-035) — removed `WithContinuousRender(false)` boilerplate from all examples. gogpu v0.43.0+ defaults to event-driven (winit 0.29 pattern).
- **deps:** gg v0.49.2 → v0.50.1, gogpu v0.42.11 → v0.43.1, wgpu v0.30.7 → v0.30.8

## [0.1.40] — 2026-06-29

### Changed

- **Event-driven rendering by default** (ADR-035) — removed `WithContinuousRender(false)` from all 6 examples and documentation. No longer needed since gogpu v0.43.0 defaults to event-driven rendering, following the winit 0.29 pattern. Enterprise research: 8/8 UI frameworks (Flutter, Qt, GTK4, SwiftUI, Compose, Iced, Gio, Chromium) enforce event-driven at entry point.
- **deps:** gg v0.49.2 → v0.50.1, gogpu v0.42.11 → v0.43.1, wgpu v0.30.7 → v0.30.8

## [0.1.39] — 2026-06-28

### Changed

- **deps:** gg v0.49.0 → v0.49.2, gogpu v0.42.8 → v0.42.11, wgpu v0.30.5 → v0.30.7, gputypes v0.5.0 → v0.5.1
  - **gg v0.49.2:** Software Vulkan (llvmpipe/SwiftShader/WARP) now fully supported for offscreen boundary textures (ADR-046). Removes `softwareMode` blanket disable — software adapters treated as full GPU implementations per enterprise pattern (Skia Graphite, Flutter Impeller, wgpu). MSAA runtime fallback 4x→1x. Error logging in `CreateOffscreenTexture` instead of silent nil return.
  - **wgpu v0.30.7:** Software backend format-aware copy + GLES LockOSThread fix.
  - **gputypes v0.5.1:** `BlockCopySize()` for format-aware buffer sizing.

## [0.1.38] — 2026-06-28

### Added

- **Layout invalidation wiring** ([#151](https://github.com/gogpu/ui/pull/151), @TimLai666) — ADR-032 Phase 2a: `BindToSchedulerLayout` + `BindToSchedulerLayoutFunc` for layout-affecting signals, `MarkNeedsLayout()` in 9 event handlers, `MarkLayoutCleanRecursive` post-layout shim. 11 signal bindings converted (button/text, checkbox/label, chip/label, badge/count, collapsible/expanded, treeview/root, listview/itemCount, gridview/itemCount+columns, datatable/rowCount, splitview/ratio). Invalidation infrastructure ready — cache activation (LayoutChild) in Phase 2b.

## [0.1.37] — 2026-06-27

### Added

- **Painter/Widget behavior separation** (ADR-034) — enterprise-level refactor moving ALL behavioral logic (cursor position, selection, text measurement, animation easing, data mapping) from theme painters to core widgets. Painters are now draw-only. Enables community theme authoring.
- **LayoutMetrics optional interface** — 7 widgets (button, checkbox, radio, slider, chip, dialog, textfield) define LayoutMetrics allowing theme painters to control spatial metrics (height, padding, fontSize, radius). 35 compile-time checked implementations across 5 painters.
- **ThemeBundle interface** (`theme.Bundle`) — packages all painters for complete theme installation. Community themes implement this to provide a full design system.
- **Gallery: 8 themes** — Fluent and Cupertino added to theme switcher alongside M3 (4 colors) + DevTools (dark/light).
- **internal/textmetrics** — shared MeasureText-based text measurement replacing per-painter charWidthRatio estimates.

### Fixed

- **DevTools cursor wrong position** ([#126](https://github.com/gogpu/ui/issues/126)) — cursor position was computed differently per theme (MeasureText vs charWidthRatio). Now all themes use widget-computed CursorRect via MeasureText.
- **Cursor height** — added 2px caretHeightOffset (Flutter `_kCaretHeightOffset` pattern) for cleaner cursor appearance.

### Changed (breaking, v0.x)

- **TextField PaintState** — expanded with pre-computed fields (DisplayText, ContentRect, CursorRect, SelectionRect, ShowCursor, ShowSelection, FontSize, ColorScheme). Pointer receiver (`*PaintState`).
- **titlebar Painter** — `DrawBackground`/`DrawControlButton` renamed to `PaintBackground`/`PaintControlButton`.
- **linechart Painter** — `PaintChart(canvas, bounds, state)` → `PaintChart(canvas, state)` with Bounds in PaintState.
- **progressbar PaintState** — `ProgressBarColorScheme` field renamed to `ColorScheme`.

### Dependencies

- gg v0.48.16 → v0.49.0
- gogpu v0.42.6 → v0.42.8
- wgpu v0.30.3 → v0.30.5

### Removed

- 4 duplicate `maskText` functions (one per theme painter — now single in core)
- 4 duplicate `contentRect` functions (now computed by widget)
- `charWidthRatio` constants from all theme textfield painters
- `easeInOut` duplicate from M3 progress painter

## [0.1.36] — 2026-06-25

### Added

- **Layout cache infrastructure** ([#143](https://github.com/gogpu/ui/pull/143), @TimLai666) — ADR-032 Phase 1: per-node constraint cache on WidgetBase, `widget.LayoutChild()` helper (symmetric with DrawChild), `MarkNeedsLayout()` with idempotency guard and upward propagation, `GOGPU_DEBUG_LAYOUT=1` debug verifier. Purely additive — mechanism is inert until containers adopt LayoutChild in Phase 2.
- **Gallery: "Show Dialog" button** — dialog section now has an interactive trigger instead of empty inline widget. Click opens modal overlay with Cancel/Confirm actions.

### Fixed

- **Radio group stale selection pixels** ([#145](https://github.com/gogpu/ui/issues/145)) — selecting a new radio button now invalidates both the newly-selected and previously-selected items. Previously only the clicked item was invalidated, causing the old "selected dot" to persist on damage-aware compositors (Wayland, Vulkan `VK_KHR_incremental_present`, DX12 `FLIP_SEQUENTIAL`). Also expanded invalidation rect to cover focus ring area (drawn 2px beyond item bounds).
- **Inconsistent glyph weights in ListView** ([#148](https://github.com/gogpu/ui/issues/148)) — RepaintBoundary GPU textures were blitted at fractional pixel positions, causing the GPU's bilinear texture sampler to interpolate between texel rows differently per item. Snapped compositor blit coordinates to integer device pixels (Flutter `ComputeIntegralTransCTM` pattern). Also snapped clip rect coordinates for consistency.
- **Collapsible content bleeding through after collapse** ([#147](https://github.com/gogpu/ui/issues/147)) — ListView items and spinner remained visible after collapsing a section. Two fixes: (1) `renderSingleBoundaryFromLayer` now updates `entry.clipRect` for invisible boundaries (stale clip prevention), (2) `compositeFromTreeRecursive` checks boundary visibility before blitting (defense-in-depth). Also added zero-area clip skip in `blitPictureLayer`.
- **Spinner arc artifact** — 1px blue line at bottom of spinner caused by rounding mismatch between damage rect (`int()` truncation) and blit position (`math.Round()`). Aligned damage tracking to use `math.Round()` matching the blit path.

### Dependencies

- gg v0.48.13 → v0.48.16
- gogpu v0.42.1 → v0.42.6
- wgpu v0.30.2 → v0.30.3

## [0.1.35] — 2026-06-21

### Added

- **Badge widget** ([#138](https://github.com/gogpu/ui/pull/138), @TimLai666) — notification badge displaying a count or status dot. Count mode (pill shape, "99+" overflow), dot mode (minimal indicator). Signal bindings with dedup, pluggable Painter, DefaultPainter (M3 error color). 99.3% coverage.
- **Chip widget** ([#139](https://github.com/gogpu/ui/pull/139), @TimLai666) — compact action/filter chip (M3 spec). Action mode (click handler), filter mode (toggleable selection with two-way signal write-back). M3 state layers (hover 0.08, press 0.12), keyboard activation (Enter/Space), focus ring. FontSize in PaintState for design system painters. 99.0% coverage.
- **Badge + Chip design system painters** — all 4 themes: M3 (Error/OnError, Outline/SecondaryContainer), DevTools (Red7/White, BorderStrong/ControlFill), Fluent (Accent/OnAccent, StrokeDefault/AccentLight), Cupertino (SystemRed/White, Separator/Accent). 86 tests.

### Performance

- **Layout pool** ([#140](https://github.com/gogpu/ui/pull/140), @TimLai666) — `sync.Pool` for per-Compute temporary slices in flex/grid/stack layouts. Flex and stack: 0 allocs/op. Grid: 2 allocs/op (config-aliased slices). ~12% faster layout compute.
- **ListView per-item decorator** — hover/selection background painted inside per-item boundary (Flutter/Android/Qt pattern). Hover change dirties only the affected item, not the entire ListView. Previously root re-recorded on every hover.
- **ListView scroll recycling** — incremental cache update reuses decorators for overlapping items during scroll. Only edge items (entering viewport) are created. RecyclerView/Flutter SliverList pattern.

### Fixed

- **Debug overlay fade** — cyan (GOGPU_DEBUG_DIRTY) and green (GOGPU_DEBUG_DAMAGE) flash-and-fade overlays now auto-fade after 400ms without requiring mouse movement. Frame skip gate was not checking overlay animation state.
- **ListView root re-record on hover** — removed legacy `invalidateItemRect` which set `window.needsRedraw=true` on every hover change, forcing root boundary re-record and full compositor blit.

### Dependencies

- gg v0.48.11 → v0.48.13
- gogpu v0.42.0 → v0.42.1
- wgpu v0.30.1 → v0.30.2
- golang.org/x/image v0.40.0 → v0.43.0
- golang.org/x/text v0.37.0 → v0.38.0 (indirect)
- go-webgpu/goffi v0.5.3 → v0.5.5 (indirect)

## [0.1.34] — 2026-06-16

### Fixed

- **Hello example gray ListView items on macOS Retina** — ListView container background was `#FAFAFA` (250,250,250), visually gray on P3 Retina displays. Changed to `#FFFFFF` (255,255,255) to match card background. Invisible on Windows sRGB, noticeable on macOS P3.

## [0.1.33] — 2026-06-16

### Fixed

- **HiDPI/Retina 2× oversized rendering** ([#129](https://github.com/gogpu/ui/issues/129), [#133](https://github.com/gogpu/ui/pull/133), @TimLai666) — per-boundary offscreen GPU textures were allocated in logical pixels while gg.Context renders at DeviceScale. On Retina (scale=2.0), textures were half-size: scene clipped to top-left quarter, then upscaled. Fix: allocate at physical pixels (`logical × DeviceScale`), new `scaleToPhysical` helper. Confirmed on macOS Retina by @lkmavi. Validated against 6 enterprise frameworks (Flutter, Chrome, Skia, Qt6, Android, Vello) — all allocate offscreen textures in physical pixels.

## [0.1.32] — 2026-06-16

### Dependencies

- gg v0.48.10 → v0.48.11 (fix #369 @TimLai666: thin strokes solid fill on GPU compute)

## [0.1.31] — 2026-06-15

### Dependencies

- gg v0.48.9 → v0.48.10 (BUG-BACKDROP-001: Vello backdrop prefix sum boundary fix, wgpu opaque handles)
- gogpu v0.41.14 → v0.42.0 (GPU struct tokens — zero unsafe.Pointer, gpucontext interfaces → opaque structs)
- gpucontext v0.19.0 → v0.21.0 (Device/Queue/Adapter/Surface/Instance → opaque struct tokens)
- wgpu v0.29.15 → v0.30.1 (opaque handles, DeviceFromHandle/DeviceToHandle type-safe helpers)

## [0.1.30] — 2026-06-15

### Changed

- **ROADMAP.md** — rewritten with future-oriented vision (Phase 5-9 timeline, ecosystem integration, design philosophy, community projects)
- **README.md** — updated metrics (198K+ LOC, 7300+ tests, 1.1M+ ecosystem)
- **ARCHITECTURE.md** — updated dependency versions, date

### Dependencies

- gg v0.48.8 → v0.48.9
- gogpu v0.41.12 → v0.41.14 (HiDPI research @TimLai666 #306)
- wgpu v0.29.14 → v0.29.15
- naga v0.17.14 → v0.17.15

## [0.1.29] — 2026-05-22

### Fixed

- **Outlined button border invisible** — `DrawRoundRect(transparent)` before `StrokeRoundRect` caused stroke to be lost in GPU SDF pipeline (gg BUG-SDF-001: zero-alpha MSAA coverage interference). Removed dead transparent fill from all 4 painters (M3, DevTools, Fluent, Cupertino). Root cause fixed upstream in gg v0.48.3.
- **Disabled Outlined button border invisible** — disabled state returned `ColorTransparent` as stroke color. Changed to `DisabledFg` for visible gray border.
- **Outlined button missing hover overlay** — M3 spec requires 8% primary tint on hover/press. Added conditional fill for Outlined variant (matching existing TextOnly pattern).
- **Checkbox left-side clipping** ([#117](https://github.com/gogpu/ui/issues/117), @celer) — stroke extended 1px beyond widget bounds in all 5 painters (Default, M3, DevTools, Fluent, Cupertino). Inset box by `borderStrokeWidth/2`.
- **Progressbar flicker** ([#112](https://github.com/gogpu/ui/issues/112), @celer) — signal-driven redraws fired on every notification even when value unchanged. Added `BindToSchedulerFunc` with deduplication predicate + `lastDrawnValue` cache.
- **Collapsible content leak** ([#122](https://github.com/gogpu/ui/issues/122), @kolkov) — collapsed content boundaries not culled from compositor. Now stamps empty `CompositorClip` via `DrawChild` when progress=0.
- **Scroll redraw artifacts on Linux** ([#111](https://github.com/gogpu/ui/issues/111), @celer) — damage ring buffer `[3]` insufficient for 4-buffer swapchains. Increased to `[4]` and fill ALL slots after full render.
- **ListView flash on theme change** ([#116](https://github.com/gogpu/ui/issues/116), @celer) — orphaned boundary textures after `SetRoot` never released. Added `gcOrphanedTextures` — walks Layer Tree, deletes entries not in live set.
- **M4 MacBook HiDPI rendering** ([#123](https://github.com/gogpu/ui/issues/123), @nzanghi-tst) — HiDPI damage rects not scaled for Retina displays. Fixed via gg v0.47.4+ deps bump.
- **macOS LCD text rendering** — LCD subpixel layout now detected from `PlatformProvider.SubpixelLayout()` via `App.PlatformProvider()` getter (ADR-024). macOS correctly gets `SubpixelNone` (grayscale AA since Mojave), Windows/Linux get `SubpixelRGB` (ClearType).
- **Memory leak on SetRoot** — cached scenes not released during `UnmountTree`; `hoveredWidget`/`capturedWidget` pinned old widget tree. Added `ClearCachedScene` in unmount + nil refs in `SetRoot`.

### Added

- **PointerCapturer** (ADR-031, [#110](https://github.com/gogpu/ui/issues/110), @celer) — widget-level mouse capture for drag operations. Slider thumb drag now works when mouse exits widget bounds. Type-assertion extension pattern (Win32 SetCapture / Qt grabMouse). Auto-releases on MouseRelease.
- **`App.PlatformProvider()`** — public getter for platform provider, enables LCD detection from desktop render loop.
- **`state.BindToSchedulerFunc`** — generic signal binding with custom predicate for deduplication. Reusable by any widget.
- **40+ new tests** across checkbox, progressbar, collapsible, desktop, slider, widget, app, theme packages.

### Dependencies

- gg v0.46.11 → v0.48.3 (text scissor fix gg#335, zero-alpha SDF skip BUG-SDF-001, stroke expander, pixel-perfect mode, HiDPI damage scaling)
- gogpu v0.35.0 → v0.39.0 (PlatformProvider delegation ADR-024, app lifecycle ADR-026, macOS system menu, custom menus, key repeat timerfd, Windows ESC, keysym-to-Unicode, AltGr/Level3)
- gpucontext v0.18.0 → v0.19.0
- wgpu v0.28.1 → v0.28.7 (macOS software blit, GLES hidden window, lint cleanup)

## [0.1.28] — 2026-05-16

### Dependencies

- gg v0.47.0 → v0.47.1 (text batch coalescing perf + HiDPI warning — ADR-031, gg#322, gg#324)
- gogpu v0.35.0 → v0.36.0 (unified XKB text input, AltGr/Level3 international layouts — gogpu#233, ADR-029)

## [0.1.27] — 2026-05-16

### Dependencies

- gg v0.46.11 → v0.47.0 (pixel-perfect mode `SetAntiAlias` — ADR-030, gg#319, gg#320)

## [0.1.26] — 2026-05-15

### Dependencies

- gogpu v0.34.8 → v0.35.0 (Browser/WASM platform support, X11 XKB constant fix — gogpu#70, gogpu#227, @unxed)
- wgpu v0.27.5 → v0.28.1 (Browser WebGPU backend + API stubs)

## [0.1.25] — 2026-05-14

### Dependencies

- gogpu v0.34.6 → v0.34.8 (Wayland keyboard layout via xkbcommon, X11 layout switching fix — gogpu#227, @paulie-g)

## [0.1.24] — 2026-05-14

### Fixed

- **Collapsible ghost pixels during animation** ([#101](https://github.com/gogpu/ui/issues/101) Thread B, @AnyCPU) — during collapse animation the clip area shrinks each frame, but the boundary's GPU texture retained stale pixels from the previous (larger) clip. `progressAdapter.Set()` now calls `InvalidateScene()` to force boundary re-recording. 1 line + 44 LOC test.
- **stampCompositorClip degenerate rects** ([#101](https://github.com/gogpu/ui/issues/101) Thread B') — zero-area or negative-area clip intersections (e.g., widget fully scrolled out of viewport) produced positioned-but-empty rects that defeated downstream `IsEmpty()` culling, allowing stale texture blits. Now normalizes to explicit zero rect. 6 lines + 99 LOC test.
- **Gallery theme dropdown resets on theme change** ([#101](https://github.com/gogpu/ui/issues/101) Thread G) — `buildGallery()` hardcoded `dropdown.Selected(0)`, losing user's theme selection when tree was rebuilt. Now tracks `galleryState.themeIdx`. 3 lines.

### Dependencies

- gg v0.46.9 → v0.46.11 (SVG HiDPI scale fix, GPU stroke EvenOdd fill rule, nil texture readback guard — [#101](https://github.com/gogpu/ui/issues/101) Threads C, F)
- wgpu v0.27.3 → v0.27.5 (flaky Windows CI fix, NULL handle guard in TransitionTextures — [#101](https://github.com/gogpu/ui/issues/101))
- gogpu v0.34.3 → v0.34.6 (macOS PUA filter, Linux EventClose, deferred SetHitTestCallback frameless drag fix — [#101](https://github.com/gogpu/ui/issues/101) Threads A, H)
- goffi v0.5.0 → v0.5.1 (amd64 struct arg passing, XMM0:XMM1 return, CGO_ENABLED=1 — [#101](https://github.com/gogpu/ui/issues/101) Thread A)
- golang.org/x/image v0.39.0 → v0.40.0
- golang.org/x/text v0.36.0 → v0.37.0

## [0.1.23] — 2026-05-13

### Added

- **Custom font loading pipeline** (TASK-UI-CJK-001, gg#304) — plugins can now load custom fonts via `ctx.Assets.LoadFont("name", data)` and they are used for rendering. Follows Flutter/Qt6/Iced universal pattern: global `FontRegistry` singleton with CSS weight matching and Inter fallback.
- **`FontRegistry`** (`internal/render/fontregistry.go`) — process-global, thread-safe (RWMutex) font registry. Pre-registers embedded Inter. Caches `*text.FontSource` by (family, weight, style). CSS weight matching via `theme/font.Registry`.
- **`StyledTextDrawer`** optional interface (`widget/canvas.go`) — `DrawStyledText(text, bounds, TextStyle)` + `MeasureStyledText(text, TextStyle)`. Implemented by both Canvas and SceneCanvas. Uses type assertion pattern consistent with `ArcStroker`, `SVGFiller`, `DeviceScaler`.
- **`TextWidget.FontFamily()`** builder method (`primitives/text.go`) — routes to `StyledTextDrawer` when custom font family set, falls back to regular `DrawText` with Inter.
- **Plugin → Registry wiring** — `MemoryAssetLoader.LoadFont()` auto-registers fonts in `GlobalFontRegistry()`. `NewDefaultPluginContext()` creates real `MemoryAssetLoader` instead of noop.
- **47 new tests** across fontregistry, canvas, scene_canvas, plugin, primitives, uitest packages.

### Fixed

- **gg v0.46.9** — fix Mac Retina rendering (gg#308, @sverrehu). `MarkDirty()` used logical pixel dimensions for texture upload region — on Retina (scale=2.0), only 1/4 of the pixmap was uploaded to the GPU texture. Regression from gg v0.45.4. Includes 3 HiDPI regression tests.

### Dependencies

- gg v0.46.8 → v0.46.9

## [0.1.21] — 2026-05-12

### Changed

- **gg v0.46.8** — fix CJK `IsCJK` propagation through scene/shaper paths. `ShapedGlyph.IsCJK` field (ADR-027) was never populated — CJK improvements (script-aware hinting, exact-size rasterization, Tier 6 routing) were silently bypassed through scene and UI compositor paths. Fixed in 6 locations: builtin shaper, HarfBuzz shaper, LayoutText, scene encoding, scene GPU/CPU decoders. Zero breaking changes, no UI modifications needed. Closes gg#304.

### Dependencies

- gg v0.46.7 → v0.46.8

## [0.1.20] — 2026-05-11

### Added

- **Layer Tree compositor in production pipeline** (ADR-007 Phase D) — `compositor/` package now drives the render loop. `OffsetLayer`, `PictureLayer`, `ClipRectLayer`, `OpacityLayer` provide structured composition with animated transform/opacity support. Replaces direct widget tree walks with Layer Tree traversal.
- **Persistent Layer Tree** (ADR-007 Phase D.5) — `UpdateLayerTree()` reuses layer objects across frames. 97.9% fewer allocations for 200 boundaries (613 → 13 allocs/op). Enterprise pattern validated by research (Flutter, Chrome, Qt6, Android, Skia all use persistent trees).
- **O(1) flat dirty boundary list** (ADR-028 Phase C) — `HasDirtyBoundaries()` replaces `NeedsRedrawInTreeNonBoundary()` O(n) tree walk for frame skip. 45× faster (1.2ns vs 58ns). Flutter `_nodesNeedingPaint` pattern with `DirtyBoundaryRegistrar` interface.
- **Multi-rect damage** (ADR-030) — per-draw dynamic scissor for multiple dirty rects. Zero pixel waste when dirty widgets are spatially distant. Ring buffer stores rect lists per frame. Threshold >16 rects merges to union (GDK/Sway pattern). Full stack: ui → gg `RenderDirectWithDamageRects` → wgpu `PresentWithDamage`.
- **Overlay content in boundary pipeline** (ADR-029 Phase E) — dropdown menus, dialogs rendered via same Layer Tree + boundary texture pipeline as main widgets. `PaintOverlayBoundaries()`, `AppendOverlaysToLayerTree()`. Scrim for modal overlays only (Flutter ModalBarrier).
- **Overlay hover blocking** — `overlayAwareHitTest()` checks overlay stack before root tree. Background widgets no longer receive hover when overlay is open.
- **Software backend e2e tests** — pixel-exact damage verification through wgpu software backend. HAL-level `RenderPassStats` proves scissor=48×48 (not full window). 9 e2e tests run in CI without GPU.
- **GPU pipeline diagnostic logging** — 7 log points behind `GOGPU_DEBUG_DAMAGE=1`: frame entry, root invalidate, per-boundary render/check, damage tracking, blit, blit path. `renderCount`/`blitCount` counters per frame.
- **~120 new tests**, 6 benchmarks across desktop, app, compositor, state, overlay packages.
- **3 enterprise research reports**: Layer Tree patterns (5 frameworks), multi-rect damage (4 APIs, 5 frameworks), ListView recycling (5 frameworks).

### Changed

- **Frame skip O(1)** — `NeedsRedrawInTreeNonBoundary` O(n) replaced with `HasDirtyBoundaries()` O(1) in desktop.draw frame skip check.
- **os.Getenv cached** — `GOGPU_DEBUG_DAMAGE` and `GOGPU_DAMAGE_BLIT` cached via `sync.Once`. Zero syscalls in hot path.
- **state.Bind deprecated** — use `BindToScheduler` for granular per-widget invalidation (enterprise pattern). `Bind` still works for backward compatibility.
- **Phase 7 documentation** — all docblocks updated from "Phase 4-5" to "Phase 7". Stale/contradictory comments removed.
- **Debug overlay + LoadOpLoad** — force full `canvas.Render` when `GOGPU_DEBUG_DAMAGE=1` to prevent green residue from LoadOpLoad preserved content.

### Fixed

- **Dropdown black background** — overlay boundary incorrectly marked as root (`IsRoot=true` from `Parent()==nil`) → `DrawGPUTextureBase` single-slot overwrote actual root. Fixed: `clearRootOnPictureLayers` after append.
- **Child boundary dirty ≠ root needsRedraw** — `onBoundaryDirty` callback called `ctx.InvalidateRect` which set `window.needsRedraw=true` forcing root re-render every frame. Fixed: use `RegisterDirtyBoundary` only.
- **Dropdown menu ctx.InvalidateRect leak** — menu.go called both `SetNeedsRedraw` AND `ctx.InvalidateRect` on RepaintBoundary. The `InvalidateRect` violated boundary isolation, forcing root re-render. Fixed: removed redundant `ctx.InvalidateRect` calls.
- **Child boundaries invisible** — `renderFromTreeRecursive` had depth limit that prevented child boundaries (spinner, ListView items) from rendering. Fixed: removed depth limit.

### Dependencies

- gg v0.46.7 (multi-rect damage API, per-draw scissor)
- gogpu v0.34.3
- wgpu v0.27.3 (software backend Stats, slog.Debug instrumentation)

### Known Issues

- Dropdown menu items: cyan/green debug overlays do not show on overlay menu items (debug visualization only, menu renders and functions correctly)
- GPU 10% for spinner 48×48 at 30fps (target <3%, scissor proven correct at HAL level)

## [0.1.19] — 2026-05-10

### Added

- **Per-boundary GPU textures** (ADR-007 Phase 7) — each RepaintBoundary rendered into own offscreen GPU texture. Clean boundaries reuse previous texture (0 GPU work). Compositor blits via non-MSAA path. No full widget tree traversal per frame.
- **0% GPU idle** — frame skip in `desktop.draw`: early return when no boundary is dirty and no widget needs redraw. Previous frame's GPU output reused. Verified 0% GPU on all 6 examples.
- **Offscreen boundary culling** — `isBoundaryVisible()` checks CompositorClip intersection before recording. Offscreen spinner → Draw never runs → ScheduleAnimationFrame not called → animation pumper stops → 0% GPU.
- **34 integration tests** for render loop pipeline — multi-frame spinner lifecycle, data ticker isolation, recording order, ScreenBounds accuracy, clean state early return, visibility matrix (14 subtests).
- **DrawChild skip pattern** (Flutter `paintChild`) — child boundaries are SKIPPED during parent recording. Each child boundary gets its own GPU texture, composed separately. Parent scene contains only non-boundary children. When a child boundary is dirty, the root re-records cheaply (child content skipped), then child re-renders its own texture.
- **Compositor scissor clipping** — ScrollView viewport clipping applied via GPU scissor rect during texture composition. Items outside the viewport are clipped at the GPU level, not during scene recording.
- **AnimationScheduler** (Flutter `scheduleFrame` pattern) — deferred animation frame requests at 30fps. Separates animation-driven from interaction-driven invalidation.
- **RepaintBoundary as WidgetBase property** (ADR-024) — `SetRepaintBoundary(true)` on any widget. Flutter pattern replaces wrapper-based approach. ListView items auto-boundary.
- **CrossAxisAlignment** for VBox/HBox — `CrossAxisCenter`, `CrossAxisStart`, `CrossAxisEnd`, `CrossAxisStretch`. Flutter `CrossAxisAlignment` equivalent.
- **TextModeController** optional interface — `widget.TextMode` enum (Auto/MSDF/Vector/Bitmap/GlyphMask) for explicit text rendering mode control during zoom (issue #94).
- **SVG icons in SceneCanvas** — `SVGRenderer` + `SVGFiller` interfaces on SceneCanvas. CPU rasterization via `RasterizerAnalytic` (bypasses GPU queueing on temp context).
- **2-level IconCache** (enterprise pattern) — Level 1: parsed `svg.Document` by pointer. Level 2: rasterized `*scene.Image` by (ptr, w, h, color) with LRU eviction (256 max). Before: 7.5ms/frame (50 icons). After: <1µs (cache hit).
- **DPI-aware icon rendering** (ADR-026) — render SVG icons at `ceil(logicalSize × deviceScale)` physical pixels. Qt6/Chromium/IntelliJ enterprise pattern. `DeviceScaler` interface propagates scale.
- **Damage rects passthrough** — dirty boundary rects → gg `SetPresentDamage()` → OS compositor partial present.
- **Debug overlays** (ADR-023) — `GOGPU_DEBUG_DIRTY=1` cyan flash on dirty widgets, `GOGPU_DEBUG_DAMAGE=1` green flash on gg damage regions.
- **Dirty tracking** — per-item `InvalidateRect` for ListView, `StampScreenOrigin` for correct screen-space positions, viewport clip in dirty collector.
- **Hover E2E tests** — 3 tests: button hover → boundary dirty propagation, deep nesting, full Window.HandleEvent chain.
- **36 IconCache tests** — 99%+ coverage on cache logic.
- **28 DPI-aware rendering tests** — scale 1x/2x, cache key separation, edge cases.

### Fixed

- **Double rendering of boundary items** (#94, #91) — `renderBoundaryTextures` used `depth > 1` threshold. ListView items (depth 1) rendered into BOTH root texture (inline) AND own textures (overlay blit). Alpha-blended overlap = ghost text artifacts. Fix: `depth > 0` — only root gets offscreen texture.
- **Inline child boundary hover** — dirty child boundaries didn't trigger root scene re-recording. Root texture stayed stale on hover/state changes. Fix: `paintBoundaryWithDepth` re-records parent when inline child dirty.
- **ListView hover background** — hover on ListView items now triggers root re-recording with DrawChild skip. Child boundaries are skipped during parent recording, so root re-records cheaply while items retain their own textures.
- **Force root re-recording** — `NeedsRedrawInTree` check in `desktop.draw` ensures root scene re-records when any descendant widget is dirty, even when the root boundary itself is clean.
- **ScreenOriginBase in recordBoundary** — `ScreenOriginBase` set from boundary widget's screen position before recording. Nested boundaries get correct screen-space origins for compositor texture placement.
- **Scrollbar track repeat timing** — Qt6-inspired timing: 500ms initial delay, 50ms repeat interval (QScrollBar pattern). Prevents root re-recording flood from polling-based repeat.
- **SVG icons missing** — temp `gg.NewContext()` with GPU accelerator active queued shapes instead of CPU pixmap rendering. `dc.Image()` returned empty. Fix: `SetRasterizerMode(RasterizerAnalytic)`.
- **TextField/Slider/LineChart width** — hardcoded preferred widths (100px, 200px, 308px). Now fill `MaxWidth` from layout constraints.
- **Nested boundary clip** — `DrawChild` for nested boundaries during BoundaryRecording draws directly (preserves parent PushClip).
- **ScreenOrigin positioning** — depth-based nesting, `ScreenOrigin()` for compositor texture placement.
- **Spinner intrinsic layout** — 48×48 ignores parent MinWidth.
- **Damage rect screen coords** — `onBoundaryDirty` callback now uses `ScreenOrigin + Bounds` for screen-space damage rect (was local bounds at 0,0).
- **CollectDirtyRegions ordering** — moved after `PaintBoundaryLayers` so `ScreenOrigin` is fresh from root recording. Fixes debug overlay showing damage at (0,0).
- **Pumper isolation** — suppress `onBoundaryDirty` when `desktop.draw` forces root `InvalidateScene`. Data tickers (1/sec) no longer restart 30fps animation pumper.
- **Viewport culling removed from BoxWidget** — compositor-level culling handles visibility (Flutter/Chrome/Qt6 pattern). Fixes spinner "floating" when viewport culling skipped `StampScreenOrigin`.

### Changed (Dependencies)

- **gg** v0.44.1 → **v0.46.4** (LCD ClearType glyph mask ADR-024, TagText scene text ADR-022, atlas zoom resilience, deferred ortho projection, blit scissor groups)
- **gogpu** v0.31.0 → **v0.34.0** (LCD ClearType, SubpixelLayout, three-mode D2 render loop, EventSource fix)
- **gpucontext** v0.16.0 → **v0.18.0** (SubpixelLayout API, AdapterInfo)

## [0.1.18] — 2026-05-01

### Changed (Dependencies)

- **gg** v0.43.7 → **v0.44.1**

## [0.1.17] — 2026-05-01

### Added

- **Offscreen fit-to-content rendering** (ADR-008, [#87](https://github.com/gogpu/ui/issues/87), @AgentNemo00) — `offscreen.WithFitSize()` measures widget preferred size before allocating canvas. `offscreen.WithMaxSize(w, h)` constrains expansion. Follows Flutter/Qt6/SwiftUI/Compose measure-then-render pattern. 5 new tests.

### Changed (Dependencies)

- **gg** v0.43.6 → **v0.43.7** (wgpu v0.26.12, gpucontext v0.16.0, naga v0.17.10)
- **gogpu** v0.30.3 → **v0.31.0** (runtime fullscreen toggle on all platforms)
- **gpucontext** v0.15.0 → **v0.16.0** (WindowChrome.SetFullscreen/IsFullscreen)

## [0.1.16] — 2026-04-29

### Changed (Dependencies)

- **gg** v0.43.5 → **v0.43.6**

## [0.1.15] — 2026-04-29

### Fixed

- **Taskmanager example**: collapsible headers (CPU, Memory, Disk) now use reactive `state.Signal`
  instead of static text — percentages update at 1 Hz matching chart data. ([#82](https://github.com/gogpu/ui/issues/82))

### Changed (Dependencies)

- **gg** v0.43.4 → **v0.43.5**
- **gogpu** v0.30.0 → **v0.30.3**

## [0.1.14] — 2026-04-27

### Added

- **Scene composition compositor** (ADR-007 Phase 4-5) — retained-mode rendering with display list caching:
  - **Full DrawTree every frame** via `render.Canvas` (gg.Context GPU pipeline)
  - **RepaintBoundary scene.Scene cache**: clean boundaries replay cached display lists via `ReplayScene` (Push/Translate/GPUSceneRenderer/Pop); dirty boundaries re-record
  - **Single GPU render pass**: all shapes (SDF, text, paths) queued into gg.Context, flushed via `FlushGPUWithView`
  - **No persistent CPU pixmap**: eliminated `drawDirtyRegions` union clip that caused jitter
  - **No RasterizerAnalytic hack**: GPU SDF shapes work natively (shadows, rounded corners)
  - Taskmanager GPU: 7-18% → 0-1%; IDE: hover lag eliminated; signals: text stable

- **Granular widget invalidation** (INVAL-001) — 11 interactive widgets migrated from `ctx.Invalidate()` (full-tree layout+redraw) to `SetNeedsRedraw + InvalidateRect` (visual-only, no layout):
  - button, checkbox, radio, slider, dropdown (trigger+menu), tabview, collapsible, treeview, toolbar, titlebar, textfield
  - 40 `ctx.Invalidate()` calls replaced; `ctx.Invalidate()` retained only for structural changes (text input, overlay open/close, expand/collapse)
  - ~50 regression tests verifying: no full invalidation on hover, needsRedraw set, InvalidateRect matches bounds, callbacks still fire

- **Retained-mode compositor foundations** (ADR-007 Phase 1-3):
  - **Upward dirty propagation** O(depth): `SetNeedsRedraw` → walks UP to nearest RepaintBoundary → stops
  - **RepaintBoundary display list cache**: `scene.Scene` per boundary (replaces image.RGBA/GPU texture)
  - **SceneCanvas vector text**: `scene.DrawText` (glyph outlines, scalable, enterprise-quality)
  - **RasterCache**: Flutter-pattern stability tracking for GPU texture promotion
  - **ListView hover**: `markItemDirty` on 2 items, not `ctx.Invalidate()` on 14

- **offscreen** — new `ui/offscreen` package for headless widget rendering without GPU/window/app.
  `offscreen.NewRenderer(w, h)` creates a CPU-only renderer; `.Render(widget)` lays out and draws;
  `.Image()` returns `*image.RGBA`. Options: `WithTheme`, `WithScale` (HiDPI), `WithBackground`.
  Material 3 light theme applied by default. ([#75](https://github.com/gogpu/ui/issues/75))

- **Slide and Fade transition widgets** — Flutter-style animation wrappers in `transition/` package.
  53 new tests, 98.8% coverage. ([#75](https://github.com/gogpu/ui/issues/75))

- **`examples/modular-compositor`** — reference multi-module offscreen rendering example
  for Magic Mirror-style architectures ([#75](https://github.com/gogpu/ui/issues/75))

### Performance

- **Taskmanager GPU load**: 7-18% → 0-1% (scene composition eliminates CPU pixmap upload overhead)
- **IDE hover**: lag eliminated (no full-tree invalidation on mouse events)
- **Gallery spinner**: renders at 60fps without neighbor widget jitter (no dirty region union)

### Fixed

- **Visual jitter** in gallery (spinner + progressbar + chart) — eliminated dirty region union clip
- **Text disappearing** in signals example — eliminated persistent pixmap + partial redraw artifacts
- **IDE hover lag** — eliminated `ctx.Invalidate()` full-tree walk on every mouse event
- **Black border** on window edges — background rect sized to GPU surface, not canvas
- **Circular progress spinner shape** — arc line cap `LineCapRound` per M3 spec

### Changed (Dependencies)

- **gg** v0.40.0 → **v0.43.4** — scene.AppendWithTranslation, scene fixes, stem hinting, LoadOp damage, opacity API
- **gogpu** v0.26.4 → **v0.30.0** — event-driven frame pacing, mouse grab, frameless windows
- **gpucontext** v0.11.0 → **v0.15.0** — type-safe TextureView handles (ADR-018), CursorMode
- **wgpu** v0.24.4 → **v0.26.8** — Metal cull mode, DWM fix, deferred destruction
- **naga** v0.17.0 → **v0.17.6** — DXIL 94/208 golden parity, ir.TypeSize

### Internal

- Removed legacy rendering hacks: `RasterizerAnalytic` force, `paintCPUDirect`, swapchain warmup
- `Window.ThemeBackground()` made public for compositor access
- `desktop.Run` uses `RenderModeHostManaged` (full draw, RepaintBoundary cache handles efficiency)
- `SceneCanvas.ReplayScene` uses `Scene.AppendWithTranslation` for coordinate offsetting

## [0.1.13] — 2026-04-08

### Changed (Dependencies)

- **gg** v0.39.4 → **v0.40.0** — alpha mask API (per-shape, per-layer, luminance, GPU interface)

## [0.1.12] — 2026-04-08

### Changed (Dependencies)

- **gg** v0.39.3 → **v0.39.4**, **gogpu** v0.26.3 → **v0.26.4**, **wgpu** v0.24.3 → **v0.24.4**
- Software backend: enterprise Present via GDI, core routing, adapter logging
- `GOGPU_GRAPHICS_API=software` now renders gg examples at ~64 FPS on CPU

## [0.1.11] — 2026-04-07

### Fixed

- **GLES crash on Linux X11** — indirect dependency wgpu updated to v0.24.3 which fixes
  SIGSEGV in `eglInitialize` caused by X11 display use-after-close (BUG-GLES-001)

## [0.1.10] — 2026-04-07

### Fixed

- **MSDF text overlapping on Retina** — gg v0.39.1 → **v0.39.3** fixes large text (28px+)
  overlapping letters on HiDPI displays (scale=2). (#247 in gg)

## [0.1.9] — 2026-04-07

### Changed (Dependencies)
- **gg** v0.39.0 → **v0.39.1**, **gogpu** v0.26.1 → **v0.26.3**, **wgpu** v0.23.9 → **v0.24.2**, **naga** v0.16.6 → **v0.17.0**
- Metal: texture flicker fix, SDF shapes fix, DX12 encoder pool, DXIL backend

## [0.1.8] — 2026-04-05

### Changed (Dependencies)
- **gg** v0.38.3 → **v0.39.0** (Path API: Iterate callback, SoA layout, GLES fixes)

### Internal
- `icon/svg.go`: adapted to gg Path.Iterate() API (replaces Path.Elements())

## [0.1.7] — 2026-04-05

### Fixed

- **Widget Gallery content invisible** — `isExpanded()` used duck-typing interface
  `IsExpanded() bool` to detect flex layout wrappers in VBox. `collapsible.Widget`
  also has `IsExpanded()` for its expand/collapse state, causing VBox to mistakenly
  treat collapsible sections as flex children and give them `MaxHeight=0`. Replaced
  with private marker interface `layoutExpander` using unexported method. Prevents
  any external type from accidentally satisfying the interface.
  (BUG-UI-GALLERY-001)
- **Gallery theme switching** — `onThemeChange` callback passed recursively instead
  of `nil` so dropdown keeps working after theme switch.

### Changed (Dependencies)
- **gg** v0.38.2 → **v0.38.3**
- **gogpu** v0.26.0 → **v0.26.1**
- **wgpu** (indirect) v0.23.0 → **v0.23.9**
- **naga** (indirect) v0.15.0 → **v0.16.6**
- **gputypes** (indirect) v0.3.0 → **v0.4.0**
- **golang.org/x/image** v0.37.0 → **v0.38.0**

## [0.1.6] — 2026-03-24

### Fixed

- **Examples: software adapter fallback** — All examples now check
  `AcceleratorCanRenderDirect()` before using GPU-direct SDF rendering.
  On CPU-only adapters (llvmpipe, SwiftShader), falls back to
  `canvas.Render(dc.RenderTarget())` universal path via PresentTexture.

### Changed (Dependencies)
- **gg** v0.38.1 → **v0.38.2** (GLES clip/scissor fixes)
- **gogpu** v0.25.0 → **v0.26.0**
- **wgpu** (indirect) v0.22.1 → **v0.23.0**
- **naga** (indirect) v0.14.8 → **v0.15.0**

## [0.1.5] — 2026-03-23

### Changed (Dependencies)
- **gg** v0.38.0 → **v0.38.1** (SVG renderer fixes, first-frame rendering improvements)

## [0.1.4] — 2026-03-21

### Added

- **DevTools design system** — Complete JetBrains-inspired theme with 22 component painters
  (dark/light mode), based on Int UI gray scale and JetBrains IDE styling. New `theme/devtools/`
  package with full painter set matching Material 3, Fluent, and Cupertino coverage.
- **Stripe toolbar widget** — New `core/stripe/` package for vertical tool window sidebars.
  Top/bottom button groups, hover/click/active states, pluggable Painter interface. JetBrains
  IDE-accurate sizing (40x40 buttons, 20px icons, 59px with labels).
- **TitleBar widget** — New `core/titlebar/` package for frameless window title bars. Leading/center
  child zones, window controls (minimize/maximize/close), hit-test delegation for proper drag areas.
- **SVG icon system** — Full SVG rendering via `gg/svg` package. `FromSVGXML` constructor loads
  JetBrains expui SVG icons with proper fill, stroke, fill-rule, stroke-linecap, `<circle>`,
  `<path>` elements. `SVGRenderer` interface on Canvas. 17 expui icons for toolbar and sidebar.
- **IDE layout example** — New `examples/ide/` demonstrating GoLand-inspired layout: frameless
  titlebar with toolbar, project tree, editor/terminal tabs, left/right tool window strips,
  status bar. Uses DevTools theme, SplitView, TabView, TreeView, Stripe, Toolbar.
- **Toolbar options** — `ButtonSize(px)` and `Gap(px)` for configurable toolbar button sizing.
  JetBrains defaults: 30x30 buttons, 10px gap.
- **SplitView FixedFirst** — Pixel-based panel sizing. First panel stays at constant width/height
  regardless of window resize. Drag updates pixel position.
- **Expanded widget** — New `primitives.Expanded()` wrapper for flex layout grow behavior.
- **LCD ClearType** — Subpixel text rendering enabled (`gg.LCDLayoutRGB`).
- **10 first-frame rendering tests** — Headless tests verifying all widgets render correctly
  on the very first Frame+DrawTo cycle.

### Fixed

- **TabView coordinate system** — TabView now uses local coordinates with PushTransform in Draw,
  matching SplitView pattern. Fixes first-frame rendering where tabs appeared at wrong positions.
- **Window focus redraw** — `HandleFocusChange` now requests redraw, fixing black window after
  losing and regaining focus in event-driven mode.
- **Toolbar NewRect width** — Fixed `NewRect(x, 0, x+itemW, h)` → `NewRect(x, 0, itemW, h)`.
  Each toolbar button was getting progressively wider.
- **Titlebar hover tracking** — Proper MouseLeave dispatch when cursor moves between toolbar
  children. Hit-test delegation via `HitTestPoint` interface.

### Changed (Dependencies)
- **gg** v0.37.3 → **v0.38.0** (SVG renderer, FillPath, ParseSVGPath, LCD ClearType)
- **gogpu** v0.24.4 → **v0.24.5**
- **gpucontext** v0.10.0 → **v0.11.0**
- **wgpu** (indirect) v0.21.3 → **v0.22.1**

### Removed

- **TextWidget.Italic()** — Dead code removed. Canvas.DrawText never rendered italic.

## [0.1.3] — 2026-03-17

### Fixed

- **Animation scheduling** — Fixed critical bug where animations only worked when the user
  moved the mouse. Root cause: `needsLayout` flag was unconditionally cleared after layout,
  clobbering the re-invalidation set by `tickAnimation()` during layout. Now checks
  `IsInvalidated()` before clearing. Affects all animated widgets (collapsible, slider,
  dialog, tabview, scrollview).

### Added

- **Animation frame pumper** — New `animPumper` goroutine requests redraws at ~60fps while
  animations are active. Automatically stops after 3 consecutive idle frames. Enables smooth
  animations in event-driven (on-demand) rendering mode.
- **BeginFrame timing** — New `ContextImpl.BeginFrame()` method calculates DeltaTime from
  inter-frame intervals with clamping to [0, 100ms]. Prevents animation jumps after
  background/resume or debugger pauses.
- **Collapsible DeltaTime clamping** — `tickAnimation()` clamps dt to [1ms, 32ms] instead
  of skipping on dt<=0. First frame always advances animation.
- **13 regression tests** — Animation scheduling (5), BeginFrame timing (5), collapsible
  animation (3). Key test verifies needsLayout is preserved when widget invalidates during
  layout.

## [0.1.2] — 2026-03-16

### Fixed

- **Inter font with Cyrillic/Greek/Vietnamese** — Replaced Latin-only Inter subsets
  (68KB) with full Inter 4.1 (412/420KB). Fixes [#49](https://github.com/gogpu/ui/issues/49).

### Changed (Dependencies)
- **gg** v0.37.1 → **v0.37.3** (universal Render, GLES/Software support)
- **gogpu** v0.24.2 → **v0.24.4** (env var, PresentTexture, GLES CompatibleSurface)
- **wgpu** (indirect) v0.21.1 → **v0.21.3** (core validation, DX12/GLES fixes)
- **naga** (indirect) v0.14.7 → **v0.14.8** (GLSL binding fix)

## [0.1.1] — 2026-03-15

### Changed (Dependencies)
- **gg** v0.37.0 → **v0.37.1**
- **gogpu** v0.24.1 → **v0.24.2**
- **wgpu** (indirect) v0.21.0 → **v0.21.1**

## [0.1.0] — 2026-03-15

### Added (Hover Tracking — TASK-UI-067)
- **W3C PointerEventSource** — wired `gpucontext.PointerEventSource.OnPointer()` for
  window Enter/Leave events. HoverTracker in Window performs hit-testing on MouseMove
  using ScreenBounds, synthesizes MouseEnter/MouseLeave for individual widgets.
  Enables hover cursors (pointer, text, resize) in production. 17 new tests.

### Fixed (Drag Cursor — TASK-UI-068)
- **Drag cursor maintained** — SplitView and Slider now set cursor on every drag MouseMove.
  Window skips ResetCursor in Frame() while mouse buttons are held. Cursor sync runs
  immediately after HandleEvent for responsive hover feedback in event-driven mode.

### Fixed (Event Coordinate Transform — TASK-UI-066)
- **ScrollView event dispatch** — mouse/wheel coordinates now transformed from screen
  space to content space before dispatching to children. Fixes click hit-testing for
  widgets inside scrolled containers. Removed redundant transforms from ListView/DataTable.

### Added (Widget Gallery Example)
- **Gallery example** (`examples/gallery/`) — comprehensive widget gallery demonstrating
  all 22 interactive widgets with live theme switching between Material 3, Fluent Design,
  and Cupertino design systems. Organized into collapsible sections by category.

### Changed (Dependencies)
- **gogpu** v0.24.0 → **v0.24.1**

### Added (Screen-Space Coordinates — TASK-UI-065)
- **ScreenBounds** (`widget/base.go`) — screen-space coordinate transform for overlay
  positioning inside ScrollView. Draw-pass transform stamping via `Canvas.TransformOffset()`
  + `widget.StampScreenOrigin()`. Dropdown/Popover use `ScreenBounds()` for correct
  positioning. Enterprise pattern (Flutter localToGlobal / Qt mapToGlobal). 72 files.

### Fixed (Collapsible)
- **Event forwarding** — Collapsible now properly forwards events to content widgets
  when expanded. Previously mouse clicks on content children were not dispatched.

### Fixed (App — Text Input)
- **OnTextInput handler** — EventBridge now uses `OnTextInput` callback for character
  input, replacing the `keyToRune` workaround that failed for non-ASCII characters
- **keyToRune removal** — removed fragile key-to-rune synthesis; character input now
  comes exclusively from the platform's text input API

### Added (Widget Canvas)
- **MeasureText** — new `widget.Canvas` interface method for measuring text dimensions
  without drawing. Returns `geometry.Size` with text width and height. Used by widgets
  for layout calculations (e.g., label width in ProgressBar, column sizing in DataTable).

### Fixed (App — Focus)
- **FocusManager integration** — Window now creates and wires a `focus.Manager` for
  Tab/Shift+Tab keyboard navigation. Key events flow through FocusManager before
  reaching the widget tree, enabling system-level focus traversal.
- **Tab focus redraw** — focus changes now properly trigger widget invalidation so
  focus rings are drawn/cleared immediately

### Fixed (Font)
- **Inter font full Unicode** — replaced Latin-only Inter font subsets with full
  Unicode Inter 4.1 font files. Enables Cyrillic, Greek, Vietnamese, and other scripts.

### Changed (Dependencies — Cascade Update)
- **gg** v0.36.4 -> **v0.37.0** (full ecosystem update for new wgpu HAL API)
- **gpucontext** v0.9.0 -> **v0.10.0** (TextureView.Destroy API change)
- **gogpu** v0.23.3 -> **v0.24.0** (new wgpu HAL integration)
- **wgpu** (indirect) -> **v0.21.0** (new HAL API, TextureView lifecycle)
- **naga** (indirect) -> **v0.14.7**

### Refactored (API Consistency)
- **TextAlign type** — `Canvas.DrawText` alignment parameter changed from raw `float32`
  to type-safe `widget.TextAlign` enum (Left/Center/Right). 65 files updated.
- **Painter naming** — linechart `DrawChart`→`PaintChart`, `ChartState`→`PaintState`;
  progressbar `ColorScheme`→`ProgressBarColorScheme`

### Added (M3 Painters for Phase 4 Widgets)
- **12 new Material 3 painters** (`theme/material3/`) — ProgressBar, Progress (circular),
  Collapsible, Popover, SplitView, GridView, LineChart, TreeView, DataTable, Toolbar,
  Menu, Docking. All with M3 color roles and tests.

### Added (Phase 4 — Enterprise Widgets)
- **TreeView** (`core/treeview/`) — hierarchical tree with expand/collapse, virtualized
  rendering, keyboard nav, indent with connector lines, selection, Painter pattern
- **DataTable** (`core/datatable/`) — sortable column table, fixed header, virtualized
  rows, row selection, column alignment, zebra striping, sort indicators
- **Toolbar** (`core/toolbar/`) — horizontal action bar with icon buttons, separators,
  spacers, custom widget items, keyboard nav
- **Menu System** (`core/menu/`) — MenuBar + ContextMenu, submenus, separators,
  disabled items, shortcut display, overlay integration

### Added (Phase 4 — Design Systems & Infrastructure)
- **Fluent Design Theme** (`theme/fluent/`) — Microsoft Fluent Design with 9 painters,
  accent color system, inner focus ring, 4px radii, light/dark variants. 42 tests.
- **Cupertino Theme** (`theme/cupertino/`) — Apple HIG with 9 painters, iOS toggle switch
  checkbox, segmented control tabview, transparent scrollbar, pill buttons. 44 tests.
- **i18n System** (`i18n/`) — Locale, Bundle, Translator with 4-level fallback,
  CLDR plural rules (6 language families), RTL detection, reactive LocaleSignal. 32 tests, 97.9%

### Added (Phase 4 — Continued)
- **Docking System** (`core/docking/`) — IDE-style dockable panels with border layout
  (Left/Right/Top/Bottom/Center zones), tabbed panel groups, auto-collapse empty zones,
  Dock/Undock/MovePanel API. 62 tests, 95.3%
- **Testing Utilities** (`uitest/`) — reusable MockCanvas (records all draw calls),
  MockContext, event factories, widget helpers, custom assertions. Replaces 30+ duplicate
  mocks across test files. 53 tests, 93.1%

### Added (Phase 4 Infrastructure)
- **Font Registry** (`theme/font/`) — CSS font-weight matching algorithm (W3C spec),
  Weight (100-900), Style (Normal/Italic), Family/Face, thread-safe Registry. 20 tests, 97.7%
- **Icon System** (`icon/`) — vector path icons (MoveTo/LineTo/CubicTo/Close), IconWidget,
  thread-safe Registry, 10 built-in Material-style icons, De Casteljau cubic Bezier. 39 tests, 97.6%
- **Drag and Drop** (`dnd/`) — DragSource/DropTarget interfaces, Manager with full lifecycle,
  5px drag threshold, Escape cancel, drop effects. Foundation for docking system.

### Added (Phase 4 Widgets)
- **Circular Progress** (`core/progress/`) — determinate arc + indeterminate spinner,
  polyline arc approximation, time-based animation, Painter pattern. 48 tests, 97.4%
- **Popover/Tooltip** (`core/popover/`) — click-triggered popover + hover-triggered tooltip,
  12 placements with auto-flip, viewport clamping, overlay integration, dismiss-on-click-outside
- **SplitView** (`core/splitview/`) — resizable split panels (H/V), draggable divider,
  min constraints, double-click collapse, handle dots, cursor change. 37 tests, 96.8%

### Added (Performance Benchmarks)
- **Benchmarks** across 5 packages: layout (flex/stack/grid/cache), signals (get/set/computed/effect/chain),
  widget tree (walk/bounds), ListView virtualization (layout/scroll/selection), animation (tween/spring/sequence).
  36 benchmarks total. Key results: ~17ns signal read, ~150ns 10-child flex layout, ~28ns tween tick,
  zero allocations on hot paths.

### Added (Dirty Region Tracking — TASK-UI-053)
- **Dirty region tracker** (`internal/dirty/`) — collects dirty widget bounds,
  merges overlapping/nearby regions, enables partial repaints. Collector walks
  widget tree via NeedsRedraw(), Tracker optimizes regions with configurable
  merge gap. Full repaint fallback when >16 regions. 43 tests, 100% coverage.

### Added (Transitions — TASK-UI-025)
- **Transition wrapper** (`transition/`) — widget enter/exit animations via wrapper
  pattern. Effects: FadeIn/Out, SlideIn/Out (4 directions), ScaleIn/Out. Show()/Hide()
  trigger animated transitions with time-based progress. OpacityPusher graceful
  degradation, retained-mode integration. 38 tests, 98.7% coverage.

### Added (Animation Presets — TASK-UI-024A)
- **M3 motion presets** (`animation/presets.go`) — Material 3 duration tokens
  (Short1..ExtraLong4), easing aliases (Standard, Emphasized, Decelerate, Accelerate),
  preset builders: FadeIn/Out, SlideIn (4 directions), ScaleIn/Out, DialogEnter/Exit,
  MenuEnter/Exit, SnackbarEnter/Exit
- **Orchestration helpers** (`animation/orchestrate.go`) — Stagger (staggered start),
  Chain, Group, RepeatN/RepeatForever, Reverse, WithDelay

### Added (GridView Widget — TASK-UI-022)
- **GridView widget** (`core/gridview/`) — virtualized 2D grid for large datasets.
  Fixed cell size with auto-fit columns, cell recycling (only visible rows rendered),
  single selection, keyboard navigation (arrows/Home/End/PgUp/PgDn), hover highlight.
  Content[C] (CDK) architecture, BuildCell convenience API, Painter pattern,
  4-level signal bindings. 90 tests, 92.1% coverage.

### Added (ListView Widget — TASK-UI-021)
- **ListView widget** (`core/listview/`) — virtualized scrollable list for large
  datasets. Fixed item height with efficient recycling: only visible items are
  laid out, drawn, and cached. Built on ScrollView for scrolling, with
  Content[C] (CDK) as internal architecture and `BuildItem` convenience API.
  Mouse click selection (single/multi/none), hover highlight, keyboard navigation
  (Up/Down/Home/End/PgUp/PgDn), divider lines. Two-way SelectedIndexSignal and
  SelectedIndicesSignal bindings. Pluggable Painter pattern with DefaultPainter
  fallback. M3 ListViewPainter with HCT-derived selection/hover colors.
- **Material 3 ListViewPainter** (`theme/material3/listview.go`) — M3 list item
  rendering with hover overlay, selection background, divider colors from theme

### Added (Box — Horizontal Layout, TASK-UI-058)
- **HBox / VBox direction** — Box widget now supports horizontal layout via
  `SetDirection(DirectionHorizontal)`, `HBox()` / `VBox()` convenience constructors,
  `DirectionSignal` reactive binding. Children laid out left-to-right with gap.
  Mount/Unmount lifecycle for signal cleanup.

### Added (LineChart Widget — TASK-UI-060)
- **LineChart widget** (`core/linechart/`) — real-time line chart for time-series
  data visualization. Multiple series with colors, rolling window (MaxPoints),
  auto-scaling Y axis, grid lines, Y-axis labels. Right-aligned scrolling
  (newest data at right edge). Pluggable Painter pattern, signal bindings,
  thread-safe PushValue. 43 tests, 98.8% coverage.

### Added (ProgressBar Widget — TASK-UI-059)
- **ProgressBar widget** (`core/progressbar/`) — linear progress bar (0-100%).
  Rounded corners via PushClipRoundRect, optional label with custom format,
  configurable bar/track/label colors. 4-level signal priority for value binding,
  Painter pattern, Mount/Unmount lifecycle. 31 tests, 99.3% coverage.

### Added (Collapsible Section Widget — TASK-UI-061)
- **Collapsible widget** (`core/collapsible/`) — expandable section with clickable
  header and animated content reveal. Tween animation with EaseInOutCubic,
  keyboard focus (Enter/Space), arrow indicator, content clipping during
  animation. Painter pattern, two-way ExpandedSignal binding.
  76 tests, 98.2% coverage.

### Fixed (ScrollView)
- **Drag sticking** — mouse drag no longer "sticks" when releasing outside the
  scrollview bounds. ButtonState tracking in event_bridge properly sends
  MouseUp for all buttons held at previous frame
- **Track page-scroll** — click on scrollbar track now scrolls by one page
  (viewport height) instead of jumping to click position
- **Track repeat** — holding mouse on scrollbar track now auto-repeats
  page scrolling (500ms initial delay, 100ms repeat interval)
- **Wheel direction** — mouse wheel now scrolls in natural direction
  (wheel up = content up = negative delta)

### Fixed (Box Widget)
- **WheelEvent dispatch** — Box now properly dispatches WheelEvent to children,
  enabling mouse wheel scrolling for ScrollView inside Box containers
- **Child clipping** — Box with border or rounded corners now calls PushClip
  to clip child content to container bounds, preventing overflow
- **Border z-order** — border is now drawn AFTER children so it renders on top
  of content instead of being obscured by child widgets

### Added (Canvas / GPU Clipping)
- **PushClipRoundRect** — new `widget.Canvas` interface method for GPU SDF-based
  rounded rectangle clipping. `Canvas` implementation delegates to `gg.ClipRoundRect()`;
  `SceneCanvas` falls back to rectangular clip (scene.Scene support pending gg#202).
  `Box.Draw` automatically uses `PushClipRoundRect` when `radius > 0`, properly
  clipping child content to rounded corners without padding workarounds

### Fixed (Canvas / GPU Clipping)
- **PushClip with gg.ClipRect** — Canvas.PushClip now sets clip rect on the
  underlying gg.Context via ClipRect(), enabling hardware GPU scissor rect
  clipping. Previously only tracked clip bounds internally without informing
  the rendering backend, so GPU-rendered shapes ignored clip regions

### Fixed (Event Bridge)
- **ButtonState tracking** — event_bridge now tracks which mouse buttons were
  held in the previous frame and synthesizes MouseUp events for buttons that
  were released between frames, preventing drag state from sticking

### Changed (Dependencies)
- **gg** v0.35.3 → **v0.36.4** (GPU GlyphMask cache, RoundRectShape SDF, scene clip support, font hinting, ClearType LCD subpixel, GPU scissor rect clipping, GPU RRect SDF clip via ClipRoundRect)
- **golang.org/x/image** v0.36.0 → **v0.37.0**
- **golang.org/x/text** v0.34.0 → **v0.35.0**
- **go-text/typesetting** v0.3.3 → **v0.3.4**

### Added (scene.Scene Integration — TASK-UI-057 SP3)
- **SceneCanvas adapter** (`internal/render/scene_canvas.go`) — implements `widget.Canvas`
  by recording drawing commands into `scene.Scene` for tile-parallel rendering.
  All shape operations (rect, round rect, circle, line) map to scene shapes.
  Text rendering via gg.Context pass-through preserves MSDF quality.
  PushClip/PopClip and PushTransform/PopTransform with internal stacks for
  visibility optimization.
- **RepaintBoundary scene integration** (`primitives/repaint_boundary.go`) —
  threshold-based rendering selection: RepaintBoundaries >= 128x128 pixels use
  `scene.Scene` + `scene.Renderer` for tile-parallel rendering. Smaller widgets
  use the traditional `gg.Context` path. Scene resources (Renderer, Scene, Pixmap)
  are lazily initialized and reused across frames. Zero breaking changes to
  `widget.Canvas` interface.

### Added (TabView Widget — TASK-UI-029)
- **TabView widget** (`core/tabview/`) — tabbed navigation container with lazy
  content switching (only selected tab laid out/drawn). Horizontal tab bar with
  Top/Bottom positioning. Click-to-select, closeable tabs (per-tab override),
  keyboard navigation (Left/Right with wrap-around, Home/End, skip disabled).
  Two-way SelectedSignal binding. Pluggable Painter pattern with DefaultPainter
  fallback. Equal-width tab distribution. 92.1% test coverage.
- **Material 3 TabViewPainter** (`theme/material3/tabview.go`) — M3 tab bar
  rendering with HCT-derived colors, 3px rounded indicator, hover overlay,
  focus ring, close button X icon, disabled state

### Added (ScrollView Widget — TASK-UI-028)
- **ScrollView widget** (`core/scrollview/`) — scrollable container with content
  clipping via PushClip/PopClip and translation via PushTransform. Vertical (default),
  horizontal, and bi-directional scrolling. Mouse wheel, keyboard navigation
  (arrows, Page Up/Down, Home/End), scrollbar thumb drag, click-on-track scrolling.
  Scrollbar visibility: auto/always/never. Two-way ScrollX/ScrollY signal bindings.
  Pluggable Painter pattern with DefaultPainter fallback. 96.5% test coverage, ~1,170 LOC.
- **Material 3 ScrollbarPainter** (`theme/material3/scrollbar.go`) — M3 scrollbar
  rendering with HCT-derived colors and opacity states (normal/hover/drag)

### Added (Animation Engine — TASK-UI-024)
- **Animation engine** (`animation/`) — comprehensive animation system with:
  - **Tween animations**: Builder pattern `To(signal, target).Duration(d).Ease(e).Start(ctrl)`.
    Delay, repeat (finite/infinite), auto-reverse, OnDone callback.
  - **Spring physics**: Damped harmonic oscillator with sub-stepped Euler integration.
    `SpringTo(signal, target).Stiffness(s).DampingRatio(d).Start(ctrl)`.
    Dual-threshold convergence (restDelta + restSpeed). Velocity preservation on retarget.
  - **CubicBezier easing**: 11-sample table + Newton-Raphson + bisection fallback (~10ns/eval).
  - **ThreePointCubic**: Exact M3 Emphasized curve (two joined cubic segments).
  - **M3 motion tokens**: 7 easing curves, 16 duration tokens (50ms-1000ms),
    4 damping ratios, 4 stiffness presets (from Jetpack Compose).
  - **Tween[T] evaluator**: Generic type mapping (Color, Point, Size) from float32 progress.
    Flutter pattern: engine drives float32, Tween maps to any type.
  - **Composition**: Sequence (chain) and Parallel for multi-animation orchestration.
  - **Controller**: Auto-cancel per signal, Tick(dt), HasActive(), CancelAll().
    Spring velocity transfer on auto-cancel. 0% CPU when idle.
  - 73 tests, 90.3% coverage, ~2,800 LOC total.

### Added (Dialog Widget — TASK-UI-014)
- **Dialog/Modal widget** (`core/dialog/`) — modal dialog with backdrop overlay,
  title, optional content widget, and action buttons. Dismissible via backdrop
  click and Escape key (configurable). Focus trapping with Tab/Shift+Tab cycling
  between action buttons. Enter/Space activates focused action. 4-tier title
  resolution (ReadonlySignal > Signal > Fn > Static). Pluggable Painter pattern
  with DefaultPainter fallback. Convenience constructors: `Alert()`, `Confirm()`.
  96.9% test coverage.
- **Material 3 DialogPainter** (`theme/material3/dialog.go`) — M3 dialog rendering
  with HCT-derived colors, 24dp corner radius, scrim backdrop, focus ring

### Added (Slider Widget — TASK-UI-015)
- **Slider widget** (`core/slider/`) — draggable slider for selecting numeric values
  within a range. Continuous and discrete (step snapping) modes. Horizontal and
  vertical orientations. Mouse drag, click-on-track, full keyboard navigation
  (arrows, Home/End, PgUp/PgDn). Two-way ValueSignal binding, DisabledSignal.
  Pluggable Painter pattern with DefaultPainter fallback. 94.6% test coverage.
- **Material 3 SliderPainter** (`theme/material3/slider.go`) — M3 slider rendering
  with HCT-derived colors, state modifiers (hover/drag/focus/disabled), tick marks

### Added (Retained-Mode Rendering — TASK-UI-057 Sub-Phase 2)
- **RepaintBoundary widget** (`primitives/repaint_boundary.go`) — caches child
  subtree as CPU-side pixel buffer (image.RGBA). When the subtree is clean, the
  cached image is composited directly instead of re-rendering descendants.
  Flutter RepaintBoundary pattern for explicit opt-in caching boundaries.
- **DrawImage on Canvas** — `widget.Canvas.DrawImage(img, at)` for blitting cached
  pixel buffers. Used by RepaintBoundary for cache compositing.
- **CachedWidgets in DrawStats** — `widget.DrawStats.CachedWidgets` counter tracks
  how many widgets were served from cache during draw traversal.

### Added (Professional Font — Inter)
- **Inter font for UI text** — replaced Go fonts (goregular/gobold) with
  Inter Regular (400) and Bold (700). Inter is designed specifically for
  computer screens and UI, used by GitHub, Figma, and VSCode. Embedded via
  `go:embed` (+136KB, latin subset). SIL OFL / Apache 2.0 license.

### Changed (Render Package)
- **Renamed `ctx` to `dc`** in render package — follows gg ecosystem convention
  where `*gg.Context` is called `dc` (drawing context), not `ctx` (`context.Context`)

### Changed (Dependencies)
- **gg** v0.34.0 → **v0.35.3** (GlyphCache, stem darkening, MSDF FontID collision fix)
- **gogpu** v0.23.0 → **v0.23.2** (Retina contentsScale fix) — examples only

### Added (Retained-Mode Rendering — TASK-UI-057 Sub-Phase 1)
- **Draw tree traversal with statistics** — `widget.DrawTree()` draws the root widget
  and collects per-widget dirty/clean statistics via `widget.DrawStats`
- **Draw statistics collection** — `widget.CollectDrawStats()` walks the tree without
  drawing, reporting dirty, clean, skipped, and total widget counts (for diagnostics)
- **FrameStats.DrawStats** — per-widget draw statistics are now included in
  `app.FrameStats`, accessible via frame callback for performance monitoring
- **Window.LastDrawStats()** — accessor for the most recent draw traversal statistics
- **Window.DrawTo() uses DrawTree** — the draw pass now collects statistics during
  rendering, providing observability into the retained-mode dirty-tracking system

### Added (Signal Lifecycle — SIGNALS-006/007/008)
- **Automatic signal binding lifecycle** — widgets with signal bindings now
  auto-subscribe on mount and auto-cleanup on unmount (no memory leaks):
  - `widget.Lifecycle` interface (`Mount(ctx)` / `Unmount()`) — opt-in for widgets with signals
  - `widget.SchedulerRef` interface — avoids circular imports between widget and state
  - `WidgetBase.AddBinding()` / `AddEffect()` / `CleanupBindings()` — binding management
  - `widget.MountTree()` / `UnmountTree()` — recursive tree lifecycle helpers
  - `Window.SetRoot()` triggers mount/unmount automatically
- **Scheduler push-based invalidation** — `Scheduler.SetOnDirty()` callback wakes
  render loop via `RequestRedraw()` when signals change. Reflush loop protection
  (max 2 re-flushes per frame) prevents infinite loops
- **ReadonlySignal widget options** — computed signals (`state.NewComputed()`) can
  now be passed to widgets:
  - button: `TextReadonlySignal`, `DisabledReadonlySignal`
  - checkbox: `LabelReadonlySignal`, `DisabledReadonlySignal`
  - radio: `GroupDisabledReadonlySignal`
  - Priority: ReadonlySignal > Signal > Fn > Static
- **All 6 widget types implement Lifecycle** — button, checkbox, radio, textfield,
  dropdown, primitives/text auto-bind signals on mount

### Added (Examples)
- **Signals demo** (`examples/signals/`) — standalone example demonstrating all signal
  features: TextSignal, ContentSignal, CheckedSignal, SelectedSignal, DisabledSignal.
  Event-driven rendering (0% CPU when idle), GPU-accelerated via ggcanvas

### Fixed
- **Disabled button text color** — DefaultPainter now uses solid gray (`RGBA 0.62`)
  for disabled text instead of near-invisible alpha-blended black (`RGBA 0.12 @ 38%`).
  Disabled background changed to visible light gray (`RGBA 0.92`)

### Dependencies
- gg v0.33.5 → v0.34.0, gogpu v0.22.11 → v0.23.0 (HiDPI support)
- gg v0.33.5 → v0.33.6, gogpu v0.22.9 → v0.22.11, wgpu v0.20.0, gputypes v0.3.0
  (wgpu enterprise-grade validation layer: core validation, typed errors, deferred errors)
- gg v0.33.3 → v0.33.5 (per-batch GPU text color fix — each DrawText call now
  renders with its own color instead of all text sharing the first call's color)

### Added (Signals Integration)
- **Reactive signal bindings for all core widgets (SIGNALS-001..005)** — push-based
  state management via coregx/signals integration across the entire widget tree:
  - button: TextSignal(Signal[string]), DisabledSignal(Signal[bool])
  - checkbox: CheckedSignal(Signal[bool]) (two-way), LabelSignal(Signal[string]),
    DisabledSignal(Signal[bool])
  - radio: SelectedSignal(Signal[string]) (two-way),
    GroupDisabledSignal(Signal[bool])
  - primitives/text: ContentSignal(ReadonlySignal[string])
  - Priority resolution: Signal > Fn > Static (backward compatible)
  - Unified PropertySignal naming convention across all widgets

### Deprecated
- textfield.Value() — use textfield.ValueSignal() instead
- dropdown.Signal() — use dropdown.SelectedSignal() instead

### Added
- **Overlay infrastructure** (`overlay/`) — window-level overlay stack for popups, dropdowns, tooltips, and modals. Stack with push/pop/remove, Container with dismiss-on-click-outside and Escape key, Position helper with viewport clamping and flip logic. 30+ tests
- **Dropdown/Select widget** (`core/dropdown/`) — full-featured dropdown with trigger, floating menu overlay, keyboard navigation (Up/Down/Enter/Escape/Home/End), mouse hover highlight, mouse wheel scrolling, max visible items with clipping, signal two-way binding, accessibility (role=combobox). 11 functional options, pluggable Painter interface, 55 tests
- **Material 3 Dropdown painter** (`theme/material3/dropdown.go`) — outlined trigger with chevron indicator, menu with hover/selected highlights, theme-derived colors
- **ThemeScope widget** (`primitives/themescope.go`) — overrides theme for widget subtree. Nested scoping (inner wins), nil passthrough, context wrapper pattern. 22 tests
- **TextField widget** (`core/textfield/`) — full-featured text input with cursor, selection, clipboard (Ctrl+A/C/X/V), password masking, validation, signal two-way binding, accessibility (role=textbox). 12 functional options, pluggable Painter interface, 55 tests
- **Material 3 TextField painter** (`theme/material3/textfield.go`) — outlined variant with theme-derived colors (Primary focus, Outline unfocused, Error invalid)
- **OverlayManager interface** (`widget/context.go`) — `PushOverlay`, `PopOverlay`, `RemoveOverlay` on Context for widget access to overlay stack
- **WindowSize on Context** (`widget/context.go`) — `WindowSize()` method for overlay positioning calculations

### Changed
- **Update gg v0.32.0 → v0.33.0** — includes image clipping (image-as-shader pattern),
  anti-aliased clip masks (4x Y-supersampling), DrawImageRounded/DrawImageCircular convenience
  methods, MSL backend fixes for Apple Silicon, and Linux/macOS SIGSEGV fix
  ([gg#155](https://github.com/gogpu/gg/issues/155),
  [naga#38](https://github.com/gogpu/naga/pull/38),
  [ui#23](https://github.com/gogpu/ui/issues/23),
  [goffi#19](https://github.com/go-webgpu/goffi/issues/19))
- **Multi-layer box shadow** — Material Design elevation now uses 3-4 concentric semi-transparent rounded rects (approximated Gaussian blur) instead of single flat rectangle. Levels 1-5 with progressive elevation
- **GPU direct rendering** — hello example switched from CPU readback (`RenderTo`) to zero-copy GPU surface rendering (`RenderDirect`). Single render pass, no CPU readback
- **Material Design card layout** — hello example wraps content card in outer container with 24px margin
- **Automatic resource cleanup** — examples updated to use gogpu `App.TrackResource()` for automatic ggcanvas shutdown

### Fixed
- **Text vertical alignment** — `DrawText` now centers text vertically within bounds using `(boundsHeight - textHeight)/2 + ascent` instead of top-anchoring at `ascent`
- **Box shadow direction** — shadow offset now includes horizontal component matching Material Design light source

### Dependencies
- gg v0.29.0 → v0.33.1 (smart rasterizer selection, image clipping, AA clip masks, FDot16 overflow fix, aaShift=2)
- gogpu v0.19.6 → v0.22.6 (Vulkan copy stride fix, X11 multi-touch, Wayland support, Metal vertex descriptor fix)
- wgpu v0.16.9 → v0.19.5 (Metal vertex descriptor, Vulkan surface validation, public API root package)
- naga v0.14.1 → v0.14.5

### Phase 2: Interactive Widgets (Complete — 16/16 tasks)

Interactive widgets with 3-layer architecture (ADR-003), keyboard focus management,
CDK foundation, overlay infrastructure, and Material Design 3 theming with pluggable painters.

#### Added

- **cdk** -- Component Development Kit foundation (ADR-003)
  - `Content[C]` polymorphic content interface for composite widgets
  - `StringContent`, `FuncContent[C]`, `WidgetContent` implementations
  - Foundation for Phase 3 container widgets (VirtualizedList, Tabs, ComboBox)
  - 15 tests, 100% coverage

- **core/button** -- Generic button widget with pluggable Painter
  - `button.Widget` with functional options pattern
  - `Painter` interface for design-system-agnostic rendering
  - `DefaultPainter` as minimal fallback (gray, no design system)
  - `PaintState` struct for painter context with `ButtonColorScheme` for theme-derived colors
  - 4 variant styles: Filled, Outlined, TextOnly, Tonal
  - 3 size presets: Small (32px), Medium (40px), Large (48px)
  - Mouse click and keyboard (Enter/Space) activation
  - Hover/press/focus visual states with color modifiers
  - Fluent styling: Padding, MinWidth, MaxWidth, SetBackground, SetRounded
  - 75+ tests (external + internal), 96%+ coverage

- **theme/material3** -- Material Design 3 theme + component painters (moved from `material3/`)
  - `ButtonPainter` implementing `core/button.Painter` with M3 visual style
  - `CheckboxPainter` implementing `core/checkbox.Painter` with M3 visual style
  - `RadioPainter` implementing `core/radio.Painter` with M3 visual style
  - Painters hold `*Theme` field and resolve colors from M3 ColorScheme instead of hardcoded values
  - M3 color palette: primary, outline, secondary container, on-colors
  - Light/Dark color schemes with 29 color roles
  - Tonal palette generation (primary, secondary, tertiary, neutral, error)
  - 15 typography roles (Display, Headline, Title, Body, Label x 3 sizes)
  - 7-level shape scale (None to Full)
  - HCT (Hue, Chroma, Tone) color space approximation via HSL
  - 50+ tests (external + internal), 97%+ coverage

- **core/checkbox** -- Toggleable checkbox widget with pluggable Painter
  - Three visual states: unchecked, checked, indeterminate
  - `Painter` interface for design-system-agnostic rendering
  - `DefaultPainter` as minimal fallback (gray, no design system)
  - Mouse click and keyboard (Space) activation
  - `LabelOpt` for text label, `Disabled` for read-only state
  - Implements `widget.Focusable` for Tab navigation with focus ring
  - 96%+ coverage

- **core/radio** -- Mutually exclusive radio group widget with pluggable Painter
  - `NewGroup` with functional options: `Items`, `Selected`, `OnChange`, `DirectionOpt`
  - `ItemDef{Value, Label}` for item definition
  - Vertical (default) and Horizontal layout directions
  - Arrow key navigation within group (Up/Down or Left/Right)
  - Space/Enter selection on focused item
  - `Painter` interface with `DefaultPainter` fallback
  - Individual items implement `widget.Focusable`
  - 96%+ coverage

- **focus** -- Keyboard focus management
  - `focus.Manager` with delegation pattern (public wrapper around internal impl)
  - Tab/Shift+Tab navigation through focusable widgets
  - Keyboard shortcut registration and dispatch
  - Focus ring drawing with configurable offset/color
  - `focus.Shortcut` for key combination matching
  - 44 tests (39 external + 5 internal)
  - 95.2% coverage (focus), 15.2% (internal/focus)

- **widget** -- Added Focusable interface and ThemeProvider
  - `IsFocusable`, `SetFocused`, `IsFocused` for keyboard focus support
  - `ThemeProvider` interface for dark/light mode queries (`IsDark()`)
  - `Context.ThemeProvider()` / `Context.SetThemeProvider()` for theme access from widgets

#### Architecture (ADR-003)

- 3-layer architecture: Foundation → CDK → Core Widgets / Design Systems
- Design-system-agnostic widgets in `core/` with pluggable `Painter` interfaces
- Design system implementations in `theme/material3/`, `fluent/` (planned), `cupertino/` (planned)
- Content[C] polymorphic pattern in `cdk/` for Phase 3 composite widgets

#### Dependencies

- gg v0.28.2 → v0.28.3 (wgpu v0.16.2 — Metal autorelease pool fix)
- gogpu v0.18.2 → v0.19.0 (cross-platform Rust backend) in hello example
- wgpu v0.16.1 → v0.16.2 in hello example

#### Statistics

- **New tests:** 440+ (core/button: 75+, core/checkbox: 40+, core/radio: 40+, core/textfield: 55, core/dropdown: 55, overlay: 30+, focus: 44, material3: 50+, cdk: 15, themescope: 22)
- **Total tests:** 1,500+
- **Total packages:** 25

---

### Phase 1: MVP

Complete MVP with accessibility, reactive state, widget primitives, and window integration.

#### Added

- **a11y** — Accessibility foundation (Day 1 requirement)
  - 35+ ARIA roles across 5 categories (Structural, Input, Display, Container, Navigation)
  - `Accessible` interface: Role, Label, Hint, Value, State, Actions
  - `AccessibilityNode` with stable uint64 IDs (atomic counter, not pointer-based)
  - `TreeProvider` interface + `MemoryTree` with O(1) ID lookup and dirty tracking
  - `Announcer` interface + `NoOpAnnouncer` default
  - `CheckedState` enum (Unchecked/Checked/Mixed)
  - 99.1% test coverage

- **state** — Reactive signals integration (coregx/signals v0.1.0)
  - Type aliases: `Signal[T]`, `ReadonlySignal[T]`, `Computed`, `Effect`
  - `Bind[T]` connects signal changes to `widget.Context.Invalidate()`
  - `BindToScheduler[T]` for batched rendering through `Scheduler`
  - `Scheduler` with `MarkDirty`, `Flush`, `Batch` and deduplication
  - `NewEffect` and `NewEffectWithCleanup` for side effects
  - 100% test coverage

- **primitives** — Basic widget primitives with Tailwind-style fluent API
  - `Box` — container with Padding, Background, Rounded, Border, Shadow, Gap
  - `Text` — static text with FontSize, Color, Bold, Italic, Align, MaxLines, Ellipsis
  - `TextFn` — reactive text via `func() string` (auto-updates with signals)
  - `Image` — image display with Fit modes (Cover, Contain, Fill, None), Rounded, Alt
  - All primitives implement `widget.Widget` and `a11y.Accessible`
  - Builders ARE widgets (no separate `.Build()` step)
  - 94.4% test coverage

- **app** — Window integration via gpucontext interfaces
  - `App` with Options pattern (`WithWindowProvider`, `WithPlatformProvider`, `WithTheme`)
  - `Window` manages widget tree lifecycle (SetRoot, Frame, HandleEvent)
  - Event bridge translates platform events to `ui/event` types
  - Headless mode for testing (nil providers, 800x600 default)
  - DPI scaling via `WindowProvider.ScaleFactor`
  - Cursor forwarding to `PlatformProvider.SetCursor`
  - Dependency inversion: imports `gpucontext` interfaces only, never `gogpu`
  - 98.6% test coverage

#### Dependencies

- Added `github.com/coregx/signals` v0.1.0
- Added `github.com/gogpu/gpucontext` v0.8.0
- Updated `github.com/gogpu/gg` v0.15.7 → v0.28.1
- Updated `github.com/gogpu/gogpu` v0.17.0 → v0.18.1 (in examples)
- Updated `github.com/gogpu/gpucontext` v0.8.0 → v0.9.0

#### Statistics

- **New LOC:** ~8,900
- **Total LOC:** ~40,000
- **New tests:** ~250
- **Total tests:** 1,017
- **Average coverage:** ~97%

---

### Phase 1.5: Extensibility Foundation

Extensibility infrastructure enabling third-party widgets, themes, and layouts:

#### Added

- **registry** — Widget factory registration
  - `RegisterWidget()` for dynamic widget creation by name
  - `CreateWidget()` for factory-based instantiation
  - `ListWidgets()` for discovering registered widgets
  - Thread-safe with `sync.RWMutex`
  - `init()` auto-registration pattern for third-party extensions
  - 100% test coverage

- **layout** — Public layout API (moved from internal)
  - `LayoutAlgorithm` interface for custom layouts
  - `LayoutTree` interface for widget tree traversal
  - `RegisterLayout()` for third-party layout algorithms
  - Built-in: Flex, VStack, HStack, ZStack, Grid
  - `LayoutStyle` for declarative styling
  - 89.5% test coverage

- **theme** — Theme System Foundation + Extensions + Registry
  - `Theme` struct with Colors, Typography, Spacing, Shadows, Radii
  - `ThemeExtension` interface (Flutter-inspired):
    - `Name()`, `Merge()`, `Lerp()`, `CopyWith()` methods
  - `Register()` / `Get()` for dynamic theme switching
  - `Mode` enum: Light, Dark, System
  - Built-in presets: Light, Dark, HighContrast, DefaultTheme
  - 100% test coverage

- **plugin** — Plugin bundling system
  - `Plugin` interface with lifecycle management
  - `Dependency` with semver constraints (>=, <, ^, ~)
  - Topological sort for dependency resolution
  - `PluginContext` with registry access
  - `PluginInfo` for metadata and priority
  - 99.4% test coverage

#### Statistics

- **Phase 1.5 LOC:** ~9,200
- **Test Coverage:** 97%+ average

---

### Phase 0: Foundation Complete

Foundation packages implemented with enterprise-grade quality:

#### Added

- **geometry** — Core geometric types for UI layout
  - `Point`, `Size`, `Rect` with float32 components (GPU-compatible)
  - `Constraints` for constraint-based layout (Flutter-inspired)
  - `Insets` for padding/margin calculations
  - 98.8% test coverage

- **event** — Type-safe event system
  - `Event` interface with timestamp and consumption tracking
  - `MouseEvent` with position, button, and modifier support
  - `KeyEvent` with key codes and text input
  - `WheelEvent` for scroll handling
  - `FocusEvent` for focus management
  - `Modifiers` bitmask for Shift/Ctrl/Alt/Meta
  - 100% test coverage

- **widget** — Core widget abstraction
  - `Widget` interface: Layout, Draw, Event, Children
  - `WidgetBase` struct with thread-safe state management
  - `Context` interface for UI state (focus, time, cursor, scale)
  - `Canvas` interface for drawing operations
  - `Color` type with float32 RGBA and helpers (Hex, Lerp, WithAlpha)
  - `CursorType` enum with 12 cursor types
  - 100% test coverage

- **internal/render** — Canvas implementation
  - `Canvas` implementing widget.Canvas using gogpu/gg
  - Clip stack with intersection-based clipping
  - Transform stack with cumulative offsets
  - 96.5% test coverage

- **internal/layout** — Layout engine
  - `FlexContainer` — Full CSS Flexbox implementation
  - `VStack`, `HStack`, `ZStack` — Simple stack layouts
  - `GridContainer` — Grid layout with auto/fixed/fractional tracks
  - 89.9% test coverage

#### Statistics

- **Phase 0 LOC:** ~10,261
- **Test Coverage:** 95%+ average

---

## Version History

| Version | Phase | Description |
|---------|-------|-------------|
| v0.1.0 | MVP | Accessibility, signals, primitives, windowing |
| v0.2.0 | Beta | Interactive widgets, Material 3 |
| v0.3.0 | RC | Virtualization, animation |
| v1.0.0 | Production | Enterprise features |

---

[Unreleased]: https://github.com/gogpu/ui/compare/main...HEAD
