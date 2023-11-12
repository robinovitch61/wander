package nomad

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hashicorp/nomad/api"
	"github.com/robinovitch61/wander/internal/tui/components/page"
	"github.com/robinovitch61/wander/internal/tui/components/viewport"
	"github.com/robinovitch61/wander/internal/tui/constants"
	"github.com/robinovitch61/wander/internal/tui/formatter"
	"github.com/robinovitch61/wander/internal/tui/keymap"
	"github.com/robinovitch61/wander/internal/tui/style"
	"strings"
	"time"
)

type Page int8

const (
	Unset Page = iota
	JobsPage
	AllTasksPage
	JobSpecPage
	JobEventsPage
	JobEventPage
	JobMetaPage
	AllocEventsPage
	AllocEventPage
	AllEventsPage
	AllEventPage
	JobTasksPage
	ExecPage
	AllocSpecPage
	LogsPage
	LoglinePage
)

func GetAllPageConfigs(width, height int, compactTables bool) map[Page]page.Config {
	return map[Page]page.Config{
		JobsPage: {
			Width: width, Height: height,
			LoadingString:    JobsPage.LoadingString(),
			SelectionEnabled: true, WrapText: false, RequestInput: false,
			CompactTableContent:      compactTables,
			ViewportConditionalStyle: constants.JobsTableStatusStyles,
		},
		AllTasksPage: {
			Width: width, Height: height,
			LoadingString:    AllTasksPage.LoadingString(),
			SelectionEnabled: true, WrapText: false, RequestInput: false,
			CompactTableContent:      compactTables,
			ViewportConditionalStyle: constants.TasksTableStatusStyles,
		},
		JobSpecPage: {
			Width: width, Height: height,
			LoadingString:    JobSpecPage.LoadingString(),
			SelectionEnabled: false, WrapText: true, RequestInput: false,
		},
		JobEventsPage: {
			Width: width, Height: height,
			LoadingString:    JobEventsPage.LoadingString(),
			SelectionEnabled: true, WrapText: false, RequestInput: false,
		},
		JobEventPage: {
			Width: width, Height: height,
			LoadingString:    JobEventPage.LoadingString(),
			SelectionEnabled: false, WrapText: true, RequestInput: false,
		},
		JobMetaPage: {
			Width: width, Height: height,
			LoadingString:    JobMetaPage.LoadingString(),
			SelectionEnabled: false, WrapText: true, RequestInput: false,
			CompactTableContent: compactTables,
		},
		AllocEventsPage: {
			Width: width, Height: height,
			LoadingString:    AllocEventsPage.LoadingString(),
			SelectionEnabled: true, WrapText: false, RequestInput: false,
		},
		AllocEventPage: {
			Width: width, Height: height,
			LoadingString:    AllocEventPage.LoadingString(),
			SelectionEnabled: false, WrapText: true, RequestInput: false,
		},
		AllEventsPage: {
			Width: width, Height: height,
			LoadingString:    AllEventsPage.LoadingString(),
			SelectionEnabled: true, WrapText: false, RequestInput: false,
		},
		AllEventPage: {
			Width: width, Height: height,
			LoadingString:    AllEventPage.LoadingString(),
			SelectionEnabled: false, WrapText: true, RequestInput: false,
		},
		JobTasksPage: {
			Width: width, Height: height,
			LoadingString:    JobTasksPage.LoadingString(),
			SelectionEnabled: true, WrapText: false, RequestInput: false,
			CompactTableContent:      compactTables,
			ViewportConditionalStyle: constants.TasksTableStatusStyles,
		},
		ExecPage: {
			Width: width, Height: height,
			LoadingString:    ExecPage.LoadingString(),
			SelectionEnabled: false, WrapText: true, RequestInput: true,
		},
		AllocSpecPage: {
			Width: width, Height: height,
			LoadingString:    AllocSpecPage.LoadingString(),
			SelectionEnabled: false, WrapText: true, RequestInput: false,
		},
		LogsPage: {
			Width: width, Height: height,
			LoadingString:    LogsPage.LoadingString(),
			SelectionEnabled: true, WrapText: false, RequestInput: false,
		},
		LoglinePage: {
			Width: width, Height: height,
			LoadingString:    LoglinePage.LoadingString(),
			SelectionEnabled: false, WrapText: true, RequestInput: false,
		},
	}
}

