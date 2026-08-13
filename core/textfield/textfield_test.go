package textfield_test

import (
	"image"
	"testing"
	"time"

	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/gesture"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// --- Construction Tests ---

func TestNew_Defaults(t *testing.T) {
	tf := textfield.New()

	if !tf.IsVisible() {
		t.Error("default text field should be visible")
	}
	if !tf.IsEnabled() {
		t.Error("default text field should be enabled")
	}
	if !tf.IsFocusable() {
		t.Error("default text field should be focusable")
	}
	if tf.Children() != nil {
		t.Error("text field should have no children")
	}
	if tf.Text() != "" {
		t.Errorf("text = %q, want empty", tf.Text())
	}
	if tf.HasError() {
		t.Error("default text field should not have error")
	}
}

func TestNew_WithPlaceholder(t *testing.T) {
	tf := textfield.New(textfield.Placeholder("Enter email"))
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	tf.Draw(ctx, canvas)

	if len(canvas.drawTexts) == 0 {
		t.Fatal("should have drawn placeholder text")
	}
	if canvas.drawTexts[0].text != "Enter email" {
		t.Errorf("placeholder = %q, want %q", canvas.drawTexts[0].text, "Enter email")
	}
}

func TestNew_WithInitialValue(t *testing.T) {
	tf := textfield.New(textfield.InitialValue("hello"))

	if tf.Text() != "hello" {
		t.Errorf("text = %q, want %q", tf.Text(), "hello")
	}
}

func TestNew_WithDisabled(t *testing.T) {
	tf := textfield.New(textfield.Disabled(true))

	if tf.IsFocusable() {
		t.Error("disabled text field should not be focusable")
	}
}

func TestNew_WithDisabledFn(t *testing.T) {
	isDisabled := true
	tf := textfield.New(textfield.DisabledFn(func() bool { return isDisabled }))

	if tf.IsFocusable() {
		t.Error("disabled text field should not be focusable")
	}

	isDisabled = false
	if !tf.IsFocusable() {
		t.Error("enabled text field should be focusable")
	}
}

func TestNew_WithMaxLength(t *testing.T) {
	tf := textfield.New(
		textfield.InitialValue("abc"),
		textfield.MaxLength(5),
	)
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)
	ctx := widget.NewContext()

	// Type characters up to limit.
	typeRune(tf, ctx, 'd')
	typeRune(tf, ctx, 'e')
	if tf.Text() != "abcde" {
		t.Errorf("text = %q, want %q", tf.Text(), "abcde")
	}

	// Try to type beyond max length.
	typeRune(tf, ctx, 'f')
	if tf.Text() != "abcde" {
		t.Errorf("text = %q after max length, want %q", tf.Text(), "abcde")
	}
}

func TestNew_WithOnChange(t *testing.T) {
	var changed string
	tf := textfield.New(textfield.OnChange(func(v string) { changed = v }))
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)
	ctx := widget.NewContext()

	typeRune(tf, ctx, 'x')

	if changed != "x" {
		t.Errorf("onChange value = %q, want %q", changed, "x")
	}
}

func TestNew_WithOnSubmit(t *testing.T) {
	var submitted string
	tf := textfield.New(
		textfield.InitialValue("test"),
		textfield.OnSubmit(func(v string) { submitted = v }),
	)
	tf.SetFocused(true)
	ctx := widget.NewContext()

	pressKey(tf, ctx, event.KeyEnter, event.ModNone)

	if submitted != "test" {
		t.Errorf("onSubmit value = %q, want %q", submitted, "test")
	}
}

func TestNew_WithInputType(t *testing.T) {
	tests := []struct {
		name      string
		inputType textfield.InputType
		wantStr   string
	}{
		{"Text", textfield.TypeText, "Text"},
		{"Password", textfield.TypePassword, "Password"},
		{"Email", textfield.TypeEmail, "Email"},
		{"Number", textfield.TypeNumber, "Number"},
		{"Search", textfield.TypeSearch, "Search"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = textfield.New(textfield.InputTypeOpt(tt.inputType))
			if tt.inputType.String() != tt.wantStr {
				t.Errorf("InputType.String() = %q, want %q", tt.inputType.String(), tt.wantStr)
			}
		})
	}
}

func TestNew_AllOptions(t *testing.T) {
	tf := textfield.New(
		textfield.Placeholder("Enter text"),
		textfield.InitialValue("hello"),
		textfield.OnChange(func(string) {}),
		textfield.OnSubmit(func(string) {}),
		textfield.InputTypeOpt(textfield.TypeEmail),
		textfield.MaxLength(100),
		textfield.Validation(func(v string) string {
			if v == "" {
				return "required"
			}
			return ""
		}),
		textfield.Disabled(false),
		textfield.A11yLabel("Email address"),
	)

	if !tf.IsFocusable() {
		t.Error("should be focusable")
	}
}

// --- Signal Binding Tests ---

func TestSignal_TwoWayBinding(t *testing.T) {
	sig := state.NewSignal("")
	tf := textfield.New(textfield.ValueSignal(sig))
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)
	ctx := widget.NewContext()

	// Type into the field.
	typeRune(tf, ctx, 'h')
	typeRune(tf, ctx, 'i')

	if sig.Get() != "hi" {
		t.Errorf("signal = %q, want %q", sig.Get(), "hi")
	}

	// Set signal externally.
	sig.Set("external")
	// Draw triggers sync.
	canvas := &recordingCanvas{}
	tf.Draw(ctx, canvas)

	if tf.Text() != "external" {
		t.Errorf("text = %q after signal set, want %q", tf.Text(), "external")
	}
}

// --- Validation Tests ---

func TestValidation_ErrorOnInvalid(t *testing.T) {
	tf := textfield.New(
		textfield.InitialValue(""),
		textfield.Validation(func(v string) string {
			if v == "" {
				return "required"
			}
			return ""
		}),
	)

	if !tf.HasError() {
		t.Error("empty value should trigger validation error")
	}
	if tf.ErrorMessage() != "required" {
		t.Errorf("error = %q, want %q", tf.ErrorMessage(), "required")
	}
}

func TestValidation_ClearsOnValid(t *testing.T) {
	tf := textfield.New(
		textfield.InitialValue(""),
		textfield.Validation(func(v string) string {
			if v == "" {
				return "required"
			}
			return ""
		}),
	)
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)
	ctx := widget.NewContext()

	typeRune(tf, ctx, 'a')

	if tf.HasError() {
		t.Error("valid value should not have error")
	}
}

func TestValidation_MultipleValidators(t *testing.T) {
	tf := textfield.New(
		textfield.InitialValue("ab"),
		textfield.Validation(
			func(v string) string {
				if len(v) < 3 {
					return "too short"
				}
				return ""
			},
			func(v string) string {
				if len(v) > 10 {
					return "too long"
				}
				return ""
			},
		),
	)

	if tf.ErrorMessage() != "too short" {
		t.Errorf("error = %q, want %q", tf.ErrorMessage(), "too short")
	}
}

