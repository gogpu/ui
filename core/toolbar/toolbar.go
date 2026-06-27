package toolbar

import (
	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// Option configures a toolbar during construction.
type Option func(*config)

// config holds the toolbar's configuration, set at construction time via options.
type config struct {
	items      []Item
	height     float32
	buttonSize float32 // 0 = use default (36px)
	gap        float32 // 0 = use default (2px); -1 = explicitly 0
	painter    Painter
}

// Items sets the toolbar's items.
func Items(items ...Item) Option {
	return func(c *config) {
		c.items = items
	}
}

// Height sets the toolbar height in logical pixels.
func Height(h float32) Option {
	return func(c *config) {
		c.height = h
	}
}

// ButtonSize sets the icon button size in logical pixels.
// Default is 36px. JetBrains IDE uses 30px for main toolbar.
func ButtonSize(px float32) Option {
	return func(c *config) { c.buttonSize = px }
}

// Gap sets the spacing between toolbar items in logical pixels.
// Default is 2px. JetBrains IDE uses 10px for main toolbar.
func Gap(px float32) Option {
	return func(c *config) { c.gap = px }
}

// PainterOpt sets the painter used to render the toolbar.
func PainterOpt(p Painter) Option {
	return func(c *config) {
		c.painter = p
	}
}

// interactionState represents the current user interaction state for a button item.
type interactionState uint8

const (
	stateNormal  interactionState = iota
	stateHover                    // mouse is over the item
	statePressed                  // mouse button is held down
)

// itemState tracks per-item interaction and layout data.
type itemState struct {
	interaction interactionState
	bounds      geometry.Rect
}

// Widget implements a horizontal toolbar with icon buttons, separators,
// spacers, and custom widget items.
//
// A toolbar is created with [New] using functional options:
//
//	tb := toolbar.New(
//	    toolbar.Items(
//	        toolbar.IconButton("New", icon.Add, onNew),
//	        toolbar.Separator(),
//	        toolbar.Spacer(),
//	        toolbar.IconButton("Settings", icon.Settings, onSettings),
//	    ),
//	    toolbar.Height(40),
//	)
type Widget struct {
	widget.WidgetBase

	cfg        config
	painter    Painter
	itemStates []itemState
	focusIndex int // index of the focused item (-1 = none)
}

// New creates a new toolbar Widget with the given options.
//
// The returned widget is visible and enabled by default. The default
// height is 40 logical pixels.
func New(opts ...Option) *Widget {
	w := &Widget{
		painter:    DefaultPainter{},
		focusIndex: noFocusIndex,
	}
	w.SetVisible(true)
	w.SetEnabled(true)

	w.cfg.height = defaultHeight

	for _, opt := range opts {
		opt(&w.cfg)
	}

	if w.cfg.painter != nil {
		w.painter = w.cfg.painter
	}

	w.itemStates = make([]itemState, len(w.cfg.items))

	// ADR-028: parent chain for upward dirty propagation.
	// Flutter: RenderObject.adoptChild sets parent on each child.
	for _, item := range w.cfg.items { //nolint:gocritic // Item is read-only here
		if item.Kind == ItemCustom && item.Widget != nil {
			type parentSetter interface{ SetParent(widget.Widget) }
			if ps, ok := item.Widget.(parentSetter); ok {
				ps.SetParent(w)
			}
		}
	}

	return w
}

// Default configuration values.
const (
	defaultHeight  float32 = 40
	buttonItemSize float32 = 36
	separatorWidth float32 = 8
	itemGap        float32 = 2
	noFocusIndex           = -1
)

// resolvedButtonSize returns the effective button size.
func (w *Widget) resolvedButtonSize() float32 {
	if w.cfg.buttonSize > 0 {
		return w.cfg.buttonSize
	}
	return buttonItemSize
}

// resolvedGap returns the effective item gap.
func (w *Widget) resolvedGap() float32 {
	if w.cfg.gap < 0 {
		return 0
	}
	if w.cfg.gap > 0 {
		return w.cfg.gap
	}
	return itemGap
}

// Text-with-icon button width constants.
const (
	textButtonMinWidth float32 = 64
	charWidthRatio     float32 = 0.55
)

// ItemCount returns the number of items in the toolbar.
func (w *Widget) ItemCount() int {
	return len(w.cfg.items)
}

// ItemAt returns the item at the given index, or an empty Item if out of range.
func (w *Widget) ItemAt(idx int) Item {
	if idx < 0 || idx >= len(w.cfg.items) {
		return Item{}
	}
	return w.cfg.items[idx]
}

// --- widget.Widget interface ---

// Layout calculates the toolbar's preferred size within the given constraints.
func (w *Widget) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	preferred := geometry.Sz(constraints.MaxWidth, w.cfg.height)
	size := constraints.Constrain(preferred)

	// Calculate item bounds.
	w.layoutItems(ctx, size)

	w.SetBounds(geometry.FromPointSize(w.Position(), size))
	return size
}

