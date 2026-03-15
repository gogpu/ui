package cupertino

import (
	"github.com/gogpu/ui/core/dialog"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// DialogPainter renders dialogs using Apple HIG design tokens.
// Cupertino dialogs follow the iOS alert/sheet style: rounded corners,
// grouped background, centered title, and stacked or horizontal action buttons.
//
// If Theme is nil, DialogPainter falls back to the default system blue palette.
type DialogPainter struct {
	Theme *Theme // nil uses default system blue fallback
}

// resolveColors returns the DialogColorScheme derived from the painter's Theme.
func (p DialogPainter) resolveColors() dialog.DialogColorScheme {
	if p.Theme == nil {
		return cupDefaultDialogColors
	}
	cs := p.Theme.Colors
	return dialog.DialogColorScheme{
		Backdrop: widget.RGBA(0, 0, 0, cupDlgBackdropAlpha),
		Surface:  cs.SecondarySystemBackground,
		Title:    cs.Label,
		Content:  cs.SecondaryLabel,
		Border:   widget.ColorTransparent,
		Shadow:   widget.RGBA(0, 0, 0, cupDlgShadowAlpha),
		ActionFg: cs.Accent,
		ActionBg: cs.Accent,
	}
}

// PaintDialog renders a dialog according to Apple HIG specifications.
func (p DialogPainter) PaintDialog(canvas widget.Canvas, ps dialog.PaintState) {
	if ps.Bounds.IsEmpty() {
		return
	}

	colors := ps.ColorScheme
	if colors == (dialog.DialogColorScheme{}) {
		colors = p.resolveColors()
	}

	// iOS alert: heavy rounded corners, no border.
	canvas.DrawRoundRect(ps.Bounds, colors.Surface, cupDlgRadius)

	// Title (centered, bold).
	if ps.Title != "" {
		titleBounds := geometry.Rect{
			Min: geometry.Pt(ps.Bounds.Min.X+cupDlgPadding, ps.Bounds.Min.Y+cupDlgPadding),
			Max: geometry.Pt(ps.Bounds.Max.X-cupDlgPadding, ps.Bounds.Min.Y+cupDlgPadding+cupDlgTitleHeight),
		}
		canvas.DrawText(ps.Title, titleBounds, cupDlgTitleFontSize, colors.Title, true, cupDlgTextAlignCenter)
	}

	// Action buttons (right-aligned, accent color text).
	p.paintActions(canvas, ps, colors)

	// Focus indicator.
	if ps.Focused {
		ringBounds := ps.Bounds.Expand(cupDlgFocusRingOffset)
		ringRadius := cupDlgRadius + cupDlgFocusRingOffset
		canvas.StrokeRoundRect(ringBounds, colors.ActionFg.WithAlpha(cupDlgFocusRingAlpha), ringRadius, cupDlgFocusRingStroke)
	}
}

// paintActions renders Cupertino-styled action buttons.
func (p DialogPainter) paintActions(canvas widget.Canvas, ps dialog.PaintState, colors dialog.DialogColorScheme) {
	if len(ps.Actions) == 0 {
		return
	}

	x := ps.Bounds.Max.X - cupDlgPadding
	y := ps.Bounds.Max.Y - cupDlgPadding - cupDlgActionHeight

	for i := len(ps.Actions) - 1; i >= 0; i-- {
		label := ps.Actions[i].Label
		btnWidth := float32(len(label))*cupDlgActionCharWidth + cupDlgActionPaddingX*2
		x -= btnWidth

		btnBounds := geometry.NewRect(x, y, btnWidth, cupDlgActionHeight)
		canvas.DrawText(label, btnBounds, cupDlgActionFontSize, colors.ActionFg, false, cupDlgTextAlignCenter)

		x -= cupDlgActionSpacing
	}
}

// cupDefaultDialogColors holds the default Cupertino dialog color scheme.
var cupDefaultDialogColors = dialog.DialogColorScheme{
	Backdrop: widget.RGBA(0, 0, 0, 0.4),
	Surface:  widget.Hex(0xF2F2F7),
	Title:    widget.RGBA(0.0, 0.0, 0.0, 1.0),
	Content:  widget.RGBA(0.235, 0.235, 0.263, 0.6),
	Border:   widget.ColorTransparent,
	Shadow:   widget.RGBA(0, 0, 0, 0.15),
	ActionFg: systemBlue,
	ActionBg: systemBlue,
}

// Cupertino dialog drawing constants.
const (
	cupDlgRadius          float32 = 14
	cupDlgPadding         float32 = 20
	cupDlgTitleHeight     float32 = 24
	cupDlgTitleFontSize   float32 = 17
	cupDlgActionHeight    float32 = 44
	cupDlgActionFontSize  float32 = 17
	cupDlgActionCharWidth float32 = 9
	cupDlgActionPaddingX  float32 = 12
	cupDlgActionSpacing   float32 = 8
	cupDlgTextAlignCenter         = widget.TextAlignCenter
	cupDlgBackdropAlpha   float32 = 0.4
	cupDlgShadowAlpha     float32 = 0.15
	cupDlgFocusRingOffset float32 = 3
	cupDlgFocusRingStroke float32 = 2.5
	cupDlgFocusRingAlpha  float32 = 0.6
)

// Compile-time check that DialogPainter implements Painter.
var _ dialog.Painter = DialogPainter{}
