# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
- Updated `github.com/gogpu/gg` v0.15.7 → v0.26.1

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
