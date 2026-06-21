<p align="center">
  <img src="https://raw.githubusercontent.com/gogpu/.github/main/assets/logo.png" alt="GoGPU Logo" width="120" />
</p>

<h1 align="center">gogpu/ui</h1>

<p align="center">
  <strong>Enterprise-Grade GUI Toolkit for Go</strong><br>
  Modern widgets, reactive state, GPU-accelerated rendering — zero CGO
</p>

<p align="center">
  <a href="https://github.com/gogpu/ui/actions"><img src="https://github.com/gogpu/ui/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://app.codecov.io/gh/gogpu/ui"><img src="https://codecov.io/gh/gogpu/ui/branch/main/graph/badge.svg" alt="Coverage"></a>
  <a href="https://goreportcard.com/report/github.com/gogpu/ui"><img src="https://goreportcard.com/badge/github.com/gogpu/ui" alt="Go Report Card"></a>
  <a href="https://pkg.go.dev/github.com/gogpu/ui"><img src="https://pkg.go.dev/badge/github.com/gogpu/ui.svg" alt="Go Reference"></a>
  <a href="https://github.com/gogpu/ui/releases/latest"><img src="https://img.shields.io/github/v/release/gogpu/ui?style=flat&label=version" alt="Version"></a>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go" alt="Go Version"></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License"></a>
  <a href="https://github.com/gogpu/ui/stargazers"><img src="https://img.shields.io/github/stars/gogpu/ui?style=flat&labelColor=555&color=yellow" alt="Stars"></a>
  <a href="https://github.com/orgs/gogpu/discussions"><img src="https://img.shields.io/badge/Discussions-join-blue" alt="Discussions"></a>
</p>

---

