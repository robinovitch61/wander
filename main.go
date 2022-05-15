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
	"wander/command"
	"wander/components/header"
	"wander/components/page"
	"wander/dev"
	"wander/message"
	"wander/nomad"
)

var (
	NomadTokenEnvVariable = "NOMAD_TOKEN"
	NomadUrlEnvVariable   = "NOMAD_ADDR"
)

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
	currentPage         page.Page
	jobsPage            page.JobsModel
	header              header.Model
	width, height       int
	initialized         bool
	nomadAllocationData nomadAllocationData
	nomadLogData        nomadLogData
	selectedAlloc       selectedAlloc
	logType             nomad.LogType
	loading             bool
	err                 error
}

func (m model) Init() tea.Cmd {
	return command.FetchJobs(m.nomadUrl, m.nomadToken)
}

func (m model) fetchPageDataCmd() tea.Cmd {
	switch m.currentPage {

	case page.Jobs:
		return command.FetchJobs(m.nomadUrl, m.nomadToken)

	case page.Allocation:
		return command.FetchAllocations(m.nomadUrl, m.nomadToken, m.jobsPage.SelectedJobId)

	case page.Logs:
		return command.FetchLogs(m.nomadUrl, m.nomadToken, m.selectedAlloc.ID, m.selectedAlloc.taskName, m.logType)
	}
	return nil
}

//func (m *model) setHeaderKeyHelp() {
//	m.header.KeyHelp = getPageKeyMapView(m.currentPage, m.header.EditingFilter, m.header.HasFilter())
//}

func (m *model) setPage(p page.Page) {
	m.currentPage = p
	//m.setHeaderKeyHelp()
}

//func (m *model) setFilter(s string) {
//m.header.Filter = s
//m.jobsPage.SetHighlightText(s)

//switch m.currentPage {
//case page.Jobs:
//	m.updateJobViewport()
//case page.Allocation:
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
func (m *model) setLoading(loadingString string) {
	m.loading = true
	//m.viewport.SetLoading(loadingString)
}

func (m *model) setWindowSize(width, height int) {
	m.width = width
	m.height = height
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	m.jobsPage, cmd = m.jobsPage.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {

	case message.ErrMsg:
		dev.Debug("errMsg")
		m.err = msg
		return m, nil

	case tea.WindowSizeMsg:
		dev.Debug("WindowSizeMsg main")
		m.setWindowSize(msg.Width, msg.Height)
		pageHeight := msg.Height - m.header.ViewHeight()

		if !m.initialized {
			jobsCommand := command.FetchJobs(m.nomadUrl, m.nomadToken)
			m.jobsPage = page.NewJobsModel(jobsCommand, msg.Width, pageHeight)
			// TODO LEO: rest of pages here?
			m.initialized = true
		} else {
			m.jobsPage.SetWindowSize(msg.Width, pageHeight)
		}

	//case message.NomadJobsMsg:
	//	dev.Debug("nomadJobsMsg main")
	//	m.jobsPage, cmd = m.jobsPage.Update(msg)
	//	cmds = append(cmds, cmd)

	//case message.NomadAllocationMsg:
	//	dev.Debug("NomadAllocationMsg")
	//	m.nomadAllocationData.allData = msg
	//	m.updateAllocationViewport()
	//	m.loading = false
	//	return m, nil
	//
	//case message.NomadLogsMsg:
	//	dev.Debug("NomadLogsMsg")
	//	m.nomadLogData.allData = msg.Data
	//	m.logType = msg.LogType
	//	m.updateLogViewport()
	//	m.loading = false
	//	return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Exit):
			return m, tea.Quit

			//case m.loading:
			//	return m, nil

			//case key.Matches(msg, m.keyMap.Forward):
			//	//switch m.currentPage {
			//	//case page.Jobs:
			//	//	m.selectedJobId = m.nomadJobData.filteredData[m.viewport.CursorRow].ID
			//	//case page.Allocation:
			//	//	alloc := m.nomadAllocationData.filteredData[m.viewport.CursorRow]
			//	//	m.selectedAlloc = selectedAlloc{alloc.ID, alloc.TaskName}
			//	//}
			//
			//	if newPage := m.currentPage.Forward(); newPage != m.currentPage {
			//		m.setFilter("")
			//		m.setPage(newPage)
			//		cmd := m.fetchPageDataCmd()
			//		m.setLoading(newPage.LoadingString())
			//		return m, cmd
			//	}
			//
			//case key.Matches(msg, m.keyMap.Back):
			//	if m.header.Filter != "" {
			//		m.setFiltering(false, true)
			//		return m, nil
			//	}
			//
			//	if newPage := m.currentPage.Backward(); newPage != m.currentPage {
			//		m.setPage(newPage)
			//		cmd := m.fetchPageDataCmd()
			//		m.setLoading(newPage.LoadingString())
			//		return m, cmd
			//	}

			// TODO LEO: put these in logs page
			//case key.Matches(msg, m.keyMap.StdOut) && m.currentPage == page.Logs:
			//	m.logType = nomad.StdOut
			//	m.setLoading(m.currentPage.LoadingString())
			//	return m, m.fetchPageDataCmd()
			//
			//case key.Matches(msg, m.keyMap.StdErr) && m.currentPage == page.Logs:
			//	m.logType = nomad.StdErr
			//	m.setLoading(m.currentPage.LoadingString())
			//	return m, m.fetchPageDataCmd()
		}
	}

	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nError: %v\n\n", m.err)
	}

	pageView := ""
	switch m.currentPage {
	case page.Jobs:
		pageView = m.jobsPage.View()
	}
	finalView := m.header.View() + "\n" + pageView
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

	keyMap := getMainKeyMap()
	firstPage := page.Jobs
	logo := []string{
		"█ █ █ █▀█ █▄ █ █▀▄ █▀▀ █▀█",
		"▀▄▀▄▀ █▀█ █ ▀█ █▄▀ ██▄ █▀▄",
	}
	return model{
		nomadToken:  nomadToken,
		nomadUrl:    nomadUrl,
		keyMap:      keyMap,
		logType:     nomad.StdOut,
		currentPage: firstPage,
		header:      header.New(logo, nomadUrl, ""),
		loading:     true,
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
