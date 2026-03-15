package dropdown_test

import (
	"image"
	"testing"

	"github.com/gogpu/ui/core/dropdown"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// --- Construction Tests ---

func TestNew_Defaults(t *testing.T) {
	dd := dropdown.New()

	if !dd.IsVisible() {
		t.Error("default dropdown should be visible")
	}
	if !dd.IsEnabled() {
		t.Error("default dropdown should be enabled")
	}
	if !dd.IsFocusable() {
		t.Error("default dropdown should be focusable")
	}
	if dd.Children() != nil {
		t.Error("dropdown should have no children")
	}
	if dd.SelectedIndex() != -1 {
		t.Errorf("SelectedIndex = %d, want -1", dd.SelectedIndex())
	}
	if dd.SelectedValue() != "" {
		t.Errorf("SelectedValue = %q, want empty", dd.SelectedValue())
	}
	if dd.IsOpen() {
		t.Error("dropdown should not be open by default")
	}
}

func TestNew_WithItems(t *testing.T) {
	dd := dropdown.New(
		dropdown.Items("Red", "Green", "Blue"),
	)

	if dd.SelectedIndex() != -1 {
		t.Errorf("SelectedIndex = %d, want -1", dd.SelectedIndex())
	}
}

func TestNew_WithItemDefs(t *testing.T) {
	items := []dropdown.ItemDef{
		{Value: "r", Label: "Red"},
		{Value: "g", Label: "Green"},
		{Value: "b", Label: "Blue", Disabled: true},
	}
	dd := dropdown.New(dropdown.ItemDefs(items))

	if dd.SelectedIndex() != -1 {
		t.Errorf("SelectedIndex = %d, want -1", dd.SelectedIndex())
	}
}

func TestNew_WithSelected(t *testing.T) {
	dd := dropdown.New(
		dropdown.Items("Red", "Green", "Blue"),
		dropdown.Selected(1),
	)

	if dd.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex = %d, want 1", dd.SelectedIndex())
	}
	if dd.SelectedValue() != "Green" {
		t.Errorf("SelectedValue = %q, want %q", dd.SelectedValue(), "Green")
	}
}

func TestNew_WithPlaceholder(t *testing.T) {
	dd := dropdown.New(
		dropdown.Items("Red", "Green", "Blue"),
		dropdown.Placeholder("Choose color..."),
	)
	dd.SetBounds(geometry.NewRect(0, 0, 200, 48))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	dd.Draw(ctx, canvas)

	if len(canvas.drawTexts) == 0 {
		t.Fatal("should have drawn text")
	}
	if canvas.drawTexts[0].text != "Choose color..." {
		t.Errorf("text = %q, want %q", canvas.drawTexts[0].text, "Choose color...")
	}
}

func TestNew_WithOnChange(t *testing.T) {
	dd := dropdown.New(
		dropdown.Items("Red", "Green", "Blue"),
		dropdown.OnChange(func(index int, value string) {}),
	)
	_ = dd
}

func TestNew_WithDisabled(t *testing.T) {
	dd := dropdown.New(dropdown.Disabled(true))

	if dd.IsFocusable() {
		t.Error("disabled dropdown should not be focusable")
	}
}

func TestNew_WithDisabledFn(t *testing.T) {
	isDisabled := true
	dd := dropdown.New(dropdown.DisabledFn(func() bool { return isDisabled }))

	if dd.IsFocusable() {
		t.Error("disabled dropdown should not be focusable")
	}

	isDisabled = false
	if !dd.IsFocusable() {
		t.Error("enabled dropdown should be focusable")
	}
}

func TestNew_WithMaxVisibleItems(t *testing.T) {
	dd := dropdown.New(
		dropdown.Items("A", "B", "C", "D", "E"),
		dropdown.MaxVisibleItems(3),
	)
	_ = dd
}

func TestNew_WithPainter(t *testing.T) {
	p := &testPainter{}
	dd := dropdown.New(
		dropdown.Items("Red", "Green", "Blue"),
		dropdown.PainterOpt(p),
	)
	dd.SetBounds(geometry.NewRect(0, 0, 200, 48))
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	dd.Draw(ctx, canvas)

	if !p.triggerCalled {
		t.Error("Draw should delegate to configured painter's PaintTrigger")
	}
}

