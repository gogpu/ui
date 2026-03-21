package devtools_test

import (
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
	"github.com/gogpu/ui/theme/devtools"
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

// --- Button painter tests ---

func TestButtonPainterImplementsInterface(t *testing.T) {
	var _ button.Painter = devtools.ButtonPainter{}
}

func TestPaintButtonEmptyBounds(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.ButtonPainter{}
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
	painter := devtools.ButtonPainter{}
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
	painter := devtools.ButtonPainter{}
	painter.PaintButton(canvas, button.PaintState{
		Text:    "Cancel",
		Variant: button.Outlined,
		Bounds:  testBounds(),
	})

	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Errorf("Outlined should draw 1 DrawRoundRect (transparent bg), got %d", len(roundRects))
	}
	if roundRects[0].color != widget.ColorTransparent {
		t.Errorf("Outlined background should be transparent, got %v", roundRects[0].color)
	}

	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) != 1 {
		t.Errorf("Outlined should draw 1 StrokeRoundRect (border), got %d", len(strokes))
	}
}

func TestPaintButtonTextOnly(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.ButtonPainter{}
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
	painter := devtools.ButtonPainter{}
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
	painter := devtools.ButtonPainter{}
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
	painter := devtools.ButtonPainter{}
	painter.PaintButton(canvas, button.PaintState{
		Text:    "Focused",
		Variant: button.Filled,
		Bounds:  testBounds(),
		Focused: true,
	})

	// DevTools focus ring = 1px StrokeRoundRect.
	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) != 1 {
		t.Errorf("focused should draw 1 StrokeRoundRect (focus ring), got %d", len(strokes))
	}
}

func TestPaintButtonWithTheme(t *testing.T) {
	theme := devtools.NewDarkTheme()
	painter := devtools.ButtonPainter{Theme: theme}
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

func TestPaintButtonTonal(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.ButtonPainter{}
	painter.PaintButton(canvas, button.PaintState{
		Text:    "Tonal",
		Variant: button.Tonal,
		Bounds:  testBounds(),
	})

	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Errorf("Tonal should draw 1 DrawRoundRect, got %d", len(roundRects))
	}
}

// --- Checkbox painter tests ---

func TestCheckboxPainterImplementsInterface(t *testing.T) {
	var _ checkbox.Painter = devtools.CheckboxPainter{}
}

func TestPaintCheckboxChecked(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.CheckboxPainter{}
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
	painter := devtools.CheckboxPainter{}
	painter.PaintCheckbox(canvas, checkbox.PaintState{
		Label:  "Opt",
		Bounds: testBounds(),
	})

	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) < 1 {
		t.Error("unchecked checkbox should draw at least 1 StrokeRoundRect (border)")
	}
}

func TestPaintCheckboxIndeterminate(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.CheckboxPainter{}
	painter.PaintCheckbox(canvas, checkbox.PaintState{
		Label:         "Partial",
		Indeterminate: true,
		Bounds:        testBounds(),
	})

	// Should draw the box + dash line.
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) < 1 {
		t.Error("indeterminate checkbox should draw at least 1 DrawRoundRect")
	}

	lines := canvas.methodCalls(methodDrawLine)
	if len(lines) < 1 {
		t.Error("indeterminate checkbox should draw at least 1 DrawLine (dash)")
	}
}

func TestPaintCheckboxEmpty(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.CheckboxPainter{}
	painter.PaintCheckbox(canvas, checkbox.PaintState{Bounds: geometry.Rect{}})
	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no calls, got %d", len(canvas.calls))
	}
}

func TestPaintCheckboxWithTheme(t *testing.T) {
	theme := devtools.NewDarkTheme()
	painter := devtools.CheckboxPainter{Theme: theme}
	canvas := &recordCanvas{}

	painter.PaintCheckbox(canvas, checkbox.PaintState{
		Label:   "Themed",
		Checked: true,
		Bounds:  testBounds(),
	})

	if len(canvas.calls) == 0 {
		t.Fatal("themed painter should produce draw calls")
	}
}

