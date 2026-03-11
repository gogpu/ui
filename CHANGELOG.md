# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
