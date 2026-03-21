package devtools

import (
	"github.com/gogpu/ui/core/stripe"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/widget"
)

// StripePainter renders stripe toolbars using DevTools design tokens.
// DevTools stripes use a Surface background with a 1px border, and
// SquareStripeButton styling: 40x40 buttons, 20px icons, 12px corner
// radius on hover, matching JetBrains IDE tool window strip styling.
//
// If Theme is nil, StripePainter falls back to the default DevTools dark palette.
type StripePainter struct {
	Theme *Theme // nil uses default DevTools dark fallback
}

// resolveColors returns DevTools-derived colors for stripe painting.
func (p StripePainter) resolveColors() dtStripeColors {
	if p.Theme == nil {
		return dtDefaultStripeColors
	}
	cs := p.Theme.Colors
	return dtStripeColors{
		Background: cs.Surface,
		Border:     cs.Border,
		Foreground: cs.OnSurfaceSecondary,
		HoverBg:    cs.ControlHover,
		PressedBg:  cs.ControlFill,
		ActiveBg:   cs.Selection,
		ActiveFg:   cs.OnSurface,
	}
}

// PaintBackground renders the stripe background with a 1px border.
func (p StripePainter) PaintBackground(canvas widget.Canvas, bounds geometry.Rect) {
	if bounds.IsEmpty() {
		return
	}
	colors := p.resolveColors()
	canvas.DrawRect(bounds, colors.Background)

	// 1px right border for left stripe, left border for right stripe.
	borderRect := geometry.NewRect(
		bounds.Max.X-dtStripeBorderWidth,
		bounds.Min.Y,
		dtStripeBorderWidth,
		bounds.Height(),
	)
	canvas.DrawRect(borderRect, colors.Border)
}

// PaintButton renders a stripe button with icon and optional label.
func (p StripePainter) PaintButton(canvas widget.Canvas, state stripe.ButtonPaintState) {
	if state.Bounds.IsEmpty() {
		return
	}
	colors := p.resolveColors()

	// Background on hover/press/active.
	bg := widget.ColorTransparent
	switch {
	case state.Pressed:
		bg = colors.PressedBg
	case state.Hovered:
		bg = colors.HoverBg
	case state.Active:
		bg = colors.ActiveBg
	}
	if !bg.IsTransparent() {
		canvas.DrawRoundRect(state.Bounds, bg, dtStripeButtonRadius)
	}

	// Icon color.
	fg := colors.Foreground
	if state.Active {
		fg = colors.ActiveFg
	}

	// Draw icon centered horizontally.
	iconBounds := dtStripeIconBounds(state.Bounds, state.ShowLabel)
	icon.Draw(canvas, state.Icon, iconBounds, fg)

	// Draw label below icon if enabled.
	if state.ShowLabel && state.Label != "" {
		textBounds := dtStripeTextBounds(state.Bounds, iconBounds)
		canvas.DrawText(state.Label, textBounds, dtStripeLabelFontSize, fg, false, widget.TextAlignCenter)
	}
}

// dtStripeColors holds resolved DevTools color roles for stripe painting.
type dtStripeColors struct {
	Background widget.Color
	Border     widget.Color
	Foreground widget.Color
	HoverBg    widget.Color
	PressedBg  widget.Color
	ActiveBg   widget.Color
	ActiveFg   widget.Color
}

// dtDefaultStripeColors holds default DevTools dark fallback colors.
// Based on JetBrains StripeButtonUi.kt dark theme values.
var dtDefaultStripeColors = dtStripeColors{
	Background: widget.Hex(0x2B2D30), // Surface
	Border:     widget.Hex(0x393B40), // Gray3
	Foreground: widget.Hex(0x9DA0A8), // Gray9 (OnSurfaceSecondary)
	HoverBg:    widget.Hex(0x43454A), // Gray4 (ControlHover)
	PressedBg:  widget.Hex(0x393B40), // Gray3 (ControlFill)
	ActiveBg:   widget.Hex(0x35373B), // Selection
	ActiveFg:   widget.Hex(0xDFE1E5), // Gray12 (OnSurface)
}

// dtStripeIconBounds calculates icon bounds within a stripe button.
func dtStripeIconBounds(btnBounds geometry.Rect, showLabel bool) geometry.Rect {
	if showLabel {
		centerX := btnBounds.Min.X + btnBounds.Width()/2
		y := btnBounds.Min.Y + dtStripeIconPaddingLabel
		return geometry.NewRect(
			centerX-dtStripeIconSize/2, y,
			dtStripeIconSize, dtStripeIconSize,
		)
	}
	centerX := btnBounds.Min.X + btnBounds.Width()/2
	centerY := btnBounds.Min.Y + btnBounds.Height()/2
	return geometry.NewRect(
		centerX-dtStripeIconSize/2, centerY-dtStripeIconSize/2,
		dtStripeIconSize, dtStripeIconSize,
	)
}

// dtStripeTextBounds calculates text bounds below the icon.
func dtStripeTextBounds(btnBounds geometry.Rect, iconRect geometry.Rect) geometry.Rect {
	y := iconRect.Max.Y + dtStripeIconTextGap
	return geometry.NewRect(
		btnBounds.Min.X+dtStripeTextPadding,
		y,
		btnBounds.Width()-dtStripeTextPadding*2,
		btnBounds.Max.Y-y,
	)
}

// DevTools stripe drawing constants matching JetBrains IDE.
const (
	dtStripeIconSize         float32 = 20
	dtStripeIconPaddingLabel float32 = 4
	dtStripeIconTextGap      float32 = 3
	dtStripeLabelFontSize    float32 = 11
	dtStripeTextPadding      float32 = 6
	dtStripeButtonRadius     float32 = 4
	dtStripeBorderWidth      float32 = 1
)

// Compile-time check that StripePainter implements Painter.
var _ stripe.Painter = StripePainter{}
