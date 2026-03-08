package radio

import (
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// --- Config Tests ---

func TestConfig_ResolvedDisabled(t *testing.T) {
	t.Run("static disabled", func(t *testing.T) {
		c := groupConfig{disabled: true}
		if !c.ResolvedDisabled() {
			t.Error("ResolvedDisabled() should be true")
		}
	})

	t.Run("static enabled", func(t *testing.T) {
		c := groupConfig{disabled: false}
		if c.ResolvedDisabled() {
			t.Error("ResolvedDisabled() should be false")
		}
	})

	t.Run("dynamic disabled", func(t *testing.T) {
		c := groupConfig{disabled: false, disabledFn: func() bool { return true }}
		if !c.ResolvedDisabled() {
			t.Error("ResolvedDisabled() with fn should be true")
		}
	})

	t.Run("dynamic overrides static", func(t *testing.T) {
		c := groupConfig{disabled: true, disabledFn: func() bool { return false }}
		if c.ResolvedDisabled() {
			t.Error("ResolvedDisabled() with fn returning false should be false")
		}
	})
}

// --- Direction Tests ---

func TestDirection_String(t *testing.T) {
	tests := []struct {
		name string
		dir  Direction
		want string
	}{
		{"vertical", Vertical, "Vertical"},
		{"horizontal", Horizontal, "Horizontal"},
		{"unknown", Direction(99), "Unknown"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.dir.String(); got != tc.want {
				t.Errorf("Direction(%d).String() = %q, want %q", tc.dir, got, tc.want)
			}
		})
	}
}

// --- Options Tests ---

func TestOptions(t *testing.T) {
	t.Run("OnChange", func(t *testing.T) {
		var c groupConfig
		called := false
		OnChange(func(string) { called = true })(&c)
		if c.onChange == nil {
			t.Fatal("onChange should not be nil")
		}
		c.onChange("test")
		if !called {
			t.Error("onChange should have been called")
		}
	})

	t.Run("Selected", func(t *testing.T) {
		var c groupConfig
		Selected("x")(&c)
		if c.selected != "x" {
			t.Errorf("selected = %q, want %q", c.selected, "x")
		}
	})

	t.Run("DirectionOpt", func(t *testing.T) {
		var c groupConfig
		DirectionOpt(Horizontal)(&c)
		if c.direction != Horizontal {
			t.Errorf("direction = %v, want %v", c.direction, Horizontal)
		}
	})

	t.Run("GroupDisabled", func(t *testing.T) {
		var c groupConfig
		GroupDisabled(true)(&c)
		if !c.disabled {
			t.Error("disabled should be true")
		}
	})

	t.Run("GroupDisabledFn", func(t *testing.T) {
		var c groupConfig
		GroupDisabledFn(func() bool { return true })(&c)
		if c.disabledFn == nil {
			t.Error("disabledFn should not be nil")
		}
	})

	t.Run("GroupA11yLabel", func(t *testing.T) {
		var c groupConfig
		GroupA11yLabel("Choose one")(&c)
		if c.a11yLabel != "Choose one" {
			t.Errorf("a11yLabel = %q, want %q", c.a11yLabel, "Choose one")
		}
	})

	t.Run("GroupPainter", func(t *testing.T) {
		var c groupConfig
		p := DefaultPainter{}
		GroupPainter(p)(&c)
		if c.painter == nil {
			t.Fatal("painter should not be nil")
		}
	})

	t.Run("Items", func(t *testing.T) {
		var c groupConfig
		Items(
			ItemDef{Value: "a", Label: "Alpha"},
			ItemDef{Value: "b", Label: "Beta"},
		)(&c)
		if len(c.items) != 2 {
			t.Fatalf("items len = %d, want 2", len(c.items))
		}
		if c.items[0].Value != "a" {
			t.Errorf("items[0].Value = %q, want %q", c.items[0].Value, "a")
		}
		if c.items[1].Label != "Beta" {
			t.Errorf("items[1].Label = %q, want %q", c.items[1].Label, "Beta")
		}
	})
}

// --- Widget Construction Tests ---

func TestNewGroup_Defaults(t *testing.T) {
	g := NewGroup()

	if !g.IsVisible() {
		t.Error("should be visible by default")
	}
	if !g.IsEnabled() {
		t.Error("should be enabled by default")
	}
	if g.selected != -1 {
		t.Errorf("selected = %d, want -1", g.selected)
	}
	if len(g.items) != 0 {
		t.Errorf("items len = %d, want 0", len(g.items))
	}
}

func TestNewGroup_WithItems(t *testing.T) {
	g := NewGroup(
		Items(
			ItemDef{Value: "a", Label: "Alpha"},
			ItemDef{Value: "b", Label: "Beta"},
		),
		Selected("b"),
	)

	if len(g.items) != 2 {
		t.Fatalf("items len = %d, want 2", len(g.items))
	}
	if g.selected != 1 {
		t.Errorf("selected = %d, want 1", g.selected)
	}
	if g.items[0].value != "a" {
		t.Errorf("items[0].value = %q, want 'a'", g.items[0].value)
	}
	if g.items[1].label != "Beta" {
		t.Errorf("items[1].label = %q, want 'Beta'", g.items[1].label)
	}
}

func TestNewGroup_DefaultPainter(t *testing.T) {
	g := NewGroup(Items(ItemDef{Value: "a", Label: "Alpha"}))
	if g.items[0].painter == nil {
		t.Error("painter should not be nil by default")
	}
	if _, ok := g.items[0].painter.(DefaultPainter); !ok {
		t.Errorf("default painter should be DefaultPainter, got %T", g.items[0].painter)
	}
}

func TestNewGroup_CustomPainter(t *testing.T) {
	custom := &internalTestPainter{}
	g := NewGroup(
		Items(ItemDef{Value: "a", Label: "Alpha"}),
		GroupPainter(custom),
	)
	if g.items[0].painter != custom {
		t.Error("painter should be the custom painter")
	}
}

func TestNewGroup_Children(t *testing.T) {
	g := NewGroup(
		Items(
			ItemDef{Value: "a", Label: "Alpha"},
			ItemDef{Value: "b", Label: "Beta"},
		),
	)
	children := g.Children()
	if len(children) != 2 {
		t.Errorf("Children() len = %d, want 2", len(children))
	}
}

func TestNewGroup_SelectedNotFound(t *testing.T) {
	g := NewGroup(
		Items(ItemDef{Value: "a", Label: "Alpha"}),
		Selected("nonexistent"),
	)
	if g.selected != -1 {
		t.Errorf("selected = %d, want -1 for nonexistent value", g.selected)
	}
}

// --- Selection Logic Tests ---

