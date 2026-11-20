package textfield_test

import (
	"testing"

	"github.com/gogpu/ui/app"
	"github.com/gogpu/ui/core/menu"
	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"
)

type recordingEditMenu struct {
	menu.DefaultPainter
	items []menu.MenuItem
}

func (p *recordingEditMenu) PaintMenu(_ widget.Canvas, st *menu.MenuPaintState) {
	p.items = st.Items
}

func TestTextFieldEditMenuCommands(t *testing.T) {
	clipboard := &testClipboard{}
	widget.RegisterClipboardProvider(clipboard)
	defer widget.RegisterClipboardProvider(nil)
	painter := &recordingEditMenu{}
	field := textfield.New(textfield.InitialValue("начало конец"), textfield.ContextMenuPainterOpt(painter))
	a := app.New()
	win := a.Window()
	defer win.Close()
	a.SetRoot(primitives.VBox(field))
	a.Frame()
	win.DrawTo(&uitest.MockCanvas{})
	win.Context().RequestFocus(field)
	key := func(k event.Key, mods event.Modifiers) {
		win.HandleEvent(event.NewKeyEvent(event.KeyPress, k, 0, mods))
	}
	key(event.KeyLeft, event.ModCtrl|event.ModShift)
	pos := field.ScreenBounds().Center()
	win.HandleEvent(event.NewMouseEvent(event.MousePress, event.ButtonRight, event.ButtonStateRight, pos, pos, event.ModNone))
	win.HandleEvent(event.NewMouseEvent(event.MouseRelease, event.ButtonRight, 0, pos, pos, event.ModNone))

	activate := func(label string) {
		t.Helper()
		if win.OverlayCount() == 0 {
			key(event.KeyMenu, event.ModNone)
		}
		if win.OverlayCount() != 1 {
			t.Fatalf("opening %s menu, overlay count=%d", label, win.OverlayCount())
		}
		a.Frame()
		win.DrawOverlays(&uitest.MockCanvas{})
		panel := win.OverlayContentWidgets()[0].(menu.PanelState)
		for range painter.items {
			key(event.KeyDown, event.ModNone)
			i := panel.HighlightedIndex()
			if i >= 0 && painter.items[i].Label == label {
				key(event.KeyEnter, event.ModNone)
				if win.OverlayCount() != 0 || !win.Context().IsFocused(field) {
					t.Fatalf("%s did not close the menu and keep focus", label)
				}
				return
			}
		}
		t.Fatalf("%s is missing or disabled", label)
	}
	win.HandleEvent(event.NewKeyEvent(event.KeyPress, event.KeyUnknown, 'x', event.ModNone))
	activate("Copy")
	if clipboard.text != "конец" || field.Text() != "начало конец" {
		t.Fatalf("copy or underlying input changed the field: %q, clipboard=%q", field.Text(), clipboard.text)
	}
	activate("Cut")
	if field.Text() != "начало " {
		t.Fatalf("cut left %q", field.Text())
	}
	activate("Undo")
	if from, to := field.Selection(); field.Text() != "начало конец" || from != 7 || to != 12 {
		t.Fatalf("undo did not restore the value and selection: %q [%d:%d]", field.Text(), from, to)
	}
	activate("Redo")
	activate("Paste")
	activate("Select All")
	activate("Delete")
	if field.Text() != "" {
		t.Fatalf("deleting the entire selection left %q", field.Text())
	}
	key(event.KeyZ, event.ModCtrl)
	if field.Text() != "начало конец" {
		t.Fatalf("undo after menu delete restored %q", field.Text())
	}
}

func TestTextFieldUndoRespectsExternalValueChanges(t *testing.T) {
	value := state.NewSignal("старое")
	field := textfield.New(textfield.ValueSignal(value), textfield.MaxLength(6))
	field.SetFocused(true)
	ctx := widget.NewContext()
	pressKey(field, ctx, event.KeyA, event.ModCtrl)
	typeRune(field, ctx, 'ч')
	pressKey(field, ctx, event.KeyZ, event.ModCtrl)
	if value.Get() != "старое" {
		t.Fatalf("undo of replacement restored %q", value.Get())
	}
	pressKey(field, ctx, event.KeyZ, event.ModCtrl|event.ModShift)
	if value.Get() != "ч" {
		t.Fatalf("redo of replacement restored %q", value.Get())
	}
	value.Set("внешнее")
	pressKey(field, ctx, event.KeyZ, event.ModCtrl)
	if value.Get() != "внешнее" {
		t.Fatalf("undo overwrote the external value: %q", value.Get())
	}
}
