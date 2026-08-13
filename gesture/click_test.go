package gesture

import (
	"testing"
	"time"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
)

// makePointerEvent creates a test pointer event.
func makePointerEvent(t PointerEventType, id int, pt PointerType, pos geometry.Point,
	btn event.Button, ts time.Duration) *PointerEvent {
	return &PointerEvent{
		Base:           event.NewBase(event.TypeMouse, event.ModNone),
		EventType:      t,
		PointerID:      id,
		PointerType:    pt,
		Position:       pos,
		GlobalPosition: pos,
		Button:         btn,
		Timestamp:      ts,
	}
}

func TestClickRecognizer_SingleClick(t *testing.T) {
	var gotDetails ClickDetails

	arena := NewArena()
	rec := NewClickRecognizer(ClickConfig{
		OnClick: func(d ClickDetails) {
			gotDetails = d
		},
	})

	// Pointer down.
	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	if !rec.AddPointer(down, arena) {
		t.Fatal("AddPointer should return true")
	}
	arena.Close(1)

	// Pointer up.
	up := makePointerEvent(PointerUp, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 50*time.Millisecond)
	rec.HandleEvent(up)

	if gotDetails.ClickCount != 1 {
		t.Errorf("ClickCount = %d, want 1", gotDetails.ClickCount)
	}
	if gotDetails.Button != event.ButtonLeft {
		t.Errorf("Button = %v, want Left", gotDetails.Button)
	}
}

func TestClickRecognizer_DoubleClick(t *testing.T) {
	var clicks []int

	arena := NewArena()
	rec := NewClickRecognizer(ClickConfig{
		OnClick: func(d ClickDetails) {
			clicks = append(clicks, d.ClickCount)
		},
	})

	// First click.
	t1 := time.Duration(0)
	down1 := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, t1)
	rec.AddPointer(down1, arena)
	arena.Close(1)
	up1 := makePointerEvent(PointerUp, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, t1+50*time.Millisecond)
	rec.HandleEvent(up1)

	// Second click within timeout.
	t2 := t1 + 150*time.Millisecond
	arena2 := NewArena()
	down2 := makePointerEvent(PointerDown, 2, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, t2)
	rec.AddPointer(down2, arena2)
	arena2.Close(2)
	up2 := makePointerEvent(PointerUp, 2, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, t2+50*time.Millisecond)
	rec.HandleEvent(up2)

	if len(clicks) != 2 {
		t.Fatalf("got %d clicks, want 2", len(clicks))
	}
	if clicks[0] != 1 {
		t.Errorf("first click count = %d, want 1", clicks[0])
	}
	if clicks[1] != 2 {
		t.Errorf("second click count = %d, want 2", clicks[1])
	}
}

func TestClickRecognizer_TripleClick(t *testing.T) {
	var clicks []int

	rec := NewClickRecognizer(ClickConfig{
		OnClick: func(d ClickDetails) {
			clicks = append(clicks, d.ClickCount)
		},
	})

	baseT := time.Duration(0)
	for i := 0; i < 3; i++ {
		ts := baseT + time.Duration(i)*150*time.Millisecond
		arena := NewArena()
		down := makePointerEvent(PointerDown, i+1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, ts)
		rec.AddPointer(down, arena)
		arena.Close(i + 1)
		up := makePointerEvent(PointerUp, i+1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, ts+50*time.Millisecond)
		rec.HandleEvent(up)
	}

	if len(clicks) != 3 {
		t.Fatalf("got %d clicks, want 3", len(clicks))
	}
	expected := []int{1, 2, 3}
	for i, want := range expected {
		if clicks[i] != want {
			t.Errorf("click[%d] = %d, want %d", i, clicks[i], want)
		}
	}
}

func TestClickRecognizer_MaxClickCount(t *testing.T) {
	tests := []struct {
		name     string
		maxCount int
		nClicks  int
		wantLast int
	}{
		{"max_3_with_4_clicks", 3, 4, 3},
		{"max_1_with_2_clicks", 1, 2, 1},
		{"max_2_with_3_clicks", 2, 3, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var lastCount int
			rec := NewClickRecognizer(ClickConfig{
				MaxClickCount: tt.maxCount,
				OnClick: func(d ClickDetails) {
					lastCount = d.ClickCount
				},
			})

			for i := 0; i < tt.nClicks; i++ {
				ts := time.Duration(i) * 150 * time.Millisecond
				arena := NewArena()
				down := makePointerEvent(PointerDown, i+1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, ts)
				rec.AddPointer(down, arena)
				arena.Close(i + 1)
				up := makePointerEvent(PointerUp, i+1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, ts+50*time.Millisecond)
				rec.HandleEvent(up)
			}

			if lastCount != tt.wantLast {
				t.Errorf("last click count = %d, want %d", lastCount, tt.wantLast)
			}
		})
	}
}

