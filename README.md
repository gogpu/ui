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
  <a href="https://github.com/gogpu/ui"><img src="https://img.shields.io/badge/status-Phase_2_Beta-brightgreen" alt="Status"></a>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go" alt="Go Version"></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License"></a>
  <a href="https://github.com/gogpu/gogpu/stargazers"><img src="https://img.shields.io/github/stars/gogpu/gogpu?style=flat&labelColor=555&color=yellow" alt="Stars"></a>
  <a href="https://github.com/gogpu/gogpu/discussions"><img src="https://img.shields.io/github/discussions/gogpu/gogpu?style=flat&labelColor=555&color=blue" alt="Discussions"></a>
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

## Quick Start

```go
package main

import (
    "fmt"

    "github.com/gogpu/gogpu"
    "github.com/gogpu/ui/app"
    "github.com/gogpu/ui/primitives"
    "github.com/gogpu/ui/state"
    "github.com/gogpu/ui/widget"
)

func main() {
    gogpuApp := gogpu.NewApp(gogpu.Config{
        Title:  "My App",
        Width:  800,
        Height: 600,
    })

    uiApp := app.New(
        app.WithWindowProvider(gogpuApp),
        app.WithPlatformProvider(gogpuApp),
    )

    // Reactive state
    count := state.NewSignal(0)

    uiApp.SetRoot(
        primitives.Box(
            primitives.Text("Hello gogpu/ui!").
                FontSize(24).
                Bold().
                Color(widget.RGBA8(33, 33, 33, 255)),

            primitives.TextFn(func() string {
                return fmt.Sprintf("Count: %d", count.Get())
            }).FontSize(18),

            primitives.Box().
                Width(200).Height(40).
                Background(widget.RGBA8(98, 0, 238, 255)).
                Rounded(8),
        ).Padding(24).Gap(12).Background(widget.RGBA8(255, 255, 255, 255)),
    )
}
```

---

## Packages

### Core (Phase 0)

| Package | Description | Coverage |
|---------|-------------|----------|
| `geometry` | Point, Size, Rect, Constraints, Insets | 98.8% |
| `event` | MouseEvent, KeyEvent, WheelEvent, FocusEvent, Modifiers | 100% |
| `widget` | Widget interface, WidgetBase, Context, Canvas, Color, ThemeProvider | 100% |
| `internal/render` | Canvas implementation using gogpu/gg | 96.5% |
| `internal/layout` | Flex, Stack, Grid layout engines | 89.9% |

### MVP (Phase 1)

| Package | Description | Coverage |
|---------|-------------|----------|
| `a11y` | Accessibility: 35+ ARIA roles, Accessible interface, Tree, Announcer | 99.1% |
| `state` | Reactive signals (coregx/signals), Binding, Scheduler with batching | 100% |
| `primitives` | Box, Text, Image widgets with fluent builder API | 94.4% |
| `app` | Window integration via gpucontext interfaces (dependency inversion) | 98.6% |

### Extensibility (Phase 1.5)

| Package | Description | Coverage |
|---------|-------------|----------|
| `layout` | Public layout API with custom algorithms | 89.5% |
| `registry` | Widget factory registration for third-party widgets | 100% |
| `theme` | Theme system with Extensions and Registry | 100% |
| `plugin` | Plugin bundling with dependency resolution | 99.4% |

### Interactive Widgets (Phase 2 — Partial)

| Package | Description | Coverage |
|---------|-------------|----------|
| `cdk` | Component Development Kit — Content[C] polymorphic pattern | 100% |
| `core/button` | Generic button with pluggable Painter, 4 variants, 3 sizes | 96%+ |
| `theme/material3` | Material Design 3 — theme (HCT color science) + theme-aware ButtonPainter | 97%+ |
| `focus` | Keyboard focus management with Tab/Shift+Tab navigation | 95.2% |
| `internal/focus` | Internal focus manager implementation | 15.2% |

