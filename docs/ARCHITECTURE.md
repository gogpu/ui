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
| theme/material3/  |  theme/fluent/    |  theme/cupertino/    |
| 21 Painters       |  9 Painters       |  9 Painters          |
| (M3 HCT colors)   |  (Acrylic/Mica)  |  (Apple HIG)          |
+-------------------+-------------------+----------------------+
|         Layer 3a: Generic Widgets (behavior)                 |
| core/button/      |  core/checkbox/   |  primitives/         |
| core/radio/       |  core/textfield/  |  Box (HBox/VBox),    |
| core/dropdown/    |  core/slider/     |  Text, Image,        |
| core/dialog/      |  core/scrollview/ |  ThemeScope,         |
| core/tabview/     |  core/listview/   |  RepaintBoundary     |
| core/gridview/    |  core/linechart/  |                      |
| core/progressbar/ |  core/progress/   |  22 interactive      |
| core/collapsible/ |  core/popover/    |  widgets in core/    |
| core/splitview/   |  core/treeview/   |                      |
| core/datatable/   |  core/toolbar/    |                      |
| core/menu/        |  core/docking/    |                      |
+-------------------+-------------------+----------------------+
|         Layer 2: Component Development Kit                   |
| cdk/              |                                          |
| Content[C]        |  (future: clickable, hoverable, overlay) |
+-------------------+------------------------------------------+
|         Layer 1: Foundation                                  |
| widget/                              |  event/               |
| Widget, WidgetBase, Context, Canvas  |  Mouse, Key, Wheel    |
| Focusable, Lifecycle, SchedulerRef   |  Focus, Modifiers     |
+--------------------------------------+-----------------------+
| geometry/                                                    |
| Point, Size, Rect, Constraints, Insets                       |
+==============================================================+
|                    Infrastructure                            |
| focus/           |  layout/          |  state/               |
| Focus Manager    |  Flex, Stack, Grid|  Signals, Binding     |
| (delegation)     |  (public API)     |  Scheduler, Lifecycle |
+------------------+-------------------+-----------------------+
| a11y/            |  registry/        |  plugin/              |
| Accessible       |  Widget Registry  |  Plugin System        |
| Node, Tree, Role |  Categories       |  Manager, Assets      |
+------------------+-------------------+-----------------------+
| animation/       |  transition/      |  icon/                |
| Tween, Spring,   |  Fade, Slide,     |  Vector paths,        |
| M3 Presets,      |  Scale, Show/Hide |  IconWidget,          |
| Orchestration    |  Enter/Exit       |  10 built-in icons    |
+------------------+-------------------+-----------------------+
| dnd/             |  theme/font/      |  i18n/                |
| DragSource,      |  Font Registry,   |  Locale, Bundle,      |
| DropTarget,      |  CSS weight match |  Translator,          |
| Manager          |  Family/Face      |  CLDR plural, RTL     |
+------------------+-------------------+-----------------------+
| uitest/          |                   |                       |
| MockCanvas,      |  MockContext,     |  Event factories,     |
| Widget helpers   |  Assertions       |  Reusable mocks       |
+------------------+-------------------+-----------------------+
| overlay/         |  render/          |  app/                 |
| Stack, Container |  Canvas factory   |  App, Window,         |
| Position         |  (wraps internal) |  EventBridge          |
+------------------+-------------------+-----------------------+
|                 Internal Implementation                      |
| internal/render  |  internal/layout  |  internal/focus       |
| Canvas (gg)      |  Flex, Stack,     |  Manager, Ring,       |
| SceneCanvas      |  Grid, Engine     |  Traversal, Shortcut  |
+------------------+-------------------+-----------------------+
| internal/dirty   |                   |                       |
| Region Tracker,  |  Merge algorithm, |  Partial repaints     |
| Collector        |  Full repaint FB  |                       |
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
| `widget/` | Core widget abstractions | `Widget`, `WidgetBase`, `Context`, `Canvas`, `Focusable`, `PointerCapturer` (ADR-031), `Lifecycle`, `SchedulerRef`, `ThemeProvider`, `Color` |
| `event/` | Input event types | `MouseEvent`, `KeyEvent`, `FocusEvent`, `WheelEvent`, `Modifiers` |
| `geometry/` | Geometric primitives | `Point`, `Size`, `Rect`, `Constraints`, `Insets` |

### Layer 2: CDK (Component Development Kit)

| Package | Purpose | Key Types |
|---------|---------|-----------|
| `cdk/` | Headless behaviors, polymorphic content | `Content[C]`, `StringContent`, `FuncContent[C]`, `WidgetContent` |

### Layer 3a: Generic Widgets (22 interactive widgets in core/)

