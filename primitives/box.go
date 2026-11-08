package primitives

import (
	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// BoxStyle holds all visual styling for a [BoxWidget].
type BoxStyle struct {
	Padding    geometry.Insets
	Background widget.Color
	Radius     float32
	Border     Border
	Shadow     Shadow
	Gap        float32

	// Explicit dimensions. Zero means unconstrained in that dimension.
	ExplicitWidth  float32
	ExplicitHeight float32
	MinWidth       float32
	MinHeight      float32
	MaxWidth       float32
	MaxHeight      float32
}

// BoxWidget is a container that lays out children vertically or horizontally
// with optional padding, background, border, rounded corners, shadow, and gap.
//
// BoxWidget implements [widget.Widget], [a11y.Accessible], and [widget.Lifecycle].
//
// Create a BoxWidget with the [Box] constructor. Use [HBox] or [VBox] for
// convenience constructors with a pre-set direction.
// CrossAxisAlignment controls how children are positioned on the cross axis.
// For VBox (vertical): cross axis = horizontal. For HBox: cross axis = vertical.
// Flutter equivalent: CrossAxisAlignment in Column/Row.
type CrossAxisAlignment int

const (
	// CrossAxisStart aligns children to the start (left for VBox, top for HBox).
	CrossAxisStart CrossAxisAlignment = iota
	// CrossAxisCenter centers children on the cross axis.
	CrossAxisCenter
	// CrossAxisEnd aligns children to the end (right for VBox, bottom for HBox).
	CrossAxisEnd
	// CrossAxisStretch stretches children to fill the cross axis (default).
	CrossAxisStretch
)

// BoxWidget is a container that lays out children vertically or horizontally
// with optional padding, background, border, rounded corners, shadow, and gap.
//
// BoxWidget implements [widget.Widget], [a11y.Accessible], and [widget.Lifecycle].
//
// Create a BoxWidget with the [Box] constructor. Use [HBox] or [VBox] for
// convenience constructors with a pre-set direction.
type BoxWidget struct {
	widget.WidgetBase

	style              BoxStyle
	direction          Direction
	directionSignal    state.ReadonlySignal[Direction]
	crossAlign         CrossAxisAlignment
	accessibilityLabel string
}

// Box creates a new container widget with the given children.
//
// Children are laid out vertically from top to bottom by default. Use
// [BoxWidget.SetDirection] to switch to horizontal layout, or use the
// [HBox] and [VBox] convenience constructors.
//
//	card := primitives.Box(
//	    primitives.Text("Title").Bold(),
//	    primitives.Text("Body"),
//	).Padding(16).Background(widget.Hex(0xFFFFFF))
func Box(children ...widget.Widget) *BoxWidget {
	b := &BoxWidget{}
	// Register the concrete container so AddChild can set the real parent on
	// each child (WidgetBase alone cannot know its enclosing type).
	b.SetSelf(b)
	b.SetVisible(true)
	b.SetEnabled(true)
	// Children live in the single WidgetBase store. AddChild establishes the
	// parent chain for upward dirty propagation (Flutter adoptChild) and mounts
	// any child added to an already-mounted box.
	for _, child := range children {
		b.AddChild(child)
	}
	return b
}

// HBox creates a new container that lays out children horizontally (left to right).
//
// This is a convenience constructor equivalent to Box(children...).SetDirection(DirectionHorizontal).
//
//	row := primitives.HBox(icon, label, spacer).Gap(8)
func HBox(children ...widget.Widget) *BoxWidget {
	b := Box(children...)
	b.direction = DirectionHorizontal
	return b
}

// VBox creates a new container that lays out children vertically (top to bottom).
//
// This is a convenience constructor equivalent to Box(children...).SetDirection(DirectionVertical).
// Since vertical is the default direction, VBox is primarily useful for readability
// when paired with [HBox] in the same codebase.
//
//	column := primitives.VBox(title, subtitle, body).Gap(4)
func VBox(children ...widget.Widget) *BoxWidget {
	return Box(children...)
}

// --- Fluent style methods ---

// Padding sets uniform padding on all edges.
func (b *BoxWidget) Padding(v float32) *BoxWidget {
	b.style.Padding = geometry.UniformInsets(v)
	return b
}

// PaddingXY sets separate horizontal and vertical padding.
func (b *BoxWidget) PaddingXY(x, y float32) *BoxWidget {
	b.style.Padding = geometry.SymmetricInsets(x, y)
	return b
}

// PaddingTop sets the top padding.
func (b *BoxWidget) PaddingTop(v float32) *BoxWidget {
	b.style.Padding.Top = v
	return b
}

// PaddingRight sets the right padding.
func (b *BoxWidget) PaddingRight(v float32) *BoxWidget {
	b.style.Padding.Right = v
	return b
}

// PaddingBottom sets the bottom padding.
func (b *BoxWidget) PaddingBottom(v float32) *BoxWidget {
	b.style.Padding.Bottom = v
	return b
}

// PaddingLeft sets the left padding.
func (b *BoxWidget) PaddingLeft(v float32) *BoxWidget {
	b.style.Padding.Left = v
	return b
}

// Background sets the background color.
func (b *BoxWidget) Background(c widget.Color) *BoxWidget {
	b.style.Background = c
	return b
}

// Rounded sets a uniform border radius.
func (b *BoxWidget) Rounded(r float32) *BoxWidget {
	b.style.Radius = r
	return b
}

// BorderStyle sets the border width and color.
func (b *BoxWidget) BorderStyle(width float32, color widget.Color) *BoxWidget {
	b.style.Border = Border{Width: width, Color: color}
	return b
}

// ShadowLevel sets the elevation shadow level (0-5).
func (b *BoxWidget) ShadowLevel(level int) *BoxWidget {
	if level < 0 {
		level = 0
	}
	if level > maxShadowLevel {
		level = maxShadowLevel
	}
	b.style.Shadow = Shadow{Level: level}
	return b
}

// Gap sets the spacing between children. The gap is applied in the layout
// direction: vertically for [DirectionVertical], horizontally for [DirectionHorizontal].
func (b *BoxWidget) Gap(v float32) *BoxWidget {
	b.style.Gap = v
	return b
}

// CrossAlign sets the cross-axis alignment for child widgets.
// For VBox: controls horizontal alignment (center, start, end, stretch).
// For HBox: controls vertical alignment.
// Default is CrossAxisStart (left-aligned for VBox).
//
// Flutter equivalent: CrossAxisAlignment on Column/Row.
//
//	primitives.VBox(spinner).CrossAlign(primitives.CrossAxisCenter)
func (b *BoxWidget) CrossAlign(a CrossAxisAlignment) *BoxWidget {
	b.crossAlign = a
	return b
}

// SetDirection sets the layout direction for child widgets.
//
//	primitives.Box(a, b, c).SetDirection(primitives.DirectionHorizontal)
func (b *BoxWidget) SetDirection(d Direction) *BoxWidget {
	if b.direction != d {
		b.direction = d
		b.SetNeedsRedraw(true)
	}
	return b
}

// DirectionSignal binds the layout direction to a read-only reactive signal.
// When set, the signal value takes precedence over the static direction set
// via [BoxWidget.SetDirection].
//
// Because BoxWidget is a container (not user-editable), only
// [state.ReadonlySignal] is accepted (no write-back capability is needed).
//
//	dir := state.NewSignal(primitives.DirectionHorizontal)
//	box := primitives.Box(a, b).DirectionSignal(dir)
func (b *BoxWidget) DirectionSignal(sig state.ReadonlySignal[Direction]) *BoxWidget {
	b.directionSignal = sig
	return b
}

// ResolvedDirection returns the effective layout direction.
// Priority: ReadonlySignal > Static.
func (b *BoxWidget) ResolvedDirection() Direction {
	if b.directionSignal != nil {
		return b.directionSignal.Get()
	}
	return b.direction
}

// Width sets an explicit width.
func (b *BoxWidget) Width(v float32) *BoxWidget {
	b.style.ExplicitWidth = v
	return b
}

// Height sets an explicit height.
func (b *BoxWidget) Height(v float32) *BoxWidget {
	b.style.ExplicitHeight = v
	return b
}

// MinWidthValue sets the minimum width.
func (b *BoxWidget) MinWidthValue(v float32) *BoxWidget {
	b.style.MinWidth = v
	return b
}

// MinHeightValue sets the minimum height.
func (b *BoxWidget) MinHeightValue(v float32) *BoxWidget {
	b.style.MinHeight = v
	return b
}

// MaxWidthValue sets the maximum width.
func (b *BoxWidget) MaxWidthValue(v float32) *BoxWidget {
	b.style.MaxWidth = v
	return b
}

// MaxHeightValue sets the maximum height.
func (b *BoxWidget) MaxHeightValue(v float32) *BoxWidget {
	b.style.MaxHeight = v
	return b
}

// Label sets a custom accessibility label for this container.
func (b *BoxWidget) Label(label string) *BoxWidget {
	b.accessibilityLabel = label
	return b
}

// Style returns the current box style (read-only snapshot).
func (b *BoxWidget) Style() BoxStyle {
	return b.style
}

// --- widget.Widget interface ---

// Layout calculates the box size by laying out children in the resolved
// direction (vertical or horizontal) with padding and gap, then constraining
// the result.
func (b *BoxWidget) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	if b.ResolvedDirection() == DirectionHorizontal {
		return b.layoutHorizontal(ctx, constraints)
	}
	return b.layoutVertical(ctx, constraints)
}

