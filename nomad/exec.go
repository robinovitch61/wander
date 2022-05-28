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
}

func ReadExecWebSocketNextMessage(ws *websocket.Conn) tea.Cmd {
	return func() tea.Msg {
		stdout, stderr, err := readNext(ws)
		if err != nil {
			return message.ErrMsg{Err: err}
		}
		return ExecWebSocketResponseMsg{StdOut: stdout, StdErr: stderr}
	}
}

func SendAndReadExecWebSocketMessage(ws *websocket.Conn, shellCmd string) tea.Cmd {
	return func() tea.Msg {
		var err error
		var stdout, stderr string
		var so, se string
		for _, char := range shellCmd + "\n" {
			// send char
			err = send(ws, string(char))
			if err != nil {
				return message.ErrMsg{Err: err}
			}

			so, se, err = readNext(ws)
			if err != nil {
				return message.ErrMsg{Err: err}
			}
			stdout += so
			stderr += se
		}
		so, se, err = readNext(ws)
		if err != nil {
			return message.ErrMsg{Err: err}
		}
		stdout += so
		stderr += se
		return ExecWebSocketResponseMsg{StdOut: stdout, StdErr: stderr}
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

func send(ws *websocket.Conn, r string) error {
	encoded := b64.StdEncoding.EncodeToString([]byte(r))
	toSend := fmt.Sprintf(`{"stdin":{"data":"%s"}}`, encoded)
	return ws.WriteMessage(1, []byte(toSend))
}

func readNext(ws *websocket.Conn) (string, string, error) {
	var stdout, stderr string

	msgType, content, err := ws.ReadMessage()
	if err != nil {
		return "", "", err
	}

	stdout, stderr, err = parseWebSocketMessage(msgType, content)
	if err != nil {
		return "", "", err
	}

	return stdout, stderr, nil
}

func parseWebSocketMessage(msgType int, content []byte) (string, string, error) {
	var err error

	response := execResponseJSON{}
	err = json.Unmarshal(content, &response)
	if err != nil {
		return "", "", err
	}

	var stdout, stderr string
	switch {
	case response.StdOut != execResponseDataJSON{}:
		if stdOutData := response.StdOut.Data; stdOutData != "" {
			decoded, err := b64.StdEncoding.DecodeString(stdOutData)
			if err != nil {
				return "", "", err
			}
			stdout += string(decoded)
		}
	default:
		panic(fmt.Sprintf("Unhandled websocket response: %s", content))
	}

	return normalizeLineEndings(stdout), normalizeLineEndings(stderr), nil
}

func normalizeLineEndings(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}