func TestSelectValue(t *testing.T) {
	g := NewGroup(
		Items(
			ItemDef{Value: "a", Label: "Alpha"},
			ItemDef{Value: "b", Label: "Beta"},
		),
	)

	g.selectValue("a")
	if g.selected != 0 {
		t.Errorf("selected = %d, want 0 after selecting 'a'", g.selected)
	}

	g.selectValue("b")
	if g.selected != 1 {
		t.Errorf("selected = %d, want 1 after selecting 'b'", g.selected)
	}
}

func TestSelectValue_CallsOnChange(t *testing.T) {
	callCount := 0
	var lastValue string
	g := NewGroup(
		Items(ItemDef{Value: "a", Label: "Alpha"}),
		OnChange(func(v string) {
			callCount++
			lastValue = v
		}),
	)

	g.selectValue("a")
	if callCount != 1 {
		t.Errorf("onChange called %d times, want 1", callCount)
	}
	if lastValue != "a" {
		t.Errorf("lastValue = %q, want 'a'", lastValue)
	}
}

func TestSelectValue_NilOnChange(t *testing.T) {
	g := NewGroup(Items(ItemDef{Value: "a", Label: "Alpha"}))
	// Should not panic.
	g.selectValue("a")
}

func TestIsSelected(t *testing.T) {
	g := NewGroup(
		Items(
			ItemDef{Value: "a", Label: "Alpha"},
			ItemDef{Value: "b", Label: "Beta"},
		),
		Selected("a"),
	)

	if !g.isSelected("a") {
		t.Error("'a' should be selected")
	}
	if g.isSelected("b") {
		t.Error("'b' should not be selected")
	}
}

func TestIsSelected_NoSelection(t *testing.T) {
	g := NewGroup(
		Items(ItemDef{Value: "a", Label: "Alpha"}),
	)

	if g.isSelected("a") {
		t.Error("'a' should not be selected when no selection")
	}
}

func TestSetSelectedLocked(t *testing.T) {
	g := NewGroup(
		Items(
			ItemDef{Value: "a", Label: "Alpha"},
			ItemDef{Value: "b", Label: "Beta"},
		),
	)

	g.mu.Lock()
	g.setSelectedLocked("b")
	g.mu.Unlock()
	if g.selected != 1 {
		t.Errorf("selected = %d, want 1", g.selected)
	}

	g.mu.Lock()
	g.setSelectedLocked("nonexistent")
	g.mu.Unlock()
	if g.selected != -1 {
		t.Errorf("selected = %d, want -1 for nonexistent", g.selected)
	}
}

// --- IsFocusable Tests ---

func TestIsFocusable_Group(t *testing.T) {
	g := NewGroup(Items(ItemDef{Value: "a", Label: "Alpha"}))
	if g.IsFocusable() {
		t.Error("group should not be focusable")
	}
}

func TestIsFocusable_Item(t *testing.T) {
	tests := []struct {
		name          string
		visible       bool
		enabled       bool
		disabled      bool
		wantFocusable bool
	}{
		{"all true", true, true, false, true},
		{"invisible", false, true, false, false},
		{"not enabled", true, false, false, false},
		{"group disabled", true, true, true, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := NewGroup(
				Items(ItemDef{Value: "a", Label: "Alpha"}),
				GroupDisabled(tc.disabled),
			)
			it := g.items[0]
			it.SetVisible(tc.visible)
			it.SetEnabled(tc.enabled)

			if got := it.IsFocusable(); got != tc.wantFocusable {
				t.Errorf("IsFocusable() = %v, want %v", got, tc.wantFocusable)
			}
		})
	}
}

func TestIsFocusable_DisabledFn(t *testing.T) {
	isDisabled := true
	g := NewGroup(
		Items(ItemDef{Value: "a", Label: "Alpha"}),
		GroupDisabledFn(func() bool { return isDisabled }),
	)

	if g.items[0].IsFocusable() {
		t.Error("should not be focusable when DisabledFn returns true")
	}

	isDisabled = false
	if !g.items[0].IsFocusable() {
		t.Error("should be focusable when DisabledFn returns false")
	}
}

// --- Layout Tests ---

func TestLayout_MinHeight(t *testing.T) {
	g := NewGroup(Items(ItemDef{Value: "a", Label: "Alpha"}))
	g.SetBounds(geometry.NewRect(0, 0, 200, 200))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 400))

	size := g.Layout(ctx, constraints)

	if size.Height < itemMinHeight {
		t.Errorf("height = %v, should be at least %v", size.Height, itemMinHeight)
	}
}

func TestLayout_WithLabel(t *testing.T) {
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 400))

	noLabel := NewGroup(Items(ItemDef{Value: "a", Label: ""}))
	noLabel.SetBounds(geometry.NewRect(0, 0, 200, 200))
	sizeNoLabel := noLabel.Layout(ctx, constraints)

	withLabel := NewGroup(Items(ItemDef{Value: "a", Label: "Some Label"}))
	withLabel.SetBounds(geometry.NewRect(0, 0, 200, 200))
	sizeWithLabel := withLabel.Layout(ctx, constraints)

	if sizeWithLabel.Width <= sizeNoLabel.Width {
		t.Error("item with label should be wider than without")
	}
}

func TestLayout_Constrained(t *testing.T) {
	g := NewGroup(Items(ItemDef{Value: "a", Label: "Alpha"}))
	g.SetBounds(geometry.NewRect(0, 0, 200, 200))
	ctx := widget.NewContext()
	constraints := geometry.Tight(geometry.Sz(50, 30))

	size := g.Layout(ctx, constraints)

	if size.Width != 50 {
		t.Errorf("width = %v, want 50 (tight constraint)", size.Width)
	}
	if size.Height != 30 {
		t.Errorf("height = %v, want 30 (tight constraint)", size.Height)
	}
}

func TestLayout_VerticalPositioning(t *testing.T) {
	g := NewGroup(
		Items(
			ItemDef{Value: "a", Label: "Alpha"},
			ItemDef{Value: "b", Label: "Beta"},
		),
		DirectionOpt(Vertical),
	)
	g.SetBounds(geometry.NewRect(10, 20, 200, 200))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 400))

	g.Layout(ctx, constraints)

	// Items are positioned in local coordinates (0,0)-based.
	// The group's Draw applies PushTransform(g.Bounds().Min).
	if g.items[0].Bounds().Min.X != 0 {
		t.Errorf("item 0 X = %v, want 0 (local coords)", g.items[0].Bounds().Min.X)
	}
	if g.items[0].Bounds().Min.Y != 0 {
		t.Errorf("item 0 Y = %v, want 0 (local coords)", g.items[0].Bounds().Min.Y)
	}

	// Second item should be below the first.
	if g.items[1].Bounds().Min.Y <= g.items[0].Bounds().Min.Y {
		t.Error("item 1 should be below item 0 in vertical layout")
	}
}

