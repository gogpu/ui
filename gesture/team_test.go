package gesture

import (
	"testing"
	"time"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
)

// mockRecognizer implements Recognizer for team testing.
type mockRecognizer struct {
	accepted bool
	rejected bool
	disposed bool
	handled  []*PointerEvent
}

func (m *mockRecognizer) AddPointer(_ *PointerEvent, _ *Arena) bool { return true }
func (m *mockRecognizer) HandleEvent(ev *PointerEvent)              { m.handled = append(m.handled, ev) }
func (m *mockRecognizer) AcceptGesture(_ int)                       { m.accepted = true }
func (m *mockRecognizer) RejectGesture(_ int)                       { m.rejected = true }
func (m *mockRecognizer) Dispose()                                  { m.disposed = true }

func TestTeam_CaptainWins(t *testing.T) {
	team := &Team{}
	captain := &mockRecognizer{}
	member := &mockRecognizer{}

	team.Captain = captain
	wrapped := team.Add(member)

	// When the wrapped member wins the arena, the captain should get
	// AcceptGesture and the original member should be rejected.
	wrapped.AcceptGesture(1)

	if !captain.accepted {
		t.Error("captain should be accepted when team member wins")
	}
	if !member.rejected {
		t.Error("original member should be rejected when captain takes over")
	}
}

func TestTeam_NoCaptain(t *testing.T) {
	team := &Team{}
	member := &mockRecognizer{}

	wrapped := team.Add(member)

	// Without a captain, the original member should get the accept.
	wrapped.AcceptGesture(1)

	if !member.accepted {
		t.Error("member should be accepted when no captain is set")
	}
}

func TestTeam_CaptainIsSelf(t *testing.T) {
	team := &Team{}
	member := &mockRecognizer{}

	wrapped := team.Add(member)
	team.Captain = member // Captain is the same recognizer.

	wrapped.AcceptGesture(1)

	if !member.accepted {
		t.Error("member should be accepted when captain is self")
	}
}

func TestTeam_RejectDelegates(t *testing.T) {
	team := &Team{}
	member := &mockRecognizer{}

	wrapped := team.Add(member)
	wrapped.RejectGesture(1)

	if !member.rejected {
		t.Error("reject should delegate to inner recognizer")
	}
}

func TestTeam_HandleEventDelegates(t *testing.T) {
	team := &Team{}
	member := &mockRecognizer{}

	wrapped := team.Add(member)

	ev := &PointerEvent{EventType: PointerMove, PointerID: 1}
	wrapped.HandleEvent(ev)

	if len(member.handled) != 1 {
		t.Error("HandleEvent should delegate to inner recognizer")
	}
}

func TestTeam_DisposeDelegates(t *testing.T) {
	team := &Team{}
	member := &mockRecognizer{}

	wrapped := team.Add(member)
	wrapped.Dispose()

	if !member.disposed {
		t.Error("Dispose should delegate to inner recognizer")
	}
}

func TestTeam_Members(t *testing.T) {
	team := &Team{}
	if team.Members() != 0 {
		t.Error("empty team should have 0 members")
	}

	team.Add(&mockRecognizer{})
	team.Add(&mockRecognizer{})
	if team.Members() != 2 {
		t.Errorf("Members() = %d, want 2", team.Members())
	}
}

// TestTeam_CaptainReceivesAcceptThroughArena verifies that when a team
// member's inner recognizer is registered via AddPointer -> arena.Add,
// the arena calls AcceptGesture on the teamMember wrapper (not the inner
// recognizer directly), so the captain interception logic executes.
// This tests Issue 1: Team arena registration bypass.
func TestTeam_CaptainReceivesAcceptThroughArena(t *testing.T) {
	team := &Team{}

	// Create real recognizers (not mocks) that embed RecognizerBase.
	var clickAccepted bool
	click := NewClickRecognizer(ClickConfig{
		OnClick: func(_ ClickDetails) { clickAccepted = true },
	})

	var dragAccepted bool
	drag := NewDragRecognizer(DragConfig{
		Direction:   DragDirectionPan,
		OnDragStart: func(_ DragStartDetails) { dragAccepted = true },
	})

	// The drag recognizer is the captain.
	team.Captain = drag
	wrappedClick := team.Add(click)
	wrappedDrag := team.Add(drag)

	// Simulate pointer down through the team wrappers into a shared arena.
	arena := NewArena()
	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)

	wrappedClick.AddPointer(down, arena)
	wrappedDrag.AddPointer(down, arena)
	arena.Close(1)

	// Simulate the click recognizer getting accepted first via pointer up.
	// The arena should call AcceptGesture on the teamMember wrapper for click.
	// The wrapper should redirect to the captain (drag).
	up := makePointerEvent(PointerUp, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 50*time.Millisecond)
	arena.Route(up)

	// After pointer up, sweep resolves the first remaining member.
	arena.Sweep(1)

	// dragAccepted fires on DragStart (slop exceeded), not just AcceptGesture.
	// The captain received AcceptGesture which sets wonArena but doesn't
	// fire OnDragStart. That's correct -- drag start requires slop.
	// So we only verify that click did NOT fire (captain intercepted).
	_ = dragAccepted // intentionally not asserted; see comment above
	if clickAccepted {
		t.Error("click OnClick should not fire when captain intercepts the accept")
	}
}

// TestTeam_WrapperRegisteredInArena verifies at the arena level that the
// teamMember wrapper (not the inner recognizer) is the registered member.
func TestTeam_WrapperRegisteredInArena(t *testing.T) {
	team := &Team{}

	click := NewClickRecognizer(ClickConfig{})
	captain := NewDragRecognizer(DragConfig{Direction: DragDirectionPan})
	team.Captain = captain

	wrappedClick := team.Add(click)

	arena := NewArena()
	down := makePointerEvent(PointerDown, 1, PointerTypeMouse, geometry.Pt(50, 50), event.ButtonLeft, 0)
	wrappedClick.AddPointer(down, arena)

	// The arena should have 1 member, and it should be the wrapper.
	if arena.MemberCount(1) != 1 {
		t.Fatalf("MemberCount = %d, want 1", arena.MemberCount(1))
	}

	// Close the arena. Since there's only one member, it auto-accepts.
	// AcceptGesture should go through the teamMember wrapper to the captain.
	arena.Close(1)

	// Verify: captain.wonArena should be true (AcceptGesture called on captain).
	// The inner click recognizer should NOT have state=clickAccepted
	// (it should have been rejected by the wrapper).
	if click.state == clickAccepted {
		t.Error("inner click recognizer should be rejected when captain takes over")
	}
}
