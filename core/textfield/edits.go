package textfield

import (
	"unicode/utf8"

	"github.com/gogpu/ui/widget"
)

type editSnapshot struct {
	text string
	sel  selection
}

const maxUndoEdits = 64

func pushEdit(stack []editSnapshot, edit editSnapshot) []editSnapshot {
	if len(stack) == maxUndoEdits {
		copy(stack, stack[1:])
		stack = stack[:len(stack)-1]
	}
	return append(stack, edit)
}

func (w *Widget) applyEdit(text string, cursor int) {
	before := w.resolvedText()
	if before != text && w.cfg.inputType != TypePassword {
		w.undo = pushEdit(w.undo, editSnapshot{before, w.sel})
		w.redo = nil
	}
	w.sel.SetCursor(cursor)
	w.setText(text)
}

func (w *Widget) restoreEdit(ctx widget.Context, from, to *[]editSnapshot) bool {
	if len(*from) == 0 {
		return true
	}
	last := len(*from) - 1
	snapshot := (*from)[last]
	(*from)[last] = editSnapshot{}
	*from = (*from)[:last]
	*to = pushEdit(*to, editSnapshot{w.resolvedText(), w.sel})
	w.sel = snapshot.sel
	w.setText(snapshot.text)
	w.notifyChange(ctx)
	return true
}

// An outside update replaces the value that the edit history was based on.
func (w *Widget) syncExternalText() {
	if text := w.resolvedText(); text != w.cfg.value {
		w.cfg.value = text
		w.undo, w.redo = nil, nil
		w.sel.SetCursor(clampPos(w.sel.cursor, utf8.RuneCountInString(text)))
		w.validate()
	}
}
