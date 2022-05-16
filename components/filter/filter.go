package filter

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	keyMap = getKeyMap()
)

type Model struct {
	prefix             string
	onUpdateFilter     func()
	keyMap             filterKeyMap
	Filter             string
	EditingFilter      bool
	PrefixStyle        lipgloss.Style
	FilterStyle        lipgloss.Style
	EditingFilterStyle lipgloss.Style
}

func New(prefix string) Model {
	return Model{
		prefix: prefix,
		keyMap: keyMap,
		PrefixStyle: lipgloss.NewStyle().
			Padding(0, 3).
			Border(lipgloss.NormalBorder(), true).
			//BorderForeground(lipgloss.Color("#000000")).
			//BorderBackground(lipgloss.Color("6")).
			Foreground(lipgloss.Color("#FFFFFF")),
		//Background(lipgloss.Color("6")),
		FilterStyle: lipgloss.NewStyle().
			Margin(0, 1).
			Foreground(lipgloss.Color("#8E8E8E")),
		EditingFilterStyle: lipgloss.NewStyle().
			Margin(0, 1).
			Padding(0, 1).
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("6")),
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.keyMap.Back) {
			m.SetFiltering(false, true)
		}

		if m.EditingFilter {
			switch {
			case key.Matches(msg, m.keyMap.Forward):
				m.SetFiltering(false, false)
			default:
				switch msg.Type {
				case tea.KeyBackspace:
					if len(m.Filter) > 0 {
						if msg.Alt {
							m.SetFilter("")
						} else {
							m.SetFilter(m.Filter[:len(m.Filter)-1])
						}
					}
				case tea.KeyRunes:
					m.SetFilter(m.Filter + msg.String())
				}
			}
		} else if key.Matches(msg, m.keyMap.Filter) {
			m.SetFiltering(true, false)
		}
	}

	return m, nil
}

func (m Model) View() string {
	var filterString string
	switch {
	case len(m.Filter) > 0:
		filterString = fmt.Sprintf("filter: %s", m.Filter)
	case m.EditingFilter:
		filterString = "type to filter"
	default:
		filterString = "<'/' to filter>"
	}
	return lipgloss.JoinHorizontal(lipgloss.Center, m.PrefixStyle.Render(m.prefix), m.formatFilterString(filterString))
}

func (m Model) ViewHeight() int {
	return lipgloss.Height(m.View())
}

func (m *Model) SetFilter(filter string) {
	m.Filter = filter
}

func (m *Model) SetFiltering(isEditingFilter, clearFilter bool) {
	m.EditingFilter = isEditingFilter
	if clearFilter {
		m.SetFilter("")
	}
}

func (m Model) formatFilterString(s string) string {
	if !m.EditingFilter && len(m.Filter) == 0 {
		return m.FilterStyle.Render(s)
	}
	return m.EditingFilterStyle.Render(s)
}
