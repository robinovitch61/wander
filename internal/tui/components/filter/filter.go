package filter

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinovitch61/wander/internal/dev"
	"github.com/robinovitch61/wander/internal/tui/style"
)

var (
	keyMap = getKeyMap()
)

type Model struct {
	prefix    string
	keyMap    filterKeyMap
	textinput textinput.Model
}

func New(prefix string) Model {
	ti := textinput.New()
	ti.Prompt = ""
	ti.SetCursorMode(textinput.CursorHide)
	return Model{
		prefix:    prefix,
		keyMap:    keyMap,
		textinput: ti,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	dev.Debug(fmt.Sprintf("filter %T", msg))
	var cmd tea.Cmd
	m.textinput, cmd = m.textinput.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.textinput.Focused() {
		m.textinput.TextStyle = style.FilterEditing
		m.textinput.PromptStyle = style.FilterEditing
		m.textinput.Cursor.TextStyle = style.FilterEditing
		if len(m.textinput.Value()) > 0 {
			// editing existing filter
			m.textinput.Prompt = "filter: "
		} else {
			// editing but no filter value yet
			m.textinput.Prompt = ""
			m.textinput.SetCursorMode(textinput.CursorHide)
			m.textinput.SetValue("type to filter")
		}
	} else {
		if len(m.textinput.Value()) > 0 {
			// filter applied, not editing
			m.textinput.Prompt = "filter: "
			m.textinput.Cursor.TextStyle = style.FilterApplied
			m.textinput.PromptStyle = style.FilterApplied
			m.textinput.TextStyle = style.FilterApplied
		} else {
			// no filter, not editing
			m.textinput.Prompt = ""
			m.textinput.PromptStyle = style.Regular
			m.textinput.TextStyle = style.Regular
			m.textinput.SetValue("'/' to filter")
		}
	}
	filterString := m.textinput.View()
	filterStringStyle := m.textinput.TextStyle.Copy().MarginLeft(1).PaddingLeft(1).PaddingRight(0)
	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		style.FilterPrefix.Render(m.prefix),
		filterStringStyle.Render(filterString),
	)
}

func (m Model) Value() string {
	return m.textinput.Value()
}

func (m Model) ViewHeight() int {
	return lipgloss.Height(m.View())
}

func (m *Model) SetPrefix(prefix string) {
	m.prefix = prefix
}

func (m Model) Focused() bool {
	return m.textinput.Focused()
}

func (m *Model) Focus() {
	m.textinput.SetCursorMode(textinput.CursorBlink)
	m.textinput.Focus()
}

func (m *Model) Blur() {
	// move cursor to end of word so right padding shows up even if cursor not at end when blurred
	m.textinput.SetCursor(len(m.textinput.Value()))

	m.textinput.SetCursorMode(textinput.CursorHide)
	m.textinput.Blur()
}

func (m *Model) BlurAndClear() {
	m.Blur()
	m.textinput.SetValue("")
}
