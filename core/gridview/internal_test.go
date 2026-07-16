package gridview

import (
	"github.com/gogpu/gg/scene"
	"image"
	"testing"

	"github.com/gogpu/ui/cdk"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// --- cellCache tests ---

func TestCellCache_Update(t *testing.T) {
	var cc cellCache
	builder := cdk.FuncContent[CellContext]{Fn: func(ctx CellContext) widget.Widget {
		return nil
	}}

	cc.update(0, 5, builder, -1, -1, 3, nil)

	if !cc.valid {
		t.Error("cache should be valid after update")
	}
	if cc.startIndex != 0 {
		t.Errorf("startIndex = %d, want 0", cc.startIndex)
	}
	if cc.endIndex != 5 {
		t.Errorf("endIndex = %d, want 5", cc.endIndex)
	}
	if len(cc.widgets) != 5 {
		t.Errorf("len(widgets) = %d, want 5", len(cc.widgets))
	}
}

func TestCellCache_Reuse(t *testing.T) {
	var cc cellCache
	callCount := 0
	builder := cdk.FuncContent[CellContext]{Fn: func(ctx CellContext) widget.Widget {
		callCount++
		return nil
	}}

	cc.update(0, 5, builder, -1, -1, 3, nil)
	if callCount != 5 {
		t.Errorf("callCount = %d, want 5", callCount)
	}

	// Second call with same range should be no-op.
	cc.update(0, 5, builder, -1, -1, 3, nil)
	if callCount != 5 {
		t.Errorf("callCount after reuse = %d, want 5", callCount)
	}
}

func TestCellCache_Invalidate_Forces_Rebuild(t *testing.T) {
	var cc cellCache
	callCount := 0
	builder := cdk.FuncContent[CellContext]{Fn: func(ctx CellContext) widget.Widget {
		callCount++
		return nil
	}}

	cc.update(0, 5, builder, -1, -1, 3, nil)
	cc.invalidate()
	cc.update(0, 5, builder, -1, -1, 3, nil)

	if callCount != 10 {
		t.Errorf("callCount = %d, want 10 (5+5 after invalidate)", callCount)
	}
}

func TestCellCache_Clear(t *testing.T) {
	var cc cellCache
	builder := cdk.FuncContent[CellContext]{Fn: func(ctx CellContext) widget.Widget {
		return nil
	}}

	cc.update(0, 5, builder, -1, -1, 3, nil)
	cc.clear()

	if cc.valid {
		t.Error("cache should be invalid after clear")
	}
	if len(cc.widgets) != 0 {
		t.Errorf("len(widgets) = %d, want 0", len(cc.widgets))
	}
}

func TestCellCache_EmptyRange(t *testing.T) {
	var cc cellCache
	builder := cdk.FuncContent[CellContext]{Fn: func(ctx CellContext) widget.Widget {
		return nil
	}}

	cc.update(0, 0, builder, -1, -1, 3, nil)

	if cc.valid {
		t.Error("cache should not be valid for empty range")
	}
}

func TestCellCache_WidgetAt(t *testing.T) {
	var cc cellCache

	// Empty cache.
	if got := cc.widgetAt(0); got != nil {
		t.Error("widgetAt on empty cache should return nil")
	}
	if got := cc.widgetAt(-1); got != nil {
		t.Error("widgetAt(-1) should return nil")
	}
}

func TestCellCache_NilBuilder(t *testing.T) {
	var cc cellCache
	cc.update(0, 3, nil, -1, -1, 3, nil)

	for i := 0; i < 3; i++ {
		if got := cc.widgetAt(i); got != nil {
			t.Errorf("widgetAt(%d) with nil builder should return nil", i)
		}
	}
}

func TestCellCache_ContextPropagation(t *testing.T) {
	var cc cellCache
	var received []CellContext
	builder := cdk.FuncContent[CellContext]{Fn: func(ctx CellContext) widget.Widget {
		received = append(received, ctx)
		return nil
	}}

	// 6 items, 3 cols: row 0 = [0,1,2], row 1 = [3,4,5]
	cc.update(0, 6, builder, 2, 4, 3, nil)

	if len(received) != 6 {
		t.Fatalf("received %d contexts, want 6", len(received))
	}

	// Index 0: row=0, col=0, not selected, not hovered.
	if received[0].Index != 0 || received[0].Row != 0 || received[0].Col != 0 ||
		received[0].IsSelected || received[0].IsHovered {
		t.Errorf("ctx[0] = %+v, want Index=0, Row=0, Col=0", received[0])
	}
	// Index 2: selected.
	if received[2].Index != 2 || !received[2].IsSelected {
		t.Errorf("ctx[2] = %+v, want Index=2, IsSelected=true", received[2])
	}
	// Index 4: hovered.
	if received[4].Index != 4 || !received[4].IsHovered {
		t.Errorf("ctx[4] = %+v, want Index=4, IsHovered=true", received[4])
	}
	// Index 3: row=1, col=0.
	if received[3].Row != 1 || received[3].Col != 0 {
		t.Errorf("ctx[3] = %+v, want Row=1, Col=0", received[3])
	}
}

func TestCellCache_ZeroCols(t *testing.T) {
	var cc cellCache
	var received []CellContext
	builder := cdk.FuncContent[CellContext]{Fn: func(ctx CellContext) widget.Widget {
		received = append(received, ctx)
		return nil
	}}

	// Zero cols should not panic.
	cc.update(0, 2, builder, -1, -1, 0, nil)

	if len(received) != 2 {
		t.Fatalf("received %d contexts, want 2", len(received))
	}
	// With zero cols, row/col should both be 0.
	if received[0].Row != 0 || received[0].Col != 0 {
		t.Errorf("ctx[0] with zero cols = %+v", received[0])
	}
}

// --- virtualContent tests ---

func TestVirtualContent_Children(t *testing.T) {
	vc := &virtualContent{}
	if got := vc.Children(); got != nil {
		t.Errorf("Children() = %v, want nil", got)
	}
}

func TestVirtualContent_Layout_NilGrid(t *testing.T) {
	vc := &virtualContent{}
	got := vc.Layout(nil, geometry.Constraints{MinWidth: 100, MaxWidth: 300, MinHeight: 100, MaxHeight: 500})
	if got != (geometry.Size{}) {
		t.Errorf("Layout with nil grid = %v, want zero", got)
	}
}

func TestVirtualContent_Draw_NilGrid(t *testing.T) {
	vc := &virtualContent{}
	// Should not panic.
	vc.Draw(nil, &mockCanvas{})
}

func TestVirtualContent_Event_NilGrid(t *testing.T) {
	vc := &virtualContent{}
	if got := vc.Event(nil, &event.MouseEvent{}); got {
		t.Error("Event with nil grid should return false")
	}
}

func TestVirtualContent_Layout_InfiniteWidth(t *testing.T) {
	gv := New(
		ItemCount(10),
		ItemSize(100, 100),
		Columns(3),
	)

	vc := &virtualContent{grid: gv}
	got := vc.Layout(nil, geometry.Constraints{
		MinWidth:  100,
		MaxWidth:  geometry.Infinity,
		MinHeight: 0,
		MaxHeight: geometry.Infinity,
	})

	// Should use MinWidth when MaxWidth is infinity.
	if got.Width != 100 {
		t.Errorf("Width = %v, want 100", got.Width)
	}
}

// --- cellIndexAtPoint tests ---

func TestCellIndexAtPoint(t *testing.T) {
	gv := New(
		ItemCount(12),
		ItemSize(100, 100),
		Columns(3),
		Gap(10),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 400, MaxWidth: 400,
		MinHeight: 400, MaxHeight: 400,
	})

	tests := []struct {
		name string
		x, y float32
		want int
	}{
		{"cell 0,0", 50, 50, 0},
		{"cell 1,0", 160, 50, 1},          // 110+50 = in second column
		{"cell 2,0", 270, 50, 2},          // 220+50
		{"cell 0,1", 50, 160, 3},          // row 1
		{"in gap x", 105, 50, -1},         // between col 0 and 1 (gap at 100-110)
		{"in gap y", 50, 105, -1},         // between row 0 and 1
		{"col out of range", 400, 50, -1}, // beyond columns
		{"beyond items", 50, 500, -1},     // beyond item count
		{"negative x", -10, 50, -1},       // negative coordinate
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := gv.cellIndexAtPoint(tt.x, tt.y)
			if got != tt.want {
				t.Errorf("cellIndexAtPoint(%v, %v) = %d, want %d", tt.x, tt.y, got, tt.want)
			}
		})
	}
}

