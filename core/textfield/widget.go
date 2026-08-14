package textfield

import (
	"unicode/utf8"

	"github.com/gogpu/gpucontext"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/gesture"
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

	// Gesture recognizer for tap/drag/multi-click text selection.
	// Handles single-click (cursor placement), double-click (word selection),
	// triple-click (select all), and drag (selection extension).
	// Replaces ad-hoc dragging field and MouseDrag/MouseDoubleClick handling.
	tapDrag *gesture.TapAndDragRecognizer

	// Interaction state.
	hovered bool

	// Validation state.
	errorMsg string

	// Styling overrides set via fluent methods.
	padding float32

	// Horizontal scroll offset (always <= 0). When text exceeds the visible
	// content area, this shifts the text left so the cursor stays visible.
	// Enterprise references: Flutter RenderEditable._showCaretOnScreen(),
	// Qt QLineEdit d->hscroll, HTML input.scrollLeft.
	scrollOffsetX float32

	// Cached text metrics from last Draw call, used by gesture handlers
	// (positionFromGlobal) that don't have access to canvas.
	cachedMetrics *textmetrics.Metrics
	// Cached layout values from last Draw call.
	cachedContentRect   geometry.Rect
	cachedDisplayText   string
	cachedFontSize      float32
	cachedIMECursorArea gpucontext.IMECursorArea
	cachedIMEAreaSet    bool
	gestureHandledTap   bool

	// IME composition is kept separate from committed text. Cursor and
	// selection positions in the preedit are UTF-8 byte offsets supplied by the
	// gpucontext v2 contract and converted to rune positions only for drawing.
	composition gpucontext.IMEComposition
	composing   bool
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

	// Create TapAndDragRecognizer for unified click/drag/multi-click handling.
	// This replaces ad-hoc MousePress, MouseDrag, and MouseDoubleClick handlers
	// with gesture system callbacks (ADR-049 Phase 3, #225).
	w.tapDrag = gesture.NewTapAndDragRecognizer(gesture.TapAndDragConfig{
		OnTapDown:    w.handleGestureTapDown,
		OnDragStart:  w.handleGestureDragStart,
		OnDragUpdate: w.handleGestureDragUpdate,
	})

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

// SetEnabled updates the widget's enabled state and invalidates any transient
// IME preedit immediately. Programmatic toggles do not necessarily produce a
// FocusEvent, so waiting until Draw would let a disabled-then-reenabled field
// resurrect stale composition text.
func (w *Widget) SetEnabled(enabled bool) {
	if !enabled {
		w.CancelIMEComposition()
	}
	w.WidgetBase.SetEnabled(enabled)
}

// SetVisible updates visibility and clears transient IME state when the field
// is hidden. This mirrors SetEnabled for callers that toggle visibility
// reactively between frames.
func (w *Widget) SetVisible(visible bool) {
	if !visible {
		w.CancelIMEComposition()
	}
	w.WidgetBase.SetVisible(visible)
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
			w.clearComposition()
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
	if !focused || disabled || !w.IsEnabled() || !w.IsVisible() {
		// A programmatic focus/disable/visibility change may not emit a
		// FocusEvent; cancel the in-memory preedit before it can reappear on
		// re-enable or on a later focus restore.
		w.clearComposition()
	}
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

	// Ensure cursor is visible within the content rect (adjusts scrollOffsetX).
	w.ensureCursorVisible(tm, contentRect, displayText)

	// TextRect: text rendering area shifted by scroll offset.
	// ContentRect stays unshifted for clipping (PushClip).
	scrolledRect := contentRect
	scrolledRect.Min.X += w.scrollOffsetX
	scrolledRect.Max.X += w.scrollOffsetX

	// Cache for gesture handlers (positionFromGlobal).
	w.cachedMetrics = tm
	w.cachedContentRect = contentRect
	w.cachedDisplayText = displayText
	w.cachedFontSize = fontSize

	// Flutter/Qt pattern: compute cursor/selection in UNSHIFTED content rect
	// (text-local coordinates), then apply scrollOffsetX uniformly.
	// This ensures cursor and text share the same offset — they never drift apart.
	showCursor := focused && !disabled && !hasSelection
	var cursorRect geometry.Rect
	if showCursor {
		cursorRect = tm.CursorRect(contentRect, displayText, w.sel.cursor, cw)
		cursorRect.Min.X += w.scrollOffsetX
		cursorRect.Max.X += w.scrollOffsetX
	}

	showSelection := hasSelection
	var selectionRect geometry.Rect
	if showSelection {
		selectionRect = tm.SelectionRect(contentRect, displayText, w.sel.anchor, w.sel.cursor)
		selectionRect.Min.X += w.scrollOffsetX
		selectionRect.Max.X += w.scrollOffsetX
	}

	compositionText, compositionRect, compositionSelectionRect, compositionCursorRect :=
		w.compositionPaintGeometry(tm, contentRect, displayText)
	showComposition := focused && w.IsVisible() && w.IsEnabled() &&
		compositionText != "" && !compositionRect.IsEmpty() && !disabled
	if showComposition && !compositionCursorRect.IsEmpty() {
		// The marked-text cursor is the only insertion caret during a visible
		// preedit. Drawing the committed caret as well produces two cursors,
		// often at different positions, while the IME is active.
		showCursor = false
	}
	if showComposition {
		// Candidate placement follows the active preedit cursor when the IME
		// provides one; otherwise it remains anchored to the committed caret.
		candidate := compositionCursorRect
		if candidate.IsEmpty() {
			candidate = cursorRect
		}
		w.cachedIMECursorArea = w.areaFromLocalRect(candidate, bounds)
	} else {
		w.cachedIMECursorArea = w.areaFromLocalRect(cursorRect, bounds)
	}
	w.cachedIMEAreaSet = true

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
		DisplayText:              displayText,
		ContentRect:              contentRect,
		TextRect:                 scrolledRect,
		CursorRect:               cursorRect,
		SelectionRect:            selectionRect,
		ShowCursor:               showCursor,
		ShowSelection:            showSelection,
		FontSize:                 fontSize,
		CompositionText:          compositionText,
		CompositionTextRect:      compositionRect,
		CompositionSelectionRect: compositionSelectionRect,
		CompositionCursorRect:    compositionCursorRect,
		ShowComposition:          showComposition,
	})
}

