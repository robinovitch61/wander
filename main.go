package main

// TODO LEO: known bugs
// - [ ] Crashes if terminal height smaller than header height
// - [ ] Cursor shows up if no viewport content
// - [ ] Can crash app by hitting Enter when jobs loading

// TODO LEO: consider:
// - can deduplicate state by only having viewport, header, etc. having copies of certain state and having top-level reference it
// - can pass updater functions that mutate parent state as props into components (e.g. onEnter(m *model) to viewport)

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

type nomadAllocationData struct {
	allAllocationData      []nomad.AllocationRowEntry
	filteredAllocationData []nomad.AllocationRowEntry
}

type nomadLogData struct {
	allLogData      []nomad.LogRow
	filteredLogData []nomad.LogRow
}

type selectedAlloc struct {
	ID, taskName string
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
	nomadJobData        nomadJobData
	nomadAllocationData nomadAllocationData
	nomadLogData        nomadLogData
	selectedJobId       string
	selectedAlloc       selectedAlloc
	logType             nomad.LogType
	err                 error
}

func (m model) Init() tea.Cmd {
	return command.FetchJobs(m.nomadUrl, m.nomadToken)
}

func (m model) fetchPageDataCmd() tea.Cmd {
	switch m.page {

	case page.Jobs:
		return command.FetchJobs(m.nomadUrl, m.nomadToken)

	case page.Allocation:
		return command.FetchAllocations(m.nomadUrl, m.nomadToken, m.selectedJobId)

	case page.Logs:
		return command.FetchLogs(m.nomadUrl, m.nomadToken, m.selectedAlloc.ID, m.selectedAlloc.taskName, m.logType)
	}
	return nil
}

func (m *model) setFiltering(isEditingFilter, clearFilter bool) {
	m.header.EditingFilter = isEditingFilter
	if clearFilter {
		m.setFilter("")
	}
}

func (m *model) setFilter(s string) {
	m.header.Filter = s

	switch m.page {
	case page.Jobs:
		m.updateJobViewport()
	case page.Allocation:
		m.updateAllocationViewport()
	case page.Logs:
		m.updateLogViewport(m.logType)
	}
}

func (m *model) updateViewport(table formatter.Table, cursorRow int) {
	m.viewport.SetHeader(strings.Join(table.HeaderRows, "\n"))
	m.viewport.SetContent(strings.Join(table.ContentRows, "\n"))
	m.viewport.SetCursorRow(cursorRow)
}

func (m *model) updateFilteredJobData() {
	var filteredJobData []nomad.JobResponseEntry
	for _, entry := range m.nomadJobData.allJobData {
		if entry.MatchesFilter(m.header.Filter) {
			filteredJobData = append(filteredJobData, entry)
		}
	}
	m.nomadJobData.filteredJobData = filteredJobData
}

func (m *model) updateJobViewport() {
	m.updateFilteredJobData()
	table := formatter.JobResponsesAsTable(m.nomadJobData.filteredJobData)
	m.updateViewport(table, 0)
}

func (m *model) updateFilteredAllocationData() {
	var filteredAllocationData []nomad.AllocationRowEntry
	for _, entry := range m.nomadAllocationData.allAllocationData {
		if entry.MatchesFilter(m.header.Filter) {
			filteredAllocationData = append(filteredAllocationData, entry)
		}
	}
	m.nomadAllocationData.filteredAllocationData = filteredAllocationData
}

func (m *model) updateAllocationViewport() {
	m.updateFilteredAllocationData()
	table := formatter.AllocationsAsTable(m.nomadAllocationData.filteredAllocationData)
	m.updateViewport(table, 0)
}

func (m *model) updateFilteredLogData() {
	var filteredLogData []nomad.LogRow
	for _, log := range m.nomadLogData.allLogData {
		if log.MatchesFilter(m.header.Filter) {
			filteredLogData = append(filteredLogData, log)
		}
	}
	m.nomadLogData.filteredLogData = filteredLogData
}

func (m *model) updateLogViewport(logType nomad.LogType) {
	m.updateFilteredLogData()
	table := formatter.LogsAsTable(m.nomadLogData.filteredLogData, logType)
	m.updateViewport(table, len(table.ContentRows)-1)
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
		m.updateJobViewport()
		return m, nil

	case message.NomadAllocationMsg:
		dev.Debug("NomadAllocationMsg")
		m.nomadAllocationData.allAllocationData = msg
		m.updateAllocationViewport()
		return m, nil

	case message.NomadLogsMsg:
		dev.Debug("NomadLogsMsg")
		m.nomadLogData.allLogData = msg.Data
		m.updateLogViewport(msg.LogType)
		return m, nil

	case message.ErrMsg:
		dev.Debug("errMsg")
		m.err = msg
		return m, nil

	case tea.KeyMsg:
		dev.Debug(fmt.Sprintf("KeyMsg '%s'", msg))

		if m.header.EditingFilter {
			switch {
			case key.Matches(msg, m.keyMap.Back):
				m.setFiltering(false, true)
			case key.Matches(msg, m.keyMap.Enter):
				m.setFiltering(false, false)
			default:
				switch msg.Type {
				case tea.KeyBackspace:
					if len(m.header.Filter) > 0 {
						if msg.Alt {
							m.setFilter("")
						} else {
							m.setFilter(m.header.Filter[:len(m.header.Filter)-1])
						}
					}
				case tea.KeyRunes:
					m.setFilter(m.header.Filter + msg.String())
				}
			}
			return m, nil
		}

		switch {
		case key.Matches(msg, m.keyMap.Exit):
			return m, tea.Quit

		case key.Matches(msg, m.keyMap.Enter):
			switch m.page {
			case page.Jobs:
				m.selectedJobId = m.nomadJobData.filteredJobData[m.viewport.CursorRow].ID
			case page.Allocation:
				alloc := m.nomadAllocationData.filteredAllocationData[m.viewport.CursorRow]
				m.selectedAlloc = selectedAlloc{alloc.ID, alloc.TaskName}
			}

			if newPage := m.page.Forward(); newPage != m.page {
				m.setFilter("")
				m.page = newPage
				cmd := m.fetchPageDataCmd()
				m.viewport.SetLoading(newPage.LoadingString())
				return m, cmd
			}

		case key.Matches(msg, m.keyMap.Back):
			if m.header.Filter != "" {
				m.setFiltering(false, true)
				return m, nil
			}

			if newPage := m.page.Backward(); newPage != m.page {
				m.page = newPage
				cmd := m.fetchPageDataCmd()
				m.viewport.SetLoading(newPage.LoadingString())
				return m, cmd
			}

		case key.Matches(msg, m.keyMap.Reload):
			cmd := m.fetchPageDataCmd()
			m.viewport.SetLoading(m.page.ReloadingString())
			return m, cmd

		case key.Matches(msg, m.keyMap.Filter):
			m.setFiltering(true, false)
			return m, nil

		case key.Matches(msg, m.keyMap.StdOut) && m.page == page.Logs:
			m.logType = nomad.StdOut
			m.viewport.SetLoading(m.page.LoadingString())
			return m, m.fetchPageDataCmd()

		case key.Matches(msg, m.keyMap.StdErr) && m.page == page.Logs:
			m.logType = nomad.StdErr
			m.viewport.SetLoading(m.page.LoadingString())
			return m, m.fetchPageDataCmd()
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
			m.logType = nomad.StdOut
			m.initialized = true

			firstPage := page.Jobs
			m.page = firstPage
			m.viewport.SetLoading(firstPage.LoadingString())

			cmd := m.fetchPageDataCmd()
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
