package main

// TODO LEO: consider:
// - can deduplicate state by only having viewport, header, etc. having copies of certain state and having top-level reference it
// - can pass updater functions that mutate parent state as props into components (e.g. onEnter(m *model) to viewport)
// - can pass pointer to main model to child components in order to avoid replicating state

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
	NomadUrlEnvVariable   = "NOMAD_ADDR"
)

type nomadJobData struct {
	allData, filteredData []nomad.JobResponseEntry
}

type nomadAllocationData struct {
	allData, filteredData []nomad.AllocationRowEntry
}

type nomadLogData struct {
	allData, filteredData []nomad.LogRow
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
	loading             bool
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

func (m *model) setHeaderKeyHelp() {
	m.header.KeyHelp = KeyMapView(m.page, m.header.EditingFilter, m.header.HasFilter())
}

func (m *model) setFiltering(isEditingFilter, clearFilter bool) {
	m.header.EditingFilter = isEditingFilter
	if clearFilter {
		m.setFilter("")
	}
	m.setHeaderKeyHelp()
}

func (m *model) setPage(p page.Page) {
	m.page = p
	m.setHeaderKeyHelp()
}

func (m *model) setFilter(s string) {
	m.header.Filter = s
	m.viewport.Highlight = s

	switch m.page {
	case page.Jobs:
		m.updateJobViewport()
	case page.Allocation:
		m.updateAllocationViewport()
	case page.Logs:
		m.updateLogViewport()
	}
}

func (m *model) updateViewport(table formatter.Table, cursorRow int) {
	m.viewport.SetHeader(strings.Join(table.HeaderRows, "\n"))
	m.viewport.SetContent(strings.Join(table.ContentRows, "\n"))
	m.viewport.SetCursorRow(cursorRow)
}

func (m *model) updateFilteredJobData() {
	var filteredJobData []nomad.JobResponseEntry
	for _, entry := range m.nomadJobData.allData {
		if entry.MatchesFilter(m.header.Filter) {
			filteredJobData = append(filteredJobData, entry)
		}
	}
	m.nomadJobData.filteredData = filteredJobData
}

func (m *model) updateJobViewport() {
	m.updateFilteredJobData()
	table := formatter.JobResponsesAsTable(m.nomadJobData.filteredData)
	m.updateViewport(table, 0)
}

func (m *model) updateFilteredAllocationData() {
	var filteredAllocationData []nomad.AllocationRowEntry
	for _, entry := range m.nomadAllocationData.allData {
		if entry.MatchesFilter(m.header.Filter) {
			filteredAllocationData = append(filteredAllocationData, entry)
		}
	}
	m.nomadAllocationData.filteredData = filteredAllocationData
}

func (m *model) updateAllocationViewport() {
	m.updateFilteredAllocationData()
	table := formatter.AllocationsAsTable(m.nomadAllocationData.filteredData)
	m.updateViewport(table, 0)
}

func (m *model) updateFilteredLogData() {
	var filteredLogData []nomad.LogRow
	for _, log := range m.nomadLogData.allData {
		if log.MatchesFilter(m.header.Filter) {
			filteredLogData = append(filteredLogData, log)
		}
	}
	m.nomadLogData.filteredData = filteredLogData
}

func (m *model) updateLogViewport() {
	m.updateFilteredLogData()
	table := formatter.LogsAsTable(m.nomadLogData.filteredData, m.logType)
	m.updateViewport(table, len(table.ContentRows)-1)
}

func (m *model) setLoading(loadingString string) {
	m.loading = true
	m.viewport.SetLoading(loadingString)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {

	case message.NomadJobsMsg:
		dev.Debug("nomadJobsMsg")
		m.nomadJobData.allData = msg
		m.updateJobViewport()
		m.loading = false
		return m, nil

	case message.NomadAllocationMsg:
		dev.Debug("NomadAllocationMsg")
		m.nomadAllocationData.allData = msg
		m.updateAllocationViewport()
		m.loading = false
		return m, nil

	case message.NomadLogsMsg:
		dev.Debug("NomadLogsMsg")
		m.nomadLogData.allData = msg.Data
		m.logType = msg.LogType
		m.updateLogViewport()
		m.loading = false
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
			case key.Matches(msg, m.keyMap.Forward):
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

		case m.loading:
			return m, nil

		case key.Matches(msg, m.keyMap.Forward):
			switch m.page {
			case page.Jobs:
				m.selectedJobId = m.nomadJobData.filteredData[m.viewport.CursorRow].ID
			case page.Allocation:
				alloc := m.nomadAllocationData.filteredData[m.viewport.CursorRow]
				m.selectedAlloc = selectedAlloc{alloc.ID, alloc.TaskName}
			}

			if newPage := m.page.Forward(); newPage != m.page {
				m.setFilter("")
				m.setPage(newPage)
				cmd := m.fetchPageDataCmd()
				m.setLoading(newPage.LoadingString())
				return m, cmd
			}

		case key.Matches(msg, m.keyMap.Back):
			if m.header.Filter != "" {
				m.setFiltering(false, true)
				return m, nil
			}

			if newPage := m.page.Backward(); newPage != m.page {
				m.setPage(newPage)
				cmd := m.fetchPageDataCmd()
				m.setLoading(newPage.LoadingString())
				return m, cmd
			}

		case key.Matches(msg, m.keyMap.Reload):
			cmd := m.fetchPageDataCmd()
			m.setLoading(m.page.ReloadingString())
			return m, cmd

		case key.Matches(msg, m.keyMap.Filter):
			m.setFiltering(true, false)
			return m, nil

		case key.Matches(msg, m.keyMap.StdOut) && m.page == page.Logs:
			m.logType = nomad.StdOut
			m.setLoading(m.page.LoadingString())
			return m, m.fetchPageDataCmd()

		case key.Matches(msg, m.keyMap.StdErr) && m.page == page.Logs:
			m.logType = nomad.StdErr
			m.setLoading(m.page.LoadingString())
			return m, m.fetchPageDataCmd()
		}

	case tea.WindowSizeMsg:
		dev.Debug("WindowSizeMsg")
		m.width, m.height = msg.Width, msg.Height

		headerHeight := m.header.ViewHeight()
		footerHeight := 0
		viewportHeight := msg.Height - (headerHeight + footerHeight)

		if !m.initialized {
			// this is the first message received and initializes the viewport size
			m.viewport = viewport.New(msg.Width, viewportHeight)
			m.viewport.SetLoading(m.page.LoadingString())
			m.initialized = true
			return m, m.fetchPageDataCmd()
		} else {
			m.viewport.SetWidth(msg.Width)
			m.viewport.SetHeight(viewportHeight)
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nError: %v\n\n", m.err)
	}

	finalView := m.header.View() + "\n" + m.viewport.View()
	return finalView
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

	keyMap := getKeyMap()
	firstPage := page.Jobs
	logo := []string{
		"█ █ █ █▀█ █▄ █ █▀▄ █▀▀ █▀█",
		"▀▄▀▄▀ █▀█ █ ▀█ █▄▀ ██▄ █▀▄",
	}
	return model{
		nomadToken: nomadToken,
		nomadUrl:   nomadUrl,
		keyMap:     keyMap,
		logType:    nomad.StdOut,
		page:       firstPage,
		header:     header.New(logo, nomadUrl, KeyMapView(firstPage, false, false)),
		loading:    true,
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
