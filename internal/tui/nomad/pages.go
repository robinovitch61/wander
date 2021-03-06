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
	JobSpecPage
	JobEventsPage
	JobEventPage
	AllocEventsPage
	AllocEventPage
	AllEventsPage
	AllEventPage
	AllocationsPage
	ExecPage
	AllocSpecPage
	LogsPage
	LoglinePage
)

func GetAllPageConfigs(width, height int, copySavePath bool) map[Page]page.Config {
	return map[Page]page.Config{
		JobsPage: {
			Width: width, Height: height,
			FilterPrefix: "Jobs", LoadingString: JobsPage.LoadingString(),
			CopySavePath: copySavePath, SelectionEnabled: true, WrapText: false, RequestInput: false,
			ViewportConditionalStyle: constants.JobsViewportConditionalStyle,
		},
		JobSpecPage: {
			Width: width, Height: height,
			LoadingString: JobSpecPage.LoadingString(),
			CopySavePath:  copySavePath, SelectionEnabled: false, WrapText: true, RequestInput: false,
		},
		JobEventsPage: {
			Width: width, Height: height,
			LoadingString: JobEventsPage.LoadingString(),
			CopySavePath:  copySavePath, SelectionEnabled: true, WrapText: false, RequestInput: false,
		},
		JobEventPage: {
			Width: width, Height: height,
			LoadingString: JobEventPage.LoadingString(),
			CopySavePath:  copySavePath, SelectionEnabled: false, WrapText: true, RequestInput: false,
		},
		AllocEventsPage: {
			Width: width, Height: height,
			LoadingString: AllocEventsPage.LoadingString(),
			CopySavePath:  copySavePath, SelectionEnabled: true, WrapText: false, RequestInput: false,
		},
		AllocEventPage: {
			Width: width, Height: height,
			LoadingString: AllocEventPage.LoadingString(),
			CopySavePath:  copySavePath, SelectionEnabled: false, WrapText: true, RequestInput: false,
		},
		AllEventsPage: {
			Width: width, Height: height,
			LoadingString: AllEventsPage.LoadingString(),
			CopySavePath:  copySavePath, SelectionEnabled: true, WrapText: false, RequestInput: false,
		},
		AllEventPage: {
			Width: width, Height: height,
			LoadingString: AllEventPage.LoadingString(),
			CopySavePath:  copySavePath, SelectionEnabled: false, WrapText: true, RequestInput: false,
		},
		AllocationsPage: {
			Width: width, Height: height,
			LoadingString: AllocationsPage.LoadingString(),
			CopySavePath:  copySavePath, SelectionEnabled: true, WrapText: false, RequestInput: false,
			ViewportConditionalStyle: constants.AllocationsViewportConditionalStyle,
		},
		ExecPage: {
			Width: width, Height: height,
			LoadingString: ExecPage.LoadingString(),
			CopySavePath:  copySavePath, SelectionEnabled: false, WrapText: true, RequestInput: true,
		},
		AllocSpecPage: {
			Width: width, Height: height,
			LoadingString: AllocSpecPage.LoadingString(),
			CopySavePath:  copySavePath, SelectionEnabled: false, WrapText: true, RequestInput: false,
		},
		LogsPage: {
			Width: width, Height: height,
			LoadingString: LogsPage.LoadingString(),
			CopySavePath:  copySavePath, SelectionEnabled: true, WrapText: false, RequestInput: false,
		},
		LoglinePage: {
			Width: width, Height: height,
			LoadingString: LoglinePage.LoadingString(),
			CopySavePath:  copySavePath, SelectionEnabled: false, WrapText: true, RequestInput: false,
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
	case JobSpecPage:
		return "job spec"
	case JobEventsPage, AllocEventsPage:
		return "events"
	case AllEventsPage:
		return "all events"
	case JobEventPage, AllocEventPage, AllEventPage:
		return "event"
	case AllocationsPage:
		return "allocations"
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
		return AllocationsPage
	case JobEventsPage:
		return JobEventPage
	case AllocEventsPage:
		return AllocEventPage
	case AllEventsPage:
		return AllEventPage
	case AllocationsPage:
		return LogsPage
	case LogsPage:
		return LoglinePage
	}
	return p
}

func (p Page) Backward() Page {
	switch p {
	case JobSpecPage:
		return JobsPage
	case JobEventsPage:
		return JobsPage
	case JobEventPage:
		return JobEventsPage
	case AllocEventsPage:
		return AllocationsPage
	case AllocEventPage:
		return AllocEventsPage
	case AllEventsPage:
		return JobsPage
	case AllEventPage:
		return AllEventsPage
	case AllocationsPage:
		return JobsPage
	case ExecPage:
		return AllocationsPage
	case AllocSpecPage:
		return AllocationsPage
	case LogsPage:
		return AllocationsPage
	case LoglinePage:
		return LogsPage
	}
	return p
}

func (p Page) GetFilterPrefix(jobID, taskName, allocID string, eventTopics Topics, eventNamespace string) string {
	switch p {
	case JobsPage:
		return "Jobs"
	case JobSpecPage:
		return fmt.Sprintf("Job Spec for %s", style.Bold.Render(jobID))
	case JobEventsPage:
		return fmt.Sprintf("Events for %s (%s)", jobID, getTopicNames(eventTopics))
	case JobEventPage:
		return fmt.Sprintf("Event for %s", jobID)
	case AllocEventsPage:
		return fmt.Sprintf("Events for %s %s", style.Bold.Render(jobID), formatter.ShortAllocID(allocID))
	case AllocEventPage:
		return fmt.Sprintf("Event for %s %s", style.Bold.Render(jobID), formatter.ShortAllocID(allocID))
	case AllEventsPage:
		return fmt.Sprintf("All Events in %s (%s)", eventNamespace, formatEventTopics(eventTopics))
	case AllEventPage:
		return fmt.Sprintf("Event")
	case AllocationsPage:
		return fmt.Sprintf("Allocations for %s", style.Bold.Render(jobID))
	case ExecPage:
		return fmt.Sprintf("Exec for %s %s", style.Bold.Render(taskName), formatter.ShortAllocID(allocID))
	case AllocSpecPage:
		return fmt.Sprintf("Allocation Spec for %s %s", style.Bold.Render(taskName), formatter.ShortAllocID(allocID))
	case LogsPage:
		return fmt.Sprintf("Logs for %s %s", style.Bold.Render(taskName), formatter.ShortAllocID(allocID))
	case LoglinePage:
		return fmt.Sprintf("Log Line for %s %s", style.Bold.Render(taskName), formatter.ShortAllocID(allocID))
	default:
		panic("page not found")
	}
}

type EventsStream struct {
	Chan      <-chan *api.Events
	Topics    Topics
	Namespace string
}

type PageLoadedMsg struct {
	Page        Page
	TableHeader []string
	AllPageRows []page.Row
	Connection  EventsStream
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
		output += style.KeyHelpKey.Render(km.Help().Key) + " " + style.KeyHelpDescription.Render(km.Help().Desc) + "    "
	}
	output = strings.TrimSpace(output)
	return output
}