func TestNew_WithSignal(t *testing.T) {
	sig := state.NewSignal(1)
	dd := dropdown.New(
		dropdown.Items("Red", "Green", "Blue"),
		dropdown.SelectedSignal(sig),
	)

	if dd.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex = %d, want 1 (from signal)", dd.SelectedIndex())
	}
}

func TestNew_WithA11yHint(t *testing.T) {
	dd := dropdown.New(dropdown.A11yHint("Select a color"))
	_ = dd
}

func TestNew_AllOptions(t *testing.T) {
	dd := dropdown.New(
		dropdown.Items("Red", "Green", "Blue"),
		dropdown.Selected(0),
		dropdown.Placeholder("Choose..."),
		dropdown.OnChange(func(int, string) {}),
		dropdown.Disabled(false),
		dropdown.MaxVisibleItems(5),
		dropdown.A11yHint("color picker"),
	)

	if !dd.IsFocusable() {
		t.Error("should be focusable")
	}
	if dd.SelectedIndex() != 0 {
		t.Errorf("SelectedIndex = %d, want 0", dd.SelectedIndex())
	}
}

// --- Selection Tests ---

func TestSelectedValue_NoSelection(t *testing.T) {
	dd := dropdown.New(dropdown.Items("Red", "Green", "Blue"))

	if dd.SelectedValue() != "" {
		t.Errorf("SelectedValue = %q, want empty", dd.SelectedValue())
	}
}

func TestSelectedValue_WithSelection(t *testing.T) {
	dd := dropdown.New(
		dropdown.Items("Red", "Green", "Blue"),
		dropdown.Selected(2),
	)

	if dd.SelectedValue() != "Blue" {
		t.Errorf("SelectedValue = %q, want %q", dd.SelectedValue(), "Blue")
	}
}

func TestSetSelectedIndex_Valid(t *testing.T) {
	dd := dropdown.New(dropdown.Items("Red", "Green", "Blue"))

	dd.SetSelectedIndex(1)

	if dd.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex = %d, want 1", dd.SelectedIndex())
	}
	if dd.SelectedValue() != "Green" {
		t.Errorf("SelectedValue = %q, want %q", dd.SelectedValue(), "Green")
	}
}

func TestSetSelectedIndex_MinusOne(t *testing.T) {
	dd := dropdown.New(
		dropdown.Items("Red", "Green", "Blue"),
		dropdown.Selected(1),
	)

	dd.SetSelectedIndex(-1)

	if dd.SelectedIndex() != -1 {
		t.Errorf("SelectedIndex = %d, want -1", dd.SelectedIndex())
	}
}

func TestSetSelectedIndex_OutOfRange(t *testing.T) {
	dd := dropdown.New(
		dropdown.Items("Red", "Green", "Blue"),
		dropdown.Selected(1),
	)

	dd.SetSelectedIndex(10)

	if dd.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex should remain 1, got %d", dd.SelectedIndex())
	}
}

func TestSetSelectedIndex_UpdatesSignal(t *testing.T) {
	sig := state.NewSignal(-1)
	dd := dropdown.New(
		dropdown.Items("Red", "Green", "Blue"),
		dropdown.SelectedSignal(sig),
	)

	dd.SetSelectedIndex(2)

	if sig.Get() != 2 {
		t.Errorf("signal.Get() = %d, want 2", sig.Get())
	}
}

// --- Layout Tests ---

func TestLayout_DefaultSize(t *testing.T) {
	dd := dropdown.New(dropdown.Items("Red", "Green", "Blue"))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 400))

	size := dd.Layout(ctx, constraints)

	if size.Width != 200 {
		t.Errorf("width = %v, want 200", size.Width)
	}
	if size.Height != 48 {
		t.Errorf("height = %v, want 48", size.Height)
	}
}

func TestLayout_TightConstraints(t *testing.T) {
	dd := dropdown.New(dropdown.Items("Red", "Green", "Blue"))
	ctx := widget.NewContext()
	constraints := geometry.Tight(geometry.Sz(100, 36))

	size := dd.Layout(ctx, constraints)

	if size.Width != 100 {
		t.Errorf("width = %v, want 100", size.Width)
	}
	if size.Height != 36 {
		t.Errorf("height = %v, want 36", size.Height)
	}
}

// --- Draw Tests ---

