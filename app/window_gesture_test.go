package app

import (
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/gesture"
	"github.com/gogpu/ui/widget"
)

// gestureAwareMock is a widget that implements gesture.GestureAware.
type gestureAwareMock struct {
	widget.WidgetBase
	recognizers   []gesture.Recognizer
	layoutSize    geometry.Size
	layoutCalled  bool
	drawCalled    bool
	eventCalled   bool
	lastEvent     event.Event
	eventCallback func(event.Event) // optional callback for event sequence tracking
}

func newGestureAwareMock(recs ...gesture.Recognizer) *gestureAwareMock {
	m := &gestureAwareMock{
		recognizers: recs,
		layoutSize:  geometry.Sz(100, 100),
	}
	m.SetVisible(true)
	m.SetEnabled(true)
	return m
}

func (m *gestureAwareMock) GestureHitTest(_ geometry.Point) []gesture.Recognizer {
	return m.recognizers
}

func (m *gestureAwareMock) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	m.layoutCalled = true
	return constraints.Constrain(m.layoutSize)
}

func (m *gestureAwareMock) Draw(_ widget.Context, _ widget.Canvas) {
	m.drawCalled = true
}

func (m *gestureAwareMock) Event(_ widget.Context, e event.Event) bool {
	m.eventCalled = true
	m.lastEvent = e
	if m.eventCallback != nil {
		m.eventCallback(e)
	}
	return false
}

// Compile-time check.
var _ gesture.GestureAware = (*gestureAwareMock)(nil)

// --- Tests ---

func TestWindow_HandlePointerEvent_NilRoot(t *testing.T) {
	a := New()
	w := a.Window()

	ev := &gesture.PointerEvent{
		EventType: gesture.PointerDown,
		PointerID: 1,
		Position:  geometry.Pt(50, 50),
	}
	// Should not panic.
	w.HandlePointerEvent(ev)
}

func TestWindow_HandlePointerEvent_NilEvent(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	// Should not panic.
	w.HandlePointerEvent(nil)
}

func TestWindow_GestureArena_LazyCreation(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	// Arena should be nil before any pointer events.
	if w.GestureArena() != nil {
		t.Error("arena should be nil before pointer events")
	}

	// First pointer event creates the arena.
	ev := &gesture.PointerEvent{
		EventType: gesture.PointerDown,
		PointerID: 1,
		Position:  geometry.Pt(50, 50),
	}
	w.HandlePointerEvent(ev)

	if w.GestureArena() == nil {
		t.Error("arena should be created after first pointer event")
	}
}

func TestWindow_HandlePointerEvent_NoGestureAwareWidgets(t *testing.T) {
	a := New()
	w := a.Window()
	// Regular mock widget (not GestureAware).
	root := newMockWidget()
	w.SetRoot(root)
	w.Frame() // layout so ScreenBounds are set

	ev := &gesture.PointerEvent{
		EventType:      gesture.PointerDown,
		PointerID:      1,
		Position:       geometry.Pt(50, 50),
		GlobalPosition: geometry.Pt(50, 50),
	}
	// Should not panic — no recognizers collected, arena closes empty.
	w.HandlePointerEvent(ev)

	// Legacy event dispatch should still work.
	me := event.NewMouseEvent(event.MousePress, event.ButtonLeft,
		event.ButtonStateLeft, geometry.Pt(50, 50), geometry.Pt(50, 50), event.ModNone)
	w.HandleEvent(me)
	if !root.eventCalled {
		t.Error("legacy HandleEvent should still work when no GestureAware widgets")
	}
}

func TestWindow_HandlePointerEvent_GestureAwareWidgetHitTest(t *testing.T) {
	a := New()
	w := a.Window()

	var clickFired bool
	click := gesture.NewClickRecognizer(gesture.ClickConfig{
		OnClick: func(_ gesture.ClickDetails) {
			clickFired = true
		},
	})

	root := newGestureAwareMock(click)
	root.layoutSize = geometry.Sz(200, 200)
	w.SetRoot(root)
	w.Frame() // layout

	// Set bounds and screen origin for hit-testing.
	// In headless mode Frame() performs layout but not Draw, so
	// ScreenOrigin (set during Draw) must be manually applied.
	root.SetBounds(geometry.NewRect(0, 0, 200, 200))
	root.SetScreenOrigin(geometry.Pt(0, 0))

	// PointerDown at a point inside the widget.
	down := &gesture.PointerEvent{
		Base:           event.NewBase(event.TypeMouse, event.ModNone),
		EventType:      gesture.PointerDown,
		PointerID:      1,
		PointerType:    gesture.PointerTypeMouse,
		Position:       geometry.Pt(100, 100),
		GlobalPosition: geometry.Pt(100, 100),
		Button:         event.ButtonLeft,
		Buttons:        event.ButtonStateLeft,
	}
	w.HandlePointerEvent(down)

	// PointerUp at the same point.
	up := &gesture.PointerEvent{
		Base:           event.NewBase(event.TypeMouse, event.ModNone),
		EventType:      gesture.PointerUp,
		PointerID:      1,
		PointerType:    gesture.PointerTypeMouse,
		Position:       geometry.Pt(100, 100),
		GlobalPosition: geometry.Pt(100, 100),
		Button:         event.ButtonLeft,
		Buttons:        0,
	}
	w.HandlePointerEvent(up)

	if !clickFired {
		t.Error("ClickRecognizer.OnClick should have fired after down+up on GestureAware widget")
	}
}

