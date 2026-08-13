package gesture

import (
	"math"
	"testing"
	"time"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
)

func TestDragRecognizer_PanDrag(t *testing.T) {
	var started, ended bool
	var updates []DragUpdateDetails

	arena := NewArena()
	rec := NewDragRecognizer(DragConfig{
		Direction: DragDirectionPan,
		OnDragStart: func(_ DragStartDetails) {
			started = true
		},
		OnDragUpdate: func(d DragUpdateDetails) {
			updates = append(updates, d)
		},
		OnDragEnd: func(_ DragEndDetails) {
			ended = true
		},
	})

	// Pointer down.
	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	rec.AddPointer(down, arena)
	arena.Close(1)

	// Move beyond slop (1px for mouse).
	move1 := makePointerEvent(PointerMove, 1, PointerTypeMouse, geometry.Pt(55, 55), event.ButtonLeft, 50*time.Millisecond)
	rec.HandleEvent(move1)

	if !started {
		t.Error("drag should have started after exceeding slop")
	}

	// Another move.
	move2 := makePointerEvent(PointerMove, 1, PointerTypeMouse, geometry.Pt(60, 60), event.ButtonLeft, 100*time.Millisecond)
	rec.HandleEvent(move2)

	if len(updates) < 1 {
		t.Fatal("should have received drag updates")
	}

	// Pointer up.
	up := makePointerEvent(PointerUp, 1, PointerTypeMouse, geometry.Pt(60, 60), event.ButtonLeft, 150*time.Millisecond)
	rec.HandleEvent(up)

	if !ended {
		t.Error("drag should have ended on pointer up")
	}
}

func TestDragRecognizer_HorizontalOnly(t *testing.T) {
	var started bool
	rec := NewDragRecognizer(DragConfig{
		Direction:   DragDirectionHorizontal,
		OnDragStart: func(_ DragStartDetails) { started = true },
	})

	arena := NewArena()
	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	rec.AddPointer(down, arena)
	arena.Close(1)

	// Move vertically only: should not trigger horizontal drag.
	moveV := makePointerEvent(PointerMove, 1, PointerTypeMouse, geometry.Pt(50, 60), event.ButtonLeft, 50*time.Millisecond)
	rec.HandleEvent(moveV)

	if started {
		t.Error("vertical movement should not trigger horizontal drag")
	}

	// Move horizontally: should trigger.
	moveH := makePointerEvent(PointerMove, 1, PointerTypeMouse, geometry.Pt(55, 60), event.ButtonLeft, 100*time.Millisecond)
	rec.HandleEvent(moveH)

	if !started {
		t.Error("horizontal movement should trigger horizontal drag")
	}
}

func TestDragRecognizer_VerticalOnly(t *testing.T) {
	var started bool
	rec := NewDragRecognizer(DragConfig{
		Direction:   DragDirectionVertical,
		OnDragStart: func(_ DragStartDetails) { started = true },
	})

	arena := NewArena()
	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	rec.AddPointer(down, arena)
	arena.Close(1)

	// Move horizontally only: should not trigger.
	moveH := makePointerEvent(PointerMove, 1, PointerTypeMouse, geometry.Pt(55, 50), event.ButtonLeft, 50*time.Millisecond)
	rec.HandleEvent(moveH)

	if started {
		t.Error("horizontal movement should not trigger vertical drag")
	}

	// Move vertically: should trigger.
	moveV := makePointerEvent(PointerMove, 1, PointerTypeMouse, geometry.Pt(55, 55), event.ButtonLeft, 100*time.Millisecond)
	rec.HandleEvent(moveV)

	if !started {
		t.Error("vertical movement should trigger vertical drag")
	}
}

func TestDragRecognizer_TouchSlop(t *testing.T) {
	var started bool
	rec := NewDragRecognizer(DragConfig{
		Direction:   DragDirectionPan,
		OnDragStart: func(_ DragStartDetails) { started = true },
	})

	arena := NewArena()
	down := makePointerEvent(PointerDown, 1, PointerTypeTouch, geometry.Pt(50, 50), event.ButtonLeft, 0)
	rec.AddPointer(down, arena)
	arena.Close(1)

	// Move 10px (less than touch slop of 18px).
	move1 := makePointerEvent(PointerMove, 1, PointerTypeTouch, geometry.Pt(60, 50), event.ButtonLeft, 50*time.Millisecond)
	rec.HandleEvent(move1)

	if started {
		t.Error("10px movement should not exceed touch slop of 18px")
	}

	// Move beyond 18px total.
	move2 := makePointerEvent(PointerMove, 1, PointerTypeTouch, geometry.Pt(70, 50), event.ButtonLeft, 100*time.Millisecond)
	rec.HandleEvent(move2)

	if !started {
		t.Error("20px movement should exceed touch slop of 18px")
	}
}

