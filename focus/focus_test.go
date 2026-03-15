package focus_test

import (
	"image"
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/focus"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// mockWidget is a test widget that implements Widget and optionally Focusable.
type mockWidget struct {
	widget.WidgetBase
	focusable bool
	children  []widget.Widget
}

func newMockWidget(id string, focusable bool) *mockWidget {
	w := &mockWidget{focusable: focusable}
	w.SetID(id)
	w.SetVisible(true)
	w.SetEnabled(true)
	return w
}

func (w *mockWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Smallest()
}

func (w *mockWidget) Draw(_ widget.Context, _ widget.Canvas) {}

func (w *mockWidget) Event(_ widget.Context, _ event.Event) bool {
	return false
}

func (w *mockWidget) Children() []widget.Widget {
	return w.children
}

func (w *mockWidget) IsFocusable() bool {
	return w.focusable && w.IsVisible() && w.IsEnabled()
}

func (w *mockWidget) addChild(child *mockWidget) {
	w.children = append(w.children, child)
}

// mockCanvas records calls for testing DrawFocusRing.
type mockCanvas struct {
	strokeRoundRects []strokeRoundRectCall
}

type strokeRoundRectCall struct {
	r           geometry.Rect
	color       widget.Color
	radius      float32
	strokeWidth float32
}

func (c *mockCanvas) Clear(_ widget.Color)                                  {}
func (c *mockCanvas) DrawRect(_ geometry.Rect, _ widget.Color)              {}
func (c *mockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) {}
func (c *mockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32) {
}

func (c *mockCanvas) StrokeRoundRect(r geometry.Rect, color widget.Color, radius float32, strokeWidth float32) {
	c.strokeRoundRects = append(c.strokeRoundRects, strokeRoundRectCall{
		r:           r,
		color:       color,
		radius:      radius,
		strokeWidth: strokeWidth,
	})
}

func (c *mockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)              {}
func (c *mockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {}
func (c *mockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32)             {}

func (c *mockCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
}

func (c *mockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}

func (c *mockCanvas) DrawImage(_ image.Image, _ geometry.Point)    {}
func (c *mockCanvas) PushClip(_ geometry.Rect)                     {}
func (c *mockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *mockCanvas) PopClip()                                     {}
func (c *mockCanvas) PushTransform(_ geometry.Point)               {}
func (c *mockCanvas) PopTransform()                                {}
func (c *mockCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }

// testTree holds widget references from buildTree for easy access.
type testTree struct {
	root *mockWidget
	a    *mockWidget // focusable
	b    *mockWidget // not focusable container
	c    *mockWidget // focusable, child of b
	d    *mockWidget // focusable
}

// buildTree creates a widget tree for testing:
//
//	root
//	 +-- a (focusable)
//	 +-- b (not focusable)
//	 |   +-- c (focusable)
//	 +-- d (focusable)
func buildTree() testTree {
	root := newMockWidget("root", false)
	a := newMockWidget("a", true)
	b := newMockWidget("b", false)
	c := newMockWidget("c", true)
	d := newMockWidget("d", true)

	b.addChild(c)
	root.addChild(a)
	root.addChild(b)
	root.addChild(d)

	return testTree{root: root, a: a, b: b, c: c, d: d}
}

// --- Manager Tests ---

func TestNew(t *testing.T) {
	root := newMockWidget("root", false)
	m := focus.New(root)

	if m.Focused() != nil {
		t.Error("new manager should have no focused widget")
	}
}

func TestNew_NilRoot(t *testing.T) {
	m := focus.New(nil)
	if m.Focused() != nil {
		t.Error("new manager with nil root should have no focused widget")
	}
	m.Next()
	m.Previous()
	m.Blur()
}

func TestFocus(t *testing.T) {
	tt := buildTree()
	m := focus.New(tt.root)

	m.Focus(tt.a)

	if m.Focused() != tt.a {
		t.Error("focused should be widget a")
	}
	if !tt.a.IsFocused() {
		t.Error("widget a should report focused")
	}
}

func TestFocus_BlursPrevious(t *testing.T) {
	tt := buildTree()
	m := focus.New(tt.root)

	m.Focus(tt.a)
	m.Focus(tt.c)

	if tt.a.IsFocused() {
		t.Error("widget a should have lost focus")
	}
	if !tt.c.IsFocused() {
		t.Error("widget c should have focus")
	}
	if m.Focused() != tt.c {
		t.Error("focused should be widget c")
	}
}

func TestFocus_SameWidget(t *testing.T) {
	tt := buildTree()
	m := focus.New(tt.root)

	m.Focus(tt.a)
	m.Focus(tt.a) // no-op

	if !tt.a.IsFocused() {
		t.Error("widget a should still be focused")
	}
}

func TestFocus_Nil(t *testing.T) {
	tt := buildTree()
	m := focus.New(tt.root)

	m.Focus(tt.a)
	m.Focus(nil) // equivalent to Blur

	if tt.a.IsFocused() {
		t.Error("widget a should have lost focus")
	}
	if m.Focused() != nil {
		t.Error("focused should be nil")
	}
}

func TestFocus_NonFocusable(t *testing.T) {
	tt := buildTree()
	m := focus.New(tt.root)

	m.Focus(tt.a)

	nf := newMockWidget("nf", false)
	m.Focus(nf) // should be no-op

	if m.Focused() != tt.a {
		t.Error("focused should still be widget a")
	}
}

func TestBlur(t *testing.T) {
	tt := buildTree()
	m := focus.New(tt.root)

	m.Focus(tt.a)
	m.Blur()

	if m.Focused() != nil {
		t.Error("focused should be nil after blur")
	}
	if tt.a.IsFocused() {
		t.Error("widget a should not be focused after blur")
	}
}

func TestBlur_WhenNothingFocused(t *testing.T) {
	root := newMockWidget("root", false)
	m := focus.New(root)

	m.Blur()

	if m.Focused() != nil {
		t.Error("focused should be nil")
	}
}

// --- Tab Navigation Tests ---

func TestNext(t *testing.T) {
	tt := buildTree()
	m := focus.New(tt.root)

	m.Next()
	if m.Focused() != tt.a {
		t.Errorf("first Next: focused = %v, want a", focusedID(m))
	}

	m.Next()
	if m.Focused() != tt.c {
		t.Errorf("second Next: focused = %v, want c", focusedID(m))
	}

	m.Next()
	if m.Focused() != tt.d {
		t.Errorf("third Next: focused = %v, want d", focusedID(m))
	}

	m.Next()
	if m.Focused() != tt.a {
		t.Errorf("fourth Next: focused = %v, want a (wrap)", focusedID(m))
	}
}

func TestPrevious(t *testing.T) {
	tt := buildTree()
	m := focus.New(tt.root)

	m.Previous()
	if m.Focused() != tt.d {
		t.Errorf("first Previous: focused = %v, want d", focusedID(m))
	}

	m.Previous()
	if m.Focused() != tt.c {
		t.Errorf("second Previous: focused = %v, want c", focusedID(m))
	}

	m.Previous()
	if m.Focused() != tt.a {
		t.Errorf("third Previous: focused = %v, want a", focusedID(m))
	}

	m.Previous()
	if m.Focused() != tt.d {
		t.Errorf("fourth Previous: focused = %v, want d (wrap)", focusedID(m))
	}
}

func TestNext_EmptyTree(t *testing.T) {
	root := newMockWidget("root", false)
	m := focus.New(root)

	m.Next()

	if m.Focused() != nil {
		t.Error("Next on empty tree should not set focus")
	}
}

func TestPrevious_EmptyTree(t *testing.T) {
	root := newMockWidget("root", false)
	m := focus.New(root)

	m.Previous()

	if m.Focused() != nil {
		t.Error("Previous on empty tree should not set focus")
	}
}

func TestNext_SingleWidget(t *testing.T) {
	root := newMockWidget("root", false)
	only := newMockWidget("only", true)
	root.addChild(only)
	m := focus.New(root)

	m.Next()
	if m.Focused() != only {
		t.Error("first Next should focus only widget")
	}

	m.Next()
	if m.Focused() != only {
		t.Error("second Next should still focus only widget")
	}
}

func TestPrevious_SingleWidget(t *testing.T) {
	root := newMockWidget("root", false)
	only := newMockWidget("only", true)
	root.addChild(only)
	m := focus.New(root)

	m.Previous()
	if m.Focused() != only {
		t.Error("first Previous should focus only widget")
	}

	m.Previous()
	if m.Focused() != only {
		t.Error("second Previous should still focus only widget")
	}
}

func TestNext_SkipsInvisible(t *testing.T) {
	root := newMockWidget("root", false)
	a := newMockWidget("a", true)
	b := newMockWidget("b", true)
	b.SetVisible(false)
	c := newMockWidget("c", true)

	root.addChild(a)
	root.addChild(b)
	root.addChild(c)
	m := focus.New(root)

	m.Next()
	m.Next()

	if m.Focused() != c {
		t.Errorf("Next should skip invisible widget: focused = %v, want c", focusedID(m))
	}
}

func TestNext_SkipsDisabled(t *testing.T) {
	root := newMockWidget("root", false)
	a := newMockWidget("a", true)
	b := newMockWidget("b", true)
	b.SetEnabled(false)
	c := newMockWidget("c", true)

	root.addChild(a)
	root.addChild(b)
	root.addChild(c)
	m := focus.New(root)

	m.Next()
	m.Next()

	if m.Focused() != c {
		t.Errorf("Next should skip disabled widget: focused = %v, want c", focusedID(m))
	}
}

func TestNext_SkipsInvisibleSubtree(t *testing.T) {
	root := newMockWidget("root", false)
	a := newMockWidget("a", true)
	container := newMockWidget("container", false)
	container.SetVisible(false)
	child := newMockWidget("child", true)
	container.addChild(child)
	b := newMockWidget("b", true)

	root.addChild(a)
	root.addChild(container)
	root.addChild(b)
	m := focus.New(root)

	m.Next()
	m.Next()

	if m.Focused() != b {
		t.Errorf("Next should skip invisible subtree: focused = %v, want b", focusedID(m))
	}
}

func TestNext_FocusedRemovedFromTree(t *testing.T) {
	root := newMockWidget("root", false)
	a := newMockWidget("a", true)
	b := newMockWidget("b", true)
	root.addChild(a)
	root.addChild(b)
	m := focus.New(root)

	m.Focus(a)

	root.children = []widget.Widget{b}

	m.Next()
	if m.Focused() != b {
		t.Errorf("Next with removed widget: focused = %v, want b", focusedID(m))
	}
}

// --- HandleKeyEvent Tests ---

func TestHandleKeyEvent_Tab(t *testing.T) {
	tt := buildTree()
	m := focus.New(tt.root)

	tabPress := event.NewKeyEvent(event.KeyPress, event.KeyTab, 0, event.ModNone)
	consumed := m.HandleKeyEvent(tabPress)

	if !consumed {
		t.Error("Tab press should be consumed")
	}
	if m.Focused() != tt.a {
		t.Errorf("Tab should focus first widget: focused = %v, want a", focusedID(m))
	}

	tabPress2 := event.NewKeyEvent(event.KeyPress, event.KeyTab, 0, event.ModNone)
	m.HandleKeyEvent(tabPress2)

	if m.Focused() != tt.c {
		t.Errorf("second Tab should focus c: focused = %v, want c", focusedID(m))
	}
}

func TestHandleKeyEvent_ShiftTab(t *testing.T) {
	tt := buildTree()
	m := focus.New(tt.root)

	shiftTab := event.NewKeyEvent(event.KeyPress, event.KeyTab, 0, event.ModShift)
	consumed := m.HandleKeyEvent(shiftTab)

	if !consumed {
		t.Error("Shift+Tab press should be consumed")
	}
	if m.Focused() != tt.d {
		t.Errorf("Shift+Tab should focus last widget: focused = %v, want d", focusedID(m))
	}
}

func TestHandleKeyEvent_TabRepeat(t *testing.T) {
	tt := buildTree()
	m := focus.New(tt.root)

	tabPress := event.NewKeyEvent(event.KeyPress, event.KeyTab, 0, event.ModNone)
	m.HandleKeyEvent(tabPress)

	tabRepeat := event.NewKeyEvent(event.KeyRepeat, event.KeyTab, 0, event.ModNone)
	consumed := m.HandleKeyEvent(tabRepeat)

	if !consumed {
		t.Error("Tab repeat should be consumed")
	}
	if m.Focused() != tt.c {
		t.Errorf("Tab repeat should advance focus: focused = %v, want c", focusedID(m))
	}
}

func TestHandleKeyEvent_TabRelease(t *testing.T) {
	root := newMockWidget("root", false)
	m := focus.New(root)

	tabRelease := event.NewKeyEvent(event.KeyRelease, event.KeyTab, 0, event.ModNone)
	consumed := m.HandleKeyEvent(tabRelease)

	if !consumed {
		t.Error("Tab release should be consumed to prevent propagation")
	}
}

func TestHandleKeyEvent_NonTab(t *testing.T) {
	root := newMockWidget("root", false)
	m := focus.New(root)

	enterPress := event.NewKeyEvent(event.KeyPress, event.KeyEnter, 0, event.ModNone)
	consumed := m.HandleKeyEvent(enterPress)

	if consumed {
		t.Error("Enter press should not be consumed")
	}
}

func TestHandleKeyEvent_Nil(t *testing.T) {
	root := newMockWidget("root", false)
	m := focus.New(root)

	consumed := m.HandleKeyEvent(nil)
	if consumed {
		t.Error("nil event should not be consumed")
	}
}

// --- Shortcut Tests ---

func TestShortcut_Register(t *testing.T) {
	root := newMockWidget("root", false)
	m := focus.New(root)

	called := false
	s := focus.Shortcut{Key: event.KeyS, Ctrl: true}
	m.RegisterShortcut(s, func() {
		called = true
	})

	ctrlS := event.NewKeyEvent(event.KeyPress, event.KeyS, 0, event.ModCtrl)
	consumed := m.HandleKeyEvent(ctrlS)

	if !consumed {
		t.Error("Ctrl+S should be consumed")
	}
	if !called {
		t.Error("shortcut handler should have been called")
	}
}

func TestShortcut_NoMatch(t *testing.T) {
	root := newMockWidget("root", false)
	m := focus.New(root)

	called := false
	s := focus.Shortcut{Key: event.KeyS, Ctrl: true}
	m.RegisterShortcut(s, func() {
		called = true
	})

	justS := event.NewKeyEvent(event.KeyPress, event.KeyS, 0, event.ModNone)
	consumed := m.HandleKeyEvent(justS)

	if consumed {
		t.Error("S without Ctrl should not be consumed")
	}
	if called {
		t.Error("handler should not be called without Ctrl")
	}
}

func TestShortcut_WithShiftAndAlt(t *testing.T) {
	root := newMockWidget("root", false)
	m := focus.New(root)

	called := false
	s := focus.Shortcut{Key: event.KeyZ, Ctrl: true, Shift: true}
	m.RegisterShortcut(s, func() {
		called = true
	})

	ctrlShiftZ := event.NewKeyEvent(event.KeyPress, event.KeyZ, 0, event.ModCtrl|event.ModShift)
	consumed := m.HandleKeyEvent(ctrlShiftZ)

	if !consumed {
		t.Error("Ctrl+Shift+Z should be consumed")
	}
	if !called {
		t.Error("handler should be called for Ctrl+Shift+Z")
	}
}

func TestShortcut_Unregister(t *testing.T) {
	root := newMockWidget("root", false)
	m := focus.New(root)

	called := false
	s := focus.Shortcut{Key: event.KeyS, Ctrl: true}
	m.RegisterShortcut(s, func() {
		called = true
	})

	m.UnregisterShortcut(s)

	ctrlS := event.NewKeyEvent(event.KeyPress, event.KeyS, 0, event.ModCtrl)
	consumed := m.HandleKeyEvent(ctrlS)

	if consumed {
		t.Error("Ctrl+S should not be consumed after unregister")
	}
	if called {
		t.Error("handler should not be called after unregister")
	}
}

func TestShortcut_Unregister_NotFound(t *testing.T) {
	root := newMockWidget("root", false)
	m := focus.New(root)

	s := focus.Shortcut{Key: event.KeyQ, Ctrl: true}
	m.UnregisterShortcut(s)
}

func TestShortcut_NilHandler(t *testing.T) {
	root := newMockWidget("root", false)
	m := focus.New(root)

	m.RegisterShortcut(focus.Shortcut{Key: event.KeyA}, nil)

	keyA := event.NewKeyEvent(event.KeyPress, event.KeyA, 0, event.ModNone)
	consumed := m.HandleKeyEvent(keyA)

	if consumed {
		t.Error("nil handler shortcut should not have been registered")
	}
}

func TestShortcut_PrecedenceOverTab(t *testing.T) {
	root := newMockWidget("root", false)
	a := newMockWidget("a", true)
	root.addChild(a)
	m := focus.New(root)

	shortcutCalled := false
	s := focus.Shortcut{Key: event.KeyTab, Ctrl: true}
	m.RegisterShortcut(s, func() {
		shortcutCalled = true
	})

	ctrlTab := event.NewKeyEvent(event.KeyPress, event.KeyTab, 0, event.ModCtrl)
	consumed := m.HandleKeyEvent(ctrlTab)

	if !consumed {
		t.Error("Ctrl+Tab should be consumed by shortcut")
	}
	if !shortcutCalled {
		t.Error("shortcut handler should have been called")
	}
	if m.Focused() != nil {
		t.Error("tab navigation should not have occurred")
	}
}

func TestShortcut_OnlyOnPress(t *testing.T) {
	root := newMockWidget("root", false)
	m := focus.New(root)

	called := false
	s := focus.Shortcut{Key: event.KeyS, Ctrl: true}
	m.RegisterShortcut(s, func() {
		called = true
	})

	ctrlSRepeat := event.NewKeyEvent(event.KeyRepeat, event.KeyS, 0, event.ModCtrl)
	m.HandleKeyEvent(ctrlSRepeat)

	if called {
		t.Error("shortcuts should only fire on press, not repeat")
	}

	ctrlSRelease := event.NewKeyEvent(event.KeyRelease, event.KeyS, 0, event.ModCtrl)
	m.HandleKeyEvent(ctrlSRelease)

	if called {
		t.Error("shortcuts should only fire on press, not release")
	}
}

// --- Shortcut.Matches Tests ---

func TestShortcut_Matches(t *testing.T) {
	tests := []struct {
		name     string
		shortcut focus.Shortcut
		key      event.Key
		mods     event.Modifiers
		want     bool
	}{
		{
			name:     "exact match",
			shortcut: focus.Shortcut{Key: event.KeyS, Ctrl: true},
			key:      event.KeyS,
			mods:     event.ModCtrl,
			want:     true,
		},
		{
			name:     "wrong key",
			shortcut: focus.Shortcut{Key: event.KeyS, Ctrl: true},
			key:      event.KeyA,
			mods:     event.ModCtrl,
			want:     false,
		},
		{
			name:     "missing ctrl",
			shortcut: focus.Shortcut{Key: event.KeyS, Ctrl: true},
			key:      event.KeyS,
			mods:     event.ModNone,
			want:     false,
		},
		{
			name:     "extra modifier present",
			shortcut: focus.Shortcut{Key: event.KeyS, Ctrl: true},
			key:      event.KeyS,
			mods:     event.ModCtrl | event.ModShift,
			want:     false,
		},
		{
			name:     "alt shortcut",
			shortcut: focus.Shortcut{Key: event.KeyF4, Alt: true},
			key:      event.KeyF4,
			mods:     event.ModAlt,
			want:     true,
		},
		{
			name:     "no modifiers",
			shortcut: focus.Shortcut{Key: event.KeyEscape},
			key:      event.KeyEscape,
			mods:     event.ModNone,
			want:     true,
		},
		{
			name:     "no modifiers but ctrl held",
			shortcut: focus.Shortcut{Key: event.KeyEscape},
			key:      event.KeyEscape,
			mods:     event.ModCtrl,
			want:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := event.NewKeyEvent(event.KeyPress, tc.key, 0, tc.mods)
			got := tc.shortcut.Matches(e)
			if got != tc.want {
				t.Errorf("Matches() = %v, want %v", got, tc.want)
			}
		})
	}
}

// --- SetRoot Tests ---

func TestSetRoot(t *testing.T) {
	root1 := newMockWidget("root1", false)
	a := newMockWidget("a", true)
	root1.addChild(a)

	root2 := newMockWidget("root2", false)
	b := newMockWidget("b", true)
	root2.addChild(b)

	m := focus.New(root1)
	m.Focus(a)

	m.SetRoot(root2)

	if m.Focused() != nil {
		t.Error("focus should be cleared when focused widget is not in new tree")
	}
	if a.IsFocused() {
		t.Error("widget a should be blurred")
	}
}

func TestSetRoot_FocusedStillPresent(t *testing.T) {
	root := newMockWidget("root", false)
	a := newMockWidget("a", true)
	b := newMockWidget("b", true)
	root.addChild(a)
	root.addChild(b)

	m := focus.New(root)
	m.Focus(a)

	m.SetRoot(root)

	if m.Focused() != a {
		t.Error("focus should be preserved when widget is still in tree")
	}
}

func TestSetRoot_Nil(t *testing.T) {
	root := newMockWidget("root", false)
	a := newMockWidget("a", true)
	root.addChild(a)

	m := focus.New(root)
	m.Focus(a)

	m.SetRoot(nil)

	if m.Focused() != nil {
		t.Error("focus should be cleared with nil root")
	}
}

// --- DrawFocusRing Tests ---

func TestDrawFocusRing(t *testing.T) {
	canvas := &mockCanvas{}
	bounds := geometry.NewRect(10, 20, 100, 50)
	color := widget.ColorBlue

	focus.DrawFocusRing(canvas, bounds, color, 4.0)

	if len(canvas.strokeRoundRects) != 1 {
		t.Fatalf("expected 1 StrokeRoundRect call, got %d", len(canvas.strokeRoundRects))
	}

	call := canvas.strokeRoundRects[0]

	// Ring should be expanded by 2.0 (default offset).
	expectedBounds := bounds.Expand(2.0)
	if call.r != expectedBounds {
		t.Errorf("ring bounds = %v, want %v", call.r, expectedBounds)
	}

	if call.color != color {
		t.Errorf("ring color = %v, want %v", call.color, color)
	}

	// radius = 4.0 + 2.0 (default offset) = 6.0
	var expectedRadius float32 = 6.0
	if call.radius != expectedRadius {
		t.Errorf("ring radius = %v, want %v", call.radius, expectedRadius)
	}

	var expectedStrokeWidth float32 = 2.0
	if call.strokeWidth != expectedStrokeWidth {
		t.Errorf("ring strokeWidth = %v, want %v", call.strokeWidth, expectedStrokeWidth)
	}
}

func TestDrawFocusRing_ZeroRadius(t *testing.T) {
	canvas := &mockCanvas{}
	bounds := geometry.NewRect(0, 0, 50, 50)

	focus.DrawFocusRing(canvas, bounds, widget.ColorRed, 0)

	if len(canvas.strokeRoundRects) != 1 {
		t.Fatalf("expected 1 call, got %d", len(canvas.strokeRoundRects))
	}

	call := canvas.strokeRoundRects[0]
	var expectedRadius float32 = 2.0
	if call.radius != expectedRadius {
		t.Errorf("radius with 0 input = %v, want %v", call.radius, expectedRadius)
	}
}

// --- Helper ---

func focusedID(m *focus.Manager) string {
	f := m.Focused()
	if f == nil {
		return "<nil>"
	}
	if w, ok := f.(*mockWidget); ok {
		return w.ID()
	}
	return "<unknown>"
}
