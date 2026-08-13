package app

import (
	"github.com/gogpu/gpucontext"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
)

// eventBridgeState tracks pointer/scroll session state shared among all
// event bridge callbacks. Keeping it in a struct avoids a growing list of
// closure-captured locals and makes the state available to helper functions.
type eventBridgeState struct {
	// pressedButtons tracks which mouse buttons are currently pressed so
	// that derived MouseMove events carry accurate ButtonState.
	pressedButtons event.ButtonState

	// lastMousePos is the last known mouse position so WheelEvents carry
	// the correct position (the platform's OnScroll callback doesn't
	// provide mouse coordinates).
	lastMousePos geometry.Point

	// mods tracks the keyboard modifiers so mouse and wheel events can
	// carry them. The platform's mouse callbacks report only a button and
	// a position, so without this every MouseEvent is built with ModNone
	// and modified clicks would not be expressible.
	mods event.Modifiers

	// mouseInsideWindow is true when the mouse pointer is known to be
	// inside the window's client area. Set by PointerEnter and legacy
	// OnMouseMove (with bounds check). Cleared by PointerLeave, focus
	// loss, and PointerCancel (mouse). Used to suppress scroll events
	// that arrive after the cursor has left (e.g. macOS momentum scroll).
	mouseInsideWindow bool
}

// pointInWindow returns true when (x, y) is inside the window's logical
// bounds. Returns true when no WindowProvider is set (headless).
func (s *eventBridgeState) pointInWindow(w *Window, x, y float32) bool {
	if w.wp == nil {
		return true
	}
	pw, ph := w.wp.Size()
	return x >= 0 && y >= 0 && x < float32(pw) && y < float32(ph)
}

// scrollAllowed returns true when a scroll event should be dispatched.
// Scrolls are allowed when the mouse is inside the window or when a
// mouse button is held (drag in progress).
func (s *eventBridgeState) scrollAllowed() bool {
	return s.mouseInsideWindow || s.pressedButtons != 0
}

// attachEventBridge registers event callbacks on the EventSource that
// translate platform events into ui/event types and dispatch them to
// the Window.
//
// Unified pointer pipeline (ADR-049 Phase 3):
// PointerEvent is the SINGLE source of all pointer input. If the platform
// provides PointerEventSource, those events drive everything. If the platform
// has only legacy mouse callbacks, they are wrapped into PointerEvents first.
//
// HandlePointerEvent is the sole entry point. It feeds the gesture arena AND
// derives MouseEvent for existing widget dispatch via HandleEvent.
//
// This function is called once during App creation when an EventSource
// is provided. The callbacks are invoked on the main thread by the host
// application's event loop.
func attachEventBridge(es gpucontext.EventSource, w *Window) {
	st := &eventBridgeState{}

	_, hasPointerSource := es.(gpucontext.PointerEventSource)

	// --- Unified pointer pipeline ---
	//
	// If the platform has PointerEventSource: OnPointer handles ALL pointer
	// events (Down/Up/Move/Cancel/Enter/Leave). Legacy mouse callbacks are
	// still wired for button state tracking but do NOT dispatch events.
	//
	// If the platform has only legacy mouse callbacks: they synthesize
	// PointerEvents and feed them through HandlePointerEvent.

	if hasPointerSource {
		// Platform provides rich pointer events. Wire legacy callbacks
		// ONLY for button state tracking (pressedButtons/lastMousePos),
		// not for event dispatch. All dispatch goes through OnPointer.
		es.OnMouseMove(func(x, y float64) {
			st.lastMousePos = geometry.Pt(float32(x), float32(y))
			st.mouseInsideWindow = st.pointInWindow(w, float32(x), float32(y))
		})
		es.OnMousePress(func(button gpucontext.MouseButton, _ float64, _ float64) {
			btn := translateMouseButton(button)
			st.pressedButtons |= buttonToState(btn)
		})
		es.OnMouseRelease(func(button gpucontext.MouseButton, _ float64, _ float64) {
			btn := translateMouseButton(button)
			st.pressedButtons &^= buttonToState(btn)
		})
	} else {
		// Legacy-only platform. Synthesize PointerEvents from mouse callbacks
		// and feed through the unified HandlePointerEvent path.
		es.OnMouseMove(func(x, y float64) {
			pos := geometry.Pt(float32(x), float32(y))
			st.lastMousePos = pos
			st.mouseInsideWindow = st.pointInWindow(w, float32(x), float32(y))
			// Only synthesize gesture moves when buttons are pressed.
			// Unpressed moves are handled as derived MouseMove by HandlePointerEvent
			// only when there is a gesture in progress; otherwise dispatch a
			// plain MouseMove for hover tracking.
			if st.pressedButtons != 0 {
				gev := synthesizePointerEvent(event.MouseMove, event.ButtonNone, st.pressedButtons, pos, st.mods)
				w.HandlePointerEvent(&gev)
			} else {
				// No gesture in progress — dispatch MouseMove directly for hover.
				e := event.NewMouseEvent(event.MouseMove, event.ButtonNone,
					st.pressedButtons, pos, pos, st.mods)
				w.HandleEvent(e)
			}
		})

		es.OnMousePress(func(button gpucontext.MouseButton, x, y float64) {
			pos := geometry.Pt(float32(x), float32(y))
			btn := translateMouseButton(button)
			st.pressedButtons |= buttonToState(btn)
			gev := synthesizePointerEvent(event.MousePress, btn, st.pressedButtons, pos, st.mods)
			w.HandlePointerEvent(&gev)
		})

		es.OnMouseRelease(func(button gpucontext.MouseButton, x, y float64) {
			pos := geometry.Pt(float32(x), float32(y))
			btn := translateMouseButton(button)
			st.pressedButtons &^= buttonToState(btn)
			gev := synthesizePointerEvent(event.MouseRelease, btn, st.pressedButtons, pos, st.mods)
			w.HandlePointerEvent(&gev)
		})
	}

	attachKeyboardBridge(es, w, &st.mods)
	attachScrollBridge(es, w, st)

	es.OnResize(func(width, height int) {
		w.HandleResize(width, height)
	})

	es.OnFocus(func(focused bool) {
		// A modifier believed to be held after the window lost focus would turn
		// the next ordinary click into a modified one: the release happened
		// somewhere else and this window never saw it.
		st.mods = event.ModNone
		if !focused {
			// The cursor may leave the window while focus is switching to
			// another application. Without a PointerLeave, the bridge
			// would continue to believe the cursor is inside, allowing
			// stale scroll events through.
			st.mouseInsideWindow = false
			st.pressedButtons = 0
		}
		w.HandleFocusChange(focused)
	})

	// Wire W3C Pointer Events for the unified pipeline. When the platform
	// has PointerEventSource, this handles ALL pointer types (Down/Up/Move/
	// Cancel/Enter/Leave). The unified HandlePointerEvent derives MouseEvent
	// for existing widget dispatch.
	attachPointerBridge(es, w, st)
}

