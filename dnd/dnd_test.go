package dnd

import (
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
)

// --- Test helpers ---

// mockDragSource implements DragSource for testing.
type mockDragSource struct {
	data        DragData
	allowDrag   bool
	dragStarted bool
	dragEnded   bool
	endAccepted bool
	startPos    geometry.Point
}

func (m *mockDragSource) DragStart(pos geometry.Point) (DragData, bool) {
	m.dragStarted = true
	m.startPos = pos
	return m.data, m.allowDrag
}

func (m *mockDragSource) DragEnd(accepted bool) {
	m.dragEnded = true
	m.endAccepted = accepted
}

// mockDropTarget implements DropTarget for testing.
type mockDropTarget struct {
	acceptKinds  []string
	entered      bool
	overCount    int
	left         bool
	dropped      bool
	dropAccepted bool
	lastOverPos  geometry.Point
	lastDropPos  geometry.Point
	lastOverData DragData
	lastDropData DragData
	overEffect   DropEffect
}

func (m *mockDropTarget) CanAccept(data DragData) bool {
	for _, k := range m.acceptKinds {
		if k == data.Kind {
			return true
		}
	}
	return false
}

func (m *mockDropTarget) DragEnter(data DragData) {
	m.entered = true
}

func (m *mockDropTarget) DragOver(data DragData, pos geometry.Point) DropEffect {
	m.overCount++
	m.lastOverPos = pos
	m.lastOverData = data
	return m.overEffect
}

func (m *mockDropTarget) DragLeave() {
	m.left = true
	m.entered = false
}

func (m *mockDropTarget) Drop(data DragData, pos geometry.Point) bool {
	m.dropped = true
	m.lastDropPos = pos
	m.lastDropData = data
	return m.dropAccepted
}

// Helper to create mouse events.
func mousePress(x, y float32) *event.MouseEvent {
	return event.NewMouseEvent(
		event.MousePress,
		event.ButtonLeft,
		event.ButtonStateLeft,
		geometry.Pt(x, y),
		geometry.Pt(x, y),
		0,
	)
}

func mouseMove(x, y float32) *event.MouseEvent {
	return event.NewMouseEvent(
		event.MouseMove,
		event.ButtonNone,
		event.ButtonStateLeft,
		geometry.Pt(x, y),
		geometry.Pt(x, y),
		0,
	)
}

func mouseDrag(x, y float32) *event.MouseEvent {
	return event.NewMouseEvent(
		event.MouseDrag,
		event.ButtonNone,
		event.ButtonStateLeft,
		geometry.Pt(x, y),
		geometry.Pt(x, y),
		0,
	)
}

func mouseRelease(x, y float32) *event.MouseEvent {
	return event.NewMouseEvent(
		event.MouseRelease,
		event.ButtonLeft,
		0,
		geometry.Pt(x, y),
		geometry.Pt(x, y),
		0,
	)
}

func rightPress(x, y float32) *event.MouseEvent {
	return event.NewMouseEvent(
		event.MousePress,
		event.ButtonRight,
		event.ButtonStateRight,
		geometry.Pt(x, y),
		geometry.Pt(x, y),
		0,
	)
}

func keyEscape() *event.KeyEvent {
	return event.NewKeyEvent(event.KeyPress, event.KeyEscape, 0, 0)
}

// --- DropEffect tests ---

func TestDropEffectString(t *testing.T) {
	tests := []struct {
		name   string
		effect DropEffect
		want   string
	}{
		{"None", DropNone, "None"},
		{"Copy", DropCopy, "Copy"},
		{"Move", DropMove, "Move"},
		{"Link", DropLink, "Link"},
		{"Unknown", DropEffect(99), "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.effect.String(); got != tt.want {
				t.Errorf("DropEffect(%d).String() = %q, want %q", tt.effect, got, tt.want)
			}
		})
	}
}

// --- DragData tests ---

