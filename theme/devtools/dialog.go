package devtools

import (
	"github.com/gogpu/ui/core/dialog"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// DialogPainter renders dialogs using DevTools design tokens.
// DevTools dialogs use Surface (#2B2D30) background with 8px radius,
// subtle shadow, 1px border, and 16px bold title — matching JetBrains
// IDE modal dialog styling.
//
// If Theme is nil, DialogPainter falls back to the default DevTools dark palette.
type DialogPainter struct {
	Theme *Theme // nil uses default DevTools dark fallback
}

// resolveColors returns the DialogColorScheme derived from the painter's Theme.
func (p DialogPainter) resolveColors() dialog.DialogColorScheme {
	if p.Theme == nil {
		return dtDefaultDialogColors
	}
	cs := p.Theme.Colors
	return dialog.DialogColorScheme{
		Backdrop: cs.Backdrop,
		Surface:  cs.Surface,
		Title:    cs.OnSurface,
		Content:  cs.OnSurfaceSecondary,
		Border:   cs.Border,
		Shadow:   cs.Shadow,
		ActionFg: cs.Primary,
		ActionBg: cs.Primary,
	}
}

// PaintDialog renders a dialog according to DevTools design specifications.
func (p DialogPainter) PaintDialog(canvas widget.Canvas, ps dialog.PaintState) {
	if ps.Bounds.IsEmpty() {
		return
	}

	colors := ps.ColorScheme
	if colors == (dialog.DialogColorScheme{}) {
		colors = p.resolveColors()
	}

	// Surface with border.
	canvas.DrawRoundRect(ps.Bounds, colors.Surface, dtDialogRadius)
	canvas.StrokeRoundRect(ps.Bounds, colors.Border, dtDialogRadius, dtDialogBorderWidth)

	// Title.
	if ps.Title != "" {
		titleBounds := geometry.Rect{
			Min: geometry.Pt(ps.Bounds.Min.X+dtDialogPadding, ps.Bounds.Min.Y+dtDialogPadding),
			Max: geometry.Pt(ps.Bounds.Max.X-dtDialogPadding, ps.Bounds.Min.Y+dtDialogPadding+dtDialogTitleHeight),
		}
		canvas.DrawText(ps.Title, titleBounds, dtDialogTitleFontSize, colors.Title, true, dtDialogTextAlignLeft)
	}

	// Action buttons.
	p.paintActions(canvas, ps, colors)

	// Focus indicator.
	if ps.Focused {
		dtDrawFocusRing(canvas, ps.Bounds, dtDialogRadius, colors.ActionFg)
	}
}

// paintActions renders DevTools-styled action buttons.
func (p DialogPainter) paintActions(canvas widget.Canvas, ps dialog.PaintState, colors dialog.DialogColorScheme) {
	if len(ps.Actions) == 0 {
		return
	}

	x := ps.Bounds.Max.X - dtDialogPadding
	y := ps.Bounds.Max.Y - dtDialogPadding - dtDialogActionHeight

	for i := len(ps.Actions) - 1; i >= 0; i-- {
		label := ps.Actions[i].Label
		btnWidth := float32(len(label))*dtDialogActionCharWidth + dtDialogActionPaddingX*2
		x -= btnWidth

		btnBounds := geometry.NewRect(x, y, btnWidth, dtDialogActionHeight)
		canvas.DrawText(label, btnBounds, dtDialogActionFontSize, colors.ActionFg, false, dtDialogTextAlignCenter)

		x -= dtDialogActionSpacing
	}
}

// DevTools dialog constants.
const (
	dtDialogRadius          float32 = 8
	dtDialogBorderWidth     float32 = 1
	dtDialogPadding         float32 = 16
	dtDialogTitleHeight     float32 = 24
	dtDialogTitleFontSize   float32 = 16
	dtDialogActionHeight    float32 = 28
	dtDialogActionFontSize  float32 = 13
	dtDialogActionCharWidth float32 = 7
	dtDialogActionPaddingX  float32 = 12
	dtDialogActionSpacing   float32 = 8
	dtDialogTextAlignLeft           = widget.TextAlignLeft
	dtDialogTextAlignCenter         = widget.TextAlignCenter
)

// dtDefaultDialogColors holds the default DevTools dark dialog color scheme.
var dtDefaultDialogColors = dialog.DialogColorScheme{
	Backdrop: widget.RGBA(0, 0, 0, 0.50),
	Surface:  widget.Hex(0x2B2D30), // Gray2 (Surface)
	Title:    widget.Hex(0xDFE1E5), // Gray12
	Content:  widget.Hex(0x9DA0A8), // Gray9
	Border:   widget.Hex(0x393B40), // Gray3 (Border)
	Shadow:   widget.RGBA(0, 0, 0, 0.50),
	ActionFg: DefaultAccentColor,
	ActionBg: DefaultAccentColor,
}

// Compile-time check that DialogPainter implements Painter.
var _ dialog.Painter = DialogPainter{}
