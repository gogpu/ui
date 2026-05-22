package cupertino_test

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
	"github.com/gogpu/ui/theme/cupertino"
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

func TestNewThemeDefaults(t *testing.T) {
	theme := cupertino.NewTheme()
	if theme == nil {
		t.Fatal("NewTheme should return non-nil")
	}
	if theme.IsDark() {
		t.Error("NewTheme should create light theme")
	}
	if theme.Radius <= 0 {
		t.Error("Radius should be positive")
	}
	// Default accent is system blue.
	accent := theme.Colors.Accent
	if accent.A == 0 {
		t.Error("accent color should not be fully transparent")
	}
}

func TestNewDarkTheme(t *testing.T) {
	theme := cupertino.NewDarkTheme()
	if theme == nil {
		t.Fatal("NewDarkTheme should return non-nil")
	}
	if !theme.IsDark() {
		t.Error("NewDarkTheme should create dark theme")
	}
}

func TestWithAccentColor(t *testing.T) {
	green := widget.Hex(0x34C759)
	theme := cupertino.NewTheme(cupertino.WithAccentColor(green))
	if theme.Colors.Accent != green {
		t.Errorf("accent should be green, got %v", theme.Colors.Accent)
	}
}

func TestOnSurface(t *testing.T) {
	theme := cupertino.NewTheme()
	onSurface := theme.OnSurface()
	if onSurface.A == 0 {
		t.Error("OnSurface should not be fully transparent")
	}
}

// --- ButtonPainter tests ---

func TestButtonPainterImplementsInterface(t *testing.T) {
	var _ button.Painter = cupertino.ButtonPainter{}
}

func TestPaintButtonEmptyBounds(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.ButtonPainter{}

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
	painter := cupertino.ButtonPainter{}

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
	if !texts[0].bold {
		t.Error("Filled variant text should be bold")
	}
}

func TestPaintButtonOutlined(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.ButtonPainter{}

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
	painter := cupertino.ButtonPainter{}

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

func TestPaintButtonTextOnlyNormal(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.ButtonPainter{}

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

func TestPaintButtonDisabled(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.ButtonPainter{}

	painter.PaintButton(canvas, button.PaintState{
		Text:     "Disabled",
		Variant:  button.Filled,
		Bounds:   testBounds(),
		Disabled: true,
	})

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("disabled should draw 1 DrawText, got %d", len(texts))
	}

	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) != 0 {
		t.Errorf("disabled should not draw focus ring, got %d StrokeRoundRect", len(strokes))
	}
}

func TestPaintButtonFocused(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.ButtonPainter{}

	painter.PaintButton(canvas, button.PaintState{
		Text:    "Focused",
		Variant: button.Filled,
		Bounds:  testBounds(),
		Focused: true,
	})

	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) != 1 {
		t.Errorf("focused should draw 1 StrokeRoundRect (focus ring), got %d", len(strokes))
	}
}

func TestPaintButtonFocusedDisabled(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.ButtonPainter{}

	painter.PaintButton(canvas, button.PaintState{
		Text:     "Focused+Disabled",
		Variant:  button.Filled,
		Bounds:   testBounds(),
		Focused:  true,
		Disabled: true,
	})

	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) != 0 {
		t.Errorf("focused+disabled should not draw focus ring, got %d StrokeRoundRect", len(strokes))
	}
}

