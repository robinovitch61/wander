package message

import (
	"wander/components/page"
	"wander/nomad"
)

// NomadLogsMsg is a message for nomad logs
type NomadLogsMsg struct {
	LogType nomad.LogType
	Data    []nomad.LogRow
}

// GoToPageMsg is a message to go to a new page
type GoToPageMsg page.Page

// ErrMsg is an error message
type ErrMsg struct{ Err error }

// ViewJobsPageMsg is a message to view the jobs page
type ViewJobsPageMsg struct{}

func (e ErrMsg) Error() string { return e.Err.Error() }
