package primitives_test

import (
	"fmt"
	"testing"

	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
)

// --- Text construction ---

func TestTextStaticContent(t *testing.T) {
	tw := primitives.Text("Hello")
	if tw.Content() != "Hello" {
		t.Errorf("expected 'Hello', got %q", tw.Content())
	}
}

func TestTextReactiveContent(t *testing.T) {
	counter := 0
	tw := primitives.TextFn(func() string {
		return fmt.Sprintf("Count: %d", counter)
	})

	if tw.Content() != "Count: 0" {
		t.Errorf("expected 'Count: 0', got %q", tw.Content())
	}

	counter = 42
	if tw.Content() != "Count: 42" {
		t.Errorf("expected 'Count: 42', got %q", tw.Content())
	}
}

func TestTextIsReactive(t *testing.T) {
	static := primitives.Text("Hello")
	if static.IsReactive() {
		t.Error("static text should not be reactive")
	}

	reactive := primitives.TextFn(func() string { return "hi" })
	if !reactive.IsReactive() {
		t.Error("TextFn should be reactive")
	}
}

func TestTextDefaultStyle(t *testing.T) {
	tw := primitives.Text("Hello")
	style := tw.Style()
	if style.FontSize != 14 {
		t.Errorf("expected default font size 14, got %f", style.FontSize)
	}
	if style.Color != widget.ColorBlack {
		t.Error("expected default color black")
	}
	if style.Bold || style.Italic {
		t.Error("should not be bold or italic by default")
	}
	if style.LineHeight != 1.2 {
		t.Errorf("expected default line height 1.2, got %f", style.LineHeight)
	}
}

func TestTextIsVisibleAndEnabled(t *testing.T) {
	tw := primitives.Text("Hello")
	if !tw.IsVisible() {
		t.Error("text should be visible by default")
	}
	if !tw.IsEnabled() {
		t.Error("text should be enabled by default")
	}
}

// --- Fluent style methods ---

func TestTextFontSize(t *testing.T) {
	tw := primitives.Text("Hello").FontSize(24)
	if tw.Style().FontSize != 24 {
		t.Errorf("expected font size 24, got %f", tw.Style().FontSize)
	}
}

func TestTextColor(t *testing.T) {
	c := widget.Hex(0xFF0000)
	tw := primitives.Text("Hello").Color(c)
	if tw.Style().Color != c {
		t.Error("color not set")
	}
}

func TestTextBold(t *testing.T) {
	tw := primitives.Text("Hello").Bold()
	if !tw.Style().Bold {
		t.Error("bold not set")
	}
}

func TestTextItalic(t *testing.T) {
	tw := primitives.Text("Hello").Italic()
	if !tw.Style().Italic {
		t.Error("italic not set")
	}
}

func TestTextAlign(t *testing.T) {
	tests := []struct {
		name  string
		align primitives.TextAlign
	}{
		{"Start", primitives.TextAlignStart},
		{"Center", primitives.TextAlignCenter},
		{"End", primitives.TextAlignEnd},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tw := primitives.Text("Hello").Align(tt.align)
			if tw.Style().Align != tt.align {
				t.Errorf("expected align %s, got %s", tt.align, tw.Style().Align)
			}
		})
	}
}

func TestTextMaxLines(t *testing.T) {
	tw := primitives.Text("Hello").MaxLines(3)
	if tw.Style().MaxLines != 3 {
		t.Errorf("expected 3 max lines, got %d", tw.Style().MaxLines)
	}
}

func TestTextEllipsis(t *testing.T) {
	tw := primitives.Text("Hello").Ellipsis()
	if tw.Style().Overflow != primitives.TextOverflowEllipsis {
		t.Errorf("expected ellipsis overflow, got %s", tw.Style().Overflow)
	}
}

func TestTextLineHeight(t *testing.T) {
	tw := primitives.Text("Hello").LineHeight(1.5)
	if tw.Style().LineHeight != 1.5 {
		t.Errorf("expected 1.5, got %f", tw.Style().LineHeight)
	}
}

