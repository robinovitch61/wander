package viewport

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinovitch61/wander/internal/dev"
	"github.com/robinovitch61/wander/internal/fileio"
	"github.com/robinovitch61/wander/internal/tui/components/toast"
	"github.com/robinovitch61/wander/internal/tui/constants"
	"github.com/robinovitch61/wander/internal/tui/style"
	"strings"
	"unicode/utf8"
)

const lineContinuationIndicator = "..."

var lenLineContinuationIndicator = runeCount(lineContinuationIndicator)

type SaveStatusMsg struct {
	SuccessMessage, Err string
}

type Model struct {
	header         []string
	wrappedHeader  []string
	content        []string
	wrappedContent []string

	// wrappedContentIdxToContentIdx maps the item at an index of wrappedContent to the index of content it is associated with (many wrappedContent indexes -> one content index)
	wrappedContentIdxToContentIdx map[int]int

	// contentIdxToFirstWrappedContentIdx maps the item at an index of content to the first index of wrappedContent it is associated with (index of content -> first index of wrappedContent)
	contentIdxToFirstWrappedContentIdx map[int]int

	// contentIdxToHeight maps the item at an index of content to its wrapped height in terminal rows
	contentIdxToHeight map[int]int

	// selectedContentIdx is the index of content of the currently selected item when selectionEnabled is true
	selectedContentIdx int
	stringToHighlight  string
	selectionEnabled   bool
	wrapText           bool

	// width is the width of the entire viewport in terminal columns
	width int
	// height is the height of the entire viewport in terminal rows
	height int
	// contentHeight is the height of the viewport in terminal rows, excluding the header and footer
	contentHeight int
	// maxLineLength is the maximum line length in terminal characters across header and content
	maxLineLength int

	keyMap viewportKeyMap

	// yOffset is the index of the first row shown on screen - wrappedContent[yOffset] if wrapText, otherwise content[yOffset]
	yOffset int
	// xOffset is the number of columns scrolled right when content lines overflow the viewport and wrapText is false
	xOffset int

	saveDialog textinput.Model
	toast      toast.Model

	HeaderStyle          lipgloss.Style
	SelectedContentStyle lipgloss.Style
	HighlightStyle       lipgloss.Style
	ContentStyle         lipgloss.Style
	FooterStyle          lipgloss.Style
}

