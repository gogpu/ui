package textfield_test

import (
	"testing"

	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
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

func TestClipboard_CopyPaste(t *testing.T) {
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
	ctx := widget.NewContext()

	// Double-click on the first character area.
	dbl := event.NewMouseEvent(event.MouseDoubleClick, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(12+2, 24), geometry.Pt(12+2, 24), event.ModNone)
	tf.Event(ctx, dbl)

	start, end := tf.Selection()
	if start != 0 || end != 5 {
		t.Errorf("selection = (%d, %d), want (0, 5) for word 'hello'", start, end)
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

func (p *testPainter) PaintTextField(_ widget.Canvas, ps textfield.PaintState) {
	p.called = true
	p.state = ps
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
	align    float32
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

func (c *recordingCanvas) DrawRoundRect(r geometry.Rect, color widget.Color, radius float32) {
	c.drawRoundRects = append(c.drawRoundRects, drawRoundRectCall{r: r, color: color, radius: radius})
}

func (c *recordingCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {}
func (c *recordingCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)                {}
func (c *recordingCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32)   {}

func (c *recordingCanvas) DrawLine(from, to geometry.Point, color widget.Color, strokeWidth float32) {
	c.drawLines = append(c.drawLines, drawLineCall{from: from, to: to, color: color, strokeWidth: strokeWidth})
}

func (c *recordingCanvas) DrawText(text string, bounds geometry.Rect, fontSize float32, color widget.Color, bold bool, align float32) {
	c.drawTexts = append(c.drawTexts, drawTextCall{text: text, bounds: bounds, fontSize: fontSize, color: color, bold: bold, align: align})
}

func (c *recordingCanvas) PushClip(_ geometry.Rect)       {}
func (c *recordingCanvas) PopClip()                       {}
func (c *recordingCanvas) PushTransform(_ geometry.Point) {}
func (c *recordingCanvas) PopTransform()                  {}

// --- mockCanvas for non-recording tests ---

type mockCanvas struct{}

func (c *mockCanvas) Clear(_ widget.Color)                                                  {}
func (c *mockCanvas) DrawRect(_ geometry.Rect, _ widget.Color)                              {}
func (c *mockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32)                 {}
func (c *mockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32)              {}
func (c *mockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {}
func (c *mockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)                {}
func (c *mockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32)   {}
func (c *mockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32)               {}

func (c *mockCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ float32) {
}

func (c *mockCanvas) PushClip(_ geometry.Rect)       {}
func (c *mockCanvas) PopClip()                       {}
func (c *mockCanvas) PushTransform(_ geometry.Point) {}
func (c *mockCanvas) PopTransform()                  {}
