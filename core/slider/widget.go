package slider

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// interactionState represents the current user interaction state.
type interactionState uint8

const (
	stateNormal   interactionState = iota
	stateHover                     // mouse is over the slider
	stateDragging                  // thumb is being dragged
)

// Widget implements a draggable slider for selecting a value from a range.
//
// A slider is created with [New] using functional options:
//
//	s := slider.New(
//	    slider.Min(0),
//	    slider.Max(100),
//	    slider.Value(50),
//	    slider.OnChange(handleChange),
//	)
//
// Fluent styling methods may be chained after construction:
//
//	s.Padding(8)
type Widget struct {
	widget.WidgetBase
	cfg         config
	interaction interactionState
	painter     Painter

	// Styling overrides set via fluent methods.
	padding float32
}

// New creates a new slider Widget with the given options.
//
// The returned widget is visible, enabled, and focusable by default.
// The default range is [0, 100] with Horizontal orientation.
func New(opts ...Option) *Widget {
	w := &Widget{
		padding: defaultPadding,
		painter: DefaultPainter{},
	}
	w.SetVisible(true)
	w.SetEnabled(true)

	// Default range.
	w.cfg.maxVal = defaultMaxVal

	for _, opt := range opts {
		opt(&w.cfg)
	}

	// Apply painter from config if set.
	if w.cfg.painter != nil {
		w.painter = w.cfg.painter
	}

	return w
}

// Default values.
const (
	defaultPadding float32 = 4
	defaultMaxVal  float32 = 100
)

// IsFocusable reports whether the slider can currently receive focus.
// A slider is focusable when it is visible, enabled, and not disabled.
func (w *Widget) IsFocusable() bool {
	return w.IsVisible() && w.IsEnabled() && !w.cfg.ResolvedDisabled()
}

// Layout calculates the slider's preferred size within the given constraints.
func (w *Widget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	// Query LayoutMetrics from painter (type assert with default fallback).
	lm := resolveSliderLayoutMetrics(w.painter)
	tr := lm.SliderThumbRadius()

	if w.cfg.orientation == Vertical {
		height := constraints.MaxHeight
		if height <= 0 || height == geometry.Infinity {
			height = verticalDefaultHeight + w.padding*2
		}
		return constraints.Constrain(geometry.Sz(
			tr*2+w.padding*2, height))
	}
	width := constraints.MaxWidth
	if width <= 0 || width == geometry.Infinity {
		width = horizontalDefaultWidth + w.padding*2
	}
	return constraints.Constrain(geometry.Sz(
		width, tr*2+w.padding*2))
}

// Layout dimension constants.
const (
	horizontalDefaultWidth float32 = 200
	verticalDefaultHeight  float32 = 200
)

// Draw renders the slider to the canvas.
func (w *Widget) Draw(ctx widget.Context, canvas widget.Canvas) {
	rangeVal := w.cfg.maxVal - w.cfg.minVal
	var progress float32
	if rangeVal > 0 {
		progress = (w.cfg.ResolvedValue() - w.cfg.minVal) / rangeVal
		if progress < 0 {
			progress = 0
		}
		if progress > 1 {
			progress = 1
		}
	}

	w.painter.PaintSlider(canvas, PaintState{
		Value:       w.cfg.ResolvedValue(),
		Min:         w.cfg.minVal,
		Max:         w.cfg.maxVal,
		Progress:    progress,
		Hovered:     w.interaction == stateHover,
		Dragging:    w.interaction == stateDragging,
		Focused:     w.IsFocused(),
		Disabled:    w.cfg.ResolvedDisabled(),
		Bounds:      w.Bounds(),
		Orientation: w.cfg.orientation,
		Marks:       w.cfg.marks,
	})
}

// Event handles an input event and returns true if consumed.
func (w *Widget) Event(ctx widget.Context, e event.Event) bool {
	return handleEvent(w, ctx, e)
}

// Children returns nil because a slider is a leaf widget.
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
	if w.cfg.readonlyValueSignal != nil {
		b := state.BindToScheduler(w.cfg.readonlyValueSignal, w, sched)
		w.AddBinding(b)
	} else if w.cfg.valueSignal != nil {
		b := state.BindToScheduler(w.cfg.valueSignal, w, sched)
		w.AddBinding(b)
	}
	if w.cfg.readonlyDisabledSignal != nil {
		b := state.BindToScheduler(w.cfg.readonlyDisabledSignal, w, sched)
		w.AddBinding(b)
	} else if w.cfg.disabledSignal != nil {
		b := state.BindToScheduler(w.cfg.disabledSignal, w, sched)
		w.AddBinding(b)
	}
}

// Unmount is called when the slider is removed from the widget tree.
// Implements [widget.Lifecycle].
func (w *Widget) Unmount() {
	// Bindings are cleaned up automatically by WidgetBase.CleanupBindings().
}

// Verify Widget implements required interfaces at compile time.
var (
	_ widget.Widget    = (*Widget)(nil)
	_ widget.Focusable = (*Widget)(nil)
	_ widget.Lifecycle = (*Widget)(nil)
)
