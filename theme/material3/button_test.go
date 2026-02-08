package material3_test

import (
	"testing"

	"github.com/gogpu/ui/core/button"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/theme/material3"
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
	align       float32
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

func (c *recordCanvas) DrawText(text string, bounds geometry.Rect, fontSize float32, color widget.Color, bold bool, align float32) {
	c.calls = append(c.calls, drawCall{
		method: methodDrawText, text: text, bounds: bounds,
		fontSize: fontSize, color: color, bold: bold, align: align,
	})
}

func (c *recordCanvas) PushClip(r geometry.Rect) {}

func (c *recordCanvas) PopClip() {}

func (c *recordCanvas) PushTransform(offset geometry.Point) {}

func (c *recordCanvas) PopTransform() {}

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

func TestButtonPainterImplementsInterface(t *testing.T) {
	// Compile-time check: assignment would fail if ButtonPainter
	// did not implement button.Painter.
	var _ button.Painter = material3.ButtonPainter{}
}

func TestPaintButtonEmptyBounds(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.ButtonPainter{}

	painter.PaintButton(canvas, button.PaintState{
		Text:    "Click",
		Variant: button.Filled,
		Bounds:  geometry.Rect{}, // empty bounds
	})

	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no draw calls, got %d", len(canvas.calls))
	}
}

func TestPaintButtonFilled(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.ButtonPainter{}

	painter.PaintButton(canvas, button.PaintState{
		Text:    "Submit",
		Variant: button.Filled,
		Bounds:  testBounds(),
	})

	// Should have DrawRoundRect (background) + DrawText.
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

	// Should NOT have StrokeRoundRect.
	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) != 0 {
		t.Errorf("Filled should not draw StrokeRoundRect, got %d", len(strokes))
	}
}

func TestPaintButtonOutlined(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.ButtonPainter{}

	painter.PaintButton(canvas, button.PaintState{
		Text:    "Cancel",
		Variant: button.Outlined,
		Bounds:  testBounds(),
	})

	// Should have DrawRoundRect (transparent bg) + StrokeRoundRect (border) + DrawText.
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Errorf("Outlined should draw 1 DrawRoundRect (transparent bg), got %d", len(roundRects))
	}
	// The background should be transparent.
	if roundRects[0].color != widget.ColorTransparent {
		t.Errorf("Outlined background should be transparent, got %v", roundRects[0].color)
	}

	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) != 1 {
		t.Errorf("Outlined should draw 1 StrokeRoundRect (border), got %d", len(strokes))
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("Outlined should draw 1 DrawText, got %d", len(texts))
	}
	if texts[0].bold {
		t.Error("Outlined variant text should not be bold")
	}
}

func TestPaintButtonTextOnlyNormal(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.ButtonPainter{}

	painter.PaintButton(canvas, button.PaintState{
		Text:    "Learn more",
		Variant: button.TextOnly,
		Bounds:  testBounds(),
		// Hovered: false, Pressed: false -- normal state
	})

	// TextOnly normal: no background, only DrawText.
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 0 {
		t.Errorf("TextOnly normal should draw 0 DrawRoundRect, got %d", len(roundRects))
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("TextOnly should draw 1 DrawText, got %d", len(texts))
	}
}

func TestPaintButtonTextOnlyHover(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.ButtonPainter{}

	painter.PaintButton(canvas, button.PaintState{
		Text:    "Learn more",
		Variant: button.TextOnly,
		Bounds:  testBounds(),
		Hovered: true,
	})

	// TextOnly hover: should draw background.
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Errorf("TextOnly hover should draw 1 DrawRoundRect, got %d", len(roundRects))
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("TextOnly hover should draw 1 DrawText, got %d", len(texts))
	}
}

func TestPaintButtonTonal(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.ButtonPainter{}

	painter.PaintButton(canvas, button.PaintState{
		Text:    "Action",
		Variant: button.Tonal,
		Bounds:  testBounds(),
	})

	// Tonal should draw DrawRoundRect (background) + DrawText.
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Errorf("Tonal should draw 1 DrawRoundRect, got %d", len(roundRects))
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("Tonal should draw 1 DrawText, got %d", len(texts))
	}
	if texts[0].bold {
		t.Error("Tonal variant text should not be bold")
	}
}

