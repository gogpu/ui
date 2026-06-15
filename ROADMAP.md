# gogpu/ui Roadmap

> **Version:** 0.1.30-dev
> **Updated:** June 2026
> **Go Version:** 1.25+

---

## Vision

**gogpu/ui** is the first enterprise-grade GUI toolkit for Go — zero CGO, GPU-accelerated, signal-driven.

Go has waited 17 years for a professional graphics ecosystem. We're building it: 1.1M+ LOC across the gogpu ecosystem, all platforms, native menus and dialogs, triple-backend WebGPU, shader compiler, and a complete GUI toolkit.

**Target applications:**
- **IDEs** — GoLand/VS Code class (docking, tabs, tree, toolbar, menus, code editor)
- **Design tools** — Photoshop/Figma class (canvas, layers, GPU compute)
- **CAD/scientific** — 3D viewport, data visualization, GPU-accelerated compute
- **Professional dashboards** — real-time charts, tables, data grids
- **Desktop apps** — Chrome/Electron replacement with native performance

**Key differentiators:**
- Pure Go by default (zero CGO), Rust backend optional via `-tags rust` (ADR-038 triple-backend)
- WebGPU-first rendering via gogpu/wgpu (Vulkan/Metal/DX12/GLES/Software/Browser)
- Signals-based reactive state (coregx/signals — hybrid push-pull, zero glitch)
- Layer Tree compositor with damage-aware blit (Flutter/Chrome patterns)
- Four design systems: Material 3, DevTools (JetBrains), Fluent, Cupertino
- Polymorphic Content[C] pattern (CDK — inspired by taiga-family/polymorpheus)
- Pluggable Painter architecture — design-system-agnostic widgets
- Enterprise features: docking, virtualization, accessibility, i18n, drag & drop

---

## Current Status

| Metric | Value |
|--------|-------|
| Packages | 56+ |
| Go Source Files | ~413 |
| Test Files | ~202 |
| Total LOC | ~198,000+ |
| Test Functions | ~7,300+ |
| Test Coverage | 97%+ |
| Linter Issues | 0 |
| Interactive Widgets | 22 |
| Design Systems | 4 (M3, DevTools, Fluent, Cupertino) |
| Painters | 61 (21 + 22 + 9 + 9) |

---

## Versioning Strategy

### Core Principle: Stay on v0.x.x

```
v0.x.x  → Active development (current — breaking changes OK)
v1.0.0  → ONLY when API stable for 1+ year (target: Dec 2026)
v2.0.0  → AVOID (requires /v2 import path)
```

### Version Progression:

```
v0.0.x  → Phase 0 Foundation               ✅ COMPLETE
v0.1.0  → Phase 1 MVP                      ✅ COMPLETE (Mar 2026)
v0.1.x  → Phase 1.5 Extensibility          ✅ COMPLETE
v0.2.0  → Phase 2 Beta                     ✅ COMPLETE
v0.2.x  → Phase 2.5 Signals Integration    ✅ COMPLETE
v0.3.0  → Phase 3 RC                       ✅ COMPLETE
v0.4.0  → Phase 4 v1.0 features            IN PROGRESS (~90%)
v0.9.0  → Pre-1.0 API freeze
v0.10+  → Stabilization
v1.0.0  → Production (target: Dec 2026)
```

### API Compatibility Patterns:

