package gridview_test

import (
	"image"
	"testing"

	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/core/gridview"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// --- Construction tests ---

func TestNew_Defaults(t *testing.T) {
	gv := gridview.New()

	if !gv.IsVisible() {
		t.Error("should be visible by default")
	}
	if !gv.IsEnabled() {
		t.Error("should be enabled by default")
	}
	if gv.GetItemCount() != 0 {
		t.Errorf("GetItemCount() = %d, want 0", gv.GetItemCount())
	}
	if gv.GetColumns() != 4 {
		t.Errorf("GetColumns() = %d, want 4 (default)", gv.GetColumns())
	}
}

func TestNew_WithItemCount(t *testing.T) {
	gv := gridview.New(
		gridview.ItemCount(100),
		gridview.ItemSize(120, 120),
	)

	if gv.GetItemCount() != 100 {
		t.Errorf("GetItemCount() = %d, want 100", gv.GetItemCount())
	}
}

func TestNew_WithBuildCell(t *testing.T) {
	called := false
	gv := gridview.New(
		gridview.ItemCount(10),
		gridview.ItemSize(100, 100),
		gridview.Columns(3),
		gridview.BuildCell(func(index int, ctx gridview.CellContext) widget.Widget {
			called = true
			return nil
		}),
	)

	ctx := widget.NewContext()
	constraints := geometry.Constraints{
		MinWidth: 400, MaxWidth: 400,
		MinHeight: 400, MaxHeight: 400,
	}
	gv.Layout(ctx, constraints)
	gv.SetBounds(geometry.NewRect(0, 0, 400, 400))
	gv.Draw(ctx, &mockCanvas{})

	if !called {
		t.Error("BuildCell callback should have been called during Draw")
	}
}

func TestNew_WithColumns(t *testing.T) {
	gv := gridview.New(
		gridview.ItemCount(20),
		gridview.Columns(5),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 600, MaxWidth: 600,
		MinHeight: 400, MaxHeight: 400,
	})

	if gv.GetColumns() != 5 {
		t.Errorf("GetColumns() = %d, want 5", gv.GetColumns())
	}
}

func TestNew_WithColumnsAuto(t *testing.T) {
	gv := gridview.New(
		gridview.ItemCount(20),
		gridview.ItemSize(100, 100),
		gridview.Gap(10),
		gridview.ColumnsAuto(true),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 340, MaxWidth: 340,
		MinHeight: 400, MaxHeight: 400,
	})

	// 340 / (100 + 10) = 3.09 => 3 columns
	if gv.GetColumns() != 3 {
		t.Errorf("GetColumns() = %d, want 3 (auto-fit)", gv.GetColumns())
	}
}

func TestNew_WithGap(t *testing.T) {
	gv := gridview.New(
		gridview.ItemCount(10),
		gridview.ItemSize(100, 100),
		gridview.Columns(2),
		gridview.Gap(10),
		gridview.BuildCell(func(index int, ctx gridview.CellContext) widget.Widget {
			return nil
		}),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 600, MaxHeight: 600,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 300, 600))
	// Should not panic.
	gv.Draw(ctx, &mockCanvas{})
}

func TestNew_WithPainter(t *testing.T) {
	p := &testPainter{}
	gv := gridview.New(
		gridview.ItemCount(4),
		gridview.ItemSize(100, 100),
		gridview.Columns(2),
		gridview.PainterOpt(p),
		gridview.BuildCell(func(index int, ctx gridview.CellContext) widget.Widget {
			return nil
		}),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 300, MaxHeight: 300,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 300, 300))
	gv.Draw(ctx, &mockCanvas{})

	if p.cellBackgroundCalls == 0 {
		t.Error("Painter.PaintCellBackground was not called")
	}
}

// --- Layout tests ---

func TestLayout_FillsAvailableSpace(t *testing.T) {
	gv := gridview.New(
		gridview.ItemCount(100),
		gridview.ItemSize(100, 100),
		gridview.Columns(3),
	)

	ctx := widget.NewContext()
	size := gv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 400,
		MinHeight: 300, MaxHeight: 500,
	})

	if size.Width != 400 || size.Height != 500 {
		t.Errorf("Layout size = %v, want (400, 500)", size)
	}
}

