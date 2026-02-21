package textfield

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// Painter draws the visual representation of a text field.
// Each design system (Material 3, Fluent, Cupertino) provides its own
// Painter implementation to render the text field in its visual style.
//
// If no Painter is set, the text field uses [DefaultPainter].
type Painter interface {
	PaintTextField(canvas widget.Canvas, state PaintState)
}

// PaintState provides the current text field state to the painter.
type PaintState struct {
	Text        string
	Placeholder string
	Focused     bool
	Hovered     bool
	Disabled    bool
	HasError    bool
	ErrorMsg    string
	CursorPos   int
	SelectStart int
	SelectEnd   int
	InputType   InputType
	Bounds      geometry.Rect
}

// TextFieldColorScheme provides theme-derived colors for text field painting.
// Zero value means the painter should use its built-in defaults.
type TextFieldColorScheme struct {
	Background  widget.Color
	Border      widget.Color
	FocusBorder widget.Color
	ErrorBorder widget.Color
	TextColor   widget.Color
	Placeholder widget.Color
	CursorColor widget.Color
	DisabledBg  widget.Color
	DisabledFg  widget.Color
	SelectionBg widget.Color
	ErrorText   widget.Color
}

// DefaultPainter provides a minimal fallback painter with no design system styling.
// It draws a simple outlined text field suitable for testing and prototyping.
type DefaultPainter struct{}

// PaintTextField renders a minimal text field with gray outline.
// If the state has a non-zero ColorScheme (via the M3 painter), use those colors.
func (p DefaultPainter) PaintTextField(canvas widget.Canvas, st PaintState) {
	if st.Bounds.IsEmpty() {
		return
	}

	paintBackground(canvas, st)
	paintBorder(canvas, st)
	paintContent(canvas, st)
	paintCursor(canvas, st)
}

// paintBackground fills the text field background.
func paintBackground(canvas widget.Canvas, st PaintState) {
	bg := defaultBg
	if st.Disabled {
		bg = defaultDisabledBg
	}
	canvas.DrawRoundRect(st.Bounds, bg, defaultCornerRadius)
}

// paintBorder draws the text field outline.
func paintBorder(canvas widget.Canvas, st PaintState) {
	borderColor := defaultBorderColor
	strokeWidth := defaultBorderWidth

	switch {
	case st.Disabled:
		borderColor = defaultDisabledBorderColor
	case st.HasError:
		borderColor = defaultErrorColor
	case st.Focused:
		borderColor = defaultFocusBorderColor
		strokeWidth = defaultFocusBorderWidth
	case st.Hovered:
		borderColor = defaultHoverBorderColor
	}

	canvas.StrokeRoundRect(st.Bounds, borderColor, defaultCornerRadius, strokeWidth)
}

// paintContent renders either the placeholder or the displayed text.
func paintContent(canvas widget.Canvas, st PaintState) {
	contentBounds := contentRect(st.Bounds)

	// Clip to content area.
	canvas.PushClip(contentBounds)
	defer canvas.PopClip()

	displayText := st.Text
	if st.InputType == TypePassword {
		displayText = maskText(len([]rune(st.Text)))
	}

	if displayText == "" && !st.Focused {
		// Draw placeholder.
		color := defaultPlaceholderColor
		if st.Disabled {
			color = defaultDisabledFg
		}
		canvas.DrawText(st.Placeholder, contentBounds, defaultFontSize, color, false, textAlignLeft)
		return
	}

	// Draw text.
	textColor := defaultTextColor
	if st.Disabled {
		textColor = defaultDisabledFg
	}

	// Draw selection background if present.
	if st.SelectStart != st.SelectEnd {
		paintSelection(canvas, contentBounds, st)
	}

	canvas.DrawText(displayText, contentBounds, defaultFontSize, textColor, false, textAlignLeft)
}

