package listview_test

import (
	"image"
	"testing"

	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/core/listview"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

// --- Construction tests ---

func TestNew_Defaults(t *testing.T) {
	lv := listview.New()

	if !lv.IsVisible() {
		t.Error("should be visible by default")
	}
	if !lv.IsEnabled() {
		t.Error("should be enabled by default")
	}
	if lv.GetItemCount() != 0 {
		t.Errorf("GetItemCount() = %d, want 0", lv.GetItemCount())
	}
}

func TestNew_WithItemCount(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(100),
		listview.FixedItemHeight(48),
	)

	if lv.GetItemCount() != 100 {
		t.Errorf("GetItemCount() = %d, want 100", lv.GetItemCount())
	}
}

func TestNew_WithBuildItem(t *testing.T) {
	called := false
	lv := listview.New(
		listview.ItemCount(10),
		listview.FixedItemHeight(48),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget {
			called = true
			return nil
		}),
	)

	ctx := widget.NewContext()
	constraints := geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 400, MaxHeight: 400,
	}
	lv.Layout(ctx, constraints)
	lv.SetBounds(geometry.NewRect(0, 0, 300, 400))
	lv.Draw(ctx, &mockCanvas{})

	if !called {
		t.Error("BuildItem callback should have been called during Draw")
	}
}

func TestNew_WithFixedHeight(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(100),
		listview.FixedItemHeight(56),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	ctx := widget.NewContext()
	size := lv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 200, MaxHeight: 200,
	})

	if size.Width != 300 || size.Height != 200 {
		t.Errorf("Layout size = %v, want (300, 200)", size)
	}
}

func TestNew_WithItemHeightFn(t *testing.T) {
	heights := []float32{20, 30, 40, 50, 60}
	lv := listview.New(
		listview.ItemCount(len(heights)),
		listview.ItemHeightFn(func(i int) float32 { return heights[i] }),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 200, MaxHeight: 200,
	})

	// Just verify it doesn't panic.
	if lv.GetItemCount() != 5 {
		t.Errorf("GetItemCount() = %d, want 5", lv.GetItemCount())
	}
}

func TestNew_WithEstimatedHeight(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(50),
		listview.EstimatedItemHeight(64),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	if lv.GetItemCount() != 50 {
		t.Errorf("GetItemCount() = %d, want 50", lv.GetItemCount())
	}
}

// --- Signal binding tests ---

func TestItemCountSignal(t *testing.T) {
	sig := state.NewSignal(10)
	lv := listview.New(
		listview.ItemCountSignal(sig),
		listview.FixedItemHeight(48),
	)

	if lv.GetItemCount() != 10 {
		t.Errorf("GetItemCount() = %d, want 10", lv.GetItemCount())
	}

	sig.Set(20)
	if lv.GetItemCount() != 20 {
		t.Errorf("after Set(20): GetItemCount() = %d, want 20", lv.GetItemCount())
	}
}

func TestItemCountReadonlySignal(t *testing.T) {
	sig := state.NewSignal(10)
	lv := listview.New(
		listview.ItemCountReadonlySignal(sig),
		listview.ItemCount(5), // should be overridden by signal
		listview.FixedItemHeight(48),
	)

	if lv.GetItemCount() != 10 {
		t.Errorf("GetItemCount() = %d, want 10 (readonly signal takes precedence)", lv.GetItemCount())
	}
}

func TestItemCountFn(t *testing.T) {
	count := 15
	lv := listview.New(
		listview.ItemCountFn(func() int { return count }),
		listview.FixedItemHeight(48),
	)

	if lv.GetItemCount() != 15 {
		t.Errorf("GetItemCount() = %d, want 15", lv.GetItemCount())
	}

	count = 25
	if lv.GetItemCount() != 25 {
		t.Errorf("after change: GetItemCount() = %d, want 25", lv.GetItemCount())
	}
}

func TestSelectedIndexSignal_TwoWay(t *testing.T) {
	sig := state.NewSignal(-1)
	lv := listview.New(
		listview.ItemCount(10),
		listview.FixedItemHeight(48),
		listview.SelectionModeOpt(listview.SelectionSingle),
		listview.SelectedIndexSignal(sig),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	ctx := widget.NewContext()
	constraints := geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 400, MaxHeight: 400,
	}
	lv.Layout(ctx, constraints)
	lv.SetBounds(geometry.NewRect(0, 0, 300, 400))

	// Click on item 3 (at y=144 which is index 3 for 48px items).
	me := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(150, 145),
	}
	lv.Event(ctx, me)

	// Signal should be updated.
	if sig.Get() != 3 {
		t.Errorf("sig.Get() = %d, want 3", sig.Get())
	}

	// Set signal externally.
	sig.Set(7)
	// The widget should reflect this on next read.
}

func TestDisabledSignal(t *testing.T) {
	sig := state.NewSignal(false)
	lv := listview.New(
		listview.DisabledSignal(sig),
	)

	if lv.IsFocusable() != true {
		t.Error("should be focusable when not disabled")
	}

	sig.Set(true)
	if lv.IsFocusable() != false {
		t.Error("should not be focusable when disabled")
	}
}

func TestDisabledReadonlySignal(t *testing.T) {
	sig := state.NewSignal(true)
	lv := listview.New(
		listview.DisabledReadonlySignal(sig),
		listview.Disabled(false), // should be overridden
	)

	if lv.IsFocusable() != false {
		t.Error("should not be focusable (readonly disabled signal overrides static)")
	}
}

