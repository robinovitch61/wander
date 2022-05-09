package header

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type Model struct {
	content      []string
	filterString string
}

func New(nomadUrl, filterString string) (m Model) {
	content := []string{fmt.Sprintf("Cluster URL: %s", nomadUrl)}
	if filterString != "" {
		content = append(content, fmt.Sprintf("Filter: %s", filterString))
	}
	return Model{content, filterString}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update() (Model, tea.Cmd) {
	return m, nil
}

func (m *Model) SetFilterString(s string) {
	// TODO LEO: should this be in the Update function?
	m.filterString = s
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