func TestDraw_DefaultPlaceholder(t *testing.T) {
	dd := dropdown.New(dropdown.Items("Red", "Green", "Blue"))
	dd.SetBounds(geometry.NewRect(0, 0, 200, 48))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	dd.Draw(ctx, canvas)

	if len(canvas.drawTexts) == 0 {
		t.Fatal("should have drawn text")
	}
	if canvas.drawTexts[0].text != "Select..." {
		t.Errorf("default placeholder = %q, want %q", canvas.drawTexts[0].text, "Select...")
	}
}

func TestDraw_SelectedItemText(t *testing.T) {
	dd := dropdown.New(
		dropdown.Items("Red", "Green", "Blue"),
		dropdown.Selected(1),
	)
	dd.SetBounds(geometry.NewRect(0, 0, 200, 48))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	dd.Draw(ctx, canvas)

	if len(canvas.drawTexts) == 0 {
		t.Fatal("should have drawn text")
	}
	if canvas.drawTexts[0].text != "Green" {
		t.Errorf("text = %q, want %q", canvas.drawTexts[0].text, "Green")
	}
}

func TestDraw_ItemDefLabel(t *testing.T) {
	items := []dropdown.ItemDef{
		{Value: "r", Label: "Red"},
		{Value: "g", Label: "Green"},
	}
	dd := dropdown.New(
		dropdown.ItemDefs(items),
		dropdown.Selected(0),
	)
	dd.SetBounds(geometry.NewRect(0, 0, 200, 48))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	dd.Draw(ctx, canvas)

	if len(canvas.drawTexts) == 0 {
		t.Fatal("should have drawn text")
	}
	if canvas.drawTexts[0].text != "Red" {
		t.Errorf("text = %q, want %q", canvas.drawTexts[0].text, "Red")
	}
}

func TestDraw_NilCanvas(t *testing.T) {
	dd := dropdown.New(dropdown.Items("Red", "Green", "Blue"))
	dd.SetBounds(geometry.NewRect(0, 0, 200, 48))
	ctx := widget.NewContext()

	// Should not panic.
	dd.Draw(ctx, nil)
}

func TestDraw_DelegatesToPainter(t *testing.T) {
	p := &testPainter{}
	dd := dropdown.New(
		dropdown.Items("Red", "Green", "Blue"),
		dropdown.Selected(0),
		dropdown.PainterOpt(p),
	)
	dd.SetBounds(geometry.NewRect(0, 0, 200, 48))
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	dd.Draw(ctx, canvas)

	if !p.triggerCalled {
		t.Error("should call PaintTrigger on custom painter")
	}
	if p.triggerState.SelectedText != "Red" {
		t.Errorf("PaintTrigger SelectedText = %q, want %q", p.triggerState.SelectedText, "Red")
	}
	if p.triggerState.Bounds.IsEmpty() {
		t.Error("PaintTrigger Bounds should not be empty")
	}
}

// --- Event Handling Tests ---

func TestEvent_DisabledIgnoresAll(t *testing.T) {
	dd := dropdown.New(
		dropdown.Items("Red", "Green", "Blue"),
		dropdown.Disabled(true),
	)
	dd.SetBounds(geometry.NewRect(0, 0, 200, 48))
	ctx := widget.NewContext()

	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(100, 24), geometry.Pt(100, 24), event.ModNone)
	consumed := dd.Event(ctx, press)

	if consumed {
		t.Error("disabled dropdown should not consume events")
	}
}

func TestEvent_MouseEnterSetsHover(t *testing.T) {
	dd := dropdown.New(dropdown.Items("Red", "Green", "Blue"))
	dd.SetBounds(geometry.NewRect(0, 0, 200, 48))
	ctx := widget.NewContext()

	enter := event.NewMouseEvent(event.MouseEnter, event.ButtonNone, 0,
		geometry.Pt(100, 24), geometry.Pt(100, 24), event.ModNone)
	consumed := dd.Event(ctx, enter)

	if !consumed {
		t.Error("should consume MouseEnter")
	}
}