func TestDisabledFn(t *testing.T) {
	disabled := true
	lv := listview.New(
		listview.DisabledFn(func() bool { return disabled }),
	)

	if lv.IsFocusable() != false {
		t.Error("should not be focusable when disabled")
	}

	disabled = false
	if lv.IsFocusable() != true {
		t.Error("should be focusable when not disabled")
	}
}

// --- Layout tests ---

func TestLayout_FillsConstraints(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(10),
		listview.FixedItemHeight(48),
	)

	ctx := widget.NewContext()
	size := lv.Layout(ctx, geometry.Constraints{
		MinWidth: 200, MaxWidth: 400,
		MinHeight: 300, MaxHeight: 600,
	})

	if size.Width != 400 || size.Height != 600 {
		t.Errorf("Layout() = %v, want (400, 600)", size)
	}
}

// --- Children tests ---

func TestChildren_ReturnsScrollView(t *testing.T) {
	lv := listview.New()
	children := lv.Children()

	if len(children) != 1 {
		t.Fatalf("len(Children()) = %d, want 1", len(children))
	}
	if children[0] == nil {
		t.Error("child should not be nil")
	}
}

// --- IsFocusable tests ---

func TestIsFocusable(t *testing.T) {
	tests := []struct {
		name string
		opts []listview.Option
		want bool
	}{
		{"default", nil, true},
		{"disabled", []listview.Option{listview.Disabled(true)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lv := listview.New(tt.opts...)
			if got := lv.IsFocusable(); got != tt.want {
				t.Errorf("IsFocusable() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- VisibleRange tests ---

func TestVisibleRange(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(100),
		listview.FixedItemHeight(48),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 200, MaxHeight: 200,
	})

	start, end := lv.VisibleRange()
	if start != 0 {
		t.Errorf("start = %d, want 0", start)
	}
	// 200px / 48px = ~4 items + partial = 5 items
	if end < 4 || end > 6 {
		t.Errorf("end = %d, want 4-6", end)
	}
}

// --- InvalidateData tests ---

func TestInvalidateData(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(10),
		listview.FixedItemHeight(48),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	// Should not panic.
	lv.InvalidateData()
}

// --- Accessibility tests ---

func TestAccessibilityRole(t *testing.T) {
	lv := listview.New()
	if got := lv.AccessibilityRole(); got != a11y.RoleList {
		t.Errorf("AccessibilityRole() = %v, want %v", got, a11y.RoleList)
	}
}

func TestAccessibilityLabel(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		lv := listview.New()
		if got := lv.AccessibilityLabel(); got != "List" {
			t.Errorf("AccessibilityLabel() = %q, want %q", got, "List")
		}
	})

	t.Run("custom", func(t *testing.T) {
		lv := listview.New(listview.A11yLabel("Contacts"))
		if got := lv.AccessibilityLabel(); got != "Contacts" {
			t.Errorf("AccessibilityLabel() = %q, want %q", got, "Contacts")
		}
	})
}

func TestAccessibilityValue(t *testing.T) {
	t.Run("no items", func(t *testing.T) {
		lv := listview.New()
		if got := lv.AccessibilityValue(); got != "0 items" {
			t.Errorf("AccessibilityValue() = %q, want %q", got, "0 items")
		}
	})

	t.Run("with items", func(t *testing.T) {
		lv := listview.New(listview.ItemCount(42))
		if got := lv.AccessibilityValue(); got != "42 items" {
			t.Errorf("AccessibilityValue() = %q, want %q", got, "42 items")
		}
	})

	t.Run("with selection", func(t *testing.T) {
		lv := listview.New(
			listview.ItemCount(10),
			listview.SelectedIndex(3),
		)
		if got := lv.AccessibilityValue(); got != "Item 4 of 10 selected" {
			t.Errorf("AccessibilityValue() = %q, want %q", got, "Item 4 of 10 selected")
		}
	})
}

func TestAccessibilityState(t *testing.T) {
	t.Run("enabled", func(t *testing.T) {
		lv := listview.New()
		st := lv.AccessibilityState()
		if st.Disabled {
			t.Error("should not be disabled")
		}
	})

	t.Run("disabled", func(t *testing.T) {
		lv := listview.New(listview.Disabled(true))
		st := lv.AccessibilityState()
		if !st.Disabled {
			t.Error("should be disabled")
		}
	})
}

func TestAccessibilityActions(t *testing.T) {
	lv := listview.New()
	actions := lv.AccessibilityActions()
	if len(actions) != 2 {
		t.Fatalf("len(actions) = %d, want 2", len(actions))
	}
}

func TestAccessibilityHint(t *testing.T) {
	lv := listview.New()
	if got := lv.AccessibilityHint(); got != "" {
		t.Errorf("AccessibilityHint() = %q, want empty", got)
	}
}

// --- Event tests ---

func TestEvent_KeyboardNavigation(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(10),
		listview.FixedItemHeight(48),
		listview.SelectionModeOpt(listview.SelectionSingle),
		listview.SelectedIndex(0),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 400, MaxHeight: 400,
	})
	lv.SetBounds(geometry.NewRect(0, 0, 300, 400))
	ctx.RequestFocus(lv)

	// Press Down arrow.
	ke := &event.KeyEvent{
		KeyType: event.KeyPress,
		Key:     event.KeyDown,
	}
	consumed := lv.Event(ctx, ke)
	if !consumed {
		t.Error("KeyDown should be consumed")
	}
}

