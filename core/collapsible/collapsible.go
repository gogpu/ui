package collapsible

import (
	"time"

	"github.com/gogpu/ui/animation"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// interactionState represents the current user interaction on the header.
type interactionState uint8

const (
	stateNormal  interactionState = iota
	stateHover                    // mouse is over the header
	statePressed                  // mouse button is held down on header
)

// Widget implements a collapsible section with a clickable header and
// expandable content area.
//
// Create with [New] using functional options:
//
//	section := collapsible.New(
//	    collapsible.Title("Details"),
//	    collapsible.Content(detailWidget),
//	    collapsible.Expanded(true),
//	)
type Widget struct {
	widget.WidgetBase
	cfg     config
	istate  interactionState
	painter Painter

	// Animation state.
	progress float32 // 0.0 = collapsed, 1.0 = expanded
	animCtrl *animation.Controller

	// Cached content size from last layout.
	contentSize geometry.Size

	// headerTitle is an internal TextWidget for the header title text.
	// It participates in Children() so dirty.Collector can track title
	// changes independently (e.g., TitleSignal updates → cyan overlay).
	headerTitle widget.Widget
}

// Default configuration values.
const (
	defaultHeaderHeight float32 = 36
	defaultAnimDuration         = 200 * time.Millisecond
)

// New creates a new collapsible section Widget with the given options.
//
// The returned widget is visible and enabled by default.
// The content starts collapsed unless [Expanded] is set to true.
func New(opts ...Option) *Widget {
	w := &Widget{
		painter:  DefaultPainter{},
		animCtrl: animation.NewController(),
	}
	w.SetVisible(true)
	w.SetEnabled(true)

	// Set defaults before applying options.
	w.cfg.headerHeight = defaultHeaderHeight
	w.cfg.animated = true
	w.cfg.duration = defaultAnimDuration

	for _, opt := range opts {
		opt(&w.cfg)
	}

	// Apply painter from config if set.
	if w.cfg.painter != nil {
		w.painter = w.cfg.painter
	}

	// Initialize progress based on expanded state.
	if w.cfg.ResolvedExpanded() {
		w.progress = 1.0
	}

	// Create internal header title widget for dirty tracking.
	w.headerTitle = newHeaderTextWidget()

	// ADR-028: parent chain for upward dirty propagation.
	// Flutter: RenderObject.adoptChild sets parent on each child.
	type parentSetter interface{ SetParent(widget.Widget) }
	if ps, ok := w.headerTitle.(parentSetter); ok {
		ps.SetParent(w)
	}
	if w.cfg.content != nil {
		if ps, ok := w.cfg.content.(parentSetter); ok {
			ps.SetParent(w)
		}
	}

	return w
}

// IsFocusable reports whether the collapsible section can receive focus.
func (w *Widget) IsFocusable() bool {
	return w.IsVisible() && w.IsEnabled()
}

// IsExpanded returns the current expanded state.
func (w *Widget) IsExpanded() bool {
	return w.cfg.ResolvedExpanded()
}

// SetExpanded sets the expanded state programmatically.
// If animated, triggers a smooth transition.
func (w *Widget) SetExpanded(expanded bool) {
	current := w.cfg.ResolvedExpanded()
	if current == expanded {
		return
	}
	w.setExpandedState(expanded)
}

// Toggle toggles the expanded state.
func (w *Widget) Toggle() {
	w.setExpandedState(!w.cfg.ResolvedExpanded())
}

// Progress returns the current animation progress (0.0 = collapsed, 1.0 = expanded).
// This is useful for testing and external monitoring.
func (w *Widget) Progress() float32 {
	return w.progress
}

// IsAnimating reports whether an expand/collapse animation is in progress.
func (w *Widget) IsAnimating() bool {
	return w.animCtrl != nil && w.animCtrl.HasActive()
}

// Layout calculates the widget's size given constraints.
//
// The height depends on the animation progress:
//   - Collapsed: headerHeight only
//   - Expanded: headerHeight + content height
//   - Animating: headerHeight + content height * progress
func (w *Widget) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	// Tick animation if active.
	w.tickAnimation(ctx)

	headerH := w.cfg.headerHeight

	// Layout content to determine its natural height.
	var contentH float32
	if w.cfg.content != nil {
		maxContentH := constraints.MaxHeight - headerH
		if maxContentH < 0 {
			maxContentH = 0
		}
		contentConstraints := geometry.BoxConstraints(
			constraints.MinWidth, constraints.MaxWidth,
			0, maxContentH,
		)
		w.contentSize = w.cfg.content.Layout(ctx, contentConstraints)
		contentH = w.contentSize.Height
	}

	// Total height = header + animated content portion.
	totalH := headerH + contentH*w.progress
	preferred := geometry.Sz(constraints.MaxWidth, totalH)
	return constraints.Constrain(preferred)
}

