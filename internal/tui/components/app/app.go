package app

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hashicorp/nomad/api"
	"github.com/itchyny/gojq"
	"github.com/robinovitch61/wander/internal/dev"
	"github.com/robinovitch61/wander/internal/tui/components/toast"
	"github.com/robinovitch61/wander/internal/tui/components/header"
	"github.com/robinovitch61/wander/internal/tui/components/page"
	"github.com/robinovitch61/wander/internal/tui/constants"
	"github.com/robinovitch61/wander/internal/tui/formatter"
	"github.com/robinovitch61/wander/internal/tui/keymap"
	"github.com/robinovitch61/wander/internal/tui/message"
	"github.com/robinovitch61/wander/internal/tui/nomad"
	"github.com/robinovitch61/wander/internal/tui/style"
)

type TLSConfig struct {
	CACert, CAPath, ClientCert, ClientKey, ServerName string
	SkipVerify                                        bool
}

type EventConfig struct {
	Topics       nomad.Topics
	Namespace    string
	JQQuery      *gojq.Code
	AllocJQQuery *gojq.Code
}

type LogConfig struct {
	Offset int
	Tail   bool
}

type Config struct {
	RootOpts                      []string
	Version                       string
	URL, Token, Region, Namespace string
	HTTPAuth                      string
	TLS                           TLSConfig
	Event                         EventConfig
	Log                           LogConfig
	CopySavePath                  bool
	UpdateSeconds                 time.Duration
	JobColumns                    []string
	AllTaskColumns                []string
	JobTaskColumns                []string
	LogoColor                     string
	StartCompact                  bool
	StartAllTasksView             bool
	CompactTables                 bool
	StartFiltering                bool
	FilterWithContext             bool
}

type Model struct {
	config    Config
	client    api.Client

	header       header.Model
	compact      bool
	currentPage  nomad.Page
	pageModels   map[nomad.Page]*page.Model

	inJobsMode   bool
	jobID        string
	jobNamespace string
	alloc        api.Allocation
	taskName     string
	logline      string
	logType      nomad.LogType

	updateID int

	lastExecContent string

	eventsStream nomad.EventsStream
	event        string
	meta         map[string]string

	logsStream      nomad.LogsStream
	lastLogFinished bool

	// adminAction is a key of TaskAdminActions (or JobAdminActions, when it exists)
	adminAction nomad.AdminAction

	width, height int
	initialized   bool
	err           error
}

func getFirstPage(c Config) nomad.Page {
	firstPage := nomad.JobsPage
	if c.StartAllTasksView {
		firstPage = nomad.AllTasksPage
	}
	return firstPage
}