// paintSelection draws the selection highlight rectangle.
func paintSelection(canvas widget.Canvas, contentBounds geometry.Rect, st PaintState) {
	selStart := st.SelectStart
	selEnd := st.SelectEnd
	if selStart > selEnd {
		selStart, selEnd = selEnd, selStart
	}

	charWidth := defaultFontSize * charWidthRatio
	x1 := contentBounds.Min.X + float32(selStart)*charWidth
	x2 := contentBounds.Min.X + float32(selEnd)*charWidth

	// Clamp to content bounds.
	if x2 > contentBounds.Max.X {
		x2 = contentBounds.Max.X
	}

	selRect := geometry.Rect{
		Min: geometry.Pt(x1, contentBounds.Min.Y),
		Max: geometry.Pt(x2, contentBounds.Max.Y),
	}
	canvas.DrawRect(selRect, defaultSelectionBg)
}

// paintCursor draws the blinking caret when the field is focused.
func paintCursor(canvas widget.Canvas, st PaintState) {
	if !st.Focused || st.Disabled {
		return
	}
	// No cursor when there is a selection.
	if st.SelectStart != st.SelectEnd {
		return
	}

	content := contentRect(st.Bounds)
	charWidth := defaultFontSize * charWidthRatio
	cursorX := content.Min.X + float32(st.CursorPos)*charWidth

	// Clamp cursor to content area.
	if cursorX > content.Max.X {
		cursorX = content.Max.X
	}

	top := geometry.Pt(cursorX, content.Min.Y+cursorTopPadding)
	bottom := geometry.Pt(cursorX, content.Max.Y-cursorBottomPadding)
	canvas.DrawLine(top, bottom, defaultCursorColor, cursorWidth)
}

// contentRect returns the inner content area after padding.
func contentRect(bounds geometry.Rect) geometry.Rect {
	return geometry.Rect{
		Min: geometry.Pt(bounds.Min.X+contentPaddingH, bounds.Min.Y+contentPaddingV),
		Max: geometry.Pt(bounds.Max.X-contentPaddingH, bounds.Max.Y-contentPaddingV),
	}
}

// maskText returns a string of dots with the given length.
func maskText(length int) string {
	runes := make([]rune, length)
	for i := range runes {
		runes[i] = passwordMaskChar
	}
	return string(runes)
}

// Painting constants.
const (
	defaultFontSize         float32 = 14
	defaultCornerRadius     float32 = 4
	defaultBorderWidth      float32 = 1
	defaultFocusBorderWidth float32 = 2
	contentPaddingH         float32 = 12
	contentPaddingV         float32 = 8
	charWidthRatio          float32 = 0.55
	textAlignLeft           float32 = 0
	cursorWidth             float32 = 1.5
	cursorTopPadding        float32 = 2
	cursorBottomPadding     float32 = 2
	passwordMaskChar                = '\u2022' // bullet character
)

// Default colors for DefaultPainter.
var (
	defaultBg                  = widget.ColorWhite
	defaultBorderColor         = widget.RGBA(0.45, 0.45, 0.45, 1.0)
	defaultFocusBorderColor    = widget.Hex(0x6750A4)
	defaultHoverBorderColor    = widget.RGBA(0.2, 0.2, 0.2, 1.0)
	defaultDisabledBorderColor = widget.RGBA(0.45, 0.45, 0.45, 0.38)
	defaultErrorColor          = widget.Hex(0xB3261E)
	defaultTextColor           = widget.RGBA(0.1, 0.1, 0.1, 1.0)
	defaultPlaceholderColor    = widget.RGBA(0.45, 0.45, 0.45, 1.0)
	defaultCursorColor         = widget.Hex(0x6750A4)
	defaultDisabledBg          = widget.RGBA(0.95, 0.95, 0.95, 1.0)
	defaultDisabledFg          = widget.RGBA(0.12, 0.12, 0.13, 0.38)
	defaultSelectionBg         = widget.Hex(0x6750A4).WithAlpha(0.2)
)
