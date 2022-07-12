package app

import "sync"

var (
	updateID    int
	updateIDMtx sync.Mutex
)

func nextUpdateID() int {
	updateIDMtx.Lock()
	defer updateIDMtx.Unlock()
	updateID++
	return updateID
}