func New(width, height int) (m Model) {
	m.saveDialog = textinput.New()
	m.saveDialog.Prompt = "> "
	m.saveDialog.PromptStyle = style.SaveDialogPromptStyle
	m.saveDialog.PlaceholderStyle = style.SaveDialogPlaceholderStyle
	m.saveDialog.TextStyle = style.SaveDialogTextStyle

	m.setWidthAndHeight(width, height)

	m.updateContentHeight()
	m.keyMap = GetKeyMap()
	m.selectionEnabled = true
	m.wrapText = false

	m.HeaderStyle = style.ViewportHeaderStyle
	m.SelectedContentStyle = style.ViewportSelectedRowStyle
	m.HighlightStyle = style.ViewportHighlightStyle
	m.FooterStyle = style.ViewportFooterStyle
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	dev.Debug(fmt.Sprintf("viewport %T", msg))
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	if m.saveDialog.Focused() {
		m.saveDialog, cmd = m.saveDialog.Update(msg)
		cmds = append(cmds, cmd)

		switch msg := msg.(type) {
		case tea.KeyMsg:
			cancel := key.Matches(msg, m.keyMap.CancelSave)
			confirm := key.Matches(msg, m.keyMap.ConfirmSave)
			if cancel || confirm {
				if confirm {
					cmds = append(cmds, m.getSaveCommand())
				}

				m.saveDialog.Blur()
				m.saveDialog.Reset()
				return m, tea.Batch(cmds...)
			}
		}
	} else {
		switch msg := msg.(type) {
		case SaveStatusMsg:
			if msg.Err != "" {
				m.toast = toast.New(fmt.Sprintf("Error: %s", msg.Err))
				m.toast.MessageStyle = style.ErrorToast.Copy().Width(m.width)
			} else {
				m.toast = toast.New(msg.SuccessMessage)
				m.toast.MessageStyle = style.SuccessToast.Copy().Width(m.width)
			}

		case tea.KeyMsg:
			switch {
			case key.Matches(msg, m.keyMap.Up):
				if m.selectionEnabled {
					m.selectedContentIdxUp(1)
				} else {
					m.viewUp(1)
				}

			case key.Matches(msg, m.keyMap.Down):
				if m.selectionEnabled {
					m.selectedContentIdxDown(1)
				} else {
					m.viewDown(1)
				}

			case key.Matches(msg, m.keyMap.Left):
				if !m.wrapText {
					m.viewLeft(m.width / 4)
				}

			case key.Matches(msg, m.keyMap.Right):
				if !m.wrapText {
					m.viewRight(m.width / 4)
				}

			case key.Matches(msg, m.keyMap.HalfPageUp):
				offset := max(1, m.getNumVisibleItems()/2)
				m.viewUp(m.contentHeight / 2)
				if m.selectionEnabled {
					m.selectedContentIdxUp(offset)
				}

			case key.Matches(msg, m.keyMap.HalfPageDown):
				offset := max(1, m.getNumVisibleItems()/2)
				m.viewDown(m.contentHeight / 2)
				if m.selectionEnabled {
					m.selectedContentIdxDown(offset)
				}

			case key.Matches(msg, m.keyMap.PageUp):
				offset := m.getNumVisibleItems()
				m.viewUp(m.contentHeight)
				if m.selectionEnabled {
					m.selectedContentIdxUp(offset)
				}

			case key.Matches(msg, m.keyMap.PageDown):
				offset := m.getNumVisibleItems()
				m.viewDown(m.contentHeight)
				if m.selectionEnabled {
					m.selectedContentIdxDown(offset)
				}

			case key.Matches(msg, m.keyMap.Top):
				if m.selectionEnabled {
					m.selectedContentIdxUp(m.yOffset + m.contentHeight)
				} else {
					m.viewUp(m.yOffset + m.contentHeight)
				}

			case key.Matches(msg, m.keyMap.Bottom):
				if m.selectionEnabled {
					m.selectedContentIdxDown(m.maxVisibleLineIdx())
				} else {
					m.viewDown(m.maxVisibleLineIdx())
				}

			case key.Matches(msg, m.keyMap.Save):
				m.saveDialog.Focus()
				cmds = append(cmds, textinput.Blink)
			}
		}
	}

	m.toast, cmd = m.toast.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var viewString string

	footerString, footerHeight := m.getFooter()

	addLineToViewString := func(line string) {
		viewString += line + "\n"
	}

	header := m.getHeader()
	for _, headerLine := range header {
		headerViewLine := m.getVisiblePartOfLine(headerLine)
		addLineToViewString(m.HeaderStyle.Render(headerViewLine))
	}

	visibleLines := m.getVisibleLines()
	hasNoHighlight := runeCount(m.stringToHighlight) == 0
	for idx, line := range visibleLines {
		isSelected := m.selectionEnabled && m.getContentIdx(m.yOffset+idx) == m.selectedContentIdx
		lineStyle := m.ContentStyle
		if isSelected {
			lineStyle = m.SelectedContentStyle
		}
		contentViewLine := m.getVisiblePartOfLine(line)

		if hasNoHighlight {
			addLineToViewString(lineStyle.Render(contentViewLine))
		} else {
			// this splitting and rejoining of styled content is expensive and causes increased flickering,
			// so only do it if something is actually highlighted
			lineChunks := strings.Split(contentViewLine, m.stringToHighlight)
			var styledChunks []string
			for _, chunk := range lineChunks {
				styledChunks = append(styledChunks, lineStyle.Render(chunk))
			}
			addLineToViewString(strings.Join(styledChunks, m.HighlightStyle.Render(m.stringToHighlight)))
		}
	}

	if footerHeight > 0 {
		// pad so footer shows up at bottom
		padCount := max(0, m.contentHeight-len(visibleLines)-footerHeight)
		viewString += strings.Repeat("\n", padCount)
		viewString += footerString
	}
	renderedViewString := style.Viewport.Width(m.width).Height(m.height).Render(viewString)

	if m.toast.Visible {
		lines := strings.Split(renderedViewString, "\n")
		lines = lines[:len(lines)-m.toast.ViewHeight()]
		renderedViewString = strings.Join(lines, "\n") + "\n" + m.toast.View()
	}

	return renderedViewString
}

