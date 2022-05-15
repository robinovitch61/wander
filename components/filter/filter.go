package filter

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"wander/dev"
)

type Model struct {
	keyMap        KeyMap
	Filter        string
	EditingFilter bool
}

func New() Model {
	return Model{}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Back):
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
			//m.updateJobViewport()
		} else if key.Matches(msg, m.keyMap.Filter) {
			m.SetFiltering(true, false)
		}

		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
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
	dev.Debug(fmt.Sprintf("isEditingFilter %t clearFilter %t", isEditingFilter, clearFilter))
	m.EditingFilter = isEditingFilter
	if clearFilter {
		m.SetFilter("")
	}
	//m.setHeaderKeyHelp()
}
