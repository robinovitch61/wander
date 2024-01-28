package page

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinovitch61/wander/internal/dev"
	"github.com/robinovitch61/wander/internal/tui/components/filter"
	"github.com/robinovitch61/wander/internal/tui/components/toast"
	"github.com/robinovitch61/wander/internal/tui/components/viewport"
	"github.com/robinovitch61/wander/internal/tui/constants"
	"github.com/robinovitch61/wander/internal/tui/keymap"
	"github.com/robinovitch61/wander/internal/tui/message"
)

type Config struct {
	Width, Height                            int
	LoadingString                            string
	SelectionEnabled, WrapText, RequestInput bool
	CompactTableContent                      bool
	ViewportConditionalStyle                 map[string]lipgloss.Style
}

type Model struct {
	width, height int

	pageData data

	viewport viewport.Model
	filter   filter.Model

	loadingString string
	loading       bool

	copySavePath bool

	doesRequestInput bool
	textinput        textinput.Model
	inputPrefix      string
	initialized      bool

	// if FilterWithContext is true, filtering doesn't remove rows, just highlights the matching text
	// and makes it so you can cycle through matches
	FilterWithContext bool
}

func New(c Config, copySavePath, startFiltering, filterWithContext bool) Model {
	pageFilter := filter.New("")
	if startFiltering {
		pageFilter.Focus()
	}
	pageViewport := viewport.New(c.Width, c.Height-pageFilter.ViewHeight(), c.CompactTableContent)
	pageViewport.SetSelectionEnabled(c.SelectionEnabled)
	pageViewport.SetWrapText(c.WrapText)
	pageViewport.ConditionalStyle = c.ViewportConditionalStyle

	var pageTextInput textinput.Model
	if c.RequestInput {
		pageTextInput = textinput.New()
		pageTextInput.Focus()
		pageTextInput.Prompt = ""
		pageTextInput.SetValue(constants.DefaultPageInput)
	}

	model := Model{
		width:             c.Width,
		height:            c.Height,
		viewport:          pageViewport,
		filter:            pageFilter,
		loadingString:     c.LoadingString,
		loading:           true,
		copySavePath:      copySavePath,
		doesRequestInput:  c.RequestInput,
		textinput:         pageTextInput,
		FilterWithContext: filterWithContext,
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
		if m.copySavePath {
			cmds = append(cmds, func() tea.Msg {
				_ = clipboard.WriteAll(msg.FullPath)
				return nil
			})
		}
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
				// done editing, apply filter
				m.filter.Blur()
				m.updateFilter()
			}
		} else {
			// not focused and hit filter - start filtering
			if key.Matches(msg, keymap.KeyMap.Filter) {
				m.filter.Focus()
				return m, textinput.Blink
			}

			// not editing filter - pass through to viewport
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}

		prevFilter := m.filter.Value()
		m.filter, cmd = m.filter.Update(msg)
		if m.filter.Value() != prevFilter {
			m.updateViewport()
		}
		cmds = append(cmds, cmd)

		// after all the filtered data is updated and commands are prepped, we can deal with contextual filtering
		if m.FilterWithContext {
			if m.filter.Focused() {
				// every keystroke modifies what filtered rows match
				m.ResetContextFilter()
			} else {
				if m.FilterWithContext && m.filter.HasFilterText() && len(m.pageData.FilteredContentIdxs) > 0 {
					switch {
					case key.Matches(msg, keymap.KeyMap.NextFilteredRow):
						m.pageData.IncrementFilteredSelectionNum()
						m.scrollViewportToContentIdx(m.pageData.CurrentFilteredContentIdx)
					case key.Matches(msg, keymap.KeyMap.PrevFilteredRow):
						m.pageData.DecrementFilteredSelectionNum()
						m.scrollViewportToContentIdx(m.pageData.CurrentFilteredContentIdx)
					}
				}
			}
		}

	default:
		m.filter, cmd = m.filter.Update(msg)
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

func (m *Model) SetAllPageRows(allPageRows []Row) {
	m.pageData.AllRows = allPageRows
	m.updateViewport()
}

func (m *Model) SetFilterPrefix(prefix string) {
	m.filter.SetPrefix(prefix)
}

func (m *Model) SetViewportSelectionToBottom() {
	m.viewport.SetSelectedContentIdx(len(m.pageData.FilteredRows) - 1)
}

func (m *Model) ScrollViewportToBottom() {
	m.viewport.ScrollToBottom()
}

func (m *Model) SetViewportXOffset(n int) {
	m.viewport.SetXOffset(n)
}

func (m *Model) SetToast(toast toast.Model, style lipgloss.Style) {
	m.viewport.SetToast(toast, style)
}

func (m *Model) HideToast() {
	m.viewport.HideToast()
}

func (m *Model) AppendToViewport(rows []Row, startOnNewLine bool) {
	newPageRows := m.pageData.AllRows
	for i, r := range rows {
		if r.Row != "" {
			if i == 0 && !startOnNewLine {
				allButLastEntry := newPageRows[:max(0, len(newPageRows)-1)]
				currentLastEntry := Row{}
				if len(m.pageData.AllRows) > 0 {
					currentLastEntry = m.pageData.AllRows[len(m.pageData.AllRows)-1]
				}
				newLastEntry := Row{Key: currentLastEntry.Key, Row: currentLastEntry.Row + r.Row}
				newPageRows = append(allButLastEntry, newLastEntry)
			} else {
				newPageRows = append(newPageRows, r)
			}
		}
	}
	m.SetAllPageRows(newPageRows)
}

