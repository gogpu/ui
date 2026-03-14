package dnd

import "github.com/gogpu/ui/geometry"

// Session represents an active drag operation.
//
// A Session is created by the Manager when a drag is initiated (mouse press
// on a DragSource followed by movement beyond the drag threshold). It tracks
// the drag data, source widget, positions, and current hover target.
//
// Session is not safe for concurrent access. It is managed exclusively
// by the Manager, which provides its own synchronization.
type Session struct {
	data       DragData
	source     DragSource
	startPos   geometry.Point
	currentPos geometry.Point
	active     bool
	target     DropTarget
	feedback   Feedback
}

// Data returns the drag data for this session.
func (s *Session) Data() DragData {
	return s.data
}

// Source returns the DragSource that initiated this session.
func (s *Session) Source() DragSource {
	return s.source
}

// StartPos returns the position where the drag was initiated.
func (s *Session) StartPos() geometry.Point {
	return s.startPos
}

// CurrentPos returns the current cursor position during the drag.
func (s *Session) CurrentPos() geometry.Point {
	return s.currentPos
}

// IsActive returns true if the drag session is currently active.
func (s *Session) IsActive() bool {
	return s.active
}

// CurrentTarget returns the DropTarget currently under the cursor, or nil.
func (s *Session) CurrentTarget() DropTarget {
	return s.target
}

// Feedback returns the current visual feedback for the drag.
func (s *Session) Feedback() Feedback {
	return s.feedback
}