func TestPaintCheckboxFocused(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.CheckboxPainter{}
	painter.PaintCheckbox(canvas, checkbox.PaintState{
		Label:   "Focus",
		Bounds:  testBounds(),
		Focused: true,
	})

	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) < 2 { //nolint:mnd // border + focus ring
		t.Errorf("focused unchecked should draw at least 2 StrokeRoundRect (border + focus ring), got %d", len(strokes))
	}
}

// --- Radio painter tests ---

func TestRadioPainterImplementsInterface(t *testing.T) {
	var _ radio.Painter = devtools.RadioPainter{}
}

func TestPaintRadioSelected(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.RadioPainter{}
	painter.PaintRadio(canvas, radio.PaintState{
		Label:    "Option A",
		Selected: true,
		Bounds:   testBounds(),
	})

	// DevTools selected radio: StrokeCircle (border) + DrawCircle (inner dot).
	strokeCircles := canvas.methodCalls(methodStrokeCircle)
	if len(strokeCircles) < 1 {
		t.Error("selected radio should draw at least 1 StrokeCircle (accent border)")
	}

	circles := canvas.methodCalls(methodDrawCircle)
	if len(circles) < 1 {
		t.Error("selected radio should draw at least 1 DrawCircle (inner dot)")
	}
}

func TestPaintRadioUnselected(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.RadioPainter{}
	painter.PaintRadio(canvas, radio.PaintState{
		Label:  "Option B",
		Bounds: testBounds(),
	})

	strokeCircles := canvas.methodCalls(methodStrokeCircle)
	if len(strokeCircles) < 1 {
		t.Error("unselected radio should draw at least 1 StrokeCircle (border)")
	}
}

func TestPaintRadioEmpty(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.RadioPainter{}
	painter.PaintRadio(canvas, radio.PaintState{Bounds: geometry.Rect{}})
	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no calls, got %d", len(canvas.calls))
	}
}

func TestPaintRadioWithTheme(t *testing.T) {
	theme := devtools.NewDarkTheme()
	painter := devtools.RadioPainter{Theme: theme}
	canvas := &recordCanvas{}

	painter.PaintRadio(canvas, radio.PaintState{
		Label:    "Themed",
		Selected: true,
		Bounds:   testBounds(),
	})

	if len(canvas.calls) == 0 {
		t.Fatal("themed painter should produce draw calls")
	}
}

func TestPaintRadioFocused(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.RadioPainter{}
	painter.PaintRadio(canvas, radio.PaintState{
		Label:   "Focus",
		Bounds:  testBounds(),
		Focused: true,
	})

	strokeCircles := canvas.methodCalls(methodStrokeCircle)
	if len(strokeCircles) < 2 { //nolint:mnd // border + focus ring
		t.Errorf("focused radio should draw at least 2 StrokeCircle (border + focus ring), got %d", len(strokeCircles))
	}
}

// --- TextField painter tests ---

func TestTextFieldPainterImplementsInterface(t *testing.T) {
	var _ textfield.Painter = devtools.TextFieldPainter{}
}

func TestPaintTextField(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.TextFieldPainter{}
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
	painter := devtools.TextFieldPainter{}
	painter.PaintTextField(canvas, textfield.PaintState{Bounds: geometry.Rect{}})
	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no calls, got %d", len(canvas.calls))
	}
}

func TestPaintTextFieldPlaceholder(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.TextFieldPainter{}
	painter.PaintTextField(canvas, textfield.PaintState{
		Text:        "",
		Placeholder: "Enter value...",
		Bounds:      testBounds(),
	})

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) < 1 {
		t.Fatal("placeholder should draw at least 1 DrawText")
	}
	if texts[0].text != "Enter value..." {
		t.Errorf("placeholder text = %q, want 'Enter value...'", texts[0].text)
	}
}

func TestPaintTextFieldFocused(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.TextFieldPainter{}
	painter.PaintTextField(canvas, textfield.PaintState{
		Text:    "Focus",
		Bounds:  testBounds(),
		Focused: true,
	})

	// Should draw cursor line when focused.
	lines := canvas.methodCalls(methodDrawLine)
	if len(lines) < 1 {
		t.Error("focused text field should draw at least 1 DrawLine (cursor)")
	}
}

