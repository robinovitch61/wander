package message

import "wander/components/page"

type ErrMsg struct{ Err error }

type ChangePageMsg struct{ NewPage page.Page }

func (e ErrMsg) Error() string { return e.Err.Error() }
