package titlebar

import (
	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// WindowChrome provides window management operations.
//
// This interface is a subset of gpucontext.WindowChrome, defined locally
// to avoid a direct dependency on the gpucontext package. Callers should
// pass the gpucontext.WindowChrome value obtained from the window provider.
type WindowChrome interface {
	// Minimize minimizes the window.
	Minimize()

	// Maximize maximizes the window.
	Maximize()

	// IsMaximized reports whether the window is currently maximized.
	IsMaximized() bool

	// Close closes the window.
	Close()
}

// HitTestResult identifies what region of the title bar a point falls in.
// This is used by the window system to determine drag, resize, and
// control button behavior.
type HitTestResult int

const (
	// HitTestClient means the point is over a child widget (not draggable).
	HitTestClient HitTestResult = iota

	// HitTestCaption means the point is over empty title bar area (drag to move).
	HitTestCaption

	// HitTestClose means the point is over the close button.
	HitTestClose

	// HitTestMaximize means the point is over the maximize/restore button.
	HitTestMaximize

	// HitTestMinimize means the point is over the minimize button.
	HitTestMinimize
)

// controlCount is the number of window control buttons (minimize, maximize, close).
const controlCount = 3

// controlIndex maps ControlType to control button index (0=min, 1=max, 2=close).
const (
	controlIdxMinimize = 0
	controlIdxMaximize = 1
	controlIdxClose    = 2
)

// interactionState represents the current user interaction state for a control button.
type interactionState uint8

const (
	stateNormal  interactionState = iota
	stateHover                    // mouse is over the control
	statePressed                  // mouse button is held down
)

// Widget implements a custom window title bar.
//
// The title bar renders a horizontal bar with three zones: leading (left),
// center, and trailing (right). The trailing zone contains window control
// buttons (minimize, maximize/restore, close) when a [WindowChrome] is provided.
//
// Empty space in the title bar acts as a drag region for window movement.
// Child widgets placed in the leading and center zones receive events normally.
//
// A Widget is created with [New] using functional options:
//
//	tb := titlebar.New(
//	    titlebar.Title("My App"),
//	    titlebar.Leading(menuBtn),
//	    titlebar.Chrome(windowChrome),
//	    titlebar.PainterOpt(painter),
//	)
type Widget struct {
	widget.WidgetBase

	cfg     config
	painter Painter

	// controlBounds holds the bounds for each control button (min, max, close).
	controlBounds [controlCount]geometry.Rect

	// controlStates holds the interaction state for each control button.
	controlStates [controlCount]interactionState

	// leadingBounds holds the calculated bounds for each leading child widget.
	leadingBounds []geometry.Rect

	// centerBounds holds the calculated bounds for each center child widget.
	centerBounds []geometry.Rect

	// hoveredChild tracks the currently hovered child for MouseLeave dispatch.
	// nil means no child is hovered.
	hoveredChild widget.Widget
}

// New creates a new title bar Widget with the given options.
//
// The returned widget is visible and enabled by default. The default
// height is 40 logical pixels.
func New(opts ...Option) *Widget {
	w := &Widget{
		painter: DefaultPainter{},
	}
	w.SetVisible(true)
	w.SetEnabled(true)

	w.cfg.height = defaultBarHeight
	w.cfg.focused = true

	for _, opt := range opts {
		opt(&w.cfg)
	}

	if w.cfg.painter != nil {
		w.painter = w.cfg.painter
	}

	w.leadingBounds = make([]geometry.Rect, len(w.cfg.leading))
	w.centerBounds = make([]geometry.Rect, len(w.cfg.center))

	// ADR-028: parent chain for upward dirty propagation.
	// Flutter: RenderObject.adoptChild sets parent on each child.
	type parentSetter interface{ SetParent(widget.Widget) }
	for _, child := range w.cfg.leading {
		if child != nil {
			if ps, ok := child.(parentSetter); ok {
				ps.SetParent(w)
			}
		}
	}
	for _, child := range w.cfg.center {
		if child != nil {
			if ps, ok := child.(parentSetter); ok {
				ps.SetParent(w)
			}
		}
	}

	return w
}

// Default configuration values.
const (
	defaultBarHeight    float32 = 40
	controlButtonWidth  float32 = 46
	controlButtonHeight float32 = 40
	childGap            float32 = 2
	leadingPadding      float32 = 4
	titleFontSize       float32 = 12
)

// Title returns the current title text.
func (w *Widget) Title() string {
	return w.cfg.title
}

// SetTitle updates the title text.
func (w *Widget) SetTitle(title string) {
	w.cfg.title = title
}

// SetFocusedState updates whether the window is focused, affecting visual appearance.
func (w *Widget) SetFocusedState(focused bool) {
	w.cfg.focused = focused
}

// HasChrome reports whether a WindowChrome is available.
func (w *Widget) HasChrome() bool {
	return w.cfg.chrome != nil
}

// HitTest returns what region of the title bar a screen-space point falls in.
// This is intended to be called from a hit-test callback registered with the
// window system.
func (w *Widget) HitTest(x, y float64) HitTestResult {
	pt := geometry.Pt(float32(x), float32(y))
	screenBounds := w.ScreenBounds()

	// Not in title bar at all.
	if !screenBounds.Contains(pt) {
		return HitTestClient
	}

	local := pt.Sub(screenBounds.Min)

	// Check control buttons (reverse order: close first for priority).
	if w.cfg.chrome != nil {
		for i := controlCount - 1; i >= 0; i-- {
			if w.controlBounds[i].Contains(local) {
				switch i {
				case controlIdxClose:
					return HitTestClose
				case controlIdxMaximize:
					return HitTestMaximize
				case controlIdxMinimize:
					return HitTestMinimize
				}
			}
		}
	}

	// Check leading and center child widgets.
	if hitTestChildren(w.cfg.leading, w.leadingBounds, local) {
		return HitTestClient
	}
	if hitTestChildren(w.cfg.center, w.centerBounds, local) {
		return HitTestClient
	}

	// Empty space = caption (drag to move).
	return HitTestCaption
}

// hitTestChildren checks if local point hits any child widget.
// Delegates to child.HitTestPoint if available (toolbar gaps = no hit).
func hitTestChildren(children []widget.Widget, bounds []geometry.Rect, local geometry.Point) bool {
	for i, child := range children {
		if child == nil || !bounds[i].Contains(local) {
			continue
		}
		if ht, ok := child.(interface{ HitTestPoint(geometry.Point) bool }); ok {
			childLocal := local.Sub(bounds[i].Min)
			if ht.HitTestPoint(childLocal) {
				return true
			}
			continue
		}
		return true
	}
	return false
}

// --- widget.Widget interface ---

// Layout calculates the title bar's preferred size within the given constraints.
func (w *Widget) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	preferred := geometry.Sz(constraints.MaxWidth, w.cfg.height)
	size := constraints.Constrain(preferred)

	w.layoutChildren(ctx, size)

	w.SetBounds(geometry.FromPointSize(w.Position(), size))
	return size
}