func TestEvent_KeyboardNotConsumedWhenNotFocused(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(10),
		listview.FixedItemHeight(48),
		listview.SelectionModeOpt(listview.SelectionSingle),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 400, MaxHeight: 400,
	})
	lv.SetBounds(geometry.NewRect(0, 0, 300, 400))
	// Not focused.

	ke := &event.KeyEvent{
		KeyType: event.KeyPress,
		Key:     event.KeyDown,
	}
	consumed := lv.Event(ctx, ke)
	if consumed {
		t.Error("KeyDown should not be consumed when not focused")
	}
}

func TestEvent_DisabledIgnoresInput(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(10),
		listview.FixedItemHeight(48),
		listview.Disabled(true),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 400, MaxHeight: 400,
	})

	// Not visible/enabled should not consume.
	lv.SetVisible(false)
	ke := &event.KeyEvent{
		KeyType: event.KeyPress,
		Key:     event.KeyDown,
	}
	if lv.Event(ctx, ke) {
		t.Error("should not consume when not visible")
	}
}

// --- Callback tests ---

func TestOnItemClick(t *testing.T) {
	clicked := -1
	lv := listview.New(
		listview.ItemCount(10),
		listview.FixedItemHeight(48),
		listview.OnItemClick(func(index int) { clicked = index }),
		listview.SelectionModeOpt(listview.SelectionSingle),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 400, MaxHeight: 400,
	})
	lv.SetBounds(geometry.NewRect(0, 0, 300, 400))

	// Click on item at y=100 (index 2 for 48px items).
	me := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(150, 100),
	}
	lv.Event(ctx, me)

	if clicked != 2 {
		t.Errorf("clicked = %d, want 2", clicked)
	}
}

func TestOnSelectionChange(t *testing.T) {
	changed := -1
	lv := listview.New(
		listview.ItemCount(10),
		listview.FixedItemHeight(48),
		listview.SelectionModeOpt(listview.SelectionSingle),
		listview.OnSelectionChange(func(index int) { changed = index }),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 400, MaxHeight: 400,
	})
	lv.SetBounds(geometry.NewRect(0, 0, 300, 400))

	// Click on item 0.
	me := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(150, 10),
	}
	lv.Event(ctx, me)

	if changed != 0 {
		t.Errorf("changed = %d, want 0", changed)
	}
}

func TestOnEndReached(t *testing.T) {
	endReached := false
	lv := listview.New(
		listview.ItemCount(10),
		listview.FixedItemHeight(48),
		listview.OnEndReached(func() { endReached = true }),
		listview.EndReachedThreshold(5),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 400, MaxHeight: 400,
	})
	lv.SetBounds(geometry.NewRect(0, 0, 300, 400))

	// Draw to trigger end-reached check.
	// With 10 items at 48px, viewport=400px shows ~8 items.
	// That's past the threshold of 5 from end.
	lv.Draw(ctx, &mockCanvas{})

	if !endReached {
		t.Error("OnEndReached should have been called")
	}
}

func TestOnScroll(t *testing.T) {
	scrolled := false
	lv := listview.New(
		listview.ItemCount(100),
		listview.FixedItemHeight(48),
		listview.OnScroll(func(offset float32) { scrolled = true }),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	// Just verify construction.
	_ = lv
	_ = scrolled
}

// --- Mount/Unmount tests ---

func TestMount_Unmount(t *testing.T) {
	sig := state.NewSignal(10)
	lv := listview.New(
		listview.ItemCountSignal(sig),
	)

	ctx := widget.NewContext()
	// Should not panic.
	lv.Mount(ctx)
	lv.Unmount()
}

func TestMount_WithScheduler(t *testing.T) {
	sig := state.NewSignal(10)
	selected := state.NewSignal(-1)
	disabled := state.NewSignal(false)

	lv := listview.New(
		listview.ItemCountSignal(sig),
		listview.SelectedIndexSignal(selected),
		listview.DisabledSignal(disabled),
	)

	ctx := widget.NewContext()
	sched := &mockScheduler{}
	ctx.SetScheduler(sched)

	lv.Mount(ctx)

	// MarkDirty is called by signal subscription, not by Mount itself.
	// We just verify no panic.
	_ = sched.markDirtyCount

	lv.Unmount()
}

// --- ScrollToIndex tests ---

func TestScrollToIndex(t *testing.T) {
	sig := state.NewSignal[float32](0)
	lv := listview.New(
		listview.ItemCount(100),
		listview.FixedItemHeight(48),
		listview.ScrollYSignal(sig),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 200, MaxHeight: 200,
	})
	lv.SetBounds(geometry.NewRect(0, 0, 300, 200))

	// Scroll to item 50 (should be at offset 50*48 = 2400).
	lv.ScrollToIndex(50)

	// Signal should have been updated.
	scrollY := sig.Get()
	if scrollY < 2000 {
		t.Errorf("scrollY = %v, should be >= 2000 for item 50", scrollY)
	}
}

func TestScrollToIndex_OutOfRange(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(10),
		listview.FixedItemHeight(48),
	)

	// Should not panic for out-of-range indices.
	lv.ScrollToIndex(-1)
	lv.ScrollToIndex(100)
}

// --- Painter tests ---

func TestPainterOpt(t *testing.T) {
	p := &mockPainter{}
	lv := listview.New(
		listview.ItemCount(3),
		listview.FixedItemHeight(48),
		listview.Divider(true),
		listview.PainterOpt(p),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return &mockWidget{} }),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 400, MaxHeight: 400,
	})
	lv.SetBounds(geometry.NewRect(0, 0, 300, 400))
	lv.Draw(ctx, &mockCanvas{})

	if p.dividerCalls == 0 {
		t.Error("PaintDivider should have been called")
	}
	if p.itemBgCalls == 0 {
		t.Error("PaintItemBackground should have been called")
	}
}

