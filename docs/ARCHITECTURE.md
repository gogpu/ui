# Architecture

> **gogpu/ui** -- Enterprise-grade GUI toolkit for Go

---

## Overview

### 3-Layer Architecture (ADR-003)

```
+--------------------------------------------------------------+
|                    User Application                          |
+==============================================================+
|            Layer 3b: Design Systems (styling)                |
| theme/material3/  |  (future)         |  (future)            |
| M3 Button/Check/  |  fluent/          |  cupertino/          |
| RadioPainter      |                   |                      |
+-------------------+-------------------+----------------------+
|         Layer 3a: Generic Widgets (behavior)                 |
| core/button/      |  core/checkbox/   |  primitives/         |
| core/radio/       |  core/textfield/  |  Box, Text, Image    |
| core/dropdown/    |  Widget, Painter  |  ThemeScope          |
+-------------------+-------------------+----------------------+
|         Layer 2: Component Development Kit                   |
| cdk/              |                                          |
| Content[C]        |  (future: clickable, hoverable, overlay) |
+-------------------+------------------------------------------+
|         Layer 1: Foundation                                  |
| widget/                              |  event/               |
| Widget, WidgetBase, Context, Canvas  |  Mouse, Key, Wheel    |
| Focusable, ThemeProvider, Color      |  Focus, Modifiers     |
+--------------------------------------+-----------------------+
| geometry/                                                    |
| Point, Size, Rect, Constraints, Insets                       |
+==============================================================+
|                    Infrastructure                            |
| focus/           |  layout/          |  state/               |
| Focus Manager    |  Flex, Stack, Grid|  Signals, Binding     |
| (delegation)     |  (public API)     |  Computed, Effect     |
+------------------+-------------------+-----------------------+
| a11y/            |  registry/        |  plugin/              |
| Accessible       |  Widget Registry  |  Plugin System        |
| Node, Tree, Role |  Categories       |  Manager, Assets      |
+------------------+-------------------+-----------------------+
| overlay/         |  render/          |  app/                 |
| Stack, Container |  Canvas factory   |  App, Window,         |
| Position         |  (wraps internal) |  EventBridge          |
+------------------+-------------------+-----------------------+
|                 Internal Implementation                      |
| internal/render  |  internal/layout  |  internal/focus       |
| Canvas (gg)      |  Flex, Stack,     |  Manager, Ring,       |
| Renderer         |  Grid, Engine     |  Traversal, Shortcut  |
+------------------+-------------------+-----------------------+
|                 External Dependencies                        |
| gogpu/gg         |  gogpu/gpucontext |  coregx/signals       |
| 2D Graphics      |  Window/Platform  |  Reactive State       |
+------------------+-------------------+-----------------------+
```

---

## Package Structure

### Layer 1: Foundation

| Package | Purpose | Key Types |
|---------|---------|-----------|
| `widget/` | Core widget abstractions | `Widget`, `WidgetBase`, `Context`, `Canvas`, `Focusable`, `ThemeProvider`, `Color` |
| `event/` | Input event types | `MouseEvent`, `KeyEvent`, `FocusEvent`, `WheelEvent`, `Modifiers` |
| `geometry/` | Geometric primitives | `Point`, `Size`, `Rect`, `Constraints`, `Insets` |

### Layer 2: CDK (Component Development Kit)

| Package | Purpose | Key Types |
|---------|---------|-----------|
| `cdk/` | Headless behaviors, polymorphic content | `Content[C]`, `StringContent`, `FuncContent[C]`, `WidgetContent` |

### Layer 3a: Generic Widgets

| Package | Purpose | Key Types |
|---------|---------|-----------|
| `core/button/` | Button widget (behavior + Painter) | `Widget`, `Painter`, `PaintState`, `ButtonColorScheme`, `DefaultPainter` |
| `core/checkbox/` | Checkbox widget (toggle + Painter) | `Widget`, `Painter`, `PaintState`, `DefaultPainter` |
| `core/radio/` | Radio group widget (selection + Painter) | `Group`, `Item`, `Painter`, `PaintState`, `DefaultPainter` |
| `core/textfield/` | Text input widget (cursor, selection, clipboard) | `Widget`, `Painter`, selection, validation |
| `core/dropdown/` | Dropdown/select widget (overlay menu) | `Widget`, `Painter`, keyboard nav, scroll |
| `primitives/` | Display-only widgets | `BoxWidget`, `TextWidget`, `ImageWidget`, `ThemeScope` |

### Layer 3b: Design Systems

| Package | Purpose | Key Types |
|---------|---------|-----------|
| `theme/material3/` | M3 design tokens + painters | `Theme`, `ButtonPainter`, `CheckboxPainter`, `RadioPainter`, `TextFieldPainter`, `DropdownPainter`, `ColorScheme`, `TypeScale`, `ShapeScale` |

### Infrastructure