func TestWindow_HandlePointerEvent_OutsideBoundsNotHit(t *testing.T) {
	a := New()
	w := a.Window()

	var clickFired bool
	click := gesture.NewClickRecognizer(gesture.ClickConfig{
		OnClick: func(_ gesture.ClickDetails) {
			clickFired = true
		},
	})

	root := newGestureAwareMock(click)
	root.layoutSize = geometry.Sz(100, 100)
	w.SetRoot(root)
	w.Frame()
	root.SetBounds(geometry.NewRect(0, 0, 100, 100))
	root.SetScreenOrigin(geometry.Pt(0, 0))

	// PointerDown OUTSIDE the widget bounds (widget is 100x100).
	down := &gesture.PointerEvent{
		EventType:      gesture.PointerDown,
		PointerID:      1,
		Position:       geometry.Pt(200, 200),
		GlobalPosition: geometry.Pt(200, 200),
		Button:         event.ButtonLeft,
	}
	w.HandlePointerEvent(down)

	up := &gesture.PointerEvent{
		EventType:      gesture.PointerUp,
		PointerID:      1,
		Position:       geometry.Pt(200, 200),
		GlobalPosition: geometry.Pt(200, 200),
	}
	w.HandlePointerEvent(up)

	if clickFired {
		t.Error("click should not fire when pointer is outside widget bounds")
	}
}

// containerGestureMock holds child widgets and implements GestureAware.
type containerGestureMock struct {
	widget.WidgetBase
	recognizers []gesture.Recognizer
}

func newContainerGestureMock(children []widget.Widget, recs ...gesture.Recognizer) *containerGestureMock {
	m := &containerGestureMock{recognizers: recs}
	m.SetVisible(true)
	m.SetEnabled(true)
	for _, c := range children {
		m.AddChild(c)
	}
	return m
}

func (m *containerGestureMock) GestureHitTest(_ geometry.Point) []gesture.Recognizer {
	return m.recognizers
}

func (m *containerGestureMock) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	return constraints.Constrain(geometry.Sz(300, 300))
}

func (m *containerGestureMock) Draw(_ widget.Context, _ widget.Canvas) {}

func (m *containerGestureMock) Event(_ widget.Context, _ event.Event) bool {
	return false
}

func TestWindow_HandlePointerEvent_NestedGestureAwareWidgets(t *testing.T) {
	a := New()
	w := a.Window()

	var parentClicked, childClicked bool

	parentClick := gesture.NewClickRecognizer(gesture.ClickConfig{
		OnClick: func(_ gesture.ClickDetails) {
			parentClicked = true
		},
	})
	childClick := gesture.NewClickRecognizer(gesture.ClickConfig{
		OnClick: func(_ gesture.ClickDetails) {
			childClicked = true
		},
	})

	child := newGestureAwareMock(childClick)
	child.layoutSize = geometry.Sz(50, 50)

	parent := newContainerGestureMock([]widget.Widget{child}, parentClick)

	w.SetRoot(parent)
	w.Frame()

	// Set bounds and screen origins for hit-testing.
	parent.SetBounds(geometry.NewRect(0, 0, 300, 300))
	parent.SetScreenOrigin(geometry.Pt(0, 0))
	child.SetBounds(geometry.NewRect(10, 10, 60, 60))
	child.SetScreenOrigin(geometry.Pt(10, 10))

	// PointerDown inside the child (which is inside the parent).
	down := &gesture.PointerEvent{
		Base:           event.NewBase(event.TypeMouse, event.ModNone),
		EventType:      gesture.PointerDown,
		PointerID:      1,
		PointerType:    gesture.PointerTypeMouse,
		Position:       geometry.Pt(30, 30),
		GlobalPosition: geometry.Pt(30, 30),
		Button:         event.ButtonLeft,
		Buttons:        event.ButtonStateLeft,
	}
	w.HandlePointerEvent(down)

	// Both parent and child recognizers should be in the arena.
	// The arena auto-resolves: with 2 members, Close() does not auto-pick.
	// On PointerUp, Sweep selects the first member.

	up := &gesture.PointerEvent{
		Base:           event.NewBase(event.TypeMouse, event.ModNone),
		EventType:      gesture.PointerUp,
		PointerID:      1,
		PointerType:    gesture.PointerTypeMouse,
		Position:       geometry.Pt(30, 30),
		GlobalPosition: geometry.Pt(30, 30),
		Button:         event.ButtonLeft,
		Buttons:        0,
	}
	w.HandlePointerEvent(up)

	// At least one of them should have been accepted (sweep picks first).
	if !parentClicked && !childClicked {
		t.Error("at least one click recognizer should fire for nested GestureAware widgets")
	}
}

