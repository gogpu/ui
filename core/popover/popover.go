package popover

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// Popover is a click-triggered floating overlay that displays arbitrary widget
// content anchored to a trigger widget. It integrates with the overlay stack
// for z-ordering and supports click-outside dismissal.
//
// Create with [NewPopover] using functional options.
type Popover struct {
	widget.WidgetBase

	cfg     config
	painter Painter
	visible bool

	// overlayWidget is the content wrapper pushed to the overlay stack.
	overlayWidget *overlayContent
}

// NewPopover creates a new Popover with the given options.
//
// The returned widget is visible and enabled by default, but the popover
// content is initially hidden until opened via click or [Popover.Show].
func NewPopover(opts ...Option) *Popover {
	p := &Popover{
		painter: DefaultPainter{},
	}
	p.cfg = defaultConfig()
	p.SetVisible(true)
	p.SetEnabled(true)

	for _, opt := range opts {
		opt(&p.cfg)
	}

	if p.cfg.painter != nil {
		p.painter = p.cfg.painter
	}

	// Initialize from signal if provided.
	if p.cfg.visibleSignal != nil && p.cfg.visibleSignal.Get() {
		p.visible = true
	}

	return p
}

// IsFocusable reports whether the popover trigger can receive focus.
func (p *Popover) IsFocusable() bool {
	return p.IsVisible() && p.IsEnabled() && !p.cfg.ResolvedDisabled()
}

// Layout calculates the popover's size. The popover itself takes the size
// of its trigger widget.
func (p *Popover) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	if p.cfg.trigger != nil {
		return p.cfg.trigger.Layout(ctx, constraints)
	}
	return constraints.Constrain(geometry.Sz(0, 0))
}

// Draw renders the trigger widget. The popover content is drawn by the
// overlay stack, not as part of this widget's draw call.
func (p *Popover) Draw(ctx widget.Context, canvas widget.Canvas) {
	if canvas == nil {
		return
	}
	if p.cfg.trigger != nil {
		p.cfg.trigger.Draw(ctx, canvas)
	}
}

// Event handles input events for the popover trigger.
func (p *Popover) Event(ctx widget.Context, e event.Event) bool {
	if p.cfg.ResolvedDisabled() {
		return false
	}

	// Delegate to trigger first.
	if p.cfg.trigger != nil && p.cfg.trigger.Event(ctx, e) {
		return true
	}

	switch ev := e.(type) {
	case *event.MouseEvent:
		return p.handleMouseEvent(ctx, ev)
	case *event.KeyEvent:
		return p.handleKeyEvent(ctx, ev)
	default:
		return false
	}
}

// Children returns the trigger widget as the sole child.
// The popover content is rendered in the overlay, not as a child.
func (p *Popover) Children() []widget.Widget {
	if p.cfg.trigger == nil {
		return nil
	}
	return []widget.Widget{p.cfg.trigger}
}

// IsOpen returns true if the popover content is currently visible.
func (p *Popover) IsOpen() bool {
	return p.visible
}

// Show opens the popover content overlay.
func (p *Popover) Show(ctx widget.Context) {
	if p.visible || p.cfg.ResolvedDisabled() {
		return
	}

	om := ctx.OverlayManager()
	if om == nil {
		return
	}

	p.visible = true

	// Update signal if bound.
	if p.cfg.visibleSignal != nil {
		p.cfg.visibleSignal.Set(true)
	}

	// Create the overlay content wrapper.
	contentWidget := p.cfg.content
	if contentWidget == nil {
		return
	}

	p.overlayWidget = newOverlayContent(contentWidget, p.painter, p.cfg.placement)

	// Lay out the content to get its natural size.
	contentSize := p.resolveContentSize(ctx)

	// Position relative to trigger using screen-space bounds.
	// ScreenBounds accounts for all parent transforms (scroll offsets, etc.).
	anchor := triggerScreenBoundsOf(p.cfg.trigger)
	if anchor.IsEmpty() {
		anchor = p.ScreenBounds()
	}
	windowSize := ctx.WindowSize()
	pos := CalculatePosition(p.cfg.placement, anchor, contentSize, windowSize, p.cfg.gap)

	p.overlayWidget.SetBounds(geometry.FromPointSize(pos, contentSize))

	// Push to overlay stack.
	om.PushOverlay(p.overlayWidget, func() {
		p.hide(ctx)
	})

	if p.cfg.onShow != nil {
		p.cfg.onShow()
	}

	ctx.Invalidate()
}

// Hide closes the popover content overlay.
func (p *Popover) Hide(ctx widget.Context) {
	p.hide(ctx)
}

// hide is the internal close implementation.
func (p *Popover) hide(ctx widget.Context) {
	if !p.visible {
		return
	}

	p.visible = false

	// Update signal if bound.
	if p.cfg.visibleSignal != nil {
		p.cfg.visibleSignal.Set(false)
	}

	if p.overlayWidget != nil {
		om := ctx.OverlayManager()
		if om != nil {
			om.RemoveOverlay(p.overlayWidget)
		}
		p.overlayWidget = nil
	}

	if p.cfg.onHide != nil {
		p.cfg.onHide()
	}

	ctx.Invalidate()
}

// Toggle opens the popover if closed, closes it if open.
func (p *Popover) Toggle(ctx widget.Context) {
	if p.visible {
		p.hide(ctx)
	} else {
		p.Show(ctx)
	}
}

