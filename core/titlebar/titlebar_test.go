package titlebar

import (
	"github.com/gogpu/gg/scene"
	"image"
	"testing"

	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// --- ControlType Tests ---

func TestControlType_String(t *testing.T) {
	tests := []struct {
		ct   ControlType
		want string
	}{
		{ControlMinimize, "Minimize"},
		{ControlMaximize, "Maximize"},
		{ControlRestore, "Restore"},
		{ControlClose, "Close"},
		{ControlType(99), "Unknown"},
	}
	for _, tc := range tests {
		if got := tc.ct.String(); got != tc.want {
			t.Errorf("ControlType(%d).String() = %q, want %q", tc.ct, got, tc.want)
		}
	}
}

// --- Construction Tests ---

func TestNew_Default(t *testing.T) {
	tb := New()

	if !tb.IsVisible() {
		t.Error("default title bar should be visible")
	}
	if !tb.IsEnabled() {
		t.Error("default title bar should be enabled")
	}
	if tb.IsFocusable() {
		t.Error("title bar should not be focusable")
	}
	if tb.Title() != "" {
		t.Errorf("Title() = %q, want empty", tb.Title())
	}
	if tb.HasChrome() {
		t.Error("HasChrome() should be false without Chrome option")
	}
	if tb.cfg.height != defaultBarHeight {
		t.Errorf("height = %v, want %v", tb.cfg.height, defaultBarHeight)
	}
}

func TestNew_WithTitle(t *testing.T) {
	tb := New(Title("My App"))

	if tb.Title() != "My App" {
		t.Errorf("Title() = %q, want %q", tb.Title(), "My App")
	}
}

func TestNew_WithHeight(t *testing.T) {
	tb := New(Height(48))

	if tb.cfg.height != 48 {
		t.Errorf("height = %v, want 48", tb.cfg.height)
	}
}

func TestNew_WithPainter(t *testing.T) {
	p := &testPainter{}
	tb := New(PainterOpt(p))

	if tb.painter != p {
		t.Error("painter should be the custom painter")
	}
}

func TestNew_WithChrome(t *testing.T) {
	chrome := &mockChrome{}
	tb := New(Chrome(chrome))

	if !tb.HasChrome() {
		t.Error("HasChrome() should be true when Chrome is set")
	}
}

func TestNew_WithLeading(t *testing.T) {
	w1 := &mockWidget{}
	w2 := &mockWidget{}
	tb := New(Leading(w1, w2))

	children := tb.Children()
	if len(children) != 2 {
		t.Fatalf("Children() len = %d, want 2", len(children))
	}
}

func TestNew_WithCenter(t *testing.T) {
	w1 := &mockWidget{}
	tb := New(Center(w1))

	children := tb.Children()
	if len(children) != 1 {
		t.Fatalf("Children() len = %d, want 1", len(children))
	}
}

func TestNew_WithFocused(t *testing.T) {
	tb := New(Focused(false))
	if tb.cfg.focused {
		t.Error("focused should be false")
	}
}

func TestSetTitle(t *testing.T) {
	tb := New(Title("Old"))
	tb.SetTitle("New")
	if tb.Title() != "New" {
		t.Errorf("Title() = %q, want %q", tb.Title(), "New")
	}
}

func TestSetFocusedState(t *testing.T) {
	tb := New()
	tb.SetFocusedState(false)
	if tb.cfg.focused {
		t.Error("focused should be false")
	}
	tb.SetFocusedState(true)
	if !tb.cfg.focused {
		t.Error("focused should be true")
	}
}

// --- Layout Tests ---

func TestLayout_Size(t *testing.T) {
	tb := New(Height(48))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 600))

	size := tb.Layout(ctx, constraints)

	if size.Height != 48 {
		t.Errorf("height = %v, want 48", size.Height)
	}
	if size.Width != 800 {
		t.Errorf("width = %v, want 800 (fills available)", size.Width)
	}
}

func TestLayout_ControlBounds_WithChrome(t *testing.T) {
	chrome := &mockChrome{}
	tb := New(Chrome(chrome))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))

	tb.Layout(ctx, constraints)

	// Controls should be on the right side.
	expectedX := float32(800) - controlButtonWidth*controlCount
	if tb.controlBounds[0].Min.X != expectedX {
		t.Errorf("minimize X = %v, want %v", tb.controlBounds[0].Min.X, expectedX)
	}

	// Each control is controlButtonWidth wide.
	for i := 0; i < controlCount; i++ {
		w := tb.controlBounds[i].Width()
		if w != controlButtonWidth {
			t.Errorf("control[%d] width = %v, want %v", i, w, controlButtonWidth)
		}
	}
}

