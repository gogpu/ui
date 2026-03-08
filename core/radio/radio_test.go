package radio_test

import (
	"testing"

	"github.com/gogpu/ui/core/radio"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// --- Construction Tests ---

func TestNewGroup_Defaults(t *testing.T) {
	rg := radio.NewGroup()

	if !rg.IsVisible() {
		t.Error("default group should be visible")
	}
	if !rg.IsEnabled() {
		t.Error("default group should be enabled")
	}
	if rg.Selected() != "" {
		t.Errorf("default group should have no selection, got %q", rg.Selected())
	}
	if rg.ItemCount() != 0 {
		t.Errorf("default group should have 0 items, got %d", rg.ItemCount())
	}
	if rg.Children() == nil {
		t.Error("Children() should return empty slice, not nil")
	}
	if len(rg.Children()) != 0 {
		t.Errorf("Children() should return 0 items, got %d", len(rg.Children()))
	}
}

func TestNewGroup_WithOptions(t *testing.T) {
	changed := false
	rg := radio.NewGroup(
		radio.Items(
			radio.ItemDef{Value: "a", Label: "Alpha"},
			radio.ItemDef{Value: "b", Label: "Beta"},
			radio.ItemDef{Value: "c", Label: "Gamma"},
		),
		radio.Selected("b"),
		radio.OnChange(func(v string) { changed = true }),
		radio.GroupDisabled(false),
		radio.GroupA11yLabel("Choose option"),
	)

	if rg.ItemCount() != 3 {
		t.Fatalf("expected 3 items, got %d", rg.ItemCount())
	}
	if rg.Selected() != "b" {
		t.Errorf("expected selected 'b', got %q", rg.Selected())
	}
	if len(rg.Children()) != 3 {
		t.Errorf("Children() should return 3 items, got %d", len(rg.Children()))
	}
	_ = changed
}

func TestNewGroup_WithPainter(t *testing.T) {
	p := &testPainter{}
	rg := radio.NewGroup(
		radio.Items(radio.ItemDef{Value: "x", Label: "Test"}),
		radio.GroupPainter(p),
	)

	rg.SetBounds(geometry.NewRect(0, 0, 200, 100))
	rg.ItemAt(0).SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	rg.Draw(ctx, canvas)

	if !p.called {
		t.Error("Draw should delegate to the configured painter")
	}
}

func TestNewGroup_SelectedNotFound(t *testing.T) {
	rg := radio.NewGroup(
		radio.Items(radio.ItemDef{Value: "a", Label: "Alpha"}),
		radio.Selected("nonexistent"),
	)

	if rg.Selected() != "" {
		t.Errorf("should have no selection when value not found, got %q", rg.Selected())
	}
}

// --- Selection Tests ---

func TestSelect_Click(t *testing.T) {
	var selectedValue string
	rg := radio.NewGroup(
		radio.Items(
			radio.ItemDef{Value: "a", Label: "Alpha"},
			radio.ItemDef{Value: "b", Label: "Beta"},
		),
		radio.OnChange(func(v string) { selectedValue = v }),
	)
	rg.SetBounds(geometry.NewRect(0, 0, 200, 100))
	rg.ItemAt(0).SetBounds(geometry.NewRect(0, 0, 200, 30))
	rg.ItemAt(1).SetBounds(geometry.NewRect(0, 30, 200, 30))
	ctx := widget.NewContext()

	// Click on item B.
	simulateClick(rg.ItemAt(1), ctx, 45)

	if selectedValue != "b" {
		t.Errorf("expected selected 'b', got %q", selectedValue)
	}
	if rg.Selected() != "b" {
		t.Errorf("group.Selected() should be 'b', got %q", rg.Selected())
	}
}

func TestSelect_Keyboard_Space(t *testing.T) {
	var selectedValue string
	rg := radio.NewGroup(
		radio.Items(
			radio.ItemDef{Value: "a", Label: "Alpha"},
			radio.ItemDef{Value: "b", Label: "Beta"},
		),
		radio.OnChange(func(v string) { selectedValue = v }),
	)
	rg.SetBounds(geometry.NewRect(0, 0, 200, 100))
	rg.ItemAt(0).SetBounds(geometry.NewRect(0, 0, 200, 30))
	rg.ItemAt(1).SetBounds(geometry.NewRect(0, 30, 200, 30))

	// Focus item A and press Space.
	rg.ItemAt(0).SetFocused(true)
	ctx := widget.NewContext()

	press := event.NewKeyEvent(event.KeyPress, event.KeySpace, 0, event.ModNone)
	rg.ItemAt(0).Event(ctx, press)

	release := event.NewKeyEvent(event.KeyRelease, event.KeySpace, 0, event.ModNone)
	rg.ItemAt(0).Event(ctx, release)

	if selectedValue != "a" {
		t.Errorf("expected selected 'a' via Space, got %q", selectedValue)
	}
}

func TestSelect_Keyboard_Enter(t *testing.T) {
	var selectedValue string
	rg := radio.NewGroup(
		radio.Items(
			radio.ItemDef{Value: "a", Label: "Alpha"},
			radio.ItemDef{Value: "b", Label: "Beta"},
		),
		radio.OnChange(func(v string) { selectedValue = v }),
	)
	rg.SetBounds(geometry.NewRect(0, 0, 200, 100))
	rg.ItemAt(0).SetBounds(geometry.NewRect(0, 0, 200, 30))
	rg.ItemAt(1).SetBounds(geometry.NewRect(0, 30, 200, 30))

	rg.ItemAt(1).SetFocused(true)
	ctx := widget.NewContext()

	press := event.NewKeyEvent(event.KeyPress, event.KeyEnter, 0, event.ModNone)
	rg.ItemAt(1).Event(ctx, press)

	release := event.NewKeyEvent(event.KeyRelease, event.KeyEnter, 0, event.ModNone)
	rg.ItemAt(1).Event(ctx, release)

	if selectedValue != "b" {
		t.Errorf("expected selected 'b' via Enter, got %q", selectedValue)
	}
}

func TestMutualExclusion(t *testing.T) {
	rg := radio.NewGroup(
		radio.Items(
			radio.ItemDef{Value: "a", Label: "Alpha"},
			radio.ItemDef{Value: "b", Label: "Beta"},
			radio.ItemDef{Value: "c", Label: "Gamma"},
		),
		radio.Selected("a"),
	)
	rg.SetBounds(geometry.NewRect(0, 0, 200, 100))
	rg.ItemAt(0).SetBounds(geometry.NewRect(0, 0, 200, 30))
	rg.ItemAt(1).SetBounds(geometry.NewRect(0, 30, 200, 30))
	rg.ItemAt(2).SetBounds(geometry.NewRect(0, 60, 200, 30))
	ctx := widget.NewContext()

	if rg.Selected() != "a" {
		t.Fatalf("initial selection should be 'a', got %q", rg.Selected())
	}

	// Select B.
	simulateClick(rg.ItemAt(1), ctx, 45)

	if rg.Selected() != "b" {
		t.Errorf("after clicking B, selected should be 'b', got %q", rg.Selected())
	}

	// Select C.
	simulateClick(rg.ItemAt(2), ctx, 75)

	if rg.Selected() != "c" {
		t.Errorf("after clicking C, selected should be 'c', got %q", rg.Selected())
	}
}

func TestSelected_Programmatic(t *testing.T) {
	rg := radio.NewGroup(
		radio.Items(
			radio.ItemDef{Value: "a", Label: "Alpha"},
			radio.ItemDef{Value: "b", Label: "Beta"},
		),
	)

	rg.Select("b")
	if rg.Selected() != "b" {
		t.Errorf("expected 'b', got %q", rg.Selected())
	}

	rg.Select("a")
	if rg.Selected() != "a" {
		t.Errorf("expected 'a', got %q", rg.Selected())
	}

	rg.Select("nonexistent")
	if rg.Selected() != "" {
		t.Errorf("expected empty for nonexistent, got %q", rg.Selected())
	}
}

// --- Arrow Key Navigation Tests ---

func TestArrowNavigation_Vertical(t *testing.T) {
	rg := radio.NewGroup(
		radio.Items(
			radio.ItemDef{Value: "a", Label: "Alpha"},
			radio.ItemDef{Value: "b", Label: "Beta"},
			radio.ItemDef{Value: "c", Label: "Gamma"},
		),
		radio.DirectionOpt(radio.Vertical),
	)
	rg.SetBounds(geometry.NewRect(0, 0, 200, 120))
	for i := 0; i < rg.ItemCount(); i++ {
		rg.ItemAt(i).SetBounds(geometry.NewRect(0, float32(i*30), 200, 30))
	}
	ctx := widget.NewContext()

	// Focus first item.
	ctx.RequestFocus(rg.ItemAt(0))

	// Press Down arrow -> should move focus to item 1.
	press := event.NewKeyEvent(event.KeyPress, event.KeyDown, 0, event.ModNone)
	consumed := rg.ItemAt(0).Event(ctx, press)

	if !consumed {
		t.Error("Down arrow should be consumed")
	}
	if !rg.ItemAt(1).IsFocused() {
		t.Error("focus should move to item 1 after Down arrow")
	}
}

func TestArrowNavigation_Vertical_WrapAround(t *testing.T) {
	rg := radio.NewGroup(
		radio.Items(
			radio.ItemDef{Value: "a", Label: "Alpha"},
			radio.ItemDef{Value: "b", Label: "Beta"},
		),
		radio.DirectionOpt(radio.Vertical),
	)
	rg.SetBounds(geometry.NewRect(0, 0, 200, 80))
	rg.ItemAt(0).SetBounds(geometry.NewRect(0, 0, 200, 30))
	rg.ItemAt(1).SetBounds(geometry.NewRect(0, 30, 200, 30))
	ctx := widget.NewContext()

	// Focus last item.
	ctx.RequestFocus(rg.ItemAt(1))

	// Press Down arrow -> should wrap to item 0.
	press := event.NewKeyEvent(event.KeyPress, event.KeyDown, 0, event.ModNone)
	rg.ItemAt(1).Event(ctx, press)

	if !rg.ItemAt(0).IsFocused() {
		t.Error("focus should wrap to item 0 after Down on last item")
	}
}

func TestArrowNavigation_Vertical_UpOnFirst(t *testing.T) {
	rg := radio.NewGroup(
		radio.Items(
			radio.ItemDef{Value: "a", Label: "Alpha"},
			radio.ItemDef{Value: "b", Label: "Beta"},
		),
		radio.DirectionOpt(radio.Vertical),
	)
	rg.SetBounds(geometry.NewRect(0, 0, 200, 80))
	rg.ItemAt(0).SetBounds(geometry.NewRect(0, 0, 200, 30))
	rg.ItemAt(1).SetBounds(geometry.NewRect(0, 30, 200, 30))
	ctx := widget.NewContext()

	// Focus first item.
	ctx.RequestFocus(rg.ItemAt(0))

	// Press Up arrow -> should wrap to last item.
	press := event.NewKeyEvent(event.KeyPress, event.KeyUp, 0, event.ModNone)
	rg.ItemAt(0).Event(ctx, press)

	if !rg.ItemAt(1).IsFocused() {
		t.Error("focus should wrap to last item after Up on first item")
	}
}

func TestArrowNavigation_Horizontal(t *testing.T) {
	rg := radio.NewGroup(
		radio.Items(
			radio.ItemDef{Value: "a", Label: "Alpha"},
			radio.ItemDef{Value: "b", Label: "Beta"},
			radio.ItemDef{Value: "c", Label: "Gamma"},
		),
		radio.DirectionOpt(radio.Horizontal),
	)
	rg.SetBounds(geometry.NewRect(0, 0, 400, 30))
	for i := 0; i < rg.ItemCount(); i++ {
		rg.ItemAt(i).SetBounds(geometry.NewRect(float32(i*100), 0, 100, 30))
	}
	ctx := widget.NewContext()

	// Focus first item.
	ctx.RequestFocus(rg.ItemAt(0))

	// Press Right arrow -> should move focus to item 1.
	press := event.NewKeyEvent(event.KeyPress, event.KeyRight, 0, event.ModNone)
	consumed := rg.ItemAt(0).Event(ctx, press)

	if !consumed {
		t.Error("Right arrow should be consumed in horizontal mode")
	}
	if !rg.ItemAt(1).IsFocused() {
		t.Error("focus should move to item 1 after Right arrow")
	}
}

func TestArrowNavigation_Horizontal_LeftRight(t *testing.T) {
	rg := radio.NewGroup(
		radio.Items(
			radio.ItemDef{Value: "a", Label: "Alpha"},
			radio.ItemDef{Value: "b", Label: "Beta"},
		),
		radio.DirectionOpt(radio.Horizontal),
	)
	rg.SetBounds(geometry.NewRect(0, 0, 400, 30))
	rg.ItemAt(0).SetBounds(geometry.NewRect(0, 0, 100, 30))
	rg.ItemAt(1).SetBounds(geometry.NewRect(100, 0, 100, 30))
	ctx := widget.NewContext()

	// Focus item 1, press Left -> item 0.
	ctx.RequestFocus(rg.ItemAt(1))
	press := event.NewKeyEvent(event.KeyPress, event.KeyLeft, 0, event.ModNone)
	rg.ItemAt(1).Event(ctx, press)

	if !rg.ItemAt(0).IsFocused() {
		t.Error("focus should move to item 0 after Left arrow")
	}
}

func TestArrowNavigation_WrongAxisIgnored(t *testing.T) {
	rg := radio.NewGroup(
		radio.Items(
			radio.ItemDef{Value: "a", Label: "Alpha"},
			radio.ItemDef{Value: "b", Label: "Beta"},
		),
		radio.DirectionOpt(radio.Vertical),
	)
	rg.SetBounds(geometry.NewRect(0, 0, 200, 80))
	rg.ItemAt(0).SetBounds(geometry.NewRect(0, 0, 200, 30))
	rg.ItemAt(1).SetBounds(geometry.NewRect(0, 30, 200, 30))
	ctx := widget.NewContext()

	// Focus item 0, press Left (wrong axis for vertical).
	ctx.RequestFocus(rg.ItemAt(0))
	press := event.NewKeyEvent(event.KeyPress, event.KeyLeft, 0, event.ModNone)
	consumed := rg.ItemAt(0).Event(ctx, press)

	if consumed {
		t.Error("Left arrow should not be consumed in vertical mode")
	}
	if !rg.ItemAt(0).IsFocused() {
		t.Error("focus should remain on item 0")
	}
}

// --- Disabled Tests ---

func TestDisabled_BlocksInteraction(t *testing.T) {
	var selectedValue string
	rg := radio.NewGroup(
		radio.Items(
			radio.ItemDef{Value: "a", Label: "Alpha"},
			radio.ItemDef{Value: "b", Label: "Beta"},
		),
		radio.OnChange(func(v string) { selectedValue = v }),
		radio.GroupDisabled(true),
	)
	rg.SetBounds(geometry.NewRect(0, 0, 200, 100))
	rg.ItemAt(0).SetBounds(geometry.NewRect(0, 0, 200, 30))
	rg.ItemAt(1).SetBounds(geometry.NewRect(0, 30, 200, 30))
	ctx := widget.NewContext()

	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(10, 10), geometry.Pt(10, 10), event.ModNone)
	consumed := rg.ItemAt(0).Event(ctx, press)

	if consumed {
		t.Error("disabled group should not consume events")
	}
	if selectedValue != "" {
		t.Error("disabled group should not fire onChange")
	}
}