func TestPaintButtonWithTheme(t *testing.T) {
	theme := cupertino.NewTheme()
	painter := cupertino.ButtonPainter{Theme: theme}
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

func TestPaintButtonCustomAccent(t *testing.T) {
	green := widget.Hex(0x34C759)
	theme := cupertino.NewTheme(cupertino.WithAccentColor(green))
	painterGreen := cupertino.ButtonPainter{Theme: theme}
	painterDefault := cupertino.ButtonPainter{}

	canvasGreen := &recordCanvas{}
	canvasDefault := &recordCanvas{}

	state := button.PaintState{
		Text:    "Color",
		Variant: button.Filled,
		Bounds:  testBounds(),
	}

	painterGreen.PaintButton(canvasGreen, state)
	painterDefault.PaintButton(canvasDefault, state)

	greenRects := canvasGreen.methodCalls(methodDrawRoundRect)
	defaultRects := canvasDefault.methodCalls(methodDrawRoundRect)

	if len(greenRects) != 1 || len(defaultRects) != 1 {
		t.Fatalf("both should draw 1 DrawRoundRect, got green=%d default=%d",
			len(greenRects), len(defaultRects))
	}

	if greenRects[0].color == defaultRects[0].color {
		t.Error("green and default themes should produce different backgrounds")
	}
}

func TestPaintButtonHoverState(t *testing.T) {
	canvas := &recordCanvas{}
	canvasNormal := &recordCanvas{}
	painter := cupertino.ButtonPainter{}

	painter.PaintButton(canvas, button.PaintState{
		Text: "Hover", Variant: button.Filled, Bounds: testBounds(), Hovered: true,
	})
	painter.PaintButton(canvasNormal, button.PaintState{
		Text: "Normal", Variant: button.Filled, Bounds: testBounds(),
	})

	hoverRects := canvas.methodCalls(methodDrawRoundRect)
	normalRects := canvasNormal.methodCalls(methodDrawRoundRect)

	if len(hoverRects) != 1 || len(normalRects) != 1 {
		t.Fatalf("both should draw 1 DrawRoundRect")
	}

	if hoverRects[0].color == normalRects[0].color {
		t.Error("hovered color should differ from normal")
	}
}

func TestPaintButtonFontSizes(t *testing.T) {
	painter := cupertino.ButtonPainter{}

	tests := []struct {
		name string
		size button.Size
	}{
		{"Small", button.Small},
		{"Medium", button.Medium},
		{"Large", button.Large},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canvas := &recordCanvas{}
			painter.PaintButton(canvas, button.PaintState{
				Text: "Test", Variant: button.Filled, Size: tt.size, Bounds: testBounds(),
			})

			texts := canvas.methodCalls(methodDrawText)
			if len(texts) != 1 {
				t.Fatalf("expected 1 DrawText call, got %d", len(texts))
			}
			if texts[0].fontSize <= 0 {
				t.Error("font size should be positive")
			}
		})
	}
}

// --- CheckboxPainter tests (iOS toggle switch) ---

func TestCheckboxPainterImplementsInterface(t *testing.T) {
	var _ checkbox.Painter = cupertino.CheckboxPainter{}
}

func TestPaintCheckboxEmptyBounds(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.CheckboxPainter{}

	painter.PaintCheckbox(canvas, checkbox.PaintState{Bounds: geometry.Rect{}})

	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no draw calls, got %d", len(canvas.calls))
	}
}

func TestPaintCheckboxChecked(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.CheckboxPainter{}

	painter.PaintCheckbox(canvas, checkbox.PaintState{
		Label:   "Toggle",
		Checked: true,
		Bounds:  testBounds(),
	})

	// Toggle ON: track (DrawRoundRect) + thumb (DrawCircle) + label (DrawText).
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) < 1 {
		t.Errorf("checked toggle should draw track, got %d DrawRoundRect", len(roundRects))
	}

	circles := canvas.methodCalls(methodDrawCircle)
	if len(circles) < 1 {
		t.Errorf("checked toggle should draw thumb, got %d DrawCircle", len(circles))
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("should draw 1 DrawText for label, got %d", len(texts))
	}
	if texts[0].text != "Toggle" {
		t.Errorf("label should be 'Toggle', got %q", texts[0].text)
	}
}

func TestPaintCheckboxUnchecked(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.CheckboxPainter{}

	painter.PaintCheckbox(canvas, checkbox.PaintState{
		Checked: false,
		Bounds:  testBounds(),
	})

	// Toggle OFF: track outline (StrokeRoundRect) + thumb (DrawCircle).
	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) < 1 {
		t.Errorf("unchecked toggle should draw track outline, got %d StrokeRoundRect", len(strokes))
	}

	circles := canvas.methodCalls(methodDrawCircle)
	if len(circles) < 1 {
		t.Errorf("unchecked toggle should draw thumb, got %d DrawCircle", len(circles))
	}
}

