package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"wander/components/viewport"
	"wander/page"
	"wander/style"
)

type mainKeyMap struct {
	Exit    key.Binding
	Forward key.Binding
	Back    key.Binding
	Reload  key.Binding
	Filter  key.Binding
	StdOut  key.Binding
	StdErr  key.Binding
}

// getKeyMap returns the main key mappings
func getKeyMap() mainKeyMap {
	return mainKeyMap{
		Exit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "exit"),
		),
		Forward: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "enter"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Reload: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reload"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		StdOut: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "stdout"),
		),
		StdErr: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "stderr"),
		),
	}
}

func KeyMapView(currentPage page.Page, isFiltering, hasFilter bool) string {
	keyHelper := help.New()
	keyHelper.ShortSeparator = "    "
	keyHelper.Styles.ShortKey = style.KeyHelpKeyStyle
	keyHelper.Styles.ShortDesc = style.KeyHelpDescriptionStyle

	mainKm := getKeyMap()
	var alwaysShown []key.Binding
	if !isFiltering {
		alwaysShown = append(alwaysShown, []key.Binding{mainKm.Exit, mainKm.Filter, mainKm.Reload}...)

		if nextPage := currentPage.Forward(); nextPage != currentPage {
			mainKm.Forward.SetHelp(mainKm.Forward.Help().Key, fmt.Sprintf("view %s", currentPage.Forward().String()))
			alwaysShown = append(alwaysShown, mainKm.Forward)
		}

		if !hasFilter {
			if prevPage := currentPage.Backward(); prevPage != currentPage {
				mainKm.Back.SetHelp(mainKm.Back.Help().Key, fmt.Sprintf("view %s", currentPage.Backward().String()))
				alwaysShown = append(alwaysShown, mainKm.Back)
			}
		} else {
			mainKm.Back.SetHelp(mainKm.Back.Help().Key, "clear filter")
			alwaysShown = append(alwaysShown, mainKm.Back)
		}
	} else {
		mainKm.Forward.SetHelp(mainKm.Forward.Help().Key, "keep filter")
		alwaysShown = append(alwaysShown, mainKm.Forward)

		mainKm.Back.SetHelp(mainKm.Back.Help().Key, "discard filter")
		alwaysShown = append(alwaysShown, mainKm.Back)
	}
	firstRow := keyHelper.ShortHelpView(alwaysShown)

	viewportKm := viewport.GetKeyMap()
	viewportAlwaysShown := []key.Binding{viewportKm.Up, viewportKm.Down, viewportKm.PageUp, viewportKm.PageDown}
	secondRow := keyHelper.ShortHelpView(viewportAlwaysShown)

	final := firstRow
	if !isFiltering {
		final += "\n" + secondRow
	}
	return final
}