func TestPaintTextFieldError(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.TextFieldPainter{}
	painter.PaintTextField(canvas, textfield.PaintState{
		Text:     "Bad",
		Bounds:   testBounds(),
		HasError: true,
		ErrorMsg: "Invalid input",
	})

	texts := canvas.methodCalls(methodDrawText)
	foundError := false
	for _, t := range texts {
		if t.text == "Invalid input" {
			foundError = true
			break
		}
	}
	if !foundError {
		t.Error("error text field should draw error message")
	}
}

func TestPaintTextFieldWithTheme(t *testing.T) {
	theme := devtools.NewDarkTheme()
	painter := devtools.TextFieldPainter{Theme: theme}
	canvas := &recordCanvas{}

	painter.PaintTextField(canvas, textfield.PaintState{
		Text:   "Themed",
		Bounds: testBounds(),
	})

	if len(canvas.calls) == 0 {
		t.Fatal("themed painter should produce draw calls")
	}
}

// --- Dropdown painter tests ---

func TestDropdownPainterImplementsInterface(t *testing.T) {
	var _ dropdown.Painter = devtools.DropdownPainter{}
}

func TestPaintDropdownTrigger(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.DropdownPainter{}
	painter.PaintTrigger(canvas, &dropdown.TriggerPaintState{
		Bounds:       testBounds(),
		SelectedText: "Item 1",
	})

	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) < 1 {
		t.Error("trigger should draw at least 1 DrawRoundRect (background)")
	}
}

func TestPaintDropdownTriggerEmpty(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.DropdownPainter{}
	painter.PaintTrigger(canvas, &dropdown.TriggerPaintState{Bounds: geometry.Rect{}})
	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no calls, got %d", len(canvas.calls))
	}
}

func TestPaintDropdownMenu(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.DropdownPainter{}
	painter.PaintMenu(canvas, &dropdown.MenuPaintState{
		Bounds: testBounds(),
		Items: []dropdown.ItemDef{
			{Value: "a", Label: "Alpha"},
			{Value: "b", Label: "Beta"},
		},
		HighlightedIndex: -1,
		SelectedIndex:    0,
		VisibleCount:     2,
		ItemHeight:       28,
	})

	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) < 1 {
		t.Error("menu should draw at least 1 DrawRoundRect (background)")
	}
}

func TestPaintDropdownMenuEmpty(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.DropdownPainter{}
	painter.PaintMenu(canvas, &dropdown.MenuPaintState{Bounds: geometry.Rect{}})
	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no calls, got %d", len(canvas.calls))
	}
}

func TestPaintDropdownWithTheme(t *testing.T) {
	theme := devtools.NewDarkTheme()
	painter := devtools.DropdownPainter{Theme: theme}
	canvas := &recordCanvas{}

	painter.PaintTrigger(canvas, &dropdown.TriggerPaintState{
		Bounds:       testBounds(),
		SelectedText: "Themed",
	})

	if len(canvas.calls) == 0 {
		t.Fatal("themed painter should produce draw calls")
	}
}

// --- Slider painter tests ---

func TestSliderPainterImplementsInterface(t *testing.T) {
	var _ slider.Painter = devtools.SliderPainter{}
}

func TestPaintSlider(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.SliderPainter{}
	painter.PaintSlider(canvas, slider.PaintState{
		Value:    50,
		Min:      0,
		Max:      100,
		Progress: 0.5,
		Bounds:   geometry.NewRect(10, 10, 200, 30),
	})

	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) < 2 { //nolint:mnd // inactive + active track
		t.Errorf("slider should draw at least 2 DrawRoundRect (inactive + active track), got %d", len(roundRects))
	}

	circles := canvas.methodCalls(methodDrawCircle)
	if len(circles) < 1 {
		t.Errorf("slider should draw at least 1 circle (thumb), got %d", len(circles))
	}
}

func TestPaintSliderEmpty(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.SliderPainter{}
	painter.PaintSlider(canvas, slider.PaintState{Bounds: geometry.Rect{}})
	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no calls, got %d", len(canvas.calls))
	}
}

