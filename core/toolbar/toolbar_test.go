package toolbar

import (
	"image"
	"testing"

	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/widget"
)

// --- Item Tests ---

func TestItemKind_String(t *testing.T) {
	tests := []struct {
		kind ItemKind
		want string
	}{
		{ItemButton, "Button"},
		{ItemSeparator, "Separator"},
		{ItemSpacer, "Spacer"},
		{ItemCustom, "Custom"},
		{ItemKind(99), "Unknown"},
	}
	for _, tc := range tests {
		if got := tc.kind.String(); got != tc.want {
			t.Errorf("ItemKind(%d).String() = %q, want %q", tc.kind, got, tc.want)
		}
	}
}

func TestIconButton_Constructor(t *testing.T) {
	called := false
	item := IconButton("New", icon.Add, func() { called = true })

	if item.Kind != ItemButton {
		t.Errorf("Kind = %v, want ItemButton", item.Kind)
	}
	if item.Label != "New" {
		t.Errorf("Label = %q, want %q", item.Label, "New")
	}
	if item.Icon.Name != "add" {
		t.Errorf("Icon.Name = %q, want %q", item.Icon.Name, "add")
	}
	if !item.Enabled {
		t.Error("should be enabled by default")
	}
	if item.ShowLabel {
		t.Error("ShowLabel should be false for IconButton")
	}
	item.OnClick()
	if !called {
		t.Error("OnClick should have been called")
	}
}

func TestTextIconButton_Constructor(t *testing.T) {
	item := TextIconButton("Open", icon.Menu, func() {})

	if item.Kind != ItemButton {
		t.Errorf("Kind = %v, want ItemButton", item.Kind)
	}
	if !item.ShowLabel {
		t.Error("ShowLabel should be true for TextIconButton")
	}
	if item.Label != "Open" {
		t.Errorf("Label = %q, want %q", item.Label, "Open")
	}
}

func TestSeparator_Constructor(t *testing.T) {
	item := Separator()

	if item.Kind != ItemSeparator {
		t.Errorf("Kind = %v, want ItemSeparator", item.Kind)
	}
	if !item.Enabled {
		t.Error("should be enabled")
	}
}

func TestSpacer_Constructor(t *testing.T) {
	item := Spacer()

	if item.Kind != ItemSpacer {
		t.Errorf("Kind = %v, want ItemSpacer", item.Kind)
	}
}

func TestCustom_Constructor(t *testing.T) {
	w := &mockWidget{}
	item := Custom(w)

	if item.Kind != ItemCustom {
		t.Errorf("Kind = %v, want ItemCustom", item.Kind)
	}
	if item.Widget != w {
		t.Error("Widget should be the provided widget")
	}
}

// --- Construction Tests ---

func TestNew_Default(t *testing.T) {
	tb := New()

	if !tb.IsVisible() {
		t.Error("default toolbar should be visible")
	}
	if !tb.IsEnabled() {
		t.Error("default toolbar should be enabled")
	}
	if !tb.IsFocusable() {
		t.Error("default toolbar should be focusable")
	}
	if tb.ItemCount() != 0 {
		t.Errorf("ItemCount() = %d, want 0", tb.ItemCount())
	}
	if tb.cfg.height != defaultHeight {
		t.Errorf("height = %v, want %v", tb.cfg.height, defaultHeight)
	}
	if tb.focusIndex != noFocusIndex {
		t.Errorf("focusIndex = %d, want %d", tb.focusIndex, noFocusIndex)
	}
}

func TestNew_WithItems(t *testing.T) {
	tb := New(Items(
		IconButton("New", icon.Add, nil),
		Separator(),
		Spacer(),
		IconButton("Settings", icon.Settings, nil),
	))

	if tb.ItemCount() != 4 {
		t.Errorf("ItemCount() = %d, want 4", tb.ItemCount())
	}
	if tb.ItemAt(0).Kind != ItemButton {
		t.Errorf("ItemAt(0).Kind = %v, want ItemButton", tb.ItemAt(0).Kind)
	}
	if tb.ItemAt(1).Kind != ItemSeparator {
		t.Errorf("ItemAt(1).Kind = %v, want ItemSeparator", tb.ItemAt(1).Kind)
	}
	if tb.ItemAt(2).Kind != ItemSpacer {
		t.Errorf("ItemAt(2).Kind = %v, want ItemSpacer", tb.ItemAt(2).Kind)
	}
}

func TestNew_WithHeight(t *testing.T) {
	tb := New(Height(56))

	if tb.cfg.height != 56 {
		t.Errorf("height = %v, want 56", tb.cfg.height)
	}
}

func TestNew_WithPainter(t *testing.T) {
	p := &testPainter{}
	tb := New(PainterOpt(p))

	if tb.painter != p {
		t.Error("painter should be the custom painter")
	}
}

func TestItemAt_OutOfRange(t *testing.T) {
	tb := New(Items(IconButton("A", icon.Add, nil)))

	empty := tb.ItemAt(-1)
	if empty.Kind != ItemButton || empty.Label != "" {
		t.Error("ItemAt(-1) should return empty Item")
	}

	empty = tb.ItemAt(5)
	if empty.Label != "" {
		t.Error("ItemAt(5) should return empty Item")
	}
}

