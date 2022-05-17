package nomad

import (
	"strings"
)

// LogRow is a log line
type LogRow string

func (e LogRow) MatchesFilter(filter string) bool {
	return strings.Contains(string(e), filter)
}

// LogType is an enum for the log type
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