| Pattern | Purpose |
|---------|---------|
| **Functional Options** | Extend API without breaking changes |
| **Interface Extension** | Optional capabilities via type assertion |
| **Config Structs** | New fields with zero-value defaults |
| **internal/** | Implementation details (can change) |
| **Pluggable Painters** | Design system independence |
| **Content[C]** | Polymorphic content rendering |

---

## Completed Phases

### Phase 0: Foundation ✅

Core packages: geometry, event, widget, internal/render, internal/layout.

### Phase 1: MVP (v0.1.0) ✅

Signals integration, basic primitives (Box, Text, Image), public layout API, theme system, window integration.

### Phase 1.5: Extensibility ✅

Widget registry, public layout API, theme system + extensions, plugin system with dependency resolution.

### Phase 2: Beta ✅

Interactive widgets (button, checkbox, radio, textfield, dropdown), overlay infrastructure, focus management, Material Design 3 (HCT color science, 21 painters), CDK Content[C] pattern, ThemeScope.

### Phase 2.5: Signals Integration ✅

Push-based reactive state for all widgets. 4-level priority (ReadonlySignal > Signal > Fn > Static). Two-way bindings for stateful widgets.

### Phase 3: Release Candidate ✅

Slider, Dialog, Animation engine (Tween, Spring, M3 motion), ScrollView, TabView, ListView (virtualized), GridView, LineChart, ProgressBar, Collapsible, SplitView, Popover/Tooltip, Transitions, Dirty region tracking.

### Phase 4: Production Features — In Progress (~90%)

**Completed:**

| Feature | Description |
|---------|-------------|
| Circular progress | Determinate arc + indeterminate spinner |
| TreeView | Hierarchical, expand/collapse, virtualized |
| DataTable | Sortable columns, fixed header, virtualized rows |
| Toolbar | Icon buttons, separators, spacers |
| Menu | MenuBar + ContextMenu, submenus, shortcuts |
| Docking | IDE-style panels, border layout, tabbed groups |
| Drag & Drop | DragSource, DropTarget, Manager |
| DevTools theme | JetBrains Int UI — 22 painters, dark/light |
| Fluent theme | Microsoft Fluent Design — 9 painters |
| Cupertino theme | Apple HIG — 9 painters |
| Font registry | CSS weight matching (W3C spec) |
| Icon system | SVG icons, 2-level cache, DPI-aware |
| i18n | Locale, CLDR plural rules, RTL, bundles |
| Offscreen renderer | Headless widget → *image.RGBA |
| Layer Tree compositor | Flutter pipeline (ADR-007) |
| Per-boundary GPU textures | MSAA offscreen, DrawChild skip |
| Persistent Layer Tree | 97.9% fewer allocs |
| O(1) frame skip | Flat dirty set, 0% GPU idle |
| Multi-rect damage | Per-draw scissor, LoadOpLoad |
| Overlay boundary pipeline | Dropdown/dialog via Layer Tree |
| Custom font pipeline | FontRegistry, StyledTextDrawer |
| PointerCapturer | ADR-031, widget-level mouse capture |
| 34 integration tests | Multi-frame lifecycle, visibility matrix |

**Remaining Phase 4:**

| Task | Priority | Status |
|------|----------|--------|
| GPU spinner <3% | P0 | scheduler.SetOnDirty lifecycle |
| ListView hover rebuild | P1 | Painter pattern: hover = repaint only |
| Texture GC | P1 | Prune orphaned boundaryTextures |
| API review + freeze | P0 | Pre-1.0 audit |

---

## Future Roadmap

### Phase 5: New Widgets (v0.5.x — Q3 2026)

Essential widgets for production applications.

| Widget | Description | Complexity | Use Case |
|--------|-------------|------------|----------|
| **RichText** | Styled text with bold/italic/links, inline formatting | Medium | Content display, help text |
| **NumberField** | Numeric input: spinner buttons, range, step | Low | Forms, settings |
| **ToggleSwitch** | iOS/Material on/off switch with animation | Low | Settings, preferences |
| **Badge** | Notification badge (dot or count) on any widget | Low | Navigation, status |
| **Chip** | Filter/action chips (M3 spec) | Low | Tags, filters |
| **SegmentedControl** | Toggle button group (iOS/Fluent style) | Medium | View switching |
| **SearchField** | Text input with search icon, clear, suggestions | Medium | Data filtering |

### Phase 6: Advanced Widgets (v0.6.x — Q4 2026)

Complex widgets for professional applications.

| Widget | Description | Complexity | Use Case |
|--------|-------------|------------|----------|
| **DatePicker** | Calendar popup, date ranges, locale-aware | High | Forms, scheduling |
| **TimePicker** | Hour/minute selection, AM/PM, 24h | Medium | Scheduling |
| **ColorPicker** | Color wheel/palette, HSL/RGB, opacity | High | Design tools |
| **Accordion** | Mutually exclusive collapsible sections | Low | Settings, FAQ |
| **Breadcrumb** | Navigation trail with separators | Low | File browser, navigation |
| **Stepper** | Multi-step wizard with progress | Medium | Onboarding, forms |
| **Sheet** | Bottom/side sheet overlay (M3 spec) | Medium | Mobile-style panels |
| **NavigationRail** | Vertical navigation (M3 spec) | Medium | App navigation |

### Phase 7: IDE & Professional Widgets (v0.7.x — Q1 2027)

Widgets that enable building professional tools.

| Widget | Description | Complexity | Use Case |
|--------|-------------|------------|----------|
| **CodeEditor** | Syntax-highlighted editing, PieceTable, GPU text | Very High | IDEs, config editors |
| **Terminal** | Terminal emulator widget, ANSI codes | Very High | IDE terminal, DevOps |
| **Canvas** | User-controlled drawing surface, pan/zoom | Medium | Design tools, diagrams |
| **Carousel** | Horizontal scroll with snap points | Medium | Image galleries |
| **VirtualTable** | DataTable + million-row virtualization | High | Data analysis, logs |

The Code Editor is being designed as a separate `gogpu/editor` module (ADR-028) with PieceTable, GPU text rendering, and syntax highlighting. Enterprise references: VS Code PieceTree, Monaco MVVM, Scintilla, Xi-editor Rope, Zed GPUI, CodeMirror 6, cosmic-text.

### Phase 8: Platform Integration (v0.8.x — Q1-Q2 2027)

Platform-specific features for native feel.

| Feature | Description | Priority |
|---------|-------------|----------|
| **Accessibility adapters** | Windows UIA, Linux AT-SPI2, macOS NSAccessibility | P1 |
| **System theme detection** | Auto light/dark switching from OS | P1 |
| **Native file dialogs** | Open/Save/Folder via system dialogs | P1 |
| **Clipboard rich content** | HTML/RTF clipboard support | P2 |
| **IME support** | Input method for CJK languages | P2 |
| **Touch/gesture input** | Pinch, swipe, long press | P2 |

### Phase 9: API Freeze & Stabilization (v0.9.x — Q2-Q3 2027)

| Task | Description |
|------|-------------|
| API audit | Review every public type, method, option |
| Breaking change sweep | Last chance for naming/signature fixes |
| Migration guide | v0.x → v1.0 upgrade path |
| Documentation polish | Complete godoc, tutorials, cookbook |
| Performance profiling | Memory, CPU, GPU benchmarks per widget |
| Fuzz testing | Edge cases in layout, event dispatch, signals |

### v1.0.0 — Production Release (Target: Dec 2026 → Hard Deadline: Nov 2027)

**Success criteria:**
- API stable for 6+ months without breaking changes
- 30+ widgets with all 4 design system painters
- WCAG 2.1 AA accessibility compliance
- 60fps with 10,000 widgets
- <100ms startup time
- Complete documentation and migration guides
- Listed in awesome-go ✅ (already achieved)

---

## Rendering Performance Roadmap (ADR-007)

> **Architecture:** Hybrid CPU+GPU — industry standard (Chrome/Skia, Flutter, GTK4, Qt).
> CPU text atlas + GPU shapes + GPU compositor. Validated by source-level analysis of 8 engines.

### Current Performance (Intel Iris Xe, v0.1.29)

| Metric | Before (v0.1.14) | Current |
|--------|-------------------|---------|
| GPU (static UI, no animations) | 8% | **0%** |
| GPU (spinner visible, 30fps) | 8% | **10%** |
| GPU (spinner offscreen) | 8% | **0%** |
| GPU readback per frame | 0 | 0 |
| Render passes (idle) | 1 | **0** (frame skip) |
| Layer allocs per frame (200 boundaries) | 613 | **13** (persistent tree) |

### Completed Rendering Phases

| Phase | What | Status |
|-------|------|--------|
| Phase 1 | Zero-readback compositor (FlushPixmap, FlushGPUWithView) | ✅ |
| Phase 2 | Scene composition (RepaintBoundary, GPU SDF, granular invalidation) | ✅ |
| Phase 3 | Per-boundary GPU textures (MSAA offscreen, DrawChild skip) | ✅ |
| Phase 4 | Layer Tree + Damage-aware blit (persistent tree, multi-rect scissor, LoadOpLoad) | ✅ |

### Future Rendering

| Phase | What | Target |
|-------|------|--------|
| Phase 5 | Spinner GPU <3% (scheduler.SetOnDirty lifecycle) | v0.4.x |
| Phase 6 | Vello compute GPU path rendering (9-stage compute pipeline) | v0.7.x |
| Phase 7 | Partial present (VK_KHR_incremental_present, DX12 partial swap) | v0.8.x |

### Performance Targets

| Metric | Current | v1.0 Target |
|--------|---------|-------------|
| GPU % (static UI) | **0%** | 0% |
| GPU % (spinner) | 10% | <3% |
| GPU % (spinner offscreen) | **0%** | 0% |
| Startup time | ~200ms | <100ms |
| Memory per widget | ~2KB | <1KB |
| 10K widgets @ 60fps | — | Target |

---

## Ecosystem Integration Roadmap

gogpu/ui is one part of a larger ecosystem. Future integration points:

| Integration | Description | Timeline |
|-------------|-------------|----------|
| **gogpu/compute** | GPU compute via ComputeProvider (Born ML pattern) | Q3 2026 |
| **gogpu/editor** | Native code editor widget (ADR-028) | Q4 2026 |
| **gogpu/g3d** | 3D viewport widget for CAD/games | 2027 |
| **Browser/WASM** | Run ui in browser via wgpu Browser backend (ADR-038) | Q4 2026 |
| **compose** | Multi-process widget composition (Unix socket, hot-plug) | Available now |
| **Born ML** | ML model inference results in ui widgets | Available now |

### Cascade Release Order

```
naga (shader compiler)
  → wgpu (WebGPU HAL)
    → gpucontext (interfaces)
      → gogpu (windowing) + gg (2D graphics)
        → ui (GUI toolkit)
```

All releases must follow this cascade. Breaking changes in lower layers require coordinated releases.

---

## Design Philosophy

### What We Build On

| Pattern | Source | Our Implementation |
|---------|--------|-------------------|
| Layer Tree compositor | Flutter, Chrome, Qt6, Android | `compositor/` package |
| Pluggable Painters | All design systems (Swing L&F, Qt styles) | Painter interfaces per widget |
| Polymorphic Content[C] | taiga-family/polymorpheus | `cdk/` package |
| Signal-driven reactivity | Angular Signals, SolidJS, Preact | `state/` + coregx/signals |
| Functional Options | Go community best practice | All widget constructors |
| RepaintBoundary | Flutter RenderObject.isRepaintBoundary | `widget.WidgetBase` property |
| Damage-aware blit | Chrome DamageTracker, Wayland damage | `desktop/` + gg + wgpu stack |

### What We Don't Do

- **No webview** — native GPU rendering, not HTML/CSS/JS
- **No CGO** — pure Go, compiles on any platform Go supports
- **No runtime code generation** — all types resolved at compile time
- **No global state** — instance-based (Scheduler, FocusManager, App)
- **No implicit side effects** — explicit lifecycle (Mount/Unmount)
- **No backend abstraction in ui** — rendering is always gg → wgpu (ADR-009)

---

## Dependencies

| Dependency | Version | Purpose |
|------------|---------|---------|
| gogpu/gg | v0.48.11 | 2D rendering + scene.Scene |
| gogpu/gogpu | v0.42.0 | Windowing, input (examples) |
| gogpu/gpucontext | v0.21.0 | Shared interfaces (opaque struct tokens) |
| coregx/signals | v0.1.0 | Reactive state management |
| golang.org/x/image | v0.41.0 | Inter font (standard) |

**Indirect:** gogpu/wgpu v0.30.1, gogpu/naga v0.17.15, gogpu/gputypes v0.5.0, go-text/typesetting v0.3.4

---

## Community & Contributions

### How to Contribute

- **Test** — run examples on different GPUs and platforms, report issues
- **API feedback** — suggest improvements to widget APIs
- **Widgets** — implement new widgets following the Painter pattern
- **Design systems** — create painters for your design system
- **Documentation** — improve godoc, write tutorials
- **Spread the word** — articles, talks, social media

### Community Projects Using ui

| Project | Author | Description |
|---------|--------|-------------|
| KiGo | @AgentNemo00 | Visual programming tool |
| PupSeek IDE | private | AI-powered IDE |
| Petri Net IDE | @paulie-g | Process modeling tool |
| f4 | @unxed | Text editor |

---

## Links

| Resource | URL |
|----------|-----|
| gogpu Organization | https://github.com/gogpu |
| UI Repository | https://github.com/gogpu/ui |
| Discussions | https://github.com/orgs/gogpu/discussions/18 |
| awesome-go listing | https://github.com/avelino/awesome-go |
| Kanban Tasks | `docs/dev/kanban/` |
| Research | `docs/dev/research/` |
| ADRs | `docs/dev/architecture/` |

---

*This roadmap evolves with the project. Last updated: June 2026.*
