// Package dnd provides drag and drop infrastructure for the gogpu/ui toolkit.
//
// The package defines two optional interfaces that widgets can implement:
//
//   - [DragSource] for widgets that can be dragged
//   - [DropTarget] for widgets that accept drops
//
// A [Manager] coordinates drag operations across the widget tree. It detects
// drag initiation (mouse press + 5px movement threshold), manages the active
// [Session], and dispatches enter/over/leave/drop events to registered
// [DropTarget] widgets.
//
// # Usage
//
// Widgets opt into drag and drop via interface implementation and type assertion:
//
//	if src, ok := w.(dnd.DragSource); ok {
//	    data, allow := src.DragStart(pos)
//	    // ...
//	}
//
// The Manager handles the full lifecycle:
//
//	mgr := dnd.NewManager()
//	mgr.RegisterTarget(myDropZone, myDropZone.Bounds())
//	// In event loop:
//	if mgr.HandleMouseEvent(mouseEvt) {
//	    // event consumed by drag/drop
//	}
//
// # Drag Detection
//
// Drag starts when:
//  1. MousePress on a widget implementing [DragSource]
//  2. Mouse moves more than 5px while pressed (prevents accidental drags)
//  3. [DragSource.DragStart] returns true
//
// # Cancellation
//
// An active drag can be canceled by pressing the Escape key. The Manager
// provides [Manager.HandleKeyEvent] for this purpose. Cancellation invokes
// [DragSource.DragEnd] with accepted=false.
//
// # Thread Safety
//
// [Manager] is safe for concurrent access. All methods are protected by a mutex.
package dnd
