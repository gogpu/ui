package primitives

import (
	"strings"
	"unicode"
	"unicode/utf8"

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

	// FontFamily specifies a custom font family name (e.g., "NotoSansCJK").
	// When empty (the default), the embedded Inter font is used.
	// Custom fonts must be registered via the plugin system or font registry
	// before they can be referenced here.
	FontFamily string
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

// FontFamily sets a custom font family name.
//
// The font must be registered via the plugin system's [plugin.AssetLoader]
// or directly via the font registry before it can be used. When set, the
// widget will use [widget.StyledTextDrawer] to render text with the custom
// font, falling back to the standard [widget.Canvas.DrawText] with Inter
// if the canvas does not support styled text.
//
// Example:
//
//	label := primitives.Text("CJK").FontFamily("NotoSansCJK").FontSize(18)
func (t *TextWidget) FontFamily(name string) *TextWidget {
	t.style.FontFamily = name
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
//
// When a custom [TextWidget.FontFamily] or [TextWidget.Italic] is set
// and the canvas supports [widget.StyledTextDrawer], the styled text
// path is used to render with the custom font. Otherwise, the standard
// [widget.Canvas.DrawText] is used with the default embedded font.
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

	color := t.resolveColor(ctx)
	lines, _ := t.lines(text, bounds.Width())
	canvas.PushClip(bounds)
	defer canvas.PopClip()
	if len(lines) == 1 {
		t.drawLine(canvas, lines[0], bounds, color)
		return
	}
	height := t.lineHeight()
	for i, line := range lines {
		y := bounds.Min.Y + float32(i)*height
		if y >= bounds.Max.Y {
			break
		}
		t.drawLine(canvas, line, geometry.NewRect(bounds.Min.X, y, bounds.Width(), height), color)
	}
}

func (t *TextWidget) drawLine(canvas widget.Canvas, text string, bounds geometry.Rect, color widget.Color) {
	// Use StyledTextDrawer path when custom font family or italic is set.
	if t.style.FontFamily != "" || t.style.Italic {
		if sd, ok := canvas.(widget.StyledTextDrawer); ok {
			sd.DrawStyledText(text, bounds, widget.TextStyle{
				FontFamily: t.style.FontFamily,
				FontSize:   t.style.FontSize,
				Bold:       t.style.Bold,
				Italic:     t.style.Italic,
				Color:      color,
				Align:      t.style.Align,
			})
			return
		}
	}

	canvas.DrawText(text, bounds, t.style.FontSize, color, t.style.Bold, t.style.Align)
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

	lines, wrapped := t.lines(text, maxWidth)
	var longest int
	for _, line := range lines {
		longest = max(longest, utf8.RuneCountInString(line))
	}
	width := float32(longest) * t.fontSize() * estimatedCharWidth
	if wrapped {
		width = maxWidth
	}
	return geometry.Sz(width, float32(len(lines))*t.lineHeight())
}

func (t *TextWidget) fontSize() float32 {
	if t.style.FontSize > 0 {
		return t.style.FontSize
	}
	return defaultFontSize
}

func (t *TextWidget) lineHeight() float32 { return t.fontSize() * t.style.LineHeight }

// lines uses the same width estimate for measuring and painting. Explicit
// newlines start a new line even when the available width is unbounded.
func (t *TextWidget) lines(text string, width float32) ([]string, bool) {
	count := utf8.RuneCountInString(text)
	limit := count + 1
	charWidth := t.fontSize() * estimatedCharWidth
	if width > 0 && width < float32(limit)*charWidth {
		limit = max(1, int(width/charWidth+0.0001))
	}
	if count <= limit && !strings.ContainsRune(text, '\n') {
		return []string{text}, false
	}
	lines := make([]string, 0, 1)
	wrapped := false
	for _, paragraph := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		runes := []rune(paragraph)
		for len(runes) > limit {
			end := limit
			for end > 0 && !unicode.IsSpace(runes[end]) {
				end--
			}
			if end == 0 {
				end = limit
			}
			lines = append(lines, string(runes[:end]))
			runes = runes[end:]
			if len(runes) > 0 && unicode.IsSpace(runes[0]) {
				runes = runes[1:]
			}
			wrapped = true
		}
		lines = append(lines, string(runes))
	}
	if n := t.style.MaxLines; n > 0 && len(lines) > n {
		lines = lines[:n]
		if t.style.Overflow == TextOverflowEllipsis {
			last := []rune(lines[n-1])
			lines[n-1] = string(last[:min(len(last), limit-1)]) + "…"
		}
	}
	return lines, wrapped
}

// Mount creates signal bindings for push-based invalidation.
// Implements [widget.Lifecycle].
func (t *TextWidget) Mount(ctx widget.Context) {
	sched := ctx.Scheduler()
	if sched == nil {
		return
	}
	if t.contentSignal != nil {
		b := state.BindToSchedulerLayout(t.contentSignal, t, sched)
		t.AddBinding(b)
	}
}

// Unmount is called when the text widget is removed from the widget tree.
// Implements [widget.Lifecycle].
func (t *TextWidget) Unmount() {
	// Bindings are cleaned up automatically by WidgetBase.CleanupBindings().
}

// Compile-time interface checks.
var (
	_ widget.Widget    = (*TextWidget)(nil)
	_ a11y.Accessible  = (*TextWidget)(nil)
	_ widget.Lifecycle = (*TextWidget)(nil)
)