// --- Layout Tests ---

func TestLayout_Size(t *testing.T) {
	tb := New(
		Items(
			IconButton("A", icon.Add, nil),
			IconButton("B", icon.Close, nil),
		),
		Height(48),
	)
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 600))

	size := tb.Layout(ctx, constraints)

	if size.Height != 48 {
		t.Errorf("height = %v, want 48", size.Height)
	}
	if size.Width != 800 {
		t.Errorf("width = %v, want 800 (fills available)", size.Width)
	}
}

func TestLayout_ItemBounds(t *testing.T) {
	tb := New(Items(
		IconButton("A", icon.Add, nil),
		IconButton("B", icon.Close, nil),
	))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 100))

	tb.Layout(ctx, constraints)

	// First item should start at x=0.
	if tb.itemStates[0].bounds.Min.X != 0 {
		t.Errorf("item 0 x = %v, want 0", tb.itemStates[0].bounds.Min.X)
	}
	// Second item should start after first item + gap.
	expectedX := buttonItemSize + itemGap
	if tb.itemStates[1].bounds.Min.X != expectedX {
		t.Errorf("item 1 x = %v, want %v", tb.itemStates[1].bounds.Min.X, expectedX)
	}
}

func TestLayout_SpacerDistribution(t *testing.T) {
	tb := New(Items(
		IconButton("A", icon.Add, nil),
		Spacer(),
		IconButton("B", icon.Close, nil),
	))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 40))

	tb.Layout(ctx, constraints)

	// Last item should be near the right edge.
	lastBounds := tb.itemStates[2].bounds
	if lastBounds.Max.X < 350 {
		t.Errorf("last item Max.X = %v, expected near 400", lastBounds.Max.X)
	}
}

func TestLayout_EmptyToolbar(t *testing.T) {
	tb := New()
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 100))

	size := tb.Layout(ctx, constraints)

	if size.Height != defaultHeight {
		t.Errorf("height = %v, want %v", size.Height, defaultHeight)
	}
}

func TestLayout_TextIconButton_Width(t *testing.T) {
	tb := New(Items(
		TextIconButton("Open File", icon.Menu, nil),
	))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 40))

	tb.Layout(ctx, constraints)

	// Text+icon button should be wider than icon-only button.
	if tb.itemStates[0].bounds.Width() <= buttonItemSize {
		t.Errorf("text+icon button width = %v, should be > %v", tb.itemStates[0].bounds.Width(), buttonItemSize)
	}
}

// --- Draw Tests ---

func TestDraw_CallsPainter(t *testing.T) {
	p := &testPainter{}
	tb := New(
		Items(
			IconButton("A", icon.Add, nil),
			Separator(),
			IconButton("B", icon.Close, nil),
		),
		PainterOpt(p),
	)
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 40))
	tb.Layout(ctx, constraints)

	canvas := &mockCanvas{}
	tb.Draw(ctx, canvas)

	if !p.toolbarPainted {
		t.Error("PaintToolbar should have been called")
	}
	if p.buttonCount != 2 {
		t.Errorf("PaintButtonItem called %d times, want 2", p.buttonCount)
	}
	if p.separatorCount != 1 {
		t.Errorf("PaintSeparator called %d times, want 1", p.separatorCount)
	}
}

func TestDraw_Invisible(t *testing.T) {
	p := &testPainter{}
	tb := New(PainterOpt(p))
	tb.SetVisible(false)

	ctx := widget.NewContext()
	canvas := &mockCanvas{}
	tb.Draw(ctx, canvas)

	if p.toolbarPainted {
		t.Error("invisible toolbar should not paint")
	}
}

func TestDraw_DefaultPainter_Background(t *testing.T) {
	tb := New(Items(IconButton("A", icon.Add, nil)))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 40))
	tb.Layout(ctx, constraints)

	canvas := &mockCanvas{}
	tb.Draw(ctx, canvas)

	if len(canvas.drawRects) == 0 {
		t.Error("should draw toolbar background rect")
	}
}

func TestDraw_DefaultPainter_Separator(t *testing.T) {
	tb := New(Items(Separator()))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 40))
	tb.Layout(ctx, constraints)

	canvas := &mockCanvas{}
	tb.Draw(ctx, canvas)

	if len(canvas.drawLines) == 0 {
		t.Error("should draw separator line")
	}
}

// --- Mouse Event Tests ---

func TestMousePress_ButtonItem(t *testing.T) {
	clicked := false
	tb := New(Items(IconButton("A", icon.Add, func() { clicked = true })))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 40))
	tb.Layout(ctx, constraints)

	center := tb.itemStates[0].bounds.Center()
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		center, center, event.ModNone)
	consumed := tb.Event(ctx, press)

	if !consumed {
		t.Error("press on button should be consumed")
	}
	if tb.itemStates[0].interaction != statePressed {
		t.Errorf("state = %v, want statePressed", tb.itemStates[0].interaction)
	}
	if clicked {
		t.Error("should not click on press, only on release")
	}
}

