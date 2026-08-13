package gesture

import (
	"time"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
)

// clickState tracks the state machine of a ClickRecognizer.
type clickState uint8

const (
	clickReady    clickState = iota // Waiting for pointer down
	clickPossible                   // Pointer down, waiting for arena or movement
	clickAccepted                   // Arena won, waiting for pointer up to fire
)

// ClickConfig configures a ClickRecognizer.
type ClickConfig struct {
	// MaxClickCount caps the click count. Default: 3 (Chromium standard).
	// Set to 1 to detect only single clicks.
	MaxClickCount int

	// OnClickDown is called when the pointer goes down with the current
	// consecutive click count. Useful for visual feedback before the
	// arena resolves.
	OnClickDown func(details ClickDownDetails)

	// OnClick is called when a click sequence completes (pointer up
	// within slop and within timing window). Provides the final click count.
	OnClick func(details ClickDetails)

	// OnClickCancel is called if the click is canceled (pointer moved
	// beyond slop, arena lost to another recognizer, pointer canceled).
	OnClickCancel func()
}

// ClickDownDetails carries information about a pointer-down in a click sequence.
type ClickDownDetails struct {
	GlobalPosition geometry.Point
	LocalPosition  geometry.Point
	ClickCount     int
	PointerType    PointerType
	Button         event.Button
	Modifiers      event.Modifiers
	Timestamp      time.Duration
}

// ClickDetails carries information about a completed click.
type ClickDetails struct {
	GlobalPosition geometry.Point
	LocalPosition  geometry.Point
	ClickCount     int
	PointerType    PointerType
	Button         event.Button
	Modifiers      event.Modifiers
	Timestamp      time.Duration
}

// ClickOption configures a ClickRecognizer via functional options.
type ClickOption func(*ClickRecognizer)

// WithPressedSignal returns a ClickOption that populates the given signal
// with the pressed state (true while pointer is down, false otherwise).
func WithPressedSignal(sig state.Signal[bool]) ClickOption {
	return func(r *ClickRecognizer) { r.pressedSignal = sig }
}

// ClickRecognizer detects single-click, double-click, and triple-click
// sequences. Click count is synthesized from timing and position constraints,
// replacing the platform-dependent MouseDoubleClick event type.
//
// State machine:
//
//	ready -> possible (PointerDown, start deadline timer)
//	  -> accepted (arena won, PointerUp -> fire OnClick with ClickCount)
//	  -> rejected (moved > slop, canceled, arena lost)
type ClickRecognizer struct {
	RecognizerBase

	config ClickConfig
	state  clickState

	// Current gesture state.
	downPosition    geometry.Point
	downGlobalPos   geometry.Point
	downButton      event.Button
	downModifiers   event.Modifiers
	downTimestamp   time.Duration
	downPointerType PointerType
	currentPointer  int

	// Multi-click tracking.
	lastUpTimestamp time.Duration
	lastUpPosition  geometry.Point
	lastUpButton    event.Button
	lastClickCount  int
	clickCount      int

	// Signal support.
	pressedSignal state.Signal[bool]
}

