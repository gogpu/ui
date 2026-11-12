package dropdown

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// recordingOverlayManager is a no-op overlay manager for Open tests.
type recordingOverlayManager struct{}

func (m *recordingOverlayManager) PushOverlay(_ widget.Widget, _ func()) {}
func (m *recordingOverlayManager) PopOverlay()                           {}
func (m *recordingOverlayManager) RemoveOverlay(_ widget.Widget)         {}

func TestItemsSignal_OpenBuildsMenuFromResolvedItems(t *testing.T) {
	sig := state.NewSignal([]string{"A", "B", "C"})
	dd := New(ItemsSignal(sig), Selected(2))
	dd.SetBounds(geometry.NewRect(0, 0, 200, 48))

	ctx := widget.NewContext()
	ctx.SetOverlayManager(&recordingOverlayManager{})
	ctx.SetWindowSize(geometry.Sz(800, 600))

	dd.Open(ctx)
	if dd.menuWidget == nil {
		t.Fatal("Open should create a menu widget")
	}
	if got := len(dd.menuWidget.items); got != 3 {
		t.Fatalf("menu items = %d, want 3", got)
	}
	if got := dd.menuWidget.selectedIndex; got != 2 {
		t.Errorf("menu selectedIndex = %d, want 2", got)
	}

	// A change while the menu is open leaves the open snapshot alone: the menu
	// picks up the new list on the next Open.
	sig.Set([]string{"A", "B"})
	if got := len(dd.menuWidget.items); got != 3 {
		t.Errorf("open menu items after change = %d, want 3 (snapshot)", got)
	}

	dd.Close(ctx)
	dd.Open(ctx)

	if got := len(dd.menuWidget.items); got != 2 {
		t.Errorf("menu items after reopen = %d, want 2", got)
	}
	if got := dd.menuWidget.selectedIndex; got != 1 {
		t.Errorf("menu selectedIndex after reopen = %d, want 1 (clamped)", got)
	}
}

// openForPick opens dd on a context with a no-op overlay manager, so a test can
// drive the open menu's row callback the way a click on a row does.
func openForPick(t *testing.T, dd *Widget) {
	t.Helper()
	dd.SetBounds(geometry.NewRect(0, 0, 200, 48))
	ctx := widget.NewContext()
	ctx.SetOverlayManager(&recordingOverlayManager{})
	ctx.SetWindowSize(geometry.Sz(800, 600))
	dd.Open(ctx)
	if dd.menuWidget == nil {
		t.Fatal("Open should create a menu widget")
	}
}

type pick struct {
	index int
	value string
}

// TestItemsSignal_PickAfterChange_ReportsWhatTheUserSaw: the open menu shows
// its snapshot. A pick made after the bound list changed must report the item
// the user clicked, at its index in the current list, never whatever now sits
// at the clicked row (LEAD-023).
func TestItemsSignal_PickAfterChange_ReportsWhatTheUserSaw(t *testing.T) {
	sig := state.NewSignal([]string{"Apple", "Banana", "Cherry"})
	var got []pick
	dd := New(ItemsSignal(sig), OnChange(func(i int, v string) { got = append(got, pick{i, v}) }))
	openForPick(t, dd)

	// The menu shows [Apple, Banana, Cherry]; the list becomes [Banana, Cherry]
	// and the user clicks row 2, "Cherry", which now sits at index 1.
	sig.Set([]string{"Banana", "Cherry"})
	dd.menuWidget.onSelect(2)

	if len(got) != 1 || got[0] != (pick{1, "Cherry"}) {
		t.Fatalf("onChange calls = %v, want [{1 Cherry}]", got)
	}
	if dd.SelectedIndex() != 1 || dd.SelectedValue() != "Cherry" {
		t.Errorf("selection = %d %q, want 1 Cherry", dd.SelectedIndex(), dd.SelectedValue())
	}
}

// TestItemsSignal_PickOfRemovedItem_IsRefused: an item that left the list can
// no longer be selected. The pick changes nothing and closes the menu, instead
// of selecting an item the user never saw.
func TestItemsSignal_PickOfRemovedItem_IsRefused(t *testing.T) {
	sig := state.NewSignal([]string{"Apple", "Banana", "Cherry"})
	var got []pick
	dd := New(ItemsSignal(sig), Selected(0), OnChange(func(i int, v string) { got = append(got, pick{i, v}) }))
	openForPick(t, dd)

	sig.Set([]string{"Xylophone", "Zebra", "Yak"})
	dd.menuWidget.onSelect(1) // "Banana", gone from the list

	if len(got) != 0 {
		t.Fatalf("onChange calls = %v, want none (Banana is gone)", got)
	}
	if dd.SelectedValue() != "Xylophone" {
		t.Errorf("SelectedValue = %q, want Xylophone (selection unchanged at index 0)", dd.SelectedValue())
	}
	if dd.IsOpen() {
		t.Error("the menu should close after a refused pick")
	}
}

// TestItemsSignal_PickAfterShrink_StillSelects: a shorter list that still holds
// the clicked item selects it, instead of dropping the pick because the clicked
// row is past the end of the new list.
func TestItemsSignal_PickAfterShrink_StillSelects(t *testing.T) {
	sig := state.NewSignal([]string{"A", "B", "C"})
	var got []pick
	dd := New(ItemsSignal(sig), OnChange(func(i int, v string) { got = append(got, pick{i, v}) }))
	openForPick(t, dd)

	sig.Set([]string{"C"})
	dd.menuWidget.onSelect(2) // "C"

	if len(got) != 1 || got[0] != (pick{0, "C"}) {
		t.Fatalf("onChange calls = %v, want [{0 C}]", got)
	}
}