// compositionPaintGeometry converts the versioned UTF-8 composition ranges
// into the rune-index geometry used by textmetrics. The committed text stays
// unchanged; the preedit is drawn at the committed caret and therefore never
// participates in selection or clipboard operations until it is committed.
func (w *Widget) compositionPaintGeometry(
	tm *textmetrics.Metrics,
	contentRect geometry.Rect,
	displayText string,
) (string, geometry.Rect, geometry.Rect, geometry.Rect) {
	if !w.composing || w.cfg.inputType == TypePassword {
		return "", geometry.Rect{}, geometry.Rect{}, geometry.Rect{}
	}
	composition := w.composition
	if !composition.IsValid() || composition.CompositionText == "" {
		return "", geometry.Rect{}, geometry.Rect{}, geometry.Rect{}
	}

	// IsValid above guarantees that all supplied byte offsets are UTF-8
	// boundaries in CompositionText, so conversion cannot fail here.
	start, _ := compositionRangeToRunes(composition.CompositionText, composition.SelectionStart, composition.SelectionEnd)
	cursorBegin, cursorEnd := composition.CursorBegin, composition.CursorEnd
	if cursorBegin < 0 || cursorEnd < 0 {
		cursorBegin, cursorEnd = 0, 0
	}
	cursorStart, _ := byteOffsetToRune(composition.CompositionText, cursorBegin)

	startX := tm.CursorX(contentRect, displayText, w.sel.cursor) + w.scrollOffsetX
	width := tm.Canvas.MeasureText(composition.CompositionText, tm.FontSize, false)
	compositionRect := geometry.NewRect(
		startX,
		contentRect.Min.Y,
		width,
		contentRect.Height(),
	)
	selectionRect := geometry.Rect{}
	if start != nil {
		selectionRect = tm.SelectionRect(compositionRect, composition.CompositionText, start[0], start[1])
	}
	cursorRect := geometry.Rect{}
	if composition.HasCursor() {
		// CursorBegin is the leading edge of the valid cursor range, so selected
		// segments remain visibly underlined instead of moving the caret.
		cursorRect = tm.CursorRect(compositionRect, composition.CompositionText, cursorStart, resolveLayoutMetrics(w.painter).TextFieldCursorWidth())
	}
	return composition.CompositionText, compositionRect, selectionRect, cursorRect
}