// layoutVertical lays out children top-to-bottom.
//
// When one or more children are wrapped in [Expanded], a two-pass algorithm
// is used: fixed children are measured first, then remaining height is divided
// equally among expanded children. Without any Expanded children, the layout
// behaves identically to the original sequential algorithm.
func (b *BoxWidget) layoutVertical(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	constraints = b.applyExplicitConstraints(constraints)

	pad := b.style.Padding
	childConstraints := constraints.Deflate(pad)
	if childConstraints.MaxWidth < 0 {
		childConstraints.MaxWidth = 0
	}
	if childConstraints.MaxHeight < 0 {
		childConstraints.MaxHeight = 0
	}

	// Snapshot the child list once per pass. A child removed from inside its
	// own Layout must not shift the slice under us (GTK save-next / Flutter
	// toList equivalent).
	children := b.Children()

	// Count expanded children to decide between single-pass and two-pass.
	expandedCount := countExpanded(children)

	if expandedCount == 0 {
		return b.layoutVerticalSimple(ctx, constraints, pad, childConstraints, children)
	}

	return b.layoutVerticalTwoPass(ctx, constraints, pad, childConstraints, children, expandedCount)
}

// layoutVerticalSimple is the original single-pass vertical layout (no Expanded children).
func (b *BoxWidget) layoutVerticalSimple(
	ctx widget.Context,
	constraints geometry.Constraints,
	pad geometry.Insets,
	childConstraints geometry.Constraints,
	children []widget.Widget,
) geometry.Size {
	childCount := len(children)
	var totalHeight float32
	var maxChildWidth float32

	for i, child := range children {
		remaining := childConstraints
		if childConstraints.HasBoundedHeight() {
			remaining.MaxHeight = childConstraints.MaxHeight - totalHeight
			if remaining.MaxHeight < 0 {
				remaining.MaxHeight = 0
			}
		}
		remaining.MinHeight = 0

		// CrossAxisStretch: give child full width. Others: let child choose.
		if b.crossAlign != CrossAxisStretch {
			remaining.MinWidth = 0
		}

		size := widget.LayoutChild(child, ctx, remaining)

		childX := pad.Left
		childY := pad.Top + totalHeight

		totalHeight += size.Height
		if i < childCount-1 {
			totalHeight += b.style.Gap
		}
		if size.Width > maxChildWidth {
			maxChildWidth = size.Width
		}

		child.(interface{ SetBounds(geometry.Rect) }).SetBounds(
			geometry.FromPointSize(geometry.Pt(childX, childY), size),
		)
	}

	contentWidth := maxChildWidth + pad.Horizontal()
	contentHeight := totalHeight + pad.Vertical()

	resultSize := constraints.Constrain(geometry.Sz(contentWidth, contentHeight))

	// Second pass: apply cross-axis alignment now that we know total width.
	if b.crossAlign == CrossAxisCenter || b.crossAlign == CrossAxisEnd {
		availWidth := resultSize.Width - pad.Horizontal()
		for _, child := range children {
			cb := child.(interface{ Bounds() geometry.Rect }).Bounds()
			cw := cb.Width()
			ch := cb.Height()
			var newX float32
			if b.crossAlign == CrossAxisCenter {
				newX = pad.Left + (availWidth-cw)/2
			} else {
				newX = pad.Left + availWidth - cw
			}
			child.(interface{ SetBounds(geometry.Rect) }).SetBounds(
				geometry.NewRect(newX, cb.Min.Y, cw, ch),
			)
		}
	}
	b.SetBounds(geometry.FromPointSize(b.Position(), resultSize))
	return resultSize
}

