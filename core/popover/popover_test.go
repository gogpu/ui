package popover_test

import (
	"image"
	"testing"
	"time"

	"github.com/gogpu/ui/core/popover"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// =============================================================================
// Placement Tests
// =============================================================================

func TestPlacement_String(t *testing.T) {
	tests := []struct {
		p    popover.Placement
		want string
	}{
		{popover.Bottom, "Bottom"},
		{popover.BottomStart, "BottomStart"},
		{popover.BottomEnd, "BottomEnd"},
		{popover.Top, "Top"},
		{popover.TopStart, "TopStart"},
		{popover.TopEnd, "TopEnd"},
		{popover.Left, "Left"},
		{popover.LeftStart, "LeftStart"},
		{popover.LeftEnd, "LeftEnd"},
		{popover.Right, "Right"},
		{popover.RightStart, "RightStart"},
		{popover.RightEnd, "RightEnd"},
		{popover.Placement(99), "Unknown"},
	}
	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			got := tc.p.String()
			if got != tc.want {
				t.Errorf("String() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestCalculatePosition_AllPlacements(t *testing.T) {
	// NewRect(x, y, w, h) -> Min=(100,100), Max=(200,140), Center=(150,120)
	anchor := geometry.NewRect(100, 100, 100, 40)
	overlaySize := geometry.Sz(80, 30)
	windowSize := geometry.Sz(800, 600)
	gap := float32(4)

	tests := []struct {
		placement popover.Placement
		wantX     float32
		wantY     float32
	}{
		{popover.Bottom, 110, 144},      // centered: 150-40=110, below: 140+4=144
		{popover.BottomStart, 100, 144}, // start: 100, below: 140+4=144
		{popover.BottomEnd, 120, 144},   // end: 200-80=120, below: 140+4=144
		{popover.Top, 110, 66},          // centered: 150-40=110, above: 100-30-4=66
		{popover.TopStart, 100, 66},
		{popover.TopEnd, 120, 66},
		{popover.Left, 16, 105}, // left: 100-80-4=16, centered: 120-15=105
		{popover.LeftStart, 16, 100},
		{popover.LeftEnd, 16, 110}, // left: 100-80-4=16, end: 140-30=110
		{popover.Right, 204, 105},  // right: 200+4=204, centered: 120-15=105
		{popover.RightStart, 204, 100},
		{popover.RightEnd, 204, 110}, // right: 200+4=204, end: 140-30=110
	}

	for _, tc := range tests {
		t.Run(tc.placement.String(), func(t *testing.T) {
			pos := popover.CalculatePosition(tc.placement, anchor, overlaySize, windowSize, gap)
			if pos.X != tc.wantX || pos.Y != tc.wantY {
				t.Errorf("pos = (%v, %v), want (%v, %v)", pos.X, pos.Y, tc.wantX, tc.wantY)
			}
		})
	}
}

func TestCalculatePosition_AutoFlip(t *testing.T) {
	// Anchor near the bottom edge; Bottom placement should flip to Top.
	// NewRect(x, y, w, h) -> Min=(100,570), Max=(200,590)
	anchor := geometry.NewRect(100, 570, 100, 20)
	overlaySize := geometry.Sz(80, 30)
	windowSize := geometry.Sz(800, 600)
	gap := float32(4)

	pos := popover.CalculatePosition(popover.Bottom, anchor, overlaySize, windowSize, gap)
	// Should flip to top: 570 - 30 - 4 = 536
	if pos.Y != 536 {
		t.Errorf("Y = %v, want 536 (flipped to top)", pos.Y)
	}
}

func TestCalculatePosition_AutoFlipRight(t *testing.T) {
	// Anchor near the right edge; Right placement should flip to Left.
	// NewRect(x, y, w, h) -> Min=(750,100), Max=(790,140)
	anchor := geometry.NewRect(750, 100, 40, 40)
	overlaySize := geometry.Sz(80, 30)
	windowSize := geometry.Sz(800, 600)
	gap := float32(4)

	pos := popover.CalculatePosition(popover.Right, anchor, overlaySize, windowSize, gap)
	// Should flip to left: 750 - 80 - 4 = 666
	if pos.X != 666 {
		t.Errorf("X = %v, want 666 (flipped to left)", pos.X)
	}
}

func TestCalculatePosition_ClampToViewport(t *testing.T) {
	// Anchor near top-left corner, overlay would go negative.
	// NewRect(x, y, w, h) -> Min=(0,0), Max=(20,20)
	anchor := geometry.NewRect(0, 0, 20, 20)
	overlaySize := geometry.Sz(80, 30)
	windowSize := geometry.Sz(800, 600)
	gap := float32(4)

	pos := popover.CalculatePosition(popover.Top, anchor, overlaySize, windowSize, gap)
	// Top would be negative, flip to bottom: 20+4=24
	if pos.Y < 0 {
		t.Errorf("Y = %v, should not be negative (clamped)", pos.Y)
	}
}

// =============================================================================
// Popover Construction Tests
// =============================================================================

func TestNewPopover_Defaults(t *testing.T) {
	pop := popover.NewPopover()

	if !pop.IsVisible() {
		t.Error("default popover should be visible")
	}
	if !pop.IsEnabled() {
		t.Error("default popover should be enabled")
	}
	if pop.IsOpen() {
		t.Error("popover should not be open by default")
	}
}

func TestNewPopover_WithTrigger(t *testing.T) {
	trigger := newMockWidget(geometry.NewRect(10, 10, 110, 50))
	pop := popover.NewPopover(
		popover.TriggerWidget(trigger),
	)

	children := pop.Children()
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}
}

func TestNewPopover_WithContent(t *testing.T) {
	content := newMockWidget(geometry.NewRect(0, 0, 200, 150))
	pop := popover.NewPopover(
		popover.Content(content),
	)

	if pop.IsOpen() {
		t.Error("should not be open until Show()")
	}
}

func TestNewPopover_WithPlacement(t *testing.T) {
	pop := popover.NewPopover(
		popover.PlacementOpt(popover.TopEnd),
	)
	// Just verifying construction doesn't panic.
	_ = pop
}

func TestNewPopover_WithGap(t *testing.T) {
	pop := popover.NewPopover(
		popover.Gap(8),
	)
	_ = pop
}

func TestNewPopover_WithCallbacks(t *testing.T) {
	showCalled := false
	hideCalled := false

	trigger := newMockWidget(geometry.NewRect(10, 10, 110, 50))
	content := newMockWidget(geometry.NewRect(0, 0, 100, 80))

	pop := popover.NewPopover(
		popover.TriggerWidget(trigger),
		popover.Content(content),
		popover.OnShow(func() { showCalled = true }),
		popover.OnHide(func() { hideCalled = true }),
	)

	ctx := widget.NewContext()
	om := setupOverlayManager(ctx)

	pop.SetBounds(geometry.NewRect(10, 10, 110, 50))
	pop.Show(ctx)

	if !showCalled {
		t.Error("OnShow callback should have been called")
	}
	if om.pushCount != 1 {
		t.Errorf("pushCount = %d, want 1", om.pushCount)
	}

	pop.Hide(ctx)

	if !hideCalled {
		t.Error("OnHide callback should have been called")
	}
}

func TestNewPopover_NoChildren_WithoutTrigger(t *testing.T) {
	pop := popover.NewPopover()

	children := pop.Children()
	if children != nil {
		t.Errorf("expected nil children without trigger, got %d", len(children))
	}
}

// =============================================================================
// Popover Show/Hide Tests
// =============================================================================

func TestPopover_ShowHide(t *testing.T) {
	trigger := newMockWidget(geometry.NewRect(10, 10, 110, 50))
	content := newMockWidget(geometry.NewRect(0, 0, 100, 80))

	pop := popover.NewPopover(
		popover.TriggerWidget(trigger),
		popover.Content(content),
	)
	pop.SetBounds(geometry.NewRect(10, 10, 110, 50))

	ctx := widget.NewContext()
	om := setupOverlayManager(ctx)

	pop.Show(ctx)
	if !pop.IsOpen() {
		t.Error("should be open after Show()")
	}
	if om.pushCount != 1 {
		t.Errorf("pushCount = %d, want 1", om.pushCount)
	}

	pop.Hide(ctx)
	if pop.IsOpen() {
		t.Error("should be closed after Hide()")
	}
	if om.removeCount != 1 {
		t.Errorf("removeCount = %d, want 1", om.removeCount)
	}
}

func TestPopover_ShowNoOpWhenOpen(t *testing.T) {
	trigger := newMockWidget(geometry.NewRect(10, 10, 110, 50))
	content := newMockWidget(geometry.NewRect(0, 0, 100, 80))

	pop := popover.NewPopover(
		popover.TriggerWidget(trigger),
		popover.Content(content),
	)
	pop.SetBounds(geometry.NewRect(10, 10, 110, 50))

	ctx := widget.NewContext()
	om := setupOverlayManager(ctx)

	pop.Show(ctx)
	pop.Show(ctx) // Should be no-op.

	if om.pushCount != 1 {
		t.Errorf("pushCount = %d, want 1 (Show when already open should be no-op)", om.pushCount)
	}
}

func TestPopover_HideNoOpWhenClosed(t *testing.T) {
	pop := popover.NewPopover()
	ctx := widget.NewContext()

	pop.Hide(ctx) // Should not panic.

	if pop.IsOpen() {
		t.Error("should remain closed")
	}
}

func TestPopover_ShowNoOpWhenDisabled(t *testing.T) {
	trigger := newMockWidget(geometry.NewRect(10, 10, 110, 50))
	content := newMockWidget(geometry.NewRect(0, 0, 100, 80))

	pop := popover.NewPopover(
		popover.TriggerWidget(trigger),
		popover.Content(content),
		popover.Disabled(true),
	)
	pop.SetBounds(geometry.NewRect(10, 10, 110, 50))

	ctx := widget.NewContext()
	setupOverlayManager(ctx)

	pop.Show(ctx)
	if pop.IsOpen() {
		t.Error("disabled popover should not open")
	}
}

func TestPopover_ShowNoOpWithoutContent(t *testing.T) {
	trigger := newMockWidget(geometry.NewRect(10, 10, 110, 50))
	pop := popover.NewPopover(
		popover.TriggerWidget(trigger),
	)
	pop.SetBounds(geometry.NewRect(10, 10, 110, 50))

	ctx := widget.NewContext()
	om := setupOverlayManager(ctx)

	pop.Show(ctx)
	// visible is set to true, but no overlay is pushed since content is nil.
	if om.pushCount != 0 {
		t.Errorf("should not push overlay without content, pushCount = %d", om.pushCount)
	}
}

func TestPopover_ShowNoOpWithoutOverlayManager(t *testing.T) {
	content := newMockWidget(geometry.NewRect(0, 0, 100, 80))
	pop := popover.NewPopover(
		popover.Content(content),
	)

	ctx := widget.NewContext() // No overlay manager set.

	pop.Show(ctx)
	if pop.IsOpen() {
		t.Error("should not open without overlay manager")
	}
}

func TestPopover_Toggle(t *testing.T) {
	trigger := newMockWidget(geometry.NewRect(10, 10, 110, 50))
	content := newMockWidget(geometry.NewRect(0, 0, 100, 80))

	pop := popover.NewPopover(
		popover.TriggerWidget(trigger),
		popover.Content(content),
	)
	pop.SetBounds(geometry.NewRect(10, 10, 110, 50))

	ctx := widget.NewContext()
	setupOverlayManager(ctx)

	pop.Toggle(ctx)
	if !pop.IsOpen() {
		t.Error("should be open after first toggle")
	}

	pop.Toggle(ctx)
	if pop.IsOpen() {
		t.Error("should be closed after second toggle")
	}
}

func TestPopover_ContentSize(t *testing.T) {
	trigger := newMockWidget(geometry.NewRect(10, 10, 110, 50))
	content := newMockWidget(geometry.NewRect(0, 0, 100, 80))

	pop := popover.NewPopover(
		popover.TriggerWidget(trigger),
		popover.Content(content),
		popover.ContentSize(250, 180),
	)
	pop.SetBounds(geometry.NewRect(10, 10, 110, 50))

	ctx := widget.NewContext()
	setupOverlayManager(ctx)

	pop.Show(ctx)
	if !pop.IsOpen() {
		t.Error("should be open")
	}
}

// =============================================================================
// Popover Event Tests
// =============================================================================

func TestPopover_ClickToToggle(t *testing.T) {
	trigger := newMockWidget(geometry.NewRect(10, 10, 110, 50))
	content := newMockWidget(geometry.NewRect(0, 0, 100, 80))

	pop := popover.NewPopover(
		popover.TriggerWidget(trigger),
		popover.Content(content),
	)
	pop.SetBounds(geometry.NewRect(10, 10, 110, 50))

	ctx := widget.NewContext()
	setupOverlayManager(ctx)

	// Click to open.
	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		geometry.Pt(50, 30), geometry.Pt(50, 30), event.ModNone)
	consumed := pop.Event(ctx, release)
	if !consumed {
		t.Error("click on trigger should be consumed")
	}
	if !pop.IsOpen() {
		t.Error("should be open after click")
	}

	// Click again to close.
	consumed = pop.Event(ctx, release)
	if !consumed {
		t.Error("second click should be consumed")
	}
	if pop.IsOpen() {
		t.Error("should be closed after second click")
	}
}

func TestPopover_ClickOutsideTrigger(t *testing.T) {
	trigger := newMockWidget(geometry.NewRect(10, 10, 110, 50))
	content := newMockWidget(geometry.NewRect(0, 0, 100, 80))

	pop := popover.NewPopover(
		popover.TriggerWidget(trigger),
		popover.Content(content),
	)
	pop.SetBounds(geometry.NewRect(10, 10, 110, 50))

	ctx := widget.NewContext()
	setupOverlayManager(ctx)

	// Click outside trigger bounds.
	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		geometry.Pt(500, 500), geometry.Pt(500, 500), event.ModNone)
	consumed := pop.Event(ctx, release)
	if consumed {
		t.Error("click outside trigger should not be consumed")
	}
	if pop.IsOpen() {
		t.Error("should not open from click outside")
	}
}

