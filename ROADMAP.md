# gogpu/ui Roadmap

> **Version:** 0.4.x (Phase 3 Complete, Phase 4 Near Complete)
> **Updated:** March 2026
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
- Three design systems: Material 3, Fluent, Cupertino

---

## Current Status

| Metric | Value |
|--------|-------|
| Packages | 55+ |
| Go Source Files | ~350 |
| Test Files | ~151 |
| Total LOC | ~150,000 |
| Test Functions | ~6,000 |
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
│  theme/material3   │  theme/fluent     │  theme/cupertino   │
│  (Complete ✅)     │  (Complete ✅)    │  (Complete ✅)     │
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
| Performance Benchmarks | 36 benchmarks across 5 packages |
| Task Manager Example | Full-featured demo with charts, tables, animations |
| Widget Gallery Example | All 22 widgets, 3 design systems, theme switching |
| Hover Tracking | W3C PointerEventSource, HoverTracker, cursor management |
| ScreenBounds | Screen-space coordinate transform for overlay positioning |
| Event Coordinate Transform | ScrollView mouse/wheel coordinate transforms |
| Inter Font Unicode | Full Unicode Inter 4.1 (Cyrillic, Greek, Vietnamese) |

**Remaining:**

| Task | Description | Priority |
|------|-------------|----------|
| Accessibility adapters | Platform-specific AT-SPI / UIA adapters | P1 |
| Documentation polish | Comprehensive API docs and guides | P2 |
| Performance optimization | Profiling, hot path optimization | P2 |
| API review | Pre-release API audit and freeze | P0 |

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
| gogpu/gg | v0.37.0 | 2D rendering + scene.Scene | ✅ Integrated |
| gogpu/gpucontext | v0.10.0 | Shared interfaces | ✅ Integrated |
| gogpu/gogpu | v0.24.1 | Windowing (examples) | ✅ Integrated |
| coregx/signals | v0.1.0 | State management | ✅ Integrated |
| golang.org/x/image | v0.37.0 | Inter font (standard) | ✅ Integrated |

**Indirect:** go-text/typesetting v0.3.4, gogpu/gputypes v0.3.0, gogpu/wgpu v0.21.0, gogpu/naga v0.14.7, golang.org/x/text v0.35.0

---

## Success Criteria

### Performance
- 60fps with 10,000 widgets
- <100ms startup time
- <1KB memory per widget

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
