package keymap

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"wander/components/page"
	"wander/components/viewport"
	"wander/style"
)

type keyMap struct {
	Exit    key.Binding
	Forward key.Binding
	Back    key.Binding
	Reload  key.Binding
	Filter  key.Binding
	StdOut  key.Binding
	StdErr  key.Binding
}

var KeyMap = keyMap{
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

func GetPageKeyHelp(currentPage page.Page, isFiltering, hasFilter bool) string {
	keyHelper := help.New()
	keyHelper.ShortSeparator = "    "
	keyHelper.Styles.ShortKey = style.KeyHelpKey
	keyHelper.Styles.ShortDesc = style.KeyHelpDescription

	var alwaysShown []key.Binding
	if !isFiltering {
		alwaysShown = append(alwaysShown, []key.Binding{KeyMap.Exit, KeyMap.Filter, KeyMap.Reload}...)

		if nextPage := currentPage.Forward(); nextPage != currentPage {
			KeyMap.Forward.SetHelp(KeyMap.Forward.Help().Key, fmt.Sprintf("view %s", currentPage.Forward().String()))
			alwaysShown = append(alwaysShown, KeyMap.Forward)
		}

		if !hasFilter {
			if prevPage := currentPage.Backward(); prevPage != currentPage {
				KeyMap.Back.SetHelp(KeyMap.Back.Help().Key, fmt.Sprintf("view %s", currentPage.Backward().String()))
				alwaysShown = append(alwaysShown, KeyMap.Back)
			}
		} else {
			KeyMap.Back.SetHelp(KeyMap.Back.Help().Key, "clear filter")
			alwaysShown = append(alwaysShown, KeyMap.Back)
		}
	} else {
		KeyMap.Forward.SetHelp(KeyMap.Forward.Help().Key, "keep filter")
		alwaysShown = append(alwaysShown, KeyMap.Forward)

		KeyMap.Back.SetHelp(KeyMap.Back.Help().Key, "discard filter")
		alwaysShown = append(alwaysShown, KeyMap.Back)
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
