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

type Config struct {
	Version, SHA, URL, Token, EventTopics, EventNamespace string
	UpdateSeconds                                         time.Duration
}

type Model struct {
	config Config

	header      header.Model
	currentPage nomad.Page
	pageModels  map[nomad.Page]*page.Model

	jobID        string
	jobNamespace string
	allocID      string
	taskName     string
	logline      string
	logType      nomad.LogType

	updateID int

	eventsStream nomad.PersistentConnection
	event        string

	execWebSocket       *websocket.Conn
	execPty             *os.File
	inPty               bool
	webSocketConnected  bool
	lastCommandFinished struct{ stdOut, stdErr bool }

	width, height int
	initialized   bool
	err           error
}

func InitialModel(c Config) Model {
	firstPage := nomad.JobsPage
	initialHeader := header.New(constants.LogoString, c.URL, getVersionString(c.Version, c.SHA), nomad.GetPageKeyHelp(firstPage, false, false, false, false, false, false))

	return Model{
		config:      c,
		header:      initialHeader,
		currentPage: firstPage,
		updateID:    nextUpdateID(),
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
	case message.CleanupCompleteMsg:
		return m, tea.Quit

	case tea.KeyMsg:
		// always exit if desired, or don't respond if typing "q" legitimately in some text input
		if key.Matches(msg, keymap.KeyMap.Exit) {
			addingQToFilter := m.currentPageFilterFocused()
			saving := m.currentPageViewportSaving()
			enteringInput := currentPageModel != nil && currentPageModel.EnteringInput()
			typingQLegitimately := msg.String() == "q" && (addingQToFilter || saving || enteringInput || m.inPty)
			if !typingQLegitimately || m.err != nil {
				return m, m.cleanupCmd()
			}
		}

		if m.currentPage == nomad.ExecPage {
			var keypress string
			if m.inPty {
				if key.Matches(msg, keymap.KeyMap.Back) {
					m.setInPty(false)
					return m, nil
				} else {
					keypress = nomad.GetKeypress(msg)
					return m, nomad.SendWebSocketMessage(m.execWebSocket, keypress)
				}
			} else if key.Matches(msg, keymap.KeyMap.Forward) && m.webSocketConnected && !m.currentPageViewportSaving() {
				m.setInPty(true)
			}
		}

		if !m.currentPageFilterFocused() && !m.currentPageViewportSaving() {
			switch {
			case key.Matches(msg, keymap.KeyMap.Forward):
				if selectedPageRow, err := m.getCurrentPageModel().GetSelectedPageRow(); err == nil {
					switch m.currentPage {
					case nomad.JobsPage:
						m.jobID, m.jobNamespace = nomad.JobIDAndNamespaceFromKey(selectedPageRow.Key)
					case nomad.EventsPage:
						m.event = selectedPageRow.Row
					case nomad.AllocationsPage:
						allocInfo, err := nomad.AllocationInfoFromKey(selectedPageRow.Key)
						if err != nil {
							m.err = err
							return m, nil
						}
						m.allocID, m.taskName = allocInfo.AllocID, allocInfo.TaskName
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

			if key.Matches(msg, keymap.KeyMap.Events) && m.currentPage == nomad.JobsPage {
				m.setPage(nomad.EventsPage)
				return m, m.getCurrentPageCmd()
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
				viewportHeightWithoutFooter := m.getCurrentPageModel().ViewportHeight() - 1 // hardcoded as known today, has to change if footer expands
				cmds = append(cmds, nomad.ResizeTty(m.execWebSocket, m.width, viewportHeightWithoutFooter))
			}
		}

	case nomad.PageLoadedMsg:
		if msg.Page == m.currentPage {
			m.getCurrentPageModel().SetHeader(msg.TableHeader)
			m.getCurrentPageModel().SetAllPageData(msg.AllPageRows)
			if m.currentPageLoading() {
				m.getCurrentPageModel().SetViewportXOffset(0)
			}
			m.getCurrentPageModel().SetLoading(false)

			switch m.currentPage {
			case nomad.JobsPage:
				if m.currentPage == nomad.JobsPage && len(msg.AllPageRows) == 0 {
					// oddly, nomad http api errors when one provides the wrong token, but returns empty results when one provides an empty token
					m.getCurrentPageModel().SetAllPageData([]page.Row{
						{"", "No job results. Is the cluster empty or no nomad token provided?"},
						{"", "Press q or ctrl+c to quit."},
					})
					m.getCurrentPageModel().SetViewportSelectionEnabled(false)
				}
			case nomad.EventsPage:
				if m.eventsStream.Body != nil {
					err := m.eventsStream.Body.Close()
					if err != nil {
						m.err = err
						return m, nil
					}
				}
				m.eventsStream = msg.Connection
				cmds = append(cmds, nomad.ReadEventsStreamNextMessage(m.eventsStream.Reader))
			case nomad.LogsPage:
				m.getCurrentPageModel().SetViewportSelectionToBottom()
			case nomad.ExecPage:
				m.getCurrentPageModel().SetInputPrefix("Enter command: ")
			}
			cmds = append(cmds, nomad.UpdatePageDataWithDelay(m.updateID, m.currentPage, m.config.UpdateSeconds))
		}

	case nomad.EventsStreamMsg:
		if m.currentPage == nomad.EventsPage {
			if !msg.Closed {
				if msg.Value != "{}" {
					scrollDown := m.getCurrentPageModel().ViewportSelectionAtBottom()
					m.getCurrentPageModel().AppendToViewport([]page.Row{{Row: msg.Value}}, true)
					if scrollDown {
						m.getCurrentPageModel().ScrollViewportToBottom()
					}
				}
				cmds = append(cmds, nomad.ReadEventsStreamNextMessage(m.eventsStream.Reader))
			}
		}

	case nomad.UpdatePageDataMsg:
		if msg.ID == m.updateID && msg.Page == m.currentPage {
			cmds = append(cmds, m.getCurrentPageCmd())
			m.updateID = nextUpdateID()
		}

	case message.PageInputReceivedMsg:
		if m.currentPage == nomad.ExecPage {
			m.getCurrentPageModel().SetLoading(true)
			return m, nomad.InitiateWebSocket(m.config.URL, m.config.Token, m.allocID, m.taskName, msg.Input)
		}

	case nomad.ExecWebSocketConnectedMsg:
		m.execWebSocket = msg.WebSocketConnection
		m.webSocketConnected = true
		m.getCurrentPageModel().SetLoading(false)
		m.setInPty(true)
		viewportHeightWithoutFooter := m.getCurrentPageModel().ViewportHeight() - 1 // hardcoded as known today, has to change if footer expands
		cmds = append(cmds, nomad.ResizeTty(m.execWebSocket, m.width, viewportHeightWithoutFooter))
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
				m.getCurrentPageModel().ScrollViewportToBottom()
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
	m.updateKeyHelp()

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

	eventsPage := page.New(m.width, pageHeight, m.getFilterPrefix(nomad.EventsPage), nomad.EventsPage.LoadingString(), true, false, false, nil)
	m.pageModels[nomad.EventsPage] = &eventsPage

	eventPage := page.New(m.width, pageHeight, m.getFilterPrefix(nomad.EventPage), nomad.EventPage.LoadingString(), false, true, false, nil)
	m.pageModels[nomad.EventPage] = &eventPage

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

func (m *Model) cleanupCmd() tea.Cmd {
	return func() tea.Msg {
		if m.eventsStream.Body != nil {
			_ = m.eventsStream.Body.Close()
		}
		if m.execWebSocket != nil {
			nomad.CloseWebSocket(m.execWebSocket)()
		}
		return message.CleanupCompleteMsg{}
	}
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
		return nomad.FetchJobs(m.config.URL, m.config.Token)
	case nomad.JobSpecPage:
		return nomad.FetchJobSpec(m.config.URL, m.config.Token, m.jobID, m.jobNamespace)
	case nomad.EventsPage:
		return nomad.FetchEventsStream(m.config.URL, m.config.Token, m.config.EventTopics, m.config.EventNamespace)
	case nomad.EventPage:
		return nomad.PrettifyLine(m.event, nomad.EventPage)
	case nomad.AllocationsPage:
		return nomad.FetchAllocations(m.config.URL, m.config.Token, m.jobID, m.jobNamespace)
	case nomad.ExecPage:
		return nomad.LoadExecPage()
	case nomad.AllocSpecPage:
		return nomad.FetchAllocSpec(m.config.URL, m.config.Token, m.allocID)
	case nomad.LogsPage:
		return nomad.FetchLogs(m.config.URL, m.config.Token, m.allocID, m.taskName, m.logType)
	case nomad.LoglinePage:
		return nomad.PrettifyLine(m.logline, nomad.LoglinePage)
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
	m.getCurrentPageModel().ScrollViewportToBottom()
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
	m.updateKeyHelp()
}

func (m *Model) updateKeyHelp() {
	m.header.KeyHelp = nomad.GetPageKeyHelp(m.currentPage, m.currentPageFilterFocused(), m.currentPageFilterApplied(), m.currentPageViewportSaving(), m.getCurrentPageModel().EnteringInput(), m.inPty, m.webSocketConnected)
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
	return page.GetFilterPrefix(m.jobID, m.taskName, m.allocID, m.config.EventTopics, m.config.EventNamespace)
}

func getVersionString(v, s string) string {
	if v == "" {
		return "built from source"
	}
	return fmt.Sprintf("%s (%s)", v, s[:7])
}
