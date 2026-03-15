package listview

import (
	"fmt"

	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/core/scrollview"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// Widget implements a virtualized list that renders only visible items.
// It composes an internal [scrollview.Widget] for scroll handling and
// delegates item rendering to a builder callback.
//
// A list view is created with [New] using functional options:
//
//	lv := listview.New(
//	    listview.ItemCount(len(data)),
//	    listview.FixedItemHeight(48),
//	    listview.BuildItem(func(ctx listview.ItemContext) widget.Widget {
//	        return buildRow(data[ctx.Index])
//	    }),
//	)
type Widget struct {
	widget.WidgetBase
	cfg     config
	painter Painter

	// Internal scroll view (composition).
	scroll  *scrollview.Widget
	virtual *virtualContent

	// Height management.
	heights heightManager

	// Widget cache for visible items.
	cache widgetCache

	// Layout state.
	viewportWidth  float32
	viewportHeight float32

	// Interaction state.
	hoveredIndex int

	// End-reached guard to avoid duplicate calls per scroll position.
	endReachedFired bool
}

// New creates a new list view Widget with the given options.
//
// The returned widget is visible, enabled, and focusable by default.
// If no [BuildItem] callback is provided, the list renders empty.
func New(opts ...Option) *Widget {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	w := &Widget{
		cfg:          cfg,
		painter:      DefaultPainter{},
		hoveredIndex: noHoveredIndex,
	}
	w.SetVisible(true)
	w.SetEnabled(true)

	if w.cfg.painter != nil {
		w.painter = w.cfg.painter
	}

	// Initialize height manager.
	w.heights = newHeightManager(&w.cfg)

	// Create virtual content widget.
	w.virtual = &virtualContent{list: w}
	w.virtual.SetVisible(true)

	// Build internal scroll view options.
	svOpts := []scrollview.Option{
		scrollview.DirectionOpt(scrollview.Vertical),
	}
	if w.cfg.scrollYSignal != nil {
		svOpts = append(svOpts, scrollview.ScrollYSignal(w.cfg.scrollYSignal))
	}
	if w.cfg.onScroll != nil {
		fn := w.cfg.onScroll
		svOpts = append(svOpts, scrollview.OnScroll(func(_, y float32) {
			fn(y)
			// Invalidate cache when scroll position changes.
			w.cache.invalidate()
			w.endReachedFired = false
		}))
	} else {
		// Still need to invalidate cache on scroll.
		svOpts = append(svOpts, scrollview.OnScroll(func(_, _ float32) {
			w.cache.invalidate()
			w.endReachedFired = false
		}))
	}

	w.scroll = scrollview.New(w.virtual, svOpts...)

	return w
}

// IsFocusable reports whether the list view can currently receive focus.
func (w *Widget) IsFocusable() bool {
	return w.IsVisible() && w.IsEnabled() && !w.cfg.ResolvedDisabled()
}

// Layout calculates the list view's size within the given constraints.
func (w *Widget) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	// Update item count (may have changed via signal).
	newCount := w.cfg.ResolvedItemCount()
	if newCount != w.heights.count {
		w.heights.updateCount(newCount)
		w.cache.invalidate()
	}

	// The list fills the available space, but clamps infinite dimensions
	// to defaults. Without this, unconstrained parents (e.g., Box with
	// vertical stacking) would cause the viewport to be "infinite",
	// defeating virtualization (all items would be "visible").
	size := constraints.Biggest()
	if size.Width >= geometry.Infinity {
		size.Width = constraints.Constrain(geometry.Sz(defaultViewportWidth, 0)).Width
	}
	if size.Height >= geometry.Infinity {
		// Use total content height clamped to default viewport height.
		totalH := w.heights.totalHeight()
		if totalH > defaultViewportHeight {
			totalH = defaultViewportHeight
		}
		size.Height = totalH
	}
	if size.Width <= 0 {
		size.Width = defaultViewportWidth
	}
	if size.Height <= 0 {
		size.Height = defaultViewportHeight
	}

	w.viewportWidth = size.Width
	w.viewportHeight = size.Height

	// Layout the internal scroll view with concrete (non-infinite) constraints.
	svConstraints := geometry.Tight(size)
	w.scroll.Layout(ctx, svConstraints)

	return size
}

