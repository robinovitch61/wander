package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"os"
	"strings"
	"time"
	"wander/components/header"
	"wander/components/page"
	"wander/components/viewport"
	"wander/constants"
	"wander/dev"
	"wander/keymap"
	"wander/message"
	"wander/pages"
	"wander/pages/jobs"
	"wander/pages/logs"
	"wander/style"
)

type toastTimeoutMsg struct{}

type model struct {
	nomadToken       string
	nomadUrl         string
	currentPage      *page.Model
	jobsPage         page.Model
	allocationsPage  page.Model
	logsPage         page.Model
	loglinePage      page.Model
	selectedJobID    string
	selectedAllocID  string
	selectedTaskName string
	selectedLogType  logs.LogType
	header           header.Model
	width, height    int
	initialized      bool
	toastMessage     string
	showToast        bool
	err              error
}

func initialModel() model {
	nomadToken := os.Getenv(constants.NomadTokenEnvVariable)
	if nomadToken == "" {
		fmt.Printf("Set environment variable %s\n", constants.NomadTokenEnvVariable)
		os.Exit(1)
	}

	nomadUrl := os.Getenv(constants.NomadUrlEnvVariable)
	if nomadUrl == "" {
		fmt.Printf("Set environment variable %s\n", constants.NomadUrlEnvVariable)
		os.Exit(1)
	}

<<<<<<< HEAD
	firstPage := pages.Jobs
||||||| parent of 20ee616 (Broken)
	logo := []string{
		"█ █ █ █▀█ █▄ █ █▀▄ █▀▀ █▀█",
		"▀▄▀▄▀ █▀█ █ ▀█ █▄▀ ██▄ █▀▄",
	}
	logoString := strings.Join(logo, "\n")
	firstPage := pages.Jobs
=======
	initialHeader := header.New(constants.LogoString, nomadUrl, "")
>>>>>>> 20ee616 (Broken)
	return model{
<<<<<<< HEAD
		nomadToken:  nomadToken,
		nomadUrl:    nomadUrl,
		currentPage: firstPage,
		header:      header.New(constants.Logo, nomadUrl, keymap.GetPageKeyHelp(firstPage)),
||||||| parent of 20ee616 (Broken)
		nomadToken:  nomadToken,
		nomadUrl:    nomadUrl,
		currentPage: firstPage,
		header:      header.New(logoString, nomadUrl, keymap.GetPageKeyHelp(firstPage)),
=======
		nomadToken: nomadToken,
		nomadUrl:   nomadUrl,
		header:     initialHeader,
>>>>>>> 20ee616 (Broken)
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
			dev.Debug("HERE")
			dev.Debug(msg.String())
			if !(m.currentPageFilterFocused() && msg.String() == "q") {
				return m, tea.Quit
			}

		default:
			if m.currentPage.Loading {
				return m, nil
			}
		}

	case message.ErrMsg:
		m.err = msg
		return m, nil

	case toastTimeoutMsg:
		m.showToast = false
		return m, nil

	case viewport.SaveStatusMsg:
		dev.Debug(msg.SuccessMessage)
		if msg.Err != "" {
			m.toastMessage = style.ErrorToast.Width(m.width).Render(fmt.Sprintf("Error: %s", msg.Err))
		} else {
			m.toastMessage = style.SuccessToast.Width(m.width).Render(msg.SuccessMessage)
		}
		m.showToast = true
		cmds = append(
			cmds,
			tea.Tick(constants.ToastDuration, func(t time.Time) tea.Msg { return toastTimeoutMsg{} }),
		)
		return m, tea.Batch(cmds...)

	case tea.WindowSizeMsg:
		m.setWindowSize(msg.Width, msg.Height)
		pageHeight := m.getPageHeight()

		if !m.initialized {
			m.jobsPage = page.New(msg.Width, pageHeight, "Jobs", "Loading jobs...")
			m.allocationsPage = page.New(msg.Width, pageHeight, "Allocations", "")
			m.logsPage = page.New(msg.Width, pageHeight, "Logs", "")
			m.loglinePage = page.New(msg.Width, pageHeight, "Log Line", "")

			m.currentPage = &m.jobsPage
			m.initialized = true
			cmds = append(cmds, jobs.FetchJobs(m.nomadUrl, m.nomadToken))
		} else {
<<<<<<< HEAD
			m.jobsPage.SetWindowSize(msg.Width, pageHeight)
			m.allocationsPage.SetWindowSize(msg.Width, pageHeight)
			m.logsPage.SetWindowSize(msg.Width, pageHeight)
			m.loglinePage.SetWindowSize(msg.Width, pageHeight)
||||||| parent of 20ee616 (Broken)
			m.getCurrentPageModel().SetWindowSize(msg.Width, pageHeight)
=======
			m.currentPage.SetWindowSize(msg.Width, pageHeight)
>>>>>>> 20ee616 (Broken)
		}

	case jobs.NomadJobsMsg:
		m.currentPage = &m.jobsPage
		m.currentPage.SetPageData(jobTable.HeaderRows, jobTable.ContentRows)
		m.currentPage.Loading = false

		// // this is how subcomponents currently tell main model to update the parent state
		// case pages.ChangePageMsg:
		// 	newPage := msg.NewPage
		// 	m.setPage(newPage)
		//
		// 	switch newPage {
		// 	case pages.Jobs:
		// 		m.jobsPage.Loading = true
		// 		return m, jobs.FetchJobs(m.nomadUrl, m.nomadToken)
		//
		// 	case pages.Allocations:
		// 		jobID := m.jobsPage.LastSelectedJobID
		// 		m.allocationsPage.SetJobID(jobID)
		// 		m.allocationsPage.Loading = true
		// 		return m, allocations.FetchAllocations(m.nomadUrl, m.nomadToken, jobID)
		//
		// 	case pages.Logs:
		// 		m.setPage(pages.Logs)
		// 		m.logsPage.ResetXOffset()
		// 		allocID, taskName := m.allocationsPage.LastSelectedAllocID, m.allocationsPage.LastSelectedTaskName
		// 		m.logsPage.SetAllocationData(allocID, taskName)
		// 		m.logsPage.Loading = true
		// 		return m, logs.FetchLogs(
		// 			m.nomadUrl,
		// 			m.nomadToken,
		// 			allocID,
		// 			taskName,
		// 			m.logsPage.LastSelectedLogType,
		// 		)
		//
		// 	case pages.Logline:
		// 		m.setPage(pages.Logline)
		// 		m.loglinePage.SetAllocationData(m.allocationsPage.LastSelectedAllocID, m.allocationsPage.LastSelectedTaskName)
		// 		m.loglinePage.SetLogline(m.logsPage.LastSelectedLogline)
		// 	}
	}

	*m.currentPage, cmd = m.currentPage.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	} else if !m.initialized {
		return ""
	}