// --- Text Editing Tests ---

func TestEdit_InsertCharacters(t *testing.T) {
	tf := textfield.New()
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)
	ctx := widget.NewContext()

	typeRune(tf, ctx, 'a')
	typeRune(tf, ctx, 'b')
	typeRune(tf, ctx, 'c')

	if tf.Text() != "abc" {
		t.Errorf("text = %q, want %q", tf.Text(), "abc")
	}
}

func TestEdit_Backspace(t *testing.T) {
	tf := textfield.New(textfield.InitialValue("abc"))
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)
	ctx := widget.NewContext()

	pressKey(tf, ctx, event.KeyBackspace, event.ModNone)

	if tf.Text() != "ab" {
		t.Errorf("text = %q, want %q", tf.Text(), "ab")
	}
}

func TestEdit_BackspaceAtStart(t *testing.T) {
	tf := textfield.New(textfield.InitialValue("abc"))
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)
	ctx := widget.NewContext()

	// Move cursor to start.
	pressKey(tf, ctx, event.KeyHome, event.ModNone)
	pressKey(tf, ctx, event.KeyBackspace, event.ModNone)

	if tf.Text() != "abc" {
		t.Errorf("text = %q, want %q", tf.Text(), "abc")
	}
}

func TestEdit_Delete(t *testing.T) {
	tf := textfield.New(textfield.InitialValue("abc"))
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)
	ctx := widget.NewContext()

	// Move cursor to start, then delete.
	pressKey(tf, ctx, event.KeyHome, event.ModNone)
	pressKey(tf, ctx, event.KeyDelete, event.ModNone)

	if tf.Text() != "bc" {
		t.Errorf("text = %q, want %q", tf.Text(), "bc")
	}
}

func TestEdit_DeleteAtEnd(t *testing.T) {
	tf := textfield.New(textfield.InitialValue("abc"))
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)
	ctx := widget.NewContext()

	pressKey(tf, ctx, event.KeyDelete, event.ModNone)

	if tf.Text() != "abc" {
		t.Errorf("text = %q, want %q", tf.Text(), "abc")
	}
}

func TestEdit_SetText(t *testing.T) {
	tf := textfield.New()
	tf.SetText("hello world")

	if tf.Text() != "hello world" {
		t.Errorf("text = %q, want %q", tf.Text(), "hello world")
	}
}

// --- Cursor Movement Tests ---

func TestCursor_ArrowLeft(t *testing.T) {
	tf := textfield.New(textfield.InitialValue("abc"))
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)
	ctx := widget.NewContext()

	pressKey(tf, ctx, event.KeyLeft, event.ModNone)

	if tf.CursorPosition() != 2 {
		t.Errorf("cursor = %d, want 2", tf.CursorPosition())
	}
}

func TestCursor_ArrowRight(t *testing.T) {
	tf := textfield.New(textfield.InitialValue("abc"))
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)
	ctx := widget.NewContext()

	// Move to start first, then right.
	pressKey(tf, ctx, event.KeyHome, event.ModNone)
	pressKey(tf, ctx, event.KeyRight, event.ModNone)

	if tf.CursorPosition() != 1 {
		t.Errorf("cursor = %d, want 1", tf.CursorPosition())
	}
}

func TestCursor_Home(t *testing.T) {
	tf := textfield.New(textfield.InitialValue("abc"))
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)
	ctx := widget.NewContext()

	pressKey(tf, ctx, event.KeyHome, event.ModNone)

	if tf.CursorPosition() != 0 {
		t.Errorf("cursor = %d, want 0", tf.CursorPosition())
	}
}

func TestCursor_End(t *testing.T) {
	tf := textfield.New(textfield.InitialValue("abc"))
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)
	ctx := widget.NewContext()

	pressKey(tf, ctx, event.KeyHome, event.ModNone)
	pressKey(tf, ctx, event.KeyEnd, event.ModNone)

	if tf.CursorPosition() != 3 {
		t.Errorf("cursor = %d, want 3", tf.CursorPosition())
	}
}

func TestCursor_CtrlLeft_WordJump(t *testing.T) {
	tf := textfield.New(textfield.InitialValue("hello world"))
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)
	ctx := widget.NewContext()

	// Cursor starts at end (position 11).
	pressKey(tf, ctx, event.KeyLeft, event.ModCtrl)

	if tf.CursorPosition() != 6 {
		t.Errorf("cursor = %d, want 6 (start of 'world')", tf.CursorPosition())
	}
}

func TestCursor_CtrlRight_WordJump(t *testing.T) {
	tf := textfield.New(textfield.InitialValue("hello world"))
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)
	ctx := widget.NewContext()

	pressKey(tf, ctx, event.KeyHome, event.ModNone)
	pressKey(tf, ctx, event.KeyRight, event.ModCtrl)

	if tf.CursorPosition() != 5 {
		t.Errorf("cursor = %d, want 5 (end of 'hello')", tf.CursorPosition())
	}
}

// --- Selection Tests ---

func TestSelection_ShiftArrow(t *testing.T) {
	tf := textfield.New(textfield.InitialValue("abc"))
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)
	ctx := widget.NewContext()

	pressKey(tf, ctx, event.KeyLeft, event.ModShift)
	pressKey(tf, ctx, event.KeyLeft, event.ModShift)

	start, end := tf.Selection()
	if start != 1 || end != 3 {
		t.Errorf("selection = (%d, %d), want (1, 3)", start, end)
	}
}

func TestSelection_SelectAll(t *testing.T) {
	tf := textfield.New(textfield.InitialValue("hello"))
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)
	ctx := widget.NewContext()

	pressKey(tf, ctx, event.KeyA, event.ModCtrl)

	start, end := tf.Selection()
	if start != 0 || end != 5 {
		t.Errorf("selection = (%d, %d), want (0, 5)", start, end)
	}
}

func TestSelection_DeleteSelection(t *testing.T) {
	tf := textfield.New(textfield.InitialValue("hello"))
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)
	ctx := widget.NewContext()

	// Select all, then type a character.
	pressKey(tf, ctx, event.KeyA, event.ModCtrl)
	typeRune(tf, ctx, 'x')

	if tf.Text() != "x" {
		t.Errorf("text = %q, want %q", tf.Text(), "x")
	}
}

func TestSelection_BackspaceDeletesSelection(t *testing.T) {
	tf := textfield.New(textfield.InitialValue("hello"))
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)
	ctx := widget.NewContext()

	// Select last 2 chars.
	pressKey(tf, ctx, event.KeyLeft, event.ModShift)
	pressKey(tf, ctx, event.KeyLeft, event.ModShift)
	pressKey(tf, ctx, event.KeyBackspace, event.ModNone)

	if tf.Text() != "hel" {
		t.Errorf("text = %q, want %q", tf.Text(), "hel")
	}
}

// --- Clipboard Tests ---