func (m *Model) SetSelectionEnabled(selectionEnabled bool) {
	m.selectionEnabled = selectionEnabled
}

func (m *Model) SetWrapText(wrapText bool) {
	m.wrapText = wrapText
	m.updateForWrapText()
}

func (m *Model) ToggleWrapText() {
	m.wrapText = !m.wrapText
	m.updateForWrapText()
}

func (m *Model) HideToast() {
	m.toast.Visible = false
}

// SetSize sets the viewport's width and height, including header.
func (m *Model) SetSize(width, height int) {
	m.setWidthAndHeight(width, height)
	m.updateContentHeight()
	m.fixViewForSelection()
}

func (m *Model) SetHeader(header []string) {
	m.header = header
	m.updateWrappedHeader()
	m.updateForHeaderAndContent()
}

func (m *Model) SetContent(content []string) {
	m.content = content
	m.updateWrappedContent()
	m.updateForHeaderAndContent()
	m.fixSelection()
}

// SetSelectedContentIdx sets the selectedContentIdx with bounds. Adjusts yOffset as necessary.
func (m *Model) SetSelectedContentIdx(n int) {
	if m.contentHeight == 0 {
		return
	}

	if maxSelectedIdx := m.maxContentIdx(); n > maxSelectedIdx {
		m.selectedContentIdx = maxSelectedIdx
	} else {
		m.selectedContentIdx = max(0, n)
	}

	m.fixViewForSelection()
}

func (m *Model) SetXOffset(n int) {
	maxXOffset := m.maxLineLength - m.width
	m.xOffset = max(0, min(maxXOffset, n))
}

func (m *Model) SetStringToHighlight(h string) {
	m.stringToHighlight = h
}

func (m *Model) ScrollToBottom() {
	m.selectedContentIdxDown(len(m.content))
	m.viewDown(len(m.content))
}

func (m Model) SelectedContentIdx() int {
	return m.selectedContentIdx
}

func (m Model) Saving() bool {
	return m.saveDialog.Focused()
}

func (m *Model) updateWrappedHeader() {
	var allWrappedHeader []string
	for _, line := range m.header {
		wrappedLinesForLine := m.getWrappedLines(line)
		for _, wrappedLine := range wrappedLinesForLine {
			allWrappedHeader = append(allWrappedHeader, wrappedLine)
		}
	}
	m.wrappedHeader = allWrappedHeader
}

func (m *Model) updateWrappedContent() {
	var allWrappedContent []string
	wrappedContentIdxToContentIdx := make(map[int]int)
	contentIdxToFirstWrappedContentIdx := make(map[int]int)
	contentIdxToHeight := make(map[int]int)

	var wrappedContentIdx int
	for contentIdx, line := range m.content {
		wrappedLinesForLine := m.getWrappedLines(line)
		contentIdxToHeight[contentIdx] = len(wrappedLinesForLine)
		for _, wrappedLine := range wrappedLinesForLine {
			allWrappedContent = append(allWrappedContent, wrappedLine)

			wrappedContentIdxToContentIdx[wrappedContentIdx] = contentIdx
			if _, exists := contentIdxToFirstWrappedContentIdx[contentIdx]; !exists {
				contentIdxToFirstWrappedContentIdx[contentIdx] = wrappedContentIdx
			}

			wrappedContentIdx += 1
		}
	}
	m.wrappedContent = allWrappedContent
	m.wrappedContentIdxToContentIdx = wrappedContentIdxToContentIdx
	m.contentIdxToFirstWrappedContentIdx = contentIdxToFirstWrappedContentIdx
	m.contentIdxToHeight = contentIdxToHeight
}

func (m *Model) updateForHeaderAndContent() {
	m.updateMaxLineLength()
	m.updateContentHeight()
	m.fixViewForSelection()
}

func (m *Model) updateForWrapText() {
	m.updateContentHeight()
	m.updateWrappedContent()
	m.updateMaxLineLength()
	m.SetXOffset(0)
	m.fixViewForSelection()
}

func (m *Model) updateMaxLineLength() {
	for _, line := range append(m.getHeader(), m.getContent()...) {
		if lineLength := runeCount(strings.TrimRight(line, " ")); lineLength > m.maxLineLength {
			m.maxLineLength = lineLength
		}
	}
}