func TestWindow_HandlePointerEvent_MoveRoutedToArena(t *testing.T) {
	a := New()
	w := a.Window()

	var dragStarted bool
	drag := gesture.NewDragRecognizer(gesture.DragConfig{
		OnDragStart: func(_ gesture.DragStartDetails) {
			dragStarted = true
		},
		OnDragUpdate: func(_ gesture.DragUpdateDetails) {},
		OnDragEnd:    func(_ gesture.DragEndDetails) {},
	})

	root := newGestureAwareMock(drag)
	root.layoutSize = geometry.Sz(400, 400)
	w.SetRoot(root)
	w.Frame()
	root.SetBounds(geometry.NewRect(0, 0, 400, 400))
	root.SetScreenOrigin(geometry.Pt(0, 0))

	// PointerDown.
	down := &gesture.PointerEvent{
		Base:           event.NewBase(event.TypeMouse, event.ModNone),
		EventType:      gesture.PointerDown,
		PointerID:      1,
		PointerType:    gesture.PointerTypeMouse,
		Position:       geometry.Pt(100, 100),
		GlobalPosition: geometry.Pt(100, 100),
		Button:         event.ButtonLeft,
		Buttons:        event.ButtonStateLeft,
	}
	w.HandlePointerEvent(down)

	// PointerMove beyond slop (PrecisePointerSlop = 1px for mouse).
	move := &gesture.PointerEvent{
		Base:           event.NewBase(event.TypeMouse, event.ModNone),
		EventType:      gesture.PointerMove,
		PointerID:      1,
		PointerType:    gesture.PointerTypeMouse,
		Position:       geometry.Pt(110, 110),
		GlobalPosition: geometry.Pt(110, 110),
		Buttons:        event.ButtonStateLeft,
	}
	w.HandlePointerEvent(move)

	if !dragStarted {
		t.Error("DragRecognizer.OnDragStart should fire after move beyond slop")
	}

	// PointerUp ends the drag.
	up := &gesture.PointerEvent{
		Base:           event.NewBase(event.TypeMouse, event.ModNone),
		EventType:      gesture.PointerUp,
		PointerID:      1,
		PointerType:    gesture.PointerTypeMouse,
		Position:       geometry.Pt(110, 110),
		GlobalPosition: geometry.Pt(110, 110),
		Buttons:        0,
	}
	w.HandlePointerEvent(up)
}

func TestWindow_HandlePointerEvent_LegacyWidgetsUnaffected(t *testing.T) {
	a := New()
	w := a.Window()

	// Root is a plain widget (NOT GestureAware).
	root := newMockWidget()
	w.SetRoot(root)
	w.Frame()

	// HandlePointerEvent should not break anything.
	down := &gesture.PointerEvent{
		EventType:      gesture.PointerDown,
		PointerID:      1,
		Position:       geometry.Pt(50, 50),
		GlobalPosition: geometry.Pt(50, 50),
	}
	w.HandlePointerEvent(down)

	// Legacy event dispatch still works.
	me := event.NewMouseEvent(event.MousePress, event.ButtonLeft,
		event.ButtonStateLeft, geometry.Pt(50, 50), geometry.Pt(50, 50), event.ModNone)
	w.HandleEvent(me)

	if !root.eventCalled {
		t.Error("legacy HandleEvent must still work for non-GestureAware widgets")
	}
}

func TestWindow_HandlePointerEvent_CancelRejectsAll(t *testing.T) {
	a := New()
	w := a.Window()

	var cancelCalled bool
	click := gesture.NewClickRecognizer(gesture.ClickConfig{
		OnClickCancel: func() {
			cancelCalled = true
		},
	})

	root := newGestureAwareMock(click)
	root.layoutSize = geometry.Sz(200, 200)
	w.SetRoot(root)
	w.Frame()
	root.SetBounds(geometry.NewRect(0, 0, 200, 200))
	root.SetScreenOrigin(geometry.Pt(0, 0))

	// PointerDown.
	down := &gesture.PointerEvent{
		Base:           event.NewBase(event.TypeMouse, event.ModNone),
		EventType:      gesture.PointerDown,
		PointerID:      1,
		PointerType:    gesture.PointerTypeMouse,
		Position:       geometry.Pt(100, 100),
		GlobalPosition: geometry.Pt(100, 100),
		Button:         event.ButtonLeft,
		Buttons:        event.ButtonStateLeft,
	}
	w.HandlePointerEvent(down)

	// PointerCancel.
	cancel := &gesture.PointerEvent{
		EventType: gesture.PointerCancel,
		PointerID: 1,
	}
	w.HandlePointerEvent(cancel)

	// After cancel, sweep should do nothing (no members or resolved).
	// The recognizer's handleCancel should have been called via Route.
	if !cancelCalled {
		t.Error("ClickRecognizer.OnClickCancel should fire after PointerCancel")
	}
}

