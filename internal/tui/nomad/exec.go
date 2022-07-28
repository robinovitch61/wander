package nomad

import (
	"bytes"
	"context"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hashicorp/nomad/api"
	"github.com/robinovitch61/wander/internal/tui/components/page"
	"github.com/robinovitch61/wander/internal/tui/message"
	"io"
	"time"
)

type ExecSessionConnectedMsg struct{}

type ExecSessionIO struct {
	StdInReader        io.Reader
	StdOutAndErrWriter io.Writer
}

type ExecWebSocketResponseMsg struct {
	Value string
	Close bool
}

func LoadExecPage() tea.Cmd {
	return func() tea.Msg {
		// this does no real work as the command input is requested before the exec websocket connects
		return PageLoadedMsg{Page: ExecPage, TableHeader: []string{}, AllPageRows: []page.Row{}}
	}
}

func InitiateWebSocket(client api.Client, namespace string, alloc api.Allocation, taskName, command string, session *ExecSessionIO, terminalSizeChan chan api.TerminalSize) tea.Cmd {
	return func() tea.Msg {
		// TODO LEO: something wrong here
		api.ClientConnTimeout = 1 * time.Nanosecond
		_, err := client.Allocations().Exec(
			context.Background(),
			&alloc,
			taskName,
			true,
			[]string{command},
			session.StdInReader,
			session.StdOutAndErrWriter,
			session.StdOutAndErrWriter,
			terminalSizeChan,
			nil,
		)
		if err != nil {
			return message.ErrMsg{Err: err}
		}

		return ExecSessionConnectedMsg{}
	}
}

func ReadExecSessionNextMessage(stdOutandErrWriter *bytes.Buffer) tea.Cmd {
	return func() tea.Msg {
		nextMsg := stdOutandErrWriter.Next(1)
		return ExecWebSocketResponseMsg{Value: string(nextMsg), Close: false}
	}
}

// func SendExecMessage(stdinReader io.Reader, keyPress string) tea.Cmd {
// 	return func() tea.Msg {
// 		_, err := stdinReader.Write([]byte(keyPress))
// 		if err != nil {
// 			return message.ErrMsg{Err: err}
// 		}
// 		return nil
// 	}
// }

func GetKeypress(msg tea.KeyMsg) (keypress string) {
	switch msg.Type {
	case tea.KeyEnter:
		keypress = "\n"
	case tea.KeySpace:
		keypress = " "
	case tea.KeyBackspace:
		if msg.Alt {
			keypress = string(rune(23))
		} else {
			keypress = string(rune(127))
		}
	case tea.KeyCtrlD:
		keypress = string(rune(4))
	case tea.KeyTab:
		keypress = string(rune(9))
	case tea.KeyUp:
		keypress = string(rune(27)) + "[A"
	case tea.KeyDown:
		keypress = string(rune(27)) + "[B"
	default:
		keypress = string(msg.Runes)
	}
	return keypress
}