| Package | Purpose | Key Types |
|---------|---------|-----------|
| `overlay/` | Overlay/popup infrastructure | `Stack`, `Container`, `Position` |
| `focus/` | Focus management (public API) | `Manager`, `Shortcut`, `DrawFocusRing` |
| `layout/` | Layout tree and algorithms | `NodeID`, `NodeLayout`, `Result`, `Algorithm` |
| `state/` | Reactive state (signals integration) | `Signal`, `ReadonlySignal`, `Computed`, `Effect`, `Binding`, `Scheduler` |
| `theme/` | Base theme system | `Theme`, `ColorPalette`, `Typography`, `SpacingScale`, `ShadowStyles`, `RadiusScale` |
| `a11y/` | Accessibility | `Accessible`, `Node`, `NodeID`, `Role`, `State`, `Action`, `Tree` |
| `app/` | Window integration | `App`, `Window`, `EventBridge`, `FrameStats` |
| `registry/` | Widget registry | `Registry`, `Category`, widget/context/canvas type aliases |
| `plugin/` | Plugin system | `Plugin`, `Manager`, `PluginContext`, `Dependency`, `AssetLoader` |
| `render/` | Public Canvas factory | `NewCanvas` (wraps internal/render) |

### Internal Packages

| Package | Purpose | Key Types |
|---------|---------|-----------|
| `internal/render/` | Canvas and Renderer backed by gg | `Canvas`, `Renderer`, `SoftwareTarget`, `RenderConfig` |
| `internal/layout/` | Layout engines | `FlexContainer`, `VStack`, `HStack`, `GridContainer`, `Engine` |
| `internal/focus/` | Focus manager implementation | `Manager`, `Shortcut`, `DrawFocusRing`, traversal helpers |

---

## Core Concepts

### Widget Interface

The `Widget` interface (`widget/widget.go`) defines the three-phase lifecycle for all UI elements:

```go
// widget/widget.go
type Widget interface {
    Layout(ctx Context, constraints geometry.Constraints) geometry.Size
    Draw(ctx Context, canvas Canvas)
    Event(ctx Context, e event.Event) bool
    Children() []Widget
}
```

- **Layout** -- Calculate size given constraints from the parent. Containers layout their children and set child bounds.
- **Draw** -- Render to a canvas. Called after layout when bounds are established.
- **Event** -- Handle user input. Returns true if the event was consumed.
- **Children** -- Return child widgets in z-order. Leaf widgets return nil.

There is no Prepaint/Paint two-phase rendering. Drawing happens in a single Draw pass.

### WidgetBase

`WidgetBase` (`widget/base.go`) provides common functionality via embedding. Fields are plain types, not signals:

```go
// widget/base.go
type WidgetBase struct {
    mu       sync.RWMutex
    bounds   geometry.Rect  // Cached layout bounds
    focused  bool           // Whether widget has focus
    visible  bool           // Whether widget is visible
    enabled  bool           // Whether widget accepts input
    id       string         // Optional ID for debugging
    children []Widget       // Child widgets
    parent   Widget         // Parent widget (if any)
}
```

All state access is protected by a `sync.RWMutex`. WidgetBase provides:

- Bounds tracking (`Bounds`, `SetBounds`, `Size`, `Position`)
- Focus state (`IsFocused`, `SetFocused`)
- Visibility and enabled state (`IsVisible`, `SetVisible`, `IsEnabled`, `SetEnabled`)
- Child management (`AddChild`, `RemoveChild`, `InsertChild`, `ClearChildren`, `ChildAt`, `ChildCount`)
- Hit testing (`ContainsPoint`)
- Coordinate conversion (`LocalToGlobal`, `GlobalToLocal`)

Defaults: visible = true, enabled = true.

### Context Interface

`Context` (`widget/context.go`) is passed through the widget tree during all phases:

```go
// widget/context.go
type Context interface {
    RequestFocus(w Widget)
    ReleaseFocus(w Widget)
    IsFocused(w Widget) bool
    FocusedWidget() Widget
    Now() time.Time
    DeltaTime() time.Duration
    Invalidate()
    InvalidateRect(r geometry.Rect)
    Cursor() CursorType
    SetCursor(cursor CursorType)
    Scale() float32
    ThemeProvider() ThemeProvider
}
```

The concrete implementation `ContextImpl` provides thread-safe focus management, time tracking, invalidation callbacks, cursor state, and theme access. It also supports:

- `SetNow(time.Time)` -- Updates time and computes delta (called per-frame)
- `IsInvalidated() bool` / `ClearInvalidation()` -- Frame-level dirty tracking
- `ResetCursor()` -- Resets cursor to default at frame start
- `SetOnInvalidate(func())` -- Callback when `Invalidate()` is called
- `SetThemeProvider(ThemeProvider)` -- Sets the active theme provider (wired by `app/window.go`)

### Canvas Interface