func TestLayout_InfiniteConstraints(t *testing.T) {
	gv := gridview.New(
		gridview.ItemCount(10),
		gridview.ItemSize(100, 100),
		gridview.Columns(2),
	)

	ctx := widget.NewContext()
	size := gv.Layout(ctx, geometry.Constraints{
		MinWidth:  200,
		MaxWidth:  geometry.Infinity,
		MinHeight: 0,
		MaxHeight: geometry.Infinity,
	})

	// Width should fallback, height should be clamped.
	if size.Width <= 0 || size.Height <= 0 {
		t.Errorf("Layout returned zero size for infinite constraints: %v", size)
	}
}

// --- Draw tests ---

func TestDraw_EmptyGrid(t *testing.T) {
	p := &testPainter{}
	gv := gridview.New(
		gridview.ItemCount(0),
		gridview.PainterOpt(p),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 300, MaxHeight: 300,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 300, 300))
	gv.Draw(ctx, &mockCanvas{})

	if p.emptyStateCalls == 0 {
		t.Error("PaintEmptyState was not called for empty grid")
	}
}

func TestDraw_Invisible(t *testing.T) {
	gv := gridview.New(gridview.ItemCount(10))
	gv.SetVisible(false)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 300, MaxHeight: 300,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 300, 300))
	// Should not panic.
	gv.Draw(ctx, &mockCanvas{})
}

func TestDraw_EmptyBounds(t *testing.T) {
	gv := gridview.New(gridview.ItemCount(10))

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 300, MaxHeight: 300,
	})
	// Don't set bounds -- they are zero.
	gv.Draw(ctx, &mockCanvas{})
}

func TestDraw_WithSelection(t *testing.T) {
	p := &testPainter{}
	gv := gridview.New(
		gridview.ItemCount(10),
		gridview.ItemSize(100, 100),
		gridview.Columns(3),
		gridview.SelectionModeOpt(gridview.SelectionSingle),
		gridview.SelectedIndex(2),
		gridview.PainterOpt(p),
		gridview.BuildCell(func(index int, ctx gridview.CellContext) widget.Widget {
			return nil
		}),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 400, MaxWidth: 400,
		MinHeight: 400, MaxHeight: 400,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 400, 400))
	gv.Draw(ctx, &mockCanvas{})

	if p.selectionCalls == 0 {
		t.Error("PaintSelection was not called for selected item")
	}
}

// --- Selection tests ---

func TestSelectionMode_String(t *testing.T) {
	tests := []struct {
		mode gridview.SelectionMode
		want string
	}{
		{gridview.SelectionNone, "None"},
		{gridview.SelectionSingle, "Single"},
		{gridview.SelectionMode(99), "Unknown"},
	}
	for _, tt := range tests {
		if got := tt.mode.String(); got != tt.want {
			t.Errorf("SelectionMode(%d).String() = %q, want %q", tt.mode, got, tt.want)
		}
	}
}

// --- Visible range tests ---

func TestVisibleRange_EmptyGrid(t *testing.T) {
	gv := gridview.New(gridview.ItemCount(0))

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 300, MaxHeight: 300,
	})

	start, end := gv.VisibleRange()
	if start != 0 || end != 0 {
		t.Errorf("VisibleRange() = (%d, %d), want (0, 0)", start, end)
	}
}

func TestVisibleRange_SmallGrid(t *testing.T) {
	gv := gridview.New(
		gridview.ItemCount(6),
		gridview.ItemSize(100, 100),
		gridview.Columns(3),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 400, MaxWidth: 400,
		MinHeight: 300, MaxHeight: 300,
	})

	start, end := gv.VisibleRange()
	// All 6 items should be visible in a 2-row grid within 300px viewport.
	if start != 0 || end != 6 {
		t.Errorf("VisibleRange() = (%d, %d), want (0, 6)", start, end)
	}
}

// --- Event tests ---

func TestEvent_Disabled(t *testing.T) {
	gv := gridview.New(
		gridview.ItemCount(10),
		gridview.Disabled(true),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 300, MaxHeight: 300,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 300, 300))

	ke := &event.KeyEvent{Key: event.KeyDown, KeyType: event.KeyPress}
	if gv.Event(ctx, ke) {
		t.Error("disabled grid should not consume events")
	}
}

