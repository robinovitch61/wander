package logline

import (
	"strings"
)

type loglineData struct {
	allData, filteredData []loglineRow
}

type loglineRow string

func (e loglineRow) MatchesFilter(filter string) bool {
	return strings.Contains(string(e), filter)
}

func logsAsString(logs []loglineRow) string {
	// is there a better way to do this in Go? Seems silly
	var logRows []string
	for _, row := range logs {
		logRows = append(logRows, string(row))
	}
	return strings.Join(logRows, "\n")
}