// resolveContentSize returns the content size, using fixed dimensions if set,
// or the content's natural layout size otherwise.
func (p *Popover) resolveContentSize(ctx widget.Context) geometry.Size {
	if p.cfg.contentWidth > 0 && p.cfg.contentHeight > 0 {
		return geometry.Sz(p.cfg.contentWidth, p.cfg.contentHeight)
	}

	windowSize := ctx.WindowSize()
	loose := geometry.Loose(windowSize)
	size := p.cfg.content.Layout(ctx, loose)

	if p.cfg.contentWidth > 0 {
		size.Width = p.cfg.contentWidth
	}
	if p.cfg.contentHeight > 0 {
		size.Height = p.cfg.contentHeight
	}

	return size
}

// handleMouseEvent processes mouse events on the trigger area.
func (p *Popover) handleMouseEvent(ctx widget.Context, e *event.MouseEvent) bool {
	if e.MouseType != event.MouseRelease || e.Button != event.ButtonLeft {
		return false
	}

	triggerBounds := triggerBoundsOf(p.cfg.trigger)
	if !triggerBounds.IsEmpty() && !triggerBounds.Contains(e.Position) {
		return false
	}

	p.Toggle(ctx)
	return true
}

// handleKeyEvent processes keyboard events when focused.
func (p *Popover) handleKeyEvent(ctx widget.Context, e *event.KeyEvent) bool {
	if !p.IsFocused() {
		return false
	}

	if e.KeyType != event.KeyPress {
		return false
	}

	switch e.Key {
	case event.KeyEnter, event.KeySpace:
		p.Toggle(ctx)
		return true
	case event.KeyEscape:
		if p.visible {
			p.hide(ctx)
			return true
		}
		return false
	default:
		return false
	}
}

// Mount creates signal bindings for push-based invalidation.
func (p *Popover) Mount(ctx widget.Context) {
	sched := ctx.Scheduler()
	if sched == nil {
		return
	}
	if p.cfg.visibleSignal != nil {
		b := state.BindToScheduler(p.cfg.visibleSignal, p, sched)
		p.AddBinding(b)
	}
}

// Unmount is called when the popover is removed from the widget tree.
func (p *Popover) Unmount() {
	// Bindings are cleaned up automatically by WidgetBase.CleanupBindings().
}

// overlayContent wraps the popover content widget for the overlay stack.
// It implements the overlay.Overlay interface indirectly via the
// OverlayManager's PushOverlay contract.
type overlayContent struct {
	widget.WidgetBase

	content   widget.Widget
	painter   Painter
	placement Placement
}

// newOverlayContent creates a new overlay content wrapper.
func newOverlayContent(content widget.Widget, painter Painter, placement Placement) *overlayContent {
	oc := &overlayContent{
		content:   content,
		painter:   painter,
		placement: placement,
	}
	oc.SetVisible(true)
	oc.SetEnabled(true)
	return oc
}

// Layout lays out the overlay content within its bounds.
func (oc *overlayContent) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	size := constraints.Constrain(oc.Bounds().Size())
	oc.SetBounds(geometry.FromPointSize(oc.Bounds().Min, size))

	if oc.content != nil {
		contentConstraints := geometry.Tight(size)
		oc.content.Layout(ctx, contentConstraints)
		if setter, ok := oc.content.(interface{ SetBounds(geometry.Rect) }); ok {
			setter.SetBounds(geometry.FromPointSize(oc.Bounds().Min, size))
		}
	}

	return size
}

// Draw renders the popover background and content.
func (oc *overlayContent) Draw(ctx widget.Context, canvas widget.Canvas) {
	if canvas == nil {
		return
	}

	oc.painter.PaintPopover(canvas, &PopoverPaintState{
		Bounds:    oc.Bounds(),
		Placement: oc.placement,
	})

	if oc.content != nil {
		oc.content.Draw(ctx, canvas)
	}
}

// Event dispatches events to the content widget.
func (oc *overlayContent) Event(ctx widget.Context, e event.Event) bool {
	if oc.content != nil {
		return oc.content.Event(ctx, e)
	}
	return false
}

// Children returns the content widget.
func (oc *overlayContent) Children() []widget.Widget {
	if oc.content == nil {
		return nil
	}
	return []widget.Widget{oc.content}
}

// SetBounds sets the bounds and propagates to the content.
func (oc *overlayContent) SetBounds(bounds geometry.Rect) {
	oc.WidgetBase.SetBounds(bounds)
	if oc.content != nil {
		if setter, ok := oc.content.(interface{ SetBounds(geometry.Rect) }); ok {
			setter.SetBounds(bounds)
		}
	}
}

// triggerBoundsOf returns the bounds of a trigger widget via type assertion.
// Returns an empty rect if the widget is nil or doesn't support Bounds().
func triggerBoundsOf(w widget.Widget) geometry.Rect {
	if w == nil {
		return geometry.Rect{}
	}
	if bg, ok := w.(interface{ Bounds() geometry.Rect }); ok {
		return bg.Bounds()
	}
	return geometry.Rect{}
}

// triggerScreenBoundsOf returns the screen-space bounds of a trigger widget.
// Screen bounds account for all parent transforms (scroll offsets, box positions).
// Falls back to local Bounds() if the widget doesn't support ScreenBounds.
func triggerScreenBoundsOf(w widget.Widget) geometry.Rect {
	if w == nil {
		return geometry.Rect{}
	}
	if sb, ok := w.(interface{ ScreenBounds() geometry.Rect }); ok {
		return sb.ScreenBounds()
	}
	return triggerBoundsOf(w)
}

// Compile-time interface checks.
var (
	_ widget.Widget    = (*Popover)(nil)
	_ widget.Focusable = (*Popover)(nil)
	_ widget.Lifecycle = (*Popover)(nil)
	_ widget.Widget    = (*overlayContent)(nil)
)