func TestEvent_MouseLeaveClearsHover(t *testing.T) {
	dd := dropdown.New(dropdown.Items("Red", "Green", "Blue"))
	dd.SetBounds(geometry.NewRect(0, 0, 200, 48))
	ctx := widget.NewContext()

	// Enter first.
	enter := event.NewMouseEvent(event.MouseEnter, event.ButtonNone, 0,
		geometry.Pt(100, 24), geometry.Pt(100, 24), event.ModNone)
	dd.Event(ctx, enter)

	// Then leave.
	leave := event.NewMouseEvent(event.MouseLeave, event.ButtonNone, 0,
		geometry.Pt(300, 24), geometry.Pt(300, 24), event.ModNone)
	consumed := dd.Event(ctx, leave)

	if !consumed {
		t.Error("should consume MouseLeave")
	}
}

func TestEvent_ClickTogglesMenu(t *testing.T) {
	dd := dropdown.New(dropdown.Items("Red", "Green", "Blue"))
	dd.SetBounds(geometry.NewRect(0, 0, 200, 48))
	ctx := widget.NewContext()
	setupOverlayManager(ctx)

	// Simulate full click.
	simulateClick(dd, ctx, 100, 24)

	if !dd.IsOpen() {
		t.Error("dropdown should be open after click")
	}
}

func TestEvent_RightClickIgnored(t *testing.T) {
	dd := dropdown.New(dropdown.Items("Red", "Green", "Blue"))
	dd.SetBounds(geometry.NewRect(0, 0, 200, 48))
	ctx := widget.NewContext()

	press := event.NewMouseEvent(event.MousePress, event.ButtonRight, 0,
		geometry.Pt(100, 24), geometry.Pt(100, 24), event.ModNone)
	consumed := dd.Event(ctx, press)

	if consumed {
		t.Error("right click should not be consumed by dropdown")
	}
}

func TestEvent_KeyEnterToggle(t *testing.T) {
	dd := dropdown.New(dropdown.Items("Red", "Green", "Blue"))
	dd.SetBounds(geometry.NewRect(0, 0, 200, 48))
	dd.SetFocused(true)
	ctx := widget.NewContext()
	setupOverlayManager(ctx)

	press := event.NewKeyEvent(event.KeyPress, event.KeyEnter, 0, event.ModNone)
	consumed := dd.Event(ctx, press)

	if !consumed {
		t.Error("Enter should be consumed by focused dropdown")
	}
	if !dd.IsOpen() {
		t.Error("Enter should open the dropdown")
	}
}

func TestEvent_KeySpaceToggle(t *testing.T) {
	dd := dropdown.New(dropdown.Items("Red", "Green", "Blue"))
	dd.SetBounds(geometry.NewRect(0, 0, 200, 48))
	dd.SetFocused(true)
	ctx := widget.NewContext()
	setupOverlayManager(ctx)

	press := event.NewKeyEvent(event.KeyPress, event.KeySpace, 0, event.ModNone)
	consumed := dd.Event(ctx, press)

	if !consumed {
		t.Error("Space should be consumed by focused dropdown")
	}
	if !dd.IsOpen() {
		t.Error("Space should open the dropdown")
	}
}

func TestEvent_KeyDownOpens(t *testing.T) {
	dd := dropdown.New(dropdown.Items("Red", "Green", "Blue"))
	dd.SetBounds(geometry.NewRect(0, 0, 200, 48))
	dd.SetFocused(true)
	ctx := widget.NewContext()
	setupOverlayManager(ctx)

	press := event.NewKeyEvent(event.KeyPress, event.KeyDown, 0, event.ModNone)
	consumed := dd.Event(ctx, press)

	if !consumed {
		t.Error("Down arrow should be consumed by focused dropdown")
	}
	if !dd.IsOpen() {
		t.Error("Down arrow should open the dropdown")
	}
}

func TestEvent_KeyUpOpens(t *testing.T) {
	dd := dropdown.New(dropdown.Items("Red", "Green", "Blue"))
	dd.SetBounds(geometry.NewRect(0, 0, 200, 48))
	dd.SetFocused(true)
	ctx := widget.NewContext()
	setupOverlayManager(ctx)

	press := event.NewKeyEvent(event.KeyPress, event.KeyUp, 0, event.ModNone)
	consumed := dd.Event(ctx, press)

	if !consumed {
		t.Error("Up arrow should be consumed by focused dropdown")
	}
	if !dd.IsOpen() {
		t.Error("Up arrow should open the dropdown")
	}
}

