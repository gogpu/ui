package event

import (
	"fmt"
	"time"

	"github.com/gogpu/ui/geometry"
)

// WheelEvent represents a mouse wheel scroll event.
//
// Delta contains the scroll amount in both directions:
//   - Positive Delta.Y means scrolling up (content moves down)
//   - Negative Delta.Y means scrolling down (content moves up)
//   - Positive Delta.X means scrolling right
//   - Negative Delta.X means scrolling left
//
// The delta values are typically in "lines" for discrete scroll wheels,
// or pixels for smooth scrolling (e.g., trackpads).
type WheelEvent struct {
	Base

	// Delta is the scroll amount (X for horizontal, Y for vertical).
	Delta geometry.Point

	// Position is the mouse position relative to the widget.
	Position geometry.Point

	// GlobalPosition is the mouse position in screen coordinates.
	GlobalPosition geometry.Point
}

// NewWheelEvent creates a new wheel event with the current time.
func NewWheelEvent(
	delta geometry.Point,
	position geometry.Point,
	globalPosition geometry.Point,
	mods Modifiers,
) *WheelEvent {
	return &WheelEvent{
		Base:           NewBase(TypeWheel, mods),
		Delta:          delta,
		Position:       position,
		GlobalPosition: globalPosition,
	}
}

// NewWheelEventWithTime creates a new wheel event with a specific timestamp.
func NewWheelEventWithTime(
	delta geometry.Point,
	position geometry.Point,
	globalPosition geometry.Point,
	mods Modifiers,
	t time.Time,
) *WheelEvent {
	return &WheelEvent{
		Base:           NewBaseWithTime(TypeWheel, t, mods),
		Delta:          delta,
		Position:       position,
		GlobalPosition: globalPosition,
	}
}

// DeltaX returns the horizontal scroll amount.
func (e *WheelEvent) DeltaX() float32 {
	return e.Delta.X
}

// DeltaY returns the vertical scroll amount.
func (e *WheelEvent) DeltaY() float32 {
	return e.Delta.Y
}

// IsScrollUp returns true if scrolling up (positive Y delta).
func (e *WheelEvent) IsScrollUp() bool {
	return e.Delta.Y > 0
}

// IsScrollDown returns true if scrolling down (negative Y delta).
func (e *WheelEvent) IsScrollDown() bool {
	return e.Delta.Y < 0
}

// IsScrollLeft returns true if scrolling left (negative X delta).
func (e *WheelEvent) IsScrollLeft() bool {
	return e.Delta.X < 0
}

// IsScrollRight returns true if scrolling right (positive X delta).
func (e *WheelEvent) IsScrollRight() bool {
	return e.Delta.X > 0
}

// IsHorizontal returns true if there is any horizontal scrolling.
func (e *WheelEvent) IsHorizontal() bool {
	return e.Delta.X != 0
}

// IsVertical returns true if there is any vertical scrolling.
func (e *WheelEvent) IsVertical() bool {
	return e.Delta.Y != 0
}

// String returns a human-readable representation of the event.
func (e *WheelEvent) String() string {
	return fmt.Sprintf("WheelEvent{Delta: %s, Position: %s, Mods: %s}",
		e.Delta, e.Position, e.Modifiers())
}
