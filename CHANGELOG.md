# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
