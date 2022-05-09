package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"strings"
	"wander/command"
	"wander/components/header"
	"wander/components/viewport"
	"wander/dev"
	"wander/formatter"
	"wander/message"
)

var (
	NomadTokenEnvVariable = "NOMAD_TOKEN"
	NomadUrlEnvVariable   = "NOMAD_URL"
)

type Page int8

func (s Page) String() string {
	switch s {
	case Unset:
		return "undefined"
	case Jobs:
		return "jobs"
	case Allocation:
		return "allocation"
	case Logs:
		return "logs"
	}
	return "unknown"
}

const (
	Unset Page = iota
	Jobs
	Allocation
	Logs
)

type model struct {
	nomadToken    string
	nomadUrl      string
	nomadJobTable formatter.Table
	header        header.Model
	page          Page
	viewport      viewport.Model
	width, height int
	initialized   bool
	err           error
}

func (m model) Init() tea.Cmd {
	return command.FetchJobs(m.nomadUrl, m.nomadToken)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	if m.page == Unset {
		m.page = Jobs
	}

	switch msg := msg.(type) {

	case message.NomadJobsMsg:
		dev.Debug("nomadJobsMsg")
		m.nomadJobTable = msg.Table
		m.viewport.SetHeader(strings.Join(msg.Table.HeaderRows, "\n"))
		m.viewport.SetContent(strings.Join(msg.Table.ContentRows, "\n"))
		return m, nil

	case message.ErrMsg:
		dev.Debug("errMsg")
		m.err = msg
		return m, nil

	case tea.KeyMsg:
		dev.Debug("KeyMsg")
		if msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyEsc {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		dev.Debug("WindowSizeMsg")
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := m.header.ViewHeight()
		dev.Debug(fmt.Sprintf("LEO HERE %d", headerHeight))
		footerHeight := 0
		viewportHeight := msg.Height - (headerHeight + footerHeight)

		if !m.initialized {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.viewport = viewport.New(msg.Width, viewportHeight)
			m.initialized = true
		} else {
			m.viewport.SetWidth(msg.Width)
			m.viewport.SetHeight(viewportHeight)
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nError: %v\n\n", m.err)
	}

	finalView := m.header.View() + "\n"

	if m.page == Jobs && m.nomadJobTable.IsEmpty() {
		finalView += "Retrieving jobs..."
	}
	finalView += m.viewport.View()
	return finalView
	//return lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(finalView)
}

func initialModel() model {
	nomadToken := os.Getenv(NomadTokenEnvVariable)
	if nomadToken == "" {
		fmt.Printf("Set environment variable %s\n", NomadTokenEnvVariable)
		os.Exit(1)
	}

	nomadUrl := os.Getenv(NomadUrlEnvVariable)
	if nomadUrl == "" {
		fmt.Printf("Set environment variable %s\n", NomadUrlEnvVariable)
		os.Exit(1)
	}

	return model{
		nomadToken: nomadToken,
		nomadUrl:   nomadUrl,
		header:     header.New(nomadUrl),
	}
}

func main() {
	program := tea.NewProgram(initialModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())

	dev.Debug("~STARTING UP~")
	if err := program.Start(); err != nil {
		fmt.Printf("Error on wander startup: %v\n", err)
		os.Exit(1)
	}
}
