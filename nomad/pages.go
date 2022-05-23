package nomad

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"wander/components/page"
	"wander/components/viewport"
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
	return fmt.Sprintf("loading %s...", p.String())
}

func (p Page) ReloadingString() string {
	return fmt.Sprintf("Reloading %s...", p.String())
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

type PageLoadedMsg struct {
	Page        Page
	TableHeader []string
	AllPageData []page.Row
}

type ChangePageMsg struct{ NewPage Page }

func GetPageKeyHelp(currentPage Page) string {
	keyHelper := help.New()
	keyHelper.ShortSeparator = "    "
	keyHelper.Styles.ShortKey = style.KeyHelpKey
	keyHelper.Styles.ShortDesc = style.KeyHelpDescription
	viewportKeyMap := viewport.GetKeyMap()

	alwaysShown := []key.Binding{keymap.KeyMap.Exit, viewportKeyMap.Save}
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

	firstRow := keyHelper.ShortHelpView(alwaysShown)

	viewportKm := viewport.GetKeyMap()
	viewportAlwaysShown := []key.Binding{viewportKm.Down, viewportKm.Up, viewportKm.PageDown, viewportKm.PageUp}
	secondRow := keyHelper.ShortHelpView(viewportAlwaysShown)

	final := firstRow + "\n" + secondRow
	return final
}