func TestDisabled_DisabledFn(t *testing.T) {
	isDisabled := true
	rg := radio.NewGroup(
		radio.Items(radio.ItemDef{Value: "a", Label: "Alpha"}),
		radio.GroupDisabledFn(func() bool { return isDisabled }),
	)
	rg.SetBounds(geometry.NewRect(0, 0, 200, 100))
	rg.ItemAt(0).SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(10, 10), geometry.Pt(10, 10), event.ModNone)
	consumed := rg.ItemAt(0).Event(ctx, press)
	if consumed {
		t.Error("disabled group should not consume events")
	}

	isDisabled = false
	consumed = rg.ItemAt(0).Event(ctx, press)
	if !consumed {
		t.Error("enabled group should consume events")
	}
}

// --- OnChange Callback Tests ---

func TestOnChange_Callback(t *testing.T) {
	callCount := 0
	var lastValue string
	rg := radio.NewGroup(
		radio.Items(
			radio.ItemDef{Value: "a", Label: "Alpha"},
			radio.ItemDef{Value: "b", Label: "Beta"},
		),
		radio.OnChange(func(v string) {
			callCount++
			lastValue = v
		}),
	)
	rg.SetBounds(geometry.NewRect(0, 0, 200, 100))
	rg.ItemAt(0).SetBounds(geometry.NewRect(0, 0, 200, 30))
	rg.ItemAt(1).SetBounds(geometry.NewRect(0, 30, 200, 30))
	ctx := widget.NewContext()

	simulateClick(rg.ItemAt(0), ctx, 10)
	if callCount != 1 {
		t.Fatalf("onChange called %d times, want 1", callCount)
	}
	if lastValue != "a" {
		t.Errorf("lastValue = %q, want 'a'", lastValue)
	}

	simulateClick(rg.ItemAt(1), ctx, 45)
	if callCount != 2 {
		t.Fatalf("onChange called %d times, want 2", callCount)
	}
	if lastValue != "b" {
		t.Errorf("lastValue = %q, want 'b'", lastValue)
	}
}

