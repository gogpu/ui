package primitives

import (
	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
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

// BoxWidget is a container that lays out children in a vertical stack with
// optional padding, background, border, rounded corners, shadow, and gap.
//
// BoxWidget implements [widget.Widget] and [a11y.Accessible].
//
// Create a BoxWidget with the [Box] constructor.
type BoxWidget struct {
	widget.WidgetBase

	style              BoxStyle
	children           []widget.Widget
	accessibilityLabel string
}

// Box creates a new container widget with the given children.
//
// Children are laid out vertically from top to bottom. Use the fluent
// methods to add styling.
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

// Gap sets the vertical spacing between children.
func (b *BoxWidget) Gap(v float32) *BoxWidget {
	b.style.Gap = v
	return b
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

// Layout calculates the box size by laying out children vertically with
// padding and gap, then constraining the result.
func (b *BoxWidget) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
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

	// Draw border
	if !b.style.Border.IsZero() {
		if b.style.Radius > 0 {
			canvas.StrokeRoundRect(bounds, b.style.Border.Color, b.style.Radius, b.style.Border.Width)
		} else {
			canvas.StrokeRect(bounds, b.style.Border.Color, b.style.Border.Width)
		}
	}

	// Draw children with transform offset for this box's position
	canvas.PushTransform(bounds.Min)
	for _, child := range b.children {
		child.Draw(ctx, canvas)
	}
	canvas.PopTransform()
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

// Compile-time interface checks.
var (
	_ widget.Widget   = (*BoxWidget)(nil)
	_ a11y.Accessible = (*BoxWidget)(nil)
)