func (p Page) DoesLoad() bool {
	noLoadPages := []Page{LoglinePage, JobEventPage, AllocEventPage, AllEventPage}
	for _, noLoadPage := range noLoadPages {
		if noLoadPage == p {
			return false
		}
	}
	return true
}

func (p Page) DoesReload() bool {
	noReloadPages := []Page{LoglinePage, JobEventsPage, JobEventPage, AllocEventsPage, AllocEventPage, AllEventsPage, AllEventPage, ExecPage}
	for _, noReloadPage := range noReloadPages {
		if noReloadPage == p {
			return false
		}
	}
	return true
}

func (p Page) ShowsTasks() bool {
	taskPages := []Page{AllTasksPage, JobTasksPage}
	for _, taskPage := range taskPages {
		if taskPage == p {
			return true
		}
	}
	return false
}

func (p Page) CanBeStart() bool {
	return p == JobsPage || p == AllTasksPage
}

func (p Page) doesUpdate() bool {
	noUpdatePages := []Page{
		LoglinePage,     // doesn't load
		ExecPage,        // doesn't reload
		LogsPage,        // currently makes scrolling impossible - solve in https://github.com/robinovitch61/wander/issues/1
		JobSpecPage,     // would require changes to make scrolling possible
		AllocSpecPage,   // would require changes to make scrolling possible
		JobEventsPage,   // constant connection, streams data
		JobEventPage,    // doesn't load
		AllocEventsPage, // constant connection, streams data
		AllocEventPage,  // doesn't load
		AllEventsPage,   // constant connection, streams data
		AllEventPage,    // doesn't load
	}
	for _, noUpdatePage := range noUpdatePages {
		if noUpdatePage == p {
			return false
		}
	}
	return true
}

func (p Page) String() string {
	switch p {
	case Unset:
		return "undefined"
	case JobsPage:
		return "jobs"
	case AllTasksPage:
		return "all tasks"
	case JobSpecPage:
		return "job spec"
	case JobEventsPage:
		return "job events"
	case AllocEventsPage:
		return "events"
	case JobMetaPage:
		return "meta"
	case AllEventsPage:
		return "all events"
	case JobEventPage, AllocEventPage, AllEventPage:
		return "event"
	case JobTasksPage:
		return "tasks"
	case ExecPage:
		return "exec"
	case AllocSpecPage:
		return "allocation spec"
	case LogsPage:
		return "logs"
	case LoglinePage:
		return "log"
	}
	return "unknown"
}

func (p Page) LoadingString() string {
	return fmt.Sprintf("Loading %s...", p.String())
}

func (p Page) Forward() Page {
	switch p {
	case JobsPage:
		return JobTasksPage
	case AllTasksPage:
		return LogsPage
	case JobEventsPage:
		return JobEventPage
	case AllocEventsPage:
		return AllocEventPage
	case AllEventsPage:
		return AllEventPage
	case JobTasksPage:
		return LogsPage
	case LogsPage:
		return LoglinePage
	}
	return p
}

func returnToPage(inJobsMode bool) Page {
	if inJobsMode {
		return JobTasksPage
	}
	return AllTasksPage
}

func (p Page) Backward(inJobsMode bool) Page {
	switch p {
	case JobSpecPage:
		return JobsPage
	case JobEventsPage:
		return JobsPage
	case JobEventPage:
		return JobEventsPage
	case JobMetaPage:
		return JobsPage
	case AllocEventsPage:
		return returnToPage(inJobsMode)
	case AllocEventPage:
		return AllocEventsPage
	case AllEventsPage:
		return JobsPage
	case AllEventPage:
		return AllEventsPage
	case JobTasksPage:
		return JobsPage
	case ExecPage:
		return returnToPage(inJobsMode)
	case AllocSpecPage:
		return returnToPage(inJobsMode)
	case LogsPage:
		return returnToPage(inJobsMode)
	case LoglinePage:
		return LogsPage
	}
	return p
}

func allocEventFilterPrefix(allocName, allocID string) string {
	return fmt.Sprintf("%s %s", style.Bold.Render(allocName), formatter.ShortAllocID(allocID))
}