func TestLayout_ControlBounds_NoChrome(t *testing.T) {
	tb := New()
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))

	tb.Layout(ctx, constraints)

	// Without chrome, control bounds should still be calculated but
	// controlsZoneWidth returns 0.
	if tb.controlsZoneWidth() != 0 {
		t.Errorf("controlsZoneWidth = %v, want 0", tb.controlsZoneWidth())
	}
}

func TestLayout_LeadingChildren(t *testing.T) {
	w1 := &mockWidget{}
	w2 := &mockWidget{}
	chrome := &mockChrome{}
	tb := New(Leading(w1, w2), Chrome(chrome))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))

	tb.Layout(ctx, constraints)

	// First leading widget should start near left edge.
	if tb.leadingBounds[0].Min.X < leadingPadding {
		t.Errorf("leading[0].Min.X = %v, should be >= %v", tb.leadingBounds[0].Min.X, leadingPadding)
	}
	// Second leading widget should be after first.
	if tb.leadingBounds[1].Min.X <= tb.leadingBounds[0].Max.X {
		t.Errorf("leading[1].Min.X = %v, should be > leading[0].Max.X = %v",
			tb.leadingBounds[1].Min.X, tb.leadingBounds[0].Max.X)
	}
}

func TestLayout_CenterChildren(t *testing.T) {
	cw := &mockWidget{}
	chrome := &mockChrome{}
	tb := New(Center(cw), Chrome(chrome))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))

	tb.Layout(ctx, constraints)

	// Center widget should be roughly centered between leading end and controls.
	centerMid := tb.centerBounds[0].Min.X + tb.centerBounds[0].Width()/2
	controlsStart := float32(800) - controlButtonWidth*controlCount
	expectedCenter := controlsStart / 2
	tolerance := float32(40)
	diff := centerMid - expectedCenter
	if diff < -tolerance || diff > tolerance {
		t.Errorf("center widget midpoint = %v, expected near %v", centerMid, expectedCenter)
	}
}

// --- Draw Tests ---

func TestDraw_CallsPainter(t *testing.T) {
	p := &testPainter{}
	chrome := &mockChrome{}
	tb := New(PainterOpt(p), Chrome(chrome))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	canvas := &mockCanvas{}
	tb.Draw(ctx, canvas)

	if !p.bgPainted {
		t.Error("PaintBackground should have been called")
	}
	if p.controlCount != controlCount {
		t.Errorf("PaintControlButton called %d times, want %d", p.controlCount, controlCount)
	}
}

func TestDraw_NoChrome_NoControls(t *testing.T) {
	p := &testPainter{}
	tb := New(PainterOpt(p))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	canvas := &mockCanvas{}
	tb.Draw(ctx, canvas)

	if p.controlCount != 0 {
		t.Errorf("PaintControlButton called %d times, want 0 (no chrome)", p.controlCount)
	}
}

func TestDraw_Invisible(t *testing.T) {
	p := &testPainter{}
	tb := New(PainterOpt(p))
	tb.SetVisible(false)

	ctx := widget.NewContext()
	canvas := &mockCanvas{}
	tb.Draw(ctx, canvas)

	if p.bgPainted {
		t.Error("invisible title bar should not paint")
	}
}

func TestDraw_TitleText(t *testing.T) {
	tb := New(Title("My App"))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	canvas := &mockCanvas{}
	tb.Draw(ctx, canvas)

	if len(canvas.drawTexts) == 0 {
		t.Error("should draw title text when no center children")
	}
	if canvas.drawTexts[0].text != "My App" {
		t.Errorf("title text = %q, want %q", canvas.drawTexts[0].text, "My App")
	}
}

func TestDraw_TitleNotShown_WithCenterWidgets(t *testing.T) {
	cw := &mockWidget{}
	tb := New(Title("My App"), Center(cw))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	canvas := &mockCanvas{}
	tb.Draw(ctx, canvas)

	// Title text should not be drawn when center widgets are present.
	for _, dt := range canvas.drawTexts {
		if dt.text == "My App" {
			t.Error("title text should not be drawn when center children exist")
		}
	}
}

func TestDraw_ControlType_Maximize(t *testing.T) {
	p := &testPainter{}
	chrome := &mockChrome{maximized: false}
	tb := New(PainterOpt(p), Chrome(chrome))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	canvas := &mockCanvas{}
	tb.Draw(ctx, canvas)

	if !p.hasControl(ControlMaximize) {
		t.Error("should draw ControlMaximize when not maximized")
	}
}