func TestDragRecognizer_RejectOnUp(t *testing.T) {
	// Pointer up without exceeding slop should reject.
	var canceled bool
	rec := NewDragRecognizer(DragConfig{
		Direction:    DragDirectionPan,
		OnDragCancel: func() { canceled = true },
	})

	arena := NewArena()
	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	rec.AddPointer(down, arena)
	arena.Close(1)

	// Up without moving.
	up := makePointerEvent(PointerUp, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 50*time.Millisecond)
	rec.HandleEvent(up)

	if canceled {
		t.Error("OnDragCancel should not fire on normal rejection (no drag started)")
	}
}

func TestDragRecognizer_Velocity(t *testing.T) {
	var endDetails DragEndDetails
	rec := NewDragRecognizer(DragConfig{
		Direction:   DragDirectionPan,
		OnDragStart: func(_ DragStartDetails) {},
		OnDragEnd: func(d DragEndDetails) {
			endDetails = d
		},
	})

	arena := NewArena()
	// Simulate 100px horizontal in 100ms = 1000px/s.
	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(0, 0), event.ButtonLeft, 0)
	rec.AddPointer(down, arena)
	arena.Close(1)

	for i := 1; i <= 5; i++ {
		ts := time.Duration(i*20) * time.Millisecond
		pos := geometry.Pt(float32(i*20), 0)
		move := makePointerEvent(PointerMove, 1, PointerTypeMouse, pos, event.ButtonLeft, ts)
		rec.HandleEvent(move)
	}

	up := makePointerEvent(PointerUp, 1, PointerTypeMouse, geometry.Pt(100, 0), event.ButtonLeft, 100*time.Millisecond)
	rec.HandleEvent(up)

	if math.Abs(float64(endDetails.Velocity.X)-1000) > 100 {
		t.Errorf("end velocity X = %.1f, want ~1000", endDetails.Velocity.X)
	}
}

func TestDragRecognizer_PrimaryDelta(t *testing.T) {
	tests := []struct {
		name      string
		direction DragDirection
		delta     geometry.Point
		wantSign  float32 // +1 or -1 for direction
	}{
		{"horizontal_right", DragDirectionHorizontal, geometry.Pt(10, 5), 1},
		{"horizontal_left", DragDirectionHorizontal, geometry.Pt(-10, 5), -1},
		{"vertical_down", DragDirectionVertical, geometry.Pt(5, 10), 1},
		{"vertical_up", DragDirectionVertical, geometry.Pt(5, -10), -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPrimary float32
			rec := NewDragRecognizer(DragConfig{
				Direction:   tt.direction,
				OnDragStart: func(_ DragStartDetails) {},
				OnDragUpdate: func(d DragUpdateDetails) {
					gotPrimary = d.PrimaryDelta
				},
			})

			arena := NewArena()
			down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
			rec.AddPointer(down, arena)
			arena.Close(1)

			// First move triggers drag start.
			moveStart := makePointerEvent(PointerMove, 1, PointerTypeMouse,
				geometry.Pt(50+tt.delta.X, 50+tt.delta.Y), event.ButtonLeft, 50*time.Millisecond)
			rec.HandleEvent(moveStart)

			// Second move to get a clean delta.
			moveDelta := makePointerEvent(PointerMove, 1, PointerTypeMouse,
				geometry.Pt(50+tt.delta.X*2, 50+tt.delta.Y*2), event.ButtonLeft, 100*time.Millisecond)
			rec.HandleEvent(moveDelta)

			if (tt.wantSign > 0 && gotPrimary <= 0) || (tt.wantSign < 0 && gotPrimary >= 0) {
				t.Errorf("primaryDelta = %.1f, want sign %.0f", gotPrimary, tt.wantSign)
			}
		})
	}
}

func TestDragRecognizer_DraggingSignal(t *testing.T) {
	sig := state.NewSignalWithOptions(false, state.Options[bool]{
		Equal: func(a, b bool) bool { return a == b },
	})

	rec := NewDragRecognizer(DragConfig{
		Direction:   DragDirectionPan,
		OnDragStart: func(_ DragStartDetails) {},
		OnDragEnd:   func(_ DragEndDetails) {},
	}, WithDraggingSignal(sig))

	arena := NewArena()
	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(0, 0), event.ButtonLeft, 0)
	rec.AddPointer(down, arena)
	arena.Close(1)

	if sig.Get() {
		t.Error("signal should be false before drag starts")
	}

	// Move to trigger drag.
	move := makePointerEvent(PointerMove, 1, PointerTypeMouse, geometry.Pt(10, 10), event.ButtonLeft, 50*time.Millisecond)
	rec.HandleEvent(move)

	if !sig.Get() {
		t.Error("signal should be true during drag")
	}

	// End drag.
	up := makePointerEvent(PointerUp, 1, PointerTypeMouse, geometry.Pt(10, 10), event.ButtonLeft, 100*time.Millisecond)
	rec.HandleEvent(up)

	if sig.Get() {
		t.Error("signal should be false after drag ends")
	}
}