func TestOnChange_NilCallback(t *testing.T) {
	rg := radio.NewGroup(
		radio.Items(radio.ItemDef{Value: "a", Label: "Alpha"}),
	)
	rg.SetBounds(geometry.NewRect(0, 0, 200, 100))
	rg.ItemAt(0).SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	// Should not panic.
	simulateClick(rg.ItemAt(0), ctx, 10)
}

// --- Focusable Tests ---

func TestFocusable_Items(t *testing.T) {
	rg := radio.NewGroup(
		radio.Items(
			radio.ItemDef{Value: "a", Label: "Alpha"},
			radio.ItemDef{Value: "b", Label: "Beta"},
		),
	)

	for i := 0; i < rg.ItemCount(); i++ {
		it := rg.ItemAt(i)
		if !it.IsFocusable() {
			t.Errorf("item %d should be focusable", i)
		}
	}
}

func TestFocusable_ItemsDisabledGroup(t *testing.T) {
	rg := radio.NewGroup(
		radio.Items(radio.ItemDef{Value: "a", Label: "Alpha"}),
		radio.GroupDisabled(true),
	)

	if rg.ItemAt(0).IsFocusable() {
		t.Error("items in disabled group should not be focusable")
	}
}

func TestFocusable_GroupNotFocusable(t *testing.T) {
	rg := radio.NewGroup(
		radio.Items(radio.ItemDef{Value: "a", Label: "Alpha"}),
	)

	if rg.IsFocusable() {
		t.Error("group itself should not be focusable (delegates to items)")
	}
}