func TestPaintCheckboxIndeterminate(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.CheckboxPainter{}

	painter.PaintCheckbox(canvas, checkbox.PaintState{
		Indeterminate: true,
		Bounds:        testBounds(),
	})

	// Indeterminate: track fill + thumb + dash line.
	lines := canvas.methodCalls(methodDrawLine)
	if len(lines) < 1 {
		t.Errorf("indeterminate should draw dash line, got %d DrawLine", len(lines))
	}
}

func TestPaintCheckboxWithTheme(t *testing.T) {
	theme := cupertino.NewTheme()
	painter := cupertino.CheckboxPainter{Theme: theme}
	canvas := &recordCanvas{}

	painter.PaintCheckbox(canvas, checkbox.PaintState{
		Checked: true,
		Bounds:  testBounds(),
	})

	if len(canvas.calls) == 0 {
		t.Fatal("themed checkbox painter should produce draw calls")
	}
}

// --- RadioPainter tests ---

func TestRadioPainterImplementsInterface(t *testing.T) {
	var _ radio.Painter = cupertino.RadioPainter{}
}

func TestPaintRadioEmptyBounds(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.RadioPainter{}

	painter.PaintRadio(canvas, radio.PaintState{Bounds: geometry.Rect{}})

	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no draw calls, got %d", len(canvas.calls))
	}
}

func TestPaintRadioSelected(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.RadioPainter{}

	painter.PaintRadio(canvas, radio.PaintState{
		Label:    "Option A",
		Selected: true,
		Bounds:   testBounds(),
	})

	// Selected: outer circle fill + inner circle + label text.
	circles := canvas.methodCalls(methodDrawCircle)
	if len(circles) < 2 {
		t.Errorf("selected radio should draw 2 circles (outer+inner), got %d", len(circles))
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("should draw 1 DrawText for label, got %d", len(texts))
	}
}

func TestPaintRadioUnselected(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.RadioPainter{}

	painter.PaintRadio(canvas, radio.PaintState{
		Selected: false,
		Bounds:   testBounds(),
	})

	// Unselected: border circle only.
	strokeCircles := canvas.methodCalls(methodStrokeCircle)
	if len(strokeCircles) < 1 {
		t.Errorf("unselected radio should draw border circle, got %d StrokeCircle", len(strokeCircles))
	}
}

func TestPaintRadioFocused(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.RadioPainter{}

	painter.PaintRadio(canvas, radio.PaintState{
		Selected: true,
		Focused:  true,
		Bounds:   testBounds(),
	})

	strokeCircles := canvas.methodCalls(methodStrokeCircle)
	if len(strokeCircles) < 1 {
		t.Errorf("focused radio should draw focus ring, got %d StrokeCircle", len(strokeCircles))
	}
}

// --- TextFieldPainter tests ---

func TestTextFieldPainterImplementsInterface(t *testing.T) {
	var _ textfield.Painter = cupertino.TextFieldPainter{}
}

func TestPaintTextFieldEmptyBounds(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.TextFieldPainter{}

	painter.PaintTextField(canvas, textfield.PaintState{Bounds: geometry.Rect{}})

	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no draw calls, got %d", len(canvas.calls))
	}
}

func TestPaintTextFieldNormal(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.TextFieldPainter{}

	painter.PaintTextField(canvas, textfield.PaintState{
		Text:   "Hello",
		Bounds: testBounds(),
	})

	// Background + border + text.
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) < 1 {
		t.Errorf("should draw background, got %d DrawRoundRect", len(roundRects))
	}

	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) < 1 {
		t.Errorf("should draw border, got %d StrokeRoundRect", len(strokes))
	}
}