`Canvas` (`widget/canvas.go`) provides drawing operations. It uses `Color` directly, not style structs:

```go
// widget/canvas.go
type Canvas interface {
    Clear(color Color)
    DrawRect(r geometry.Rect, color Color)
    StrokeRect(r geometry.Rect, color Color, strokeWidth float32)
    DrawRoundRect(r geometry.Rect, color Color, radius float32)
    StrokeRoundRect(r geometry.Rect, color Color, radius float32, strokeWidth float32)
    DrawCircle(center geometry.Point, radius float32, color Color)
    StrokeCircle(center geometry.Point, radius float32, color Color, strokeWidth float32)
    DrawLine(from, to geometry.Point, color Color, strokeWidth float32)
    DrawText(text string, bounds geometry.Rect, fontSize float32, color Color, bold bool, align float32)
    PushClip(r geometry.Rect)
    PopClip()
    PushTransform(offset geometry.Point)
    PopTransform()
}
```

Key design decisions:
- `DrawText` takes bounds, fontSize, color, bold flag, and alignment (0=left, 0.5=center, 1=right)
- Clip and transform use push/pop stacks (not Save/Restore)
- PushTransform applies a translation offset (not a full matrix)

### Focusable Interface

`Focusable` (`widget/focusable.go`) is an opt-in interface for widgets that accept keyboard focus:

```go
// widget/focusable.go
type Focusable interface {
    IsFocusable() bool
    SetFocused(focused bool)
    IsFocused() bool
}
```

`WidgetBase` already implements `SetFocused` and `IsFocused`. Concrete widgets only need to implement `IsFocusable()` to opt in.

### Color Type

`Color` is defined in `widget/canvas.go` (not a separate `color.go` file):

```go
type Color struct {
    R, G, B, A float32
}
```

Constructors: `RGBA`, `RGB`, `RGBA8`, `RGB8`, `Hex`, `HexA`. Methods: `WithAlpha`, `Lerp`, `IsOpaque`, `IsTransparent`, `RGBA8`. Predefined constants: `ColorBlack`, `ColorWhite`, `ColorRed`, `ColorGreen`, `ColorBlue`, etc.

---

## Event System

### Event Interface

All events implement `event.Event` (`event/event.go`):

```go
type Event interface {
    Type() Type
    Time() time.Time
    Handled() bool
    SetHandled()
    Modifiers() Modifiers
}
```

The `Base` struct provides the common implementation. Events use pointer receivers so `SetHandled()` works correctly.

### Event Types

| Type Enum | Struct | Sub-types |
|-----------|--------|-----------|
| `TypeMouse` | `MouseEvent` | `MousePress`, `MouseRelease`, `MouseMove`, `MouseEnter`, `MouseLeave`, `MouseDrag`, `MouseDoubleClick` |
| `TypeKey` | `KeyEvent` | `KeyPress`, `KeyRelease`, `KeyRepeat` |
| `TypeFocus` | `FocusEvent` | `FocusGained`, `FocusLost` |
| `TypeWheel` | `WheelEvent` | (scroll delta in X/Y) |
| `TypeTouch` | -- | Defined but not yet implemented |
| `TypeText` | -- | Defined but not yet implemented |
| `TypeDrop` | -- | Defined but not yet implemented |
| `TypeResize` | -- | Defined but not yet implemented |

### MouseEvent

```go
type MouseEvent struct {
    Base
    MouseType      MouseEventType
    Button         Button
    Buttons        ButtonState
    Position       geometry.Point
    GlobalPosition geometry.Point
    ClickCount     int
}
```

Mouse buttons: `ButtonNone`, `ButtonLeft`, `ButtonRight`, `ButtonMiddle`, `ButtonX1`, `ButtonX2`.
Button state is a bitmask with `ButtonStateLeft`, `ButtonStateRight`, etc.

### KeyEvent

```go
type KeyEvent struct {
    Base
    KeyType  KeyEventType
    Key      Key
    Rune     rune
    ScanCode uint32
}
```

Comprehensive key code constants: letters (A-Z), digits (0-9), function keys (F1-F24), navigation, editing, modifiers, numpad, symbols, and media keys.

### Modifiers

```go
type Modifiers uint8

const (
    ModNone     Modifiers = 0
    ModShift    Modifiers = 1 << iota
    ModCtrl
    ModAlt
    ModSuper
    ModCapsLock
    ModNumLock
)
```

Methods: `Has`, `HasAny`, `IsShift`, `IsCtrl`, `IsAlt`, `IsSuper`, `With`, `Without`.

### Event Propagation

Events are dispatched from the root widget down through the tree. A widget's `Event` method returns `true` to consume the event and stop propagation. There is no explicit capture/bubble phase -- widgets check bounds and delegate to children as appropriate.

---

## Button Widget

The `core/button/` package is the first concrete widget implementation. It demonstrates the 3-layer architecture: generic widget behavior in `core/button/`, pluggable visual styling via the `Painter` interface, and Material 3 styling in `theme/material3/`.