**Total: ~45,000+ lines of code | 21 packages | 1,194+ tests | ~97% average coverage**

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    User Application                         │
├─────────────────────────────────────────────────────────────┤
│  theme/material3/  │  fluent/        │  cupertino/          │
│  Theme + Painters  │   (Planned)     │    (Planned)         │
│  (Complete ✅)     │                 │                      │
├─────────────────────────────────────────────────────────────┤
│  core/button/      │  docking/       │  animation/          │
│  focus/            │  DockingHost    │  Animation, Spring   │
│  (Complete ✅)     │  (Phase 3+)    │  (Phase 3)            │
├──────────────┬──────────────────────────────────────────────┤
│  cdk/        │  Content[C] polymorphic pattern              │
│  (Complete ✅)│  Headless behaviors (Phase 3)               │
├──────────────┴──────────────────────────────────────────────┤
│  primitives/       │  app/           │  a11y/               │
│  Box, Text, Image  │  Window, Loop   │  Roles, Tree, Node   │
│  (Complete ✅)     │  (Complete ✅) │  (Complete ✅)       │
├─────────────────────────────────────────────────────────────┤
│  registry/         │  plugin/        │  theme/              │
│  Widget Factory    │  Plugin System  │  Theme + Extensions  │
│  (Complete ✅)     │  (Complete ✅) │  (Complete ✅)       │
├─────────────────────────────────────────────────────────────┤
│  layout/           │  state/                                │
│  VStack, HStack,   │  Signals, Binding,                     │
│  Grid, Flexbox     │  Scheduler                             │
│  (Complete ✅)     │  (Complete ✅)                        │
├─────────────────────────────────────────────────────────────┤
│  widget/           │  event/         │  geometry/           │
│  Widget, Context   │  Mouse, Key     │  Point, Rect         │
│  (Complete ✅)     │  (Complete ✅) │  (Complete ✅)       │
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

// Image
primitives.Image(mySource).
    Size(48, 48).
    Cover().
    Rounded(24).
    Alt("User avatar")
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

### Phase 2: Beta (In Progress)

- [x] Button widget (4 variants, 3 sizes, keyboard activation)
- [x] Material Design 3 theme (HCT color science, light/dark schemes)
- [x] Keyboard navigation (focus management, Tab/Shift+Tab, shortcuts)
- [ ] TextField widget
- [ ] Checkbox & Radio
- [ ] Container widgets (ScrollView, Panel)
- [ ] Typography system

### Phase 3: Release Candidate

- [ ] Virtualized lists and grids
- [ ] Animation engine (Spring, Tween)
- [ ] Gesture recognition
- [ ] Tabs, SplitView, Accordion

### Phase 4: Production (v1.0)

- [ ] IDE-style docking system
- [ ] Platform accessibility adapters (UIA, AT-SPI2, NSAccessibility)
- [ ] Drag & drop
- [ ] Additional themes (Fluent, Cupertino)

---

## Requirements

| Dependency | Purpose | Status |
|------------|---------|--------|
| Go 1.25+ | Language runtime | Required |
| [gogpu/gg](https://github.com/gogpu/gg) | 2D graphics rendering | Integrated |
| [gogpu/gpucontext](https://github.com/gogpu/gpucontext) | Shared interfaces | Integrated |
| [coregx/signals](https://github.com/coregx/signals) | Reactive state management | Integrated |

---

## Installation

```bash
go get github.com/gogpu/ui@latest
```

---

## Related Projects

| Project | Description |
|---------|-------------|
| [gogpu/gogpu](https://github.com/gogpu/gogpu) | Graphics framework — GPU abstraction, windowing, input |
| [gogpu/gg](https://github.com/gogpu/gg) | 2D graphics — Canvas API, GPU text |
| [gogpu/wgpu](https://github.com/gogpu/wgpu) | Pure Go WebGPU — Vulkan, Metal, GLES, Software |
| [gogpu/naga](https://github.com/gogpu/naga) | Shader compiler — WGSL to SPIR-V, MSL, GLSL |

**Total ecosystem: 250K+ lines of Pure Go** — no CGO, no Rust, no C.

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

<p align="center">
  <strong>gogpu/ui</strong> — Enterprise-grade GUI for Go<br>
  <sub>Part of the <a href="https://github.com/gogpu">GoGPU</a> ecosystem</sub>
</p>
