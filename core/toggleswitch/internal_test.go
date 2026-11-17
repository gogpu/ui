package toggleswitch

import (
	"math"
	"testing"
	"time"

	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/animation"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"
)

// --- helpers ---

// recordingPainter records every PaintState and inherits DefaultPainter
// metrics and motion.
type recordingPainter struct {
	DefaultPainter
	states []PaintState
}

func (p *recordingPainter) PaintSwitch(_ widget.Canvas, s PaintState) {
	p.states = append(p.states, s)
}

func (p *recordingPainter) last(t *testing.T) PaintState {
	t.Helper()
	if len(p.states) == 0 {
		t.Fatal("painter was never called")
	}
	return p.states[len(p.states)-1]
}

var t0 = time.Unix(1000, 0)

// newTestSwitch returns a switch with a recording painter laid out at
// 200x40, and a context whose clock reads t0.
func newTestSwitch(opts ...Option) (*Widget, *recordingPainter, *widget.ContextImpl) {
	p := &recordingPainter{}
	w := New(append([]Option{PainterOpt(p)}, opts...)...)
	w.SetBounds(geometry.NewRect(0, 0, 200, 40))
	ctx := widget.NewContext()
	ctx.SetNow(t0)
	return w, p, ctx
}

func mouse(typ event.MouseEventType, x, y float32) *event.MouseEvent {
	button := event.ButtonNone
	var buttons event.ButtonState
	switch typ {
	case event.MousePress:
		button, buttons = event.ButtonLeft, event.ButtonStateLeft
	case event.MouseRelease:
		button = event.ButtonLeft
	case event.MouseMove:
		buttons = event.ButtonStateLeft
	}
	pos := geometry.Pt(x, y)
	return event.NewMouseEvent(typ, button, buttons, pos, pos, event.ModNone)
}

func click(w *Widget, ctx widget.Context) {
	w.Event(ctx, mouse(event.MousePress, 20, 20))
	w.Event(ctx, mouse(event.MouseRelease, 20, 20))
}

func drawAt(w *Widget, ctx *widget.ContextImpl, at time.Time) {
	ctx.SetNow(at)
	w.Draw(ctx, &uitest.MockCanvas{})
}

func near(a, b float32) bool {
	return math.Abs(float64(a-b)) < 1e-4
}

// --- config ---

func TestConfig_Priorities(t *testing.T) {
	labelSig := state.NewSignal("signal")
	roLabel := state.NewSignal("readonly")
	c := config{label: "static"}
	if got := c.ResolvedLabel(); got != "static" {
		t.Errorf("static label = %q", got)
	}
	c.labelFn = func() string { return "fn" }
	if got := c.ResolvedLabel(); got != "fn" {
		t.Errorf("fn label = %q", got)
	}
	c.labelSignal = labelSig
	if got := c.ResolvedLabel(); got != "signal" {
		t.Errorf("signal label = %q", got)
	}
	c.readonlyLabelSig = roLabel
	if got := c.ResolvedLabel(); got != "readonly" {
		t.Errorf("readonly label = %q", got)
	}

	c.checked = true
	c.checkedFn = func() bool { return false }
	if c.ResolvedChecked() {
		t.Error("CheckedFn should override Checked")
	}
	c.checkedSignal = state.NewSignal(true)
	if !c.ResolvedChecked() {
		t.Error("CheckedSignal should override CheckedFn")
	}

	c.disabled = true
	c.disabledFn = func() bool { return false }
	if c.ResolvedDisabled() {
		t.Error("DisabledFn should override Disabled")
	}
	c.disabledSignal = state.NewSignal(true)
	if !c.ResolvedDisabled() {
		t.Error("DisabledSignal should override DisabledFn")
	}
	c.readonlyDisabledSig = state.NewSignal(false)
	if c.ResolvedDisabled() {
		t.Error("DisabledReadonlySignal should override DisabledSignal")
	}
}

// --- construction and layout ---