// layoutVerticalTwoPass measures fixed children first, then distributes
// remaining height equally among Expanded children.
func (b *BoxWidget) layoutVerticalTwoPass(
	ctx widget.Context,
	constraints geometry.Constraints,
	pad geometry.Insets,
	childConstraints geometry.Constraints,
	children []widget.Widget,
	expandedCount int,
) geometry.Size {
	childCount := len(children)

	// Pass 1: measure fixed (non-expanded) children.
	fixedSizes := make([]geometry.Size, childCount)
	var fixedHeight float32
	var maxChildWidth float32

	for i, child := range children {
		if isExpanded(child) {
			continue
		}
		remaining := childConstraints
		remaining.MinHeight = 0
		size := widget.LayoutChild(child, ctx, remaining)
		fixedSizes[i] = size
		fixedHeight += size.Height
		if size.Width > maxChildWidth {
			maxChildWidth = size.Width
		}
	}

	// Calculate gap total and remaining height for expanded children.
	gapTotal := b.style.Gap * float32(childCount-1)
	var expandedHeight float32
	if childConstraints.HasBoundedHeight() {
		expandedHeight = (childConstraints.MaxHeight - fixedHeight - gapTotal) / float32(expandedCount)
		if expandedHeight < 0 {
			expandedHeight = 0
		}
	}

	// Pass 2: layout expanded children and position all.
	var totalHeight float32
	for i, child := range children {
		var size geometry.Size
		if isExpanded(child) {
			ec := geometry.Constraints{
				MinWidth:  childConstraints.MinWidth,
				MaxWidth:  childConstraints.MaxWidth,
				MinHeight: expandedHeight,
				MaxHeight: expandedHeight,
			}
			size = widget.LayoutChild(child, ctx, ec)
		} else {
			size = fixedSizes[i]
		}

		childX := pad.Left
		childY := pad.Top + totalHeight
		child.(interface{ SetBounds(geometry.Rect) }).SetBounds(
			geometry.FromPointSize(geometry.Pt(childX, childY), size),
		)

		totalHeight += size.Height
		if i < childCount-1 {
			totalHeight += b.style.Gap
		}
		if size.Width > maxChildWidth {
			maxChildWidth = size.Width
		}
	}

	contentWidth := maxChildWidth + pad.Horizontal()
	contentHeight := totalHeight + pad.Vertical()

	resultSize := constraints.Constrain(geometry.Sz(contentWidth, contentHeight))
	b.SetBounds(geometry.FromPointSize(b.Position(), resultSize))
	return resultSize
}

