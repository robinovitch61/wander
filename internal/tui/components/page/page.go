package page

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinovitch61/wander/internal/dev"
	"github.com/robinovitch61/wander/internal/tui/components/filter"
	"github.com/robinovitch61/wander/internal/tui/components/toast"
	"github.com/robinovitch61/wander/internal/tui/components/viewport"
	"github.com/robinovitch61/wander/internal/tui/keymap"
	"github.com/robinovitch61/wander/internal/tui/message"
	"strings"
)

type Model struct {
	width, height int

	pageData data

	viewport viewport.Model
	filter   filter.Model

	loadingString string
	loading       bool

	doesRequestInput bool
	textinput        textinput.Model
	needsNewInput    bool
	inputPrefix      string
	initialized      bool
}

func New(
	width, height int,
	filterPrefix, loadingString string,
	selectionEnabled, wrapText, requestInput bool,
) Model {
	pageFilter := filter.New(filterPrefix)
	pageViewport := viewport.New(width, height-pageFilter.ViewHeight())
	pageViewport.SetSelectionEnabled(selectionEnabled)
	pageViewport.SetWrapText(wrapText)

	needsNewInput := false
	var pageTextInput textinput.Model
	if requestInput {
		pageTextInput = textinput.New()
		pageTextInput.Focus()
		pageTextInput.Prompt = ""
		pageTextInput.SetValue("/bin/sh")
		needsNewInput = true
	}

	model := Model{
		width:            width,
		height:           height,
		viewport:         pageViewport,
		filter:           pageFilter,
		loadingString:    loadingString,
		loading:          true,
		doesRequestInput: requestInput,
		textinput:        pageTextInput,
		needsNewInput:    needsNewInput,
	}
	return model
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	dev.Debug(fmt.Sprintf("page %T", msg))
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	if m.EnteringInput() {
		if !m.initialized {
			m.initialized = true
			return m, textinput.Blink
		} else {
			switch msg := msg.(type) {
			case tea.KeyMsg:
				if msg.String() == "enter" && len(m.textinput.Value()) > 0 {
					m.needsNewInput = false
					return m, func() tea.Msg { return message.PageInputReceivedMsg{Input: m.textinput.Value()} }
				}
			}

			m.textinput, cmd = m.textinput.Update(msg)
			return m, cmd
		}
	}

	if m.viewport.Saving() {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case viewport.SaveStatusMsg:
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)

	case toast.TimeoutMsg:
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keymap.KeyMap.Back):
			m.clearFilter()

		case key.Matches(msg, keymap.KeyMap.Wrap):
			m.viewport.ToggleWrapText()
		}

		if m.filter.Focused() {
			switch {
			case key.Matches(msg, keymap.KeyMap.Forward):
				m.filter.Blur()
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
	var content string
	if m.loading {
		content = fmt.Sprintf(m.loadingString)
	} else {
		if m.EnteringInput() {
			content = m.inputPrefix + m.textinput.View()
		} else {
			content = m.viewport.View()
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left, m.filter.View(), content)
}

func (m *Model) SetWindowSize(width, height int) {
	m.width, m.height = width, height
	m.viewport.SetSize(width, height-m.filter.ViewHeight())
}

func (m *Model) SetHeader(header []string) {
	m.viewport.SetHeader(header)
}

func (m *Model) SetInputPrefix(p string) {
	m.inputPrefix = p
}

func (m *Model) SetViewportStyle(headerStyle, contentStyle lipgloss.Style) {
	m.viewport.HeaderStyle = headerStyle
	m.viewport.ContentStyle = contentStyle
}

func (m *Model) SetLoading(isLoading bool) {
	m.loading = isLoading
}

func (m *Model) SetAllPageData(allPageData []Row) {
	m.pageData.All = allPageData
	m.updateViewport()
}

func (m *Model) SetFilterPrefix(prefix string) {
	m.filter.SetPrefix(prefix)
}

func (m *Model) SetViewportSelectionToBottom() {
	m.viewport.SetSelectedContentIdx(len(m.pageData.Filtered) - 1)
}

func (m *Model) ScrollViewportToBottom() {
	m.viewport.ScrollToBottom()
}

func (m *Model) SetViewportXOffset(n int) {
	m.viewport.SetXOffset(n)
}

func (m *Model) HideToast() {
	m.viewport.HideToast()
}

func (m *Model) AppendToViewport(rows []Row, startOnNewLine bool) {
	newPageData := m.pageData.All
	for i, r := range rows {
		if r.Row != "" {
			if i == 0 && !startOnNewLine {
				allButLastEntry := newPageData[:max(0, len(newPageData)-1)]
				currentLastEntry := Row{}
				if len(m.pageData.All) > 0 {
					currentLastEntry = m.pageData.All[len(m.pageData.All)-1]
				}
				newLastEntry := Row{Key: currentLastEntry.Key, Row: currentLastEntry.Row + r.Row}
				newPageData = append(allButLastEntry, newLastEntry)
			} else {
				newPageData = append(newPageData, r)
			}
		}
	}
	m.SetAllPageData(newPageData)
	// maxHeight := m.height - m.filter.ViewHeight()
	// if shownPageHeight := len(m.pageData.All); shownPageHeight < maxHeight {
	// 	m.viewport.SetSize(m.width, shownPageHeight)
	// } else {
	// 	m.viewport.SetSize(m.width, maxHeight)
	// }
	m.viewport.ScrollToBottom()
}

func (m *Model) SetDoesNeedNewInput() {
	if !m.doesRequestInput {
		return
	}
	m.initialized = false
	m.needsNewInput = true
}

func (m *Model) SetViewportPromptVisible(v bool) {
	m.viewport.SetShowPrompt(v)
}

func (m Model) Loading() bool {
	return m.loading
}

func (m Model) GetSelectedPageRow() (Row, error) {
	selectedRow := m.viewport.SelectedContentIdx()
	if filtered := m.pageData.Filtered; len(filtered) > 0 && selectedRow >= 0 && selectedRow < len(filtered) {
		return filtered[selectedRow], nil
	}
	return Row{}, fmt.Errorf("bad thing")
}

func (m Model) EnteringInput() bool {
	return m.doesRequestInput && m.needsNewInput
}

func (m Model) FilterFocused() bool {
	return m.filter.Focused()
}

func (m Model) FilterApplied() bool {
	return len(m.filter.Filter) > 0
}

func (m Model) ViewportSaving() bool {
	return m.viewport.Saving()
}

func (m Model) ViewportHeight() int {
	return lipgloss.Height(m.viewport.View())
}

func (m *Model) clearFilter() {
	m.filter.BlurAndClear()
	m.updateViewport()
}

func (m *Model) updateViewport() {
	m.viewport.SetStringToHighlight(m.filter.Filter)
	m.updateFilteredData()
	m.viewport.SetContent(rowsToStrings(m.pageData.Filtered))
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