// Draw renders the list view to the canvas.
func (w *Widget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !w.IsVisible() {
		return
	}

	bounds := w.Bounds()
	if bounds.IsEmpty() {
		return
	}

	// Set scroll view bounds to match our bounds.
	w.scroll.SetBounds(bounds)

	// Delegate drawing to the internal scroll view.
	// The scroll view clips, translates, and draws our virtual content.
	w.scroll.Draw(ctx, canvas)
}

// Event handles an input event and returns true if consumed.
func (w *Widget) Event(ctx widget.Context, e event.Event) bool {
	if !w.IsVisible() || !w.IsEnabled() {
		return false
	}

	// Ensure scroll view bounds are set before event dispatch.
	// ScrollView transforms event coordinates using its bounds.
	bounds := w.Bounds()
	if !bounds.IsEmpty() {
		w.scroll.SetBounds(bounds)
	}

	// Handle keyboard events at the list level.
	if ke, ok := e.(*event.KeyEvent); ok {
		if handleListKeyEvent(w, ctx, ke) {
			return true
		}
	}

	// Delegate other events to the scroll view (handles wheel, scrollbar, etc.).
	return w.scroll.Event(ctx, e)
}

// Children returns the internal scroll view as the single child.
func (w *Widget) Children() []widget.Widget {
	if w.scroll == nil {
		return nil
	}
	return []widget.Widget{w.scroll}
}

// Mount creates signal bindings for push-based invalidation.
// Implements [widget.Lifecycle].
func (w *Widget) Mount(ctx widget.Context) {
	sched := ctx.Scheduler()
	if sched == nil {
		return
	}

	// Bind item count signals.
	if w.cfg.readonlyItemCountSignal != nil {
		b := state.BindToScheduler(w.cfg.readonlyItemCountSignal, w, sched)
		w.AddBinding(b)
	} else if w.cfg.itemCountSignal != nil {
		b := state.BindToScheduler(w.cfg.itemCountSignal, w, sched)
		w.AddBinding(b)
	}

	// Bind selected index signals.
	if w.cfg.readonlySelectedIndexSignal != nil {
		b := state.BindToScheduler(w.cfg.readonlySelectedIndexSignal, w, sched)
		w.AddBinding(b)
	} else if w.cfg.selectedIndexSignal != nil {
		b := state.BindToScheduler(w.cfg.selectedIndexSignal, w, sched)
		w.AddBinding(b)
	}

	// Bind disabled signals.
	if w.cfg.readonlyDisabledSignal != nil {
		b := state.BindToScheduler(w.cfg.readonlyDisabledSignal, w, sched)
		w.AddBinding(b)
	} else if w.cfg.disabledSignal != nil {
		b := state.BindToScheduler(w.cfg.disabledSignal, w, sched)
		w.AddBinding(b)
	}

	// Mount internal scroll view.
	w.scroll.Mount(ctx)
}

// Unmount is called when the list view is removed from the widget tree.
// Implements [widget.Lifecycle].
func (w *Widget) Unmount() {
	// Unmount internal scroll view.
	w.scroll.Unmount()
	// Bindings are cleaned up automatically by WidgetBase.CleanupBindings().
}

// --- Public API ---

// ScrollToIndex scrolls to make the item at the given index visible.
// If the item is already fully visible, this is a no-op.
func (w *Widget) ScrollToIndex(index int) {
	itemCount := w.cfg.ResolvedItemCount()
	if index < 0 || index >= itemCount {
		return
	}

	itemTop := w.heights.offsetAt(index)
	itemBottom := itemTop + w.heights.heightAt(index)
	scrollY := w.currentScrollY()

	// Already visible: no-op.
	if itemTop >= scrollY && itemBottom <= scrollY+w.viewportHeight {
		return
	}

	var newScrollY float32
	if itemTop < scrollY {
		// Item is above viewport -- scroll up.
		newScrollY = itemTop
	} else {
		// Item is below viewport -- scroll down.
		newScrollY = itemBottom - w.viewportHeight
		if newScrollY < 0 {
			newScrollY = 0
		}
	}

	w.setScrollY(newScrollY)
}

