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
	"wander/pages/allocations"
	"wander/pages/jobs"
	"wander/pages/logs"
	"wander/style"
)

type toastTimeoutMsg struct{}

type model struct {
	nomadUrl    string
	nomadToken  string
	header      header.Model
	currentPage pages.Page
	jobsPage    page.Model
	// jobsData        []jobs.jobResponseEntry
	allocationsPage page.Model
	// allocationsData []allocations.AllocationResponseEntry
	logsPage page.Model
	// logsData        []logs.LogRow
	loglinePage page.Model
	// loglineData     []logline.LoglineRow
	jobID         string
	allocID       string
	taskName      string
	logType       logs.LogType
	width, height int
	initialized   bool
	toastMessage  string
	showToast     bool
	err           error
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

	firstPage := pages.Jobs
	initialHeader := header.New(constants.LogoString, nomadUrl, "")

	return model{
		nomadUrl:    nomadUrl,
		nomadToken:  nomadToken,
		header:      initialHeader,
		currentPage: firstPage,
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
		// always exit if desired, or don't respond if loading
		if key.Matches(msg, keymap.KeyMap.Exit) {
			if addingQToFilter := m.currentPageFilterFocused() && msg.String() == "q"; !addingQToFilter {
				return m, tea.Quit
			}
		} else if m.currentPageLoading() {
			return m, nil
		}

		switch {
		case key.Matches(msg, keymap.KeyMap.Forward):
			switch m.currentPage {
			case pages.Jobs:
				if selectedPageRow, err := m.jobsPage.GetSelectedPageRow(); err == nil {
					m.jobID = jobs.JobIDFromKey(selectedPageRow.Key)
					m.currentPage = pages.Allocations
					m.allocationsPage.Loading = true
					return m, allocations.FetchAllocations(m.nomadUrl, m.nomadToken, m.jobID)
				}
			default:
				panic("IMPLEMENT ME")
			}
		default:
			panic("IMPLEMENT ME")
		}

	case message.ErrMsg:
		m.err = msg
		return m, nil

	case toastTimeoutMsg:
		m.showToast = false
		return m, nil

	case viewport.SaveStatusMsg:
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
		m.width, m.height = msg.Width, msg.Height
		if !m.initialized {
			m.initialize()
			cmds = append(cmds, jobs.FetchJobs(m.nomadUrl, m.nomadToken))
		} else {
			m.setPageWindowSize()
		}

	case jobs.NomadJobsMsg:
		m.currentPage = pages.Jobs
		m.jobsPage.SetHeader(msg.TableHeader)
		m.jobsPage.SetAllPageData(msg.AllPageData)

		// // this is how subcomponents currently tell main model to update the parent state
		// case pages.ChangePageMsg:
		// 	newPage := msg.NewPage
		// 	m.setPage(newPage)
		//
		// 	switch newPage {
		// 	case pages.Jobs:
		// 		m.jobsPage.loading = true
		// 		return m, jobs.FetchJobs(m.nomadUrl, m.nomadToken)
		//
		// 	case pages.Allocations:
		// 		jobID := m.jobsPage.LastSelectedJobID
		// 		m.allocationsPage.SetJobID(jobID)
		// 		m.allocationsPage.loading = true
		// 		return m, allocations.FetchAllocations(m.nomadUrl, m.nomadToken, jobID)
		//
		// 	case pages.Logs:
		// 		m.setPage(pages.Logs)
		// 		m.logsPage.ResetXOffset()
		// 		allocID, taskName := m.allocationsPage.LastSelectedAllocID, m.allocationsPage.LastSelectedTaskName
		// 		m.logsPage.SetAllocationData(allocID, taskName)
		// 		m.logsPage.loading = true
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
	default:
		panic("page not found")
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	} else if !m.initialized {
		return ""
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

	pageView = m.header.View() + "\n" + pageView

	if m.showToast {
		lines := strings.Split(pageView, "\n")
		lines = lines[:len(lines)-lipgloss.Height(m.toastMessage)]
		pageView = strings.Join(lines, "\n") + "\n" + m.toastMessage
	}

	return pageView
}

func (m *model) initialize() {
	pageHeight := m.getPageHeight()
	m.jobsPage = page.New(m.width, pageHeight, "Jobs", "loading jobs...")
	m.allocationsPage = page.New(m.width, pageHeight, "Allocations", "")
	m.logsPage = page.New(m.width, pageHeight, "Logs", "")
	m.loglinePage = page.New(m.width, pageHeight, "Log Line", "")
	m.initialized = true
}

func (m *model) setPageWindowSize() {
	m.getCurrentPageModel().SetWindowSize(m.width, m.getPageHeight())
}

func (m model) getCurrentPageModel() *page.Model {
	switch m.currentPage {
	case pages.Jobs:
		return &m.jobsPage
	case pages.Allocations:
		return &m.allocationsPage
	case pages.Logs:
		return &m.logsPage
	case pages.Logline:
		return &m.loglinePage
	default:
		panic("page not found")
	}
}

func (m model) getPageHeight() int {
	return m.height - m.header.ViewHeight()
}

func (m model) currentPageLoading() bool {
	return m.getCurrentPageModel().Loading
}

func (m model) currentPageFilterFocused() bool {
	return m.getCurrentPageModel().FilterFocused()
}

func main() {
	program := tea.NewProgram(initialModel(), tea.WithAltScreen())

	dev.Debug("~STARTING UP~")
	if err := program.Start(); err != nil {
		fmt.Printf("Error on wander startup: %v", err)
		os.Exit(1)
	}
}
