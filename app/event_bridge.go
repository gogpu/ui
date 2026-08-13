package app

import (
	"github.com/gogpu/gpucontext"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
)

// eventBridgeState is shared by the platform callbacks installed by
// attachEventBridge. EventSource callbacks run on the UI thread, so the state
// does not need synchronization.
type eventBridgeState struct {
	pressedButtons  event.ButtonState
	lastMousePos    geometry.Point
	cursorInside    bool
	leaveDispatched bool
	mods            event.Modifiers
}

// pointInWindow reports whether pos is inside the window's logical content
// bounds. A size can be unavailable briefly during window creation; preserve
// event delivery until the first real size arrives in that case.
func pointInWindow(w *Window, pos geometry.Point) bool {
	size := w.WindowSize()
	if size.IsEmpty() {
		return true
	}
	return geometry.FromPointSize(geometry.Point{}, size).Contains(pos)
}

// handleBridgeMouseMove rejects ordinary mouse movement outside the window.
// A held button bypasses the gate so pointer capture and drag tracking keep
// receiving moves after the pointer crosses the window edge.
func handleBridgeMouseMove(w *Window, state *eventBridgeState, x, y float64) {
	pos := geometry.Pt(float32(x), float32(y))
	inside := pointInWindow(w, pos)
	if !inside && state.pressedButtons == 0 {
		if state.cursorInside {
			// Some platforms keep sending moves in a narrow area outside the
			// window without first sending PointerLeave. Synthesize one leave
			// on the transition so widget hover and cursor state are cleared.
			state.cursorInside = false
			state.leaveDispatched = true
			w.HandleEvent(event.NewMouseEvent(
				event.MouseLeave,
				event.ButtonNone,
				state.pressedButtons,
				pos,
				pos,
				state.mods,
			))
		}
		return
	}

	state.cursorInside = inside
	if inside {
		state.leaveDispatched = false
	}
	state.lastMousePos = pos
	w.HandleEvent(event.NewMouseEvent(
		event.MouseMove,
		event.ButtonNone,
		state.pressedButtons,
		pos,
		pos, // global position same as local for root dispatch
		state.mods,
	))
}