// Draw renders the collapsible section.
func (w *Widget) Draw(ctx widget.Context, canvas widget.Canvas) {
	bounds := w.Bounds()
	if bounds.IsEmpty() {
		return
	}

	headerBounds := geometry.NewRect(
		bounds.Min.X, bounds.Min.Y,
		bounds.Width(), w.cfg.headerHeight,
	)

	// Set bounds and stamp screen origin on header title widget for dirty tracking.
	if w.headerTitle != nil {
		if setter, ok := w.headerTitle.(interface{ SetBounds(geometry.Rect) }); ok {
			setter.SetBounds(headerBounds)
		}
		widget.StampScreenOrigin(w.headerTitle, canvas)
	}

	// Paint header via the painter.
	w.painter.PaintHeader(canvas, HeaderState{
		Title:         w.cfg.ResolvedTitle(),
		Expanded:      w.cfg.ResolvedExpanded(),
		Hovered:       w.istate == stateHover,
		Pressed:       w.istate == statePressed,
		Focused:       w.IsFocused(),
		Bounds:        headerBounds,
		ArrowProgress: w.progress,
		HeaderColor:   w.cfg.headerColor,
		ArrowColor:    w.cfg.arrowColor,
	})

	// Draw content if visible (progress > 0).
	if w.progress > 0 && w.cfg.content != nil {
		w.drawContent(ctx, canvas, bounds)
	} else if w.cfg.content != nil {
		// Collapsed (progress == 0): stamp an empty clip on content so child
		// boundaries receive a zero-size CompositorClip. This causes
		// isBoundaryLayerVisible to cull them, preventing stale textures
		// from being blitted at the old position (#122).
		emptyClip := geometry.Rect{}
		canvas.PushClip(emptyClip)
		widget.DrawChild(w.cfg.content, ctx, canvas)
		canvas.PopClip()
	}
}

// drawContent renders the content area with clipping.
func (w *Widget) drawContent(ctx widget.Context, canvas widget.Canvas, bounds geometry.Rect) {
	contentTop := bounds.Min.Y + w.cfg.headerHeight
	visibleH := w.contentSize.Height * w.progress

	// Clip content to the current animated height.
	clipRect := geometry.NewRect(
		bounds.Min.X, contentTop,
		bounds.Width(), visibleH,
	)
	canvas.PushClip(clipRect)

	// Position the content widget.
	contentBounds := geometry.NewRect(
		bounds.Min.X, contentTop,
		w.contentSize.Width, w.contentSize.Height,
	)
	if setter, ok := w.cfg.content.(interface{ SetBounds(geometry.Rect) }); ok {
		setter.SetBounds(contentBounds)
	}
	widget.StampScreenOrigin(w.cfg.content, canvas)
	w.cfg.content.Draw(ctx, canvas)

	canvas.PopClip()
}

// Event handles input events and returns true if consumed.
func (w *Widget) Event(ctx widget.Context, e event.Event) bool {
	// Let header handle its own events first.
	if handleEvent(w, ctx, e) {
		return true
	}

	// Forward to content widget when expanded.
	if w.IsExpanded() && w.cfg.content != nil {
		return w.cfg.content.Event(ctx, e)
	}

	return false
}

// Children returns the content widget for tree traversal.
// The content is always returned even when collapsed, to allow the framework
// to manage lifecycle and focus traversal.
func (w *Widget) Children() []widget.Widget {
	children := make([]widget.Widget, 0, 2)
	if w.headerTitle != nil {
		children = append(children, w.headerTitle)
	}
	if w.cfg.content != nil {
		children = append(children, w.cfg.content)
	}
	if len(children) == 0 {
		return nil
	}
	return children
}

