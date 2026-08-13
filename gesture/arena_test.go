package gesture

import (
	"testing"
)

// mockMember implements ArenaMember for testing.
type mockMember struct {
	name     string
	accepted bool
	rejected bool
}

func newMock(name string) *mockMember {
	return &mockMember{name: name}
}

func (m *mockMember) AcceptGesture(_ int) { m.accepted = true }
func (m *mockMember) RejectGesture(_ int) { m.rejected = true }

func TestArena_Add(t *testing.T) {
	a := NewArena()
	m := newMock("A")

	entry := a.Add(1, m)
	if entry.PointerID != 1 {
		t.Errorf("entry.PointerID = %d, want 1", entry.PointerID)
	}
	if entry.Member != m {
		t.Error("entry.Member != m")
	}
	if a.MemberCount(1) != 1 {
		t.Errorf("MemberCount = %d, want 1", a.MemberCount(1))
	}
}

func TestArena_Close_SingleMember(t *testing.T) {
	a := NewArena()
	m := newMock("A")

	a.Add(1, m)
	a.Close(1)

	// With only one member, closing should auto-resolve.
	if !m.accepted {
		t.Error("single member should be accepted after Close")
	}
	if a.IsResolved(1) {
		// Arena should be cleaned up after resolution.
		t.Error("resolved arena should be cleaned up")
	}
}

func TestArena_Close_MultipleMembersNoEagerWinner(t *testing.T) {
	a := NewArena()
	m1 := newMock("A")
	m2 := newMock("B")

	a.Add(1, m1)
	a.Add(1, m2)
	a.Close(1)

	// Two members: no auto-resolve until one accepts/rejects or sweep.
	if m1.accepted || m2.accepted {
		t.Error("no member should be accepted with 2 members and no eager winner")
	}
}

func TestArena_ResolveAccepted_Open(t *testing.T) {
	a := NewArena()
	m1 := newMock("A")
	m2 := newMock("B")

	a.Add(1, m1)
	a.Add(1, m2)

	// Resolve Accepted while arena is still open: store as eager winner.
	a.Resolve(1, m1, Accepted)
	if m1.accepted {
		t.Error("eager winner should not be accepted until arena closes")
	}

	// Close triggers eager winner resolution.
	a.Close(1)
	if !m1.accepted {
		t.Error("eager winner should be accepted after Close")
	}
	if !m2.rejected {
		t.Error("loser should be rejected after Close")
	}
}

func TestArena_ResolveAccepted_Closed(t *testing.T) {
	a := NewArena()
	m1 := newMock("A")
	m2 := newMock("B")

	a.Add(1, m1)
	a.Add(1, m2)
	a.Close(1)

	// Resolve Accepted after arena is closed: wins immediately.
	a.Resolve(1, m1, Accepted)
	if !m1.accepted {
		t.Error("member should be accepted when resolving Accepted on closed arena")
	}
	if !m2.rejected {
		t.Error("loser should be rejected")
	}
}

func TestArena_ResolveRejected(t *testing.T) {
	a := NewArena()
	m1 := newMock("A")
	m2 := newMock("B")

	a.Add(1, m1)
	a.Add(1, m2)
	a.Close(1)

	a.Resolve(1, m1, Rejected)
	if !m1.rejected {
		t.Error("rejected member should have RejectGesture called")
	}
	// Only one member left after close -> auto-resolve.
	if !m2.accepted {
		t.Error("remaining member should be auto-accepted")
	}
}

func TestArena_Sweep(t *testing.T) {
	a := NewArena()
	m1 := newMock("A")
	m2 := newMock("B")

	a.Add(1, m1)
	a.Add(1, m2)
	a.Close(1)

	// Sweep gives victory to the first remaining member.
	a.Sweep(1)
	if !m1.accepted {
		t.Error("first member should win sweep")
	}
	if !m2.rejected {
		t.Error("second member should be rejected on sweep")
	}
}

func TestArena_HoldRelease(t *testing.T) {
	a := NewArena()
	m1 := newMock("A")
	m2 := newMock("B")

	a.Add(1, m1)
	a.Add(1, m2)
	a.Close(1)

	a.Hold(1)
	a.Sweep(1) // Should not resolve because held.

	if m1.accepted || m2.accepted {
		t.Error("sweep should not resolve a held arena")
	}

	a.Release(1)
	// Release does not automatically sweep.
	if m1.accepted || m2.accepted {
		t.Error("release alone should not sweep")
	}

	// Manual sweep after release.
	a.Sweep(1)
	if !m1.accepted {
		t.Error("first member should win sweep after release")
	}
}