type testClipboard struct{ text string }

func (c *testClipboard) ClipboardRead() (string, error)   { return c.text, nil }
func (c *testClipboard) ClipboardWrite(text string) error { c.text = text; return nil }

func TestClipboard_CopyPaste(t *testing.T) {
	widget.RegisterClipboardProvider(&testClipboard{})

	tf := textfield.New(textfield.InitialValue("hello"))
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)
	ctx := widget.NewContext()

	// Select all.
	pressKey(tf, ctx, event.KeyA, event.ModCtrl)
	// Copy.
	pressKey(tf, ctx, event.KeyC, event.ModCtrl)
	// Move to end.
	pressKey(tf, ctx, event.KeyEnd, event.ModNone)
	// Paste.
	pressKey(tf, ctx, event.KeyV, event.ModCtrl)

	if tf.Text() != "hellohello" {
		t.Errorf("text = %q, want %q", tf.Text(), "hellohello")
	}
}

func TestClipboard_Cut(t *testing.T) {
	widget.RegisterClipboardProvider(&testClipboard{})

	tf := textfield.New(textfield.InitialValue("hello"))
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)
	ctx := widget.NewContext()

	// Select all.
	pressKey(tf, ctx, event.KeyA, event.ModCtrl)
	// Cut.
	pressKey(tf, ctx, event.KeyX, event.ModCtrl)

	if tf.Text() != "" {
		t.Errorf("text = %q, want empty after cut", tf.Text())
	}

	// Paste.
	pressKey(tf, ctx, event.KeyV, event.ModCtrl)

	if tf.Text() != "hello" {
		t.Errorf("text = %q, want %q after paste", tf.Text(), "hello")
	}
}

// --- Mouse Interaction Tests ---

func TestMouse_ClickFocuses(t *testing.T) {
	tf := textfield.New()
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	ctx := widget.NewContext()

	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(50, 24), geometry.Pt(50, 24), event.ModNone)
	tf.Event(ctx, press)

	if !tf.IsFocused() {
		t.Error("click should focus the text field")
	}
}

func TestMouse_HoverCursor(t *testing.T) {
	tf := textfield.New()
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	ctx := widget.NewContext()

	enter := event.NewMouseEvent(event.MouseEnter, event.ButtonNone, 0,
		geometry.Pt(50, 24), geometry.Pt(50, 24), event.ModNone)
	tf.Event(ctx, enter)

	if ctx.Cursor() != widget.CursorText {
		t.Errorf("cursor = %v, want CursorText", ctx.Cursor())
	}

	leave := event.NewMouseEvent(event.MouseLeave, event.ButtonNone, 0,
		geometry.Pt(-1, -1), geometry.Pt(-1, -1), event.ModNone)
	tf.Event(ctx, leave)

	if ctx.Cursor() != widget.CursorDefault {
		t.Errorf("cursor = %v, want CursorDefault after leave", ctx.Cursor())
	}
}

func TestMouse_DoubleClickSelectsWord(t *testing.T) {
	tf := textfield.New(textfield.InitialValue("hello world"))
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)

	// Double-click is now handled by the TapAndDragRecognizer.
	// Simulate two rapid taps at the same position through the gesture system.
	recs := tf.GestureHitTest(geometry.Pt(0, 0))
	if len(recs) == 0 {
		t.Fatal("TextField should have a TapAndDragRecognizer")
	}
	arena := gesture.NewArena()
	tapPos := geometry.Pt(12+2, 24)
	ts := 100 * time.Millisecond

	// First tap: down + up.
	down1 := &gesture.PointerEvent{
		Base:           event.NewBase(event.TypeMouse, event.ModNone),
		EventType:      gesture.PointerDown,
		PointerID:      1,
		PointerType:    gesture.PointerTypeMouse,
		Position:       tapPos,
		GlobalPosition: tapPos,
		Button:         event.ButtonLeft,
		Buttons:        event.ButtonStateLeft,
		Timestamp:      ts,
	}
	recs[0].AddPointer(down1, arena)
	arena.Close(1)

	up1 := &gesture.PointerEvent{
		Base:           event.NewBase(event.TypeMouse, event.ModNone),
		EventType:      gesture.PointerUp,
		PointerID:      1,
		PointerType:    gesture.PointerTypeMouse,
		Position:       tapPos,
		GlobalPosition: tapPos,
		Timestamp:      ts + 50*time.Millisecond,
	}
	recs[0].HandleEvent(up1)
	arena.Sweep(1)

	// Second tap (within DoubleTapTimeout): down + up.
	ts2 := ts + 100*time.Millisecond
	down2 := &gesture.PointerEvent{
		Base:           event.NewBase(event.TypeMouse, event.ModNone),
		EventType:      gesture.PointerDown,
		PointerID:      1,
		PointerType:    gesture.PointerTypeMouse,
		Position:       tapPos,
		GlobalPosition: tapPos,
		Button:         event.ButtonLeft,
		Buttons:        event.ButtonStateLeft,
		Timestamp:      ts2,
	}
	recs[0].AddPointer(down2, arena)
	arena.Close(1)

	// After the second tap-down with ConsecutiveTapCount=2, word selection
	// should have been triggered by OnTapDown.
	start, end := tf.Selection()
	if start != 0 || end != 5 {
		t.Errorf("selection = (%d, %d), want (0, 5) for word 'hello'", start, end)
	}
}

func TestMouse_TripleClickSelectsAll(t *testing.T) {
	tf := textfield.New(textfield.InitialValue("hello world"))
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)

	recs := tf.GestureHitTest(geometry.Pt(0, 0))
	if len(recs) == 0 {
		t.Fatal("TextField should have a TapAndDragRecognizer")
	}
	arena := gesture.NewArena()
	tapPos := geometry.Pt(14, 24)

	// Three rapid taps at the same position.
	ts := 100 * time.Millisecond
	for i := 0; i < 3; i++ {
		tapTS := ts + time.Duration(i)*100*time.Millisecond
		down := &gesture.PointerEvent{
			Base:           event.NewBase(event.TypeMouse, event.ModNone),
			EventType:      gesture.PointerDown,
			PointerID:      1,
			PointerType:    gesture.PointerTypeMouse,
			Position:       tapPos,
			GlobalPosition: tapPos,
			Button:         event.ButtonLeft,
			Buttons:        event.ButtonStateLeft,
			Timestamp:      tapTS,
		}
		recs[0].AddPointer(down, arena)
		arena.Close(1)

		up := &gesture.PointerEvent{
			Base:           event.NewBase(event.TypeMouse, event.ModNone),
			EventType:      gesture.PointerUp,
			PointerID:      1,
			PointerType:    gesture.PointerTypeMouse,
			Position:       tapPos,
			GlobalPosition: tapPos,
			Timestamp:      tapTS + 30*time.Millisecond,
		}
		recs[0].HandleEvent(up)
		arena.Sweep(1)
	}

	// Triple click selects all text.
	start, end := tf.Selection()
	runeCount := len([]rune("hello world"))
	if start != 0 || end != runeCount {
		t.Errorf("selection = (%d, %d), want (0, %d) for select-all", start, end, runeCount)
	}
}

