package gesture

// Disposition is the result of a recognizer's arena evaluation.
type Disposition uint8

const (
	// Accepted indicates the recognizer claims victory in the arena.
	Accepted Disposition = iota

	// Rejected indicates the recognizer withdraws from the arena.
	Rejected
)

// ArenaEntry is a handle returned by Arena.Add, used for tracking
// a member's registration in the arena.
type ArenaEntry struct {
	// PointerID is the pointer this entry is registered for.
	PointerID int

	// Member is the arena participant.
	Member ArenaMember
}

// ArenaMember is the interface that all gesture arena participants implement.
// Each recognizer that wants to claim a pointer sequence registers as an
// ArenaMember in the arena for that pointer's ID.
type ArenaMember interface {
	// AcceptGesture is called when this member wins the arena.
	// The member should commit its gesture (fire callbacks, transition state).
	AcceptGesture(pointerID int)

	// RejectGesture is called when this member loses the arena.
	// The member should reset its internal state and release resources.
	RejectGesture(pointerID int)
}

// arenaState tracks the state of a single pointer's arena competition.
type arenaState struct {
	members     []ArenaMember
	isOpen      bool // true during PointerDown dispatch, false after Close
	isHeld      bool // true when Hold is active, prevents Sweep
	eagerWinner ArenaMember
	resolved    bool // true after a winner has been declared
}

// Arena manages gesture disambiguation for a single window.
//
// When a PointerDown event occurs, all interested recognizers add themselves
// to the arena for that pointer ID. As pointer events arrive, recognizers
// evaluate whether the gesture matches their pattern. A recognizer calls
// Resolve(Accepted) to claim victory or Resolve(Rejected) to withdraw.
//
// Resolution rules (Flutter GestureArenaManager protocol):
//   - Resolve(Accepted) while arena is open: store as eager winner.
//   - Resolve(Rejected): remove member, call RejectGesture.
//   - Arena closes (end of PointerDown dispatch): if eager winner exists, it
//     wins; if 1 member remains, it wins.
//   - Sweep (after PointerUp): first remaining member wins (last resort).
//   - Hold/Release: prevents sweep (used by multi-tap between taps).
type Arena struct {
	arenas map[int]*arenaState

	// pendingRoute maps pointer IDs to members that should receive events.
	// This is populated when members are added and used by Route.
	pendingRoute map[int][]ArenaMember
}

// NewArena creates a new gesture arena.
func NewArena() *Arena {
	return &Arena{
		arenas:       make(map[int]*arenaState),
		pendingRoute: make(map[int][]ArenaMember),
	}
}

// Add registers a member in the arena for the given pointer ID.
// Must be called during PointerDown dispatch; the arena closes at end of dispatch.
// Returns an ArenaEntry for tracking.
func (a *Arena) Add(pointerID int, member ArenaMember) ArenaEntry {
	state := a.getOrCreateState(pointerID)
	if state.resolved {
		// Arena already resolved for this pointer; reject immediately.
		member.RejectGesture(pointerID)
		return ArenaEntry{PointerID: pointerID, Member: member}
	}

	state.members = append(state.members, member)
	a.pendingRoute[pointerID] = append(a.pendingRoute[pointerID], member)

	return ArenaEntry{PointerID: pointerID, Member: member}
}

// Close marks the arena for a pointer as closed (no more members can join).
// Called at the end of PointerDown dispatch. If exactly one member remains
// or an eager winner exists, resolves immediately.
func (a *Arena) Close(pointerID int) {
	state, ok := a.arenas[pointerID]
	if !ok {
		return
	}
	state.isOpen = false
	a.tryResolve(pointerID, state)
}

// Resolve declares the member's disposition for the given pointer ID.
func (a *Arena) Resolve(pointerID int, member ArenaMember, disposition Disposition) {
	state, ok := a.arenas[pointerID]
	if !ok || state.resolved {
		return
	}

	switch disposition {
	case Accepted:
		if state.isOpen {
			// Arena still open: store as eager winner, resolve when closed.
			state.eagerWinner = member
		} else {
			// Arena closed: this member wins immediately.
			a.resolveWinner(pointerID, state, member)
		}
	case Rejected:
		a.removeMember(pointerID, state, member)
		member.RejectGesture(pointerID)
		// After removal, check if auto-resolve is possible.
		if !state.isOpen {
			a.tryResolve(pointerID, state)
		}
	}
}

// Hold prevents the arena from sweeping for the given pointer ID.
// Used by multi-click recognizers between taps to defer resolution.
func (a *Arena) Hold(pointerID int) {
	if state, ok := a.arenas[pointerID]; ok {
		state.isHeld = true
	}
}

// Release allows the arena to sweep again for the given pointer ID.
// If sweep was pending, it executes immediately.
func (a *Arena) Release(pointerID int) {
	state, ok := a.arenas[pointerID]
	if !ok {
		return
	}
	state.isHeld = false
}

