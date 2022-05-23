package pages

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
)

type Page int8

const (
	Unset Page = iota
	Jobs
	Allocations
	Logs
	Logline
)

func (p Page) String() string {
	switch p {
	case Unset:
		return "undefined"
	case Jobs:
		return "jobs"
	case Allocations:
		return "allocations"
	case Logs:
		return "logs"
	case Logline:
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
	case Jobs:
		return Allocations
	case Allocations:
		return Logs
	case Logs:
		return Logline
	}
	return p
}

func (p Page) Backward() Page {
	switch p {
	case Allocations:
		return Jobs
	case Logs:
		return Allocations
	case Logline:
		return Logs
	}
	return p
}

type ChangePageMsg struct{ NewPage Page }

func ToJobsPageCmd() tea.Msg {
	return ChangePageMsg{NewPage: Jobs}
}

func ToAllocationsPageCmd() tea.Msg {
	return ChangePageMsg{NewPage: Allocations}
}

func ToLogsPageCmd() tea.Msg {
	return ChangePageMsg{NewPage: Logs}
}

func ToLoglinePageCmd() tea.Msg {
	return ChangePageMsg{NewPage: Logline}
}

// TODO LEO: Make this a map of pages.Page -> rows of help?
// func GetPageKeyHelp(currentPage page.Model) string {
// 	keyHelper := help.New()
// 	keyHelper.ShortSeparator = "    "
// 	keyHelper.Styles.ShortKey = style.KeyHelpKey
// 	keyHelper.Styles.ShortDesc = style.KeyHelpDescription
// 	viewportKeyMap := viewport.GetKeyMap()
//
// 	alwaysShown := []key.Binding{KeyMap.Exit, viewportKeyMap.Save}
// 	if currentPage != pages.Logline {
// 		alwaysShown = append(alwaysShown, KeyMap.Reload)
// 	}
//
// 	if nextPage := currentPage.Forward(); nextPage != currentPage {
// 		KeyMap.Forward.SetHelp(KeyMap.Forward.Help().Key, fmt.Sprintf("view %s", currentPage.Forward().String()))
// 		alwaysShown = append(alwaysShown, KeyMap.Forward)
// 	}
//
// 	if prevPage := currentPage.Backward(); prevPage != currentPage {
// 		KeyMap.Back.SetHelp(KeyMap.Back.Help().Key, fmt.Sprintf("view %s", currentPage.Backward().String()))
// 		alwaysShown = append(alwaysShown, KeyMap.Back)
// 	}
//
// 	firstRow := keyHelper.ShortHelpView(alwaysShown)
//
// 	viewportKm := viewport.GetKeyMap()
// 	viewportAlwaysShown := []key.Binding{viewportKm.Down, viewportKm.Up, viewportKm.PageDown, viewportKm.PageUp}
// 	secondRow := keyHelper.ShortHelpView(viewportAlwaysShown)
//
// 	final := firstRow + "\n" + secondRow
// 	return final
// }
