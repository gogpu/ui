package fluent_test

import (
	"github.com/gogpu/gg/scene"
	"image"
	"testing"

	"github.com/gogpu/ui/core/button"
	"github.com/gogpu/ui/core/checkbox"
	"github.com/gogpu/ui/core/dialog"
	"github.com/gogpu/ui/core/dropdown"
	"github.com/gogpu/ui/core/radio"
	"github.com/gogpu/ui/core/scrollview"
	"github.com/gogpu/ui/core/slider"
	"github.com/gogpu/ui/core/tabview"
	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/theme/fluent"
	"github.com/gogpu/ui/widget"
)

// drawCall represents a single drawing operation recorded by recordCanvas.
type drawCall struct {
	method      string
	bounds      geometry.Rect
	color       widget.Color
	radius      float32
	strokeWidth float32
	text        string
	fontSize    float32
	bold        bool
	align       widget.TextAlign
}

// recordCanvas records all drawing operations for test assertions.
type recordCanvas struct {
	calls []drawCall
}

func (c *recordCanvas) Clear(color widget.Color) {
	c.calls = append(c.calls, drawCall{method: methodClear, color: color})
}

func (c *recordCanvas) DrawRect(r geometry.Rect, color widget.Color) {
	c.calls = append(c.calls, drawCall{method: methodDrawRect, bounds: r, color: color})
}

func (c *recordCanvas) FillRectDirect(_ geometry.Rect, _ widget.Color) {}

func (c *recordCanvas) StrokeRect(r geometry.Rect, color widget.Color, strokeWidth float32) {
	c.calls = append(c.calls, drawCall{method: methodStrokeRect, bounds: r, color: color, strokeWidth: strokeWidth})
}

func (c *recordCanvas) DrawRoundRect(r geometry.Rect, color widget.Color, radius float32) {
	c.calls = append(c.calls, drawCall{method: methodDrawRoundRect, bounds: r, color: color, radius: radius})
}

func (c *recordCanvas) StrokeRoundRect(r geometry.Rect, color widget.Color, radius float32, strokeWidth float32) {
	c.calls = append(c.calls, drawCall{
		method: methodStrokeRoundRect, bounds: r, color: color,
		radius: radius, strokeWidth: strokeWidth,
	})
}

func (c *recordCanvas) DrawCircle(center geometry.Point, radius float32, color widget.Color) {
	c.calls = append(c.calls, drawCall{method: methodDrawCircle, color: color, radius: radius})
}

func (c *recordCanvas) StrokeCircle(center geometry.Point, radius float32, color widget.Color, strokeWidth float32) {
	c.calls = append(c.calls, drawCall{method: methodStrokeCircle, color: color, radius: radius, strokeWidth: strokeWidth})
}
func (c *recordCanvas) StrokeArc(_ geometry.Point, _ float32, _, _ float64, _ widget.Color, _ float32) {
}

func (c *recordCanvas) DrawLine(from, to geometry.Point, color widget.Color, strokeWidth float32) {
	c.calls = append(c.calls, drawCall{method: methodDrawLine, color: color, strokeWidth: strokeWidth})
}

func (c *recordCanvas) DrawText(text string, bounds geometry.Rect, fontSize float32, color widget.Color, bold bool, align widget.TextAlign) {
	c.calls = append(c.calls, drawCall{
		method: methodDrawText, text: text, bounds: bounds,
		fontSize: fontSize, color: color, bold: bold, align: align,
	})
}

func (c *recordCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}

func (c *recordCanvas) DrawImage(_ image.Image, _ geometry.Point) {}

func (c *recordCanvas) PushClip(_ geometry.Rect)                     {}
func (c *recordCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *recordCanvas) PopClip()                                     {}
func (c *recordCanvas) PushTransform(_ geometry.Point)               {}
func (c *recordCanvas) PopTransform()                                {}
func (c *recordCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }
func (c *recordCanvas) ScreenOriginBase() geometry.Point             { return geometry.Point{} }
func (c *recordCanvas) ClipBounds() geometry.Rect                    { return geometry.NewRect(0, 0, 10000, 10000) }
func (c *recordCanvas) ReplayScene(_ *scene.Scene)                   {}

