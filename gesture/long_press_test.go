package gesture

import (
	"testing"
	"time"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
)

func TestLongPressRecognizer_BasicLongPress(t *testing.T) {
	var pressDownFired, longPressFired, upFired bool

	rec := NewLongPressRecognizer(LongPressConfig{
		OnLongPressDown: func(_ LongPressDetails) {
			pressDownFired = true
		},
		OnLongPress: func(_ LongPressDetails) {
			longPressFired = true
		},
		OnLongPressUp: func(_ LongPressDetails) {
			upFired = true
		},
	})

	arena := NewArena()
	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	rec.AddPointer(down, arena)
	arena.Close(1)

	// Simulate frame ticks.
	// At 50ms: PressTimeout not reached (100ms).
	needsMore := rec.CheckTimer(50 * time.Millisecond)
	if !needsMore {
		t.Error("should need more frames before timeout")
	}
	if pressDownFired {
		t.Error("PressTimeout (100ms) not reached at 50ms")
	}

	// At 100ms: PressTimeout reached.
	needsMore = rec.CheckTimer(100 * time.Millisecond)
	if !needsMore {
		t.Error("should still need frames before LongPressTimeout")
	}
	if !pressDownFired {
		t.Error("PressTimeout should fire at 100ms")
	}

	// At 499ms: LongPressTimeout not reached.
	needsMore = rec.CheckTimer(499 * time.Millisecond)
	if !needsMore {
		t.Error("should still need frames before 500ms")
	}
	if longPressFired {
		t.Error("LongPressTimeout (500ms) not reached at 499ms")
	}

	// At 500ms: LongPressTimeout reached.
	needsMore = rec.CheckTimer(500 * time.Millisecond)
	if needsMore {
		t.Error("should not need more frames after LongPressTimeout")
	}
	if !longPressFired {
		t.Error("LongPressTimeout should fire at 500ms")
	}

	// Release.
	up := makePointerEvent(PointerUp, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 600*time.Millisecond)
	rec.HandleEvent(up)

	if !upFired {
		t.Error("OnLongPressUp should fire on pointer up after long press")
	}
}

func TestLongPressRecognizer_CancelOnMove(t *testing.T) {
	var canceled bool
	rec := NewLongPressRecognizer(LongPressConfig{
		OnLongPressCancel: func() { canceled = true },
	})

	arena := NewArena()
	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	rec.AddPointer(down, arena)
	arena.Close(1)

	// Move beyond slop before timeout.
	move := makePointerEvent(PointerMove, 1, PointerTypeMouse, geometry.Pt(55, 55), event.ButtonLeft, 50*time.Millisecond)
	rec.HandleEvent(move)

	if !canceled {
		t.Error("long press should be canceled when pointer moves beyond slop")
	}
}

func TestLongPressRecognizer_CancelOnEarlyUp(t *testing.T) {
	var longPressFired bool
	rec := NewLongPressRecognizer(LongPressConfig{
		OnLongPress: func(_ LongPressDetails) { longPressFired = true },
	})

	arena := NewArena()
	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	rec.AddPointer(down, arena)
	arena.Close(1)

	// Release before timeout.
	up := makePointerEvent(PointerUp, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 200*time.Millisecond)
	rec.HandleEvent(up)

	if longPressFired {
		t.Error("long press should not fire on early release")
	}
}

func TestLongPressRecognizer_LongPressDrag(t *testing.T) {
	var moveUpdates []LongPressMoveDetails

	rec := NewLongPressRecognizer(LongPressConfig{
		OnLongPress: func(_ LongPressDetails) {},
		OnLongPressMoveUpdate: func(d LongPressMoveDetails) {
			moveUpdates = append(moveUpdates, d)
		},
	})

	arena := NewArena()
	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	rec.AddPointer(down, arena)
	arena.Close(1)

	// Trigger long press.
	rec.CheckTimer(500 * time.Millisecond)

	// Move while long press is active (long-press-drag).
	move1 := makePointerEvent(PointerMove, 1, PointerTypeMouse, geometry.Pt(60, 60), event.ButtonLeft, 600*time.Millisecond)
	rec.HandleEvent(move1)

	move2 := makePointerEvent(PointerMove, 1, PointerTypeMouse, geometry.Pt(70, 70), event.ButtonLeft, 700*time.Millisecond)
	rec.HandleEvent(move2)

	if len(moveUpdates) != 2 {
		t.Errorf("got %d move updates, want 2", len(moveUpdates))
	}
}