func TestDragData(t *testing.T) {
	data := DragData{Kind: "text", Payload: "hello"}
	if data.Kind != "text" {
		t.Errorf("Kind = %q, want %q", data.Kind, "text")
	}
	if data.Payload != "hello" {
		t.Errorf("Payload = %v, want %v", data.Payload, "hello")
	}
}

// --- Session tests ---

func TestSessionAccessors(t *testing.T) {
	src := &mockDragSource{data: DragData{Kind: "widget"}, allowDrag: true}
	tgt := &mockDropTarget{acceptKinds: []string{"widget"}, overEffect: DropCopy}

	s := &Session{
		data:       DragData{Kind: "widget", Payload: 42},
		source:     src,
		startPos:   geometry.Pt(10, 20),
		currentPos: geometry.Pt(30, 40),
		active:     true,
		target:     tgt,
		feedback:   Feedback{Effect: DropCopy, Label: "Move here"},
	}

	if s.Data().Kind != "widget" {
		t.Errorf("Data().Kind = %q, want %q", s.Data().Kind, "widget")
	}
	if s.Data().Payload != 42 {
		t.Errorf("Data().Payload = %v, want %v", s.Data().Payload, 42)
	}
	if s.Source() != src {
		t.Error("Source() mismatch")
	}
	if s.StartPos() != geometry.Pt(10, 20) {
		t.Errorf("StartPos() = %v, want (10,20)", s.StartPos())
	}
	if s.CurrentPos() != geometry.Pt(30, 40) {
		t.Errorf("CurrentPos() = %v, want (30,40)", s.CurrentPos())
	}
	if !s.IsActive() {
		t.Error("IsActive() = false, want true")
	}
	if s.CurrentTarget() != tgt {
		t.Error("CurrentTarget() mismatch")
	}
	if s.Feedback().Effect != DropCopy {
		t.Errorf("Feedback().Effect = %v, want DropCopy", s.Feedback().Effect)
	}
	if s.Feedback().Label != "Move here" {
		t.Errorf("Feedback().Label = %q, want %q", s.Feedback().Label, "Move here")
	}
}

func TestSessionInactive(t *testing.T) {
	s := &Session{}
	if s.IsActive() {
		t.Error("new session should not be active")
	}
	if s.CurrentTarget() != nil {
		t.Error("new session should have nil target")
	}
}

// --- Manager basic tests ---

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager() returned nil")
	}
	if m.IsDragging() {
		t.Error("new manager should not be dragging")
	}
	if m.CurrentSession() != nil {
		t.Error("new manager should have nil session")
	}
	if m.TargetCount() != 0 {
		t.Errorf("TargetCount() = %d, want 0", m.TargetCount())
	}
}

func TestRegisterTarget(t *testing.T) {
	m := NewManager()
	tgt := &mockDropTarget{acceptKinds: []string{"text"}}
	bounds := geometry.NewRect(0, 0, 100, 100)

	m.RegisterTarget(tgt, bounds)
	if m.TargetCount() != 1 {
		t.Errorf("TargetCount() = %d, want 1", m.TargetCount())
	}

	// Register same target again updates bounds.
	m.RegisterTarget(tgt, geometry.NewRect(10, 10, 110, 110))
	if m.TargetCount() != 1 {
		t.Errorf("TargetCount() = %d after re-register, want 1", m.TargetCount())
	}

	// Register nil does nothing.
	m.RegisterTarget(nil, bounds)
	if m.TargetCount() != 1 {
		t.Errorf("TargetCount() = %d after nil register, want 1", m.TargetCount())
	}
}

func TestUnregisterTarget(t *testing.T) {
	m := NewManager()
	tgt := &mockDropTarget{acceptKinds: []string{"text"}}
	bounds := geometry.NewRect(0, 0, 100, 100)

	m.RegisterTarget(tgt, bounds)
	m.UnregisterTarget(tgt)
	if m.TargetCount() != 0 {
		t.Errorf("TargetCount() = %d after unregister, want 0", m.TargetCount())
	}

	// Unregister non-existent target does nothing.
	m.UnregisterTarget(tgt)
	if m.TargetCount() != 0 {
		t.Errorf("TargetCount() = %d after double unregister, want 0", m.TargetCount())
	}

	// Unregister nil does nothing.
	m.UnregisterTarget(nil)
}