func TestMouseRelease_FiresClick(t *testing.T) {
	clicked := false
	tb := New(Items(IconButton("A", icon.Add, func() { clicked = true })))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 40))
	tb.Layout(ctx, constraints)

	center := tb.itemStates[0].bounds.Center()
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		center, center, event.ModNone)
	tb.Event(ctx, press)

	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		center, center, event.ModNone)
	consumed := tb.Event(ctx, release)

	if !consumed {
		t.Error("release should be consumed")
	}
	if !clicked {
		t.Error("click handler should have been called")
	}
	if tb.itemStates[0].interaction != stateHover {
		t.Errorf("state = %v, want stateHover (released inside)", tb.itemStates[0].interaction)
	}
}

func TestMouseRelease_Outside_NoClick(t *testing.T) {
	clicked := false
	tb := New(Items(IconButton("A", icon.Add, func() { clicked = true })))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 40))
	tb.Layout(ctx, constraints)

	center := tb.itemStates[0].bounds.Center()
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		center, center, event.ModNone)
	tb.Event(ctx, press)

	outside := geometry.Pt(300, 20)
	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		outside, outside, event.ModNone)
	tb.Event(ctx, release)

	if clicked {
		t.Error("should not click when released outside item")
	}
}

func TestMouseMove_HoverState(t *testing.T) {
	tb := New(Items(IconButton("A", icon.Add, nil)))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 40))
	tb.Layout(ctx, constraints)

	center := tb.itemStates[0].bounds.Center()
	move := event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0,
		center, center, event.ModNone)
	tb.Event(ctx, move)

	if tb.itemStates[0].interaction != stateHover {
		t.Errorf("state = %v, want stateHover", tb.itemStates[0].interaction)
	}
}

func TestMouseLeave_ClearsHover(t *testing.T) {
	tb := New(Items(IconButton("A", icon.Add, nil)))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 40))
	tb.Layout(ctx, constraints)

	// Set hover first.
	center := tb.itemStates[0].bounds.Center()
	move := event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0,
		center, center, event.ModNone)
	tb.Event(ctx, move)

	leave := event.NewMouseEvent(event.MouseLeave, event.ButtonNone, 0,
		geometry.Pt(-1, -1), geometry.Pt(-1, -1), event.ModNone)
	tb.Event(ctx, leave)

	if tb.itemStates[0].interaction != stateNormal {
		t.Errorf("state = %v, want stateNormal after leave", tb.itemStates[0].interaction)
	}
}

func TestMousePress_RightButton_Ignored(t *testing.T) {
	tb := New(Items(IconButton("A", icon.Add, nil)))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 40))
	tb.Layout(ctx, constraints)

	center := tb.itemStates[0].bounds.Center()
	press := event.NewMouseEvent(event.MousePress, event.ButtonRight, event.ButtonStateRight,
		center, center, event.ModNone)
	consumed := tb.Event(ctx, press)

	if consumed {
		t.Error("right button press should not be consumed")
	}
}

func TestMousePress_DisabledItem_Ignored(t *testing.T) {
	item := IconButton("A", icon.Add, nil)
	item.Enabled = false
	tb := New(Items(item))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 40))
	tb.Layout(ctx, constraints)

	center := tb.itemStates[0].bounds.Center()
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		center, center, event.ModNone)
	consumed := tb.Event(ctx, press)

	if consumed {
		t.Error("disabled item should not consume press")
	}
}

func TestMousePress_Separator_Ignored(t *testing.T) {
	tb := New(Items(Separator()))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 40))
	tb.Layout(ctx, constraints)

	center := tb.itemStates[0].bounds.Center()
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		center, center, event.ModNone)
	consumed := tb.Event(ctx, press)

	if consumed {
		t.Error("separator should not consume press")
	}
}

func TestMousePress_NilOnClick(t *testing.T) {
	tb := New(Items(IconButton("A", icon.Add, nil)))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 40))
	tb.Layout(ctx, constraints)

	center := tb.itemStates[0].bounds.Center()
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		center, center, event.ModNone)
	tb.Event(ctx, press)

	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		center, center, event.ModNone)
	// Should not panic with nil OnClick.
	tb.Event(ctx, release)
}

// --- Keyboard Tests ---

func TestKeyboard_ArrowNavigation(t *testing.T) {
	tb := New(Items(
		IconButton("A", icon.Add, nil),
		Separator(),
		IconButton("B", icon.Close, nil),
	))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 40))
	tb.Layout(ctx, constraints)
	tb.SetFocused(true)

	// Right arrow should focus first button.
	right := event.NewKeyEvent(event.KeyPress, event.KeyRight, 0, event.ModNone)
	consumed := tb.Event(ctx, right)

	if !consumed {
		t.Error("right arrow should be consumed")
	}
	if tb.focusIndex != 0 {
		t.Errorf("focusIndex = %d, want 0", tb.focusIndex)
	}

	// Right arrow again should skip separator and focus second button.
	consumed = tb.Event(ctx, right)
	if !consumed {
		t.Error("second right arrow should be consumed")
	}
	if tb.focusIndex != 2 {
		t.Errorf("focusIndex = %d, want 2 (skipping separator)", tb.focusIndex)
	}

	// Left arrow should go back to first button.
	left := event.NewKeyEvent(event.KeyPress, event.KeyLeft, 0, event.ModNone)
	consumed = tb.Event(ctx, left)
	if !consumed {
		t.Error("left arrow should be consumed")
	}
	if tb.focusIndex != 0 {
		t.Errorf("focusIndex = %d, want 0", tb.focusIndex)
	}
}

