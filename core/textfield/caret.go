package textfield

import (
	"time"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// CaretKeyframe is one step of a caret blink cycle: from At until the next
// keyframe, the caret is drawn at Opacity.
type CaretKeyframe struct {
	At      time.Duration
	Opacity float32
}

// CaretBlink describes how the caret blinks while the field has focus and
// no selection. The zero value does not blink: the caret stays visible.
//
// Opacity steps from one keyframe to the next without interpolation, so the
// field repaints only when a step is due and the window stays idle between
// steps. The cycle restarts, caret visible, whenever the field gains focus,
// the caret moves, or the text changes, so the caret never disappears while
// the user types.
type CaretBlink struct {
	// Cycle is the length of one blink cycle.
	Cycle time.Duration

	// Keyframes are the opacity steps of one cycle, sorted by At, each in
	// [0, Cycle). Before the first keyframe the caret is fully visible.
	// The slice is read, never modified.
	Keyframes []CaretKeyframe
}

// DefaultCaretBlinkInterval is how long a hard-blinking caret stays visible
// and then hidden: 500 ms, the half period of Flutter and Jetpack Compose,
// and the default of IntelliJ-based IDEs.
const DefaultCaretBlinkInterval = 500 * time.Millisecond

// HardCaretBlink returns a blink that shows the caret for interval, then
// hides it for interval, switching sharply, as Material, Win32 and JetBrains
// carets do. An interval of zero or less returns the zero CaretBlink, which
// does not blink.
func HardCaretBlink(interval time.Duration) CaretBlink {
	if interval <= 0 {
		return CaretBlink{}
	}
	return CaretBlink{
		Cycle: 2 * interval,
		Keyframes: []CaretKeyframe{
			{At: 0, Opacity: 1},
			{At: interval, Opacity: 0},
		},
	}
}

// Opacity returns the caret opacity, from 0 to 1, elapsed after the cycle
// started.
func (b CaretBlink) Opacity(elapsed time.Duration) float32 {
	if !b.Blinks() {
		return 1
	}
	return b.opacityAt(b.phase(elapsed))
}

// NextChange returns how long after elapsed the opacity next changes, or
// zero when it never changes.
func (b CaretBlink) NextChange(elapsed time.Duration) time.Duration {
	if !b.Blinks() {
		return 0
	}
	elapsed = max(elapsed, 0)
	phase := b.phase(elapsed)
	current := b.opacityAt(phase)
	cycleStart := elapsed - phase
	// The opacity can only change at a cycle start or a keyframe. Blinks saw
	// two different opacities among those, so the current cycle and the next
	// one hold a change.
	for cycle := range 2 {
		base := cycleStart + time.Duration(cycle)*b.Cycle
		if at := base; at > elapsed && b.opacityAt(0) != current {
			return at - elapsed
		}
		for _, k := range b.Keyframes {
			if k.At < 0 || k.At >= b.Cycle {
				continue
			}
			if at := base + k.At; at > elapsed && b.opacityAt(k.At) != current {
				return at - elapsed
			}
		}
	}
	return 0
}

// Blinks reports whether the caret opacity changes over a cycle.
func (b CaretBlink) Blinks() bool {
	if b.Cycle <= 0 {
		return false
	}
	first := b.opacityAt(0)
	for _, k := range b.Keyframes {
		if k.At >= 0 && k.At < b.Cycle && b.opacityAt(k.At) != first {
			return true
		}
	}
	return false
}

func (b CaretBlink) phase(elapsed time.Duration) time.Duration {
	if elapsed <= 0 {
		return 0
	}
	return elapsed % b.Cycle
}

func (b CaretBlink) opacityAt(phase time.Duration) float32 {
	opacity := float32(1)
	for _, k := range b.Keyframes {
		if k.At > phase {
			break
		}
		opacity = k.Opacity
	}
	return min(max(opacity, 0), 1)
}

// CaretBlinker is implemented by painters that set how their caret blinks.
// A text field whose painter does not implement it blinks like
// [DefaultPainter]. The [CaretBlinkOpt] option overrides both.
type CaretBlinker interface {
	TextFieldCaretBlink() CaretBlink
}

// PaintCursor draws the caret for st: a vertical bar as wide as
// st.CursorRect, centered on the insertion point at st.CursorRect.Min.X and
// spanning its height, in color with its alpha scaled by st.CursorOpacity.
// A positive radius rounds the bar's ends; it is capped at half the width.
// PaintCursor draws nothing when st.ShowCursor is false.
func PaintCursor(canvas widget.Canvas, st *PaintState, color widget.Color, radius float32) {
	if !st.ShowCursor {
		return
	}
	if st.CursorOpacity > 0 && st.CursorOpacity < 1 {
		color = color.WithAlpha(color.A * st.CursorOpacity)
	}

	width := st.CursorRect.Width()
	x := st.CursorRect.Min.X
	if radius <= 0 {
		top := geometry.Pt(x, st.CursorRect.Min.Y)
		bottom := geometry.Pt(x, st.CursorRect.Max.Y)
		canvas.DrawLine(top, bottom, color, width)
		return
	}
	bar := geometry.NewRect(x-width/2, st.CursorRect.Min.Y, width, st.CursorRect.Height())
	canvas.DrawRoundRect(bar, color, min(radius, width/2))
}

// caretBlinkState restarts the blink cycle when the caret appears, moves, or
// the text around it changes.
type caretBlinkState struct {
	shown   bool
	shownAt time.Time
	pos     int
	text    string
}

// caretOpacity returns the opacity to draw the caret at, 0 when it is hidden,
// and schedules the field's next redraw for the next blink step.
//
// The caret is drawn solid when the context cannot schedule redraws:
// without a later redraw, a caret drawn in its hidden step would stay hidden.
func (w *Widget) caretOpacity(ctx widget.Context, show bool, text string) float32 {
	rs, _ := ctx.(widget.RedrawScheduler)
	st := &w.caret
	if !show {
		if st.shown && rs != nil {
			rs.CancelRedraw(w)
		}
		st.shown = false
		return 0
	}

	now := ctx.Now()
	if !st.shown || st.pos != w.sel.cursor || st.text != text {
		st.shown = true
		st.shownAt = now
		st.pos = w.sel.cursor
		st.text = text
	}

	blink := w.resolveCaretBlink()
	elapsed := now.Sub(st.shownAt)
	next := blink.NextChange(elapsed)
	if next <= 0 {
		if rs != nil {
			rs.CancelRedraw(w)
		}
		return 1
	}
	if rs == nil || !rs.ScheduleRedraw(w, next) {
		return 1
	}
	return blink.Opacity(elapsed)
}

// resolveCaretBlink returns the blink set with [CaretBlinkOpt], else the
// painter's, else [DefaultPainter]'s.
func (w *Widget) resolveCaretBlink() CaretBlink {
	if w.cfg.caretBlink != nil {
		return *w.cfg.caretBlink
	}
	if b, ok := w.painter.(CaretBlinker); ok {
		return b.TextFieldCaretBlink()
	}
	return DefaultPainter{}.TextFieldCaretBlink()
}

// windowFocused reports whether the window has OS focus, assuming it does
// when the context cannot tell.
func windowFocused(ctx widget.Context) bool {
	if r, ok := ctx.(widget.WindowFocusReporter); ok {
		return r.IsWindowFocused()
	}
	return true
}
