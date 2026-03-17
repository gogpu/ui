package widget

import (
	"sync"
	"testing"
	"time"

	"github.com/gogpu/ui/geometry"
)

func TestCursorType_String(t *testing.T) {
	tests := []struct {
		cursor CursorType
		want   string
	}{
		{CursorDefault, "Default"},
		{CursorPointer, "Pointer"},
		{CursorText, "Text"},
		{CursorCrosshair, "Crosshair"},
		{CursorMove, "Move"},
		{CursorResizeNS, "ResizeNS"},
		{CursorResizeEW, "ResizeEW"},
		{CursorResizeNESW, "ResizeNESW"},
		{CursorResizeNWSE, "ResizeNWSE"},
		{CursorNotAllowed, "NotAllowed"},
		{CursorWait, "Wait"},
		{CursorNone, "None"},
		{CursorType(255), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.cursor.String()
			if got != tt.want {
				t.Errorf("CursorType(%d).String() = %q, want %q", tt.cursor, got, tt.want)
			}
		})
	}
}

func TestNewContext(t *testing.T) {
	ctx := NewContext()
	if ctx == nil {
		t.Fatal("NewContext returned nil")
	}

	// Check defaults
	if ctx.FocusedWidget() != nil {
		t.Error("expected no focused widget by default")
	}
	if ctx.Scale() != 1.0 {
		t.Errorf("Scale() = %v, want 1.0", ctx.Scale())
	}
	if ctx.Cursor() != CursorDefault {
		t.Errorf("Cursor() = %v, want CursorDefault", ctx.Cursor())
	}
	if ctx.IsInvalidated() {
		t.Error("expected not invalidated initially")
	}
	if ctx.DeltaTime() != 0 {
		t.Errorf("DeltaTime() = %v, want 0", ctx.DeltaTime())
	}
	if ctx.ThemeProvider() != nil {
		t.Error("expected nil ThemeProvider by default")
	}
}

func TestContextImpl_Focus(t *testing.T) {
	ctx := NewContext()
	widget1 := newMockWidget()
	widget2 := newMockWidget()

	// Initially no focus
	if ctx.FocusedWidget() != nil {
		t.Error("expected no focused widget initially")
	}
	if ctx.IsFocused(widget1) {
		t.Error("widget1 should not be focused")
	}

	// Request focus for widget1
	ctx.RequestFocus(widget1)
	if ctx.FocusedWidget() != widget1 {
		t.Error("widget1 should be focused")
	}
	if !ctx.IsFocused(widget1) {
		t.Error("IsFocused(widget1) should be true")
	}
	if widget1.IsFocused() != true {
		t.Error("widget1.IsFocused() should be true")
	}

	// Request focus for widget2 (should unfocus widget1)
	ctx.RequestFocus(widget2)
	if ctx.FocusedWidget() != widget2 {
		t.Error("widget2 should be focused")
	}
	if ctx.IsFocused(widget1) {
		t.Error("widget1 should not be focused")
	}
	if widget1.IsFocused() {
		t.Error("widget1.IsFocused() should be false after losing focus")
	}
	if !ctx.IsFocused(widget2) {
		t.Error("IsFocused(widget2) should be true")
	}

	// Request focus for already focused widget (no-op)
	ctx.RequestFocus(widget2)
	if ctx.FocusedWidget() != widget2 {
		t.Error("widget2 should still be focused")
	}

	// Release focus
	ctx.ReleaseFocus(widget2)
	if ctx.FocusedWidget() != nil {
		t.Error("no widget should be focused after release")
	}
	if widget2.IsFocused() {
		t.Error("widget2.IsFocused() should be false after release")
	}

	// Release focus from wrong widget (no-op)
	ctx.RequestFocus(widget1)
	ctx.ReleaseFocus(widget2) // widget2 doesn't have focus
	if ctx.FocusedWidget() != widget1 {
		t.Error("widget1 should still be focused")
	}
}

func TestContextImpl_RequestFocus_Nil(t *testing.T) {
	ctx := NewContext()
	widget1 := newMockWidget()

	ctx.RequestFocus(widget1)
	ctx.RequestFocus(nil) // Should clear focus

	if ctx.FocusedWidget() != nil {
		t.Error("focusing nil should clear focus")
	}
	if widget1.IsFocused() {
		t.Error("widget1 should lose focus when nil is focused")
	}
}