| Package | Purpose | Key Types |
|---------|---------|-----------|
| `core/button/` | Button widget (behavior + Painter) | `Widget`, `Painter`, `PaintState`, `ButtonColorScheme`, `DefaultPainter` |
| `core/checkbox/` | Checkbox widget (toggle + Painter) | `Widget`, `Painter`, `PaintState`, `DefaultPainter` |
| `core/radio/` | Radio group widget (selection + Painter) | `Group`, `Item`, `Painter`, `PaintState`, `DefaultPainter` |
| `core/textfield/` | Text input widget (cursor, selection, clipboard) | `Widget`, `Painter`, selection, validation |
| `core/dropdown/` | Dropdown/select widget (overlay menu) | `Widget`, `Painter`, keyboard nav, scroll |
| `core/slider/` | Slider widget (continuous/discrete, H/V) | `Widget`, `Painter`, `PaintState`, `DefaultPainter` |
| `core/dialog/` | Modal dialog (backdrop, actions, focus trap) | `Widget`, `Painter`, `Alert`, `Confirm` |
| `core/scrollview/` | Scrollable container (V/H/both) | `Widget`, `Painter`, wheel, keyboard, drag |
| `core/tabview/` | Tabbed navigation (lazy content) | `Widget`, `Tab`, `Painter`, `DefaultPainter` |
| `core/listview/` | Virtualized list (fixed-height items, recycling) | `Widget`, `Painter`, selection, keyboard nav |
| `core/gridview/` | Virtualized 2D grid (auto-fit columns, cell recycling) | `Widget`, `Painter`, Content[C], selection |
| `core/linechart/` | Real-time line chart (multiple series, rolling window) | `Widget`, `Painter`, `Series`, thread-safe PushValue |
| `core/progressbar/` | Linear progress bar (0-100%, rounded corners) | `Widget`, `Painter`, signal binding |
| `core/progress/` | Circular progress (determinate arc + spinner) | `Widget`, `Painter`, polyline arc |
| `core/collapsible/` | Expandable section (animated expand/collapse) | `Widget`, `Painter`, Tween animation |
| `core/popover/` | Popover (click) + Tooltip (hover), 12 placements | `Widget`, `Painter`, auto-flip, overlay |
| `core/splitview/` | Resizable split panels (H/V, draggable divider) | `Widget`, `Painter`, min constraints, collapse |
| `core/treeview/` | Hierarchical tree (expand/collapse, virtualized) | `Widget`, `Painter`, indent, connector lines |
| `core/datatable/` | Sortable column table (fixed header, virtualized rows) | `Widget`, `Painter`, column sort, zebra striping |
| `core/toolbar/` | Horizontal action bar (icon buttons, separators) | `Widget`, `Painter`, spacers, custom items |
| `core/menu/` | MenuBar + ContextMenu (submenus, shortcuts) | `MenuBar`, `ContextMenu`, `Painter`, overlay |
| `core/docking/` | IDE-style dockable panels (border layout, tabbed groups) | `Host`, `Panel`, `Painter`, Dock/Undock API |
| `primitives/` | Display-only widgets + RepaintBoundary | `BoxWidget` (HBox/VBox), `TextWidget`, `ImageWidget`, `ThemeScope`, `RepaintBoundary` |

### Layer 3b: Design Systems

| Package | Purpose | Key Types |
|---------|---------|-----------|
| `theme/material3/` | M3 design tokens + 21 painters | `Theme`, `ButtonPainter`, `CheckboxPainter`, `RadioPainter`, `TextFieldPainter`, `DropdownPainter`, `SliderPainter`, `DialogPainter`, `ScrollbarPainter`, `TabViewPainter`, `ListViewPainter`, `GridViewPainter`, `LineChartPainter`, `ProgressBarPainter`, `ProgressPainter`, `CollapsiblePainter`, `PopoverPainter`, `SplitViewPainter`, `TreeViewPainter`, `DataTablePainter`, `ToolbarPainter`, `MenuPainter`, `DockingPainter`, `ColorScheme`, `TypeScale`, `ShapeScale` |
| `theme/fluent/` | Microsoft Fluent Design + 9 painters | `Theme`, accent color system, inner focus ring, 4px radii, light/dark |
| `theme/cupertino/` | Apple HIG + 9 painters | `Theme`, iOS toggle switch, segmented control, pill buttons |

### Infrastructure

