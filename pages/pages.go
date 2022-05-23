package pages

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
)

type Page int8

const (
	Unset Page = iota
	Jobs
	Allocations
	Logs
	Logline
)

func (p Page) String() string {
	switch p {
	case Unset:
		return "undefined"
	case Jobs:
		return "jobs"
	case Allocations:
		return "allocations"
	case Logs:
		return "logs"
	case Logline:
		return "log"
	}
	return "unknown"
}

func (p Page) LoadingString() string {
	return fmt.Sprintf("loading %s...", p.String())
}

func (p Page) ReloadingString() string {
	return fmt.Sprintf("Reloading %s...", p.String())
}

func (p Page) Forward() Page {
	switch p {
	case Jobs:
		return Allocations
	case Allocations:
		return Logs
	case Logs:
		return Logline
	}
	return p
}

func (p Page) Backward() Page {
	switch p {
	case Allocations:
		return Jobs
	case Logs:
		return Allocations
	case Logline:
		return Logs
	}
	return p
}

type ChangePageMsg struct{ NewPage Page }

func ToJobsPageCmd() tea.Msg {
	return ChangePageMsg{NewPage: Jobs}
}

func ToAllocationsPageCmd() tea.Msg {
	return ChangePageMsg{NewPage: Allocations}
}

func ToLogsPageCmd() tea.Msg {
	return ChangePageMsg{NewPage: Logs}
}

func ToLoglinePageCmd() tea.Msg {
	return ChangePageMsg{NewPage: Logline}
}
