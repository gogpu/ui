package fluent

import (
	"github.com/gogpu/ui/core/dialog"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// DialogPainter renders dialogs using Fluent Design tokens.
// Fluent dialogs have a clean surface with subtle border and smaller
// corner radius compared to Material 3.
//
// If Theme is nil, DialogPainter falls back to the default Fluent Blue palette.
type DialogPainter struct {
	Theme *Theme // nil uses default Fluent Blue fallback
}

// resolveColors returns the DialogColorScheme derived from the painter's Theme.
func (p DialogPainter) resolveColors() dialog.DialogColorScheme {
	if p.Theme == nil {
		return flDefaultDialogColors
	}
	cs := p.Theme.Colors
	return dialog.DialogColorScheme{
		Backdrop: cs.Backdrop,
		Surface:  cs.Surface,
		Title:    cs.OnSurface,
		Content:  cs.OnSurfaceSecond,
		Border:   cs.StrokeDefault,
		Shadow:   cs.Shadow,
		ActionFg: cs.Accent,
		ActionBg: cs.Accent,
	}
}

// PaintDialog renders a dialog according to Fluent Design specifications.
func (p DialogPainter) PaintDialog(canvas widget.Canvas, ps dialog.PaintState) {
	if ps.Bounds.IsEmpty() {
		return
	}

	colors := ps.ColorScheme
	if colors == (dialog.DialogColorScheme{}) {
		colors = p.resolveColors()
	}

	// Surface with border.
	canvas.DrawRoundRect(ps.Bounds, colors.Surface, flDialogRadius)
	canvas.StrokeRoundRect(ps.Bounds, colors.Border, flDialogRadius, flDialogBorderWidth)

	// Title.
	if ps.Title != "" {
		titleBounds := geometry.Rect{
			Min: geometry.Pt(ps.Bounds.Min.X+flDialogPadding, ps.Bounds.Min.Y+flDialogPadding),
			Max: geometry.Pt(ps.Bounds.Max.X-flDialogPadding, ps.Bounds.Min.Y+flDialogPadding+flDialogTitleHeight),
		}
		canvas.DrawText(ps.Title, titleBounds, flDialogTitleFontSize, colors.Title, true, flDialogTextAlignLeft)
	}

	// Action buttons.
	p.paintActions(canvas, ps, colors)

	// Focus indicator.
	if ps.Focused {
		flDrawFocusRing(canvas, ps.Bounds, flDialogRadius, colors.ActionFg)
	}
}

// paintActions renders Fluent-styled action buttons.
func (p DialogPainter) paintActions(canvas widget.Canvas, ps dialog.PaintState, colors dialog.DialogColorScheme) {
	if len(ps.Actions) == 0 {
		return
	}

	x := ps.Bounds.Max.X - flDialogPadding
	y := ps.Bounds.Max.Y - flDialogPadding - flDialogActionHeight

	for i := len(ps.Actions) - 1; i >= 0; i-- {
		label := ps.Actions[i].Label
		btnWidth := float32(len(label))*flDialogActionCharWidth + flDialogActionPaddingX*2
		x -= btnWidth

		btnBounds := geometry.NewRect(x, y, btnWidth, flDialogActionHeight)
		canvas.DrawText(label, btnBounds, flDialogActionFontSize, colors.ActionFg, false, flDialogTextAlignCenter)

		x -= flDialogActionSpacing
	}
}

// Fluent dialog constants.
const (
	flDialogRadius          float32 = 8
	flDialogBorderWidth     float32 = 1
	flDialogPadding         float32 = 24
	flDialogTitleHeight     float32 = 28
	flDialogTitleFontSize   float32 = 20
	flDialogActionHeight    float32 = 32
	flDialogActionFontSize  float32 = 14
	flDialogActionCharWidth float32 = 8
	flDialogActionPaddingX  float32 = 12
	flDialogActionSpacing   float32 = 8
	flDialogTextAlignLeft           = widget.TextAlignLeft
	flDialogTextAlignCenter         = widget.TextAlignCenter
)

// flDefaultDialogColors holds the default Fluent dialog color scheme.
var flDefaultDialogColors = dialog.DialogColorScheme{
	Backdrop: widget.RGBA(0, 0, 0, 0.30),
	Surface:  widget.ColorWhite,
	Title:    widget.Hex(0x1A1A1A),
	Content:  widget.Hex(0x616161),
	Border:   widget.RGBA(0, 0, 0, 0.14),
	Shadow:   widget.RGBA(0, 0, 0, 0.14),
	ActionFg: DefaultAccentColor,
	ActionBg: DefaultAccentColor,
}

// Compile-time check that DialogPainter implements Painter.
var _ dialog.Painter = DialogPainter{}
