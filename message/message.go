package message

import (
	b64 "encoding/base64"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
	"wander/dev"
	"wander/formatter"
)

type ErrMsg struct{ Err error }

func (e ErrMsg) Error() string { return e.Err.Error() }

func SendExec(shellCmd string, ws *websocket.Conn) tea.Cmd {
	return func() tea.Msg {
		json, err := formatter.JsonEncodedTokenArray(shellCmd)
		if err == nil && ws != nil {
			dev.Debug(json)
			// TODO LEO: return cmd here that does the websocket async
		}

		encoded := b64.StdEncoding.EncodeToString([]byte(json))
		err = ws.WriteMessage(1, []byte(fmt.Sprintf("{\"stdin\": {\"data\": \"%s\"}}", encoded)))
		if err != nil {
			return ErrMsg{Err: err}
		}

		var messageType int
		var messageContent []byte
		for {
			messageType, messageContent, err = ws.ReadMessage()
			dev.Debug(fmt.Sprintf("type %d content %s", messageType, string(messageContent)))

		}

		// for _, line := range strings.Split(string(messageContent), "\n") {
		//
		// }
		// return AppendToPageMsg{}
	}
}
