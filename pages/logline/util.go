package logline

import (
	"strings"
	"wander/formatter"
)

type loglineData struct {
	allData, filteredData []loglineRow
}

type loglineRow string

func (e loglineRow) MatchesFilter(filter string) bool {
	return strings.Contains(string(e), filter)
}

func logsAsTable(logs []loglineRow) formatter.Table {
	var logRows [][]string
	for _, row := range logs {
		logRows = append(logRows, []string{string(row)})
	}

	return formatter.GetRenderedTableAsString(
		[]string{"Content"},
		logRows,
	)
}
