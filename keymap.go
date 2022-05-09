package main

import (
	"github.com/charmbracelet/bubbles/key"
)

type mainKeyMap struct {
	Exit   key.Binding
	Enter  key.Binding
	Back   key.Binding
	Reload key.Binding
	Filter key.Binding
}

// getKeyMap returns the main key mappings
func getKeyMap() mainKeyMap {
	return mainKeyMap{
		Exit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q/ctrl+c", "exit"),
		),
		Enter: key.NewBinding(
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
	}
}
