package splitview

import (
	"time"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// Orientation controls how the two panels are arranged.
type Orientation uint8

// Orientation constants.
const (
	// Horizontal arranges panels left (first) and right (second).
	Horizontal Orientation = iota

	// Vertical arranges panels top (first) and bottom (second).
	Vertical
)

// orientationNames maps each Orientation to its human-readable name.
var orientationNames = [...]string{
	Horizontal: orientationHorizontalStr,
	Vertical:   orientationVerticalStr,
}

// String constants for Orientation.String to satisfy goconst.
const (
	orientationHorizontalStr = "Horizontal"
	orientationVerticalStr   = "Vertical"
	orientationUnknownStr    = "Unknown"
)

// String returns a human-readable name for the orientation.
func (o Orientation) String() string {
	if int(o) < len(orientationNames) {
		return orientationNames[o]
	}
	return orientationUnknownStr
}

// Option configures a split view during construction.
type Option func(*config)

// config holds the split view's configuration, set at construction time via options.
type config struct {
	first        widget.Widget
	second       widget.Widget
	orientation  Orientation
	ratio        float32
	minFirst     float32
	minSecond    float32
	dividerWidth float32
	collapsible  bool

	fixedFirst float32 // pixel size for first panel (0 = use ratio)

	ratioSignal         state.Signal[float32]
	readonlyRatioSignal state.ReadonlySignal[float32]

	onRatioChange func(float32)
	painter       Painter
	colorScheme   DividerColorScheme
}

// ResolvedRatio returns the current split ratio.
// Priority: ReadonlySignal > Signal > Static.
func (c *config) ResolvedRatio() float32 {
	if c.readonlyRatioSignal != nil {
		return c.readonlyRatioSignal.Get()
	}
	if c.ratioSignal != nil {
		return c.ratioSignal.Get()
	}
	return c.ratio
}

// resolvedDividerWidth returns the configured divider width or the default.
func (c *config) resolvedDividerWidth() float32 {
	if c.dividerWidth > 0 {
		return c.dividerWidth
	}
	return defaultDividerWidth
}

// First sets the first panel widget (left for Horizontal, top for Vertical).
func First(w widget.Widget) Option {
	return func(c *config) { c.first = w }
}

// Second sets the second panel widget (right for Horizontal, bottom for Vertical).
func Second(w widget.Widget) Option {
	return func(c *config) { c.second = w }
}

// OrientationOpt sets the split orientation.
// Default is [Horizontal].
func OrientationOpt(o Orientation) Option {
	return func(c *config) { c.orientation = o }
}

// InitialRatio sets the initial split ratio (0.0 to 1.0).
// A ratio of 0.3 means the first panel takes 30% of the available space.
// Default is 0.5.
func InitialRatio(r float32) Option {
	return func(c *config) { c.ratio = clampRatio(r) }
}

// FixedFirst sets a fixed pixel size for the first panel.
// Unlike ratio-based sizing, this keeps the first panel at a constant
// pixel width (horizontal) or height (vertical) regardless of window size.
// The second panel fills the remaining space.
// Set to 0 (default) to use ratio-based sizing.
func FixedFirst(px float32) Option {
	return func(c *config) { c.fixedFirst = px }
}

// MinFirst sets the minimum size (width or height) of the first panel in pixels.
func MinFirst(px float32) Option {
	return func(c *config) { c.minFirst = px }
}

// MinSecond sets the minimum size (width or height) of the second panel in pixels.
func MinSecond(px float32) Option {
	return func(c *config) { c.minSecond = px }
}

// DividerWidth sets the divider thickness in pixels.
// Default is 6 pixels.
func DividerWidth(w float32) Option {
	return func(c *config) { c.dividerWidth = w }
}

// CollapsibleOpt enables double-click-to-collapse on the divider.
// When enabled, double-clicking the divider collapses the first panel.
// Double-clicking again restores it to the previous ratio.
func CollapsibleOpt(v bool) Option {
	return func(c *config) { c.collapsible = v }
}

// RatioSignal binds the split ratio to a reactive signal.
// This is a TWO-WAY binding: the widget reads the ratio from the signal,
// and when the user drags the divider, the new ratio is written back.
func RatioSignal(sig state.Signal[float32]) Option {
	return func(c *config) { c.ratioSignal = sig }
}

// RatioReadonlySignal binds the split ratio to a read-only signal.
// When set, this takes highest precedence over all other ratio sources.
func RatioReadonlySignal(sig state.ReadonlySignal[float32]) Option {
	return func(c *config) { c.readonlyRatioSignal = sig }
}

// OnRatioChange sets the callback invoked when the split ratio changes.
func OnRatioChange(fn func(float32)) Option {
	return func(c *config) { c.onRatioChange = fn }
}

// PainterOpt sets the painter used to render the divider.
// Each design system provides its own painter. If not set,
// [DefaultPainter] is used.
func PainterOpt(p Painter) Option {
	return func(c *config) { c.painter = p }
}

// ColorSchemeOpt sets the theme-derived color scheme for the divider.
func ColorSchemeOpt(cs DividerColorScheme) Option {
	return func(c *config) { c.colorScheme = cs }
}

// Default values.
const (
	defaultRatio        float32 = 0.5
	defaultDividerWidth float32 = 6
	defaultMinPanel     float32 = 0
)

// doubleClickThreshold is the maximum time between two clicks to count as a double-click.
const doubleClickThreshold = 400 * time.Millisecond

// Widget implements a resizable split panel container with a draggable divider.
//
// A split view is created with [New] using functional options:
//
//	split := splitview.New(
//	    splitview.First(leftPanel),
//	    splitview.Second(rightPanel),
//	    splitview.OrientationOpt(splitview.Horizontal),
//	    splitview.InitialRatio(0.3),
//	)
type Widget struct {
	widget.WidgetBase
	cfg     config
	painter Painter

	// Interaction state.
	hovered        bool
	dragging       bool
	dragStart      geometry.Point
	dragStartRatio float32

	// Collapse state.
	collapsed    bool
	preCollapse  float32 // ratio before collapse, for restore
	lastClickAt  time.Time
	lastClickPos geometry.Point
}

// New creates a new split view Widget with the given options.
//
// The returned widget is visible and enabled by default.
// The default orientation is [Horizontal] with a 50/50 split.
func New(opts ...Option) *Widget {
	w := &Widget{
		painter: DefaultPainter{},
	}
	w.SetVisible(true)
	w.SetEnabled(true)

	w.cfg.ratio = defaultRatio

	for _, opt := range opts {
		opt(&w.cfg)
	}

	if w.cfg.painter != nil {
		w.painter = w.cfg.painter
	}

	return w
}

// Layout calculates panel sizes and positions children.
func (w *Widget) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	size := constraints.Biggest()
	if size.Width <= 0 || size.Height <= 0 {
		fallback := geometry.Sz(defaultViewportWidth, defaultViewportHeight)
		if constraints.MaxWidth > 0 {
			fallback.Width = constraints.MaxWidth
		}
		if constraints.MaxHeight > 0 {
			fallback.Height = constraints.MaxHeight
		}
		size = fallback
	}

	divW := w.cfg.resolvedDividerWidth()

	// If fixedFirst is set, recompute ratio from pixel size each layout pass.
	// This keeps the first panel at a constant pixel size regardless of window size.
	if w.cfg.fixedFirst > 0 {
		totalSpace := size.Width
		if w.cfg.orientation == Vertical {
			totalSpace = size.Height
		}
		totalSpace -= divW
		if totalSpace > 0 {
			w.cfg.ratio = clampRatio(w.cfg.fixedFirst / totalSpace)
		}
	}

	ratio := w.effectiveRatio()

	// Use local origin (0,0) for child placement. The parent positions us
	// via SetBounds after Layout returns, and Draw applies PushTransform
	// to map local coordinates to parent space. Using w.Bounds().Min here
	// is incorrect because Bounds is stale during Layout (parent calls
	// Layout first, then SetBounds).
	origin := geometry.Point{}
	firstRect, secondRect := computePanelRects(size, origin, ratio, divW, w.cfg.orientation)

	w.layoutChild(ctx, w.cfg.first, firstRect)
	w.layoutChild(ctx, w.cfg.second, secondRect)

	return size
}

