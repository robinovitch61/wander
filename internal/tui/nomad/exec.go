package nomad

import (
	"bufio"
	"context"
	"fmt"
	"github.com/hashicorp/nomad/api"
	"github.com/moby/term"
	"github.com/robinovitch61/wander/internal/tui/formatter"
	"golang.org/x/exp/maps"
	"golang.org/x/sys/unix"
	"io"
	"os"
	"os/signal"
	"syscall"
)

// TODO:
// - [x] try with bash
// - [x] improve help text on wander exec
// - [x] clear screen on start of exec
// - [x] don't print config file used on exec
// - [x] warning message that "you're in wander"
// - [x] cmd human friendly (not full 36char id)
// - [x] code review this
// - [ ] update gif
// - [ ] fix wander serve exec

type ExecCompleteMsg struct {
	Output string
}

type StdoutProxy struct {
	SavedOutput []byte
}

func (so *StdoutProxy) Write(p []byte) (n int, err error) {
	so.SavedOutput = append(so.SavedOutput, p...)
	return os.Stdout.Write(p)
}

func findAllocsForJobPrefix(client api.Client, jobName string) map[string][]*api.AllocationListStub {
	allocs := make(map[string][]*api.AllocationListStub)
	jobs, _, err := client.Jobs().PrefixList(jobName)
	if err != nil || len(jobs) == 0 {
		return allocs
	}
	for _, job := range jobs {
		jobAllocs, _, err := client.Jobs().Allocations(job.ID, true, &api.QueryOptions{Namespace: job.Namespace})
		if err != nil || len(jobAllocs) == 0 {
			continue
		}
		allocs[job.ID] = jobAllocs
	}
	return allocs
}

func AllocExec(client *api.Client, allocID, task string, args []string) (int, error) {
	alloc, _, err := client.Allocations().Info(allocID, nil)
	if err != nil {
		// maybe allocID is actually a job name
		foundAllocs := findAllocsForJobPrefix(*client, allocID)
		if len(foundAllocs) > 0 {
			if len(foundAllocs) == 1 && len(maps.Values(foundAllocs)[0]) == 1 && maps.Values(foundAllocs)[0][0] != nil {
				// only one job with one allocation found, use that
				alloc, _, err = client.Allocations().Info(maps.Values(foundAllocs)[0][0].ID, nil)
			} else {
				// multiple jobs and/or allocations found, print them and exit
				for job, jobAllocs := range foundAllocs {
					fmt.Printf("allocations for job %s:\n", job)
					for _, alloc := range jobAllocs {
						fmt.Printf("  %s (%s in %s)\n", formatter.ShortAllocID(alloc.ID), alloc.Name, alloc.Namespace)
					}
				}
				return 1, nil
			}
		} else {
			// maybe allocID is short form of id
			shortIDAllocs, _, err := client.Allocations().List(&api.QueryOptions{Prefix: allocID})
			if err != nil {
				return 1, fmt.Errorf("no jobs or allocation id for %s found: %v", allocID, err)
			}
			if len(shortIDAllocs) > 1 {
				// rare but possible that uuid prefixes match
				fmt.Printf("prefix %s matched multiple allocations:\n", allocID)
				for _, alloc := range shortIDAllocs {
					fmt.Printf("  %s (%s in %s)\n", formatter.ShortAllocID(alloc.ID), alloc.Name, alloc.Namespace)
				}
				return 1, err
			} else if len(shortIDAllocs) == 1 {
				alloc, _, err = client.Allocations().Info(shortIDAllocs[0].ID, nil)
			} else {
				return 1, fmt.Errorf("no allocations found for alloc id %s", allocID)
			}
		}
	}

	// if task is blank, user has assumed that there is only one task in the allocation and wants to
	// use that
	if task == "" {
		if len(alloc.TaskStates) == 1 {
			for taskName := range alloc.TaskStates {
				task = taskName
			}
		} else {
			fmt.Printf("multiple tasks found in allocation %s (%s in %s)\n", formatter.ShortAllocID(alloc.ID), alloc.Name, alloc.Namespace)
			for taskName := range alloc.TaskStates {
				fmt.Printf("  %s\n", taskName)
			}
		}
	}

	code, err := execImpl(client, alloc, task, args, "~", os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		return 1, err
	}
	return code, nil
}

