package jobs

import (
	tea "github.com/charmbracelet/bubbletea"
	"strings"
	"wander/components/viewport"
	"wander/dev"
	"wander/formatter"
	"wander/nomad"
)

type nomadJobData struct {
	allData, filteredData []nomad.JobResponseEntry
}

type Model struct {
	fetchDataCommand tea.Cmd
	initialized      bool
	nomadJobData     nomadJobData
	width, height    int
	viewport         viewport.Model
	highlightText    string
	SelectedJobId    string
}

func (m Model) New() Model {
	return Model{}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		dev.Debug("WindowSizeMsg")
		m.width, m.height = msg.Width, msg.Height

		headerHeight := 0
		footerHeight := 0
		viewportHeight := msg.Height - (headerHeight + footerHeight)

		if !m.initialized {
			// this is the first message received and initializes the viewport size
			m.viewport = viewport.New(msg.Width, viewportHeight)
			m.viewport.SetLoading("LOADING")
			m.initialized = true
			return m, m.fetchDataCommand
		} else {
			m.viewport.SetWidth(msg.Width)
			m.viewport.SetHeight(viewportHeight)
		}
	}
	return m, nil
}

func (m Model) View() string {
	return ""
}

func (m *Model) SetHighlightText(h string) {
	m.highlightText = h
}

func (m *Model) updateTable(table formatter.Table, cursorRow int) {
	m.viewport.SetHeader(strings.Join(table.HeaderRows, "\n"))
	m.viewport.SetContent(strings.Join(table.ContentRows, "\n"))
	m.viewport.SetCursorRow(cursorRow)
}

func (m *Model) updateFilteredJobData() {
	var filteredJobData []nomad.JobResponseEntry
	for _, entry := range m.nomadJobData.allData {
		if entry.MatchesFilter(m.highlightText) {
			filteredJobData = append(filteredJobData, entry)
		}
	}
	m.nomadJobData.filteredData = filteredJobData
}

func (m *Model) updateJobViewport() {
	m.updateFilteredJobData()
	table := formatter.JobResponsesAsTable(m.nomadJobData.filteredData)
	m.updateTable(table, 0)
}
