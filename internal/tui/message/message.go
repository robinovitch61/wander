package message

type ErrMsg struct{ Err error }

func (e ErrMsg) Error() string { return e.Err.Error() }

type PageInputReceivedMsg struct {
	Input string
}

type CleanupCompleteMsg struct{}
