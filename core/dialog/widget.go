package dialog

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// Default layout constants.
const (
	defaultMaxWidth  float32 = 560
	windowMargin     float32 = 0.9 // dialog can occupy up to 90% of window
	minDialogWidth   float32 = 280
	minDialogHeight  float32 = 120
	contentPadding   float32 = 24
	titleAreaHeight  float32 = 52 // padding + title + spacing
	actionAreaHeight float32 = 60 // padding + action buttons + padding
)

// Widget implements a modal dialog with title, content, and action buttons.
//
// A dialog is created in a hidden state with [New]. Call [Widget.Show] to
// push it as a modal overlay, and [Widget.Close] to remove it.
//
// The dialog renders a semi-transparent backdrop that blocks interaction
// with the background, and a centered surface containing the title,
// optional content widget, and action buttons.
type Widget struct {
	widget.WidgetBase
	cfg     config
	painter Painter
	visible bool

	// Overlay surface widget manages the dialog surface content.
	surface *surfaceWidget
}

// New creates a new dialog Widget with the given options.
//
// The returned widget is NOT visible until [Widget.Show] is called.
// By default, the dialog is dismissible (backdrop click) and closes on Escape.
func New(opts ...Option) *Widget {
	w := &Widget{
		painter: DefaultPainter{},
	}
	// Defaults.
	w.cfg.dismissible = true
	w.cfg.escToClose = true
	w.cfg.maxWidth = defaultMaxWidth

	for _, opt := range opts {
		opt(&w.cfg)
	}

	if w.cfg.painter != nil {
		w.painter = w.cfg.painter
	}

	w.SetEnabled(true)

	return w
}

// IsOpen returns true if the dialog is currently visible as an overlay.
func (w *Widget) IsOpen() bool {
	return w.visible
}

// Show pushes the dialog as a modal overlay via the context's OverlayManager.
// If the dialog is already open, this is a no-op.
func (w *Widget) Show(ctx widget.Context) {
	if w.visible {
		return
	}

	om := ctx.OverlayManager()
	if om == nil {
		return
	}

	w.visible = true
	w.SetVisible(true)

	// Create the surface widget for overlay rendering.
	w.surface = newSurfaceWidget(w, ctx)

	om.PushOverlay(w.surface, func() {
		w.doClose(ctx)
	})

	// ADR-028: visual only — overlay display is handled by DrawOverlays.
	w.SetNeedsRedraw(true)
}

// Close removes the dialog from the overlay stack.
// If the dialog is not open, this is a no-op.
func (w *Widget) Close(ctx widget.Context) {
	w.doClose(ctx)
}

// doClose is the internal close implementation.
func (w *Widget) doClose(ctx widget.Context) {
	if !w.visible {
		return
	}

	w.visible = false
	w.SetVisible(false)

	if w.surface != nil {
		om := ctx.OverlayManager()
		if om != nil {
			om.RemoveOverlay(w.surface)
		}
		w.surface = nil
	}

	if w.cfg.onClose != nil {
		w.cfg.onClose()
	}

	// ADR-028: visual only — overlay removal handled by DrawOverlays.
	w.SetNeedsRedraw(true)
}

// Layout calculates the dialog's preferred size. When shown as an overlay,
// layout is handled by the surfaceWidget. This returns zero size when hidden.
func (w *Widget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	if !w.visible {
		return geometry.Size{}
	}
	return constraints.Constrain(geometry.Sz(defaultMaxWidth, minDialogHeight))
}

// Draw renders the dialog. When shown as an overlay, drawing is handled
// by the surfaceWidget. This is a no-op when hidden.
func (w *Widget) Draw(_ widget.Context, _ widget.Canvas) {
	// Rendering is handled by surfaceWidget in the overlay stack.
}

// Event handles input events. When shown as an overlay, events are handled
// by the surfaceWidget. This returns false when hidden.
func (w *Widget) Event(_ widget.Context, _ event.Event) bool {
	return false
}

// Children returns nil because the dialog's content is rendered via the
// overlay surface, not as a direct child.
func (w *Widget) Children() []widget.Widget {
	return nil
}

// Mount creates signal bindings for push-based invalidation.
// Implements [widget.Lifecycle].
func (w *Widget) Mount(ctx widget.Context) {
	sched := ctx.Scheduler()
	if sched == nil {
		return
	}
	if w.cfg.readonlyTitleSignal != nil {
		b := state.BindToScheduler(w.cfg.readonlyTitleSignal, w, sched)
		w.AddBinding(b)
	} else if w.cfg.titleSignal != nil {
		b := state.BindToScheduler(w.cfg.titleSignal, w, sched)
		w.AddBinding(b)
	}
}

// Unmount is called when the dialog is removed from the widget tree.
// Implements [widget.Lifecycle].
func (w *Widget) Unmount() {
	// Bindings are cleaned up automatically by WidgetBase.CleanupBindings().
}

