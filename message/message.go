package message

import (
	"wander/components/page"
	"wander/pages"
)

type ErrMsg struct{ Err error }

func (e ErrMsg) Error() string { return e.Err.Error() }

type PageLoadMsg struct {
	Page        pages.Page
	TableHeader []string
	AllPageData []page.Row
}
