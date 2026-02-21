package textfield

import "unicode"

// selection tracks cursor position and text selection state.
type selection struct {
	cursor    int    // cursor position (byte offset in runes slice)
	anchor    int    // anchor for selection (same as cursor when no selection)
	clipboard string // internal clipboard (placeholder for platform clipboard)
}

// HasSelection returns true if text is selected (anchor != cursor).
func (s *selection) HasSelection() bool {
	return s.anchor != s.cursor
}

// OrderedRange returns the selection range as (start, end) where start <= end.
func (s *selection) OrderedRange() (int, int) {
	if s.anchor <= s.cursor {
		return s.anchor, s.cursor
	}
	return s.cursor, s.anchor
}

// ClearSelection collapses the selection to the cursor position.
func (s *selection) ClearSelection() {
	s.anchor = s.cursor
}

// SelectAll selects all text given the total rune count.
func (s *selection) SelectAll(runeCount int) {
	s.anchor = 0
	s.cursor = runeCount
}

// SetCursor moves the cursor to the given position and clears selection.
func (s *selection) SetCursor(pos int) {
	s.cursor = pos
	s.anchor = pos
}

// SetCursorKeepSelection moves the cursor without clearing the anchor,
// extending or shrinking the selection.
func (s *selection) SetCursorKeepSelection(pos int) {
	s.cursor = pos
}

// clampPos ensures a position is within [0, maxPos].
func clampPos(pos, maxPos int) int {
	if pos < 0 {
		return 0
	}
	if pos > maxPos {
		return maxPos
	}
	return pos
}

// nextWordBoundary finds the next word boundary after pos in the given runes.
func nextWordBoundary(runes []rune, pos int) int {
	n := len(runes)
	if pos >= n {
		return n
	}
	// Skip current word characters.
	i := pos
	for i < n && isWordChar(runes[i]) {
		i++
	}
	// If we didn't move, skip non-word characters.
	if i == pos {
		for i < n && !isWordChar(runes[i]) {
			i++
		}
	}
	return i
}

// prevWordBoundary finds the previous word boundary before pos in the given runes.
func prevWordBoundary(runes []rune, pos int) int {
	if pos <= 0 {
		return 0
	}
	i := pos
	// Skip non-word characters backwards.
	for i > 0 && !isWordChar(runes[i-1]) {
		i--
	}
	// Skip word characters backwards.
	for i > 0 && isWordChar(runes[i-1]) {
		i--
	}
	return i
}

// isWordChar returns true if the rune is a letter, digit, or underscore.
func isWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

// wordBoundsAt returns the (start, end) of the word at the given position.
// Used for double-click word selection.
func wordBoundsAt(runes []rune, pos int) (int, int) {
	n := len(runes)
	if n == 0 {
		return 0, 0
	}
	// Clamp position.
	if pos >= n {
		pos = n - 1
	}
	if pos < 0 {
		pos = 0
	}

	// If at a non-word char, select the run of non-word chars.
	if !isWordChar(runes[pos]) {
		start := pos
		for start > 0 && !isWordChar(runes[start-1]) {
			start--
		}
		end := pos + 1
		for end < n && !isWordChar(runes[end]) {
			end++
		}
		return start, end
	}

	// Otherwise select the word.
	start := pos
	for start > 0 && isWordChar(runes[start-1]) {
		start--
	}
	end := pos + 1
	for end < n && isWordChar(runes[end]) {
		end++
	}
	return start, end
}

// copyToClipboard stores the given text in the internal clipboard.
// This is a placeholder; a real implementation would use platform APIs.
func (s *selection) copyToClipboard(text string) {
	s.clipboard = text
}

// pasteFromClipboard returns the text from the internal clipboard.
// This is a placeholder; a real implementation would use platform APIs.
func (s *selection) pasteFromClipboard() string {
	return s.clipboard
}
