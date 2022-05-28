package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
	"os"
	"wander/components/header"
	"wander/components/page"
	"wander/constants"
	"wander/dev"
	"wander/keymap"
	"wander/message"
	nomad "wander/nomad"
	"wander/style"
)

type model struct {
	nomadUrl   string
	nomadToken string

	header          header.Model
	currentPage     nomad.Page
	jobsPage        page.Model
	jobSpecPage     page.Model
	allocationsPage page.Model
	allocSpecPage   page.Model
	logsPage        page.Model
	loglinePage     page.Model
	execPage        page.Model

	jobID         string
	jobNamespace  string
	allocID       string
	taskName      string
	logline       string
	logType       nomad.LogType
	execWebSocket *websocket.Conn

	width, height int
	initialized   bool
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
	dev.DebugMsg("main", msg)
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		addingQToFilter := m.currentPageFilterFocused()
		saving := m.currentPageViewportSaving()
		inTerminal := m.currentPageIsTerminal()
		editingText := addingQToFilter || saving || inTerminal

		if key.Matches(msg, keymap.KeyMap.Exit) {
			noQuit := editingText && msg.String() == "q"
			if !noQuit {
				return m, tea.Quit
			}
		}

		if !editingText || inTerminal {
			switch {
			case key.Matches(msg, keymap.KeyMap.Forward):
				if selectedPageRow, err := m.getCurrentPageModel().GetSelectedPageRow(); err == nil {
					switch m.currentPage {
					case nomad.JobsPage:
						m.jobID, m.jobNamespace = nomad.JobIDAndNamespaceFromKey(selectedPageRow.Key)
					case nomad.AllocationsPage:
						m.allocID, m.taskName = nomad.AllocIDAndTaskNameFromKey(selectedPageRow.Key)
					case nomad.LogsPage:
						m.logline = selectedPageRow.Row
					}

					nextPage := m.currentPage.Forward()
					if nextPage != m.currentPage {
						m.setPage(nextPage)
						return m, m.getCurrentPageCmd()
					}
				}

			case key.Matches(msg, keymap.KeyMap.Back):
				if !m.currentPageFilterApplied() && !saving {
					prevPage := m.currentPage.Backward()
					if prevPage != m.currentPage {
						m.getCurrentPageModel().ExitTerminal()
						m.setPage(prevPage)
						return m, m.getCurrentPageCmd()
					}
				}

			case key.Matches(msg, keymap.KeyMap.Reload):
				if m.currentPage.Loads() && !inTerminal {
					m.getCurrentPageModel().SetLoading(true)
					return m, m.getCurrentPageCmd()
				}
			}

			if key.Matches(msg, keymap.KeyMap.Spec) {
				if selectedPageRow, err := m.getCurrentPageModel().GetSelectedPageRow(); err == nil {
					switch m.currentPage {
					case nomad.JobsPage:
						m.jobID, m.jobNamespace = nomad.JobIDAndNamespaceFromKey(selectedPageRow.Key)
						m.setPage(nomad.JobSpecPage)
						return m, m.getCurrentPageCmd()
					case nomad.AllocationsPage:
						m.allocID, m.taskName = nomad.AllocIDAndTaskNameFromKey(selectedPageRow.Key)
						m.setPage(nomad.AllocSpecPage)
						return m, m.getCurrentPageCmd()
					}
				}
			}

			if key.Matches(msg, keymap.KeyMap.Exec) && m.currentPage == nomad.AllocationsPage {
				if selectedPageRow, err := m.getCurrentPageModel().GetSelectedPageRow(); err == nil {
					m.allocID, m.taskName = nomad.AllocIDAndTaskNameFromKey(selectedPageRow.Key)
					m.setPage(nomad.ExecPage)
					return m, m.getCurrentPageCmd()
				}
			} else if m.currentPage == nomad.LogsPage {
				switch {
				case key.Matches(msg, keymap.KeyMap.StdOut):
					if m.logType != nomad.StdOut {
						m.logType = nomad.StdOut
						m.getCurrentPageModel().SetViewportStyle(style.ViewportHeaderStyle, style.StdOut)
						m.getCurrentPageModel().SetLoading(true)
						return m, m.getCurrentPageCmd()
					}

				case key.Matches(msg, keymap.KeyMap.StdErr):
					if m.logType != nomad.StdErr {
						m.logType = nomad.StdErr
						stdErrHeaderStyle := style.ViewportHeaderStyle.Copy().Inherit(style.StdErr)
						m.getCurrentPageModel().SetViewportStyle(stdErrHeaderStyle, style.StdErr)
						m.getCurrentPageModel().SetLoading(true)
						return m, m.getCurrentPageCmd()
					}
				}
			}
		}

	case message.ErrMsg:
		m.err = msg
		return m, nil

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if !m.initialized {
			m.initialize()
			cmds = append(cmds, m.getCurrentPageCmd())
		} else {
			m.setPageWindowSize()
		}

	case nomad.PageLoadedMsg:
		m.setPage(msg.Page)
		currentPageModel := m.getCurrentPageModel()
		currentPageModel.SetHeader(msg.TableHeader)
		currentPageModel.SetAllPageData(msg.AllPageData)
		currentPageModel.SetLoading(false)
		currentPageModel.SetViewportXOffset(0)
		if m.currentPage == nomad.LogsPage {
			m.logsPage.SetViewportCursorToBottom()
		}

	case page.TerminalEnterMsg:
		dev.Debug(msg.Cmd)
		if msg.Init {
			dev.Debug("INIT")
			return m, nomad.InitiateExecWebSocketConnection(m.nomadUrl, m.nomadToken, m.allocID, m.taskName, msg.Cmd)
		} else {
			dev.Debug("SESSION")
			cmds = append(cmds, nomad.SendAndReadExecWebSocketMessage(m.execWebSocket, msg.Cmd))
		}

	case nomad.ExecWebSocketConnectedMsg:
		dev.Debug("CONNECTED")
		m.execWebSocket = msg.WebSocketConnection
		cmds = append(cmds, nomad.ReadExecWebSocketNextMessage(m.execWebSocket))

	case nomad.ExecWebSocketResponseMsg:
		dev.Debug("WS RESPONSE")
		dev.Debug(msg.StdOut)
		var newPageData []page.Row
		stdOutRows := strings.Split(msg.StdOut, "\n")
		stdErrRows := strings.Split(msg.StdErr, "\n")
		for _, row := range append(stdOutRows, stdErrRows...) {
			if strings.TrimSpace(row) != "" {
				newPageData = append(newPageData, page.Row{Row: row})
			}
		}
		m.getCurrentPageModel().AppendPageData(newPageData)
	}

	currentPageModel := m.getCurrentPageModel()
	*currentPageModel, cmd = currentPageModel.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	} else if !m.initialized {
		return ""
	}

	pageView := m.header.View() + "\n" + m.getCurrentPageModel().View()

	return pageView
}