func TestEvent_Invisible(t *testing.T) {
	gv := gridview.New(gridview.ItemCount(10))
	gv.SetVisible(false)

	ctx := widget.NewContext()
	ke := &event.KeyEvent{Key: event.KeyDown, KeyType: event.KeyPress}
	if gv.Event(ctx, ke) {
		t.Error("invisible grid should not consume events")
	}
}

func TestEvent_NotEnabled(t *testing.T) {
	gv := gridview.New(gridview.ItemCount(10))
	gv.SetEnabled(false)

	ctx := widget.NewContext()
	ke := &event.KeyEvent{Key: event.KeyDown, KeyType: event.KeyPress}
	if gv.Event(ctx, ke) {
		t.Error("non-enabled grid should not consume events")
	}
}

// --- Keyboard navigation tests ---

func TestKeyboard_ArrowRight(t *testing.T) {
	gv := gridview.New(
		gridview.ItemCount(12),
		gridview.ItemSize(100, 100),
		gridview.Columns(3),
		gridview.SelectionModeOpt(gridview.SelectionSingle),
		gridview.SelectedIndex(0),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 400, MaxWidth: 400,
		MinHeight: 500, MaxHeight: 500,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 400, 500))
	gv.SetFocused(true)

	ke := &event.KeyEvent{Key: event.KeyRight, KeyType: event.KeyPress}
	consumed := gv.Event(ctx, ke)

	if !consumed {
		t.Error("KeyRight should be consumed when selection is enabled")
	}
}

func TestKeyboard_ArrowDown(t *testing.T) {
	gv := gridview.New(
		gridview.ItemCount(12),
		gridview.ItemSize(100, 100),
		gridview.Columns(3),
		gridview.SelectionModeOpt(gridview.SelectionSingle),
		gridview.SelectedIndex(0),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 400, MaxWidth: 400,
		MinHeight: 500, MaxHeight: 500,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 400, 500))
	gv.SetFocused(true)

	ke := &event.KeyEvent{Key: event.KeyDown, KeyType: event.KeyPress}
	consumed := gv.Event(ctx, ke)

	if !consumed {
		t.Error("KeyDown should be consumed when selection is enabled")
	}
}

func TestKeyboard_HomeEnd(t *testing.T) {
	gv := gridview.New(
		gridview.ItemCount(12),
		gridview.ItemSize(100, 100),
		gridview.Columns(3),
		gridview.SelectionModeOpt(gridview.SelectionSingle),
		gridview.SelectedIndex(5),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 400, MaxWidth: 400,
		MinHeight: 500, MaxHeight: 500,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 400, 500))
	gv.SetFocused(true)

	// Home
	ke := &event.KeyEvent{Key: event.KeyHome, KeyType: event.KeyPress}
	if !gv.Event(ctx, ke) {
		t.Error("KeyHome should be consumed")
	}

	// End
	ke = &event.KeyEvent{Key: event.KeyEnd, KeyType: event.KeyPress}
	if !gv.Event(ctx, ke) {
		t.Error("KeyEnd should be consumed")
	}
}

func TestKeyboard_PageUpDown(t *testing.T) {
	gv := gridview.New(
		gridview.ItemCount(100),
		gridview.ItemSize(100, 100),
		gridview.Columns(3),
		gridview.SelectionModeOpt(gridview.SelectionSingle),
		gridview.SelectedIndex(50),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 400, MaxWidth: 400,
		MinHeight: 300, MaxHeight: 300,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 400, 300))
	gv.SetFocused(true)

	ke := &event.KeyEvent{Key: event.KeyPageDown, KeyType: event.KeyPress}
	if !gv.Event(ctx, ke) {
		t.Error("PageDown should be consumed")
	}

	ke = &event.KeyEvent{Key: event.KeyPageUp, KeyType: event.KeyPress}
	if !gv.Event(ctx, ke) {
		t.Error("PageUp should be consumed")
	}
}

func TestKeyboard_EnterInvokesClick(t *testing.T) {
	clicked := -1
	gv := gridview.New(
		gridview.ItemCount(10),
		gridview.ItemSize(100, 100),
		gridview.Columns(3),
		gridview.SelectionModeOpt(gridview.SelectionSingle),
		gridview.SelectedIndex(3),
		gridview.OnCellClick(func(index int) { clicked = index }),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 400, MaxWidth: 400,
		MinHeight: 400, MaxHeight: 400,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 400, 400))
	gv.SetFocused(true)

	ke := &event.KeyEvent{Key: event.KeyEnter, KeyType: event.KeyPress}
	consumed := gv.Event(ctx, ke)

	if !consumed {
		t.Error("Enter should be consumed when click callback is set")
	}
	if clicked != 3 {
		t.Errorf("clicked = %d, want 3", clicked)
	}
}

