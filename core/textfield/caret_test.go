package textfield_test

import (
	"testing"
	"time"

	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

const ms = time.Millisecond

func TestHardCaretBlink(t *testing.T) {
	b := textfield.HardCaretBlink(500 * ms)

	tests := []struct {
		elapsed time.Duration
		opacity float32
		next    time.Duration
	}{
		{elapsed: 0, opacity: 1, next: 500 * ms},
		{elapsed: 200 * ms, opacity: 1, next: 300 * ms},
		{elapsed: 499 * ms, opacity: 1, next: 1 * ms},
		{elapsed: 500 * ms, opacity: 0, next: 500 * ms},
		{elapsed: 999 * ms, opacity: 0, next: 1 * ms},
		{elapsed: 1000 * ms, opacity: 1, next: 500 * ms},
		{elapsed: 2750 * ms, opacity: 0, next: 250 * ms},
		{elapsed: -time.Second, opacity: 1, next: 500 * ms},
	}
	for _, tt := range tests {
		if got := b.Opacity(tt.elapsed); got != tt.opacity {
			t.Errorf("Opacity(%v) = %v, want %v", tt.elapsed, got, tt.opacity)
		}
		if got := b.NextChange(tt.elapsed); got != tt.next {
			t.Errorf("NextChange(%v) = %v, want %v", tt.elapsed, got, tt.next)
		}
	}
}

// TestCaretBlinkNextChangeSkipsUnchangedKeyframes checks that a keyframe
// repeating the current opacity does not cost a redraw, within a cycle and
// across the wrap to the next one.
func TestCaretBlinkNextChangeSkipsUnchangedKeyframes(t *testing.T) {
	b := textfield.CaretBlink{
		Cycle: time.Second,
		Keyframes: []textfield.CaretKeyframe{
			{At: 0, Opacity: 1},
			{At: 500 * ms, Opacity: 1},
			{At: 600 * ms, Opacity: 0.5},
			{At: 700 * ms, Opacity: 0},
			{At: 800 * ms, Opacity: 0},
			{At: 900 * ms, Opacity: 0.5},
		},
	}

	tests := []struct {
		elapsed time.Duration
		opacity float32
		next    time.Duration
	}{
		{elapsed: 0, opacity: 1, next: 600 * ms},
		{elapsed: 650 * ms, opacity: 0.5, next: 50 * ms},
		{elapsed: 700 * ms, opacity: 0, next: 200 * ms},
		{elapsed: 950 * ms, opacity: 0.5, next: 50 * ms},
		{elapsed: 1000 * ms, opacity: 1, next: 600 * ms},
	}
	for _, tt := range tests {
		if got := b.Opacity(tt.elapsed); got != tt.opacity {
			t.Errorf("Opacity(%v) = %v, want %v", tt.elapsed, got, tt.opacity)
		}
		if got := b.NextChange(tt.elapsed); got != tt.next {
			t.Errorf("NextChange(%v) = %v, want %v", tt.elapsed, got, tt.next)
		}
	}
}

func TestCaretBlinkThatDoesNotBlink(t *testing.T) {
	blinks := map[string]textfield.CaretBlink{
		"zero value":        {},
		"zero interval":     textfield.HardCaretBlink(0),
		"constant opacity":  {Cycle: time.Second, Keyframes: []textfield.CaretKeyframe{{At: 0, Opacity: 1}, {At: 500 * ms, Opacity: 1}}},
		"no cycle":          {Keyframes: []textfield.CaretKeyframe{{At: 0, Opacity: 1}, {At: 500 * ms, Opacity: 0}}},
		"step past the end": {Cycle: time.Second, Keyframes: []textfield.CaretKeyframe{{At: 0, Opacity: 1}, {At: time.Second, Opacity: 0}}},
	}
	for name, b := range blinks {
		if b.Blinks() {
			t.Errorf("%s: Blinks() = true, want false", name)
		}
		if got := b.Opacity(700 * ms); got != 1 {
			t.Errorf("%s: Opacity = %v, want 1", name, got)
		}
		if got := b.NextChange(700 * ms); got != 0 {
			t.Errorf("%s: NextChange = %v, want 0", name, got)
		}
	}
}

// redrawLog records what a text field asked of the context's
// widget.RedrawScheduler.
type redrawLog struct {
	scheduled []time.Duration
	cancels   int
}

func (l *redrawLog) last() time.Duration {
	if len(l.scheduled) == 0 {
		return 0
	}
	return l.scheduled[len(l.scheduled)-1]
}

// newBlinkContext returns a context that can schedule redraws, at time t0.
func newBlinkContext(t0 time.Time) (*widget.ContextImpl, *redrawLog) {
	ctx := widget.NewContext()
	ctx.SetNow(t0)
	log := &redrawLog{}
	ctx.SetOnScheduleRedraw(
		func(_ widget.Widget, d time.Duration) { log.scheduled = append(log.scheduled, d) },
		func(widget.Widget) { log.cancels++ },
	)
	return ctx, log
}

func newFocusedField(p textfield.Painter, opts ...textfield.Option) *textfield.Widget {
	opts = append([]textfield.Option{textfield.InitialValue("hello"), textfield.PainterOpt(p)}, opts...)
	tf := textfield.New(opts...)
	tf.SetBounds(geometry.NewRect(0, 0, 300, 48))
	tf.SetFocused(true)
	return tf
}

// caretAt draws tf at time at and returns the painted state.
func caretAt(t *testing.T, tf *textfield.Widget, ctx *widget.ContextImpl, p *testPainter, at time.Time) textfield.PaintState {
	t.Helper()
	ctx.SetNow(at)
	tf.Draw(ctx, &mockCanvas{})
	return p.state
}

func TestCaret_BlinksAndSchedulesEachStep(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx, log := newBlinkContext(t0)
	p := &testPainter{}
	tf := newFocusedField(p)

	st := caretAt(t, tf, ctx, p, t0)
	if !st.ShowCursor || st.CursorOpacity != 1 {
		t.Fatalf("at focus: ShowCursor=%v CursorOpacity=%v, want visible at 1", st.ShowCursor, st.CursorOpacity)
	}
	if log.last() != textfield.DefaultCaretBlinkInterval {
		t.Errorf("scheduled redraw in %v, want %v", log.last(), textfield.DefaultCaretBlinkInterval)
	}

	st = caretAt(t, tf, ctx, p, t0.Add(500*ms))
	if st.ShowCursor || st.CursorOpacity != 0 {
		t.Errorf("at 500ms: ShowCursor=%v CursorOpacity=%v, want hidden", st.ShowCursor, st.CursorOpacity)
	}
	if !st.CursorRect.IsEmpty() {
		t.Errorf("hidden caret has CursorRect %v, want empty", st.CursorRect)
	}
	if log.last() != 500*ms {
		t.Errorf("scheduled redraw in %v, want 500ms", log.last())
	}

	// A redraw for another reason mid-step keeps the phase and asks for the
	// remainder of the step.
	st = caretAt(t, tf, ctx, p, t0.Add(1200*ms))
	if !st.ShowCursor {
		t.Error("at 1200ms: caret hidden, want visible")
	}
	if log.last() != 300*ms {
		t.Errorf("scheduled redraw in %v, want 300ms", log.last())
	}
}

func TestCaret_RestartsVisibleOnEdit(t *testing.T) {
	edits := map[string]func(tf *textfield.Widget, ctx widget.Context){
		"typing":       func(tf *textfield.Widget, ctx widget.Context) { typeRune(tf, ctx, 'x') },
		"moving caret": func(tf *textfield.Widget, ctx widget.Context) { pressKey(tf, ctx, event.KeyLeft, event.ModNone) },
		"setting text": func(tf *textfield.Widget, _ widget.Context) { tf.SetText("hello world") },
	}
	for name, edit := range edits {
		t.Run(name, func(t *testing.T) {
			t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
			ctx, log := newBlinkContext(t0)
			p := &testPainter{}
			tf := newFocusedField(p)

			caretAt(t, tf, ctx, p, t0)
			if st := caretAt(t, tf, ctx, p, t0.Add(600*ms)); st.ShowCursor {
				t.Fatal("precondition: caret should be hidden at 600ms")
			}

			edit(tf, ctx)
			st := caretAt(t, tf, ctx, p, t0.Add(600*ms))
			if !st.ShowCursor || st.CursorOpacity != 1 {
				t.Errorf("after edit: ShowCursor=%v CursorOpacity=%v, want visible at 1", st.ShowCursor, st.CursorOpacity)
			}
			if log.last() != 500*ms {
				t.Errorf("scheduled redraw in %v, want a full 500ms step", log.last())
			}
		})
	}
}

func TestCaret_FocusLossCancelsAndRegainRestarts(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx, log := newBlinkContext(t0)
	p := &testPainter{}
	tf := newFocusedField(p)

	caretAt(t, tf, ctx, p, t0)

	tf.SetFocused(false)
	if st := caretAt(t, tf, ctx, p, t0.Add(100*ms)); st.ShowCursor {
		t.Error("unfocused field shows a caret")
	}
	if log.cancels != 1 {
		t.Errorf("CancelRedraw calls = %d, want 1", log.cancels)
	}

	// 700ms into the old cycle would be hidden; regaining focus starts anew.
	tf.SetFocused(true)
	if st := caretAt(t, tf, ctx, p, t0.Add(700*ms)); !st.ShowCursor || st.CursorOpacity != 1 {
		t.Errorf("after refocus: ShowCursor=%v CursorOpacity=%v, want visible at 1", st.ShowCursor, st.CursorOpacity)
	}
	if log.last() != 500*ms {
		t.Errorf("scheduled redraw in %v, want 500ms", log.last())
	}
}

func TestCaret_HiddenWhileWindowInactive(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx, log := newBlinkContext(t0)
	p := &testPainter{}
	tf := newFocusedField(p)

	caretAt(t, tf, ctx, p, t0)
	scheduled := len(log.scheduled)

	ctx.SetWindowFocused(false)
	if st := caretAt(t, tf, ctx, p, t0.Add(100*ms)); st.ShowCursor {
		t.Error("caret shown in an inactive window")
	}
	if len(log.scheduled) != scheduled || log.cancels != 1 {
		t.Errorf("inactive window: %d new schedules, %d cancels; want 0 and 1",
			len(log.scheduled)-scheduled, log.cancels)
	}

	ctx.SetWindowFocused(true)
	if st := caretAt(t, tf, ctx, p, t0.Add(700*ms)); !st.ShowCursor || st.CursorOpacity != 1 {
		t.Errorf("window active again: ShowCursor=%v CursorOpacity=%v, want visible at 1", st.ShowCursor, st.CursorOpacity)
	}
}

// TestCaret_SolidWithoutRedrawScheduler covers contexts that cannot wake
// the field later: a caret drawn in its hidden step would stay hidden, so
// it is drawn solid instead.
func TestCaret_SolidWithoutRedrawScheduler(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx := widget.NewContext()
	p := &testPainter{}
	tf := newFocusedField(p)

	caretAt(t, tf, ctx, p, t0)
	if st := caretAt(t, tf, ctx, p, t0.Add(500*ms)); !st.ShowCursor || st.CursorOpacity != 1 {
		t.Errorf("ShowCursor=%v CursorOpacity=%v, want a solid caret", st.ShowCursor, st.CursorOpacity)
	}
}

// blinkPainter is a painter with its own caret blink.
type blinkPainter struct {
	testPainter
	blink textfield.CaretBlink
}

func (p *blinkPainter) TextFieldCaretBlink() textfield.CaretBlink { return p.blink }

func TestCaret_BlinkComesFromPainterThenOption(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("painter", func(t *testing.T) {
		ctx, log := newBlinkContext(t0)
		p := &blinkPainter{blink: textfield.HardCaretBlink(100 * ms)}
		tf := newFocusedField(p)
		caretAt(t, tf, ctx, &p.testPainter, t0)
		if log.last() != 100*ms {
			t.Errorf("scheduled redraw in %v, want the painter's 100ms", log.last())
		}
	})

	t.Run("option disables blinking", func(t *testing.T) {
		ctx, log := newBlinkContext(t0)
		p := &blinkPainter{blink: textfield.HardCaretBlink(100 * ms)}
		tf := newFocusedField(p, textfield.CaretBlinkOpt(textfield.CaretBlink{}))
		caretAt(t, tf, ctx, &p.testPainter, t0)
		st := caretAt(t, tf, ctx, &p.testPainter, t0.Add(150*ms))
		if !st.ShowCursor || st.CursorOpacity != 1 {
			t.Errorf("ShowCursor=%v CursorOpacity=%v, want a steady caret", st.ShowCursor, st.CursorOpacity)
		}
		if len(log.scheduled) != 0 {
			t.Errorf("steady caret scheduled %v", log.scheduled)
		}
	})
}

func TestPaintCursor(t *testing.T) {
	rect := geometry.NewRect(50, 10, 2, 20)
	color := widget.RGBA(0.2, 0.4, 0.6, 0.8)

	t.Run("square ends draw a centered line", func(t *testing.T) {
		canvas := &recordingCanvas{}
		textfield.PaintCursor(canvas, &textfield.PaintState{ShowCursor: true, CursorRect: rect, CursorOpacity: 0.5}, color, 0)
		if len(canvas.drawLines) != 1 || len(canvas.drawRoundRects) != 0 {
			t.Fatalf("lines=%d roundRects=%d, want 1 line", len(canvas.drawLines), len(canvas.drawRoundRects))
		}
		line := canvas.drawLines[0]
		if line.from != geometry.Pt(50, 10) || line.to != geometry.Pt(50, 30) || line.strokeWidth != 2 {
			t.Errorf("line %v-%v width %v, want (50,10)-(50,30) width 2", line.from, line.to, line.strokeWidth)
		}
		if want := color.WithAlpha(0.4); line.color != want {
			t.Errorf("color %v, want alpha scaled by opacity %v", line.color, want)
		}
	})

	t.Run("rounded ends draw a centered bar", func(t *testing.T) {
		canvas := &recordingCanvas{}
		textfield.PaintCursor(canvas, &textfield.PaintState{ShowCursor: true, CursorRect: rect}, color, 2)
		if len(canvas.drawRoundRects) != 1 || len(canvas.drawLines) != 0 {
			t.Fatalf("lines=%d roundRects=%d, want 1 round rect", len(canvas.drawLines), len(canvas.drawRoundRects))
		}
		bar := canvas.drawRoundRects[0]
		if want := geometry.NewRect(49, 10, 2, 20); bar.r != want {
			t.Errorf("bar %v, want %v centered on x=50", bar.r, want)
		}
		if bar.radius != 1 {
			t.Errorf("radius %v, want 1 (capped at half the width)", bar.radius)
		}
		if bar.color != color {
			t.Errorf("color %v, want %v: zero CursorOpacity means opaque", bar.color, color)
		}
	})

	t.Run("hidden caret draws nothing", func(t *testing.T) {
		canvas := &recordingCanvas{}
		textfield.PaintCursor(canvas, &textfield.PaintState{CursorRect: rect, CursorOpacity: 1}, color, 0)
		if len(canvas.drawLines)+len(canvas.drawRoundRects) != 0 {
			t.Error("PaintCursor drew with ShowCursor false")
		}
	})
}
