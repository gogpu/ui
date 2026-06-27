package chip

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// interactionState represents the current user interaction state.
type interactionState uint8

const (
	stateNormal  interactionState = iota
	stateHover                    // pointer is over the chip
	statePressed                  // chip is held down
)

// Widget implements a compact, interactive chip.
//
// A chip is created with [New] using functional options:
//
//	c := chip.New(
//	    chip.Label("Tag"),
//	    chip.OnClick(handleTag),
//	)
//
// Fluent styling methods may be chained after construction:
//
//	c.Padding(2)
type Widget struct {
	widget.WidgetBase
	cfg     config
	state   interactionState
	painter Painter

	// Styling overrides set via fluent methods.
	padding float32
}

// New creates a new chip Widget with the given options.
//
// The returned widget is visible, enabled, and focusable by default.
func New(opts ...Option) *Widget {
	w := &Widget{
		painter: DefaultPainter{},
	}
	w.SetVisible(true)
	w.SetEnabled(true)

	for _, opt := range opts {
		opt(&w.cfg)
	}

	if w.cfg.painter != nil {
		w.painter = w.cfg.painter
	}

	return w
}

// Default sizing values.
const (
	defaultChipHeight float32 = 32
	defaultChipRadius float32 = 8
	defaultFontSize   float32 = 14
	labelPaddingX     float32 = 12
	// charWidthRatio approximates character width as a fraction of font size
	// for layout estimation.
	charWidthRatio float32 = 0.55
	minChipWidth   float32 = 32
)

// IsFocusable reports whether the chip can currently receive focus.
// Implements [widget.Focusable].
func (w *Widget) IsFocusable() bool {
	return w.IsVisible() && w.IsEnabled() && !w.cfg.ResolvedDisabled()
}

// Selected returns the chip's current resolved selected state.
func (w *Widget) Selected() bool {
	return w.cfg.ResolvedSelected()
}

// SetSelected updates the chip's selected state.
//
// If a writable [SelectedSignal] is bound, the new value is written back to it
// (two-way binding). With only a static value, the chip's own state is updated.
// When the selected state is driven by a read-only source ([SelectedFn] or
// [SelectedReadonlySignal]), this is a no-op: the owner controls that source.
func (w *Widget) SetSelected(sel bool) {
	if w.cfg.ResolvedSelected() == sel {
		return
	}
	w.applySelected(sel)
	w.SetNeedsRedraw(true)
}

// applySelected writes the new selected state to the appropriate writable
// source: the bound [SelectedSignal] if present, otherwise the static field.
// A [SelectedFn] or [SelectedReadonlySignal] source is read-only, so the chip
// leaves it untouched and relies on the owner to update it.
func (w *Widget) applySelected(sel bool) {
	if w.cfg.selectedSignal != nil {
		w.cfg.selectedSignal.Set(sel)
		return
	}
	if w.cfg.selectedFn == nil && w.cfg.readonlySelectedSignal == nil {
		w.cfg.selected = sel
	}
}

// Layout calculates the chip's preferred size within the given constraints.
func (w *Widget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	// Query LayoutMetrics from painter (type assert with default fallback).
	lm := resolveChipLayoutMetrics(w.painter)

	fontSize := lm.ChipFontSize()
	minW := lm.ChipMinWidth()
	padX := lm.ChipPadding()

	text := w.cfg.ResolvedLabel()
	textWidth := float32(len(text)) * fontSize * charWidthRatio

	contentW := textWidth + padX*2
	if contentW < minW {
		contentW = minW
	}

	preferred := geometry.Sz(
		contentW+w.padding*2,
		defaultChipHeight+w.padding*2,
	)
	return constraints.Constrain(preferred)
}

// Draw renders the chip to the canvas.
func (w *Widget) Draw(_ widget.Context, canvas widget.Canvas) {
	// Query LayoutMetrics from painter (type assert with default fallback).
	lm := resolveChipLayoutMetrics(w.painter)

	w.painter.PaintChip(canvas, PaintState{
		Label:       w.cfg.ResolvedLabel(),
		Bounds:      w.contentBounds(),
		Radius:      lm.ChipRadius(),
		FontSize:    lm.ChipFontSize(),
		Selectable:  w.cfg.selectable,
		Selected:    w.cfg.ResolvedSelected(),
		Hovered:     w.state == stateHover,
		Pressed:     w.state == statePressed,
		Focused:     w.IsFocused(),
		Disabled:    w.cfg.ResolvedDisabled(),
		ColorScheme: w.cfg.colorScheme,
	})
}

// contentBounds returns the chip bounds inset by the outer padding.
func (w *Widget) contentBounds() geometry.Rect {
	b := w.Bounds()
	if w.padding == 0 {
		return b
	}
	return geometry.NewRect(
		b.Min.X+w.padding,
		b.Min.Y+w.padding,
		b.Width()-w.padding*2,
		b.Height()-w.padding*2,
	)
}

// Event handles an input event and returns true if consumed.
func (w *Widget) Event(ctx widget.Context, e event.Event) bool {
	return handleEvent(w, ctx, e)
}

// Children returns nil because a chip is a leaf widget.
func (w *Widget) Children() []widget.Widget {
	return nil
}

// Mount creates signal bindings for push-based invalidation.
// Implements [widget.Lifecycle].
func (w *Widget) Mount(ctx widget.Context) {
	sched := ctx.Scheduler()
	if sched == nil {
		return
	}
	if w.cfg.readonlyLabelSignal != nil {
		w.AddBinding(state.BindToScheduler(w.cfg.readonlyLabelSignal, w, sched))
	} else if w.cfg.labelSignal != nil {
		w.AddBinding(state.BindToScheduler(w.cfg.labelSignal, w, sched))
	}
	if w.cfg.readonlySelectedSignal != nil {
		w.AddBinding(state.BindToScheduler(w.cfg.readonlySelectedSignal, w, sched))
	} else if w.cfg.selectedSignal != nil {
		w.AddBinding(state.BindToScheduler(w.cfg.selectedSignal, w, sched))
	}
	if w.cfg.readonlyDisabledSignal != nil {
		w.AddBinding(state.BindToScheduler(w.cfg.readonlyDisabledSignal, w, sched))
	} else if w.cfg.disabledSignal != nil {
		w.AddBinding(state.BindToScheduler(w.cfg.disabledSignal, w, sched))
	}
}

// Unmount is called when the chip is removed from the widget tree.
// Implements [widget.Lifecycle].
func (w *Widget) Unmount() {
	// Bindings are cleaned up automatically by WidgetBase.CleanupBindings().
}

// Padding sets the outer padding around the chip content.
// Returns the widget for method chaining.
func (w *Widget) Padding(v float32) *Widget {
	w.padding = v
	return w
}

// Verify Widget implements required interfaces at compile time.
var (
	_ widget.Widget    = (*Widget)(nil)
	_ widget.Focusable = (*Widget)(nil)
	_ widget.Lifecycle = (*Widget)(nil)
)
