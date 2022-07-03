package nomad

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
	"github.com/robinovitch61/wander/internal/tui/constants"
	"github.com/robinovitch61/wander/internal/tui/formatter"
	"github.com/robinovitch61/wander/internal/tui/message"
	"strings"
	"time"
)

type ExecWebSocketConnectedMsg struct {
	WebSocketConnection *websocket.Conn
}

type ExecWebSocketResponseMsg struct {
	StdOut, StdErr string
	Close          bool
}

type ExecWebSocketHeartbeatMsg struct{}

func InitiateWebSocket(host, token, allocID, taskName, command string) tea.Cmd {
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

func ReadExecWebSocketNextMessage(ws *websocket.Conn) tea.Cmd {
	return func() tea.Msg {
		nextMsg := readNext(ws)
		if nextMsg.Err != nil {
			return message.ErrMsg{Err: nextMsg.Err}
		}
		return ExecWebSocketResponseMsg{StdOut: nextMsg.StdOut, StdErr: nextMsg.StdErr, Close: nextMsg.Close}
	}
}

func SendWebSocketMessage(ws *websocket.Conn, keyPress string) tea.Cmd {
	return func() tea.Msg {
		var err error
		err = sendStdInData(ws, keyPress)
		if err != nil {
			return message.ErrMsg{Err: err}
		}
		return nil
	}
}

func ResizeTty(ws *websocket.Conn, width, height int) tea.Cmd {
	return func() tea.Msg {
		var err error
		err = sendTtyResize(ws, width, height)
		if err != nil {
			return message.ErrMsg{Err: err}
		}
		return nil
	}
}

func SendHeartbeatWithDelay() tea.Cmd {
	return tea.Tick(constants.ExecWebSocketHeartbeatDuration, func(t time.Time) tea.Msg { return ExecWebSocketHeartbeatMsg{} })
}

func SendHeartbeat(ws *websocket.Conn) tea.Cmd {
	return func() tea.Msg {
		var err error
		err = ws.WriteMessage(1, []byte("{}"))
		if err != nil {
			return message.ErrMsg{Err: err}
		}
		return SendHeartbeatWithDelay()
	}
}

func CloseWebSocket(ws *websocket.Conn) tea.Cmd {
	return func() tea.Msg {
		var err error
		err = sendStdInData(ws, string(rune(4)))
		if err != nil {
			if !strings.Contains(err.Error(), "write: broken pipe") {
				return message.ErrMsg{Err: err}
			}
		}
		return nil
	}
}

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

func sendStdInData(ws *websocket.Conn, r string) error {
	encoded := b64.StdEncoding.EncodeToString([]byte(r))
	toSend := fmt.Sprintf(`{"stdin":{"data":"%s"}}`, encoded)
	return ws.WriteMessage(1, []byte(toSend))
}

func sendTtyResize(ws *websocket.Conn, width, height int) error {
	toSend := fmt.Sprintf(`{"tty_size": {"height": %d, "width": %d}}`, height, width)
	return ws.WriteMessage(1, []byte(toSend))
}

func readNext(ws *websocket.Conn) parsedWebSocketMessage {
	msgType, content, err := ws.ReadMessage()
	if err != nil {
		closedConnUsed := strings.Contains(err.Error(), "use of closed network connection")
		abnormalClosure := strings.Contains(err.Error(), "close 1006")
		if closedConnUsed || abnormalClosure {
			return parsedWebSocketMessage{Close: true}
		}
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
