package pages

import "fmt"

type Page int8

const (
	Unset Page = iota
	Jobs
	Allocations
	Logs
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
	}
	return "unknown"
}

func (p Page) LoadingString() string {
	return fmt.Sprintf("Loading %s...", p.String())
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
	}
	return p
}

func (p Page) Backward() Page {
	switch p {
	case Allocations:
		return Jobs
	case Logs:
		return Allocations
	}
	return p
}