func TestHitTestGestureAware_InvisibleWidgetSkipped(t *testing.T) {
	a := New()
	w := a.Window()

	click := gesture.NewClickRecognizer(gesture.ClickConfig{})
	root := newGestureAwareMock(click)
	root.layoutSize = geometry.Sz(200, 200)
	root.SetVisible(false) // Invisible!
	w.SetRoot(root)
	w.Frame()
	root.SetBounds(geometry.NewRect(0, 0, 200, 200))
	root.SetScreenOrigin(geometry.Pt(0, 0))

	recs := w.hitTestGestureAware(geometry.Pt(100, 100))
	if len(recs) != 0 {
		t.Errorf("hitTestGestureAware returned %d recognizers for invisible widget, want 0", len(recs))
	}
}

// TestWindow_GestureAndDerived_DerivedMouseEventFiresAfterGesture verifies that
// the derived MouseEvent from HandlePointerEvent Part 2 still reaches the widget
// AFTER the gesture arena processes in Part 1. This is the integration test for
// the race condition fix: gesture callbacks must not corrupt state that derived
// MouseEvent handlers depend on.
func TestWindow_GestureAndDerived_DerivedMouseEventFiresAfterGesture(t *testing.T) {
	a := New()
	w := a.Window()

	// Track the sequence of calls.
	var sequence []string

	click := gesture.NewClickRecognizer(gesture.ClickConfig{
		OnClickDown: func(_ gesture.ClickDownDetails) {
			sequence = append(sequence, "gesture:down")
		},
		OnClick: func(_ gesture.ClickDetails) {
			sequence = append(sequence, "gesture:click")
		},
	})

	root := newGestureAwareMock(click)
	root.layoutSize = geometry.Sz(200, 200)
	// Override Event to track derived events.
	root.eventCallback = func(e event.Event) {
		if me, ok := e.(*event.MouseEvent); ok {
			sequence = append(sequence, "derived:"+me.MouseType.String())
		}
	}
	w.SetRoot(root)
	w.Frame()
	root.SetBounds(geometry.NewRect(0, 0, 200, 200))
	root.SetScreenOrigin(geometry.Pt(0, 0))

	// PointerDown.
	down := &gesture.PointerEvent{
		Base:           event.NewBase(event.TypeMouse, event.ModNone),
		EventType:      gesture.PointerDown,
		PointerID:      1,
		PointerType:    gesture.PointerTypeMouse,
		Position:       geometry.Pt(100, 100),
		GlobalPosition: geometry.Pt(100, 100),
		Button:         event.ButtonLeft,
		Buttons:        event.ButtonStateLeft,
	}
	w.HandlePointerEvent(down)

	// PointerUp.
	up := &gesture.PointerEvent{
		Base:           event.NewBase(event.TypeMouse, event.ModNone),
		EventType:      gesture.PointerUp,
		PointerID:      1,
		PointerType:    gesture.PointerTypeMouse,
		Position:       geometry.Pt(100, 100),
		GlobalPosition: geometry.Pt(100, 100),
		Button:         event.ButtonLeft,
		Buttons:        0,
	}
	w.HandlePointerEvent(up)

	// Verify ordering: gesture callbacks fire in Part 1, derived events in Part 2.
	// Expected sequence:
	// 1. gesture:down   (Part 1: AddPointer -> OnClickDown)
	// 2. derived:Enter  (Part 2: trackMouseButtonOwnership -> updateHover on first press)
	// 3. derived:Press  (Part 2: derived MousePress)
	// 4. gesture:click  (Part 1: Route -> handleUp -> fireClick)
	// 5. derived:Release (Part 2: derived MouseRelease)
	expected := []string{"gesture:down", "derived:Enter", "derived:Press", "gesture:click", "derived:Release"}
	if len(sequence) != len(expected) {
		t.Fatalf("sequence length = %d, want %d\nsequence: %v", len(sequence), len(expected), sequence)
	}
	for i, exp := range expected {
		if sequence[i] != exp {
			t.Errorf("sequence[%d] = %q, want %q\nfull sequence: %v", i, sequence[i], exp, sequence)
		}
	}
}
