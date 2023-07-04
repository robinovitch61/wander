package header

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/robinovitch61/wander/internal/tui/style"
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
			logoStyle.Padding(0).Margin(0).Render("WANDER"),
			style.KeyHelp.Render(m.keyHelp),
			style.Regular.Copy().Padding(0, 2, 0, 0).Render(m.version),
			clusterUrl,
		)
	}
	logo := logoStyle.Render(m.logo)
	left := style.Header.Render(lipgloss.JoinVertical(lipgloss.Center, logo, m.version, clusterUrl))
	styledKeyHelp := style.KeyHelp.Render(m.keyHelp)
	return lipgloss.JoinHorizontal(lipgloss.Center, left, styledKeyHelp)
}

func (m Model) ViewHeight() int {
	return lipgloss.Height(m.View())
}

func (m *Model) SetKeyHelp(keyHelp string) {
	m.keyHelp = keyHelp
}

func (m *Model) ToggleCompact() {
	m.compact = !m.compact
}
