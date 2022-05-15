package viewport

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
	"wander/dev"
	"wander/style"
)

const lineContinuationIndicator = "..."
const lenLineContinuationIndicator = len(lineContinuationIndicator)

type Model struct {
	width         int
	height        int
	contentHeight int // excludes header height, should always be internal
	keyMap        viewportKeyMap

	// Currently, causes flickering if enabled.
	mouseWheelEnabled bool

	// yOffset is the vertical scroll position of the text.
	yOffset int

	// xOffset is the horizontal scroll position of the text.
	xOffset int

	// CursorRow is the row index of the cursor.
	CursorRow int

	// Highlight is the text to highlight (case sensitive), used for search, filter etc.
	Highlight string

	initialized   bool
	header        []string
	lines         []string
	maxLineLength int
}

func New(width, height int) (m Model) {
	m.width = width
	m.height = height
	m.setInitialValues()
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	dev.Debug(fmt.Sprintf("viewport %T", msg))
	if !m.initialized {
		m.setInitialValues()
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Up):
			m.cursorRowUp(1)

		case key.Matches(msg, m.keyMap.Down):
			m.cursorRowDown(1)

		case key.Matches(msg, m.keyMap.Left):
			m.viewLeft(m.width / 4)

		case key.Matches(msg, m.keyMap.Right):
			m.viewRight(m.width / 4)

		case key.Matches(msg, m.keyMap.HalfPageUp):
			m.viewUp(m.contentHeight / 2)
			m.cursorRowUp(m.contentHeight / 2)

		case key.Matches(msg, m.keyMap.HalfPageDown):
			m.viewDown(m.contentHeight / 2)
			m.cursorRowDown(m.contentHeight / 2)

		case key.Matches(msg, m.keyMap.PageUp):
			m.viewUp(m.contentHeight)
			m.cursorRowUp(m.contentHeight)

		case key.Matches(msg, m.keyMap.PageDown):
			m.viewDown(m.contentHeight)
			m.cursorRowDown(m.contentHeight)

		case key.Matches(msg, m.keyMap.Top):
			m.cursorRowUp(m.yOffset + m.CursorRow)

		case key.Matches(msg, m.keyMap.Bottom):
			yOffset := m.maxLinesIdx()
			m.cursorRowDown(yOffset)
		}

		//dev.Debug(fmt.Sprintf("selection %d, yoffset %d, height %d, contentHeight %d, len(m.lines) %d", m.CursorRow, m.yOffset, m.height, m.contentHeight, len(m.lines)))

	case tea.MouseMsg:
		if !m.mouseWheelEnabled {
			break
		}
		switch msg.Type {
		case tea.MouseWheelUp:
			m.cursorRowUp(1)

		case tea.MouseWheelDown:
			m.cursorRowDown(1)
		}
	}

	// could return non-nil cmd in the future
	return m, nil
}

func (m Model) View() string {
	nothingHighlighted := len(m.Highlight) == 0
	var viewLines string

	for _, headerLine := range m.header {
		viewLines += style.TableHeaderStyle.Render(m.getVisiblePartOfLine(headerLine)) + "\n"
	}

	for idx, line := range m.visibleLines() {
		isSelected := m.yOffset+idx == m.CursorRow
		line = m.getVisiblePartOfLine(line)

		if nothingHighlighted {
			if isSelected {
				viewLines += style.ViewportCursorRow.Render(line)
			} else {
				viewLines += line
			}
		} else {
			// this splitting and rejoining of styled lines is expensive and causes increased flickering,
			// so only do it if something is actually highlighted
			styledHighlight := style.ViewportHighlight.Render(m.Highlight)
			if isSelected {
				lineChunks := strings.Split(line, m.Highlight)
				var styledChunks []string
				for _, chunk := range lineChunks {
					styledChunks = append(styledChunks, style.ViewportCursorRow.Render(chunk))
				}
				viewLines += strings.Join(styledChunks, styledHighlight)
			} else {
				line = strings.ReplaceAll(line, m.Highlight, styledHighlight)
				viewLines += line
			}
		}

		viewLines += "\n"
	}
	trimmedViewLines := strings.TrimSpace(viewLines)
	renderedViewLines := style.Viewport.Width(m.width).Height(m.height).Render(trimmedViewLines)
	return renderedViewLines
}

func (m Model) ContentEmpty() bool {
	return len(m.header) == 0 && len(m.lines) == 0
}

// SetSize sets the viewport's width and height, including header.
func (m *Model) SetSize(width, height int) {
	m.width, m.height = width, height
	m.setContentHeight()
	m.fixState()
}

// SetHeader sets the pager's header content.
func (m *Model) SetHeader(header string) {
	m.header = strings.Split(normalizeLineEndings(header), "\n")
	m.setContentHeight()
}

