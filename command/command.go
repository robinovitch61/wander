package command

import (
	tea "github.com/charmbracelet/bubbletea"
	"strings"
	"wander/message"
	"wander/nomad"
)

func FetchLogs(url, token, allocId, taskName string, logType nomad.LogType) tea.Cmd {
	return func() tea.Msg {
		// TODO LEO: error handling
		logTypeString := "stdout"
		switch logType {
		case nomad.StdOut:
			logTypeString = "stdout"
		case nomad.StdErr:
			logTypeString = "stderr"
		}

		body, _ := nomad.GetLogs(url, token, allocId, taskName, logTypeString)
		//simulateLoading()
		//body := MockLogsResponse
		var logRows []nomad.LogRow
		for _, log := range strings.Split(string(body), "\n") {
			logRows = append(logRows, nomad.LogRow(log))
		}
		return message.NomadLogsMsg{LogType: logType, Data: logRows}
	}
}