func TestTextFluentChaining(t *testing.T) {
	tw := primitives.Text("Hello").
		FontSize(18).
		Color(widget.ColorRed).
		Bold().
		Italic().
		Align(primitives.TextAlignCenter).
		MaxLines(2).
		Ellipsis().
		LineHeight(1.4)

	style := tw.Style()
	if style.FontSize != 18 {
		t.Error("font size not chained")
	}
	if !style.Bold || !style.Italic {
		t.Error("bold/italic not chained")
	}
	if style.Align != primitives.TextAlignCenter {
		t.Error("align not chained")
	}
	if style.MaxLines != 2 {
		t.Error("max lines not chained")
	}
	if style.Overflow != primitives.TextOverflowEllipsis {
		t.Error("overflow not chained")
	}
	if style.LineHeight != 1.4 {
		t.Error("line height not chained")
	}
}

// --- Layout ---

func TestTextLayoutEmptyString(t *testing.T) {
	tw := primitives.Text("")
	ctx := widget.NewContext()
	size := tw.Layout(ctx, geometry.Loose(geometry.Sz(200, 200)))
	if size.Width != 0 || size.Height != 0 {
		t.Errorf("empty text should have zero size, got %s", size)
	}
}

func TestTextLayoutSingleLine(t *testing.T) {
	tw := primitives.Text("Hello").FontSize(14)
	ctx := widget.NewContext()
	size := tw.Layout(ctx, geometry.Loose(geometry.Sz(500, 500)))

	// 5 chars * 0.6 * 14 = 42, height = 14 * 1.2 = 16.8
	if size.Width < 40 || size.Width > 50 {
		t.Errorf("unexpected single-line width: %f", size.Width)
	}
	if size.Height < 15 || size.Height > 20 {
		t.Errorf("unexpected single-line height: %f", size.Height)
	}
}

func TestTextLayoutWraps(t *testing.T) {
	// 20 chars * 0.6 * 14 = 168 natural width, constrain to 100
	tw := primitives.Text("Hello World 12345678").FontSize(14)
	ctx := widget.NewContext()
	size := tw.Layout(ctx, geometry.Loose(geometry.Sz(100, 500)))

	// Should wrap to multiple lines
	singleLineHeight := float32(14 * 1.2)
	if size.Height <= singleLineHeight+0.1 {
		t.Errorf("text should wrap: height=%f, singleLine=%f", size.Height, singleLineHeight)
	}
}

func TestTextLayoutMaxLinesTruncates(t *testing.T) {
	// Long text that would wrap to many lines
	tw := primitives.Text("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA").
		FontSize(14).MaxLines(2)
	ctx := widget.NewContext()
	size := tw.Layout(ctx, geometry.Loose(geometry.Sz(100, 500)))

	maxHeight := float32(2) * 14 * 1.2
	if size.Height > maxHeight+0.1 {
		t.Errorf("max lines should limit height: got %f, want <= %f", size.Height, maxHeight)
	}
}

func TestTextLayoutUnbounded(t *testing.T) {
	tw := primitives.Text("Hello").FontSize(14)
	ctx := widget.NewContext()
	size := tw.Layout(ctx, geometry.Expand())

	// Should be a single line with computed width
	if size.Width < 40 {
		t.Errorf("unbounded width too small: %f", size.Width)
	}
	lineH := float32(14 * 1.2)
	if size.Height < lineH-1 || size.Height > lineH+1 {
		t.Errorf("unbounded should be single line: height=%f", size.Height)
	}
}

func TestTextLayoutReactive(t *testing.T) {
	text := "Short"
	tw := primitives.TextFn(func() string { return text }).FontSize(14)
	ctx := widget.NewContext()

	size1 := tw.Layout(ctx, geometry.Loose(geometry.Sz(500, 500)))

	text = "A much longer string"
	size2 := tw.Layout(ctx, geometry.Loose(geometry.Sz(500, 500)))

	if size2.Width <= size1.Width {
		t.Errorf("longer text should be wider: %f <= %f", size2.Width, size1.Width)
	}
}

func TestTextLayoutSetsBounds(t *testing.T) {
	tw := primitives.Text("Hello").FontSize(14)
	ctx := widget.NewContext()
	size := tw.Layout(ctx, geometry.Loose(geometry.Sz(500, 500)))

	bounds := tw.Bounds()
	if bounds.Width() != size.Width || bounds.Height() != size.Height {
		t.Errorf("bounds should match layout size: bounds=%s, size=%s", bounds.Size(), size)
	}
}

// --- Draw ---