func TestPaintTextFieldPlaceholder(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.TextFieldPainter{}

	painter.PaintTextField(canvas, textfield.PaintState{
		Text:        "",
		Placeholder: "Enter text...",
		Bounds:      testBounds(),
	})

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) < 1 {
		t.Fatalf("should draw placeholder text, got %d DrawText", len(texts))
	}
	if texts[0].text != "Enter text..." {
		t.Errorf("placeholder text should be 'Enter text...', got %q", texts[0].text)
	}
}

func TestPaintTextFieldWithTheme(t *testing.T) {
	theme := cupertino.NewTheme()
	painter := cupertino.TextFieldPainter{Theme: theme}
	canvas := &recordCanvas{}

	painter.PaintTextField(canvas, textfield.PaintState{
		Text:   "Themed",
		Bounds: testBounds(),
	})

	if len(canvas.calls) == 0 {
		t.Fatal("themed text field painter should produce draw calls")
	}
}

// --- DropdownPainter tests ---

func TestDropdownPainterImplementsInterface(t *testing.T) {
	var _ dropdown.Painter = cupertino.DropdownPainter{}
}

func TestPaintDropdownTriggerEmptyBounds(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.DropdownPainter{}

	painter.PaintTrigger(canvas, &dropdown.TriggerPaintState{})

	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no draw calls, got %d", len(canvas.calls))
	}
}

func TestPaintDropdownTriggerNormal(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.DropdownPainter{}

	painter.PaintTrigger(canvas, &dropdown.TriggerPaintState{
		Bounds:       testBounds(),
		SelectedText: "Option 1",
	})

	// Background + border + text + chevron (2 lines).
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) < 1 {
		t.Errorf("should draw background, got %d DrawRoundRect", len(roundRects))
	}

	lines := canvas.methodCalls(methodDrawLine)
	if len(lines) < 2 {
		t.Errorf("should draw chevron (2 lines), got %d DrawLine", len(lines))
	}
}

func TestPaintDropdownMenuNormal(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.DropdownPainter{}

	painter.PaintMenu(canvas, &dropdown.MenuPaintState{
		Bounds:       geometry.NewRect(10, 50, 120, 120),
		Items:        []dropdown.ItemDef{{Value: "a"}, {Value: "b"}},
		ItemHeight:   40,
		VisibleCount: 3,
	})

	// Menu bg + border + item texts.
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) < 1 {
		t.Errorf("should draw menu background, got %d DrawRoundRect", len(roundRects))
	}
}

// --- SliderPainter tests ---

func TestSliderPainterImplementsInterface(t *testing.T) {
	var _ slider.Painter = cupertino.SliderPainter{}
}

func TestPaintSliderEmptyBounds(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.SliderPainter{}

	painter.PaintSlider(canvas, slider.PaintState{Bounds: geometry.Rect{}})

	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no draw calls, got %d", len(canvas.calls))
	}
}

func TestPaintSliderHorizontal(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.SliderPainter{}

	painter.PaintSlider(canvas, slider.PaintState{
		Value:       50,
		Min:         0,
		Max:         100,
		Progress:    0.5,
		Bounds:      geometry.NewRect(10, 10, 200, 40),
		Orientation: slider.Horizontal,
	})

	// Inactive track + active track + thumb circle + thumb border.
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) < 2 {
		t.Errorf("should draw inactive+active tracks, got %d DrawRoundRect", len(roundRects))
	}

	circles := canvas.methodCalls(methodDrawCircle)
	if len(circles) < 1 {
		t.Errorf("should draw thumb circle, got %d DrawCircle", len(circles))
	}
}

func TestPaintSliderVertical(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.SliderPainter{}

	painter.PaintSlider(canvas, slider.PaintState{
		Value:       50,
		Min:         0,
		Max:         100,
		Progress:    0.5,
		Bounds:      geometry.NewRect(10, 10, 40, 200),
		Orientation: slider.Vertical,
	})

	if len(canvas.calls) == 0 {
		t.Fatal("vertical slider should produce draw calls")
	}
}

