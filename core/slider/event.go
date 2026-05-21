package slider

import (
	"math"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// handleEvent processes input events for the slider widget.
// It manages hover, drag, click-on-track, and keyboard navigation.
func handleEvent(w *Widget, ctx widget.Context, e event.Event) bool {
	// Disabled sliders ignore all interaction.
	if w.cfg.ResolvedDisabled() {
		return false
	}

	switch ev := e.(type) {
	case *event.MouseEvent:
		return handleMouseEvent(w, ctx, ev)
	case *event.KeyEvent:
		return handleKeyEvent(w, ctx, ev)
	default:
		return false
	}
}

// handleMouseEvent processes mouse events for hover, press, drag, and click-on-track.
func handleMouseEvent(w *Widget, ctx widget.Context, e *event.MouseEvent) bool {
	switch e.MouseType {
	case event.MouseEnter:
		w.interaction = stateHover
		ctx.SetCursor(widget.CursorPointer)
		w.SetNeedsRedraw(true)
		ctx.InvalidateRect(w.Bounds())
		return true

	case event.MouseLeave:
		if w.interaction != stateDragging {
			w.interaction = stateNormal
			ctx.SetCursor(widget.CursorDefault)
		}
		w.SetNeedsRedraw(true)
		ctx.InvalidateRect(w.Bounds())
		return true

	case event.MousePress:
		return handleMousePress(w, ctx, e)

	case event.MouseRelease:
		return handleMouseRelease(w, ctx, e)

	case event.MouseMove:
		return handleMouseMove(w, ctx, e)

	default:
		return false
	}
}

// handleMousePress handles a mouse button press on the slider.
func handleMousePress(w *Widget, ctx widget.Context, e *event.MouseEvent) bool {
	if e.Button != event.ButtonLeft {
		return false
	}

	ctx.RequestFocus(w)
	w.interaction = stateDragging
	ctx.SetCursor(widget.CursorPointer)

	// ADR-031: Capture pointer so mouse events are delivered directly
	// during drag, even when the cursor moves outside widget bounds.
	if pc, ok := ctx.(widget.PointerCapturer); ok {
		pc.CapturePointer(w)
	}

	// Set value based on click position.
	newValue := valueFromPosition(w, e.Position)
	setValue(w, ctx, newValue)

	return true
}

// handleMouseRelease handles a mouse button release.
func handleMouseRelease(w *Widget, ctx widget.Context, e *event.MouseEvent) bool {
	if e.Button != event.ButtonLeft {
		return false
	}

	wasDragging := w.interaction == stateDragging

	// ADR-031: Release pointer capture explicitly. The Window also
	// auto-releases on MouseRelease when all buttons are up, but
	// explicit release is cleaner and handles edge cases.
	if pc, ok := ctx.(widget.PointerCapturer); ok {
		pc.ReleasePointer(w)
	}

	if w.Bounds().Contains(e.Position) {
		w.interaction = stateHover
	} else {
		w.interaction = stateNormal
		ctx.SetCursor(widget.CursorDefault)
	}
	w.SetNeedsRedraw(true)
	ctx.InvalidateRect(w.Bounds())

	return wasDragging
}

// handleMouseMove handles mouse movement for drag tracking.
func handleMouseMove(w *Widget, ctx widget.Context, e *event.MouseEvent) bool {
	if w.interaction != stateDragging {
		return false
	}

	newValue := valueFromPosition(w, e.Position)
	setValue(w, ctx, newValue)
	ctx.SetCursor(widget.CursorPointer) // Maintain drag cursor on every move

	return true
}

// handleKeyEvent processes keyboard events for slider value adjustment.
func handleKeyEvent(w *Widget, ctx widget.Context, e *event.KeyEvent) bool {
	if !w.IsFocused() {
		return false
	}

	if e.KeyType != event.KeyPress && e.KeyType != event.KeyRepeat {
		return false
	}

	rangeVal := w.cfg.maxVal - w.cfg.minVal
	if rangeVal <= 0 {
		return false
	}

	stepVal := w.cfg.step
	if stepVal <= 0 {
		// Default keyboard step: 1% of range or 1, whichever is smaller.
		stepVal = rangeVal * keyboardStepPercent
		if stepVal > keyboardStepMax {
			stepVal = keyboardStepMax
		}
	}

	largeStep := rangeVal * keyboardLargeStepPercent

	current := w.cfg.ResolvedValue()

	switch e.Key {
	case event.KeyRight, event.KeyUp:
		setValue(w, ctx, current+stepVal)
		return true

	case event.KeyLeft, event.KeyDown:
		setValue(w, ctx, current-stepVal)
		return true

	case event.KeyPageUp:
		setValue(w, ctx, current+largeStep)
		return true

	case event.KeyPageDown:
		setValue(w, ctx, current-largeStep)
		return true

	case event.KeyHome:
		setValue(w, ctx, w.cfg.minVal)
		return true

	case event.KeyEnd:
		setValue(w, ctx, w.cfg.maxVal)
		return true

	default:
		return false
	}
}

// valueFromPosition converts a mouse position to a slider value.
func valueFromPosition(w *Widget, pos geometry.Point) float32 {
	bounds := w.Bounds()
	rangeVal := w.cfg.maxVal - w.cfg.minVal
	if rangeVal <= 0 {
		return w.cfg.minVal
	}

	var progress float32
	if w.cfg.orientation == Vertical {
		trackTop := bounds.Min.Y + thumbRadius
		trackBottom := bounds.Max.Y - thumbRadius
		trackLen := trackBottom - trackTop
		if trackLen <= 0 {
			return w.cfg.minVal
		}
		// Vertical: bottom = 0, top = 1.
		progress = (trackBottom - pos.Y) / trackLen
	} else {
		trackLeft := bounds.Min.X + thumbRadius
		trackRight := bounds.Max.X - thumbRadius
		trackWidth := trackRight - trackLeft
		if trackWidth <= 0 {
			return w.cfg.minVal
		}
		progress = (pos.X - trackLeft) / trackWidth
	}

	// Clamp progress to [0, 1].
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	return w.cfg.minVal + progress*rangeVal
}

// setValue updates the slider's value, clamping and snapping as configured,
// then fires the onChange callback and invalidates the widget.
func setValue(w *Widget, ctx widget.Context, rawValue float32) {
	newValue := clampAndSnap(rawValue, w.cfg.minVal, w.cfg.maxVal, w.cfg.step)

	currentValue := w.cfg.ResolvedValue()
	if newValue == currentValue {
		return
	}

	// TWO-WAY: if a ValueSignal is bound, write back the new value.
	if w.cfg.valueSignal != nil {
		w.cfg.valueSignal.Set(newValue)
	} else if w.cfg.valueFn == nil {
		// Update the static value field so subsequent reads reflect the change,
		// unless a dynamic ValueFn is in use (in which case the caller manages state).
		w.cfg.value = newValue
	}

	if w.cfg.onChange != nil {
		w.cfg.onChange(newValue)
	}

	w.SetNeedsRedraw(true)
	ctx.InvalidateRect(w.Bounds())
}

// clampAndSnap clamps val to [minVal, maxVal] and snaps to nearest step if step > 0.
func clampAndSnap(val, minVal, maxVal, step float32) float32 {
	if val < minVal {
		val = minVal
	}
	if val > maxVal {
		val = maxVal
	}

	if step > 0 {
		// Snap to nearest step relative to minVal.
		offset := val - minVal
		steps := float32(math.Round(float64(offset / step)))
		val = minVal + steps*step

		// Re-clamp after snapping.
		if val < minVal {
			val = minVal
		}
		if val > maxVal {
			val = maxVal
		}
	}

	return val
}

// Keyboard navigation constants.
const (
	keyboardStepPercent      float32 = 0.01 // 1% of range
	keyboardStepMax          float32 = 1    // max keyboard step
	keyboardLargeStepPercent float32 = 0.10 // 10% of range for PgUp/PgDn
)