### Pluggable Painter Pattern

The button defines a `Painter` interface that design systems implement. This separates behavior (click handling, focus, states) from visual rendering (colors, shapes, typography):

```go
// core/button/painter.go
type Painter interface {
    PaintButton(canvas widget.Canvas, state PaintState)
}

// theme/material3/button.go
type ButtonPainter struct {
    Theme *Theme  // nil = default M3 purple fallback
}
var _ button.Painter = ButtonPainter{}
```

If no painter is set on the button, `DefaultPainter` (a minimal gray style) is used.

### Color Resolution Chain

Colors flow through a 4-level priority chain:

```
1. Explicit override (SetBackground)     →  state.Background
2. PaintState.ColorScheme (non-zero)     →  pre-resolved colors
3. Painter.resolveColors() from Theme    →  Theme.Colors → ButtonColorScheme
4. Painter built-in defaults             →  m3DefaultColors or gray fallback
```

`ButtonColorScheme` is a value struct with 10 color fields (FilledBg, FilledFg, OutlinedBorder, etc.) that carries theme-derived colors. It lives in `core/button/` and uses only `widget.Color`, so `core/button/` never imports `theme/material3/`.

### Construction with Functional Options

```go
import "github.com/gogpu/ui/core/button"
import "github.com/gogpu/ui/theme/material3"

// Generic button with M3 styling
m3 := material3.New(widget.Hex(0x6750A4))
btn := button.New(
    button.Text("Submit"),
    button.OnClick(func() { /* ... */ }),
    button.VariantOpt(button.Filled),
    button.SizeOpt(button.Large),
    button.PainterOpt(material3.ButtonPainter{Theme: m3}),
)
```

Available options: `Text`/`TextOpt`, `TextFn`, `OnClick`, `Disabled`, `DisabledFn`, `VariantOpt`, `SizeOpt`, `PainterOpt`, `A11yHint`, `BackgroundOpt`/`Background`, `RoundedOpt`/`Rounded`.

### Fluent Styling Methods

```go
btn.Padding(16).SetBackground(color).SetRounded(8).MinWidth(200)
```

Methods: `Padding`, `PaddingXY`, `SetBackground`, `SetRounded`, `MinWidth`, `MaxWidth`.

### Variants and Sizes

Variants: `Filled` (default, solid background), `Outlined` (border only), `TextOnly` (text only, hover highlight), `Tonal` (tinted background).

Sizes: `Small` (32px height, 12px font), `Medium` (40px height, 14px font), `Large` (48px height, 16px font).

### Interaction States

Internal `interactionState`: `stateNormal`, `stateHover`, `statePressed`. Colors are adjusted using `Lerp` for hover (lighten 10%) and pressed (darken 15%).

### Config with Dynamic Resolution

The button config supports both static values and dynamic functions:

```go
type config struct {
    text       string
    textFn     func() string        // Takes precedence over text
    disabled   bool
    disabledFn func() bool          // Takes precedence over disabled
    onClick    func()
    variant    Variant
    size       Size
    painter    Painter              // nil = DefaultPainter
    // ...
}
```

`ResolvedText()` and `ResolvedDisabled()` prefer the function over the static value.

### Interface Compliance

```go
var _ widget.Widget    = (*Widget)(nil)
var _ widget.Focusable = (*Widget)(nil)
```

The button is a leaf widget (`Children()` returns nil). It embeds `widget.WidgetBase` and implements `IsFocusable()` as `IsVisible() && IsEnabled() && !ResolvedDisabled()`.

### File Organization

| File | Responsibility |
|------|----------------|
| `core/button/widget.go` | Widget struct, New, Layout, Draw, Event, Children |
| `core/button/options.go` | Functional option types |
| `core/button/button.go` | Convenience aliases (Text, Background, Rounded) |
| `core/button/config.go` | Internal config struct with resolution logic |
| `core/button/variants.go` | Variant/Size enums and size constants |
| `core/button/paint.go` | Drawing helpers, color palette, focus ring |
| `core/button/painter.go` | Painter interface, PaintState, ButtonColorScheme, DefaultPainter |
| `core/button/event.go` | Mouse and keyboard event handling |
| `core/button/styling.go` | Fluent styling methods |

---

## Checkbox Widget

The `core/checkbox/` package implements a toggleable checkbox with three visual states: unchecked, checked, and indeterminate. Like the button, it uses the pluggable `Painter` pattern for design-system-agnostic rendering.

### Check States

- **Unchecked** (default) -- empty box with a border
- **Checked** -- filled box with a checkmark
- **Indeterminate** -- filled box with a horizontal dash (for "select all" scenarios)

### Construction

```go
cb := checkbox.New(
    checkbox.LabelOpt("Accept terms"),
    checkbox.OnToggle(func(checked bool) { /* ... */ }),
    checkbox.Checked(true),
    checkbox.Disabled(false),
)
```

