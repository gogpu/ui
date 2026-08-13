package gesture

// Recognizer is the interface for all gesture recognizers.
//
// Recognizers are stateful objects that observe a stream of PointerEvents
// and decide whether the sequence matches a specific gesture pattern
// (click, drag, long-press, pinch, etc.).
//
// Lifecycle:
//  1. AddPointer is called for each PointerDown; the recognizer decides
//     whether to compete in the arena for this pointer.
//  2. HandleEvent receives all subsequent events for tracked pointers.
//  3. The recognizer calls Arena.Resolve(Accepted) or Resolve(Rejected).
//  4. AcceptGesture/RejectGesture is called by the arena.
//  5. Dispose releases resources when the recognizer is removed.
type Recognizer interface {
	ArenaMember

	// AddPointer is called when a new pointer goes down.
	// If the recognizer is interested, it should add itself to the arena
	// and begin tracking the pointer. If not interested, it should return
	// false and will not receive further events for this pointer.
	AddPointer(ev *PointerEvent, arena *Arena) bool

	// HandleEvent processes a pointer event for a tracked pointer.
	// Called for PointerMove, PointerUp, and PointerCancel after AddPointer
	// returned true.
	HandleEvent(ev *PointerEvent)

	// Dispose releases resources. Called when the widget is unmounted.
	Dispose()
}

// RecognizerBase provides common functionality for recognizer implementations.
// Embed this in concrete recognizers.
type RecognizerBase struct {
	// arena is set when the recognizer joins an arena via StartTrackingPointer.
	arena *Arena

	// trackedPointers maps pointer IDs to arena entries.
	trackedPointers map[int]ArenaEntry

	// deviceKind is the input device classification, set from the first
	// PointerDown event. Determines slop thresholds.
	deviceKind DeviceKind

	// memberOverride, when non-nil, is registered as the arena member
	// instead of the recognizer itself. Used by Team to ensure the
	// teamMember wrapper receives AcceptGesture/RejectGesture calls
	// from the arena (enabling captain interception).
	memberOverride ArenaMember
}

// SetMemberOverride sets an ArenaMember that will be registered in the arena
// instead of the recognizer itself. This is used by Team to ensure the
// teamMember wrapper (not the inner recognizer) is the arena participant,
// so that AcceptGesture flows through the captain interception logic.
func (r *RecognizerBase) SetMemberOverride(m ArenaMember) {
	r.memberOverride = m
}

// StartTrackingPointer registers the recognizer in the arena for this pointer.
// The member parameter is the concrete recognizer (or team wrapper) that should
// be registered as the arena member. If a memberOverride is set (via
// SetMemberOverride), it takes precedence over the member parameter.
func (r *RecognizerBase) StartTrackingPointer(pointerID int, arena *Arena, member ArenaMember) {
	r.arena = arena
	if r.trackedPointers == nil {
		r.trackedPointers = make(map[int]ArenaEntry)
	}
	// If a member override is set (e.g., by Team), register the override
	// so the arena calls AcceptGesture/RejectGesture on the wrapper.
	registered := member
	if r.memberOverride != nil {
		registered = r.memberOverride
	}
	entry := arena.Add(pointerID, registered)
	r.trackedPointers[pointerID] = entry
}

// StopTrackingPointer removes tracking for a pointer.
func (r *RecognizerBase) StopTrackingPointer(pointerID int) {
	delete(r.trackedPointers, pointerID)
}

// IsTrackingPointer reports whether the recognizer is tracking the given pointer.
func (r *RecognizerBase) IsTrackingPointer(pointerID int) bool {
	_, ok := r.trackedPointers[pointerID]
	return ok
}

// ResolvePointer resolves the arena for a tracked pointer.
// If a memberOverride is set, the override is used as the member identity
// so the arena correctly matches the registered participant.
func (r *RecognizerBase) ResolvePointer(pointerID int, disposition Disposition, member ArenaMember) {
	if r.arena == nil {
		return
	}
	resolved := member
	if r.memberOverride != nil {
		resolved = r.memberOverride
	}
	r.arena.Resolve(pointerID, resolved, disposition)
}

// Slop returns the drag detection threshold for the current device kind.
func (r *RecognizerBase) Slop() float32 {
	return SlopForDevice(r.deviceKind)
}

// SetDeviceKind records the device kind from a pointer event.
func (r *RecognizerBase) SetDeviceKind(pt PointerType) {
	r.deviceKind = pt.DeviceKind()
}

// Dispose resets the base recognizer state.
func (r *RecognizerBase) Dispose() {
	r.arena = nil
	r.trackedPointers = nil
	r.memberOverride = nil
}

// baseProvider is implemented by recognizers embedding RecognizerBase.
// Used by Team to access the embedded base for setting member overrides.
type baseProvider interface {
	base() *RecognizerBase
}

// base returns a pointer to this RecognizerBase.
// Satisfies the baseProvider interface for all types that embed RecognizerBase.
func (r *RecognizerBase) base() *RecognizerBase {
	return r
}

// recognizerBase extracts the embedded RecognizerBase from a Recognizer.
// Returns nil if the recognizer does not embed RecognizerBase (e.g., a mock).
func recognizerBase(r Recognizer) *RecognizerBase {
	if bp, ok := r.(baseProvider); ok {
		return bp.base()
	}
	return nil
}