func TestLongPressRecognizer_TouchSlop(t *testing.T) {
	var canceled bool
	rec := NewLongPressRecognizer(LongPressConfig{
		OnLongPressCancel: func() { canceled = true },
	})

	arena := NewArena()
	down := makePointerEvent(PointerDown, 1, PointerTypeTouch, geometry.Pt(50, 50), event.ButtonLeft, 0)
	rec.AddPointer(down, arena)
	arena.Close(1)

	// Move 10px (less than touch slop of 18px): should not cancel.
	move1 := makePointerEvent(PointerMove, 1, PointerTypeTouch, geometry.Pt(60, 50), event.ButtonLeft, 50*time.Millisecond)
	rec.HandleEvent(move1)

	if canceled {
		t.Error("10px movement should not cancel touch long press (slop is 18px)")
	}

	// Move beyond 18px from down position.
	move2 := makePointerEvent(PointerMove, 1, PointerTypeTouch, geometry.Pt(70, 50), event.ButtonLeft, 100*time.Millisecond)
	rec.HandleEvent(move2)

	if !canceled {
		t.Error("20px movement should cancel touch long press")
	}
}

func TestLongPressRecognizer_ActiveSignal(t *testing.T) {
	sig := state.NewSignalWithOptions(false, state.Options[bool]{
		Equal: func(a, b bool) bool { return a == b },
	})

	rec := NewLongPressRecognizer(LongPressConfig{
		OnLongPress: func(_ LongPressDetails) {},
	}, WithLongPressActiveSignal(sig))

	arena := NewArena()
	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	rec.AddPointer(down, arena)
	arena.Close(1)

	if sig.Get() {
		t.Error("signal should be false before long press triggers")
	}

	rec.CheckTimer(500 * time.Millisecond)

	if !sig.Get() {
		t.Error("signal should be true after long press triggers")
	}

	up := makePointerEvent(PointerUp, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 600*time.Millisecond)
	rec.HandleEvent(up)

	if sig.Get() {
		t.Error("signal should be false after pointer up")
	}
}

func TestLongPressRecognizer_PointerCancel(t *testing.T) {
	var canceled bool
	rec := NewLongPressRecognizer(LongPressConfig{
		OnLongPress:       func(_ LongPressDetails) {},
		OnLongPressCancel: func() { canceled = true },
	})

	arena := NewArena()
	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	rec.AddPointer(down, arena)
	arena.Close(1)

	// Trigger long press.
	rec.CheckTimer(500 * time.Millisecond)

	// Cancel.
	cancel := makePointerEvent(PointerCancel, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 600*time.Millisecond)
	rec.HandleEvent(cancel)

	if !canceled {
		t.Error("OnLongPressCancel should fire on PointerCancel during active long press")
	}
}

func TestLongPressRecognizer_CheckTimerNotPossible(t *testing.T) {
	rec := NewLongPressRecognizer(LongPressConfig{})

	// CheckTimer in ready state should return false.
	needsMore := rec.CheckTimer(500 * time.Millisecond)
	if needsMore {
		t.Error("CheckTimer should return false when not in possible state")
	}
}

// TestLongPressRecognizer_SlopResolvesRejected verifies that when a
// long-press recognizer exceeds slop in a multi-recognizer arena, it
// resolves Rejected so the arena can auto-accept remaining members.
// Regression test for Issue 4: ghost member blocking auto-resolution.
func TestLongPressRecognizer_SlopResolvesRejected(t *testing.T) {
	arena := NewArena()

	lp := NewLongPressRecognizer(LongPressConfig{
		OnLongPressCancel: func() {},
	})
	other := newMock("other")

	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	lp.AddPointer(down, arena)
	arena.Add(1, other)
	arena.Close(1)

	// Move beyond slop. LongPress should resolve Rejected.
	move := makePointerEvent(PointerMove, 1, PointerTypeMouse, geometry.Pt(55, 55), event.ButtonLeft, 50*time.Millisecond)
	arena.Route(move)

	// The other member should be auto-accepted after long press rejects.
	if !other.accepted {
		t.Error("other member should be auto-accepted after long press resolves Rejected on slop")
	}
}

func TestLongPressRecognizer_IgnoreWrongPointer(t *testing.T) {
	var canceled bool
	rec := NewLongPressRecognizer(LongPressConfig{
		OnLongPressCancel: func() { canceled = true },
	})

	arena := NewArena()
	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	rec.AddPointer(down, arena)
	arena.Close(1)

	// Move with wrong pointer ID.
	move := makePointerEvent(PointerMove, 99, PointerTypeMouse, geometry.Pt(100, 100), event.ButtonLeft, 50*time.Millisecond)
	rec.HandleEvent(move)

	if canceled {
		t.Error("should ignore events for wrong pointer ID")
	}
}
