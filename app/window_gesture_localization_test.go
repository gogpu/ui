package app

// Regression tests for pointer-event localization in the unified pointer
// pipeline (ADR-049): events delivered to gesture recognizers must be in the
// recognizer's widget-local coordinate space (the same space as the derived
// MouseEvents its widget receives), not in window coordinates.
//
// Regression history: recognizers received the raw window-coordinate event,
// so widgets away from the window origin (e.g. the IDE demo's project tree,
// offset by the title bar and tool stripe) mis-hit their own geometry. A
// tree click selected a row shifted down by the tree's window offset.

import (
	"testing"

	"github.com/gogpu/ui/core/treeview"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/gesture"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/state"
)

// TestHandlePointerEvent_LocalizesRecognizerCoordinates verifies that a
// ClickRecognizer collected from a widget placed at a non-zero window offset
// receives LocalPosition in the widget's coordinate space while
// GlobalPosition stays in window space, and that the derived MouseEvent
// delivered to the widget carries the same position.
func TestHandlePointerEvent_LocalizesRecognizerCoordinates(t *testing.T) {
	a := New()
	w := a.Window()

	var gotLocal, gotGlobal geometry.Point
	var fired bool
	click := gesture.NewClickRecognizer(gesture.ClickConfig{
		OnClick: func(d gesture.ClickDetails) {
			fired = true
			gotLocal = d.LocalPosition
			gotGlobal = d.GlobalPosition
		},
	})

	// Widget at window offset (300, 200), nested in a container that places
	// it there so both dispatch paths agree on its position.
	target := newGestureAwareMock(click)
	target.SetBounds(geometry.NewRect(0, 0, 200, 100))
	target.SetScreenOrigin(geometry.Pt(300, 200))

	inner := primitives.Box(target)
	inner.SetBounds(geometry.NewRect(300, 200, 200, 100))
	inner.SetScreenOrigin(geometry.Pt(300, 200))
	root := primitives.VBox(inner)
	root.SetBounds(geometry.NewRect(0, 0, 1024, 700))
	root.SetScreenOrigin(geometry.Pt(0, 0))
	w.SetRoot(root)

	// Click at window (350, 250) — widget-local (50, 50).
	clickAt(w, geometry.Pt(350, 250), geometry.Pt(352, 251))

	if !fired {
		t.Fatal("OnClick never fired")
	}
	if want := geometry.Pt(52, 51); gotLocal != want {
		t.Errorf("ClickDetails.LocalPosition = %v, want %v (widget-local)", gotLocal, want)
	}
	if want := geometry.Pt(352, 251); gotGlobal != want {
		t.Errorf("ClickDetails.GlobalPosition = %v, want %v (window)", gotGlobal, want)
	}

	// The derived MouseEvent delivered to the widget's Event handler must be
	// in the same coordinate space as the recognizer's LocalPosition.
	me, ok := target.lastEvent.(*event.MouseEvent)
	if !ok {
		t.Fatalf("last event = %T, want *event.MouseEvent", target.lastEvent)
	}
	if me.Position != geometry.Pt(52, 51) {
		t.Errorf("derived MouseEvent position = %v, want (52, 51)", me.Position)
	}
}