func TestKeyboard_NoSelectionMode(t *testing.T) {
	gv := gridview.New(
		gridview.ItemCount(10),
		gridview.ItemSize(100, 100),
		gridview.Columns(3),
		gridview.SelectionModeOpt(gridview.SelectionNone),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 400, MaxWidth: 400,
		MinHeight: 400, MaxHeight: 400,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 400, 400))
	gv.SetFocused(true)

	ke := &event.KeyEvent{Key: event.KeyRight, KeyType: event.KeyPress}
	if gv.Event(ctx, ke) {
		t.Error("KeyRight should not be consumed when selection mode is None")
	}
}

func TestKeyboard_NotFocused(t *testing.T) {
	gv := gridview.New(
		gridview.ItemCount(10),
		gridview.ItemSize(100, 100),
		gridview.Columns(3),
		gridview.SelectionModeOpt(gridview.SelectionSingle),
		gridview.SelectedIndex(0),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 400, MaxWidth: 400,
		MinHeight: 400, MaxHeight: 400,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 400, 400))
	// Not focused.

	ke := &event.KeyEvent{Key: event.KeyRight, KeyType: event.KeyPress}
	if gv.Event(ctx, ke) {
		t.Error("KeyRight should not be consumed when not focused")
	}
}

func TestKeyboard_EmptyGrid(t *testing.T) {
	gv := gridview.New(
		gridview.ItemCount(0),
		gridview.SelectionModeOpt(gridview.SelectionSingle),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 300, MaxHeight: 300,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 300, 300))
	gv.SetFocused(true)

	ke := &event.KeyEvent{Key: event.KeyDown, KeyType: event.KeyPress}
	if gv.Event(ctx, ke) {
		t.Error("should not consume key events on empty grid")
	}
}

func TestKeyboard_KeyRelease_Ignored(t *testing.T) {
	gv := gridview.New(
		gridview.ItemCount(10),
		gridview.SelectionModeOpt(gridview.SelectionSingle),
		gridview.SelectedIndex(0),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 300, MaxHeight: 300,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 300, 300))
	gv.SetFocused(true)

	ke := &event.KeyEvent{Key: event.KeyDown, KeyType: event.KeyRelease}
	if gv.Event(ctx, ke) {
		t.Error("KeyRelease should not be consumed")
	}
}

func TestKeyboard_UnhandledKey(t *testing.T) {
	gv := gridview.New(
		gridview.ItemCount(10),
		gridview.SelectionModeOpt(gridview.SelectionSingle),
		gridview.SelectedIndex(0),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 300, MaxHeight: 300,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 300, 300))
	gv.SetFocused(true)

	ke := &event.KeyEvent{Key: event.KeyTab, KeyType: event.KeyPress}
	if gv.Event(ctx, ke) {
		t.Error("Tab key should not be consumed by grid")
	}
}

// --- Signal binding tests ---

func TestItemCountSignal(t *testing.T) {
	sig := state.NewSignal(10)
	gv := gridview.New(
		gridview.ItemCountSignal(sig),
		gridview.ItemSize(100, 100),
	)

	if gv.GetItemCount() != 10 {
		t.Errorf("GetItemCount() = %d, want 10", gv.GetItemCount())
	}

	sig.Set(20)
	if gv.GetItemCount() != 20 {
		t.Errorf("GetItemCount() after signal update = %d, want 20", gv.GetItemCount())
	}
}

func TestItemCountFn(t *testing.T) {
	count := 10
	gv := gridview.New(
		gridview.ItemCountFn(func() int { return count }),
	)

	if gv.GetItemCount() != 10 {
		t.Errorf("GetItemCount() = %d, want 10", gv.GetItemCount())
	}

	count = 20
	if gv.GetItemCount() != 20 {
		t.Errorf("GetItemCount() after update = %d, want 20", gv.GetItemCount())
	}
}

