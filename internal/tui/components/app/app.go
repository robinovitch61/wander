package app

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
	"github.com/robinovitch61/wander/internal/dev"
	"github.com/robinovitch61/wander/internal/tui/components/header"
	"github.com/robinovitch61/wander/internal/tui/components/page"
	"github.com/robinovitch61/wander/internal/tui/constants"
	"github.com/robinovitch61/wander/internal/tui/formatter"
	"github.com/robinovitch61/wander/internal/tui/keymap"
	"github.com/robinovitch61/wander/internal/tui/message"
	"github.com/robinovitch61/wander/internal/tui/nomad"
	"github.com/robinovitch61/wander/internal/tui/style"
	"os"
	"strings"
	"time"
)

type Model struct {
	nomadUrl   string
	nomadToken string

	header      header.Model
	currentPage nomad.Page
	pageModels  map[nomad.Page]*page.Model

	jobID        string
	jobNamespace string
	allocID      string
	taskName     string
	logline      string
	logType      nomad.LogType

	execWebSocket       *websocket.Conn
	execPty             *os.File
	inPty               bool
	webSocketConnected  bool
	lastCommandFinished struct{ stdOut, stdErr bool }

	updateInterval time.Duration

	width, height int
	initialized   bool
	err           error
}

func InitialModel(version, sha, url, token string, updateSeconds int) Model {
	firstPage := nomad.JobsPage
	initialHeader := header.New(constants.LogoString, url, getVersionString(version, sha), nomad.GetPageKeyHelp(firstPage, false, false, false, false, false, false))
	updateInterval := time.Second * time.Duration(updateSeconds)

	return Model{
		nomadUrl:       url,
		nomadToken:     token,
		header:         initialHeader,
		currentPage:    firstPage,
		updateInterval: updateInterval,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	dev.Debug(fmt.Sprintf("main %T", msg))
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	currentPageModel := m.getCurrentPageModel()
	if currentPageModel != nil && currentPageModel.EnteringInput() {
		*currentPageModel, cmd = currentPageModel.Update(msg)
		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// always exit if desired, or don't respond if typing "q" legitimately in some text input
		if key.Matches(msg, keymap.KeyMap.Exit) {
			addingQToFilter := m.currentPageFilterFocused()
			saving := m.currentPageViewportSaving()
			enteringInput := currentPageModel != nil && currentPageModel.EnteringInput()
			typingQLegitimately := msg.String() == "q" && (addingQToFilter || saving || enteringInput || m.inPty)
			if !typingQLegitimately {
				return m, tea.Quit
			}
		}

		if m.inPty {
			var keypress string
			if key.Matches(msg, keymap.KeyMap.Back) {
				m.setInPty(false)
			} else {
				keypress = nomad.GetKeypress(msg)
			}
			return m, nomad.SendWebSocketMessage(m.execWebSocket, keypress)
		}

		if !m.currentPageFilterFocused() && !m.currentPageViewportSaving() {
			switch {
			case key.Matches(msg, keymap.KeyMap.Forward):
				if selectedPageRow, err := m.getCurrentPageModel().GetSelectedPageRow(); err == nil {
					switch m.currentPage {
					case nomad.JobsPage:
						m.jobID, m.jobNamespace = nomad.JobIDAndNamespaceFromKey(selectedPageRow.Key)
					case nomad.AllocationsPage:
						allocInfo, err := nomad.AllocationInfoFromKey(selectedPageRow.Key)
						if err != nil {
							m.err = err
							return m, nil
						}
						m.allocID, m.taskName = allocInfo.AllocID, allocInfo.TaskName
					case nomad.ExecPage:
						if !m.getCurrentPageModel().EnteringInput() && m.webSocketConnected {
							m.setInPty(true)
						}
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
				if !m.currentPageFilterApplied() {
					switch m.currentPage {
					case nomad.ExecPage:
						if !m.getCurrentPageModel().EnteringInput() {
							cmds = append(cmds, nomad.CloseWebSocket(m.execWebSocket))
						}
						m.getCurrentPageModel().SetDoesNeedNewInput()
					}

					backPage := m.currentPage.Backward()
					if backPage != m.currentPage {
						m.setPage(backPage)
						cmds = append(cmds, m.getCurrentPageCmd())
						return m, tea.Batch(cmds...)
					}
				}

			case key.Matches(msg, keymap.KeyMap.Reload):
				if m.currentPage.DoesReload() {
					m.getCurrentPageModel().SetLoading(true)
					return m, m.getCurrentPageCmd()
				}
			}

			if key.Matches(msg, keymap.KeyMap.Exec) {
				if selectedPageRow, err := m.getCurrentPageModel().GetSelectedPageRow(); err == nil {
					if m.currentPage == nomad.AllocationsPage {
						allocInfo, err := nomad.AllocationInfoFromKey(selectedPageRow.Key)
						if err != nil {
							m.err = err
							return m, nil
						}
						if allocInfo.Running {
							m.allocID, m.taskName = allocInfo.AllocID, allocInfo.TaskName
							m.setPage(nomad.ExecPage)
							return m, m.getCurrentPageCmd()
						}
					}
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
						allocInfo, err := nomad.AllocationInfoFromKey(selectedPageRow.Key)
						if err != nil {
							m.err = err
							return m, nil
						}
						m.allocID, m.taskName = allocInfo.AllocID, allocInfo.TaskName
						m.setPage(nomad.AllocSpecPage)
						return m, m.getCurrentPageCmd()
					}
				}
			}

			if m.currentPage == nomad.LogsPage {
				switch {
				case key.Matches(msg, keymap.KeyMap.StdOut):
					if !m.currentPageLoading() && m.logType != nomad.StdOut {
						m.logType = nomad.StdOut
						m.getCurrentPageModel().SetViewportStyle(style.ViewportHeaderStyle, style.StdOut)
						m.getCurrentPageModel().SetLoading(true)
						return m, m.getCurrentPageCmd()
					}

				case key.Matches(msg, keymap.KeyMap.StdErr):
					if !m.currentPageLoading() && m.logType != nomad.StdErr {
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
			if m.currentPage == nomad.ExecPage {
				cmds = append(cmds, nomad.ResizeTty(m.execWebSocket, m.width, m.getCurrentPageModel().ViewportHeight()))
			}
		}

	case nomad.PageLoadedMsg:
		if msg.Page == m.currentPage {
			m.getCurrentPageModel().SetHeader(msg.TableHeader)
			m.getCurrentPageModel().SetAllPageData(msg.AllPageData)
			m.getCurrentPageModel().SetLoading(false)
			m.getCurrentPageModel().SetViewportXOffset(0)

			switch m.currentPage {
			case nomad.LogsPage:
				m.getCurrentPageModel().SetViewportSelectionToBottom()
			case nomad.ExecPage:
				m.getCurrentPageModel().SetInputPrefix("Enter command: ")
			}
			cmds = append(cmds, nomad.UpdatePageDataWithDelay(m.currentPage, m.updateInterval))
		}

	case nomad.UpdatePageDataMsg:
		if msg.Page == m.currentPage {
			cmds = append(cmds, m.getCurrentPageCmd())
		}

	case message.PageInputReceivedMsg:
		if m.currentPage == nomad.ExecPage {
			m.getCurrentPageModel().SetLoading(true)
			return m, nomad.InitiateWebSocket(m.nomadUrl, m.nomadToken, m.allocID, m.taskName, msg.Input)
		}

	case nomad.ExecWebSocketConnectedMsg:
		m.execWebSocket = msg.WebSocketConnection
		m.webSocketConnected = true
		m.getCurrentPageModel().SetLoading(false)
		m.setInPty(true)
		cmds = append(cmds, nomad.ResizeTty(m.execWebSocket, m.width, m.getCurrentPageModel().ViewportHeight()))
		cmds = append(cmds, nomad.ReadExecWebSocketNextMessage(m.execWebSocket))
		cmds = append(cmds, nomad.SendHeartbeatWithDelay())

	case nomad.ExecWebSocketHeartbeatMsg:
		if m.currentPage == nomad.ExecPage && m.webSocketConnected {
			cmds = append(cmds, nomad.SendHeartbeat(m.execWebSocket))
			cmds = append(cmds, nomad.SendHeartbeatWithDelay())
			return m, tea.Batch(cmds...)
		}
		return m, nil

	case nomad.ExecWebSocketResponseMsg:
		if m.currentPage == nomad.ExecPage {
			if msg.Close {
				m.webSocketConnected = false
				m.setInPty(false)
				m.getCurrentPageModel().AppendToViewport([]page.Row{{Row: constants.ExecWebSocketClosed}}, true)
			} else {
				m.appendToViewport(msg.StdOut, m.lastCommandFinished.stdOut)
				m.appendToViewport(msg.StdErr, m.lastCommandFinished.stdErr)
				m.updateLastCommandFinished(msg.StdOut, msg.StdErr)
				cmds = append(cmds, nomad.ReadExecWebSocketNextMessage(m.execWebSocket))
			}
		}
	}

	currentPageModel = m.getCurrentPageModel()
	if currentPageModel != nil && !currentPageModel.EnteringInput() {
		*currentPageModel, cmd = currentPageModel.Update(msg)
		cmds = append(cmds, cmd)
	}
	m.header.KeyHelp = nomad.GetPageKeyHelp(m.currentPage, m.currentPageFilterFocused(), m.currentPageFilterApplied(), m.currentPageViewportSaving(), m.getCurrentPageModel().EnteringInput(), m.inPty, m.webSocketConnected)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err) + "\n\nif this seems wrong, consider opening an issue here: https://github.com/robinovitch61/wander/issues/new/choose" + "\n\nq/ctrl+c to quit"
	} else if !m.initialized {
		return ""
	}

	pageView := m.header.View() + "\n" + m.getCurrentPageModel().View()

	return pageView
}

func (m *Model) initialize() {
	pageHeight := m.getPageHeight()

	m.pageModels = make(map[nomad.Page]*page.Model)

	jobsPage := page.New(m.width, pageHeight, m.getFilterPrefix(nomad.JobsPage), nomad.JobsPage.LoadingString(), true, false, false, constants.JobsViewportConditionalStyle)
	m.pageModels[nomad.JobsPage] = &jobsPage

	jobSpecPage := page.New(m.width, pageHeight, m.getFilterPrefix(nomad.JobSpecPage), nomad.JobSpecPage.LoadingString(), false, true, false, nil)
	m.pageModels[nomad.JobSpecPage] = &jobSpecPage

	allocationsPage := page.New(m.width, pageHeight, m.getFilterPrefix(nomad.AllocationsPage), nomad.AllocationsPage.LoadingString(), true, false, false, constants.AllocationsViewportConditionalStyle)
	m.pageModels[nomad.AllocationsPage] = &allocationsPage

	execPage := page.New(m.width, pageHeight, m.getFilterPrefix(nomad.ExecPage), nomad.ExecPage.LoadingString(), false, true, true, nil)
	m.pageModels[nomad.ExecPage] = &execPage

	allocSpecPage := page.New(m.width, pageHeight, m.getFilterPrefix(nomad.AllocSpecPage), nomad.AllocSpecPage.LoadingString(), false, true, false, nil)
	m.pageModels[nomad.AllocSpecPage] = &allocSpecPage

	logsPage := page.New(m.width, pageHeight, m.getFilterPrefix(nomad.LogsPage), nomad.LogsPage.LoadingString(), true, false, false, nil)
	m.pageModels[nomad.LogsPage] = &logsPage

	loglinePage := page.New(m.width, pageHeight, m.getFilterPrefix(nomad.LoglinePage), nomad.LoglinePage.LoadingString(), false, true, false, nil)
	m.pageModels[nomad.LoglinePage] = &loglinePage

	m.initialized = true
}

func (m *Model) setPageWindowSize() {
	for _, pm := range m.pageModels {
		pm.SetWindowSize(m.width, m.getPageHeight())
	}
}

func (m *Model) setPage(page nomad.Page) {
	m.getCurrentPageModel().HideToast()
	m.currentPage = page
	m.getCurrentPageModel().SetFilterPrefix(m.getFilterPrefix(page))
	if page.DoesLoad() {
		m.getCurrentPageModel().SetLoading(true)
	} else {
		m.getCurrentPageModel().SetLoading(false)
	}
}

func (m *Model) getCurrentPageModel() *page.Model {
	return m.pageModels[m.currentPage]
}

func (m *Model) getCurrentPageCmd() tea.Cmd {
	switch m.currentPage {
	case nomad.JobsPage:
		return nomad.FetchJobs(m.nomadUrl, m.nomadToken)
	case nomad.JobSpecPage:
		return nomad.FetchJobSpec(m.nomadUrl, m.nomadToken, m.jobID, m.jobNamespace)
	case nomad.AllocationsPage:
		return nomad.FetchAllocations(m.nomadUrl, m.nomadToken, m.jobID, m.jobNamespace)
	case nomad.ExecPage:
		return func() tea.Msg {
			return nomad.PageLoadedMsg{Page: nomad.ExecPage, TableHeader: []string{}, AllPageData: []page.Row{}}
		}
	case nomad.AllocSpecPage:
		return nomad.FetchAllocSpec(m.nomadUrl, m.nomadToken, m.allocID)
	case nomad.LogsPage:
		return nomad.FetchLogs(m.nomadUrl, m.nomadToken, m.allocID, m.taskName, m.logType)
	case nomad.LoglinePage:
		return nomad.FetchLogLine(m.logline)
	default:
		panic("page load command not found")
	}
}

func (m *Model) appendToViewport(content string, startOnNewLine bool) {
	stringRows := strings.Split(content, "\n")
	var pageRows []page.Row
	for _, row := range stringRows {
		stripped := formatter.StripANSI(row)
		pageRows = append(pageRows, page.Row{Row: stripped})
	}
	m.getCurrentPageModel().AppendToViewport(pageRows, startOnNewLine)
}

// updateLastCommandFinished updates lastCommandFinished, which is necessary
// because some data gets received in chunks in which a trailing \n indicates
// finished content, otherwise more content is expected (e.g. the exec
// websocket behaves this way when returning long content)
func (m *Model) updateLastCommandFinished(stdOut, stdErr string) {
	m.lastCommandFinished.stdOut = false
	m.lastCommandFinished.stdErr = false
	if strings.HasSuffix(stdOut, "\n") {
		m.lastCommandFinished.stdOut = true
	}
	if strings.HasSuffix(stdErr, "\n") {
		m.lastCommandFinished.stdErr = true
	}
}

func (m *Model) setInPty(inPty bool) {
	m.inPty = inPty
	m.getCurrentPageModel().SetViewportPromptVisible(inPty)
	if inPty {
		m.getCurrentPageModel().ScrollViewportToBottom()
	}
}

func (m Model) getPageHeight() int {
	return m.height - m.header.ViewHeight()
}

func (m Model) currentPageLoading() bool {
	return m.getCurrentPageModel().Loading()
}

func (m Model) currentPageFilterFocused() bool {
	return m.getCurrentPageModel().FilterFocused()
}

func (m Model) currentPageFilterApplied() bool {
	return m.getCurrentPageModel().FilterApplied()
}

func (m Model) currentPageViewportSaving() bool {
	return m.getCurrentPageModel().ViewportSaving()
}

func (m Model) getFilterPrefix(page nomad.Page) string {
	return page.GetFilterPrefix(m.jobID, m.taskName, m.allocID)
}

func getVersionString(v, s string) string {
	if v == "" {
		return "built from source"
	}
	return fmt.Sprintf("%s (%s)", v, s[:7])
}