func TestUpdateTargetBounds(t *testing.T) {
	m := NewManager()
	tgt := &mockDropTarget{acceptKinds: []string{"text"}}
	m.RegisterTarget(tgt, geometry.NewRect(0, 0, 50, 50))

	m.UpdateTargetBounds(tgt, geometry.NewRect(10, 10, 60, 60))
	if m.TargetCount() != 1 {
		t.Errorf("TargetCount() changed after bounds update")
	}

	// Update non-existent does nothing.
	other := &mockDropTarget{acceptKinds: []string{"text"}}
	m.UpdateTargetBounds(other, geometry.NewRect(0, 0, 10, 10))
	if m.TargetCount() != 1 {
		t.Errorf("TargetCount() changed after non-existent update")
	}

	// Update nil does nothing.
	m.UpdateTargetBounds(nil, geometry.NewRect(0, 0, 10, 10))
}

// --- Full drag lifecycle tests ---

func TestDragLifecycleSuccessfulDrop(t *testing.T) {
	m := NewManager()
	src := &mockDragSource{
		data:      DragData{Kind: "text", Payload: "hello"},
		allowDrag: true,
	}
	tgt := &mockDropTarget{
		acceptKinds:  []string{"text"},
		dropAccepted: true,
		overEffect:   DropMove,
	}
	m.RegisterTarget(tgt, geometry.NewRect(50, 50, 100, 100))

	// 1. Press on source.
	consumed := m.HandleMouseEvent(mousePress(10, 10), src)
	if consumed {
		t.Error("press should not be consumed (widgets need it for clicks)")
	}
	if m.IsDragging() {
		t.Error("should not be dragging yet (threshold not met)")
	}

	// 2. Move within threshold — no drag yet.
	consumed = m.HandleMouseEvent(mouseMove(12, 12), nil)
	if consumed {
		t.Error("sub-threshold move should not be consumed")
	}
	if m.IsDragging() {
		t.Error("sub-threshold move should not start drag")
	}

	// 3. Move beyond threshold — drag starts.
	consumed = m.HandleMouseEvent(mouseMove(20, 10), nil)
	if !consumed {
		t.Error("supra-threshold move should be consumed")
	}
	if !m.IsDragging() {
		t.Error("should be dragging after threshold")
	}
	if !src.dragStarted {
		t.Error("DragStart should have been called")
	}

	// 4. Move over target.
	consumed = m.HandleMouseEvent(mouseMove(100, 100), nil)
	if !consumed {
		t.Error("drag move over target should be consumed")
	}
	if !tgt.entered {
		t.Error("DragEnter should have been called")
	}
	if tgt.overCount == 0 {
		t.Error("DragOver should have been called")
	}

	// 5. Release over target — drop.
	consumed = m.HandleMouseEvent(mouseRelease(100, 100), nil)
	if !consumed {
		t.Error("release during drag should be consumed")
	}
	if !tgt.dropped {
		t.Error("Drop should have been called")
	}
	if tgt.lastDropData.Kind != "text" {
		t.Errorf("Drop data kind = %q, want %q", tgt.lastDropData.Kind, "text")
	}
	if !src.dragEnded {
		t.Error("DragEnd should have been called")
	}
	if !src.endAccepted {
		t.Error("DragEnd accepted should be true")
	}
	if m.IsDragging() {
		t.Error("should not be dragging after drop")
	}
}

