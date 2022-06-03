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
const lenLineContinuationIndicator = len(lineContinuationIndicator)

type SaveStatusMsg struct {
	SuccessMessage, Err string
}

type Model struct {
	header            []string
	content           []string
	cursorEnabled     bool
	cursorRow         int
	StringToHighlight string
	wrapText          bool
	mouseWheelEnabled bool

	width         int
	height        int
	contentHeight int
	maxLineLength int

	keyMap     viewportKeyMap
	saveDialog textinput.Model

	yOffset int
	xOffset int

	toastMessage string
	showToast    bool

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
	m.mouseWheelEnabled = false

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
			switch {
			case key.Matches(msg, m.keyMap.CancelSave):
				m.saveDialog.Blur()
				m.saveDialog.Reset()

			case key.Matches(msg, m.keyMap.ConfirmSave):
				var content string
				for _, line := range append(m.header, m.content...) {
					content += strings.TrimRight(line, " ") + "\n"
				}
				cmds = append(cmds, saveCommand(m.saveDialog.Value(), content))
				m.saveDialog.Blur()
				m.saveDialog.Reset()
			}
		}
	} else {
		switch msg := msg.(type) {
		case toast.ToastTimeoutMsg:
			m.showToast = false
			return m, nil

		case SaveStatusMsg:
			if msg.Err != "" {
				m.toastMessage = style.ErrorToast.Width(m.width).Render(fmt.Sprintf("Error: %s", msg.Err))
			} else {
				m.toastMessage = style.SuccessToast.Width(m.width).Render(msg.SuccessMessage)
			}
			m.showToast = true
			cmds = append(cmds, toast.GetToastTimeoutCmd())
			return m, tea.Batch(cmds...)

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
					m.cursorRowDown(m.maxLinesIdx())
				} else {
					m.viewDown(m.maxLinesIdx())
				}

			case key.Matches(msg, m.keyMap.Save):
				m.saveDialog.Focus()
				cmds = append(cmds, textinput.Blink)
			}

		case tea.MouseMsg:
			if !m.mouseWheelEnabled {
				break
			}
			switch msg.Type {
			case tea.MouseWheelUp:
				m.cursorRowUp(1)

			case tea.MouseWheelDown:
				m.cursorRowDown(1)
			}
		}
	}
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var viewString string

	nothingHighlighted := len(m.StringToHighlight) == 0
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
			styledHighlight := m.HighlightStyle.Render(m.StringToHighlight)
			lineStyle := m.ContentStyle
			if isSelected {
				lineStyle = m.CursorRowStyle
			}
			for _, line := range parsedLines {
				lineChunks := strings.Split(line, m.StringToHighlight)
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

	if m.showToast {
		lines := strings.Split(renderedViewString, "\n")
		lines = lines[:len(lines)-lipgloss.Height(m.toastMessage)]
		renderedViewString = strings.Join(lines, "\n") + "\n" + m.toastMessage
	}

	return renderedViewString
}

func (m *Model) SetCursorEnabled(cursorEnabled bool) {
	m.cursorEnabled = cursorEnabled
}

func (m *Model) SetWrapText(wrapText bool) {
	// TODO LEO: currently can't wrap text with cursor enabled due to mismatch between contentHeight and what
	// View() actually returns

	// idea for wrapping: model internally maintains wrappedHeader, wrappedContent []wrapped
	// where type wrapped struct { unwrappedIdx int, value string }
	// unwrappedIdx represents cursorRow when wrapped
	if m.cursorEnabled {
		m.wrapText = false
	} else {
		m.wrapText = wrapText
	}
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
	m.xOffset = max(0, min(m.maxLineLength, n))
}

func (m Model) CursorRow() int {
	return m.cursorRow
}

func (m Model) Saving() bool {
	return m.saveDialog.Focused()
}

