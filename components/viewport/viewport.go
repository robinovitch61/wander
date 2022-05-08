package viewport

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

// New returns a new model with the given width and height as well as default
// keymappings.
func New(width, height int) (m Model) {
	m.width = width
	m.height = height
	m.setInitialValues()
	return m
}

// Model is the Bubble Tea model for this viewport element.
type Model struct {
	width         int
	height        int
	contentHeight int
	keyMap        viewportKeyMap

	// Whether to respond to the mouse. The mouse must be enabled in
	// Bubble Tea for this to work. For details, see the Bubble Tea docs.
	// Currently, causes flickering if enabled.
	mouseWheelEnabled bool

	// yOffset is the vertical scroll position of the text.
	yOffset int

	// cursorRow is the row index of the cursor.
	cursorRow int

	// styleViewport applies a lipgloss style to the viewport. Realistically, it's most
	// useful for setting borders, margins and padding.
	styleViewport  lipgloss.Style
	styleCursorRow lipgloss.Style

	initialized bool
	header      []string
	lines       []string
}

func (m *Model) setInitialValues() {
	m.contentHeight = m.height - len(m.header)
	m.keyMap = getKeyMap()
	m.mouseWheelEnabled = false
	m.styleViewport = lipgloss.NewStyle().Background(lipgloss.Color("#000000"))
	m.styleCursorRow = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	m.initialized = true
}

// Init exists to satisfy the tea.Model interface for composability purposes.
func (m Model) Init() tea.Cmd {
	return nil
}

func normalizeLineEndings(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}

// SetHeight sets the pager's height, including header.
func (m *Model) SetHeight(h int) {
	m.height = h
	m.contentHeight = h - len(m.header)
	if m.cursorRow > m.lastVisibleLineIdx() {
		m.cursorRow = m.lastVisibleLineIdx()
	}
}

// SetWidth sets the pager's width.
func (m *Model) SetWidth(w int) {
	m.width = w
}

// SetHeader sets the pager's header content.
func (m *Model) SetHeader(header string) {
	m.header = strings.Split(normalizeLineEndings(header), "\n")
	m.contentHeight = m.height - len(m.header)
}

// SetContent sets the pager's text content.
func (m *Model) SetContent(s string) {
	m.lines = strings.Split(normalizeLineEndings(s), "\n")
}

// maxLinesIdx returns the maximum index of the model's lines
func (m *Model) maxLinesIdx() int {
	return len(m.lines) - 1
}

// lastVisibleLineIdx returns the maximum visible line index
func (m *Model) lastVisibleLineIdx() int {
	return min(m.maxLinesIdx(), m.yOffset+m.contentHeight-1)
}

// maxYOffset returns the maximum yOffset (the yOffset that shows the final screen)
func (m *Model) maxYOffset() int {
	if m.maxLinesIdx() < m.contentHeight {
		return 0
	}
	return m.maxLinesIdx() - m.contentHeight + 1
}

// maxCursorRow returns the maximum cursorRow
func (m *Model) maxCursorRow() int {
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

// setCursorRow sets the cursorRow with bounds. Adjusts yOffset as necessary.
func (m *Model) setCursorRow(n int) {
	if maxSelection := m.maxCursorRow(); n > maxSelection {
		m.cursorRow = maxSelection
	} else {
		m.cursorRow = max(0, n)
	}

	if lastVisibleLineIdx := m.lastVisibleLineIdx(); m.cursorRow > lastVisibleLineIdx {
		m.viewDown(m.cursorRow - lastVisibleLineIdx)
	} else if m.cursorRow < m.yOffset {
		m.viewUp(m.yOffset - m.cursorRow)
	}
}

// visibleLines retrieves the visible lines based on the yOffset
func (m *Model) visibleLines() []string {
	start := m.yOffset
	end := start + m.contentHeight
	if end > m.maxLinesIdx() {
		return m.lines[start:]
	}
	return m.lines[start:end]
}

// cursorRowDown moves the cursorRow down by the given number of lines.
func (m *Model) cursorRowDown(n int) {
	m.setCursorRow(m.cursorRow + n)
}

// cursorRowUp moves the cursorRow up by the given number of lines.
func (m *Model) cursorRowUp(n int) {
	m.setCursorRow(m.cursorRow - n)
}

// viewDown moves the view down by the given number of lines.
func (m *Model) viewDown(n int) {
	m.setYOffset(m.yOffset + n)
}

// viewUp moves the view up by the given number of lines.
func (m *Model) viewUp(n int) {
	m.setYOffset(m.yOffset - n)
}

// Update handles standard message-based viewport updates.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.initialized {
		m.setInitialValues()
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Down):
			m.cursorRowDown(1)

		case key.Matches(msg, m.keyMap.Up):
			m.cursorRowUp(1)

		case key.Matches(msg, m.keyMap.HalfPageDown):
			m.viewDown(m.contentHeight / 2)
			m.cursorRowDown(m.contentHeight / 2)

		case key.Matches(msg, m.keyMap.HalfPageUp):
			m.viewUp(m.contentHeight / 2)
			m.cursorRowUp(m.contentHeight / 2)

		case key.Matches(msg, m.keyMap.PageDown):
			m.viewDown(m.contentHeight)
			m.cursorRowDown(m.contentHeight)

		case key.Matches(msg, m.keyMap.PageUp):
			m.viewUp(m.contentHeight)
			m.cursorRowUp(m.contentHeight)
		}
		//dev.Debug(fmt.Sprintf("selection %d, yoffset %d, height %d, contentHeight %d, len(m.lines) %d", m.cursorRow, m.yOffset, m.height, m.contentHeight, len(m.lines)))

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

// View returns the string representing the viewport.
func (m Model) View() string {
	visibleLines := m.visibleLines()

	viewLines := strings.Join(m.header, "\n") + "\n"
	for idx, line := range visibleLines {
		if m.yOffset+idx == m.cursorRow {
			viewLines += m.styleCursorRow.Render(line)
		} else {
			viewLines += line
		}

		if idx != len(visibleLines)-1 {
			viewLines += "\n"
		}
	}
	return m.styleViewport.Width(m.width).Height(m.height).Render(viewLines) // width and height are variable
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
