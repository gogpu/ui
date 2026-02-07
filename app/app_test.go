package app

import (
	"testing"

	"github.com/gogpu/gpucontext"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/theme"
	"github.com/gogpu/ui/widget"
)

// --- Mock types for testing ---

// mockWidget implements widget.Widget for testing.
type mockWidget struct {
	widget.WidgetBase
	layoutCalled bool
	drawCalled   bool
	eventCalled  bool
	lastEvent    event.Event
	layoutSize   geometry.Size
}

func newMockWidget() *mockWidget {
	return &mockWidget{
		layoutSize: geometry.Sz(100, 50),
	}
}

func (m *mockWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	m.layoutCalled = true
	return c.Constrain(m.layoutSize)
}

func (m *mockWidget) Draw(_ widget.Context, _ widget.Canvas) {
	m.drawCalled = true
}

func (m *mockWidget) Event(_ widget.Context, e event.Event) bool {
	m.eventCalled = true
	m.lastEvent = e
	return true
}

// mockWindowProvider implements gpucontext.WindowProvider for testing.
type mockWindowProvider struct {
	width, height int
	scale         float64
	redrawCount   int
}

func (m *mockWindowProvider) Size() (int, int) {
	return m.width, m.height
}

func (m *mockWindowProvider) ScaleFactor() float64 {
	return m.scale
}

func (m *mockWindowProvider) RequestRedraw() {
	m.redrawCount++
}

// mockPlatformProvider implements gpucontext.PlatformProvider for testing.
type mockPlatformProvider struct {
	lastCursor gpucontext.CursorShape
	darkMode   bool
	fontScale  float32
}

func (m *mockPlatformProvider) ClipboardRead() (string, error) { return "", nil }
func (m *mockPlatformProvider) ClipboardWrite(string) error    { return nil }
func (m *mockPlatformProvider) SetCursor(c gpucontext.CursorShape) {
	m.lastCursor = c
}
func (m *mockPlatformProvider) DarkMode() bool     { return m.darkMode }
func (m *mockPlatformProvider) ReduceMotion() bool { return false }
func (m *mockPlatformProvider) HighContrast() bool { return false }
func (m *mockPlatformProvider) FontScale() float32 { return m.fontScale }

// mockEventSource implements gpucontext.EventSource for testing.
type mockEventSource struct {
	onKeyPress            func(gpucontext.Key, gpucontext.Modifiers)
	onKeyRelease          func(gpucontext.Key, gpucontext.Modifiers)
	onTextInput           func(string)
	onMouseMove           func(float64, float64)
	onMousePress          func(gpucontext.MouseButton, float64, float64)
	onMouseRelease        func(gpucontext.MouseButton, float64, float64)
	onScroll              func(float64, float64)
	onResize              func(int, int)
	onFocus               func(bool)
	onIMECompositionStart func()
	onIMECompositionEnd   func(string)
}

func (m *mockEventSource) OnKeyPress(fn func(gpucontext.Key, gpucontext.Modifiers)) {
	m.onKeyPress = fn
}
func (m *mockEventSource) OnKeyRelease(fn func(gpucontext.Key, gpucontext.Modifiers)) {
	m.onKeyRelease = fn
}
func (m *mockEventSource) OnTextInput(fn func(string)) {
	m.onTextInput = fn
}
func (m *mockEventSource) OnMouseMove(fn func(float64, float64)) {
	m.onMouseMove = fn
}
func (m *mockEventSource) OnMousePress(fn func(gpucontext.MouseButton, float64, float64)) {
	m.onMousePress = fn
}
func (m *mockEventSource) OnMouseRelease(fn func(gpucontext.MouseButton, float64, float64)) {
	m.onMouseRelease = fn
}
func (m *mockEventSource) OnScroll(fn func(float64, float64)) {
	m.onScroll = fn
}
func (m *mockEventSource) OnResize(fn func(int, int)) {
	m.onResize = fn
}
func (m *mockEventSource) OnFocus(fn func(bool)) {
	m.onFocus = fn
}
func (m *mockEventSource) OnIMECompositionStart(fn func()) {
	m.onIMECompositionStart = fn
}
func (m *mockEventSource) OnIMECompositionUpdate(func(gpucontext.IMEState)) {}
func (m *mockEventSource) OnIMECompositionEnd(fn func(string)) {
	m.onIMECompositionEnd = fn
}

// --- App tests ---

func TestNewApp_Headless(t *testing.T) {
	a := New()

	if a == nil {
		t.Fatal("New() returned nil")
	}
	if a.Window() == nil {
		t.Fatal("Window() returned nil")
	}
	if a.Theme() == nil {
		t.Fatal("Theme() returned nil")
	}
	if a.Scheduler() == nil {
		t.Fatal("Scheduler() returned nil")
	}
}

func TestNewApp_WithWindowProvider(t *testing.T) {
	wp := &mockWindowProvider{width: 1024, height: 768, scale: 2.0}
	a := New(WithWindowProvider(wp))

	if a.Window().WindowSize().Width != 1024 {
		t.Errorf("window width = %v, want 1024", a.Window().WindowSize().Width)
	}
	if a.Window().WindowSize().Height != 768 {
		t.Errorf("window height = %v, want 768", a.Window().WindowSize().Height)
	}
	if a.Window().Context().Scale() != 2.0 {
		t.Errorf("scale = %v, want 2.0", a.Window().Context().Scale())
	}
}