### Interaction

- Mouse click (left button) toggles checked state
- Space key toggles when focused
- Disabled checkboxes ignore all interaction
- Implements `widget.Focusable` for Tab navigation

---

## Radio Group Widget

The `core/radio/` package implements a mutually exclusive radio group with configurable layout direction (vertical or horizontal) and arrow key navigation.

### Construction

```go
rg := radio.NewGroup(
    radio.Items(
        radio.ItemDef{Value: "s", Label: "Small"},
        radio.ItemDef{Value: "m", Label: "Medium"},
        radio.ItemDef{Value: "l", Label: "Large"},
    ),
    radio.Selected("m"),
    radio.OnChange(func(v string) { /* ... */ }),
    radio.DirectionOpt(radio.Horizontal),
)
```

### Layout Direction

- `Vertical` (default) -- items stacked top-to-bottom, Up/Down arrow keys
- `Horizontal` -- items placed left-to-right, Left/Right arrow keys

### Interaction

- Mouse click selects an item and deselects the previous one
- Arrow keys navigate between items within the group
- Space/Enter on a focused item selects it
- Individual items implement `widget.Focusable` for Tab navigation

---

## Focus Management

### Delegation Pattern

Focus management uses a public/internal delegation pattern:

- `focus/` (public) -- Thin wrapper that delegates to `internal/focus/`
- `internal/focus/` -- Full implementation (Manager, traversal, shortcuts)

```go
// focus/focus.go (public)
type Manager struct {
    impl *ifocus.Manager
}

func (m *Manager) Focus(w widget.Focusable) { m.impl.Focus(w) }
func (m *Manager) Blur()                     { m.impl.Blur() }
func (m *Manager) Next()                     { m.impl.Next() }
func (m *Manager) Previous()                 { m.impl.Previous() }
func (m *Manager) HandleKeyEvent(e *event.KeyEvent) bool { return m.impl.HandleKeyEvent(e) }
```

### Internal Focus Manager

`internal/focus/Manager` tracks:
- `root widget.Widget` -- Widget tree root for traversal
- `focused widget.Focusable` -- Currently focused widget
- `shortcuts []shortcutEntry` -- Registered keyboard shortcuts

Tab order is depth-first traversal. The manager collects focusable widgets by recursively walking the tree, skipping invisible and disabled subtrees.

### Key Event Handling

Priority order in `HandleKeyEvent`:
1. Registered keyboard shortcuts (on KeyPress only)
2. Tab key -- next focusable widget
3. Shift+Tab -- previous focusable widget

### Shortcuts

```go
// focus/shortcut.go
type Shortcut struct {
    Key   event.Key
    Ctrl  bool
    Shift bool
    Alt   bool
}
```

### Focus Ring Drawing

```go
focus.DrawFocusRing(canvas, bounds, color, radius)
```

Draws a rounded rectangle outline offset by `DefaultFocusRingOffset` (2px) with `DefaultFocusRingStrokeWidth` (2px).

---

## Rendering Pipeline

### Render Loop

The frame cycle in `app/Window.Frame()`:

```
1. Update time (ctx.SetNow)
2. Reset cursor to default
3. Flush pending state changes (scheduler.Flush)
4. Update DPI scale factor
5. Update window size from provider
6. Layout pass (if needsLayout flag is set)
   - Create tight constraints from window size
   - Call root.Layout(ctx, constraints)
   - Set root bounds
7. Draw pass
   - Call root.Draw(ctx, canvas) via DrawTo
8. Sync cursor to platform
9. Clear invalidation flags
10. Report frame statistics (if callback set)
```

There is no Prepaint pass. The render loop is two-phase: Layout then Draw.

### Canvas Implementation

`internal/render/Canvas` wraps `gg.Context` (gogpu/gg 2D rasterizer):

- Manages clip stack and transform stack internally
- Clip intersection computed manually; visibility checked per draw call
- Transform is translation-only (offset accumulation)
- Text rendering uses Go standard fonts (goregular/gobold) via `gg/text.FontSource`
- Color conversion: `widget.Color` (float32) to `gg.RGBA` (float64) via `ToGGColor`/`FromGGColor`

### Renderer

`internal/render/Renderer` manages the frame lifecycle:

- `BeginFrame(background)` -- Resets canvas, clears with background color, returns Canvas
- `EndFrame()` -- Returns the gg.Context for image extraction
- `Resize(w, h)` -- Recreates context and canvas on size change

### Public Canvas Factory

`render/` package provides a public factory:

```go
canvas := render.NewCanvas(ggContext, width, height)
```

This wraps `internal/render.NewCanvas` and returns a `widget.Canvas`.

---

## State Management

The `state/` package wraps `coregx/signals` for reactive state management.

### Signal

```go
count := state.NewSignal(0)
count.Set(5)
fmt.Println(count.Get()) // 5
```