func TestEvent_KeyEscapeCloses(t *testing.T) {
	dd := dropdown.New(dropdown.Items("Red", "Green", "Blue"))
	dd.SetBounds(geometry.NewRect(0, 0, 200, 48))
	dd.SetFocused(true)
	ctx := widget.NewContext()
	setupOverlayManager(ctx)

	// Open first.
	dd.Open(ctx)
	if !dd.IsOpen() {
		t.Fatal("dropdown should be open")
	}

	// Press Escape.
	esc := event.NewKeyEvent(event.KeyPress, event.KeyEscape, 0, event.ModNone)
	consumed := dd.Event(ctx, esc)

	if !consumed {
		t.Error("Escape should be consumed when dropdown is open")
	}
	if dd.IsOpen() {
		t.Error("Escape should close the dropdown")
	}
}

func TestEvent_EscapeNotConsumedWhenClosed(t *testing.T) {
	dd := dropdown.New(dropdown.Items("Red", "Green", "Blue"))
	dd.SetFocused(true)
	ctx := widget.NewContext()

	esc := event.NewKeyEvent(event.KeyPress, event.KeyEscape, 0, event.ModNone)
	consumed := dd.Event(ctx, esc)

	if consumed {
		t.Error("Escape should not be consumed when dropdown is closed")
	}
}

func TestEvent_KeyIgnoredWhenNotFocused(t *testing.T) {
	dd := dropdown.New(dropdown.Items("Red", "Green", "Blue"))
	ctx := widget.NewContext()

	press := event.NewKeyEvent(event.KeyPress, event.KeyEnter, 0, event.ModNone)
	consumed := dd.Event(ctx, press)

	if consumed {
		t.Error("key events should be ignored when not focused")
	}
}

func TestEvent_KeyReleaseIgnored(t *testing.T) {
	dd := dropdown.New(dropdown.Items("Red", "Green", "Blue"))
	dd.SetFocused(true)
	ctx := widget.NewContext()

	release := event.NewKeyEvent(event.KeyRelease, event.KeyEnter, 0, event.ModNone)
	consumed := dd.Event(ctx, release)

	if consumed {
		t.Error("key release should not be consumed")
	}
}

// --- Open/Close Tests ---

func TestOpen_CreatesOverlay(t *testing.T) {
	dd := dropdown.New(dropdown.Items("Red", "Green", "Blue"))
	dd.SetBounds(geometry.NewRect(0, 0, 200, 48))
	ctx := widget.NewContext()
	om := setupOverlayManager(ctx)

	dd.Open(ctx)

	if !dd.IsOpen() {
		t.Error("dropdown should be open after Open()")
	}
	if om.pushCount != 1 {
		t.Errorf("PushOverlay called %d times, want 1", om.pushCount)
	}
}

func TestOpen_NoOpIfAlreadyOpen(t *testing.T) {
	dd := dropdown.New(dropdown.Items("Red", "Green", "Blue"))
	dd.SetBounds(geometry.NewRect(0, 0, 200, 48))
	ctx := widget.NewContext()
	om := setupOverlayManager(ctx)

	dd.Open(ctx)
	dd.Open(ctx) // Should be no-op.

	if om.pushCount != 1 {
		t.Errorf("PushOverlay called %d times, want 1", om.pushCount)
	}
}

func TestOpen_NoOpIfDisabled(t *testing.T) {
	dd := dropdown.New(
		dropdown.Items("Red", "Green", "Blue"),
		dropdown.Disabled(true),
	)
	dd.SetBounds(geometry.NewRect(0, 0, 200, 48))
	ctx := widget.NewContext()
	setupOverlayManager(ctx)

	dd.Open(ctx)

	if dd.IsOpen() {
		t.Error("disabled dropdown should not open")
	}
}

func TestOpen_NoOpWithoutOverlayManager(t *testing.T) {
	dd := dropdown.New(dropdown.Items("Red", "Green", "Blue"))
	dd.SetBounds(geometry.NewRect(0, 0, 200, 48))
	ctx := widget.NewContext()
	// No overlay manager set.

	dd.Open(ctx)

	if dd.IsOpen() {
		t.Error("dropdown should not open without overlay manager")
	}
}