func TestTextField_GestureAwareInterface(t *testing.T) {
	tf := textfield.New()

	// Verify GestureAware interface is implemented.
	ga, ok := interface{}(tf).(gesture.GestureAware)
	if !ok {
		t.Fatal("TextField should implement gesture.GestureAware")
	}

	recs := ga.GestureHitTest(geometry.Pt(0, 0))
	if len(recs) != 1 {
		t.Errorf("GestureHitTest() returned %d, want 1", len(recs))
	}
}

// --- Disabled State Tests ---

func TestDisabled_BlocksKeyInput(t *testing.T) {
	tf := textfield.New(textfield.Disabled(true))
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)
	ctx := widget.NewContext()

	consumed := typeRune(tf, ctx, 'x')

	if consumed {
		t.Error("disabled text field should not consume key events")
	}
	if tf.Text() != "" {
		t.Error("disabled text field should not accept input")
	}
}

func TestDisabled_BlocksMouseInput(t *testing.T) {
	tf := textfield.New(textfield.Disabled(true))
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	ctx := widget.NewContext()

	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(50, 24), geometry.Pt(50, 24), event.ModNone)
	consumed := tf.Event(ctx, press)

	if consumed {
		t.Error("disabled text field should not consume mouse events")
	}
}

// --- Password Mode Tests ---

func TestPassword_DrawsMaskedText(t *testing.T) {
	tf := textfield.New(
		textfield.InitialValue("secret"),
		textfield.InputTypeOpt(textfield.TypePassword),
	)
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	tf.Draw(ctx, canvas)

	// The drawn text should be bullets, not the actual text.
	for _, dt := range canvas.drawTexts {
		if dt.text == "secret" {
			t.Error("password field should not draw plaintext")
		}
	}
}

// --- Focus Tests ---

func TestFocusable_Interface(t *testing.T) {
	var f widget.Focusable = textfield.New()
	_ = f
}

func TestFocus_SetFocused(t *testing.T) {
	tf := textfield.New()

	tf.SetFocused(true)
	if !tf.IsFocused() {
		t.Error("should be focused after SetFocused(true)")
	}

	tf.SetFocused(false)
	if tf.IsFocused() {
		t.Error("should not be focused after SetFocused(false)")
	}
}

func TestFocusable_VisibleAndEnabled(t *testing.T) {
	tf := textfield.New()

	if !tf.IsFocusable() {
		t.Error("visible+enabled text field should be focusable")
	}

	tf.SetVisible(false)
	if tf.IsFocusable() {
		t.Error("invisible text field should not be focusable")
	}

	tf.SetVisible(true)
	tf.SetEnabled(false)
	if tf.IsFocusable() {
		t.Error("disabled text field should not be focusable")
	}
}

// --- Accessibility Tests ---

func TestA11y_Role(t *testing.T) {
	tf := textfield.New()

	if tf.AccessibleRole() != a11y.RoleTextField {
		t.Errorf("role = %v, want RoleTextField", tf.AccessibleRole())
	}
}

func TestA11y_Label(t *testing.T) {
	tf := textfield.New(
		textfield.Placeholder("Enter email"),
		textfield.A11yLabel("Email address"),
	)

	if tf.AccessibleLabel() != "Email address" {
		t.Errorf("label = %q, want %q", tf.AccessibleLabel(), "Email address")
	}
}

func TestA11y_LabelFallsBackToPlaceholder(t *testing.T) {
	tf := textfield.New(textfield.Placeholder("Enter email"))

	if tf.AccessibleLabel() != "Enter email" {
		t.Errorf("label = %q, want %q", tf.AccessibleLabel(), "Enter email")
	}
}

func TestA11y_PasswordValueMasked(t *testing.T) {
	tf := textfield.New(
		textfield.InitialValue("secret"),
		textfield.InputTypeOpt(textfield.TypePassword),
	)

	val := tf.AccessibleValue()
	if val == "secret" {
		t.Error("password accessible value should be masked")
	}
	if len([]rune(val)) != 6 {
		t.Errorf("masked value length = %d, want 6", len([]rune(val)))
	}
}

// --- Layout Tests ---

func TestLayout_Size(t *testing.T) {
	tf := textfield.New()
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 400))

	size := tf.Layout(ctx, constraints)

	if size.Width < 100 {
		t.Errorf("width = %v, should be at least 100", size.Width)
	}
	if size.Height < 40 {
		t.Errorf("height = %v, should be at least 40", size.Height)
	}
}

func TestLayout_TightConstraints(t *testing.T) {
	tf := textfield.New()
	ctx := widget.NewContext()
	constraints := geometry.Tight(geometry.Sz(200, 50))

	size := tf.Layout(ctx, constraints)

	if size.Width != 200 {
		t.Errorf("width = %v, want 200", size.Width)
	}
	if size.Height != 50 {
		t.Errorf("height = %v, want 50", size.Height)
	}
}

// --- Draw Tests ---

func TestDraw_DelegatesToPainter(t *testing.T) {
	p := &testPainter{}
	tf := textfield.New(
		textfield.Placeholder("Test"),
		textfield.PainterOpt(p),
	)
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	tf.Draw(ctx, canvas)

	if !p.called {
		t.Error("Draw should delegate to the configured painter")
	}
	if p.state.Placeholder != "Test" {
		t.Errorf("PaintState.Placeholder = %q, want %q", p.state.Placeholder, "Test")
	}
}

func TestDraw_DoesNotPanicWithBounds(t *testing.T) {
	tf := textfield.New(textfield.Placeholder("Test"))
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	tf.Draw(ctx, canvas)
}

// --- Pre-computed PaintState Tests (ADR-034) ---

func TestDraw_PrecomputedDisplayText(t *testing.T) {
	p := &testPainter{}
	tf := textfield.New(
		textfield.InitialValue("secret"),
		textfield.InputTypeOpt(textfield.TypePassword),
		textfield.PainterOpt(p),
	)
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	tf.Draw(ctx, canvas)

	if p.state.DisplayText == "secret" {
		t.Error("password DisplayText should be masked, not plaintext")
	}
	if len([]rune(p.state.DisplayText)) != 6 {
		t.Errorf("password DisplayText should have 6 bullets, got %d runes", len([]rune(p.state.DisplayText)))
	}
}

func TestDraw_PrecomputedDisplayText_PlainText(t *testing.T) {
	p := &testPainter{}
	tf := textfield.New(
		textfield.InitialValue("hello"),
		textfield.PainterOpt(p),
	)
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	tf.Draw(ctx, canvas)

	if p.state.DisplayText != "hello" {
		t.Errorf("DisplayText = %q, want %q", p.state.DisplayText, "hello")
	}
}

