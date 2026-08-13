package gesture

import (
	"math"
	"time"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
)

// DragDirection constrains which axis the drag recognizer responds to.
type DragDirection uint8

const (
	// DragDirectionPan allows drag in both axes.
	DragDirectionPan DragDirection = iota

	// DragDirectionHorizontal restricts drag to the horizontal axis.
	DragDirectionHorizontal

	// DragDirectionVertical restricts drag to the vertical axis.
	DragDirectionVertical
)

// DragDirection string constants for goconst compliance.
const (
	dragPanStr        = "Pan"
	dragHorizontalStr = "Horizontal"
	dragVerticalStr   = "Vertical"
)

// String returns a human-readable name for the drag direction.
func (d DragDirection) String() string {
	switch d {
	case DragDirectionPan:
		return dragPanStr
	case DragDirectionHorizontal:
		return dragHorizontalStr
	case DragDirectionVertical:
		return dragVerticalStr
	default:
		return pointerUnknownStr
	}
}

// dragState tracks the state machine of a DragRecognizer.
type dragState uint8

const (
	dragReady    dragState = iota // Waiting for pointer down
	dragPossible                  // Pointer down, accumulating delta
	dragAccepted                  // Drag confirmed, firing updates
)

// DragConfig configures a DragRecognizer.
type DragConfig struct {
	// Direction constrains which axis is recognized.
	Direction DragDirection

	// OnDragStart is called when movement exceeds the slop threshold.
	OnDragStart func(details DragStartDetails)

	// OnDragUpdate is called for each pointer move during an active drag.
	OnDragUpdate func(details DragUpdateDetails)

	// OnDragEnd is called when the pointer is released during a drag.
	// Includes velocity for fling detection.
	OnDragEnd func(details DragEndDetails)

	// OnDragCancel is called if the drag is canceled.
	OnDragCancel func()
}

// DragStartDetails carries information about the start of a drag.
type DragStartDetails struct {
	GlobalPosition geometry.Point
	LocalPosition  geometry.Point
	PointerType    PointerType
	Timestamp      time.Duration
}

// DragUpdateDetails carries information about a drag movement.
type DragUpdateDetails struct {
	GlobalPosition geometry.Point
	LocalPosition  geometry.Point
	Delta          geometry.Point // Movement since last update
	PrimaryDelta   float32        // Movement along the drag axis
	Timestamp      time.Duration
}

// DragEndDetails carries information about the end of a drag.
type DragEndDetails struct {
	Velocity        geometry.Point // Pixels per second at release
	PrimaryVelocity float32        // Velocity along the drag axis
}

// DragOption configures a DragRecognizer via functional options.
type DragOption func(*DragRecognizer)

// WithDraggingSignal returns a DragOption that populates the given signal
// with the current drag state (true while dragging, false otherwise).
func WithDraggingSignal(sig state.Signal[bool]) DragOption {
	return func(r *DragRecognizer) { r.draggingSignal = sig }
}

// WithDragPositionSignal returns a DragOption that populates the given
// signal with the current drag position during an active drag.
func WithDragPositionSignal(sig state.Signal[geometry.Point]) DragOption {
	return func(r *DragRecognizer) { r.positionSignal = sig }
}

// DragRecognizer detects drag gestures (pan, vertical-only, horizontal-only).
// Replaces ad-hoc drag logic in Slider, SplitView, ScrollView, and Docking.
//
// State machine:
//
//	ready -> possible (PointerDown, accumulate delta)
//	  -> accepted (delta > slop, fire OnDragStart)
//	  -> updates (PointerMove while accepted, fire OnDragUpdate)
//	  -> ended (PointerUp, fire OnDragEnd with velocity)
type DragRecognizer struct {
	RecognizerBase

	config DragConfig
	state  dragState

	// Current gesture tracking.
	currentPointer  int
	wonArena        bool // Arena accepted this recognizer (may precede slop).
	downPosition    geometry.Point
	downGlobalPos   geometry.Point
	downPointerType PointerType
	downTimestamp   time.Duration
	lastPosition    geometry.Point
	lastGlobalPos   geometry.Point
	lastTimestamp   time.Duration

	// Velocity tracking.
	velocity *VelocityTracker

	// Signal support.
	draggingSignal state.Signal[bool]
	positionSignal state.Signal[geometry.Point]
}

