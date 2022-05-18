package header

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
	"wander/style"
)

type Model struct {
	logo     string
	nomadUrl string
	KeyHelp  string
}

func New(logo string, nomadUrl, keyHelp string) (m Model) {
	return Model{logo: logo, nomadUrl: nomadUrl, KeyHelp: keyHelp}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	logo := style.Logo.Render(m.logo)
	clusterUrl := style.ClusterUrl.Render(fmt.Sprintf("URL: %s", m.nomadUrl))
	left := style.Header.Render(lipgloss.JoinVertical(lipgloss.Center, logo, clusterUrl))
	styledKeyHelp := style.KeyHelp.Render(m.KeyHelp)
	return lipgloss.JoinHorizontal(lipgloss.Center, left, styledKeyHelp)
}

func (m Model) ViewHeight() int {
	return len(strings.Split(m.View(), "\n"))
}