| Package | Purpose | Key Types |
|---------|---------|-----------|
| `overlay/` | Overlay/popup infrastructure | `Stack`, `Container`, `Position` |
| `focus/` | Focus management (public API) | `Manager`, `Shortcut`, `DrawFocusRing` |
| `layout/` | Layout tree and algorithms | `NodeID`, `NodeLayout`, `Result`, `Algorithm` |
| `state/` | Reactive state with push-pull lifecycle | `Signal`, `ReadonlySignal`, `Computed`, `Effect`, `Binding`, `BindToSchedulerFunc`, `Scheduler` |
| `theme/` | Base theme system | `Theme`, `ColorPalette`, `Typography`, `SpacingScale`, `ShadowStyles`, `RadiusScale` |
| `theme/font/` | Font registry (CSS weight matching, W3C spec) | `Registry`, `Family`, `Face`, `Weight`, `Style` |
| `a11y/` | Accessibility | `Accessible`, `Node`, `NodeID`, `Role`, `State`, `Action`, `Tree` |
| `animation/` | Animation engine + M3 presets + orchestration | `Controller`, `To`, `SpringTo`, `Sequence`, `Parallel`, `CubicBezier`, `Stagger`, `Chain` |
| `transition/` | Widget enter/exit animations | `Wrapper`, `FadeIn`, `FadeOut`, `SlideIn`, `SlideOut`, `ScaleIn`, `ScaleOut` |
| `icon/` | Vector path icons (10 built-in Material icons) | `Icon`, `IconWidget`, `Registry`, De Casteljau cubic Bezier |
| `i18n/` | Internationalization (CLDR plural rules, RTL) | `Locale`, `Bundle`, `Translator`, `LocaleSignal` |
| `dnd/` | Drag and drop | `DragSource`, `DropTarget`, `Manager`, 5px threshold |
| `uitest/` | Testing utilities (reusable mocks) | `MockCanvas`, `MockContext`, event factories, assertions |
| `app/` | Window integration + FocusManager | `App`, `Window`, `EventBridge`, `FrameStats`, `FocusManager`, `PlatformProvider()` |
| `registry/` | Widget registry | `Registry`, `Category`, widget/context/canvas type aliases |
| `plugin/` | Plugin system | `Plugin`, `Manager`, `PluginContext`, `Dependency`, `AssetLoader` |
| `render/` | Public Canvas factory | `NewCanvas` (wraps internal/render) |
| `offscreen/` | Headless widget rendering (no GPU/window) | `Renderer`, `NewRenderer`, `WithTheme`, `WithScale`, `WithBackground` |

### Internal Packages

| Package | Purpose | Key Types |
|---------|---------|-----------|
| `internal/render/` | Canvas, SceneCanvas, FontRegistry, Renderer backed by gg | `Canvas`, `SceneCanvas`, `FontRegistry`, `Renderer`, `SoftwareTarget`, `RenderConfig` |
| `internal/layout/` | Layout engines | `FlexContainer`, `VStack`, `HStack`, `GridContainer`, `Engine` |
| `internal/focus/` | Focus manager implementation | `Manager`, `Shortcut`, `DrawFocusRing`, traversal helpers |
| `internal/dirty/` | Dirty region tracking | `Tracker`, `Collector`, merge algorithm, partial repaints |

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

`WidgetBase` (`widget/base.go`) provides common functionality via embedding:

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
    bindings []Unbinder     // Signal bindings (cleaned up on unmount)
    effects  []Stopper      // Effects (stopped on unmount)
    mounted  bool           // Whether widget is currently in the mounted tree
}
```

All state access is protected by a `sync.RWMutex`. WidgetBase provides:

- Bounds tracking (`Bounds`, `SetBounds`, `Size`, `Position`)
- Focus state (`IsFocused`, `SetFocused`)
- Visibility and enabled state (`IsVisible`, `SetVisible`, `IsEnabled`, `SetEnabled`)
- Child management (`AddChild`, `RemoveChild`, `InsertChild`, `ClearChildren`, `ChildAt`, `ChildCount`)
- Hit testing (`ContainsPoint`)
- Coordinate conversion (`LocalToGlobal`, `GlobalToLocal`)
- Signal binding lifecycle (`AddBinding`, `AddEffect`, `CleanupBindings`, `IsMounted`, `SetMounted`)

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
    OverlayManager() OverlayManager
    WindowSize() geometry.Size
    Scheduler() SchedulerRef
}

// SchedulerRef avoids circular imports between widget and state.
type SchedulerRef interface {
    MarkDirty(w Widget)
}
```

The concrete implementation `ContextImpl` provides thread-safe focus management, time tracking, invalidation callbacks, cursor state, and theme access. It also supports:

- `SetNow(time.Time)` -- Updates time and computes delta (called per-frame)
- `IsInvalidated() bool` / `ClearInvalidation()` -- Frame-level dirty tracking
- `ResetCursor()` -- Resets cursor to default at frame start
- `SetOnInvalidate(func())` -- Callback when `Invalidate()` is called
- `SetThemeProvider(ThemeProvider)` -- Sets the active theme provider (wired by `app/window.go`)
- `SetScheduler(SchedulerRef)` -- Sets the signal scheduler (wired by `app/window.go`)

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
    DrawText(text string, bounds geometry.Rect, fontSize float32, color Color, bold bool, align TextAlign)
    MeasureText(text string, fontSize float32, bold bool) geometry.Size
    DrawImage(img image.Image, at geometry.Point)
    PushClip(r geometry.Rect)
    PushClipRoundRect(r geometry.Rect, radius float32)
    PopClip()
    PushTransform(offset geometry.Point)
    PopTransform()
    TransformOffset() geometry.Point
}
```

Key design decisions:
- `DrawText` takes bounds, fontSize, color, bold flag, and alignment (`TextAlignLeft`, `TextAlignCenter`, `TextAlignRight`)
- `MeasureText` returns text dimensions without drawing (for layout calculations)
- `PushClipRoundRect` provides GPU SDF-based rounded rectangle clipping
- `DrawImage` blits cached pixel buffers (used by RepaintBoundary)
- `TransformOffset` returns the current cumulative transform offset (used by `StampScreenOrigin` for ScreenBounds)
- Clip and transform use push/pop stacks (not Save/Restore)
- PushTransform applies a translation offset (not a full matrix)

### Custom Font Support (StyledTextDrawer)

Canvas supports custom fonts via the optional `StyledTextDrawer` interface:

```go
// widget/canvas.go
type TextStyle struct {
    FontFamily string
    FontSize   float32
    Bold       bool
    Italic     bool
    Color      Color
    Align      TextAlign
}

