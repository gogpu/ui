package tabview

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// Widget implements a tabbed navigation container with configurable appearance.
//
// A tabview is created with [New] using functional options:
//
//	tv := tabview.New(
//	    []tabview.Tab{
//	        {Label: "Home", Content: homeWidget},
//	        {Label: "Settings", Content: settingsWidget},
//	    },
//	    tabview.PositionOpt(tabview.Top),
//	    tabview.OnSelect(func(idx int) { log.Println("selected:", idx) }),
//	)
type Widget struct {
	widget.WidgetBase
	cfg     config
	painter Painter

	// Computed layout state.
	tabBarBounds geometry.Rect
	tabStates    []TabState
}

// New creates a new tabview Widget with the given tabs and options.
//
// The returned widget is visible, enabled, and focusable by default.
func New(tabs []Tab, opts ...Option) *Widget {
	w := &Widget{
		painter: DefaultPainter{},
	}
	w.SetVisible(true)
	w.SetEnabled(true)

	// Copy tabs to prevent external mutation.
	w.cfg.tabs = make([]Tab, len(tabs))
	copy(w.cfg.tabs, tabs)

	for _, opt := range opts {
		opt(&w.cfg)
	}

	// Apply painter from config if set.
	if w.cfg.painter != nil {
		w.painter = w.cfg.painter
	}

	// Initialize tab states.
	w.tabStates = make([]TabState, len(tabs))

	return w
}

// IsFocusable reports whether the tabview can currently receive focus.
func (w *Widget) IsFocusable() bool {
	return w.IsVisible() && w.IsEnabled()
}

// Layout calculates the tabview's preferred size within the given constraints.
// Only the selected tab's content is laid out (lazy).
func (w *Widget) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	totalSize := constraints.Constrain(geometry.Sz(constraints.MaxWidth, constraints.MaxHeight))

	// Calculate tab bar bounds and individual tab bounds.
	w.computeTabLayout(totalSize)

	// Sync tab states (disabled, closeable, etc.) for event handling.
	w.updateTabStates(w.cfg.ResolvedSelected())

	// Layout only the selected tab's content.
	selectedIdx := w.cfg.ResolvedSelected()
	if selectedIdx >= 0 && selectedIdx < len(w.cfg.tabs) {
		tab := &w.cfg.tabs[selectedIdx]
		if tab.Content != nil {
			contentBounds := w.contentBounds(totalSize)
			contentConstraints := geometry.Tight(contentBounds.Size())
			tab.Content.Layout(ctx, contentConstraints)

			// Set bounds on content widget.
			if setter, ok := tab.Content.(interface{ SetBounds(geometry.Rect) }); ok {
				setter.SetBounds(contentBounds)
			}
		}
	}

	return totalSize
}

// Draw renders the tabview to the canvas.
// Only the selected tab's content is drawn (lazy).
func (w *Widget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if w.Bounds().IsEmpty() {
		return
	}

	// Update tab states for the painter.
	selectedIdx := w.cfg.ResolvedSelected()
	w.updateTabStates(selectedIdx)

	// Paint the tab bar.
	w.painter.PaintTabBar(canvas, PaintState{
		Bounds:      w.tabBarBounds,
		Tabs:        w.tabStates,
		SelectedIdx: selectedIdx,
		Position:    w.cfg.position,
		Focused:     w.IsFocused(),
	})

	// Draw only the selected tab's content.
	if selectedIdx >= 0 && selectedIdx < len(w.cfg.tabs) {
		tab := &w.cfg.tabs[selectedIdx]
		if tab.Content != nil {
			contentBounds := w.contentBounds(w.Bounds().Size())
			canvas.PushClip(contentBounds)
			widget.StampScreenOrigin(tab.Content, canvas)
			tab.Content.Draw(ctx, canvas)
			canvas.PopClip()
		}
	}
}

// Event handles an input event and returns true if consumed.
func (w *Widget) Event(ctx widget.Context, e event.Event) bool {
	// Forward events to selected tab's content first.
	selectedIdx := w.cfg.ResolvedSelected()
	if selectedIdx >= 0 && selectedIdx < len(w.cfg.tabs) {
		tab := &w.cfg.tabs[selectedIdx]
		if tab.Content != nil {
			if tab.Content.Event(ctx, e) {
				return true
			}
		}
	}

	return handleEvent(w, ctx, e)
}