`Signal[T]` and `ReadonlySignal[T]` are type aliases for `signals.Signal[T]` and `signals.ReadonlySignal[T]`.

### Computed

```go
fullName := state.NewComputed(func() string {
    return firstName.Get() + " " + lastName.Get()
}, firstName.AsReadonly(), lastName.AsReadonly())
```

Lazy evaluation with memoization. Dependencies must be passed explicitly.

### Effect

```go
eff := state.NewEffect(func() {
    fmt.Println("count is", count.Get())
}, count.AsReadonly())
defer eff.Stop()
```

Runs immediately and re-runs on dependency changes. `NewEffectWithCleanup` supports cleanup callbacks.

### Binding

`Binding` connects a signal to a widget's invalidation lifecycle:

```go
binding := state.Bind(sig, ctx)
defer binding.Unbind()
```

When the signal changes, the widget's context is invalidated, marking it for re-render.

### Scheduler

`Scheduler` batches widget re-render requests:

```go
sched := state.NewScheduler(func(dirty []widget.Widget) {
    // Re-render dirty widgets
})
sched.MarkDirty(widget)
sched.Flush() // Calls flush function with deduplicated widget list
```

Supports explicit batching via `Batch` method. Instance-based (no global state), thread-safe.

### Widget Signal Bindings

All core widgets support reactive signal bindings via the `PropertySignal` naming pattern.
Signal values take highest priority over dynamic functions (`Fn`) and static values.

**One-way bindings** (widget reads from signal):
```go
label := state.NewSignal("Click me")
btn := button.New(
    button.TextSignal(label),
    button.DisabledSignal(state.NewSignal(false)),
)
label.Set("Updated!") // Button text updates on next draw
```

**Two-way bindings** (widget reads and writes back):
```go
checked := state.NewSignal(false)
cb := checkbox.New(
    checkbox.CheckedSignal(checked),
    checkbox.OnToggle(func(v bool) {
        fmt.Println("toggled to", v)
    }),
)
// User clicks checkbox → checked signal updated
// checked.Set(true) → checkbox updates on next draw
```

**Available signal options:**

| Widget | Option | Type | Binding |
|--------|--------|------|---------|
| `button` | `TextSignal` | `Signal[string]` | one-way |
| `button` | `DisabledSignal` | `Signal[bool]` | one-way |
| `checkbox` | `CheckedSignal` | `Signal[bool]` | two-way |
| `checkbox` | `LabelSignal` | `Signal[string]` | one-way |
| `checkbox` | `DisabledSignal` | `Signal[bool]` | one-way |
| `radio` | `SelectedSignal` | `Signal[string]` | two-way |
| `radio` | `GroupDisabledSignal` | `Signal[bool]` | one-way |
| `textfield` | `ValueSignal` | `Signal[string]` | two-way |
| `dropdown` | `SelectedSignal` | `Signal[int]` | two-way |
| `primitives/text` | `ContentSignal` | `ReadonlySignal[string]` | one-way |

Priority resolution: Signal > Fn > Static.

---

## Theme System

### Architecture

The theme system has three interconnected layers:

```
widget/theme.go          ThemeProvider interface (IsDark)
        ↑                         ↑ implements
theme/                   Base theme (ColorPalette, Typography, Spacing)
        ↑                         ↑ extends
theme/material3/         M3 design tokens (ColorScheme from HCT seed)
        ↓                         ↓ produces
core/button/             ButtonColorScheme (10 color fields)
```

**Import direction:** `theme/material3/` imports `core/button/` and `widget/`. Neither `core/button/` nor `widget/` imports any theme package. This avoids import cycles.

### ThemeProvider Interface

`widget/theme.go` defines a minimal interface that concrete themes implement:

```go
// widget/theme.go
type ThemeProvider interface {
    IsDark() bool
}
```

The `Context` interface exposes `ThemeProvider()` so widgets can query the active theme. `ContextImpl` stores the provider and wires it via `SetThemeProvider()`, called by `app/window.go` when the app's theme is set or changed at runtime.

### Theme Struct

```go
// theme/theme.go
type Theme struct {
    Name       string
    Mode       ThemeMode          // ModeLight, ModeDark, ModeSystem
    Colors     ColorPalette
    Typography Typography
    Spacing    SpacingScale
    Shadows    ShadowStyles
    Radii      RadiusScale
    Extensions map[string]any     // Simple key-value extensions
    typedExts  *typedExtensions   // Type-safe ThemeExtension instances
}
```

Themes are created with `theme.New(name, mode)`, `theme.DefaultLight()`, or `theme.DefaultDark()`.

Functional methods: `Clone`, `WithName`, `WithMode`, `WithColors`, `WithTypography`, `WithSpacing`, `WithShadows`, `WithRadii`, `ScaleTypography`, `ScaleSpacing`, `Compact`, `Comfortable`.

### Extensions