func TestDraw_ControlType_Restore(t *testing.T) {
	p := &testPainter{}
	chrome := &mockChrome{maximized: true}
	tb := New(PainterOpt(p), Chrome(chrome))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	canvas := &mockCanvas{}
	tb.Draw(ctx, canvas)

	if !p.hasControl(ControlRestore) {
		t.Error("should draw ControlRestore when maximized")
	}
}

func TestDraw_LeadingChildren(t *testing.T) {
	drawn := false
	cw := &drawTrackingWidget{onDraw: func() { drawn = true }}
	tb := New(Leading(cw))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	canvas := &mockCanvas{}
	tb.Draw(ctx, canvas)

	if !drawn {
		t.Error("leading child Draw should be called")
	}
}

// --- Mouse Event Tests ---

func TestMousePress_ControlButton(t *testing.T) {
	chrome := &mockChrome{}
	tb := New(Chrome(chrome))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	// Press on close button.
	closeCenter := tb.controlBounds[controlIdxClose].Center()
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		closeCenter, closeCenter, event.ModNone)
	consumed := tb.Event(ctx, press)

	if !consumed {
		t.Error("press on control button should be consumed")
	}
	if tb.controlStates[controlIdxClose] != statePressed {
		t.Errorf("close state = %v, want statePressed", tb.controlStates[controlIdxClose])
	}
}

func TestMouseRelease_FiresClose(t *testing.T) {
	chrome := &mockChrome{}
	tb := New(Chrome(chrome))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	closeCenter := tb.controlBounds[controlIdxClose].Center()
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		closeCenter, closeCenter, event.ModNone)
	tb.Event(ctx, press)

	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		closeCenter, closeCenter, event.ModNone)
	tb.Event(ctx, release)

	if !chrome.closeCalled {
		t.Error("Close() should have been called")
	}
}

func TestMouseRelease_FiresMinimize(t *testing.T) {
	chrome := &mockChrome{}
	tb := New(Chrome(chrome))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	minCenter := tb.controlBounds[controlIdxMinimize].Center()
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		minCenter, minCenter, event.ModNone)
	tb.Event(ctx, press)

	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		minCenter, minCenter, event.ModNone)
	tb.Event(ctx, release)

	if !chrome.minimizeCalled {
		t.Error("Minimize() should have been called")
	}
}

func TestMouseRelease_FiresMaximize(t *testing.T) {
	chrome := &mockChrome{}
	tb := New(Chrome(chrome))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	maxCenter := tb.controlBounds[controlIdxMaximize].Center()
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		maxCenter, maxCenter, event.ModNone)
	tb.Event(ctx, press)

	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		maxCenter, maxCenter, event.ModNone)
	tb.Event(ctx, release)

	if !chrome.maximizeCalled {
		t.Error("Maximize() should have been called")
	}
}

func TestMouseRelease_Outside_NoAction(t *testing.T) {
	chrome := &mockChrome{}
	tb := New(Chrome(chrome))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	closeCenter := tb.controlBounds[controlIdxClose].Center()
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		closeCenter, closeCenter, event.ModNone)
	tb.Event(ctx, press)

	// Release far away from close button.
	outside := geometry.Pt(10, 20)
	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		outside, outside, event.ModNone)
	tb.Event(ctx, release)

	if chrome.closeCalled {
		t.Error("Close() should not be called when released outside")
	}
}

func TestMouseMove_HoverState(t *testing.T) {
	chrome := &mockChrome{}
	tb := New(Chrome(chrome))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	closeCenter := tb.controlBounds[controlIdxClose].Center()
	move := event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0,
		closeCenter, closeCenter, event.ModNone)
	tb.Event(ctx, move)

	if tb.controlStates[controlIdxClose] != stateHover {
		t.Errorf("close state = %v, want stateHover", tb.controlStates[controlIdxClose])
	}
}

func TestMouseLeave_ClearsHover(t *testing.T) {
	chrome := &mockChrome{}
	tb := New(Chrome(chrome))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	// Set hover first.
	closeCenter := tb.controlBounds[controlIdxClose].Center()
	move := event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0,
		closeCenter, closeCenter, event.ModNone)
	tb.Event(ctx, move)

	leave := event.NewMouseEvent(event.MouseLeave, event.ButtonNone, 0,
		geometry.Pt(-1, -1), geometry.Pt(-1, -1), event.ModNone)
	tb.Event(ctx, leave)

	if tb.controlStates[controlIdxClose] != stateNormal {
		t.Errorf("close state = %v, want stateNormal after leave", tb.controlStates[controlIdxClose])
	}
}

