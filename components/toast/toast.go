package toast

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"time"
	"wander/constants"
	"wander/dev"
	"wander/style"
)

type Model struct {
	message      string
	timeout      time.Duration
	initialized  bool
	Visible      bool
	MessageStyle lipgloss.Style
}

func New(message string) Model {
	return Model{
		message:      message,
		timeout:      constants.ToastDuration,
		Visible:      true,
		MessageStyle: style.SuccessToast,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	dev.Debug(fmt.Sprintf("toast %T", msg))
	if !m.initialized {
		m.initialized = true
		return m, m.timeoutAfterDuration()
	}

	switch msg.(type) {
	case TimeoutMsg:
		m.Visible = false
	}

	return m, nil
}

func (m Model) View() string {
	if m.Visible {
		return m.MessageStyle.Render(m.message)
	}
	return ""
}

func (m Model) ViewHeight() int {
	return lipgloss.Height(m.View())
}

// Msg and Cmds

type TimeoutMsg struct{}

func (m Model) timeoutAfterDuration() tea.Cmd {
	return tea.Tick(m.timeout, func(t time.Time) tea.Msg { return TimeoutMsg{} })
}
