package logline

import (
	"strings"
)

type loglineData struct {
	allData, filteredData []LoglineRow
}

type LoglineRow string

func (e LoglineRow) MatchesFilter(filter string) bool {
	return strings.Contains(string(e), filter)
}

func logsAsString(logs []LoglineRow) string {
	// is there a better way to do this in Go? Seems silly
	var logRows []string
	for _, row := range logs {
		logRows = append(logRows, string(row))
	}
	return strings.Join(logRows, "\n")
}