func TestPopover_RightClickIgnored(t *testing.T) {
	trigger := newMockWidget(geometry.NewRect(10, 10, 110, 50))
	content := newMockWidget(geometry.NewRect(0, 0, 100, 80))

	pop := popover.NewPopover(
		popover.TriggerWidget(trigger),
		popover.Content(content),
	)
	pop.SetBounds(geometry.NewRect(10, 10, 110, 50))

	ctx := widget.NewContext()
	setupOverlayManager(ctx)

	// Right-click should not toggle.
	release := event.NewMouseEvent(event.MouseRelease, event.ButtonRight, 0,
		geometry.Pt(50, 30), geometry.Pt(50, 30), event.ModNone)
	pop.Event(ctx, release)

	if pop.IsOpen() {
		t.Error("right-click should not open popover")
	}
}

func TestPopover_DisabledIgnoresEvents(t *testing.T) {
	trigger := newMockWidget(geometry.NewRect(10, 10, 110, 50))
	content := newMockWidget(geometry.NewRect(0, 0, 100, 80))

	pop := popover.NewPopover(
		popover.TriggerWidget(trigger),
		popover.Content(content),
		popover.Disabled(true),
	)
	pop.SetBounds(geometry.NewRect(10, 10, 110, 50))

	ctx := widget.NewContext()
	setupOverlayManager(ctx)

	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		geometry.Pt(50, 30), geometry.Pt(50, 30), event.ModNone)
	consumed := pop.Event(ctx, release)

	if consumed {
		t.Error("disabled popover should not consume events")
	}
	if pop.IsOpen() {
		t.Error("disabled popover should not open")
	}
}

