package page

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
	"wander/components/filter"
	"wander/components/viewport"
	"wander/keymap"
	"wander/pages"
)

type Model struct {
	url, token    string
	pageData      pageData
	width, height int
	viewport      viewport.Model
	filter        filter.Model
	filterPrefix  string
	prevPageMsg   pages.ChangePageMsg
	nextPageMsg   pages.ChangePageMsg
	reloadPageMsg pages.ChangePageMsg
	columns       []string
	loadingString string
	Loading       bool
}

func New(
	url, token string,
	width, height int,
	filterPrefix, loadingString string,
	prevPageMsg, nextPageMsg, reloadPageMsg pages.ChangePageMsg,
	columns []string,
) Model {
	pageFilter := filter.New(filterPrefix)
	model := Model{
		url:           url,
		token:         token,
		width:         width,
		height:        height,
		viewport:      viewport.New(width, height-pageFilter.ViewHeight()),
		filter:        pageFilter,
		prevPageMsg:   prevPageMsg,
		nextPageMsg:   nextPageMsg,
		columns:       columns,
		loadingString: loadingString,
		Loading:       true,
	}
	return model
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	if m.viewport.Saving() {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.filter.Focused() {
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
				return m, func() tea.Msg { return m.reloadPageMsg }

			case key.Matches(msg, keymap.KeyMap.Forward):
				return m, func() tea.Msg { return m.nextPageMsg }

			case key.Matches(msg, keymap.KeyMap.Back):
				if len(m.filter.Filter) == 0 {
					return m, pages.ToJobsPageCmd
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
			m.updateViewport()
		}
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	content := fmt.Sprintf(m.loadingString)
	if !m.Loading {
		content = m.viewport.View()
	}
	return lipgloss.JoinVertical(lipgloss.Left, m.filter.View(), content)
}

func (m *Model) SetWindowSize(width, height int) {
	m.width, m.height = width, height
	m.viewport.SetSize(width, height-m.filter.ViewHeight())
}

func (m *Model) clearFilter() {
	m.filter.BlurAndClear()
	m.updateViewport()
}

func (m *Model) updateFilteredData() {
	var filteredData []string
	for _, entry := range m.pageData.allData {
		if strings.Contains(entry, m.filter.Filter) {
			filteredData = append(filteredData, entry)
		}
	}
	m.pageData.filteredData = filteredData
}

func (m *Model) updateViewport() {
	m.viewport.Highlight = m.filter.Filter
	m.updateFilteredData()
	table := dataAsTable(m.pageData.filteredData, m.columns)
	m.viewport.SetHeaderAndContent(
		strings.Join(table.HeaderRows, "\n"),
		strings.Join(table.ContentRows, "\n"),
	)
	m.viewport.SetCursorRow(0)
}
