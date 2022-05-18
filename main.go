package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"strings"
	"wander/components/header"
	"wander/dev"
	"wander/keymap"
	"wander/message"
	"wander/pages"
	"wander/pages/allocations"
	"wander/pages/jobs"
	"wander/pages/logline"
	"wander/pages/logs"
)

var (
	NomadTokenEnvVariable = "NOMAD_TOKEN"
	NomadUrlEnvVariable   = "NOMAD_ADDR"
)

type model struct {
	nomadToken       string
	nomadUrl         string
	currentPage      pages.Page
	jobsPage         jobs.Model
	allocationsPage  allocations.Model
	logsPage         logs.Model
	loglinePage      logline.Model
	selectedJobID    string
	selectedAllocID  string
	selectedTaskName string
	selectedLogType  logs.LogType
	header           header.Model
	width, height    int
	initialized      bool
	err              error
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

	logo := []string{
		"█ █ █ █▀█ █▄ █ █▀▄ █▀▀ █▀█",
		"▀▄▀▄▀ █▀█ █ ▀█ █▄▀ ██▄ █▀▄",
	}
	logoString := strings.Join(logo, "\n")
	firstPage := pages.Jobs
	return model{
		nomadToken:  nomadToken,
		nomadUrl:    nomadUrl,
		currentPage: firstPage,
		header:      header.New(logoString, nomadUrl, keymap.GetPageKeyHelp(firstPage)),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	dev.Debug(fmt.Sprintf("main %T", msg))
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keymap.KeyMap.Exit):
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
			m.loglinePage = logline.New("", msg.Width, pageHeight)
			m.initialized = true
		} else {
			m.jobsPage.SetWindowSize(msg.Width, pageHeight)
			m.allocationsPage.SetWindowSize(msg.Width, pageHeight)
			m.logsPage.SetWindowSize(msg.Width, pageHeight)
			m.loglinePage.SetWindowSize(msg.Width, pageHeight)
		}

	// this is how subcomponents currently tell main model to update the parent state
	case message.ChangePageMsg:
		newPage := msg.NewPage
		m.setPage(newPage)
		m.header.KeyHelp = keymap.GetPageKeyHelp(newPage)

		switch newPage {
		case pages.Jobs:
			m.jobsPage.Loading = true
			return m, jobs.FetchJobs(m.nomadUrl, m.nomadToken)

		case pages.Allocations:
			jobID := m.jobsPage.LastSelectedJobID
			m.allocationsPage.SetJobID(jobID)
			m.allocationsPage.Loading = true
			return m, allocations.FetchAllocations(m.nomadUrl, m.nomadToken, jobID)

		case pages.Logs:
			m.setPage(pages.Logs)
			allocID, taskName := m.allocationsPage.LastSelectedAllocID, m.allocationsPage.LastSelectedTaskName
			m.logsPage.SetAllocationData(allocID, taskName)
			m.logsPage.Loading = true
			return m, logs.FetchLogs(
				m.nomadUrl,
				m.nomadToken,
				allocID,
				taskName,
				m.logsPage.LastSelectedLogType,
			)

		case pages.Logline:
			m.setPage(pages.Logline)
			m.loglinePage.SetLogline(m.logsPage.LastSelectedLogline)
		}
	}

	switch m.currentPage {
	case pages.Jobs:
		m.jobsPage, cmd = m.jobsPage.Update(msg)
		cmds = append(cmds, cmd)
	case pages.Allocations:
		m.allocationsPage, cmd = m.allocationsPage.Update(msg)
		cmds = append(cmds, cmd)
	case pages.Logs:
		m.logsPage, cmd = m.logsPage.Update(msg)
		cmds = append(cmds, cmd)
	case pages.Logline:
		m.loglinePage, cmd = m.loglinePage.Update(msg)
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
	case pages.Jobs:
		pageView = m.jobsPage.View()
	case pages.Allocations:
		pageView = m.allocationsPage.View()
	case pages.Logs:
		pageView = m.logsPage.View()
	case pages.Logline:
		pageView = m.loglinePage.View()
	}

	finalView := m.header.View() + "\n" + pageView
	return finalView
}

func (m *model) setPage(p pages.Page) {
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
	case pages.Jobs:
		return m.jobsPage.Loading
	case pages.Allocations:
		return m.allocationsPage.Loading
	case pages.Logs:
		return m.logsPage.Loading
	case pages.Logline:
		return false
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
