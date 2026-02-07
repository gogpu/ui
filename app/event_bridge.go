package app

import (
	"github.com/gogpu/gpucontext"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
)

// attachEventBridge registers event callbacks on the EventSource that
// translate platform events into ui/event types and dispatch them to
// the Window.
//
// This function is called once during App creation when an EventSource
// is provided. The callbacks are invoked on the main thread by the host
// application's event loop.
func attachEventBridge(es gpucontext.EventSource, w *Window) {
	es.OnMouseMove(func(x, y float64) {
		pos := geometry.Pt(float32(x), float32(y))
		e := event.NewMouseEvent(
			event.MouseMove,
			event.ButtonNone,
			0, // no buttons pressed for move
			pos,
			pos, // global position same as local for root dispatch
			event.ModNone,
		)
		w.HandleEvent(e)
	})

	es.OnMousePress(func(button gpucontext.MouseButton, x, y float64) {
		pos := geometry.Pt(float32(x), float32(y))
		btn := translateMouseButton(button)
		e := event.NewMouseEvent(
			event.MousePress,
			btn,
			buttonToState(btn),
			pos,
			pos,
			event.ModNone,
		)
		w.HandleEvent(e)
	})

	es.OnMouseRelease(func(button gpucontext.MouseButton, x, y float64) {
		pos := geometry.Pt(float32(x), float32(y))
		btn := translateMouseButton(button)
		e := event.NewMouseEvent(
			event.MouseRelease,
			btn,
			0, // button is now released
			pos,
			pos,
			event.ModNone,
		)
		w.HandleEvent(e)
	})

	es.OnKeyPress(func(key gpucontext.Key, mods gpucontext.Modifiers) {
		uiKey := translateKey(key)
		uiMods := translateModifiers(mods)
		e := event.NewKeyEvent(
			event.KeyPress,
			uiKey,
			0, // rune is delivered via OnTextInput
			uiMods,
		)
		w.HandleEvent(e)
	})

	es.OnKeyRelease(func(key gpucontext.Key, mods gpucontext.Modifiers) {
		uiKey := translateKey(key)
		uiMods := translateModifiers(mods)
		e := event.NewKeyEvent(
			event.KeyRelease,
			uiKey,
			0,
			uiMods,
		)
		w.HandleEvent(e)
	})

	es.OnScroll(func(dx, dy float64) {
		delta := geometry.Pt(float32(dx), float32(dy))
		e := event.NewWheelEvent(
			delta,
			geometry.Pt(0, 0), // scroll position is not provided by EventSource
			geometry.Pt(0, 0),
			event.ModNone,
		)
		w.HandleEvent(e)
	})

	es.OnResize(func(width, height int) {
		w.HandleResize(width, height)
	})

	es.OnFocus(func(focused bool) {
		w.HandleFocusChange(focused)
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