func TestPaintSliderVertical(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.SliderPainter{}
	painter.PaintSlider(canvas, slider.PaintState{
		Value:       50,
		Min:         0,
		Max:         100,
		Progress:    0.5,
		Bounds:      geometry.NewRect(10, 10, 30, 200),
		Orientation: slider.Vertical,
	})

	if len(canvas.calls) == 0 {
		t.Error("vertical slider should produce draw calls")
	}
}

func TestPaintSliderWithMarks(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.SliderPainter{}
	painter.PaintSlider(canvas, slider.PaintState{
		Value:    50,
		Min:      0,
		Max:      100,
		Progress: 0.5,
		Bounds:   geometry.NewRect(10, 10, 200, 30),
		Marks: []slider.Mark{
			{Value: 25},
			{Value: 75},
		},
	})

	circles := canvas.methodCalls(methodDrawCircle)
	if len(circles) < 3 { //nolint:mnd // thumb + 2 marks
		t.Errorf("slider with marks should draw at least 3 circles (thumb + 2 marks), got %d", len(circles))
	}
}

func TestPaintSliderWithTheme(t *testing.T) {
	theme := devtools.NewDarkTheme()
	painter := devtools.SliderPainter{Theme: theme}
	canvas := &recordCanvas{}

	painter.PaintSlider(canvas, slider.PaintState{
		Value:    50,
		Min:      0,
		Max:      100,
		Progress: 0.5,
		Bounds:   geometry.NewRect(10, 10, 200, 30),
	})

	if len(canvas.calls) == 0 {
		t.Fatal("themed painter should produce draw calls")
	}
}

// --- Dialog painter tests ---

func TestDialogPainterImplementsInterface(t *testing.T) {
	var _ dialog.Painter = devtools.DialogPainter{}
}

func TestPaintDialog(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.DialogPainter{}
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
	painter := devtools.DialogPainter{}
	painter.PaintDialog(canvas, dialog.PaintState{Bounds: geometry.Rect{}})
	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no calls, got %d", len(canvas.calls))
	}
}

func TestPaintDialogNoTitle(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.DialogPainter{}
	painter.PaintDialog(canvas, dialog.PaintState{
		Bounds: geometry.NewRect(100, 100, 300, 200),
	})

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 0 {
		t.Errorf("dialog without title should draw 0 DrawText, got %d", len(texts))
	}
}

func TestPaintDialogFocused(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.DialogPainter{}
	painter.PaintDialog(canvas, dialog.PaintState{
		Title:   "Focus",
		Bounds:  geometry.NewRect(100, 100, 300, 200),
		Focused: true,
	})

	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) < 2 { //nolint:mnd // border + focus ring
		t.Errorf("focused dialog should draw at least 2 StrokeRoundRect (border + focus ring), got %d", len(strokes))
	}
}

func TestPaintDialogWithTheme(t *testing.T) {
	theme := devtools.NewDarkTheme()
	painter := devtools.DialogPainter{Theme: theme}
	canvas := &recordCanvas{}

	painter.PaintDialog(canvas, dialog.PaintState{
		Title:  "Themed",
		Bounds: geometry.NewRect(100, 100, 300, 200),
	})

	if len(canvas.calls) == 0 {
		t.Fatal("themed painter should produce draw calls")
	}
}

// --- Scrollbar painter tests ---

func TestScrollbarPainterImplementsInterface(t *testing.T) {
	var _ scrollview.Painter = devtools.ScrollbarPainter{}
}

func TestPaintScrollbar(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.ScrollbarPainter{}
	painter.PaintScrollbar(canvas, scrollview.PaintState{
		Bounds:         geometry.NewRect(10, 10, 200, 300),
		VScrollVisible: true,
		VThumbRect:     geometry.NewRect(198, 20, 6, 60),
		VTrackRect:     geometry.NewRect(198, 10, 6, 300),
	})

	if len(canvas.calls) == 0 {
		t.Error("scrollbar should produce draw calls")
	}
}

