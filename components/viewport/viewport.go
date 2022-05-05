package viewport

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
	"wander/dev"
)

// New returns a new model with the given width and height as well as default
// keymappings.
func New(width, height int) (m Model) {
	m.Width = width
	m.Height = height
	m.setInitialValues()
	return m
}

// Model is the Bubble Tea model for this viewport element.
type Model struct {
	Width  int
	Height int
	KeyMap KeyMap

	// Whether or not to respond to the mouse. The mouse must be enabled in
	// Bubble Tea for this to work. For details, see the Bubble Tea docs.
	MouseWheelEnabled bool

	// The number of lines the mouse wheel will scroll. By default, this is 3.
	MouseWheelDelta int

	// YOffset is the vertical scroll position of the text.
	YOffset int

	// CursorRow is the row index of the cursor.
	CursorRow int

	// Style applies a lipgloss style to the viewport. Realistically, it's most
	// useful for setting borders, margins and padding.
	Style lipgloss.Style

	initialized bool
	lines       []string
}

func (m *Model) setInitialValues() {
	m.KeyMap = DefaultKeyMap()
	m.MouseWheelEnabled = true
	m.MouseWheelDelta = 3
	m.Style = lipgloss.NewStyle()
	m.initialized = true
}

// Init exists to satisfy the tea.Model interface for composability purposes.
func (m Model) Init() tea.Cmd {
	return nil
}

// SetContent set the pager's text content.
func (m *Model) SetContent(s string) {
	s = strings.ReplaceAll(s, "\r\n", "\n") // normalize line endings
	m.lines = strings.Split(s, "\n")
}

// maxLinesIdx returns the maximum index of the model's lines
func (m *Model) maxLinesIdx() int {
	return len(m.lines) - 1
}

func (m *Model) lastVisibleLineIdx() int {
	return min(m.maxLinesIdx(), m.YOffset + m.Height - 1)
}

// maxYOffset returns the maximum YOffset (the YOffset that shows the final screen)
func (m *Model) maxYOffset() int {
	if m.maxLinesIdx() < m.Height {
		return 0
	}
	return m.maxLinesIdx() - m.Height + 1
}

// maxSelection returns the maximum CursorRow
func (m *Model) maxSelection() int {
	return len(m.lines) - 1
}

// SetYOffset sets the YOffset with bounds
func (m *Model) SetYOffset(n int) {
	if maxYOffset := m.maxYOffset(); n > maxYOffset {
		m.YOffset = maxYOffset
	} else {
		m.YOffset = max(0, n)
	}
}

// SetCursorRow sets the CursorRow with bounds. Adjusts YOffset as necessary.
func (m *Model) SetCursorRow(n int) {
	if maxSelection := m.maxSelection(); n > maxSelection {
		m.CursorRow = maxSelection
	} else {
		m.CursorRow = max(0, n)
	}

	if lastVisibleLineIdx := m.lastVisibleLineIdx(); m.CursorRow > lastVisibleLineIdx {
		m.viewDown(m.CursorRow - lastVisibleLineIdx)
	} else if m.CursorRow < m.YOffset {
		m.viewUp(m.YOffset - m.CursorRow)
	}
}

// visibleLines retrieves the visible lines based on the YOffset
func (m *Model) visibleLines() []string {
	start := m.YOffset
	end := min(m.maxLinesIdx()+1, m.YOffset+m.Height)
	return m.lines[start:end]
}

// selectionDown moves the CursorRow down by the given number of lines.
func (m *Model) selectionDown(n int) {
	m.SetCursorRow(m.CursorRow + n)
}

// selectionUp moves the CursorRow up by the given number of lines.
func (m *Model) selectionUp(n int) {
	m.SetCursorRow(m.CursorRow - n)
}

// viewDown moves the view down by the given number of lines.
func (m *Model) viewDown(n int) {
	m.SetYOffset(m.YOffset + n)
}

// viewUp moves the view up by the given number of lines. Returns the new
// lines to show.
func (m *Model) viewUp(n int) {
	m.SetYOffset(m.YOffset - n)
}

// Update handles standard message-based viewport updates.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m.updateAsModel(msg)
}

// Author's note: this method has been broken out to make it easier to
// potentially transition Update to satisfy tea.Model.
func (m Model) updateAsModel(msg tea.Msg) (Model, tea.Cmd) {
	if !m.initialized {
		m.setInitialValues()
	}

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Down):
			m.selectionDown(1)

		case key.Matches(msg, m.KeyMap.Up):
			m.selectionUp(1)
		}
		dev.Debug(fmt.Sprintf("selection %d, yoffset %d, height %d, len(m.lines) %d, firstline %s, lastline %s", m.CursorRow, m.YOffset, m.Height, len(m.lines), m.lines[0], m.lines[len(m.lines)-1]))

	case tea.MouseMsg:
		if !m.MouseWheelEnabled {
			break
		}
		switch msg.Type {
		case tea.MouseWheelUp:
			m.viewUp(m.MouseWheelDelta)

		case tea.MouseWheelDown:
			m.viewDown(m.MouseWheelDelta)
		}
	}

	return m, cmd
}

// View renders the viewport into a string.
func (m Model) View() string {
	visibleLines := m.visibleLines()

	viewLines := ""
	for idx, line := range visibleLines {
		if m.YOffset+idx == m.CursorRow {
			viewLines += ">"
		}

		viewLines += line

		if idx != len(visibleLines)-1 {
			viewLines += "\n"
		}
	}

	return m.Style.Copy().
		UnsetWidth().
		UnsetHeight().
		Render(viewLines)
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
