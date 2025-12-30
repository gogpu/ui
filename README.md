<p align="center">
  <img src="https://raw.githubusercontent.com/gogpu/.github/main/assets/logo.png" alt="GoGPU Logo" width="120" />
</p>

<h1 align="center">gogpu/ui</h1>

<p align="center">
  <strong>Enterprise-Grade GUI Toolkit for Go</strong><br>
  Modern widgets, reactive state, GPU-accelerated rendering
</p>

<p align="center">
  <a href="https://github.com/gogpu/ui/actions"><img src="https://github.com/gogpu/ui/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/gogpu/ui"><img src="https://img.shields.io/badge/status-foundation-brightgreen" alt="Status"></a>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go" alt="Go Version"></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License"></a>
  <a href="https://github.com/gogpu/gogpu/stargazers"><img src="https://img.shields.io/github/stars/gogpu/gogpu?style=flat&labelColor=555&color=yellow" alt="Stars"></a>
  <a href="https://github.com/gogpu/gogpu/discussions"><img src="https://img.shields.io/github/discussions/gogpu/gogpu?style=flat&labelColor=555&color=blue" alt="Discussions"></a>
</p>

---

> **Join the Discussion:** Help shape the future of Go GUI! Share your ideas, report issues, and discuss features at our [GitHub Discussions](https://github.com/orgs/gogpu/discussions/18).

---

## Overview

**gogpu/ui** is a reference implementation of a professional GUI library for Go, designed for building:

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
| **Virtualization** | Yes | Limited | Manual |
| **IDE docking** | Yes | No | No |

---

## Status: Foundation Complete (v0.0.x)

> **Phase 0 Foundation is complete!** Core packages are implemented and tested.

### Implemented Packages

| Package | Description | Coverage |
|---------|-------------|----------|
| `geometry` | Point, Size, Rect, Constraints, Insets | 100% |
| `event` | MouseEvent, KeyEvent, WheelEvent, Modifiers | 100% |
| `widget` | Widget interface, WidgetBase, Context, Canvas, Color | 100% |
| `internal/render` | Canvas implementation using gogpu/gg | 96.5% |
| `internal/layout` | Flex, Stack, Grid layout engines | 89.9% |

**Total: ~10,261 lines of code with 95%+ average test coverage**

### Current Focus

- Phase 1: MVP with signals integration and window support
- API refinement based on community feedback

**Watch/Star the repo to follow development!**

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    User Application                         │
├─────────────────────────────────────────────────────────────┤
│  theme/material3   │  theme/fluent   │  theme/cupertino     │
│    (Planned)       │   (Planned)     │    (Planned)         │
├─────────────────────────────────────────────────────────────┤
│  widgets/         │  docking/        │  animation/          │
│  Button, TextField│  DockingHost     │  Animation, Spring   │
│  (Planned)        │  (Planned)       │  (Planned)           │
├─────────────────────────────────────────────────────────────┤
│  layout/                            │  state/               │
│  VStack, HStack, Grid, Flexbox      │  Signals              │
│  (Internal ✅)                      │  (Planned)            │
├─────────────────────────────────────────────────────────────┤
│  widget/                            │  event/               │
│  Widget, WidgetBase, Context        │  Mouse, Keyboard      │
│  (Complete ✅)                      │  (Complete ✅)        │
├─────────────────────────────────────────────────────────────┤
│  geometry/        │  internal/render │  internal/layout     │
│  Point, Rect      │  Canvas impl     │  Flex, Stack, Grid   │
│  (Complete ✅)    │  (Complete ✅)   │  (Complete ✅)       │
├─────────────────────────────────────────────────────────────┤
│  gogpu/gg          │  gogpu/gogpu    │  coregx/signals      │
│  2D Graphics       │  Windowing      │  State Management    │
└─────────────────────────────────────────────────────────────┘
```

---

## Planned API

```go
package main

import (
    "fmt"

    "github.com/gogpu/gogpu"
    "github.com/gogpu/ui/layout"
    "github.com/gogpu/ui/widgets"
    "github.com/coregx/signals"
)

func main() {
    app := gogpu.NewApp(gogpu.Config{
        Title:  "My Application",
        Width:  1280,
        Height: 720,
    })

    // Reactive state
    count := signals.New(0)

    // Declarative UI
    root := layout.VStack(
        widgets.Text("Counter Demo").FontSize(24),

        layout.HStack(
            widgets.Button("-").OnClick(func() {
                count.Set(count.Get() - 1)
            }),

            widgets.Text(signals.Computed(func() string {
                return fmt.Sprintf("Count: %d", count.Get())
            })),

            widgets.Button("+").OnClick(func() {
                count.Set(count.Get() + 1)
            }),
        ).Spacing(8),

        widgets.TextField().
            Placeholder("Enter text...").
            Width(300),
    ).Spacing(16).Padding(24)

    app.SetRoot(root)
    app.Run()
}
```

> **Note:** This is the target API design. Foundation is complete, widgets are in development.

---

## Implementation Progress

### Foundation (Phase 0) ✅

- [x] Geometry types (Point, Size, Rect, Constraints)
- [x] Event system (Mouse, Keyboard, Wheel, Focus)
- [x] Widget interface and WidgetBase
- [x] Canvas interface and implementation
- [x] Layout engine (Flex, Stack, Grid)
- [x] Color type with utilities

### Phase 1: MVP (In Progress)

- [ ] Signals integration (coregx/signals)
- [ ] Basic primitives (Box, Text, Image)
- [ ] Public layout API
- [ ] Theme system foundation
- [ ] Window integration (gogpu/gogpu)

### Phase 2: Beta

- [ ] Button, TextField, Label
- [ ] Checkbox, Radio, Switch
- [ ] Slider, Progress
- [ ] Dropdown, Select
- [ ] Material Design 3 theme

### Phase 3: Release Candidate

- [ ] List, Table, Tree (virtualized)
- [ ] Tabs, Accordion, SplitView
- [ ] Animation engine
- [ ] ScrollView with physics

### Phase 4: Production

- [ ] IDE-style docking
- [ ] Drag & drop
- [ ] Accessibility (WCAG 2.1 AA)
- [ ] Additional themes (Fluent, Cupertino)

---

## Requirements

| Dependency | Purpose | Status |
|------------|---------|--------|
| Go 1.25+ | Language runtime | Required |
| [gogpu/gg](https://github.com/gogpu/gg) | 2D graphics rendering | ✅ Integrated |
| [gogpu/gogpu](https://github.com/gogpu/gogpu) | Windowing and GPU abstraction | Phase 1 |
| [coregx/signals](https://github.com/coregx/signals) | Reactive state management | Phase 1 |

---

## Installation

```bash
go get github.com/gogpu/ui@latest
```

> **Note:** Currently provides foundation packages only. Full widget library coming in v0.1.0.

---

## Roadmap

| Phase | Version | Description | Status |
|-------|---------|-------------|--------|
| **Phase 0** | v0.0.x | Foundation: geometry, event, widget, layout | ✅ Complete |
| **Phase 1** | v0.1.0 | MVP: Signals, primitives, windowing | 🔄 In Progress |
| **Phase 2** | v0.2.0 | Beta: Widgets, Material 3 | Planned |
| **Phase 3** | v0.3.0 | RC: Virtualization, animation | Planned |
| **Phase 4** | v1.0.0 | Production: Docking, a11y, themes | Planned |

Full details: [ROADMAP.md](ROADMAP.md)

---

## Related Projects

| Project | Description | Purpose |
|---------|-------------|---------|
| [gogpu/gg](https://github.com/gogpu/gg) | 2D graphics | Canvas API, scene graph, GPU text |
| [gogpu/wgpu](https://github.com/gogpu/wgpu) | Pure Go WebGPU | Vulkan, Metal, GLES, Software backends |
| [gogpu/gogpu](https://github.com/gogpu/gogpu) | Graphics framework | GPU abstraction, windowing, input |
| [gogpu/naga](https://github.com/gogpu/naga) | Shader compiler | WGSL → SPIR-V, MSL, GLSL |

**Total ecosystem: 200K+ lines of Pure Go** — no CGO, no Rust, no C.

---

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

**Ways to contribute:**
- Design discussions in [GitHub Discussions](https://github.com/orgs/gogpu/discussions/18)
- API feedback and suggestions
- Documentation improvements
- Code contributions (see open issues)

---

## License

MIT License — see [LICENSE](LICENSE) for details.

---

<p align="center">
  <strong>gogpu/ui</strong> — Enterprise-grade GUI for Go<br>
  <sub>Part of the <a href="https://github.com/gogpu">GoGPU</a> ecosystem</sub>
</p>