// Mount creates signal bindings for push-based invalidation.
// Implements [widget.Lifecycle].
func (w *Widget) Mount(ctx widget.Context) {
	sched := ctx.Scheduler()
	if sched == nil {
		return
	}
	// Bind title signals to HEADER widget (not self) so dirty.Collector
	// reports header bounds, not full collapsible bounds.
	titleTarget := w.headerTitle
	if titleTarget == nil {
		titleTarget = w
	}
	if w.cfg.readonlyTitleSignal != nil {
		b := state.BindToScheduler(w.cfg.readonlyTitleSignal, titleTarget, sched)
		w.AddBinding(b)
	} else if w.cfg.titleSignal != nil {
		b := state.BindToScheduler(w.cfg.titleSignal, titleTarget, sched)
		w.AddBinding(b)
	}
	if w.cfg.readonlyExpandedSignal != nil {
		b := state.BindToSchedulerLayout(w.cfg.readonlyExpandedSignal, w, sched)
		w.AddBinding(b)
	} else if w.cfg.expandedSignal != nil {
		b := state.BindToSchedulerLayout(w.cfg.expandedSignal, w, sched)
		w.AddBinding(b)
	}
}

// Unmount is called when the widget is removed from the tree.
// Implements [widget.Lifecycle].
func (w *Widget) Unmount() {
	// Cancel any running animation.
	if w.animCtrl != nil {
		w.animCtrl.CancelAll()
	}
}

// setExpandedState updates the expanded state and starts animation if needed.
func (w *Widget) setExpandedState(expanded bool) {
	// Update state source.
	if w.cfg.expandedSignal != nil {
		w.cfg.expandedSignal.Set(expanded)
	} else {
		w.cfg.expanded = expanded
	}

	// Start animation or instant toggle.
	if w.cfg.animated {
		w.startAnimation(expanded)
	} else {
		if expanded {
			w.progress = 1.0
		} else {
			w.progress = 0.0
		}
	}

	w.SetNeedsRedraw(true)

	if w.cfg.onToggle != nil {
		w.cfg.onToggle(expanded)
	}
}

// startAnimation begins a tween animation for expand/collapse.
func (w *Widget) startAnimation(expanding bool) {
	target := float32(0.0)
	if expanding {
		target = 1.0
	}

	// Use a progress signal adapter so animation.To can drive it.
	adapter := &progressAdapter{w: w}
	animation.To(adapter, target).
		From(w.progress).
		Duration(w.cfg.duration).
		Ease(animation.EaseInOutCubic).
		Start(w.animCtrl)
}

// tickAnimation advances animation by delta time from context.
func (w *Widget) tickAnimation(ctx widget.Context) {
	if w.animCtrl == nil || !w.animCtrl.HasActive() {
		return
	}

	dt := ctx.DeltaTime()
	if dt < 1*time.Millisecond {
		dt = 1 * time.Millisecond
	}
	if dt > 32*time.Millisecond {
		dt = 32 * time.Millisecond
	}
	wasActive := w.animCtrl.HasActive()
	w.animCtrl.Tick(dt)

	if w.animCtrl.HasActive() {
		w.SetNeedsRedraw(true)
		ctx.Invalidate()
	} else if wasActive {
		w.SetNeedsRedraw(true)
		ctx.Invalidate()
	}
}

// progressAdapter implements the signalFloat32 interface to allow animation.To
// to drive the widget's progress field directly.
type progressAdapter struct {
	w *Widget
}

// Get returns the current progress.
func (a *progressAdapter) Get() float32 {
	return a.w.progress
}

// Set updates the progress and marks the widget for redraw.
// During collapse animation, the clip area shrinks each frame.
// InvalidateScene ensures the enclosing boundary's GPU texture is
// re-recorded so stale pixels outside the new clip are cleared.
func (a *progressAdapter) Set(v float32) {
	a.w.progress = v
	a.w.SetNeedsRedraw(true)
	a.w.InvalidateScene()
}

// Verify Widget implements required interfaces at compile time.
var (
	_ widget.Widget    = (*Widget)(nil)
	_ widget.Focusable = (*Widget)(nil)
	_ widget.Lifecycle = (*Widget)(nil)
)
