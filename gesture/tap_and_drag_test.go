package gesture

import (
	"testing"
	"time"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
)

func TestTapAndDrag_SingleTap(t *testing.T) {
	var tapDownCount, tapUpCount int

	rec := NewTapAndDragRecognizer(TapAndDragConfig{
		OnTapDown: func(d TapDragDownDetails) {
			tapDownCount = d.ConsecutiveTapCount
		},
		OnTapUp: func(d TapDragUpDetails) {
			tapUpCount = d.ConsecutiveTapCount
		},
	})

	arena := NewArena()
	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	rec.AddPointer(down, arena)
	arena.Close(1)

	up := makePointerEvent(PointerUp, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 50*time.Millisecond)
	rec.HandleEvent(up)

	if tapDownCount != 1 {
		t.Errorf("tapDownCount = %d, want 1", tapDownCount)
	}
	if tapUpCount != 1 {
		t.Errorf("tapUpCount = %d, want 1", tapUpCount)
	}
}

func TestTapAndDrag_DoubleTapDrag(t *testing.T) {
	var dragStartCount int
	var dragUpdates []TapDragUpdateDetails

	rec := NewTapAndDragRecognizer(TapAndDragConfig{
		OnTapDown: func(_ TapDragDownDetails) {},
		OnTapUp:   func(_ TapDragUpDetails) {},
		OnDragStart: func(d TapDragStartDetails) {
			dragStartCount = d.ConsecutiveTapCount
		},
		OnDragUpdate: func(d TapDragUpdateDetails) {
			dragUpdates = append(dragUpdates, d)
		},
		OnDragEnd: func(_ TapDragEndDetails) {},
	})

	// First tap.
	arena1 := NewArena()
	down1 := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	rec.AddPointer(down1, arena1)
	arena1.Close(1)
	up1 := makePointerEvent(PointerUp, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 50*time.Millisecond)
	rec.HandleEvent(up1)

	// Second tap -> drag (double-tap-drag for word selection).
	t2 := 150 * time.Millisecond
	arena2 := NewArena()
	down2 := makePointerEvent(PointerDown, 2, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, t2)
	rec.AddPointer(down2, arena2)
	arena2.Close(2)

	// Move beyond slop to start drag.
	move := makePointerEvent(PointerMove, 2, PointerTypeMouse, geometry.Pt(60, 50), event.ButtonLeft, t2+50*time.Millisecond)
	rec.HandleEvent(move)

	if dragStartCount != 2 {
		t.Errorf("drag start consecutiveTapCount = %d, want 2", dragStartCount)
	}

	// Continue drag.
	move2 := makePointerEvent(PointerMove, 2, PointerTypeMouse, geometry.Pt(80, 50), event.ButtonLeft, t2+100*time.Millisecond)
	rec.HandleEvent(move2)

	if len(dragUpdates) == 0 {
		t.Fatal("should have received drag updates")
	}
	if dragUpdates[len(dragUpdates)-1].ConsecutiveTapCount != 2 {
		t.Errorf("drag update consecutiveTapCount = %d, want 2", dragUpdates[len(dragUpdates)-1].ConsecutiveTapCount)
	}

	// End drag.
	up2 := makePointerEvent(PointerUp, 2, PointerTypeMouse, geometry.Pt(80, 50), event.ButtonLeft, t2+150*time.Millisecond)
	rec.HandleEvent(up2)
}

func TestTapAndDrag_TripleTap(t *testing.T) {
	var lastTapUpCount int

	rec := NewTapAndDragRecognizer(TapAndDragConfig{
		OnTapDown: func(_ TapDragDownDetails) {},
		OnTapUp: func(d TapDragUpDetails) {
			lastTapUpCount = d.ConsecutiveTapCount
		},
	})

	for i := 0; i < 3; i++ {
		ts := time.Duration(i) * 150 * time.Millisecond
		arena := NewArena()
		down := makePointerEvent(PointerDown, i+1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, ts)
		rec.AddPointer(down, arena)
		arena.Close(i + 1)
		up := makePointerEvent(PointerUp, i+1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, ts+50*time.Millisecond)
		rec.HandleEvent(up)
	}

	if lastTapUpCount != 3 {
		t.Errorf("triple tap count = %d, want 3", lastTapUpCount)
	}
}

func TestTapAndDrag_TapCountReset(t *testing.T) {
	var tapCounts []int

	rec := NewTapAndDragRecognizer(TapAndDragConfig{
		OnTapDown: func(d TapDragDownDetails) {
			tapCounts = append(tapCounts, d.ConsecutiveTapCount)
		},
		OnTapUp: func(_ TapDragUpDetails) {},
	})

	// First tap.
	arena1 := NewArena()
	down1 := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	rec.AddPointer(down1, arena1)
	arena1.Close(1)
	up1 := makePointerEvent(PointerUp, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 50*time.Millisecond)
	rec.HandleEvent(up1)

	// Second tap after timeout.
	t2 := 500 * time.Millisecond
	arena2 := NewArena()
	down2 := makePointerEvent(PointerDown, 2, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, t2)
	rec.AddPointer(down2, arena2)
	arena2.Close(2)
	up2 := makePointerEvent(PointerUp, 2, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, t2+50*time.Millisecond)
	rec.HandleEvent(up2)

	if len(tapCounts) != 2 {
		t.Fatalf("got %d taps, want 2", len(tapCounts))
	}
	if tapCounts[1] != 1 {
		t.Errorf("second tap count = %d, want 1 (reset due to timeout)", tapCounts[1])
	}
}

