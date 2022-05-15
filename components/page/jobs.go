package page

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
	"wander/components/viewport"
	"wander/dev"
	"wander/formatter"
	"wander/message"
	"wander/nomad"
)

type nomadJobData struct {
	allData, filteredData []nomad.JobResponseEntry
}

type JobsModel struct {
	fetchDataCommand tea.Cmd
	nomadJobData     nomadJobData
	width, height    int
	viewport         viewport.Model
	keyMap           pageKeyMap
	loading          bool
	SelectedJobId    string
	Filter           string
	EditingFilter    bool
}

func NewJobsModel(fetchDataCommand tea.Cmd, width, height int) JobsModel {
	return JobsModel{
		fetchDataCommand: fetchDataCommand,
		width:            width,
		height:           height,
		viewport:         viewport.New(width, height-1), // TODO LEO: m.filter.ViewHeight() once component
		keyMap:           getKeyMap(),
		loading:          true,
	}
}

func (m JobsModel) Update(msg tea.Msg) (JobsModel, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {

	case message.NomadJobsMsg:
		m.nomadJobData.allData = msg
		m.updateJobViewport()
		m.loading = false

	case tea.KeyMsg:
		cmd = m.handleKeyPress(msg)
		cmds = append(cmds, cmd)
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m JobsModel) View() string {
	if m.loading {
		return "Loading jobs..."
	}

	jobsPageView := "Jobs"
	// TODO LEO: filter component View here?
	if m.EditingFilter || len(m.Filter) > 0 {
		jobsPageView += fmt.Sprintf(" (filter: %s)", m.Filter)
	} else {
		jobsPageView += " <'/' to filter>"
	}

	return lipgloss.JoinVertical(lipgloss.Left, jobsPageView, m.viewport.View())
}

func (m *JobsModel) SetWindowSize(width, height int) {
	dev.Debug("HERE")
	m.width, m.height = width, height
	m.viewport.SetSize(width, height-1) // TODO LEO: m.filter.ViewHeight() once component
}

func (m *JobsModel) updateTable(table formatter.Table, cursorRow int) {
	m.viewport.SetHeader(strings.Join(table.HeaderRows, "\n"))
	m.viewport.SetContent(strings.Join(table.ContentRows, "\n"))
	m.viewport.SetCursorRow(cursorRow)
}

func (m *JobsModel) updateFilteredJobData() {
	var filteredJobData []nomad.JobResponseEntry
	for _, entry := range m.nomadJobData.allData {
		if entry.MatchesFilter(m.Filter) {
			filteredJobData = append(filteredJobData, entry)
		}
	}
	m.nomadJobData.filteredData = filteredJobData
}

func (m *JobsModel) updateJobViewport() {
	m.updateFilteredJobData()
	table := formatter.JobResponsesAsTable(m.nomadJobData.filteredData)
	m.updateTable(table, 0)
}

func (m *JobsModel) handleKeyPress(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.keyMap.Back):
		m.setFiltering(false, true)
		m.updateJobViewport()
	}

	if m.EditingFilter {
		switch {
		case key.Matches(msg, m.keyMap.Forward):
			m.setFiltering(false, false)
		default:
			switch msg.Type {
			case tea.KeyBackspace:
				if len(m.Filter) > 0 {
					if msg.Alt {
						m.setFilter("")
					} else {
						m.setFilter(m.Filter[:len(m.Filter)-1])
					}
				}
			case tea.KeyRunes:
				m.setFilter(m.Filter + msg.String())
			}
		}
		m.updateJobViewport()
	} else {
		switch {
		case key.Matches(msg, m.keyMap.Reload):
			m.loading = true
			return m.fetchDataCommand

		case key.Matches(msg, m.keyMap.Filter):
			m.setFiltering(true, false)
		}
	}
	return nil
}

func (m *JobsModel) setFilter(filter string) {
	m.Filter = filter
}

func (m *JobsModel) setFiltering(isEditingFilter, clearFilter bool) {
	dev.Debug(fmt.Sprintf("isEditingFilter %t clearFilter %t", isEditingFilter, clearFilter))
	m.EditingFilter = isEditingFilter
	if clearFilter {
		m.setFilter("")
	}
	//m.setHeaderKeyHelp()
}
