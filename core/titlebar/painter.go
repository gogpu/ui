package titlebar

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// Painter draws the visual representation of a title bar.
// Each design system (Material 3, Fluent, DevTools) provides its own
// Painter implementation to render the title bar in its visual style.
//
// If no Painter is set, the title bar uses [DefaultPainter].
type Painter interface {
	// DrawBackground draws the title bar background.
	DrawBackground(canvas widget.Canvas, bounds geometry.Rect, state BackgroundState)

	// DrawControlButton draws a window control button (minimize, maximize, close).
	DrawControlButton(canvas widget.Canvas, bounds geometry.Rect, control ControlType, state ControlState)
}

// ControlType identifies a window control button.
type ControlType uint8

const (
	// ControlMinimize is the minimize window button.
	ControlMinimize ControlType = iota

	// ControlMaximize is the maximize window button (when window is not maximized).
	ControlMaximize

	// ControlRestore is the restore window button (when window is maximized).
	ControlRestore

	// ControlClose is the close window button.
	ControlClose
)

// controlTypeNames maps each ControlType to its human-readable name.
var controlTypeNames = [...]string{
	ControlMinimize: "Minimize",
	ControlMaximize: "Maximize",
	ControlRestore:  "Restore",
	ControlClose:    "Close",
}

// String returns a human-readable name for the control type.
func (ct ControlType) String() string {
	if int(ct) < len(controlTypeNames) {
		return controlTypeNames[ct]
	}
	return "Unknown"
}

// BackgroundState provides the current title bar background state to the painter.
type BackgroundState struct {
	// Focused indicates whether the application window has input focus.
	Focused bool
}

// ControlState provides the current window control button state to the painter.
type ControlState struct {
	// Hovered indicates the mouse cursor is over the control button.
	Hovered bool

	// Pressed indicates the mouse button is held down on the control button.
	Pressed bool
}

// DefaultPainter provides a minimal fallback painter with no design system styling.
// It draws a simple dark bar with basic window controls.
type DefaultPainter struct{}

// DrawBackground renders a dark background bar.
func (p DefaultPainter) DrawBackground(canvas widget.Canvas, bounds geometry.Rect, _ BackgroundState) {
	if bounds.IsEmpty() {
		return
	}
	canvas.DrawRect(bounds, defaultBarBg)
}

// DrawControlButton renders a minimal window control button.
func (p DefaultPainter) DrawControlButton(canvas widget.Canvas, bounds geometry.Rect, control ControlType, state ControlState) {
	if bounds.IsEmpty() {
		return
	}

	// Background on hover/press.
	bg := controlBackground(control, state)
	if !bg.IsTransparent() {
		canvas.DrawRect(bounds, bg)
	}

	// Icon color.
	fg := defaultControlFg
	if control == ControlClose && state.Hovered {
		fg = widget.ColorWhite
	}

	// Draw icon glyph.
	cx := bounds.Min.X + bounds.Width()/2
	cy := bounds.Min.Y + bounds.Height()/2

	switch control {
	case ControlMinimize:
		// Horizontal line.
		canvas.DrawLine(
			geometry.Pt(cx-defaultIconHalf, cy),
			geometry.Pt(cx+defaultIconHalf, cy),
			fg, defaultIconStroke,
		)
	case ControlMaximize:
		// Square outline.
		r := geometry.NewRect(cx-defaultIconHalf, cy-defaultIconHalf, defaultIconSize, defaultIconSize)
		canvas.StrokeRect(r, fg, defaultIconStroke)
	case ControlRestore:
		// Two overlapping squares (smaller).
		offset := float32(2)
		half := defaultIconHalf - 1
		// Back square (offset up-right).
		back := geometry.NewRect(cx-half+offset, cy-half-offset, half*2, half*2)
		canvas.StrokeRect(back, fg, defaultIconStroke)
		// Front square.
		front := geometry.NewRect(cx-half, cy-half, half*2, half*2)
		canvas.DrawRect(front, defaultBarBg) // Fill to cover back square overlap.
		canvas.StrokeRect(front, fg, defaultIconStroke)
	case ControlClose:
		// X shape via two lines.
		canvas.DrawLine(
			geometry.Pt(cx-defaultIconHalf, cy-defaultIconHalf),
			geometry.Pt(cx+defaultIconHalf, cy+defaultIconHalf),
			fg, defaultCloseStroke,
		)
		canvas.DrawLine(
			geometry.Pt(cx+defaultIconHalf, cy-defaultIconHalf),
			geometry.Pt(cx-defaultIconHalf, cy+defaultIconHalf),
			fg, defaultCloseStroke,
		)
	}
}

// Default colors for DefaultPainter.
var (
	defaultBarBg          = widget.Hex(0x2B2D30)
	defaultControlFg      = widget.Hex(0xDFE1E5)
	defaultControlHoverBg = widget.RGBA(1, 1, 1, 0.10)
	defaultControlPressBg = widget.RGBA(1, 1, 1, 0.06)
	defaultCloseHoverBg   = widget.Hex(0xC42B1C)
	defaultClosePressBg   = widget.Hex(0xB22A1A)
)

// Default icon drawing constants.
const (
	defaultIconSize    float32 = 10
	defaultIconHalf    float32 = 5
	defaultIconStroke  float32 = 1
	defaultCloseStroke float32 = 1.5
)

// controlBackground returns the background color for a control button based on
// its type and interaction state.
func controlBackground(control ControlType, state ControlState) widget.Color {
	if control == ControlClose {
		return closeControlBackground(state)
	}
	return normalControlBackground(state)
}

// closeControlBackground returns the background color for the close button.
func closeControlBackground(state ControlState) widget.Color {
	if state.Pressed {
		return defaultClosePressBg
	}
	if state.Hovered {
		return defaultCloseHoverBg
	}
	return widget.ColorTransparent
}

// normalControlBackground returns the background color for non-close controls.
func normalControlBackground(state ControlState) widget.Color {
	if state.Pressed {
		return defaultControlPressBg
	}
	if state.Hovered {
		return defaultControlHoverBg
	}
	return widget.ColorTransparent
}

// Compile-time check that DefaultPainter implements Painter.
var _ Painter = DefaultPainter{}