func TestKeyboard_EnterActivation(t *testing.T) {
	clicked := false
	tb := New(Items(IconButton("A", icon.Add, func() { clicked = true })))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 40))
	tb.Layout(ctx, constraints)
	tb.SetFocused(true)
	tb.focusIndex = 0

	press := event.NewKeyEvent(event.KeyPress, event.KeyEnter, 0, event.ModNone)
	consumed := tb.Event(ctx, press)
	if !consumed {
		t.Error("Enter press should be consumed")
	}
	if tb.itemStates[0].interaction != statePressed {
		t.Errorf("state = %v, want statePressed", tb.itemStates[0].interaction)
	}

	release := event.NewKeyEvent(event.KeyRelease, event.KeyEnter, 0, event.ModNone)
	consumed = tb.Event(ctx, release)
	if !consumed {
		t.Error("Enter release should be consumed")
	}
	if !clicked {
		t.Error("Enter release should trigger click")
	}
}

func TestKeyboard_SpaceActivation(t *testing.T) {
	clicked := false
	tb := New(Items(IconButton("A", icon.Add, func() { clicked = true })))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 40))
	tb.Layout(ctx, constraints)
	tb.SetFocused(true)
	tb.focusIndex = 0

	press := event.NewKeyEvent(event.KeyPress, event.KeySpace, 0, event.ModNone)
	tb.Event(ctx, press)

	release := event.NewKeyEvent(event.KeyRelease, event.KeySpace, 0, event.ModNone)
	tb.Event(ctx, release)

	if !clicked {
		t.Error("Space release should trigger click")
	}
}

func TestKeyboard_NotFocused_Ignored(t *testing.T) {
	clicked := false
	tb := New(Items(IconButton("A", icon.Add, func() { clicked = true })))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 40))
	tb.Layout(ctx, constraints)
	tb.SetFocused(false)
	tb.focusIndex = 0

	press := event.NewKeyEvent(event.KeyPress, event.KeyEnter, 0, event.ModNone)
	consumed := tb.Event(ctx, press)

	if consumed {
		t.Error("Enter should not be consumed when not focused")
	}
	if clicked {
		t.Error("should not click when not focused")
	}
}

func TestKeyboard_DisabledItem_Ignored(t *testing.T) {
	clicked := false
	item := IconButton("A", icon.Add, func() { clicked = true })
	item.Enabled = false
	tb := New(Items(item))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 40))
	tb.Layout(ctx, constraints)
	tb.SetFocused(true)
	tb.focusIndex = 0

	press := event.NewKeyEvent(event.KeyPress, event.KeyEnter, 0, event.ModNone)
	consumed := tb.Event(ctx, press)

	if consumed {
		t.Error("disabled item should not consume key events")
	}
	if clicked {
		t.Error("disabled item should not fire click")
	}
}

func TestKeyboard_NoFocusIndex_Activation(t *testing.T) {
	tb := New(Items(IconButton("A", icon.Add, nil)))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 40))
	tb.Layout(ctx, constraints)
	tb.SetFocused(true)

	press := event.NewKeyEvent(event.KeyPress, event.KeyEnter, 0, event.ModNone)
	consumed := tb.Event(ctx, press)

	if consumed {
		t.Error("Enter with no focus index should not be consumed")
	}
}

func TestKeyboard_TabNavigation(t *testing.T) {
	tb := New(Items(
		IconButton("A", icon.Add, nil),
		IconButton("B", icon.Close, nil),
	))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 40))
	tb.Layout(ctx, constraints)
	tb.SetFocused(true)

	tab := event.NewKeyEvent(event.KeyPress, event.KeyTab, 0, event.ModNone)
	tb.Event(ctx, tab)

	if tb.focusIndex != 0 {
		t.Errorf("focusIndex = %d, want 0", tb.focusIndex)
	}

	tb.Event(ctx, tab)
	if tb.focusIndex != 1 {
		t.Errorf("focusIndex = %d, want 1", tb.focusIndex)
	}

	// Shift+Tab goes backward.
	shiftTab := event.NewKeyEvent(event.KeyPress, event.KeyTab, 0, event.ModShift)
	tb.Event(ctx, shiftTab)
	if tb.focusIndex != 0 {
		t.Errorf("focusIndex = %d, want 0 after Shift+Tab", tb.focusIndex)
	}
}

func TestKeyboard_ArrowPastEnd(t *testing.T) {
	tb := New(Items(IconButton("A", icon.Add, nil)))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 40))
	tb.Layout(ctx, constraints)
	tb.SetFocused(true)
	tb.focusIndex = 0

	right := event.NewKeyEvent(event.KeyPress, event.KeyRight, 0, event.ModNone)
	consumed := tb.Event(ctx, right)

	if consumed {
		t.Error("right arrow past end should not be consumed")
	}
	if tb.focusIndex != 0 {
		t.Errorf("focusIndex = %d, should remain 0", tb.focusIndex)
	}
}