func TestPaintButtonDisabled(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.ButtonPainter{}

	painter.PaintButton(canvas, button.PaintState{
		Text:     "Disabled",
		Variant:  button.Filled,
		Bounds:   testBounds(),
		Disabled: true,
	})

	// Disabled Filled: should draw background + text, but no focus ring.
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Errorf("disabled should draw 1 DrawRoundRect, got %d", len(roundRects))
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("disabled should draw 1 DrawText, got %d", len(texts))
	}

	// Should NOT draw focus ring even if focused.
	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) != 0 {
		t.Errorf("disabled should not draw focus ring, got %d StrokeRoundRect", len(strokes))
	}
}

func TestPaintButtonDisabledOutlined(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.ButtonPainter{}

	painter.PaintButton(canvas, button.PaintState{
		Text:     "Disabled",
		Variant:  button.Outlined,
		Bounds:   testBounds(),
		Disabled: true,
	})

	// Disabled Outlined: transparent background for DrawRoundRect, border stroke, text.
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Errorf("disabled outlined should draw 1 DrawRoundRect, got %d", len(roundRects))
	}
	if roundRects[0].color != widget.ColorTransparent {
		t.Errorf("disabled outlined bg should be transparent, got %v", roundRects[0].color)
	}
}

func TestPaintButtonFocused(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.ButtonPainter{}

	painter.PaintButton(canvas, button.PaintState{
		Text:    "Focused",
		Variant: button.Filled,
		Bounds:  testBounds(),
		Focused: true,
	})

	// Focused: should draw background + text + focus ring (StrokeRoundRect).
	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) != 1 {
		t.Errorf("focused should draw 1 StrokeRoundRect (focus ring), got %d", len(strokes))
	}
}

func TestPaintButtonFocusedDisabled(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.ButtonPainter{}

	painter.PaintButton(canvas, button.PaintState{
		Text:     "Focused+Disabled",
		Variant:  button.Filled,
		Bounds:   testBounds(),
		Focused:  true,
		Disabled: true,
	})

	// Focused+Disabled: should NOT draw focus ring.
	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) != 0 {
		t.Errorf("focused+disabled should not draw focus ring, got %d StrokeRoundRect", len(strokes))
	}
}

func TestPaintButtonCustomBackground(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.ButtonPainter{}

	customBg := widget.Hex(0xFF0000) // red
	painter.PaintButton(canvas, button.PaintState{
		Text:       "Custom",
		Variant:    button.Filled,
		Bounds:     testBounds(),
		Background: &customBg,
	})

	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Fatalf("custom bg should draw 1 DrawRoundRect, got %d", len(roundRects))
	}

	// Background should be the custom color (not modified since no hover/press).
	got := roundRects[0].color
	if got != customBg {
		t.Errorf("custom background: got %v, want %v", got, customBg)
	}
}

func TestPaintButtonCustomRadius(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.ButtonPainter{}

	customRadius := float32(20)
	painter.PaintButton(canvas, button.PaintState{
		Text:    "Rounded",
		Variant: button.Filled,
		Bounds:  testBounds(),
		Radius:  &customRadius,
	})

	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Fatalf("custom radius should draw 1 DrawRoundRect, got %d", len(roundRects))
	}

	if roundRects[0].radius != 20 {
		t.Errorf("custom radius: got %f, want 20", roundRects[0].radius)
	}
}

func TestM3ResolvedForeground(t *testing.T) {
	// We test the visible behavior by checking text colors for each variant.
	tests := []struct {
		name     string
		variant  button.Variant
		disabled bool
	}{
		{"Filled enabled", button.Filled, false},
		{"Outlined enabled", button.Outlined, false},
		{"TextOnly enabled", button.TextOnly, false},
		{"Tonal enabled", button.Tonal, false},
		{"Filled disabled", button.Filled, true},
		{"Outlined disabled", button.Outlined, true},
	}

	painter := material3.ButtonPainter{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canvas := &recordCanvas{}
			painter.PaintButton(canvas, button.PaintState{
				Text:     "Test",
				Variant:  tt.variant,
				Bounds:   testBounds(),
				Disabled: tt.disabled,
			})

			texts := canvas.methodCalls(methodDrawText)
			if len(texts) != 1 {
				t.Fatalf("expected 1 DrawText call, got %d", len(texts))
			}

			// All variants should produce a non-transparent text color.
			fg := texts[0].color
			if fg.A == 0 {
				t.Error("foreground color should not be fully transparent")
			}
		})
	}
}

