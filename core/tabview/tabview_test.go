package tabview_test

import (
	"image"
	"testing"

	"github.com/gogpu/ui/core/tabview"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// --- Construction Tests ---

func TestNew_Default(t *testing.T) {
	tabs := []tabview.Tab{
		{Label: "Tab1"},
		{Label: "Tab2"},
	}
	tv := tabview.New(tabs)

	if !tv.IsVisible() {
		t.Error("default tabview should be visible")
	}
	if !tv.IsEnabled() {
		t.Error("default tabview should be enabled")
	}
	if !tv.IsFocusable() {
		t.Error("default tabview should be focusable")
	}
	if tv.TabCount() != 2 {
		t.Errorf("TabCount() = %d, want 2", tv.TabCount())
	}
	if tv.SelectedIndex() != 0 {
		t.Errorf("SelectedIndex() = %d, want 0", tv.SelectedIndex())
	}
}

func TestNew_EmptyTabs(t *testing.T) {
	tv := tabview.New(nil)
	if tv.TabCount() != 0 {
		t.Errorf("TabCount() = %d, want 0", tv.TabCount())
	}
	if tv.Children() != nil {
		t.Error("Children() should be nil for empty tabs")
	}
}

func TestNew_WithSelectedIndex(t *testing.T) {
	tabs := []tabview.Tab{
		{Label: "Tab1"},
		{Label: "Tab2"},
		{Label: "Tab3"},
	}
	tv := tabview.New(tabs, tabview.SelectedIndex(2))
	if tv.SelectedIndex() != 2 {
		t.Errorf("SelectedIndex() = %d, want 2", tv.SelectedIndex())
	}
}

func TestNew_WithPosition(t *testing.T) {
	tabs := []tabview.Tab{{Label: "Tab1"}}

	tests := []struct {
		name     string
		position tabview.TabPosition
	}{
		{"top", tabview.Top},
		{"bottom", tabview.Bottom},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tv := tabview.New(tabs, tabview.PositionOpt(tc.position))
			_ = tv
		})
	}
}

func TestNew_AllOptions(t *testing.T) {
	tabs := []tabview.Tab{
		{Label: "Home"},
		{Label: "Settings", Closeable: true},
	}
	sig := state.NewSignal(0)
	tv := tabview.New(tabs,
		tabview.PositionOpt(tabview.Top),
		tabview.Closeable(true),
		tabview.OnSelect(func(_ int) {}),
		tabview.OnClose(func(_ int) {}),
		tabview.SelectedIndex(1),
		tabview.SelectedSignalOpt(sig),
		tabview.PainterOpt(tabview.DefaultPainter{}),
	)
	_ = tv
}

func TestNew_WithReadonlySignal(t *testing.T) {
	base := state.NewSignal(1)
	computed := state.NewComputed(func() int {
		return base.Get()
	}, base)

	tabs := []tabview.Tab{
		{Label: "Tab1"},
		{Label: "Tab2"},
		{Label: "Tab3"},
	}
	tv := tabview.New(tabs, tabview.SelectedReadonlySignalOpt(computed))
	if tv.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex() = %d, want 1 (from readonly signal)", tv.SelectedIndex())
	}

	base.Set(2)
	if tv.SelectedIndex() != 2 {
		t.Errorf("SelectedIndex() = %d, want 2 (after readonly signal update)", tv.SelectedIndex())
	}
}

// --- Children Tests ---

func TestChildren_WithContent(t *testing.T) {
	content1 := &mockWidget{}
	content2 := &mockWidget{}
	tabs := []tabview.Tab{
		{Label: "Tab1", Content: content1},
		{Label: "Tab2", Content: content2},
	}
	tv := tabview.New(tabs)
	children := tv.Children()

	if len(children) != 2 {
		t.Fatalf("Children() len = %d, want 2", len(children))
	}
}

func TestChildren_NilContent(t *testing.T) {
	tabs := []tabview.Tab{
		{Label: "Tab1"},
		{Label: "Tab2"},
	}
	tv := tabview.New(tabs)
	if tv.Children() != nil {
		t.Error("Children() should be nil when all tabs have nil content")
	}
}

