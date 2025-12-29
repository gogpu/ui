package event

import (
	"fmt"
	"time"
)

// FocusEventType represents the specific type of focus event.
type FocusEventType uint8

// Focus event type constants.
const (
	// FocusGained indicates the widget gained keyboard focus.
	FocusGained FocusEventType = iota + 1

	// FocusLost indicates the widget lost keyboard focus.
	FocusLost
)

// String returns a human-readable name for the focus event type.
func (t FocusEventType) String() string {
	switch t {
	case FocusGained:
		return "Gained"
	case FocusLost:
		return "Lost"
	default:
		return unknownStr
	}
}

// FocusEvent represents a focus change event.
//
// Focus events are sent when a widget gains or loses keyboard focus.
// Only one widget can have focus at a time within a window.
type FocusEvent struct {
	Base

	// FocusType is the specific type of focus event.
	FocusType FocusEventType
}

// NewFocusEvent creates a new focus event with the current time.
func NewFocusEvent(focusType FocusEventType) *FocusEvent {
	return &FocusEvent{
		Base:      NewBase(TypeFocus, ModNone),
		FocusType: focusType,
	}
}

// NewFocusEventWithTime creates a new focus event with a specific timestamp.
func NewFocusEventWithTime(focusType FocusEventType, t time.Time) *FocusEvent {
	return &FocusEvent{
		Base:      NewBaseWithTime(TypeFocus, t, ModNone),
		FocusType: focusType,
	}
}

// IsGained returns true if this is a focus gained event.
func (e *FocusEvent) IsGained() bool {
	return e.FocusType == FocusGained
}

// IsLost returns true if this is a focus lost event.
func (e *FocusEvent) IsLost() bool {
	return e.FocusType == FocusLost
}

// String returns a human-readable representation of the event.
func (e *FocusEvent) String() string {
	return fmt.Sprintf("FocusEvent{Type: %s}", e.FocusType)
}