func TestKeyboard_OtherKeys_Ignored(t *testing.T) {
	tb := New(Items(IconButton("A", icon.Add, nil)))
	ctx := widget.NewContext()
	tb.SetFocused(true)

	key := event.NewKeyEvent(event.KeyPress, event.KeyA, 'a', event.ModNone)
	consumed := tb.Event(ctx, key)

	if consumed {
		t.Error("other keys should not be consumed")
	}
}

// --- Disabled Toolbar Tests ---

func TestDisabledToolbar_IgnoresEvents(t *testing.T) {
	tb := New(Items(IconButton("A", icon.Add, nil)))
	tb.SetEnabled(false)
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 40))
	tb.Layout(ctx, constraints)

	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(18, 20), geometry.Pt(18, 20), event.ModNone)
	consumed := tb.Event(ctx, press)

	if consumed {
		t.Error("disabled toolbar should not consume events")
	}
}

// --- IsFocusable Tests ---

func TestIsFocusable(t *testing.T) {
	tests := []struct {
		name    string
		visible bool
		enabled bool
		want    bool
	}{
		{"visible+enabled", true, true, true},
		{"invisible", false, true, false},
		{"disabled", true, false, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tb := New()
			tb.SetVisible(tc.visible)
			tb.SetEnabled(tc.enabled)
			if got := tb.IsFocusable(); got != tc.want {
				t.Errorf("IsFocusable() = %v, want %v", got, tc.want)
			}
		})
	}
}

// --- Children Tests ---

func TestChildren_NoCustomWidgets(t *testing.T) {
	tb := New(Items(
		IconButton("A", icon.Add, nil),
		Separator(),
	))

	children := tb.Children()
	if children != nil {
		t.Errorf("Children() = %v, want nil (no custom widgets)", children)
	}
}

func TestChildren_WithCustomWidgets(t *testing.T) {
	w := &mockWidget{}
	tb := New(Items(
		IconButton("A", icon.Add, nil),
		Custom(w),
	))

	children := tb.Children()
	if len(children) != 1 {
		t.Fatalf("len(Children()) = %d, want 1", len(children))
	}
	if children[0] != w {
		t.Error("Children()[0] should be the custom widget")
	}
}

// --- Accessibility Tests ---

func TestAccessibility(t *testing.T) {
	tb := New()

	if role := tb.AccessibilityRole(); role != a11y.RoleToolbar {
		t.Errorf("AccessibilityRole() = %v, want RoleToolbar", role)
	}
	if label := tb.AccessibilityLabel(); label != a11yLabel {
		t.Errorf("AccessibilityLabel() = %q, want %q", label, a11yLabel)
	}
	if hint := tb.AccessibilityHint(); hint != "" {
		t.Errorf("AccessibilityHint() = %q, want empty", hint)
	}
	if value := tb.AccessibilityValue(); value != "" {
		t.Errorf("AccessibilityValue() = %q, want empty", value)
	}
}

func TestAccessibilityState_Disabled(t *testing.T) {
	tb := New()
	tb.SetEnabled(false)

	state := tb.AccessibilityState()
	if !state.Disabled {
		t.Error("disabled toolbar should report Disabled=true")
	}
}

func TestAccessibilityState_Hidden(t *testing.T) {
	tb := New()
	tb.SetVisible(false)

	state := tb.AccessibilityState()
	if !state.Hidden {
		t.Error("hidden toolbar should report Hidden=true")
	}
}

func TestAccessibilityActions_Nil(t *testing.T) {
	tb := New()
	if actions := tb.AccessibilityActions(); actions != nil {
		t.Errorf("AccessibilityActions() = %v, want nil", actions)
	}
}

// --- Painter Interface Tests ---

func TestDefaultPainter_PaintToolbar_EmptyBounds(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	p.PaintToolbar(canvas, PaintToolbarState{Bounds: geometry.Rect{}})

	if len(canvas.drawRects) > 0 {
		t.Error("should not paint with empty bounds")
	}
}

func TestDefaultPainter_PaintButtonItem_EmptyBounds(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	p.PaintButtonItem(canvas, PaintButtonState{Bounds: geometry.Rect{}})

	if len(canvas.drawRoundRects) > 0 {
		t.Error("should not paint with empty bounds")
	}
}

func TestDefaultPainter_PaintSeparator_EmptyBounds(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	p.PaintSeparator(canvas, geometry.Rect{})

	if len(canvas.drawLines) > 0 {
		t.Error("should not paint with empty bounds")
	}
}

func TestDefaultPainter_PaintButtonItem_HoverState(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	bounds := geometry.NewRect(0, 0, 36, 40)
	p.PaintButtonItem(canvas, PaintButtonState{
		Icon:    icon.Add,
		Hovered: true,
		Bounds:  bounds,
	})

	if len(canvas.drawRoundRects) == 0 {
		t.Error("hovered button should draw background round rect")
	}
}

func TestDefaultPainter_PaintButtonItem_PressedState(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	bounds := geometry.NewRect(0, 0, 36, 40)
	p.PaintButtonItem(canvas, PaintButtonState{
		Icon:    icon.Add,
		Pressed: true,
		Bounds:  bounds,
	})

	if len(canvas.drawRoundRects) == 0 {
		t.Error("pressed button should draw background round rect")
	}
}