func TestSelectedIndexSignal(t *testing.T) {
	sig := state.NewSignal(5)
	gv := gridview.New(
		gridview.ItemCount(20),
		gridview.ItemSize(100, 100),
		gridview.Columns(3),
		gridview.SelectionModeOpt(gridview.SelectionSingle),
		gridview.SelectedIndexSignal(sig),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 400, MaxWidth: 400,
		MinHeight: 500, MaxHeight: 500,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 400, 500))

	// Signal reads work.
	if got := sig.Get(); got != 5 {
		t.Errorf("signal initial = %d, want 5", got)
	}
}

func TestColumnsSignal(t *testing.T) {
	sig := state.NewSignal(3)
	gv := gridview.New(
		gridview.ItemCount(20),
		gridview.ItemSize(100, 100),
		gridview.ColumnsSignal(sig),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 400, MaxWidth: 400,
		MinHeight: 400, MaxHeight: 400,
	})

	if gv.GetColumns() != 3 {
		t.Errorf("GetColumns() = %d, want 3", gv.GetColumns())
	}

	sig.Set(5)
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 600, MaxWidth: 600,
		MinHeight: 400, MaxHeight: 400,
	})

	if gv.GetColumns() != 5 {
		t.Errorf("GetColumns() after signal update = %d, want 5", gv.GetColumns())
	}
}

func TestDisabledSignal(t *testing.T) {
	sig := state.NewSignal(false)
	gv := gridview.New(
		gridview.ItemCount(10),
		gridview.DisabledSignal(sig),
	)

	if !gv.IsFocusable() {
		t.Error("should be focusable when not disabled")
	}

	sig.Set(true)
	if gv.IsFocusable() {
		t.Error("should not be focusable when disabled")
	}
}

func TestDisabledFn(t *testing.T) {
	disabled := false
	gv := gridview.New(
		gridview.ItemCount(10),
		gridview.DisabledFn(func() bool { return disabled }),
	)

	if !gv.IsFocusable() {
		t.Error("should be focusable when not disabled")
	}

	disabled = true
	if gv.IsFocusable() {
		t.Error("should not be focusable when disabled")
	}
}

func TestReadonlySignals(t *testing.T) {
	countSig := state.NewSignal(15)
	selSig := state.NewSignal(3)
	colSig := state.NewSignal(5)
	disSig := state.NewSignal(true)

	gv := gridview.New(
		gridview.ItemCountReadonlySignal(countSig),
		gridview.SelectedIndexReadonlySignal(selSig),
		gridview.ColumnsReadonlySignal(colSig),
		gridview.DisabledReadonlySignal(disSig),
	)

	if gv.GetItemCount() != 15 {
		t.Errorf("GetItemCount() = %d, want 15", gv.GetItemCount())
	}
	if gv.IsFocusable() {
		t.Error("should not be focusable when disabled via readonly signal")
	}
}

// --- Children tests ---

func TestChildren(t *testing.T) {
	gv := gridview.New(gridview.ItemCount(5))
	children := gv.Children()
	if len(children) != 1 {
		t.Errorf("Children() len = %d, want 1 (scroll view)", len(children))
	}
}

func TestChildren_NilScroll(t *testing.T) {
	// This tests the nil guard, though New() always creates scroll.
	gv := &gridview.Widget{}
	children := gv.Children()
	if children != nil {
		t.Errorf("Children() = %v, want nil for zero-value Widget", children)
	}
}

// --- Accessibility tests ---

func TestAccessibility_Role(t *testing.T) {
	gv := gridview.New(gridview.ItemCount(10))
	if gv.AccessibilityRole() != a11y.RoleGrid {
		t.Errorf("AccessibilityRole() = %v, want RoleGrid", gv.AccessibilityRole())
	}
}

func TestAccessibility_Label(t *testing.T) {
	gv := gridview.New()
	if gv.AccessibilityLabel() != "Grid" {
		t.Errorf("AccessibilityLabel() = %q, want %q", gv.AccessibilityLabel(), "Grid")
	}

	gv2 := gridview.New(gridview.A11yLabel("Image Gallery"))
	if gv2.AccessibilityLabel() != "Image Gallery" {
		t.Errorf("AccessibilityLabel() = %q, want %q", gv2.AccessibilityLabel(), "Image Gallery")
	}
}