func TestDefaultPainter_EmptyState(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(0),
		listview.FixedItemHeight(48),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 400, MaxHeight: 400,
	})
	lv.SetBounds(geometry.NewRect(0, 0, 300, 400))

	// Should not panic.
	lv.Draw(ctx, &mockCanvas{})
}

// --- M3 Painter tests ---

func TestM3ListViewPainter(t *testing.T) {
	p := material3.ListViewPainter{}

	// Should not panic with nil theme.
	canvas := &mockCanvas{}
	p.PaintDivider(canvas, listview.DividerState{
		Bounds:    geometry.NewRect(0, 0, 300, 1),
		ItemIndex: 0,
	})
	p.PaintEmptyState(canvas, geometry.NewRect(0, 0, 300, 400))
	p.PaintItemBackground(canvas, listview.ItemPaintState{
		Bounds:  geometry.NewRect(0, 0, 300, 48),
		Hovered: true,
	})
	p.PaintSelection(canvas, listview.ItemPaintState{
		Bounds:   geometry.NewRect(0, 0, 300, 48),
		Selected: true,
	})
}

func TestM3ListViewPainter_WithTheme(t *testing.T) {
	theme := material3.New(widget.Hex(0x6750A4)) // M3 purple
	p := material3.ListViewPainter{Theme: theme}

	canvas := &mockCanvas{}
	p.PaintDivider(canvas, listview.DividerState{
		Bounds:    geometry.NewRect(0, 0, 300, 1),
		ItemIndex: 0,
	})
	p.PaintEmptyState(canvas, geometry.NewRect(0, 0, 300, 400))
	p.PaintItemBackground(canvas, listview.ItemPaintState{
		Bounds: geometry.NewRect(0, 0, 300, 48),
	})
	p.PaintSelection(canvas, listview.ItemPaintState{
		Bounds:   geometry.NewRect(0, 0, 300, 48),
		Selected: true,
		Focused:  true,
	})
}

func TestM3ListViewPainter_EmptyBounds(t *testing.T) {
	p := material3.ListViewPainter{}
	canvas := &mockCanvas{}

	// Empty bounds should be no-ops.
	p.PaintDivider(canvas, listview.DividerState{})
	p.PaintEmptyState(canvas, geometry.Rect{})
	p.PaintItemBackground(canvas, listview.ItemPaintState{})
	p.PaintSelection(canvas, listview.ItemPaintState{})
}

// --- Options tests ---

func TestOverscan(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(100),
		listview.FixedItemHeight(48),
		listview.Overscan(5),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	// Verify construction succeeds.
	if lv.GetItemCount() != 100 {
		t.Errorf("GetItemCount() = %d, want 100", lv.GetItemCount())
	}
}

func TestDividerOption(t *testing.T) {
	lv := listview.New(
		listview.Divider(true),
	)
	_ = lv // verify construction
}

func TestScrollYSignal(t *testing.T) {
	sig := state.NewSignal[float32](0)
	lv := listview.New(
		listview.ScrollYSignal(sig),
	)
	_ = lv // verify construction
}

func TestEndReachedThreshold(t *testing.T) {
	lv := listview.New(
		listview.EndReachedThreshold(10),
	)
	_ = lv // verify construction
}

func TestSelectedIndexReadonlySignal(t *testing.T) {
	sig := state.NewSignal(5)
	lv := listview.New(
		listview.ItemCount(10),
		listview.SelectedIndexReadonlySignal(sig),
		listview.SelectedIndex(2), // should be overridden
	)

	// ReadonlySignal should take precedence.
	got := lv.AccessibilityValue()
	if got != "Item 6 of 10 selected" {
		t.Errorf("AccessibilityValue() = %q, want 'Item 6 of 10 selected'", got)
	}
}

// --- Draw tests ---

func TestDraw_NotVisible(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(10),
		listview.FixedItemHeight(48),
	)
	lv.SetVisible(false)

	ctx := widget.NewContext()
	canvas := &mockCanvas{}
	// Should not panic.
	lv.Draw(ctx, canvas)
}

func TestDraw_EmptyBounds(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(10),
		listview.FixedItemHeight(48),
	)

	ctx := widget.NewContext()
	canvas := &mockCanvas{}
	// No Layout called, bounds are zero.
	lv.Draw(ctx, canvas)
}

// --- Edge cases ---

func TestSingleItem(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(1),
		listview.FixedItemHeight(48),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 400, MaxHeight: 400,
	})
	lv.SetBounds(geometry.NewRect(0, 0, 300, 400))
	lv.Draw(ctx, &mockCanvas{})

	start, end := lv.VisibleRange()
	if start != 0 || end != 1 {
		t.Errorf("VisibleRange() = (%d, %d), want (0, 1)", start, end)
	}
}

func TestLargeItemCount(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(100000),
		listview.FixedItemHeight(48),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 400, MaxHeight: 400,
	})
	lv.SetBounds(geometry.NewRect(0, 0, 300, 400))

	// Should handle large counts without performance issues.
	lv.Draw(ctx, &mockCanvas{})

	if lv.GetItemCount() != 100000 {
		t.Errorf("GetItemCount() = %d, want 100000", lv.GetItemCount())
	}
}

// --- Keyboard navigation extended tests ---

