package toast

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinovitch61/wander/internal/dev"
	"github.com/robinovitch61/wander/internal/tui/constants"
	"github.com/robinovitch61/wander/internal/tui/style"
	"sync"
	"time"
)

var (
	lastID int
	idMtx  sync.Mutex
)

type Model struct {
	id           int
	message      string
	timeout      time.Duration
	initialized  bool
	Visible      bool
	MessageStyle lipgloss.Style
}

func New(message string) Model {
	return Model{
		id:           nextID(),
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

	switch msg := msg.(type) {
	case TimeoutMsg:
		if msg.ID > 0 && msg.ID != m.id {
			return m, nil
		}

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

// Value and Cmds

type TimeoutMsg struct {
	ID int
}

func (m Model) timeoutAfterDuration() tea.Cmd {
	return tea.Tick(m.timeout, func(t time.Time) tea.Msg { return TimeoutMsg{m.id} })
}

// Helpers

func nextID() int {
	idMtx.Lock()
	defer idMtx.Unlock()
	lastID++
	return lastID
}
