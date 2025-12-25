<p align="center">
  <img src="https://raw.githubusercontent.com/gogpu/.github/main/assets/logo.png" alt="GoGPU Logo" width="120" />
</p>

<h1 align="center">gogpu/ui</h1>

<p align="center">
  <strong>Enterprise-Grade GUI Toolkit for Go</strong><br>
  Modern widgets, reactive state, GPU-accelerated rendering
</p>

<p align="center">
  <a href="https://github.com/gogpu/ui"><img src="https://img.shields.io/badge/version-v0.0.0-blue" alt="Version"></a>
  <a href="https://github.com/gogpu/ui"><img src="https://img.shields.io/badge/status-planning-orange" alt="Status"></a>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go" alt="Go Version"></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License"></a>
  <a href="https://github.com/gogpu/gogpu/stargazers"><img src="https://img.shields.io/github/stars/gogpu/gogpu?style=flat&labelColor=555&color=yellow" alt="Stars"></a>
  <a href="https://github.com/gogpu/gogpu/discussions"><img src="https://img.shields.io/github/discussions/gogpu/gogpu?style=flat&labelColor=555&color=blue" alt="Discussions"></a>
</p>

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

## Status: Planning (v0.0.0)

> **Development has not yet started.** The project is in the design and planning phase.

Current focus:
- Architecture design
- API specification
- Dependency coordination with gogpu ecosystem

**Watch/Star the repo to be notified when development begins!**

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
│  VStack, HStack, Grid, Flexbox      │  Signals              │
├─────────────────────────────────────────────────────────────┤
│  core/                              │  event/               │
│  Widget, WidgetBase, Context        │  Mouse, Keyboard      │
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
    "github.com/gogpu/gogpu"
    "github.com/gogpu/ui"
    "github.com/gogpu/ui/widgets"
    "github.com/gogpu/ui/layout"
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

> **Note:** This is the target API design, not yet implemented.

---

## Planned Features

### Core
- [x] Widget interface design
- [ ] Signals integration (coregx/signals)
- [ ] Event system (mouse, keyboard, focus)
- [ ] Rendering pipeline (gogpu/gg)

### Widgets
- [ ] Button, TextField, Label
- [ ] Checkbox, Radio, Switch
- [ ] Slider, Progress
- [ ] Dropdown, Select, ComboBox
- [ ] List, Table, Tree (virtualized)
- [ ] Tabs, Accordion, SplitView
- [ ] Dialog, Popover, Tooltip

### Layout
- [ ] VStack, HStack (Flexbox)
- [ ] Grid (CSS Grid-like)
- [ ] Absolute positioning
- [ ] ScrollView

### Themes
- [ ] Material Design 3
- [ ] Microsoft Fluent
- [ ] Apple Cupertino

### Enterprise
- [ ] IDE-style docking
- [ ] Drag & drop
- [ ] Virtualization (100K+ items)
- [ ] Animation engine
- [ ] Accessibility (WCAG 2.1 AA)
- [ ] Internationalization (RTL, i18n)

---

## Requirements

| Dependency | Version | Purpose |
|------------|---------|---------|
| Go | 1.25+ | Language runtime |
| gogpu/gg | 0.13.0+ | 2D graphics |
| gogpu/gogpu | 0.8.0+ | Windowing |
| coregx/signals | 0.1.0+ | State management |

---

## Roadmap

| Phase | Version | Description |
|-------|---------|-------------|
| **Phase 1** | v0.1.0 | MVP: Core, layout, events |
| **Phase 2** | v0.2.0 | Beta: Widgets, Material 3 |
| **Phase 3** | v0.3.0 | RC: Virtualization, animation |
| **Phase 4** | v1.0.0 | Production: Docking, a11y, themes |

Full details: [ROADMAP.md](ROADMAP.md)

---

## Related Projects

| Project | Description | Status |
|---------|-------------|--------|
| [gogpu/gogpu](https://github.com/gogpu/gogpu) | Graphics framework, windowing | v0.4.0 |
| [gogpu/gg](https://github.com/gogpu/gg) | 2D graphics library | v0.4.0 |
| [gogpu/wgpu](https://github.com/gogpu/wgpu) | Pure Go WebGPU | v0.4.0 |
| [gogpu/naga](https://github.com/gogpu/naga) | Shader compiler | v0.6.0 |

---

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

**Ways to contribute:**
- Design discussions in Issues
- API feedback
- Documentation improvements
- Research on GUI patterns

---

## License

MIT License — see [LICENSE](LICENSE) for details.

---

<p align="center">
  <strong>gogpu/ui</strong> — Enterprise-grade GUI for Go
</p>