func TestPaintScrollbarEmpty(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.ScrollbarPainter{}
	painter.PaintScrollbar(canvas, scrollview.PaintState{Bounds: geometry.Rect{}})
	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no calls, got %d", len(canvas.calls))
	}
}

func TestPaintScrollbarHorizontal(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.ScrollbarPainter{}
	painter.PaintScrollbar(canvas, scrollview.PaintState{
		Bounds:         geometry.NewRect(10, 10, 200, 300),
		HScrollVisible: true,
		HThumbRect:     geometry.NewRect(20, 298, 60, 6),
		HTrackRect:     geometry.NewRect(10, 298, 200, 6),
	})

	if len(canvas.calls) == 0 {
		t.Error("horizontal scrollbar should produce draw calls")
	}
}

func TestPaintScrollbarWithTheme(t *testing.T) {
	theme := devtools.NewDarkTheme()
	painter := devtools.ScrollbarPainter{Theme: theme}
	canvas := &recordCanvas{}

	painter.PaintScrollbar(canvas, scrollview.PaintState{
		Bounds:         geometry.NewRect(10, 10, 200, 300),
		VScrollVisible: true,
		VThumbRect:     geometry.NewRect(198, 20, 6, 60),
		VTrackRect:     geometry.NewRect(198, 10, 6, 300),
	})

	if len(canvas.calls) == 0 {
		t.Fatal("themed painter should produce draw calls")
	}
}

// --- TabView painter tests ---

func TestTabViewPainterImplementsInterface(t *testing.T) {
	var _ tabview.Painter = devtools.TabViewPainter{}
}

func TestPaintTabBar(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.TabViewPainter{}
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
	if len(texts) < 2 { //nolint:mnd // 2 tab labels
		t.Errorf("tab bar should draw at least 2 DrawText (tab labels), got %d", len(texts))
	}
}

func TestPaintTabBarEmpty(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.TabViewPainter{}
	painter.PaintTabBar(canvas, tabview.PaintState{Bounds: geometry.Rect{}})
	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no calls, got %d", len(canvas.calls))
	}
}

func TestPaintTabBarFocused(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.TabViewPainter{}
	painter.PaintTabBar(canvas, tabview.PaintState{
		Bounds: geometry.NewRect(0, 0, 400, 40),
		Tabs: []tabview.TabState{
			{Label: "Tab 1", Bounds: geometry.NewRect(0, 0, 100, 40), Selected: true},
		},
		SelectedIdx: 0,
		Focused:     true,
	})

	strokes := canvas.methodCalls(methodStrokeRect)
	if len(strokes) < 1 {
		t.Error("focused tab bar should draw at least 1 StrokeRect (focus ring)")
	}
}

func TestPaintTabBarWithCloseable(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.TabViewPainter{}
	painter.PaintTabBar(canvas, tabview.PaintState{
		Bounds: geometry.NewRect(0, 0, 400, 40),
		Tabs: []tabview.TabState{
			{
				Label:             "Closeable",
				Bounds:            geometry.NewRect(0, 0, 120, 40),
				Selected:          true,
				Closeable:         true,
				CloseButtonBounds: geometry.NewRect(100, 12, 16, 16),
			},
		},
		SelectedIdx: 0,
	})

	lines := canvas.methodCalls(methodDrawLine)
	if len(lines) < 2 { //nolint:mnd // close button X = 2 lines
		t.Errorf("closeable tab should draw at least 2 DrawLine (X icon), got %d", len(lines))
	}
}

func TestPaintTabBarHover(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.TabViewPainter{}
	painter.PaintTabBar(canvas, tabview.PaintState{
		Bounds: geometry.NewRect(0, 0, 400, 40),
		Tabs: []tabview.TabState{
			{Label: "Tab 1", Bounds: geometry.NewRect(0, 0, 100, 40), Hovered: true},
		},
		SelectedIdx: -1,
	})

	rects := canvas.methodCalls(methodDrawRect)
	if len(rects) < 2 { //nolint:mnd // background + hover
		t.Errorf("hovered tab should draw at least 2 DrawRect (bg + hover), got %d", len(rects))
	}
}

