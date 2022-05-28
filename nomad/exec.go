package nomad

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
	"strings"
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
			"tty":     "true",
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
	Close          bool
}

func ReadExecWebSocketNextMessage(ws *websocket.Conn) tea.Cmd {
	return func() tea.Msg {
		nextMsg := readNext(ws)
		if nextMsg.Err != nil {
			return message.ErrMsg{Err: nextMsg.Err}
		}
		return ExecWebSocketResponseMsg{StdOut: nextMsg.StdOut, StdErr: nextMsg.StdErr, Close: nextMsg.Close}
	}
}

func SendAndReadExecWebSocketMessage(ws *websocket.Conn, shellCmd string) tea.Cmd {
	return func() tea.Msg {
		var err error
		var finalMsg parsedWebSocketMessage
		for _, char := range shellCmd + "\n" {
			err = send(ws, string(char))
			if err != nil {
				return message.ErrMsg{Err: err}
			}
			finalMsg = appendNextMsg(ws, finalMsg)
			if finalMsg.Err != nil {
				return message.ErrMsg{Err: finalMsg.Err}
			}
		}
		// after sending all chars including \n, next message will be response of command
		finalMsg = appendNextMsg(ws, finalMsg)
		if finalMsg.Err != nil {
			return message.ErrMsg{Err: finalMsg.Err}
		}
		return ExecWebSocketResponseMsg{StdOut: finalMsg.StdOut, StdErr: finalMsg.StdErr, Close: finalMsg.Close}
	}
}

type exitJSON struct {
	ExitCode int `json:"exit_code"`
}

type execResponseDataJSON struct {
	Data   string   `json:"data"`
	Close  bool     `json:"close"`
	Result exitJSON `json:"result"`
}

type execResponseJSON struct {
	StdOut execResponseDataJSON `json:"stdout"`
	StdErr execResponseDataJSON `json:"stderr"`
	Exited bool                 `json:"exited"`
}

func appendNextMsg(ws *websocket.Conn, prevMsg parsedWebSocketMessage) parsedWebSocketMessage {
	nextMsg := readNext(ws)
	return parsedWebSocketMessage{
		StdOut: prevMsg.StdOut + nextMsg.StdOut,
		StdErr: prevMsg.StdErr + nextMsg.StdErr,
		Close:  prevMsg.Close || nextMsg.Close,
		Err:    nextMsg.Err,
	}
}

func send(ws *websocket.Conn, r string) error {
	encoded := b64.StdEncoding.EncodeToString([]byte(r))
	toSend := fmt.Sprintf(`{"stdin":{"data":"%s"}}`, encoded)
	return ws.WriteMessage(1, []byte(toSend))
}

func readNext(ws *websocket.Conn) parsedWebSocketMessage {
	// TODO LEO: with large responses, multiple messages per stdin :/
	msgType, content, err := ws.ReadMessage()
	if err != nil {
		return parsedWebSocketMessage{Err: err}
	}
	return parseWebSocketMessage(msgType, content)
}

type parsedWebSocketMessage struct {
	StdOut, StdErr string
	Close          bool
	Err            error
}

func parseWebSocketMessage(msgType int, content []byte) parsedWebSocketMessage {
	var err error

	response := execResponseJSON{}
	err = json.Unmarshal(content, &response)
	if err != nil {
		return parsedWebSocketMessage{Err: err}
	}

	var stdout, stderr string
	switch {
	case response.StdOut != execResponseDataJSON{}:
		if stdOutData := response.StdOut.Data; stdOutData != "" {
			decoded, err := b64.StdEncoding.DecodeString(stdOutData)
			if err != nil {
				return parsedWebSocketMessage{Err: err}
			}
			stdout += string(decoded)
		}

	case response.StdErr != execResponseDataJSON{}:
		if stdErrData := response.StdErr.Data; stdErrData != "" {
			decoded, err := b64.StdEncoding.DecodeString(stdErrData)
			if err != nil {
				return parsedWebSocketMessage{Err: err}
			}
			stderr += string(decoded)
		} else if stdErrClose := response.StdErr.Close; stdErrClose {
			return parsedWebSocketMessage{Close: true}
		}

	case response.Exited:
		return parsedWebSocketMessage{Close: true}

	default:
		panic(fmt.Sprintf("Unhandled websocket response: %s (msgType %d)", content, msgType))
	}

	return parsedWebSocketMessage{
		StdOut: normalizeLineEndings(stdout),
		StdErr: normalizeLineEndings(stderr),
	}
}

func normalizeLineEndings(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}