type StyledTextDrawer interface {
    DrawStyledText(text string, bounds geometry.Rect, style TextStyle)
    MeasureStyledText(text string, style TextStyle) float32
}
```

Both `Canvas` and `SceneCanvas` implement `StyledTextDrawer`. Widgets use type assertion (`if sd, ok := canvas.(widget.StyledTextDrawer); ok { ... }`), consistent with `ArcStroker`, `SVGFiller`, `DeviceScaler`, and `TextModeController`.

Font resolution uses a global `FontRegistry` (process-singleton, RWMutex) with CSS weight matching and Inter fallback. Plugins register fonts via `ctx.Assets.LoadFont("name", data)` which auto-registers in the global registry. This follows the universal pattern from Flutter (`FontCollection`), Qt6 (`QFontDatabase`), and Iced (`FontSystem`).

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

`ResolvedText()` and `ResolvedDisabled()` use priority: ReadonlySignal > Signal > Fn > Static.

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
3. Flush pending signal changes (reflush loop, max 2 re-flushes)
   - scheduler.Flush() processes dirty widgets from signal changes
   - If flush triggers new signal changes, re-flush (up to 2 times)
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

### Widget Lifecycle (Mount/Unmount)

Widgets that use signal bindings implement the optional `Lifecycle` interface:

```go
// widget/lifecycle.go
type Lifecycle interface {
    Mount(ctx Context)   // Called when added to tree — create signal bindings
    Unmount()            // Called when removed — cleanup (bindings auto-cleaned)
}
```

The framework manages lifecycle automatically:

- `Window.SetRoot(w)` calls `UnmountTree(oldRoot)` then `MountTree(newRoot, ctx)`
- `MountTree` walks the widget tree recursively, calling `Mount(ctx)` on each `Lifecycle` implementor
- `UnmountTree` walks bottom-up, calling `CleanupBindings()` then `Unmount()` on each widget
- Widgets without signals need not implement `Lifecycle` — they are unaffected

All widget types with signal bindings implement `Lifecycle`: button, checkbox, radio, textfield, dropdown, slider, dialog, scrollview, tabview, listview, gridview, linechart, progressbar, collapsible, popover, splitview, treeview, datatable, toolbar, menu, docking, primitives/text, primitives/box.

### Retained-Mode Rendering

The framework uses a hybrid immediate/retained rendering model with three
levels of optimization:

**Level 1: Frame-level skip (implemented)**
When no widget in the tree has its `needsRedraw` flag set, `Window.DrawTo()`
returns `false` and the host application reuses the previous frame's GPU
framebuffer. This means idle UIs consume zero CPU for the draw phase.

**Level 2: Draw statistics (implemented)**
`widget.DrawTree()` performs the draw traversal and collects per-widget
`DrawStats` (dirty, clean, skipped, total counts). These stats are exposed
via `FrameStats.DrawStats` for performance monitoring and validation.

**Level 3: Per-widget pixel caching (implemented, Sub-Phase 2)**
Clean subtrees are composited from cached pixel buffers instead of re-drawn.
**RepaintBoundary** (ADR-024) is a WidgetBase property (`SetRepaintBoundary(true)`).
Each boundary has its own `scene.Scene` for display list caching.

**Level 4: Layer Tree Compositor + Damage-Aware Blit (ADR-007 Phase D+, v0.1.20)**

Enterprise retained-mode compositor with Layer Tree, per-boundary GPU textures,
persistent tree reuse, multi-rect damage, and overlay boundary pipeline:

```
desktop.draw()
  → Frame()                         signals, layout, animations
  → [O(1) FRAME SKIP]              HasDirtyBoundaries || NeedsRedraw || NeedsAnimationFrame
  → PaintBoundaryLayers()            re-record dirty+visible boundaries (Flutter flushPaint)
  → PaintOverlayBoundaries()         re-record dirty overlay content boundaries
  → UpdateLayerTree()                persistent Layer Tree (97.9% fewer allocs)
  → AppendOverlaysToLayerTree()      overlay boundaries in Layer Tree (Z-order on top)
  → CollectDirtyRegions()            dirty tracker for debug overlay
  → renderBoundaryTexturesFromTree() Layer Tree walk → per-boundary GPU textures (MSAA)
  → compositeTexturesFromTree()      Layer Tree walk → blit all textures to surface (non-MSAA)
  → DrawOverlayScrim()               modal backdrop only (non-modal = no scrim)
  → RenderDirectWithDamage()         LoadOpLoad + scissor to damage rect (damage-aware blit)
    OR canvas.Render()               LoadOpClear + full blit (when root changed or debug active)
