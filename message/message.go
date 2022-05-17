package message

type ErrMsg struct{ Err error }

type ViewJobsMsg struct{}

type ViewAllocationsMsg struct{}

type ViewLogsMsg struct{}

func (e ErrMsg) Error() string { return e.Err.Error() }