// layoutHorizontal lays out children left-to-right.
//
// When one or more children are wrapped in [Expanded], a two-pass algorithm
// is used: fixed children are measured first, then remaining width is divided
// equally among expanded children. Without any Expanded children, the layout
// behaves identically to the original sequential algorithm.
func (b *BoxWidget) layoutHorizontal(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	constraints = b.applyExplicitConstraints(constraints)

	pad := b.style.Padding
	childConstraints := constraints.Deflate(pad)
	if childConstraints.MaxWidth < 0 {
		childConstraints.MaxWidth = 0
	}
	if childConstraints.MaxHeight < 0 {
		childConstraints.MaxHeight = 0
	}

	// Snapshot the child list once per pass (see layoutVertical).
	children := b.Children()

	// Count expanded children to decide between single-pass and two-pass.
	expandedCount := countExpanded(children)

	if expandedCount == 0 {
		return b.layoutHorizontalSimple(ctx, constraints, pad, childConstraints, children)
	}

	return b.layoutHorizontalTwoPass(ctx, constraints, pad, childConstraints, children, expandedCount)
}

// layoutHorizontalSimple is the original single-pass horizontal layout (no Expanded children).
func (b *BoxWidget) layoutHorizontalSimple(
	ctx widget.Context,
	constraints geometry.Constraints,
	pad geometry.Insets,
	childConstraints geometry.Constraints,
	children []widget.Widget,
) geometry.Size {
	childCount := len(children)
	var totalWidth float32
	var maxChildHeight float32

	for i, child := range children {
		remaining := childConstraints
		if childConstraints.HasBoundedWidth() {
			remaining.MaxWidth = childConstraints.MaxWidth - totalWidth
			if remaining.MaxWidth < 0 {
				remaining.MaxWidth = 0
			}
		}
		remaining.MinWidth = 0

		size := widget.LayoutChild(child, ctx, remaining)

		childX := pad.Left + totalWidth
		childY := pad.Top
		child.(interface{ SetBounds(geometry.Rect) }).SetBounds(
			geometry.FromPointSize(geometry.Pt(childX, childY), size),
		)

		totalWidth += size.Width
		if i < childCount-1 {
			totalWidth += b.style.Gap
		}
		if size.Height > maxChildHeight {
			maxChildHeight = size.Height
		}
	}

	contentWidth := totalWidth + pad.Horizontal()
	contentHeight := maxChildHeight + pad.Vertical()

	resultSize := constraints.Constrain(geometry.Sz(contentWidth, contentHeight))
	b.SetBounds(geometry.FromPointSize(b.Position(), resultSize))
	return resultSize
}