func TestChildren_MixedContent(t *testing.T) {
	content := &mockWidget{}
	tabs := []tabview.Tab{
		{Label: "Tab1", Content: content},
		{Label: "Tab2"},
	}
	tv := tabview.New(tabs)
	children := tv.Children()

	if len(children) != 1 {
		t.Fatalf("Children() len = %d, want 1", len(children))
	}
}

// --- Layout Tests ---

func TestLayout_TopPosition(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 100)}
	tabs := []tabview.Tab{
		{Label: "Tab1", Content: content},
	}
	tv := tabview.New(tabs, tabview.PositionOpt(tabview.Top))
	tv.SetBounds(geometry.NewRect(0, 0, 400, 300))
	ctx := widget.NewContext()
	constraints := geometry.Tight(geometry.Sz(400, 300))

	size := tv.Layout(ctx, constraints)

	if size.Width != 400 {
		t.Errorf("width = %v, want 400", size.Width)
	}
	if size.Height != 300 {
		t.Errorf("height = %v, want 300", size.Height)
	}
}

func TestLayout_BottomPosition(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 100)}
	tabs := []tabview.Tab{
		{Label: "Tab1", Content: content},
	}
	tv := tabview.New(tabs, tabview.PositionOpt(tabview.Bottom))
	tv.SetBounds(geometry.NewRect(0, 0, 400, 300))
	ctx := widget.NewContext()
	constraints := geometry.Tight(geometry.Sz(400, 300))

	size := tv.Layout(ctx, constraints)

	if size.Width != 400 {
		t.Errorf("width = %v, want 400", size.Width)
	}
	if size.Height != 300 {
		t.Errorf("height = %v, want 300", size.Height)
	}
}

func TestLayout_OnlySelectedContentLaidOut(t *testing.T) {
	content1 := &mockWidget{preferredSize: geometry.Sz(200, 100)}
	content2 := &mockWidget{preferredSize: geometry.Sz(200, 100)}
	tabs := []tabview.Tab{
		{Label: "Tab1", Content: content1},
		{Label: "Tab2", Content: content2},
	}
	tv := tabview.New(tabs, tabview.SelectedIndex(0))
	tv.SetBounds(geometry.NewRect(0, 0, 400, 300))
	ctx := widget.NewContext()
	constraints := geometry.Tight(geometry.Sz(400, 300))

	tv.Layout(ctx, constraints)

	if content1.layoutCount != 1 {
		t.Errorf("selected content laid out %d times, want 1", content1.layoutCount)
	}
	if content2.layoutCount != 0 {
		t.Errorf("non-selected content laid out %d times, want 0", content2.layoutCount)
	}
}

func TestLayout_NilContent(t *testing.T) {
	tabs := []tabview.Tab{
		{Label: "Tab1"},
	}
	tv := tabview.New(tabs)
	tv.SetBounds(geometry.NewRect(0, 0, 400, 300))
	ctx := widget.NewContext()
	constraints := geometry.Tight(geometry.Sz(400, 300))

	// Should not panic.
	tv.Layout(ctx, constraints)
}

// --- Draw Tests ---

func TestDraw_DoesNotPanicWithBounds(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 100)}
	tabs := []tabview.Tab{
		{Label: "Tab1", Content: content},
		{Label: "Tab2"},
	}
	tv := tabview.New(tabs)
	tv.SetBounds(geometry.NewRect(0, 0, 400, 300))
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	// Layout first, then draw.
	tv.Layout(ctx, geometry.Tight(geometry.Sz(400, 300)))
	tv.Draw(ctx, canvas)
}

func TestDraw_EmptyBoundsSkips(t *testing.T) {
	tabs := []tabview.Tab{{Label: "Tab1"}}
	tv := tabview.New(tabs)
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	// No bounds set — should skip.
	tv.Draw(ctx, canvas)

	if len(canvas.drawTexts) != 0 {
		t.Error("should not draw text with empty bounds")
	}
}

func TestDraw_OnlySelectedContentDrawn(t *testing.T) {
	content1 := &mockWidget{preferredSize: geometry.Sz(200, 100)}
	content2 := &mockWidget{preferredSize: geometry.Sz(200, 100)}
	tabs := []tabview.Tab{
		{Label: "Tab1", Content: content1},
		{Label: "Tab2", Content: content2},
	}
	tv := tabview.New(tabs, tabview.SelectedIndex(1))
	tv.SetBounds(geometry.NewRect(0, 0, 400, 300))
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	tv.Layout(ctx, geometry.Tight(geometry.Sz(400, 300)))
	tv.Draw(ctx, canvas)

	if content1.drawCount != 0 {
		t.Errorf("non-selected content drawn %d times, want 0", content1.drawCount)
	}
	if content2.drawCount != 1 {
		t.Errorf("selected content drawn %d times, want 1", content2.drawCount)
	}
}

