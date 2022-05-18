package message

import "wander/pages"

type ErrMsg struct{ Err error }

type ChangePageMsg struct{ NewPage pages.Page }

func (e ErrMsg) Error() string { return e.Err.Error() }
