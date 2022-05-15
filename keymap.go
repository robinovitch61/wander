package main

import (
	"github.com/charmbracelet/bubbles/key"
)

type mainKeyMap struct {
	Exit    key.Binding
	Forward key.Binding
	Back    key.Binding
}

func getMainKeyMap() mainKeyMap {
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
	}
}
