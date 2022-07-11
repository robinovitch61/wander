package nomad

import (
	"bufio"
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinovitch61/wander/internal/tui/components/page"
	"github.com/robinovitch61/wander/internal/tui/components/viewport"
	"github.com/robinovitch61/wander/internal/tui/constants"
	"github.com/robinovitch61/wander/internal/tui/formatter"
	"github.com/robinovitch61/wander/internal/tui/keymap"
	"github.com/robinovitch61/wander/internal/tui/style"
	"io"
	"strings"
	"time"
)

type Page int8

const (
	Unset Page = iota
	JobsPage
	JobSpecPage
	EventsPage
	EventPage
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
		EventsPage: {
			Width: width, Height: height,
			LoadingString: EventsPage.LoadingString(),
			CopySavePath:  copySavePath, SelectionEnabled: true, WrapText: false, RequestInput: false,
		},
		EventPage: {
			Width: width, Height: height,
			LoadingString: EventPage.LoadingString(),
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
	noLoadPages := []Page{LoglinePage, EventPage}
	for _, noLoadPage := range noLoadPages {
		if noLoadPage == p {
			return false
		}
	}
	return true
}

func (p Page) DoesReload() bool {
	noReloadPages := []Page{LoglinePage, EventsPage, EventPage, ExecPage}
	for _, noReloadPage := range noReloadPages {
		if noReloadPage == p {
			return false
		}
	}
	return true
}

func (p Page) doesUpdate() bool {
	noUpdatePages := []Page{
		LoglinePage,   // doesn't load
		ExecPage,      // doesn't reload
		LogsPage,      // currently makes scrolling impossible - solve in https://github.com/robinovitch61/wander/issues/1
		JobSpecPage,   // would require changes to make scrolling possible
		AllocSpecPage, // would require changes to make scrolling possible
		EventsPage,    // constant connection, streams data
		EventPage,     // doesn't load
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
	case EventsPage:
		return "events"
	case EventPage:
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
	case EventsPage:
		return EventPage
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
	case EventsPage:
		return JobsPage
	case EventPage:
		return EventsPage
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

func (p Page) GetFilterPrefix(jobID, taskName, allocID, eventTopics, eventNamespace string) string {
	switch p {
	case JobsPage:
		return "Jobs"
	case JobSpecPage:
		return fmt.Sprintf("Job Spec for %s", style.Bold.Render(jobID))
	case EventsPage:
		return fmt.Sprintf("Events in %s for %s", eventNamespace, formatEventTopics(eventTopics))
	case EventPage:
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

type PersistentConnection struct {
	Reader *bufio.Reader
	Body   io.ReadCloser
}

type PageLoadedMsg struct {
	Page        Page
	TableHeader []string
	AllPageRows []page.Row
	Connection  PersistentConnection
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

func GetPageKeyHelp(currentPage Page, filterFocused, filterApplied, saving, enteringInput, inPty, webSocketConnected bool) string {
	firstRow := []key.Binding{keymap.KeyMap.Exit}

	if currentPage.DoesReload() && !saving && !filterFocused {
		firstRow = append(firstRow, keymap.KeyMap.Reload)
	}

	viewportKeyMap := viewport.GetKeyMap()
	secondRow := []key.Binding{viewportKeyMap.Save, keymap.KeyMap.Wrap}
	thirdRow := []key.Binding{viewportKeyMap.Down, viewportKeyMap.Up, viewportKeyMap.PageDown, viewportKeyMap.PageUp}

	var fourthRow []key.Binding
	if nextPage := currentPage.Forward(); nextPage != currentPage {
		changeKeyHelp(&keymap.KeyMap.Forward, fmt.Sprintf("view %s", currentPage.Forward().String()))
		fourthRow = append(fourthRow, keymap.KeyMap.Forward)
	}

	if filterApplied {
		changeKeyHelp(&keymap.KeyMap.Back, "remove filter")
		fourthRow = append(fourthRow, keymap.KeyMap.Back)
	} else if prevPage := currentPage.Backward(); prevPage != currentPage {
		changeKeyHelp(&keymap.KeyMap.Back, fmt.Sprintf("view %s", currentPage.Backward().String()))
		fourthRow = append(fourthRow, keymap.KeyMap.Back)
	}

	if currentPage == JobsPage || currentPage == AllocationsPage {
		fourthRow = append(fourthRow, keymap.KeyMap.Spec)
	} else if currentPage == LogsPage {
		fourthRow = append(fourthRow, keymap.KeyMap.StdOut)
		fourthRow = append(fourthRow, keymap.KeyMap.StdErr)
	}

	if currentPage == JobsPage {
		fourthRow = append(fourthRow, keymap.KeyMap.Events)
	}

	if currentPage == AllocationsPage {
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