func TestDefaultPainter_PaintButtonItem_Disabled(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	bounds := geometry.NewRect(0, 0, 36, 40)
	p.PaintButtonItem(canvas, PaintButtonState{
		Icon:     icon.Add,
		Disabled: true,
		Bounds:   bounds,
	})

	// Disabled items should not have hover/press background.
	if len(canvas.drawRoundRects) > 0 {
		t.Error("disabled button should not draw hover/press background")
	}
}

func TestDefaultPainter_PaintButtonItem_FocusRing(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	bounds := geometry.NewRect(10, 5, 46, 35)
	p.PaintButtonItem(canvas, PaintButtonState{
		Icon:    icon.Add,
		Focused: true,
		Bounds:  bounds,
	})

	found := false
	for _, call := range canvas.strokeRoundRects {
		if call.r.Min.X < 10 {
			found = true
			break
		}
	}
	if !found {
		t.Error("focused button should draw a focus ring (expanded stroke)")
	}
}

func TestDefaultPainter_PaintButtonItem_FocusRing_NotWhenDisabled(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	bounds := geometry.NewRect(10, 5, 46, 35)
	p.PaintButtonItem(canvas, PaintButtonState{
		Icon:     icon.Add,
		Focused:  true,
		Disabled: true,
		Bounds:   bounds,
	})

	for _, call := range canvas.strokeRoundRects {
		if call.r.Min.X < 10 {
			t.Error("disabled focused button should not draw focus ring")
			break
		}
	}
}

func TestDefaultPainter_PaintButtonItem_ShowLabel(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	bounds := geometry.NewRect(0, 0, 100, 40)
	p.PaintButtonItem(canvas, PaintButtonState{
		Label:     "Open",
		Icon:      icon.Menu,
		ShowLabel: true,
		Bounds:    bounds,
	})

	if len(canvas.drawTexts) == 0 {
		t.Error("should draw label text when ShowLabel is true")
	}
	if canvas.drawTexts[0].text != "Open" {
		t.Errorf("text = %q, want %q", canvas.drawTexts[0].text, "Open")
	}
}

func TestDefaultPainter_PaintButtonItem_NoLabel(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	bounds := geometry.NewRect(0, 0, 36, 40)
	p.PaintButtonItem(canvas, PaintButtonState{
		Label: "Open",
		Icon:  icon.Menu,
		// ShowLabel is false
		Bounds: bounds,
	})

	if len(canvas.drawTexts) > 0 {
		t.Error("should not draw label text when ShowLabel is false")
	}
}

// --- Helper Function Tests ---

func TestIconBoundsForItem(t *testing.T) {
	t.Run("centered icon", func(t *testing.T) {
		bounds := geometry.NewRect(0, 0, 36, 40)
		iconB := iconBoundsForItem(bounds, false)

		if iconB.IsEmpty() {
			t.Error("icon bounds should not be empty")
		}
		// Icon should be centered.
		centerX := bounds.Min.X + bounds.Width()/2
		iconCenterX := iconB.Min.X + iconB.Width()/2
		if abs32(centerX-iconCenterX) > 1 {
			t.Errorf("icon not centered: center=%v, iconCenter=%v", centerX, iconCenterX)
		}
	})

	t.Run("icon with label", func(t *testing.T) {
		bounds := geometry.NewRect(0, 0, 100, 40)
		iconB := iconBoundsForItem(bounds, true)

		// Icon should be on the left.
		if iconB.Min.X != iconPadding {
			t.Errorf("icon.Min.X = %v, want %v", iconB.Min.X, iconPadding)
		}
	})
}

func TestTextBoundsForItem(t *testing.T) {
	itemBounds := geometry.NewRect(0, 0, 100, 40)
	// NewRect(x, y, w, h): icon at (6, 10) with size 20x20 -> Max.X = 26
	iconRect := geometry.NewRect(6, 10, 20, 20)
	textB := textBoundsForItem(itemBounds, iconRect)

	expectedX := iconRect.Max.X + textIconGap
	if textB.Min.X != expectedX {
		t.Errorf("text.Min.X = %v, want %v", textB.Min.X, expectedX)
	}
	if textB.Min.Y != itemBounds.Min.Y {
		t.Errorf("text.Min.Y = %v, want %v", textB.Min.Y, itemBounds.Min.Y)
	}
}

func TestButtonItemWidth(t *testing.T) {
	tb := New()

	t.Run("icon only", func(t *testing.T) {
		item := IconButton("A", icon.Add, nil)
		if got := tb.buttonItemWidth(item); got != buttonItemSize {
			t.Errorf("width = %v, want %v", got, buttonItemSize)
		}
	})

	t.Run("with label", func(t *testing.T) {
		item := TextIconButton("Open File", icon.Menu, nil)
		width := tb.buttonItemWidth(item)
		if width <= buttonItemSize {
			t.Errorf("width = %v, should be > %v for text+icon", width, buttonItemSize)
		}
	})

	t.Run("short label enforces min width", func(t *testing.T) {
		item := TextIconButton("X", icon.Add, nil)
		width := tb.buttonItemWidth(item)
		if width < textButtonMinWidth {
			t.Errorf("width = %v, should be >= %v", width, textButtonMinWidth)
		}
	})
}