// layoutChildren calculates bounds for leading, center, and control button zones.
func (w *Widget) layoutChildren(ctx widget.Context, available geometry.Size) {
	// Calculate control buttons zone (right side).
	controlsWidth := w.controlsZoneWidth()
	controlsX := available.Width - controlsWidth

	// Layout control buttons.
	for i := 0; i < controlCount; i++ {
		x := controlsX + float32(i)*controlButtonWidth
		w.controlBounds[i] = geometry.NewRect(x, 0, controlButtonWidth, available.Height)
	}

	// Layout leading children (left side).
	x := leadingPadding
	for i, child := range w.cfg.leading {
		if child == nil {
			w.leadingBounds[i] = geometry.Rect{}
			continue
		}
		if i > 0 {
			x += childGap
		}
		childConstraints := geometry.Loose(geometry.Sz(controlsX-x, available.Height))
		sz := child.Layout(ctx, childConstraints)
		cy := (available.Height - sz.Height) / 2
		bounds := geometry.NewRect(x, cy, sz.Width, sz.Height)
		w.leadingBounds[i] = bounds
		setBounds(child, bounds)
		x += sz.Width
	}

	// Calculate center zone: between leading end and controls start.
	leadingEnd := x
	centerAvailable := controlsX - leadingEnd

	// Layout center children.
	var centerTotalWidth float32
	centerSizes := make([]geometry.Size, len(w.cfg.center))
	for i, child := range w.cfg.center {
		if child == nil {
			continue
		}
		if i > 0 {
			centerTotalWidth += childGap
		}
		childConstraints := geometry.Loose(geometry.Sz(centerAvailable, available.Height))
		sz := child.Layout(ctx, childConstraints)
		centerSizes[i] = sz
		centerTotalWidth += sz.Width
	}

	// Center the center zone widgets.
	centerStart := leadingEnd + (centerAvailable-centerTotalWidth)/2
	if centerStart < leadingEnd {
		centerStart = leadingEnd
	}
	cx := centerStart
	for i, child := range w.cfg.center {
		if child == nil {
			w.centerBounds[i] = geometry.Rect{}
			continue
		}
		if i > 0 {
			cx += childGap
		}
		cy := (available.Height - centerSizes[i].Height) / 2
		bounds := geometry.NewRect(cx, cy, centerSizes[i].Width, centerSizes[i].Height)
		w.centerBounds[i] = bounds
		setBounds(child, bounds)
		cx += centerSizes[i].Width
	}
}