func TestLayout_HorizontalPositioning(t *testing.T) {
	g := NewGroup(
		Items(
			ItemDef{Value: "a", Label: "Alpha"},
			ItemDef{Value: "b", Label: "Beta"},
		),
		DirectionOpt(Horizontal),
	)
	g.SetBounds(geometry.NewRect(10, 20, 400, 50))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 400))

	g.Layout(ctx, constraints)

	// Items are positioned in local coordinates (0,0)-based.
	if g.items[0].Bounds().Min.X != 0 {
		t.Errorf("item 0 X = %v, want 0 (local coords)", g.items[0].Bounds().Min.X)
	}

	// Second item should be to the right of the first.
	if g.items[1].Bounds().Min.X <= g.items[0].Bounds().Min.X {
		t.Error("item 1 should be to the right of item 0 in horizontal layout")
	}
}

// --- State Transition Tests ---

func TestStateTransition_HoverEnterLeave(t *testing.T) {
	g := NewGroup(Items(ItemDef{Value: "a", Label: "Alpha"}))
	g.SetBounds(geometry.NewRect(0, 0, 200, 40))
	g.items[0].SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	enterEvt := event.NewMouseEvent(event.MouseEnter, event.ButtonNone, 0,
		geometry.Pt(10, 10), geometry.Pt(10, 10), event.ModNone)
	consumed := handleItemEvent(g.items[0], ctx, enterEvt)

	if !consumed {
		t.Error("MouseEnter should be consumed")
	}
	if g.items[0].state != stateHover {
		t.Errorf("state = %v, want stateHover", g.items[0].state)
	}

	leaveEvt := event.NewMouseEvent(event.MouseLeave, event.ButtonNone, 0,
		geometry.Pt(300, 300), geometry.Pt(300, 300), event.ModNone)
	consumed = handleItemEvent(g.items[0], ctx, leaveEvt)

	if !consumed {
		t.Error("MouseLeave should be consumed")
	}
	if g.items[0].state != stateNormal {
		t.Errorf("state = %v, want stateNormal", g.items[0].state)
	}
}

func TestStateTransition_PressRelease(t *testing.T) {
	selected := false
	g := NewGroup(
		Items(ItemDef{Value: "a", Label: "Alpha"}),
		OnChange(func(string) { selected = true }),
	)
	g.SetBounds(geometry.NewRect(0, 0, 200, 40))
	g.items[0].SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	pressEvt := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(10, 10), geometry.Pt(10, 10), event.ModNone)
	consumed := handleItemEvent(g.items[0], ctx, pressEvt)

	if !consumed {
		t.Error("MousePress should be consumed")
	}
	if g.items[0].state != statePressed {
		t.Errorf("state = %v, want statePressed", g.items[0].state)
	}
	if selected {
		t.Error("should not select on press, only on release")
	}

	releaseEvt := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		geometry.Pt(10, 10), geometry.Pt(10, 10), event.ModNone)
	consumed = handleItemEvent(g.items[0], ctx, releaseEvt)

	if !consumed {
		t.Error("MouseRelease should be consumed")
	}
	if !selected {
		t.Error("should have selected on release inside bounds")
	}
	if g.items[0].state != stateHover {
		t.Errorf("state = %v, want stateHover (released inside bounds)", g.items[0].state)
	}
}

func TestStateTransition_PressReleaseOutside(t *testing.T) {
	selected := false
	g := NewGroup(
		Items(ItemDef{Value: "a", Label: "Alpha"}),
		OnChange(func(string) { selected = true }),
	)
	g.SetBounds(geometry.NewRect(0, 0, 200, 40))
	g.items[0].SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	pressEvt := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(10, 10), geometry.Pt(10, 10), event.ModNone)
	handleItemEvent(g.items[0], ctx, pressEvt)

	releaseEvt := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		geometry.Pt(300, 300), geometry.Pt(300, 300), event.ModNone)
	handleItemEvent(g.items[0], ctx, releaseEvt)

	if selected {
		t.Error("should not select when released outside bounds")
	}
	if g.items[0].state != stateNormal {
		t.Errorf("state = %v, want stateNormal (released outside)", g.items[0].state)
	}
}

func TestStateTransition_RightButton_Ignored(t *testing.T) {
	g := NewGroup(Items(ItemDef{Value: "a", Label: "Alpha"}))
	g.SetBounds(geometry.NewRect(0, 0, 200, 40))
	g.items[0].SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	pressEvt := event.NewMouseEvent(event.MousePress, event.ButtonRight, event.ButtonStateRight,
		geometry.Pt(10, 10), geometry.Pt(10, 10), event.ModNone)
	consumed := handleItemEvent(g.items[0], ctx, pressEvt)

	if consumed {
		t.Error("right button press should not be consumed")
	}
	if g.items[0].state != stateNormal {
		t.Error("state should remain normal for right button")
	}
}

// --- Disabled State Tests ---

func TestDisabled_IgnoresEvents(t *testing.T) {
	selected := false
	g := NewGroup(
		Items(ItemDef{Value: "a", Label: "Alpha"}),
		OnChange(func(string) { selected = true }),
		GroupDisabled(true),
	)
	g.SetBounds(geometry.NewRect(0, 0, 200, 40))
	g.items[0].SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	enterEvt := event.NewMouseEvent(event.MouseEnter, event.ButtonNone, 0,
		geometry.Pt(10, 10), geometry.Pt(10, 10), event.ModNone)
	consumed := handleItemEvent(g.items[0], ctx, enterEvt)
	if consumed {
		t.Error("disabled item should not consume MouseEnter")
	}

	pressEvt := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(10, 10), geometry.Pt(10, 10), event.ModNone)
	consumed = handleItemEvent(g.items[0], ctx, pressEvt)
	if consumed {
		t.Error("disabled item should not consume MousePress")
	}

	if selected {
		t.Error("disabled item should not fire select")
	}
}

func TestDisabledFn_Reactive(t *testing.T) {
	isDisabled := false
	selected := false
	g := NewGroup(
		Items(ItemDef{Value: "a", Label: "Alpha"}),
		OnChange(func(string) { selected = true }),
		GroupDisabledFn(func() bool { return isDisabled }),
	)
	g.SetBounds(geometry.NewRect(0, 0, 200, 40))
	g.items[0].SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	pressEvt := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(10, 10), geometry.Pt(10, 10), event.ModNone)
	handleItemEvent(g.items[0], ctx, pressEvt)

	releaseEvt := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		geometry.Pt(10, 10), geometry.Pt(10, 10), event.ModNone)
	handleItemEvent(g.items[0], ctx, releaseEvt)

	if !selected {
		t.Error("should select when not disabled")
	}

	selected = false
	isDisabled = true

	pressEvt2 := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(10, 10), geometry.Pt(10, 10), event.ModNone)
	consumed := handleItemEvent(g.items[0], ctx, pressEvt2)

	if consumed {
		t.Error("disabled item should not consume events")
	}
	if selected {
		t.Error("disabled item should not fire select")
	}
}

// --- Keyboard Activation Tests ---

