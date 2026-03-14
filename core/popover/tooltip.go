package popover

import (
	"time"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// defaultTooltipDelay is the default hover delay before showing a tooltip.
const defaultTooltipDelay = 500 * time.Millisecond

// Tooltip is a hover-triggered floating label anchored to a trigger widget.
// It appears after a configurable delay when the mouse enters the trigger,
// and disappears when the mouse leaves.
//
// Create with [NewTooltip] using functional options.
type Tooltip struct {
	widget.WidgetBase

	cfg     config
	painter Painter
	visible bool
	hovered bool

	// hoverStart tracks when the mouse entered the trigger for delay timing.
	hoverStart time.Time

	// overlayWidget is the tooltip overlay pushed to the stack.
	overlayWidget *tooltipOverlay
}

// NewTooltip creates a new Tooltip with the given options.
//
// The tooltip is initially hidden and appears on hover after the configured delay.
func NewTooltip(opts ...Option) *Tooltip {
	t := &Tooltip{
		painter: DefaultPainter{},
	}
	t.cfg = defaultConfig()
	t.SetVisible(true)
	t.SetEnabled(true)

	for _, opt := range opts {
		opt(&t.cfg)
	}

	if t.cfg.painter != nil {
		t.painter = t.cfg.painter
	}

	return t
}

// Layout calculates the tooltip's size. The tooltip itself takes the size
// of its trigger widget.
func (t *Tooltip) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	// Check if delay has elapsed and tooltip should appear.
	if t.hovered && !t.visible && !t.hoverStart.IsZero() {
		if ctx.Now().Sub(t.hoverStart) >= t.cfg.delay {
			t.show(ctx)
		}
	}

	if t.cfg.trigger != nil {
		return t.cfg.trigger.Layout(ctx, constraints)
	}
	return constraints.Constrain(geometry.Sz(0, 0))
}

// Draw renders the trigger widget. The tooltip content is drawn by the
// overlay stack.
func (t *Tooltip) Draw(ctx widget.Context, canvas widget.Canvas) {
	if canvas == nil {
		return
	}
	if t.cfg.trigger != nil {
		t.cfg.trigger.Draw(ctx, canvas)
	}
}

// Event handles input events, primarily mouse enter/leave on the trigger.
func (t *Tooltip) Event(ctx widget.Context, e event.Event) bool {
	if t.cfg.ResolvedDisabled() {
		return false
	}

	// Delegate to trigger first.
	if t.cfg.trigger != nil && t.cfg.trigger.Event(ctx, e) {
		return true
	}

	me, ok := e.(*event.MouseEvent)
	if !ok {
		return false
	}

	return t.handleMouseEvent(ctx, me)
}

// Children returns the trigger widget as the sole child.
func (t *Tooltip) Children() []widget.Widget {
	if t.cfg.trigger == nil {
		return nil
	}
	return []widget.Widget{t.cfg.trigger}
}

// IsOpen returns true if the tooltip is currently visible.
func (t *Tooltip) IsOpen() bool {
	return t.visible
}

// Text returns the tooltip text.
func (t *Tooltip) Text() string {
	return t.cfg.tooltipText
}

// handleMouseEvent processes mouse events for hover detection.
func (t *Tooltip) handleMouseEvent(ctx widget.Context, e *event.MouseEvent) bool {
	switch e.MouseType {
	case event.MouseEnter:
		t.hovered = true
		t.hoverStart = ctx.Now()
		// Request continuous frames for delay check.
		ctx.Invalidate()
		return false // Don't consume enter events.

	case event.MouseLeave:
		t.hovered = false
		t.hoverStart = time.Time{}
		if t.visible {
			t.hide(ctx)
		}
		return false // Don't consume leave events.

	case event.MouseMove:
		// If already showing, update position if needed.
		return false

	default:
		// Any click hides the tooltip.
		if t.visible && (e.MouseType == event.MousePress) {
			t.hide(ctx)
		}
		return false
	}
}