```

**GPU performance:** 0% idle (frame skip), 10% with visible spinner at 30fps
(48x48 scissor proven at HAL level via software backend e2e tests).

**Layer Tree compositor (ADR-007 Phase D):**
The `compositor/` package provides a structured layer tree that drives the
production render loop. `OffsetLayer` positions boundaries in window coordinates.
`PictureLayer` owns a cached `scene.Scene`, `BoundaryCacheKey`, `ScreenOrigin`,
and `ClipRect`. `ClipRectLayer` provides viewport clipping for ScrollView items.
`OpacityLayer` supports alpha blending on cached textures (via gg
`DrawGPUTextureWithOpacity`). Layer Tree traversal replaces direct widget tree
walks for rendering and compositing.

**Persistent Layer Tree (ADR-007 Phase D.5):**
`UpdateLayerTree()` reuses layer objects across frames instead of rebuilding
per-frame. For 200 boundaries: 613 allocs/op down to 13 allocs/op (97.9%
reduction). Enterprise pattern validated by research across Flutter, Chrome,
Qt6, Android, and Skia -- all use persistent trees.

**O(1) frame skip (ADR-028 Phase C):**
`HasDirtyBoundaries()` checks a flat dirty boundary set instead of the
previous O(n) `NeedsRedrawInTreeNonBoundary` tree walk. 45x faster (1.2ns
vs 58ns). Flutter `_nodesNeedingPaint` pattern with `DirtyBoundaryRegistrar`
interface.

**Multi-rect damage (ADR-030):**
Per-draw dynamic scissor for multiple dirty rects. Zero pixel waste when dirty
widgets are spatially distant. Ring buffer stores rect lists per frame. Threshold
of >16 rects merges to union (GDK/Sway pattern). Full stack: ui ->
gg `RenderDirectWithDamageRects` -> wgpu `PresentWithDamage`.

**Overlay boundary pipeline (ADR-029 Phase E):**
Dropdown menus, dialogs, and other overlays rendered via the same Layer Tree and
boundary texture pipeline as main widgets. `PaintOverlayBoundaries()` re-records
dirty overlay scenes. `AppendOverlaysToLayerTree()` adds overlays after the main
tree for correct Z-order. Scrim applies only for modal overlays (Flutter
ModalBarrier pattern). `overlayAwareHitTest()` blocks hover on background widgets
when an overlay is open.

Each RepaintBoundary is rendered into its own GPU offscreen texture. Child
boundaries (depth > 0) are **skipped** during parent recording (DrawChild skip
pattern -- Flutter `paintChild`). Each child boundary gets its own GPU texture,
composed separately during Layer Tree traversal. When a child boundary is dirty,
the root re-records cheaply (child content skipped), and the child re-renders
its own texture independently.

**Offscreen boundary culling:**
`isBoundaryLayerVisible()` checks CompositorClip intersection before recording.
Offscreen animated widgets (spinner scrolled out of view) are not recorded ->
`ScheduleAnimationFrame` not called -> animation pumper stops -> 0% GPU.

**DrawChild skip pattern (Flutter paintChild):**
During `recordBoundary`, the `BoundaryRecorder` checks each child: if the child
has `IsRepaintBoundary() == true`, it is skipped (not drawn into the parent
scene). Instead, the child's GPU texture is composed at the correct position
during Layer Tree compositing with GPU scissor clipping applied per viewport
(ScrollView). This means parent re-recording is cheap -- it only draws
non-boundary children (text, backgrounds, dividers) while boundary children
retain their cached textures.

**Force root re-recording:**
`desktop.draw` checks `NeedsRedrawInTreeNonBoundary` on the root widget.
If any non-boundary descendant is dirty, root re-records. Boundary descendants
manage their own dirty state independently. The `onBoundaryDirty` callback is
suppressed during this forced invalidation to prevent restarting the animation
pumper from data tickers.

**Compositor scissor clipping:**
Items inside ScrollView viewports are clipped via GPU scissor rect during
Layer Tree compositing, not during scene recording. Each boundary group in
the blit pass has per-group scissor applied.

**Software backend e2e tests:**
The wgpu software backend (`hal/software`) provides deterministic GPU pipeline
for CI. HAL-level `RenderPassStats` proves scissor=48x48 (not full window).
Pixel-exact readback verifies damage preservation across frames. 9 e2e tests
run without GPU hardware.

**ScreenOriginBase:**
`recordBoundary` sets `ScreenOriginBase` from the boundary widget's screen
position before recording child content. This ensures nested boundaries get
correct screen-space origins for compositor texture placement (fixes nested
boundary positioning in ScrollView).

**Scrollbar track repeat (Qt6 timing):**
Track repeat uses Qt6 `QScrollBar` timing: 500ms initial delay, 50ms repeat
interval. Event-driven (no polling goroutine) to prevent root re-recording
flood.

**SVG icon rendering** uses CPU rasterization (`RasterizerAnalytic`) into
scene.Image, with a 2-level LRU IconCache (Level 1: parsed docs, Level 2:
rasterized bitmaps by ptr+size+color). DPI-aware: renders at physical pixel
size (`ceil(logical × deviceScale)`).

The dirty-tracking flow:

```
Widget state change (hover, click, signal)
  → SetNeedsRedraw(true)
    → propagateDirtyUpward(parent) → root boundary → InvalidateScene()
      → RegisterDirtyBoundary() → flat dirty set (O(1))
        → RequestRedraw()
          → desktop.draw: HasDirtyBoundaries() O(1) check
            → PaintBoundaryLayers: recordBoundary() with DrawChild skip
            → PaintOverlayBoundaries: re-record overlay content
            → UpdateLayerTree: persistent tree reuse
            → AppendOverlaysToLayerTree: overlay Z-order
            → renderBoundaryTexturesFromTree: Layer Tree → GPU textures
            → compositeTexturesFromTree: Layer Tree → blit + scissor
            → RenderDirectWithDamage: LoadOpLoad + damage rect → surface