func TestAccessibility_Value(t *testing.T) {
	gv := gridview.New(gridview.ItemCount(10))
	want := "10 items"
	if got := gv.AccessibilityValue(); got != want {
		t.Errorf("AccessibilityValue() = %q, want %q", got, want)
	}

	gv2 := gridview.New(
		gridview.ItemCount(10),
		gridview.SelectedIndex(3),
	)
	want2 := "Item 4 of 10 selected"
	if got := gv2.AccessibilityValue(); got != want2 {
		t.Errorf("AccessibilityValue() = %q, want %q", got, want2)
	}
}

func TestAccessibility_Hint(t *testing.T) {
	gv := gridview.New()
	if gv.AccessibilityHint() != "" {
		t.Errorf("AccessibilityHint() = %q, want empty", gv.AccessibilityHint())
	}
}

func TestAccessibility_State(t *testing.T) {
	gv := gridview.New(gridview.Disabled(true))
	st := gv.AccessibilityState()
	if !st.Disabled {
		t.Error("AccessibilityState().Disabled should be true")
	}
}

func TestAccessibility_Actions(t *testing.T) {
	gv := gridview.New()
	actions := gv.AccessibilityActions()
	if len(actions) != 2 {
		t.Errorf("AccessibilityActions() len = %d, want 2", len(actions))
	}
}

// --- IsFocusable tests ---

func TestIsFocusable(t *testing.T) {
	gv := gridview.New(gridview.ItemCount(10))

	if !gv.IsFocusable() {
		t.Error("should be focusable by default")
	}

	gv.SetVisible(false)
	if gv.IsFocusable() {
		t.Error("should not be focusable when invisible")
	}

	gv.SetVisible(true)
	gv.SetEnabled(false)
	if gv.IsFocusable() {
		t.Error("should not be focusable when not enabled")
	}
}

// --- InvalidateData test ---

func TestInvalidateData(t *testing.T) {
	gv := gridview.New(
		gridview.ItemCount(10),
		gridview.ItemSize(100, 100),
		gridview.Columns(3),
	)

	// Should not panic.
	gv.InvalidateData()
}

// --- ScrollToIndex test ---

func TestScrollToIndex(t *testing.T) {
	gv := gridview.New(
		gridview.ItemCount(100),
		gridview.ItemSize(100, 100),
		gridview.Columns(3),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 400, MaxWidth: 400,
		MinHeight: 200, MaxHeight: 200,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 400, 200))

	// Should not panic for various indices.
	gv.ScrollToIndex(0)
	gv.ScrollToIndex(50)
	gv.ScrollToIndex(99)
	gv.ScrollToIndex(-1)  // out of range
	gv.ScrollToIndex(100) // out of range
}

// --- Mount/Unmount tests ---

func TestMountUnmount(t *testing.T) {
	gv := gridview.New(
		gridview.ItemCount(10),
		gridview.ItemSize(100, 100),
		gridview.Columns(3),
	)

	ctx := widget.NewContext()
	// Should not panic.
	gv.Mount(ctx)
	gv.Unmount()
}

func TestMountWithSignals(t *testing.T) {
	countSig := state.NewSignal(10)
	selSig := state.NewSignal(0)
	disSig := state.NewSignal(false)
	colSig := state.NewSignal(3)

	gv := gridview.New(
		gridview.ItemCountSignal(countSig),
		gridview.SelectedIndexSignal(selSig),
		gridview.DisabledSignal(disSig),
		gridview.ColumnsSignal(colSig),
	)

	ctx := widget.NewContext()
	// Should not panic (scheduler may be nil in test context).
	gv.Mount(ctx)
	gv.Unmount()
}

// --- CellContent option test ---

func TestCellContent(t *testing.T) {
	called := false
	gv := gridview.New(
		gridview.ItemCount(4),
		gridview.ItemSize(100, 100),
		gridview.Columns(2),
		gridview.CellContent(testCellContent{fn: func(ctx gridview.CellContext) widget.Widget {
			called = true
			return nil
		}}),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 300, MaxHeight: 300,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 300, 300))
	gv.Draw(ctx, &mockCanvas{})

	if !called {
		t.Error("CellContent.Render was not called")
	}
}

// --- OnSelectionChange test ---