// execImpl invokes the Alloc Exec api call, it also prepares and restores terminal states as necessary.
func execImpl(client *api.Client, alloc *api.Allocation, task string,
	command []string, escapeChar string, stdin io.Reader, stdout, stderr io.WriteCloser) (int, error) {

	// attempt to clear screen
	os.Stdout.Write([]byte("\033c"))
	fmt.Println(fmt.Sprintf("Exec session for %s (%s), task %s", alloc.Name, formatter.ShortAllocID(alloc.ID), task))

	sizeCh := make(chan api.TerminalSize, 1)

	ctx, cancelFn := context.WithCancel(context.Background())
	actuallyCancel := func() {
		cancelFn()
	}
	defer actuallyCancel()

	inCleanup, err := setRawTerminal(stdin)
	if err != nil {
		return -1, err
	}
	defer inCleanup()

	outCleanup, err := setRawTerminalOutput(stdout)
	if err != nil {
		return -1, err
	}
	defer outCleanup()

	sizeCleanup, err := watchTerminalSize(stdin, sizeCh)
	if err != nil {
		return -1, err
	}
	defer sizeCleanup()

	stdin = NewReader(stdin, escapeChar[0], func(c byte) bool {
		switch c {
		case '.':
			// need to restore tty state so error reporting here
			// gets emitted at beginning of line
			outCleanup()
			inCleanup()

			stderr.Write([]byte("\nConnection closed\n"))
			cancelFn()
			return true
		default:
			return false
		}
	})

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		for range signalCh {
			cancelFn()
		}
	}()

	return client.Allocations().Exec(ctx,
		alloc, task, true, command, stdin, stdout, stderr, sizeCh, nil)
}

// setRawTerminal sets the stream terminal in raw mode, so process captures
// Ctrl+C and other commands to forward to remote process.
// It returns a cleanup function that restores terminal to original mode.
func setRawTerminal(stream interface{}) (cleanup func(), err error) {
	fd, _ := term.GetFdInfo(stream)

	state, err := term.SetRawTerminal(fd)
	if err != nil {
		return nil, err
	}

	return func() {
		term.RestoreTerminal(fd, state)
	}, nil
}

// setRawTerminalOutput sets the output stream in Windows to raw mode,
// so it disables LF -> CRLF translation.
// It's basically a no-op on unix.
func setRawTerminalOutput(stream interface{}) (cleanup func(), err error) {
	fd, _ := term.GetFdInfo(stream)

	state, err := term.SetRawTerminalOutput(fd)
	//_, err = term.SetRawTerminalOutput(fd)
	if err != nil {
		return nil, err
	}

	return func() {
		term.RestoreTerminal(fd, state)
	}, nil
}

// watchTerminalSize watches terminal size changes to propagate to remote tty.
func watchTerminalSize(out io.Reader, resize chan<- api.TerminalSize) (func(), error) {
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

	go func() {
		// send initial size
		sendTerminalSize()
	}()

	return cancel, nil
}

func setupWindowNotification(ch chan<- os.Signal) {
	signal.Notify(ch, unix.SIGWINCH)
}

// Handler is a callback for handling an escaped char.  Reader would skip
// the escape char and passed char if returns true; otherwise, it preserves them
// in output
type Handler func(c byte) bool

