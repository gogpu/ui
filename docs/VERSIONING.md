# Versioning Policy

> **gogpu/ui** follows semantic versioning with a conservative v0 strategy.

---

## Version Semantics

```
v0.x.y  → Development (breaking changes allowed)
v1.x.y  → Stable (backward compatible only)
v2.x.y  → Major breaking change (AVOID!)
```

### Why We Stay on v0.x.x

1. **API Freedom** — Can iterate without breaking user code expectations
2. **User Expectations** — v0 clearly signals "API may change"
3. **Avoiding v2** — v2 requires `/v2` import path (ecosystem disruption)

---

## Version Progression Plan

```
v0.1.0   Phase 1: MVP (core, layout, events)
v0.2.0   Phase 2: Beta (widgets, Material 3)
v0.3.0   Phase 3: RC (virtualization, animation)
v0.4.0+  Additional features, refinements
v0.9.0   API freeze (no new features)
v0.10.0+ Bug fixes, stabilization (12+ months)
v1.0.0   Production release
```

### Criteria for v1.0.0:

- [ ] API stable for 12+ months
- [ ] 80%+ test coverage
- [ ] Zero known critical bugs
- [ ] Complete documentation
- [ ] 20+ example applications
- [ ] Production use by 3+ projects

---

## Backward Compatibility Patterns

### 1. Functional Options

**Use for:** Configurable types, constructors

```go
// v0.1.0 — Initial API
func NewButton(text string, opts ...ButtonOption) *Button

// v0.5.0 — Add feature (NO breaking change)
func WithIcon(icon *Icon) ButtonOption
func WithVariant(v Variant) ButtonOption

// User code:
btn := NewButton("Save")                    // Still works!
btn := NewButton("Save", WithIcon(icon))    // New feature
```

### 2. Interface Extension

**Use for:** Optional widget capabilities

```go
// v0.1.0 — Core interface
type Widget interface {
    Layout(ctx *LayoutContext) Size
    Paint(ctx *PaintContext)
}

// v0.3.0 — Extended capability (NO breaking change)
type Focusable interface {
    Widget
    Focus()
    Blur()
}

// Usage via type assertion:
if f, ok := widget.(Focusable); ok {
    f.Focus()
}
```

### 3. Config Structs

**Use for:** Configuration with many options

```go
// v0.1.0 — Initial config
type AppConfig struct {
    Title  string
    Width  int
    Height int
}

// v0.3.0 — New fields (NO breaking change)
type AppConfig struct {
    Title  string
    Width  int
    Height int
    Theme  *Theme   // nil = default (zero value)
    DPI    float32  // 0 = auto (zero value)
}

// User code:
cfg := AppConfig{Title: "App", Width: 800, Height: 600}  // Still works!
```

### 4. Internal Packages

**Use for:** Implementation details

```go
// Public API — must be stable
import "github.com/gogpu/ui/widgets"

// Internal — can change freely (users cannot import)
import "github.com/gogpu/ui/internal/render"
```

### 5. Experimental Package

**Use for:** Unstable features

```go
// Explicitly unstable — users accept risk
import "github.com/gogpu/ui/experimental/docking"
```

---

## What to AVOID

### Breaking Function Signatures

```go
// ❌ WRONG: Breaking change
// v0.1.0
func NewButton(text string) *Button

// v0.2.0 — Breaks all existing code!
func NewButton(text string, icon *Icon) *Button

// ✅ CORRECT: Use options
func NewButton(text string, opts ...ButtonOption) *Button
```

### Breaking Interface Methods

```go
// ❌ WRONG: Breaking change
// v0.1.0
type Widget interface {
    Paint(ctx *PaintContext)
}

// v0.2.0 — Breaks all Widget implementations!
type Widget interface {
    Paint(ctx *PaintContext)
    PaintOverlay(ctx *PaintContext)  // NEW required method
}

// ✅ CORRECT: Separate optional interface
type OverlayPainter interface {
    PaintOverlay(ctx *PaintContext)
}
```

### Breaking Struct Fields

```go
// ❌ WRONG: Breaking change
// v0.1.0
type Button struct {
    Text string
}

// v0.2.0 — Breaks struct literals!
type Button struct {
    Label string  // Renamed from Text
}

// ✅ CORRECT: Keep field, add new one
type Button struct {
    Text  string  // Keep for compatibility
    Label string  // New field (alternative)
}
```

---

## Deprecation Process

### Step 1: Mark Deprecated (v0.x)

```go
// Deprecated: Use NewButtonWithOptions instead.
func NewButton(text string) *Button {
    return NewButtonWithOptions(text)
}
```

### Step 2: Document Migration

```go
// Migration guide in CHANGELOG.md:
// v0.5.0: NewButton deprecated, use NewButtonWithOptions
```

### Step 3: Remove in Major Version Only

```go
// v1.0.0: Deprecated items may be removed
// v2.0.0: AVOID v2 at all costs!
```

---

## Avoiding v2.0.0

### Why v2 is Problematic:

```go
// Before v2
import "github.com/gogpu/ui"

// After v2 — ALL IMPORTS MUST CHANGE
import "github.com/gogpu/ui/v2"
```

### Strategies to Avoid v2:

1. **Design API carefully from start**
2. **Use functional options everywhere**
3. **Use interface extension**
4. **Hide implementation in internal/**
5. **Use experimental/ for unstable features**
6. **Long v0.x development cycle**
7. **Extended v1.x stabilization period**

---

## Go Module Best Practices

### Go 1.25+ Features:

```go
// Generics — type-safe, no breaking changes needed
type Signal[T any] interface {
    Get() T
    Set(T)
}

// Range over functions (Go 1.23+)
func (l *List) Items() iter.Seq[Widget]
```

### Module File:

```go
// go.mod
module github.com/gogpu/ui

go 1.25

require (
    github.com/gogpu/gg v0.13.0
    github.com/gogpu/gogpu v0.8.0
    github.com/coregx/signals v0.1.0
)
```

---

## Summary

| Principle | Implementation |
|-----------|----------------|
| Stay on v0 | Long development cycle |
| Backward compatible | Functional options, interfaces |
| Hide implementation | internal/ packages |
| Unstable features | experimental/ package |
| Avoid v2 | Careful API design |

---

*This policy ensures long-term stability while allowing innovation.*