func TestPopover_DisabledFn(t *testing.T) {
	disabled := true
	trigger := newMockWidget(geometry.NewRect(10, 10, 110, 50))
	content := newMockWidget(geometry.NewRect(0, 0, 100, 80))

	pop := popover.NewPopover(
		popover.TriggerWidget(trigger),
		popover.Content(content),
		popover.DisabledFn(func() bool { return disabled }),
	)
	pop.SetBounds(geometry.NewRect(10, 10, 110, 50))

	ctx := widget.NewContext()
	setupOverlayManager(ctx)

	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		geometry.Pt(50, 30), geometry.Pt(50, 30), event.ModNone)
	pop.Event(ctx, release)

	if pop.IsOpen() {
		t.Error("disabled popover should not open")
	}

	disabled = false
	pop.Event(ctx, release)

	if !pop.IsOpen() {
		t.Error("non-disabled popover should open")
	}
}

func TestPopover_KeyboardToggle(t *testing.T) {
	trigger := newMockWidget(geometry.NewRect(10, 10, 110, 50))
	content := newMockWidget(geometry.NewRect(0, 0, 100, 80))

	pop := popover.NewPopover(
		popover.TriggerWidget(trigger),
		popover.Content(content),
	)
	pop.SetBounds(geometry.NewRect(10, 10, 110, 50))
	pop.SetFocused(true)

	ctx := widget.NewContext()
	setupOverlayManager(ctx)

	// Enter key should toggle.
	enterKey := event.NewKeyEvent(event.KeyPress, event.KeyEnter, 0, event.ModNone)
	consumed := pop.Event(ctx, enterKey)
	if !consumed {
		t.Error("Enter key should be consumed")
	}
	if !pop.IsOpen() {
		t.Error("should be open after Enter")
	}

	// Escape key should close.
	escKey := event.NewKeyEvent(event.KeyPress, event.KeyEscape, 0, event.ModNone)
	consumed = pop.Event(ctx, escKey)
	if !consumed {
		t.Error("Escape key should be consumed when open")
	}
	if pop.IsOpen() {
		t.Error("should be closed after Escape")
	}
}

func TestPopover_SpaceKeyToggle(t *testing.T) {
	trigger := newMockWidget(geometry.NewRect(10, 10, 110, 50))
	content := newMockWidget(geometry.NewRect(0, 0, 100, 80))

	pop := popover.NewPopover(
		popover.TriggerWidget(trigger),
		popover.Content(content),
	)
	pop.SetFocused(true)
	pop.SetBounds(geometry.NewRect(10, 10, 110, 50))

	ctx := widget.NewContext()
	setupOverlayManager(ctx)

	space := event.NewKeyEvent(event.KeyPress, event.KeySpace, ' ', event.ModNone)
	consumed := pop.Event(ctx, space)
	if !consumed {
		t.Error("Space key should be consumed")
	}
	if !pop.IsOpen() {
		t.Error("should be open after Space")
	}
}

func TestPopover_EscapeWhenClosedNotConsumed(t *testing.T) {
	pop := popover.NewPopover()
	pop.SetFocused(true)

	ctx := widget.NewContext()

	escKey := event.NewKeyEvent(event.KeyPress, event.KeyEscape, 0, event.ModNone)
	consumed := pop.Event(ctx, escKey)

	if consumed {
		t.Error("Escape when closed should not be consumed")
	}
}

func TestPopover_KeyEventNotFocused(t *testing.T) {
	pop := popover.NewPopover()

	ctx := widget.NewContext()

	enterKey := event.NewKeyEvent(event.KeyPress, event.KeyEnter, 0, event.ModNone)
	consumed := pop.Event(ctx, enterKey)

	if consumed {
		t.Error("key event when not focused should not be consumed")
	}
}

func TestPopover_KeyReleaseIgnored(t *testing.T) {
	pop := popover.NewPopover()
	pop.SetFocused(true)

	ctx := widget.NewContext()

	enterRelease := event.NewKeyEvent(event.KeyRelease, event.KeyEnter, 0, event.ModNone)
	consumed := pop.Event(ctx, enterRelease)

	if consumed {
		t.Error("key release should not be consumed")
	}
}

func TestPopover_UnhandledKeyNotConsumed(t *testing.T) {
	pop := popover.NewPopover()
	pop.SetFocused(true)

	ctx := widget.NewContext()

	tabKey := event.NewKeyEvent(event.KeyPress, event.KeyTab, '\t', event.ModNone)
	consumed := pop.Event(ctx, tabKey)

	if consumed {
		t.Error("unhandled key should not be consumed")
	}
}

// =============================================================================
// Popover Focusable Tests
// =============================================================================

func TestPopover_IsFocusable(t *testing.T) {
	pop := popover.NewPopover()

	if !pop.IsFocusable() {
		t.Error("visible+enabled popover should be focusable")
	}

	pop.SetVisible(false)
	if pop.IsFocusable() {
		t.Error("invisible popover should not be focusable")
	}

	pop.SetVisible(true)
	pop.SetEnabled(false)
	if pop.IsFocusable() {
		t.Error("disabled popover should not be focusable")
	}
}

func TestPopover_IsFocusable_DisabledOpt(t *testing.T) {
	pop := popover.NewPopover(popover.Disabled(true))

	if pop.IsFocusable() {
		t.Error("popover with Disabled(true) should not be focusable")
	}
}

// =============================================================================
// Popover Layout & Draw Tests
// =============================================================================

func TestPopover_Layout_WithTrigger(t *testing.T) {
	trigger := newMockWidget(geometry.NewRect(0, 0, 120, 40))
	pop := popover.NewPopover(
		popover.TriggerWidget(trigger),
	)

	ctx := widget.NewContext()
	size := pop.Layout(ctx, geometry.Loose(geometry.Sz(800, 600)))

	if size.Width != 120 || size.Height != 40 {
		t.Errorf("size = %v, want (120, 40)", size)
	}
}

func TestPopover_Layout_WithoutTrigger(t *testing.T) {
	pop := popover.NewPopover()

	ctx := widget.NewContext()
	size := pop.Layout(ctx, geometry.Loose(geometry.Sz(800, 600)))

	if size.Width != 0 || size.Height != 0 {
		t.Errorf("size = %v, want (0, 0)", size)
	}
}

func TestPopover_Draw_NilCanvas(t *testing.T) {
	pop := popover.NewPopover()
	ctx := widget.NewContext()

	pop.Draw(ctx, nil) // Should not panic.
}

func TestPopover_Draw_WithTrigger(t *testing.T) {
	trigger := &drawRecordingWidget{}
	trigger.SetBounds(geometry.NewRect(10, 10, 110, 50))
	trigger.SetVisible(true)

	pop := popover.NewPopover(
		popover.TriggerWidget(trigger),
	)
	pop.SetBounds(geometry.NewRect(10, 10, 110, 50))

	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	pop.Draw(ctx, canvas)

	if !trigger.drawn {
		t.Error("trigger should have been drawn")
	}
}

// =============================================================================
// Popover Signal Binding Tests
// =============================================================================

func TestPopover_VisibleSignal(t *testing.T) {
	sig := state.NewSignal(false)
	trigger := newMockWidget(geometry.NewRect(10, 10, 110, 50))
	content := newMockWidget(geometry.NewRect(0, 0, 100, 80))

	pop := popover.NewPopover(
		popover.TriggerWidget(trigger),
		popover.Content(content),
		popover.VisibleSignal(sig),
	)
	pop.SetBounds(geometry.NewRect(10, 10, 110, 50))

	ctx := widget.NewContext()
	setupOverlayManager(ctx)

	pop.Show(ctx)

	if !sig.Get() {
		t.Error("signal should be true after Show()")
	}

	pop.Hide(ctx)

	if sig.Get() {
		t.Error("signal should be false after Hide()")
	}
}

