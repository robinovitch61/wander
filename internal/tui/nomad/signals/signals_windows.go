//go:build windows
// +build windows

package signals

import (
	"github.com/hashicorp/nomad/api"
	"io"
)

// WatchTerminalSize watches terminal size changes to propagate to remote tty.
func WatchTerminalSize(out io.Reader, resize chan<- api.TerminalSize) (func(), error) {
	// not available on windows because windows does not
	// implement syscall.SIGWINCH
	return func() {}, nil
}