func taskFilterPrefix(taskName, allocName string) string {
	return fmt.Sprintf("%s in %s", style.Bold.Render(taskName), allocName)
}

func namespaceFilterPrefix(namespace string) string {
	if namespace == "*" {
		return "All Namespaces"
	}
	return fmt.Sprintf("Namespace %s", style.Bold.Render(namespace))
}

func (p Page) GetFilterPrefix(namespace, jobID, taskName, allocName, allocID string, eventTopics Topics, eventNamespace string) string {
	switch p {
	case JobsPage:
		return fmt.Sprintf("Jobs in %s", namespaceFilterPrefix(namespace))
	case AllTasksPage:
		return fmt.Sprintf("All Tasks in %s", namespaceFilterPrefix(namespace))
	case JobSpecPage:
		return fmt.Sprintf("Spec for Job %s", style.Bold.Render(jobID))
	case JobEventsPage:
		return fmt.Sprintf("Events for Job %s (%s)", style.Bold.Render(jobID), getTopicNames(eventTopics))
	case JobEventPage:
		return fmt.Sprintf("Event for Job %s", style.Bold.Render(jobID))
	case JobMetaPage:
		return fmt.Sprintf("Meta for Job %s", jobID)
	case AllocEventsPage:
		return fmt.Sprintf("Events for Allocation %s", allocEventFilterPrefix(allocName, allocID))
	case AllocEventPage:
		return fmt.Sprintf("Event for Allocation %s", allocEventFilterPrefix(allocName, allocID))
	case AllEventsPage:
		return fmt.Sprintf("All Events in Namespace %s (%s)", eventNamespace, formatEventTopics(eventTopics))
	case AllEventPage:
		return fmt.Sprintf("Event")
	case JobTasksPage:
		return fmt.Sprintf("Tasks for Job %s", style.Bold.Render(jobID))
	case ExecPage:
		return fmt.Sprintf("Exec for Task %s", taskFilterPrefix(taskName, allocName))
	case AllocSpecPage:
		return fmt.Sprintf("Spec for Allocation %s %s", style.Bold.Render(allocName), formatter.ShortAllocID(allocID))
	case LogsPage:
		return fmt.Sprintf("Logs for Task %s", taskFilterPrefix(taskName, allocName))
	case LoglinePage:
		return fmt.Sprintf("Log Line for Task %s", taskFilterPrefix(taskName, allocName))
	default:
		panic("page not found")
	}
}

type EventsStream struct {
	Chan      <-chan *api.Events
	Topics    Topics
	Namespace string
}

type LogsStream struct {
	Chan    <-chan *api.StreamFrame
	LogType LogType
}

type PageLoadedMsg struct {
	Page         Page
	TableHeader  []string
	AllPageRows  []page.Row
	EventsStream EventsStream
	LogsStream   LogsStream
}

type UpdatePageDataMsg struct {
	ID   int
	Page Page
}

func UpdatePageDataWithDelay(id int, p Page, d time.Duration) tea.Cmd {
	if p.doesUpdate() && d > 0 {
		return tea.Tick(d, func(t time.Time) tea.Msg { return UpdatePageDataMsg{id, p} })
	}
	return nil
}

func getShortHelp(bindings []key.Binding) string {
	var output string
	for _, km := range bindings {
		output += style.KeyHelpKey.Render(km.Help().Key) + " " + style.KeyHelpDescription.Render(km.Help().Desc) + "  "
	}
	output = strings.TrimSpace(output)
	return output
}

func changeKeyHelp(k *key.Binding, h string) {
	k.SetHelp(k.Help().Key, h)
}