func TestM3DisabledBackground(t *testing.T) {
	painter := material3.ButtonPainter{}

	tests := []struct {
		name        string
		variant     button.Variant
		expectEmpty bool // Outlined/TextOnly disabled = transparent
	}{
		{"Filled disabled", button.Filled, false},
		{"Outlined disabled", button.Outlined, true},
		{"TextOnly disabled", button.TextOnly, true},
		{"Tonal disabled", button.Tonal, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canvas := &recordCanvas{}
			painter.PaintButton(canvas, button.PaintState{
				Text:     "Test",
				Variant:  tt.variant,
				Bounds:   testBounds(),
				Disabled: true,
			})

			roundRects := canvas.methodCalls(methodDrawRoundRect)
			if tt.variant == button.TextOnly {
				// TextOnly disabled + not hovered = no background drawn.
				if len(roundRects) != 0 {
					t.Errorf("TextOnly disabled should draw 0 DrawRoundRect, got %d", len(roundRects))
				}
				return
			}

			if len(roundRects) < 1 {
				t.Fatalf("expected at least 1 DrawRoundRect, got %d", len(roundRects))
			}

			bg := roundRects[0].color
			if tt.expectEmpty && bg != widget.ColorTransparent {
				t.Errorf("expected transparent background, got %v", bg)
			}
		})
	}
}

func TestM3ApplyStateHover(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.ButtonPainter{}

	painter.PaintButton(canvas, button.PaintState{
		Text:    "Hover",
		Variant: button.Filled,
		Bounds:  testBounds(),
		Hovered: true,
	})

	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Fatalf("hover should draw 1 DrawRoundRect, got %d", len(roundRects))
	}

	// Hovered color should be lighter than the base (lerped toward white).
	// We can verify it differs from the non-hovered version.
	canvasNormal := &recordCanvas{}
	painter.PaintButton(canvasNormal, button.PaintState{
		Text:    "Normal",
		Variant: button.Filled,
		Bounds:  testBounds(),
	})

	normalRects := canvasNormal.methodCalls(methodDrawRoundRect)
	if len(normalRects) != 1 {
		t.Fatalf("normal should draw 1 DrawRoundRect, got %d", len(normalRects))
	}

	hoveredColor := roundRects[0].color
	normalColor := normalRects[0].color
	if hoveredColor == normalColor {
		t.Error("hovered color should differ from normal color")
	}
}

func TestM3ApplyStatePress(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.ButtonPainter{}

	painter.PaintButton(canvas, button.PaintState{
		Text:    "Press",
		Variant: button.Filled,
		Bounds:  testBounds(),
		Pressed: true,
	})

	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Fatalf("press should draw 1 DrawRoundRect, got %d", len(roundRects))
	}

	// Pressed color should be darker than the base (lerped toward black).
	canvasNormal := &recordCanvas{}
	painter.PaintButton(canvasNormal, button.PaintState{
		Text:    "Normal",
		Variant: button.Filled,
		Bounds:  testBounds(),
	})

	normalRects := canvasNormal.methodCalls(methodDrawRoundRect)
	if len(normalRects) != 1 {
		t.Fatalf("normal should draw 1 DrawRoundRect, got %d", len(normalRects))
	}

	pressedColor := roundRects[0].color
	normalColor := normalRects[0].color
	if pressedColor == normalColor {
		t.Error("pressed color should differ from normal color")
	}
}

func TestM3FontSize(t *testing.T) {
	painter := material3.ButtonPainter{}

	tests := []struct {
		name     string
		size     button.Size
		wantSize float32
	}{
		{"Small", button.Small, 12},
		{"Medium", button.Medium, 14},
		{"Large", button.Large, 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canvas := &recordCanvas{}
			painter.PaintButton(canvas, button.PaintState{
				Text:    "Test",
				Variant: button.Filled,
				Size:    tt.size,
				Bounds:  testBounds(),
			})

			texts := canvas.methodCalls(methodDrawText)
			if len(texts) != 1 {
				t.Fatalf("expected 1 DrawText call, got %d", len(texts))
			}

			if texts[0].fontSize != tt.wantSize {
				t.Errorf("font size for %s: got %f, want %f",
					tt.name, texts[0].fontSize, tt.wantSize)
			}
		})
	}
}

