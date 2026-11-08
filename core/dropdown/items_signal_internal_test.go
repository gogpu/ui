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