// layoutHorizontalTwoPass measures fixed children first, then distributes
// remaining width equally among Expanded children.
func (b *BoxWidget) layoutHorizontalTwoPass(
	ctx widget.Context,
	constraints geometry.Constraints,
	pad geometry.Insets,
	childConstraints geometry.Constraints,
	children []widget.Widget,
	expandedCount int,
) geometry.Size {
	childCount := len(children)

	// Pass 1: measure fixed (non-expanded) children.
	fixedSizes := make([]geometry.Size, childCount)
	var fixedWidth float32
	var maxChildHeight float32

	for i, child := range children {
		if isExpanded(child) {
			continue
		}
		remaining := childConstraints
		remaining.MinWidth = 0
		size := widget.LayoutChild(child, ctx, remaining)
		fixedSizes[i] = size
		fixedWidth += size.Width
		if size.Height > maxChildHeight {
			maxChildHeight = size.Height
		}
	}

	// Calculate gap total and remaining width for expanded children.
	gapTotal := b.style.Gap * float32(childCount-1)
	var expandedWidth float32
	if childConstraints.HasBoundedWidth() {
		expandedWidth = (childConstraints.MaxWidth - fixedWidth - gapTotal) / float32(expandedCount)
		if expandedWidth < 0 {
			expandedWidth = 0
		}
	}

	// Pass 2: layout expanded children and position all.
	var totalWidth float32
	for i, child := range children {
		var size geometry.Size
		if isExpanded(child) {
			ec := geometry.Constraints{
				MinWidth:  expandedWidth,
				MaxWidth:  expandedWidth,
				MinHeight: childConstraints.MinHeight,
				MaxHeight: childConstraints.MaxHeight,
			}
			size = widget.LayoutChild(child, ctx, ec)
		} else {
			size = fixedSizes[i]
		}

		childX := pad.Left + totalWidth
		childY := pad.Top
		child.(interface{ SetBounds(geometry.Rect) }).SetBounds(
			geometry.FromPointSize(geometry.Pt(childX, childY), size),
		)

		totalWidth += size.Width
		if i < childCount-1 {
			totalWidth += b.style.Gap
		}
		if size.Height > maxChildHeight {
			maxChildHeight = size.Height
		}
	}

	contentWidth := totalWidth + pad.Horizontal()
	contentHeight := maxChildHeight + pad.Vertical()

	resultSize := constraints.Constrain(geometry.Sz(contentWidth, contentHeight))
	b.SetBounds(geometry.FromPointSize(b.Position(), resultSize))
	return resultSize
}