func TestDragRecognizer_PositionSignal(t *testing.T) {
	sig := state.NewSignal(geometry.Point{})

	rec := NewDragRecognizer(DragConfig{
		Direction:    DragDirectionPan,
		OnDragStart:  func(_ DragStartDetails) {},
		OnDragUpdate: func(_ DragUpdateDetails) {},
		OnDragEnd:    func(_ DragEndDetails) {},
	}, WithDragPositionSignal(sig))

	arena := NewArena()
	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(0, 0), event.ButtonLeft, 0)
	rec.AddPointer(down, arena)
	arena.Close(1)

	// Move to start drag and update position.
	move := makePointerEvent(PointerMove, 1, PointerTypeMouse, geometry.Pt(25, 30), event.ButtonLeft, 50*time.Millisecond)
	rec.HandleEvent(move)

	pos := sig.Get()
	if pos.X != 25 || pos.Y != 30 {
		t.Errorf("position signal = %v, want (25, 30)", pos)
	}
}

func TestDragRecognizer_Cancel(t *testing.T) {
	var canceled bool
	rec := NewDragRecognizer(DragConfig{
		Direction:    DragDirectionPan,
		OnDragStart:  func(_ DragStartDetails) {},
		OnDragCancel: func() { canceled = true },
	})

	arena := NewArena()
	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(0, 0), event.ButtonLeft, 0)
	rec.AddPointer(down, arena)
	arena.Close(1)

	// Move to start drag.
	move := makePointerEvent(PointerMove, 1, PointerTypeMouse, geometry.Pt(10, 10), event.ButtonLeft, 50*time.Millisecond)
	rec.HandleEvent(move)

	// Cancel.
	cancel := makePointerEvent(PointerCancel, 1, PointerTypeMouse, geometry.Pt(10, 10), event.ButtonLeft, 100*time.Millisecond)
	rec.HandleEvent(cancel)

	if !canceled {
		t.Error("OnDragCancel should be called on PointerCancel during active drag")
	}
}

func TestDragRecognizer_IgnoreWrongPointer(t *testing.T) {
	var started bool
	rec := NewDragRecognizer(DragConfig{
		Direction:   DragDirectionPan,
		OnDragStart: func(_ DragStartDetails) { started = true },
	})

	arena := NewArena()
	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(0, 0), event.ButtonLeft, 0)
	rec.AddPointer(down, arena)
	arena.Close(1)

	// Move with wrong pointer ID.
	move := makePointerEvent(PointerMove, 99, PointerTypeMouse, geometry.Pt(100, 100), event.ButtonLeft, 50*time.Millisecond)
	rec.HandleEvent(move)

	if started {
		t.Error("should ignore events for wrong pointer ID")
	}
}

// TestDragRecognizer_UpWithoutDragResolvesRejected verifies that when a
// drag recognizer receives pointer-up without exceeding slop in a multi-
// recognizer arena, it resolves Rejected so the arena can auto-accept
// the remaining member.
// Regression test for Issue 3: ghost member blocking auto-resolution.
func TestDragRecognizer_UpWithoutDragResolvesRejected(t *testing.T) {
	arena := NewArena()

	drag := NewDragRecognizer(DragConfig{
		Direction: DragDirectionPan,
	})
	other := newMock("other")

	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	drag.AddPointer(down, arena)
	arena.Add(1, other)
	arena.Close(1)

	// Pointer up without any movement. Drag should reject (slop not exceeded).
	up := makePointerEvent(PointerUp, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 50*time.Millisecond)
	arena.Route(up)

	// The other member should be auto-accepted after drag rejects.
	if !other.accepted {
		t.Error("other member should be auto-accepted after drag resolves Rejected on pointer up without drag")
	}
}

func TestDragDirection_String(t *testing.T) {
	tests := []struct {
		dir  DragDirection
		want string
	}{
		{DragDirectionPan, "Pan"},
		{DragDirectionHorizontal, "Horizontal"},
		{DragDirectionVertical, "Vertical"},
		{DragDirection(99), "Unknown"},
	}

	for _, tt := range tests {
		got := tt.dir.String()
		if got != tt.want {
			t.Errorf("DragDirection(%d).String() = %q, want %q", tt.dir, got, tt.want)
		}
	}
}
