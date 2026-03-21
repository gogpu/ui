package stripe

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/widget"
)

// Painter draws the visual representation of a stripe toolbar.
// Each design system (DevTools, Material 3, Fluent) provides its own
// Painter implementation to render the stripe in its visual style.
//
// If no Painter is set, the stripe uses [DefaultPainter].
type Painter interface {
	// PaintBackground renders the stripe background and optional border.
	PaintBackground(canvas widget.Canvas, bounds geometry.Rect)

	// PaintButton renders a single button within the stripe.
	PaintButton(canvas widget.Canvas, state ButtonPaintState)
}

// ButtonPaintState provides the current button state to the painter.
type ButtonPaintState struct {
	// Bounds is the button's rectangle in local coordinates.
	Bounds geometry.Rect

	// Icon is the vector icon data for the button.
	Icon icon.IconData

	// Label is the button's text label.
	Label string

	// Active indicates this button is the currently selected tool window.
	Active bool

	// Hovered indicates the mouse cursor is over this button.
	Hovered bool

	// Pressed indicates the mouse button is held down on this button.
	Pressed bool

	// ShowLabel controls whether the text label is rendered below the icon.
	ShowLabel bool
}

// DefaultPainter provides a minimal fallback painter with no design system styling.
// It draws simple gray buttons suitable for testing and prototyping.
type DefaultPainter struct{}

// PaintBackground renders a light gray background for the stripe.
func (p DefaultPainter) PaintBackground(canvas widget.Canvas, bounds geometry.Rect) {
	if bounds.IsEmpty() {
		return
	}
	canvas.DrawRect(bounds, defaultStripeBg)
}

// PaintButton renders a button with icon and optional text label.
func (p DefaultPainter) PaintButton(canvas widget.Canvas, state ButtonPaintState) {
	if state.Bounds.IsEmpty() {
		return
	}

	// Background on hover/press/active.
	bg := widget.ColorTransparent
	switch {
	case state.Pressed:
		bg = defaultPressedBg
	case state.Hovered:
		bg = defaultHoverBg
	case state.Active:
		bg = defaultActiveBg
	}
	if !bg.IsTransparent() {
		canvas.DrawRoundRect(state.Bounds, bg, defaultBtnRadius)
	}

	// Icon color.
	fg := defaultFgColor
	if state.Active {
		fg = defaultActiveFg
	}

	// Draw icon centered horizontally.
	iconBounds := buttonIconBounds(state.Bounds, state.ShowLabel)
	icon.Draw(canvas, state.Icon, iconBounds, fg)

	// Draw label below icon if enabled.
	if state.ShowLabel && state.Label != "" {
		textBounds := buttonTextBounds(state.Bounds, iconBounds)
		canvas.DrawText(state.Label, textBounds, defaultLabelFontSize, fg, false, widget.TextAlignCenter)
	}
}

// buttonIconBounds calculates the icon bounds within a button.
// With labels: icon is at the top, centered horizontally.
// Without labels: icon is centered both horizontally and vertically.
func buttonIconBounds(btnBounds geometry.Rect, showLabel bool) geometry.Rect {
	if showLabel {
		// Icon centered horizontally, near top of button.
		centerX := btnBounds.Min.X + btnBounds.Width()/2
		y := btnBounds.Min.Y + defaultIconPaddingLabel
		return geometry.NewRect(
			centerX-defaultIconSize/2,
			y,
			defaultIconSize,
			defaultIconSize,
		)
	}
	// Icon centered both ways.
	centerX := btnBounds.Min.X + btnBounds.Width()/2
	centerY := btnBounds.Min.Y + btnBounds.Height()/2
	return geometry.NewRect(
		centerX-defaultIconSize/2,
		centerY-defaultIconSize/2,
		defaultIconSize,
		defaultIconSize,
	)
}

// buttonTextBounds calculates text bounds below the icon.
func buttonTextBounds(btnBounds geometry.Rect, iconRect geometry.Rect) geometry.Rect {
	y := iconRect.Max.Y + defaultIconTextGap
	return geometry.NewRect(
		btnBounds.Min.X+defaultTextPaddingH,
		y,
		btnBounds.Width()-defaultTextPaddingH*2,
		btnBounds.Max.Y-y,
	)
}

// Painting constants matching JetBrains IDE sizing.
const (
	defaultIconSize         float32 = 20
	defaultIconPaddingLabel float32 = 4
	defaultIconPaddingOnly  float32 = 5
	defaultIconTextGap      float32 = 3
	defaultLabelFontSize    float32 = 11
	defaultTextPaddingH     float32 = 6
	defaultBtnRadius        float32 = 12
)

// Default colors for DefaultPainter.
var (
	defaultStripeBg  = widget.Hex(0xF5F5F5)
	defaultHoverBg   = widget.RGBA(0, 0, 0, 0.08)
	defaultPressedBg = widget.RGBA(0, 0, 0, 0.16)
	defaultActiveBg  = widget.RGBA(0, 0, 0, 0.12)
	defaultFgColor   = widget.RGBA(0.4, 0.4, 0.4, 1.0)
	defaultActiveFg  = widget.RGBA(0.1, 0.1, 0.1, 1.0)
)

// Compile-time check that DefaultPainter implements Painter.
var _ Painter = DefaultPainter{}
