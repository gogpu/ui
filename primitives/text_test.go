package primitives_test

import (
	"fmt"
	"testing"

	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/state"
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
	if style.Bold {
		t.Error("should not be bold by default")
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
		Align(primitives.TextAlignCenter).
		MaxLines(2).
		Ellipsis().
		LineHeight(1.4)

	style := tw.Style()
	if style.FontSize != 18 {
		t.Error("font size not chained")
	}
	if !style.Bold {
		t.Error("bold not chained")
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
	accState := tw.AccessibilityState()
	if accState.Hidden || accState.Disabled {
		t.Error("default state should be visible and enabled")
	}

	tw.SetVisible(false)
	accState = tw.AccessibilityState()
	if !accState.Hidden {
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
		{primitives.TextAlignStart, "Left"},
		{primitives.TextAlignCenter, "Center"},
		{primitives.TextAlignEnd, "Right"},
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

// --- Theme-aware default color ---

func TestTextDefaultColor_WithTheme(t *testing.T) {
	// When a ThemeProvider is set, Text should use OnSurface color.
	onSurface := widget.Hex(0x1C1B1F) // M3 light OnSurface
	tp := &testThemeProvider{onSurface: onSurface}

	ctx := widget.NewContext()
	ctx.SetThemeProvider(tp)

	tw := primitives.Text("Hello").FontSize(14)
	canvas := &mockCanvas{}
	_ = tw.Layout(ctx, geometry.Loose(geometry.Sz(200, 200)))
	tw.Draw(ctx, canvas)

	if canvas.drawTextCount == 0 {
		t.Fatal("expected DrawText call")
	}
	if canvas.lastTextColor != onSurface {
		t.Errorf("text color = %+v, want theme OnSurface %+v",
			canvas.lastTextColor, onSurface)
	}
}

func TestTextExplicitColor_OverridesTheme(t *testing.T) {
	// Explicit .Color() always wins over theme.
	onSurface := widget.Hex(0x1C1B1F)
	tp := &testThemeProvider{onSurface: onSurface}

	ctx := widget.NewContext()
	ctx.SetThemeProvider(tp)

	explicitRed := widget.ColorRed
	tw := primitives.Text("Hello").FontSize(14).Color(explicitRed)
	canvas := &mockCanvas{}
	_ = tw.Layout(ctx, geometry.Loose(geometry.Sz(200, 200)))
	tw.Draw(ctx, canvas)

	if canvas.drawTextCount == 0 {
		t.Fatal("expected DrawText call")
	}
	if canvas.lastTextColor != explicitRed {
		t.Errorf("text color = %+v, want explicit %+v",
			canvas.lastTextColor, explicitRed)
	}
}

func TestTextNoTheme_FallsBackToBlack(t *testing.T) {
	// Without a theme, text should default to black.
	ctx := widget.NewContext() // no theme provider

	tw := primitives.Text("Hello").FontSize(14)
	canvas := &mockCanvas{}
	_ = tw.Layout(ctx, geometry.Loose(geometry.Sz(200, 200)))
	tw.Draw(ctx, canvas)

	if canvas.drawTextCount == 0 {
		t.Fatal("expected DrawText call")
	}
	if canvas.lastTextColor != widget.ColorBlack {
		t.Errorf("text color = %+v, want ColorBlack %+v",
			canvas.lastTextColor, widget.ColorBlack)
	}
}

func TestTextReactiveFn_UsesThemeColor(t *testing.T) {
	// TextFn should also use theme colors.
	onSurface := widget.Hex(0xE6E1E5) // M3 dark OnSurface
	tp := &testThemeProvider{onSurface: onSurface, dark: true}

	ctx := widget.NewContext()
	ctx.SetThemeProvider(tp)

	tw := primitives.TextFn(func() string { return "Dynamic" }).FontSize(14)
	canvas := &mockCanvas{}
	_ = tw.Layout(ctx, geometry.Loose(geometry.Sz(200, 200)))
	tw.Draw(ctx, canvas)

	if canvas.drawTextCount == 0 {
		t.Fatal("expected DrawText call")
	}
	if canvas.lastTextColor != onSurface {
		t.Errorf("text color = %+v, want theme OnSurface %+v",
			canvas.lastTextColor, onSurface)
	}
}

// --- Signal Binding Tests ---

func TestTextContentSignal(t *testing.T) {
	sig := state.NewSignal("Signal Text")
	tw := primitives.Text("").ContentSignal(sig).FontSize(14)

	if tw.Content() != "Signal Text" {
		t.Errorf("content = %q, want %q", tw.Content(), "Signal Text")
	}
	if !tw.IsReactive() {
		t.Error("signal-bound text should be reactive")
	}
}

func TestTextContentSignalUpdate(t *testing.T) {
	sig := state.NewSignal("Initial")
	tw := primitives.Text("").ContentSignal(sig).FontSize(14)

	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	// First draw with initial value.
	_ = tw.Layout(ctx, geometry.Loose(geometry.Sz(200, 200)))
	tw.Draw(ctx, canvas)

	if canvas.drawTextCount == 0 {
		t.Fatal("expected DrawText call")
	}
	if canvas.lastText != "Initial" {
		t.Errorf("text = %q, want %q", canvas.lastText, "Initial")
	}

	// Update signal and redraw.
	sig.Set("Updated")
	canvas = &mockCanvas{}
	_ = tw.Layout(ctx, geometry.Loose(geometry.Sz(200, 200)))
	tw.Draw(ctx, canvas)

	if canvas.drawTextCount == 0 {
		t.Fatal("expected DrawText call after signal update")
	}
	if canvas.lastText != "Updated" {
		t.Errorf("text = %q, want %q", canvas.lastText, "Updated")
	}
}

func TestTextContentSignalPriority(t *testing.T) {
	t.Run("Signal overrides Fn", func(t *testing.T) {
		sig := state.NewSignal("signal")
		tw := primitives.TextFn(func() string { return "fn" }).ContentSignal(sig)

		if tw.Content() != "signal" {
			t.Errorf("content = %q, want %q (signal should override fn)", tw.Content(), "signal")
		}
	})

	t.Run("Signal overrides static", func(t *testing.T) {
		sig := state.NewSignal("signal")
		tw := primitives.Text("static").ContentSignal(sig)

		if tw.Content() != "signal" {
			t.Errorf("content = %q, want %q (signal should override static)", tw.Content(), "signal")
		}
	})

	t.Run("Fn used when no signal", func(t *testing.T) {
		tw := primitives.TextFn(func() string { return "fn" })

		if tw.Content() != "fn" {
			t.Errorf("content = %q, want %q", tw.Content(), "fn")
		}
	})

	t.Run("Static used when no signal and no fn", func(t *testing.T) {
		tw := primitives.Text("static")

		if tw.Content() != "static" {
			t.Errorf("content = %q, want %q", tw.Content(), "static")
		}
	})
}

// testThemeProvider is a minimal ThemeProvider for testing theme-aware primitives.
type testThemeProvider struct {
	dark      bool
	onSurface widget.Color
}

func (tp *testThemeProvider) IsDark() bool {
	return tp.dark
}

func (tp *testThemeProvider) OnSurface() widget.Color {
	return tp.onSurface
}

// --- Lifecycle Tests ---

func TestTextWidget_LifecycleInterface(t *testing.T) {
	var _ widget.Lifecycle = primitives.Text("hello")
}

func TestTextWidget_Mount_CreatesBindings(t *testing.T) {
	sig := state.NewSignal("hello")
	tw := primitives.Text("").ContentSignal(sig.AsReadonly())

	sched := state.NewScheduler(func(_ []widget.Widget) {})
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	tw.Mount(ctx)

	dirtyCount := 0
	sched.SetOnDirty(func() { dirtyCount++ })
	sig.Set("world")

	if dirtyCount == 0 {
		t.Error("signal change should mark widget dirty after mount")
	}
}

func TestTextWidget_Unmount_CleansBindings(t *testing.T) {
	sig := state.NewSignal("hello")
	tw := primitives.Text("").ContentSignal(sig.AsReadonly())

	sched := state.NewScheduler(func(_ []widget.Widget) {})
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	tw.Mount(ctx)
	tw.CleanupBindings()
	tw.Unmount()

	sig.Set("world")

	if sched.PendingCount() != 0 {
		t.Error("signal change after unmount should not mark widget dirty")
	}
}
