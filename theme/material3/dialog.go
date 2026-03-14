package material3

import (
	"github.com/gogpu/ui/core/dialog"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// DialogPainter renders dialogs using Material 3 design tokens.
// It applies the M3 dialog specification: surface tint, headline small
// typography, and text action buttons.
//
// If Theme is nil, DialogPainter falls back to the default M3 purple palette.
type DialogPainter struct {
	Theme *Theme // nil uses default M3 purple fallback
}

// resolveColors returns the DialogColorScheme derived from the painter's Theme.
// If Theme is nil, it returns the default M3 dialog color scheme.
func (p DialogPainter) resolveColors() dialog.DialogColorScheme {
	if p.Theme == nil {
		return m3DefaultDialogColors
	}
	cs := p.Theme.Colors
	return dialog.DialogColorScheme{
		Backdrop: widget.RGBA(0, 0, 0, m3DialogScrimAlpha),
		Surface:  cs.SurfaceContainerHigh,
		Title:    cs.OnSurface,
		Content:  cs.OnSurfaceVariant,
		Border:   widget.ColorTransparent,
		Shadow:   widget.RGBA(0, 0, 0, m3DialogShadowAlpha),
		ActionFg: cs.Primary,
		ActionBg: cs.Primary,
	}
}

// PaintDialog renders a dialog according to Material 3 specifications.
func (p DialogPainter) PaintDialog(canvas widget.Canvas, ps dialog.PaintState) {
	if ps.Bounds.IsEmpty() {
		return
	}

	// Determine the color scheme to use.
	colors := ps.ColorScheme
	if colors == (dialog.DialogColorScheme{}) {
		colors = p.resolveColors()
	}

	// Draw surface with M3 corner radius (28dp for full-screen dialog, 24dp standard).
	canvas.DrawRoundRect(ps.Bounds, colors.Surface, m3DialogRadius)

	// Draw title (headline small: 24px).
	if ps.Title != "" {
		titleBounds := geometry.Rect{
			Min: geometry.Pt(ps.Bounds.Min.X+m3DialogPadding, ps.Bounds.Min.Y+m3DialogPadding),
			Max: geometry.Pt(ps.Bounds.Max.X-m3DialogPadding, ps.Bounds.Min.Y+m3DialogPadding+m3DialogTitleHeight),
		}
		canvas.DrawText(ps.Title, titleBounds, m3DialogTitleFontSize, colors.Title, false, m3DialogTextAlignLeft)
	}

	// Draw action buttons (text buttons, right-aligned).
	p.paintActions(canvas, ps, colors)

	// Focus indicator.
	if ps.Focused {
		ringBounds := ps.Bounds.Expand(m3DialogFocusRingOffset)
		ringRadius := m3DialogRadius + m3DialogFocusRingOffset
		canvas.StrokeRoundRect(ringBounds, colors.ActionFg.WithAlpha(m3DialogFocusRingAlpha), ringRadius, m3DialogFocusRingStroke)
	}
}

// paintActions renders M3-styled action buttons.
func (p DialogPainter) paintActions(canvas widget.Canvas, ps dialog.PaintState, colors dialog.DialogColorScheme) {
	if len(ps.Actions) == 0 {
		return
	}

	// Layout actions from right to left.
	x := ps.Bounds.Max.X - m3DialogPadding
	y := ps.Bounds.Max.Y - m3DialogPadding - m3DialogActionHeight

	for i := len(ps.Actions) - 1; i >= 0; i-- {
		label := ps.Actions[i].Label
		btnWidth := float32(len(label))*m3DialogActionCharWidth + m3DialogActionPaddingX*2
		x -= btnWidth

		btnBounds := geometry.NewRect(x, y, btnWidth, m3DialogActionHeight)
		canvas.DrawText(label, btnBounds, m3DialogActionFontSize, colors.ActionFg, false, m3DialogTextAlignCenter)

		x -= m3DialogActionSpacing
	}
}

// M3 dialog constants.
const (
	m3DialogRadius          float32 = 24
	m3DialogPadding         float32 = 24
	m3DialogTitleHeight     float32 = 28
	m3DialogTitleFontSize   float32 = 24
	m3DialogActionHeight    float32 = 36
	m3DialogActionFontSize  float32 = 14
	m3DialogActionCharWidth float32 = 8
	m3DialogActionPaddingX  float32 = 12
	m3DialogActionSpacing   float32 = 8
	m3DialogTextAlignLeft           = widget.TextAlignLeft
	m3DialogTextAlignCenter         = widget.TextAlignCenter
	m3DialogScrimAlpha      float32 = 0.32
	m3DialogShadowAlpha     float32 = 0.15
	m3DialogFocusRingOffset float32 = 2
	m3DialogFocusRingStroke float32 = 2
	m3DialogFocusRingAlpha  float32 = 0.7
)

// m3DefaultDialogColors holds the default M3 purple color scheme for dialogs.
// Used as a fallback when no Theme is provided.
var m3DefaultDialogColors = dialog.DialogColorScheme{
	Backdrop: widget.RGBA(0, 0, 0, 0.32),
	Surface:  widget.Hex(0xECE6F0),    // M3 surface container high
	Title:    widget.Hex(0x1C1B1F),    // M3 on-surface
	Content:  widget.Hex(0x49454F),    // M3 on-surface-variant
	Border:   widget.ColorTransparent, // M3 dialogs have no border
	Shadow:   widget.RGBA(0, 0, 0, 0.15),
	ActionFg: widget.Hex(0x6750A4), // M3 primary
	ActionBg: widget.Hex(0x6750A4), // M3 primary
}

// Compile-time check that DialogPainter implements Painter.
var _ dialog.Painter = DialogPainter{}
