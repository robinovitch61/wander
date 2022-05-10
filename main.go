package main

// TODO LEO: known bugs
// - [ ] Crashes if terminal height smaller than header height

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"strings"
	"wander/command"
	"wander/components/header"
	"wander/components/viewport"
	"wander/dev"
	"wander/formatter"
	"wander/message"
	"wander/nomad"
	"wander/page"
)

var (
	NomadTokenEnvVariable = "NOMAD_TOKEN"
	NomadUrlEnvVariable   = "NOMAD_URL"
)

type nomadJobData struct {
	allJobData      []nomad.JobResponseEntry
	filteredJobData []nomad.JobResponseEntry
}

type model struct {
	nomadToken          string
	nomadUrl            string
	keyMap              mainKeyMap
	page                page.Page
	header              header.Model
	viewport            viewport.Model
	width, height       int
	initialized         bool
	editingActiveFilter bool
	activeFilter        string
	nomadJobData        nomadJobData
	selectedJobId       string // TODO LEO use this
	err                 error
}

func (m model) Init() tea.Cmd {
	return command.FetchJobs(m.nomadUrl, m.nomadToken)
}

func fetchPageDataCmd(m model) tea.Cmd {
	switch m.page {

	case page.Jobs:
		return command.FetchJobs(m.nomadUrl, m.nomadToken)

	case page.Allocation:
		return command.FetchAllocations(m.nomadUrl, m.nomadToken, m.selectedJobId)
	}
	return nil
}

func (m *model) setFiltering(editingActiveFilter, clearActiveFilter bool) {
	m.editingActiveFilter = editingActiveFilter
	dev.Debug(fmt.Sprintf("Set editingActiveFilter %t", editingActiveFilter))
	m.header.SetEditingFilter(editingActiveFilter)
	if clearActiveFilter {
		m.setActiveFilter("")
	}
}

func (m *model) setActiveFilter(s string) {
	m.activeFilter = s
	m.header.SetFilterString(m.activeFilter)

	switch m.page {
	case page.Jobs:
		m.updateJobsViewport()
	case page.Allocation:
		// TODO LEO: implement
	}
}

func (m *model) updateViewport(table formatter.Table) {
	m.viewport.SetHeader(strings.Join(table.HeaderRows, "\n"))
	m.viewport.SetContent(strings.Join(table.ContentRows, "\n"))
}

func (m *model) updateFilteredJobData() {
	var filteredJobData []nomad.JobResponseEntry
	for _, entry := range m.nomadJobData.allJobData {
		if entry.MatchesFilter(m.activeFilter) {
			filteredJobData = append(filteredJobData, entry)
		}
	}
	m.nomadJobData.filteredJobData = filteredJobData
}

func (m *model) updateJobsViewport() {
	m.updateFilteredJobData()
	table := formatter.JobResponseAsTable(m.nomadJobData.filteredJobData)
	m.updateViewport(table)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {

	case message.NomadJobsMsg:
		dev.Debug("nomadJobsMsg")
		m.nomadJobData.allJobData = msg
		m.updateJobsViewport()
		return m, nil

	case message.NomadAllocationMsg:
		dev.Debug("NomadAllocationMsg")
		m.updateViewport(msg.Table)
		return m, nil

	case message.ErrMsg:
		dev.Debug("errMsg")
		m.err = msg
		return m, nil

	case tea.KeyMsg:
		dev.Debug(fmt.Sprintf("KeyMsg '%s'", msg))

		if m.editingActiveFilter {
			switch {
			case key.Matches(msg, m.keyMap.Back):
				m.setFiltering(false, true)
			case key.Matches(msg, m.keyMap.Enter):
				m.setFiltering(false, false)
			default:
				switch msg.Type {
				case tea.KeyBackspace:
					if len(m.activeFilter) > 0 {
						m.setActiveFilter(m.activeFilter[:len(m.activeFilter)-1])
					}
				case tea.KeyRunes:
					m.setActiveFilter(m.activeFilter + msg.String())
				}
			}
			return m, nil
		}

		switch {
		case key.Matches(msg, m.keyMap.Exit):
			return m, tea.Quit

		case key.Matches(msg, m.keyMap.Enter):
			if m.page == page.Jobs {
				m.selectedJobId = m.nomadJobData.filteredJobData[m.viewport.CursorRow].ID
			}

			if newPage := m.page.Forward(); newPage != m.page {
				m.setActiveFilter("")
				m.page = newPage
				cmd := fetchPageDataCmd(m)
				m.viewport.SetLoading(newPage.LoadingString())
				return m, cmd
			}

		case key.Matches(msg, m.keyMap.Back):
			m.setFiltering(false, true)
			if newPage := m.page.Backward(); newPage != m.page {
				m.page = newPage
				cmd := fetchPageDataCmd(m)
				m.viewport.SetLoading(newPage.LoadingString())
				return m, cmd
			}

		case key.Matches(msg, m.keyMap.Reload):
			cmd := fetchPageDataCmd(m)
			m.viewport.SetLoading(m.page.ReloadingString())
			return m, cmd

		case key.Matches(msg, m.keyMap.Filter):
			m.setFiltering(true, true)
			return m, nil
		}

	case tea.WindowSizeMsg:
		dev.Debug("WindowSizeMsg")
		m.width, m.height = msg.Width, msg.Height

		headerHeight := m.header.ViewHeight()
		footerHeight := 0
		viewportHeight := msg.Height - (headerHeight + footerHeight)

		if !m.initialized {
			// this is the first message received and the initial entrypoint to the app
			// TODO LEO: separate loading/reloading out of viewport, then can initialize app and viewport separately
			// tradeoff here is can't have multiple components loading at same time then?
			m.keyMap = getKeyMap()
			m.viewport = viewport.New(msg.Width, viewportHeight)
			m.initialized = true

			firstPage := page.Jobs
			m.page = firstPage
			m.viewport.SetLoading(firstPage.LoadingString())

			cmd := fetchPageDataCmd(m)
			return m, cmd
		} else {
			m.viewport.SetWidth(msg.Width)
			m.viewport.SetHeight(viewportHeight)
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	m.header, cmd = m.header.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nError: %v\n\n", m.err)
	}

	finalView := m.header.View() + "\n" + m.viewport.View()
	return finalView
	//return lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(finalView)
}

func initialModel() model {
	nomadToken := os.Getenv(NomadTokenEnvVariable)
	if nomadToken == "" {
		fmt.Printf("Set environment variable %s\n", NomadTokenEnvVariable)
		os.Exit(1)
	}

	nomadUrl := os.Getenv(NomadUrlEnvVariable)
	if nomadUrl == "" {
		fmt.Printf("Set environment variable %s\n", NomadUrlEnvVariable)
		os.Exit(1)
	}

	return model{
		nomadToken: nomadToken,
		nomadUrl:   nomadUrl,
		header:     header.New(nomadUrl, ""),
	}
}

func main() {
	program := tea.NewProgram(initialModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())

	dev.Debug("~STARTING UP~")
	if err := program.Start(); err != nil {
		fmt.Printf("Error on wander startup: %v\n", err)
		os.Exit(1)
	}
}