// areaFromLocalRect converts a widget-local caret rectangle to window-local
// logical DIP coordinates as required by gpucontext.IMECursorArea.
func (w *Widget) areaFromLocalRect(rect, bounds geometry.Rect) gpucontext.IMECursorArea {
	if rect.IsEmpty() {
		return gpucontext.IMECursorArea{}
	}
	origin := w.ScreenOrigin()
	return gpucontext.IMECursorArea{
		X:      float64(origin.X + rect.Min.X - bounds.Min.X),
		Y:      float64(origin.Y + rect.Min.Y - bounds.Min.Y),
		Width:  float64(rect.Width()),
		Height: float64(rect.Height()),
	}
}

// byteOffsetToRune converts a validated UTF-8 byte offset to a rune index.
func byteOffsetToRune(text string, offset int) (int, bool) {
	if offset < 0 || offset > len(text) || !utf8.ValidString(text) {
		return 0, false
	}
	if offset < len(text) && !utf8.RuneStart(text[offset]) {
		return 0, false
	}
	return len([]rune(text[:offset])), true
}

func compositionRangeToRunes(text string, start, end int) ([]int, bool) {
	if start == end {
		return nil, true
	}
	first, ok := byteOffsetToRune(text, start)
	if !ok {
		return nil, false
	}
	last, ok := byteOffsetToRune(text, end)
	if !ok || first > last {
		return nil, false
	}
	return []int{first, last}, true
}

// IMEEnabled reports whether this field should own the platform IME.
// Password fields deliberately disable native composition to avoid leaking
// preedit/surrounding text into IME candidate stores.
func (w *Widget) IMEEnabled() bool {
	return w.IsFocused() && w.IsVisible() && w.IsEnabled() &&
		!w.cfg.ResolvedDisabled() && w.cfg.inputType != TypePassword
}

// IMEContentType returns the advisory purpose and privacy hints for this
// field's input type.
func (w *Widget) IMEContentType() (gpucontext.ContentPurpose, gpucontext.ContentHint) {
	switch w.cfg.inputType {
	case TypePassword:
		return gpucontext.ContentPurposePassword,
			gpucontext.ContentHintHiddenText | gpucontext.ContentHintSensitiveData
	case TypeEmail:
		return gpucontext.ContentPurposeEmail, gpucontext.ContentHintNone
	case TypeNumber:
		// gpucontext deliberately models the semantic purpose separately from
		// presentation hints; there is no digits-only hint in the contract.
		return gpucontext.ContentPurposeNumber, gpucontext.ContentHintNone
	case TypeSearch:
		return gpucontext.ContentPurposeNormal, gpucontext.ContentHintCompletion
	default:
		return gpucontext.ContentPurposeNormal, gpucontext.ContentHintNone
	}
}

// IMESurroundingText returns the committed UTF-8 text and cursor/anchor byte
// offsets. It returns an empty payload whenever IME is disabled or unfocused.
func (w *Widget) IMESurroundingText() gpucontext.IMESurroundingText {
	if !w.IMEEnabled() {
		return gpucontext.IMESurroundingText{}
	}
	text := w.resolvedText()
	runes := []rune(text)
	cursor := clampPos(w.sel.cursor, len(runes))
	anchor := clampPos(w.sel.anchor, len(runes))
	return gpucontext.IMESurroundingText{
		Text:   text,
		Cursor: runeToByteOffset(runes, cursor),
		Anchor: runeToByteOffset(runes, anchor),
	}
}

// IMECursorArea returns the latest caret rectangle in window-local logical
// DIP coordinates. Before the first draw it falls back to the field's content
// origin, which is still a safe candidate anchor for native IMEs.
func (w *Widget) IMECursorArea() gpucontext.IMECursorArea {
	if w.cachedIMEAreaSet {
		return w.cachedIMECursorArea
	}
	bounds := w.Bounds()
	lm := resolveLayoutMetrics(w.painter)
	hPad, vPad := lm.ContentPadding()
	area := geometry.NewRect(bounds.Min.X+hPad, bounds.Min.Y+vPad, lm.TextFieldCursorWidth(), bounds.Height()-2*vPad)
	return w.areaFromLocalRect(area, bounds)
}

