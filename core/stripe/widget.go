package stripe

import (
	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// interactionState represents the current user interaction state for a button.
type interactionState uint8

const (
	stateNormal  interactionState = iota
	stateHover                    // mouse is over the button
	statePressed                  // mouse button is held down
)

// buttonState tracks per-button interaction and layout data.
type buttonState struct {
	interaction interactionState
	bounds      geometry.Rect
}

// Widget implements a vertical tool window sidebar strip with icon buttons.
//
// A stripe is created with [New] using functional options:
//
//	s := stripe.New(
//	    stripe.TopItems(
//	        stripe.Button{ID: "project", Label: "Project", Icon: icon.FolderClosed},
//	    ),
//	    stripe.BottomItems(
//	        stripe.Button{ID: "terminal", Label: "Terminal", Icon: icon.Terminal},
//	    ),
//	    stripe.ActiveID("terminal"),
//	    stripe.ShowLabels(true),
//	)
type Widget struct {
	widget.WidgetBase

	cfg      config
	painter  Painter
	activeID string

	topStates    []buttonState
	bottomStates []buttonState
	hoveredIdx   int // index into allButtons() or noHover
}

// noHover is the sentinel value indicating no button is hovered.
const noHover = -1

// New creates a new stripe Widget with the given options.
//
// The returned widget is visible and enabled by default.
func New(opts ...Option) *Widget {
	w := &Widget{
		painter:    DefaultPainter{},
		hoveredIdx: noHover,
	}
	w.SetVisible(true)
	w.SetEnabled(true)

	for _, opt := range opts {
		opt(&w.cfg)
	}

	w.activeID = w.cfg.activeID

	if w.cfg.painter != nil {
		w.painter = w.cfg.painter
	}

	w.topStates = make([]buttonState, len(w.cfg.topItems))
	w.bottomStates = make([]buttonState, len(w.cfg.bottomItems))
	return w
}

// ActiveButtonID returns the ID of the currently active button.
func (w *Widget) ActiveButtonID() string {
	return w.activeID
}

// SetActiveID sets the currently active button by ID.
func (w *Widget) SetActiveID(id string) {
	w.activeID = id
}

// TopItemCount returns the number of top-group buttons.
func (w *Widget) TopItemCount() int {
	return len(w.cfg.topItems)
}

// BottomItemCount returns the number of bottom-group buttons.
func (w *Widget) BottomItemCount() int {
	return len(w.cfg.bottomItems)
}

// --- Sizing ---

// Default sizing constants.
const (
	defaultWidthLabels    float32 = 64 // default width with labels
	defaultWidthIconsOnly float32 = 40 // default width without labels
	buttonHeightLabels    float32 = 57 // button height with 1-line label
	buttonHeightIconsOnly float32 = 40 // button height without label
)

// resolvedWidth returns the effective strip width.
func (w *Widget) resolvedWidth() float32 {
	if w.cfg.width > 0 {
		return w.cfg.width
	}
	if w.cfg.showLabels {
		return defaultWidthLabels
	}
	return defaultWidthIconsOnly
}

// resolvedButtonHeight returns the effective button height.
func (w *Widget) resolvedButtonHeight() float32 {
	if w.cfg.showLabels {
		return buttonHeightLabels
	}
	return buttonHeightIconsOnly
}

// --- widget.Widget interface ---

// Layout calculates the stripe's preferred size within the given constraints.
// The stripe takes its configured width and fills available height.
func (w *Widget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	stripW := w.resolvedWidth()
	preferred := geometry.Sz(stripW, constraints.MaxHeight)
	size := constraints.Constrain(preferred)

	w.layoutButtons(size)
	w.SetBounds(geometry.FromPointSize(w.Position(), size))
	return size
}

// layoutButtons calculates bounds for all buttons in local coordinates.
func (w *Widget) layoutButtons(available geometry.Size) {
	btnH := w.resolvedButtonHeight()
	stripW := available.Width

	// Top group: stacked from y=0 downward.
	var y float32
	for i := range w.cfg.topItems {
		w.topStates[i].bounds = geometry.NewRect(0, y, stripW, btnH)
		y += btnH
	}

	// Bottom group: stacked from bottom upward.
	y = available.Height
	for i := len(w.cfg.bottomItems) - 1; i >= 0; i-- {
		y -= btnH
		w.bottomStates[i].bounds = geometry.NewRect(0, y, stripW, btnH)
	}
}

// Draw renders the stripe background and all buttons.
func (w *Widget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !w.IsVisible() {
		return
	}

	bounds := w.Bounds()

	// Draw background at screen position.
	w.painter.PaintBackground(canvas, bounds)

	// Draw buttons in local coordinates.
	canvas.PushTransform(bounds.Min)

	for i := range w.cfg.topItems {
		w.painter.PaintButton(canvas, ButtonPaintState{
			Bounds:    w.topStates[i].bounds,
			Icon:      w.cfg.topItems[i].Icon,
			Label:     w.cfg.topItems[i].Label,
			Active:    w.cfg.topItems[i].ID == w.activeID,
			Hovered:   w.topStates[i].interaction == stateHover,
			Pressed:   w.topStates[i].interaction == statePressed,
			ShowLabel: w.cfg.showLabels,
		})
	}

	for i := range w.cfg.bottomItems {
		w.painter.PaintButton(canvas, ButtonPaintState{
			Bounds:    w.bottomStates[i].bounds,
			Icon:      w.cfg.bottomItems[i].Icon,
			Label:     w.cfg.bottomItems[i].Label,
			Active:    w.cfg.bottomItems[i].ID == w.activeID,
			Hovered:   w.bottomStates[i].interaction == stateHover,
			Pressed:   w.bottomStates[i].interaction == statePressed,
			ShowLabel: w.cfg.showLabels,
		})
	}

	canvas.PopTransform()
}

// Event handles input events for the stripe.
func (w *Widget) Event(ctx widget.Context, e event.Event) bool {
	if !w.IsVisible() || !w.IsEnabled() {
		return false
	}

	me, ok := e.(*event.MouseEvent)
	if !ok {
		return false
	}
	return w.handleMouseEvent(ctx, me)
}

// handleMouseEvent processes mouse events for stripe buttons.
func (w *Widget) handleMouseEvent(ctx widget.Context, e *event.MouseEvent) bool {
	local := e.Position.Sub(w.Bounds().Min)

	switch e.MouseType {
	case event.MouseMove:
		return w.handleMove(ctx, local)
	case event.MouseLeave:
		return w.clearAllHover(ctx)
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
	default:
		return false
	}
}

// handleMove updates hover states based on cursor position.
func (w *Widget) handleMove(ctx widget.Context, local geometry.Point) bool {
	group, idx := w.hitTest(local)
	changed := false

	// Clear previous hover states.
	for i := range w.topStates {
		if w.topStates[i].interaction == stateHover {
			w.topStates[i].interaction = stateNormal
			changed = true
		}
	}
	for i := range w.bottomStates {
		if w.bottomStates[i].interaction == stateHover {
			w.bottomStates[i].interaction = stateNormal
			changed = true
		}
	}

	// Set new hover.
	if idx >= 0 {
		states := w.statesForGroup(group)
		if states[idx].interaction == stateNormal {
			states[idx].interaction = stateHover
			changed = true
		}
	}

	if changed {
		ctx.Invalidate()
	}
	return changed
}

// handlePress sets the pressed state on the hit button.
func (w *Widget) handlePress(ctx widget.Context, local geometry.Point) bool {
	group, idx := w.hitTest(local)
	if idx < 0 {
		return false
	}

	states := w.statesForGroup(group)
	states[idx].interaction = statePressed
	ctx.Invalidate()
	return true
}

// handleRelease fires the click handler if released on the same button.
func (w *Widget) handleRelease(ctx widget.Context, local geometry.Point) bool {
	// Find which button was pressed.
	pressedGroup, pressedIdx := w.findPressed()
	if pressedIdx < 0 {
		return false
	}

	// Clear pressed state.
	pressedStates := w.statesForGroup(pressedGroup)
	pressedStates[pressedIdx].interaction = stateNormal

	// Check if released on the same button.
	releaseGroup, releaseIdx := w.hitTest(local)
	if releaseGroup == pressedGroup && releaseIdx == pressedIdx {
		pressedStates[pressedIdx].interaction = stateHover

		btn := w.buttonForGroup(pressedGroup, pressedIdx)
		w.activeID = btn.ID
		if btn.OnClick != nil {
			btn.OnClick()
		}
		if w.cfg.onSelect != nil {
			w.cfg.onSelect(btn.ID)
		}
	}

	ctx.Invalidate()
	return true
}

// clearAllHover resets all hover/pressed states to normal.
func (w *Widget) clearAllHover(ctx widget.Context) bool {
	changed := false
	for i := range w.topStates {
		if w.topStates[i].interaction != stateNormal {
			w.topStates[i].interaction = stateNormal
			changed = true
		}
	}
	for i := range w.bottomStates {
		if w.bottomStates[i].interaction != stateNormal {
			w.bottomStates[i].interaction = stateNormal
			changed = true
		}
	}
	if changed {
		ctx.Invalidate()
	}
	return changed
}

// hitTest returns which group (false=top, true=bottom) and index the point hits.
// Returns -1 for index if no button is hit.
func (w *Widget) hitTest(local geometry.Point) (bottom bool, idx int) {
	for i := range w.topStates {
		if w.topStates[i].bounds.Contains(local) {
			return false, i
		}
	}
	for i := range w.bottomStates {
		if w.bottomStates[i].bounds.Contains(local) {
			return true, i
		}
	}
	return false, noHover
}

// findPressed returns the group and index of the currently pressed button.
func (w *Widget) findPressed() (bottom bool, idx int) {
	for i := range w.topStates {
		if w.topStates[i].interaction == statePressed {
			return false, i
		}
	}
	for i := range w.bottomStates {
		if w.bottomStates[i].interaction == statePressed {
			return true, i
		}
	}
	return false, noHover
}

// statesForGroup returns the button states slice for the given group.
func (w *Widget) statesForGroup(bottom bool) []buttonState {
	if bottom {
		return w.bottomStates
	}
	return w.topStates
}

// buttonForGroup returns the Button at the given index in the given group.
func (w *Widget) buttonForGroup(bottom bool, idx int) Button {
	if bottom {
		return w.cfg.bottomItems[idx]
	}
	return w.cfg.topItems[idx]
}

// Children returns nil. Stripe buttons are not child widgets.
func (w *Widget) Children() []widget.Widget {
	return nil
}

// --- a11y.Accessible interface ---

// AccessibilityRole returns [a11y.RoleToolbar].
func (w *Widget) AccessibilityRole() a11y.Role {
	return a11y.RoleToolbar
}

// AccessibilityLabel returns "Tool Window Strip".
func (w *Widget) AccessibilityLabel() string {
	return a11yLabel
}

// AccessibilityHint returns an empty string.
func (w *Widget) AccessibilityHint() string {
	return ""
}

// AccessibilityValue returns the active button ID.
func (w *Widget) AccessibilityValue() string {
	return w.activeID
}

// AccessibilityState returns the current accessibility state.
func (w *Widget) AccessibilityState() a11y.State {
	return a11y.State{
		Disabled: !w.IsEnabled(),
		Hidden:   !w.IsVisible(),
	}
}

// AccessibilityActions returns nil. Actions are on individual buttons.
func (w *Widget) AccessibilityActions() []a11y.Action {
	return nil
}

// a11yLabel is the accessibility label for the stripe.
const a11yLabel = "Tool Window Strip"

// Compile-time interface checks.
var (
	_ widget.Widget   = (*Widget)(nil)
	_ a11y.Accessible = (*Widget)(nil)
)