func (m *Model) setWidthAndHeight(width, height int) {
	m.width, m.height = width, height
	m.updateWrappedHeader()
	m.updateWrappedContent()
	m.updateSaveDialogPlaceholder()
}

// fixViewForSelection adjusts the view given the current selection
func (m *Model) fixViewForSelection() {
	currentLineIdx := m.getCurrentLineIdx()
	lastVisibleLineIdx := m.lastVisibleLineIdx()
	offScreenRowCount := currentLineIdx - lastVisibleLineIdx
	if offScreenRowCount >= 0 || m.lastContentItemSelected() {
		heightOffset := m.contentIdxToHeight[m.selectedContentIdx] - 1
		if !m.wrapText {
			heightOffset = 0
		}
		m.viewDown(offScreenRowCount + heightOffset)
	} else if currentLineIdx < m.yOffset {
		m.viewUp(m.yOffset - currentLineIdx)
	}

	if maxYOffset := m.maxYOffset(); m.yOffset > maxYOffset {
		m.setYOffset(maxYOffset)
	}
}

func (m *Model) fixSelection() {
	if !m.selectionEnabled {
		return
	}
	if m.selectedContentIdx > m.maxContentIdx() {
		m.selectedContentIdx = 0
	}
}

func (m *Model) updateContentHeight() {
	_, footerHeight := m.getFooter()
	contentHeight := m.height - len(m.getHeader()) - footerHeight
	m.contentHeight = max(0, contentHeight)
}

// setYOffset sets the yOffset with bounds
func (m *Model) setYOffset(n int) {
	if maxYOffset := m.maxYOffset(); n > maxYOffset {
		m.yOffset = maxYOffset
	} else {
		m.yOffset = max(0, n)
	}
}

// selectedContentIdxDown moves the selectedContentIdx down by the given number of terminal rows
func (m *Model) selectedContentIdxDown(n int) {
	m.SetSelectedContentIdx(m.selectedContentIdx + n)
}

// selectedContentIdxUp moves the selectedContentIdx up by the given number of terminal rows
func (m *Model) selectedContentIdxUp(n int) {
	m.SetSelectedContentIdx(m.selectedContentIdx - n)
}

// viewDown moves the view down by the given number of terminal rows
func (m *Model) viewDown(n int) {
	m.setYOffset(m.yOffset + n)
}

// viewUp moves the view up by the given number of terminal rows
func (m *Model) viewUp(n int) {
	m.setYOffset(m.yOffset - n)
}

// viewLeft moves the view left the given number of columns
func (m *Model) viewLeft(n int) {
	m.SetXOffset(m.xOffset - n)
}

// viewRight moves the view right the given number of columns
func (m *Model) viewRight(n int) {
	m.SetXOffset(m.xOffset + n)
}

func (m *Model) updateSaveDialogPlaceholder() {
	padding := m.width - runeCount(constants.SaveDialogPlaceholder) - runeCount(m.saveDialog.Prompt)
	padding = max(0, padding)
	placeholder := constants.SaveDialogPlaceholder + strings.Repeat(" ", padding)
	m.saveDialog.Placeholder = placeholder[:min(runeCount(placeholder), m.width)]
}

func (m Model) getHeader() []string {
	if m.wrapText {
		return m.wrappedHeader
	}
	return m.header
}

func (m Model) getContent() []string {
	if m.wrapText {
		return m.wrappedContent
	}
	return m.content
}

// lastVisibleLineIdx returns the maximum visible line index
func (m Model) lastVisibleLineIdx() int {
	return min(m.maxVisibleLineIdx(), m.yOffset+m.contentHeight-1)
}

// maxYOffset returns the maximum yOffset (the yOffset that shows the final screen)
func (m Model) maxYOffset() int {
	if m.maxVisibleLineIdx() < m.contentHeight {
		return 0
	}
	return len(m.getContent()) - m.contentHeight
}

func (m *Model) maxVisibleLineIdx() int {
	return len(m.getContent()) - 1
}

func (m Model) maxContentIdx() int {
	return len(m.content) - 1
}

// getVisibleLines retrieves the visible content based on the yOffset and contentHeight
func (m Model) getVisibleLines() []string {
	maxVisibleLineIdx := m.maxVisibleLineIdx()
	start := max(0, min(maxVisibleLineIdx, m.yOffset))
	end := start + m.contentHeight
	if end > maxVisibleLineIdx {
		return m.getContent()[start:]
	}
	return m.getContent()[start:end]
}