> **Join the Discussion:** Help shape the future of Go GUI! Share your ideas, report issues, and discuss features at our [GitHub Discussions](https://github.com/orgs/gogpu/discussions/18).

---

## Overview

**gogpu/ui** is an enterprise-grade GUI toolkit for Go, designed for building:

- **IDEs** (GoLand, VS Code class)
- **Design Tools** (Photoshop, Figma class)
- **CAD Applications**
- **Professional Dashboards**
- **Chrome/Electron Replacement Apps**

### Key Differentiators

| Feature | gogpu/ui | Fyne | Gio |
|---------|----------|------|-----|
| **CGO-free** | Yes | No | Yes |
| **WebGPU rendering** | Yes | OpenGL | Direct GPU |
| **Reactive state** | Signals | Binding | Events |
| **Layout engine** | Flexbox + Grid | Custom | Flex |
| **Accessibility** | Day 1 (ARIA roles) | Limited | Limited |
| **Plugin system** | Yes | No | No |

---

## Screenshot

<p align="center">
  <img src="https://github.com/user-attachments/assets/f67636bc-7426-4865-9269-11e2a81124ad" alt="gogpu/ui Widget Demo" width="600" />
</p>

<p align="center"><em>examples/hello — Checkboxes, Radio Buttons, ListView (1000 items), Material Design 3</em></p>

---

## Quick Start

> **Important:** The gogpu ecosystem is **pure Go with zero CGO**. You must set `CGO_ENABLED=0` (the Go default) — do **not** enable CGO.

The fastest way to get started is to clone and run one of the included examples:

```bash
git clone https://github.com/gogpu/ui.git
cd ui/examples/hello
go run .
```

Here is a minimal example using `desktop.Run` (recommended):

```go
package main

import (
    "log"

    _ "github.com/gogpu/gg/gpu"
    "github.com/gogpu/gogpu"
    "github.com/gogpu/ui/app"
    "github.com/gogpu/ui/desktop"
    "github.com/gogpu/ui/primitives"
    "github.com/gogpu/ui/widget"
)

func main() {
    gogpuApp := gogpu.NewApp(gogpu.DefaultConfig().
        WithTitle("My App").
        WithSize(800, 600).
        WithContinuousRender(false)) // Event-driven: 0% CPU when idle

    uiApp := app.New(
        app.WithWindowProvider(gogpuApp),
        app.WithPlatformProvider(gogpuApp),
        app.WithEventSource(gogpuApp.EventSource()),
    )

    uiApp.SetRoot(
        primitives.Box(
            primitives.Text("Hello gogpu/ui!").
                FontSize(24).Bold().
                Color(widget.RGBA8(33, 33, 33, 255)),
            primitives.Text("Enterprise-grade GUI for Go").
                FontSize(16).
                Color(widget.RGBA8(100, 100, 100, 255)),
        ).Padding(24).Gap(12).Background(widget.RGBA8(255, 255, 255, 255)),
    )

    if err := desktop.Run(gogpuApp, uiApp); err != nil {
        log.Fatal(err)
    }
}
```

---

## Packages

### Core (Phase 0)

| Package | Description | Coverage |
|---------|-------------|----------|
| `geometry` | Point, Size, Rect, Constraints, Insets | 98.8% |
| `event` | MouseEvent, KeyEvent, WheelEvent, FocusEvent, Modifiers | 100% |
| `widget` | Widget, WidgetBase, Context, Canvas, Lifecycle (mount/unmount), SchedulerRef | 100% |
| `internal/render` | Canvas, SceneCanvas, IconCache (2-level LRU), DPI-aware SVG | 96.5% |
| `internal/layout` | Flex, Stack, Grid layout engines | 89.9% |

### MVP (Phase 1)

| Package | Description | Coverage |
|---------|-------------|----------|
| `a11y` | Accessibility: 35+ ARIA roles, Accessible interface, Tree, Announcer | 99.1% |
| `state` | Reactive signals, Binding, Scheduler with push-based invalidation | 100% |
| `primitives` | Box, Text, Image widgets with fluent builder API | 94.4% |
| `app` | Window integration via gpucontext interfaces (dependency inversion) | 98.6% |

### Extensibility (Phase 1.5)

| Package | Description | Coverage |
|---------|-------------|----------|
| `layout` | Public layout API with custom algorithms | 89.5% |
| `registry` | Widget factory registration for third-party widgets | 100% |
| `theme` | Theme system with Extensions and Registry | 100% |
| `plugin` | Plugin bundling with dependency resolution | 99.4% |

### Interactive Widgets (Phase 2 — Complete)

| Package | Description | Coverage |
|---------|-------------|----------|
| `cdk` | Component Development Kit — Content[C] polymorphic pattern | 100% |
| `core/button` | Generic button with pluggable Painter, 4 variants, 3 sizes, signal bindings | 96%+ |
| `core/checkbox` | Toggleable checkbox with checked/unchecked/indeterminate states, signal bindings | 96%+ |
| `core/radio` | Mutually exclusive radio group with vertical/horizontal layout, signal bindings | 96%+ |
| `core/textfield` | Text input with cursor, selection, clipboard, validation, signal bindings | 96%+ |
| `core/slider` | Slider: continuous/discrete, horizontal/vertical, drag+keyboard, signal bindings | 94.6% |
| `core/dialog` | Modal dialog: backdrop overlay, action buttons, focus trapping, Alert/Confirm | 96.9% |
| `core/dropdown` | Dropdown/select with overlay menu, keyboard navigation, signal bindings | 96%+ |
| `overlay` | Overlay/popup stack, container, position helper | 95%+ |
| `primitives` | Box, Text, Image, RepaintBoundary (GPU texture caching via Layer Tree compositor) | 94.4% |
| `theme/material3` | Material Design 3 — theme (HCT color science) + 21 component painters | 97%+ |
| `focus` | Keyboard focus management with Tab/Shift+Tab navigation | 95.2% |
| `internal/focus` | Internal focus manager implementation | 15.2% |

### Phase 3 (Complete)

| Package | Description | Coverage |
|---------|-------------|----------|
| `animation` | Animation engine: tween, spring, M3 presets, orchestration (Stagger/Chain/Repeat/Reverse) | 92.3% |
| `transition` | Widget enter/exit animations: Fade, Slide, Scale effects, Show/Hide wrapper | 98.7% |
| `core/scrollview` | Scrollable container: vertical/horizontal/both, wheel+keyboard+drag, PushClip/PushTransform, signal bindings | 96.5% |
| `core/tabview` | Tabbed navigation: lazy content switching, closeable tabs, keyboard nav, Top/Bottom position, signal bindings | 92.1% |
| `core/listview` | Virtualized list: fixed-height items, recycling, single/multi selection, keyboard nav, M3 painter | 96%+ |
| `core/gridview` | Virtualized 2D grid: fixed cell size, auto-fit columns, cell recycling, selection, keyboard nav | 92.1% |
| `core/linechart` | Real-time line chart: multiple series, rolling window, grid, Y-axis labels, thread-safe PushValue | 98.8% |
| `core/progressbar` | Linear progress bar: 0-100%, rounded corners, label, signal binding, PushClipRoundRect | 99.3% |
| `core/collapsible` | Expandable section: animated expand/collapse, keyboard focus, arrow indicator, Tween animation | 98.2% |
| `primitives` | Box (HBox/VBox direction), Text, Image, ThemeScope, RepaintBoundary | 94%+ |

### Phase 4 (In Progress)

| Package | Description | Coverage |
|---------|-------------|----------|
| `core/progress` | Circular progress: determinate arc + indeterminate spinner, polyline approximation | 97.4% |
| `core/popover` | Popover (click) + Tooltip (hover): 12 placements, auto-flip, overlay integration | 97.1% |
| `core/splitview` | Resizable split panels: draggable divider, min constraints, collapse, handle dots | 96.8% |
| `core/treeview` | Hierarchical tree: expand/collapse, virtualized rendering, connector lines, keyboard nav | 96%+ |
| `core/datatable` | Sortable column table: fixed header, virtualized rows, zebra striping, sort indicators | 96%+ |
| `core/toolbar` | Horizontal action bar: icon buttons, separators, spacers, custom widget items | 96%+ |
| `core/menu` | MenuBar + ContextMenu: submenus, separators, disabled items, shortcut display | 96%+ |
| `core/docking` | IDE-style dockable panels: border layout, tabbed groups, Dock/Undock API | 95.3% |
| `core/badge` | Notification badge: count mode (pill, "99+" overflow), dot mode, signal bindings with dedup | 99.3% |
| `core/chip` | Action/filter chip (M3 spec): toggleable selection, two-way signal write-back, state layers | 99.0% |
| `theme/material3` | 21 component painters (all widgets covered) | 97%+ |
| `theme/devtools` | **JetBrains DevTools**: 22 painters, Int UI gray scale, dark/light, IDE-style | 96%+ |
| `theme/fluent` | Microsoft Fluent Design: 9 painters, accent colors, inner focus ring, light/dark | 96%+ |
| `theme/cupertino` | Apple HIG: 9 painters, iOS toggle switch, segmented control, pill buttons | 96%+ |
| `theme/font` | Font Registry: CSS weight matching (W3C spec), Weight 100-900, Style, Family/Face | 97.7% |
| `icon` | SVG icons (JetBrains expui), 2-level cache, DPI-aware rasterization, gg/svg renderer | 97%+ |
| `i18n` | Internationalization: Locale, Bundle, Translator, CLDR plural rules, RTL, LocaleSignal | 97.9% |
| `dnd` | Drag and drop: DragSource/DropTarget interfaces, Manager, 5px threshold, Escape cancel | 99.3% |
| `offscreen` | Headless widget rendering: CPU-only `*image.RGBA` output, no GPU/window/app required | 100% |
| `uitest` | Testing utilities: MockCanvas, MockContext, event factories, widget helpers, assertions | 93.1% |
| `internal/dirty` | Dirty region tracking: Collector, Tracker, merge algorithm, partial repaints | 100% |

| `compositor` | Layer Tree compositor: OffsetLayer, PictureLayer, ClipRectLayer, OpacityLayer — production render pipeline | 95%+ |

**Total: ~198,000+ lines of code | 56+ packages | ~7,300+ tests | 97%+ average coverage**

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    User Application                         │
├─────────────────────────────────────────────────────────────┤
│  theme/material3/  │  theme/devtools/ │  theme/fluent/      │
│  21 Painters       │  22 Painters     │  9 Painters         │
│  theme/cupertino/  │                  │                     │
│  9 Painters        │                  │                     │
├─────────────────────────────────────────────────────────────┤
│  24 Interactive Widgets (core/)                             │
│  button, checkbox, radio, textfield, dropdown, slider,      │
│  dialog, scrollview, tabview, listview, gridview, linechart,│
│  progressbar, progress, collapsible, popover, splitview,    │
│  treeview, datatable, toolbar, menu, docking                │
├──────────────┬──────────────────────────────────────────────┤
│  cdk/        │  Content[C] polymorphic pattern              │
├──────────────┴──────────────────────────────────────────────┤
│  primitives/       │  animation/     │  transition/         │
│  Box (HBox/VBox),  │  Tween, Spring  │  Fade, Slide, Scale  │
│  Text, Image,      │  M3 Presets,    │  Enter/Exit          │
│  RepaintBoundary   │  Orchestration  │                      │
├─────────────────────────────────────────────────────────────┤
│  icon/  │  i18n/  │  dnd/   │  uitest/ │  theme/font/       │
├─────────────────────────────────────────────────────────────┤
│  app/ + FocusManager │  focus/ │  overlay/ │  render/       │
├─────────────────────────────────────────────────────────────┤
│  desktop/            (Layer Tree Compositor + Damage-Aware Blit)    │
│  compositor/         (Production: OffsetLayer, PictureLayer, Opacity)│
│  offscreen/          (headless widget → *image.RGBA)        │
├─────────────────────────────────────────────────────────────┤
│  layout/           │  state/         │  a11y/               │
│  Flex, Stack, Grid │  Signals, Bind  │  ARIA Roles, Tree    │
├─────────────────────────────────────────────────────────────┤
│  registry/         │  plugin/        │  theme/              │
├─────────────────────────────────────────────────────────────┤
│  widget/           │  event/         │  geometry/           │
│  Widget, Context   │  Mouse, Key     │  Point, Rect         │
│  Canvas, Lifecycle │  Wheel, Focus   │  Constraints         │
├─────────────────────────────────────────────────────────────┤
│  internal/render   │  internal/layout│  internal/focus      │
│  Canvas, Scene,    │  Flex, Grid     │  Manager, Ring       │
│  IconCache (LRU)   │  internal/dirty │  Tracker, Collector  │
├─────────────────────────────────────────────────────────────┤
│  gogpu/gg          │  gpucontext     │  coregx/signals      │
│  2D Graphics       │  Shared Ifaces  │  State Management    │
└─────────────────────────────────────────────────────────────┘
```

### Dependency Principle

```
ui → gpucontext (interfaces)       ← dependency inversion
ui → gg (2D rendering)
ui → coregx/signals (reactive)

gogpu → gpucontext (implements)    ← concrete implementation
gg → wgpu → naga                   ← internal to gg
```

**ui never imports gogpu, wgpu, or naga directly.**

### Render Pipeline

Enterprise-grade retained-mode rendering (ADR-007):

1. **O(1) frame skip** -- flat dirty boundary set, no tree walks when idle (0% GPU)
2. **Layer Tree composition** -- OffsetLayer, PictureLayer, OpacityLayer, ClipRectLayer
3. **Per-boundary GPU textures** -- dirty boundaries re-render to MSAA offscreen texture, clean reuse cached
4. **Damage-aware blit** -- LoadOpLoad + multi-rect scissor, only dirty pixels touch the GPU
5. **Persistent tree** -- layer objects reused across frames (97.9% fewer allocations)

Validated by enterprise research: Flutter, Chrome, Qt6, Android, Skia patterns.
Software backend e2e tests prove scissor=48x48 at HAL level.

---

## Examples

| Example | Description |
|---------|-------------|
| [`examples/hello`](examples/hello) | Widget demo: checkbox, radio, ListView (1000 items), M3 theme, event-driven GPU rendering |
| [`examples/signals`](examples/signals) | Reactive signals: TextSignal, ContentSignal, CheckedSignal, SelectedSignal, DisabledSignal |
| [`examples/taskmanager`](examples/taskmanager) | Full task manager: charts, tables, animations, real-time data |
| [`examples/gallery`](examples/gallery) | Widget gallery: all 24 widgets, 4 design systems (M3/DevTools/Fluent/Cupertino), theme switching |
| [`examples/ide`](examples/ide) | GoLand-inspired IDE layout: DevTools theme, toolbar, tree, tabs, terminal, SVG icons |
| [`examples/modular-compositor`](examples/modular-compositor) | Multi-module offscreen rendering: clock + notification compositor ([#75](https://github.com/gogpu/ui/issues/75)) |

Run any example:

```bash
cd examples/gallery
go run .
```

---

## API Examples

### Primitives

```go
// Box — universal container
primitives.Box(children...).
    Padding(16).
    PaddingXY(24, 8).
    Background(theme.Surface).
    Rounded(8).
    BorderStyle(1, theme.Outline).
    ShadowLevel(2).
    Gap(8)

// Text — static
primitives.Text("Hello World").
    FontSize(24).
    Bold().
    Color(theme.OnSurface).
    Align(primitives.TextAlignCenter).
    MaxLines(2).
    Ellipsis()

// Text — reactive (auto-updates when signal changes)
primitives.TextFn(func() string {
    return fmt.Sprintf("Count: %d", count.Get())
}).FontSize(18)

// Text — signal binding (auto-updates when signal changes)
title := state.NewSignal("Hello World")
primitives.NewText("").ContentSignal(title).FontSize(24)

// Image
primitives.Image(mySource).
    Size(48, 48).
    Cover().
    Rounded(24).
    Alt("User avatar")
```

### Slider

```go
// Basic slider
s := slider.New(
    slider.Min(0),
    slider.Max(100),
    slider.Value(50),
    slider.OnChange(func(v float32) { fmt.Println("Value:", v) }),
)

// Discrete slider with step snapping and marks
s := slider.New(
    slider.Min(0), slider.Max(100), slider.Step(10),
    slider.Marks([]slider.Mark{{Value: 0}, {Value: 50}, {Value: 100}}),
    slider.PainterOpt(material3.SliderPainter{Theme: m3}),
)

// Two-way signal binding
volume := state.NewSignal[float32](75)
s := slider.New(
    slider.Min(0), slider.Max(100),
    slider.ValueSignal(volume),  // reads AND writes back on drag
)
```

### Dialog

```go
// Simple alert dialog
d := dialog.Alert("Error", "File not found.", func() { log.Println("dismissed") })
d.Show(ctx)

// Confirmation dialog
d := dialog.Confirm("Delete?", "This cannot be undone.",
    func() { log.Println("canceled") },
    func() { deleteItem() },
)
d.Show(ctx)

// Custom dialog with M3 painter
d := dialog.New(
    dialog.Title("Settings"),
    dialog.Content(settingsWidget),
    dialog.Actions(
        dialog.Action{Label: "Cancel", Variant: dialog.VariantTextOnly},
        dialog.Action{Label: "Save", Variant: dialog.VariantFilled, Default: true,
            OnClick: func() { saveSettings() }},
    ),
    dialog.PainterOpt(material3.DialogPainter{Theme: m3}),
)
d.Show(ctx)
```

### Animation

```go
// Tween animation with M3 easing
ctrl := animation.NewController()
opacity := state.NewSignal[float32](0)
animation.To(opacity, 1.0).
    Duration(animation.DurationMedium2).
    Ease(animation.M3Standard).
    Start(ctrl)

// Spring physics (critically damped)
position := state.NewSignal[float32](0)
animation.SpringTo(position, 200).
    Stiffness(animation.StiffnessMedium).
    DampingRatio(animation.DampingNoBouncy).
    Start(ctrl)

// Color tween (Flutter pattern: float32 engine + Tween[T])
progress := state.NewSignal[float32](0)
animation.To(progress, 1.0).Duration(300*time.Millisecond).Start(ctrl)
colorTween := animation.NewColorTween(startColor, endColor)
// in Draw: canvas.DrawRect(bounds, colorTween.At(progress.Get()))

// Parallel composition
animation.Parallel(
    animation.To(opacity, 1.0).Duration(200*time.Millisecond),
    animation.To(translateY, 0).Duration(300*time.Millisecond),
).Start(ctrl)
```

### ScrollView

```go
// Basic vertical scrollview
sv := scrollview.New(longContentWidget,
    scrollview.DirectionOpt(scrollview.Vertical),
    scrollview.ScrollbarOpt(scrollview.ScrollbarAuto),
    scrollview.OnScroll(func(x, y float32) { fmt.Printf("scroll: %.0f\n", y) }),
)

// 2D scrollview with signal binding
scrollY := state.NewSignal[float32](0)
sv := scrollview.New(largeCanvas,
    scrollview.DirectionOpt(scrollview.Both),
    scrollview.ScrollYSignal(scrollY), // two-way binding
    scrollview.PainterOpt(material3.ScrollbarPainter{Theme: m3}),
)
```

### TabView

```go
// Basic tab view
tv := tabview.New([]tabview.Tab{
    {Label: "Profile", Content: profileWidget},
    {Label: "Settings", Content: settingsWidget},
    {Label: "About", Content: aboutWidget},
}, tabview.OnSelect(func(idx int) { fmt.Println("Tab:", idx) }))

// Closeable tabs with M3 painter and signal binding
selected := state.NewSignal[int](0)
tv := tabview.New(tabs,
    tabview.Closeable(true),
    tabview.SelectedSignalOpt(selected),
    tabview.PainterOpt(material3.TabViewPainter{Theme: m3}),
    tabview.OnClose(func(idx int) { removeDynamicTab(idx) }),
)
```

### Reactive State

```go
// Create a signal
name := state.NewSignal("World")

// Computed value (auto-updates)
greeting := state.NewComputed(func() string {
    return "Hello, " + name.Get() + "!"
})

// Bind signal to widget invalidation
binding := state.Bind(name, ctx)
defer binding.Unbind()

// Batch multiple changes (single re-render)
scheduler.Batch(func() {
    firstName.Set("Alice")
    lastName.Set("Smith")
    age.Set(30)
})
```

### Widget Signal Bindings

```go
// Bind signals directly to widget properties
label := state.NewSignal("Submit")
disabled := state.NewSignal(false)

btn := button.New(
    button.TextSignal(label),
    button.DisabledSignal(disabled),
    button.OnClick(func() {
        label.Set("Processing...")
        disabled.Set(true)
    }),
)

// Two-way binding: checkbox state synced with signal
agreed := state.NewSignal(false)
cb := checkbox.New(
    checkbox.CheckedSignal(agreed),
    checkbox.LabelOpt("I agree to terms"),
)
```

### Accessibility

```go
// Every widget implements a11y.Accessible
func (b *MyButton) AccessibilityRole() a11y.Role   { return a11y.RoleButton }
func (b *MyButton) AccessibilityLabel() string      { return b.text }
func (b *MyButton) AccessibilityActions() []a11y.Action {
    return []a11y.Action{a11y.ActionClick}
}

// Accessibility tree with stable node IDs
root := a11y.NewNode(a11y.RoleWindow, "My Application")
tree := a11y.NewMemoryTree(root)
button := a11y.NewNode(a11y.RoleButton, "Save")  // stable uint64 ID
tree.Insert(root, button)
```

### Offscreen Rendering

```go
// Render widgets to image without GPU, window, or app
r := offscreen.NewRenderer(400, 120)
r.Render(primitives.Text("Hello, World!").FontSize(24))
img := r.Image() // *image.RGBA — ready for png.Encode, testing, compositing

// HiDPI with dark theme and white background
dark := material3.NewDark(widget.Hex(0x00BFA5))
r := offscreen.NewRenderer(800, 240,
    offscreen.WithTheme(dark),
    offscreen.WithScale(2.0),
    offscreen.WithBackground(widget.ColorWhite),
)
r.Render(myWidgetTree)
```

### Window Integration

```go
// ui connects to windowing via interfaces (not concrete types)
uiApp := app.New(
    app.WithWindowProvider(gogpuApp),    // gpucontext.WindowProvider
    app.WithPlatformProvider(gogpuApp),  // gpucontext.PlatformProvider
    app.WithTheme(myTheme),
)

uiApp.SetRoot(rootWidget)

// Headless mode for testing (no window needed)
testApp := app.New()  // works without any providers
testApp.SetRoot(rootWidget)
testApp.Window().Frame()  // processes layout + draw
```

---

## Implementation Progress

### Phase 0: Foundation ✅

- [x] Geometry types (Point, Size, Rect, Constraints, Insets)
- [x] Event system (Mouse, Keyboard, Wheel, Focus, Modifiers)
- [x] Widget interface, WidgetBase, Context, Canvas
- [x] Layout engines (Flexbox, Stack, Grid)
- [x] Canvas implementation (gogpu/gg)

### Phase 1: MVP ✅

- [x] Accessibility foundation (35+ ARIA roles, Accessible interface, Tree)
- [x] Reactive signals integration (coregx/signals, Binding, Scheduler)
- [x] Basic primitives (Box, Text, Image with fluent API)
- [x] Window integration (app package via gpucontext interfaces)

### Phase 1.5: Extensibility ✅

- [x] Widget Registry (third-party registration)
- [x] Public Layout API (custom algorithms)
- [x] Theme System + Extensions + Registry
- [x] Plugin System (bundling, dependency resolution)

### Phase 2: Beta ✅

- [x] Button widget (4 variants, 3 sizes, keyboard activation)
- [x] Checkbox widget (checked/unchecked/indeterminate, pluggable Painter)
- [x] Radio group widget (vertical/horizontal, arrow key navigation)
- [x] TextField widget (cursor, selection, clipboard, validation)
- [x] Dropdown/Select widget (overlay menu, keyboard nav, scroll)
- [x] Overlay infrastructure (stack, container, position)
- [x] Material Design 3 theme (HCT color science, 21 painters)
- [x] Keyboard navigation (focus management, Tab/Shift+Tab, shortcuts)
- [x] ThemeScope (theme override for widget subtrees)
- [x] Event-driven rendering (0% CPU when idle)
- [x] Reactive signal bindings for all widgets (TextSignal, CheckedSignal, SelectedSignal, DisabledSignal, ContentSignal)

### Phase 3: Release Candidate ✅

- [x] Retained-mode rendering: dirty tracking, DrawTree, DrawStats (SP1)
- [x] RepaintBoundary: per-widget pixel caching (SP2)
- [x] scene.Scene integration: tile-parallel rendering via SceneCanvas (SP3)
- [x] Slider widget (continuous/discrete, horizontal/vertical, M3 painter)
- [x] Dialog/Modal widget (backdrop, actions, focus trapping, M3 painter)
- [x] Animation engine (Tween, Spring, CubicBezier, M3 motion tokens)
- [x] Animation presets (M3 motion tokens) + orchestration (Stagger, Chain, Repeat)
- [x] Transitions: enter/exit animations (Fade, Slide, Scale)
- [x] ScrollView widget (vertical/horizontal/both, wheel+keyboard+drag)
- [x] TabView widget (lazy content, closeable tabs, keyboard nav)
- [x] ListView widget (virtualized list, recycling, single/multi selection, M3 painter)
- [x] GridView widget (virtualized 2D grid, auto-fit columns, cell recycling)
- [x] LineChart widget (real-time, multiple series, rolling window)
- [x] ProgressBar widget (linear, rounded corners, signal binding)
- [x] Collapsible section widget (animated expand/collapse)
- [x] Box HBox/VBox direction support
- [x] Dirty region tracking (internal/dirty)
- [x] Performance benchmarks (36 across 5 packages)

### Phase 4: Production (v1.0) — In Progress

- [x] Circular progress indicator (determinate arc + indeterminate spinner)
- [x] SplitView (resizable split panels, draggable divider)
- [x] Popover/Tooltip (12 placements, auto-flip, overlay)
- [x] TreeView (hierarchical, expand/collapse, virtualized)
- [x] DataTable (sortable columns, fixed header, virtualized rows)
- [x] Toolbar (icon buttons, separators, spacers)
- [x] Menu system (MenuBar + ContextMenu, submenus, shortcuts)
- [x] IDE-style docking system (border layout, tabbed groups)
- [x] Drag & drop foundation (DragSource, DropTarget, Manager)
- [x] Fluent Design theme (9 painters, accent colors)
- [x] Cupertino theme (9 painters, iOS-style)
- [x] i18n (CLDR plural rules, RTL detection, LocaleSignal)
- [x] Icon system (vector paths, 10 Material icons, De Casteljau)
- [x] Font registry (CSS weight matching, W3C spec)
- [x] Testing utilities (MockCanvas, MockContext, assertions)
- [x] Dirty region tracking (merge algorithm, partial repaints)
- [x] Performance benchmarks (36 across 5 packages)
- [x] Hover tracking (W3C PointerEventSource, HoverTracker, cursor management)
- [x] ScreenBounds coordinate system (overlay positioning, hit-testing)
- [x] Event coordinate transforms (ScrollView content-space dispatch)
- [x] Inter font full Unicode (Cyrillic, Greek, Vietnamese)
- [x] MeasureText on Canvas (layout calculations without drawing)
- [x] FocusManager integration in Window (Tab/Shift+Tab navigation)
- [x] OnTextInput handler (platform character input API)
- [x] Task Manager example (charts, tables, animations)
- [x] Widget Gallery example (all widgets, 4 design systems, theme switching)
- [x] Incremental rendering pipeline (ADR-004): frame skip, persistent pixmap, dirty regions
- [x] Auto RepaintBoundary in ListView (per-item pixel caching)
- [x] DrawStats observability (CachedWidgets, DirtyRegionCount)
- [x] Tracker.Intersects() fast path in RepaintBoundary
- [x] Centralized ImageCache with LRU eviction (64MB, thread-safe)
- [x] Offscreen renderer (headless widget → *image.RGBA, no GPU/window)
- [ ] Platform accessibility adapters (UIA, AT-SPI2, NSAccessibility)
- [ ] Performance optimization pass

---

## Requirements

| Dependency | Purpose | Status |
|------------|---------|--------|
| Go 1.25+ | Language runtime | Required |
| CGO_ENABLED=0 | Pure Go (no C compiler needed) | Default |
| [gogpu/gogpu](https://github.com/gogpu/gogpu) | Windowing, input, GPU surface | For examples |
| [gogpu/gg](https://github.com/gogpu/gg) | 2D graphics rendering | Integrated |
| [gogpu/gpucontext](https://github.com/gogpu/gpucontext) | Shared interfaces | Integrated |
| [coregx/signals](https://github.com/coregx/signals) | Reactive state management | Integrated |

> **Note:** The entire ecosystem is pure Go. Do **not** set `CGO_ENABLED=1` — this will cause build errors from the `goffi` package which requires CGO to be disabled.

---

## Installation

```bash
# UI toolkit (library)
go get github.com/gogpu/ui@latest

# For runnable applications you also need gogpu (windowing) and gg (rendering):
go get github.com/gogpu/gogpu@latest
go get github.com/gogpu/gg@latest
```

---

## Related Projects

| Project | Description |
|---------|-------------|
| [gogpu/gogpu](https://github.com/gogpu/gogpu) | Graphics framework — GPU abstraction, windowing, input |
| [gogpu/gg](https://github.com/gogpu/gg) | 2D graphics — Canvas API, GPU text |
| [gogpu/wgpu](https://github.com/gogpu/wgpu) | Pure Go WebGPU — Vulkan, Metal, GLES, Software |
| [gogpu/naga](https://github.com/gogpu/naga) | Shader compiler — WGSL to SPIR-V, MSL, GLSL |

**Total ecosystem: 1.1M+ lines of Pure Go** — zero CGO, Rust optional via `-tags rust` (ADR-038).

---

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

**Ways to contribute:**
- Test the packages, report bugs
- API feedback and suggestions
- Documentation improvements
- Spread the word (Reddit, Hacker News, Dev.to)
- Code contributions (see open issues)

---

## License

MIT License — see [LICENSE](LICENSE) for details.

---

## Star History

<a href="https://star-history.com/#gogpu/ui&Date">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=gogpu/ui&type=Date&theme=dark" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/svg?repos=gogpu/ui&type=Date" />
   <img alt="Star History Chart" src="https://api.star-history.com/svg?repos=gogpu/ui&type=Date" />
 </picture>
</a>

---

<p align="center">
  <strong>gogpu/ui</strong> — Enterprise-grade GUI for Go<br>
  <sub>Part of the <a href="https://github.com/gogpu">GoGPU</a> ecosystem</sub>
</p>