// NewDragRecognizer creates a recognizer that detects drag gestures.
func NewDragRecognizer(cfg DragConfig, opts ...DragOption) *DragRecognizer {
	r := &DragRecognizer{
		config:         cfg,
		state:          dragReady,
		velocity:       NewVelocityTracker(),
		currentPointer: noPointer,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// AddPointer is called when a new pointer goes down.
func (r *DragRecognizer) AddPointer(ev *PointerEvent, arena *Arena) bool {
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
	r.lastTimestamp = ev.Timestamp
	r.state = dragPossible

	r.velocity.Reset()
	r.velocity.AddPosition(ev.Timestamp, ev.GlobalPosition)

	return true
}

// HandleEvent processes pointer events for the tracked pointer.
func (r *DragRecognizer) HandleEvent(ev *PointerEvent) {
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
// This may happen before slop is exceeded (single-member auto-accept).
// Actual drag start is deferred until slop is exceeded.
func (r *DragRecognizer) AcceptGesture(pointerID int) {
	r.wonArena = true
	// Do NOT transition to dragAccepted until slop is exceeded.
	// The drag state machine requires movement beyond slop regardless
	// of arena resolution (same behavior as Flutter monodrag.dart).
}

// RejectGesture is called when this recognizer loses the arena.
func (r *DragRecognizer) RejectGesture(pointerID int) {
	r.reset()
	if r.config.OnDragCancel != nil {
		r.config.OnDragCancel()
	}
}

// Dispose releases resources.
func (r *DragRecognizer) Dispose() {
	r.RecognizerBase.Dispose()
	r.draggingSignal = nil
	r.positionSignal = nil
}

// handleMove checks for drag start (slop exceeded) and fires updates.
func (r *DragRecognizer) handleMove(ev *PointerEvent) {
	r.velocity.AddPosition(ev.Timestamp, ev.GlobalPosition)

	switch r.state {
	case dragPossible:
		if r.exceedsSlop(ev) {
			// Slop exceeded: transition to drag.
			if !r.wonArena {
				// Still competing in the arena; request acceptance.
				r.ResolvePointer(ev.PointerID, Accepted, r)
			}
			// Whether we already won or just requested, start the drag.
			r.state = dragAccepted
			r.fireDragStart()
			r.fireDragUpdate(ev)
		}
	case dragAccepted:
		r.fireDragUpdate(ev)
	}
}

// handleUp ends the drag or rejects if never started.
func (r *DragRecognizer) handleUp(ev *PointerEvent) {
	r.velocity.AddPosition(ev.Timestamp, ev.GlobalPosition)

	switch r.state {
	case dragAccepted:
		r.fireDragEnd()
	case dragPossible:
		// Pointer released before slop was exceeded: this was not a drag.
		// Resolve as rejected so the arena can auto-resolve remaining
		// members (prevents ghost member blocking other recognizers).
		if !r.wonArena {
			r.ResolvePointer(ev.PointerID, Rejected, r)
		}
	}

	r.StopTrackingPointer(ev.PointerID)
	r.state = dragReady
	r.currentPointer = noPointer
	r.wonArena = false
}

// handleCancel resets the recognizer.
func (r *DragRecognizer) handleCancel() {
	wasDragging := r.state == dragAccepted
	r.reset()
	if wasDragging {
		if r.config.OnDragCancel != nil {
			r.config.OnDragCancel()
		}
	}
}

// exceedsSlop checks whether pointer movement exceeds the slop threshold
// along the configured drag axis.
func (r *DragRecognizer) exceedsSlop(ev *PointerEvent) bool {
	delta := ev.Position.Sub(r.downPosition)
	slop := r.Slop()

	switch r.config.Direction {
	case DragDirectionHorizontal:
		return float32(math.Abs(float64(delta.X))) > slop
	case DragDirectionVertical:
		return float32(math.Abs(float64(delta.Y))) > slop
	default: // Pan
		return delta.Length() > slop
	}
}

// fireDragStart fires the OnDragStart callback and updates signals.
func (r *DragRecognizer) fireDragStart() {
	if r.draggingSignal != nil {
		r.draggingSignal.Set(true)
	}

	if r.config.OnDragStart != nil {
		r.config.OnDragStart(DragStartDetails{
			GlobalPosition: r.downGlobalPos,
			LocalPosition:  r.downPosition,
			PointerType:    r.downPointerType,
			Timestamp:      r.downTimestamp,
		})
	}
}

// fireDragUpdate fires the OnDragUpdate callback and updates signals.
func (r *DragRecognizer) fireDragUpdate(ev *PointerEvent) {
	delta := ev.Position.Sub(r.lastPosition)
	primaryDelta := r.primaryComponent(delta)

	r.lastPosition = ev.Position
	r.lastGlobalPos = ev.GlobalPosition
	r.lastTimestamp = ev.Timestamp

	if r.positionSignal != nil {
		r.positionSignal.Set(ev.Position)
	}

	if r.config.OnDragUpdate != nil {
		r.config.OnDragUpdate(DragUpdateDetails{
			GlobalPosition: ev.GlobalPosition,
			LocalPosition:  ev.Position,
			Delta:          delta,
			PrimaryDelta:   primaryDelta,
			Timestamp:      ev.Timestamp,
		})
	}
}

// fireDragEnd fires the OnDragEnd callback and updates signals.
func (r *DragRecognizer) fireDragEnd() {
	vel := r.velocity.Velocity()
	primaryVel := r.primaryComponent(vel)

	if r.draggingSignal != nil {
		r.draggingSignal.Set(false)
	}

	if r.config.OnDragEnd != nil {
		r.config.OnDragEnd(DragEndDetails{
			Velocity:        vel,
			PrimaryVelocity: primaryVel,
		})
	}
}

// primaryComponent extracts the component along the configured drag axis.
func (r *DragRecognizer) primaryComponent(p geometry.Point) float32 {
	switch r.config.Direction {
	case DragDirectionHorizontal:
		return p.X
	case DragDirectionVertical:
		return p.Y
	default:
		return p.Length()
	}
}

// reset returns the recognizer to the ready state.
func (r *DragRecognizer) reset() {
	if r.currentPointer != noPointer {
		r.StopTrackingPointer(r.currentPointer)
	}
	r.state = dragReady
	r.currentPointer = noPointer
	r.wonArena = false
	if r.draggingSignal != nil {
		r.draggingSignal.Set(false)
	}
}
