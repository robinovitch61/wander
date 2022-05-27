package nomad

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
	"wander/components/page"
	"wander/dev"
	"wander/message"
)

func FetchExecSession(host, token, allocID, taskName string) tea.Cmd {
	return func() tea.Msg {
		// strip off any other protocol in the host
		host = strings.Split(host, "://")[1]
		dev.Debug("HOST")
		dev.Debug(host)

		path := fmt.Sprintf("/v1/client/allocation/%s/exec", allocID)
		params := map[string]string{
			"command": "[\"echo\", \"hi\"]",
			"task":    taskName,
			// "ws_handshake": "true",
			// "token":        token,
		}

		ws, err := getWebsocketConnection(host, path, token, params)
		if err != nil {
			return message.ErrMsg{Err: err}
		}

		return PageLoadedMsg{
			Page:        ExecPage,
			TableHeader: []string{},
			AllPageData: []page.Row{},
			Websocket:   ws,
		}
	}
}