// controlsZoneWidth returns the total width of the window controls zone.
func (w *Widget) controlsZoneWidth() float32 {
	if w.cfg.chrome == nil {
		return 0
	}
	return controlButtonWidth * controlCount
}

// Draw renders the title bar background, child widgets, and control buttons.
func (w *Widget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !w.IsVisible() {
		return
	}

	bounds := w.Bounds()

	// Draw background.
	w.painter.PaintBackground(canvas, bounds, BackgroundState{
		Focused: w.cfg.focused,
	})

	// Draw children with transform offset.
	canvas.PushTransform(bounds.Min)

	// Draw leading children.
	for _, child := range w.cfg.leading {
		if child == nil {
			continue
		}
		widget.StampScreenOrigin(child, canvas)
		child.Draw(ctx, canvas)
	}

	// Draw title text if no center children and title is set.
	if len(w.cfg.center) == 0 && w.cfg.title != "" {
		controlsWidth := w.controlsZoneWidth()
		titleBounds := geometry.NewRect(0, 0, bounds.Width()-controlsWidth, bounds.Height())
		canvas.DrawText(w.cfg.title, titleBounds, titleFontSize, defaultControlFg, false, widget.TextAlignCenter)
	}

	// Draw center children.
	for _, child := range w.cfg.center {
		if child == nil {
			continue
		}
		widget.StampScreenOrigin(child, canvas)
		child.Draw(ctx, canvas)
	}

	// Draw control buttons.
	if w.cfg.chrome != nil {
		for i := 0; i < controlCount; i++ {
			ct := w.controlTypeForIndex(i)
			w.painter.PaintControlButton(canvas, w.controlBounds[i], ct, ControlState{
				Hovered: w.controlStates[i] == stateHover,
				Pressed: w.controlStates[i] == statePressed,
			})
		}
	}

	canvas.PopTransform()
}

// controlTypeForIndex returns the ControlType for the given control index.
func (w *Widget) controlTypeForIndex(idx int) ControlType {
	switch idx {
	case controlIdxMinimize:
		return ControlMinimize
	case controlIdxMaximize:
		if w.cfg.chrome != nil && w.cfg.chrome.IsMaximized() {
			return ControlRestore
		}
		return ControlMaximize
	case controlIdxClose:
		return ControlClose
	default:
		return ControlClose
	}
}

// Event handles input events for the title bar and its children.
func (w *Widget) Event(ctx widget.Context, e event.Event) bool {
	if !w.IsVisible() || !w.IsEnabled() {
		return false
	}

	switch ev := e.(type) {
	case *event.MouseEvent:
		return w.handleMouseEvent(ctx, ev)
	default:
		// Dispatch to child widgets.
		return w.dispatchToChildren(ctx, e)
	}
}

