package chip_test

import (
	"testing"

	"github.com/gogpu/ui/core/chip"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"
)

// --- Construction & defaults ---

func TestNew_Defaults(t *testing.T) {
	c := chip.New()
	if !c.IsVisible() {
		t.Error("new chip should be visible")
	}
	if !c.IsEnabled() {
		t.Error("new chip should be enabled")
	}
	if !c.IsFocusable() {
		t.Error("new chip should be focusable")
	}
	if c.Selected() {
		t.Error("new chip should not be selected")
	}
}

func TestIsFocusable_DisabledNotFocusable(t *testing.T) {
	if chip.New(chip.Disabled(true)).IsFocusable() {
		t.Error("disabled chip should not be focusable")
	}
}

// --- Resolution priority ---

func TestResolvedLabel_Priority(t *testing.T) {
	roSig := state.NewSignal("ro")
	rwSig := state.NewSignal("rw")
	tests := []struct {
		name string
		opts []chip.Option
		want string
	}{
		{"static", []chip.Option{chip.Label("s")}, "s"},
		{"fn over static", []chip.Option{chip.Label("s"), chip.LabelFn(func() string { return "f" })}, "f"},
		{"signal over fn", []chip.Option{chip.LabelFn(func() string { return "f" }), chip.LabelSignal(rwSig)}, "rw"},
		{"readonly over signal", []chip.Option{chip.LabelSignal(rwSig), chip.LabelReadonlySignal(roSig)}, "ro"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := chip.New(tt.opts...)
			canvas := drawAt(c)
			uitest.AssertDrawnText(t, canvas, tt.want)
		})
	}
}