// computeDialogBounds calculates the centered dialog surface bounds
// within the given window size.
func (w *Widget) computeDialogBounds(windowSize geometry.Size) geometry.Rect {
	// Query LayoutMetrics from painter (type assert with default fallback).
	lm := resolveDialogLayoutMetrics(w.painter)

	maxW := w.cfg.maxWidth
	if maxW <= 0 {
		maxW = lm.DialogMaxWidth()
	}

	// Constrain to window bounds.
	availW := windowSize.Width * windowMargin
	availH := windowSize.Height * windowMargin

	if maxW > availW {
		maxW = availW
	}
	if maxW < minDialogWidth && availW >= minDialogWidth {
		maxW = minDialogWidth
	}

	// Compute height based on content.
	dialogH := titleAreaHeight + actionAreaHeight
	if w.cfg.content != nil {
		dialogH += contentPadding * 2 // content padding top + bottom
	}

	maxH := w.cfg.maxHeight
	if maxH <= 0 || maxH > availH {
		maxH = availH
	}
	if dialogH > maxH {
		dialogH = maxH
	}
	if dialogH < minDialogHeight {
		dialogH = minDialogHeight
	}

	// Center in window.
	x := (windowSize.Width - maxW) / 2
	y := (windowSize.Height - dialogH) / 2

	return geometry.NewRect(x, y, maxW, dialogH)
}

// computeActionRects pre-computes action button rects using MeasureText for
// accurate text measurement. This is the preferred path when a canvas is
// available (during Draw). Buttons are laid out right-to-left.
func computeActionRects(canvas widget.Canvas, dialogBounds geometry.Rect, actions []Action) []geometry.Rect {
	if len(actions) == 0 {
		return nil
	}

	rects := make([]geometry.Rect, len(actions))
	x := dialogBounds.Max.X - dialogPadding
	y := dialogBounds.Max.Y - dialogPadding - actionHeight

	for i := len(actions) - 1; i >= 0; i-- {
		textW := canvas.MeasureText(actions[i].Label, actionFontSize, false)
		btnWidth := textW + actionPaddingX*2
		x -= btnWidth
		rects[i] = geometry.NewRect(x, y, btnWidth, actionHeight)
		x -= actionSpacing
	}
	return rects
}

// computeActionRectsLegacy computes action button rects using the character
// count estimate. Used for hit-testing when no canvas is available.
func computeActionRectsLegacy(dialogBounds geometry.Rect, actions []Action) []geometry.Rect {
	if len(actions) == 0 {
		return nil
	}

	rects := make([]geometry.Rect, len(actions))
	x := dialogBounds.Max.X - dialogPadding
	y := dialogBounds.Max.Y - dialogPadding - actionHeight

	for i := len(actions) - 1; i >= 0; i-- {
		btnWidth := float32(len(actions[i].Label))*actionCharWidth + actionPaddingX*2
		x -= btnWidth
		rects[i] = geometry.NewRect(x, y, btnWidth, actionHeight)
		x -= actionSpacing
	}
	return rects
}

// Compile-time interface checks.
var (
	_ widget.Widget    = (*Widget)(nil)
	_ widget.Lifecycle = (*Widget)(nil)
)

// surfaceWidget is the overlay widget that renders the backdrop and dialog surface.
// It implements [overlay.Overlay] to participate in the overlay stack.
type surfaceWidget struct {
	widget.WidgetBase
	dialog     *Widget
	ctx        widget.Context
	focusIndex int // index of focused action button (-1 = none)
}

// newSurfaceWidget creates a new surface widget for rendering the dialog overlay.
func newSurfaceWidget(d *Widget, ctx widget.Context) *surfaceWidget {
	s := &surfaceWidget{
		dialog:     d,
		ctx:        ctx,
		focusIndex: -1,
	}
	s.SetVisible(true)
	s.SetEnabled(true)

	// Set initial focus to the default action.
	for i, a := range d.cfg.actions {
		if a.Default {
			s.focusIndex = i
		}
	}
	// If no default, focus the last action (common UX pattern).
	if s.focusIndex < 0 && len(d.cfg.actions) > 0 {
		s.focusIndex = len(d.cfg.actions) - 1
	}

	return s
}

// Layout fills the entire window.
func (s *surfaceWidget) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	windowSize := ctx.WindowSize()
	size := constraints.Constrain(windowSize)
	s.SetBounds(geometry.NewRect(0, 0, size.Width, size.Height))
	return size
}

