package gesture

import (
	"time"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
)

// tapDragState tracks the state machine of a TapAndDragRecognizer.
type tapDragState uint8

const (
	tapDragReady    tapDragState = iota // Waiting for pointer down
	tapDragPossible                     // Pointer down, waiting for slop or up
	tapDragDragging                     // Drag confirmed with tap count
)

// TapAndDragConfig configures a TapAndDragRecognizer.
type TapAndDragConfig struct {
	// OnTapDown is called on pointer down with the current tap count.
	OnTapDown func(details TapDragDownDetails)

	// OnTapUp is called on pointer up without exceeding drag slop.
	OnTapUp func(details TapDragUpDetails)

	// OnDragStart is called when movement exceeds slop during a tap sequence.
	OnDragStart func(details TapDragStartDetails)

	// OnDragUpdate is called for each move during a tap-drag.
	OnDragUpdate func(details TapDragUpdateDetails)

	// OnDragEnd is called when the pointer is released during a drag.
	OnDragEnd func(details TapDragEndDetails)

	// OnCancel is called if the gesture is canceled.
	OnCancel func()
}

// TapDragDownDetails carries pointer-down information with tap count.
type TapDragDownDetails struct {
	GlobalPosition      geometry.Point
	LocalPosition       geometry.Point
	ConsecutiveTapCount int // 1=single, 2=double, 3=triple
	PointerType         PointerType
	Button              event.Button
	Modifiers           event.Modifiers
}

// TapDragUpDetails carries pointer-up information with tap count.
type TapDragUpDetails struct {
	GlobalPosition      geometry.Point
	LocalPosition       geometry.Point
	ConsecutiveTapCount int
}

// TapDragStartDetails carries drag-start information with tap count.
type TapDragStartDetails struct {
	GlobalPosition      geometry.Point
	LocalPosition       geometry.Point
	ConsecutiveTapCount int
	PointerType         PointerType
}

// TapDragUpdateDetails carries drag-update information with tap count.
type TapDragUpdateDetails struct {
	GlobalPosition      geometry.Point
	LocalPosition       geometry.Point
	Delta               geometry.Point
	ConsecutiveTapCount int
}

// TapDragEndDetails carries drag-end information.
type TapDragEndDetails struct {
	Velocity            geometry.Point
	ConsecutiveTapCount int
}

// TapAndDragRecognizer combines click-count tracking with drag detection.
// Every callback receives ConsecutiveTapCount, enabling:
//   - Double-tap + drag = word-by-word selection (TextField)
//   - Triple-tap + drag = line-by-line selection (TextField)
//
// This is the Flutter TapAndDragGestureRecognizer pattern.
type TapAndDragRecognizer struct {
	RecognizerBase

	config TapAndDragConfig
	state  tapDragState

	// Current gesture tracking.
	currentPointer  int
	downPosition    geometry.Point
	downGlobalPos   geometry.Point
	downPointerType PointerType
	downButton      event.Button
	downModifiers   event.Modifiers
	downTimestamp   time.Duration
	lastPosition    geometry.Point
	lastGlobalPos   geometry.Point

	// Multi-tap tracking (same logic as ClickRecognizer).
	lastUpTimestamp     time.Duration
	lastUpPosition      geometry.Point
	lastUpButton        event.Button
	lastTapCount        int
	consecutiveTapCount int

	// Velocity tracking.
	velocity *VelocityTracker
}

// NewTapAndDragRecognizer creates a combined tap-and-drag recognizer.
func NewTapAndDragRecognizer(cfg TapAndDragConfig) *TapAndDragRecognizer {
	return &TapAndDragRecognizer{
		config:         cfg,
		state:          tapDragReady,
		velocity:       NewVelocityTracker(),
		currentPointer: noPointer,
	}
}

// AddPointer is called when a new pointer goes down.
func (r *TapAndDragRecognizer) AddPointer(ev *PointerEvent, arena *Arena) bool {
	if ev.EventType != PointerDown {
		return false
	}

	r.SetDeviceKind(ev.PointerType)
	r.StartTrackingPointer(ev.PointerID, arena, r)

	r.currentPointer = ev.PointerID
	r.downPosition = ev.Position
	r.downGlobalPos = ev.GlobalPosition
	r.downPointerType = ev.PointerType
	r.downButton = ev.Button
	r.downModifiers = ev.Modifiers()
	r.downTimestamp = ev.Timestamp
	r.lastPosition = ev.Position
	r.lastGlobalPos = ev.GlobalPosition

	r.velocity.Reset()
	r.velocity.AddPosition(ev.Timestamp, ev.GlobalPosition)

	// Compute consecutive tap count.
	r.consecutiveTapCount = r.computeTapCount(ev)
	r.state = tapDragPossible

	if r.config.OnTapDown != nil {
		r.config.OnTapDown(TapDragDownDetails{
			GlobalPosition:      ev.GlobalPosition,
			LocalPosition:       ev.Position,
			ConsecutiveTapCount: r.consecutiveTapCount,
			PointerType:         ev.PointerType,
			Button:              ev.Button,
			Modifiers:           ev.Modifiers(),
		})
	}

	return true
}

