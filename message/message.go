package message

import (
	"wander/formatter"
	"wander/nomad"
)

// NomadJobsMsg is a message for nomad jobs
type NomadJobsMsg struct {
	Table     formatter.Table
	TableData []nomad.JobResponseEntry
}

// NomadAllocationMsg is a message for nomad allocations
type NomadAllocationMsg struct {
	Table formatter.Table
}

// ErrMsg is an error message
type ErrMsg struct{ err error }

func (e ErrMsg) Error() string { return e.err.Error() }
