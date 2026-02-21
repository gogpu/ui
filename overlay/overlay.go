// Package overlay provides an overlay stack for displaying content above the
// normal widget tree. Overlays are used by dropdowns, tooltips, dialogs, and
// any widget that needs to float above the main UI.
//
// The core abstraction is [Stack], owned by the Window, which manages a
// last-in-first-out collection of [Overlay] widgets. Events are dispatched to
// the top overlay before reaching the regular widget tree, and overlays are
// drawn after (on top of) the regular tree.
//
// Widgets push overlays via the [widget.OverlayManager] interface obtained from
// [widget.Context], and never import the overlay package directly. This avoids
// circular dependencies.
package overlay

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// Overlay represents content displayed above the normal widget tree.
// It extends widget.Widget with dismissal and modality support.
type Overlay interface {
	widget.Widget

	// Dismiss is called when the overlay should close (e.g. click outside, Escape key).
	Dismiss()

	// Modal returns true if the overlay blocks interaction with content below.
	// A modal overlay consumes all events that are not handled by its content.
	Modal() bool
}

// Stack manages a last-in-first-out collection of overlays for a window.
//
// The stack supports push, pop, and targeted removal. When an overlay is
// removed, all overlays above it in the stack are also removed (maintaining
// stack semantics). An optional onChange callback is invoked whenever the
// stack is modified, which typically triggers a redraw.
//
// Stack is NOT safe for concurrent access. All operations must occur on the
// main/UI thread.
type Stack struct {
	overlays []Overlay
	onChange func()
}

// NewStack creates an empty overlay stack. The onChange callback, if non-nil,
// is called whenever the stack contents change (push, pop, or remove).
func NewStack(onChange func()) *Stack {
	return &Stack{onChange: onChange}
}

// Push adds an overlay to the top of the stack.
func (s *Stack) Push(o Overlay) {
	if o == nil {
		return
	}
	s.overlays = append(s.overlays, o)
	s.notify()
}

// Pop removes and returns the top overlay. Returns nil if the stack is empty.
func (s *Stack) Pop() Overlay {
	if len(s.overlays) == 0 {
		return nil
	}
	top := s.overlays[len(s.overlays)-1]
	s.overlays[len(s.overlays)-1] = nil // clear reference for GC
	s.overlays = s.overlays[:len(s.overlays)-1]
	s.notify()
	return top
}

// Remove removes the given overlay and all overlays above it in the stack.
// This maintains stack semantics: you cannot remove an overlay from the
// middle without also removing everything on top of it.
func (s *Stack) Remove(o Overlay) {
	for i, existing := range s.overlays {
		if existing == o {
			// Clear references for GC.
			for j := i; j < len(s.overlays); j++ {
				s.overlays[j] = nil
			}
			s.overlays = s.overlays[:i]
			s.notify()
			return
		}
	}
}

// Top returns the topmost overlay, or nil if the stack is empty.
func (s *Stack) Top() Overlay {
	if len(s.overlays) == 0 {
		return nil
	}
	return s.overlays[len(s.overlays)-1]
}

// IsEmpty returns true if the stack contains no overlays.
func (s *Stack) IsEmpty() bool {
	return len(s.overlays) == 0
}

// Len returns the number of overlays in the stack.
func (s *Stack) Len() int {
	return len(s.overlays)
}

// List returns the overlay slice in bottom-to-top order.
// The returned slice should not be modified by the caller.
func (s *Stack) List() []Overlay {
	return s.overlays
}

// HandleEvent dispatches an event to the overlay stack. Events are sent to
// the top overlay first. If the top overlay is modal and does not consume
// the event, the event is still blocked from reaching the normal widget tree.
//
// Returns true if the event was consumed or blocked by a modal overlay.
func (s *Stack) HandleEvent(ctx widget.Context, e event.Event) bool {
	if len(s.overlays) == 0 {
		return false
	}
	top := s.overlays[len(s.overlays)-1]
	if top.Event(ctx, e) {
		return true
	}
	// Modal overlays block all events from reaching the widget tree,
	// even if they did not explicitly consume the event.
	return top.Modal()
}

// Layout lays out all overlays with the given window-sized constraints.
func (s *Stack) Layout(ctx widget.Context, windowSize geometry.Size) {
	constraints := geometry.Tight(windowSize)
	for _, o := range s.overlays {
		size := o.Layout(ctx, constraints)
		if setter, ok := o.(interface{ SetBounds(geometry.Rect) }); ok {
			setter.SetBounds(geometry.NewRect(0, 0, size.Width, size.Height))
		}
	}
}

// Draw draws all overlays in bottom-to-top order.
func (s *Stack) Draw(ctx widget.Context, canvas widget.Canvas) {
	for _, o := range s.overlays {
		o.Draw(ctx, canvas)
	}
}

// notify calls the onChange callback if set.
func (s *Stack) notify() {
	if s.onChange != nil {
		s.onChange()
	}
}
