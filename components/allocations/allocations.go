package allocations

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
	"wander/components/filter"
	"wander/components/page"
	"wander/components/viewport"
	"wander/dev"
	"wander/message"
	"wander/style"
)

type Model struct {
	url, token           string
	nomadAllocationData  nomadAllocationData
	width, height        int
	viewport             viewport.Model
	filter               filter.Model
	keyMap               page.KeyMap
	jobID                string
	Loading              bool
	LastSelectedAllocID  string
	LastSelectedTaskName string
}

func New(url, token string, width, height int) Model {
	allocationsFilter := filter.New("Allocations")
	model := Model{
		url:      url,
		token:    token,
		width:    width,
		height:   height,
		viewport: viewport.New(width, height-allocationsFilter.ViewHeight()),
		filter:   allocationsFilter,
		keyMap:   page.GetKeyMap(),
		Loading:  true,
	}
	return model
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	dev.Debug(fmt.Sprintf("allocations %T", msg))

	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {

	case nomadAllocationMsg:
		m.nomadAllocationData.allData = msg
		m.updateAllocationViewport()
		m.Loading = false

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Reload):
			m.Loading = true
			cmds = append(cmds, func() tea.Msg { return message.ViewAllocationsMsg{} })

		case key.Matches(msg, m.keyMap.Forward):
			if !m.filter.EditingFilter {
				selectedAlloc := m.nomadAllocationData.filteredData[m.viewport.CursorRow]
				m.LastSelectedAllocID = selectedAlloc.ID
				m.LastSelectedTaskName = selectedAlloc.TaskName
				return m, func() tea.Msg { return message.ViewLogsMsg{} }
			}

		case key.Matches(msg, m.keyMap.Back):
			if !m.filter.EditingFilter && len(m.filter.Filter) == 0 {
				return m, func() tea.Msg { return message.ViewJobsMsg{} }
			}
		}

		prevFilter := m.filter.Filter
		m.filter, cmd = m.filter.Update(msg)
		if m.filter.Filter != prevFilter {
			m.updateAllocationViewport()
		}
		cmds = append(cmds, cmd)

		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	content := fmt.Sprintf("Loading allocations for %s...", m.jobID)
	if !m.Loading {
		content = m.viewport.View()
	}
	return lipgloss.JoinVertical(lipgloss.Left, m.filter.View(), content)
}

func (m *Model) SetWindowSize(width, height int) {
	m.width, m.height = width, height
	m.viewport.SetSize(width, height-m.filter.ViewHeight())
}

func (m *Model) SetJobID(jobID string) {
	m.jobID = jobID
	m.filter.SetPrefix(fmt.Sprintf("Allocations for %s", style.Bold.Render(jobID)))
}

func (m *Model) updateFilteredAllocationData() {
	var filteredAllocationData []allocationRowEntry
	for _, entry := range m.nomadAllocationData.allData {
		if entry.MatchesFilter(m.filter.Filter) {
			filteredAllocationData = append(filteredAllocationData, entry)
		}
	}
	m.nomadAllocationData.filteredData = filteredAllocationData
}

func (m *Model) updateAllocationViewport() {
	m.viewport.Highlight = m.filter.Filter
	m.updateFilteredAllocationData()
	table := allocationsAsTable(m.nomadAllocationData.filteredData)
	m.viewport.SetHeaderAndContent(
		strings.Join(table.HeaderRows, "\n"),
		strings.Join(table.ContentRows, "\n"),
	)
	m.viewport.SetCursorRow(0)
}
