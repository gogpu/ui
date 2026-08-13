package gesture

import (
	"time"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
)

// longPressState tracks the state machine of a LongPressRecognizer.
type longPressState uint8

const (
	longPressReady    longPressState = iota // Waiting for pointer down
	longPressPossible                       // Pointer down, waiting for timeout
	longPressAccepted                       // Long press triggered
)

// LongPressConfig configures a LongPressRecognizer.
type LongPressConfig struct {
	// OnLongPressDown is called after PressTimeout (100ms) if the pointer
	// is still within slop. Used for visual feedback (ripple, highlight).
	OnLongPressDown func(details LongPressDetails)

	// OnLongPress is called when the long-press duration (500ms) is reached.
	OnLongPress func(details LongPressDetails)

	// OnLongPressMoveUpdate is called if the pointer moves after a
	// long-press has been recognized (long-press-drag).
	OnLongPressMoveUpdate func(details LongPressMoveDetails)

	// OnLongPressUp is called when the pointer is released after a long-press.
	OnLongPressUp func(details LongPressDetails)

	// OnLongPressCancel is called if the long-press is canceled.
	OnLongPressCancel func()
}

// LongPressDetails carries information about a long-press event.
type LongPressDetails struct {
	GlobalPosition geometry.Point
	LocalPosition  geometry.Point
	PointerType    PointerType
}

// LongPressMoveDetails carries movement information during a long-press-drag.
type LongPressMoveDetails struct {
	GlobalPosition geometry.Point
	LocalPosition  geometry.Point
	Delta          geometry.Point
}

// LongPressOption configures a LongPressRecognizer via functional options.
type LongPressOption func(*LongPressRecognizer)

// WithLongPressActiveSignal returns a LongPressOption that populates the
// given signal with the long-press active state.
func WithLongPressActiveSignal(sig state.Signal[bool]) LongPressOption {
	return func(r *LongPressRecognizer) { r.activeSignal = sig }
}

// LongPressRecognizer detects long-press gestures (hold without moving for
// 500ms). Required for context menus on touch devices.
//
// Timer implementation uses frame-based polling via CheckTimer, called by
// the animation scheduler. The recognizer records the PointerDown timestamp
// and checks elapsed time on each frame tick. This keeps all gesture logic
// on the main thread, avoiding concurrency issues.
type LongPressRecognizer struct {
	RecognizerBase

	config LongPressConfig
	state  longPressState

	// Current gesture tracking.
	currentPointer  int
	downPosition    geometry.Point
	downGlobalPos   geometry.Point
	downPointerType PointerType
	downTimestamp   time.Duration
	lastPosition    geometry.Point
	lastGlobalPos   geometry.Point

	// Timer state.
	pressDownFired bool // Whether OnLongPressDown has been called (PressTimeout)
	longPressFired bool // Whether OnLongPress has been called (LongPressTimeout)
	wonArena       bool // Arena accepted this recognizer (may precede timeout).

	// Signal support.
	activeSignal state.Signal[bool]
}

