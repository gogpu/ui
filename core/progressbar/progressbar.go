package progressbar

import (
	"fmt"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// Widget implements a linear progress bar showing a value between 0 and 1.
//
// A progress bar is created with [New] using functional options:
//
//	bar := progressbar.New(
//	    progressbar.Value(0.65),
//	    progressbar.ShowLabel(true),
//	    progressbar.Height(20),
//	)
//
// Fluent styling methods may be chained after construction:
//
//	bar.Padding(8)
type Widget struct {
	widget.WidgetBase
	cfg     config
	painter Painter

	// Styling overrides set via fluent methods.
	padding float32

	// lastDrawnValue caches the value from the most recent Draw call.
	// Signal bindings compare against this to suppress redundant redraws
	// when the signal fires with an unchanged value (#112).
	lastDrawnValue float64
	// lastDrawnValueSet tracks whether lastDrawnValue has been populated
	// by at least one Draw call. Zero is a valid progress value, so we
	// need a separate flag.
	lastDrawnValueSet bool
}

// New creates a new progress bar Widget with the given options.
//
// The returned widget is visible and enabled by default. It is not focusable
// because progress bars are display-only widgets.
func New(opts ...Option) *Widget {
	w := &Widget{
		padding: defaultPadding,
		painter: DefaultPainter{},
	}
	w.SetVisible(true)
	w.SetEnabled(true)

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
	defaultPadding    float32 = 4
	defaultBarHeight  float32 = 8
	defaultBarRadius  float32 = 4
	defaultFontSize   float32 = 11
	preferredBarWidth float32 = 200
	percentMultiplier float64 = 100
)

// SetValue updates the progress bar's static value.
// The value is clamped to [0, 1].
func (w *Widget) SetValue(v float64) {
	v = clampValue(v)
	if w.cfg.value != v {
		w.cfg.value = v
		w.SetNeedsRedraw(true)
	}
}

// Value returns the current resolved value (0 to 1).
func (w *Widget) Value() float64 {
	return clampValue(w.cfg.ResolvedValue())
}

// Layout calculates the progress bar's preferred size within the given constraints.
func (w *Widget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	barH := w.cfg.height
	if barH <= 0 {
		barH = defaultBarHeight
	}

	preferred := geometry.Sz(
		preferredBarWidth+w.padding*2,
		barH+w.padding*2,
	)

	return constraints.Constrain(preferred)
}

// Draw renders the progress bar to the canvas.
func (w *Widget) Draw(_ widget.Context, canvas widget.Canvas) {
	value := clampValue(w.cfg.ResolvedValue())
	w.lastDrawnValue = value
	w.lastDrawnValueSet = true

	barH := w.cfg.height
	if barH <= 0 {
		barH = defaultBarHeight
	}
	radius := w.cfg.radius
	if radius < 0 {
		radius = defaultBarRadius
	}
	// Use default radius when not explicitly set (zero value).
	if !w.cfg.radiusSet {
		radius = defaultBarRadius
	}

	label := w.resolveLabel(value)

	w.painter.PaintProgressBar(canvas, PaintState{
		Value:                  value,
		Bounds:                 w.Bounds(),
		BarHeight:              barH,
		Radius:                 radius,
		ShowLabel:              w.cfg.showLabel,
		Label:                  label,
		Disabled:               w.cfg.ResolvedDisabled(),
		ProgressBarColorScheme: w.cfg.colorScheme,
	})
}

// resolveLabel computes the label text for the given value.
func (w *Widget) resolveLabel(value float64) string {
	if !w.cfg.showLabel {
		return ""
	}
	if w.cfg.formatLabel != nil {
		return w.cfg.formatLabel(value)
	}
	return fmt.Sprintf("%.0f%%", value*percentMultiplier)
}

// Event handles an input event. Progress bars are display-only and always
// return false (events are never consumed).
func (w *Widget) Event(_ widget.Context, _ event.Event) bool {
	return false
}

// Children returns nil because a progress bar is a leaf widget.
func (w *Widget) Children() []widget.Widget {
	return nil
}

// Mount creates signal bindings for push-based invalidation.
// Implements [widget.Lifecycle].
//
// The value signal binding deduplicates notifications: if the signal fires
// with the same value that was last drawn, the widget is not marked dirty.
// This prevents flicker from redundant redraws (#112).
func (w *Widget) Mount(ctx widget.Context) {
	sched := ctx.Scheduler()
	if sched == nil {
		return
	}
	// Value signal: deduplicate by comparing against lastDrawnValue.
	if w.cfg.readonlyValueSignal != nil {
		b := w.bindValueSignal(w.cfg.readonlyValueSignal, sched)
		w.AddBinding(b)
	} else if w.cfg.valueSignal != nil {
		b := w.bindValueSignal(w.cfg.valueSignal, sched)
		w.AddBinding(b)
	}
	// Disabled signal: no deduplication needed (bool changes are infrequent).
	if w.cfg.readonlyDisabledSignal != nil {
		b := state.BindToScheduler(w.cfg.readonlyDisabledSignal, w, sched)
		w.AddBinding(b)
	} else if w.cfg.disabledSignal != nil {
		b := state.BindToScheduler(w.cfg.disabledSignal, w, sched)
		w.AddBinding(b)
	}
}

// bindValueSignal creates a deduplicating binding for a float64 signal.
// MarkDirty is only called when the clamped value differs from the last
// drawn value, preventing flicker from redundant redraws (#112).
func (w *Widget) bindValueSignal(sig state.ReadonlySignal[float64], sched widget.SchedulerRef) *state.Binding {
	return state.BindToSchedulerFunc(sig, func(newVal float64) bool {
		clamped := clampValue(newVal)
		if w.lastDrawnValueSet && clamped == w.lastDrawnValue {
			return false // suppress: value unchanged
		}
		return true
	}, w, sched)
}

// Unmount is called when the progress bar is removed from the widget tree.
// Implements [widget.Lifecycle].
func (w *Widget) Unmount() {
	// Bindings are cleaned up automatically by WidgetBase.CleanupBindings().
}

// Padding sets the padding around the progress bar.
// Returns the widget for method chaining.
func (w *Widget) Padding(v float32) *Widget {
	w.padding = v
	return w
}

// clampValue clamps v to the range [0, 1].
func clampValue(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// Verify Widget implements required interfaces at compile time.
var (
	_ widget.Widget    = (*Widget)(nil)
	_ widget.Lifecycle = (*Widget)(nil)
)