// NewReader returns a reader that escapes the c character (following new lines),
// in the same manner OpenSSH handling, which defaults to `~`.
//
// For illustrative purposes, we use `~` in documentation as a shorthand for escaping character.
//
// If following a new line, reader sees:
//   - `~~`, only one is emitted
//   - `~.` (or any character), the handler is invoked with the character.
//     If handler returns true, `~.` will be skipped; otherwise, it's propagated.
//   - `~` and it's the last character in stream, it's propagated
//
// Appearances of `~` when not preceded by a new line are propagated unmodified.
func NewReader(r io.Reader, c byte, h Handler) io.Reader {
	pr, pw := io.Pipe()
	reader := &reader{
		impl:       r, // stdin
		escapeChar: c,
		handler:    h,
		pr:         pr,
		pw:         pw,
	}
	go reader.pipe()
	return reader
}

// lookState represents the state of reader for what character of `\n~.` sequence
// reader is looking for
type lookState int

const (
	// sLookNewLine indicates that reader is looking for new line
	sLookNewLine lookState = iota

	// sLookEscapeChar indicates that reader is looking for ~
	sLookEscapeChar

	// sLookChar indicates that reader just read `~` is waiting for next character
	// before acting
	sLookChar
)

// to ease comments, i'll assume escape character to be `~`
type reader struct {
	impl       io.Reader
	escapeChar uint8
	handler    Handler

	// buffers
	pw *io.PipeWriter
	pr *io.PipeReader
}

func (r *reader) Read(buf []byte) (int, error) {
	return r.pr.Read(buf)
}

func (r *reader) pipe() {
	rb := make([]byte, 4096)
	bw := bufio.NewWriter(r.pw)

	state := sLookEscapeChar

	for {
		n, err := r.impl.Read(rb)

		if n > 0 {
			state = r.processBuf(bw, rb, n, state)
			bw.Flush()
			if state == sLookChar {
				// terminated with ~ - let's read one more character
				n, err = r.impl.Read(rb[:1])
				if n == 1 {
					state = sLookNewLine
					if rb[0] == r.escapeChar {
						// only emit escape character once
						bw.WriteByte(rb[0])
						bw.Flush()
					} else if r.handler(rb[0]) {
						// skip if handled
					} else {
						bw.WriteByte(r.escapeChar)
						bw.WriteByte(rb[0])
						bw.Flush()
						if rb[0] == '\n' || rb[0] == '\r' {
							state = sLookEscapeChar
						}
					}
				}
			}
		}

		if err != nil {
			// write ~ if it's the last thing
			if state == sLookChar {
				bw.WriteByte(r.escapeChar)
			}
			bw.Flush()
			r.pw.CloseWithError(err)
			break
		}
	}
}

// processBuf process buffer and emits all output to writer
// if the last part of buffer is a new line followed by sequnce, it writes
// all output until the new line and returns sLookChar
func (r *reader) processBuf(bw io.Writer, buf []byte, n int, s lookState) lookState {
	i := 0

	wi := 0

START:
	if s == sLookEscapeChar && buf[i] == r.escapeChar {
		if i+1 >= n {
			// buf terminates with ~ - write all before
			bw.Write(buf[wi:i])
			return sLookChar
		}

		nc := buf[i+1]
		if nc == r.escapeChar {
			// skip one escape char
			bw.Write(buf[wi:i])
			i++
			wi = i
		} else if r.handler(nc) {
			// skip both characters
			bw.Write(buf[wi:i])
			i = i + 2
			wi = i
		} else if nc == '\n' || nc == '\r' {
			i = i + 2
			s = sLookEscapeChar
			goto START
		} else {
			i = i + 2
			// need to write everything keep going
		}
	}

	// search until we get \n~, or buf terminates
	for {
		if i >= n {
			// got to end without new line, write and return
			bw.Write(buf[wi:n])
			return sLookNewLine
		}

		if buf[i] == '\n' || buf[i] == '\r' {
			// buf terminated at new line
			if i+1 >= n {
				bw.Write(buf[wi:n])
				return sLookEscapeChar
			}

			// peek to see escape character go back to START if so
			if buf[i+1] == r.escapeChar {
				s = sLookEscapeChar
				i++
				goto START
			}
		}

		i++
	}
}
