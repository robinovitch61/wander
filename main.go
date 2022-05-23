package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"os"
	"strings"
	"wander/components/header"
	"wander/components/page"
	"wander/components/toast"
	"wander/components/viewport"
	"wander/constants"
	"wander/dev"
	"wander/keymap"
	"wander/message"
	"wander/nomad"
	"wander/style"
)

type model struct {
	nomadUrl        string
	nomadToken      string
	header          header.Model
	currentPage     nomad.Page
	jobsPage        page.Model
	allocationsPage page.Model
	logsPage        page.Model
	loglinePage     page.Model
	jobID           string
	allocID         string
	taskName        string
	logline         string
	logType         nomad.LogType
	width, height   int
	initialized     bool
	toastMessage    string
	showToast       bool
	err             error
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

	firstPage := nomad.JobsPage
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
		// always exit if desired, or don't respond if editing filter or saving
		if key.Matches(msg, keymap.KeyMap.Exit) {
			addingQToFilter := m.currentPageFilterFocused()
			saving := m.currentPageViewportSaving()
			typingQWhileFilteringOrSaving := (addingQToFilter || saving) && msg.String() == "q"
			if !typingQWhileFilteringOrSaving {
				return m, tea.Quit
			}
		}

		// TODO LEO: make actions during loading safe in the future
		if m.currentPageLoading() {
			return m, nil
		}

		if !m.currentPageFilterFocused() && !m.currentPageViewportSaving() {
			switch {
			case key.Matches(msg, keymap.KeyMap.Forward):
				if selectedPageRow, err := m.getCurrentPageModel().GetSelectedPageRow(); err == nil {
					switch m.currentPage {
					case nomad.JobsPage:
						m.jobID = nomad.JobIDFromKey(selectedPageRow.Key)
					case nomad.AllocationsPage:
						m.allocID, m.taskName = nomad.AllocIDAndTaskNameFromKey(selectedPageRow.Key)
					case nomad.LogsPage:
						m.logline = selectedPageRow.Row
					}

					nextPage := m.currentPage.Forward()
					if nextPage != m.currentPage {
						m.setPage(nextPage)
						m.getCurrentPageModel().SetLoading(true)
						return m, m.getCurrentPageLoadCmd()
					}
				}

			case key.Matches(msg, keymap.KeyMap.Back):
				if !m.currentPageFilterApplied() {
					prevPage := m.currentPage.Backward()
					if prevPage != m.currentPage {
						m.setPage(prevPage)
						m.getCurrentPageModel().SetLoading(true)
						return m, m.getCurrentPageLoadCmd()
					}
				}

			case key.Matches(msg, keymap.KeyMap.Reload):
				m.getCurrentPageModel().SetLoading(true)
				return m, m.getCurrentPageLoadCmd()
			}
		}

	case message.ErrMsg:
		m.err = msg
		return m, nil

	case toast.ToastTimeoutMsg:
		m.showToast = false
		return m, nil

	case viewport.SaveStatusMsg:
		if msg.Err != "" {
			m.toastMessage = style.ErrorToast.Width(m.width).Render(fmt.Sprintf("Error: %s", msg.Err))
		} else {
			m.toastMessage = style.SuccessToast.Width(m.width).Render(msg.SuccessMessage)
		}
		m.showToast = true
		cmds = append(cmds, toast.GetToastTimeoutCmd())
		return m, tea.Batch(cmds...)

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if !m.initialized {
			m.initialize()
			cmds = append(cmds, m.getCurrentPageLoadCmd())
		} else {
			m.setPageWindowSize()
		}

	case nomad.PageLoadedMsg:
		dev.Debug(msg.Page.String())
		m.setPage(msg.Page)
		m.getCurrentPageModel().SetHeader(msg.TableHeader)
		m.getCurrentPageModel().SetAllPageData(msg.AllPageData)
		m.getCurrentPageModel().SetLoading(false)
	}

	currentPageModel := m.getCurrentPageModel()
	*currentPageModel, cmd = currentPageModel.Update(msg)
	cmds = append(cmds, cmd)
	// switch m.currentPage {
	// case nomad.JobsPage:
	// 	m.jobsPage, cmd = m.jobsPage.Update(msg)
	// 	cmds = append(cmds, cmd)
	// case nomad.AllocationsPage:
	// 	m.allocationsPage, cmd = m.allocationsPage.Update(msg)
	// 	cmds = append(cmds, cmd)
	// case nomad.LogsPage:
	// 	m.logsPage, cmd = m.logsPage.Update(msg)
	// 	cmds = append(cmds, cmd)
	// case nomad.LoglinePage:
	// 	m.loglinePage, cmd = m.loglinePage.Update(msg)
	// 	cmds = append(cmds, cmd)
	// default:
	// 	panic("page not found")
	// }

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	} else if !m.initialized {
		return ""
	}

	pageView := m.header.View() + "\n" + m.getCurrentPageModel().View()

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

func (m *model) setPage(page nomad.Page) {
	m.currentPage = page
	m.header.KeyHelp = nomad.GetPageKeyHelp(page)
}

func (m *model) getCurrentPageModel() *page.Model {
	switch m.currentPage {
	case nomad.JobsPage:
		return &m.jobsPage
	case nomad.AllocationsPage:
		return &m.allocationsPage
	case nomad.LogsPage:
		return &m.logsPage
	case nomad.LoglinePage:
		return &m.loglinePage
	default:
		panic("current page model not found")
	}
}

func (m model) getCurrentPageLoadCmd() tea.Cmd {
	switch m.currentPage {
	case nomad.JobsPage:
		return nomad.FetchJobs(m.nomadUrl, m.nomadToken)
	case nomad.AllocationsPage:
		return nomad.FetchAllocations(m.nomadUrl, m.nomadToken, m.jobID)
	case nomad.LogsPage:
		return nomad.FetchLogs(m.nomadUrl, m.nomadToken, m.allocID, m.taskName, m.logType)
	case nomad.LoglinePage:
		return nomad.FetchLogLine(m.logline)
	default:
		panic("page load command not found")
	}
}

func (m model) getPageHeight() int {
	return m.height - m.header.ViewHeight()
}

func (m model) currentPageLoading() bool {
	return m.getCurrentPageModel().Loading()
}

func (m model) currentPageFilterFocused() bool {
	return m.getCurrentPageModel().FilterFocused()
}

func (m model) currentPageFilterApplied() bool {
	return m.getCurrentPageModel().FilterApplied()
}

func (m model) currentPageViewportSaving() bool {
	return m.getCurrentPageModel().ViewportSaving()
}

func main() {
	program := tea.NewProgram(initialModel(), tea.WithAltScreen())

	dev.Debug("~STARTING UP~")
	if err := program.Start(); err != nil {
		fmt.Printf("Error on wander startup: %v", err)
		os.Exit(1)
	}
}