// Method name constants to satisfy goconst.
const (
	methodClear           = "Clear"
	methodDrawRect        = "DrawRect"
	methodStrokeRect      = "StrokeRect"
	methodDrawRoundRect   = "DrawRoundRect"
	methodStrokeRoundRect = "StrokeRoundRect"
	methodDrawCircle      = "DrawCircle"
	methodStrokeCircle    = "StrokeCircle"
	methodDrawLine        = "DrawLine"
	methodDrawText        = "DrawText"
)

func (c *recordCanvas) methodCalls(method string) []drawCall {
	var result []drawCall
	for _, call := range c.calls {
		if call.method == method {
			result = append(result, call)
		}
	}
	return result
}

// Compile-time check that recordCanvas implements Canvas.
var _ widget.Canvas = (*recordCanvas)(nil)

// testBounds returns a standard non-empty rectangle for tests.
func testBounds() geometry.Rect {
	return geometry.NewRect(10, 10, 120, 40)
}

// --- Theme creation tests ---

func TestNewTheme(t *testing.T) {
	theme := fluent.NewTheme()
	if theme == nil {
		t.Fatal("NewTheme() returned nil")
	}
	if theme.IsDark() {
		t.Error("NewTheme() should return light theme")
	}
	if theme.Colors.Accent.A == 0 {
		t.Error("Accent color should have non-zero alpha")
	}
	if theme.Colors.Surface.A == 0 {
		t.Error("Surface color should have non-zero alpha")
	}
}

func TestNewDarkTheme(t *testing.T) {
	theme := fluent.NewDarkTheme()
	if theme == nil {
		t.Fatal("NewDarkTheme() returned nil")
	}
	if !theme.IsDark() {
		t.Error("NewDarkTheme() should return dark theme")
	}

	// Dark surface should be dark.
	surfaceLum := luminance(theme.Colors.Surface)
	if surfaceLum > 0.2 {
		t.Errorf("dark surface luminance = %f, should be < 0.2", surfaceLum)
	}
}

func TestNewThemeWithAccentColor(t *testing.T) {
	customAccent := widget.Hex(0x744DA9) // purple
	theme := fluent.NewTheme(fluent.WithAccentColor(customAccent))

	if theme.Colors.Accent != customAccent {
		t.Errorf("accent = %v, want %v", theme.Colors.Accent, customAccent)
	}
}

func TestThemeOnSurface(t *testing.T) {
	theme := fluent.NewTheme()
	onSurface := theme.OnSurface()
	if onSurface.A == 0 {
		t.Error("OnSurface should have non-zero alpha")
	}
	if onSurface != theme.Colors.OnSurface {
		t.Error("OnSurface() should equal Colors.OnSurface")
	}
}

func TestThemeImplementsThemeProvider(t *testing.T) {
	var _ widget.ThemeProvider = (*fluent.Theme)(nil)
}

func TestLightScheme(t *testing.T) {
	cs := fluent.LightScheme(fluent.DefaultAccentColor)
	assertNonZero(t, "Accent", cs.Accent)
	assertNonZero(t, "Surface", cs.Surface)
	assertNonZero(t, "OnSurface", cs.OnSurface)
	assertNonZero(t, "Error", cs.Error)

	// Light surface should be bright.
	surfaceLum := luminance(cs.Surface)
	if surfaceLum < 0.9 {
		t.Errorf("light surface luminance = %f, should be > 0.9", surfaceLum)
	}
}

func TestDarkScheme(t *testing.T) {
	cs := fluent.DarkScheme(fluent.DefaultAccentColor)
	assertNonZero(t, "Accent", cs.Accent)
	assertNonZero(t, "OnSurface", cs.OnSurface)

	surfaceLum := luminance(cs.Surface)
	if surfaceLum > 0.2 {
		t.Errorf("dark surface luminance = %f, should be < 0.2", surfaceLum)
	}
}