func TestFocus_SetFocused(t *testing.T) {
	rg := radio.NewGroup(
		radio.Items(radio.ItemDef{Value: "a", Label: "Alpha"}),
	)

	it := rg.ItemAt(0)
	it.SetFocused(true)
	if !it.IsFocused() {
		t.Error("should be focused after SetFocused(true)")
	}

	it.SetFocused(false)
	if it.IsFocused() {
		t.Error("should not be focused after SetFocused(false)")
	}
}

// --- Layout Tests ---

func TestLayout_Vertical(t *testing.T) {
	rg := radio.NewGroup(
		radio.Items(
			radio.ItemDef{Value: "a", Label: "Alpha"},
			radio.ItemDef{Value: "b", Label: "Beta"},
		),
		radio.DirectionOpt(radio.Vertical),
	)
	rg.SetBounds(geometry.NewRect(0, 0, 200, 200))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 400))

	size := rg.Layout(ctx, constraints)

	if size.Width <= 0 {
		t.Error("width should be positive")
	}
	if size.Height <= 0 {
		t.Error("height should be positive")
	}
}

func TestLayout_Horizontal(t *testing.T) {
	rg := radio.NewGroup(
		radio.Items(
			radio.ItemDef{Value: "a", Label: "Alpha"},
			radio.ItemDef{Value: "b", Label: "Beta"},
		),
		radio.DirectionOpt(radio.Horizontal),
	)
	rg.SetBounds(geometry.NewRect(0, 0, 400, 50))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 400))

	size := rg.Layout(ctx, constraints)

	if size.Width <= 0 {
		t.Error("width should be positive")
	}
	if size.Height <= 0 {
		t.Error("height should be positive")
	}
}

