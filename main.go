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
	"wander/components/allocations"
	"wander/components/header"
	"wander/components/jobs"
	"wander/components/logs"
	"wander/components/page"
	"wander/dev"
	"wander/message"
)

var (
	NomadTokenEnvVariable = "NOMAD_TOKEN"
	NomadUrlEnvVariable   = "NOMAD_ADDR"
)

type model struct {
	nomadToken      string
	nomadUrl        string
	keyMap          mainKeyMap
	currentPage     page.Page
	jobsPage        jobs.Model
	allocationsPage allocations.Model
	logsPage        logs.Model
	header          header.Model
	width, height   int
	initialized     bool
	err             error
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

	keyMap := getMainKeyMap()
	logo := []string{
		"█ █ █ █▀█ █▄ █ █▀▄ █▀▀ █▀█",
		"▀▄▀▄▀ █▀█ █ ▀█ █▄▀ ██▄ █▀▄",
	}
	logoString := strings.Join(logo, "\n")
	return model{
		nomadToken:  nomadToken,
		nomadUrl:    nomadUrl,
		keyMap:      keyMap,
		currentPage: page.Jobs,
		header:      header.New(logoString, nomadUrl, ""),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

//func (m model) fetchPageDataCmd() tea.Cmd {
//	switch m.currentPage {
//
//	case page.Jobs:
//		return command.fetchJobs(m.nomadUrl, m.nomadToken)
//
//	case page.Allocations:
//		return command.fetchAllocations(m.nomadUrl, m.nomadToken, m.jobsPage.SelectedJobId)
//
//	case page.Logs:
//		return command.FetchLogs(m.nomadUrl, m.nomadToken, m.selectedAlloc.ID, m.selectedAlloc.taskName, m.logType)
//	}
//	return nil
//}

//func (m *model) setHeaderKeyHelp() {
//	m.header.KeyHelp = getPageKeyMapView(m.currentPage, m.header.EditingFilter, m.header.HasFilter())
//}

//func (m *model) setFilter(s string) {
//m.header.Filter = s
//m.jobsPage.SetHighlightText(s)

//switch m.currentPage {
//case page.Jobs:
//	m.updateJobViewport()
//case page.Allocations:
//	m.updateAllocationViewport()
//case page.Logs:
//	m.updateLogViewport()
//}
//}

//func (m *model) updateFilteredAllocationData() {
//	var filteredAllocationData []nomad.AllocationRowEntry
//	for _, entry := range m.nomadAllocationData.allData {
//		if entry.MatchesFilter(m.header.Filter) {
//			filteredAllocationData = append(filteredAllocationData, entry)
//		}
//	}
//	m.nomadAllocationData.filteredData = filteredAllocationData
//}
//
//func (m *model) updateAllocationViewport() {
//	m.updateFilteredAllocationData()
//	table := formatter.AllocationsAsTable(m.nomadAllocationData.filteredData)
//	m.updateViewport(table, 0)
//}
//
//func (m *model) updateFilteredLogData() {
//	var filteredLogData []nomad.LogRow
//	for _, log := range m.nomadLogData.allData {
//		if log.MatchesFilter(m.header.Filter) {
//			filteredLogData = append(filteredLogData, log)
//		}
//	}
//	m.nomadLogData.filteredData = filteredLogData
//}
//
//func (m *model) updateLogViewport() {
//	m.updateFilteredLogData()
//	table := formatter.LogsAsTable(m.nomadLogData.filteredData, m.logType)
//	m.updateViewport(table, len(table.ContentRows)-1)
//}
//

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	dev.Debug(fmt.Sprintf("main %T", msg))
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Exit):
			return m, tea.Quit

		default:
			if m.currentPageLoading() {
				return m, nil
			}
		}

	case message.ErrMsg:
		m.err = msg
		return m, nil
	case tea.WindowSizeMsg:
		m.setWindowSize(msg.Width, msg.Height)
		pageHeight := m.getPageHeight()

		if !m.initialized {
			m.jobsPage = jobs.New(m.nomadUrl, m.nomadToken, msg.Width, pageHeight)
			m.allocationsPage = allocations.New(m.nomadUrl, m.nomadToken, msg.Width, pageHeight)
			m.logsPage = logs.New(m.nomadUrl, m.nomadToken, msg.Width, pageHeight)
			m.initialized = true
		} else {
			// TODO LEO: dry
			m.jobsPage.SetWindowSize(msg.Width, pageHeight)
			m.allocationsPage.SetWindowSize(msg.Width, pageHeight)
			m.logsPage.SetWindowSize(msg.Width, pageHeight)
		}

	case message.ViewJobsMsg:
		m.setPage(page.Jobs)
		m.jobsPage.ClearFilter()
		return m, jobs.FetchJobs(m.nomadUrl, m.nomadToken)

	case message.ViewAllocationsMsg:
		m.setPage(page.Allocations)
		jobID := m.jobsPage.LastSelectedJobID
		m.allocationsPage.SetJobID(jobID)
		return m, allocations.FetchAllocations(m.nomadUrl, m.nomadToken, jobID)

	case message.ViewLogsMsg:
		m.setPage(page.Logs)
		allocID, taskName := m.allocationsPage.LastSelectedAllocID, m.allocationsPage.LastSelectedTaskName
		m.logsPage.SetAllocationData(allocID, taskName)
		return m, logs.FetchLogs(
			m.nomadUrl,
			m.nomadToken,
			allocID,
			taskName,
			m.logsPage.LastSelectedLogType,
		)
	}

	switch m.currentPage {
	case page.Jobs:
		m.jobsPage, cmd = m.jobsPage.Update(msg)
		cmds = append(cmds, cmd)

	case page.Allocations:
		m.allocationsPage, cmd = m.allocationsPage.Update(msg)
		cmds = append(cmds, cmd)

	case page.Logs:
		m.logsPage, cmd = m.logsPage.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}

	pageView := ""
	switch m.currentPage {
	case page.Jobs:
		pageView = m.jobsPage.View()

	case page.Allocations:
		pageView = m.allocationsPage.View()

	case page.Logs:
		pageView = m.logsPage.View()
	}

	finalView := m.header.View() + "\n" + pageView
	return finalView
}

func (m *model) setPage(p page.Page) {
	m.currentPage = p
	//m.setHeaderKeyHelp()
}

func (m *model) setWindowSize(width, height int) {
	m.width, m.height = width, height
}

func (m model) getPageHeight() int {
	return m.height - m.header.ViewHeight()
}

func (m model) currentPageLoading() bool {
	switch m.currentPage {
	case page.Jobs:
		return m.jobsPage.Loading
	case page.Allocations:
		return m.allocationsPage.Loading
	case page.Logs:
		return m.logsPage.Loading
	}
	return true
}

func main() {
	program := tea.NewProgram(initialModel(), tea.WithAltScreen())

	dev.Debug("~STARTING UP~")
	if err := program.Start(); err != nil {
		fmt.Printf("Error on wander startup: %v", err)
		os.Exit(1)
	}
}