func TestDraw_AllPositions(t *testing.T) {
	positions := []tabview.TabPosition{tabview.Top, tabview.Bottom}
	for _, pos := range positions {
		t.Run(pos.String(), func(t *testing.T) {
			tabs := []tabview.Tab{{Label: "Tab1"}}
			tv := tabview.New(tabs, tabview.PositionOpt(pos))
			tv.SetBounds(geometry.NewRect(0, 0, 400, 300))
			ctx := widget.NewContext()
			canvas := &mockCanvas{}

			tv.Layout(ctx, geometry.Tight(geometry.Sz(400, 300)))
			tv.Draw(ctx, canvas)
		})
	}
}

// --- Event Handling Tests ---

func TestEvent_ClickSelectsTab(t *testing.T) {
	selected := -1
	tabs := []tabview.Tab{
		{Label: "Tab1"},
		{Label: "Tab2"},
		{Label: "Tab3"},
	}
	tv := tabview.New(tabs,
		tabview.SelectedIndex(0),
		tabview.OnSelect(func(idx int) { selected = idx }),
	)
	tv.SetBounds(geometry.NewRect(0, 0, 300, 300))
	ctx := widget.NewContext()
	tv.Layout(ctx, geometry.Tight(geometry.Sz(300, 300)))

	// Click on the second tab (at x=150, which is center of tab2 with 100px tabs).
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(150, 24), geometry.Pt(150, 24), event.ModNone)
	consumed := tv.Event(ctx, press)

	if !consumed {
		t.Error("click on tab should be consumed")
	}
	if selected != 1 {
		t.Errorf("selected = %d, want 1", selected)
	}
	if tv.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex() = %d, want 1", tv.SelectedIndex())
	}
}

func TestEvent_ClickDisabledTabIgnored(t *testing.T) {
	tabs := []tabview.Tab{
		{Label: "Tab1"},
		{Label: "Tab2", Disabled: true},
	}
	tv := tabview.New(tabs, tabview.SelectedIndex(0))
	tv.SetBounds(geometry.NewRect(0, 0, 200, 300))
	ctx := widget.NewContext()
	tv.Layout(ctx, geometry.Tight(geometry.Sz(200, 300)))

	// Click on the disabled second tab.
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(150, 24), geometry.Pt(150, 24), event.ModNone)
	tv.Event(ctx, press)

	if tv.SelectedIndex() != 0 {
		t.Errorf("SelectedIndex() = %d, want 0 (disabled tab should not be selected)", tv.SelectedIndex())
	}
}

func TestEvent_CloseButton(t *testing.T) {
	closedIdx := -1
	tabs := []tabview.Tab{
		{Label: "Tab1"},
		{Label: "Tab2"},
	}
	tv := tabview.New(tabs,
		tabview.Closeable(true),
		tabview.OnClose(func(idx int) { closedIdx = idx }),
	)
	tv.SetBounds(geometry.NewRect(0, 0, 300, 300))
	ctx := widget.NewContext()
	tv.Layout(ctx, geometry.Tight(geometry.Sz(300, 300)))

	// Click on the close button area of the first tab (near right edge).
	// Tab1 is 0-150px wide; close button is at right edge minus padding.
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(130, 24), geometry.Pt(130, 24), event.ModNone)
	consumed := tv.Event(ctx, press)

	if !consumed {
		t.Error("click on close button should be consumed")
	}
	if closedIdx != 0 {
		t.Errorf("closedIdx = %d, want 0", closedIdx)
	}
}

func TestEvent_RightClickIgnored(t *testing.T) {
	tabs := []tabview.Tab{
		{Label: "Tab1"},
		{Label: "Tab2"},
	}
	tv := tabview.New(tabs)
	tv.SetBounds(geometry.NewRect(0, 0, 200, 300))
	ctx := widget.NewContext()
	tv.Layout(ctx, geometry.Tight(geometry.Sz(200, 300)))

	press := event.NewMouseEvent(event.MousePress, event.ButtonRight, 0,
		geometry.Pt(50, 24), geometry.Pt(50, 24), event.ModNone)
	consumed := tv.Event(ctx, press)

	if consumed {
		t.Error("right-click should not be consumed")
	}
}