func TestDragLifecycleRejectedDrop(t *testing.T) {
	m := NewManager()
	src := &mockDragSource{
		data:      DragData{Kind: "text", Payload: "hello"},
		allowDrag: true,
	}
	tgt := &mockDropTarget{
		acceptKinds:  []string{"text"},
		dropAccepted: false, // target rejects
		overEffect:   DropNone,
	}
	m.RegisterTarget(tgt, geometry.NewRect(50, 50, 100, 100))

	// Start drag.
	m.HandleMouseEvent(mousePress(10, 10), src)
	m.HandleMouseEvent(mouseMove(20, 10), nil)

	// Move over target, then drop.
	m.HandleMouseEvent(mouseMove(100, 100), nil)
	m.HandleMouseEvent(mouseRelease(100, 100), nil)

	if !src.dragEnded {
		t.Error("DragEnd should have been called")
	}
	if src.endAccepted {
		t.Error("DragEnd accepted should be false when target rejects")
	}
}

func TestDragCancelWithEscape(t *testing.T) {
	m := NewManager()
	src := &mockDragSource{
		data:      DragData{Kind: "widget"},
		allowDrag: true,
	}
	tgt := &mockDropTarget{
		acceptKinds: []string{"widget"},
		overEffect:  DropMove,
	}
	m.RegisterTarget(tgt, geometry.NewRect(50, 50, 100, 100))

	// Start drag and move over target.
	m.HandleMouseEvent(mousePress(10, 10), src)
	m.HandleMouseEvent(mouseMove(20, 10), nil)
	m.HandleMouseEvent(mouseMove(100, 100), nil)

	if !tgt.entered {
		t.Error("should have entered target")
	}

	// Cancel with Escape.
	consumed := m.HandleKeyEvent(keyEscape())
	if !consumed {
		t.Error("Escape during drag should be consumed")
	}
	if m.IsDragging() {
		t.Error("should not be dragging after cancel")
	}
	if !tgt.left {
		t.Error("DragLeave should have been called on cancel")
	}
	if !src.dragEnded {
		t.Error("DragEnd should have been called")
	}
	if src.endAccepted {
		t.Error("DragEnd accepted should be false on cancel")
	}
}

func TestDragCancelWithManagerCancel(t *testing.T) {
	m := NewManager()
	src := &mockDragSource{
		data:      DragData{Kind: "widget"},
		allowDrag: true,
	}

	// Start drag.
	m.HandleMouseEvent(mousePress(10, 10), src)
	m.HandleMouseEvent(mouseMove(20, 10), nil)

	if !m.IsDragging() {
		t.Fatal("should be dragging")
	}

	m.Cancel()
	if m.IsDragging() {
		t.Error("should not be dragging after Cancel()")
	}
	if !src.dragEnded {
		t.Error("DragEnd should have been called")
	}
}

func TestDragSourceDisallows(t *testing.T) {
	m := NewManager()
	src := &mockDragSource{
		data:      DragData{Kind: "text"},
		allowDrag: false, // disallow
	}

	m.HandleMouseEvent(mousePress(10, 10), src)
	consumed := m.HandleMouseEvent(mouseMove(20, 10), nil)
	if consumed {
		t.Error("move should not be consumed when drag disallowed")
	}
	if m.IsDragging() {
		t.Error("should not be dragging when source disallows")
	}
	if !src.dragStarted {
		t.Error("DragStart should still have been called")
	}
}

func TestDragNoSourceOnPress(t *testing.T) {
	m := NewManager()

	// Press with nil source.
	consumed := m.HandleMouseEvent(mousePress(20, 20), nil)
	if consumed {
		t.Error("press with nil source should not be consumed")
	}

	// Move should not start drag.
	consumed = m.HandleMouseEvent(mouseMove(20, 10), nil)
	if consumed {
		t.Error("move without prior source should not be consumed")
	}
}

func TestDragRightClickIgnored(t *testing.T) {
	m := NewManager()
	src := &mockDragSource{data: DragData{Kind: "text"}, allowDrag: true}

	consumed := m.HandleMouseEvent(rightPress(10, 10), src)
	if consumed {
		t.Error("right click should not be consumed")
	}

	consumed = m.HandleMouseEvent(mouseMove(20, 10), nil)
	if consumed {
		t.Error("move after right click should not start drag")
	}
}

