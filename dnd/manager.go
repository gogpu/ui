package dnd

import (
	"sync"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
)

// targetEntry associates a DropTarget with its current bounds.
type targetEntry struct {
	target DropTarget
	bounds geometry.Rect
}

// Manager coordinates drag and drop operations across the widget tree.
//
// It handles the full drag lifecycle: detection (mouse press + threshold),
// session management, target enter/over/leave dispatch, and drop/cancel.
//
// Manager is safe for concurrent access. All exported methods are
// protected by a mutex.
type Manager struct {
	mu sync.Mutex

	// session is the currently active drag session, or nil if idle.
	session *Session

	// targets holds all registered drop targets and their bounds.
	targets []targetEntry

	// pendingSource tracks a potential drag source before the threshold is met.
	pendingSource DragSource

	// pendingPos is the mouse press position for threshold detection.
	pendingPos geometry.Point

	// pendingActive is true when we are tracking a potential drag (mouse pressed).
	pendingActive bool
}

// NewManager creates a new drag and drop Manager.
func NewManager() *Manager {
	return &Manager{}
}

// RegisterTarget registers a DropTarget with its bounds for hit testing.
//
// If the target is already registered, its bounds are updated.
func (m *Manager) RegisterTarget(target DropTarget, bounds geometry.Rect) {
	if target == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, entry := range m.targets {
		if entry.target == target {
			m.targets[i].bounds = bounds
			return
		}
	}
	m.targets = append(m.targets, targetEntry{target: target, bounds: bounds})
}

// UnregisterTarget removes a DropTarget from the manager.
//
// If the target is currently the hover target of an active drag, DragLeave
// is called before removal.
func (m *Manager) UnregisterTarget(target DropTarget) {
	if target == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	// If removing the current hover target, notify it.
	if m.session != nil && m.session.target == target {
		m.session.target.DragLeave()
		m.session.target = nil
		m.session.feedback = Feedback{}
	}

	for i, entry := range m.targets {
		if entry.target != target {
			continue
		}
		lastIdx := len(m.targets) - 1
		m.targets[i] = m.targets[lastIdx]
		m.targets[lastIdx] = targetEntry{} // clear for GC
		m.targets = m.targets[:lastIdx]
		return
	}
}

// UpdateTargetBounds updates the bounds for a registered target.
//
// This should be called after layout changes to keep hit testing accurate.
func (m *Manager) UpdateTargetBounds(target DropTarget, bounds geometry.Rect) {
	if target == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, entry := range m.targets {
		if entry.target == target {
			m.targets[i].bounds = bounds
			return
		}
	}
}

// IsDragging returns true if a drag operation is currently active.
func (m *Manager) IsDragging() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.session != nil && m.session.active
}

// CurrentSession returns the active drag session, or nil if no drag is active.
//
// The returned Session should not be stored beyond the current event cycle,
// as it may be invalidated when the drag ends.
func (m *Manager) CurrentSession() *Session {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.session
}

// HandleMouseEvent processes a mouse event for drag and drop.
//
// Returns true if the event was consumed by drag/drop handling.
// The source parameter is the widget under the cursor for drag initiation;
// it may be nil if no DragSource is under the cursor.
func (m *Manager) HandleMouseEvent(evt *event.MouseEvent, source DragSource) bool {
	if evt == nil {
		return false
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	switch evt.MouseType {
	case event.MousePress:
		m.handlePress(evt, source)
		return false
	case event.MouseMove, event.MouseDrag:
		return m.handleMove(evt)
	case event.MouseRelease:
		return m.handleRelease(evt)
	default:
		return false
	}
}

// HandleKeyEvent processes a key event for drag cancellation.
//
// Returns true if the event was consumed (Escape pressed during active drag).
func (m *Manager) HandleKeyEvent(evt *event.KeyEvent) bool {
	if evt == nil {
		return false
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	if evt.KeyType == event.KeyPress && evt.Key == event.KeyEscape {
		if m.session != nil && m.session.active {
			m.cancelDrag()
			return true
		}
		// Also cancel pending drag detection.
		if m.pendingActive {
			m.pendingActive = false
			m.pendingSource = nil
			return true
		}
	}
	return false
}

// Cancel cancels the active drag operation, if any.
//
// This is equivalent to the user pressing Escape during a drag.
func (m *Manager) Cancel() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cancelDrag()
}

// TargetCount returns the number of registered drop targets.
func (m *Manager) TargetCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.targets)
}

// handlePress starts tracking a potential drag.
func (m *Manager) handlePress(evt *event.MouseEvent, source DragSource) {
	// Only left button initiates drag.
	if evt.Button != event.ButtonLeft {
		return
	}
	if source == nil {
		return
	}

	m.pendingSource = source
	m.pendingPos = evt.Position
	m.pendingActive = true
	// Don't consume press — widgets still need it for clicks.
}

// handleMove checks threshold and dispatches drag events.
func (m *Manager) handleMove(evt *event.MouseEvent) bool {
	// Active drag: dispatch to targets.
	if m.session != nil && m.session.active {
		m.session.currentPos = evt.Position
		m.updateTarget(evt.Position)
		return true
	}

	// Pending drag: check threshold.
	if !m.pendingActive {
		return false
	}
	dist := evt.Position.Distance(m.pendingPos)
	if dist < dragThreshold {
		return false
	}

	// Threshold exceeded — start drag.
	data, allow := m.pendingSource.DragStart(m.pendingPos)
	m.pendingActive = false
	if !allow {
		m.pendingSource = nil
		return false
	}

	m.session = &Session{
		data:       data,
		source:     m.pendingSource,
		startPos:   m.pendingPos,
		currentPos: evt.Position,
		active:     true,
	}
	m.pendingSource = nil
	m.updateTarget(evt.Position)
	return true
}

// handleRelease completes or cancels the drag.
func (m *Manager) handleRelease(evt *event.MouseEvent) bool {
	// Clear pending drag on any release.
	if m.pendingActive {
		m.pendingActive = false
		m.pendingSource = nil
	}

	if m.session == nil || !m.session.active {
		return false
	}

	accepted := false
	if m.session.target != nil {
		accepted = m.session.target.Drop(m.session.data, evt.Position)
	}

	m.session.source.DragEnd(accepted)
	m.session.active = false
	m.session = nil
	return true
}

// cancelDrag cancels an active drag session.
func (m *Manager) cancelDrag() {
	if m.session == nil || !m.session.active {
		return
	}
	if m.session.target != nil {
		m.session.target.DragLeave()
	}
	m.session.source.DragEnd(false)
	m.session.active = false
	m.session = nil
}

// updateTarget finds the target under the cursor and dispatches enter/over/leave.
func (m *Manager) updateTarget(pos geometry.Point) {
	var hit DropTarget
	for _, entry := range m.targets {
		if entry.bounds.Contains(pos) && entry.target.CanAccept(m.session.data) {
			hit = entry.target
			break
		}
	}

	prev := m.session.target
	if prev == hit {
		// Same target — dispatch DragOver.
		if hit != nil {
			effect := hit.DragOver(m.session.data, pos)
			m.session.feedback = Feedback{Effect: effect}
		}
		return
	}

	// Target changed.
	if prev != nil {
		prev.DragLeave()
	}
	m.session.target = hit
	if hit != nil {
		hit.DragEnter(m.session.data)
		effect := hit.DragOver(m.session.data, pos)
		m.session.feedback = Feedback{Effect: effect}
	} else {
		m.session.feedback = Feedback{}
	}
}