func TestCellIndexAtPoint_ZeroColumns(t *testing.T) {
	gv := &Widget{}
	// Zero effective columns.
	if got := gv.cellIndexAtPoint(50, 50); got != noHoveredIndex {
		t.Errorf("cellIndexAtPoint with zero cols = %d, want %d", got, noHoveredIndex)
	}
}

func TestCellIndexAtPoint_ZeroItemSize(t *testing.T) {
	gv := New(
		ItemCount(10),
		ItemSize(0, 0),
		Columns(3),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 300, MaxHeight: 300,
	})

	if got := gv.cellIndexAtPoint(50, 50); got != noHoveredIndex {
		t.Errorf("cellIndexAtPoint with zero item size = %d, want %d", got, noHoveredIndex)
	}
}

// --- ceilDiv tests ---

func TestCeilDiv(t *testing.T) {
	tests := []struct {
		a, b int
		want int
	}{
		{10, 3, 4},
		{9, 3, 3},
		{0, 3, 0},
		{1, 3, 1},
		{10, 0, 0}, // zero divisor
		{7, 4, 2},
	}
	for _, tt := range tests {
		got := ceilDiv(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("ceilDiv(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

// --- visibleRowRange tests ---

func TestVisibleRowRange(t *testing.T) {
	gv := New(
		ItemCount(30),
		ItemSize(100, 100),
		Columns(3),
		Gap(10),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 400, MaxWidth: 400,
		MinHeight: 250, MaxHeight: 250,
	})

	tests := []struct {
		name                string
		scrollY, viewportH  float32
		wantFirst, wantLast int
	}{
		{"top", 0, 250, 0, 2},
		{"scrolled down", 220, 250, 2, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			first, last := gv.visibleRowRange(tt.scrollY, tt.viewportH)
			if first != tt.wantFirst || last != tt.wantLast {
				t.Errorf("visibleRowRange(%v, %v) = (%d, %d), want (%d, %d)",
					tt.scrollY, tt.viewportH, first, last, tt.wantFirst, tt.wantLast)
			}
		})
	}
}

func TestVisibleRowRange_EmptyGrid(t *testing.T) {
	gv := New(ItemCount(0))
	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 300, MaxHeight: 300,
	})

	first, last := gv.visibleRowRange(0, 300)
	if first != 0 || last != 0 {
		t.Errorf("visibleRowRange empty = (%d, %d), want (0, 0)", first, last)
	}
}

