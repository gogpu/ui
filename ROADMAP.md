# gogpu/ui Roadmap

> **Version:** 0.2.x (Phase 2 In Progress)
> **Updated:** February 2026
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

---

## Current Status

### Phase 0: Foundation ✅ COMPLETE

| Package | Description | LOC | Coverage |
|---------|-------------|-----|----------|
| `geometry` | Point, Size, Rect, Constraints, Insets | ~800 | 100% |
| `event` | MouseEvent, KeyEvent, WheelEvent, Modifiers | ~600 | 100% |
| `widget` | Widget, WidgetBase, Context, Canvas, Color | ~2,956 | 100% |
| `internal/render` | Canvas implementation using gogpu/gg | ~1,740 | 96.5% |
| `internal/layout` | Flex, Stack, Grid layout engines | ~4,165 | 89.9% |
| **Total** | | **~10,261** | **95%+** |

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
v0.2.0  → Phase 2 Beta (In Progress)
v0.3.0  → Phase 3 RC
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
│  theme/material3   │  theme/fluent   │  theme/cupertino     │
│  (Complete ✅)     │   (Phase 4)     │    (Phase 4)         │
├─────────────────────────────────────────────────────────────┤
│  button/  focus/   │  docking/       │  animation/          │
│  (Complete ✅)     │  DockingHost    │  Animation, Spring  │
│  TextField (Next)  │  (Phase 4)      │  (Phase 3)           │
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
- ✅ geometry — Point, Size, Rect, Constraints, Insets
- ✅ event — MouseEvent, KeyEvent, WheelEvent, FocusEvent, Modifiers
- ✅ widget — Widget interface, WidgetBase, Context, Canvas, Color
- ✅ internal/render — Canvas implementation using gogpu/gg
- ✅ internal/layout — Engine, FlexContainer, VStack, HStack, ZStack, Grid

**Statistics:**
- ~10,261 lines of code
- 95%+ test coverage
- 0 linter issues

---

### Phase 1: MVP (v0.1.0) ✅ COMPLETE

**Goal:** Working foundation with basic widgets

**Tasks (10 tasks, ~12K LOC):**

| Task | Description | Status | LOC |
|------|-------------|--------|-----|
| TASK-UI-001 | Core Widget Interface | ✅ Done (Phase 0) | — |
| TASK-UI-002 | Signals Integration | ✅ Done | ~800 |
| TASK-UI-003 | WidgetBase Composition | ✅ Done (Phase 0) | — |
| TASK-UI-004 | Basic Primitives (Box, Text, Image) | ✅ Done | ~1,200 |
| TASK-UI-005 | Stack Layout (VStack, HStack) | ✅ Done (Phase 0) | — |
| TASK-UI-006 | Flexbox Layout Engine | ✅ Done (Phase 0) | — |
| TASK-UI-007 | Event System | ✅ Done (Phase 0) | — |
| TASK-UI-008 | Theme System Foundation | ✅ Done | ~1,200 |
| TASK-UI-009 | Rendering Pipeline | ✅ Done (Phase 0) | — |
| TASK-UI-010 | Window Integration | ✅ Done | ~800 |

**Delivered:**
- Signals integration (coregx/signals)
- Basic primitives (Box, Text, Image)
- Public layout API
- Theme system foundation
- Window integration (app package via gpucontext interfaces)

---

### Phase 1.5: Extensibility Foundation (v0.1.x) ✅ COMPLETE

**Goal:** Enable community to create custom widgets, themes, and layouts

**Completed:** 2026-01-16 | **Total LOC:** ~9,200 | **Coverage:** 97%+

| Task | Description | Status | LOC |
|------|-------------|--------|-----|
| ~~TASK-UI-041~~ | Widget Registry | ✅ Done | ~1,340 |
| ~~TASK-UI-042~~ | ThemeExtension Interface | ✅ Done | ~760 |
| ~~TASK-UI-043~~ | Public Layout API | ✅ Done | ~3,720 |
| ~~TASK-UI-044~~ | Theme Registry | ✅ Done | ~1,180 |
| ~~TASK-UI-045~~ | Plugin System | ✅ Done | ~3,040 |
| ~~TASK-UI-046~~ | Community Extension Guidelines | ✅ Done | ~2,000 |

**Implemented Packages:**
- `registry/` — Widget factory registration (100% coverage)
- `layout/` — Public layout API with custom algorithms (89.5% coverage)
- `theme/` — Theme System + Extensions + Registry (100% coverage)
- `plugin/` — Plugin bundling with dependency resolution (99.4% coverage)

**Why Extensibility First?**
- Community can create extensions from v0.1.x
- Third-party widgets/themes before v1.0
- Ecosystem growth enables faster adoption

---

### Phase 2: Beta (v0.2.0) 🔄 IN PROGRESS

**Goal:** Complete widget library

**Tasks (10 tasks, ~10K LOC):**