func TestDraw_PrecomputedContentRect(t *testing.T) {
	p := &testPainter{}
	tf := textfield.New(
		textfield.InitialValue("test"),
		textfield.PainterOpt(p),
	)
	tf.SetBounds(geometry.NewRect(10, 20, 300, 48))
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	tf.Draw(ctx, canvas)

	cr := p.state.ContentRect
	if cr.IsEmpty() {
		t.Error("ContentRect should not be empty")
	}
	// ContentRect should be inside bounds (inset by padding).
	bounds := tf.Bounds()
	if cr.Min.X <= bounds.Min.X || cr.Min.Y <= bounds.Min.Y {
		t.Errorf("ContentRect.Min should be inset from bounds: cr=%v, bounds=%v", cr, bounds)
	}
	if cr.Max.X >= bounds.Max.X || cr.Max.Y >= bounds.Max.Y {
		t.Errorf("ContentRect.Max should be inset from bounds: cr=%v, bounds=%v", cr, bounds)
	}
}

func TestDraw_PrecomputedShowCursor_Focused(t *testing.T) {
	p := &testPainter{}
	tf := textfield.New(
		textfield.InitialValue("hello"),
		textfield.PainterOpt(p),
	)
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	tf.Draw(ctx, canvas)

	if !p.state.ShowCursor {
		t.Error("ShowCursor should be true when focused with no selection")
	}
	if p.state.CursorRect.IsEmpty() {
		t.Error("CursorRect should not be empty when ShowCursor is true")
	}
}

func TestDraw_PrecomputedShowCursor_Unfocused(t *testing.T) {
	p := &testPainter{}
	tf := textfield.New(
		textfield.InitialValue("hello"),
		textfield.PainterOpt(p),
	)
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	// Not focused.
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	tf.Draw(ctx, canvas)

	if p.state.ShowCursor {
		t.Error("ShowCursor should be false when not focused")
	}
}

func TestDraw_PrecomputedShowCursor_Disabled(t *testing.T) {
	p := &testPainter{}
	tf := textfield.New(
		textfield.InitialValue("hello"),
		textfield.Disabled(true),
		textfield.PainterOpt(p),
	)
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	tf.Draw(ctx, canvas)

	if p.state.ShowCursor {
		t.Error("ShowCursor should be false when disabled")
	}
}

func TestDraw_PrecomputedSelection(t *testing.T) {
	p := &testPainter{}
	tf := textfield.New(
		textfield.InitialValue("hello"),
		textfield.PainterOpt(p),
	)
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)
	ctx := widget.NewContext()

	// Select last 2 chars.
	pressKey(tf, ctx, event.KeyLeft, event.ModShift)
	pressKey(tf, ctx, event.KeyLeft, event.ModShift)

	canvas := &mockCanvas{}
	tf.Draw(ctx, canvas)

	if !p.state.ShowSelection {
		t.Error("ShowSelection should be true when selection exists")
	}
	if p.state.SelectionRect.IsEmpty() {
		t.Error("SelectionRect should not be empty when ShowSelection is true")
	}
	if p.state.ShowCursor {
		t.Error("ShowCursor should be false when selection exists")
	}
}

func TestDraw_PrecomputedFontSize(t *testing.T) {
	p := &testPainter{}
	tf := textfield.New(
		textfield.PainterOpt(p),
	)
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	tf.Draw(ctx, canvas)

	if p.state.FontSize <= 0 {
		t.Errorf("FontSize = %v, want > 0", p.state.FontSize)
	}
}

func TestLayoutMetrics_DefaultPainter(t *testing.T) {
	var lm textfield.LayoutMetrics = textfield.DefaultPainter{}

	h, v := lm.ContentPadding()
	if h <= 0 || v <= 0 {
		t.Errorf("ContentPadding = (%v, %v), want positive values", h, v)
	}
	if lm.TextFieldFontSize() <= 0 {
		t.Errorf("TextFieldFontSize = %v, want > 0", lm.TextFieldFontSize())
	}
	if lm.TextFieldCursorWidth() <= 0 {
		t.Errorf("TextFieldCursorWidth = %v, want > 0", lm.TextFieldCursorWidth())
	}
	if lm.TextFieldCornerRadius() <= 0 {
		t.Errorf("TextFieldCornerRadius = %v, want > 0", lm.TextFieldCornerRadius())
	}
}

// --- Widget Interface Compliance ---

func TestWidgetInterface(t *testing.T) {
	var w widget.Widget = textfield.New()
	_ = w
}

func TestFocusableInterface(t *testing.T) {
	var f widget.Focusable = textfield.New()
	_ = f
}

// --- Fluent Styling Tests ---

func TestFluent_Padding(t *testing.T) {
	tf := textfield.New()
	result := tf.Padding(16)

	if result != tf {
		t.Error("fluent methods should return the same widget")
	}
}

// --- Tab Key Propagation ---

func TestTab_NotConsumed(t *testing.T) {
	tf := textfield.New()
	tf.SetFocused(true)
	ctx := widget.NewContext()

	consumed := pressKey(tf, ctx, event.KeyTab, event.ModNone)

	if consumed {
		t.Error("Tab key should not be consumed (for focus navigation)")
	}
}

// --- Edge Cases ---

func TestUnfocused_IgnoresKeyEvents(t *testing.T) {
	tf := textfield.New()
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	ctx := widget.NewContext()

	consumed := typeRune(tf, ctx, 'a')

	if consumed {
		t.Error("unfocused text field should not consume key events")
	}
}

func TestControlChars_Filtered(t *testing.T) {
	tf := textfield.New()
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)
	ctx := widget.NewContext()

	// Try to insert a control character.
	e := event.NewKeyEvent(event.KeyPress, event.KeyUnknown, '\x01', event.ModNone)
	tf.Event(ctx, e)

	if tf.Text() != "" {
		t.Errorf("control chars should be filtered, text = %q", tf.Text())
	}
}

func TestArrowLeft_CollapseSelection(t *testing.T) {
	tf := textfield.New(textfield.InitialValue("hello"))
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)
	ctx := widget.NewContext()

	// Select some text.
	pressKey(tf, ctx, event.KeyLeft, event.ModShift)
	pressKey(tf, ctx, event.KeyLeft, event.ModShift)

	// Arrow left without shift should collapse to selection start.
	pressKey(tf, ctx, event.KeyLeft, event.ModNone)

	start, end := tf.Selection()
	if start != end {
		t.Error("selection should be collapsed after arrow without shift")
	}
	if tf.CursorPosition() != 3 {
		t.Errorf("cursor = %d, want 3", tf.CursorPosition())
	}
}