func TestEvent_ClickOutsideTabBar(t *testing.T) {
	tabs := []tabview.Tab{{Label: "Tab1"}}
	tv := tabview.New(tabs)
	tv.SetBounds(geometry.NewRect(0, 0, 200, 300))
	ctx := widget.NewContext()
	tv.Layout(ctx, geometry.Tight(geometry.Sz(200, 300)))

	// Click in the content area (below tab bar).
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(100, 200), geometry.Pt(100, 200), event.ModNone)
	consumed := tv.Event(ctx, press)

	if consumed {
		t.Error("click outside tab bar should not be consumed by tab bar handler")
	}
}

// --- Keyboard Navigation Tests ---

func TestEvent_ArrowRightNavigation(t *testing.T) {
	tabs := []tabview.Tab{
		{Label: "Tab1"},
		{Label: "Tab2"},
		{Label: "Tab3"},
	}
	tv := tabview.New(tabs, tabview.SelectedIndex(0))
	tv.SetFocused(true)
	ctx := widget.NewContext()

	key := event.NewKeyEvent(event.KeyPress, event.KeyRight, 0, event.ModNone)
	consumed := tv.Event(ctx, key)

	if !consumed {
		t.Error("arrow right should be consumed")
	}
	if tv.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex() = %d, want 1", tv.SelectedIndex())
	}
}

func TestEvent_ArrowLeftNavigation(t *testing.T) {
	tabs := []tabview.Tab{
		{Label: "Tab1"},
		{Label: "Tab2"},
		{Label: "Tab3"},
	}
	tv := tabview.New(tabs, tabview.SelectedIndex(2))
	tv.SetFocused(true)
	ctx := widget.NewContext()

	key := event.NewKeyEvent(event.KeyPress, event.KeyLeft, 0, event.ModNone)
	consumed := tv.Event(ctx, key)

	if !consumed {
		t.Error("arrow left should be consumed")
	}
	if tv.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex() = %d, want 1", tv.SelectedIndex())
	}
}

func TestEvent_ArrowRightWrapsAround(t *testing.T) {
	tabs := []tabview.Tab{
		{Label: "Tab1"},
		{Label: "Tab2"},
	}
	tv := tabview.New(tabs, tabview.SelectedIndex(1))
	tv.SetFocused(true)
	ctx := widget.NewContext()

	key := event.NewKeyEvent(event.KeyPress, event.KeyRight, 0, event.ModNone)
	tv.Event(ctx, key)

	if tv.SelectedIndex() != 0 {
		t.Errorf("SelectedIndex() = %d, want 0 (wrapped around)", tv.SelectedIndex())
	}
}

func TestEvent_ArrowLeftWrapsAround(t *testing.T) {
	tabs := []tabview.Tab{
		{Label: "Tab1"},
		{Label: "Tab2"},
	}
	tv := tabview.New(tabs, tabview.SelectedIndex(0))
	tv.SetFocused(true)
	ctx := widget.NewContext()

	key := event.NewKeyEvent(event.KeyPress, event.KeyLeft, 0, event.ModNone)
	tv.Event(ctx, key)

	if tv.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex() = %d, want 1 (wrapped around)", tv.SelectedIndex())
	}
}

func TestEvent_ArrowSkipsDisabledTabs(t *testing.T) {
	tabs := []tabview.Tab{
		{Label: "Tab1"},
		{Label: "Tab2", Disabled: true},
		{Label: "Tab3"},
	}
	tv := tabview.New(tabs, tabview.SelectedIndex(0))
	tv.SetFocused(true)
	ctx := widget.NewContext()

	key := event.NewKeyEvent(event.KeyPress, event.KeyRight, 0, event.ModNone)
	tv.Event(ctx, key)

	if tv.SelectedIndex() != 2 {
		t.Errorf("SelectedIndex() = %d, want 2 (skipped disabled)", tv.SelectedIndex())
	}
}