func TestDragTargetTransition(t *testing.T) {
	m := NewManager()
	src := &mockDragSource{
		data:      DragData{Kind: "text", Payload: "data"},
		allowDrag: true,
	}
	tgt1 := &mockDropTarget{acceptKinds: []string{"text"}, overEffect: DropCopy}
	tgt2 := &mockDropTarget{acceptKinds: []string{"text"}, overEffect: DropMove}
	m.RegisterTarget(tgt1, geometry.NewRect(0, 50, 50, 100))
	m.RegisterTarget(tgt2, geometry.NewRect(60, 50, 110, 100))

	// Start drag.
	m.HandleMouseEvent(mousePress(10, 10), src)
	m.HandleMouseEvent(mouseMove(20, 10), nil)

	// Move over tgt1.
	m.HandleMouseEvent(mouseMove(25, 75), nil)
	if !tgt1.entered {
		t.Error("tgt1 should have been entered")
	}

	// Move over tgt2.
	m.HandleMouseEvent(mouseMove(80, 75), nil)
	if !tgt1.left {
		t.Error("tgt1 DragLeave should have been called")
	}
	if !tgt2.entered {
		t.Error("tgt2 should have been entered")
	}

	// Move out of all targets.
	m.HandleMouseEvent(mouseMove(200, 200), nil)
	if !tgt2.left {
		t.Error("tgt2 DragLeave should have been called")
	}

	// Release outside any target.
	m.HandleMouseEvent(mouseRelease(200, 200), nil)
	if !src.dragEnded {
		t.Error("DragEnd should have been called")
	}
	if src.endAccepted {
		t.Error("DragEnd accepted should be false (no target)")
	}
}

func TestDragTargetCannotAccept(t *testing.T) {
	m := NewManager()
	src := &mockDragSource{
		data:      DragData{Kind: "widget"},
		allowDrag: true,
	}
	tgt := &mockDropTarget{acceptKinds: []string{"text"}} // wrong kind
	m.RegisterTarget(tgt, geometry.NewRect(50, 50, 100, 100))

	// Start drag and move over target area.
	m.HandleMouseEvent(mousePress(10, 10), src)
	m.HandleMouseEvent(mouseMove(20, 10), nil)
	m.HandleMouseEvent(mouseMove(100, 100), nil)

	if tgt.entered {
		t.Error("target should not be entered (incompatible kind)")
	}
}

func TestDragStayOnSameTarget(t *testing.T) {
	m := NewManager()
	src := &mockDragSource{
		data:      DragData{Kind: "text"},
		allowDrag: true,
	}
	tgt := &mockDropTarget{acceptKinds: []string{"text"}, overEffect: DropCopy}
	m.RegisterTarget(tgt, geometry.NewRect(50, 50, 100, 100))

	// Start drag and move to target.
	m.HandleMouseEvent(mousePress(10, 10), src)
	m.HandleMouseEvent(mouseMove(20, 10), nil)
	m.HandleMouseEvent(mouseMove(100, 100), nil)

	initialOverCount := tgt.overCount

	// Move within same target.
	m.HandleMouseEvent(mouseMove(110, 110), nil)
	if tgt.overCount != initialOverCount+1 {
		t.Errorf("DragOver count = %d, want %d", tgt.overCount, initialOverCount+1)
	}
	if tgt.left {
		t.Error("should not have left target")
	}
}

func TestDragMouseDragEventType(t *testing.T) {
	m := NewManager()
	src := &mockDragSource{
		data:      DragData{Kind: "text"},
		allowDrag: true,
	}

	m.HandleMouseEvent(mousePress(10, 10), src)

	// Use MouseDrag event type (instead of MouseMove).
	consumed := m.HandleMouseEvent(mouseDrag(20, 10), nil)
	if !consumed {
		t.Error("MouseDrag event should trigger drag start")
	}
	if !m.IsDragging() {
		t.Error("should be dragging after MouseDrag beyond threshold")
	}
}