// NewClickRecognizer creates a recognizer that detects click sequences.
func NewClickRecognizer(cfg ClickConfig, opts ...ClickOption) *ClickRecognizer {
	if cfg.MaxClickCount <= 0 {
		cfg.MaxClickCount = MaxClickCount
	}
	r := &ClickRecognizer{
		config:         cfg,
		state:          clickReady,
		currentPointer: noPointer,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// AddPointer is called when a new pointer goes down. The click recognizer
// is always interested in pointer-down events for click detection.
func (r *ClickRecognizer) AddPointer(ev *PointerEvent, arena *Arena) bool {
	if ev.EventType != PointerDown {
		return false
	}

	r.SetDeviceKind(ev.PointerType)
	r.StartTrackingPointer(ev.PointerID, arena, r)

	r.currentPointer = ev.PointerID
	r.downPosition = ev.Position
	r.downGlobalPos = ev.GlobalPosition
	r.downButton = ev.Button
	r.downModifiers = ev.Modifiers()
	r.downTimestamp = ev.Timestamp
	r.downPointerType = ev.PointerType

	// Compute click count from multi-click sequence.
	r.clickCount = r.computeClickCount(ev)
	r.state = clickPossible

	if r.pressedSignal != nil {
		r.pressedSignal.Set(true)
	}

	if r.config.OnClickDown != nil {
		r.config.OnClickDown(ClickDownDetails{
			GlobalPosition: ev.GlobalPosition,
			LocalPosition:  ev.Position,
			ClickCount:     r.clickCount,
			PointerType:    ev.PointerType,
			Button:         ev.Button,
			Modifiers:      ev.Modifiers(),
			Timestamp:      ev.Timestamp,
		})
	}

	return true
}

// HandleEvent processes pointer events for the tracked pointer.
func (r *ClickRecognizer) HandleEvent(ev *PointerEvent) {
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
func (r *ClickRecognizer) AcceptGesture(pointerID int) {
	if r.state == clickPossible {
		r.state = clickAccepted
	}
}

// RejectGesture is called when this recognizer loses the arena.
func (r *ClickRecognizer) RejectGesture(pointerID int) {
	r.reset()
	if r.config.OnClickCancel != nil {
		r.config.OnClickCancel()
	}
}

// Dispose releases resources.
func (r *ClickRecognizer) Dispose() {
	r.RecognizerBase.Dispose()
	r.pressedSignal = nil
}

// handleMove checks if the pointer has moved beyond the slop threshold.
func (r *ClickRecognizer) handleMove(ev *PointerEvent) {
	if r.state != clickPossible && r.state != clickAccepted {
		return
	}

	dist := ev.Position.Distance(r.downPosition)
	if dist > r.Slop() {
		// Moved too far; cancel this click regardless of arena state.
		// If still competing in the arena, resolve rejected so the arena
		// can auto-resolve remaining members (prevents ghost member).
		if r.state == clickPossible {
			r.ResolvePointer(r.currentPointer, Rejected, r)
		}
		r.reset()
		if r.config.OnClickCancel != nil {
			r.config.OnClickCancel()
		}
	}
}

// handleUp completes the click if the recognizer has been accepted.
func (r *ClickRecognizer) handleUp(ev *PointerEvent) {
	switch r.state {
	case clickPossible:
		// Not yet accepted by arena. Resolve as accepted (arena may auto-resolve).
		r.ResolvePointer(ev.PointerID, Accepted, r)
		// If we got accepted (state changed to clickAccepted), fire the click.
		if r.state == clickAccepted {
			r.fireClick(ev)
		}
	case clickAccepted:
		r.fireClick(ev)
	}
}

// handleCancel resets the recognizer on pointer cancellation.
func (r *ClickRecognizer) handleCancel() {
	r.reset()
	if r.config.OnClickCancel != nil {
		r.config.OnClickCancel()
	}
}

// fireClick fires the OnClick callback with the current click count.
func (r *ClickRecognizer) fireClick(ev *PointerEvent) {
	// Record for next multi-click computation.
	r.lastUpTimestamp = ev.Timestamp
	r.lastUpPosition = ev.GlobalPosition
	r.lastUpButton = r.downButton
	r.lastClickCount = r.clickCount

	details := ClickDetails{
		GlobalPosition: ev.GlobalPosition,
		LocalPosition:  ev.Position,
		ClickCount:     r.clickCount,
		PointerType:    r.downPointerType,
		Button:         r.downButton,
		Modifiers:      r.downModifiers,
		Timestamp:      ev.Timestamp,
	}

	r.StopTrackingPointer(ev.PointerID)
	r.state = clickReady

	if r.pressedSignal != nil {
		r.pressedSignal.Set(false)
	}

	if r.config.OnClick != nil {
		r.config.OnClick(details)
	}
}

// computeClickCount determines the click count for a new pointer-down event
// based on timing and position relative to the last click.
//
// State machine (from ADR-049 Appendix B):
//
//	if (elapsed < DoubleTapTimeout AND
//	    distance < DoubleTapSlop [touch only] AND
//	    elapsed > DoubleTapMinTime AND
//	    same button):
//	  clickCount = min(lastClickCount + 1, MaxClickCount)
//	else:
//	  clickCount = 1
func (r *ClickRecognizer) computeClickCount(ev *PointerEvent) int {
	if r.lastClickCount == 0 {
		return 1
	}

	elapsed := ev.Timestamp - r.lastUpTimestamp
	if elapsed < DoubleTapMinTime || elapsed > DoubleTapTimeout {
		return 1
	}

	// Button must match for multi-click.
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

	next := r.lastClickCount + 1
	if next > r.config.MaxClickCount {
		next = r.config.MaxClickCount
	}
	return next
}

// reset returns the recognizer to the ready state.
func (r *ClickRecognizer) reset() {
	if r.currentPointer != noPointer {
		r.StopTrackingPointer(r.currentPointer)
	}
	r.state = clickReady
	r.currentPointer = noPointer
	if r.pressedSignal != nil {
		r.pressedSignal.Set(false)
	}
}
