package viewport

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
	content        []string
	wrappedContent []string
	// wrappedOffsets maps an index of content to the number of terminal rows it takes up when wrapped
	wrappedOffsets    map[int]int
	cursorEnabled     bool
	cursorRow         int
	stringToHighlight string
	wrapText          bool

	// width is the width of the entire viewport in terminal columns
	width int
	// height is the height of the entire viewport in terminal rows
	height int
	// contentHeight is the height of the viewport in terminal rows, excluding the terminal rows taken up by the header
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

	HeaderStyle    lipgloss.Style
	CursorRowStyle lipgloss.Style
	HighlightStyle lipgloss.Style
	ContentStyle   lipgloss.Style
	FooterStyle    lipgloss.Style
}

func New(width, height int) (m Model) {
	m.width, m.height = width, height

	m.saveDialog = textinput.New()
	m.saveDialog.Prompt = "> "
	m.saveDialog.Placeholder = m.getSaveDialogPlaceholder()
	m.saveDialog.PromptStyle = style.SaveDialogPromptStyle
	m.saveDialog.PlaceholderStyle = style.SaveDialogPlaceholderStyle
	m.saveDialog.TextStyle = style.SaveDialogTextStyle

	m.setContentHeight()
	m.keyMap = GetKeyMap()
	m.cursorEnabled = true
	m.wrapText = false

	m.HeaderStyle = style.ViewportHeaderStyle
	m.CursorRowStyle = style.ViewportCursorRowStyle
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
				if m.cursorEnabled {
					m.cursorRowUp(1)
				} else {
					m.viewUp(1)
				}

			case key.Matches(msg, m.keyMap.Down):
				if m.cursorEnabled {
					m.cursorRowDown(1)
				} else {
					m.viewDown(1)
				}

			case key.Matches(msg, m.keyMap.Left):
				m.viewLeft(m.width / 4)

			case key.Matches(msg, m.keyMap.Right):
				m.viewRight(m.width / 4)

			case key.Matches(msg, m.keyMap.HalfPageUp):
				m.viewUp(m.contentHeight / 2)
				if m.cursorEnabled {
					m.cursorRowUp(m.contentHeight / 2)
				}

			case key.Matches(msg, m.keyMap.HalfPageDown):
				m.viewDown(m.contentHeight / 2)
				if m.cursorEnabled {
					m.cursorRowDown(m.contentHeight / 2)
				}

			case key.Matches(msg, m.keyMap.PageUp):
				m.viewUp(m.contentHeight)
				if m.cursorEnabled {
					m.cursorRowUp(m.contentHeight)
				}

			case key.Matches(msg, m.keyMap.PageDown):
				m.viewDown(m.contentHeight)
				if m.cursorEnabled {
					m.cursorRowDown(m.contentHeight)
				}

			case key.Matches(msg, m.keyMap.Top):
				if m.cursorEnabled {
					m.cursorRowUp(m.yOffset + m.cursorRow)
				} else {
					m.viewUp(m.yOffset + m.cursorRow)
				}

			case key.Matches(msg, m.keyMap.Bottom):
				if m.cursorEnabled {
					m.cursorRowDown(m.maxContentIndex())
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

	nothingHighlighted := runeCount(m.stringToHighlight) == 0
	footerString, footerHeight := m.getFooter()
	lineCount := 0
	viewportHeightWithoutFooter := m.height - footerHeight

	addLineToViewString := func(line string) {
		if lineCount < viewportHeightWithoutFooter {
			viewString += line + "\n"
			lineCount += 1
		}
	}

	for _, headerLine := range m.header {
		for _, line := range m.lineToViewLines(headerLine) {
			addLineToViewString(m.HeaderStyle.Render(line))
		}
	}

	for idx, line := range m.visibleLines() {
		isSelected := m.cursorEnabled && m.yOffset+idx == m.cursorRow
		parsedLines := m.lineToViewLines(line)

		if nothingHighlighted {
			for _, line := range parsedLines {
				if isSelected {
					addLineToViewString(m.CursorRowStyle.Render(line))
				} else {
					addLineToViewString(m.ContentStyle.Render(line))
				}
			}
		} else {
			// this splitting and rejoining of styled content is expensive and causes increased flickering,
			// so only do it if something is actually highlighted
			styledHighlight := m.HighlightStyle.Render(m.stringToHighlight)
			lineStyle := m.ContentStyle
			if isSelected {
				lineStyle = m.CursorRowStyle
			}
			for _, line := range parsedLines {
				lineChunks := strings.Split(line, m.stringToHighlight)
				var styledChunks []string
				for _, chunk := range lineChunks {
					styledChunks = append(styledChunks, lineStyle.Render(chunk))
				}
				addLineToViewString(strings.Join(styledChunks, styledHighlight))
			}
		}
	}

	if footerHeight > 0 {
		// pad so footer shows up at bottom
		for lineCount < viewportHeightWithoutFooter {
			viewString += "\n"
			lineCount += 1
		}
		viewString += footerString
	}
	trimmedViewLines := strings.Trim(viewString, "\n")
	renderedViewString := style.Viewport.Width(m.width).Height(m.height).Render(trimmedViewLines)

	if m.toast.Visible {
		lines := strings.Split(renderedViewString, "\n")
		lines = lines[:len(lines)-m.toast.ViewHeight()]
		renderedViewString = strings.Join(lines, "\n") + "\n" + m.toast.View()
	}

	return renderedViewString
}

func (m *Model) SetCursorEnabled(cursorEnabled bool) {
	m.cursorEnabled = cursorEnabled
}

func (m *Model) SetWrapText(wrapText bool) {
	// idea for wrapping: model internally maintains wrappedHeader, wrappedContent []wrapped
	// where type wrapped struct { unwrappedIdx int, value string }
	// unwrappedIdx represents cursorRow when wrapped
	m.wrapText = wrapText
	m.fixState()
}

func (m *Model) ToggleWrapText() {
	m.wrapText = !m.wrapText
	m.fixState()
}

func (m *Model) HideToast() {
	m.toast.Visible = false
}

// SetSize sets the viewport's width and height, including header.
func (m *Model) SetSize(width, height int) {
	m.setWidthAndHeight(width, height)
	m.setContentHeight()
	m.fixState()
}

func (m *Model) SetHeader(header []string) {
	m.header = header
	m.updateMaxLineLength()
	m.fixState()
}

func (m *Model) SetContent(content []string) {
	m.content = content
	m.updateWrappedContent()
	m.updateMaxLineLength()
	m.setContentHeight()
	m.fixState()
}

// SetCursorRow sets the cursorRow with bounds. Adjusts yOffset as necessary.
func (m *Model) SetCursorRow(n int) {
	if m.contentHeight == 0 {
		return
	}

	if maxSelection := m.maxCursorRow(); n > maxSelection {
		m.cursorRow = maxSelection
	} else {
		m.cursorRow = max(0, n)
	}

	if lastVisibleLineIdx := m.lastVisibleLineIdx(); m.cursorRow > lastVisibleLineIdx {
		m.viewDown(m.cursorRow - lastVisibleLineIdx)
	} else if m.cursorRow < m.yOffset {
		m.viewUp(m.yOffset - m.cursorRow)
	}
}

func (m *Model) SetXOffset(n int) {
	maxXOffset := m.maxLineLength - m.width
	m.xOffset = max(0, min(maxXOffset, n))
}

func (m *Model) SetStringToHighlight(h string) {
	m.stringToHighlight = h
}

func (m Model) CursorRow() int {
	return m.cursorRow
}

func (m Model) Saving() bool {
	return m.saveDialog.Focused()
}

func (m *Model) updateWrappedContent() {
	var allWrappedContent []string
	wrappedOffsets := make(map[int]int)
	for i, line := range m.currentContent() {
		wrappedLinesForLine := m.getWrappedLines(line)
		wrappedOffsets[i] = len(wrappedLinesForLine)
		for _, wrappedLine := range wrappedLinesForLine {
			allWrappedContent = append(allWrappedContent, wrappedLine)
		}
	}
	m.wrappedContent = allWrappedContent
	m.wrappedOffsets = wrappedOffsets
}

func (m *Model) updateMaxLineLength() {
	for _, line := range append(m.header, m.currentContent()...) {
		if lineLength := runeCount(strings.TrimRight(line, " ")); lineLength > m.maxLineLength {
			m.maxLineLength = lineLength
		}
	}
}

func (m *Model) setWidthAndHeight(width, height int) {
	m.width, m.height = width, height
	m.saveDialog.Placeholder = m.getSaveDialogPlaceholder()
}

// fixCursorRow adjusts the cursor to be in a visible location if it is outside the visible content
func (m *Model) fixCursorRow() {
	if m.cursorRow > m.lastVisibleLineIdx() {
		m.SetCursorRow(m.lastVisibleLineIdx())
	}
}

// fixYOffset adjusts the yOffset such that it's not above the maximum value
func (m *Model) fixYOffset() {
	if maxYOffset := m.maxYOffset(); m.yOffset > maxYOffset {
		m.setYOffset(maxYOffset)
	}
}

// fixState fixes cursorRow and yOffset
func (m *Model) fixState() {
	m.fixYOffset()
	m.fixCursorRow()
}

func (m *Model) setContentHeight() {
	_, footerHeight := m.getFooter()
	contentHeight := m.height - len(m.header) - footerHeight
	m.contentHeight = max(0, contentHeight)
}

// maxContentIndex returns the maximum index of the model's content
func (m *Model) maxContentIndex() int {
	return len(m.currentContent()) - 1
}

// setYOffset sets the yOffset with bounds.
func (m *Model) setYOffset(n int) {
	if maxYOffset := m.maxYOffset(); n > maxYOffset {
		m.yOffset = maxYOffset
	} else {
		m.yOffset = max(0, n)
	}
}

// cursorRowDown moves the cursorRow down by the given number of content.
func (m *Model) cursorRowDown(n int) {
	m.SetCursorRow(m.cursorRow + n)
}

// cursorRowUp moves the cursorRow up by the given number of content.
func (m *Model) cursorRowUp(n int) {
	m.SetCursorRow(m.cursorRow - n)
}

// viewDown moves the view down by the given number of content.
func (m *Model) viewDown(n int) {
	m.setYOffset(m.yOffset + n)
}

// viewUp moves the view up by the given number of content.
func (m *Model) viewUp(n int) {
	m.setYOffset(m.yOffset - n)
}

// viewLeft moves the view left the given number of columns.
func (m *Model) viewLeft(n int) {
	m.SetXOffset(m.xOffset - n)
}

// viewRight moves the view right the given number of columns.
func (m *Model) viewRight(n int) {
	m.SetXOffset(m.xOffset + n)
}

func (m Model) currentContent() []string {
	if m.wrapText {
		return m.wrappedContent
	}
	return m.content
}

func (m Model) getSaveDialogPlaceholder() string {
	padding := m.width - runeCount(constants.SaveDialogPlaceholder) - runeCount(m.saveDialog.Prompt)
	padding = max(0, padding)
	placeholder := constants.SaveDialogPlaceholder + strings.Repeat(" ", padding)
	return placeholder[:min(runeCount(placeholder), m.width)]
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
	return len(m.currentContent()) - m.contentHeight
}

// maxCursorRow returns the maximum cursorRow
func (m Model) maxCursorRow() int {
	return len(m.currentContent()) - 1
}

// visibleLines retrieves the visible content based on the yOffset
func (m Model) visibleLines() []string {
	start := min(len(m.currentContent()), m.yOffset)
	end := start + m.contentHeight
	if end > m.maxContentIndex() {
		return m.currentContent()[start:]
	}
	return m.currentContent()[start:end]
}

func (m Model) getVisiblePartOfLine(line string) string {
	rightTrimmedLineLength := runeCount(strings.TrimRight(line, " "))
	end := min(runeCount(line), m.xOffset+m.width)
	start := min(end, m.xOffset)
	line = line[start:end]
	if m.xOffset+m.width < rightTrimmedLineLength {
		line = line[:runeCount(line)-lenLineContinuationIndicator] + lineContinuationIndicator
	}
	if m.xOffset > 0 {
		line = lineContinuationIndicator + line[min(runeCount(line), lenLineContinuationIndicator):]
	}
	return line
}

func (m Model) getWrappedLines(line string) []string {
	if runeCount(line) < m.width {
		return []string{line}
	}

	line = strings.TrimSpace(line)

	var wrappedLines []string
	for {
		lineWidth := runeCount(line)
		if lineWidth == 0 {
			break
		}

		width := m.width
		if lineWidth < m.width {
			width = lineWidth
		}

		wrappedLines = append(wrappedLines, line[0:width])
		line = line[width:]
	}

	return wrappedLines
}

func (m Model) lineToViewLines(line string) []string {
	if m.wrapText {
		return m.getWrappedLines(line)
	} else {
		return []string{m.getVisiblePartOfLine(line)}
	}
}

func (m Model) getFooter() (string, int) {
	numerator := m.cursorRow + 1

	if m.saveDialog.Focused() {
		footer := lipgloss.NewStyle().MaxWidth(m.width).Render(m.saveDialog.View())
		return footer, lipgloss.Height(footer)
	}

	// if cursor is disabled, percentage should show from the bottom of the visible content
	// such that panning the view to the bottom shows 100%
	if !m.cursorEnabled {
		numerator = m.yOffset + m.contentHeight
	}

	if numLines := len(m.currentContent()); numLines > m.height-len(m.header) {
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
	for _, line := range append(m.header, m.currentContent()...) {
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

func runeCount(a string) int {
	return utf8.RuneCountInString(a)
}

// func normalizeLineEndings(s string) string {
// 	return strings.ReplaceAll(s, "\r\n", "\n")
// }
