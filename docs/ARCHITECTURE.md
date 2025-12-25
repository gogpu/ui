# Architecture

> **gogpu/ui** — Enterprise-grade GUI library for Go

---

## Overview

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
│  VStack, HStack, Grid, Flexbox      │  coregx/signals       │
├─────────────────────────────────────────────────────────────┤
│  core/                              │  event/               │
│  Widget, WidgetBase, Context        │  Mouse, Keyboard      │
├─────────────────────────────────────────────────────────────┤
│  internal/render   │  internal/platform                     │
│  Canvas, Renderer  │  Win32, Cocoa, X11                     │
├─────────────────────────────────────────────────────────────┤
│  gogpu/gg          │  gogpu/gogpu    │  coregx/signals      │
│  2D Graphics       │  Windowing      │  State Management    │
└─────────────────────────────────────────────────────────────┘
```

---

## Package Structure

### Public Packages (Stable API)

| Package | Purpose | Stability |
|---------|---------|-----------|
| `core/` | Widget interface, Context, Geometry | Stable |
| `state/` | Signals integration | Stable |
| `layout/` | VStack, HStack, Grid, Flexbox | Stable |
| `widgets/` | Button, TextField, etc. | Stable |
| `theme/` | Theme interface and presets | Stable |
| `animation/` | Animation engine | Stable |
| `docking/` | IDE-style docking | Stable |
| `a11y/` | Accessibility | Stable |
| `i18n/` | Internationalization | Stable |
| `testing/` | Test utilities | Stable |

### Private Packages (Internal)

| Package | Purpose | Stability |
|---------|---------|-----------|
| `internal/render/` | Rendering pipeline | Can change |
| `internal/platform/` | Platform integration | Can change |
| `internal/layout/` | Layout algorithms | Can change |

### Unstable Packages (Experimental)

| Package | Purpose | Stability |
|---------|---------|-----------|
| `experimental/` | Unstable features | May change/remove |

---

## Core Concepts

### Widget Interface

```go
// core/widget.go
type Widget interface {
    // Layout calculates size given constraints
    Layout(ctx *LayoutContext) Size

    // Paint renders the widget
    Paint(ctx *PaintContext)

    // HandleEvent processes input events
    HandleEvent(event Event) bool
}
```

### Optional Interfaces (Extension Pattern)

```go
// Focusable — widgets that can receive focus
type Focusable interface {
    Widget
    Focus()
    Blur()
    IsFocused() bool
}

// Accessible — widgets with accessibility support
type Accessible interface {
    Widget
    AccessibilityRole() Role
    AccessibilityLabel() string
}

// Usage:
if f, ok := widget.(Focusable); ok {
    f.Focus()
}
```

### Composition via WidgetBase

```go
// core/widget_base.go
type WidgetBase struct {
    bounds   Rect
    visible  Signal[bool]
    enabled  Signal[bool]
    children []Widget
    parent   Widget
}

// User's custom widget:
type MyWidget struct {
    core.WidgetBase  // Embed for composition
    customField string
}
```

---

## State Management

### Signals (coregx/signals)

```go
// Reactive state
count := signals.NewSignal(0)

// Computed values
doubled := signals.NewComputed(func() int {
    return count.Get() * 2
})

// Side effects
signals.NewEffect(func() {
    fmt.Println("Count changed:", count.Get())
})
```

### Integration with Widgets

```go
// Widgets use signals for reactive properties
type Button struct {
    core.WidgetBase

    text     string
    disabled Signal[bool]  // Reactive property

    OnClick func()
}

// Auto-update on signal change
btn.disabled.Set(true)  // Widget automatically re-renders
```

---

## Layout System

### Layout Primitives

```go
// VStack — vertical layout
layout.VStack(
    widgets.Text{Content: "Title"},
    widgets.Button{Text: "Click"},
).WithSpacing(16)

// HStack — horizontal layout
layout.HStack(
    widgets.Button{Text: "Cancel"},
    widgets.Button{Text: "OK"},
).WithSpacing(8)

// Flexbox — CSS Flexbox-like
layout.Flex{
    Direction: layout.Row,
    Children: []Widget{
        Box{}.Flex(1),   // flex-grow: 1
        Box{}.Flex(2),   // flex-grow: 2
    },
}

// Grid — CSS Grid-like
layout.Grid{
    Columns: []GridTrack{Fr(1), Px(200), Fr(2)},
    Rows:    []GridTrack{Px(60), Fr(1)},
}
```

### Constraints Model

```go
type Constraints struct {
    MinWidth, MaxWidth   float32
    MinHeight, MaxHeight float32
}