func TestResolvedSelected_Priority(t *testing.T) {
	tests := []struct {
		name string
		opts []chip.Option
		want bool
	}{
		{"static", []chip.Option{chip.Selected(true)}, true},
		{"fn over static", []chip.Option{chip.Selected(false), chip.SelectedFn(func() bool { return true })}, true},
		{"signal over fn", []chip.Option{chip.SelectedFn(func() bool { return false }), chip.SelectedSignal(state.NewSignal(true))}, true},
		{"readonly over signal", []chip.Option{chip.SelectedSignal(state.NewSignal(false)), chip.SelectedReadonlySignal(state.NewSignal(true))}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := chip.New(tt.opts...).Selected(); got != tt.want {
				t.Errorf("Selected() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolvedDisabled_Priority(t *testing.T) {
	// ResolvedDisabled is unexported; assert its effect via IsFocusable
	// (a disabled chip is not focusable).
	tests := []struct {
		name      string
		opts      []chip.Option
		wantFocus bool
	}{
		{"static disabled", []chip.Option{chip.Disabled(true)}, false},
		{"fn over static", []chip.Option{chip.Disabled(false), chip.DisabledFn(func() bool { return true })}, false},
		{"signal over fn", []chip.Option{chip.DisabledFn(func() bool { return false }), chip.DisabledSignal(state.NewSignal(true))}, false},
		{"readonly over signal", []chip.Option{chip.DisabledSignal(state.NewSignal(false)), chip.DisabledReadonlySignal(state.NewSignal(true))}, false},
		{"enabled", []chip.Option{chip.Disabled(false)}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := chip.New(tt.opts...).IsFocusable(); got != tt.wantFocus {
				t.Errorf("IsFocusable() = %v, want %v", got, tt.wantFocus)
			}
		})
	}
}

// --- Layout ---

func TestLayout_WidthGrowsWithLabel(t *testing.T) {
	short := uitest.LayoutWidget(chip.New(chip.Label("a")), 500, 100)
	long := uitest.LayoutWidget(chip.New(chip.Label("a much longer label")), 500, 100)
	if long.Width <= short.Width {
		t.Errorf("long width %v should exceed short width %v", long.Width, short.Width)
	}
}

func TestLayout_MinWidth(t *testing.T) {
	size := uitest.LayoutWidget(chip.New(chip.Label("")), 500, 100)
	if size.Width < 32 {
		t.Errorf("width = %v, want >= 32 (min)", size.Width)
	}
}

func TestLayout_Height(t *testing.T) {
	size := uitest.LayoutWidget(chip.New(chip.Label("x")), 500, 100)
	if size.Height != 32 {
		t.Errorf("height = %v, want 32", size.Height)
	}
}

func TestLayout_PaddingAddsSize(t *testing.T) {
	base := uitest.LayoutWidget(chip.New(chip.Label("x")), 500, 100)
	padded := uitest.LayoutWidget(chip.New(chip.Label("x")).Padding(5), 500, 100)
	if padded.Height != base.Height+10 {
		t.Errorf("padded height = %v, want base %v + 10", padded.Height, base.Height)
	}
}

func TestLayout_RespectsConstraints(t *testing.T) {
	size := uitest.LayoutWidget(chip.New(chip.Label("very long label here")), 20, 20)
	if size.Width > 20 || size.Height > 20 {
		t.Errorf("size = %v, want constrained to 20x20", size)
	}
}

// --- Draw ---

func drawAt(c *chip.Widget) *uitest.MockCanvas {
	c.SetBounds(geometry.NewRect(0, 0, 80, 32))
	return uitest.DrawWidget(c)
}

func TestDraw_UnselectedHasBorder(t *testing.T) {
	c := chip.New(chip.Label("Tag"))
	canvas := drawAt(c)
	if len(canvas.StrokeRoundRects) == 0 {
		t.Error("unselected chip should stroke a border")
	}
	uitest.AssertDrawnText(t, canvas, "Tag")
}

func TestDraw_SelectedFillsBackground(t *testing.T) {
	c := chip.New(chip.Label("Tag"), chip.Selectable(true), chip.Selected(true))
	canvas := drawAt(c)
	if len(canvas.RoundRects) == 0 {
		t.Error("selected chip should fill a background round rect")
	}
}

func TestDraw_DisabledUsesDisabledBackground(t *testing.T) {
	canvas := drawAt(chip.New(chip.Label("Tag"), chip.Disabled(true)))
	if len(canvas.RoundRects) == 0 {
		t.Error("disabled chip should fill a background round rect")
	}
}

func TestDraw_FocusedDrawsRing(t *testing.T) {
	c := chip.New(chip.Label("Tag"))
	c.SetFocused(true)
	canvas := drawAt(c)
	// Unselected draws 1 border stroke; focus adds a 2nd stroke (the ring).
	if len(canvas.StrokeRoundRects) < 2 {
		t.Errorf("focused chip strokes = %d, want >= 2 (border + ring)", len(canvas.StrokeRoundRects))
	}
}

func TestDraw_HoverAddsStateLayer(t *testing.T) {
	c := chip.New(chip.Label("Tag"), chip.Selected(true), chip.Selectable(true))
	base := drawAt(c)
	baseFills := len(base.RoundRects)

	// Hover the chip; the state layer adds an extra filled round rect.
	c.SetBounds(geometry.NewRect(0, 0, 80, 32))
	c.Event(uitest.NewMockContext(), uitest.MouseEnter(10, 10))
	hovered := uitest.DrawWidget(c)
	if len(hovered.RoundRects) <= baseFills {
		t.Errorf("hover fills = %d, want > base %d (state layer)", len(hovered.RoundRects), baseFills)
	}
}

func TestDraw_ColorSchemeOverride(t *testing.T) {
	scheme := chip.ChipColorScheme{
		SelectedBackground: widget.Hex(0x00FF00),
		SelectedLabel:      widget.Hex(0x000000),
	}
	c := chip.New(chip.Label("x"), chip.Selectable(true), chip.Selected(true), chip.ColorSchemeOpt(scheme))
	canvas := drawAt(c)
	if len(canvas.RoundRects) == 0 {
		t.Fatal("expected a filled round rect")
	}
	uitest.AssertColorEqual(t, canvas.RoundRects[0].Color, widget.Hex(0x00FF00))
}

func TestDraw_PaddingInsetsContent(t *testing.T) {
	c := chip.New(chip.Label("x"), chip.Selected(true), chip.Selectable(true)).Padding(4)
	c.SetBounds(geometry.NewRect(0, 0, 100, 40))
	canvas := uitest.DrawWidget(c)
	if len(canvas.RoundRects) == 0 {
		t.Fatal("expected a filled round rect")
	}
	got := canvas.RoundRects[0].Bounds
	want := geometry.NewRect(4, 4, 92, 32)
	if got != want {
		t.Errorf("content bounds = %v, want %v", got, want)
	}
}

// --- Selection state mutation ---

func TestSetSelected_Uncontrolled(t *testing.T) {
	c := chip.New(chip.Selectable(true))
	c.SetSelected(true)
	if !c.Selected() {
		t.Error("SetSelected(true) should select in uncontrolled mode")
	}
}

func TestSetSelected_WritesBackToSignal(t *testing.T) {
	sig := state.NewSignal(false)
	c := chip.New(chip.Selectable(true), chip.SelectedSignal(sig))
	c.SetSelected(true)
	if !sig.Get() {
		t.Error("SetSelected should write back to a bound SelectedSignal")
	}
	if !c.Selected() {
		t.Error("Selected() should reflect the written-back signal value")
	}
}

func TestSetSelected_SameValueNoOp(t *testing.T) {
	c := chip.New(chip.Selectable(true))
	c.SetSelected(false) // already false -> early return, no change
	if c.Selected() {
		t.Error("SetSelected(false) on an unselected chip should be a no-op")
	}
}

func TestSetSelected_ReadonlySourceNoOp(t *testing.T) {
	roSig := state.NewSignal(false)
	c := chip.New(chip.Selectable(true), chip.SelectedReadonlySignal(roSig))
	c.SetSelected(true)
	if c.Selected() {
		t.Error("SetSelected must not change a read-only selected source")
	}

	cFn := chip.New(chip.Selectable(true), chip.SelectedFn(func() bool { return false }))
	cFn.SetSelected(true)
	if cFn.Selected() {
		t.Error("SetSelected must not change a function-driven selected source")
	}
}

// --- Events ---

func TestClick_ActionChipFiresOnClick(t *testing.T) {
	clicked := false
	c := chip.New(chip.Label("Go"), chip.OnClick(func() { clicked = true }))
	c.SetBounds(geometry.NewRect(0, 0, 80, 32))
	if !uitest.SimulateClick(c, 10, 10) {
		t.Error("press should be consumed")
	}
	if !clicked {
		t.Error("OnClick should fire on click")
	}
}

func TestClick_FilterChipTogglesAndNotifies(t *testing.T) {
	var got bool
	var fired int
	c := chip.New(chip.Label("Filter"), chip.Selectable(true),
		chip.OnSelectedChanged(func(sel bool) { got = sel; fired++ }))
	c.SetBounds(geometry.NewRect(0, 0, 80, 32))

	uitest.SimulateClick(c, 10, 10)
	if fired != 1 || !got {
		t.Errorf("after first click: fired=%d got=%v, want 1/true", fired, got)
	}
	if !c.Selected() {
		t.Error("filter chip should be selected after first click")
	}

	uitest.SimulateClick(c, 10, 10)
	if fired != 2 || got {
		t.Errorf("after second click: fired=%d got=%v, want 2/false", fired, got)
	}
	if c.Selected() {
		t.Error("filter chip should be deselected after second click")
	}
}

func TestClick_FilterSignalWriteBack(t *testing.T) {
	sig := state.NewSignal(false)
	c := chip.New(chip.Selectable(true), chip.SelectedSignal(sig))
	c.SetBounds(geometry.NewRect(0, 0, 80, 32))

	uitest.SimulateClick(c, 10, 10)
	if !sig.Get() {
		t.Error("clicking a SelectedSignal-bound chip should write back to the signal")
	}
	if !c.Selected() {
		t.Error("chip should render as selected after write-back")
	}

	uitest.SimulateClick(c, 10, 10)
	if sig.Get() {
		t.Error("a second click should toggle the signal back to false")
	}
}

func TestClick_FilterReadonlyControlled(t *testing.T) {
	// A read-only source: the chip cannot write, but must still report the
	// intended new value via OnSelectedChanged so the owner can update state.
	roSig := state.NewSignal(false)
	var notified bool
	c := chip.New(chip.Selectable(true), chip.SelectedReadonlySignal(roSig),
		chip.OnSelectedChanged(func(sel bool) { notified = sel }))
	c.SetBounds(geometry.NewRect(0, 0, 80, 32))

	uitest.SimulateClick(c, 10, 10)
	if !notified {
		t.Error("OnSelectedChanged should report the intended new state (true)")
	}
	if c.Selected() {
		t.Error("read-only controlled chip must not flip its own state")
	}
}

func TestClick_DisabledIgnored(t *testing.T) {
	clicked := false
	c := chip.New(chip.Disabled(true), chip.OnClick(func() { clicked = true }))
	c.SetBounds(geometry.NewRect(0, 0, 80, 32))
	if uitest.SimulateClick(c, 10, 10) {
		t.Error("disabled chip should not consume events")
	}
	if clicked {
		t.Error("disabled chip should not fire OnClick")
	}
}

func TestMouseRelease_OutsideDoesNotActivate(t *testing.T) {
	clicked := false
	c := chip.New(chip.OnClick(func() { clicked = true }))
	c.SetBounds(geometry.NewRect(0, 0, 80, 32))
	ctx := uitest.NewMockContext()
	c.Event(ctx, uitest.Click(10, 10))     // press inside
	c.Event(ctx, uitest.Release(200, 200)) // release outside
	if clicked {
		t.Error("release outside bounds should not activate")
	}
}

func TestMousePress_NonLeftIgnored(t *testing.T) {
	c := chip.New()
	c.SetBounds(geometry.NewRect(0, 0, 80, 32))
	ctx := uitest.NewMockContext()
	if c.Event(ctx, uitest.RightClick(10, 10)) {
		t.Error("right click press should not be consumed")
	}
}

func TestHover_SetsCursorAndConsumes(t *testing.T) {
	c := chip.New()
	c.SetBounds(geometry.NewRect(0, 0, 80, 32))
	ctx := uitest.NewMockContext()
	if !c.Event(ctx, uitest.MouseEnter(10, 10)) {
		t.Error("MouseEnter should be consumed")
	}
	uitest.AssertCursor(t, ctx, widget.CursorPointer)
	if !c.Event(ctx, uitest.MouseLeave(10, 10)) {
		t.Error("MouseLeave should be consumed")
	}
	uitest.AssertCursor(t, ctx, widget.CursorDefault)
}

func TestKeyboard_ActivatesWhenFocused(t *testing.T) {
	for _, key := range []event.Key{event.KeyEnter, event.KeySpace} {
		clicked := false
		c := chip.New(chip.OnClick(func() { clicked = true }))
		c.SetFocused(true)
		ctx := uitest.NewMockContext()
		c.Event(ctx, uitest.KeyPress(key, event.ModNone))
		c.Event(ctx, uitest.KeyRelease(key, event.ModNone))
		if !clicked {
			t.Errorf("key %v should activate a focused chip", key)
		}
	}
}

func TestKeyboard_ActivationInvalidates(t *testing.T) {
	for _, key := range []event.Key{event.KeyEnter, event.KeySpace} {
		c := chip.New(chip.Selectable(true))
		c.SetBounds(geometry.NewRect(10, 10, 80, 32))
		c.SetFocused(true)

		// Each key edge changes the pressed visual with no mouse event in
		// between, so it must invalidate like the mouse path does.
		for _, ev := range []*event.KeyEvent{
			uitest.KeyPress(key, event.ModNone),
			uitest.KeyRelease(key, event.ModNone),
		} {
			ctx := uitest.NewMockContext()
			c.Event(ctx, ev)
			if len(ctx.InvalidatedRects) == 0 {
				t.Errorf("%v %v should invalidate the chip bounds", key, ev.KeyType)
			}
		}
	}
}

func TestKeyboard_IgnoredWhenNotFocused(t *testing.T) {
	clicked := false
	c := chip.New(chip.OnClick(func() { clicked = true }))
	ctx := uitest.NewMockContext()
	if c.Event(ctx, uitest.KeyPress(event.KeyEnter, event.ModNone)) {
		t.Error("unfocused chip should not consume key events")
	}
	if clicked {
		t.Error("unfocused chip should not activate")
	}
}

func TestChildren_Nil(t *testing.T) {
	if chip.New().Children() != nil {
		t.Error("chip should have no children")
	}
}

// --- Custom painter ---

type recordPainter struct{ last chip.PaintState }

func (p *recordPainter) PaintChip(_ widget.Canvas, ps chip.PaintState) { p.last = ps }

func TestPainterOpt_ReceivesState(t *testing.T) {
	p := &recordPainter{}
	c := chip.New(chip.Label("Z"), chip.Selectable(true), chip.Selected(true), chip.PainterOpt(p))
	drawAt(c)
	if p.last.Label != "Z" || !p.last.Selected || !p.last.Selectable {
		t.Errorf("painter state = %+v, want Label=Z Selected Selectable", p.last)
	}
}

func TestDefaultPainter_EmptyBoundsNoOp(t *testing.T) {
	canvas := &uitest.MockCanvas{}
	chip.DefaultPainter{}.PaintChip(canvas, chip.PaintState{Label: "x"})
	if canvas.TotalDrawCalls() != 0 {
		t.Errorf("empty bounds draw calls = %d, want 0", canvas.TotalDrawCalls())
	}
}

func TestDraw_UsesDefaultFontSize(t *testing.T) {
	canvas := drawAt(chip.New(chip.Label("Tag")))
	if len(canvas.Texts) != 1 {
		t.Fatalf("texts = %d, want 1", len(canvas.Texts))
	}
	if canvas.Texts[0].FontSize != 14 {
		t.Errorf("font size = %v, want 14 (default)", canvas.Texts[0].FontSize)
	}
}

func TestDefaultPainter_RespectsFontSize(t *testing.T) {
	canvas := &uitest.MockCanvas{}
	chip.DefaultPainter{}.PaintChip(canvas, chip.PaintState{
		Label:    "x",
		Bounds:   geometry.NewRect(0, 0, 40, 20),
		FontSize: 22,
	})
	if len(canvas.Texts) != 1 {
		t.Fatalf("texts = %d, want 1", len(canvas.Texts))
	}
	if canvas.Texts[0].FontSize != 22 {
		t.Errorf("font size = %v, want 22 (explicit)", canvas.Texts[0].FontSize)
	}
}

func TestDefaultPainter_FontSizeFallback(t *testing.T) {
	canvas := &uitest.MockCanvas{}
	chip.DefaultPainter{}.PaintChip(canvas, chip.PaintState{
		Label:  "x",
		Bounds: geometry.NewRect(0, 0, 40, 20),
		// FontSize left 0 -> painter falls back to its default.
	})
	if len(canvas.Texts) != 1 {
		t.Fatalf("texts = %d, want 1", len(canvas.Texts))
	}
	if canvas.Texts[0].FontSize != 14 {
		t.Errorf("font size = %v, want 14 (fallback)", canvas.Texts[0].FontSize)
	}
}

// --- Mount / signals ---

type mockScheduler struct{ dirtyCount int }

func (s *mockScheduler) MarkDirty(_ widget.Widget) { s.dirtyCount++ }

func mountWith(t *testing.T, c *chip.Widget) *mockScheduler {
	t.Helper()
	sched := &mockScheduler{}
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)
	c.Mount(ctx)
	return sched
}

func TestMount_LabelSignalNotifies(t *testing.T) {
	sig := state.NewSignal("a")
	c := chip.New(chip.LabelSignal(sig))
	sched := mountWith(t, c)
	sig.Set("b")
	if sched.dirtyCount == 0 {
		t.Error("label change should notify scheduler")
	}
}

func TestMount_ReadonlyLabelSignalNotifies(t *testing.T) {
	sig := state.NewSignal("a")
	c := chip.New(chip.LabelReadonlySignal(sig))
	sched := mountWith(t, c)
	sig.Set("b")
	if sched.dirtyCount == 0 {
		t.Error("readonly label change should notify scheduler")
	}
}

func TestMount_SelectedSignalNotifies(t *testing.T) {
	sig := state.NewSignal(false)
	c := chip.New(chip.Selectable(true), chip.SelectedSignal(sig))
	sched := mountWith(t, c)
	sig.Set(true)
	if sched.dirtyCount == 0 {
		t.Error("selected change should notify scheduler")
	}
}

func TestMount_ReadonlySelectedSignalNotifies(t *testing.T) {
	sig := state.NewSignal(false)
	c := chip.New(chip.Selectable(true), chip.SelectedReadonlySignal(sig))
	sched := mountWith(t, c)
	sig.Set(true)
	if sched.dirtyCount == 0 {
		t.Error("readonly selected change should notify scheduler")
	}
}

func TestMount_DisabledSignalNotifies(t *testing.T) {
	sig := state.NewSignal(false)
	c := chip.New(chip.DisabledSignal(sig))
	sched := mountWith(t, c)
	sig.Set(true)
	if sched.dirtyCount == 0 {
		t.Error("disabled change should notify scheduler")
	}
}

func TestMount_ReadonlyDisabledSignalNotifies(t *testing.T) {
	sig := state.NewSignal(false)
	c := chip.New(chip.DisabledReadonlySignal(sig))
	sched := mountWith(t, c)
	sig.Set(true)
	if sched.dirtyCount == 0 {
		t.Error("readonly disabled change should notify scheduler")
	}
}

func TestMount_NilSchedulerNoPanic(t *testing.T) {
	c := chip.New(chip.LabelSignal(state.NewSignal("a")))
	c.Mount(widget.NewContext()) // scheduler nil; must not panic
}

func TestUnmount_CleanupStopsNotifications(t *testing.T) {
	sig := state.NewSignal("a")
	c := chip.New(chip.LabelSignal(sig))
	sched := mountWith(t, c)
	c.CleanupBindings()
	c.Unmount()
	sched.dirtyCount = 0
	sig.Set("z")
	if sched.dirtyCount > 0 {
		t.Error("scheduler should not be notified after cleanup")
	}
}

// --- Fluent & interface compliance ---

func TestPadding_Chaining(t *testing.T) {
	c := chip.New()
	if c.Padding(2) != c {
		t.Error("Padding should return the same widget")
	}
}

func TestWidget_ImplementsInterfaces(t *testing.T) {
	var _ widget.Widget = chip.New()
	var _ widget.Focusable = chip.New()
	var _ widget.Lifecycle = chip.New()
}

// --- Additional painter / event branch coverage ---

func TestDraw_UnselectedColorSchemeOverride(t *testing.T) {
	scheme := chip.ChipColorScheme{
		Background: widget.Hex(0x112233), // opaque -> triggers fill path
		Border:     widget.Hex(0x445566),
		Label:      widget.Hex(0x778899),
	}
	canvas := drawAt(chip.New(chip.Label("x"), chip.ColorSchemeOpt(scheme)))
	if len(canvas.RoundRects) == 0 {
		t.Fatal("opaque scheme background should produce a fill")
	}
	uitest.AssertColorEqual(t, canvas.RoundRects[0].Color, widget.Hex(0x112233))
	if len(canvas.StrokeRoundRects) == 0 {
		t.Fatal("unselected chip should stroke a border")
	}
	uitest.AssertColorEqual(t, canvas.StrokeRoundRects[0].Color, widget.Hex(0x445566))
}

func TestDraw_DisabledColorSchemeOverride(t *testing.T) {
	scheme := chip.ChipColorScheme{
		DisabledBackground: widget.Hex(0xABCDEF),
		DisabledLabel:      widget.Hex(0x123456),
	}
	canvas := drawAt(chip.New(chip.Label("x"), chip.Disabled(true), chip.ColorSchemeOpt(scheme)))
	if len(canvas.RoundRects) == 0 {
		t.Fatal("disabled chip should fill a background")
	}
	uitest.AssertColorEqual(t, canvas.RoundRects[0].Color, widget.Hex(0xABCDEF))
}

func TestDraw_PressedAddsStateLayer(t *testing.T) {
	c := chip.New(chip.Label("Tag"), chip.Selectable(true), chip.Selected(true))
	base := drawAt(c)
	baseFills := len(base.RoundRects)

	c.SetBounds(geometry.NewRect(0, 0, 80, 32))
	c.Event(uitest.NewMockContext(), uitest.Click(10, 10)) // press -> statePressed
	pressed := uitest.DrawWidget(c)
	if len(pressed.RoundRects) <= baseFills {
		t.Errorf("pressed fills = %d, want > base %d (state layer)", len(pressed.RoundRects), baseFills)
	}
}

func TestEvent_NonMouseNonKeyIgnored(t *testing.T) {
	c := chip.New()
	c.SetBounds(geometry.NewRect(0, 0, 80, 32))
	if c.Event(uitest.NewMockContext(), uitest.WheelScroll(10, 10, 1)) {
		t.Error("wheel events should not be consumed by a chip")
	}
}

func TestKeyboard_OtherKeyIgnored(t *testing.T) {
	c := chip.New()
	c.SetFocused(true)
	if c.Event(uitest.NewMockContext(), uitest.KeyPress(event.KeyTab, event.ModNone)) {
		t.Error("non-activation key should not be consumed")
	}
}

func TestKeyboard_RepeatEventIgnored(t *testing.T) {
	c := chip.New()
	c.SetFocused(true)
	// A key-repeat for Enter is neither an initial press nor a release;
	// handleActivationKey should ignore it.
	repeat := event.NewKeyEvent(event.KeyRepeat, event.KeyEnter, 0, event.ModNone)
	if c.Event(uitest.NewMockContext(), repeat) {
		t.Error("key repeat event should not be consumed")
	}
}