func changeKeyHelp(k *key.Binding, h string) {
	k.SetHelp(k.Help().Key, h)
}

func GetPageKeyHelp(currentPage Page, filterFocused, filterApplied, saving, enteringInput, inPty, webSocketConnected bool, logType LogType) string {
	firstRow := []key.Binding{keymap.KeyMap.Exit}

	if currentPage.DoesReload() && !saving && !filterFocused {
		firstRow = append(firstRow, keymap.KeyMap.Reload)
	}

	viewportKeyMap := viewport.GetKeyMap()
	secondRow := []key.Binding{viewportKeyMap.Save, keymap.KeyMap.Wrap}
	thirdRow := []key.Binding{viewportKeyMap.Down, viewportKeyMap.Up, viewportKeyMap.PageDown, viewportKeyMap.PageUp}

	var fourthRow []key.Binding
	if nextPage := currentPage.Forward(); nextPage != currentPage {
		changeKeyHelp(&keymap.KeyMap.Forward, currentPage.Forward().String())
		fourthRow = append(fourthRow, keymap.KeyMap.Forward)
	}

	if filterApplied {
		changeKeyHelp(&keymap.KeyMap.Back, "remove filter")
		fourthRow = append(fourthRow, keymap.KeyMap.Back)
	} else if prevPage := currentPage.Backward(); prevPage != currentPage {
		changeKeyHelp(&keymap.KeyMap.Back, fmt.Sprintf("%s", currentPage.Backward().String()))
		fourthRow = append(fourthRow, keymap.KeyMap.Back)
	}

	if currentPage == JobsPage || currentPage == AllocationsPage {
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
	}

	if currentPage == AllocationsPage {
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
			secondRow = []key.Binding{keymap.KeyMap.Back}
			return getShortHelp(firstRow) + "\n" + getShortHelp(secondRow)
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
