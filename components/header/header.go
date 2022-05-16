package header

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
	"wander/style"
)

type Model struct {
	logo     []string
	nomadUrl string
	KeyHelp  string
}

var (
//clusterURLPrefix = style.Bold.Render("Nomad:")
//filterPrefix     = style.Bold.Render("Filter:")
)

func New(logo []string, nomadUrl, keyHelp string) (m Model) {
	return Model{logo: logo, nomadUrl: nomadUrl, KeyHelp: keyHelp}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	logo := style.Logo.Render(strings.Join(m.logo, "\n"))
	styledKeyHelp := style.KeyHelp.Render(m.KeyHelp)
	top := lipgloss.JoinHorizontal(lipgloss.Center, logo, styledKeyHelp)
	clusterUrl := style.Bold.Render(m.nomadUrl)
	//headerLeft := lipgloss.JoinVertical(lipgloss.Left, logo, clusterUrl)
	//viewString += "\n" + m.formatFilterString(fmt.Sprintf("%s ", filterPrefix))
	//if m.EditingFilter {
	//	if m.Filter == "" {
	//		viewString += m.formatFilterString("<type to filter>")
	//	}
	//} else {
	//	if m.Filter == "" {
	//		viewString += m.formatFilterString("none ('/' to filter)")
	//	}
	//}
	//viewString += m.formatFilterString(m.Filter)
	//styledViewString := style.Header.Render(viewString)
	return lipgloss.JoinVertical(lipgloss.Left, clusterUrl, top)
}

func (m Model) ViewHeight() int {
	return len(strings.Split(m.View(), "\n"))
}

//func (m Model) HasFilter() bool {
//	return len(m.Filter) > 0
//}