func TestLayout_Empty(t *testing.T) {
	rg := radio.NewGroup()
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 400))

	size := rg.Layout(ctx, constraints)

	if size.Width != 0 || size.Height != 0 {
		t.Errorf("empty group should have zero size, got %v", size)
	}
}

func TestLayout_TightConstraints(t *testing.T) {
	rg := radio.NewGroup(
		radio.Items(radio.ItemDef{Value: "a", Label: "Alpha"}),
	)
	rg.SetBounds(geometry.NewRect(0, 0, 200, 200))
	ctx := widget.NewContext()
	constraints := geometry.Tight(geometry.Sz(60, 30))

	size := rg.Layout(ctx, constraints)

	if size.Width != 60 {
		t.Errorf("width = %v, want 60", size.Width)
	}
	if size.Height != 30 {
		t.Errorf("height = %v, want 30", size.Height)
	}
}

// --- Draw Tests ---

func TestDraw_DelegatesToPainter(t *testing.T) {
	p := &testPainter{}
	rg := radio.NewGroup(
		radio.Items(
			radio.ItemDef{Value: "a", Label: "Alpha"},
		),
		radio.GroupPainter(p),
		radio.Selected("a"),
	)
	rg.SetBounds(geometry.NewRect(0, 0, 200, 40))
	rg.ItemAt(0).SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	rg.Draw(ctx, canvas)

	if !p.called {
		t.Error("Draw should delegate to the configured painter")
	}
	if p.state.Label != "Alpha" {
		t.Errorf("PaintState.Label = %q, want %q", p.state.Label, "Alpha")
	}
	if !p.state.Selected {
		t.Error("PaintState.Selected should be true for selected item")
	}
	if p.state.Bounds.IsEmpty() {
		t.Error("PaintState.Bounds should not be empty")
	}
}