func TestEvent_HomeSelectsFirst(t *testing.T) {
	tabs := []tabview.Tab{
		{Label: "Tab1"},
		{Label: "Tab2"},
		{Label: "Tab3"},
	}
	tv := tabview.New(tabs, tabview.SelectedIndex(2))
	tv.SetFocused(true)
	ctx := widget.NewContext()

	key := event.NewKeyEvent(event.KeyPress, event.KeyHome, 0, event.ModNone)
	consumed := tv.Event(ctx, key)

	if !consumed {
		t.Error("Home key should be consumed")
	}
	if tv.SelectedIndex() != 0 {
		t.Errorf("SelectedIndex() = %d, want 0", tv.SelectedIndex())
	}
}

func TestEvent_EndSelectsLast(t *testing.T) {
	tabs := []tabview.Tab{
		{Label: "Tab1"},
		{Label: "Tab2"},
		{Label: "Tab3"},
	}
	tv := tabview.New(tabs, tabview.SelectedIndex(0))
	tv.SetFocused(true)
	ctx := widget.NewContext()

	key := event.NewKeyEvent(event.KeyPress, event.KeyEnd, 0, event.ModNone)
	consumed := tv.Event(ctx, key)

	if !consumed {
		t.Error("End key should be consumed")
	}
	if tv.SelectedIndex() != 2 {
		t.Errorf("SelectedIndex() = %d, want 2", tv.SelectedIndex())
	}
}

func TestEvent_HomeSkipsDisabled(t *testing.T) {
	tabs := []tabview.Tab{
		{Label: "Tab1", Disabled: true},
		{Label: "Tab2"},
		{Label: "Tab3"},
	}
	tv := tabview.New(tabs, tabview.SelectedIndex(2))
	tv.SetFocused(true)
	ctx := widget.NewContext()

	key := event.NewKeyEvent(event.KeyPress, event.KeyHome, 0, event.ModNone)
	tv.Event(ctx, key)

	if tv.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex() = %d, want 1 (first enabled)", tv.SelectedIndex())
	}
}

func TestEvent_EndSkipsDisabled(t *testing.T) {
	tabs := []tabview.Tab{
		{Label: "Tab1"},
		{Label: "Tab2"},
		{Label: "Tab3", Disabled: true},
	}
	tv := tabview.New(tabs, tabview.SelectedIndex(0))
	tv.SetFocused(true)
	ctx := widget.NewContext()

	key := event.NewKeyEvent(event.KeyPress, event.KeyEnd, 0, event.ModNone)
	tv.Event(ctx, key)

	if tv.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex() = %d, want 1 (last enabled)", tv.SelectedIndex())
	}
}

func TestEvent_KeyIgnoredWhenNotFocused(t *testing.T) {
	tabs := []tabview.Tab{
		{Label: "Tab1"},
		{Label: "Tab2"},
	}
	tv := tabview.New(tabs, tabview.SelectedIndex(0))
	// Not focused.
	ctx := widget.NewContext()

	key := event.NewKeyEvent(event.KeyPress, event.KeyRight, 0, event.ModNone)
	consumed := tv.Event(ctx, key)

	if consumed {
		t.Error("key event should not be consumed when not focused")
	}
	if tv.SelectedIndex() != 0 {
		t.Errorf("SelectedIndex() = %d, want 0 (unchanged)", tv.SelectedIndex())
	}
}

func TestEvent_KeyReleaseIgnored(t *testing.T) {
	tabs := []tabview.Tab{
		{Label: "Tab1"},
		{Label: "Tab2"},
	}
	tv := tabview.New(tabs, tabview.SelectedIndex(0))
	tv.SetFocused(true)
	ctx := widget.NewContext()

	key := event.NewKeyEvent(event.KeyRelease, event.KeyRight, 0, event.ModNone)
	consumed := tv.Event(ctx, key)

	if consumed {
		t.Error("key release should not be consumed")
	}
}

func TestEvent_UnhandledKeyIgnored(t *testing.T) {
	tabs := []tabview.Tab{
		{Label: "Tab1"},
		{Label: "Tab2"},
	}
	tv := tabview.New(tabs, tabview.SelectedIndex(0))
	tv.SetFocused(true)
	ctx := widget.NewContext()

	key := event.NewKeyEvent(event.KeyPress, event.KeyA, 0, event.ModNone)
	consumed := tv.Event(ctx, key)

	if consumed {
		t.Error("unhandled key should not be consumed")
	}
}