func TestContextImpl_Time(t *testing.T) {
	ctx := NewContext()

	// Initial time should be set
	initialNow := ctx.Now()
	if initialNow.IsZero() {
		t.Error("initial Now() should not be zero")
	}

	// Update time
	time.Sleep(10 * time.Millisecond)
	newTime := time.Now()
	ctx.BeginFrame(newTime)

	// Check new time
	if ctx.Now() != newTime {
		t.Error("Now() should return the set time")
	}

	// Check delta time
	delta := ctx.DeltaTime()
	if delta < 10*time.Millisecond {
		t.Errorf("DeltaTime() = %v, expected >= 10ms", delta)
	}
}

func TestContextImpl_Invalidate(t *testing.T) {
	ctx := NewContext()

	// Initially not invalidated
	if ctx.IsInvalidated() {
		t.Error("expected not invalidated initially")
	}

	// Invalidate
	ctx.Invalidate()
	if !ctx.IsInvalidated() {
		t.Error("expected invalidated after Invalidate()")
	}

	// Clear invalidation
	ctx.ClearInvalidation()
	if ctx.IsInvalidated() {
		t.Error("expected not invalidated after ClearInvalidation()")
	}
}

func TestContextImpl_InvalidateCallback(t *testing.T) {
	ctx := NewContext()
	called := false
	ctx.SetOnInvalidate(func() {
		called = true
	})

	ctx.Invalidate()
	if !called {
		t.Error("invalidate callback should have been called")
	}
}

func TestContextImpl_InvalidateRect(t *testing.T) {
	ctx := NewContext()

	// Initially empty rect
	if !ctx.InvalidatedRect().IsEmpty() {
		t.Error("expected empty invalidated rect initially")
	}

	// Invalidate a rect
	r1 := geometry.NewRect(0, 0, 100, 100)
	ctx.InvalidateRect(r1)
	got := ctx.InvalidatedRect()
	if got != r1 {
		t.Errorf("InvalidatedRect() = %v, want %v", got, r1)
	}

	// Invalidate another rect (should union)
	r2 := geometry.NewRect(50, 50, 100, 100)
	ctx.InvalidateRect(r2)
	got = ctx.InvalidatedRect()
	expected := r1.Union(r2)
	if got != expected {
		t.Errorf("InvalidatedRect() = %v, want %v", got, expected)
	}

	// Clear
	ctx.ClearInvalidation()
	if !ctx.InvalidatedRect().IsEmpty() {
		t.Error("expected empty invalidated rect after clear")
	}
}

func TestContextImpl_InvalidateRect_WhenFullInvalidated(t *testing.T) {
	ctx := NewContext()

	// Full invalidation first
	ctx.Invalidate()

	// InvalidateRect should be no-op when already fully invalidated
	rectCalled := false
	ctx.SetOnInvalidateRect(func(_ geometry.Rect) {
		rectCalled = true
	})

	ctx.InvalidateRect(geometry.NewRect(0, 0, 100, 100))
	if rectCalled {
		t.Error("InvalidateRect callback should not be called when already fully invalidated")
	}
}

func TestContextImpl_InvalidateRectCallback(t *testing.T) {
	ctx := NewContext()
	var calledRect geometry.Rect
	ctx.SetOnInvalidateRect(func(r geometry.Rect) {
		calledRect = r
	})

	r := geometry.NewRect(10, 20, 30, 40)
	ctx.InvalidateRect(r)
	if calledRect != r {
		t.Errorf("callback received %v, want %v", calledRect, r)
	}
}

func TestContextImpl_Cursor(t *testing.T) {
	ctx := NewContext()

	// Default cursor
	if ctx.Cursor() != CursorDefault {
		t.Error("expected default cursor initially")
	}

	// Change cursor
	ctx.SetCursor(CursorPointer)
	if ctx.Cursor() != CursorPointer {
		t.Error("expected pointer cursor after SetCursor")
	}

	// Reset cursor
	ctx.ResetCursor()
	if ctx.Cursor() != CursorDefault {
		t.Error("expected default cursor after reset")
	}
}

