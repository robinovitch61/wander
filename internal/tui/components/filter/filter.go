package filter

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinovitch61/wander/internal/dev"
	"github.com/robinovitch61/wander/internal/tui/style"
)

var (
	keyMap = getKeyMap()
)

type Model struct {
	prefix         string
	onUpdateFilter func()
	keyMap         filterKeyMap
	focus          bool
	Filter         string
}

func New(prefix string) Model {
	return Model{
		prefix: prefix,
		keyMap: keyMap,
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
						m.setFilter("")
					} else {
						m.setFilter(m.Filter[:len(m.Filter)-1])
					}
				}
			case tea.KeyRunes:
				// without this check, matches M+Backspace as \x18\u007f, etc.
				if len(msg.String()) == 1 {
					m.setFilter(m.Filter + msg.String())
				}
			case tea.KeySpace:
				m.setFilter(m.Filter + msg.String())
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
		filterString = "'/' to filter"
	}
	return lipgloss.JoinHorizontal(lipgloss.Center, style.FilterPrefix.Render(m.prefix), m.formatFilterString(filterString))
}

func (m Model) ViewHeight() int {
	return lipgloss.Height(m.View())
}

func (m *Model) SetPrefix(prefix string) {
	m.prefix = prefix
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

func (m *Model) setFilter(filter string) {
	m.Filter = filter
}

func (m Model) formatFilterString(s string) string {
	if !m.focus {
		if len(m.Filter) == 0 {
			return style.Filter.Render(s)
		} else {
			return style.FilterApplied.Render(s)
		}
	}
	return style.FilterEditing.Render(s)
}