func TestAsTheme(t *testing.T) {
	flTheme := fluent.NewTheme()
	generic := flTheme.AsTheme()
	if generic == nil {
		t.Fatal("AsTheme() returned nil")
	}
	if generic.Name != "Fluent Light" {
		t.Errorf("Name = %q, want %q", generic.Name, "Fluent Light")
	}
	if generic.Colors.Primary.A == 0 {
		t.Error("Primary should have non-zero alpha")
	}

	darkGeneric := fluent.NewDarkTheme().AsTheme()
	if darkGeneric.Name != "Fluent Dark" {
		t.Errorf("dark Name = %q, want %q", darkGeneric.Name, "Fluent Dark")
	}
}

func TestDifferentAccentsProduceDifferentSchemes(t *testing.T) {
	blue := fluent.LightScheme(widget.Hex(0x0078D4))
	red := fluent.LightScheme(widget.Hex(0xD40000))

	if colorEqual(blue.Accent, red.Accent) {
		t.Error("blue and red accents should produce different schemes")
	}
}

// --- Button painter tests ---

func TestButtonPainterImplementsInterface(t *testing.T) {
	var _ button.Painter = fluent.ButtonPainter{}
}

func TestPaintButtonEmptyBounds(t *testing.T) {
	canvas := &recordCanvas{}
	painter := fluent.ButtonPainter{}
	painter.PaintButton(canvas, button.PaintState{
		Text:    "Click",
		Variant: button.Filled,
		Bounds:  geometry.Rect{},
	})
	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no draw calls, got %d", len(canvas.calls))
	}
}

func TestPaintButtonFilled(t *testing.T) {
	canvas := &recordCanvas{}
	painter := fluent.ButtonPainter{}
	painter.PaintButton(canvas, button.PaintState{
		Text:    "Submit",
		Variant: button.Filled,
		Bounds:  testBounds(),
	})

	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Errorf("Filled should draw 1 DrawRoundRect, got %d", len(roundRects))
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("Filled should draw 1 DrawText, got %d", len(texts))
	}
	if texts[0].text != "Submit" {
		t.Errorf("text should be 'Submit', got %q", texts[0].text)
	}
}

func TestPaintButtonOutlined(t *testing.T) {
	canvas := &recordCanvas{}
	painter := fluent.ButtonPainter{}
	painter.PaintButton(canvas, button.PaintState{
		Text:    "Cancel",
		Variant: button.Outlined,
		Bounds:  testBounds(),
	})

	// Normal (not hovered): no DrawRoundRect, only StrokeRoundRect (border).
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 0 {
		t.Errorf("Outlined normal should draw 0 DrawRoundRect, got %d", len(roundRects))
	}

	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) != 1 {
		t.Errorf("Outlined should draw 1 StrokeRoundRect (border), got %d", len(strokes))
	}
}

func TestPaintButtonOutlinedHovered(t *testing.T) {
	canvas := &recordCanvas{}
	painter := fluent.ButtonPainter{}
	painter.PaintButton(canvas, button.PaintState{
		Text:    "Cancel",
		Variant: button.Outlined,
		Bounds:  testBounds(),
		Hovered: true,
	})

	// Hovered: DrawRoundRect (hover overlay) + StrokeRoundRect (border).
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Errorf("Outlined hovered should draw 1 DrawRoundRect (hover overlay), got %d", len(roundRects))
	}

	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) != 1 {
		t.Errorf("Outlined hovered should draw 1 StrokeRoundRect (border), got %d", len(strokes))
	}
}

func TestPaintButtonTextOnly(t *testing.T) {
	canvas := &recordCanvas{}
	painter := fluent.ButtonPainter{}
	painter.PaintButton(canvas, button.PaintState{
		Text:    "Learn more",
		Variant: button.TextOnly,
		Bounds:  testBounds(),
	})

	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 0 {
		t.Errorf("TextOnly normal should draw 0 DrawRoundRect, got %d", len(roundRects))
	}
}

