package dnd

import "github.com/gogpu/ui/geometry"

// dragThreshold is the minimum distance in pixels the mouse must move
// from the press point before a drag operation begins. This prevents
// accidental drags during normal clicks.
const dragThreshold float32 = 5.0

// DropEffect indicates the visual feedback for a drop operation.
type DropEffect uint8

// DropEffect constants.
const (
	// DropNone indicates the target does not accept the drop.
	DropNone DropEffect = iota

	// DropCopy indicates the dragged data will be copied to the target.
	DropCopy

	// DropMove indicates the dragged data will be moved to the target.
	DropMove

	// DropLink indicates the target will create a link/reference to the data.
	DropLink
)

// dropEffectNames maps each DropEffect to its human-readable name.
var dropEffectNames = [...]string{
	DropNone: dropEffectNoneStr,
	DropCopy: dropEffectCopyStr,
	DropMove: dropEffectMoveStr,
	DropLink: dropEffectLinkStr,
}

// String constants for DropEffect.String to satisfy goconst.
const (
	dropEffectNoneStr    = "None"
	dropEffectCopyStr    = "Copy"
	dropEffectMoveStr    = "Move"
	dropEffectLinkStr    = "Link"
	dropEffectUnknownStr = "Unknown"
)

// String returns a human-readable name for the drop effect.
func (e DropEffect) String() string {
	if int(e) < len(dropEffectNames) {
		return dropEffectNames[e]
	}
	return dropEffectUnknownStr
}

// DragData carries the payload during a drag operation.
//
// Kind identifies the type of data being dragged (e.g., "text", "widget",
// "file"). DropTargets use Kind to determine whether they can accept
// the data without inspecting the payload.
type DragData struct {
	// Kind is a type identifier for the dragged data (e.g., "text", "widget", "file").
	Kind string

	// Payload is the actual data being dragged. The concrete type depends on Kind.
	Payload any
}

// DragSource is implemented by widgets that can initiate drag operations.
//
// Widgets implement this interface to participate as drag sources. The Manager
// detects DragSource via type assertion when processing mouse press events.
type DragSource interface {
	// DragStart is called when a drag operation begins (after the mouse
	// moves beyond the drag threshold). It returns the drag data and
	// whether the drag should be allowed.
	DragStart(pos geometry.Point) (DragData, bool)

	// DragEnd is called when a drag operation completes. The accepted
	// parameter is true if a DropTarget accepted the data, false if the
	// drag was canceled or no target accepted it.
	DragEnd(accepted bool)
}

// DropTarget is implemented by widgets that can accept dropped data.
//
// Widgets implement this interface to participate as drop targets. Targets
// are registered with the Manager along with their bounds.
type DropTarget interface {
	// CanAccept returns true if this target accepts data of the given kind.
	// This is called before DragEnter to quickly filter incompatible targets.
	CanAccept(data DragData) bool

	// DragEnter is called when the drag cursor enters the target's bounds.
	DragEnter(data DragData)

	// DragOver is called while the drag cursor moves over the target.
	// Returns the DropEffect to display as visual feedback.
	DragOver(data DragData, pos geometry.Point) DropEffect

	// DragLeave is called when the drag cursor leaves the target's bounds.
	DragLeave()

	// Drop is called when the data is released over the target.
	// Returns true if the target accepted the data.
	Drop(data DragData, pos geometry.Point) bool
}

// Feedback carries visual feedback information during an active drag.
type Feedback struct {
	// Effect is the current drop effect at the cursor position.
	Effect DropEffect

	// Label is optional text displayed near the cursor during drag.
	Label string
}