// Sweep resolves all open arenas: first remaining member wins.
// Called after PointerUp dispatch completes. Does nothing if the arena
// is held.
//
// After sweep resolution, the full cleanup (including pendingRoute)
// is performed since the pointer sequence is complete.
func (a *Arena) Sweep(pointerID int) {
	state, ok := a.arenas[pointerID]
	if !ok || state.resolved || state.isHeld {
		// If no arena state but pendingRoute exists (already resolved
		// via Close), clean up the route now that PointerUp is done.
		if !ok {
			delete(a.pendingRoute, pointerID)
		}
		return
	}

	if len(state.members) > 0 {
		winner := state.members[0]
		a.resolveWinner(pointerID, state, winner)
	} else {
		// No members left; clean up.
		a.cleanup(pointerID)
	}

	// Full cleanup after sweep — pointer sequence is complete.
	delete(a.pendingRoute, pointerID)
}

// Route dispatches a pointer event to all members tracking the given pointer.
// Called for PointerMove, PointerUp, and PointerCancel events.
func (a *Arena) Route(ev *PointerEvent) {
	members, ok := a.pendingRoute[ev.PointerID]
	if !ok {
		return
	}
	// Iterate over a copy of the slice because HandleEvent may modify the arena.
	routeMembers := make([]ArenaMember, len(members))
	copy(routeMembers, members)
	for _, m := range routeMembers {
		if r, ok := m.(Recognizer); ok {
			r.HandleEvent(ev)
		}
	}
}

// MemberCount returns the number of members currently in the arena for a pointer.
// Returns 0 if no arena exists for the pointer.
func (a *Arena) MemberCount(pointerID int) int {
	state, ok := a.arenas[pointerID]
	if !ok {
		return 0
	}
	return len(state.members)
}

// IsResolved reports whether the arena for the given pointer has been resolved.
func (a *Arena) IsResolved(pointerID int) bool {
	state, ok := a.arenas[pointerID]
	if !ok {
		return false
	}
	return state.resolved
}

// IsHeld reports whether the arena for the given pointer is held.
func (a *Arena) IsHeld(pointerID int) bool {
	state, ok := a.arenas[pointerID]
	if !ok {
		return false
	}
	return state.isHeld
}

// getOrCreateState returns the arena state for the given pointer, creating
// a new open arena if one does not exist.
func (a *Arena) getOrCreateState(pointerID int) *arenaState {
	state, ok := a.arenas[pointerID]
	if !ok {
		state = &arenaState{isOpen: true}
		a.arenas[pointerID] = state
	}
	return state
}

// tryResolve attempts to auto-resolve the arena if conditions are met:
// - Eager winner exists, or
// - Exactly one member remains.
func (a *Arena) tryResolve(pointerID int, state *arenaState) {
	if state.resolved {
		return
	}

	if state.eagerWinner != nil {
		a.resolveWinner(pointerID, state, state.eagerWinner)
		return
	}

	if len(state.members) == 1 {
		a.resolveWinner(pointerID, state, state.members[0])
	}
}

// resolveWinner declares the winner and rejects all other members.
//
// After resolution, the arena state is marked resolved and losers are
// rejected, but pendingRoute is preserved so the winning recognizer
// continues to receive PointerMove/Up/Cancel events via Route.
// Full cleanup happens in Sweep (after PointerUp) or when the last
// member is removed.
func (a *Arena) resolveWinner(pointerID int, state *arenaState, winner ArenaMember) {
	state.resolved = true

	// Reject all losers.
	for _, m := range state.members {
		if m != winner {
			m.RejectGesture(pointerID)
		}
	}

	// Retain only the winner in the route list so subsequent
	// PointerMove/Up events reach it via Route.
	a.pendingRoute[pointerID] = []ArenaMember{winner}

	// Accept the winner.
	winner.AcceptGesture(pointerID)

	// Clean up arena state (no longer needed for resolution decisions),
	// but keep pendingRoute alive for event delivery.
	delete(a.arenas, pointerID)
}

// removeMember removes a member from the arena's member list.
func (a *Arena) removeMember(pointerID int, state *arenaState, member ArenaMember) {
	for i, m := range state.members {
		if m == member {
			state.members = append(state.members[:i], state.members[i+1:]...)
			break
		}
	}
	// Also remove from pendingRoute.
	if route, ok := a.pendingRoute[pointerID]; ok {
		for i, m := range route {
			if m == member {
				a.pendingRoute[pointerID] = append(route[:i], route[i+1:]...)
				break
			}
		}
	}
}

// cleanup removes all state for a resolved pointer.
func (a *Arena) cleanup(pointerID int) {
	delete(a.arenas, pointerID)
	delete(a.pendingRoute, pointerID)
}
