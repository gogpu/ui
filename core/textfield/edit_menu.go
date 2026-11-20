package textfield

import (
	"github.com/gogpu/ui/core/menu"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

func (w *Widget) canCopy() bool {
	return w.cfg.inputType != TypePassword && w.sel.HasSelection()
}

func (w *Widget) showEditMenu(ctx widget.Context, at geometry.Point, click *geometry.Point) bool {
	if ctx.OverlayManager() == nil {
		return false
	}
	if w.closeEditMenu != nil {
		w.closeEditMenu()
	}
	if click != nil && !w.sel.HasSelection() {
		w.sel.SetCursor(clampPos(positionFromLocal(w, *click), len(w.textRunes())))
	}
	ctx.RequestFocus(w)
	w.SetNeedsRedraw(true)

	item := func(label, shortcut string, enabled bool, action func()) menu.MenuItem {
		if !enabled {
			return menu.ItemDisabled(label, shortcut)
		}
		return menu.Item(label, shortcut, func() {
			if widget.IsDetached(w) || w.cfg.ResolvedDisabled() {
				return
			}
			w.syncExternalText()
			ctx.RequestFocus(w)
			action()
		})
	}
	entries := []menu.MenuItem{
		item("Undo", "Ctrl+Z", len(w.undo) != 0, func() { w.restoreEdit(ctx, &w.undo, &w.redo) }),
		item("Redo", "Ctrl+Y", len(w.redo) != 0, func() { w.restoreEdit(ctx, &w.redo, &w.undo) }),
		menu.Sep(),
		item("Cut", "Ctrl+X", w.canCopy(), func() { handleCut(w, ctx) }),
		item("Copy", "Ctrl+C", w.canCopy(), func() { handleCopy(w) }),
		item("Paste", "Ctrl+V", widget.ClipboardRead() != "", func() { handlePaste(w, ctx) }),
		item("Delete", "Del", w.sel.HasSelection(), func() {
			if w.sel.HasSelection() {
				w.deleteSelection()
				w.notifyChange(ctx)
			}
		}),
		menu.Sep(),
		item("Select All", "Ctrl+A", w.resolvedText() != "", func() { handleSelectAll(w, ctx) }),
	}
	opts := []menu.ContextMenuOption{menu.ContextCaptureKeys()}
	if w.cfg.menuPainter != nil {
		opts = append(opts, menu.ContextPainterOpt(w.cfg.menuPainter))
	}
	cm := menu.NewContextMenu(entries, opts...)
	cm.Show(ctx, at)
	w.closeEditMenu = func() { cm.Hide(ctx) }
	return true
}
