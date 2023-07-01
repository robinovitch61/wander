package keymap

import (
	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	Back        key.Binding
	Exec        key.Binding
	Exit        key.Binding
	Compact     key.Binding
	JobEvents   key.Binding
	JobMeta     key.Binding
	AllocEvents key.Binding
	AllEvents   key.Binding
	Filter      key.Binding
	Forward     key.Binding
	Reload      key.Binding
	StdOut      key.Binding
	StdErr      key.Binding
	Spec        key.Binding
	Wrap        key.Binding
}

var KeyMap = keyMap{
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Exec: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "exec"),
	),
	Exit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q/ctrl+c", "exit"),
	),
	Compact: key.NewBinding(
		key.WithKeys("ctrl+h"),
		key.WithHelp("ctrl+h", "compact"),
	),
	JobEvents: key.NewBinding(
		key.WithKeys("v"),
		key.WithHelp("v", "events"),
	),
	JobMeta: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "meta"),
	),
	AllocEvents: key.NewBinding(
		key.WithKeys("v"),
		key.WithHelp("v", "events"),
	),
	AllEvents: key.NewBinding(
		key.WithKeys("V"),
		key.WithHelp("V", "all events"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	Forward: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "enter"),
	),
	Reload: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "reload"),
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
		key.WithHelp("p", "spec"),
	),
	Wrap: key.NewBinding(
		key.WithKeys("ctrl+w"),
		key.WithHelp("ctrl+w", "toggle wrap"),
	),
}