// Draw renders the box background, border, shadow, and then draws all children.
func (b *BoxWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !b.IsVisible() {
		return
	}

	bounds := b.Bounds()

	// Draw shadow layers (outermost first, innermost last).
	if !b.style.Shadow.IsZero() {
		b.drawShadow(canvas, bounds)
	}

	// Draw background
	if !b.style.Background.IsTransparent() {
		if b.style.Radius > 0 {
			canvas.DrawRoundRect(bounds, b.style.Background, b.style.Radius)
		} else {
			canvas.DrawRect(bounds, b.style.Background)
		}
	}

	// Clip children when the box has a border or rounded corners,
	// so content doesn't overflow the visual boundary.
	// Uses PushClipRoundRect for rounded corners (GPU SDF clip),
	// falls back to rectangular PushClip otherwise.
	clipChildren := !b.style.Border.IsZero() || b.style.Radius > 0
	if clipChildren {
		if b.style.Radius > 0 {
			canvas.PushClipRoundRect(bounds, b.style.Radius)
		} else {
			canvas.PushClip(bounds)
		}
	}

	// Draw children with transform offset for this box's position.
	// No viewport culling here — DrawChild skip pattern makes offscreen
	// boundary children cheap (O(1) stamp + skip). Compositor-level culling
	// via CompositorClip handles GPU texture rendering and compositing.
	// Flutter: PaintingContext.paintChild always calls paint on all children;
	// visibility culling is in the compositor (addRetained vs addToScene).
	canvas.PushTransform(bounds.Min)
	// Iterate a snapshot so a child removed from inside Draw does not shift the
	// slice or skip a sibling.
	for _, child := range b.Children() {
		widget.StampScreenOrigin(child, canvas)
		widget.DrawChild(child, ctx, canvas)
	}
	canvas.PopTransform()

	if clipChildren {
		canvas.PopClip()
	}

	// Draw border AFTER children so it renders on top of content.
	if !b.style.Border.IsZero() {
		if b.style.Radius > 0 {
			canvas.StrokeRoundRect(bounds, b.style.Border.Color, b.style.Radius, b.style.Border.Width)
		} else {
			canvas.StrokeRect(bounds, b.style.Border.Color, b.style.Border.Width)
		}
	}
}

// drawShadow renders multi-layer concentric rounded rectangles that
// approximate a Gaussian blur shadow. Each layer is slightly larger and
// more offset than the previous one, with decreasing alpha, creating
// a soft gradient falloff that looks like a real elevation shadow.
func (b *BoxWidget) drawShadow(canvas widget.Canvas, bounds geometry.Rect) {
	layers := shadowLayers(b.style.Shadow.Level)
	for _, layer := range layers {
		r := bounds.Expand(layer.spread).TranslateXY(layer.offsetX, layer.offsetY)
		color := widget.RGBA(0, 0, 0, layer.alpha)
		radius := b.style.Radius + layer.radiusExtra
		if radius > 0 {
			canvas.DrawRoundRect(r, color, radius)
		} else {
			canvas.DrawRect(r, color)
		}
	}
}

// Event dispatches the event to children. Returns true if any child consumed it.
func (b *BoxWidget) Event(ctx widget.Context, e event.Event) bool {
	if !b.IsVisible() || !b.IsEnabled() {
		return false
	}

	// For mouse events, translate position to local coordinates and
	// dispatch only to children whose bounds contain the position.
	if me, ok := e.(*event.MouseEvent); ok {
		return b.dispatchMouseEvent(ctx, me)
	}

	// For wheel events, translate position to local coordinates
	// so hit-testing in children works correctly.
	if we, ok := e.(*event.WheelEvent); ok {
		return b.dispatchWheelEvent(ctx, we)
	}

	// Non-mouse events (keyboard, focus) broadcast to all children. Snapshot
	// the list once so a child removed during dispatch does not shift it.
	children := b.Children()
	for i := len(children) - 1; i >= 0; i-- {
		if children[i].Event(ctx, e) {
			return true
		}
	}
	return false
}

// dispatchMouseEvent translates mouse coordinates to Box-local space
// and dispatches only to children whose bounds contain the position.
// This mirrors PushTransform(bounds.Min) used in Draw.
func (b *BoxWidget) dispatchMouseEvent(ctx widget.Context, e *event.MouseEvent) bool {
	local := *e
	local.Position = e.Position.Sub(b.Bounds().Min)

	children := b.Children()
	for i := len(children) - 1; i >= 0; i-- {
		child := children[i]
		if bw, ok := child.(interface{ Bounds() geometry.Rect }); ok {
			if !bw.Bounds().Contains(local.Position) {
				continue
			}
		}
		if child.Event(ctx, &local) {
			return true
		}
	}
	return false
}

