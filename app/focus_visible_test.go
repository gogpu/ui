package app

import (
	"testing"

	"github.com/gogpu/ui/core/button"
	"github.com/gogpu/ui/core/checkbox"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
)

// TestFocusVisible_PointerVsKeyboard drives the full event chain — mouse
// press/release on a checkbox, then Tab — and checks the :focus-visible
// contract end to end: clicking focuses without the indicator, Tab shows it.
func TestFocusVisible_PointerVsKeyboard(t *testing.T) {
	cb := checkbox.New(checkbox.LabelOpt("Test checkbox"))
	btn := button.New(button.TextOpt("Action"))

	root := primitives.Box(cb, btn).Padding(32).Gap(12)

	wp := &mockWindowProvider{width: 800, height: 600, scale: 1.0}
	a := New(WithWindowProvider(wp))
	a.SetRoot(root)

	w := a.Window()
	w.needsLayout = true
	w.Frame()

	click := func(b interface{ Bounds() geometry.Rect }) {
		pos := b.Bounds().Center()
		press := event.NewMouseEvent(
			event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
			pos, pos, event.ModNone,
		)
		w.HandleEvent(press)
		w.HandleEvent(event.NewMouseEvent(
			event.MouseRelease, event.ButtonLeft, 0,
			pos, pos, event.ModNone,
		))
	}

	// Click the checkbox: it must take focus (keyboard activation works)
	// but not show the focus indicator — the checked state is feedback
	// enough for a pointer interaction.
	click(cb)
	if !cb.IsFocused() {
		t.Fatal("checkbox should be focused after click")
	}
	if cb.IsFocusVisible() {
		t.Error("clicking a checkbox must not show the focus indicator")
	}

	// Click the button: same contract.
	click(btn)
	if !btn.IsFocused() {
		t.Fatal("button should be focused after click")
	}
	if btn.IsFocusVisible() {
		t.Error("clicking a button must not show the focus indicator")
	}

	// Tab away and back: focus arrives via the keyboard and the indicator
	// must appear.
	tab := &event.KeyEvent{Key: event.KeyTab, KeyType: event.KeyPress}
	w.HandleEvent(tab)
	if !cb.IsFocused() {
		t.Fatal("Tab should move focus back to the checkbox")
	}
	if !cb.IsFocusVisible() {
		t.Error("Tab-focused checkbox should show the focus indicator")
	}
	if btn.IsFocusVisible() {
		t.Error("indicator must move with focus")
	}

	// Clicking the keyboard-focused checkbox hides the indicator again,
	// the way browsers re-evaluate :focus-visible on pointerdown.
	click(cb)
	if !cb.IsFocused() {
		t.Fatal("checkbox should stay focused after re-click")
	}
	if cb.IsFocusVisible() {
		t.Error("re-clicking a keyboard-focused widget should hide the indicator")
	}
}

// TestFocusVisible_TabRingOnSingleFocusable covers Tab when the clicked
// widget is the only focusable one: focus does not move, so the window must
// still be told to repaint the ring that Tab turns on.
func TestFocusVisible_TabRingOnSingleFocusable(t *testing.T) {
	btn := button.New(button.TextOpt("Only"))
	a := New(WithWindowProvider(&mockWindowProvider{width: 400, height: 300, scale: 1.0}))
	a.SetRoot(primitives.Box(btn).Padding(20))
	w := a.Window()
	w.needsLayout = true
	w.Frame()

	pos := btn.Bounds().Center()
	w.HandleEvent(event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft, pos, pos, event.ModNone))
	w.HandleEvent(event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0, pos, pos, event.ModNone))
	if !btn.IsFocused() || btn.IsFocusVisible() {
		t.Fatalf("after click: focused=%v visible=%v, want focused without indicator", btn.IsFocused(), btn.IsFocusVisible())
	}
	// Paint the click: the render loop clears the dirty marks once drawn.
	w.Frame()
	w.ClearDirtyBoundaries()
	btn.ClearRedraw()

	w.HandleEvent(&event.KeyEvent{Key: event.KeyTab, KeyType: event.KeyPress})
	if !btn.IsFocusVisible() {
		t.Fatal("Tab should show the indicator on the focused widget")
	}
	if !btn.NeedsRedraw() {
		t.Error("the ring Tab turned on must be repainted, but the widget is not marked dirty")
	}
}

// TestFocusVisible_MouseMoveKeepsKeyboardModality checks that only a press
// counts as pointer interaction: after keyboard navigation, moving the mouse
// over the window must not hide the ring of focus granted next.
func TestFocusVisible_MouseMoveKeepsKeyboardModality(t *testing.T) {
	first := button.New(button.TextOpt("First"))
	second := button.New(button.TextOpt("Second"))
	a := New(WithWindowProvider(&mockWindowProvider{width: 400, height: 300, scale: 1.0}))
	a.SetRoot(primitives.Box(first, second).Padding(20).Gap(8))
	w := a.Window()
	w.needsLayout = true
	w.Frame()

	w.HandleEvent(&event.KeyEvent{Key: event.KeyTab, KeyType: event.KeyPress})
	if !first.IsFocusVisible() {
		t.Fatal("Tab should focus the first button with the indicator")
	}

	pos := second.Bounds().Center()
	w.HandleEvent(event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0, pos, pos, event.ModNone))

	// Focus moved by the application (for example after a timer), not by a
	// click: the last real interaction was the keyboard.
	w.ctx.RequestFocus(second)
	if !second.IsFocusVisible() {
		t.Error("a mouse move must not switch the modality to pointer")
	}
}