func TestVisibleRowRange_ZeroCellStep(t *testing.T) {
	gv := New(
		ItemCount(10),
		ItemSize(0, 0),
		Gap(0),
	)
	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 300, MaxHeight: 300,
	})

	first, last := gv.visibleRowRange(0, 300)
	if first != 0 || last != 0 {
		t.Errorf("visibleRowRange zero step = (%d, %d), want (0, 0)", first, last)
	}
}

// --- totalContentHeight tests ---

func TestTotalContentHeight(t *testing.T) {
	gv := New(
		ItemCount(12),
		ItemSize(100, 100),
		Columns(3),
		Gap(10),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 400, MaxWidth: 400,
		MinHeight: 400, MaxHeight: 400,
	})

	// 12 items / 3 cols = 4 rows
	// 4 * 100 + 3 * 10 = 430
	got := gv.totalContentHeight()
	if got != 430 {
		t.Errorf("totalContentHeight() = %v, want 430", got)
	}
}

func TestTotalContentHeight_Empty(t *testing.T) {
	gv := New(ItemCount(0))
	if got := gv.totalContentHeight(); got != 0 {
		t.Errorf("totalContentHeight() empty = %v, want 0", got)
	}
}

// --- computeEffectiveColumns tests ---

func TestComputeEffectiveColumns_MinOne(t *testing.T) {
	gv := New(
		ItemCount(10),
		ItemSize(100, 100),
		Columns(0), // zero columns
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 300, MaxHeight: 300,
	})

	if gv.GetColumns() < 1 {
		t.Errorf("GetColumns() with zero config = %d, want >= 1", gv.GetColumns())
	}
}

func TestComputeEffectiveColumns_AutoZeroStep(t *testing.T) {
	gv := New(
		ItemCount(10),
		ItemSize(0, 100),
		Gap(0),
		ColumnsAuto(true),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 300, MaxHeight: 300,
	})

	// Zero item width + zero gap = zero step => should clamp to 1.
	if gv.GetColumns() < 1 {
		t.Errorf("GetColumns() with zero step = %d, want >= 1", gv.GetColumns())
	}
}