func TestClickRecognizer_TimingReset(t *testing.T) {
	var clicks []int
	rec := NewClickRecognizer(ClickConfig{
		OnClick: func(d ClickDetails) {
			clicks = append(clicks, d.ClickCount)
		},
	})

	// First click.
	arena1 := NewArena()
	down1 := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	rec.AddPointer(down1, arena1)
	arena1.Close(1)
	up1 := makePointerEvent(PointerUp, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 50*time.Millisecond)
	rec.HandleEvent(up1)

	// Second click after DoubleTapTimeout (too slow).
	t2 := 500 * time.Millisecond
	arena2 := NewArena()
	down2 := makePointerEvent(PointerDown, 2, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, t2)
	rec.AddPointer(down2, arena2)
	arena2.Close(2)
	up2 := makePointerEvent(PointerUp, 2, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, t2+50*time.Millisecond)
	rec.HandleEvent(up2)

	if len(clicks) != 2 {
		t.Fatalf("got %d clicks, want 2", len(clicks))
	}
	if clicks[1] != 1 {
		t.Errorf("second click count = %d, want 1 (reset due to timeout)", clicks[1])
	}
}

func TestClickRecognizer_DoubleTapMinTime(t *testing.T) {
	var clicks []int
	rec := NewClickRecognizer(ClickConfig{
		OnClick: func(d ClickDetails) {
			clicks = append(clicks, d.ClickCount)
		},
	})

	// First click.
	arena1 := NewArena()
	down1 := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	rec.AddPointer(down1, arena1)
	arena1.Close(1)
	up1 := makePointerEvent(PointerUp, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 50*time.Millisecond)
	rec.HandleEvent(up1)

	// Second click too fast (anti-bounce).
	t2 := 50*time.Millisecond + 10*time.Millisecond // 60ms from start, 10ms from up
	arena2 := NewArena()
	down2 := makePointerEvent(PointerDown, 2, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, t2)
	rec.AddPointer(down2, arena2)
	arena2.Close(2)
	up2 := makePointerEvent(PointerUp, 2, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, t2+50*time.Millisecond)
	rec.HandleEvent(up2)

	if len(clicks) < 2 {
		t.Fatalf("got %d clicks, want 2", len(clicks))
	}
	if clicks[1] != 1 {
		t.Errorf("second click count = %d, want 1 (reset due to anti-bounce)", clicks[1])
	}
}

func TestClickRecognizer_ButtonMismatch(t *testing.T) {
	var clicks []int
	rec := NewClickRecognizer(ClickConfig{
		OnClick: func(d ClickDetails) {
			clicks = append(clicks, d.ClickCount)
		},
	})

	// First click with left button.
	arena1 := NewArena()
	down1 := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	rec.AddPointer(down1, arena1)
	arena1.Close(1)
	up1 := makePointerEvent(PointerUp, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 50*time.Millisecond)
	rec.HandleEvent(up1)

	// Second click with right button.
	t2 := 150 * time.Millisecond
	arena2 := NewArena()
	down2 := makePointerEvent(PointerDown, 2, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonRight, t2)
	rec.AddPointer(down2, arena2)
	arena2.Close(2)
	up2 := makePointerEvent(PointerUp, 2, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonRight, t2+50*time.Millisecond)
	rec.HandleEvent(up2)

	if len(clicks) < 2 {
		t.Fatalf("got %d clicks, want 2", len(clicks))
	}
	if clicks[1] != 1 {
		t.Errorf("second click count = %d, want 1 (reset due to button mismatch)", clicks[1])
	}
}

func TestClickRecognizer_DistanceResetTouch(t *testing.T) {
	var clicks []int
	rec := NewClickRecognizer(ClickConfig{
		OnClick: func(d ClickDetails) {
			clicks = append(clicks, d.ClickCount)
		},
	})

	// First tap.
	arena1 := NewArena()
	down1 := makePointerEvent(PointerDown, 1, PointerTypeTouch, geometry.Pt(50, 50), event.ButtonLeft, 0)
	rec.AddPointer(down1, arena1)
	arena1.Close(1)
	up1 := makePointerEvent(PointerUp, 1, PointerTypeTouch, geometry.Pt(50, 50), event.ButtonLeft, 50*time.Millisecond)
	rec.HandleEvent(up1)

	// Second tap too far away.
	t2 := 150 * time.Millisecond
	arena2 := NewArena()
	down2 := makePointerEvent(PointerDown, 2, PointerTypeTouch, geometry.Pt(200, 200), event.ButtonLeft, t2)
	rec.AddPointer(down2, arena2)
	arena2.Close(2)
	up2 := makePointerEvent(PointerUp, 2, PointerTypeTouch, geometry.Pt(200, 200), event.ButtonLeft, t2+50*time.Millisecond)
	rec.HandleEvent(up2)

	if len(clicks) < 2 {
		t.Fatalf("got %d clicks, want 2", len(clicks))
	}
	if clicks[1] != 1 {
		t.Errorf("second click count = %d, want 1 (reset due to distance)", clicks[1])
	}
}