func GetPageKeyHelp(
	currentPage Page,
	filterFocused, filterApplied, saving, enteringInput, inPty, webSocketConnected bool,
	logType LogType,
	compact, inJobsMode bool,
) string {
	if compact {
		changeKeyHelp(&keymap.KeyMap.Compact, "expand header")
		return getShortHelp([]key.Binding{keymap.KeyMap.Compact})
	} else {
		changeKeyHelp(&keymap.KeyMap.Compact, "compact")
	}

	if filterFocused || enteringInput {
		keymap.KeyMap.Exit.SetHelp("ctrl+c", "exit")
	} else {
		keymap.KeyMap.Exit.SetHelp("q/ctrl+c", "exit")
	}

	firstRow := []key.Binding{keymap.KeyMap.Exit}

	if !saving && !filterFocused {
		firstRow = append(firstRow, keymap.KeyMap.Compact)
		if currentPage.DoesReload() {
			firstRow = append(firstRow, keymap.KeyMap.Reload)
		}
	}

	viewportKeyMap := viewport.GetKeyMap()
	secondRow := []key.Binding{viewportKeyMap.Save, keymap.KeyMap.Wrap}
	thirdRow := []key.Binding{viewportKeyMap.Down, viewportKeyMap.Up, viewportKeyMap.PageDown, viewportKeyMap.PageUp, viewportKeyMap.Bottom, viewportKeyMap.Top}

	var fourthRow []key.Binding
	if nextPage := currentPage.Forward(); nextPage != currentPage {
		changeKeyHelp(&keymap.KeyMap.Forward, currentPage.Forward().String())
		fourthRow = append(fourthRow, keymap.KeyMap.Forward)
	}

	if filterApplied {
		changeKeyHelp(&keymap.KeyMap.Back, "remove filter")
		fourthRow = append(fourthRow, keymap.KeyMap.Back)
	} else if prevPage := currentPage.Backward(inJobsMode); prevPage != currentPage {
		changeKeyHelp(&keymap.KeyMap.Back, fmt.Sprintf("%s", currentPage.Backward(inJobsMode).String()))
		fourthRow = append(fourthRow, keymap.KeyMap.Back)
	}

	if currentPage == JobsPage || currentPage.ShowsTasks() {
		if currentPage == JobsPage {
			fourthRow = append(fourthRow, keymap.KeyMap.TasksMode)
		} else if currentPage == AllTasksPage {
			fourthRow = append(fourthRow, keymap.KeyMap.JobsMode)
		}
		fourthRow = append(fourthRow, keymap.KeyMap.Spec)
	} else if currentPage == LogsPage {
		if logType == StdOut {
			fourthRow = append(fourthRow, keymap.KeyMap.StdErr)
		} else {
			fourthRow = append(fourthRow, keymap.KeyMap.StdOut)
		}
	}

	if currentPage == JobsPage {
		fourthRow = append(fourthRow, keymap.KeyMap.JobEvents)
		fourthRow = append(fourthRow, keymap.KeyMap.AllEvents)
		fourthRow = append(fourthRow, keymap.KeyMap.JobMeta)
	}

	if currentPage.ShowsTasks() {
		fourthRow = append(fourthRow, keymap.KeyMap.AllocEvents)
		fourthRow = append(fourthRow, keymap.KeyMap.Exec)
	}

	if currentPage == ExecPage {
		if enteringInput {
			changeKeyHelp(&keymap.KeyMap.Forward, "run command")
			secondRow = append(fourthRow, keymap.KeyMap.Forward)
			return getShortHelp(firstRow) + "\n" + getShortHelp(secondRow)
		}
		if inPty {
			changeKeyHelp(&keymap.KeyMap.Back, "disable input")
			return getShortHelp([]key.Binding{keymap.KeyMap.Back})
		} else {
			if webSocketConnected {
				changeKeyHelp(&keymap.KeyMap.Forward, "enable input")
				fourthRow = append(fourthRow, keymap.KeyMap.Forward)
			}
		}
	}

	if saving {
		changeKeyHelp(&keymap.KeyMap.Forward, "confirm save")
		changeKeyHelp(&keymap.KeyMap.Back, "cancel save")
		secondRow = []key.Binding{keymap.KeyMap.Back, keymap.KeyMap.Forward}
		return getShortHelp(firstRow) + "\n" + getShortHelp(secondRow)
	}

	if filterFocused {
		changeKeyHelp(&keymap.KeyMap.Forward, "apply filter")
		changeKeyHelp(&keymap.KeyMap.Back, "cancel filter")
		secondRow = []key.Binding{keymap.KeyMap.Back, keymap.KeyMap.Forward}
		return getShortHelp(firstRow) + "\n" + getShortHelp(secondRow)
	}

	var final string
	for _, row := range [][]key.Binding{firstRow, secondRow, thirdRow, fourthRow} {
		final += getShortHelp(row) + "\n"
	}

	return strings.TrimRight(final, "\n")
}
