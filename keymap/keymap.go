package keymap

import (
	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	Exit    key.Binding
	Forward key.Binding
	Back    key.Binding
	Reload  key.Binding
	Filter  key.Binding
	StdOut  key.Binding
	StdErr  key.Binding
	Spec    key.Binding
	Exec    key.Binding
}

var KeyMap = keyMap{
	Exit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q/ctrl+c", "exit"),
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
	Spec: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "view spec"),
	),
	Exec: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "exec"),
	),
}