func TestUnregisterTargetDuringDrag(t *testing.T) {
	m := NewManager()
	src := &mockDragSource{
		data:      DragData{Kind: "text"},
		allowDrag: true,
	}
	tgt := &mockDropTarget{acceptKinds: []string{"text"}, overEffect: DropCopy}
	m.RegisterTarget(tgt, geometry.NewRect(50, 50, 100, 100))

	// Start drag and enter target.
	m.HandleMouseEvent(mousePress(10, 10), src)
	m.HandleMouseEvent(mouseMove(20, 10), nil)
	m.HandleMouseEvent(mouseMove(100, 100), nil)

	if !tgt.entered {
		t.Fatal("should have entered target")
	}

	// Unregister target during drag.
	m.UnregisterTarget(tgt)
	if !tgt.left {
		t.Error("DragLeave should be called when unregistering hover target")
	}
	if m.CurrentSession().CurrentTarget() != nil {
		t.Error("session target should be nil after unregister")
	}
}

func TestHandleNilEvents(t *testing.T) {
	m := NewManager()
	if m.HandleMouseEvent(nil, nil) {
		t.Error("nil mouse event should return false")
	}
	if m.HandleKeyEvent(nil) {
		t.Error("nil key event should return false")
	}
}

func TestEscapeCancelsPendingDrag(t *testing.T) {
	m := NewManager()
	src := &mockDragSource{
		data:      DragData{Kind: "text"},
		allowDrag: true,
	}

	// Press to start pending drag, but don't exceed threshold.
	m.HandleMouseEvent(mousePress(10, 10), src)

	// Escape cancels pending.
	consumed := m.HandleKeyEvent(keyEscape())
	if !consumed {
		t.Error("Escape should cancel pending drag")
	}

	// Now move beyond threshold — should not start drag.
	consumed = m.HandleMouseEvent(mouseMove(20, 10), nil)
	if consumed {
		t.Error("move after Escape should not start drag")
	}
	if m.IsDragging() {
		t.Error("should not be dragging after pending canceled")
	}
}

func TestEscapeWithNoDrag(t *testing.T) {
	m := NewManager()
	consumed := m.HandleKeyEvent(keyEscape())
	if consumed {
		t.Error("Escape with no drag should not be consumed")
	}
}

func TestReleaseWithNoDrag(t *testing.T) {
	m := NewManager()
	consumed := m.HandleMouseEvent(mouseRelease(10, 10), nil)
	if consumed {
		t.Error("release with no drag should not be consumed")
	}
}

func TestReleaseClearsPending(t *testing.T) {
	m := NewManager()
	src := &mockDragSource{
		data:      DragData{Kind: "text"},
		allowDrag: true,
	}

	// Press, then release without exceeding threshold.
	m.HandleMouseEvent(mousePress(10, 10), src)
	m.HandleMouseEvent(mouseRelease(12, 12), nil)

	// Subsequent move should not start drag.
	consumed := m.HandleMouseEvent(mouseMove(20, 10), nil)
	if consumed {
		t.Error("move after release should not start drag")
	}
}

func TestMouseEnterEventIgnored(t *testing.T) {
	m := NewManager()
	enterEvt := event.NewMouseEvent(
		event.MouseEnter,
		event.ButtonNone,
		0,
		geometry.Pt(10, 10),
		geometry.Pt(10, 10),
		0,
	)
	consumed := m.HandleMouseEvent(enterEvt, nil)
	if consumed {
		t.Error("MouseEnter should not be consumed")
	}
}

func TestCancelWithNoActiveDrag(t *testing.T) {
	m := NewManager()
	// Should not panic.
	m.Cancel()
	if m.IsDragging() {
		t.Error("should not be dragging")
	}
}

// --- Feedback tests ---

func TestFeedbackStruct(t *testing.T) {
	f := Feedback{Effect: DropCopy, Label: "Copy here"}
	if f.Effect != DropCopy {
		t.Errorf("Effect = %v, want DropCopy", f.Effect)
	}
	if f.Label != "Copy here" {
		t.Errorf("Label = %q, want %q", f.Label, "Copy here")
	}
}

// --- DragVisual tests ---