func TestPaintButtonTextOnlyHover(t *testing.T) {
	canvas := &recordCanvas{}
	painter := fluent.ButtonPainter{}
	painter.PaintButton(canvas, button.PaintState{
		Text:    "Hover",
		Variant: button.TextOnly,
		Bounds:  testBounds(),
		Hovered: true,
	})

	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Errorf("TextOnly hover should draw 1 DrawRoundRect, got %d", len(roundRects))
	}
}

func TestPaintButtonDisabled(t *testing.T) {
	canvas := &recordCanvas{}
	painter := fluent.ButtonPainter{}
	painter.PaintButton(canvas, button.PaintState{
		Text:     "Disabled",
		Variant:  button.Filled,
		Bounds:   testBounds(),
		Disabled: true,
	})

	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Errorf("disabled should draw 1 DrawRoundRect, got %d", len(roundRects))
	}

	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) != 0 {
		t.Errorf("disabled should not draw focus ring, got %d StrokeRoundRect", len(strokes))
	}
}

func TestPaintButtonFocused(t *testing.T) {
	canvas := &recordCanvas{}
	painter := fluent.ButtonPainter{}
	painter.PaintButton(canvas, button.PaintState{
		Text:    "Focused",
		Variant: button.Filled,
		Bounds:  testBounds(),
		Focused: true,
	})

	// Fluent focus ring = inner StrokeRoundRect.
	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) != 1 {
		t.Errorf("focused should draw 1 StrokeRoundRect (focus ring), got %d", len(strokes))
	}
}

func TestPaintButtonWithTheme(t *testing.T) {
	theme := fluent.NewTheme()
	painter := fluent.ButtonPainter{Theme: theme}
	canvas := &recordCanvas{}

	painter.PaintButton(canvas, button.PaintState{
		Text:    "Themed",
		Variant: button.Filled,
		Bounds:  testBounds(),
	})

	if len(canvas.calls) == 0 {
		t.Fatal("themed painter should produce draw calls")
	}
}

// --- Checkbox painter tests ---

func TestCheckboxPainterImplementsInterface(t *testing.T) {
	var _ checkbox.Painter = fluent.CheckboxPainter{}
}

func TestPaintCheckboxChecked(t *testing.T) {
	canvas := &recordCanvas{}
	painter := fluent.CheckboxPainter{}
	painter.PaintCheckbox(canvas, checkbox.PaintState{
		Label:   "Accept",
		Checked: true,
		Bounds:  testBounds(),
	})

	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) < 1 {
		t.Error("checked checkbox should draw at least 1 DrawRoundRect")
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("should draw 1 DrawText (label), got %d", len(texts))
	}
	if texts[0].text != "Accept" {
		t.Errorf("label = %q, want 'Accept'", texts[0].text)
	}
}

func TestPaintCheckboxUnchecked(t *testing.T) {
	canvas := &recordCanvas{}
	painter := fluent.CheckboxPainter{}
	painter.PaintCheckbox(canvas, checkbox.PaintState{
		Label:  "Opt",
		Bounds: testBounds(),
	})

	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) < 1 {
		t.Error("unchecked checkbox should draw at least 1 StrokeRoundRect (border)")
	}
}

func TestPaintCheckboxEmpty(t *testing.T) {
	canvas := &recordCanvas{}
	painter := fluent.CheckboxPainter{}
	painter.PaintCheckbox(canvas, checkbox.PaintState{Bounds: geometry.Rect{}})
	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no calls, got %d", len(canvas.calls))
	}
}

// --- Radio painter tests ---

func TestRadioPainterImplementsInterface(t *testing.T) {
	var _ radio.Painter = fluent.RadioPainter{}
}

