# gogpu/ui Roadmap

> **Version:** 0.1.23 (Custom Font Pipeline + Mac Retina Fix + CJK)
> **Updated:** May 2026
> **Go Version:** 1.25+

---

## Vision

**gogpu/ui** is a reference implementation of an enterprise-grade GUI library for Go.

**Target applications:**
- IDEs (GoLand-class)
- Design tools (Photoshop, Illustrator)
- CAD applications
- Chrome/Electron-class applications
- Professional dashboards

**Key differentiators:**
- Pure Go (zero CGO)
- WebGPU-first rendering via gogpu/wgpu
- Signals-based state management (coregx/signals)
- Enterprise features: docking, virtualization, accessibility
- Four design systems: Material 3, DevTools (JetBrains), Fluent, Cupertino

---

## Current Status

| Metric | Value |
|--------|-------|
| Packages | 56+ |
| Go Source Files | ~612 |
| Test Files | ~200 |
| Total LOC | ~189,000+ |
| Test Functions | ~7,200+ |
| Test Coverage | 97%+ |
| Linter Issues | 0 |

---

## Versioning Strategy

### Core Principle: Stay on v0.x.x

```
v0.x.x  → Active development (current)
v1.0.0  → ONLY when API stable for 1+ year
v2.0.0  → AVOID (requires /v2 import path)
```

### Version Progression:

```
v0.0.x  → Phase 0 Foundation ✅ COMPLETE
v0.1.0  → Phase 1 MVP ✅ COMPLETE
v0.1.x  → Phase 1.5 Extensibility ✅ COMPLETE
v0.2.0  → Phase 2 Beta ✅ COMPLETE
v0.2.x  → Phase 2.5 Signals Integration ✅ COMPLETE
v0.3.0  → Phase 3 RC ✅ COMPLETE
v0.4.0  → Phase 4 v1.0 features (in progress)
v0.9.0  → Pre-1.0 API freeze
v0.10+  → Stabilization
v1.0.0  → Production (when ready)
```

### API Compatibility Patterns:

| Pattern | Purpose |
|---------|---------|
| **Functional Options** | Extend API without breaking changes |
| **Interface Extension** | Optional capabilities via type assertion |
| **Config Structs** | New fields with zero-value defaults |
| **internal/** | Implementation details (can change) |
| **experimental/** | Unstable features (may change/remove) |

### Repository Strategy: Mono-repo

| Aspect | Multi-repo | Mono-repo (chosen) |
|--------|------------|-------------------|
| Versioning | Matrix | Single version |
| Diamond deps | Possible | Impossible |
| Atomic changes | Difficult | Easy |
| v2 risk | High | Low |

**Full policy:** `docs/VERSIONING.md`

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    User Application                         │
├─────────────────────────────────────────────────────────────┤
│  theme/material3  │ theme/devtools │ theme/fluent │ theme/cupertino │
│  (Complete)       │ (Complete)     │ (Complete)   │ (Complete)      │
├─────────────────────────────────────────────────────────────┤
│  core/button/      │  animation/ ✅    │  core/docking/ ✅  │
│  core/checkbox/    │  Tween, Spring    │  DockingHost       │
│  core/radio/       │  M3 motion        │  Zone, Panel       │
│  core/textfield/   │                   │                    │
│  core/dropdown/    │  transition/ ✅   │  dnd/ ✅           │
│  core/slider/ ✅   │  Enter/exit       │  Drag & Drop       │
│  core/dialog/ ✅   │  effects          │                    │
│  core/scrollview/✅│                   │  uitest/ ✅        │
│  core/tabview/ ✅  │  internal/        │  Test utilities    │
│  core/listview/ ✅ │  render/          │                    │
│  core/gridview/ ✅ │  Canvas +         │  i18n/ ✅          │
│  core/linechart/ ✅│  SceneCanvas ✅   │  Internationalize  │
│  core/progressbar/✅│  (tile-parallel  │                    │
│  core/progress/ ✅ │   scene.Scene)    │  icon/ ✅          │
│  core/collapsible/✅│                  │  Icon system       │
│  core/splitview/ ✅│                   │                    │
│  core/popover/ ✅  │                   │  theme/font/ ✅    │
│  core/treeview/ ✅ │                   │  Typography        │
│  core/datatable/ ✅│                   │                    │
│  core/toolbar/ ✅  │                   │                    │
│  core/menu/ ✅     │                   │                    │
│  focus/ overlay/ ✅│                   │                    │
├─────────────────────────────────────────────────────────────┤
│  layout/                            │  state/               │
│  VStack, HStack, Grid, Flexbox      │  coregx/signals       │
│  (Complete ✅)                      │  (Complete ✅)       │
├─────────────────────────────────────────────────────────────┤
│  widget/                            │  event/               │
│  Widget, WidgetBase, Context        │  Mouse, Keyboard      │
│  (Complete ✅)                      │  (Complete ✅)       │
├─────────────────────────────────────────────────────────────┤
│  geometry/        │  internal/render │  internal/layout     │
│  Point, Rect      │  Canvas impl     │  Flex, Stack, Grid   │
│  (Complete ✅)    │  (Complete ✅)   │  (Complete ✅)      │
├─────────────────────────────────────────────────────────────┤
│  gogpu/gg          │  gogpu/gogpu    │  coregx/signals      │
│  2D Graphics ✅    │  Windowing      │  State Management    │
└─────────────────────────────────────────────────────────────┘
```

---

## Phases

### Phase 0: Foundation ✅ COMPLETE

**Goal:** Core packages for building widgets

**Completed:**
- geometry — Point, Size, Rect, Constraints, Insets
- event — MouseEvent, KeyEvent, WheelEvent, FocusEvent, Modifiers
- widget — Widget interface, WidgetBase, Context, Canvas, Color
- internal/render — Canvas implementation using gogpu/gg
- internal/layout — Engine, FlexContainer, VStack, HStack, ZStack, Grid

---

### Phase 1: MVP (v0.1.0) ✅ COMPLETE

**Goal:** Working foundation with basic widgets

**Delivered:**
- Signals integration (coregx/signals)
- Basic primitives (Box, Text, Image)
- Public layout API
- Theme system foundation
- Window integration (app package via gpucontext interfaces)

---

### Phase 1.5: Extensibility Foundation (v0.1.x) ✅ COMPLETE

**Goal:** Enable community to create custom widgets, themes, and layouts

**Implemented Packages:**
- `registry/` — Widget factory registration (100% coverage)
- `layout/` — Public layout API with custom algorithms (89.5% coverage)
- `theme/` — Theme System + Extensions + Registry (100% coverage)
- `plugin/` — Plugin bundling with dependency resolution (99.4% coverage)

---

### Phase 2: Beta (v0.2.0) ✅ COMPLETE

**Goal:** Interactive widget library with Material Design 3

**Implemented Packages:**
- `core/button/` — Interactive button, 4 variants, 3 sizes, pluggable Painter
- `core/checkbox/` — Toggleable checkbox, 3 states
- `core/radio/` — Radio group, vertical/horizontal, arrow key navigation
- `core/textfield/` — Text input, cursor, selection, clipboard, validation
- `core/dropdown/` — Dropdown/select, overlay menu, keyboard nav, scroll
- `overlay/` — Overlay stack, container, position helper
- `focus/` — Keyboard focus management with Tab/Shift+Tab
- `theme/material3/` — M3 theme (HCT color science) + widget painters
- `cdk/` — Component Development Kit, Content[C] pattern
- `primitives/themescope.go` — Theme override for widget subtrees

---

### Phase 2.5: Signals Integration (v0.2.x) ✅ COMPLETE

**Goal:** Push-based reactive state for all widgets

**Key decisions:**
- `PropertySignal` naming convention: `TextSignal()`, `CheckedSignal()`, etc.
- Priority: ReadonlySignal > Signal > Fn > Static
- Two-way binding for stateful widgets (checkbox, radio, textfield, dropdown)
- One-way for display widgets (button text, labels, primitives)

---

### Phase 3: RC (v0.3.0) ✅ COMPLETE

**Goal:** Enterprise features, rendering optimizations, containers

**Completed:**

| Task | Description |
|------|-------------|
| Retained-mode SP1 | Dirty tracking, DrawTree, DrawStats, FrameStats |
| Retained-mode SP2 | RepaintBoundary: per-widget pixel caching |
| Retained-mode SP3 | scene.Scene integration (tile-parallel rendering, SceneCanvas) |
| Slider widget | Continuous/discrete, horizontal/vertical, M3 painter |
| Dialog widget | Modal/modeless, action buttons, focus trapping, M3 painter |
| Animation engine | Tween, Spring, CubicBezier, M3 tokens, Sequence/Parallel |
| ScrollView widget | Vertical/horizontal/both, wheel+keyboard+drag, M3 painter |
| TabView widget | Tab strip, lazy content, closeable tabs, keyboard nav, M3 painter |
| ListView widget | Virtualized list with recycling, selection, keyboard nav, M3 painter |
| GridView widget | Virtualized 2D grid with cell recycling |
| LineChart widget | Data visualization with series, axes, legends |
| ProgressBar widget | Determinate/indeterminate, linear progress |
| Progress widget | Circular/spinner progress indicators |
| Collapsible widget | Expandable/collapsible section with animation |
| SplitView widget | Resizable split panes (horizontal/vertical) |
| Popover/Tooltip | Floating content with anchor positioning |
| Transitions | Enter/exit transition effects for widget animations |
| Animation Presets | Pre-built animation orchestrations (M3 Motion) |
| Dirty Region Tracking | Optimized redraw with region-based invalidation |
| Performance Benchmarks | Comprehensive benchmark suite |
| HBox direction | Horizontal layout direction in Box |
| Task Manager Example | Full-featured demo application |

---

### Phase 4: v1.0 — In Progress

**Goal:** Production-ready enterprise library

**Completed:**

| Task | Description |
|------|-------------|
| Fluent Theme | Windows Fluent Design System painters for all widgets |
| Cupertino Theme | Apple HIG design system painters for all widgets |
| Typography System | Font registry, weights, styles, families |
| Icon System | Icon registry, drawing, widget |
| Internationalization | Locale, direction (LTR/RTL), plural rules, bundles |
| Drag & Drop | Session management, visual feedback, drop targets |
| Docking System | DockingHost, Zone, Panel for IDE-style layouts |
| Testing Utilities | Mock canvas, context, event helpers, assertions |
| TreeView widget | Hierarchical tree with expand/collapse, node management |
| DataTable widget | Column-based data display with sorting |
| Toolbar widget | Action bar with items and overflow |
| Menu widget | Menu bar, context menu, menu items |
| Dirty Region Tracking | Region collector, merge algorithm, partial repaints |
| **Layer Tree Compositor (ADR-007)** | **Flutter pipeline: PaintBoundaryLayers → BuildLayerTree → replayLayerTree** |
| **Per-boundary GPU textures** | **Each RepaintBoundary → own offscreen GPU texture** |
| **DrawChild skip (Flutter paintChild)** | **Child boundaries SKIPPED during parent recording** |
| **Compositor scissor clipping** | **Items clipped by ScrollView viewport** |
| **0% GPU idle (frame skip)** | **Early return when nothing dirty — 0% GPU on static UI** |
| **Offscreen boundary culling** | **Spinner offscreen → recording skipped → pumper stops** |
| **34 integration tests** | **Multi-frame lifecycle, visibility matrix, damage rects** |
| ListView auto RepaintBoundary | Per-item pixel caching for virtualized lists |
| DrawStats observability | CachedWidgets, DirtyRegionCount, DrawStatsProvider |
| Tracker.Intersects() fast path | O(regions) spatial check in RepaintBoundary |
| Centralized ImageCache | LRU eviction (64MB), thread-safe, per-Window lifecycle |
| Offscreen Renderer | Headless widget → *image.RGBA without GPU/window |
| Performance Benchmarks | 36 benchmarks across 5 packages |
| Task Manager Example | Full-featured demo with charts, tables, animations |
| Widget Gallery Example | All 22 widgets, 4 design systems, theme switching |
| Modular Compositor Example | Multi-module offscreen rendering (Magic Mirror pattern) |
| Hover Tracking | W3C PointerEventSource, HoverTracker, cursor management |
| ScreenBounds | Screen-space coordinate transform for overlay positioning |
| Event Coordinate Transform | ScrollView mouse/wheel coordinate transforms |
| Inter Font Unicode | Full Unicode Inter 4.1 (Cyrillic, Greek, Vietnamese) |
| **Custom Font Loading Pipeline** | **FontRegistry (global singleton, CSS weight matching), StyledTextDrawer optional interface, Plugin→Registry wiring, TextWidget.FontFamily()** |
| **Mac Retina Fix** | **gg v0.46.9 MarkDirty() logical→physical pixel fix (gg#308, @sverrehu)** |
| **CJK IsCJK Propagation** | **gg v0.46.8 ShapedGlyph.IsCJK through scene/shaper paths (gg#304)** |

**Remaining:**

| Task | Description | Priority |
|------|-------------|----------|
| **GPU spinner <3%** | **scheduler.SetOnDirty needsRedraw lifecycle optimization** | **P0** |
| **ListView hover rebuild** | **Painter pattern: hover = repaint, not widget rebuild (~15 LOC)** | **P1** |
| **Texture GC** | **Prune orphaned boundaryTextures entries (~20 LOC)** | **P1** |
| Accessibility adapters | Platform-specific AT-SPI / UIA adapters | P1 |
| RichText widget | Styled text with inline formatting, links | P2 |
| NumberField widget | Numeric input with increment/decrement, ranges | P2 |
| DatePicker widget | Calendar popup, date range selection | P2 |
| TimePicker widget | Time selection with hour/minute/AM-PM | P2 |
| ColorPicker widget | Color wheel, palette, opacity slider | P2 |
| Accordion widget | Mutually exclusive collapsible sections | P3 |
| Breadcrumb widget | Navigation breadcrumb trail | P3 |
| Stepper widget | Multi-step wizard/form progress | P3 |
| Documentation polish | Comprehensive API docs and guides | P2 |
| API review | Pre-release API audit and freeze | P0 |

---

## Rendering Performance Roadmap (ADR-007)

> **Architecture:** Hybrid CPU+GPU — industry standard (Chrome/Skia, Flutter, GTK4, Qt).
> CPU text atlas + GPU shapes + GPU compositor. Validated by source-level analysis of 8 engines.

### Current State (Intel Iris Xe, v0.1.19)

| Metric | Before (v0.1.14) | After v0.1.19 |
|--------|-------------------|---------------|
| GPU (static UI, no animations) | 8% | **0%** |
| GPU (spinner visible, 30fps) | 8% | **8%** |
| GPU (spinner offscreen) | 8% | **0%** |
| GPU readback per frame | 0 | 0 |
| Render passes (idle) | 1 | **0** (frame skip) |
| Offscreen boundary cost | Always recorded | **Culled** (CompositorClip) |

### Phase 1: Zero-Readback Compositor ✅ Done

Single-pass compositor (Flutter OffsetLayer / Chrome cc pattern):
- `FlushPixmap`: CPU pixmap upload without GPU readback
- `DrawGPUTextureBase`: pixmap as base layer (drawn first)
- `FlushGPUWithView`: GPU shapes overlay (same render pass)

### Phase 2: Scene Composition Compositor (ADR-007) ✅ Done

- **Scene composition**: full DrawTree every frame via render.Canvas (gg.Context GPU pipeline).
  RepaintBoundary cache hit = replay cached scene.Scene; miss = re-record.
  Single FlushGPUWithView render pass per frame. No retained CPU pixmap.
- **Granular widget invalidation** (INVAL-001): 11 interactive widgets migrated from
  `ctx.Invalidate()` to `SetNeedsRedraw + InvalidateRect`. ~50 regression tests.
- **GPU SDF shapes natively**: no RasterizerAnalytic hack, shadows and rounded corners
  rendered via GPU accelerator.
- **Upward dirty propagation**: O(depth) to nearest RepaintBoundary, O(1) guard.

### Phase 3: Per-Boundary GPU Textures (ADR-007 Phase 7) ✅ Done

- **Per-boundary GPU textures**: each RepaintBoundary → own offscreen MSAA texture
- **DrawChild skip**: child boundaries SKIPPED during parent BoundaryRecording (Flutter paintChild)
- **Compositor scissor clipping**: items clipped by parent viewport (ScrollView)
- **Frame skip**: early return in desktop.draw when nothing dirty → 0% GPU idle
- **Offscreen boundary culling**: isBoundaryVisible checks CompositorClip intersection
- **Pumper isolation**: ScheduleAnimationFrame only pumper trigger, data tickers don't restart 30fps
- **34 integration tests**: multi-frame lifecycle, visibility matrix, damage rects, recording order

### Phase 4: Layer Tree + Damage-Aware Compositor (ADR-007 Phase D, ADR-030) ✅ Done

- **Layer Tree compositor in production** — `compositor/` drives render loop (OffsetLayer, PictureLayer, ClipRectLayer, OpacityLayer)
- **Persistent Layer Tree** — `UpdateLayerTree()` reuses layers (97.9% fewer allocs, 613→13)
- **O(1) frame skip** — flat dirty boundary set replaces O(n) tree walk (45× faster)
- **Multi-rect damage** — per-draw dynamic scissor, ring buffer, 16-rect merge threshold
- **LoadOpLoad** — preserve previous framebuffer, blit only damage rects
- **Partial present** — `PresentWithDamage` sends dirty rects through full stack
- **Overlay boundary pipeline** — dropdown/dialog content via same Layer Tree
- **Remaining:** scheduler.SetOnDirty lifecycle → spinner GPU 10% → <3%

### Phase 5: Vello Compute Integration — Future

Full Vello 9-stage compute pipeline for GPU-accelerated path rendering:
- `internal/gpu/tilecompute/` already exists (CPU reference)
- GPU dispatch via wgpu compute shaders

### Performance Targets

| Metric | Phase 2 | Phase 3 ✅ | Phase 4 ✅ | Phase 5 |
|--------|---------|-----------|-----------|---------|
| GPU % (static UI) | 8% | **0%** | **0%** | 0% |
| GPU % (spinner) | 8% | 8% | **10%** (48×48 scissor) | <1% |
| GPU % (spinner offscreen) | 8% | **0%** | **0%** | 0% |
| GPU readback | 0 | 0 | 0 | 0 |

---

## New Widgets Roadmap

### Near-term (v0.4.x)

| Widget | Description | Complexity |
|--------|-------------|------------|
| **RichText** | Styled text with bold/italic/links, inline formatting | Medium |
| **NumberField** | Numeric input: spinner buttons, range clamping, step | Low |
| **ToggleSwitch** | iOS/Material on/off switch with animation | Low |
| **Badge** | Notification badge (dot or count) on any widget | Low |
| **Chip** | Filter/action chips (M3 spec) | Low |
| **SegmentedControl** | Toggle button group (iOS/Fluent style) | Medium |

### Mid-term (v0.5.x)

| Widget | Description | Complexity |
|--------|-------------|------------|
| **DatePicker** | Calendar popup, date ranges, locale-aware | High |
| **TimePicker** | Hour/minute selection, AM/PM, 24h formats | Medium |
| **ColorPicker** | Color wheel/palette, HSL/RGB, opacity | High |
| **Accordion** | Mutually exclusive collapsible sections | Low |
| **Breadcrumb** | Navigation breadcrumb with separators | Low |
| **Stepper** | Multi-step wizard with progress indicator | Medium |
| **SearchField** | Text input with search icon, clear, suggestions | Medium |

### Long-term (v0.6.x+)

| Widget | Description | Complexity |
|--------|-------------|------------|
| **RichTextEditor** | Editable rich text (ProseMirror-inspired) | Very High |
| **Sheet** | Bottom/side sheet overlay (M3 spec) | Medium |
| **NavigationRail** | Vertical navigation (M3 spec) | Medium |
| **Carousel** | Horizontal scroll with snap points | Medium |
| **VirtualTable** | DataTable + virtualized rows (10K+ rows) | High |
| **CodeEditor** | Syntax-highlighted code editing (IDE widget) | Very High |
| **Terminal** | Terminal emulator widget | Very High |
| **Canvas** | User-controlled drawing surface | Medium |

---

## Total Scope

| Phase | Status |
|-------|--------|
| Phase 0 (Foundation) | ✅ Complete |
| Phase 1 (MVP) | ✅ Complete |
| Phase 1.5 (Extensibility) | ✅ Complete |
| Phase 2 (Beta) | ✅ Complete |
| Phase 2.5 (Signals) | ✅ Complete |
| Phase 3 (RC) | ✅ Complete |
| Phase 4 (v1.0) | In Progress (~90%) |

---

## Dependencies

| Dependency | Version | Purpose | Status |
|------------|---------|---------|--------|
| gogpu/gg | v0.46.11 | 2D rendering + scene.Scene | ✅ Integrated |
| gogpu/gpucontext | v0.18.0 | Shared interfaces | ✅ Integrated |
| gogpu/gogpu | v0.35.0 | Windowing, Browser/WASM (examples) | ✅ Integrated |
| coregx/signals | v0.1.0 | State management | ✅ Integrated |
| golang.org/x/image | v0.39.0 | Inter font (standard) | ✅ Integrated |

**Indirect:** go-text/typesetting v0.3.4, gogpu/gputypes v0.5.0, gogpu/wgpu v0.28.1, gogpu/naga v0.17.14, goffi v0.5.1, golang.org/x/text v0.37.0

---

## Success Criteria

### Performance
- 60fps with 10,000 widgets
- <100ms startup time
- <1KB memory per widget
- 0% GPU on static UI ✅

### Quality
- 80%+ test coverage (current: 97%+)
- WCAG 2.1 AA compliance
- Zero known critical bugs

### Ecosystem
- 20+ example applications
- Complete API documentation
- Migration guides from Fyne/Gio

---

## Links

| Resource | URL |
|----------|-----|
| gogpu Organization | https://github.com/gogpu |
| UI Repository | https://github.com/gogpu/ui |
| Discussions | https://github.com/orgs/gogpu/discussions/18 |
| Kanban Tasks | `docs/dev/kanban/` |
| Research | `docs/dev/research/` |

---

*This roadmap is updated as the project evolves.*
