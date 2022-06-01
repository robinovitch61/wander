package nomad

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
	"wander/components/page"
	"wander/formatter"
	"wander/message"
)

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
	return "Unknown"
}

func (p LogType) ShortString() string {
	switch p {
	case StdOut:
		return "stdout"
	case StdErr:
		return "stderr"
	}
	return "unknown"
}

func FetchLogs(url, token, allocID, taskName string, logType LogType) tea.Cmd {
	return func() tea.Msg {
		params := map[string]string{
			"task":   taskName,
			"type":   logType.ShortString(),
			"origin": "end",
			"offset": "1000000",
			"plain":  "true",
		}
		fullPath := fmt.Sprintf("%s%s%s", url, "/v1/client/fs/logs/", allocID)
		body, err := get(fullPath, token, params)
		if err != nil {
			return message.ErrMsg{Err: err}
		}

		logRows := strings.Split(string(body), "\n")

		tableHeader, allPageData := logsAsTable(logRows, logType)
		return PageLoadedMsg{Page: LogsPage, TableHeader: tableHeader, AllPageData: allPageData}
	}
}

func logsAsTable(logs []string, logType LogType) ([]string, []page.Row) {
	var logRows [][]string
	var keys []string
	for _, row := range logs {
		if stripped := strings.TrimSpace(row); stripped != "" {
			logRows = append(logRows, []string{row})
		}
		keys = append(keys, "")
	}

	columns := []string{logType.String()}
	table := formatter.GetRenderedTableAsString(columns, logRows)

	var rows []page.Row
	for idx, row := range table.ContentRows {
		rows = append(rows, page.Row{Key: keys[idx], Row: row})
	}

	return table.HeaderRows, rows
}