// attachScrollBridge wires scroll event callbacks.
//
// If the platform provides ScrollEventSource (position-carrying scroll events),
// OnScrollEvent is used and OnScroll is NOT wired. This avoids double dispatch.
// If only the basic OnScroll is available, it uses lastMousePos from legacy
// mouse tracking.
//
// Scrolls are filtered by mouseInsideWindow to prevent dispatch when the cursor
// has left the window (e.g. macOS momentum/inertial scroll).
func attachScrollBridge(es gpucontext.EventSource, w *Window, st *eventBridgeState) {
	if ses, ok := es.(gpucontext.ScrollEventSource); ok {
		ses.OnScrollEvent(func(sev gpucontext.ScrollEvent) {
			// Decide whether to trust the position embedded in the event
			// or fall back to the independently tracked mouse position.
			// Some backends report physical (out-of-bounds) coordinates or
			// zero for events synthesized from a touchpad.
			//
			// Decision matrix:
			// 1. Reported non-zero position inside window → use it.
			// 2. Reported (0,0): ambiguous — could be the real window
			//    corner or an uninitialized zero from the backend. Fall
			//    back to lastMousePos when the cursor is inside; use (0,0)
			//    only when lastMousePos is also (0,0) (confirming it).
			// 3. Reported outside, cursor tracked inside (or dragging) →
			//    fall back to lastMousePos.
			// 4. Neither source can confirm inside → suppress.
			reportedInBounds := st.pointInWindow(w, float32(sev.X), float32(sev.Y))
			isZeroPos := sev.X == 0 && sev.Y == 0

			var pos geometry.Point
			switch {
			case reportedInBounds && !isZeroPos:
				// Non-zero position inside bounds. Trusted.
				pos = geometry.Pt(float32(sev.X), float32(sev.Y))

			case isZeroPos && st.mouseInsideWindow:
				// Zero position while cursor is tracked inside. Use lastMousePos
				// which is the independently confirmed position. When lastMousePos
				// is also (0,0), we correctly use (0,0).
				pos = st.lastMousePos

			case st.scrollAllowed():
				// Reported position is outside (or untrusted) but cursor is
				// tracked inside or dragging. Use the last known good position.
				pos = st.lastMousePos

			default:
				// Cursor is outside and no drag is active. Suppress.
				return
			}

			delta := geometry.Pt(float32(sev.DeltaX), float32(sev.DeltaY))
			e := event.NewWheelEvent(delta, pos, pos, translateModifiers(sev.Modifiers))
			w.HandleEvent(e)
		})
		// Do NOT wire OnScroll — ScrollEventSource replaces it entirely.
		return
	}

	es.OnScroll(func(dx, dy float64) {
		if !st.scrollAllowed() {
			return
		}
		delta := geometry.Pt(float32(dx), float32(dy))
		e := event.NewWheelEvent(
			delta,
			st.lastMousePos,
			st.lastMousePos,
			st.mods,
		)
		w.HandleEvent(e)
	})
}

