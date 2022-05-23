package page

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
	"wander/components/filter"
	"wander/components/viewport"
	"wander/keymap"
)

type Model struct {
	width, height int
	pageData      data
	viewport      viewport.Model
	filter        filter.Model
	loadingString string
	Loading       bool
}

func New(
	width, height int,
	filterPrefix, loadingString string,
) Model {
	pageFilter := filter.New(filterPrefix)
	pageViewport := viewport.New(width, height-pageFilter.ViewHeight())
	model := Model{
		width:         width,
		height:        height,
		viewport:      pageViewport,
		filter:        pageFilter,
		loadingString: loadingString,
		Loading:       true,
	}
	return model
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	if m.viewport.Saving() {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.filter.Focused() {
			switch {
			case key.Matches(msg, keymap.KeyMap.Forward):
				m.filter.Blur()

			case key.Matches(msg, keymap.KeyMap.Back):
				m.clearFilter()
			}
		} else {
			switch {
			case key.Matches(msg, keymap.KeyMap.Filter):
				m.filter.Focus()
				return m, nil
			}

			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}

		prevFilter := m.filter.Filter
		m.filter, cmd = m.filter.Update(msg)
		if m.filter.Filter != prevFilter {
			m.updateViewport()
		}
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	content := fmt.Sprintf(m.loadingString)
	if !m.Loading {
		content = m.viewport.View()
	}
	return lipgloss.JoinVertical(lipgloss.Left, m.filter.View(), content)
}

func (m *Model) SetWindowSize(width, height int) {
	m.width, m.height = width, height
	m.viewport.SetSize(width, height-m.filter.ViewHeight())
}

func (m Model) GetSelectedPageRow() (Row, error) {
	if filtered := m.pageData.Filtered; len(filtered) > 0 && m.viewport.CursorRow < len(filtered) {
		return filtered[m.viewport.CursorRow], nil
	}
	return Row{}, fmt.Errorf("bad thing")
}

func (m Model) SetHeader(header []string) {
	m.viewport.SetHeader(header)
}

func (m *Model) SetAllPageData(allPageData []Row) {
	m.pageData.All = allPageData
	m.updateViewport()
}

func (m Model) FilterFocused() bool {
	return m.filter.Focused()
}

func (m *Model) clearFilter() {
	m.filter.BlurAndClear()
	m.updateViewport()
}

func (m *Model) updateViewport() {
	m.viewport.Highlight = m.filter.Filter
	m.updateFilteredData()
	m.viewport.SetContent(rowsToStrings(m.pageData.Filtered))
	m.viewport.SetCursorRow(0)
}

func (m *Model) updateFilteredData() {
	if m.filter.Filter == "" {
		m.pageData.Filtered = m.pageData.All
	} else {
		var filteredData []Row
		for _, entry := range m.pageData.All {
			if strings.Contains(entry.Row, m.filter.Filter) {
				filteredData = append(filteredData, entry)
			}
		}
		m.pageData.Filtered = filteredData
	}
}
