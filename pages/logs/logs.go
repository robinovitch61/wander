package logs

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
	"wander/components/filter"
	"wander/components/viewport"
	"wander/dev"
	"wander/formatter"
	"wander/keymap"
	"wander/pages"
	"wander/style"
)

type Model struct {
	url, token          string
	logsData            logsData
	width, height       int
	viewport            viewport.Model
	filter              filter.Model
	allocID             string
	taskName            string
	Loading             bool
	LastSelectedLogType LogType
	LastSelectedLogline string
}

const filterPrefix = "Logs"

func New(url, token string, width, height int) Model {
	logsFilter := filter.New(filterPrefix)
	model := Model{
		url:      url,
		token:    token,
		width:    width,
		height:   height,
		viewport: viewport.New(width, height-logsFilter.ViewHeight()),
		filter:   logsFilter,
		Loading:  true,
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
		m.logsData.allData = msg.Data
		m.updateLogViewport()
		m.Loading = false

	case tea.KeyMsg:
		if m.filter.EditingFilter || key.Matches(msg, keymap.KeyMap.Filter) {
			prevFilter := m.filter.Filter
			m.filter, cmd = m.filter.Update(msg)
			if m.filter.Filter != prevFilter {
				m.updateLogViewport()
			}
			cmds = append(cmds, cmd)
		} else {
			switch {
			case key.Matches(msg, keymap.KeyMap.Reload):
				m.Loading = true
				cmds = append(cmds, FetchLogs(m.url, m.token, m.allocID, m.taskName, m.LastSelectedLogType))

			case key.Matches(msg, keymap.KeyMap.Forward):
				if len(m.logsData.filteredData) > 0 {
					m.LastSelectedLogline = string(m.logsData.filteredData[m.viewport.CursorRow])
					return m, pages.ToLoglinePageCmd
				}

			case key.Matches(msg, keymap.KeyMap.Back):
				if len(m.filter.Filter) == 0 {
					return m, pages.ToAllocationsPageCmd
				} else {
					m.ClearFilter()
				}

			case key.Matches(msg, keymap.KeyMap.StdOut):
				m.setLogType(StdOut)
				m.Loading = true
				return m, pages.ToLogsPageCmd

			case key.Matches(msg, keymap.KeyMap.StdErr):
				m.setLogType(StdErr)
				m.Loading = true
				return m, pages.ToLogsPageCmd
			}

			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	content := fmt.Sprintf("Loading %s logs for %s...", m.LastSelectedLogType.ShortString(), m.taskName)
	if !m.Loading {
		if m.LastSelectedLogType == StdOut {
			m.viewport.ContentStyle = style.StdOut
			m.viewport.HeaderStyle = style.StdOut.Copy().Bold(true)
		} else {
			m.viewport.ContentStyle = style.StdErr
			m.viewport.HeaderStyle = style.StdErr.Copy().Bold(true)
		}
		content = m.viewport.View()
	}
	return lipgloss.JoinVertical(lipgloss.Left, m.filter.View(), content)
}

func (m *Model) SetWindowSize(width, height int) {
	m.width, m.height = width, height
	m.viewport.SetSize(width, height-m.filter.ViewHeight())
}

func (m *Model) SetAllocationData(allocID, taskName string) {
	m.allocID, m.taskName = allocID, taskName
	m.filter.SetPrefix(fmt.Sprintf("%s for %s %s", filterPrefix, style.Bold.Render(taskName), formatter.ShortAllocID(allocID)))
}

func (m *Model) ClearFilter() {
	m.filter.SetFilter("")
	m.updateLogViewport()
}

func (m *Model) ResetXOffset() {
	m.viewport.SetXOffset(0)
}

func (m *Model) setLogType(logType LogType) {
	m.LastSelectedLogType = logType
}

func (m *Model) updateFilteredLogData() {
	var filteredLogData []logRow
	for _, entry := range m.logsData.allData {
		if entry.MatchesFilter(m.filter.Filter) {
			filteredLogData = append(filteredLogData, entry)
		}
	}
	m.logsData.filteredData = filteredLogData
}

func (m *Model) updateLogViewport() {
	m.viewport.Highlight = m.filter.Filter
	m.updateFilteredLogData()
	table := logsAsTable(m.logsData.filteredData, m.LastSelectedLogType)
	m.viewport.SetHeaderAndContent(
		strings.Join(table.HeaderRows, "\n"),
		strings.Join(table.ContentRows, "\n"),
	)
	m.viewport.SetCursorRow(len(m.logsData.filteredData) - 1)
}
