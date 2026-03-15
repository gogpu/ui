package app

import (
	"testing"

	"github.com/gogpu/gpucontext"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

func TestEventBridge_MouseMove(t *testing.T) {
	es := &mockEventSource{}
	a := New(WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)

	// Simulate mouse move.
	es.onMouseMove(100.0, 200.0)

	if !root.eventCalled {
		t.Fatal("event not dispatched")
	}
	me, ok := root.lastEvent.(*event.MouseEvent)
	if !ok {
		t.Fatal("expected MouseEvent")
	}
	if me.MouseType != event.MouseMove {
		t.Errorf("mouse type = %v, want Move", me.MouseType)
	}
	if me.Position.X != 100.0 || me.Position.Y != 200.0 {
		t.Errorf("position = %v, want (100, 200)", me.Position)
	}
}

func TestEventBridge_MousePress(t *testing.T) {
	es := &mockEventSource{}
	a := New(WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)

	es.onMousePress(gpucontext.MouseButtonLeft, 50.0, 75.0)

	if !root.eventCalled {
		t.Fatal("event not dispatched")
	}
	me, ok := root.lastEvent.(*event.MouseEvent)
	if !ok {
		t.Fatal("expected MouseEvent")
	}
	if me.MouseType != event.MousePress {
		t.Errorf("mouse type = %v, want Press", me.MouseType)
	}
	if me.Button != event.ButtonLeft {
		t.Errorf("button = %v, want Left", me.Button)
	}
	if !me.Buttons.IsLeftPressed() {
		t.Error("left button should be in pressed state")
	}
}

func TestEventBridge_MouseRelease(t *testing.T) {
	es := &mockEventSource{}
	a := New(WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)

	es.onMouseRelease(gpucontext.MouseButtonRight, 30.0, 40.0)

	if !root.eventCalled {
		t.Fatal("event not dispatched")
	}
	me, ok := root.lastEvent.(*event.MouseEvent)
	if !ok {
		t.Fatal("expected MouseEvent")
	}
	if me.MouseType != event.MouseRelease {
		t.Errorf("mouse type = %v, want Release", me.MouseType)
	}
	if me.Button != event.ButtonRight {
		t.Errorf("button = %v, want Right", me.Button)
	}
}

func TestEventBridge_KeyPress(t *testing.T) {
	es := &mockEventSource{}
	a := New(WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)

	es.onKeyPress(gpucontext.KeyA, gpucontext.ModShift|gpucontext.ModControl)

	if !root.eventCalled {
		t.Fatal("event not dispatched")
	}
	ke, ok := root.lastEvent.(*event.KeyEvent)
	if !ok {
		t.Fatal("expected KeyEvent")
	}
	if ke.KeyType != event.KeyPress {
		t.Errorf("key type = %v, want Press", ke.KeyType)
	}
	if ke.Key != event.KeyA {
		t.Errorf("key = %v, want A", ke.Key)
	}
	if !ke.Modifiers().IsShift() {
		t.Error("expected Shift modifier")
	}
	if !ke.Modifiers().IsCtrl() {
		t.Error("expected Ctrl modifier")
	}
}

func TestEventBridge_KeyRelease(t *testing.T) {
	es := &mockEventSource{}
	a := New(WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)

	es.onKeyRelease(gpucontext.KeyEscape, 0)

	if !root.eventCalled {
		t.Fatal("event not dispatched")
	}
	ke, ok := root.lastEvent.(*event.KeyEvent)
	if !ok {
		t.Fatal("expected KeyEvent")
	}
	if ke.KeyType != event.KeyRelease {
		t.Errorf("key type = %v, want Release", ke.KeyType)
	}
	if ke.Key != event.KeyEscape {
		t.Errorf("key = %v, want Escape", ke.Key)
	}
}

func TestEventBridge_Scroll(t *testing.T) {
	es := &mockEventSource{}
	a := New(WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)

	es.onScroll(0.0, -3.0)

	if !root.eventCalled {
		t.Fatal("event not dispatched")
	}
	we, ok := root.lastEvent.(*event.WheelEvent)
	if !ok {
		t.Fatal("expected WheelEvent")
	}
	if we.Delta.X != 0.0 {
		t.Errorf("delta X = %v, want 0", we.Delta.X)
	}
	if we.Delta.Y != -3.0 {
		t.Errorf("delta Y = %v, want -3", we.Delta.Y)
	}
}

func TestEventBridge_Resize(t *testing.T) {
	es := &mockEventSource{}
	a := New(WithEventSource(es))

	es.onResize(1920, 1080)

	size := a.Window().WindowSize()
	if size.Width != 1920 || size.Height != 1080 {
		t.Errorf("size = %v, want (1920, 1080)", size)
	}
	if !a.Window().NeedsLayout() {
		t.Error("resize should mark layout as needed")
	}
}

func TestEventBridge_Focus(t *testing.T) {
	es := &mockEventSource{}
	a := New(WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)

	es.onFocus(true)

	if !root.eventCalled {
		t.Fatal("focus event not dispatched")
	}
	fe, ok := root.lastEvent.(*event.FocusEvent)
	if !ok {
		t.Fatal("expected FocusEvent")
	}
	if !fe.IsGained() {
		t.Error("expected focus gained")
	}
}

func TestEventBridge_FocusLost(t *testing.T) {
	es := &mockEventSource{}
	a := New(WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)

	es.onFocus(false)

	if !root.eventCalled {
		t.Fatal("focus event not dispatched")
	}
	fe, ok := root.lastEvent.(*event.FocusEvent)
	if !ok {
		t.Fatal("expected FocusEvent")
	}
	if !fe.IsLost() {
		t.Error("expected focus lost")
	}
}

// --- Translation function tests ---

func TestTranslateMouseButton(t *testing.T) {
	tests := []struct {
		name string
		in   gpucontext.MouseButton
		want event.Button
	}{
		{"Left", gpucontext.MouseButtonLeft, event.ButtonLeft},
		{"Right", gpucontext.MouseButtonRight, event.ButtonRight},
		{"Middle", gpucontext.MouseButtonMiddle, event.ButtonMiddle},
		{"Button4", gpucontext.MouseButton4, event.ButtonX1},
		{"Button5", gpucontext.MouseButton5, event.ButtonX2},
		{"Unknown", gpucontext.MouseButton(99), event.ButtonNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := translateMouseButton(tt.in)
			if got != tt.want {
				t.Errorf("translateMouseButton(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestButtonToState(t *testing.T) {
	tests := []struct {
		name string
		in   event.Button
		want event.ButtonState
	}{
		{"Left", event.ButtonLeft, event.ButtonStateLeft},
		{"Right", event.ButtonRight, event.ButtonStateRight},
		{"Middle", event.ButtonMiddle, event.ButtonStateMiddle},
		{"X1", event.ButtonX1, event.ButtonStateX1},
		{"X2", event.ButtonX2, event.ButtonStateX2},
		{"None", event.ButtonNone, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buttonToState(tt.in)
			if got != tt.want {
				t.Errorf("buttonToState(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestTranslateKey(t *testing.T) {
	tests := []struct {
		name string
		in   gpucontext.Key
		want event.Key
	}{
		// Letters
		{"A", gpucontext.KeyA, event.KeyA},
		{"B", gpucontext.KeyB, event.KeyB},
		{"C", gpucontext.KeyC, event.KeyC},
		{"D", gpucontext.KeyD, event.KeyD},
		{"E", gpucontext.KeyE, event.KeyE},
		{"F", gpucontext.KeyF, event.KeyF},
		{"G", gpucontext.KeyG, event.KeyG},
		{"H", gpucontext.KeyH, event.KeyH},
		{"I", gpucontext.KeyI, event.KeyI},
		{"J", gpucontext.KeyJ, event.KeyJ},
		{"K", gpucontext.KeyK, event.KeyK},
		{"L", gpucontext.KeyL, event.KeyL},
		{"M", gpucontext.KeyM, event.KeyM},
		{"N", gpucontext.KeyN, event.KeyN},
		{"O", gpucontext.KeyO, event.KeyO},
		{"P", gpucontext.KeyP, event.KeyP},
		{"Q", gpucontext.KeyQ, event.KeyQ},
		{"R", gpucontext.KeyR, event.KeyR},
		{"S", gpucontext.KeyS, event.KeyS},
		{"T", gpucontext.KeyT, event.KeyT},
		{"U", gpucontext.KeyU, event.KeyU},
		{"V", gpucontext.KeyV, event.KeyV},
		{"W", gpucontext.KeyW, event.KeyW},
		{"X", gpucontext.KeyX, event.KeyX},
		{"Y", gpucontext.KeyY, event.KeyY},
		{"Z", gpucontext.KeyZ, event.KeyZ},

		// Numbers
		{"0", gpucontext.Key0, event.Key0},
		{"1", gpucontext.Key1, event.Key1},
		{"2", gpucontext.Key2, event.Key2},
		{"3", gpucontext.Key3, event.Key3},
		{"4", gpucontext.Key4, event.Key4},
		{"5", gpucontext.Key5, event.Key5},
		{"6", gpucontext.Key6, event.Key6},
		{"7", gpucontext.Key7, event.Key7},
		{"8", gpucontext.Key8, event.Key8},
		{"9", gpucontext.Key9, event.Key9},

		// Function keys
		{"F1", gpucontext.KeyF1, event.KeyF1},
		{"F2", gpucontext.KeyF2, event.KeyF2},
		{"F3", gpucontext.KeyF3, event.KeyF3},
		{"F4", gpucontext.KeyF4, event.KeyF4},
		{"F5", gpucontext.KeyF5, event.KeyF5},
		{"F6", gpucontext.KeyF6, event.KeyF6},
		{"F7", gpucontext.KeyF7, event.KeyF7},
		{"F8", gpucontext.KeyF8, event.KeyF8},
		{"F9", gpucontext.KeyF9, event.KeyF9},
		{"F10", gpucontext.KeyF10, event.KeyF10},
		{"F11", gpucontext.KeyF11, event.KeyF11},
		{"F12", gpucontext.KeyF12, event.KeyF12},

		// Navigation
		{"Escape", gpucontext.KeyEscape, event.KeyEscape},
		{"Tab", gpucontext.KeyTab, event.KeyTab},
		{"Backspace", gpucontext.KeyBackspace, event.KeyBackspace},
		{"Enter", gpucontext.KeyEnter, event.KeyEnter},
		{"Space", gpucontext.KeySpace, event.KeySpace},
		{"Insert", gpucontext.KeyInsert, event.KeyInsert},
		{"Delete", gpucontext.KeyDelete, event.KeyDelete},
		{"Home", gpucontext.KeyHome, event.KeyHome},
		{"End", gpucontext.KeyEnd, event.KeyEnd},
		{"PageUp", gpucontext.KeyPageUp, event.KeyPageUp},
		{"PageDown", gpucontext.KeyPageDown, event.KeyPageDown},
		{"Left", gpucontext.KeyLeft, event.KeyLeft},
		{"Right", gpucontext.KeyRight, event.KeyRight},
		{"Up", gpucontext.KeyUp, event.KeyUp},
		{"Down", gpucontext.KeyDown, event.KeyDown},

		// Modifiers
		{"LeftShift", gpucontext.KeyLeftShift, event.KeyLeftShift},
		{"RightShift", gpucontext.KeyRightShift, event.KeyRightShift},
		{"LeftControl", gpucontext.KeyLeftControl, event.KeyLeftCtrl},
		{"RightControl", gpucontext.KeyRightControl, event.KeyRightCtrl},
		{"LeftAlt", gpucontext.KeyLeftAlt, event.KeyLeftAlt},
		{"RightAlt", gpucontext.KeyRightAlt, event.KeyRightAlt},
		{"LeftSuper", gpucontext.KeyLeftSuper, event.KeyLeftSuper},
		{"RightSuper", gpucontext.KeyRightSuper, event.KeyRightSuper},

		// Punctuation
		{"Minus", gpucontext.KeyMinus, event.KeyMinus},
		{"Equal", gpucontext.KeyEqual, event.KeyEqual},
		{"LeftBracket", gpucontext.KeyLeftBracket, event.KeyLeftBracket},
		{"RightBracket", gpucontext.KeyRightBracket, event.KeyRightBracket},
		{"Backslash", gpucontext.KeyBackslash, event.KeyBackslash},
		{"Semicolon", gpucontext.KeySemicolon, event.KeySemicolon},
		{"Apostrophe", gpucontext.KeyApostrophe, event.KeyApostrophe},
		{"Grave", gpucontext.KeyGrave, event.KeyGrave},
		{"Comma", gpucontext.KeyComma, event.KeyComma},
		{"Period", gpucontext.KeyPeriod, event.KeyPeriod},
		{"Slash", gpucontext.KeySlash, event.KeySlash},

		// Numpad
		{"Numpad0", gpucontext.KeyNumpad0, event.KeyNumpad0},
		{"Numpad1", gpucontext.KeyNumpad1, event.KeyNumpad1},
		{"Numpad2", gpucontext.KeyNumpad2, event.KeyNumpad2},
		{"Numpad3", gpucontext.KeyNumpad3, event.KeyNumpad3},
		{"Numpad4", gpucontext.KeyNumpad4, event.KeyNumpad4},
		{"Numpad5", gpucontext.KeyNumpad5, event.KeyNumpad5},
		{"Numpad6", gpucontext.KeyNumpad6, event.KeyNumpad6},
		{"Numpad7", gpucontext.KeyNumpad7, event.KeyNumpad7},
		{"Numpad8", gpucontext.KeyNumpad8, event.KeyNumpad8},
		{"Numpad9", gpucontext.KeyNumpad9, event.KeyNumpad9},
		{"NumpadDecimal", gpucontext.KeyNumpadDecimal, event.KeyNumpadDecimal},
		{"NumpadDivide", gpucontext.KeyNumpadDivide, event.KeyNumpadDivide},
		{"NumpadMultiply", gpucontext.KeyNumpadMultiply, event.KeyNumpadMultiply},
		{"NumpadSubtract", gpucontext.KeyNumpadSubtract, event.KeyNumpadSubtract},
		{"NumpadAdd", gpucontext.KeyNumpadAdd, event.KeyNumpadAdd},
		{"NumpadEnter", gpucontext.KeyNumpadEnter, event.KeyNumpadEnter},

		// Other
		{"CapsLock", gpucontext.KeyCapsLock, event.KeyCapsLock},
		{"ScrollLock", gpucontext.KeyScrollLock, event.KeyScrollLock},
		{"NumLock", gpucontext.KeyNumLock, event.KeyNumLock},
		{"PrintScreen", gpucontext.KeyPrintScreen, event.KeyPrintScreen},
		{"Pause", gpucontext.KeyPause, event.KeyPause},

		// Unknown
		{"Unknown", gpucontext.Key(9999), event.KeyUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := translateKey(tt.in)
			if got != tt.want {
				t.Errorf("translateKey(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestTranslateModifiers(t *testing.T) {
	tests := []struct {
		name string
		in   gpucontext.Modifiers
		want event.Modifiers
	}{
		{"None", 0, event.ModNone},
		{"Shift", gpucontext.ModShift, event.ModShift},
		{"Control", gpucontext.ModControl, event.ModCtrl},
		{"Alt", gpucontext.ModAlt, event.ModAlt},
		{"Super", gpucontext.ModSuper, event.ModSuper},
		{"ShiftCtrl", gpucontext.ModShift | gpucontext.ModControl, event.ModShift | event.ModCtrl},
		{"All", gpucontext.ModShift | gpucontext.ModControl | gpucontext.ModAlt | gpucontext.ModSuper,
			event.ModShift | event.ModCtrl | event.ModAlt | event.ModSuper},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := translateModifiers(tt.in)
			if got != tt.want {
				t.Errorf("translateModifiers(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestWidgetCursorToPlatform(t *testing.T) {
	tests := []struct {
		name string
		in   widget.CursorType
		want gpucontext.CursorShape
	}{
		{"Default", widget.CursorDefault, gpucontext.CursorDefault},
		{"Pointer", widget.CursorPointer, gpucontext.CursorPointer},
		{"Text", widget.CursorText, gpucontext.CursorText},
		{"Crosshair", widget.CursorCrosshair, gpucontext.CursorCrosshair},
		{"Move", widget.CursorMove, gpucontext.CursorMove},
		{"ResizeNS", widget.CursorResizeNS, gpucontext.CursorResizeNS},
		{"ResizeEW", widget.CursorResizeEW, gpucontext.CursorResizeEW},
		{"ResizeNESW", widget.CursorResizeNESW, gpucontext.CursorResizeNESW},
		{"ResizeNWSE", widget.CursorResizeNWSE, gpucontext.CursorResizeNWSE},
		{"NotAllowed", widget.CursorNotAllowed, gpucontext.CursorNotAllowed},
		{"Wait", widget.CursorWait, gpucontext.CursorWait},
		{"None", widget.CursorNone, gpucontext.CursorNone},
		{"UnknownFallback", widget.CursorType(99), gpucontext.CursorDefault},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := widgetCursorToPlatform(tt.in)
			if got != tt.want {
				t.Errorf("widgetCursorToPlatform(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestEventBridge_MouseButton_AllVariants(t *testing.T) {
	buttons := []struct {
		name string
		btn  gpucontext.MouseButton
		want event.Button
	}{
		{"Left", gpucontext.MouseButtonLeft, event.ButtonLeft},
		{"Right", gpucontext.MouseButtonRight, event.ButtonRight},
		{"Middle", gpucontext.MouseButtonMiddle, event.ButtonMiddle},
	}

	for _, tt := range buttons {
		t.Run(tt.name, func(t *testing.T) {
			es := &mockEventSource{}
			a := New(WithEventSource(es))
			root := newMockWidget()
			a.SetRoot(root)

			es.onMousePress(tt.btn, 10.0, 20.0)

			me, ok := root.lastEvent.(*event.MouseEvent)
			if !ok {
				t.Fatal("expected MouseEvent")
			}
			if me.Button != tt.want {
				t.Errorf("button = %v, want %v", me.Button, tt.want)
			}
		})
	}
}

func TestEventBridge_Resize_WithLayout(t *testing.T) {
	es := &mockEventSource{}
	a := New(WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)

	// Initial frame.
	a.Frame()
	root.layoutCalled = false

	// Resize via event bridge.
	es.onResize(500, 400)

	// Next frame should relayout.
	a.Frame()
	if !root.layoutCalled {
		t.Error("resize via event bridge should trigger relayout")
	}
}

func TestEventBridge_NoEventSource(t *testing.T) {
	// App without event source should work fine.
	a := New()
	root := newMockWidget()
	a.SetRoot(root)
	a.Frame()

	if !root.layoutCalled {
		t.Error("layout should work without event source")
	}
}

func TestSetFrameCallback_NilWindow(t *testing.T) {
	// Test the guard for nil window.
	a := &App{}
	// Should not panic.
	a.SetFrameCallback(func(_ FrameStats) {})
}

// Verify compile-time interface satisfaction for mocks.
var (
	_ gpucontext.WindowProvider     = (*mockWindowProvider)(nil)
	_ gpucontext.PlatformProvider   = (*mockPlatformProvider)(nil)
	_ gpucontext.EventSource        = (*mockEventSource)(nil)
	_ gpucontext.PointerEventSource = (*mockEventSource)(nil)
	_ widget.Canvas                 = (*mockCanvas)(nil)
	_ widget.Widget                 = (*mockWidget)(nil)
	_ widget.Widget                 = (*cursorSettingWidget)(nil)
	_ widget.Widget                 = (*cursorSettingOnLayoutWidget)(nil)
)

// --- PointerEventSource tests ---

func TestEventBridge_PointerEnter(t *testing.T) {
	es := &mockEventSource{}
	a := New(WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)

	if es.onPointer == nil {
		t.Fatal("OnPointer was not registered")
	}

	es.onPointer(gpucontext.PointerEvent{
		Type:        gpucontext.PointerEnter,
		X:           100.0,
		Y:           200.0,
		PointerType: gpucontext.PointerTypeMouse,
		IsPrimary:   true,
		Modifiers:   gpucontext.ModShift,
	})

	if !root.eventCalled {
		t.Fatal("PointerEnter event not dispatched")
	}
	me, ok := root.lastEvent.(*event.MouseEvent)
	if !ok {
		t.Fatal("expected MouseEvent")
	}
	if me.MouseType != event.MouseEnter {
		t.Errorf("mouse type = %v, want Enter", me.MouseType)
	}
	if me.Position.X != 100.0 || me.Position.Y != 200.0 {
		t.Errorf("position = %v, want (100, 200)", me.Position)
	}
	if !me.Modifiers().IsShift() {
		t.Error("expected Shift modifier")
	}
}

func TestEventBridge_PointerLeave(t *testing.T) {
	es := &mockEventSource{}
	a := New(WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)

	es.onPointer(gpucontext.PointerEvent{
		Type:        gpucontext.PointerLeave,
		X:           0.0,
		Y:           0.0,
		PointerType: gpucontext.PointerTypeMouse,
		IsPrimary:   true,
	})

	if !root.eventCalled {
		t.Fatal("PointerLeave event not dispatched")
	}
	me, ok := root.lastEvent.(*event.MouseEvent)
	if !ok {
		t.Fatal("expected MouseEvent")
	}
	if me.MouseType != event.MouseLeave {
		t.Errorf("mouse type = %v, want Leave", me.MouseType)
	}
}

func TestEventBridge_PointerMove_Ignored(t *testing.T) {
	es := &mockEventSource{}
	a := New(WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)

	// PointerMove should be ignored (already handled by OnMouseMove).
	es.onPointer(gpucontext.PointerEvent{
		Type: gpucontext.PointerMove,
		X:    50.0,
		Y:    50.0,
	})

	if root.eventCalled {
		t.Error("PointerMove should not dispatch via OnPointer (handled by OnMouseMove)")
	}
}

func TestEventBridge_PointerEnter_UpdatesLastMousePos(t *testing.T) {
	es := &mockEventSource{}
	a := New(WithEventSource(es))
	root := newMockWidget()
	a.SetRoot(root)

	// PointerEnter should update lastMousePos so subsequent scroll events
	// carry the correct position.
	es.onPointer(gpucontext.PointerEvent{
		Type: gpucontext.PointerEnter,
		X:    300.0,
		Y:    400.0,
	})

	// Reset event tracking.
	root.eventCalled = false
	root.lastEvent = nil

	// Scroll should use the position from PointerEnter.
	es.onScroll(0.0, -1.0)

	we, ok := root.lastEvent.(*event.WheelEvent)
	if !ok {
		t.Fatal("expected WheelEvent")
	}
	if we.Position.X != 300.0 || we.Position.Y != 400.0 {
		t.Errorf("wheel position = %v, want (300, 400)", we.Position)
	}
}

func TestEventBridge_PointerEventSource_Registration(t *testing.T) {
	es := &mockEventSource{}
	_ = New(WithEventSource(es))

	if es.onPointer == nil {
		t.Error("OnPointer callback was not registered")
	}
}

// Verify no unused imports by using geometry in a test.
var _ = geometry.Pt(0, 0)
