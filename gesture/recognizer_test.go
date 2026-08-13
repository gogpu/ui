package gesture

import (
	"testing"
	"time"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
)

func TestRecognizerBase_IsTrackingPointer(t *testing.T) {
	rb := &RecognizerBase{}
	if rb.IsTrackingPointer(1) {
		t.Error("should not be tracking any pointer initially")
	}

	arena := NewArena()
	mock := newMock("A")
	rb.StartTrackingPointer(1, arena, mock)

	if !rb.IsTrackingPointer(1) {
		t.Error("should be tracking pointer 1")
	}
	if rb.IsTrackingPointer(2) {
		t.Error("should not be tracking pointer 2")
	}

	rb.StopTrackingPointer(1)
	if rb.IsTrackingPointer(1) {
		t.Error("should not be tracking pointer 1 after stop")
	}
}

func TestRecognizerBase_Dispose(t *testing.T) {
	rb := &RecognizerBase{}
	arena := NewArena()
	mock := newMock("A")
	rb.StartTrackingPointer(1, arena, mock)

	rb.Dispose()
	if rb.arena != nil {
		t.Error("arena should be nil after dispose")
	}
	if rb.trackedPointers != nil {
		t.Error("trackedPointers should be nil after dispose")
	}
}

func TestRecognizerBase_ResolvePointerNilArena(t *testing.T) {
	rb := &RecognizerBase{}
	// Should not panic with nil arena.
	rb.ResolvePointer(1, Accepted, newMock("A"))
}

func TestRecognizerBase_SlopByDevice(t *testing.T) {
	rb := &RecognizerBase{deviceKind: DeviceKindPrecise}
	if rb.Slop() != PrecisePointerSlop {
		t.Errorf("Slop() = %f, want %f", rb.Slop(), PrecisePointerSlop)
	}

	rb.SetDeviceKind(PointerTypeTouch)
	if rb.Slop() != TouchSlop {
		t.Errorf("Slop() = %f, want %f", rb.Slop(), TouchSlop)
	}
}

func TestClickRecognizer_PressedSignal(t *testing.T) {
	sig := state.NewSignalWithOptions(false, state.Options[bool]{
		Equal: func(a, b bool) bool { return a == b },
	})

	rec := NewClickRecognizer(ClickConfig{
		OnClick: func(_ ClickDetails) {},
	}, WithPressedSignal(sig))

	arena := NewArena()
	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	rec.AddPointer(down, arena)
	arena.Close(1)

	if !sig.Get() {
		t.Error("pressed signal should be true after pointer down")
	}

	up := makePointerEvent(PointerUp, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 50*time.Millisecond)
	rec.HandleEvent(up)

	if sig.Get() {
		t.Error("pressed signal should be false after click completes")
	}
}

func TestClickRecognizer_PointerCancel(t *testing.T) {
	var cancelFired bool
	rec := NewClickRecognizer(ClickConfig{
		OnClickCancel: func() { cancelFired = true },
	})

	arena := NewArena()
	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	rec.AddPointer(down, arena)
	arena.Close(1)

	cancel := makePointerEvent(PointerCancel, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 50*time.Millisecond)
	rec.HandleEvent(cancel)

	if !cancelFired {
		t.Error("OnClickCancel should fire on PointerCancel")
	}
}

func TestClickRecognizer_Dispose(t *testing.T) {
	rec := NewClickRecognizer(ClickConfig{})
	rec.Dispose()
	// Should not panic.
}

func TestDragRecognizer_Dispose(t *testing.T) {
	rec := NewDragRecognizer(DragConfig{})
	rec.Dispose()
	// Should not panic.
}

func TestLongPressRecognizer_Dispose(t *testing.T) {
	rec := NewLongPressRecognizer(LongPressConfig{})
	rec.Dispose()
	// Should not panic.
}

func TestTapAndDragRecognizer_Dispose(t *testing.T) {
	rec := NewTapAndDragRecognizer(TapAndDragConfig{})
	rec.Dispose()
	// Should not panic.
}

