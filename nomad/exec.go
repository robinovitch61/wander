package nomad

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
	"strings"
	"wander/message"
)

type WebSocketConnectedMsg struct {
	Conn *websocket.Conn
}

func InitiateWebSocketConnection(host, token, allocID, taskName, command string) tea.Cmd {
	return func() tea.Msg {
		// strip off any other protocol in the host
		secure := false
		if strings.Contains(host, "https://") {
			secure = true
		}

		host = strings.Split(host, "://")[1]

		path := fmt.Sprintf("/v1/client/allocation/%s/exec", allocID)
		params := map[string]string{
			"command": command,
			"task":    taskName,
		}

		ws, err := getWebSocketConnection(secure, host, path, token, params)
		if err != nil {
			return message.ErrMsg{Err: err}
		}

		return WebSocketConnectedMsg{Conn: ws}
	}
}