func (w *Widget) Layout(ctx *LayoutContext) Size {
    // Measure within constraints
    size := w.measureChildren(ctx.Constraints)
    return ctx.Constraints.Constrain(size)
}
```

---

## Rendering Pipeline

### Canvas Abstraction

```go
// internal/render/canvas.go
type Canvas interface {
    // Drawing
    DrawRect(rect Rect, style RectStyle)
    DrawRoundedRect(rect Rect, radius float32, style RectStyle)
    DrawText(text string, pos Point, style TextStyle)
    DrawImage(img Image, rect Rect)
    DrawPath(path Path, style PathStyle)

    // State
    Save()
    Restore()
    Translate(x, y float32)
    Clip(rect Rect)
}
```

### Render Loop

```
1. Process Events (from platform)
2. Update State (signals trigger effects)
3. Layout Pass (if dirty)
4. Paint Pass (only dirty regions)
5. Present (to screen)
```

### Dirty Tracking

```go
// Only re-render changed widgets
func (w *WidgetBase) MarkDirty() {
    renderer.AddDirty(w)
}

// Signal changes trigger dirty
signals.NewEffect(func() {
    value := signal.Get()
    w.MarkDirty()
})
```

---

## Event System

### Event Types

```go
// Mouse events
type MouseEvent struct {
    Type     EventType  // Enter, Leave, Move, Down, Up, Click
    Position Point
    Button   MouseButton
    Modifiers Modifiers
}

// Keyboard events
type KeyEvent struct {
    Type      EventType  // Down, Up, Press
    Key       Key
    Modifiers Modifiers
}

// Focus events
type FocusEvent struct {
    Type          EventType  // Focus, Blur
    RelatedTarget Widget
}
```

### Event Propagation

```
1. Capture phase: root → target
2. Target phase: target handles
3. Bubble phase: target → root
4. StopPropagation() halts at any point
```

---

## Theme System

### Theme Structure

```go
type Theme struct {
    Colors     ColorPalette
    Typography Typography
    Spacing    SpacingScale
    Shadows    ShadowStyles
    Radii      RadiusScale
}

type ColorPalette struct {
    Primary      Color
    OnPrimary    Color
    Secondary    Color
    Background   Color
    Surface      Color
    Error        Color
    // ...
}
```

### Theme Presets

```go
// Material Design 3
theme := material3.Theme(seedColor)

// Microsoft Fluent
theme := fluent.Theme()

// Apple HIG
theme := cupertino.Theme()
```

---

## Accessibility

### ARIA-like Roles

```go
type Role int

const (
    RoleButton Role = iota
    RoleCheckbox
    RoleTextField
    RoleSlider
    RoleList
    RoleListItem
    RoleDialog
    // ...
)
```

### Platform Integration

```go
// Windows: UI Automation
// macOS: NSAccessibility
// Linux: AT-SPI

type PlatformAccessibility interface {
    CreateAccessibleNode(widget Accessible) AccessibleNode
    Announce(message string, priority Priority)
}
```

---

## Performance

### Virtualization

```go
// Only render visible items
widgets.VirtualizedList{
    ItemCount:  100000,
    ItemHeight: 50,
    RenderItem: func(index int) Widget {
        return widgets.Text{Content: items[index]}
    },
}
```

### Memory Pooling

```go
// Reuse widget objects
var pool = sync.Pool{
    New: func() any { return &WidgetBase{} },
}

func AcquireWidget() *WidgetBase {
    return pool.Get().(*WidgetBase)
}
```

### Batched Rendering

```go
// Batch similar draw calls
type DrawBatcher struct {
    rects []RectBatch
    texts []TextBatch
}

func (b *DrawBatcher) Flush(canvas Canvas) {
    // Single draw call per batch
}
```

---

## Dependencies

| Dependency | Purpose | Version |
|------------|---------|---------|
| gogpu/gg | 2D rendering | v0.13.0+ |
| gogpu/gogpu | Windowing | v0.8.0+ |
| coregx/signals | State management | v0.1.0+ |

---

## Design Principles

### 1. Composition over Inheritance

```go
// ✅ Composition
type Button struct {
    core.WidgetBase  // Embed
    text string
}

// ❌ Inheritance (not possible in Go anyway)
type Button extends WidgetBase { ... }
```

### 2. Headless Core

```go
// Core = behavior only
type Checkbox struct {
    Checked bool
    OnChange func(bool)
    // NO styling here
}

// Styling via theme
theme.ApplyStyle(checkbox)
```

### 3. Declarative API

```go
func App() Widget {
    return VStack(
        Text("Hello"),
        Button{Text: "Click", OnClick: handler},
    ).Padding(16)
}
```

### 4. Type Safety

```go
// Generics for type-safe state
type Signal[T any] interface {
    Get() T
    Set(T)
}

name := NewSignal[string]("John")
age := NewSignal[int](30)
```

---

*This architecture enables building enterprise-grade applications in pure Go.*
