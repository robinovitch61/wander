package main

import (
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"io/ioutil"
	"net/http"
	"os"
)

var (
	NomadTokenEnvVariable = "NOMAD_TOKEN"
	NomadUrlEnvVariable   = "NOMAD_URL"
)

type model struct {
	nomadToken   string
	nomadUrl     string
	nomadJobList []string
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
	switch msg := msg.(type) {

	case nomadJobListMsg:
		m.nomadJobList = msg
		return m, nil

	case errMsg:
		m.err = msg
		return m, nil

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyEsc {
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() string {
	// If there's an error, print it out and don't do anything else.
	if m.err != nil {
		return fmt.Sprintf("\nError: %v\n\n", m.err)
	}

	// Tell the user we're doing something.
	s := "Loading..."

	// When the server responds with a status, add it to the current line.
	if len(m.nomadJobList) > 0 {
		s = ""
		for _, job := range m.nomadJobList {
			s += fmt.Sprintf("\n%s", job)
		}
	}

	// Send off whatever we came up with above for rendering.
	return "\n" + s + "\n\n"
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
	if err := tea.NewProgram(initialModel()).Start(); err != nil {
		fmt.Printf("Error on wander startup: %v\n", err)
		os.Exit(1)
	}
}