func TestClickRecognizer_MouseNoDistanceConstraint(t *testing.T) {
	var clicks []int
	rec := NewClickRecognizer(ClickConfig{
		OnClick: func(d ClickDetails) {
			clicks = append(clicks, d.ClickCount)
		},
	})

	// First click.
	arena1 := NewArena()
	down1 := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	rec.AddPointer(down1, arena1)
	arena1.Close(1)
	up1 := makePointerEvent(PointerUp, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 50*time.Millisecond)
	rec.HandleEvent(up1)

	// Second click at distant position (mouse: no distance constraint).
	t2 := 150 * time.Millisecond
	arena2 := NewArena()
	down2 := makePointerEvent(PointerDown, 2, PointerTypeMouse, geometry.Pt(500, 500), event.ButtonLeft, t2)
	rec.AddPointer(down2, arena2)
	arena2.Close(2)
	up2 := makePointerEvent(PointerUp, 2, PointerTypeMouse, geometry.Pt(500, 500), event.ButtonLeft, t2+50*time.Millisecond)
	rec.HandleEvent(up2)

	if len(clicks) < 2 {
		t.Fatalf("got %d clicks, want 2", len(clicks))
	}
	if clicks[1] != 2 {
		t.Errorf("second click count = %d, want 2 (mouse has no distance constraint)", clicks[1])
	}
}

func TestClickRecognizer_CancelOnSlop(t *testing.T) {
	var canceled bool
	rec := NewClickRecognizer(ClickConfig{
		OnClick:       func(_ ClickDetails) { t.Error("OnClick should not be called") },
		OnClickCancel: func() { canceled = true },
	})

	arena := NewArena()
	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	rec.AddPointer(down, arena)
	arena.Close(1)

	// Move beyond precise pointer slop (1px).
	move := makePointerEvent(PointerMove, 1, PointerTypeMouse, geometry.Pt(55, 55), event.ButtonLeft, 50*time.Millisecond)
	rec.HandleEvent(move)

	if !canceled {
		t.Error("click should be canceled when pointer moves beyond slop")
	}
}

func TestClickRecognizer_OnClickDown(t *testing.T) {
	var gotDown ClickDownDetails
	rec := NewClickRecognizer(ClickConfig{
		OnClickDown: func(d ClickDownDetails) {
			gotDown = d
		},
		OnClick: func(_ ClickDetails) {},
	})

	arena := NewArena()
	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	rec.AddPointer(down, arena)

	if gotDown.ClickCount != 1 {
		t.Errorf("OnClickDown count = %d, want 1", gotDown.ClickCount)
	}
	if gotDown.Button != event.ButtonLeft {
		t.Errorf("OnClickDown button = %v, want Left", gotDown.Button)
	}
}

// TestClickRecognizer_SlopResolvesRejected verifies that when a click
// recognizer exceeds slop in a multi-recognizer arena, it resolves
// Rejected so the arena can auto-accept the remaining recognizer.
// Regression test for Issue 2: ghost member blocking auto-resolution.
func TestClickRecognizer_SlopResolvesRejected(t *testing.T) {
	// Two recognizers compete: click + drag. When the click exceeds slop,
	// it must resolve Rejected so the drag auto-accepts.
	var dragWon bool

	arena := NewArena()

	click := NewClickRecognizer(ClickConfig{
		OnClick: func(_ ClickDetails) { t.Error("click should not fire") },
	})
	drag := newMock("drag")

	// Both register for the same pointer.
	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	click.AddPointer(down, arena)
	arena.Add(1, drag)
	arena.Close(1)

	// Move beyond slop. Click should resolve Rejected.
	move := makePointerEvent(PointerMove, 1, PointerTypeMouse, geometry.Pt(55, 55), event.ButtonLeft, 50*time.Millisecond)
	arena.Route(move)

	// After click rejects, drag should be auto-accepted (only member left).
	dragWon = drag.accepted
	if !dragWon {
		t.Error("drag should be auto-accepted after click resolves Rejected on slop")
	}
}

// TestClickDragArena_MovementExceedsSlop exercises a multi-recognizer arena
// with a real ClickRecognizer and DragRecognizer. Movement exceeds slop:
// click rejects -> drag auto-accepts and fires OnDragStart.
func TestClickDragArena_MovementExceedsSlop(t *testing.T) {
	var clickFired, dragStarted bool

	arena := NewArena()

	click := NewClickRecognizer(ClickConfig{
		OnClick: func(_ ClickDetails) { clickFired = true },
	})
	drag := NewDragRecognizer(DragConfig{
		Direction:   DragDirectionPan,
		OnDragStart: func(_ DragStartDetails) { dragStarted = true },
	})

	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	click.AddPointer(down, arena)
	drag.AddPointer(down, arena)
	arena.Close(1)

	// Move beyond slop (1px for mouse). Both recognizers see the move via Route.
	move := makePointerEvent(PointerMove, 1, PointerTypeMouse, geometry.Pt(55, 55), event.ButtonLeft, 50*time.Millisecond)
	arena.Route(move)

	if clickFired {
		t.Error("click should not fire after movement exceeds slop")
	}
	if !dragStarted {
		t.Error("drag should start after movement exceeds slop and click rejects")
	}
}
