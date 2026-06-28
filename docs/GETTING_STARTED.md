# Getting Started with gogpu/ui

This guide walks you through building your first application with gogpu/ui,
from installation to custom widgets.

---

## Installation

```bash
go get github.com/gogpu/ui@latest
```

For windowed applications you also need the application framework and canvas:

```bash
go get github.com/gogpu/gogpu@latest
go get github.com/gogpu/gg@latest
```

**Requirements:** Go 1.25+, zero CGO (pure Go, works on all platforms).

---

## Minimal Application

Every gogpu/ui application has three parts:
1. A **gogpu.App** for windowing and GPU context
2. A **ui app.App** that manages the widget tree
3. A **draw callback** that bridges gg rendering to ui widgets

```go
package main

import (
    "fmt"
    "log"

    "github.com/gogpu/gg"
    _ "github.com/gogpu/gg/gpu"
    "github.com/gogpu/gg/integration/ggcanvas"
    "github.com/gogpu/gogpu"
    "github.com/gogpu/ui/app"
    "github.com/gogpu/ui/core/button"
    "github.com/gogpu/ui/primitives"
    "github.com/gogpu/ui/render"
    "github.com/gogpu/ui/theme/material3"
    "github.com/gogpu/ui/widget"
)

func main() {
    gogpuApp := gogpu.NewApp(gogpu.DefaultConfig().
        WithTitle("My First App").
        WithSize(600, 400))

    m3 := material3.New(widget.Hex(0x6750A4))

    uiApp := app.New(
        app.WithWindowProvider(gogpuApp),
        app.WithPlatformProvider(gogpuApp),
        app.WithEventSource(gogpuApp.EventSource()),
    )
    uiApp.SetRoot(
        primitives.Box(
            primitives.Text("Hello, gogpu/ui!").FontSize(24).Bold(),
            button.New(
                button.TextOpt("Click Me"),
                button.OnClick(func() { fmt.Println("Clicked!") }),
                button.PainterOpt(material3.ButtonPainter{Theme: m3}),
            ),
        ).Padding(24).Gap(12),
    )

    var canvas *ggcanvas.Canvas
    gogpuApp.OnDraw(func(dc *gogpu.Context) {
        w, h := dc.Width(), dc.Height()
        if w <= 0 || h <= 0 {
            return
        }
        if canvas == nil {
            provider := gogpuApp.GPUContextProvider()
            if provider == nil {
                return
            }
            var err error
            canvas, err = ggcanvas.New(provider, w, h)
            if err != nil {
                log.Printf("ggcanvas: %v", err)
                return
            }
        }
        uiApp.Frame()
        cw, ch := canvas.Size()
        if cw != w || ch != h {
            if err := canvas.Resize(w, h); err != nil {
                log.Printf("resize: %v", err)
            }
            cw, ch = w, h
        }
        sv := dc.SurfaceView()
        sw, sh := dc.SurfaceSize()
        canvas.Draw(func(cc *gg.Context) {
            cc.SetRGBA(0.94, 0.94, 0.94, 1)
            cc.DrawRectangle(0, 0, float64(cw), float64(ch))
            cc.Fill()
            widgetCanvas := render.NewCanvas(cc, cw, ch)
            uiApp.Window().DrawTo(widgetCanvas)
        })
        if err := canvas.RenderDirect(sv, sw, sh); err != nil {
            log.Printf("render: %v", err)
        }
    })
    gogpuApp.OnClose(func() { gg.CloseAccelerator() })

    if err := gogpuApp.Run(); err != nil {
        log.Fatal(err)
    }
}
```

---

## Layout with Box, HBox, VBox

`primitives.Box` is the primary layout container. Children stack vertically
by default. Use `HBox` for horizontal layout or `VBox` for explicit vertical.

```go
// Vertical stack (default)
column := primitives.Box(
    primitives.Text("Title").FontSize(20).Bold(),
    primitives.Text("Subtitle").FontSize(14),
).Padding(16).Gap(8)

// Horizontal row
row := primitives.HBox(
    primitives.Text("Left"),
    primitives.Text("Right"),
).Gap(12)

// Card with background, border, shadow
card := primitives.Box(
    primitives.Text("Card content"),
).
    Padding(24).
    Background(widget.RGBA8(255, 255, 255, 255)).
    Rounded(12).
    BorderStyle(1, widget.RGBA8(200, 200, 200, 255)).
    ShadowLevel(2)

// Fixed dimensions
header := primitives.Box(
    primitives.Text("Header"),
).Height(64).Background(widget.Hex(0x6750A4))
```

---

## Interactive Widgets

All interactive widgets use functional options for construction and support
pluggable painters for design-system independence.

### Button

```go
btn := button.New(
    button.TextOpt("Submit"),
    button.OnClick(func() { fmt.Println("submitted") }),
    button.VariantOpt(button.Filled),   // Filled, Outlined, TextOnly, Tonal
    button.SizeOpt(button.Medium),      // Small, Medium, Large
    button.PainterOpt(material3.ButtonPainter{Theme: m3}),
)
```

### Checkbox

```go
cb := checkbox.New(
    checkbox.LabelOpt("Accept terms"),
    checkbox.Checked(true),
    checkbox.OnToggle(func(checked bool) {
        fmt.Println("checked:", checked)
    }),
)
```

### Radio Group

```go
rg := radio.NewGroup(
    radio.Items(
        radio.ItemDef{Value: "sm", Label: "Small"},
        radio.ItemDef{Value: "md", Label: "Medium"},
        radio.ItemDef{Value: "lg", Label: "Large"},
    ),
    radio.Selected("md"),
    radio.OnChange(func(v string) { fmt.Println("size:", v) }),
)
```