func TestPaintRadioSelected(t *testing.T) {
	canvas := &recordCanvas{}
	painter := fluent.RadioPainter{}
	painter.PaintRadio(canvas, radio.PaintState{
		Label:    "Option A",
		Selected: true,
		Bounds:   testBounds(),
	})

	circles := canvas.methodCalls(methodDrawCircle)
	if len(circles) < 2 {
		t.Errorf("selected radio should draw at least 2 circles (outer + inner), got %d", len(circles))
	}
}

func TestPaintRadioUnselected(t *testing.T) {
	canvas := &recordCanvas{}
	painter := fluent.RadioPainter{}
	painter.PaintRadio(canvas, radio.PaintState{
		Label:  "Option B",
		Bounds: testBounds(),
	})

	strokeCircles := canvas.methodCalls(methodStrokeCircle)
	if len(strokeCircles) < 1 {
		t.Error("unselected radio should draw at least 1 StrokeCircle (border)")
	}
}

// --- TextField painter tests ---

func TestTextFieldPainterImplementsInterface(t *testing.T) {
	var _ textfield.Painter = fluent.TextFieldPainter{}
}

func TestPaintTextField(t *testing.T) {
	canvas := &recordCanvas{}
	painter := fluent.TextFieldPainter{}
	painter.PaintTextField(canvas, textfield.PaintState{
		Text:   "Hello",
		Bounds: testBounds(),
	})

	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) < 1 {
		t.Error("text field should draw at least 1 DrawRoundRect (background)")
	}

	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) < 1 {
		t.Error("text field should draw at least 1 StrokeRoundRect (border)")
	}
}

func TestPaintTextFieldEmpty(t *testing.T) {
	canvas := &recordCanvas{}
	painter := fluent.TextFieldPainter{}
	painter.PaintTextField(canvas, textfield.PaintState{Bounds: geometry.Rect{}})
	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no calls, got %d", len(canvas.calls))
	}
}

// --- Dropdown painter tests ---

func TestDropdownPainterImplementsInterface(t *testing.T) {
	var _ dropdown.Painter = fluent.DropdownPainter{}
}

func TestPaintDropdownTrigger(t *testing.T) {
	canvas := &recordCanvas{}
	painter := fluent.DropdownPainter{}
	painter.PaintTrigger(canvas, &dropdown.TriggerPaintState{
		Bounds:       testBounds(),
		SelectedText: "Item 1",
	})

	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) < 1 {
		t.Error("trigger should draw at least 1 DrawRoundRect (background)")
	}
}

func TestPaintDropdownMenu(t *testing.T) {
	canvas := &recordCanvas{}
	painter := fluent.DropdownPainter{}
	painter.PaintMenu(canvas, &dropdown.MenuPaintState{
		Bounds: testBounds(),
		Items: []dropdown.ItemDef{
			{Value: "a", Label: "Alpha"},
			{Value: "b", Label: "Beta"},
		},
		HighlightedIndex: -1,
		SelectedIndex:    0,
		VisibleCount:     2,
		ItemHeight:       32,
	})

	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) < 1 {
		t.Error("menu should draw at least 1 DrawRoundRect (background)")
	}
}

// --- Slider painter tests ---

func TestSliderPainterImplementsInterface(t *testing.T) {
	var _ slider.Painter = fluent.SliderPainter{}
}

func TestPaintSlider(t *testing.T) {
	canvas := &recordCanvas{}
	painter := fluent.SliderPainter{}
	painter.PaintSlider(canvas, slider.PaintState{
		Value:    50,
		Min:      0,
		Max:      100,
		Progress: 0.5,
		Bounds:   geometry.NewRect(10, 10, 200, 30),
	})

	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) < 2 {
		t.Errorf("slider should draw at least 2 DrawRoundRect (inactive + active track), got %d", len(roundRects))
	}

	circles := canvas.methodCalls(methodDrawCircle)
	if len(circles) < 2 {
		t.Errorf("slider should draw at least 2 circles (thumb border + thumb fill), got %d", len(circles))
	}
}

func TestPaintSliderEmpty(t *testing.T) {
	canvas := &recordCanvas{}
	painter := fluent.SliderPainter{}
	painter.PaintSlider(canvas, slider.PaintState{Bounds: geometry.Rect{}})
	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no calls, got %d", len(canvas.calls))
	}
}