func TestClose(t *testing.T) {
	dd := dropdown.New(dropdown.Items("Red", "Green", "Blue"))
	dd.SetBounds(geometry.NewRect(0, 0, 200, 48))
	ctx := widget.NewContext()
	om := setupOverlayManager(ctx)

	dd.Open(ctx)
	dd.Close(ctx)

	if dd.IsOpen() {
		t.Error("dropdown should be closed after Close()")
	}
	if om.removeCount != 1 {
		t.Errorf("RemoveOverlay called %d times, want 1", om.removeCount)
	}
}

func TestClose_NoOpIfAlreadyClosed(t *testing.T) {
	dd := dropdown.New(dropdown.Items("Red", "Green", "Blue"))
	ctx := widget.NewContext()

	dd.Close(ctx) // Should be no-op.

	if dd.IsOpen() {
		t.Error("dropdown should remain closed")
	}
}

// --- Focus Tests ---

func TestFocus_SetFocused(t *testing.T) {
	dd := dropdown.New(dropdown.Items("Red"))

	dd.SetFocused(true)
	if !dd.IsFocused() {
		t.Error("should be focused after SetFocused(true)")
	}

	dd.SetFocused(false)
	if dd.IsFocused() {
		t.Error("should not be focused after SetFocused(false)")
	}
}

func TestFocusable_VisibleAndEnabled(t *testing.T) {
	dd := dropdown.New(dropdown.Items("Red"))

	if !dd.IsFocusable() {
		t.Error("visible+enabled dropdown should be focusable")
	}

	dd.SetVisible(false)
	if dd.IsFocusable() {
		t.Error("invisible dropdown should not be focusable")
	}

	dd.SetVisible(true)
	dd.SetEnabled(false)
	if dd.IsFocusable() {
		t.Error("disabled dropdown should not be focusable")
	}
}

// --- Accessibility Tests ---

func TestA11y_Role(t *testing.T) {
	dd := dropdown.New()

	if dd.A11yRole() != "combobox" {
		t.Errorf("A11yRole = %q, want %q", dd.A11yRole(), "combobox")
	}
}

func TestA11y_Label(t *testing.T) {
	dd := dropdown.New(dropdown.A11yHint("Select a color"))

	if dd.A11yLabel() != "Select a color" {
		t.Errorf("A11yLabel = %q, want %q", dd.A11yLabel(), "Select a color")
	}
}

func TestA11y_LabelDefault(t *testing.T) {
	dd := dropdown.New()

	if dd.A11yLabel() != "dropdown" {
		t.Errorf("A11yLabel = %q, want %q", dd.A11yLabel(), "dropdown")
	}
}

func TestA11y_Value(t *testing.T) {
	dd := dropdown.New(
		dropdown.Items("Red", "Green", "Blue"),
		dropdown.Selected(1),
	)

	if dd.A11yValue() != "Green" {
		t.Errorf("A11yValue = %q, want %q", dd.A11yValue(), "Green")
	}
}

func TestA11y_ValueEmpty(t *testing.T) {
	dd := dropdown.New(
		dropdown.Items("Red", "Green", "Blue"),
		dropdown.Placeholder("Pick one"),
	)

	if dd.A11yValue() != "Pick one" {
		t.Errorf("A11yValue = %q, want %q", dd.A11yValue(), "Pick one")
	}
}

func TestA11y_Expanded(t *testing.T) {
	dd := dropdown.New(dropdown.Items("Red", "Green", "Blue"))
	dd.SetBounds(geometry.NewRect(0, 0, 200, 48))
	ctx := widget.NewContext()
	setupOverlayManager(ctx)

	if dd.A11yExpanded() {
		t.Error("should not be expanded initially")
	}

	dd.Open(ctx)

	if !dd.A11yExpanded() {
		t.Error("should be expanded after Open()")
	}
}

// --- ItemDef Tests ---

func TestItemDef_DisplayText(t *testing.T) {
	tests := []struct {
		name string
		item dropdown.ItemDef
		want string
	}{
		{"value only", dropdown.ItemDef{Value: "red"}, "red"},
		{"label and value", dropdown.ItemDef{Value: "r", Label: "Red"}, "Red"},
		{"empty value", dropdown.ItemDef{}, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.item.DisplayText()
			if got != tc.want {
				t.Errorf("DisplayText() = %q, want %q", got, tc.want)
			}
		})
	}
}

// --- Widget Interface Compliance ---