// attachEventBridge registers event callbacks on the EventSource that
// translate platform events into ui/event types and dispatch them to
// the Window.
//
// This function is called once during App creation when an EventSource
// is provided. The callbacks are invoked on the main thread by the host
// application's event loop.
//
// Window-bound checks belong here: Window.HandleEvent also serves the public
// App.HandleEvent API and uitest, whose caller-supplied coordinates must remain
// available for synthetic input.
func attachEventBridge(es gpucontext.EventSource, w *Window) {
	state := &eventBridgeState{}

	es.OnMouseMove(func(x, y float64) {
		handleBridgeMouseMove(w, state, x, y)
	})

	es.OnMousePress(func(button gpucontext.MouseButton, x, y float64) {
		pos := geometry.Pt(float32(x), float32(y))
		btn := translateMouseButton(button)
		state.pressedButtons |= buttonToState(btn)
		state.cursorInside = pointInWindow(w, pos)
		if state.cursorInside {
			state.leaveDispatched = false
		}
		state.lastMousePos = pos
		e := event.NewMouseEvent(
			event.MousePress,
			btn,
			state.pressedButtons,
			pos,
			pos,
			state.mods,
		)
		w.HandleEvent(e)
	})

	es.OnMouseRelease(func(button gpucontext.MouseButton, x, y float64) {
		pos := geometry.Pt(float32(x), float32(y))
		btn := translateMouseButton(button)
		state.pressedButtons &^= buttonToState(btn)
		state.cursorInside = pointInWindow(w, pos)
		if state.cursorInside {
			state.leaveDispatched = false
		}
		state.lastMousePos = pos
		e := event.NewMouseEvent(
			event.MouseRelease,
			btn,
			state.pressedButtons,
			pos,
			pos,
			state.mods,
		)
		w.HandleEvent(e)
	})

	es.OnKeyPress(func(key gpucontext.Key, platMods gpucontext.Modifiers) {
		uiKey := translateKey(key)
		uiMods := translateModifiers(platMods)
		// A key event reports the modifiers held BEFORE it, so pressing Shift
		// alone reports no Shift. Fold the key itself in, or holding a modifier
		// and clicking — with no other key in between, which is the whole
		// gesture — would leave the state empty.
		state.mods = uiMods | modifierForKey(uiKey)
		// Rune=0: character input is delivered separately via OnTextInput.
		// KeyPress only carries the key code for navigation (arrows, Tab,
		// Backspace, etc.) and modifier detection (Ctrl+C, etc.).
		e := event.NewKeyEvent(
			event.KeyPress,
			uiKey,
			0,
			uiMods,
		)
		w.HandleEvent(e)
	})

	es.OnKeyRelease(func(key gpucontext.Key, platMods gpucontext.Modifiers) {
		uiKey := translateKey(key)
		uiMods := translateModifiers(platMods)
		// Releasing a modifier clears it: the reported state still contains it.
		state.mods = uiMods &^ modifierForKey(uiKey)
		e := event.NewKeyEvent(
			event.KeyRelease,
			uiKey,
			0,
			uiMods,
		)
		w.HandleEvent(e)
	})

	es.OnTextInput(func(text string) {
		for _, r := range text {
			e := event.NewKeyEvent(
				event.KeyPress,
				event.KeyUnknown,
				r,
				event.ModNone,
			)
			w.HandleEvent(e)
		}
	})

	attachScrollBridge(es, w, state)

	es.OnResize(func(width, height int) {
		w.HandleResize(width, height)
	})

	es.OnFocus(func(focused bool) {
		// A modifier believed to be held after the window lost focus would turn
		// the next ordinary click into a modified one: the release happened
		// somewhere else and this window never saw it.
		state.mods = event.ModNone
		if !focused {
			// Focus loss does not guarantee matching mouse releases or a
			// PointerLeave. Reset bridge ownership at this boundary; otherwise
			// stale buttons would bypass the outside-window gate indefinitely.
			state.pressedButtons = 0
			state.cursorInside = false
		}
		w.HandleFocusChange(focused)
	})

	// Wire W3C Pointer Events for Enter/Leave (cursor/hover support).
	attachPointerBridge(es, w, state)
}

// attachScrollBridge prefers ScrollEventSource because it reports the pointer
// position for each wheel event. The basic EventSource callback has no
// position, so it falls back to the last pointer position and cursor state.
func attachScrollBridge(es gpucontext.EventSource, w *Window, state *eventBridgeState) {
	if scrollSource, ok := es.(gpucontext.ScrollEventSource); ok {
		scrollSource.OnScrollEvent(func(scroll gpucontext.ScrollEvent) {
			pos := geometry.Pt(float32(scroll.X), float32(scroll.Y))
			inside := pointInWindow(w, pos)
			zeroWithoutPosition := pos.IsZero() &&
				(!state.cursorInside || !state.lastMousePos.IsZero())
			if !inside || zeroWithoutPosition {
				// A few EventSource implementations historically reported wheel
				// positions in the wrong coordinate space, or as (0,0). Treat the
				// position as untrusted unless the independently tracked cursor
				// state (or an active drag) confirms that the event belongs here.
				if !state.cursorInside && state.pressedButtons == 0 {
					return
				}
				pos = state.lastMousePos
			}

			delta := geometry.Pt(float32(scroll.DeltaX), float32(scroll.DeltaY))
			w.HandleEvent(event.NewWheelEvent(
				delta,
				pos,
				pos,
				translateModifiers(scroll.Modifiers),
			))
		})
		// gogpu dispatches a detailed scroll event to both its detailed and
		// legacy callbacks. Register exactly one to avoid duplicate wheels.
		return
	}

	es.OnScroll(func(dx, dy float64) {
		if !state.cursorInside && state.pressedButtons == 0 {
			return
		}

		delta := geometry.Pt(float32(dx), float32(dy))
		w.HandleEvent(event.NewWheelEvent(
			delta,
			state.lastMousePos,
			state.lastMousePos,
			state.mods,
		))
	})
}