func TestArrowRight_CollapseSelection(t *testing.T) {
	tf := textfield.New(textfield.InitialValue("hello"))
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)
	ctx := widget.NewContext()

	// Select some text (from end backwards).
	pressKey(tf, ctx, event.KeyLeft, event.ModShift)
	pressKey(tf, ctx, event.KeyLeft, event.ModShift)

	// Arrow right without shift should collapse to selection end.
	pressKey(tf, ctx, event.KeyRight, event.ModNone)

	start, end := tf.Selection()
	if start != end {
		t.Error("selection should be collapsed after arrow without shift")
	}
	if tf.CursorPosition() != 5 {
		t.Errorf("cursor = %d, want 5", tf.CursorPosition())
	}
}

func TestShiftHome_Selection(t *testing.T) {
	tf := textfield.New(textfield.InitialValue("hello"))
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)
	ctx := widget.NewContext()

	pressKey(tf, ctx, event.KeyHome, event.ModShift)

	start, end := tf.Selection()
	if start != 0 || end != 5 {
		t.Errorf("selection = (%d, %d), want (0, 5)", start, end)
	}
}

func TestShiftEnd_Selection(t *testing.T) {
	tf := textfield.New(textfield.InitialValue("hello"))
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)
	ctx := widget.NewContext()

	pressKey(tf, ctx, event.KeyHome, event.ModNone)
	pressKey(tf, ctx, event.KeyEnd, event.ModShift)

	start, end := tf.Selection()
	if start != 0 || end != 5 {
		t.Errorf("selection = (%d, %d), want (0, 5)", start, end)
	}
}

// --- PaintState / ColorScheme Tests ---

func TestPaintState_ColorScheme(t *testing.T) {
	scheme := textfield.TextFieldColorScheme{
		Background:  widget.ColorWhite,
		Border:      widget.ColorGray,
		FocusBorder: widget.ColorBlue,
		ErrorBorder: widget.ColorRed,
		TextColor:   widget.ColorBlack,
		Placeholder: widget.ColorLightGray,
		CursorColor: widget.ColorBlue,
		DisabledBg:  widget.ColorLightGray,
		DisabledFg:  widget.ColorDarkGray,
		SelectionBg: widget.ColorCyan,
		ErrorText:   widget.ColorRed,
	}

	ps := textfield.PaintState{
		Text:    "test",
		Focused: true,
		Bounds:  geometry.NewRect(0, 0, 300, 48),
	}

	// Verify the scheme has the expected values.
	if scheme.FocusBorder != widget.ColorBlue {
		t.Error("FocusBorder should be blue")
	}
	_ = ps
}

// --- Horizontal Scroll Tests (Issue #212) ---

// narrowFieldWidth creates a narrow field width where text will overflow.
// With default painter: contentPaddingH=12 on each side, so content area = 80-24 = 56px.
// With MeasureText returning len(runes)*fontSize*0.5, at fontSize=14 each rune = 7px.
// So 8 runes = 56px fills the content area, 9+ triggers scrolling.
const narrowFieldWidth float32 = 80

func newNarrowField(text string) (*textfield.Widget, widget.Context, *testPainter) {
	p := &testPainter{}
	tf := textfield.New(
		textfield.InitialValue(text),
		textfield.PainterOpt(p),
	)
	tf.SetBounds(geometry.NewRect(0, 0, narrowFieldWidth, 48))
	tf.SetFocused(true)
	ctx := widget.NewContext()
	return tf, ctx, p
}

func TestScroll_NoScrollWhenTextFits(t *testing.T) {
	tf, ctx, p := newNarrowField("short")
	canvas := &mockCanvas{}

	tf.Draw(ctx, canvas)

	if tf.ScrollOffsetX() != 0 {
		t.Errorf("scrollOffsetX = %v, want 0 (text fits)", tf.ScrollOffsetX())
	}
	// Cursor should be within content rect.
	if p.state.ShowCursor && p.state.CursorRect.Min.X < p.state.ContentRect.Min.X {
		t.Error("cursor should be within content rect when text fits")
	}
}

func TestScroll_ScrollsWhenTextOverflows(t *testing.T) {
	// "abcdefghijklmnop" = 16 runes * 7px = 112px, content area ~56px.
	// Cursor starts at end (position 16). Text must scroll left.
	tf, ctx, _ := newNarrowField("abcdefghijklmnop")
	canvas := &mockCanvas{}

	tf.Draw(ctx, canvas)

	if tf.ScrollOffsetX() >= 0 {
		t.Errorf("scrollOffsetX = %v, want < 0 (text overflows, cursor at end)", tf.ScrollOffsetX())
	}
}

func TestScroll_CursorVisibleAfterTyping(t *testing.T) {
	tf, ctx, p := newNarrowField("")
	canvas := &mockCanvas{}

	// Type characters until text overflows the content area.
	for _, r := range "abcdefghijklmnop" {
		typeRune(tf, ctx, r)
	}

	tf.Draw(ctx, canvas)

	// Cursor must be visible within content rect.
	if p.state.ShowCursor {
		cr := p.state.CursorRect
		ct := p.state.ContentRect
		if cr.Min.X < ct.Min.X || cr.Min.X > ct.Max.X {
			t.Errorf("cursor at X=%v outside content rect [%v, %v]",
				cr.Min.X, ct.Min.X, ct.Max.X)
		}
	}
}

func TestScroll_HomeResetsScroll(t *testing.T) {
	tf, ctx, p := newNarrowField("abcdefghijklmnop")
	canvas := &mockCanvas{}

	// Draw once to establish scroll state (cursor at end).
	tf.Draw(ctx, canvas)
	scrollBefore := tf.ScrollOffsetX()
	if scrollBefore >= 0 {
		t.Fatalf("precondition failed: scrollOffsetX = %v, want < 0", scrollBefore)
	}

	// Press Home to go to position 0.
	pressKey(tf, ctx, event.KeyHome, event.ModNone)
	tf.Draw(ctx, canvas)

	// After Home, cursor is at position 0. Scroll should adjust toward 0
	// (showing text from the beginning).
	if tf.ScrollOffsetX() != 0 {
		t.Errorf("scrollOffsetX = %v after Home, want 0", tf.ScrollOffsetX())
	}

	// Cursor should be near the left edge of content rect.
	if p.state.ShowCursor {
		cr := p.state.CursorRect
		ct := p.state.ContentRect
		// Cursor at Home should be at or very near the content rect left edge.
		if cr.Min.X < ct.Min.X || cr.Min.X > ct.Min.X+10 {
			t.Errorf("cursor at Home X=%v, expected near content left %v", cr.Min.X, ct.Min.X)
		}
	}
}