func (m Model) getVisiblePartOfLine(line string) string {
	rightTrimmedLineLength := runeCount(strings.TrimRight(line, " "))
	end := min(runeCount(line), m.xOffset+m.width)
	start := min(end, m.xOffset)
	line = line[start:end]
	if m.xOffset+m.width < rightTrimmedLineLength {
		truncate := max(0, runeCount(line)-lenLineContinuationIndicator)
		line = line[:truncate] + lineContinuationIndicator
	}
	if m.xOffset > 0 {
		line = lineContinuationIndicator + line[min(runeCount(line), lenLineContinuationIndicator):]
	}
	return line
}

func (m Model) getContentIdx(wrappedContentIdx int) int {
	if !m.wrapText {
		return wrappedContentIdx
	}
	return m.wrappedContentIdxToContentIdx[wrappedContentIdx]
}

func (m Model) getCurrentLineIdx() int {
	if m.wrapText {
		return m.contentIdxToFirstWrappedContentIdx[m.selectedContentIdx]
	}
	return m.selectedContentIdx
}

func (m Model) getWrappedLines(line string) []string {
	if runeCount(line) < m.width {
		return []string{line}
	}
	line = strings.TrimRight(line, " ")
	return splitLineIntoSizedChunks(line, m.width)
}

func (m Model) getNumVisibleItems() int {
	if !m.wrapText {
		return m.contentHeight
	}

	var itemCount int
	var rowCount int
	contentIdx := m.wrappedContentIdxToContentIdx[m.yOffset]
	for rowCount < m.contentHeight {
		if height, exists := m.contentIdxToHeight[contentIdx]; exists {
			rowCount += height
		} else {
			break
		}
		contentIdx += 1
		itemCount += 1
	}
	return itemCount
}

func (m Model) lastContentItemSelected() bool {
	return m.selectedContentIdx == len(m.content)-1
}

func (m Model) getFooter() (string, int) {
	numerator := m.selectedContentIdx + 1
	denominator := len(m.content)
	totalNumLines := len(m.getContent())

	if m.saveDialog.Focused() {
		footer := lipgloss.NewStyle().MaxWidth(m.width).Render(m.saveDialog.View())
		return footer, lipgloss.Height(footer)
	}

	// if selection is disabled, percentage should show from the bottom of the visible content
	// such that panning the view to the bottom shows 100%
	if !m.selectionEnabled {
		numerator = m.yOffset + m.contentHeight
		denominator = totalNumLines
	}

	if totalNumLines >= m.height-len(m.getHeader()) {
		percentScrolled := percent(numerator, denominator)
		footerString := fmt.Sprintf("%d%% (%d/%d)", percentScrolled, numerator, denominator)
		renderedFooterString := m.FooterStyle.Copy().MaxWidth(m.width).Render(footerString)
		footerHeight := lipgloss.Height(renderedFooterString)
		return renderedFooterString, footerHeight
	}
	return "", 0
}

func (m Model) getSaveCommand() tea.Cmd {
	var content string
	for _, line := range append(m.getHeader(), m.getContent()...) {
		content += strings.TrimRight(line, " ") + "\n"
	}

	return func() tea.Msg {
		savePathWithFileName, err := fileio.SaveToFile(m.saveDialog.Value(), content)
		if err != nil {
			return SaveStatusMsg{Err: err.Error()}
		}
		return SaveStatusMsg{SuccessMessage: fmt.Sprintf("Success: saved to %s", savePathWithFileName)}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func percent(a, b int) int {
	return int(float32(a) / float32(b) * 100)
}

func splitLineIntoSizedChunks(line string, chunkSize int) []string {
	var wrappedLines []string
	for {
		lineWidth := runeCount(line)
		if lineWidth == 0 {
			break
		}

		width := chunkSize
		if lineWidth < chunkSize {
			width = lineWidth
		}

		wrappedLines = append(wrappedLines, line[0:width])
		line = line[width:]
	}
	return wrappedLines
}

func runeCount(a string) int {
	return utf8.RuneCountInString(a)
}

// func normalizeLineEndings(s string) string {
// 	return strings.ReplaceAll(s, "\r\n", "\n")
// }