// attachKeyboardBridge wires keyboard and text input callbacks.
func attachKeyboardBridge(es gpucontext.EventSource, w *Window, mods *event.Modifiers) {
	es.OnKeyPress(func(key gpucontext.Key, platMods gpucontext.Modifiers) {
		uiKey := translateKey(key)
		uiMods := translateModifiers(platMods)
		// A key event reports the modifiers held BEFORE it, so pressing Shift
		// alone reports no Shift. Fold the key itself in, or holding a modifier
		// and clicking — with no other key in between, which is the whole
		// gesture — would leave the state empty.
		*mods = uiMods | modifierForKey(uiKey)
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
		*mods = uiMods &^ modifierForKey(uiKey)
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
}

// attachPointerBridge wires W3C PointerEventSource for the unified pointer
// pipeline (ADR-049 Phase 3).
//
// When the platform provides PointerEventSource, ALL pointer events flow
// through this function. Enter/Leave are dispatched as MouseEvents directly
// for mouse pointers only (touch/pen enter/leave are ignored to avoid
// disturbing mouse state). Down/Up/Move/Cancel are converted to
// gesture.PointerEvent and fed to HandlePointerEvent, which both feeds the
// gesture arena AND derives MouseEvents for existing widget dispatch.
//
// When the platform does not provide PointerEventSource, this function is
// a no-op (legacy mouse callbacks handle synthesis in attachEventBridge).
func attachPointerBridge(
	es gpucontext.EventSource,
	w *Window,
	st *eventBridgeState,
) {
	pes, ok := es.(gpucontext.PointerEventSource)
	if !ok {
		return
	}

	pes.OnPointer(func(ev gpucontext.PointerEvent) {
		isMouse := ev.PointerType == gpucontext.PointerTypeMouse ||
			ev.PointerType == 0 // zero = unspecified, treat as mouse

		switch ev.Type {
		case gpucontext.PointerEnter:
			// Only mouse enter/leave affect mouse tracking state.
			// Touch/pen enter/leave must not arm scroll fallback or
			// dispatch mouse events (they are separate pointer streams).
			if !isMouse {
				return
			}
			pos := geometry.Pt(float32(ev.X), float32(ev.Y))
			st.lastMousePos = pos
			st.mouseInsideWindow = true
			e := event.NewMouseEvent(
				event.MouseEnter,
				event.ButtonNone,
				st.pressedButtons,
				pos, pos,
				translateModifiers(ev.Modifiers),
			)
			w.HandleEvent(e)

		case gpucontext.PointerLeave:
			if !isMouse {
				return
			}
			pos := geometry.Pt(float32(ev.X), float32(ev.Y))
			st.mouseInsideWindow = false
			e := event.NewMouseEvent(
				event.MouseLeave,
				event.ButtonNone,
				st.pressedButtons,
				pos, pos,
				translateModifiers(ev.Modifiers),
			)
			w.HandleEvent(e)

		case gpucontext.PointerDown:
			// Update button tracking from rich pointer data.
			btn := convertPointerButton(ev.Button)
			st.pressedButtons |= buttonToState(btn)
			pos := geometry.Pt(float32(ev.X), float32(ev.Y))
			st.lastMousePos = pos
			// Unified path: convert and feed through HandlePointerEvent.
			if gev, ok := convertPointerEvent(ev); ok {
				w.HandlePointerEvent(&gev)
			}

		case gpucontext.PointerUp:
			btn := convertPointerButton(ev.Button)
			st.pressedButtons &^= buttonToState(btn)
			pos := geometry.Pt(float32(ev.X), float32(ev.Y))
			st.lastMousePos = pos
			if gev, ok := convertPointerEvent(ev); ok {
				w.HandlePointerEvent(&gev)
			}

		case gpucontext.PointerMove:
			pos := geometry.Pt(float32(ev.X), float32(ev.Y))
			st.lastMousePos = pos
			if gev, ok := convertPointerEvent(ev); ok {
				w.HandlePointerEvent(&gev)
			}

		case gpucontext.PointerCancel:
			// Only cancel mouse state for mouse cancels; touch/pen
			// cancels must not disturb mouse capture or held buttons.
			if isMouse {
				st.mouseInsideWindow = false
				st.pressedButtons = 0
				w.cancelPointerState()
			}
			if gev, ok := convertPointerEvent(ev); ok {
				w.HandlePointerEvent(&gev)
			}
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
