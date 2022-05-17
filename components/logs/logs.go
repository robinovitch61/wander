package logs

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
	"wander/components/filter"
	"wander/components/page"
	"wander/components/viewport"
	"wander/dev"
	"wander/message"
)

type Model struct {
	url, token    string
	nomadLogsData nomadLogsData
	width, height int
	viewport      viewport.Model
	filter        filter.Model
	keyMap        page.KeyMap
	loading       bool
	allocID       string
	taskName      string
	logType       LogType
}

func New(url, token string, width, height int) Model {
	logsFilter := filter.New("Logs")
	model := Model{
		url:      url,
		token:    token,
		width:    width,
		height:   height,
		viewport: viewport.New(width, height-logsFilter.ViewHeight()),
		filter:   logsFilter,
		keyMap:   page.GetKeyMap(),
		loading:  true,
	}
	return model
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	dev.Debug(fmt.Sprintf("logs %T", msg))

	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {

	case nomadLogsMsg:
		m.nomadLogsData.allData = msg.Data
		m.updateLogViewport()
		m.loading = false

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Reload):
			m.loading = true
			cmds = append(cmds, FetchLogs(m.url, m.token, m.allocID, m.taskName, m.logType))

		case key.Matches(msg, m.keyMap.Back):
			if !m.filter.EditingFilter {
				return m, func() tea.Msg { return message.ViewAllocationsMsg{} }
			}
		}

		prevFilter := m.filter.Filter
		m.filter, cmd = m.filter.Update(msg)
		if m.filter.Filter != prevFilter {
			m.updateLogViewport()
		}
		cmds = append(cmds, cmd)

		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	content := fmt.Sprintf("Loading logs for %s...", m.taskName)
	if !m.loading {
		content = m.viewport.View()
	}
	return lipgloss.JoinVertical(lipgloss.Left, m.filter.View(), content)
}

func (m *Model) SetWindowSize(width, height int) {
	m.width, m.height = width, height
	m.viewport.SetSize(width, height-m.filter.ViewHeight())
}

func (m *Model) SetAllocationData(allocId, taskName string) {
	m.allocID, m.taskName = allocId, taskName
}

func (m *Model) SetLogType(logType LogType) {
	m.logType = logType
}

func (m *Model) ClearFilter() {
	m.filter.SetFilter("")
	m.updateLogViewport()
}

func (m *Model) updateFilteredLogData() {
	var filteredLogData []logRow
	for _, entry := range m.nomadLogsData.allData {
		if entry.MatchesFilter(m.filter.Filter) {
			filteredLogData = append(filteredLogData, entry)
		}
	}
	m.nomadLogsData.filteredData = filteredLogData
}

func (m *Model) updateLogViewport() {
	m.updateFilteredLogData()
	table := logsAsTable(m.nomadLogsData.filteredData, m.logType)
	m.viewport.SetHeaderAndContent(
		strings.Join(table.HeaderRows, "\n"),
		strings.Join(table.ContentRows, "\n"),
	)
	m.viewport.SetCursorRow(0)
}