// layoutItems calculates bounds for each item within the toolbar.
func (w *Widget) layoutItems(ctx widget.Context, available geometry.Size) {
	if len(w.cfg.items) == 0 {
		return
	}

	// First pass: calculate fixed-width items and count spacers.
	var fixedWidth float32
	var spacerCount int
	var gapCount int

	for i, item := range w.cfg.items { //nolint:gocritic // Item is read-only here
		switch item.Kind {
		case ItemButton:
			fixedWidth += w.buttonItemWidth(item)
		case ItemSeparator:
			fixedWidth += separatorWidth
		case ItemSpacer:
			spacerCount++
		case ItemCustom:
			if item.Widget != nil {
				childConstraints := geometry.Loose(geometry.Sz(available.Width, available.Height))
				sz := item.Widget.Layout(ctx, childConstraints)
				fixedWidth += sz.Width
			}
		}
		if i > 0 {
			gapCount++
		}
	}
	gap := w.resolvedGap()
	fixedWidth += float32(gapCount) * gap

	// Calculate spacer width.
	var spacerW float32
	if spacerCount > 0 {
		remaining := available.Width - fixedWidth
		if remaining < 0 {
			remaining = 0
		}
		spacerW = remaining / float32(spacerCount)
	}

	// Second pass: assign bounds.
	var x float32
	for i, item := range w.cfg.items { //nolint:gocritic // Item is read-only here
		if i > 0 {
			x += gap
		}

		var itemW float32
		switch item.Kind {
		case ItemButton:
			itemW = w.buttonItemWidth(item)
		case ItemSeparator:
			itemW = separatorWidth
		case ItemSpacer:
			itemW = spacerW
		case ItemCustom:
			if item.Widget != nil {
				itemW = item.Widget.(interface{ Bounds() geometry.Rect }).Bounds().Width()
			}
		}

		bounds := geometry.NewRect(x, 0, itemW, available.Height)
		w.itemStates[i].bounds = bounds

		// Position custom widgets.
		if item.Kind == ItemCustom && item.Widget != nil {
			childH := item.Widget.(interface{ Bounds() geometry.Rect }).Bounds().Height()
			cy := (available.Height - childH) / 2
			item.Widget.(interface{ SetBounds(geometry.Rect) }).SetBounds(
				geometry.NewRect(x, cy, itemW, childH),
			)
		}

		x += itemW
	}
}

// buttonItemWidth calculates the width for a button item.
func (w *Widget) buttonItemWidth(item Item) float32 {
	if item.ShowLabel && item.Label != "" {
		textW := float32(len(item.Label)) * defaultFontSize * charWidthRatio
		width := iconPadding + maxIconSize + textIconGap + textW + iconPadding
		if width < textButtonMinWidth {
			width = textButtonMinWidth
		}
		return width
	}
	return w.resolvedButtonSize()
}

// Draw renders the toolbar background and all items.
func (w *Widget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !w.IsVisible() {
		return
	}

	bounds := w.Bounds()

	// Draw toolbar background.
	w.painter.PaintToolbar(canvas, PaintToolbarState{
		Bounds: bounds,
	})

	// Draw items with transform offset.
	canvas.PushTransform(bounds.Min)
	for i, item := range w.cfg.items { //nolint:gocritic
		itemBounds := w.itemStates[i].bounds
		if itemBounds.IsEmpty() {
			continue
		}

		switch item.Kind {
		case ItemButton:
			// Pre-compute icon/text bounds (ADR-034 Phase 4).
			iBounds := iconBoundsForItem(itemBounds, item.ShowLabel)
			var tBounds geometry.Rect
			if item.ShowLabel && item.Label != "" {
				tBounds = textBoundsForItem(itemBounds, iBounds)
			}
			w.painter.PaintButtonItem(canvas, PaintButtonState{
				Label:      item.Label,
				Icon:       item.Icon,
				ShowLabel:  item.ShowLabel,
				Hovered:    w.itemStates[i].interaction == stateHover,
				Pressed:    w.itemStates[i].interaction == statePressed,
				Focused:    w.focusIndex == i,
				Disabled:   !item.Enabled,
				Bounds:     itemBounds,
				IconBounds: iBounds,
				TextBounds: tBounds,
			})
		case ItemSeparator:
			w.painter.PaintSeparator(canvas, itemBounds)
		case ItemSpacer:
			// Spacers are invisible.
		case ItemCustom:
			if item.Widget != nil {
				widget.StampScreenOrigin(item.Widget, canvas)
				item.Widget.Draw(ctx, canvas)
			}
		}
	}
	canvas.PopTransform()
}