// attachPointerBridge wires W3C PointerEventSource for Enter/Leave events.
//
// The platform generates PointerEnter when the mouse enters the window
// and PointerLeave when it leaves. These are essential for resetting
// hover state when the mouse exits the window entirely.
//
// PointerMove/Down/Up are already handled by the legacy OnMouseMove,
// OnMousePress, and OnMouseRelease callbacks. Enter, Leave, and Cancel are
// handled here because the legacy EventSource has no equivalent callbacks.
func attachPointerBridge(es gpucontext.EventSource, w *Window, state *eventBridgeState) {
	pes, ok := es.(gpucontext.PointerEventSource)
	if !ok {
		return
	}

	pes.OnPointer(func(ev gpucontext.PointerEvent) {
		// Legacy mouse state and cursor hover must not be changed by an
		// independent touch or pen pointer.
		if ev.PointerType != gpucontext.PointerTypeMouse {
			return
		}

		switch ev.Type {
		case gpucontext.PointerEnter:
			pos := geometry.Pt(float32(ev.X), float32(ev.Y))
			state.cursorInside = true
			state.leaveDispatched = false
			state.lastMousePos = pos
			e := event.NewMouseEvent(
				event.MouseEnter,
				event.ButtonNone,
				state.pressedButtons,
				pos, pos,
				translateModifiers(ev.Modifiers),
			)
			w.HandleEvent(e)

		case gpucontext.PointerLeave:
			pos := geometry.Pt(float32(ev.X), float32(ev.Y))
			state.cursorInside = false
			if state.leaveDispatched {
				return
			}
			state.leaveDispatched = true
			e := event.NewMouseEvent(
				event.MouseLeave,
				event.ButtonNone,
				state.pressedButtons,
				pos, pos,
				translateModifiers(ev.Modifiers),
			)
			w.HandleEvent(e)

		case gpucontext.PointerCancel:
			// The platform has ended the gesture without guaranteeing legacy
			// mouse-release callbacks. Drop bridge and window capture state, then
			// fail closed until a new pointer event establishes cursor state.
			state.pressedButtons = 0
			state.cursorInside = false
			w.cancelPointerState()
		}
	})
}

// translateMouseButton converts gpucontext.MouseButton to event.Button.
func translateMouseButton(btn gpucontext.MouseButton) event.Button {
	switch btn {
	case gpucontext.MouseButtonLeft:
		return event.ButtonLeft
	case gpucontext.MouseButtonRight:
		return event.ButtonRight
	case gpucontext.MouseButtonMiddle:
		return event.ButtonMiddle
	case gpucontext.MouseButton4:
		return event.ButtonX1
	case gpucontext.MouseButton5:
		return event.ButtonX2
	default:
		return event.ButtonNone
	}
}

// buttonToState converts a single event.Button to a ButtonState bitmask.
func buttonToState(btn event.Button) event.ButtonState {
	switch btn {
	case event.ButtonLeft:
		return event.ButtonStateLeft
	case event.ButtonRight:
		return event.ButtonStateRight
	case event.ButtonMiddle:
		return event.ButtonStateMiddle
	case event.ButtonX1:
		return event.ButtonStateX1
	case event.ButtonX2:
		return event.ButtonStateX2
	default:
		return 0
	}
}

