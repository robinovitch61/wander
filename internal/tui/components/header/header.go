package header

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/robinovitch61/wander/internal/tui/style"
	"strings"
)

type Model struct {
	logo, logoColor, nomadUrl, version, keyHelp string
	compact                                     bool
}

func New(logo string, logoColor string, nomadUrl, version, keyHelp string) (m Model) {
	return Model{logo: logo, logoColor: logoColor, nomadUrl: nomadUrl, version: version, keyHelp: keyHelp}
}

func (m Model) View() string {
	logoStyle := style.Logo.Copy()
	if m.logoColor != "" {
		logoStyle.Foreground(lipgloss.Color(m.logoColor))
	}
	clusterUrl := style.ClusterUrl.Render(m.nomadUrl)
	if m.compact {
		return lipgloss.JoinHorizontal(
			lipgloss.Center,
			logoStyle.Padding(0).Render("WANDER")+" ",
			m.version+" ",
			clusterUrl,
			style.KeyHelp.Render(m.keyHelp),
		)
	}
	logo := logoStyle.Render(m.logo)
	left := style.Header.Render(lipgloss.JoinVertical(lipgloss.Center, logo, m.version, clusterUrl))
	styledKeyHelp := style.KeyHelp.Render(m.keyHelp)
	return lipgloss.JoinHorizontal(lipgloss.Center, left, styledKeyHelp)
}

func (m Model) ViewHeight() int {
	return len(strings.Split(m.View(), "\n"))
}

func (m *Model) SetKeyHelp(keyHelp string) {
	m.keyHelp = keyHelp
}

func (m *Model) ToggleCompact() {
	m.compact = !m.compact
}