func TestScroll_EndScrollsToShowCursor(t *testing.T) {
	tf, ctx, p := newNarrowField("abcdefghijklmnop")
	canvas := &mockCanvas{}

	// Move to Home first.
	pressKey(tf, ctx, event.KeyHome, event.ModNone)
	tf.Draw(ctx, canvas)
	if tf.ScrollOffsetX() != 0 {
		t.Fatalf("precondition failed: scroll should be 0 after Home")
	}

	// Press End to go to end.
	pressKey(tf, ctx, event.KeyEnd, event.ModNone)
	tf.Draw(ctx, canvas)

	if tf.ScrollOffsetX() >= 0 {
		t.Errorf("scrollOffsetX = %v after End, want < 0", tf.ScrollOffsetX())
	}

	// Cursor at end should be visible within content rect.
	if p.state.ShowCursor {
		cr := p.state.CursorRect
		ct := p.state.ContentRect
		if cr.Min.X > ct.Max.X {
			t.Errorf("cursor at End X=%v exceeds content right edge %v", cr.Min.X, ct.Max.X)
		}
	}
}

func TestScroll_ArrowLeftScrollsBack(t *testing.T) {
	tf, ctx, _ := newNarrowField("abcdefghijklmnop")
	canvas := &mockCanvas{}

	// Cursor starts at end, text is scrolled left.
	tf.Draw(ctx, canvas)
	scrollEnd := tf.ScrollOffsetX()

	// Press left arrow multiple times to move cursor back.
	for i := 0; i < 10; i++ {
		pressKey(tf, ctx, event.KeyLeft, event.ModNone)
	}
	tf.Draw(ctx, canvas)

	// Scroll should have changed (less negative or zero).
	if tf.ScrollOffsetX() <= scrollEnd {
		t.Errorf("scrollOffsetX = %v after leftward movement, expected > %v",
			tf.ScrollOffsetX(), scrollEnd)
	}
}

func TestScroll_BackspaceAdjustsScroll(t *testing.T) {
	tf, ctx, _ := newNarrowField("abcdefghijklmnop")
	canvas := &mockCanvas{}

	// Draw to establish initial scroll.
	tf.Draw(ctx, canvas)

	// Delete all characters via backspace.
	for range 16 {
		pressKey(tf, ctx, event.KeyBackspace, event.ModNone)
	}
	tf.Draw(ctx, canvas)

	// After deleting all text, scroll should reset to 0.
	if tf.ScrollOffsetX() != 0 {
		t.Errorf("scrollOffsetX = %v after deleting all text, want 0", tf.ScrollOffsetX())
	}
}

func TestScroll_OffsetNeverPositive(t *testing.T) {
	tf, ctx, _ := newNarrowField("abcdefghijklmnop")
	canvas := &mockCanvas{}

	// Home.
	pressKey(tf, ctx, event.KeyHome, event.ModNone)
	tf.Draw(ctx, canvas)

	if tf.ScrollOffsetX() > 0 {
		t.Errorf("scrollOffsetX = %v, must never be > 0", tf.ScrollOffsetX())
	}

	// Keep pressing left at position 0.
	pressKey(tf, ctx, event.KeyLeft, event.ModNone)
	tf.Draw(ctx, canvas)

	if tf.ScrollOffsetX() > 0 {
		t.Errorf("scrollOffsetX = %v after left at pos 0, must never be > 0", tf.ScrollOffsetX())
	}
}

func TestScroll_ScrollOffsetXGetter(t *testing.T) {
	tf := textfield.New()
	if tf.ScrollOffsetX() != 0 {
		t.Errorf("new widget scrollOffsetX = %v, want 0", tf.ScrollOffsetX())
	}
}

func TestScroll_CursorRectWithinContentRect(t *testing.T) {
	// Verify cursor rect stays within content rect bounds after scrolling.
	tf, ctx, p := newNarrowField("abcdefghijklmnop")
	canvas := &mockCanvas{}

	// Test at various cursor positions.
	positions := []event.Key{event.KeyHome, event.KeyEnd}
	for _, key := range positions {
		pressKey(tf, ctx, key, event.ModNone)
		tf.Draw(ctx, canvas)

		if p.state.ShowCursor {
			cr := p.state.CursorRect
			ct := p.state.ContentRect
			// CursorRect.Min.X should be within ContentRect horizontal bounds
			// (with a small margin tolerance for scrollMargin).
			if cr.Min.X < ct.Min.X-1 || cr.Min.X > ct.Max.X+1 {
				t.Errorf("key=%v: cursor X=%v outside content rect [%v, %v]",
					key, cr.Min.X, ct.Min.X, ct.Max.X)
			}
		}
	}
}

func TestScroll_ContentRectUnchangedByScroll(t *testing.T) {
	// Verify that ContentRect in PaintState is the original (unscrolled) rect,
	// ensuring painters clip to the correct visible area.
	tf, ctx, p := newNarrowField("abcdefghijklmnop")
	canvas := &mockCanvas{}

	tf.Draw(ctx, canvas)

	// ContentRect should match the bounds minus padding, NOT shifted by scroll.
	bounds := tf.Bounds()
	cr := p.state.ContentRect
	if cr.Min.X <= bounds.Min.X {
		t.Errorf("ContentRect.Min.X=%v should be > bounds.Min.X=%v (padding)", cr.Min.X, bounds.Min.X)
	}
	if cr.Max.X >= bounds.Max.X {
		t.Errorf("ContentRect.Max.X=%v should be < bounds.Max.X=%v (padding)", cr.Max.X, bounds.Max.X)
	}
}

func TestScroll_MouseClickWithScroll(t *testing.T) {
	// When text is scrolled, a click at the left edge of the field
	// should position the cursor at the first visible rune, not rune 0.
	tf, ctx, _ := newNarrowField("abcdefghijklmnop")
	canvas := &mockCanvas{}

	// Draw to establish scroll state (cursor at end, text scrolled left).
	tf.Draw(ctx, canvas)
	if tf.ScrollOffsetX() >= 0 {
		t.Fatalf("precondition: expected scroll < 0 for overflowing text")
	}

	// Click at the left edge of the content area (just inside padding).
	// With scroll active, this should NOT place cursor at position 0.
	leftEdge := geometry.Pt(13, 24) // Just past the 12px left padding.
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		leftEdge, leftEdge, event.ModNone)
	tf.Event(ctx, press)

	// The cursor should be at a position > 0 because text is scrolled.
	if tf.CursorPosition() == 0 {
		t.Error("click at left edge with scroll should not place cursor at position 0")
	}
}

func TestScroll_DeleteReducesScroll(t *testing.T) {
	// After deleting text that makes the remaining text fit, scroll should reset.
	tf, ctx, _ := newNarrowField("abcdefghijklmnop")
	canvas := &mockCanvas{}
	tf.Draw(ctx, canvas)

	// Select all and delete.
	pressKey(tf, ctx, event.KeyA, event.ModCtrl)
	pressKey(tf, ctx, event.KeyBackspace, event.ModNone)
	tf.Draw(ctx, canvas)

	if tf.ScrollOffsetX() != 0 {
		t.Errorf("scrollOffsetX = %v after clearing all text, want 0", tf.ScrollOffsetX())
	}
}

