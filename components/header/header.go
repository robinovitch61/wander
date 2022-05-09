package header

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type Model struct {
	content []string
}

func New(nomadUrl string) (m Model) {
	content := []string{fmt.Sprintf("Cluster URL: %s", nomadUrl)}
	return Model{content}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update() (Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	viewString := strings.Join(m.content, "\n")
	return lipgloss.NewStyle().
		Padding(0, 1).
		Border(lipgloss.RoundedBorder(), true).
		Render(viewString)
}

func (m Model) ViewHeight() int {
	return len(strings.Split(m.View(), "\n"))
}