func TestPaintTabBarWithTheme(t *testing.T) {
	theme := devtools.NewDarkTheme()
	painter := devtools.TabViewPainter{Theme: theme}
	canvas := &recordCanvas{}

	painter.PaintTabBar(canvas, tabview.PaintState{
		Bounds: geometry.NewRect(0, 0, 400, 40),
		Tabs: []tabview.TabState{
			{Label: "Themed", Bounds: geometry.NewRect(0, 0, 100, 40), Selected: true},
		},
		SelectedIdx: 0,
	})

	if len(canvas.calls) == 0 {
		t.Fatal("themed painter should produce draw calls")
	}
}

func TestPaintTabBarBottom(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.TabViewPainter{}
	painter.PaintTabBar(canvas, tabview.PaintState{
		Bounds: geometry.NewRect(0, 0, 400, 40),
		Tabs: []tabview.TabState{
			{Label: "Tab 1", Bounds: geometry.NewRect(0, 0, 100, 40), Selected: true},
		},
		SelectedIdx: 0,
		Position:    tabview.Bottom,
	})

	// Indicator should be at the top of the tab for bottom position.
	rects := canvas.methodCalls(methodDrawRect)
	if len(rects) < 2 { //nolint:mnd // background + indicator
		t.Errorf("bottom tab bar should draw at least 2 DrawRect (bg + indicator), got %d", len(rects))
	}
}

// --- Additional coverage tests ---

func TestPaintButtonDisabledOutlined(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.ButtonPainter{}
	painter.PaintButton(canvas, button.PaintState{
		Text:     "Disabled",
		Variant:  button.Outlined,
		Bounds:   testBounds(),
		Disabled: true,
	})

	// Outlined disabled bg should be transparent.
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Errorf("disabled outlined should draw 1 DrawRoundRect (transparent), got %d", len(roundRects))
	}
	if roundRects[0].color != widget.ColorTransparent {
		t.Errorf("disabled outlined bg should be transparent, got %v", roundRects[0].color)
	}
}

func TestPaintButtonPressedState(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.ButtonPainter{}
	painter.PaintButton(canvas, button.PaintState{
		Text:    "Press",
		Variant: button.Filled,
		Bounds:  testBounds(),
		Pressed: true,
	})

	if len(canvas.calls) == 0 {
		t.Error("pressed button should produce draw calls")
	}
}

func TestPaintButtonCustomBackground(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.ButtonPainter{}
	customBg := widget.Hex(0xFF0000)
	painter.PaintButton(canvas, button.PaintState{
		Text:       "Custom",
		Variant:    button.Filled,
		Bounds:     testBounds(),
		Background: &customBg,
	})

	if len(canvas.calls) == 0 {
		t.Error("button with custom bg should produce draw calls")
	}
}

func TestPaintCheckboxDisabled(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.CheckboxPainter{}
	painter.PaintCheckbox(canvas, checkbox.PaintState{
		Label:    "Disabled",
		Checked:  true,
		Bounds:   testBounds(),
		Disabled: true,
	})

	if len(canvas.calls) == 0 {
		t.Error("disabled checkbox should produce draw calls")
	}
}

func TestPaintCheckboxDisabledUnchecked(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.CheckboxPainter{}
	painter.PaintCheckbox(canvas, checkbox.PaintState{
		Label:    "Disabled",
		Bounds:   testBounds(),
		Disabled: true,
	})

	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) < 1 {
		t.Error("disabled unchecked should draw at least 1 StrokeRoundRect (border)")
	}
}

func TestPaintRadioDisabled(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.RadioPainter{}
	painter.PaintRadio(canvas, radio.PaintState{
		Label:    "Disabled",
		Selected: true,
		Bounds:   testBounds(),
		Disabled: true,
	})

	if len(canvas.calls) == 0 {
		t.Error("disabled radio should produce draw calls")
	}
}

func TestPaintTextFieldPassword(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.TextFieldPainter{}
	painter.PaintTextField(canvas, textfield.PaintState{
		Text:      "secret",
		InputType: textfield.TypePassword,
		Bounds:    testBounds(),
	})

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) < 1 {
		t.Fatal("password field should draw text")
	}
	// Should draw bullet characters, not plaintext.
	if texts[0].text == "secret" {
		t.Error("password field should mask text")
	}
}