// handleMouseEvent processes mouse events for the title bar.
func (w *Widget) handleMouseEvent(ctx widget.Context, e *event.MouseEvent) bool {
	local := e.Position.Sub(w.Bounds().Min)

	switch e.MouseType {
	case event.MousePress:
		if e.Button != event.ButtonLeft {
			return false
		}
		return w.handlePress(ctx, local)
	case event.MouseRelease:
		if e.Button != event.ButtonLeft {
			return false
		}
		return w.handleRelease(ctx, local)
	case event.MouseMove:
		return w.handleMove(ctx, local)
	case event.MouseLeave:
		// Send MouseLeave to hovered child before clearing.
		if w.hoveredChild != nil {
			w.hoveredChild.Event(ctx, &event.MouseEvent{
				MouseType: event.MouseLeave,
				Position:  local,
			})
			w.hoveredChild = nil
		}
		return w.clearControlHovers(ctx)
	default:
		return false
	}
}

// handlePress processes a left mouse press.
func (w *Widget) handlePress(ctx widget.Context, local geometry.Point) bool {
	// Check control buttons first.
	if w.cfg.chrome != nil {
		for i := 0; i < controlCount; i++ {
			if w.controlBounds[i].Contains(local) {
				w.controlStates[i] = statePressed
				w.SetNeedsRedraw(true)
				ctx.InvalidateRect(w.Bounds())
				return true
			}
		}
	}

	// Dispatch to leading children.
	for i, child := range w.cfg.leading {
		if child == nil {
			continue
		}
		if w.leadingBounds[i].Contains(local) {
			return child.Event(ctx, &event.MouseEvent{
				MouseType: event.MousePress,
				Button:    event.ButtonLeft,
				Position:  local,
			})
		}
	}

	// Dispatch to center children.
	for i, child := range w.cfg.center {
		if child == nil {
			continue
		}
		if w.centerBounds[i].Contains(local) {
			return child.Event(ctx, &event.MouseEvent{
				MouseType: event.MousePress,
				Button:    event.ButtonLeft,
				Position:  local,
			})
		}
	}

	// Caption area -- consumed (allows drag).
	return true
}

// handleRelease processes a left mouse release.
func (w *Widget) handleRelease(ctx widget.Context, local geometry.Point) bool {
	// Find pressed control and fire action.
	for i := 0; i < controlCount; i++ {
		if w.controlStates[i] != statePressed {
			continue
		}

		// Release always clears pressed state.
		if w.controlBounds[i].Contains(local) {
			w.controlStates[i] = stateHover
		} else {
			w.controlStates[i] = stateNormal
		}
		w.SetNeedsRedraw(true)
		ctx.InvalidateRect(w.Bounds())

		// Fire action only if released on the same button.
		if w.controlBounds[i].Contains(local) && w.cfg.chrome != nil {
			w.fireControlAction(i)
		}
		return true
	}

	// Dispatch to leading/center children.
	for i, child := range w.cfg.leading {
		if child == nil {
			continue
		}
		if w.leadingBounds[i].Contains(local) {
			return child.Event(ctx, &event.MouseEvent{
				MouseType: event.MouseRelease,
				Button:    event.ButtonLeft,
				Position:  local,
			})
		}
	}
	for i, child := range w.cfg.center {
		if child == nil {
			continue
		}
		if w.centerBounds[i].Contains(local) {
			return child.Event(ctx, &event.MouseEvent{
				MouseType: event.MouseRelease,
				Button:    event.ButtonLeft,
				Position:  local,
			})
		}
	}

	return true
}

// fireControlAction performs the window action for the given control index.
func (w *Widget) fireControlAction(idx int) {
	switch idx {
	case controlIdxMinimize:
		w.cfg.chrome.Minimize()
	case controlIdxMaximize:
		w.cfg.chrome.Maximize()
	case controlIdxClose:
		w.cfg.chrome.Close()
	}
}

