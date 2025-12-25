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

## Design Decisions

Based on research of modern UI frameworks (GPUI, Xilem, Floem, iced, Dioxus), we made the following architectural decisions:

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Reactivity** | Fine-grained signals | O(affected) updates, not O(n). Proven by Floem/Lapce. |
| **Styling** | Tailwind-style builders | Type-safe, IDE autocomplete, AI-friendly. |
| **Layout** | Flexbox + incremental cache | Industry standard (Taffy-like). |
| **Accessibility** | AccessKit schema | Cross-platform standard. |
| **Effects** | Batched updates | Multiple changes = single render. |

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
| `a11y/` | Accessibility (AccessKit) | Stable |
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

    // Prepaint prepares GPU resources (optional, for batching)
    Prepaint(ctx *PrepaintContext) any

    // Paint renders the widget using prepaint state
    Paint(state any, ctx *PaintContext)

    // HandleEvent processes input events
    HandleEvent(event Event) EventResult
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

// Accessible — widgets with accessibility support (AccessKit-compatible)
type Accessible interface {
    Widget
    AccessibilityNode() AccessNode
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

## Styling API

### Tailwind-style Builders (Primary API)

All widgets use fluent method chaining for styling:

```go
// Clean, linear, type-safe
Button("Submit").
    Padding(16).
    PaddingX(24).
    Background(theme.Primary).
    Rounded(8).
    Shadow(1).
    OnHover(func(s *Style) {
        s.Background(theme.PrimaryDark)
    })
```

### Style Composition

Reusable styles via functional composition:

```go
// Define reusable styles
var PrimaryButton = Style(
    Padding(16),
    PaddingX(24),
    Background(theme.Primary),
    Rounded(8),
)

var DangerButton = Style(
    PrimaryButton,  // Inherit
    Background(theme.Error),
)

// Apply to widgets
Button("Submit").Apply(PrimaryButton)
Button("Delete").Apply(DangerButton)
```

### Why This Approach?

| Criteria | Tailwind-style | Struct-based | CSS-like |
|----------|---------------|--------------|----------|
| Type safety | 100% | 95% | 60% |
| IDE autocomplete | Excellent | Good | Poor |
| AI code generation | Optimal | Verbose | Error-prone |
| Compile-time errors | Yes | Mostly | No |

---

## State Management

### Signals (coregx/signals)

```go
// Reactive state
count := signals.NewSignal(0)

// Computed values (auto-tracked dependencies)
doubled := signals.NewComputed(func() int {
    return count.Get() * 2
})

// Side effects
signals.NewEffect(func() {
    fmt.Println("Count changed:", count.Get())
})
```

### Fine-grained Reactivity

Widgets automatically subscribe to signals used in their render functions:

```go
func Counter() Widget {
    count := signals.NewSignal(0)

    return VStack(
        // Label auto-subscribes to count
        Label(func() string {
            return fmt.Sprintf("Count: %d", count.Get())
        }),
        Button("+").OnClick(func() {
            count.Set(count.Get() + 1)  // Only Label re-renders
        }),
    ).Gap(8)
}
```

### Effect Batching

Multiple signal changes within the same event handler are batched:

```go
// Without batching: 3 renders
// With batching: 1 render
func handleFormSubmit() {
    name.Set("John")      // queued
    email.Set("j@x.com")  // queued
    age.Set(30)           // queued
}   // → single render at end
```

---

## Layout System

### Layout Primitives

```go
// VStack — vertical layout
VStack(
    Text("Title").Font(typography.H1),
    Text("Subtitle").Color(theme.OnSurfaceVariant),
    Button("Action"),
).Gap(16).Padding(24)

// HStack — horizontal layout
HStack(
    Button("Cancel").Variant(Outlined),
    Spacer(),
    Button("Save").Variant(Filled),
).Gap(8)

// Flex — CSS Flexbox
Flex(
    Box().Flex(1),   // flex-grow: 1
    Box().Flex(2),   // flex-grow: 2
).Direction(Row).JustifyBetween()

// Grid — CSS Grid
Grid(children...).
    Columns(Fr(1), Px(200), Fr(2)).
    Rows(Px(60), Fr(1)).
    Gap(16)
```

### Constraints Model

```go
type Constraints struct {
    MinWidth, MaxWidth   float32
    MinHeight, MaxHeight float32
}

func (w *Widget) Layout(ctx *LayoutContext) Size {
    size := w.measureChildren(ctx.Constraints)
    return ctx.Constraints.Constrain(size)
}
```

### Incremental Layout

Layout results are cached and only recomputed for dirty subtrees:

```go
type LayoutCache struct {
    size        Size
    constraints Constraints
    valid       bool
}

func (w *WidgetBase) Layout(ctx *LayoutContext) Size {
    if w.cache.valid && w.cache.constraints == ctx.Constraints {
        return w.cache.size  // Use cached result
    }
    // Recompute only if dirty
    size := w.computeLayout(ctx)
    w.cache = LayoutCache{size, ctx.Constraints, true}
    return size
}
```

---

## Rendering Pipeline

### Two-Phase Rendering

```go
// Phase 1: Prepaint (collect GPU resources)
state := widget.Prepaint(ctx)

// Phase 2: Paint (emit draw commands)
widget.Paint(state, ctx)
```

### Canvas Abstraction

```go
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
2. Update State (signals trigger effects, batched)
3. Layout Pass (only dirty subtrees)
4. Prepaint Pass (collect GPU resources)
5. Paint Pass (emit draw commands)
6. Present (to screen)
```

---

## Event System

### Event Types

```go
// Mouse events
type MouseEvent struct {
    Type      EventType  // Enter, Leave, Move, Down, Up, Click
    Position  Point
    Button    MouseButton
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
    Primary          Color
    OnPrimary        Color
    PrimaryContainer Color
    Secondary        Color
    Background       Color
    Surface          Color
    OnSurface        Color
    OnSurfaceVariant Color
    Error            Color
    Outline          Color
}
```

### Theme Presets

```go
// Material Design 3 (default)
theme := material3.Theme(material3.WithSeedColor(seedColor))

// Microsoft Fluent
theme := fluent.Theme()

// Apple Human Interface
theme := cupertino.Theme()
```

---

## Accessibility

### AccessKit Integration

We use AccessKit-compatible schema for cross-platform accessibility:

```go
type AccessNode struct {
    ID          uint64
    Role        AccessRole
    Name        string
    Description string
    Value       string
    Actions     []AccessAction
    Bounds      Rect
    Children    []uint64
    States      AccessState
}

type AccessState struct {
    Selected bool
    Expanded bool
    Checked  *bool  // nil = not applicable
    Disabled bool
    Focused  bool
}
```

### Platform Adapters

```go
// Windows: UI Automation
// macOS: NSAccessibility
// Linux: AT-SPI (D-Bus)

type PlatformAccessibility interface {
    CreateNode(widget Accessible) AccessibleNode
    UpdateNode(node AccessibleNode)
    RemoveNode(node AccessibleNode)
    Announce(message string, priority Priority)
}
```

---

## Performance

### Virtualization

```go
VirtualizedList(
    ItemCount(100000),
    ItemHeight(50),
    RenderItem(func(index int) Widget {
        return ListItem(items[index])
    }),
)
```

### Memory Pooling

```go
var nodePool = sync.Pool{
    New: func() any { return &LayoutNode{} },
}
```

### Batched Rendering

```go
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

## Proposals Under Discussion

The following patterns are under community review. See [GitHub Discussion #18](https://github.com/gogpu/gogpu/discussions/18) for feedback.

| Pattern | Description | Status |
|---------|-------------|--------|
| **linkedSignal** | Writable computed with reset capability for forms | Proposed |
| **Signal Forms** | FormGroup/FormField abstraction with validation | Proposed |
| **Context DI** | Type-safe injection for Theme, Logger, Config | Proposed |
| **Control Flow DSL** | If/For/Switch builders | Proposed |
| **Feature Stores** | State management best practices | Proposed |

---

## Design Principles

### 1. Composition over Inheritance

```go
// Embed WidgetBase for shared functionality
type Button struct {
    core.WidgetBase
    text string
}
```

### 2. Tailwind-style API

```go
// Fluent, type-safe, IDE-friendly
Button("Submit").Padding(16).Background(theme.Primary).Rounded(8)
```

### 3. Fine-grained Reactivity

```go
// Only affected widgets re-render
Label(func() string { return count.Get() })  // Auto-subscribes
```

### 4. Type Safety

```go
// Generics for compile-time safety
name := signals.NewSignal[string]("John")
age := signals.NewSignal[int](30)
```

### 5. Zero Allocations in Hot Paths

```go
// Pool layout nodes, batch draw calls
node := nodePool.Get().(*LayoutNode)
defer nodePool.Put(node)
```

---

*This architecture enables building enterprise-grade applications in pure Go.*
