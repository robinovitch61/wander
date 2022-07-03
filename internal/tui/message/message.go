package message

import tea "github.com/charmbracelet/bubbletea"

type ErrMsg struct{ Err error }

func (e ErrMsg) Error() string { return e.Err.Error() }

func ErrCmd(err error) func() tea.Msg {
	return func() tea.Msg {
		return ErrMsg{err}
	}
}

type PageInputReceivedMsg struct {
	Input string
}