func TestNewDragVisualActive(t *testing.T) {
	tgt := &mockDropTarget{acceptKinds: []string{"text"}, overEffect: DropCopy}
	s := &Session{
		data:       DragData{Kind: "text"},
		source:     &mockDragSource{},
		startPos:   geometry.Pt(10, 20),
		currentPos: geometry.Pt(50, 60),
		active:     true,
		target:     tgt,
		feedback:   Feedback{Effect: DropCopy, Label: "drop"},
	}

	v := NewDragVisual(s)
	if v.CursorPos != geometry.Pt(50, 60) {
		t.Errorf("CursorPos = %v, want (50,60)", v.CursorPos)
	}
	if v.Offset != geometry.Pt(40, 40) {
		t.Errorf("Offset = %v, want (40,40)", v.Offset)
	}
	if v.Effect != DropCopy {
		t.Errorf("Effect = %v, want DropCopy", v.Effect)
	}
	if !v.OverTarget {
		t.Error("OverTarget = false, want true")
	}
	if v.Label != "drop" {
		t.Errorf("Label = %q, want %q", v.Label, "drop")
	}
}

func TestNewDragVisualNilSession(t *testing.T) {
	v := NewDragVisual(nil)
	if v.OverTarget {
		t.Error("nil session should give OverTarget=false")
	}
	if v.Effect != DropNone {
		t.Errorf("nil session effect = %v, want DropNone", v.Effect)
	}
}

func TestNewDragVisualInactiveSession(t *testing.T) {
	s := &Session{active: false}
	v := NewDragVisual(s)
	if v.OverTarget {
		t.Error("inactive session should give OverTarget=false")
	}
}

func TestNewDragVisualNoTarget(t *testing.T) {
	s := &Session{
		startPos:   geometry.Pt(10, 10),
		currentPos: geometry.Pt(50, 50),
		active:     true,
		target:     nil,
	}
	v := NewDragVisual(s)
	if v.OverTarget {
		t.Error("no target should give OverTarget=false")
	}
}

// --- Concurrent access tests ---

func TestManagerConcurrentAccess(t *testing.T) {
	m := NewManager()
	tgt := &mockDropTarget{acceptKinds: []string{"text"}, overEffect: DropCopy}
	m.RegisterTarget(tgt, geometry.NewRect(50, 50, 100, 100))

	done := make(chan struct{})
	go func() {
		defer close(done)
		for range 100 {
			_ = m.IsDragging()
			_ = m.CurrentSession()
			_ = m.TargetCount()
		}
	}()

	for range 100 {
		m.RegisterTarget(&mockDropTarget{acceptKinds: []string{"text"}}, geometry.NewRect(0, 0, 10, 10))
		_ = m.IsDragging()
	}
	<-done
}

// --- Feedback during drag ---

func TestFeedbackUpdatedDuringDrag(t *testing.T) {
	m := NewManager()
	src := &mockDragSource{
		data:      DragData{Kind: "text"},
		allowDrag: true,
	}
	tgt := &mockDropTarget{acceptKinds: []string{"text"}, overEffect: DropMove}
	m.RegisterTarget(tgt, geometry.NewRect(50, 50, 100, 100))

	// Start drag.
	m.HandleMouseEvent(mousePress(10, 10), src)
	m.HandleMouseEvent(mouseMove(20, 10), nil)

	// Before entering target — no feedback.
	session := m.CurrentSession()
	if session.Feedback().Effect != DropNone {
		t.Errorf("feedback before target = %v, want DropNone", session.Feedback().Effect)
	}

	// Enter target.
	m.HandleMouseEvent(mouseMove(100, 100), nil)
	session = m.CurrentSession()
	if session.Feedback().Effect != DropMove {
		t.Errorf("feedback on target = %v, want DropMove", session.Feedback().Effect)
	}

	// Leave target.
	m.HandleMouseEvent(mouseMove(200, 200), nil)
	session = m.CurrentSession()
	if session.Feedback().Effect != DropNone {
		t.Errorf("feedback off target = %v, want DropNone", session.Feedback().Effect)
	}
}