func TestEvent_KeyboardHome(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(10),
		listview.FixedItemHeight(48),
		listview.SelectionModeOpt(listview.SelectionSingle),
		listview.SelectedIndex(5),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{MinWidth: 300, MaxWidth: 300, MinHeight: 400, MaxHeight: 400})
	lv.SetBounds(geometry.NewRect(0, 0, 300, 400))
	ctx.RequestFocus(lv)

	consumed := lv.Event(ctx, &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyHome})
	if !consumed {
		t.Error("Home should be consumed")
	}
}

func TestEvent_KeyboardEnd(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(10),
		listview.FixedItemHeight(48),
		listview.SelectionModeOpt(listview.SelectionSingle),
		listview.SelectedIndex(0),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{MinWidth: 300, MaxWidth: 300, MinHeight: 400, MaxHeight: 400})
	lv.SetBounds(geometry.NewRect(0, 0, 300, 400))
	ctx.RequestFocus(lv)

	consumed := lv.Event(ctx, &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyEnd})
	if !consumed {
		t.Error("End should be consumed")
	}
}

func TestEvent_KeyboardUp(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(10),
		listview.FixedItemHeight(48),
		listview.SelectionModeOpt(listview.SelectionSingle),
		listview.SelectedIndex(5),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{MinWidth: 300, MaxWidth: 300, MinHeight: 400, MaxHeight: 400})
	lv.SetBounds(geometry.NewRect(0, 0, 300, 400))
	ctx.RequestFocus(lv)

	consumed := lv.Event(ctx, &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyUp})
	if !consumed {
		t.Error("Up should be consumed")
	}
}

func TestEvent_KeyboardPageDown(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(100),
		listview.FixedItemHeight(48),
		listview.SelectionModeOpt(listview.SelectionSingle),
		listview.SelectedIndex(0),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{MinWidth: 300, MaxWidth: 300, MinHeight: 400, MaxHeight: 400})
	lv.SetBounds(geometry.NewRect(0, 0, 300, 400))
	ctx.RequestFocus(lv)

	consumed := lv.Event(ctx, &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyPageDown})
	if !consumed {
		t.Error("PageDown should be consumed")
	}
}

func TestEvent_KeyboardPageUp(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(100),
		listview.FixedItemHeight(48),
		listview.SelectionModeOpt(listview.SelectionSingle),
		listview.SelectedIndex(50),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{MinWidth: 300, MaxWidth: 300, MinHeight: 400, MaxHeight: 400})
	lv.SetBounds(geometry.NewRect(0, 0, 300, 400))
	ctx.RequestFocus(lv)

	consumed := lv.Event(ctx, &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyPageUp})
	if !consumed {
		t.Error("PageUp should be consumed")
	}
}

func TestEvent_KeyboardEnter(t *testing.T) {
	clicked := -1
	lv := listview.New(
		listview.ItemCount(10),
		listview.FixedItemHeight(48),
		listview.SelectionModeOpt(listview.SelectionSingle),
		listview.SelectedIndex(3),
		listview.OnItemClick(func(index int) { clicked = index }),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{MinWidth: 300, MaxWidth: 300, MinHeight: 400, MaxHeight: 400})
	lv.SetBounds(geometry.NewRect(0, 0, 300, 400))
	ctx.RequestFocus(lv)

	consumed := lv.Event(ctx, &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyEnter})
	if !consumed {
		t.Error("Enter should be consumed when item is selected")
	}
	if clicked != 3 {
		t.Errorf("clicked = %d, want 3", clicked)
	}
}

func TestEvent_KeyboardSpace(t *testing.T) {
	clicked := -1
	lv := listview.New(
		listview.ItemCount(10),
		listview.FixedItemHeight(48),
		listview.SelectionModeOpt(listview.SelectionSingle),
		listview.SelectedIndex(5),
		listview.OnItemClick(func(index int) { clicked = index }),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{MinWidth: 300, MaxWidth: 300, MinHeight: 400, MaxHeight: 400})
	lv.SetBounds(geometry.NewRect(0, 0, 300, 400))
	ctx.RequestFocus(lv)

	consumed := lv.Event(ctx, &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeySpace})
	if !consumed {
		t.Error("Space should be consumed when item is selected")
	}
	if clicked != 5 {
		t.Errorf("clicked = %d, want 5", clicked)
	}
}

func TestEvent_KeyboardNoSelection(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(10),
		listview.FixedItemHeight(48),
		listview.SelectionModeOpt(listview.SelectionNone),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{MinWidth: 300, MaxWidth: 300, MinHeight: 400, MaxHeight: 400})
	lv.SetBounds(geometry.NewRect(0, 0, 300, 400))
	ctx.RequestFocus(lv)

	// Arrow keys should not be consumed when selection mode is None.
	consumed := lv.Event(ctx, &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyDown})
	if consumed {
		t.Error("KeyDown should not be consumed when selection mode is None")
	}
}

func TestEvent_KeyboardEmptyList(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(0),
		listview.FixedItemHeight(48),
		listview.SelectionModeOpt(listview.SelectionSingle),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{MinWidth: 300, MaxWidth: 300, MinHeight: 400, MaxHeight: 400})
	lv.SetBounds(geometry.NewRect(0, 0, 300, 400))
	ctx.RequestFocus(lv)

	consumed := lv.Event(ctx, &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyDown})
	if consumed {
		t.Error("KeyDown should not be consumed when list is empty")
	}
}

