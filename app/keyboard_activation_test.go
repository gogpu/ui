package app

import (
	"testing"

	"github.com/gogpu/ui/core/button"
	"github.com/gogpu/ui/core/checkbox"
	"github.com/gogpu/ui/core/chip"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
)

// TestKeyboardActivation_SchedulesRepaint drives Tab + Enter/Space through
// the window and checks that each key edge that changes what a control looks
// like also tells the window to repaint. Without it the render loop skips the
// frame (desktop.Run's needsFrame is false) and the screen keeps showing the
// old state: a checkbox toggled with Space looks unchanged until something
// else happens to repaint it.
func TestKeyboardActivation_SchedulesRepaint(t *testing.T) {
	tests := []struct {
		name    string
		control func() widget.Widget
		key     event.Key
	}{
		{"button/Enter", func() widget.Widget { return button.New(button.TextOpt("OK")) }, event.KeyEnter},
		{"button/Space", func() widget.Widget { return button.New(button.TextOpt("OK")) }, event.KeySpace},
		{"checkbox/Space", func() widget.Widget { return checkbox.New(checkbox.LabelOpt("Remember me")) }, event.KeySpace},
		{"chip/Enter", func() widget.Widget { return chip.New(chip.Label("Tag"), chip.Selectable(true)) }, event.KeyEnter},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			control := tt.control()
			a := New() // headless: Frame draws and clears the redraw flag
			a.SetRoot(primitives.Box(control).Padding(16))
			w := a.Window()
			w.Frame()

			w.HandleEvent(event.NewKeyEvent(event.KeyPress, event.KeyTab, 0, event.ModNone))
			w.Frame()
			w.ClearDirtyBoundaries() // the render loop does this once it has painted
			if f, ok := control.(interface{ IsFocused() bool }); !ok || !f.IsFocused() {
				t.Fatal("setup: Tab should focus the control")
			}
			if w.NeedsRedraw() || w.HasDirtyBoundaries() {
				t.Fatal("setup: Frame should have consumed the redraw request")
			}

			for _, edge := range []event.KeyEventType{event.KeyPress, event.KeyRelease} {
				w.HandleEvent(event.NewKeyEvent(edge, tt.key, 0, event.ModNone))
				// The same work check desktop.Run makes before painting a frame.
				if !w.NeedsRedraw() && !w.HasDirtyBoundaries() {
					t.Errorf("%v %v changed the control's state but scheduled no repaint", tt.key, edge)
				}
				w.Frame()
				w.ClearDirtyBoundaries()
			}
		})
	}
}