// --- Dialog painter tests ---

func TestDialogPainterImplementsInterface(t *testing.T) {
	var _ dialog.Painter = fluent.DialogPainter{}
}

func TestPaintDialog(t *testing.T) {
	canvas := &recordCanvas{}
	painter := fluent.DialogPainter{}
	painter.PaintDialog(canvas, dialog.PaintState{
		Title:  "Confirm",
		Bounds: geometry.NewRect(100, 100, 300, 200),
		Actions: []dialog.Action{
			{Label: "OK"},
			{Label: "Cancel"},
		},
	})

	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) < 1 {
		t.Error("dialog should draw at least 1 DrawRoundRect (surface)")
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) < 1 {
		t.Error("dialog should draw at least 1 DrawText (title)")
	}
}

func TestPaintDialogEmpty(t *testing.T) {
	canvas := &recordCanvas{}
	painter := fluent.DialogPainter{}
	painter.PaintDialog(canvas, dialog.PaintState{Bounds: geometry.Rect{}})
	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no calls, got %d", len(canvas.calls))
	}
}

// --- Scrollbar painter tests ---

func TestScrollbarPainterImplementsInterface(t *testing.T) {
	var _ scrollview.Painter = fluent.ScrollbarPainter{}
}

func TestPaintScrollbar(t *testing.T) {
	canvas := &recordCanvas{}
	painter := fluent.ScrollbarPainter{}
	painter.PaintScrollbar(canvas, scrollview.PaintState{
		Bounds:         geometry.NewRect(10, 10, 200, 300),
		VScrollVisible: true,
		VThumbRect:     geometry.NewRect(198, 20, 8, 60),
		VTrackRect:     geometry.NewRect(198, 10, 8, 300),
	})

	if len(canvas.calls) == 0 {
		t.Error("scrollbar should produce draw calls")
	}
}

// --- TabView painter tests ---

func TestTabViewPainterImplementsInterface(t *testing.T) {
	var _ tabview.Painter = fluent.TabViewPainter{}
}

func TestPaintTabBar(t *testing.T) {
	canvas := &recordCanvas{}
	painter := fluent.TabViewPainter{}
	painter.PaintTabBar(canvas, tabview.PaintState{
		Bounds: geometry.NewRect(0, 0, 400, 40),
		Tabs: []tabview.TabState{
			{Label: "Tab 1", Bounds: geometry.NewRect(0, 0, 100, 40), Selected: true},
			{Label: "Tab 2", Bounds: geometry.NewRect(100, 0, 100, 40)},
		},
		SelectedIdx: 0,
	})

	// Should draw background + tab text.
	rects := canvas.methodCalls(methodDrawRect)
	if len(rects) < 1 {
		t.Error("tab bar should draw at least 1 DrawRect (background)")
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) < 2 {
		t.Errorf("tab bar should draw at least 2 DrawText (tab labels), got %d", len(texts))
	}
}

func TestPaintTabBarEmpty(t *testing.T) {
	canvas := &recordCanvas{}
	painter := fluent.TabViewPainter{}
	painter.PaintTabBar(canvas, tabview.PaintState{Bounds: geometry.Rect{}})
	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no calls, got %d", len(canvas.calls))
	}
}

// --- Helpers ---

func assertNonZero(t *testing.T, name string, c widget.Color) {
	t.Helper()
	if c.R == 0 && c.G == 0 && c.B == 0 && c.A == 0 {
		t.Errorf("%s should not be zero-value color", name)
	}
}

func luminance(c widget.Color) float32 {
	return 0.299*c.R + 0.587*c.G + 0.114*c.B
}

func colorEqual(a, b widget.Color) bool {
	const tolerance = 0.05
	return absF32(a.R-b.R) < tolerance &&
		absF32(a.G-b.G) < tolerance &&
		absF32(a.B-b.B) < tolerance
}

func absF32(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}
