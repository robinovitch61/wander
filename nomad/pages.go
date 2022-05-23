package nomad

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"strings"
	"wander/components/page"
	"wander/components/viewport"
	"wander/formatter"
	"wander/keymap"
	"wander/style"
)

type Page int8

const (
	Unset Page = iota
	JobsPage
	AllocationsPage
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
	case AllocationsPage:
		return "allocations"
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
	case AllocationsPage:
		return JobsPage
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
	case AllocationsPage:
		return fmt.Sprintf("Allocations for %s", jobID)
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
	alwaysShown := []key.Binding{keymap.KeyMap.Exit}

	if currentPage != LoglinePage {
		alwaysShown = append(alwaysShown, keymap.KeyMap.Reload)
	}

	if nextPage := currentPage.Forward(); nextPage != currentPage {
		keymap.KeyMap.Forward.SetHelp(keymap.KeyMap.Forward.Help().Key, fmt.Sprintf("view %s", currentPage.Forward().String()))
		alwaysShown = append(alwaysShown, keymap.KeyMap.Forward)
	}

	if prevPage := currentPage.Backward(); prevPage != currentPage {
		keymap.KeyMap.Back.SetHelp(keymap.KeyMap.Back.Help().Key, fmt.Sprintf("view %s", currentPage.Backward().String()))
		alwaysShown = append(alwaysShown, keymap.KeyMap.Back)
	}

	if currentPage == LogsPage {
		alwaysShown = append(alwaysShown, keymap.KeyMap.StdOut)
		alwaysShown = append(alwaysShown, keymap.KeyMap.StdErr)
	}

	firstRow := getShortHelp(alwaysShown)

	viewportKeyMap := viewport.GetKeyMap()
	viewportAlwaysShown := []key.Binding{viewportKeyMap.Down, viewportKeyMap.Up, viewportKeyMap.PageDown, viewportKeyMap.PageUp, viewportKeyMap.Save}
	secondRow := getShortHelp(viewportAlwaysShown)

	return firstRow + "\n" + secondRow
}