func TestScroll_PasteTriggersScroll(t *testing.T) {
	tf, ctx, _ := newNarrowField("")
	canvas := &mockCanvas{}

	// Type "ab", select all, copy.
	typeRune(tf, ctx, 'a')
	typeRune(tf, ctx, 'b')
	pressKey(tf, ctx, event.KeyA, event.ModCtrl)
	pressKey(tf, ctx, event.KeyC, event.ModCtrl)
	pressKey(tf, ctx, event.KeyEnd, event.ModNone)

	// Paste many times to overflow.
	for range 10 {
		pressKey(tf, ctx, event.KeyV, event.ModCtrl)
	}
	tf.Draw(ctx, canvas)

	// Text should now be "ab" * 10 + "ab" = 22 chars, definitely overflowing.
	if tf.ScrollOffsetX() >= 0 {
		t.Errorf("scrollOffsetX = %v after pasting overflow text, want < 0", tf.ScrollOffsetX())
	}
}

// --- Helper functions ---

func typeRune(tf *textfield.Widget, ctx widget.Context, r rune) bool {
	e := event.NewKeyEvent(event.KeyPress, event.KeyUnknown, r, event.ModNone)
	return tf.Event(ctx, e)
}

func pressKey(tf *textfield.Widget, ctx widget.Context, key event.Key, mods event.Modifiers) bool {
	e := event.NewKeyEvent(event.KeyPress, key, 0, mods)
	return tf.Event(ctx, e)
}

// --- testPainter records calls ---

type testPainter struct {
	called bool
	state  textfield.PaintState
}

func (p *testPainter) PaintTextField(_ widget.Canvas, ps *textfield.PaintState) {
	p.called = true
	p.state = *ps
}

// --- recordingCanvas records draw calls for verification ---

type recordingCanvas struct {
	drawTexts      []drawTextCall
	drawRoundRects []drawRoundRectCall
	drawRects      []drawRectCall
	drawLines      []drawLineCall
}

type drawTextCall struct {
	text     string
	bounds   geometry.Rect
	fontSize float32
	color    widget.Color
	bold     bool
	align    widget.TextAlign
}

type drawRoundRectCall struct {
	r      geometry.Rect
	color  widget.Color
	radius float32
}

type drawRectCall struct {
	r     geometry.Rect
	color widget.Color
}

type drawLineCall struct {
	from, to    geometry.Point
	color       widget.Color
	strokeWidth float32
}

func (c *recordingCanvas) Clear(_ widget.Color)                                  {}
func (c *recordingCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) {}

func (c *recordingCanvas) DrawRect(r geometry.Rect, color widget.Color) {
	c.drawRects = append(c.drawRects, drawRectCall{r: r, color: color})
}

func (c *recordingCanvas) FillRectDirect(_ geometry.Rect, _ widget.Color) {}

func (c *recordingCanvas) DrawRoundRect(r geometry.Rect, color widget.Color, radius float32) {
	c.drawRoundRects = append(c.drawRoundRects, drawRoundRectCall{r: r, color: color, radius: radius})
}

func (c *recordingCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {}
func (c *recordingCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)                {}
func (c *recordingCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32)   {}
func (c *recordingCanvas) StrokeArc(_ geometry.Point, _ float32, _, _ float64, _ widget.Color, _ float32) {
}

func (c *recordingCanvas) DrawLine(from, to geometry.Point, color widget.Color, strokeWidth float32) {
	c.drawLines = append(c.drawLines, drawLineCall{from: from, to: to, color: color, strokeWidth: strokeWidth})
}

func (c *recordingCanvas) DrawText(text string, bounds geometry.Rect, fontSize float32, color widget.Color, bold bool, align widget.TextAlign) {
	c.drawTexts = append(c.drawTexts, drawTextCall{text: text, bounds: bounds, fontSize: fontSize, color: color, bold: bold, align: align})
}

func (c *recordingCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}

func (c *recordingCanvas) DrawImage(_ image.Image, _ geometry.Point)    {}
func (c *recordingCanvas) PushClip(_ geometry.Rect)                     {}
func (c *recordingCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *recordingCanvas) PopClip()                                     {}
func (c *recordingCanvas) PushTransform(_ geometry.Point)               {}
func (c *recordingCanvas) PopTransform()                                {}
func (c *recordingCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }
func (c *recordingCanvas) ScreenOriginBase() geometry.Point             { return geometry.Point{} }
func (c *recordingCanvas) ClipBounds() geometry.Rect                    { return geometry.NewRect(0, 0, 10000, 10000) }
func (c *recordingCanvas) ReplayScene(_ widget.SceneCache)              {}

// --- mockCanvas for non-recording tests ---

type mockCanvas struct{}

func (c *mockCanvas) Clear(_ widget.Color)                                                  {}
func (c *mockCanvas) DrawRect(_ geometry.Rect, _ widget.Color)                              {}
func (c *mockCanvas) FillRectDirect(_ geometry.Rect, _ widget.Color)                        {}
func (c *mockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32)                 {}
func (c *mockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32)              {}
func (c *mockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {}
func (c *mockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)                {}
func (c *mockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32)   {}
func (c *mockCanvas) StrokeArc(_ geometry.Point, _ float32, _, _ float64, _ widget.Color, _ float32) {
}
func (c *mockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32) {}

func (c *mockCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
}

func (c *mockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}

func (c *mockCanvas) DrawImage(_ image.Image, _ geometry.Point)    {}
func (c *mockCanvas) PushClip(_ geometry.Rect)                     {}
func (c *mockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *mockCanvas) PopClip()                                     {}
func (c *mockCanvas) PushTransform(_ geometry.Point)               {}
func (c *mockCanvas) PopTransform()                                {}
func (c *mockCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }
func (c *mockCanvas) ScreenOriginBase() geometry.Point             { return geometry.Point{} }
func (c *mockCanvas) ClipBounds() geometry.Rect                    { return geometry.NewRect(0, 0, 10000, 10000) }
func (c *mockCanvas) ReplayScene(_ widget.SceneCache)              {}

// --- Lifecycle Tests ---

func TestLifecycleInterface(t *testing.T) {
	var _ widget.Lifecycle = textfield.New()
}

func TestMount_CreatesBindings(t *testing.T) {
	sig := state.NewSignal("hello")
	tf := textfield.New(textfield.ValueSignal(sig))

	sched := state.NewScheduler(func(_ []widget.Widget) {})
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	tf.Mount(ctx)

	dirtyCount := 0
	sched.SetOnDirty(func() { dirtyCount++ })
	sig.Set("world")

	if dirtyCount == 0 {
		t.Error("signal change should mark widget dirty after mount")
	}
}

func TestUnmount_CleansBindings(t *testing.T) {
	sig := state.NewSignal("hello")
	tf := textfield.New(textfield.ValueSignal(sig))

	sched := state.NewScheduler(func(_ []widget.Widget) {})
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	tf.Mount(ctx)
	tf.CleanupBindings()
	tf.Unmount()

	sig.Set("world")

	if sched.PendingCount() != 0 {
		t.Error("signal change after unmount should not mark widget dirty")
	}
}