func TestMousePress_RightButton_Ignored(t *testing.T) {
	chrome := &mockChrome{}
	tb := New(Chrome(chrome))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	closeCenter := tb.controlBounds[controlIdxClose].Center()
	press := event.NewMouseEvent(event.MousePress, event.ButtonRight, event.ButtonStateRight,
		closeCenter, closeCenter, event.ModNone)
	consumed := tb.Event(ctx, press)

	if consumed {
		t.Error("right button press should not be consumed")
	}
}

func TestMousePress_CaptionArea(t *testing.T) {
	chrome := &mockChrome{}
	tb := New(Chrome(chrome))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	// Press in empty area (not on controls or children).
	center := geometry.Pt(200, 20)
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		center, center, event.ModNone)
	consumed := tb.Event(ctx, press)

	if !consumed {
		t.Error("press in caption area should be consumed (for dragging)")
	}
}

func TestMousePress_LeadingChild(t *testing.T) {
	consumed := false
	cw := &eventTrackingWidget{onEvent: func() { consumed = true }}
	tb := New(Leading(cw))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	// Press on leading child.
	childCenter := tb.leadingBounds[0].Center()
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		childCenter, childCenter, event.ModNone)
	tb.Event(ctx, press)

	if !consumed {
		t.Error("press on leading child should dispatch to child")
	}
}

// --- Disabled Tests ---

func TestDisabled_IgnoresEvents(t *testing.T) {
	chrome := &mockChrome{}
	tb := New(Chrome(chrome))
	tb.SetEnabled(false)
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	closeCenter := tb.controlBounds[controlIdxClose].Center()
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		closeCenter, closeCenter, event.ModNone)
	consumed := tb.Event(ctx, press)

	if consumed {
		t.Error("disabled title bar should not consume events")
	}
}

// --- HitTest Tests ---

func TestHitTest_Caption(t *testing.T) {
	chrome := &mockChrome{}
	tb := New(Chrome(chrome))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	// Hit empty area.
	result := tb.HitTest(200, 20)
	if result != HitTestCaption {
		t.Errorf("HitTest(200, 20) = %v, want HitTestCaption", result)
	}
}

func TestHitTest_ControlButtons(t *testing.T) {
	chrome := &mockChrome{}
	tb := New(Chrome(chrome))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	tests := []struct {
		name string
		idx  int
		want HitTestResult
	}{
		{"minimize", controlIdxMinimize, HitTestMinimize},
		{"maximize", controlIdxMaximize, HitTestMaximize},
		{"close", controlIdxClose, HitTestClose},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			center := tb.controlBounds[tc.idx].Center()
			result := tb.HitTest(float64(center.X), float64(center.Y))
			if result != tc.want {
				t.Errorf("HitTest = %v, want %v", result, tc.want)
			}
		})
	}
}

func TestHitTest_LeadingChild(t *testing.T) {
	cw := &mockWidget{}
	chrome := &mockChrome{}
	tb := New(Leading(cw), Chrome(chrome))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	center := tb.leadingBounds[0].Center()
	result := tb.HitTest(float64(center.X), float64(center.Y))
	if result != HitTestClient {
		t.Errorf("HitTest over child = %v, want HitTestClient", result)
	}
}

func TestHitTest_Outside(t *testing.T) {
	tb := New()
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	result := tb.HitTest(200, 100)
	if result != HitTestClient {
		t.Errorf("HitTest outside = %v, want HitTestClient", result)
	}
}

func TestHitTest_NoChrome_Caption(t *testing.T) {
	tb := New()
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	result := tb.HitTest(200, 20)
	if result != HitTestCaption {
		t.Errorf("HitTest without chrome = %v, want HitTestCaption", result)
	}
}

// --- Children Tests ---

func TestChildren_Empty(t *testing.T) {
	tb := New()
	if children := tb.Children(); children != nil {
		t.Errorf("Children() = %v, want nil", children)
	}
}

func TestChildren_LeadingAndCenter(t *testing.T) {
	w1 := &mockWidget{}
	w2 := &mockWidget{}
	w3 := &mockWidget{}
	tb := New(Leading(w1, w2), Center(w3))

	children := tb.Children()
	if len(children) != 3 {
		t.Fatalf("len(Children()) = %d, want 3", len(children))
	}
}

func TestChildren_NilFilter(t *testing.T) {
	tb := New(Leading(nil, &mockWidget{}))
	children := tb.Children()
	if len(children) != 1 {
		t.Fatalf("len(Children()) = %d, want 1 (nil filtered)", len(children))
	}
}

// --- Accessibility Tests ---