func TestNewApp_WithPlatformProvider(t *testing.T) {
	pp := &mockPlatformProvider{darkMode: true, fontScale: 1.5}
	a := New(WithPlatformProvider(pp))

	if a.Window() == nil {
		t.Fatal("Window() returned nil")
	}
	// Platform provider is accessible through the window.
	// Verify it was stored by triggering cursor sync.
	a.SetRoot(newMockWidget())
	a.Frame()

	if pp.lastCursor != gpucontext.CursorDefault {
		t.Errorf("cursor = %v, want Default", pp.lastCursor)
	}
}

func TestNewApp_WithTheme(t *testing.T) {
	dark := theme.DefaultDark()
	a := New(WithTheme(dark))

	if a.Theme() != dark {
		t.Error("theme was not set correctly")
	}
	if !a.Theme().IsDark() {
		t.Error("expected dark theme")
	}
}

func TestNewApp_WithEventSource(t *testing.T) {
	es := &mockEventSource{}
	a := New(WithEventSource(es))

	if a == nil {
		t.Fatal("New() returned nil")
	}
	// Verify callbacks were registered.
	if es.onMouseMove == nil {
		t.Error("OnMouseMove was not registered")
	}
	if es.onMousePress == nil {
		t.Error("OnMousePress was not registered")
	}
	if es.onKeyPress == nil {
		t.Error("OnKeyPress was not registered")
	}
	if es.onResize == nil {
		t.Error("OnResize was not registered")
	}
	if es.onFocus == nil {
		t.Error("OnFocus was not registered")
	}
}

func TestApp_SetRoot(t *testing.T) {
	a := New()
	w := newMockWidget()
	a.SetRoot(w)

	if a.Window().Root() != w {
		t.Error("root widget not set correctly")
	}
}

func TestApp_SetTheme(t *testing.T) {
	a := New()
	dark := theme.DefaultDark()
	a.SetTheme(dark)

	if a.Theme() != dark {
		t.Error("theme was not updated")
	}
	if a.Window().Theme() != dark {
		t.Error("window theme was not updated")
	}
}

func TestApp_SetTheme_Nil(t *testing.T) {
	a := New()
	original := a.Theme()
	a.SetTheme(nil)

	if a.Theme() != original {
		t.Error("nil theme should be ignored")
	}
}

func TestApp_Frame_NoRoot(t *testing.T) {
	a := New()
	// Should not panic.
	a.Frame()
}

func TestApp_Frame_WithRoot(t *testing.T) {
	a := New()
	w := newMockWidget()
	a.SetRoot(w)
	a.Frame()

	if !w.layoutCalled {
		t.Error("layout was not called on root widget")
	}
}

func TestApp_HandleEvent(t *testing.T) {
	a := New()
	w := newMockWidget()
	a.SetRoot(w)

	e := event.NewKeyEvent(event.KeyPress, event.KeyA, 'a', event.ModNone)
	a.HandleEvent(e)

	if !w.eventCalled {
		t.Error("event was not dispatched to root widget")
	}
	if w.lastEvent != e {
		t.Error("wrong event dispatched")
	}
}

func TestApp_HandleEvent_NoRoot(t *testing.T) {
	a := New()
	e := event.NewKeyEvent(event.KeyPress, event.KeyA, 'a', event.ModNone)
	// Should not panic.
	a.HandleEvent(e)
}

func TestApp_SetFrameCallback(t *testing.T) {
	a := New()
	w := newMockWidget()
	a.SetRoot(w)

	var stats FrameStats
	a.SetFrameCallback(func(s FrameStats) {
		stats = s
	})

	a.Frame()

	if stats.FrameStart.IsZero() {
		t.Error("frame start was not set")
	}
	if stats.TotalDuration < 0 {
		t.Error("total duration was negative")
	}
	if !stats.LayoutPerformed {
		t.Error("layout should have been performed on first frame")
	}
}

func TestApp_SetFrameCallback_SecondFrame(t *testing.T) {
	a := New()
	w := newMockWidget()
	a.SetRoot(w)

	var callCount int
	a.SetFrameCallback(func(_ FrameStats) {
		callCount++
	})

	a.Frame()
	a.Frame()

	if callCount != 2 {
		t.Errorf("callback called %d times, want 2", callCount)
	}
}

func TestApp_AllOptions(t *testing.T) {
	wp := &mockWindowProvider{width: 640, height: 480, scale: 1.5}
	pp := &mockPlatformProvider{fontScale: 1.0}
	es := &mockEventSource{}
	th := theme.DefaultDark()

	a := New(
		WithWindowProvider(wp),
		WithPlatformProvider(pp),
		WithEventSource(es),
		WithTheme(th),
	)

	if a.Window().WindowSize().Width != 640 {
		t.Errorf("width = %v, want 640", a.Window().WindowSize().Width)
	}
	if a.Theme() != th {
		t.Error("theme not set")
	}
}