// Event handles input events for the toolbar and its items.
func (w *Widget) Event(ctx widget.Context, e event.Event) bool {
	if !w.IsVisible() || !w.IsEnabled() {
		return false
	}

	switch ev := e.(type) {
	case *event.MouseEvent:
		return w.handleMouseEvent(ctx, ev)
	case *event.KeyEvent:
		return w.handleKeyEvent(ctx, ev)
	default:
		// Dispatch to custom widget items.
		return w.dispatchToCustomItems(ctx, e)
	}
}

// handleMouseEvent processes mouse events for toolbar items.
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
		return w.clearHoverStates(ctx)

	default:
		return false
	}
}

// handlePress finds the item under the cursor and sets it to pressed state.
func (w *Widget) handlePress(ctx widget.Context, local geometry.Point) bool {
	idx := w.hitTest(local)
	if idx < 0 {
		return false
	}

	item := w.cfg.items[idx]
	if item.Kind != ItemButton || !item.Enabled {
		// Dispatch to custom widget if applicable.
		if item.Kind == ItemCustom && item.Widget != nil {
			return item.Widget.Event(ctx, &event.MouseEvent{
				MouseType: event.MousePress,
				Button:    event.ButtonLeft,
				Position:  local,
			})
		}
		return false
	}

	w.itemStates[idx].interaction = statePressed
	w.focusIndex = idx
	w.SetNeedsRedraw(true)
	ctx.InvalidateRect(w.Bounds())
	return true
}

// handleRelease finds the item under the cursor and fires its click handler.
func (w *Widget) handleRelease(ctx widget.Context, local geometry.Point) bool {
	// Find which item was pressed.
	pressedIdx := noFocusIndex
	for i := range w.itemStates {
		if w.itemStates[i].interaction == statePressed {
			pressedIdx = i
			break
		}
	}

	if pressedIdx < 0 {
		return false
	}

	wasPressed := pressedIdx
	// Update interaction state based on cursor position.
	idx := w.hitTest(local)
	if idx == wasPressed {
		w.itemStates[wasPressed].interaction = stateHover
	} else {
		w.itemStates[wasPressed].interaction = stateNormal
	}
	w.SetNeedsRedraw(true)
	ctx.InvalidateRect(w.Bounds())

	// Fire click if released on the same item that was pressed.
	if idx == wasPressed {
		item := w.cfg.items[wasPressed]
		if item.Kind == ItemButton && item.Enabled && item.OnClick != nil {
			item.OnClick()
		}
	}
	return true
}

// handleMove updates hover states based on cursor position.
func (w *Widget) handleMove(ctx widget.Context, local geometry.Point) bool {
	idx := w.hitTest(local)
	changed := false

	for i := range w.itemStates {
		item := w.cfg.items[i]
		if item.Kind != ItemButton || !item.Enabled {
			continue
		}

		newState := resolveHoverState(w.itemStates[i].interaction, i == idx)
		if w.itemStates[i].interaction != newState {
			w.itemStates[i].interaction = newState
			changed = true
		}
	}

	if changed {
		w.SetNeedsRedraw(true)
		ctx.InvalidateRect(w.Bounds())
	}
	return changed
}

// resolveHoverState determines the new interaction state for an item during
// mouse move. Pressed items retain their state regardless of cursor position.
func resolveHoverState(current interactionState, underCursor bool) interactionState {
	if current == statePressed {
		return statePressed
	}
	if underCursor {
		return stateHover
	}
	return stateNormal
}

// clearHoverStates resets all item interaction states to normal.
func (w *Widget) clearHoverStates(ctx widget.Context) bool {
	changed := false
	for i := range w.itemStates {
		if w.itemStates[i].interaction != stateNormal {
			w.itemStates[i].interaction = stateNormal
			changed = true
		}
	}
	if changed {
		w.SetNeedsRedraw(true)
		ctx.InvalidateRect(w.Bounds())
	}
	return changed
}

// handleKeyEvent processes keyboard events for the toolbar.
func (w *Widget) handleKeyEvent(ctx widget.Context, e *event.KeyEvent) bool {
	if !w.IsFocused() {
		return false
	}

	switch e.Key {
	case event.KeyTab:
		return w.handleTabKey(ctx, e)
	case event.KeyLeft:
		return w.handleArrowKey(ctx, e, -1)
	case event.KeyRight:
		return w.handleArrowKey(ctx, e, 1)
	case event.KeyEnter, event.KeySpace:
		return w.handleActivationKey(ctx, e)
	default:
		return false
	}
}

// handleTabKey moves focus between toolbar items.
func (w *Widget) handleTabKey(ctx widget.Context, e *event.KeyEvent) bool {
	if e.KeyType != event.KeyPress {
		return false
	}

	direction := 1
	if e.Modifiers().IsShift() {
		direction = -1
	}
	return w.moveFocus(ctx, direction)
}

