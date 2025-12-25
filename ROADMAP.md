# gogpu/ui Roadmap

> **Version:** 0.0.0 (Planning)
> **Updated:** December 2025
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

## Versioning Strategy

### Core Principle: Stay on v0.x.x

```
v0.x.x  → Active development (current)
v1.0.0  → ONLY when API stable for 1+ year
v2.0.0  → AVOID (requires /v2 import path)
```

### Version Progression:

```
v0.1.0  → Phase 1 MVP
v0.2.0  → Phase 2 Beta
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
│    (Optional)      │   (Optional)    │    (Optional)        │
├─────────────────────────────────────────────────────────────┤
│  widgets/         │  docking/        │  animation/          │
│  Button, TextField│  DockingHost     │  Animation, Spring   │
│  Dropdown, etc.   │  FloatingWindow  │  Transitions         │
├─────────────────────────────────────────────────────────────┤
│  layout/                            │  state/               │
│  VStack, HStack, Grid, Flexbox      │  coregx/signals       │
├─────────────────────────────────────────────────────────────┤
│  core/                              │  event/               │
│  Widget, WidgetBase, Context        │  Mouse, Keyboard      │
├─────────────────────────────────────────────────────────────┤
│  render/                            │  typography/          │
│  Canvas, Renderer                   │  Font, TextStyle      │
├─────────────────────────────────────────────────────────────┤
│  gogpu/gg          │  gogpu/gogpu    │  coregx/signals      │
│  2D Graphics       │  Windowing      │  State Management    │
└─────────────────────────────────────────────────────────────┘
```

---

## Phases

### Phase 1: MVP (v0.1.0)

**Goal:** Working foundation with basic widgets

**Tasks (10 tasks, ~12K LOC):**

| Task | Description | LOC |
|------|-------------|-----|
| TASK-UI-001 | Core Widget Interface | 1,500 |
| TASK-UI-002 | Signals Integration | 800 |
| TASK-UI-003 | WidgetBase Composition | 600 |
| TASK-UI-004 | Basic Primitives (Box, Text, Image) | 1,200 |
| TASK-UI-005 | Stack Layout (VStack, HStack) | 800 |
| TASK-UI-006 | Flexbox Layout Engine | 2,500 |
| TASK-UI-007 | Event System | 1,000 |
| TASK-UI-008 | Theme System Foundation | 1,200 |
| TASK-UI-009 | Rendering Pipeline | 1,500 |
| TASK-UI-010 | Window Integration | 800 |

**Deliverables:**
- Working hello world app
- Basic layout system
- Event handling
- Theme support

---

### Phase 2: Beta (v0.2.0)

**Goal:** Complete widget library

**Tasks (10 tasks, ~10K LOC):**

| Task | Description | LOC |
|------|-------------|-----|
| TASK-UI-011 | Button Widget | 600 |
| TASK-UI-012 | TextField Widget | 1,200 |
| TASK-UI-013 | Checkbox & Radio | 600 |
| TASK-UI-014 | Dropdown/Select | 900 |
| TASK-UI-015 | Slider Widget | 500 |
| TASK-UI-016 | Progress Indicators | 400 |
| TASK-UI-017 | Material 3 Theme | 1,500 |
| TASK-UI-018 | Typography System | 600 |
| TASK-UI-019 | Icon System | 400 |
| TASK-UI-020 | Keyboard Navigation | 700 |

**Deliverables:**
- All standard widgets
- Material 3 theme
- Full keyboard support

---

### Phase 3: RC (v0.3.0)

**Goal:** Enterprise features

**Tasks (10 tasks, ~10K LOC):**

| Task | Description | LOC |
|------|-------------|-----|
| TASK-UI-021 | VirtualizedList | 1,200 |
| TASK-UI-022 | VirtualizedGrid | 800 |
| TASK-UI-023 | Grid Layout Engine | 1,500 |
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

| Phase | Tasks | Estimated LOC |
|-------|-------|---------------|
| Phase 1 (MVP) | 10 | ~12K |
| Phase 2 (Beta) | 10 | ~10K |
| Phase 3 (RC) | 10 | ~10K |
| Phase 4 (v1.0) | 10 | ~24K |
| **Total** | **40** | **~56K LOC** |

---

## Dependencies

| Dependency | Version | Purpose |
|------------|---------|---------|
| gogpu/gg | v0.13.0+ | 2D rendering |
| gogpu/gogpu | v0.8.0+ | Windowing |
| gogpu/wgpu | v0.7.0+ | WebGPU backend |
| coregx/signals | v0.1.0+ | State management |

---

## Success Criteria

### Performance
- 60fps with 10,000 widgets
- <100ms startup time
- <1KB memory per widget

### Quality
- 80%+ test coverage
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
| Kanban Tasks | `docs/dev/kanban/` |
| Research | `docs/dev/research/` |

---

*This roadmap is updated as the project evolves.*