func TestEvent_AllDisabledArrowNoChange(t *testing.T) {
	tabs := []tabview.Tab{
		{Label: "Tab1", Disabled: true},
		{Label: "Tab2", Disabled: true},
	}
	tv := tabview.New(tabs, tabview.SelectedIndex(0))
	tv.SetFocused(true)
	ctx := widget.NewContext()

	key := event.NewKeyEvent(event.KeyPress, event.KeyRight, 0, event.ModNone)
	consumed := tv.Event(ctx, key)

	if consumed {
		t.Error("arrow with all disabled tabs should not be consumed")
	}
}

// --- Signal Binding Tests ---

func TestSelectedSignal_Bidirectional(t *testing.T) {
	sig := state.NewSignal(0)
	tabs := []tabview.Tab{
		{Label: "Tab1"},
		{Label: "Tab2"},
		{Label: "Tab3"},
	}
	tv := tabview.New(tabs, tabview.SelectedSignalOpt(sig))

	// Signal -> widget.
	sig.Set(2)
	if tv.SelectedIndex() != 2 {
		t.Errorf("SelectedIndex() = %d, want 2 (from signal)", tv.SelectedIndex())
	}

	// Widget -> signal (via keyboard).
	tv.SetFocused(true)
	tv.SetBounds(geometry.NewRect(0, 0, 300, 300))
	ctx := widget.NewContext()
	tv.Layout(ctx, geometry.Tight(geometry.Sz(300, 300)))

	key := event.NewKeyEvent(event.KeyPress, event.KeyLeft, 0, event.ModNone)
	tv.Event(ctx, key)

	if sig.Get() != 1 {
		t.Errorf("signal.Get() = %d, want 1 (from widget keyboard nav)", sig.Get())
	}
}

func TestSelectedSignal_Priority(t *testing.T) {
	sig := state.NewSignal(2)
	tabs := []tabview.Tab{
		{Label: "Tab1"},
		{Label: "Tab2"},
		{Label: "Tab3"},
	}
	tv := tabview.New(tabs,
		tabview.SelectedIndex(0),
		tabview.SelectedSignalOpt(sig),
	)

	if tv.SelectedIndex() != 2 {
		t.Errorf("SelectedIndex() = %d, want 2 (signal overrides static)", tv.SelectedIndex())
	}
}

func TestReadonlySignal_HighestPriority(t *testing.T) {
	rw := state.NewSignal(0)
	ro := state.NewComputed(func() int { return 2 })

	tabs := []tabview.Tab{
		{Label: "Tab1"},
		{Label: "Tab2"},
		{Label: "Tab3"},
	}
	tv := tabview.New(tabs,
		tabview.SelectedIndex(0),
		tabview.SelectedSignalOpt(rw),
		tabview.SelectedReadonlySignalOpt(ro),
	)

	if tv.SelectedIndex() != 2 {
		t.Errorf("SelectedIndex() = %d, want 2 (readonly signal has highest priority)", tv.SelectedIndex())
	}
}

// --- Focus Tests ---

func TestFocus_SetFocused(t *testing.T) {
	tabs := []tabview.Tab{{Label: "Tab1"}}
	tv := tabview.New(tabs)

	tv.SetFocused(true)
	if !tv.IsFocused() {
		t.Error("should be focused after SetFocused(true)")
	}

	tv.SetFocused(false)
	if tv.IsFocused() {
		t.Error("should not be focused after SetFocused(false)")
	}
}

func TestFocusable_VisibleAndEnabled(t *testing.T) {
	tabs := []tabview.Tab{{Label: "Tab1"}}
	tv := tabview.New(tabs)

	if !tv.IsFocusable() {
		t.Error("visible+enabled tabview should be focusable")
	}

	tv.SetVisible(false)
	if tv.IsFocusable() {
		t.Error("invisible tabview should not be focusable")
	}

	tv.SetVisible(true)
	tv.SetEnabled(false)
	if tv.IsFocusable() {
		t.Error("disabled tabview should not be focusable")
	}
}

// --- Lifecycle Tests ---

func TestLifecycleInterface(t *testing.T) {
	tabs := []tabview.Tab{{Label: "Tab1"}}
	var _ widget.Lifecycle = tabview.New(tabs)
}