```

Key functions:
- `PaintBoundaryLayersWithContext(root, _, ctx)` -- re-records dirty boundaries
- `PaintOverlayBoundaries(overlays, ctx)` -- re-records dirty overlay boundaries
- `UpdateLayerTree(root, tree)` -- persistent Layer Tree update (reuses layers)
- `AppendOverlaysToLayerTree(overlays, tree)` -- overlays after main tree
- `renderBoundaryTexturesFromTree(tree, cc)` -- Layer Tree walk -> GPU textures
- `compositeTexturesFromTree(tree, cc, w, h)` -- Layer Tree walk -> blit + scissor
- `HasDirtyBoundaries()` -- O(1) flat dirty set check for frame skip
- `recordBoundary(w, ctx)` -- records scene with DrawChild skip for child boundaries
- `widget.ClearRedrawInTree(w)` -- clears all flags recursively
- `widget.MarkRedrawInTree(w)` -- marks all widgets dirty (used by resize, theme change)
- `widget.NeedsRedrawInTree(w)` -- checks if any descendant needs redraw

### Canvas Implementation

`internal/render/Canvas` wraps `gg.Context` (gogpu/gg 2D rasterizer):

- Manages clip stack and transform stack internally
- Clip intersection computed manually; visibility checked per draw call
- Transform is translation-only (offset accumulation)
- Text rendering uses Inter font by default (Regular/Bold) via `gg/text.FontSource`
- Custom fonts resolved via global `FontRegistry` (CSS weight matching, `*text.FontSource` caching)
- Implements `StyledTextDrawer` for custom font rendering alongside standard `DrawText`
- Color conversion: `widget.Color` (float32) to `gg.RGBA` (float64) via `ToGGColor`/`FromGGColor`

### Renderer

`internal/render/Renderer` manages the frame lifecycle:

- `BeginFrame(background)` -- Resets canvas, clears with background color, returns Canvas
- `EndFrame()` -- Returns the gg.Context for image extraction
- `Resize(w, h)` -- Recreates context and canvas on size change

### SceneCanvas (Tile-Parallel Rendering)

`internal/render/SceneCanvas` implements `widget.Canvas` by recording commands
into a `scene.Scene` instead of executing them immediately:

- Shape operations (rect, round rect, circle, line) map to `scene.Shape` types
- Text is rendered via gg.Context pass-through, captured as pixels, added as `scene.Image`
- Clip and transform stacks mirror Canvas behavior for visibility optimization
- Used by `RepaintBoundary.renderWithScene()` for large widget subtrees
- `scene.Renderer` rasterizes the scene tile-parallel (64x64 tiles, goroutine worker pool)
- Result is `gg.Pixmap.ToImage()` -> `*image.RGBA` for cache compositing

### Public Canvas Factory

`render/` package provides a public factory:

```go
canvas := render.NewCanvas(ggContext, width, height)
```

This wraps `internal/render.NewCanvas` and returns a `widget.Canvas`.

### Offscreen Rendering

The `offscreen/` package renders widget trees into `*image.RGBA` without a GPU,
window, or running application. It uses CPU-only rasterization via `gg.NewContext`:

```go
r := offscreen.NewRenderer(400, 120)
r.Render(primitives.Text("Hello").FontSize(24))
img := r.Image() // *image.RGBA
```

Options: `WithTheme` (default: M3 light), `WithScale` (HiDPI), `WithBackground`.

Internally, `offscreen.Renderer.Render()` creates a `widget.ContextImpl`, runs
`Layout()` to size the widget tree, then `widget.DrawTree()` to render it
(including `StampScreenOrigin` for correct screen-space coordinates).

Use cases: screenshot testing, multi-process compositors, PDF/image export, CI.

### Incremental Rendering (ADR-004)

The rendering pipeline uses a three-level incremental strategy to minimize per-frame work:

```
Level 1 — Frame Skip:
  DrawTo() returns false when tree is clean → 0 CPU, 0 GPU upload

