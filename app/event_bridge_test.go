package app

import (
	"testing"

	"github.com/gogpu/gpucontext"
	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

func TestEventBridge_V2IMECommitExactlyOnce(t *testing.T) {
	es := &mockV2EventSource{}
	a := New(WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)

	if es.onIMECompositionStart == nil || es.onIMECompositionUpdateV2 == nil || es.onIMECompositionEnd == nil {
		t.Fatal("versioned IME callbacks were not attached")
	}
	es.onIMECompositionStart()
	composition := gpucontext.IMEComposition{
		CompositionText: "to",
		CursorBegin:     2,
		CursorEnd:       2,
		SelectionStart:  0,
		SelectionEnd:    2,
	}
	es.onIMECompositionUpdateV2(composition)
	es.onIMECompositionEnd("東京")
	// A backend may still echo the result through the legacy text callback;
	// the bridge must drop that echo after routing the explicit end commit.
	es.onTextInput("東京")

	var keyEvents, updates, ends int
	for _, raw := range root.events {
		switch e := raw.(type) {
		case *event.KeyEvent:
			if e.HasRune() {
				keyEvents++
			}
		case *event.IMEEvent:
			if e.IsCompositionUpdate() {
				updates++
			}
			if e.IsCompositionEnd() {
				ends++
				if e.Committed != "東京" {
					t.Fatalf("end commit = %q, want 東京", e.Committed)
				}
			}
		}
	}
	if keyEvents != 0 || updates != 1 || ends != 1 {
		t.Fatalf("IME routed key=%d updates=%d ends=%d events=%#v, want 0/1/1", keyEvents, updates, ends, root.events)
	}
}

func TestEventBridge_V2IMETextInputBeforeEndIsNotDuplicated(t *testing.T) {
	es := &mockV2EventSource{}
	a := New(WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)
	es.onIMECompositionStart()
	es.onTextInput("候補")
	es.onIMECompositionEnd("候補")

	var keyEvents, committedEnds int
	for _, raw := range root.events {
		switch e := raw.(type) {
		case *event.KeyEvent:
			if e.HasRune() {
				keyEvents++
			}
		case *event.IMEEvent:
			if e.IsCompositionEnd() && e.Committed != "" {
				committedEnds++
			}
		}
	}
	if keyEvents != len([]rune("候補")) || committedEnds != 0 {
		t.Fatalf("routed key=%d committed ends=%d events=%#v, want %d/0", keyEvents, committedEnds, root.events, len([]rune("候補")))
	}
}

func TestEventBridge_LegacyIMEUpdateConvertsRangesAndDeduplicates(t *testing.T) {
	es := &mockEventSource{}
	a := New(WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)
	if es.onIMECompositionUpdate == nil {
		t.Fatal("legacy IME update callback was not attached")
	}

	es.onIMECompositionUpdate(gpucontext.IMEState{
		CompositionText: "かな",
		CursorPos:       len("か"),
		SelectionStart:  0,
		SelectionEnd:    len("かな"),
	})
	es.onTextInput("仮名")
	es.onIMECompositionEnd("仮名")

	var updates, keyEvents, committedEnds int
	for _, raw := range root.events {
		switch e := raw.(type) {
		case *event.IMEEvent:
			if e.IsCompositionUpdate() {
				updates++
				if e.Composition.CursorBegin != len("か") || e.Composition.CursorEnd != len("か") {
					t.Fatalf("legacy composition ranges = %#v, want collapsed UTF-8 cursor", e.Composition)
				}
			}
			if e.IsCompositionEnd() && e.Committed != "" {
				committedEnds++
			}
		case *event.KeyEvent:
			if e.HasRune() {
				keyEvents++
			}
		}
	}
	if updates != 1 || keyEvents != len([]rune("仮名")) || committedEnds != 0 {
		t.Fatalf("legacy routed updates=%d keys=%d committed ends=%d events=%#v, want 1/%d/0", updates, keyEvents, committedEnds, root.events, len([]rune("仮名")))
	}
}

func TestWindowIMESyncFocusPrivacyAndCandidateArea(t *testing.T) {
	wp := &mockIMEWindowProvider{mockWindowProvider: mockWindowProvider{width: 500, height: 300, scale: 1}}
	a := New(WithWindowProvider(wp))
	tf := textfield.New(textfield.InitialValue("é你"))
	a.SetRoot(tf)
	a.Window().Context().RequestFocus(tf)
	a.Window().Frame()

	if len(wp.enabled) == 0 || !wp.enabled[len(wp.enabled)-1] {
		t.Fatalf("enabled calls = %#v, want focused text field enabled", wp.enabled)
	}
	if len(wp.surrounding) == 0 || wp.surrounding[len(wp.surrounding)-1].Text != "é你" {
		t.Fatalf("surrounding calls = %#v, want committed UTF-8 text", wp.surrounding)
	}
	if len(wp.areas) == 0 || wp.areas[len(wp.areas)-1].Height <= 0 {
		t.Fatalf("cursor areas = %#v, want non-empty candidate rect", wp.areas)
	}

	beforeSurrounding := len(wp.surrounding)
	a.Window().HandleFocusChange(false)
	if wp.enabled[len(wp.enabled)-1] {
		t.Fatalf("focus loss enabled calls = %#v, want disabled", wp.enabled)
	}
	if len(wp.surrounding) != beforeSurrounding {
		t.Fatalf("focus loss sent surrounding text: %#v", wp.surrounding)
	}

	password := textfield.New(textfield.InitialValue("secret"), textfield.InputTypeOpt(textfield.TypePassword))
	a.Window().Context().ReleaseFocus(tf)
	a.SetRoot(password)
	a.Window().Context().RequestFocus(password)
	a.Window().HandleFocusChange(true)
	a.Window().Frame()
	if wp.enabled[len(wp.enabled)-1] {
		t.Fatal("password field enabled native IME")
	}
	if len(wp.contentTypes) == 0 || wp.contentTypes[len(wp.contentTypes)-1][0] != gpucontext.ContentPurposePassword {
		t.Fatalf("content type calls = %#v, want password purpose", wp.contentTypes)
	}
	if len(wp.surrounding) != beforeSurrounding {
		t.Fatalf("password field sent surrounding text: %#v", wp.surrounding)
	}
}

func newBoundedEventBridgeRoot(es gpucontext.EventSource) *mockWidget {
	wp := &mockWindowProvider{width: 400, height: 300, scale: 1}
	a := New(WithWindowProvider(wp), WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)
	return root
}

func resetEventBridgeRoot(root *mockWidget) {
	root.eventCalled = false
	root.lastEvent = nil
}

func TestEventBridge_MouseMove(t *testing.T) {
	es := &mockEventSource{}
	a := New(WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)

	// Unified pipeline: PointerMove through OnPointer derives MouseMove.
	es.onPointer(gpucontext.PointerEvent{
		Type:        gpucontext.PointerMove,
		X:           100.0,
		Y:           200.0,
		PointerType: gpucontext.PointerTypeMouse,
	})

	if !root.eventCalled {
		t.Fatal("event not dispatched")
	}
	me, ok := root.lastEvent.(*event.MouseEvent)
	if !ok {
		t.Fatal("expected MouseEvent")
	}
	if me.MouseType != event.MouseMove {
		t.Errorf("mouse type = %v, want Move", me.MouseType)
	}
	if me.Position.X != 100.0 || me.Position.Y != 200.0 {
		t.Errorf("position = %v, want (100, 200)", me.Position)
	}
}

func TestEventBridge_MouseMoveOutsideWindow(t *testing.T) {
	es := &mockEventSource{}
	root := newBoundedEventBridgeRoot(es)

	es.onMouseMove(450, 100)

	if root.eventCalled {
		t.Errorf("outside move dispatched %T, want no event", root.lastEvent)
	}
}

func TestEventBridge_PointerLeaveDispatchesMouseLeave(t *testing.T) {
	es := &mockEventSource{}
	root := newBoundedEventBridgeRoot(es)

	// Enter via PointerEnter, then leave via PointerLeave.
	es.onPointer(gpucontext.PointerEvent{
		Type:        gpucontext.PointerEnter,
		PointerType: gpucontext.PointerTypeMouse,
		X:           100,
		Y:           100,
	})
	resetEventBridgeRoot(root)

	es.onPointer(gpucontext.PointerEvent{
		Type:        gpucontext.PointerLeave,
		PointerType: gpucontext.PointerTypeMouse,
		X:           460,
		Y:           100,
	})

	leave, ok := root.lastEvent.(*event.MouseEvent)
	if !ok || leave.MouseType != event.MouseLeave {
		t.Fatalf("PointerLeave event = %T %#v, want MouseLeave", root.lastEvent, root.lastEvent)
	}
}

func TestEventBridge_DragOutsideViaPointerEvents(t *testing.T) {
	es := &mockEventSource{}
	a := New(WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)

	// Drag via PointerDown -> PointerMove -> PointerUp dispatches correctly.
	es.onPointer(gpucontext.PointerEvent{
		Type:        gpucontext.PointerDown,
		X:           100,
		Y:           100,
		PointerType: gpucontext.PointerTypeMouse,
		Button:      gpucontext.ButtonLeft,
		Buttons:     gpucontext.ButtonsLeft,
	})
	resetEventBridgeRoot(root)

	es.onPointer(gpucontext.PointerEvent{
		Type:        gpucontext.PointerMove,
		X:           450,
		Y:           100,
		PointerType: gpucontext.PointerTypeMouse,
		Buttons:     gpucontext.ButtonsLeft,
	})

	move, ok := root.lastEvent.(*event.MouseEvent)
	if !ok || move.MouseType != event.MouseMove {
		t.Fatalf("drag move event = %T %#v, want MouseMove", root.lastEvent, root.lastEvent)
	}
	if !move.Buttons.IsLeftPressed() {
		t.Error("drag move should carry left button state")
	}

	resetEventBridgeRoot(root)
	es.onPointer(gpucontext.PointerEvent{
		Type:        gpucontext.PointerUp,
		X:           450,
		Y:           100,
		PointerType: gpucontext.PointerTypeMouse,
		Button:      gpucontext.ButtonLeft,
		Buttons:     0,
	})
	release, ok := root.lastEvent.(*event.MouseEvent)
	if !ok || release.MouseType != event.MouseRelease {
		t.Fatalf("release event = %T %#v, want MouseRelease", root.lastEvent, root.lastEvent)
	}
}

func TestEventBridge_MousePress(t *testing.T) {
	es := &mockEventSource{}
	a := New(WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)

	// Unified pipeline: PointerDown through OnPointer derives MousePress.
	es.onPointer(gpucontext.PointerEvent{
		Type:        gpucontext.PointerDown,
		X:           50.0,
		Y:           75.0,
		PointerType: gpucontext.PointerTypeMouse,
		Button:      gpucontext.ButtonLeft,
		Buttons:     gpucontext.ButtonsLeft,
	})

	if !root.eventCalled {
		t.Fatal("event not dispatched")
	}
	me, ok := root.lastEvent.(*event.MouseEvent)
	if !ok {
		t.Fatal("expected MouseEvent")
	}
	if me.MouseType != event.MousePress {
		t.Errorf("mouse type = %v, want Press", me.MouseType)
	}
	if me.Button != event.ButtonLeft {
		t.Errorf("button = %v, want Left", me.Button)
	}
	if !me.Buttons.IsLeftPressed() {
		t.Error("left button should be in pressed state")
	}
}

func TestEventBridge_MouseRelease(t *testing.T) {
	es := &mockEventSource{}
	a := New(WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)

	// Unified pipeline: PointerUp through OnPointer derives MouseRelease.
	es.onPointer(gpucontext.PointerEvent{
		Type:        gpucontext.PointerUp,
		X:           30.0,
		Y:           40.0,
		PointerType: gpucontext.PointerTypeMouse,
		Button:      gpucontext.ButtonRight,
		Buttons:     0,
	})

	if !root.eventCalled {
		t.Fatal("event not dispatched")
	}
	me, ok := root.lastEvent.(*event.MouseEvent)
	if !ok {
		t.Fatal("expected MouseEvent")
	}
	if me.MouseType != event.MouseRelease {
		t.Errorf("mouse type = %v, want Release", me.MouseType)
	}
	if me.Button != event.ButtonRight {
		t.Errorf("button = %v, want Right", me.Button)
	}
}

func TestEventBridge_KeyPress(t *testing.T) {
	es := &mockEventSource{}
	a := New(WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)

	es.onKeyPress(gpucontext.KeyA, gpucontext.ModShift|gpucontext.ModControl)

	if !root.eventCalled {
		t.Fatal("event not dispatched")
	}
	ke, ok := root.lastEvent.(*event.KeyEvent)
	if !ok {
		t.Fatal("expected KeyEvent")
	}
	if ke.KeyType != event.KeyPress {
		t.Errorf("key type = %v, want Press", ke.KeyType)
	}
	if ke.Key != event.KeyA {
		t.Errorf("key = %v, want A", ke.Key)
	}
	if !ke.Modifiers().IsShift() {
		t.Error("expected Shift modifier")
	}
	if !ke.Modifiers().IsCtrl() {
		t.Error("expected Ctrl modifier")
	}
}

func TestEventBridge_KeyRelease(t *testing.T) {
	es := &mockEventSource{}
	a := New(WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)

	es.onKeyRelease(gpucontext.KeyEscape, 0)

	if !root.eventCalled {
		t.Fatal("event not dispatched")
	}
	ke, ok := root.lastEvent.(*event.KeyEvent)
	if !ok {
		t.Fatal("expected KeyEvent")
	}
	if ke.KeyType != event.KeyRelease {
		t.Errorf("key type = %v, want Release", ke.KeyType)
	}
	if ke.Key != event.KeyEscape {
		t.Errorf("key = %v, want Escape", ke.Key)
	}
}

func TestEventBridge_Scroll(t *testing.T) {
	es := &mockEventSource{}
	a := New(WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)

	// The basic EventSource scroll callback has no position. Establish the
	// pointer as in-bounds before scrolling.
	es.onMouseMove(100, 100)
	resetEventBridgeRoot(root)
	es.onScroll(0.0, -3.0)

	if !root.eventCalled {
		t.Fatal("event not dispatched")
	}
	we, ok := root.lastEvent.(*event.WheelEvent)
	if !ok {
		t.Fatal("expected WheelEvent")
	}
	if we.Delta.X != 0.0 {
		t.Errorf("delta X = %v, want 0", we.Delta.X)
	}
	if we.Delta.Y != -3.0 {
		t.Errorf("delta Y = %v, want -3", we.Delta.Y)
	}
}

func TestEventBridge_ScrollOutsideWindow_Fallback(t *testing.T) {
	es := &mockEventSource{}
	root := newBoundedEventBridgeRoot(es)

	// Enter the window, then leave. Scrolls after leave should be suppressed.
	es.onPointer(gpucontext.PointerEvent{
		Type: gpucontext.PointerEnter, PointerType: gpucontext.PointerTypeMouse,
		X: 100, Y: 100,
	})
	es.onPointer(gpucontext.PointerEvent{
		Type: gpucontext.PointerLeave, PointerType: gpucontext.PointerTypeMouse,
		X: 450, Y: 100,
	})
	resetEventBridgeRoot(root)
	es.onScroll(0, -3)

	if root.eventCalled {
		t.Errorf("outside scroll dispatched %T, want no event", root.lastEvent)
	}

	// Re-enter the window. Scrolls should be dispatched again.
	es.onPointer(gpucontext.PointerEvent{
		Type: gpucontext.PointerEnter, PointerType: gpucontext.PointerTypeMouse,
		X: 100, Y: 100,
	})
	resetEventBridgeRoot(root)
	es.onScroll(0, -3)
	if _, ok := root.lastEvent.(*event.WheelEvent); !ok {
		t.Fatalf("scroll after re-entry = %T, want WheelEvent", root.lastEvent)
	}
}

func TestEventBridge_DetailedScrollUsesEventPositionAndBounds(t *testing.T) {
	es := &mockScrollEventSource{}
	root := newBoundedEventBridgeRoot(es)

	if es.onScrollEvent == nil {
		t.Fatal("OnScrollEvent callback was not registered")
	}
	if es.onScroll != nil {
		t.Fatal("legacy OnScroll callback registered with detailed source; wheels would dispatch twice")
	}

	// Scroll at a position inside the window bounds (400x300).
	es.onScrollEvent(gpucontext.ScrollEvent{
		X: 120, Y: 130,
		DeltaX: 2, DeltaY: -4,
		Modifiers: gpucontext.ModShift,
	})
	wheel, ok := root.lastEvent.(*event.WheelEvent)
	if !ok {
		t.Fatalf("detailed scroll event = %T, want WheelEvent", root.lastEvent)
	}
	if wheel.Position != geometry.Pt(120, 130) {
		t.Errorf("wheel position = %v, want (120, 130)", wheel.Position)
	}
	if wheel.Delta != geometry.Pt(2, -4) {
		t.Errorf("wheel delta = %v, want (2, -4)", wheel.Delta)
	}
	if !wheel.Modifiers().IsShift() {
		t.Error("detailed wheel lost its Shift modifier")
	}

	// Scroll at a position outside the window bounds should be suppressed
	// when the cursor is not tracked inside.
	resetEventBridgeRoot(root)
	es.onScrollEvent(gpucontext.ScrollEvent{X: 450, Y: 130, DeltaY: -4})
	if root.eventCalled {
		t.Errorf("outside detailed scroll dispatched %T, want no event", root.lastEvent)
	}
}

func TestEventBridge_DetailedScrollFallsBackForUntrustedPosition(t *testing.T) {
	es := &mockScrollEventSource{}
	root := newBoundedEventBridgeRoot(es)

	// Establish the pointer as inside the window via PointerEnter +
	// legacy move tracking (which updates lastMousePos).
	es.onPointer(gpucontext.PointerEvent{
		Type: gpucontext.PointerEnter, PointerType: gpucontext.PointerTypeMouse,
		X: 100, Y: 100,
	})
	es.onMouseMove(100, 100)
	resetEventBridgeRoot(root)

	// A scroll event with an out-of-bounds reported position should fall
	// back to lastMousePos when the cursor is known to be inside.
	es.onScrollEvent(gpucontext.ScrollEvent{X: 1000, Y: 700, DeltaY: -2})
	wheel, ok := root.lastEvent.(*event.WheelEvent)
	if !ok {
		t.Fatalf("out-of-bounds reported position event = %T, want WheelEvent", root.lastEvent)
	}
	if wheel.Position != geometry.Pt(100, 100) {
		t.Errorf("fallback position = %v, want last trusted position (100, 100)", wheel.Position)
	}

	// A scroll event with zero position should also use fallback.
	resetEventBridgeRoot(root)
	es.onScrollEvent(gpucontext.ScrollEvent{DeltaY: -2})
	wheel, ok = root.lastEvent.(*event.WheelEvent)
	if !ok {
		t.Fatalf("zero reported position event = %T, want WheelEvent", root.lastEvent)
	}
	if wheel.Position != geometry.Pt(100, 100) {
		t.Errorf("zero-position fallback = %v, want last trusted position (100, 100)", wheel.Position)
	}

	// After the cursor leaves, an untrusted zero-position scroll should
	// NOT revive a stale in-window position (e.g. macOS momentum scroll).
	es.onPointer(gpucontext.PointerEvent{
		Type: gpucontext.PointerLeave, PointerType: gpucontext.PointerTypeMouse,
		X: 450, Y: 100,
	})
	resetEventBridgeRoot(root)
	es.onScrollEvent(gpucontext.ScrollEvent{DeltaY: -2, IsMomentum: true})
	if root.eventCalled {
		t.Errorf("zero-position momentum after exit dispatched %T, want no event", root.lastEvent)
	}
}

func TestEventBridge_DetailedZeroPositionRequiresInsideState(t *testing.T) {
	es := &mockScrollEventSource{}
	root := newBoundedEventBridgeRoot(es)

	// (0,0) is a real in-window corner, so keep it when PointerEnter has
	// independently established that the cursor is there.
	es.onPointer(gpucontext.PointerEvent{Type: gpucontext.PointerEnter})
	resetEventBridgeRoot(root)
	es.onScrollEvent(gpucontext.ScrollEvent{DeltaY: -2})
	wheel, ok := root.lastEvent.(*event.WheelEvent)
	if !ok {
		t.Fatalf("trusted origin scroll = %T, want WheelEvent", root.lastEvent)
	}
	if !wheel.Position.IsZero() {
		t.Errorf("trusted origin position = %v, want (0,0)", wheel.Position)
	}

	// Once the cursor leaves, the same all-zero event is an untrusted
	// no-position report and must not revive scrolling outside the window.
	es.onPointer(gpucontext.PointerEvent{Type: gpucontext.PointerLeave})
	resetEventBridgeRoot(root)
	es.onScrollEvent(gpucontext.ScrollEvent{DeltaY: -2})
	if root.eventCalled {
		t.Errorf("zero-position scroll after leave dispatched %T, want no event", root.lastEvent)
	}
}

func TestEventBridge_DetailedScrollDuringDragOutside(t *testing.T) {
	es := &mockScrollEventSource{}
	root := newBoundedEventBridgeRoot(es)

	es.onMousePress(gpucontext.MouseButtonLeft, 100, 100)
	resetEventBridgeRoot(root)
	es.onScrollEvent(gpucontext.ScrollEvent{X: 450, Y: 100, DeltaY: -2})

	if _, ok := root.lastEvent.(*event.WheelEvent); !ok {
		t.Fatalf("drag scroll event = %T, want WheelEvent", root.lastEvent)
	}
}

func TestEventBridge_FocusLossInvalidatesFallbackScrollPosition(t *testing.T) {
	es := &mockEventSource{}
	root := newBoundedEventBridgeRoot(es)

	// Establish the cursor as inside via PointerEnter.
	es.onPointer(gpucontext.PointerEvent{
		Type: gpucontext.PointerEnter, PointerType: gpucontext.PointerTypeMouse,
		X: 100, Y: 100,
	})
	es.onFocus(false)
	resetEventBridgeRoot(root)
	es.onScroll(0, -2)

	if root.eventCalled {
		t.Errorf("scroll after focus loss dispatched %T, want no event", root.lastEvent)
	}
}

func TestEventBridge_FocusLossCancelsHeldButtons(t *testing.T) {
	es := &mockScrollEventSource{}
	root := newBoundedEventBridgeRoot(es)

	// Press left button via PointerDown (the unified pipeline dispatch path).
	es.onPointer(gpucontext.PointerEvent{
		Type: gpucontext.PointerDown, PointerType: gpucontext.PointerTypeMouse,
		X: 100, Y: 100,
		Button: gpucontext.ButtonLeft, Buttons: gpucontext.ButtonsLeft,
	})
	es.onFocus(false)
	resetEventBridgeRoot(root)

	// After focus loss, outside scroll should be suppressed
	// (mouseInsideWindow=false, pressedButtons=0).
	es.onScrollEvent(gpucontext.ScrollEvent{X: 450, Y: 100, DeltaY: -2})
	if root.eventCalled {
		t.Errorf("outside scroll after focus loss dispatched %T, want no event", root.lastEvent)
	}

	// A new gesture starts from a clean button state rather than inheriting
	// the lost left-button release. Use PointerDown for dispatch.
	es.onPointer(gpucontext.PointerEvent{
		Type: gpucontext.PointerDown, PointerType: gpucontext.PointerTypeMouse,
		X: 100, Y: 100,
		Button: gpucontext.ButtonRight, Buttons: gpucontext.ButtonsRight,
	})
	press, ok := root.lastEvent.(*event.MouseEvent)
	if !ok {
		t.Fatalf("new press event = %T, want MouseEvent", root.lastEvent)
	}
	if press.Buttons != event.ButtonStateRight {
		t.Errorf("new press buttons = %v, want right only", press.Buttons)
	}
}

func TestEventBridge_PointerCancelCancelsHeldButtonsAndCapture(t *testing.T) {
	es := &mockScrollEventSource{}
	wp := &mockWindowProvider{width: 400, height: 300, scale: 1}
	a := New(WithWindowProvider(wp), WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)
	w := a.Window()

	// Use PointerDown so HandleEvent updates mouseButtonsHeld.
	es.onPointer(gpucontext.PointerEvent{
		Type: gpucontext.PointerDown, PointerType: gpucontext.PointerTypeMouse,
		X: 100, Y: 100,
		Button: gpucontext.ButtonLeft, Buttons: gpucontext.ButtonsLeft,
	})
	w.ctx.CapturePointer(root)
	if w.capturedWidget != root {
		t.Fatal("precondition: root should hold pointer capture")
	}

	// Mouse PointerCancel should clear capture and held buttons.
	es.onPointer(gpucontext.PointerEvent{
		Type:        gpucontext.PointerCancel,
		PointerType: gpucontext.PointerTypeMouse,
	})
	if w.capturedWidget != nil {
		t.Error("capturedWidget should be nil after PointerCancel")
	}
	if w.mouseButtonsHeld != 0 {
		t.Errorf("mouseButtonsHeld = %v after PointerCancel, want none", w.mouseButtonsHeld)
	}
	release, ok := root.lastEvent.(*event.MouseEvent)
	if !ok || release.MouseType != event.MouseRelease || release.Buttons != 0 {
		t.Fatalf("captured widget cancellation event = %T %#v, want final MouseRelease", root.lastEvent, root.lastEvent)
	}

	// After cancel, outside scroll should be suppressed.
	resetEventBridgeRoot(root)
	es.onScrollEvent(gpucontext.ScrollEvent{X: 450, Y: 100, DeltaY: -2})
	if root.eventCalled {
		t.Errorf("outside scroll after PointerCancel dispatched %T, want no event", root.lastEvent)
	}
}

func TestEventBridge_NonMousePointerEventsPreserveMouseState(t *testing.T) {
	es := &mockEventSource{}
	wp := &mockWindowProvider{width: 400, height: 300, scale: 1}
	a := New(WithWindowProvider(wp), WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)
	w := a.Window()

	// Touch/pen PointerEnter should NOT dispatch mouse events or arm scroll.
	for _, pointerType := range []gpucontext.PointerType{
		gpucontext.PointerTypeTouch,
		gpucontext.PointerTypePen,
	} {
		resetEventBridgeRoot(root)
		es.onPointer(gpucontext.PointerEvent{
			Type:        gpucontext.PointerEnter,
			PointerType: pointerType,
			PointerID:   2,
		})
		if root.eventCalled {
			t.Errorf("%v PointerEnter dispatched %T, want no event", pointerType, root.lastEvent)
		}
		es.onScroll(0, -2)
		if root.eventCalled {
			t.Errorf("%v PointerEnter armed mouse scroll fallback", pointerType)
		}
	}

	// Establish mouse as inside via PointerEnter (mouse).
	es.onPointer(gpucontext.PointerEvent{
		Type: gpucontext.PointerEnter, PointerType: gpucontext.PointerTypeMouse,
		X: 100, Y: 100,
	})
	resetEventBridgeRoot(root)

	// Touch/pen PointerLeave should NOT dispatch mouse events.
	for _, pointerType := range []gpucontext.PointerType{
		gpucontext.PointerTypeTouch,
		gpucontext.PointerTypePen,
	} {
		es.onPointer(gpucontext.PointerEvent{
			Type:        gpucontext.PointerLeave,
			PointerType: pointerType,
			PointerID:   2,
		})
	}
	if root.eventCalled {
		t.Errorf("non-mouse enter/leave dispatched %T, want no event", root.lastEvent)
	}

	// The touch/pen leaves must not invalidate the independently tracked
	// in-window mouse position.
	es.onScroll(0, -2)
	if _, ok := root.lastEvent.(*event.WheelEvent); !ok {
		t.Fatalf("mouse scroll after non-mouse leave = %T, want WheelEvent", root.lastEvent)
	}

	// Press mouse button via PointerDown (updates w.mouseButtonsHeld).
	es.onPointer(gpucontext.PointerEvent{
		Type: gpucontext.PointerDown, PointerType: gpucontext.PointerTypeMouse,
		X: 100, Y: 100,
		Button: gpucontext.ButtonLeft, Buttons: gpucontext.ButtonsLeft,
	})
	w.ctx.CapturePointer(root)

	// Touch/pen PointerCancel should NOT clear mouse capture.
	for _, pointerType := range []gpucontext.PointerType{
		gpucontext.PointerTypeTouch,
		gpucontext.PointerTypePen,
	} {
		es.onPointer(gpucontext.PointerEvent{
			Type:        gpucontext.PointerCancel,
			PointerType: pointerType,
			PointerID:   2,
		})
	}

	if w.capturedWidget != root {
		t.Error("touch PointerCancel cleared unrelated mouse capture")
	}
	if w.mouseButtonsHeld != event.ButtonStateLeft {
		t.Errorf("mouseButtonsHeld = %v after touch PointerCancel, want left", w.mouseButtonsHeld)
	}

	// Mouse drag via PointerMove should work normally after touch cancel.
	resetEventBridgeRoot(root)
	es.onPointer(gpucontext.PointerEvent{
		Type: gpucontext.PointerMove, PointerType: gpucontext.PointerTypeMouse,
		X: 450, Y: 100,
		Buttons: gpucontext.ButtonsLeft,
	})
	move, ok := root.lastEvent.(*event.MouseEvent)
	if !ok || move.MouseType != event.MouseMove {
		t.Fatalf("mouse drag after touch PointerCancel = %T, want MouseMove", root.lastEvent)
	}
	if !move.Buttons.IsLeftPressed() {
		t.Error("mouse drag lost left-button state after touch PointerCancel")
	}
}

func TestEventBridge_Resize(t *testing.T) {
	es := &mockEventSource{}
	a := New(WithEventSource(es))

	es.onResize(1920, 1080)

	size := a.Window().WindowSize()
	if size.Width != 1920 || size.Height != 1080 {
		t.Errorf("size = %v, want (1920, 1080)", size)
	}
	if !a.Window().NeedsLayout() {
		t.Error("resize should mark layout as needed")
	}
}

func TestEventBridge_Focus(t *testing.T) {
	es := &mockEventSource{}
	a := New(WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)

	es.onFocus(true)

	if !root.eventCalled {
		t.Fatal("focus event not dispatched")
	}
	fe, ok := root.lastEvent.(*event.FocusEvent)
	if !ok {
		t.Fatal("expected FocusEvent")
	}
	if !fe.IsGained() {
		t.Error("expected focus gained")
	}
}

func TestEventBridge_FocusLost(t *testing.T) {
	es := &mockEventSource{}
	a := New(WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)

	es.onFocus(false)

	if !root.eventCalled {
		t.Fatal("focus event not dispatched")
	}
	fe, ok := root.lastEvent.(*event.FocusEvent)
	if !ok {
		t.Fatal("expected FocusEvent")
	}
	if !fe.IsLost() {
		t.Error("expected focus lost")
	}
}

// --- Translation function tests ---

func TestTranslateMouseButton(t *testing.T) {
	tests := []struct {
		name string
		in   gpucontext.MouseButton
		want event.Button
	}{
		{"Left", gpucontext.MouseButtonLeft, event.ButtonLeft},
		{"Right", gpucontext.MouseButtonRight, event.ButtonRight},
		{"Middle", gpucontext.MouseButtonMiddle, event.ButtonMiddle},
		{"Button4", gpucontext.MouseButton4, event.ButtonX1},
		{"Button5", gpucontext.MouseButton5, event.ButtonX2},
		{"Unknown", gpucontext.MouseButton(99), event.ButtonNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := translateMouseButton(tt.in)
			if got != tt.want {
				t.Errorf("translateMouseButton(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestButtonToState(t *testing.T) {
	tests := []struct {
		name string
		in   event.Button
		want event.ButtonState
	}{
		{"Left", event.ButtonLeft, event.ButtonStateLeft},
		{"Right", event.ButtonRight, event.ButtonStateRight},
		{"Middle", event.ButtonMiddle, event.ButtonStateMiddle},
		{"X1", event.ButtonX1, event.ButtonStateX1},
		{"X2", event.ButtonX2, event.ButtonStateX2},
		{"None", event.ButtonNone, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buttonToState(tt.in)
			if got != tt.want {
				t.Errorf("buttonToState(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestTranslateKey(t *testing.T) {
	tests := []struct {
		name string
		in   gpucontext.Key
		want event.Key
	}{
		// Letters
		{"A", gpucontext.KeyA, event.KeyA},
		{"B", gpucontext.KeyB, event.KeyB},
		{"C", gpucontext.KeyC, event.KeyC},
		{"D", gpucontext.KeyD, event.KeyD},
		{"E", gpucontext.KeyE, event.KeyE},
		{"F", gpucontext.KeyF, event.KeyF},
		{"G", gpucontext.KeyG, event.KeyG},
		{"H", gpucontext.KeyH, event.KeyH},
		{"I", gpucontext.KeyI, event.KeyI},
		{"J", gpucontext.KeyJ, event.KeyJ},
		{"K", gpucontext.KeyK, event.KeyK},
		{"L", gpucontext.KeyL, event.KeyL},
		{"M", gpucontext.KeyM, event.KeyM},
		{"N", gpucontext.KeyN, event.KeyN},
		{"O", gpucontext.KeyO, event.KeyO},
		{"P", gpucontext.KeyP, event.KeyP},
		{"Q", gpucontext.KeyQ, event.KeyQ},
		{"R", gpucontext.KeyR, event.KeyR},
		{"S", gpucontext.KeyS, event.KeyS},
		{"T", gpucontext.KeyT, event.KeyT},
		{"U", gpucontext.KeyU, event.KeyU},
		{"V", gpucontext.KeyV, event.KeyV},
		{"W", gpucontext.KeyW, event.KeyW},
		{"X", gpucontext.KeyX, event.KeyX},
		{"Y", gpucontext.KeyY, event.KeyY},
		{"Z", gpucontext.KeyZ, event.KeyZ},

		// Numbers
		{"0", gpucontext.Key0, event.Key0},
		{"1", gpucontext.Key1, event.Key1},
		{"2", gpucontext.Key2, event.Key2},
		{"3", gpucontext.Key3, event.Key3},
		{"4", gpucontext.Key4, event.Key4},
		{"5", gpucontext.Key5, event.Key5},
		{"6", gpucontext.Key6, event.Key6},
		{"7", gpucontext.Key7, event.Key7},
		{"8", gpucontext.Key8, event.Key8},
		{"9", gpucontext.Key9, event.Key9},

		// Function keys
		{"F1", gpucontext.KeyF1, event.KeyF1},
		{"F2", gpucontext.KeyF2, event.KeyF2},
		{"F3", gpucontext.KeyF3, event.KeyF3},
		{"F4", gpucontext.KeyF4, event.KeyF4},
		{"F5", gpucontext.KeyF5, event.KeyF5},
		{"F6", gpucontext.KeyF6, event.KeyF6},
		{"F7", gpucontext.KeyF7, event.KeyF7},
		{"F8", gpucontext.KeyF8, event.KeyF8},
		{"F9", gpucontext.KeyF9, event.KeyF9},
		{"F10", gpucontext.KeyF10, event.KeyF10},
		{"F11", gpucontext.KeyF11, event.KeyF11},
		{"F12", gpucontext.KeyF12, event.KeyF12},

		// Navigation
		{"Escape", gpucontext.KeyEscape, event.KeyEscape},
		{"Tab", gpucontext.KeyTab, event.KeyTab},
		{"Backspace", gpucontext.KeyBackspace, event.KeyBackspace},
		{"Enter", gpucontext.KeyEnter, event.KeyEnter},
		{"Space", gpucontext.KeySpace, event.KeySpace},
		{"Insert", gpucontext.KeyInsert, event.KeyInsert},
		{"Delete", gpucontext.KeyDelete, event.KeyDelete},
		{"Home", gpucontext.KeyHome, event.KeyHome},
		{"End", gpucontext.KeyEnd, event.KeyEnd},
		{"PageUp", gpucontext.KeyPageUp, event.KeyPageUp},
		{"PageDown", gpucontext.KeyPageDown, event.KeyPageDown},
		{"Left", gpucontext.KeyLeft, event.KeyLeft},
		{"Right", gpucontext.KeyRight, event.KeyRight},
		{"Up", gpucontext.KeyUp, event.KeyUp},
		{"Down", gpucontext.KeyDown, event.KeyDown},

		// Modifiers
		{"LeftShift", gpucontext.KeyLeftShift, event.KeyLeftShift},
		{"RightShift", gpucontext.KeyRightShift, event.KeyRightShift},
		{"LeftControl", gpucontext.KeyLeftControl, event.KeyLeftCtrl},
		{"RightControl", gpucontext.KeyRightControl, event.KeyRightCtrl},
		{"LeftAlt", gpucontext.KeyLeftAlt, event.KeyLeftAlt},
		{"RightAlt", gpucontext.KeyRightAlt, event.KeyRightAlt},
		{"LeftSuper", gpucontext.KeyLeftSuper, event.KeyLeftSuper},
		{"RightSuper", gpucontext.KeyRightSuper, event.KeyRightSuper},

		// Punctuation
		{"Minus", gpucontext.KeyMinus, event.KeyMinus},
		{"Equal", gpucontext.KeyEqual, event.KeyEqual},
		{"LeftBracket", gpucontext.KeyLeftBracket, event.KeyLeftBracket},
		{"RightBracket", gpucontext.KeyRightBracket, event.KeyRightBracket},
		{"Backslash", gpucontext.KeyBackslash, event.KeyBackslash},
		{"Semicolon", gpucontext.KeySemicolon, event.KeySemicolon},
		{"Apostrophe", gpucontext.KeyApostrophe, event.KeyApostrophe},
		{"Grave", gpucontext.KeyGrave, event.KeyGrave},
		{"Comma", gpucontext.KeyComma, event.KeyComma},
		{"Period", gpucontext.KeyPeriod, event.KeyPeriod},
		{"Slash", gpucontext.KeySlash, event.KeySlash},

		// Numpad
		{"Numpad0", gpucontext.KeyNumpad0, event.KeyNumpad0},
		{"Numpad1", gpucontext.KeyNumpad1, event.KeyNumpad1},
		{"Numpad2", gpucontext.KeyNumpad2, event.KeyNumpad2},
		{"Numpad3", gpucontext.KeyNumpad3, event.KeyNumpad3},
		{"Numpad4", gpucontext.KeyNumpad4, event.KeyNumpad4},
		{"Numpad5", gpucontext.KeyNumpad5, event.KeyNumpad5},
		{"Numpad6", gpucontext.KeyNumpad6, event.KeyNumpad6},
		{"Numpad7", gpucontext.KeyNumpad7, event.KeyNumpad7},
		{"Numpad8", gpucontext.KeyNumpad8, event.KeyNumpad8},
		{"Numpad9", gpucontext.KeyNumpad9, event.KeyNumpad9},
		{"NumpadDecimal", gpucontext.KeyNumpadDecimal, event.KeyNumpadDecimal},
		{"NumpadDivide", gpucontext.KeyNumpadDivide, event.KeyNumpadDivide},
		{"NumpadMultiply", gpucontext.KeyNumpadMultiply, event.KeyNumpadMultiply},
		{"NumpadSubtract", gpucontext.KeyNumpadSubtract, event.KeyNumpadSubtract},
		{"NumpadAdd", gpucontext.KeyNumpadAdd, event.KeyNumpadAdd},
		{"NumpadEnter", gpucontext.KeyNumpadEnter, event.KeyNumpadEnter},

		// Other
		{"CapsLock", gpucontext.KeyCapsLock, event.KeyCapsLock},
		{"ScrollLock", gpucontext.KeyScrollLock, event.KeyScrollLock},
		{"NumLock", gpucontext.KeyNumLock, event.KeyNumLock},
		{"PrintScreen", gpucontext.KeyPrintScreen, event.KeyPrintScreen},
		{"Pause", gpucontext.KeyPause, event.KeyPause},

		// Unknown
		{"Unknown", gpucontext.Key(9999), event.KeyUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := translateKey(tt.in)
			if got != tt.want {
				t.Errorf("translateKey(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestTranslateModifiers(t *testing.T) {
	tests := []struct {
		name string
		in   gpucontext.Modifiers
		want event.Modifiers
	}{
		{"None", 0, event.ModNone},
		{"Shift", gpucontext.ModShift, event.ModShift},
		{"Control", gpucontext.ModControl, event.ModCtrl},
		{"Alt", gpucontext.ModAlt, event.ModAlt},
		{"Super", gpucontext.ModSuper, event.ModSuper},
		{"ShiftCtrl", gpucontext.ModShift | gpucontext.ModControl, event.ModShift | event.ModCtrl},
		{"All", gpucontext.ModShift | gpucontext.ModControl | gpucontext.ModAlt | gpucontext.ModSuper,
			event.ModShift | event.ModCtrl | event.ModAlt | event.ModSuper},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := translateModifiers(tt.in)
			if got != tt.want {
				t.Errorf("translateModifiers(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestWidgetCursorToPlatform(t *testing.T) {
	tests := []struct {
		name string
		in   widget.CursorType
		want gpucontext.CursorShape
	}{
		{"Default", widget.CursorDefault, gpucontext.CursorDefault},
		{"Pointer", widget.CursorPointer, gpucontext.CursorPointer},
		{"Text", widget.CursorText, gpucontext.CursorText},
		{"Crosshair", widget.CursorCrosshair, gpucontext.CursorCrosshair},
		{"Move", widget.CursorMove, gpucontext.CursorMove},
		{"ResizeNS", widget.CursorResizeNS, gpucontext.CursorResizeNS},
		{"ResizeEW", widget.CursorResizeEW, gpucontext.CursorResizeEW},
		{"ResizeNESW", widget.CursorResizeNESW, gpucontext.CursorResizeNESW},
		{"ResizeNWSE", widget.CursorResizeNWSE, gpucontext.CursorResizeNWSE},
		{"NotAllowed", widget.CursorNotAllowed, gpucontext.CursorNotAllowed},
		{"Wait", widget.CursorWait, gpucontext.CursorWait},
		{"None", widget.CursorNone, gpucontext.CursorNone},
		{"UnknownFallback", widget.CursorType(99), gpucontext.CursorDefault},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := widgetCursorToPlatform(tt.in)
			if got != tt.want {
				t.Errorf("widgetCursorToPlatform(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestEventBridge_MouseButton_AllVariants(t *testing.T) {
	buttons := []struct {
		name    string
		platBtn gpucontext.Button
		want    event.Button
	}{
		{"Left", gpucontext.ButtonLeft, event.ButtonLeft},
		{"Right", gpucontext.ButtonRight, event.ButtonRight},
		{"Middle", gpucontext.ButtonMiddle, event.ButtonMiddle},
	}

	for _, tt := range buttons {
		t.Run(tt.name, func(t *testing.T) {
			es := &mockEventSource{}
			a := New(WithEventSource(es))
			root := newMockWidget()
			a.SetRoot(root)

			// Unified pipeline: PointerDown with different buttons.
			es.onPointer(gpucontext.PointerEvent{
				Type:        gpucontext.PointerDown,
				X:           10.0,
				Y:           20.0,
				PointerType: gpucontext.PointerTypeMouse,
				Button:      tt.platBtn,
			})

			me, ok := root.lastEvent.(*event.MouseEvent)
			if !ok {
				t.Fatal("expected MouseEvent")
			}
			if me.Button != tt.want {
				t.Errorf("button = %v, want %v", me.Button, tt.want)
			}
		})
	}
}

func TestEventBridge_Resize_WithLayout(t *testing.T) {
	es := &mockEventSource{}
	a := New(WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)

	// Initial frame.
	a.Frame()
	root.layoutCalled = false

	// Resize via event bridge.
	es.onResize(500, 400)

	// Next frame should relayout.
	a.Frame()
	if !root.layoutCalled {
		t.Error("resize via event bridge should trigger relayout")
	}
}

func TestEventBridge_NoEventSource(t *testing.T) {
	// App without event source should work fine.
	a := New()
	root := newMockWidget()
	a.SetRoot(root)
	a.Frame()

	if !root.layoutCalled {
		t.Error("layout should work without event source")
	}
}

func TestSetFrameCallback_NilWindow(t *testing.T) {
	// Test the guard for nil window.
	a := &App{}
	// Should not panic.
	a.SetFrameCallback(func(_ FrameStats) {})
}

// Verify compile-time interface satisfaction for mocks.
var (
	_ gpucontext.WindowProvider     = (*mockWindowProvider)(nil)
	_ gpucontext.PlatformProvider   = (*mockPlatformProvider)(nil)
	_ gpucontext.EventSource        = (*mockEventSource)(nil)
	_ gpucontext.PointerEventSource = (*mockEventSource)(nil)
	_ gpucontext.EventSource        = (*mockScrollEventSource)(nil)
	_ gpucontext.ScrollEventSource  = (*mockScrollEventSource)(nil)
	_ widget.Canvas                 = (*mockCanvas)(nil)
	_ widget.Widget                 = (*mockWidget)(nil)
	_ widget.Widget                 = (*cursorSettingWidget)(nil)
	_ widget.Widget                 = (*cursorSettingOnLayoutWidget)(nil)
)

// --- PointerEventSource tests ---

func TestEventBridge_PointerEnter(t *testing.T) {
	es := &mockEventSource{}
	a := New(WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)

	if es.onPointer == nil {
		t.Fatal("OnPointer was not registered")
	}

	es.onPointer(gpucontext.PointerEvent{
		Type:        gpucontext.PointerEnter,
		X:           100.0,
		Y:           200.0,
		PointerType: gpucontext.PointerTypeMouse,
		IsPrimary:   true,
		Modifiers:   gpucontext.ModShift,
	})

	if !root.eventCalled {
		t.Fatal("PointerEnter event not dispatched")
	}
	me, ok := root.lastEvent.(*event.MouseEvent)
	if !ok {
		t.Fatal("expected MouseEvent")
	}
	if me.MouseType != event.MouseEnter {
		t.Errorf("mouse type = %v, want Enter", me.MouseType)
	}
	if me.Position.X != 100.0 || me.Position.Y != 200.0 {
		t.Errorf("position = %v, want (100, 200)", me.Position)
	}
	if !me.Modifiers().IsShift() {
		t.Error("expected Shift modifier")
	}
}

func TestEventBridge_PointerLeave(t *testing.T) {
	es := &mockEventSource{}
	a := New(WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)

	es.onPointer(gpucontext.PointerEvent{
		Type:        gpucontext.PointerLeave,
		X:           0.0,
		Y:           0.0,
		PointerType: gpucontext.PointerTypeMouse,
		IsPrimary:   true,
	})

	if !root.eventCalled {
		t.Fatal("PointerLeave event not dispatched")
	}
	me, ok := root.lastEvent.(*event.MouseEvent)
	if !ok {
		t.Fatal("expected MouseEvent")
	}
	if me.MouseType != event.MouseLeave {
		t.Errorf("mouse type = %v, want Leave", me.MouseType)
	}
}

func TestEventBridge_PointerMove_DispatchesMouseEvent(t *testing.T) {
	es := &mockEventSource{}
	a := New(WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)

	// In the unified pipeline, PointerMove through OnPointer derives a
	// MouseMove event and dispatches it via HandleEvent.
	es.onPointer(gpucontext.PointerEvent{
		Type:        gpucontext.PointerMove,
		X:           50.0,
		Y:           50.0,
		PointerType: gpucontext.PointerTypeMouse,
	})

	if !root.eventCalled {
		t.Error("PointerMove should dispatch via unified pipeline as derived MouseMove")
	}
	me, ok := root.lastEvent.(*event.MouseEvent)
	if !ok {
		t.Fatal("expected MouseEvent")
	}
	if me.MouseType != event.MouseMove {
		t.Errorf("mouse type = %v, want Move", me.MouseType)
	}
}

func TestEventBridge_PointerEnter_UpdatesLastMousePos(t *testing.T) {
	es := &mockEventSource{}
	a := New(WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)

	// PointerEnter should update lastMousePos so subsequent scroll events
	// carry the correct position.
	es.onPointer(gpucontext.PointerEvent{
		Type: gpucontext.PointerEnter,
		X:    300.0,
		Y:    400.0,
	})

	// Reset event tracking.
	root.eventCalled = false
	root.lastEvent = nil

	// Scroll should use the position from PointerEnter.
	es.onScroll(0.0, -1.0)

	we, ok := root.lastEvent.(*event.WheelEvent)
	if !ok {
		t.Fatal("expected WheelEvent")
	}
	if we.Position.X != 300.0 || we.Position.Y != 400.0 {
		t.Errorf("wheel position = %v, want (300, 400)", we.Position)
	}
}

func TestEventBridge_PointerEventSource_Registration(t *testing.T) {
	es := &mockEventSource{}
	_ = New(WithEventSource(es))

	if es.onPointer == nil {
		t.Error("OnPointer callback was not registered")
	}
}

// Verify no unused imports by using geometry in a test.
var _ = geometry.Pt(0, 0)
