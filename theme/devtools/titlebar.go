package devtools

import (
	"math"

	"github.com/gogpu/ui/core/titlebar"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/widget"
)

// TitleBarPainter renders title bars using DevTools design tokens.
//
// DevTools title bars use a dark header background (#27282E) that is always
// dark even in light mode, matching JetBrains IDE behavior. Window control
// buttons are 46x40px with no border and Windows-style hover effects.
//
// If Theme is nil, TitleBarPainter falls back to default DevTools dark colors.
type TitleBarPainter struct {
	Theme *Theme // nil uses default DevTools dark fallback
}

// resolveColors returns DevTools-derived colors for title bar painting.
func (p TitleBarPainter) resolveColors() dtTitleBarColors {
	if p.Theme == nil {
		return dtDefaultTitleBarColors
	}
	cs := p.Theme.Colors
	return dtTitleBarColors{
		Background:       cs.HeaderBackground,
		BorderColor:      cs.Border,
		IconColor:        cs.HeaderForeground,
		ControlHoverBg:   cs.ControlHover,
		ControlPressBg:   cs.ControlFill,
		CloseHoverBg:     dtCloseHoverRed,
		ClosePressBg:     dtClosePressDarkRed,
		CloseHoverIconFg: widget.ColorWhite,
	}
}

// PaintBackground renders the dark header background with a 1px bottom border.
func (p TitleBarPainter) PaintBackground(canvas widget.Canvas, bounds geometry.Rect, _ titlebar.BackgroundState) {
	if bounds.IsEmpty() {
		return
	}
	colors := p.resolveColors()

	// Dark header background.
	canvas.DrawRect(bounds, colors.Background)

	// 1px bottom border.
	borderRect := geometry.NewRect(
		bounds.Min.X,
		bounds.Max.Y-dtTitleBarBorderWidth,
		bounds.Width(),
		dtTitleBarBorderWidth,
	)
	canvas.DrawRect(borderRect, colors.BorderColor)
}

// PaintControlButton renders a window control button with Windows-style hover.
func (p TitleBarPainter) PaintControlButton(canvas widget.Canvas, bounds geometry.Rect, control titlebar.ControlType, state titlebar.ControlState) {
	if bounds.IsEmpty() {
		return
	}
	colors := p.resolveColors()

	// Background on hover/press.
	isClose := control == titlebar.ControlClose
	bg := dtControlBackground(isClose, state, colors)
	if !bg.IsTransparent() {
		canvas.DrawRect(bounds, bg)
	}

	// Icon color.
	fg := colors.IconColor
	if isClose && state.Hovered {
		fg = colors.CloseHoverIconFg
	}

	// Draw SVG icon centered within control button bounds.
	var iconData icon.IconData
	switch control {
	case titlebar.ControlMinimize:
		iconData = icon.WindowMinimize
	case titlebar.ControlMaximize:
		iconData = icon.WindowMaximize
	case titlebar.ControlRestore:
		iconData = icon.WindowRestore
	case titlebar.ControlClose:
		iconData = icon.WindowClose
	}
	const controlIconSize float32 = 16
	cx := bounds.Min.X + bounds.Width()/2
	cy := bounds.Min.Y + bounds.Height()/2
	iconBounds := geometry.NewRect(
		float32(math.Round(float64(cx-controlIconSize/2))),
		float32(math.Round(float64(cy-controlIconSize/2))),
		controlIconSize, controlIconSize,
	)
	icon.Draw(canvas, iconData, iconBounds, fg)
}

// dtControlBackground returns the background color for a control button based on
// whether it is the close button and its interaction state.
func dtControlBackground(isClose bool, state titlebar.ControlState, colors dtTitleBarColors) widget.Color {
	if isClose {
		return dtCloseBackground(state, colors)
	}
	return dtNormalBackground(state, colors)
}

// dtCloseBackground returns the background color for the close button.
func dtCloseBackground(state titlebar.ControlState, colors dtTitleBarColors) widget.Color {
	if state.Pressed {
		return colors.ClosePressBg
	}
	if state.Hovered {
		return colors.CloseHoverBg
	}
	return widget.ColorTransparent
}

// dtNormalBackground returns the background color for non-close controls.
func dtNormalBackground(state titlebar.ControlState, colors dtTitleBarColors) widget.Color {
	if state.Pressed {
		return colors.ControlPressBg
	}
	if state.Hovered {
		return colors.ControlHoverBg
	}
	return widget.ColorTransparent
}

// dtTitleBarColors holds resolved DevTools colors for title bar painting.
type dtTitleBarColors struct {
	Background       widget.Color
	BorderColor      widget.Color
	IconColor        widget.Color
	ControlHoverBg   widget.Color
	ControlPressBg   widget.Color
	CloseHoverBg     widget.Color
	ClosePressBg     widget.Color
	CloseHoverIconFg widget.Color
}

// dtDefaultTitleBarColors holds default DevTools dark fallback colors.
var dtDefaultTitleBarColors = dtTitleBarColors{
	Background:       widget.Hex(0x27282E),
	BorderColor:      widget.Hex(0x393B40), // Gray3
	IconColor:        widget.Hex(0xDFE1E5), // Gray12
	ControlHoverBg:   widget.Hex(0x43454A), // Gray4
	ControlPressBg:   widget.Hex(0x393B40), // Gray3
	CloseHoverBg:     dtCloseHoverRed,
	ClosePressBg:     dtClosePressDarkRed,
	CloseHoverIconFg: widget.ColorWhite,
}

// DevTools title bar color constants.
var (
	dtCloseHoverRed     = widget.Hex(0xC42B1C)
	dtClosePressDarkRed = widget.Hex(0xB22A1A)
)

// DevTools title bar drawing constants.
const (
	dtTitleBarBorderWidth float32 = 1
	dtCtrlIconSize        float32 = 10
	dtCtrlIconHalf        float32 = 5
	dtCtrlRestoreHalf     float32 = 4
	dtCtrlIconStroke      float32 = 1
	dtCtrlCloseStroke     float32 = 1.5
)

// Compile-time check that TitleBarPainter implements titlebar.Painter.
var _ titlebar.Painter = TitleBarPainter{}