// Default viewport dimensions used as fallback.
const (
	defaultViewportWidth  float32 = 400
	defaultViewportHeight float32 = 300
)

// layoutChild measures and positions a single child widget.
func (w *Widget) layoutChild(ctx widget.Context, child widget.Widget, rect geometry.Rect) {
	if child == nil {
		return
	}
	childConstraints := geometry.Constraints{
		MinWidth:  rect.Width(),
		MaxWidth:  rect.Width(),
		MinHeight: rect.Height(),
		MaxHeight: rect.Height(),
	}
	child.Layout(ctx, childConstraints)
	if setter, ok := child.(interface{ SetBounds(geometry.Rect) }); ok {
		setter.SetBounds(rect)
	}
}

// Draw renders both panels and the divider.
func (w *Widget) Draw(ctx widget.Context, canvas widget.Canvas) {
	bounds := w.Bounds()
	if bounds.IsEmpty() {
		return
	}

	// Push transform for local coordinate space (children use local bounds).
	canvas.PushTransform(bounds.Min)

	// Draw first panel.
	if w.cfg.first != nil {
		widget.StampScreenOrigin(w.cfg.first, canvas)
		w.cfg.first.Draw(ctx, canvas)
	}

	// Draw divider.
	divRect := w.dividerRect()
	ps := PaintState{
		DividerRect: divRect,
		Orientation: w.cfg.orientation,
		Hovered:     w.hovered,
		Dragging:    w.dragging,
		Collapsed:   w.collapsed,
		ColorScheme: w.cfg.colorScheme,
	}
	w.painter.PaintDivider(canvas, ps)

	// Draw second panel.
	if w.cfg.second != nil {
		widget.StampScreenOrigin(w.cfg.second, canvas)
		w.cfg.second.Draw(ctx, canvas)
	}

	canvas.PopTransform()
}