func TestPaintSliderDisabled(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.SliderPainter{}

	painter.PaintSlider(canvas, slider.PaintState{
		Value:    50,
		Min:      0,
		Max:      100,
		Progress: 0.5,
		Bounds:   geometry.NewRect(10, 10, 200, 40),
		Disabled: true,
	})

	if len(canvas.calls) == 0 {
		t.Fatal("disabled slider should still produce draw calls")
	}
}

// --- DialogPainter tests ---

func TestDialogPainterImplementsInterface(t *testing.T) {
	var _ dialog.Painter = cupertino.DialogPainter{}
}

func TestPaintDialogEmptyBounds(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.DialogPainter{}

	painter.PaintDialog(canvas, dialog.PaintState{Bounds: geometry.Rect{}})

	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no draw calls, got %d", len(canvas.calls))
	}
}

func TestPaintDialogNormal(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.DialogPainter{}

	painter.PaintDialog(canvas, dialog.PaintState{
		Title:  "Alert",
		Bounds: geometry.NewRect(50, 50, 300, 200),
		Actions: []dialog.Action{
			{Label: "OK"},
			{Label: "Cancel"},
		},
	})

	// Surface + title text + action texts.
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) < 1 {
		t.Errorf("should draw dialog surface, got %d DrawRoundRect", len(roundRects))
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) < 3 {
		t.Errorf("should draw title + 2 action texts, got %d DrawText", len(texts))
	}
}

func TestPaintDialogFocused(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.DialogPainter{}

	painter.PaintDialog(canvas, dialog.PaintState{
		Title:   "Focused",
		Bounds:  geometry.NewRect(50, 50, 300, 200),
		Focused: true,
	})

	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) < 1 {
		t.Errorf("focused dialog should draw focus ring, got %d StrokeRoundRect", len(strokes))
	}
}

// --- ScrollbarPainter tests ---

func TestScrollbarPainterImplementsInterface(t *testing.T) {
	var _ scrollview.Painter = cupertino.ScrollbarPainter{}
}

func TestPaintScrollbarEmptyBounds(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.ScrollbarPainter{}

	painter.PaintScrollbar(canvas, scrollview.PaintState{Bounds: geometry.Rect{}})

	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no draw calls, got %d", len(canvas.calls))
	}
}

func TestPaintScrollbarVertical(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.ScrollbarPainter{}

	painter.PaintScrollbar(canvas, scrollview.PaintState{
		Bounds:         geometry.NewRect(0, 0, 200, 400),
		VScrollVisible: true,
		VTrackRect:     geometry.NewRect(190, 0, 8, 400),
		VThumbRect:     geometry.NewRect(190, 0, 8, 100),
	})

	if len(canvas.calls) == 0 {
		t.Fatal("scrollbar painter should produce draw calls")
	}
}

func TestPaintScrollbarWithTheme(t *testing.T) {
	theme := cupertino.NewTheme()
	painter := cupertino.ScrollbarPainter{Theme: theme}
	canvas := &recordCanvas{}

	painter.PaintScrollbar(canvas, scrollview.PaintState{
		Bounds:         geometry.NewRect(0, 0, 200, 400),
		VScrollVisible: true,
		VTrackRect:     geometry.NewRect(190, 0, 8, 400),
		VThumbRect:     geometry.NewRect(190, 0, 8, 100),
	})

	if len(canvas.calls) == 0 {
		t.Fatal("themed scrollbar painter should produce draw calls")
	}
}

// --- TabViewPainter tests (segmented control) ---

func TestTabViewPainterImplementsInterface(t *testing.T) {
	var _ tabview.Painter = cupertino.TabViewPainter{}
}

func TestPaintTabBarEmptyBounds(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.TabViewPainter{}

	painter.PaintTabBar(canvas, tabview.PaintState{Bounds: geometry.Rect{}})

	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no draw calls, got %d", len(canvas.calls))
	}
}