func TestOnSelectionChange(t *testing.T) {
	changed := -1
	gv := gridview.New(
		gridview.ItemCount(10),
		gridview.ItemSize(100, 100),
		gridview.Columns(3),
		gridview.SelectionModeOpt(gridview.SelectionSingle),
		gridview.SelectedIndex(0),
		gridview.OnSelectionChange(func(index int) { changed = index }),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 400, MaxWidth: 400,
		MinHeight: 400, MaxHeight: 400,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 400, 400))
	gv.SetFocused(true)

	// Move right: 0 -> 1
	ke := &event.KeyEvent{Key: event.KeyRight, KeyType: event.KeyPress}
	gv.Event(ctx, ke)

	if changed != 1 {
		t.Errorf("OnSelectionChange received %d, want 1", changed)
	}
}

// --- OnScroll option test ---

func TestOnScrollOption(t *testing.T) {
	gv := gridview.New(
		gridview.ItemCount(10),
		gridview.OnScroll(func(offset float32) {
			// callback set -- should not panic
		}),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 300, MaxHeight: 300,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 300, 300))
	gv.Draw(ctx, &mockCanvas{})
}

// --- ScrollYSignal option test ---

func TestScrollYSignal(t *testing.T) {
	sig := state.NewSignal[float32](0)
	gv := gridview.New(
		gridview.ItemCount(100),
		gridview.ItemSize(100, 100),
		gridview.ScrollYSignal(sig),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 300, MaxHeight: 300,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 300, 300))
	gv.Draw(ctx, &mockCanvas{})
}

// --- ColumnsAuto with tiny viewport ---

func TestColumnsAuto_TinyViewport(t *testing.T) {
	gv := gridview.New(
		gridview.ItemCount(20),
		gridview.ItemSize(200, 200),
		gridview.ColumnsAuto(true),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 50, MaxWidth: 50,
		MinHeight: 200, MaxHeight: 200,
	})

	// With viewport 50px and item 200px, should clamp to 1 column.
	if gv.GetColumns() < 1 {
		t.Errorf("GetColumns() = %d, want >= 1", gv.GetColumns())
	}
}

// --- DefaultPainter tests ---

func TestDefaultPainter_PaintEmptyState(t *testing.T) {
	p := gridview.DefaultPainter{}
	c := &mockCanvas{}
	// Should not panic.
	p.PaintEmptyState(c, geometry.NewRect(0, 0, 300, 300))
	p.PaintEmptyState(c, geometry.Rect{}) // empty bounds
}

func TestDefaultPainter_PaintCellBackground(t *testing.T) {
	p := gridview.DefaultPainter{}
	c := &mockCanvas{}

	// Hovered cell.
	p.PaintCellBackground(c, gridview.CellPaintState{
		Bounds:  geometry.NewRect(0, 0, 100, 100),
		Hovered: true,
	})

	// Empty bounds.
	p.PaintCellBackground(c, gridview.CellPaintState{})

	// Disabled + hovered: should not draw.
	p.PaintCellBackground(c, gridview.CellPaintState{
		Bounds:   geometry.NewRect(0, 0, 100, 100),
		Hovered:  true,
		Disabled: true,
	})

	// With color scheme.
	p.PaintCellBackground(c, gridview.CellPaintState{
		Bounds:  geometry.NewRect(0, 0, 100, 100),
		Hovered: true,
		ColorScheme: gridview.GridColorScheme{
			HoverColor: widget.RGBA(1, 0, 0, 0.1),
		},
	})
}

func TestDefaultPainter_PaintSelection(t *testing.T) {
	p := gridview.DefaultPainter{}
	c := &mockCanvas{}

	// Selected cell.
	p.PaintSelection(c, gridview.CellPaintState{
		Bounds:   geometry.NewRect(0, 0, 100, 100),
		Selected: true,
	})

	// Selected + focused.
	p.PaintSelection(c, gridview.CellPaintState{
		Bounds:   geometry.NewRect(0, 0, 100, 100),
		Selected: true,
		Focused:  true,
	})

	// Not selected.
	p.PaintSelection(c, gridview.CellPaintState{
		Bounds: geometry.NewRect(0, 0, 100, 100),
	})

	// Empty bounds.
	p.PaintSelection(c, gridview.CellPaintState{Selected: true})

	// With color scheme.
	p.PaintSelection(c, gridview.CellPaintState{
		Bounds:   geometry.NewRect(0, 0, 100, 100),
		Selected: true,
		Focused:  true,
		ColorScheme: gridview.GridColorScheme{
			SelectionColor: widget.RGBA(0, 0, 1, 0.2),
			FocusColor:     widget.RGBA(0, 1, 0, 0.5),
		},
	})

	// Focused + disabled: should not draw focus border.
	p.PaintSelection(c, gridview.CellPaintState{
		Bounds:   geometry.NewRect(0, 0, 100, 100),
		Selected: true,
		Focused:  true,
		Disabled: true,
	})
}

