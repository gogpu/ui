package event

import (
	"fmt"
	"time"

	"github.com/gogpu/gpucontext"
)

// IMEEventType identifies one stage of an input-method interaction.
type IMEEventType uint8

const (
	// IMECompositionStart indicates that the IME started a preedit.
	IMECompositionStart IMEEventType = iota + 1
	// IMECompositionUpdate replaces the current preedit.
	IMECompositionUpdate
	// IMECompositionEnd ends the preedit and may commit text.
	IMECompositionEnd
	// IMECanceled cancels the preedit without committing it.
	IMECanceled
	// IMEDisabled indicates that the platform disabled text input.
	IMEDisabled
	// IMEDeleteSurrounding asks the focused field to delete text around its
	// insertion point.
	IMEDeleteSurrounding
)

const (
	imeCompositionStartStr  = "CompositionStart"
	imeCompositionUpdateStr = "CompositionUpdate"
	imeCompositionEndStr    = "CompositionEnd"
	imeCanceledStr          = "Canceled"
	imeDisabledStr          = "Disabled"
	imeDeleteSurroundingStr = "DeleteSurrounding"
)

// String returns a stable human-readable event name.
func (t IMEEventType) String() string {
	switch t {
	case IMECompositionStart:
		return imeCompositionStartStr
	case IMECompositionUpdate:
		return imeCompositionUpdateStr
	case IMECompositionEnd:
		return imeCompositionEndStr
	case IMECanceled:
		return imeCanceledStr
	case IMEDisabled:
		return imeDisabledStr
	case IMEDeleteSurrounding:
		return imeDeleteSurroundingStr
	default:
		return unknownStr
	}
}

// IMEEvent is delivered to the focused text input widget.
//
// Composition ranges are UTF-8 byte offsets into Composition. The event is
// mutable only through the embedded Base handled flag; payloads are snapshots
// and must not be retained for mutation by a widget.
type IMEEvent struct {
	Base

	// IMEType identifies the lifecycle operation represented by this event.
	IMEType IMEEventType

	// Composition is the current preedit for IMECompositionUpdate.
	Composition gpucontext.IMEComposition

	// Committed is the text committed by IMECompositionEnd. It is empty for a
	// canceled or empty commit.
	Committed string

	// Delete contains a delete-surrounding request for IMEDeleteSurrounding.
	Delete gpucontext.IMEDeleteSurroundingEvent
}

// NewIMECompositionStartEvent creates a composition-start event.
func NewIMECompositionStartEvent() *IMEEvent {
	return &IMEEvent{Base: NewBase(TypeIME, ModNone), IMEType: IMECompositionStart}
}

// NewIMECompositionUpdateEvent creates a preedit update event.
func NewIMECompositionUpdateEvent(composition gpucontext.IMEComposition) *IMEEvent {
	return &IMEEvent{
		Base:        NewBase(TypeIME, ModNone),
		IMEType:     IMECompositionUpdate,
		Composition: composition,
	}
}

// NewIMECompositionEndEvent creates a preedit-end event.
func NewIMECompositionEndEvent(committed string) *IMEEvent {
	return &IMEEvent{
		Base:      NewBase(TypeIME, ModNone),
		IMEType:   IMECompositionEnd,
		Committed: committed,
	}
}

// NewIMECanceledEvent creates an event that cancels the active preedit.
func NewIMECanceledEvent() *IMEEvent {
	return &IMEEvent{Base: NewBase(TypeIME, ModNone), IMEType: IMECanceled}
}

// NewIMEDisabledEvent creates an event that clears platform text input state.
func NewIMEDisabledEvent() *IMEEvent {
	return &IMEEvent{Base: NewBase(TypeIME, ModNone), IMEType: IMEDisabled}
}

// NewIMEDeleteSurroundingEvent creates a delete-surrounding request.
func NewIMEDeleteSurroundingEvent(request gpucontext.IMEDeleteSurroundingEvent) *IMEEvent {
	return &IMEEvent{
		Base:    NewBase(TypeIME, ModNone),
		IMEType: IMEDeleteSurrounding,
		Delete:  request,
	}
}

// NewIMECompositionStart is a concise alias for
// [NewIMECompositionStartEvent].
func NewIMECompositionStart() *IMEEvent { return NewIMECompositionStartEvent() }

// NewIMECompositionUpdate is a concise alias for
// [NewIMECompositionUpdateEvent].
func NewIMECompositionUpdate(composition gpucontext.IMEComposition) *IMEEvent {
	return NewIMECompositionUpdateEvent(composition)
}

// NewIMECompositionEnd is a concise alias for [NewIMECompositionEndEvent].
func NewIMECompositionEnd(committed string) *IMEEvent {
	return NewIMECompositionEndEvent(committed)
}

// NewIMECanceled is a concise alias for [NewIMECanceledEvent].
func NewIMECanceled() *IMEEvent { return NewIMECanceledEvent() }

// NewIMEDisabled is a concise alias for [NewIMEDisabledEvent].
func NewIMEDisabled() *IMEEvent { return NewIMEDisabledEvent() }

// NewIMEDeleteSurrounding is a concise alias for
// [NewIMEDeleteSurroundingEvent].
func NewIMEDeleteSurrounding(request gpucontext.IMEDeleteSurroundingEvent) *IMEEvent {
	return NewIMEDeleteSurroundingEvent(request)
}

// IsCompositionStart reports whether this event starts a composition.
func (e *IMEEvent) IsCompositionStart() bool { return e != nil && e.IMEType == IMECompositionStart }

// IsCompositionUpdate reports whether this event updates a preedit.
func (e *IMEEvent) IsCompositionUpdate() bool { return e != nil && e.IMEType == IMECompositionUpdate }

// IsCompositionEnd reports whether this event ends a composition.
func (e *IMEEvent) IsCompositionEnd() bool { return e != nil && e.IMEType == IMECompositionEnd }

// IsCanceled reports whether this event cancels a composition.
func (e *IMEEvent) IsCanceled() bool { return e != nil && e.IMEType == IMECanceled }

// IsDisabled reports whether this event disables text input.
func (e *IMEEvent) IsDisabled() bool { return e != nil && e.IMEType == IMEDisabled }

// IsDeleteSurrounding reports whether this event requests surrounding text
// deletion.
func (e *IMEEvent) IsDeleteSurrounding() bool {
	return e != nil && e.IMEType == IMEDeleteSurrounding
}

// NewIMECompositionUpdateEventWithTime creates a preedit update with a fixed
// timestamp, primarily for deterministic event tests.
func NewIMECompositionUpdateEventWithTime(
	composition gpucontext.IMEComposition,
	t time.Time,
) *IMEEvent {
	return &IMEEvent{
		Base:        NewBaseWithTime(TypeIME, t, ModNone),
		IMEType:     IMECompositionUpdate,
		Composition: composition,
	}
}

// String returns a concise representation useful in event diagnostics.
func (e *IMEEvent) String() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf("IMEEvent{Type: %s, Committed: %q}", e.IMEType, e.Committed)
}

var _ Event = (*IMEEvent)(nil)
