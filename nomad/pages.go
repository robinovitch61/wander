package nomad

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/robinovitch61/wander/components/page"
	"github.com/robinovitch61/wander/components/viewport"
	"github.com/robinovitch61/wander/formatter"
	"github.com/robinovitch61/wander/keymap"
	"github.com/robinovitch61/wander/style"
	"strings"
)

type Page int8

const (
	Unset Page = iota
	JobsPage
	JobSpecPage
	AllocationsPage
	AllocSpecPage
	LogsPage
	LoglinePage
)

func (p Page) Loads() bool {
	noLoadPages := []Page{LoglinePage}
	for _, noLoadPage := range noLoadPages {
		if noLoadPage == p {
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
	case AllocationsPage:
		return "allocations"
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
	case AllocationsPage:
		return JobsPage
	case AllocSpecPage:
		return AllocationsPage
	case LogsPage:
		return AllocationsPage
	case LoglinePage:
		return LogsPage
	}
	return p
}

func (p Page) GetFilterPrefix(jobID, taskName, allocID string) string {
	switch p {
	case JobsPage:
		return "Jobs"
	case JobSpecPage:
		return fmt.Sprintf("Job Spec for %s", style.Bold.Render(jobID))
	case AllocationsPage:
		return fmt.Sprintf("Allocations for %s", style.Bold.Render(jobID))
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

type PageLoadedMsg struct {
	Page        Page
	TableHeader []string
	AllPageData []page.Row
}

type ChangePageMsg struct{ NewPage Page }

func getShortHelp(bindings []key.Binding) string {
	var output string
	for _, km := range bindings {
		output += style.KeyHelpKey.Render(km.Help().Key) + " " + style.KeyHelpDescription.Render(km.Help().Desc) + "    "
	}
	output = strings.TrimSpace(output)
	return output
}

func GetPageKeyHelp(currentPage Page) string {
	firstRow := []key.Binding{keymap.KeyMap.Exit, keymap.KeyMap.Wrap}

	if currentPage.Loads() {
		firstRow = append(firstRow, keymap.KeyMap.Reload)
	}

	viewportKeyMap := viewport.GetKeyMap()
	secondRow := []key.Binding{viewportKeyMap.Down, viewportKeyMap.Up, viewportKeyMap.PageDown, viewportKeyMap.PageUp, viewportKeyMap.Save}

	var thirdRow []key.Binding
	if nextPage := currentPage.Forward(); nextPage != currentPage {
		keymap.KeyMap.Forward.SetHelp(keymap.KeyMap.Forward.Help().Key, fmt.Sprintf("view %s", currentPage.Forward().String()))
		thirdRow = append(thirdRow, keymap.KeyMap.Forward)
	}

	if prevPage := currentPage.Backward(); prevPage != currentPage {
		keymap.KeyMap.Back.SetHelp(keymap.KeyMap.Back.Help().Key, fmt.Sprintf("view %s", currentPage.Backward().String()))
		thirdRow = append(thirdRow, keymap.KeyMap.Back)
	}

	if currentPage == JobsPage || currentPage == AllocationsPage {
		thirdRow = append(thirdRow, keymap.KeyMap.Spec)
	} else if currentPage == LogsPage {
		thirdRow = append(thirdRow, keymap.KeyMap.StdOut)
		thirdRow = append(thirdRow, keymap.KeyMap.StdErr)
	}

	var final string
	for _, row := range [][]key.Binding{firstRow, secondRow, thirdRow} {
		final += getShortHelp(row) + "\n"
	}

	return strings.TrimRight(final, "\n")
}