func runeToByteOffset(runes []rune, index int) int {
	index = clampPos(index, len(runes))
	return len(string(runes[:index]))
}

// scrollMargin is the horizontal margin in pixels to keep between the cursor
// and the visible edge when scrolling. Prevents the cursor from sitting
// exactly at the boundary, matching Flutter's _kCaretGap behavior.
const scrollMargin float32 = 2

// ensureCursorVisible adjusts scrollOffsetX so the cursor stays within the
// visible content rect. Called during Draw() after layout metrics are resolved.
//
// scrollOffsetX is always <= 0 (text shifts left when overflowing right).
// When text fits entirely within contentRect, scrollOffsetX is clamped to 0.
func (w *Widget) ensureCursorVisible(tm *textmetrics.Metrics, contentRect geometry.Rect, displayText string) {
	// Measure the full text width.
	fullTextWidth := tm.Canvas.MeasureText(displayText, tm.FontSize, false)
	contentWidth := contentRect.Width()

	// If text fits, no scrolling needed.
	if fullTextWidth <= contentWidth {
		w.scrollOffsetX = 0
		return
	}

	// Compute cursor X offset from content origin using MeasureText directly
	// (NOT CursorX, which clamps to contentRect.Max.X and would hide overflow).
	runes := []rune(displayText)
	runePos := w.sel.cursor
	if runePos > len(runes) {
		runePos = len(runes)
	}
	var cursorRelative float32
	if runePos > 0 {
		cursorRelative = tm.Canvas.MeasureText(string(runes[:runePos]), tm.FontSize, false)
	}

	// Apply current scroll offset to get the visual cursor position.
	visualCursorX := cursorRelative + w.scrollOffsetX

	// If cursor is past the right edge, scroll left to reveal it.
	if visualCursorX > contentWidth-scrollMargin {
		w.scrollOffsetX = contentWidth - scrollMargin - cursorRelative
	}

	// If cursor is past the left edge, scroll right to reveal it.
	if visualCursorX < scrollMargin {
		w.scrollOffsetX = scrollMargin - cursorRelative
	}

	// Clamp: never scroll right of origin (would show empty space on left).
	if w.scrollOffsetX > 0 {
		w.scrollOffsetX = 0
	}

	// Clamp: never scroll so far left that right side shows empty space.
	maxScroll := -(fullTextWidth - contentWidth)
	if w.scrollOffsetX < maxScroll {
		w.scrollOffsetX = maxScroll
	}
}

// ScrollOffsetX returns the current horizontal scroll offset.
// This value is always <= 0. A value of 0 means no scrolling.
func (w *Widget) ScrollOffsetX() float32 {
	return w.scrollOffsetX
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
	w.clearComposition()
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
	// Dispose recognizer to release arena references.
	if w.tapDrag != nil {
		w.tapDrag.Dispose()
	}
	// Bindings are cleaned up automatically by WidgetBase.CleanupBindings().
}

// GestureHitTest returns the gesture recognizers for a pointer event at pos.
// Implements [gesture.GestureAware] for the unified pointer pipeline (ADR-049).
// TextField is a leaf widget — always returns recognizers (hit-test already
// confirmed bounds containment).
func (w *Widget) GestureHitTest(_ geometry.Point) []gesture.Recognizer {
	if w.tapDrag == nil {
		return nil
	}
	return []gesture.Recognizer{w.tapDrag}
}

