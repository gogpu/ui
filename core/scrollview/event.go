package scrollview

import (
	"time"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// trackRepeatState holds state for repeated page-scroll while the mouse
// button is held down on the scrollbar track.
type trackRepeatState struct {
	active    bool
	axis      dragAxis  // which axis is being scrolled
	direction int       // -1 for up/left, +1 for down/right
	clickPos  float32   // click position on the scroll axis
	lastFire  time.Time // time of last scroll action
	count     int       // number of repeats fired (0 = initial click)
}

// Track repeat timing (Qt6 QScrollBar pattern: 500ms initial, 50ms repeat).
const (
	trackRepeatInitialDelay = 500 * time.Millisecond
	trackRepeatInterval     = 50 * time.Millisecond
)

// handleEvent processes input events for the scroll view widget.
func handleEvent(w *Widget, ctx widget.Context, e event.Event) bool {
	switch ev := e.(type) {
	case *event.WheelEvent:
		return handleWheelEvent(w, ctx, ev)
	case *event.MouseEvent:
		return handleMouseEvent(w, ctx, ev)
	case *event.KeyEvent:
		return handleKeyEvent(w, ctx, ev)
	default:
		return false
	}
}

// handleWheelEvent processes mouse wheel events for scrolling.
func handleWheelEvent(w *Widget, ctx widget.Context, e *event.WheelEvent) bool {
	if !w.Bounds().Contains(e.Position) {
		return false
	}

	step := w.cfg.resolvedScrollStep()
	dx := e.Delta.X * step
	dy := e.Delta.Y * step

	// Restrict scroll axes based on direction.
	switch w.cfg.direction {
	case Vertical:
		dx = 0
	case Horizontal:
		dy = 0
	case Both:
		// Allow both
	}

	if dx == 0 && dy == 0 {
		return false
	}

	scrollX := w.cfg.ResolvedScrollX() + dx
	scrollY := w.cfg.ResolvedScrollY() + dy

	setScroll(w, ctx, scrollX, scrollY)
	return true
}

// handleMouseEvent processes mouse events for scrollbar thumb dragging.
func handleMouseEvent(w *Widget, ctx widget.Context, e *event.MouseEvent) bool {
	switch e.MouseType {
	case event.MouseEnter:
		if !w.hovered {
			w.hovered = true
			w.MarkRedrawLocal()
			ctx.InvalidateRect(w.Bounds())
		}
		return false // Don't consume enter events -- let children handle them too.

	case event.MouseLeave:
		if w.hovered {
			w.hovered = false
			if w.dragging == dragNone {
				w.MarkRedrawLocal()
				ctx.InvalidateRect(w.Bounds())
			}
		}
		return false

	case event.MousePress:
		return handleMousePress(w, ctx, e)

	case event.MouseRelease:
		return handleMouseRelease(w, ctx, e)

	case event.MouseMove:
		return handleMouseMove(w, ctx, e)

	default:
		return false
	}
}

// handleMousePress handles mouse button press on scrollbar thumbs.
func handleMousePress(w *Widget, ctx widget.Context, e *event.MouseEvent) bool {
	if e.Button != event.ButtonLeft {
		return false
	}

	// Check if press is on a scrollbar thumb.
	vThumb, hThumb := w.computeThumbRects()

	if w.canScrollY() && !vThumb.IsEmpty() && vThumb.Contains(e.Position) {
		w.dragging = dragVertical
		w.dragStart = e.Position
		w.dragScrollStart = w.cfg.ResolvedScrollY()
		ctx.RequestFocus(w)
		w.MarkRedrawLocal()
		ctx.InvalidateRect(w.Bounds())
		return true
	}

	if w.canScrollX() && !hThumb.IsEmpty() && hThumb.Contains(e.Position) {
		w.dragging = dragHorizontal
		w.dragStart = e.Position
		w.dragScrollStart = w.cfg.ResolvedScrollX()
		ctx.RequestFocus(w)
		w.MarkRedrawLocal()
		ctx.InvalidateRect(w.Bounds())
		return true
	}

	// Check if press is on the track (page scroll).
	// Clicking the track above/left of the thumb scrolls one page up/left;
	// clicking below/right scrolls one page down/right.
	// This follows the Windows convention (most predictable for users).
	vTrack, hTrack := w.computeTrackRects()

	if w.canScrollY() && !vTrack.IsEmpty() && vTrack.Contains(e.Position) {
		dir := 1
		if !vThumb.IsEmpty() && e.Position.Y < vThumb.Min.Y {
			dir = -1
		}
		pageSize := w.viewportSize.Height
		setScroll(w, ctx, w.cfg.ResolvedScrollX(), w.cfg.ResolvedScrollY()+float32(dir)*pageSize)

		// Start track repeat for continuous scrolling while held.
		now := ctx.Now()
		w.trackRepeat = trackRepeatState{
			active:    true,
			axis:      dragVertical,
			direction: dir,
			clickPos:  e.Position.Y,
			lastFire:  now,
		}
		ctx.RequestFocus(w)
		return true
	}

	if w.canScrollX() && !hTrack.IsEmpty() && hTrack.Contains(e.Position) {
		dir := 1
		if !hThumb.IsEmpty() && e.Position.X < hThumb.Min.X {
			dir = -1
		}
		pageSize := w.viewportSize.Width
		setScroll(w, ctx, w.cfg.ResolvedScrollX()+float32(dir)*pageSize, w.cfg.ResolvedScrollY())

		// Start track repeat.
		now := ctx.Now()
		w.trackRepeat = trackRepeatState{
			active:    true,
			axis:      dragHorizontal,
			direction: dir,
			clickPos:  e.Position.X,
			lastFire:  now,
		}
		ctx.RequestFocus(w)
		return true
	}

	return false
}

// handleMouseRelease handles mouse button release.
func handleMouseRelease(w *Widget, ctx widget.Context, e *event.MouseEvent) bool {
	if e.Button != event.ButtonLeft {
		return false
	}

	wasDragging := w.dragging != dragNone
	wasRepeating := w.trackRepeat.active
	w.dragging = dragNone
	w.trackRepeat.active = false
	if wasDragging || wasRepeating {
		w.MarkRedrawLocal()
		ctx.InvalidateRect(w.Bounds())
	}
	return wasDragging || wasRepeating
}

// handleMouseMove handles mouse movement for drag tracking.
func handleMouseMove(w *Widget, ctx widget.Context, e *event.MouseEvent) bool {
	if w.dragging == dragNone {
		return false
	}

	// If the left button is no longer pressed, the MouseRelease was lost
	// (e.g., released outside widget bounds). Clear the drag state.
	if !e.Buttons.IsLeftPressed() {
		w.dragging = dragNone
		w.trackRepeat.active = false
		w.MarkRedrawLocal()
		ctx.InvalidateRect(w.Bounds())
		return false
	}

	_, _ = w.computeThumbRects() // ensure rects are computed
	_, hTrack := w.computeTrackRects()
	vTrack, _ := w.computeTrackRects()

	if w.dragging == dragVertical {
		deltaPixels := e.Position.Y - w.dragStart.Y
		trackLen := vTrack.Height()
		thumbSize := computeThumbSize(w.viewportSize.Height, w.contentSize.Height, trackLen)
		scrollableTrack := trackLen - thumbSize
		if scrollableTrack > 0 {
			maxScrollY := w.contentSize.Height - w.viewportSize.Height
			newScrollY := w.dragScrollStart + deltaPixels*(maxScrollY/scrollableTrack)
			setScroll(w, ctx, w.cfg.ResolvedScrollX(), newScrollY)
		}
		return true
	}

	if w.dragging == dragHorizontal {
		deltaPixels := e.Position.X - w.dragStart.X
		trackLen := hTrack.Width()
		thumbSize := computeThumbSize(w.viewportSize.Width, w.contentSize.Width, trackLen)
		scrollableTrack := trackLen - thumbSize
		if scrollableTrack > 0 {
			maxScrollX := w.contentSize.Width - w.viewportSize.Width
			newScrollX := w.dragScrollStart + deltaPixels*(maxScrollX/scrollableTrack)
			setScroll(w, ctx, newScrollX, w.cfg.ResolvedScrollY())
		}
		return true
	}

	return false
}

// handleKeyEvent processes keyboard events for scroll navigation.
func handleKeyEvent(w *Widget, ctx widget.Context, e *event.KeyEvent) bool {
	if !w.IsFocused() {
		return false
	}

	if e.KeyType != event.KeyPress && e.KeyType != event.KeyRepeat {
		return false
	}

	step := w.cfg.resolvedScrollStep()
	scrollX := w.cfg.ResolvedScrollX()
	scrollY := w.cfg.ResolvedScrollY()

	switch e.Key {
	case event.KeyDown:
		setScroll(w, ctx, scrollX, scrollY+step)
		return true
	case event.KeyUp:
		setScroll(w, ctx, scrollX, scrollY-step)
		return true
	case event.KeyRight:
		setScroll(w, ctx, scrollX+step, scrollY)
		return true
	case event.KeyLeft:
		setScroll(w, ctx, scrollX-step, scrollY)
		return true
	case event.KeyPageDown:
		setScroll(w, ctx, scrollX, scrollY+w.viewportSize.Height)
		return true
	case event.KeyPageUp:
		setScroll(w, ctx, scrollX, scrollY-w.viewportSize.Height)
		return true
	case event.KeyHome:
		setScroll(w, ctx, 0, 0)
		return true
	case event.KeyEnd:
		maxY := w.contentSize.Height - w.viewportSize.Height
		if maxY < 0 {
			maxY = 0
		}
		setScroll(w, ctx, scrollX, maxY)
		return true
	default:
		return false
	}
}

// setScroll updates the scroll position, clamping to valid bounds.
func setScroll(w *Widget, ctx widget.Context, rawX, rawY float32) {
	newX := clampScroll(rawX, w.contentSize.Width, w.viewportSize.Width)
	newY := clampScroll(rawY, w.contentSize.Height, w.viewportSize.Height)

	currentX := w.cfg.ResolvedScrollX()
	currentY := w.cfg.ResolvedScrollY()

	if newX == currentX && newY == currentY {
		return
	}

	// TWO-WAY: write back to signals if bound.
	if w.cfg.scrollXSignal != nil {
		w.cfg.scrollXSignal.Set(newX)
	} else {
		w.cfg.scrollX = newX
	}

	if w.cfg.scrollYSignal != nil {
		w.cfg.scrollYSignal.Set(newY)
	} else {
		w.cfg.scrollY = newY
	}

	if w.cfg.onScroll != nil {
		w.cfg.onScroll(newX, newY)
	}

	w.SetNeedsRedraw(true)
	ctx.InvalidateRect(w.Bounds())
}

// Gesture-based scroll helpers removed: ScrollView does not implement
// GestureAware. Scrollbar drag is handled by Event() handler
// (MousePress/Move/Release) with proper scrollbar hit-testing.

// clampScroll clamps a scroll offset to [0, maxScroll].
func clampScroll(offset, contentSize, viewportSize float32) float32 {
	maxScroll := contentSize - viewportSize
	if maxScroll < 0 {
		maxScroll = 0
	}
	if offset < 0 {
		offset = 0
	}
	if offset > maxScroll {
		offset = maxScroll
	}
	return offset
}

// scrollFromTrackClick converts a click position on the track to a scroll offset.
// It centers the thumb at the click position.
func scrollFromTrackClick(clickPos, trackStart, trackLen, viewportSize, contentSize float32) float32 {
	maxScroll := contentSize - viewportSize
	if maxScroll <= 0 {
		return 0
	}
	thumbSize := computeThumbSize(viewportSize, contentSize, trackLen)
	scrollableTrack := trackLen - thumbSize
	if scrollableTrack <= 0 {
		return 0
	}
	// Center thumb at click position.
	thumbCenter := clickPos - trackStart - thumbSize/2
	if thumbCenter < 0 {
		thumbCenter = 0
	}
	if thumbCenter > scrollableTrack {
		thumbCenter = scrollableTrack
	}
	return thumbCenter / scrollableTrack * maxScroll
}

// computeThumbSize calculates the thumb size proportional to viewport/content ratio.
func computeThumbSize(viewportSize, contentSize, trackLen float32) float32 {
	if contentSize <= 0 || viewportSize <= 0 {
		return minThumbSize
	}
	ratio := viewportSize / contentSize
	if ratio >= 1 {
		return trackLen
	}
	thumbSize := ratio * trackLen
	if thumbSize < minThumbSize {
		thumbSize = minThumbSize
	}
	return thumbSize
}

// Scroll step defaults.
const defaultScrollStep float32 = 40

// dragAxis represents which scrollbar is being dragged.
type dragAxis uint8

const (
	dragNone       dragAxis = iota
	dragVertical            // vertical scrollbar thumb is being dragged
	dragHorizontal          // horizontal scrollbar thumb is being dragged
)

// computeThumbPosition calculates the thumb offset within the track.
func computeThumbPosition(scrollOffset, maxScroll, trackLen, thumbSize float32) float32 {
	if maxScroll <= 0 {
		return 0
	}
	scrollableTrack := trackLen - thumbSize
	if scrollableTrack <= 0 {
		return 0
	}
	return (scrollOffset / maxScroll) * scrollableTrack
}

// computeScrollbarRect calculates the scrollbar track rectangle.
func computeScrollbarRect(bounds geometry.Rect, axis dragAxis, hasOther bool) geometry.Rect {
	totalWidth := scrollbarWidth + scrollbarPadding*2

	if axis == dragVertical {
		height := bounds.Height()
		if hasOther {
			height -= totalWidth // Leave space for horizontal scrollbar.
		}
		if height <= 0 {
			return geometry.Rect{}
		}
		return geometry.NewRect(
			bounds.Max.X-totalWidth,
			bounds.Min.Y,
			totalWidth,
			height,
		)
	}

	// Horizontal.
	width := bounds.Width()
	if hasOther {
		width -= totalWidth // Leave space for vertical scrollbar.
	}
	if width <= 0 {
		return geometry.Rect{}
	}
	return geometry.NewRect(
		bounds.Min.X,
		bounds.Max.Y-totalWidth,
		width,
		totalWidth,
	)
}