// Draw renders the backdrop and dialog surface.
func (s *surfaceWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if canvas == nil {
		return
	}

	d := s.dialog

	// Draw backdrop.
	backdropColor := defaultBackdropColor
	canvas.DrawRect(s.Bounds(), backdropColor)

	// Compute dialog bounds.
	windowSize := ctx.WindowSize()
	dialogBounds := d.computeDialogBounds(windowSize)

	// Pre-compute action button rects using MeasureText (ADR-034 Phase 4).
	actionRects := computeActionRects(canvas, dialogBounds, d.cfg.actions)

	// Delegate to painter for the dialog surface.
	d.painter.PaintDialog(canvas, PaintState{
		Title:       d.cfg.ResolvedTitle(),
		HasContent:  d.cfg.content != nil,
		Actions:     d.cfg.actions,
		Focused:     s.focusIndex >= 0,
		Bounds:      dialogBounds,
		ActionRects: actionRects,
	})

	// Draw content widget if present.
	if d.cfg.content != nil {
		contentBounds := geometry.Rect{
			Min: geometry.Pt(dialogBounds.Min.X+contentPadding, dialogBounds.Min.Y+titleAreaHeight),
			Max: geometry.Pt(dialogBounds.Max.X-contentPadding, dialogBounds.Max.Y-actionAreaHeight),
		}
		if setter, ok := d.cfg.content.(interface{ SetBounds(geometry.Rect) }); ok {
			setter.SetBounds(contentBounds)
		}
		d.cfg.content.Draw(ctx, canvas)
	}
}

// Event handles input events for the dialog overlay.
func (s *surfaceWidget) Event(ctx widget.Context, e event.Event) bool {
	switch ev := e.(type) {
	case *event.KeyEvent:
		return s.handleKeyEvent(ctx, ev)
	case *event.MouseEvent:
		return s.handleMouseEvent(ctx, ev)
	default:
		// Modal: consume all events to prevent interaction with background.
		return true
	}
}

// handleKeyEvent processes keyboard events.
func (s *surfaceWidget) handleKeyEvent(ctx widget.Context, e *event.KeyEvent) bool {
	d := s.dialog

	if e.KeyType != event.KeyPress && e.KeyType != event.KeyRepeat {
		return true // consume but ignore releases for modality
	}

	switch e.Key {
	case event.KeyEscape:
		if d.cfg.escToClose {
			d.doClose(ctx)
		}
		return true

	case event.KeyTab:
		s.cycleFocus(ctx, e.IsShift())
		return true

	case event.KeyEnter, event.KeySpace:
		if s.focusIndex >= 0 && s.focusIndex < len(d.cfg.actions) {
			action := d.cfg.actions[s.focusIndex]
			if action.OnClick != nil {
				action.OnClick()
			}
			d.doClose(ctx)
		}
		return true

	default:
		return true // modal: consume all keys
	}
}

// cycleFocus moves focus to the next or previous action button.
func (s *surfaceWidget) cycleFocus(ctx widget.Context, reverse bool) {
	n := len(s.dialog.cfg.actions)
	if n == 0 {
		return
	}
	if reverse {
		s.focusIndex--
		if s.focusIndex < 0 {
			s.focusIndex = n - 1
		}
	} else {
		s.focusIndex++
		if s.focusIndex >= n {
			s.focusIndex = 0
		}
	}
	// ADR-028: visual only — focus highlight moved between buttons.
	s.SetNeedsRedraw(true)
	ctx.InvalidateRect(s.Bounds())
}

// handleMouseEvent processes mouse events.
func (s *surfaceWidget) handleMouseEvent(ctx widget.Context, e *event.MouseEvent) bool {
	d := s.dialog

	if e.MouseType != event.MousePress {
		return true // consume but only act on press
	}

	// Check if click is outside dialog surface.
	windowSize := ctx.WindowSize()
	dialogBounds := d.computeDialogBounds(windowSize)

	if !dialogBounds.Contains(e.Position) {
		// Click on backdrop.
		if d.cfg.dismissible {
			d.doClose(ctx)
		}
		return true
	}

	// Check if click is on an action button.
	s.handleActionClick(ctx, e, dialogBounds)

	return true
}

// handleActionClick checks if a mouse press hit an action button.
func (s *surfaceWidget) handleActionClick(ctx widget.Context, e *event.MouseEvent, dialogBounds geometry.Rect) {
	d := s.dialog
	if len(d.cfg.actions) == 0 {
		return
	}

	// Use the same pre-computed rects as painting (ADR-034 Phase 4).
	// Fall back to legacy charWidth estimate if no canvas available.
	actionRects := computeActionRectsLegacy(dialogBounds, d.cfg.actions)

	for i, rect := range actionRects {
		if rect.Contains(e.Position) {
			action := d.cfg.actions[i]
			if action.OnClick != nil {
				action.OnClick()
			}
			d.doClose(ctx)
			return
		}
	}
}

// Children returns nil; the surface has no child widgets in the tree.
func (s *surfaceWidget) Children() []widget.Widget {
	return nil
}

// Dismiss is called by the overlay stack when the overlay should close.
func (s *surfaceWidget) Dismiss() {
	s.dialog.doClose(s.ctx)
}

// Modal returns true; dialogs always block interaction with the background.
func (s *surfaceWidget) Modal() bool {
	return true
}