// handleMove updates hover states for control buttons and children.
func (w *Widget) handleMove(ctx widget.Context, local geometry.Point) bool {
	changed := false

	if w.cfg.chrome != nil {
		for i := 0; i < controlCount; i++ {
			newState := resolveHover(w.controlStates[i], w.controlBounds[i].Contains(local))
			if w.controlStates[i] != newState {
				w.controlStates[i] = newState
				changed = true
			}
		}
	}

	// Find which child (if any) the cursor is over.
	var currentChild widget.Widget
	for i, child := range w.cfg.leading {
		if child != nil && w.leadingBounds[i].Contains(local) {
			currentChild = child
			break
		}
	}
	if currentChild == nil {
		for i, child := range w.cfg.center {
			if child != nil && w.centerBounds[i].Contains(local) {
				currentChild = child
				break
			}
		}
	}

	// Send MouseLeave to previous child if cursor moved away.
	if w.hoveredChild != nil && w.hoveredChild != currentChild {
		w.hoveredChild.Event(ctx, &event.MouseEvent{
			MouseType: event.MouseLeave,
			Position:  local,
		})
		changed = true
	}

	// Send MouseMove to current child.
	if currentChild != nil {
		currentChild.Event(ctx, &event.MouseEvent{
			MouseType: event.MouseMove,
			Position:  local,
		})
		changed = true
	}

	w.hoveredChild = currentChild

	if changed {
		w.SetNeedsRedraw(true)
		ctx.InvalidateRect(w.Bounds())
	}
	return changed
}

// resolveHover determines the new state for a control during mouse move.
// Pressed controls retain their state regardless of cursor position.
func resolveHover(current interactionState, underCursor bool) interactionState {
	if current == statePressed {
		return statePressed
	}
	if underCursor {
		return stateHover
	}
	return stateNormal
}

// clearControlHovers resets all control interaction states to normal.
func (w *Widget) clearControlHovers(ctx widget.Context) bool {
	changed := false
	for i := 0; i < controlCount; i++ {
		if w.controlStates[i] != stateNormal {
			w.controlStates[i] = stateNormal
			changed = true
		}
	}
	if changed {
		w.SetNeedsRedraw(true)
		ctx.InvalidateRect(w.Bounds())
	}
	return changed
}

// dispatchToChildren dispatches non-mouse events to child widgets.
func (w *Widget) dispatchToChildren(ctx widget.Context, e event.Event) bool {
	for _, child := range w.cfg.leading {
		if child == nil {
			continue
		}
		if child.Event(ctx, e) {
			return true
		}
	}
	for _, child := range w.cfg.center {
		if child == nil {
			continue
		}
		if child.Event(ctx, e) {
			return true
		}
	}
	return false
}

// Children returns all child widgets (leading + center).
func (w *Widget) Children() []widget.Widget {
	var children []widget.Widget
	for _, child := range w.cfg.leading {
		if child != nil {
			children = append(children, child)
		}
	}
	for _, child := range w.cfg.center {
		if child != nil {
			children = append(children, child)
		}
	}
	if len(children) == 0 {
		return nil
	}
	return children
}

// IsFocusable reports whether the title bar can receive keyboard focus.
// Title bars are not typically focusable.
func (w *Widget) IsFocusable() bool {
	return false
}

// --- a11y.Accessible interface ---

// AccessibilityRole returns [a11y.RoleBanner].
func (w *Widget) AccessibilityRole() a11y.Role {
	return a11y.RoleBanner
}

// AccessibilityLabel returns "Title Bar".
func (w *Widget) AccessibilityLabel() string {
	return a11yLabel
}

// AccessibilityHint returns the window title as a hint.
func (w *Widget) AccessibilityHint() string {
	return w.cfg.title
}

// AccessibilityValue returns an empty string.
func (w *Widget) AccessibilityValue() string {
	return ""
}

// AccessibilityState returns the accessibility state.
func (w *Widget) AccessibilityState() a11y.State {
	return a11y.State{
		Disabled: !w.IsEnabled(),
		Hidden:   !w.IsVisible(),
	}
}

// AccessibilityActions returns nil.
func (w *Widget) AccessibilityActions() []a11y.Action {
	return nil
}

const a11yLabel = "Title Bar"

// setBounds sets the bounds on a widget that supports it.
func setBounds(child widget.Widget, bounds geometry.Rect) {
	if setter, ok := child.(interface{ SetBounds(geometry.Rect) }); ok {
		setter.SetBounds(bounds)
	}
}

// Compile-time interface checks.
var (
	_ widget.Widget    = (*Widget)(nil)
	_ widget.Focusable = (*Widget)(nil)
	_ a11y.Accessible  = (*Widget)(nil)
)