// --- HitTest Tests ---

func TestHitTest(t *testing.T) {
	tb := New(Items(
		IconButton("A", icon.Add, nil),
		Separator(),
		IconButton("B", icon.Close, nil),
	))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 40))
	tb.Layout(ctx, constraints)

	t.Run("hit first button", func(t *testing.T) {
		center := tb.itemStates[0].bounds.Center()
		if got := tb.hitTest(center); got != 0 {
			t.Errorf("hitTest = %d, want 0", got)
		}
	})

	t.Run("hit last button", func(t *testing.T) {
		center := tb.itemStates[2].bounds.Center()
		if got := tb.hitTest(center); got != 2 {
			t.Errorf("hitTest = %d, want 2", got)
		}
	})

	t.Run("miss", func(t *testing.T) {
		if got := tb.hitTest(geometry.Pt(399, 20)); got != noFocusIndex {
			t.Errorf("hitTest = %d, want %d", got, noFocusIndex)
		}
	})
}

// --- Custom Widget Tests ---

func TestLayout_CustomWidget(t *testing.T) {
	cw := &mockWidget{}
	tb := New(Items(
		IconButton("A", icon.Add, nil),
		Custom(cw),
	))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 40))
	tb.Layout(ctx, constraints)

	cwBounds := cw.Bounds()
	if cwBounds.IsEmpty() {
		t.Error("custom widget should have non-empty bounds after layout")
	}
}

func TestDraw_CustomWidget(t *testing.T) {
	drawn := false
	cw := &drawTrackingWidget{onDraw: func() { drawn = true }}
	tb := New(Items(Custom(cw)))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 40))
	tb.Layout(ctx, constraints)

	canvas := &mockCanvas{}
	tb.Draw(ctx, canvas)

	if !drawn {
		t.Error("custom widget Draw should be called")
	}
}

func TestDispatchToCustomItems(t *testing.T) {
	consumed := false
	cw := &eventTrackingWidget{onEvent: func() { consumed = true }}
	tb := New(Items(Custom(cw)))
	ctx := widget.NewContext()

	// Non-mouse/non-key events dispatch to custom items.
	fe := &event.FocusEvent{}
	tb.Event(ctx, fe)

	if !consumed {
		t.Error("event should be dispatched to custom widget")
	}
}

func TestMousePress_Empty_NoHit(t *testing.T) {
	tb := New(Items(IconButton("A", icon.Add, nil)))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 40))
	tb.Layout(ctx, constraints)

	// Click far away from any item.
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(300, 20), geometry.Pt(300, 20), event.ModNone)
	consumed := tb.Event(ctx, press)

	if consumed {
		t.Error("press outside all items should not be consumed")
	}
}

func TestMouseMove_NoChange(t *testing.T) {
	tb := New(Items(Separator()))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 40))
	tb.Layout(ctx, constraints)

	// Move over a separator (not a button) -- no state change.
	center := tb.itemStates[0].bounds.Center()
	move := event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0,
		center, center, event.ModNone)
	consumed := tb.Event(ctx, move)

	if consumed {
		t.Error("move over separator should return false (no button state change)")
	}
}

func TestKeyboard_ArrowKey_Release_Ignored(t *testing.T) {
	tb := New(Items(IconButton("A", icon.Add, nil)))
	ctx := widget.NewContext()
	tb.SetFocused(true)

	release := event.NewKeyEvent(event.KeyRelease, event.KeyRight, 0, event.ModNone)
	consumed := tb.Event(ctx, release)

	if consumed {
		t.Error("arrow key release should not be consumed")
	}
}

func TestKeyboard_TabKey_Release_Ignored(t *testing.T) {
	tb := New(Items(IconButton("A", icon.Add, nil)))
	ctx := widget.NewContext()
	tb.SetFocused(true)

	release := event.NewKeyEvent(event.KeyRelease, event.KeyTab, 0, event.ModNone)
	consumed := tb.Event(ctx, release)

	if consumed {
		t.Error("tab key release should not be consumed")
	}
}

func TestMoveFocus_EmptyToolbar(t *testing.T) {
	tb := New()
	ctx := widget.NewContext()
	tb.SetFocused(true)

	right := event.NewKeyEvent(event.KeyPress, event.KeyRight, 0, event.ModNone)
	consumed := tb.Event(ctx, right)

	if consumed {
		t.Error("arrow in empty toolbar should not be consumed")
	}
}

func TestMoveFocus_LeftFromStart(t *testing.T) {
	tb := New(Items(IconButton("A", icon.Add, nil)))
	ctx := widget.NewContext()
	tb.SetFocused(true)

	// Left from no-focus starts at len(items), goes backward, finds item 0.
	left := event.NewKeyEvent(event.KeyPress, event.KeyLeft, 0, event.ModNone)
	consumed := tb.Event(ctx, left)

	if !consumed {
		t.Error("left arrow from no-focus should find the last focusable item")
	}
	if tb.focusIndex != 0 {
		t.Errorf("focusIndex = %d, want 0", tb.focusIndex)
	}
}

