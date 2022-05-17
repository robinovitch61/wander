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
	styledKeyHelp := style.KeyHelp.Render(m.KeyHelp)
	top := lipgloss.JoinHorizontal(lipgloss.Center, logo, styledKeyHelp)
	clusterUrl := style.Bold.Copy().Padding(0, 0, 0, 1).Render(fmt.Sprintf("URL: %s", m.nomadUrl))
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
	return lipgloss.JoinVertical(lipgloss.Left, top, clusterUrl)
}

func (m Model) ViewHeight() int {
	return len(strings.Split(m.View(), "\n"))
}

//func (m Model) HasFilter() bool {
//	return len(m.Filter) > 0
//}