Level 2 — Dirty Region Redraw (persistent pixmap):
  dirty.Collector walks tree → dirty.Tracker collects regions → Optimize merges
  For each region: PushClip → DrawRect(background) → draw intersecting widgets → PopClip
  Pixmap persists between frames (Qt QBackingStore pattern)

Level 3 — RepaintBoundary Cache (subtree isolation):
  Container widgets auto-wrap children (ListView items)
  Cache hit → DrawImage blit (zero re-render)
  Cache miss → offscreen render → centralized ImageCache (LRU, 64MB)
```

Key components:
- `dirty.Tracker` + `dirty.Collector` — region tracking with merge optimization
- `RepaintBoundary` — pixel caching with scene.Renderer for large subtrees
- `ImageCache` — centralized LRU cache with memory budget and eviction
- `DrawStatsProvider` — observability (CachedWidgets, DirtyWidgets)
- `DirtyTrackerProvider` — O(regions) `Intersects()` fast path in RepaintBoundary

See `docs/dev/architecture/ADR-004-INCREMENTAL-RENDERING.md` for full design.

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

`Scheduler` batches widget re-render requests with push-based invalidation:

```go
sched := state.NewScheduler(func(dirty []widget.Widget) {
    // Re-render dirty widgets
})
sched.MarkDirty(widget)
sched.Flush() // Calls flush function with deduplicated widget list
```

Supports explicit batching via `Batch` method. Instance-based (no global state), thread-safe.

**Push-based invalidation:** `SetOnDirty(fn)` registers a callback fired when the pending set transitions from empty to non-empty. In `app/Window`, this is wired to `RequestRedraw()` — the render loop wakes up only when signals actually change.

**Reflush protection:** `Frame()` calls `Flush()` in a loop (max 2 re-flushes) to drain widgets that become dirty during flush callbacks. This prevents infinite loops from circular signal dependencies.

`Scheduler` satisfies the `widget.SchedulerRef` interface (`MarkDirty(Widget)`).

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
| `button` | `TextReadonlySignal` | `ReadonlySignal[string]` | one-way |
| `button` | `DisabledSignal` | `Signal[bool]` | one-way |
| `button` | `DisabledReadonlySignal` | `ReadonlySignal[bool]` | one-way |
| `checkbox` | `CheckedSignal` | `Signal[bool]` | two-way |
| `checkbox` | `LabelSignal` | `Signal[string]` | one-way |
| `checkbox` | `LabelReadonlySignal` | `ReadonlySignal[string]` | one-way |
| `checkbox` | `DisabledSignal` | `Signal[bool]` | one-way |
| `checkbox` | `DisabledReadonlySignal` | `ReadonlySignal[bool]` | one-way |
| `radio` | `SelectedSignal` | `Signal[string]` | two-way |
| `radio` | `GroupDisabledSignal` | `Signal[bool]` | one-way |
| `radio` | `GroupDisabledReadonlySignal` | `ReadonlySignal[bool]` | one-way |
| `textfield` | `ValueSignal` | `Signal[string]` | two-way |
| `dropdown` | `SelectedSignal` | `Signal[int]` | two-way |
| `primitives/text` | `ContentSignal` | `ReadonlySignal[string]` | one-way |

Priority resolution: ReadonlySignal > Signal > Fn > Static.

`ReadonlySignal` variants enable computed properties (via `state.NewComputed()`) to drive widget state. Two-way binding signals (CheckedSignal, SelectedSignal, ValueSignal) do not have readonly variants since the widget writes back to them.

### Signal Lifecycle (Hybrid Push-Pull)

All widgets with signal bindings implement `widget.Lifecycle`. On `Mount(ctx)`, each widget creates `BindToScheduler` subscriptions for its signals. On `Unmount()`, bindings are cleaned up automatically via `WidgetBase.CleanupBindings()`.

**Push path:** `Signal.Set()` -> subscriber callback -> `Scheduler.MarkDirty(widget)` -> `SetOnDirty` callback -> `RequestRedraw()` -> next frame starts

**Pull path:** Widget reads `signal.Get()` lazily during `Layout()` / `Draw()` — value is always current

This hybrid push-pull model (inspired by Angular Signals) eliminates the diamond problem, prevents glitch states, and reduces unnecessary frames to zero when no signals change.

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
- **FocusManager integration** — Tab/Shift+Tab navigation via `focus.Manager`
- Focus change handling
- Cursor sync to platform
- Frame statistics reporting via `FrameCallback`

The Window creates a `focus.Manager` and wires it to the widget tree root.
Key events flow through the FocusManager before reaching the widget tree,
enabling Tab navigation and keyboard shortcut dispatch.

### EventBridge

`app.EventBridge` translates `gpucontext` events into `event.*` types and dispatches them to the Window.

**Event pipeline:**
```
gpucontext (native OS events)
  -> EventBridge (OnPointer, OnTextInput, OnKeyboard)
    -> Window.HandleEvent()
      -> HoverTracker (hit-test ScreenBounds, synthesize Enter/Leave)
      -> FocusManager.HandleKeyEvent() (Tab/Shift+Tab, shortcuts)
      -> Root Widget tree (depth-first dispatch)