// Event handles input events for divider dragging.
//
// Mouse and wheel event positions are translated from parent-local space
// to SplitView-local space before hit-testing and child dispatch, matching
// the coordinate convention used by [primitives.BoxWidget].
func (w *Widget) Event(ctx widget.Context, e event.Event) bool {
	// Translate mouse events to local coordinates.
	if me, ok := e.(*event.MouseEvent); ok {
		local := *me
		local.Position = me.Position.Sub(w.Bounds().Min)
		return w.handleMouseEvent(ctx, &local)
	}

	// Translate wheel events to local coordinates.
	if we, ok := e.(*event.WheelEvent); ok {
		local := *we
		local.Position = we.Position.Sub(w.Bounds().Min)
		return w.handleWheelEvent(ctx, &local)
	}

	// Non-positional events (keyboard, focus) broadcast to children.
	if w.cfg.first != nil {
		if w.cfg.first.Event(ctx, e) {
			return true
		}
	}
	if w.cfg.second != nil {
		if w.cfg.second.Event(ctx, e) {
			return true
		}
	}

	return false
}

// handleMouseEvent processes a mouse event in local coordinates.
func (w *Widget) handleMouseEvent(ctx widget.Context, me *event.MouseEvent) bool {
	// Check divider interactions first.
	if consumed := w.handleDividerEvent(ctx, me); consumed {
		return true
	}

	// Then dispatch to children.
	if w.cfg.first != nil {
		if w.cfg.first.Event(ctx, me) {
			return true
		}
	}
	if w.cfg.second != nil {
		if w.cfg.second.Event(ctx, me) {
			return true
		}
	}

	return false
}

// handleWheelEvent dispatches a wheel event (in local coordinates) to children.
func (w *Widget) handleWheelEvent(ctx widget.Context, we *event.WheelEvent) bool {
	if w.cfg.first != nil {
		if w.cfg.first.Event(ctx, we) {
			return true
		}
	}
	if w.cfg.second != nil {
		if w.cfg.second.Event(ctx, we) {
			return true
		}
	}
	return false
}

// handleDividerEvent processes mouse events related to the divider.
// The mouse event positions must be in SplitView-local coordinates.
func (w *Widget) handleDividerEvent(ctx widget.Context, me *event.MouseEvent) bool {
	switch me.MouseType {
	case event.MouseMove:
		return w.handleMouseMove(ctx, me)
	case event.MousePress:
		return w.handleMousePress(ctx, me)
	case event.MouseRelease:
		return w.handleMouseRelease(ctx, me)
	case event.MouseEnter, event.MouseLeave:
		// Update hover on enter/leave only if over divider.
		wasHovered := w.hovered
		w.hovered = w.dividerRect().Contains(me.Position)
		if w.hovered != wasHovered {
			w.updateCursor(ctx)
			ctx.Invalidate()
		}
		return false
	default:
		return false
	}
}

// handleMouseMove handles mouse movement for drag tracking and hover.
func (w *Widget) handleMouseMove(ctx widget.Context, me *event.MouseEvent) bool {
	if w.dragging {
		// If left button is no longer pressed, clear drag state.
		if !me.Buttons.IsLeftPressed() {
			w.dragging = false
			w.hovered = false
			ctx.SetCursor(widget.CursorDefault)
			ctx.Invalidate()
			return false
		}

		w.updateRatioFromDrag(ctx, me.Position)
		w.updateCursor(ctx) // Maintain drag cursor on every move
		return true
	}

	// Update hover state for cursor.
	wasHovered := w.hovered
	w.hovered = w.dividerRect().Contains(me.Position)
	if w.hovered != wasHovered {
		w.updateCursor(ctx)
		ctx.Invalidate()
	}

	return false
}

