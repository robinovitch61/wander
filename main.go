package main

import (
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"strings"
	"wander/components/viewport"
	"wander/formatter"
	"wander/nomad"
)

var (
	NomadTokenEnvVariable = "NOMAD_TOKEN"
	NomadUrlEnvVariable   = "NOMAD_URL"
)

type model struct {
	nomadToken string
	nomadUrl   string
	nomadJobs  []string
	viewport   viewport.Model
	ready      bool
	err        error
}

// messages
type nomadJobsMsg []string

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

// commands
func fetchJobs(url, token string) tea.Cmd {
	return func() tea.Msg {
		// TODO LEO: error handling
		//body, _ := nomad.GetJobs(url, token)
		body := MockJobsResponse
		var jobResponse []nomad.JobResponseEntry
		if err := json.Unmarshal(body, &jobResponse); err != nil {
			fmt.Println("Failed to decode response")
		}

		return nomadJobsMsg(formatter.JobResponsesAsTable(jobResponse))
	}
}

func (m model) Init() tea.Cmd {
	return fetchJobs(m.nomadUrl, m.nomadToken)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {

	case nomadJobsMsg:
		m.nomadJobs = msg
		m.viewport.SetContent(strings.Join(m.nomadJobs, "\n"))
		return m, nil

	case errMsg:
		m.err = msg
		return m, nil

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyEsc {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		headerHeight := 0
		footerHeight := 0
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
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

	if len(m.nomadJobs) == 0 {
		return "Retrieving jobs..."
	}
	return m.viewport.View()
}

func initialModel() model {
	nomadToken := os.Getenv(NomadTokenEnvVariable)
	if nomadToken == "" {
		fmt.Printf("Set environment variable %s", NomadTokenEnvVariable)
		os.Exit(1)
	}

	nomadUrl := os.Getenv(NomadUrlEnvVariable)
	if nomadUrl == "" {
		fmt.Printf("Set environment variable %s", NomadUrlEnvVariable)
		os.Exit(1)
	}
	return model{
		nomadToken: nomadToken,
		nomadUrl:   nomadUrl,
		nomadJobs:  nil,
		err:        nil,
	}
}

func main() {
	program := tea.NewProgram(initialModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())

	if err := program.Start(); err != nil {
		fmt.Printf("Error on wander startup: %v\n", err)
		os.Exit(1)
	}
}