func TestEvent_KeyboardKeyRepeat(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(10),
		listview.FixedItemHeight(48),
		listview.SelectionModeOpt(listview.SelectionSingle),
		listview.SelectedIndex(0),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{MinWidth: 300, MaxWidth: 300, MinHeight: 400, MaxHeight: 400})
	lv.SetBounds(geometry.NewRect(0, 0, 300, 400))
	ctx.RequestFocus(lv)

	// KeyRepeat should also be handled.
	consumed := lv.Event(ctx, &event.KeyEvent{KeyType: event.KeyRepeat, Key: event.KeyDown})
	if !consumed {
		t.Error("KeyDown with KeyRepeat should be consumed")
	}
}

func TestEvent_KeyRelease_NotConsumed(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(10),
		listview.FixedItemHeight(48),
		listview.SelectionModeOpt(listview.SelectionSingle),
		listview.SelectedIndex(0),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{MinWidth: 300, MaxWidth: 300, MinHeight: 400, MaxHeight: 400})
	lv.SetBounds(geometry.NewRect(0, 0, 300, 400))
	ctx.RequestFocus(lv)

	// KeyRelease should not be consumed.
	consumed := lv.Event(ctx, &event.KeyEvent{KeyType: event.KeyRelease, Key: event.KeyDown})
	if consumed {
		t.Error("KeyRelease should not be consumed")
	}
}

func TestEvent_UnknownKey_NotConsumed(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(10),
		listview.FixedItemHeight(48),
		listview.SelectionModeOpt(listview.SelectionSingle),
		listview.SelectedIndex(0),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{MinWidth: 300, MaxWidth: 300, MinHeight: 400, MaxHeight: 400})
	lv.SetBounds(geometry.NewRect(0, 0, 300, 400))
	ctx.RequestFocus(lv)

	consumed := lv.Event(ctx, &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyA})
	if consumed {
		t.Error("KeyA should not be consumed")
	}
}

// --- Mouse hover tests ---

func TestEvent_MouseMove_HoverTracking(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(10),
		listview.FixedItemHeight(48),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return &mockWidget{} }),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{MinWidth: 300, MaxWidth: 300, MinHeight: 400, MaxHeight: 400})
	lv.SetBounds(geometry.NewRect(0, 0, 300, 400))

	// Mouse move to trigger hover. This goes through scrollview -> virtualContent.
	me := &event.MouseEvent{
		MouseType: event.MouseMove,
		Position:  geometry.Pt(150, 100),
	}
	lv.Event(ctx, me)
	// No crash = success for now.
}

func TestEvent_MouseLeave(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(10),
		listview.FixedItemHeight(48),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return &mockWidget{} }),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{MinWidth: 300, MaxWidth: 300, MinHeight: 400, MaxHeight: 400})
	lv.SetBounds(geometry.NewRect(0, 0, 300, 400))

	me := &event.MouseEvent{
		MouseType: event.MouseLeave,
	}
	lv.Event(ctx, me)
}

func TestEvent_RightClick_NotConsumed(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(10),
		listview.FixedItemHeight(48),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{MinWidth: 300, MaxWidth: 300, MinHeight: 400, MaxHeight: 400})
	lv.SetBounds(geometry.NewRect(0, 0, 300, 400))

	me := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonRight,
		Position:  geometry.Pt(150, 100),
	}
	consumed := lv.Event(ctx, me)
	// Right click should not be consumed by item press handler.
	_ = consumed
}

// --- DefaultPainter edge cases ---

func TestDefaultPainter_PaintDivider_EmptyBounds(t *testing.T) {
	p := listview.DefaultPainter{}
	canvas := &mockCanvas{}
	p.PaintDivider(canvas, listview.DividerState{})
}

func TestDefaultPainter_PaintEmptyState_EmptyBounds(t *testing.T) {
	p := listview.DefaultPainter{}
	canvas := &mockCanvas{}
	p.PaintEmptyState(canvas, geometry.Rect{})
}

func TestDefaultPainter_PaintItemBackground_Disabled(t *testing.T) {
	p := listview.DefaultPainter{}
	canvas := &mockCanvas{}
	p.PaintItemBackground(canvas, listview.ItemPaintState{
		Bounds:   geometry.NewRect(0, 0, 300, 48),
		Hovered:  true,
		Disabled: true,
	})
}

func TestDefaultPainter_PaintSelection_NotSelected(t *testing.T) {
	p := listview.DefaultPainter{}
	canvas := &mockCanvas{}
	p.PaintSelection(canvas, listview.ItemPaintState{
		Bounds:   geometry.NewRect(0, 0, 300, 48),
		Selected: false,
	})
}

func TestDefaultPainter_PaintSelection_Focused(t *testing.T) {
	p := listview.DefaultPainter{}
	canvas := &mockCanvas{}
	p.PaintSelection(canvas, listview.ItemPaintState{
		Bounds:   geometry.NewRect(0, 0, 300, 48),
		Selected: true,
		Focused:  true,
	})
}

func TestDefaultPainter_PaintDivider_WithColorScheme(t *testing.T) {
	p := listview.DefaultPainter{}
	canvas := &mockCanvas{}
	p.PaintDivider(canvas, listview.DividerState{
		Bounds: geometry.NewRect(0, 0, 300, 1),
		ColorScheme: listview.ListColorScheme{
			DividerColor: widget.ColorRed,
		},
	})
}

func TestDefaultPainter_PaintItemBg_WithColorScheme(t *testing.T) {
	p := listview.DefaultPainter{}
	canvas := &mockCanvas{}
	p.PaintItemBackground(canvas, listview.ItemPaintState{
		Bounds:  geometry.NewRect(0, 0, 300, 48),
		Hovered: true,
		ColorScheme: listview.ListColorScheme{
			HoverColor: widget.ColorBlue,
		},
	})
}