func TestWidgetInterface(t *testing.T) {
	var w widget.Widget = dropdown.New()
	_ = w
}

func TestFocusableInterface(t *testing.T) {
	var f widget.Focusable = dropdown.New()
	_ = f
}

// --- Helper functions ---

func simulateClick(dd *dropdown.Widget, ctx widget.Context, x, y float32) {
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(x, y), geometry.Pt(x, y), event.ModNone)
	dd.Event(ctx, press)

	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		geometry.Pt(x, y), geometry.Pt(x, y), event.ModNone)
	dd.Event(ctx, release)
}

func setupOverlayManager(ctx *widget.ContextImpl) *mockOverlayManager {
	om := &mockOverlayManager{}
	ctx.SetOverlayManager(om)
	ctx.SetWindowSize(geometry.Sz(800, 600))
	return om
}

// --- Mock Overlay Manager ---

type mockOverlayManager struct {
	pushCount   int
	popCount    int
	removeCount int
	lastWidget  widget.Widget
}

func (m *mockOverlayManager) PushOverlay(w widget.Widget, _ func()) {
	m.pushCount++
	m.lastWidget = w
}

func (m *mockOverlayManager) PopOverlay() {
	m.popCount++
}

func (m *mockOverlayManager) RemoveOverlay(_ widget.Widget) {
	m.removeCount++
}

// --- Test Painter ---

type testPainter struct {
	triggerCalled bool
	menuCalled    bool
	triggerState  dropdown.TriggerPaintState
	menuState     dropdown.MenuPaintState
}

func (p *testPainter) PaintTrigger(_ widget.Canvas, st *dropdown.TriggerPaintState) {
	p.triggerCalled = true
	p.triggerState = *st
}

func (p *testPainter) PaintMenu(_ widget.Canvas, st *dropdown.MenuPaintState) {
	p.menuCalled = true
	p.menuState = *st
}

// --- Recording Canvas ---

type recordingCanvas struct {
	drawTexts      []drawTextCall
	drawRoundRects []drawRoundRectCall
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

func (c *recordingCanvas) Clear(_ widget.Color)                                  {}
func (c *recordingCanvas) DrawRect(_ geometry.Rect, _ widget.Color)              {}
func (c *recordingCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) {}

func (c *recordingCanvas) DrawRoundRect(r geometry.Rect, color widget.Color, radius float32) {
	c.drawRoundRects = append(c.drawRoundRects, drawRoundRectCall{r: r, color: color, radius: radius})
}

func (c *recordingCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {}
func (c *recordingCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)                {}
func (c *recordingCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32)   {}
func (c *recordingCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32)               {}

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

// --- Mock Canvas ---

type mockCanvas struct{}

func (c *mockCanvas) Clear(_ widget.Color)                                                  {}
func (c *mockCanvas) DrawRect(_ geometry.Rect, _ widget.Color)                              {}
func (c *mockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32)                 {}
func (c *mockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32)              {}
func (c *mockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {}
func (c *mockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)                {}
func (c *mockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32)   {}
func (c *mockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32)               {}

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

// --- Lifecycle Tests ---

func TestLifecycleInterface(t *testing.T) {
	var _ widget.Lifecycle = dropdown.New()
}

func TestMount_CreatesBindings(t *testing.T) {
	sig := state.NewSignal(0)
	dd := dropdown.New(
		dropdown.Items("A", "B", "C"),
		dropdown.SelectedSignal(sig),
	)

	sched := state.NewScheduler(func(_ []widget.Widget) {})
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	dd.Mount(ctx)

	dirtyCount := 0
	sched.SetOnDirty(func() { dirtyCount++ })
	sig.Set(1)

	if dirtyCount == 0 {
		t.Error("signal change should mark widget dirty after mount")
	}
}

func TestUnmount_CleansBindings(t *testing.T) {
	sig := state.NewSignal(0)
	dd := dropdown.New(
		dropdown.Items("A", "B"),
		dropdown.SelectedSignal(sig),
	)

	sched := state.NewScheduler(func(_ []widget.Widget) {})
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	dd.Mount(ctx)
	dd.CleanupBindings()
	dd.Unmount()

	sig.Set(1)

	if sched.PendingCount() != 0 {
		t.Error("signal change after unmount should not mark widget dirty")
	}
}