func TestMount_CreatesBindings(t *testing.T) {
	sig := state.NewSignal(0)
	tabs := []tabview.Tab{
		{Label: "Tab1"},
		{Label: "Tab2"},
	}
	tv := tabview.New(tabs, tabview.SelectedSignalOpt(sig))

	dirtyCount := 0
	sched := state.NewScheduler(func(_ []widget.Widget) {})
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	tv.Mount(ctx)
	sched.SetOnDirty(func() { dirtyCount++ })
	sig.Set(1)

	if dirtyCount == 0 {
		t.Error("signal change should mark widget dirty after mount")
	}
}

func TestMount_ReadonlySignal(t *testing.T) {
	base := state.NewSignal(0)
	computed := state.NewComputed(func() int { return base.Get() }, base)

	tabs := []tabview.Tab{
		{Label: "Tab1"},
		{Label: "Tab2"},
	}
	tv := tabview.New(tabs, tabview.SelectedReadonlySignalOpt(computed))

	dirtyCount := 0
	sched := state.NewScheduler(func(_ []widget.Widget) {})
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	tv.Mount(ctx)
	sched.SetOnDirty(func() { dirtyCount++ })
	base.Set(1)

	if dirtyCount == 0 {
		t.Error("computed signal change should mark widget dirty after mount")
	}
}

func TestUnmount_CleansBindings(t *testing.T) {
	sig := state.NewSignal(0)
	tabs := []tabview.Tab{
		{Label: "Tab1"},
		{Label: "Tab2"},
	}
	tv := tabview.New(tabs, tabview.SelectedSignalOpt(sig))

	sched := state.NewScheduler(func(_ []widget.Widget) {})
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	tv.Mount(ctx)
	tv.CleanupBindings()
	tv.Unmount()

	sig.Set(1)

	if sched.PendingCount() != 0 {
		t.Error("signal change after unmount should not mark widget dirty")
	}
}

func TestMount_NoScheduler_NoPanic(t *testing.T) {
	sig := state.NewSignal(0)
	tabs := []tabview.Tab{{Label: "Tab1"}}
	tv := tabview.New(tabs, tabview.SelectedSignalOpt(sig))
	ctx := widget.NewContext()

	// Should not panic even without scheduler.
	tv.Mount(ctx)
}

func TestMount_NoSignals_NoPanic(t *testing.T) {
	tabs := []tabview.Tab{{Label: "Tab1"}}
	tv := tabview.New(tabs)

	sched := state.NewScheduler(func(_ []widget.Widget) {})
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	// Should not panic with no signals.
	tv.Mount(ctx)
}

// --- Widget Interface Compliance ---

func TestWidgetInterface(t *testing.T) {
	tabs := []tabview.Tab{{Label: "Tab1"}}
	var w widget.Widget = tabview.New(tabs)
	_ = w
}

func TestFocusableInterface(t *testing.T) {
	tabs := []tabview.Tab{{Label: "Tab1"}}
	var f widget.Focusable = tabview.New(tabs)
	_ = f
}

// --- Mouse Hover Tests ---

func TestEvent_MouseMoveUpdatesHover(t *testing.T) {
	tabs := []tabview.Tab{
		{Label: "Tab1"},
		{Label: "Tab2"},
	}
	tv := tabview.New(tabs)
	tv.SetBounds(geometry.NewRect(0, 0, 200, 300))
	ctx := widget.NewContext()
	tv.Layout(ctx, geometry.Tight(geometry.Sz(200, 300)))

	move := event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(50, 24), geometry.Pt(50, 24), event.ModNone)
	consumed := tv.Event(ctx, move)

	if !consumed {
		t.Error("mouse move over tab should be consumed")
	}
}

func TestEvent_MouseLeave(t *testing.T) {
	tabs := []tabview.Tab{
		{Label: "Tab1"},
		{Label: "Tab2"},
	}
	tv := tabview.New(tabs)
	tv.SetBounds(geometry.NewRect(0, 0, 200, 300))
	ctx := widget.NewContext()
	tv.Layout(ctx, geometry.Tight(geometry.Sz(200, 300)))

	// First hover.
	move := event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(50, 24), geometry.Pt(50, 24), event.ModNone)
	tv.Event(ctx, move)

	// Then leave.
	leave := event.NewMouseEvent(event.MouseLeave, event.ButtonNone, 0,
		geometry.Pt(-1, -1), geometry.Pt(-1, -1), event.ModNone)
	consumed := tv.Event(ctx, leave)

	if !consumed {
		t.Error("mouse leave should be consumed when hover was active")
	}
}