func TestArena_NoMembersAfterReject(t *testing.T) {
	a := NewArena()
	m1 := newMock("A")

	a.Add(1, m1)
	a.Close(1)
	// Single member auto-resolved, arena cleaned up.

	// Sweep on non-existent arena should be a no-op.
	a.Sweep(1)
}

func TestArena_MultiplePointers(t *testing.T) {
	a := NewArena()
	m1 := newMock("ptr1")
	m2 := newMock("ptr2")

	a.Add(1, m1)
	a.Add(2, m2)
	a.Close(1)
	a.Close(2)

	// Each pointer's arena resolves independently.
	if !m1.accepted {
		t.Error("pointer 1 member should be accepted")
	}
	if !m2.accepted {
		t.Error("pointer 2 member should be accepted")
	}
}

func TestArena_AddAfterCleanup(t *testing.T) {
	a := NewArena()
	m1 := newMock("A")

	a.Add(1, m1)
	a.Close(1) // m1 auto-accepted, arena cleaned up.

	// After cleanup, adding to the same pointer ID creates a fresh arena.
	m2 := newMock("B")
	a.Add(1, m2)

	if m2.rejected {
		t.Error("member added to fresh arena (after cleanup) should not be rejected")
	}
	if a.MemberCount(1) != 1 {
		t.Errorf("MemberCount = %d, want 1", a.MemberCount(1))
	}
}

func TestArena_MemberCount(t *testing.T) {
	a := NewArena()

	if a.MemberCount(1) != 0 {
		t.Error("empty arena should have 0 members")
	}

	m1 := newMock("A")
	m2 := newMock("B")
	m3 := newMock("C")

	a.Add(1, m1)
	if a.MemberCount(1) != 1 {
		t.Errorf("MemberCount = %d, want 1", a.MemberCount(1))
	}

	a.Add(1, m2)
	a.Add(1, m3)
	if a.MemberCount(1) != 3 {
		t.Errorf("MemberCount = %d, want 3", a.MemberCount(1))
	}

	a.Close(1)
	a.Resolve(1, m1, Rejected)
	if a.MemberCount(1) != 2 {
		t.Errorf("MemberCount after reject = %d, want 2", a.MemberCount(1))
	}
}

func TestArena_IsHeld(t *testing.T) {
	a := NewArena()
	m := newMock("A")

	if a.IsHeld(1) {
		t.Error("non-existent arena should not be held")
	}

	a.Add(1, m)
	if a.IsHeld(1) {
		t.Error("new arena should not be held")
	}

	a.Hold(1)
	if !a.IsHeld(1) {
		t.Error("held arena should report IsHeld")
	}

	a.Release(1)
	if a.IsHeld(1) {
		t.Error("released arena should not be held")
	}
}

func TestArena_AllRejectThenSweep(t *testing.T) {
	a := NewArena()
	m1 := newMock("A")
	m2 := newMock("B")

	a.Add(1, m1)
	a.Add(1, m2)
	a.Close(1)

	a.Resolve(1, m1, Rejected)
	// m2 is the only member after close -> auto-accepted.
	if !m2.accepted {
		t.Error("last remaining member should be auto-accepted")
	}
}

func TestArena_Route_CallsHandleEvent(t *testing.T) {
	a := NewArena()
	rec := &routeTestRecognizer{}

	a.Add(1, rec)

	// Route should call HandleEvent on recognizers.
	ev := &PointerEvent{
		EventType: PointerMove,
		PointerID: 1,
	}
	a.Route(ev)

	if !rec.handleCalled {
		t.Error("Route should call HandleEvent on recognizers")
	}
}

// routeTestRecognizer is a minimal Recognizer for testing Route.
type routeTestRecognizer struct {
	handleCalled bool
}

func (r *routeTestRecognizer) AcceptGesture(_ int) {}
func (r *routeTestRecognizer) RejectGesture(_ int) {}
func (r *routeTestRecognizer) Dispose()            {}
func (r *routeTestRecognizer) HandleEvent(_ *PointerEvent) {
	r.handleCalled = true
}
func (r *routeTestRecognizer) AddPointer(_ *PointerEvent, _ *Arena) bool {
	return true
}
