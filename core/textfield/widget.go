package textfield

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/internal/textmetrics"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// Widget implements a full-featured text input field with validation,
// selection, and accessibility support.
//
// A text field is created with [New] using functional options:
//
//	field := textfield.New(
//	    textfield.Placeholder("Enter email"),
//	    textfield.OnChange(handleChange),
//	    textfield.InputType(textfield.TypeEmail),
//	    textfield.MaxLength(255),
//	)
//
// Fluent styling methods may be chained after construction:
//
//	field.Padding(12)
type Widget struct {
	widget.WidgetBase
	cfg     config
	sel     selection
	painter Painter

	// Interaction state.
	hovered  bool
	dragging bool

	// Validation state.
	errorMsg string

	// Styling overrides set via fluent methods.
	padding float32

	// Cached text metrics from last Draw call, used by event handlers
	// (positionFromMouse) that don't have access to canvas.
	cachedMetrics *textmetrics.Metrics
	// Cached layout values from last Draw call.
	cachedContentRect geometry.Rect
	cachedDisplayText string
	cachedFontSize    float32
}

// New creates a new text field Widget with the given options.
//
// The returned widget is visible, enabled, and focusable by default.
// Use options to configure placeholder, change handler, input type, etc.
func New(opts ...Option) *Widget {
	w := &Widget{
		padding: defaultPaddingValue,
		painter: DefaultPainter{},
	}
	w.SetVisible(true)
	w.SetEnabled(true)

	for _, opt := range opts {
		opt(&w.cfg)
	}

	// Apply painter from config if set.
	if w.cfg.painter != nil {
		w.painter = w.cfg.painter
	}

	// Initialize text from signal if bound.
	if w.cfg.signal != nil {
		w.cfg.value = w.cfg.signal.Get()
	}

	// Place cursor at end of initial value.
	runes := []rune(w.cfg.value)
	w.sel.SetCursor(len(runes))

	// Run initial validation.
	if len(w.cfg.validation) > 0 {
		w.errorMsg = runValidation(w.cfg.validation, w.cfg.value)
	}

	return w
}

// Default padding value.
const defaultPaddingValue float32 = 4

// Layout sizing constants.
const (
	defaultFieldHeight float32 = 48
	minFieldWidth      float32 = 100
)

// IsFocusable reports whether the text field can currently receive focus.
// A text field is focusable when it is visible, enabled, and not disabled.
func (w *Widget) IsFocusable() bool {
	return w.IsVisible() && w.IsEnabled() && !w.cfg.ResolvedDisabled()
}

// Layout calculates the text field's preferred size within the given constraints.
func (w *Widget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	width := constraints.MaxWidth
	if width <= 0 || width == geometry.Infinity {
		width = minFieldWidth
	}
	return constraints.Constrain(geometry.Sz(width, defaultFieldHeight))
}

// Draw renders the text field to the canvas.
func (w *Widget) Draw(_ widget.Context, canvas widget.Canvas) {
	// Sync from signal if bound.
	if w.cfg.signal != nil {
		current := w.cfg.signal.Get()
		if current != w.cfg.value {
			w.cfg.value = current
			runes := []rune(current)
			w.sel.SetCursor(clampPos(w.sel.cursor, len(runes)))
		}
	}

	// Query LayoutMetrics from painter (type assert with default fallback).
	lm := resolveLayoutMetrics(w.painter)

	// Compute pre-computed fields.
	text := w.resolvedText()
	bounds := w.Bounds()
	focused := w.IsFocused()
	disabled := w.cfg.ResolvedDisabled()
	hasSelection := w.sel.anchor != w.sel.cursor

	displayText := text
	if w.cfg.inputType == TypePassword {
		displayText = maskText(len([]rune(text)))
	}

	hPad, vPad := lm.ContentPadding()
	contentRect := geometry.Rect{
		Min: geometry.Pt(bounds.Min.X+hPad, bounds.Min.Y+vPad),
		Max: geometry.Pt(bounds.Max.X-hPad, bounds.Max.Y-vPad),
	}

	fontSize := lm.TextFieldFontSize()
	cw := lm.TextFieldCursorWidth()

	// Build text metrics for cursor/selection computation.
	tm := &textmetrics.Metrics{Canvas: canvas, FontSize: fontSize}

	// Cache for event handlers (positionFromMouse).
	w.cachedMetrics = tm
	w.cachedContentRect = contentRect
	w.cachedDisplayText = displayText
	w.cachedFontSize = fontSize

	// Compute cursor rect (if applicable).
	showCursor := focused && !disabled && !hasSelection
	var cursorRect geometry.Rect
	if showCursor {
		cursorRect = tm.CursorRect(contentRect, displayText, w.sel.cursor, cw)
	}

	// Compute selection rect (if applicable).
	showSelection := hasSelection
	var selectionRect geometry.Rect
	if showSelection {
		selectionRect = tm.SelectionRect(contentRect, displayText, w.sel.anchor, w.sel.cursor)
	}

	w.painter.PaintTextField(canvas, &PaintState{
		// Legacy fields.
		Text:        text,
		Placeholder: w.cfg.placeholder,
		Focused:     focused,
		Hovered:     w.hovered,
		Disabled:    disabled,
		HasError:    w.errorMsg != "",
		ErrorMsg:    w.errorMsg,
		CursorPos:   w.sel.cursor,
		SelectStart: w.sel.anchor,
		SelectEnd:   w.sel.cursor,
		InputType:   w.cfg.inputType,
		Bounds:      bounds,

		// Pre-computed fields.
		DisplayText:   displayText,
		ContentRect:   contentRect,
		CursorRect:    cursorRect,
		SelectionRect: selectionRect,
		ShowCursor:    showCursor,
		ShowSelection: showSelection,
		FontSize:      fontSize,
	})
}