func (m *Model) updateMaxLineLength() {
	for _, line := range append(m.header, m.content...) {
		if lineLength := len(strings.TrimRight(line, " ")); lineLength > m.maxLineLength {
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

// maxLinesIdx returns the maximum index of the model's content
func (m *Model) maxLinesIdx() int {
	return len(m.content) - 1
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
	m.SetXOffset(min(m.maxLineLength-m.width, m.xOffset+n))
}

func (m Model) getSaveDialogPlaceholder() string {
	padding := m.width - utf8.RuneCountInString(constants.SaveDialogPlaceholder) - utf8.RuneCountInString(m.saveDialog.Prompt)
	padding = max(0, padding)
	placeholder := constants.SaveDialogPlaceholder + strings.Repeat(" ", padding)
	return placeholder[:min(len(placeholder), m.width)]
}

// lastVisibleLineIdx returns the maximum visible line index
func (m Model) lastVisibleLineIdx() int {
	return min(m.maxLinesIdx(), m.yOffset+m.contentHeight-1)
}

// maxYOffset returns the maximum yOffset (the yOffset that shows the final screen)
func (m Model) maxYOffset() int {
	if m.maxLinesIdx() < m.contentHeight {
		return 0
	}
	return len(m.content) - m.contentHeight
}

// maxCursorRow returns the maximum cursorRow
func (m Model) maxCursorRow() int {
	return len(m.content) - 1
}

// visibleLines retrieves the visible content based on the yOffset
func (m Model) visibleLines() []string {
	start := min(len(m.content), m.yOffset)
	end := start + m.contentHeight
	if end > m.maxLinesIdx() {
		return m.content[start:]
	}
	return m.content[start:end]
}

func (m Model) getVisiblePartOfLine(line string) string {
	rightTrimmedLineLength := len(strings.TrimRight(line, " "))
	end := min(len(line), m.xOffset+m.width)
	start := min(end, m.xOffset)
	line = line[start:end]
	if m.xOffset+m.width < rightTrimmedLineLength {
		line = line[:len(line)-lenLineContinuationIndicator] + lineContinuationIndicator
	}
	if m.xOffset > 0 {
		line = lineContinuationIndicator + line[min(len(line), lenLineContinuationIndicator):]
	}
	return line
}

func (m Model) getWrappedLines(line string) []string {
	if utf8.RuneCountInString(line) < m.width {
		return []string{line}
	}

	var lines []string
	l := ""
	for pos, b := range []rune(line) {
		l += string(b)
		if pos != 0 && (pos+1)%m.width == 0 {
			lines = append(lines, l)
			l = ""
		}
	}
	lines = append(lines, l)
	return lines
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
		return lipgloss.NewStyle().MaxWidth(m.width).Render(m.saveDialog.View()), 1
	}

	// if cursor is disabled, percentage should show from the bottom of the visible content
	// such that panning the view to the bottom shows 100%
	if !m.cursorEnabled {
		numerator = m.yOffset + len(m.visibleLines())
	}

	if numLines := len(m.content); numLines > m.height-len(m.header) {
		percentScrolled := percent(numerator, numLines)
		footerString := fmt.Sprintf("%d%% (%d/%d)", percentScrolled, numerator, numLines)
		renderedFooterString := m.FooterStyle.Copy().MaxWidth(m.width).Render(footerString)
		footerHeight := lipgloss.Height(renderedFooterString)
		return renderedFooterString, footerHeight
	}
	return "", 0
}

func saveCommand(saveDialogValue string, fileContent string) tea.Cmd {
	return func() tea.Msg {
		savePathWithFileName, err := fileio.SaveToFile(saveDialogValue, fileContent)
		if err != nil {
			return SaveStatusMsg{SuccessMessage: "", Err: err.Error()}
		}
		successMessage := fmt.Sprintf("Success: saved to %s", savePathWithFileName)
		return SaveStatusMsg{SuccessMessage: successMessage, Err: ""}
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

// func normalizeLineEndings(s string) string {
// 	return strings.ReplaceAll(s, "\r\n", "\n")
// }