// HandleEvent processes pointer events for the tracked pointer.
func (r *TapAndDragRecognizer) HandleEvent(ev *PointerEvent) {
	if ev.PointerID != r.currentPointer {
		return
	}

	switch ev.EventType {
	case PointerMove:
		r.handleMove(ev)
	case PointerUp:
		r.handleUp(ev)
	case PointerCancel:
		r.handleCancel()
	}
}

// AcceptGesture is called when this recognizer wins the arena.
func (r *TapAndDragRecognizer) AcceptGesture(pointerID int) {
	// State transitions happen in handleMove/handleUp.
}

// RejectGesture is called when this recognizer loses the arena.
func (r *TapAndDragRecognizer) RejectGesture(pointerID int) {
	r.resetState()
	if r.config.OnCancel != nil {
		r.config.OnCancel()
	}
}

// Dispose releases resources.
func (r *TapAndDragRecognizer) Dispose() {
	r.RecognizerBase.Dispose()
}

// handleMove checks for drag start and fires updates.
func (r *TapAndDragRecognizer) handleMove(ev *PointerEvent) {
	r.velocity.AddPosition(ev.Timestamp, ev.GlobalPosition)

	switch r.state {
	case tapDragPossible:
		delta := ev.Position.Sub(r.downPosition)
		if delta.Length() > r.Slop() {
			r.state = tapDragDragging
			r.ResolvePointer(ev.PointerID, Accepted, r)
			r.fireDragStart()
			r.fireDragUpdate(ev)
		}
	case tapDragDragging:
		r.fireDragUpdate(ev)
	}
}

// handleUp completes the tap or drag.
func (r *TapAndDragRecognizer) handleUp(ev *PointerEvent) {
	r.velocity.AddPosition(ev.Timestamp, ev.GlobalPosition)

	switch r.state {
	case tapDragPossible:
		// Pointer released without dragging: this is a tap.
		r.ResolvePointer(ev.PointerID, Accepted, r)

		// Record for next multi-tap computation.
		r.lastUpTimestamp = ev.Timestamp
		r.lastUpPosition = ev.GlobalPosition
		r.lastUpButton = r.downButton
		r.lastTapCount = r.consecutiveTapCount

		if r.config.OnTapUp != nil {
			r.config.OnTapUp(TapDragUpDetails{
				GlobalPosition:      ev.GlobalPosition,
				LocalPosition:       ev.Position,
				ConsecutiveTapCount: r.consecutiveTapCount,
			})
		}

	case tapDragDragging:
		// End the drag.
		vel := r.velocity.Velocity()
		r.lastUpTimestamp = ev.Timestamp
		r.lastUpPosition = ev.GlobalPosition
		r.lastUpButton = r.downButton
		r.lastTapCount = r.consecutiveTapCount

		if r.config.OnDragEnd != nil {
			r.config.OnDragEnd(TapDragEndDetails{
				Velocity:            vel,
				ConsecutiveTapCount: r.consecutiveTapCount,
			})
		}
	}

	r.StopTrackingPointer(ev.PointerID)
	r.state = tapDragReady
	r.currentPointer = noPointer
}

// handleCancel resets the recognizer.
func (r *TapAndDragRecognizer) handleCancel() {
	r.resetState()
	if r.config.OnCancel != nil {
		r.config.OnCancel()
	}
}

// fireDragStart fires the OnDragStart callback.
func (r *TapAndDragRecognizer) fireDragStart() {
	if r.config.OnDragStart != nil {
		r.config.OnDragStart(TapDragStartDetails{
			GlobalPosition:      r.downGlobalPos,
			LocalPosition:       r.downPosition,
			ConsecutiveTapCount: r.consecutiveTapCount,
			PointerType:         r.downPointerType,
		})
	}
}

// fireDragUpdate fires the OnDragUpdate callback.
func (r *TapAndDragRecognizer) fireDragUpdate(ev *PointerEvent) {
	delta := ev.Position.Sub(r.lastPosition)
	r.lastPosition = ev.Position
	r.lastGlobalPos = ev.GlobalPosition

	if r.config.OnDragUpdate != nil {
		r.config.OnDragUpdate(TapDragUpdateDetails{
			GlobalPosition:      ev.GlobalPosition,
			LocalPosition:       ev.Position,
			Delta:               delta,
			ConsecutiveTapCount: r.consecutiveTapCount,
		})
	}
}

// computeTapCount determines the consecutive tap count for a new pointer-down.
func (r *TapAndDragRecognizer) computeTapCount(ev *PointerEvent) int {
	if r.lastTapCount == 0 {
		return 1
	}

	elapsed := ev.Timestamp - r.lastUpTimestamp
	if elapsed < DoubleTapMinTime || elapsed > DoubleTapTimeout {
		return 1
	}

	if ev.Button != r.lastUpButton {
		return 1
	}

	// For touch devices, check spatial constraint.
	if ev.PointerType.DeviceKind() == DeviceKindTouch {
		dist := ev.GlobalPosition.Distance(r.lastUpPosition)
		if dist > DoubleTapSlop {
			return 1
		}
	}

	next := r.lastTapCount + 1
	if next > MaxClickCount {
		return 1
	}
	return next
}

// resetState returns the recognizer to the ready state.
func (r *TapAndDragRecognizer) resetState() {
	if r.currentPointer != noPointer {
		r.StopTrackingPointer(r.currentPointer)
	}
	r.state = tapDragReady
	r.currentPointer = noPointer
}
