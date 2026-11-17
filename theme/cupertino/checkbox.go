package cupertino

import (
	"github.com/gogpu/ui/core/checkbox"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// CheckboxPainter renders checkboxes the way macOS draws them (AppKit
// checkboxes, and UIKit's checkbox style on Mac Catalyst): a 14pt rounded
// square with a thin border that fills with the accent color and shows a
// check mark when checked, or a dash in the mixed (indeterminate) state.
//
// Like the native control, the checkbox has no hover effect. Pressing
// shades the box, and keyboard focus draws the macOS focus ring around it.
// The light or dark appearance follows [Theme.IsDark].
//
// If Theme is nil, CheckboxPainter uses the light appearance with the
// default system blue accent.
type CheckboxPainter struct {
	Theme *Theme // nil uses the light system blue appearance
}

// cupCheckboxPalette holds the colors of every checkbox state.
type cupCheckboxPalette struct {
	fill           widget.Color // unchecked box
	border         widget.Color // unchecked box outline
	checkedFill    widget.Color // checked and mixed box
	mark           widget.Color // check mark and dash
	pressOverlay   widget.Color // shade over the pressed box
	label          widget.Color
	focusRing      widget.Color
	disabledFill   widget.Color
	disabledBorder widget.Color
	disabledMark   widget.Color
	disabledLabel  widget.Color
}

// cupDefaultCheckboxPalette is the palette used when the painter has no Theme.
var cupDefaultCheckboxPalette = func() cupCheckboxPalette {
	cs := lightColors(systemBlue)
	return cupCheckboxPaletteFor(&cs, false)
}()

// cupCheckboxPaletteFor builds the checkbox palette for a color scheme.
// The fills, borders and disabled colors are the macOS values Flutter's
// CupertinoCheckbox uses; the dark unchecked fill flattens the native
// vertical gradient (white at 14% to 29%) to its middle.
func cupCheckboxPaletteFor(cs *ColorScheme, dark bool) cupCheckboxPalette {
	pal := cupCheckboxPalette{
		fill:          widget.ColorWhite,
		border:        widget.Hex(0xD1D1D6),
		checkedFill:   cs.Accent,
		mark:          cs.OnAccent,
		pressOverlay:  widget.ColorBlack.WithAlpha(cupCbPressedOpacity),
		label:         cs.Label,
		focusRing:     cupFocusColor(cs.Accent),
		disabledFill:  widget.ColorWhite.WithAlpha(0.5),
		disabledMark:  widget.ColorBlack.WithAlpha(0.25),
		disabledLabel: cs.TertiaryLabel,
	}
	if dark {
		pal.fill = widget.ColorWhite.WithAlpha(0.21)
		pal.border = widget.RGBA(0.5, 0.5, 0.5, 0.2)
		pal.mark = widget.Hex(0xDEE8F8)
		pal.pressOverlay = widget.ColorWhite.WithAlpha(cupCbPressedOpacity)
		pal.disabledFill = widget.ColorWhite.WithAlpha(0.11)
		pal.disabledMark = widget.ColorWhite.WithAlpha(0.25)
	}
	pal.disabledBorder = pal.border.WithAlpha(pal.border.A * 0.5)
	return pal
}

// palette returns the checkbox palette derived from the painter's Theme.
func (p CheckboxPainter) palette() cupCheckboxPalette {
	if p.Theme == nil {
		return cupDefaultCheckboxPalette
	}
	return cupCheckboxPaletteFor(&p.Theme.Colors, p.Theme.IsDark())
}

// override replaces palette colors with the non-zero colors of s.
func (pal *cupCheckboxPalette) override(s checkbox.CheckboxColorScheme) {
	for _, o := range [...]struct {
		dst *widget.Color
		src widget.Color
	}{
		{&pal.checkedFill, s.CheckedBg},
		{&pal.mark, s.CheckedFg},
		{&pal.border, s.UncheckedBorder},
		{&pal.label, s.LabelColor},
		{&pal.disabledFill, s.DisabledBg},
		{&pal.disabledMark, s.DisabledFg},
		{&pal.disabledLabel, s.DisabledFg},
		{&pal.focusRing, s.FocusRing},
	} {
		if o.src != (widget.Color{}) {
			*o.dst = o.src
		}
	}
}

// PaintCheckbox renders a checkbox as a macOS checkbox.
func (p CheckboxPainter) PaintCheckbox(canvas widget.Canvas, state checkbox.PaintState) {
	if state.Bounds.IsEmpty() {
		return
	}

	pal := p.palette()
	if state.ColorScheme != (checkbox.CheckboxColorScheme{}) {
		pal.override(state.ColorScheme)
	}
	if state.Background != nil {
		pal.checkedFill = *state.Background
	}

	box := cupCheckboxBoxRect(state.Bounds)
	on := state.Checked || state.Indeterminate
	cupPaintCheckboxBox(canvas, box, on, state, pal)

	if on {
		mark := pal.mark
		if state.Disabled {
			mark = pal.disabledMark
		}
		if state.Indeterminate {
			cupDrawCheckboxStroke(canvas, mark,
				cupCheckboxPoint(box, cupCbDashStartX, cupCbDashY),
				cupCheckboxPoint(box, cupCbDashEndX, cupCbDashY),
			)
		} else {
			cupDrawCheckboxStroke(canvas, mark,
				cupCheckboxPoint(box, cupCbCheckP1X, cupCbCheckP1Y),
				cupCheckboxPoint(box, cupCbCheckP2X, cupCbCheckP2Y),
				cupCheckboxPoint(box, cupCbCheckP3X, cupCbCheckP3Y),
			)
		}
	}

	if state.Label != "" {
		fg := pal.label
		if state.Disabled {
			fg = pal.disabledLabel
		}
		canvas.DrawText(state.Label, cupCheckboxLabelBounds(state.Bounds), cupCbFontSize, fg, false, cupCbTextAlignLeft)
	}

	if state.Focused && !state.Disabled {
		canvas.StrokeRoundRect(box.Expand(cupCbFocusRingOffset), pal.focusRing,
			cupCbCornerRadius+cupCbFocusRingOffset, cupCbFocusRingStroke)
	}
}

// cupPaintCheckboxBox draws the box fill, its outline, and the press shade.
// A checked box has no outline; a disabled one is outlined in every state.
func cupPaintCheckboxBox(canvas widget.Canvas, box geometry.Rect, on bool, state checkbox.PaintState, pal cupCheckboxPalette) {
	fill, border := pal.fill, pal.border
	switch {
	case state.Disabled:
		fill, border = pal.disabledFill, pal.disabledBorder
	case on:
		fill, border = pal.checkedFill, widget.Color{}
	}

	canvas.DrawRoundRect(box, fill, cupCbCornerRadius)
	if state.Pressed && !state.Disabled {
		canvas.DrawRoundRect(box, pal.pressOverlay, cupCbCornerRadius)
	}
	if border != (widget.Color{}) {
		// Stroke inside the box so the outline stays within its 14pt.
		inset := cupCbBorderWidth / 2
		canvas.StrokeRoundRect(box.Expand(-inset), border, cupCbCornerRadius-inset, cupCbBorderWidth)
	}
}

// cupDrawCheckboxStroke strokes a polyline. Opaque strokes get round caps
// and joins from dots at the vertices; translucent ones skip the dots,
// which would double the alpha where they overlap the lines.
func cupDrawCheckboxStroke(canvas widget.Canvas, color widget.Color, pts ...geometry.Point) {
	for i := 1; i < len(pts); i++ {
		canvas.DrawLine(pts[i-1], pts[i], color, cupCbMarkStroke)
	}
	if !color.IsOpaque() {
		return
	}
	for _, pt := range pts {
		canvas.DrawCircle(pt, cupCbMarkStroke/2, color)
	}
}

// cupCheckboxPoint maps a point given in box fractions to canvas coordinates.
func cupCheckboxPoint(box geometry.Rect, fx, fy float32) geometry.Point {
	return geometry.Pt(box.Min.X+box.Width()*fx, box.Min.Y+box.Height()*fy)
}

// cupCheckboxBoxRect returns the box, vertically centered at the left edge.
func cupCheckboxBoxRect(bounds geometry.Rect) geometry.Rect {
	y := bounds.Min.Y + (bounds.Height()-cupCbBoxSize)/2
	return geometry.NewRect(bounds.Min.X, y, cupCbBoxSize, cupCbBoxSize)
}

// cupCheckboxLabelBounds returns the label area to the right of the box.
func cupCheckboxLabelBounds(bounds geometry.Rect) geometry.Rect {
	labelW := max(bounds.Width()-cupCbBoxSize-cupCbLabelGap, 0)
	return geometry.NewRect(bounds.Min.X+cupCbBoxSize+cupCbLabelGap, bounds.Min.Y, labelW, bounds.Height())
}

// macOS checkbox drawing constants (regular control size).
const (
	cupCbBoxSize        float32 = 14
	cupCbCornerRadius   float32 = 4
	cupCbBorderWidth    float32 = 1
	cupCbLabelGap       float32 = 6
	cupCbFontSize       float32 = 13
	cupCbTextAlignLeft          = widget.TextAlignLeft
	cupCbPressedOpacity float32 = 0.15
	cupCbMarkStroke     float32 = 2

	// The focus ring is centered 1pt outside the box.
	cupCbFocusRingOffset float32 = 1
	cupCbFocusRingStroke float32 = 3.5

	// Check mark vertices, as fractions of the box size.
	cupCbCheckP1X float32 = 0.22
	cupCbCheckP1Y float32 = 0.54
	cupCbCheckP2X float32 = 0.40
	cupCbCheckP2Y float32 = 0.75
	cupCbCheckP3X float32 = 0.78
	cupCbCheckP3Y float32 = 0.25

	// Mixed-state dash, as fractions of the box size.
	cupCbDashStartX float32 = 0.25
	cupCbDashEndX   float32 = 0.75
	cupCbDashY      float32 = 0.5
)

// CheckboxBoxSize returns the macOS checkbox box size.
func (CheckboxPainter) CheckboxBoxSize() float32 { return cupCbBoxSize }

// CheckboxLabelGap returns the macOS gap between box and label.
func (CheckboxPainter) CheckboxLabelGap() float32 { return cupCbLabelGap }

// CheckboxFontSize returns the macOS control font size for the label.
func (CheckboxPainter) CheckboxFontSize() float32 { return cupCbFontSize }

// Compile-time checks.
var (
	_ checkbox.Painter       = CheckboxPainter{}
	_ checkbox.LayoutMetrics = CheckboxPainter{}
)
