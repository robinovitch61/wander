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
	ID           int
	message      string
	Timeout      time.Duration
	initialized  bool
	Visible      bool
	MessageStyle lipgloss.Style
}

func New(message string) Model {
	return Model{
		ID:           nextID(),
		message:      message,
		Timeout:      constants.ToastDuration,
		Visible:      true,
		MessageStyle: style.SuccessToast,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	dev.Debug(fmt.Sprintf("toast %T", msg))
	switch msg := msg.(type) {
	case TimeoutMsg:
		if msg.ID > 0 && msg.ID != m.ID {
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

type TimeoutMsg struct {
	ID int
}

func (m Model) timeoutAfterDuration() tea.Cmd {
	return tea.Tick(m.Timeout, func(t time.Time) tea.Msg { return TimeoutMsg{m.ID} })
}

func nextID() int {
	idMtx.Lock()
	defer idMtx.Unlock()
	lastID++
	return lastID
}