func TestPopover_VisibleSignalInit(t *testing.T) {
	sig := state.NewSignal(true)

	pop := popover.NewPopover(
		popover.VisibleSignal(sig),
	)

	if !pop.IsOpen() {
		t.Error("popover should initialize as open from signal=true")
	}
}

func TestPopover_Mount_CreatesBindings(t *testing.T) {
	sig := state.NewSignal(false)
	pop := popover.NewPopover(
		popover.VisibleSignal(sig),
	)

	sched := state.NewScheduler(func(_ []widget.Widget) {})
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	pop.Mount(ctx)

	dirtyCount := 0
	sched.SetOnDirty(func() { dirtyCount++ })
	sig.Set(true)

	if dirtyCount == 0 {
		t.Error("signal change should mark widget dirty after mount")
	}
}

func TestPopover_Unmount_CleansBindings(t *testing.T) {
	sig := state.NewSignal(false)
	pop := popover.NewPopover(
		popover.VisibleSignal(sig),
	)

	sched := state.NewScheduler(func(_ []widget.Widget) {})
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	pop.Mount(ctx)
	pop.CleanupBindings()
	pop.Unmount()

	sig.Set(true)

	if sched.PendingCount() != 0 {
		t.Error("signal change after unmount should not mark widget dirty")
	}
}

func TestPopover_Mount_NoScheduler(t *testing.T) {
	pop := popover.NewPopover()
	ctx := widget.NewContext() // No scheduler.

	pop.Mount(ctx) // Should not panic.
}

func TestPopover_Mount_NoSignal(t *testing.T) {
	pop := popover.NewPopover()

	sched := state.NewScheduler(func(_ []widget.Widget) {})
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	pop.Mount(ctx) // Should not panic.
}

// =============================================================================
// Popover MaxWidth Option Test
// =============================================================================

func TestNewPopover_MaxWidth(t *testing.T) {
	pop := popover.NewPopover(
		popover.MaxWidth(500),
	)
	_ = pop // Verifies construction.
}

// =============================================================================
// Tooltip Construction Tests
// =============================================================================

func TestNewTooltip_Defaults(t *testing.T) {
	tip := popover.NewTooltip()

	if !tip.IsVisible() {
		t.Error("default tooltip should be visible")
	}
	if !tip.IsEnabled() {
		t.Error("default tooltip should be enabled")
	}
	if tip.IsOpen() {
		t.Error("tooltip should not be open by default")
	}
	if tip.Text() != "" {
		t.Errorf("Text() = %q, want empty", tip.Text())
	}
}

func TestNewTooltip_WithText(t *testing.T) {
	tip := popover.NewTooltip(
		popover.TooltipText("Save document"),
	)

	if tip.Text() != "Save document" {
		t.Errorf("Text() = %q, want %q", tip.Text(), "Save document")
	}
}

func TestNewTooltip_WithOptions(t *testing.T) {
	trigger := newMockWidget(geometry.NewRect(10, 10, 110, 50))
	tip := popover.NewTooltip(
		popover.TriggerWidget(trigger),
		popover.TooltipText("Help text"),
		popover.PlacementOpt(popover.Top),
		popover.Delay(300*time.Millisecond),
		popover.Gap(8),
	)

	children := tip.Children()
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}
}

func TestNewTooltip_NoChildren_WithoutTrigger(t *testing.T) {
	tip := popover.NewTooltip()

	children := tip.Children()
	if children != nil {
		t.Errorf("expected nil children without trigger, got %d", len(children))
	}
}

// =============================================================================
// Tooltip Hover Tests
// =============================================================================

func TestTooltip_HoverShowHide(t *testing.T) {
	trigger := newMockWidget(geometry.NewRect(10, 10, 110, 50))
	tip := popover.NewTooltip(
		popover.TriggerWidget(trigger),
		popover.TooltipText("Tip text"),
		popover.Delay(0), // Zero delay for testing.
	)
	tip.SetBounds(geometry.NewRect(10, 10, 110, 50))

	ctx := widget.NewContext()
	now := time.Now()
	ctx.SetNow(now)
	setupOverlayManager(ctx)

	// Mouse enter.
	enter := event.NewMouseEvent(event.MouseEnter, event.ButtonNone, 0,
		geometry.Pt(50, 30), geometry.Pt(50, 30), event.ModNone)
	tip.Event(ctx, enter)

	// Layout triggers show after delay (0ms).
	ctx.SetNow(now.Add(time.Millisecond))
	tip.Layout(ctx, geometry.Loose(geometry.Sz(800, 600)))

	if !tip.IsOpen() {
		t.Error("should be open after hover + delay")
	}

	// Mouse leave.
	leave := event.NewMouseEvent(event.MouseLeave, event.ButtonNone, 0,
		geometry.Pt(200, 200), geometry.Pt(200, 200), event.ModNone)
	tip.Event(ctx, leave)

	if tip.IsOpen() {
		t.Error("should be closed after mouse leave")
	}
}

func TestTooltip_HoverWithDelay(t *testing.T) {
	trigger := newMockWidget(geometry.NewRect(10, 10, 110, 50))
	tip := popover.NewTooltip(
		popover.TriggerWidget(trigger),
		popover.TooltipText("Delayed tip"),
		popover.Delay(100*time.Millisecond),
	)
	tip.SetBounds(geometry.NewRect(10, 10, 110, 50))

	ctx := widget.NewContext()
	now := time.Now()
	ctx.SetNow(now)
	setupOverlayManager(ctx)

	// Mouse enter.
	enter := event.NewMouseEvent(event.MouseEnter, event.ButtonNone, 0,
		geometry.Pt(50, 30), geometry.Pt(50, 30), event.ModNone)
	tip.Event(ctx, enter)

	// Layout before delay elapsed: should not show.
	ctx.SetNow(now.Add(50 * time.Millisecond))
	tip.Layout(ctx, geometry.Loose(geometry.Sz(800, 600)))

	if tip.IsOpen() {
		t.Error("should not be open before delay elapses")
	}

	// Layout after delay elapsed: should show.
	ctx.SetNow(now.Add(150 * time.Millisecond))
	tip.Layout(ctx, geometry.Loose(geometry.Sz(800, 600)))

	if !tip.IsOpen() {
		t.Error("should be open after delay elapses")
	}
}

func TestTooltip_ClickHides(t *testing.T) {
	trigger := newMockWidget(geometry.NewRect(10, 10, 110, 50))
	tip := popover.NewTooltip(
		popover.TriggerWidget(trigger),
		popover.TooltipText("Click me"),
		popover.Delay(0),
	)
	tip.SetBounds(geometry.NewRect(10, 10, 110, 50))

	ctx := widget.NewContext()
	now := time.Now()
	ctx.SetNow(now)
	setupOverlayManager(ctx)

	// Hover to show.
	enter := event.NewMouseEvent(event.MouseEnter, event.ButtonNone, 0,
		geometry.Pt(50, 30), geometry.Pt(50, 30), event.ModNone)
	tip.Event(ctx, enter)
	ctx.SetNow(now.Add(time.Millisecond))
	tip.Layout(ctx, geometry.Loose(geometry.Sz(800, 600)))

	if !tip.IsOpen() {
		t.Fatal("should be open")
	}

	// Click should hide.
	click := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(50, 30), geometry.Pt(50, 30), event.ModNone)
	tip.Event(ctx, click)

	if tip.IsOpen() {
		t.Error("should be closed after click")
	}
}