// TestHandlePointerEvent_TreeViewAtWindowOffset_SelectsClickedRow reproduces
// the IDE-demo layout (title bar 40px, left stripe 64px, 220px tree panel,
// 24px rows) and verifies that clicking a row selects exactly that row —
// both when clicking the label and when clicking the empty space beside it.
func TestHandlePointerEvent_TreeViewAtWindowOffset_SelectsClickedRow(t *testing.T) {
	var selections []string
	selected := state.NewSignal("")

	tree := treeview.New(
		treeview.Root(&treeview.TreeNode{
			ID: "root", Label: "myproject", Expanded: true,
			Children: []*treeview.TreeNode{
				{ID: "cmd", Label: "cmd", Expanded: true, Children: []*treeview.TreeNode{
					{ID: "cmd-main", Label: "main.go"},
				}},
				{ID: "pkg", Label: "pkg", Expanded: true, Children: []*treeview.TreeNode{
					{ID: "pkg-server", Label: "server"},
					{ID: "pkg-config", Label: "config"},
				}},
				{ID: "internal", Label: "internal"},
				{ID: "go-mod", Label: "go.mod"},
				{ID: "go-sum", Label: "go.sum"},
				{ID: "readme", Label: "README.md"},
				{ID: "license", Label: "LICENSE"},
				{ID: "makefile", Label: "Makefile"},
			},
		}),
		treeview.ItemHeight(24),
		treeview.SelectionModeOpt(treeview.SelectionSingle),
		treeview.SelectedNodeSignal(selected),
		treeview.OnSelect(func(n *treeview.TreeNode) {
			selections = append(selections, n.Label)
		}),
	)

	// Layout mimicking examples/ide/main.go: 40px title bar, 64px stripe,
	// then the tree panel. Child bounds are parent-local, as set by Layout.
	titleBar := primitives.Box()
	middleRow := primitives.HBox(
		primitives.Box(), // left stripe
		primitives.VBox(tree),
	)
	root := primitives.VBox(titleBar, primitives.Expanded(middleRow))

	root.SetBounds(geometry.NewRect(0, 0, 1024, 700))
	root.SetScreenOrigin(geometry.Pt(0, 0))
	titleBar.SetBounds(geometry.NewRect(0, 0, 1024, 40))
	titleBar.SetScreenOrigin(geometry.Pt(0, 0))
	expanded := root.Children()[1].(*primitives.ExpandedWidget)
	expanded.SetBounds(geometry.NewRect(0, 40, 1024, 632))
	expanded.SetScreenOrigin(geometry.Pt(0, 40))
	middleRow.SetBounds(geometry.NewRect(0, 0, 1024, 632))
	middleRow.SetScreenOrigin(geometry.Pt(0, 40))
	stripe := middleRow.Children()[0].(*primitives.BoxWidget)
	stripe.SetBounds(geometry.NewRect(0, 0, 64, 632))
	stripe.SetScreenOrigin(geometry.Pt(0, 40))
	treeWrap := middleRow.Children()[1].(*primitives.BoxWidget)
	treeWrap.SetBounds(geometry.NewRect(64, 0, 220, 632))
	treeWrap.SetScreenOrigin(geometry.Pt(64, 40))
	tree.SetBounds(geometry.NewRect(0, 0, 220, 632))
	tree.SetScreenOrigin(geometry.Pt(64, 40))

	a := New()
	w := a.Window()
	w.SetRoot(root)

	// Flattened rows (24px each, window offset 40): row 2 "main.go" spans
	// window Y [88, 112); row 3 "pkg" spans [112, 136); row 11 "Makefile"
	// spans [304, 328).
	cases := []struct {
		name      string
		pos       geometry.Point
		wantID    string
		wantFires int
	}{
		{"click on row label", geometry.Pt(100, 100), "cmd-main", 1},
		// Same row via the empty space to the right of the label: the
		// selection must not change and no second OnSelect may fire.
		{"click empty space in same row", geometry.Pt(250, 100), "cmd-main", 0},
		{"click empty space in another row", geometry.Pt(250, 124), "pkg", 1},
		{"click last row label", geometry.Pt(100, 316), "makefile", 1},
	}
	for _, tc := range cases {
		selections = nil
		clickAt(w, tc.pos, tc.pos)

		if len(selections) != tc.wantFires {
			t.Errorf("%s: got %d OnSelect calls %v, want %d",
				tc.name, len(selections), selections, tc.wantFires)
		}
		if got := selected.Get(); got != tc.wantID {
			t.Errorf("%s: selected node ID = %q, want %q", tc.name, got, tc.wantID)
		}
	}
}

// clickAt dispatches a full click (pointer down + pointer up) through the
// unified pointer pipeline. downPos and upPos are window coordinates.
func clickAt(w *Window, downPos, upPos geometry.Point) {
	w.HandlePointerEvent(&gesture.PointerEvent{
		Base:           event.NewBase(event.TypeMouse, event.ModNone),
		EventType:      gesture.PointerDown,
		PointerID:      1,
		PointerType:    gesture.PointerTypeMouse,
		Position:       downPos,
		GlobalPosition: downPos,
		Button:         event.ButtonLeft,
		Buttons:        event.ButtonStateLeft,
	})
	w.HandlePointerEvent(&gesture.PointerEvent{
		Base:           event.NewBase(event.TypeMouse, event.ModNone),
		EventType:      gesture.PointerUp,
		PointerID:      1,
		PointerType:    gesture.PointerTypeMouse,
		Position:       upPos,
		GlobalPosition: upPos,
		Button:         event.ButtonLeft,
	})
}
