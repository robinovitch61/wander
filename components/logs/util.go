package logs

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
	"wander/formatter"
	"wander/message"
	"wander/nomad"
)

type nomadLogsData struct {
	allData, filteredData []logRow
}

type nomadLogsMsg struct {
	LogType LogType
	Data    []logRow
}

type logRow string

func (e logRow) MatchesFilter(filter string) bool {
	return strings.Contains(string(e), filter)
}

type LogType int8

const (
	StdOut LogType = iota
	StdErr
)

func (p LogType) String() string {
	switch p {
	case StdOut:
		return "Stdout Logs"
	case StdErr:
		return "Stderr Logs"
	}
	return "Stdout Logs"
}

func (p LogType) ShortString() string {
	switch p {
	case StdOut:
		return "stdout"
	case StdErr:
		return "stderr"
	}
	return "stdout"
}

func FetchLogs(url, token, allocID, taskName string, logType LogType) tea.Cmd {
	return func() tea.Msg {
		logTypeString := "stdout"
		switch logType {
		case StdOut:
			logTypeString = "stdout"
		case StdErr:
			logTypeString = "stderr"
		}

		params := map[string]string{
			"task":   taskName,
			"type":   logTypeString,
			"origin": "end",
			"offset": "1000000",
			"plain":  "true",
		}
		fullPath := fmt.Sprintf("%s%s%s", url, "/v1/client/fs/logs/", allocID)
		body, err := nomad.Get(fullPath, token, params)
		if err != nil {
			return message.ErrMsg{Err: err}
		}

		var logRows []logRow
		for _, log := range strings.Split(string(body), "\n") {
			logRows = append(logRows, logRow(log))
		}
		return nomadLogsMsg{LogType: logType, Data: logRows}
	}
}

func logsAsTable(logs []logRow, logType LogType) formatter.Table {
	var logRows [][]string
	// ignore the first log line because it may be truncated due to offset
	// TODO LEO: check if there's actually a truncated line based on the offset size and log char length^
	//for _, row := range logs[1:] {
	for _, row := range logs {
		if stripped := strings.TrimSpace(string(row)); stripped != "" {
			logRows = append(logRows, []string{stripped})
		}
	}

	return formatter.GetRenderedTableAsString(
		[]string{logType.String()},
		logRows,
	)
}