func (m *model) initialize() {
	pageHeight := m.getPageHeight()
	m.jobsPage = page.New(m.width, pageHeight, m.getFilterPrefix(nomad.JobsPage), nomad.JobsPage.LoadingString(), true, false, false)
	m.jobSpecPage = page.New(m.width, pageHeight, m.getFilterPrefix(nomad.JobSpecPage), nomad.JobSpecPage.LoadingString(), false, true, false)
	m.allocationsPage = page.New(m.width, pageHeight, m.getFilterPrefix(nomad.AllocationsPage), nomad.AllocationsPage.LoadingString(), true, false, false)
	m.allocSpecPage = page.New(m.width, pageHeight, m.getFilterPrefix(nomad.AllocSpecPage), nomad.AllocSpecPage.LoadingString(), false, true, false)
	m.logsPage = page.New(m.width, pageHeight, m.getFilterPrefix(nomad.LogsPage), nomad.LogsPage.LoadingString(), true, false, false)
	m.loglinePage = page.New(m.width, pageHeight, m.getFilterPrefix(nomad.LoglinePage), nomad.LoglinePage.LoadingString(), false, true, false)
	m.execPage = page.New(m.width, pageHeight, m.getFilterPrefix(nomad.ExecPage), nomad.ExecPage.LoadingString(), false, true, true)
	m.initialized = true
}

func (m *model) setPageWindowSize() {
	m.jobsPage.SetWindowSize(m.width, m.getPageHeight())
	m.jobSpecPage.SetWindowSize(m.width, m.getPageHeight())
	m.allocationsPage.SetWindowSize(m.width, m.getPageHeight())
	m.allocSpecPage.SetWindowSize(m.width, m.getPageHeight())
	m.logsPage.SetWindowSize(m.width, m.getPageHeight())
	m.loglinePage.SetWindowSize(m.width, m.getPageHeight())
	m.execPage.SetWindowSize(m.width, m.getPageHeight())
}

func (m *model) setPage(page nomad.Page) {
	m.currentPage = page
	m.header.KeyHelp = nomad.GetPageKeyHelp(page)
	m.getCurrentPageModel().SetFilterPrefix(m.getFilterPrefix(page))
	if page.Loads() {
		m.getCurrentPageModel().SetLoading(true)
	} else {
		m.getCurrentPageModel().SetLoading(false)
	}
}

func (m *model) getCurrentPageModel() *page.Model {
	switch m.currentPage {
	case nomad.JobsPage:
		return &m.jobsPage
	case nomad.JobSpecPage:
		return &m.jobSpecPage
	case nomad.AllocationsPage:
		return &m.allocationsPage
	case nomad.AllocSpecPage:
		return &m.allocSpecPage
	case nomad.LogsPage:
		return &m.logsPage
	case nomad.LoglinePage:
		return &m.loglinePage
	case nomad.ExecPage:
		return &m.execPage
	default:
		panic("current page model not found")
	}
}

func (m *model) getCurrentPageCmd() tea.Cmd {
	switch m.currentPage {
	case nomad.JobsPage:
		return nomad.FetchJobs(m.nomadUrl, m.nomadToken)
	case nomad.JobSpecPage:
		return nomad.FetchJobSpec(m.nomadUrl, m.nomadToken, m.jobID, m.jobNamespace)
	case nomad.AllocationsPage:
		return nomad.FetchAllocations(m.nomadUrl, m.nomadToken, m.jobID, m.jobNamespace)
	case nomad.AllocSpecPage:
		return nomad.FetchAllocSpec(m.nomadUrl, m.nomadToken, m.allocID)
	case nomad.LogsPage:
		return nomad.FetchLogs(m.nomadUrl, m.nomadToken, m.allocID, m.taskName, m.logType)
	case nomad.LoglinePage:
		return nomad.FetchLogLine(m.logline)
	case nomad.ExecPage:
		// don't actually fetch the websocket right away, as it needs an initial command
		return func() tea.Msg { return nomad.PageLoadedMsg{Page: nomad.ExecPage} }
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

func (m model) currentPageIsTerminal() bool {
	return m.getCurrentPageModel().IsTerminal()
}

func (m model) getFilterPrefix(page nomad.Page) string {
	return page.GetFilterPrefix(m.jobID, m.taskName, m.allocID)
}

func main() {
	program := tea.NewProgram(initialModel(), tea.WithAltScreen())

	dev.Debug("~STARTING UP~")
	if err := program.Start(); err != nil {
		fmt.Printf("Error on wander startup: %v", err)
		os.Exit(1)
	}
}
