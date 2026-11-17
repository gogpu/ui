package toggleswitch

import (
	"time"
	"unicode/utf8"

	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/gesture"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// Widget implements an on/off switch with an animated, draggable thumb.
//
// A switch is created with [New] using functional options:
//
//	sw := toggleswitch.New(
//	    toggleswitch.Label("Notifications"),
//	    toggleswitch.CheckedSignal(enabled),
//	)
type Widget struct {
	widget.WidgetBase
	cfg     config
	painter Painter

	hovered     bool
	pointerDown bool
	keyDown     bool

	// Thumb drag tracking, in the coordinate space of Bounds.
	dragging    bool
	pressX      float32
	pressPos    float32 // thumb position when the press started
	dragPos     float32
	lastPointer geometry.Point

	// Animated values, advanced from the frame clock on each draw.
	position  tween
	hoverAnim tween
	pressAnim tween
	animReady bool

	// Gesture recognizers for the unified pointer pipeline (ADR-049).
	clickRec *gesture.ClickRecognizer
	dragRec  *gesture.DragRecognizer
	teamRecs []gesture.Recognizer

	padding float32
}

// New creates a switch with the given options.
//
// The returned widget is visible, enabled, focusable, and off unless
// configured otherwise.
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
	if w.cfg.painter != nil {
		w.painter = w.cfg.painter
	}

	// The recognizers only claim the pointer in the gesture arena so that an
	// enclosing scrollable does not take over a thumb drag. All state changes
	// happen in the derived mouse events (event.go): gesture callbacks run
	// before those events and would race with them.
	w.clickRec = gesture.NewClickRecognizer(gesture.ClickConfig{MaxClickCount: 1})
	w.dragRec = gesture.NewDragRecognizer(gesture.DragConfig{Direction: gesture.DragDirectionHorizontal})
	team := &gesture.Team{Captain: w.dragRec}
	w.teamRecs = []gesture.Recognizer{
		team.Add(w.clickRec),
		team.Add(w.dragRec),
	}

	return w
}

// Layout constants.
const (
	defaultPadding float32 = 4
	charWidthRatio float32 = 0.55
	minHeight      float32 = 24
)

// IsChecked reports whether the switch is on.
func (w *Widget) IsChecked() bool {
	return w.cfg.ResolvedChecked()
}

// Padding sets the space around the switch graphic and label.
// Returns the widget for method chaining.
func (w *Widget) Padding(v float32) *Widget {
	w.padding = v
	return w
}

// IsFocusable reports whether the switch can currently receive focus.
func (w *Widget) IsFocusable() bool {
	return w.IsVisible() && !w.isDisabled()
}

// isDisabled reports whether the switch ignores input.
func (w *Widget) isDisabled() bool {
	return !w.IsEnabled() || w.cfg.ResolvedDisabled()
}

// isPressed reports whether the pointer or the Space key holds the switch.
func (w *Widget) isPressed() bool {
	return w.pointerDown || (w.keyDown && w.IsFocused())
}

// Layout calculates the preferred size within the given constraints.
func (w *Widget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	lm := resolveLayoutMetrics(w.painter)
	size := lm.SwitchSize()

	width := size.Width + w.padding*2
	height := max(size.Height+w.padding*2, minHeight)
	if label := w.cfg.ResolvedLabel(); label != "" {
		textWidth := float32(utf8.RuneCountInString(label)) * lm.SwitchFontSize() * charWidthRatio
		width += lm.SwitchLabelGap() + textWidth
	}

	return constraints.Constrain(geometry.Sz(width, height))
}

// Draw renders the switch and keeps requesting frames while it animates.
func (w *Widget) Draw(ctx widget.Context, canvas widget.Canvas) {
	now := ctx.Now()
	w.syncAnimations(now)
	w.painter.PaintSwitch(canvas, w.paintState(now))

	if !w.animating(now) {
		return
	}
	w.SetNeedsRedraw(true)
	if sched, ok := ctx.(widget.AnimationScheduler); ok {
		sched.ScheduleAnimationFrame()
	} else {
		ctx.InvalidateRect(w.Bounds())
	}
}

// syncAnimations retargets the animated values to the current state.
// Changes made by events or signals since the last draw start animating here.
func (w *Widget) syncAnimations(now time.Time) {
	disabled := w.isDisabled()
	position := unit(w.cfg.ResolvedChecked())
	if w.dragging {
		position = w.dragPos
	}
	hover := unit(w.hovered && !disabled)
	press := unit(w.isPressed() && !disabled)

	// The first draw shows the initial state as is. Without a frame clock
	// (zero time) there is nothing to animate against.
	if now.IsZero() || !w.animReady {
		w.position.jump(position)
		w.hoverAnim.jump(hover)
		w.pressAnim.jump(press)
		w.animReady = !now.IsZero()
		return
	}

	m := resolveMotion(w.painter)
	if w.dragging {
		w.position.jump(position)
	} else {
		w.position.animateTo(position, now, m.Toggle, m.ToggleEasing)
	}
	w.hoverAnim.animateTo(hover, now, m.Hover, nil)
	w.pressAnim.animateTo(press, now, m.Press, nil)
}

// animating reports whether any animated value is still moving at now.
func (w *Widget) animating(now time.Time) bool {
	return w.position.running(now) || w.hoverAnim.running(now) || w.pressAnim.running(now)
}

