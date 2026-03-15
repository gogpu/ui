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
type BoxWidget struct {
	widget.WidgetBase

	style              BoxStyle
	direction          Direction
	directionSignal    state.ReadonlySignal[Direction]
	children           []widget.Widget
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
	b := &BoxWidget{
		children: children,
	}
	b.SetVisible(true)
	b.SetEnabled(true)
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

// layoutVertical lays out children top-to-bottom (the original behavior).
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

	var totalHeight float32
	var maxChildWidth float32
	childCount := len(b.children)

	for i, child := range b.children {
		remaining := childConstraints
		if childConstraints.HasBoundedHeight() {
			remaining.MaxHeight = childConstraints.MaxHeight - totalHeight
			if remaining.MaxHeight < 0 {
				remaining.MaxHeight = 0
			}
		}
		remaining.MinHeight = 0

		size := child.Layout(ctx, remaining)

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

	var totalWidth float32
	var maxChildHeight float32
	childCount := len(b.children)

	for i, child := range b.children {
		remaining := childConstraints
		if childConstraints.HasBoundedWidth() {
			remaining.MaxWidth = childConstraints.MaxWidth - totalWidth
			if remaining.MaxWidth < 0 {
				remaining.MaxWidth = 0
			}
		}
		remaining.MinWidth = 0

		size := child.Layout(ctx, remaining)

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
	canvas.PushTransform(bounds.Min)
	for _, child := range b.children {
		widget.StampScreenOrigin(child, canvas)
		child.Draw(ctx, canvas)
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

	// Non-mouse events (keyboard, focus) broadcast to all children.
	for i := len(b.children) - 1; i >= 0; i-- {
		if b.children[i].Event(ctx, e) {
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

	for i := len(b.children) - 1; i >= 0; i-- {
		child := b.children[i]
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

	for i := len(b.children) - 1; i >= 0; i-- {
		child := b.children[i]
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

// Children returns the box's child widgets.
func (b *BoxWidget) Children() []widget.Widget {
	if len(b.children) == 0 {
		return nil
	}
	result := make([]widget.Widget, len(b.children))
	copy(result, b.children)
	return result
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