| Task | Description | Status | LOC |
|------|-------------|--------|-----|
| ~~TASK-UI-011~~ | Button Widget | ✅ Done | ~2,400 |
| TASK-UI-012 | TextField Widget | 📋 Pending | 1,200 |
| ~~TASK-UI-013~~ | Checkbox & Radio | ✅ Done | ~2,000 |
| TASK-UI-014 | Dropdown/Select | 📋 Pending | 900 |
| TASK-UI-015 | Slider Widget | 📋 Pending | 500 |
| TASK-UI-016 | Progress Indicators | 📋 Pending | 400 |
| ~~TASK-UI-017~~ | Material 3 Theme | ✅ Done | ~1,800 |
| TASK-UI-018 | Typography System | 📋 Pending | 600 |
| TASK-UI-019 | Icon System | 📋 Pending | 400 |
| ~~TASK-UI-020~~ | Keyboard Navigation (Focus) | ✅ Done | ~1,600 |

**Implemented Packages:**
- `core/button/` — Interactive button widget, 4 variants, 3 sizes (96%+ coverage)
- `core/checkbox/` — Toggleable checkbox with checked/unchecked/indeterminate (96%+ coverage)
- `core/radio/` — Radio group with vertical/horizontal layout, arrow key navigation (96%+ coverage)
- `focus/` — Keyboard focus management with Tab/Shift+Tab (95.2% coverage)
- `internal/focus/` — Internal focus manager implementation
- `theme/material3/` — Material Design 3 with HCT color science (97%+ coverage)
- `widget/focusable.go` — Focusable interface (IsFocusable, SetFocused, IsFocused)

**Remaining Deliverables:**
- TextField, Dropdown, Slider widgets
- Progress indicators
- Typography and icon systems

---

### Phase 3: RC (v0.3.0)

**Goal:** Enterprise features

**Tasks (10 tasks, ~10K LOC):**

| Task | Description | LOC |
|------|-------------|-----|
| TASK-UI-021 | VirtualizedList | 1,200 |
| TASK-UI-022 | VirtualizedGrid | 800 |
| TASK-UI-023 | Grid Layout Engine | ✅ Done (Phase 0) |
| TASK-UI-024 | Animation Engine | 1,000 |
| TASK-UI-025 | Transitions | 600 |
| TASK-UI-026 | Dialog/Modal | 700 |
| TASK-UI-027 | Popover/Tooltip | 600 |
| TASK-UI-028 | ScrollView | 600 |
| TASK-UI-029 | TabView | 500 |
| TASK-UI-030 | SplitView | 400 |

**Deliverables:**
- Virtualization for large datasets
- Animation system
- Complex layouts

---

### Phase 4: v1.0

**Goal:** Production-ready enterprise library

**Tasks (10 tasks, ~23K LOC):**

| Task | Description | LOC |
|------|-------------|-----|
| TASK-UI-031 | Docking System | 2,500 |
| TASK-UI-032 | Drag & Drop | 800 |
| TASK-UI-033 | Accessibility (A11y) - Pure Go AccessKit | 2,200 |
| TASK-UI-034 | Internationalization (i18n) | 600 |
| TASK-UI-035 | Fluent Theme | 1,000 |
| TASK-UI-036 | Cupertino Theme | 1,000 |
| TASK-UI-037 | Testing Utilities | 800 |
| TASK-UI-038 | Documentation | 10,000 |
| TASK-UI-039 | Examples | 3,000 |
| TASK-UI-040 | Performance Optimization | 1,500 |

**Deliverables:**
- IDE-style docking
- WCAG 2.1 AA compliance
- Multi-language support
- 3 theme presets
- Comprehensive docs

---

## Total Scope

| Phase | Tasks | Estimated LOC | Status |
|-------|-------|---------------|--------|
| Phase 0 (Foundation) | 5 packages | ~10K | ✅ Complete |
| Phase 1 (MVP) | 10 | ~12K | ✅ Complete |
| Phase 1.5 (Extensibility) | 6 | ~9K | ✅ Complete |
| Phase 2 (Beta) | 10 | ~10K | 🔄 In Progress (4/10) |
| Phase 3 (RC) | 10 | ~8K | Planned |
| Phase 4 (v1.0) | 10 | ~24K | Planned |
| **Total** | **51+** | **~73K LOC** | |

---

## Dependencies

| Dependency | Version | Purpose | Status |
|------------|---------|---------|--------|
| gogpu/gg | v0.28.1+ | 2D rendering | ✅ Integrated |
| gogpu/gogpu | v0.18.1+ | Windowing (event-driven) | ✅ Integrated |
| gogpu/gpucontext | v0.9.0+ | Shared interfaces | ✅ Integrated |
| gogpu/wgpu | v0.16.0+ | WebGPU backend | Via gg |
| coregx/signals | v0.1.0+ | State management | ✅ Integrated |

---

## Success Criteria

### Performance
- 60fps with 10,000 widgets
- <100ms startup time
- <1KB memory per widget

### Quality
- 80%+ test coverage (current: 95%+)
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
