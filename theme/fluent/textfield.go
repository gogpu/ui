package fluent

import (
	"time"

	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// TextFieldPainter renders text fields using Fluent Design tokens.
// Fluent text fields feature a subtle bottom accent border on focus
// and clean rectangular shape with small corner radius.
//
// TextFieldPainter implements [textfield.LayoutMetrics] to provide Fluent
// spatial metrics (14px font, 12/8px padding, 1px cursor) used by the
// widget to compute pre-computed PaintState fields.
//
// The caret follows the Windows text caret rather than the accent: a 1 px
// bar (the default SPI_GETCARETWIDTH) in the text color, as WinUI's
// TextControlCaretBrush is, blinking sharply every 530 ms (the default
// GetCaretBlinkTime).
//
// If Theme is nil, TextFieldPainter falls back to the default Fluent Blue palette.
type TextFieldPainter struct {
	Theme *Theme // nil uses default Fluent Blue fallback
}

// ContentPadding returns the Fluent horizontal and vertical padding.
func (TextFieldPainter) ContentPadding() (float32, float32) {
	return flTFContentPaddingH, flTFContentPaddingV
}

// TextFieldFontSize returns the Fluent font size.
func (TextFieldPainter) TextFieldFontSize() float32 { return flTFFontSize }

// TextFieldCursorWidth returns the Fluent cursor width.
func (TextFieldPainter) TextFieldCursorWidth() float32 { return flTFCursorWidth }

// TextFieldCornerRadius returns the Fluent corner radius.
func (TextFieldPainter) TextFieldCornerRadius() float32 { return flTFCornerRadius }

// TextFieldCaretBlink returns the Windows default caret blink: visible and
// hidden for 530 ms each, switching sharply.
func (TextFieldPainter) TextFieldCaretBlink() textfield.CaretBlink { return flTFCaretBlink }

// resolveColors returns the TextFieldColorScheme derived from the painter's Theme.
func (p TextFieldPainter) resolveColors() textfield.TextFieldColorScheme {
	if p.Theme == nil {
		return flDefaultTextFieldColors
	}
	cs := p.Theme.Colors
	return textfield.TextFieldColorScheme{
		Background:  cs.SurfaceTertiary,
		Border:      cs.StrokeDefault,
		FocusBorder: cs.Accent,
		ErrorBorder: cs.Error,
		TextColor:   cs.OnSurface,
		Placeholder: cs.OnSurfaceSecond,
		CursorColor: cs.OnSurface,
		DisabledBg:  cs.FillDisable,
		DisabledFg:  cs.OnSurfaceSecond.WithAlpha(flDisabledAlpha),
		SelectionBg: cs.Accent.WithAlpha(0.2),
		ErrorText:   cs.Error,
	}
}

// PaintTextField renders a text field according to Fluent Design specifications.
// Cursor and selection positions come from pre-computed PaintState fields.
func (p TextFieldPainter) PaintTextField(canvas widget.Canvas, st *textfield.PaintState) {
	if st.Bounds.IsEmpty() {
		return
	}

	colors := p.resolveColors()
	fontSize := st.FontSize
	if fontSize <= 0 {
		fontSize = flTFFontSize
	}

	flPaintTFBackground(canvas, st, colors)
	flPaintTFBorder(canvas, st, colors)
	flPaintTFContent(canvas, st, colors, fontSize)
	flPaintTFCursorFromState(canvas, st, colors)
	flPaintTFError(canvas, st, colors)
}

// flPaintTFBackground draws the text field background.
func flPaintTFBackground(canvas widget.Canvas, st *textfield.PaintState, colors textfield.TextFieldColorScheme) {
	bg := colors.Background
	if st.Disabled {
		bg = colors.DisabledBg
	}
	canvas.DrawRoundRect(st.Bounds, bg, flTFCornerRadius)
}

// flPaintTFBorder draws the text field outline.
func flPaintTFBorder(canvas widget.Canvas, st *textfield.PaintState, colors textfield.TextFieldColorScheme) {
	borderColor := colors.Border
	strokeWidth := flTFBorderWidth

	switch {
	case st.Disabled:
		borderColor = colors.DisabledFg
	case st.HasError:
		borderColor = colors.ErrorBorder
		strokeWidth = flTFFocusBorderWidth
	case st.Focused:
		borderColor = colors.FocusBorder
		strokeWidth = flTFFocusBorderWidth
	case st.Hovered:
		borderColor = colors.TextColor
	}

	canvas.StrokeRoundRect(st.Bounds, borderColor, flTFCornerRadius, strokeWidth)
}

// flPaintTFContent draws the text or placeholder using pre-computed fields.
func flPaintTFContent(canvas widget.Canvas, st *textfield.PaintState, colors textfield.TextFieldColorScheme, fontSize float32) {
	canvas.PushClip(st.ContentRect)
	defer canvas.PopClip()

	if st.DisplayText == "" && !st.Focused {
		color := colors.Placeholder
		if st.Disabled {
			color = colors.DisabledFg
		}
		canvas.DrawText(st.Placeholder, st.ContentRect, fontSize, color, false, flTFTextAlignLeft)
		return
	}

	textColor := colors.TextColor
	if st.Disabled {
		textColor = colors.DisabledFg
	}

	// Selection highlight from pre-computed rect.
	if st.ShowSelection {
		canvas.DrawRect(st.SelectionRect, colors.SelectionBg)
	}

	canvas.DrawText(st.DisplayText, st.TextRect, fontSize, textColor, false, flTFTextAlignLeft)
}

// flPaintTFCursorFromState draws the caret using pre-computed CursorRect.
func flPaintTFCursorFromState(canvas widget.Canvas, st *textfield.PaintState, colors textfield.TextFieldColorScheme) {
	textfield.PaintCursor(canvas, st, colors.CursorColor, 0)
}

// flPaintTFError draws the error message below the text field.
func flPaintTFError(canvas widget.Canvas, st *textfield.PaintState, colors textfield.TextFieldColorScheme) {
	if !st.HasError || st.ErrorMsg == "" {
		return
	}

	errBounds := geometry.Rect{
		Min: geometry.Pt(st.Bounds.Min.X+flTFContentPaddingH, st.Bounds.Max.Y+flTFErrorTopGap),
		Max: geometry.Pt(st.Bounds.Max.X, st.Bounds.Max.Y+flTFErrorTopGap+flTFErrorFontSize+flTFErrorBottomPad),
	}
	canvas.DrawText(st.ErrorMsg, errBounds, flTFErrorFontSize, colors.ErrorText, false, flTFTextAlignLeft)
}

// flDefaultTextFieldColors holds the default Fluent text field color scheme.
var flDefaultTextFieldColors = textfield.TextFieldColorScheme{
	Background:  widget.Hex(0xFAFAFA),
	Border:      widget.RGBA(0, 0, 0, 0.14),
	FocusBorder: DefaultAccentColor,
	ErrorBorder: widget.Hex(0xC42B1C),
	TextColor:   widget.Hex(0x1A1A1A),
	Placeholder: widget.Hex(0x616161),
	CursorColor: widget.Hex(0x1A1A1A), // text color, like TextControlCaretBrush
	DisabledBg:  widget.RGBA(0, 0, 0, 0.04),
	DisabledFg:  widget.RGBA(0.38, 0.38, 0.38, 0.38),
	SelectionBg: DefaultAccentColor.WithAlpha(0.2),
	ErrorText:   widget.Hex(0xC42B1C),
}

// Fluent text field drawing constants.
const (
	flTFCornerRadius     float32 = 4
	flTFBorderWidth      float32 = 1
	flTFFocusBorderWidth float32 = 2
	flTFContentPaddingH  float32 = 12
	flTFContentPaddingV  float32 = 8
	flTFFontSize         float32 = 14
	flTFTextAlignLeft            = widget.TextAlignLeft
	flTFCursorWidth      float32 = 1
	flTFErrorFontSize    float32 = 12
	flTFErrorTopGap      float32 = 4
	flTFErrorBottomPad   float32 = 4
)

// flTFCaretBlinkInterval is the Windows default caret blink time, which
// GetCaretBlinkTime reports when the user has not changed it.
const flTFCaretBlinkInterval = 530 * time.Millisecond

// flTFCaretBlink is the Windows default caret blink.
var flTFCaretBlink = textfield.HardCaretBlink(flTFCaretBlinkInterval)

// Compile-time checks.
var (
	_ textfield.Painter       = TextFieldPainter{}
	_ textfield.LayoutMetrics = TextFieldPainter{}
	_ textfield.CaretBlinker  = TextFieldPainter{}
)