func TestDraw_DoesNotPanicWithBounds(t *testing.T) {
	rg := radio.NewGroup(
		radio.Items(radio.ItemDef{Value: "a", Label: "Alpha"}),
	)
	rg.SetBounds(geometry.NewRect(0, 0, 200, 40))
	rg.ItemAt(0).SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	rg.Draw(ctx, canvas)
}

func TestDraw_EmptyBounds(t *testing.T) {
	rg := radio.NewGroup(
		radio.Items(radio.ItemDef{Value: "a", Label: "Alpha"}),
	)
	// Do not set bounds (empty).
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	rg.Draw(ctx, canvas)

	// With empty item bounds, painter should skip drawing.
	if len(canvas.drawCircles) > 0 || len(canvas.strokeCircles) > 0 || len(canvas.drawTexts) > 0 {
		t.Error("should not draw anything with empty item bounds")
	}
}

// --- PaintState ColorScheme Tests ---

func TestPaintState_ColorScheme(t *testing.T) {
	scheme := radio.RadioColorScheme{
		SelectedBg:       widget.ColorRed,
		SelectedFg:       widget.ColorWhite,
		UnselectedBorder: widget.ColorGray,
		LabelColor:       widget.ColorBlack,
		DisabledBg:       widget.ColorLightGray,
		DisabledFg:       widget.ColorDarkGray,
		FocusRing:        widget.ColorBlue,
	}

	ps := radio.PaintState{
		Label:       "Test",
		Selected:    true,
		ColorScheme: scheme,
		Bounds:      geometry.NewRect(0, 0, 100, 40),
	}

	if ps.ColorScheme.SelectedBg != widget.ColorRed {
		t.Error("ColorScheme.SelectedBg should be red")
	}
	if ps.ColorScheme.SelectedFg != widget.ColorWhite {
		t.Error("ColorScheme.SelectedFg should be white")
	}
}

// --- Item Tests ---

func TestItem_Value(t *testing.T) {
	rg := radio.NewGroup(
		radio.Items(radio.ItemDef{Value: "test-val", Label: "Test Label"}),
	)

	it := rg.ItemAt(0)
	if it.Value() != "test-val" {
		t.Errorf("Value() = %q, want %q", it.Value(), "test-val")
	}
	if it.Label() != "Test Label" {
		t.Errorf("Label() = %q, want %q", it.Label(), "Test Label")
	}
}

func TestItem_Children(t *testing.T) {
	rg := radio.NewGroup(
		radio.Items(radio.ItemDef{Value: "a", Label: "Alpha"}),
	)

	if rg.ItemAt(0).Children() != nil {
		t.Error("item Children() should return nil")
	}
}

// --- Widget Interface Compliance ---

func TestWidgetInterface_Group(t *testing.T) {
	var w widget.Widget = radio.NewGroup()
	_ = w
}

func TestWidgetInterface_Item(t *testing.T) {
	rg := radio.NewGroup(
		radio.Items(radio.ItemDef{Value: "a", Label: "Alpha"}),
	)
	var w widget.Widget = rg.ItemAt(0)
	_ = w
}

func TestFocusableInterface_Item(t *testing.T) {
	rg := radio.NewGroup(
		radio.Items(radio.ItemDef{Value: "a", Label: "Alpha"}),
	)
	var f widget.Focusable = rg.ItemAt(0)
	_ = f
}

// --- Event Handling Tests ---

func TestEvent_MouseClickCycle(t *testing.T) {
	var selectedValue string
	rg := radio.NewGroup(
		radio.Items(radio.ItemDef{Value: "a", Label: "Alpha"}),
		radio.OnChange(func(v string) { selectedValue = v }),
	)
	rg.SetBounds(geometry.NewRect(0, 0, 200, 40))
	rg.ItemAt(0).SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	enter := event.NewMouseEvent(event.MouseEnter, event.ButtonNone, 0,
		geometry.Pt(10, 10), geometry.Pt(10, 10), event.ModNone)
	rg.ItemAt(0).Event(ctx, enter)

	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(10, 10), geometry.Pt(10, 10), event.ModNone)
	rg.ItemAt(0).Event(ctx, press)

	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		geometry.Pt(10, 10), geometry.Pt(10, 10), event.ModNone)
	rg.ItemAt(0).Event(ctx, release)

	if selectedValue != "a" {
		t.Error("should have selected after full mouse cycle")
	}
}