// resolveLayoutMetrics returns the LayoutMetrics from the painter if it
// implements that interface, otherwise returns DefaultPainter metrics.
func resolveLayoutMetrics(p Painter) LayoutMetrics {
	if lm, ok := p.(LayoutMetrics); ok {
		return lm
	}
	return DefaultPainter{}
}

// Event handles an input event and returns true if consumed.
func (w *Widget) Event(ctx widget.Context, e event.Event) bool {
	return handleEvent(w, ctx, e)
}

// Children returns nil because a text field is a leaf widget.
func (w *Widget) Children() []widget.Widget {
	return nil
}

// Text returns the current text value.
func (w *Widget) Text() string {
	return w.resolvedText()
}

// SetText sets the text value programmatically and revalidates.
func (w *Widget) SetText(text string) {
	w.setText(text)
	runes := []rune(text)
	w.sel.SetCursor(clampPos(w.sel.cursor, len(runes)))
	w.validate()
}

// ErrorMessage returns the current validation error message, or empty string if valid.
func (w *Widget) ErrorMessage() string {
	return w.errorMsg
}

// HasError returns true if the field has a validation error.
func (w *Widget) HasError() bool {
	return w.errorMsg != ""
}

// CursorPosition returns the current cursor position as a rune index.
func (w *Widget) CursorPosition() int {
	return w.sel.cursor
}

// Selection returns the current selection range as (start, end) rune indices.
// If start == end, there is no selection.
func (w *Widget) Selection() (int, int) {
	return w.sel.OrderedRange()
}

// resolvedText returns the current text, preferring signal over config value.
func (w *Widget) resolvedText() string {
	if w.cfg.signal != nil {
		return w.cfg.signal.Get()
	}
	return w.cfg.value
}

// setText updates the internal text value and the signal if bound.
func (w *Widget) setText(text string) {
	w.cfg.value = text
	if w.cfg.signal != nil {
		w.cfg.signal.Set(text)
	}
}

// textRunes returns the current text as a rune slice.
func (w *Widget) textRunes() []rune {
	return []rune(w.resolvedText())
}

// insertText inserts text at the current cursor position.
func (w *Widget) insertText(text string) {
	runes := w.textRunes()
	pos := w.sel.cursor
	insertRunes := []rune(text)

	// Check max length.
	if w.cfg.maxLength > 0 {
		remaining := w.cfg.maxLength - len(runes)
		if remaining <= 0 {
			return
		}
		if len(insertRunes) > remaining {
			insertRunes = insertRunes[:remaining]
		}
	}

	newRunes := make([]rune, 0, len(runes)+len(insertRunes))
	newRunes = append(newRunes, runes[:pos]...)
	newRunes = append(newRunes, insertRunes...)
	newRunes = append(newRunes, runes[pos:]...)
	w.setText(string(newRunes))
	w.sel.SetCursor(pos + len(insertRunes))
}

// deleteSelection deletes the selected text and places cursor at selection start.
func (w *Widget) deleteSelection() {
	if !w.sel.HasSelection() {
		return
	}

	runes := w.textRunes()
	start, end := w.sel.OrderedRange()
	newRunes := make([]rune, 0, len(runes)-(end-start))
	newRunes = append(newRunes, runes[:start]...)
	newRunes = append(newRunes, runes[end:]...)
	w.setText(string(newRunes))
	w.sel.SetCursor(start)
}

// notifyChange fires the onChange callback and runs validation.
func (w *Widget) notifyChange(ctx widget.Context) {
	w.validate()
	if w.cfg.onChange != nil {
		w.cfg.onChange(w.resolvedText())
	}
	// ADR-028: visual only — text content changed within fixed-size field.
	w.SetNeedsRedraw(true)
	ctx.InvalidateRect(w.Bounds())
}

// validate runs all configured validation functions.
func (w *Widget) validate() {
	w.errorMsg = runValidation(w.cfg.validation, w.resolvedText())
}

// Padding sets the padding around the text field content.
// Returns the widget for method chaining.
func (w *Widget) Padding(v float32) *Widget {
	w.padding = v
	return w
}

// Mount creates signal bindings for push-based invalidation.
// Implements [widget.Lifecycle].
func (w *Widget) Mount(ctx widget.Context) {
	sched := ctx.Scheduler()
	if sched == nil {
		return
	}
	if w.cfg.signal != nil {
		b := state.BindToScheduler(w.cfg.signal, w, sched)
		w.AddBinding(b)
	}
}

// Unmount is called when the text field is removed from the widget tree.
// Implements [widget.Lifecycle].
func (w *Widget) Unmount() {
	// Bindings are cleaned up automatically by WidgetBase.CleanupBindings().
}

// Verify Widget implements required interfaces at compile time.
var (
	_ widget.Widget    = (*Widget)(nil)
	_ widget.Focusable = (*Widget)(nil)
	_ widget.Lifecycle = (*Widget)(nil)
)
