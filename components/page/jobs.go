package page

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
	filter           filter.Model
	keyMap           pageKeyMap
	loading          bool
	SelectedJobId    string
}

func NewJobsModel(fetchDataCommand tea.Cmd, width, height int) JobsModel {
	model := JobsModel{
		fetchDataCommand: fetchDataCommand,
		width:            width,
		height:           height,
		viewport:         viewport.New(width, height-1), // TODO LEO: m.filter.ViewHeight() once component
		keyMap:           getKeyMap(),
		loading:          true,
	}
	model.filter = filter.New(func() {
		dev.Debug("HI")
		dev.Debug(fmt.Sprintf("model pointer %p", &model))
		model.updateFilteredJobData()
		dev.Debug(fmt.Sprintf("model pointer %p", &model))
	})
	return model
}

func (m JobsModel) Update(msg tea.Msg) (JobsModel, tea.Cmd) {
	dev.Debug(fmt.Sprintf("jobs %T", msg))
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {

	case message.NomadJobsMsg:
		m.nomadJobData.allData = msg
		m.updateJobViewport()
		m.loading = false

	//case message.UpdatedFilterMsg:
	//	m.updateJobViewport()

	case tea.KeyMsg:
		if key.Matches(msg, m.keyMap.Reload) {
			m.loading = true
			cmds = append(cmds, m.fetchDataCommand)
		}

		m.filter, cmd = m.filter.Update(msg)
		cmds = append(cmds, cmd)

		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m JobsModel) View() string {
	if m.loading {
		return "Loading jobs..."
	}

	return lipgloss.JoinVertical(lipgloss.Left, m.filter.View(), m.viewport.View())
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
	dev.Debug("HERE")
	dev.Debug(fmt.Sprintf("filter %s", m.filter.Filter))
	dev.Debug(fmt.Sprintf("model pointer m %p", &m))
	var filteredJobData []nomad.JobResponseEntry
	for _, entry := range m.nomadJobData.allData {
		if entry.MatchesFilter(m.filter.Filter) {
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