func TestEvent_DisabledIgnoresAll(t *testing.T) {
	var selectedValue string
	rg := radio.NewGroup(
		radio.Items(radio.ItemDef{Value: "a", Label: "Alpha"}),
		radio.OnChange(func(v string) { selectedValue = v }),
		radio.GroupDisabled(true),
	)
	rg.SetBounds(geometry.NewRect(0, 0, 200, 40))
	rg.ItemAt(0).SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(10, 10), geometry.Pt(10, 10), event.ModNone)
	consumed := rg.ItemAt(0).Event(ctx, press)

	if consumed {
		t.Error("disabled item should not consume events")
	}
	if selectedValue != "" {
		t.Error("disabled item should not fire onChange")
	}
}

func TestEvent_RightButtonIgnored(t *testing.T) {
	rg := radio.NewGroup(
		radio.Items(radio.ItemDef{Value: "a", Label: "Alpha"}),
	)
	rg.SetBounds(geometry.NewRect(0, 0, 200, 40))
	rg.ItemAt(0).SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	press := event.NewMouseEvent(event.MousePress, event.ButtonRight, event.ButtonStateRight,
		geometry.Pt(10, 10), geometry.Pt(10, 10), event.ModNone)
	consumed := rg.ItemAt(0).Event(ctx, press)

	if consumed {
		t.Error("right button press should not be consumed")
	}
}

// --- Group.Event delegation Tests ---

func TestGroupEvent_DelegatesToItems(t *testing.T) {
	var selectedValue string
	rg := radio.NewGroup(
		radio.Items(
			radio.ItemDef{Value: "a", Label: "Alpha"},
			radio.ItemDef{Value: "b", Label: "Beta"},
		),
		radio.OnChange(func(v string) { selectedValue = v }),
	)
	rg.SetBounds(geometry.NewRect(0, 0, 200, 100))
	rg.ItemAt(0).SetBounds(geometry.NewRect(0, 0, 200, 30))
	rg.ItemAt(1).SetBounds(geometry.NewRect(0, 30, 200, 30))
	ctx := widget.NewContext()

	// Send press through the group, not directly to the item.
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(10, 10), geometry.Pt(10, 10), event.ModNone)
	consumed := rg.Event(ctx, press)

	if !consumed {
		t.Error("Group.Event should delegate to items and consume the press")
	}

	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		geometry.Pt(10, 10), geometry.Pt(10, 10), event.ModNone)
	rg.Event(ctx, release)

	if selectedValue != "a" {
		t.Errorf("expected 'a' selected via Group.Event, got %q", selectedValue)
	}
}

func TestGroupEvent_UnhandledReturns_False(t *testing.T) {
	rg := radio.NewGroup(
		radio.Items(radio.ItemDef{Value: "a", Label: "Alpha"}),
		radio.GroupDisabled(true),
	)
	rg.SetBounds(geometry.NewRect(0, 0, 200, 40))
	rg.ItemAt(0).SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(10, 10), geometry.Pt(10, 10), event.ModNone)
	consumed := rg.Event(ctx, press)

	if consumed {
		t.Error("disabled group Event should not consume events")
	}
}

// --- Helper functions ---

// simulateClickDefaultX is the fixed X coordinate used for simulated clicks.
const simulateClickDefaultX float32 = 10

func simulateClick(it *radio.Item, ctx widget.Context, y float32) {
	pt := geometry.Pt(simulateClickDefaultX, y)
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		pt, pt, event.ModNone)
	it.Event(ctx, press)

	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		pt, pt, event.ModNone)
	it.Event(ctx, release)
}

// --- testPainter records the call to PaintRadio ---

type testPainter struct {
	called bool
	state  radio.PaintState
}

func (p *testPainter) PaintRadio(_ widget.Canvas, ps radio.PaintState) {
	p.called = true
	p.state = ps
}

// --- recordingCanvas records draw calls for verification ---

type recordingCanvas struct {
	drawTexts      []drawTextCall
	drawCircles    []drawCircleCall
	strokeCircles  []strokeCircleCall
	drawRoundRects []drawRoundRectCall
}

type drawTextCall struct {
	text     string
	bounds   geometry.Rect
	fontSize float32
	color    widget.Color
	bold     bool
	align    float32
}

type drawCircleCall struct {
	center geometry.Point
	radius float32
	color  widget.Color
}