### Dropdown

```go
dd := dropdown.New(
    dropdown.Items("Red", "Green", "Blue"),
    dropdown.Placeholder("Pick a color"),
    dropdown.OnChange(func(idx int, val string) {
        fmt.Println("color:", val)
    }),
)
```

### Slider

```go
s := slider.New(
    slider.Min(0),
    slider.Max(100),
    slider.Value(50),
    slider.OnChange(func(v float32) {
        fmt.Printf("value: %.0f\n", v)
    }),
)
```

### TextField

```go
tf := textfield.New(
    textfield.Placeholder("Enter your name"),
    textfield.OnChange(func(text string) {
        fmt.Println("input:", text)
    }),
)
```

---

## Theming

gogpu/ui ships with three design systems. Each provides painters for
every widget type.

### Material Design 3

```go
import "github.com/gogpu/ui/theme/material3"

m3 := material3.New(widget.Hex(0x6750A4)) // seed color

btn := button.New(
    button.TextOpt("M3 Button"),
    button.PainterOpt(material3.ButtonPainter{Theme: m3}),
)
```

### Fluent Design (Windows)

```go
import "github.com/gogpu/ui/theme/fluent"

fl := fluent.NewTheme() // default Windows accent blue
// fl := fluent.NewDarkTheme() // dark variant

btn := button.New(
    button.TextOpt("Fluent Button"),
    button.PainterOpt(fluent.ButtonPainter{Theme: fl}),
)
```

### Cupertino (Apple HIG)

```go
import "github.com/gogpu/ui/theme/cupertino"

cu := cupertino.NewTheme()
// cu := cupertino.NewDarkTheme()

btn := button.New(
    button.TextOpt("Cupertino Button"),
    button.PainterOpt(cupertino.ButtonPainter{Theme: cu}),
)
```

---

## Reactive State with Signals

The `state` package provides reactive primitives that automatically
invalidate widgets when values change.

### Signal (read-write)

```go
import "github.com/gogpu/ui/state"

count := state.NewSignal(0)

// Bind to a button label
btn := button.New(
    button.TextSignal(state.NewComputed(func() string {
        return fmt.Sprintf("Count: %d", count.Get())
    })),
    button.OnClick(func() {
        count.Set(count.Get() + 1)
    }),
)
```

### Two-Way Binding

Stateful widgets like checkbox, slider, and dropdown support two-way
signal bindings. The widget both reads from and writes to the signal:

```go
volume := state.NewSignal[float32](50)

s := slider.New(
    slider.Min(0),
    slider.Max(100),
    slider.ValueSignal(volume), // two-way: reads AND writes
)

// volume.Get() always reflects the slider's current value
```

### Computed Values

```go
firstName := state.NewSignal("John")
lastName := state.NewSignal("Doe")

fullName := state.NewComputed(func() string {
    return firstName.Get() + " " + lastName.Get()
})

label := primitives.Text("").ContentSignal(fullName)
```

---

## Custom Widgets

Implement the `widget.Widget` interface to create custom widgets. Embed
`widget.WidgetBase` for common functionality (bounds, visibility, enabled).

```go
type Counter struct {
    widget.WidgetBase
    count int
}

func NewCounter() *Counter {
    c := &Counter{}
    c.SetVisible(true)
    c.SetEnabled(true)
    return c
}

func (c *Counter) Layout(ctx widget.Context, cons geometry.Constraints) geometry.Size {
    return cons.Constrain(geometry.Sz(120, 40))
}

func (c *Counter) Draw(ctx widget.Context, canvas widget.Canvas) {
    b := c.Bounds()
    canvas.FillRect(b, widget.RGBA8(240, 240, 240, 255))
    canvas.DrawText(fmt.Sprintf("Count: %d", c.count), b.X+8, b.Y+28,
        widget.RGBA8(0, 0, 0, 255), 16)
}

func (c *Counter) Event(ctx widget.Context, e event.Event) bool {
    if me, ok := e.(*event.MouseEvent); ok && me.MouseType == event.MousePress {
        if c.Bounds().Contains(me.Position) {
            c.count++
            c.MarkDirty()
            return true
        }
    }
    return false
}

func (c *Counter) Children() []widget.Widget { return nil }
```

For production widgets, define a `Painter` interface to separate behavior
from rendering, allowing different design systems to style the widget:

```go
// In your widget package:
type CounterPainter interface {
    PaintCounter(canvas widget.Canvas, state *CounterPaintState)
}

// In theme/material3/:
type CounterPainter struct{ Theme *Theme }
func (p CounterPainter) PaintCounter(canvas widget.Canvas, state *CounterPaintState) {
    // Material 3 styled rendering
}
```

---

## Examples

The `examples/` directory contains working applications:

| Example | Description |
|---------|-------------|
| `examples/hello/` | Widget demo with checkboxes, radio buttons, ListView |
| `examples/signals/` | Reactive state management patterns |
| `examples/taskmanager/` | Full task manager with charts, tables, animations |
| `examples/gallery/` | Widget gallery with all 22 widgets, 3 design systems, theme switching |

Run any example:

```bash
cd examples/gallery
go run .
```

---

## Next Steps

- Browse the [API documentation](https://pkg.go.dev/github.com/gogpu/ui)
- Read `docs/ARCHITECTURE.md` for the system design
- Read `docs/EXTENSIONS.md` for creating plugins and custom themes
- Read `docs/VERSIONING.md` for compatibility guarantees
- Check `ROADMAP.md` for planned features