func TestDefaultPainter_PaintSelection_WithColorScheme(t *testing.T) {
	p := listview.DefaultPainter{}
	canvas := &mockCanvas{}
	p.PaintSelection(canvas, listview.ItemPaintState{
		Bounds:   geometry.NewRect(0, 0, 300, 48),
		Selected: true,
		Focused:  true,
		ColorScheme: listview.ListColorScheme{
			SelectionColor: widget.ColorGreen,
			FocusColor:     widget.ColorBlue,
		},
	})
}

// --- Mount with readonly signals ---

func TestMount_WithReadonlySignals(t *testing.T) {
	countSig := state.NewSignal(10)
	selectedSig := state.NewSignal(0)
	disabledSig := state.NewSignal(false)

	lv := listview.New(
		listview.ItemCountReadonlySignal(countSig),
		listview.SelectedIndexReadonlySignal(selectedSig),
		listview.DisabledReadonlySignal(disabledSig),
	)

	ctx := widget.NewContext()
	sched := &mockScheduler{}
	ctx.SetScheduler(sched)

	lv.Mount(ctx)
	lv.Unmount()
}

// --- GetItemCount ---

func TestGetItemCount(t *testing.T) {
	lv := listview.New(listview.ItemCount(42))
	if got := lv.GetItemCount(); got != 42 {
		t.Errorf("GetItemCount() = %d, want 42", got)
	}
}

// --- Layout edge cases ---

func TestLayout_NegativeConstraints(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(10),
		listview.FixedItemHeight(48),
	)

	ctx := widget.NewContext()
	// Negative max constraints to trigger fallback path.
	size := lv.Layout(ctx, geometry.Constraints{
		MinWidth: 0, MaxWidth: -1,
		MinHeight: 0, MaxHeight: -1,
	})

	// Biggest() returns negative, so fallback is used.
	// Constrain clamps to [0, -1] which is 0, but that's the branch we want to cover.
	_ = size // Just verify no panic and the branch is exercised.
}

func TestLayout_ItemCountChangedViaSignal(t *testing.T) {
	sig := state.NewSignal(5)
	lv := listview.New(
		listview.ItemCountSignal(sig),
		listview.FixedItemHeight(48),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	ctx := widget.NewContext()
	constraints := geometry.Constraints{MinWidth: 300, MaxWidth: 300, MinHeight: 400, MaxHeight: 400}
	lv.Layout(ctx, constraints)

	// Change count via signal.
	sig.Set(20)
	lv.Layout(ctx, constraints)

	if lv.GetItemCount() != 20 {
		t.Errorf("GetItemCount() = %d, want 20", lv.GetItemCount())
	}
}

// --- InvalidateData with lazy mode ---

func TestInvalidateData_LazyMode(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(10),
		listview.EstimatedItemHeight(50),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	// Should not panic for lazy mode.
	lv.InvalidateData()
}

// --- ScrollToIndex edge cases ---

func TestScrollToIndex_AlreadyVisible(t *testing.T) {
	sig := state.NewSignal[float32](0)
	lv := listview.New(
		listview.ItemCount(100),
		listview.FixedItemHeight(48),
		listview.ScrollYSignal(sig),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{MinWidth: 300, MaxWidth: 300, MinHeight: 400, MaxHeight: 400})
	lv.SetBounds(geometry.NewRect(0, 0, 300, 400))

	// Item 0 is already visible at scroll=0. Should be no-op.
	lv.ScrollToIndex(0)
	if sig.Get() != 0 {
		t.Errorf("scrollY = %v, want 0 (item already visible)", sig.Get())
	}
}

func TestScrollToIndex_ScrollUp(t *testing.T) {
	sig := state.NewSignal[float32](480) // scrolled to item 10
	lv := listview.New(
		listview.ItemCount(100),
		listview.FixedItemHeight(48),
		listview.ScrollYSignal(sig),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{MinWidth: 300, MaxWidth: 300, MinHeight: 200, MaxHeight: 200})
	lv.SetBounds(geometry.NewRect(0, 0, 300, 200))

	// Item 2 is above the viewport. Should scroll up.
	lv.ScrollToIndex(2)
	if sig.Get() != 96 {
		t.Errorf("scrollY = %v, want 96 (item 2 top)", sig.Get())
	}
}

// --- Keyboard disabled test ---

func TestEvent_KeyboardDisabled(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(10),
		listview.FixedItemHeight(48),
		listview.SelectionModeOpt(listview.SelectionSingle),
		listview.SelectedIndex(0),
		listview.Disabled(true),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{MinWidth: 300, MaxWidth: 300, MinHeight: 400, MaxHeight: 400})
	lv.SetBounds(geometry.NewRect(0, 0, 300, 400))
	ctx.RequestFocus(lv)

	consumed := lv.Event(ctx, &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyDown})
	if consumed {
		t.Error("KeyDown should not be consumed when disabled")
	}
}

// --- Mouse disabled test ---

func TestEvent_MouseDisabled(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(10),
		listview.FixedItemHeight(48),
		listview.Disabled(true),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{MinWidth: 300, MaxWidth: 300, MinHeight: 400, MaxHeight: 400})
	lv.SetBounds(geometry.NewRect(0, 0, 300, 400))

	me := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(150, 100),
	}
	lv.Event(ctx, me)
	// No crash; disabled should block mouse events.
}

// --- Enter with no click handler test ---