func TestAccessibility(t *testing.T) {
	tb := New(Title("Test App"))

	if role := tb.AccessibilityRole(); role != a11y.RoleBanner {
		t.Errorf("AccessibilityRole() = %v, want RoleBanner", role)
	}
	if label := tb.AccessibilityLabel(); label != a11yLabel {
		t.Errorf("AccessibilityLabel() = %q, want %q", label, a11yLabel)
	}
	if hint := tb.AccessibilityHint(); hint != "Test App" {
		t.Errorf("AccessibilityHint() = %q, want %q", hint, "Test App")
	}
	if value := tb.AccessibilityValue(); value != "" {
		t.Errorf("AccessibilityValue() = %q, want empty", value)
	}
}

func TestAccessibilityState_Disabled(t *testing.T) {
	tb := New()
	tb.SetEnabled(false)

	state := tb.AccessibilityState()
	if !state.Disabled {
		t.Error("disabled title bar should report Disabled=true")
	}
}

func TestAccessibilityState_Hidden(t *testing.T) {
	tb := New()
	tb.SetVisible(false)

	state := tb.AccessibilityState()
	if !state.Hidden {
		t.Error("hidden title bar should report Hidden=true")
	}
}

func TestAccessibilityActions_Nil(t *testing.T) {
	tb := New()
	if actions := tb.AccessibilityActions(); actions != nil {
		t.Errorf("AccessibilityActions() = %v, want nil", actions)
	}
}

// --- DefaultPainter Tests ---

func TestDefaultPainter_PaintBackground_EmptyBounds(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	p.PaintBackground(canvas, geometry.Rect{}, BackgroundState{})

	if len(canvas.drawRects) > 0 {
		t.Error("should not paint with empty bounds")
	}
}

func TestDefaultPainter_PaintBackground(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	bounds := geometry.NewRect(0, 0, 800, 40)
	p.PaintBackground(canvas, bounds, BackgroundState{Focused: true})

	if len(canvas.drawRects) == 0 {
		t.Error("should draw background rect")
	}
}

func TestDefaultPainter_PaintControlButton_EmptyBounds(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	p.PaintControlButton(canvas, geometry.Rect{}, ControlClose, ControlState{})

	if len(canvas.drawRects) > 0 || len(canvas.drawLines) > 0 {
		t.Error("should not paint with empty bounds")
	}
}

func TestDefaultPainter_PaintControlButton_Minimize(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	bounds := geometry.NewRect(0, 0, 46, 40)
	p.PaintControlButton(canvas, bounds, ControlMinimize, ControlState{})

	if len(canvas.drawLines) == 0 {
		t.Error("minimize should draw a line")
	}
}

func TestDefaultPainter_PaintControlButton_Maximize(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	bounds := geometry.NewRect(0, 0, 46, 40)
	p.PaintControlButton(canvas, bounds, ControlMaximize, ControlState{})

	if len(canvas.strokeRects) == 0 {
		t.Error("maximize should draw a stroked rect")
	}
}

func TestDefaultPainter_PaintControlButton_Restore(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	bounds := geometry.NewRect(0, 0, 46, 40)
	p.PaintControlButton(canvas, bounds, ControlRestore, ControlState{})

	// Restore draws two rects.
	if len(canvas.strokeRects) < 2 {
		t.Errorf("restore should draw 2 stroked rects, got %d", len(canvas.strokeRects))
	}
}

func TestDefaultPainter_PaintControlButton_Close(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	bounds := geometry.NewRect(0, 0, 46, 40)
	p.PaintControlButton(canvas, bounds, ControlClose, ControlState{})

	// Close draws 2 lines (X shape).
	if len(canvas.drawLines) != 2 {
		t.Errorf("close should draw 2 lines, got %d", len(canvas.drawLines))
	}
}

func TestDefaultPainter_PaintControlButton_CloseHover(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	bounds := geometry.NewRect(0, 0, 46, 40)
	p.PaintControlButton(canvas, bounds, ControlClose, ControlState{Hovered: true})

	// Should draw red background.
	if len(canvas.drawRects) == 0 {
		t.Error("close hover should draw background rect")
	}
}

func TestDefaultPainter_PaintControlButton_MinimizeHover(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	bounds := geometry.NewRect(0, 0, 46, 40)
	p.PaintControlButton(canvas, bounds, ControlMinimize, ControlState{Hovered: true})

	// Should draw hover background.
	if len(canvas.drawRects) == 0 {
		t.Error("minimize hover should draw background rect")
	}
}

func TestDefaultPainter_PaintControlButton_Pressed(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	bounds := geometry.NewRect(0, 0, 46, 40)
	p.PaintControlButton(canvas, bounds, ControlMinimize, ControlState{Pressed: true})

	if len(canvas.drawRects) == 0 {
		t.Error("pressed control should draw background rect")
	}
}

