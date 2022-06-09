package viewport

// TODO LEO: Remove
// - terminal rows mean a literal row in the terminal
// - selection maps to an entry in content, even if that entry wraps to more than one terminal row

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
	"unicode/utf8"
	"wander/components/toast"
	"wander/constants"
	"wander/dev"
	"wander/fileio"
	"wander/style"
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

	// wrappedContentIdxToContentIdx maps indexes of wrappedContent to the indexes of content they are associated with
	wrappedContentIdxToContentIdx map[int]int
	selectionEnabled              bool

	// selectedContentIdx is the index of content of the currently selected item when selectionEnabled is true
	selectedContentIdx int
	stringToHighlight  string
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

	// yOffset indexes into content to the first shown line in the viewport. top line = content[yOffset]
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
				m.saveDialog.Blur()
				m.saveDialog.Reset()

				if confirm {
					cmds = append(cmds, m.getSaveCommand())
					// return m, tea.Batch(cmds...) // TODO LEO: Confirm ok
				}
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
				m.viewLeft(m.width / 4)

			case key.Matches(msg, m.keyMap.Right):
				m.viewRight(m.width / 4)

			case key.Matches(msg, m.keyMap.HalfPageUp):
				m.viewUp(m.contentHeight / 2)
				if m.selectionEnabled {
					m.selectedContentIdxUp(m.contentHeight / 2)
				}

			case key.Matches(msg, m.keyMap.HalfPageDown):
				m.viewDown(m.contentHeight / 2)
				if m.selectionEnabled {
					m.selectedContentIdxDown(m.contentHeight / 2)
				}

			case key.Matches(msg, m.keyMap.PageUp):
				m.viewUp(m.contentHeight)
				if m.selectionEnabled {
					m.selectedContentIdxUp(m.contentHeight)
				}

			case key.Matches(msg, m.keyMap.PageDown):
				m.viewDown(m.contentHeight)
				if m.selectionEnabled {
					m.selectedContentIdxDown(m.contentHeight)
				}

			case key.Matches(msg, m.keyMap.Top):
				if m.selectionEnabled {
					m.selectedContentIdxUp(m.yOffset + m.contentHeight)
				} else {
					m.viewUp(m.yOffset + m.contentHeight)
				}

			case key.Matches(msg, m.keyMap.Bottom):
				if m.selectionEnabled {
					m.selectedContentIdxDown(m.maxContentIndex())
				} else {
					m.viewDown(m.maxContentIndex())
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
	dev.Debug(fmt.Sprintf("contentHeight %d, footerHeight %d", m.contentHeight, footerHeight))

	addLineToViewString := func(line string) {
		viewString += line + "\n"
	}

	for _, headerLine := range m.getHeader() {
		headerViewLine := m.getVisiblePartOfLine(headerLine)
		addLineToViewString(m.HeaderStyle.Render(headerViewLine))
	}

	visibleLines := m.getVisibleLines()
	dev.Debug(fmt.Sprintf("LEN VISIBLE %d", len(visibleLines)))
	for idx, line := range visibleLines {
		isSelected := m.selectionEnabled && m.getContentIdx(m.yOffset+idx) == m.selectedContentIdx
		lineStyle := m.ContentStyle
		if isSelected {
			lineStyle = m.SelectedContentStyle
		}
		contentViewLine := m.getVisiblePartOfLine(line)

		if runeCount(m.stringToHighlight) == 0 {
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
		padCount := max(0, m.contentHeight-len(visibleLines)-len(m.getHeader()))
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
	// idea for wrapping: model internally maintains wrappedHeader, wrappedContent []wrapped
	// where type wrapped struct { unwrappedIdx int, value string }
	// unwrappedIdx represents selectedContentIdx when wrapped
	m.wrapText = wrapText
	m.updateContentHeight()
	m.fixState()
}

func (m *Model) ToggleWrapText() {
	m.wrapText = !m.wrapText
	m.updateContentHeight()
	m.fixState()
}

func (m *Model) HideToast() {
	m.toast.Visible = false
}

// SetSize sets the viewport's width and height, including header.
func (m *Model) SetSize(width, height int) {
	m.setWidthAndHeight(width, height)
	m.updateContentHeight()
	m.fixState()
}

func (m *Model) SetHeader(header []string) {
	m.header = header
	m.updateWrappedHeader()
	// TODO LEO: dedupe these three lines
	m.updateMaxLineLength()
	m.updateContentHeight()
	m.fixState()
}

func (m *Model) SetContent(content []string) {
	m.content = content
	m.updateWrappedContent()
	m.updateMaxLineLength()
	m.updateContentHeight()
	m.fixState()
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

	if lastVisibleLineIdx := m.lastVisibleLineIdx(); m.selectedContentIdx > lastVisibleLineIdx {
		m.viewDown(m.selectedContentIdx - lastVisibleLineIdx)
	} else if m.selectedContentIdx < m.yOffset {
		m.viewUp(m.yOffset - m.selectedContentIdx)
	}
}

func (m *Model) SetXOffset(n int) {
	maxXOffset := m.maxLineLength - m.width
	m.xOffset = max(0, min(maxXOffset, n))
}

func (m *Model) SetStringToHighlight(h string) {
	m.stringToHighlight = h
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
	var wrappedContentIdx int
	for contentIdx, line := range m.content {
		wrappedLinesForLine := m.getWrappedLines(line)
		for _, wrappedLine := range wrappedLinesForLine {
			allWrappedContent = append(allWrappedContent, wrappedLine)
			wrappedContentIdxToContentIdx[wrappedContentIdx] = contentIdx
			wrappedContentIdx += 1
		}
	}
	m.wrappedContent = allWrappedContent
	m.wrappedContentIdxToContentIdx = wrappedContentIdxToContentIdx
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

// fixSelectedContentIdx adjusts the selection to be in a visible location if it is outside the visible content
func (m *Model) fixSelectedContentIdx() {
	if m.selectedContentIdx > m.lastVisibleLineIdx() {
		m.SetSelectedContentIdx(m.lastVisibleLineIdx())
	}
}

// fixYOffset adjusts the yOffset such that it's not above the maximum value
func (m *Model) fixYOffset() {
	if maxYOffset := m.maxYOffset(); m.yOffset > maxYOffset {
		m.setYOffset(maxYOffset)
	}
}

// fixState fixes selectedContentIdx and yOffset
func (m *Model) fixState() {
	m.fixYOffset()
	m.fixSelectedContentIdx()
}

func (m *Model) updateContentHeight() {
	_, footerHeight := m.getFooter()
	contentHeight := m.height - len(m.getHeader()) - footerHeight
	m.contentHeight = max(0, contentHeight)
}

// maxContentIndex returns the maximum index of the model's content
func (m *Model) maxContentIndex() int {
	return len(m.getContent()) - 1
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
	return min(m.maxContentIndex(), m.yOffset+m.contentHeight-1)
}

// maxYOffset returns the maximum yOffset (the yOffset that shows the final screen)
func (m Model) maxYOffset() int {
	if m.maxContentIndex() < m.contentHeight {
		return 0
	}
	return len(m.getContent()) - m.contentHeight
}

// maxContentIdx returns the maximum content index
func (m Model) maxContentIdx() int {
	return len(m.getContent()) - 1
}

// getVisibleLines retrieves the visible content based on the yOffset and contentHeight
func (m Model) getVisibleLines() []string {
	start := max(0, min(m.maxContentIdx(), m.yOffset))
	end := start + m.contentHeight
	if end > m.maxContentIndex() {
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

func (m Model) getWrappedLines(line string) []string {
	if runeCount(line) < m.width {
		return []string{line}
	}
	line = strings.TrimSpace(line)
	return splitLineIntoSizedChunks(line, m.width)
}

func (m Model) getFooter() (string, int) {
	numerator := m.selectedContentIdx + 1

	if m.saveDialog.Focused() {
		footer := lipgloss.NewStyle().MaxWidth(m.width).Render(m.saveDialog.View())
		return footer, lipgloss.Height(footer)
	}

	// if selection is disabled, percentage should show from the bottom of the visible content
	// such that panning the view to the bottom shows 100%
	if !m.selectionEnabled {
		numerator = m.yOffset + m.contentHeight
	}

	if numLines := len(m.getContent()); numLines > m.height-len(m.getHeader()) {
		percentScrolled := percent(numerator, numLines)
		footerString := fmt.Sprintf("%d%% (%d/%d)", percentScrolled, numerator, numLines)
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