// handleMousePress handles mouse button press on the divider.
func (w *Widget) handleMousePress(ctx widget.Context, me *event.MouseEvent) bool {
	if me.Button != event.ButtonLeft {
		return false
	}

	divRect := w.dividerRect()
	if !divRect.Contains(me.Position) {
		return false
	}

	// Check for double-click (collapse/expand).
	if w.cfg.collapsible {
		now := ctx.Now()
		if now.Sub(w.lastClickAt) < doubleClickThreshold && w.isNearLastClick(me.Position) {
			w.toggleCollapse(ctx)
			w.lastClickAt = time.Time{} // Reset to prevent triple-click.
			return true
		}
		w.lastClickAt = now
		w.lastClickPos = me.Position
	}

	// Start drag.
	w.dragging = true
	w.dragStart = me.Position
	w.dragStartRatio = w.effectiveRatio()
	w.updateCursor(ctx)
	ctx.Invalidate()
	return true
}

// handleMouseRelease handles mouse button release.
func (w *Widget) handleMouseRelease(ctx widget.Context, me *event.MouseEvent) bool {
	if me.Button != event.ButtonLeft {
		return false
	}

	wasDragging := w.dragging
	w.dragging = false
	if wasDragging {
		w.hovered = w.dividerRect().Contains(me.Position)
		w.updateCursor(ctx)
		ctx.Invalidate()
	}
	return wasDragging
}

// isNearLastClick checks if a position is close to the last click position.
func (w *Widget) isNearLastClick(p geometry.Point) bool {
	const threshold float32 = 10
	dx := p.X - w.lastClickPos.X
	dy := p.Y - w.lastClickPos.Y
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx < threshold && dy < threshold
}

// toggleCollapse toggles the collapsed state of the first panel.
func (w *Widget) toggleCollapse(ctx widget.Context) {
	if w.collapsed {
		// Restore previous ratio.
		w.collapsed = false
		w.setRatio(ctx, w.preCollapse)
	} else {
		// Collapse first panel.
		w.preCollapse = w.effectiveRatio()
		w.collapsed = true
		w.setRatio(ctx, 0)
	}
}

// updateRatioFromDrag calculates the new ratio based on drag position.
func (w *Widget) updateRatioFromDrag(ctx widget.Context, pos geometry.Point) {
	bounds := w.Bounds()
	divW := w.cfg.resolvedDividerWidth()

	var totalSpace float32
	var delta float32

	if w.cfg.orientation == Horizontal {
		totalSpace = bounds.Width() - divW
		delta = pos.X - w.dragStart.X
	} else {
		totalSpace = bounds.Height() - divW
		delta = pos.Y - w.dragStart.Y
	}

	if totalSpace <= 0 {
		return
	}

	newRatio := w.dragStartRatio + delta/totalSpace
	newRatio = w.clampRatioToConstraints(newRatio, totalSpace)

	// Uncollapse if dragged away from zero.
	if w.collapsed && newRatio > 0 {
		w.collapsed = false
	}

	w.setRatio(ctx, newRatio)
}

// clampRatioToConstraints applies min panel constraints to the ratio.
func (w *Widget) clampRatioToConstraints(ratio, totalSpace float32) float32 {
	ratio = clampRatio(ratio)

	if totalSpace <= 0 {
		return ratio
	}

	// Apply minimum first panel constraint.
	if w.cfg.minFirst > 0 {
		minRatio := w.cfg.minFirst / totalSpace
		if ratio < minRatio {
			ratio = minRatio
		}
	}

	// Apply minimum second panel constraint.
	if w.cfg.minSecond > 0 {
		maxRatio := 1.0 - w.cfg.minSecond/totalSpace
		if ratio > maxRatio {
			ratio = maxRatio
		}
	}

	return ratio
}

// setRatio updates the split ratio, writes to signal if bound, and fires callback.
// If fixedFirst is active, the pixel size is updated from the new ratio.
func (w *Widget) setRatio(ctx widget.Context, ratio float32) {
	ratio = clampRatio(ratio)
	current := w.cfg.ResolvedRatio()

	if ratio == current {
		return
	}

	// Update fixedFirst pixel size so Layout preserves the dragged position.
	if w.cfg.fixedFirst > 0 {
		bounds := w.Bounds()
		totalSpace := bounds.Width()
		if w.cfg.orientation == Vertical {
			totalSpace = bounds.Height()
		}
		totalSpace -= w.cfg.resolvedDividerWidth()
		if totalSpace > 0 {
			w.cfg.fixedFirst = ratio * totalSpace
		}
	}

	// TWO-WAY: write back to signal if bound.
	if w.cfg.ratioSignal != nil {
		w.cfg.ratioSignal.Set(ratio)
	} else {
		w.cfg.ratio = ratio
	}

	if w.cfg.onRatioChange != nil {
		w.cfg.onRatioChange(ratio)
	}

	w.SetNeedsRedraw(true)
	ctx.Invalidate()
}

