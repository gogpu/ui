package dnd

import "github.com/gogpu/ui/geometry"

// DragVisual describes how to render drag feedback.
//
// Applications can inspect the active Session and its Feedback to render
// custom drag visuals (ghost widgets, drop indicators, etc.). This type
// provides common visual parameters that painters can use.
type DragVisual struct {
	// CursorPos is the current cursor position during the drag.
	CursorPos geometry.Point

	// Offset is the distance from the cursor to the drag ghost's origin.
	// Typically set to the offset from the original click point within
	// the source widget.
	Offset geometry.Point

	// Effect is the current drop effect for visual feedback.
	Effect DropEffect

	// OverTarget is true when the cursor is over a valid drop target.
	OverTarget bool

	// Label is optional text to display near the cursor.
	Label string
}

// NewDragVisual creates a DragVisual from an active Session.
//
// Returns a zero DragVisual if the session is nil or inactive.
func NewDragVisual(session *Session) DragVisual {
	if session == nil || !session.active {
		return DragVisual{}
	}
	offset := session.currentPos.Sub(session.startPos)
	return DragVisual{
		CursorPos:  session.currentPos,
		Offset:     offset,
		Effect:     session.feedback.Effect,
		OverTarget: session.target != nil,
		Label:      session.feedback.Label,
	}
}