func TestTooltip_DisabledIgnoresHover(t *testing.T) {
	trigger := newMockWidget(geometry.NewRect(10, 10, 110, 50))
	tip := popover.NewTooltip(
		popover.TriggerWidget(trigger),
		popover.TooltipText("Disabled"),
		popover.Disabled(true),
		popover.Delay(0),
	)
	tip.SetBounds(geometry.NewRect(10, 10, 110, 50))

	ctx := widget.NewContext()
	setupOverlayManager(ctx)

	enter := event.NewMouseEvent(event.MouseEnter, event.ButtonNone, 0,
		geometry.Pt(50, 30), geometry.Pt(50, 30), event.ModNone)
	consumed := tip.Event(ctx, enter)

	if consumed {
		t.Error("disabled tooltip should not consume events")
	}
}

func TestTooltip_ShowNoOpWithoutOverlayManager(t *testing.T) {
	trigger := newMockWidget(geometry.NewRect(10, 10, 110, 50))
	tip := popover.NewTooltip(
		popover.TriggerWidget(trigger),
		popover.TooltipText("No overlay"),
		popover.Delay(0),
	)
	tip.SetBounds(geometry.NewRect(10, 10, 110, 50))

	ctx := widget.NewContext() // No overlay manager.
	now := time.Now()
	ctx.SetNow(now)

	enter := event.NewMouseEvent(event.MouseEnter, event.ButtonNone, 0,
		geometry.Pt(50, 30), geometry.Pt(50, 30), event.ModNone)
	tip.Event(ctx, enter)
	ctx.SetNow(now.Add(time.Millisecond))
	tip.Layout(ctx, geometry.Loose(geometry.Sz(800, 600)))

	if tip.IsOpen() {
		t.Error("should not open without overlay manager")
	}
}

// =============================================================================
// Tooltip Layout & Draw Tests
// =============================================================================

func TestTooltip_Layout_WithTrigger(t *testing.T) {
	trigger := newMockWidget(geometry.NewRect(0, 0, 80, 32))
	tip := popover.NewTooltip(
		popover.TriggerWidget(trigger),
		popover.TooltipText("Test"),
	)

	ctx := widget.NewContext()
	size := tip.Layout(ctx, geometry.Loose(geometry.Sz(800, 600)))

	if size.Width != 80 || size.Height != 32 {
		t.Errorf("size = %v, want (80, 32)", size)
	}
}

func TestTooltip_Layout_WithoutTrigger(t *testing.T) {
	tip := popover.NewTooltip()

	ctx := widget.NewContext()
	size := tip.Layout(ctx, geometry.Loose(geometry.Sz(800, 600)))

	if size.Width != 0 || size.Height != 0 {
		t.Errorf("size = %v, want (0, 0)", size)
	}
}

func TestTooltip_Draw_NilCanvas(t *testing.T) {
	tip := popover.NewTooltip()
	ctx := widget.NewContext()

	tip.Draw(ctx, nil) // Should not panic.
}

func TestTooltip_Draw_WithTrigger(t *testing.T) {
	trigger := &drawRecordingWidget{}
	trigger.SetBounds(geometry.NewRect(10, 10, 80, 40))
	trigger.SetVisible(true)

	tip := popover.NewTooltip(
		popover.TriggerWidget(trigger),
		popover.TooltipText("Test"),
	)

	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	tip.Draw(ctx, canvas)

	if !trigger.drawn {
		t.Error("trigger should have been drawn")
	}
}

// =============================================================================
// Tooltip Signal Binding Tests
// =============================================================================

func TestTooltip_VisibleSignal(t *testing.T) {
	sig := state.NewSignal(false)
	trigger := newMockWidget(geometry.NewRect(10, 10, 110, 50))
	tip := popover.NewTooltip(
		popover.TriggerWidget(trigger),
		popover.TooltipText("Signal tip"),
		popover.VisibleSignal(sig),
		popover.Delay(0),
	)
	tip.SetBounds(geometry.NewRect(10, 10, 110, 50))

	ctx := widget.NewContext()
	now := time.Now()
	ctx.SetNow(now)
	setupOverlayManager(ctx)

	// Hover to show.
	enter := event.NewMouseEvent(event.MouseEnter, event.ButtonNone, 0,
		geometry.Pt(50, 30), geometry.Pt(50, 30), event.ModNone)
	tip.Event(ctx, enter)
	ctx.SetNow(now.Add(time.Millisecond))
	tip.Layout(ctx, geometry.Loose(geometry.Sz(800, 600)))

	if !sig.Get() {
		t.Error("signal should be true when shown")
	}

	// Leave to hide.
	leave := event.NewMouseEvent(event.MouseLeave, event.ButtonNone, 0,
		geometry.Pt(200, 200), geometry.Pt(200, 200), event.ModNone)
	tip.Event(ctx, leave)

	if sig.Get() {
		t.Error("signal should be false when hidden")
	}
}

func TestTooltip_Mount_CreatesBindings(t *testing.T) {
	sig := state.NewSignal(false)
	tip := popover.NewTooltip(
		popover.TooltipText("Test"),
		popover.VisibleSignal(sig),
	)

	sched := state.NewScheduler(func(_ []widget.Widget) {})
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	tip.Mount(ctx)

	dirtyCount := 0
	sched.SetOnDirty(func() { dirtyCount++ })
	sig.Set(true)

	if dirtyCount == 0 {
		t.Error("signal change should mark widget dirty after mount")
	}
}

func TestTooltip_Unmount_CleansBindings(t *testing.T) {
	sig := state.NewSignal(false)
	tip := popover.NewTooltip(
		popover.TooltipText("Test"),
		popover.VisibleSignal(sig),
	)

	sched := state.NewScheduler(func(_ []widget.Widget) {})
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	tip.Mount(ctx)
	tip.CleanupBindings()
	tip.Unmount()

	sig.Set(true)

	if sched.PendingCount() != 0 {
		t.Error("signal change after unmount should not mark widget dirty")
	}
}

func TestTooltip_Mount_NoScheduler(t *testing.T) {
	tip := popover.NewTooltip()
	ctx := widget.NewContext()

	tip.Mount(ctx) // Should not panic.
}

func TestTooltip_OnShowOnHideCallbacks(t *testing.T) {
	showCalled := false
	hideCalled := false

	trigger := newMockWidget(geometry.NewRect(10, 10, 110, 50))
	tip := popover.NewTooltip(
		popover.TriggerWidget(trigger),
		popover.TooltipText("Callbacks"),
		popover.Delay(0),
		popover.OnShow(func() { showCalled = true }),
		popover.OnHide(func() { hideCalled = true }),
	)
	tip.SetBounds(geometry.NewRect(10, 10, 110, 50))

	ctx := widget.NewContext()
	now := time.Now()
	ctx.SetNow(now)
	setupOverlayManager(ctx)

	// Hover to show.
	enter := event.NewMouseEvent(event.MouseEnter, event.ButtonNone, 0,
		geometry.Pt(50, 30), geometry.Pt(50, 30), event.ModNone)
	tip.Event(ctx, enter)
	ctx.SetNow(now.Add(time.Millisecond))
	tip.Layout(ctx, geometry.Loose(geometry.Sz(800, 600)))

	if !showCalled {
		t.Error("OnShow callback should have been called")
	}

	// Leave to hide.
	leave := event.NewMouseEvent(event.MouseLeave, event.ButtonNone, 0,
		geometry.Pt(200, 200), geometry.Pt(200, 200), event.ModNone)
	tip.Event(ctx, leave)

	if !hideCalled {
		t.Error("OnHide callback should have been called")
	}
}

// =============================================================================
// Tooltip MouseMove Event
// =============================================================================

