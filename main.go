package main

import (
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"wander/components/viewport"
)

var (
	NomadTokenEnvVariable = "NOMAD_TOKEN"
	NomadUrlEnvVariable   = "NOMAD_URL"
)

type model struct {
	nomadToken   string
	nomadUrl     string
	nomadJobList []string
	viewport     viewport.Model
	ready        bool
	err          error
}

// messages
type nomadJobListMsg []string

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

// commands
func fetchJobNames(url, token string) tea.Cmd {
	return func() tea.Msg {
		client := &http.Client{}
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/jobs", url), nil)
		if err != nil {
			return nomadJobListMsg([]string{err.Error()})
		}
		req.Header.Set("X-Nomad-Token", token)
		resp, err := client.Do(req)
		if err != nil {
			return nomadJobListMsg([]string{err.Error()})
		}

		type JobResponseEntry struct {
			Name string
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nomadJobListMsg([]string{err.Error()})
		}
		var jobResponse []JobResponseEntry
		if err := json.Unmarshal(body, &jobResponse); err != nil {
			fmt.Println("Failed to decode response")
		}

		var jobs []string
		for _, entry := range jobResponse {
			jobs = append(jobs, entry.Name)
		}
		return nomadJobListMsg(jobs)
	}
}

func (m model) Init() tea.Cmd {
	return fetchJobNames(m.nomadUrl, m.nomadToken)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {

	case nomadJobListMsg:
		m.nomadJobList = msg
		m.viewport.SetContent(strings.Join(m.nomadJobList, "\n"))
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
			m.viewport.YPosition = headerHeight
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

	if len(m.nomadJobList) == 0 {
		return "Loading..."
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
		nomadToken:   nomadToken,
		nomadUrl:     nomadUrl,
		nomadJobList: nil,
		err:          nil,
	}
}

func main() {
	program := tea.NewProgram(initialModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())

	if err := program.Start(); err != nil {
		fmt.Printf("Error on wander startup: %v\n", err)
		os.Exit(1)
	}
}