Two extension mechanisms:
1. `map[string]any` -- Simple key-value storage via `SetExtension`/`GetExtension`
2. `ThemeExtension` interface -- Type-safe extensions with `Lerp`, `Merge`, `CopyWith` support via `RegisterExtension`/`TypedExtension`

Extension merging and interpolation enable theme inheritance and animated transitions.

### Material 3

`theme/material3/` implements Material Design 3 (Material You):

```go
theme := material3.New(widget.Hex(0x6750A4))     // Light from seed color
theme := material3.NewDark(widget.Hex(0x6750A4))  // Dark from seed color
```

The `material3.Theme` struct contains:
- `Colors ColorScheme` -- Full M3 color scheme (29 roles) derived from seed color via HCT color science
- `Typography TypeScale` -- M3 type scale (15 roles)
- `Shape ShapeScale` -- M3 corner radius scale (7 levels)

Color generation uses HCT (Hue, Chroma, Tone) to derive primary, secondary, tertiary, neutral, and error palettes from a single seed color. The palette generator is in `theme/material3/palette.go` and `theme/material3/hct.go`.

### Theme → Widget Color Flow

`material3.ButtonPainter` holds a `*Theme` field. At paint time it calls `resolveColors()` which maps M3 `ColorScheme` roles to `ButtonColorScheme` fields:

```go
// theme/material3/button.go
func (p ButtonPainter) resolveColors() button.ButtonColorScheme {
    cs := p.Theme.Colors
    return button.ButtonColorScheme{
        FilledBg:       cs.Primary,
        FilledFg:       cs.OnPrimary,
        OutlinedBorder: cs.Outline,
        TonalBg:        cs.SecondaryContainer,
        TonalFg:        cs.OnSecondaryContainer,
        // ...
    }
}
```

When `Theme` is nil, a built-in default purple palette (`m3DefaultColors`) is used as fallback. The resolved colors are passed to the button via `ButtonColorScheme` in `PaintState`, allowing the core button to remain design-system-agnostic.

Changing the seed color produces an entirely different palette -- a red seed gives red-derived primaries, a green seed gives green-derived primaries, etc. This enables "Material You" dynamic theming from a single hex value.

---

## Layout System

### Public Layout API

`layout/` provides the public layout tree abstraction:

- `NodeID` -- Identifies nodes in the layout tree
- `NodeLayout` -- Position and size output for a node
- `Result` -- Output of a layout computation
- `Algorithm` -- Interface for layout algorithm implementations
- `LayoutTree` -- Manages the node tree for algorithms to operate on
- `Style` -- Layout style properties
- `Flex`, `Stack`, `Grid` -- Public layout constructors

### Internal Layout Engines

`internal/layout/` contains the actual layout algorithms:

**FlexContainer** -- Full CSS Flexbox implementation:
- Directions: `Row`, `RowReverse`, `Column`, `ColumnReverse`
- Justify: `Start`, `End`, `Center`, `SpaceBetween`, `SpaceAround`, `SpaceEvenly`
- Align: `Start`, `End`, `Center`, `Stretch`, `Baseline`
- Wrap modes: `NoWrap`, `Wrap`, `WrapReverse`
- Per-item: `Grow`, `Shrink`, `Basis`, `AlignSelf`
- Gap and CrossGap for spacing

**VStack / HStack** -- Simplified vertical/horizontal stacking with spacing and alignment (`StackAlignStart`, `StackAlignCenter`, `StackAlignEnd`, `StackAlignStretch`).

**GridContainer** -- CSS Grid-like layout:
- Track sizing: `TrackAuto`, `TrackFixed`, `TrackFraction` (like CSS `fr` units)
- Row and column track definitions

**Engine** -- Layout orchestrator with optional caching:
- Cache keyed by element ID + constraints
- Dirty tracking for incremental updates
- Two-pass intrinsic sizing via `LayoutWithIntrinsics`
- Statistics tracking (cache hits/misses, layout calls)

---

## Accessibility

### Accessible Interface

```go
// a11y/accessible.go
type Accessible interface {
    AccessibilityRole() Role
    AccessibilityLabel() string
    AccessibilityHint() string
    AccessibilityValue() string
    AccessibilityState() State
    AccessibilityActions() []Action
}
```

Roles include `RoleButton`, `RoleSlider`, `RoleCheckbox`, etc. States include `Disabled`, `Selected`, `Expanded`, `Checked`, and numeric value ranges (`ValueMin`, `ValueMax`, `ValueNow`).

### Node and Tree

`a11y.Node` represents a single element in the accessibility tree:
- Stable `NodeID` (atomic uint64 counter)
- Role, Label, Hint, Value, State, Actions, Bounds, Children
- Thread-safe via `sync.RWMutex`

`a11y.Tree` manages the full accessibility tree with `NewNodeFromAccessible` for building nodes from widgets.

---

## Application Layer

### App

