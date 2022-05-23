package toast

import (
	tea "github.com/charmbracelet/bubbletea"
	"time"
	"wander/constants"
)

// TODO LEO: Make this a whole component with Model/Update/View

type ToastTimeoutMsg struct{}

func GetToastTimeoutCmd() tea.Cmd {
	return tea.Tick(constants.ToastDuration, func(t time.Time) tea.Msg { return ToastTimeoutMsg{} })
}