func TestContextImpl_Scale(t *testing.T) {
	ctx := NewContext()

	// Default scale
	if ctx.Scale() != 1.0 {
		t.Errorf("Scale() = %v, want 1.0", ctx.Scale())
	}

	// Change scale
	ctx.SetScale(2.0)
	if ctx.Scale() != 2.0 {
		t.Errorf("Scale() = %v, want 2.0", ctx.Scale())
	}

	ctx.SetScale(1.5)
	if ctx.Scale() != 1.5 {
		t.Errorf("Scale() = %v, want 1.5", ctx.Scale())
	}
}

func TestContextImpl_ThreadSafety(t *testing.T) {
	ctx := NewContext()
	widget1 := newMockWidget()
	widget2 := newMockWidget()

	var wg sync.WaitGroup
	const numGoroutines = 100

	// Concurrent focus operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				ctx.RequestFocus(widget1)
			} else {
				ctx.RequestFocus(widget2)
			}
			_ = ctx.FocusedWidget()
			_ = ctx.IsFocused(widget1)
		}(i)
	}

	// Concurrent time operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = ctx.Now()
			_ = ctx.DeltaTime()
			ctx.BeginFrame(time.Now())
		}()
	}

	// Concurrent invalidation
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx.Invalidate()
			ctx.InvalidateRect(geometry.NewRect(0, 0, 100, 100))
			_ = ctx.IsInvalidated()
		}()
	}

	// Concurrent cursor operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ctx.SetCursor(CursorType(i % 12))
			_ = ctx.Cursor()
		}(i)
	}

	wg.Wait()
	// If we get here without deadlock or panic, the test passes
}

func TestContextImpl_ThemeProvider_NilByDefault(t *testing.T) {
	ctx := NewContext()
	if ctx.ThemeProvider() != nil {
		t.Error("expected nil ThemeProvider by default (headless mode)")
	}
}

func TestContextImpl_ThemeProvider_RoundTrip(t *testing.T) {
	ctx := NewContext()
	tp := &mockThemeProvider{dark: true}

	ctx.SetThemeProvider(tp)
	got := ctx.ThemeProvider()
	if got == nil {
		t.Fatal("ThemeProvider() returned nil after SetThemeProvider")
	}
	if !got.IsDark() {
		t.Error("ThemeProvider().IsDark() = false, want true")
	}

	// Set to nil (clear)
	ctx.SetThemeProvider(nil)
	if ctx.ThemeProvider() != nil {
		t.Error("expected nil ThemeProvider after SetThemeProvider(nil)")
	}
}

func TestContextImpl_ThemeProvider_ThreadSafety(t *testing.T) {
	ctx := NewContext()
	tp1 := &mockThemeProvider{dark: false}
	tp2 := &mockThemeProvider{dark: true}

	var wg sync.WaitGroup
	const numGoroutines = 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				ctx.SetThemeProvider(tp1)
			} else {
				ctx.SetThemeProvider(tp2)
			}
			tp := ctx.ThemeProvider()
			if tp != nil {
				_ = tp.IsDark()
			}
		}(i)
	}

	wg.Wait()
	// If we get here without deadlock or panic, the test passes
}

// mockThemeProvider is a minimal ThemeProvider implementation for testing.
type mockThemeProvider struct {
	dark      bool
	onSurface Color
}

func (m *mockThemeProvider) IsDark() bool {
	return m.dark
}

func (m *mockThemeProvider) OnSurface() Color {
	if m.onSurface != (Color{}) {
		return m.onSurface
	}
	// Sensible default: black for light, white for dark.
	if m.dark {
		return ColorWhite
	}
	return ColorBlack
}

// --- BeginFrame timing tests ---
//
// These tests verify the BeginFrame method that was introduced to separate
// inter-frame timing from event-time updates (SetNow). Before this fix,
// DeltaTime was coupled to SetNow, causing incorrect timing when multiple
// HandleEvent calls happened between frames.

