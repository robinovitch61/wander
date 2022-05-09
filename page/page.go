package page

import "fmt"

type Page int8

const (
	Unset Page = iota
	Jobs
	Allocation
	Logs
)

func (p Page) String() string {
	switch p {
	case Unset:
		return "undefined"
	case Jobs:
		return "jobs"
	case Allocation:
		return "allocation"
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
		return Allocation
	}
	return p
}

func (p Page) Backward() Page {
	switch p {
	case Allocation:
		return Jobs
	}
	return p
}