// VisibleRange returns the indices of currently visible items [start, end).
func (w *Widget) VisibleRange() (start, end int) {
	return w.heights.visibleRange(w.currentScrollY(), w.viewportHeight, 0)
}

// InvalidateData signals that the underlying data has changed.
// This invalidates the widget cache and triggers re-layout.
func (w *Widget) InvalidateData() {
	w.cache.invalidate()
	w.heights.updateCount(w.cfg.ResolvedItemCount())
	if w.heights.mode == heightLazy {
		w.heights.initLazy()
	}
}

// GetItemCount returns the current item count.
func (w *Widget) GetItemCount() int {
	return w.cfg.ResolvedItemCount()
}

// --- Accessibility ---

// AccessibilityRole returns the ARIA role for this widget.
func (w *Widget) AccessibilityRole() a11y.Role {
	return a11y.RoleList
}

// AccessibilityLabel returns the accessibility label.
func (w *Widget) AccessibilityLabel() string {
	if w.cfg.a11yLabel != "" {
		return w.cfg.a11yLabel
	}
	return "List"
}

// AccessibilityHint returns the accessibility hint.
func (w *Widget) AccessibilityHint() string {
	return ""
}

// AccessibilityValue returns the current list state as a string.
func (w *Widget) AccessibilityValue() string {
	count := w.cfg.ResolvedItemCount()
	selected := w.cfg.ResolvedSelectedIndex()
	if selected >= 0 && selected < count {
		return fmt.Sprintf("Item %d of %d selected", selected+1, count)
	}
	return fmt.Sprintf("%d items", count)
}

// AccessibilityState returns the current accessibility state.
func (w *Widget) AccessibilityState() a11y.State {
	return a11y.State{
		Disabled: w.cfg.ResolvedDisabled(),
	}
}

// AccessibilityActions returns the list of supported actions.
func (w *Widget) AccessibilityActions() []a11y.Action {
	return []a11y.Action{a11y.ActionScrollUp, a11y.ActionScrollDown}
}

// --- Internal helpers ---

// currentScrollY returns the current vertical scroll offset.
func (w *Widget) currentScrollY() float32 {
	_, y := w.scroll.ScrollOffset()
	return y
}

// setScrollY updates the scroll Y position.
func (w *Widget) setScrollY(y float32) {
	if w.cfg.scrollYSignal != nil {
		w.cfg.scrollYSignal.Set(y)
	}
	// The scroll view will pick up the new value on next layout/draw.
}

// viewportBounds returns the viewport bounds in local coordinates.
func (w *Widget) viewportBounds() geometry.Rect {
	return geometry.NewRect(0, 0, w.viewportWidth, w.viewportHeight)
}

// checkEndReached fires the OnEndReached callback if the visible range
// is close to the end of the list.
func (w *Widget) checkEndReached(visibleEnd, itemCount int) {
	if w.cfg.onEndReached == nil || w.endReachedFired {
		return
	}
	threshold := w.cfg.endReachedThreshold
	if threshold <= 0 {
		threshold = defaultEndReachedThreshold
	}
	if itemCount > 0 && visibleEnd >= itemCount-threshold {
		w.endReachedFired = true
		w.cfg.onEndReached()
	}
}

// Default viewport dimensions used as fallback.
const (
	defaultViewportWidth  float32 = 200
	defaultViewportHeight float32 = 400
)

// Verify Widget implements required interfaces at compile time.
var (
	_ widget.Widget    = (*Widget)(nil)
	_ widget.Focusable = (*Widget)(nil)
	_ widget.Lifecycle = (*Widget)(nil)
	_ a11y.Accessible  = (*Widget)(nil)
)
