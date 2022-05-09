package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"strings"
	"wander/command"
	"wander/components/header"
	"wander/components/viewport"
	"wander/dev"
	"wander/message"
	"wander/nomad"
	"wander/page"
)

var (
	NomadTokenEnvVariable = "NOMAD_TOKEN"
	NomadUrlEnvVariable   = "NOMAD_URL"
)

type model struct {
	nomadToken    string
	nomadUrl      string
	keyMap        mainKeyMap
	page          page.Page
	header        header.Model
	viewport      viewport.Model
	width, height int
	initialized   bool
	err           error
	nomadJobsList []nomad.JobResponseEntry
}

func (m model) Init() tea.Cmd {
	return command.FetchJobs(m.nomadUrl, m.nomadToken)
}

func fetchPageDataCmd(m model) tea.Cmd {
	switch m.page {

	case page.Jobs:
		return command.FetchJobs(m.nomadUrl, m.nomadToken)

	case page.Allocation:
		selectedJob := m.nomadJobsList[m.viewport.CursorRow]                     // TODO LEO: this may not remain true with search/filtering
		return command.FetchAllocation(m.nomadUrl, m.nomadToken, selectedJob.ID) // TODO LEO: correct ID
	}
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {

	case message.NomadJobsMsg:
		dev.Debug("nomadJobsMsg")
		m.nomadJobsList = msg.Jobs
		m.viewport.SetHeader(strings.Join(msg.Table.HeaderRows, "\n"))
		m.viewport.SetContent(strings.Join(msg.Table.ContentRows, "\n"))
		return m, nil

	case message.NomadAllocationMsg:
		dev.Debug("NomadAllocationMsg")
		m.viewport.SetHeader(strings.Join(msg.Table.HeaderRows, "\n"))
		m.viewport.SetContent(strings.Join(msg.Table.ContentRows, "\n"))
		return m, nil

	case message.ErrMsg:
		dev.Debug("errMsg")
		m.err = msg
		return m, nil

	case tea.KeyMsg:
		dev.Debug(fmt.Sprintf("KeyMsg '%s'", msg))
		switch {

		case key.Matches(msg, m.keyMap.Exit):
			return m, tea.Quit

		case key.Matches(msg, m.keyMap.Enter):
			if newPage := m.page.Forward(); newPage != m.page {
				m.page = newPage
				cmd := fetchPageDataCmd(m)
				m.viewport.SetLoading(newPage.LoadingString())
				return m, cmd
			}

		case key.Matches(msg, m.keyMap.Back):
			if newPage := m.page.Backward(); newPage != m.page {
				m.page = newPage
				cmd := fetchPageDataCmd(m)
				m.viewport.SetLoading(newPage.LoadingString())
				return m, cmd
			}

		case key.Matches(msg, m.keyMap.Reload):
			cmd := fetchPageDataCmd(m)
			m.viewport.SetLoading(m.page.ReloadingString())
			return m, cmd
		}

	case tea.WindowSizeMsg:
		dev.Debug("WindowSizeMsg")
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := m.header.ViewHeight()
		footerHeight := 0
		viewportHeight := msg.Height - (headerHeight + footerHeight)

		if !m.initialized {
			// this is the first message received and the initial entrypoint to the app
			m.keyMap = getKeyMap()
			m.viewport = viewport.New(msg.Width, viewportHeight)
			m.initialized = true

			firstPage := page.Jobs
			m.page = firstPage
			m.viewport.SetLoading(firstPage.LoadingString())

			cmd := fetchPageDataCmd(m)
			return m, cmd
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

	finalView := m.header.View() + "\n" + m.viewport.View()
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
