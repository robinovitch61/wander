package keymap

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"wander/components/viewport"
	"wander/pages"
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

func GetPageKeyHelp(currentPage pages.Page) string {
	keyHelper := help.New()
	keyHelper.ShortSeparator = "    "
	keyHelper.Styles.ShortKey = style.KeyHelpKey
	keyHelper.Styles.ShortDesc = style.KeyHelpDescription

	alwaysShown := []key.Binding{KeyMap.Exit}
	if currentPage != pages.Logline {
		alwaysShown = append(alwaysShown, KeyMap.Reload)
	}

	if nextPage := currentPage.Forward(); nextPage != currentPage {
		KeyMap.Forward.SetHelp(KeyMap.Forward.Help().Key, fmt.Sprintf("view %s", currentPage.Forward().String()))
		alwaysShown = append(alwaysShown, KeyMap.Forward)
	}

	if prevPage := currentPage.Backward(); prevPage != currentPage {
		KeyMap.Back.SetHelp(KeyMap.Back.Help().Key, fmt.Sprintf("view %s", currentPage.Backward().String()))
		alwaysShown = append(alwaysShown, KeyMap.Back)
	}

	firstRow := keyHelper.ShortHelpView(alwaysShown)

	viewportKm := viewport.GetKeyMap()
	viewportAlwaysShown := []key.Binding{viewportKm.Down, viewportKm.Up, viewportKm.PageDown, viewportKm.PageUp}
	secondRow := keyHelper.ShortHelpView(viewportAlwaysShown)

	final := firstRow + "\n" + secondRow
	return final
}