// dispatchWheelEvent translates wheel event position to Box-local space
// and dispatches to children whose bounds contain the position.
func (b *BoxWidget) dispatchWheelEvent(ctx widget.Context, e *event.WheelEvent) bool {
	local := *e
	local.Position = e.Position.Sub(b.Bounds().Min)

	children := b.Children()
	for i := len(children) - 1; i >= 0; i-- {
		child := children[i]
		if bw, ok := child.(interface{ Bounds() geometry.Rect }); ok {
			if !bw.Bounds().Contains(local.Position) {
				continue
			}
		}
		if child.Event(ctx, &local) {
			return true
		}
	}
	return false
}

// --- a11y.Accessible interface ---

// AccessibilityRole returns [a11y.RoleGenericContainer].
func (b *BoxWidget) AccessibilityRole() a11y.Role {
	return a11y.RoleGenericContainer
}

// AccessibilityLabel returns the custom label, or empty string if none set.
func (b *BoxWidget) AccessibilityLabel() string {
	return b.accessibilityLabel
}

// AccessibilityHint returns an empty string. Containers typically have no hint.
func (b *BoxWidget) AccessibilityHint() string {
	return ""
}

// AccessibilityValue returns an empty string. Containers have no value.
func (b *BoxWidget) AccessibilityValue() string {
	return ""
}

// AccessibilityState returns the default state.
func (b *BoxWidget) AccessibilityState() a11y.State {
	return a11y.State{
		Disabled: !b.IsEnabled(),
		Hidden:   !b.IsVisible(),
	}
}

// AccessibilityActions returns nil. Containers have no actions.
func (b *BoxWidget) AccessibilityActions() []a11y.Action {
	return nil
}

// countExpanded returns the number of children that are wrapped in [Expanded].
func countExpanded(children []widget.Widget) int {
	var count int
	for _, child := range children {
		if isExpanded(child) {
			count++
		}
	}
	return count
}

// applyExplicitConstraints integrates explicit dimensions and min/max into
// the parent constraints.
func (b *BoxWidget) applyExplicitConstraints(c geometry.Constraints) geometry.Constraints {
	if b.style.ExplicitWidth > 0 {
		c.MinWidth = b.style.ExplicitWidth
		c.MaxWidth = b.style.ExplicitWidth
	}
	if b.style.ExplicitHeight > 0 {
		c.MinHeight = b.style.ExplicitHeight
		c.MaxHeight = b.style.ExplicitHeight
	}
	if b.style.MinWidth > 0 && b.style.MinWidth > c.MinWidth {
		c.MinWidth = b.style.MinWidth
	}
	if b.style.MinHeight > 0 && b.style.MinHeight > c.MinHeight {
		c.MinHeight = b.style.MinHeight
	}
	if b.style.MaxWidth > 0 && b.style.MaxWidth < c.MaxWidth {
		c.MaxWidth = b.style.MaxWidth
	}
	if b.style.MaxHeight > 0 && b.style.MaxHeight < c.MaxHeight {
		c.MaxHeight = b.style.MaxHeight
	}
	return c
}

// Mount creates signal bindings for push-based invalidation.
// Implements [widget.Lifecycle].
func (b *BoxWidget) Mount(ctx widget.Context) {
	sched := ctx.Scheduler()
	if sched == nil {
		return
	}
	if b.directionSignal != nil {
		binding := state.BindToScheduler(b.directionSignal, b, sched)
		b.AddBinding(binding)
	}
}

// Unmount is called when the box widget is removed from the widget tree.
// Implements [widget.Lifecycle].
func (b *BoxWidget) Unmount() {
	// Bindings are cleaned up automatically by WidgetBase.CleanupBindings().
}

// Compile-time interface checks.
var (
	_ widget.Widget    = (*BoxWidget)(nil)
	_ a11y.Accessible  = (*BoxWidget)(nil)
	_ widget.Lifecycle = (*BoxWidget)(nil)
)
