package nomad

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hashicorp/nomad/api"
	"github.com/robinovitch61/wander/internal/dev"
	"github.com/robinovitch61/wander/internal/tui/components/page"
	"github.com/robinovitch61/wander/internal/tui/formatter"
	"strings"
	"time"
)

type LogType int8

const (
	StdOut LogType = iota
	StdErr
)

type LogsStreamMsg struct {
	Value string // may include line breaks
	Type  LogType
}

func (p LogType) String() string {
	switch p {
	case StdOut:
		return "Stdout Logs"
	case StdErr:
		return "Stderr Logs"
	}
	return "unknown"
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

func FetchLogs(client api.Client, alloc api.Allocation, taskName string, logType LogType, logOffset int) tea.Cmd {
	return func() tea.Msg {
		// This is currently very important and strange. The logs api attempts to go through the node directly
		// by default. The default timeout for this is 1 second. If it fails, it falls silently to going through
		// the server. Since it always fails, at least in my Nomad setup, make it timeout immediately by setting
		// the timeout to something tiny.
		api.ClientConnTimeout = 1 * time.Microsecond

		closeLogConn := make(chan struct{}) // never closed for now
		dev.Debug(fmt.Sprintf("offset is %d", logOffset))
		logsChan, _ := client.AllocFS().Logs( // TODO LEO: deal with error channel
			&alloc,
			true,
			taskName,
			logType.ShortString(),
			"end",
			int64(logOffset),
			closeLogConn,
			nil,
		)

		tableHeader, allPageData := logsAsTable([]string{}, logType)
		return PageLoadedMsg{Page: LogsPage, TableHeader: tableHeader, AllPageRows: allPageData, LogsStream: LogsStream{logsChan, logType}}
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

func ReadLogsStreamNextMessage(c LogsStream) tea.Cmd {
	return func() tea.Msg {
		line := <-c.Chan
		dev.Debug(fmt.Sprintf("log offset %d", line.Offset))
		tabReplacedLine := strings.ReplaceAll(string(line.Data), "\t", "    ")
		return LogsStreamMsg{Value: tabReplacedLine, Type: c.LogType}
	}
}
