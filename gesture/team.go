package gesture

// Team groups recognizers that cooperate rather than compete.
//
// Within a team, when one member would win the arena, the captain (if set)
// is given the chance to claim instead. This enables widgets like Slider
// where Tap (click-to-position) and Drag (thumb-drag) should cooperate:
// if the user starts dragging, the drag recognizer wins without waiting
// for the tap's timeout.
//
// Flutter equivalent: GestureArenaTeam.
type Team struct {
	// Captain is the preferred winner when a team member would win.
	// If nil, the original winner keeps the victory.
	Captain ArenaMember

	members []teamMember
}

// teamMember wraps a recognizer that belongs to a team. It intercepts
// AcceptGesture to redirect to the team captain if one is set.
type teamMember struct {
	inner Recognizer
	team  *Team
}

// AcceptGesture intercepts the arena accept. If the team has a captain
// and the captain is not this member's inner recognizer, the captain
// receives AcceptGesture instead and this member is rejected.
func (m *teamMember) AcceptGesture(pointerID int) {
	if m.team.Captain != nil && m.team.Captain != m.inner {
		m.team.Captain.AcceptGesture(pointerID)
		m.inner.RejectGesture(pointerID)
	} else {
		m.inner.AcceptGesture(pointerID)
	}
}

// RejectGesture delegates to the inner recognizer.
func (m *teamMember) RejectGesture(pointerID int) {
	m.inner.RejectGesture(pointerID)
}

// AddPointer delegates to the inner recognizer, ensuring that the
// teamMember wrapper (not the inner recognizer) is registered in the
// arena. This is critical: when the arena calls AcceptGesture, it must
// call it on the teamMember so the captain interception logic executes.
func (m *teamMember) AddPointer(ev *PointerEvent, arena *Arena) bool {
	// Set the member override on the inner recognizer's RecognizerBase
	// so that StartTrackingPointer registers this teamMember wrapper
	// as the arena participant instead of the inner recognizer.
	if base := recognizerBase(m.inner); base != nil {
		base.SetMemberOverride(m)
	}
	return m.inner.AddPointer(ev, arena)
}

// HandleEvent delegates to the inner recognizer.
func (m *teamMember) HandleEvent(ev *PointerEvent) {
	m.inner.HandleEvent(ev)
}

// Dispose delegates to the inner recognizer.
func (m *teamMember) Dispose() {
	m.inner.Dispose()
}

// Add adds a recognizer to this team. The returned Recognizer is a wrapper
// that intercepts arena accept to support team captain logic.
func (t *Team) Add(r Recognizer) Recognizer {
	m := &teamMember{
		inner: r,
		team:  t,
	}
	t.members = append(t.members, *m)
	return m
}

// Members returns the number of recognizers in this team.
func (t *Team) Members() int {
	return len(t.members)
}
