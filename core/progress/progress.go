package progress

import (
	"fmt"
	"math"
	"time"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// Widget implements a circular progress indicator with determinate and
// indeterminate modes.
//
// A circular progress indicator is created with [New] using functional options:
//
//	indicator := progress.New(
//	    progress.Value(0.65),
//	    progress.Size(48),
//	    progress.ShowLabel(true),
//	)
type Widget struct {
	widget.WidgetBase
	cfg     config
	painter Painter

	// Animation state for indeterminate mode.
	startTime time.Time
}

// New creates a new circular progress indicator with the given options.
//
// The returned widget is visible and enabled by default. It is not focusable
// because progress indicators are display-only widgets.
func New(opts ...Option) *Widget {
	w := &Widget{
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

	// Indeterminate spinners animate every frame (SetNeedsRedraw in Draw).
	// Mark as RepaintBoundary so dirty propagation stops here — only the
	// spinner's 48×48 scene re-records, not the entire parent tree.
	// Flutter: CircularProgressIndicator is always at its own boundary.
	if w.cfg.indeterminate {
		w.SetRepaintBoundary(true)
	}

	return w
}

// Default dimensions and constants.
const (
	defaultDiameter    float32 = 48
	defaultStrokeWidth float32 = 4
	defaultFontSize    float32 = 11
	percentMultiplier  float64 = 100

	// M3 animation timing (Flutter reference: progress_indicator.dart).
	// One grow/shrink cycle takes 1333ms; rotation period is ~1568ms.
	arcCycleDuration     = 1.333       // seconds per arc expand+contract cycle
	rotationDuration     = 1.568       // seconds per full rotation
	indeterminateArcSpan = math.Pi / 2 // fallback for DefaultPainter
)

// SetValue updates the indicator's static value.
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

// IsIndeterminate returns true if the indicator is in indeterminate (spinner) mode.
func (w *Widget) IsIndeterminate() bool {
	return w.cfg.indeterminate
}

// Layout calculates the indicator's preferred size within the given constraints.
func (w *Widget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	diameter := w.cfg.diameter
	if diameter <= 0 {
		diameter = defaultDiameter
	}

	// Circular progress prefers diameter×diameter but must not expand
	// beyond that when parent gives wide constraints (VBox MinWidth=parent).
	// Tighten MaxWidth/MaxHeight to diameter so Constrain doesn't expand.
	// Flutter: CircularProgressIndicator wrapped in SizedBox(diameter).
	// Circular indicator is intrinsically sized: always diameter×diameter.
	// Ignore parent MinWidth/MinHeight (VBox gives MinWidth=parent width).
	// Respect parent MaxWidth/MaxHeight only if smaller than diameter (Tight).
	sw := diameter
	sh := diameter
	if constraints.MaxWidth < sw {
		sw = constraints.MaxWidth
	}
	if constraints.MaxHeight < sh {
		sh = constraints.MaxHeight
	}
	return geometry.Sz(sw, sh)
}

// Draw renders the circular progress indicator to the canvas.
func (w *Widget) Draw(ctx widget.Context, canvas widget.Canvas) {
	bounds := w.Bounds()
	if bounds.IsEmpty() {
		return
	}

	if w.cfg.indeterminate {
		w.drawIndeterminate(ctx, canvas, bounds)
	} else {
		w.drawDeterminate(ctx, canvas, bounds)
	}
}

// drawDeterminate renders the progress indicator directly via the painter.
func (w *Widget) drawDeterminate(_ widget.Context, canvas widget.Canvas, bounds geometry.Rect) {
	diameter := w.cfg.diameter
	if diameter <= 0 {
		diameter = defaultDiameter
	}
	strokeW := w.cfg.strokeWidth
	if strokeW <= 0 {
		strokeW = defaultStrokeWidth
	}

	value := clampValue(w.cfg.ResolvedValue())
	ps := PaintState{
		Bounds:      bounds,
		Diameter:    diameter,
		StrokeWidth: strokeW,
		Disabled:    w.cfg.ResolvedDisabled(),
		ColorScheme: w.cfg.colorScheme,
		Value:       value,
		ShowLabel:   w.cfg.showLabel,
	}
	if ps.ShowLabel {
		ps.Label = w.resolveLabel(value)
	}

	// Pre-compute geometry (ADR-034 Phase 4).
	ps.Center, ps.Radius = ComputeCenterRadius(ps)

	w.painter.PaintProgress(canvas, ps)
}

// drawIndeterminate renders the spinner via the painter.
func (w *Widget) drawIndeterminate(ctx widget.Context, canvas widget.Canvas, bounds geometry.Rect) {
	diameter := w.cfg.diameter
	if diameter <= 0 {
		diameter = defaultDiameter
	}
	strokeW := w.cfg.strokeWidth
	if strokeW <= 0 {
		strokeW = defaultStrokeWidth
	}

	elapsed := w.elapsedSeconds(ctx)
	rotation := computeRotation(elapsed)
	phase := computeAnimationPhase(elapsed)

	ps := PaintState{
		Bounds:         bounds,
		Diameter:       diameter,
		StrokeWidth:    strokeW,
		Disabled:       w.cfg.ResolvedDisabled(),
		ColorScheme:    w.cfg.colorScheme,
		Indeterminate:  true,
		Rotation:       rotation,
		AnimationPhase: phase,
	}

	// Pre-compute geometry and arc angles (ADR-034 Phase 4).
	ps.Center, ps.Radius = ComputeCenterRadius(ps)
	ps.ArcStartAngle, ps.ArcSweepAngle = ComputeArcAngles(phase, rotation)

	w.painter.PaintProgress(canvas, ps)
	w.SetNeedsRedraw(true)

	// Request next animation frame via deferred scheduling (Flutter scheduleFrame
	// pattern). Does NOT trigger immediate RequestRedraw — the animation pumper
	// controls actual frame rate. Falls back to immediate InvalidateRect if
	// AnimationScheduler not available (headless tests, legacy contexts).
	if sched, ok := ctx.(widget.AnimationScheduler); ok {
		sched.ScheduleAnimationFrame()
	} else {
		ctx.InvalidateRect(w.Bounds())
	}
}

// elapsedSeconds returns seconds since the spinner started.
func (w *Widget) elapsedSeconds(ctx widget.Context) float64 {
	now := ctx.Now()
	if w.startTime.IsZero() {
		w.startTime = now
	}
	return now.Sub(w.startTime).Seconds()
}

// computeRotation returns the continuous rotation angle in radians.
func computeRotation(elapsed float64) float64 {
	return (elapsed / rotationDuration) * 2 * math.Pi
}

// computeAnimationPhase returns a 0-1 sawtooth phase for the arc grow/shrink cycle.
func computeAnimationPhase(elapsed float64) float64 {
	return math.Mod(elapsed/arcCycleDuration, 1.0)
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

// Event handles an input event. Progress indicators are display-only and always
// return false (events are never consumed).
func (w *Widget) Event(_ widget.Context, _ event.Event) bool {
	return false
}

// Children returns nil because a progress indicator is a leaf widget.
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

// Unmount is called when the progress indicator is removed from the widget tree.
// Implements [widget.Lifecycle].
func (w *Widget) Unmount() {
	// Bindings are cleaned up automatically by WidgetBase.CleanupBindings().
}

// Option configures a circular progress indicator during construction.
type Option func(*config)

// Value sets the indicator's initial static value (0.0 to 1.0).
// Values outside [0, 1] are clamped during rendering.
func Value(v float64) Option {
	return func(c *config) {
		c.value = v
	}
}

// ValueFn sets a dynamic value function that is evaluated on each draw.
// When set, this takes precedence over the static value but not over
// a signal set via [ValueSignal] or [ValueReadonlySignal].
func ValueFn(fn func() float64) Option {
	return func(c *config) {
		c.valueFn = fn
	}
}

// ValueSignal binds the indicator's value to a reactive signal.
// This is a one-way read binding: the widget reads the value from the signal.
// When set, the signal value takes precedence over both [ValueFn] and [Value]
// but not over [ValueReadonlySignal].
func ValueSignal(sig state.Signal[float64]) Option {
	return func(c *config) {
		c.valueSignal = sig
	}
}

// ValueReadonlySignal binds the indicator's value to a read-only signal.
// This is useful for computed signals created via [state.NewComputed].
// When set, this takes highest precedence over all other value sources.
func ValueReadonlySignal(sig state.ReadonlySignal[float64]) Option {
	return func(c *config) {
		c.readonlyValueSignal = sig
	}
}

// Size sets the indicator's diameter in logical pixels. Default is 48.
func Size(diameter float32) Option {
	return func(c *config) {
		c.diameter = diameter
	}
}

// StrokeWidth sets the arc stroke width in logical pixels. Default is 4.
func StrokeWidth(w float32) Option {
	return func(c *config) {
		c.strokeWidth = w
	}
}

// TrackColor sets the background circle color.
func TrackColor(color widget.Color) Option {
	return func(c *config) {
		c.colorScheme.Track = color
		c.colorScheme.trackSet = true
	}
}

// IndicatorColor sets the progress arc color.
func IndicatorColor(color widget.Color) Option {
	return func(c *config) {
		c.colorScheme.Indicator = color
		c.colorScheme.indicatorSet = true
	}
}

// ShowLabel enables or disables the percentage label in the center.
// Only applies to determinate mode.
func ShowLabel(show bool) Option {
	return func(c *config) {
		c.showLabel = show
	}
}

// FormatLabelFn sets a custom label formatting function.
// The function receives the current value (0.0 to 1.0) and returns
// the label string. If nil, the default "65%" format is used.
func FormatLabelFn(fn func(float64) string) Option {
	return func(c *config) {
		c.formatLabel = fn
	}
}

// Indeterminate sets whether the indicator shows as a spinning arc.
// When true, the value is ignored and the arc rotates continuously.
func Indeterminate(v bool) Option {
	return func(c *config) {
		c.indeterminate = v
	}
}

// ColorSchemeOpt sets the full color scheme for painting.
// This overrides the painter's built-in defaults.
func ColorSchemeOpt(cs ProgressColorScheme) Option {
	return func(c *config) {
		c.colorScheme = cs
	}
}

// Disabled sets the indicator's disabled state.
func Disabled(d bool) Option {
	return func(c *config) {
		c.disabled = d
	}
}

// DisabledFn sets a dynamic function for the disabled state.
// When set, this takes precedence over the static value but not
// over a signal set via [DisabledSignal].
func DisabledFn(fn func() bool) Option {
	return func(c *config) {
		c.disabledFn = fn
	}
}

// DisabledSignal binds the disabled state to a reactive signal.
// When set, the signal value takes precedence over both [DisabledFn]
// and [Disabled] but not over [DisabledReadonlySignal].
func DisabledSignal(sig state.Signal[bool]) Option {
	return func(c *config) {
		c.disabledSignal = sig
	}
}

// DisabledReadonlySignal binds the disabled state to a read-only signal.
// When set, this takes highest precedence over all other disabled sources.
func DisabledReadonlySignal(sig state.ReadonlySignal[bool]) Option {
	return func(c *config) {
		c.readonlyDisabledSignal = sig
	}
}

// PainterOpt sets the painter used to render the progress indicator.
// Each design system provides its own painter. If not set,
// [DefaultPainter] is used.
func PainterOpt(p Painter) Option {
	return func(c *config) {
		c.painter = p
	}
}

// config holds the circular progress indicator's configuration.
type config struct {
	value                  float64
	valueFn                func() float64
	valueSignal            state.Signal[float64]
	readonlyValueSignal    state.ReadonlySignal[float64]
	diameter               float32
	strokeWidth            float32
	showLabel              bool
	formatLabel            func(float64) string
	indeterminate          bool
	disabled               bool
	disabledFn             func() bool
	disabledSignal         state.Signal[bool]
	readonlyDisabledSignal state.ReadonlySignal[bool]
	colorScheme            ProgressColorScheme
	painter                Painter
}

// ResolvedValue returns the current progress value.
// Priority: ReadonlySignal > Signal > Fn > Static.
func (c *config) ResolvedValue() float64 {
	if c.readonlyValueSignal != nil {
		return c.readonlyValueSignal.Get()
	}
	if c.valueSignal != nil {
		return c.valueSignal.Get()
	}
	if c.valueFn != nil {
		return c.valueFn()
	}
	return c.value
}

// ResolvedDisabled returns the current disabled state.
// Priority: ReadonlySignal > Signal > Fn > Static.
func (c *config) ResolvedDisabled() bool {
	if c.readonlyDisabledSignal != nil {
		return c.readonlyDisabledSignal.Get()
	}
	if c.disabledSignal != nil {
		return c.disabledSignal.Get()
	}
	if c.disabledFn != nil {
		return c.disabledFn()
	}
	return c.disabled
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

// ProgressColorScheme provides theme-derived colors for progress indicator painting.
// Zero value means the painter should use its built-in defaults.
type ProgressColorScheme struct {
	Indicator         widget.Color // progress arc color
	Track             widget.Color // background circle color
	Label             widget.Color // label text color
	DisabledIndicator widget.Color // arc color when disabled
	DisabledTrack     widget.Color // track color when disabled
	indicatorSet      bool         // true if Indicator was explicitly set
	trackSet          bool         // true if Track was explicitly set
}

// Verify Widget implements required interfaces at compile time.
var (
	_ widget.Widget    = (*Widget)(nil)
	_ widget.Lifecycle = (*Widget)(nil)
)
