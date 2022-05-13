package header

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
	"wander/style"
)

type Model struct {
	content       []string
	KeyHelp       string
	Filter        string
	EditingFilter bool
}

var (
	clusterURLPrefix = style.Bold.Render("Cluster URL:")
	filterPrefix     = style.Bold.Render("Filter:")
)

func New(nomadUrl, filterString, keyHelp string) (m Model) {
	content := []string{fmt.Sprintf("%s %s", clusterURLPrefix, nomadUrl)}
	if filterString != "" {
		content = append(content, fmt.Sprintf("%s %s", filterPrefix, filterString))
	}
	return Model{content: content, KeyHelp: keyHelp, Filter: filterString}
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
	return style.EditingText.Render(s)
}

func (m Model) View() string {
	viewString := strings.Join(m.content, "\n")
	viewString += "\n" + m.formatFilterString(fmt.Sprintf("%s ", filterPrefix))
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
	styledViewString := style.Header.Render(viewString)
	styledKeyHelp := style.KeyHelp.Render(m.KeyHelp)
	return lipgloss.JoinHorizontal(0.3, styledViewString, styledKeyHelp)
}

func (m Model) ViewHeight() int {
	return len(strings.Split(m.View(), "\n"))
}

func (m Model) HasFilter() bool {
	return len(m.Filter) > 0
}
