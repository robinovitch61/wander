package jobs

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"wander/components/filter"
	"wander/components/viewport"
	"wander/dev"
	"wander/keymap"
	"wander/pages"
)

type Model struct {
	initialized bool
	url, token  string
	// jobsData          jobsData
	width, height     int
	viewport          viewport.Model
	filter            filter.Model
	Loading           bool
	LastSelectedJobID string
}

func New(url, token string, width, height int) Model {
	jobsFilter := filter.New("Jobs")
	model := Model{
		url:      url,
		token:    token,
		width:    width,
		height:   height,
		viewport: viewport.New(width, height-jobsFilter.ViewHeight()),
		filter:   jobsFilter,
		Loading:  true,
	}
	return model
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	dev.Debug(fmt.Sprintf("jobs %T", msg))

	if !m.initialized {
		m.initialized = true
		return m, FetchJobs(m.url, m.token)
	}

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
	case NomadJobsMsg:
		// m.jobsData.allData = msg
		m.updateJobViewport()
		m.Loading = false

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
				return m, pages.ToJobsPageCmd

			// case key.Matches(msg, keymap.KeyMap.Forward):
			// 	if len(m.jobsData.filteredData) > 0 {
			// 		m.LastSelectedJobID = m.jobsData.filteredData[m.viewport.CursorRow].ID
			// 		return m, pages.ToAllocationsPageCmd
			// 	}

			case key.Matches(msg, keymap.KeyMap.Back):
				m.clearFilter()
			}

			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}

		// filter won't respond to key messages if not focused
		prevFilter := m.filter.Filter
		m.filter, cmd = m.filter.Update(msg)
		if m.filter.Filter != prevFilter {
			m.updateJobViewport()
		}
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	content := "loading jobs..."
	if !m.Loading {
		content = m.viewport.View()
	}
	return lipgloss.JoinVertical(lipgloss.Left, m.filter.View(), content)
}

func (m Model) FilterFocused() bool {
	return m.filter.Focused()
}

func (m *Model) SetWindowSize(width, height int) {
	m.width, m.height = width, height
	m.viewport.SetSize(width, height-m.filter.ViewHeight())
}

func (m *Model) clearFilter() {
	m.filter.BlurAndClear()
	m.updateJobViewport()
}

// func (m *Model) updateFilteredJobData() {
// 	var filteredJobData []JobResponseEntry
// 	for _, entry := range m.jobsData.allData {
// 		if entry.MatchesFilter(m.filter.Filter) {
// 			filteredJobData = append(filteredJobData, entry)
// 		}
// 	}
// 	m.jobsData.filteredData = filteredJobData
// }

func (m *Model) updateJobViewport() {
	m.viewport.Highlight = m.filter.Filter
	// m.updateFilteredJobData()
	// table := JobResponsesAsTable(m.jobsData.filteredData)
	// m.viewport.SetHeaderAndContent(
	// 	strings.Join(table.HeaderRows, "\n"),
	// 	strings.Join(table.ContentRows, "\n"),
	// )
	m.viewport.SetCursorRow(0)
}