func TestPaintTextFieldSelection(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.TextFieldPainter{}
	painter.PaintTextField(canvas, textfield.PaintState{
		Text:        "Hello World",
		Bounds:      testBounds(),
		Focused:     true,
		SelectStart: 0,
		SelectEnd:   5,
	})

	// Should draw selection rectangle.
	rects := canvas.methodCalls(methodDrawRect)
	if len(rects) < 1 {
		t.Error("text field with selection should draw at least 1 DrawRect (selection)")
	}
}

func TestPaintTextFieldDisabled(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.TextFieldPainter{}
	painter.PaintTextField(canvas, textfield.PaintState{
		Text:     "Disabled",
		Bounds:   testBounds(),
		Disabled: true,
	})

	if len(canvas.calls) == 0 {
		t.Error("disabled text field should produce draw calls")
	}
}

func TestPaintTextFieldDisabledPlaceholder(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.TextFieldPainter{}
	painter.PaintTextField(canvas, textfield.PaintState{
		Text:        "",
		Placeholder: "Disabled placeholder",
		Bounds:      testBounds(),
		Disabled:    true,
	})

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) < 1 {
		t.Fatal("disabled placeholder should draw text")
	}
}

func TestPaintTextFieldHovered(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.TextFieldPainter{}
	painter.PaintTextField(canvas, textfield.PaintState{
		Text:    "Hovered",
		Bounds:  testBounds(),
		Hovered: true,
	})

	if len(canvas.calls) == 0 {
		t.Error("hovered text field should produce draw calls")
	}
}

func TestPaintDropdownTriggerDisabled(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.DropdownPainter{}
	painter.PaintTrigger(canvas, &dropdown.TriggerPaintState{
		Bounds:       testBounds(),
		SelectedText: "Disabled",
		Disabled:     true,
	})

	if len(canvas.calls) == 0 {
		t.Error("disabled trigger should produce draw calls")
	}
}

func TestPaintDropdownTriggerFocused(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.DropdownPainter{}
	painter.PaintTrigger(canvas, &dropdown.TriggerPaintState{
		Bounds:       testBounds(),
		SelectedText: "Focused",
		Focused:      true,
	})

	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) < 2 { //nolint:mnd // border + focus ring
		t.Errorf("focused trigger should draw at least 2 StrokeRoundRect (border + focus), got %d", len(strokes))
	}
}

func TestPaintDropdownTriggerPlaceholder(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.DropdownPainter{}
	painter.PaintTrigger(canvas, &dropdown.TriggerPaintState{
		Bounds:        testBounds(),
		SelectedText:  "Choose...",
		IsPlaceholder: true,
	})

	if len(canvas.calls) == 0 {
		t.Error("placeholder trigger should produce draw calls")
	}
}

func TestPaintDropdownMenuHighlighted(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.DropdownPainter{}
	painter.PaintMenu(canvas, &dropdown.MenuPaintState{
		Bounds: testBounds(),
		Items: []dropdown.ItemDef{
			{Value: "a", Label: "Alpha"},
			{Value: "b", Label: "Beta", Disabled: true},
		},
		HighlightedIndex: 0,
		SelectedIndex:    -1,
		VisibleCount:     2,
		ItemHeight:       28,
	})

	rects := canvas.methodCalls(methodDrawRect)
	if len(rects) < 1 {
		t.Error("menu with highlighted item should draw at least 1 DrawRect (highlight)")
	}
}

func TestPaintSliderDisabled(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.SliderPainter{}
	painter.PaintSlider(canvas, slider.PaintState{
		Value:    50,
		Min:      0,
		Max:      100,
		Progress: 0.5,
		Bounds:   geometry.NewRect(10, 10, 200, 30),
		Disabled: true,
	})

	if len(canvas.calls) == 0 {
		t.Error("disabled slider should produce draw calls")
	}
}