// translateKey converts gpucontext.Key to event.Key.
//
//nolint:gocyclo,cyclop,funlen,maintidx // Key mapping requires a large switch statement by design.
func translateKey(key gpucontext.Key) event.Key {
	switch key {
	// Letters
	case gpucontext.KeyA:
		return event.KeyA
	case gpucontext.KeyB:
		return event.KeyB
	case gpucontext.KeyC:
		return event.KeyC
	case gpucontext.KeyD:
		return event.KeyD
	case gpucontext.KeyE:
		return event.KeyE
	case gpucontext.KeyF:
		return event.KeyF
	case gpucontext.KeyG:
		return event.KeyG
	case gpucontext.KeyH:
		return event.KeyH
	case gpucontext.KeyI:
		return event.KeyI
	case gpucontext.KeyJ:
		return event.KeyJ
	case gpucontext.KeyK:
		return event.KeyK
	case gpucontext.KeyL:
		return event.KeyL
	case gpucontext.KeyM:
		return event.KeyM
	case gpucontext.KeyN:
		return event.KeyN
	case gpucontext.KeyO:
		return event.KeyO
	case gpucontext.KeyP:
		return event.KeyP
	case gpucontext.KeyQ:
		return event.KeyQ
	case gpucontext.KeyR:
		return event.KeyR
	case gpucontext.KeyS:
		return event.KeyS
	case gpucontext.KeyT:
		return event.KeyT
	case gpucontext.KeyU:
		return event.KeyU
	case gpucontext.KeyV:
		return event.KeyV
	case gpucontext.KeyW:
		return event.KeyW
	case gpucontext.KeyX:
		return event.KeyX
	case gpucontext.KeyY:
		return event.KeyY
	case gpucontext.KeyZ:
		return event.KeyZ

	// Numbers
	case gpucontext.Key0:
		return event.Key0
	case gpucontext.Key1:
		return event.Key1
	case gpucontext.Key2:
		return event.Key2
	case gpucontext.Key3:
		return event.Key3
	case gpucontext.Key4:
		return event.Key4
	case gpucontext.Key5:
		return event.Key5
	case gpucontext.Key6:
		return event.Key6
	case gpucontext.Key7:
		return event.Key7
	case gpucontext.Key8:
		return event.Key8
	case gpucontext.Key9:
		return event.Key9

	// Function keys
	case gpucontext.KeyF1:
		return event.KeyF1
	case gpucontext.KeyF2:
		return event.KeyF2
	case gpucontext.KeyF3:
		return event.KeyF3
	case gpucontext.KeyF4:
		return event.KeyF4
	case gpucontext.KeyF5:
		return event.KeyF5
	case gpucontext.KeyF6:
		return event.KeyF6
	case gpucontext.KeyF7:
		return event.KeyF7
	case gpucontext.KeyF8:
		return event.KeyF8
	case gpucontext.KeyF9:
		return event.KeyF9
	case gpucontext.KeyF10:
		return event.KeyF10
	case gpucontext.KeyF11:
		return event.KeyF11
	case gpucontext.KeyF12:
		return event.KeyF12

	// Navigation
	case gpucontext.KeyEscape:
		return event.KeyEscape
	case gpucontext.KeyTab:
		return event.KeyTab
	case gpucontext.KeyBackspace:
		return event.KeyBackspace
	case gpucontext.KeyEnter:
		return event.KeyEnter
	case gpucontext.KeySpace:
		return event.KeySpace
	case gpucontext.KeyInsert:
		return event.KeyInsert
	case gpucontext.KeyDelete:
		return event.KeyDelete
	case gpucontext.KeyHome:
		return event.KeyHome
	case gpucontext.KeyEnd:
		return event.KeyEnd
	case gpucontext.KeyPageUp:
		return event.KeyPageUp
	case gpucontext.KeyPageDown:
		return event.KeyPageDown
	case gpucontext.KeyLeft:
		return event.KeyLeft
	case gpucontext.KeyRight:
		return event.KeyRight
	case gpucontext.KeyUp:
		return event.KeyUp
	case gpucontext.KeyDown:
		return event.KeyDown

	// Modifiers
	case gpucontext.KeyLeftShift:
		return event.KeyLeftShift
	case gpucontext.KeyRightShift:
		return event.KeyRightShift
	case gpucontext.KeyLeftControl:
		return event.KeyLeftCtrl
	case gpucontext.KeyRightControl:
		return event.KeyRightCtrl
	case gpucontext.KeyLeftAlt:
		return event.KeyLeftAlt
	case gpucontext.KeyRightAlt:
		return event.KeyRightAlt
	case gpucontext.KeyLeftSuper:
		return event.KeyLeftSuper
	case gpucontext.KeyRightSuper:
		return event.KeyRightSuper

	// Punctuation
	case gpucontext.KeyMinus:
		return event.KeyMinus
	case gpucontext.KeyEqual:
		return event.KeyEqual
	case gpucontext.KeyLeftBracket:
		return event.KeyLeftBracket
	case gpucontext.KeyRightBracket:
		return event.KeyRightBracket
	case gpucontext.KeyBackslash:
		return event.KeyBackslash
	case gpucontext.KeySemicolon:
		return event.KeySemicolon
	case gpucontext.KeyApostrophe:
		return event.KeyApostrophe
	case gpucontext.KeyGrave:
		return event.KeyGrave
	case gpucontext.KeyComma:
		return event.KeyComma
	case gpucontext.KeyPeriod:
		return event.KeyPeriod
	case gpucontext.KeySlash:
		return event.KeySlash

	// Numpad
	case gpucontext.KeyNumpad0:
		return event.KeyNumpad0
	case gpucontext.KeyNumpad1:
		return event.KeyNumpad1
	case gpucontext.KeyNumpad2:
		return event.KeyNumpad2
	case gpucontext.KeyNumpad3:
		return event.KeyNumpad3
	case gpucontext.KeyNumpad4:
		return event.KeyNumpad4
	case gpucontext.KeyNumpad5:
		return event.KeyNumpad5
	case gpucontext.KeyNumpad6:
		return event.KeyNumpad6
	case gpucontext.KeyNumpad7:
		return event.KeyNumpad7
	case gpucontext.KeyNumpad8:
		return event.KeyNumpad8
	case gpucontext.KeyNumpad9:
		return event.KeyNumpad9
	case gpucontext.KeyNumpadDecimal:
		return event.KeyNumpadDecimal
	case gpucontext.KeyNumpadDivide:
		return event.KeyNumpadDivide
	case gpucontext.KeyNumpadMultiply:
		return event.KeyNumpadMultiply
	case gpucontext.KeyNumpadSubtract:
		return event.KeyNumpadSubtract
	case gpucontext.KeyNumpadAdd:
		return event.KeyNumpadAdd
	case gpucontext.KeyNumpadEnter:
		return event.KeyNumpadEnter

	// Other
	case gpucontext.KeyCapsLock:
		return event.KeyCapsLock
	case gpucontext.KeyScrollLock:
		return event.KeyScrollLock
	case gpucontext.KeyNumLock:
		return event.KeyNumLock
	case gpucontext.KeyPrintScreen:
		return event.KeyPrintScreen
	case gpucontext.KeyPause:
		return event.KeyPause

	default:
		return event.KeyUnknown
	}
}

// translateModifiers converts gpucontext.Modifiers to event.Modifiers.
// modifierForKey is the modifier bit a key sets while it is held, or ModNone
// for anything that is not a modifier.
func modifierForKey(k event.Key) event.Modifiers {
	switch k {
	case event.KeyLeftShift, event.KeyRightShift:
		return event.ModShift
	case event.KeyLeftCtrl, event.KeyRightCtrl:
		return event.ModCtrl
	case event.KeyLeftAlt, event.KeyRightAlt:
		return event.ModAlt
	case event.KeyLeftSuper, event.KeyRightSuper:
		return event.ModSuper
	}
	return event.ModNone
}

func translateModifiers(mods gpucontext.Modifiers) event.Modifiers {
	var result event.Modifiers
	if mods.HasShift() {
		result |= event.ModShift
	}
	if mods.HasControl() {
		result |= event.ModCtrl
	}
	if mods.HasAlt() {
		result |= event.ModAlt
	}
	if mods.HasSuper() {
		result |= event.ModSuper
	}
	return result
}