func TestIconBoundsForItem_ZeroHeight(t *testing.T) {
	// Item bounds with very small height where iconPadding*2 > height.
	bounds := geometry.NewRect(0, 0, 36, 4)
	iconB := iconBoundsForItem(bounds, false)

	// h = 4 - 12 = -8 -> clamped to 0, so icon size = 0.
	if iconB.Width() != 0 || iconB.Height() != 0 {
		t.Errorf("icon with tiny bounds should have 0 size, got %v x %v", iconB.Width(), iconB.Height())
	}
}

// --- abs32 helper ---

func abs32(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

// --- Mock Types ---

type testPainter struct {
	toolbarPainted bool
	buttonCount    int
	separatorCount int
}

func (p *testPainter) PaintToolbar(_ widget.Canvas, _ PaintToolbarState) {
	p.toolbarPainted = true
}

func (p *testPainter) PaintButtonItem(_ widget.Canvas, _ PaintButtonState) {
	p.buttonCount++
}

func (p *testPainter) PaintSeparator(_ widget.Canvas, _ geometry.Rect) {
	p.separatorCount++
}

type mockWidget struct {
	widget.WidgetBase
}

func (w *mockWidget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	size := constraints.Constrain(geometry.Sz(60, 30))
	w.SetBounds(geometry.FromPointSize(w.Position(), size))
	return size
}

func (w *mockWidget) Draw(_ widget.Context, _ widget.Canvas) {}

func (w *mockWidget) Event(_ widget.Context, _ event.Event) bool { return false }

func (w *mockWidget) Children() []widget.Widget { return nil }

type drawTrackingWidget struct {
	widget.WidgetBase
	onDraw func()
}

func (w *drawTrackingWidget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	size := constraints.Constrain(geometry.Sz(60, 30))
	w.SetBounds(geometry.FromPointSize(w.Position(), size))
	return size
}

func (w *drawTrackingWidget) Draw(_ widget.Context, _ widget.Canvas) {
	if w.onDraw != nil {
		w.onDraw()
	}
}

func (w *drawTrackingWidget) Event(_ widget.Context, _ event.Event) bool { return false }

func (w *drawTrackingWidget) Children() []widget.Widget { return nil }

type eventTrackingWidget struct {
	widget.WidgetBase
	onEvent func()
}

func (w *eventTrackingWidget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	size := constraints.Constrain(geometry.Sz(60, 30))
	w.SetBounds(geometry.FromPointSize(w.Position(), size))
	return size
}

func (w *eventTrackingWidget) Draw(_ widget.Context, _ widget.Canvas) {}

func (w *eventTrackingWidget) Event(_ widget.Context, _ event.Event) bool {
	if w.onEvent != nil {
		w.onEvent()
	}
	return true
}

func (w *eventTrackingWidget) Children() []widget.Widget { return nil }

// --- mockCanvas ---

type mockCanvas struct {
	drawRects        []drawRectCall
	drawRoundRects   []drawRoundRectCall
	strokeRoundRects []strokeRoundRectCall
	drawTexts        []drawTextCall
	drawLines        []drawLineCall
}

type drawRectCall struct {
	r     geometry.Rect
	color widget.Color
}

type drawRoundRectCall struct {
	r      geometry.Rect
	color  widget.Color
	radius float32
}

type strokeRoundRectCall struct {
	r           geometry.Rect
	color       widget.Color
	radius      float32
	strokeWidth float32
}

type drawTextCall struct {
	text     string
	bounds   geometry.Rect
	fontSize float32
	color    widget.Color
	bold     bool
	align    widget.TextAlign
}

type drawLineCall struct {
	from, to    geometry.Point
	color       widget.Color
	strokeWidth float32
}

func (c *mockCanvas) Clear(_ widget.Color) {}

func (c *mockCanvas) DrawRect(r geometry.Rect, color widget.Color) {
	c.drawRects = append(c.drawRects, drawRectCall{r: r, color: color})
}

func (c *mockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) {}

func (c *mockCanvas) DrawRoundRect(r geometry.Rect, color widget.Color, radius float32) {
	c.drawRoundRects = append(c.drawRoundRects, drawRoundRectCall{r: r, color: color, radius: radius})
}

func (c *mockCanvas) StrokeRoundRect(r geometry.Rect, color widget.Color, radius float32, strokeWidth float32) {
	c.strokeRoundRects = append(c.strokeRoundRects, strokeRoundRectCall{r: r, color: color, radius: radius, strokeWidth: strokeWidth})
}

func (c *mockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)              {}
func (c *mockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {}

func (c *mockCanvas) DrawLine(from, to geometry.Point, color widget.Color, strokeWidth float32) {
	c.drawLines = append(c.drawLines, drawLineCall{from: from, to: to, color: color, strokeWidth: strokeWidth})
}

func (c *mockCanvas) DrawText(text string, bounds geometry.Rect, fontSize float32, color widget.Color, bold bool, align widget.TextAlign) {
	c.drawTexts = append(c.drawTexts, drawTextCall{text: text, bounds: bounds, fontSize: fontSize, color: color, bold: bold, align: align})
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