<<<<<<< HEAD
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
||||||| parent of 20ee616 (Broken)
	pageView := m.getCurrentPageModel().View()
=======
	pageView := m.currentPage.View()
>>>>>>> 20ee616 (Broken)

	pageView = m.header.View() + "\n" + pageView

	if m.showToast {
		lines := strings.Split(pageView, "\n")
		lines = lines[:len(lines)-lipgloss.Height(m.toastMessage)]
		pageView = strings.Join(lines, "\n") + "\n" + m.toastMessage
	}

	return pageView
}


// func (m *model) setPage(p pages.Page) {
// 	m.currentPage = p
// 	// m.setHeaderKeyHelp()
// }

func (m *model) setWindowSize(width, height int) {
	m.width, m.height = width, height
}

func (m model) getPageHeight() int {
	return m.height - m.header.ViewHeight()
}

// func (m model) currentPageLoading() bool {
// 	switch m.currentPage {
// 	case pages.Jobs:
// 		return m.jobsPage.Loading
// 	case pages.Allocations:
// 		return m.allocationsPage.Loading
// 	case pages.Logs:
// 		return m.logsPage.Loading
// 	case pages.Logline:
// 		return false
// 	}
// 	return true
// }

func (m model) currentPageFilterFocused() bool {
	switch m.currentPage {
	case pages.Jobs:
		return m.jobsPage.FilterFocused()
	case pages.Allocations:
		return m.allocationsPage.FilterFocused()
	case pages.Logs:
		return m.logsPage.FilterFocused()
	case pages.Logline:
		return m.loglinePage.FilterFocused()
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