// effectiveRatio returns the current ratio, accounting for collapsed state.
func (w *Widget) effectiveRatio() float32 {
	if w.collapsed {
		return 0
	}
	return w.cfg.ResolvedRatio()
}

// dividerRect returns the current divider rectangle in local coordinates.
//
// Local coordinates start at (0,0) within the SplitView. The Draw method
// applies PushTransform(bounds.Min) to map these to parent space.
func (w *Widget) dividerRect() geometry.Rect {
	bounds := w.Bounds()
	if bounds.IsEmpty() {
		return geometry.Rect{}
	}

	divW := w.cfg.resolvedDividerWidth()
	ratio := w.effectiveRatio()

	if w.cfg.orientation == Horizontal {
		totalSpace := bounds.Width() - divW
		firstWidth := totalSpace * ratio
		return geometry.NewRect(
			firstWidth,
			0,
			divW,
			bounds.Height(),
		)
	}

	totalSpace := bounds.Height() - divW
	firstHeight := totalSpace * ratio
	return geometry.NewRect(
		0,
		firstHeight,
		bounds.Width(),
		divW,
	)
}

// updateCursor sets the appropriate resize cursor based on orientation and hover state.
func (w *Widget) updateCursor(ctx widget.Context) {
	if w.dragging || w.hovered {
		if w.cfg.orientation == Horizontal {
			ctx.SetCursor(widget.CursorResizeEW)
		} else {
			ctx.SetCursor(widget.CursorResizeNS)
		}
	} else {
		ctx.SetCursor(widget.CursorDefault)
	}
}

// Children returns the two panel widgets.
func (w *Widget) Children() []widget.Widget {
	var children []widget.Widget
	if w.cfg.first != nil {
		children = append(children, w.cfg.first)
	}
	if w.cfg.second != nil {
		children = append(children, w.cfg.second)
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
	if w.cfg.readonlyRatioSignal != nil {
		b := state.BindToScheduler(w.cfg.readonlyRatioSignal, w, sched)
		w.AddBinding(b)
	} else if w.cfg.ratioSignal != nil {
		b := state.BindToScheduler(w.cfg.ratioSignal, w, sched)
		w.AddBinding(b)
	}
}

// Unmount is called when the split view is removed from the widget tree.
// Implements [widget.Lifecycle].
func (w *Widget) Unmount() {
	// Bindings are cleaned up automatically by WidgetBase.CleanupBindings().
}

// Ratio returns the current split ratio.
func (w *Widget) Ratio() float32 {
	return w.cfg.ResolvedRatio()
}

// IsCollapsed reports whether the first panel is currently collapsed.
func (w *Widget) IsCollapsed() bool {
	return w.collapsed
}

// FirstPanel returns the first panel widget.
func (w *Widget) FirstPanel() widget.Widget {
	return w.cfg.first
}

// SecondPanel returns the second panel widget.
func (w *Widget) SecondPanel() widget.Widget {
	return w.cfg.second
}

// computePanelRects calculates the panel rectangles given the split parameters.
func computePanelRects(size geometry.Size, origin geometry.Point, ratio, divW float32, orient Orientation) (first, second geometry.Rect) {
	if orient == Horizontal {
		totalSpace := size.Width - divW
		firstWidth := totalSpace * ratio
		secondWidth := totalSpace - firstWidth

		first = geometry.NewRect(origin.X, origin.Y, firstWidth, size.Height)
		second = geometry.NewRect(origin.X+firstWidth+divW, origin.Y, secondWidth, size.Height)
	} else {
		totalSpace := size.Height - divW
		firstHeight := totalSpace * ratio
		secondHeight := totalSpace - firstHeight

		first = geometry.NewRect(origin.X, origin.Y, size.Width, firstHeight)
		second = geometry.NewRect(origin.X, origin.Y+firstHeight+divW, size.Width, secondHeight)
	}
	return first, second
}

// clampRatio clamps a ratio to the valid range [0, 1].
func clampRatio(r float32) float32 {
	if r < 0 {
		return 0
	}
	if r > 1 {
		return 1
	}
	return r
}

// Verify Widget implements required interfaces at compile time.
var (
	_ widget.Widget    = (*Widget)(nil)
	_ widget.Lifecycle = (*Widget)(nil)
)