func TestKeyboard_SpaceActivation(t *testing.T) {
	selected := false
	g := NewGroup(
		Items(ItemDef{Value: "a", Label: "Alpha"}),
		OnChange(func(string) { selected = true }),
	)
	g.items[0].SetFocused(true)
	g.items[0].SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	pressEvt := event.NewKeyEvent(event.KeyPress, event.KeySpace, 0, event.ModNone)
	consumed := handleItemEvent(g.items[0], ctx, pressEvt)
	if !consumed {
		t.Error("Space press should be consumed when focused")
	}
	if g.items[0].state != statePressed {
		t.Errorf("state = %v, want statePressed", g.items[0].state)
	}

	releaseEvt := event.NewKeyEvent(event.KeyRelease, event.KeySpace, 0, event.ModNone)
	consumed = handleItemEvent(g.items[0], ctx, releaseEvt)
	if !consumed {
		t.Error("Space release should be consumed when focused")
	}
	if !selected {
		t.Error("Space release should trigger selection")
	}
	if g.items[0].state != stateNormal {
		t.Errorf("state = %v, want stateNormal after key release", g.items[0].state)
	}
}

func TestKeyboard_EnterActivation(t *testing.T) {
	selected := false
	g := NewGroup(
		Items(ItemDef{Value: "a", Label: "Alpha"}),
		OnChange(func(string) { selected = true }),
	)
	g.items[0].SetFocused(true)
	g.items[0].SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	pressEvt := event.NewKeyEvent(event.KeyPress, event.KeyEnter, 0, event.ModNone)
	consumed := handleItemEvent(g.items[0], ctx, pressEvt)
	if !consumed {
		t.Error("Enter press should be consumed when focused")
	}

	releaseEvt := event.NewKeyEvent(event.KeyRelease, event.KeyEnter, 0, event.ModNone)
	consumed = handleItemEvent(g.items[0], ctx, releaseEvt)
	if !consumed {
		t.Error("Enter release should be consumed when focused")
	}
	if !selected {
		t.Error("Enter release should trigger selection")
	}
}

func TestKeyboard_NotFocused_Ignored(t *testing.T) {
	selected := false
	g := NewGroup(
		Items(ItemDef{Value: "a", Label: "Alpha"}),
		OnChange(func(string) { selected = true }),
	)
	g.items[0].SetFocused(false)
	g.items[0].SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	pressEvt := event.NewKeyEvent(event.KeyPress, event.KeySpace, 0, event.ModNone)
	consumed := handleItemEvent(g.items[0], ctx, pressEvt)

	if consumed {
		t.Error("Space should not be consumed when not focused")
	}
	if selected {
		t.Error("should not select when not focused")
	}
}

func TestKeyboard_OtherKeys_Ignored(t *testing.T) {
	g := NewGroup(Items(ItemDef{Value: "a", Label: "Alpha"}))
	g.items[0].SetFocused(true)
	g.items[0].SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	pressEvt := event.NewKeyEvent(event.KeyPress, event.KeyA, 'a', event.ModNone)
	consumed := handleItemEvent(g.items[0], ctx, pressEvt)

	if consumed {
		t.Error("key A should not be consumed by radio item")
	}
}

func TestKeyboard_Disabled_Ignored(t *testing.T) {
	selected := false
	g := NewGroup(
		Items(ItemDef{Value: "a", Label: "Alpha"}),
		OnChange(func(string) { selected = true }),
		GroupDisabled(true),
	)
	g.items[0].SetFocused(true)
	g.items[0].SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	pressEvt := event.NewKeyEvent(event.KeyPress, event.KeySpace, 0, event.ModNone)
	consumed := handleItemEvent(g.items[0], ctx, pressEvt)

	if consumed {
		t.Error("disabled item should not consume key events")
	}
	if selected {
		t.Error("disabled item should not fire select")
	}
}

// --- Navigation Tests ---

func TestMoveFocus_Forward(t *testing.T) {
	g := NewGroup(
		Items(
			ItemDef{Value: "a", Label: "Alpha"},
			ItemDef{Value: "b", Label: "Beta"},
			ItemDef{Value: "c", Label: "Gamma"},
		),
	)
	for i := range g.items {
		g.items[i].SetBounds(geometry.NewRect(0, float32(i*30), 200, 30))
	}
	ctx := widget.NewContext()
	ctx.RequestFocus(g.items[0])

	g.moveFocus(g.items[0], ctx, 1)

	if !g.items[1].IsFocused() {
		t.Error("focus should move to item 1")
	}
}

func TestMoveFocus_Backward(t *testing.T) {
	g := NewGroup(
		Items(
			ItemDef{Value: "a", Label: "Alpha"},
			ItemDef{Value: "b", Label: "Beta"},
			ItemDef{Value: "c", Label: "Gamma"},
		),
	)
	for i := range g.items {
		g.items[i].SetBounds(geometry.NewRect(0, float32(i*30), 200, 30))
	}
	ctx := widget.NewContext()
	ctx.RequestFocus(g.items[2])

	g.moveFocus(g.items[2], ctx, -1)

	if !g.items[1].IsFocused() {
		t.Error("focus should move to item 1")
	}
}

func TestMoveFocus_WrapForward(t *testing.T) {
	g := NewGroup(
		Items(
			ItemDef{Value: "a", Label: "Alpha"},
			ItemDef{Value: "b", Label: "Beta"},
		),
	)
	for i := range g.items {
		g.items[i].SetBounds(geometry.NewRect(0, float32(i*30), 200, 30))
	}
	ctx := widget.NewContext()
	ctx.RequestFocus(g.items[1])

	g.moveFocus(g.items[1], ctx, 1)

	if !g.items[0].IsFocused() {
		t.Error("focus should wrap to item 0")
	}
}

func TestMoveFocus_WrapBackward(t *testing.T) {
	g := NewGroup(
		Items(
			ItemDef{Value: "a", Label: "Alpha"},
			ItemDef{Value: "b", Label: "Beta"},
		),
	)
	for i := range g.items {
		g.items[i].SetBounds(geometry.NewRect(0, float32(i*30), 200, 30))
	}
	ctx := widget.NewContext()
	ctx.RequestFocus(g.items[0])

	g.moveFocus(g.items[0], ctx, -1)

	if !g.items[1].IsFocused() {
		t.Error("focus should wrap to item 1")
	}
}

// --- Draw Tests ---

func TestDraw_EmptyBounds(t *testing.T) {
	g := NewGroup(Items(ItemDef{Value: "a", Label: "Alpha"}))
	// Do not set bounds.
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	g.Draw(ctx, canvas)

	// With empty item bounds, painter should skip drawing.
	if len(canvas.drawCircles) > 0 || len(canvas.strokeCircles) > 0 ||
		len(canvas.drawTexts) > 0 {
		t.Error("should not draw anything with empty bounds")
	}
}

