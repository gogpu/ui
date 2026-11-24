package app_test

import (
	"testing"

	"github.com/gogpu/ui/app"
	"github.com/gogpu/ui/core/dialog"
	"github.com/gogpu/ui/core/dropdown"
	"github.com/gogpu/ui/event"
)

func TestDropdownSelectionKeepsCallbackDialog(t *testing.T) {
	application := app.New()
	window := application.Window()
	ctx := window.Context()

	var opened *dialog.Widget
	selector := dropdown.New(
		dropdown.Items("Red", "Green"),
		dropdown.Selected(0),
		dropdown.OnChange(func(index int, value string) {
			if index != 1 || value != "Green" {
				t.Errorf("callback got (%d, %q), want (1, Green)", index, value)
			}
			opened = dialog.Alert("Picked", value, nil)
			opened.Show(ctx)
		}),
	)
	application.SetRoot(selector)
	application.Frame()
	selector.Open(ctx)
	application.Frame()
	if count := window.OverlayCount(); count != 1 {
		t.Fatalf("opening menu produced %d overlays, want 1", count)
	}

	application.HandleEvent(event.NewKeyEvent(event.KeyPress, event.KeyDown, 0, event.ModNone))
	application.HandleEvent(event.NewKeyEvent(event.KeyPress, event.KeyEnter, 0, event.ModNone))
	application.Frame()
	if selector.IsOpen() || selector.SelectedIndex() != 1 {
		t.Fatal("selection should update the index and close the menu")
	}
	if opened == nil || !opened.IsOpen() || window.OverlayCount() != 1 {
		t.Fatalf("callback dialog should own the remaining overlay, got %d", window.OverlayCount())
	}
	opened.Close(ctx)
	if count := window.OverlayCount(); count != 0 {
		t.Fatalf("closing the dialog left %d overlays, want 0", count)
	}
}