func TestOptions_SourcePriorities(t *testing.T) {
	label := state.NewSignal("signal")
	computed := state.NewComputed(func() string { return "computed:" + label.Get() }, label)
	fn := func() string { return "fn" }

	labels := []struct {
		name string
		opts []Option
		want string
	}{
		{"static", []Option{Label("static")}, "static"},
		{"fn over static", []Option{Label("static"), LabelFn(fn)}, "fn"},
		{"signal over fn", []Option{LabelFn(fn), LabelSignal(label)}, "signal"},
		{"readonly signal over signal", []Option{LabelSignal(label), LabelReadonlySignal(computed)}, "computed:signal"},
	}
	for _, tt := range labels {
		if got := New(tt.opts...).AccessibilityLabel(); got != tt.want {
			t.Errorf("%s: label = %q, want %q", tt.name, got, tt.want)
		}
	}

	enabled := func() bool { return false }
	disabled := []struct {
		name string
		opts []Option
		want bool
	}{
		{"fn over static", []Option{Disabled(true), DisabledFn(enabled)}, false},
		{"signal over fn", []Option{DisabledFn(enabled), DisabledSignal(state.NewSignal(true))}, true},
		{"readonly signal over signal", []Option{
			DisabledSignal(state.NewSignal(true)),
			DisabledReadonlySignal(state.NewComputed(enabled)),
		}, false},
	}
	for _, tt := range disabled {
		if got := New(tt.opts...).AccessibilityState().Disabled; got != tt.want {
			t.Errorf("%s: disabled = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestNew_Defaults(t *testing.T) {
	w := New()
	if !w.IsVisible() || !w.IsEnabled() || !w.IsFocusable() {
		t.Error("new switch should be visible, enabled, and focusable")
	}
	if w.IsChecked() {
		t.Error("new switch should be off")
	}
	if _, ok := w.painter.(DefaultPainter); !ok {
		t.Errorf("painter = %T, want DefaultPainter", w.painter)
	}
	if w.Children() != nil {
		t.Error("switch should have no children")
	}
	if got := len(w.GestureHitTest(geometry.Point{})); got != 2 {
		t.Errorf("GestureHitTest returned %d recognizers, want 2", got)
	}
	if !New(Checked(true)).IsChecked() {
		t.Error("Checked(true) should start on")
	}
}

func TestDisabled_NotFocusable(t *testing.T) {
	if New(Disabled(true)).IsFocusable() {
		t.Error("disabled switch should not be focusable")
	}
	w := New()
	w.SetEnabled(false)
	if w.IsFocusable() {
		t.Error("switch with SetEnabled(false) should not be focusable")
	}
}

func TestLayout(t *testing.T) {
	ctx := widget.NewContext()
	loose := geometry.Loose(geometry.Sz(1000, 1000))

	size := New().Layout(ctx, loose)
	want := DefaultPainter{}.SwitchSize()
	if size.Width != want.Width+2*defaultPadding || size.Height != want.Height+2*defaultPadding {
		t.Errorf("size without label = %v, want footprint %v plus padding", size, want)
	}

	labeled := New(Label("Wi-Fi")).Layout(ctx, loose)
	if labeled.Width <= size.Width {
		t.Errorf("label should widen the switch: %v vs %v", labeled.Width, size.Width)
	}

	tight := New(Label("A long label")).Layout(ctx, geometry.Loose(geometry.Sz(40, 20)))
	if tight.Width > 40 || tight.Height > 20 {
		t.Errorf("layout ignored constraints: %v", tight)
	}

	padded := New().Padding(10).Layout(ctx, loose)
	if padded.Width != want.Width+20 {
		t.Errorf("padded width = %v, want %v", padded.Width, want.Width+20)
	}
}

func TestPaintState_Geometry(t *testing.T) {
	w, p, ctx := newTestSwitch(Label("Sound"))
	w.Draw(ctx, &uitest.MockCanvas{})
	s := p.last(t)

	size := DefaultPainter{}.SwitchSize()
	wantControl := geometry.NewRect(defaultPadding, (40-size.Height)/2, size.Width, size.Height)
	if s.ControlBounds != wantControl {
		t.Errorf("ControlBounds = %v, want %v", s.ControlBounds, wantControl)
	}
	if s.LabelBounds.Min.X != wantControl.Max.X+defaultLabelGap {
		t.Errorf("label starts at %v, want %v", s.LabelBounds.Min.X, wantControl.Max.X+defaultLabelGap)
	}
	if s.LabelBounds.Max.X != 200-defaultPadding {
		t.Errorf("label ends at %v, want %v", s.LabelBounds.Max.X, 200-defaultPadding)
	}

	w2, p2, ctx2 := newTestSwitch()
	w2.Draw(ctx2, &uitest.MockCanvas{})
	if !p2.last(t).LabelBounds.IsEmpty() {
		t.Error("LabelBounds should be empty without a label")
	}
}

// --- pointer interaction ---

func TestClick_Toggles(t *testing.T) {
	var got []bool
	w, _, ctx := newTestSwitch(OnToggle(func(on bool) { got = append(got, on) }))

	click(w, ctx)
	click(w, ctx)

	if len(got) != 2 || !got[0] || got[1] {
		t.Errorf("OnToggle values = %v, want [true false]", got)
	}
	if w.IsChecked() {
		t.Error("two clicks should leave the switch off")
	}
}

func TestClick_PressRequestsFocusAndCapture(t *testing.T) {
	w, _, ctx := newTestSwitch()
	var captured widget.Widget
	ctx.SetOnCapturePointer(func(wgt widget.Widget) { captured = wgt })

	if !w.Event(ctx, mouse(event.MousePress, 20, 20)) {
		t.Fatal("press should be consumed")
	}
	if ctx.FocusedWidget() != w {
		t.Error("press should request focus")
	}
	if captured != w {
		t.Error("press should capture the pointer")
	}
}

func TestClick_ReleaseOutsideDoesNotToggle(t *testing.T) {
	w, _, ctx := newTestSwitch()
	w.Event(ctx, mouse(event.MousePress, 20, 20))
	w.Event(ctx, mouse(event.MouseRelease, 20, 300))
	if w.IsChecked() {
		t.Error("release outside should not toggle")
	}
}

func TestClick_RightButtonIgnored(t *testing.T) {
	w, _, ctx := newTestSwitch()
	pos := geometry.Pt(20, 20)
	press := event.NewMouseEvent(event.MousePress, event.ButtonRight, event.ButtonStateRight, pos, pos, event.ModNone)
	release := event.NewMouseEvent(event.MouseRelease, event.ButtonRight, 0, pos, pos, event.ModNone)
	if w.Event(ctx, press) || w.Event(ctx, release) {
		t.Error("right button should not be consumed")
	}
	if w.IsChecked() {
		t.Error("right click should not toggle")
	}
}

func TestDisabled_IgnoresInput(t *testing.T) {
	toggled := false
	w, p, ctx := newTestSwitch(Disabled(true), OnToggle(func(bool) { toggled = true }))

	if w.Event(ctx, mouse(event.MouseEnter, 20, 20)) {
		t.Error("disabled switch should not consume enter")
	}
	if w.Event(ctx, mouse(event.MousePress, 20, 20)) {
		t.Error("disabled switch should not consume press")
	}
	w.Event(ctx, mouse(event.MouseRelease, 20, 20))
	w.SetFocused(true)
	w.Event(ctx, event.NewKeyEvent(event.KeyPress, event.KeySpace, 0, event.ModNone))
	w.Event(ctx, event.NewKeyEvent(event.KeyRelease, event.KeySpace, 0, event.ModNone))
	if toggled {
		t.Error("disabled switch should not toggle")
	}

	w.Draw(ctx, &uitest.MockCanvas{})
	s := p.last(t)
	if !s.Disabled || s.Hovered || s.HoverAmount != 0 || s.PressAmount != 0 {
		t.Errorf("disabled paint state shows interaction: %+v", s)
	}
}

func TestDisabledSignal_HoverReturnsWhenEnabled(t *testing.T) {
	disabled := state.NewSignal(true)
	w, p, ctx := newTestSwitch(DisabledSignal(disabled))
	w.Event(ctx, mouse(event.MouseEnter, 20, 20))
	disabled.Set(false)
	w.Draw(ctx, &uitest.MockCanvas{})
	if !p.last(t).Hovered {
		t.Error("pointer entered while disabled should show hover once enabled")
	}
}

func TestCheckedSignal_TwoWay(t *testing.T) {
	sig := state.NewSignal(false)
	w, _, ctx := newTestSwitch(CheckedSignal(sig))
	click(w, ctx)
	if !sig.Get() {
		t.Error("click should write true to the signal")
	}
	sig.Set(false)
	if w.IsChecked() {
		t.Error("switch should read the signal")
	}
}

func TestCheckedFn_CallerOwnsState(t *testing.T) {
	var toggledTo []bool
	w, _, ctx := newTestSwitch(
		CheckedFn(func() bool { return true }),
		OnToggle(func(on bool) { toggledTo = append(toggledTo, on) }),
	)
	click(w, ctx)
	if len(toggledTo) != 1 || toggledTo[0] {
		t.Errorf("OnToggle = %v, want [false]", toggledTo)
	}
	if w.cfg.checked {
		t.Error("static value should stay untouched with CheckedFn")
	}
}

// --- keyboard ---

func TestKeyboard_SpaceTogglesOnRelease(t *testing.T) {
	w, p, ctx := newTestSwitch()
	w.SetFocused(true)

	if !w.Event(ctx, event.NewKeyEvent(event.KeyPress, event.KeySpace, 0, event.ModNone)) {
		t.Fatal("Space press should be consumed")
	}
	if w.IsChecked() {
		t.Error("Space press alone should not toggle")
	}
	w.Draw(ctx, &uitest.MockCanvas{})
	if !p.last(t).Pressed {
		t.Error("held Space should show the pressed state")
	}

	w.Event(ctx, event.NewKeyEvent(event.KeyRelease, event.KeySpace, 0, event.ModNone))
	if !w.IsChecked() {
		t.Error("Space release should toggle")
	}
}

func TestKeyboard_IgnoredKeysAndFocus(t *testing.T) {
	w, _, ctx := newTestSwitch()
	w.Event(ctx, event.NewKeyEvent(event.KeyPress, event.KeySpace, 0, event.ModNone))
	w.Event(ctx, event.NewKeyEvent(event.KeyRelease, event.KeySpace, 0, event.ModNone))
	if w.IsChecked() {
		t.Error("unfocused switch should ignore Space")
	}

	w.SetFocused(true)
	if w.Event(ctx, event.NewKeyEvent(event.KeyPress, event.KeyEnter, 0, event.ModNone)) {
		t.Error("Enter should not be consumed")
	}
	if w.Event(ctx, event.NewKeyEvent(event.KeyRelease, event.KeySpace, 0, event.ModNone)) {
		t.Error("Space release without press should not be consumed")
	}
	if w.IsChecked() {
		t.Error("stray Space release should not toggle")
	}
}

// --- drag ---

func TestDrag_PastHalfTurnsOn(t *testing.T) {
	w, p, ctx := newTestSwitch()
	travel := DefaultPainter{}.SwitchThumbTravel()

	w.Event(ctx, mouse(event.MousePress, 20, 20))
	w.Event(ctx, mouse(event.MouseMove, 20+travel*0.75, 20))
	w.Draw(ctx, &uitest.MockCanvas{})
	s := p.last(t)
	if !s.Dragging || !near(s.Position, 0.75) {
		t.Errorf("mid-drag state: Dragging=%v Position=%v, want true 0.75", s.Dragging, s.Position)
	}
	if !s.Pressed {
		t.Error("dragging should count as pressed")
	}

	w.Event(ctx, mouse(event.MouseRelease, 20+travel*0.75, 20))
	if !w.IsChecked() {
		t.Error("releasing past the middle should turn the switch on")
	}
}

func TestDrag_BeforeHalfStaysOff(t *testing.T) {
	toggled := false
	w, _, ctx := newTestSwitch(OnToggle(func(bool) { toggled = true }))
	travel := DefaultPainter{}.SwitchThumbTravel()

	w.Event(ctx, mouse(event.MousePress, 20, 20))
	w.Event(ctx, mouse(event.MouseMove, 20+travel*0.4, 20))
	w.Event(ctx, mouse(event.MouseRelease, 20+travel*0.4, 20))
	if toggled || w.IsChecked() {
		t.Error("a drag that stays on the off side should not toggle, even released inside")
	}
}

func TestDrag_FromOnToOff(t *testing.T) {
	w, _, ctx := newTestSwitch(Checked(true))
	w.Draw(ctx, &uitest.MockCanvas{})

	w.Event(ctx, mouse(event.MousePress, 40, 20))
	w.Event(ctx, mouse(event.MouseMove, 10, 20))
	w.Event(ctx, mouse(event.MouseRelease, 10, 20))
	if w.IsChecked() {
		t.Error("dragging the thumb to the off side should turn the switch off")
	}
}

func TestDrag_SmallMoveIsStillAClick(t *testing.T) {
	w, _, ctx := newTestSwitch()
	w.Event(ctx, mouse(event.MousePress, 20, 20))
	w.Event(ctx, mouse(event.MouseMove, 20+dragSlop/2, 20))
	if w.dragging {
		t.Fatal("movement within the slop should not start a drag")
	}
	w.Event(ctx, mouse(event.MouseRelease, 20+dragSlop/2, 20))
	if !w.IsChecked() {
		t.Error("press and release within the slop should toggle like a click")
	}
}

func TestDrag_OvershootOutsideStillCommits(t *testing.T) {
	w, _, ctx := newTestSwitch()
	w.Event(ctx, mouse(event.MousePress, 20, 20))
	w.Event(ctx, mouse(event.MouseMove, 260, 70))
	w.Event(ctx, mouse(event.MouseRelease, 260, 70))
	if !w.IsChecked() {
		t.Error("a drag released past the switch should keep the thumb's side")
	}
}

func TestDrag_CancelReleaseDoesNotCommit(t *testing.T) {
	w, p, ctx := newTestSwitch()
	w.Draw(ctx, &uitest.MockCanvas{})
	w.Event(ctx, mouse(event.MousePress, 20, 20))
	w.Event(ctx, mouse(event.MouseMove, 60, 20))
	if !w.dragging || w.dragPos != 1 {
		t.Fatalf("precondition: dragging=%v dragPos=%v, want true 1", w.dragging, w.dragPos)
	}

	// The Window cancels with a release just outside the widget's corner.
	w.Event(ctx, mouse(event.MouseRelease, -1, -1))
	if w.IsChecked() {
		t.Error("a canceled drag should not toggle")
	}

	// The thumb animates back from where the drag left it.
	drawAt(w, ctx, t0.Add(time.Millisecond))
	s := p.last(t)
	if s.Dragging || s.Position < 0.9 || s.TurningOn {
		t.Errorf("after cancel: Dragging=%v Position=%v TurningOn=%v, want false, ~1, false",
			s.Dragging, s.Position, s.TurningOn)
	}
	drawAt(w, ctx, t0.Add(time.Second))
	if got := p.last(t).Position; got != 0 {
		t.Errorf("thumb settled at %v, want 0", got)
	}
}

func TestDrag_DisabledMidDragDropsGesture(t *testing.T) {
	disabled := state.NewSignal(false)
	w, _, ctx := newTestSwitch(DisabledSignal(disabled))
	w.Event(ctx, mouse(event.MousePress, 20, 20))
	w.Event(ctx, mouse(event.MouseMove, 60, 20))
	disabled.Set(true)
	w.Event(ctx, mouse(event.MouseMove, 70, 20))
	w.Event(ctx, mouse(event.MouseRelease, 70, 20))
	if w.IsChecked() || w.dragging || w.pointerDown {
		t.Error("disabling mid-drag should end the gesture without toggling")
	}
}

// --- animation ---

func TestAnimation_FirstDrawShowsStateWithoutAnimating(t *testing.T) {
	scheduled := 0
	w, p, ctx := newTestSwitch(Checked(true))
	ctx.SetOnScheduleAnimation(func() { scheduled++ })
	w.Draw(ctx, &uitest.MockCanvas{})

	s := p.last(t)
	if s.Position != 1 || s.Progress != 1 || !s.TurningOn {
		t.Errorf("initial on state: Position=%v Progress=%v TurningOn=%v", s.Position, s.Progress, s.TurningOn)
	}
	if scheduled != 0 {
		t.Errorf("idle switch scheduled %d frames", scheduled)
	}
}

func TestAnimation_ToggleAnimatesThumb(t *testing.T) {
	scheduled := 0
	w, p, ctx := newTestSwitch()
	ctx.SetOnScheduleAnimation(func() { scheduled++ })
	w.Draw(ctx, &uitest.MockCanvas{})
	click(w, ctx)

	full := DefaultPainter{}.SwitchMotion().Toggle
	drawAt(w, ctx, t0)
	s := p.last(t)
	if s.Position != 0 || !s.Checked || !s.TurningOn {
		t.Errorf("animation start: Position=%v Checked=%v TurningOn=%v", s.Position, s.Checked, s.TurningOn)
	}
	if scheduled == 0 || !w.NeedsRedraw() {
		t.Error("an animating switch should request another frame")
	}

	drawAt(w, ctx, t0.Add(full/4))
	s = p.last(t)
	if !near(s.Progress, 0.25) || !near(s.Position, animation.EaseInOutCubic(0.25)) {
		t.Errorf("quarter way: Progress=%v Position=%v", s.Progress, s.Position)
	}

	drawAt(w, ctx, t0.Add(full))
	s = p.last(t)
	if s.Position != 1 || s.Progress != 1 {
		t.Errorf("end: Position=%v Progress=%v, want 1", s.Position, s.Progress)
	}
	before := scheduled
	drawAt(w, ctx, t0.Add(full+time.Millisecond))
	if scheduled != before {
		t.Error("a settled switch should stop requesting frames")
	}
}

func TestAnimation_ReverseMidFlightContinues(t *testing.T) {
	w, p, ctx := newTestSwitch()
	w.Draw(ctx, &uitest.MockCanvas{})
	click(w, ctx)
	full := DefaultPainter{}.SwitchMotion().Toggle

	drawAt(w, ctx, t0)
	drawAt(w, ctx, t0.Add(full/2))
	mid := p.last(t).Position

	click(w, ctx)
	drawAt(w, ctx, t0.Add(full/2))
	s := p.last(t)
	if !near(s.Position, mid) {
		t.Errorf("reversing jumped from %v to %v", mid, s.Position)
	}
	if s.TurningOn || s.Checked {
		t.Error("reversed switch should be heading off")
	}

	// The way back covers half the track, so it takes half the time.
	drawAt(w, ctx, t0.Add(full/2+full/2))
	if got := p.last(t).Position; got != 0 {
		t.Errorf("reverse trip ended at %v, want 0", got)
	}
}

func TestAnimation_SignalChangeAnimates(t *testing.T) {
	sig := state.NewSignal(false)
	w, p, ctx := newTestSwitch(CheckedSignal(sig))
	w.Draw(ctx, &uitest.MockCanvas{})

	sig.Set(true)
	drawAt(w, ctx, t0.Add(time.Millisecond))
	if s := p.last(t); s.Position != 0 || !s.TurningOn {
		t.Errorf("external change should start a trip from off: %+v", s)
	}
	drawAt(w, ctx, t0.Add(time.Second))
	if got := p.last(t).Position; got != 1 {
		t.Errorf("Position = %v, want 1", got)
	}
}

func TestAnimation_ZeroClockIsInstant(t *testing.T) {
	w, p, _ := newTestSwitch()
	ctx := uitest.NewMockContext()
	ctx.TimeVal = time.Time{}

	w.Draw(ctx, &uitest.MockCanvas{})
	click(w, ctx)
	w.Event(ctx, mouse(event.MouseEnter, 20, 20))
	ctx.Reset()
	w.ClearRedraw()
	w.Draw(ctx, &uitest.MockCanvas{})

	s := p.last(t)
	if s.Position != 1 || s.HoverAmount != 1 {
		t.Errorf("zero clock: Position=%v HoverAmount=%v, want 1 1", s.Position, s.HoverAmount)
	}
	if len(ctx.InvalidatedRects) != 0 || w.NeedsRedraw() {
		t.Error("zero clock should not request animation frames")
	}
}

func TestAnimation_HoverAndPressFade(t *testing.T) {
	w, p, ctx := newTestSwitch()
	w.Draw(ctx, &uitest.MockCanvas{})
	m := DefaultPainter{}.SwitchMotion()

	w.Event(ctx, mouse(event.MouseEnter, 20, 20))
	drawAt(w, ctx, t0)
	drawAt(w, ctx, t0.Add(m.Hover/2))
	if got := p.last(t).HoverAmount; !near(got, 0.5) {
		t.Errorf("HoverAmount halfway = %v, want 0.5", got)
	}

	w.Event(ctx, mouse(event.MousePress, 20, 20))
	drawAt(w, ctx, t0.Add(m.Hover/2))
	drawAt(w, ctx, t0.Add(m.Hover/2+m.Press))
	s := p.last(t)
	if s.PressAmount != 1 || s.HoverAmount != 1 {
		t.Errorf("held press: PressAmount=%v HoverAmount=%v, want 1 1", s.PressAmount, s.HoverAmount)
	}

	w.Event(ctx, mouse(event.MouseRelease, 20, 20))
	w.Event(ctx, mouse(event.MouseLeave, 300, 20))
	end := t0.Add(m.Hover/2 + m.Press)
	drawAt(w, ctx, end)
	drawAt(w, ctx, end.Add(time.Second))
	s = p.last(t)
	if s.PressAmount != 0 || s.HoverAmount != 0 || s.Hovered || s.Pressed {
		t.Errorf("after leave: %+v", s)
	}
}

func TestAnimation_UnmountResets(t *testing.T) {
	w, p, ctx := newTestSwitch()
	w.Draw(ctx, &uitest.MockCanvas{})
	w.Event(ctx, mouse(event.MouseEnter, 20, 20))
	w.Event(ctx, mouse(event.MousePress, 20, 20))
	w.Unmount()

	drawAt(w, ctx, t0.Add(time.Millisecond))
	s := p.last(t)
	if s.Pressed || s.Hovered || s.PressAmount != 0 || s.HoverAmount != 0 {
		t.Errorf("remounted switch should start at rest: %+v", s)
	}
}

// --- lifecycle ---

func TestMount_BindsSignals(t *testing.T) {
	checked := state.NewSignal(false)
	label := state.NewSignal("label")
	labelBase := state.NewSignal("a")
	roLabel := state.NewComputed(func() string { return labelBase.Get() }, labelBase)
	disabled := state.NewSignal(false)
	disabledBase := state.NewSignal(false)
	roDisabled := state.NewComputed(func() bool { return disabledBase.Get() }, disabledBase)

	tests := []struct {
		name    string
		opts    []Option
		change  func()
		ignored func() // changes a source that a higher-priority one shadows
	}{
		{"checked", []Option{CheckedSignal(checked)}, func() { checked.Set(!checked.Get()) }, nil},
		{"label", []Option{LabelSignal(label)}, func() { label.Set(label.Get() + "!") }, nil},
		{
			"readonly label", []Option{LabelSignal(label), LabelReadonlySignal(roLabel)},
			func() { labelBase.Set(labelBase.Get() + "!") },
			func() { label.Set(label.Get() + "!") },
		},
		{"disabled", []Option{DisabledSignal(disabled)}, func() { disabled.Set(!disabled.Get()) }, nil},
		{
			"readonly disabled", []Option{DisabledSignal(disabled), DisabledReadonlySignal(roDisabled)},
			func() { disabledBase.Set(!disabledBase.Get()) },
			func() { disabled.Set(!disabled.Get()) },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := New(tt.opts...)
			sched := state.NewScheduler(func([]widget.Widget) {})
			ctx := widget.NewContext()
			ctx.SetScheduler(sched)
			w.Mount(ctx)

			if tt.ignored != nil {
				tt.ignored()
				if sched.PendingCount() != 0 {
					t.Error("a shadowed signal should not mark the switch dirty")
				}
			}
			tt.change()
			if sched.PendingCount() != 1 {
				t.Fatalf("pending = %d after a signal change, want 1", sched.PendingCount())
			}
			sched.Flush()

			w.CleanupBindings()
			w.Unmount()
			tt.change()
			if sched.PendingCount() != 0 {
				t.Error("a signal change after unmount should not mark the switch dirty")
			}
		})
	}
}

func TestMount_WithoutSchedulerIsNoop(t *testing.T) {
	sig := state.NewSignal(false)
	w := New(CheckedSignal(sig))
	w.Mount(widget.NewContext())
	sig.Set(true)
	if !w.IsChecked() {
		t.Error("switch should still read the signal without a scheduler")
	}
}

// --- tween ---

func TestTween_ProportionalDuration(t *testing.T) {
	var tw tween
	tw.jump(0.5)
	tw.animateTo(1, t0, 200*time.Millisecond, nil)
	if tw.duration != 100*time.Millisecond {
		t.Errorf("half trip duration = %v, want 100ms", tw.duration)
	}
	if !tw.rising() || !tw.running(t0.Add(50*time.Millisecond)) {
		t.Error("tween should be rising and running mid-trip")
	}
	if got := tw.value(t0.Add(50 * time.Millisecond)); !near(got, 0.75) {
		t.Errorf("value mid-trip = %v, want 0.75", got)
	}
	if tw.running(t0.Add(100 * time.Millisecond)) {
		t.Error("tween should finish after its duration")
	}
	if got := tw.value(t0.Add(-time.Second)); got != 0.5 {
		t.Errorf("value before start = %v, want 0.5", got)
	}

	tw.animateTo(1, t0, time.Second, nil)
	if tw.duration != 100*time.Millisecond {
		t.Error("retargeting to the same value should keep the segment")
	}
}

func TestTween_EasingAndZeroDuration(t *testing.T) {
	var tw tween
	tw.jump(0)
	square := func(f float32) float32 { return f * f }
	tw.animateTo(1, t0, 100*time.Millisecond, square)
	if got := tw.value(t0.Add(50 * time.Millisecond)); !near(got, 0.25) {
		t.Errorf("eased value = %v, want 0.25", got)
	}
	if got := tw.linear(t0.Add(50 * time.Millisecond)); !near(got, 0.5) {
		t.Errorf("linear value = %v, want 0.5", got)
	}

	tw.animateTo(0, t0, 0, nil)
	if tw.running(t0) || tw.value(t0) != 0 {
		t.Error("zero duration should be instant")
	}
}

// --- a11y ---

func TestAccessibility(t *testing.T) {
	w := New(Label("Bluetooth"), A11yHint("Turns Bluetooth on or off"), Checked(true))
	if w.AccessibilityRole() != a11y.RoleSwitch {
		t.Errorf("role = %v, want RoleSwitch", w.AccessibilityRole())
	}
	if w.AccessibilityLabel() != "Bluetooth" || w.AccessibilityHint() != "Turns Bluetooth on or off" {
		t.Error("label or hint not reported")
	}
	if w.AccessibilityValue() != "" {
		t.Error("value should be empty")
	}
	st := w.AccessibilityState()
	if st.Checked != a11y.CheckedTrue || st.Disabled {
		t.Errorf("state = %+v", st)
	}
	if acts := w.AccessibilityActions(); len(acts) != 1 || acts[0] != a11y.ActionClick {
		t.Errorf("actions = %v", acts)
	}

	off := New(Disabled(true))
	if off.AccessibilityState().Checked != a11y.CheckedFalse || !off.AccessibilityState().Disabled {
		t.Error("disabled off switch state not reported")
	}
	if off.AccessibilityActions() != nil {
		t.Error("disabled switch should offer no actions")
	}
}

// --- DefaultPainter ---

func TestDefaultPainter_Draws(t *testing.T) {
	w := New(Label("Airplane mode"), Checked(true))
	w.SetBounds(geometry.NewRect(0, 0, 200, 40))
	w.SetFocused(true)
	canvas := &uitest.MockCanvas{}
	w.Draw(widget.NewContext(), canvas)

	if len(canvas.RoundRects) != 1 || canvas.RoundRects[0].Color != defaultOnTrack {
		t.Errorf("track = %+v, want one on-colored round rect", canvas.RoundRects)
	}
	if len(canvas.Circles) != 1 {
		t.Fatalf("thumb circles = %d, want 1", len(canvas.Circles))
	}
	track := canvas.RoundRects[0].Bounds
	if got, want := canvas.Circles[0].Center.X, track.Max.X-track.Height()/2; !near(got, want) {
		t.Errorf("on thumb center x = %v, want %v", got, want)
	}
	uitest.AssertDrawnText(t, canvas, "Airplane mode")
	if len(canvas.StrokeRoundRects) != 1 {
		t.Error("focused switch should draw a focus ring")
	}
}

func TestDefaultPainter_EmptyControl(t *testing.T) {
	canvas := &uitest.MockCanvas{}
	DefaultPainter{}.PaintSwitch(canvas, PaintState{})
	if canvas.TotalDrawCalls() != 0 {
		t.Error("empty control bounds should draw nothing")
	}
}