// handleGestureTapDown is the OnTapDown callback for the TapAndDragRecognizer.
// Handles cursor placement (single), word selection (double), and select-all (triple).
func (w *Widget) handleGestureTapDown(details gesture.TapDragDownDetails) {
	if details.Button != event.ButtonLeft {
		return
	}

	// Request focus on any tap.
	// The context is not available here directly, so we defer focus
	// request to the Event handler (MousePress derived from PointerDown
	// still calls RequestFocus). The cursor placement happens immediately.

	runes := w.textRunes()
	pos := positionFromGlobal(w, details.LocalPosition)

	switch details.ConsecutiveTapCount {
	case 1:
		// Single click: place cursor, optionally extend selection with Shift.
		if details.Modifiers.IsShift() {
			w.sel.SetCursorKeepSelection(pos)
		} else {
			w.sel.SetCursor(pos)
		}
	case 2:
		// Double click: select word at position.
		start, end := wordBoundsAt(runes, pos)
		w.sel.anchor = start
		w.sel.cursor = end
		w.gestureHandledTap = true
	default:
		// Triple click (or more): select all text.
		w.sel.selectAll(len(runes))
		w.gestureHandledTap = true
	}

	// ADR-028: visual only.
	w.SetNeedsRedraw(true)
}

// handleGestureDragStart is the OnDragStart callback. Begins selection drag.
func (w *Widget) handleGestureDragStart(_ gesture.TapDragStartDetails) {
	// Drag started — selection extension will happen in handleGestureDragUpdate.
	// No action needed on start; anchor was set in handleGestureTapDown.
}

// handleGestureDragUpdate is the OnDragUpdate callback. Extends selection
// based on the consecutive tap count:
//   - 1 = character-by-character selection
//   - 2 = word-by-word selection
//   - 3 = line-by-line (select-all for single-line TextField)
func (w *Widget) handleGestureDragUpdate(details gesture.TapDragUpdateDetails) {
	runes := w.textRunes()
	pos := positionFromGlobal(w, details.LocalPosition)

	switch details.ConsecutiveTapCount {
	case 1:
		// Character selection: extend cursor without moving anchor.
		w.sel.SetCursorKeepSelection(pos)
	case 2:
		// Word-by-word selection: snap to word boundaries.
		_, end := wordBoundsAt(runes, pos)
		w.sel.SetCursorKeepSelection(end)
	default:
		// Line/all selection: snap to full text.
		w.sel.selectAll(len(runes))
	}

	// ADR-028: visual only.
	w.SetNeedsRedraw(true)
}

// positionFromLocal converts a draw-local position (the coordinate space used
// by derived MouseEvents after Box dispatch translation) to a rune index.
// This is the position relative to the widget's parent, matching the coordinate
// space of Bounds() and cachedContentRect.
func positionFromLocal(w *Widget, localPos geometry.Point) int {
	runes := w.textRunes()

	// Use cached metrics from last Draw if available.
	if w.cachedMetrics != nil {
		adjustedX := localPos.X - w.scrollOffsetX
		return w.cachedMetrics.RuneIndexFromX(
			w.cachedContentRect,
			w.cachedDisplayText,
			adjustedX,
		)
	}

	// Fallback: approximate using layout metrics padding.
	lm := resolveLayoutMetrics(w.painter)
	hPad, _ := lm.ContentPadding()
	bounds := w.Bounds()
	localX := localPos.X - bounds.Min.X - hPad - w.scrollOffsetX

	if localX <= 0 {
		return 0
	}

	fontSize := lm.TextFieldFontSize()
	charW := fontSize * 0.55
	pos := int(localX / charW)
	return clampPos(pos, len(runes))
}

// positionFromGlobal converts a window-coordinate (global) position to a rune
// index. Used by gesture recognizer callbacks where positions are in window
// coordinates (gesture.PointerEvent.GlobalPosition).
//
// Converts global to draw-local by subtracting the widget's ScreenOrigin
// (accumulated parent transforms) and adding Bounds().Min (parent-local offset).
// This produces the same coordinate space as derived MouseEvent positions.
func positionFromGlobal(w *Widget, globalPos geometry.Point) int {
	so := w.ScreenOrigin()
	localPos := globalPos.Sub(so).Add(w.Bounds().Min)
	return positionFromLocal(w, localPos)
}

// Verify Widget implements required interfaces at compile time.
var (
	_ widget.Widget        = (*Widget)(nil)
	_ widget.Focusable     = (*Widget)(nil)
	_ widget.Lifecycle     = (*Widget)(nil)
	_ gesture.GestureAware = (*Widget)(nil)
)