func TestDefaultPainter_PaintControlButton_ClosePressed(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	bounds := geometry.NewRect(0, 0, 46, 40)
	p.PaintControlButton(canvas, bounds, ControlClose, ControlState{Pressed: true})

	if len(canvas.drawRects) == 0 {
		t.Error("close pressed should draw dark red background")
	}
}

// --- resolveHover Tests ---

func TestResolveHover(t *testing.T) {
	tests := []struct {
		name        string
		current     interactionState
		underCursor bool
		want        interactionState
	}{
		{"pressed stays pressed", statePressed, false, statePressed},
		{"pressed stays pressed under cursor", statePressed, true, statePressed},
		{"normal to hover", stateNormal, true, stateHover},
		{"hover to normal", stateHover, false, stateNormal},
		{"hover stays hover", stateHover, true, stateHover},
		{"normal stays normal", stateNormal, false, stateNormal},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveHover(tc.current, tc.underCursor)
			if got != tc.want {
				t.Errorf("resolveHover(%v, %v) = %v, want %v", tc.current, tc.underCursor, got, tc.want)
			}
		})
	}
}

// --- Event dispatch to children ---

func TestDispatchToChildren_FocusEvent(t *testing.T) {
	consumed := false
	cw := &eventTrackingWidget{onEvent: func() { consumed = true }}
	tb := New(Leading(cw))
	ctx := widget.NewContext()

	fe := &event.FocusEvent{}
	tb.Event(ctx, fe)

	if !consumed {
		t.Error("focus event should be dispatched to children")
	}
}

func TestMouseRelease_ToLeadingChild(t *testing.T) {
	consumed := false
	cw := &eventTrackingWidget{onEvent: func() { consumed = true }}
	tb := New(Leading(cw))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	childCenter := tb.leadingBounds[0].Center()
	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		childCenter, childCenter, event.ModNone)
	tb.Event(ctx, release)

	if !consumed {
		t.Error("release on leading child should dispatch to child")
	}
}

func TestMouseMove_ForwardsToChildren(t *testing.T) {
	moved := false
	cw := &eventTrackingWidget{onEvent: func() { moved = true }}
	tb := New(Leading(cw))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	childCenter := tb.leadingBounds[0].Center()
	move := event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0,
		childCenter, childCenter, event.ModNone)
	tb.Event(ctx, move)

	if !moved {
		t.Error("mouse move should be forwarded to child")
	}
}

// --- setBounds helper ---

func TestSetBounds_Helper(t *testing.T) {
	w := &mockWidget{}
	bounds := geometry.NewRect(10, 20, 100, 50)
	setBounds(w, bounds)
	if w.Bounds() != bounds {
		t.Errorf("Bounds() = %v, want %v", w.Bounds(), bounds)
	}
}

// --- Additional coverage tests ---

func TestHitTest_CenterChild(t *testing.T) {
	cw := &mockWidget{}
	chrome := &mockChrome{}
	tb := New(Center(cw), Chrome(chrome))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	center := tb.centerBounds[0].Center()
	result := tb.HitTest(float64(center.X), float64(center.Y))
	if result != HitTestClient {
		t.Errorf("HitTest over center child = %v, want HitTestClient", result)
	}
}

func TestMousePress_CenterChild(t *testing.T) {
	consumed := false
	cw := &eventTrackingWidget{onEvent: func() { consumed = true }}
	chrome := &mockChrome{}
	tb := New(Center(cw), Chrome(chrome))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	childCenter := tb.centerBounds[0].Center()
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		childCenter, childCenter, event.ModNone)
	tb.Event(ctx, press)

	if !consumed {
		t.Error("press on center child should dispatch to child")
	}
}

func TestMouseRelease_ToCenterChild(t *testing.T) {
	consumed := false
	cw := &eventTrackingWidget{onEvent: func() { consumed = true }}
	chrome := &mockChrome{}
	tb := New(Center(cw), Chrome(chrome))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	childCenter := tb.centerBounds[0].Center()
	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		childCenter, childCenter, event.ModNone)
	tb.Event(ctx, release)

	if !consumed {
		t.Error("release on center child should dispatch to child")
	}
}

func TestMouseMove_CenterChild(t *testing.T) {
	moved := false
	cw := &eventTrackingWidget{onEvent: func() { moved = true }}
	chrome := &mockChrome{}
	tb := New(Center(cw), Chrome(chrome))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	childCenter := tb.centerBounds[0].Center()
	move := event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0,
		childCenter, childCenter, event.ModNone)
	tb.Event(ctx, move)

	if !moved {
		t.Error("mouse move should be forwarded to center child")
	}
}