func TestTapAndDrag_DragEndVelocity(t *testing.T) {
	var endDetails TapDragEndDetails

	rec := NewTapAndDragRecognizer(TapAndDragConfig{
		OnTapDown:   func(_ TapDragDownDetails) {},
		OnDragStart: func(_ TapDragStartDetails) {},
		OnDragEnd: func(d TapDragEndDetails) {
			endDetails = d
		},
	})

	arena := NewArena()
	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(0, 0), event.ButtonLeft, 0)
	rec.AddPointer(down, arena)
	arena.Close(1)

	// Drag to start.
	for i := 1; i <= 5; i++ {
		ts := time.Duration(i*20) * time.Millisecond
		move := makePointerEvent(PointerMove, 1, PointerTypeMouse, geometry.Pt(float32(i*20), 0), event.ButtonLeft, ts)
		rec.HandleEvent(move)
	}

	// End drag.
	up := makePointerEvent(PointerUp, 1, PointerTypeMouse, geometry.Pt(100, 0), event.ButtonLeft, 100*time.Millisecond)
	rec.HandleEvent(up)

	if endDetails.ConsecutiveTapCount != 1 {
		t.Errorf("drag end consecutiveTapCount = %d, want 1", endDetails.ConsecutiveTapCount)
	}
	if endDetails.Velocity.X <= 0 {
		t.Error("drag end velocity X should be positive")
	}
}

func TestTapAndDrag_Cancel(t *testing.T) {
	var canceled bool
	rec := NewTapAndDragRecognizer(TapAndDragConfig{
		OnTapDown: func(_ TapDragDownDetails) {},
		OnCancel:  func() { canceled = true },
	})

	arena := NewArena()
	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	rec.AddPointer(down, arena)
	arena.Close(1)

	cancel := makePointerEvent(PointerCancel, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 50*time.Millisecond)
	rec.HandleEvent(cancel)

	if !canceled {
		t.Error("OnCancel should fire on PointerCancel")
	}
}

func TestTapAndDrag_ButtonMismatchResets(t *testing.T) {
	var tapCounts []int

	rec := NewTapAndDragRecognizer(TapAndDragConfig{
		OnTapDown: func(d TapDragDownDetails) {
			tapCounts = append(tapCounts, d.ConsecutiveTapCount)
		},
		OnTapUp: func(_ TapDragUpDetails) {},
	})

	// First tap with left button.
	arena1 := NewArena()
	down1 := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	rec.AddPointer(down1, arena1)
	arena1.Close(1)
	up1 := makePointerEvent(PointerUp, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 50*time.Millisecond)
	rec.HandleEvent(up1)

	// Second tap with right button.
	t2 := 150 * time.Millisecond
	arena2 := NewArena()
	down2 := makePointerEvent(PointerDown, 2, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonRight, t2)
	rec.AddPointer(down2, arena2)
	arena2.Close(2)
	up2 := makePointerEvent(PointerUp, 2, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonRight, t2+50*time.Millisecond)
	rec.HandleEvent(up2)

	if len(tapCounts) < 2 {
		t.Fatalf("got %d taps, want 2", len(tapCounts))
	}
	if tapCounts[1] != 1 {
		t.Errorf("second tap count = %d, want 1 (reset due to button mismatch)", tapCounts[1])
	}
}

func TestTapAndDrag_IgnoreWrongPointer(t *testing.T) {
	var dragStarted bool
	rec := NewTapAndDragRecognizer(TapAndDragConfig{
		OnTapDown:   func(_ TapDragDownDetails) {},
		OnDragStart: func(_ TapDragStartDetails) { dragStarted = true },
	})

	arena := NewArena()
	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	rec.AddPointer(down, arena)
	arena.Close(1)

	// Move with wrong pointer ID.
	move := makePointerEvent(PointerMove, 99, PointerTypeMouse, geometry.Pt(100, 100), event.ButtonLeft, 50*time.Millisecond)
	rec.HandleEvent(move)

	if dragStarted {
		t.Error("should ignore events for wrong pointer ID")
	}
}

func TestTapAndDrag_MaxClickCountWraps(t *testing.T) {
	var tapCounts []int

	rec := NewTapAndDragRecognizer(TapAndDragConfig{
		OnTapDown: func(d TapDragDownDetails) {
			tapCounts = append(tapCounts, d.ConsecutiveTapCount)
		},
		OnTapUp: func(_ TapDragUpDetails) {},
	})

	// 4 consecutive taps: should wrap to 1 after MaxClickCount (3).
	for i := 0; i < 4; i++ {
		ts := time.Duration(i) * 150 * time.Millisecond
		arena := NewArena()
		down := makePointerEvent(PointerDown, i+1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, ts)
		rec.AddPointer(down, arena)
		arena.Close(i + 1)
		up := makePointerEvent(PointerUp, i+1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, ts+50*time.Millisecond)
		rec.HandleEvent(up)
	}

	if len(tapCounts) != 4 {
		t.Fatalf("got %d taps, want 4", len(tapCounts))
	}
	expected := []int{1, 2, 3, 1}
	for i, want := range expected {
		if tapCounts[i] != want {
			t.Errorf("tap[%d] = %d, want %d", i, tapCounts[i], want)
		}
	}
}