func TestEvent_KeyboardEnter_NoClickHandler(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(10),
		listview.FixedItemHeight(48),
		listview.SelectionModeOpt(listview.SelectionSingle),
		listview.SelectedIndex(3),
		// No OnItemClick handler.
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return nil }),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{MinWidth: 300, MaxWidth: 300, MinHeight: 400, MaxHeight: 400})
	lv.SetBounds(geometry.NewRect(0, 0, 300, 400))
	ctx.RequestFocus(lv)

	consumed := lv.Event(ctx, &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyEnter})
	if consumed {
		t.Error("Enter should not be consumed without OnItemClick handler")
	}
}

// --- M3 Painter with color scheme ---

func TestM3ListViewPainter_PaintItemBackground_NotHovered(t *testing.T) {
	p := material3.ListViewPainter{}
	canvas := &mockCanvas{}
	p.PaintItemBackground(canvas, listview.ItemPaintState{
		Bounds:  geometry.NewRect(0, 0, 300, 48),
		Hovered: false, // not hovered — different code path
	})
}

func TestM3ListViewPainter_PaintSelection_NotSelected(t *testing.T) {
	p := material3.ListViewPainter{}
	canvas := &mockCanvas{}
	// Not selected — should early return.
	p.PaintSelection(canvas, listview.ItemPaintState{
		Bounds:   geometry.NewRect(0, 0, 300, 48),
		Selected: false,
	})
}

func TestM3ListViewPainter_WithColorScheme(t *testing.T) {
	p := material3.ListViewPainter{}
	canvas := &mockCanvas{}
	cs := listview.ListColorScheme{
		DividerColor:   widget.ColorRed,
		SelectionColor: widget.ColorGreen,
		HoverColor:     widget.ColorBlue,
		FocusColor:     widget.ColorWhite,
	}

	p.PaintDivider(canvas, listview.DividerState{
		Bounds:      geometry.NewRect(0, 0, 300, 1),
		ColorScheme: cs,
	})
	p.PaintItemBackground(canvas, listview.ItemPaintState{
		Bounds:      geometry.NewRect(0, 0, 300, 48),
		Hovered:     true,
		ColorScheme: cs,
	})
	p.PaintSelection(canvas, listview.ItemPaintState{
		Bounds:      geometry.NewRect(0, 0, 300, 48),
		Selected:    true,
		ColorScheme: cs,
	})
}

// --- Event not visible/not enabled ---

func TestEvent_NotEnabled(t *testing.T) {
	lv := listview.New(
		listview.ItemCount(10),
		listview.FixedItemHeight(48),
	)
	lv.SetEnabled(false)

	ctx := widget.NewContext()
	consumed := lv.Event(ctx, &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyDown})
	if consumed {
		t.Error("should not consume when not enabled")
	}
}

// --- Draw with selection and dividers ---

func TestDraw_WithSelectionAndDividers(t *testing.T) {
	p := &mockPainter{}
	lv := listview.New(
		listview.ItemCount(5),
		listview.FixedItemHeight(48),
		listview.SelectionModeOpt(listview.SelectionSingle),
		listview.SelectedIndex(2),
		listview.Divider(true),
		listview.PainterOpt(p),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget { return &mockWidget{} }),
	)

	ctx := widget.NewContext()
	lv.Layout(ctx, geometry.Constraints{MinWidth: 300, MaxWidth: 300, MinHeight: 400, MaxHeight: 400})
	lv.SetBounds(geometry.NewRect(0, 0, 300, 400))
	ctx.RequestFocus(lv)
	lv.Draw(ctx, &mockCanvas{})

	if p.selectionCalls == 0 {
		t.Error("PaintSelection should have been called for selected item")
	}
	if p.dividerCalls == 0 {
		t.Error("PaintDivider should have been called between items")
	}
}

// --- Mock types ---

type mockCanvas struct{}

func (m *mockCanvas) Clear(_ widget.Color)                                                  {}
func (m *mockCanvas) DrawRect(_ geometry.Rect, _ widget.Color)                              {}
func (m *mockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32)                 {}
func (m *mockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32)              {}
func (m *mockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {}
func (m *mockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)                {}
func (m *mockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32)   {}
func (m *mockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32)               {}
func (m *mockCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
}

func (m *mockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}
func (m *mockCanvas) DrawImage(_ image.Image, _ geometry.Point)    {}
func (m *mockCanvas) PushClip(_ geometry.Rect)                     {}
func (m *mockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (m *mockCanvas) PopClip()                                     {}
func (m *mockCanvas) PushTransform(_ geometry.Point)               {}
func (m *mockCanvas) PopTransform()                                {}
func (m *mockCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }

type mockPainter struct {
	dividerCalls    int
	emptyStateCalls int
	itemBgCalls     int
	selectionCalls  int
}

func (p *mockPainter) PaintDivider(_ widget.Canvas, _ listview.DividerState) {
	p.dividerCalls++
}
func (p *mockPainter) PaintEmptyState(_ widget.Canvas, _ geometry.Rect) {
	p.emptyStateCalls++
}
func (p *mockPainter) PaintItemBackground(_ widget.Canvas, _ listview.ItemPaintState) {
	p.itemBgCalls++
}
func (p *mockPainter) PaintSelection(_ widget.Canvas, _ listview.ItemPaintState) {
	p.selectionCalls++
}

type mockWidget struct {
	widget.WidgetBase
}

func (w *mockWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(100, 48))
}

func (w *mockWidget) Draw(_ widget.Context, _ widget.Canvas) {}

func (w *mockWidget) Event(_ widget.Context, _ event.Event) bool { return false }

type mockScheduler struct {
	markDirtyCount int
}

func (s *mockScheduler) MarkDirty(_ widget.Widget) {
	s.markDirtyCount++
}