func TestDraw_SelectedState(t *testing.T) {
	g := NewGroup(
		Items(ItemDef{Value: "a", Label: "Alpha"}),
		Selected("a"),
	)
	g.items[0].SetBounds(geometry.NewRect(10, 10, 100, 30))
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	g.Draw(ctx, canvas)

	// Selected: DrawCircle (outer filled) + DrawCircle (inner dot) + DrawText.
	if len(canvas.drawCircles) < 2 {
		t.Errorf("selected should draw at least 2 circles, got %d", len(canvas.drawCircles))
	}
	if len(canvas.drawTexts) == 0 {
		t.Error("should draw label text")
	}
	if canvas.drawTexts[0].text != "Alpha" {
		t.Errorf("text = %q, want %q", canvas.drawTexts[0].text, "Alpha")
	}
}

func TestDraw_UnselectedState(t *testing.T) {
	g := NewGroup(
		Items(ItemDef{Value: "a", Label: "Alpha"}),
	)
	g.items[0].SetBounds(geometry.NewRect(10, 10, 100, 30))
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	g.Draw(ctx, canvas)

	// Unselected: StrokeCircle (border) + DrawText.
	if len(canvas.strokeCircles) == 0 {
		t.Error("unselected should draw border (StrokeCircle)")
	}
	if len(canvas.drawCircles) > 0 {
		t.Error("unselected should not draw filled circles")
	}
	if len(canvas.drawTexts) == 0 {
		t.Error("should draw label text")
	}
}

func TestDraw_NoLabel(t *testing.T) {
	g := NewGroup(Items(ItemDef{Value: "a", Label: ""}))
	g.items[0].SetBounds(geometry.NewRect(10, 10, 30, 30))
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	g.Draw(ctx, canvas)

	if len(canvas.drawTexts) > 0 {
		t.Error("item without label should not draw text")
	}
}

func TestDraw_FocusRing(t *testing.T) {
	g := NewGroup(Items(ItemDef{Value: "a", Label: "Alpha"}))
	g.items[0].SetBounds(geometry.NewRect(10, 10, 100, 30))
	g.items[0].SetFocused(true)
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	g.Draw(ctx, canvas)

	// Focus ring: StrokeCircle with expanded radius.
	found := false
	for _, call := range canvas.strokeCircles {
		if call.radius > outerRadius {
			found = true
			break
		}
	}
	if !found {
		t.Error("focused item should draw a focus ring (expanded StrokeCircle)")
	}
}

func TestDraw_FocusRing_NotWhenDisabled(t *testing.T) {
	g := NewGroup(
		Items(ItemDef{Value: "a", Label: "Alpha"}),
		GroupDisabled(true),
	)
	g.items[0].SetBounds(geometry.NewRect(10, 10, 100, 30))
	g.items[0].SetFocused(true)
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	g.Draw(ctx, canvas)

	for _, call := range canvas.strokeCircles {
		if call.radius > outerRadius {
			t.Error("disabled focused item should not draw focus ring")
			break
		}
	}
}

func TestDraw_DisabledSelected(t *testing.T) {
	g := NewGroup(
		Items(ItemDef{Value: "a", Label: "Alpha"}),
		Selected("a"),
		GroupDisabled(true),
	)
	g.items[0].SetBounds(geometry.NewRect(0, 0, 100, 30))
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	g.Draw(ctx, canvas)

	if len(canvas.drawCircles) < 2 {
		t.Fatal("disabled selected should still draw circles")
	}
	bg := canvas.drawCircles[0].color
	if bg == defaultSelectedBg {
		t.Error("disabled background should differ from normal selected background")
	}
}

func TestDraw_DisabledUnselected(t *testing.T) {
	g := NewGroup(
		Items(ItemDef{Value: "a", Label: "Alpha"}),
		GroupDisabled(true),
	)
	g.items[0].SetBounds(geometry.NewRect(0, 0, 100, 30))
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	g.Draw(ctx, canvas)

	if len(canvas.strokeCircles) == 0 {
		t.Error("disabled unselected should draw border")
	}
	borderColor := canvas.strokeCircles[0].color
	if borderColor == defaultUnselectedBorder {
		t.Error("disabled border should differ from normal border")
	}
}

// --- Painting Helper Tests ---

func TestApplyStateModifier(t *testing.T) {
	base := widget.ColorRed

	normal := applyStateModifier(base, false, false)
	if normal != base {
		t.Error("normal state should return base color unchanged")
	}

	hover := applyStateModifier(base, true, false)
	if hover == base {
		t.Error("hover state should modify the base color")
	}

	pressed := applyStateModifier(base, false, true)
	if pressed == base {
		t.Error("pressed state should modify the base color")
	}
}

func TestApplyStateModifier_PressedOverridesHover(t *testing.T) {
	base := widget.ColorRed

	result := applyStateModifier(base, true, true)
	pressedOnly := applyStateModifier(base, false, true)
	if result != pressedOnly {
		t.Error("pressed should take precedence over hovered")
	}
}

func TestRadioCircleGeometry(t *testing.T) {
	bounds := geometry.NewRect(10, 10, 100, 30)
	center, radius := radioCircleGeometry(bounds)

	if radius != outerRadius {
		t.Errorf("radius = %v, want %v", radius, outerRadius)
	}
	if center.X != 10+outerRadius {
		t.Errorf("center.X = %v, want %v", center.X, 10+outerRadius)
	}
	expectedY := float32(10) + 30/2
	if center.Y != expectedY {
		t.Errorf("center.Y = %v, want %v", center.Y, expectedY)
	}
}

func TestRadioLabelBounds(t *testing.T) {
	bounds := geometry.NewRect(10, 10, 100, 30)
	label := radioLabelBounds(bounds)

	expectedX := float32(10) + outerRadius*2 + labelGap
	if label.Min.X != expectedX {
		t.Errorf("label X = %v, want %v", label.Min.X, expectedX)
	}
	if label.Min.Y != 10 {
		t.Errorf("label Y = %v, want 10", label.Min.Y)
	}
	if label.Width() <= 0 {
		t.Error("label width should be positive")
	}
}

func TestRadioLabelBounds_NarrowBounds(t *testing.T) {
	// Bounds too narrow for label.
	bounds := geometry.NewRect(0, 0, 10, 30)
	label := radioLabelBounds(bounds)

	if label.Width() < 0 {
		t.Error("label width should not be negative")
	}
}

// --- Painter Interface Tests ---

func TestPainter_DelegationToPainter(t *testing.T) {
	p := &internalTestPainter{}
	g := NewGroup(
		Items(ItemDef{Value: "a", Label: "Test"}),
		GroupPainter(p),
		Selected("a"),
	)
	g.items[0].SetBounds(geometry.NewRect(0, 0, 100, 30))
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	g.Draw(ctx, canvas)

	if !p.called {
		t.Error("Draw should delegate to the configured painter")
	}
	if p.state.Label != "Test" {
		t.Errorf("PaintState.Label = %q, want %q", p.state.Label, "Test")
	}
	if !p.state.Selected {
		t.Error("PaintState.Selected should be true")
	}
	if p.state.Bounds.IsEmpty() {
		t.Error("PaintState.Bounds should not be empty")
	}
}