```

The EventBridge also handles:
- **ButtonState tracking** — synthesizes MouseUp events for buttons released between frames
- **OnTextInput** — character input handler for text fields (separate from KeyPress)
- **OnPointer** — W3C PointerEventSource for window Enter/Leave events

### HoverTracker and Cursor Management

The Window includes a `HoverTracker` that performs hit-testing on every MouseMove event using `ScreenBounds` (screen-space coordinates). It synthesizes `MouseEnter`/`MouseLeave` events for individual widgets, enabling hover cursors (pointer, text, resize) in production.

**Cursor lifecycle per frame:**
1. `Frame()` resets cursor to default (unless mouse buttons are held -- drag cursor protection)
2. Widget tree processes events, widgets call `ctx.SetCursor()` as needed
3. HoverTracker runs cursor sync immediately after `HandleEvent` for responsive feedback
4. After draw pass, cursor is synced to the platform provider

**Drag cursor protection:** When mouse buttons are held (drag in progress), `ResetCursor` is skipped so the drag cursor (e.g., resize for SplitView) is maintained throughout the drag operation.

### ScreenBounds (Coordinate System)

`WidgetBase.ScreenBounds()` returns the widget's bounds in screen-space coordinates. During the draw pass, `Canvas.TransformOffset()` + `widget.StampScreenOrigin()` stamp each widget's screen origin as transforms accumulate. This enables:

- **Overlay positioning** inside ScrollView (Dropdown, Popover use `ScreenBounds()` for correct placement)
- **Hit-testing** for hover tracking (mouse position in screen space matches widget screen bounds)
- Enterprise pattern equivalent to Flutter's `localToGlobal` / Qt's `mapToGlobal`

### Event Coordinate Transform (ScrollView)

ScrollView transforms mouse/wheel coordinates from screen space to content space before dispatching to children. This ensures hit-testing works correctly for widgets inside scrolled containers. ListView and DataTable rely on this transform rather than implementing their own.

---

## Primitives

### BoxWidget

Container that lays out children in a vertical (default) or horizontal stack:

```go
// Vertical (default)
card := primitives.Box(
    primitives.Text("Title").Bold(),
    primitives.Text("Body"),
).Padding(16).Background(widget.Hex(0xFFFFFF)).Rounded(8)

// Horizontal layout
row := primitives.HBox(label, input, button).Gap(8)

// Direction via signal binding
dir := state.NewSignal(primitives.DirectionHorizontal)
box := primitives.Box(children...).DirectionSignal(dir)
```

Supports: padding, background, border, rounded corners, shadow, gap, direction (HBox/VBox), explicit dimensions (width/height/min/max), PushClipRoundRect for child clipping.

Implements `widget.Widget`, `a11y.Accessible`, and `widget.Lifecycle` (for signal cleanup).

### TextWidget

Renders text with configurable font size, color, bold, alignment, and optional font family. The `FontFamily(name)` builder method routes to `StyledTextDrawer` for custom font rendering when available, with Inter fallback.

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
- `AssetLoader` handles asset management for plugins (fonts, icons, images)
- `MemoryAssetLoader.LoadFont()` auto-registers fonts in the global `FontRegistry` — loaded fonts are immediately available to all widgets via `StyledTextDrawer`

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
| `github.com/gogpu/gg` | 2D graphics + scene.Scene tile-parallel rendering | v0.48.11 |
| `github.com/gogpu/gpucontext` | Shared GPU interfaces (opaque struct tokens) | v0.21.0 |
| `github.com/gogpu/gogpu` | Application framework, windowing, Browser/WASM (examples only) | v0.42.0 |
| `github.com/coregx/signals` | Reactive state management | v0.1.0 |
| `golang.org/x/image` | Font rendering infrastructure | v0.41.0 |

**Indirect:** gogpu/wgpu v0.30.1, gogpu/naga v0.17.15, gogpu/gputypes v0.5.0, go-text/typesetting v0.3.4, golang.org/x/text v0.36.0

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

### 6. Opt-in Lifecycle for Signal Binding

Widgets that use reactive signals implement `Lifecycle` (opt-in via type assertion). This follows the Flutter `initState`/`dispose` pattern — explicit lifecycle hooks for resource management:

```go
if lc, ok := w.(widget.Lifecycle); ok {
    lc.Mount(ctx)   // Subscribe to signals
    // later...
    lc.Unmount()    // Unsubscribe (auto via CleanupBindings)
}
```

Widgets without signals are unaffected — no performance cost, no code changes.

### 7. Thread Safety via Mutexes

`WidgetBase` and `ContextImpl` use `sync.RWMutex` for state protection. Canvas and Renderer are NOT thread-safe and must be used from the UI thread.

### 8. Value Semantics for Geometry

All types in `geometry/` are small structs passed by value. Operations return new values without modifying the receiver. No heap allocations in hot paths.

---

*This document reflects the actual codebase as of June 15, 2026 (v0.1.29 — PointerCapturer ADR-031, Layer Tree compositor, damage-aware blit, 12 bug fixes, 4 design systems with 61 painters).*