func TestPaintButtonTextOnlyPressed(t *testing.T) {
	canvas := &recordCanvas{}
	painter := material3.ButtonPainter{}

	painter.PaintButton(canvas, button.PaintState{
		Text:    "Press me",
		Variant: button.TextOnly,
		Bounds:  testBounds(),
		Pressed: true,
	})

	// TextOnly pressed: should draw background (same as hover).
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Errorf("TextOnly pressed should draw 1 DrawRoundRect, got %d", len(roundRects))
	}
}

func TestButtonPainterWithThemeSameAsDefault(t *testing.T) {
	// ButtonPainter with the default M3 purple seed should produce identical
	// output to a nil-Theme painter (which uses hardcoded M3 purple defaults).
	defaultSeed := widget.Hex(0x6750A4)
	painterWithTheme := material3.ButtonPainter{Theme: material3.New(defaultSeed)}
	painterNilTheme := material3.ButtonPainter{}

	canvasA := &recordCanvas{}
	canvasB := &recordCanvas{}

	state := button.PaintState{
		Text:    "Test",
		Variant: button.Filled,
		Bounds:  testBounds(),
	}

	painterWithTheme.PaintButton(canvasA, state)
	painterNilTheme.PaintButton(canvasB, state)

	// Both should produce drawing calls (background + text).
	if len(canvasA.calls) == 0 {
		t.Fatal("themed painter should produce draw calls")
	}
	if len(canvasB.calls) == 0 {
		t.Fatal("nil-theme painter should produce draw calls")
	}

	// Both should produce the same number of calls.
	if len(canvasA.calls) != len(canvasB.calls) {
		t.Errorf("call count mismatch: themed=%d, nil-theme=%d",
			len(canvasA.calls), len(canvasB.calls))
	}
}

func TestButtonPainterNilThemeFallback(t *testing.T) {
	// A nil-Theme ButtonPainter should use the default M3 purple palette
	// and produce valid draw calls.
	painter := material3.ButtonPainter{}
	canvas := &recordCanvas{}

	painter.PaintButton(canvas, button.PaintState{
		Text:    "Default",
		Variant: button.Filled,
		Bounds:  testBounds(),
	})

	// Should produce DrawRoundRect (bg) + DrawText.
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 1 {
		t.Fatalf("nil-theme Filled should draw 1 DrawRoundRect, got %d", len(roundRects))
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("nil-theme Filled should draw 1 DrawText, got %d", len(texts))
	}

	// Text should be white (M3 on-primary default).
	fg := texts[0].color
	if fg != widget.ColorWhite {
		t.Errorf("nil-theme Filled text color should be white, got %v", fg)
	}
}

func TestButtonPainterRedTheme(t *testing.T) {
	// A red-seeded theme should produce colors that differ from the default purple.
	redTheme := material3.New(widget.Hex(0xFF0000))
	purpleTheme := material3.New(widget.Hex(0x6750A4))

	painterRed := material3.ButtonPainter{Theme: redTheme}
	painterPurple := material3.ButtonPainter{Theme: purpleTheme}

	canvasRed := &recordCanvas{}
	canvasPurple := &recordCanvas{}

	state := button.PaintState{
		Text:    "Color",
		Variant: button.Filled,
		Bounds:  testBounds(),
	}

	painterRed.PaintButton(canvasRed, state)
	painterPurple.PaintButton(canvasPurple, state)

	redRects := canvasRed.methodCalls(methodDrawRoundRect)
	purpleRects := canvasPurple.methodCalls(methodDrawRoundRect)

	if len(redRects) != 1 || len(purpleRects) != 1 {
		t.Fatalf("both should draw 1 DrawRoundRect, got red=%d purple=%d",
			len(redRects), len(purpleRects))
	}

	// The filled background colors should differ between red and purple themes.
	if redRects[0].color == purpleRects[0].color {
		t.Error("red and purple themes should produce different filled backgrounds")
	}
}
