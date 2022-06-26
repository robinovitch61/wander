package header

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/robinovitch61/wander/internal/tui/style"
	"strings"
)

type Model struct {
	logo, nomadUrl, version, KeyHelp string
}

func New(logo string, nomadUrl, version, keyHelp string) (m Model) {
	return Model{logo: logo, nomadUrl: nomadUrl, version: version, KeyHelp: keyHelp}
}

func (m Model) View() string {
	logo := style.Logo.Render(m.logo)
	clusterUrl := style.ClusterUrl.Render(m.nomadUrl)
	left := style.Header.Render(lipgloss.JoinVertical(lipgloss.Center, logo, clusterUrl, m.version))
	styledKeyHelp := style.KeyHelp.Render(m.KeyHelp)
	return lipgloss.JoinHorizontal(lipgloss.Center, left, styledKeyHelp)
}

func (m Model) ViewHeight() int {
	return len(strings.Split(m.View(), "\n"))
}