`app.App` is the entry point, bridging the widget tree with the windowing system:

```go
a := app.New(
    app.WithWindowProvider(wp),
    app.WithPlatformProvider(pp),
    app.WithEventSource(es),
    app.WithTheme(myTheme),
)
a.SetRoot(rootWidget)
a.Frame() // Called from host render loop
```

- Uses `gpucontext.WindowProvider` for window geometry and redraw requests
- Uses `gpucontext.PlatformProvider` for cursor management
- Uses `gpucontext.EventSource` for input events
- Operates in headless mode (800x600, 1x scale) when providers are nil

### Window

`app.Window` manages the widget tree for a single window:
- Layout pass with tight constraints from window size
- Draw pass via `DrawTo(canvas)`
- Event dispatch to root widget
- Focus change handling
- Cursor sync to platform
- Frame statistics reporting via `FrameCallback`

### EventBridge

`app.EventBridge` translates `gpucontext` events into `event.*` types and dispatches them to the Window.

---

## Primitives

### BoxWidget

Container that lays out children in a vertical stack:

```go
card := primitives.Box(
    primitives.Text("Title").Bold(),
    primitives.Text("Body"),
).Padding(16).Background(widget.Hex(0xFFFFFF)).Rounded(8)
```

Supports: padding, background, border, rounded corners, shadow, gap, explicit dimensions (width/height/min/max).

Implements `widget.Widget` and `a11y.Accessible`.

### TextWidget

Renders text with configurable font size, color, bold, and alignment.

### ImageWidget

Renders an image within bounds.

---

## Plugin System

The `plugin/` package provides a plugin architecture for bundling UI components:

```go
type Plugin interface {
    Name() string
    Version() string
    Dependencies() []Dependency
    Init(ctx *PluginContext) error
    Shutdown() error
}
```

- `Manager` handles registration, dependency resolution, and initialization order
- `PluginContext` provides access to widget registry, theme registry, and asset loader
- `Dependency` declares required plugins with version constraints
- `AssetLoader` handles asset management for plugins

---

## Widget Registry

The `registry/` package provides a global registry for widget factories:

- Register widget constructors by name and category
- Categories: `CategoryInput`, `CategoryDisplay`, `CategoryContainer`, `CategoryCustom`
- Type aliases for `widget.Widget`, `widget.Context`, `widget.Canvas`, etc. to simplify imports

---

## Dependencies

| Dependency | Purpose | Version |
|------------|---------|---------|
| `github.com/gogpu/gg` | 2D graphics backend for Canvas | v0.32.0 |
| `github.com/gogpu/gpucontext` | Window/Platform provider interfaces | v0.9.0 |
| `github.com/coregx/signals` | Reactive state management | v0.1.0 |
| `golang.org/x/image` | Standard Go fonts (goregular, gobold) | v0.36.0 |

Go version: **1.25.0**

---

## Design Principles

### 1. Composition over Inheritance

Widgets embed `WidgetBase` for shared functionality. Optional interfaces (`Focusable`, `Accessible`) are checked via type assertion:

```go
if f, ok := w.(widget.Focusable); ok && f.IsFocusable() {
    // Widget supports focus
}
```

### 2. Functional Options for Construction

Widgets use the functional options pattern:

```go
btn := button.New(
    button.Text("Submit"),
    button.OnClick(handleSubmit),
    button.VariantOpt(button.Filled),
)
```

### 3. Interface-Driven Architecture

Core abstractions are interfaces (`Widget`, `Context`, `Canvas`, `Focusable`, `Accessible`, `Plugin`). This enables testing with mocks and alternative implementations.

### 4. Delegation for Internal Complexity

Public packages provide clean APIs while delegating to internal implementations:
- `focus/` delegates to `internal/focus/`
- `render/` delegates to `internal/render/`
- `layout/` delegates to `internal/layout/`

### 5. Pluggable Painters for Design System Independence

Generic widgets in `core/` define behavior and delegate visual rendering to a `Painter` interface. Each design system provides its own Painter:

```go
// core/button/   defines: Painter interface + ButtonColorScheme value struct
// core/checkbox/ defines: Painter interface + PaintState
// core/radio/    defines: Painter interface + PaintState
// theme/material3/ provides: ButtonPainter, CheckboxPainter, RadioPainter
```

This lets the same widget render as Material 3, Fluent, or Cupertino by swapping the Painter. Colors flow as a value struct (`ButtonColorScheme`) -- no import cycle between `core/` and `theme/`.

### 6. Thread Safety via Mutexes

`WidgetBase` and `ContextImpl` use `sync.RWMutex` for state protection. Canvas and Renderer are NOT thread-safe and must be used from the UI thread.

### 7. Value Semantics for Geometry

All types in `geometry/` are small structs passed by value. Operations return new values without modifying the receiver. No heap allocations in hot paths.

---

*This document reflects the actual codebase as of March 1, 2026.*