func TestPaintSliderFocused(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.SliderPainter{}
	painter.PaintSlider(canvas, slider.PaintState{
		Value:    50,
		Min:      0,
		Max:      100,
		Progress: 0.5,
		Bounds:   geometry.NewRect(10, 10, 200, 30),
		Focused:  true,
	})

	strokeCircles := canvas.methodCalls(methodStrokeCircle)
	if len(strokeCircles) < 1 {
		t.Error("focused slider should draw at least 1 StrokeCircle (focus ring)")
	}
}

func TestPaintTabBarDisabledTab(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.TabViewPainter{}
	painter.PaintTabBar(canvas, tabview.PaintState{
		Bounds: geometry.NewRect(0, 0, 400, 40),
		Tabs: []tabview.TabState{
			{Label: "Disabled", Bounds: geometry.NewRect(0, 0, 100, 40), Disabled: true},
		},
		SelectedIdx: -1,
	})

	if len(canvas.calls) == 0 {
		t.Error("tab bar with disabled tab should produce draw calls")
	}
}

// --- All painters with light theme ---

func TestAllPaintersWithLightTheme(t *testing.T) {
	theme := devtools.NewTheme()

	tests := []struct {
		name string
		fn   func(*recordCanvas)
	}{
		{"ButtonPainter", func(c *recordCanvas) {
			devtools.ButtonPainter{Theme: theme}.PaintButton(c, button.PaintState{
				Text: "Light", Variant: button.Filled, Bounds: testBounds(),
			})
		}},
		{"CheckboxPainter", func(c *recordCanvas) {
			devtools.CheckboxPainter{Theme: theme}.PaintCheckbox(c, checkbox.PaintState{
				Label: "Light", Checked: true, Bounds: testBounds(),
			})
		}},
		{"RadioPainter", func(c *recordCanvas) {
			devtools.RadioPainter{Theme: theme}.PaintRadio(c, radio.PaintState{
				Label: "Light", Selected: true, Bounds: testBounds(),
			})
		}},
		{"TextFieldPainter", func(c *recordCanvas) {
			devtools.TextFieldPainter{Theme: theme}.PaintTextField(c, textfield.PaintState{
				Text: "Light", Bounds: testBounds(),
			})
		}},
		{"DropdownPainter", func(c *recordCanvas) {
			devtools.DropdownPainter{Theme: theme}.PaintTrigger(c, &dropdown.TriggerPaintState{
				SelectedText: "Light", Bounds: testBounds(),
			})
		}},
		{"SliderPainter", func(c *recordCanvas) {
			devtools.SliderPainter{Theme: theme}.PaintSlider(c, slider.PaintState{
				Value: 50, Min: 0, Max: 100, Progress: 0.5,
				Bounds: geometry.NewRect(10, 10, 200, 30),
			})
		}},
		{"DialogPainter", func(c *recordCanvas) {
			devtools.DialogPainter{Theme: theme}.PaintDialog(c, dialog.PaintState{
				Title: "Light", Bounds: geometry.NewRect(100, 100, 300, 200),
			})
		}},
		{"ScrollbarPainter", func(c *recordCanvas) {
			devtools.ScrollbarPainter{Theme: theme}.PaintScrollbar(c, scrollview.PaintState{
				Bounds: geometry.NewRect(10, 10, 200, 300), VScrollVisible: true,
				VThumbRect: geometry.NewRect(198, 20, 6, 60),
				VTrackRect: geometry.NewRect(198, 10, 6, 300),
			})
		}},
		{"TabViewPainter", func(c *recordCanvas) {
			devtools.TabViewPainter{Theme: theme}.PaintTabBar(c, tabview.PaintState{
				Bounds: geometry.NewRect(0, 0, 400, 40),
				Tabs: []tabview.TabState{
					{Label: "Light", Bounds: geometry.NewRect(0, 0, 100, 40), Selected: true},
				},
				SelectedIdx: 0,
			})
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canvas := &recordCanvas{}
			tt.fn(canvas)
			if len(canvas.calls) == 0 {
				t.Errorf("%s with light theme should produce draw calls", tt.name)
			}
		})
	}
}
