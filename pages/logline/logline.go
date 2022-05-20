package logline

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"wander/components/filter"
	"wander/components/viewport"
	"wander/dev"
	"wander/formatter"
	"wander/keymap"
	"wander/pages"
	"wander/style"
)

type Model struct {
	logline       string
	loglineData   loglineData
	width, height int
	viewport      viewport.Model
	filter        filter.Model
}

const filterPrefix = "Log Line"

func New(logline string, width, height int) Model {
	splitLoglines := splitLogline(logline)
	loglineFilter := filter.New(filterPrefix)
	loglineViewport := viewport.New(width, height-loglineFilter.ViewHeight())
	loglineViewport.SetCursorEnabled(false)
	model := Model{
		logline:     logline,
		loglineData: loglineData{splitLoglines, splitLoglines},
		width:       width,
		height:      height,
		viewport:    loglineViewport,
		filter:      loglineFilter,
	}
	return model
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	dev.Debug(fmt.Sprintf("logline %T", msg))

	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.viewport.Saving() {
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		} else if m.filter.Focused() {
			switch {
			case key.Matches(msg, keymap.KeyMap.Forward):
				m.filter.Blur()

			case key.Matches(msg, keymap.KeyMap.Back):
				m.clearFilter()
			}
		} else {
			switch {
			case key.Matches(msg, keymap.KeyMap.Filter):
				m.filter.Focus()
				return m, nil

			case key.Matches(msg, keymap.KeyMap.Reload):
				return m, pages.ToAllocationsPageCmd

			case key.Matches(msg, keymap.KeyMap.Back):
				if len(m.filter.Filter) == 0 {
					return m, pages.ToLogsPageCmd
				} else {
					m.clearFilter()
				}
			}

			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}

		// filter won't respond to key messages if not focused
		prevFilter := m.filter.Filter
		m.filter, cmd = m.filter.Update(msg)
		if m.filter.Filter != prevFilter {
			m.updateLoglineViewport()
		}
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, m.filter.View(), m.viewport.View())
}

func (m *Model) SetWindowSize(width, height int) {
	m.width, m.height = width, height
	m.viewport.SetSize(width, height-m.filter.ViewHeight())
}

func (m *Model) clearFilter() {
	m.filter.BlurAndClear()
	m.updateLoglineViewport()
}

func (m *Model) SetLogline(logline string) {
	m.logline = logline
	m.loglineData.allData = splitLogline(logline)
	m.updateLoglineViewport()
}

func (m *Model) SetAllocationData(allocID, taskName string) {
	m.filter.SetPrefix(fmt.Sprintf("%s for %s %s", filterPrefix, style.Bold.Render(taskName), formatter.ShortAllocID(allocID)))
}

func (m *Model) updateFilteredLogData() {
	var filteredLogData []loglineRow
	for _, entry := range m.loglineData.allData {
		if entry.MatchesFilter(m.filter.Filter) {
			filteredLogData = append(filteredLogData, entry)
		}
	}
	m.loglineData.filteredData = filteredLogData
}

func (m *Model) updateLoglineViewport() {
	m.viewport.Highlight = m.filter.Filter
	m.updateFilteredLogData()
	content := logsAsString(m.loglineData.filteredData)
	m.viewport.SetHeaderAndContent("", content)
	m.viewport.SetCursorRow(0)
}

func prettyPrint(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
}

func splitLogline(logline string) []loglineRow {
	pretty, err := prettyPrint([]byte(logline))
	if err != nil {
		return []loglineRow{loglineRow(logline)}
	}

	var splitLoglines []loglineRow
	for _, row := range bytes.Split(pretty, []byte("\n")) {
		splitLoglines = append(splitLoglines, loglineRow(row))
	}

	return splitLoglines
}