// --- TabPosition String Tests ---

func TestTabPosition_String(t *testing.T) {
	tests := []struct {
		pos  tabview.TabPosition
		want string
	}{
		{tabview.Top, "Top"},
		{tabview.Bottom, "Bottom"},
		{tabview.TabPosition(99), "Unknown"},
	}
	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			if got := tc.pos.String(); got != tc.want {
				t.Errorf("String() = %q, want %q", got, tc.want)
			}
		})
	}
}

// --- Event forwarding to content ---

func TestEvent_ForwardedToSelectedContent(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 100)}
	tabs := []tabview.Tab{
		{Label: "Tab1", Content: content},
	}
	tv := tabview.New(tabs)
	tv.SetBounds(geometry.NewRect(0, 0, 200, 300))
	ctx := widget.NewContext()
	tv.Layout(ctx, geometry.Tight(geometry.Sz(200, 300)))

	// Send an event — content will consume it.
	content.consumeEvents = true
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(100, 200), geometry.Pt(100, 200), event.ModNone)
	consumed := tv.Event(ctx, press)

	if !consumed {
		t.Error("event should be consumed by content widget")
	}
	if content.eventCount != 1 {
		t.Errorf("content event count = %d, want 1", content.eventCount)
	}
}

// --- Per-tab closeable ---

func TestPerTabCloseable(t *testing.T) {
	tabs := []tabview.Tab{
		{Label: "Tab1"},
		{Label: "Tab2", Closeable: true},
	}
	tv := tabview.New(tabs) // Global closeable is false.
	tv.SetBounds(geometry.NewRect(0, 0, 200, 300))
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	tv.Layout(ctx, geometry.Tight(geometry.Sz(200, 300)))
	tv.Draw(ctx, canvas) // Should not panic.
}

// --- Selecting same tab is no-op ---

func TestEvent_SelectSameTabNoCallback(t *testing.T) {
	callCount := 0
	tabs := []tabview.Tab{
		{Label: "Tab1"},
		{Label: "Tab2"},
	}
	tv := tabview.New(tabs,
		tabview.SelectedIndex(0),
		tabview.OnSelect(func(_ int) { callCount++ }),
	)
	tv.SetBounds(geometry.NewRect(0, 0, 200, 300))
	ctx := widget.NewContext()
	tv.Layout(ctx, geometry.Tight(geometry.Sz(200, 300)))

	// Click on already-selected tab.
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(50, 24), geometry.Pt(50, 24), event.ModNone)
	tv.Event(ctx, press)

	if callCount != 0 {
		t.Errorf("callCount = %d, want 0 (selecting same tab should be no-op)", callCount)
	}
}

// --- mockWidget for testing ---

type mockWidget struct {
	widget.WidgetBase
	preferredSize geometry.Size
	layoutCount   int
	drawCount     int
	eventCount    int
	consumeEvents bool
}

func (w *mockWidget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	w.layoutCount++
	return constraints.Constrain(w.preferredSize)
}

func (w *mockWidget) Draw(_ widget.Context, _ widget.Canvas) {
	w.drawCount++
}

func (w *mockWidget) Event(_ widget.Context, _ event.Event) bool {
	w.eventCount++
	return w.consumeEvents
}

func (w *mockWidget) Children() []widget.Widget {
	return nil
}

// --- recordingCanvas records draw calls for verification ---

type recordingCanvas struct {
	drawTexts []drawTextCall
}

type drawTextCall struct {
	text     string
	bounds   geometry.Rect
	fontSize float32
	color    widget.Color
	bold     bool
	align    widget.TextAlign
}

func (c *recordingCanvas) Clear(_ widget.Color)                                     {}
func (c *recordingCanvas) DrawRect(_ geometry.Rect, _ widget.Color)                 {}
func (c *recordingCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32)    {}
func (c *recordingCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32) {}
func (c *recordingCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {
}
func (c *recordingCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)              {}
func (c *recordingCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {}
func (c *recordingCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32)             {}

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