// paintState assembles the painter input for the frame at now.
func (w *Widget) paintState(now time.Time) PaintState {
	lm := resolveLayoutMetrics(w.painter)
	disabled := w.isDisabled()
	checked := w.cfg.ResolvedChecked()

	control := w.controlBounds(lm)
	label := w.cfg.ResolvedLabel()
	var labelBounds geometry.Rect
	if label != "" {
		labelBounds = w.labelBounds(lm, control)
	}

	turningOn := checked
	if w.position.running(now) {
		turningOn = w.position.rising()
	}

	return PaintState{
		Label:         label,
		Checked:       checked,
		Hovered:       w.hovered && !disabled,
		Pressed:       w.isPressed() && !disabled,
		Dragging:      w.dragging,
		Focused:       w.IsFocused(),
		Disabled:      disabled,
		Position:      w.position.value(now),
		Progress:      clampUnit(w.position.linear(now)),
		TurningOn:     turningOn,
		HoverAmount:   w.hoverAnim.value(now),
		PressAmount:   w.pressAnim.value(now),
		Bounds:        w.Bounds(),
		ControlBounds: control,
		LabelBounds:   labelBounds,
	}
}

// thumbPosition returns where the thumb is shown at now, in [0, 1].
func (w *Widget) thumbPosition(now time.Time) float32 {
	if w.animReady && w.position.running(now) {
		return clampUnit(w.position.value(now))
	}
	return unit(w.cfg.ResolvedChecked())
}

// controlBounds returns the switch graphic area: after the leading padding,
// centered vertically.
func (w *Widget) controlBounds(lm LayoutMetrics) geometry.Rect {
	b := w.Bounds()
	size := lm.SwitchSize()
	return geometry.NewRect(b.Min.X+w.padding, b.Min.Y+(b.Height()-size.Height)/2, size.Width, size.Height)
}

// labelBounds returns the label area right of the switch graphic.
func (w *Widget) labelBounds(lm LayoutMetrics, control geometry.Rect) geometry.Rect {
	b := w.Bounds()
	x := control.Max.X + lm.SwitchLabelGap()
	return geometry.NewRect(x, b.Min.Y, max(b.Max.X-w.padding-x, 0), b.Height())
}

// unit converts b to 1 (true) or 0 (false).
func unit(b bool) float32 {
	if b {
		return 1
	}
	return 0
}

// Event handles an input event and returns true if consumed.
func (w *Widget) Event(ctx widget.Context, e event.Event) bool {
	return handleEvent(w, ctx, e)
}

// Children returns nil because a switch is a leaf widget.
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
	if w.cfg.checkedSignal != nil {
		w.AddBinding(state.BindToScheduler(w.cfg.checkedSignal, w, sched))
	}
	if w.cfg.readonlyLabelSig != nil {
		w.AddBinding(state.BindToSchedulerLayout(w.cfg.readonlyLabelSig, w, sched))
	} else if w.cfg.labelSignal != nil {
		w.AddBinding(state.BindToSchedulerLayout(w.cfg.labelSignal, w, sched))
	}
	if w.cfg.readonlyDisabledSig != nil {
		w.AddBinding(state.BindToScheduler(w.cfg.readonlyDisabledSig, w, sched))
	} else if w.cfg.disabledSignal != nil {
		w.AddBinding(state.BindToScheduler(w.cfg.disabledSignal, w, sched))
	}
}

// Unmount releases gesture recognizers and drops interaction state, so a
// remounted switch starts at rest. Implements [widget.Lifecycle].
func (w *Widget) Unmount() {
	if w.clickRec != nil {
		w.clickRec.Dispose()
	}
	if w.dragRec != nil {
		w.dragRec.Dispose()
	}
	w.hovered = false
	w.pointerDown = false
	w.keyDown = false
	w.dragging = false
	w.animReady = false
	// Bindings are cleaned up automatically by WidgetBase.CleanupBindings().
}

// GestureHitTest returns the gesture recognizers for a pointer event at pos.
// Implements [gesture.GestureAware] for the unified pointer pipeline (ADR-049).
func (w *Widget) GestureHitTest(_ geometry.Point) []gesture.Recognizer {
	return w.teamRecs
}

// AccessibilityRole returns [a11y.RoleSwitch].
func (w *Widget) AccessibilityRole() a11y.Role {
	return a11y.RoleSwitch
}

// AccessibilityLabel returns the switch label.
func (w *Widget) AccessibilityLabel() string {
	return w.cfg.ResolvedLabel()
}

// AccessibilityHint returns the hint set with [A11yHint].
func (w *Widget) AccessibilityHint() string {
	return w.cfg.a11yHint
}

// AccessibilityValue returns an empty string; the value is reported
// through [a11y.State.Checked].
func (w *Widget) AccessibilityValue() string {
	return ""
}

// AccessibilityState reports the on/off value, focus, and availability.
func (w *Widget) AccessibilityState() a11y.State {
	checked := a11y.CheckedFalse
	if w.cfg.ResolvedChecked() {
		checked = a11y.CheckedTrue
	}
	return a11y.State{
		Disabled: w.isDisabled(),
		Checked:  checked,
		Focused:  w.IsFocused(),
		Hidden:   !w.IsVisible(),
	}
}

// AccessibilityActions returns the click action while the switch is enabled.
func (w *Widget) AccessibilityActions() []a11y.Action {
	if w.isDisabled() {
		return nil
	}
	return a11yActions
}

// a11yActions is shared by all switches; callers must not modify it.
var a11yActions = []a11y.Action{a11y.ActionClick}

// Verify Widget implements required interfaces at compile time.
var (
	_ widget.Widget        = (*Widget)(nil)
	_ widget.Focusable     = (*Widget)(nil)
	_ widget.Lifecycle     = (*Widget)(nil)
	_ gesture.GestureAware = (*Widget)(nil)
	_ a11y.Accessible      = (*Widget)(nil)
)
