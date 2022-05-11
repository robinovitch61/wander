package header

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type Model struct {
	content       []string
	Filter        string
	EditingFilter bool
}

func New(nomadUrl, filterString string) (m Model) {
	content := []string{fmt.Sprintf("Cluster URL: %s", nomadUrl)}
	if filterString != "" {
		content = append(content, fmt.Sprintf("Filter: %s", filterString))
	}
	return Model{content: content, Filter: filterString}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m Model) formatFilterString(s string) string {
	if !m.EditingFilter {
		return s
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Background(lipgloss.Color("6")).
		Render(s)
}

func (m Model) View() string {
	viewString := strings.Join(m.content, "\n")
	viewString += "\n" + m.formatFilterString("Filter: ")
	if m.EditingFilter {
		if m.Filter == "" {
			viewString += m.formatFilterString("<type to filter>")
		}
	} else {
		if m.Filter == "" {
			viewString += m.formatFilterString("none ('/' to filter)")
		}
	}
	viewString += m.formatFilterString(m.Filter)
	return lipgloss.NewStyle().
		Padding(0, 1).
		Border(lipgloss.RoundedBorder(), true).
		Render(viewString)
}

func (m Model) ViewHeight() int {
	return len(strings.Split(m.View(), "\n"))
}