// Children returns the content widgets of all tabs.
// Only the selected tab's content is laid out and drawn, but all
// content widgets are reported for tree traversal purposes.
func (w *Widget) Children() []widget.Widget {
	var children []widget.Widget
	for i := range w.cfg.tabs {
		if w.cfg.tabs[i].Content != nil {
			children = append(children, w.cfg.tabs[i].Content)
		}
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
	if w.cfg.readonlySelectedSignal != nil {
		b := state.BindToScheduler(w.cfg.readonlySelectedSignal, w, sched)
		w.AddBinding(b)
	} else if w.cfg.selectedSignal != nil {
		b := state.BindToScheduler(w.cfg.selectedSignal, w, sched)
		w.AddBinding(b)
	}
}

// Unmount is called when the tabview is removed from the widget tree.
// Implements [widget.Lifecycle].
func (w *Widget) Unmount() {
	// Bindings are cleaned up automatically by WidgetBase.CleanupBindings().
}

// TabCount returns the number of tabs.
func (w *Widget) TabCount() int {
	return len(w.cfg.tabs)
}

// SelectedIndex returns the currently selected tab index.
func (w *Widget) SelectedIndex() int {
	return w.cfg.ResolvedSelected()
}

// computeTabLayout calculates tab bar and individual tab bounds.
func (w *Widget) computeTabLayout(totalSize geometry.Size) {
	bounds := w.Bounds()
	tabCount := len(w.cfg.tabs)

	// Tab bar position.
	switch w.cfg.position {
	case Bottom:
		w.tabBarBounds = geometry.NewRect(
			bounds.Min.X,
			bounds.Min.Y+totalSize.Height-tabBarHeight,
			totalSize.Width,
			tabBarHeight,
		)
	default: // Top
		w.tabBarBounds = geometry.NewRect(
			bounds.Min.X,
			bounds.Min.Y,
			totalSize.Width,
			tabBarHeight,
		)
	}

	if tabCount == 0 {
		return
	}

	// Equal-width tabs.
	tabWidth := totalSize.Width / float32(tabCount)
	for i := range w.tabStates {
		x := w.tabBarBounds.Min.X + float32(i)*tabWidth
		w.tabStates[i].Bounds = geometry.NewRect(
			x,
			w.tabBarBounds.Min.Y,
			tabWidth,
			tabBarHeight,
		)

		// Close button bounds.
		isCloseable := w.cfg.closeable || w.cfg.tabs[i].Closeable
		if isCloseable && !w.cfg.tabs[i].Disabled {
			cbX := x + tabWidth - tabPaddingX - closeButtonSize
			cbY := w.tabBarBounds.Min.Y + (tabBarHeight-closeButtonSize)/2
			w.tabStates[i].CloseButtonBounds = geometry.NewRect(
				cbX, cbY,
				closeButtonSize, closeButtonSize,
			)
		} else {
			w.tabStates[i].CloseButtonBounds = geometry.Rect{}
		}
	}
}

// contentBounds returns the bounds for the content area.
func (w *Widget) contentBounds(totalSize geometry.Size) geometry.Rect {
	bounds := w.Bounds()
	contentHeight := totalSize.Height - tabBarHeight
	switch w.cfg.position {
	case Bottom:
		return geometry.NewRect(
			bounds.Min.X,
			bounds.Min.Y,
			totalSize.Width,
			contentHeight,
		)
	default: // Top
		return geometry.NewRect(
			bounds.Min.X,
			bounds.Min.Y+tabBarHeight,
			totalSize.Width,
			contentHeight,
		)
	}
}

// updateTabStates refreshes the tab states for the painter.
func (w *Widget) updateTabStates(selectedIdx int) {
	for i := range w.tabStates {
		w.tabStates[i].Label = w.cfg.tabs[i].Label
		w.tabStates[i].Selected = i == selectedIdx
		w.tabStates[i].Disabled = w.cfg.tabs[i].Disabled
		w.tabStates[i].Closeable = w.cfg.closeable || w.cfg.tabs[i].Closeable
	}
}

// Verify Widget implements required interfaces at compile time.
var (
	_ widget.Widget    = (*Widget)(nil)
	_ widget.Focusable = (*Widget)(nil)
	_ widget.Lifecycle = (*Widget)(nil)
)