func TestBeginFrame_CalculatesDeltaTime(t *testing.T) {
	ctx := NewContext()
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	// First BeginFrame resets lastFrame so the second call has a known baseline.
	ctx.BeginFrame(t0)

	t1 := t0.Add(16 * time.Millisecond)
	ctx.BeginFrame(t1)

	dt := ctx.DeltaTime()
	if dt != 16*time.Millisecond {
		t.Errorf("DeltaTime() = %v, want 16ms", dt)
	}
	if ctx.Now() != t1 {
		t.Errorf("Now() = %v, want %v", ctx.Now(), t1)
	}
}

func TestBeginFrame_ClampsNegativeDeltaToZero(t *testing.T) {
	// Time going backward can happen with system clock adjustments.
	ctx := NewContext()
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 100_000_000, time.UTC) // 100ms
	t1 := time.Date(2026, 1, 1, 0, 0, 0, 50_000_000, time.UTC)  // 50ms (earlier!)

	ctx.BeginFrame(t0)
	ctx.BeginFrame(t1)

	if ctx.DeltaTime() != 0 {
		t.Errorf("DeltaTime() = %v, want 0 (negative delta should be clamped)", ctx.DeltaTime())
	}
}

func TestBeginFrame_ClampsLargeDeltaTo100ms(t *testing.T) {
	// Large delta happens when app resumes from background or debugger pause.
	// Without clamping, animations would jump by seconds.
	ctx := NewContext()
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := t0.Add(5 * time.Second) // 5 seconds gap

	ctx.BeginFrame(t0)
	ctx.BeginFrame(t1)

	if ctx.DeltaTime() != maxDeltaTime {
		t.Errorf("DeltaTime() = %v, want %v (large delta should be clamped)", ctx.DeltaTime(), maxDeltaTime)
	}
}

func TestBeginFrame_IndependentOfSetNow(t *testing.T) {
	// DeltaTime should be based on BeginFrame-to-BeginFrame interval,
	// not affected by SetNow calls in between (which happen during
	// HandleEvent).
	ctx := NewContext()
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	tEvent := t0.Add(50 * time.Millisecond) // Simulates HandleEvent timestamp
	t1 := t0.Add(100 * time.Millisecond)    // Next frame

	ctx.BeginFrame(t0)
	ctx.SetNow(tEvent) // This should NOT affect DeltaTime calculation
	ctx.BeginFrame(t1)

	dt := ctx.DeltaTime()
	if dt != 100*time.Millisecond {
		t.Errorf("DeltaTime() = %v, want 100ms (should be based on lastFrame, not Now)", dt)
	}
}

// --- Invalidation lifecycle tests ---
//
// These tests verify that IsInvalidated persists correctly and is only
// cleared by ClearInvalidation. The animation scheduling fix depends on
// checking IsInvalidated() after layout to decide whether to clear needsLayout.

func TestIsInvalidated_PersistsUntilClear(t *testing.T) {
	ctx := NewContext()

	if ctx.IsInvalidated() {
		t.Error("should not be invalidated initially")
	}

	ctx.Invalidate()
	if !ctx.IsInvalidated() {
		t.Error("should be invalidated after Invalidate()")
	}

	// Multiple Invalidate calls should not change state.
	ctx.Invalidate()
	if !ctx.IsInvalidated() {
		t.Error("should still be invalidated after second Invalidate()")
	}

	ctx.ClearInvalidation()
	if ctx.IsInvalidated() {
		t.Error("should not be invalidated after ClearInvalidation()")
	}
}

func TestContextImpl_Interface(t *testing.T) {
	// Verify ContextImpl implements Context
	var _ Context = (*ContextImpl)(nil)
}

func BenchmarkContextImpl_IsFocused(b *testing.B) {
	ctx := NewContext()
	widget := newMockWidget()
	ctx.RequestFocus(widget)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ctx.IsFocused(widget)
	}
}

func BenchmarkContextImpl_RequestFocus(b *testing.B) {
	ctx := NewContext()
	widget1 := newMockWidget()
	widget2 := newMockWidget()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			ctx.RequestFocus(widget1)
		} else {
			ctx.RequestFocus(widget2)
		}
	}
}

func BenchmarkContextImpl_Now(b *testing.B) {
	ctx := NewContext()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ctx.Now()
	}
}

func BenchmarkContextImpl_Invalidate(b *testing.B) {
	ctx := NewContext()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.Invalidate()
	}
}
