<p align="center">
  <img src="https://raw.githubusercontent.com/gogpu/.github/main/assets/logo.png" alt="GoGPU Logo" width="120" />
</p>

<h1 align="center">gogpu/ui</h1>

<p align="center">
  <strong>Pure Go GUI Toolkit</strong><br>
  Modern widgets, layouts, and styling — built on GoGPU
</p>

<p align="center">
  <a href="https://github.com/gogpu/ui"><img src="https://img.shields.io/badge/status-planned-orange" alt="Status"></a>
  <a href="https://github.com/gogpu/gogpu"><img src="https://img.shields.io/badge/requires-gogpu-blue" alt="Requires"></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License"></a>
</p>

---

## Status: Planned

> **This project is not yet started.** It will be developed after `gogpu/gogpu` reaches v0.2.0 (basic 2D rendering).
>
> **Star the repo to be notified when development begins!**

---

## Vision

A modern, GPU-accelerated GUI toolkit for Go that:

- **Zero CGO** — Pure Go, simple `go build`
- **GPU Rendered** — Smooth 60fps UI with hardware acceleration
- **Immediate + Retained** — Flexible rendering modes
- **Cross-Platform** — Windows, Linux, macOS
- **Themeable** — Built-in dark/light themes, custom styling

---

## Planned Features

### Widgets
- [ ] Button, Label, TextInput
- [ ] Checkbox, Radio, Slider
- [ ] Dropdown, ComboBox
- [ ] List, Table, Tree
- [ ] Tabs, Accordion
- [ ] Modal, Tooltip, Popup
- [ ] ScrollView, SplitPane

### Layouts
- [ ] Flex (row/column)
- [ ] Grid
- [ ] Stack
- [ ] Absolute positioning

### Styling
- [ ] CSS-like styling
- [ ] Themes (dark/light)
- [ ] Custom fonts
- [ ] Icons (embedded SVG)

### Accessibility
- [ ] Keyboard navigation
- [ ] Screen reader support
- [ ] High contrast mode

---

## Architecture

```
┌─────────────────────────────────────┐
│         Your Application            │
├─────────────────────────────────────┤
│            gogpu/ui                 │  ← This library
│    Widgets, Layouts, Styling        │
├─────────────────────────────────────┤
│            gogpu/gg                 │
│         2D Graphics API             │
├─────────────────────────────────────┤
│           gogpu/gogpu               │
│      GPU, Window, Input, Math       │
├─────────────────────────────────────┤
│          WebGPU Runtime             │
└─────────────────────────────────────┘
```

---

## Target API

```go
package main

import (
    "github.com/gogpu/gogpu"
    "github.com/gogpu/ui"
)

func main() {
    app := gogpu.NewApp(gogpu.Config{
        Title: "My App",
        Width: 800,
        Height: 600,
    })

    // Create UI
    root := ui.Column(
        ui.Label("Hello, GoGPU!").FontSize(24),
        ui.Row(
            ui.Button("Click Me").OnClick(func() {
                println("Clicked!")
            }),
            ui.Button("Cancel"),
        ).Gap(8),
        ui.TextInput().Placeholder("Enter text..."),
    ).Padding(16).Gap(12)

    app.SetRoot(root)
    app.Run()
}
```

> **Note:** This API is a design target, not implemented yet.

---

## Inspiration

- [egui](https://github.com/emilk/egui) (Rust) — Immediate mode GUI
- [Gio](https://gioui.org) (Go) — Portable GUI
- [Flutter](https://flutter.dev) — Widget composition
- [SwiftUI](https://developer.apple.com/xcode/swiftui/) — Declarative syntax
- [Tailwind CSS](https://tailwindcss.com) — Utility-first styling

---

## Timeline

| Phase | Milestone | Depends On |
|-------|-----------|------------|
| Phase 1 | Basic widgets | gogpu v0.2.0 |
| Phase 2 | Layouts | Phase 1 |
| Phase 3 | Styling/Themes | Phase 2 |
| Phase 4 | Advanced widgets | Phase 3 |
| Phase 5 | Accessibility | Phase 4 |

---

## Related Projects

| Project | Description |
|---------|-------------|
| [gogpu/gogpu](https://github.com/gogpu/gogpu) | Graphics framework (foundation) |
| [gogpu/gg](https://github.com/gogpu/gg) | 2D graphics library |
| [gogpu/naga](https://github.com/gogpu/naga) | Shader compiler |

---

## Contributing

This project is in the planning phase. Contributions to the design are welcome!

- Open issues to discuss widget designs
- Share inspiration from other GUI toolkits
- Help define the styling system

---

## License

MIT License — see [LICENSE](LICENSE) for details.

---

<p align="center">
  <strong>gogpu/ui</strong> — The GUI toolkit Go deserves
</p>