func TestTextDrawNoPanicEmpty(t *testing.T) {
	tw := primitives.Text("")
	ctx := widget.NewContext()
	canvas := &mockCanvas{}
	_ = tw.Layout(ctx, geometry.Loose(geometry.Sz(100, 100)))
	tw.Draw(ctx, canvas) // Should not panic
}

func TestTextDrawRendersText(t *testing.T) {
	tw := primitives.Text("Hello").FontSize(14)
	ctx := widget.NewContext()
	canvas := &mockCanvas{}
	_ = tw.Layout(ctx, geometry.Loose(geometry.Sz(100, 100)))
	tw.Draw(ctx, canvas)

	if canvas.drawTextCount == 0 {
		t.Error("text draw should call DrawText")
	}
}

func TestTextDrawInvisible(t *testing.T) {
	tw := primitives.Text("Hello")
	tw.SetVisible(false)
	ctx := widget.NewContext()
	canvas := &mockCanvas{}
	_ = tw.Layout(ctx, geometry.Loose(geometry.Sz(100, 100)))
	tw.Draw(ctx, canvas)

	if canvas.drawRectCount != 0 {
		t.Error("invisible text should not draw")
	}
}

// --- Event ---

func TestTextEventNotConsumed(t *testing.T) {
	tw := primitives.Text("Hello")
	ctx := widget.NewContext()
	e := &event.Base{}
	if tw.Event(ctx, e) {
		t.Error("text should not consume events")
	}
}

// --- Children ---

func TestTextChildrenNil(t *testing.T) {
	tw := primitives.Text("Hello")
	if tw.Children() != nil {
		t.Error("text should have no children")
	}
}

// --- Accessibility ---

func TestTextAccessibilityRole(t *testing.T) {
	tw := primitives.Text("Hello")
	if tw.AccessibilityRole() != a11y.RoleLabel {
		t.Errorf("expected RoleLabel, got %s", tw.AccessibilityRole())
	}
}

func TestTextAccessibilityLabelStatic(t *testing.T) {
	tw := primitives.Text("Hello World")
	if tw.AccessibilityLabel() != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", tw.AccessibilityLabel())
	}
}

func TestTextAccessibilityLabelReactive(t *testing.T) {
	text := "Initial"
	tw := primitives.TextFn(func() string { return text })
	if tw.AccessibilityLabel() != "Initial" {
		t.Errorf("expected 'Initial', got %q", tw.AccessibilityLabel())
	}

	text = "Updated"
	if tw.AccessibilityLabel() != "Updated" {
		t.Errorf("expected 'Updated', got %q", tw.AccessibilityLabel())
	}
}

func TestTextAccessibilityState(t *testing.T) {
	tw := primitives.Text("Hello")
	state := tw.AccessibilityState()
	if state.Hidden || state.Disabled {
		t.Error("default state should be visible and enabled")
	}

	tw.SetVisible(false)
	state = tw.AccessibilityState()
	if !state.Hidden {
		t.Error("invisible text should report Hidden=true")
	}
}

func TestTextAccessibilityActions(t *testing.T) {
	tw := primitives.Text("Hello")
	if tw.AccessibilityActions() != nil {
		t.Error("text should have no actions")
	}
}

func TestTextAccessibilityHint(t *testing.T) {
	tw := primitives.Text("Hello")
	if tw.AccessibilityHint() != "" {
		t.Error("text should have no hint")
	}
}

func TestTextAccessibilityValue(t *testing.T) {
	tw := primitives.Text("Hello")
	if tw.AccessibilityValue() != "" {
		t.Error("text should have no value")
	}
}

// --- Style enums ---

func TestTextAlignString(t *testing.T) {
	tests := []struct {
		align primitives.TextAlign
		want  string
	}{
		{primitives.TextAlignStart, "Start"},
		{primitives.TextAlignCenter, "Center"},
		{primitives.TextAlignEnd, "End"},
		{primitives.TextAlign(99), "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.align.String(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTextOverflowString(t *testing.T) {
	tests := []struct {
		overflow primitives.TextOverflow
		want     string
	}{
		{primitives.TextOverflowClip, "Clip"},
		{primitives.TextOverflowEllipsis, "Ellipsis"},
		{primitives.TextOverflow(99), "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.overflow.String(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
