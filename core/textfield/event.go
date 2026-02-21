package textfield

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/widget"
)

// handleEvent processes input events for the text field widget.
// It manages hover, focus, text editing, and selection states.
func handleEvent(w *Widget, ctx widget.Context, e event.Event) bool {
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

// handleMouseEvent processes mouse events for focus, cursor placement, and selection.
func handleMouseEvent(w *Widget, ctx widget.Context, e *event.MouseEvent) bool {
	switch e.MouseType {
	case event.MouseEnter:
		w.hovered = true
		ctx.SetCursor(widget.CursorText)
		ctx.Invalidate()
		return true

	case event.MouseLeave:
		w.hovered = false
		ctx.SetCursor(widget.CursorDefault)
		ctx.Invalidate()
		return true

	case event.MousePress:
		return handleMousePress(w, ctx, e)

	case event.MouseRelease:
		w.dragging = false
		return true

	case event.MouseDrag:
		return handleMouseDrag(w, ctx, e)

	case event.MouseDoubleClick:
		return handleDoubleClick(w, ctx, e)

	default:
		return false
	}
}

// handleMousePress handles a mouse press event to place the cursor.
func handleMousePress(w *Widget, ctx widget.Context, e *event.MouseEvent) bool {
	if e.Button != event.ButtonLeft {
		return false
	}

	ctx.RequestFocus(w)
	ctx.Invalidate()

	pos := positionFromMouse(w, e)
	if e.Modifiers().IsShift() {
		w.sel.SetCursorKeepSelection(pos)
	} else {
		w.sel.SetCursor(pos)
	}
	w.dragging = true
	return true
}

// handleMouseDrag handles mouse drag for text selection.
func handleMouseDrag(w *Widget, ctx widget.Context, e *event.MouseEvent) bool {
	if !w.dragging {
		return false
	}
	pos := positionFromMouse(w, e)
	w.sel.SetCursorKeepSelection(pos)
	ctx.Invalidate()
	return true
}

// handleDoubleClick selects the word at the click position.
func handleDoubleClick(w *Widget, ctx widget.Context, e *event.MouseEvent) bool {
	if e.Button != event.ButtonLeft {
		return false
	}
	runes := w.textRunes()
	pos := positionFromMouse(w, e)
	start, end := wordBoundsAt(runes, pos)
	w.sel.anchor = start
	w.sel.cursor = end
	ctx.Invalidate()
	return true
}

// positionFromMouse converts a mouse position to a rune index in the text.
func positionFromMouse(w *Widget, e *event.MouseEvent) int {
	bounds := w.Bounds()
	localX := e.Position.X - bounds.Min.X - contentPaddingH

	if localX <= 0 {
		return 0
	}

	charW := defaultFontSize * charWidthRatio
	pos := int(localX / charW)

	runes := w.textRunes()
	return clampPos(pos, len(runes))
}

// handleKeyEvent processes keyboard events for text editing and navigation.
func handleKeyEvent(w *Widget, ctx widget.Context, e *event.KeyEvent) bool {
	if !w.IsFocused() {
		return false
	}

	// Only handle press and repeat events.
	if e.KeyType == event.KeyRelease {
		return false
	}

	shift := e.Modifiers().IsShift()
	ctrl := e.Modifiers().IsCtrl()

	switch {
	case ctrl && e.Key == event.KeyA:
		return handleSelectAll(w, ctx)
	case ctrl && e.Key == event.KeyC:
		return handleCopy(w)
	case ctrl && e.Key == event.KeyX:
		return handleCut(w, ctx)
	case ctrl && e.Key == event.KeyV:
		return handlePaste(w, ctx)
	case e.Key == event.KeyLeft:
		return handleArrowLeft(w, ctx, shift, ctrl)
	case e.Key == event.KeyRight:
		return handleArrowRight(w, ctx, shift, ctrl)
	case e.Key == event.KeyHome:
		return handleHome(w, ctx, shift)
	case e.Key == event.KeyEnd:
		return handleEnd(w, ctx, shift)
	case e.Key == event.KeyBackspace:
		return handleBackspace(w, ctx)
	case e.Key == event.KeyDelete:
		return handleDelete(w, ctx)
	case e.Key == event.KeyEnter, e.Key == event.KeyNumpadEnter:
		return handleEnter(w)
	case e.Key == event.KeyTab:
		return false // Let tab propagate for focus navigation
	default:
		return handleCharInput(w, ctx, e)
	}
}

// handleSelectAll selects all text in the field.
func handleSelectAll(w *Widget, ctx widget.Context) bool {
	runes := w.textRunes()
	w.sel.SelectAll(len(runes))
	ctx.Invalidate()
	return true
}

// handleCopy copies selected text to the clipboard.
func handleCopy(w *Widget) bool {
	if !w.sel.HasSelection() {
		return true
	}
	runes := w.textRunes()
	start, end := w.sel.OrderedRange()
	w.sel.copyToClipboard(string(runes[start:end]))
	return true
}

// handleCut cuts selected text to the clipboard.
func handleCut(w *Widget, ctx widget.Context) bool {
	if !w.sel.HasSelection() {
		return true
	}
	runes := w.textRunes()
	start, end := w.sel.OrderedRange()
	w.sel.copyToClipboard(string(runes[start:end]))
	w.deleteSelection()
	w.notifyChange(ctx)
	return true
}

// handlePaste inserts clipboard text at the cursor position.
func handlePaste(w *Widget, ctx widget.Context) bool {
	text := w.sel.pasteFromClipboard()
	if text == "" {
		return true
	}

	w.deleteSelection()
	w.insertText(text)
	w.notifyChange(ctx)
	return true
}

// handleArrowLeft moves the cursor left, optionally extending selection.
func handleArrowLeft(w *Widget, ctx widget.Context, shift, ctrl bool) bool {
	runes := w.textRunes()
	var newPos int
	if ctrl {
		newPos = prevWordBoundary(runes, w.sel.cursor)
	} else {
		if !shift && w.sel.HasSelection() {
			start, _ := w.sel.OrderedRange()
			newPos = start
		} else {
			newPos = clampPos(w.sel.cursor-1, len(runes))
		}
	}

	if shift {
		w.sel.SetCursorKeepSelection(newPos)
	} else {
		w.sel.SetCursor(newPos)
	}
	ctx.Invalidate()
	return true
}

// handleArrowRight moves the cursor right, optionally extending selection.
func handleArrowRight(w *Widget, ctx widget.Context, shift, ctrl bool) bool {
	runes := w.textRunes()
	var newPos int
	if ctrl {
		newPos = nextWordBoundary(runes, w.sel.cursor)
	} else {
		if !shift && w.sel.HasSelection() {
			_, end := w.sel.OrderedRange()
			newPos = end
		} else {
			newPos = clampPos(w.sel.cursor+1, len(runes))
		}
	}

	if shift {
		w.sel.SetCursorKeepSelection(newPos)
	} else {
		w.sel.SetCursor(newPos)
	}
	ctx.Invalidate()
	return true
}

// handleHome moves the cursor to the beginning of the line.
func handleHome(w *Widget, ctx widget.Context, shift bool) bool {
	if shift {
		w.sel.SetCursorKeepSelection(0)
	} else {
		w.sel.SetCursor(0)
	}
	ctx.Invalidate()
	return true
}

// handleEnd moves the cursor to the end of the line.
func handleEnd(w *Widget, ctx widget.Context, shift bool) bool {
	runes := w.textRunes()
	if shift {
		w.sel.SetCursorKeepSelection(len(runes))
	} else {
		w.sel.SetCursor(len(runes))
	}
	ctx.Invalidate()
	return true
}

// handleBackspace deletes the character before the cursor or the selection.
func handleBackspace(w *Widget, ctx widget.Context) bool {
	if w.sel.HasSelection() {
		w.deleteSelection()
		w.notifyChange(ctx)
		return true
	}

	if w.sel.cursor == 0 {
		return true
	}

	runes := w.textRunes()
	pos := w.sel.cursor
	newRunes := make([]rune, 0, len(runes)-1)
	newRunes = append(newRunes, runes[:pos-1]...)
	newRunes = append(newRunes, runes[pos:]...)
	w.setText(string(newRunes))
	w.sel.SetCursor(pos - 1)
	w.notifyChange(ctx)
	return true
}

// handleDelete deletes the character after the cursor or the selection.
func handleDelete(w *Widget, ctx widget.Context) bool {
	if w.sel.HasSelection() {
		w.deleteSelection()
		w.notifyChange(ctx)
		return true
	}

	runes := w.textRunes()
	if w.sel.cursor >= len(runes) {
		return true
	}

	pos := w.sel.cursor
	newRunes := make([]rune, 0, len(runes)-1)
	newRunes = append(newRunes, runes[:pos]...)
	newRunes = append(newRunes, runes[pos+1:]...)
	w.setText(string(newRunes))
	w.notifyChange(ctx)
	return true
}

// handleEnter submits the text field value.
func handleEnter(w *Widget) bool {
	if w.cfg.onSubmit != nil {
		w.cfg.onSubmit(w.resolvedText())
	}
	return true
}

// handleCharInput inserts a printable character at the cursor position.
func handleCharInput(w *Widget, ctx widget.Context, e *event.KeyEvent) bool {
	if !e.HasRune() {
		return false
	}

	r := e.Rune
	// Filter out control characters.
	if r < ' ' {
		return false
	}

	w.deleteSelection()

	// Check max length.
	runes := w.textRunes()
	if w.cfg.maxLength > 0 && len(runes) >= w.cfg.maxLength {
		return true
	}

	w.insertText(string(r))
	w.notifyChange(ctx)
	return true
}
