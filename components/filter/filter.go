package filter

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"wander/dev"
)

var (
	keyMap = getKeyMap()
)

type Model struct {
	onUpdateFilter func()
	keyMap         filterKeyMap
	Filter         string
	EditingFilter  bool
}

func New(onUpdateFilter func()) Model {
	return Model{
		onUpdateFilter: onUpdateFilter,
		keyMap:         keyMap,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	dev.Debug(fmt.Sprintf("filter %T", msg))
	var (
		cmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.keyMap.Back) {
			m.SetFiltering(false, true)
			m.onUpdateFilter()
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
				m.onUpdateFilter()
			}
		} else if key.Matches(msg, m.keyMap.Filter) {
			m.SetFiltering(true, false)
		}
	}

	return m, cmd
}

func (m Model) View() string {
	filterView := "Jobs" // TODO LEO: input this to New and put in local state
	if m.EditingFilter || len(m.Filter) > 0 {
		filterView += fmt.Sprintf(" (filter: %s)", m.Filter)
	} else {
		filterView += " <'/' to filter>"
	}
	return filterView
}

func (m *Model) SetFilter(filter string) {
	m.Filter = filter
}

func (m *Model) SetFiltering(isEditingFilter, clearFilter bool) {
	m.EditingFilter = isEditingFilter
	if clearFilter {
		m.SetFilter("")
	}
	//m.setHeaderKeyHelp()
}