func TestPainter_PaintStateFields(t *testing.T) {
	p := &internalTestPainter{}
	g := NewGroup(
		Items(ItemDef{Value: "a", Label: "Test"}),
		GroupPainter(p),
		Selected("a"),
	)
	g.items[0].SetBounds(geometry.NewRect(0, 0, 100, 30))
	g.items[0].SetFocused(true)
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	g.Draw(ctx, canvas)

	if !p.state.Selected {
		t.Error("PaintState.Selected should be true")
	}
	if !p.state.Focused {
		t.Error("PaintState.Focused should be true")
	}
}

// --- ColorScheme Path Tests ---

func TestPaintWithColorScheme_Selected(t *testing.T) {
	scheme := RadioColorScheme{
		SelectedBg:       widget.ColorRed,
		SelectedFg:       widget.ColorWhite,
		UnselectedBorder: widget.ColorGray,
		LabelColor:       widget.ColorBlack,
		DisabledBg:       widget.ColorLightGray,
		DisabledFg:       widget.ColorDarkGray,
		FocusRing:        widget.ColorBlue,
	}
	canvas := &internalMockCanvas{}
	ps := PaintState{
		Label:       "CS",
		Selected:    true,
		Bounds:      geometry.NewRect(0, 0, 100, 30),
		ColorScheme: scheme,
	}

	p := DefaultPainter{}
	p.PaintRadio(canvas, ps)

	if len(canvas.drawCircles) < 2 {
		t.Fatalf("selected with scheme should draw 2 circles, got %d", len(canvas.drawCircles))
	}
	// Outer circle should use scheme's SelectedBg.
	if canvas.drawCircles[0].color != widget.ColorRed {
		t.Error("outer circle should use ColorScheme.SelectedBg")
	}
	// Inner dot should use scheme's SelectedFg.
	if canvas.drawCircles[1].color != widget.ColorWhite {
		t.Error("inner dot should use ColorScheme.SelectedFg")
	}
	// Label should use scheme's LabelColor.
	if len(canvas.drawTexts) != 1 {
		t.Fatalf("should draw 1 text, got %d", len(canvas.drawTexts))
	}
	if canvas.drawTexts[0].color != widget.ColorBlack {
		t.Error("label should use ColorScheme.LabelColor")
	}
}

func TestPaintWithColorScheme_Unselected(t *testing.T) {
	scheme := RadioColorScheme{
		SelectedBg:       widget.ColorRed,
		SelectedFg:       widget.ColorWhite,
		UnselectedBorder: widget.Hex(0x112233),
		LabelColor:       widget.ColorBlack,
		DisabledBg:       widget.ColorLightGray,
		DisabledFg:       widget.ColorDarkGray,
		FocusRing:        widget.ColorBlue,
	}
	canvas := &internalMockCanvas{}
	ps := PaintState{
		Label:       "CS",
		Selected:    false,
		Bounds:      geometry.NewRect(0, 0, 100, 30),
		ColorScheme: scheme,
	}

	p := DefaultPainter{}
	p.PaintRadio(canvas, ps)

	if len(canvas.strokeCircles) == 0 {
		t.Fatal("unselected with scheme should draw border")
	}
	if canvas.strokeCircles[0].color != widget.Hex(0x112233) {
		t.Error("border should use ColorScheme.UnselectedBorder")
	}
}

func TestPaintWithColorScheme_DisabledSelected(t *testing.T) {
	scheme := RadioColorScheme{
		SelectedBg:       widget.ColorRed,
		SelectedFg:       widget.ColorWhite,
		UnselectedBorder: widget.ColorGray,
		LabelColor:       widget.ColorBlack,
		DisabledBg:       widget.Hex(0xAAAAAA),
		DisabledFg:       widget.Hex(0x999999),
		FocusRing:        widget.ColorBlue,
	}
	canvas := &internalMockCanvas{}
	ps := PaintState{
		Label:       "DS",
		Selected:    true,
		Disabled:    true,
		Bounds:      geometry.NewRect(0, 0, 100, 30),
		ColorScheme: scheme,
	}

	p := DefaultPainter{}
	p.PaintRadio(canvas, ps)

	if len(canvas.drawCircles) < 2 {
		t.Fatalf("disabled selected should draw 2 circles, got %d", len(canvas.drawCircles))
	}
	if canvas.drawCircles[0].color != widget.Hex(0xAAAAAA) {
		t.Error("disabled bg should use ColorScheme.DisabledBg")
	}
	if canvas.drawCircles[1].color != widget.Hex(0x999999) {
		t.Error("disabled fg should use ColorScheme.DisabledFg")
	}
	// Label should use scheme's DisabledFg.
	if len(canvas.drawTexts) != 1 {
		t.Fatalf("should draw 1 text, got %d", len(canvas.drawTexts))
	}
	if canvas.drawTexts[0].color != widget.Hex(0x999999) {
		t.Error("disabled label should use ColorScheme.DisabledFg")
	}
}

func TestPaintWithColorScheme_DisabledUnselected(t *testing.T) {
	scheme := RadioColorScheme{
		SelectedBg:       widget.ColorRed,
		SelectedFg:       widget.ColorWhite,
		UnselectedBorder: widget.ColorGray,
		LabelColor:       widget.ColorBlack,
		DisabledBg:       widget.Hex(0xAAAAAA),
		DisabledFg:       widget.Hex(0x999999),
		FocusRing:        widget.ColorBlue,
	}
	canvas := &internalMockCanvas{}
	ps := PaintState{
		Label:       "DU",
		Selected:    false,
		Disabled:    true,
		Bounds:      geometry.NewRect(0, 0, 100, 30),
		ColorScheme: scheme,
	}

	p := DefaultPainter{}
	p.PaintRadio(canvas, ps)

	if len(canvas.strokeCircles) == 0 {
		t.Fatal("disabled unselected should draw border")
	}
	if canvas.strokeCircles[0].color != widget.Hex(0x999999) {
		t.Error("disabled border should use ColorScheme.DisabledFg")
	}
}

