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

	widget.PlaySound(widget.SoundAlert)

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

// computeDialogBounds calculates the centered dialog surface bounds within the
// given window size.
//
// It also lays the content widget out. The dialog is what gives the content a
// size: the surface fills the window, so nothing else would lay it out, and a
// container drawn without a layout draws nothing, because its children keep
// empty bounds (#241). The measured content height is part of the dialog
// height, so the content is not clipped to the padding alone.
func (w *Widget) computeDialogBounds(ctx widget.Context, windowSize geometry.Size) geometry.Rect {
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

	// Lay the content out within the width the dialog will have, and take its
	// height into the dialog height.
	maxH := w.cfg.maxHeight
	if maxH <= 0 || maxH > availH {
		maxH = availH
	}
	chromeHeight := titleAreaHeight + actionAreaHeight
	padding := contentPadding
	if w.cfg.content != nil {
		chromeHeight += padding * 2
	}
	if w.cfg.contentOnly {
		chromeHeight, padding = 0, 0
	}
	var contentSize geometry.Size
	if w.cfg.content != nil {
		width := max(0, maxW-padding*2)
		minWidth := float32(0)
		if w.cfg.contentOnly {
			minWidth = width
		}
		contentSize = widget.LayoutChild(w.cfg.content,
			ctx,
			geometry.BoxConstraints(minWidth, width, 0, max(0, maxH-chromeHeight)),
		)
	}

	dialogH := chromeHeight + contentSize.Height
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
func computeActionRects(canvas widget.Canvas, dialogBounds geometry.Rect, actions []Action, pad, actH float32) []geometry.Rect {
	if len(actions) == 0 {
		return nil
	}

	rects := make([]geometry.Rect, len(actions))
	x := dialogBounds.Max.X - pad
	y := dialogBounds.Max.Y - pad - actH

	for i := len(actions) - 1; i >= 0; i-- {
		textW := canvas.MeasureText(actions[i].Label, actionFontSize, false)
		btnWidth := textW + actionPaddingX*2
		x -= btnWidth
		rects[i] = geometry.NewRect(x, y, btnWidth, actH)
		x -= actionSpacing
	}
	return rects
}

// computeActionRectsLegacy computes action button rects using the character
// count estimate. Used for hit-testing when no canvas is available.
func computeActionRectsLegacy(dialogBounds geometry.Rect, actions []Action, pad, actH float32) []geometry.Rect {
	if len(actions) == 0 {
		return nil
	}

	rects := make([]geometry.Rect, len(actions))
	x := dialogBounds.Max.X - pad
	y := dialogBounds.Max.Y - pad - actH

	for i := len(actions) - 1; i >= 0; i-- {
		btnWidth := float32(len(actions[i].Label))*actionCharWidth + actionPaddingX*2
		x -= btnWidth
		rects[i] = geometry.NewRect(x, y, btnWidth, actH)
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
	dialog          *Widget
	ctx             widget.Context
	focusIndex      int // index of focused action button (-1 = none)
	hoverIndex      int // index of hovered action button (-1 = none)
	lastActionRects []geometry.Rect
}

// newSurfaceWidget creates a new surface widget for rendering the dialog overlay.
func newSurfaceWidget(d *Widget, ctx widget.Context) *surfaceWidget {
	s := &surfaceWidget{
		dialog:     d,
		ctx:        ctx,
		focusIndex: -1,
		hoverIndex: -1,
	}
	s.SetVisible(true)
	s.SetEnabled(true)
	if child, ok := d.cfg.content.(interface{ SetParent(widget.Widget) }); ok {
		child.SetParent(s)
	}
	if d.cfg.contentOnly {
		return s
	}

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
	s.positionContent(ctx)
	return size
}

func (s *surfaceWidget) positionContent(ctx widget.Context) geometry.Rect {
	bounds := s.dialog.computeDialogBounds(ctx, ctx.WindowSize())
	contentBounds := bounds
	if !s.dialog.cfg.contentOnly {
		contentBounds = geometry.Rect{
			Min: geometry.Pt(bounds.Min.X+contentPadding, bounds.Min.Y+titleAreaHeight),
			Max: geometry.Pt(bounds.Max.X-contentPadding, bounds.Max.Y-actionAreaHeight),
		}
	}
	if child, ok := s.dialog.cfg.content.(interface{ SetBounds(geometry.Rect) }); ok {
		child.SetBounds(contentBounds)
	}
	return bounds
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

	// Compute dialog bounds (and lay the content out).
	dialogBounds := s.positionContent(ctx)

	// Pre-compute action button rects using MeasureText (ADR-034 Phase 4).
	title, actions := d.cfg.ResolvedTitle(), d.cfg.actions
	if d.cfg.contentOnly {
		title, actions = "", nil
	}
	pad := resolveDialogLayoutMetrics(d.painter).DialogPadding()
	actH := resolveDialogActionHeight(d.painter)
	actionRects := computeActionRects(canvas, dialogBounds, actions, pad, actH)
	s.lastActionRects = actionRects

	// Delegate to painter for the dialog surface.
	d.painter.PaintDialog(canvas, PaintState{
		Title:            title,
		HasContent:       d.cfg.content != nil,
		Actions:          actions,
		Focused:          s.focusIndex >= 0,
		FocusedAction:    s.focusIndex,
		HasHoveredAction: s.hoverIndex >= 0,
		HoveredAction:    s.hoverIndex,
		Bounds:           dialogBounds,
		ActionRects:      actionRects,
	})

	// Draw content widget if present.
	if d.cfg.content != nil {
		widget.StampScreenOrigin(d.cfg.content, canvas)
		widget.DrawChild(d.cfg.content, ctx, canvas)
	}
}

// Event handles input events for the dialog overlay.
func (s *surfaceWidget) Event(ctx widget.Context, e event.Event) bool {
	switch ev := e.(type) {
	case *event.KeyEvent:
		if ev.Key != event.KeyEscape && s.dialog.cfg.content != nil {
			if s.dialog.cfg.content.Event(ctx, e) {
				return true
			}
			if ev.Key == event.KeyTab && (s.dialog.cfg.contentOnly || len(s.dialog.cfg.actions) == 0) {
				return false // let the window traverse focusable controls
			}
		}
		return s.handleKeyEvent(ctx, ev)
	case *event.MouseEvent:
		if s.dispatchContentPointer(ctx, ev.Position, ev) {
			if ev.MouseType == event.MouseMove && s.hoverIndex >= 0 {
				s.hoverIndex = -1
				ctx.SetCursor(widget.CursorDefault)
				s.SetNeedsRedraw(true)
				ctx.InvalidateRect(s.Bounds())
			}
			return true
		}
		return s.handleMouseEvent(ctx, ev)
	case *event.WheelEvent:
		s.dispatchContentPointer(ctx, ev.Position, ev)
		return true
	default:
		// Modal: consume all events to prevent interaction with background.
		return true
	}
}

func (s *surfaceWidget) dispatchContentPointer(ctx widget.Context, point geometry.Point, e event.Event) bool {
	content := s.dialog.cfg.content
	if child, ok := content.(interface{ Bounds() geometry.Rect }); ok && child.Bounds().Contains(point) {
		return content.Event(ctx, e)
	}
	return false
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
		if !d.cfg.contentOnly {
			s.cycleFocus(ctx, e.IsShift())
		}
		return true

	case event.KeyEnter, event.KeySpace:
		if !d.cfg.contentOnly && s.focusIndex >= 0 && s.focusIndex < len(d.cfg.actions) {
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

func (s *surfaceWidget) actionRectsForHit(dialogBounds geometry.Rect) []geometry.Rect {
	actions := s.dialog.cfg.actions
	if len(s.lastActionRects) == len(actions) {
		return s.lastActionRects
	}
	pad := resolveDialogLayoutMetrics(s.dialog.painter).DialogPadding()
	actH := resolveDialogActionHeight(s.dialog.painter)
	return computeActionRectsLegacy(dialogBounds, actions, pad, actH)
}

// handleMouseEvent processes mouse events.
func (s *surfaceWidget) handleMouseEvent(ctx widget.Context, e *event.MouseEvent) bool {
	d := s.dialog

	switch e.MouseType {
	case event.MouseMove:
		if !d.cfg.contentOnly && len(d.cfg.actions) > 0 {
			dialogBounds := d.computeDialogBounds(ctx, ctx.WindowSize())
			hovered := -1
			for i, rect := range s.actionRectsForHit(dialogBounds) {
				if rect.Contains(e.Position) {
					hovered = i
					break
				}
			}
			if hovered != s.hoverIndex {
				s.hoverIndex = hovered
				if hovered >= 0 {
					ctx.SetCursor(widget.CursorPointer)
				} else {
					ctx.SetCursor(widget.CursorDefault)
				}
				s.SetNeedsRedraw(true)
				ctx.InvalidateRect(s.Bounds())
			}
		}
		return true

	case event.MouseLeave:
		if s.hoverIndex >= 0 {
			s.hoverIndex = -1
			ctx.SetCursor(widget.CursorDefault)
			s.SetNeedsRedraw(true)
			ctx.InvalidateRect(s.Bounds())
		}
		return true

	case event.MousePress:
		// Check if click is outside dialog surface.
		windowSize := ctx.WindowSize()
		dialogBounds := d.computeDialogBounds(ctx, windowSize)

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

	default:
		return true // consume other mouse events for modality
	}
}

// handleActionClick checks if a mouse press hit an action button.
func (s *surfaceWidget) handleActionClick(ctx widget.Context, e *event.MouseEvent, dialogBounds geometry.Rect) {
	d := s.dialog
	if d.cfg.contentOnly || len(d.cfg.actions) == 0 {
		return
	}

	actionRects := s.actionRectsForHit(dialogBounds)

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

// Children returns the dialog content for mounting, focus and input routing.
func (s *surfaceWidget) Children() []widget.Widget {
	if s.dialog.cfg.content == nil {
		return nil
	}
	return []widget.Widget{s.dialog.cfg.content}
}

// Dismiss is called by the overlay stack when the overlay should close.
func (s *surfaceWidget) Dismiss() {
	s.dialog.doClose(s.ctx)
}

// Modal returns true; dialogs always block interaction with the background.
func (s *surfaceWidget) Modal() bool {
	return true
}