// handleArrowKey moves focus left or right between items.
func (w *Widget) handleArrowKey(ctx widget.Context, e *event.KeyEvent, direction int) bool {
	if e.KeyType != event.KeyPress {
		return false
	}
	return w.moveFocus(ctx, direction)
}

// moveFocus moves focus to the next or previous focusable item.
func (w *Widget) moveFocus(ctx widget.Context, direction int) bool {
	if len(w.cfg.items) == 0 {
		return false
	}

	start := w.focusIndex
	if start < 0 {
		if direction > 0 {
			start = -1
		} else {
			start = len(w.cfg.items)
		}
	}

	idx := start + direction
	for idx >= 0 && idx < len(w.cfg.items) {
		item := w.cfg.items[idx]
		if item.Kind == ItemButton && item.Enabled {
			w.focusIndex = idx
			w.SetNeedsRedraw(true)
			ctx.InvalidateRect(w.Bounds())
			return true
		}
		idx += direction
	}
	return false
}

// handleActivationKey activates the currently focused button item.
func (w *Widget) handleActivationKey(_ widget.Context, e *event.KeyEvent) bool {
	if w.focusIndex < 0 || w.focusIndex >= len(w.cfg.items) {
		return false
	}

	item := w.cfg.items[w.focusIndex]
	if item.Kind != ItemButton || !item.Enabled {
		return false
	}

	switch e.KeyType {
	case event.KeyPress:
		w.itemStates[w.focusIndex].interaction = statePressed
		return true
	case event.KeyRelease:
		wasPressed := w.itemStates[w.focusIndex].interaction == statePressed
		w.itemStates[w.focusIndex].interaction = stateNormal
		if wasPressed && item.OnClick != nil {
			item.OnClick()
		}
		return true
	default:
		return false
	}
}

// hitTest returns the index of the item at the given local point, or -1.
func (w *Widget) hitTest(local geometry.Point) int {
	for i := range w.itemStates {
		if w.itemStates[i].bounds.Contains(local) {
			return i
		}
	}
	return noFocusIndex
}

// HitTestPoint returns true if the local-space point hits a toolbar item
// (button, separator, or custom widget). Returns false for empty gaps
// between items and spacers — allowing the parent to treat gaps as drag area.
func (w *Widget) HitTestPoint(local geometry.Point) bool {
	for i, item := range w.cfg.items { //nolint:gocritic
		if item.Kind == ItemSpacer {
			continue // spacers are not interactive
		}
		if w.itemStates[i].bounds.Contains(local) {
			return true
		}
	}
	return false
}

// dispatchToCustomItems dispatches events to custom widget items.
func (w *Widget) dispatchToCustomItems(ctx widget.Context, e event.Event) bool {
	for _, item := range w.cfg.items { //nolint:gocritic
		if item.Kind == ItemCustom && item.Widget != nil {
			if item.Widget.Event(ctx, e) {
				return true
			}
		}
	}
	return false
}

// Children returns the custom widget items embedded in the toolbar.
func (w *Widget) Children() []widget.Widget {
	var children []widget.Widget
	for _, item := range w.cfg.items { //nolint:gocritic
		if item.Kind == ItemCustom && item.Widget != nil {
			children = append(children, item.Widget)
		}
	}
	return children
}

// IsFocusable reports whether the toolbar can receive focus.
func (w *Widget) IsFocusable() bool {
	return w.IsVisible() && w.IsEnabled()
}

// --- a11y.Accessible interface ---

// AccessibilityRole returns [a11y.RoleToolbar].
func (w *Widget) AccessibilityRole() a11y.Role {
	return a11y.RoleToolbar
}

// AccessibilityLabel returns "Toolbar".
func (w *Widget) AccessibilityLabel() string {
	return a11yLabel
}

// AccessibilityHint returns an empty string.
func (w *Widget) AccessibilityHint() string {
	return ""
}

// AccessibilityValue returns an empty string.
func (w *Widget) AccessibilityValue() string {
	return ""
}

// AccessibilityState returns the default state.
func (w *Widget) AccessibilityState() a11y.State {
	return a11y.State{
		Disabled: !w.IsEnabled(),
		Hidden:   !w.IsVisible(),
	}
}

// AccessibilityActions returns nil. Toolbar actions are on individual items.
func (w *Widget) AccessibilityActions() []a11y.Action {
	return nil
}

// a11yLabel is the accessibility label for the toolbar.
const a11yLabel = "Toolbar"

// Compile-time interface checks.
var (
	_ widget.Widget    = (*Widget)(nil)
	_ widget.Focusable = (*Widget)(nil)
	_ a11y.Accessible  = (*Widget)(nil)
)