// show displays the tooltip overlay.
func (t *Tooltip) show(ctx widget.Context) {
	if t.visible || t.cfg.ResolvedDisabled() {
		return
	}

	om := ctx.OverlayManager()
	if om == nil {
		return
	}

	t.visible = true

	// Update signal if bound.
	if t.cfg.visibleSignal != nil {
		t.cfg.visibleSignal.Set(true)
	}

	// Calculate tooltip size from text.
	tooltipSize := t.calculateTooltipSize()

	// Create the overlay widget.
	t.overlayWidget = newTooltipOverlay(t.cfg.tooltipText, t.painter, t.cfg.placement, tooltipSize)

	// Position relative to trigger.
	triggerBounds := triggerBoundsOf(t.cfg.trigger)
	if triggerBounds.IsEmpty() {
		triggerBounds = t.Bounds()
	}
	windowSize := ctx.WindowSize()
	pos := CalculatePosition(t.cfg.placement, triggerBounds, tooltipSize, windowSize, t.cfg.gap)

	t.overlayWidget.SetBounds(geometry.FromPointSize(pos, tooltipSize))

	// Push to overlay stack. Tooltip is non-modal, dismiss on any outside event.
	om.PushOverlay(t.overlayWidget, func() {
		t.hide(ctx)
	})

	if t.cfg.onShow != nil {
		t.cfg.onShow()
	}

	ctx.Invalidate()
}

// hide removes the tooltip overlay.
func (t *Tooltip) hide(ctx widget.Context) {
	if !t.visible {
		return
	}

	t.visible = false

	// Update signal if bound.
	if t.cfg.visibleSignal != nil {
		t.cfg.visibleSignal.Set(false)
	}

	if t.overlayWidget != nil {
		om := ctx.OverlayManager()
		if om != nil {
			om.RemoveOverlay(t.overlayWidget)
		}
		t.overlayWidget = nil
	}

	if t.cfg.onHide != nil {
		t.cfg.onHide()
	}

	ctx.Invalidate()
}

// calculateTooltipSize estimates the tooltip size based on text content.
func (t *Tooltip) calculateTooltipSize() geometry.Size {
	// Estimate width: ~7px per character (approximate for default font size).
	charWidth := t.cfg.tooltipFontSize * tooltipCharWidthRatio
	textWidth := float32(len(t.cfg.tooltipText)) * charWidth
	if textWidth > t.cfg.maxWidth {
		textWidth = t.cfg.maxWidth
	}

	width := textWidth + t.cfg.tooltipPaddingH*2
	height := t.cfg.tooltipFontSize + t.cfg.tooltipPaddingV*2

	return geometry.Sz(width, height)
}

// Mount creates signal bindings for push-based invalidation.
func (t *Tooltip) Mount(ctx widget.Context) {
	sched := ctx.Scheduler()
	if sched == nil {
		return
	}
	if t.cfg.visibleSignal != nil {
		b := state.BindToScheduler(t.cfg.visibleSignal, t, sched)
		t.AddBinding(b)
	}
}

// Unmount is called when the tooltip is removed from the widget tree.
func (t *Tooltip) Unmount() {
	// Bindings are cleaned up automatically by WidgetBase.CleanupBindings().
}

// tooltipOverlay is the overlay widget that displays the tooltip text.
type tooltipOverlay struct {
	widget.WidgetBase

	text      string
	painter   Painter
	placement Placement
	size      geometry.Size
}

// newTooltipOverlay creates a new tooltip overlay widget.
func newTooltipOverlay(text string, painter Painter, placement Placement, size geometry.Size) *tooltipOverlay {
	to := &tooltipOverlay{
		text:      text,
		painter:   painter,
		placement: placement,
		size:      size,
	}
	to.SetVisible(true)
	to.SetEnabled(true)
	return to
}

// Layout returns the pre-calculated tooltip size.
func (to *tooltipOverlay) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	return constraints.Constrain(to.size)
}

// Draw renders the tooltip background and text.
func (to *tooltipOverlay) Draw(_ widget.Context, canvas widget.Canvas) {
	if canvas == nil {
		return
	}

	to.painter.PaintTooltip(canvas, &TooltipPaintState{
		Bounds:    to.Bounds(),
		Text:      to.text,
		Placement: to.placement,
	})
}

// Event returns false; tooltips do not consume events.
func (to *tooltipOverlay) Event(_ widget.Context, _ event.Event) bool {
	return false
}

// Children returns nil; tooltip overlay is a leaf widget.
func (to *tooltipOverlay) Children() []widget.Widget {
	return nil
}

// Compile-time interface checks.
var (
	_ widget.Widget    = (*Tooltip)(nil)
	_ widget.Lifecycle = (*Tooltip)(nil)
	_ widget.Widget    = (*tooltipOverlay)(nil)
)
