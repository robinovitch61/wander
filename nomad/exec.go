package nomad

import (
	b64 "encoding/base64"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
	"strings"
	"time"
	"wander/dev"
	"wander/formatter"
	"wander/message"
)

type ExecWebSocketConnectedMsg struct {
	WebSocketConnection *websocket.Conn
}

func InitiateExecWebSocketConnection(host, token, allocID, taskName, command string) tea.Cmd {
	return func() tea.Msg {
		jsonCommand, err := formatter.JsonEncodedTokenArray(command)
		if err != nil {
			return message.ErrMsg{Err: err}
		}

		secure := false
		if strings.Contains(host, "https://") {
			secure = true
		}

		host = strings.Split(host, "://")[1]

		path := fmt.Sprintf("/v1/client/allocation/%s/exec", allocID)
		params := map[string]string{
			"command": jsonCommand,
			"task":    taskName,
		}

		ws, err := getWebSocketConnection(secure, host, path, token, params)
		if err != nil {
			return message.ErrMsg{Err: err}
		}

		return ExecWebSocketConnectedMsg{WebSocketConnection: ws}
	}
}

type ExecWebSocketResponseMsg struct {
	StdOut, StdErr string
}

func ReadExecWebSocketNextMessage(ws *websocket.Conn) tea.Cmd {
	return func() tea.Msg {
		// TODO LEO: don't think we actually want this timeout as it cancels all future reads
		// setTimeoutErr := ws.SetReadDeadline(time.Now().Add(3 * time.Second))
		// if setTimeoutErr != nil {
		// 	return message.ErrMsg{Err: setTimeoutErr}
		// }

		dev.Debug("READING WS")
		msgType, content, err := ws.ReadMessage()
		dev.Debug(fmt.Sprintf("READ MESSAGE type %d content %s err %e", msgType, content, err))

		if err != nil {
			// timeouts can occur when there's just no response, like sending "/bin/sh" as a first command
			if !strings.Contains(err.Error(), "i/o timeout") {
				return message.ErrMsg{Err: err}
			} else {
				dev.Debug("RESET")
				setTimeoutErr := ws.SetReadDeadline(time.Time{})
				if setTimeoutErr != nil {
					return message.ErrMsg{Err: setTimeoutErr}
				}
			}
		}

		// TODO LEO: parse response here, handle output, decode the base64 and return real msg
		return ExecWebSocketResponseMsg{StdOut: "hi"}
	}
}

func SendAndReadExecWebSocketMessage(ws *websocket.Conn, shellCmd string) tea.Cmd {
	return func() tea.Msg {
		var err error
		dev.Debug(fmt.Sprintf("SENDING %s", shellCmd))
		// jsonCmd, err := formatter.JsonEncodedTokenArray(shellCmd)
		// if err != nil {
		// 	return message.ErrMsg{Err: err}
		// }
		encodedCmd := b64.StdEncoding.EncodeToString([]byte(shellCmd))

		toSend := fmt.Sprintf(`{"stdin": {"data": %s}}`, encodedCmd)

		err = ws.WriteMessage(1, []byte(toSend))
		if err != nil {
			return message.ErrMsg{Err: err}
		}
		dev.Debug("SENT")

		// TODO LEO: don't think we actually want this timeout as it cancels all future reads
		// setTimeoutErr := ws.SetReadDeadline(time.Now().Add(3 * time.Second))
		// if setTimeoutErr != nil {
		// 	return message.ErrMsg{Err: setTimeoutErr}
		// }

		dev.Debug("READING WS")
		msgType, content, err := ws.ReadMessage()
		dev.Debug(fmt.Sprintf("READ MESSAGE type %d content %s err %e", msgType, content, err))

		if err != nil {
			// timeouts can occur when there's just no response, like sending "/bin/sh" as a first command
			if !strings.Contains(err.Error(), "i/o timeout") {
				return message.ErrMsg{Err: err}
			}
		}

		// TODO LEO: parse response here, handle output, decode the base64 and return real msg
		return ExecWebSocketResponseMsg{StdOut: "hi"}
	}
}
