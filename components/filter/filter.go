package filter

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"wander/dev"
)

var (
	keyMap = getKeyMap()
)

type Model struct {
	prefix             string
	onUpdateFilter     func()
	keyMap             filterKeyMap
	focus              bool
	Filter             string
	PrefixStyle        lipgloss.Style
	FilterStyle        lipgloss.Style
	AppliedFilterStyle lipgloss.Style
	EditingFilterStyle lipgloss.Style
}

func New(prefix string) Model {
	return Model{
		prefix: prefix,
		keyMap: keyMap,
		PrefixStyle: lipgloss.NewStyle().
			Padding(0, 3).
			Border(lipgloss.NormalBorder(), true).
			Foreground(lipgloss.Color("#FFFFFF")),
		FilterStyle: lipgloss.NewStyle().
			Margin(0, 1).
			Foreground(lipgloss.Color("#8E8E8E")),
		AppliedFilterStyle: lipgloss.NewStyle().
			Margin(0, 1).
			Padding(0, 1).
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#00A095")),
		EditingFilterStyle: lipgloss.NewStyle().
			Margin(0, 1).
			Padding(0, 1).
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("6")),
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	dev.Debug(fmt.Sprintf("filter %T", msg))
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.focus {
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
				// without this check, matches M+Backspace as \x18\u007f, etc.
				if len(msg.String()) == 1 {
					dev.Debug(fmt.Sprintf("filter key %s", msg))
					m.SetFilter(m.Filter + msg.String())
				}
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	var filterString string
	switch {
	case len(m.Filter) > 0:
		filterString = fmt.Sprintf("filter: %s", m.Filter)
	case m.focus:
		filterString = "type to filter"
	default:
		filterString = "<'/' to filter>"
	}
	return lipgloss.JoinHorizontal(lipgloss.Center, m.PrefixStyle.Render(m.prefix), m.formatFilterString(filterString))
}

func (m Model) ViewHeight() int {
	return lipgloss.Height(m.View())
}

func (m *Model) SetPrefix(prefix string) {
	m.prefix = prefix
}

func (m *Model) SetFilter(filter string) {
	m.Filter = filter
}

func (m Model) Focused() bool {
	return m.focus
}

func (m *Model) Focus() {
	m.focus = true
}

func (m *Model) Blur() {
	m.focus = false
}

func (m *Model) BlurAndClear() {
	m.Blur()
	m.Filter = ""
}

func (m Model) formatFilterString(s string) string {
	if !m.focus {
		if len(m.Filter) == 0 {
			return m.FilterStyle.Render(s)
		} else {
			return m.AppliedFilterStyle.Render(s)
		}
	}
	return m.EditingFilterStyle.Render(s)
}
