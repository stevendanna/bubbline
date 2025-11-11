package textarea

import (
	"fmt"
	"strings"
)

// EmptyValue returns true iff the value is empty.
func (m *Model) EmptyValue() bool {
	return len(m.value) == 0 || (len(m.value) == 1 && len(m.value[0]) == 0)
}

// NumLinesInValue returns the number of logical lines in the value.
func (m *Model) NumLinesInValue() int {
	return len(m.value)
}

// CursorPos retrieves the position of the cursor inside the input.
func (m *Model) CursorPos() int {
	return m.col
}

// CursorRight moves the cursor to the right by the specified amount.
func (m *Model) CursorRight(n int) {
	for i := 0; i < n; i++ {
		m.characterRight()
	}
}

// CurrentLine retrieves the current line as a string.
func (m *Model) CurrentLine() string {
	return string(m.value[m.row])
}

// ClearLine clears the current line.
func (m *Model) ClearLine() {
	m.value[m.row] = m.value[m.row][:0]
	m.col = 0
	m.lastCharOffset = 0
}

// InsertNewline inserts a newline character at the cursor.
func (m *Model) InsertNewline() {
	if m.MaxHeight > 0 && len(m.value) >= m.MaxHeight {
		return
	}
	m.col = clamp(m.col, 0, len(m.value[m.row]))
	m.splitLine(m.row, m.col)
}

// DeleteCharacterForward deletes the character at the cursor.
func (m *Model) DeleteCharacterForward() {
	if len(m.value[m.row]) > 0 && m.col < len(m.value[m.row]) {
		m.value[m.row] = append(m.value[m.row][:m.col], m.value[m.row][m.col+1:]...)
	}
	if m.col >= len(m.value[m.row]) {
		m.mergeLineBelow(m.row)
	}
}

// DeleteCharactersBackward deletes n characters before the cursor.
func (m *Model) DeleteCharactersBackward(n int) {
	for n > 0 {
		m.col = clamp(m.col, 0, len(m.value[m.row]))
		if m.col <= 0 {
			m.mergeLineAbove(m.row)
			n--
			continue
		}
		if len(m.value[m.row]) > 0 {
			d := n
			if d > len(m.value[m.row]) {
				d = len(m.value[m.row])
			}
			m.value[m.row] = append(m.value[m.row][:max(0, m.col-d)], m.value[m.row][m.col:]...)
			if m.col > 0 {
				m.SetCursor(m.col - d)
			}
			n -= d
		}
	}
}

// MoveTo moves the cursor to the specified position.
func (m *Model) MoveTo(row, col int) {
	m.row = clamp(row, 0, len(m.value)-1)
	m.SetCursor(clamp(col, 0, len(m.value[m.row])))
}

// ValueRunes retrieves the current value decomposed as runes.
func (m *Model) ValueRunes() [][]rune {
	return m.value
}

// AtBeginningOfLine returns true if the cursor is at the beginning of
// an empty line line.
func (m *Model) AtBeginningOfEmptyLine() bool {
	return m.col == 0 && len(m.value[m.row]) == 0
}

// AtFirstLineOfInputAndView returns true if the cursor is on the first line
// of the input and viewport.
func (m *Model) AtFirstLineOfInputAndView() bool {
	li := m.LineInfo()
	return m.row == 0 && li.RowOffset == 0
}

// AtEndOfInput returns true if the cursor is on the last line of the input and viewport.
func (m *Model) AtLastLineOfInputAndView() bool {
	li := m.LineInfo()
	return m.row >= len(m.value)-1 && li.RowOffset == li.Height-1
}

// ResetViewCursorDown scrolls the viewport so that the cursor
// is position on the bottom line.
func (m Model) ResetViewCursorDown() {
	row := m.cursorLineNumber()
	m.viewport.SetYOffset(row - m.viewport.Height + 1)
}

// LogicalHeight returns the number of lines needed in a viewport to
// show the entire value.
func (m Model) LogicalHeight() int {
	logicalHeight := 0
	nl := m.LineCount()
	for row := 0; row < nl; row++ {
		li := m.LineInfoAt(row, 0)
		logicalHeight += li.Height
	}
	return logicalHeight
}

// overwriteRune overwrites the rune at the cursor position.
func (m *Model) overwriteRune(r rune) {
	// If we're at the end of the line, or if the input rune is a
	// newline, simply insert it.  Otherwise, overwrite.
	if r == '\n' || r == '\r' || (m.col >= len(m.value[m.row])) {
		m.InsertRune(r)
		return
	}
	m.value[m.row][m.col] = r
	m.SetCursor(m.col + 1)
}

// Debug returns debug details about the state of the model.
func (m Model) Debug() string {
	var buf strings.Builder
	fmt.Fprintf(&buf, "focus: %v\n", m.focus)
	fmt.Fprintf(&buf, "promptWidth: %d\n", m.promptWidth)
	fmt.Fprintf(&buf, "width: %d, height: %d\n", m.width, m.height)
	fmt.Fprintf(&buf, "col: %d, row: %d\n", m.col, m.row)
	fmt.Fprintf(&buf, "lastCharOffset: %d\n", m.lastCharOffset)
	for l, line := range m.value {
		fmt.Fprintf(&buf, "line %d: %v (%q)\n", l, line, string(line))
	}
	return buf.String()
}