// --- handleContentMouseEvent tests ---

func TestHandleContentMouseEvent_Move(t *testing.T) {
	gv := New(
		ItemCount(12),
		ItemSize(100, 100),
		Columns(3),
		Gap(0),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 400, MaxWidth: 400,
		MinHeight: 400, MaxHeight: 400,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 400, 400))

	// Mouse move over cell 0.
	me := &event.MouseEvent{
		MouseType: event.MouseMove,
		Position:  geometry.Pt(50, 50),
	}
	consumed := handleContentEvent(gv, ctx, me)
	// Move should not consume.
	if consumed {
		t.Error("MouseMove should not be consumed")
	}
	// Hovered index should update.
	if gv.hoveredIndex != 0 {
		t.Errorf("hoveredIndex = %d, want 0", gv.hoveredIndex)
	}

	// Move to cell 4 (row 1, col 1).
	me.Position = geometry.Pt(150, 150)
	handleContentEvent(gv, ctx, me)
	if gv.hoveredIndex != 4 {
		t.Errorf("hoveredIndex = %d, want 4", gv.hoveredIndex)
	}
}

func TestHandleContentMouseEvent_Press(t *testing.T) {
	clicked := -1
	gv := New(
		ItemCount(12),
		ItemSize(100, 100),
		Columns(3),
		Gap(0),
		SelectionModeOpt(SelectionSingle),
		SelectedIndex(-1),
		OnCellClick(func(index int) { clicked = index }),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 400, MaxWidth: 400,
		MinHeight: 400, MaxHeight: 400,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 400, 400))

	// Click cell 5 (row 1, col 2).
	me := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(250, 150),
	}
	consumed := handleContentEvent(gv, ctx, me)

	if !consumed {
		t.Error("MousePress on cell should be consumed")
	}
	if clicked != 5 {
		t.Errorf("clicked = %d, want 5", clicked)
	}
	if gv.cfg.ResolvedSelectedIndex() != 5 {
		t.Errorf("selectedIndex = %d, want 5", gv.cfg.ResolvedSelectedIndex())
	}
}

func TestHandleContentMouseEvent_Leave(t *testing.T) {
	gv := New(
		ItemCount(12),
		ItemSize(100, 100),
		Columns(3),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 400, MaxWidth: 400,
		MinHeight: 400, MaxHeight: 400,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 400, 400))

	// Hover first.
	gv.hoveredIndex = 5

	// Mouse leave.
	me := &event.MouseEvent{MouseType: event.MouseLeave}
	handleContentEvent(gv, ctx, me)

	if gv.hoveredIndex != noHoveredIndex {
		t.Errorf("hoveredIndex after leave = %d, want %d", gv.hoveredIndex, noHoveredIndex)
	}
}

func TestHandleContentMouseEvent_Leave_NoHover(t *testing.T) {
	gv := New(
		ItemCount(12),
		ItemSize(100, 100),
		Columns(3),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 400, MaxWidth: 400,
		MinHeight: 400, MaxHeight: 400,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 400, 400))

	// Leave with no hover active -- should not invalidate.
	me := &event.MouseEvent{MouseType: event.MouseLeave}
	handleContentEvent(gv, ctx, me)
	// Should not panic.
}

func TestHandleContentMouseEvent_Disabled(t *testing.T) {
	gv := New(
		ItemCount(12),
		ItemSize(100, 100),
		Columns(3),
		Disabled(true),
	)

	ctx := widget.NewContext()
	me := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(50, 50),
	}
	if handleContentEvent(gv, ctx, me) {
		t.Error("disabled grid should not handle mouse events")
	}
}

func TestHandleContentMouseEvent_RightClick(t *testing.T) {
	gv := New(
		ItemCount(12),
		ItemSize(100, 100),
		Columns(3),
		SelectionModeOpt(SelectionSingle),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 400, MaxWidth: 400,
		MinHeight: 400, MaxHeight: 400,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 400, 400))

	me := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonRight,
		Position:  geometry.Pt(50, 50),
	}
	if handleContentEvent(gv, ctx, me) {
		t.Error("right click should not be consumed")
	}
}