func TestTooltip_MouseMoveNotConsumed(t *testing.T) {
	trigger := newMockWidget(geometry.NewRect(10, 10, 110, 50))
	tip := popover.NewTooltip(
		popover.TriggerWidget(trigger),
		popover.TooltipText("Move"),
	)
	tip.SetBounds(geometry.NewRect(10, 10, 110, 50))

	ctx := widget.NewContext()

	move := event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(50, 30), geometry.Pt(50, 30), event.ModNone)
	consumed := tip.Event(ctx, move)

	if consumed {
		t.Error("mouse move should not be consumed")
	}
}

// =============================================================================
// Painter Tests
// =============================================================================

func TestDefaultPainter_PaintPopover(t *testing.T) {
	painter := popover.DefaultPainter{}
	canvas := &recordingCanvas{}

	painter.PaintPopover(canvas, &popover.PopoverPaintState{
		Bounds:    geometry.NewRect(100, 100, 300, 250),
		Placement: popover.Bottom,
	})

	// Should draw shadow, background, and border.
	if len(canvas.drawRoundRects) < 2 {
		t.Errorf("expected at least 2 round rects (shadow + bg), got %d", len(canvas.drawRoundRects))
	}
}

func TestDefaultPainter_PaintPopover_EmptyBounds(t *testing.T) {
	painter := popover.DefaultPainter{}
	canvas := &recordingCanvas{}

	painter.PaintPopover(canvas, &popover.PopoverPaintState{
		Bounds: geometry.Rect{},
	})

	if len(canvas.drawRoundRects) != 0 {
		t.Error("should not draw anything for empty bounds")
	}
}

func TestDefaultPainter_PaintTooltip(t *testing.T) {
	painter := popover.DefaultPainter{}
	canvas := &recordingCanvas{}

	painter.PaintTooltip(canvas, &popover.TooltipPaintState{
		Bounds:    geometry.NewRect(100, 100, 250, 130),
		Text:      "Help text",
		Placement: popover.Bottom,
	})

	if len(canvas.drawRoundRects) < 1 {
		t.Errorf("expected at least 1 round rect (bg), got %d", len(canvas.drawRoundRects))
	}
	if len(canvas.drawTexts) < 1 {
		t.Error("expected at least 1 text draw call")
	}
	if canvas.drawTexts[0].text != "Help text" {
		t.Errorf("text = %q, want %q", canvas.drawTexts[0].text, "Help text")
	}
}

func TestDefaultPainter_PaintTooltip_EmptyBounds(t *testing.T) {
	painter := popover.DefaultPainter{}
	canvas := &recordingCanvas{}

	painter.PaintTooltip(canvas, &popover.TooltipPaintState{
		Bounds: geometry.Rect{},
		Text:   "Should not draw",
	})

	if len(canvas.drawRoundRects) != 0 {
		t.Error("should not draw anything for empty bounds")
	}
}

func TestCustomPainter(t *testing.T) {
	painter := &testPainter{}
	trigger := newMockWidget(geometry.NewRect(10, 10, 110, 50))
	content := newMockWidget(geometry.NewRect(0, 0, 100, 80))

	pop := popover.NewPopover(
		popover.TriggerWidget(trigger),
		popover.Content(content),
		popover.PainterOpt(painter),
	)
	pop.SetBounds(geometry.NewRect(10, 10, 110, 50))

	ctx := widget.NewContext()
	setupOverlayManager(ctx)

	pop.Show(ctx)
	// Custom painter would be called when the overlay draws.
	_ = pop
}

// =============================================================================
// DismissOnClickOutside Option Test
// =============================================================================

func TestPopover_DismissOnClickOutsideOption(t *testing.T) {
	pop := popover.NewPopover(
		popover.DismissOnClickOutside(false),
	)
	_ = pop // Just verifying construction works.
}

// =============================================================================
// OverlayContent Tests
// =============================================================================

func TestPopover_OverlayContentDrawsBackground(t *testing.T) {
	trigger := newMockWidget(geometry.NewRect(10, 10, 110, 50))
	contentWidget := &drawRecordingWidget{}
	contentWidget.SetBounds(geometry.NewRect(0, 0, 100, 80))
	contentWidget.SetVisible(true)

	pop := popover.NewPopover(
		popover.TriggerWidget(trigger),
		popover.Content(contentWidget),
	)
	pop.SetBounds(geometry.NewRect(10, 10, 110, 50))

	ctx := widget.NewContext()
	om := setupOverlayManager(ctx)

	pop.Show(ctx)

	// The overlay content widget should have been pushed.
	if om.pushCount != 1 {
		t.Errorf("pushCount = %d, want 1", om.pushCount)
	}
}

// =============================================================================
// Interface Compliance
// =============================================================================

func TestPopover_WidgetInterface(t *testing.T) {
	var w widget.Widget = popover.NewPopover()
	_ = w
}

func TestPopover_FocusableInterface(t *testing.T) {
	var f widget.Focusable = popover.NewPopover()
	_ = f
}

func TestPopover_LifecycleInterface(t *testing.T) {
	var l widget.Lifecycle = popover.NewPopover()
	_ = l
}

func TestTooltip_WidgetInterface(t *testing.T) {
	var w widget.Widget = popover.NewTooltip()
	_ = w
}

func TestTooltip_LifecycleInterface(t *testing.T) {
	var l widget.Lifecycle = popover.NewTooltip()
	_ = l
}

// =============================================================================
// Tooltip Non-mouse Events
// =============================================================================

func TestTooltip_NonMouseEventNotConsumed(t *testing.T) {
	tip := popover.NewTooltip()
	ctx := widget.NewContext()

	keyEvent := event.NewKeyEvent(event.KeyPress, event.KeyEnter, 0, event.ModNone)
	consumed := tip.Event(ctx, keyEvent)

	if consumed {
		t.Error("non-mouse event should not be consumed by tooltip")
	}
}

// =============================================================================
// Popover Non-mouse/Non-key Events
// =============================================================================

func TestPopover_WheelEventNotConsumed(t *testing.T) {
	pop := popover.NewPopover()
	ctx := widget.NewContext()

	wheelEvent := &event.WheelEvent{Delta: geometry.Pt(0, 1)}
	consumed := pop.Event(ctx, wheelEvent)

	if consumed {
		t.Error("wheel event should not be consumed by popover")
	}
}

// =============================================================================
// Auto-Flip Coverage: All Sides
// =============================================================================

func TestCalculatePosition_AutoFlipTop(t *testing.T) {
	// Anchor near the top edge; Top should flip to Bottom.
	anchor := geometry.NewRect(100, 0, 100, 20)
	overlaySize := geometry.Sz(80, 30)
	windowSize := geometry.Sz(800, 600)
	gap := float32(4)

	pos := popover.CalculatePosition(popover.Top, anchor, overlaySize, windowSize, gap)
	// Top would be -34, should flip to bottom: 20+4=24
	if pos.Y != 24 {
		t.Errorf("Y = %v, want 24 (flipped to bottom)", pos.Y)
	}
}

func TestCalculatePosition_AutoFlipLeft(t *testing.T) {
	// Anchor near the left edge; Left should flip to Right.
	anchor := geometry.NewRect(0, 100, 20, 40)
	overlaySize := geometry.Sz(80, 30)
	windowSize := geometry.Sz(800, 600)
	gap := float32(4)

	pos := popover.CalculatePosition(popover.Left, anchor, overlaySize, windowSize, gap)
	// Left would be 0-80-4=-84, should flip to right: 20+4=24
	if pos.X != 24 {
		t.Errorf("X = %v, want 24 (flipped to right)", pos.X)
	}
}

