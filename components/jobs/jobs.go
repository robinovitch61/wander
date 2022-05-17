package jobs

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
)

type nomadJobData struct {
	allData, filteredData []jobResponseEntry
}

type NomadJobsMsg []jobResponseEntry

type Model struct {
	initialized   bool
	url, token    string
	nomadJobData  nomadJobData
	width, height int
	viewport      viewport.Model
	filter        filter.Model
	keyMap        page.KeyMap
	loading       bool
	SelectedJobId string
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
		keyMap:   page.GetKeyMap(),
		loading:  true,
	}
	return model
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	dev.Debug(fmt.Sprintf("jobs %T", msg))

	if !m.initialized {
		m.initialized = true
		return m, fetchJobs(m.url, m.token)
	}

	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {

	case NomadJobsMsg:
		m.nomadJobData.allData = msg
		m.updateJobViewport()
		m.loading = false

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Reload):
			m.loading = true
			cmds = append(cmds, fetchJobs(m.url, m.token))

		case key.Matches(msg, m.keyMap.Forward):
			if !m.filter.EditingFilter {
				return m, func() tea.Msg {
					selectedId := m.nomadJobData.filteredData[m.viewport.CursorRow].ID
					return JobSelectedMsg{selectedId}
				}
			}
		}

		prevFilter := m.filter.Filter
		m.filter, cmd = m.filter.Update(msg)
		if m.filter.Filter != prevFilter {
			m.updateJobViewport()
		}
		cmds = append(cmds, cmd)

		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.loading {
		return "Loading jobs..."
	}
	return lipgloss.JoinVertical(lipgloss.Left, m.filter.View(), m.viewport.View())
}

func (m *Model) SetWindowSize(width, height int) {
	m.width, m.height = width, height
	m.viewport.SetSize(width, height-m.filter.ViewHeight())
}

func (m *Model) updateFilteredJobData() {
	var filteredJobData []jobResponseEntry
	for _, entry := range m.nomadJobData.allData {
		if entry.MatchesFilter(m.filter.Filter) {
			filteredJobData = append(filteredJobData, entry)
		}
	}
	m.nomadJobData.filteredData = filteredJobData
}

func (m *Model) updateJobViewport() {
	m.updateFilteredJobData()
	table := jobResponsesAsTable(m.nomadJobData.filteredData)
	m.viewport.SetHeader(strings.Join(table.HeaderRows, "\n"))
	m.viewport.SetContent(strings.Join(table.ContentRows, "\n"))
	m.viewport.SetCursorRow(0)
}