func TestClickRecognizer_IgnoreNonDown(t *testing.T) {
	rec := NewClickRecognizer(ClickConfig{})
	arena := NewArena()
	// AddPointer with a move event should return false.
	move := makePointerEvent(PointerMove, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	if rec.AddPointer(move, arena) {
		t.Error("AddPointer should return false for non-down events")
	}
}

func TestDragRecognizer_IgnoreNonDown(t *testing.T) {
	rec := NewDragRecognizer(DragConfig{})
	arena := NewArena()
	up := makePointerEvent(PointerUp, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	if rec.AddPointer(up, arena) {
		t.Error("AddPointer should return false for non-down events")
	}
}

func TestLongPressRecognizer_IgnoreNonDown(t *testing.T) {
	rec := NewLongPressRecognizer(LongPressConfig{})
	arena := NewArena()
	cancel := makePointerEvent(PointerCancel, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	if rec.AddPointer(cancel, arena) {
		t.Error("AddPointer should return false for non-down events")
	}
}

func TestTapAndDragRecognizer_IgnoreNonDown(t *testing.T) {
	rec := NewTapAndDragRecognizer(TapAndDragConfig{})
	arena := NewArena()
	move := makePointerEvent(PointerMove, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	if rec.AddPointer(move, arena) {
		t.Error("AddPointer should return false for non-down events")
	}
}

func TestClickRecognizer_RejectGestureViaArena(t *testing.T) {
	var cancelFired bool
	rec := NewClickRecognizer(ClickConfig{
		OnClickCancel: func() { cancelFired = true },
	})

	// Two recognizers compete; one wins, one loses.
	winner := newMock("winner")
	arena := NewArena()

	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	rec.AddPointer(down, arena)
	arena.Add(1, winner)
	arena.Close(1)

	// Winner takes the arena.
	arena.Resolve(1, winner, Accepted)

	if !cancelFired {
		t.Error("OnClickCancel should fire when arena rejects the click recognizer")
	}
}

func TestDragRecognizer_RejectGestureViaArena(t *testing.T) {
	var cancelFired bool
	rec := NewDragRecognizer(DragConfig{
		OnDragCancel: func() { cancelFired = true },
	})

	winner := newMock("winner")
	arena := NewArena()

	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	rec.AddPointer(down, arena)
	arena.Add(1, winner)
	arena.Close(1)

	arena.Resolve(1, winner, Accepted)

	if !cancelFired {
		t.Error("OnDragCancel should fire when arena rejects the drag recognizer")
	}
}

func TestLongPressRecognizer_RejectGestureViaArena(t *testing.T) {
	var cancelFired bool
	rec := NewLongPressRecognizer(LongPressConfig{
		OnLongPressCancel: func() { cancelFired = true },
	})

	winner := newMock("winner")
	arena := NewArena()

	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	rec.AddPointer(down, arena)
	arena.Add(1, winner)
	arena.Close(1)

	arena.Resolve(1, winner, Accepted)

	if !cancelFired {
		t.Error("OnLongPressCancel should fire when arena rejects the long press recognizer")
	}
}

// TestNoPointerSentinel verifies that recognizers use -1 (noPointer) as
// the sentinel for "no active pointer", not 0. PointerID 0 is valid
// per W3C Pointer Events spec.
func TestNoPointerSentinel(t *testing.T) {
	t.Run("click_accepts_pointer_zero", func(t *testing.T) {
		var clicked bool
		rec := NewClickRecognizer(ClickConfig{
			OnClick: func(_ ClickDetails) { clicked = true },
		})

		arena := NewArena()
		down := makePointerEvent(PointerDown, 0, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
		if !rec.AddPointer(down, arena) {
			t.Fatal("AddPointer should accept PointerID=0")
		}
		arena.Close(0)

		up := makePointerEvent(PointerUp, 0, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 50*time.Millisecond)
		rec.HandleEvent(up)

		if !clicked {
			t.Error("click should fire for PointerID=0")
		}
	})

	t.Run("drag_accepts_pointer_zero", func(t *testing.T) {
		var started bool
		rec := NewDragRecognizer(DragConfig{
			Direction:   DragDirectionPan,
			OnDragStart: func(_ DragStartDetails) { started = true },
		})

		arena := NewArena()
		down := makePointerEvent(PointerDown, 0, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
		if !rec.AddPointer(down, arena) {
			t.Fatal("AddPointer should accept PointerID=0")
		}
		arena.Close(0)

		move := makePointerEvent(PointerMove, 0, PointerTypeMouse, geometry.Pt(55, 55), event.ButtonLeft, 50*time.Millisecond)
		rec.HandleEvent(move)

		if !started {
			t.Error("drag should start for PointerID=0")
		}
	})

	t.Run("longpress_accepts_pointer_zero", func(t *testing.T) {
		var fired bool
		rec := NewLongPressRecognizer(LongPressConfig{
			OnLongPress: func(_ LongPressDetails) { fired = true },
		})

		arena := NewArena()
		down := makePointerEvent(PointerDown, 0, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
		if !rec.AddPointer(down, arena) {
			t.Fatal("AddPointer should accept PointerID=0")
		}
		arena.Close(0)

		rec.CheckTimer(500 * time.Millisecond)

		if !fired {
			t.Error("long press should fire for PointerID=0")
		}
	})
}

// TestRecognizerBase_MemberOverride verifies that SetMemberOverride
// causes StartTrackingPointer to register the override in the arena.
func TestRecognizerBase_MemberOverride(t *testing.T) {
	rb := &RecognizerBase{}
	override := newMock("override")
	rb.SetMemberOverride(override)

	arena := NewArena()
	selfMock := newMock("self")
	rb.StartTrackingPointer(1, arena, selfMock)
	arena.Close(1)

	// The override should be accepted (it's the registered member),
	// not the selfMock that was passed as the member parameter.
	if !override.accepted {
		t.Error("override should be accepted as the arena member")
	}
	if selfMock.accepted {
		t.Error("self should not be accepted when override is set")
	}
}

// TestRecognizerBase_ResolvePointerUsesOverride verifies that
// ResolvePointer uses the member override for arena identity matching.
func TestRecognizerBase_ResolvePointerUsesOverride(t *testing.T) {
	rb := &RecognizerBase{}
	override := newMock("override")
	rb.SetMemberOverride(override)

	arena := NewArena()
	other := newMock("other")
	rb.StartTrackingPointer(1, arena, newMock("self"))
	arena.Add(1, other)
	arena.Close(1)

	// Resolve rejected. Since override is the registered member,
	// the arena should remove the override and auto-accept other.
	rb.ResolvePointer(1, Rejected, newMock("self"))

	if !other.accepted {
		t.Error("other should be auto-accepted after override resolves Rejected")
	}
}

func TestArena_RouteNonExistentPointer(t *testing.T) {
	a := NewArena()
	ev := &PointerEvent{EventType: PointerMove, PointerID: 99}
	// Should not panic.
	a.Route(ev)
}

func TestArena_CloseNonExistent(t *testing.T) {
	a := NewArena()
	// Should not panic.
	a.Close(99)
}

func TestArena_SweepNonExistent(t *testing.T) {
	a := NewArena()
	// Should not panic.
	a.Sweep(99)
}

func TestArena_HoldNonExistent(t *testing.T) {
	a := NewArena()
	// Should not panic.
	a.Hold(99)
}

func TestArena_ReleaseNonExistent(t *testing.T) {
	a := NewArena()
	// Should not panic.
	a.Release(99)
}
