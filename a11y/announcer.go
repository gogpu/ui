package a11y

// Priority represents the urgency of a live region announcement.
//
// The priority determines how screen readers handle the announcement
// relative to what they are currently speaking.
type Priority uint8

// Priority constants.
const (
	// PriorityLow indicates the announcement can wait until the screen reader
	// finishes its current speech. This is appropriate for non-urgent status
	// updates (equivalent to ARIA aria-live="polite").
	PriorityLow Priority = iota

	// PriorityHigh indicates the announcement should interrupt current speech.
	// This is appropriate for urgent messages like errors or alerts
	// (equivalent to ARIA aria-live="assertive").
	PriorityHigh
)

// String returns a human-readable name for the priority.
func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "Low"
	case PriorityHigh:
		return "High"
	default:
		return unknownStr
	}
}

// Announcer is the interface for making live region announcements to
// assistive technology.
//
// Live region announcements are used to notify screen readers of dynamic
// content changes that are not captured by the accessibility tree structure.
// For example, a chat application might announce new messages, or a form
// might announce validation errors.
//
// Platform adapters provide implementations that use the native accessibility
// APIs. When no platform adapter is registered, [NoOpAnnouncer] is used.
//
// # Example
//
//	var announcer a11y.Announcer = a11y.NoOpAnnouncer{}
//	announcer.Announce("File saved successfully", a11y.PriorityLow)
//	announcer.Announce("Error: invalid input", a11y.PriorityHigh)
type Announcer interface {
	// Announce delivers a message to assistive technology.
	//
	// The message is spoken by the screen reader according to the given priority.
	// With [PriorityLow], the message waits for the current speech to finish.
	// With [PriorityHigh], the message interrupts current speech immediately.
	//
	// This method must be safe to call from any goroutine.
	Announce(message string, priority Priority)
}

// NoOpAnnouncer is a default [Announcer] that discards all announcements.
//
// It is used as the default announcer when no platform adapter is registered.
// This ensures code that makes announcements does not need nil checks.
//
// NoOpAnnouncer is safe for concurrent use.
type NoOpAnnouncer struct{}

// Announce discards the message. This is a no-op implementation.
func (NoOpAnnouncer) Announce(_ string, _ Priority) {
	// Intentionally empty: no platform adapter available.
}

// Verify NoOpAnnouncer implements Announcer.
var _ Announcer = NoOpAnnouncer{}
