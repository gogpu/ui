package toolbar

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/widget"
)

// Painter draws the visual representation of a toolbar.
// Each design system (Material 3, Fluent, Cupertino) provides its own
// Painter implementation to render the toolbar in its visual style.
//
// If no Painter is set, the toolbar uses [DefaultPainter].
type Painter interface {
	// PaintToolbar renders the toolbar background.
	PaintToolbar(canvas widget.Canvas, state PaintToolbarState)

	// PaintButtonItem renders a single button item within the toolbar.
	PaintButtonItem(canvas widget.Canvas, state PaintButtonState)

	// PaintSeparator renders a separator item within the toolbar.
	PaintSeparator(canvas widget.Canvas, bounds geometry.Rect)
}

// PaintToolbarState provides the current toolbar state to the painter.
type PaintToolbarState struct {
	Bounds geometry.Rect
}

// PaintButtonState provides the current button item state to the painter.
type PaintButtonState struct {
	Label     string
	Icon      icon.IconData
	ShowLabel bool
	Hovered   bool
	Pressed   bool
	Focused   bool
	Disabled  bool
	Bounds    geometry.Rect
}

// DefaultPainter provides a minimal fallback painter with no design system styling.
// It draws simple gray buttons suitable for testing and prototyping.
type DefaultPainter struct{}

// PaintToolbar renders a light gray background for the toolbar.
func (p DefaultPainter) PaintToolbar(canvas widget.Canvas, state PaintToolbarState) {
	if state.Bounds.IsEmpty() {
		return
	}
	canvas.DrawRect(state.Bounds, defaultToolbarBg)
}

// PaintButtonItem renders a button item with icon and optional text label.
func (p DefaultPainter) PaintButtonItem(canvas widget.Canvas, state PaintButtonState) {
	if state.Bounds.IsEmpty() {
		return
	}

	// Background on hover/press (disabled items get no feedback).
	if !state.Disabled {
		bg := widget.ColorTransparent
		if state.Pressed {
			bg = defaultPressedBg
		} else if state.Hovered {
			bg = defaultHoverBg
		}
		if !bg.IsTransparent() {
			canvas.DrawRoundRect(state.Bounds, bg, defaultItemRadius)
		}
	}

	// Icon color.
	fg := defaultIconColor
	if state.Disabled {
		fg = defaultDisabledColor
	}

	// Draw icon centered in the button area.
	iconBounds := iconBoundsForItem(state.Bounds, state.ShowLabel)
	icon.Draw(canvas, state.Icon, iconBounds, fg)

	// Draw label text if ShowLabel is true.
	if state.ShowLabel && state.Label != "" {
		textBounds := textBoundsForItem(state.Bounds, iconBounds)
		canvas.DrawText(state.Label, textBounds, defaultFontSize, fg, false, widget.TextAlignLeft)
	}

	// Focus ring.
	if state.Focused && !state.Disabled {
		ringBounds := state.Bounds.Expand(focusRingOffset)
		canvas.StrokeRoundRect(ringBounds, focusRingColor, defaultItemRadius+focusRingOffset, focusRingStrokeWidth)
	}
}

// PaintSeparator renders a vertical line.
func (p DefaultPainter) PaintSeparator(canvas widget.Canvas, bounds geometry.Rect) {
	if bounds.IsEmpty() {
		return
	}
	centerX := bounds.Min.X + bounds.Width()/2
	canvas.DrawLine(
		geometry.Pt(centerX, bounds.Min.Y+separatorInset),
		geometry.Pt(centerX, bounds.Max.Y-separatorInset),
		defaultSeparatorColor, separatorStrokeWidth,
	)
}

// iconBoundsForItem calculates the bounds for the icon within a button item.
func iconBoundsForItem(itemBounds geometry.Rect, showLabel bool) geometry.Rect {
	h := itemBounds.Height() - iconPadding*2
	if h < 0 {
		h = 0
	}
	iconSize := h
	if iconSize > maxIconSize {
		iconSize = maxIconSize
	}
	if showLabel {
		// Icon on the left side.
		x := itemBounds.Min.X + iconPadding
		y := itemBounds.Min.Y + (itemBounds.Height()-iconSize)/2
		return geometry.NewRect(x, y, iconSize, iconSize)
	}
	// Icon centered.
	centerX := itemBounds.Min.X + itemBounds.Width()/2
	centerY := itemBounds.Min.Y + itemBounds.Height()/2
	return geometry.NewRect(
		centerX-iconSize/2,
		centerY-iconSize/2,
		iconSize,
		iconSize,
	)
}

// textBoundsForItem calculates the bounds for the label text, placed to the
// right of the icon bounds.
func textBoundsForItem(itemBounds geometry.Rect, iconRect geometry.Rect) geometry.Rect {
	x := iconRect.Max.X + textIconGap
	return geometry.NewRect(
		x,
		itemBounds.Min.Y,
		itemBounds.Max.X-iconPadding-x,
		itemBounds.Height(),
	)
}

// Painting constants.
const (
	defaultItemRadius    float32 = 6
	iconPadding          float32 = 6
	maxIconSize          float32 = 20
	textIconGap          float32 = 4
	defaultFontSize      float32 = 12
	separatorInset       float32 = 6
	separatorStrokeWidth float32 = 1
	focusRingOffset      float32 = 2
	focusRingStrokeWidth float32 = 2
)

// Default colors for DefaultPainter.
var (
	defaultToolbarBg      = widget.Hex(0xF5F5F5)
	defaultHoverBg        = widget.RGBA(0, 0, 0, 0.08)
	defaultPressedBg      = widget.RGBA(0, 0, 0, 0.16)
	defaultIconColor      = widget.RGBA(0.2, 0.2, 0.2, 1.0)
	defaultDisabledColor  = widget.RGBA(0.2, 0.2, 0.2, 0.38)
	defaultSeparatorColor = widget.RGBA(0, 0, 0, 0.12)
	focusRingColor        = widget.Hex(0x6750A4).WithAlpha(0.7)
)

// Compile-time check that DefaultPainter implements Painter.
var _ Painter = DefaultPainter{}