// NewLongPressRecognizer creates a recognizer that detects long-press gestures.
func NewLongPressRecognizer(cfg LongPressConfig, opts ...LongPressOption) *LongPressRecognizer {
	r := &LongPressRecognizer{
		config:         cfg,
		state:          longPressReady,
		currentPointer: noPointer,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// AddPointer is called when a new pointer goes down.
func (r *LongPressRecognizer) AddPointer(ev *PointerEvent, arena *Arena) bool {
	if ev.EventType != PointerDown {
		return false
	}

	r.SetDeviceKind(ev.PointerType)
	r.StartTrackingPointer(ev.PointerID, arena, r)

	r.currentPointer = ev.PointerID
	r.downPosition = ev.Position
	r.downGlobalPos = ev.GlobalPosition
	r.downPointerType = ev.PointerType
	r.downTimestamp = ev.Timestamp
	r.lastPosition = ev.Position
	r.lastGlobalPos = ev.GlobalPosition
	r.pressDownFired = false
	r.longPressFired = false
	r.state = longPressPossible

	return true
}

// HandleEvent processes pointer events for the tracked pointer.
func (r *LongPressRecognizer) HandleEvent(ev *PointerEvent) {
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

// CheckTimer checks whether the long-press timeout has been reached.
// Must be called from the animation frame loop with the current timestamp.
// This is the frame-based timer approach (no goroutines).
//
// Returns true if the recognizer needs continued animation frames.
func (r *LongPressRecognizer) CheckTimer(now time.Duration) bool {
	if r.state != longPressPossible {
		return false
	}

	elapsed := now - r.downTimestamp

	// Check PressTimeout (100ms) for visual feedback.
	if !r.pressDownFired && elapsed >= PressTimeout {
		r.pressDownFired = true
		if r.config.OnLongPressDown != nil {
			r.config.OnLongPressDown(LongPressDetails{
				GlobalPosition: r.downGlobalPos,
				LocalPosition:  r.downPosition,
				PointerType:    r.downPointerType,
			})
		}
	}

	// Check LongPressTimeout (500ms) for long-press trigger.
	if !r.longPressFired && elapsed >= LongPressTimeout {
		r.longPressFired = true
		r.state = longPressAccepted

		if r.activeSignal != nil {
			r.activeSignal.Set(true)
		}

		// Resolve as accepted in the arena.
		r.ResolvePointer(r.currentPointer, Accepted, r)

		if r.config.OnLongPress != nil {
			r.config.OnLongPress(LongPressDetails{
				GlobalPosition: r.downGlobalPos,
				LocalPosition:  r.downPosition,
				PointerType:    r.downPointerType,
			})
		}
		return false // No more animation frames needed.
	}

	return true // Continue animation frames.
}

// AcceptGesture is called when this recognizer wins the arena.
// May happen before LongPressTimeout (single-member auto-accept).
// Actual long press is deferred until timeout via CheckTimer.
func (r *LongPressRecognizer) AcceptGesture(pointerID int) {
	r.wonArena = true
}

// RejectGesture is called when this recognizer loses the arena.
func (r *LongPressRecognizer) RejectGesture(pointerID int) {
	r.reset()
	if r.config.OnLongPressCancel != nil {
		r.config.OnLongPressCancel()
	}
}

// Dispose releases resources.
func (r *LongPressRecognizer) Dispose() {
	r.RecognizerBase.Dispose()
	r.activeSignal = nil
}

// handleMove checks slop and fires long-press-drag updates.
func (r *LongPressRecognizer) handleMove(ev *PointerEvent) {
	switch r.state {
	case longPressPossible:
		// Check if pointer moved beyond slop.
		dist := ev.Position.Distance(r.downPosition)
		if dist > r.Slop() {
			// Cancel long press. Resolve rejected in the arena so other
			// recognizers can be auto-accepted (prevents ghost member).
			r.ResolvePointer(r.currentPointer, Rejected, r)
			r.reset()
			if r.config.OnLongPressCancel != nil {
				r.config.OnLongPressCancel()
			}
		}
	case longPressAccepted:
		// Long press is active; fire move updates (long-press-drag).
		delta := ev.Position.Sub(r.lastPosition)
		r.lastPosition = ev.Position
		r.lastGlobalPos = ev.GlobalPosition

		if r.config.OnLongPressMoveUpdate != nil {
			r.config.OnLongPressMoveUpdate(LongPressMoveDetails{
				GlobalPosition: ev.GlobalPosition,
				LocalPosition:  ev.Position,
				Delta:          delta,
			})
		}
	}
}

// handleUp ends the long-press gesture.
func (r *LongPressRecognizer) handleUp(ev *PointerEvent) {
	switch r.state {
	case longPressAccepted:
		if r.config.OnLongPressUp != nil {
			r.config.OnLongPressUp(LongPressDetails{
				GlobalPosition: ev.GlobalPosition,
				LocalPosition:  ev.Position,
				PointerType:    r.downPointerType,
			})
		}
	case longPressPossible:
		// Pointer released before long press timeout; reject.
		r.ResolvePointer(ev.PointerID, Rejected, r)
	}

	r.reset()
}

// handleCancel resets the recognizer.
func (r *LongPressRecognizer) handleCancel() {
	wasActive := r.state == longPressAccepted
	r.reset()
	if wasActive {
		if r.config.OnLongPressCancel != nil {
			r.config.OnLongPressCancel()
		}
	}
}

// reset returns the recognizer to the ready state.
func (r *LongPressRecognizer) reset() {
	if r.currentPointer != noPointer {
		r.StopTrackingPointer(r.currentPointer)
	}
	r.state = longPressReady
	r.currentPointer = noPointer
	r.pressDownFired = false
	r.longPressFired = false
	r.wonArena = false
	if r.activeSignal != nil {
		r.activeSignal.Set(false)
	}
}