func InitialModel(c Config) Model {
	firstPage := getFirstPage(c)
	initialHeader := header.New(
		constants.LogoString,
		c.LogoColor,
		c.URL,
		c.Version,
		nomad.GetPageKeyHelp(firstPage, false, false, false, nomad.StdOut, false, !c.StartAllTasksView),
	)
	return Model{
		config:      c,
		header:      initialHeader,
		currentPage: firstPage,
		updateID:    nextUpdateID(),
		inJobsMode:  !c.StartAllTasksView,
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
		cmd = m.handleKeyMsg(msg)
		if cmd != nil {
			return m, cmd
		}

	case message.ErrMsg:
		m.err = msg
		return m, nil

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if !m.initialized {
			err := m.initialize()
			if err != nil {
				m.err = err
				return m, nil
			}
			cmds = append(cmds, m.getCurrentPageCmd())
		} else {
			m.setPageWindowSize()
		}

	case nomad.PageLoadedMsg:
		if msg.Page == m.currentPage {
			m.getCurrentPageModel().SetHeader(msg.TableHeader)
			m.getCurrentPageModel().SetAllPageRows(msg.AllPageRows)
			if m.currentPageLoading() {
				m.getCurrentPageModel().SetViewportXOffset(0)
			}
			if m.getCurrentPageModel().FilterWithContext {
				m.getCurrentPageModel().ResetContextFilter()
			}
			m.getCurrentPageModel().SetLoading(false)

			if m.currentPage.CanBeFirstPage() && len(msg.AllPageRows) == 0 {
				// oddly, nomad http api errors when one provides the wrong token,
				// but returns empty results when one provides an empty token
				m.getCurrentPageModel().SetHeader([]string{"Error"})
				m.getCurrentPageModel().SetAllPageRows([]page.Row{
					{Key: "", Row: "No results. Is the cluster empty or was no nomad token provided?"},
					{Key: "", Row: "Press q or ctrl+c to quit."},
				})
				m.getCurrentPageModel().SetViewportSelectionEnabled(false)
			}

			switch m.currentPage {
			case nomad.JobEventsPage, nomad.AllEventsPage:
				m.eventsStream = msg.EventsStream
				cmds = append(cmds, nomad.ReadEventsStreamNextMessage(m.eventsStream, m.config.Event.JQQuery))
			case nomad.AllocEventsPage:
				m.eventsStream = msg.EventsStream
				cmds = append(cmds, nomad.ReadEventsStreamNextMessage(m.eventsStream, m.config.Event.AllocJQQuery))
			case nomad.LogsPage:
				m.getCurrentPageModel().SetViewportSelectionToBottom()
				if m.config.Log.Tail {
					m.logsStream = msg.LogsStream
					m.lastLogFinished = true
					cmds = append(cmds, nomad.ReadLogsStreamNextMessage(m.logsStream))
				}
			case nomad.ExecPage:
				m.getCurrentPageModel().SetInputPrefix("Enter command: ")
			}
			cmds = append(cmds, nomad.UpdatePageDataWithDelay(m.updateID, m.currentPage, m.config.UpdateSeconds))
		}

	case nomad.EventsStreamMsg:
		if m.currentPage == nomad.JobEventsPage || m.currentPage == nomad.AllocEventsPage || m.currentPage == nomad.AllEventsPage {
			if fmt.Sprint(msg.Topics) == fmt.Sprint(m.eventsStream.Topics) {
				// sticky scroll down, i.e. if at bottom already, keep scrolling to bottom as new ones are added
				scrollDown := m.getCurrentPageModel().ViewportSelectionAtBottom()
				for _, event := range msg.Events {
					if event.CompleteValue == "{}" {
						continue
					}
					m.getCurrentPageModel().AppendToViewport([]page.Row{{Key: event.CompleteValue, Row: event.JQValue}}, true)
				}
				if scrollDown {
					m.getCurrentPageModel().ScrollViewportToBottom()
				}
			}
			query := m.config.Event.JQQuery
			if m.currentPage == nomad.AllocEventsPage {
				query = m.config.Event.AllocJQQuery
			}
			cmds = append(cmds, nomad.ReadEventsStreamNextMessage(m.eventsStream, query))
		}

	case nomad.LogsStreamMsg:
		if m.currentPage == nomad.LogsPage && m.logType == msg.Type {
			logLines := strings.Split(msg.Value, "\n")

			// finish with the last log line if necessary
			if !m.lastLogFinished {
				m.getCurrentPageModel().AppendToViewport([]page.Row{{Row: logLines[0]}}, false)
				logLines = logLines[1:]
			}

			// append all the new log rows in this chunk to the viewport at once
			scrollDown := m.getCurrentPageModel().ViewportSelectionAtBottom()
			var allRows []page.Row
			for _, logLine := range logLines {
				allRows = append(allRows, page.Row{Row: logLine})
			}
			m.getCurrentPageModel().AppendToViewport(allRows, true)
			if scrollDown {
				m.getCurrentPageModel().ScrollViewportToBottom()
			}

			m.lastLogFinished = strings.HasSuffix(msg.Value, "\n")
			cmds = append(cmds, nomad.ReadLogsStreamNextMessage(m.logsStream))
		}

	case nomad.UpdatePageDataMsg:
		if msg.ID == m.updateID && msg.Page == m.currentPage {
			cmds = append(cmds, m.getCurrentPageCmd())
			m.updateID = nextUpdateID()
		}

	case nomad.ExecCompleteMsg:
		if m.currentPage == nomad.ExecPage {
			m.getCurrentPageModel().SetDoesNeedNewInput()
			m.lastExecContent = strings.TrimSpace(msg.Output)
			m.setPage(nomad.ExecCompletePage)
			cmds = append(cmds, m.getCurrentPageCmd())
		}

	case message.PageInputReceivedMsg:
		if m.currentPage == nomad.ExecPage {
			// run the same wander executable even if there is a different one in the path
			ex, err := os.Executable()
			if err != nil {
				m.err = err
				return m, nil
			}
			dir := path.Dir(ex)

			args := []string{"exec"}
			// pass the same cli opts to wander exec as passed into the current wander root command
			args = append(args, m.config.RootOpts...)
			args = append(args, []string{
				m.alloc.ID,
				"--task",
				m.taskName,
				msg.Input,
			}...)
			c := exec.Command(fmt.Sprintf("%s/wander", dir), args...)
			c.Env = os.Environ()

			stdoutProxy := &nomad.StdoutProxy{}
			c.Stdout = stdoutProxy
			m.getCurrentPageModel().SetDoesNeedNewInput()
			return m, tea.ExecProcess(c, func(err error) tea.Msg {
				return nomad.ExecCompleteMsg{Output: string(stdoutProxy.SavedOutput)}
			})
		}

	case nomad.TaskAdminActionCompleteMsg:
		m.getCurrentPageModel().SetToast(
			toast.New(
				fmt.Sprintf(
					"%s completed successfully",
					nomad.GetTaskAdminText(
						m.adminAction, msg.TaskName, msg.AllocName, msg.AllocID))),
			style.SuccessToast,
		)

	case nomad.TaskAdminActionFailedMsg:
		m.getCurrentPageModel().SetToast(
			toast.New(
				fmt.Sprintf(
					"%s failed with error: %s",
					nomad.GetTaskAdminText(
						m.adminAction, msg.TaskName, msg.AllocName, msg.AllocID),
					msg.Error())),
			style.ErrorToast,
		)

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

func (m *Model) initialize() error {
	client, err := m.config.Client()
	if err != nil {
		return err
	}
	m.client = *client

	firstPage := getFirstPage(m.config)

	m.pageModels = make(map[nomad.Page]*page.Model)
	for k, pageConfig := range nomad.GetAllPageConfigs(m.width, m.getPageHeight(), m.config.CompactTables) {
		startFiltering := m.config.StartFiltering && k == firstPage
		p := page.New(pageConfig, m.config.CopySavePath, startFiltering, m.config.FilterWithContext)
		m.pageModels[k] = &p
	}

	if m.config.StartCompact {
		m.toggleCompact()
	}

	m.getCurrentPageModel().SetFilterPrefix(m.getFilterPrefix(m.currentPage))

	m.initialized = true
	return nil
}

func (m *Model) cleanupCmd() tea.Cmd {
	return func() tea.Msg {
		return message.CleanupCompleteMsg{}
	}
}

func (m *Model) setPageWindowSize() {
	for _, pm := range m.pageModels {
		pm.SetWindowSize(m.width, m.getPageHeight())
	}
}

func (m *Model) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	var cmds []tea.Cmd
	currentPageModel := m.getCurrentPageModel()

	// always exit if desired, or don't respond if typing "q" legitimately in some text input
	if key.Matches(msg, keymap.KeyMap.Exit) {
		addingQToFilter := m.currentPageFilterFocused()
		saving := m.currentPageViewportSaving()
		enteringInput := currentPageModel != nil && currentPageModel.EnteringInput()
		typingQLegitimately := msg.String() == "q" && (addingQToFilter || saving || enteringInput)
		if !typingQLegitimately || m.err != nil {
			return m.cleanupCmd()
		}
	}

	if !m.currentPageFilterFocused() && !m.currentPageViewportSaving() {
		switch {

		case key.Matches(msg, keymap.KeyMap.Compact):
			m.toggleCompact()
			return nil

		case key.Matches(msg, keymap.KeyMap.Forward):
			if selectedPageRow, err := m.getCurrentPageModel().GetSelectedPageRow(); err == nil {
				switch m.currentPage {
				case nomad.JobsPage:
					m.jobID, m.jobNamespace = nomad.JobIDAndNamespaceFromKey(selectedPageRow.Key)
				case nomad.JobEventsPage, nomad.AllocEventsPage, nomad.AllEventsPage:
					m.event = selectedPageRow.Key
				case nomad.LogsPage:
					m.logline = selectedPageRow.Row
				case nomad.TaskAdminPage:
					m.adminAction = nomad.KeyToAdminAction(selectedPageRow.Key)
				case nomad.TaskAdminConfirmPage:
					if selectedPageRow.Key == constants.ConfirmationKey {
						cmds = append(cmds, nomad.GetCmdForTaskAdminAction(
							m.client, m.adminAction, m.taskName, m.alloc.Name, m.alloc.ID))
					} else {
						backPage := m.currentPage.Backward(m.inJobsMode)
						m.setPage(backPage)
						cmds = append(cmds, m.getCurrentPageCmd())
						return tea.Batch(cmds...)
					}
				default:
					if m.currentPage.ShowsTasks() {
						taskInfo, err := nomad.TaskInfoFromKey(selectedPageRow.Key)
						if err != nil {
							m.err = err
							return nil
						}
						m.alloc, m.taskName = taskInfo.Alloc, taskInfo.TaskName
					}
				}

				nextPage := m.currentPage.Forward(m.inJobsMode)
				if nextPage != m.currentPage {
					m.setPage(nextPage)
					cmds = append(cmds, m.getCurrentPageCmd())
					return tea.Batch(cmds...)
				}
			}

		case key.Matches(msg, keymap.KeyMap.Back):
			if !m.currentPageFilterApplied() {
				switch m.currentPage {
				case nomad.ExecPage:
					m.getCurrentPageModel().SetDoesNeedNewInput()
				}

				backPage := m.currentPage.Backward(m.inJobsMode)
				if backPage != m.currentPage {
					m.setPage(backPage)
					cmds = append(cmds, m.getCurrentPageCmd())
					return tea.Batch(cmds...)
				}
			}

		case key.Matches(msg, keymap.KeyMap.Reload):
			if m.currentPage.DoesReload() {
				m.getCurrentPageModel().SetLoading(true)
				return m.getCurrentPageCmd()
			}

		}

		if key.Matches(msg, keymap.KeyMap.Exec) {
			if selectedPageRow, err := m.getCurrentPageModel().GetSelectedPageRow(); err == nil {
				if m.currentPage.ShowsTasks() {
					taskInfo, err := nomad.TaskInfoFromKey(selectedPageRow.Key)
					if err != nil {
						m.err = err
						return nil
					}
					if taskInfo.Running {
						m.alloc, m.taskName = taskInfo.Alloc, taskInfo.TaskName
						m.setPage(nomad.ExecPage)
						return m.getCurrentPageCmd()
					}
				}
			}
		}

		if key.Matches(msg, keymap.KeyMap.Stats) {
			if selectedPageRow, err := m.getCurrentPageModel().GetSelectedPageRow(); err == nil {
				if m.currentPage.ShowsTasks() {
					taskInfo, err := nomad.TaskInfoFromKey(selectedPageRow.Key)
					if err != nil {
						m.err = err
						return nil
					}
					if taskInfo.Running {
						m.alloc, m.taskName = taskInfo.Alloc, taskInfo.TaskName
						m.setPage(nomad.StatsPage)
						return m.getCurrentPageCmd()
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
					return m.getCurrentPageCmd()
				default:
					if m.currentPage.ShowsTasks() {
						taskInfo, err := nomad.TaskInfoFromKey(selectedPageRow.Key)
						if err != nil {
							m.err = err
							return nil
						}
						m.alloc, m.taskName = taskInfo.Alloc, taskInfo.TaskName
						m.setPage(nomad.AllocSpecPage)
						return m.getCurrentPageCmd()
					}
				}
			}
		}

		if key.Matches(msg, keymap.KeyMap.JobEvents) && m.currentPage == nomad.JobsPage {
			if selectedPageRow, err := m.getCurrentPageModel().GetSelectedPageRow(); err == nil {
				m.jobID, m.jobNamespace = nomad.JobIDAndNamespaceFromKey(selectedPageRow.Key)
				m.setPage(nomad.JobEventsPage)
				return m.getCurrentPageCmd()
			}
		}

		if key.Matches(msg, keymap.KeyMap.TasksMode) && m.currentPage == nomad.JobsPage {
			if m.inJobsMode {
				m.setPage(nomad.AllTasksPage)
				m.inJobsMode = false
				return m.getCurrentPageCmd()
			}
		}

		if key.Matches(msg, keymap.KeyMap.JobsMode) && m.currentPage == nomad.AllTasksPage {
			if !m.inJobsMode {
				m.setPage(nomad.JobsPage)
				m.inJobsMode = true
				return m.getCurrentPageCmd()
			}
		}

		if key.Matches(msg, keymap.KeyMap.JobMeta) && m.currentPage == nomad.JobsPage {
			if selectedPageRow, err := m.getCurrentPageModel().GetSelectedPageRow(); err == nil {
				m.jobID, m.jobNamespace = nomad.JobIDAndNamespaceFromKey(selectedPageRow.Key)
				m.setPage(nomad.JobMetaPage)
				return m.getCurrentPageCmd()
			}
		}

		if key.Matches(msg, keymap.KeyMap.AllocEvents) && m.currentPage.ShowsTasks() {
			if selectedPageRow, err := m.getCurrentPageModel().GetSelectedPageRow(); err == nil {
				taskInfo, err := nomad.TaskInfoFromKey(selectedPageRow.Key)
				if err != nil {
					m.err = err
					return nil
				}
				m.alloc, m.taskName = taskInfo.Alloc, taskInfo.TaskName
				m.setPage(nomad.AllocEventsPage)
				return m.getCurrentPageCmd()
			}
		}

		if key.Matches(msg, keymap.KeyMap.AllEvents) && m.currentPage == nomad.JobsPage {
			m.setPage(nomad.AllEventsPage)
			return m.getCurrentPageCmd()
		}

		if key.Matches(msg, keymap.KeyMap.AdminMenu) && m.currentPage.HasAdminMenu() {
			if selectedPageRow, err := m.getCurrentPageModel().GetSelectedPageRow(); err == nil {
				// Get task info from the currently selected row
				taskInfo, err := nomad.TaskInfoFromKey(selectedPageRow.Key)
				if err != nil {
					m.err = err
					return nil
				}
				if taskInfo.Running {
					m.alloc, m.taskName = taskInfo.Alloc, taskInfo.TaskName
					m.setPage(nomad.TaskAdminPage)
					return m.getCurrentPageCmd()
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
					return m.getCurrentPageCmd()
				}

			case key.Matches(msg, keymap.KeyMap.StdErr):
				if !m.currentPageLoading() && m.logType != nomad.StdErr {
					m.logType = nomad.StdErr
					stdErrHeaderStyle := style.ViewportHeaderStyle.Copy().Inherit(style.StdErr)
					m.getCurrentPageModel().SetViewportStyle(stdErrHeaderStyle, style.StdErr)
					m.getCurrentPageModel().SetLoading(true)
					return m.getCurrentPageCmd()
				}
			}
		}
	}

	return nil
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

func (m *Model) appendToViewport(content string, startOnNewLine bool) {
	stringRows := strings.Split(content, "\n")
	var pageRows []page.Row
	for _, row := range stringRows {
		stripOS := formatter.StripOSCommandSequences(row)
		stripped := formatter.StripANSI(stripOS)
		// bell seems to mess with parent terminal
		if stripped != "\a" {
			pageRows = append(pageRows, page.Row{Row: stripped})
		}
	}
	m.getCurrentPageModel().AppendToViewport(pageRows, startOnNewLine)
	m.getCurrentPageModel().ScrollViewportToBottom()
}

func (m *Model) updateKeyHelp() {
	newKeyHelp := nomad.GetPageKeyHelp(m.currentPage, m.currentPageFilterFocused(), m.currentPageFilterApplied(), m.currentPageViewportSaving(), m.logType, m.compact, m.inJobsMode)
	m.header.SetKeyHelp(newKeyHelp)
}

func (m *Model) toggleCompact() {
	m.compact = !m.compact
	m.header.ToggleCompact()
	m.updateKeyHelp()
	for _, pm := range m.pageModels {
		pm.ToggleCompact()
	}
	m.setPageWindowSize()
}

func (m Model) getCurrentPageCmd() tea.Cmd {
	switch m.currentPage {
	case nomad.JobsPage:
		return nomad.FetchJobs(m.client, m.config.JobColumns)
	case nomad.AllTasksPage:
		return nomad.FetchAllTasks(m.client, m.config.AllTaskColumns)
	case nomad.JobSpecPage:
		return nomad.FetchJobSpec(m.client, m.jobID, m.jobNamespace)
	case nomad.JobEventsPage:
		return nomad.FetchEventsStream(m.client, nomad.TopicsForJob(m.config.Event.Topics, m.jobID), m.jobNamespace, nomad.JobEventsPage)
	case nomad.JobEventPage:
		return nomad.PrettifyLine(m.event, nomad.JobEventPage)
	case nomad.JobMetaPage:
		return nomad.FetchJobMeta(m.client, m.jobID, m.jobNamespace)
	case nomad.AllocEventsPage:
		return nomad.FetchEventsStream(m.client, nomad.TopicsForAlloc(m.config.Event.Topics, m.alloc.ID), m.jobNamespace, nomad.AllocEventsPage)
	case nomad.AllocEventPage:
		return nomad.PrettifyLine(m.event, nomad.AllocEventPage)
	case nomad.AllEventsPage:
		return nomad.FetchEventsStream(m.client, m.config.Event.Topics, m.config.Event.Namespace, nomad.AllEventsPage)
	case nomad.AllEventPage:
		return nomad.PrettifyLine(m.event, nomad.AllEventPage)
	case nomad.JobTasksPage:
		return nomad.FetchTasksForJob(m.client, m.jobID, m.jobNamespace, m.config.JobTaskColumns)
	case nomad.ExecPage:
		return func() tea.Msg {
			// this does no async work, just moves to request the command input
			return nomad.PageLoadedMsg{Page: nomad.ExecPage, TableHeader: []string{}, AllPageRows: []page.Row{}}
		}
	case nomad.ExecCompletePage:
		return func() tea.Msg {
			// this does no async work, just shows the output of the prior exec session
			var allPageRows []page.Row
			for _, row := range strings.Split(m.lastExecContent, "\n") {
				row = strings.ReplaceAll(row, "\r", "")
				if len(row) == 0 {
					continue
				}
				allPageRows = append(allPageRows, page.Row{Row: formatter.StripOSCommandSequences(formatter.StripANSI(row))})
			}
			return nomad.PageLoadedMsg{Page: nomad.ExecCompletePage, TableHeader: []string{"Exec Session Output"}, AllPageRows: allPageRows}
		}
	case nomad.AllocSpecPage:
		return nomad.FetchAllocSpec(m.client, m.alloc.ID)
	case nomad.LogsPage:
		return nomad.FetchLogs(m.client, m.alloc, m.taskName, m.logType, m.config.Log.Offset, m.config.Log.Tail)
	case nomad.LoglinePage:
		return nomad.PrettifyLine(m.logline, nomad.LoglinePage)
	case nomad.StatsPage:
		return nomad.FetchStats(m.client, m.alloc.ID, m.alloc.Name)
	case nomad.TaskAdminPage:
		return func() tea.Msg {
			// this does no async work, just constructs the task admin menu
			var rows []page.Row
			for action := range nomad.TaskAdminActions {
				rows = append(rows, page.Row{
					Key: nomad.AdminActionToKey(action),
					Row: nomad.GetTaskAdminText(action, m.taskName, m.alloc.Name, m.alloc.ID),
				})
			}
			return nomad.PageLoadedMsg{
				Page:        nomad.TaskAdminPage,
				TableHeader: []string{"Available Admin Actions"},
				AllPageRows: rows,
			}
		}

	case nomad.TaskAdminConfirmPage:
		return func() tea.Msg {
			// this does no async work, just constructs the confirmation page
			confirmationText := nomad.GetTaskAdminText(m.adminAction, m.taskName, m.alloc.Name, m.alloc.ID)
			confirmationText = strings.ToLower(confirmationText[:1]) + confirmationText[1:]
			return nomad.PageLoadedMsg{
				Page:        nomad.TaskAdminConfirmPage,
				TableHeader: []string{"Are you sure?"},
				AllPageRows: []page.Row{
					{Key: "Cancel", Row: "Cancel"},
					{Key: constants.ConfirmationKey, Row: fmt.Sprintf("Yes, %s", confirmationText)},
				},
			}
		}

	default:
		panic(fmt.Sprintf("Load command for page:%s not found", m.currentPage))
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
	return page.GetFilterPrefix(m.config.Namespace, m.jobID, m.taskName, m.alloc.Name, m.alloc.ID, m.config.Event.Topics, m.config.Event.Namespace)
}
