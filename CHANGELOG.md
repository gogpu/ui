# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
- **Tasks Completed:** 5/6 (83%)

---

### Phase 0: Foundation Complete

Foundation packages implemented with enterprise-grade quality:

#### Added

- **geometry** — Core geometric types for UI layout
  - `Point`, `Size`, `Rect` with float32 components (GPU-compatible)
  - `Constraints` for constraint-based layout (Flutter-inspired)
  - `Insets` for padding/margin calculations
  - 100% test coverage

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
  - Color conversion utilities (widget.Color ↔ gg.RGBA)
  - `Renderer` for render cycle orchestration
  - `RenderTarget` interface with `SoftwareTarget` implementation
  - 96.5% test coverage

- **internal/layout** — Layout engine
  - `Engine` with caching and dirty tracking
  - `FlexContainer` — Full CSS Flexbox implementation
    - Direction: Row, RowReverse, Column, ColumnReverse
    - Justify: Start, End, Center, SpaceBetween, SpaceAround, SpaceEvenly
    - Align: Start, End, Center, Stretch, Baseline
    - flex-grow, flex-shrink, flex-basis support
  - `VStack`, `HStack`, `ZStack` — Simple stack layouts
  - `GridContainer` — Grid layout with auto/fixed/fractional tracks
  - 89.9% test coverage

#### Statistics

- **Total Lines of Code:** ~10,261
- **Test Coverage:** 95%+ average
- **Linter Issues:** 0

### Planned for v0.1.0 (Phase 1: MVP)

- [ ] Signals integration (coregx/signals)
- [ ] Basic primitives (Box, Text, Image)
- [ ] Public layout API
- [ ] Theme system foundation
- [ ] Window integration (gogpu/gogpu)

---

## Version History

| Version | Phase | Description |
|---------|-------|-------------|
| v0.1.0 | MVP | Core, layout, events, windowing |
| v0.2.0 | Beta | Widgets, Material 3 |
| v0.3.0 | RC | Virtualization, animation |
| v1.0.0 | Production | Enterprise features |

---

[Unreleased]: https://github.com/gogpu/ui/compare/main...HEAD
