package scrollview

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// Widget implements a scrollable container that clips and translates its
// content child widget.
//
// A scroll view is created with [New] using functional options:
//
//	sv := scrollview.New(content,
//	    scrollview.DirectionOpt(scrollview.Vertical),
//	    scrollview.OnScroll(handleScroll),
//	)
type Widget struct {
	widget.WidgetBase
	cfg     config
	content widget.Widget
	painter Painter

	// Cached layout measurements.
	contentSize  geometry.Size
	viewportSize geometry.Size

	// Interaction state.
	hovered         bool
	dragging        dragAxis
	dragStart       geometry.Point
	dragScrollStart float32

	// Track repeat state: page-scroll repeats while mouse is held on the track.
	trackRepeat trackRepeatState
}

// New creates a new scroll view Widget wrapping the given content widget.
//
// The returned widget is visible, enabled, and focusable by default.
// The default direction is [Vertical] with [ScrollbarAuto] visibility.
func New(content widget.Widget, opts ...Option) *Widget {
	w := &Widget{
		content: content,
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

// IsFocusable reports whether the scroll view can currently receive focus.
func (w *Widget) IsFocusable() bool {
	return w.IsVisible() && w.IsEnabled()
}

// Layout calculates the scroll view's size and measures its content.
//
// The viewport is constrained to the parent's constraints. The content
// is measured with unconstrained dimensions along the scroll axis to
// determine its natural size.
func (w *Widget) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	// The viewport fills the available space.
	w.viewportSize = constraints.Biggest()
	if w.viewportSize.Width <= 0 || w.viewportSize.Height <= 0 {
		w.viewportSize = constraints.Constrain(geometry.Sz(defaultViewportWidth, defaultViewportHeight))
	}

	if w.content == nil {
		return w.viewportSize
	}

	// Build content constraints: unconstrained along scroll axes.
	contentConstraints := w.buildContentConstraints()

	// Measure content.
	w.contentSize = w.content.Layout(ctx, contentConstraints)

	// Set content bounds at (0, 0) with its natural size.
	if setter, ok := w.content.(interface{ SetBounds(geometry.Rect) }); ok {
		setter.SetBounds(geometry.NewRect(0, 0, w.contentSize.Width, w.contentSize.Height))
	}

	return w.viewportSize
}

// buildContentConstraints creates constraints for measuring the content widget.
// Axes that scroll are unconstrained; non-scrolling axes are constrained to viewport.
func (w *Widget) buildContentConstraints() geometry.Constraints {
	switch w.cfg.direction {
	case Vertical:
		return geometry.Constraints{
			MinWidth:  w.viewportSize.Width,
			MaxWidth:  w.viewportSize.Width,
			MinHeight: 0,
			MaxHeight: geometry.Infinity,
		}
	case Horizontal:
		return geometry.Constraints{
			MinWidth:  0,
			MaxWidth:  geometry.Infinity,
			MinHeight: w.viewportSize.Height,
			MaxHeight: w.viewportSize.Height,
		}
	case Both:
		return geometry.Constraints{
			MinWidth:  0,
			MaxWidth:  geometry.Infinity,
			MinHeight: 0,
			MaxHeight: geometry.Infinity,
		}
	default:
		return geometry.Constraints{
			MinWidth:  w.viewportSize.Width,
			MaxWidth:  w.viewportSize.Width,
			MinHeight: 0,
			MaxHeight: geometry.Infinity,
		}
	}
}

// Default viewport dimensions used as fallback.
const (
	defaultViewportWidth  float32 = 200
	defaultViewportHeight float32 = 200
)

// Draw renders the scroll view to the canvas.
//
// Drawing order:
//  1. Push clip to viewport bounds
//  2. Push transform for scroll offset
//  3. Draw content
//  4. Pop transform
//  5. Pop clip
//  6. Draw scrollbar(s) on top
func (w *Widget) Draw(ctx widget.Context, canvas widget.Canvas) {
	bounds := w.Bounds()
	if bounds.IsEmpty() {
		return
	}

	// Process track repeat (page-scroll while mouse held on track).
	w.tickTrackRepeat(ctx)

	scrollX := w.cfg.ResolvedScrollX()
	scrollY := w.cfg.ResolvedScrollY()

	// Clip content to viewport.
	canvas.PushClip(bounds)
	canvas.PushTransform(geometry.Pt(bounds.Min.X-scrollX, bounds.Min.Y-scrollY))

	if w.content != nil {
		widget.StampScreenOrigin(w.content, canvas)
		w.content.Draw(ctx, canvas)
	}

	canvas.PopTransform()
	canvas.PopClip()

	// Draw scrollbar(s) on top.
	w.paintScrollbars(canvas)
}

// tickTrackRepeat fires the next page-scroll if the repeat timer has elapsed.
// Called from Draw so the self-invalidation loop drives continuous scrolling.
func (w *Widget) tickTrackRepeat(ctx widget.Context) {
	if !w.trackRepeat.active {
		return
	}

	now := ctx.Now()
	elapsed := now.Sub(w.trackRepeat.lastFire)

	// Initial press has a longer delay before first repeat.
	delay := trackRepeatInterval
	if w.trackRepeat.count == 0 {
		delay = trackRepeatInitialDelay
	}

	if elapsed < delay {
		// Not yet time — request another frame to check again.
		ctx.Invalidate()
		return
	}

	// Check if thumb has reached or passed the click position — stop.
	if w.trackRepeatReached() {
		w.trackRepeat.active = false
		return
	}

	if w.trackRepeat.axis == dragVertical {
		pageSize := w.viewportSize.Height
		setScroll(w, ctx, w.cfg.ResolvedScrollX(), w.cfg.ResolvedScrollY()+float32(w.trackRepeat.direction)*pageSize)
	} else {
		pageSize := w.viewportSize.Width
		setScroll(w, ctx, w.cfg.ResolvedScrollX()+float32(w.trackRepeat.direction)*pageSize, w.cfg.ResolvedScrollY())
	}

	w.trackRepeat.lastFire = now
	w.trackRepeat.count++

	// Request next frame for continuous repeat.
	ctx.Invalidate()
}

// trackRepeatReached returns true if the scrollbar thumb has reached
// or passed the original click position during track repeat scrolling.
func (w *Widget) trackRepeatReached() bool {
	vThumb, hThumb := w.computeThumbRects()
	if w.trackRepeat.axis == dragVertical && !vThumb.IsEmpty() {
		center := (vThumb.Min.Y + vThumb.Max.Y) / 2
		return (w.trackRepeat.direction > 0 && center >= w.trackRepeat.clickPos) ||
			(w.trackRepeat.direction < 0 && center <= w.trackRepeat.clickPos)
	}
	if w.trackRepeat.axis == dragHorizontal && !hThumb.IsEmpty() {
		center := (hThumb.Min.X + hThumb.Max.X) / 2
		return (w.trackRepeat.direction > 0 && center >= w.trackRepeat.clickPos) ||
			(w.trackRepeat.direction < 0 && center <= w.trackRepeat.clickPos)
	}
	return false
}

// paintScrollbars renders scrollbar overlays.
func (w *Widget) paintScrollbars(canvas widget.Canvas) {
	vThumb, hThumb := w.computeThumbRects()
	vTrack, hTrack := w.computeTrackRects()

	ps := PaintState{
		Bounds:    w.Bounds(),
		Direction: w.cfg.direction,
		Focused:   w.IsFocused(),
		Hovered:   w.hovered,
		Dragging:  w.dragging != dragNone,

		VScrollVisible: w.shouldShowVScrollbar(),
		VThumbRect:     vThumb,
		VTrackRect:     vTrack,

		HScrollVisible: w.shouldShowHScrollbar(),
		HThumbRect:     hThumb,
		HTrackRect:     hTrack,
	}

	w.painter.PaintScrollbar(canvas, ps)
}

// Event handles an input event and returns true if consumed.
//
// Event coordinates are transformed from parent space to content space before
// dispatching to the content widget. This is the inverse of the Draw transform
// (PushTransform with scroll offset), following the universal GUI framework
// pattern used by Flutter, Qt, WPF, and Gio.
//
// Scrollbar interactions are NOT transformed because scrollbars are drawn
// on top of the viewport in parent-space coordinates.
func (w *Widget) Event(ctx widget.Context, e event.Event) bool {
	// Handle scrollbar interactions FIRST so the thumb drag always works,
	// even when content would consume the same mouse event (e.g., item click).
	if me, ok := e.(*event.MouseEvent); ok {
		if w.dragging != dragNone || w.isOnScrollbar(me.Position) {
			return handleEvent(w, ctx, e)
		}
	}

	// Only dispatch to content if the event position is within the viewport.
	// This prevents content from receiving events when the mouse is outside.
	// Skip this check if bounds haven't been set yet (empty rect).
	if !w.isEventInsideViewport(e) {
		return handleEvent(w, ctx, e)
	}

	// Transform event coordinates to content space and dispatch to content.
	if w.content != nil {
		contentEvent := w.transformToContentSpace(e)
		if consumed := w.content.Event(ctx, contentEvent); consumed {
			return true
		}
	}

	return handleEvent(w, ctx, e)
}

// isEventInsideViewport reports whether a positional event is within the
// scroll view's viewport bounds. Returns true for non-positional events
// (key, focus) and when bounds are empty (not yet set).
func (w *Widget) isEventInsideViewport(e event.Event) bool {
	bounds := w.Bounds()
	if bounds.IsEmpty() {
		return true // bounds not set yet, allow dispatch
	}

	if me, ok := e.(*event.MouseEvent); ok {
		return bounds.Contains(me.Position)
	}
	if we, ok := e.(*event.WheelEvent); ok {
		return bounds.Contains(we.Position)
	}
	return true // non-positional events always pass through
}

// transformToContentSpace converts event coordinates from parent space to
// content space by applying the inverse of the Draw transform.
//
// Draw applies: PushTransform(Pt(bounds.Min.X - scrollX, bounds.Min.Y - scrollY))
// So screen-to-content is: contentPos = screenPos - bounds.Min + scrollOffset
//
// Non-positional events (key, focus) pass through unchanged.
func (w *Widget) transformToContentSpace(e event.Event) event.Event {
	bounds := w.Bounds()
	scrollX := w.cfg.ResolvedScrollX()
	scrollY := w.cfg.ResolvedScrollY()
	offset := geometry.Pt(scrollX-bounds.Min.X, scrollY-bounds.Min.Y)

	switch ev := e.(type) {
	case *event.MouseEvent:
		local := *ev
		local.Position = ev.Position.Add(offset)
		return &local
	case *event.WheelEvent:
		local := *ev
		local.Position = ev.Position.Add(offset)
		return &local
	default:
		return e
	}
}

// Children returns the content widget as the single child.
func (w *Widget) Children() []widget.Widget {
	if w.content == nil {
		return nil
	}
	return []widget.Widget{w.content}
}

// Mount creates signal bindings for push-based invalidation.
// Implements [widget.Lifecycle].
func (w *Widget) Mount(ctx widget.Context) {
	sched := ctx.Scheduler()
	if sched == nil {
		return
	}
	if w.cfg.readonlyScrollXSignal != nil {
		b := state.BindToScheduler(w.cfg.readonlyScrollXSignal, w, sched)
		w.AddBinding(b)
	} else if w.cfg.scrollXSignal != nil {
		b := state.BindToScheduler(w.cfg.scrollXSignal, w, sched)
		w.AddBinding(b)
	}
	if w.cfg.readonlyScrollYSignal != nil {
		b := state.BindToScheduler(w.cfg.readonlyScrollYSignal, w, sched)
		w.AddBinding(b)
	} else if w.cfg.scrollYSignal != nil {
		b := state.BindToScheduler(w.cfg.scrollYSignal, w, sched)
		w.AddBinding(b)
	}
}

// Unmount is called when the scroll view is removed from the widget tree.
// Implements [widget.Lifecycle].
func (w *Widget) Unmount() {
	// Bindings are cleaned up automatically by WidgetBase.CleanupBindings().
}

// Content returns the scroll view's content widget.
func (w *Widget) Content() widget.Widget {
	return w.content
}

// ScrollOffset returns the current scroll offset.
func (w *Widget) ScrollOffset() (x, y float32) {
	return w.cfg.ResolvedScrollX(), w.cfg.ResolvedScrollY()
}

// ViewportSize returns the current viewport size.
func (w *Widget) ViewportSize() geometry.Size {
	return w.viewportSize
}

// ContentSize returns the measured content size.
func (w *Widget) ContentSize() geometry.Size {
	return w.contentSize
}

// ScrollbarInset returns the width reserved for visible scrollbars.
// When the vertical scrollbar is shown, this is the total scrollbar width
// (track + padding). Otherwise zero.
func (w *Widget) ScrollbarInset() float32 {
	if w.shouldShowVScrollbar() {
		return scrollbarWidth + scrollbarPadding*2
	}
	return 0
}

// canScrollX reports whether horizontal scrolling is possible.
func (w *Widget) canScrollX() bool {
	if w.cfg.direction == Vertical {
		return false
	}
	return w.contentSize.Width > w.viewportSize.Width
}

// canScrollY reports whether vertical scrolling is possible.
func (w *Widget) canScrollY() bool {
	if w.cfg.direction == Horizontal {
		return false
	}
	return w.contentSize.Height > w.viewportSize.Height
}

// shouldShowVScrollbar returns true if the vertical scrollbar should be drawn.
func (w *Widget) shouldShowVScrollbar() bool {
	if w.cfg.direction == Horizontal {
		return false
	}
	switch w.cfg.scrollbar {
	case ScrollbarAlways:
		return true
	case ScrollbarNever:
		return false
	default: // ScrollbarAuto
		return w.contentSize.Height > w.viewportSize.Height
	}
}

// isOnScrollbar reports whether the given point is within the scrollbar track area.
func (w *Widget) isOnScrollbar(p geometry.Point) bool {
	vTrack, hTrack := w.computeTrackRects()
	if w.shouldShowVScrollbar() && !vTrack.IsEmpty() && vTrack.Contains(p) {
		return true
	}
	if w.shouldShowHScrollbar() && !hTrack.IsEmpty() && hTrack.Contains(p) {
		return true
	}
	return false
}

// shouldShowHScrollbar returns true if the horizontal scrollbar should be drawn.
func (w *Widget) shouldShowHScrollbar() bool {
	if w.cfg.direction == Vertical {
		return false
	}
	switch w.cfg.scrollbar {
	case ScrollbarAlways:
		return true
	case ScrollbarNever:
		return false
	default: // ScrollbarAuto
		return w.contentSize.Width > w.viewportSize.Width
	}
}

// computeTrackRects calculates the track rectangles for both scrollbars.
func (w *Widget) computeTrackRects() (vTrack, hTrack geometry.Rect) {
	bounds := w.Bounds()
	showV := w.shouldShowVScrollbar()
	showH := w.shouldShowHScrollbar()

	if showV {
		vTrack = computeScrollbarRect(bounds, dragVertical, showH)
	}
	if showH {
		hTrack = computeScrollbarRect(bounds, dragHorizontal, showV)
	}
	return vTrack, hTrack
}

// computeThumbRects calculates the thumb rectangles for both scrollbars.
func (w *Widget) computeThumbRects() (vThumb, hThumb geometry.Rect) {
	vTrack, hTrack := w.computeTrackRects()

	if w.shouldShowVScrollbar() && !vTrack.IsEmpty() {
		vThumb = w.computeVThumbRect(vTrack)
	}
	if w.shouldShowHScrollbar() && !hTrack.IsEmpty() {
		hThumb = w.computeHThumbRect(hTrack)
	}
	return vThumb, hThumb
}

// computeVThumbRect calculates the vertical thumb rectangle within the track.
func (w *Widget) computeVThumbRect(track geometry.Rect) geometry.Rect {
	trackLen := track.Height() - scrollbarPadding*2
	if trackLen <= 0 {
		return geometry.Rect{}
	}

	thumbSize := computeThumbSize(w.viewportSize.Height, w.contentSize.Height, trackLen)
	maxScroll := w.contentSize.Height - w.viewportSize.Height
	thumbPos := computeThumbPosition(w.cfg.ResolvedScrollY(), maxScroll, trackLen, thumbSize)

	return geometry.NewRect(
		track.Min.X+scrollbarPadding,
		track.Min.Y+scrollbarPadding+thumbPos,
		scrollbarWidth,
		thumbSize,
	)
}

// computeHThumbRect calculates the horizontal thumb rectangle within the track.
func (w *Widget) computeHThumbRect(track geometry.Rect) geometry.Rect {
	trackLen := track.Width() - scrollbarPadding*2
	if trackLen <= 0 {
		return geometry.Rect{}
	}

	thumbSize := computeThumbSize(w.viewportSize.Width, w.contentSize.Width, trackLen)
	maxScroll := w.contentSize.Width - w.viewportSize.Width
	thumbPos := computeThumbPosition(w.cfg.ResolvedScrollX(), maxScroll, trackLen, thumbSize)

	return geometry.NewRect(
		track.Min.X+scrollbarPadding+thumbPos,
		track.Min.Y+scrollbarPadding,
		thumbSize,
		scrollbarWidth,
	)
}

// Padding sets the content padding. Returns the widget for method chaining.
func (w *Widget) Padding(_ float32) *Widget {
	// Reserved for future use. Currently a no-op.
	return w
}

// Verify Widget implements required interfaces at compile time.
var (
	_ widget.Widget    = (*Widget)(nil)
	_ widget.Focusable = (*Widget)(nil)
	_ widget.Lifecycle = (*Widget)(nil)
)