func (m *Model) SetDoesNeedNewInput() {
	if !m.doesRequestInput {
		return
	}
	m.initialized = false
}

func (m *Model) SetViewportPromptVisible(v bool) {
	m.viewport.SetShowPrompt(v)
}

func (m *Model) SetViewportSelectionEnabled(v bool) {
	m.viewport.SetSelectionEnabled(v)
}

func (m *Model) setIndexesOfFilteredRows(idxs []int) {
	m.pageData.FilteredContentIdxs = idxs
	m.updateFilter()
}

func (m *Model) ToggleCompact() {
	m.filter.ToggleCompact()
}

func (m *Model) ResetContextFilter() {
	if !m.FilterWithContext {
		return
	}
	m.pageData.FilteredSelectionNum = 0
	if len(m.pageData.FilteredContentIdxs) > 0 {
		m.pageData.CurrentFilteredContentIdx = m.pageData.FilteredContentIdxs[m.pageData.FilteredSelectionNum]
		m.scrollViewportToContentIdx(m.pageData.CurrentFilteredContentIdx)
	}
}

func (m Model) Loading() bool {
	return m.loading
}

func (m Model) GetSelectedPageRow() (Row, error) {
	if !m.viewport.SelectionEnabled() {
		return Row{}, fmt.Errorf("selection disabled")
	}
	selectedRow := m.viewport.SelectedContentIdx()
	if filtered := m.pageData.FilteredRows; len(filtered) > 0 && selectedRow >= 0 && selectedRow < len(filtered) {
		return filtered[selectedRow], nil
	}
	return Row{}, fmt.Errorf("selection invalid")
}

func (m Model) ViewportSelectionAtBottom() bool {
	if !m.viewport.SelectionEnabled() {
		return false
	}
	if len(m.pageData.FilteredRows) == 0 {
		return true
	}
	return m.viewport.SelectedContentIdx() == len(m.pageData.FilteredRows)-1
}

func (m Model) EnteringInput() bool {
	return m.doesRequestInput
}

func (m Model) FilterFocused() bool {
	return m.filter.Focused()
}

func (m Model) FilterApplied() bool {
	return len(m.filter.Value()) > 0
}

func (m Model) ViewportSaving() bool {
	return m.viewport.Saving()
}

func (m Model) ViewportHeight() int {
	return lipgloss.Height(m.viewport.View())
}

func (m *Model) clearFilter() {
	m.filter.BlurAndClear()
	m.setIndexesOfFilteredRows([]int{})
	m.updateViewport()
}

func (m *Model) updateViewport() {
	m.viewport.SetStringToHighlight(m.filter.Value())
	m.updateFilteredData()
	m.viewport.SetContent(rowsToStrings(m.pageData.FilteredRows))
}

func (m *Model) updateFilteredData() {
	if !m.filter.HasFilterText() {
		m.pageData.FilteredRows = m.pageData.AllRows
		m.setIndexesOfFilteredRows([]int{})
	} else if m.FilterWithContext {
		m.pageData.FilteredRows = m.pageData.AllRows
		var indexesOfFilteredRows []int
		for i, entry := range m.pageData.AllRows {
			if strings.Contains(entry.Row, m.filter.Value()) {
				indexesOfFilteredRows = append(indexesOfFilteredRows, i)
			}
		}
		m.setIndexesOfFilteredRows(indexesOfFilteredRows)
	} else {
		var filteredData []Row
		for _, entry := range m.pageData.AllRows {
			if strings.Contains(entry.Row, m.filter.Value()) {
				filteredData = append(filteredData, entry)
			}
		}
		m.pageData.FilteredRows = filteredData
	}
}

func (m *Model) updateFilter() {
	if !m.FilterWithContext {
		return
	}

	if len(m.pageData.FilteredContentIdxs) == 0 {
		m.filter.SetSuffix(" (no matches) ")
	} else if m.filter.Focused() {
		m.filter.SetSuffix(
			fmt.Sprintf(
				" (%d/%d, %s to apply) ",
				m.pageData.FilteredSelectionNum+1,
				len(m.pageData.FilteredContentIdxs),
				keymap.KeyMap.Forward.Help().Key,
			),
		)
	} else {
		m.filter.SetSuffix(
			fmt.Sprintf(
				" (%d/%d, %s/%s to cycle) ",
				m.pageData.FilteredSelectionNum+1,
				len(m.pageData.FilteredContentIdxs),
				keymap.KeyMap.NextFilteredRow.Help().Key,
				keymap.KeyMap.PrevFilteredRow.Help().Key,
			),
		)
	}
}

func (m *Model) scrollViewportToContentIdx(contentIdx int) {
	m.viewport.SetSelectedContentIdx(contentIdx)
	m.viewport.SpecialContentIdx = m.pageData.CurrentFilteredContentIdx
	m.updateFilter()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
