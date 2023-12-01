//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || aix
// +build darwin dragonfly freebsd linux netbsd openbsd solaris aix

package signals

import (
	"context"
	"github.com/hashicorp/nomad/api"
	"github.com/moby/term"
	"io"
	"os"
	"os/signal"
	"syscall"
)

// WatchTerminalSize watches terminal size changes to propagate to remote tty.
func WatchTerminalSize(out io.Reader, resize chan<- api.TerminalSize) (func(), error) {
	fd, _ := term.GetFdInfo(out)

	ctx, cancel := context.WithCancel(context.Background())

	signalCh := make(chan os.Signal, 1)
	setupWindowNotification(signalCh)

	sendTerminalSize := func() {
		s, err := term.GetWinsize(fd)
		if err != nil {
			return
		}

		resize <- api.TerminalSize{
			Height: int(s.Height),
			Width:  int(s.Width),
		}
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-signalCh:
				sendTerminalSize()
			}
		}
	}()

	sendTerminalSize()
	return cancel, nil
}

func setupWindowNotification(ch chan<- os.Signal) {
	signal.Notify(ch, syscall.SIGWINCH)
}