// --- BuildCell context propagation ---

func TestBuildCell_ContextPropagation(t *testing.T) {
	var received []gridview.CellContext
	gv := gridview.New(
		gridview.ItemCount(6),
		gridview.ItemSize(100, 100),
		gridview.Columns(3),
		gridview.SelectionModeOpt(gridview.SelectionSingle),
		gridview.SelectedIndex(2),
		gridview.BuildCell(func(index int, ctx gridview.CellContext) widget.Widget {
			received = append(received, ctx)
			return nil
		}),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 400, MaxWidth: 400,
		MinHeight: 300, MaxHeight: 300,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 400, 300))
	gv.Draw(ctx, &mockCanvas{})

	if len(received) != 6 {
		t.Fatalf("received %d contexts, want 6", len(received))
	}

	// Check index 0: row=0, col=0, not selected.
	if received[0].Index != 0 || received[0].Row != 0 || received[0].Col != 0 || received[0].IsSelected {
		t.Errorf("ctx[0] = %+v, want Index=0, Row=0, Col=0, IsSelected=false", received[0])
	}

	// Check index 2: row=0, col=2, selected.
	if received[2].Index != 2 || received[2].Row != 0 || received[2].Col != 2 || !received[2].IsSelected {
		t.Errorf("ctx[2] = %+v, want Index=2, Row=0, Col=2, IsSelected=true", received[2])
	}

	// Check index 3: row=1, col=0.
	if received[3].Index != 3 || received[3].Row != 1 || received[3].Col != 0 {
		t.Errorf("ctx[3] = %+v, want Index=3, Row=1, Col=0", received[3])
	}
}

// --- Enter/Space with no callback ---

func TestKeyboard_EnterNoCallback(t *testing.T) {
	gv := gridview.New(
		gridview.ItemCount(10),
		gridview.SelectionModeOpt(gridview.SelectionSingle),
		gridview.SelectedIndex(0),
		// No OnCellClick.
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 300, MaxHeight: 300,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 300, 300))
	gv.SetFocused(true)

	ke := &event.KeyEvent{Key: event.KeyEnter, KeyType: event.KeyPress}
	if gv.Event(ctx, ke) {
		t.Error("Enter without click callback should return false")
	}
}

// --- Space key invokes click ---

func TestKeyboard_SpaceInvokesClick(t *testing.T) {
	clicked := -1
	gv := gridview.New(
		gridview.ItemCount(10),
		gridview.SelectionModeOpt(gridview.SelectionSingle),
		gridview.SelectedIndex(2),
		gridview.OnCellClick(func(index int) { clicked = index }),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 300, MaxHeight: 300,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 300, 300))
	gv.SetFocused(true)

	ke := &event.KeyEvent{Key: event.KeySpace, KeyType: event.KeyPress}
	consumed := gv.Event(ctx, ke)

	if !consumed {
		t.Error("Space should be consumed when click callback is set")
	}
	if clicked != 2 {
		t.Errorf("clicked = %d, want 2", clicked)
	}
}

// --- Test helpers ---

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

type testPainter struct {
	cellBackgroundCalls int
	selectionCalls      int
	emptyStateCalls     int
}

func (p *testPainter) PaintCellBackground(_ widget.Canvas, _ gridview.CellPaintState) {
	p.cellBackgroundCalls++
}
func (p *testPainter) PaintSelection(_ widget.Canvas, _ gridview.CellPaintState) {
	p.selectionCalls++
}
func (p *testPainter) PaintEmptyState(_ widget.Canvas, _ geometry.Rect) {
	p.emptyStateCalls++
}

type testCellContent struct {
	fn func(ctx gridview.CellContext) widget.Widget
}

func (t testCellContent) Render(ctx gridview.CellContext) widget.Widget {
	if t.fn != nil {
		return t.fn(ctx)
	}
	return nil
}