func TestDispatchToChildren_CenterFocusEvent(t *testing.T) {
	consumed := false
	cw := &eventTrackingWidget{onEvent: func() { consumed = true }}
	tb := New(Center(cw))
	ctx := widget.NewContext()

	fe := &event.FocusEvent{}
	tb.Event(ctx, fe)

	if !consumed {
		t.Error("focus event should be dispatched to center children")
	}
}

func TestDispatchToChildren_NilChildren(t *testing.T) {
	tb := New(Leading(nil), Center(nil))
	ctx := widget.NewContext()

	fe := &event.FocusEvent{}
	consumed := tb.Event(ctx, fe)

	if consumed {
		t.Error("nil children should not consume events")
	}
}

func TestMouseRelease_RightButton_Ignored(t *testing.T) {
	chrome := &mockChrome{}
	tb := New(Chrome(chrome))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	release := event.NewMouseEvent(event.MouseRelease, event.ButtonRight, 0,
		geometry.Pt(200, 20), geometry.Pt(200, 20), event.ModNone)
	consumed := tb.Event(ctx, release)

	if consumed {
		t.Error("right button release should not be consumed")
	}
}

func TestMousePress_NilLeadingChild(t *testing.T) {
	chrome := &mockChrome{}
	tb := New(Leading(nil), Chrome(chrome))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(10, 20), geometry.Pt(10, 20), event.ModNone)
	// Should not panic.
	tb.Event(ctx, press)
}

func TestMouseRelease_CaptionArea(t *testing.T) {
	chrome := &mockChrome{}
	tb := New(Chrome(chrome))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	// Release in caption area (no pressed control).
	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		geometry.Pt(200, 20), geometry.Pt(200, 20), event.ModNone)
	consumed := tb.Event(ctx, release)

	// Should still be consumed (caption release).
	if !consumed {
		t.Error("release in caption area should be consumed")
	}
}

func TestControlTypeForIndex_Unknown(t *testing.T) {
	chrome := &mockChrome{}
	tb := New(Chrome(chrome))

	// Invalid index should default to ControlClose.
	ct := tb.controlTypeForIndex(99)
	if ct != ControlClose {
		t.Errorf("controlTypeForIndex(99) = %v, want ControlClose", ct)
	}
}

func TestMouseLeave_NoChange(t *testing.T) {
	tb := New() // No chrome, all states are normal.
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	leave := event.NewMouseEvent(event.MouseLeave, event.ButtonNone, 0,
		geometry.Pt(-1, -1), geometry.Pt(-1, -1), event.ModNone)
	consumed := tb.Event(ctx, leave)

	if consumed {
		t.Error("leave with no hovered controls should not be consumed")
	}
}

func TestDraw_CenterChildren(t *testing.T) {
	drawn := false
	cw := &drawTrackingWidget{onDraw: func() { drawn = true }}
	tb := New(Center(cw))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	canvas := &mockCanvas{}
	tb.Draw(ctx, canvas)

	if !drawn {
		t.Error("center child Draw should be called")
	}
}

func TestLayout_NilLeadingChild(t *testing.T) {
	tb := New(Leading(nil, &mockWidget{}))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))

	// Should not panic.
	tb.Layout(ctx, constraints)

	if tb.leadingBounds[0] != (geometry.Rect{}) {
		t.Error("nil leading child should have empty bounds")
	}
}

func TestLayout_NilCenterChild(t *testing.T) {
	tb := New(Center(nil, &mockWidget{}))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))

	// Should not panic.
	tb.Layout(ctx, constraints)

	if tb.centerBounds[0] != (geometry.Rect{}) {
		t.Error("nil center child should have empty bounds")
	}
}

func TestDraw_NilChildren_NoPanic(t *testing.T) {
	tb := New(Leading(nil), Center(nil))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 40))
	tb.Layout(ctx, constraints)

	canvas := &mockCanvas{}
	// Should not panic.
	tb.Draw(ctx, canvas)
}

// --- Mock Types ---

type mockChrome struct {
	maximized      bool
	closeCalled    bool
	minimizeCalled bool
	maximizeCalled bool
}

func (c *mockChrome) Minimize()         { c.minimizeCalled = true }
func (c *mockChrome) Maximize()         { c.maximizeCalled = true }
func (c *mockChrome) IsMaximized() bool { return c.maximized }
func (c *mockChrome) Close()            { c.closeCalled = true }

type testPainter struct {
	bgPainted    bool
	controlCount int
	controls     []ControlType
}

func (p *testPainter) PaintBackground(_ widget.Canvas, _ geometry.Rect, _ BackgroundState) {
	p.bgPainted = true
}

func (p *testPainter) PaintControlButton(_ widget.Canvas, _ geometry.Rect, control ControlType, _ ControlState) {
	p.controlCount++
	p.controls = append(p.controls, control)
}