type strokeCircleCall struct {
	center      geometry.Point
	radius      float32
	color       widget.Color
	strokeWidth float32
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

func (c *recordingCanvas) DrawCircle(center geometry.Point, radius float32, color widget.Color) {
	c.drawCircles = append(c.drawCircles, drawCircleCall{center: center, radius: radius, color: color})
}

func (c *recordingCanvas) StrokeCircle(center geometry.Point, radius float32, color widget.Color, strokeWidth float32) {
	c.strokeCircles = append(c.strokeCircles, strokeCircleCall{center: center, radius: radius, color: color, strokeWidth: strokeWidth})
}

func (c *recordingCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32) {}

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

// --- Signal Binding Tests (public API) ---

func TestRadioSelectedSignal(t *testing.T) {
	sig := state.NewSignal("b")
	rg := radio.NewGroup(
		radio.Items(
			radio.ItemDef{Value: "a", Label: "Alpha"},
			radio.ItemDef{Value: "b", Label: "Beta"},
			radio.ItemDef{Value: "c", Label: "Gamma"},
		),
		radio.SelectedSignal(sig),
	)

	// Signal provides initial selection.
	if rg.Selected() != "b" {
		t.Errorf("Selected() = %q, want %q (from signal)", rg.Selected(), "b")
	}

	// External signal update is reflected.
	sig.Set("c")
	if rg.Selected() != "c" {
		t.Errorf("Selected() = %q, want %q (after signal update)", rg.Selected(), "c")
	}

	// User click writes back to signal (two-way).
	rg.SetBounds(geometry.NewRect(0, 0, 200, 100))
	rg.ItemAt(0).SetBounds(geometry.NewRect(0, 0, 200, 30))
	rg.ItemAt(1).SetBounds(geometry.NewRect(0, 30, 200, 30))
	rg.ItemAt(2).SetBounds(geometry.NewRect(0, 60, 200, 30))
	ctx := widget.NewContext()

	simulateClick(rg.ItemAt(0), ctx, 10)

	if sig.Get() != "a" {
		t.Errorf("signal = %q, want %q (two-way write-back)", sig.Get(), "a")
	}
	if rg.Selected() != "a" {
		t.Errorf("Selected() = %q, want %q", rg.Selected(), "a")
	}
}

func TestRadioGroupDisabledSignal(t *testing.T) {
	sig := state.NewSignal(true)
	rg := radio.NewGroup(
		radio.Items(radio.ItemDef{Value: "a", Label: "Alpha"}),
		radio.GroupDisabledSignal(sig),
	)

	if rg.ItemAt(0).IsFocusable() {
		t.Error("items should not be focusable when disabled via signal")
	}

	sig.Set(false)

	if !rg.ItemAt(0).IsFocusable() {
		t.Error("items should be focusable when enabled via signal")
	}
}

func TestRadioSignalPriority(t *testing.T) {
	t.Run("SelectedSignal overrides Selected", func(t *testing.T) {
		sig := state.NewSignal("b")
		rg := radio.NewGroup(
			radio.Items(
				radio.ItemDef{Value: "a", Label: "Alpha"},
				radio.ItemDef{Value: "b", Label: "Beta"},
			),
			radio.Selected("a"),
			radio.SelectedSignal(sig),
		)

		if rg.Selected() != "b" {
			t.Errorf("Selected() = %q, want %q (signal > static)", rg.Selected(), "b")
		}
	})

	t.Run("GroupDisabledSignal overrides GroupDisabledFn and GroupDisabled", func(t *testing.T) {
		sig := state.NewSignal(true)
		rg := radio.NewGroup(
			radio.Items(radio.ItemDef{Value: "a", Label: "Alpha"}),
			radio.GroupDisabled(false),
			radio.GroupDisabledFn(func() bool { return false }),
			radio.GroupDisabledSignal(sig),
		)

		if rg.ItemAt(0).IsFocusable() {
			t.Error("signal(true) should override fn(false) and static(false)")
		}
	})
}

func TestRadioSelectedSignal_WithOnChange(t *testing.T) {
	sig := state.NewSignal("")
	var lastValue string
	callCount := 0
	rg := radio.NewGroup(
		radio.Items(
			radio.ItemDef{Value: "a", Label: "Alpha"},
			radio.ItemDef{Value: "b", Label: "Beta"},
		),
		radio.SelectedSignal(sig),
		radio.OnChange(func(v string) {
			callCount++
			lastValue = v
		}),
	)
	rg.SetBounds(geometry.NewRect(0, 0, 200, 100))
	rg.ItemAt(0).SetBounds(geometry.NewRect(0, 0, 200, 30))
	rg.ItemAt(1).SetBounds(geometry.NewRect(0, 30, 200, 30))
	ctx := widget.NewContext()

	simulateClick(rg.ItemAt(0), ctx, 10)

	if callCount != 1 {
		t.Errorf("onChange called %d times, want 1", callCount)
	}
	if lastValue != "a" {
		t.Errorf("lastValue = %q, want %q", lastValue, "a")
	}
	if sig.Get() != "a" {
		t.Errorf("signal = %q, want %q", sig.Get(), "a")
	}
}