func TestCalculatePosition_AutoFlipStartEnd(t *testing.T) {
	// Test various start/end flips.
	tests := []struct {
		name      string
		placement popover.Placement
		anchor    geometry.Rect
	}{
		{"BottomStart_flip", popover.BottomStart, geometry.NewRect(100, 580, 100, 20)},
		{"BottomEnd_flip", popover.BottomEnd, geometry.NewRect(100, 580, 100, 20)},
		{"TopStart_flip", popover.TopStart, geometry.NewRect(100, 0, 100, 20)},
		{"TopEnd_flip", popover.TopEnd, geometry.NewRect(100, 0, 100, 20)},
		{"LeftStart_flip", popover.LeftStart, geometry.NewRect(0, 100, 20, 40)},
		{"LeftEnd_flip", popover.LeftEnd, geometry.NewRect(0, 100, 20, 40)},
		{"RightStart_flip", popover.RightStart, geometry.NewRect(780, 100, 20, 40)},
		{"RightEnd_flip", popover.RightEnd, geometry.NewRect(780, 100, 20, 40)},
	}

	overlaySize := geometry.Sz(80, 30)
	windowSize := geometry.Sz(800, 600)
	gap := float32(4)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pos := popover.CalculatePosition(tc.placement, tc.anchor, overlaySize, windowSize, gap)
			// Just verify no negative values (clamped) and within bounds.
			if pos.X < 0 || pos.Y < 0 {
				t.Errorf("position (%v, %v) should not be negative", pos.X, pos.Y)
			}
			if pos.X+overlaySize.Width > windowSize.Width {
				t.Errorf("X+Width = %v, exceeds window width %v", pos.X+overlaySize.Width, windowSize.Width)
			}
			if pos.Y+overlaySize.Height > windowSize.Height {
				t.Errorf("Y+Height = %v, exceeds window height %v", pos.Y+overlaySize.Height, windowSize.Height)
			}
		})
	}
}

func TestCalculatePosition_ClampAllEdges(t *testing.T) {
	// Overlay bigger than window should be clamped to (0,0).
	anchor := geometry.NewRect(0, 0, 10, 10)
	overlaySize := geometry.Sz(1000, 1000)
	windowSize := geometry.Sz(800, 600)
	gap := float32(0)

	pos := popover.CalculatePosition(popover.Bottom, anchor, overlaySize, windowSize, gap)
	if pos.X != 0 {
		t.Errorf("X = %v, want 0 (clamped)", pos.X)
	}
	if pos.Y != 0 {
		t.Errorf("Y = %v, want 0 (clamped)", pos.Y)
	}
}

// =============================================================================
// Overlay Widget Exercise (via captured overlay)
// =============================================================================

func TestPopover_OverlayWidgetMethods(t *testing.T) {
	trigger := newMockWidget(geometry.NewRect(10, 10, 100, 40))
	contentWidget := &drawRecordingWidget{}
	contentWidget.SetBounds(geometry.NewRect(0, 0, 100, 80))
	contentWidget.SetVisible(true)

	pop := popover.NewPopover(
		popover.TriggerWidget(trigger),
		popover.Content(contentWidget),
	)
	pop.SetBounds(geometry.NewRect(10, 10, 100, 40))

	ctx := widget.NewContext()
	om := &capturingOverlayManager{}
	ctx.SetOverlayManager(om)
	ctx.SetWindowSize(geometry.Sz(800, 600))

	pop.Show(ctx)

	if om.captured == nil {
		t.Fatal("overlay widget should have been captured")
	}

	overlay := om.captured

	// Get bounds via type assertion.
	boundsGetter, ok := overlay.(interface{ Bounds() geometry.Rect })
	if !ok {
		t.Fatal("overlay should support Bounds()")
	}
	bounds := boundsGetter.Bounds()

	// Test Layout.
	size := overlay.Layout(ctx, geometry.Tight(bounds.Size()))
	if size.Width <= 0 || size.Height <= 0 {
		t.Errorf("overlay size = %v, want positive", size)
	}

	// Test Draw.
	canvas := &mockCanvas{}
	overlay.Draw(ctx, canvas)

	// Test Draw with nil canvas.
	overlay.Draw(ctx, nil) // Should not panic.

	// Test Event.
	me := event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(50, 50), geometry.Pt(50, 50), event.ModNone)
	overlay.Event(ctx, me) // Just exercise the path.

	// Test Children.
	children := overlay.Children()
	if len(children) != 1 {
		t.Errorf("expected 1 child, got %d", len(children))
	}
}

func TestTooltip_OverlayWidgetMethods(t *testing.T) {
	trigger := newMockWidget(geometry.NewRect(10, 10, 100, 40))
	tip := popover.NewTooltip(
		popover.TriggerWidget(trigger),
		popover.TooltipText("Test tooltip"),
		popover.Delay(0),
	)
	tip.SetBounds(geometry.NewRect(10, 10, 100, 40))

	ctx := widget.NewContext()
	now := time.Now()
	ctx.SetNow(now)
	om := &capturingOverlayManager{}
	ctx.SetOverlayManager(om)
	ctx.SetWindowSize(geometry.Sz(800, 600))

	// Hover to trigger.
	enter := event.NewMouseEvent(event.MouseEnter, event.ButtonNone, 0,
		geometry.Pt(50, 30), geometry.Pt(50, 30), event.ModNone)
	tip.Event(ctx, enter)
	ctx.SetNow(now.Add(time.Millisecond))
	tip.Layout(ctx, geometry.Loose(geometry.Sz(800, 600)))

	if om.captured == nil {
		t.Fatal("tooltip overlay should have been captured")
	}

	tipOverlay := om.captured

	// Get bounds via type assertion.
	boundsGetter, ok := tipOverlay.(interface{ Bounds() geometry.Rect })
	if !ok {
		t.Fatal("tooltip overlay should support Bounds()")
	}
	bounds := boundsGetter.Bounds()

	// Test Layout.
	size := tipOverlay.Layout(ctx, geometry.Tight(bounds.Size()))
	if size.Width <= 0 || size.Height <= 0 {
		t.Errorf("overlay size = %v, want positive", size)
	}

	// Test Draw.
	canvas := &recordingCanvas{}
	tipOverlay.Draw(ctx, canvas)

	if len(canvas.drawRoundRects) == 0 {
		t.Error("tooltip overlay should have drawn background")
	}
	if len(canvas.drawTexts) == 0 {
		t.Error("tooltip overlay should have drawn text")
	}

	// Test Draw with nil canvas.
	tipOverlay.Draw(ctx, nil) // Should not panic.

	// Test Event (tooltips don't consume events).
	me := event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(50, 50), geometry.Pt(50, 50), event.ModNone)
	consumed := tipOverlay.Event(ctx, me)
	if consumed {
		t.Error("tooltip overlay should not consume events")
	}

	// Test Children.
	children := tipOverlay.Children()
	if children != nil {
		t.Error("tooltip overlay should have no children")
	}
}

// =============================================================================
// Unmount Coverage
// =============================================================================

func TestPopover_Unmount(t *testing.T) {
	pop := popover.NewPopover()
	pop.Unmount() // Should not panic.
}

func TestTooltip_Unmount(t *testing.T) {
	tip := popover.NewTooltip()
	tip.Unmount() // Should not panic.
}

// =============================================================================
// Tooltip Text Clamping
// =============================================================================

func TestTooltip_LongTextClamped(t *testing.T) {
	trigger := newMockWidget(geometry.NewRect(10, 10, 100, 40))
	longText := "This is a very long tooltip text that should be clamped to the maximum width setting"
	tip := popover.NewTooltip(
		popover.TriggerWidget(trigger),
		popover.TooltipText(longText),
		popover.MaxWidth(150),
		popover.Delay(0),
	)
	tip.SetBounds(geometry.NewRect(10, 10, 100, 40))

	ctx := widget.NewContext()
	now := time.Now()
	ctx.SetNow(now)
	om := &capturingOverlayManager{}
	ctx.SetOverlayManager(om)
	ctx.SetWindowSize(geometry.Sz(800, 600))

	enter := event.NewMouseEvent(event.MouseEnter, event.ButtonNone, 0,
		geometry.Pt(50, 30), geometry.Pt(50, 30), event.ModNone)
	tip.Event(ctx, enter)
	ctx.SetNow(now.Add(time.Millisecond))
	tip.Layout(ctx, geometry.Loose(geometry.Sz(800, 600)))

	if !tip.IsOpen() {
		t.Error("tooltip should be open")
	}
}