// SetContent sets the pager's text content.
func (m *Model) SetContent(s string) {
	lines := strings.Split(normalizeLineEndings(s), "\n")
	maxLineLength := 0
	for _, line := range lines {
		if lineLength := len(strings.TrimSpace(line)); lineLength > maxLineLength {
			maxLineLength = lineLength
		}
	}
	m.lines = lines
	m.maxLineLength = maxLineLength
	m.fixState()
}

// SetCursorRow sets the CursorRow with bounds. Adjusts yOffset as necessary.
func (m *Model) SetCursorRow(n int) {
	if m.contentHeight == 0 {
		return
	}

	if maxSelection := m.maxCursorRow(); n > maxSelection {
		m.CursorRow = maxSelection
	} else {
		m.CursorRow = max(0, n)
	}

	if lastVisibleLineIdx := m.lastVisibleLineIdx(); m.CursorRow > lastVisibleLineIdx {
		m.viewDown(m.CursorRow - lastVisibleLineIdx)
	} else if m.CursorRow < m.yOffset {
		m.viewUp(m.yOffset - m.CursorRow)
	}
}

func (m *Model) setInitialValues() {
	m.setContentHeight()
	m.keyMap = GetKeyMap()
	m.mouseWheelEnabled = false
	m.initialized = true
}

func normalizeLineEndings(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}

// fixCursorRow adjusts the cursor to be in a visible location if it is outside the visible content
func (m *Model) fixCursorRow() {
	if m.CursorRow > m.lastVisibleLineIdx() {
		m.SetCursorRow(m.lastVisibleLineIdx())
	}
}

// fixYOffset adjusts the yOffset such that it's not above the maximum value
func (m *Model) fixYOffset() {
	if maxYOffset := m.maxYOffset(); m.yOffset > maxYOffset {
		m.setYOffset(maxYOffset)
	}
}

// fixState fixes CursorRow and yOffset
func (m *Model) fixState() {
	m.fixYOffset()
	m.fixCursorRow()
}

func (m *Model) setContentHeight() {
	m.contentHeight = max(0, m.height-len(m.header))
}

// maxLinesIdx returns the maximum index of the model's lines
func (m *Model) maxLinesIdx() int {
	return len(m.lines) - 1
}

// lastVisibleLineIdx returns the maximum visible line index
func (m Model) lastVisibleLineIdx() int {
	return min(m.maxLinesIdx(), m.yOffset+m.contentHeight-1)
}

// maxYOffset returns the maximum yOffset (the yOffset that shows the final screen)
func (m Model) maxYOffset() int {
	if m.maxLinesIdx() < m.contentHeight {
		return 0
	}
	return len(m.lines) - m.contentHeight
}

// maxCursorRow returns the maximum CursorRow
func (m Model) maxCursorRow() int {
	return len(m.lines) - 1
}

// setYOffset sets the yOffset with bounds.
func (m *Model) setYOffset(n int) {
	if maxYOffset := m.maxYOffset(); n > maxYOffset {
		m.yOffset = maxYOffset
	} else {
		m.yOffset = max(0, n)
	}
}

// visibleLines retrieves the visible lines based on the yOffset
func (m Model) visibleLines() []string {
	start := m.yOffset
	end := start + m.contentHeight
	if end > m.maxLinesIdx() {
		return m.lines[start:]
	}
	return m.lines[start:end]
}

// cursorRowDown moves the CursorRow down by the given number of lines.
func (m *Model) cursorRowDown(n int) {
	m.SetCursorRow(m.CursorRow + n)
}

// cursorRowUp moves the CursorRow up by the given number of lines.
func (m *Model) cursorRowUp(n int) {
	m.SetCursorRow(m.CursorRow - n)
}

// viewDown moves the view down by the given number of lines.
func (m *Model) viewDown(n int) {
	m.setYOffset(m.yOffset + n)
}

// viewUp moves the view up by the given number of lines.
func (m *Model) viewUp(n int) {
	m.setYOffset(m.yOffset - n)
}

func (m *Model) setXOffset(n int) {
	m.xOffset = max(0, n)
}

// viewLeft moves the view left the given number of columns.
func (m *Model) viewLeft(n int) {
	m.setXOffset(m.xOffset - n)
}

// viewRight moves the view right the given number of columns.
func (m *Model) viewRight(n int) {
	m.setXOffset(min(m.maxLineLength-m.width, m.xOffset+n))
}

func (m Model) getVisiblePartOfLine(line string) string {
	trimmedLineLength := len(strings.TrimSpace(line))
	if len(line) > m.width {
		//dev.Debug(fmt.Sprintf("len(line) %d, m.xOffset %d, m.width %d", len(line), m.xOffset, m.width))
		line = line[m.xOffset:min(len(line), m.xOffset+m.width)]
		if m.xOffset+m.width+1 < trimmedLineLength {
			line = line[:len(line)-lenLineContinuationIndicator] + lineContinuationIndicator
		}
		if m.xOffset > 0 {
			line = lineContinuationIndicator + line[lenLineContinuationIndicator:]
		}
	}
	return line
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