func TestPaintWithColorScheme_FocusRing(t *testing.T) {
	scheme := RadioColorScheme{
		SelectedBg:       widget.ColorRed,
		SelectedFg:       widget.ColorWhite,
		UnselectedBorder: widget.ColorGray,
		LabelColor:       widget.ColorBlack,
		DisabledBg:       widget.ColorLightGray,
		DisabledFg:       widget.ColorDarkGray,
		FocusRing:        widget.Hex(0x0000FF),
	}
	canvas := &internalMockCanvas{}
	ps := PaintState{
		Label:       "FR",
		Selected:    false,
		Focused:     true,
		Bounds:      geometry.NewRect(0, 0, 100, 30),
		ColorScheme: scheme,
	}

	p := DefaultPainter{}
	p.PaintRadio(canvas, ps)

	// Should have focus ring (StrokeCircle with expanded radius).
	found := false
	for _, call := range canvas.strokeCircles {
		if call.radius > outerRadius && call.color == widget.Hex(0x0000FF) {
			found = true
			break
		}
	}
	if !found {
		t.Error("should draw focus ring with ColorScheme.FocusRing color")
	}
}

func TestResolveLabelColor(t *testing.T) {
	scheme := RadioColorScheme{
		LabelColor: widget.ColorRed,
		DisabledFg: widget.ColorGray,
	}

	tests := []struct {
		name      string
		state     PaintState
		hasScheme bool
		want      widget.Color
	}{
		{
			"default no scheme",
			PaintState{},
			false,
			defaultLabelColor,
		},
		{
			"with scheme label",
			PaintState{ColorScheme: scheme},
			true,
			widget.ColorRed,
		},
		{
			"disabled no scheme",
			PaintState{Disabled: true},
			false,
			defaultDisabledFg,
		},
		{
			"disabled with scheme",
			PaintState{Disabled: true, ColorScheme: scheme},
			true,
			widget.ColorGray,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveLabelColor(tc.state, tc.hasScheme)
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

// --- internalTestPainter records the call to PaintRadio ---

type internalTestPainter struct {
	called bool
	state  PaintState
}

func (p *internalTestPainter) PaintRadio(_ widget.Canvas, ps PaintState) {
	p.called = true
	p.state = ps
}

// --- internalMockCanvas records canvas calls for testing ---

type internalMockCanvas struct {
	drawCircles    []internalDrawCircleCall
	strokeCircles  []internalStrokeCircleCall
	drawTexts      []internalDrawTextCall
	drawRoundRects []internalDrawRoundRectCall
}

type internalDrawCircleCall struct {
	center geometry.Point
	radius float32
	color  widget.Color
}

type internalStrokeCircleCall struct {
	center      geometry.Point
	radius      float32
	color       widget.Color
	strokeWidth float32
}

type internalDrawTextCall struct {
	text     string
	bounds   geometry.Rect
	fontSize float32
	color    widget.Color
	bold     bool
	align    float32
}

type internalDrawRoundRectCall struct {
	r      geometry.Rect
	color  widget.Color
	radius float32
}

func (c *internalMockCanvas) Clear(_ widget.Color)                                  {}
func (c *internalMockCanvas) DrawRect(_ geometry.Rect, _ widget.Color)              {}
func (c *internalMockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) {}

func (c *internalMockCanvas) DrawRoundRect(r geometry.Rect, color widget.Color, radius float32) {
	c.drawRoundRects = append(c.drawRoundRects, internalDrawRoundRectCall{r: r, color: color, radius: radius})
}

func (c *internalMockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {
}

func (c *internalMockCanvas) DrawCircle(center geometry.Point, radius float32, color widget.Color) {
	c.drawCircles = append(c.drawCircles, internalDrawCircleCall{center: center, radius: radius, color: color})
}

func (c *internalMockCanvas) StrokeCircle(center geometry.Point, radius float32, color widget.Color, strokeWidth float32) {
	c.strokeCircles = append(c.strokeCircles, internalStrokeCircleCall{center: center, radius: radius, color: color, strokeWidth: strokeWidth})
}

func (c *internalMockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32) {}

func (c *internalMockCanvas) DrawText(text string, bounds geometry.Rect, fontSize float32, color widget.Color, bold bool, align float32) {
	c.drawTexts = append(c.drawTexts, internalDrawTextCall{text: text, bounds: bounds, fontSize: fontSize, color: color, bold: bold, align: align})
}

func (c *internalMockCanvas) PushClip(_ geometry.Rect)       {}
func (c *internalMockCanvas) PopClip()                       {}
func (c *internalMockCanvas) PushTransform(_ geometry.Point) {}
func (c *internalMockCanvas) PopTransform()                  {}

// --- onChange dedup test ---

func TestSelectValue_SameValue_NoOnChange(t *testing.T) {
	callCount := 0
	g := NewGroup(
		Items(ItemDef{Value: "a", Label: "Alpha"}),
		OnChange(func(string) { callCount++ }),
		Selected("a"),
	)

	g.selectValue("a") // re-select same value
	if callCount != 0 {
		t.Errorf("onChange should not fire on re-selecting same value, called %d times", callCount)
	}
}

// --- Signal Binding Tests ---

func TestConfig_ResolvedSelected_Signal(t *testing.T) {
	sig := state.NewSignal("b")
	c := groupConfig{selected: "a", selectedSignal: sig}

	if got := c.ResolvedSelected(); got != "b" {
		t.Errorf("ResolvedSelected() = %q, want %q (signal value)", got, "b")
	}
}

func TestConfig_ResolvedSelected_NoSignal(t *testing.T) {
	c := groupConfig{selected: "a"}

	if got := c.ResolvedSelected(); got != "a" {
		t.Errorf("ResolvedSelected() = %q, want %q (static value)", got, "a")
	}
}

func TestConfig_ResolvedDisabled_Signal(t *testing.T) {
	sig := state.NewSignal(true)
	c := groupConfig{disabled: false, disabledFn: func() bool { return false }, disabledSignal: sig}

	if !c.ResolvedDisabled() {
		t.Error("ResolvedDisabled() should be true (signal value)")
	}
}

func TestConfig_ResolvedDisabled_SignalPriority(t *testing.T) {
	sig := state.NewSignal(true)

	t.Run("signal beats fn and static", func(t *testing.T) {
		c := groupConfig{disabled: false, disabledFn: func() bool { return false }, disabledSignal: sig}
		if !c.ResolvedDisabled() {
			t.Error("signal(true) should override fn(false) and static(false)")
		}
	})

	t.Run("fn beats static when no signal", func(t *testing.T) {
		c := groupConfig{disabled: false, disabledFn: func() bool { return true }}
		if !c.ResolvedDisabled() {
			t.Error("fn(true) should override static(false)")
		}
	})

	t.Run("static used when no signal and no fn", func(t *testing.T) {
		c := groupConfig{disabled: true}
		if !c.ResolvedDisabled() {
			t.Error("static(true) should be used")
		}
	})
}

func TestConfig_ResolvedSelected_SignalPriority(t *testing.T) {
	sig := state.NewSignal("signal-val")

	t.Run("signal beats static", func(t *testing.T) {
		c := groupConfig{selected: "static-val", selectedSignal: sig}
		if got := c.ResolvedSelected(); got != "signal-val" {
			t.Errorf("ResolvedSelected() = %q, want %q (signal > static)", got, "signal-val")
		}
	})

	t.Run("static used when no signal", func(t *testing.T) {
		c := groupConfig{selected: "static-val"}
		if got := c.ResolvedSelected(); got != "static-val" {
			t.Errorf("ResolvedSelected() = %q, want %q", got, "static-val")
		}
	})
}

func TestRadioSelectedSignal_TwoWay(t *testing.T) {
	sig := state.NewSignal("a")
	g := NewGroup(
		Items(
			ItemDef{Value: "a", Label: "Alpha"},
			ItemDef{Value: "b", Label: "Beta"},
		),
		SelectedSignal(sig),
	)
	g.SetBounds(geometry.NewRect(0, 0, 200, 100))
	g.items[0].SetBounds(geometry.NewRect(0, 0, 200, 30))
	g.items[1].SetBounds(geometry.NewRect(0, 30, 200, 30))

	// Signal → widget: initial value should be "a".
	if g.Selected() != "a" {
		t.Errorf("Selected() = %q, want %q (from signal)", g.Selected(), "a")
	}

	// Widget → signal: simulate click on item B.
	ctx := widget.NewContext()
	pressEvt := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(10, 45), geometry.Pt(10, 45), event.ModNone)
	handleItemEvent(g.items[1], ctx, pressEvt)

	releaseEvt := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		geometry.Pt(10, 45), geometry.Pt(10, 45), event.ModNone)
	handleItemEvent(g.items[1], ctx, releaseEvt)

	// Signal should be updated.
	if sig.Get() != "b" {
		t.Errorf("signal value = %q, want %q (two-way write-back)", sig.Get(), "b")
	}

	// Widget reads from signal.
	if g.Selected() != "b" {
		t.Errorf("Selected() = %q, want %q", g.Selected(), "b")
	}
}

func TestRadioSelectedSignal_TwoWay_Keyboard(t *testing.T) {
	sig := state.NewSignal("")
	g := NewGroup(
		Items(
			ItemDef{Value: "a", Label: "Alpha"},
			ItemDef{Value: "b", Label: "Beta"},
		),
		SelectedSignal(sig),
	)
	g.items[0].SetFocused(true)
	g.items[0].SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	press := event.NewKeyEvent(event.KeyPress, event.KeySpace, 0, event.ModNone)
	handleItemEvent(g.items[0], ctx, press)

	release := event.NewKeyEvent(event.KeyRelease, event.KeySpace, 0, event.ModNone)
	handleItemEvent(g.items[0], ctx, release)

	if sig.Get() != "a" {
		t.Errorf("signal value = %q, want %q (keyboard activation)", sig.Get(), "a")
	}
}

func TestRadioSelectedSignal_ExternalUpdate(t *testing.T) {
	sig := state.NewSignal("a")
	g := NewGroup(
		Items(
			ItemDef{Value: "a", Label: "Alpha"},
			ItemDef{Value: "b", Label: "Beta"},
		),
		SelectedSignal(sig),
	)

	// External code updates signal.
	sig.Set("b")

	// Widget should reflect the new signal value.
	if g.Selected() != "b" {
		t.Errorf("Selected() = %q, want %q (external signal update)", g.Selected(), "b")
	}
	if !g.isSelected("b") {
		t.Error("isSelected('b') should be true after external signal update")
	}
	if g.isSelected("a") {
		t.Error("isSelected('a') should be false after external signal update")
	}
}

func TestRadioGroupDisabledSignal(t *testing.T) {
	sig := state.NewSignal(true)
	g := NewGroup(
		Items(ItemDef{Value: "a", Label: "Alpha"}),
		GroupDisabledSignal(sig),
	)

	if !g.cfg.ResolvedDisabled() {
		t.Error("should be disabled when signal is true")
	}
	if g.items[0].IsFocusable() {
		t.Error("items should not be focusable when disabled via signal")
	}

	sig.Set(false)

	if g.cfg.ResolvedDisabled() {
		t.Error("should be enabled when signal is false")
	}
	if !g.items[0].IsFocusable() {
		t.Error("items should be focusable when enabled via signal")
	}
}

func TestRadioGroupDisabledSignal_IgnoresEvents(t *testing.T) {
	selected := false
	sig := state.NewSignal(true)
	g := NewGroup(
		Items(ItemDef{Value: "a", Label: "Alpha"}),
		OnChange(func(string) { selected = true }),
		GroupDisabledSignal(sig),
	)
	g.SetBounds(geometry.NewRect(0, 0, 200, 40))
	g.items[0].SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	pressEvt := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(10, 10), geometry.Pt(10, 10), event.ModNone)
	consumed := handleItemEvent(g.items[0], ctx, pressEvt)

	if consumed {
		t.Error("disabled (via signal) item should not consume events")
	}
	if selected {
		t.Error("disabled (via signal) item should not fire onChange")
	}

	// Enable via signal.
	sig.Set(false)

	consumed = handleItemEvent(g.items[0], ctx, pressEvt)
	if !consumed {
		t.Error("enabled item should consume events")
	}
}

func TestRadioSelectedSignal_OnChangeStillFires(t *testing.T) {
	sig := state.NewSignal("")
	var lastValue string
	callCount := 0
	g := NewGroup(
		Items(
			ItemDef{Value: "a", Label: "Alpha"},
			ItemDef{Value: "b", Label: "Beta"},
		),
		SelectedSignal(sig),
		OnChange(func(v string) {
			callCount++
			lastValue = v
		}),
	)
	g.SetBounds(geometry.NewRect(0, 0, 200, 100))
	g.items[0].SetBounds(geometry.NewRect(0, 0, 200, 30))
	g.items[1].SetBounds(geometry.NewRect(0, 30, 200, 30))
	ctx := widget.NewContext()

	// Click item A.
	pressEvt := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(10, 10), geometry.Pt(10, 10), event.ModNone)
	handleItemEvent(g.items[0], ctx, pressEvt)

	releaseEvt := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		geometry.Pt(10, 10), geometry.Pt(10, 10), event.ModNone)
	handleItemEvent(g.items[0], ctx, releaseEvt)

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

func TestOptions_SelectedSignal(t *testing.T) {
	var c groupConfig
	sig := state.NewSignal("x")
	SelectedSignal(sig)(&c)
	if c.selectedSignal == nil {
		t.Error("selectedSignal should not be nil")
	}
	if c.selectedSignal.Get() != "x" {
		t.Errorf("selectedSignal.Get() = %q, want %q", c.selectedSignal.Get(), "x")
	}
}

func TestOptions_GroupDisabledSignal(t *testing.T) {
	var c groupConfig
	sig := state.NewSignal(true)
	GroupDisabledSignal(sig)(&c)
	if c.disabledSignal == nil {
		t.Error("disabledSignal should not be nil")
	}
	if !c.disabledSignal.Get() {
		t.Error("disabledSignal.Get() should be true")
	}
}
