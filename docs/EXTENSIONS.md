# Creating Extensions for gogpu/ui

> **Version:** 0.1.x | **Updated:** January 2026

This guide explains how to extend gogpu/ui with custom widgets, themes, layouts, and plugins.

---

## Table of Contents

1. [Overview](#overview)
2. [Creating Custom Widgets](#creating-custom-widgets)
3. [Creating Custom Themes](#creating-custom-themes)
4. [Creating Custom Layouts](#creating-custom-layouts)
5. [Creating Plugins](#creating-plugins)
6. [Best Practices](#best-practices)

---

## Overview

gogpu/ui provides four extension points:

| Extension | Package | Purpose |
|-----------|---------|---------|
| **Widgets** | `registry/` | Custom UI components |
| **Themes** | `theme/` | Visual styling |
| **Layouts** | `layout/` | Custom positioning algorithms |
| **Plugins** | `plugin/` | Bundle all of the above |

All extensions use the `init()` auto-registration pattern for seamless integration.

---

## Creating Custom Widgets

### Step 1: Implement the Widget Interface

```go
package mywidgets

import (
    "github.com/gogpu/ui/geometry"
    "github.com/gogpu/ui/event"
    "github.com/gogpu/ui/widget"
)

// MyButton is a custom button widget.
type MyButton struct {
    widget.WidgetBase
    label   string
    onClick func()
}

// NewMyButton creates a new button with the given label.
func NewMyButton(label string) *MyButton {
    return &MyButton{label: label}
}

// Layout calculates the widget size.
func (b *MyButton) Layout(wc widget.Context, constraints geometry.Constraints) geometry.Size {
    // Button needs at least 100x40 pixels
    return constraints.Constrain(geometry.Size{Width: 100, Height: 40})
}

// Draw renders the button.
func (b *MyButton) Draw(wc widget.Context, canvas widget.Canvas) {
    bounds := b.Bounds()

    // Draw background
    canvas.SetFillColor(widget.Hex(0x2196F3))
    canvas.FillRoundedRect(bounds, 8)

    // Draw label (centered)
    canvas.SetFillColor(widget.White)
    canvas.DrawText(b.label, bounds.Center())
}

// HandleEvent processes user input.
func (b *MyButton) HandleEvent(wc widget.Context, ev event.Event) bool {
    if mouse, ok := ev.(*event.MouseEvent); ok {
        if mouse.Type == event.MousePress && mouse.Button == event.ButtonLeft {
            if b.onClick != nil {
                b.onClick()
            }
            return true
        }
    }
    return false
}

// OnClick sets the click handler.
func (b *MyButton) OnClick(handler func()) *MyButton {
    b.onClick = handler
    return b
}
```

### Step 2: Register the Widget Factory

```go
package mywidgets

import "github.com/gogpu/ui/registry"

func init() {
    registry.RegisterWidget("my-button", createMyButton, registry.WidgetInfo{
        Name:        "my-button",
        Description: "Custom styled button",
        Category:    registry.CategoryInput,
        Version:     "1.0.0",
    })
}

func createMyButton(config map[string]any) (registry.Widget, error) {
    label, _ := config["label"].(string)
    if label == "" {
        label = "Button"
    }
    return NewMyButton(label), nil
}
```

### Step 3: Use the Widget

```go
package main

import (
    "github.com/gogpu/ui/registry"
    _ "github.com/example/mywidgets" // Auto-registers via init()
)

func main() {
    // Create by name (dynamic)
    btn, err := registry.CreateWidget("my-button", map[string]any{
        "label": "Click Me",
    })

    // Or create directly (static)
    btn := mywidgets.NewMyButton("Click Me").OnClick(func() {
        fmt.Println("Clicked!")
    })
}
```

---

## Creating Custom Themes

### Step 1: Create a Theme

```go
package mytheme

import (
    "github.com/gogpu/ui/theme"
    "github.com/gogpu/ui/widget"
)

// CorporateTheme returns a branded theme.
func CorporateTheme() *theme.Theme {
    t := theme.DefaultLight()

    // Brand colors
    t.Colors.Primary = widget.Hex(0x1A237E)      // Deep blue
    t.Colors.Secondary = widget.Hex(0xFFC107)    // Amber
    t.Colors.PrimaryVariant = widget.Hex(0x534BAE)

    // Custom typography
    t.Typography.FontFamily = "Inter"

    // Tighter spacing
    t.Spacing = theme.CompactSpacing()

    return t
}
```

### Step 2: Register the Theme

```go
package mytheme

import "github.com/gogpu/ui/theme"

func init() {
    theme.Register("corporate", CorporateTheme(), theme.ThemeInfo{
        Name:        "corporate",
        Description: "Corporate brand theme",
        Author:      "My Company",
        Version:     "1.0.0",
    })
}
```

### Step 3: Use the Theme

```go
package main

import (
    "github.com/gogpu/ui/theme"
    _ "github.com/example/mytheme" // Auto-registers
)

func main() {
    // Get by name
    t, ok := theme.Get("corporate")
    if !ok {
        t = theme.DefaultLight()
    }

    // Use theme properties
    primaryColor := t.Colors.Primary
    bodyStyle := t.Typography.BodyMedium
    padding := t.Spacing.M
}
```

### Creating Theme Extensions

For component-specific theming, implement `ThemeExtension`:

```go
package mywidgets

import "github.com/gogpu/ui/theme"

// ButtonTheme contains button-specific styling.
type ButtonTheme struct {
    BorderRadius float32
    Elevation    int
    RippleColor  widget.Color
}

func (b *ButtonTheme) Name() string { return "my-button-theme" }

func (b *ButtonTheme) Merge(other theme.ThemeExtension) theme.ThemeExtension {
    if o, ok := other.(*ButtonTheme); ok {
        return &ButtonTheme{
            BorderRadius: o.BorderRadius,
            Elevation:    o.Elevation,
            RippleColor:  o.RippleColor,
        }
    }
    return b
}

func (b *ButtonTheme) Lerp(other theme.ThemeExtension, t float32) theme.ThemeExtension {
    if o, ok := other.(*ButtonTheme); ok {
        return &ButtonTheme{
            BorderRadius: b.BorderRadius + (o.BorderRadius-b.BorderRadius)*t,
            Elevation:    b.Elevation, // No interpolation for int
            RippleColor:  b.RippleColor.Lerp(o.RippleColor, t),
        }
    }
    return b
}

func (b *ButtonTheme) CopyWith(overrides map[string]any) theme.ThemeExtension {
    copy := *b
    if v, ok := overrides["borderRadius"].(float32); ok {
        copy.BorderRadius = v
    }
    return &copy
}

// Usage
func init() {
    theme.RegisterExtension("my-button-theme", &ButtonTheme{
        BorderRadius: 8,
        Elevation:    2,
        RippleColor:  widget.RGBA(0, 0, 0, 0.1),
    })
}
```

---

## Creating Custom Layouts

### Step 1: Implement LayoutAlgorithm

```go
package mylayout

import (
    "github.com/gogpu/ui/geometry"
    "github.com/gogpu/ui/layout"
)

// MasonryLayout arranges children in a Pinterest-style grid.
type MasonryLayout struct {
    Columns int
    Gap     float32
}

func (m *MasonryLayout) Name() string {
    return "masonry"
}

func (m *MasonryLayout) Compute(
    tree layout.LayoutTree,
    root layout.NodeID,
    available geometry.Size,
) layout.Result {
    children := tree.Children(root)
    if len(children) == 0 {
        return layout.Result{Size: geometry.Size{}}
    }

    // Calculate column width
    totalGap := float32(m.Columns-1) * m.Gap
    colWidth := (available.Width - totalGap) / float32(m.Columns)

    // Track column heights
    colHeights := make([]float32, m.Columns)
    positions := make(map[layout.NodeID]geometry.Point)

    for _, child := range children {
        // Find shortest column
        minCol := 0
        for i := 1; i < m.Columns; i++ {
            if colHeights[i] < colHeights[minCol] {
                minCol = i
            }
        }

        // Position child
        x := float32(minCol) * (colWidth + m.Gap)
        y := colHeights[minCol]
        positions[child] = geometry.Point{X: x, Y: y}

        // Measure child and update column height
        childSize := tree.Measure(child, geometry.Constraints{
            MinWidth:  colWidth,
            MaxWidth:  colWidth,
            MinHeight: 0,
            MaxHeight: available.Height,
        })
        colHeights[minCol] += childSize.Height + m.Gap
    }

    // Find max height
    maxHeight := float32(0)
    for _, h := range colHeights {
        if h > maxHeight {
            maxHeight = h
        }
    }

    return layout.Result{
        Size:      geometry.Size{Width: available.Width, Height: maxHeight},
        Positions: positions,
    }
}
```

### Step 2: Register the Layout

```go
package mylayout

import "github.com/gogpu/ui/layout"

func init() {
    layout.RegisterLayout("masonry", &MasonryLayout{
        Columns: 3,
        Gap:     16,
    })
}
```

### Step 3: Use the Layout

```go
// Get registered layout
algo, ok := layout.GetLayout("masonry")

// Or use directly
masonry := &mylayout.MasonryLayout{Columns: 4, Gap: 8}
result := masonry.Compute(tree, rootID, availableSize)
```

---

## Creating Plugins

Plugins bundle widgets, themes, and layouts into a single distributable package.

### Step 1: Implement the Plugin Interface

```go
package myplugin

import (
    "github.com/gogpu/ui/plugin"
    "github.com/gogpu/ui/widget"
)

type MyPlugin struct{}

func (p *MyPlugin) Name() string    { return "my-plugin" }
func (p *MyPlugin) Version() string { return "1.0.0" }

func (p *MyPlugin) Dependencies() []plugin.Dependency {
    return []plugin.Dependency{
        // Optional: depend on other plugins
        // {Name: "base-widgets", Version: ">=1.0.0"},
    }
}

func (p *MyPlugin) Init(ctx *plugin.PluginContext) error {
    // Register widgets
    ctx.Widgets.Register("my-button", createMyButton)
    ctx.Widgets.Register("my-card", createMyCard)

    // Register themes
    ctx.Themes.Register("my-light", myLightTheme())
    ctx.Themes.Register("my-dark", myDarkTheme())

    // Register layouts
    ctx.Layouts.Register("masonry", &MasonryLayout{})

    // Load assets
    ctx.Assets.RegisterFont("my-font", myFontData)
    ctx.Assets.RegisterIcon("my-icons", myIconSet)

    return nil
}

func (p *MyPlugin) Shutdown() error {
    // Cleanup resources if needed
    return nil
}
```

### Step 2: Register the Plugin

```go
package myplugin

import "github.com/gogpu/ui/plugin"

func init() {
    plugin.Register(&MyPlugin{}, plugin.PluginInfo{
        Name:        "my-plugin",
        Description: "A complete UI component library",
        Version:     "1.0.0",
        Author:      "My Team",
        License:     "MIT",
        Tags:        []string{"widgets", "themes", "material"},
    })
}
```

### Step 3: Use the Plugin

```go
package main

import (
    "github.com/gogpu/ui/plugin"
    "github.com/gogpu/ui/registry"
    "github.com/gogpu/ui/theme"

    _ "github.com/example/myplugin" // Auto-registers
)

func main() {
    // Initialize all plugins (resolves dependencies)
    if err := plugin.Initialize(); err != nil {
        log.Fatal(err)
    }
    defer plugin.Shutdown()

    // List available plugins
    for _, name := range plugin.List() {
        info, _ := plugin.Info(name)
        fmt.Printf("%s v%s by %s\n", name, info.Version, info.Author)
    }

    // Use components from plugins
    btn, _ := registry.CreateWidget("my-button", nil)
    t, _ := theme.Get("my-light")
}
```

---

## Best Practices

### Widget Development

1. **Embed WidgetBase** for common functionality
2. **Validate config** in factory functions
3. **Provide sensible defaults** for optional parameters
4. **Return descriptive errors** for invalid configuration
5. **Use semantic versioning** for widget versions

### Theme Development

1. **Start from DefaultLight/DefaultDark** as a base
2. **Follow Material 3** color semantics
3. **Test both light and dark modes**
4. **Provide accessibility variants** (high contrast)
5. **Use ThemeExtension** for component-specific styling

### Layout Development

1. **Handle empty children** gracefully
2. **Respect constraints** from parent
3. **Cache expensive calculations** when possible
4. **Return accurate size** in Result

### Plugin Development

1. **Declare dependencies** explicitly
2. **Use semantic versioning** for compatibility
3. **Initialize resources** in Init(), cleanup in Shutdown()
4. **Avoid global state** outside of registries
5. **Document all exported components**

### General

1. **Use init() for auto-registration** - enables blank imports
2. **Provide both static and dynamic** creation options
3. **Write comprehensive tests** (aim for 80%+ coverage)
4. **Follow Go naming conventions** (exported = public API)
5. **Document with godoc comments**

---

## Publishing Extensions

### Package Structure

```
github.com/yourname/ui-extension/
├── go.mod
├── README.md
├── LICENSE
├── doc.go           # Package documentation
├── widgets/         # Custom widgets
│   ├── button.go
│   └── card.go
├── themes/          # Custom themes
│   └── corporate.go
├── layouts/         # Custom layouts
│   └── masonry.go
└── plugin.go        # Plugin registration
```

### go.mod Example

```go
module github.com/yourname/ui-extension

go 1.25

require github.com/gogpu/ui v0.1.0
```

### README Template

```markdown
# My UI Extension

Custom widgets, themes, and layouts for gogpu/ui.

## Installation

go get github.com/yourname/ui-extension

## Usage

import _ "github.com/yourname/ui-extension"

## Components

### Widgets
- `my-button` - Custom styled button
- `my-card` - Material card component

### Themes
- `corporate` - Corporate brand theme

### Layouts
- `masonry` - Pinterest-style grid
```

---

## Version Compatibility

| gogpu/ui | Extension API |
|----------|---------------|
| v0.1.x | Stable (registry, theme, layout, plugin) |
| v0.2.x | Stable + new widget primitives |

Extensions built for v0.1.x will remain compatible with future v0.x releases.

---

*For questions and support, visit [GitHub Discussions](https://github.com/orgs/gogpu/discussions).*