func (p *testPainter) hasControl(ct ControlType) bool {
	for _, c := range p.controls {
		if c == ct {
			return true
		}
	}
	return false
}

type mockWidget struct {
	widget.WidgetBase
}

func (w *mockWidget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	size := constraints.Constrain(geometry.Sz(60, 30))
	w.SetBounds(geometry.FromPointSize(w.Position(), size))
	return size
}

func (w *mockWidget) Draw(_ widget.Context, _ widget.Canvas) {}

func (w *mockWidget) Event(_ widget.Context, _ event.Event) bool { return false }

func (w *mockWidget) Children() []widget.Widget { return nil }

type drawTrackingWidget struct {
	widget.WidgetBase
	onDraw func()
}

func (w *drawTrackingWidget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	size := constraints.Constrain(geometry.Sz(60, 30))
	w.SetBounds(geometry.FromPointSize(w.Position(), size))
	return size
}

func (w *drawTrackingWidget) Draw(_ widget.Context, _ widget.Canvas) {
	if w.onDraw != nil {
		w.onDraw()
	}
}

func (w *drawTrackingWidget) Event(_ widget.Context, _ event.Event) bool { return false }

func (w *drawTrackingWidget) Children() []widget.Widget { return nil }

type eventTrackingWidget struct {
	widget.WidgetBase
	onEvent func()
}

func (w *eventTrackingWidget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	size := constraints.Constrain(geometry.Sz(60, 30))
	w.SetBounds(geometry.FromPointSize(w.Position(), size))
	return size
}

func (w *eventTrackingWidget) Draw(_ widget.Context, _ widget.Canvas) {}

func (w *eventTrackingWidget) Event(_ widget.Context, _ event.Event) bool {
	if w.onEvent != nil {
		w.onEvent()
	}
	return true
}

func (w *eventTrackingWidget) Children() []widget.Widget { return nil }

// --- mockCanvas ---

type mockCanvas struct {
	drawRects      []drawRectCall
	strokeRects    []strokeRectCall
	drawRoundRects []drawRoundRectCall
	drawTexts      []drawTextCall
	drawLines      []drawLineCall
}

type drawRectCall struct {
	r     geometry.Rect
	color widget.Color
}

type strokeRectCall struct {
	r           geometry.Rect
	color       widget.Color
	strokeWidth float32
}

type drawRoundRectCall struct {
	r      geometry.Rect
	color  widget.Color
	radius float32
}

type drawTextCall struct {
	text     string
	bounds   geometry.Rect
	fontSize float32
	color    widget.Color
	bold     bool
	align    widget.TextAlign
}

type drawLineCall struct {
	from, to    geometry.Point
	color       widget.Color
	strokeWidth float32
}

func (c *mockCanvas) Clear(_ widget.Color) {}

func (c *mockCanvas) DrawRect(r geometry.Rect, color widget.Color) {
	c.drawRects = append(c.drawRects, drawRectCall{r: r, color: color})
}

func (c *mockCanvas) FillRectDirect(_ geometry.Rect, _ widget.Color) {}

func (c *mockCanvas) StrokeRect(r geometry.Rect, color widget.Color, strokeWidth float32) {
	c.strokeRects = append(c.strokeRects, strokeRectCall{r: r, color: color, strokeWidth: strokeWidth})
}

func (c *mockCanvas) DrawRoundRect(r geometry.Rect, color widget.Color, radius float32) {
	c.drawRoundRects = append(c.drawRoundRects, drawRoundRectCall{r: r, color: color, radius: radius})
}

func (c *mockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {}

func (c *mockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)              {}
func (c *mockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {}
func (c *mockCanvas) StrokeArc(_ geometry.Point, _ float32, _, _ float64, _ widget.Color, _ float32) {
}

func (c *mockCanvas) DrawLine(from, to geometry.Point, color widget.Color, strokeWidth float32) {
	c.drawLines = append(c.drawLines, drawLineCall{from: from, to: to, color: color, strokeWidth: strokeWidth})
}

func (c *mockCanvas) DrawText(text string, bounds geometry.Rect, fontSize float32, color widget.Color, bold bool, align widget.TextAlign) {
	c.drawTexts = append(c.drawTexts, drawTextCall{text: text, bounds: bounds, fontSize: fontSize, color: color, bold: bold, align: align})
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
func (c *mockCanvas) ScreenOriginBase() geometry.Point             { return geometry.Point{} }
func (c *mockCanvas) ClipBounds() geometry.Rect                    { return geometry.NewRect(0, 0, 10000, 10000) }
func (c *mockCanvas) ReplayScene(_ *scene.Scene)                   {}
