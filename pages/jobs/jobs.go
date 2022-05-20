package jobs

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
	"wander/components/filter"
	"wander/components/viewport"
	"wander/dev"
	"wander/keymap"
	"wander/pages"
)

type jobsData struct {
	allData, filteredData []jobResponseEntry
}

type nomadJobsMsg []jobResponseEntry

type Model struct {
	initialized       bool
	url, token        string
	jobsData          jobsData
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

	switch msg := msg.(type) {
	case nomadJobsMsg:
		m.jobsData.allData = msg
		m.updateJobViewport()
		m.Loading = false

	case tea.KeyMsg:
		if m.filter.Focused() || key.Matches(msg, keymap.KeyMap.Filter) {
			prevFilter := m.filter.Filter
			m.filter, cmd = m.filter.Update(msg)
			if m.filter.Filter != prevFilter {
				m.updateJobViewport()
			}
			cmds = append(cmds, cmd)
		} else {
			switch {
			case key.Matches(msg, keymap.KeyMap.Reload):
				m.Loading = true
				cmds = append(cmds, FetchJobs(m.url, m.token))

			case key.Matches(msg, keymap.KeyMap.Forward):
				if len(m.jobsData.filteredData) > 0 {
					m.LastSelectedJobID = m.jobsData.filteredData[m.viewport.CursorRow].ID
					return m, pages.ToAllocationsPageCmd
				}

			case key.Matches(msg, keymap.KeyMap.Back):
				m.ClearFilter()
			}

			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	content := "Loading jobs..."
	if !m.Loading {
		content = m.viewport.View()
	}
	return lipgloss.JoinVertical(lipgloss.Left, m.filter.View(), content)
}

func (m *Model) SetWindowSize(width, height int) {
	m.width, m.height = width, height
	m.viewport.SetSize(width, height-m.filter.ViewHeight())
}

func (m *Model) ClearFilter() {
	m.filter.SetFilter("")
	m.updateJobViewport()
}

func (m *Model) updateFilteredJobData() {
	var filteredJobData []jobResponseEntry
	for _, entry := range m.jobsData.allData {
		if entry.MatchesFilter(m.filter.Filter) {
			filteredJobData = append(filteredJobData, entry)
		}
	}
	m.jobsData.filteredData = filteredJobData
}

func (m *Model) updateJobViewport() {
	m.viewport.Highlight = m.filter.Filter
	m.updateFilteredJobData()
	table := jobResponsesAsTable(m.jobsData.filteredData)
	m.viewport.SetHeaderAndContent(
		strings.Join(table.HeaderRows, "\n"),
		strings.Join(table.ContentRows, "\n"),
	)
	m.viewport.SetCursorRow(0)
}
