package page

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
	"wander/components/filter"
	"wander/components/toast"
	"wander/components/viewport"
	"wander/dev"
	"wander/keymap"
)

type Model struct {
	width, height     int
	pageData          data
	viewport          viewport.Model
	filter            filter.Model
	loadingString     string
	loading           bool
	isTerminal        bool
	prompt            textinput.Model
	promptInitialized bool
}

func New(
	width, height int,
	filterPrefix, loadingString string,
	cursorEnabled, wrapText, isTerminal bool,
) Model {
	pageFilter := filter.New(filterPrefix)
	viewportHeight := height - pageFilter.ViewHeight()
	if isTerminal {
		viewportHeight = 0
	}
	pageViewport := viewport.New(width, viewportHeight)
	pageViewport.SetCursorEnabled(cursorEnabled)
	pageViewport.SetWrapText(wrapText)
	model := Model{
		width:         width,
		height:        height,
		viewport:      pageViewport,
		filter:        pageFilter,
		loadingString: loadingString,
		loading:       true,
		isTerminal:    isTerminal,
	}
	return model
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	dev.Debug(fmt.Sprintf("page %T", msg))
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	if m.isTerminal {
		if !m.promptInitialized {
			m.prompt = textinput.New()
			m.prompt.Focus()
			cmds = append(cmds, textinput.Blink)
			m.promptInitialized = true
		}

		if !m.ViewportSaving() {
			m.prompt, cmd = m.prompt.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	if m.viewport.Saving() {
		m.prompt.Blur()
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case viewport.SaveStatusMsg:
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)

	case toast.ToastTimeoutMsg:
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)

	case tea.KeyMsg:
		if key.Matches(msg, keymap.KeyMap.Forward) && m.isTerminal {
			shellCmd := m.prompt.Value()
			m.prompt.Reset()
			dev.Debug(shellCmd)
			// TODO LEO: return cmd here that does the websocket async
		}

		if key.Matches(msg, keymap.KeyMap.Back) {
			m.clearFilter()
		}

		if m.filter.Focused() {
			switch {
			case key.Matches(msg, keymap.KeyMap.Forward):
				m.filter.Blur()
			}
		} else {
			switch {
			case key.Matches(msg, keymap.KeyMap.Filter):
				m.filter.Focus()
				return m, nil
			}

			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}

		prevFilter := m.filter.Filter
		m.filter, cmd = m.filter.Update(msg)
		if m.filter.Filter != prevFilter {
			m.updateViewport()
		}
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	blocks := []string{m.filter.View()}
	content := fmt.Sprintf(m.loadingString)
	if !m.loading {
		content = m.viewport.View()
	}
	blocks = append(blocks, content)
	if m.isTerminal && !m.loading {
		blocks = append(blocks, m.prompt.View())
	}
	return lipgloss.JoinVertical(lipgloss.Left, blocks...)
}

func (m *Model) SetWindowSize(width, height int) {
	m.width, m.height = width, height
	m.viewport.SetSize(width, height-m.filter.ViewHeight())
}

func (m *Model) SetHeader(header []string) {
	m.viewport.SetHeader(header)
}

func (m *Model) SetViewportStyle(headerStyle, contentStyle lipgloss.Style) {
	m.viewport.HeaderStyle = headerStyle
	m.viewport.ContentStyle = contentStyle
}

func (m *Model) SetLoading(isLoading bool) {
	m.loading = isLoading
}

func (m *Model) SetAllPageData(allPageData []Row) {
	m.pageData.All = allPageData
	m.updateViewport()
}

func (m *Model) SetFilterPrefix(prefix string) {
	m.filter.SetPrefix(prefix)
}

func (m *Model) SetViewportCursorToBottom() {
	m.viewport.SetCursorRow(len(m.pageData.Filtered) - 1)
}

func (m *Model) SetViewportXOffset(n int) {
	m.viewport.SetXOffset(n)
}

func (m *Model) ExitTerminal() {
	if m.isTerminal {
		m.promptInitialized = false
	}
}

func (m Model) Loading() bool {
	return m.loading
}

func (m Model) GetSelectedPageRow() (Row, error) {
	cursorRow := m.viewport.CursorRow()
	if filtered := m.pageData.Filtered; len(filtered) > 0 && cursorRow >= 0 && cursorRow < len(filtered) {
		return filtered[cursorRow], nil
	}
	return Row{}, fmt.Errorf("bad thing")
}

func (m Model) FilterFocused() bool {
	return m.filter.Focused()
}

func (m Model) FilterApplied() bool {
	return len(m.filter.Filter) > 0
}

func (m Model) ViewportSaving() bool {
	return m.viewport.Saving()
}

func (m Model) IsTerminal() bool {
	return m.isTerminal
}

func (m *Model) clearFilter() {
	m.filter.BlurAndClear()
	m.updateViewport()
}

func (m *Model) updateViewport() {
	m.viewport.Highlight = m.filter.Filter
	m.updateFilteredData()
	m.viewport.SetContent(rowsToStrings(m.pageData.Filtered))
	m.viewport.SetCursorRow(0)
}

func (m *Model) updateFilteredData() {
	if m.filter.Filter == "" {
		m.pageData.Filtered = m.pageData.All
	} else {
		var filteredData []Row
		for _, entry := range m.pageData.All {
			if strings.Contains(entry.Row, m.filter.Filter) {
				filteredData = append(filteredData, entry)
			}
		}
		m.pageData.Filtered = filteredData
	}
}