func TestHandleContentMouseEvent_ClickOutside(t *testing.T) {
	gv := New(
		ItemCount(4),
		ItemSize(100, 100),
		Columns(2),
		SelectionModeOpt(SelectionSingle),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 300, MaxHeight: 300,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 300, 300))

	// Click beyond items.
	me := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(250, 250), // beyond 4 items in 2x2 grid
	}
	if handleContentEvent(gv, ctx, me) {
		t.Error("click outside cells should not be consumed")
	}
}

func TestHandleContentEvent_NonMouse(t *testing.T) {
	gv := New(ItemCount(10))
	ctx := widget.NewContext()

	ke := &event.KeyEvent{Key: event.KeyDown}
	if handleContentEvent(gv, ctx, ke) {
		t.Error("non-mouse event should not be consumed")
	}
}

func TestHandleContentMouseEvent_UnhandledType(t *testing.T) {
	gv := New(
		ItemCount(10),
		ItemSize(100, 100),
		Columns(3),
	)

	ctx := widget.NewContext()
	me := &event.MouseEvent{
		MouseType: event.MouseRelease,
		Position:  geometry.Pt(50, 50),
	}
	if handleContentEvent(gv, ctx, me) {
		t.Error("MouseRelease should not be consumed")
	}
}

// --- Layout edge cases ---

func TestLayout_ZeroWidthHeight(t *testing.T) {
	gv := New(
		ItemCount(10),
		ItemSize(100, 100),
		Columns(3),
	)

	ctx := widget.NewContext()
	size := gv.Layout(ctx, geometry.Constraints{
		MinWidth: 0, MaxWidth: 0,
		MinHeight: 0, MaxHeight: 0,
	})

	// Should use default fallbacks.
	if size.Width <= 0 || size.Height <= 0 {
		t.Errorf("Layout returned non-positive size: %v", size)
	}
}

// --- setScrollY with signal ---

func TestSetScrollY_WithSignal(t *testing.T) {
	gv := New(
		ItemCount(100),
		ItemSize(100, 100),
		Columns(3),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 400, MaxWidth: 400,
		MinHeight: 200, MaxHeight: 200,
	})

	// Without signal, setScrollY should not panic.
	gv.setScrollY(100)
}

// --- setSelectedIndex with signal ---

func TestSetSelectedIndex_SameValue(t *testing.T) {
	gv := New(
		ItemCount(10),
		SelectionModeOpt(SelectionSingle),
		SelectedIndex(5),
	)

	ctx := widget.NewContext()
	// Setting same value should be no-op.
	gv.setSelectedIndex(ctx, 5)
}

// --- Mount with scheduler tests ---

type mockScheduler struct {
	dirtyCount int
}

func (s *mockScheduler) MarkDirty(_ widget.Widget) {
	s.dirtyCount++
}

func TestMount_WithScheduler_AllSignals(t *testing.T) {
	countSig := state.NewSignal(10)
	selSig := state.NewSignal(0)
	disSig := state.NewSignal(false)
	colSig := state.NewSignal(3)

	gv := New(
		ItemCountSignal(countSig),
		SelectedIndexSignal(selSig),
		DisabledSignal(disSig),
		ColumnsSignal(colSig),
	)

	sched := &mockScheduler{}
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	gv.Mount(ctx)

	// Verify bindings were created by changing signals.
	countSig.Set(20)
	selSig.Set(5)
	disSig.Set(true)
	colSig.Set(5)

	if sched.dirtyCount != 4 {
		t.Errorf("dirtyCount = %d, want 4 (one per signal change)", sched.dirtyCount)
	}

	gv.Unmount()
}

func TestMount_WithScheduler_ReadonlySignals(t *testing.T) {
	countSig := state.NewSignal(10)
	selSig := state.NewSignal(0)
	disSig := state.NewSignal(false)
	colSig := state.NewSignal(3)

	gv := New(
		ItemCountReadonlySignal(countSig),
		SelectedIndexReadonlySignal(selSig),
		DisabledReadonlySignal(disSig),
		ColumnsReadonlySignal(colSig),
	)

	sched := &mockScheduler{}
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	gv.Mount(ctx)

	countSig.Set(20)
	selSig.Set(5)
	disSig.Set(true)
	colSig.Set(5)

	if sched.dirtyCount != 4 {
		t.Errorf("dirtyCount = %d, want 4", sched.dirtyCount)
	}

	gv.Unmount()
}

