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
	"wander/components/viewport"
	"wander/dev"
	"wander/keymap"
	"wander/message"
	"wander/pages"
	"wander/pages/allocations"
	"wander/pages/jobs"
	"wander/pages/logline"
	"wander/pages/logs"
	"wander/style"
)

var (
	nomadTokenEnvVariable = "NOMAD_TOKEN"
	nomadUrlEnvVariable   = "NOMAD_ADDR"
)

type toastTimeoutMsg struct{}

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
	toastMessage     string
	showToast        bool
	err              error
}

func initialModel() model {
	nomadToken := os.Getenv(nomadTokenEnvVariable)
	if nomadToken == "" {
		fmt.Printf("Set environment variable %s\n", nomadTokenEnvVariable)
		os.Exit(1)
	}

	nomadUrl := os.Getenv(nomadUrlEnvVariable)
	if nomadUrl == "" {
		fmt.Printf("Set environment variable %s\n", nomadUrlEnvVariable)
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

	case toastTimeoutMsg:
		m.showToast = false
		return m, nil

	case viewport.SaveStatusMsg:
		dev.Debug(msg.SuccessMessage)
		if msg.Err != "" {
			m.toastMessage = style.ErrorToast.Width(m.width).Render(msg.Err)
		} else {
			m.toastMessage = style.SuccessToast.Width(m.width).Render(msg.SuccessMessage)
		}
		m.showToast = true
		cmds = append(
			cmds,
			tea.Tick(time.Second*5, func(t time.Time) tea.Msg { return toastTimeoutMsg{} }),
		)
		return m, tea.Batch(cmds...)

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
			m.getCurrentPageModel().SetWindowSize(msg.Width, pageHeight)
		}

	// this is how subcomponents currently tell main model to update the parent state
	case pages.ChangePageMsg:
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
			m.logsPage.ResetXOffset()
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
			m.loglinePage.SetAllocationData(m.allocationsPage.LastSelectedAllocID, m.allocationsPage.LastSelectedTaskName)
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

	pageView := m.getCurrentPageModel().View()

	pageView = m.header.View() + "\n" + pageView

	if m.showToast {
		lines := strings.Split(pageView, "\n")
		lines = lines[:len(lines)-lipgloss.Height(m.toastMessage)]
		pageView = strings.Join(lines, "\n") + "\n" + m.toastMessage
	}

	return pageView
}

type currentPageModel interface {
	View() string
	SetWindowSize(width, height int)
}

func (m model) getCurrentPageModel() currentPageModel {
	switch m.currentPage {
	case pages.Jobs:
		return &m.jobsPage
	case pages.Allocations:
		return &m.allocationsPage
	case pages.Logs:
		return &m.logsPage
	case pages.Logline:
		return &m.loglinePage
	}
	return nil
}

func (m *model) setPage(p pages.Page) {
	m.currentPage = p
	// m.setHeaderKeyHelp()
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