func TestPaintTabBarSegmentedControl(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.TabViewPainter{}

	painter.PaintTabBar(canvas, tabview.PaintState{
		Bounds: geometry.NewRect(10, 10, 300, 32),
		Tabs: []tabview.TabState{
			{Label: "First", Bounds: geometry.NewRect(10, 10, 100, 32), Selected: true},
			{Label: "Second", Bounds: geometry.NewRect(110, 10, 100, 32)},
			{Label: "Third", Bounds: geometry.NewRect(210, 10, 100, 32)},
		},
		SelectedIdx: 0,
	})

	// Background pill + selected segment + texts.
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) < 2 {
		t.Errorf("should draw background pill + selected segment, got %d DrawRoundRect", len(roundRects))
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 3 {
		t.Errorf("should draw 3 tab labels, got %d DrawText", len(texts))
	}
}

func TestPaintTabBarFocused(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.TabViewPainter{}

	painter.PaintTabBar(canvas, tabview.PaintState{
		Bounds: geometry.NewRect(10, 10, 300, 32),
		Tabs: []tabview.TabState{
			{Label: "Tab", Bounds: geometry.NewRect(10, 10, 100, 32)},
		},
		Focused: true,
	})

	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) < 1 {
		t.Errorf("focused tab bar should draw focus ring, got %d StrokeRoundRect", len(strokes))
	}
}

func TestPaintTabBarWithCloseable(t *testing.T) {
	canvas := &recordCanvas{}
	painter := cupertino.TabViewPainter{}

	closeBtn := geometry.NewRect(80, 15, 16, 16)
	painter.PaintTabBar(canvas, tabview.PaintState{
		Bounds: geometry.NewRect(10, 10, 200, 32),
		Tabs: []tabview.TabState{
			{
				Label:             "Closeable",
				Bounds:            geometry.NewRect(10, 10, 100, 32),
				Closeable:         true,
				CloseButtonBounds: closeBtn,
			},
		},
		SelectedIdx: -1,
	})

	// Should draw close button X (2 lines).
	lines := canvas.methodCalls(methodDrawLine)
	if len(lines) < 2 {
		t.Errorf("closeable tab should draw X icon (2 lines), got %d DrawLine", len(lines))
	}
}

func TestPaintTabBarWithTheme(t *testing.T) {
	theme := cupertino.NewTheme()
	painter := cupertino.TabViewPainter{Theme: theme}
	canvas := &recordCanvas{}

	painter.PaintTabBar(canvas, tabview.PaintState{
		Bounds: geometry.NewRect(10, 10, 200, 32),
		Tabs: []tabview.TabState{
			{Label: "Tab 1", Bounds: geometry.NewRect(10, 10, 100, 32), Selected: true},
		},
		SelectedIdx: 0,
	})

	if len(canvas.calls) == 0 {
		t.Fatal("themed tabview painter should produce draw calls")
	}
}

// --- Dark theme variant tests ---

func TestDarkThemeButtonColors(t *testing.T) {
	darkTheme := cupertino.NewDarkTheme()
	lightTheme := cupertino.NewTheme()

	painterDark := cupertino.ButtonPainter{Theme: darkTheme}
	painterLight := cupertino.ButtonPainter{Theme: lightTheme}

	canvasDark := &recordCanvas{}
	canvasLight := &recordCanvas{}

	state := button.PaintState{
		Text: "Test", Variant: button.Filled, Bounds: testBounds(),
	}

	painterDark.PaintButton(canvasDark, state)
	painterLight.PaintButton(canvasLight, state)

	// Both should produce draw calls.
	if len(canvasDark.calls) == 0 {
		t.Fatal("dark theme painter should produce draw calls")
	}
	if len(canvasLight.calls) == 0 {
		t.Fatal("light theme painter should produce draw calls")
	}
}

func TestDarkThemeCheckboxColors(t *testing.T) {
	darkTheme := cupertino.NewDarkTheme()
	painter := cupertino.CheckboxPainter{Theme: darkTheme}
	canvas := &recordCanvas{}

	painter.PaintCheckbox(canvas, checkbox.PaintState{
		Checked: true,
		Bounds:  testBounds(),
	})

	if len(canvas.calls) == 0 {
		t.Fatal("dark theme checkbox painter should produce draw calls")
	}
}