func TestMount_ReadonlyTakesPrecedence(t *testing.T) {
	roSig := state.NewSignal(10)
	rwSig := state.NewSignal(5)

	gv := New(
		ItemCountReadonlySignal(roSig),
		ItemCountSignal(rwSig),
	)

	sched := &mockScheduler{}
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	gv.Mount(ctx)

	// Only readonly should be bound (higher priority).
	roSig.Set(20)
	rwSig.Set(15)

	// Only 1 dirty mark (from readonly), not 2.
	if sched.dirtyCount != 1 {
		t.Errorf("dirtyCount = %d, want 1 (only readonly bound)", sched.dirtyCount)
	}

	gv.Unmount()
}

// --- ResolvedSelectedIndex coverage ---

func TestResolvedSelectedIndex_Signal(t *testing.T) {
	sig := state.NewSignal(7)
	cfg := config{selectedIndexSignal: sig, selectedIndex: 3}

	if got := cfg.ResolvedSelectedIndex(); got != 7 {
		t.Errorf("ResolvedSelectedIndex() = %d, want 7", got)
	}
}

func TestResolvedSelectedIndex_Readonly(t *testing.T) {
	sig := state.NewSignal(9)
	cfg := config{readonlySelectedIndexSignal: sig, selectedIndex: 3}

	if got := cfg.ResolvedSelectedIndex(); got != 9 {
		t.Errorf("ResolvedSelectedIndex() = %d, want 9", got)
	}
}

// --- ResolvedColumns coverage ---

func TestResolvedColumns_Signal(t *testing.T) {
	sig := state.NewSignal(6)
	cfg := config{columnsSignal: sig, columns: 3}

	if got := cfg.ResolvedColumns(); got != 6 {
		t.Errorf("ResolvedColumns() = %d, want 6", got)
	}
}

// --- VisibleRange edge cases ---

func TestVisibleRange_ZeroColumns(t *testing.T) {
	gv := &Widget{}
	start, end := gv.VisibleRange()
	if start != 0 || end != 0 {
		t.Errorf("VisibleRange() zero cols = (%d, %d), want (0, 0)", start, end)
	}
}

// --- moveSelectionByPage edge cases ---

func TestMoveSelectionByPage_ZeroCellStep(t *testing.T) {
	gv := New(
		ItemCount(10),
		ItemSize(0, 0),
		Gap(0),
		Columns(3),
		SelectionModeOpt(SelectionSingle),
		SelectedIndex(0),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 300, MaxHeight: 300,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 300, 300))
	gv.SetFocused(true)

	ke := &event.KeyEvent{Key: event.KeyPageDown, KeyType: event.KeyPress}
	// Should not panic despite zero cell step.
	gv.Event(ctx, ke)
}

// --- drawVisibleCells with zero cols ---

func TestDrawVisibleCells_ZeroCols(t *testing.T) {
	gv := New(
		ItemCount(5),
		ItemSize(0, 0),
		Columns(0),
	)

	ctx := widget.NewContext()
	gv.Layout(ctx, geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 300, MaxHeight: 300,
	})
	gv.SetBounds(geometry.NewRect(0, 0, 300, 300))
	// Should not panic.
	gv.Draw(ctx, &mockCanvas{})
}

// --- ScrollToIndex edge cases ---

func TestScrollToIndex_ZeroCols(t *testing.T) {
	gv := &Widget{}
	// Should not panic with zero effective columns.
	gv.ScrollToIndex(0)
}

// --- mockCanvas for internal tests ---

type mockCanvas struct{}

func (m *mockCanvas) Clear(_ widget.Color)                                                  {}
func (m *mockCanvas) DrawRect(_ geometry.Rect, _ widget.Color)                              {}
func (m *mockCanvas) FillRectDirect(_ geometry.Rect, _ widget.Color)                        {}
func (m *mockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32)                 {}
func (m *mockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32)              {}
func (m *mockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {}
func (m *mockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)                {}
func (m *mockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32)   {}
func (m *mockCanvas) StrokeArc(_ geometry.Point, _ float32, _, _ float64, _ widget.Color, _ float32) {
}
func (m *mockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32) {}
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
func (m *mockCanvas) ScreenOriginBase() geometry.Point             { return geometry.Point{} }
func (m *mockCanvas) ClipBounds() geometry.Rect                    { return geometry.NewRect(0, 0, 10000, 10000) }
func (m *mockCanvas) ReplayScene(_ *scene.Scene)                   {}