// =============================================================================
// Popover resolveContentSize partial width/height
// =============================================================================

func TestPopover_ContentSizePartialWidth(t *testing.T) {
	trigger := newMockWidget(geometry.NewRect(10, 10, 100, 40))
	content := newMockWidget(geometry.NewRect(0, 0, 100, 80))

	pop := popover.NewPopover(
		popover.TriggerWidget(trigger),
		popover.Content(content),
		popover.ContentSize(250, 0), // Only width set.
	)
	pop.SetBounds(geometry.NewRect(10, 10, 100, 40))

	ctx := widget.NewContext()
	setupOverlayManager(ctx)

	pop.Show(ctx)
	if !pop.IsOpen() {
		t.Error("should be open")
	}
}

func TestPopover_ContentSizePartialHeight(t *testing.T) {
	trigger := newMockWidget(geometry.NewRect(10, 10, 100, 40))
	content := newMockWidget(geometry.NewRect(0, 0, 100, 80))

	pop := popover.NewPopover(
		popover.TriggerWidget(trigger),
		popover.Content(content),
		popover.ContentSize(0, 150), // Only height set.
	)
	pop.SetBounds(geometry.NewRect(10, 10, 100, 40))

	ctx := widget.NewContext()
	setupOverlayManager(ctx)

	pop.Show(ctx)
	if !pop.IsOpen() {
		t.Error("should be open")
	}
}

// =============================================================================
// Popover with trigger consuming events
// =============================================================================

func TestPopover_TriggerConsumesEvent(t *testing.T) {
	trigger := &consumingWidget{}
	trigger.SetBounds(geometry.NewRect(10, 10, 100, 40))
	trigger.SetVisible(true)

	pop := popover.NewPopover(
		popover.TriggerWidget(trigger),
	)
	pop.SetBounds(geometry.NewRect(10, 10, 100, 40))

	ctx := widget.NewContext()

	me := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		geometry.Pt(50, 30), geometry.Pt(50, 30), event.ModNone)
	consumed := pop.Event(ctx, me)

	if !consumed {
		t.Error("trigger should consume the event")
	}
	if pop.IsOpen() {
		t.Error("popover should not open when trigger consumes event")
	}
}

func TestTooltip_TriggerConsumesEvent(t *testing.T) {
	trigger := &consumingWidget{}
	trigger.SetBounds(geometry.NewRect(10, 10, 100, 40))
	trigger.SetVisible(true)

	tip := popover.NewTooltip(
		popover.TriggerWidget(trigger),
		popover.TooltipText("Test"),
	)
	tip.SetBounds(geometry.NewRect(10, 10, 100, 40))

	ctx := widget.NewContext()

	me := event.NewMouseEvent(event.MouseEnter, event.ButtonNone, 0,
		geometry.Pt(50, 30), geometry.Pt(50, 30), event.ModNone)
	consumed := tip.Event(ctx, me)

	if !consumed {
		t.Error("trigger should consume the event")
	}
}

// =============================================================================
// triggerBoundsOf with nil trigger
// =============================================================================

func TestPopover_NilTriggerBounds(t *testing.T) {
	content := newMockWidget(geometry.NewRect(0, 0, 100, 80))

	pop := popover.NewPopover(
		popover.Content(content),
	)
	pop.SetBounds(geometry.NewRect(50, 50, 100, 40))

	ctx := widget.NewContext()
	setupOverlayManager(ctx)

	// Show with nil trigger should use popover's own bounds.
	pop.Show(ctx)
	if !pop.IsOpen() {
		t.Error("should be open")
	}
}

// =============================================================================
// Tooltip with nil trigger bounds
// =============================================================================

func TestTooltip_NilTriggerBounds(t *testing.T) {
	tip := popover.NewTooltip(
		popover.TooltipText("No trigger"),
		popover.Delay(0),
	)
	tip.SetBounds(geometry.NewRect(50, 50, 100, 40))

	ctx := widget.NewContext()
	now := time.Now()
	ctx.SetNow(now)
	setupOverlayManager(ctx)

	// Directly exercise the show path without trigger.
	enter := event.NewMouseEvent(event.MouseEnter, event.ButtonNone, 0,
		geometry.Pt(80, 70), geometry.Pt(80, 70), event.ModNone)
	tip.Event(ctx, enter)
	ctx.SetNow(now.Add(time.Millisecond))
	tip.Layout(ctx, geometry.Loose(geometry.Sz(800, 600)))

	if !tip.IsOpen() {
		t.Error("should be open")
	}
}

// =============================================================================
// Helpers
// =============================================================================

func setupOverlayManager(ctx *widget.ContextImpl) *mockOverlayManager {
	om := &mockOverlayManager{}
	ctx.SetOverlayManager(om)
	ctx.SetWindowSize(geometry.Sz(800, 600))
	return om
}

// mockOverlayManager implements widget.OverlayManager for testing.
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

// capturingOverlayManager captures the pushed widget for testing.
type capturingOverlayManager struct {
	captured widget.Widget
}

func (m *capturingOverlayManager) PushOverlay(w widget.Widget, _ func()) {
	m.captured = w
}

func (m *capturingOverlayManager) PopOverlay() {}

func (m *capturingOverlayManager) RemoveOverlay(_ widget.Widget) {}

// consumingWidget always consumes events.
type consumingWidget struct {
	widget.WidgetBase
}

func (w *consumingWidget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	return constraints.Constrain(w.Bounds().Size())
}

func (w *consumingWidget) Draw(_ widget.Context, _ widget.Canvas) {}

func (w *consumingWidget) Event(_ widget.Context, _ event.Event) bool {
	return true // Always consumes.
}

func (w *consumingWidget) Children() []widget.Widget {
	return nil
}

// mockWidget is a simple widget for testing.
type mockWidget struct {
	widget.WidgetBase
	layoutSize geometry.Size
}

func newMockWidget(bounds geometry.Rect) *mockWidget {
	w := &mockWidget{
		layoutSize: bounds.Size(),
	}
	w.SetBounds(bounds)
	w.SetVisible(true)
	w.SetEnabled(true)
	return w
}

func (w *mockWidget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	return constraints.Constrain(w.layoutSize)
}

func (w *mockWidget) Draw(_ widget.Context, _ widget.Canvas) {}

func (w *mockWidget) Event(_ widget.Context, _ event.Event) bool {
	return false
}

func (w *mockWidget) Children() []widget.Widget {
	return nil
}

// drawRecordingWidget records whether Draw was called.
type drawRecordingWidget struct {
	widget.WidgetBase
	drawn bool
}

func (w *drawRecordingWidget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	return constraints.Constrain(w.Bounds().Size())
}

func (w *drawRecordingWidget) Draw(_ widget.Context, _ widget.Canvas) {
	w.drawn = true
}

func (w *drawRecordingWidget) Event(_ widget.Context, _ event.Event) bool {
	return false
}

func (w *drawRecordingWidget) Children() []widget.Widget {
	return nil
}

// testPainter records paint calls.
type testPainter struct {
	popoverCalled bool
	tooltipCalled bool
}

func (p *testPainter) PaintPopover(_ widget.Canvas, _ *popover.PopoverPaintState) {
	p.popoverCalled = true
}

func (p *testPainter) PaintTooltip(_ widget.Canvas, _ *popover.TooltipPaintState) {
	p.tooltipCalled = true
}

// Recording canvas for paint verification.
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

// mockCanvas is a no-op canvas for testing.
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
