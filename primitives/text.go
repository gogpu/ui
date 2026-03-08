package primitives

import (
	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// defaultFontSize is the font size used when none is specified.
const defaultFontSize float32 = 14

// defaultLineHeight is the default line height multiplier relative to font size.
const defaultLineHeight float32 = 1.2

// estimatedCharWidth is a rough ratio of character width to font size for
// simple text measurement heuristics. Real measurement is provided by the
// Canvas implementation.
const estimatedCharWidth float32 = 0.6

// TextStyle holds all visual styling for a [TextWidget].
type TextStyle struct {
	FontSize   float32
	Color      widget.Color
	Bold       bool
	Italic     bool
	Align      TextAlign
	MaxLines   int
	Overflow   TextOverflow
	LineHeight float32
}

// TextWidget displays static or reactive text content.
//
// TextWidget implements [widget.Widget] and [a11y.Accessible].
//
// Create a TextWidget with [Text] (static) or [TextFn] (reactive).
//
// # Theme-Aware Default Color
//
// By default, TextWidget uses the theme's OnSurface color for text
// when a ThemeProvider is available on the context. If no theme is set,
// the fallback is [widget.ColorBlack]. An explicit call to [TextWidget.Color]
// always takes precedence over the theme default.
type TextWidget struct {
	widget.WidgetBase

	style         TextStyle
	content       string
	fn            func() string
	contentSignal state.ReadonlySignal[string]
	colorExplicit bool // true when Color() was called explicitly
}

// Text creates a new text widget with static content.
//
//	label := primitives.Text("Hello World").FontSize(18).Bold()
func Text(content string) *TextWidget {
	t := &TextWidget{
		content: content,
		style: TextStyle{
			FontSize:   defaultFontSize,
			Color:      widget.ColorBlack,
			LineHeight: defaultLineHeight,
		},
	}
	t.SetVisible(true)
	t.SetEnabled(true)
	return t
}

// TextFn creates a new text widget with reactive content.
//
// The function fn is called during layout and draw to obtain the current
// text. When the function reads a signal's value, changes to that signal
// will cause re-layout and re-draw when the binding is set up externally.
//
//	counter := state.NewSignal(0)
//	label := primitives.TextFn(func() string {
//	    return fmt.Sprintf("Count: %d", counter.Get())
//	}).FontSize(14)
func TextFn(fn func() string) *TextWidget {
	t := &TextWidget{
		fn: fn,
		style: TextStyle{
			FontSize:   defaultFontSize,
			Color:      widget.ColorBlack,
			LineHeight: defaultLineHeight,
		},
	}
	t.SetVisible(true)
	t.SetEnabled(true)
	return t
}

// ContentSignal binds the text content to a read-only reactive signal.
// When set, the signal value takes precedence over both [TextFn] and static
// content set via [Text].
//
// Because TextWidget is display-only, only [state.ReadonlySignal] is accepted
// (no write-back capability is needed).
//
//	name := state.NewSignal("Alice")
//	label := primitives.Text("").ContentSignal(name).FontSize(14)
func (t *TextWidget) ContentSignal(sig state.ReadonlySignal[string]) *TextWidget {
	t.contentSignal = sig
	return t
}

// --- Fluent style methods ---

// FontSize sets the font size in logical pixels.
func (t *TextWidget) FontSize(size float32) *TextWidget {
	t.style.FontSize = size
	return t
}

// Color sets the text color explicitly.
//
// An explicit color always takes precedence over the theme's default
// OnSurface color.
func (t *TextWidget) Color(c widget.Color) *TextWidget {
	t.style.Color = c
	t.colorExplicit = true
	return t
}

// Bold enables bold font weight.
func (t *TextWidget) Bold() *TextWidget {
	t.style.Bold = true
	return t
}

// Italic enables italic font style.
func (t *TextWidget) Italic() *TextWidget {
	t.style.Italic = true
	return t
}

// Align sets horizontal text alignment.
func (t *TextWidget) Align(a TextAlign) *TextWidget {
	t.style.Align = a
	return t
}

// MaxLines limits the number of displayed lines. Zero means unlimited.
func (t *TextWidget) MaxLines(n int) *TextWidget {
	t.style.MaxLines = n
	return t
}

// Ellipsis enables truncation with "..." when text overflows.
func (t *TextWidget) Ellipsis() *TextWidget {
	t.style.Overflow = TextOverflowEllipsis
	return t
}

// LineHeight sets the line height multiplier. The default is 1.2.
func (t *TextWidget) LineHeight(v float32) *TextWidget {
	t.style.LineHeight = v
	return t
}

// Style returns the current text style (read-only snapshot).
func (t *TextWidget) Style() TextStyle {
	return t.style
}

// Content returns the current text content.
//
// Resolution priority: [ContentSignal] > [TextFn] > static [Text].
func (t *TextWidget) Content() string {
	if t.contentSignal != nil {
		return t.contentSignal.Get()
	}
	if t.fn != nil {
		return t.fn()
	}
	return t.content
}

// IsReactive returns true if the widget uses a function or signal for its content.
func (t *TextWidget) IsReactive() bool {
	return t.contentSignal != nil || t.fn != nil
}

// --- widget.Widget interface ---

// Layout measures the text and returns the constrained size.
//
// Text measurement uses a simple character-width heuristic. Real font
// measurement is delegated to the Canvas implementation in production.
func (t *TextWidget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	text := t.Content()
	size := t.measureText(text, constraints.MaxWidth)
	resultSize := constraints.Constrain(size)
	t.SetBounds(geometry.FromPointSize(t.Position(), resultSize))
	return resultSize
}

// Draw renders the text content.
//
// The text color is resolved with the following priority:
//  1. Explicit color set via [TextWidget.Color] (always wins)
//  2. ThemeProvider's OnSurface color (if a theme is active)
//  3. [widget.ColorBlack] (fallback when no theme is set)
func (t *TextWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !t.IsVisible() {
		return
	}

	text := t.Content()
	if text == "" {
		return
	}

	bounds := t.Bounds()
	if bounds.IsEmpty() {
		return
	}

	var align float32
	switch t.style.Align {
	case TextAlignCenter:
		align = 0.5
	case TextAlignEnd:
		align = 1.0
	}

	color := t.resolveColor(ctx)
	canvas.DrawText(text, bounds, t.style.FontSize, color, t.style.Bold, align)
}

// resolveColor returns the text color using the priority chain:
// explicit > ThemeProvider.OnSurface() > ColorBlack.
func (t *TextWidget) resolveColor(ctx widget.Context) widget.Color {
	if t.colorExplicit {
		return t.style.Color
	}
	if ctx != nil {
		if tp := ctx.ThemeProvider(); tp != nil {
			return tp.OnSurface()
		}
	}
	return t.style.Color // defaults to ColorBlack from constructor
}

// Event returns false. Text widgets do not consume events.
func (t *TextWidget) Event(_ widget.Context, _ event.Event) bool {
	return false
}

// Children returns nil. Text is a leaf widget.
func (t *TextWidget) Children() []widget.Widget {
	return nil
}

// --- a11y.Accessible interface ---

// AccessibilityRole returns [a11y.RoleLabel].
func (t *TextWidget) AccessibilityRole() a11y.Role {
	return a11y.RoleLabel
}

// AccessibilityLabel returns the text content.
func (t *TextWidget) AccessibilityLabel() string {
	return t.Content()
}

// AccessibilityHint returns an empty string.
func (t *TextWidget) AccessibilityHint() string {
	return ""
}

// AccessibilityValue returns an empty string.
func (t *TextWidget) AccessibilityValue() string {
	return ""
}

// AccessibilityState returns the default state.
func (t *TextWidget) AccessibilityState() a11y.State {
	return a11y.State{
		Hidden: !t.IsVisible(),
	}
}

// AccessibilityActions returns nil. Text has no actions.
func (t *TextWidget) AccessibilityActions() []a11y.Action {
	return nil
}

// --- internal ---

// measureText estimates the size of text using a simple heuristic.
//
// It counts runes, applies the character width ratio, and wraps lines
// when the available width is bounded. MaxLines and Ellipsis are respected.
func (t *TextWidget) measureText(text string, maxWidth float32) geometry.Size {
	if text == "" {
		return geometry.Size{}
	}

	fontSize := t.style.FontSize
	if fontSize <= 0 {
		fontSize = defaultFontSize
	}
	lineH := fontSize * t.style.LineHeight
	charW := fontSize * estimatedCharWidth

	runes := []rune(text)
	runeCount := len(runes)

	if !isBoundedWidth(maxWidth) || maxWidth <= 0 {
		// Unbounded: single line
		width := float32(runeCount) * charW
		return geometry.Sz(width, lineH)
	}

	// Calculate characters per line
	charsPerLine := int(maxWidth / charW)
	if charsPerLine < 1 {
		charsPerLine = 1
	}

	lines := (runeCount + charsPerLine - 1) / charsPerLine
	if lines < 1 {
		lines = 1
	}

	// Apply max lines
	if t.style.MaxLines > 0 && lines > t.style.MaxLines {
		lines = t.style.MaxLines
	}

	width := maxWidth
	if lines == 1 {
		singleLineWidth := float32(runeCount) * charW
		if singleLineWidth < maxWidth {
			width = singleLineWidth
		}
	}

	height := float32(lines) * lineH
	return geometry.Sz(width, height)
}

// isBoundedWidth returns true if the width value is bounded (not infinity).
func isBoundedWidth(v float32) bool {
	return v < geometry.Infinity
}

// Compile-time interface checks.
var (
	_ widget.Widget   = (*TextWidget)(nil)
	_ a11y.Accessible = (*TextWidget)(nil)
)
